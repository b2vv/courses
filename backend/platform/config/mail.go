package config

import (
	"fmt"
	stdmail "net/mail"
)

// SMTPConfig is the SMTP delivery endpoint. Authentication fields will be added
// only when the deployment has an SMTP provider that needs them.
type SMTPConfig struct {
	Host string
	Port string
	From string
}

// MailConfig reads the unauthenticated SMTP configuration used by local Mailpit.
func MailConfig(getenv Getenv) (SMTPConfig, error) {
	host, port, err := endpoint(getenv, "SMTP_HOST", "SMTP_PORT", "1025")
	if err != nil {
		return SMTPConfig{}, err
	}
	from := getenv("MAIL_FROM")
	parsed, err := stdmail.ParseAddress(from)
	if err != nil || parsed.Address != from {
		return SMTPConfig{}, fmt.Errorf("MAIL_FROM must be an email address")
	}
	return SMTPConfig{Host: host, Port: port, From: from}, nil
}
