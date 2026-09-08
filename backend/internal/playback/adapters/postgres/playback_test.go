package postgres_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	playbackpostgres "courses/internal/playback/adapters/postgres"
	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	playbackUserID   = "40000000-0000-4000-8000-000000000001"
	playbackLessonID = "70000000-0000-4000-8000-000000000001"
	playbackVideoID  = "50000000-0000-4000-8000-000000000001"
)

type playbackQueryer struct {
	query string
	args  []any
	row   playbackRow
}

func (database *playbackQueryer) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	database.query, database.args = query, args
	return database.row
}

type playbackRow struct {
	ownerID        string
	status         string
	videoID        string
	clientSequence int64
	videoPresent   bool
	err            error
}

func (result playbackRow) Scan(dest ...any) error {
	if result.err != nil {
		return result.err
	}
	*dest[0].(*string) = result.ownerID
	*dest[1].(*string) = result.status
	*dest[2].(*pgtype.Text) = pgtype.Text{String: result.videoID, Valid: result.videoPresent}
	*dest[3].(*int64) = result.clientSequence
	return nil
}

func TestLessonPlaybackReaderLoadsServerSelectedVideoAndProgress(t *testing.T) {
	t.Parallel()
	database := &playbackQueryer{row: playbackRow{
		ownerID: "40000000-0000-4000-8000-000000000002", status: "published",
		videoID: playbackVideoID, videoPresent: true, clientSequence: 12,
	}}
	reader := playbackpostgres.NewLessonPlaybackReader(database)

	got, err := reader.LoadLessonPlayback(context.Background(), playbackUserID, playbackLessonID)
	if err != nil {
		t.Fatal(err)
	}
	want := domain.LessonPlayback{
		VideoID: playbackVideoID, ClientSequence: 12,
		Course: domain.VideoCourse{OwnerID: "40000000-0000-4000-8000-000000000002", Status: domain.Published},
	}
	if got != want {
		t.Fatalf("playback=%+v, want %+v", got, want)
	}
	if len(database.args) != 2 || database.args[0] != playbackUserID || database.args[1] != playbackLessonID {
		t.Fatalf("args=%v", database.args)
	}
	for _, fragment := range []string{"course_lessons", "course_modules", "courses", "video_assets", "lesson_progress", "state='ready'", "generation DESC", "evicted_at IS NULL"} {
		if !strings.Contains(database.query, fragment) {
			t.Fatalf("query missing %q: %s", fragment, database.query)
		}
	}
}

func TestLessonPlaybackReaderClassifiesMissingAndCorruptRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		row  playbackRow
		want error
	}{
		{name: "lesson missing", row: playbackRow{err: pgx.ErrNoRows}, want: domain.ErrLessonNotFound},
		{name: "video not ready", row: playbackRow{ownerID: "owner", status: "published"}, want: domain.ErrVideoNotReady},
		{name: "database unavailable", row: playbackRow{err: errors.New("connection failed")}, want: domain.ErrDependencyUnavailable},
		{name: "unknown status", row: playbackRow{ownerID: "owner", status: "unknown", videoID: playbackVideoID, videoPresent: true}, want: domain.ErrDependencyUnavailable},
		{name: "negative stored sequence", row: playbackRow{ownerID: "owner", status: "published", videoID: playbackVideoID, videoPresent: true, clientSequence: -1}, want: domain.ErrDependencyUnavailable},
		{name: "oversized stored sequence", row: playbackRow{ownerID: "owner", status: "published", videoID: playbackVideoID, videoPresent: true, clientSequence: int64(domain.MaximumClientSequence) + 1}, want: domain.ErrDependencyUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := playbackpostgres.NewLessonPlaybackReader(&playbackQueryer{row: test.row}).LoadLessonPlayback(context.Background(), playbackUserID, playbackLessonID)
			if !errors.Is(err, test.want) {
				t.Fatalf("err=%v, want %v", err, test.want)
			}
		})
	}
}
