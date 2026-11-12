package imapx

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/emersion/go-imap/client"

	"webmail/backend/internal/config"
)

const (
	poolIdleTTL     = 10 * time.Minute // discard connections idle longer than this
	poolHealthAfter = 30 * time.Second // NOOP-check connections idle longer than this
	poolSweepEvery  = time.Minute
)

type entry struct {
	c        *client.Client
	user     string
	passHash string
	lastUsed time.Time
}

// Pool keeps logged-in IMAP connections per user for reuse, avoiding the
// TLS+LOGIN handshake cost on every API request.
type Pool struct {
	mu         sync.Mutex
	cfg        *config.Config
	idle       []*entry
	maxPerUser int
}

func NewPool(cfg *config.Config) *Pool {
	max := 4
	p := &Pool{cfg: cfg, maxPerUser: max}
	go p.sweeper()
	return p
}

func passFingerprint(pass string) string {
	sum := sha256.Sum256([]byte(pass))
	return hex.EncodeToString(sum[:8])
}

func (p *Pool) dial(user, pass string) (*client.Client, error) {
	return Dial(p.cfg, user, pass)
}

// acquire returns a healthy, logged-in connection (reused if possible).
func (p *Pool) acquire(user, pass string) (*client.Client, error) {
	fp := passFingerprint(pass)
	var picked *entry
	var stale []*entry

	p.mu.Lock()
	for i := len(p.idle) - 1; i >= 0; i-- {
		e := p.idle[i]
		if e.user != user {
			continue
		}
		if e.passHash == fp {
			p.idle = append(p.idle[:i], p.idle[i+1:]...)
			picked = e
			break
		}
		// same user, different password (re-login) -> drop old connection
		p.idle = append(p.idle[:i], p.idle[i+1:]...)
		stale = append(stale, e)
	}
	p.mu.Unlock()

	for _, e := range stale {
		e.c.Logout()
	}

	if picked != nil {
		idleFor := time.Since(picked.lastUsed)
		if idleFor > poolIdleTTL {
			picked.c.Logout()
			return p.dial(user, pass)
		}
		if idleFor > poolHealthAfter {
			if err := picked.c.Noop(); err != nil {
				picked.c.Logout()
				return p.dial(user, pass)
			}
		}
		return picked.c, nil
	}
	return p.dial(user, pass)
}

func (p *Pool) release(user, pass string, c *client.Client) {
	p.mu.Lock()
	count := 0
	for _, e := range p.idle {
		if e.user == user {
			count++
		}
	}
	if count >= p.maxPerUser {
		p.mu.Unlock()
		c.Logout()
		return
	}
	p.idle = append(p.idle, &entry{
		c:        c,
		user:     user,
		passHash: passFingerprint(pass),
		lastUsed: time.Now(),
	})
	p.mu.Unlock()
}

func (p *Pool) discard(c *client.Client) {
	if c != nil {
		c.Logout()
	}
}

// Do runs fn with a healthy pooled connection. On error the connection is
// discarded (conservative: avoids reusing a possibly desynced connection).
func (p *Pool) Do(user, pass string, fn func(c *client.Client) error) error {
	c, err := p.acquire(user, pass)
	if err != nil {
		return err
	}
	if err := fn(c); err != nil {
		p.discard(c)
		return err
	}
	p.release(user, pass, c)
	return nil
}

// Close logs out every pooled connection.
func (p *Pool) Close() {
	p.mu.Lock()
	idle := p.idle
	p.idle = nil
	p.mu.Unlock()
	for _, e := range idle {
		e.c.Logout()
	}
}

func (p *Pool) sweeper() {
	for range time.Tick(poolSweepEvery) {
		cut := time.Now().Add(-poolIdleTTL)
		var expired []*entry
		p.mu.Lock()
		keep := p.idle[:0]
		for _, e := range p.idle {
			if e.lastUsed.Before(cut) {
				expired = append(expired, e)
			} else {
				keep = append(keep, e)
			}
		}
		p.idle = keep
		p.mu.Unlock()
		for _, e := range expired {
			e.c.Logout()
		}
	}
}
