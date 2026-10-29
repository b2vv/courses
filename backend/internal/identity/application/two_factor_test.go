package application_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha1" // #nosec G505 -- RFC 6238 test vector.
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"testing"
	"time"

	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

const twoFactorUserID = "40000000-0000-4000-8000-000000000001"
const twoFactorSecret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

type twoFactorStore struct {
	state                  ports.TwoFactor
	set, enabled, disabled int
	lastCounter            int64
}

func (s *twoFactorStore) GetTwoFactor(context.Context, string) (ports.TwoFactor, error) {
	return s.state, nil
}

// AdvanceTotpCounter mirrors the repository's conditional UPDATE: a step is consumed (true) only when
// it is strictly newer than the last recorded one, so a replayed code returns false.
func (s *twoFactorStore) AdvanceTotpCounter(_ context.Context, _ string, counter int64) (bool, error) {
	if counter <= s.lastCounter {
		return false, nil
	}
	s.lastCounter = counter
	return true, nil
}
func (s *twoFactorStore) SetTotpSecret(_ context.Context, _ string, secret string) error {
	s.set++
	s.state.Secret = secret
	s.state.Enabled = false
	return nil
}
func (s *twoFactorStore) EnableTotp(context.Context, string) error {
	s.enabled++
	s.state.Enabled = true
	return nil
}
func (s *twoFactorStore) DisableTotp(context.Context, string) error {
	s.disabled++
	s.state = ports.TwoFactor{}
	s.lastCounter = 0 // mirrors DisableUserTotp resetting totp_last_counter
	return nil
}

type recoveryStore struct {
	codes   map[[32]byte]bool
	deleted int
}

func (s *recoveryStore) InsertCodes(_ context.Context, _ string, digests [][32]byte) error {
	if s.codes == nil {
		s.codes = map[[32]byte]bool{}
	}
	for _, d := range digests {
		s.codes[d] = false
	}
	return nil
}
func (s *recoveryStore) ConsumeCode(_ context.Context, _ string, digest [32]byte) (bool, error) {
	used, ok := s.codes[digest]
	if !ok || used {
		return false, nil
	}
	s.codes[digest] = true
	return true, nil
}
func (s *recoveryStore) DeleteCodes(context.Context, string) error {
	s.deleted++
	s.codes = map[[32]byte]bool{}
	return nil
}

type pendingStore struct {
	values map[string]ports.PendingTwoFactor
}

type passwordVerifier struct {
	err   error
	calls int
}

func (v *passwordVerifier) VerifyPassword(context.Context, string, string) error {
	v.calls++
	return v.err
}

func (s *pendingStore) Put(_ context.Context, id string, data ports.PendingTwoFactor, _ time.Duration) error {
	if s.values == nil {
		s.values = map[string]ports.PendingTwoFactor{}
	}
	s.values[id] = data
	return nil
}
func (s *pendingStore) Take(_ context.Context, id string) (ports.PendingTwoFactor, bool, error) {
	data, ok := s.values[id]
	delete(s.values, id)
	return data, ok, nil
}

func TestConfirmTwoFactorEnablesAndReturnsRecoveryCodesOnce(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store, recovery := &twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret}}, &recoveryStore{}
	service := application.NewTwoFactor(store, recovery, nil, &sessions{}, func() time.Time { return now })
	result, err := service.ConfirmTwoFactor(context.Background(), twoFactorUserID, totp(twoFactorSecret, now))
	if err != nil {
		t.Fatal(err)
	}
	if !store.state.Enabled || len(result.RecoveryCodes) != domain.RecoveryCodeCount || len(recovery.codes) != domain.RecoveryCodeCount {
		t.Fatalf("confirmation not persisted: %+v", result)
	}
}

func TestEnrollTwoFactorPersistsAnUnconfirmedSecret(t *testing.T) {
	store := &twoFactorStore{}
	service := application.NewTwoFactor(store, &recoveryStore{}, nil, &sessions{}, time.Now)
	result, err := service.EnrollTwoFactor(context.Background(), twoFactorUserID)
	if err != nil || result.Secret == "" || result.URI == "" || store.set != 1 || store.state.Enabled {
		t.Fatalf("enroll = %+v, %v; state=%+v", result, err, store.state)
	}
}

func TestCompleteTwoFactorConsumesPendingAndCanUseTOTP(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	recovery := &recoveryStore{}
	store := &twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}
	pending := &pendingStore{}
	issued := ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: twoFactorUserID, CreatedAt: now, ExpiresAt: now.Add(3 * 24 * time.Hour)}, Token: validToken('A')}
	service := application.NewTwoFactor(store, recovery, pending, &sessions{issued: issued}, func() time.Time { return now })
	if err := pending.Put(context.Background(), "pending", ports.PendingTwoFactor{UserID: twoFactorUserID, RememberMe: true}, time.Minute); err != nil {
		t.Fatal(err)
	}
	result, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "pending", Code: totp(twoFactorSecret, now)})
	if err != nil || result.Token != issued.Token {
		t.Fatalf("complete = %+v, %v", result, err)
	}
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "pending", Code: totp(twoFactorSecret, now)}); !errors.Is(err, domain.ErrAuthenticationRequired) {
		t.Fatalf("reused pending: %v", err)
	}
}

