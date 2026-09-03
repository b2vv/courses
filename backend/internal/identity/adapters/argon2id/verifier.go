package argon2id

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memory      uint32 = 64 * 1024
	iterations  uint32 = 3
	parallelism uint8  = 2
	saltLength         = 16
	keyLength          = 32
)

var ErrInvalidEncoding = errors.New("invalid argon2id encoding")

type Verifier struct{ dummyHash string }

func NewVerifier() *Verifier {
	return &Verifier{dummyHash: encode("courseforge-invalid-user", []byte("courseforge-dumm"))}
}

func (verifier *Verifier) Verify(encoded, password string) (bool, error) {
	if len(encoded) > 512 || len(password) > 1024 {
		return false, ErrInvalidEncoding
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[0] != "" || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=65536,t=3,p=2" {
		return false, ErrInvalidEncoding
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != saltLength {
		return false, ErrInvalidEncoding
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != keyLength {
		return false, ErrInvalidEncoding
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func (verifier *Verifier) VerifyDummy(password string) error {
	_, err := verifier.Verify(verifier.dummyHash, password)
	return err
}

func encode(password string, salt []byte) string {
	hash := argon2.IDKey([]byte(password), salt, iterations, memory, parallelism, keyLength)
	return fmt.Sprintf("$argon2id$v=19$m=65536,t=3,p=2$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(hash))
}
