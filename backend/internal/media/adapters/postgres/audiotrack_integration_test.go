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

// noMasterRewrite is the delete rewrite callback for tests that assert DB state
// only; L3 master regeneration is covered separately by testDeleteAudioTrackRegeneratesMaster.
func noMasterRewrite([]domain.AudioTrack) error { return nil }

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
	tracks, err := h.repository.ListAudioTracks(h.ctx, ownerActor(), assetID)
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

	tracks, err := h.repository.ListAudioTracks(h.ctx, ownerActor(), mediaAssetOne)
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

	var rewrittenWith []domain.AudioTrack
	if err := h.repository.SetDefaultAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour), func(ready []domain.AudioTrack) error {
		rewrittenWith = ready
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackTwo || defaults != 1 {
		t.Fatalf("after swap default=%s count=%d", id, defaults)
	}
	// The rewrite must see the NEW default (trackTwo), not the old one — it reads
	// inside the transaction after the swap, so the master carries the correct
	// DEFAULT=YES line (D-S5L1-5b).
	found := false
	for _, track := range rewrittenWith {
		if track.ID == mediaTrackTwo && track.IsDefault {
			found = true
		}
		if track.ID == mediaTrackOne && track.IsDefault {
			t.Fatalf("rewrite still sees trackOne as default: %+v", rewrittenWith)
		}
	}
	if !found {
		t.Fatalf("rewrite did not see trackTwo as default: %+v", rewrittenWith)
	}
	if err := h.repository.SetDefaultAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("missing target: %v", err)
	}
	h.insertTrack(t, mediaTrackThree, mediaAssetOne, string(domain.UploadedTrack), "de", "Ger", false, domain.Processing, mediaNow.Add(2*time.Minute))
	if err := h.repository.SetDefaultAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrInvalidRequest) {
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
	if err := h.repository.SetDefaultAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Hour), noMasterRewrite); err != nil {
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
			errs[i] = h.repository.SetDefaultAudioTrack(h.ctx, ownerActor(), mediaAssetOne, targets[i], mediaNow.Add(time.Hour), noMasterRewrite)
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

func (h *mediaHarness) testPromoteAudioTrackReadiesVideoOnlyAsset(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Queued, mediaNow)

	var rewritten []domain.AudioTrack
	if _, err := h.repository.PromoteAudioTrack(h.ctx, mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Minute), func(ready []domain.AudioTrack) error {
		rewritten = ready
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	asset, err := h.repository.Load(h.ctx, mediaAssetOne)
	if err != nil || asset.State != domain.Ready {
		t.Fatalf("asset=%+v err=%v want ready (T11)", asset, err)
	}
	track, err := h.repository.LoadAudioTrack(h.ctx, mediaAssetOne, mediaTrackOne)
	if err != nil || track.State != domain.Ready || !track.IsDefault {
		t.Fatalf("track=%+v err=%v want ready default", track, err)
	}
	if len(rewritten) != 1 || rewritten[0].ID != mediaTrackOne || !rewritten[0].IsDefault {
		t.Fatalf("rewrite ready set=%+v", rewritten)
	}
}

func (h *mediaHarness) testPromoteAudioTrackKeepsEmbeddedDefault(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Processing, mediaNow.Add(time.Minute))

	var readyCount, defaults int
	if _, err := h.repository.PromoteAudioTrack(h.ctx, mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Minute), func(ready []domain.AudioTrack) error {
		readyCount = len(ready)
		for _, track := range ready {
			if track.IsDefault {
				defaults++
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if readyCount != 2 || defaults != 1 {
		t.Fatalf("ready=%d defaults=%d want 2 ready, 1 default", readyCount, defaults)
	}
	if id, count := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackOne || count != 1 {
		t.Fatalf("default=%s count=%d want the embedded track to stay default", id, count)
	}
}

func (h *mediaHarness) testPromoteAudioTrackEvictsOldestReady(t *testing.T) {
	h.reset(t)
	// Two older Ready generations already fill the lesson's cap; the video-only
	// generation now gaining its audio makes three, so promoting it must evict the
	// oldest exactly as the video publish path does (D4).
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertAsset(t, mediaAssetTwo, mediaLessonOne, 2, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertAsset(t, mediaAssetThree, mediaLessonOne, 3, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetThree, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Queued, mediaNow)

	evicted, err := h.repository.PromoteAudioTrack(h.ctx, mediaAssetThree, mediaTrackOne, mediaNow.Add(time.Minute), func([]domain.AudioTrack) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if evicted == nil || evicted.ID != mediaAssetOne {
		t.Fatalf("evicted=%+v want the oldest generation %s", evicted, mediaAssetOne)
	}
	for _, want := range []struct {
		id      string
		evicted bool
	}{
		{mediaAssetOne, true},
		{mediaAssetTwo, false},
		{mediaAssetThree, false},
	} {
		var isEvicted bool
		if err := h.pool.QueryRow(h.ctx, `SELECT evicted_at IS NOT NULL FROM video_assets WHERE id=$1`, want.id).Scan(&isEvicted); err != nil {
			t.Fatal(err)
		}
		if isEvicted != want.evicted {
			t.Fatalf("asset %s evicted=%v want %v", want.id, isEvicted, want.evicted)
		}
	}
	asset, err := h.repository.Load(h.ctx, mediaAssetThree)
	if err != nil || asset.State != domain.Ready {
		t.Fatalf("promoted asset=%+v err=%v want ready", asset, err)
	}
}

func (h *mediaHarness) testFailAudioTrackFailsVideoOnlyAsset(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Processing, mediaNow)

	if err := h.repository.FailAudioTrack(h.ctx, mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	track, err := h.repository.LoadAudioTrack(h.ctx, mediaAssetOne, mediaTrackOne)
	if err != nil || track.State != domain.Failed {
		t.Fatalf("track=%+v err=%v want failed", track, err)
	}
	asset, err := h.repository.Load(h.ctx, mediaAssetOne)
	if err != nil || asset.State != domain.Failed {
		t.Fatalf("asset=%+v err=%v want failed (T12)", asset, err)
	}
}

func (h *mediaHarness) testFailAudioTrackKeepsAssetWithOtherAudio(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Processing, mediaNow.Add(time.Minute))

	if err := h.repository.FailAudioTrack(h.ctx, mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	asset, err := h.repository.Load(h.ctx, mediaAssetOne)
	if err != nil || asset.State != domain.Ready {
		t.Fatalf("asset=%+v err=%v want still ready", asset, err)
	}
}

func (h *mediaHarness) testConcurrentPromoteAudioTrack(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Queued, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "de", "Ger", false, domain.Queued, mediaNow.Add(time.Minute))

	// Two audio transcodes finish at once; the per-asset lock serialises their
	// master rewrites and default election — the result is both tracks Ready with
	// exactly one default, never two (D7 single-writer).
	targets := []string{mediaTrackOne, mediaTrackTwo}
	errs := make([]error, len(targets))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := range targets {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, errs[i] = h.repository.PromoteAudioTrack(h.ctx, mediaAssetOne, targets[i], mediaNow.Add(time.Hour), func([]domain.AudioTrack) error {
				return nil
			})
		}(i)
	}
	close(start)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent promote %d: %v", i, err)
		}
	}
	tracks, err := h.repository.ListAudioTracks(h.ctx, ownerActor(), mediaAssetOne)
	if err != nil {
		t.Fatal(err)
	}
	ready, defaults := 0, 0
	for _, track := range tracks {
		if track.State == domain.Ready {
			ready++
		}
		if track.IsDefault {
			defaults++
		}
	}
	if ready != 2 || defaults != 1 {
		t.Fatalf("after concurrent promote ready=%d defaults=%d want 2, 1", ready, defaults)
	}
	if asset, err := h.repository.Load(h.ctx, mediaAssetOne); err != nil || asset.State != domain.Ready {
		t.Fatalf("asset=%+v err=%v want ready", asset, err)
	}
}

func (h *mediaHarness) testDeleteAudioTrackReElection(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))
	h.insertTrack(t, mediaTrackThree, mediaAssetOne, string(domain.UploadedTrack), "de", "Ger", false, domain.Failed, mediaNow.Add(2*time.Minute))

	// Deleting the default re-elects the earliest remaining Ready track (trackTwo).
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Hour), noMasterRewrite); err != nil {
		t.Fatal(err)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackTwo || defaults != 1 {
		t.Fatalf("re-elected default=%s count=%d", id, defaults)
	}
	// trackTwo is now the only Ready track; a Ready asset refuses to drop it.
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("dropping last ready track: %v", err)
	}
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour), noMasterRewrite); err != nil {
		t.Fatalf("deleting a non-default failed track: %v", err)
	}
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackThree, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("deleting a missing track: %v", err)
	}

	// When the asset is not Ready, dropping the last Ready track is allowed.
	h.insertAsset(t, mediaAssetTwo, mediaLessonTwo, 1, domain.AwaitingAudio, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackThree, mediaAssetTwo, string(domain.UploadedTrack), "uk", "Ukr", true, domain.Ready, mediaNow)
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetTwo, mediaTrackThree, mediaNow.Add(time.Hour), noMasterRewrite); err != nil {
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

func (h *mediaHarness) testAudioTrackAuthoringOwnership(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))

	// mediaInstructorB owns a different course, so every authoring call is refused
	// before it touches a track (D-S5L1-3).
	stranger := domain.Actor{UserID: mediaInstructorB}
	if _, err := h.repository.ListAudioTracks(h.ctx, stranger, mediaAssetOne); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("list by a non-owner: %v", err)
	}
	if err := h.repository.SetDefaultAudioTrack(h.ctx, stranger, mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("set-default by a non-owner: %v", err)
	}
	if _, err := h.repository.UpdateAudioTrackMetadata(h.ctx, stranger, mediaAssetOne, mediaTrackTwo, "de", "Ger", mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("rename by a non-owner: %v", err)
	}
	if err := h.repository.DeleteAudioTrack(h.ctx, stranger, mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("delete by a non-owner: %v", err)
	}
	// Nothing changed under the refused calls.
	if tracks, err := h.repository.ListAudioTracks(h.ctx, ownerActor(), mediaAssetOne); err != nil || len(tracks) != 2 {
		t.Fatalf("tracks=%+v err=%v want 2 untouched", tracks, err)
	}
}

