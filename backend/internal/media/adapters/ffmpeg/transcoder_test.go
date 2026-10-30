package ffmpeg

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"courses/internal/media/domain"
)

const (
	testAssetID   = "11111111-1111-4111-8111-111111111111"
	testClaimLeaf = "0123456789abcdef0123456789abcdef"
)

func TestEnforceTreeModesMakesOutputNotWorldReadable(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "360p")
	if err := os.MkdirAll(sub, 0o777); err != nil {
		t.Fatal(err)
	}
	files := []string{
		filepath.Join(root, "master.m3u8"),
		filepath.Join(sub, "index.m3u8"),
		filepath.Join(sub, "init.mp4"),
		filepath.Join(sub, "segment_00000.m4s"),
	}
	for _, file := range files {
		if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := enforceTreeModes(root); err != nil {
		t.Fatal(err)
	}

	for _, file := range files {
		info, err := os.Stat(file)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != fileMode {
			t.Fatalf("%s mode=%o, want %o", file, info.Mode().Perm(), fileMode)
		}
	}
	for _, directory := range []string{root, sub} {
		info, err := os.Stat(directory)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != directoryMode {
			t.Fatalf("%s mode=%o, want %o", directory, info.Mode().Perm(), directoryMode)
		}
	}
}

func TestParseProbeAndAcceptSource(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name      string
		document  string
		want      domain.SourceProbe
		wantError error
	}{
		{
			name:     "valid H264 AAC source",
			document: `{"streams":[{"codec_name":"h264","codec_type":"video","width":1920,"height":1040,"ignored":"value"},{"codec_name":"aac","codec_type":"audio"}],"format":{"duration":"12.375"},"ignored":true}`,
			want:     domain.SourceProbe{Width: 1920, Height: 1040, VideoCodec: domain.CodecH264, AudioCodec: domain.CodecAAC, DurationSeconds: 12.375},
		},
		{
			name:      "non video source",
			document:  `{"streams":[],"format":{"duration":"0.25"}}`,
			wantError: domain.ErrUnsupportedSource,
		},
		{
			name:      "video without audio",
			document:  `{"streams":[{"codec_name":"h264","codec_type":"video","width":640,"height":360}],"format":{"duration":"4.0"}}`,
			wantError: domain.ErrUnsupportedSource,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			probe, err := parseProbe([]byte(test.document))
			if err == nil {
				err = domain.AcceptSource(probe)
			}
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error=%v want=%v", err, test.wantError)
			}
			if test.wantError == nil && probe != test.want {
				t.Fatalf("probe=%+v want=%+v", probe, test.want)
			}
		})
	}
}

func TestParseProbeRefusesMissingOrInvalidDuration(t *testing.T) {
	t.Parallel()
	for _, document := range []string{
		`{"streams":[{"codec_name":"h264","codec_type":"video","width":640,"height":360},{"codec_name":"aac","codec_type":"audio"}],"format":{}}`,
		`{"streams":[{"codec_name":"h264","codec_type":"video","width":640,"height":360},{"codec_name":"aac","codec_type":"audio"}],"format":{"duration":"NaN"}}`,
		`{"streams":[{"codec_name":"h264","codec_type":"video","width":640,"height":360},{"codec_name":"aac","codec_type":"audio"}],"format":{"duration":"-1"}}`,
	} {
		if _, err := parseProbe([]byte(document)); !errors.Is(err, errInvalidTranscodeInput) {
			t.Fatalf("parseProbe(%s) error=%v want=%v", document, err, errInvalidTranscodeInput)
		}
	}
}

func TestProbeArgumentsAreFixed(t *testing.T) {
	t.Parallel()
	transcoder := New("/media")
	got, err := transcoder.probeArguments(sourceKey())
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"-v", "error", "-show_streams", "-show_format", "-print_format", "json", "--", "/media/source/11111111-1111-4111-8111-111111111111/original"}
	assertArguments(t, got, want)
}

func TestRenderArgumentsForLandscapeAndPortraitSources(t *testing.T) {
	t.Parallel()
	transcoder := New("/media")
	for _, test := range []struct {
		name  string
		probe domain.SourceProbe
		want  []string
	}{
		{
			name:  "1920 by 1040",
			probe: domain.SourceProbe{Width: 1920, Height: 1040},
			want:  expectedTwoRungRenderArguments(),
		},
		{
			name:  "1080 by 1920 portrait",
			probe: domain.SourceProbe{Width: 1080, Height: 1920},
			want:  expectedThreeRungRenderArguments(),
		},
		{
			name:  "320 by 240 below the first rung",
			probe: domain.SourceProbe{Width: 320, Height: 240},
			want:  expectedOneRungRenderArguments(),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := transcoder.renderArguments(sourceKey(), workKey(), domain.SelectLadder(test.probe))
			if err != nil {
				t.Fatal(err)
			}
			assertArguments(t, got, test.want)
		})
	}
}

