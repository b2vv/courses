package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	transcripthttp "courses/internal/transcript/adapters/http"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/httpserver"
)

type stubDraftRepository struct {
	draft       ports.LessonDraft
	lessonAsset *ports.PendingRecognition
}

func (s *stubDraftRepository) GetLessonDraft(context.Context, string) (ports.LessonDraft, error) {
	return s.draft, nil
}
func (s *stubDraftRepository) PendingRecognitions(context.Context, int) ([]ports.PendingRecognition, error) {
	return nil, nil
}
func (s *stubDraftRepository) Claim(context.Context, string, string, int32, bool, time.Time) (ports.ClaimedDraft, error) {
	return ports.ClaimedDraft{}, nil
}
func (s *stubDraftRepository) SaveDraft(context.Context, domain.TranscriptDraft, string, time.Time) (string, error) {
	return "", nil
}
func (s *stubDraftRepository) Complete(context.Context, domain.TranscriptDraft, string, time.Time) error {
	return nil
}
func (s *stubDraftRepository) Recover(context.Context, time.Time, int) ([]ports.RecoveredDraft, error) {
	return nil, nil
}
func (s *stubDraftRepository) LessonAsset(context.Context, string) (*ports.PendingRecognition, error) {
	return s.lessonAsset, nil
}

type stubCourseReader struct{ course domain.Course }

func (s *stubCourseReader) LessonCourse(context.Context, string) (domain.Course, error) {
	return s.course, nil
}

type stubPublisher struct {
	err   error
	token string
}

func (s *stubPublisher) SaveLessonTranscript(context.Context, ports.SaveTranscript) (string, error) {
	return s.token, s.err
}

func serveDraft(t *testing.T, drafts ports.DraftRepository, courses application.CourseReader, publisher application.TranscriptPublisher, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	drafting := application.NewDrafting(drafts, courses, publisher, nil, time.Now, nil)
	handler := transcripthttp.NewDraftHandler(drafting)
	server := httpserver.NewHandler(httpserver.Options{
		Service: "draft-test",
		Register: func(mux *http.ServeMux) {
			handler.RegisterAdmin(mux,
				func(next http.Handler) http.Handler { return next },
				func(next http.Handler) http.Handler { return next },
			)
		},
	})
	response := httptest.NewRecorder()
	server.ServeHTTP(response, r)
	return response
}

func TestC05PublishRefusesStaleToken(t *testing.T) {
	drafts := &stubDraftRepository{
		draft: ports.LessonDraft{
			Found: true,
			Draft: domain.TranscriptDraft{
				LessonID:   lessonID,
				AssetID:    "50000000-0000-4000-8000-000000000001",
				Generation: 1,
				State:      domain.DraftReady,
				Language:   "uk",
				Segments: []domain.Segment{
					{ID: "70000000-0000-4000-8000-0000000000aa", StartMs: 0, EndMs: 1000, Text: "текст"},
				},
			},
			UpdatedAt:       "2026-01-01T00:00:00Z",
			ReadyGeneration: 1,
		},
	}
	courses := &stubCourseReader{course: domain.Course{OwnerID: ownerID, Status: domain.Draft}}
	publisher := &stubPublisher{err: domain.ErrTranscriptModified}

	body, _ := json.Marshal(map[string]string{"expectedUpdatedAt": "stale-token"})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/lessons/"+lessonID+"/draft/publish", bytes.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Request-ID", "test-c05")
	r = r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: ownerID, Role: "instructor"}))

	response := serveDraft(t, drafts, courses, publisher, r)

	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d want=409 (C05: stale token must be refused): %s", response.Code, response.Body.String())
	}
	var errBody struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &errBody); err != nil {
		t.Fatal(err)
	}
	if errBody.Error.Code != "transcript_modified" {
		t.Fatalf("code=%s want=transcript_modified (C05)", errBody.Error.Code)
	}
}

func TestPublishSuccessReturnsToken(t *testing.T) {
	drafts := &stubDraftRepository{
		draft: ports.LessonDraft{
			Found: true,
			Draft: domain.TranscriptDraft{
				LessonID:   lessonID,
				AssetID:    "50000000-0000-4000-8000-000000000001",
				Generation: 1,
				State:      domain.DraftReady,
				Language:   "uk",
				Segments: []domain.Segment{
					{ID: "70000000-0000-4000-8000-0000000000aa", StartMs: 0, EndMs: 1000, Text: "текст"},
				},
			},
			UpdatedAt:       "2026-01-01T00:00:00Z",
			ReadyGeneration: 1,
		},
	}
	courses := &stubCourseReader{course: domain.Course{OwnerID: ownerID, Status: domain.Draft}}
	publisher := &stubPublisher{token: "2026-01-01T00:00:01Z"}

	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/lessons/"+lessonID+"/draft/publish", nil)
	r.Header.Set("X-Request-ID", "test-publish-ok")
	r = r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: ownerID, Role: "instructor"}))

	response := serveDraft(t, drafts, courses, publisher, r)

	if response.Code != http.StatusOK {
		t.Fatalf("status=%d want=200: %s", response.Code, response.Body.String())
	}
	var body struct {
		UpdatedAt string `json:"updatedAt"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.UpdatedAt != "2026-01-01T00:00:01Z" {
		t.Fatalf("updatedAt=%s want=2026-01-01T00:00:01Z", body.UpdatedAt)
	}
}

func TestRegenerateAlreadyRunningReturnsConflict(t *testing.T) {
	dur := 60.0
	const runningAssetID = "22222222-2222-4222-8222-222222222222"
	drafts := &stubDraftRepository{draft: ports.LessonDraft{
		Found: true,
		Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: runningAssetID, Generation: 1, State: domain.DraftRunning},
	}, lessonAsset: &ports.PendingRecognition{LessonID: lessonID, AssetID: runningAssetID, Generation: 1, DurationSeconds: &dur}}
	courses := &stubCourseReader{course: domain.Course{OwnerID: ownerID, Status: domain.Draft}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/lessons/"+lessonID+"/draft/regenerate", nil)
	r.Header.Set("X-Request-ID", "test-regenerate-running")
	r = r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: ownerID, Role: "instructor"}))
	response := serveDraft(t, drafts, courses, &stubPublisher{}, r)
	if response.Code != http.StatusConflict {
		t.Fatalf("status=%d want=409: %s", response.Code, response.Body.String())
	}
	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "already_running" {
		t.Fatalf("code=%s want already_running", body.Error.Code)
	}
}
