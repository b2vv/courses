package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"io"
	"log/slog"
)

var testLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

const (
	lessonID   = "11111111-1111-4111-8111-111111111111"
	assetID    = "22222222-2222-4222-8222-222222222222"
	generation = int32(1)
	claimToken = "2026-01-01T00:00:00Z"
)

func testClock() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

func segmentID(n int) string {
	ids := []string{
		"aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
		"bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
		"cccccccc-cccc-4ccc-8ccc-cccccccccccc",
	}
	return ids[n]
}

func testSegments(n int) []domain.Segment {
	segments := make([]domain.Segment, n)
	for i := range segments {
		c := float64(0.9)
		segments[i] = domain.Segment{
			ID: segmentID(i), StartMs: int64(i * 1000), EndMs: int64((i + 1) * 1000),
			Text: "hello", Confidence: &c,
		}
	}
	return segments
}

// --- fakes ---

type fakeDrafts struct {
	claimResult        ports.ClaimedDraft
	claimErr           error
	completeErr        error
	pendingResult      []ports.PendingRecognition
	pendingErr         error
	recoverResult      []ports.RecoveredDraft
	recoverErr         error
	completedDraft     domain.TranscriptDraft
	completedToken     string
	claimCalls         int
	completeCalls      int
	completeContextErr error
}

func (f *fakeDrafts) GetLessonDraft(context.Context, string) (ports.LessonDraft, error) {
	return ports.LessonDraft{}, nil
}

func (f *fakeDrafts) PendingRecognitions(_ context.Context, _ int) ([]ports.PendingRecognition, error) {
	return f.pendingResult, f.pendingErr
}

func (f *fakeDrafts) Claim(context.Context, string, string, int32, time.Time) (ports.ClaimedDraft, error) {
	f.claimCalls++
	return f.claimResult, f.claimErr
}

func (f *fakeDrafts) SaveDraft(context.Context, domain.TranscriptDraft, string, time.Time) (string, error) {
	return "", nil
}

func (f *fakeDrafts) Complete(ctx context.Context, draft domain.TranscriptDraft, token string, _ time.Time) error {
	f.completeCalls++
	f.completedDraft = draft
	f.completedToken = token
	f.completeContextErr = ctx.Err()
	return f.completeErr
}

func (f *fakeDrafts) Recover(context.Context, time.Time, int) ([]ports.RecoveredDraft, error) {
	return f.recoverResult, f.recoverErr
}

func (f *fakeDrafts) LessonAsset(context.Context, string) (*ports.PendingRecognition, error) {
	return nil, nil
}

type fakeTranscriber struct {
	result ports.Recognition
	err    error
}

func (f *fakeTranscriber) Transcribe(context.Context, string, domain.Language) (ports.Recognition, error) {
	return f.result, f.err
}

type fakeAudio struct {
	path     string
	err      error
	released bool
}

func (f *fakeAudio) Extract(context.Context, string) (string, func(), error) {
	if f.err != nil {
		return "", nil, f.err
	}
	return f.path, func() { f.released = true }, nil
}

type fakeQueue struct {
	jobs []ports.RecognitionJob
	err  error
}

func (f *fakeQueue) Enqueue(_ context.Context, job ports.RecognitionJob, _ time.Duration) error {
	if f.err != nil {
		return f.err
	}
	f.jobs = append(f.jobs, job)
	return nil
}

// --- Process tests ---

func TestProcessHappyPath(t *testing.T) {
	t.Parallel()
	segments := testSegments(2)
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: segments, SpeechSeconds: 30.0},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.Failed {
		t.Fatalf("expected success, got failed: %s", outcome.Reason)
	}
	if drafts.completedDraft.State != domain.DraftReady {
		t.Fatalf("state=%s want=ready", drafts.completedDraft.State)
	}
	if drafts.completedDraft.Language != "uk" {
		t.Fatalf("language=%s want=uk", drafts.completedDraft.Language)
	}
	if len(drafts.completedDraft.Segments) != 2 {
		t.Fatalf("segments=%d want=2", len(drafts.completedDraft.Segments))
	}
	if drafts.completedToken != claimToken {
		t.Fatalf("token=%s want=%s", drafts.completedToken, claimToken)
	}
}

func TestProcessNotClaimedReturnsNil(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: false},
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil || outcome.Failed {
		t.Fatalf("err=%v outcome=%+v", err, outcome)
	}
	if drafts.completeCalls != 0 {
		t.Fatal("Complete must not be called when not claimed")
	}
}

