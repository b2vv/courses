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
	registry               *prometheus.Registry
	Requests               *prometheus.CounterVec
	Duration               *prometheus.HistogramVec
	DatabaseDuration       *prometheus.HistogramVec
	QueueDepth             *prometheus.GaugeVec
	PlaybackDatabaseChecks prometheus.Counter
	TranscodingJobDuration *prometheus.HistogramVec
	TranscodingFailures    *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	metrics := &Metrics{
		registry:               prometheus.NewRegistry(),
		Requests:               prometheus.NewCounterVec(prometheus.CounterOpts{Name: "courses_http_requests_total", Help: "Completed HTTP requests."}, []string{"method", "route", "status"}),
		Duration:               prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "courses_http_request_duration_seconds", Help: "HTTP request latency.", Buckets: prometheus.DefBuckets}, []string{"method", "route"}),
		DatabaseDuration:       prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "courses_database_query_duration_seconds", Help: "Database query latency without SQL text.", Buckets: prometheus.DefBuckets}, []string{"outcome"}),
		QueueDepth:             prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "courses_queue_depth", Help: "Asynq tasks by state; -1 means inspection failed."}, []string{"queue", "state"}),
		PlaybackDatabaseChecks: prometheus.NewCounter(prometheus.CounterOpts{Name: "courses_playback_authorization_database_checks_total", Help: "Playback authorization checks that reached PostgreSQL."}),
		TranscodingJobDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "courses_transcoding_job_duration_seconds", Help: "Media transcoding job duration.", Buckets: prometheus.ExponentialBuckets(1, 2, 12)}, []string{"outcome"}),
		TranscodingFailures:    prometheus.NewCounterVec(prometheus.CounterOpts{Name: "courses_transcoding_failures_total", Help: "Media transcoding failures by bounded reason."}, []string{"reason"}),
	}
	for _, outcome := range []string{"ready", "failed", "cancelled"} {
		metrics.TranscodingJobDuration.WithLabelValues(outcome)
	}
	for _, reason := range []string{"unsupported_source", "probe_failed", "render_failed", "publish_failed", "storage_unavailable"} {
		metrics.TranscodingFailures.WithLabelValues(reason)
	}
	metrics.registry.MustRegister(metrics.Requests, metrics.Duration, metrics.DatabaseDuration, metrics.QueueDepth, metrics.PlaybackDatabaseChecks, metrics.TranscodingJobDuration, metrics.TranscodingFailures, collectors.NewGoCollector(), collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
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
