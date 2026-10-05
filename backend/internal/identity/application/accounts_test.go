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
func (f *accountUsers) UpdatePassword(context.Context, string, string) error { return nil }
func (f *accountUsers) UpdateEmail(context.Context, string, string) error    { return nil }
func (f *accountUsers) FindUser(context.Context, string) (domain.User, error) {
	return f.found, f.findErr
}

type accountHasher struct {
	hash string
	err  error
}

func (f accountHasher) Hash(string) (string, error) { return f.hash, f.err }

type accountTokens struct {
	issued      []domain.AccountToken
	digest      []byte
	consumeID   string
	consumeErr  error
	invalidated []string
}

func (f *accountTokens) Issue(_ context.Context, token domain.AccountToken, digest []byte) error {
	f.issued = append(f.issued, token)
	f.digest = append([]byte(nil), digest...)
	return nil
}
func (f *accountTokens) Consume(_ context.Context, _ domain.AccountTokenPurpose, digest []byte) (string, string, error) {
	f.digest = append([]byte(nil), digest...)
	return f.consumeID, "", f.consumeErr
}
func (f *accountTokens) InvalidateByPurpose(_ context.Context, id string, _ domain.AccountTokenPurpose) error {
	f.invalidated = append(f.invalidated, id)
	return nil
}

type accountMailer struct {
	calls int
	err   error
}

func (f *accountMailer) SendVerification(context.Context, string, string, string) error {
	f.calls++
	return f.err
}
func (f *accountMailer) SendPasswordReset(context.Context, string, string, string) error { return nil }
func (f *accountMailer) SendEmailChange(context.Context, string, string, string) error   { return nil }

func TestRegisterCreatesUnverifiedUserIssuesTokenAndContinuesAfterMailFailure(t *testing.T) {
	now := time.Date(2026, 10, 5, 12, 0, 0, 0, time.UTC)
	users := &accountUsers{}
	tokens := &accountTokens{}
	mailer := &accountMailer{err: errors.New("smtp unavailable")}
	session := &sessions{issued: ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: "40000000-0000-4000-8000-000000000001", CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: validToken('A')}}
	accounts := application.NewAccounts(users, users, accountHasher{hash: "hash"}, tokens, mailer, session, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
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
	accounts := application.NewAccounts(users, users, accountHasher{hash: "hash"}, &accountTokens{}, &accountMailer{}, session, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
	_, err := accounts.Register(context.Background(), application.RegisterInput{Email: "student@example.test", Password: "correct password", DisplayName: "Student"}, "uk")
	if !errors.Is(err, domain.ErrEmailTaken) || session.rotateCalls != 0 {
		t.Fatalf("err=%v sessions=%d", err, session.rotateCalls)
	}
}

func TestVerifyEmailConsumesTokenAndMarksUserVerified(t *testing.T) {
	users := &accountUsers{}
	tokens := &accountTokens{consumeID: "40000000-0000-4000-8000-000000000001"}
	accounts := application.NewAccounts(users, users, accountHasher{}, tokens, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
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
			accounts := application.NewAccounts(&accountUsers{}, &accountUsers{}, accountHasher{}, &accountTokens{consumeErr: test.err}, &accountMailer{}, &sessions{}, time.Now, func() string { return "40000000-0000-4000-8000-000000000001" })
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
	accounts := application.NewAccounts(users, users, accountHasher{}, tokens, mailer, &sessions{}, func() time.Time { return now }, func() string { return "40000000-0000-4000-8000-000000000001" })
	if sent, err := accounts.ResendVerification(context.Background(), users.found.ID, "uk"); err != nil || !sent || len(tokens.invalidated) != 1 || len(tokens.issued) != 1 {
		t.Fatalf("sent=%v err=%v tokens=%+v", sent, err, tokens)
	}
	users.found.EmailVerified = true
	if sent, err := accounts.ResendVerification(context.Background(), users.found.ID, "uk"); err != nil || sent || len(tokens.invalidated) != 1 {
		t.Fatalf("idempotence: sent=%v err=%v", sent, err)
	}
}
