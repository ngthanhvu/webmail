package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

const cookieName = "webmail_session"

type Data struct {
	Username string `json:"u"`
	Password string `json:"p"`
	Expires  int64  `json:"e"`
}

type Manager struct {
	key []byte
	ttl time.Duration
}

func New(secret string, ttlHours int) *Manager {
	sum := sha256.Sum256([]byte(secret))
	return &Manager{key: sum[:], ttl: time.Duration(ttlHours) * time.Hour}
}

func (m *Manager) aead() (cipher.AEAD, error) {
	block, err := aes.NewCipher(m.key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func (m *Manager) seal(d Data) (string, error) {
	plaintext, err := json.Marshal(d)
	if err != nil {
		return "", err
	}
	gcm, err := m.aead()
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ct := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ct), nil
}

func (m *Manager) unseal(s string) (Data, error) {
	var d Data
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return d, errors.New("token không hợp lệ")
	}
	gcm, err := m.aead()
	if err != nil {
		return d, err
	}
	if len(raw) < gcm.NonceSize() {
		return d, errors.New("token không hợp lệ")
	}
	nonce, ct := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return d, errors.New("token không hợp lệ")
	}
	if err := json.Unmarshal(plaintext, &d); err != nil {
		return d, err
	}
	if time.Now().Unix() > d.Expires {
		return d, errors.New("phiên đã hết hạn")
	}
	return d, nil
}

func (m *Manager) SetCookie(w http.ResponseWriter, username, password string) error {
	tok, err := m.seal(Data{
		Username: username,
		Password: password,
		Expires:  time.Now().Add(m.ttl).Unix(),
	})
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(m.ttl.Seconds()),
	})
	return nil
}

func (m *Manager) Clear(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func (m *Manager) Get(r *http.Request) (Data, error) {
	ck, err := r.Cookie(cookieName)
	if err != nil {
		return Data{}, errors.New("chưa đăng nhập")
	}
	return m.unseal(ck.Value)
}
