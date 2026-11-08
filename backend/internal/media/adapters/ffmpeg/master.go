package ffmpeg

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
)

// audioGroupID is the HLS audio group ffmpeg emits for "agroup:aac"; every audio
// EXT-X-MEDIA and every video EXT-X-STREAM-INF reference it so a player can pair
// any variant with any track.
const audioGroupID = "group_aac"

// RegenerateMaster rewrites the published master.m3u8 of the asset from the given
// Ready audio set (D7). It is a full rewrite of the audio section — never an
// append — reading the current master, replacing every EXT-X-MEDIA:TYPE=AUDIO
// line with one line per Ready track, and ensuring each video EXT-X-STREAM-INF
// advertises the audio group and codec. The ffmpeg-measured video variant lines
// (BANDWIDTH/RESOLUTION) are preserved verbatim. The write is atomic (temp file
// in the same directory + rename) so nginx never serves a half-written master,
// and idempotent so a retried job reproduces the same file. The caller serialises
// invocations per asset.
func (transcoder *Transcoder) RegenerateMaster(ctx context.Context, assetID string, audios []ports.AudioRendition) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if !validUUID(assetID) || len(audios) == 0 {
		return errInvalidTranscodeInput
	}
	masterPath := filepath.Join(transcoder.root, string(domain.StorageHLS), assetID, "master.m3u8")
	existing, err := os.ReadFile(masterPath)
	if err != nil {
		return fmt.Errorf("read master: %w", err)
	}
	rewritten, err := rewriteMasterText(string(existing), audios)
	if err != nil {
		return err
	}
	if err := writeFileAtomic(masterPath, []byte(rewritten), fileMode); err != nil {
		return fmt.Errorf("write master: %w", err)
	}
	return nil
}

// rewriteMasterText replaces the audio section of a master playlist with one
// EXT-X-MEDIA line per Ready track, preserving the header and the video
// EXT-X-STREAM-INF blocks (and their following URI lines) exactly as ffmpeg wrote
// them. A video-only base master (no audio group) is upgraded in place: each
// STREAM-INF gains AUDIO="group_aac" and an mp4a codec. It requires at least one
// audio rendition — the D4 invariant forbids a Ready asset without a Ready track,
// so the master is never regenerated empty.
func rewriteMasterText(existing string, audios []ports.AudioRendition) (string, error) {
	if len(audios) == 0 {
		return "", errInvalidTranscodeInput
	}
	lines := strings.Split(existing, "\n")
	head := make([]string, 0, len(lines))
	body := make([]string, 0, len(lines))
	seenStream := false
	for _, line := range lines {
		if !seenStream && strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			seenStream = true
		}
		if !seenStream {
			if strings.HasPrefix(line, "#EXT-X-MEDIA:") {
				continue // drop the existing audio section; it is rebuilt below
			}
			head = append(head, line)
			continue
		}
		if strings.HasPrefix(line, "#EXT-X-STREAM-INF:") {
			line = ensureAudioGroup(line)
		}
		body = append(body, line)
	}
	if !seenStream {
		return "", errInvalidTranscodeInput
	}
	media := make([]string, 0, len(audios))
	for _, audio := range audios {
		media = append(media, formatMediaLine(audio))
	}
	out := make([]string, 0, len(head)+len(media)+len(body))
	out = append(out, head...)
	out = append(out, media...)
	out = append(out, body...)
	return strings.Join(out, "\n"), nil
}

func formatMediaLine(audio ports.AudioRendition) string {
	def := "NO"
	if audio.Default {
		def = "YES"
	}
	return `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="` + audioGroupID +
		`",NAME="` + audio.Name +
		`",LANGUAGE="` + audio.Language +
		`",DEFAULT=` + def +
		`,AUTOSELECT=YES,CHANNELS="2",URI="` + audioURI(audio) + `"`
}

func audioURI(audio ports.AudioRendition) string {
	if audio.Embedded {
		return "audio/index.m3u8"
	}
	return domain.AudioLeafPrefix + audio.TrackID + "/index.m3u8"
}

// ensureAudioGroup makes a video EXT-X-STREAM-INF line reference the audio group:
// it adds AUDIO="group_aac" when absent (a video-only base master) and an mp4a
// codec to CODECS when missing. It is a no-op on a line that already carries both,
// so an audio-carrying base master is preserved verbatim.
func ensureAudioGroup(streamInf string) string {
	streamInf = ensureCodecsHaveAudio(streamInf)
	if !strings.Contains(streamInf, "AUDIO=") {
		streamInf += `,AUDIO="` + audioGroupID + `"`
	}
	return streamInf
}

func ensureCodecsHaveAudio(line string) string {
	const key = `CODECS="`
	index := strings.Index(line, key)
	if index < 0 {
		return line
	}
	start := index + len(key)
	closing := strings.Index(line[start:], `"`)
	if closing < 0 {
		return line
	}
	if strings.Contains(line[start:start+closing], "mp4a") {
		return line
	}
	return line[:start+closing] + ",mp4a.40.2" + line[start+closing:]
}

// writeFileAtomic writes data to a temporary file in the destination directory,
// gives it mode, then renames it over path. The rename is atomic within the
// directory's filesystem, so a reader (nginx) sees either the old or the new
// master, never a partial one.
func writeFileAtomic(path string, data []byte, mode os.FileMode) error {
	directory := filepath.Dir(path)
	temporary, err := os.CreateTemp(directory, ".master-*.tmp")
	if err != nil {
		return err
	}
	name := temporary.Name()
	defer os.Remove(name)
	if _, err := temporary.Write(data); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Chmod(mode); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(name, path)
}
