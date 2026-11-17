package main

import (
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"
)

func TestWorkerRejectsInvalidMultilingualBeforeDependencies(t *testing.T) {
	t.Setenv("POSTGRES_HOST", "")
	for _, value := range []string{"invalid", "true", "false"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("ASR_MULTILINGUAL_DEFAULT", value)
			err := run(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
			want := "POSTGRES_HOST is required"
			if value == "invalid" {
				want = "ASR_MULTILINGUAL_DEFAULT must be true or false"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("worker startup ASR=%s error=%v; want %q", value, err, want)
			}
		})
	}
}
