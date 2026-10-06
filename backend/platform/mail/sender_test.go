package mail

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	stdmail "net/mail"
	"strings"
	"testing"
)

func TestEncodeMessageBuildsUTF8MultipartAlternative(t *testing.T) {
	t.Parallel()

	raw, err := encodeMessage("no-reply@courseforge.local", Message{
		To:       "student@example.com",
		Subject:  "Підтвердіть email",
		TextBody: "Вітаємо, студенте!",
		HTMLBody: "<p>Вітаємо, <strong>студенте</strong>!</p>",
	})
	if err != nil {
		t.Fatal(err)
	}

	message, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	if got := message.Header.Get("To"); got != "student@example.com" {
		t.Fatalf("To = %q", got)
	}
	if subject := message.Header.Get("Subject"); !strings.Contains(subject, "=?UTF-8?") {
		t.Fatalf("Subject %q is not RFC 2047 encoded", subject)
	}
	mediaType, params, err := mime.ParseMediaType(message.Header.Get("Content-Type"))
	if err != nil || mediaType != "multipart/alternative" {
		t.Fatalf("Content-Type = %q, %v", message.Header.Get("Content-Type"), err)
	}
	reader := multipart.NewReader(message.Body, params["boundary"])
	for _, want := range []struct{ mediaType, body string }{
		{"text/plain", "Вітаємо, студенте!"},
		{"text/html", "<p>Вітаємо, <strong>студенте</strong>!</p>"},
	} {
		part, err := reader.NextPart()
		if err != nil {
			t.Fatal(err)
		}
		mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil || mediaType != want.mediaType {
			t.Fatalf("part Content-Type = %q, %v", part.Header.Get("Content-Type"), err)
		}
		body, err := io.ReadAll(part)
		if err != nil || string(body) != want.body {
			t.Fatalf("part body = %q, %v", body, err)
		}
	}
	if _, err := reader.NextPart(); err != io.EOF {
		t.Fatalf("third part error = %v, want EOF", err)
	}
}

func TestEncodeMessageRejectsHeaderInjectionAndInvalidAddresses(t *testing.T) {
	t.Parallel()

	for name, message := range map[string]Message{
		"missing recipient": {Subject: "subject", TextBody: "text", HTMLBody: "<p>html</p>"},
		"newline recipient": {To: "student@example.com\r\nBcc: leak@example.com", Subject: "subject", TextBody: "text", HTMLBody: "<p>html</p>"},
		"newline subject":   {To: "student@example.com", Subject: "subject\r\nBcc: leak@example.com", TextBody: "text", HTMLBody: "<p>html</p>"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := encodeMessage("no-reply@courseforge.local", message); err == nil {
				t.Fatal("encodeMessage() succeeded")
			}
		})
	}
}
