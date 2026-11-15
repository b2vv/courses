//go:build integration

package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	containers "courses/internal/testsupport/containers"
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
	container, err := containers.Start(ctx, func(ctx context.Context) (*postgrescontainer.PostgresContainer, error) {
		return postgrescontainer.Run(ctx, "postgres:17.5",
			postgrescontainer.WithDatabase("draft_recover_test"),
			postgrescontainer.WithUsername("courses"),
			postgrescontainer.WithPassword("ephemeral-test-only"),
			testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
		)
	})
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
	t.Run("M2-R1 migration 00031 maps populated drafts and round trips", func(t *testing.T) { h.testDraftTrackMigration(t, provider) })

	t.Run("M2-R8 running draft past per-row P-3 timeout is recovered", h.testExpiredDraftRecovered)
	t.Run("M2-R8 running draft within per-row P-3 timeout is not recovered", h.testFreshDraftNotRecovered)
	t.Run("M2-R8 recovered draft completed as failed timeout", h.testRecoverThenComplete)
	t.Run("M2-R8 concurrent re-Claim between Recover and Complete yields ErrDraftSuperseded", h.testRecoverRaceWithReClaim)
	t.Run("deadline fails running draft with timeout", h.testProcessDeadlineFailsDraft)
	t.Run("M2-R3 C21 rejects foreign lesson asset track non-ready and generation", h.testClaimRejectsForeignAsset)
	t.Run("C22 deletes a draft when its asset is deleted", h.testAssetDeleteCascadesDraft)
	t.Run("C08 lesson deletion removes the draft before completion", h.testLessonDeleteCascadesDraft)
	t.Run("M2-R4 C09 slow old completion stays hidden after asset replacement", h.testOldCompletionAfterNewClaim)
	t.Run("M2-R4 C09 delayed old Ready track claim is superseded", h.testDelayedOldClaim)
	t.Run("M2-R4 re-claiming the same track fences slow completion", h.testReclaimSameGeneration)
	t.Run("M2-R2 C01 selects only current Ready assets missing drafts", h.testPendingRecognitionsSelector)
	t.Run("M2-R2 S6 pending follows each ready track regardless of default", h.testPendingPerTrack)
	t.Run("M2-R3 S6 claim records selected track leaf language and mode", h.testClaimRecordsTrack)
	t.Run("M2-R7 S6 automatic duplicate preserves stale output until requested", h.testAutomaticDuplicateNotReclaimed)
	t.Run("M2-R7 S6 staleness follows track language and mode", h.testDraftStaleAfterMetadataChange)
	t.Run("S6 a rerun leaves the published transcript untouched", h.testRerunLeavesTranscriptUntouched)
	t.Run("M2-R3 S6 joined source leaf keeps audio track shape", h.testTrackSourceLeafShape)
	t.Run("M2-R8 tracks recover and complete independently", h.testTracksRecoverIndependently)
	t.Run("M2-R9 deleting one track preserves sibling draft and published language", h.testTrackDeletePreservesTranscript)
}

type blockingAudio struct{}

