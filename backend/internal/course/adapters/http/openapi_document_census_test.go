package http_test

import (
	"courses/internal/course/domain"
	"courses/internal/openapispec"
	"courses/internal/platform/openapigen"
	"encoding/json"
	"reflect"
	"testing"
)

// TestDocumentEnumCensus pins the hand-written lesson-document enums the overlay publishes to the
// domain sets and constant their validators enforce (§16.22 follow-up, 2026-10-29). locale,
// contentType and schemaVersion are plain scalar fields on the document DTO, so the reflector emits
// them without an enum and the overlay is the only place the constraint lives; before the cutover
// the frozen hand spec was its only guard, and that is gone. The census makes the domain the single
// source and fails on any drift between it and the emitted contract.
func TestDocumentEnumCensus(t *testing.T) {
	doc, err := openapigen.GenerateFromSource(openapispec.AllOpenAPIContributions(), "../../../../openapi/openapi.overlay.yaml", "../../../..")
	if err != nil {
		t.Fatal(err)
	}
	assertStringEnum := func(schema, field string, want []string) {
		t.Helper()
		s := doc.Components.Schemas[schema]
		if s == nil || s.Value == nil {
			t.Fatalf("schema %s missing", schema)
		}
		prop := s.Value.Properties[field]
		if prop == nil || prop.Value == nil {
			t.Fatalf("%s.%s missing", schema, field)
		}
		got := make([]string, 0, len(prop.Value.Enum))
		for _, v := range prop.Value.Enum {
			str, ok := v.(string)
			if !ok {
				t.Fatalf("%s.%s enum carries a non-string value %v", schema, field, v)
			}
			got = append(got, str)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("%s.%s enum = %v, domain set = %v", schema, field, got, want)
		}
	}
	assertStringEnum("LessonDocument", "locale", domain.LessonDocumentLocales)
	assertStringEnum("LessonSection", "contentType", domain.ImageContentTypes)

	// schemaVersion is a plain integer field whose single legal value is the current version. Compare
	// through JSON so the enum's numeric representation (float64 after loading) does not matter.
	prop := doc.Components.Schemas["LessonDocument"].Value.Properties["schemaVersion"]
	if prop == nil || prop.Value == nil {
		t.Fatal("LessonDocument.schemaVersion missing")
	}
	gotVersion, err := json.Marshal(prop.Value.Enum)
	if err != nil {
		t.Fatal(err)
	}
	wantVersion, err := json.Marshal([]any{domain.CurrentLessonDocumentSchemaVersion})
	if err != nil {
		t.Fatal(err)
	}
	if string(gotVersion) != string(wantVersion) {
		t.Errorf("schemaVersion enum = %s, domain constant = %s", gotVersion, wantVersion)
	}
}
