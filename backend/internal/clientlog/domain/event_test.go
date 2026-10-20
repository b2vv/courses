package domain

import (
	"errors"
	"strings"
	"testing"
)

func TestNewClientEventNormalizesLevel(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   string
		want Level
		err  bool
	}{
		{in: "", want: LevelError},
		{in: "error", want: LevelError},
		{in: "warn", want: LevelWarn},
		{in: "info", want: LevelInfo},
		{in: "fatal", err: true},
		{in: "ERROR", err: true},
	}
	for _, tc := range cases {
		got, err := NewClientEvent(Input{Level: tc.in, Message: "x"})
		if tc.err {
			if !errors.Is(err, ErrInvalidEvent) {
				t.Fatalf("level %q: want ErrInvalidEvent, got %v", tc.in, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("level %q: unexpected error %v", tc.in, err)
		}
		if got.Level != tc.want {
			t.Fatalf("level %q: want %q, got %q", tc.in, tc.want, got.Level)
		}
	}
}

func TestNewClientEventFillsSentinels(t *testing.T) {
	t.Parallel()
	event, err := NewClientEvent(Input{Message: "boot failed"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Route != unknownSentinel {
		t.Fatalf("empty route: want %q, got %q", unknownSentinel, event.Route)
	}
	if event.Release != unknownSentinel {
		t.Fatalf("empty release: want %q, got %q", unknownSentinel, event.Release)
	}
}

func TestNewClientEventScrubsSecretsInMessageAndStack(t *testing.T) {
	t.Parallel()
	// A single message carrying several §17 shapes at once: an in-text key=value,
	// a quoted JSON field, a cookie, a Bearer header and a JWT.
	message := `login failed password=hunter2 for "token":"abc123" cookie: session=deadbeef Authorization: Bearer sk_live_9f8g7h and jwt eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.sig3xyz`
	stack := `at post (https://app/x?access_token=zzz111) merchant_secret=topsecret`

	event, err := NewClientEvent(Input{Level: "error", Message: message, ErrorStack: stack})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	forbidden := []string{
		"hunter2", "abc123", "deadbeef", "sk_live_9f8g7h",
		"eyJhbGciOiJIUzI1NiJ9", "zzz111", "topsecret",
	}
	for _, secret := range forbidden {
		if strings.Contains(event.Message, secret) || strings.Contains(event.ErrorStack, secret) {
			t.Fatalf("secret %q survived redaction: message=%q stack=%q", secret, event.Message, event.ErrorStack)
		}
	}
	// The keys themselves (context) survive — only values are masked.
	if !strings.Contains(event.Message, "password") || !strings.Contains(event.Message, redactedPlaceholder) {
		t.Fatalf("redaction removed the key or placeholder: %q", event.Message)
	}
}

func TestNewClientEventTruncatesOversizedFields(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", maxMessageLength+500)
	longStack := strings.Repeat("b", maxErrorStackLength+500)
	event, err := NewClientEvent(Input{Level: "error", Message: long, ErrorStack: longStack})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len([]rune(event.Message)) != maxMessageLength {
		t.Fatalf("message not truncated to cap: got %d", len([]rune(event.Message)))
	}
	if len([]rune(event.ErrorStack)) != maxErrorStackLength {
		t.Fatalf("stack not truncated to cap: got %d", len([]rune(event.ErrorStack)))
	}
}

func TestNewClientEventKeepsSafeFields(t *testing.T) {
	t.Parallel()
	event, err := NewClientEvent(Input{
		Level:     "warn",
		Message:   "render boundary caught",
		Route:     "/learn/esoterica",
		Release:   "a1b2c3d",
		UserAgent: "Chrome",
		ErrorName: "TypeError",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if event.Message != "render boundary caught" || event.Route != "/learn/esoterica" ||
		event.Release != "a1b2c3d" || event.UserAgent != "Chrome" || event.ErrorName != "TypeError" {
		t.Fatalf("safe fields mangled: %+v", event)
	}
}

func TestNewClientEventScrubsTokenInRoute(t *testing.T) {
	t.Parallel()
	event, err := NewClientEvent(Input{Level: "error", Message: "x", Route: "/learn/x?token=leak999"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(event.Route, "leak999") {
		t.Fatalf("token survived in route: %q", event.Route)
	}
}
