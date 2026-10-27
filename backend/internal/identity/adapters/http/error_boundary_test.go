package http

import (
	"courses/internal/identity/domain"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestErrorBoundaries(t *testing.T) {
	writers := map[string]func(http.ResponseWriter, *http.Request, error){
		"writeError": func(w http.ResponseWriter, r *http.Request, err error) { writeError(w, r, err, "") },
	}
	assertResponse := func(t *testing.T, writer func(http.ResponseWriter, *http.Request, error), err error, status int, code string) {
		t.Helper()
		response := httptest.NewRecorder()
		writer(response, httptest.NewRequest(http.MethodGet, "/", nil), err)
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if decodeErr := json.Unmarshal(response.Body.Bytes(), &body); decodeErr != nil {
			t.Fatal(decodeErr)
		}
		if response.Code != status || body.Error.Code != code {
			t.Fatalf("got %d/%q, want %d/%q", response.Code, body.Error.Code, status, code)
		}
	}
	for name, writer := range writers {
		t.Run(name+"/unknown", func(t *testing.T) {
			for _, err := range []error{errors.New("boom database=private user=private"), fmt.Errorf("query: %w", errors.New("boom")), errors.New("invalid_request"), errors.New("course_not_found"), errors.New("origin_forbidden")} {
				assertResponse(t, writer, err, http.StatusServiceUnavailable, "dependency_unavailable")
			}
		})
	}
	cases := []struct {
		writer string
		err    error
		status int
		code   string
	}{
		{"writeError", domain.ErrInvalidUser, http.StatusBadRequest, "invalid_request"},
		{"writeError", domain.ErrInvalidCredentials, http.StatusUnauthorized, "invalid_credentials"},
		{"writeError", domain.ErrAuthenticationRequired, http.StatusUnauthorized, "authentication_required"},
		{"writeError", domain.ErrUserNotFound, http.StatusNotFound, "user_not_found"},
		{"writeError", domain.ErrCannotTargetSelf, http.StatusConflict, "cannot_target_self"},
		{"writeError", domain.ErrAdminImmutable, http.StatusConflict, "admin_immutable"},
		{"writeError", domain.ErrLastAdmin, http.StatusConflict, "last_admin"},
		{"writeError", domain.ErrEmailTaken, http.StatusConflict, "email_taken"},
		{"writeError", domain.ErrWeakPassword, http.StatusUnprocessableEntity, "weak_password"},
		{"writeError", domain.ErrInvalidEmail, http.StatusUnprocessableEntity, "invalid_email"},
		{"writeError", domain.ErrInvalidDisplayName, http.StatusUnprocessableEntity, "invalid_display_name"},
		{"writeError", domain.ErrPasswordUnchanged, http.StatusUnprocessableEntity, "password_unchanged"},
		{"writeError", domain.ErrTokenUsed, http.StatusGone, "token_used"},
		{"writeError", domain.ErrInvalidAccountToken, http.StatusBadRequest, "invalid_account_token"},
		{"writeError", domain.ErrRateLimited, http.StatusTooManyRequests, "rate_limited"},
		{"writeError", domain.ErrInvalidTOTP, http.StatusUnprocessableEntity, "invalid_totp"},
		{"writeError", domain.ErrRecoveryCodeInvalid, http.StatusUnauthorized, "recovery_code_invalid"},
		{"writeError", domain.ErrTwoFactorAlreadyEnabled, http.StatusConflict, "twofactor_already_enabled"},
		{"writeError", domain.ErrTwoFactorNotEnabled, http.StatusConflict, "twofactor_not_enabled"},
		{"writeError", domain.ErrDependencyUnavailable, http.StatusServiceUnavailable, "dependency_unavailable"},
	}
	for _, tc := range cases {
		t.Run(tc.writer+"/"+tc.code, func(t *testing.T) {
			assertResponse(t, writers[tc.writer], tc.err, tc.status, tc.code)
			assertResponse(t, writers[tc.writer], fmt.Errorf("private infrastructure context: %w", tc.err), tc.status, tc.code)
		})
	}
}

func TestStaticTransportErrorOverrides(t *testing.T) {
	for _, tc := range []struct {
		code   string
		status int
	}{{"origin_forbidden", http.StatusForbidden}, {"invalid_request", http.StatusBadRequest}, {"forbidden", http.StatusForbidden}, {"grant_forbidden", http.StatusForbidden}} {
		response := httptest.NewRecorder()
		writeError(response, httptest.NewRequest(http.MethodGet, "/", nil), errors.New("boom"), tc.code)
		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if response.Code != tc.status || body.Error.Code != tc.code {
			t.Fatalf("got %d/%q, want %d/%q", response.Code, body.Error.Code, tc.status, tc.code)
		}
	}
}
