package alert

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
	"time"
)

//go:embed templates/email.html
var emailHTML string

var emailTemplate = template.Must(template.New("email").Parse(emailHTML))

func renderEmail(subject, body string) (string, error) {
	body = strings.ReplaceAll(strings.ReplaceAll(body, "\r\n", "\n"), "\r", "\n")
	var out bytes.Buffer
	err := emailTemplate.Execute(&out, struct {
		Subject    string
		Paragraphs []string
	}{subject, strings.Split(body, "\n\n")})
	return out.String(), err
}

// buildEmail includes both readable plain text and the styled HTML alternative.
func buildEmail(from, to, subject, body string) ([]byte, error) {
	for _, header := range []string{from, to, subject} {
		if strings.ContainsAny(header, "\r\n") {
			return nil, fmt.Errorf("email headers must not contain line breaks")
		}
	}
	sender, err := mail.ParseAddress(from)
	if err != nil {
		return nil, fmt.Errorf("invalid email sender: %w", err)
	}
	recipient, err := mail.ParseAddress(to)
	if err != nil {
		return nil, fmt.Errorf("invalid email recipient: %w", err)
	}
	html, err := renderEmail(subject, body)
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	parts := multipart.NewWriter(&out)
	fmt.Fprintf(&out, "From: %s\r\nTo: %s\r\nSubject: %s\r\nDate: %s\r\nMIME-Version: 1.0\r\nContent-Type: multipart/alternative; boundary=%q\r\n\r\n",
		sender.String(), recipient.String(), mime.QEncoding.Encode("UTF-8", subject), time.Now().UTC().Format(time.RFC1123Z), parts.Boundary())
	for _, alternative := range []struct{ contentType, content string }{
		{"text/plain; charset=UTF-8", body},
		{"text/html; charset=UTF-8", html},
	} {
		header := textproto.MIMEHeader{}
		header.Set("Content-Type", alternative.contentType)
		header.Set("Content-Transfer-Encoding", "quoted-printable")
		part, err := parts.CreatePart(header)
		if err != nil {
			return nil, err
		}
		encoded := quotedprintable.NewWriter(part)
		if _, err := encoded.Write([]byte(alternative.content)); err != nil {
			return nil, err
		}
		if err := encoded.Close(); err != nil {
			return nil, err
		}
	}
	if err := parts.Close(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
