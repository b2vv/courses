// Package openapigen reflects module-owned transport values into a parallel OpenAPI document.
// It depends on no business module. Contributions and the overlay are explicit inputs.
package openapigen

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"

	"courses/platform/httpserver"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3gen"
	"github.com/oasdiff/yaml"
)

// Field identifies a Go field on its declaring type, including promoted fields.
type Field struct {
	Type reflect.Type
	Name string
}

type Response struct {
	Body any
	// Component optionally names a reusable response (not a schema).
	Component string
	// Empty declares an intentional bodyless response (e.g. 201/202/204). A nil Body without
	// Empty is a forgotten body and fails closed, whatever the status.
	Empty bool
}

// Param declares parameter structure; validation and prose belong to the overlay.
type Param struct {
	Name     string
	In       string
	Type     string
	Enum     []string
	Required bool
}
type RouteBinding struct {
	Method, Path string
	// Request declares a required JSON body; nil means no request binding.
	Request   any
	Responses map[int]Response
	Params    []Param
}
type ModuleContribution struct {
	Routes         []RouteBinding
	Enums          map[reflect.Type][]string
	Overrides      map[reflect.Type]*openapi3.Schema
	NameMap        map[reflect.Type]string
	NullableFields map[Field]bool
	// FieldRefs names reusable scalar schemas for transport fields.
	FieldRefs map[Field]string
	// Components declares non-reflectable schemas/headers, e.g. a nullable wrapper.
	Components openapi3.Components
}

type generator struct {
	errorCodes []string
	config     ModuleContribution
	names      map[string]reflect.Type
	seen       map[reflect.Type]bool
	request    bool
}

// Generate builds JSON request and response bindings. Request-only bindings carry
// a descriptive default response without claiming a generated response schema.
func Generate(contribs []ModuleContribution, overlayPath string) (*openapi3.T, error) {
	return generate(contribs, overlayPath, nil)
}

// GenerateFromSource resolves the complete production error surface before
// generating any schemas. Unknown expressions/provenance prevent generation.
func GenerateFromSource(contribs []ModuleContribution, overlayPath, backendRoot string) (*openapi3.T, error) {
	census, err := ScanErrorCodes(backendRoot, TransportErrorCodes())
	if err != nil {
		return nil, err
	}
	return generate(contribs, overlayPath, census.Codes)
}

