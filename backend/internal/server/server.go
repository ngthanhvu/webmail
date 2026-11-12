package server

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/emersion/go-imap/client"

	"webmail/backend/internal/config"
	"webmail/backend/internal/imapx"
	"webmail/backend/internal/mimex"
	"webmail/backend/internal/session"
	"webmail/backend/internal/smtpx"
)

type Server struct {
	cfg  *config.Config
	sess *session.Manager
	pool *imapx.Pool
	mux  *http.ServeMux
}

func New(cfg *config.Config) *Server {
	s := &Server{
		cfg:  cfg,
		sess: session.New(cfg.AppSecret, cfg.SessionHours),
		pool: imapx.NewPool(cfg),
		mux:  http.NewServeMux(),
	}
	s.mux.HandleFunc("POST /api/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/logout", s.handleLogout)
	s.mux.HandleFunc("GET /api/me", s.auth(s.handleMe))
	s.mux.HandleFunc("GET /api/mailboxes", s.auth(s.handleMailboxes))
	s.mux.HandleFunc("GET /api/messages", s.auth(s.handleListMessages))
	s.mux.HandleFunc("GET /api/messages/{uid}", s.auth(s.handleGetMessage))
	s.mux.HandleFunc("GET /api/messages/{uid}/attachment", s.auth(s.handleAttachment))
	s.mux.HandleFunc("POST /api/messages/{uid}/flags", s.auth(s.handleFlags))
	s.mux.HandleFunc("POST /api/send", s.auth(s.handleSend))
	s.mux.HandleFunc("/", s.handleStatic)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

func (s *Server) auth(next func(http.ResponseWriter, *http.Request, session.Data)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := s.sess.Get(r)
		if err != nil {
			writeErr(w, http.StatusUnauthorized, "Phiên hết hạn, vui lòng đăng nhập lại")
			return
		}
		next(w, r, d)
	}
}

// ---------- handlers ----------

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil || req.Username == "" || req.Password == "" {
		writeErr(w, http.StatusBadRequest, "Vui lòng nhập tài khoản và mật khẩu")
		return
	}
	// Verify credentials via IMAP; the connection is kept in the pool so
	// subsequent requests are fast.
	if err := s.pool.Do(req.Username, req.Password, func(c *client.Client) error {
		return c.Noop()
	}); err != nil {
		writeErr(w, http.StatusUnauthorized, "Đăng nhập thất bại: "+err.Error())
		return
	}
	if err := s.sess.SetCookie(w, req.Username, req.Password); err != nil {
		writeErr(w, http.StatusInternalServerError, "Lỗi tạo phiên đăng nhập")
		return
	}
	writeJSON(w, map[string]any{"ok": true, "username": req.Username})
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	s.sess.Clear(w)
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request, d session.Data) {
	writeJSON(w, map[string]any{"username": d.Username})
}

func (s *Server) handleMailboxes(w http.ResponseWriter, r *http.Request, d session.Data) {
	boxes, err := imapx.ListMailboxesPooled(s.pool, d.Username, d.Password, 3)
	if err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, map[string]any{"mailboxes": boxes})
}

func (s *Server) handleListMessages(w http.ResponseWriter, r *http.Request, d session.Data) {
	q := r.URL.Query()
	mailbox := q.Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var total int
	var msgs []imapx.Summary
	err := s.pool.Do(d.Username, d.Password, func(c *client.Client) error {
		var e error
		total, msgs, e = imapx.ListMessages(c, mailbox, page, limit)
		return e
	})
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{
		"mailbox":  mailbox,
		"page":     page,
		"limit":    limit,
		"total":    total,
		"messages": msgs,
	})
}

func (s *Server) handleGetMessage(w http.ResponseWriter, r *http.Request, d session.Data) {
	uid, err := strconv.ParseUint(r.PathValue("uid"), 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "UID không hợp lệ")
		return
	}
	mailbox := r.URL.Query().Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}
	markSeen := r.URL.Query().Get("peek") != "1"

	var msg *mimex.ParsedMessage
	err = s.pool.Do(d.Username, d.Password, func(c *client.Client) error {
		var e error
		msg, e = imapx.GetMessage(c, mailbox, uint32(uid), markSeen)
		return e
	})
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, msg)
}

func (s *Server) handleAttachment(w http.ResponseWriter, r *http.Request, d session.Data) {
	uid, err := strconv.ParseUint(r.PathValue("uid"), 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "UID không hợp lệ")
		return
	}
	q := r.URL.Query()
	mailbox, section := q.Get("mailbox"), q.Get("section")
	if mailbox == "" {
		mailbox = "INBOX"
	}

	var part mimex.PartContent
	err = s.pool.Do(d.Username, d.Password, func(c *client.Client) error {
		var e error
		part, e = imapx.GetAttachment(c, mailbox, uint32(uid), section)
		return e
	})
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}

	ct := part.ContentType
	if ct == "" {
		ct = q.Get("contentType")
	}
	if ct == "" {
		ct = "application/octet-stream"
	}
	filename := part.Filename
	if filename == "" {
		filename = q.Get("filename")
	}
	if filename == "" {
		filename = "attachment-" + section
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Length", strconv.Itoa(len(part.Data)))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if isASCII(filename) {
		w.Header().Set("Content-Disposition", `attachment; filename="`+strings.ReplaceAll(filename, `"`, `'`)+`"`)
	} else {
		w.Header().Set("Content-Disposition", "attachment; filename*=UTF-8''"+urlPathEscape(filename))
	}
	w.Write(part.Data)
}

