package postgres_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	playbackpostgres "courses/internal/playback/adapters/postgres"
	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5"
)

type queryer struct {
	query string
	args  []any
	row   pgx.Row
}

func TestVideoAccessReaderClassifiesAbsentUnavailableAndCorruptRows(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		row  row
		want error
	}{
		{name: "absent video", row: row{err: pgx.ErrNoRows}, want: domain.ErrForbidden},
		{name: "database unavailable", row: row{err: errors.New("connection failed")}, want: domain.ErrDependencyUnavailable},
		{name: "unknown course status", row: row{owner: "owner", status: "unknown"}, want: domain.ErrDependencyUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			metric := &counter{}
			_, err := playbackpostgres.NewVideoAccessReader(&queryer{row: test.row}, metric).LoadVideoCourse(context.Background(), "50000000-0000-4000-8000-000000000001")
			if !errors.Is(err, test.want) || metric.value != 1 {
				t.Fatalf("err=%v metric=%d, want %v and 1", err, metric.value, test.want)
			}
		})
	}
}

func (database *queryer) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	database.query, database.args = query, args
	return database.row
}

type row struct {
	owner  string
	status string
	err    error
}

func (result row) Scan(dest ...any) error {
	if result.err != nil {
		return result.err
	}
	*dest[0].(*string) = result.owner
	*dest[1].(*string) = result.status
	return nil
}

type counter struct{ value int }

func (metric *counter) Inc() { metric.value++ }

func TestVideoAccessReaderLoadsCourseForTheExactVideo(t *testing.T) {
	t.Parallel()
	metric := &counter{}
	database := &queryer{row: row{owner: "40000000-0000-4000-8000-000000000002", status: "published"}}
	reader := playbackpostgres.NewVideoAccessReader(database, metric)

	course, err := reader.LoadVideoCourse(context.Background(), "50000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if course != (domain.VideoCourse{OwnerID: "40000000-0000-4000-8000-000000000002", Status: domain.Published}) {
		t.Fatalf("course=%+v", course)
	}
	if len(database.args) != 1 || database.args[0] != "50000000-0000-4000-8000-000000000001" {
		t.Fatalf("args=%v", database.args)
	}
	for _, relation := range []string{"video_assets", "course_lessons", "course_modules", "courses"} {
		if !strings.Contains(database.query, relation) {
			t.Fatalf("query does not join %s: %s", relation, database.query)
		}
	}
	if metric.value != 1 {
		t.Fatalf("metric=%d, want 1", metric.value)
	}
}
