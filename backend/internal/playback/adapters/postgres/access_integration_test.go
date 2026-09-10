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

func TestVideoAccessReaderAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("playback_test"),
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
		ownerID    = "49000000-0000-4000-8000-000000000001"
		courseID   = "19000000-0000-4000-8000-000000000001"
		moduleID   = "69000000-0000-4000-8000-000000000001"
		lessonID   = "79000000-0000-4000-8000-000000000001"
		videoID    = "89000000-0000-4000-8000-000000000001"
		newVideoID = "89000000-0000-4000-8000-000000000002"
		viewerID   = "49000000-0000-4000-8000-000000000002"
	)
	fixtures := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users(id,email,display_name,role,password_hash)
		 VALUES($1,'owner@playback.test','Playback owner','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, []any{ownerID}},
		{`INSERT INTO users(id,email,display_name,role,password_hash)
		 VALUES($1,'viewer@playback.test','Playback viewer','student','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, []any{viewerID}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at)
		 VALUES($1,$2,'playback-test','Playback test','Playback test summary','Playback test description','published',now())`, []any{courseID, ownerID}},
		{`INSERT INTO course_modules(id,course_id,title,position) VALUES($1,$2,'Module',0)`, []any{moduleID, courseID}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($1,$2,'Lesson','',0)`, []any{lessonID, moduleID}},
		{`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
		 VALUES($1,$2,1,'ready',1,1,'test.mp4',ARRAY['360p'],now()+interval '1 day',now(),now())`, []any{videoID, lessonID}},
	}
	for _, fixture := range fixtures {
		if _, err := pool.Exec(ctx, fixture.query, fixture.args...); err != nil {
			t.Fatal(err)
		}
	}
	metric := &counter{}
	reader := playbackpostgres.NewVideoAccessReader(pool, metric)
	course, err := reader.LoadVideoCourse(ctx, videoID)
	if err != nil {
		t.Fatal(err)
	}
	if course != (domain.VideoCourse{OwnerID: ownerID, Status: domain.Published}) {
		t.Fatalf("course=%+v", course)
	}
	playbackReader := playbackpostgres.NewLessonPlaybackReader(pool)
	playback, err := playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID)
	if err != nil || playback.VideoID != videoID || playback.ClientSequence != 0 || playback.Course != course {
		t.Fatalf("first playback=%+v err=%v", playback, err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at)
		VALUES($1,$2,12,12,now())`, viewerID, lessonID); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,expires_at,created_at,updated_at)
		VALUES($1,$2,2,'ready',1,1,'new.mp4',ARRAY['360p'],now()+interval '1 day',now(),now())`, newVideoID, lessonID); err != nil {
		t.Fatal(err)
	}
	playback, err = playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID)
	if err != nil || playback.VideoID != newVideoID || playback.ClientSequence != 12 {
		t.Fatalf("current playback=%+v err=%v", playback, err)
	}
	if _, err := pool.Exec(ctx, `UPDATE video_assets SET evicted_at=now(),updated_at=now() WHERE id=$1`, videoID); err != nil {
		t.Fatal(err)
	}
	if _, err := reader.LoadVideoCourse(ctx, videoID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("evicted video: %v", err)
	}
	if metric.value != 2 {
		t.Fatalf("metric=%d, want 2", metric.value)
	}
	if _, err := pool.Exec(ctx, `UPDATE video_assets SET evicted_at=now(),updated_at=now() WHERE id=$1`, newVideoID); err != nil {
		t.Fatal(err)
	}
	if _, err := playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID); !errors.Is(err, domain.ErrVideoNotReady) {
		t.Fatalf("no ready video: %v", err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM course_lessons WHERE id=$1`, lessonID); err != nil {
		t.Fatal(err)
	}
	if _, err := playbackReader.LoadLessonPlayback(ctx, viewerID, lessonID); !errors.Is(err, domain.ErrLessonNotFound) {
		t.Fatalf("missing lesson: %v", err)
	}
}
