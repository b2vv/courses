package http_test

import (
	"courses/internal/course/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"testing"
)

// TestAdminCourseLengthCensus pins the maxLength the overlay publishes for each admin course text
// field to the domain length constant its validators enforce, so the two cannot drift apart
// (§16.22 follow-up, 2026-10-29). The reflected contract carries no maxLength of its own — the
// overlay supplies every value here — and the domain constant is the single source the validators
// (NewCourse, ValidateCourseDraft, ValidateSlug, the node validators) and this test both read.
func TestAdminCourseLengthCensus(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]uint64{
		"title":       domain.MaximumTitle,
		"summary":     domain.MaximumSummary,
		"description": domain.MaximumDescription,
		"slug":        domain.MaximumSlug,
	}
	// Each schema exposes the subset of text fields it carries; the census checks every maxLength
	// present and then requires each field to have been found at least once (non-vacuous).
	seen := map[string]int{}
	for _, name := range []string{"Course", "AdminCourse", "AdminCourseTree", "AdminCourseWrite"} {
		schema := doc.Components.Schemas[name]
		if schema == nil || schema.Value == nil {
			t.Fatalf("schema %s missing", name)
		}
		for field, limit := range want {
			prop := schema.Value.Properties[field]
			if prop == nil || prop.Value == nil {
				continue
			}
			got := prop.Value.MaxLength
			if got == nil {
				t.Errorf("%s.%s has no maxLength; the overlay must publish the domain limit", name, field)
				continue
			}
			if *got != limit {
				t.Errorf("%s.%s maxLength = %d, domain constant = %d", name, field, *got, limit)
			}
			seen[field]++
		}
	}
	for field := range want {
		if seen[field] == 0 {
			t.Errorf("no admin course schema published a maxLength for %q; the census would be vacuous", field)
		}
	}
}
