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
	recoverAsset5 = "50000000-0000-4000-8000-000000000005"
	recoverAsset6 = "50000000-0000-4000-8000-000000000006"
	recoverAsset7 = "50000000-0000-4000-8000-000000000007"
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
	t.Run("C21 rejects an asset belonging to another lesson", h.testClaimRejectsForeignAsset)
	t.Run("C22 deletes a draft when its asset is deleted", h.testAssetDeleteCascadesDraft)
	t.Run("C09 old completion after new claim is superseded", h.testOldCompletionAfterNewClaim)
	t.Run("C09 delayed old claim cannot overwrite the new draft", h.testDelayedOldClaim)
	t.Run("re-claiming the same generation succeeds", h.testReclaimSameGeneration)
	t.Run("C01 selects only ready assets missing their current draft", h.testPendingRecognitionsSelector)
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
		($1,$5,'L1','',0,$6,$6),($2,$5,'L2','',1,$6,$6),($3,$5,'L3','',2,$6,$6),($4,$5,'L4','',3,$6,$6)`,
			[]any{recoverLesson1, recoverLesson2, recoverLesson3, recoverLesson4, draftModule, draftNow}},
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

func (h *draftHarness) insertOrphanAssetWithDuration(t *testing.T, assetID string, durationSeconds float64) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO video_assets(
		id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,
		renditions,duration_seconds,expires_at,created_at,updated_at)
		VALUES($1,NULL,1,'ready',1024,1024,'orphan.mp4','{720p}',$2,$3,$3,$3)`,
		assetID, durationSeconds, draftNow); err != nil {
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

func (h *draftHarness) completeReadyDraft(t *testing.T, lessonID, assetID string, generation int32, claimAt time.Time) {
	t.Helper()
	claimed, err := h.repository.Claim(h.ctx, lessonID, assetID, generation, claimAt)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if err := h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
		State: domain.DraftReady, Language: domain.Language("en"),
		Segments: []domain.Segment{{ID: "60000000-0000-4000-8000-000000000001", EndMs: 1000, Text: "ready"}},
	}, claimed.Token, claimAt.Add(time.Second)); err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func (h *draftHarness) testClaimRejectsForeignAsset(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson2, 1, 60)
	_, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, draftNow)
	if err != domain.ErrNotFound {
		t.Fatalf("Claim err=%v; want ErrNotFound", err)
	}
	var drafts int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts WHERE lesson_id=$1::text::uuid`, recoverLesson1).Scan(&drafts); err != nil {
		t.Fatal(err)
	}
	if drafts != 0 {
		t.Fatalf("draft rows for foreign lesson=%d; want 0", drafts)
	}
}

func (h *draftHarness) testAssetDeleteCascadesDraft(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	if _, err := h.pool.Exec(h.ctx, `DELETE FROM video_assets WHERE id=$1::text::uuid`, recoverAsset1); err != nil {
		t.Fatalf("DELETE video_assets: %v", err)
	}
	var drafts int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts WHERE lesson_id=$1::text::uuid`, recoverLesson1).Scan(&drafts); err != nil {
		t.Fatal(err)
	}
	// D8-15 deliberately chooses CASCADE: SET NULL would retain a draft that names no source asset.
	if drafts != 0 {
		t.Fatalf("draft rows after asset delete=%d; want 0", drafts)
	}
}

func (h *draftHarness) testOldCompletionAfterNewClaim(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	old, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, draftNow)
	if err != nil {
		t.Fatal(err)
	}
	newClaim, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset2, 2, draftNow.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if err := h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID: recoverLesson1, AssetID: recoverAsset1, Generation: 1,
		State: domain.DraftReady, Language: "en",
		Segments: []domain.Segment{{ID: "60000000-0000-4000-8000-000000000001", EndMs: 1000, Text: "old"}},
	}, old.Token, draftNow.Add(2*time.Second)); err != domain.ErrDraftSuperseded {
		t.Fatalf("old Complete err=%v; want ErrDraftSuperseded", err)
	}
	if newClaim.Token == old.Token {
		t.Fatal("new Claim reused old optimistic-lock token")
	}
}

func (h *draftHarness) testDelayedOldClaim(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	if _, err := h.pool.Exec(h.ctx, `UPDATE video_assets SET evicted_at=$2, updated_at=$2 WHERE id=$1::text::uuid`, recoverAsset1, draftNow.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	h.completeReadyDraft(t, recoverLesson1, recoverAsset2, 2, draftNow.Add(2*time.Second))

	// The old asset is still present (only soft-evicted), so Claim must reject it at the UPSERT.
	old, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, draftNow.Add(4*time.Second))
	if err != nil {
		t.Fatalf("delayed old Claim: %v", err)
	}
	if old.Claimed {
		t.Fatal("delayed old Claim was accepted")
	}
	draft, err := h.repository.GetLessonDraft(h.ctx, recoverLesson1)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Draft.Generation != 2 || draft.Draft.Segments[0].Text != "ready" {
		t.Fatalf("delayed old worker overwrote draft: generation=%d text=%q; want generation 2/ready", draft.Draft.Generation, draft.Draft.Segments[0].Text)
	}
}

func (h *draftHarness) testReclaimSameGeneration(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	first, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, draftNow)
	if err != nil {
		t.Fatalf("first Claim: %v", err)
	}
	second, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, draftNow.Add(time.Second))
	if err != nil {
		t.Fatalf("same-generation Claim: %v", err)
	}
	if !second.Claimed {
		t.Fatal("same-generation Claim was not claimed")
	}
	if second.Token == first.Token {
		t.Fatal("same-generation Claim reused the previous optimistic-lock token")
	}
}

func (h *draftHarness) testPendingRecognitionsSelector(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.completeReadyDraft(t, recoverLesson2, recoverAsset2, 1, draftNow)
	h.insertAssetWithDuration(t, recoverAsset3, recoverLesson3, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset4, recoverLesson3, 2, 60)
	h.completeReadyDraft(t, recoverLesson3, recoverAsset3, 1, draftNow)
	h.insertAssetWithDuration(t, recoverAsset5, recoverLesson4, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset6, recoverLesson4, 2, 60)
	h.completeReadyDraft(t, recoverLesson4, recoverAsset6, 2, draftNow)
	h.insertOrphanAssetWithDuration(t, recoverAsset7, 60)
	pending, err := h.repository.PendingRecognitions(h.ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatalf("pending=%+v; want lesson1/gen1 and lesson3/gen2 only", pending)
	}
	if pending[0].LessonID != recoverLesson1 || pending[1].LessonID != recoverLesson3 || pending[1].Generation != 2 {
		t.Fatalf("pending=%+v; want lesson1/gen1 and lesson3/gen2", pending)
	}
	asset, err := h.repository.LessonAsset(h.ctx, recoverLesson3)
	if err != nil {
		t.Fatal(err)
	}
	if asset == nil || asset.AssetID != recoverAsset4 || asset.Generation != 2 {
		t.Fatalf("current asset=%+v; want asset %s/generation 2", asset, recoverAsset4)
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
