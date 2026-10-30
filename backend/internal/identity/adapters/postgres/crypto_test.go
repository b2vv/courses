//go:build integration

package postgres

import (
	"bytes"

	"courses/internal/identity/adapters/identitycrypto"
)

// testKeys builds the 2FA at-rest crypto from a fixed in-test master so repository tests exercise the
// real encrypt/decrypt and keyed-HMAC paths (GetTwoFactor decrypts on read; a nil keys would panic).
func testKeys() *identitycrypto.Keys {
	keys, err := identitycrypto.New(bytes.Repeat([]byte{0x2a}, 32))
	if err != nil {
		panic(err)
	}
	return keys
}
