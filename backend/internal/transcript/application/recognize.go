package application

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

const (
	noSpeechThreshold = 1.0

	timeoutMultiplier = 1.0
	timeoutConstant   = 300 * time.Second

	pendingLimit = 50
)

type Clock func() time.Time

type RecognizeOutcome struct {
	Failed   bool
	NoSpeech bool
	Reason   string
}

type Recognize struct {
	drafts      ports.DraftRepository
	transcriber ports.Transcriber
	audio       ports.AudioSource
	queue       ports.RecognitionQueue
	now         Clock
	logger      *slog.Logger
}

func NewRecognize(
	drafts ports.DraftRepository,
	transcriber ports.Transcriber,
	audio ports.AudioSource,
	queue ports.RecognitionQueue,
	now Clock,
	logger *slog.Logger,
) *Recognize {
	return &Recognize{
		drafts:      drafts,
		transcriber: transcriber,
		audio:       audio,
		queue:       queue,
		now:         now,
		logger:      logger,
	}
}

func (r *Recognize) Process(ctx context.Context, job ports.RecognitionJob) (RecognizeOutcome, error) {
	now := r.timestamp()
	claimed, err := r.drafts.Claim(ctx, job.LessonID, job.AssetID, job.Generation, now)
	if err != nil {
		return RecognizeOutcome{}, err
	}
	if !claimed.Claimed {
		return RecognizeOutcome{}, nil
	}

	audioPath, release, err := r.audio.Extract(ctx, job.AssetID)
	if errors.Is(err, domain.ErrNotFound) {
		return r.fail(ctx, claimed, job, "source_not_found")
	}
	if err != nil {
		if ctx.Err() != nil {
			return RecognizeOutcome{}, ctx.Err()
		}
		return r.fail(ctx, claimed, job, "audio_extraction_failed")
	}
	defer release()

	recognition, err := r.transcriber.Transcribe(ctx, audioPath, job.Language)
	if err != nil {
		if ctx.Err() != nil {
			return RecognizeOutcome{}, ctx.Err()
		}
		return r.fail(ctx, claimed, job, "engine_error")
	}

	r.logger.Info("recognition complete",
		slog.String("lessonId", job.LessonID),
		slog.Float64("speechSeconds", recognition.SpeechSeconds),
		slog.Int("segments", len(recognition.Segments)),
		slog.String("language", string(recognition.Language)),
	)

	if recognition.SpeechSeconds < noSpeechThreshold {
		return r.complete(ctx, claimed, domain.TranscriptDraft{
			LessonID:   job.LessonID,
			AssetID:    job.AssetID,
			Generation: job.Generation,
			State:      domain.DraftNoSpeech,
		})
	}

	draft, err := domain.NewTranscriptDraft(domain.TranscriptDraft{
		LessonID:   job.LessonID,
		AssetID:    job.AssetID,
		Generation: job.Generation,
		State:      domain.DraftReady,
		Language:   recognition.Language,
		Segments:   recognition.Segments,
	})
	if errors.Is(err, domain.ErrSegmentLimit) {
		return r.fail(ctx, claimed, job, "segment_limit")
	}
	if err != nil {
		return r.fail(ctx, claimed, job, "invalid_output")
	}

	return r.complete(ctx, claimed, draft)
}

func (r *Recognize) complete(ctx context.Context, claimed ports.ClaimedDraft, draft domain.TranscriptDraft) (RecognizeOutcome, error) {
	err := r.drafts.Complete(ctx, draft, claimed.Token, r.timestamp())
	if errors.Is(err, domain.ErrDraftSuperseded) || errors.Is(err, domain.ErrNotFound) {
		return RecognizeOutcome{}, nil
	}
	if err != nil {
		return RecognizeOutcome{}, err
	}
	if draft.State == domain.DraftFailed {
		return RecognizeOutcome{Failed: true, Reason: draft.Failure}, nil
	}
	if draft.State == domain.DraftNoSpeech {
		return RecognizeOutcome{NoSpeech: true}, nil
	}
	return RecognizeOutcome{}, nil
}

func (r *Recognize) fail(ctx context.Context, claimed ports.ClaimedDraft, job ports.RecognitionJob, reason string) (RecognizeOutcome, error) {
	return r.complete(ctx, claimed, domain.TranscriptDraft{
		LessonID:   job.LessonID,
		AssetID:    job.AssetID,
		Generation: job.Generation,
		State:      domain.DraftFailed,
		Failure:    reason,
	})
}

// EnqueuePending polls for ready assets without a draft for their current generation and
// enqueues a recognition task for each (plan P-1). It runs on the worker's startup ticker.
func (r *Recognize) EnqueuePending(ctx context.Context) error {
	pending, err := r.drafts.PendingRecognitions(ctx, pendingLimit)
	if err != nil {
		return err
	}
	for _, p := range pending {
		if p.DurationSeconds == nil {
			if err := r.failBeforeStart(ctx, p, "unknown_duration"); err != nil {
				r.logger.Error("failed to reject nil-duration draft",
					slog.String("lessonId", p.LessonID), slog.Any("error", err))
			}
			continue
		}
		timeout := TaskTimeout(*p.DurationSeconds)
		if err := r.queue.Enqueue(ctx, ports.RecognitionJob{
			LessonID:   p.LessonID,
			AssetID:    p.AssetID,
			Generation: p.Generation,
		}, timeout); err != nil {
			r.logger.Error("failed to enqueue recognition",
				slog.String("lessonId", p.LessonID), slog.Any("error", err))
			continue
		}
	}
	return nil
}

func (r *Recognize) failBeforeStart(ctx context.Context, p ports.PendingRecognition, reason string) error {
	now := r.timestamp()
	claimed, err := r.drafts.Claim(ctx, p.LessonID, p.AssetID, p.Generation, now)
	if err != nil {
		return err
	}
	if !claimed.Claimed {
		return nil
	}
	return r.drafts.Complete(ctx, domain.TranscriptDraft{
		LessonID:   p.LessonID,
		AssetID:    p.AssetID,
		Generation: p.Generation,
		State:      domain.DraftFailed,
		Failure:    reason,
	}, claimed.Token, r.timestamp())
}

// TaskTimeout derives the queue timeout from the audio duration (plan P-3).
func TaskTimeout(durationSeconds float64) time.Duration {
	return time.Duration(durationSeconds*timeoutMultiplier)*time.Second + timeoutConstant
}

func (r *Recognize) timestamp() time.Time {
	return r.now().UTC().Truncate(time.Microsecond)
}
