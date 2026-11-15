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

func (f *draftFakeDrafts) ListLessonDrafts(context.Context, string) ([]ports.LessonDraft, error) {
	if f.lessonDraft.Draft.TrackID == "" {
		return nil, f.lessonErr
	}
	return []ports.LessonDraft{f.lessonDraft}, f.lessonErr
}
func (f *draftFakeDrafts) PendingRecognitions(context.Context, int) ([]ports.PendingRecognition, error) {
	return nil, nil
}
func (f *draftFakeDrafts) Claim(context.Context, string, string, int32, string, bool, time.Time) (ports.ClaimedDraft, error) {
	return f.claimResult, nil
}
func (f *draftFakeDrafts) Complete(_ context.Context, _ domain.TranscriptDraft, _ string, _ time.Time) error {
	f.completeCalls++
	return f.completeErr
}
func (f *draftFakeDrafts) Recover(context.Context, time.Time, int) ([]ports.RecoveredDraft, error) {
	return nil, nil
}
func (f *draftFakeDrafts) LessonTrack(context.Context, string, string) (*ports.PendingRecognition, error) {
	return f.lessonAsset, f.lessonAErr
}

var ownerActor = domain.Actor{UserID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"}
var ownedCourse = domain.Course{OwnerID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Status: domain.Draft}

// --- GetDraft tests ---

func TestGetDraftFound(t *testing.T) {
	t.Parallel()
	drafts := &draftFakeDrafts{
		lessonDraft: ports.LessonDraft{
			IsDefault: true,
			Draft: domain.TranscriptDraft{
				LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, TrackLanguage: "uk",
				State: domain.DraftReady, Language: "uk", Segments: testSegments(2),
			},
			UpdatedAt:     claimToken,
			TrackLanguage: "uk",
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
		t.Fatal("must not be stale when track metadata matches")
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
			IsDefault: true,
			Draft: domain.TranscriptDraft{
				LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, TrackLanguage: "uk",
				State: domain.DraftReady, Language: "uk", Segments: testSegments(1),
			},
			UpdatedAt:     claimToken,
			TrackLanguage: "en",
		},
	}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	status, err := d.GetDraft(context.Background(), ownerActor, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	if !status.Stale {
		t.Fatal("must be stale when track languages differ")
	}
}

// --- Regenerate tests ---

func TestRegenerateEnqueues(t *testing.T) {
	t.Parallel()
	dur := float64(120)
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{
			LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, DurationSeconds: &dur,
		},
	}
	queue := &fakeQueue{}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, queue, testClock, testLogger)

	if err := d.Regenerate(context.Background(), ownerActor, lessonID); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs=%d want=1", len(queue.jobs))
	}
	// Multi-audio S6: the author's Regenerate may re-recognise a draft that heard the default.
	if !queue.jobs[0].Requested {
		t.Fatal("a Regenerate job is not marked requested")
	}
}

func TestRegenerateRunningCurrentDraftIsVisibleAndDoesNotEnqueue(t *testing.T) {
	t.Parallel()
	dur := float64(120)
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, DurationSeconds: &dur},
		lessonDraft: ports.LessonDraft{IsDefault: true, UpdatedAt: claimToken, Draft: domain.TranscriptDraft{
			LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, State: domain.DraftRunning,
		}},
	}
	queue := &fakeQueue{}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, queue, testClock, testLogger)

	if err := d.Regenerate(context.Background(), ownerActor, lessonID); !errors.Is(err, domain.ErrAlreadyRunning) {
		t.Fatalf("err=%v want already_running", err)
	}
	if len(queue.jobs) != 0 {
		t.Fatalf("jobs=%d want=0", len(queue.jobs))
	}
}

func TestRegenerateNoAsset(t *testing.T) {
	t.Parallel()
	d := NewDrafting(&draftFakeDrafts{}, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	err := d.Regenerate(context.Background(), ownerActor, lessonID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("err=%v want=not_found", err)
	}
}

func TestRegenerateForbidden(t *testing.T) {
	t.Parallel()
	otherActor := domain.Actor{UserID: "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"}
	d := NewDrafting(&draftFakeDrafts{}, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	err := d.Regenerate(context.Background(), otherActor, lessonID)
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("err=%v want=forbidden", err)
	}
}

func TestRegenerateNilDuration(t *testing.T) {
	t.Parallel()
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{
			LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, DurationSeconds: nil,
		},
	}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, &fakeQueue{}, testClock, testLogger)

	err := d.Regenerate(context.Background(), ownerActor, lessonID)
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("err=%v want=dependency_unavailable", err)
	}
}

func TestRegenerateTargetsDefaultTrack(t *testing.T) {
	t.Parallel()
	dur := float64(60)
	drafts := &draftFakeDrafts{
		lessonAsset: &ports.PendingRecognition{
			LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, DurationSeconds: &dur,
		},
	}
	queue := &fakeQueue{}
	d := NewDrafting(drafts, &fakeCourseReader{course: ownedCourse}, queue, testClock, testLogger)

	if err := d.Regenerate(context.Background(), ownerActor, lessonID); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].TrackID != trackID {
		t.Fatalf("jobs=%+v want default track %s", queue.jobs, trackID)
	}
}
