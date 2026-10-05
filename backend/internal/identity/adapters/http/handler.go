package http

import (
	"bytes"
	"context"
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
}

type userKey struct{}

type loginRequest struct {
	Email    string
	Password string
}

type userDTO struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type instructorDTO struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

func NewHandler(identity identityService, instructors instructorLister, administration administrationService, config Config) *Handler {
	return &Handler{identity: identity, instructors: instructors, administration: administration, config: config}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/auth/login", handler.RequireSameOrigin(http.HandlerFunc(handler.login)))
	mux.Handle("GET /api/v1/auth/me", handler.Authenticate(http.HandlerFunc(handler.me)))
	mux.Handle("POST /api/v1/auth/logout", handler.RequireSameOrigin(http.HandlerFunc(handler.logout)))
	mux.Handle("GET /api/v1/admin/instructors", handler.Authenticate(http.HandlerFunc(handler.listInstructors)))
	mux.Handle("GET /api/v1/admin/users", handler.Authenticate(http.HandlerFunc(handler.listUsers)))
	mux.Handle("POST /api/v1/admin/users/{id}/status", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.setUserStatus))))
	mux.Handle("DELETE /api/v1/admin/users/{id}", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.deleteUser))))
	mux.Handle("POST /api/v1/admin/users/{id}/password", handler.RequireSameOrigin(handler.Authenticate(http.HandlerFunc(handler.resetUserPassword))))
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
	result, err := handler.identity.Login(r.Context(), application.LoginInput{Email: input.Email, Password: input.Password, PresentedToken: presented})
	if err != nil {
		writeError(w, r, err, "")
		return
	}
	http.SetCookie(w, handler.sessionCookie(result.Token, result.Session.ExpiresAt, int(result.Session.ExpiresAt.Sub(result.Session.CreatedAt).Seconds())))
	httpserver.WriteJSON(w, http.StatusOK, mapUser(result.User))
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

func identityHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

func mapUser(user domain.User) userDTO {
	return userDTO{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Role: string(user.Role)}
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
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Identity service is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
