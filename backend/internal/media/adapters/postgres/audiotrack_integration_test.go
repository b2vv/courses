//go:build integration

package postgres

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"courses/internal/media/domain"
	"github.com/pressly/goose/v3"
)

const (
	mediaTrackOne   = "a2000000-0000-4000-8000-000000000001"
	mediaTrackTwo   = "a2000000-0000-4000-8000-000000000002"
	mediaTrackThree = "a2000000-0000-4000-8000-000000000003"
)

func (h *mediaHarness) insertTrack(t *testing.T, id, assetID, kind, language, label string, isDefault bool, state domain.State, createdAt time.Time) {
	t.Helper()
	leaf := domain.EmbeddedAudioLeaf
	if kind == string(domain.UploadedTrack) {
		leaf = domain.AudioLeafPrefix + id
	}
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(
		id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5,$6,$7,$8,$9,$9)`,
		id, assetID, kind, language, label, isDefault, string(state), leaf, createdAt); err != nil {
		t.Fatal(err)
	}
}

func (h *mediaHarness) defaultTrackID(t *testing.T, assetID string) (string, int) {
	t.Helper()
	tracks, err := h.repository.ListAudioTracks(h.ctx, assetID)
	if err != nil {
		t.Fatal(err)
	}
	defaults := 0
	id := ""
	for _, track := range tracks {
		if track.IsDefault {
			defaults++
			id = track.ID
		}
	}
	return id, defaults
}

func (h *mediaHarness) testAudioTracksPersistAndList(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Українська", false, domain.Ready, mediaNow.Add(time.Minute))
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)

	tracks, err := h.repository.ListAudioTracks(h.ctx, mediaAssetOne)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 2 || tracks[0].ID != mediaTrackOne || tracks[1].ID != mediaTrackTwo {
		t.Fatalf("ordered tracks=%+v", tracks)
	}
	if tracks[0].Kind != domain.EmbeddedTrack || tracks[0].SourceLeaf != domain.EmbeddedAudioLeaf || !tracks[0].IsDefault {
		t.Fatalf("embedded track mapped wrong: %+v", tracks[0])
	}
	if tracks[1].Kind != domain.UploadedTrack || tracks[1].SourceLeaf != domain.AudioLeafPrefix+mediaTrackTwo || tracks[1].Language != "uk" {
		t.Fatalf("uploaded track mapped wrong: %+v", tracks[1])
	}
}

func (h *mediaHarness) testAudioTrackSchemaGuards(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)

	rejects := []struct {
		name string
		exec func() error
	}{
		{name: "embedded must be ready", exec: func() error {
			_, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
				VALUES($1::text::uuid,$2::text::uuid,'embedded','en','Original',false,'processing','original',$3,$3)`, mediaTrackOne, mediaAssetOne, mediaNow)
			return err
		}},
		{name: "uploaded leaf must match id", exec: func() error {
			_, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
				VALUES($1::text::uuid,$2::text::uuid,'uploaded','uk','Ukr',false,'created','original',$3,$3)`, mediaTrackOne, mediaAssetOne, mediaNow)
			return err
		}},
		{name: "default must be ready", exec: func() error {
			_, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
				VALUES($1::text::uuid,$2::text::uuid,'uploaded','uk','Ukr',true,'processing',$3,$4,$4)`, mediaTrackOne, mediaAssetOne, domain.AudioLeafPrefix+mediaTrackOne, mediaNow)
			return err
		}},
		{name: "awaiting_audio accepted on video_assets", exec: func() error {
			// Inverse guard: the widened state CHECK must ACCEPT the new state.
			_, err := h.pool.Exec(h.ctx, `UPDATE video_assets SET state='awaiting_audio', updated_at=$2 WHERE id=$1::text::uuid`, mediaAssetOne, mediaNow.Add(time.Minute))
			if err != nil {
				return err
			}
			// Restore so later asserts see a ready asset if reused.
			_, err = h.pool.Exec(h.ctx, `UPDATE video_assets SET state='ready', updated_at=$2 WHERE id=$1::text::uuid`, mediaAssetOne, mediaNow.Add(2*time.Minute))
			return err
		}},
	}
	for _, reject := range rejects {
		t.Run(reject.name, func(t *testing.T) {
			err := reject.exec()
			if reject.name == "awaiting_audio accepted on video_assets" {
				if err != nil {
					t.Fatalf("widened CHECK rejected awaiting_audio: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected a constraint violation, got nil")
			}
		})
	}

	// Two defaults for one asset: the partial unique index rejects the second.
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	_, err := h.pool.Exec(h.ctx, `INSERT INTO audio_tracks(id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,'uploaded','uk','Ukr',true,'ready',$3,$4,$4)`, mediaTrackTwo, mediaAssetOne, domain.AudioLeafPrefix+mediaTrackTwo, mediaNow.Add(time.Minute))
	if err == nil {
		t.Fatal("second default must violate the one-default-per-asset index")
	}
}

func (h *mediaHarness) testSetDefaultAudioTrack(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))

	if err := h.repository.SetDefaultAudioTrack(h.ctx, mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackTwo || defaults != 1 {
		t.Fatalf("after swap default=%s count=%d", id, defaults)
	}
	if err := h.repository.SetDefaultAudioTrack(h.ctx, mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour)); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("missing target: %v", err)
	}
	h.insertTrack(t, mediaTrackThree, mediaAssetOne, string(domain.UploadedTrack), "de", "Ger", false, domain.Processing, mediaNow.Add(2*time.Minute))
	if err := h.repository.SetDefaultAudioTrack(h.ctx, mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour)); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("non-ready target: %v", err)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackTwo || defaults != 1 {
		t.Fatalf("default changed after failed swaps: default=%s count=%d", id, defaults)
	}
}

