package openapigen

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

type example struct {
	ID       string  `json:"id"`
	Label    *string `json:"label"`
	Optional *string `json:"optional,omitempty"`
	Children []child `json:"children"`
	Child    *child  `json:"child"`
}
type child struct {
	Value int `json:"value"`
}
type custom string

func (custom) MarshalJSON() ([]byte, error) { return []byte(`"custom"`), nil }

type pointerCustom struct {
	Value string `json:"value"`
}

func (*pointerCustom) MarshalJSON() ([]byte, error) { return []byte(`"custom"`), nil }

type enum string

func overlay(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "overlay.yaml")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}
func contribution(v any) ModuleContribution {
	return ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/examples", Responses: map[int]Response{200: {Body: v}}}}}
}
func exampleContribution() ModuleContribution {
	c := contribution(example{})
	c.NullableFields = map[Field]bool{{reflect.TypeFor[example](), "Children"}: false}
	return c
}
func TestGenerateTwiceIsByteIdentical(t *testing.T) {
	c := exampleContribution()
	path := overlay(t, "{}")
	first, err := Generate([]ModuleContribution{c}, path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate([]ModuleContribution{c}, path)
	if err != nil {
		t.Fatal(err)
	}
	a, err := MarshalYAML(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := MarshalYAML(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("regeneration changed bytes")
	}
	s := first.Components.Schemas["example"].Value
	if !reflect.DeepEqual(s.Required, []string{"child", "children", "id", "label"}) {
		t.Fatalf("required = %v", s.Required)
	}
	if !s.Properties["label"].Value.Nullable || !s.Properties["optional"].Value.Nullable || !s.Properties["child"].Value.Nullable {
		t.Fatal("pointer nullability lost")
	}
	if s.Properties["children"].Value.Nullable {
		t.Fatal("explicit non-null slice became nullable")
	}
	if first.Components.Schemas["child"].Value.Nullable {
		t.Fatal("pointer polluted canonical component")
	}
	if s.AdditionalProperties.Has == nil || *s.AdditionalProperties.Has {
		t.Fatal("object not closed")
	}
}
func TestFailClosed(t *testing.T) {
	for _, tc := range []struct {
		name        string
		c           ModuleContribution
		patch, want string
	}{
		{"marshaler", contribution(struct {
			Value custom `json:"value"`
		}{}), "{}", "json.Marshaler"},
		{"pointer marshaler", contribution(struct {
			Value pointerCustom `json:"value"`
		}{}), "{}", "json.Marshaler"},
		{"enum", contribution(struct {
			Value enum `json:"value"`
		}{}), "{}", "string enum"},
		{"slice decision", contribution(example{}), "{}", "slice nullability"},
		{"orphan schema", exampleContribution(), "components:\n  schemas:\n    Typo: {description: orphan}\n", "orphan"},
		{"orphan field", exampleContribution(), "components:\n  schemas:\n    example:\n      properties:\n        typo: {description: orphan}\n", "orphan"},
		{"orphan path", exampleContribution(), "paths:\n  /typo:\n    get: {summary: orphan}\n", "orphan"},
		{"unknown keyword", exampleContribution(), "components:\n  schemas:\n    example: {minLenght: 1}\n", "invalid"},
		{"malformed", exampleContribution(), "[broken", "overlay"},
		{"empty", exampleContribution(), "", "overlay"},
		{"multiple documents", exampleContribution(), "{}\n---\n{}", "exactly one"},
		{"duplicate YAML key", exampleContribution(), "info:\n  title: A\n  title: B\n", "overlay"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Generate([]ModuleContribution{tc.c}, overlay(t, tc.patch))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
}
func TestMissingOverlay(t *testing.T) {
	_, err := Generate([]ModuleContribution{exampleContribution()}, filepath.Join(t.TempDir(), "missing"))
	if err == nil {
		t.Fatal("missing overlay accepted")
	}
}
func TestNonInjectiveNameMap(t *testing.T) {
	c := exampleContribution()
	c.NameMap = map[reflect.Type]string{reflect.TypeFor[example](): "Same", reflect.TypeFor[child](): "Same"}
	_, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err == nil || !strings.Contains(err.Error(), "collision") {
		t.Fatalf("got %v", err)
	}
}
func TestRegisteredEnumAndOverride(t *testing.T) {
	c := contribution(struct {
		Value custom `json:"value"`
		State enum   `json:"state"`
	}{})
	c.Enums = map[reflect.Type][]string{reflect.TypeFor[custom](): {"custom"}, reflect.TypeFor[enum](): {"ready", "done"}}
	c.Overrides = map[reflect.Type]*openapi3.Schema{reflect.TypeFor[custom](): openapi3.NewStringSchema()}
	if _, err := Generate([]ModuleContribution{c}, overlay(t, "{}")); err != nil {
		t.Fatal(err)
	}
}
func TestNullableSliceDecision(t *testing.T) {
	c := exampleContribution()
	c.NullableFields[Field{reflect.TypeFor[example](), "Children"}] = true
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	if !doc.Components.Schemas["example"].Value.Properties["children"].Value.Nullable {
		t.Fatal("nullable decision lost")
	}
}

func TestStructMarshalerOverrideIsScalarComponent(t *testing.T) {
	c := contribution(struct {
		Value pointerCustom `json:"value"`
	}{})
	c.Overrides = map[reflect.Type]*openapi3.Schema{reflect.TypeFor[pointerCustom](): openapi3.NewStringSchema()}
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	s := doc.Components.Schemas["pointerCustom"].Value
	if !s.Type.Is("string") || len(s.Properties) != 0 {
		t.Fatalf("override retained reflected shape: %+v", s)
	}
}
func TestUntaggedJSONFieldsAndIgnoredFields(t *testing.T) {
	c := contribution(struct {
		Value    string
		Hidden   custom `json:"-"`
		Optional string `json:",omitempty"`
	}{})
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	s := doc.Paths.Value("/examples").Get.Responses.Value("200").Value.Content["application/json"].Schema.Value
	if !reflect.DeepEqual(s.Required, []string{"Value"}) || len(s.Properties) != 2 {
		t.Fatalf("unexpected fields: %+v", s)
	}
}
