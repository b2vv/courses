package asynq

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"courses/platform/observability"
	asynqlib "github.com/hibiken/asynq"
)

type fakeEnqueueClient struct {
	task    *asynqlib.Task
	err     error
	options []asynqlib.Option
}

func (c *fakeEnqueueClient) EnqueueContext(_ context.Context, task *asynqlib.Task, options ...asynqlib.Option) (*asynqlib.TaskInfo, error) {
	c.task = task
	c.options = options
	return nil, c.err
}

func TestQueueUsesAssetGenerationAndStableMarker(t *testing.T) {
	job := ports.RecognitionJob{
		AssetID: "50000000-0000-4000-8000-000000000001", Generation: 7, TrackID: "track-one",
		TaskMarker: "2026-09-24T12:34:56.123456789Z",
	}
	if got := recognitionTaskID(job); got != "50000000-0000-4000-8000-000000000001:7:track-one:20260924T123456.123456Z" {
		t.Fatalf("task id=%q", got)
	}
	if got := recognitionTaskID(job); got != recognitionTaskID(job) {
		t.Fatal("task id is not stable")
	}
	if got := recognitionTaskID(ports.RecognitionJob{AssetID: job.AssetID, Generation: job.Generation, TrackID: job.TrackID}); got != job.AssetID+":7:track-one:none" {
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

// Requested survives the queue; a legacy task missing trackId is refused (MS-D3).
func TestPayloadCarriesRequested(t *testing.T) {
	for _, requested := range []bool{true, false} {
		client := &fakeEnqueueClient{}
		job := ports.RecognitionJob{LessonID: "40000000-0000-4000-8000-000000000001", AssetID: "50000000-0000-4000-8000-000000000001", Generation: 1, TrackID: "70000000-0000-4000-8000-000000000001", Requested: requested}
		if err := newQueue(client).Enqueue(context.Background(), job, time.Minute); err != nil {
			t.Fatal(err)
		}
		payload, err := decodePayload(client.task.Payload())
		if err != nil || payload.Requested != requested {
			t.Fatalf("requested=%v decoded=%+v err=%v", requested, payload, err)
		}
	}
	legacy, err := decodePayload([]byte(`{"lessonId":"40000000-0000-4000-8000-000000000001","assetId":"50000000-0000-4000-8000-000000000001","generation":1}`))
	if !errors.Is(err, errInvalidTaskPayload) {
		t.Fatalf("legacy payload=%+v err=%v, want invalid_task_payload", legacy, err)
	}
}

func TestTwoTracksEnqueueDistinctTaskIDs(t *testing.T) {
	job := ports.RecognitionJob{LessonID: "lesson", AssetID: "asset", Generation: 7, TaskMarker: "marker"}
	for _, track := range []string{"one", "two"} {
		job.TrackID = track
		client := &fakeEnqueueClient{}
		if err := newQueue(client).Enqueue(context.Background(), job, time.Minute); err != nil {
			t.Fatal(err)
		}
		var taskID string
		for _, option := range client.options {
			if option.Type() == asynqlib.TaskIDOpt {
				taskID = option.Value().(string)
			}
		}
		if want := "asset:7:" + track + ":marker"; taskID != want {
			t.Fatalf("TaskID=%q want %q", taskID, want)
		}
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(client.task.Payload(), &payload); err != nil {
			t.Fatal(err)
		}
		if _, exists := payload["language"]; exists {
			t.Fatal("job payload still carries a language override")
		}
	}
}

func TestLegacyTaskRefusedWithoutProcessingOrRetries(t *testing.T) {
	client := &fakeEnqueueClient{}
	// A pre-cutover producer has no trackId. Enqueue's MaxRetry(0) makes a refusal terminal.
	if err := newQueue(client).Enqueue(context.Background(), ports.RecognitionJob{LessonID: "lesson", AssetID: "asset", Generation: 1}, time.Minute); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, option := range client.options {
		if option.Type() == asynqlib.MaxRetryOpt {
			found = true
			if option.Value() != 0 {
				t.Fatalf("MaxRetry=%v want 0", option.Value())
			}
		}
	}
	if !found {
		t.Fatal("missing MaxRetry option")
	}
	called := false
	handler := NewHandler(func(context.Context, ports.RecognitionJob) (application.RecognizeOutcome, error) {
		called = true
		return application.RecognizeOutcome{}, nil
	}, observability.NewMetrics())
	if err := handler.ProcessTask(context.Background(), client.task); !errors.Is(err, errInvalidTaskPayload) {
		t.Fatalf("err=%v want invalid payload", err)
	}
	if called {
		t.Fatal("process called for a task without trackId")
	}
}

func TestTrackIDUsesThePublishedWireKey(t *testing.T) {
	const track = "70000000-0000-4000-8000-000000000001"
	client := &fakeEnqueueClient{}
	if err := newQueue(client).Enqueue(context.Background(), ports.RecognitionJob{LessonID: "lesson", AssetID: "asset", Generation: 1, TrackID: track}, time.Minute); err != nil {
		t.Fatal(err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(client.task.Payload(), &payload); err != nil {
		t.Fatal(err)
	}
	if string(payload["trackId"]) != `"70000000-0000-4000-8000-000000000001"` {
		t.Fatalf("wire trackId=%s; want literal track UUID", payload["trackId"])
	}
	decoded, err := decodePayload([]byte(`{"lessonId":"lesson","assetId":"asset","generation":1,"trackId":"70000000-0000-4000-8000-000000000001"}`))
	if err != nil || decoded.TrackID != track {
		t.Fatalf("hand-authored trackId decoded=%q err=%v", decoded.TrackID, err)
	}
}

// Stop after Claim to observe the task-to-application identity boundary without an engine.
type taskClaimRecorder struct {
	ports.DraftRepository
	calls int
	track string
}

func (r *taskClaimRecorder) Claim(_ context.Context, _, _ string, _ int32, track string, _ bool, _ time.Time) (ports.ClaimedDraft, error) {
	r.calls++
	r.track = track
	return ports.ClaimedDraft{}, nil
}
func TestTaskHandlerPassesTrackIDToClaim(t *testing.T) {
	drafts := &taskClaimRecorder{}
	recognize := application.NewRecognize(drafts, nil, nil, nil, func() time.Time { return time.Date(2026, 11, 17, 12, 0, 0, 0, time.UTC) }, nil)
	handler := NewHandler(recognize.Process, observability.NewMetrics())
	task := asynqlib.NewTask(TaskTypeRecognize, []byte(`{"lessonId":"40000000-0000-4000-8000-000000000001","assetId":"50000000-0000-4000-8000-000000000001","generation":2,"trackId":"70000000-0000-4000-8000-000000000002","requested":true}`))
	if err := handler.ProcessTask(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	if drafts.calls != 1 || drafts.track != "70000000-0000-4000-8000-000000000002" {
		t.Fatalf("Claim calls=%d track=%q; want one claim for 70000000-0000-4000-8000-000000000002", drafts.calls, drafts.track)
	}
}
