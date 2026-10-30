// Package identitycrypto holds the at-rest protection for two-factor secrets: AES-256-GCM encryption
// for the reversible TOTP secret and a keyed HMAC for the one-way recovery-code hashes. Both sub-keys
// are HKDF-derived from one master key (env IDENTITY_2FA_SECRET_KEY) with distinct info labels, so a
// single configured secret never gets reused byte-for-byte across two primitives (NIST SP 800-108).
//
// Why keyed, not argon2id, for recovery codes (SA13-I / I-F6, owner 2026-10-30): the phase already
// introduces a server-side key for the TOTP secret, and a keyed MAC makes the low-entropy codes
// unforgeable without that key regardless of their 40 bits — while keeping the existing atomic,
// indexed by-hash consume. argon2id would only win if the key leaked alongside the database, but in
// that case the co-located TOTP secret is already decryptable and 2FA is already lost.
package identitycrypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hkdf"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
)

// minMasterKeyBytes is the smallest master accepted: 32 bytes of entropy feed two 256-bit sub-keys.
const minMasterKeyBytes = 32

const (
	infoEncryption = "totp-secret-aes256gcm"
	infoRecovery   = "recovery-code-hmac"
	subKeyBytes    = 32
)

// ErrShortMasterKey means the configured master key carried fewer than 32 bytes of entropy.
var ErrShortMasterKey = errors.New("identity 2FA master key must be at least 32 bytes")

// ErrInvalidCiphertext means a stored TOTP secret could not be decrypted: it was truncated, tampered,
// or written under a different key. It is never treated as plaintext (forward-only, fail-closed).
var ErrInvalidCiphertext = errors.New("identity 2FA secret ciphertext invalid")

// Keys derives and holds the two sub-keys for one identity's two-factor at-rest protection.
type Keys struct {
	aead    cipher.AEAD
	hmacKey []byte
}

// New derives the AES and HMAC sub-keys from master via HKDF-SHA256 and builds the GCM AEAD. master
// must carry at least 32 bytes; a shorter key is rejected rather than stretched.
func New(master []byte) (*Keys, error) {
	if len(master) < minMasterKeyBytes {
		return nil, ErrShortMasterKey
	}
	encryptionKey, err := hkdf.Key(sha256.New, master, nil, infoEncryption, subKeyBytes)
	if err != nil {
		return nil, err
	}
	recoveryKey, err := hkdf.Key(sha256.New, master, nil, infoRecovery, subKeyBytes)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Keys{aead: aead, hmacKey: recoveryKey}, nil
}

// EncryptSecret seals plaintext under AES-256-GCM with a fresh random nonce, returning
// base64(nonce || ciphertext||tag). A new nonce per call keeps re-encryption of the same secret
// unlinkable.
func (keys *Keys) EncryptSecret(plaintext string) (string, error) {
	nonce := make([]byte, keys.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := keys.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawStdEncoding.EncodeToString(sealed), nil
}

// DecryptSecret reverses EncryptSecret. A truncated, tampered, or wrong-key input yields
// ErrInvalidCiphertext (the GCM auth tag fails), never a silent plaintext fallback.
func (keys *Keys) DecryptSecret(encoded string) (string, error) {
	sealed, err := base64.RawStdEncoding.DecodeString(encoded)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	nonceSize := keys.aead.NonceSize()
	if len(sealed) < nonceSize {
		return "", ErrInvalidCiphertext
	}
	nonce, ciphertext := sealed[:nonceSize], sealed[nonceSize:]
	plaintext, err := keys.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}
	return string(plaintext), nil
}

// RecoveryHash is the deterministic, keyed hash stored for a recovery code: hex(HMAC-SHA256(k_hmac,
// code)). Deterministic so the existing indexed by-hash lookup/consume is unchanged; keyed so the
// stored value is unforgeable without the master key even at 40 bits of code entropy. The 32-byte MAC
// hex-encodes to 64 chars, the same width as the superseded sha256 hex, so the code_hash column is
// unchanged.
func (keys *Keys) RecoveryHash(code string) string {
	mac := hmac.New(sha256.New, keys.hmacKey)
	mac.Write([]byte(code))
	return hex.EncodeToString(mac.Sum(nil))
}
