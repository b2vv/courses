package ports

import (
	"context"

	"courses/internal/identity/domain"
)

type UserWriter interface {
	CreateUser(context.Context, domain.User, string) (domain.User, error)
	SetEmailVerified(context.Context, string) error
	UpdatePassword(context.Context, string, string) error
	UpdateEmail(context.Context, string, string) error
}

type TokenStore interface {
	Issue(context.Context, domain.AccountToken, []byte) error
	Consume(context.Context, domain.AccountTokenPurpose, []byte) (userID string, newEmail string, err error)
	InvalidateByPurpose(context.Context, string, domain.AccountTokenPurpose) error
}

type RateLimiter interface {
	Allow(context.Context, string) (bool, error)
	Reset(context.Context, string) error
}

type Mailer interface {
	SendVerification(context.Context, string, string, string) error
	SendPasswordReset(context.Context, string, string, string) error
	SendEmailChange(context.Context, string, string, string) error
}
