package http_test

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"

	"courses/internal/course/domain"
	identitydomain "courses/internal/identity/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
)

func TestAdminStatusRequestMatchesGeneratedOpenAPI(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), filepath.Join("..", "..", "..", "..", "openapi", "openapi.overlay.yaml"), filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, field string
		omit, valid bool
	}{
		{name: "accepted publish", valid: true},
		{name: "null status", field: "status"},
		{name: "null expectedUpdatedAt", field: "expectedUpdatedAt"},
		{name: "missing status", field: "status", omit: true},
		{name: "missing expectedUpdatedAt", field: "expectedUpdatedAt", omit: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repository := seededAdminWriteRepository(domain.Draft, instructorUserID)
			identity := &adminIdentity{}
			handler := adminReadHandler(t, repository, identity)
			before := repository.courses[adminCourseID]
			body := statusBody(domain.Published, before.UpdatedAt)
			if tc.field != "" {
				if tc.omit {
					delete(body, tc.field)
				} else {
					body[tc.field] = nil
				}
			}
			// The fixture restores exactly the bytes consumed by the real registered handler.
			response, request := serveAdminJSON(t, handler, identity, identityUser(identitydomain.Admin), http.MethodPost, lifecyclePath(), body)
			if tc.valid {
				if response.Code != http.StatusOK || repository.transitionCalls != 1 || repository.courses[adminCourseID].Status != domain.Published {
					t.Fatalf("handler did not apply transition: status=%d calls=%d body=%s", response.Code, repository.transitionCalls, response.Body.String())
				}
			} else if response.Code != http.StatusBadRequest || repository.transitionCalls != 0 {
				t.Fatalf("invalid body reached transition: status=%d calls=%d", response.Code, repository.transitionCalls)
			}
			route, params, err := router.FindRoute(request)
			if err != nil {
				t.Fatal(err)
			}
			err = openapi3filter.ValidateRequest(context.Background(), &openapi3filter.RequestValidationInput{Request: request, PathParams: params, Route: route})
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v ValidateRequest=%v", tc.valid, err)
			}
		})
	}
}
