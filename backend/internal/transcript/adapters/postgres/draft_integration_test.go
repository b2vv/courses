//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	transcriptpostgres "courses/internal/transcript/adapters/postgres"
	transcriptapplication "courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	draftInstructor = "30000000-0000-4000-8000-0000000000d1"
	draftCourse     = "10000000-0000-4000-8000-0000000000d1"
	draftModule     = "20000000-0000-4000-8000-0000000000d1"

	recoverLesson1 = "40000000-0000-4000-8000-000000000001"
	recoverLesson2 = "40000000-0000-4000-8000-000000000002"
	recoverLesson3 = "40000000-0000-4000-8000-000000000003"
	recoverLesson4 = "40000000-0000-4000-8000-000000000004"

	recoverAsset1 = "50000000-0000-4000-8000-000000000001"
	recoverAsset2 = "50000000-0000-4000-8000-000000000002"
	recoverAsset3 = "50000000-0000-4000-8000-000000000003"
	recoverAsset4 = "50000000-0000-4000-8000-000000000004"
)

var draftNow = time.Date(2026, 9, 22, 12, 0, 0, 0, time.UTC)

type draftHarness struct {
	ctx        context.Context
	pool       *pgxpool.Pool
	repository *transcriptpostgres.Repository
}

func TestDraftRecoverAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("draft_recover_test"),
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

	h := &draftHarness{ctx: ctx, pool: pool, repository: transcriptpostgres.NewRepository(pool)}

	t.Run("running draft past per-row P-3 timeout is recovered", h.testExpiredDraftRecovered)
	t.Run("running draft within per-row P-3 timeout is not recovered", h.testFreshDraftNotRecovered)
	t.Run("recovered draft completed as failed timeout", h.testRecoverThenComplete)
	t.Run("concurrent re-Claim between Recover and Complete yields ErrDraftSuperseded", h.testRecoverRaceWithReClaim)
	t.Run("deadline fails running draft with timeout", h.testProcessDeadlineFailsDraft)
}

type blockingAudio struct{}

func (blockingAudio) Extract(context.Context, string) (string, func(), error) {
	return "/tmp/audio.wav", func() {}, nil
}

type blockingTranscriber struct{}

func (blockingTranscriber) Transcribe(ctx context.Context, _ string, _ domain.Language) (ports.Recognition, error) {
	<-ctx.Done()
	return ports.Recognition{}, ctx.Err()
}

func (h *draftHarness) testProcessDeadlineFailsDraft(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 0.01)
	r := transcriptapplication.NewRecognize(
		h.repository, blockingTranscriber{}, blockingAudio{}, nil, time.Now, slog.Default())
	ctx, cancel := context.WithTimeout(h.ctx, 500*time.Millisecond)
	defer cancel()
	outcome, err := r.Process(ctx, ports.RecognitionJob{
		LessonID: recoverLesson1, AssetID: recoverAsset1, Generation: 1,
	})
	if err != nil || !outcome.Failed || outcome.Reason != "timeout" {
		t.Fatalf("outcome=%+v err=%v want failed timeout", outcome, err)
	}
	var state, failure string
	if err := h.pool.QueryRow(h.ctx,
		`SELECT state, failure FROM asr_drafts WHERE lesson_id = $1::text::uuid`, recoverLesson1).
		Scan(&state, &failure); err != nil {
		t.Fatal(err)
	}
	if state != string(domain.DraftFailed) || failure != "timeout" {
		t.Fatalf("state=%s failure=%s want failed/timeout", state, failure)
	}
}

func (h *draftHarness) reset(t *testing.T) {
	t.Helper()
	statements := []struct {
		sql  string
		args []any
	}{
		{`TRUNCATE asr_drafts, video_assets, course_lessons, course_modules, courses, users CASCADE`, nil},
		{`INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at) VALUES
		($1,'draft@t.local','D','instructor','$argon2id$v=19$m=65536,t=3,p=2$c3R1ZGVudC1kZW1vLXNsdA$5S1MhXX3wgDMnIKuygFRfUNVvTtRktPECxV3ZUBg0is',$2,$2)`,
			[]any{draftInstructor, draftNow}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at) VALUES
		($1,$2,'recover','Recover','S','D','draft',$3,$3)`,
			[]any{draftCourse, draftInstructor, draftNow}},
		{`INSERT INTO course_modules(id,course_id,title,position,created_at,updated_at) VALUES($1,$2,'M',0,$3,$3)`,
			[]any{draftModule, draftCourse, draftNow}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position,created_at,updated_at) VALUES
		($1,$4,'L1','',0,$5,$5),($2,$4,'L2','',1,$5,$5),($3,$4,'L3','',2,$5,$5)`,
			[]any{recoverLesson1, recoverLesson2, recoverLesson3, draftModule, draftNow}},
	}
	for _, s := range statements {
		if _, err := h.pool.Exec(h.ctx, s.sql, s.args...); err != nil {
			t.Fatal(err)
		}
	}
}

func (h *draftHarness) insertAssetWithDuration(t *testing.T, assetID, lessonID string, generation int, durationSeconds float64) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO video_assets(
		id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,
		renditions,duration_seconds,expires_at,created_at,updated_at)
		VALUES($1,$2,$3,'ready',1024,1024,'fixture.mp4','{720p}',$4,$5,$5,$5)`,
		assetID, lessonID, generation, durationSeconds, draftNow); err != nil {
		t.Fatal(err)
	}
}

func (h *draftHarness) insertRunningDraft(t *testing.T, lessonID, assetID string, generation int32, updatedAt time.Time) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO asr_drafts(
		lesson_id,asset_id,generation,state,segments,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,$3,'running','[]'::jsonb,$4,$5)`,
		lessonID, assetID, generation, updatedAt, updatedAt); err != nil {
		t.Fatal(err)
	}
}

