package http

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/platform/httpserver"
)

const maximumLoginBody = 8 * 1024

type identityService interface {
	Login(context.Context, application.LoginInput) (application.LoginResult, error)
	Authenticate(context.Context, string) (domain.User, error)
	Logout(context.Context, string) error
}

type accountsService interface {
	Register(context.Context, application.RegisterInput, string) (application.RegisterResult, error)
	VerifyEmail(context.Context, string) error
	ResendVerification(context.Context, string, string) (bool, error)
	RequestPasswordReset(context.Context, string, string, string) error
	ResetPassword(context.Context, string, string) error
	ChangePassword(context.Context, string, string, string, []byte) error
	RequestEmailChange(context.Context, string, string, string) error
	ConfirmEmailChange(context.Context, string) error
}

type twoFactorService interface {
	EnrollTwoFactor(context.Context, string, ...string) (application.EnrollTwoFactorResult, error)
	ConfirmTwoFactor(context.Context, string, string) (application.ConfirmTwoFactorResult, error)
	CompleteTwoFactor(context.Context, application.CompleteTwoFactorInput) (application.CompleteTwoFactorResult, error)
	DisableTwoFactor(context.Context, string, string, string) error
	RegenerateRecoveryCodes(context.Context, string, string) (application.ConfirmTwoFactorResult, error)
	AdminDisableTwoFactor(context.Context, domain.User, string) error
}

type instructorLister interface {
	ListInstructors(context.Context) ([]domain.User, error)
}

type administrationService interface {
	ListUsers(context.Context, application.ListUsersInput) (application.UserPage, error)
	SetStatus(context.Context, domain.Actor, string, domain.AccountStatus) error
	Delete(context.Context, domain.Actor, string) error
	ResetPassword(context.Context, string, string) error
}

type Handler struct {
	identity       identityService
	instructors    instructorLister
	administration administrationService
	config         Config
	accounts       accountsService
	twoFactor      twoFactorService
}

type userKey struct{}

type loginRequest struct {
	Email      string
	Password   string
	RememberMe bool
}

type userDTO struct {
	ID               string `json:"id"`
	Email            string `json:"email"`
	DisplayName      string `json:"displayName"`
	Role             string `json:"role"`
	EmailVerified    bool   `json:"emailVerified"`
	TwoFactorEnabled bool   `json:"twoFactorEnabled"`
}

type instructorDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

func NewHandler(identity identityService, instructors instructorLister, administration administrationService, config Config, accountServices ...accountsService) *Handler {
	handler := &Handler{identity: identity, instructors: instructors, administration: administration, config: config}
	if len(accountServices) > 0 {
		handler.accounts = accountServices[0]
	}
	return handler
}

// WithTwoFactor connects the optional account-security use cases without
// broadening the constructor used by older composition roots and tests.
func (handler *Handler) WithTwoFactor(service twoFactorService) *Handler {
	handler.twoFactor = service
	return handler
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/auth/login", handler.RequireSameOrigin(http.HandlerFunc(handler.login)))
	mux.Handle("POST /api/v1/auth/register", handler.RequireSameOrigin(http.HandlerFunc(handler.register)))
	mux.Handle("POST /api/v1/auth/verify-email", handler.RequireSameOrigin(http.HandlerFunc(handler.verifyEmail)))
	mux.Handle("POST /api/v1/auth/resend-verification", handler.RequireSameOrigin(http.HandlerFunc(handler.resendVerification)))
	mux.Handle("POST /api/v1/auth/password/forgot", handler.RequireSameOrigin(http.HandlerFunc(handler.requestPasswordReset)))
	mux.Handle("POST /api/v1/auth/password/reset", handler.RequireSameOrigin(http.HandlerFunc(handler.resetPassword)))
	mux.Handle("POST /api/v1/auth/password/change", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.changePassword))))
	mux.Handle("POST /api/v1/auth/email/change", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.requestEmailChange))))
	mux.Handle("POST /api/v1/auth/email/change/confirm", handler.RequireSameOrigin(http.HandlerFunc(handler.confirmEmailChange)))
	mux.Handle("POST /api/v1/auth/2fa/enroll", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.enrollTwoFactor))))
	mux.Handle("POST /api/v1/auth/2fa/confirm", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.confirmTwoFactor))))
	mux.Handle("POST /api/v1/auth/2fa/complete", handler.RequireSameOrigin(http.HandlerFunc(handler.completeTwoFactor)))
	mux.Handle("POST /api/v1/auth/2fa/disable", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.disableTwoFactor))))
	mux.Handle("POST /api/v1/auth/2fa/recovery-codes", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.regenerateRecoveryCodes))))
	mux.Handle("GET /api/v1/auth/me", handler.Authenticate(http.HandlerFunc(handler.me)))
	mux.Handle("POST /api/v1/auth/logout", handler.RequireSameOrigin(http.HandlerFunc(handler.logout)))
	mux.Handle("GET /api/v1/admin/instructors", handler.Authenticate(http.HandlerFunc(handler.listInstructors)))
	mux.Handle("GET /api/v1/admin/users", handler.Authenticate(http.HandlerFunc(handler.listUsers)))
	mux.Handle("POST /api/v1/admin/users/{id}/status", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.setUserStatus))))
	mux.Handle("DELETE /api/v1/admin/users/{id}", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.deleteUser))))
	mux.Handle("POST /api/v1/admin/users/{id}/password", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.resetUserPassword))))
	mux.Handle("POST /api/v1/admin/users/{id}/2fa/disable", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.adminDisableTwoFactor))))
}

