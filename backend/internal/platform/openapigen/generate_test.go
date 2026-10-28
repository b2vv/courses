package openapigen

import (
	"bytes"
	"encoding/json"
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

// Two differently-shaped Go types mapped to one component name must still fail closed. Since §16.19
// this is caught at emission by the sameJSON guard as a "conflicting schema", not a priori by a
// NameMap-injectivity check; an identical shared schema is now allowed
// (TestSharedComponentNameAllowsIdenticalSchemaOnly).
func TestNonInjectiveNameMap(t *testing.T) {
	c := exampleContribution()
	c.NameMap = map[reflect.Type]string{reflect.TypeFor[example](): "Same", reflect.TypeFor[child](): "Same"}
	_, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err == nil || !strings.Contains(err.Error(), "conflicting schema") {
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

func TestBodylessResponseEmitsNoContent(t *testing.T) {
	// A 202 declares Empty (intentional bodyless on a content-bearing status); a 200 has a body.
	c := ModuleContribution{Routes: []RouteBinding{{Method: "POST", Path: "/ping", Request: child{}, Responses: map[int]Response{
		202: {Empty: true},
		200: {Body: child{}},
	}}}}
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	responses := doc.Paths.Value("/ping").Post.Responses
	noContent := responses.Value("202").Value
	if noContent == nil || len(noContent.Content) != 0 {
		t.Fatalf("202 Empty must carry no content: %+v", noContent)
	}
	if noContent.Description == nil {
		t.Fatal("202 must keep a description")
	}
	if bodied := responses.Value("200").Value; bodied.Content["application/json"] == nil {
		t.Fatal("a bodied response lost its content schema")
	}
	// Fail closed: a nil body without Empty is a forgotten body, whatever the status.
	bad := ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/oops", Responses: map[int]Response{200: {}}}}}
	if _, err := Generate([]ModuleContribution{bad}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "nil response body for status 200") {
		t.Fatalf("nil body without Empty should fail closed, got %v", err)
	}
	// Fail closed: declaring both a body and Empty is contradictory.
	both := ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/both", Responses: map[int]Response{200: {Body: child{}, Empty: true}}}}}
	if _, err := Generate([]ModuleContribution{both}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "Empty together with a body") {
		t.Fatalf("body+Empty should fail closed, got %v", err)
	}
}

func TestOneOfResponseEmitsAlternatives(t *testing.T) {
	// A 200 whose body is exactly one of two named shapes (login: AuthenticatedUser | TwoFactorRequired).
	type authUser struct {
		ID string `json:"id"`
	}
	type twoFactor struct {
		PendingID string `json:"pendingId"`
	}
	names := map[reflect.Type]string{reflect.TypeFor[authUser](): "AuthenticatedUser", reflect.TypeFor[twoFactor](): "TwoFactorRequired"}
	c := ModuleContribution{
		Routes:  []RouteBinding{{Method: "POST", Path: "/login", Responses: map[int]Response{200: {OneOf: []any{authUser{}, twoFactor{}}}}}},
		NameMap: names,
	}
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	schema := doc.Paths.Value("/login").Post.Responses.Value("200").Value.Content["application/json"].Schema.Value
	if len(schema.OneOf) != 2 {
		t.Fatalf("expected two oneOf members, got %+v", schema.OneOf)
	}
	// Named members emit as component refs, in slice order (order is contract-significant).
	if schema.OneOf[0].Ref != "#/components/schemas/AuthenticatedUser" || schema.OneOf[1].Ref != "#/components/schemas/TwoFactorRequired" {
		t.Fatalf("oneOf members/order wrong: %q, %q", schema.OneOf[0].Ref, schema.OneOf[1].Ref)
	}
	if doc.Components.Schemas["AuthenticatedUser"] == nil || doc.Components.Schemas["TwoFactorRequired"] == nil {
		t.Fatal("oneOf member components missing")
	}
	// Fail closed: a single alternative is a Body, not a OneOf.
	one := ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/one", Responses: map[int]Response{200: {OneOf: []any{authUser{}}}}}}, NameMap: names}
	if _, err := Generate([]ModuleContribution{one}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "at least two") {
		t.Fatalf("single-member OneOf should fail closed, got %v", err)
	}
	// Fail closed: Body and OneOf together are contradictory.
	bodyAndOneOf := ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/mix", Responses: map[int]Response{200: {Body: authUser{}, OneOf: []any{authUser{}, twoFactor{}}}}}}, NameMap: names}
	if _, err := Generate([]ModuleContribution{bodyAndOneOf}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "both Body and OneOf") {
		t.Fatalf("Body+OneOf should fail closed, got %v", err)
	}
	// Fail closed: Empty and OneOf together are contradictory.
	emptyAndOneOf := ModuleContribution{Routes: []RouteBinding{{Method: "GET", Path: "/e", Responses: map[int]Response{200: {Empty: true, OneOf: []any{authUser{}, twoFactor{}}}}}}, NameMap: names}
	if _, err := Generate([]ModuleContribution{emptyAndOneOf}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "Empty together with a body") {
		t.Fatalf("Empty+OneOf should fail closed, got %v", err)
	}
}

func TestHeaderParameterComponent(t *testing.T) {
	// A reusable header parameter (e.g. Idempotency-Key) emits a components.parameters entry and the
	// operation references it; a plain header stays inline.
	c := ModuleContribution{Routes: []RouteBinding{{
		Method: "POST", Path: "/buy", Request: child{},
		Params: []Param{
			{Name: "Idempotency-Key", In: "header", Type: "string", Required: true, Component: "IdempotencyKey"},
			{Name: "X-Trace", In: "header", Type: "string"},
		},
		Responses: map[int]Response{200: {Body: child{}}},
	}}}
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	comp := doc.Components.Parameters["IdempotencyKey"]
	if comp == nil || comp.Value == nil || comp.Value.In != "header" || comp.Value.Name != "Idempotency-Key" || !comp.Value.Required {
		t.Fatalf("reusable header parameter not emitted as a component: %+v", comp)
	}
	var sawRef, sawInline bool
	for _, p := range doc.Paths.Value("/buy").Post.Parameters {
		if p.Ref == "#/components/parameters/IdempotencyKey" {
			sawRef = true
		}
		if p.Value != nil && p.Value.Name == "X-Trace" && p.Value.In == "header" {
			sawInline = true
		}
	}
	if !sawRef {
		t.Fatal("operation does not reference the IdempotencyKey parameter component")
	}
	if !sawInline {
		t.Fatal("plain header parameter should stay inline")
	}
	// Fail closed: the same component name reused with a differing definition conflicts.
	bad := ModuleContribution{Routes: []RouteBinding{
		{Method: "GET", Path: "/a", Params: []Param{{Name: "Idempotency-Key", In: "header", Type: "string", Required: true, Component: "IdempotencyKey"}}, Responses: map[int]Response{200: {Body: child{}}}},
		{Method: "GET", Path: "/b", Params: []Param{{Name: "Idempotency-Key", In: "header", Type: "string", Required: false, Component: "IdempotencyKey"}}, Responses: map[int]Response{200: {Body: child{}}}},
	}}
	if _, err := Generate([]ModuleContribution{bad}, overlay(t, "{}")); err == nil || !strings.Contains(err.Error(), "conflicting parameter") {
		t.Fatalf("differing reusable parameter should fail closed, got %v", err)
	}
}

func TestSharedComponentNameAllowsIdenticalSchemaOnly(t *testing.T) {
	type ratingA struct {
		Average float64 `json:"average"`
		Count   int     `json:"count"`
	}
	type ratingB struct {
		Average float64 `json:"average"`
		Count   int     `json:"count"`
	}
	type ratingC struct {
		Average float64 `json:"average"`
		Count   int64   `json:"count"`
	}
	path := overlay(t, "{}")
	// Two distinct Go types, same component name, identical schema: allowed (a shared schema such
	// as Rating that several modules each map their own DTO onto).
	ok := ModuleContribution{
		Routes: []RouteBinding{
			{Method: "GET", Path: "/a", Responses: map[int]Response{200: {Body: ratingA{}}}},
			{Method: "GET", Path: "/b", Responses: map[int]Response{200: {Body: ratingB{}}}},
		},
		NameMap: map[reflect.Type]string{reflect.TypeFor[ratingA](): "Rating", reflect.TypeFor[ratingB](): "Rating"},
	}
	doc, err := Generate([]ModuleContribution{ok}, path)
	if err != nil {
		t.Fatalf("identical shared schema must be allowed: %v", err)
	}
	if doc.Components.Schemas["Rating"] == nil {
		t.Fatal("shared Rating component missing")
	}
	// Same name but a different shape (int64 vs int) must still fail closed.
	bad := ModuleContribution{
		Routes: []RouteBinding{
			{Method: "GET", Path: "/a", Responses: map[int]Response{200: {Body: ratingA{}}}},
			{Method: "GET", Path: "/c", Responses: map[int]Response{200: {Body: ratingC{}}}},
		},
		NameMap: map[reflect.Type]string{reflect.TypeFor[ratingA](): "Rating", reflect.TypeFor[ratingC](): "Rating"},
	}
	if _, err := Generate([]ModuleContribution{bad}, path); err == nil || !strings.Contains(err.Error(), "conflicting schema") {
		t.Fatalf("a differing shared schema must conflict, got %v", err)
	}
}

func paramContribution() ModuleContribution {
	c := contribution(child{})
	c.Routes[0].Path = "/examples/{slug}"
	c.Routes[0].Params = []Param{
		{Name: "sort", In: "query", Type: "string", Enum: []string{"recent", "title"}},
		{Name: "slug", In: "path", Type: "string", Required: true},
		{Name: "limit", In: "query", Type: "integer"},
	}
	return c
}

func TestParametersFromGoWithMetadata(t *testing.T) {
	c := paramContribution()
	patch := overlay(t, `paths:
  /examples/{slug}:
    get:
      parameters:
        limit:
          description: Page size
          schema: {minimum: 1, maximum: 50, default: 20}
        slug:
          schema: {minLength: 1, maxLength: 120, pattern: '^[a-z]+$'}
        sort:
          schema: {default: recent}
`)
	doc, err := Generate([]ModuleContribution{c}, patch)
	if err != nil {
		t.Fatal(err)
	}
	params := doc.Paths.Value("/examples/{slug}").Get.Parameters
	if len(params) != 3 {
		t.Fatalf("parameters = %v", params)
	}
	limit, slug, order := params[0].Value, params[1].Value, params[2].Value
	if limit.Name != "limit" || limit.In != "query" || limit.Required || !limit.Schema.Value.Type.Is("integer") || limit.Description != "Page size" || *limit.Schema.Value.Min != 1 || *limit.Schema.Value.Max != 50 || limit.Schema.Value.Default != float64(20) {
		t.Fatalf("limit lost structure or metadata: %+v", limit)
	}
	if slug.Name != "slug" || slug.In != "path" || !slug.Required || !slug.Schema.Value.Type.Is("string") || slug.Schema.Value.MinLength != 1 || *slug.Schema.Value.MaxLength != 120 || slug.Schema.Value.Pattern != "^[a-z]+$" {
		t.Fatalf("slug lost structure or metadata: %+v", slug)
	}
	if order.Name != "sort" || !reflect.DeepEqual(order.Schema.Value.Enum, []any{"recent", "title"}) || order.Schema.Value.Default != "recent" {
		t.Fatalf("sort lost enum or default: %+v", order)
	}
	// Reordering caller-owned input cannot affect emitted bytes, and generation
	// must not sort the caller's slice in place.
	if c.Routes[0].Params[0].Name != "sort" {
		t.Fatal("mutated caller parameters")
	}
	c.Routes[0].Params[0], c.Routes[0].Params[2] = c.Routes[0].Params[2], c.Routes[0].Params[0]
	again, err := Generate([]ModuleContribution{c}, patch)
	if err != nil {
		t.Fatal(err)
	}
	first, err := MarshalYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalYAML(again)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("parameter order changed generated bytes")
	}
}

func TestParameterMetadataFailsClosed(t *testing.T) {
	for _, tc := range []struct{ name, metadata, want string }{
		{"orphan", "typo: {description: orphan}", "orphan overlay parameter"},
		{"array replacement", "[{name: slug, in: path}]", "must be a mapping"},
		{"delete parameters", "null", "must be a mapping"},
		{"delete parameter", "slug: null", "invalid parameter metadata"},
		{"name", "slug: {name: other}", "unsupported parameter metadata"},
		{"in", "slug: {in: query}", "unsupported parameter metadata"},
		{"required", "slug: {required: false}", "unsupported parameter metadata"},
		{"type", "slug: {schema: {type: integer}}", "unsupported parameter metadata"},
		{"enum", "sort: {schema: {enum: [other]}}", "unsupported parameter metadata"},
		{"ref", "slug: {$ref: '#/components/parameters/other'}", "unsupported parameter metadata"},
		{"unknown validation", "slug: {schema: {minLenght: 1}}", "unsupported parameter metadata"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			patch := "paths:\n  /examples/{slug}:\n    get:\n      parameters:\n        " + tc.metadata + "\n"
			_, err := Generate([]ModuleContribution{paramContribution()}, overlay(t, patch))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("want %q, got %v", tc.want, err)
			}
		})
	}
	t.Run("no Go parameters", func(t *testing.T) {
		_, err := Generate([]ModuleContribution{contribution(child{})}, overlay(t, "paths:\n  /examples:\n    get:\n      parameters:\n        limit: {description: orphan}\n"))
		if err == nil || !strings.Contains(err.Error(), "orphan overlay parameter") {
			t.Fatalf("got %v", err)
		}
	})
}

