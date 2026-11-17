//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pressly/goose/v3"
)

func (h *draftHarness) testDraftTrackSchemaConstraints(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertRunningDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	for _, tc := range []struct{ name, query, code string }{
		{"unique source leaf", `INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,declared_bytes,created_at,updated_at) SELECT 'a6000000-0000-4000-8000-000000000002',video_asset_id,'embedded','fr','Duplicate',false,'ready','original',0,created_at,updated_at FROM audio_tracks WHERE id='a0000000-0000-4000-8000-000000000001'`, "23505"},
		{"language length", `UPDATE asr_drafts SET track_language='en-abcdefgh-abcdefgh-abcdefgh-abcdefgh'`, "23514"},
		{"language shape", `UPDATE asr_drafts SET track_language='en..us'`, "23514"},
		{"language required", `UPDATE asr_drafts SET track_language=NULL`, "23502"},
		{"track mode required", `UPDATE audio_tracks SET multilingual=NULL`, "23502"},
		{"snapshot mode required", `UPDATE asr_drafts SET track_multilingual=NULL`, "23502"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.pool.Exec(h.ctx, tc.query)
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != tc.code {
				t.Fatalf("%s: error=%v; want SQLSTATE %s", tc.name, err, tc.code)
			}
		})
	}
	var columns string
	err := h.pool.QueryRow(h.ctx, `SELECT string_agg(a.attname,',' ORDER BY k.ordinality) FROM pg_index i JOIN pg_class c ON c.oid=i.indexrelid CROSS JOIN LATERAL unnest(i.indkey) WITH ORDINALITY k(attnum,ordinality) JOIN pg_attribute a ON a.attrelid=i.indrelid AND a.attnum=k.attnum WHERE c.oid='asr_drafts_lesson'::regclass`).Scan(&columns)
	if err != nil || columns != "lesson_id" {
		t.Fatalf("asr_drafts_lesson columns=%q err=%v; want lesson_id", columns, err)
	}
}

