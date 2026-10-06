package domain

import (
	"encoding/hex"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type Role string

const (
	Student    Role = "student"
	Instructor Role = "instructor"
	Admin      Role = "admin"
)

func (role Role) Valid() bool {
	return role == Student || role == Instructor || role == Admin
}

func (role Role) CanLearn() bool {
	return role.Valid()
}

func (role Role) CanAccessAdministration() bool {
	return role == Instructor || role == Admin
}

func (role Role) CanAdministerAccounts() bool {
	return role == Admin
}

type User struct {
	ID               string
	Email            string
	DisplayName      string
	Role             Role
	EmailVerified    bool
	TwoFactorEnabled bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
	DisabledAt       *time.Time
	DeletedAt        *time.Time
}

func NewUser(user User) (User, error) {
	email, err := NormalizeEmail(user.Email)
	if err != nil || email != user.Email || !validID(user.ID) || !validDisplayName(user.DisplayName) || !user.Role.Valid() {
		return User{}, ErrInvalidUser
	}
	if user.CreatedAt.IsZero() || user.UpdatedAt.Before(user.CreatedAt) {
		return User{}, ErrInvalidUser
	}
	if user.DisabledAt != nil && (user.DisabledAt.Before(user.CreatedAt) || user.DisabledAt.After(user.UpdatedAt)) {
		return User{}, ErrInvalidUser
	}
	if user.DeletedAt != nil && (user.DeletedAt.Before(user.CreatedAt) || user.DeletedAt.After(user.UpdatedAt) || user.DisabledAt == nil) {
		return User{}, ErrInvalidUser
	}
	return user, nil
}

func (user User) Active() bool {
	return user.DisabledAt == nil && user.DeletedAt == nil
}

func NormalizeEmail(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", ErrInvalidEmail
	}
	normalized := strings.ToLower(strings.TrimSpace(value))
	if len(normalized) < 3 || len(normalized) > 254 || strings.Count(normalized, "@") != 1 {
		return "", ErrInvalidEmail
	}
	local, host, found := strings.Cut(normalized, "@")
	if !found || local == "" || host == "" {
		return "", ErrInvalidEmail
	}
	for _, character := range normalized {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return "", ErrInvalidEmail
		}
	}
	return normalized, nil
}

func validDisplayName(value string) bool {
	if !utf8.ValidString(value) || value == "" || strings.TrimSpace(value) != value || utf8.RuneCountInString(value) > 120 {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}
