package filesystem

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"courses/internal/media/domain"
)

const (
	testAssetID   = "11111111-1111-4111-8111-111111111111"
	testClaimLeaf = "22222222222242228222222222222222"
	testOtherID   = "33333333-3333-4333-8333-333333333333"
)

var errFixtureRead = errors.New("fixture read failed")

func TestB11StorageCannotBeWalkedOutOf(t *testing.T) {
	storage, rootPath := openTestStorage(t)
	ctx := context.Background()

	for name, key := range map[string]domain.StorageKey{
		"dot dot":     {Kind: domain.StorageTemporary, ID: testAssetID, Leaf: "../escape"},
		"absolute":    {Kind: domain.StorageTemporary, ID: testAssetID, Leaf: "/etc/passwd"},
		"non UUID id": {Kind: domain.StorageTemporary, ID: "not-a-uuid", Leaf: "upload"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := storage.Append(ctx, key, 0, strings.NewReader("escaped")); err == nil {
				t.Fatal("Append() error = nil")
			}
		})
	}
	if _, err := os.Stat(filepath.Join(rootPath, "escape")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("escape path exists or stat failed: %v", err)
	}

	if err := os.Symlink("/etc", filepath.Join(rootPath, "hls", testAssetID)); err != nil {
		t.Fatal(err)
	}
	if _, err := storage.Size(ctx, domain.StorageKey{Kind: domain.StorageHLS, ID: testAssetID}); err == nil {
		t.Fatal("Size() followed a symlink outside the root")
	}

	outsidePath := t.TempDir()
	outsideFile := filepath.Join(outsidePath, "original")
	writeFixture(t, outsideFile, "outside")
	if err := os.Symlink(outsidePath, filepath.Join(rootPath, "source", testOtherID)); err != nil {
		t.Fatal(err)
	}
	outsideKey := domain.StorageKey{Kind: domain.StorageSource, ID: testOtherID, Leaf: "original"}
	if err := storage.Truncate(ctx, outsideKey, 0); err == nil {
		t.Fatal("Truncate() followed a symlink outside the root")
	}
	assertFile(t, outsideFile, "outside")
}

func TestAppendTruncatesToOffsetAndMovePreservesBytes(t *testing.T) {
	storage, rootPath := openTestStorage(t)
	ctx := context.Background()
	scratch := domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID, Leaf: "upload"}

	if written, err := storage.Append(ctx, scratch, 0, strings.NewReader("abcdef")); err != nil || written != 6 {
		t.Fatalf("first Append() = %d, %v", written, err)
	}
	if _, err := storage.Append(ctx, scratch, 7, strings.NewReader("invalid")); !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("short-file Append() error = %v", err)
	}
	assertFile(t, filepath.Join(rootPath, "temporary", testAssetID+"-upload"), "abcdef")

	failingBody := io.MultiReader(strings.NewReader("XY"), errorReader{})
	if _, err := storage.Append(ctx, scratch, 3, failingBody); !errors.Is(err, errFixtureRead) {
		t.Fatalf("failed Append() error = %v", err)
	}
	assertFile(t, filepath.Join(rootPath, "temporary", testAssetID+"-upload"), "abc")
	if written, err := storage.Append(ctx, scratch, 3, strings.NewReader("XYZ")); err != nil || written != 3 {
		t.Fatalf("repairing Append() = %d, %v", written, err)
	}
	if err := storage.Truncate(ctx, scratch, 5); err != nil {
		t.Fatal(err)
	}
	if size, err := storage.Size(ctx, scratch); err != nil || size != 5 {
		t.Fatalf("Size() = %d, %v", size, err)
	}

	source := domain.StorageKey{Kind: domain.StorageSource, ID: testAssetID, Leaf: "original"}
	if err := storage.Move(ctx, scratch, source); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(rootPath, "source", testAssetID, "original"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "abcXY" {
		t.Fatalf("moved content = %q", content)
	}
}

func TestHeadReturnsWholeShortFile(t *testing.T) {
	storage, _ := openTestStorage(t)
	key := domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID, Leaf: "upload"}
	if _, err := storage.Append(context.Background(), key, 0, strings.NewReader("short")); err != nil {
		t.Fatal(err)
	}

	head, err := storage.Head(context.Background(), key, 12)
	if err != nil {
		t.Fatal(err)
	}
	if string(head) != "short" {
		t.Fatalf("Head() = %q", head)
	}
}

