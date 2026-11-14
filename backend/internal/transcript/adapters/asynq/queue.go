package asynq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	asynqlib "github.com/hibiken/asynq"
	redisclient "github.com/redis/go-redis/v9"
)

const TaskTypeRecognize = "asr:recognize"

type enqueueClient interface {
	EnqueueContext(context.Context, *asynqlib.Task, ...asynqlib.Option) (*asynqlib.TaskInfo, error)
}

type Queue struct{ client enqueueClient }

func NewQueue(client redisclient.UniversalClient) *Queue {
	return newQueue(asynqlib.NewClientFromRedisClient(client))
}

func newQueue(client enqueueClient) *Queue { return &Queue{client: client} }

var _ ports.RecognitionQueue = (*Queue)(nil)

func (queue *Queue) Enqueue(ctx context.Context, job ports.RecognitionJob, timeout time.Duration) error {
	payload, err := json.Marshal(taskPayload{
		LessonID:   job.LessonID,
		AssetID:    job.AssetID,
		Generation: job.Generation,
		Language:   string(job.Language),
		TaskMarker: job.TaskMarker,
		Requested:  job.Requested,
	})
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	taskID := recognitionTaskID(job)
	_, err = queue.client.EnqueueContext(ctx, asynqlib.NewTask(TaskTypeRecognize, payload),
		asynqlib.TaskID(taskID),
		asynqlib.Queue("asr"),
		asynqlib.MaxRetry(0),
		asynqlib.Timeout(timeout),
		asynqlib.Retention(24*time.Hour),
	)
	if errors.Is(err, asynqlib.ErrTaskIDConflict) {
		return domain.ErrAlreadyRunning
	}
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	return nil
}

func recognitionTaskID(job ports.RecognitionJob) string {
	marker := "none"
	if job.TaskMarker != "" {
		if parsed, err := time.Parse(time.RFC3339Nano, job.TaskMarker); err == nil {
			marker = parsed.UTC().Truncate(time.Microsecond).Format("20060102T150405.000000Z")
		} else {
			marker = job.TaskMarker
		}
	}
	return fmt.Sprintf("%s:%d:%s", job.AssetID, job.Generation, marker)
}

type taskPayload struct {
	LessonID   string `json:"lessonId"`
	AssetID    string `json:"assetId"`
	Generation int32  `json:"generation"`
	Language   string `json:"language,omitempty"`
	TaskMarker string `json:"taskMarker,omitempty"`
	// Requested marks the author's Regenerate (multi-audio S6). A task enqueued before S6 has no
	// field and decodes as false — an automatic job, which is what every pre-S6 sweep task was.
	Requested bool `json:"requested,omitempty"`
}
