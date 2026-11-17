//go:build integration

package postgres_test

import (
	"strings"
	"testing"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

// S6 successors: recognition belongs to each ready track, independently of the default.
const (
	recoverTrack2 = "a6000000-0000-4000-8000-000000000002"
	recoverTrack3 = "a6000000-0000-4000-8000-000000000003"
	recoverTrack4 = "a6000000-0000-4000-8000-000000000004"
	recoverTrack5 = "a6000000-0000-4000-8000-000000000005"
	recoverTrack6 = "a6000000-0000-4000-8000-000000000006"
)

func (h *draftHarness) pendingTracks(t *testing.T) map[string]ports.PendingRecognition {
	t.Helper()
	pending, err := h.repository.PendingRecognitions(h.ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	tracks := map[string]ports.PendingRecognition{}
	for _, p := range pending {
		if _, exists := tracks[p.TrackID]; exists {
			t.Fatalf("duplicate pending track: %+v", p)
		}
		tracks[p.TrackID] = p
	}
	return tracks
}

func (h *draftHarness) lessonDraft(t *testing.T, lessonID, trackID string) ports.LessonDraft {
	t.Helper()
	drafts, err := h.repository.ListLessonDrafts(h.ctx, lessonID)
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range drafts {
		if d.Draft.TrackID == trackID {
			return d
		}
	}
	t.Fatalf("track %s missing from drafts %+v", trackID, drafts)
	return ports.LessonDraft{}
}

func (h *draftHarness) completeDraftText(t *testing.T, draft domain.TranscriptDraft, token, text string, at time.Time) {
	t.Helper()
	draft.State, draft.Language = domain.DraftReady, "en"
	draft.Segments = []domain.Segment{{ID: "60000000-0000-4000-8000-000000000001", EndMs: 1000, Text: text}}
	if err := h.repository.Complete(h.ctx, draft, token, at); err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func (h *draftHarness) testPendingPerTrack(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	// All non-ready track states are excluded, including failed.
	for i, state := range []string{"created", "uploading", "queued", "processing", "failed"} {
		id := []string{recoverTrack3, recoverTrack4, recoverTrack5, recoverTrack6, initialTrackID(recoverAsset7)}[i]
		h.insertTrack(t, id, recoverAsset1, "uploaded", "zz-"+id[len(id)-4:], state, false)
	}
	// A former Ready asset stays in the table, but its tracks are not candidates.
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset3, recoverLesson2, 2, 60)
	h.completeReadyDraft(t, recoverLesson2, recoverAsset3, 2, draftNow)
	// awaiting_audio is an asset state; audio tracks use created/uploading/queued/processing.
	h.insertAssetWithDuration(t, recoverAsset4, recoverLesson3, 1, 60)
	h.exec(t, `UPDATE video_assets SET state='awaiting_audio' WHERE id=$1::text::uuid`, recoverAsset4)
	assertTwo := func() {
		t.Helper()
		got := h.pendingTracks(t)
		if len(got) != 2 {
			t.Fatalf("pending=%+v; want two tracks of current asset", got)
		}
		for _, id := range []string{initialTrackID(recoverAsset1), recoverTrack2} {
			p, ok := got[id]
			if !ok || p.AssetID != recoverAsset1 || p.LessonID != recoverLesson1 || p.Generation != 1 || p.DurationSeconds == nil || *p.DurationSeconds != 60 {
				t.Fatalf("pending[%s]=%+v found=%v", id, p, ok)
			}
		}
	}
	assertTwo()
	// Even absence of a default does not suppress ready tracks.
	h.exec(t, `UPDATE audio_tracks SET is_default=false WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	assertTwo()
	h.exec(t, `UPDATE audio_tracks SET is_default=true WHERE id=$1::text::uuid`, recoverTrack2)
	assertTwo()
}

func (h *draftHarness) testClaimRecordsTrack(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "DE-AT", "ready", false)
	h.exec(t, `UPDATE audio_tracks SET multilingual=true WHERE id=$1::text::uuid`, recoverTrack2)
	claimed, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, recoverTrack2, false, draftNow)
	if err != nil || !claimed.Claimed {
		t.Fatalf("claim=%+v err=%v", claimed, err)
	}
	if claimed.Draft.TrackID != recoverTrack2 || claimed.Draft.TrackLanguage != "de-at" || !claimed.Draft.TrackMultilingual || claimed.Draft.SourceLeaf != "audio-"+recoverTrack2 {
		t.Fatalf("wrong selected-track snapshot: %+v", claimed)
	}
	stored := h.lessonDraft(t, recoverLesson1, recoverTrack2)
	if stored.Draft.TrackLanguage != "de-at" || !stored.Draft.TrackMultilingual || stored.Draft.SourceLeaf != "audio-"+recoverTrack2 || stored.IsDefault {
		t.Fatalf("stored snapshot=%+v", stored)
	}
	// Selecting a ready non-default track remains legal without any default.
	h.exec(t, `UPDATE audio_tracks SET is_default=false WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	selected, err := h.repository.LessonTrack(h.ctx, recoverLesson1, recoverTrack2)
	if err != nil || selected == nil || selected.TrackID != recoverTrack2 {
		t.Fatalf("selected=%+v err=%v", selected, err)
	}
	original, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
	if err != nil || !original.Claimed || original.Draft.SourceLeaf != "original" || original.Draft.TrackMultilingual {
		t.Fatalf("embedded claim=%+v err=%v", original, err)
	}
}

func (h *draftHarness) testAutomaticDuplicateNotReclaimed(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	trackID := initialTrackID(recoverAsset1)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	before := h.lessonDraft(t, recoverLesson1, trackID)
	h.exec(t, `UPDATE audio_tracks SET language='de',multilingual=true WHERE id=$1::text::uuid`, trackID)
	duplicate, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, trackID, false, draftNow.Add(time.Minute))
	if err != nil || duplicate.Claimed {
		t.Fatalf("automatic claim=%+v err=%v", duplicate, err)
	}
	kept := h.lessonDraft(t, recoverLesson1, trackID)
	if !kept.Draft.Stale(kept.TrackLanguage, kept.TrackMultilingual) || kept.Draft.State != domain.DraftReady || kept.Draft.Segments[0].Text != "ready" || kept.UpdatedAt != before.UpdatedAt || kept.Draft.TrackLanguage != "en" || kept.Draft.TrackMultilingual {
		t.Fatalf("automatic duplicate changed stale output: %+v", kept)
	}
	if pending := h.pendingTracks(t); len(pending) != 0 {
		t.Fatalf("stale draft became pending: %+v", pending)
	}
	manual, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, trackID, true, draftNow.Add(2*time.Minute))
	if err != nil || !manual.Claimed || manual.Draft.TrackLanguage != "de" || !manual.Draft.TrackMultilingual || manual.Token == before.UpdatedAt {
		t.Fatalf("requested claim=%+v err=%v", manual, err)
	}
	fresh := h.lessonDraft(t, recoverLesson1, trackID)
	if fresh.Draft.State != domain.DraftRunning || len(fresh.Draft.Segments) != 0 || fresh.Draft.Stale(fresh.TrackLanguage, fresh.TrackMultilingual) {
		t.Fatalf("requested snapshot=%+v", fresh)
	}
}

func (h *draftHarness) testDraftStaleAfterMetadataChange(t *testing.T) {
	for _, item := range []struct {
		name, from, to string
		mode, stale    bool
	}{
		{"en to de", "en", "de", false, true},
		{"und to uk", "und", "uk", false, false},
		{"und family to uk", "und-x-original", "uk", false, false},
		{"regional rename", "en-US", "en-GB", false, true},
		{"multilingual flip", "en", "en", true, true},
		{"und mode flip", "und", "uk", true, true},
	} {
		t.Run(item.name, func(t *testing.T) {
			h.reset(t)
			h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
			trackID := initialTrackID(recoverAsset1)
			h.exec(t, `UPDATE audio_tracks SET language=$2 WHERE id=$1::text::uuid`, trackID, item.from)
			h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
			h.exec(t, `UPDATE audio_tracks SET language=$2,multilingual=$3 WHERE id=$1::text::uuid`, trackID, item.to, item.mode)
			draft := h.lessonDraft(t, recoverLesson1, trackID)
			if got := draft.Draft.Stale(draft.TrackLanguage, draft.TrackMultilingual); got != item.stale {
				t.Fatalf("stale=%v want=%v draft=%+v", got, item.stale, draft)
			}
			if got := h.pendingTracks(t); len(got) != 0 {
				t.Fatalf("metadata change became pending: %+v", got)
			}
		})
	}
	// Changing the default alone no longer makes either track's output stale.
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	h.exec(t, `UPDATE audio_tracks SET is_default=false WHERE video_asset_id=$1::text::uuid`, recoverAsset1)
	h.exec(t, `UPDATE audio_tracks SET is_default=true WHERE id=$1::text::uuid`, recoverTrack2)
	draft := h.lessonDraft(t, recoverLesson1, initialTrackID(recoverAsset1))
	if draft.IsDefault || draft.Draft.Stale(draft.TrackLanguage, draft.TrackMultilingual) {
		t.Fatalf("default switch changed staleness: %+v", draft)
	}
}

func (h *draftHarness) seedPublishedTranscript(t *testing.T) {
	t.Helper()
	h.exec(t, `INSERT INTO transcripts(lesson_id,language,authored_against_generation,created_at,updated_at)
		VALUES($1::text::uuid,'en',1,$2,$2)`, recoverLesson1, draftNow)
	h.exec(t, `INSERT INTO transcript_segments(id,lesson_id,language,ordinal,start_ms,end_ms,text)
		VALUES('61000000-0000-4000-8000-000000000001',$1::text::uuid,'en',0,0,1000,'Published by author')`, recoverLesson1)
}

func (h *draftHarness) publishedSnapshot(t *testing.T) string {
	t.Helper()
	var value string
	if err := h.pool.QueryRow(h.ctx, `SELECT row_to_json(t)::text||'|'||row_to_json(s)::text
		FROM transcripts t JOIN transcript_segments s USING (lesson_id,language)
		WHERE t.lesson_id=$1::text::uuid AND t.language='en'`, recoverLesson1).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func (h *draftHarness) testRerunLeavesTranscriptUntouched(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.seedPublishedTranscript(t)
	before := h.publishedSnapshot(t)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	rerun, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), true, draftNow.Add(time.Minute))
	if err != nil || !rerun.Claimed {
		t.Fatalf("rerun=%+v err=%v", rerun, err)
	}
	h.completeDraftText(t, rerun.Draft, rerun.Token, "rerun", draftNow.Add(2*time.Minute))
	if after := h.publishedSnapshot(t); after != before {
		t.Fatalf("published transcript changed: %q -> %q", before, after)
	}
}

