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

const (
	adminID   = "40000000-0000-4000-8000-000000000003"
	studentID = "40000000-0000-4000-8000-000000000001"
	deletedID = "40000000-0000-4000-8000-000000000004"
	missingID = "40000000-0000-4000-8000-000000000099"
)

type administrationService struct {
	list                 func(application.ListUsersInput) (application.UserPage, error)
	err                  error
	listCalls            int
	setStatusCalls       int
	deleteCalls          int
	resetPasswordCalls   int
	lastList             application.ListUsersInput
	lastActor            domain.Actor
	lastID, lastPassword string
	lastStatus           domain.AccountStatus
}

func (service *administrationService) ListUsers(_ context.Context, input application.ListUsersInput) (application.UserPage, error) {
	service.listCalls++
	service.lastList = input
	if service.list != nil {
		return service.list(input)
	}
	return application.UserPage{}, service.err
}

func (service *administrationService) SetStatus(_ context.Context, actor domain.Actor, id string, status domain.AccountStatus) error {
	service.setStatusCalls++
	service.lastActor, service.lastID, service.lastStatus = actor, id, status
	return service.err
}

func (service *administrationService) Delete(_ context.Context, actor domain.Actor, id string) error {
	service.deleteCalls++
	service.lastActor, service.lastID = actor, id
	return service.err
}

func (service *administrationService) ResetPassword(_ context.Context, id, password string) error {
	service.resetPasswordCalls++
	service.lastID, service.lastPassword = id, password
	return service.err
}

func (service *administrationService) calls() int {
	return service.listCalls + service.setStatusCalls + service.deleteCalls + service.resetPasswordCalls
}

