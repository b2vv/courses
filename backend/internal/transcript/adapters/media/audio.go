package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

// Compile-time check.
var _ ports.AudioSource = (*Audio)(nil)

const processWaitDelay = 2 * time.Second

// Audio implements ports.AudioSource by extracting 16 kHz mono audio from a media source using
// ffmpeg. The source path is derived from the media module's storage layout (plan D8-3):
// <mediaRoot>/source/<assetID>/<leaf>, where the leaf is the default audio track's — "original"
// for the video's own sound, "audio-<trackId>" for an uploaded track (multi-audio S6, D-S6-4). The adapter owns this derivation so that the caller
// — the application layer — does not need to know where media keeps its bytes.
//
// The extracted audio is written to a temporary file under `tmpDir`; the caller MUST call the
// returned `release` function promptly, because an hour of 16 kHz mono WAV is ~115 MB.
type Audio struct {
	mediaRoot string
	tmpDir    string
}

func NewAudio(mediaRoot, tmpDir string) *Audio {
	return &Audio{
		mediaRoot: filepath.Clean(mediaRoot),
		tmpDir:    filepath.Clean(tmpDir),
	}
}

func (audio *Audio) Extract(ctx context.Context, assetID, leaf string) (string, func(), error) {
	if err := domain.ValidateID(assetID); err != nil {
		return "", nil, err
	}
	// The leaf comes from the database, but it is joined into a path: only the two shapes the
	// media module writes may pass (path safety).
	if !domain.ValidAudioLeaf(leaf) {
		return "", nil, domain.ErrInvalidRequest
	}

	sourcePath := filepath.Join(audio.mediaRoot, "source", assetID, leaf)
	if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
		return "", nil, domain.ErrNotFound
	} else if err != nil {
		return "", nil, fmt.Errorf("stat source %s: %w", assetID, err)
	}

	outputFile, err := os.CreateTemp(audio.tmpDir, "asr-audio-*.wav")
	if err != nil {
		return "", nil, fmt.Errorf("create temp audio file: %w", err)
	}
	outputPath := outputFile.Name()
	// Close immediately — ffmpeg writes to the path, not the descriptor.
	outputFile.Close()

	cleanup := func() { os.Remove(outputPath) }

	if err := extractAudio(ctx, sourcePath, outputPath); err != nil {
		cleanup()
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", nil, err
		}
		return "", nil, fmt.Errorf("extract audio from %s: %w", assetID, err)
	}

	return outputPath, cleanup, nil
}

// stderrLimit bounds the stderr buffer. ffmpeg can produce megabytes of warnings on a corrupt
// source; collecting without a limit risks OOM under the container's mem_limit.
const stderrLimit = 64 * 1024

// extractAudio runs ffmpeg to produce 16 kHz mono WAV. This is the format faster-whisper
// expects: 16000 Hz sample rate, single channel. WAV avoids any codec dependency on the
// consumer side and keeps the extraction step lossless (PCM s16le).
func extractAudio(ctx context.Context, sourcePath, outputPath string) error {
	arguments := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", sourcePath,
		"-vn",
		"-ac", "1",
		"-ar", "16000",
		"-c:a", "pcm_s16le",
		"-f", "wav",
		"--", outputPath,
	}

	command := exec.CommandContext(ctx, "ffmpeg", arguments...)
	command.WaitDelay = processWaitDelay

	stderr := &limitedBuffer{limit: stderrLimit}
	command.Stderr = stderr

	if err := command.Run(); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		return fmt.Errorf("%w: %s", err, stderr.buf.String())
	}

	// Verify the output is non-empty. A zero-byte WAV means ffmpeg found no audio stream
	// in the source, which is an ordinary rejection (the media module accepts video-only
	// sources for transcoding but they have nothing to recognise).
	info, err := os.Stat(outputPath)
	if err != nil || info.Size() == 0 {
		return fmt.Errorf("ffmpeg produced no audio output")
	}

	return nil
}

// limitedBuffer collects up to `limit` bytes. Unlike the whisper adapter's boundedWriter it
// needs no mutex: ffmpeg's stderr pipe is consumed by exec.Cmd's single copying goroutine.
type limitedBuffer struct {
	limit int
	buf   bytes.Buffer
}

func (lb *limitedBuffer) Write(p []byte) (int, error) {
	n := len(p)
	remaining := lb.limit - lb.buf.Len()
	if remaining <= 0 {
		return n, nil
	}
	if len(p) > remaining {
		p = p[:remaining]
	}
	lb.buf.Write(p)
	return n, nil
}