func (handler *Handler) login(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	input, err := decodeLogin(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	presented, duplicate := sessionCookie(r)
	if duplicate {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	result, err := handler.identity.Login(r.Context(), application.LoginInput{Email: input.Email, Password: input.Password, RememberMe: input.RememberMe, PresentedToken: presented, IP: clientIP(r)})
	if errors.Is(err, domain.ErrTwoFactorRequired) {
		httpserver.WriteJSON(w, http.StatusOK, twoFactorRequiredDTO{Status: string(domain.ErrTwoFactorRequired), PendingID: result.PendingID})
		return
	}
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	http.SetCookie(w, handler.sessionCookie(result.Token, result.Session.ExpiresAt, int(result.Session.ExpiresAt.Sub(result.Session.CreatedAt).Seconds())))
	httpserver.WriteJSON(w, http.StatusOK, mapUser(result.User))
}

type twoFactorRequiredDTO struct {
	Status    string `json:"status"`
	PendingID string `json:"pendingId"`
}
type enrollTwoFactorDTO struct {
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauthUri"`
}
type recoveryCodesDTO struct {
	RecoveryCodes []string `json:"recoveryCodes"`
}
type twoFactorCodeRequest struct{ Code string }
type completeTwoFactorRequest struct {
	PendingID string
	Code      string
}
type disableTwoFactorRequest struct {
	Password string
	Code     string
}

func (handler *Handler) twoFactorUnavailable(w http.ResponseWriter, r *http.Request) bool {
	if handler.twoFactor != nil {
		return false
	}
	writeError(w, r, domain.ErrDependencyUnavailable, "")
	return true
}

func currentUser(w http.ResponseWriter, r *http.Request) (domain.User, bool) {
	user, ok := r.Context().Value(userKey{}).(domain.User)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
	}
	return user, ok
}

func (handler *Handler) enrollTwoFactor(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	user, ok := currentUser(w, r)
	if !ok || handler.twoFactorUnavailable(w, r) {
		return
	}
	result, err := handler.twoFactor.EnrollTwoFactor(r.Context(), user.ID, user.Email)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, enrollTwoFactorDTO{Secret: result.Secret, OTPAuthURI: result.URI})
}

func (handler *Handler) confirmTwoFactor(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	user, ok := currentUser(w, r)
	request, err := decodeTwoFactorCode(w, r, true)
	if !ok || err != nil {
		if err != nil {
			writeError(w, r, err, "invalid_request")
		}
		return
	}
	if handler.twoFactorUnavailable(w, r) {
		return
	}
	result, err := handler.twoFactor.ConfirmTwoFactor(r.Context(), user.ID, request.Code)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, recoveryCodesDTO{RecoveryCodes: result.RecoveryCodes})
}

