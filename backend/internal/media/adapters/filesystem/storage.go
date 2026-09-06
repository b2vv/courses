package filesystem

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"syscall"

	"courses/internal/media/domain"
)

const (
	directoryMode = 0o750
	fileMode      = 0o640
)

var errInvalidStorageKey = errors.New("invalid storage key")

type Storage struct {
	root *os.Root
}

func New(rootPath string) (*Storage, error) {
	root, err := os.OpenRoot(rootPath)
	if err != nil {
		return nil, err
	}
	for _, directory := range []string{
		string(domain.StorageSource),
		string(domain.StorageHLS),
		string(domain.StorageThumbnail),
		string(domain.StorageTemporary),
	} {
		if err := root.MkdirAll(directory, directoryMode); err != nil {
			return nil, errors.Join(err, root.Close())
		}
	}
	return &Storage{root: root}, nil
}

func (storage *Storage) Close() error {
	return storage.root.Close()
}

func (storage *Storage) Free(ctx context.Context) (free int64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	directory, err := storage.root.Open(".")
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()

	var statistics syscall.Statfs_t
	if err := syscall.Fstatfs(int(directory.Fd()), &statistics); err != nil {
		return 0, err
	}
	return int64(statistics.Bavail) * int64(statistics.Bsize), nil
}

func (storage *Storage) Size(ctx context.Context, key domain.StorageKey) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	name, err := resolve(key)
	if err != nil {
		return 0, err
	}
	information, err := storage.root.Stat(name)
	if err != nil {
		return 0, err
	}
	return information.Size(), nil
}

func (storage *Storage) Head(ctx context.Context, key domain.StorageKey, n int) (head []byte, err error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if n < 0 {
		return nil, fmt.Errorf("head length %d: %w", n, errInvalidStorageKey)
	}
	name, err := resolve(key)
	if err != nil {
		return nil, err
	}
	file, err := storage.root.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	return io.ReadAll(io.LimitReader(file, int64(n)))
}

func (storage *Storage) Append(ctx context.Context, key domain.StorageKey, offset int64, body io.Reader) (written int64, err error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if offset < 0 {
		return 0, fmt.Errorf("append offset %d: %w", offset, errInvalidStorageKey)
	}
	name, err := resolve(key)
	if err != nil {
		return 0, err
	}
	if err := storage.root.MkdirAll(path.Dir(name), directoryMode); err != nil {
		return 0, err
	}
	flags := os.O_WRONLY
	if offset == 0 {
		flags |= os.O_CREATE
	}
	file, err := storage.root.OpenFile(name, flags, fileMode)
	if err != nil {
		return 0, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()

	if err := reconcileLength(file, offset); err != nil {
		return 0, err
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return 0, err
	}
	written, err = io.Copy(file, body)
	if err != nil {
		return written, rollback(file, offset, err)
	}
	if err := ctx.Err(); err != nil {
		return written, rollback(file, offset, err)
	}
	if err := file.Sync(); err != nil {
		return written, rollback(file, offset, err)
	}
	return written, nil
}

func (storage *Storage) Truncate(ctx context.Context, key domain.StorageKey, size int64) (err error) {
	if err := ctx.Err(); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("truncate size %d: %w", size, errInvalidStorageKey)
	}
	name, err := resolve(key)
	if err != nil {
		return err
	}
	flags := os.O_WRONLY
	if size == 0 {
		if err := storage.root.MkdirAll(path.Dir(name), directoryMode); err != nil {
			return err
		}
		flags |= os.O_CREATE
	}
	file, err := storage.root.OpenFile(name, flags, fileMode)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	if err := reconcileLength(file, size); err != nil {
		return err
	}
	return file.Sync()
}

func reconcileLength(file *os.File, offset int64) error {
	information, err := file.Stat()
	if err != nil {
		return err
	}
	if information.Size() < offset {
		return fmt.Errorf("file length %d is below offset %d: %w", information.Size(), offset, io.ErrUnexpectedEOF)
	}
	if information.Size() > offset {
		return file.Truncate(offset)
	}
	return nil
}

