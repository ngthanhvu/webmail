package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	HTTPAddr     string
	AppSecret    string
	SessionHours int
	FrontendDist string

	IMAPMode       string // tls | starttls | plain
	IMAPHost       string
	IMAPPort       int
	IMAPSkipVerify bool

	SMTPMode       string // starttls | tls | plain
	SMTPHost       string
	SMTPPort       int
	SMTPUser       string
	SMTPPass       string
	SMTPFrom       string
	SMTPSkipVerify bool

	SentMailbox  string
	TrashMailbox string
}

func Load() *Config {
	return &Config{
		HTTPAddr:     getenv("HTTP_ADDR", ":8080"),
		AppSecret:    getenv("APP_SECRET", "dev-secret-change-me"),
		SessionHours: getint("SESSION_HOURS", 24),
		FrontendDist: getenv("FRONTEND_DIST", ""),

		IMAPMode:       strings.ToLower(getenv("IMAP_MODE", "tls")),
		IMAPHost:       getenv("IMAP_HOST", "localhost"),
		IMAPPort:       getint("IMAP_PORT", 993),
		IMAPSkipVerify: getbool("IMAP_SKIP_VERIFY", false),

		SMTPMode:       strings.ToLower(getenv("SMTP_MODE", "starttls")),
		SMTPHost:       getenv("SMTP_HOST", "localhost"),
		SMTPPort:       getint("SMTP_PORT", 587),
		SMTPUser:       getenv("SMTP_USER", ""),
		SMTPPass:       getenv("SMTP_PASS", ""),
		SMTPFrom:       getenv("SMTP_FROM", ""),
		SMTPSkipVerify: getbool("SMTP_SKIP_VERIFY", false),

		SentMailbox:  getenv("MAILBOX_SENT", "Sent"),
		TrashMailbox: getenv("MAILBOX_TRASH", "Trash"),
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getint(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getbool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}
