package imapx

import (
	"bytes"
	"crypto/tls"
	"net"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"

	"webmail/backend/internal/config"
	"webmail/backend/internal/mimex"
)

const (
	dialTimeout          = 10 * time.Second
	handshakeLoginTimout = 15 * time.Second
)

// Dial connects and logs into the IMAP server with the given credentials.
// Connect + handshake + login are bounded by timeouts so a blocked or
// unresponsive server fails fast instead of hanging the request.
func Dial(cfg *config.Config, username, password string) (*client.Client, error) {
	addr := net.JoinHostPort(cfg.IMAPHost, strconv.Itoa(cfg.IMAPPort))
	tlsConf := &tls.Config{ServerName: cfg.IMAPHost, InsecureSkipVerify: cfg.IMAPSkipVerify}

	rawConn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("kết nối IMAP (timeout 10s): %w", err)
	}
	_ = rawConn.SetDeadline(time.Now().Add(handshakeLoginTimout))

	var c *client.Client
	switch cfg.IMAPMode {
	case "starttls":
		c, err = client.New(rawConn)
		if err == nil {
			if err = c.StartTLS(tlsConf); err != nil {
				rawConn.Close()
				return nil, fmt.Errorf("IMAP STARTTLS: %w", err)
			}
		}
	case "plain":
		c, err = client.New(rawConn)
	default:
		tlsConn := tls.Client(rawConn, tlsConf)
		if err := tlsConn.Handshake(); err != nil {
			rawConn.Close()
			return nil, fmt.Errorf("bắt tay TLS với IMAP: %w", err)
		}
		c, err = client.New(tlsConn)
	}
	if err != nil {
		rawConn.Close()
		return nil, fmt.Errorf("kết nối IMAP: %w", err)
	}

	if err := c.Login(username, password); err != nil {
		c.Logout()
		return nil, fmt.Errorf("đăng nhập IMAP thất bại: %w", err)
	}
	_ = rawConn.SetDeadline(time.Time{})
	return c, nil
}

// Close logs the session out, ignoring errors.
func Close(c *client.Client) {
	if c != nil {
		c.Logout()
	}
}

// ---------- mailboxes ----------

type Mailbox struct {
	Name       string   `json:"name"`
	Delimiter  string   `json:"delimiter"`
	Attributes []string `json:"attributes"`
	Total      int      `json:"total"`
	Unseen     int      `json:"unseen"`
}

func ListMailboxes(c *client.Client, withStatus bool) ([]Mailbox, error) {
	ch := make(chan *imap.MailboxInfo, 64)
	if err := c.List("", "*", ch); err != nil {
		return nil, fmt.Errorf("lỗi liệt kê hộp thư: %w", err)
	}
	out := []Mailbox{}
	for info := range ch {
		if info.Name == "" {
			continue
		}
		mb := Mailbox{
			Name:       info.Name,
			Delimiter:  string(info.Delimiter),
			Attributes: info.Attributes,
		}
		if withStatus {
			if st, err := c.Status(info.Name, []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen}); err == nil && st != nil {
				mb.Total = int(st.Messages)
				mb.Unseen = int(st.Unseen)
			}
		}
		out = append(out, mb)
	}
	sortMailboxes(out)
	return out, nil
}

// ListMailboxesPooled lists mailboxes and fills total/unseen counts using
// parallel STATUS commands over pooled connections.
func ListMailboxesPooled(p *Pool, user, pass string, workers int) ([]Mailbox, error) {
	var boxes []Mailbox
	err := p.Do(user, pass, func(c *client.Client) error {
		var e error
		boxes, e = ListMailboxes(c, false)
		return e
	})
	if err != nil {
		return nil, err
	}
	if len(boxes) == 0 {
		return boxes, nil
	}
	if workers < 1 {
		workers = 3
	}

	type status struct{ total, unseen int }
	results := make([]status, len(boxes))
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := range boxes {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			name := boxes[i].Name
			var st status
			err := p.Do(user, pass, func(c *client.Client) error {
				s, e := c.Status(name, []imap.StatusItem{imap.StatusMessages, imap.StatusUnseen})
				if e != nil {
					return e
				}
				if s != nil {
					st = status{total: int(s.Messages), unseen: int(s.Unseen)}
				}
				return nil
			})
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			results[i] = st
		}(i)
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	for i := range boxes {
		boxes[i].Total = results[i].total
		boxes[i].Unseen = results[i].unseen
	}
	return boxes, nil
}