func TestAdminUsersListContract(t *testing.T) {
	now := time.Date(2026, 9, 6, 12, 0, 0, 123456000, time.UTC)
	disabledAt := now.Add(-time.Hour)
	deletedAt := now
	users := []domain.User{
		{ID: studentID, Email: "active@courseforge.local", DisplayName: "Active user", Role: domain.Student, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now},
		{ID: "40000000-0000-4000-8000-000000000002", Email: "blocked@courseforge.local", DisplayName: "Blocked user", Role: domain.Instructor, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now, DisabledAt: &disabledAt},
		{ID: deletedID, Email: "deleted@courseforge.local", DisplayName: "Deleted user", Role: domain.Student, CreatedAt: now.Add(-2 * time.Hour), UpdatedAt: now, DisabledAt: &disabledAt, DeletedAt: &deletedAt},
	}
	service := &administrationService{list: func(input application.ListUsersInput) (application.UserPage, error) {
		if input.Status == domain.AccountDeleted {
			return application.UserPage{Items: users[2:], Limit: input.Limit, Offset: input.Offset, Total: 1}, nil
		}
		return application.UserPage{Items: users[:2], Limit: input.Limit, Offset: input.Offset, Total: 2}, nil
	}}
	identity := &identityService{user: adminUser(domain.Admin)}
	handler := adminUsersHandler(t, identity, service)

	response := serveAdminUsersContract(t, handler, adminUsersRequest(http.MethodGet, "/api/v1/admin/users?limit=50&offset=10&q=%25_!", "", false, true), true)
	if response.Code != http.StatusOK || service.lastList != (application.ListUsersInput{Limit: 50, Offset: 10, Search: "%_!"}) {
		t.Fatalf("status=%d input=%+v body=%s", response.Code, service.lastList, response.Body.String())
	}
	var page struct {
		Items []struct {
			ID, Email, DisplayName, Role, Status, UpdatedAt string
		} `json:"items"`
		Pagination struct {
			Limit, Offset int
			Total         int64
		} `json:"pagination"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 2 || page.Items[0].ID != studentID || page.Items[0].Email != "active@courseforge.local" || page.Items[0].DisplayName != "Active user" || page.Items[0].Role != "student" || page.Items[0].Status != "active" || page.Items[1].Status != "blocked" || page.Items[0].UpdatedAt != "2026-09-06T12:00:00.123456Z" || page.Pagination.Limit != 50 || page.Pagination.Offset != 10 || page.Pagination.Total != 2 {
		t.Fatalf("page=%+v", page)
	}

	response = serveAdminUsersContract(t, handler, adminUsersRequest(http.MethodGet, "/api/v1/admin/users?status=deleted", "", false, true), true)
	if response.Code != http.StatusOK || service.lastList.Status != domain.AccountDeleted || !strings.Contains(response.Body.String(), `"status":"deleted"`) || !strings.Contains(response.Body.String(), `"email":"deleted@courseforge.local"`) {
		t.Fatalf("status=%d input=%+v body=%s", response.Code, service.lastList, response.Body.String())
	}

	service.list = func(input application.ListUsersInput) (application.UserPage, error) {
		return application.UserPage{Items: nil, Limit: input.Limit, Offset: input.Offset, Total: 0}, nil
	}
	response = serveAdminUsersContract(t, handler, adminUsersRequest(http.MethodGet, "/api/v1/admin/users?status=deleted", "", false, true), true)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"items":[]`) || !strings.Contains(response.Body.String(), `"total":0`) {
		t.Fatalf("empty page status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestAdminUserMutationsContract(t *testing.T) {
	service := &administrationService{}
	handler := adminUsersHandler(t, &identityService{user: adminUser(domain.Admin)}, service)
	for _, test := range []struct {
		name, method, path, body string
		status                   int
	}{
		{name: "block", method: http.MethodPost, path: "/api/v1/admin/users/" + studentID + "/status", body: `{"status":"blocked"}`, status: 200},
		{name: "restore", method: http.MethodPost, path: "/api/v1/admin/users/" + studentID + "/status", body: `{"status":"active"}`, status: 200},
		{name: "delete", method: http.MethodDelete, path: "/api/v1/admin/users/" + studentID, status: 204},
		{name: "reset blocked target", method: http.MethodPost, path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"new password"}`, status: 204},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := serveAdminUsersContract(t, handler, adminUsersRequest(test.method, test.path, test.body, true, true), true)
			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
		})
	}
	if service.setStatusCalls != 2 || service.deleteCalls != 1 || service.resetPasswordCalls != 1 || service.lastActor.UserID != adminID || service.lastID != studentID || service.lastPassword != "new password" {
		t.Fatalf("service=%+v", service)
	}
}

func TestAdminPasswordBoundsAndStrictBodies(t *testing.T) {
	tests := []struct {
		name, path, body string
		status           int
		wantCalls        int
		validRequest     bool
	}{
		{name: "empty password", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":""}`, status: 400},
		{name: "password over 1024 bytes", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"` + strings.Repeat("x", 1025) + `"}`, status: 400},
		{name: "password exactly 1024 bytes", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"` + strings.Repeat("x", 1024) + `"}`, status: 204, wantCalls: 1, validRequest: true},
		{name: "multibyte password exactly 1024 bytes", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"` + strings.Repeat("ї", 512) + `"}`, status: 204, wantCalls: 1, validRequest: true},
		{name: "multibyte password over 1024 bytes", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"` + strings.Repeat("ї", 513) + `"}`, status: 400},
		{name: "whitespace-only password", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"   "}`, status: 204, wantCalls: 1, validRequest: true},
		{name: "same as current password", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"existing password"}`, status: 204, wantCalls: 1, validRequest: true},
		{name: "unknown member", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"x","role":"admin"}`, status: 400},
		{name: "duplicate member", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"x","password":"y"}`, status: 400},
		{name: "malformed body", path: "/api/v1/admin/users/" + studentID + "/password", body: `{`, status: 400},
		{name: "trailing JSON", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"x"} {}`, status: 400},
		{name: "invalid UTF-8", path: "/api/v1/admin/users/" + studentID + "/password", body: "{\"password\":\"" + string([]byte{0xff}) + "\"}", status: 400},
		{name: "oversized body", path: "/api/v1/admin/users/" + studentID + "/password", body: `{"password":"` + strings.Repeat("x", 9000) + `"}`, status: 400},
		{name: "illegal status", path: "/api/v1/admin/users/" + studentID + "/status", body: `{"status":"deleted"}`, status: 400},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := &administrationService{}
			handler := adminUsersHandler(t, &identityService{user: adminUser(domain.Admin)}, service)
			response := serveAdminUsersContract(t, handler, adminUsersRequest(http.MethodPost, test.path, test.body, true, true), test.validRequest)
			if response.Code != test.status || service.calls() != test.wantCalls {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, service.calls(), response.Body.String())
			}
			if response.Code == http.StatusBadRequest {
				assertAdminUsersError(t, response, "invalid_request")
			}
		})
	}
}

func TestAdminUsersRefusalMatrix(t *testing.T) {
	operations := []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/admin/users", ""},
		{http.MethodPost, "/api/v1/admin/users/" + studentID + "/status", `{"status":"blocked"}`},
		{http.MethodDelete, "/api/v1/admin/users/" + studentID, ""},
		{http.MethodPost, "/api/v1/admin/users/" + studentID + "/password", `{"password":"x"}`},
	}
	for _, caller := range []struct {
		name   string
		role   domain.Role
		code   string
		status int
		cookie bool
	}{
		{name: "anonymous", code: "authentication_required", status: 401},
		{name: "student", role: domain.Student, code: "forbidden", status: 403, cookie: true},
		{name: "instructor", role: domain.Instructor, code: "forbidden", status: 403, cookie: true},
	} {
		for _, operation := range operations {
			t.Run(caller.name+" "+operation.method+" "+operation.path, func(t *testing.T) {
				service := &administrationService{}
				identity := &identityService{user: adminUser(caller.role)}
				response := serveAdminUsersContract(t, adminUsersHandler(t, identity, service), adminUsersRequest(operation.method, operation.path, operation.body, operation.method != http.MethodGet, caller.cookie), true)
				if response.Code != caller.status || service.calls() != 0 {
					t.Fatalf("status=%d calls=%d body=%s", response.Code, service.calls(), response.Body.String())
				}
				assertAdminUsersError(t, response, caller.code)
			})
		}
	}

	service := &administrationService{}
	disabledIdentity := &identityService{err: domain.ErrAuthenticationRequired}
	response := serveAdminUsersContract(t, adminUsersHandler(t, disabledIdentity, service), adminUsersRequest(http.MethodGet, "/api/v1/admin/users", "", false, true), true)
	if response.Code != http.StatusUnauthorized || service.calls() != 0 {
		t.Fatalf("disabled admin status=%d calls=%d", response.Code, service.calls())
	}
	assertAdminUsersError(t, response, "authentication_required")

	for _, operation := range operations[1:] {
		service = &administrationService{}
		identity := &identityService{user: adminUser(domain.Admin)}
		response = serveAdminUsersContract(t, adminUsersHandler(t, identity, service), adminUsersRequest(operation.method, operation.path, operation.body, false, true), true)
		if response.Code != http.StatusForbidden || identity.authenticateCalls != 0 || service.calls() != 0 {
			t.Fatalf("foreign origin status=%d auth=%d calls=%d", response.Code, identity.authenticateCalls, service.calls())
		}
		assertAdminUsersError(t, response, "origin_forbidden")
	}
}

func TestAdminUsersBusinessErrorMapping(t *testing.T) {
	for _, test := range []struct {
		name, method, path, body, code string
		failure                        error
		status                         int
		wantCalls                      int
	}{
		{name: "cannot block self", method: http.MethodPost, path: "/api/v1/admin/users/" + adminID + "/status", body: `{"status":"blocked"}`, failure: domain.ErrCannotTargetSelf, status: 409, code: "cannot_target_self", wantCalls: 1},
		{name: "cannot delete self", method: http.MethodDelete, path: "/api/v1/admin/users/" + adminID, failure: domain.ErrCannotTargetSelf, status: 409, code: "cannot_target_self", wantCalls: 1},
		{name: "admin immutable", method: http.MethodDelete, path: "/api/v1/admin/users/40000000-0000-4000-8000-000000000005", failure: domain.ErrAdminImmutable, status: 409, code: "admin_immutable", wantCalls: 1},
		{name: "last admin", method: http.MethodPost, path: "/api/v1/admin/users/40000000-0000-4000-8000-000000000005/status", body: `{"status":"blocked"}`, failure: domain.ErrLastAdmin, status: 409, code: "last_admin", wantCalls: 1},
		{name: "unknown status target", method: http.MethodPost, path: "/api/v1/admin/users/" + missingID + "/status", body: `{"status":"blocked"}`, failure: domain.ErrUserNotFound, status: 404, code: "user_not_found", wantCalls: 1},
		{name: "unknown delete target", method: http.MethodDelete, path: "/api/v1/admin/users/" + missingID, failure: domain.ErrUserNotFound, status: 404, code: "user_not_found", wantCalls: 1},
		{name: "unknown reset target", method: http.MethodPost, path: "/api/v1/admin/users/" + missingID + "/password", body: `{"password":"x"}`, failure: domain.ErrUserNotFound, status: 404, code: "user_not_found", wantCalls: 1},
		{name: "deleted reset target", method: http.MethodPost, path: "/api/v1/admin/users/" + deletedID + "/password", body: `{"password":"x"}`, failure: domain.ErrUserNotFound, status: 404, code: "user_not_found", wantCalls: 1},
		{name: "dependency unavailable", method: http.MethodGet, path: "/api/v1/admin/users", failure: domain.ErrDependencyUnavailable, status: 503, code: "dependency_unavailable", wantCalls: 1},
		{name: "unexpected failure", method: http.MethodGet, path: "/api/v1/admin/users", failure: errors.New("private database detail"), status: 500, code: "internal_error", wantCalls: 1},
		{name: "malformed id", method: http.MethodDelete, path: "/api/v1/admin/users/not-a-uuid", status: 400, code: "invalid_request"},
	} {
		t.Run(test.name, func(t *testing.T) {
			service := &administrationService{err: test.failure}
			response := serveAdminUsersContract(t, adminUsersHandler(t, &identityService{user: adminUser(domain.Admin)}, service), adminUsersRequest(test.method, test.path, test.body, test.method != http.MethodGet, true), test.name != "malformed id")
			if response.Code != test.status || service.calls() != test.wantCalls || strings.Contains(response.Body.String(), "private database detail") {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, service.calls(), response.Body.String())
			}
			assertAdminUsersError(t, response, test.code)
		})
	}
}

func TestAdminUsersListRejectsInvalidQueryBeforeUseCase(t *testing.T) {
	for _, query := range []string{
		"?unknown=x", "?limit=0", "?limit=51", "?limit=1&limit=2", "?offset=10001",
		"?status=unknown", "?q=" + strings.Repeat("x", 101), "?q=%00",
	} {
		service := &administrationService{}
		response := serveAdminUsersContract(t, adminUsersHandler(t, &identityService{user: adminUser(domain.Admin)}, service), adminUsersRequest(http.MethodGet, "/api/v1/admin/users"+query, "", false, true), false)
		if response.Code != http.StatusBadRequest || service.calls() != 0 {
			t.Fatalf("query=%q status=%d calls=%d body=%s", query, response.Code, service.calls(), response.Body.String())
		}
		assertAdminUsersError(t, response, "invalid_request")
	}
}

func adminUsersHandler(t *testing.T, identity *identityService, administration *administrationService) http.Handler {
	t.Helper()
	config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
	if err != nil {
		t.Fatal(err)
	}
	adapter := identityhttp.NewHandler(identity, identity, administration, config)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: adapter.Register})
}

func adminUser(role domain.Role) domain.User {
	user := domain.User{ID: adminID, Email: "admin@courseforge.local", DisplayName: "Admin", Role: role}
	if role == domain.Student {
		user.ID, user.Email, user.DisplayName = studentID, "student@courseforge.local", "Student"
	}
	if role == domain.Instructor {
		user.ID, user.Email, user.DisplayName = "40000000-0000-4000-8000-000000000002", "instructor@courseforge.local", "Instructor"
	}
	return user
}

func adminUsersRequest(method, path, body string, validOrigin, cookie bool) *http.Request {
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if validOrigin {
		request.Header.Set("Origin", "http://localhost:8080")
	} else if method != http.MethodGet {
		request.Header.Set("Origin", "https://evil.example")
	}
	if cookie {
		request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
	}
	return request
}

func serveAdminUsersContract(t *testing.T, handler http.Handler, request *http.Request, validRequest bool) *httptest.ResponseRecorder {
	t.Helper()
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
	body, err := io.ReadAll(request.Body)
	if err != nil {
		t.Fatal(err)
	}
	request.Body = io.NopCloser(bytes.NewReader(body))
	route, parameters, err := router.FindRoute(request)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route}
	if validRequest {
		if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
			t.Fatal(err)
		}
	}
	request.Body = io.NopCloser(bytes.NewReader(body))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("admin identity response is cacheable: %v", response.Header())
	}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input,
		Status:                 response.Code,
		Header:                 response.Header(),
		Body:                   io.NopCloser(bytes.NewReader(response.Body.Bytes())),
	}); err != nil {
		t.Fatal(err)
	}
	return response
}

func assertAdminUsersError(t *testing.T, response *httptest.ResponseRecorder, code string) {
	t.Helper()
	var envelope httpserver.ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != code || envelope.Error.RequestID != response.Header().Get("X-Request-ID") {
		t.Fatalf("error=%+v", envelope)
	}
}
