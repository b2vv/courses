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

	asrasynq "courses/internal/transcript/adapters/asynq"
	asrmedia "courses/internal/transcript/adapters/media"
	asrpostgres "courses/internal/transcript/adapters/postgres"
	asrwhisper "courses/internal/transcript/adapters/whisper"
	"courses/internal/transcript/application"
	"courses/platform/httpserver"
	"courses/platform/observability"
	redisinfrastructure "courses/platform/redis"
	"courses/platform/runtime"
	"github.com/hibiken/asynq"
)

const (
	mediaRoot    = "/media"
	pollInterval = 60 * time.Second
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

	repository := asrpostgres.NewRepository(dependencies.Database)
	queue := asrasynq.NewQueue(dependencies.Redis)
	engine := asrwhisper.NewEngine(
		envOr("WHISPER_PYTHON", "/opt/whisper/bin/python3"),
		envOr("WHISPER_SCRIPT", "/opt/whisper/transcribe.py"),
		envOr("WHISPER_MODEL_DIR", "/opt/whisper/models"),
		logger,
	)
	audio := asrmedia.NewAudio(mediaRoot, os.TempDir())
	recognize := application.NewRecognize(repository, engine, audio, queue, time.Now, logger)
	if err := recognize.Recover(ctx); err != nil {
		return errors.New("asr recovery failed")
	}

	var queueHealthy atomic.Bool
	queueHealthy.Store(true)
	server := asynq.NewServer(queueOption, asynq.Config{
		Concurrency: 1, Queues: map[string]int{"asr": 1}, ShutdownTimeout: 10 * time.Second,
		Logger:              redisinfrastructure.QueueLogger{Logger: logger},
		ErrorHandler:        asrasynq.NewErrorHandler(logger),
		HealthCheckInterval: 5 * time.Second,
		HealthCheckFunc:     func(err error) { queueHealthy.Store(err == nil) },
	})
	mux := asynq.NewServeMux()
	mux.Handle(asrasynq.TaskTypeRecognize, asrasynq.NewHandler(recognize.Process, dependencies.Metrics))
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

	pollCtx, stopPoll := context.WithCancel(ctx)
	pollDone := make(chan struct{})
	go func() {
		defer close(pollDone)
		runPolling(pollCtx, recognize, logger)
	}()
	defer func() { stopPoll(); <-pollDone }()

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
				}{Service: "asr-worker", Stage: "asr", Handlers: []string{asrasynq.TaskTypeRecognize}})
			})
		},
	})
	logger.Info("worker starting", "address", runtime.Address(":8083"), "registeredHandlers", 1)
	return httpserver.Run(ctx, runtime.Address(":8083"), handler)
}

func runPolling(ctx context.Context, recognize *application.Recognize, logger *slog.Logger) {
	if err := recognize.EnqueuePending(ctx); err != nil {
		logger.Error("initial poll failed", "error", err.Error())
	}
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := recognize.EnqueuePending(ctx); err != nil {
				logger.Error("poll failed", "error", err.Error())
			}
		}
	}
}

func modelReady() error {
	dir := envOr("WHISPER_MODEL_DIR", "/opt/whisper/models")
	info, err := os.Stat(dir)
	if err != nil {
		return errors.New("model weights not available")
	}
	if !info.IsDir() {
		return errors.New("model path is not a directory")
	}
	return nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
