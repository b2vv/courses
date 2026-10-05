package domain

import (
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	minimumPasswordLength                = 8
	maximumPasswordLength                = 128
	maximumRegistrationDisplayNameLength = 100

	EmailVerificationTokenLifetime = 24 * time.Hour
	PasswordResetTokenLifetime     = time.Hour
)

type Registration struct {
	Email       string
	Password    string
	DisplayName string
}

func NewRegistration(registration Registration) (Registration, error) {
	email, err := NormalizeEmail(registration.Email)
	if err != nil {
		return Registration{}, err
	}
	if !validPassword(registration.Password) {
		return Registration{}, ErrWeakPassword
	}
	if !validRegistrationDisplayName(registration.DisplayName) {
		return Registration{}, ErrInvalidDisplayName
	}
	registration.Email = email
	return registration, nil
}

type AccountTokenPurpose string

const (
	EmailVerificationToken AccountTokenPurpose = "email_verify"
	PasswordResetToken     AccountTokenPurpose = "password_reset"
	EmailChangeToken       AccountTokenPurpose = "email_change"
)

func (purpose AccountTokenPurpose) Valid() bool {
	return purpose == EmailVerificationToken || purpose == PasswordResetToken || purpose == EmailChangeToken
}

type AccountToken struct {
	ID         string
	UserID     string
	Purpose    AccountTokenPurpose
	ExpiresAt  time.Time
	ConsumedAt *time.Time
	CreatedAt  time.Time
}

func NewAccountToken(token AccountToken) (AccountToken, error) {
	if !validID(token.ID) || !validID(token.UserID) || !token.Purpose.Valid() || token.CreatedAt.IsZero() || !token.ExpiresAt.After(token.CreatedAt) {
		return AccountToken{}, ErrInvalidAccountToken
	}
	if token.ConsumedAt != nil && token.ConsumedAt.Before(token.CreatedAt) {
		return AccountToken{}, ErrInvalidAccountToken
	}
	return token, nil
}

func (token AccountToken) UsableAt(now time.Time) bool {
	return token.ConsumedAt == nil && !now.Before(token.CreatedAt) && now.Before(token.ExpiresAt)
}

func ValidatePasswordChange(current, next string) error {
	if current == next {
		return ErrPasswordUnchanged
	}
	if !validPassword(next) {
		return ErrWeakPassword
	}
	return nil
}

func ValidateEmailChange(email string) (string, error) {
	return NormalizeEmail(email)
}

func validPassword(value string) bool {
	return utf8.ValidString(value) && utf8.RuneCountInString(value) >= minimumPasswordLength && utf8.RuneCountInString(value) <= maximumPasswordLength
}

func validRegistrationDisplayName(value string) bool {
	if !utf8.ValidString(value) || value == "" || strings.TrimSpace(value) != value || utf8.RuneCountInString(value) > maximumRegistrationDisplayNameLength {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