func sortMailboxes(list []Mailbox) {
	rank := func(mb Mailbox) int {
		switch mb.Name {
		case "INBOX":
			return 0
		}
		for _, a := range mb.Attributes {
			switch a {
			case `\Sent`:
				return 1
			case `\Drafts`:
				return 2
			case `\Junk`:
				return 4
			case `\Trash`:
				return 5
			case `\Archive`:
				return 6
			}
		}
		name := strings.ToLower(mb.Name)
		base := name
		if i := strings.LastIndex(name, "/"); i >= 0 {
			base = name[i+1:]
		}
		if i := strings.LastIndex(base, ">"); i >= 0 { // "[Gmail]/Sent Mail" style handled loosely below
			base = base[i+1:]
		}
		switch base {
		case "sent", "sent messages", "sent items", "sent mail", "đã gửi":
			return 1
		case "drafts", "draft", "nháp":
			return 2
		case "junk", "spam", "bulk mail", "thư rác":
			return 4
		case "trash", "deleted messages", "deleted items", "bin", "thùng rác":
			return 5
		case "archive", "all mail", "archived":
			return 6
		}
		return 3
	}
	// stable insertion-ish sort by rank then name
	for i := 1; i < len(list); i++ {
		for j := i; j > 0; j-- {
			a, b := list[j-1], list[j]
			ra, rb := rank(a), rank(b)
			if ra < rb || (ra == rb && strings.ToLower(a.Name) <= strings.ToLower(b.Name)) {
				break
			}
			list[j-1], list[j] = list[j], list[j-1]
		}
	}
}

// ---------- message list ----------

type Summary struct {
	UID       uint32    `json:"uid"`
	Subject   string    `json:"subject"`
	From      string    `json:"from"`
	FromEmail string    `json:"fromEmail"`
	Date      time.Time `json:"date"`
	Size      uint32    `json:"size"`
	Flags     []string  `json:"flags"`
	Seen      bool      `json:"seen"`
	Flagged   bool      `json:"flagged"`
	Snippet   string    `json:"snippet"`
}

func ListMessages(c *client.Client, mailbox string, page, limit int) (int, []Summary, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if page < 0 {
		page = 0
	}
	mbox, err := c.Select(mailbox, true)
	if err != nil {
		return 0, nil, fmt.Errorf("không mở được hộp thư %q: %w", mailbox, err)
	}
	total := int(mbox.Messages)
	if total == 0 {
		return total, []Summary{}, nil
	}
	end := total - page*limit
	start := end - limit + 1
	if end < 1 {
		return total, []Summary{}, nil
	}
	if start < 1 {
		start = 1
	}

	set := new(imap.SeqSet)
	seqs := make([]uint32, 0, end-start+1)
	for i := start; i <= end; i++ {
		seqs = append(seqs, uint32(i))
	}
	set.AddNum(seqs...)

	textSection := &imap.BodySectionName{
		BodyPartName: imap.BodyPartName{Specifier: imap.TextSpecifier},
		Peek:         true,
		Partial:      []int{0, 4096},
	}
	metaItems := []imap.FetchItem{
		imap.FetchUid,
		imap.FetchFlags,
		imap.FetchEnvelope,
		imap.FetchRFC822Size,
		imap.FetchInternalDate,
	}

	// Single combined fetch (metadata + partial body) saves a round trip.
	// Some servers (e.g. GreenMail) respond non-compliantly to partial
	// fetches and return nothing - then retry with metadata only.
	ch := make(chan *imap.Message, limit)
	if err := c.Fetch(set, append(metaItems, textSection.FetchItem()), ch); err != nil {
		return 0, nil, fmt.Errorf("lỗi tải danh sách thư: %w", err)
	}
	var msgs []Summary
	for m := range ch {
		if m == nil {
			continue
		}
		msgs = append(msgs, summaryFrom(m))
		if r := m.GetBody(textSection); r != nil {
			msgs[len(msgs)-1].Snippet = mimex.ExtractSnippet(readAll(r))
		}
	}

	if len(msgs) == 0 {
		// fallback: metadata only (server broke on the partial fetch)
		ch = make(chan *imap.Message, limit)
		if err := c.Fetch(set, metaItems, ch); err != nil {
			return 0, nil, fmt.Errorf("lỗi tải danh sách thư: %w", err)
		}
		for m := range ch {
			if m == nil {
				continue
			}
			msgs = append(msgs, summaryFrom(m))
		}
	}
	return total, msgs, nil
}

