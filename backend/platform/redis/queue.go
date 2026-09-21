package redis

import (
	"context"
	"log/slog"
	"time"

	"courses/platform/observability"
	"github.com/hibiken/asynq"
)

// QueueLogger does not emit Asynq's free-form arguments, which can include task
// payloads and connection details. Dependency health and error level remain visible.
type QueueLogger struct{ Logger *slog.Logger }

func (l QueueLogger) Debug(...interface{}) { l.Logger.Debug("queue runtime event") }
func (l QueueLogger) Info(...interface{})  { l.Logger.Info("queue runtime event") }
func (l QueueLogger) Warn(...interface{}) {
	l.Logger.Warn("queue runtime warning; check readiness and queue metrics")
}
func (l QueueLogger) Error(...interface{}) {
	l.Logger.Error("queue runtime failure; check readiness and queue metrics")
}
func (l QueueLogger) Fatal(...interface{}) {
	l.Logger.Error("queue runtime fatal error; check readiness and queue metrics")
}

// MonitorQueue updates gauges in a background loop, never by blocking metrics scrapes.
func MonitorQueue(ctx context.Context, inspector *asynq.Inspector, metrics *observability.Metrics, logger *slog.Logger, queue string) {
	update := func() {
		// GetQueueInfo does not wrap its missing-queue error with the public
		// ErrQueueNotFound sentinel. List first to distinguish absent from failed.
		queues, err := inspector.Queues()
		info := &asynq.QueueInfo{}
		if err == nil {
			for _, q := range queues {
				if q == queue {
					info, err = inspector.GetQueueInfo(q)
					break
				}
			}
		}
		if err != nil {
			for _, state := range []string{"pending", "active", "scheduled", "retry", "archived"} {
				metrics.QueueDepth.WithLabelValues(queue, state).Set(-1)
			}
			logger.Error("queue inspection failed")
			return
		}
		for state, count := range map[string]int{"pending": info.Pending, "active": info.Active, "scheduled": info.Scheduled, "retry": info.Retry, "archived": info.Archived} {
			metrics.QueueDepth.WithLabelValues(queue, state).Set(float64(count))
		}
	}
	update()
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			update()
		}
	}
}
