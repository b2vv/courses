//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/media/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (h *mediaHarness) assertMultilingualMode(t *testing.T, track domain.AudioTrack, want bool) {
	t.Helper()
	var stored bool
	if err := h.pool.QueryRow(h.ctx, `SELECT multilingual FROM audio_tracks WHERE id=$1`, track.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	loaded, err := h.repository.LoadAudioTrack(h.ctx, track.VideoAssetID, track.ID)
	if err != nil {
		t.Fatal(err)
	}
	if track.Multilingual != want || stored != want || loaded.Multilingual != want {
		t.Fatalf("multilingual returned=%v SQL=%v loaded=%v want=%v", track.Multilingual, stored, loaded.Multilingual, want)
	}
}

func (h *mediaHarness) testMultilingualPersistence(t *testing.T) {
	for _, want := range []bool{false, true} {
		t.Run(map[bool]string{false: "false", true: "true"}[want], func(t *testing.T) {
			t.Run("create", func(t *testing.T) {
				h.reset(t)
				h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
				track, err := domain.NewUploadedTrack(domain.AudioTrack{ID: mediaTrackOne, VideoAssetID: mediaAssetOne, Kind: domain.UploadedTrack, Language: "uk", Label: "Ukr", State: domain.Created, SourceLeaf: domain.AudioLeafPrefix + mediaTrackOne, DeclaredBytes: 1024, OriginalName: "audio.aac", ExpiresAt: mediaNow.Add(time.Hour), CreatedAt: mediaNow, UpdatedAt: mediaNow, Multilingual: want})
				if err != nil {
					t.Fatal(err)
				}
				stored, err := h.repository.CreateAudioTrack(h.ctx, track)
				if err != nil {
					t.Fatal(err)
				}
				h.assertMultilingualMode(t, stored, want)
			})
			t.Run("mint", func(t *testing.T) {
				h.reset(t)
				h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
				stored, err := h.repository.MintAudioTrack(h.ctx, ownerActor(), mediaAssetOne, 1024, "audio.aac", "uk", "Ukr", want, mediaNow.Add(time.Hour), mediaNow)
				if err != nil {
					t.Fatal(err)
				}
				h.assertMultilingualMode(t, stored, want)
			})
			t.Run("publish", func(t *testing.T) {
				h.reset(t)
				h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Processing, mediaNow.Add(time.Hour), nil)
				track := embeddedTrackFor(mediaAssetOne)
				track.Multilingual = want
				if _, _, err := h.repository.Publish(h.ctx, mediaAssetOne, mediaClaimA, []domain.Rendition{domain.Rendition360p}, nil, mediaNow, track, func(domain.VideoAsset) error { return nil }); err != nil {
					t.Fatal(err)
				}
				tracks, err := h.repository.ListAudioTracks(h.ctx, ownerActor(), mediaAssetOne)
				if err != nil || len(tracks) != 1 {
					t.Fatalf("published tracks=%+v err=%v", tracks, err)
				}
				h.assertMultilingualMode(t, tracks[0], want)
			})
		})
	}
}

func (h *mediaHarness) testTrackMetadata(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertTrack(t, mediaTrackTwo, mediaAssetOne, string(domain.UploadedTrack), "uk", "Ukr", false, domain.Ready, mediaNow.Add(time.Minute))
	calls := 0
	for _, track := range []struct{ id, language, label string }{{mediaTrackOne, "en", "Original"}, {mediaTrackTwo, "uk", "Ukr"}} {
		_, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, track.id, track.language, track.label, nil, mediaNow.Add(time.Hour), func([]domain.AudioTrack) error { calls++; return nil })
		if err != nil {
			t.Fatal(err)
		}
	}
	if calls != 0 {
		t.Fatalf("metadata no-ops rewrote master %d times; want 0", calls)
	}
	_, err := h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "uk", "New label", nil, mediaNow.Add(time.Hour), func(tracks []domain.AudioTrack) error {
		calls++
		if len(tracks) != 2 || tracks[1].ID != mediaTrackTwo || tracks[1].Label != "New label" {
			t.Fatalf("label-only rewrite saw %+v", tracks)
		}
		return nil
	})
	if err != nil || calls != 1 {
		t.Fatalf("label-only rewrite calls=%d err=%v; want 1 and nil", calls, err)
	}
	diskFull := errors.New("disk full")
	_, err = h.repository.UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "uk", "Rejected label", nil, mediaNow.Add(time.Hour), func([]domain.AudioTrack) error { return diskFull })
	if !errors.Is(err, diskFull) {
		t.Fatalf("label-only rewrite error=%v want disk full", err)
	}
	track, err := h.repository.LoadAudioTrack(h.ctx, mediaAssetOne, mediaTrackTwo)
	if err != nil || track.Label != "New label" {
		t.Fatalf("rollback label=%q err=%v want New label", track.Label, err)
	}
}

