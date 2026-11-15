package http_test

import (
	"testing"

	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
)

func TestM2R10OpenAPISurface(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	base := "/api/v1/admin/lessons/{lessonId}"
	for _, suffix := range []string{"/draft", "/draft/regenerate", "/draft/publish"} {
		if doc.Paths.Value(base+suffix) != nil {
			t.Fatalf("removed OpenAPI path still present: %s", suffix)
		}
	}
	for _, suffix := range []string{"/drafts", "/transcripts"} {
		path := doc.Paths.Value(base + suffix)
		if path == nil || path.Get == nil || path.Get.Responses.Status(200) == nil {
			t.Fatalf("missing inventory GET: %s", suffix)
		}
	}
	path := doc.Paths.Value(base + "/drafts/{trackId}/regenerate")
	if path == nil || path.Post == nil {
		t.Fatal("missing per-track regenerate")
	}
	if path.Post.RequestBody != nil {
		t.Fatal("regenerate must have no request body schema")
	}
	for _, status := range []int{202, 400, 401, 403, 404, 409, 503} {
		if path.Post.Responses.Status(status) == nil {
			t.Fatalf("missing regenerate status=%d", status)
		}
	}
	draft := doc.Components.Schemas["LessonDraft"].Value
	for _, field := range []string{"trackId", "trackLanguage", "trackMultilingual", "state", "language", "failure", "stale", "updatedAt", "assetId", "generation", "segments"} {
		if draft.Properties[field] == nil {
			t.Fatalf("missing draft field %s", field)
		}
	}
	if doc.Components.Schemas["RegenerateLessonDraftRequest"] != nil {
		t.Fatal("legacy request schema remains")
	}
}
