package ffmpeg

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"courses/internal/media/domain"
)

const (
	processWaitDelay = 2 * time.Second
	directoryMode    = 0o750
)

var errInvalidTranscodeInput = errors.New("invalid transcoder input")

type Transcoder struct {
	root string
}

func New(root string) *Transcoder {
	return &Transcoder{root: filepath.Clean(root)}
}

func (transcoder *Transcoder) Probe(ctx context.Context, source domain.StorageKey) (domain.SourceProbe, error) {
	arguments, err := transcoder.probeArguments(source)
	if err != nil {
		return domain.SourceProbe{}, err
	}
	var output bytes.Buffer
	if err := runCommand(ctx, "ffprobe", arguments, &output, nil); err != nil {
		return domain.SourceProbe{}, fmt.Errorf("ffprobe failed: %w", err)
	}
	probe, err := parseProbe(output.Bytes())
	if err != nil {
		return domain.SourceProbe{}, fmt.Errorf("invalid ffprobe output: %w", err)
	}
	return probe, nil
}

func (transcoder *Transcoder) Render(ctx context.Context, source, work domain.StorageKey, renditions []domain.Rendition) error {
	arguments, err := transcoder.renderArguments(source, work, renditions)
	if err != nil {
		return err
	}
	workPath, err := transcoder.resolveWork(work, false)
	if err != nil {
		return err
	}
	for _, directory := range append([]string{"audio"}, renditionNames(renditions)...) {
		if err := os.MkdirAll(filepath.Join(workPath, directory), directoryMode); err != nil {
			return fmt.Errorf("prepare transcode output: %w", err)
		}
	}
	if err := runCommand(ctx, "ffmpeg", arguments, io.Discard, nil); err != nil {
		return fmt.Errorf("ffmpeg render failed: %w", err)
	}
	return nil
}

func (transcoder *Transcoder) Thumbnail(ctx context.Context, source, destination domain.StorageKey, probe domain.SourceProbe) error {
	arguments, err := transcoder.thumbnailArguments(source, destination, probe)
	if err != nil {
		return err
	}
	destinationPath, err := transcoder.resolveWork(destination, true)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), directoryMode); err != nil {
		return fmt.Errorf("prepare thumbnail output: %w", err)
	}
	if err := runCommand(ctx, "ffmpeg", arguments, io.Discard, nil); err != nil {
		return fmt.Errorf("ffmpeg thumbnail failed: %w", err)
	}
	return nil
}

func (transcoder *Transcoder) probeArguments(source domain.StorageKey) ([]string, error) {
	sourcePath, err := transcoder.resolveSource(source)
	if err != nil {
		return nil, err
	}
	return []string{"-v", "error", "-show_streams", "-show_format", "-print_format", "json", "--", sourcePath}, nil
}

func (transcoder *Transcoder) renderArguments(source, work domain.StorageKey, renditions []domain.Rendition) ([]string, error) {
	sourcePath, err := transcoder.resolveSource(source)
	if err != nil {
		return nil, err
	}
	workPath, err := transcoder.resolveWork(work, false)
	if err != nil || source.ID != work.ID {
		return nil, errInvalidTranscodeInput
	}
	settings, err := settingsFor(renditions)
	if err != nil {
		return nil, err
	}

	arguments := []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", sourcePath,
		"-filter_complex", filterComplex(settings),
	}
	for index, setting := range settings {
		stream := strconv.Itoa(index)
		arguments = append(arguments,
			"-map", "[v"+stream+"]",
			"-c:v:"+stream, "libx264",
			"-profile:v:"+stream, "high",
		)
		if setting.level != "" {
			arguments = append(arguments, "-level:v:"+stream, setting.level)
		}
		arguments = append(arguments,
			"-pix_fmt:v:"+stream, "yuv420p",
			"-crf:v:"+stream, "21",
			"-maxrate:v:"+stream, setting.maximumRate,
			"-bufsize:v:"+stream, setting.bufferSize,
			"-preset:v:"+stream, "veryfast",
			"-force_key_frames:v:"+stream, "expr:gte(t,n_forced*2)",
		)
	}
	arguments = append(arguments,
		"-map", "0:a:0",
		"-c:a:0", "aac",
		"-profile:a:0", "aac_low",
		"-b:a:0", "128k",
		"-ac", "2",
		"-f", "hls",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-hls_segment_type", "fmp4",
		"-hls_fmp4_init_filename", "init.mp4",
		"-hls_segment_filename", filepath.Join(workPath, "%v", "segment_%05d.m4s"),
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", variantMap(settings),
		filepath.Join(workPath, "%v", "index.m3u8"),
	)
	return arguments, nil
}

func (transcoder *Transcoder) thumbnailArguments(source, destination domain.StorageKey, probe domain.SourceProbe) ([]string, error) {
	sourcePath, err := transcoder.resolveSource(source)
	if err != nil {
		return nil, err
	}
	destinationPath, err := transcoder.resolveWork(destination, true)
	if err != nil || source.ID != destination.ID {
		return nil, errInvalidTranscodeInput
	}
	seek := "1"
	if probe.DurationSeconds < 1 {
		seek = "0"
	}
	return []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-ss", seek,
		"-i", sourcePath,
		"-frames:v", "1",
		"-vf", "scale='min(640,iw)':-2",
		"-f", "image2",
		destinationPath,
	}, nil
}

