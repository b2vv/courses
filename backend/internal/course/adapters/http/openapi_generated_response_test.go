package http_test

import (
	"bytes"
	"context"
	"io"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// Catalog Layer B exercises real handler bytes against freshly generated schemas.
// The all-module mux/coverage gate is still a later S0.1 run.
func TestCatalogHandlersMatchGeneratedOpenAPI(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), filepath.Join("..", "..", "..", "..", "openapi", "openapi.overlay.yaml"), filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/courses", "/api/v1/courses?q=empty", "/api/v1/courses/go-foundations", "/api/v1/courses/missing", "/api/v1/courses?limit=0"} {
		t.Run(path, func(t *testing.T) {
			request := httptest.NewRequest("GET", path, nil)
			response := httptest.NewRecorder()
			handler(nil).ServeHTTP(response, request)
			route, params, err := router.FindRoute(request)
			if err != nil {
				t.Fatal(err)
			}
			input := &openapi3filter.RequestValidationInput{Request: request, PathParams: params, Route: route}
			if err := openapi3filter.ValidateResponse(context.Background(), &openapi3filter.ResponseValidationInput{RequestValidationInput: input, Status: response.Code, Header: response.Header(), Body: io.NopCloser(bytes.NewReader(response.Body.Bytes()))}); err != nil {
				t.Fatal(err)
			}
		})
	}
}