func TestProcessNoSpeech(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{SpeechSeconds: 0.3},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil || outcome.Failed {
		t.Fatalf("err=%v outcome=%+v", err, outcome)
	}
	if !outcome.NoSpeech {
		t.Fatal("outcome.NoSpeech must be true for sub-threshold speech")
	}
	if drafts.completedDraft.State != domain.DraftNoSpeech {
		t.Fatalf("state=%s want=no_speech", drafts.completedDraft.State)
	}
}

func TestProcessSourceNotFound(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{
		err: domain.ErrNotFound,
	}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Failed || outcome.Reason != "source_not_found" {
		t.Fatalf("outcome=%+v", outcome)
	}
	if drafts.completedDraft.State != domain.DraftFailed {
		t.Fatalf("state=%s want=failed", drafts.completedDraft.State)
	}
}

func TestProcessEngineError(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		err: errors.New("engine crashed"),
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Failed || outcome.Reason != "engine_error" {
		t.Fatalf("outcome=%+v", outcome)
	}
	if drafts.completedDraft.State != domain.DraftFailed {
		t.Fatalf("state=%s want=failed", drafts.completedDraft.State)
	}
}

func TestProcessContextCancelledDuringTranscribe(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewRecognize(drafts, &fakeTranscriber{
		err: context.Canceled,
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	_, err := r.Process(ctx, ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want=context.Canceled", err)
	}
	if drafts.completeCalls != 0 {
		t.Fatal("Complete must not be called on context cancellation")
	}
}

func TestProcessContextCancelledDuringExtract(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{
		err: errors.New("extract failed"),
	}, &fakeQueue{}, testClock, testLogger)

	_, err := r.Process(ctx, ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v want=context.Canceled", err)
	}
}

func TestProcessDeadlineDuringExtractFailsDraft(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken}}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{err: context.DeadlineExceeded}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(ctx, ports.RecognitionJob{LessonID: lessonID, AssetID: assetID, Generation: generation})
	if err != nil || !outcome.Failed || outcome.Reason != "timeout" {
		t.Fatalf("outcome=%+v err=%v want failed timeout", outcome, err)
	}
	if drafts.completedDraft.State != domain.DraftFailed || drafts.completedDraft.Failure != "timeout" {
		t.Fatalf("draft=%+v want failed/timeout", drafts.completedDraft)
	}
	if drafts.completeContextErr != nil {
		t.Fatalf("terminal write used canceled context: %v", drafts.completeContextErr)
	}
}

func TestProcessDeadlineDuringTranscribeFailsDraft(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken}}
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	r := NewRecognize(drafts, &fakeTranscriber{err: context.DeadlineExceeded}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(ctx, ports.RecognitionJob{LessonID: lessonID, AssetID: assetID, Generation: generation})
	if err != nil || !outcome.Failed || outcome.Reason != "timeout" {
		t.Fatalf("outcome=%+v err=%v want failed timeout", outcome, err)
	}
	if drafts.completedDraft.State != domain.DraftFailed || drafts.completedDraft.Failure != "timeout" {
		t.Fatalf("draft=%+v want failed/timeout", drafts.completedDraft)
	}
	if drafts.completeContextErr != nil {
		t.Fatalf("terminal write used canceled context: %v", drafts.completeContextErr)
	}
}

func TestProcessSegmentLimitPermanent(t *testing.T) {
	t.Parallel()
	segments := make([]domain.Segment, domain.MaximumSegments+1)
	for i := range segments {
		c := float64(0.9)
		segments[i] = domain.Segment{
			ID:      "aaaaaaaa-aaaa-4aaa-8aaa-" + padHex(i),
			StartMs: int64(i * 1000), EndMs: int64((i + 1) * 1000),
			Text: "word", Confidence: &c,
		}
	}
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: segments, SpeechSeconds: 3600},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Failed || outcome.Reason != "segment_limit" {
		t.Fatalf("outcome=%+v", outcome)
	}
	if drafts.completedDraft.Failure != "segment_limit" {
		t.Fatalf("failure=%s", drafts.completedDraft.Failure)
	}
}

func TestProcessSupersededOnComplete(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
		completeErr: domain.ErrDraftSuperseded,
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: testSegments(1), SpeechSeconds: 5.0},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil || outcome.Failed {
		t.Fatalf("superseded must be silent success: err=%v outcome=%+v", err, outcome)
	}
}

