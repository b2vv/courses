package domain

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidUser             Error = "invalid_user"
	ErrInvalidEmail            Error = "invalid_email"
	ErrInvalidDisplayName      Error = "invalid_display_name"
	ErrWeakPassword            Error = "weak_password"
	ErrPasswordUnchanged       Error = "password_unchanged"
	ErrEmailTaken              Error = "email_taken"
	ErrTokenUsed               Error = "token_used"
	ErrInvalidAccountToken     Error = "invalid_account_token"
	ErrInvalidSession          Error = "invalid_session"
	ErrInvalidCredentials      Error = "invalid_credentials"
	ErrRateLimited             Error = "rate_limited"
	ErrAuthenticationRequired  Error = "authentication_required"
	ErrForbidden               Error = "forbidden"
	ErrDependencyUnavailable   Error = "dependency_unavailable"
	ErrCannotTargetSelf        Error = "cannot_target_self"
	ErrAdminImmutable          Error = "admin_immutable"
	ErrLastAdmin               Error = "last_admin"
	ErrUserNotFound            Error = "user_not_found"
	ErrInvalidTOTP             Error = "invalid_totp"
	ErrTwoFactorAlreadyEnabled Error = "twofactor_already_enabled"
	ErrTwoFactorNotEnabled     Error = "twofactor_not_enabled"
	ErrRecoveryCodeInvalid     Error = "recovery_code_invalid"
	ErrTwoFactorRequired       Error = "two_factor_required"
)
