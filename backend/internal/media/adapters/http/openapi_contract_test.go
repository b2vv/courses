package http_test

import (
	"bytes"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestMediaOpenAPIFidelity(t *testing.T) {
	generate := func() *openapi3.T {
		t.Helper()
		doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
		if err != nil {
			t.Fatal(err)
		}
		return doc
	}
	generated := generate()
	hand, err := openapi3.NewLoader().LoadFromFile("../../../../openapi/openapi.yaml")
	if err != nil {
		t.Fatal(err)
	}
	got, want := fidelitySurface(t, normalizedDocument(t, generated)), fidelitySurface(t, normalizedDocument(t, hand))
	if diff := firstDifference("$", got, want); diff != "" {
		t.Fatal(diff)
	}
	// Prove the list ref closure reaches video members, and the request is included.
	schemas := got["components"].(map[string]any)["schemas"].(map[string]any)
	properties := schemas["MediaVideo"].(map[string]any)["properties"].(map[string]any)
	value, exists := properties["state"]
	if !exists {
		t.Fatal("video state missing before mutation")
	}
	delete(properties, "state")
	if firstDifference("$", got, want) == "" {
		t.Fatal("removed video member escaped fidelity")
	}
	properties["state"] = value
	request := schemas["MediaTicketRequest"].(map[string]any)["properties"].(map[string]any)
	declared, exists := request["declaredBytes"]
	if !exists {
		t.Fatal("request declaredBytes missing before mutation")
	}
	delete(request, "declaredBytes")
	if firstDifference("$", got, want) == "" {
		t.Fatal("removed request member escaped fidelity")
	}
	request["declaredBytes"] = declared
	t.Log("video and request ref-closure member removals detected")
	first, err := openapigen.MarshalYAML(generated)
	if err != nil {
		t.Fatal(err)
	}
	second, err := openapigen.MarshalYAML(generate())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("regeneration is not byte-identical")
	}
	t.Log("all seven media paths and full reference closure equal hand; two generations byte-identical")
}

func fidelitySurface(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	want := map[string]any{"openapi": source["openapi"], "info": source["info"], "servers": source["servers"], "paths": map[string]any{}, "components": map[string]any{}}
	paths := want["paths"].(map[string]any)
	for _, path := range []string{
		"/api/v1/admin/lessons/{lessonId}/video",
		"/api/v1/admin/courses/{courseId}/videos",
		"/api/v1/admin/videos/{videoId}/retry",
		"/api/v1/admin/videos/{videoId}/audio-tracks",
		"/api/v1/admin/videos/{videoId}/audio-tracks/{trackId}",
		"/api/v1/admin/videos/{videoId}/audio-tracks/{trackId}/default",
	} {
		value, ok := source["paths"].(map[string]any)[path]
		if !ok || value == nil {
			t.Fatalf("missing path %s", path)
		}
		paths[path] = value
	}
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

	collect(paths)
	schemas := want["components"].(map[string]any)["schemas"].(map[string]any)
	for _, name := range []string{"MediaTicketRequest", "MediaUploadTicket", "MediaVideo", "CourseVideos", "MediaAudioTrack", "MediaAudioTracks", "MediaAudioRename", "AdminTimestamp", "ErrorResponse"} {
		if schemas[name] == nil {
			t.Fatalf("missing DTO in reference closure: %s", name)
		}
	}
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
				// Parameter order is not semantic; compare every parameter by (in,name).
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
