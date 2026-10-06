package ports

import (
	"context"
	"time"

	"courses/internal/identity/domain"
)

type CredentialVerifier interface {
	VerifyCredentials(context.Context, string, string) (domain.User, error)
}

type IssuedSession struct {
	Session domain.Session
	Token   string
}

type SessionStore interface {
	Rotate(context.Context, string, string, time.Time, time.Time) (IssuedSession, error)
	Resolve(context.Context, string, time.Time) (domain.User, error)
	Revoke(context.Context, string) error
	DeleteIdentitySessionsForUser(context.Context, string) error
	DeleteUserSessionsExceptCurrent(context.Context, string, []byte) error
}
