package mimex

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"regexp"
	"strings"
	"time"
)

func nowRFC2822() string {
	return time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")
}

func timeNow() time.Time {
	return time.Now()
}

var qpSnippetRe = regexp.MustCompile(`=[0-9A-F]{2}`)
var htmlTagRe = regexp.MustCompile(`(?s)<[^>]*>`)

// ExtractSnippet builds a short preview from a partial body fetch (best effort).
func ExtractSnippet(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	s := string(b)
	if strings.Contains(s, "=\n") || qpSnippetRe.MatchString(s) {
		if out, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(s))); err == nil && len(out) > 0 {
			s = string(out)
		}
	}
	low := strings.ToLower(s)
	if strings.Contains(low, "<html") || strings.Contains(low, "<div") || strings.Contains(low, "<br") || strings.Contains(low, "<p") {
		s = htmlTagRe.ReplaceAllString(s, " ")
		s = decodeBasicEntities(s)
	}
	s = strings.Join(strings.Fields(s), " ")
	runes := []rune(s)
	if len(runes) > 140 {
		runes = runes[:140]
	}
	return string(runes)
}

func decodeBasicEntities(s string) string {
	repl := strings.NewReplacer(
		"&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&quot;", `"`, "&#39;", "'",
	)
	return repl.Replace(s)
}

// ---------- outgoing message building ----------

type OutgoingAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Outgoing struct {
	FromName  string
	FromEmail string
	To        []string
	Cc        []string
	Bcc       []string
	Subject   string
	Text      string
	HTML      string
	InReplyTo string
	Refs      string
	Attachments []OutgoingAttachment
}

type mimePart struct {
	headers []string // full header lines "Key: value"
	data    []byte
}

// Build renders the outgoing message as raw RFC 822 bytes.
func Build(o Outgoing) ([]byte, error) {
	var body []byte
	var contentType string
	boundary := ""

	attachmentParts := make([]mimePart, 0, len(o.Attachments))
	for _, a := range o.Attachments {
		ct := a.ContentType
		if ct == "" {
			ct = "application/octet-stream"
		}
		attachmentParts = append(attachmentParts, mimePart{
			headers: []string{
				"Content-Type: " + ct,
				"Content-Transfer-Encoding: base64",
				"Content-Disposition: attachment; filename=" + quoteFilename(a.Filename),
			},
			data: []byte(wrap76(base64.StdEncoding.EncodeToString(a.Data))),
		})
	}

	plain := mimePart{
		headers: []string{
			"Content-Type: text/plain; charset=utf-8",
			"Content-Transfer-Encoding: base64",
		},
		data: []byte(wrap76(base64.StdEncoding.EncodeToString([]byte(o.Text)))),
	}

	switch {
	case o.HTML == "" && len(attachmentParts) == 0:
		contentType = "text/plain; charset=utf-8"
		body = plain.data
	case o.HTML == "":
		boundary = randomBoundary()
		contentType = "multipart/mixed; boundary=\"" + boundary + "\""
		body = multipartBody(boundary, append([]mimePart{plain}, attachmentParts...))
	case len(attachmentParts) == 0:
		boundary = randomBoundary()
		contentType = "multipart/alternative; boundary=\"" + boundary + "\""
		html := mimePart{
			headers: []string{
				"Content-Type: text/html; charset=utf-8",
				"Content-Transfer-Encoding: base64",
			},
			data: []byte(wrap76(base64.StdEncoding.EncodeToString([]byte(o.HTML)))),
		}
		body = multipartBody(boundary, []mimePart{plain, html})
	default:
		altBoundary := randomBoundary()
		html := mimePart{
			headers: []string{
				"Content-Type: text/html; charset=utf-8",
				"Content-Transfer-Encoding: base64",
			},
			data: []byte(wrap76(base64.StdEncoding.EncodeToString([]byte(o.HTML)))),
		}
		alt := mimePart{
			headers: []string{"Content-Type: multipart/alternative; boundary=\"" + altBoundary + "\""},
			data:    multipartBody(altBoundary, []mimePart{plain, html}),
		}
		boundary = randomBoundary()
		contentType = "multipart/mixed; boundary=\"" + boundary + "\""
		body = multipartBody(boundary, append([]mimePart{alt}, attachmentParts...))
	}

	var msg bytes.Buffer
	writeHeader(&msg, "From", formatAddr(o.FromName, o.FromEmail))
	if len(o.To) > 0 {
		writeHeader(&msg, "To", strings.Join(o.To, ", "))
	}
	if len(o.Cc) > 0 {
		writeHeader(&msg, "Cc", strings.Join(o.Cc, ", "))
	}
	writeHeader(&msg, "Subject", encodeHeaderWord(o.Subject))
	writeHeader(&msg, "Date", nowRFC2822())
	writeHeader(&msg, "Message-Id", genMessageID(o.FromEmail))
	writeHeader(&msg, "MIME-Version", "1.0")
	if o.InReplyTo != "" {
		writeHeader(&msg, "In-Reply-To", o.InReplyTo)
	}
	if o.Refs != "" {
		writeHeader(&msg, "References", o.Refs)
	}
	writeHeader(&msg, "Content-Type", contentType)
	if boundary == "" {
		writeHeader(&msg, "Content-Transfer-Encoding", "base64")
	}
	msg.WriteString("\r\n")
	msg.Write(body)
	return msg.Bytes(), nil
}

func multipartBody(boundary string, parts []mimePart) []byte {
	var b bytes.Buffer
	for _, p := range parts {
		b.WriteString("--" + boundary + "\r\n")
		for _, h := range p.headers {
			b.WriteString(h)
			b.WriteString("\r\n")
		}
		b.WriteString("\r\n")
		b.Write(p.data)
		if !bytes.HasSuffix(p.data, []byte("\r\n")) {
			b.WriteString("\r\n")
		}
	}
	b.WriteString("--" + boundary + "--\r\n")
	return b.Bytes()
}

func writeHeader(w io.Writer, key, value string) {
	fmt.Fprintf(w, "%s: %s\r\n", key, value)
}

func wrap76(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(s); i += 76 {
		end := i + 76
		if end > len(s) {
			end = len(s)
		}
		b.WriteString(s[i:end])
		b.WriteString("\r\n")
	}
	return b.String()
}

func randomBoundary() string {
	buf := make([]byte, 16)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func genMessageID(fromEmail string) string {
	domain := "localhost"
	if addr, err := mail.ParseAddress(fromEmail); err == nil && strings.Contains(addr.Address, "@") {
		domain = addr.Address[strings.Index(addr.Address, "@")+1:]
	}
	buf := make([]byte, 16)
	rand.Read(buf)
	return "<" + hex.EncodeToString(buf) + "." + fmt.Sprintf("%d", timeNow().Unix()) + "@" + domain + ">"
}

func quoteFilename(name string) string {
	name = strings.ReplaceAll(name, `"`, `'`)
	name = strings.ReplaceAll(name, "\r", "")
	name = strings.ReplaceAll(name, "\n", "")
	if name == "" {
		name = "attachment"
	}
	if isASCII(name) {
		return `"` + name + `"`
	}
	return mime.BEncoding.Encode("utf-8", name)
}

func formatAddr(name, email string) string {
	if name == "" {
		return email
	}
	return encodeHeaderWord(name) + " <" + email + ">"
}

func encodeHeaderWord(s string) string {
	if isASCII(s) {
		return s
	}
	return mime.BEncoding.Encode("utf-8", s)
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}
