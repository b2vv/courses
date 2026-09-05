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

	coursehttp "courses/internal/course/adapters/http"
	"courses/internal/course/application"
	"courses/internal/course/domain"
	identityhttp "courses/internal/identity/adapters/http"
	identityapplication "courses/internal/identity/application"
	identitydomain "courses/internal/identity/domain"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const (
	adminUserID      = "40000000-0000-4000-8000-000000000003"
	instructorUserID = "40000000-0000-4000-8000-000000000002"
	otherOwnerID     = "40000000-0000-4000-8000-000000000004"
	adminCourseID    = "10000000-0000-4000-8000-000000000001"
	missingCourseID  = "10000000-0000-4000-8000-000000000099"
)

type adminRepository struct {
	listCalls, loadCalls int
	lastActor            domain.Actor
	lastQuery            domain.Query
	lastStatus           domain.Status
	ownerID              string
	failure              error
}

func (repository *adminRepository) List(_ context.Context, actor domain.Actor, query domain.Query, status domain.Status) (domain.AuthoringPage, error) {
	repository.listCalls++
	repository.lastActor = actor
	repository.lastQuery = query
	repository.lastStatus = status
	if repository.failure != nil {
		return domain.AuthoringPage{}, repository.failure
	}
	items := []domain.AuthoringCourse{adminCourseSummary(repository.owner())}
	if actor.Role == domain.ActorAdmin {
		items = append(items, adminCourseSummary(otherOwnerID))
	}
	return domain.AuthoringPage{Items: items, Limit: query.Limit, Offset: query.Offset, Total: int64(len(items))}, nil
}

func (repository *adminRepository) Load(_ context.Context, id string) (domain.CourseTree, error) {
	repository.loadCalls++
	if repository.failure != nil {
		return domain.CourseTree{}, repository.failure
	}
	if id == missingCourseID {
		return domain.CourseTree{}, domain.ErrCourseNotFound
	}
	return adminCourseTree(repository.owner()), nil
}

func (*adminRepository) Create(context.Context, domain.Actor, domain.CourseDraft, time.Time) (domain.CourseTree, error) {
	return domain.CourseTree{}, errors.New("unexpected create")
}
func (*adminRepository) Save(context.Context, domain.Actor, domain.CourseTree, time.Time, time.Time) (domain.CourseTree, error) {
	return domain.CourseTree{}, errors.New("unexpected save")
}
func (*adminRepository) Transition(context.Context, domain.Actor, string, domain.Status, domain.Status, time.Time, time.Time) (domain.CourseTree, error) {
	return domain.CourseTree{}, errors.New("unexpected transition")
}
func (*adminRepository) Delete(context.Context, domain.Actor, string, time.Time) error {
	return errors.New("unexpected delete")
}

func (repository *adminRepository) owner() string {
	if repository.ownerID != "" {
		return repository.ownerID
	}
	return instructorUserID
}

type adminIdentity struct {
	user                  identitydomain.User
	instructors           []identitydomain.User
	authenticationFailure error
	authenticateCalls     int
	instructorListCalls   int
}

func (*adminIdentity) Login(context.Context, identityapplication.LoginInput) (identityapplication.LoginResult, error) {
	return identityapplication.LoginResult{}, errors.New("unexpected login")
}
func (identity *adminIdentity) Authenticate(context.Context, string) (identitydomain.User, error) {
	identity.authenticateCalls++
	if identity.authenticationFailure != nil {
		return identitydomain.User{}, identity.authenticationFailure
	}
	return identity.user, nil
}
func (*adminIdentity) Logout(context.Context, string) error { return errors.New("unexpected logout") }
func (identity *adminIdentity) ListInstructors(context.Context) ([]identitydomain.User, error) {
	identity.instructorListCalls++
	return identity.instructors, nil
}

