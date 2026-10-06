package domain_test

import (
	"errors"
	"testing"
	"time"

	"courses/internal/identity/domain"
)

func TestSessionInvariantsAndExpiryBoundary(t *testing.T) {
	t.Parallel()
	created := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	valid := domain.Session{
		ID:        "50000000-0000-4000-8000-000000000001",
		UserID:    "40000000-0000-4000-8000-000000000001",
		CreatedAt: created,
		ExpiresAt: created.Add(12 * time.Hour),
	}
	session, err := domain.NewSession(valid)
	if err != nil {
		t.Fatal(err)
	}
	if session.IsActiveAt(created.Add(-time.Nanosecond)) {
		t.Fatal("session is active before creation")
	}
	if !session.IsActiveAt(created) || !session.IsActiveAt(valid.ExpiresAt.Add(-time.Nanosecond)) {
		t.Fatal("session should be active inside its half-open lifetime")
	}
	if session.IsActiveAt(valid.ExpiresAt) || session.IsActiveAt(valid.ExpiresAt.Add(time.Nanosecond)) {
		t.Fatal("session must be expired at and after expires_at")
	}

	for _, test := range []struct {
		name   string
		change func(*domain.Session)
	}{
		{name: "invalid session id", change: func(session *domain.Session) { session.ID = "invalid" }},
		{name: "invalid user id", change: func(session *domain.Session) { session.UserID = "invalid" }},
		{name: "missing creation", change: func(session *domain.Session) { session.CreatedAt = time.Time{} }},
		{name: "equal expiry", change: func(session *domain.Session) { session.ExpiresAt = created }},
		{name: "expiry before creation", change: func(session *domain.Session) { session.ExpiresAt = created.Add(-time.Second) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			session := valid
			test.change(&session)
			if _, err := domain.NewSession(session); !errors.Is(err, domain.ErrInvalidSession) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestIdentityErrorsAreStable(t *testing.T) {
	t.Parallel()
	for errorValue, want := range map[error]string{
		domain.ErrInvalidUser:             "invalid_user",
		domain.ErrInvalidEmail:            "invalid_email",
		domain.ErrInvalidDisplayName:      "invalid_display_name",
		domain.ErrWeakPassword:            "weak_password",
		domain.ErrPasswordUnchanged:       "password_unchanged",
		domain.ErrEmailTaken:              "email_taken",
		domain.ErrTokenUsed:               "token_used",
		domain.ErrInvalidAccountToken:     "invalid_account_token",
		domain.ErrInvalidSession:          "invalid_session",
		domain.ErrInvalidCredentials:      "invalid_credentials",
		domain.ErrAuthenticationRequired:  "authentication_required",
		domain.ErrForbidden:               "forbidden",
		domain.ErrDependencyUnavailable:   "dependency_unavailable",
		domain.ErrCannotTargetSelf:        "cannot_target_self",
		domain.ErrAdminImmutable:          "admin_immutable",
		domain.ErrLastAdmin:               "last_admin",
		domain.ErrUserNotFound:            "user_not_found",
		domain.ErrInvalidTOTP:             "invalid_totp",
		domain.ErrTwoFactorAlreadyEnabled: "twofactor_already_enabled",
		domain.ErrTwoFactorNotEnabled:     "twofactor_not_enabled",
		domain.ErrRecoveryCodeInvalid:     "recovery_code_invalid",
	} {
		if errorValue.Error() != want {
			t.Fatalf("got %q, want %q", errorValue.Error(), want)
		}
	}
}