func (h *mediaHarness) lockTrackRow(t *testing.T, id string) pgx.Tx {
	t.Helper()
	tx, err := h.pool.Begin(h.ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	if _, err := tx.Exec(h.ctx, `SELECT id FROM audio_tracks WHERE id=$1 FOR UPDATE`, id); err != nil {
		t.Fatal(err)
	}
	return tx
}

func (h *mediaHarness) testTrackMetadataForeignLock(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertAsset(t, mediaAssetTwo, mediaLessonTwo, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetTwo, string(domain.EmbeddedTrack), "en", "Foreign", true, domain.Ready, mediaNow)
	h.lockTrackRow(t, mediaTrackOne) // completed SELECT establishes the foreign row lock before the call
	ctx, cancel := context.WithTimeout(h.ctx, 2*time.Second)
	defer cancel()
	_, err := h.repository.UpdateAudioTrackMetadata(ctx, ownerActor(), mediaAssetOne, mediaTrackOne, "en", "Foreign", nil, mediaNow, noMasterRewrite)
	if !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("foreign locked track error=%v want video_not_found", err)
	}
}

// Observe PostgreSQL's actual contested lock, rather than inferring scheduling from elapsed time.
func (h *mediaHarness) waitForTrackLock(ctx context.Context, fragment string, done <-chan error) (bool, error) {
	for ctx.Err() == nil {
		select {
		case err := <-done:
			return false, err
		default:
		}
		var blocked bool
		err := h.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE datname=current_database() AND wait_event_type='Lock' AND position($1 in query)>0)`, fragment).Scan(&blocked)
		if err != nil {
			return false, err
		}
		if blocked {
			return true, nil
		}
	}
	return false, ctx.Err()
}

func (h *mediaHarness) testTrackMetadataLockQueryFailure(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	tx := h.lockTrackRow(t, mediaTrackOne)
	ctx, cancel := context.WithCancel(h.ctx)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := h.repository.UpdateAudioTrackMetadata(ctx, ownerActor(), mediaAssetOne, mediaTrackOne, "en", "Original", nil, mediaNow, noMasterRewrite)
		done <- err
	}()
	observe, stop := context.WithTimeout(h.ctx, 5*time.Second)
	defer stop()
	blocked, err := h.waitForTrackLock(observe, "FROM audio_tracks", done)
	cancel()
	if !blocked {
		_ = tx.Rollback(h.ctx)
		t.Fatalf("prior-read lock was not reached: %v", err)
	}
	failure := <-done
	_ = tx.Rollback(h.ctx)
	if err := failure; !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("prior-read failure=%v want dependency_unavailable", err)
	}
}

type trackReadKey struct{}
type trackReadTracer struct{ reached, release chan struct{} }

func (tr *trackReadTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if strings.HasPrefix(data.SQL, "SELECT ") && strings.Contains(data.SQL, " FROM audio_tracks\n WHERE ") {
		return context.WithValue(ctx, trackReadKey{}, true)
	}
	return ctx
}
func (tr *trackReadTracer) TraceQueryEnd(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryEndData) {
	if marked, _ := ctx.Value(trackReadKey{}).(bool); marked {
		close(tr.reached)
		<-tr.release
	}
}

func (h *mediaHarness) testTrackMetadataExcludesSweep(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, mediaAssetOne, mediaLessonOne, 1, domain.Ready, mediaNow.Add(time.Hour), nil)
	h.insertTrack(t, mediaTrackOne, mediaAssetOne, string(domain.EmbeddedTrack), "en", "Original", true, domain.Ready, mediaNow)
	h.insertUploadTrack(t, mediaTrackTwo, mediaAssetOne, "uk", domain.Failed, 1024, 0, mediaNow.Add(-time.Hour), mediaNow)
	cfg, err := pgxpool.ParseConfig(h.connection)
	if err != nil {
		t.Fatal(err)
	}
	tr := &trackReadTracer{make(chan struct{}), make(chan struct{})}
	cfg.ConnConfig.Tracer = tr
	pool, err := pgxpool.NewWithConfig(h.ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	defer func() {
		select {
		case <-tr.release:
		default:
			close(tr.release)
		}
	}()
	metadata := make(chan error, 1)
	go func() {
		_, err := NewRepository(pool).UpdateAudioTrackMetadata(h.ctx, ownerActor(), mediaAssetOne, mediaTrackTwo, "uk", "New label", nil, mediaNow.Add(time.Hour), noMasterRewrite)
		metadata <- err
	}()
	select {
	case <-tr.reached:
	case <-time.After(5 * time.Second):
		t.Fatal("metadata prior read not reached")
	}
	sweep := make(chan error, 1)
	go func() {
		sweep <- h.repository.DeleteSweptAudioTrackLocked(h.ctx, mediaTrackTwo, mediaNow, func() error { return nil })
	}()
	observe, stop := context.WithTimeout(h.ctx, 5*time.Second)
	defer stop()
	blocked, sweepErr := h.waitForTrackLock(observe, "SELECT EXISTS (", sweep)
	close(tr.release)
	metaErr := <-metadata
	if blocked {
		sweepErr = <-sweep
	}
	if !blocked {
		t.Fatalf("sweeper entered before metadata completed: sweep=%v metadata=%v; want contested row lock", sweepErr, metaErr)
	}
	if metaErr != nil || sweepErr != nil {
		t.Fatalf("serialized metadata=%v sweep=%v", metaErr, sweepErr)
	}
}