func (h *draftHarness) testTrackSourceLeafShape(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	for _, leaf := range []string{"../original", "audio-", "audio-NOT-A-UUID", "Original"} {
		_, err := h.pool.Exec(h.ctx, `UPDATE audio_tracks SET source_leaf=$2 WHERE id=$1::text::uuid`, initialTrackID(recoverAsset1), leaf)
		if err == nil || !strings.Contains(err.Error(), "audio_tracks_leaf_matches_kind") {
			t.Errorf("leaf=%q err=%v; want track shape CHECK", leaf, err)
		}
	}
	claim, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow)
	if err != nil || !claim.Claimed || claim.Draft.SourceLeaf != "original" {
		t.Fatalf("claim after refused edits=%+v err=%v", claim, err)
	}
}

func (h *draftHarness) testTracksRecoverIndependently(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	h.exec(t, `UPDATE audio_tracks SET multilingual=true WHERE id=$1::text::uuid`, recoverTrack2)
	first, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, initialTrackID(recoverAsset1), false, draftNow.Add(-400*time.Second))
	if err != nil || !first.Claimed {
		t.Fatalf("first=%+v err=%v", first, err)
	}
	second, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, recoverTrack2, false, draftNow.Add(-200*time.Second))
	if err != nil || !second.Claimed {
		t.Fatalf("second=%+v err=%v", second, err)
	}
	recovered, err := h.repository.Recover(h.ctx, draftNow, 50)
	if err != nil || len(recovered) != 1 || recovered[0].Draft.TrackID != first.Draft.TrackID || recovered[0].Token != first.Token {
		t.Fatalf("recover=%+v err=%v", recovered, err)
	}
	// Token from the other track cannot complete this row.
	if err := h.repository.Complete(h.ctx, second.Draft, first.Token, draftNow); err != domain.ErrDraftSuperseded {
		t.Fatalf("cross-track token err=%v", err)
	}
	recovered[0].Draft.State, recovered[0].Draft.Failure = domain.DraftFailed, "timeout"
	if err := h.repository.Complete(h.ctx, recovered[0].Draft, recovered[0].Token, draftNow); err != nil {
		t.Fatal(err)
	}
	stillRunning := h.lessonDraft(t, recoverLesson1, recoverTrack2)
	if stillRunning.Draft.State != domain.DraftRunning || stillRunning.UpdatedAt != second.Token {
		t.Fatalf("sibling changed: %+v", stillRunning)
	}
	// Later the second track expires with its own identity, token and mode snapshot.
	recovered, err = h.repository.Recover(h.ctx, draftNow.Add(200*time.Second), 50)
	if err != nil || len(recovered) != 1 || recovered[0].Draft.TrackID != recoverTrack2 || recovered[0].Token != second.Token || recovered[0].Draft.TrackLanguage != "de" || !recovered[0].Draft.TrackMultilingual {
		t.Fatalf("second recover=%+v err=%v", recovered, err)
	}
	h.completeDraftText(t, recovered[0].Draft, recovered[0].Token, "second completed", draftNow.Add(201*time.Second))
	if firstDraft := h.lessonDraft(t, recoverLesson1, first.Draft.TrackID); firstDraft.Draft.State != domain.DraftFailed || firstDraft.Draft.Failure != "timeout" {
		t.Fatalf("first track changed: %+v", firstDraft)
	}
	if secondDraft := h.lessonDraft(t, recoverLesson1, recoverTrack2); secondDraft.Draft.State != domain.DraftReady || secondDraft.Draft.Segments[0].Text != "second completed" {
		t.Fatalf("second track=%+v", secondDraft)
	}
}