func rollback(file *os.File, offset int64, cause error) error {
	return errors.Join(cause, file.Truncate(offset), file.Sync())
}

func (storage *Storage) Move(ctx context.Context, from, to domain.StorageKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fromName, err := resolve(from)
	if err != nil {
		return err
	}
	toName, err := resolve(to)
	if err != nil {
		return err
	}
	if err := storage.root.MkdirAll(path.Dir(toName), directoryMode); err != nil {
		return err
	}
	return storage.root.Rename(fromName, toName)
}

func (storage *Storage) Publish(ctx context.Context, work, destination domain.StorageKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if work.Kind != domain.StorageTemporary || work.ID != destination.ID || work.Leaf == "upload" || strings.HasSuffix(work.Leaf, ".jpg") ||
		destination.Kind != domain.StorageHLS || destination.Leaf != "" {
		return errInvalidStorageKey
	}
	workName, err := resolve(work)
	if err != nil {
		return err
	}
	destinationName, err := resolve(destination)
	if err != nil {
		return err
	}
	if err := storage.root.MkdirAll(path.Dir(destinationName), directoryMode); err != nil {
		return err
	}
	if _, err := storage.root.Stat(workName); errors.Is(err, os.ErrNotExist) {
		if _, destinationErr := storage.root.Stat(destinationName); destinationErr == nil {
			return nil
		} else if !errors.Is(destinationErr, os.ErrNotExist) {
			return destinationErr
		}
		return err
	} else if err != nil {
		return err
	}
	// If both paths exist, the destination is a stale pre-commit publish.
	// Remove it so rename still lands on a fresh destination.
	if err := storage.root.RemoveAll(destinationName); err != nil {
		return err
	}
	return storage.root.Rename(workName, destinationName)
}

func (storage *Storage) Purge(ctx context.Context, key domain.StorageKey) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if key.Kind == domain.StorageTemporary && key.Leaf == "" {
		return storage.purgeTemporary(ctx, key.ID)
	}
	name, err := resolve(key)
	if err != nil {
		return err
	}
	return storage.root.RemoveAll(name)
}

func (storage *Storage) purgeTemporary(ctx context.Context, id string) (err error) {
	if !validUUID(id) {
		return errInvalidStorageKey
	}
	directory, err := storage.root.Open(string(domain.StorageTemporary))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, directory.Close()) }()

	entries, err := directory.ReadDir(-1)
	if err != nil {
		return err
	}
	prefix := id + "-"
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		if strings.HasPrefix(entry.Name(), prefix) {
			if err := storage.root.RemoveAll(path.Join(string(domain.StorageTemporary), entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func resolve(key domain.StorageKey) (string, error) {
	if !validUUID(key.ID) {
		return "", errInvalidStorageKey
	}
	switch key.Kind {
	case domain.StorageSource:
		if key.Leaf == "" {
			return path.Join(string(key.Kind), key.ID), nil
		}
		if key.Leaf == "original" {
			return path.Join(string(key.Kind), key.ID, key.Leaf), nil
		}
	case domain.StorageHLS:
		if key.Leaf == "" {
			return path.Join(string(key.Kind), key.ID), nil
		}
	case domain.StorageThumbnail:
		if key.Leaf == "" {
			return path.Join(string(key.Kind), key.ID+".jpg"), nil
		}
	case domain.StorageTemporary:
		if key.Leaf == "upload" || validClaimLeaf(key.Leaf) {
			return path.Join(string(key.Kind), key.ID+"-"+key.Leaf), nil
		}
	}
	return "", errInvalidStorageKey
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}

func validClaimLeaf(value string) bool {
	value = strings.TrimSuffix(value, ".jpg")
	if len(value) != 32 || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 16
}
