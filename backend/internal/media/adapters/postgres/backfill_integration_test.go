//go:build integration

package postgres

import (
	"context"
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"courses/internal/media/domain"
	"github.com/pressly/goose/v3"
)

const (
	backfillVersion  = 28
	backfillPrevious = 27
)

func backfillAsset(n int) string { return "83000000-0000-4000-8000-0000000000" + twoDigits(n) }
func backfillTrack(n int) string { return "a3000000-0000-4000-8000-0000000000" + twoDigits(n) }

func twoDigits(n int) string {
	return string([]byte{byte('0' + n/10), byte('0' + n%10)})
}

// backfillTrackID mirrors md5('multi-audio-backfill:' || asset_id)::uuid (D-B2).
func backfillTrackID(assetID string) string {
	sum := md5.Sum([]byte("multi-audio-backfill:" + assetID))
	value := hex.EncodeToString(sum[:])
	return value[0:8] + "-" + value[8:12] + "-" + value[12:16] + "-" + value[16:20] + "-" + value[20:32]
}

type backfillRow struct {
	id, kind, language, label, state, leaf string
	isDefault                              bool
}

func (h *mediaHarness) migrationProvider(t *testing.T) *goose.Provider {
	t.Helper()
	database, err := sql.Open("pgx", h.connection)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	return provider
}

// atPreviousVersion resets the fixture, steps the schema down to just before the
// backfill and restores the head schema (on empty data) when the subtest ends.
func (h *mediaHarness) atPreviousVersion(t *testing.T) *goose.Provider {
	t.Helper()
	provider := h.migrationProvider(t)
	h.reset(t)
	if _, err := provider.DownTo(h.ctx, backfillPrevious); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		h.reset(t)
		if _, err := provider.Up(context.Background()); err != nil {
			t.Errorf("restore head schema: %v", err)
		}
	})
	return provider
}

