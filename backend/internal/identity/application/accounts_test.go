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

type accountUsers struct {
	created              domain.User
	createErr, verifyErr error
	verified             []string
	found                domain.User
	findErr              error
	updatedPassword      string
	updatedEmail         string
	lookups              []string
}

func (f *accountUsers) CreateUser(_ context.Context, user domain.User, _ string) (domain.User, error) {
	f.created = user
	if f.createErr != nil {
		return domain.User{}, f.createErr
	}
	return user, nil
}
func (f *accountUsers) SetEmailVerified(_ context.Context, id string) error {
	f.verified = append(f.verified, id)
	return f.verifyErr
}
func (f *accountUsers) UpdatePassword(_ context.Context, _ string, hash string) error {
	f.updatedPassword = hash
	return nil
}
func (f *accountUsers) UpdateEmail(_ context.Context, _ string, email string) error {
	f.updatedEmail = email
	return nil
}
func (f *accountUsers) FindUser(_ context.Context, identifier string) (domain.User, error) {
	f.lookups = append(f.lookups, identifier)
	return f.found, f.findErr
}

type accountHasher struct {
	hash     string
	err      error
	password string
}

func (f *accountHasher) Hash(password string) (string, error) {
	f.password = password
	return f.hash, f.err
}

type accountTokens struct {
	issued       []domain.AccountToken
	digest       []byte
	consumeID    string
	consumeEmail string
	consumeErr   error
	invalidated  []string
}

func (f *accountTokens) Issue(_ context.Context, token domain.AccountToken, digest []byte) error {
	f.issued = append(f.issued, token)
	f.digest = append([]byte(nil), digest...)
	return nil
}
func (f *accountTokens) Consume(_ context.Context, _ domain.AccountTokenPurpose, digest []byte) (string, string, error) {
	f.digest = append([]byte(nil), digest...)
	return f.consumeID, f.consumeEmail, f.consumeErr
}
func (f *accountTokens) InvalidateByPurpose(_ context.Context, id string, _ domain.AccountTokenPurpose) error {
	f.invalidated = append(f.invalidated, id)
	return nil
}

type accountMailer struct {
	calls, passwordResets, emailChanges int
	err                                 error
}

func (f *accountMailer) SendVerification(context.Context, string, string, string) error {
	f.calls++
	return f.err
}
func (f *accountMailer) SendPasswordReset(context.Context, string, string, string) error {
	f.passwordResets++
	return f.err
}
func (f *accountMailer) SendEmailChange(context.Context, string, string, string) error {
	f.emailChanges++
	return f.err
}

func TestRegisterCreatesUnverifiedUserIssuesTokenAndContinuesAfterMailFailure(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	users := &accountUsers{}
	tokens := &accountTokens{}
	mailer := &accountMailer{err: errors.New("smtp unavailable")}
	session := &sessions{issued: ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: "40000000-0000-4000-8000-000000000001", CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: validToken('A')}}
	accounts := application.NewAccounts(users, users, &accountHasher{hash: "hash"}, tokens, mailer, session, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
	result, err := accounts.Register(context.Background(), application.RegisterInput{Email: " Student@CourseForge.Local ", Password: "correct password", DisplayName: "Student"}, "en")
	if err != nil {
		t.Fatal(err)
	}
	if result.VerificationEmailSent || result.User.EmailVerified || result.User.Email != "student@courseforge.local" || session.rotateCalls != 1 {
		t.Fatalf("wrong result: %+v", result)
	}
	if len(tokens.issued) != 1 || !tokens.issued[0].ExpiresAt.Equal(now.Add(domain.EmailVerificationTokenLifetime)) || len(tokens.digest) != 32 {
		t.Fatalf("token was not issued correctly: %+v", tokens)
	}
}

