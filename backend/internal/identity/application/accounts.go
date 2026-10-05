package application

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

type Accounts struct {
	users       ports.UserWriter
	reader      ports.UserReader
	passwords   ports.PasswordHasher
	tokens      ports.TokenStore
	mailer      ports.Mailer
	sessions    ports.SessionStore
	limiter     ports.RateLimiter
	credentials ports.CredentialVerifier
	now         Clock
	newID       func() string
}

type RegisterInput struct {
	Email       string
	Password    string
	DisplayName string
}

type RegisterResult struct {
	User                  domain.User
	Session               domain.Session
	Token                 string
	VerificationEmailSent bool
}

func NewAccounts(users ports.UserWriter, reader ports.UserReader, passwords ports.PasswordHasher, tokens ports.TokenStore, mailer ports.Mailer, sessions ports.SessionStore, now Clock, newID func() string, dependencies ...AccountsDependency) *Accounts {
	accounts := &Accounts{users: users, reader: reader, passwords: passwords, tokens: tokens, mailer: mailer, sessions: sessions, now: now, newID: newID}
	for _, dependency := range dependencies {
		if dependency.Limiter != nil {
			accounts.limiter = dependency.Limiter
		}
		if dependency.Credentials != nil {
			accounts.credentials = dependency.Credentials
		}
	}
	return accounts
}

type AccountsDependency struct {
	Limiter     ports.RateLimiter
	Credentials ports.CredentialVerifier
}

