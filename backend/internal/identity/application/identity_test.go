package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

type credentials struct {
	calls    int
	email    string
	password string
	user     domain.User
	err      error
}

func (credentials *credentials) VerifyCredentials(_ context.Context, email, password string) (domain.User, error) {
	credentials.calls++
	credentials.email = email
	credentials.password = password
	return credentials.user, credentials.err
}

type sessions struct {
	rotateCalls  int
	resolveCalls int
	revokeCalls  int
	userID       string
	presented    string
	createdAt    time.Time
	expiresAt    time.Time
	token        string
	issued       ports.IssuedSession
	resolvedUser domain.User
	rotateErr    error
	resolveErr   error
	revokeErr    error
}

func (sessions *sessions) Rotate(_ context.Context, userID, presentedToken string, createdAt, expiresAt time.Time) (ports.IssuedSession, error) {
	sessions.rotateCalls++
	sessions.userID = userID
	sessions.presented = presentedToken
	sessions.createdAt = createdAt
	sessions.expiresAt = expiresAt
	return sessions.issued, sessions.rotateErr
}

func (sessions *sessions) Resolve(_ context.Context, token string, _ time.Time) (domain.User, error) {
	sessions.resolveCalls++
	sessions.token = token
	return sessions.resolvedUser, sessions.resolveErr
}

func (sessions *sessions) Revoke(_ context.Context, token string) error {
	sessions.revokeCalls++
	sessions.token = token
	return sessions.revokeErr
}

func TestLoginNormalizesCredentialsAndIssuesTwelveHourSession(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	user := validUser(now)
	issued := ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: user.ID, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: validToken('A')}
	credentialPort := &credentials{user: user}
	sessionPort := &sessions{issued: issued}
	identity := application.NewIdentity(credentialPort, sessionPort, func() time.Time { return now })

	result, err := identity.Login(context.Background(), application.LoginInput{Email: " Student@CourseForge.Local ", Password: "correct horse", PresentedToken: validToken('B')})
	if err != nil {
		t.Fatal(err)
	}
	if credentialPort.email != user.Email || credentialPort.password != "correct horse" || sessionPort.userID != user.ID || sessionPort.presented != validToken('B') {
		t.Fatalf("wrong adapter inputs: credentials=%+v sessions=%+v", credentialPort, sessionPort)
	}
	if !sessionPort.createdAt.Equal(now) || !sessionPort.expiresAt.Equal(now.Add(12*time.Hour)) {
		t.Fatalf("wrong lifetime: %s to %s", sessionPort.createdAt, sessionPort.expiresAt)
	}
	if result.User != user || result.Session != issued.Session || result.Token != issued.Token {
		t.Fatalf("wrong login result: %+v", result)
	}
}

func TestLoginRejectsInvalidInputBeforePortsAndPropagatesExplicitErrors(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		input    application.LoginInput
		want     error
		portCall bool
	}{
		{name: "invalid email syntax", input: application.LoginInput{Email: "invalid", Password: "password"}, want: domain.ErrInvalidCredentials},
		{name: "empty password", input: application.LoginInput{Email: "a@b", Password: ""}, want: domain.ErrInvalidCredentials},
		{name: "oversized password", input: application.LoginInput{Email: "a@b", Password: repeat("p", 1025)}, want: domain.ErrInvalidCredentials},
		{name: "wrong password", input: application.LoginInput{Email: "a@b", Password: "wrong"}, want: domain.ErrInvalidCredentials, portCall: true},
		{name: "database unavailable", input: application.LoginInput{Email: "a@b", Password: "password"}, want: domain.ErrDependencyUnavailable, portCall: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			credentialPort := &credentials{err: test.want}
			sessionPort := &sessions{}
			_, err := application.NewIdentity(credentialPort, sessionPort, time.Now).Login(context.Background(), test.input)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v", err)
			}
			if got := credentialPort.calls > 0; got != test.portCall {
				t.Fatalf("credential call=%v", got)
			}
			if sessionPort.rotateCalls != 0 {
				t.Fatal("session rotated after failed credentials")
			}
		})
	}
}

func TestAuthenticateAndLogoutRejectMalformedTokensWithoutStoreLookup(t *testing.T) {
	t.Parallel()
	for _, token := range []string{"", "short", repeat("!", 43), validToken('A') + "A"} {
		t.Run(token, func(t *testing.T) {
			t.Parallel()
			store := &sessions{}
			identity := application.NewIdentity(&credentials{}, store, time.Now)
			if _, err := identity.Authenticate(context.Background(), token); !errors.Is(err, domain.ErrAuthenticationRequired) {
				t.Fatalf("authenticate: %v", err)
			}
			if err := identity.Logout(context.Background(), token); err != nil {
				t.Fatalf("logout: %v", err)
			}
			if store.resolveCalls != 0 || store.revokeCalls != 0 {
				t.Fatalf("store called for malformed token: %+v", store)
			}
		})
	}
}

func TestAuthenticateAndLogoutUseOneTimestampAndPropagateStoreErrors(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	token := validToken('A')
	user := validUser(now)
	store := &sessions{resolvedUser: user}
	identity := application.NewIdentity(&credentials{}, store, func() time.Time { return now })
	if got, err := identity.Authenticate(context.Background(), token); err != nil || got != user {
		t.Fatalf("user=%+v err=%v", got, err)
	}
	if store.resolveCalls != 1 || store.token != token {
		t.Fatalf("resolve call: %+v", store)
	}
	store.revokeErr = domain.ErrDependencyUnavailable
	if err := identity.Logout(context.Background(), token); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("logout: %v", err)
	}
	if store.revokeCalls != 1 {
		t.Fatalf("revoke calls=%d", store.revokeCalls)
	}
}

func validUser(now time.Time) domain.User {
	return domain.User{ID: "40000000-0000-4000-8000-000000000001", Email: "student@courseforge.local", DisplayName: "Студентка Курсу", Role: domain.Student, CreatedAt: now, UpdatedAt: now}
}

func validToken(character byte) string {
	return repeat(string(character), 43)
}

func repeat(value string, count int) string {
	result := ""
	for range count {
		result += value
	}
	return result
}
