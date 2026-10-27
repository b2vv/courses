package http_test

import (
	coursehttp "courses/internal/course/adapters/http"
	"courses/internal/course/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestConstSetsEmittedAndPublishedNarrowing(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	assertEnum := func(label string, got []any, want []string) {
		t.Helper()
		values := make([]any, len(want))
		for i, v := range want {
			values[i] = v
		}
		if !reflect.DeepEqual(got, values) {
			t.Fatalf("%s enum = %v, want %v", label, got, want)
		}
	}
	for _, name := range []string{"Course", "AdminCourseStatusRequest"} {
		assertEnum(name, doc.Components.Schemas[name].Value.Properties["status"].Value.Enum, enumStrings(domain.CourseStatuses))
	}
	assertEnum("published-only", doc.Components.Schemas["CourseDetail"].Value.Properties["status"].Value.Enum, []string{string(domain.Published)})
	for _, p := range doc.Paths.Value("/api/v1/courses").Get.Parameters {
		if p.Value.Name == "sort" {
			assertEnum("sort", p.Value.Schema.Value.Enum, enumStrings(domain.PublicCatalogSorts))
		}
	}
	// Document routes are not in this slice. Exercise the production enum registry
	// through the generator on a scalar field, without inventing a Block marshaler override.
	c := openapigen.ModuleContribution{Enums: coursehttp.OpenAPIContribution().Enums}
	c.Routes = []openapigen.RouteBinding{{Method: "GET", Path: "/probe", Responses: map[int]openapigen.Response{200: {Body: struct {
		Type domain.BlockType `json:"type"`
	}{}}}}}
	path := filepath.Join(t.TempDir(), "overlay.yaml")
	if err := os.WriteFile(path, []byte("{}"), 0600); err != nil {
		t.Fatal(err)
	}
	probe, err := openapigen.Generate([]openapigen.ModuleContribution{c}, path)
	if err != nil {
		t.Fatal(err)
	}
	schema := probe.Paths.Value("/probe").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	assertEnum("BlockType", schema.Properties["type"].Value.Enum, enumStrings(domain.BlockTypes))
}

func enumStrings[T ~string](values []T) []string {
	result := make([]string, len(values))
	for i, v := range values {
		result[i] = string(v)
	}
	return result
}
