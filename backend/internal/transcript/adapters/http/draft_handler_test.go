package http_test

import (
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

func (s *stubDraftRepository) ListLessonDrafts(context.Context, string) ([]ports.LessonDraft, error) {
	if s.draft.Draft.TrackID == "" {
		return nil, nil
	}
	return []ports.LessonDraft{s.draft}, nil
}
func (s *stubDraftRepository) PendingRecognitions(context.Context, int) ([]ports.PendingRecognition, error) {
	return nil, nil
}
func (s *stubDraftRepository) Claim(context.Context, string, string, int32, string, bool, time.Time) (ports.ClaimedDraft, error) {
	return ports.ClaimedDraft{}, nil
}
func (s *stubDraftRepository) Complete(context.Context, domain.TranscriptDraft, string, time.Time) error {
	return nil
}
func (s *stubDraftRepository) Recover(context.Context, time.Time, int) ([]ports.RecoveredDraft, error) {
	return nil, nil
}
func (s *stubDraftRepository) LessonTrack(context.Context, string, string) (*ports.PendingRecognition, error) {
	return s.lessonAsset, nil
}

type stubCourseReader struct{ course domain.Course }

func (s *stubCourseReader) LessonCourse(context.Context, string) (domain.Course, error) {
	return s.course, nil
}

func serveDraft(t *testing.T, drafts ports.DraftRepository, courses application.CourseReader, r *http.Request) *httptest.ResponseRecorder {
	t.Helper()
	drafting := application.NewDrafting(drafts, courses, nil, time.Now, nil)
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

func TestRegenerateAlreadyRunningReturnsConflict(t *testing.T) {
	dur := 60.0
	const runningAssetID = "22222222-2222-4222-8222-222222222222"
	drafts := &stubDraftRepository{draft: ports.LessonDraft{
		IsDefault: true,
		Draft:     domain.TranscriptDraft{LessonID: lessonID, AssetID: runningAssetID, Generation: 1, TrackID: "70000000-0000-4000-8000-000000000001", State: domain.DraftRunning},
	}, lessonAsset: &ports.PendingRecognition{LessonID: lessonID, AssetID: runningAssetID, Generation: 1, TrackID: "70000000-0000-4000-8000-000000000001", DurationSeconds: &dur}}
	courses := &stubCourseReader{course: domain.Course{OwnerID: ownerID, Status: domain.Draft}}
	r := httptest.NewRequest(http.MethodPost, "/api/v1/admin/lessons/"+lessonID+"/draft/regenerate", nil)
	r.Header.Set("X-Request-ID", "test-regenerate-running")
	r = r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: ownerID, Role: "instructor"}))
	response := serveDraft(t, drafts, courses, r)
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
