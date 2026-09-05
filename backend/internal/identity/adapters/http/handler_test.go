package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	identityhttp "courses/internal/identity/adapters/http"
	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

type identityService struct {
	loginCalls, authenticateCalls, logoutCalls, listInstructorCalls int
	logoutTokens                                                    []string
	loginResult                                                     application.LoginResult
	user                                                            domain.User
	instructors                                                     []domain.User
	err                                                             error
}

func (service *identityService) Login(context.Context, application.LoginInput) (application.LoginResult, error) {
	service.loginCalls++
	return service.loginResult, service.err
}
func (service *identityService) Authenticate(context.Context, string) (domain.User, error) {
	service.authenticateCalls++
	return service.user, service.err
}
func (service *identityService) Logout(_ context.Context, token string) error {
	service.logoutCalls++
	service.logoutTokens = append(service.logoutTokens, token)
	return service.err
}
func (service *identityService) ListInstructors(context.Context) ([]domain.User, error) {
	service.listInstructorCalls++
	return service.instructors, service.err
}

func TestLoginHTTPValidationAndCookie(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	user := domain.User{ID: "40000000-0000-4000-8000-000000000001", Email: "student@courseforge.local", DisplayName: "Студентка", Role: domain.Student, CreatedAt: now, UpdatedAt: now}
	service := &identityService{loginResult: application.LoginResult{User: user, Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: user.ID, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: strings.Repeat("A", 43)}}
	handler := testHandler(t, service)

	request := requestWithOrigin("POST", "/api/v1/auth/login", `{"email":"student@courseforge.local","password":"demo"}`)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.loginCalls != 1 || !strings.Contains(response.Body.String(), `"role":"student"`) {
		t.Fatalf("status=%d body=%s calls=%d", response.Code, response.Body.String(), service.loginCalls)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != identityhttp.SessionCookieName || !cookies[0].HttpOnly || cookies[0].Secure || cookies[0].SameSite != http.SameSiteLaxMode || cookies[0].Path != "/" || cookies[0].MaxAge != 12*60*60 {
		t.Fatalf("cookie=%+v", cookies)
	}
	if response.Header().Get("Cache-Control") != "no-store" || response.Header().Get("X-Request-ID") == "" {
		t.Fatalf("headers=%v", response.Header())
	}

	for _, body := range []string{`{`, `{}`, `{"email":1,"password":"x"}`, `{"email":"a@b","email":"c@d","password":"x"}`, `{"email":"a@b","password":"x","role":"admin"}`} {
		before := service.loginCalls
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, requestWithOrigin("POST", "/api/v1/auth/login", body))
		if response.Code != http.StatusBadRequest || service.loginCalls != before {
			t.Fatalf("body=%s status=%d calls=%d", body, response.Code, service.loginCalls)
		}
	}
}

