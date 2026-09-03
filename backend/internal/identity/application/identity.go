package application

import (
	"context"
	"encoding/base64"
	"time"

	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

const (
	sessionLifetime  = 12 * time.Hour
	operationTimeout = 3 * time.Second
	// Session timestamps are written to the session store and read back before
	// they are trusted. Stores persist them at microsecond resolution, so a
	// clock reading is normalized before it becomes part of that round trip.
	storeResolution = time.Microsecond
)

type Clock func() time.Time

type Identity struct {
	credentials ports.CredentialVerifier
	sessions    ports.SessionStore
	now         Clock
}

type LoginInput struct {
	Email          string
	Password       string
	PresentedToken string
}

type LoginResult struct {
	User    domain.User
	Session domain.Session
	Token   string
}

func NewIdentity(credentials ports.CredentialVerifier, sessions ports.SessionStore, now Clock) *Identity {
	return &Identity{credentials: credentials, sessions: sessions, now: now}
}

func (identity *Identity) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil || len(input.Password) == 0 || len(input.Password) > 1024 {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	user, err := identity.credentials.VerifyCredentials(ctx, email, input.Password)
	if err != nil {
		return LoginResult{}, err
	}
	if _, err := domain.NewUser(user); err != nil || !user.Active() {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	now := identity.timestamp()
	issued, err := identity.sessions.Rotate(ctx, user.ID, input.PresentedToken, now, now.Add(sessionLifetime))
	if err != nil {
		return LoginResult{}, err
	}
	session, err := domain.NewSession(issued.Session)
	if err != nil || session.UserID != user.ID || !session.CreatedAt.Equal(now) || !session.ExpiresAt.Equal(now.Add(sessionLifetime)) || !validToken(issued.Token) {
		return LoginResult{}, domain.ErrDependencyUnavailable
	}
	return LoginResult{User: user, Session: session, Token: issued.Token}, nil
}

func (identity *Identity) Authenticate(ctx context.Context, token string) (domain.User, error) {
	if !validToken(token) {
		return domain.User{}, domain.ErrAuthenticationRequired
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	user, err := identity.sessions.Resolve(ctx, token, identity.timestamp())
	if err != nil {
		return domain.User{}, err
	}
	if _, err := domain.NewUser(user); err != nil {
		return domain.User{}, domain.ErrDependencyUnavailable
	}
	if !user.Active() {
		return domain.User{}, domain.ErrAuthenticationRequired
	}
	return user, nil
}

func (identity *Identity) Logout(ctx context.Context, token string) error {
	if !validToken(token) {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	return identity.sessions.Revoke(ctx, token)
}

func (identity *Identity) timestamp() time.Time {
	return identity.now().UTC().Truncate(storeResolution)
}

func validToken(token string) bool {
	if len(token) != 43 {
		return false
	}
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil && len(decoded) == 32
}
