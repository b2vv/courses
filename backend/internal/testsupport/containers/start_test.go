package containers

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type fakeContainer struct {
	testcontainers.Container
	terminated int
}

func (c *fakeContainer) Terminate(context.Context, ...testcontainers.TerminateOption) error {
	c.terminated++
	return nil
}

func TestStartReturnsSuccessfulFirstAttempt(t *testing.T) {
	want := &fakeContainer{}
	calls := 0
	got, err := Start(context.Background(), func(context.Context) (*fakeContainer, error) {
		calls++
		return want, nil
	})
	if err != nil || got != want || calls != 1 || want.terminated != 0 {
		t.Fatalf("Start() = (%p, %v), calls=%d terminated=%d", got, err, calls, want.terminated)
	}
}

func TestStartTerminatesRaceContainersBeforeRetrying(t *testing.T) {
	withTinyRetryConfig(t)
	first, second, want := &fakeContainer{}, &fakeContainer{}, &fakeContainer{}
	calls := 0
	got, err := Start(context.Background(), func(context.Context) (*fakeContainer, error) {
		calls++
		switch calls {
		case 1:
			return first, raceError()
		case 2:
			return second, raceError()
		default:
			return want, nil
		}
	})
	if err != nil || got != want || calls != 3 || first.terminated != 1 || second.terminated != 1 || want.terminated != 0 {
		t.Fatalf("Start() = (%p, %v), calls=%d terminations=(%d, %d, %d)", got, err, calls, first.terminated, second.terminated, want.terminated)
	}
}

func TestStartReturnsNonRaceErrorAndContainerUnchanged(t *testing.T) {
	want := &fakeContainer{}
	wantErr := errors.New("ordinary failure")
	calls := 0
	got, err := Start(context.Background(), func(context.Context) (*fakeContainer, error) {
		calls++
		return want, wantErr
	})
	if !errors.Is(err, wantErr) || got != want || calls != 1 || want.terminated != 0 {
		t.Fatalf("Start() = (%p, %v), calls=%d terminated=%d", got, err, calls, want.terminated)
	}
}

func TestStartNamesUpstreamFixWhenRetryBudgetExpires(t *testing.T) {
	withTinyRetryConfig(t)
	retryConfig.budget = 3 * time.Millisecond
	calls := 0
	got, err := Start(context.Background(), func(context.Context) (*fakeContainer, error) {
		calls++
		return &fakeContainer{}, raceError()
	})
	// The last race container was terminated, so none is handed back (no double Terminate in the caller).
	if got != nil || err == nil || !strings.Contains(err.Error(), "#3868") || calls < 2 {
		t.Fatalf("Start() = (%p, %v), calls=%d", got, err, calls)
	}
}

func TestStartStopsRetryingWhenContextIsCancelled(t *testing.T) {
	withTinyRetryConfig(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := 0
	got, err := Start(ctx, func(context.Context) (*fakeContainer, error) {
		calls++
		cancel()
		return &fakeContainer{}, raceError()
	})
	if got != nil || !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("Start() = (%p, %v), calls=%d", got, err, calls)
	}
}

func raceError() error {
	return errors.New(`wait for reaper abc: unexpected container status "created"`)
}

func withTinyRetryConfig(t *testing.T) {
	t.Helper()
	old := retryConfig
	retryConfig = retrySettings{initialBackoff: time.Millisecond, maxBackoff: time.Millisecond, budget: 100 * time.Millisecond, wait: old.wait}
	t.Cleanup(func() { retryConfig = old })
}
