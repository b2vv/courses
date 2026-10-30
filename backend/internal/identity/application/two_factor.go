package application

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"time"

	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

const pendingTwoFactorTTL = 5 * time.Minute

type TwoFactor struct {
	store      ports.TwoFactorStore
	recovery   ports.RecoveryCodeStore
	pending    ports.PendingTwoFactorStore
	sessions   ports.SessionStore
	passwords  ports.PasswordVerifier
	limiter    ports.RateLimiter
	now        Clock
	newPending func() (string, error)
	logger     *slog.Logger
}

type TwoFactorDependency struct {
	Passwords  ports.PasswordVerifier
	Limiter    ports.RateLimiter
	Logger     *slog.Logger
	NewPending func() (string, error)
}

type EnrollTwoFactorResult struct{ Secret, URI string }
type ConfirmTwoFactorResult struct{ RecoveryCodes []string }
type CompleteTwoFactorInput struct{ PendingID, Code, PresentedToken string }
type CompleteTwoFactorResult struct {
	Session domain.Session
	Token   string
}

func NewTwoFactor(store ports.TwoFactorStore, recovery ports.RecoveryCodeStore, pending ports.PendingTwoFactorStore, sessions ports.SessionStore, now Clock, dependencies ...TwoFactorDependency) *TwoFactor {
	service := &TwoFactor{store: store, recovery: recovery, pending: pending, sessions: sessions, now: now, newPending: newPendingID, logger: slog.Default()}
	if len(dependencies) > 0 {
		dependency := dependencies[0]
		service.passwords, service.limiter = dependency.Passwords, dependency.Limiter
		if dependency.Logger != nil {
			service.logger = dependency.Logger
		}
		if dependency.NewPending != nil {
			service.newPending = dependency.NewPending
		}
	}
	return service
}

func (service *TwoFactor) EnrollTwoFactor(ctx context.Context, userID string, accounts ...string) (EnrollTwoFactorResult, error) {
	state, err := service.store.GetTwoFactor(ctx, userID)
	if err != nil {
		return EnrollTwoFactorResult{}, err
	}
	if state.Enabled {
		return EnrollTwoFactorResult{}, domain.ErrTwoFactorAlreadyEnabled
	}
	secret, err := domain.GenerateTOTPSecret()
	if err != nil {
		return EnrollTwoFactorResult{}, domain.ErrDependencyUnavailable
	}
	account := userID
	if len(accounts) > 0 && accounts[0] != "" {
		account = accounts[0]
	}
	uri, err := domain.TOTPURI("CourseForge", account, secret)
	if err != nil {
		return EnrollTwoFactorResult{}, domain.ErrDependencyUnavailable
	}
	if err := service.store.SetTotpSecret(ctx, userID, secret); err != nil {
		return EnrollTwoFactorResult{}, err
	}
	return EnrollTwoFactorResult{Secret: secret, URI: uri}, nil
}

func (service *TwoFactor) ConfirmTwoFactor(ctx context.Context, userID, code string) (ConfirmTwoFactorResult, error) {
	state, err := service.store.GetTwoFactor(ctx, userID)
	if err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	if state.Enabled {
		return ConfirmTwoFactorResult{}, domain.ErrTwoFactorAlreadyEnabled
	}
	// Consume the enrolment code's step as well, so the very code that proved enrolment cannot be
	// immediately replayed as the first login (SA13-B / I-F5).
	proven, err := service.verifyAndConsumeTOTP(ctx, userID, state.Secret, code)
	if err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	if !proven {
		return ConfirmTwoFactorResult{}, domain.ErrInvalidTOTP
	}
	if err := service.store.EnableTotp(ctx, userID); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	codes, err := domain.GenerateRecoveryCodes()
	if err != nil {
		return ConfirmTwoFactorResult{}, domain.ErrDependencyUnavailable
	}
	if err := service.recovery.InsertCodes(ctx, userID, codes); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	return ConfirmTwoFactorResult{RecoveryCodes: codes}, nil
}

