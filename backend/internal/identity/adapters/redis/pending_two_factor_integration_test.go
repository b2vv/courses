//go:build integration

package redis

import (
	"context"
	"testing"
	"time"

	"courses/internal/identity/ports"
	containers "courses/internal/testsupport/containers"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestPendingTwoFactorStoreAgainstRedis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := containers.Start(ctx, func(ctx context.Context) (*testcontainers.DockerContainer, error) {
		return testcontainers.Run(ctx, "redis:7.4-bookworm", testcontainers.WithExposedPorts("6379/tcp"), testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections").WithStartupTimeout(3*time.Minute)))
	})
	if err != nil {
		t.Fatalf("Redis integration requires Docker: %v", err)
	}
	t.Cleanup(func() { _ = container.Terminate(context.Background()) })
	endpoint, err := container.PortEndpoint(ctx, "6379/tcp", "")
	if err != nil {
		t.Fatal(err)
	}
	client := redisclient.NewClient(&redisclient.Options{Addr: endpoint})
	t.Cleanup(func() { _ = client.Close() })
	store, err := NewPendingTwoFactorStore(client)
	if err != nil {
		t.Fatal(err)
	}
	want := ports.PendingTwoFactor{UserID: "40000000-0000-4000-8000-000000000001", RememberMe: true, LoginRateLimitKey: "login:learner@example.com:203.0.113.7"}
	if err := store.Put(ctx, "pending-id", want, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, ok, err := store.Take(ctx, "pending-id")
	if err != nil || !ok || got != want {
		t.Fatalf("first take: %+v, %v, %v", got, ok, err)
	}
	if _, ok, err := store.Take(ctx, "pending-id"); err != nil || ok {
		t.Fatalf("second take must be empty: ok=%v err=%v", ok, err)
	}
}