func (h *draftHarness) testDraftTrackRollbackAcrossAssetsAndTiedTracks(t *testing.T, provider *goose.Provider) {
	h.reset(t)
	t.Cleanup(func() {
		if _, err := provider.Up(context.Background()); err != nil {
			t.Errorf("restore head: %v", err)
		}
	})
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	h.insertTrack(t, recoverTrack3, recoverAsset2, "uploaded", "fr", "ready", false)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	for _, row := range []struct{ asset, track, language string }{{recoverAsset2, recoverTrack3, "fr"}, {recoverAsset1, recoverTrack2, "de"}} {
		h.exec(t, `INSERT INTO asr_drafts(lesson_id,asset_id,generation,audio_track_id,track_language,state,segments,created_at,updated_at) VALUES($1::text::uuid,$2::text::uuid,1,$3::text::uuid,$4,'running','[]'::jsonb,$5,$5)`, recoverLesson1, row.asset, row.track, row.language, draftNow)
	}
	if _, err := provider.DownTo(h.ctx, 30); err != nil {
		t.Fatalf("rollback two assets of one lesson: %v", err)
	}
	var count int
	var leaf string
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*),min(source_leaf) FROM asr_drafts`).Scan(&count, &leaf); err != nil || count != 1 || leaf != "audio-"+recoverTrack2 {
		t.Fatalf("rollback survivors=%d leaf=%q err=%v; want one ascending-track winner %s", count, leaf, err, recoverTrack2)
	}
	if _, err := provider.Up(h.ctx); err != nil {
		t.Fatalf("round-trip up: %v", err)
	}
	var track string
	if err := h.pool.QueryRow(h.ctx, `SELECT audio_track_id::text FROM asr_drafts`).Scan(&track); err != nil || track != recoverTrack2 {
		t.Fatalf("round-trip track=%q err=%v; want %s", track, err, recoverTrack2)
	}
}

func (h *draftHarness) testDraftTrackCurrentAssetAndMetadata(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	for _, track := range []string{initialTrackID(recoverAsset1), recoverTrack2} {
		claim, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, track, false, draftNow)
		if err != nil || !claim.Claimed {
			t.Fatalf("claim=%+v err=%v", claim, err)
		}
		h.completeDraftText(t, claim.Draft, claim.Token, "recognized", draftNow.Add(123456*time.Microsecond))
	}
	embedded := h.lessonDraft(t, recoverLesson1, initialTrackID(recoverAsset1))
	uploaded := h.lessonDraft(t, recoverLesson1, recoverTrack2)
	if !embedded.IsEmbedded || uploaded.IsEmbedded {
		t.Fatalf("embedded flags original=%v uploaded=%v; want true/false", embedded.IsEmbedded, uploaded.IsEmbedded)
	}
	if !embedded.IsDefault || uploaded.IsDefault {
		t.Fatalf("initial default flags original=%v uploaded=%v; want true/false", embedded.IsDefault, uploaded.IsDefault)
	}
	if embedded.UpdatedAt != "2026-09-22T12:00:00.123456Z" {
		t.Fatalf("updatedAt=%q; want 2026-09-22T12:00:00.123456Z", embedded.UpdatedAt)
	}
	h.exec(t, `UPDATE audio_tracks SET is_default=false WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	h.exec(t, `UPDATE audio_tracks SET is_default=true WHERE id=$1::text::uuid`, recoverTrack2)
	if h.lessonDraft(t, recoverLesson1, initialTrackID(recoverAsset1)).IsDefault || !h.lessonDraft(t, recoverLesson1, recoverTrack2).IsDefault {
		t.Fatal("default flags must follow the current selection")
	}
	// With neither track default, the real repository's embedded flag controls editor ordering.
	h.exec(t, `UPDATE audio_tracks SET is_default=false WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	inventory, err := application.NewDrafting(h.repository, h.repository, nil, time.Now, nil).ListLessonDrafts(h.ctx, domain.Actor{UserID: draftInstructor}, recoverLesson1)
	if err != nil || len(inventory) != 2 || inventory[0].Draft.TrackID != initialTrackID(recoverAsset1) {
		t.Fatalf("editor order=%+v err=%v; want embedded first", inventory, err)
	}
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	h.exec(t, `UPDATE video_assets SET evicted_at=$1 WHERE id=$2::text::uuid`, draftNow, recoverAsset2)
	old, err := h.repository.LessonTrack(h.ctx, recoverLesson1, initialTrackID(recoverAsset1))
	if err != nil || old == nil || old.AssetID != recoverAsset1 {
		t.Fatalf("non-evicted fallback=%+v err=%v; want older ready asset", old, err)
	}
	evicted, err := h.repository.LessonTrack(h.ctx, recoverLesson1, initialTrackID(recoverAsset2))
	if err != nil || evicted != nil {
		t.Fatalf("evicted selection=%+v err=%v; want absent", evicted, err)
	}
	listed, err := h.repository.ListLessonDrafts(h.ctx, recoverLesson1)
	if err != nil || len(listed) != 2 {
		t.Fatalf("older non-evicted inventory=%+v err=%v; want two drafts", listed, err)
	}
}

func (h *draftHarness) testDraftTrackPendingEligibilityOrderAndLimit(t *testing.T) {
	h.reset(t)
	h.insertOrphanAssetWithDuration(t, recoverAsset3, 60)
	h.insertTrack(t, recoverTrack4, recoverAsset3, "embedded", "en", "ready", true)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.exec(t, `UPDATE video_assets SET evicted_at=$1 WHERE id=$2::text::uuid`, draftNow, recoverAsset2)
	// Insert the higher UUID before the lower to make storage order disagree with queue order.
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.exec(t, `DELETE FROM audio_tracks WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	h.insertTrack(t, recoverTrack3, recoverAsset1, "uploaded", "fr", "ready", false)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	rows, err := h.repository.PendingRecognitions(h.ctx, 10)
	if err != nil {
		t.Fatalf("pending must exclude detached/evicted ready assets without scan error: %v", err)
	}
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.TrackID)
	}
	if !reflect.DeepEqual(ids, []string{recoverTrack2, recoverTrack3}) {
		t.Fatalf("pending order=%v; want [%s %s] only", ids, recoverTrack2, recoverTrack3)
	}
	rows, err = h.repository.PendingRecognitions(h.ctx, 1)
	if err != nil || len(rows) != 1 || rows[0].TrackID != recoverTrack2 {
		t.Fatalf("pending limit=1 rows=%+v err=%v; want first ascending track only", rows, err)
	}
}

