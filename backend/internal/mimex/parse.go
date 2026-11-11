package mimex

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"net/url"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/unicode"
)

const maxBodyBytes = 32 << 20 // 32MB

// MaxMessageBytes is the maximum raw message size fetched from IMAP.
const MaxMessageBytes = maxBodyBytes

type Addr struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Attachment struct {
	Section     string `json:"section"`
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	Size        int    `json:"size"`
}

type ParsedMessage struct {
	UID         uint32       `json:"uid"`
	Flags       []string     `json:"flags"`
	Subject     string       `json:"subject"`
	From        Addr         `json:"from"`
	To          []Addr       `json:"to"`
	Cc          []Addr       `json:"cc"`
	Date        time.Time    `json:"date"`
	MessageID   string       `json:"messageId"`
	InReplyTo   string       `json:"inReplyTo"`
	References  string       `json:"references"`
	Text        string       `json:"text"`
	HTML        string       `json:"html"`
	Attachments []Attachment `json:"attachments"`
}

type PartContent struct {
	Filename    string
	ContentType string
	Data        []byte
}

func CharsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "us-ascii", "utf-8", "utf8", "ascii", "":
		return input, nil
	case "iso-8859-1", "latin1":
		return charmap.ISO8859_1.NewDecoder().Reader(input), nil
	case "iso-8859-15", "latin9":
		return charmap.ISO8859_15.NewDecoder().Reader(input), nil
	case "windows-1252", "cp1252":
		return charmap.Windows1252.NewDecoder().Reader(input), nil
	case "windows-1250", "cp1250":
		return charmap.Windows1250.NewDecoder().Reader(input), nil
	case "windows-1251", "cp1251":
		return charmap.Windows1251.NewDecoder().Reader(input), nil
	case "windows-1258", "cp1258":
		return charmap.Windows1258.NewDecoder().Reader(input), nil
	case "utf-16", "utf-16le", "unicode":
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder().Reader(input), nil
	default:
		return nil, fmt.Errorf("charset không hỗ trợ: %q", charset)
	}
}

var wordDecoder = &mime.WordDecoder{CharsetReader: CharsetReader}

// DecodeHeaderWords decodes RFC 2047 encoded-words (=?utf-8?B?...?=).
func DecodeHeaderWords(s string) string {
	if !strings.Contains(s, "=?") {
		return s
	}
	if out, err := wordDecoder.DecodeHeader(s); err == nil {
		return out
	}
	return s
}

// Parse parses a raw RFC 822 message into a display-friendly structure.
func Parse(raw []byte) (*ParsedMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("thư rỗng")
	}
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("lỗi phân tích thư: %w", err)
	}
	out := &ParsedMessage{
		To:          []Addr{},
		Cc:          []Addr{},
		Attachments: []Attachment{},
	}
	h := msg.Header
	out.Subject = DecodeHeaderWords(h.Get("Subject"))
	out.MessageID = strings.TrimSpace(h.Get("Message-Id"))
	out.InReplyTo = strings.TrimSpace(h.Get("In-Reply-To"))
	out.References = strings.TrimSpace(h.Get("References"))
	if date, err := h.Date(); err == nil {
		out.Date = date
	}
	if from := parseAddrList(h, "From"); len(from) > 0 {
		out.From = from[0]
	}
	out.To = parseAddrList(h, "To")
	out.Cc = parseAddrList(h, "Cc")

	body, _ := io.ReadAll(io.LimitReader(msg.Body, maxBodyBytes))
	walkContent(textproto.MIMEHeader(h), body, "", out)
	return out, nil
}

func parseAddrList(h mail.Header, key string) []Addr {
	raw := strings.TrimSpace(h.Get(key))
	if raw == "" {
		return nil
	}
	list, err := h.AddressList(key)
	if err != nil {
		var out []Addr
		for _, s := range strings.Split(raw, ",") {
			s = strings.TrimSpace(DecodeHeaderWords(s))
			if s == "" {
				continue
			}
			if i := strings.Index(s, "<"); i >= 0 && strings.HasSuffix(s, ">") {
				name := strings.TrimSpace(strings.Trim(s[:i], "\""))
				email := strings.Trim(s[i:], "<> ")
				out = append(out, Addr{Name: name, Email: email})
			} else {
				out = append(out, Addr{Email: s})
			}
		}
		return out
	}
	var out []Addr
	for _, a := range list {
		out = append(out, Addr{Name: DecodeHeaderWords(a.Name), Email: a.Address})
	}
	return out
}

