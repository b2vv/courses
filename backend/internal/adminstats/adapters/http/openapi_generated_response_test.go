package http_test

import (
	"bytes"
	"context"
	"courses/internal/adminstats/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"io"
	"testing"
	"time"
)

func TestAdminStatsHandlersMatchGeneratedOpenAPI(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, path string
		empty      bool
		status     int
	}{
		{"snapshot", "/api/v1/admin/stats", false, 200},
		{"snapshot nullable averages and empty array", "/api/v1/admin/stats", true, 200},
		{"timeseries", "/api/v1/admin/stats/timeseries?metric=signups&bucket=day", false, 200},
		{"empty timeseries", "/api/v1/admin/stats/timeseries?metric=orders&bucket=week", true, 200},
		{"invalid range", "/api/v1/admin/stats/timeseries?metric=signups&bucket=day&from=2026-05-01T00:00:00Z&to=2026-01-01T00:00:00Z", false, 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &fakeRepository{}
			if !tc.empty {
				repo.snapshot = sampleSnapshot()
				repo.points = []domain.TimeseriesPoint{{BucketStart: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), Value: 4}}
			}
			req, rec := request(t, newServer(t, repo, true), tc.path, "admin")
			if rec.Code != tc.status {
				t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.status, rec.Body.String())
			}
			route, params, err := router.FindRoute(req)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: req, PathParams: params, Route: route}
			if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: rec.Code, Header: rec.Header(), Body: io.NopCloser(bytes.NewReader(rec.Body.Bytes()))}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