func (service *TwoFactor) StartLogin(ctx context.Context, userID string, rememberMe bool, loginRateLimitKey string) (string, error) {
	if service.pending == nil {
		return "", domain.ErrDependencyUnavailable
	}
	pendingID, err := service.newPending()
	if err != nil {
		return "", domain.ErrDependencyUnavailable
	}
	pending := ports.PendingTwoFactor{UserID: userID, RememberMe: rememberMe, LoginRateLimitKey: loginRateLimitKey}
	if err := service.pending.Put(ctx, pendingID, pending, pendingTwoFactorTTL); err != nil {
		return "", err
	}
	return pendingID, nil
}

func (service *TwoFactor) CompleteTwoFactor(ctx context.Context, input CompleteTwoFactorInput) (CompleteTwoFactorResult, error) {
	if service.pending == nil || input.PendingID == "" {
		return CompleteTwoFactorResult{}, domain.ErrAuthenticationRequired
	}
	pending, ok, err := service.pending.Take(ctx, input.PendingID)
	if err != nil {
		return CompleteTwoFactorResult{}, err
	}
	if !ok {
		return CompleteTwoFactorResult{}, domain.ErrAuthenticationRequired
	}
	// Key the 2FA attempt budget on the account, not the ephemeral pending id: an
	// attacker who knows the password mints a fresh pending id per login
	// (identity.Login), so a per-pending budget would refill each attempt (audit
	// H1 / SA01 Bug B). Take is atomic, so restore the pending on a rate-limit —
	// as the failed-proof branch below does — to keep a legitimate user's budget.
	if err := service.allow(ctx, "two-factor:"+pending.UserID); err != nil {
		if putErr := service.pending.Put(ctx, input.PendingID, pending, pendingTwoFactorTTL); putErr != nil {
			return CompleteTwoFactorResult{}, putErr
		}
		return CompleteTwoFactorResult{}, err
	}
	state, err := service.store.GetTwoFactor(ctx, pending.UserID)
	if err != nil {
		return CompleteTwoFactorResult{}, err
	}
	valid := false
	if state.Enabled {
		// A matched TOTP is accepted only if its step was not already consumed; a replay yields
		// valid=false and falls through to the recovery-code check, so it surfaces as a plain
		// ErrInvalidTOTP with no replay-versus-wrong-code oracle (SA13-B / I-F5).
		valid, err = service.verifyAndConsumeTOTP(ctx, pending.UserID, state.Secret, input.Code)
		if err != nil {
			return CompleteTwoFactorResult{}, err
		}
	}
	if !valid {
		valid, err = service.recovery.ConsumeCode(ctx, pending.UserID, input.Code)
		if err != nil {
			return CompleteTwoFactorResult{}, err
		}
	}
	if !valid {
		// Take is atomic. Put the same state back only after a failed proof so a
		// legitimate user has the bounded retry budget; a successful proof never
		// restores it.
		if err := service.pending.Put(ctx, input.PendingID, pending, pendingTwoFactorTTL); err != nil {
			return CompleteTwoFactorResult{}, err
		}
		return CompleteTwoFactorResult{}, domain.ErrInvalidTOTP
	}
	lifetime := sessionLifetime
	if pending.RememberMe {
		lifetime = RememberMeLifetime
	}
	now := service.now().UTC().Truncate(storeResolution)
	issued, err := service.sessions.Rotate(ctx, pending.UserID, input.PresentedToken, now, now.Add(lifetime))
	if err != nil {
		return CompleteTwoFactorResult{}, err
	}
	session, err := validTwoFactorIssuedSession(issued, pending.UserID, now, lifetime)
	if err != nil {
		return CompleteTwoFactorResult{}, err
	}
	// A successful second factor clears both throttles for this account: the 2FA
	// attempt budget (so a legitimate user who logs in several times in the window
	// is not blocked by their own successful logins) and the login throttle
	// carried from the password step (SA01 Bug A). A brute-forcer never reaches
	// success, so never clears either.
	if service.limiter != nil {
		if err := service.limiter.Reset(ctx, "two-factor:"+pending.UserID); err != nil {
			return CompleteTwoFactorResult{}, err
		}
		if pending.LoginRateLimitKey != "" {
			if err := service.limiter.Reset(ctx, pending.LoginRateLimitKey); err != nil {
				return CompleteTwoFactorResult{}, err
			}
		}
	}
	return CompleteTwoFactorResult{Session: session, Token: issued.Token}, nil
}

