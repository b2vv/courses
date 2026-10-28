package http_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	identityhttp "courses/internal/identity/adapters/http"
	identitydomain "courses/internal/identity/domain"
	"courses/internal/media/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// Layer B validates real serialized responses only. It deliberately does not claim
// conformance between the generated request schema and the bespoke strict decoder.
func TestMediaHandlersMatchGeneratedOpenAPI(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	type responseCase struct {
		name                                        string
		failure                                     error
		status                                      int
		anonymous, student, invalidQuery, emptyList bool
	}
	for _, operation := range []struct {
		name, method, path, body string
		success                  int
	}{
		{"mint", "POST", "/api/v1/admin/lessons/" + mediaLessonID + "/video", `{"declaredBytes":12,"originalName":"lesson.mp4"}`, 201},
		{"list", "GET", "/api/v1/admin/courses/" + mediaCourseID + "/videos", "", 200},
		{"retry", "POST", "/api/v1/admin/videos/" + mediaVideoID + "/retry", "", 200},
	} {
		cases := []responseCase{
			{name: "success", status: operation.success},
			{name: "invalid query", status: 400, invalidQuery: true},
			{name: "anonymous", status: 401, anonymous: true},
			{name: "student", status: 403, student: true},
			{name: "non-owner", status: 403, failure: domain.ErrCourseForbidden},
			{name: "internal", status: 500, failure: errors.New("private dependency detail")},
			{name: "unavailable", status: 503, failure: domain.ErrDependencyUnavailable},
		}
		switch operation.name {
		case "mint":
			cases = append(cases, responseCase{name: "missing lesson", failure: domain.ErrLessonNotFound, status: 404}, responseCase{name: "conflict", failure: domain.ErrUploadInProgress, status: 409})
		case "retry":
			cases = append(cases, responseCase{name: "missing video", failure: domain.ErrVideoNotFound, status: 404}, responseCase{name: "not retryable", failure: domain.ErrVideoNotRetryable, status: 409})
		case "list":
			cases = append(cases, responseCase{name: "empty list", status: 200, emptyList: true})
		}
		for _, tc := range cases {
			t.Run(operation.name+"/"+tc.name, func(t *testing.T) {
				repository := defaultMediaRepository()
				mint := repository.mint
				repository.mint = func(ctx context.Context, actor domain.Actor, id string, size int64, name string, expires, now time.Time) (domain.VideoAsset, error) {
					if tc.failure != nil {
						return domain.VideoAsset{}, tc.failure
					}
					return mint(ctx, actor, id, size, name, expires, now)
				}
				repository.list = func(context.Context, domain.Actor, string) ([]domain.VideoAsset, error) {
					if tc.failure != nil || tc.emptyList {
						return nil, tc.failure
					}
					// Exercise every domain state, optional failureReason both omitted and present.
					assets := make([]domain.VideoAsset, 0, len(domain.States))
					for _, state := range domain.States {
						asset := mediaAsset(mediaVideoID, state, 1)
						if state == domain.Failed {
							asset.FailureReason = domain.RenderFailed
						}
						assets = append(assets, asset)
					}
					return assets, nil
				}
				repository.retry = func(context.Context, domain.Actor, string, time.Duration, time.Time) (domain.VideoAsset, error) {
					return mediaAsset(mediaVideoID, domain.Queued, 1), tc.failure
				}
				user := mediaUser(mediaAdminID, identitydomain.Admin)
				if tc.student {
					user = mediaUser(mediaInstructorID, identitydomain.Student)
				}
				identity := &mediaIdentity{user: user}
				handler := mediaAPIHandler(t, repository, &mediaQueue{}, identity)
				path := operation.path
				if tc.invalidQuery {
					path += "?unexpected=1"
				}
				req := httptest.NewRequest(operation.method, path, strings.NewReader(operation.body))
				if operation.body != "" {
					req.Header.Set("Content-Type", "application/json")
				}
				if operation.method == http.MethodPost {
					req.Header.Set("Origin", "http://localhost:8080")
				}
				if !tc.anonymous {
					req.AddCookie(&http.Cookie{Name: identityhttp.SessionCookieName, Value: strings.Repeat("A", 43)})
				}
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != tc.status {
					t.Fatalf("status=%d want=%d body=%s", rec.Code, tc.status, rec.Body.String())
				}
				if tc.emptyList && rec.Body.String() != "{\"items\":[]}\n" {
					t.Fatalf("empty list=%s", rec.Body.String())
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
}
