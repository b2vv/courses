package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

type fakeCourseReader struct {
	course domain.Course
	err    error
}

func (f *fakeCourseReader) LessonCourse(context.Context, string) (domain.Course, error) {
	return f.course, f.err
}

type draftFakeDrafts struct {
	lessonDraft   ports.LessonDraft
	lessonErr     error
	lessonAsset   *ports.PendingRecognition
	lessonAErr    error
	claimResult   ports.ClaimedDraft
	completeErr   error
	completeCalls int
}

func (f *draftFakeDrafts) GetLessonDraft(context.Context, string) (ports.LessonDraft, error) {
	return f.lessonDraft, f.lessonErr
}
func (f *draftFakeDrafts) PendingRecognitions(context.Context, int) ([]ports.PendingRecognition, error) {
	return nil, nil
}
func (f *draftFakeDrafts) Claim(context.Context, string, string, int32, time.Time) (ports.ClaimedDraft, error) {
	return f.claimResult, nil
}
func (f *draftFakeDrafts) SaveDraft(context.Context, domain.TranscriptDraft, string, time.Time) (string, error) {
	return "", nil
}
func (f *draftFakeDrafts) Complete(_ context.Context, _ domain.TranscriptDraft, _ string, _ time.Time) error {
	f.completeCalls++
	return f.completeErr
}
func (f *draftFakeDrafts) Recover(context.Context, time.Time, int) ([]ports.RecoveredDraft, error) {
	return nil, nil
}
func (f *draftFakeDrafts) LessonAsset(context.Context, string) (*ports.PendingRecognition, error) {
	return f.lessonAsset, f.lessonAErr
}

var ownerActor = domain.Actor{UserID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
var ownedCourse = domain.Course{OwnerID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Status: domain.Draft}

// --- GetDraft tests ---

func TestGetDraftFound(t *testing.T) {
	t.Parallel()
	drafts := &draftFakeDrafts{
		lessonDraft: ports.LessonDraft{
			Found: true,
			Draft: domain.TranscriptDraft{
				LessonID: lessonID, AssetID: assetID, Generation: 1,
				State: domain.DraftReady, Language: "uk", Segments: testSegments(2),
			},
			UpdatedAt:       claimToken,
			ReadyGeneration: 1,
		},
	}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	status, err := d.GetDraft(context.Background(), ownerActor, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Found {
		t.Fatal("draft must be found")
	}
	if status.State != domain.DraftReady {
		t.Fatalf("state=%s want=ready", status.State)
	}
	if status.Language != "uk" {
		t.Fatalf("language=%s want=uk", status.Language)
	}
	if len(status.Segments) != 2 {
		t.Fatalf("segments=%d want=2", len(status.Segments))
	}
	if status.Stale {
		t.Fatal("must not be stale when generations match")
	}
}

func TestGetDraftNotFound(t *testing.T) {
	t.Parallel()
	d := NewDrafting(&draftFakeDrafts{}, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	status, err := d.GetDraft(context.Background(), ownerActor, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	if status.Found {
		t.Fatal("must not be found")
	}
}

func TestGetDraftForbidden(t *testing.T) {
	t.Parallel()
	otherActor := domain.Actor{UserID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"}
	d := NewDrafting(&draftFakeDrafts{}, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	_, err := d.GetDraft(context.Background(), otherActor, lessonID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err=%v want=forbidden", err)
	}
}

func TestGetDraftStale(t *testing.T) {
	t.Parallel()
	drafts := &draftFakeDrafts{
		lessonDraft: ports.LessonDraft{
			Found: true,
			Draft: domain.TranscriptDraft{
				LessonID: lessonID, AssetID: assetID, Generation: 1,
				State: domain.DraftReady, Language: "uk", Segments: testSegments(1),
			},
			UpdatedAt:       claimToken,
			ReadyGeneration: 2,
		},
	}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	status, err := d.GetDraft(context.Background(), ownerActor, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Stale {
		t.Fatal("must be stale when generations differ")
	}
}

// --- Regenerate tests ---

func TestRegenerateEnqueues(t *testing.T) {
	t.Parallel()
	dur := float64(120)
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{
			LessonID: lessonID, AssetID: assetID, Generation: 1, DurationSeconds: &dur,
		},
	}
	queue := &fakeQueue{}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, queue, testClock, testLogger)

	if err := d.Regenerate(context.Background(), ownerActor, lessonID, "uk"); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs=%d want=1", len(queue.jobs))
	}
	if queue.jobs[0].Language != "uk" {
		t.Fatalf("language=%s want=uk", queue.jobs[0].Language)
	}
}

func TestRegenerateNoAsset(t *testing.T) {
	t.Parallel()
	d := NewDrafting(&draftFakeDrafts{}, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	err := d.Regenerate(context.Background(), ownerActor, lessonID, "")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err=%v want=not_found", err)
	}
}

func TestRegenerateForbidden(t *testing.T) {
	t.Parallel()
	otherActor := domain.Actor{UserID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"}
	d := NewDrafting(&draftFakeDrafts{}, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	err := d.Regenerate(context.Background(), otherActor, lessonID, "")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err=%v want=forbidden", err)
	}
}

func TestRegenerateNilDuration(t *testing.T) {
	t.Parallel()
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{
			LessonID: lessonID, AssetID: assetID, Generation: 1, DurationSeconds: nil,
		},
	}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	err := d.Regenerate(context.Background(), ownerActor, lessonID, "")
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("err=%v want=dependency_unavailable", err)
	}
}

func TestRegenerateAutoLanguage(t *testing.T) {
	t.Parallel()
	dur := float64(60)
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{
			LessonID: lessonID, AssetID: assetID, Generation: 1, DurationSeconds: &dur,
		},
	}
	queue := &fakeQueue{}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, queue, testClock, testLogger)

	if err := d.Regenerate(context.Background(), ownerActor, lessonID, ""); err != nil {
		t.Fatal(err)
	}
	if queue.jobs[0].Language != "" {
		t.Fatalf("language=%q want empty for auto-detect", queue.jobs[0].Language)
	}
}
