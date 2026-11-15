package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identityhttp "courses/internal/identity/adapters/http"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	transcripthttp "courses/internal/transcript/adapters/http"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/httpserver"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

const draftTrackID = "71000000-0000-4000-8000-000000000001"
const draftAssetID = "72000000-0000-4000-8000-000000000001"

type stubDraftRepository struct {
	drafts      []ports.LessonDraft
	lessonAsset *ports.PendingRecognition
	gotTrackID  string
}

func (s *stubDraftRepository) ListLessonDrafts(context.Context, string) ([]ports.LessonDraft, error) {
	return s.drafts, nil
}
func (*stubDraftRepository) PendingRecognitions(context.Context, int) ([]ports.PendingRecognition, error) {
	return nil, nil
}
func (*stubDraftRepository) Claim(context.Context, string, string, int32, string, bool, time.Time) (ports.ClaimedDraft, error) {
	return ports.ClaimedDraft{}, nil
}
func (*stubDraftRepository) Complete(context.Context, domain.TranscriptDraft, string, time.Time) error {
	return nil
}
func (*stubDraftRepository) Recover(context.Context, time.Time, int) ([]ports.RecoveredDraft, error) {
	return nil, nil
}
func (s *stubDraftRepository) LessonTrack(_ context.Context, _, trackID string) (*ports.PendingRecognition, error) {
	s.gotTrackID = trackID
	return s.lessonAsset, nil
}

type stubCourseReader struct {
	course domain.Course
	err    error
}

func (s *stubCourseReader) LessonCourse(context.Context, string) (domain.Course, error) {
	return s.course, s.err
}

type draftQueue struct{ jobs []ports.RecognitionJob }

func (q *draftQueue) Enqueue(_ context.Context, job ports.RecognitionJob, _ time.Duration) error {
	q.jobs = append(q.jobs, job)
	return nil
}

func draftServer(t *testing.T, drafts *stubDraftRepository, courses *stubCourseReader, queue *draftQueue) http.Handler {
	t.Helper()
	config, err := identityhttp.NewConfig("test", "http://localhost:8080", false)
	if err != nil {
		t.Fatal(err)
	}
	identity := identityhttp.NewHandler(nil, nil, nil, config)
	handler := transcripthttp.NewDraftHandler(application.NewDrafting(drafts, courses, queue, time.Now, nil))
	return httpserver.NewHandler(httpserver.Options{Service: "draft-test", Register: func(mux *http.ServeMux) {
		handler.RegisterAdmin(mux, func(next http.Handler) http.Handler { return next }, identity.RequireSameOrigin)
	}})
}
func draftRequest(method, suffix, body, user, origin string) *http.Request {
	r := httptest.NewRequest(method, "/api/v1/admin/lessons/"+lessonID+suffix, strings.NewReader(body))
	r.ContentLength = -1 // exercise streaming-body rejection, not only ContentLength
	if user != "" {
		r = r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: user, Role: "instructor"}))
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}
func assertM2R10Error(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status=%d want=%d body=%s", rec.Code, status, rec.Body.String())
	}
	var body httpserver.ErrorResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != code {
		t.Fatalf("code=%s want=%s", body.Error.Code, code)
	}
}
func validateMS2Response(t *testing.T, r *http.Request, rec *httptest.ResponseRecorder) {
	t.Helper()
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	route, params, err := router.FindRoute(r)
	if err != nil {
		t.Fatal(err)
	}
	input := &openapi3filter.RequestValidationInput{Request: r, PathParams: params, Route: route}
	if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: rec.Code, Header: rec.Header(), Body: rec.Result().Body}); err != nil {
		t.Fatal(err)
	}
}

func TestM2R10RegenerateHTTP(t *testing.T) {
	for _, tc := range []struct {
		name, user, origin, body             string
		missingTrack, missingLesson, running bool
		status                               int
		code                                 string
	}{
		{name: "accepted", user: ownerID, origin: "http://localhost:8080", status: 202},
		{name: "missing track", user: ownerID, origin: "http://localhost:8080", missingTrack: true, status: 404, code: "track_not_found"},
		{name: "missing lesson", user: ownerID, origin: "http://localhost:8080", missingLesson: true, status: 404, code: "lesson_not_found"},
		{name: "running", user: ownerID, origin: "http://localhost:8080", running: true, status: 409, code: "already_running"},
		{name: "foreign course", user: viewerID, origin: "http://localhost:8080", status: 403, code: "course_forbidden"},
		{name: "anonymous", origin: "http://localhost:8080", status: 401, code: "authentication_required"},
		{name: "foreign origin", user: ownerID, origin: "https://foreign.example", status: 403, code: "origin_forbidden"},
		{name: "missing origin", user: ownerID, status: 403, code: "origin_forbidden"},
		{name: "object body", user: ownerID, origin: "http://localhost:8080", body: "{}", status: 400, code: "invalid_request"},
		{name: "legacy language body", user: ownerID, origin: "http://localhost:8080", body: `{"language":"uk"}`, status: 400, code: "invalid_request"},
		{name: "whitespace body", user: ownerID, origin: "http://localhost:8080", body: " ", status: 400, code: "invalid_request"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			duration := 60.0
			drafts := &stubDraftRepository{lessonAsset: &ports.PendingRecognition{LessonID: lessonID, AssetID: draftAssetID, Generation: 2, TrackID: draftTrackID, DurationSeconds: &duration}}
			courses := &stubCourseReader{course: domain.Course{OwnerID: ownerID}}
			if tc.missingLesson {
				courses.err = domain.ErrNotFound
			}
			if tc.missingTrack {
				drafts.lessonAsset = nil
			}
			if tc.running {
				drafts.drafts = []ports.LessonDraft{{Draft: domain.TranscriptDraft{TrackID: draftTrackID, State: domain.DraftRunning}}}
			}
			queue := &draftQueue{}
			r := draftRequest("POST", "/drafts/"+draftTrackID+"/regenerate", tc.body, tc.user, tc.origin)
			rec := httptest.NewRecorder()
			draftServer(t, drafts, courses, queue).ServeHTTP(rec, r)
			if tc.code != "" {
				assertM2R10Error(t, rec, tc.status, tc.code)
			} else if rec.Code != tc.status {
				t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
			}
			if tc.status == 202 {
				if drafts.gotTrackID != draftTrackID || len(queue.jobs) != 1 || queue.jobs[0].TrackID != draftTrackID || !queue.jobs[0].Requested {
					t.Fatalf("lookup=%q jobs=%+v", drafts.gotTrackID, queue.jobs)
				}
			} else if len(queue.jobs) != 0 {
				t.Fatal("refusal enqueued work")
			}
			validateMS2Response(t, r, rec)
		})
	}
}

