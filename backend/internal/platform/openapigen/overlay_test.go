package openapigen

import (
	"strings"
	"testing"
)

func TestPruneUnreferencedSchemas(t *testing.T) {
	// Used references Nested (transitively); Orphan is referenced by nothing outside its own
	// definition. Only Orphan must be pruned; a one-pass "0 $ref anywhere" scan keeping Nested alive
	// only through Used's reference is exactly what this guards.
	doc := map[string]any{
		"paths": map[string]any{
			"/x": map[string]any{"get": map[string]any{"responses": map[string]any{
				"200": map[string]any{"content": map[string]any{"application/json": map[string]any{
					"schema": map[string]any{"$ref": "#/components/schemas/Used"},
				}}},
			}}},
		},
		"components": map[string]any{"schemas": map[string]any{
			"Used":   map[string]any{"properties": map[string]any{"n": map[string]any{"$ref": "#/components/schemas/Nested"}}},
			"Nested": map[string]any{"type": "string"},
			"Orphan": map[string]any{"type": "object"},
		}},
	}
	pruneUnreferencedSchemas(doc)
	schemas := doc["components"].(map[string]any)["schemas"].(map[string]any)
	if _, ok := schemas["Orphan"]; ok {
		t.Error("unreferenced schema Orphan was not pruned")
	}
	if _, ok := schemas["Used"]; !ok {
		t.Error("path-referenced schema Used was pruned")
	}
	if _, ok := schemas["Nested"]; !ok {
		t.Error("transitively-referenced schema Nested was pruned")
	}
}

func TestOverlayOwnedRequestBody(t *testing.T) {
	c := contribution(child{})
	c.Routes[0].Method = "POST"
	patch := `paths:
  /examples:
    post:
      requestBody:
        required: true
        content:
          multipart/form-data:
            schema:
              type: object
              required: [file]
              properties:
                file: {type: string, format: binary}
`
	doc, err := Generate([]ModuleContribution{c}, overlay(t, patch))
	if err != nil {
		t.Fatal(err)
	}
	body := doc.Paths.Value("/examples").Post.RequestBody
	if body == nil || body.Value == nil || !body.Value.Required {
		t.Fatal("overlay body missing")
	}
	schema := body.Value.Content["multipart/form-data"].Schema.Value
	file := schema.Properties["file"]
	if file == nil || !file.Value.Type.Is("string") || file.Value.Format != "binary" || len(schema.Required) != 1 || schema.Required[0] != "file" {
		t.Fatal("multipart file shape lost")
	}
	for _, tc := range []struct {
		name, patch string
		request     any
	}{
		{name: "response component stays anchored", patch: patch + `components:
  schemas:
    child:
      properties:
        orphan: {type: string}
`},
		{name: "JSON request stays anchored", request: child{}, patch: `paths:
  /examples:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                orphan: {type: string}
`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bounded := contribution(child{})
			bounded.Routes[0].Method = "POST"
			bounded.Routes[0].Request = tc.request
			_, err := Generate([]ModuleContribution{bounded}, overlay(t, tc.patch))
			if err == nil || !strings.Contains(err.Error(), "orphan overlay key") {
				t.Fatalf("want orphan overlay key, got %v", err)
			}
		})
	}
}