func (accounts *Accounts) Register(ctx context.Context, input RegisterInput, lang string) (RegisterResult, error) {
	registration, err := domain.NewRegistration(domain.Registration(input))
	if err != nil {
		return RegisterResult{}, err
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	hash, err := accounts.passwords.Hash(registration.Password)
	if err != nil {
		return RegisterResult{}, domain.ErrDependencyUnavailable
	}
	now := accounts.timestamp()
	user, err := domain.NewUser(domain.User{ID: accounts.newID(), Email: registration.Email, DisplayName: registration.DisplayName, Role: domain.Student, CreatedAt: now, UpdatedAt: now})
	if err != nil {
		return RegisterResult{}, domain.ErrDependencyUnavailable
	}
	user, err = accounts.users.CreateUser(ctx, user, hash)
	if err != nil {
		return RegisterResult{}, err
	}
	rawToken, err := accounts.issueVerification(ctx, user.ID, now)
	if err != nil {
		return RegisterResult{}, err
	}
	emailSent := accounts.mailer.SendVerification(ctx, user.Email, rawToken, lang) == nil
	issued, err := accounts.sessions.Rotate(ctx, user.ID, "", now, now.Add(sessionLifetime))
	if err != nil {
		return RegisterResult{}, err
	}
	session, err := validIssuedSession(issued, user.ID, now)
	if err != nil {
		return RegisterResult{}, err
	}
	return RegisterResult{User: user, Session: session, Token: issued.Token, VerificationEmailSent: emailSent}, nil
}

func (accounts *Accounts) VerifyEmail(ctx context.Context, rawToken string) error {
	digest, ok := accountTokenDigest(rawToken)
	if !ok {
		return domain.ErrTokenUsed
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	userID, _, err := accounts.tokens.Consume(ctx, domain.EmailVerificationToken, digest[:])
	if err != nil {
		return err
	}
	return accounts.users.SetEmailVerified(ctx, userID)
}

func (accounts *Accounts) ResendVerification(ctx context.Context, identifier, lang string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	user, err := accounts.reader.FindUser(ctx, identifier)
	if err != nil {
		return false, err
	}
	if user.EmailVerified {
		return false, nil
	}
	if err := accounts.tokens.InvalidateByPurpose(ctx, user.ID, domain.EmailVerificationToken); err != nil {
		return false, err
	}
	rawToken, err := accounts.issueVerification(ctx, user.ID, accounts.timestamp())
	if err != nil {
		return false, err
	}
	return accounts.mailer.SendVerification(ctx, user.Email, rawToken, lang) == nil, nil
}

func (accounts *Accounts) RequestPasswordReset(ctx context.Context, email, ip, lang string) error {
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	if err := accounts.allow(ctx, passwordResetRateLimitKey(email, ip)); err != nil {
		return err
	}
	normalizedEmail, err := domain.NormalizeEmail(email)
	if err != nil {
		return nil
	}
	user, err := accounts.reader.FindUser(ctx, normalizedEmail)
	if errors.Is(err, domain.ErrUserNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := accounts.tokens.InvalidateByPurpose(ctx, user.ID, domain.PasswordResetToken); err != nil {
		return err
	}
	rawToken, err := accounts.issueToken(ctx, user.ID, domain.PasswordResetToken, "", domain.PasswordResetTokenLifetime, accounts.timestamp())
	if err != nil {
		return err
	}
	_ = accounts.mailer.SendPasswordReset(ctx, user.Email, rawToken, lang)
	return nil
}

func (accounts *Accounts) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	digest, ok := accountTokenDigest(rawToken)
	if !ok {
		return domain.ErrTokenUsed
	}
	if err := domain.ValidatePassword(newPassword); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	userID, _, err := accounts.tokens.Consume(ctx, domain.PasswordResetToken, digest[:])
	if err != nil {
		return err
	}
	hash, err := accounts.passwords.Hash(newPassword)
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	if err := accounts.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	return accounts.sessions.DeleteIdentitySessionsForUser(ctx, userID)
}

func (accounts *Accounts) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string, currentDigest []byte) error {
	if accounts.credentials == nil {
		return domain.ErrDependencyUnavailable
	}
	if err := domain.ValidatePasswordChange(oldPassword, newPassword); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	user, err := accounts.reader.FindUser(ctx, userID)
	if err != nil {
		return err
	}
	verified, err := accounts.credentials.VerifyCredentials(ctx, user.Email, oldPassword)
	if err != nil {
		return err
	}
	if verified.ID != userID {
		return domain.ErrInvalidCredentials
	}
	hash, err := accounts.passwords.Hash(newPassword)
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	if err := accounts.users.UpdatePassword(ctx, userID, hash); err != nil {
		return err
	}
	return accounts.sessions.DeleteUserSessionsExceptCurrent(ctx, userID, currentDigest)
}

func (accounts *Accounts) RequestEmailChange(ctx context.Context, userID, newEmail, lang string) error {
	email, err := domain.ValidateEmailChange(newEmail)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	if _, err := accounts.reader.FindUser(ctx, email); err == nil {
		return domain.ErrEmailTaken
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	rawToken, err := accounts.issueToken(ctx, userID, domain.EmailChangeToken, email, domain.EmailVerificationTokenLifetime, accounts.timestamp())
	if err != nil {
		return err
	}
	_ = accounts.mailer.SendEmailChange(ctx, email, rawToken, lang)
	return nil
}

func (accounts *Accounts) ConfirmEmailChange(ctx context.Context, rawToken string) error {
	digest, ok := accountTokenDigest(rawToken)
	if !ok {
		return domain.ErrTokenUsed
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	userID, email, err := accounts.tokens.Consume(ctx, domain.EmailChangeToken, digest[:])
	if err != nil {
		return err
	}
	email, err = domain.ValidateEmailChange(email)
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	if _, err := accounts.reader.FindUser(ctx, email); err == nil {
		return domain.ErrEmailTaken
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	return accounts.users.UpdateEmail(ctx, userID, email)
}

func (accounts *Accounts) issueVerification(ctx context.Context, userID string, now time.Time) (string, error) {
	return accounts.issueToken(ctx, userID, domain.EmailVerificationToken, "", domain.EmailVerificationTokenLifetime, now)
}

func (accounts *Accounts) issueToken(ctx context.Context, userID string, purpose domain.AccountTokenPurpose, newEmail string, lifetime time.Duration, now time.Time) (string, error) {
	material, err := NewTokenMaterial()
	if err != nil {
		return "", domain.ErrDependencyUnavailable
	}
	token, err := domain.NewAccountToken(domain.AccountToken{ID: accounts.newID(), UserID: userID, Purpose: purpose, NewEmail: newEmail, CreatedAt: now, ExpiresAt: now.Add(lifetime)})
	if err != nil {
		return "", domain.ErrDependencyUnavailable
	}
	if err := accounts.tokens.Issue(ctx, token, material.Digest[:]); err != nil {
		return "", err
	}
	return material.Raw, nil
}

func (accounts *Accounts) allow(ctx context.Context, key string) error {
	if accounts.limiter == nil {
		return nil
	}
	allowed, err := accounts.limiter.Allow(ctx, key)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	return nil
}

func passwordResetRateLimitKey(email, ip string) string {
	return "password-reset:" + strings.ToLower(strings.TrimSpace(email)) + ":" + ip
}

func (accounts *Accounts) timestamp() time.Time {
	return accounts.now().UTC().Truncate(storeResolution)
}

func accountTokenDigest(rawToken string) ([sha256.Size]byte, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(rawToken)
	if err != nil || len(raw) != accountTokenBytes {
		return [sha256.Size]byte{}, false
	}
	return sha256.Sum256(raw), true
}

func validIssuedSession(issued ports.IssuedSession, userID string, createdAt time.Time) (domain.Session, error) {
	session, err := domain.NewSession(issued.Session)
	if err != nil || session.UserID != userID || !session.CreatedAt.Equal(createdAt) || !session.ExpiresAt.Equal(createdAt.Add(sessionLifetime)) || !validToken(issued.Token) {
		return domain.Session{}, domain.ErrDependencyUnavailable
	}
	return session, nil
}
