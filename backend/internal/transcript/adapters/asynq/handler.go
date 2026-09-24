package asynq

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/observability"
	asynqlib "github.com/hibiken/asynq"
)

var errInvalidTaskPayload = errors.New("invalid asr task payload")

type Handler struct {
	process func(context.Context, ports.RecognitionJob) (application.RecognizeOutcome, error)
	metrics *observability.Metrics
}

func NewHandler(
	process func(context.Context, ports.RecognitionJob) (application.RecognizeOutcome, error),
	metrics *observability.Metrics,
) *Handler {
	return &Handler{process: process, metrics: metrics}
}

func (handler *Handler) ProcessTask(ctx context.Context, task *asynqlib.Task) error {
	started := time.Now()
	metricOutcome := "ready"
	defer func() {
		handler.metrics.RecognitionJobDuration.WithLabelValues(metricOutcome).Observe(time.Since(started).Seconds())
	}()

	payload, err := decodePayload(task.Payload())
	if err != nil {
		metricOutcome = "failed"
		return err
	}
	outcome, err := handler.process(ctx, ports.RecognitionJob{
		LessonID:   payload.LessonID,
		AssetID:    payload.AssetID,
		Generation: payload.Generation,
		Language:   domain.Language(payload.Language),
		TaskMarker: payload.TaskMarker,
	})
	if err != nil {
		metricOutcome = "failed"
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			metricOutcome = "cancelled"
		}
		return err
	}
	if outcome.Failed {
		metricOutcome = "failed"
		handler.metrics.RecognitionFailures.WithLabelValues(outcome.Reason).Inc()
	} else if outcome.NoSpeech {
		metricOutcome = "no_speech"
	}
	return nil
}

func NewErrorHandler(logger *slog.Logger) asynqlib.ErrorHandler {
	return asynqlib.ErrorHandlerFunc(func(ctx context.Context, task *asynqlib.Task, err error) {
		lessonID := "unknown"
		if payload, payloadErr := decodePayload(task.Payload()); payloadErr == nil {
			lessonID = payload.LessonID
		}
		reason := "dependency_unavailable"
		if errors.Is(err, errInvalidTaskPayload) {
			reason = "invalid_task_payload"
		} else if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			reason = "cancelled"
		}
		logger.Error("asr job failed", "lessonId", lessonID, "reason", reason)
	})
}

func decodePayload(encoded []byte) (taskPayload, error) {
	var payload taskPayload
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return payload, errInvalidTaskPayload
	}
	if payload.LessonID == "" || payload.AssetID == "" || payload.Generation < 1 {
		return payload, errInvalidTaskPayload
	}
	return payload, nil
}