func generate(contribs []ModuleContribution, overlayPath string, errorCodes []string) (*openapi3.T, error) {
	g := generator{errorCodes: errorCodes, config: ModuleContribution{Enums: map[reflect.Type][]string{}, Overrides: map[reflect.Type]*openapi3.Schema{}, NameMap: map[reflect.Type]string{}, NullableFields: map[Field]bool{}, FieldRefs: map[Field]string{}}, names: map[string]reflect.Type{}, seen: map[reflect.Type]bool{}}
	doc := &openapi3.T{OpenAPI: "3.0.3", Info: &openapi3.Info{Title: "API", Version: "1.0.0"}, Paths: openapi3.NewPaths(), Components: &openapi3.Components{Schemas: openapi3.Schemas{}, Responses: openapi3.ResponseBodies{}, Headers: openapi3.Headers{}}}
	for _, c := range contribs {
		if err := mergeRegistry(g.config.Enums, c.Enums); err != nil {
			return nil, err
		}
		if err := mergeRegistry(g.config.Overrides, c.Overrides); err != nil {
			return nil, err
		}
		if err := mergeRegistry(g.config.NameMap, c.NameMap); err != nil {
			return nil, err
		}
		if err := mergeRegistry(g.config.NullableFields, c.NullableFields); err != nil {
			return nil, err
		}
		if err := mergeRegistry(g.config.FieldRefs, c.FieldRefs); err != nil {
			return nil, err
		}
		if err := mergeRegistry(doc.Components.Schemas, c.Components.Schemas); err != nil {
			return nil, err
		}
		if err := mergeRegistry(doc.Components.Headers, c.Components.Headers); err != nil {
			return nil, err
		}
		g.config.Routes = append(g.config.Routes, c.Routes...)
	}
	g.config.Components.Schemas = doc.Components.Schemas
	for t, name := range g.config.NameMap {
		if t == nil || name == "" {
			return nil, fmt.Errorf("invalid name mapping %v: %q", t, name)
		}
		if old, ok := g.names[name]; ok && old != t {
			return nil, fmt.Errorf("component name collision %q: %v and %v", name, old, t)
		}
		g.names[name] = t
	}
	sort.Slice(g.config.Routes, func(i, j int) bool {
		a, b := g.config.Routes[i], g.config.Routes[j]
		if a.Path == b.Path {
			return a.Method < b.Method
		}
		return a.Path < b.Path
	})
	for _, r := range g.config.Routes {
		if !strings.HasPrefix(r.Path, "/") || !slices.Contains([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS", "TRACE"}, r.Method) {
			return nil, fmt.Errorf("invalid route %s %s", r.Method, r.Path)
		}
		item := doc.Paths.Value(r.Path)
		if item == nil {
			item = &openapi3.PathItem{}
			doc.Paths.Set(r.Path, item)
		}
		if item.GetOperation(r.Method) != nil {
			return nil, fmt.Errorf("duplicate route %s %s", r.Method, r.Path)
		}
		op := &openapi3.Operation{Responses: openapi3.NewResponses()}
		params := slices.Clone(r.Params)
		sort.Slice(params, func(i, j int) bool { return params[i].Name < params[j].Name })
		for i, p := range params {
			if p.Name == "" || (i > 0 && params[i-1].Name == p.Name) ||
				!slices.Contains([]string{"query", "path"}, p.In) ||
				!slices.Contains([]string{"integer", "string"}, p.Type) ||
				(p.In == "path" && (!p.Required || !strings.Contains(r.Path, "{"+p.Name+"}"))) ||
				(len(p.Enum) > 0 && p.Type != "string") {
				return nil, fmt.Errorf("invalid parameter %q on %s %s", p.Name, r.Method, r.Path)
			}
			schema := &openapi3.Schema{Type: &openapi3.Types{p.Type}}
			for _, value := range p.Enum {
				schema.Enum = append(schema.Enum, value)
			}
			op.Parameters = append(op.Parameters, &openapi3.ParameterRef{Value: &openapi3.Parameter{
				Name: p.Name, In: p.In, Required: p.Required, Schema: &openapi3.SchemaRef{Value: schema},
			}})
		}
		if r.Request != nil {
			ref, err := g.reflectBody(doc, r.Request, true)
			if err != nil {
				return nil, err
			}
			op.RequestBody = &openapi3.RequestBodyRef{Value: &openapi3.RequestBody{Required: true, Content: openapi3.Content{"application/json": &openapi3.MediaType{Schema: ref}}}}
		}
		if len(r.Responses) == 0 {
			if r.Request == nil {
				return nil, fmt.Errorf("route without responses: %s", r.Path)
			}
			description := "Response schemas are not generated for this request-only binding."
			op.Responses.Set("default", &openapi3.ResponseRef{Value: &openapi3.Response{Description: &description}})
		} else {
			op.Responses.Delete("default")
		}
		statuses := make([]int, 0, len(r.Responses))
		for status := range r.Responses {
			statuses = append(statuses, status)
		}
		sort.Ints(statuses)
		for _, status := range statuses {
			if status < 100 || status > 599 {
				return nil, fmt.Errorf("invalid response status %d", status)
			}
			response := r.Responses[status]
			desc := ""
			value := &openapi3.Response{Description: &desc}
			// A bodyless response must be declared explicitly (Empty); a nil Body without Empty is
			// a forgotten body and fails closed, whatever the status.
			switch {
			case response.Empty && response.Body != nil:
				return nil, fmt.Errorf("response %d declares both a body and Empty", status)
			case response.Empty:
				// no content schema
			case response.Body == nil:
				return nil, fmt.Errorf("nil response body for status %d; set Empty for an intentional bodyless response", status)
			default:
				ref, err := g.reflectBody(doc, response.Body, false)
				if err != nil {
					return nil, err
				}
				value.Content = openapi3.Content{"application/json": &openapi3.MediaType{Schema: ref}}
			}
			resp := &openapi3.ResponseRef{Value: value}
			if response.Component != "" {
				if old, ok := doc.Components.Responses[response.Component]; ok && !sameJSON(old, resp) {
					return nil, fmt.Errorf("conflicting response %s", response.Component)
				}
				doc.Components.Responses[response.Component] = resp
				resp = &openapi3.ResponseRef{Ref: "#/components/responses/" + response.Component}
			}
			op.Responses.Set(strconv.Itoa(status), resp)
		}
		item.SetOperation(r.Method, op)
	}
	if err := g.completeOverrides(doc); err != nil {
		return nil, err
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var skeleton map[string]any
	if err = json.Unmarshal(raw, &skeleton); err != nil {
		return nil, err
	}
	merged, err := applyOverlay(skeleton, overlayPath)
	if err != nil {
		return nil, err
	}
	raw, err = json.Marshal(merged)
	if err != nil {
		return nil, err
	}
	loader := openapi3.NewLoader()
	result, err := loader.LoadFromData(raw)
	if err != nil {
		return nil, fmt.Errorf("merged document: %w", err)
	}
	if err = result.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid merged document: %w", err)
	}
	return result, nil
}

// Each body gets a fresh traversal: a prior response must not bypass request checks.
func (g *generator) reflectBody(doc *openapi3.T, body any, request bool) (*openapi3.SchemaRef, error) {
	g.request = request
	g.seen = map[reflect.Type]bool{}
	t := reflect.TypeOf(body)
	if t == nil {
		return nil, fmt.Errorf("nil response body")
	}
	if request && t.Kind() == reflect.Pointer {
		return nil, fmt.Errorf("request root must be a value; declare pointer nullability on fields")
	}
	if err := g.check(t); err != nil {
		return nil, err
	}
	schemas := openapi3.Schemas{}
	ref, err := openapi3gen.NewSchemaRefForValue(body, schemas,
		openapi3gen.CreateComponentSchemas(openapi3gen.ExportComponentSchemasOptions{ExportComponentSchemas: true, ExportTopLevelSchema: true}),
		openapi3gen.CreateTypeNameGenerator(g.name), openapi3gen.UseAllExportedFields(),
		openapi3gen.CreateFieldNameGenerator(func(f reflect.StructField, _ string) string {
			name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
			if name == "" {
				return f.Name
			}
			return name
		}),
		openapi3gen.ThrowErrorOnCycle(), openapi3gen.SchemaCustomizer(g.customize))
	if err != nil {
		return nil, fmt.Errorf("reflect %v: %w", t, err)
	}
	for name, s := range schemas {
		if old, ok := doc.Components.Schemas[name]; ok {
			if !sameJSON(old, s) {
				return nil, fmt.Errorf("conflicting schema %s", name)
			}
		} else {
			doc.Components.Schemas[name] = s
		}
	}
	// The reflector keeps root arrays inline even with an explicit type name.
	// Lift only explicitly named array roots; nested slices retain their shape.
	if name, named := g.config.NameMap[t]; named && ref.Ref == "" && ref.Value != nil && ref.Value.Type.Is("array") {
		if old, exists := doc.Components.Schemas[name]; exists && !sameJSON(old, ref) {
			return nil, fmt.Errorf("conflicting schema %s", name)
		}
		doc.Components.Schemas[name] = ref
		ref = &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
	}
	if t.Kind() == reflect.Pointer {
		ref = &openapi3.SchemaRef{Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{ref}, Nullable: true}}
	}
	return ref, nil
}

