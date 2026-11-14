package media

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"courses/internal/transcript/domain"
)

func TestExtractReturnsNotFoundWhenSourceIsMissing(t *testing.T) {
	t.Parallel()
	audio := NewAudio(t.TempDir(), t.TempDir())
	_, _, err := audio.Extract(context.Background(), "11111111-1111-4111-8111-111111111111", "original")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("error=%v want=ErrNotFound", err)
	}
}

func TestExtractRefusesInvalidAssetID(t *testing.T) {
	t.Parallel()
	audio := NewAudio(t.TempDir(), t.TempDir())
	for _, id := range []string{"", "not-a-uuid", "00000000-0000-0000-0000-000000000000"} {
		if _, _, err := audio.Extract(context.Background(), id, "original"); err == nil {
			t.Fatalf("invalid assetID %q must be refused", id)
		}
	}
}

func TestExtractProducesAudioFileAndReleaseRemovesIt(t *testing.T) {
	t.Parallel()
	requireFFmpeg(t)

	mediaRoot := t.TempDir()
	tmpDir := t.TempDir()
	assetID := "11111111-1111-4111-8111-111111111111"
	createSourceSilence(t, mediaRoot, assetID, 1)

	audio := NewAudio(mediaRoot, tmpDir)
	audioPath, release, err := audio.Extract(context.Background(), assetID, "original")
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(audioPath)
	if err != nil {
		t.Fatalf("extracted file does not exist: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("extracted file is empty")
	}

	release()
	if _, err := os.Stat(audioPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("release must remove the temporary file")
	}
}

func TestExtractReturnsContextErrorWhenAlreadyCancelled(t *testing.T) {
	t.Parallel()
	requireFFmpeg(t)

	mediaRoot := t.TempDir()
	tmpDir := t.TempDir()
	assetID := "11111111-1111-4111-8111-111111111111"
	createSourceSilence(t, mediaRoot, assetID, 1)

	audio := NewAudio(mediaRoot, tmpDir)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before calling Extract

	_, _, err := audio.Extract(ctx, assetID, "original")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error=%v want=context.Canceled", err)
	}

	// No temp files should be leaked.
	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".wav" {
			t.Fatalf("temp file %s was not cleaned up", entry.Name())
		}
	}
}

func TestExtractCleansUpOnFFmpegFailure(t *testing.T) {
	t.Parallel()
	mediaRoot := t.TempDir()
	tmpDir := t.TempDir()
	assetID := "11111111-1111-4111-8111-111111111111"
	sourceDir := filepath.Join(mediaRoot, "source", assetID)
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "original"), []byte("not a video"), 0o644); err != nil {
		t.Fatal(err)
	}

	audio := NewAudio(mediaRoot, tmpDir)
	_, _, err := audio.Extract(context.Background(), assetID, "original")
	if err == nil {
		t.Fatal("ffmpeg failure must be an error")
	}

	entries, _ := os.ReadDir(tmpDir)
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) == ".wav" {
			t.Fatalf("temp file %s was not cleaned up after failure", entry.Name())
		}
	}
}

// helpers

func requireFFmpeg(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		t.Skipf("ffmpeg not available: %v", err)
	}
}

func createSourceSilence(t *testing.T, mediaRoot, assetID string, seconds int) {
	t.Helper()
	sourceDir := filepath.Join(mediaRoot, "source", assetID)
	if err := os.MkdirAll(sourceDir, 0o755); err != nil {
		t.Fatal(err)
	}
	outputPath := filepath.Join(sourceDir, "original")
	cmd := exec.Command("ffmpeg",
		"-nostdin", "-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "anullsrc=r=44100:cl=mono",
		"-t", fmt.Sprintf("%d", seconds),
		"-c:a", "pcm_s16le", "-f", "wav", "-y", outputPath,
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("generate test silence: %v", err)
	}
}

// Multi-audio S6 (D-S6-4): the leaf is joined into a path, so only the two media shapes pass.
func TestExtractRefusesAnUnsafeLeafBeforeTouchingTheDisk(t *testing.T) {
	t.Parallel()
	audio := NewAudio(t.TempDir(), t.TempDir())
	for _, leaf := range []string{"", "../original", "audio-", "audio-../../x", "Original"} {
		if _, _, err := audio.Extract(context.Background(), "11111111-1111-4111-8111-111111111111", leaf); !errors.Is(err, domain.ErrInvalidRequest) {
			t.Errorf("leaf %q err=%v, want ErrInvalidRequest", leaf, err)
		}
	}
}

// An uploaded track's source is audio only; the same ffmpeg arguments extract it.
func TestExtractReadsAnUploadedTrackLeaf(t *testing.T) {
	t.Parallel()
	requireFFmpeg(t)
	mediaRoot := t.TempDir()
	assetID := "11111111-1111-4111-8111-111111111111"
	leaf := "audio-70000000-0000-4000-8000-000000000001"
	directory := filepath.Join(mediaRoot, "source", assetID)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	// m4a container, like an uploaded track; the source leaf has no extension, as media stores it.
	output, err := exec.Command("ffmpeg", "-nostdin", "-loglevel", "error", "-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-c:a", "aac", "-f", "mp4", "-y", filepath.Join(directory, leaf)).CombinedOutput()
	if err != nil {
		t.Fatalf("make audio-only source: %v: %s", err, output)
	}
	audio := NewAudio(mediaRoot, t.TempDir())
	path, release, err := audio.Extract(context.Background(), assetID, leaf)
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if info, err := os.Stat(path); err != nil || info.Size() == 0 {
		t.Fatalf("extracted=%v err=%v", info, err)
	}
	if _, _, err := audio.Extract(context.Background(), assetID, "original"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the original leaf of an audio-only asset err=%v, want ErrNotFound", err)
	}
}