func summaryFrom(m *imap.Message) Summary {
	s := Summary{
		UID:   m.Uid,
		Size:  m.Size,
		Date:  m.InternalDate,
		Flags: m.Flags,
	}
	if m.Envelope != nil {
		s.Subject = mimex.DecodeHeaderWords(m.Envelope.Subject)
		if !m.Envelope.Date.IsZero() {
			s.Date = m.Envelope.Date
		}
		froms := envelopeAddrs(m.Envelope.From)
		if len(froms) > 0 {
			s.FromEmail = froms[0].Email
			if froms[0].Name != "" {
				s.From = froms[0].Name
			} else {
				s.From = froms[0].Email
			}
		}
	}
	s.Seen = hasFlag(m.Flags, imap.SeenFlag)
	s.Flagged = hasFlag(m.Flags, imap.FlaggedFlag)
	return s
}

func readAll(r io.Reader) []byte {
	if r == nil {
		return nil
	}
	b, err := io.ReadAll(io.LimitReader(r, 8192))
	if err != nil {
		return nil
	}
	return b
}

func envelopeAddrs(list []*imap.Address) []mimex.Addr {
	out := make([]mimex.Addr, 0, len(list))
	for _, a := range list {
		if a == nil {
			continue
		}
		email := a.Address()
		name := mimex.DecodeHeaderWords(a.PersonalName)
		out = append(out, mimex.Addr{Name: name, Email: email})
	}
	return out
}

func hasFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

// ---------- single message ----------

func GetMessage(c *client.Client, mailbox string, uid uint32, markSeen bool) (*mimex.ParsedMessage, error) {
	if _, err := c.Select(mailbox, false); err != nil {
		return nil, fmt.Errorf("không mở được hộp thư %q: %w", mailbox, err)
	}
	set := new(imap.SeqSet)
	set.AddNum(uid)
	fullSection := &imap.BodySectionName{Peek: true}
	items := []imap.FetchItem{
		imap.FetchUid,
		imap.FetchFlags,
		fullSection.FetchItem(),
	}
	ch := make(chan *imap.Message, 1)
	if err := c.UidFetch(set, items, ch); err != nil {
		return nil, fmt.Errorf("lỗi tải thư: %w", err)
	}
	m := <-ch
	if m == nil {
		return nil, fmt.Errorf("không tìm thấy thư")
	}
	raw := readAllFull(m.GetBody(fullSection))
	msg, err := mimex.Parse(raw)
	if err != nil {
		return nil, err
	}
	msg.UID = uid
	msg.Flags = m.Flags
	if markSeen {
		_ = storeFlags(c, set, imap.AddFlags, []string{imap.SeenFlag})
	}
	return msg, nil
}

func readAllFull(r io.Reader) []byte {
	if r == nil {
		return nil
	}
	b, err := io.ReadAll(io.LimitReader(r, mimex.MaxMessageBytes))
	if err != nil {
		return nil
	}
	return b
}

