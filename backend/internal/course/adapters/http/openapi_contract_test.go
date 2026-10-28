package http_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3"
)

// Catalog fidelity owns its complete surface; the composition package owns the artifact.
func TestCatalogOpenAPIFidelity(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "openapi")
	generated, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), filepath.Join(dir, "openapi.overlay.yaml"), filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	hand, err := openapi3.NewLoader().LoadFromFile(filepath.Join(dir, "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := fidelitySurface(t, normalizedDocument(t, generated))
	source := normalizedDocument(t, hand)
	want := fidelitySurface(t, source)
	if diff := firstDifference("$", got, want); diff != "" {
		t.Error(diff)
		// Report each component independently so one missing metadata block does
		// not hide unrelated shape mismatches in this growing surface.
		for group, entries := range got["components"].(map[string]any) {
			for name, value := range entries.(map[string]any) {
				if d := firstDifference(group+"/"+name, value, want["components"].(map[string]any)[group].(map[string]any)[name]); d != "" {
					t.Error(d)
				}
			}
		}
		for path, value := range got["paths"].(map[string]any) {
			if d := firstDifference(path, value, want["paths"].(map[string]any)[path]); d != "" {
				t.Error(d)
			}
		}
		t.FailNow()
	}
	// Prove that the comparison includes the error enum through the response ref
	// closure. Removing any emitted member must make this exact fidelity gate red.
	schemas := got["components"].(map[string]any)["schemas"].(map[string]any)
	errorSchema := schemas["ErrorResponse"].(map[string]any)
	errorObject := errorSchema["properties"].(map[string]any)["error"].(map[string]any)
	code := errorObject["properties"].(map[string]any)["code"].(map[string]any)
	values := code["enum"].([]any)
	if len(values) == 0 {
		t.Fatal("vacuous error-code fidelity")
	}
	code["enum"] = values[1:]
	if diff := firstDifference("$", got, want); diff == "" {
		t.Fatal("error enum mutation escaped fidelity")
	}
	code["enum"] = values
	t.Logf("error enum ref-closure mutation detected; %d emitted codes compared", len(values))
	lessonProperties := schemas["LearningLesson"].(map[string]any)["properties"].(map[string]any)
	videoState, exists := lessonProperties["videoState"]
	if !exists {
		t.Fatal("vacuous learning fidelity: videoState missing")
	}
	delete(lessonProperties, "videoState")
	if diff := firstDifference("$", got, want); diff == "" {
		t.Fatal("learning videoState mutation escaped fidelity")
	}
	lessonProperties["videoState"] = videoState
	t.Log("learning videoState deletion detected by fidelity")
	spanProperties := schemas["LessonSpan"].(map[string]any)["properties"].(map[string]any)
	spanText, exists := spanProperties["text"]
	if !exists {
		t.Fatal("vacuous document fidelity: LessonSpan.text missing")
	}
	delete(spanProperties, "text")
	if diff := firstDifference("$", got, want); diff == "" {
		t.Fatal("nested document property deletion escaped fidelity")
	}
	spanProperties["text"] = spanText
	t.Log("nested LessonSpan.text deletion detected by fidelity")
	first, err := openapigen.MarshalYAML(generated)
	if err != nil {
		t.Fatal(err)
	}
	again, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), filepath.Join(dir, "openapi.overlay.yaml"), filepath.Join("..", "..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	second, err := openapigen.MarshalYAML(again)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("catalog/learning/document regeneration is not byte-identical")
	}
	t.Log("catalog, learning, document and status request + complete referenced component closures equal hand spec; two generations byte-identical")
}

