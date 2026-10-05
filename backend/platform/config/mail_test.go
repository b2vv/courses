package config_test

import (
	"strings"
	"testing"

	"courses/platform/config"
)

func TestMailConfig(t *testing.T) {
	t.Parallel()

	config, err := config.MailConfig(environment(map[string]string{
		"SMTP_HOST": "mailpit",
		"MAIL_FROM": "no-reply@courseforge.local",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if config.Host != "mailpit" || config.Port != "1025" || config.From != "no-reply@courseforge.local" {
		t.Fatalf("MailConfig() = %#v", config)
	}
}

func TestMailConfigRejectsInvalidEnvironment(t *testing.T) {
	t.Parallel()

	valid := map[string]string{
		"SMTP_HOST": "mailpit",
		"SMTP_PORT": "1025",
		"MAIL_FROM": "no-reply@courseforge.local",
	}
	for name, override := range map[string]map[string]string{
		"missing host":      {"SMTP_HOST": ""},
		"port is not TCP":   {"SMTP_PORT": "smtp"},
		"port is zero":      {"SMTP_PORT": "0"},
		"missing from":      {"MAIL_FROM": ""},
		"invalid from":      {"MAIL_FROM": "not-an-email"},
		"display name only": {"MAIL_FROM": "Courseforge <no-reply@courseforge.local>"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			values := map[string]string{}
			for key, value := range valid {
				values[key] = value
			}
			for key, value := range override {
				values[key] = value
			}
			_, err := config.MailConfig(environment(values))
			if err == nil {
				t.Fatal("MailConfig() succeeded")
			}
			if !strings.Contains(err.Error(), "SMTP_") && !strings.Contains(err.Error(), "MAIL_FROM") {
				t.Fatalf("error %q does not identify environment input", err)
			}
		})
	}
}