func (h *mediaHarness) testUpdateAudioTrackMetadata(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))

	// The rename is rendered into the master (NAME/LANGUAGE), so the rewrite must see
	// the new values inside the transaction, before commit.
	var rewrittenWith []domain.AudioTrack
	updated, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "uk-UA", "Українська", mediaNow.Add(time.Hour), func(ready []domain.AudioTrack) error {
		rewrittenWith = ready
		return nil
	})
	if err != nil || updated.Language != "uk-UA" || updated.Label != "Українська" {
		t.Fatalf("renamed=%+v err=%v", updated, err)
	}
	if len(rewrittenWith) != 2 || rewrittenWith[1].ID != mediaTrackTwo || rewrittenWith[1].Label != "Українська" || rewrittenWith[1].Language != "uk-UA" {
		t.Fatalf("rename rewrote the master with %+v, want the renamed uk track", rewrittenWith)
	}
	// A failing master rewrite aborts the rename: the database keeps the old label.
	if _, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "uk", "Rolled back", mediaNow.Add(time.Hour), func([]domain.AudioTrack) error {
		return errors.New("disk full")
	}); err == nil {
		t.Fatal("a failing master rewrite must fail the rename")
	}
	if track, err := h.repository.LoadAudioTrack(h.ctx, mediaAssetOne, mediaTrackTwo); err != nil || track.Label != "Українська" {
		t.Fatalf("after aborted rename track=%+v err=%v, want the committed label kept", track, err)
	}
	// Renaming to the embedded track's language collides on the (asset, language)
	// unique index and surfaces as ErrDuplicateLanguage (409), not a generic 400.
	if _, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "en", "Clash", mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrDuplicateLanguage) {
		t.Fatalf("duplicate language: %v", err)
	}
	// A missing track under the asset is not found.
	if _, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackThree, "de", "Ger", mediaNow.Add(time.Hour), noMasterRewrite); !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("missing track: %v", err)
	}
}