func TestParameterMetadataToleratesRefSibling(t *testing.T) {
	// An operation mixing an inline path parameter with a $ref header component: enriching the
	// inline one via the overlay must not trip over the ref sibling (which has no inline name).
	c := ModuleContribution{
		Routes: []RouteBinding{{
			Method: "POST", Path: "/courses/{slug}/buy", Request: child{},
			Params: []Param{
				{Name: "slug", In: "path", Type: "string", Required: true},
				{Name: "Idempotency-Key", In: "header", Type: "string", Required: true, Component: "IdempotencyKey"},
			},
			Responses: map[int]Response{201: {Body: child{}}},
		}},
	}
	patch := overlay(t, `paths:
  /courses/{slug}/buy:
    post:
      parameters:
        slug:
          schema: {format: uuid}
`)
	doc, err := Generate([]ModuleContribution{c}, patch)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range doc.Paths.Value("/courses/{slug}/buy").Post.Parameters {
		if p.Value != nil && p.Value.Name == "slug" && p.Value.Schema.Value.Format != "uuid" {
			t.Fatalf("inline parameter metadata not applied alongside a $ref sibling: %+v", p.Value)
		}
	}
	if doc.Components.Parameters["IdempotencyKey"] == nil {
		t.Fatal("ref parameter component missing")
	}
}

