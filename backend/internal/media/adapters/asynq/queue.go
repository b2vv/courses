package asynq

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"courses/internal/media/domain"
	asynqlib "github.com/hibiken/asynq"
	redisclient "github.com/redis/go-redis/v9"
)

const TaskTypeTranscode = "media:transcode"

type enqueueClient interface {
	EnqueueContext(context.Context, *asynqlib.Task, ...asynqlib.Option) (*asynqlib.TaskInfo, error)
}

type Queue struct{ client enqueueClient }

func NewQueue(client redisclient.UniversalClient) *Queue {
	return newQueue(asynqlib.NewClientFromRedisClient(client))
}

func newQueue(client enqueueClient) *Queue { return &Queue{client: client} }

func (queue *Queue) Enqueue(ctx context.Context, assetID string, attempt int) error {
	payload, err := json.Marshal(struct {
		AssetID string `json:"assetId"`
		Attempt int    `json:"attempt"`
	}{AssetID: assetID, Attempt: attempt})
	if err != nil {
		return domain.ErrInternal
	}
	return queue.enqueue(ctx, payload, assetID+":"+strconv.Itoa(attempt))
}

func (queue *Queue) EnqueueAudioTrack(ctx context.Context, assetID, trackID string, attempt int) error {
	payload, err := json.Marshal(struct {
		AssetID string `json:"assetId"`
		TrackID string `json:"trackId"`
		Attempt int    `json:"attempt"`
	}{AssetID: assetID, TrackID: trackID, Attempt: attempt})
	if err != nil {
		return domain.ErrInternal
	}
	return queue.enqueue(ctx, payload, assetID+":"+trackID+":"+strconv.Itoa(attempt))
}

func (queue *Queue) enqueue(ctx context.Context, payload []byte, taskID string) error {
	_, err := queue.client.EnqueueContext(ctx, asynqlib.NewTask(TaskTypeTranscode, payload),
		asynqlib.TaskID(taskID),
		asynqlib.Queue("media"),
		asynqlib.MaxRetry(3),
		asynqlib.Timeout(30*time.Minute),
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
