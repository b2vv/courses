package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/identity/domain"
)

func TestNewRegistration(t *testing.T) {
	t.Parallel()
	valid := domain.Registration{Email: " Student@Example.Test ", Password: "password", DisplayName: "Student"}
	for _, test := range []struct {
		name    string
		change  func(*domain.Registration)
		wantErr error
	}{
		{name: "normalizes email", change: func(*domain.Registration) {}},
		{name: "minimum password", change: func(registration *domain.Registration) { registration.Password = strings.Repeat("p", 8) }},
		{name: "maximum password", change: func(registration *domain.Registration) { registration.Password = strings.Repeat("p", 128) }},
		{name: "minimum display name", change: func(registration *domain.Registration) { registration.DisplayName = "n" }},
		{name: "maximum display name", change: func(registration *domain.Registration) { registration.DisplayName = strings.Repeat("n", 100) }},
		{name: "invalid email", change: func(registration *domain.Registration) { registration.Email = "invalid" }, wantErr: domain.ErrInvalidEmail},
		{name: "short password", change: func(registration *domain.Registration) { registration.Password = "short" }, wantErr: domain.ErrWeakPassword},
		{name: "long password", change: func(registration *domain.Registration) { registration.Password = strings.Repeat("p", 129) }, wantErr: domain.ErrWeakPassword},
		{name: "empty display name", change: func(registration *domain.Registration) { registration.DisplayName = "" }, wantErr: domain.ErrInvalidDisplayName},
		{name: "long display name", change: func(registration *domain.Registration) { registration.DisplayName = strings.Repeat("n", 101) }, wantErr: domain.ErrInvalidDisplayName},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			registration := valid
			test.change(&registration)
			got, err := domain.NewRegistration(registration)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("got %v, want %v", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got.Email != "student@example.test" {
				t.Fatalf("got email %q", got.Email)
			}
		})
	}
}

func TestAccountTokenPolicy(t *testing.T) {
	t.Parallel()
	created := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	valid := domain.AccountToken{ID: "50000000-0000-4000-8000-000000000001", UserID: "40000000-0000-4000-8000-000000000001", Purpose: domain.EmailVerificationToken, CreatedAt: created, ExpiresAt: created.Add(domain.EmailVerificationTokenLifetime)}
	for _, test := range []struct {
		name    string
		change  func(*domain.AccountToken)
		wantErr error
	}{
		{name: "valid", change: func(*domain.AccountToken) {}},
		{name: "unknown purpose", change: func(token *domain.AccountToken) { token.Purpose = "unknown" }, wantErr: domain.ErrInvalidAccountToken},
		{name: "bad id", change: func(token *domain.AccountToken) { token.ID = "bad" }, wantErr: domain.ErrInvalidAccountToken},
		{name: "expiry equals creation", change: func(token *domain.AccountToken) { token.ExpiresAt = token.CreatedAt }, wantErr: domain.ErrInvalidAccountToken},
		{name: "non-email-change cannot carry new email", change: func(token *domain.AccountToken) { token.NewEmail = "new@example.test" }, wantErr: domain.ErrInvalidAccountToken},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			token := valid
			test.change(&token)
			got, err := domain.NewAccountToken(token)
			if test.wantErr != nil {
				if !errors.Is(err, test.wantErr) {
					t.Fatalf("got %v, want %v", err, test.wantErr)
				}
				return
			}
			if err != nil || !got.UsableAt(created) || got.UsableAt(got.ExpiresAt) {
				t.Fatalf("token validity: token=%+v err=%v", got, err)
			}
		})
	}

	emailChange := valid
	emailChange.Purpose = domain.EmailChangeToken
	emailChange.NewEmail = "new@example.test"
	if _, err := domain.NewAccountToken(emailChange); err != nil {
		t.Fatalf("valid email-change token: %v", err)
	}
	emailChange.NewEmail = "not-an-email"
	if _, err := domain.NewAccountToken(emailChange); !errors.Is(err, domain.ErrInvalidAccountToken) {
		t.Fatalf("invalid email-change token: %v", err)
	}

	consumed := valid
	consumedAt := created.Add(time.Minute)
	consumed.ConsumedAt = &consumedAt
	token, err := domain.NewAccountToken(consumed)
	if err != nil {
		t.Fatal(err)
	}
	if token.UsableAt(created.Add(2 * time.Minute)) {
		t.Fatal("consumed token is usable")
	}
	if domain.PasswordResetTokenLifetime != time.Hour || domain.EmailVerificationTokenLifetime != 24*time.Hour {
		t.Fatal("unexpected token lifetimes")
	}
}

func TestPasswordAndEmailChangePolicies(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		old  string
		new  string
		want error
	}{
		{name: "valid password", old: "old-password", new: "new-password"},
		{name: "same password", old: "same-password", new: "same-password", want: domain.ErrPasswordUnchanged},
		{name: "weak password", old: "old-password", new: "short", want: domain.ErrWeakPassword},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.ValidatePasswordChange(test.old, test.new); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}

	for _, test := range []struct {
		name  string
		email string
		want  error
	}{
		{name: "normalizes email", email: " New@Example.Test "},
		{name: "invalid email", email: "not-an-email", want: domain.ErrInvalidEmail},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.ValidateEmailChange(test.email)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if test.want == nil && got != "new@example.test" {
				t.Fatalf("got %q", got)
			}
		})
	}
}
