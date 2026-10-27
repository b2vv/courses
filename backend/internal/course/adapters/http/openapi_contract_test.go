package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"courses/internal/platform/openapigen"
	"github.com/getkin/kin-openapi/openapi3"
)

// Run with UPDATE_OPENAPI_GENERATED=1 to emit the parallel artifact. Normal tests
// require the checked-in artifact to match; neither mode writes the hand spec.
func TestCatalogOpenAPIFidelity(t *testing.T) {
	dir := filepath.Join("..", "..", "..", "..", "openapi")
	generated, err := openapigen.Generate([]openapigen.ModuleContribution{OpenAPIContribution()}, filepath.Join(dir, "openapi.overlay.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	hand, err := openapi3.NewLoader().LoadFromFile(filepath.Join(dir, "openapi.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	got := normalizedDocument(t, generated)
	requestPath := "/api/v1/admin/courses/{id}/status"
	// Only this request-only operation is projected. Keep every generated catalog
	// path/component in the comparison so unexpected schema additions still fail.
	got["paths"].(map[string]any)[requestPath] = fidelitySurface(t, got)["paths"].(map[string]any)[requestPath]
	source := normalizedDocument(t, hand)
	want := fidelitySurface(t, source)
	if diff := firstDifference("$", got, want); diff != "" {
		t.Fatal(diff)
	}
	first, err := openapigen.MarshalYAML(generated)
	if err != nil {
		t.Fatal(err)
	}
	again, err := openapigen.Generate([]openapigen.ModuleContribution{OpenAPIContribution()}, filepath.Join(dir, "openapi.overlay.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := openapigen.MarshalYAML(again)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("catalog regeneration is not byte-identical")
	}
	artifact := filepath.Join(dir, "openapi.generated.yaml")
	if os.Getenv("UPDATE_OPENAPI_GENERATED") == "1" {
		if err = os.WriteFile(artifact, first, 0644); err != nil {
			t.Fatal(err)
		}
	}
	checked, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(checked, first) {
		t.Fatal("parallel artifact is stale; run UPDATE_OPENAPI_GENERATED=1 go test -run TestCatalogOpenAPIFidelity ./internal/course/adapters/http")
	}
	t.Log("catalog and status request + complete referenced component closures equal hand spec; two generations byte-identical")
}

// Preserve whole catalog fidelity and add the request-only route's complete ref closure.
func fidelitySurface(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	want := map[string]any{"openapi": source["openapi"], "info": source["info"], "servers": source["servers"], "paths": map[string]any{}, "components": map[string]any{}}
	paths := want["paths"].(map[string]any)
	for _, path := range []string{"/api/v1/courses", "/api/v1/courses/{slug}"} {
		paths[path] = source["paths"].(map[string]any)[path]
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
