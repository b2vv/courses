package observability

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func NewLogger(service string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})).With("service", service)
}

// NewTracing uses OTLP's standard environment variables when an endpoint is set.
// Without an exporter spans still get valid trace IDs for local correlation.
func NewTracing(ctx context.Context, service string) (*sdktrace.TracerProvider, error) {
	options := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(resource.NewSchemaless(attribute.String("service.name", service))),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(0.1))),
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" || os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != "" {
		exporter, err := otlptracehttp.New(ctx, otlptracehttp.WithTimeout(3*time.Second))
		if err != nil {
			return nil, fmt.Errorf("initialize trace exporter: %w", err)
		}
		options = append(options, sdktrace.WithBatcher(exporter, sdktrace.WithExportTimeout(3*time.Second)))
	}
	return sdktrace.NewTracerProvider(options...), nil
}