func TestRenderArgumentsPinAppleConformance(t *testing.T) {
	t.Parallel()
	transcoder := New("/media")
	got, err := transcoder.renderArguments(sourceKey(), workKey(), []domain.Rendition{
		domain.Rendition360p,
		domain.Rendition720p,
		domain.Rendition1080p,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertArguments(t, got, expectedThreeRungRenderArguments())
	joined := strings.Join(got, " ")
	for _, forbidden := range []string{"-level:v:2", "avc3", "frame_rate", "r_frame_rate"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("argv contains %q: %s", forbidden, joined)
		}
	}
}

func TestThumbnailArgumentsUseDurationWithoutUpscaling(t *testing.T) {
	t.Parallel()
	transcoder := New("/media")
	for _, test := range []struct {
		name     string
		duration float64
		seek     string
	}{
		{name: "short source", duration: 0.75, seek: "0"},
		{name: "one second source", duration: 1, seek: "1"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := transcoder.thumbnailArguments(sourceKey(), thumbnailWorkKey(), domain.SourceProbe{DurationSeconds: test.duration})
			if err != nil {
				t.Fatal(err)
			}
			assertArguments(t, got, expectedThumbnailArguments(test.seek))
		})
	}
}

func TestOriginalNameCannotReachRenderOrThumbnailArguments(t *testing.T) {
	t.Parallel()
	asset := domain.VideoAsset{ID: testAssetID, OriginalName: `"; rm -rf / --preset x.mp4`}
	transcoder := New("/media")
	render, err := transcoder.renderArguments(
		domain.StorageKey{Kind: domain.StorageSource, ID: asset.ID, Leaf: "original"},
		workKey(),
		[]domain.Rendition{domain.Rendition360p, domain.Rendition720p},
	)
	if err != nil {
		t.Fatal(err)
	}
	thumbnail, err := transcoder.thumbnailArguments(
		domain.StorageKey{Kind: domain.StorageSource, ID: asset.ID, Leaf: "original"},
		thumbnailWorkKey(),
		domain.SourceProbe{DurationSeconds: 2},
	)
	if err != nil {
		t.Fatal(err)
	}
	assertArguments(t, render, expectedTwoRungRenderArguments())
	assertArguments(t, thumbnail, expectedThumbnailArguments("1"))
	for _, argv := range [][]string{render, thumbnail} {
		if strings.Contains(strings.Join(argv, "\x00"), asset.OriginalName) {
			t.Fatalf("original filename reached argv: %q", argv)
		}
	}
}

func TestArgumentsRejectNonFixedStorageKeysAndLadders(t *testing.T) {
	t.Parallel()
	transcoder := New("/media")
	if _, err := transcoder.renderArguments(
		sourceKey(),
		domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID, Leaf: "../escape"},
		[]domain.Rendition{domain.Rendition360p},
	); err == nil {
		t.Fatal("unsafe work key accepted")
	}
	if _, err := transcoder.renderArguments(sourceKey(), workKey(), []domain.Rendition{domain.Rendition720p}); err == nil {
		t.Fatal("non-prefix ladder accepted")
	}
}

func TestCancellationEndsDirectProcessWithinWaitDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- runCommand(ctx, "/bin/sleep", []string{"30"}, io.Discard, started)
	}()

	select {
	case <-started:
	case err := <-done:
		t.Fatalf("/bin/sleep did not start: %v", err)
	case <-time.After(processWaitDelay):
		t.Fatal("timed out waiting for /bin/sleep to start")
	}

	cancelledAt := time.Now()
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("error=%v want context.Canceled", err)
		}
		if elapsed := time.Since(cancelledAt); elapsed > processWaitDelay {
			t.Fatalf("process exit took %s; WaitDelay=%s", elapsed, processWaitDelay)
		}
	case <-time.After(processWaitDelay):
		t.Fatalf("/bin/sleep remained alive beyond WaitDelay=%s", processWaitDelay)
	}
}

func expectedThumbnailArguments(seek string) []string {
	return []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-ss", seek,
		"-i", "/media/source/11111111-1111-4111-8111-111111111111/original",
		"-frames:v", "1",
		"-vf", "scale='min(640,iw)':-2",
		"-f", "image2",
		"/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef.jpg",
	}
}

