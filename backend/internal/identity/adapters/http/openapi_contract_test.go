package http_test

import (
	"bytes"
	identityhttp "courses/internal/identity/adapters/http"
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

func TestIdentityOpenAPIFidelity(t *testing.T) {
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
	// Prove the reference closure includes nested response and request members.
	schemas := got["components"].(map[string]any)["schemas"].(map[string]any)
	for _, probe := range []struct{ schema, field string }{{"AuthenticatedUser", "twoFactorEnabled"}, {"TwoFactorRequired", "pendingId"}, {"LoginRequest", "rememberMe"}, {"AdminUser", "status"}, {"OptionalEmailRequest", "email"}} {
		properties := schemas[probe.schema].(map[string]any)["properties"].(map[string]any)
		value, exists := properties[probe.field]
		if !exists {
			t.Fatalf("missing probe field %s.%s", probe.schema, probe.field)
		}
		delete(properties, probe.field)
		if firstDifference("$", got, want) == "" {
			t.Fatalf("removed %s.%s escaped fidelity", probe.schema, probe.field)
		}
		properties[probe.field] = value
	}
	// oneOf order is part of the frozen contract, not an unordered set.
	login := got["paths"].(map[string]any)["/api/v1/auth/login"].(map[string]any)["post"].(map[string]any)
	response := login["responses"].(map[string]any)["200"].(map[string]any)
	schema := response["content"].(map[string]any)["application/json"].(map[string]any)["schema"].(map[string]any)
	alternatives := schema["oneOf"].([]any)
	if len(alternatives) != 2 {
		t.Fatal("login must have two alternatives")
	}
	alternatives[0], alternatives[1] = alternatives[1], alternatives[0]
	if firstDifference("$", got, want) == "" {
		t.Fatal("swapped login alternatives escaped fidelity")
	}
	alternatives[0], alternatives[1] = alternatives[1], alternatives[0]
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
	t.Log("all 22 identity operations and full reference closure equal hand; two generations byte-identical")
}

func fidelitySurface(t *testing.T, source map[string]any) map[string]any {
	t.Helper()
	want := map[string]any{"openapi": source["openapi"], "info": source["info"], "servers": source["servers"], "paths": map[string]any{}, "components": map[string]any{}}
	paths := want["paths"].(map[string]any)
	for _, route := range identityhttp.OpenAPIContribution().Routes {
		pathItem, ok := source["paths"].(map[string]any)[route.Path].(map[string]any)
		if !ok {
			t.Fatalf("missing path %s", route.Path)
		}
		method := strings.ToLower(route.Method)
		operation, ok := pathItem[method]
		if !ok {
			t.Fatalf("missing operation %s %s", route.Method, route.Path)
		}
		selected, ok := paths[route.Path].(map[string]any)
		if !ok {
			selected = map[string]any{}
			paths[route.Path] = selected
		}
		selected[method] = operation
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
	for _, name := range []string{"AuthenticatedUser", "Instructor", "AdminUser", "AdminUserPage", "Pagination", "LoginRequest", "TwoFactorRequired", "TwoFactorEnrollment", "RecoveryCodes", "TwoFactorCodeRequest", "CompleteTwoFactorRequest", "DisableTwoFactorRequest", "RegisterRequest", "TokenRequest", "EmailRequest", "OptionalEmailRequest", "ResetPasswordRequest", "ChangePasswordRequest", "NewEmailRequest", "AdminUserStatusRequest", "AdminUserPasswordRequest", "AdminTimestamp", "ErrorResponse"} {
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
