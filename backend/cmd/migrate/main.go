package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"courses/platform/config"
	"courses/platform/observability"
	"github.com/jackc/pgx/v5/pgconn"
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
		attributes := []any{"reason", err.Error()}
		if refusal, ok := authoredRefusal(err); ok {
			attributes = append(attributes, "version", refusal.version, "sqlstate", refusal.sqlState, "message", refusal.message)
		}
		logger.Error("migration failed", attributes...)
		os.Exit(1)
	}
	logger.Info("migrations complete", "applied", count)
}

func migrate(ctx context.Context) (int, error) {
	databaseURL, err := config.PostgresDSN(os.Getenv)
	if err != nil {
		return 0, err
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
		return 0, executionError{cause: err}
	}
	return len(results), nil
}

// executionError keeps the generic text that reaches the log — SQL errors can quote schema and
// data — while still carrying the cause, so authoredRefusal can find a refusal a migration
// raised on purpose.
type executionError struct{ cause error }

func (executionError) Error() string {
	return "database migration execution failed; inspect versioned SQL against the target schema"
}

func (e executionError) Unwrap() error { return e.cause }

type refusal struct {
	version  int64
	sqlState string
	message  string
}

// authoredRefusal returns the message of a migration's own RAISE EXCEPTION (SQLSTATE P0001,
// raise_exception) so the operator sees which rows block it (multi-subtitles MS1 M-D3). Only
// P0001 is surfaced: that is the code a migration raises deliberately. P0001 alone does not prove
// authorship, so the rule for migrations is that a RAISE EXCEPTION message carries only ids and
// counts — never a user-authored value. Every other database error stays behind the generic text.
func authoredRefusal(err error) (refusal, bool) {
	var database *pgconn.PgError
	if !errors.As(err, &database) || database.Code != "P0001" {
		return refusal{}, false
	}
	found := refusal{sqlState: database.Code, message: database.Message}
	var partial *goose.PartialError
	if errors.As(err, &partial) && partial.Failed != nil && partial.Failed.Source != nil {
		found.version = partial.Failed.Source.Version
	}
	return found, true
}
