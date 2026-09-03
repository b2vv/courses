package httpserver_test

import (
	"context"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"courses/platform/httpserver"
)

func TestShutdownDrainsAnInFlightResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	entered, release := make(chan struct{}), make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(entered)
		select {
		case <-release:
			_, _ = w.Write([]byte("completed"))
		case <-r.Context().Done():
			t.Error("in-flight request canceled before drain")
		}
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- httpserver.Serve(ctx, listener, handler) }()
	result := make(chan string, 1)
	go func() {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://" + listener.Addr().String())
		if err != nil {
			result <- err.Error()
			return
		}
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			result <- err.Error()
			return
		}
		result <- string(body)
	}()
	select {
	case <-entered:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not enter handler")
	}
	cancel()
	select {
	case err := <-done:
		t.Fatalf("shutdown returned before draining: %v", err)
	case <-time.After(30 * time.Millisecond):
	}
	close(release)
	select {
	case body := <-result:
		if body != "completed" {
			t.Fatalf("got %q", body)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("response did not drain")
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("server did not exit")
	}
}