func (blockingAudio) Extract(context.Context, string, string) (string, func(), error) {
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
		LessonID: recoverLesson1, AssetID: recoverAsset1, Generation: 1, TrackID: initialTrackID(recoverAsset1),
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
		{`INSERT INTO course_lessons(id,module_id,title,position,created_at,updated_at) VALUES
		($1,$5,'L1',0,$6,$6),($2,$5,'L2',1,$6,$6),($3,$5,'L3',2,$6,$6),($4,$5,'L4',3,$6,$6)`,
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
	h.insertTrack(t, initialTrackID(assetID), assetID, "embedded", "en", "ready", true)
}

func initialTrackID(assetID string) string { return "a" + assetID[1:] }

func (h *draftHarness) insertTrack(t *testing.T, id, assetID, kind, language, state string, isDefault bool) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(
		id,video_asset_id,kind,language,label,is_default,state,source_leaf,declared_bytes,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,$3,$4,'Track',$5,$6,
		CASE WHEN $3='uploaded' THEN 'audio-'||$1 ELSE 'original' END,
		CASE WHEN $3='uploaded' THEN 1024 ELSE 0 END,$7,$7)`,
		id, assetID, kind, language, isDefault, state, draftNow); err != nil {
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
	claimed, err := h.repository.Claim(h.ctx, lessonID, assetID, generation, initialTrackID(assetID), false, updatedAt)
	if err != nil || !claimed.Claimed {
		t.Fatalf("seed running draft: claim=%+v err=%v", claimed, err)
	}
}

func (h *draftHarness) completeReadyDraft(t *testing.T, lessonID, assetID string, generation int32, claimAt time.Time) {
	t.Helper()
	claimed, err := h.repository.Claim(h.ctx, lessonID, assetID, generation, initialTrackID(assetID), false, claimAt)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if err := h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID: lessonID, AssetID: assetID, Generation: generation, TrackID: initialTrackID(assetID),
		State: domain.DraftReady, Language: domain.Language("en"),
		Segments: []domain.Segment{{ID: "60000000-0000-4000-8000-000000000001", EndMs: 1000, Text: "ready"}},
	}, claimed.Token, claimAt.Add(time.Second)); err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func (h *draftHarness) testClaimRejectsForeignAsset(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "processing", false)
	for _, item := range []struct {
		name, lesson, asset, track string
		generation                 int32
	}{
		{"foreign lesson", recoverLesson2, recoverAsset1, initialTrackID(recoverAsset1), 1},
		{"foreign asset", recoverLesson1, recoverAsset2, initialTrackID(recoverAsset2), 1},
		{"foreign track", recoverLesson1, recoverAsset1, initialTrackID(recoverAsset2), 1},
		{"missing track", recoverLesson1, recoverAsset1, recoverTrack3, 1},
		{"non-ready track", recoverLesson1, recoverAsset1, recoverTrack2, 1},
		{"wrong generation", recoverLesson1, recoverAsset1, initialTrackID(recoverAsset1), 2},
	} {
		t.Run(item.name, func(t *testing.T) {
			claim, err := h.repository.Claim(h.ctx, item.lesson, item.asset, item.generation, item.track, false, draftNow)
			if err != domain.ErrNotFound || claim.Claimed {
				t.Fatalf("claim=%+v err=%v; want NotFound", claim, err)
			}
		})
	}
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("invalid claims wrote %d drafts, err=%v", count, err)
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

func (h *draftHarness) testLessonDeleteCascadesDraft(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	claimed, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
	if err != nil {
		t.Fatalf("Claim: %v", err)
	}
	if !claimed.Claimed {
		t.Fatal("Claim must create a running draft")
	}

	if _, err := h.pool.Exec(h.ctx, `DELETE FROM course_lessons WHERE id=$1::text::uuid`, recoverLesson1); err != nil {
		t.Fatalf("DELETE course_lessons: %v", err)
	}
	var drafts int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts WHERE lesson_id=$1::text::uuid`, recoverLesson1).Scan(&drafts); err != nil {
		t.Fatal(err)
	}
	if drafts != 0 {
		t.Fatalf("draft rows after lesson delete=%d; want 0", drafts)
	}

	err = h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID: recoverLesson1, AssetID: recoverAsset1, Generation: 1, TrackID: initialTrackID(recoverAsset1),
		State: domain.DraftReady, Language: domain.Language("en"),
		Segments: []domain.Segment{{ID: "60000000-0000-4000-8000-000000000002", EndMs: 1000, Text: "ready"}},
	}, claimed.Token, draftNow.Add(time.Second))
	if err != domain.ErrNotFound {
		t.Fatalf("Complete err=%v; want ErrNotFound", err)
	}
}

func (h *draftHarness) testOldCompletionAfterNewClaim(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	old, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
	if err != nil || !old.Claimed {
		t.Fatalf("old claim=%+v err=%v", old, err)
	}
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset2, 2, draftNow.Add(time.Second))
	// Different assets have different rows: old completion may finish, but is never offered.
	h.completeDraftText(t, old.Draft, old.Token, "old", draftNow.Add(3*time.Second))
	drafts, err := h.repository.ListLessonDrafts(h.ctx, recoverLesson1)
	if err != nil || len(drafts) != 1 {
		t.Fatalf("drafts=%+v err=%v", drafts, err)
	}
	if drafts[0].Draft.TrackID != initialTrackID(recoverAsset2) || drafts[0].Draft.Segments[0].Text != "ready" {
		t.Fatalf("old output was offered: %+v", drafts)
	}
}

func (h *draftHarness) testDelayedOldClaim(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset2, 2, draftNow)
	// Keep both assets Ready and non-evicted: latest-generation membership is the guard.
	old, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow.Add(4*time.Second))
	if err != domain.ErrDraftSuperseded || old.Claimed {
		t.Fatalf("old claim=%+v err=%v; want Superseded", old, err)
	}
	drafts, err := h.repository.ListLessonDrafts(h.ctx, recoverLesson1)
	if err != nil || len(drafts) != 1 || drafts[0].Draft.Generation != 2 || drafts[0].Draft.Segments[0].Text != "ready" {
		t.Fatalf("delayed claim changed current drafts=%+v err=%v", drafts, err)
	}
}