func (h *mediaHarness) tracksOf(t *testing.T, assetID string) []backfillRow {
	t.Helper()
	rows, err := h.pool.Query(h.ctx, `SELECT id::text,kind,language,label,state,source_leaf,is_default
		FROM audio_tracks WHERE video_asset_id=$1::text::uuid ORDER BY kind, id`, assetID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var tracks []backfillRow
	for rows.Next() {
		var row backfillRow
		if err := rows.Scan(&row.id, &row.kind, &row.language, &row.label, &row.state, &row.leaf, &row.isDefault); err != nil {
			t.Fatal(err)
		}
		tracks = append(tracks, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return tracks
}

func (h *mediaHarness) schemaVersion(t *testing.T, provider *goose.Provider) int64 {
	t.Helper()
	version, err := provider.GetDBVersion(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	return version
}

func (h *mediaHarness) testEmbeddedBackfillPopulated(t *testing.T) {
	provider := h.atPreviousVersion(t)
	evictedAt := mediaNow.Add(-time.Minute)
	type fixture struct {
		name          string
		asset         int
		lesson        string
		generation    int
		state         domain.State
		evicted       *time.Time
		backfilled    bool
		language      string
		existingCount int
	}
	fixtures := []fixture{
		{name: "A ready, no tracks", asset: 1, lesson: mediaLessonOne, generation: 1, state: domain.Ready, backfilled: true, language: "und"},
		{name: "A2 ready, unfinished upload", asset: 2, lesson: mediaLessonOne, generation: 2, state: domain.Ready, backfilled: true, language: "und", existingCount: 1},
		{name: "A3 ready, failed upload", asset: 3, lesson: mediaLessonOne, generation: 3, state: domain.Ready, backfilled: true, language: "und", existingCount: 1},
		{name: "B ready but evicted", asset: 4, lesson: mediaLessonOne, generation: 4, state: domain.Ready, evicted: &evictedAt},
		{name: "C ready with a Publish embedded track", asset: 5, lesson: mediaLessonOne, generation: 5, state: domain.Ready, existingCount: 1},
		{name: "D video-only promoted by an upload", asset: 6, lesson: mediaLessonOne, generation: 6, state: domain.Ready, existingCount: 1},
		{name: "E processing", asset: 7, lesson: mediaLessonTwo, generation: 1, state: domain.Processing},
		{name: "F failed", asset: 8, lesson: mediaLessonTwo, generation: 2, state: domain.Failed},
		{name: "G awaiting audio", asset: 9, lesson: mediaLessonTwo, generation: 3, state: domain.AwaitingAudio, existingCount: 1},
		{name: "H ready, queued upload holds und", asset: 10, lesson: mediaLessonTwo, generation: 4, state: domain.Ready, backfilled: true, language: "und-x-original", existingCount: 1},
	}
	for _, item := range fixtures {
		h.insertAsset(t, backfillAsset(item.asset), item.lesson, item.generation, item.state, mediaNow.Add(time.Hour), item.evicted)
	}
	uploaded := string(domain.UploadedTrack)
	h.insertTrack(t, backfillTrack(2), backfillAsset(2), uploaded, "en", "English", false, domain.Created, mediaNow)
	h.insertTrack(t, backfillTrack(3), backfillAsset(3), uploaded, "en", "English", false, domain.Failed, mediaNow)
	h.insertTrack(t, backfillTrack(5), backfillAsset(5), string(domain.EmbeddedTrack), "und", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, backfillTrack(6), backfillAsset(6), uploaded, "uk", "Українська", true, domain.Ready, mediaNow)
	h.insertTrack(t, backfillTrack(9), backfillAsset(9), uploaded, "uk", "Українська", false, domain.Processing, mediaNow)
	h.insertTrack(t, backfillTrack(10), backfillAsset(10), uploaded, "und", "Невідома", false, domain.Queued, mediaNow)
	before := map[int][]backfillRow{}
	for _, item := range fixtures {
		before[item.asset] = h.tracksOf(t, backfillAsset(item.asset))
	}

	assertBackfilled := func(stage string) {
		t.Helper()
		for _, item := range fixtures {
			tracks := h.tracksOf(t, backfillAsset(item.asset))
			if !item.backfilled {
				if len(tracks) != len(before[item.asset]) {
					t.Errorf("%s %s: tracks=%+v, want untouched %+v", stage, item.name, tracks, before[item.asset])
				}
				continue
			}
			if len(tracks) != item.existingCount+1 {
				t.Errorf("%s %s: tracks=%+v, want %d existing + 1 embedded", stage, item.name, tracks, item.existingCount)
				continue
			}
			want := backfillRow{id: backfillTrackID(backfillAsset(item.asset)), kind: "embedded", language: item.language,
				label: "Original", state: "ready", leaf: "original", isDefault: true}
			if tracks[0] != want {
				t.Errorf("%s %s: embedded=%+v, want %+v", stage, item.name, tracks[0], want)
			}
			for _, existing := range tracks[1:] {
				if existing.isDefault || existing.kind != "uploaded" {
					t.Errorf("%s %s: existing upload changed: %+v", stage, item.name, existing)
				}
			}
		}
	}

	if _, err := provider.UpTo(h.ctx, backfillVersion); err != nil {
		t.Fatal(err)
	}
	assertBackfilled("up")

	// Reversibility (plan: not idempotence): Down removes exactly the backfilled rows.
	if _, err := provider.DownTo(h.ctx, backfillPrevious); err != nil {
		t.Fatal(err)
	}
	for _, item := range fixtures {
		if tracks := h.tracksOf(t, backfillAsset(item.asset)); len(tracks) != len(before[item.asset]) {
			t.Errorf("down %s: tracks=%+v, want the pre-backfill %+v", item.name, tracks, before[item.asset])
		}
	}
	if _, err := provider.UpTo(h.ctx, backfillVersion); err != nil {
		t.Fatal(err)
	}
	assertBackfilled("re-up")

	// Consequence 1 of the plan: an upload finishing on a backfilled asset no longer
	// rewrites the master without the original sound; the embedded track stays default.
	if _, err := h.pool.Exec(h.ctx, `UPDATE audio_tracks SET state='processing' WHERE id=$1::text::uuid`, backfillTrack(2)); err != nil {
		t.Fatal(err)
	}
	var rewritten []domain.AudioTrack
	if _, err := h.repository.PromoteAudioTrack(h.ctx, backfillAsset(2), backfillTrack(2), mediaNow.Add(time.Minute), func(ready []domain.AudioTrack) error {
		rewritten = ready
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	embeddedDefault := false
	for _, track := range rewritten {
		if track.Kind == domain.EmbeddedTrack && track.IsDefault {
			embeddedDefault = true
		}
	}
	if len(rewritten) != 2 || !embeddedDefault {
		t.Fatalf("promote on a backfilled asset rewrote %+v, want 2 renditions with the embedded default", rewritten)
	}
}

func (h *mediaHarness) testEmbeddedBackfillRefusesBothLanguagesTaken(t *testing.T) {
	provider := h.atPreviousVersion(t)
	h.insertAsset(t, backfillAsset(1), mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertAsset(t, backfillAsset(2), mediaLessonOne, 2, domain.Ready, mediaNow.Add(time.Hour), nil)
	uploaded := string(domain.UploadedTrack)
	h.insertTrack(t, backfillTrack(1), backfillAsset(1), uploaded, "und", "Перша", false, domain.Queued, mediaNow)
	h.insertTrack(t, backfillTrack(2), backfillAsset(1), uploaded, "und-x-original", "Друга", false, domain.Processing, mediaNow)
	_, err := provider.UpTo(h.ctx, backfillVersion)
	if err == nil || !strings.Contains(err.Error(), "already uses both und and und-x-original") {
		t.Fatalf("up error=%v, want the named language refusal", err)
	}
	if version := h.schemaVersion(t, provider); version != backfillPrevious {
		t.Fatalf("version=%d, want %d after the refusal", version, backfillPrevious)
	}
	if tracks := h.tracksOf(t, backfillAsset(2)); len(tracks) != 0 {
		t.Fatalf("a refused migration left rows: %+v", tracks)
	}
}

func (h *mediaHarness) testEmbeddedBackfillRefusesBrokenReadyGate(t *testing.T) {
	provider := h.atPreviousVersion(t)
	h.insertAsset(t, backfillAsset(1), mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertAsset(t, backfillAsset(2), mediaLessonOne, 2, domain.Ready, mediaNow.Add(time.Hour), nil)
	// Asset 1 has a Ready track but no default: the backfill skips it and the gate must refuse.
	h.insertTrack(t, backfillTrack(1), backfillAsset(1), string(domain.UploadedTrack), "en", "English", false, domain.Ready, mediaNow)
	_, err := provider.UpTo(h.ctx, backfillVersion)
	if err == nil || !strings.Contains(err.Error(), "has no ready default audio track") {
		t.Fatalf("up error=%v, want the Ready-gate refusal", err)
	}
	if version := h.schemaVersion(t, provider); version != backfillPrevious {
		t.Fatalf("version=%d, want %d after the refusal", version, backfillPrevious)
	}
	if tracks := h.tracksOf(t, backfillAsset(2)); len(tracks) != 0 {
		t.Fatalf("the candidate kept its backfill row after the rollback: %+v", tracks)
	}
}

func (h *mediaHarness) testEmbeddedBackfillWaitsForWriters(t *testing.T) {
	provider := h.atPreviousVersion(t)
	h.insertAsset(t, backfillAsset(1), mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertAsset(t, backfillAsset(2), mediaLessonOne, 2, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	writer, err := h.pool.Begin(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Rollback(context.Background())
	// An uncommitted write holds ROW EXCLUSIVE on audio_tracks, as a mint/promote would.
	if _, err := writer.Exec(h.ctx, `INSERT INTO audio_tracks(
		id,video_asset_id,kind,language,label,is_default,state,source_leaf,declared_bytes,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,'uploaded','en','English',false,'created',$3,1024,$4,$4)`,
		backfillTrack(1), backfillAsset(2), domain.AudioLeafPrefix+backfillTrack(1), mediaNow); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	_, err = provider.UpTo(h.ctx, backfillVersion)
	if err == nil || !strings.Contains(err.Error(), "lock timeout") {
		t.Fatalf("up error=%v, want a lock timeout while a writer holds audio_tracks", err)
	}
	if waited := time.Since(started); waited < 4*time.Second {
		t.Fatalf("the migration gave up after %s, want it to wait for the 5s lock_timeout", waited)
	}
	if version := h.schemaVersion(t, provider); version != backfillPrevious {
		t.Fatalf("version=%d, want %d after the lock timeout", version, backfillPrevious)
	}
	if err := writer.Rollback(h.ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := provider.UpTo(h.ctx, backfillVersion); err != nil {
		t.Fatalf("up after the writer finished: %v", err)
	}
	if tracks := h.tracksOf(t, backfillAsset(1)); len(tracks) != 1 || tracks[0].id != backfillTrackID(backfillAsset(1)) {
		t.Fatalf("backfill after the writer finished=%+v", tracks)
	}
}