func (handler *Handler) completeTwoFactor(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	request, err := decodeCompleteTwoFactor(w, r)
	if err != nil {
		writeError(w, r, err, "invalid_request")
		return
	}
	if handler.twoFactorUnavailable(w, r) {
		return
	}
	presented, duplicate := sessionCookie(r)
	if duplicate {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	}
	result, err := handler.twoFactor.CompleteTwoFactor(r.Context(), application.CompleteTwoFactorInput{PendingID: request.PendingID, Code: request.Code, PresentedToken: presented})
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	http.SetCookie(w, handler.sessionCookie(result.Token, result.Session.ExpiresAt, int(result.Session.ExpiresAt.Sub(result.Session.CreatedAt).Seconds())))
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) disableTwoFactor(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	user, ok := currentUser(w, r)
	request, err := decodeDisableTwoFactor(w, r)
	if !ok || err != nil {
		if err != nil {
			writeError(w, r, err, "invalid_request")
		}
		return
	}
	if handler.twoFactorUnavailable(w, r) {
		return
	}
	if err := handler.twoFactor.DisableTwoFactor(r.Context(), user.ID, request.Password, request.Code); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) regenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	user, ok := currentUser(w, r)
	request, err := decodeTwoFactorCode(w, r, true)
	if !ok || err != nil {
		if err != nil {
			writeError(w, r, err, "invalid_request")
		}
		return
	}
	if handler.twoFactorUnavailable(w, r) {
		return
	}
	result, err := handler.twoFactor.RegenerateRecoveryCodes(r.Context(), user.ID, request.Code)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, recoveryCodesDTO{RecoveryCodes: result.RecoveryCodes})
}

