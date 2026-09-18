//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

func segment(id, text string, startMs, endMs int64, documentIDs ...string) domain.Segment {
	return domain.Segment{
		ID: id, StartMs: startMs, EndMs: endMs, Text: text, DocumentIDs: documentIDs,
	}
}

func runWriteTests(t *testing.T, ctx context.Context, pool *pgxpool.Pool, authoring *application.Authoring, reading *application.Reading) {
	owner := domain.Actor{UserID: ownerAccount}
	stranger := domain.Actor{UserID: strangerAccount}

	t.Run("a stranger cannot write, including the very first transcript", func(t *testing.T) {
		// Authorization reads the course directly, so writing the first transcript of a lesson is
		// refused exactly as writing the tenth would be. If it went through the transcript, an
		// empty lesson would have no course to refuse against.
		_, err := authoring.Save(ctx, stranger, application.SaveRequest{
			LessonID: bareLesson, Language: "uk",
			Segments: []domain.Segment{segment(newSegmentID(1), "чуже", 0, 100)},
		})
		if err != domain.ErrForbidden {
			t.Fatalf("stranger writing err = %v; want forbidden", err)
		}
	})

	t.Run("the owner writes, reads back, and replaces", func(t *testing.T) {
		token, err := authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: bareLesson, Language: "uk",
			Segments: []domain.Segment{
				segment(newSegmentID(2), "перша", 0, 1000),
				segment(newSegmentID(3), "друга", 1000, 2000),
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		got, err := reading.Get(ctx, owner, bareLesson, "uk")
		if err != nil || len(got.Segments) != 2 {
			t.Fatalf("read back: segments=%d err=%v; want 2", len(got.Segments), err)
		}

		// Replacement, not merge: one segment in means one segment out, not three.
		if _, err := authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: bareLesson, Language: "uk", ExpectedUpdatedAt: token,
			Segments: []domain.Segment{segment(newSegmentID(4), "єдина", 0, 500)},
		}); err != nil {
			t.Fatal(err)
		}
		got, err = reading.Get(ctx, owner, bareLesson, "uk")
		if err != nil || len(got.Segments) != 1 {
			t.Fatalf("after replace: segments=%d err=%v; want 1", len(got.Segments), err)
		}
	})

	t.Run("a stale token is refused and the other author's work survives", func(t *testing.T) {
		first, err := reading.Get(ctx, owner, bareLesson, "uk")
		if err != nil || !first.Found {
			t.Fatal("expected a transcript to exist")
		}
		_, err = authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: bareLesson, Language: "uk",
			ExpectedUpdatedAt: "2020-01-01T00:00:00Z",
			Segments:          []domain.Segment{segment(newSegmentID(5), "затирання", 0, 100)},
		})
		if err != domain.ErrTranscriptModified {
			t.Fatalf("stale token err = %v; want transcript_modified", err)
		}
		after, err := reading.Get(ctx, owner, bareLesson, "uk")
		if err != nil || len(after.Segments) != len(first.Segments) {
			t.Fatalf("the refused write changed the transcript: %d -> %d", len(first.Segments), len(after.Segments))
		}
	})

	t.Run("a second writer into the same empty lesson loses instead of appending", func(t *testing.T) {
		// Without a token there is nothing to compare, so the unique key is what makes the
		// author's belief checkable. Appending here would also walk straight past the ceiling.
		if _, err := authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: replacedLesson, Language: "en",
			Segments: []domain.Segment{segment(newSegmentID(6), "first", 0, 100)},
		}); err != nil {
			t.Fatal(err)
		}
		_, err := authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: replacedLesson, Language: "en",
			Segments: []domain.Segment{segment(newSegmentID(7), "second", 0, 100)},
		})
		if err != domain.ErrTranscriptExists {
			t.Fatalf("second writer err = %v; want transcript_exists", err)
		}
	})

	t.Run("a document outside this course's library is a named refusal, not a 500", func(t *testing.T) {
		_, err := authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: uploadingLesson, Language: "en",
			Segments: []domain.Segment{
				segment(newSegmentID(8), "з чужим лінком", 0, 100, "60000000-0000-4000-8000-0000000000ff"),
			},
		})
		if err != domain.ErrInvalidRequest {
			t.Fatalf("unknown document err = %v; want invalid_request", err)
		}
	})

	t.Run("the write stamps the generation the player would pick", func(t *testing.T) {
		// replacedLesson has ready generations 1 and 2; writing now must stamp 2, so the text is
		// current rather than instantly stale.
		if _, err := authoring.Save(ctx, owner, application.SaveRequest{
			LessonID: replacedLesson, Language: "de",
			Segments: []domain.Segment{segment(newSegmentID(9), "aktuell", 0, 100)},
		}); err != nil {
			t.Fatal(err)
		}
		got, err := reading.Get(ctx, owner, replacedLesson, "de")
		if err != nil || got.Stale {
			t.Fatalf("freshly written transcript: stale=%v err=%v; want not stale", got.Stale, err)
		}
	})
}