func TestInvalidGoParameters(t *testing.T) {
	for _, p := range []Param{
		{In: "query", Type: "string"},
		{Name: "x", In: "cookie", Type: "string"}, // header is now valid (§16.21); cookie is not a supported location
		{Name: "x", In: "header", Type: "object"}, // a supported location with an unsupported type still fails
		{Name: "x", In: "query", Type: "object"},
		{Name: "slug", In: "path", Type: "string"},
		{Name: "absent", In: "path", Type: "string", Required: true},
		{Name: "x", In: "query", Type: "integer", Enum: []string{"one"}},
		{Name: "sort", In: "query", Type: "string"},
	} {
		c := paramContribution()
		c.Routes[0].Params = append(c.Routes[0].Params, p)
		_, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
		if err == nil || !strings.Contains(err.Error(), "invalid parameter") {
			t.Fatalf("parameter %+v: %v", p, err)
		}
	}
}

func TestParameterMetadataDoesNotInterceptResponseField(t *testing.T) {
	c := contribution(struct {
		Parameters string `json:"parameters"`
	}{})
	_, err := Generate([]ModuleContribution{c}, overlay(t, `paths:
  /examples:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                properties:
                  parameters: {description: An ordinary response field}
`))
	if err != nil {
		t.Fatal(err)
	}
}

