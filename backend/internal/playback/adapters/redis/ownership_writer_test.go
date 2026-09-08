package redis_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	playbackredis "courses/internal/playback/adapters/redis"
	"courses/internal/playback/domain"
	redisclient "github.com/redis/go-redis/v9"
)

const (
	writerUserID     = "40000000-0000-4000-8000-000000000001"
	writerPlaybackID = "11111111111111111111111111111111"
	writerVideoID    = "50000000-0000-4000-8000-000000000001"
)

type ownershipClient struct {
	setKey        string
	setValue      any
	setExpiration time.Duration
	setCalls      int
	setResult     *redisclient.StatusCmd
	evalScript    string
	evalKeys      []string
	evalArgs      []any
	evalResult    *redisclient.Cmd
}

func (client *ownershipClient) Set(_ context.Context, key string, value any, expiration time.Duration) *redisclient.StatusCmd {
	client.setKey, client.setValue, client.setExpiration = key, value, expiration
	client.setCalls++
	return client.setResult
}

func (client *ownershipClient) Eval(_ context.Context, script string, keys []string, args ...any) *redisclient.Cmd {
	client.evalScript, client.evalKeys, client.evalArgs = script, keys, args
	return client.evalResult
}

func TestTakeAlwaysSetsTheSingleUserOwnershipKey(t *testing.T) {
	t.Parallel()
	client := &ownershipClient{setResult: redisclient.NewStatusResult("OK", nil)}
	writer := playbackredis.NewOwnershipWriter(client)
	want := domain.Ownership{PlaybackID: writerPlaybackID, VideoID: writerVideoID, LessonID: "50000000-0000-4000-8000-000000000002"}

	if err := writer.Take(context.Background(), writerUserID, want, 3*time.Minute); err != nil {
		t.Fatal(err)
	}
	if err := writer.Take(context.Background(), writerUserID, want, 3*time.Minute); err != nil {
		t.Fatal(err)
	}
	if client.setCalls != 2 || client.setKey != "playback:owner:"+writerUserID || client.setExpiration != 3*time.Minute {
		t.Fatalf("calls=%d key=%q expiration=%v", client.setCalls, client.setKey, client.setExpiration)
	}
	encoded, ok := client.setValue.([]byte)
	if !ok {
		t.Fatalf("stored value type %T, want []byte", client.setValue)
	}
	var got map[string]string
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || got["playbackId"] != want.PlaybackID || got["videoId"] != want.VideoID || got["lessonId"] != want.LessonID {
		t.Fatalf("stored=%v", got)
	}
}

func TestRenewMapsEachDistinctLuaValue(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		lua  int64
		want domain.OwnershipRenewal
	}{
		{name: "minus one means missing", lua: -1, want: domain.OwnershipMissing},
		{name: "zero means taken over", lua: 0, want: domain.OwnershipTakenOver},
		{name: "one means renewed", lua: 1, want: domain.OwnershipRenewed},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			client := &ownershipClient{evalResult: redisclient.NewCmdResult(test.lua, nil)}
			got, err := playbackredis.NewOwnershipWriter(client).Renew(context.Background(), writerUserID, writerPlaybackID, 3*time.Minute)
			if err != nil || got != test.want {
				t.Fatalf("renewal=%v err=%v, want %v", got, err, test.want)
			}
			if len(client.evalKeys) != 1 || client.evalKeys[0] != "playback:owner:"+writerUserID || len(client.evalArgs) != 2 || client.evalArgs[0] != writerPlaybackID || client.evalArgs[1] != int64(180) {
				t.Fatalf("keys=%v args=%v", client.evalKeys, client.evalArgs)
			}
			if !strings.Contains(client.evalScript, "cjson.decode") || !strings.Contains(client.evalScript, "stored.playbackId ~= ARGV[1]") || strings.Contains(client.evalScript, "v ~= ARGV[1]") {
				t.Fatalf("script does not compare decoded playbackId: %s", client.evalScript)
			}
		})
	}
}

func TestOwnershipWritesClassifyRedisFailuresAsUnavailable(t *testing.T) {
	t.Parallel()
	failure := errors.New("connection failed")
	client := &ownershipClient{
		setResult:  redisclient.NewStatusResult("", failure),
		evalResult: redisclient.NewCmdResult(nil, failure),
	}
	writer := playbackredis.NewOwnershipWriter(client)
	if err := writer.Take(context.Background(), writerUserID, domain.Ownership{PlaybackID: writerPlaybackID, VideoID: writerVideoID, LessonID: "lesson"}, time.Minute); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("take: %v", err)
	}
	if _, err := writer.Renew(context.Background(), writerUserID, writerPlaybackID, time.Minute); !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("renew: %v", err)
	}
}

func TestRenewRejectsUnexpectedLuaValues(t *testing.T) {
	t.Parallel()
	client := &ownershipClient{evalResult: redisclient.NewCmdResult(int64(2), nil)}
	_, err := playbackredis.NewOwnershipWriter(client).Renew(context.Background(), writerUserID, writerPlaybackID, time.Minute)
	if !errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v, want dependency unavailable", err)
	}
}
