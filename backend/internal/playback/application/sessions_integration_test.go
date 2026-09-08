//go:build integration

package application_test

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	playbackpostgres "courses/internal/playback/adapters/postgres"
	playbackredis "courses/internal/playback/adapters/redis"
	"courses/internal/playback/application"
	"courses/internal/playback/domain"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSessionsOwnershipFailuresCannotWriteProgress(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	redisContainer, err := testcontainers.Run(
		ctx,
		"redis:7.4-bookworm",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections").WithStartupTimeout(time.Minute)),
	)
	if err != nil {
		t.Fatalf("sessions integration requires Docker Redis: %v", err)
	}
	t.Cleanup(func() {
		if err := redisContainer.Terminate(context.Background()); err != nil {
			t.Errorf("terminate Redis container: %v", err)
		}
	})
	redisEndpoint, err := redisContainer.PortEndpoint(ctx, "6379/tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	redisConnection := redisclient.NewClient(&redisclient.Options{Addr: redisEndpoint})
	t.Cleanup(func() { _ = redisConnection.Close() })

	postgresContainer, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("playback_sessions_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)),
	)
	if err != nil {
		t.Fatalf("sessions integration requires Docker PostgreSQL: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(context.Background()); err != nil {
			t.Errorf("terminate PostgreSQL container: %v", err)
		}
	})
	postgresConnection, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", postgresConnection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	migrations := filepath.Join("..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, postgresConnection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	const (
		courseID = "12000000-0000-4000-8000-000000000001"
		moduleID = "62000000-0000-4000-8000-000000000001"
	)
	fixtures := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO users(id,email,display_name,role,password_hash)
		 VALUES($1,'sessions@playback.test','Sessions owner','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA')`, []any{userID}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,published_at)
		 VALUES($1,$2,'playback-sessions-test','Playback sessions','Playback sessions summary','Playback sessions description','published',now())`, []any{courseID, userID}},
		{`INSERT INTO course_modules(id,course_id,title,position) VALUES($1,$2,'Module',0)`, []any{moduleID, courseID}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position) VALUES($1,$2,'Lesson','',0)`, []any{lessonID, moduleID}},
		{`INSERT INTO video_assets(id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,renditions,duration_seconds,expires_at,created_at,updated_at)
		 VALUES($1,$2,1,'ready',1,1,'test.mp4',ARRAY['360p'],100,now()+interval '1 day',now(),now())`, []any{videoID, lessonID}},
		{`INSERT INTO lesson_progress(user_id,lesson_id,position_seconds,client_seq,updated_at)
		 VALUES($1,$2,10,1,now())`, []any{userID, lessonID}},
	}
	for _, fixture := range fixtures {
		if _, err := pool.Exec(ctx, fixture.query, fixture.args...); err != nil {
			t.Fatal(err)
		}
	}

	now := time.Date(2026, time.September, 8, 16, 0, 0, 0, time.UTC)
	sessions := application.NewSessions(
		playbackredis.NewOwnershipWriter(redisConnection),
		playbackpostgres.NewProgressWriter(pool),
		3*time.Minute,
		func() time.Time { return now },
	)
	readProgress := func() (float64, int64) {
		t.Helper()
		var position float64
		var sequence int64
		if err := pool.QueryRow(ctx, `SELECT position_seconds,client_seq FROM lesson_progress WHERE user_id=$1 AND lesson_id=$2`, userID, lessonID).Scan(&position, &sequence); err != nil {
			t.Fatal(err)
		}
		return position, sequence
	}
	update := func(position float64, sequence int64) domain.ProgressUpdate {
		t.Helper()
		value, err := domain.NewProgressUpdate(position, sequence, "tick")
		if err != nil {
			t.Fatal(err)
		}
		return value
	}

	t.Run("R31 missing playback returns expired without progress", func(t *testing.T) {
		if err := redisConnection.Del(ctx, "playback:owner:"+userID).Err(); err != nil {
			t.Fatal(err)
		}
		if err := sessions.Heartbeat(ctx, userID, playbackID, lessonID, update(20, 2)); !errors.Is(err, domain.ErrPlaybackSessionExpired) {
			t.Fatalf("heartbeat: %v", err)
		}
		if position, sequence := readProgress(); position != 10 || sequence != 1 {
			t.Fatalf("position=%v sequence=%d", position, sequence)
		}
	})

	t.Run("R33 taken over playback cannot write progress", func(t *testing.T) {
		if err := sessions.Take(ctx, userID, domain.Ownership{PlaybackID: "22222222222222222222222222222222", VideoID: videoID, LessonID: lessonID}); err != nil {
			t.Fatal(err)
		}
		if err := sessions.Heartbeat(ctx, userID, playbackID, lessonID, update(20, 2)); !errors.Is(err, domain.ErrPlaybackTakenOver) {
			t.Fatalf("heartbeat: %v", err)
		}
		if position, sequence := readProgress(); position != 10 || sequence != 1 {
			t.Fatalf("position=%v sequence=%d", position, sequence)
		}
	})

	t.Run("owner renews and writes progress", func(t *testing.T) {
		if err := sessions.Take(ctx, userID, domain.Ownership{PlaybackID: playbackID, VideoID: videoID, LessonID: lessonID}); err != nil {
			t.Fatal(err)
		}
		if err := sessions.Heartbeat(ctx, userID, playbackID, lessonID, update(20, 2)); err != nil {
			t.Fatal(err)
		}
		if position, sequence := readProgress(); position != 20 || sequence != 2 {
			t.Fatalf("position=%v sequence=%d", position, sequence)
		}
	})

	t.Run("B35 huge sequence is rejected before persistence", func(t *testing.T) {
		if _, err := domain.NewProgressUpdate(30, 1<<62, "tick"); !errors.Is(err, domain.ErrInvalidProgress) {
			t.Fatalf("validation: %v", err)
		}
		if position, sequence := readProgress(); position != 20 || sequence != 2 {
			t.Fatalf("position=%v sequence=%d", position, sequence)
		}
	})
}
