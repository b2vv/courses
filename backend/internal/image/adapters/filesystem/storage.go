package filesystem

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"strings"

	"courses/internal/image/ports"
)

const fileMode = 0o640

var errInvalidImageKey = errors.New("invalid image key")

type Storage struct {
	root *os.Root
}

var _ ports.Storage = (*Storage)(nil)

func New(rootPath string) (*Storage, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	return &Storage{root: root}, nil
}

func (s *Storage) Close() error { return s.root.Close() }

func (s *Storage) Put(ctx context.Context, id, extension string, r io.Reader) (written int64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if !validUUID(id) || !validExtension(extension) {
		return 0, errInvalidImageKey
	}
	name := id + "." + extension
	file, err := s.root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fileMode)
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	written, err = io.Copy(file, r)
	if err != nil {
		return written, errors.Join(err, s.root.Remove(name))
	}
	if err := ctx.Err(); err != nil {
		return written, errors.Join(err, s.root.Remove(name))
	}
	if err := file.Sync(); err != nil {
		return written, errors.Join(err, s.root.Remove(name))
	}
	return written, nil
}

// validExtension gates the filename suffix to the served image allowlist; nginx maps each to its
// Content-Type by extension. Kept here (not just the caller) so the traversal guard is complete.
func validExtension(extension string) bool {
	switch extension {
	case "jpg", "png", "webp":
		return true
	default:
		return false
	}
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}
