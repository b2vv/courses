//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	playbackpostgres "courses/internal/playback/adapters/postgres"
	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestProgressWriterAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("playback_progress_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
	)
	if err != nil {
		t.Fatalf("PostgreSQL integration requires Docker: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate test container: %v", err)
		}
	})
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	const (
		userID   = "41000000-0000-4000-8000-000000000001"
		courseID = "11000000-0000-4000-8000-000000000001"
		moduleID = "61000000-0000-4000-8000-000000000001"
		lessonID = "71000000-0000-4000-8000-000000000001"
		videoID  = "81000000-0000-4000-8000-000000000001"
	)
	fixtures := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users(id,email,display_name,role,password_hash)
		 VALUES($1,'progress@playback.test','Progress owner','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, []any{userID}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at)
		 VALUES($1,$2,'playback-progress-test','Playback progress','Playback progress summary','Playback progress description','published',now())`, []any{courseID, userID}},
		{`INSERT INTO course_modules(id,course_id,title,position) VALUES($1,$2,'Module',0)`, []any{moduleID, courseID}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($1,$2,'Lesson','',0)`, []any{lessonID, moduleID}},
		{`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,duration_seconds,expires_at,created_at,updated_at)
		 VALUES($1,$2,1,'ready',1,1,'test.mp4',ARRAY['360p'],100,now()+interval '1 day',now(),now())`, []any{videoID, lessonID}},
	}
	for _, fixture := range fixtures {
		if _, err := pool.Exec(ctx, fixture.query, fixture.args...); err != nil {
			t.Fatal(err)
		}
	}

	type savedProgress struct {
		position  float64
		sequence  int64
		completed bool
	}
	now := time.Date(2026, time.September, 8, 15, 0, 0, 0, time.UTC)
	prepare := func(saved *savedProgress, duration *float64) {
		t.Helper()
		if _, err := pool.Exec(ctx, `DELETE FROM lesson_progress WHERE user_id=$1 AND lesson_id=$2`, userID, lessonID); err != nil {
			t.Fatal(err)
		}
		if _, err := pool.Exec(ctx, `UPDATE video_assets SET duration_seconds=$2 WHERE id=$1`, videoID, duration); err != nil {
			t.Fatal(err)
		}
		if saved == nil {
			return
		}
		var completedAt *time.Time
		if saved.completed {
			completedAt = &now
		}
		if _, err := pool.Exec(ctx, `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,completed_at,client_seq,updated_at)
			VALUES($1,$2,$3,$4,$5,$6)`, userID, lessonID, saved.position, completedAt, saved.sequence, now.Add(-time.Minute)); err != nil {
			t.Fatal(err)
		}
	}
	read := func() savedProgress {
		t.Helper()
		var got savedProgress
		if err := pool.QueryRow(ctx, `SELECT position_seconds,client_seq,completed_at IS NOT NULL
			FROM lesson_progress WHERE user_id=$1 AND lesson_id=$2`, userID, lessonID).Scan(&got.position, &got.sequence, &got.completed); err != nil {
			t.Fatal(err)
		}
		return got
	}
	writer := playbackpostgres.NewProgressWriter(pool)
	duration100, duration90 := 100.0, 90.0
	tests := []struct {
		name     string
		saved    *savedProgress
		duration *float64
		position float64
		sequence int64
		reason   string
		want     savedProgress
	}{
		{name: "B05 below threshold remains incomplete", saved: &savedProgress{position: 90, sequence: 1}, duration: &duration100, position: 94.9, sequence: 2, reason: "tick", want: savedProgress{position: 94.9, sequence: 2}},
		{name: "B05 exact threshold completes", saved: &savedProgress{position: 90, sequence: 1}, duration: &duration100, position: 95, sequence: 2, reason: "tick", want: savedProgress{position: 95, sequence: 2, completed: true}},
		{name: "B41 replacement restore does not complete without growth", saved: &savedProgress{position: 94, sequence: 1}, duration: &duration90, position: 94, sequence: 2, reason: "seek", want: savedProgress{position: 90, sequence: 2}},
		{name: "B42 playing to the end completes", saved: &savedProgress{position: 90, sequence: 1}, duration: &duration100, position: 100, sequence: 2, reason: "ended", want: savedProgress{position: 100, sequence: 2, completed: true}},
		{name: "B44 position over duration is clamped", saved: &savedProgress{position: 80, sequence: 1}, duration: &duration90, position: 120, sequence: 2, reason: "tick", want: savedProgress{position: 90, sequence: 2, completed: true}},
		{name: "unknown duration neither clamps nor completes", saved: &savedProgress{position: 2, sequence: 1}, position: 120, sequence: 2, reason: "tick", want: savedProgress{position: 120, sequence: 2}},
		{name: "B11 stale offline replay cannot rewind progress", saved: &savedProgress{position: 80, sequence: 9}, duration: &duration100, position: 20, sequence: 4, reason: "ended", want: savedProgress{position: 80, sequence: 9}},
		{name: "periodic position regression is ignored", saved: &savedProgress{position: 80, sequence: 9}, duration: &duration100, position: 20, sequence: 10, reason: "tick", want: savedProgress{position: 80, sequence: 9}},
		{name: "event position regression is recorded", saved: &savedProgress{position: 80, sequence: 9}, duration: &duration100, position: 20, sequence: 10, reason: "seek", want: savedProgress{position: 20, sequence: 10}},
		{name: "completed lesson does not regress on exit", saved: &savedProgress{position: 100, sequence: 3, completed: true}, position: 0, sequence: 4, reason: "nav", want: savedProgress{position: 100, sequence: 3, completed: true}},
		{name: "first client sequence zero is accepted", position: 1, sequence: 0, reason: "tick", want: savedProgress{position: 1, sequence: 0}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			prepare(test.saved, test.duration)
			update, err := domain.NewProgressUpdate(test.position, test.sequence, test.reason)
			if err != nil {
				t.Fatal(err)
			}
			if err := writer.Save(ctx, userID, lessonID, update, now); err != nil {
				t.Fatal(err)
			}
			if got := read(); got != test.want {
				t.Fatalf("progress=%+v, want %+v", got, test.want)
			}
		})
	}

	update, err := domain.NewProgressUpdate(1, 1, "tick")
	if err != nil {
		t.Fatal(err)
	}
	if err := writer.Save(ctx, userID, "72000000-0000-4000-8000-000000000001", update, now); !errors.Is(err, domain.ErrLessonNotFound) {
		t.Fatalf("missing lesson: %v", err)
	}
}
