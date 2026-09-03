package argon2id_test

import (
	"encoding/base64"
	"fmt"
	"testing"

	passwords "courses/internal/identity/adapters/argon2id"
	"golang.org/x/crypto/argon2"
)

func TestVerifierAcceptsOnlySupportedArgon2idEncoding(t *testing.T) {
	t.Parallel()
	verifier := passwords.NewVerifier()
	encoded := encode("courseforge-demo", []byte("sixteen-byte-slt"))
	matched, err := verifier.Verify(encoded, "courseforge-demo")
	if err != nil || !matched {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
	matched, err = verifier.Verify(encoded, "wrong-password")
	if err != nil || matched {
		t.Fatalf("matched=%v err=%v", matched, err)
	}
	if err := verifier.VerifyDummy("unknown-password"); err != nil {
		t.Fatal(err)
	}

	for _, encoded := range []string{
		"plaintext",
		"$argon2i$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=16$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=1,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=65536,t=3,p=2$bad$bad",
	} {
		if _, err := verifier.Verify(encoded, "password"); err == nil {
			t.Fatalf("accepted malformed hash %q", encoded)
		}
	}
}

func encode(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, 3, 64*1024, 2, 32)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}