func TestM2R10DraftsListHTTP(t *testing.T) {
	makeDraft := func(id, language string, def, embedded bool) ports.LessonDraft {
		return ports.LessonDraft{IsDefault: def, IsEmbedded: embedded, TrackLanguage: domain.Language(language), UpdatedAt: "2026-11-15T10:00:00Z", Draft: domain.TranscriptDraft{TrackID: id, TrackLanguage: domain.Language(language), State: domain.DraftReady, Language: domain.Language(language), Segments: []domain.Segment{{ID: documentID, StartMs: 0, EndMs: 1000, Text: "recognized", Speaker: "Lecturer"}}, AssetID: draftAssetID, Generation: 2}}
	}
	drafts := &stubDraftRepository{drafts: []ports.LessonDraft{
		makeDraft("71000000-0000-4000-8000-000000000005", "uk", false, false),
		makeDraft("71000000-0000-4000-8000-000000000004", "en", false, false),
		makeDraft("71000000-0000-4000-8000-000000000003", "en", false, false),
		makeDraft("71000000-0000-4000-8000-000000000002", "pl", false, true),
		makeDraft(draftTrackID, "zh", true, false),
	}}
	drafts.drafts[0].TrackMultilingual = true      // mode changed since claim
	drafts.drafts[1].TrackLanguage = "de"          // language changed; snapshot must remain en
	drafts.drafts[2].Draft.TrackLanguage = "und-1" // unknown-family snapshot remains fresh
	courses := &stubCourseReader{course: domain.Course{OwnerID: ownerID}}
	r := draftRequest("GET", "/drafts", "", ownerID, "")
	rec := httptest.NewRecorder()
	draftServer(t, drafts, courses, &draftQueue{}).ServeHTTP(rec, r)
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response=%+v %s", rec, rec.Body.String())
	}
	var body struct {
		Drafts []struct {
			TrackID, TrackLanguage   string
			TrackMultilingual, Stale bool
			Segments                 []json.RawMessage
		}
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	// Current de sorts before en; default and embedded retain priority.
	expected := []string{draftTrackID, "71000000-0000-4000-8000-000000000002", "71000000-0000-4000-8000-000000000004", "71000000-0000-4000-8000-000000000003", "71000000-0000-4000-8000-000000000005"}
	if len(body.Drafts) != len(expected) {
		t.Fatalf("drafts=%s", rec.Body.String())
	}
	for i, draft := range body.Drafts {
		if draft.TrackID != expected[i] || draft.Segments == nil {
			t.Fatalf("order/segments=%s", rec.Body.String())
		}
	}
	if !body.Drafts[2].Stale || body.Drafts[2].TrackLanguage != "en" || body.Drafts[3].Stale || !body.Drafts[4].Stale || body.Drafts[4].TrackMultilingual {
		t.Fatalf("snapshots/stale=%s", rec.Body.String())
	}
	validateMS2Response(t, r, rec)
	for _, tc := range []struct {
		name, user string
		missing    bool
		status     int
		code       string
	}{
		{"empty", ownerID, false, 200, ""}, {"foreign", viewerID, false, 403, "course_forbidden"}, {"anonymous", "", false, 401, "authentication_required"}, {"missing", ownerID, true, 404, "lesson_not_found"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &stubDraftRepository{}
			course := &stubCourseReader{course: domain.Course{OwnerID: ownerID}}
			if tc.missing {
				course.err = domain.ErrNotFound
			}
			req := draftRequest("GET", "/drafts", "", tc.user, "")
			result := httptest.NewRecorder()
			draftServer(t, repo, course, &draftQueue{}).ServeHTTP(result, req)
			if tc.code != "" {
				assertM2R10Error(t, result, tc.status, tc.code)
			} else if result.Code != 200 || result.Body.String() != "{\"drafts\":[]}\n" {
				t.Fatalf("empty=%d %s", result.Code, result.Body.String())
			}
			validateMS2Response(t, req, result)
		})
	}
}
func TestM2R10RemovedRoutesReturn404(t *testing.T) {
	server := draftServer(t, &stubDraftRepository{}, &stubCourseReader{}, &draftQueue{})
	for _, tc := range []struct{ method, path string }{{"GET", "/draft"}, {"POST", "/draft/regenerate"}, {"POST", "/draft/publish"}} {
		r := draftRequest(tc.method, tc.path, "", ownerID, "http://localhost:8080")
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, r)
		if rec.Code != 404 {
			t.Fatalf("%s %s=%d", tc.method, tc.path, rec.Code)
		}
	}
}
