package http_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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
				if tc.running {
					var body httpserver.ErrorResponse
					if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
						t.Fatal(err)
					}
					if body.Error.Message != "Recognition is already running for this track" {
						t.Fatalf("running message=%q; want Recognition is already running for this track", body.Error.Message)
					}
				}
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

func TestDraftInventoryPreservesRecognitionFields(t *testing.T) {
	confidence := 0.8
	drafts := &stubDraftRepository{drafts: []ports.LessonDraft{
		{IsDefault: true, TrackLanguage: "uk", Draft: domain.TranscriptDraft{TrackID: draftTrackID, TrackLanguage: "uk", State: domain.DraftReady, AssetID: draftAssetID, Language: "en", Segments: []domain.Segment{{ID: documentID, StartMs: 123, EndMs: 1456, Text: "recognized", Confidence: &confidence, Speaker: "Lecturer"}}}},
		{TrackLanguage: "uk", Draft: domain.TranscriptDraft{TrackID: "71000000-0000-4000-8000-000000000002", TrackLanguage: "uk", State: domain.DraftFailed, AssetID: draftAssetID, Failure: "engine_error"}},
		{TrackLanguage: "uk", Draft: domain.TranscriptDraft{TrackID: "71000000-0000-4000-8000-000000000003", TrackLanguage: "uk", State: domain.DraftRunning, AssetID: draftAssetID}},
		{TrackLanguage: "uk", Draft: domain.TranscriptDraft{TrackID: "71000000-0000-4000-8000-000000000004", TrackLanguage: "uk", State: domain.DraftNoSpeech, AssetID: draftAssetID}},
	}}
	rec := httptest.NewRecorder()
	draftServer(t, drafts, &stubCourseReader{course: domain.Course{OwnerID: ownerID}}, &draftQueue{}).ServeHTTP(rec, draftRequest("GET", "/drafts", "", ownerID, ""))
	if rec.Code != 200 {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Drafts []struct {
			State, AssetID string
			Language       *string
			Failure        *string
			Segments       []struct {
				ID             string
				StartMs, EndMs int64
				Confidence     *float64
				Speaker        *string
			}
		}
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Drafts) != 4 {
		t.Fatalf("draft count=%d", len(body.Drafts))
	}
	for i, want := range []string{"ready", "failed", "running", "no_speech"} {
		row := body.Drafts[i]
		if row.State != want {
			t.Fatalf("draft %d state=%q; want %q", i, row.State, want)
		}
		if row.AssetID != "72000000-0000-4000-8000-000000000001" {
			t.Fatalf("draft %d assetId=%q; want 72000000-0000-4000-8000-000000000001", i, row.AssetID)
		}
		if row.Segments == nil || (i > 0 && len(row.Segments) != 0) {
			t.Fatalf("draft %d segments=%+v; want an array, empty for non-ready states", i, row.Segments)
		}
	}
	ready, failed := body.Drafts[0], body.Drafts[1]
	if ready.Language == nil || *ready.Language != "en" || failed.Language != nil {
		t.Fatalf("recognized languages ready=%v failed=%v; want en and null", ready.Language, failed.Language)
	}
	if failed.Failure == nil || *failed.Failure != "engine_error" {
		t.Fatalf("failure=%v; want engine_error", failed.Failure)
	}
	if len(ready.Segments) != 1 {
		t.Fatalf("segments=%d", len(ready.Segments))
	}
	segment := ready.Segments[0]
	if segment.ID != documentID {
		t.Fatalf("segment id=%q; want %s", segment.ID, documentID)
	}
	if segment.StartMs != 123 || segment.EndMs != 1456 {
		t.Fatalf("segment times=%d..%d; want 123..1456", segment.StartMs, segment.EndMs)
	}
	if segment.Confidence == nil || *segment.Confidence != 0.8 {
		t.Fatalf("confidence=%v; want 0.8", segment.Confidence)
	}
	if segment.Speaker == nil || *segment.Speaker != "Lecturer" {
		t.Fatalf("speaker=%v; want Lecturer", segment.Speaker)
	}
}

type failingDraftBody struct{}

func (failingDraftBody) Read([]byte) (int, error) { return 0, errors.New("body read failed") }
func (failingDraftBody) Close() error             { return nil }

var _ io.ReadCloser = failingDraftBody{}

func TestRegenerateRefusesBodyReadFailure(t *testing.T) {
	duration := 60.0
	drafts := &stubDraftRepository{lessonAsset: &ports.PendingRecognition{TrackID: draftTrackID, DurationSeconds: &duration}}
	queue := &draftQueue{}
	request := draftRequest("POST", "/drafts/"+draftTrackID+"/regenerate", "", ownerID, "http://localhost:8080")
	request.Body = failingDraftBody{}
	rec := httptest.NewRecorder()
	draftServer(t, drafts, &stubCourseReader{course: domain.Course{OwnerID: ownerID}}, queue).ServeHTTP(rec, request)
	assertM2R10Error(t, rec, 400, "invalid_request")
	if len(queue.jobs) != 0 || drafts.gotTrackID != "" {
		t.Fatalf("body failure did work: lookup=%q jobs=%v", drafts.gotTrackID, queue.jobs)
	}
}