func (h *mediaHarness) testSetDefaultAdverseOrder(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	// The target was created BEFORE the current default, so a naive single-statement
	// swap would flip the earlier row to default while the later row is still default
	// — the rank-swap that a non-deferrable partial unique index rejects by scan order.
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow.Add(time.Minute))
	if err := h.repository.SetDefaultAudioTrack(h.ctx, mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Hour)); err != nil {
		t.Fatalf("adverse-order swap: %v", err)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackOne || defaults != 1 {
		t.Fatalf("after adverse swap default=%s count=%d", id, defaults)
	}
}

func (h *mediaHarness) testConcurrentSetDefault(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))

	// Two swaps to different targets race on the same FOR UPDATE lock; they must
	// serialize, both succeed, and leave exactly one default — never two or zero.
	targets := []string{mediaTrackOne, mediaTrackTwo}
	errs := make([]error, len(targets))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range targets {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			errs[i] = h.repository.SetDefaultAudioTrack(h.ctx, mediaAssetOne, targets[i], mediaNow.Add(time.Hour))
		}(i)
	}
	close(start)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent swap %d: %v", i, err)
		}
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); defaults != 1 || (id != mediaTrackOne && id != mediaTrackTwo) {
		t.Fatalf("after concurrent swaps default=%s count=%d", id, defaults)
	}
}

func (h *mediaHarness) testDeleteAudioTrackReElection(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))
	h.insertTrack(t, mediaTrackThree, mediaAssetOne, string(domain.UploadedTrack), "de", "Ger", false, domain.Failed, mediaNow.Add(2*time.Minute))

	// Deleting the default re-elects the earliest remaining Ready track (trackTwo).
	if err := h.repository.DeleteAudioTrack(h.ctx, mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackTwo || defaults != 1 {
		t.Fatalf("re-elected default=%s count=%d", id, defaults)
	}
	// trackTwo is now the only Ready track; a Ready asset refuses to drop it.
	if err := h.repository.DeleteAudioTrack(h.ctx, mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour)); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("dropping last ready track: %v", err)
	}
	if err := h.repository.DeleteAudioTrack(h.ctx, mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour)); err != nil {
		t.Fatalf("deleting a non-default failed track: %v", err)
	}
	if err := h.repository.DeleteAudioTrack(h.ctx, mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour)); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("deleting a missing track: %v", err)
	}

	// When the asset is not Ready, dropping the last Ready track is allowed.
	h.insertAsset(t, mediaAssetTwo, mediaLessonTwo, 1, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackThree, mediaAssetTwo, string(domain.UploadedTrack), "uk", "Ukr", true, domain.Ready, mediaNow)
	if err := h.repository.DeleteAudioTrack(h.ctx, mediaAssetTwo, mediaTrackThree, mediaNow.Add(time.Hour)); err != nil {
		t.Fatalf("awaiting-audio asset may drop last ready: %v", err)
	}
}

func (h *mediaHarness) testCreateAudioTrack(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)

	track, err := domain.NewUploadedTrack(domain.AudioTrack{
		ID:           mediaTrackOne,
		VideoAssetID: mediaAssetOne,
		Kind:         domain.UploadedTrack,
		Language:     "uk",
		Label:        "Українська",
		State:        domain.Created,
		SourceLeaf:   domain.AudioLeafPrefix + mediaTrackOne,
		CreatedAt:    mediaNow,
		UpdatedAt:    mediaNow,
	})
	if err != nil {
		t.Fatal(err)
	}
	stored, err := h.repository.CreateAudioTrack(h.ctx, track)
	if err != nil {
		t.Fatal(err)
	}
	if stored.ID != mediaTrackOne || stored.State != domain.Created || stored.SourceLeaf != domain.AudioLeafPrefix+mediaTrackOne {
		t.Fatalf("stored=%+v", stored)
	}
	if _, defaults := h.defaultTrackID(t, mediaAssetOne); defaults != 0 {
		t.Fatalf("a fresh uploaded track must not be default, count=%d", defaults)
	}

	orphan := track
	orphan.ID = mediaTrackTwo
	orphan.VideoAssetID = mediaAssetThree // no such asset
	orphan.SourceLeaf = domain.AudioLeafPrefix + mediaTrackTwo
	if _, err := h.repository.CreateAudioTrack(h.ctx, orphan); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("orphan track: %v", err)
	}
}

func (h *mediaHarness) testMultiAudioMigrationReversible(t *testing.T) {
	database, err := sql.Open("pgx", h.connection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.DownTo(h.ctx, 24); err != nil {
		t.Fatal(err)
	}
	var present *string
	if err := h.pool.QueryRow(h.ctx, `SELECT to_regclass('audio_tracks')::text`).Scan(&present); err != nil {
		t.Fatal(err)
	}
	if present != nil {
		t.Fatalf("down migration left audio_tracks present: %v", *present)
	}
	if _, err := provider.Up(h.ctx); err != nil {
		t.Fatal(err)
	}
	if err := h.pool.QueryRow(h.ctx, `SELECT to_regclass('audio_tracks')::text`).Scan(&present); err != nil {
		t.Fatal(err)
	}
	if present == nil {
		t.Fatal("up migration did not restore audio_tracks")
	}
}