// P-3 formula: timeout = duration_seconds × 1s + 300s.
// A draft is expired when updated_at < now - timeout.

func (h *draftHarness) testExpiredDraftRecovered(t *testing.T) {
	h.reset(t)

	// Asset with 60s duration → timeout = 60 + 300 = 360s.
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60.0)
	// Draft updated 400s ago → 400 > 360, should be recovered.
	h.insertRunningDraft(t, recoverLesson1, recoverAsset1, 1, draftNow.Add(-400*time.Second))

	// Asset with 600s duration → timeout = 600 + 300 = 900s.
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 600.0)
	// Draft updated 1000s ago → 1000 > 900, should be recovered.
	h.insertRunningDraft(t, recoverLesson2, recoverAsset2, 1, draftNow.Add(-1000*time.Second))

	recovered, err := h.repository.Recover(h.ctx, draftNow, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 2 {
		t.Fatalf("recovered=%d; want 2", len(recovered))
	}
	for _, d := range recovered {
		if d.Draft.State != domain.DraftRunning {
			t.Errorf("lesson=%s state=%s; want running", d.Draft.LessonID, d.Draft.State)
		}
		if d.Token == "" {
			t.Errorf("lesson=%s token empty", d.Draft.LessonID)
		}
	}
}

func (h *draftHarness) testFreshDraftNotRecovered(t *testing.T) {
	h.reset(t)

	// Asset with 60s duration → timeout = 360s.
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60.0)
	// Draft updated 200s ago → 200 < 360, within timeout — must NOT be recovered.
	h.insertRunningDraft(t, recoverLesson1, recoverAsset1, 1, draftNow.Add(-200*time.Second))

	// Asset with 600s duration → timeout = 900s.
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 600.0)
	// Draft updated 800s ago → 800 < 900, within timeout — must NOT be recovered.
	h.insertRunningDraft(t, recoverLesson2, recoverAsset2, 1, draftNow.Add(-800*time.Second))

	recovered, err := h.repository.Recover(h.ctx, draftNow, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 0 {
		t.Fatalf("recovered=%d; want 0 (all within P-3 timeout)", len(recovered))
	}
}

func (h *draftHarness) testRecoverThenComplete(t *testing.T) {
	h.reset(t)

	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60.0)
	h.insertRunningDraft(t, recoverLesson1, recoverAsset1, 1, draftNow.Add(-400*time.Second))

	recovered, err := h.repository.Recover(h.ctx, draftNow, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 1 {
		t.Fatalf("recovered=%d; want 1", len(recovered))
	}

	d := recovered[0]
	completeNow := draftNow.Add(time.Second)
	err = h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID:   d.Draft.LessonID,
		AssetID:    d.Draft.AssetID,
		Generation: d.Draft.Generation,
		State:      domain.DraftFailed,
		Failure:    "timeout",
	}, d.Token, completeNow)
	if err != nil {
		t.Fatalf("Complete after Recover: %v", err)
	}

	// Verify the row is now 'failed' with reason 'timeout'.
	var state, failure string
	if err := h.pool.QueryRow(h.ctx, `SELECT state, failure FROM asr_drafts WHERE lesson_id = $1::text::uuid`,
		recoverLesson1).Scan(&state, &failure); err != nil {
		t.Fatal(err)
	}
	if state != "failed" || failure != "timeout" {
		t.Fatalf("state=%s failure=%s; want failed/timeout", state, failure)
	}
}

func (h *draftHarness) testRecoverRaceWithReClaim(t *testing.T) {
	h.reset(t)

	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60.0)
	h.insertRunningDraft(t, recoverLesson1, recoverAsset1, 1, draftNow.Add(-400*time.Second))

	// Step 1: Recover sees the draft as timed out.
	recovered, err := h.repository.Recover(h.ctx, draftNow, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovered) != 1 {
		t.Fatalf("recovered=%d; want 1", len(recovered))
	}
	recoveredToken := recovered[0].Token

	// Step 2: Between Recover reading and the caller calling Complete, a new Claim
	// arrives (e.g. regeneration or re-enqueue). This moves updated_at, so the old
	// token no longer matches.
	claimNow := draftNow.Add(time.Second)
	claimed, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, claimNow)
	if err != nil {
		t.Fatalf("re-Claim: %v", err)
	}
	if !claimed.Claimed {
		t.Fatal("re-Claim must succeed")
	}

	// Step 3: Complete with the old token must fail as superseded, not silently succeed.
	completeNow := draftNow.Add(2 * time.Second)
	err = h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID:   recoverLesson1,
		AssetID:    recoverAsset1,
		Generation: 1,
		State:      domain.DraftFailed,
		Failure:    "timeout",
	}, recoveredToken, completeNow)
	if err != domain.ErrDraftSuperseded {
		t.Fatalf("Complete with stale token err=%v; want ErrDraftSuperseded", err)
	}

	// The row must still be 'running' from the new Claim, not 'failed'.
	var state string
	if err := h.pool.QueryRow(h.ctx, `SELECT state FROM asr_drafts WHERE lesson_id = $1::text::uuid`,
		recoverLesson1).Scan(&state); err != nil {
		t.Fatal(err)
	}
	if state != "running" {
		t.Fatalf("state=%s; want running (new claim is untouched)", state)
	}
}