func TestProcessNotFoundOnComplete(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
		completeErr: domain.ErrNotFound,
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: testSegments(1), SpeechSeconds: 5.0},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil || outcome.Failed {
		t.Fatalf("not-found must be silent success: err=%v outcome=%+v", err, outcome)
	}
}

func TestProcessReleasesAudioOnSuccess(t *testing.T) {
	t.Parallel()
	audio := &fakeAudio{path: "/tmp/audio.wav"}
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: testSegments(1), SpeechSeconds: 5.0},
	}, audio, &fakeQueue{}, testClock, testLogger)

	r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if !audio.released {
		t.Fatal("audio must be released after transcription")
	}
}

func TestProcessReleasesAudioOnEngineError(t *testing.T) {
	t.Parallel()
	audio := &fakeAudio{path: "/tmp/audio.wav"}
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		err: errors.New("engine failed"),
	}, audio, &fakeQueue{}, testClock, testLogger)

	r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if !audio.released {
		t.Fatal("audio must be released after engine error")
	}
}

func TestProcessPassesLanguageHint(t *testing.T) {
	t.Parallel()
	var passedHint domain.Language
	transcriber := &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: testSegments(1), SpeechSeconds: 5.0},
	}
	original := transcriber.Transcribe
	_ = original
	captureTranscriber := &capturingTranscriber{
		inner:  transcriber,
		onCall: func(hint domain.Language) { passedHint = hint },
	}
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, captureTranscriber, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation, Language: "uk",
	})
	if passedHint != "uk" {
		t.Fatalf("hint=%q want=uk", passedHint)
	}
}

type capturingTranscriber struct {
	inner  *fakeTranscriber
	onCall func(domain.Language)
}

func (c *capturingTranscriber) Transcribe(ctx context.Context, path string, hint domain.Language) (ports.Recognition, error) {
	c.onCall(hint)
	return c.inner.Transcribe(ctx, path, hint)
}

func TestProcessNoSpeechOutcomeFlag(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{SpeechSeconds: 0.3},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.NoSpeech {
		t.Fatal("outcome.NoSpeech must be true")
	}
	if outcome.Failed {
		t.Fatal("no_speech must not be marked as failed")
	}
}

func TestProcessBoundaryAtExactlyOneSecond(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: testSegments(1), SpeechSeconds: 1.0},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if outcome.NoSpeech {
		t.Fatal("exactly 1.0s must NOT be no_speech")
	}
	if drafts.completedDraft.State != domain.DraftReady {
		t.Fatalf("state=%s want=ready", drafts.completedDraft.State)
	}
}

func TestProcessClaimError(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimErr: domain.ErrDependencyUnavailable,
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	_, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("err=%v want=dependency_unavailable", err)
	}
}

func TestProcessInvalidOutput(t *testing.T) {
	t.Parallel()
	badSegment := domain.Segment{
		ID: "", StartMs: 0, EndMs: 1000, Text: "hello",
	}
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
	}
	r := NewRecognize(drafts, &fakeTranscriber{
		result: ports.Recognition{Language: "uk", Segments: []domain.Segment{badSegment}, SpeechSeconds: 5.0},
	}, &fakeAudio{path: "/tmp/audio.wav"}, &fakeQueue{}, testClock, testLogger)

	outcome, err := r.Process(context.Background(), ports.RecognitionJob{
		LessonID: lessonID, AssetID: assetID, Generation: generation,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !outcome.Failed || outcome.Reason != "invalid_output" {
		t.Fatalf("outcome=%+v", outcome)
	}
}

// --- EnqueuePending tests ---

func TestEnqueuePendingEnqueuesJobs(t *testing.T) {
	t.Parallel()
	dur := float64(120)
	drafts := &fakeDrafts{
		pendingResult: []ports.PendingRecognition{
			{LessonID: lessonID, AssetID: assetID, Generation: generation, DurationSeconds: &dur},
		},
	}
	queue := &fakeQueue{}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, queue, testClock, testLogger)

	if err := r.EnqueuePending(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs=%d want=1", len(queue.jobs))
	}
	if queue.jobs[0].LessonID != lessonID {
		t.Fatalf("lessonId=%s", queue.jobs[0].LessonID)
	}
}

