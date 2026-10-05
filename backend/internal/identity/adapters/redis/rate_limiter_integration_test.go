//go:build integration

package redis

import (
	"context"
	"log/slog"
	"testing"
	"time"

	containers "courses/internal/testsupport/containers"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRateLimiterFixedWindowAgainstRedis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*testcontainers.DockerContainer, error) {
		return testcontainers.Run(
			ctx,
			"redis:7.4-bookworm",
			testcontainers.WithExposedPorts("6379/tcp"),
			testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections").WithStartupTimeout(3*time.Minute)),
		)
	})
	if err != nil {
		t.Fatalf("Redis integration requires Docker: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate Redis container: %v", err)
		}
	})
	endpoint, err := container.PortEndpoint(ctx, "6379/tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	client := redisclient.NewClient(&redisclient.Options{Addr: endpoint})
	t.Cleanup(func() { _ = client.Close() })

	limiter, err := NewRateLimiter(client, 2, time.Minute, slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	for request := 1; request <= 2; request++ {
		allowed, err := limiter.Allow(ctx, "login:student@example.test")
		if err != nil || !allowed {
			t.Fatalf("request %d: allowed=%v err=%v", request, allowed, err)
		}
	}
	allowed, err := limiter.Allow(ctx, "login:student@example.test")
	if err != nil || allowed {
		t.Fatalf("over limit: allowed=%v err=%v", allowed, err)
	}
	if err := limiter.Reset(ctx, "login:student@example.test"); err != nil {
		t.Fatal(err)
	}
	allowed, err = limiter.Allow(ctx, "login:student@example.test")
	if err != nil || !allowed {
		t.Fatalf("after reset: allowed=%v err=%v", allowed, err)
	}
}
