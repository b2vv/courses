package jsonbody_test

import (
	"testing"

	"courses/platform/jsonbody"
)

func TestValidateUniqueKeys(t *testing.T) {
	cases := []struct {
		name string
		body string
		ok   bool
	}{
		{"flat object", `{"a":1,"b":2}`, true},
		{"nested object unique", `{"a":{"c":1},"b":[{"d":2}]}`, true},
		{"empty object", `{}`, true},
		{"top-level duplicate", `{"a":1,"a":2}`, false},
		{"nested duplicate", `{"a":{"c":1,"c":2}}`, false},
		{"duplicate inside array element", `{"a":[{"c":1,"c":2}]}`, false},
		{"not an object", `[1,2]`, false},
		{"scalar", `42`, false},
		{"trailing garbage after object", `{"a":1} 7`, false},
		{"second object after first", `{"a":1}{"b":2}`, false},
		{"empty body", ``, false},
		{"unterminated object", `{"a":1`, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := jsonbody.ValidateUniqueKeys([]byte(c.body))
			if c.ok && err != nil {
				t.Fatalf("ValidateUniqueKeys(%s) = %v; want nil", c.body, err)
			}
			if !c.ok && err == nil {
				t.Fatalf("ValidateUniqueKeys(%s) = nil; want error", c.body)
			}
		})
	}
}