// GetAttachment fetches a single MIME part by IMAP section number.
func GetAttachment(c *client.Client, mailbox string, uid uint32, section string) (mimex.PartContent, error) {
	if section == "" {
		return mimex.PartContent{}, fmt.Errorf("thiếu phần đính kèm")
	}
	path, err := parseSectionPath(section)
	if err != nil {
		return mimex.PartContent{}, err
	}
	if _, err := c.Select(mailbox, true); err != nil {
		return mimex.PartContent{}, fmt.Errorf("không mở được hộp thư %q: %w", mailbox, err)
	}
	set := new(imap.SeqSet)
	set.AddNum(uid)
	partSection := &imap.BodySectionName{
		BodyPartName: imap.BodyPartName{Path: path},
		Peek:         true,
	}
	ch := make(chan *imap.Message, 1)
	if err := c.UidFetch(set, []imap.FetchItem{partSection.FetchItem()}, ch); err != nil {
		return mimex.PartContent{}, fmt.Errorf("lỗi tải đính kèm: %w", err)
	}
	m := <-ch
	if m == nil {
		return mimex.PartContent{}, fmt.Errorf("không tìm thấy thư")
	}
	raw := readAllFull(m.GetBody(partSection))
	if len(raw) == 0 {
		return mimex.PartContent{}, fmt.Errorf("không tìm thấy phần đính kèm")
	}
	return mimex.ParsePart(raw)
}

// parseSectionPath converts "2.1" into [2,1].
func parseSectionPath(section string) ([]int, error) {
	parts := strings.Split(section, ".")
	path := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 1 {
			return nil, fmt.Errorf("số phần không hợp lệ: %q", section)
		}
		path = append(path, n)
	}
	return path, nil
}

// ---------- flags ----------

func UpdateFlags(c *client.Client, mailbox string, uid uint32, seen, flagged, deleted *bool, trashMailbox string) error {
	if _, err := c.Select(mailbox, false); err != nil {
		return fmt.Errorf("không mở được hộp thư %q: %w", mailbox, err)
	}
	set := new(imap.SeqSet)
	set.AddNum(uid)

	if flagged != nil {
		var op imap.FlagsOp = imap.AddFlags
		if !*flagged {
			op = imap.RemoveFlags
		}
		if err := storeFlags(c, set, op, []string{imap.FlaggedFlag}); err != nil {
			return err
		}
	}
	if seen != nil {
		var op imap.FlagsOp = imap.AddFlags
		if !*seen {
			op = imap.RemoveFlags
		}
		if err := storeFlags(c, set, op, []string{imap.SeenFlag}); err != nil {
			return err
		}
	}
	if deleted == nil {
		return nil
	}
	if *deleted {
		if trashMailbox != "" && mailbox != trashMailbox {
			if err := c.UidCopy(set, trashMailbox); err == nil {
				if err := storeFlags(c, set, imap.AddFlags, []string{imap.DeletedFlag}); err != nil {
					return err
				}
				return expunge(c)
			}
		}
		if err := storeFlags(c, set, imap.AddFlags, []string{imap.DeletedFlag}); err != nil {
			return err
		}
		return expunge(c)
	}
	return storeFlags(c, set, imap.RemoveFlags, []string{imap.DeletedFlag})
}

func storeFlags(c *client.Client, set *imap.SeqSet, op imap.FlagsOp, flags []string) error {
	item := imap.FormatFlagsOp(op, true)
	vals := make([]interface{}, 0, len(flags))
	for _, f := range flags {
		vals = append(vals, f)
	}
	return c.UidStore(set, item, vals, nil)
}

func expunge(c *client.Client) error {
	ch := make(chan uint32, 32)
	return c.Expunge(ch)
}

// ---------- append ----------

type literal struct {
	r *bytes.Reader
}

func newLiteral(b []byte) literal {
	return literal{r: bytes.NewReader(b)}
}

func (l literal) Read(p []byte) (int, error) { return l.r.Read(p) }
func (l literal) Len() int                   { return int(l.r.Size()) }

// TryAppend appends the raw message to the first existing mailbox in names.
func TryAppend(c *client.Client, names []string, msg []byte) error {
	var lastErr error
	for _, name := range names {
		if name == "" {
			continue
		}
		if err := c.Append(name, []string{imap.SeenFlag}, time.Now(), newLiteral(msg)); err == nil {
			return nil
		} else {
			lastErr = err
		}
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("không có hộp thư để lưu")
	}
	return lastErr
}
