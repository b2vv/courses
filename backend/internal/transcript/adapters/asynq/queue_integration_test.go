//go:build integration

package asynq

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	asynqlib "github.com/hibiken/asynq"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestQueueTaskMarkersSurviveCompletedAndArchivedKeys(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := testcontainers.Run(ctx, "redis:7.4-bookworm",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections").WithStartupTimeout(3*time.Minute)))
	if err != nil {
		t.Fatalf("Redis integration requires Docker: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	endpoint, err := container.PortEndpoint(ctx, "6379/tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	redis := redisclient.NewClient(&redisclient.Options{Addr: endpoint})
	t.Cleanup(func() { _ = redis.Close() })

	queue := NewQueue(redis)
	job := ports.RecognitionJob{AssetID: "50000000-0000-4000-8000-000000000001", Generation: 1, TaskMarker: "2026-09-24T12:00:00Z"}
	if err := queue.Enqueue(ctx, job, time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := queue.Enqueue(ctx, job, time.Minute); !errors.Is(err, domain.ErrAlreadyRunning) {
		t.Fatalf("same marker err=%v want already_running", err)
	}

	server := asynqlib.NewServerFromRedisClient(redis, asynqlib.Config{Queues: map[string]int{"asr": 1}})
	handler := asynqlib.NewServeMux()
	handler.HandleFunc(TaskTypeRecognize, func(context.Context, *asynqlib.Task) error { return nil })
	go func() { _ = server.Run(handler) }()
	completedID := recognitionTaskID(job)
	waitForTaskState(t, ctx, redis, completedID, asynqlib.TaskStateCompleted)
	server.Shutdown()

	newJob := job
	newJob.TaskMarker = "2026-09-24T12:00:01Z"
	if err := queue.Enqueue(ctx, newJob, time.Minute); err != nil {
		t.Fatalf("new marker after completion: %v", err)
	}
	if err := queue.Enqueue(ctx, job, time.Minute); !errors.Is(err, domain.ErrAlreadyRunning) {
		t.Fatalf("old marker after completion err=%v want already_running", err)
	}

	failingJob := job
	failingJob.Generation = 2
	failingJob.TaskMarker = "2026-09-24T12:00:02Z"
	if err := queue.Enqueue(ctx, failingJob, time.Minute); err != nil {
		t.Fatal(err)
	}
	failingServer := asynqlib.NewServerFromRedisClient(redis, asynqlib.Config{Queues: map[string]int{"asr": 1}})
	failingMux := asynqlib.NewServeMux()
	failingMux.HandleFunc(TaskTypeRecognize, func(context.Context, *asynqlib.Task) error { return errors.New("engine failed") })
	go func() { _ = failingServer.Run(failingMux) }()
	waitForTaskState(t, ctx, redis, recognitionTaskID(failingJob), asynqlib.TaskStateArchived)
	failingServer.Shutdown()

	if err := queue.Enqueue(ctx, failingJob, time.Minute); !errors.Is(err, domain.ErrAlreadyRunning) {
		t.Fatalf("archived marker err=%v want already_running", err)
	}
	newFailingJob := failingJob
	newFailingJob.TaskMarker = "2026-09-24T12:00:03Z"
	if err := queue.Enqueue(ctx, newFailingJob, time.Minute); err != nil {
		t.Fatalf("new marker after archive: %v", err)
	}
}

func waitForTaskState(t *testing.T, ctx context.Context, redis redisclient.UniversalClient, id string, want asynqlib.TaskState) {
	t.Helper()
	inspector := asynqlib.NewInspectorFromRedisClient(redis)
	deadline := time.NewTimer(30 * time.Second)
	ticker := time.NewTicker(25 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		info, err := inspector.GetTaskInfo("asr", id)
		if err == nil && info != nil && info.State == want {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-deadline.C:
			var state asynqlib.TaskState
			if info != nil {
				state = info.State
			}
			t.Fatalf("task %q state=%v err=%v want=%v", id, state, err, want)
		case <-ticker.C:
		}
	}
}
