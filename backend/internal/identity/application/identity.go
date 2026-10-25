package application

import (
	"context"
	"encoding/base64"
	"strings"
	"time"

	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

const (
	sessionLifetime    = 12 * time.Hour
	RememberMeLifetime = 3 * 24 * time.Hour
	operationTimeout   = 3 * time.Second
	// Session timestamps are written to the session store and read back before
	// they are trusted. Stores persist them at microsecond resolution, so a
	// clock reading is normalized before it becomes part of that round trip.
	storeResolution = time.Microsecond
)

type Clock func() time.Time

type Identity struct {
	credentials ports.CredentialVerifier
	sessions    ports.SessionStore
	limiter     ports.RateLimiter
	twoFactor   *TwoFactor
	now         Clock
}

type LoginInput struct {
	Email          string
	Password       string
	PresentedToken string
	RememberMe     bool
	IP             string
}

type LoginResult struct {
	User      domain.User
	Session   domain.Session
	Token     string
	PendingID string
}

// WithTwoFactor adds the optional second-factor step. It is a separate
// composition-root concern until the persistence adapter exists in A13b.
func (identity *Identity) WithTwoFactor(twoFactor *TwoFactor) *Identity {
	identity.twoFactor = twoFactor
	return identity
}

func NewIdentity(credentials ports.CredentialVerifier, sessions ports.SessionStore, now Clock, limiters ...ports.RateLimiter) *Identity {
	identity := &Identity{credentials: credentials, sessions: sessions, now: now}
	if len(limiters) > 0 {
		identity.limiter = limiters[0]
	}
	return identity
}

func (identity *Identity) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email, err := domain.NormalizeEmail(input.Email)
	if err != nil || len(input.Password) == 0 || len(input.Password) > 1024 {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	ctx, cancel := context.WithTimeout(ctx, operationTimeout)
	defer cancel()
	if err := identity.allow(ctx, loginRateLimitKey(input.Email, input.IP)); err != nil {
		return LoginResult{}, err
	}
	user, err := identity.credentials.VerifyCredentials(ctx, email, input.Password)
	if err != nil {
		return LoginResult{}, err
	}
	if _, err := domain.NewUser(user); err != nil || !user.Active() {
		return LoginResult{}, domain.ErrInvalidCredentials
	}
	if identity.twoFactor != nil {
		state, err := identity.twoFactor.store.GetTwoFactor(ctx, user.ID)
		if err != nil {
			return LoginResult{}, err
		}
		if state.Enabled {
			// Carry the login throttle key into the pending so CompleteTwoFactor
			// resets it on success only. Resetting here would clear the 5/15min
			// login counter on every correct password, letting an attacker who
			// knows the password brute-force the second factor (audit H1 / SA01).
			pendingID, err := identity.twoFactor.StartLogin(ctx, user.ID, input.RememberMe, loginRateLimitKey(input.Email, input.IP))
			if err != nil {
				return LoginResult{}, err
			}
			return LoginResult{User: user, PendingID: pendingID}, domain.ErrTwoFactorRequired
		}
	}
	now := identity.timestamp()
	lifetime := sessionLifetime
	if input.RememberMe {
		lifetime = RememberMeLifetime
	}
	issued, err := identity.sessions.Rotate(ctx, user.ID, input.PresentedToken, now, now.Add(lifetime))
	if err != nil {
		return LoginResult{}, err
	}
	session, err := domain.NewSession(issued.Session)
	if err != nil || session.UserID != user.ID || !session.CreatedAt.Equal(now) || !session.ExpiresAt.Equal(now.Add(lifetime)) || !validToken(issued.Token) {
		return LoginResult{}, domain.ErrDependencyUnavailable
	}
	if identity.limiter != nil {
		if err := identity.limiter.Reset(ctx, loginRateLimitKey(input.Email, input.IP)); err != nil {
			return LoginResult{}, err
		}
	}
	return LoginResult{User: user, Session: session, Token: issued.Token}, nil
}

func (identity *Identity) allow(ctx context.Context, key string) error {
	if identity.limiter == nil {
		return nil
	}
	allowed, err := identity.limiter.Allow(ctx, key)
	if err != nil {
		return err
	}
	if !allowed {
		return domain.ErrRateLimited
	}
	return nil
}

func loginRateLimitKey(email, ip string) string {
	return "login:" + strings.ToLower(strings.TrimSpace(email)) + ":" + ip
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