func (handler *Handler) register(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	input, err := decodeRegister(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	result, err := handler.accounts.Register(r.Context(), application.RegisterInput{Email: input.Email, Password: input.Password, DisplayName: input.DisplayName}, language(r))
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	http.SetCookie(w, handler.sessionCookie(result.Token, result.Session.ExpiresAt, int(result.Session.ExpiresAt.Sub(result.Session.CreatedAt).Seconds())))
	httpserver.WriteJSON(w, http.StatusCreated, mapUser(result.User))
}

func (handler *Handler) verifyEmail(w http.ResponseWriter, r *http.Request) {
	handler.tokenAction(w, r, func(token string) error { return handler.accounts.VerifyEmail(r.Context(), token) })
}
func (handler *Handler) confirmEmailChange(w http.ResponseWriter, r *http.Request) {
	handler.tokenAction(w, r, func(token string) error { return handler.accounts.ConfirmEmailChange(r.Context(), token) })
}

func (handler *Handler) tokenAction(w http.ResponseWriter, r *http.Request, action func(string) error) {
	identityHeaders(w)
	request, err := decodeToken(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	if err := action(request.Token); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) resendVerification(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	request, err := decodeOptionalEmail(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	identifier := request.Email
	if token, duplicate := sessionCookie(r); duplicate {
		writeError(w, r, domain.ErrInvalidUser, "invalid_request")
		return
	} else if token != "" {
		user, authenticationErr := handler.identity.Authenticate(r.Context(), token)
		if authenticationErr != nil {
			writeError(w, r, authenticationErr, "")
			return
		}
		identifier = user.ID
	}
	if identifier == "" {
		writeError(w, r, domain.ErrInvalidUser, "")
		return
	}
	if _, err := handler.accounts.ResendVerification(r.Context(), identifier, language(r)); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) requestPasswordReset(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	request, err := decodeEmail(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	if err := handler.accounts.RequestPasswordReset(r.Context(), request.Email, clientIP(r), language(r)); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (handler *Handler) resetPassword(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	request, err := decodeResetPassword(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	if err := handler.accounts.ResetPassword(r.Context(), request.Token, request.Password); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) changePassword(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	request, err := decodeChangePassword(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	user, ok := r.Context().Value(userKey{}).(domain.User)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return
	}
	token, duplicate := sessionCookie(r)
	if duplicate || token == "" {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return
	}
	digest, ok := sessionDigest(token)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	if err := handler.accounts.ChangePassword(r.Context(), user.ID, request.OldPassword, request.NewPassword, digest); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) requestEmailChange(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	request, err := decodeNewEmail(w, r)
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	user, ok := r.Context().Value(userKey{}).(domain.User)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return
	}
	if handler.accounts == nil {
		writeError(w, r, domain.ErrDependencyUnavailable, "")
		return
	}
	if err := handler.accounts.RequestEmailChange(r.Context(), user.ID, request.NewEmail, language(r)); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) me(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	user, ok := r.Context().Value(userKey{}).(domain.User)
	if !ok {
		writeError(w, r, domain.ErrAuthenticationRequired, "")
		return
	}
	httpserver.WriteJSON(w, http.StatusOK, mapUser(user))
}

func (handler *Handler) logout(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	tokens := sessionCookies(r)
	http.SetCookie(w, handler.expiredCookie())
	for _, token := range tokens {
		if err := handler.identity.Logout(r.Context(), token); err != nil {
			writeError(w, r, err, "")
			return
		}
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identityHeaders(w)
		token, duplicate := sessionCookie(r)
		if duplicate || token == "" {
			writeError(w, r, domain.ErrAuthenticationRequired, "")
			return
		}
		user, err := handler.identity.Authenticate(r.Context(), token)
		if err != nil {
			writeError(w, r, err, "")
			return
		}
		ctx := context.WithValue(r.Context(), userKey{}, user)
		ctx = httpserver.WithActor(ctx, httpserver.Actor{UserID: user.ID, Role: string(user.Role), EmailVerified: user.EmailVerified})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (handler *Handler) RequireSameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identityHeaders(w)
		values := r.Header.Values("Origin")
		if len(values) != 1 || values[0] == "null" || !handler.config.acceptsOrigin(values[0]) {
			writeError(w, r, domain.ErrForbidden, "origin_forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (handler *Handler) RequireAdministrator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userKey{}).(domain.User)
		if !ok {
			writeError(w, r, domain.ErrAuthenticationRequired, "")
			return
		}
		if !user.Role.CanAdministerAccounts() {
			writeError(w, r, domain.ErrForbidden, "grant_forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (handler *Handler) listInstructors(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	if _, ok := administrator(w, r); !ok {
		return
	}
	users, err := handler.instructors.ListInstructors(r.Context())
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	instructors := make([]instructorDTO, 0, len(users))
	for _, user := range users {
		instructors = append(instructors, instructorDTO{ID: user.ID, DisplayName: user.DisplayName, Role: string(user.Role)})
	}
	httpserver.WriteJSON(w, http.StatusOK, instructors)
}

func (handler *Handler) sessionCookie(value string, expires time.Time, maximumAge int) *http.Cookie {
	return &http.Cookie{Name: SessionCookieName, Value: value, Path: "/", Expires: expires, MaxAge: maximumAge, HttpOnly: true, Secure: handler.config.secureCookie, SameSite: http.SameSiteLaxMode}
}

func (handler *Handler) expiredCookie() *http.Cookie {
	return handler.sessionCookie("", time.Unix(1, 0).UTC(), -1)
}

// Every session cookie the request carried. Callers decide what more than one
// means: login and authentication refuse the request untouched, while logout
// revokes all of them, because leaving a live session behind would make a
// successful logout a lie.
func sessionCookies(r *http.Request) []string {
	var values []string
	for _, cookie := range r.Cookies() {
		if cookie.Name == SessionCookieName {
			values = append(values, cookie.Value)
		}
	}
	return values
}

func sessionCookie(r *http.Request) (string, bool) {
	values := sessionCookies(r)
	if len(values) != 1 {
		return "", len(values) > 1
	}
	return values[0], false
}

func decodeLogin(w http.ResponseWriter, r *http.Request) (loginRequest, error) {
	if mediaType := strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]); mediaType != "application/json" {
		return loginRequest{}, domain.ErrInvalidUser
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumLoginBody)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return loginRequest{}, domain.ErrInvalidUser
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return loginRequest{}, domain.ErrInvalidUser
	}
	seen := make(map[string]bool, 2)
	result := loginRequest{}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || seen[key] {
			return loginRequest{}, domain.ErrInvalidUser
		}
		seen[key] = true
		switch key {
		case "email":
			err = decoder.Decode(&result.Email)
		case "password":
			err = decoder.Decode(&result.Password)
		case "rememberMe":
			err = decoder.Decode(&result.RememberMe)
		default:
			return loginRequest{}, domain.ErrInvalidUser
		}
		if err != nil {
			return loginRequest{}, domain.ErrInvalidUser
		}
	}
	if _, err := decoder.Token(); err != nil || !seen["email"] || !seen["password"] || len(result.Email) > 254 || len(result.Password) > 1024 {
		return loginRequest{}, domain.ErrInvalidUser
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return loginRequest{}, domain.ErrInvalidUser
	}
	return result, nil
}

type registerRequest struct {
	Email       string
	Password    string
	DisplayName string
}
type tokenRequest struct{ Token string }
type emailRequest struct{ Email string }
type resetPasswordRequest struct {
	Token    string
	Password string
}
type changePasswordRequest struct {
	OldPassword string
	NewPassword string
}
type newEmailRequest struct{ NewEmail string }

func decodeRegister(w http.ResponseWriter, r *http.Request) (registerRequest, error) {
	var result registerRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{
		"email":       func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Email) },
		"password":    func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Password) },
		"displayName": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.DisplayName) },
	}, "email", "password", "displayName")
	return result, err
}
func decodeToken(w http.ResponseWriter, r *http.Request) (tokenRequest, error) {
	var result tokenRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{"token": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Token) }}, "token")
	return result, err
}
func decodeEmail(w http.ResponseWriter, r *http.Request) (emailRequest, error) {
	var result emailRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{"email": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Email) }}, "email")
	return result, err
}
func decodeOptionalEmail(w http.ResponseWriter, r *http.Request) (emailRequest, error) {
	var result emailRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{"email": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Email) }})
	return result, err
}
func decodeResetPassword(w http.ResponseWriter, r *http.Request) (resetPasswordRequest, error) {
	var result resetPasswordRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{
		"token": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Token) }, "password": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Password) },
	}, "token", "password")
	return result, err
}
func decodeChangePassword(w http.ResponseWriter, r *http.Request) (changePasswordRequest, error) {
	var result changePasswordRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{
		"oldPassword": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.OldPassword) }, "newPassword": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.NewPassword) },
	}, "oldPassword", "newPassword")
	return result, err
}
func decodeNewEmail(w http.ResponseWriter, r *http.Request) (newEmailRequest, error) {
	var result newEmailRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{"newEmail": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.NewEmail) }}, "newEmail")
	return result, err
}
func decodeTwoFactorCode(w http.ResponseWriter, r *http.Request, required bool) (twoFactorCodeRequest, error) {
	var result twoFactorCodeRequest
	requiredFields := []string{}
	if required {
		requiredFields = append(requiredFields, "code")
	}
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{
		"code": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Code) },
	}, requiredFields...)
	if err != nil || result.Code == "" || len(result.Code) > 128 {
		return twoFactorCodeRequest{}, domain.ErrInvalidUser
	}
	return result, nil
}
func decodeCompleteTwoFactor(w http.ResponseWriter, r *http.Request) (completeTwoFactorRequest, error) {
	var result completeTwoFactorRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{
		"pendingId": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.PendingID) },
		"code":      func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Code) },
	}, "pendingId", "code")
	if err != nil || result.PendingID == "" || len(result.PendingID) > 128 || result.Code == "" || len(result.Code) > 128 {
		return completeTwoFactorRequest{}, domain.ErrInvalidUser
	}
	return result, nil
}
func decodeDisableTwoFactor(w http.ResponseWriter, r *http.Request) (disableTwoFactorRequest, error) {
	var result disableTwoFactorRequest
	err := decodeObject(w, r, map[string]func(json.RawMessage) error{
		"password": func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Password) },
		"code":     func(raw json.RawMessage) error { return json.Unmarshal(raw, &result.Code) },
	}, "password", "code")
	if err != nil || result.Password == "" || len(result.Password) > 1024 || result.Code == "" || len(result.Code) > 128 {
		return disableTwoFactorRequest{}, domain.ErrInvalidUser
	}
	return result, nil
}

