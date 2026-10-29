package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	mediaasynq "courses/internal/media/adapters/asynq"
	mediafilesystem "courses/internal/media/adapters/filesystem"
	mediapostgres "courses/internal/media/adapters/postgres"
	"courses/internal/media/adapters/tus"
	mediaapplication "courses/internal/media/application"
	"courses/platform/httpserver"
	"courses/platform/observability"
	"courses/platform/runtime"
)

const mediaRoot = "/media"

func main() {
	logger := observability.NewLogger("upload")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := run(ctx, logger); err != nil {
		logger.Error("upload service stopped with an error", "reason", err.Error())
		os.Exit(1)
	}
	logger.Info("shutdown complete")
}

func run(ctx context.Context, logger *slog.Logger) error {
	dependencies, err := runtime.Open(ctx, "upload", logger)
	if err != nil {
		return err
	}
	defer dependencies.Close()
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
	upload := mediaapplication.NewUpload(repository, storage, queue, time.Now)
	server := tus.NewServer(upload)
	handler := httpserver.NewHandler(httpserver.Options{
		Service: "upload", Logger: logger, Metrics: dependencies.Metrics, MetricsToken: os.Getenv("METRICS_TOKEN"),
		Tracer: dependencies.Tracing.Tracer("http"),
		Ready: func(request context.Context) error {
			return errors.Join(dependencies.Ready(request), storage.Ready(request))
		},
		Register: server.Register,
	})
	logger.Info("upload service starting", "address", runtime.Address(":8082"))
	return httpserver.Run(ctx, runtime.Address(":8082"), handler)
}