func (h *draftHarness) testTrackDeletePreservesTranscript(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset1, "uploaded", "de", "ready", false)
	h.seedPublishedTranscript(t)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	other, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, recoverTrack2, false, draftNow)
	if err != nil || !other.Claimed {
		t.Fatalf("other=%+v err=%v", other, err)
	}
	h.completeDraftText(t, other.Draft, other.Token, "sibling", draftNow.Add(time.Second))
	before := h.publishedSnapshot(t)
	h.exec(t, `DELETE FROM audio_tracks WHERE id=$1::text::uuid`, initialTrackID(recoverAsset1))
	drafts, err := h.repository.ListLessonDrafts(h.ctx, recoverLesson1)
	if err != nil || len(drafts) != 1 || drafts[0].Draft.TrackID != recoverTrack2 || drafts[0].Draft.Segments[0].Text != "sibling" {
		t.Fatalf("drafts after delete=%+v err=%v", drafts, err)
	}
	var count int
	if err := h.pool.QueryRow(h.ctx, `SELECT count(*) FROM asr_drafts WHERE audio_track_id=$1::text::uuid`, initialTrackID(recoverAsset1)).Scan(&count); err != nil || count != 0 {
		t.Fatalf("deleted track drafts=%d err=%v", count, err)
	}
	if after := h.publishedSnapshot(t); after != before {
		t.Fatalf("track deletion changed published language: %q -> %q", before, after)
	}
}

