// Package mail delivers application email through SMTP.
package mail

import (
	"bytes"
	"context"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/http"
	stdmail "net/mail"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"

	"courses/platform/config"
)

// Message is a multipart/alternative email. Both bodies are provided so mail
// clients can choose the representation they support.
type Message struct {
	To       string
	Subject  string
	TextBody string
	HTMLBody string
}

// Sender sends mail through an unauthenticated SMTP endpoint.
type Sender struct {
	endpoint string
	host     string
	from     string
}

// NewSender builds a Sender from platform configuration.
func NewSender(mailConfig config.SMTPConfig) Sender {
	return Sender{
		endpoint: net.JoinHostPort(mailConfig.Host, mailConfig.Port),
		host:     mailConfig.Host,
		from:     mailConfig.From,
	}
}

// Send encodes and delivers msg. The SMTP connection respects ctx cancellation
// and deadlines even though net/smtp itself has no context-aware API.
func (s Sender) Send(ctx context.Context, msg Message) error {
	raw, err := encodeMessage(s.from, msg)
	if err != nil {
		return fmt.Errorf("encode email: %w", err)
	}
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	dialer := net.Dialer{}
	connection, err := dialer.DialContext(ctx, "tcp", s.endpoint)
	if err != nil {
		return fmt.Errorf("dial SMTP server: %w", err)
	}
	defer connection.Close()
	stopCancelDeadline := context.AfterFunc(ctx, func() {
		_ = connection.SetDeadline(time.Now())
	})
	defer stopCancelDeadline()
	if deadline, ok := ctx.Deadline(); ok {
		if err := connection.SetDeadline(deadline); err != nil {
			return fmt.Errorf("set SMTP deadline: %w", err)
		}
	}

	client, err := smtp.NewClient(connection, s.host)
	if err != nil {
		return fmt.Errorf("start SMTP client: %w", err)
	}
	defer client.Close()
	if err := client.Mail(s.from); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}
	if err := client.Rcpt(msg.To); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP data: %w", err)
	}
	if _, err := writer.Write(raw); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP data: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP data: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("quit SMTP client: %w", err)
	}
	return nil
}

func encodeMessage(from string, msg Message) ([]byte, error) {
	if err := validAddress("MAIL_FROM", from); err != nil {
		return nil, err
	}
	if err := validAddress("message recipient", msg.To); err != nil {
		return nil, err
	}
	if strings.ContainsAny(msg.Subject, "\r\n") {
		return nil, fmt.Errorf("message subject must not contain a newline")
	}

	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)
	for _, part := range []struct {
		contentType string
		body        string
	}{
		{contentType: "text/plain; charset=UTF-8", body: msg.TextBody},
		{contentType: "text/html; charset=UTF-8", body: msg.HTMLBody},
	} {
		header := textproto.MIMEHeader{}
		header.Set("Content-Type", part.contentType)
		header.Set("Content-Transfer-Encoding", "quoted-printable")
		writer, err := multipartWriter.CreatePart(header)
		if err != nil {
			return nil, fmt.Errorf("create MIME part: %w", err)
		}
		quotedPrintable := quotedprintable.NewWriter(writer)
		if _, err := quotedPrintable.Write([]byte(part.body)); err != nil {
			return nil, fmt.Errorf("encode MIME part: %w", err)
		}
		if err := quotedPrintable.Close(); err != nil {
			return nil, fmt.Errorf("finish MIME part: %w", err)
		}
	}
	if err := multipartWriter.Close(); err != nil {
		return nil, fmt.Errorf("finish MIME message: %w", err)
	}

	header := textproto.MIMEHeader{}
	header.Set("From", from)
	header.Set("To", msg.To)
	header.Set("Subject", mime.QEncoding.Encode("UTF-8", msg.Subject))
	header.Set("Date", time.Now().UTC().Format(time.RFC1123Z))
	header.Set("MIME-Version", "1.0")
	header.Set("Content-Type", mime.FormatMediaType("multipart/alternative", map[string]string{"boundary": multipartWriter.Boundary()}))
	var encoded bytes.Buffer
	if err := http.Header(header).Write(&encoded); err != nil {
		return nil, fmt.Errorf("write MIME headers: %w", err)
	}
	encoded.WriteString("\r\n")
	encoded.Write(body.Bytes())
	return encoded.Bytes(), nil
}

func validAddress(label, value string) error {
	if strings.ContainsAny(value, "\r\n") {
		return fmt.Errorf("%s must not contain a newline", label)
	}
	parsed, err := stdmail.ParseAddress(value)
	if err != nil || parsed.Address != value {
		return fmt.Errorf("%s must be an email address", label)
	}
	return nil
}