func mergeRegistry[K comparable, V any](dst, src map[K]V) error {
	for k, v := range src {
		if _, ok := dst[k]; ok {
			return fmt.Errorf("duplicate contribution key %v", k)
		}
		dst[k] = v
	}
	return nil
}
func sameJSON(a, b any) bool {
	x, e := json.Marshal(a)
	y, f := json.Marshal(b)
	return e == nil && f == nil && string(x) == string(y)
}
func base(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}
func (g *generator) name(t reflect.Type) string {
	if n, ok := g.config.NameMap[t]; ok {
		return n
	}
	return t.Name()
}

type jsonField struct {
	field    reflect.StructField
	owner    reflect.Type
	name     string
	optional bool
}

func fields(t reflect.Type) ([]jsonField, error) {
	var out []jsonField
	names := map[string]bool{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		tag := f.Tag.Get("json")
		name, options, _ := strings.Cut(tag, ",")
		if name == "-" {
			continue
		}
		if f.Anonymous && name == "" {
			if base(f.Type).Kind() != reflect.Struct {
				return nil, fmt.Errorf("unsupported embed %v.%s", t, f.Name)
			}
			nested, err := fields(base(f.Type))
			if err != nil {
				return nil, err
			}
			out = append(out, nested...)
			continue
		}
		if f.PkgPath != "" {
			continue
		}
		if name == "" {
			name = f.Name
		}
		for _, option := range strings.Split(options, ",") {
			if option != "" && option != "omitempty" {
				return nil, fmt.Errorf("unsupported json option %q on %v.%s", option, t, f.Name)
			}
		}
		out = append(out, jsonField{f, t, name, slices.Contains(strings.Split(options, ","), "omitempty")})
	}
	for _, f := range out {
		if names[f.name] {
			return nil, fmt.Errorf("ambiguous json field %v.%s", t, f.name)
		}
		names[f.name] = true
	}
	return out, nil
}
func (g *generator) check(t reflect.Type) error {
	t = base(t)
	if g.seen[t] {
		return nil
	}
	g.seen[t] = true
	override, overridden := g.config.Overrides[t]
	if overridden && override == nil {
		return fmt.Errorf("nil override %v", t)
	}
	marshaler := reflect.TypeFor[json.Marshaler]()
	if (t.Implements(marshaler) || reflect.PointerTo(t).Implements(marshaler)) && !overridden {
		return fmt.Errorf("unregistered json.Marshaler %v", t)
	}
	if t.Kind() == reflect.String && t.PkgPath() != "" {
		if len(g.config.Enums[t]) == 0 {
			return fmt.Errorf("unregistered string enum %v", t)
		}
	}
	if t.Kind() == reflect.Struct {
		name := g.name(t)
		if name != "" {
			if old, ok := g.names[name]; ok && old != t {
				return fmt.Errorf("component name collision %q", name)
			}
			g.names[name] = t
		}
	}
	if overridden {
		return nil
	}
	switch t.Kind() {
	case reflect.Struct:
		fs, err := fields(t)
		if err != nil {
			return err
		}
		for _, f := range fs {
			if g.request && f.field.Type.Kind() == reflect.Pointer {
				if _, ok := g.config.NullableFields[Field{f.owner, f.field.Name}]; !ok {
					return fmt.Errorf("missing request pointer nullability decision: %v.%s", f.owner, f.field.Name)
				}
			}
			if base(f.field.Type).Kind() == reflect.Slice {
				if _, ok := g.config.NullableFields[Field{f.owner, f.field.Name}]; !ok {
					return fmt.Errorf("missing slice nullability decision: %v.%s", f.owner, f.field.Name)
				}
			}
			if err := g.check(f.field.Type); err != nil {
				return err
			}
		}
	case reflect.Slice:
		return g.check(t.Elem())
	case reflect.Map, reflect.Interface, reflect.Func, reflect.Chan, reflect.Array, reflect.UnsafePointer:
		return fmt.Errorf("unsupported type %v: explicit override required", t)
	}
	return nil
}
func (g *generator) customize(_ string, t reflect.Type, _ reflect.StructTag, s *openapi3.Schema) error {
	override, overridden := g.config.Overrides[t]
	if overridden {
		copied, err := copySchema(override)
		if err != nil {
			return err
		}
		copied.Nullable = copied.Nullable || s.Nullable
		*s = *copied
	}
	if values, ok := g.config.Enums[t]; ok {
		for _, value := range values {
			s.Enum = append(s.Enum, value)
		}
	}
	if t == reflect.TypeFor[httpserver.ErrorDetail]() {
		if len(g.errorCodes) == 0 {
			return fmt.Errorf("ErrorDetail requires resolved Go error codes: use GenerateFromSource")
		}
		code := s.Properties["code"]
		if code == nil || code.Value == nil {
			return fmt.Errorf("ErrorDetail.code is missing")
		}
		for _, value := range g.errorCodes {
			code.Value.Enum = append(code.Value.Enum, value)
		}
	}
	if overridden || t.Kind() != reflect.Struct {
		return nil
	}
	fs, err := fields(t)
	if err != nil {
		return err
	}
	closed := false
	s.AdditionalProperties = openapi3.AdditionalProperties{Has: &closed}
	for _, f := range fs {
		if !f.optional {
			s.Required = append(s.Required, f.name)
		}
		prop := s.Properties[f.name]
		if prop == nil {
			return fmt.Errorf("reflection omitted %v.%s", t, f.name)
		}
		if f.field.Type.Kind() == reflect.Pointer {
			nullable := true
			if g.request {
				nullable = g.config.NullableFields[Field{f.owner, f.field.Name}]
			}
			if prop.Ref != "" && strings.HasPrefix(prop.Ref, "#/components/") {
				s.Properties[f.name] = &openapi3.SchemaRef{Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{prop}, Nullable: nullable}}
			} else if prop.Value != nil {
				prop.Value.Nullable = nullable
			}
		}
		if nullable, ok := g.config.NullableFields[Field{f.owner, f.field.Name}]; ok {
			prop = s.Properties[f.name]
			if prop.Value != nil {
				prop.Value.Nullable = nullable || (!g.request && f.field.Type.Kind() == reflect.Pointer)
			}
		}
		if name, ok := g.config.FieldRefs[Field{f.owner, f.field.Name}]; ok {
			target := g.config.Components.Schemas[name]
			prop := s.Properties[f.name]
			if target == nil || target.Value == nil || target.Value.Nullable || prop.Value == nil ||
				!slices.Contains([]reflect.Kind{reflect.String, reflect.Bool, reflect.Int, reflect.Int64, reflect.Float64}, base(f.field.Type).Kind()) ||
				!reflect.DeepEqual(target.Value.Type, prop.Value.Type) {
				return fmt.Errorf("incompatible scalar field reference %v.%s: %s", t, f.name, name)
			}
			ref := &openapi3.SchemaRef{Ref: "#/components/schemas/" + name}
			if s.Properties[f.name].Value != nil && s.Properties[f.name].Value.Nullable {
				ref = &openapi3.SchemaRef{Value: &openapi3.Schema{AllOf: openapi3.SchemaRefs{ref}, Nullable: true}}
			}
			s.Properties[f.name] = ref
		}
	}
	sort.Strings(s.Required)
	return nil
}

