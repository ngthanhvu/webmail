package smtpx

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"webmail/backend/internal/config"
)

// plainAuth implements AUTH PLAIN without net/smtp's client-side TLS
// restriction (implicit-TLS via tls.Dial is not detected by net/smtp).
type plainAuth struct {
	username, password string
}

func (a *plainAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	resp := []byte("\x00" + a.username + "\x00" + a.password)
	return "PLAIN", resp, nil
}

func (a *plainAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		return nil, errors.New("thách thức SMTP PLAIN không xác định")
	}
	return nil, nil
}

type loginAuth struct {
	username, password string
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte(a.username), nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}
	switch strings.ToLower(string(fromServer)) {
	case "username:":
		return []byte(a.username), nil
	case "password:":
		return []byte(a.password), nil
	default:
		return nil, errors.New("thách thức SMTP LOGIN không xác định")
	}
}

// selectAuth picks an AUTH mechanism advertised by the server.
func selectAuth(c *smtp.Client, username, password string) smtp.Auth {
	if ok, params := c.Extension("AUTH"); ok {
		for _, mech := range strings.Fields(strings.ToUpper(params)) {
			if i := strings.IndexByte(mech, '='); i >= 0 {
				mech = mech[:i]
			}
			switch mech {
			case "PLAIN":
				return &plainAuth{username: username, password: password}
			case "LOGIN":
				return &loginAuth{username: username, password: password}
			}
		}
	}
	return &plainAuth{username: username, password: password}
}

// Send delivers msg (raw RFC 822) to all recipients over SMTP.
func Send(cfg *config.Config, username, password, from string, recipients []string, msg []byte) error {
	if len(recipients) == 0 {
		return fmt.Errorf("thiếu người nhận")
	}
	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)
	tlsConf := &tls.Config{ServerName: cfg.SMTPHost, InsecureSkipVerify: cfg.SMTPSkipVerify}

	var c *smtp.Client
	var err error
	if cfg.SMTPMode == "tls" {
		conn, dialErr := tls.Dial("tcp", addr, tlsConf)
		if dialErr != nil {
			return fmt.Errorf("kết nối SMTP: %w", dialErr)
		}
		c, err = smtp.NewClient(conn, cfg.SMTPHost)
	} else {
		c, err = smtp.Dial(addr)
	}
	if err != nil {
		return fmt.Errorf("kết nối SMTP: %w", err)
	}
	defer c.Close()

	if cfg.SMTPMode == "starttls" {
		if ok, _ := c.Extension("STARTTLS"); ok {
			if err := c.StartTLS(tlsConf); err != nil {
				return fmt.Errorf("SMTP STARTTLS: %w", err)
			}
		}
	}

	authUser, authPass := cfg.SMTPUser, cfg.SMTPPass
	if authUser == "" {
		authUser, authPass = username, password
	}
	if authUser != "" {
		if err := c.Auth(selectAuth(c, authUser, authPass)); err != nil {
			return fmt.Errorf("xác thực SMTP: %w", err)
		}
	}

	if err := c.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL FROM: %w", err)
	}
	for _, rcpt := range recipients {
		if err := c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("SMTP RCPT TO %s: %w", rcpt, err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("gửi dữ liệu SMTP: %w", err)
	}
	if err := w.Close(); err != nil {
		return fmt.Errorf("đóng dữ liệu SMTP: %w", err)
	}
	return c.Quit()
}
