package main

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestAPIStartupHelper(t *testing.T) {
	if os.Getenv("API_STARTUP_TEST_HELPER") == "1" {
		main()
		os.Exit(0)
	}
}

func TestAPIRejectsInvalidMultilingualBeforeDependencies(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{"invalid", "true", "false"} {
		t.Run(value, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "-test.run=^TestAPIStartupHelper$")
			// Isolated environment: valid preceding startup settings, no ambient credentials or network endpoints.
			cmd.Env = []string{"API_STARTUP_TEST_HELPER=1", "APP_ENV=test", "COOKIE_SECURE=false", "PUBLIC_ORIGIN=http://localhost", "SMTP_HOST=localhost", "MAIL_FROM=test@example.test", "IDENTITY_2FA_SECRET_KEY=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=", "ASR_MULTILINGUAL_DEFAULT=" + value}
			output, err := cmd.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatal(ctx.Err())
			}
			want := "POSTGRES_HOST is required"
			if value == "invalid" {
				want = "ASR_MULTILINGUAL_DEFAULT must be true or false"
			}
			if err == nil || !strings.Contains(string(output), want) {
				t.Fatalf("startup ASR=%s error=%v output=%s; want %q", value, err, output, want)
			}
		})
	}
}
