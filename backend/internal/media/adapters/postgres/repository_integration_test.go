//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	coursepostgres "courses/internal/course/adapters/postgres"
	coursedomain "courses/internal/course/domain"
	"courses/internal/media/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mediaInstructorA = "42000000-0000-4000-8000-000000000001"
	mediaInstructorB = "42000000-0000-4000-8000-000000000002"
	mediaCourseID    = "12000000-0000-4000-8000-000000000001"
	mediaModuleID    = "62000000-0000-4000-8000-000000000001"
	mediaLessonOne   = "72000000-0000-4000-8000-000000000001"
	mediaLessonTwo   = "72000000-0000-4000-8000-000000000002"
	mediaAssetOne    = "82000000-0000-4000-8000-000000000001"
)

var mediaNow = time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

type mediaHarness struct {
	ctx        context.Context
	connection string
	pool       *pgxpool.Pool
	repository *Repository
}

type mintTraceKey struct{}

type blockingMintTracer struct {
	reached     chan struct{}
	release     chan struct{}
	reachedOnce sync.Once
	releaseOnce sync.Once
}

func (tracer *blockingMintTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if strings.Contains(data.SQL, "FOR SHARE OF courses") {
		return context.WithValue(ctx, mintTraceKey{}, true)
	}
	return ctx
}

func (tracer *blockingMintTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	if blocked, _ := ctx.Value(mintTraceKey{}).(bool); !blocked {
		return
	}
	tracer.reachedOnce.Do(func() { close(tracer.reached) })
	<-tracer.release
}

func (tracer *blockingMintTracer) unblock() { tracer.releaseOnce.Do(func() { close(tracer.release) }) }

func TestMediaRepositoryAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("media_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)),
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
	harness := &mediaHarness{ctx: ctx, connection: connection, pool: pool, repository: NewRepository(pool)}

	t.Run("P06 replacing a video keeps two ready generations", harness.testTwoReadyGenerations)
	t.Run("B16 only expired uploads are taken over", harness.testExpiredUploadTakeover)
	t.Run("B22 generation remains monotonic after eviction", harness.testMonotonicGeneration)
	t.Run("R28 ownership and authoring writes serialize", harness.testOwnershipSerialization)
	t.Run("R38 concurrent mints map either unique collision", harness.testConcurrentMints)
	t.Run("lifecycle writes lock policy guard and advance", harness.testLifecycleWrites)
}

func (h *mediaHarness) testTwoReadyGenerations(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	replacement, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, "replacement.mp4", mediaNow.Add(time.Hour), mediaNow)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `UPDATE video_assets
		SET state='ready', received_bytes=declared_bytes, renditions=ARRAY['360p'], updated_at=$2
		WHERE id=$1`, replacement.ID, mediaNow.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	assets, err := h.repository.ListCourseVideos(h.ctx, ownerActor(), mediaCourseID)
	if err != nil {
		t.Fatal(err)
	}
	if len(assets) != 2 || assets[0].Generation != 2 || assets[1].Generation != 1 ||
		assets[0].State != domain.Ready || assets[1].State != domain.Ready ||
		assets[0].EvictedAt != nil || assets[1].EvictedAt != nil {
		t.Fatalf("ready generations=%+v", assets)
	}
}

func (h *mediaHarness) testExpiredUploadTakeover(t *testing.T) {
	t.Run("created", func(t *testing.T) {
		h.reset(t)
		h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Created, mediaNow.Add(-time.Second), nil)
		minted, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, "new.mp4", mediaNow.Add(time.Hour), mediaNow)
		if err != nil || minted.Generation != 2 {
			t.Fatalf("minted=%+v err=%v", minted, err)
		}
		var evicted bool
		if err := h.pool.QueryRow(h.ctx, `SELECT evicted_at=$2 FROM video_assets WHERE id=$1`, mediaAssetOne, mediaNow).Scan(&evicted); err != nil || !evicted {
			t.Fatalf("expired asset evicted=%v err=%v", evicted, err)
		}
	})
	t.Run("processing", func(t *testing.T) {
		h.reset(t)
		h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Processing, mediaNow.Add(-time.Second), nil)
		_, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, "blocked.mp4", mediaNow.Add(time.Hour), mediaNow)
		if !errors.Is(err, domain.ErrUploadInProgress) {
			t.Fatalf("mint error=%v", err)
		}
		var count int
		var evicted bool
		if err := h.pool.QueryRow(h.ctx, `SELECT count(*), bool_or(evicted_at IS NOT NULL) FROM video_assets WHERE lesson_id=$1`, mediaLessonOne).Scan(&count, &evicted); err != nil {
			t.Fatal(err)
		}
		if count != 1 || evicted {
			t.Fatalf("processing rows=%d evicted=%v", count, evicted)
		}
	})
}

