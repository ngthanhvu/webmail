package mimex

import (
	"mime"
	"strings"
	"testing"
)

func TestBuildAndParseTextOnly(t *testing.T) {
	raw, err := Build(Outgoing{
		FromName:  "Nguyễn Văn A",
		FromEmail: "a@example.com",
		To:        []string{"b@example.com"},
		Subject:   "Xin chào tiếng Việt",
		Text:      "Đây là nội dung thư test",
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Subject != "Xin chào tiếng Việt" {
		t.Errorf("subject = %q", msg.Subject)
	}
	if msg.Text != "Đây là nội dung thư test" {
		t.Errorf("text = %q", msg.Text)
	}
	if msg.From.Email != "a@example.com" {
		t.Errorf("from = %+v", msg.From)
	}
	if len(msg.To) != 1 || msg.To[0].Email != "b@example.com" {
		t.Errorf("to = %+v", msg.To)
	}
}

func TestBuildAndParseAlternativeWithAttachment(t *testing.T) {
	raw, err := Build(Outgoing{
		FromName:  "Sender",
		FromEmail: "s@example.com",
		To:        []string{"r@example.com"},
		Subject:   "html + attachment",
		Text:      "plain version",
		HTML:      "<p>html version</p>",
		Attachments: []OutgoingAttachment{
			{Filename: "doc.txt", ContentType: "text/plain", Data: []byte("hello attachment")},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if msg.Text != "plain version" {
		t.Errorf("text = %q", msg.Text)
	}
	if !strings.Contains(msg.HTML, "html version") {
		t.Errorf("html = %q", msg.HTML)
	}
	if len(msg.Attachments) != 1 {
		t.Fatalf("attachments = %+v", msg.Attachments)
	}
	att := msg.Attachments[0]
	if att.Filename != "doc.txt" {
		t.Errorf("filename = %q", att.Filename)
	}
	if att.Section != "2" {
		t.Errorf("section = %q", att.Section)
	}
	if att.Size != len("hello attachment") {
		t.Errorf("size = %d", att.Size)
	}
}

func TestParsePart(t *testing.T) {
	raw := "Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Disposition: attachment; filename=\"demo.txt\"\r\n" +
		"Content-Transfer-Encoding: base64\r\n" +
		"\r\n" +
		"aGVsbG8gcGFydA==\r\n"
	pc, err := ParsePart([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if string(pc.Data) != "hello part" {
		t.Errorf("data = %q", pc.Data)
	}
	if pc.Filename != "demo.txt" {
		t.Errorf("filename = %q", pc.Filename)
	}
	if pc.ContentType != "text/plain" {
		t.Errorf("contentType = %q", pc.ContentType)
	}
}

func TestExtractSnippet(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"plain text here", "plain text here"},
		{"=C3=A0 text with qp", "à text with qp"},
		{"<p>Hello <b>World</b></p>", "Hello World"},
	}
	for _, tt := range tests {
		got := ExtractSnippet([]byte(tt.in))
		if got != tt.want {
			t.Errorf("ExtractSnippet(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestDecodeHeaderWords(t *testing.T) {
	encoded := mime.BEncoding.Encode("utf-8", "Xin chào")
	got := DecodeHeaderWords(encoded)
	if got != "Xin chào" {
		t.Errorf("got = %q", got)
	}
	if got := DecodeHeaderWords("plain"); got != "plain" {
		t.Errorf("plain = %q", got)
	}
}

func TestDecodeHeaderWordsLatin1(t *testing.T) {
	// "caffè" quoted-printable iso-8859-1
	got := DecodeHeaderWords("=?iso-8859-1?Q?caff=E8?=")
	if got != "caffè" {
		t.Errorf("got = %q", got)
	}
}