func TestEnqueuePendingFailsNilDuration(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		claimResult: ports.ClaimedDraft{Claimed: true, Token: claimToken},
		pendingResult: []ports.PendingRecognition{
			{LessonID: lessonID, AssetID: assetID, Generation: generation, DurationSeconds: nil},
		},
	}
	queue := &fakeQueue{}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, queue, testClock, testLogger)

	if err := r.EnqueuePending(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(queue.jobs) != 0 {
		t.Fatal("nil-duration must not enqueue")
	}
	if drafts.claimCalls != 1 {
		t.Fatalf("claim calls=%d want=1", drafts.claimCalls)
	}
	if drafts.completeCalls != 1 {
		t.Fatalf("complete calls=%d want=1", drafts.completeCalls)
	}
	if drafts.completedDraft.Failure != "unknown_duration" {
		t.Fatalf("failure=%s want=unknown_duration", drafts.completedDraft.Failure)
	}
}

func TestEnqueuePendingContinuesOnQueueError(t *testing.T) {
	t.Parallel()
	dur1 := float64(60)
	dur2 := float64(120)
	drafts := &fakeDrafts{
		pendingResult: []ports.PendingRecognition{
			{LessonID: lessonID, AssetID: assetID, Generation: 1, DurationSeconds: &dur1},
			{LessonID: "33333333-3333-4333-8333-333333333333", AssetID: "44444444-4444-4444-8444-444444444444", Generation: 1, DurationSeconds: &dur2},
		},
	}
	callCount := 0
	queue := &fakeQueue{err: nil}
	errorOnFirst := &countingQueue{
		inner: queue,
		onEnqueue: func() error {
			callCount++
			if callCount == 1 {
				return errors.New("queue down")
			}
			return nil
		},
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, errorOnFirst, testClock, testLogger)

	if err := r.EnqueuePending(context.Background()); err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Fatalf("queue calls=%d want=2 (must continue after first error)", callCount)
	}
}

func TestEnqueuePendingSkipsAlreadyRunningAndContinues(t *testing.T) {
	t.Parallel()
	dur := float64(60)
	secondLesson := "33333333-3333-4333-8333-333333333333"
	secondAsset := "44444444-4444-4444-8444-444444444444"
	drafts := &fakeDrafts{pendingResult: []ports.PendingRecognition{
		{LessonID: lessonID, AssetID: assetID, Generation: 1, DurationSeconds: &dur},
		{LessonID: secondLesson, AssetID: secondAsset, Generation: 1, DurationSeconds: &dur},
	}}
	queue := &fakeQueue{}
	callCount := 0
	queueWithConflict := &countingQueue{
		inner: queue,
		onEnqueue: func() error {
			callCount++
			if callCount == 1 {
				return domain.ErrAlreadyRunning
			}
			return nil
		},
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, queueWithConflict, testClock, testLogger)

	if err := r.EnqueuePending(context.Background()); err != nil {
		t.Fatal(err)
	}
	if callCount != 2 {
		t.Fatalf("queue calls=%d want=2", callCount)
	}
	if len(queue.jobs) != 1 || queue.jobs[0].LessonID != secondLesson {
		t.Fatalf("jobs=%v want only second pending job", queue.jobs)
	}
}

type countingQueue struct {
	inner     *fakeQueue
	onEnqueue func() error
}

func (c *countingQueue) Enqueue(ctx context.Context, job ports.RecognitionJob, timeout time.Duration) error {
	if err := c.onEnqueue(); err != nil {
		return err
	}
	return c.inner.Enqueue(ctx, job, timeout)
}

// --- Recover tests ---

func TestRecoverFailsTimedOutDraft(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		recoverResult: []ports.RecoveredDraft{
			{Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: assetID, Generation: generation, State: domain.DraftRunning}, Token: claimToken},
		},
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	if err := r.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if drafts.completeCalls != 1 {
		t.Fatalf("complete calls=%d want=1", drafts.completeCalls)
	}
	if drafts.completedDraft.State != domain.DraftFailed {
		t.Fatalf("state=%s want=failed", drafts.completedDraft.State)
	}
	if drafts.completedDraft.Failure != "timeout" {
		t.Fatalf("failure=%s want=timeout", drafts.completedDraft.Failure)
	}
	if drafts.completedToken != claimToken {
		t.Fatalf("token=%s want=%s", drafts.completedToken, claimToken)
	}
}

func TestRecoverNoDrafts(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	if err := r.Recover(context.Background()); err != nil {
		t.Fatal(err)
	}
	if drafts.completeCalls != 0 {
		t.Fatal("Complete must not be called when no drafts recovered")
	}
}

func TestRecoverRepositoryError(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{recoverErr: domain.ErrDependencyUnavailable}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	err := r.Recover(context.Background())
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("err=%v want=dependency_unavailable", err)
	}
}

