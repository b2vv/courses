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
	// AdvanceTotpCounter records that the given TOTP step was consumed, returning true only if it
	// was strictly newer than the account's previously recorded step. A false result means the step
	// (and therefore the code) was already used — a replay — and must be rejected. The update is
	// atomic, so two concurrent submissions of the same code see exactly one true (SA13-B / I-F5).
	AdvanceTotpCounter(ctx context.Context, userID string, counter int64) (bool, error)
}

type RecoveryCodeStore interface {
	InsertCodes(context.Context, string, [][32]byte) error
	ConsumeCode(context.Context, string, [32]byte) (bool, error)
	DeleteCodes(context.Context, string) error
}

type PendingTwoFactor struct {
	UserID     string
	RememberMe bool
	// LoginRateLimitKey carries the login throttle key from the password step to
	// the second-factor step so that a successful CompleteTwoFactor — and only a
	// success — resets it. The 2FA-required login path therefore stays counted
	// until the second factor completes (audit H1 / SA01 Bug A).
	LoginRateLimitKey string
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
