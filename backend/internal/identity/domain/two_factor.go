package domain

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1" // #nosec G505 -- RFC 6238's required SHA-1 TOTP profile.
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	totpPeriod        = 30 * time.Second
	totpDigits        = 6
	totpSecretBytes   = 20
	RecoveryCodeCount = 10
	recoveryCodeBytes = 5
)

var base32NoPadding = base32.StdEncoding.WithPadding(base32.NoPadding)

// GenerateTOTPSecret returns a canonical, unpadded Base32 secret with 160 bits of entropy.
func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, totpSecretBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base32NoPadding.EncodeToString(bytes), nil
}

// VerifyTOTP accepts a six-digit RFC 6238 SHA-1 code from one adjacent 30-second period.
func VerifyTOTP(secret, code string, now time.Time) bool {
	key, ok := decodeTOTPSecret(secret)
	if !ok || !validTOTPCode(code) || now.Unix() < 0 {
		return false
	}
	for offset := int64(-1); offset <= 1; offset++ {
		counter := now.Unix()/int64(totpPeriod/time.Second) + offset
		if counter < 0 {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(totpCode(key, uint64(counter))), []byte(code)) == 1 {
			return true
		}
	}
	return false
}

// TOTPURI produces the otpauth URI consumed by authenticator applications.
func TOTPURI(issuer, account, secret string) (string, error) {
	if issuer == "" || account == "" {
		return "", ErrInvalidTOTP
	}
	if _, ok := decodeTOTPSecret(secret); !ok {
		return "", ErrInvalidTOTP
	}
	uri := url.URL{Scheme: "otpauth", Host: "totp", Path: issuer + ":" + account}
	query := url.Values{
		"algorithm": {"SHA1"},
		"digits":    {strconv.Itoa(totpDigits)},
		"issuer":    {issuer},
		"period":    {strconv.FormatInt(int64(totpPeriod/time.Second), 10)},
		"secret":    {secret},
	}
	uri.RawQuery = query.Encode()
	return uri.String(), nil
}

// GenerateRecoveryCodes returns plaintext codes exactly once; callers persist only their hashes.
func GenerateRecoveryCodes() ([]string, error) {
	codes := make([]string, 0, RecoveryCodeCount)
	seen := make(map[string]struct{}, RecoveryCodeCount)
	for len(codes) < RecoveryCodeCount {
		bytes := make([]byte, recoveryCodeBytes)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}
		encoded := base32NoPadding.EncodeToString(bytes)
		code := encoded[:4] + "-" + encoded[4:]
		if _, duplicate := seen[code]; duplicate {
			continue
		}
		seen[code] = struct{}{}
		codes = append(codes, code)
	}
	return codes, nil
}

// ValidateRecoveryCode keeps the one-time policy independent from hash implementation.
// matched comes from a constant-time hash verifier and consumed from durable state.
func ValidateRecoveryCode(matched, consumed bool) error {
	if !matched || consumed {
		return ErrRecoveryCodeInvalid
	}
	return nil
}

// ValidateTwoFactorEnrollment requires a proof from the secret before it can be enabled.
func ValidateTwoFactorEnrollment(enabled bool, secret, code string, now time.Time) error {
	if enabled {
		return ErrTwoFactorAlreadyEnabled
	}
	if !VerifyTOTP(secret, code, now) {
		return ErrInvalidTOTP
	}
	return nil
}

// ValidateTwoFactorDisable requires either a valid TOTP proof or an available recovery code.
func ValidateTwoFactorDisable(enabled bool, secret, code string, recoveryCodeAvailable bool, now time.Time) error {
	if !enabled {
		return ErrTwoFactorNotEnabled
	}
	if VerifyTOTP(secret, code, now) || recoveryCodeAvailable {
		return nil
	}
	return ErrRecoveryCodeInvalid
}

func decodeTOTPSecret(secret string) ([]byte, bool) {
	if secret == "" || strings.ToUpper(secret) != secret || strings.Contains(secret, "=") {
		return nil, false
	}
	decoded, err := base32NoPadding.DecodeString(secret)
	if err != nil || len(decoded) < totpSecretBytes {
		return nil, false
	}
	return decoded, true
}

func totpCode(secret []byte, counter uint64) string {
	message := make([]byte, 8)
	binary.BigEndian.PutUint64(message, counter)
	mac := hmac.New(sha1.New, secret)
	_, _ = mac.Write(message)
	sum := mac.Sum(nil)
	offset := sum[len(sum)-1] & 0x0f
	value := binary.BigEndian.Uint32(sum[offset:offset+4]) & 0x7fffffff
	return strconv.FormatInt(int64(value%1_000_000+1_000_000), 10)[1:]
}

func validTOTPCode(code string) bool {
	if len(code) != totpDigits {
		return false
	}
	for _, character := range code {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
