package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
	coursehttp "courses/internal/course/adapters/http"
	"courses/internal/course/application"
	"courses/internal/course/domain"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

type learningReadRepository struct {
	calls  int
	course domain.LearningCourse
	err    error
}

func (repository *learningReadRepository) GetLearningCourse(_ context.Context, _, _ string) (domain.LearningCourse, error) {
	repository.calls++
	return repository.course, repository.err
}

func TestLearningHTTPMatchesOpenAPIAndConcreteProgressContract(t *testing.T) {
	repository := &learningReadRepository{course: learningHTTPFixture()}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/learn/courses/learning-course", nil)
	response := httptest.NewRecorder()
	learningHTTPHandler(repository, learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"})).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("cache-control=%q", response.Header().Get("Cache-Control"))
	}
	var body struct {
		ContinueLessonID *string `json:"continueLessonId"`
		Modules          []struct {
			ID      string `json:"id"`
			Lessons []struct {
				ID         string  `json:"id"`
				VideoState *string `json:"videoState"`
				Position   float64 `json:"position"`
				Completed  bool    `json:"completed"`
				ClientSeq  int64   `json:"clientSeq"`
			} `json:"lessons"`
		} `json:"modules"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.ContinueLessonID == nil || *body.ContinueLessonID != learningHTTPLessonID || len(body.Modules) != 1 || body.Modules[0].ID != learningHTTPModuleID || len(body.Modules[0].Lessons) != 1 {
		t.Fatalf("learning body=%s", response.Body.String())
	}
	lesson := body.Modules[0].Lessons[0]
	if lesson.ID != learningHTTPLessonID || lesson.VideoState != nil || lesson.Position != 0 || lesson.Completed || lesson.ClientSeq != 0 {
		t.Fatalf("B27 lesson=%+v", lesson)
	}
	assertLearningHTTPContract(t, request, response, true)
}

func TestLearningHTTPBoundariesAndDenials(t *testing.T) {
	for _, test := range []struct {
		name         string
		path         string
		authenticate func(http.Handler) http.Handler
		course       domain.LearningCourse
		failure      error
		wantStatus   int
		wantCode     string
		wantCalls    int
	}{
		{name: "authentication required", path: "/api/v1/learn/courses/learning-course", authenticate: learningRejected(http.StatusUnauthorized), wantStatus: http.StatusUnauthorized, wantCalls: 0},
		{name: "invalid slug before repository", path: "/api/v1/learn/courses/Invalid", authenticate: learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"}), wantStatus: http.StatusBadRequest, wantCode: "invalid_slug", wantCalls: 0},
		{name: "course not found", path: "/api/v1/learn/courses/missing-course", authenticate: learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"}), failure: domain.ErrCourseNotFound, wantStatus: http.StatusNotFound, wantCode: "course_not_found", wantCalls: 1},
		{name: "R17 unavailable course", path: "/api/v1/learn/courses/learning-course", authenticate: learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"}), course: learningHTTPArchivedFixture(), wantStatus: http.StatusForbidden, wantCode: "course_forbidden", wantCalls: 1},
		{name: "dependency unavailable", path: "/api/v1/learn/courses/learning-course", authenticate: learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"}), failure: domain.ErrDependencyUnavailable, wantStatus: http.StatusServiceUnavailable, wantCode: "dependency_unavailable", wantCalls: 1},
		{name: "query rejected", path: "/api/v1/learn/courses/learning-course?lesson=1", authenticate: learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"}), wantStatus: http.StatusBadRequest, wantCode: "invalid_query", wantCalls: 0},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &learningReadRepository{course: test.course, err: test.failure}
			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			learningHTTPHandler(repository, test.authenticate).ServeHTTP(response, request)
			if response.Code != test.wantStatus || repository.calls != test.wantCalls {
				t.Fatalf("status=%d calls=%d body=%s", response.Code, repository.calls, response.Body.String())
			}
			if test.wantCode != "" {
				var envelope httpserver.ErrorResponse
				if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
					t.Fatal(err)
				}
				if envelope.Error.Code != test.wantCode {
					t.Fatalf("code=%q body=%s", envelope.Error.Code, response.Body.String())
				}
				assertLearningHTTPContract(t, request, response, test.wantCalls > 0)
			}
		})
	}
}

func TestR09LearningForeignDraftDeniedAndPublishedAllowed(t *testing.T) {
	for _, test := range []accesstest.Case{
		accessCase(t, accesstest.Stranger, "draft"),
		accessCase(t, accesstest.Stranger, "published"),
	} {
		t.Run(test.Name, func(t *testing.T) {
			repository := &learningReadRepository{course: learningHTTPAccessFixture(test)}
			request := httptest.NewRequest(http.MethodGet, "/api/v1/learn/courses/learning-course", nil)
			response := httptest.NewRecorder()
			learningHTTPHandler(repository, learningAuthenticated(learningHTTPActor(test))).ServeHTTP(response, request)

			if test.Want {
				if response.Code != http.StatusOK {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
			} else {
				if response.Code != http.StatusForbidden || !strings.Contains(response.Body.String(), `"code":"course_forbidden"`) {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
			}
			assertLearningHTTPContract(t, request, response, true)
		})
	}
}

func accessCase(t *testing.T, role accesstest.Role, status string) accesstest.Case {
	t.Helper()
	for _, test := range accesstest.TruthTable {
		if test.Role == role && test.Status == status {
			return test
		}
	}
	t.Fatalf("missing access truth-table row role=%q status=%q", role, status)
	return accesstest.Case{}
}

func learningHTTPActor(test accesstest.Case) httpserver.Actor {
	actor := httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"}
	switch test.Role {
	case accesstest.Owner:
		actor.UserID, actor.Role = learningHTTPOwnerID, "instructor"
	case accesstest.Administrator:
		actor.Role = "admin"
	}
	return actor
}

func learningHTTPAccessFixture(test accesstest.Case) domain.LearningCourse {
	fixture := learningHTTPFixtureWithLessons([]domain.LearningLesson{{
		ID: learningHTTPLessonID, Title: "Lesson", VideoState: domain.VideoAbsent,
	}})
	fixture.Status = domain.Status(test.Status)
	return fixture
}

func TestB01LearningCourseWithoutLessonsUsesEmptyArraysAndNullContinue(t *testing.T) {
	course, err := domain.NewLearningCourse(learningHTTPFixtureWithLessons(nil))
	if err != nil {
		t.Fatal(err)
	}
	repository := &learningReadRepository{course: course}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/learn/courses/learning-course", nil)
	response := httptest.NewRecorder()
	learningHTTPHandler(repository, learningAuthenticated(httpserver.Actor{UserID: learningHTTPViewerID, Role: "student"})).ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"lessons":[]`) || !strings.Contains(response.Body.String(), `"continueLessonId":null`) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestR21AnonymousPublicCourseProgramStillExposesNoLessonIDs(t *testing.T) {
	response := httptest.NewRecorder()
	handler(nil).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/courses/go-foundations", nil))
	if response.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Program []struct {
			Lessons []map[string]json.RawMessage `json:"lessons"`
		} `json:"program"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	for _, module := range body.Program {
		for _, lesson := range module.Lessons {
			if _, leaked := lesson["id"]; leaked || len(lesson) != 1 || lesson["title"] == nil {
				t.Fatalf("public lesson leaked fields: %s", response.Body.String())
			}
		}
	}
}

func learningHTTPHandler(repository *learningReadRepository, authenticate func(http.Handler) http.Handler) http.Handler {
	adapter := coursehttp.NewLearningHandler(application.NewLearning(repository))
	return httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) {
		adapter.Register(mux, authenticate)
	}})
}

func learningAuthenticated(actor httpserver.Actor) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), actor)))
		})
	}
}

func learningRejected(status int) func(http.Handler) http.Handler {
	return func(http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) })
	}
}

func assertLearningHTTPContract(t *testing.T, request *http.Request, response *httptest.ResponseRecorder, validRequest bool) {
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
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{
		RequestValidationInput: input,
		Status:                 response.Code, Header: response.Header(),
		Body: io.NopCloser(bytes.NewReader(response.Body.Bytes())),
	}); err != nil {
		t.Fatal(err)
	}
}

const (
	learningHTTPViewerID = "40000000-0000-4000-8000-000000000001"
	learningHTTPOwnerID  = "40000000-0000-4000-8000-000000000002"
	learningHTTPModuleID = "60000000-0000-4000-8000-000000000001"
	learningHTTPLessonID = "70000000-0000-4000-8000-000000000001"
)

func learningHTTPFixture() domain.LearningCourse {
	course, err := domain.NewLearningCourse(learningHTTPFixtureWithLessons([]domain.LearningLesson{{
		ID: learningHTTPLessonID, Title: "Lesson", VideoState: domain.VideoAbsent,
	}}))
	if err != nil {
		panic(err)
	}
	return course
}

func learningHTTPArchivedFixture() domain.LearningCourse {
	course := learningHTTPFixture()
	course.Status = domain.Archived
	return course
}

func learningHTTPFixtureWithLessons(lessons []domain.LearningLesson) domain.LearningCourse {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	return domain.LearningCourse{
		Course: domain.Course{
			ID: "50000000-0000-4000-8000-000000000001", Slug: "learning-course",
			Title: "Learning course", Summary: "Learning summary", Description: "Learning description",
			Status: domain.Published, CreatedAt: now, UpdatedAt: now, PublishedAt: &now,
		},
		InstructorID: learningHTTPOwnerID,
		Modules:      []domain.LearningModule{{ID: learningHTTPModuleID, Title: "Module", Lessons: lessons}},
	}
}