type requestPointers struct {
	Required *string `json:"required"`
	Optional *string `json:"optional,omitempty"`
	Child    *child  `json:"child"`
}

func TestRequestPointerDecisions(t *testing.T) {
	c := contribution(child{})
	c.Routes = append(c.Routes, RouteBinding{Method: "POST", Path: "/requests", Request: requestPointers{}})
	c.NullableFields = map[Field]bool{
		{reflect.TypeFor[requestPointers](), "Required"}: false,
		{reflect.TypeFor[requestPointers](), "Optional"}: true,
		{reflect.TypeFor[requestPointers](), "Child"}:    false,
	}
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	body := doc.Paths.Value("/requests").Post.RequestBody.Value
	s := body.Content["application/json"].Schema.Value
	if !body.Required || !reflect.DeepEqual(s.Required, []string{"child", "required"}) {
		t.Fatalf("request required semantics: %+v", s)
	}
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{"required":"yes","child":{"value":1}}`, true},
		{`{"required":"yes","optional":null,"child":{"value":1}}`, true},
		{`{"required":"yes","optional":"present","child":{"value":1}}`, true},
		{`{"required":null,"child":{"value":1}}`, false},
		{`{"child":{"value":1}}`, false},
		{`{"required":"yes","child":null}`, false},
	} {
		var v any
		if err := json.Unmarshal([]byte(tc.raw), &v); err != nil {
			t.Fatal(err)
		}
		if err := s.VisitJSON(v); (err == nil) != tc.valid {
			t.Fatalf("%s: valid=%v error=%v", tc.raw, tc.valid, err)
		}
	}
	if !s.Properties["optional"].Value.Nullable || s.Properties["required"].Value.Nullable {
		t.Fatal("explicit nullable decisions lost")
	}
	for field := range c.NullableFields {
		t.Run("missing decision "+field.Name, func(t *testing.T) {
			decision := c.NullableFields[field]
			delete(c.NullableFields, field)
			_, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
			c.NullableFields[field] = decision
			if err == nil || !strings.Contains(err.Error(), "missing request pointer nullability decision") {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestSharedRequestResponseCannotBypassDecisions(t *testing.T) {
	// Response is visited first; a shared type still needs request decisions.
	c := contribution(requestPointers{})
	c.Routes = append(c.Routes, RouteBinding{Method: "POST", Path: "/requests", Request: requestPointers{}})
	_, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err == nil || !strings.Contains(err.Error(), "missing request pointer nullability decision") {
		t.Fatalf("got %v", err)
	}
	c.NullableFields = map[Field]bool{
		{reflect.TypeFor[requestPointers](), "Required"}: false,
		{reflect.TypeFor[requestPointers](), "Optional"}: true,
		{reflect.TypeFor[requestPointers](), "Child"}:    false,
	}
	_, err = Generate([]ModuleContribution{c}, overlay(t, "{}"))
	if err == nil || !strings.Contains(err.Error(), "conflicting schema") {
		t.Fatalf("got %v", err)
	}
}

func TestScalarFieldReference(t *testing.T) {
	c := contribution(child{})
	c.Routes = append(c.Routes, RouteBinding{Method: "POST", Path: "/requests", Request: requestPointers{}})
	c.NullableFields = map[Field]bool{
		{reflect.TypeFor[requestPointers](), "Required"}: false,
		{reflect.TypeFor[requestPointers](), "Optional"}: true,
		{reflect.TypeFor[requestPointers](), "Child"}:    false,
	}
	c.FieldRefs = map[Field]string{{reflect.TypeFor[requestPointers](), "Required"}: "Text"}
	for _, tc := range []struct {
		name   string
		schema *openapi3.Schema
		valid  bool
	}{
		{"matching scalar", openapi3.NewStringSchema(), true},
		{"wrong scalar", openapi3.NewIntegerSchema(), false},
		{"nullable target", openapi3.NewStringSchema().WithNullable(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c.Components.Schemas = openapi3.Schemas{"Text": {Value: tc.schema}}
			doc, err := Generate([]ModuleContribution{c}, overlay(t, "{}"))
			if !tc.valid {
				if err == nil || !strings.Contains(err.Error(), "incompatible scalar field reference") {
					t.Fatalf("got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if doc.Components.Schemas["requestPointers"].Value.Properties["required"].Ref != "#/components/schemas/Text" {
				t.Fatal("field reference lost")
			}
		})
	}
}

func TestParameterFormatMetadata(t *testing.T) {
	c := paramContribution()
	doc, err := Generate([]ModuleContribution{c}, overlay(t, "paths:\n  /examples/{slug}:\n    get:\n      parameters:\n        slug:\n          schema: {format: uuid}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if doc.Paths.Value("/examples/{slug}").Get.Parameters[1].Value.Schema.Value.Format != "uuid" {
		t.Fatal("format lost")
	}
}

func TestNamedRootSlice(t *testing.T) {
	path := overlay(t, "{}")
	inline, err := Generate([]ModuleContribution{contribution([]child{})}, path)
	if err != nil {
		t.Fatal(err)
	}
	inlineSchema := inline.Paths.Value("/examples").Get.Responses.Value("200").Value.Content["application/json"].Schema
	if inlineSchema.Ref != "" || !inlineSchema.Value.Type.Is("array") {
		t.Fatal("baseline is not an inline array")
	}
	c := contribution([]child{})
	c.NameMap = map[reflect.Type]string{reflect.TypeFor[[]child](): "Children"}
	generate := func() *openapi3.T {
		t.Helper()
		doc, err := Generate([]ModuleContribution{c}, path)
		if err != nil {
			t.Fatal(err)
		}
		return doc
	}
	doc := generate()
	root := doc.Paths.Value("/examples").Get.Responses.Value("200").Value.Content["application/json"].Schema
	if root.Ref != "#/components/schemas/Children" {
		t.Fatalf("root ref = %q", root.Ref)
	}
	component := doc.Components.Schemas["Children"]
	if component == nil || !sameJSON(component, inlineSchema) {
		t.Fatal("named component differs from reflected inline array")
	}
	first, err := MarshalYAML(doc)
	if err != nil {
		t.Fatal(err)
	}
	second, err := MarshalYAML(generate())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("named slice regeneration changed bytes")
	}
	// A different contributor owns the same component name with a conflicting shape.
	other := ModuleContribution{Components: openapi3.Components{Schemas: openapi3.Schemas{"Children": {Value: openapi3.NewStringSchema()}}}}
	if _, err := Generate([]ModuleContribution{c, other}, path); err == nil || !strings.Contains(err.Error(), "conflicting schema Children") {
		t.Fatalf("conflicting component accepted: %v", err)
	}
	// Identical declarations remain compatible with the existing sameJSON rule.
	other.Components.Schemas["Children"] = inlineSchema
	if _, err := Generate([]ModuleContribution{c, other}, path); err != nil {
		t.Fatal(err)
	}
}

func TestNamedRootSlicePreservesPointerAndStructRoots(t *testing.T) {
	path := overlay(t, "{}")
	children := []child{}
	for _, tc := range []struct {
		name    string
		body    any
		names   map[reflect.Type]string
		ref     string
		pointer bool
	}{
		{"struct", child{}, nil, "#/components/schemas/child", false},
		{"pointer struct", &child{}, nil, "#/components/schemas/child", true},
		{"pointer named slice", &children, map[reflect.Type]string{reflect.TypeFor[*[]child](): "Children"}, "#/components/schemas/Children", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := contribution(tc.body)
			c.NameMap = tc.names
			doc, err := Generate([]ModuleContribution{c}, path)
			if err != nil {
				t.Fatal(err)
			}
			root := doc.Paths.Value("/examples").Get.Responses.Value("200").Value.Content["application/json"].Schema
			if tc.pointer {
				if root.Value == nil || !root.Value.Nullable || len(root.Value.AllOf) != 1 {
					t.Fatal("nullable pointer wrapper lost")
				}
				root = root.Value.AllOf[0]
			}
			if root.Ref != tc.ref {
				t.Fatalf("root ref=%q want=%q", root.Ref, tc.ref)
			}
		})
	}
}
