//go:build integration

package postgres_test

import (
	"strings"
	"testing"
	"time"

	"courses/internal/transcript/domain"
)

// Multi-audio S6 (docs/multi-audio/plan-s6.md): recognition hears the default audio track.

func (h *draftHarness) pendingLessons(t *testing.T) map[string]bool {
	t.Helper()
	pending, err := h.repository.PendingRecognitions(h.ctx, 50)
	if err != nil {
		t.Fatal(err)
	}
	lessons := map[string]bool{}
	for _, p := range pending {
		lessons[p.LessonID] = true
	}
	return lessons
}

func (h *draftHarness) completeDraftText(t *testing.T, lessonID, assetID string, generation int32, token, text string, at time.Time) {
	t.Helper()
	if err := h.repository.Complete(h.ctx, domain.TranscriptDraft{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
		State: domain.DraftReady, Language: domain.Language("en"),
		Segments: []domain.Segment{{ID: "60000000-0000-4000-8000-000000000001", EndMs: 1000, Text: text}},
	}, token, at); err != nil {
		t.Fatalf("Complete: %v", err)
	}
}

func (h *draftHarness) testPendingFollowsDefaultLeaf(t *testing.T) {
	h.reset(t)
	// (a) a pre-S6 draft (NULL leaf) on an asset whose default is the embedded original.
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO asr_drafts(lesson_id,asset_id,generation,state,language,segments,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,1,'ready','en','[{"id":"60000000-0000-4000-8000-000000000001","startMs":0,"endMs":1000,"text":"old"}]'::jsonb,$3,$3)`,
		recoverLesson1, recoverAsset1, draftNow); err != nil {
		t.Fatal(err)
	}
	// (b) a draft that heard the original, after the default moved to an uploaded track.
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	h.completeReadyDraft(t, recoverLesson2, recoverAsset2, 1, draftNow)
	h.insertDefaultTrack(t, recoverAsset2, "uploaded", "")
	// (c) a draft that heard the current default.
	h.insertAssetWithDuration(t, recoverAsset3, recoverLesson3, 1, 60)
	h.completeReadyDraft(t, recoverLesson3, recoverAsset3, 1, draftNow)
	// (d) a Ready asset with no default track and no draft.
	h.insertAssetWithDuration(t, recoverAsset4, recoverLesson4, 1, 60)
	if _, err := h.pool.Exec(h.ctx, `DELETE FROM audio_tracks WHERE video_asset_id=$1::text::uuid`, recoverAsset4); err != nil {
		t.Fatal(err)
	}

	got := h.pendingLessons(t)
	for lesson, want := range map[string]bool{recoverLesson1: false, recoverLesson2: true, recoverLesson3: false, recoverLesson4: false} {
		if got[lesson] != want {
			t.Errorf("lesson %s pending=%v, want %v (all=%v)", lesson, got[lesson], want, got)
		}
	}
}

func (h *draftHarness) testClaimRecordsDefaultLeaf(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	uploaded := h.insertDefaultTrack(t, recoverAsset1, "uploaded", "")
	claimed, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, false, draftNow)
	if err != nil || !claimed.Claimed {
		t.Fatalf("claim=%+v err=%v", claimed, err)
	}
	if claimed.Draft.SourceLeaf != uploaded || !strings.HasPrefix(uploaded, "audio-") {
		t.Fatalf("claimed leaf=%q, want the uploaded default %q", claimed.Draft.SourceLeaf, uploaded)
	}
	var stored string
	if err := h.pool.QueryRow(h.ctx, `SELECT source_leaf FROM asr_drafts WHERE lesson_id=$1::text::uuid`, recoverLesson1).Scan(&stored); err != nil || stored != uploaded {
		t.Fatalf("stored leaf=%q err=%v, want %q", stored, err, uploaded)
	}

	// An asset without a default Ready track is not claimable: recognition waits for one (D5).
	h.insertAssetWithDuration(t, recoverAsset2, recoverLesson2, 1, 60)
	if _, err := h.pool.Exec(h.ctx, `DELETE FROM audio_tracks WHERE video_asset_id=$1::text::uuid`, recoverAsset2); err != nil {
		t.Fatal(err)
	}
	if _, err := h.repository.Claim(h.ctx, recoverLesson2, recoverAsset2, 1, true, draftNow); err != domain.ErrNotFound {
		t.Fatalf("claim without a default err=%v, want ErrNotFound", err)
	}
}

func (h *draftHarness) testAutomaticDuplicateNotReclaimed(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	uploaded := h.insertDefaultTrack(t, recoverAsset1, "uploaded", "")

	// Two automatic replacements were queued (before and after the first run finished).
	first, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, false, draftNow.Add(time.Minute))
	if err != nil || !first.Claimed || first.Draft.SourceLeaf != uploaded {
		t.Fatalf("first replacement claim=%+v err=%v", first, err)
	}
	h.completeDraftText(t, recoverLesson1, recoverAsset1, 1, first.Token, "first replacement", draftNow.Add(2*time.Minute))
	second, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, false, draftNow.Add(3*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if second.Claimed {
		t.Fatal("a duplicate automatic task re-claimed a draft that already heard the default")
	}
	draft, err := h.repository.GetLessonDraft(h.ctx, recoverLesson1)
	if err != nil {
		t.Fatal(err)
	}
	if draft.Draft.State != domain.DraftReady || draft.Draft.Segments[0].Text != "first replacement" {
		t.Fatalf("draft=%+v, want the first replacement kept", draft.Draft)
	}
	if h.pendingLessons(t)[recoverLesson1] {
		t.Fatal("a draft that heard the default is still pending")
	}

	// The author's Regenerate may still re-recognise the same audio.
	manual, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, true, draftNow.Add(4*time.Minute))
	if err != nil || !manual.Claimed {
		t.Fatalf("requested claim=%+v err=%v", manual, err)
	}
}

func (h *draftHarness) testDraftStaleAfterDefaultChange(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	read := func() bool {
		t.Helper()
		draft, err := h.repository.GetLessonDraft(h.ctx, recoverLesson1)
		if err != nil {
			t.Fatal(err)
		}
		return draft.Draft.Stale(draft.ReadyGeneration, draft.DefaultLeaf)
	}
	if read() {
		t.Fatal("a draft that heard the default reads stale")
	}
	h.insertDefaultTrack(t, recoverAsset1, "uploaded", "")
	if !read() {
		t.Fatal("a draft that heard another track than the default does not read stale")
	}
}

func (h *draftHarness) testRerunLeavesTranscriptUntouched(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO transcripts(lesson_id,language,authored_against_generation,created_at,updated_at)
		VALUES($1::text::uuid,'uk',1,$2,$2)`, recoverLesson1, draftNow); err != nil {
		t.Fatal(err)
	}
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO transcript_segments(id,lesson_id,language,ordinal,start_ms,end_ms,text)
		VALUES('61000000-0000-4000-8000-000000000001',$1::text::uuid,'uk',0,0,1000,'Опубліковано автором')`, recoverLesson1); err != nil {
		t.Fatal(err)
	}
	snapshot := func() string {
		t.Helper()
		var value string
		if err := h.pool.QueryRow(h.ctx, `SELECT transcripts.authored_against_generation::text||'|'||transcripts.updated_at::text||'|'||segments.text
			FROM transcripts JOIN transcript_segments AS segments USING (lesson_id, language)
			WHERE transcripts.lesson_id=$1::text::uuid`, recoverLesson1).Scan(&value); err != nil {
			t.Fatal(err)
		}
		return value
	}
	before := snapshot()
	h.completeReadyDraft(t, recoverLesson1, recoverAsset1, 1, draftNow)
	h.insertDefaultTrack(t, recoverAsset1, "uploaded", "")
	rerun, err := h.repository.Claim(h.ctx, recoverLesson1, recoverAsset1, 1, false, draftNow.Add(time.Minute))
	if err != nil || !rerun.Claimed {
		t.Fatalf("rerun claim=%+v err=%v", rerun, err)
	}
	h.completeDraftText(t, recoverLesson1, recoverAsset1, 1, rerun.Token, "rerun", draftNow.Add(2*time.Minute))
	if after := snapshot(); after != before {
		t.Fatalf("the published transcript changed: %q -> %q", before, after)
	}
}

func (h *draftHarness) testDraftSourceLeafShape(t *testing.T) {
	h.reset(t)
	h.insertAssetWithDuration(t, recoverAsset1, recoverLesson1, 1, 60)
	for _, leaf := range []string{"../original", "audio-", "audio-NOT-A-UUID", "Original"} {
		_, err := h.pool.Exec(h.ctx, `INSERT INTO asr_drafts(lesson_id,asset_id,generation,state,segments,source_leaf,created_at,updated_at)
			VALUES($1::text::uuid,$2::text::uuid,1,'running','[]'::jsonb,$3,$4,$4)`, recoverLesson1, recoverAsset1, leaf, draftNow)
		if err == nil || !strings.Contains(err.Error(), "asr_drafts_source_leaf_shape") {
			t.Errorf("leaf %q err=%v, want the shape CHECK", leaf, err)
		}
	}
}