func (service *TwoFactor) DisableTwoFactor(ctx context.Context, userID, password, code string) error {
	if service.passwords == nil {
		return domain.ErrDependencyUnavailable
	}
	if err := service.passwords.VerifyPassword(ctx, userID, password); err != nil {
		return err
	}
	if err := service.authorizeFactor(ctx, userID, code); err != nil {
		return err
	}
	if err := service.store.DisableTotp(ctx, userID); err != nil {
		return err
	}
	return service.recovery.DeleteCodes(ctx, userID)
}

func (service *TwoFactor) RegenerateRecoveryCodes(ctx context.Context, userID, code string) (ConfirmTwoFactorResult, error) {
	if err := service.authorizeFactor(ctx, userID, code); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	if err := service.recovery.DeleteCodes(ctx, userID); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	codes, err := domain.GenerateRecoveryCodes()
	if err != nil {
		return ConfirmTwoFactorResult{}, domain.ErrDependencyUnavailable
	}
	if err := service.recovery.InsertCodes(ctx, userID, codes); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	return ConfirmTwoFactorResult{RecoveryCodes: codes}, nil
}

func (service *TwoFactor) AdminDisableTwoFactor(ctx context.Context, actor domain.User, targetUserID string) error {
	if !actor.Role.CanAdministerAccounts() {
		return domain.ErrForbidden
	}
	if err := service.store.DisableTotp(ctx, targetUserID); err != nil {
		return err
	}
	if err := service.recovery.DeleteCodes(ctx, targetUserID); err != nil {
		return err
	}
	service.logger.Info("identity two-factor disabled by administrator", "actor_id", actor.ID, "target_user_id", targetUserID)
	return nil
}

func (service *TwoFactor) authorizeFactor(ctx context.Context, userID, code string) error {
	state, err := service.store.GetTwoFactor(ctx, userID)
	if err != nil {
		return err
	}
	if !state.Enabled {
		return domain.ErrTwoFactorNotEnabled
	}
	// A fresh TOTP proof authorizes the change; a replay of an already-used step does not and
	// falls through to the recovery-code check (SA13-B / I-F5).
	proven, err := service.verifyAndConsumeTOTP(ctx, userID, state.Secret, code)
	if err != nil {
		return err
	}
	if proven {
		return nil
	}
	used, err := service.recovery.ConsumeCode(ctx, userID, code)
	if err != nil {
		return err
	}
	if !used {
		return domain.ErrRecoveryCodeInvalid
	}
	return nil
}

// verifyAndConsumeTOTP reports whether code is a currently-valid TOTP proof for the account whose
// step has not been used before. On a match it atomically advances the stored step (via
// AdvanceTotpCounter), so the same code can never authorize two actions — the core of the I-F5
// replay fix. A wrong code or a replay both return false (no error), letting callers fall back to a
// recovery code and surface a uniform failure.
func (service *TwoFactor) verifyAndConsumeTOTP(ctx context.Context, userID, secret, code string) (bool, error) {
	counter, ok := domain.VerifyTOTPAt(secret, code, service.now())
	if !ok {
		return false, nil
	}
	return service.store.AdvanceTotpCounter(ctx, userID, counter)
}
func (service *TwoFactor) allow(ctx context.Context, key string) error {
	if service.limiter == nil {
		return nil
	}
	allowed, err := service.limiter.Allow(ctx, key)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	return nil
}
func newPendingID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

func validTwoFactorIssuedSession(issued ports.IssuedSession, userID string, createdAt time.Time, lifetime time.Duration) (domain.Session, error) {
	session, err := domain.NewSession(issued.Session)
	if err != nil || session.UserID != userID || !session.CreatedAt.Equal(createdAt) || !session.ExpiresAt.Equal(createdAt.Add(lifetime)) || !validToken(issued.Token) {
		return domain.Session{}, domain.ErrDependencyUnavailable
	}
	return session, nil
}
