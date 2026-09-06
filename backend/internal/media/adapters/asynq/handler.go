package asynq

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
	"courses/platform/observability"
	asynqlib "github.com/hibiken/asynq"
)

var errInvalidTaskPayload = errors.New("invalid media task payload")

type taskPayload struct {
	AssetID string `json:"assetId"`
	Attempt *int   `json:"attempt"`
}

type Handler struct {
	process func(context.Context, ports.Job) (application.TranscodeOutcome, error)
	metrics *observability.Metrics
}

func NewHandler(process func(context.Context, ports.Job) (application.TranscodeOutcome, error), metrics *observability.Metrics) *Handler {
	return &Handler{process: process, metrics: metrics}
}

func (handler *Handler) ProcessTask(ctx context.Context, task *asynqlib.Task) error {
	started := time.Now()
	metricOutcome := "ready"
	defer func() {
		handler.metrics.TranscodingJobDuration.WithLabelValues(metricOutcome).Observe(time.Since(started).Seconds())
	}()

	payload, err := decodeTaskPayload(task.Payload())
	if err != nil {
		metricOutcome = "failed"
		return err
	}
	retryCount, _ := asynqlib.GetRetryCount(ctx)
	maxRetry, _ := asynqlib.GetMaxRetry(ctx)
	outcome, err := handler.process(ctx, ports.Job{
		AssetID: payload.AssetID, Attempt: *payload.Attempt, RetryCount: retryCount, MaxRetry: maxRetry,
	})
	if err != nil {
		metricOutcome = "failed"
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			metricOutcome = "cancelled"
		}
		return err
	}
	if outcome.Permanent {
		metricOutcome = "failed"
		handler.metrics.TranscodingFailures.WithLabelValues(string(outcome.Reason)).Inc()
		return fmt.Errorf("%s: %w", outcome.Reason, asynqlib.SkipRetry)
	}
	if outcome.Failed {
		metricOutcome = "failed"
		handler.metrics.TranscodingFailures.WithLabelValues(string(outcome.Reason)).Inc()
		return errors.New(string(outcome.Reason))
	}
	return nil
}

func NewErrorHandler(logger *slog.Logger) asynqlib.ErrorHandler {
	return asynqlib.ErrorHandlerFunc(func(ctx context.Context, task *asynqlib.Task, err error) {
		payload, payloadErr := decodeTaskPayload(task.Payload())
		assetID, state, reason := "", "unknown", "dependency_unavailable"
		if payloadErr == nil {
			assetID = payload.AssetID
		}
		retryCount, _ := asynqlib.GetRetryCount(ctx)
		maxRetry, _ := asynqlib.GetMaxRetry(ctx)
		switch {
		case errors.Is(err, errInvalidTaskPayload):
			reason = "invalid_task_payload"
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			state, reason = string(domain.Queued), "cancelled"
		case errors.Is(err, asynqlib.SkipRetry):
			state, reason = string(domain.Failed), string(domain.UnsupportedSource)
		default:
			switch domain.FailureReason(err.Error()) {
			case domain.UnsupportedSource, domain.ProbeFailed, domain.RenderFailed, domain.PublishFailed, domain.StorageUnavailable:
				reason = err.Error()
				state = string(domain.Queued)
				if retryCount >= maxRetry {
					state = string(domain.Failed)
				}
			}
		}
		logger.Error("media job failed", "assetId", assetID, "state", state, "reason", reason,
			"retryCount", retryCount, "maxRetry", maxRetry)
	})
}

func decodeTaskPayload(encoded []byte) (taskPayload, error) {
	var payload taskPayload
	if err := json.Unmarshal(encoded, &payload); err != nil || !validTaskAssetID(payload.AssetID) || payload.Attempt == nil || *payload.Attempt < 0 {
		return payload, errInvalidTaskPayload
	}
	return payload, nil
}

func validTaskAssetID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}
