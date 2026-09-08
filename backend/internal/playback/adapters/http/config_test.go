package http_test

import (
	"strings"
	"testing"

	playbackhttp "courses/internal/playback/adapters/http"
)

func TestPlaybackConfigUsesSafeDefaults(t *testing.T) {
	t.Parallel()
	config, err := playbackhttp.ConfigFromEnvironment(func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if config.HeartbeatSeconds() != 30 || config.OwnershipTTLSeconds() != 180 {
		t.Fatalf("heartbeat=%d ttl=%d", config.HeartbeatSeconds(), config.OwnershipTTLSeconds())
	}
}

func TestPlaybackConfigAcceptsTheFastTestProfile(t *testing.T) {
	t.Parallel()
	values := map[string]string{
		"PLAYBACK_HEARTBEAT_SECONDS":     "2",
		"PLAYBACK_OWNERSHIP_TTL_SECONDS": "6",
	}
	config, err := playbackhttp.ConfigFromEnvironment(func(key string) string { return values[key] })
	if err != nil {
		t.Fatal(err)
	}
	if config.HeartbeatSeconds() != 2 || config.OwnershipTTLSeconds() != 6 {
		t.Fatalf("heartbeat=%d ttl=%d", config.HeartbeatSeconds(), config.OwnershipTTLSeconds())
	}
}

func TestPlaybackConfigRejectsBrokenCadenceAtStartup(t *testing.T) {
	t.Parallel()
	_, err := playbackhttp.NewConfig(30, 60)
	if err == nil || !strings.Contains(err.Error(), "heartbeatSeconds * 3 <= ownershipTtlSeconds") {
		t.Fatalf("got %v, want named playback cadence invariant", err)
	}
}

func TestPlaybackConfigRejectsNonPositiveAndNonDecimalValues(t *testing.T) {
	t.Parallel()
	tests := []map[string]string{
		{"PLAYBACK_HEARTBEAT_SECONDS": "0", "PLAYBACK_OWNERSHIP_TTL_SECONDS": "180"},
		{"PLAYBACK_HEARTBEAT_SECONDS": "30", "PLAYBACK_OWNERSHIP_TTL_SECONDS": "-1"},
		{"PLAYBACK_HEARTBEAT_SECONDS": "thirty", "PLAYBACK_OWNERSHIP_TTL_SECONDS": "180"},
	}
	for _, values := range tests {
		if _, err := playbackhttp.ConfigFromEnvironment(func(key string) string { return values[key] }); err == nil {
			t.Fatalf("values=%v: want error", values)
		}
	}
}
