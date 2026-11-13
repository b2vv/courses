//go:build integration

package ffmpeg

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"courses/internal/media/domain"
)

// TestAudioProbeAndRenderWithRealFFmpeg drives the two audio transcode steps that
// each hid a real-adapter resolver mismatch — ffprobe on the uploaded source and
// the HLS render — against the real ffmpeg/ffprobe binaries on a real AAC file.
// A fake transcoder (the earlier S5-L0 coverage) could not have caught the probe
// resolver rejecting audio-<trackID>; this test fails if that path breaks again.
func TestAudioProbeAndRenderWithRealFFmpeg(t *testing.T) {
	requireFFmpegBinaries(t)
	root := t.TempDir()
	transcoder := New(root)
	const (
		assetID = "11111111-1111-4111-8111-111111111111"
		trackID = "44444444-4444-4444-8444-444444444444"
		token   = "abcdef0123456789abcdef0123456789"
	)

	sourceDir := filepath.Join(root, "source", assetID)
	if err := os.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatal(err)
	}
	sourcePath := filepath.Join(sourceDir, domain.AudioLeafPrefix+trackID)
	generateAACFixture(t, sourcePath)

	source := domain.StorageKey{Kind: domain.StorageSource, ID: assetID, Leaf: domain.AudioLeafPrefix + trackID}
	work := domain.StorageKey{Kind: domain.StorageTemporary, ID: assetID, Leaf: token}

	probe, err := transcoder.Probe(context.Background(), source)
	if err != nil {
		t.Fatalf("Probe of a real uploaded audio source: %v", err)
	}
	if probe.AudioCodec != domain.CodecAAC {
		t.Fatalf("probed audio codec = %q, want %q", probe.AudioCodec, domain.CodecAAC)
	}
	if err := domain.AcceptAudioTrackSource(probe); err != nil {
		t.Fatalf("AcceptAudioTrackSource(probe) = %v", err)
	}

	if err := transcoder.RenderAudioTrack(context.Background(), source, work); err != nil {
		t.Fatalf("RenderAudioTrack against real ffmpeg: %v", err)
	}
	workPath := filepath.Join(root, "temporary", assetID+"-"+token)
	for _, name := range []string{"index.m3u8", "init.mp4"} {
		if _, err := os.Stat(filepath.Join(workPath, name)); err != nil {
			t.Fatalf("expected %s in the HLS output: %v", name, err)
		}
	}
	segments, err := filepath.Glob(filepath.Join(workPath, "segment_*.m4s"))
	if err != nil {
		t.Fatal(err)
	}
	if len(segments) == 0 {
		t.Fatal("RenderAudioTrack produced no HLS segments")
	}
}

func requireFFmpegBinaries(t *testing.T) {
	t.Helper()
	for _, binary := range []string{"ffmpeg", "ffprobe"} {
		if _, err := exec.LookPath(binary); err != nil {
			t.Skipf("%s not on PATH; skipping real-ffmpeg integration test", binary)
		}
	}
}

func generateAACFixture(t *testing.T, path string) {
	t.Helper()
	cmd := exec.Command("ffmpeg",
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-c:a", "aac", "-b:a", "128k", "-movflags", "+faststart",
		"-f", "mp4", path)
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("generate AAC fixture: %v\n%s", err, output)
	}
}
