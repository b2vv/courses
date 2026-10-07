// Package containers retries the temporary testcontainers-go Ryuk create-to-start race.
//
// TEMPORARY(testcontainers-go#3868): remove this package and its call sites when a
// testcontainers-go release contains PR #3868; then run the integration gate three times.
//
// In testcontainers-go v0.44.0, DockerProvider.CreateContainer first creates the target
// (docker.go:1314-1323) and then acquires the reaper (docker.go:1360-1362). On a failed
// acquisition it returns that target with the error (docker.go:1360-1361). GenericContainer
// returns its non-nil target on create/start errors (generic.go:68-89); testcontainers.Run
// preserves it (generic.go:130-140), and the postgres module wraps it in *PostgresContainer
// (modules/postgres/postgres.go:168-183).
// This package terminates that returned target before retrying the specific race; it never
// terminates a reaper. The failed reuse path returns a nil reaper (reaper.go:277-289), while
// reaperSpawner.instance is assigned only after success (reaper.go:302-304), so the next call
// can perform a fresh reuseOrCreate.
package containers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

type retrySettings struct {
	initialBackoff time.Duration
	maxBackoff     time.Duration
	budget         time.Duration
	wait           func(context.Context, time.Duration) error
}

var retryConfig = retrySettings{
	initialBackoff: 200 * time.Millisecond,
	maxBackoff:     2 * time.Second,
	budget:         30 * time.Second,
	wait: func(ctx context.Context, delay time.Duration) error {
		timer := time.NewTimer(delay)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	},
}

// Start starts a test container, retrying only testcontainers-go's temporary Ryuk race.
func Start[T testcontainers.Container](ctx context.Context, start func(context.Context) (T, error)) (T, error) {
	settings := retryConfig
	retryCtx, cancel := context.WithTimeout(ctx, settings.budget)
	defer cancel()
	deadline, _ := retryCtx.Deadline()
	backoff := settings.initialBackoff

	for {
		// The budget bounds only this retry loop. Each start gets the caller's context unchanged: a short
		// deadline on container creation is itself what orphans a reaper upstream (#3918), and callers allow
		// minutes for image pulls and database start-up.
		container, err := start(ctx)
		if err == nil {
			return container, nil
		}
		if !isReaperRace(err) {
			return container, err
		}
		// K5: a retried race used to vanish — the attempt that failed left no trace once the retry
		// passed, so a flake's message was lost. Log it before retrying; the gate keeps its output.
		log.Printf("testcontainers reaper race (will retry): %v", err)

		// The target was created before reaper acquisition. Remove only that target.
		if terminateErr := testcontainers.TerminateContainer(container); terminateErr != nil {
			log.Printf("testcontainers reaper race: terminate returned target: %v", terminateErr)
		}
		// It is gone now: never hand a terminated container back, or the caller's cleanup terminates it twice.
		var zero T
		container = zero

		if err := ctx.Err(); err != nil {
			return container, err
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return container, retryExhausted(err)
		}
		if backoff > remaining {
			backoff = remaining
		}
		if err := settings.wait(retryCtx, backoff); err != nil {
			if parentErr := ctx.Err(); parentErr != nil {
				return container, parentErr
			}
			return container, retryExhausted(err)
		}
		if backoff < settings.maxBackoff {
			backoff *= 2
			if backoff > settings.maxBackoff {
				backoff = settings.maxBackoff
			}
		}
		if time.Until(deadline) <= 0 {
			return container, retryExhausted(err)
		}
	}
}

func isReaperRace(err error) bool {
	message := err.Error()
	return strings.Contains(message, "wait for reaper") && strings.Contains(message, `unexpected container status "created"`)
}

func retryExhausted(err error) error {
	return fmt.Errorf("testcontainers-go reaper race retry exhausted; upgrade to a release containing #3868: %w", err)
}