func TestB12PublishIsAtomicAndPurgeIsIdempotent(t *testing.T) {
	storage, rootPath := openTestStorage(t)
	ctx := context.Background()
	work := domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID, Leaf: testClaimLeaf}
	destination := domain.StorageKey{Kind: domain.StorageHLS, ID: testAssetID}
	workPath := filepath.Join(rootPath, "temporary", testAssetID+"-"+testClaimLeaf)
	destinationPath := filepath.Join(rootPath, "hls", testAssetID)

	writeFixture(t, filepath.Join(workPath, "master.m3u8"), "complete")
	writeFixture(t, filepath.Join(workPath, "360p", "index.m3u8"), "variant")
	if err := storage.Publish(ctx, work, destination); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(destinationPath, "master.m3u8"), "complete")
	if _, err := os.Stat(workPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("published work still exists or stat failed: %v", err)
	}

	writeFixture(t, filepath.Join(workPath, "master.m3u8"), "replacement")
	writeFixture(t, filepath.Join(destinationPath, "stale"), "stale")
	if err := storage.Publish(ctx, work, destination); err != nil {
		t.Fatal(err)
	}
	assertFile(t, filepath.Join(destinationPath, "master.m3u8"), "replacement")
	if _, err := os.Stat(filepath.Join(destinationPath, "stale")); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("stale destination survived or stat failed: %v", err)
	}

	if err := storage.Publish(ctx, work, destination); err != nil {
		t.Fatalf("already-published Publish() = %v", err)
	}
	assertFile(t, filepath.Join(destinationPath, "master.m3u8"), "replacement")

	if err := storage.Purge(ctx, destination); err != nil {
		t.Fatal(err)
	}
	if err := storage.Publish(ctx, work, destination); err == nil {
		t.Fatal("interrupted Publish() error = nil")
	}
	if _, err := os.Stat(destinationPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("interrupted destination exists or stat failed: %v", err)
	}
	if err := storage.Purge(ctx, destination); err != nil {
		t.Fatal(err)
	}
	if err := storage.Purge(ctx, destination); err != nil {
		t.Fatalf("second Purge() = %v", err)
	}
}

func TestPurgeRemovesEveryTemporaryAssetPath(t *testing.T) {
	storage, rootPath := openTestStorage(t)
	writeFixture(t, filepath.Join(rootPath, "temporary", testAssetID+"-upload"), "source")
	writeFixture(t, filepath.Join(rootPath, "temporary", testAssetID+"-"+testClaimLeaf, "master.m3u8"), "work")
	writeFixture(t, filepath.Join(rootPath, "temporary", testAssetID+"-"+testClaimLeaf+".jpg"), "thumbnail")
	writeFixture(t, filepath.Join(rootPath, "temporary", testOtherID+"-upload"), "other")

	if err := storage.Purge(context.Background(), domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID}); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(rootPath, "temporary"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != testOtherID+"-upload" {
		t.Fatalf("temporary entries = %+v", entries)
	}
}

func TestFreeReturnsPositiveBytes(t *testing.T) {
	storage, _ := openTestStorage(t)
	free, err := storage.Free(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if free <= 0 {
		t.Fatalf("Free() = %d", free)
	}
}

func TestReadyWritesAndRemovesItsProbe(t *testing.T) {
	storage, rootPath := openTestStorage(t)
	if err := storage.Ready(context.Background()); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Join(rootPath, "temporary"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("temporary entries after readiness probe = %+v", entries)
	}
}

func openTestStorage(t *testing.T) (*Storage, string) {
	t.Helper()
	rootPath := t.TempDir()
	storage, err := New(rootPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := storage.Close(); err != nil {
			t.Error(err)
		}
	})
	return storage, rootPath
}

func writeFixture(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0o640); err != nil {
		t.Fatal(err)
	}
}

func assertFile(t *testing.T, name, want string) {
	t.Helper()
	content, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != want {
		t.Fatalf("%s = %q", name, content)
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) {
	return 0, errFixtureRead
}