// TestCompleteTwoFactorRejectsReplayedTOTP pins SA13-B / I-F5: a TOTP code valid within its window
// is accepted once and then refused on a second login (fresh pending, same clock), because its step
// was consumed. Without the consume the same code would log in repeatedly for ~90s.
func TestCompleteTwoFactorRejectsReplayedTOTP(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store := &twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}
	pending := &pendingStore{}
	// RememberMe is left false, so the completed session must carry the 12-hour ordinary lifetime;
	// a 72-hour issued session here would fail validTwoFactorIssuedSession as dependency_unavailable.
	issued := ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: twoFactorUserID, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: validToken('A')}
	service := application.NewTwoFactor(store, &recoveryStore{}, pending, &sessions{issued: issued}, func() time.Time { return now })
	code := totp(twoFactorSecret, now)
	if err := pending.Put(context.Background(), "p1", ports.PendingTwoFactor{UserID: twoFactorUserID}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "p1", Code: code}); err != nil {
		t.Fatalf("first login: %v", err)
	}
	// A brand-new pending (as a real attacker replaying the captured code would mint) with the same
	// still-in-window code must now fail: the step is already consumed.
	if err := pending.Put(context.Background(), "p2", ports.PendingTwoFactor{UserID: twoFactorUserID}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "p2", Code: code}); !errors.Is(err, domain.ErrInvalidTOTP) {
		t.Fatalf("replayed TOTP = %v, want ErrInvalidTOTP", err)
	}
}

func TestCompleteTwoFactorConsumesARecoveryCode(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	recoveryCode := "ABCD-EFGH"
	recovery := &recoveryStore{codes: map[[32]byte]bool{sha256.Sum256([]byte(recoveryCode)): false}}
	pending := &pendingStore{}
	issued := ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: twoFactorUserID, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: validToken('A')}
	service := application.NewTwoFactor(&twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, recovery, pending, &sessions{issued: issued}, func() time.Time { return now })
	if err := pending.Put(context.Background(), "pending", ports.PendingTwoFactor{UserID: twoFactorUserID}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "pending", Code: recoveryCode}); err != nil {
		t.Fatal(err)
	}
	if !recovery.codes[sha256.Sum256([]byte(recoveryCode))] {
		t.Fatal("recovery code was not consumed")
	}
}

func TestAdminDisableTwoFactorRequiresAdmin(t *testing.T) {
	service := application.NewTwoFactor(&twoFactorStore{}, &recoveryStore{}, nil, &sessions{}, time.Now)
	err := service.AdminDisableTwoFactor(context.Background(), domain.User{Role: domain.Student}, twoFactorUserID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err = %v", err)
	}
}

func TestAdminDisableTwoFactorRemovesSecretAndRecoveryCodes(t *testing.T) {
	store, recovery := &twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, &recoveryStore{}
	service := application.NewTwoFactor(store, recovery, nil, &sessions{}, time.Now)
	actor := domain.User{ID: "40000000-0000-4000-8000-000000000002", Role: domain.Admin}
	if err := service.AdminDisableTwoFactor(context.Background(), actor, twoFactorUserID); err != nil {
		t.Fatal(err)
	}
	if store.disabled != 1 || recovery.deleted != 1 {
		t.Fatalf("admin disable calls: store=%+v recovery=%+v", store, recovery)
	}
}

func TestDisableAndRegenerateRequireASecondFactor(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	store, recovery, verifier := &twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, &recoveryStore{}, &passwordVerifier{}
	service := application.NewTwoFactor(store, recovery, nil, &sessions{}, func() time.Time { return now }, application.TwoFactorDependency{Passwords: verifier})
	if err := service.DisableTwoFactor(context.Background(), twoFactorUserID, "password", totp(twoFactorSecret, now)); err != nil {
		t.Fatal(err)
	}
	if store.disabled != 1 || recovery.deleted != 1 || verifier.calls != 1 {
		t.Fatalf("disable calls: store=%+v recovery=%+v verifier=%+v", store, recovery, verifier)
	}
	store.state = ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}
	result, err := service.RegenerateRecoveryCodes(context.Background(), twoFactorUserID, totp(twoFactorSecret, now))
	if err != nil || len(result.RecoveryCodes) != domain.RecoveryCodeCount || len(recovery.codes) != domain.RecoveryCodeCount {
		t.Fatalf("regenerate = %+v, %v", result, err)
	}
}

