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
	})
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	taskID := fmt.Sprintf("%s:%d", job.LessonID, job.Generation)
	_, err = queue.client.EnqueueContext(ctx, asynqlib.NewTask(TaskTypeRecognize, payload),
		asynqlib.TaskID(taskID),
		asynqlib.Queue("asr"),
		asynqlib.MaxRetry(0),
		asynqlib.Timeout(timeout),
		asynqlib.Retention(24*time.Hour),
	)
	if errors.Is(err, asynqlib.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	return nil
}

type taskPayload struct {
	LessonID   string `json:"lessonId"`
	AssetID    string `json:"assetId"`
	Generation int32  `json:"generation"`
	Language   string `json:"language,omitempty"`
}
