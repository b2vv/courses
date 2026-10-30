package config

import (
	"encoding/base64"
	"testing"
)

func TestIdentity2FAKey(t *testing.T) {
	valid := base64.StdEncoding.EncodeToString(make([]byte, 32))
	cases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"unset", "", true},
		{"not base64", "@@@not-base64@@@", true},
		{"too short", base64.StdEncoding.EncodeToString(make([]byte, 31)), true},
		{"exactly 32", valid, false},
		{"longer than 32", base64.StdEncoding.EncodeToString(make([]byte, 48)), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			getenv := func(string) string { return tc.value }
			key, err := Identity2FAKey(getenv)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("value %q: expected error, got key of %d bytes", tc.value, len(key))
				}
				return
			}
			if err != nil {
				t.Fatalf("value %q: unexpected error %v", tc.value, err)
			}
			if len(key) < minIdentity2FAKeyBytes {
				t.Fatalf("returned key too short: %d bytes", len(key))
			}
		})
	}
}
