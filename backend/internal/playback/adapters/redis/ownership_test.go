package redis_test

import (
	"context"
	"errors"
	"testing"

	playbackredis "courses/internal/playback/adapters/redis"
	"courses/internal/playback/domain"
	redisclient "github.com/redis/go-redis/v9"
)

type getter struct {
	key    string
	result *redisclient.StringCmd
}

func TestMissingOwnershipIsAPlaybackDenial(t *testing.T) {
	t.Parallel()
	client := &getter{result: redisclient.NewStringResult("", redisclient.Nil)}
	reader := playbackredis.NewOwnershipReader(client)

	_, err := reader.Load(context.Background(), "40000000-0000-4000-8000-000000000001")
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("got %v, want forbidden", err)
	}
}

func TestCorruptOwnershipIsDependencyUnavailable(t *testing.T) {
	t.Parallel()
	tests := []string{
		`not-json`,
		`{"playbackId":"11111111111111111111111111111111","videoId":"50000000-0000-4000-8000-000000000001"}`,
	}
	for _, stored := range tests {
		client := &getter{result: redisclient.NewStringResult(stored, nil)}
		_, err := playbackredis.NewOwnershipReader(client).Load(context.Background(), "40000000-0000-4000-8000-000000000001")
		if !errors.Is(err, domain.ErrDependencyUnavailable) {
			t.Fatalf("stored=%q got %v, want dependency unavailable", stored, err)
		}
	}
}

func (client *getter) Get(_ context.Context, key string) *redisclient.StringCmd {
	client.key = key
	return client.result
}

func TestOwnershipReaderLoadsTheAuthenticatedUsersSlot(t *testing.T) {
	t.Parallel()
	client := &getter{result: redisclient.NewStringResult(`{"playbackId":"11111111111111111111111111111111","videoId":"50000000-0000-4000-8000-000000000001","lessonId":"50000000-0000-4000-8000-000000000002"}`, nil)}
	reader := playbackredis.NewOwnershipReader(client)

	owner, err := reader.Load(context.Background(), "40000000-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatal(err)
	}
	if client.key != "playback:owner:40000000-0000-4000-8000-000000000001" {
		t.Fatalf("key=%q", client.key)
	}
	if owner.PlaybackID != "11111111111111111111111111111111" || owner.VideoID != "50000000-0000-4000-8000-000000000001" || owner.LessonID != "50000000-0000-4000-8000-000000000002" {
		t.Fatalf("owner=%+v", owner)
	}
}
