// Package mail renders identity emails and hands them to the platform SMTP sender.
package mail

import (
	"bytes"
	"context"
	"fmt"
	htmltemplate "html/template"
	"io"
	"net/url"
	"strings"
	texttemplate "text/template"

	"courses/internal/identity/ports"
	platformmail "courses/platform/mail"
)

type Sender interface {
	Send(context.Context, platformmail.Message) error
}

type Mailer struct {
	sender Sender
	origin string
}

var _ ports.Mailer = (*Mailer)(nil)

func New(sender Sender, publicOrigin string) (*Mailer, error) {
	origin, err := validOrigin(publicOrigin)
	if err != nil {
		return nil, err
	}
	return &Mailer{sender: sender, origin: origin}, nil
}

func (mailer *Mailer) SendVerification(ctx context.Context, recipient, token, lang string) error {
	return mailer.send(ctx, recipient, token, lang, verification)
}

func (mailer *Mailer) SendPasswordReset(ctx context.Context, recipient, token, lang string) error {
	return mailer.send(ctx, recipient, token, lang, passwordReset)
}

func (mailer *Mailer) SendEmailChange(ctx context.Context, recipient, token, lang string) error {
	return mailer.send(ctx, recipient, token, lang, emailChange)
}

type messageKind struct{ path, ukSubject, enSubject, ukText, enText string }

var (
	verification  = messageKind{"/verify-email", "Підтвердьте електронну пошту", "Verify your email", "Підтвердьте електронну пошту", "Verify your email"}
	passwordReset = messageKind{"/password/reset", "Скиньте пароль", "Reset your password", "Скиньте пароль", "Reset your password"}
	emailChange   = messageKind{"/email/change/confirm", "Підтвердьте зміну електронної пошти", "Confirm your email change", "Підтвердьте зміну електронної пошти", "Confirm your email change"}
)

func (mailer *Mailer) send(ctx context.Context, recipient, token, lang string, kind messageKind) error {
	data := struct{ Heading, Link string }{Heading: kind.ukText, Link: mailer.link(kind.path, token)}
	subject := kind.ukSubject
	if lang == "en" {
		data.Heading, subject = kind.enText, kind.enSubject
	}
	text, err := renderText(data)
	if err != nil {
		return err
	}
	html, err := renderHTML(data)
	if err != nil {
		return err
	}
	return mailer.sender.Send(ctx, platformmail.Message{To: recipient, Subject: subject, TextBody: text, HTMLBody: html})
}

func (mailer *Mailer) link(path, token string) string {
	return mailer.origin + path + "?token=" + url.QueryEscape(token)
}

func renderText(data any) (string, error) {
	return render(texttemplate.Must(texttemplate.New("email").Parse("{{.Heading}}\n\n{{.Link}}\n")), data)
}

func renderHTML(data any) (string, error) {
	return render(htmltemplate.Must(htmltemplate.New("email").Parse("<p>{{.Heading}}</p><p><a href=\"{{.Link}}\">{{.Link}}</a></p>")), data)
}

type executor interface {
	Execute(io.Writer, any) error
}

func render(template executor, data any) (string, error) {
	var result bytes.Buffer
	if err := template.Execute(&result, data); err != nil {
		return "", fmt.Errorf("render email: %w", err)
	}
	return result.String(), nil
}

func validOrigin(value string) (string, error) {
	origin, err := url.Parse(value)
	if err != nil || origin.Scheme == "" || origin.Host == "" || origin.User != nil || origin.RawQuery != "" || origin.Fragment != "" || (origin.Path != "" && origin.Path != "/") || (origin.Scheme != "http" && origin.Scheme != "https") {
		return "", fmt.Errorf("PUBLIC_ORIGIN must be an absolute HTTP origin")
	}
	return strings.TrimSuffix(value, "/"), nil
}
