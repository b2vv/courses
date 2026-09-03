package http_test

import (
	"testing"

	identityhttp "courses/internal/identity/adapters/http"
)

func TestIdentityRuntimeConfigurationMatrix(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name, environment, origin string
		secure                    bool
		valid                     bool
	}{
		{name: "loopback development", environment: "development", origin: "http://localhost:8080", valid: true},
		{name: "IPv4 loopback", environment: "test", origin: "http://127.0.0.1:8080", valid: true},
		{name: "IPv6 loopback", environment: "development", origin: "http://[::1]:8080", valid: true},
		{name: "secure production", environment: "production", origin: "https://courses.example.com", secure: true, valid: true},
		{name: "unknown environment", environment: "staging", origin: "https://courses.example.com", secure: true},
		{name: "insecure remote", environment: "development", origin: "http://courses.example.com"},
		{name: "insecure production", environment: "production", origin: "https://courses.example.com"},
		{name: "HTTP production", environment: "production", origin: "http://localhost", secure: true},
		{name: "origin path", environment: "development", origin: "http://localhost/path"},
		{name: "origin credentials", environment: "development", origin: "http://user@localhost"},
		{name: "origin query", environment: "development", origin: "http://localhost?x=1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, err := identityhttp.NewConfig(test.environment, test.origin, test.secure)
			if (err == nil) != test.valid {
				t.Fatalf("valid=%v err=%v", test.valid, err)
			}
		})
	}
}

func TestConfigFromEnvironmentRejectsMissingOrInvalidBoolean(t *testing.T) {
	t.Parallel()
	values := map[string]string{"APP_ENV": "development", "PUBLIC_ORIGIN": "http://localhost:8080", "COOKIE_SECURE": "false"}
	if _, err := identityhttp.ConfigFromEnvironment(func(key string) string { return values[key] }); err != nil {
		t.Fatal(err)
	}
	values["COOKIE_SECURE"] = "sometimes"
	if _, err := identityhttp.ConfigFromEnvironment(func(key string) string { return values[key] }); err == nil {
		t.Fatal("accepted invalid COOKIE_SECURE")
	}
}