// Preserve whole catalog fidelity and add the request-only route's complete ref closure.
func fidelitySurface(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	want := map[string]any{"openapi": source["openapi"], "info": source["info"], "servers": source["servers"], "paths": map[string]any{}, "components": map[string]any{}}
	paths := want["paths"].(map[string]any)
	for _, path := range []string{"/api/v1/admin/courses", "/api/v1/admin/courses/{id}", "/api/v1/courses", "/api/v1/courses/{slug}", "/api/v1/learn/courses/{slug}", "/api/v1/admin/lessons/{lessonId}/document", "/api/v1/admin/lessons/{lessonId}/document/publish", "/api/v1/learn/lessons/{lessonId}/document"} {
		value, ok := source["paths"].(map[string]any)[path]
		if !ok || value == nil {
			t.Fatalf("missing path %s", path)
		}
		paths[path] = value
	}
	// Take the transitive reference closure, including responses and headers.
	// Rating is explicitly part of the requested surface even though catalog inlines it.
	needed := map[string]bool{}
	var collect func(any)
	collect = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				if k == "$ref" {
					ref := v.(string)
					if needed[ref] {
						continue
					}
					needed[ref] = true
					parts := strings.Split(strings.TrimPrefix(ref, "#/components/"), "/")
					component := source["components"].(map[string]any)[parts[0]].(map[string]any)[parts[1]]
					if component == nil {
						t.Fatalf("missing component %s", ref)
					}
					group, ok := want["components"].(map[string]any)[parts[0]].(map[string]any)
					if !ok {
						group = map[string]any{}
						want["components"].(map[string]any)[parts[0]] = group
					}
					group[parts[1]] = component
					collect(component)
				} else {
					collect(v)
				}
			}
		case []any:
			for _, v := range x {
				collect(v)
			}
		}
	}
	requestPath := "/api/v1/admin/courses/{id}/status"
	operation := source["paths"].(map[string]any)[requestPath].(map[string]any)["post"].(map[string]any)
	request := operation["requestBody"]
	if request == nil {
		t.Fatal("status request binding is missing")
	}
	paths[requestPath] = map[string]any{"post": map[string]any{"requestBody": request, "parameters": operation["parameters"]}}
	collect(paths)
	collect(map[string]any{"$ref": "#/components/schemas/Rating"})
	return want
}

func normalizedDocument(t *testing.T, doc *openapi3.T) map[string]any {
	t.Helper()
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	var value map[string]any
	if err = json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	var normalize func(any)
	normalize = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				// Parameters are identified by (in, name); generator sorts them.
				if k == "parameters" {
					if a, ok := v.([]any); ok {
						sort.Slice(a, func(i, j int) bool {
							x, y := a[i].(map[string]any), a[j].(map[string]any)
							return fmt.Sprint(x["in"], ":", x["name"]) < fmt.Sprint(y["in"], ":", y["name"])
						})
					}
				}
				if k == "required" {
					if a, ok := v.([]any); ok {
						sort.Slice(a, func(i, j int) bool { return a[i].(string) < a[j].(string) })
					}
				}
				normalize(v)
			}
		case []any:
			for _, v := range x {
				normalize(v)
			}
		}
	}
	normalize(value)
	return value
}
func firstDifference(path string, a, b any) string {
	if reflect.DeepEqual(a, b) {
		return ""
	}
	am, aok := a.(map[string]any)
	bm, bok := b.(map[string]any)
	if aok && bok {
		keys := map[string]bool{}
		for k := range am {
			keys[k] = true
		}
		for k := range bm {
			keys[k] = true
		}
		ordered := make([]string, 0, len(keys))
		for k := range keys {
			ordered = append(ordered, k)
		}
		sort.Strings(ordered)
		for _, k := range ordered {
			if diff := firstDifference(path+"/"+k, am[k], bm[k]); diff != "" {
				return diff
			}
		}
	}
	return fmt.Sprintf("%s: generated=%v; hand=%v", path, a, b)
}

// Keep schema proof independent of path/response metadata so a missing error
// description cannot prevent the per-field non-vacuity check from running.
func TestAdminCourseSchemaFidelity(t *testing.T) {
	generated, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	hand, err := openapi3.NewLoader().LoadFromFile("../../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	got := normalizedDocument(t, generated)["components"].(map[string]any)["schemas"].(map[string]any)
	want := normalizedDocument(t, hand)["components"].(map[string]any)["schemas"].(map[string]any)
	for _, name := range []string{"AdminCoursePage", "AdminCourse", "AdminCourseStatusCounts", "AdminCourseTree", "AdminCourseWrite", "AdminCourseSaveRequest", "AdminCourseDeleteRequest", "AdminModule", "AdminLesson"} {
		t.Run(name, func(t *testing.T) {
			component, ok := got[name].(map[string]any)
			if !ok {
				t.Fatalf("missing admin fidelity component %s", name)
			}
			if diff := firstDifference(name, component, want[name]); diff != "" {
				t.Fatal(diff)
			}
			properties := component["properties"].(map[string]any)
			if len(properties) == 0 {
				t.Fatalf("vacuous admin fidelity %s", name)
			}
			for field, value := range properties {
				delete(properties, field)
				if firstDifference(name, component, want[name]) == "" {
					t.Fatalf("%s.%s deletion escaped fidelity", name, field)
				}
				properties[field] = value
			}
		})
	}
}
