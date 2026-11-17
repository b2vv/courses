package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	mediaffmpeg "courses/internal/media/adapters/ffmpeg"
	mediapostgres "courses/internal/media/adapters/postgres"
	mediaapplication "courses/internal/media/application"
	"courses/platform/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
)

type mode int

const (
	modeWorker mode = iota
	modeRegenerateMasters
)

// commandMode picks what the binary does. No argument is the long-running worker; the
// one-shot `regenerate-masters` (multi-subtitles MS1 M-D4) rewrites every Ready asset's master
// from the database and exits — the compose service `media-masters` runs it before the worker
// starts, so a migration that changes rendered track metadata reaches the manifests on disk.
func commandMode(arguments []string) (mode, error) {
	switch {
	case len(arguments) == 0:
		return modeWorker, nil
	case len(arguments) == 1 && arguments[0] == "regenerate-masters":
		return modeRegenerateMasters, nil
	default:
		return modeWorker, errors.New("unknown command; expected no argument or regenerate-masters")
	}
}

// regenerate opens the worker's dependencies and rewrites the masters. It deliberately does
// not recover or sweep, start the queue or serve HTTP: it is a one-shot that must terminate.
func regenerate(ctx context.Context, logger *slog.Logger) error {
	dependencies, err := runtime.Open(ctx, "media-masters", logger)
	if err != nil {
		return err
	}
	defer dependencies.Close()
	result, err := regenerateMasters(ctx, dependencies.Database, mediaRoot)
	if err != nil {
		return err
	}
	logger.Info("masters regenerated", "regenerated", result.Regenerated, "skipped", len(result.Skipped), "failed", len(result.Failed))
	if len(result.Failed) > 0 {
		return fmt.Errorf("%d master(s) could not be rewritten: %v", len(result.Failed), result.Failed)
	}
	return nil
}

// regenerateMasters wires the production adapters — the repository's serialised rewrite and
// the ffmpeg master writer over root — exactly as the API's rename path does.
func regenerateMasters(ctx context.Context, database *pgxpool.Pool, root string) (mediaapplication.MasterRegeneration, error) {
	repository := mediapostgres.NewRepository(database)
	tracks := mediaapplication.NewAudioTracks(repository, mediaffmpeg.New(root), false, time.Now)
	return tracks.RegenerateAllMasters(ctx)
}