func decodeObject(w http.ResponseWriter, r *http.Request, fields map[string]func(json.RawMessage) error, required ...string) error {
	if mediaType := strings.TrimSpace(strings.Split(r.Header.Get("Content-Type"), ";")[0]); mediaType != "application/json" {
		return domain.ErrInvalidUser
	}
	r.Body = http.MaxBytesReader(w, r.Body, maximumLoginBody)
	decoder := json.NewDecoder(r.Body)
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return domain.ErrInvalidUser
	}
	seen := make(map[string]bool, len(fields))
	for decoder.More() {
		key, err := decoder.Token()
		name, ok := key.(string)
		if err != nil || !ok || seen[name] {
			return domain.ErrInvalidUser
		}
		seen[name] = true
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return domain.ErrInvalidUser
		}
		decode, ok := fields[name]
		if !ok || decode(raw) != nil {
			return domain.ErrInvalidUser
		}
	}
	if _, err := decoder.Token(); err != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return domain.ErrInvalidUser
	}
	for _, key := range required {
		if !seen[key] {
			return domain.ErrInvalidUser
		}
	}
	return nil
}

func language(r *http.Request) string {
	if strings.HasPrefix(strings.ToLower(r.Header.Get("Accept-Language")), "en") {
		return "en"
	}
	return "uk"
}
func clientIP(r *http.Request) string {
	host := r.RemoteAddr
	if index := strings.LastIndex(host, ":"); index > 0 {
		return host[:index]
	}
	return host
}
func sessionDigest(token string) ([]byte, bool) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(raw) != 32 {
		return nil, false
	}
	digest := sha256.Sum256(raw)
	return digest[:], true
}

func identityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

func mapUser(user domain.User) userDTO {
	return userDTO{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: string(user.Role), EmailVerified: user.EmailVerified, TwoFactorEnabled: user.TwoFactorEnabled}
}

func writeError(w http.ResponseWriter, r *http.Request, err error, override string) {
	switch {
	case override == "origin_forbidden":
		httpserver.WriteError(w, r, http.StatusForbidden, override, "Request origin is not allowed")
	case override == "invalid_request", errors.Is(err, domain.ErrInvalidUser):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_request", "Request body is invalid")
	case errors.Is(err, domain.ErrInvalidCredentials):
		httpserver.WriteError(w, r, http.StatusUnauthorized, "invalid_credentials", "Email or password is invalid")
	case errors.Is(err, domain.ErrAuthenticationRequired):
		httpserver.WriteError(w, r, http.StatusUnauthorized, "authentication_required", "Authentication is required")
	case override == "forbidden":
		httpserver.WriteError(w, r, http.StatusForbidden, "forbidden", "Permission is required")
	case override == "grant_forbidden":
		httpserver.WriteError(w, r, http.StatusForbidden, "grant_forbidden", "Administrator permission is required")
	case errors.Is(err, domain.ErrUserNotFound):
		httpserver.WriteError(w, r, http.StatusNotFound, "user_not_found", "User was not found")
	case errors.Is(err, domain.ErrCannotTargetSelf):
		httpserver.WriteError(w, r, http.StatusConflict, "cannot_target_self", "An administrator cannot target their own account")
	case errors.Is(err, domain.ErrAdminImmutable):
		httpserver.WriteError(w, r, http.StatusConflict, "admin_immutable", "Administrator accounts cannot be deleted")
	case errors.Is(err, domain.ErrLastAdmin):
		httpserver.WriteError(w, r, http.StatusConflict, "last_admin", "The last active administrator cannot be blocked")
	case errors.Is(err, domain.ErrEmailTaken):
		httpserver.WriteError(w, r, http.StatusConflict, "email_taken", "Email is already registered")
	case errors.Is(err, domain.ErrWeakPassword):
		httpserver.WriteError(w, r, http.StatusUnprocessableEntity, "weak_password", "Request validation failed")
	case errors.Is(err, domain.ErrInvalidEmail):
		httpserver.WriteError(w, r, http.StatusUnprocessableEntity, "invalid_email", "Request validation failed")
	case errors.Is(err, domain.ErrInvalidDisplayName):
		httpserver.WriteError(w, r, http.StatusUnprocessableEntity, "invalid_display_name", "Request validation failed")
	case errors.Is(err, domain.ErrPasswordUnchanged):
		httpserver.WriteError(w, r, http.StatusUnprocessableEntity, "password_unchanged", "Request validation failed")
	case errors.Is(err, domain.ErrTokenUsed):
		httpserver.WriteError(w, r, http.StatusGone, "token_used", "Token has expired or was already used")
	case errors.Is(err, domain.ErrInvalidAccountToken):
		httpserver.WriteError(w, r, http.StatusBadRequest, "invalid_account_token", "Token is invalid")
	case errors.Is(err, domain.ErrRateLimited):
		httpserver.WriteError(w, r, http.StatusTooManyRequests, "rate_limited", "Too many requests")
	case errors.Is(err, domain.ErrInvalidTOTP):
		httpserver.WriteError(w, r, http.StatusUnprocessableEntity, "invalid_totp", "Two-factor code is invalid")
	case errors.Is(err, domain.ErrRecoveryCodeInvalid):
		httpserver.WriteError(w, r, http.StatusUnauthorized, "recovery_code_invalid", "Recovery code is invalid")
	case errors.Is(err, domain.ErrTwoFactorAlreadyEnabled):
		httpserver.WriteError(w, r, http.StatusConflict, "twofactor_already_enabled", "Two-factor authentication is already enabled")
	case errors.Is(err, domain.ErrTwoFactorNotEnabled):
		httpserver.WriteError(w, r, http.StatusConflict, "twofactor_not_enabled", "Two-factor authentication is not enabled")
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Identity service is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