func (h *draftHarness) testReclaimSameGeneration(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	first, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
	if err != nil {
		t.Fatalf("first Claim: %v", err)
	}
	// The author's Regenerate re-claims the same generation (requested); an automatic duplicate
	// may not — see testAutomaticDuplicateNotReclaimed.
	second, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), true, draftNow.Add(time.Second))
	if err != nil {
		t.Fatalf("same-generation Claim: %v", err)
	}
	if !second.Claimed {
		t.Fatal("same-generation Claim was not claimed")
	}
	if second.Token == first.Token {
		t.Fatal("same-generation Claim reused the previous optimistic-lock token")
	}
	h.completeDraftText(t, second.Draft, second.Token, "new attempt", draftNow.Add(2*time.Second))
	first.Draft.State, first.Draft.Language = domain.DraftReady, "en"
	first.Draft.Segments = []domain.Segment{{ID: "60000000-0000-4000-8000-000000000001", EndMs: 1000, Text: "slow old attempt"}}
	if err := h.repository.Complete(h.ctx, first.Draft, first.Token, draftNow.Add(3*time.Second)); err != domain.ErrDraftSuperseded {
		t.Fatalf("slow Complete err=%v; want Superseded", err)
	}
	if got := h.lessonDraft(t, recoverLesson1, first.Draft.TrackID); got.Draft.Segments[0].Text != "new attempt" {
		t.Fatalf("slow worker overwrote new text: %+v", got)
	}
}

func (h *draftHarness) testPendingRecognitionsSelector(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.completeReadyDraft(t, recoverLesson2, recoverAsset2, 1, draftNow)
	h.insertAssetWithDuration(t, recoverAsset3, recoverLesson3, 1, 60)
	h.completeReadyDraft(t, recoverLesson3, recoverAsset3, 1, draftNow)
	h.insertAssetWithDuration(t, recoverAsset4, recoverLesson3, 2, 60)
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
	asset, err := h.repository.LessonTrack(h.ctx, recoverLesson3, initialTrackID(recoverAsset4))
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
		TrackID:    d.Draft.TrackID,
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
	// arrives. This moves updated_at, so the old token no longer matches. Since multi-audio S6
	// an existing draft is re-claimed only by the author's
	// Regenerate (requested); an automatic re-enqueue is refused (D-S6-2).
	claimNow := draftNow.Add(time.Second)
	claimed, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), true, claimNow)
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
		TrackID:    initialTrackID(recoverAsset1),
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

