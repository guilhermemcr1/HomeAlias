package alert

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEmailAlternatives(t *testing.T) {
	const subject = "Teste de conexão — HomeAlias"
	const body = "Olá, Guilherme!\n\nO envio de e-mail funcionou."
	raw, err := buildEmail("HomeAlias <alerts@example.com>", "Guilherme <user@example.com>", subject, body)
	if err != nil {
		t.Fatal(err)
	}
	message, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	decodedSubject, err := new(mime.WordDecoder).DecodeHeader(message.Header.Get("Subject"))
	if err != nil || decodedSubject != subject {
		t.Fatalf("subject did not round-trip: %q, %v", decodedSubject, err)
	}
	if message.Header.Get("From") != `"HomeAlias" <alerts@example.com>` {
		t.Fatalf("missing sender: %s", message.Header.Get("From"))
	}
	if _, err := time.Parse(time.RFC1123Z, message.Header.Get("Date")); err != nil {
		t.Fatalf("invalid sending date: %v", err)
	}
	mediaType, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/alternative" {
		t.Fatalf("invalid content type: %s, %v", mediaType, err)
	}
	reader := multipart.NewReader(message.Body, params["boundary"])
	for i, expected := range []string{"text/plain", "text/html"} {
		part, err := reader.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		kind, charset, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil || kind != expected || charset["charset"] != "UTF-8" {
			t.Fatalf("invalid part %d: %s", i, part.Header)
		}
		content, err := io.ReadAll(part)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 && strings.ReplaceAll(string(content), "\r\n", "\n") != body {
			t.Fatalf("plain text was changed: %q", content)
		}
		if i == 1 && (!bytes.Contains(content, []byte("Olá, Guilherme!")) || !bytes.Contains(content, []byte("<html lang=\"pt-BR\">"))) {
			t.Fatal("HTML alternative lost content or language")
		}
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatalf("multipart was not properly closed: %v", err)
	}
}

func TestEmailRejectsInvalidHeaders(t *testing.T) {
	for _, tc := range []struct{ from, to, subject string }{
		{"alerts@example.com\r\nBcc: intruder@example.com", "user@example.com", "Teste"},
		{"alerts@example.com", "user@example.com\nBcc: intruder@example.com", "Teste"},
		{"alerts@example.com", "user@example.com", "Teste\r\nBcc: intruder@example.com"},
		{"invalid", "user@example.com", "Teste"},
		{"alerts@example.com", "invalid", "Teste"},
	} {
		if _, err := buildEmail(tc.from, tc.to, tc.subject, "Teste"); err == nil {
			t.Fatal("invalid header accepted")
		}
	}
}

func TestEmailEscapesContent(t *testing.T) {
	html, err := renderEmail(`<img src=x onerror=alert(1)>`, "<script>alert(1)</script>\n\nNome & domínio")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(html, "<script>") || strings.Contains(html, "<img src=x") || !strings.Contains(html, "Nome &amp; domínio") {
		t.Fatal("dynamic content was not HTML-escaped")
	}
}

func TestEmailPreview(t *testing.T) {
	html, err := renderEmail("Teste do canal de e-mail", "Este é um teste de envio do HomeAlias.\n\nSe esta mensagem chegou até você, o envio para este endereço funcionou.\n\nVocê pode voltar à página Alertas para consultar o resultado do teste.")
	if err != nil {
		t.Fatal(err)
	}
	// Optional artifact for visual QA, generated from the production template.
	if dir := os.Getenv("HOMEALIAS_EMAIL_PREVIEW_DIR"); dir != "" {
		if err := os.WriteFile(filepath.Join(dir, "homealias-email-preview.html"), []byte(html), 0600); err != nil {
			t.Fatal(err)
		}
	}
}
