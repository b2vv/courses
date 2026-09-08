//go:build integration

package redis

import (
	"context"
	"testing"
	"time"

	"courses/internal/playback/domain"
	redisclient "github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestRenewOwnershipLuaAgainstRedis(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	container, err := testcontainers.Run(
		ctx,
		"redis:7.4-bookworm",
		testcontainers.WithExposedPorts("6379/tcp"),
		testcontainers.WithWaitStrategy(wait.ForLog("Ready to accept connections").WithStartupTimeout(time.Minute)),
	)
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

	const (
		userID     = "40000000-0000-4000-8000-000000000001"
		playbackID = "11111111111111111111111111111111"
	)
	key := ownershipKey(userID)
	assertLuaResult := func(name string, want int64) {
		t.Helper()
		t.Run(name, func(t *testing.T) {
			got, err := client.Eval(ctx, renewOwnershipScript, []string{key}, playbackID, int64(180)).Int64()
			if err != nil || got != want {
				t.Fatalf("result=%d err=%v, want %d", got, err, want)
			}
		})
	}

	assertLuaResult("missing key returns minus one", -1)
	writer := NewOwnershipWriter(client)
	if err := writer.Take(ctx, userID, domain.Ownership{
		PlaybackID: "22222222222222222222222222222222",
		VideoID:    "50000000-0000-4000-8000-000000000001",
		LessonID:   "60000000-0000-4000-8000-000000000001",
	}, time.Minute); err != nil {
		t.Fatal(err)
	}
	assertLuaResult("foreign playback id returns zero", 0)
	if ttl, err := client.TTL(ctx, key).Result(); err != nil || ttl > 90*time.Second {
		t.Fatalf("foreign owner ttl=%v err=%v, must not be renewed", ttl, err)
	}
	if err := writer.Take(ctx, userID, domain.Ownership{
		PlaybackID: playbackID,
		VideoID:    "50000000-0000-4000-8000-000000000001",
		LessonID:   "60000000-0000-4000-8000-000000000001",
	}, time.Minute); err != nil {
		t.Fatal(err)
	}
	assertLuaResult("matching playback id returns one", 1)
	if ttl, err := client.TTL(ctx, key).Result(); err != nil || ttl < 2*time.Minute {
		t.Fatalf("ttl=%v err=%v, want renewed near 3m", ttl, err)
	}
	if err := client.Set(ctx, key, "not-json", time.Minute).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := writer.Renew(ctx, userID, playbackID, 3*time.Minute); err != domain.ErrDependencyUnavailable {
		t.Fatalf("corrupt ownership: %v", err)
	}
}
