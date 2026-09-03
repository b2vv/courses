package observability

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics belongs to one executable; no default global registry is used.
type Metrics struct {
	registry         *prometheus.Registry
	Requests         *prometheus.CounterVec
	Duration         *prometheus.HistogramVec
	DatabaseDuration *prometheus.HistogramVec
	QueueDepth       *prometheus.GaugeVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		registry:         prometheus.NewRegistry(),
		Requests:         prometheus.NewCounterVec(prometheus.CounterOpts{Name: "courses_http_requests_total", Help: "Completed HTTP requests."}, []string{"method", "route", "status"}),
		Duration:         prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "courses_http_request_duration_seconds", Help: "HTTP request latency.", Buckets: prometheus.DefBuckets}, []string{"method", "route"}),
		DatabaseDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "courses_database_query_duration_seconds", Help: "Database query latency without SQL text.", Buckets: prometheus.DefBuckets}, []string{"outcome"}),
		QueueDepth:       prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "courses_queue_depth", Help: "Asynq tasks by state; -1 means inspection failed."}, []string{"queue", "state"}),
	}
	metrics.registry.MustRegister(metrics.Requests, metrics.Duration, metrics.DatabaseDuration, metrics.QueueDepth, collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	return metrics
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *Metrics) ObserveDatabase(start time.Time, err error) {
	outcome := "success"
	if err != nil {
		outcome = "error"
	}
	m.DatabaseDuration.WithLabelValues(outcome).Observe(time.Since(start).Seconds())
}
