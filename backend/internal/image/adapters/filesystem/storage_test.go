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

const (
	imageID  = "12345678-1234-4234-8234-123456789abc"
	imageExt = "png"
	imageKey = imageID + "." + imageExt
)

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
	written, err := storage.Put(context.Background(), imageID, imageExt, bytes.NewReader(payload))
	if err != nil || written != int64(len(payload)) {
		t.Fatalf("Put = %d, %v; want %d, nil", written, err, len(payload))
	}
	name := filepath.Join(directory, imageKey)
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
			written, err := storage.Put(context.Background(), id, imageExt, bytes.NewReader([]byte("image")))
			if written != 0 || !errors.Is(err, errInvalidImageKey) {
				t.Fatalf("Put = %d, %v; want 0, invalid image key", written, err)
			}
			assertEmpty(t, directory)
		})
	}
}

func TestPutInvalidExtension(t *testing.T) {
	for _, ext := range []string{"", "gif", "jpeg", "exe", "png/../escape", "PNG"} {
		t.Run(ext, func(t *testing.T) {
			storage, directory := newTestStorage(t)
			written, err := storage.Put(context.Background(), imageID, ext, bytes.NewReader([]byte("image")))
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
	if _, err := storage.Put(context.Background(), imageID, imageExt, bytes.NewReader(payload)); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Put(context.Background(), imageID, imageExt, bytes.NewReader([]byte("replacement"))); err == nil {
		t.Fatal("duplicate Put succeeded")
	}
	got, err := os.ReadFile(filepath.Join(directory, imageKey))
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("original bytes = %q, %v; want %q", got, err, payload)
	}
}

func TestPutCancelledContext(t *testing.T) {
	storage, directory := newTestStorage(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	written, err := storage.Put(ctx, imageID, imageExt, bytes.NewReader([]byte("image")))
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
			written, err := storage.Put(ctx, imageID, imageExt, reader)
			if written != int64(len("partial")) || !errors.Is(err, wantErr) {
				t.Fatalf("Put = %d, %v; want %d, %v", written, err, len("partial"), wantErr)
			}
			assertEmpty(t, directory)
		})
	}
}

func TestDelete(t *testing.T) {
	storage, directory := newTestStorage(t)
	if _, err := storage.Put(context.Background(), imageID, imageExt, bytes.NewReader([]byte("image bytes"))); err != nil {
		t.Fatal(err)
	}
	if err := storage.Delete(context.Background(), imageID, imageExt); err != nil {
		t.Fatalf("Delete = %v; want nil", err)
	}
	assertEmpty(t, directory)
}

func TestDeleteMissingFileErrors(t *testing.T) {
	storage, _ := newTestStorage(t)
	if err := storage.Delete(context.Background(), imageID, imageExt); err == nil {
		t.Fatal("Delete of a missing file returned nil; want an error the caller can join")
	}
}

func TestDeleteInvalidKey(t *testing.T) {
	for _, tc := range []struct{ id, ext string }{
		{"../escape", imageExt},
		{"not-a-uuid", imageExt},
		{"00000000-0000-0000-0000-000000000000", imageExt},
		{imageID, "gif"},
		{imageID, "png/../escape"},
	} {
		t.Run(tc.id+"."+tc.ext, func(t *testing.T) {
			storage, directory := newTestStorage(t)
			// A real file sits beside the invalid key; a rejected Delete must not touch it.
			if _, err := storage.Put(context.Background(), imageID, imageExt, bytes.NewReader([]byte("keep"))); err != nil {
				t.Fatal(err)
			}
			if err := storage.Delete(context.Background(), tc.id, tc.ext); !errors.Is(err, errInvalidImageKey) {
				t.Fatalf("Delete = %v; want invalid image key", err)
			}
			if _, err := os.Stat(filepath.Join(directory, imageKey)); err != nil {
				t.Fatalf("existing file disturbed by rejected Delete: %v", err)
			}
		})
	}
}

func TestDeleteCancelledContext(t *testing.T) {
	storage, directory := newTestStorage(t)
	if _, err := storage.Put(context.Background(), imageID, imageExt, bytes.NewReader([]byte("keep"))); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := storage.Delete(ctx, imageID, imageExt); !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete = %v; want context canceled", err)
	}
	if _, err := os.Stat(filepath.Join(directory, imageKey)); err != nil {
		t.Fatalf("file removed despite cancelled context: %v", err)
	}
}

type readerFunc func([]byte) (int, error)

func (r readerFunc) Read(p []byte) (int, error) { return r(p) }
