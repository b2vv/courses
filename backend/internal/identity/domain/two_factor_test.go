package domain_test

import (
	"errors"
	"regexp"
	"testing"
	"time"

	"courses/internal/identity/domain"
)

const rfc6238Secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

func TestVerifyTOTP(t *testing.T) {
	t.Parallel()
	now := time.Unix(59, 0).UTC() // RFC 6238, SHA-1, 6 rightmost digits: 287082.
	for _, test := range []struct {
		name   string
		secret string
		code   string
		at     time.Time
		want   bool
	}{
		{name: "current step", secret: rfc6238Secret, code: "287082", at: now, want: true},
		{name: "previous step in window", secret: rfc6238Secret, code: "755224", at: now, want: true},
		{name: "next step in window", secret: rfc6238Secret, code: "359152", at: now, want: true},
		{name: "outside window", secret: rfc6238Secret, code: "969429", at: now, want: false},
		{name: "wrong six digits", secret: rfc6238Secret, code: "287083", at: now, want: false},
		{name: "malformed code", secret: rfc6238Secret, code: "28708", at: now, want: false},
		{name: "non numeric code", secret: rfc6238Secret, code: "28708x", at: now, want: false},
		{name: "malformed secret", secret: "not base32", code: "287082", at: now, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := domain.VerifyTOTP(test.secret, test.code, test.at); got != test.want {
				t.Fatalf("VerifyTOTP() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestVerifyTOTPAt(t *testing.T) {
	t.Parallel()
	now := time.Unix(59, 0).UTC() // step = 59/30 = 1; codes from RFC 6238 SHA-1 vectors.
	for _, test := range []struct {
		name        string
		code        string
		wantCounter int64
		wantOK      bool
	}{
		{name: "current step reports its counter", code: "287082", wantCounter: 1, wantOK: true},
		{name: "previous step reports its counter", code: "755224", wantCounter: 0, wantOK: true},
		{name: "next step reports its counter", code: "359152", wantCounter: 2, wantOK: true},
		{name: "outside window", code: "969429", wantCounter: 0, wantOK: false},
		{name: "wrong code", code: "287083", wantCounter: 0, wantOK: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			counter, ok := domain.VerifyTOTPAt(rfc6238Secret, test.code, now)
			if ok != test.wantOK || counter != test.wantCounter {
				t.Fatalf("VerifyTOTPAt() = (%d, %t), want (%d, %t)", counter, ok, test.wantCounter, test.wantOK)
			}
		})
	}
}

func TestTOTPURI(t *testing.T) {
	t.Parallel()
	got, err := domain.TOTPURI("CourseForge", "student@example.test", rfc6238Secret)
	if err != nil {
		t.Fatal(err)
	}
	const want = "otpauth://totp/CourseForge:student@example.test?algorithm=SHA1&digits=6&issuer=CourseForge&period=30&secret=GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	if got != want {
		t.Fatalf("TOTPURI() = %q, want %q", got, want)
	}
	for _, test := range []struct {
		name    string
		issuer  string
		account string
		secret  string
	}{
		{name: "blank issuer", issuer: "", account: "student@example.test", secret: rfc6238Secret},
		{name: "blank account", issuer: "CourseForge", account: "", secret: rfc6238Secret},
		{name: "invalid secret", issuer: "CourseForge", account: "student@example.test", secret: "bad secret"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if _, err := domain.TOTPURI(test.issuer, test.account, test.secret); !errors.Is(err, domain.ErrInvalidTOTP) {
				t.Fatalf("TOTPURI() error = %v, want %v", err, domain.ErrInvalidTOTP)
			}
		})
	}
}

func TestGenerateTOTPSecret(t *testing.T) {
	t.Parallel()
	first, err := domain.GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	second, err := domain.GenerateTOTPSecret()
	if err != nil {
		t.Fatal(err)
	}
	if first == second || !regexp.MustCompile(`^[A-Z2-7]+$`).MatchString(first) || len(first) < 32 {
		t.Fatalf("unexpected TOTP secrets %q and %q", first, second)
	}
}

func TestGenerateRecoveryCodes(t *testing.T) {
	t.Parallel()
	codes, err := domain.GenerateRecoveryCodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(codes) != domain.RecoveryCodeCount {
		t.Fatalf("got %d codes, want %d", len(codes), domain.RecoveryCodeCount)
	}
	format := regexp.MustCompile(`^[A-Z2-7]{4}-[A-Z2-7]{4}$`)
	seen := make(map[string]struct{}, len(codes))
	for _, code := range codes {
		if !format.MatchString(code) {
			t.Fatalf("recovery code %q has an unsafe format", code)
		}
		if _, exists := seen[code]; exists {
			t.Fatalf("duplicate recovery code %q", code)
		}
		seen[code] = struct{}{}
	}
}

func TestTwoFactorPolicies(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		act  func() error
		want error
	}{
		{name: "recovery code cannot be reused", act: func() error { return domain.ValidateRecoveryCode(false, true) }, want: domain.ErrRecoveryCodeInvalid},
		{name: "recovery code must match", act: func() error { return domain.ValidateRecoveryCode(false, false) }, want: domain.ErrRecoveryCodeInvalid},
		{name: "matching unused recovery code is valid", act: func() error { return domain.ValidateRecoveryCode(true, false) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.act(); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}
