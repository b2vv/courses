package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"courses/platform/httpserver"
	"courses/platform/observability"
	redisinfrastructure "courses/platform/redis"
	"courses/platform/runtime"
	"github.com/hibiken/asynq"
)

func main() {
	logger := observability.NewLogger("asr-worker")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, logger); err != nil {
		logger.Error("worker stopped with an error", "reason", err.Error())
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func run(ctx context.Context, logger *slog.Logger) error {
	dependencies, err := runtime.Open(ctx, "asr-worker", logger)
	if err != nil {
		return err
	}
	defer dependencies.Close()
	queueOption, err := redisinfrastructure.QueueOption(dependencies.RedisURL)
	if err != nil {
		return err
	}
	var queueHealthy atomic.Bool
	queueHealthy.Store(true)
	server := asynq.NewServer(queueOption, asynq.Config{
		Concurrency: 1, Queues: map[string]int{"asr": 1}, ShutdownTimeout: 10 * time.Second,
		Logger:              redisinfrastructure.QueueLogger{Logger: logger},
		HealthCheckInterval: 5 * time.Second,
		HealthCheckFunc:     func(err error) { queueHealthy.Store(err == nil) },
	})
	mux := asynq.NewServeMux()
	if err := server.Start(mux); err != nil {
		return errors.New("queue startup failed")
	}
	defer server.Shutdown()
	inspector := asynq.NewInspector(queueOption)
	defer func() {
		if err := inspector.Close(); err != nil {
			logger.Error("queue inspector shutdown failed")
		}
	}()
	monitor, stopMonitor := context.WithCancel(ctx)
	monitorDone := make(chan struct{})
	go func() {
		defer close(monitorDone)
		redisinfrastructure.MonitorQueue(monitor, inspector, dependencies.Metrics, logger, "asr")
	}()
	defer func() { stopMonitor(); <-monitorDone }()
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "asr-worker", Logger: logger, Metrics: dependencies.Metrics,
		Tracer: dependencies.Tracing.Tracer("http"),
		Ready: func(request context.Context) error {
			if ctx.Err() != nil || !queueHealthy.Load() {
				return errors.New("worker is not ready")
			}
			if err := modelReady(); err != nil {
				return err
			}
			return dependencies.Ready(request)
		},
		Register: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
				httpserver.WriteJSON(w, http.StatusOK, struct {
					Service  string   `json:"service"`
					Stage    string   `json:"stage"`
					Handlers []string `json:"handlers"`
				}{Service: "asr-worker", Stage: "asr", Handlers: []string{}})
			})
		},
	})
	logger.Info("worker starting", "address", runtime.Address(":8083"))
	return httpserver.Run(ctx, runtime.Address(":8083"), handler)
}

func modelReady() error {
	dir := os.Getenv("WHISPER_MODEL_DIR")
	if dir == "" {
		dir = "/opt/whisper/models"
	}
	info, err := os.Stat(dir)
	if err != nil {
		return errors.New("model weights not available")
	}
	if !info.IsDir() {
		return errors.New("model path is not a directory")
	}
	return nil
}
