package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

type inventoryCourseReader struct {
	calls int
	err   error
}

func (s *inventoryCourseReader) LessonCourse(context.Context, string) (domain.Course, error) {
	s.calls++
	return ownedCourse, s.err
}

type inventoryDrafts struct {
	draftFakeDrafts
	rows                  []ports.LessonDraft
	listCalls, trackCalls int
}

func (s *inventoryDrafts) ListLessonDrafts(context.Context, string) ([]ports.LessonDraft, error) {
	s.listCalls++
	return s.rows, s.lessonErr
}
func (s *inventoryDrafts) LessonTrack(ctx context.Context, lesson, track string) (*ports.PendingRecognition, error) {
	s.trackCalls++
	return s.draftFakeDrafts.LessonTrack(ctx, lesson, track)
}

type inventoryStub struct {
	ports.TranscriptRepository
	courseCalls, listCalls int
	err                    error
}

func (s *inventoryStub) LessonCourse(context.Context, string) (domain.Course, error) {
	s.courseCalls++
	return ownedCourse, s.err
}
func (s *inventoryStub) ListLessonTranscripts(context.Context, string) ([]ports.TranscriptInventory, error) {
	s.listCalls++
	return nil, nil
}

func TestInventoryRejectsInvalidIdentityBeforeReads(t *testing.T) {
	for _, tc := range []struct {
		name, lesson string
		actor        domain.Actor
		want         error
	}{
		{"lesson", "bad", ownerActor, domain.ErrInvalidRequest},
		{"actor", lessonID, domain.Actor{UserID: "bad", Administrator: true}, domain.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			courses, drafts := &inventoryCourseReader{}, &inventoryDrafts{}
			_, err := NewDrafting(drafts, courses, &fakeQueue{}, testClock, testLogger).ListLessonDrafts(context.Background(), tc.actor, tc.lesson)
			if !errors.Is(err, tc.want) || courses.calls != 0 || drafts.listCalls != 0 {
				t.Fatalf("draft inventory err=%v course reads=%d draft reads=%d; want %v and zero reads", err, courses.calls, drafts.listCalls, tc.want)
			}
			repo := &inventoryStub{}
			_, err = NewAuthoring(repo).List(context.Background(), tc.actor, tc.lesson)
			if !errors.Is(err, tc.want) || repo.courseCalls != 0 || repo.listCalls != 0 {
				t.Fatalf("transcript inventory err=%v course reads=%d inventory reads=%d; want %v and zero reads", err, repo.courseCalls, repo.listCalls, tc.want)
			}
		})
	}
}

func TestInventoryPropagatesIndependentReadErrors(t *testing.T) {
	sentinel := errors.New("inventory unavailable")
	repo := &inventoryStub{err: sentinel}
	if _, err := NewAuthoring(repo).List(context.Background(), ownerActor, lessonID); !errors.Is(err, sentinel) || repo.listCalls != 0 {
		t.Fatalf("course error=%v inventory reads=%d; want sentinel and zero inventory reads", err, repo.listCalls)
	}
	drafts := &inventoryDrafts{draftFakeDrafts: draftFakeDrafts{lessonErr: sentinel}}
	if _, err := NewDrafting(drafts, &inventoryCourseReader{}, &fakeQueue{}, testClock, testLogger).ListLessonDrafts(context.Background(), ownerActor, lessonID); !errors.Is(err, sentinel) || drafts.listCalls != 1 {
		t.Fatalf("draft-list error=%v calls=%d; want sentinel after one read", err, drafts.listCalls)
	}
}

func TestDraftInventoryBreaksEqualLanguageTiesByTrackID(t *testing.T) {
	const higher = "70000000-0000-4000-8000-000000000002"
	drafts := &inventoryDrafts{rows: []ports.LessonDraft{{TrackLanguage: "en", Draft: domain.TranscriptDraft{TrackID: higher}}, {TrackLanguage: "en", Draft: domain.TranscriptDraft{TrackID: trackID}}}}
	rows, err := NewDrafting(drafts, &inventoryCourseReader{}, &fakeQueue{}, testClock, testLogger).ListLessonDrafts(context.Background(), ownerActor, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{rows[0].Draft.TrackID, rows[1].Draft.TrackID}
	if !reflect.DeepEqual(ids, []string{trackID, higher}) {
		t.Fatalf("track order=%v; want [%s %s]", ids, trackID, higher)
	}
}

func TestRegenerateRejectsInvalidTrackBeforeReads(t *testing.T) {
	for _, track := range []string{"", "bad", "00000000-0000-0000-0000-000000000000"} {
		courses, drafts, queue := &inventoryCourseReader{}, &inventoryDrafts{}, &fakeQueue{}
		err := NewDrafting(drafts, courses, queue, testClock, testLogger).Regenerate(context.Background(), ownerActor, lessonID, track)
		if !errors.Is(err, domain.ErrInvalidRequest) || courses.calls != 0 || drafts.trackCalls != 0 || len(queue.jobs) != 0 {
			t.Fatalf("track=%q err=%v reads=%d/%d jobs=%v; want invalid_request and no work", track, err, courses.calls, drafts.trackCalls, queue.jobs)
		}
	}
}

func TestRegenerateIgnoresAnotherRunningTrack(t *testing.T) {
	duration := 60.0
	drafts := &inventoryDrafts{draftFakeDrafts: draftFakeDrafts{lessonAsset: &ports.PendingRecognition{LessonID: lessonID, AssetID: assetID, Generation: 1, TrackID: trackID, DurationSeconds: &duration}}, rows: []ports.LessonDraft{{Draft: domain.TranscriptDraft{TrackID: "70000000-0000-4000-8000-000000000002", State: domain.DraftRunning}}}}
	queue := &fakeQueue{}
	err := NewDrafting(drafts, &inventoryCourseReader{}, queue, testClock, testLogger).Regenerate(context.Background(), ownerActor, lessonID, trackID)
	if err != nil || len(queue.jobs) != 1 || queue.jobs[0].TrackID != trackID {
		t.Fatalf("regenerate err=%v jobs=%+v; want selected track enqueued", err, queue.jobs)
	}
}