func (h *mediaHarness) testMonotonicGeneration(t *testing.T) {
	h.reset(t)
	evictedAt := mediaNow.Add(-time.Hour)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 7, domain.Ready, mediaNow.Add(time.Hour), &evictedAt)
	minted, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, "generation-eight.mp4", mediaNow.Add(time.Hour), mediaNow)
	if err != nil || minted.Generation != 8 {
		t.Fatalf("minted=%+v err=%v", minted, err)
	}
}

func (h *mediaHarness) testOwnershipSerialization(t *testing.T) {
	t.Run("committed reassignment is observed", func(t *testing.T) {
		h.reset(t)
		tx, err := h.pool.Begin(h.ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(context.Background())
		if _, err := tx.Exec(h.ctx, `UPDATE courses SET instructor_id=$2 WHERE id=$1`, mediaCourseID, mediaInstructorB); err != nil {
			t.Fatal(err)
		}
		result := make(chan error, 1)
		go func() {
			_, err := h.repository.Mint(h.ctx, domain.Actor{UserID: mediaInstructorB}, mediaLessonOne, 1024, "reassigned.mp4", mediaNow.Add(time.Hour), mediaNow)
			result <- err
		}()
		h.waitForLock(t, "FOR SHARE OF course")
		if err := tx.Commit(h.ctx); err != nil {
			t.Fatal(err)
		}
		if err := <-result; err != nil {
			t.Fatalf("mint after reassignment: %v", err)
		}
	})

	t.Run("mint refuses old owner while reassignment waits", func(t *testing.T) {
		h.reset(t)
		repository, tracedPool, tracer := h.tracedRepository(t)
		defer tracedPool.Close()
		defer tracer.unblock()
		minted := make(chan error, 1)
		go func() {
			_, err := repository.Mint(h.ctx, domain.Actor{UserID: mediaInstructorB}, mediaLessonOne, 1024, "refused-before-change.mp4", mediaNow.Add(time.Hour), mediaNow)
			minted <- err
		}()
		<-tracer.reached
		reassigned := make(chan error, 1)
		go func() {
			_, err := h.pool.Exec(h.ctx, `UPDATE courses SET instructor_id=$2 WHERE id=$1`, mediaCourseID, mediaInstructorB)
			reassigned <- err
		}()
		h.waitForLock(t, "UPDATE courses SET instructor_id")
		tracer.unblock()
		if err := <-minted; !errors.Is(err, domain.ErrCourseForbidden) {
			t.Fatalf("mint result=%v", err)
		}
		if err := <-reassigned; err != nil {
			t.Fatalf("reassignment result: %v", err)
		}
	})

	t.Run("concurrent authoring save does not deadlock", func(t *testing.T) {
		h.reset(t)
		repository, tracedPool, tracer := h.tracedRepository(t)
		defer tracedPool.Close()
		defer tracer.unblock()
		courses := coursepostgres.NewRepository(h.pool)
		tree, err := courses.Load(h.ctx, mediaCourseID)
		if err != nil {
			t.Fatal(err)
		}
		tree.Modules[0].Title = "Edited while minting"
		results := make(chan error, 2)
		go func() {
			_, err := repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, "alongside-save.mp4", mediaNow.Add(time.Hour), mediaNow)
			results <- err
		}()
		<-tracer.reached
		go func() {
			_, err := courses.Save(h.ctx, coursedomain.Actor{UserID: mediaInstructorA, Role: coursedomain.ActorInstructor}, tree, tree.UpdatedAt, tree.UpdatedAt.Add(time.Second))
			results <- err
		}()
		h.waitForLock(t, "LockAuthoringCourse")
		tracer.unblock()
		for range 2 {
			if err := <-results; err != nil {
				t.Fatalf("concurrent mint/save: %v", err)
			}
		}
	})
}

