package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"courses/platform/observability"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func main() {
	logger := observability.NewLogger("migrate")
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	ctx, deadline := context.WithTimeout(ctx, 2*time.Minute)
	defer deadline()
	count, err := migrate(ctx)
	if err != nil {
		logger.Error("migration failed", "reason", err.Error())
		os.Exit(1)
	}
	logger.Info("migrations complete", "applied", count)
}

func migrate(ctx context.Context) (int, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return 0, fmt.Errorf("DATABASE_URL is required")
	}
	directory := os.Getenv("MIGRATIONS_DIR")
	if directory == "" {
		directory = "migrations"
	}
	if _, err := os.Stat(directory); err != nil {
		return 0, fmt.Errorf("migration directory unavailable")
	}
	files, err := filepath.Glob(filepath.Join(directory, "*.sql"))
	if err != nil {
		return 0, fmt.Errorf("invalid migration directory")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return 0, fmt.Errorf("invalid database configuration")
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	ping, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(ping); err != nil {
		return 0, fmt.Errorf("database unavailable")
	}
	if len(files) == 0 {
		return 0, nil
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS(directory), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		return 0, fmt.Errorf("invalid migration files")
	}
	results, err := provider.Up(ctx)
	if err != nil {
		return 0, fmt.Errorf("database migration execution failed; inspect versioned SQL against the target schema")
	}
	return len(results), nil
}
