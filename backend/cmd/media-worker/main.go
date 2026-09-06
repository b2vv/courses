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

	mediaasynq "courses/internal/media/adapters/asynq"
	mediaffmpeg "courses/internal/media/adapters/ffmpeg"
	mediafilesystem "courses/internal/media/adapters/filesystem"
	mediapostgres "courses/internal/media/adapters/postgres"
	mediaapplication "courses/internal/media/application"
	"courses/platform/httpserver"
	"courses/platform/observability"
	redisinfrastructure "courses/platform/redis"
	"courses/platform/runtime"
	"github.com/hibiken/asynq"
)

const mediaRoot = "/media"

func main() {
	logger := observability.NewLogger("media-worker")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, logger); err != nil {
		logger.Error("worker stopped with an error", "reason", err.Error())
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func run(ctx context.Context, logger *slog.Logger) error {
	dependencies, err := runtime.Open(ctx, "media-worker", logger)
	if err != nil {
		return err
	}
	defer dependencies.Close()
	queueOption, err := redisinfrastructure.QueueOption(dependencies.RedisURL)
	if err != nil {
		return err
	}
	storage, err := mediafilesystem.New(mediaRoot)
	if err != nil {
		return err
	}
	defer func() {
		if err := storage.Close(); err != nil {
			logger.Error("media storage shutdown failed")
		}
	}()
	repository := mediapostgres.NewRepository(dependencies.Database)
	queue := mediaasynq.NewQueue(dependencies.Redis)
	transcode := mediaapplication.NewTranscode(repository, storage, mediaffmpeg.New(mediaRoot), queue, time.Now)
	if err := transcode.Recover(ctx); err != nil {
		return errors.New("media recovery failed")
	}
	if err := transcode.Sweep(ctx); err != nil {
		return errors.New("media sweep failed")
	}
	var queueHealthy atomic.Bool
	queueHealthy.Store(true)
	server := asynq.NewServer(queueOption, asynq.Config{
		Concurrency: 1, Queues: map[string]int{"media": 1}, ShutdownTimeout: 10 * time.Second,
		Logger:              redisinfrastructure.QueueLogger{Logger: logger},
		ErrorHandler:        mediaasynq.NewErrorHandler(logger),
		HealthCheckInterval: 5 * time.Second,
		HealthCheckFunc:     func(err error) { queueHealthy.Store(err == nil) },
	})
	mux := asynq.NewServeMux()
	mux.Handle(mediaasynq.TaskTypeTranscode, mediaasynq.NewHandler(transcode.Process, dependencies.Metrics))
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
		redisinfrastructure.MonitorQueue(monitor, inspector, dependencies.Metrics, logger)
	}()
	defer func() { stopMonitor(); <-monitorDone }()
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "media-worker", Logger: logger, Metrics: dependencies.Metrics,
		Tracer: dependencies.Tracing.Tracer("http"),
		Ready: func(request context.Context) error {
			if ctx.Err() != nil || !queueHealthy.Load() {
				return errors.New("worker is not ready")
			}
			return errors.Join(dependencies.Ready(request), storage.Ready(request))
		},
		Register: func(mux *http.ServeMux) {
			mux.HandleFunc("GET /status", func(w http.ResponseWriter, _ *http.Request) {
				httpserver.WriteJSON(w, http.StatusOK, struct {
					Service  string   `json:"service"`
					Stage    string   `json:"stage"`
					Handlers []string `json:"handlers"`
				}{Service: "media-worker", Stage: "media", Handlers: []string{mediaasynq.TaskTypeTranscode}})
			})
		},
	})
	logger.Info("worker starting", "address", runtime.Address(":8081"), "registeredHandlers", 1)
	return httpserver.Run(ctx, runtime.Address(":8081"), handler)
}