func (h *mediaHarness) testConcurrentMints(t *testing.T) {
	t.Run("pending index", func(t *testing.T) {
		h.reset(t)
		if _, err := h.pool.Exec(h.ctx, `DROP INDEX video_assets_generation`); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if _, err := h.pool.Exec(context.Background(), `CREATE UNIQUE INDEX video_assets_generation ON video_assets (lesson_id, generation)`); err != nil {
				t.Errorf("restore generation index: %v", err)
			}
		}()
		h.assertConcurrentMintOutcomes(t)
	})

	t.Run("generation index", func(t *testing.T) {
		h.reset(t)
		h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
		if _, err := h.pool.Exec(h.ctx, `DROP INDEX video_assets_pending_per_lesson`); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if _, err := h.pool.Exec(context.Background(), `CREATE UNIQUE INDEX video_assets_pending_per_lesson ON video_assets (lesson_id)
				WHERE lesson_id IS NOT NULL AND evicted_at IS NULL AND state IN ('created','uploading','queued','processing')`); err != nil {
				t.Errorf("restore pending index: %v", err)
			}
		}()
		h.assertConcurrentMintOutcomes(t)
	})
}

func (h *mediaHarness) assertConcurrentMintOutcomes(t *testing.T) {
	t.Helper()
	start := make(chan struct{})
	results := make(chan error, 2)
	for _, name := range []string{"first.mp4", "second.mp4"} {
		go func() {
			<-start
			_, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, name, mediaNow.Add(time.Hour), mediaNow)
			results <- err
		}()
	}
	close(start)
	var successes, conflicts int
	for range 2 {
		err := <-results
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrUploadInProgress):
			conflicts++
		default:
			t.Fatalf("mint result=%v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("successes=%d conflicts=%d", successes, conflicts)
	}
}

func (h *mediaHarness) testLifecycleWrites(t *testing.T) {
	h.reset(t)
	asset, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonOne, 1024, "lifecycle.mp4", mediaNow.Add(time.Hour), mediaNow)
	if err != nil {
		t.Fatal(err)
	}
	policyFailure := errors.New("storage unavailable")
	if _, err := h.repository.AppendChunk(h.ctx, asset.ID, 0, mediaNow, func(domain.VideoAsset) (int64, error) {
		return 0, policyFailure
	}); !errors.Is(err, policyFailure) {
		t.Fatalf("append policy error=%v", err)
	}
	if stored, err := h.repository.Load(h.ctx, asset.ID); err != nil || stored.State != domain.Created || stored.ReceivedBytes != 0 {
		t.Fatalf("after refused append asset=%+v err=%v", stored, err)
	}
	asset, err = h.repository.AppendChunk(h.ctx, asset.ID, 0, mediaNow.Add(time.Second), func(locked domain.VideoAsset) (int64, error) {
		if locked.State != domain.Created || locked.ReceivedBytes != 0 {
			t.Fatalf("append policy asset=%+v", locked)
		}
		return locked.DeclaredBytes, nil
	})
	if err != nil || asset.State != domain.Uploading || asset.ReceivedBytes != asset.DeclaredBytes {
		t.Fatalf("appended asset=%+v err=%v", asset, err)
	}
	if _, err := h.repository.Complete(h.ctx, asset.ID, mediaNow.Add(2*time.Second), func(domain.VideoAsset) error { return policyFailure }); !errors.Is(err, policyFailure) {
		t.Fatalf("complete policy error=%v", err)
	}
	asset, err = h.repository.Complete(h.ctx, asset.ID, mediaNow.Add(2*time.Second), func(locked domain.VideoAsset) error {
		if locked.State != domain.Uploading || locked.ReceivedBytes != locked.DeclaredBytes {
			t.Fatalf("complete policy asset=%+v", locked)
		}
		return nil
	})
	if err != nil || asset.State != domain.Queued {
		t.Fatalf("completed asset=%+v err=%v", asset, err)
	}
	asset, claimed, err := h.repository.Claim(h.ctx, asset.ID, "11111111111141118111111111111111", mediaNow.Add(3*time.Second))
	if err != nil || !claimed || asset.State != domain.Processing {
		t.Fatalf("claimed=%v asset=%+v err=%v", claimed, asset, err)
	}
	current, claimed, err := h.repository.Claim(h.ctx, asset.ID, "22222222222242228222222222222222", mediaNow.Add(4*time.Second))
	if err != nil || claimed || current.State != domain.Processing {
		t.Fatalf("second claim=%v asset=%+v err=%v", claimed, current, err)
	}
	if _, err := h.repository.Retry(h.ctx, ownerActor(), asset.ID, 40*time.Minute, mediaNow.Add(4*time.Second)); !errors.Is(err, domain.ErrVideoNotRetryable) {
		t.Fatalf("fresh retry error=%v", err)
	}
	retried, err := h.repository.Retry(h.ctx, ownerActor(), asset.ID, 40*time.Minute, mediaNow.Add(44*time.Minute))
	if err != nil || retried.State != domain.Queued || retried.Attempt != 1 || retried.ClaimToken != "" {
		t.Fatalf("retried asset=%+v err=%v", retried, err)
	}
	cancellable, err := h.repository.Mint(h.ctx, ownerActor(), mediaLessonTwo, 1024, "cancel.mp4", mediaNow.Add(time.Hour), mediaNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := h.repository.Cancel(h.ctx, cancellable.ID, mediaNow.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	if _, err := h.repository.Load(h.ctx, cancellable.ID); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("cancelled load error=%v", err)
	}
}

