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

type Handler struct {
	identity identityService
	config   Config
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

func NewHandler(identity identityService, config Config) *Handler {
	return &Handler{identity: identity, config: config}
}

func (handler *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/auth/login", handler.login)
	mux.Handle("GET /api/v1/auth/me", handler.authenticate(http.HandlerFunc(handler.me)))
	mux.HandleFunc("POST /api/v1/auth/logout", handler.logout)
}

func (handler *Handler) login(w http.ResponseWriter, r *http.Request) {
	identityHeaders(w)
	if !handler.sameOrigin(r) {
		writeError(w, r, domain.ErrForbidden, "origin_forbidden")
		return
	}
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
	if !handler.sameOrigin(r) {
		writeError(w, r, domain.ErrForbidden, "origin_forbidden")
		return
	}
	token, duplicate := sessionCookie(r)
	http.SetCookie(w, handler.expiredCookie())
	if duplicate {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if err := handler.identity.Logout(r.Context(), token); err != nil {
		writeError(w, r, err, "")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) authenticate(next http.Handler) http.Handler {
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
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	})
}

func (handler *Handler) sameOrigin(r *http.Request) bool {
	values := r.Header.Values("Origin")
	return len(values) == 1 && values[0] != "null" && handler.config.acceptsOrigin(values[0])
}

func (handler *Handler) sessionCookie(value string, expires time.Time, maximumAge int) *http.Cookie {
	return &http.Cookie{Name: SessionCookieName, Value: value, Path: "/", Expires: expires, MaxAge: maximumAge, HttpOnly: true, Secure: handler.config.secureCookie, SameSite: http.SameSiteLaxMode}
}

func (handler *Handler) expiredCookie() *http.Cookie {
	return handler.sessionCookie("", time.Unix(1, 0).UTC(), -1)
}

func sessionCookie(r *http.Request) (string, bool) {
	value := ""
	count := 0
	for _, cookie := range r.Cookies() {
		if cookie.Name == SessionCookieName {
			value = cookie.Value
			count++
		}
	}
	return value, count > 1
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
	case errors.Is(err, domain.ErrDependencyUnavailable):
		httpserver.WriteError(w, r, http.StatusServiceUnavailable, "dependency_unavailable", "Identity service is temporarily unavailable")
	default:
		httpserver.WriteError(w, r, http.StatusInternalServerError, "internal_error", "Unexpected server error")
	}
}