func TestRegisterReturnsEmailTakenWithoutSession(t *testing.T) {
	users := &accountUsers{createErr: domain.ErrEmailTaken}
	session := &sessions{}
	accounts := application.NewAccounts(users, users, &accountHasher{hash: "hash"}, &accountTokens{}, &accountMailer{}, session, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
	_, err := accounts.Register(context.Background(), application.RegisterInput{Email: "student@example.test", Password: "correct password", DisplayName: "Student"}, "uk")
	if !errors.Is(err, domain.ErrEmailTaken) || session.rotateCalls != 0 {
		t.Fatalf("err=%v sessions=%d", err, session.rotateCalls)
	}
}

func TestVerifyEmailConsumesTokenAndMarksUserVerified(t *testing.T) {
	users := &accountUsers{}
	tokens := &accountTokens{consumeID: "40000000-0000-4000-8000-000000000001"}
	accounts := application.NewAccounts(users, users, &accountHasher{}, tokens, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
	if err := accounts.VerifyEmail(context.Background(), validToken('A')); err != nil {
		t.Fatal(err)
	}
	if len(users.verified) != 1 || len(tokens.digest) != 32 {
		t.Fatalf("verification calls: %+v %+v", users, tokens)
	}
}

func TestVerifyEmailReturnsUsedAndExpiredTokenStoreErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
	}{{name: "used", err: domain.ErrTokenUsed}, {name: "expired", err: domain.ErrTokenUsed}} {
		t.Run(test.name, func(t *testing.T) {
			accounts := application.NewAccounts(&accountUsers{}, &accountUsers{}, &accountHasher{}, &accountTokens{consumeErr: test.err}, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
			if err := accounts.VerifyEmail(context.Background(), validToken('A')); !errors.Is(err, test.err) {
				t.Fatal(err)
			}
		})
	}
}

func TestResendVerificationInvalidatesThenIssuesUnlessAlreadyVerified(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	users := &accountUsers{found: validUser(now)}
	tokens := &accountTokens{}
	mailer := &accountMailer{}
	accounts := application.NewAccounts(users, users, &accountHasher{}, tokens, mailer, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
	if sent, err := accounts.ResendVerification(context.Background(), users.found.ID, "203.0.113.9", "uk"); err != nil || !sent || len(tokens.invalidated) != 1 || len(tokens.issued) != 1 {
		t.Fatalf("sent=%v err=%v tokens=%+v", sent, err, tokens)
	}
	users.found.EmailVerified = true
	if sent, err := accounts.ResendVerification(context.Background(), users.found.ID, "203.0.113.9", "uk"); err != nil || sent || len(tokens.invalidated) != 1 {
		t.Fatalf("idempotence: sent=%v err=%v", sent, err)
	}
}

// SA12/M12: an unauthenticated resend for an unknown account must NOT leak a
// yes/no oracle. Mirror RequestPasswordReset: a missing account is neutral
// (false, nil) and issues nothing — never ErrUserNotFound (which the handler
// maps to 404, distinguishing it from the 204 a real account returns).
func TestResendVerificationIsNeutralForUnknownAccount(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	users := &accountUsers{found: validUser(now), findErr: domain.ErrUserNotFound}
	tokens := &accountTokens{}
	accounts := application.NewAccounts(users, users, &accountHasher{}, tokens, &accountMailer{}, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
	if sent, err := accounts.ResendVerification(context.Background(), "ghost@courseforge.local", "203.0.113.9", "uk"); err != nil || sent {
		t.Fatalf("unknown account must be neutral: sent=%v err=%v", sent, err)
	}
	if len(tokens.issued) != 0 || len(tokens.invalidated) != 0 {
		t.Fatalf("unknown account must not touch tokens: %+v", tokens)
	}
}

// SA12 NB-1: the repository matches the stored (normalized) email exactly, so a
// submitted email must be normalized before the lookup — otherwise a mixed-case
// address silently finds nothing and, now that resend is neutral, looks like a
// success with no email sent. A user ID (authenticated path) passes unchanged.
func TestResendVerificationNormalizesSubmittedEmailButNotUserID(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	users := &accountUsers{found: validUser(now)}
	accounts := application.NewAccounts(users, users, &accountHasher{}, &accountTokens{}, &accountMailer{}, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
	if _, err := accounts.ResendVerification(context.Background(), "  Student@CourseForge.Local ", "203.0.113.9", "uk"); err != nil {
		t.Fatal(err)
	}
	if _, err := accounts.ResendVerification(context.Background(), users.found.ID, "203.0.113.9", "uk"); err != nil {
		t.Fatal(err)
	}
	if len(users.lookups) != 2 || users.lookups[0] != "student@courseforge.local" || users.lookups[1] != users.found.ID {
		t.Fatalf("lookups=%q", users.lookups)
	}
}

// SA12/M12: resend must be rate-limited by submitted identifier + IP, before
// the account lookup — so an attacker cannot hammer the oracle. Mirror
// RequestPasswordReset's limiter ordering.
func TestResendVerificationStopsBeforeLookupWhenRateLimited(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	users := &accountUsers{found: validUser(now)}
	tokens := &accountTokens{}
	attempts := &limiter{}
	accounts := application.NewAccounts(users, users, &accountHasher{}, tokens, &accountMailer{}, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" }, application.AccountsDependency{Limiter: attempts})
	if _, err := accounts.ResendVerification(context.Background(), "student@courseforge.local", "203.0.113.9", "uk"); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatalf("rate limit must deny: err=%v", err)
	}
	if attempts.allowCalls != 1 || attempts.key == "" || len(tokens.invalidated) != 0 || len(tokens.issued) != 0 {
		t.Fatalf("rate limit did not stop before lookup: limiter=%+v tokens=%+v", attempts, tokens)
	}
}

func TestRequestPasswordResetIsNeutralAndRateLimitedBySubmittedEmailAndIP(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		findErr error
		issued  int
	}{
		{name: "existing account", issued: 1},
		{name: "missing account", findErr: domain.ErrUserNotFound},
	} {
		t.Run(test.name, func(t *testing.T) {
			users := &accountUsers{found: validUser(now), findErr: test.findErr}
			tokens := &accountTokens{}
			accounts := application.NewAccounts(users, users, &accountHasher{}, tokens, &accountMailer{}, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
			if err := accounts.RequestPasswordReset(context.Background(), "Student@CourseForge.Local", "203.0.113.9", "uk"); err != nil {
				t.Fatal(err)
			}
			if len(tokens.issued) != test.issued {
				t.Fatalf("issued=%d", len(tokens.issued))
			}
		})
	}
}

func TestRequestPasswordResetStopsBeforeLookupWhenRateLimited(t *testing.T) {
	users := &accountUsers{found: validUser(time.Now())}
	attempts := &limiter{}
	accounts := application.NewAccounts(users, users, &accountHasher{}, &accountTokens{}, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" }, application.AccountsDependency{Limiter: attempts})
	if err := accounts.RequestPasswordReset(context.Background(), "Student@CourseForge.Local", "203.0.113.9", "uk"); !errors.Is(err, domain.ErrRateLimited) {
		t.Fatal(err)
	}
	if users.updatedPassword != "" || attempts.allowCalls != 1 || attempts.key == "" {
		t.Fatalf("rate limit did not stop reset: users=%+v limiter=%+v", users, attempts)
	}
}

func TestResetPasswordConsumesValidTokenAndDeletesAllSessions(t *testing.T) {
	users := &accountUsers{}
	hasher := &accountHasher{hash: "new-hash"}
	sessionStore := &sessions{}
	accounts := application.NewAccounts(users, users, hasher, &accountTokens{consumeID: "40000000-0000-4000-8000-000000000001"}, &accountMailer{}, sessionStore, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
	if err := accounts.ResetPassword(context.Background(), validToken('A'), "new correct password"); err != nil {
		t.Fatal(err)
	}
	if hasher.password != "new correct password" || users.updatedPassword != "new-hash" || sessionStore.deleteAllCalls != 1 {
		t.Fatalf("password reset side effects are wrong: hasher=%+v users=%+v sessions=%+v", hasher, users, sessionStore)
	}
}

func TestResetPasswordRejectsUsedAndExpiredTokens(t *testing.T) {
	for _, name := range []string{"used", "expired"} {
		t.Run(name, func(t *testing.T) {
			accounts := application.NewAccounts(&accountUsers{}, &accountUsers{}, &accountHasher{}, &accountTokens{consumeErr: domain.ErrTokenUsed}, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
			if err := accounts.ResetPassword(context.Background(), validToken('A'), "new correct password"); !errors.Is(err, domain.ErrTokenUsed) {
				t.Fatal(err)
			}
		})
	}
}

func TestChangePasswordVerifiesOldPasswordAndKeepsCurrentSession(t *testing.T) {
	now := time.Now()
	users := &accountUsers{found: validUser(now)}
	hasher := &accountHasher{hash: "new-hash"}
	store := &sessions{}
	digest := make([]byte, 32)
	accounts := application.NewAccounts(users, users, hasher, &accountTokens{}, &accountMailer{}, store, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" }, application.AccountsDependency{Credentials: &credentials{user: users.found}})
	if err := accounts.ChangePassword(context.Background(), users.found.ID, "old correct password", "new correct password", digest); err != nil {
		t.Fatal(err)
	}
	if store.deleteExceptCalls != 1 || string(store.currentDigest) != string(digest) || users.updatedPassword != "new-hash" {
		t.Fatalf("change password side effects are wrong: %+v %+v", users, store)
	}
	accounts = application.NewAccounts(users, users, hasher, &accountTokens{}, &accountMailer{}, store, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" }, application.AccountsDependency{Credentials: &credentials{err: domain.ErrInvalidCredentials}})
	if err := accounts.ChangePassword(context.Background(), users.found.ID, "wrong", "new correct password", digest); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatal(err)
	}
}

func TestEmailChangePrechecksAndRechecksUniqueness(t *testing.T) {
	now := time.Now()
	newEmail := "new@example.test"
	users := &accountUsers{findErr: domain.ErrUserNotFound}
	tokens := &accountTokens{}
	mailer := &accountMailer{}
	accounts := application.NewAccounts(users, users, &accountHasher{}, tokens, mailer, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
	if err := accounts.RequestEmailChange(context.Background(), "40000000-0000-4000-8000-000000000001", newEmail, "en"); err != nil || len(tokens.issued) != 1 || tokens.issued[0].NewEmail != newEmail || mailer.emailChanges != 1 {
		t.Fatalf("request email change: err=%v tokens=%+v mailer=%+v", err, tokens, mailer)
	}
	users.found = validUser(now)
	users.findErr = nil
	tokens.consumeID, tokens.consumeEmail = "40000000-0000-4000-8000-000000000001", newEmail
	if err := accounts.ConfirmEmailChange(context.Background(), validToken('A')); !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatal(err)
	}
	if users.updatedEmail != "" {
		t.Fatalf("TOCTOU collision updated email: %q", users.updatedEmail)
	}
}

func TestEmailChangeRejectsTakenAddressOnRequestAndUpdatesOnConfirm(t *testing.T) {
	now := time.Now()
	users := &accountUsers{found: validUser(now)}
	accounts := application.NewAccounts(users, users, &accountHasher{}, &accountTokens{}, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
	if err := accounts.RequestEmailChange(context.Background(), users.found.ID, "taken@example.test", "uk"); !errors.Is(err, domain.ErrEmailTaken) {
		t.Fatal(err)
	}
	users.findErr = domain.ErrUserNotFound
	tokens := &accountTokens{consumeID: users.found.ID, consumeEmail: "new@example.test"}
	accounts = application.NewAccounts(users, users, &accountHasher{}, tokens, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
	if err := accounts.ConfirmEmailChange(context.Background(), validToken('A')); err != nil {
		t.Fatal(err)
	}
	if users.updatedEmail != "new@example.test" {
		t.Fatalf("email was not updated and verified: %+v", users)
	}
}
