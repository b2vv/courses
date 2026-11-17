package application

import (
	"context"
	"testing"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

// Multi-audio S6: the audio heard is the leaf the claim recorded, not a fixed "original".
func TestProcessExtractsTheClaimedDefaultLeaf(t *testing.T) {
	t.Parallel()
	const uploaded = "audio-70000000-0000-4000-8000-000000000001"
	drafts := &fakeDrafts{claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken, Draft: domain.TranscriptDraft{TrackID: trackID, TrackLanguage: "uk"}}}
	drafts.claimResult.Draft.SourceLeaf = uploaded
	audio := &fakeAudio{path: "/tmp/audio.wav"}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: testSegments(1), SpeechSeconds: 30},
	}, audio, &fakeQueue{}, testClock, testLogger)

	if _, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation, TrackID: trackID, Requested: true,
	}); err != nil {
		t.Fatal(err)
	}
	if audio.leaf != uploaded {
		t.Fatalf("extracted leaf=%q, want the claimed %q", audio.leaf, uploaded)
	}
	if len(drafts.claimRequested) != 1 || !drafts.claimRequested[0] {
		t.Fatalf("claim requested=%v, want the job's true", drafts.claimRequested)
	}
}

// A job whose claim is refused (a duplicate automatic task) hears nothing.
func TestProcessRefusedClaimExtractsNothing(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{claimResult: ports.ClaimedDraft{Claimed: false}}
	audio := &fakeAudio{path: "/tmp/audio.wav"}
	r := NewRecognize(drafts, &fakeTranscriber{}, audio, &fakeQueue{}, testClock, testLogger)
	if _, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation, TrackID: trackID,
	}); err != nil {
		t.Fatal(err)
	}
	if audio.leaf != "" || drafts.completeCalls != 0 {
		t.Fatalf("a refused claim extracted %q and completed %d times", audio.leaf, drafts.completeCalls)
	}
	if len(drafts.claimRequested) != 1 || drafts.claimRequested[0] {
		t.Fatalf("an automatic job claimed with requested=%v", drafts.claimRequested)
	}
}

// The sweep enqueues automatic jobs: they must not carry the author's Requested.
func TestEnqueuePendingJobsAreAutomatic(t *testing.T) {
	t.Parallel()
	duration := 60.0
	queue := &fakeQueue{}
	drafts := &fakeDrafts{pendingResult: []ports.PendingRecognition{{
		LessonID: lessonID, AssetID: assetID, Generation: generation, TrackID: trackID, DurationSeconds: &duration,
	}}}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, queue, testClock, testLogger)
	if err := r.EnqueuePending(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].Requested || queue.jobs[0].TrackID != trackID {
		t.Fatalf("pending jobs=%+v, want one automatic job", queue.jobs)
	}
}

func TestRecoveryAndMissingDurationRetainTrackIdentity(t *testing.T) {
	for _, recovery := range []bool{false, true} {
		drafts := &fakeDrafts{claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken}}
		if recovery {
			drafts.recoverResult = []ports.RecoveredDraft{{Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: assetID, Generation: generation, TrackID: trackID}, Token: claimToken}}
		} else {
			drafts.pendingResult = []ports.PendingRecognition{{LessonID: lessonID, AssetID: assetID, Generation: generation, TrackID: trackID}}
		}
		r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)
		var err error
		if recovery {
			err = r.Recover(context.Background())
		} else {
			err = r.EnqueuePending(context.Background())
		}
		if err != nil {
			t.Fatal(err)
		}
		if !recovery && (len(drafts.claimTracks) != 1 || drafts.claimTracks[0] != trackID) {
			t.Fatalf("nil-duration claim tracks=%v; want [%s]", drafts.claimTracks, trackID)
		}
		if drafts.completeCalls != 1 || drafts.completedDraft.TrackID != trackID {
			t.Fatalf("recovery=%v completed track=%q calls=%d; want %s once", recovery, drafts.completedDraft.TrackID, drafts.completeCalls, trackID)
		}
	}
}
