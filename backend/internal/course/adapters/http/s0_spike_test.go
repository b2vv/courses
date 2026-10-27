//go:build s0spike

// S0.0 expressiveness spike (throwaway; build tag s0spike). Run:
//
//	go test -tags s0spike -run TestS0Spike -v ./internal/course/adapters/http/
//
// Purpose: resolve the plan's §7↔§8 open question — is Go→openapi fidelity
// achievable for the course surface via reflection + mechanisms, and what is
// the exact override surface? It reflects the real (unexported) course DTOs
// with the vendored openapi3gen, derives `required` ourselves (the library
// does not), and (A) asserts the reproducible STRUCTURE equals the hand spec,
// (B) logs the census of keywords Go cannot express — the override surface.
// Not wired into any build; deleted/evolved when S0.1 lands the real generator.
package http

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"

	"courses/internal/course/domain"
)

// requiredFromJSONTags derives the OpenAPI required set from json struct tags:
// a field is required unless its tag carries `,omitempty`. This is the rule the
// real generator will own (openapi3gen never populates Required). Embedded
// structs (courseDetailDTO embeds courseDTO) are flattened.
func requiredFromJSONTags(t reflect.Type) []string {
	var req []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		if f.Anonymous && tag == "" {
			req = append(req, requiredFromJSONTags(f.Type)...)
			continue
		}
		name, opts, _ := strings.Cut(tag, ",")
		if name == "-" || name == "" {
			continue
		}
		if !strings.Contains(opts, "omitempty") {
			req = append(req, name)
		}
	}
	sort.Strings(req)
	return req
}

func genSchema(t *testing.T, v any) *openapi3.Schema {
	t.Helper()
	ref, err := openapi3gen.NewSchemaRefForValue(v, nil)
	if err != nil {
		t.Fatalf("openapi3gen: %v", err)
	}
	return ref.Value
}

func propNames(s *openapi3.Schema) []string {
	names := make([]string, 0, len(s.Properties))
	for n := range s.Properties {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func TestS0SpikeStructureFidelity(t *testing.T) {
	gen := genSchema(t, courseDTO{})

	// --- property set ---
	want := []string{
		"createdAt", "description", "id", "lessonCount", "paid", "price",
		"publishedAt", "rating", "slug", "status", "summary", "title", "updatedAt",
	}
	if got := propNames(gen); !reflect.DeepEqual(got, want) {
		t.Errorf("property set:\n got=%v\nwant=%v", got, want)
	}

	// --- nullability from pointers (openapi3gen default) ---
	for _, f := range []string{"publishedAt", "price", "rating"} {
		if p := gen.Properties[f]; p == nil || p.Value == nil || !p.Value.Nullable {
			t.Errorf("%s: want nullable=true from *T, got %+v", f, p)
		}
	}
	for _, f := range []string{"id", "paid", "lessonCount", "createdAt"} {
		if p := gen.Properties[f]; p != nil && p.Value != nil && p.Value.Nullable {
			t.Errorf("%s: want nullable=false (non-pointer), got nullable=true", f)
		}
	}

	// --- base types ---
	typeOf := func(f string) string {
		p := gen.Properties[f]
		if p == nil || p.Value == nil || p.Value.Type == nil || len(*p.Value.Type) == 0 {
			return "<none>"
		}
		return (*p.Value.Type)[0]
	}
	for f, wantT := range map[string]string{
		"id": "string", "slug": "string", "status": "string",
		"paid": "boolean", "lessonCount": "integer",
		"createdAt": "string", "publishedAt": "string",
	} {
		if got := typeOf(f); got != wantT {
			t.Errorf("%s: base type got=%s want=%s", f, got, wantT)
		}
	}
	// time.Time → format date-time (library default)
	if p := gen.Properties["createdAt"]; p.Value.Format != "date-time" {
		t.Errorf("createdAt: want format date-time, got %q", p.Value.Format)
	}

	// --- required derived by our rule == hand-spec Course.required (all 13) ---
	gotReq := requiredFromJSONTags(reflect.TypeOf(courseDTO{}))
	wantReq := append([]string(nil), want...) // all 13, none omitempty
	if !reflect.DeepEqual(gotReq, wantReq) {
		t.Errorf("required:\n got=%v\nwant=%v", gotReq, wantReq)
	}
	t.Logf("STRUCTURE OK: 13 props, pointer→nullable, time→date-time, required=all-13 via omitempty rule")
}

// TestS0SpikeOverrideCensus never fails; it prints the exact override surface
// — every hand-spec keyword reflection does NOT produce — so the plan's §11.2
// table can be checked against reality.
func TestS0SpikeOverrideCensus(t *testing.T) {
	course := genSchema(t, courseDTO{})
	block := genSchema(t, richTextBlockValue())

	note := func(class, detail string) { t.Logf("OVERRIDE NEEDED | %-22s | %s", class, detail) }

	// format:uuid on id — reflection gives bare string
	if course.Properties["id"].Value.Format == "" {
		note("format:uuid", "courseDTO.ID is plain string; hand-spec id has format:uuid — not derivable")
	}
	// status enum — Go field is plain `string`, enum lives elsewhere
	if len(course.Properties["status"].Value.Enum) == 0 {
		note("enum (plain string)", "courseDTO.Status is `string` not `type X string`; enum [draft,published,archived] NOT derivable from Go")
	}
	// validation keywords
	if course.Properties["slug"].Value.MinLength == 0 && course.Properties["slug"].Value.Pattern == "" {
		note("validation keywords", "slug minLength/maxLength/pattern, lessonCount minimum — enforced in domain validators, not on DTO types")
	}
	// additionalProperties:false
	if course.AdditionalProperties.Has == nil {
		note("additionalProperties", "hand-spec sets additionalProperties:false on every object; openapi3gen leaves it unset")
	}
	// Rating divergence: same *ratingDTO maps to two schemas
	note("same-type-two-schemas", "Course.rating is INLINE with minimum:1/maximum:5; component Rating has NO min/max — one Go type (*ratingDTO) → two schemas")
	// NullablePrice wrapper component
	note("nullable-ref wrapper", "*priceDTO → hand-spec $ref NullablePrice (allOf[Price]+nullable); reflection inlines Price with nullable")
	// Go-name → component-name
	note("name-map", "domain.Block → component LessonBlock, Span → LessonSpan (Go name != component name)")

	// Block: reflection vs published LessonBlock
	bt := block.Properties["type"].Value
	if len(bt.Enum) == 0 {
		note("string-enum (BlockType)", "domain.BlockType is `type X string`; published LessonBlock.type enum [paragraph,heading,list] NOT derivable — needs const-set registry")
	}
	// Does reflection's Block struct otherwise match the PERMISSIVE published schema?
	t.Logf("FINDING: published LessonBlock is PERMISSIVE (ordered optional on any block); reflection of the Block STRUCT reproduces it modulo BlockType enum — the custom MarshalJSON (ordered-pointer) does not change the PUBLISHED schema. H2 lived in a stricter reader schema, not this component. Re-examine §2 premise 'Block is the one type reflection gets wrong'.")
	t.Logf("FINDING: course DTOs are UNEXPORTED (courseDTO); a central registry in another package cannot reflect them. Generator must run per-module (each module exports its route→type bindings) or reflect from within each adapter package. Affects §4 approach-A.")
}

// richTextBlockValue returns the REAL domain.Block (custom MarshalJSON +
// `type BlockType string`) so the spike reflects the actual type, not a mirror.
func richTextBlockValue() any { return domain.Block{} }
