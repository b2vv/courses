package http_test

import (
	"bytes"
	"context"
	imagehttp "courses/internal/image/adapters/http"
	"courses/internal/image/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"courses/platform/httpserver"
	"errors"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestImageHandlersMatchGeneratedOpenAPI(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	var fixture bytes.Buffer
	if err := png.Encode(&fixture, image.NewRGBA(image.Rect(0, 0, 1, 1))); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name                      string
		data                      []byte
		failure                   error
		missingFile, missingActor bool
		status                    int
	}{
		{name: "created multipart", data: fixture.Bytes(), status: 201},
		{name: "missing file", missingFile: true, status: 400},
		{name: "missing actor", missingActor: true, status: 403},
		{name: "lesson missing", data: fixture.Bytes(), failure: domain.ErrLessonNotFound, status: 404},
		{name: "too large", data: make([]byte, domain.MaximumImageBytes+(512<<10)), status: 413},
		{name: "unsupported", data: []byte("GIF89a............"), status: 415},
		{name: "dependency unavailable", data: fixture.Bytes(), failure: errors.New("unavailable"), status: 503},
	} {
		t.Run(tc.name, func(t *testing.T) {
			uploader := &fakeUploader{image: domain.Image{ID: "123e4567-e89b-42d3-a456-426614174000", ContentType: domain.ContentTypePNG, Width: 1, Height: 1}, err: tc.failure}
			adapter := imagehttp.NewHandler(uploader, &fakeAuthorizer{})
			authenticate := authenticated(httpserver.Actor{UserID: "admin-1", Role: "admin", EmailVerified: true})
			if tc.missingActor {
				authenticate = passThrough
			}
			actorMiddleware := authenticate
			authenticate = func(next http.Handler) http.Handler {
				downstream := actorMiddleware(next)
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Match identity Authenticate's response header, as in the adminstats harness.
					w.Header().Set("Cache-Control", "no-store")
					downstream.ServeHTTP(w, r)
				})
			}
			handler := httpserver.NewHandler(httpserver.Options{Service: "api", Register: func(mux *http.ServeMux) { adapter.RegisterAdmin(mux, authenticate, passThrough) }})
			req := uploadRequest(t, tc.data, tc.missingFile)
			req.URL.Path = "/api/v1/admin/lessons/123e4567-e89b-42d3-a456-426614174001/images"
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
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
