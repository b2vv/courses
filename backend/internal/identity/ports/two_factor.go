package ports

import (
	"context"
	"time"
)

// TwoFactor is the durable TOTP state for one identity.
type TwoFactor struct {
	Secret  string
	Enabled bool
}

type TwoFactorStore interface {
	GetTwoFactor(context.Context, string) (TwoFactor, error)
	SetTotpSecret(context.Context, string, string) error
	EnableTotp(context.Context, string) error
	DisableTotp(context.Context, string) error
}

type RecoveryCodeStore interface {
	InsertCodes(context.Context, string, [][32]byte) error
	ConsumeCode(context.Context, string, [32]byte) (bool, error)
	DeleteCodes(context.Context, string) error
}

type PendingTwoFactor struct {
	UserID     string
	RememberMe bool
}

type PendingTwoFactorStore interface {
	Put(context.Context, string, PendingTwoFactor, time.Duration) error
	Take(context.Context, string) (PendingTwoFactor, bool, error)
}

// PasswordVerifier is deliberately separate from PasswordHasher: callers of a
// password re-authentication flow must not obtain password hashes.
type PasswordVerifier interface {
	VerifyPassword(context.Context, string, string) error
}
