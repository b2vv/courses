package config

import (
	"encoding/base64"
	"fmt"
)

// minIdentity2FAKeyBytes matches identitycrypto's master-key floor: 32 bytes feed two 256-bit
// HKDF sub-keys (TOTP-secret AES, recovery-code HMAC).
const minIdentity2FAKeyBytes = 32

// Identity2FAKey decodes the master key that protects two-factor secrets at rest from the base64
// value in IDENTITY_2FA_SECRET_KEY. It is required (fail-closed): a missing, non-base64, or too-short
// value is an error so the service refuses to start rather than run with unprotected TOTP secrets.
func Identity2FAKey(getenv Getenv) ([]byte, error) {
	encoded := getenv("IDENTITY_2FA_SECRET_KEY")
	if encoded == "" {
		return nil, fmt.Errorf("IDENTITY_2FA_SECRET_KEY is required")
	}
	key, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("IDENTITY_2FA_SECRET_KEY must be base64")
	}
	if len(key) < minIdentity2FAKeyBytes {
		return nil, fmt.Errorf("IDENTITY_2FA_SECRET_KEY must decode to at least %d bytes", minIdentity2FAKeyBytes)
	}
	return key, nil
}