func (s *Server) handleFlags(w http.ResponseWriter, r *http.Request, d session.Data) {
	uid, err := strconv.ParseUint(r.PathValue("uid"), 10, 32)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "UID không hợp lệ")
		return
	}
	mailbox := r.URL.Query().Get("mailbox")
	if mailbox == "" {
		mailbox = "INBOX"
	}
	var req struct {
		Seen    *bool `json:"seen"`
		Flagged *bool `json:"flagged"`
		Deleted *bool `json:"deleted"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4<<10)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "Dữ liệu không hợp lệ")
		return
	}
	if err := s.pool.Do(d.Username, d.Password, func(c *client.Client) error {
		return imapx.UpdateFlags(c, mailbox, uint32(uid), req.Seen, req.Flagged, req.Deleted, s.cfg.TrashMailbox)
	}); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, map[string]any{"ok": true})
}

func (s *Server) handleSend(w http.ResponseWriter, r *http.Request, d session.Data) {
	var req struct {
		To         []string `json:"to"`
		Cc         []string `json:"cc"`
		Bcc        []string `json:"bcc"`
		Subject    string   `json:"subject"`
		Text       string   `json:"text"`
		HTML       string   `json:"html"`
		InReplyTo  string   `json:"inReplyTo"`
		References string   `json:"references"`
		Attachments []struct {
			Filename    string `json:"filename"`
			ContentType string `json:"contentType"`
			ContentB64  string `json:"contentBase64"`
		} `json:"attachments"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 64<<20)).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "Dữ liệu không hợp lệ: "+err.Error())
		return
	}

	from := s.cfg.SMTPFrom
	if from == "" {
		from = d.Username
	}
	var recipients []string
	for _, list := range [][]string{req.To, req.Cc, req.Bcc} {
		for _, raw := range list {
			addr := strings.TrimSpace(raw)
			if addr == "" {
				continue
			}
			parsed, err := mail.ParseAddress(addr)
			if err != nil {
				writeErr(w, http.StatusBadRequest, "Địa chỉ không hợp lệ: "+addr)
				return
			}
			recipients = append(recipients, parsed.Address)
		}
	}
	if len(recipients) == 0 {
		writeErr(w, http.StatusBadRequest, "Vui lòng nhập ít nhất một người nhận")
		return
	}

	atts := make([]mimex.OutgoingAttachment, 0, len(req.Attachments))
	for _, a := range req.Attachments {
		data, err := decodeBase64Data(a.ContentB64)
		if err != nil {
			writeErr(w, http.StatusBadRequest, "Tệp đính kèm không hợp lệ: "+a.Filename)
			return
		}
		atts = append(atts, mimex.OutgoingAttachment{
			Filename:    a.Filename,
			ContentType: a.ContentType,
			Data:        data,
		})
	}

	out := mimex.Outgoing{
		FromName:    localPart(d.Username),
		FromEmail:   from,
		To:          req.To,
		Cc:          req.Cc,
		Bcc:         req.Bcc,
		Subject:     req.Subject,
		Text:        req.Text,
		HTML:        req.HTML,
		InReplyTo:   req.InReplyTo,
		Refs:        req.References,
		Attachments: atts,
	}
	msg, err := mimex.Build(out)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "Lỗi tạo thư: "+err.Error())
		return
	}

	if err := smtpx.Send(s.cfg, d.Username, d.Password, from, recipients, msg); err != nil {
		writeErr(w, http.StatusBadGateway, "Gửi thư thất bại: "+err.Error())
		return
	}

	// best-effort: append a copy to the Sent mailbox
	if err := s.pool.Do(d.Username, d.Password, func(c *client.Client) error {
		names := []string{s.cfg.SentMailbox, "Sent", "Sent Messages", "Sent Items", "INBOX.Sent", "INBOX/Sent", "[Gmail]/Sent Mail"}
		if err := imapx.TryAppend(c, names, msg); err != nil {
			log.Printf("webmail: không lưu thư vào hộp Sent: %v", err)
		}
		return nil
	}); err != nil {
		log.Printf("webmail: không kết nối được để lưu hộp Sent: %v", err)
	}

	writeJSON(w, map[string]any{"ok": true})
}

// ---------- static files (prod) ----------

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeErr(w, http.StatusNotFound, "Không tìm thấy API")
		return
	}
	dist := s.cfg.FrontendDist
	if dist == "" {
		writeErr(w, http.StatusNotFound, "Không cấu hình frontend")
		return
	}
	name := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
	if name == "" || name == "." {
		name = "index.html"
	}
	full := filepath.Join(dist, name)
	if st, err := os.Stat(full); err == nil && !st.IsDir() {
		http.ServeFile(w, r, full)
		return
	}
	indexPath := filepath.Join(dist, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		writeErr(w, http.StatusNotFound, "Không tìm thấy frontend")
		return
	}
	http.ServeFile(w, r, indexPath)
}

// ---------- helpers ----------

func localPart(email string) string {
	if i := strings.LastIndex(email, "@"); i > 0 {
		return email[:i]
	}
	return email
}

func decodeBase64Data(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, ","); i >= 0 && strings.HasPrefix(s, "data:") {
		s = s[i+1:]
	}
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r", ""), "\n", "")
	return base64.StdEncoding.DecodeString(s)
}

func urlPathEscape(s string) string {
	var b strings.Builder
	for _, c := range []byte(s) {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '.' || c == '_' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteString("%")
			b.WriteByte("0123456789ABCDEF"[c>>4])
			b.WriteByte("0123456789ABCDEF"[c&0xF])
		}
	}
	return b.String()
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > 127 {
			return false
		}
	}
	return true
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]any{"error": msg})
}
