// Package jsonbody validates raw JSON request bodies before they are decoded.
package jsonbody

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// ErrNotUniqueObject reports that the body is not a single JSON object with unique keys at
// every nesting level. Callers map it to their own transport error; this value is never
// surfaced to clients, so one shared sentinel across modules is correct.
var ErrNotUniqueObject = errors.New("jsonbody: body is not a unique-keyed JSON object")

// ValidateUniqueKeys returns nil when body is exactly one well-formed JSON object whose keys
// are unique at every level, and ErrNotUniqueObject otherwise. encoding/json silently keeps
// the last of several duplicate keys, so a body that Decode accepts can still hide a
// collision; this walks the token stream to refuse that before Decode runs. The check is
// recursive on purpose: a duplicate key nested inside the body is refused, not only one at
// the top level.
func ValidateUniqueKeys(body []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	opening, err := decoder.Token()
	if err != nil || opening != json.Delim('{') {
		return ErrNotUniqueObject
	}
	if err := validateObject(decoder); err != nil {
		return err
	}
	// Nothing may follow the object: a second value is as malformed as a duplicate key.
	if decoder.Decode(&struct{}{}) != io.EOF {
		return ErrNotUniqueObject
	}
	return nil
}

func validateObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return ErrNotUniqueObject
		}
		if _, exists := seen[key]; exists {
			return ErrNotUniqueObject
		}
		seen[key] = struct{}{}
		if err := validateValue(decoder); err != nil {
			return err
		}
	}
	closing, err := decoder.Token()
	if err != nil || closing != json.Delim('}') {
		return ErrNotUniqueObject
	}
	return nil
}

func validateValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return ErrNotUniqueObject
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		// A scalar value: nothing to descend into.
		return nil
	}
	switch delimiter {
	case '{':
		return validateObject(decoder)
	case '[':
		for decoder.More() {
			if err := validateValue(decoder); err != nil {
				return err
			}
		}
		closing, err := decoder.Token()
		if err != nil || closing != json.Delim(']') {
			return ErrNotUniqueObject
		}
		return nil
	default:
		return ErrNotUniqueObject
	}
}