// walkContent recursively walks a MIME entity. Section numbers follow IMAP
// part addressing ("1", "2", "2.1", ...) so attachments can be re-fetched
// via BODY.PEEK[<section>].
func walkContent(header textproto.MIMEHeader, body []byte, section string, out *ParsedMessage) {
	mediaType := "text/plain"
	var params map[string]string
	if ct := header.Get("Content-Type"); ct != "" {
		if mt, p, err := mime.ParseMediaType(ct); err == nil {
			mediaType, params = mt, p
		}
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return
		}
		mr := multipart.NewReader(bytes.NewReader(body), boundary)
		for i := 1; ; i++ {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			partBody, err := io.ReadAll(io.LimitReader(p, maxBodyBytes))
			p.Close()
			if err != nil {
				break
			}
			sec := strconv.Itoa(i)
			if section != "" {
				sec = section + "." + sec
			}
			walkContent(p.Header, partBody, sec, out)
		}
		return
	}

	decoded := decodeCTE(body, header.Get("Content-Transfer-Encoding"))

	if mediaType == "message/rfc822" || mediaType == "message/global" {
		sub, err := mail.ReadMessage(bytes.NewReader(decoded))
		if err != nil {
			return
		}
		subBody, _ := io.ReadAll(io.LimitReader(sub.Body, maxBodyBytes))
		walkContent(textproto.MIMEHeader(sub.Header), subBody, section, out)
		return
	}

	filename := partFilename(header, params)
	disposition := strings.ToLower(header.Get("Content-Disposition"))
	if filename != "" || strings.Contains(disposition, "attachment") {
		if filename == "" {
			filename = "part-" + section
		}
		out.Attachments = append(out.Attachments, Attachment{
			Section:     section,
			Filename:    filename,
			ContentType: mediaType,
			Size:        len(decoded),
		})
		return
	}

	switch {
	case strings.HasPrefix(mediaType, "text/plain"):
		if out.Text == "" {
			out.Text = string(decoded)
		} else {
			out.Text += "\n" + string(decoded)
		}
	case strings.HasPrefix(mediaType, "text/html"):
		if out.HTML == "" {
			out.HTML = string(decoded)
		}
	case header.Get("Content-Id") != "":
		// inline part referenced by CID, offer it as attachment
		out.Attachments = append(out.Attachments, Attachment{
			Section:     section,
			Filename:    filename,
			ContentType: mediaType,
			Size:        len(decoded),
		})
	}
}

// ParsePart parses a single fetched MIME part (headers + body) for download.
func ParsePart(raw []byte) (PartContent, error) {
	pc := PartContent{ContentType: "application/octet-stream"}
	if len(raw) == 0 {
		return pc, fmt.Errorf("phần đính kèm rỗng")
	}
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		// Some servers return only the part content (no MIME header).
		// Fallback: treat the whole literal as encoded content.
		pc.Data = heuristicDecode(raw)
		return pc, nil
	}
	h := msg.Header
	mediaType := "application/octet-stream"
	var params map[string]string
	if mt, p, err := mime.ParseMediaType(h.Get("Content-Type")); err == nil {
		mediaType, params = mt, p
	}
	body, _ := io.ReadAll(io.LimitReader(msg.Body, maxBodyBytes))
	pc.ContentType = mediaType
	pc.Filename = partFilename(textproto.MIMEHeader(h), params)
	pc.Data = decodeCTE(body, h.Get("Content-Transfer-Encoding"))
	return pc, nil
}

// heuristicDecode guesses the transfer encoding of a header-less part.
func heuristicDecode(raw []byte) []byte {
	compact := make([]byte, 0, len(raw))
	for _, c := range raw {
		switch c {
		case '\r', '\n', ' ', '\t':
		default:
			compact = append(compact, c)
		}
	}
	if len(compact) < 16 || len(compact)%4 != 0 {
		return raw
	}
	for _, c := range compact {
		if !(c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '+' || c == '/' || c == '=') {
			return raw
		}
	}
	if out, err := base64.StdEncoding.DecodeString(string(compact)); err == nil && len(out) > 0 {
		return out
	}
	return raw
}

func decodeCTE(b []byte, enc string) []byte {	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "base64":
		cleaned := make([]byte, 0, len(b))
		for _, c := range b {
			switch c {
			case '\r', '\n', ' ', '\t':
			default:
				cleaned = append(cleaned, c)
			}
		}
		for len(cleaned)%4 != 0 {
			cleaned = append(cleaned, '=')
		}
		if out, err := base64.StdEncoding.DecodeString(string(cleaned)); err == nil {
			return out
		}
		return b
	case "quoted-printable":
		if out, err := io.ReadAll(quotedprintable.NewReader(bytes.NewReader(b))); err == nil {
			return out
		}
		return b
	default:
		return b
	}
}

func partFilename(header textproto.MIMEHeader, ctParams map[string]string) string {
	if cd := header.Get("Content-Disposition"); cd != "" {
		if _, params, err := mime.ParseMediaType(cd); err == nil {
			if f := extendedParam(params, "filename"); f != "" {
				return DecodeHeaderWords(f)
			}
		}
	}
	if f, ok := ctParams["name"]; ok && f != "" {
		return DecodeHeaderWords(f)
	}
	return ""
}

// extendedParam resolves filename / filename* (RFC 2231) parameters.
func extendedParam(params map[string]string, key string) string {
	if v, ok := params[key]; ok && v != "" {
		return v
	}
	if v, ok := params[key+"*"]; ok && v != "" {
		if i := strings.Index(v, "''"); i >= 0 {
			if dec, err := url.PathUnescape(v[i+2:]); err == nil {
				return dec
			}
		}
		if dec, err := url.PathUnescape(v); err == nil {
			return dec
		}
		return v
	}
	var sb strings.Builder
	for i := 0; ; i++ {
		k := fmt.Sprintf("%s*%d", key, i)
		v, ok := params[k]
		if !ok {
			break
		}
		if i == 0 {
			if j := strings.Index(v, "''"); j >= 0 {
				v = v[j+2:]
			}
		}
		sb.WriteString(v)
	}
	if sb.Len() == 0 {
		return ""
	}
	if dec, err := url.PathUnescape(sb.String()); err == nil {
		return dec
	}
	return sb.String()
}
