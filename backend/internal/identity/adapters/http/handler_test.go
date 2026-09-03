package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identityhttp "courses/internal/identity/adapters/http"
	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/platform/httpserver"
)

type identityService struct {
	loginCalls, authenticateCalls, logoutCalls int
	loginResult                                application.LoginResult
	user                                       domain.User
	err                                        error
}

func (service *identityService) Login(context.Context, application.LoginInput) (application.LoginResult, error) {
	service.loginCalls++
	return service.loginResult, service.err
}
func (service *identityService) Authenticate(context.Context, string) (domain.User, error) {
	service.authenticateCalls++
	return service.user, service.err
}
func (service *identityService) Logout(context.Context, string) error {
	service.logoutCalls++
	return service.err
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
	adapter := identityhttp.NewHandler(service, config)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: adapter.Register})
}

func requestWithOrigin(method, path, body string) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.Header.Set("Origin", "http://localhost:8080")
	request.Header.Set("Content-Type", "application/json")
	return request
}