func TestCompleteTwoFactorRateLimitsRepeatedInvalidProofs(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	pending, attempts := &pendingStore{}, &limiter{allowed: true}
	service := application.NewTwoFactor(&twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, &recoveryStore{}, pending, &sessions{}, func() time.Time { return now }, application.TwoFactorDependency{Limiter: attempts})
	if err := pending.Put(context.Background(), "pending", ports.PendingTwoFactor{UserID: twoFactorUserID}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "pending", Code: "000000"}); !errors.Is(err, domain.ErrInvalidTOTP) {
		t.Fatal(err)
	}
	attempts.allowed = false
	// The 2FA attempt budget must be keyed on the account, not the ephemeral
	// pending id: otherwise an attacker who knows the password mints a fresh
	// pending id per login (identity.Login) and refills the TOTP guess budget
	// each time (audit H1 / SA01 Bug B).
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "pending", Code: "000000"}); !errors.Is(err, domain.ErrRateLimited) || attempts.key != "two-factor:"+twoFactorUserID {
		t.Fatalf("err=%v key=%q", err, attempts.key)
	}
}

// TestCompleteTwoFactorResetsLoginLimitOnSuccess pins SA01 Option 1: a
// successful second factor must reset the login rate limiter using the key that
// Login carried into the pending. This is the only point where the 2FA-required
// login path is allowed to clear the login throttle.
func TestCompleteTwoFactorResetsLoginLimitOnSuccess(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	loginKey := "login:learner@example.com:203.0.113.7"
	limit := &limiter{allowed: true}
	pending := &pendingStore{}
	issued := ports.IssuedSession{Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: twoFactorUserID, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: validToken('A')}
	service := application.NewTwoFactor(&twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, &recoveryStore{}, pending, &sessions{issued: issued}, func() time.Time { return now }, application.TwoFactorDependency{Limiter: limit})
	if err := pending.Put(context.Background(), "pending", ports.PendingTwoFactor{UserID: twoFactorUserID, LoginRateLimitKey: loginKey}, time.Minute); err != nil {
		t.Fatal(err)
	}
	if _, err := service.CompleteTwoFactor(context.Background(), application.CompleteTwoFactorInput{PendingID: "pending", Code: totp(twoFactorSecret, now)}); err != nil {
		t.Fatal(err)
	}
	// Success clears both the login throttle (Bug A) and the per-account 2FA
	// attempt budget (so repeated successful logins are not self-throttled).
	reset := map[string]bool{}
	for _, key := range limit.resetKeys {
		reset[key] = true
	}
	if !reset[loginKey] || !reset["two-factor:"+twoFactorUserID] {
		t.Fatalf("success must reset both the login key and the per-account 2FA key: resetKeys=%v", limit.resetKeys)
	}
}

// TestLoginDoesNotResetRateLimitUntilSecondFactor pins SA01 Bug A: the
// 2FA-required branch of Login must NOT reset the login rate limiter. Resetting
// it there lets an attacker who knows the password log in unboundedly (each call
// succeeds on the password and clears the 5/15min login counter), minting a fresh
// pending id every time and defeating the login throttle entirely.
func TestLoginDoesNotResetRateLimitUntilSecondFactor(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	user := validUser(now)
	loginLimiter := &limiter{allowed: true}
	twoFactor := application.NewTwoFactor(&twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, &recoveryStore{}, &pendingStore{}, &sessions{}, func() time.Time { return now }, application.TwoFactorDependency{NewPending: func() (string, error) { return "pending", nil }})
	identity := application.NewIdentity(&credentials{user: user}, &sessions{}, func() time.Time { return now }, loginLimiter).WithTwoFactor(twoFactor)
	if _, err := identity.Login(context.Background(), application.LoginInput{Email: user.Email, Password: "password"}); !errors.Is(err, domain.ErrTwoFactorRequired) {
		t.Fatalf("login err=%v, want ErrTwoFactorRequired", err)
	}
	if loginLimiter.resetCalls != 0 {
		t.Fatalf("login reset the rate limiter on the 2FA-required branch (resetCalls=%d); it must stay counted until the second factor completes", loginLimiter.resetCalls)
	}
}

func TestLoginDefersSessionUntilSecondFactor(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	user := validUser(now)
	pending := &pendingStore{}
	service := application.NewTwoFactor(&twoFactorStore{state: ports.TwoFactor{Secret: twoFactorSecret, Enabled: true}}, &recoveryStore{}, pending, &sessions{}, func() time.Time { return now }, application.TwoFactorDependency{NewPending: func() (string, error) { return "pending", nil }})
	identity := application.NewIdentity(&credentials{user: user}, &sessions{}, func() time.Time { return now }).WithTwoFactor(service)
	result, err := identity.Login(context.Background(), application.LoginInput{Email: user.Email, Password: "password", RememberMe: true})
	if !errors.Is(err, domain.ErrTwoFactorRequired) || result.PendingID != "pending" || pending.values["pending"].UserID != user.ID || !pending.values["pending"].RememberMe {
		t.Fatalf("login result=%+v err=%v pending=%+v", result, err, pending.values)
	}
}

func totp(secret string, at time.Time) string {
	key := []byte("12345678901234567890")
	message := make([]byte, 8)
	binary.BigEndian.PutUint64(message, uint64(at.Unix()/30))
	h := hmac.New(sha1.New, key)
	_, _ = h.Write(message)
	sum := h.Sum(nil)
	_ = secret
	offset := sum[len(sum)-1] & 15
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1_000_000)
}
