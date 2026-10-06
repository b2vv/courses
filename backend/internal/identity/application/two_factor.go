package application

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
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
	if err := domain.ValidateTwoFactorEnrollment(state.Enabled, state.Secret, code, service.now()); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	if err := service.store.EnableTotp(ctx, userID); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	codes, err := domain.GenerateRecoveryCodes()
	if err != nil {
		return ConfirmTwoFactorResult{}, domain.ErrDependencyUnavailable
	}
	if err := service.recovery.InsertCodes(ctx, userID, recoveryDigests(codes)); err != nil {
		return ConfirmTwoFactorResult{}, err
	}
	return ConfirmTwoFactorResult{RecoveryCodes: codes}, nil
}

func (service *TwoFactor) StartLogin(ctx context.Context, userID string, rememberMe bool) (string, error) {
	if service.pending == nil {
		return "", domain.ErrDependencyUnavailable
	}
	pendingID, err := service.newPending()
	if err != nil {
		return "", domain.ErrDependencyUnavailable
	}
	if err := service.pending.Put(ctx, pendingID, ports.PendingTwoFactor{UserID: userID, RememberMe: rememberMe}, pendingTwoFactorTTL); err != nil {
		return "", err
	}
	return pendingID, nil
}

func (service *TwoFactor) CompleteTwoFactor(ctx context.Context, input CompleteTwoFactorInput) (CompleteTwoFactorResult, error) {
	if service.pending == nil || input.PendingID == "" {
		return CompleteTwoFactorResult{}, domain.ErrAuthenticationRequired
	}
	if err := service.allow(ctx, "two-factor:"+input.PendingID); err != nil {
		return CompleteTwoFactorResult{}, err
	}
	pending, ok, err := service.pending.Take(ctx, input.PendingID)
	if err != nil {
		return CompleteTwoFactorResult{}, err
	}
	if !ok {
		return CompleteTwoFactorResult{}, domain.ErrAuthenticationRequired
	}
	state, err := service.store.GetTwoFactor(ctx, pending.UserID)
	if err != nil {
		return CompleteTwoFactorResult{}, err
	}
	valid := state.Enabled && domain.VerifyTOTP(state.Secret, input.Code, service.now())
	if !valid {
		valid, err = service.recovery.ConsumeCode(ctx, pending.UserID, recoveryDigest(input.Code))
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
	if err := service.recovery.InsertCodes(ctx, userID, recoveryDigests(codes)); err != nil {
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
	if domain.VerifyTOTP(state.Secret, code, service.now()) {
		return nil
	}
	consumed, err := service.recovery.ConsumeCode(ctx, userID, recoveryDigest(code))
	if err != nil {
		return err
	}
	return domain.ValidateTwoFactorDisable(true, state.Secret, code, consumed, service.now())
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
func recoveryDigest(code string) [32]byte { return sha256.Sum256([]byte(code)) }
func recoveryDigests(codes []string) [][32]byte {
	digests := make([][32]byte, len(codes))
	for i, code := range codes {
		digests[i] = recoveryDigest(code)
	}
	return digests
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
