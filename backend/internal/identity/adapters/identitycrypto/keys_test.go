package identitycrypto

import (
	"crypto/hkdf"
	"crypto/sha256"
	"strings"
	"testing"
)

func testMaster() []byte {
	master := make([]byte, 32)
	for i := range master {
		master[i] = byte(i + 1)
	}
	return master
}

func TestNewRejectsShortMaster(t *testing.T) {
	if _, err := New(make([]byte, 31)); err != ErrShortMasterKey {
		t.Fatalf("31-byte master: got %v, want ErrShortMasterKey", err)
	}
	if _, err := New(testMaster()); err != nil {
		t.Fatalf("32-byte master: unexpected error %v", err)
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	secret := "JBSWY3DPEHPK3PXP"
	encoded, err := keys.EncryptSecret(secret)
	if err != nil {
		t.Fatal(err)
	}
	if encoded == secret || strings.Contains(encoded, secret) {
		t.Fatalf("ciphertext leaks plaintext: %q", encoded)
	}
	back, err := keys.DecryptSecret(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if back != secret {
		t.Fatalf("round-trip: got %q, want %q", back, secret)
	}
}

func TestEncryptUsesFreshNonce(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	first, err := keys.EncryptSecret("same-secret")
	if err != nil {
		t.Fatal(err)
	}
	second, err := keys.EncryptSecret("same-secret")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("two encryptions of the same secret produced identical ciphertext (nonce reuse)")
	}
}

func TestDecryptRejectsTamper(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := keys.EncryptSecret("secret")
	if err != nil {
		t.Fatal(err)
	}
	raw := []byte(encoded)
	raw[len(raw)-1] ^= 0x01 // flip a bit in the last base64 char (the tag region)
	if _, err := keys.DecryptSecret(string(raw)); err != ErrInvalidCiphertext {
		t.Fatalf("tampered ciphertext: got %v, want ErrInvalidCiphertext", err)
	}
}

func TestDecryptRejectsWrongKey(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := keys.EncryptSecret("secret")
	if err != nil {
		t.Fatal(err)
	}
	other := testMaster()
	other[0] ^= 0xFF
	wrong, err := New(other)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wrong.DecryptSecret(encoded); err != ErrInvalidCiphertext {
		t.Fatalf("decrypt under wrong key: got %v, want ErrInvalidCiphertext", err)
	}
}

func TestDecryptRejectsGarbage(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keys.DecryptSecret("not base64!!!"); err != ErrInvalidCiphertext {
		t.Fatalf("non-base64: got %v, want ErrInvalidCiphertext", err)
	}
	if _, err := keys.DecryptSecret("AAAA"); err != ErrInvalidCiphertext {
		t.Fatalf("too short for nonce: got %v, want ErrInvalidCiphertext", err)
	}
}

func TestRecoveryHashDeterministicAndDistinct(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	a1 := keys.RecoveryHash("ABCD-EFGH")
	a2 := keys.RecoveryHash("ABCD-EFGH")
	if a1 != a2 {
		t.Fatal("RecoveryHash not deterministic for the same code")
	}
	if len(a1) != 64 {
		t.Fatalf("hash width: got %d hex chars, want 64", len(a1))
	}
	if b := keys.RecoveryHash("ZZZZ-ZZZZ"); b == a1 {
		t.Fatal("distinct codes produced the same hash")
	}
}

func TestRecoveryHashIsKeyed(t *testing.T) {
	keys, err := New(testMaster())
	if err != nil {
		t.Fatal(err)
	}
	other := testMaster()
	other[5] ^= 0xFF
	wrong, err := New(other)
	if err != nil {
		t.Fatal(err)
	}
	if keys.RecoveryHash("ABCD-EFGH") == wrong.RecoveryHash("ABCD-EFGH") {
		t.Fatal("same code under different master keys produced the same HMAC (not keyed)")
	}
}

func TestSubKeysAreSeparated(t *testing.T) {
	// Guards the domain separation directly: the two HKDF info labels must derive different sub-keys
	// from the same master, or encryption and recovery would share key material. White-box: re-derive
	// both with the package's own labels and assert they differ.
	master := testMaster()
	encryptionKey, err := hkdf.Key(sha256.New, master, nil, infoEncryption, subKeyBytes)
	if err != nil {
		t.Fatal(err)
	}
	recoveryKey, err := hkdf.Key(sha256.New, master, nil, infoRecovery, subKeyBytes)
	if err != nil {
		t.Fatal(err)
	}
	if string(encryptionKey) == string(recoveryKey) {
		t.Fatal("HKDF derived identical sub-keys for distinct info labels (no domain separation)")
	}
}
