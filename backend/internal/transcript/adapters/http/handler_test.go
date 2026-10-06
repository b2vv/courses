package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"courses/internal/contracts/access/accesstest"
	transcripthttp "courses/internal/transcript/adapters/http"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const (
	lessonID   = "70000000-0000-4000-8000-000000000001"
	viewerID   = "40000000-0000-4000-8000-000000000001"
	ownerID    = "40000000-0000-4000-8000-000000000002"
	documentID = "60000000-0000-4000-8000-000000000001"
)

type stubRepository struct {
	result ports.LessonTranscript
	err    error

	// gotLanguage records what the handler passed down, so the query parameter is proven to
	// reach the read rather than being accepted and dropped.
	gotLanguage domain.Language

	saved   *ports.SaveTranscript
	saveErr error
}

func (stub *stubRepository) GetLessonTranscript(_ context.Context, _, _ string, language domain.Language) (ports.LessonTranscript, error) {
	stub.gotLanguage = language
	return stub.result, stub.err
}

func (stub *stubRepository) CourseOfID(context.Context, string) (domain.Course, error) {
	return stub.result.Course, stub.err
}

func (stub *stubRepository) CourseDocuments(context.Context, string) ([]domain.DocumentLink, error) {
	return nil, stub.err
}

func (stub *stubRepository) AddCourseDocument(context.Context, string, domain.DocumentLink) error {
	return stub.saveErr
}

func (stub *stubRepository) RemoveCourseDocument(context.Context, string, string) error {
	return stub.saveErr
}

func (stub *stubRepository) LessonCourse(context.Context, string) (domain.Course, error) {
	return stub.result.Course, stub.err
}

func (stub *stubRepository) SaveLessonTranscript(_ context.Context, save ports.SaveTranscript) (string, error) {
	stub.saved = &save
	if stub.saveErr != nil {
		return "", stub.saveErr
	}
	return "2026-09-18T00:00:00Z", nil
}

func published(segments []domain.Segment, documents []domain.DocumentLink, generation, ready int32) ports.LessonTranscript {
	transcript, err := domain.NewTranscript(domain.Transcript{
		LessonID: lessonID, Language: "uk", Segments: segments, Documents: documents,
		AuthoredAgainstGeneration: generation,
	})
	if err != nil {
		panic(err)
	}
	return ports.LessonTranscript{
		Found: true, ReadyGeneration: ready, UpdatedAt: "2026-09-18T00:00:00Z",
		Course:     domain.Course{OwnerID: ownerID, Status: domain.Published},
		Transcript: transcript,
	}
}

func activeMember(result ports.LessonTranscript) ports.LessonTranscript {
	result.Course.Membership = "active"
	return result
}

func request(t *testing.T, target string, actor httpserver.Actor) *http.Request {
	t.Helper()
	if actor.UserID != "" {
		actor.EmailVerified = true
	}
	r := httptest.NewRequest(http.MethodGet, target, nil)
	if actor.UserID != "" {
		r = r.WithContext(httpserver.WithActor(r.Context(), actor))
	}
	return r
}

// serve runs the request through the real transport stack rather than a bare mux: the contract
// requires X-Request-ID on every response, and that header is the transport's job. A test that
// skipped it would assert against a document the running server does not satisfy.
func serve(t *testing.T, repository ports.TranscriptRepository, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	handler := transcripthttp.NewHandler(application.NewReading(repository))
	server := httpserver.NewHandler(httpserver.Options{
		Service: "transcript-test",
		Register: func(mux *http.ServeMux) {
			handler.Register(mux, func(next http.Handler) http.Handler { return next })
		},
	})
	response := httptest.NewRecorder()
	server.ServeHTTP(response, r)
	return response
}