type probeDocument struct {
	Streams []struct {
		CodecName string `json:"codec_name"`
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
	} `json:"format"`
}

func parseProbe(document []byte) (domain.SourceProbe, error) {
	var decoded probeDocument
	if err := json.Unmarshal(document, &decoded); err != nil {
		return domain.SourceProbe{}, errInvalidTranscodeInput
	}
	duration, err := strconv.ParseFloat(decoded.Format.Duration, 64)
	if err != nil || duration < 0 || math.IsNaN(duration) || math.IsInf(duration, 0) {
		return domain.SourceProbe{}, errInvalidTranscodeInput
	}
	probe := domain.SourceProbe{DurationSeconds: duration}
	for _, stream := range decoded.Streams {
		switch stream.CodecType {
		case "video":
			if probe.VideoCodec == "" {
				probe.Width = stream.Width
				probe.Height = stream.Height
				probe.VideoCodec = stream.CodecName
			}
		case "audio":
			if probe.AudioCodec == "" {
				probe.AudioCodec = stream.CodecName
			}
		}
	}
	return probe, nil
}

type renditionSetting struct {
	name        string
	width       int
	height      int
	maximumRate string
	bufferSize  string
	level       string
}

func settingsFor(renditions []domain.Rendition) ([]renditionSetting, error) {
	fixed := []struct {
		rendition domain.Rendition
		setting   renditionSetting
	}{
		// B-19: 365k starved the floor rank on demanding content (measured VMAF 33 on a high-motion
		// 1080p source, ~60 on natural film per docs/media/research-codecs.md §6). 600k lifts it
		// (~+14 VMAF on the same source) while staying a genuine low-bandwidth rung reachable on a
		// ~700 kbps link and far below 720p's 3000k, so ABR still has a real lower level.
		{domain.Rendition360p, renditionSetting{name: "360p", width: 640, height: 360, maximumRate: "600k", bufferSize: "1200k", level: "4.1"}},
		{domain.Rendition720p, renditionSetting{name: "720p", width: 1280, height: 720, maximumRate: "3000k", bufferSize: "6000k", level: "4.1"}},
		{domain.Rendition1080p, renditionSetting{name: "1080p", width: 1920, height: 1080, maximumRate: "6000k", bufferSize: "12000k"}},
	}
	if len(renditions) == 0 || len(renditions) > len(fixed) {
		return nil, errInvalidTranscodeInput
	}
	settings := make([]renditionSetting, len(renditions))
	for index, rendition := range renditions {
		if rendition != fixed[index].rendition {
			return nil, errInvalidTranscodeInput
		}
		settings[index] = fixed[index].setting
	}
	return settings, nil
}

func filterComplex(settings []renditionSetting) string {
	var filter strings.Builder
	fmt.Fprintf(&filter, "[0:v]split=%d", len(settings))
	for index := range settings {
		fmt.Fprintf(&filter, "[in%d]", index)
	}
	for index, setting := range settings {
		fmt.Fprintf(&filter, ";[in%d]scale=w='min(%d,iw)':h='min(%d,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v%d]", index, setting.width, setting.height, index)
	}
	return filter.String()
}

func variantMap(settings []renditionSetting) string {
	variants := []string{"a:0,agroup:aac,name:audio,default:yes"}
	for index, setting := range settings {
		variants = append(variants, fmt.Sprintf("v:%d,agroup:aac,name:%s", index, setting.name))
	}
	return strings.Join(variants, " ")
}

func renditionNames(renditions []domain.Rendition) []string {
	names := make([]string, len(renditions))
	for index, rendition := range renditions {
		names[index] = string(rendition)
	}
	return names
}

func (transcoder *Transcoder) resolveSource(key domain.StorageKey) (string, error) {
	if key.Kind != domain.StorageSource || key.Leaf != "original" || !validUUID(key.ID) {
		return "", errInvalidTranscodeInput
	}
	return filepath.Join(transcoder.root, string(key.Kind), key.ID, key.Leaf), nil
}

func (transcoder *Transcoder) resolveWork(key domain.StorageKey, thumbnail bool) (string, error) {
	if key.Kind != domain.StorageTemporary || !validUUID(key.ID) {
		return "", errInvalidTranscodeInput
	}
	leaf := key.Leaf
	if thumbnail {
		if !strings.HasSuffix(leaf, ".jpg") {
			return "", errInvalidTranscodeInput
		}
		leaf = strings.TrimSuffix(leaf, ".jpg")
	}
	if !validClaimToken(leaf) {
		return "", errInvalidTranscodeInput
	}
	name := key.ID + "-" + leaf
	if thumbnail {
		name += ".jpg"
	}
	return filepath.Join(transcoder.root, string(key.Kind), name), nil
}

func validUUID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}

func validClaimToken(value string) bool {
	if len(value) != 32 || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == 16
}

func runCommand(ctx context.Context, name string, arguments []string, stdout io.Writer, started chan<- struct{}) error {
	command := exec.CommandContext(ctx, name, arguments...)
	command.WaitDelay = processWaitDelay
	command.Stdout = stdout
	if err := command.Start(); err != nil {
		return err
	}
	if started != nil {
		close(started)
	}
	err := command.Wait()
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	return err
}
