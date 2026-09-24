package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"courses/platform/httpserver"
)

func TestC16ReadyzRequiresModelFile(t *testing.T) {
	modelDir := t.TempDir()
	t.Setenv("WHISPER_MODEL_DIR", modelDir)
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "asr-worker",
		Ready:   func(_ context.Context) error { return modelReady() },
	})

	check := func(t *testing.T, want int) {
		t.Helper()
		request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != want {
			t.Fatalf("/readyz status=%d; want %d", response.Code, want)
		}
	}

	t.Run("missing model", func(t *testing.T) { check(t, http.StatusServiceUnavailable) })
	modelPath := filepath.Join(modelDir, "model.bin")
	if err := os.Mkdir(modelPath, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Run("model path is a directory", func(t *testing.T) { check(t, http.StatusServiceUnavailable) })
	if err := os.Remove(modelPath); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(modelPath, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Run("model file exists", func(t *testing.T) { check(t, http.StatusOK) })
}
