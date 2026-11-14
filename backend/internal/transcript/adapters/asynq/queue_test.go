package asynq

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	asynqlib "github.com/hibiken/asynq"
)

type fakeEnqueueClient struct {
	task *asynqlib.Task
	err  error
}

func (c *fakeEnqueueClient) EnqueueContext(_ context.Context, task *asynqlib.Task, _ ...asynqlib.Option) (*asynqlib.TaskInfo, error) {
	c.task = task
	return nil, c.err
}

func TestQueueUsesAssetGenerationAndStableMarker(t *testing.T) {
	job := ports.RecognitionJob{
		AssetID: "50000000-0000-4000-8000-000000000001", Generation: 7,
		TaskMarker: "2026-09-24T12:34:56.123456789Z",
	}
	if got := recognitionTaskID(job); got != "50000000-0000-4000-8000-000000000001:7:20260924T123456.123456Z" {
		t.Fatalf("task id=%q", got)
	}
	if got := recognitionTaskID(job); got != recognitionTaskID(job) {
		t.Fatal("task id is not stable")
	}
	if got := recognitionTaskID(ports.RecognitionJob{AssetID: job.AssetID, Generation: job.Generation}); got != job.AssetID+":7:none" {
		t.Fatalf("empty marker task id=%q", got)
	}
}

func TestQueueMapsTaskIDConflictToAlreadyRunning(t *testing.T) {
	client := &fakeEnqueueClient{err: asynqlib.ErrTaskIDConflict}
	err := newQueue(client).Enqueue(context.Background(), ports.RecognitionJob{
		AssetID: "50000000-0000-4000-8000-000000000001", Generation: 1,
	}, time.Minute)
	if !errors.Is(err, domain.ErrAlreadyRunning) {
		t.Fatalf("err=%v want already_running", err)
	}
}

// Multi-audio S6: Requested survives the queue, and a pre-S6 payload without it is automatic.
func TestPayloadCarriesRequested(t *testing.T) {
	for _, requested := range []bool{true, false} {
		client := &fakeEnqueueClient{}
		job := ports.RecognitionJob{LessonID: "40000000-0000-4000-8000-000000000001", AssetID: "50000000-0000-4000-8000-000000000001", Generation: 1, Requested: requested}
		if err := newQueue(client).Enqueue(context.Background(), job, time.Minute); err != nil {
			t.Fatal(err)
		}
		payload, err := decodePayload(client.task.Payload())
		if err != nil || payload.Requested != requested {
			t.Fatalf("requested=%v decoded=%+v err=%v", requested, payload, err)
		}
	}
	legacy, err := decodePayload([]byte(`{"lessonId":"40000000-0000-4000-8000-000000000001","assetId":"50000000-0000-4000-8000-000000000001","generation":1}`))
	if err != nil || legacy.Requested {
		t.Fatalf("legacy payload=%+v err=%v, want automatic", legacy, err)
	}
}
