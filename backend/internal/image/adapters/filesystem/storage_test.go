package filesystem

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

const imageID = "12345678-1234-4234-8234-123456789abc"

func newTestStorage(t *testing.T) (*Storage, string) {
	t.Helper()
	directory := t.TempDir()
	storage, err := New(directory)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := storage.Close(); err != nil {
			t.Errorf("close storage: %v", err)
		}
	})
	return storage, directory
}

func assertEmpty(t *testing.T, directory string) {
	t.Helper()
	entries, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("unexpected files: %v", entries)
	}
}

func TestPut(t *testing.T) {
	storage, directory := newTestStorage(t)
	payload := []byte("image bytes")
	written, err := storage.Put(context.Background(), imageID, bytes.NewReader(payload))
	if err != nil || written != int64(len(payload)) {
		t.Fatalf("Put = %d, %v; want %d, nil", written, err, len(payload))
	}
	name := filepath.Join(directory, imageID)
	got, err := os.ReadFile(name)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("stored bytes = %q, %v; want %q", got, err, payload)
	}
	info, err := os.Stat(name)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o640 {
		t.Fatalf("mode = %o; want 640", info.Mode().Perm())
	}
}

func TestPutInvalidID(t *testing.T) {
	for _, id := range []string{"../escape", "not-a-uuid", "00000000-0000-0000-0000-000000000000", "12345678-1234-4234-8234-123456789ABC", "12345678-1234-4234-8234-123456789abg", "12345678_1234-4234-8234-123456789abc"} {
		t.Run(id, func(t *testing.T) {
			storage, directory := newTestStorage(t)
			written, err := storage.Put(context.Background(), id, bytes.NewReader([]byte("image")))
			if written != 0 || !errors.Is(err, errInvalidImageKey) {
				t.Fatalf("Put = %d, %v; want 0, invalid image key", written, err)
			}
			assertEmpty(t, directory)
		})
	}
}

func TestPutDuplicatePreservesOriginal(t *testing.T) {
	storage, directory := newTestStorage(t)
	payload := []byte("original image")
	if _, err := storage.Put(context.Background(), imageID, bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Put(context.Background(), imageID, bytes.NewReader([]byte("replacement"))); err == nil {
		t.Fatal("duplicate Put succeeded")
	}
	got, err := os.ReadFile(filepath.Join(directory, imageID))
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("original bytes = %q, %v; want %q", got, err, payload)
	}
}

func TestPutCancelledContext(t *testing.T) {
	storage, directory := newTestStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	written, err := storage.Put(ctx, imageID, bytes.NewReader([]byte("image")))
	if written != 0 || !errors.Is(err, context.Canceled) {
		t.Fatalf("Put = %d, %v; want 0, context canceled", written, err)
	}
	assertEmpty(t, directory)
}

func TestPutRemovesPartialFile(t *testing.T) {
	readError := errors.New("reader failed")
	for _, duringCopy := range []bool{false, true} {
		name := "read failure"
		if duringCopy {
			name = "canceled during copy"
		}
		t.Run(name, func(t *testing.T) {
			storage, directory := newTestStorage(t)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			wantErr := readError
			reader := readerFunc(func(p []byte) (int, error) {
				n := copy(p, "partial")
				if duringCopy {
					cancel()
					return n, io.EOF
				}
				return n, readError
			})
			if duringCopy {
				wantErr = context.Canceled
			}
			written, err := storage.Put(ctx, imageID, reader)
			if written != int64(len("partial")) || !errors.Is(err, wantErr) {
				t.Fatalf("Put = %d, %v; want %d, %v", written, err, len("partial"), wantErr)
			}
			assertEmpty(t, directory)
		})
	}
}

type readerFunc func([]byte) (int, error)

func (r readerFunc) Read(p []byte) (int, error) { return r(p) }
