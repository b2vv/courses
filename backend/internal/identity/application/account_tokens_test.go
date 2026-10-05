package application_test

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"

	"courses/internal/identity/application"
)

func TestNewTokenMaterialReturnsOpaqueRawTokenAndItsDigest(t *testing.T) {
	t.Parallel()

	material, err := application.NewTokenMaterial()
	if err != nil {
		t.Fatal(err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(material.Raw)
	if err != nil {
		t.Fatalf("raw token encoding: %v", err)
	}
	if len(raw) != 32 {
		t.Fatalf("raw token has %d bytes, want 32", len(raw))
	}
	want := sha256.Sum256(raw)
	if material.Digest != want {
		t.Fatalf("wrong digest: got %x, want %x", material.Digest, want)
	}
}
