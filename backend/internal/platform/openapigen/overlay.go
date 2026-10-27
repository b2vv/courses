package openapigen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	yaml3 "github.com/oasdiff/yaml3"
)

// Overlay uses recursive map merge, array replacement, and null deletion.
// A replacement $ref replaces the whole schema site (OAS 3.0 forbids ref siblings).
// $ref: null first expands a generated ref/allOf wrapper, then applies metadata:
// this preserves reflected fields/required when the published contract is inline.
func applyOverlay(doc map[string]any, path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("overlay: %w", err)
	}
	dec := yaml3.NewDecoder(bytes.NewReader(data))
	var patch map[string]any
	if err = dec.Decode(&patch); err != nil {
		return nil, fmt.Errorf("overlay: %w", err)
	}
	if patch == nil {
		return nil, fmt.Errorf("overlay must be a mapping")
	}
	var extra any
	if err = dec.Decode(&extra); err != io.EOF {
		return nil, fmt.Errorf("overlay must contain exactly one YAML document")
	}
	// Snapshot refs before edits, so expansion cannot depend on map iteration order.
	raw, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	var original map[string]any
	if err = json.Unmarshal(raw, &original); err != nil {
		return nil, err
	}
	if err = mergeOverlay(doc, patch, original, nil); err != nil {
		return nil, err
	}
	return doc, nil
}
func mergeOverlay(dst, patch, original map[string]any, path []string) error {
	if value, ok := patch["$ref"]; ok {
		if value != nil {
			ref, ok := value.(string)
			if !ok || !strings.HasPrefix(ref, "#/components/") {
				return fmt.Errorf("invalid overlay reference at %v", path)
			}
			if _, err := lookup(original, ref); err != nil {
				return err
			}
			if len(patch) != 1 {
				return fmt.Errorf("reference has siblings at %v", path)
			}
			clear(dst)
			dst["$ref"] = ref
			return nil
		}
		expanded, err := expand(dst, original, map[string]bool{})
		if err != nil {
			return fmt.Errorf("inline at %v: %w", path, err)
		}
		clear(dst)
		for k, v := range expanded {
			dst[k] = v
		}
	}
	keys := make([]string, 0, len(patch))
	for k := range patch {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		value := patch[key]
		if key == "$ref" && value == nil {
			continue
		}
		old, exists := dst[key]
		if key == "parameters" && (len(path) == 2 || len(path) == 3) && path[0] == "paths" {
			if len(path) != 3 {
				return fmt.Errorf("parameter metadata must be operation-scoped at %v", path)
			}
			if err := mergeParamMetadata(old, value, path); err != nil {
				return err
			}
			continue
		}
		if anchored(path) && !exists {
			return fmt.Errorf("orphan overlay key %s/%s", strings.Join(path, "/"), key)
		}
		if value == nil {
			if !exists {
				return fmt.Errorf("orphan overlay deletion %v/%s", path, key)
			}
			delete(dst, key)
			continue
		}
		next := append(append([]string(nil), path...), key)
		if object, ok := value.(map[string]any); ok {
			target, ok := old.(map[string]any)
			if exists && !ok {
				return fmt.Errorf("overlay shape mismatch at %v", next)
			}
			if !ok {
				target = map[string]any{}
				dst[key] = target
			}
			if err := mergeOverlay(target, object, original, next); err != nil {
				return err
			}
		} else {
			if _, ok := old.(map[string]any); ok {
				return fmt.Errorf("overlay shape mismatch at %v", next)
			}
			dst[key] = value
		}
	}
	return nil
}
func anchored(path []string) bool {
	if len(path) == 1 && path[0] == "paths" {
		return true
	}
	if len(path) == 2 && path[0] == "paths" {
		return true
	}
	if len(path) == 2 && path[0] == "components" {
		return true
	}
	if len(path) > 0 && (path[len(path)-1] == "properties" || path[len(path)-1] == "responses") {
		return true
	}
	return false
}
func lookup(doc map[string]any, ref string) (map[string]any, error) {
	if !strings.HasPrefix(ref, "#/") {
		return nil, fmt.Errorf("external reference %s", ref)
	}
	var current any = doc
	for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("unresolved reference %s", ref)
		}
		current = m[strings.ReplaceAll(strings.ReplaceAll(part, "~1", "/"), "~0", "~")]
	}
	m, ok := current.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unresolved reference %s", ref)
	}
	raw, _ := json.Marshal(m)
	var copy map[string]any
	if err := json.Unmarshal(raw, &copy); err != nil {
		return nil, err
	}
	return copy, nil
}
func expand(site, original map[string]any, seen map[string]bool) (map[string]any, error) {
	if ref, ok := site["$ref"].(string); ok {
		if seen[ref] {
			return nil, fmt.Errorf("cyclic inline reference %s", ref)
		}
		seen[ref] = true
		target, err := lookup(original, ref)
		if err != nil {
			return nil, err
		}
		if _, chained := target["$ref"]; chained {
			return expand(target, original, seen)
		}
		return target, nil
	}
	if all, ok := site["allOf"].([]any); ok && len(all) == 1 {
		child, ok := all[0].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid allOf")
		}
		expanded, err := expand(child, original, seen)
		if err != nil {
			return nil, err
		}
		for k, v := range site {
			if k != "allOf" {
				expanded[k] = v
			}
		}
		return expanded, nil
	}
	return nil, fmt.Errorf("expected generated ref or single allOf wrapper")
}

// Parameter metadata is keyed by name within a path/method. Arrays and structural
// keywords are rejected so the overlay cannot replace Go-declared parameters.
func mergeParamMetadata(old, patch any, path []string) error {
	metadata, ok := patch.(map[string]any)
	if !ok {
		return fmt.Errorf("parameter metadata must be a mapping at %v", path)
	}
	params, _ := old.([]any)
	declared := map[string]map[string]any{}
	for _, entry := range params {
		param := entry.(map[string]any)
		declared[param["name"].(string)] = param
	}
	names := make([]string, 0, len(metadata))
	for name := range metadata {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		param, exists := declared[name]
		if !exists {
			return fmt.Errorf("orphan overlay parameter %v/%s", path, name)
		}
		fields, ok := metadata[name].(map[string]any)
		if !ok {
			return fmt.Errorf("invalid parameter metadata %v/%s", path, name)
		}
		for key, value := range fields {
			switch key {
			case "description":
				if _, ok := value.(string); !ok {
					return fmt.Errorf("invalid parameter description %v/%s", path, name)
				}
				param[key] = value
			case "schema":
				validation, ok := value.(map[string]any)
				if !ok {
					return fmt.Errorf("invalid parameter validation %v/%s", path, name)
				}
				schema := param["schema"].(map[string]any)
				for keyword, value := range validation {
					switch keyword {
					case "minimum", "maximum", "default", "minLength", "maxLength", "pattern", "description":
						schema[keyword] = value
					default:
						return fmt.Errorf("unsupported parameter metadata %v/%s/schema/%s", path, name, keyword)
					}
				}
			default:
				return fmt.Errorf("unsupported parameter metadata %v/%s/%s", path, name, key)
			}
		}
	}
	return nil
}