func TestAdminReadHTTPMatchesOpenAPI(t *testing.T) {
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

	for _, test := range []struct {
		name, path, code string
		status           int
		role             identitydomain.Role
		ownerID          string
		authenticated    bool
		authFailure      error
		wantCourseCalls  int
		wantPeopleCalls  int
		validRequest     bool
	}{
		{name: "admin lists every course", path: "/api/v1/admin/courses", status: 200, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "admin list search parameters", path: "/api/v1/admin/courses?limit=50&offset=10000&status=archived", status: 200, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "instructor lists own courses", path: "/api/v1/admin/courses", status: 200, role: identitydomain.Instructor, ownerID: instructorUserID, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "admin reads a tree", path: "/api/v1/admin/courses/" + adminCourseID, status: 200, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "owner reads a tree", path: "/api/v1/admin/courses/" + adminCourseID, status: 200, role: identitydomain.Instructor, ownerID: instructorUserID, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "P07 admin lists owner candidates", path: "/api/v1/admin/instructors", status: 200, role: identitydomain.Admin, authenticated: true, wantPeopleCalls: 1, validRequest: true},
		{name: "R01 anonymous list", path: "/api/v1/admin/courses", code: "authentication_required", status: 401, validRequest: true},
		{name: "R01 anonymous tree", path: "/api/v1/admin/courses/" + adminCourseID, code: "authentication_required", status: 401, validRequest: true},
		{name: "R01 anonymous instructors", path: "/api/v1/admin/instructors", code: "authentication_required", status: 401, validRequest: true},
		{name: "R02 student list", path: "/api/v1/admin/courses", code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true, validRequest: true},
		{name: "student role is refused before query validation", path: "/api/v1/admin/courses?status=unknown", code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true},
		{name: "R02 student tree", path: "/api/v1/admin/courses/" + adminCourseID, code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true, validRequest: true},
		{name: "R02 student instructors", path: "/api/v1/admin/instructors", code: "forbidden", status: 403, role: identitydomain.Student, authenticated: true, validRequest: true},
		{name: "R03 disabled account", path: "/api/v1/admin/courses", code: "authentication_required", status: 401, role: identitydomain.Instructor, authenticated: true, authFailure: identitydomain.ErrAuthenticationRequired, validRequest: true},
		{name: "R04 foreign instructor", path: "/api/v1/admin/courses/" + adminCourseID, code: "course_forbidden", status: 403, role: identitydomain.Instructor, ownerID: otherOwnerID, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "R05 instructor owner candidates", path: "/api/v1/admin/instructors", code: "forbidden", status: 403, role: identitydomain.Instructor, authenticated: true, validRequest: true},
		{name: "R06 unknown id", path: "/api/v1/admin/courses/" + missingCourseID, code: "course_not_found", status: 404, role: identitydomain.Admin, authenticated: true, wantCourseCalls: 1, validRequest: true},
		{name: "R07 malformed id", path: "/api/v1/admin/courses/not-a-uuid", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
		{name: "invalid list status", path: "/api/v1/admin/courses?status=unknown", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
		{name: "unknown list parameter", path: "/api/v1/admin/courses?q=course", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
		{name: "repeated list parameter", path: "/api/v1/admin/courses?limit=1&limit=2", code: "invalid_request", status: 400, role: identitydomain.Admin, authenticated: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &adminRepository{ownerID: test.ownerID}
			identity := &adminIdentity{
				user:                  identityUser(test.role),
				instructors:           []identitydomain.User{identityUser(identitydomain.Admin), identityUser(identitydomain.Instructor)},
				authenticationFailure: test.authFailure,
			}
			handler := adminReadHandler(t, repository, identity)
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			if test.authenticated {
				request.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != test.status {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			if response.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("response is cacheable: %v", response.Header())
			}
			if repository.listCalls+repository.loadCalls != test.wantCourseCalls || identity.instructorListCalls != test.wantPeopleCalls {
				t.Fatalf("calls: list=%d load=%d instructors=%d", repository.listCalls, repository.loadCalls, identity.instructorListCalls)
			}
			if test.code != "" {
				assertAdminError(t, response, test.code)
			}
			if test.name == "P07 admin lists owner candidates" {
				assertInstructorResponse(t, response)
			}
			if test.name == "instructor lists own courses" && repository.lastActor != (domain.Actor{UserID: instructorUserID, Role: domain.ActorInstructor}) {
				t.Fatalf("identity role was not mapped explicitly: %+v", repository.lastActor)
			}
			if test.name == "admin list search parameters" && (repository.lastQuery.Limit != 50 || repository.lastQuery.Offset != 10000 || repository.lastStatus != domain.Archived) {
				t.Fatalf("search parameters were not mapped: query=%+v status=%s", repository.lastQuery, repository.lastStatus)
			}

			route, parameters, err := router.FindRoute(request)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: request, PathParams: parameters, Route: route}
			if test.validRequest {
				if err := openapi3filter.ValidateRequest(context.Background(), input); err != nil {
					t.Fatal(err)
				}
			}
			if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
				RequestValidationInput: input,
				Status:                 response.Code,
				Header:                 response.Header(),
				Body:                   io.NopCloser(bytes.NewReader(response.Body.Bytes())),
			}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestAdminReadWrongMethod(t *testing.T) {
	repository := &adminRepository{}
	identity := &adminIdentity{user: identityUser(identitydomain.Admin)}
	response := httptest.NewRecorder()
	adminReadHandler(t, repository, identity).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/api/v1/admin/courses", nil))
	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertAdminError(t, response, "method_not_allowed")
	if repository.listCalls+repository.loadCalls != 0 {
		t.Fatal("wrong method reached the repository")
	}
}

func adminReadHandler(t *testing.T, repository *adminRepository, identity *adminIdentity) http.Handler {
	t.Helper()
	config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
	if err != nil {
		t.Fatal(err)
	}
	authoring := application.NewAuthoring(repository, func() time.Time { return time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC) })
	courseHandler := coursehttp.NewHandler(nil, authoring)
	identityHandler := identityhttp.NewHandler(identity, identity, config)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		identityHandler.Register(mux)
		courseHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)
	}})
}

