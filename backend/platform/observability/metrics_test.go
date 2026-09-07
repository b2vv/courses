package observability

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMetricsArePreinitialised(t *testing.T) {
	response := httptest.NewRecorder()
	NewMetrics().Handler().ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	body := response.Body.String()
	for _, series := range []string{
		`courses_playback_authorization_database_checks_total 0`,
		`courses_transcoding_job_duration_seconds_count{outcome="ready"} 0`,
		`courses_transcoding_job_duration_seconds_count{outcome="failed"} 0`,
		`courses_transcoding_job_duration_seconds_count{outcome="cancelled"} 0`,
		`courses_transcoding_failures_total{reason="unsupported_source"} 0`,
		`courses_transcoding_failures_total{reason="probe_failed"} 0`,
		`courses_transcoding_failures_total{reason="render_failed"} 0`,
		`courses_transcoding_failures_total{reason="publish_failed"} 0`,
		`courses_transcoding_failures_total{reason="storage_unavailable"} 0`,
	} {
		if !strings.Contains(body, series) {
			t.Errorf("missing zero-value series %q", series)
		}
	}
}
