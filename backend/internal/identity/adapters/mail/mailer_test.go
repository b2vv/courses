package mail_test

import (
	"context"
	"strings"
	"testing"

	identitymail "courses/internal/identity/adapters/mail"
	platformmail "courses/platform/mail"
)

type sender struct{ message platformmail.Message }

func (sender *sender) Send(_ context.Context, message platformmail.Message) error {
	sender.message = message
	return nil
}

func TestMailerRendersLocalizedVerificationMessagesAndEscapesHTML(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, language, subject, text string
	}{
		{name: "ukrainian", language: "uk", subject: "Підтвердьте", text: "Підтвердьте"},
		{name: "english", language: "en", subject: "Verify", text: "Verify"},
		{name: "defaults to ukrainian", language: "other", subject: "Підтвердьте", text: "Підтвердьте"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			delivery := &sender{}
			mailer, err := identitymail.New(delivery, "https://courses.example")
			if err != nil {
				t.Fatal(err)
			}
			if err := mailer.SendVerification(context.Background(), "student@example.test", "a&b", test.language); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(delivery.message.Subject, test.subject) || !strings.Contains(delivery.message.TextBody, test.text) {
				t.Fatalf("message is not localized: %+v", delivery.message)
			}
			if !strings.Contains(delivery.message.TextBody, "https://courses.example/verify-email?token=a%26b") {
				t.Fatalf("unsafe or missing text link: %q", delivery.message.TextBody)
			}
			if strings.Contains(delivery.message.HTMLBody, "a&b") || !strings.Contains(delivery.message.HTMLBody, "a%26b") {
				t.Fatalf("token is not URL-escaped in HTML: %q", delivery.message.HTMLBody)
			}
		})
	}
}

func TestMailerRendersEveryAccountMessage(t *testing.T) {
	t.Parallel()
	delivery := &sender{}
	mailer, err := identitymail.New(delivery, "https://courses.example")
	if err != nil {
		t.Fatal(err)
	}
	for _, send := range []func(context.Context, string, string, string) error{mailer.SendPasswordReset, mailer.SendEmailChange} {
		if err := send(context.Background(), "student@example.test", "token", "en"); err != nil {
			t.Fatal(err)
		}
		if delivery.message.Subject == "" || delivery.message.TextBody == "" || delivery.message.HTMLBody == "" {
			t.Fatalf("incomplete message: %+v", delivery.message)
		}
	}
}
