package application

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

const accountTokenBytes = 32

// TokenMaterial keeps the bearer token separate from the value safe to persist.
// Raw is sent only to the account holder; Digest is stored by TokenStore.
type TokenMaterial struct {
	Raw    string
	Digest [sha256.Size]byte
}

func NewTokenMaterial() (TokenMaterial, error) {
	raw := make([]byte, accountTokenBytes)
	if _, err := io.ReadFull(rand.Reader, raw); err != nil {
		return TokenMaterial{}, fmt.Errorf("read account token randomness: %w", err)
	}
	return TokenMaterial{
		Raw:    base64.RawURLEncoding.EncodeToString(raw),
		Digest: sha256.Sum256(raw),
	}, nil
}