func expectedTwoRungRenderArguments() []string {
	return []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", "/media/source/11111111-1111-4111-8111-111111111111/original",
		"-filter_complex", "[0:v]split=2[in0][in1];[in0]scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v0];[in1]scale=w='min(1280,iw)':h='min(720,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v1]",
		"-map", "[v0]", "-c:v:0", "libx264", "-profile:v:0", "high", "-level:v:0", "4.1", "-pix_fmt:v:0", "yuv420p", "-crf:v:0", "21", "-maxrate:v:0", "600k", "-bufsize:v:0", "1200k", "-preset:v:0", "veryfast", "-force_key_frames:v:0", "expr:gte(t,n_forced*2)",
		"-map", "[v1]", "-c:v:1", "libx264", "-profile:v:1", "high", "-level:v:1", "4.1", "-pix_fmt:v:1", "yuv420p", "-crf:v:1", "21", "-maxrate:v:1", "3000k", "-bufsize:v:1", "6000k", "-preset:v:1", "veryfast", "-force_key_frames:v:1", "expr:gte(t,n_forced*2)",
		"-map", "0:a:0", "-c:a:0", "aac", "-profile:a:0", "aac_low", "-b:a:0", "128k", "-ac", "2",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod", "-hls_flags", "independent_segments", "-hls_segment_type", "fmp4", "-hls_fmp4_init_filename", "init.mp4",
		"-hls_segment_filename", "/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef/%v/segment_%05d.m4s",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "a:0,agroup:aac,name:audio,default:yes v:0,agroup:aac,name:360p v:1,agroup:aac,name:720p",
		"/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef/%v/index.m3u8",
	}
}

func expectedOneRungRenderArguments() []string {
	return []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", "/media/source/11111111-1111-4111-8111-111111111111/original",
		"-filter_complex", "[0:v]split=1[in0];[in0]scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v0]",
		"-map", "[v0]", "-c:v:0", "libx264", "-profile:v:0", "high", "-level:v:0", "4.1", "-pix_fmt:v:0", "yuv420p", "-crf:v:0", "21", "-maxrate:v:0", "600k", "-bufsize:v:0", "1200k", "-preset:v:0", "veryfast", "-force_key_frames:v:0", "expr:gte(t,n_forced*2)",
		"-map", "0:a:0", "-c:a:0", "aac", "-profile:a:0", "aac_low", "-b:a:0", "128k", "-ac", "2",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod", "-hls_flags", "independent_segments", "-hls_segment_type", "fmp4", "-hls_fmp4_init_filename", "init.mp4",
		"-hls_segment_filename", "/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef/%v/segment_%05d.m4s",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "a:0,agroup:aac,name:audio,default:yes v:0,agroup:aac,name:360p",
		"/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef/%v/index.m3u8",
	}
}

func expectedThreeRungRenderArguments() []string {
	return []string{
		"-nostdin", "-hide_banner", "-loglevel", "error", "-y",
		"-i", "/media/source/11111111-1111-4111-8111-111111111111/original",
		"-filter_complex", "[0:v]split=3[in0][in1][in2];[in0]scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v0];[in1]scale=w='min(1280,iw)':h='min(720,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v1];[in2]scale=w='min(1920,iw)':h='min(1080,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2[v2]",
		"-map", "[v0]", "-c:v:0", "libx264", "-profile:v:0", "high", "-level:v:0", "4.1", "-pix_fmt:v:0", "yuv420p", "-crf:v:0", "21", "-maxrate:v:0", "600k", "-bufsize:v:0", "1200k", "-preset:v:0", "veryfast", "-force_key_frames:v:0", "expr:gte(t,n_forced*2)",
		"-map", "[v1]", "-c:v:1", "libx264", "-profile:v:1", "high", "-level:v:1", "4.1", "-pix_fmt:v:1", "yuv420p", "-crf:v:1", "21", "-maxrate:v:1", "3000k", "-bufsize:v:1", "6000k", "-preset:v:1", "veryfast", "-force_key_frames:v:1", "expr:gte(t,n_forced*2)",
		"-map", "[v2]", "-c:v:2", "libx264", "-profile:v:2", "high", "-pix_fmt:v:2", "yuv420p", "-crf:v:2", "21", "-maxrate:v:2", "6000k", "-bufsize:v:2", "12000k", "-preset:v:2", "veryfast", "-force_key_frames:v:2", "expr:gte(t,n_forced*2)",
		"-map", "0:a:0", "-c:a:0", "aac", "-profile:a:0", "aac_low", "-b:a:0", "128k", "-ac", "2",
		"-f", "hls", "-hls_time", "6", "-hls_playlist_type", "vod", "-hls_flags", "independent_segments", "-hls_segment_type", "fmp4", "-hls_fmp4_init_filename", "init.mp4",
		"-hls_segment_filename", "/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef/%v/segment_%05d.m4s",
		"-master_pl_name", "master.m3u8",
		"-var_stream_map", "a:0,agroup:aac,name:audio,default:yes v:0,agroup:aac,name:360p v:1,agroup:aac,name:720p v:2,agroup:aac,name:1080p",
		"/media/temporary/11111111-1111-4111-8111-111111111111-0123456789abcdef0123456789abcdef/%v/index.m3u8",
	}
}

func sourceKey() domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageSource, ID: testAssetID, Leaf: "original"}
}

func workKey() domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID, Leaf: testClaimLeaf}
}

func thumbnailWorkKey() domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: testAssetID, Leaf: testClaimLeaf + ".jpg"}
}

func assertArguments(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv mismatch\n got: %#v\nwant: %#v", got, want)
	}
}