func TestTranscriptReadMatchesOpenAPI(t *testing.T) {
	segment := domain.Segment{
		ID: "70000000-0000-4000-8000-0000000000aa", StartMs: 0, EndMs: 1000,
		Text: "перша", Speaker: "А", DocumentIDs: []string{documentID},
	}
	link := domain.DocumentLink{ID: documentID, Label: "Doc", URL: "https://example.test/a"}
	repository := &stubRepository{result: activeMember(published([]domain.Segment{segment}, []domain.DocumentLink{link}, 1, 1))}

	r := request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript", httpserver.Actor{UserID: viewerID, Role: "student"})
	response := serve(t, repository, r)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200: %s", response.Code, response.Body.String())
	}
	assertContract(t, r, response, true)

	var body struct {
		LessonID  string  `json:"lessonId"`
		Language  *string `json:"language"`
		UpdatedAt *string `json:"updatedAt"`
		Stale     bool    `json:"stale"`
		Segments  []struct {
			DocumentIDs []string `json:"documentIds"`
		} `json:"segments"`
		Documents []struct {
			ID string `json:"id"`
		} `json:"documents"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Language == nil || *body.Language != "uk" {
		t.Fatalf("language = %v; want uk", body.Language)
	}
	// Without the token on the read, an author could never obtain what the write demands: the
	// editing loop would not close, and every save would look like a first write.
	if body.UpdatedAt == nil || *body.UpdatedAt == "" {
		t.Fatal("the read does not carry the optimistic-lock token")
	}
	// The segment references; the library travels once. This is the shape that keeps the
	// worst case at ~13 MiB instead of ~75 MiB, so it is asserted, not assumed.
	if len(body.Segments) != 1 || len(body.Segments[0].DocumentIDs) != 1 || body.Segments[0].DocumentIDs[0] != documentID {
		t.Fatalf("segment does not reference the document by id: %+v", body.Segments)
	}
	if len(body.Documents) != 1 || body.Documents[0].ID != documentID {
		t.Fatalf("library entry missing from the response: %+v", body.Documents)
	}
	if response.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control = %q; want no-store: a transcript is private", response.Header().Get("Cache-Control"))
	}
}

func TestAbsentTranscriptIsAStateNotAMissingResource(t *testing.T) {
	repository := &stubRepository{result: ports.LessonTranscript{
		Found: false, Course: domain.Course{OwnerID: ownerID, Status: domain.Published, Membership: "active"},
	}}
	r := request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript", httpserver.Actor{UserID: viewerID, Role: "student"})
	response := serve(t, repository, r)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200 — a lesson without a transcript is a state", response.Code)
	}
	assertContract(t, r, response, true)
}

func TestTheReadTakesNoPagingParameters(t *testing.T) {
	// D-4: the lesson is served whole so that client-side search cannot lie about a miss. An
	// unknown parameter is refused rather than ignored, so a caller who thinks it is paging
	// finds out instead of silently receiving everything.
	for _, target := range []string{
		"/api/v1/learn/lessons/" + lessonID + "/transcript?limit=10",
		"/api/v1/learn/lessons/" + lessonID + "/transcript?offset=10",
		"/api/v1/learn/lessons/" + lessonID + "/transcript?page=2",
	} {
		repository := &stubRepository{result: activeMember(published(nil, nil, 0, 0))}
		response := serve(t, repository, request(t, target, httpserver.Actor{UserID: viewerID, Role: "student"}))
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d; want 400", target, response.Code)
		}
	}
}

func TestLanguageReachesTheReadAndIsNormalised(t *testing.T) {
	repository := &stubRepository{result: activeMember(published(nil, nil, 0, 0))}
	serve(t, repository, request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript?language=UK", httpserver.Actor{UserID: viewerID, Role: "student"}))
	if repository.gotLanguage != "uk" {
		t.Fatalf("language reached the read as %q; want uk", repository.gotLanguage)
	}
}

func TestRefusalDoesNotDependOnWhetherATranscriptExists(t *testing.T) {
	draft := domain.Course{OwnerID: ownerID, Status: domain.Draft}
	for _, test := range []struct {
		name   string
		result ports.LessonTranscript
	}{
		{name: "draft course with a transcript", result: ports.LessonTranscript{Found: true, Course: draft, Transcript: published(nil, nil, 0, 0).Transcript}},
		{name: "draft course without a transcript", result: ports.LessonTranscript{Found: false, Course: draft}},
	} {
		t.Run(test.name, func(t *testing.T) {
			r := request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript", httpserver.Actor{UserID: viewerID, Role: "student"})
			response := serve(t, &stubRepository{result: test.result}, r)
			if response.Code != http.StatusForbidden {
				t.Fatalf("status = %d; want 403 in both cases — otherwise absence tells the caller the course exists", response.Code)
			}
			assertContract(t, r, response, true)
		})
	}
}

func TestPay19PaidTranscriptRequiresPurchase(t *testing.T) {
	result := published(nil, nil, 0, 0)
	result.Course.Paid = true
	r := request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript", httpserver.Actor{UserID: viewerID, Role: "student"})
	response := serve(t, &stubRepository{result: result}, r)
	if response.Code != http.StatusForbidden || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"course_purchase_required"`)) {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertContract(t, r, response, true)
}

func TestR09TranscriptForeignDraftDeniedAndActiveMemberAllowed(t *testing.T) {
	for _, test := range []accesstest.Case{
		transcriptAccessCase(t, accesstest.Stranger, "draft"),
		transcriptAccessCase(t, accesstest.Stranger, "published"),
	} {
		t.Run(test.Name, func(t *testing.T) {
			result := published(nil, nil, 0, 0)
			result.Course.Status = domain.Status(test.Status)
			if test.Want {
				result.Course.Membership = "active"
			}
			r := request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript", transcriptAccessActor(test))
			response := serve(t, &stubRepository{result: result}, r)

			if test.Want {
				if response.Code != http.StatusOK {
					t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
				}
			} else if response.Code != http.StatusForbidden {
				t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
			}
			assertContract(t, r, response, true)
		})
	}
}

func transcriptAccessCase(t *testing.T, role accesstest.Role, status string) accesstest.Case {
	t.Helper()
	for _, test := range accesstest.TruthTable {
		if test.Role == role && test.Status == status {
			return test
		}
	}
	t.Fatalf("missing access truth-table row role=%q status=%q", role, status)
	return accesstest.Case{}
}

func transcriptAccessActor(test accesstest.Case) httpserver.Actor {
	actor := httpserver.Actor{UserID: viewerID, Role: "student"}
	switch test.Role {
	case accesstest.Owner:
		actor.UserID, actor.Role = ownerID, "instructor"
	case accesstest.Administrator:
		actor.Role = "admin"
	}
	return actor
}

func TestUnauthenticatedIsRefusedBeforeAnythingElse(t *testing.T) {
	r := request(t, "/api/v1/learn/lessons/"+lessonID+"/transcript", httpserver.Actor{})
	response := serve(t, &stubRepository{result: published(nil, nil, 0, 0)}, r)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d; want 401", response.Code)
	}
	assertContract(t, r, response, true)
}

func assertContract(t *testing.T, request *http.Request, response *httptest.ResponseRecorder, validRequest bool) {
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
