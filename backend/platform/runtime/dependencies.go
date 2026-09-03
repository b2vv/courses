package runtime

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"courses/platform/observability"
	"courses/platform/postgres"
	redisinfrastructure "courses/platform/redis"
	"github.com/jackc/pgx/v5/pgxpool"
	redisclient "github.com/redis/go-redis/v9"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Dependencies struct {
	Database *pgxpool.Pool
	Redis    *redisclient.Client
	RedisURL string
	Metrics  *observability.Metrics
	Tracing  *sdktrace.TracerProvider
	Logger   *slog.Logger
}

func Open(ctx context.Context, service string, logger *slog.Logger) (*Dependencies, error) {
	databaseURL, redisURL := os.Getenv("DATABASE_URL"), os.Getenv("REDIS_URL")
	if databaseURL == "" || redisURL == "" {
		return nil, fmt.Errorf("DATABASE_URL and REDIS_URL are required")
	}
	startup, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tracing, err := observability.NewTracing(startup, service)
	if err != nil {
		return nil, fmt.Errorf("tracing initialization failed")
	}
	d := &Dependencies{Metrics: observability.NewMetrics(), Tracing: tracing, Logger: logger, RedisURL: redisURL}
	d.Database, err = postgres.Open(startup, databaseURL, d.Metrics, tracing.Tracer("postgres"))
	if err != nil {
		d.Close()
		return nil, err
	}
	d.Redis, err = redisinfrastructure.Open(startup, redisURL)
	if err != nil {
		d.Close()
		return nil, err
	}
	return d, nil
}

func (d *Dependencies) Ready(ctx context.Context) error {
	return errors.Join(d.Database.Ping(ctx), d.Redis.Ping(ctx).Err())
}

// Close is called only after HTTP and queue workers have drained.
func (d *Dependencies) Close() {
	if d.Redis != nil {
		if err := d.Redis.Close(); err != nil {
			d.Logger.Error("Redis shutdown failed")
		}
	}
	if d.Database != nil {
		d.Database.Close()
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := d.Tracing.Shutdown(ctx); err != nil {
		d.Logger.Error("trace exporter shutdown failed")
	}
}

func Address(fallback string) string {
	if address := os.Getenv("HTTP_ADDR"); address != "" {
		return address
	}
	return fallback
}