func (h *mediaHarness) testDeleteAudioTrackRegeneratesMaster(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))

	// L3: deleting the uk track must regenerate the master from the remaining Ready
	// set (the embedded track only), under the lock, before commit.
	var rewrittenWith []domain.AudioTrack
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, mediaNow.Add(time.Hour), func(ready []domain.AudioTrack) error {
		rewrittenWith = ready
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(rewrittenWith) != 1 || rewrittenWith[0].ID != mediaTrackOne {
		t.Fatalf("master rewritten with %+v, want only the embedded track", rewrittenWith)
	}
	if tracks, err := h.repository.ListAudioTracks(h.ctx, ownerActor(), mediaAssetOne); err != nil || len(tracks) != 1 || tracks[0].ID != mediaTrackOne {
		t.Fatalf("after delete tracks=%+v err=%v", tracks, err)
	}
}

func (h *mediaHarness) testDeleteEmbeddedAudioTrack(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))

	// Deleting the default embedded track re-elects the remaining Ready uploaded
	// track and regenerates from it; the embedded track is not special-cased.
	var rewrittenWith []domain.AudioTrack
	if err := h.repository.DeleteAudioTrack(h.ctx, ownerActor(), mediaAssetOne, mediaTrackOne, mediaNow.Add(time.Hour), func(ready []domain.AudioTrack) error {
		rewrittenWith = ready
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(rewrittenWith) != 1 || rewrittenWith[0].ID != mediaTrackTwo || !rewrittenWith[0].IsDefault {
		t.Fatalf("after embedded delete rewritten=%+v, want uk re-elected default", rewrittenWith)
	}
	if id, defaults := h.defaultTrackID(t, mediaAssetOne); id != mediaTrackTwo || defaults != 1 {
		t.Fatalf("re-elected default=%s count=%d", id, defaults)
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
