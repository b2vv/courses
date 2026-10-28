// Command openapi-gen writes the generated OpenAPI artifact from the Go-owned
// module contributions. It is the CLI the `make generate` target runs; the
// TestOpenAPIArtifact test produces the same bytes from the same two functions,
// so `check-generated` reddens on any handler/DTO drift that is not regenerated.
package main

import (
	"flag"
	"fmt"
	"os"

	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
)

func main() {
	overlay := flag.String("overlay", "openapi/openapi.overlay.yaml", "path to the OpenAPI overlay")
	root := flag.String("root", ".", "backend root the generator resolves Go sources against")
	out := flag.String("out", "openapi/openapi.yaml", "path to write the generated artifact")
	flag.Parse()

	if err := generate(*overlay, *root, *out); err != nil {
		fmt.Fprintln(os.Stderr, "openapi-gen:", err)
		os.Exit(1)
	}
}

func generate(overlayPath, backendRoot, outPath string) error {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), overlayPath, backendRoot)
	if err != nil {
		return err
	}
	raw, err := openapigen.MarshalYAML(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(outPath, raw, 0o644)
}
