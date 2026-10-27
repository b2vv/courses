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
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

type repository struct{ failure error }

func (r repository) ListPublished(_ context.Context, q domain.Query) (domain.Page, error) {
	if r.failure != nil {
		return domain.Page{}, r.failure
	}
	items := []domain.Course{}
	if q.Offset == 0 && q.Search != "empty" {
		items = append(items, fixture())
	}
	return domain.Page{Items: items, Limit: q.Limit, Offset: q.Offset, Total: 1}, nil
}
func (r repository) GetPublished(_ context.Context, slug string) (domain.CourseDetail, error) {
	if r.failure != nil {
		return domain.CourseDetail{}, r.failure
	}
	if slug != "go-foundations" {
		return domain.CourseDetail{}, domain.ErrCourseNotFound
	}
	return domain.CourseDetail{
		Course: fixture(),
		Program: []domain.ProgramModule{
			{Title: "HTTP API", Lessons: []domain.ProgramLesson{{Title: "Handlers"}, {Title: "Contracts"}}},
			{Title: "PostgreSQL", Lessons: []domain.ProgramLesson{{Title: "Transactions"}}},
		},
	}, nil
}
func fixture() domain.Course {
	now := time.Date(2026, 9, 3, 10, 0, 0, 0, time.UTC)
	return domain.Course{ID: "10000000-0000-4000-8000-000000000001", Slug: "go-foundations", Title: "Go foundations", Summary: "Build a reliable API", Description: "Learn modules, PostgreSQL and tests.", Status: domain.Published, CreatedAt: now, UpdatedAt: now, PublishedAt: &now, Price: &domain.Price{Minor: 50000, Currency: "UAH"}, Paid: true}
}
func handler(failure error) http.Handler {
	adapter := coursehttp.NewHandler(application.NewCatalog(repository{failure: failure}), nil)
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: adapter.Register})
}

func TestCatalogHTTPMatchesOpenAPI(t *testing.T) {
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
	for _, tc := range []struct {
		name, path, code string
		status           int
		failure          error
		validRequest     bool
	}{
		{"default list", "/api/v1/courses", "", 200, nil, true},
		{"boundaries", "/api/v1/courses?limit=50&offset=10000&q=" + strings.Repeat("x", 100) + "&sort=title", "", 200, nil, true},
		{"empty list", "/api/v1/courses?q=empty", "", 200, nil, true},
		{"detail", "/api/v1/courses/go-foundations", "", 200, nil, true},
		{"missing detail", "/api/v1/courses/missing-course", "course_not_found", 404, nil, true},
		{"draft hidden", "/api/v1/courses/draft-course", "course_not_found", 404, nil, true},
		{"archived hidden", "/api/v1/courses/archived-course", "course_not_found", 404, nil, true},
		{"bad limit", "/api/v1/courses?limit=0", "invalid_query", 400, nil, false},
		{"repeated parameter", "/api/v1/courses?limit=1&limit=2", "invalid_query", 400, nil, false},
		{"unknown parameter", "/api/v1/courses?role=admin", "invalid_query", 400, nil, false},
		{"decimal", "/api/v1/courses?limit=1.5", "invalid_query", 400, nil, false},
		{"negative", "/api/v1/courses?offset=-1", "invalid_query", 400, nil, false},
		{"empty numeric", "/api/v1/courses?limit=", "invalid_query", 400, nil, false},
		{"huge numeric", "/api/v1/courses?offset=99999999999999999999", "invalid_query", 400, nil, false},
		{"plus prefix", "/api/v1/courses?limit=%2B1", "invalid_query", 400, nil, false},
		{"max limit", "/api/v1/courses?limit=51", "invalid_query", 400, nil, false},
		{"max offset", "/api/v1/courses?offset=10001", "invalid_query", 400, nil, false},
		{"long Unicode", "/api/v1/courses?q=" + strings.Repeat("ї", 101), "invalid_query", 400, nil, false},
		{"unknown sort", "/api/v1/courses?sort=popular", "invalid_query", 400, nil, false},
		{"empty sort", "/api/v1/courses?sort=", "invalid_query", 400, nil, false},
		{"bad encoding", "/api/v1/courses?q=%ZZ", "invalid_query", 400, nil, false},
		{"NUL input", "/api/v1/courses?q=%00", "invalid_query", 400, nil, false},
		{"invalid slug", "/api/v1/courses/Uppercase", "invalid_slug", 400, nil, false},
		{"detail query rejected", "/api/v1/courses/go-foundations?limit=1", "invalid_query", 400, nil, false},
		{"dependency", "/api/v1/courses", "dependency_unavailable", 503, domain.ErrDependencyUnavailable, true},
		{"unexpected", "/api/v1/courses", "dependency_unavailable", 503, errors.New("password=private SQL secret"), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest("GET", tc.path, nil)
			response := httptest.NewRecorder()
			handler(tc.failure).ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("status %d body=%s", response.Code, response.Body.String())
			}
			if response.Code >= 400 && (strings.Contains(response.Body.String(), "password=private") || strings.Contains(response.Body.String(), "SQL secret")) {
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
			if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: response.Code, Header: response.Header(), Body: io.NopCloser(bytes.NewReader(response.Body.Bytes()))}); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestListPreservesDefaultsAndRequestedPagination(t *testing.T) {
	for _, tc := range []struct {
		path                 string
		limit, offset, items int
	}{
		{"/api/v1/courses", 20, 0, 1},
		{"/api/v1/courses?limit=50&offset=10000", 50, 10000, 0},
		{"/api/v1/courses?q=empty", 20, 0, 0},
	} {
		response := httptest.NewRecorder()
		handler(nil).ServeHTTP(response, httptest.NewRequest("GET", tc.path, nil))
		var body struct {
			Items      []json.RawMessage `json:"items"`
			Pagination struct {
				Limit  int `json:"limit"`
				Offset int `json:"offset"`
			} `json:"pagination"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Pagination.Limit != tc.limit || body.Pagination.Offset != tc.offset || len(body.Items) != tc.items {
			t.Fatalf("%s: unexpected transport mapping: %s", tc.path, response.Body.String())
		}
	}
}

func TestPublishedCourseDetailExposesOnlyOrderedProgramTitles(t *testing.T) {
	response := httptest.NewRecorder()
	handler(nil).ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/courses/go-foundations", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status %d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Program json.RawMessage `json:"program"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	want := `[{"title":"HTTP API","lessons":[{"title":"Handlers"},{"title":"Contracts"}]},{"title":"PostgreSQL","lessons":[{"title":"Transactions"}]}]`
	if string(body.Program) != want {
		t.Fatalf("program = %s, want %s", body.Program, want)
	}
}

func TestUnpublishedCourseProgramsAreNotExposed(t *testing.T) {
	for _, slug := range []string{"draft-course", "archived-course"} {
		response := httptest.NewRecorder()
		handler(nil).ServeHTTP(response, httptest.NewRequest("GET", "/api/v1/courses/"+slug, nil))
		if response.Code != http.StatusNotFound {
			t.Fatalf("%s: status %d body=%s", slug, response.Code, response.Body.String())
		}
		if strings.Contains(response.Body.String(), "program") || strings.Contains(response.Body.String(), "Handlers") {
			t.Fatalf("%s leaked a program: %s", slug, response.Body.String())
		}
	}
}
