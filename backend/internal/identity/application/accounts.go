package application

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"time"

	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

type Accounts struct {
	users     ports.UserWriter
	reader    ports.UserReader
	passwords ports.PasswordHasher
	tokens    ports.TokenStore
	mailer    ports.Mailer
	sessions  ports.SessionStore
	now       Clock
	newID     func() string
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

func NewAccounts(users ports.UserWriter, reader ports.UserReader, passwords ports.PasswordHasher, tokens ports.TokenStore, mailer ports.Mailer, sessions ports.SessionStore, now Clock, newID func() string) *Accounts {
	return &Accounts{users: users, reader: reader, passwords: passwords, tokens: tokens, mailer: mailer, sessions: sessions, now: now, newID: newID}
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

func (accounts *Accounts) issueVerification(ctx context.Context, userID string, now time.Time) (string, error) {
	material, err := NewTokenMaterial()
	if err != nil {
		return "", domain.ErrDependencyUnavailable
	}
	token, err := domain.NewAccountToken(domain.AccountToken{ID: accounts.newID(), UserID: userID, Purpose: domain.EmailVerificationToken, CreatedAt: now, ExpiresAt: now.Add(domain.EmailVerificationTokenLifetime)})
	if err != nil {
		return "", domain.ErrDependencyUnavailable
	}
	if err := accounts.tokens.Issue(ctx, token, material.Digest[:]); err != nil {
		return "", err
	}
	return material.Raw, nil
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