func TestOriginAuthenticationAndLogoutMatrix(t *testing.T) {
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	user := domain.User{ID: "40000000-0000-4000-8000-000000000001", Email: "student@courseforge.local", DisplayName: "Студентка", Role: domain.Student, CreatedAt: now, UpdatedAt: now}
	token := strings.Repeat("A", 43)
	service := &identityService{user: user}
	handler := testHandler(t, service)

	for _, origin := range []string{"", "null", "https://evil.example", "http://localhost:8081"} {
		request := httptest.NewRequest("POST", "/api/v1/auth/logout", nil)
		if origin != "" {
			request.Header.Set("Origin", origin)
		}
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusForbidden || service.logoutCalls != 0 {
			t.Fatalf("origin=%q status=%d calls=%d", origin, response.Code, service.logoutCalls)
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/auth/me", nil))
	if response.Code != http.StatusUnauthorized || service.authenticateCalls != 0 {
		t.Fatalf("missing cookie status=%d calls=%d", response.Code, service.authenticateCalls)
	}
	request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	request.Header.Set("Cookie", identityhttp.SessionCookieName+"="+token)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK || service.authenticateCalls != 1 || !strings.Contains(response.Body.String(), "student@courseforge.local") {
		t.Fatalf("me status=%d body=%s calls=%d", response.Code, response.Body.String(), service.authenticateCalls)
	}

	request = requestWithOrigin("POST", "/api/v1/auth/logout", "")
	request.Header.Set("Cookie", identityhttp.SessionCookieName+"="+token)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusNoContent || service.logoutCalls != 1 {
		t.Fatalf("logout status=%d calls=%d", response.Code, service.logoutCalls)
	}
	cookie := response.Result().Cookies()[0]
	if cookie.MaxAge != -1 || !cookie.Expires.Before(time.Now()) || cookie.Path != "/" {
		t.Fatalf("expired cookie=%+v", cookie)
	}

	service.err = domain.ErrDependencyUnavailable
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || response.Result().Cookies()[0].MaxAge != -1 {
		t.Fatalf("failed logout status=%d cookie=%+v", response.Code, response.Result().Cookies())
	}
}

func TestDuplicateSessionCookieFailsClosed(t *testing.T) {
	service := &identityService{}
	handler := testHandler(t, service)
	request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	request.Header.Set("Cookie", identityhttp.SessionCookieName+"="+strings.Repeat("A", 43)+"; "+identityhttp.SessionCookieName+"="+strings.Repeat("B", 43))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized || service.authenticateCalls != 0 {
		t.Fatalf("status=%d calls=%d", response.Code, service.authenticateCalls)
	}
}

func TestIdentityHTTPExplicitErrorMatrix(t *testing.T) {
	token := strings.Repeat("A", 43)
	service := &identityService{err: domain.ErrInvalidCredentials}
	handler := testHandler(t, service)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, requestWithOrigin("POST", "/api/v1/auth/login", `{"email":"invalid","password":"wrong"}`))
	if response.Code != http.StatusUnauthorized || !strings.Contains(response.Body.String(), `"code":"invalid_credentials"`) {
		t.Fatalf("invalid credentials status=%d body=%s", response.Code, response.Body.String())
	}

	service.err = domain.ErrDependencyUnavailable
	request := httptest.NewRequest("GET", "/api/v1/auth/me", nil)
	request.Header.Set("Cookie", identityhttp.SessionCookieName+"="+token)
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusServiceUnavailable || !strings.Contains(response.Body.String(), `"code":"dependency_unavailable"`) {
		t.Fatalf("dependency status=%d body=%s", response.Code, response.Body.String())
	}

	before := service.loginCalls
	request = requestWithOrigin("POST", "/api/v1/auth/login", `{"email":"a@b","password":"x"}`)
	request.Header.Add("Origin", "http://localhost:8080")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden || service.loginCalls != before {
		t.Fatalf("duplicate origin status=%d calls=%d", response.Code, service.loginCalls)
	}

	response = httptest.NewRecorder()
	handler.ServeHTTP(response, requestWithOrigin("POST", "/api/v1/auth/login", `{"email":"a@b","password":"`+strings.Repeat("x", 9000)+`"}`))
	if response.Code != http.StatusBadRequest {
		t.Fatalf("oversized body status=%d", response.Code)
	}
}

func testHandler(t *testing.T, service *identityService) http.Handler {
	t.Helper()
	config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
	if err != nil {
		t.Fatal(err)
	}
	adapter := identityhttp.NewHandler(service, service, config)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: adapter.Register})
}