func (h *draftHarness) exec(t *testing.T, query string, args ...any) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, query, args...); err != nil {
		t.Fatal(err)
	}
}

func (h *draftHarness) testM2R10TrackSelection(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson1, 2, 60)
	h.insertAssetWithDuration(t, recoverAsset3, recoverLesson2, 1, 60)
	h.insertTrack(t, recoverTrack2, recoverAsset2, "uploaded", "de", "ready", false)
	h.insertTrack(t, recoverTrack3, recoverAsset2, "uploaded", "pl", "processing", false)
	for _, tc := range []struct {
		name, lesson, track string
		found               bool
	}{
		{"requested non-default", recoverLesson1, recoverTrack2, true},
		{"current default", recoverLesson1, initialTrackID(recoverAsset2), true},
		{"old Ready asset", recoverLesson1, initialTrackID(recoverAsset1), false},
		{"foreign lesson", recoverLesson1, initialTrackID(recoverAsset3), false},
		{"non-ready track", recoverLesson1, recoverTrack3, false},
		{"missing track", recoverLesson1, recoverTrack4, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := h.repository.LessonTrack(h.ctx, tc.lesson, tc.track)
			if err != nil || (got != nil) != tc.found {
				t.Fatalf("track=%+v err=%v found=%v", got, err, tc.found)
			}
			if got != nil && (got.TrackID != tc.track || got.AssetID != recoverAsset2) {
				t.Fatalf("wrong selection=%+v", got)
			}
		})
	}
	h.completeReadyDraft(t, recoverLesson1, recoverAsset2, 2, draftNow)
	if drafts, err := h.repository.ListLessonDrafts(h.ctx, recoverLesson1); err != nil || len(drafts) != 1 {
		t.Fatalf("Ready draft setup=%+v err=%v", drafts, err)
	}
	// One pending asset per lesson (video_assets_pending_per_lesson): demote only the latest; the
	// older Ready asset has no drafts, so the list must be empty either way.
	h.exec(t, `UPDATE video_assets SET state='processing' WHERE id=$1::text::uuid`, recoverAsset2)
	if drafts, err := h.repository.ListLessonDrafts(h.ctx, recoverLesson1); err != nil || len(drafts) != 0 {
		t.Fatalf("no Ready drafts=%+v err=%v", drafts, err)
	}
}