func TestRecoverSupersededIsSilent(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		recoverResult: []ports.RecoveredDraft{
			{Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: assetID, Generation: generation, State: domain.DraftRunning}, Token: claimToken},
		},
		completeErr: domain.ErrDraftSuperseded,
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	if err := r.Recover(context.Background()); err != nil {
		t.Fatalf("superseded must not propagate: err=%v", err)
	}
}

func TestRecoverNotFoundIsSilent(t *testing.T) {
	t.Parallel()
	drafts := &fakeDrafts{
		recoverResult: []ports.RecoveredDraft{
			{Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: assetID, Generation: generation, State: domain.DraftRunning}, Token: claimToken},
		},
		completeErr: domain.ErrNotFound,
	}
	r := NewRecognize(drafts, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	if err := r.Recover(context.Background()); err != nil {
		t.Fatalf("not-found must not propagate: err=%v", err)
	}
}

func TestRecoverContinuesOnCompleteError(t *testing.T) {
	t.Parallel()
	secondLesson := "33333333-3333-4333-8333-333333333333"
	secondAsset := "44444444-4444-4444-8444-444444444444"
	completeCallCount := 0
	drafts := &fakeDrafts{
		recoverResult: []ports.RecoveredDraft{
			{Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: assetID, Generation: 1, State: domain.DraftRunning}, Token: "tok1"},
			{Draft: domain.TranscriptDraft{LessonID: secondLesson, AssetID: secondAsset, Generation: 1, State: domain.DraftRunning}, Token: "tok2"},
		},
	}
	errOnce := &errorOnceDrafts{
		inner: drafts,
		onComplete: func() error {
			completeCallCount++
			if completeCallCount == 1 {
				return domain.ErrDependencyUnavailable
			}
			return nil
		},
	}
	r := NewRecognize(errOnce, &fakeTranscriber{}, &fakeAudio{}, &fakeQueue{}, testClock, testLogger)

	if err := r.Recover(context.Background()); err != nil {
		t.Fatalf("must not propagate transient Complete error: err=%v", err)
	}
	if completeCallCount != 2 {
		t.Fatalf("complete calls=%d want=2 (must continue after first error)", completeCallCount)
	}
}

type errorOnceDrafts struct {
	inner      *fakeDrafts
	onComplete func() error
}

func (e *errorOnceDrafts) GetLessonDraft(ctx context.Context, id string) (ports.LessonDraft, error) {
	return e.inner.GetLessonDraft(ctx, id)
}
func (e *errorOnceDrafts) PendingRecognitions(ctx context.Context, limit int) ([]ports.PendingRecognition, error) {
	return e.inner.PendingRecognitions(ctx, limit)
}
func (e *errorOnceDrafts) Claim(ctx context.Context, l, a string, g int32, now time.Time) (ports.ClaimedDraft, error) {
	return e.inner.Claim(ctx, l, a, g, now)
}
func (e *errorOnceDrafts) SaveDraft(ctx context.Context, d domain.TranscriptDraft, t string, now time.Time) (string, error) {
	return e.inner.SaveDraft(ctx, d, t, now)
}
func (e *errorOnceDrafts) Complete(ctx context.Context, d domain.TranscriptDraft, t string, now time.Time) error {
	return e.onComplete()
}
func (e *errorOnceDrafts) Recover(ctx context.Context, now time.Time, limit int) ([]ports.RecoveredDraft, error) {
	return e.inner.Recover(ctx, now, limit)
}
func (e *errorOnceDrafts) LessonAsset(ctx context.Context, id string) (*ports.PendingRecognition, error) {
	return e.inner.LessonAsset(ctx, id)
}

// --- TaskTimeout tests ---

func TestTaskTimeout(t *testing.T) {
	t.Parallel()
	got := TaskTimeout(600)
	want := 600*time.Second + 300*time.Second
	if got != want {
		t.Fatalf("timeout=%v want=%v", got, want)
	}
}

func TestTaskTimeoutShortAudio(t *testing.T) {
	t.Parallel()
	got := TaskTimeout(2)
	want := 2*time.Second + 300*time.Second
	if got != want {
		t.Fatalf("timeout=%v want=%v", got, want)
	}
}

// helpers

func padHex(n int) string {
	s := "000000000000"
	hex := []byte(s)
	for i := len(hex) - 1; n > 0 && i >= 0; i-- {
		digit := n % 16
		if digit < 10 {
			hex[i] = byte('0' + digit)
		} else {
			hex[i] = byte('a' + digit - 10)
		}
		n /= 16
	}
	return string(hex)
}