func identityUser(role identitydomain.Role) identitydomain.User {
	id := adminUserID
	displayName := "Адміністратор Курсу"
	email := "admin@courseforge.local"
	if role == identitydomain.Instructor {
		id, displayName, email = instructorUserID, "Викладач Курсу", "instructor@courseforge.local"
	}
	if role == identitydomain.Student {
		id, displayName, email = "40000000-0000-4000-8000-000000000001", "Студентка Курсу", "student@courseforge.local"
	}
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	return identitydomain.User{ID: id, Email: email, DisplayName: displayName, Role: role, CreatedAt: now, UpdatedAt: now}
}

func adminCourseSummary(ownerID string) domain.AuthoringCourse {
	return domain.AuthoringCourse{
		ID: adminCourseID, Slug: "go-foundations", Title: "Go foundations", Status: domain.Draft,
		InstructorID: ownerID, UpdatedAt: time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC),
	}
}

func adminCourseTree(ownerID string) domain.CourseTree {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	return domain.CourseTree{
		ID: adminCourseID, Slug: "go-foundations", Title: "Go foundations", Summary: "Build a reliable API",
		Description: "Learn modules, PostgreSQL and tests.", Status: domain.Draft, InstructorID: ownerID,
		CreatedAt: now, UpdatedAt: now,
		Modules: []domain.Module{{
			ID: "20000000-0000-4000-8000-000000000001", Title: "First module", Position: 0,
			Lessons: []domain.Lesson{{ID: "30000000-0000-4000-8000-000000000001", Title: "First lesson", Description: "Lesson text", Position: 0}},
		}},
	}
}

func assertAdminError(t *testing.T, response *httptest.ResponseRecorder, code string) {
	t.Helper()
	var envelope httpserver.ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != code || envelope.Error.RequestID == "" || envelope.Error.RequestID != response.Header().Get("X-Request-ID") {
		t.Fatalf("bad envelope: %+v", envelope)
	}
}

func assertInstructorResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	var body []struct {
		ID          string `json:"id"`
		DisplayName string `json:"displayName"`
		Role        string `json:"role"`
	}
	var members []map[string]json.RawMessage
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(response.Body.Bytes(), &members); err != nil {
		t.Fatal(err)
	}
	if len(body) != 2 || len(members) != 2 {
		t.Fatalf("instructors=%s", response.Body.String())
	}
	want := []struct {
		id, displayName, role string
	}{
		{adminUserID, "Адміністратор Курсу", "admin"},
		{instructorUserID, "Викладач Курсу", "instructor"},
	}
	for index, instructor := range members {
		if len(instructor) != 3 || instructor["id"] == nil || instructor["displayName"] == nil || instructor["role"] == nil || instructor["email"] != nil {
			t.Fatalf("owner candidate leaks fields: %s", response.Body.String())
		}
		if body[index].ID != want[index].id || body[index].DisplayName != want[index].displayName || body[index].Role != want[index].role {
			t.Fatalf("owner candidates=%s", response.Body.String())
		}
	}
}