func (h *mediaHarness) reset(t *testing.T) {
	t.Helper()
	statements := []struct {
		sql       string
		arguments []any
	}{
		{`TRUNCATE video_assets, course_lessons, course_modules, courses, users CASCADE`, nil},
		{`INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at) VALUES
		($1,'media-a@courseforge.local','Media A','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',$3,$3),
		($2,'media-b@courseforge.local','Media B','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',$3,$3)`,
			[]any{mediaInstructorA, mediaInstructorB, mediaNow}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at) VALUES
		($1,$2,'media-course','Media course','Media summary','Media description','draft',$3,$3)`,
			[]any{mediaCourseID, mediaInstructorA, mediaNow}},
		{`INSERT INTO course_modules(id,course_id,title,position,created_at,updated_at) VALUES($1,$2,'Media module',0,$3,$3)`,
			[]any{mediaModuleID, mediaCourseID, mediaNow}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position,created_at,updated_at) VALUES
		($1,$3,'Lesson one','',0,$4,$4),($2,$3,'Lesson two','',1,$4,$4)`,
			[]any{mediaLessonOne, mediaLessonTwo, mediaModuleID, mediaNow}},
	}
	for _, statement := range statements {
		if _, err := h.pool.Exec(h.ctx, statement.sql, statement.arguments...); err != nil {
			t.Fatal(err)
		}
	}
}

func (h *mediaHarness) insertAsset(t *testing.T, id, lessonID string, generation int, state domain.State, expiresAt time.Time, evictedAt *time.Time) {
	t.Helper()
	received := int64(0)
	claimToken := ""
	failureReason := ""
	renditions := []string{}
	if state == domain.Queued || state == domain.Processing || state == domain.Ready || state == domain.Failed {
		received = 1024
	}
	if state == domain.Processing {
		claimToken = "33333333333343338333333333333333"
	}
	if state == domain.Ready {
		renditions = []string{"360p"}
	}
	if state == domain.Failed {
		failureReason = string(domain.RenderFailed)
	}
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO video_assets(
		id,lesson_id,generation,state,claim_token,declared_bytes,received_bytes,original_name,
		failure_reason,renditions,expires_at,evicted_at,created_at,updated_at)
		VALUES($1,$2,$3,$4,NULLIF($5,'')::uuid,1024,$6,'fixture.mp4',NULLIF($7,''),$8,$9,$10,$11,$11)`,
		id, lessonID, generation, string(state), claimToken, received, failureReason, renditions, expiresAt, evictedAt, mediaNow.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
}

func (h *mediaHarness) waitForLock(t *testing.T, queryFragment string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		var blocked bool
		err := h.pool.QueryRow(h.ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_stat_activity
			WHERE datname=current_database() AND wait_event_type='Lock' AND position($1 in query) > 0
		)`, queryFragment).Scan(&blocked)
		if err != nil {
			t.Fatal(err)
		}
		if blocked {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("query containing %q did not block", queryFragment)
}

func (h *mediaHarness) tracedRepository(t *testing.T) (*Repository, *pgxpool.Pool, *blockingMintTracer) {
	t.Helper()
	config, err := pgxpool.ParseConfig(h.connection)
	if err != nil {
		t.Fatal(err)
	}
	tracer := &blockingMintTracer{reached: make(chan struct{}), release: make(chan struct{})}
	config.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(h.ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	return NewRepository(pool), pool, tracer
}

func ownerActor() domain.Actor { return domain.Actor{UserID: mediaInstructorA} }
