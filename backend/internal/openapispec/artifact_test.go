package openapispec

import (
	"bytes"
	"courses/internal/platform/openapigen"
	"os"
	"testing"
)

// TestOpenAPIArtifact is the sole writer of the authoritative openapi.yaml.
func TestOpenAPIArtifact(t *testing.T) {
	generate := func() []byte {
		t.Helper()
		doc, err := openapigen.GenerateFromSource(AllOpenAPIContributions(), "../../openapi/openapi.overlay.yaml", "../..")
		if err != nil {
			t.Fatal(err)
		}
		for _, contribution := range AllOpenAPIContributions() {
			for _, route := range contribution.Routes {
				p := doc.Paths.Value(route.Path)
				if p == nil || p.GetOperation(route.Method) == nil {
					t.Fatalf("missing contributed route %s %s", route.Method, route.Path)
				}
			}
		}
		raw, err := openapigen.MarshalYAML(doc)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	first, second := generate(), generate()
	if !bytes.Equal(first, second) {
		t.Fatal("regeneration is not byte-identical")
	}
	const artifact = "../../openapi/openapi.yaml"
	if os.Getenv("UPDATE_OPENAPI_GENERATED") == "1" {
		if err := os.WriteFile(artifact, first, 0644); err != nil {
			t.Fatal(err)
		}
	}
	checked, err := os.ReadFile(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(checked, first) {
		t.Fatal("openapi.yaml stale; run UPDATE_OPENAPI_GENERATED=1 go test -run TestOpenAPIArtifact ./internal/openapispec")
	}
	t.Log("all contributions present; two generations byte-identical; checked artifact matches")
}