// M2-R1 exercises the migration itself against the old populated schema, not hand-built v31 rows.
func (h *draftHarness) testDraftTrackMigration(t *testing.T, provider *goose.Provider) {
	if _, err := provider.DownTo(h.ctx, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(h.ctx, 30); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Restore the shared harness schema even when an assertion fails mid-migration.
		if _, err := provider.Up(context.Background()); err != nil {
			t.Errorf("restore head: %v", err)
		}
	})
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset3, recoverLesson3, 1, 60)
	h.exec(t, `UPDATE audio_tracks SET language='EN-US' WHERE id=$1::text::uuid`, initialTrackID(recoverAsset1))
	h.insertTrack(t, recoverTrack2, recoverAsset2, "uploaded", "UK", "ready", false)
	// C's leaf names a real but failed track: no non-failed match may survive backfill.
	h.insertTrack(t, recoverTrack3, recoverAsset3, "uploaded", "de", "failed", false)
	for _, item := range []struct {
		lesson, asset string
		leaf          *string
	}{
		{recoverLesson1, recoverAsset1, nil},
		{recoverLesson2, recoverAsset2, stringPointer("audio-" + recoverTrack2)},
		{recoverLesson3, recoverAsset3, stringPointer("audio-" + recoverTrack3)},
	} {
		h.exec(t, `INSERT INTO asr_drafts(lesson_id,asset_id,generation,state,segments,source_leaf,created_at,updated_at)
			VALUES($1::text::uuid,$2::text::uuid,1,'running','[]'::jsonb,$3,$4,$4)`, item.lesson, item.asset, item.leaf, draftNow)
	}
	if _, err := provider.UpTo(h.ctx, 31); err != nil {
		t.Fatalf("00031 populated up: %v", err)
	}
	assertMapped := func(stage string, want map[string]string, wantTracks int) {
		t.Helper()
		rows, err := h.pool.Query(h.ctx, `SELECT lesson_id::text,audio_track_id::text,track_language,track_multilingual FROM asr_drafts`)
		if err != nil {
			t.Fatal(err)
		}
		defer rows.Close()
		got := map[string]string{}
		for rows.Next() {
			var lesson, track, language string
			var mode bool
			if err := rows.Scan(&lesson, &track, &language, &mode); err != nil {
				t.Fatal(err)
			}
			if mode {
				t.Fatalf("%s: migrated draft mode=true", stage)
			}
			got[lesson] = track + "|" + language
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			t.Fatalf("%s: mapped drafts=%v; want %v", stage, got, want)
		}
		var total, modes int
		if err := h.pool.QueryRow(h.ctx, `SELECT count(*),count(*) FILTER (WHERE multilingual) FROM audio_tracks`).Scan(&total, &modes); err != nil || total != wantTracks || modes != 0 {
			t.Fatalf("%s: tracks=%d multilingual=%d err=%v", stage, total, modes, err)
		}
	}
	assertMapped("up", map[string]string{
		recoverLesson1: initialTrackID(recoverAsset1) + "|en-us", // A embedded
		recoverLesson2: recoverTrack2 + "|uk",                    // B uploaded; C deleted
	}, 5)
	var primaryKey string
	if err := h.pool.QueryRow(h.ctx, `SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='asr_drafts'::regclass AND contype='p'`).Scan(&primaryKey); err != nil || primaryKey != "PRIMARY KEY (audio_track_id)" {
		t.Fatalf("PK=%q err=%v", primaryKey, err)
	}
	// Down keeps one draft per lesson (MS2a review w1): the default track's even when another
	// track's attempt is newer (lesson B), else the newest attempt (lesson C, no default draft).
	h.exec(t, `INSERT INTO asr_drafts(lesson_id,asset_id,generation,audio_track_id,track_language,state,segments,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,1,$3::text::uuid,'en','running','[]'::jsonb,$4,$4)`,
		recoverLesson2, recoverAsset2, initialTrackID(recoverAsset2), draftNow.Add(-time.Hour))
	h.insertTrack(t, recoverTrack4, recoverAsset3, "uploaded", "fr", "ready", false)
	h.insertTrack(t, recoverTrack5, recoverAsset3, "uploaded", "pl", "ready", false)
	for track, at := range map[string]time.Time{recoverTrack4: draftNow.Add(-time.Hour), recoverTrack5: draftNow} {
		h.exec(t, `INSERT INTO asr_drafts(lesson_id,asset_id,generation,audio_track_id,track_language,state,segments,created_at,updated_at)
			VALUES($1::text::uuid,$2::text::uuid,1,$3::text::uuid,'xx','running','[]'::jsonb,$4,$4)`,
			recoverLesson3, recoverAsset3, track, at)
	}
	// Round-trip on populated rows must restore source leaves and map back to the same tracks.
	if _, err := provider.DownTo(h.ctx, 30); err != nil {
		t.Fatalf("00031 down: %v", err)
	}
	leaves := map[string]string{}
	leafRows, err := h.pool.Query(h.ctx, `SELECT lesson_id::text, COALESCE(source_leaf,'') FROM asr_drafts`)
	if err != nil {
		t.Fatal(err)
	}
	for leafRows.Next() {
		var lesson, leaf string
		if err := leafRows.Scan(&lesson, &leaf); err != nil {
			t.Fatal(err)
		}
		leaves[lesson] = leaf
	}
	leafRows.Close()
	if want := map[string]string{recoverLesson1: "original", recoverLesson2: "original", recoverLesson3: "audio-" + recoverTrack5}; fmt.Sprint(leaves) != fmt.Sprint(want) {
		t.Fatalf("down survivors=%v; want default-first then newest %v", leaves, want)
	}
	if _, err := provider.UpTo(h.ctx, 31); err != nil {
		t.Fatalf("00031 re-up: %v", err)
	}
	assertMapped("re-up", map[string]string{
		recoverLesson1: initialTrackID(recoverAsset1) + "|en-us",
		recoverLesson2: initialTrackID(recoverAsset2) + "|en",
		recoverLesson3: recoverTrack5 + "|pl",
	}, 7)
	// Lesson B's surviving draft now hangs on its embedded track; deleting that track takes it.
	h.exec(t, `DELETE FROM audio_tracks WHERE id=$1::text::uuid`, initialTrackID(recoverAsset2))
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts`).Scan(&count); err != nil || count != 2 {
		t.Fatalf("draft count after track cascade=%d err=%v", count, err)
	}
	if got := h.lessonDraft(t, recoverLesson1, initialTrackID(recoverAsset1)); got.Draft.TrackLanguage != "en-us" {
		t.Fatalf("cascade removed wrong draft: %+v", got)
	}
}

func stringPointer(value string) *string { return &value }
