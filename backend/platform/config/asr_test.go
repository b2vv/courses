package config_test

import (
	"strings"
	"testing"

	"courses/platform/config"
)

func TestASRMultilingualDefault(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		values  map[string]string
		want    bool
		invalid bool
	}{
		{name: "unset"},
		{name: "empty", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": ""}},
		{name: "false", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": "false"}},
		{name: "true", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": "true"}, want: true},
		{name: "invalid", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": "yes"}, invalid: true},
		{name: "numeric", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": "1"}, invalid: true},
		{name: "uppercase", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": "TRUE"}, invalid: true},
		{name: "whitespace", values: map[string]string{"ASR_MULTILINGUAL_DEFAULT": " true "}, invalid: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := config.ASRMultilingualDefault(environment(test.values))
			if got != test.want || (err != nil) != test.invalid {
				t.Fatalf("got %v, %v; want %v, invalid=%v", got, err, test.want, test.invalid)
			}
			if err != nil && !strings.Contains(err.Error(), "ASR_MULTILINGUAL_DEFAULT") {
				t.Fatalf("unnamed config error: %v", err)
			}
		})
	}
}
