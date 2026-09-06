package asynq

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/media/domain"
	asynqlib "github.com/hibiken/asynq"
)

const queueAssetID = "50000000-0000-4000-8000-000000000001"

type fakeEnqueueClient struct {
	task    *asynqlib.Task
	options []asynqlib.Option
	err     error
}

func (client *fakeEnqueueClient) EnqueueContext(_ context.Context, task *asynqlib.Task, options ...asynqlib.Option) (*asynqlib.TaskInfo, error) {
	client.task, client.options = task, options
	return nil, client.err
}

func TestQueueEnqueuesTheD10Task(t *testing.T) {
	client := &fakeEnqueueClient{}
	if err := newQueue(client).Enqueue(context.Background(), queueAssetID, 2); err != nil {
		t.Fatal(err)
	}
	if client.task.Type() != TaskTypeTranscode || string(client.task.Payload()) != `{"assetId":"50000000-0000-4000-8000-000000000001","attempt":2}` {
		t.Fatalf("type=%q payload=%s", client.task.Type(), client.task.Payload())
	}
	want := map[asynqlib.OptionType]any{
		asynqlib.TaskIDOpt:    queueAssetID + ":2",
		asynqlib.QueueOpt:     "media",
		asynqlib.MaxRetryOpt:  3,
		asynqlib.TimeoutOpt:   30 * time.Minute,
		asynqlib.RetentionOpt: 24 * time.Hour,
	}
	if len(client.options) != len(want) {
		t.Fatalf("options=%v", client.options)
	}
	for _, option := range client.options {
		if option.Value() != want[option.Type()] {
			t.Fatalf("option %v=%v, want %v", option.Type(), option.Value(), want[option.Type()])
		}
		delete(want, option.Type())
	}
	if len(want) != 0 {
		t.Fatalf("missing options=%v", want)
	}
}

func TestQueueMapsDuplicateAndDependencyErrors(t *testing.T) {
	for _, test := range []struct {
		name string
		err  error
		want error
	}{
		{name: "duplicate is success", err: asynqlib.ErrTaskIDConflict},
		{name: "redis detail is hidden", err: errors.New("redis password and address"), want: domain.ErrDependencyUnavailable},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := newQueue(&fakeEnqueueClient{err: test.err}).Enqueue(context.Background(), queueAssetID, 0)
			if !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if test.want != nil && err.Error() != test.want.Error() {
				t.Fatalf("private detail escaped: %v", err)
			}
		})
	}
}