func (h *draftHarness) testDraftTrackClaimAssetAndReset(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	claim, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset2, 1, initialTrackID(recoverAsset1), false, draftNow)
	if !errors.Is(err, domain.ErrNotFound) || claim.Claimed {
		t.Fatalf("mismatched asset claim=%+v err=%v; want not_found", claim, err)
	}
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts`).Scan(&count); err != nil || count != 0 {
		t.Fatalf("invalid claim wrote %d drafts err=%v", count, err)
	}
	claim, err = h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
	if err != nil || !claim.Claimed {
		t.Fatalf("valid claim=%+v err=%v", claim, err)
	}
	h.completeDraftText(t, claim.Draft, claim.Token, "ready", draftNow.Add(time.Second))
	claim, err = h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), true, draftNow.Add(2*time.Second))
	if err != nil || !claim.Claimed {
		t.Fatalf("ready re-claim=%+v err=%v", claim, err)
	}
	row := h.lessonDraft(t, recoverLesson1, claim.Draft.TrackID)
	if row.Draft.Language != "" || len(row.Draft.Segments) != 0 || row.Draft.State != domain.DraftRunning {
		t.Fatalf("ready reset=%+v; want running without recognized language or segments", row.Draft)
	}
	claim.Draft.State, claim.Draft.Failure = domain.DraftFailed, "engine_error"
	if err := h.repository.Complete(h.ctx, claim.Draft, claim.Token, draftNow.Add(3*time.Second)); err != nil {
		t.Fatal(err)
	}
	claim, err = h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), true, draftNow.Add(4*time.Second))
	if err != nil || !claim.Claimed {
		t.Fatalf("failed re-claim=%+v err=%v; want claimed", claim, err)
	}
	row = h.lessonDraft(t, recoverLesson1, claim.Draft.TrackID)
	if row.Draft.State != domain.DraftRunning || row.Draft.Failure != "" {
		t.Fatalf("failed reset=%+v; want running without failure", row.Draft)
	}
}

func (h *draftHarness) testDraftTrackRecoveryStateOrderAndLimit(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.exec(t, `DELETE FROM audio_tracks WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	for i, track := range []string{recoverTrack2, recoverTrack3, recoverTrack4, recoverTrack5, recoverTrack6} {
		h.insertTrack(t, track, recoverAsset1, "uploaded", []string{"de", "fr", "pl", "uk", "es"}[i], "ready", false)
		at := draftNow.Add(-time.Hour)
		if i == 1 {
			at = draftNow.Add(-2 * time.Hour)
		}
		claim, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, track, false, at)
		if err != nil || !claim.Claimed {
			t.Fatalf("seed=%+v err=%v", claim, err)
		}
		if i >= 2 {
			claim.Draft.State = domain.DraftNoSpeech
			if i == 2 {
				h.completeDraftText(t, claim.Draft, claim.Token, "terminal", draftNow.Add(-time.Hour))
				continue
			}
			if i == 3 {
				claim.Draft.State, claim.Draft.Failure = domain.DraftFailed, "engine_error"
			}
			if err := h.repository.Complete(h.ctx, claim.Draft, claim.Token, draftNow.Add(-time.Hour)); err != nil {
				t.Fatal(err)
			}
		}
	}
	rows, err := h.repository.Recover(h.ctx, draftNow, 10)
	if err != nil || len(rows) != 2 || rows[0].Draft.TrackID != recoverTrack3 || rows[1].Draft.TrackID != recoverTrack2 {
		t.Fatalf("recovery=%+v err=%v; want oldest running track then newer, no terminal rows", rows, err)
	}
	rows, err = h.repository.Recover(h.ctx, draftNow, 1)
	if err != nil || len(rows) != 1 || rows[0].Draft.TrackID != recoverTrack3 {
		t.Fatalf("recovery limit=1 rows=%+v err=%v; want oldest only", rows, err)
	}
	h.exec(t, `UPDATE asr_drafts SET updated_at=$1 WHERE audio_track_id=$2::text::uuid`, draftNow.Add(-time.Hour), recoverTrack3)
	rows, err = h.repository.Recover(h.ctx, draftNow, 1)
	if err != nil || len(rows) != 1 || rows[0].Draft.TrackID != recoverTrack2 {
		t.Fatalf("tied recovery limit=1 rows=%+v err=%v; want ascending track %s", rows, err, recoverTrack2)
	}
	// Completing that bounded batch leaves its sibling eligible on the next sweep.
	rows[0].Draft.State, rows[0].Draft.Failure = domain.DraftFailed, "timeout"
	if err := h.repository.Complete(h.ctx, rows[0].Draft, rows[0].Token, draftNow); err != nil {
		t.Fatal(err)
	}
	rows, err = h.repository.Recover(h.ctx, draftNow, 1)
	if err != nil || len(rows) != 1 || rows[0].Draft.TrackID != recoverTrack3 {
		t.Fatalf("next recovery batch=%+v err=%v; want remaining running sibling", rows, err)
	}
}

func (h *draftHarness) testDraftTrackClaimConcurrentTrackDeletion(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	ctx, cancel := context.WithTimeout(h.ctx, 10*time.Second)
	defer cancel()
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	var deletingPID int
	if err := tx.QueryRow(ctx, `SELECT pg_backend_pid()`).Scan(&deletingPID); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM audio_tracks WHERE id=$1::text::uuid`, initialTrackID(recoverAsset1)); err != nil {
		t.Fatal(err)
	}
	type result struct {
		claim ports.ClaimedDraft
		err   error
	}
	done := make(chan result, 1)
	go func() {
		claim, err := h.repository.Claim(ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
		done <- result{claim, err}
	}()
	// Observe the real INSERT waiting on the deleting transaction's FK lock before commit.
	for {
		var blocked bool
		err := h.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)) AND wait_event_type='Lock' AND wait_event='transactionid' AND query LIKE '%INSERT INTO asr_drafts%')`, deletingPID).Scan(&blocked)
		if err != nil {
			t.Fatalf("observe FK lock wait: %v", err)
		}
		if blocked {
			break
		}
		select {
		case got := <-done:
			t.Fatalf("claim returned before FK lock wait: %+v", got)
		default:
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-done:
		if !errors.Is(got.err, domain.ErrNotFound) || got.claim.Claimed {
			t.Fatalf("claim after committed track deletion=%+v err=%v; want not_found", got.claim, got.err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}