func requestWithOrigin(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Origin", "http://localhost:8080")
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestIdentityHTTPMatchesOpenAPI(t *testing.T) {
	document, err := openapi3.NewLoader().LoadFromFile(filepath.Join("..", "..", "..", "..", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(context.Background()); err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(document)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	user := domain.User{ID: "40000000-0000-4000-8000-000000000001", Email: "student@courseforge.local", DisplayName: "Студентка", Role: domain.Student, CreatedAt: now, UpdatedAt: now}
	success := application.LoginResult{User: user, Session: domain.Session{ID: "50000000-0000-4000-8000-000000000001", UserID: user.ID, CreatedAt: now, ExpiresAt: now.Add(12 * time.Hour)}, Token: strings.Repeat("A", 43)}
	for _, tc := range []struct {
		name, method, path, body, code string
		status                         int
		origin, cookie                 bool
		failure                        error
		validRequest                   bool
	}{
		{name: "login succeeds", method: "POST", path: "/api/v1/auth/login", body: `{"email":"student@courseforge.local","password":"demo"}`, status: 200, origin: true, validRequest: true},
		{name: "login without origin", method: "POST", path: "/api/v1/auth/login", body: `{"email":"student@courseforge.local","password":"demo"}`, code: "origin_forbidden", status: 403, validRequest: true},
		{name: "login with malformed body", method: "POST", path: "/api/v1/auth/login", body: `{"email":"a@b","password":"x","role":"admin"}`, code: "invalid_request", status: 400, origin: true},
		{name: "login with wrong credentials", method: "POST", path: "/api/v1/auth/login", body: `{"email":"student@courseforge.local","password":"wrong"}`, code: "invalid_credentials", status: 401, origin: true, failure: domain.ErrInvalidCredentials, validRequest: true},
		{name: "login while identity is unavailable", method: "POST", path: "/api/v1/auth/login", body: `{"email":"student@courseforge.local","password":"demo"}`, code: "dependency_unavailable", status: 503, origin: true, failure: domain.ErrDependencyUnavailable, validRequest: true},
		{name: "login with an unexpected failure", method: "POST", path: "/api/v1/auth/login", body: `{"email":"student@courseforge.local","password":"demo"}`, code: "internal_error", status: 500, origin: true, failure: errors.New("password=private argon secret"), validRequest: true},
		{name: "me with a session", method: "GET", path: "/api/v1/auth/me", status: 200, cookie: true, validRequest: true},
		{name: "me without a session", method: "GET", path: "/api/v1/auth/me", code: "authentication_required", status: 401, validRequest: true},
		{name: "me with a revoked session", method: "GET", path: "/api/v1/auth/me", code: "authentication_required", status: 401, cookie: true, failure: domain.ErrAuthenticationRequired, validRequest: true},
		{name: "logout revokes", method: "POST", path: "/api/v1/auth/logout", status: 204, origin: true, cookie: true, validRequest: true},
		{name: "logout without a session", method: "POST", path: "/api/v1/auth/logout", status: 204, origin: true, validRequest: true},
		{name: "logout without origin", method: "POST", path: "/api/v1/auth/logout", code: "origin_forbidden", status: 403, cookie: true, validRequest: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := &identityService{loginResult: success, user: user, err: tc.failure}
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}
			if tc.origin {
				request.Header.Set("Origin", "http://localhost:8080")
			}
			if tc.cookie {
				request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
			}
			route, parameters, err := router.FindRoute(request)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route}
			if tc.validRequest {
				if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
					t.Fatal(err)
				}
			}
			// Request validation drains the body, so the handler gets a fresh reader.
			request.Body = io.NopCloser(strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			testHandler(t, service).ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d body=%s", response.Code, response.Body.String())
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("identity response is cacheable: %v", response.Header())
			}
			if strings.Contains(response.Body.String(), "password=private") || strings.Contains(response.Body.String(), "argon secret") {
				t.Fatal("error details leaked")
			}
			if tc.code != "" {
				var envelope httpserver.ErrorResponse
				if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.Error.Code != tc.code || envelope.Error.RequestID != response.Header().Get("X-Request-ID") {
					t.Fatalf("bad envelope: %+v", envelope)
				}
			}
			if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: response.Code, Header: response.Header(), Body: io.NopCloser(bytes.NewReader(response.Body.Bytes()))}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// A logout that answers 204 while a presented session stays alive on the server
// is a lie the browser cannot detect. Login and authentication refuse a
// duplicated cookie untouched; logout cannot, because refusing would leave the
// session live, so it revokes every token the request carried.
func TestLogoutRevokesEverySessionCookieTheRequestCarried(t *testing.T) {
	service := &identityService{}
	request := requestWithOrigin("POST", "/api/v1/auth/logout", "")
	request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
	request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("B", 43)})
	response := httptest.NewRecorder()
	testHandler(t, service).ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if len(service.logoutTokens) != 2 ||
		service.logoutTokens[0] != strings.Repeat("A", 43) ||
		service.logoutTokens[1] != strings.Repeat("B", 43) {
		t.Fatalf("a presented session survived logout: %v", service.logoutTokens)
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Value != "" || cookies[0].MaxAge >= 0 {
		t.Fatalf("browser cookie not cleared: %+v", cookies)
	}
}

func TestLogoutReportsAFailureToRevokeAnyPresentedSession(t *testing.T) {
	service := &identityService{err: domain.ErrDependencyUnavailable}
	request := requestWithOrigin("POST", "/api/v1/auth/logout", "")
	request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
	request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("B", 43)})
	response := httptest.NewRecorder()
	testHandler(t, service).ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("a failed revocation claimed success: status=%d", response.Code)
	}
}