// MarshalYAML uses kin-openapi's pinned YAML dependency. JSON map ordering and the
// YAML encoder sort mapping keys; required is a set and is sorted recursively.
// Ordered arrays (allOf, parameters, enum) are never arbitrarily reordered.
func MarshalYAML(doc *openapi3.T) ([]byte, error) {
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var v any
	if err = json.Unmarshal(raw, &v); err != nil {
		return nil, err
	}
	sortRequired(v)
	return yaml.Marshal(v)
}
func sortRequired(v any) {
	switch x := v.(type) {
	case map[string]any:
		for k, child := range x {
			if k == "required" {
				if list, ok := child.([]any); ok {
					sort.Slice(list, func(i, j int) bool { return list[i].(string) < list[j].(string) })
				}
			}
			sortRequired(child)
		}
	case []any:
		for _, child := range x {
			sortRequired(child)
		}
	}
}

// The library exports only component bodies with properties. An overridden struct
// can serialize as a scalar; supply its body when the library emits that ref.
func (g *generator) completeOverrides(doc *openapi3.T) error {
	raw, err := json.Marshal(doc)
	if err != nil {
		return err
	}
	var tree any
	if err = json.Unmarshal(raw, &tree); err != nil {
		return err
	}
	needed := map[string]bool{}
	var walk func(any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, v := range x {
				if k == "$ref" {
					if ref, ok := v.(string); ok {
						needed[ref] = true
					}
				} else {
					walk(v)
				}
			}
		case []any:
			for _, v := range x {
				walk(v)
			}
		}
	}
	walk(tree)
	for t, s := range g.config.Overrides {
		name := g.name(t)
		if doc.Components.Schemas[name] != nil || !needed["#/components/schemas/"+name] {
			continue
		}
		copied, err := copySchema(s)
		if err != nil {
			return err
		}
		doc.Components.Schemas[name] = &openapi3.SchemaRef{Value: copied}
	}
	return nil
}
func copySchema(s *openapi3.Schema) (*openapi3.Schema, error) {
	raw, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	var out openapi3.Schema
	if err = json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	closeObjects(&out)
	return &out, nil
}
func closeObjects(s *openapi3.Schema) {
	if s.Type.Is("object") {
		closed := false
		s.AdditionalProperties = openapi3.AdditionalProperties{Has: &closed}
	}
	for _, p := range s.Properties {
		if p.Value != nil {
			closeObjects(p.Value)
		}
	}
	if s.Items != nil && s.Items.Value != nil {
		closeObjects(s.Items.Value)
	}
	for _, refs := range []openapi3.SchemaRefs{s.AllOf, s.AnyOf, s.OneOf} {
		for _, ref := range refs {
			if ref.Value != nil {
				closeObjects(ref.Value)
			}
		}
	}
}
