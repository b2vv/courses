package domain_test

import (
	"errors"
	"testing"

	"courses/internal/media/domain"
)

func TestAcceptSource(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		probe domain.SourceProbe
		want  error
	}{
		{name: "H.264 and AAC", probe: domain.SourceProbe{Width: 1920, Height: 1080, VideoCodec: domain.CodecH264, AudioCodec: domain.CodecAAC}},
		{name: "HEVC and AAC", probe: domain.SourceProbe{Width: 1920, Height: 1080, VideoCodec: domain.CodecHEVC, AudioCodec: domain.CodecAAC}},
		{name: "H.264 and MP3", probe: domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: domain.CodecH264, AudioCodec: domain.CodecMP3}},
		{name: "H.264 and Opus", probe: domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: domain.CodecH264, AudioCodec: "opus"}, want: domain.ErrUnsupportedSource},
		{name: "video without audio", probe: domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: domain.CodecH264}, want: domain.ErrUnsupportedSource},
		{name: "audio without video", probe: domain.SourceProbe{AudioCodec: domain.CodecAAC}, want: domain.ErrUnsupportedSource},
		{name: "unknown video codec", probe: domain.SourceProbe{Width: 1280, Height: 720, VideoCodec: "vp9", AudioCodec: domain.CodecAAC}, want: domain.ErrUnsupportedSource},
		{name: "zero dimensions", probe: domain.SourceProbe{VideoCodec: domain.CodecH264, AudioCodec: domain.CodecAAC}, want: domain.ErrUnsupportedSource},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.AcceptSource(test.probe); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestSourceProbeUsableDuration(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		duration float64
		want     *float64
	}{
		{name: "negative is absent", duration: -1},
		{name: "zero is absent"},
		{name: "positive is present", duration: 61.25, want: durationPointer(61.25)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := (domain.SourceProbe{DurationSeconds: test.duration}).UsableDurationSeconds()
			if got == nil || test.want == nil {
				if got != nil || test.want != nil {
					t.Fatalf("got %v, want %v", got, test.want)
				}
				return
			}
			if *got != *test.want {
				t.Fatalf("got %v, want %v", *got, *test.want)
			}
		})
	}
}

func TestAcceptContainer(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		head []byte
		want error
	}{
		{name: "MP4 isom", head: []byte("\x00\x00\x00\x18ftypisom")},
		{name: "MP4 iso2", head: []byte("\x00\x00\x00\x18ftypiso2")},
		{name: "MP4 iso4", head: []byte("\x00\x00\x00\x18ftypiso4")},
		{name: "MP4 iso5", head: []byte("\x00\x00\x00\x18ftypiso5")},
		{name: "MP4 iso6", head: []byte("\x00\x00\x00\x18ftypiso6")},
		{name: "MP4 mp41", head: []byte("\x00\x00\x00\x18ftypmp41")},
		{name: "MP4 mp42", head: []byte("\x00\x00\x00\x18ftypmp42")},
		{name: "MP4 avc1", head: []byte("\x00\x00\x00\x18ftypavc1")},
		{name: "MP4 mmp4", head: []byte("\x00\x00\x00\x18ftypmmp4")},
		{name: "MP4 dash", head: []byte("\x00\x00\x00\x18ftypdash")},
		{name: "MP4 M4V", head: []byte("\x00\x00\x00\x18ftypM4V ")},
		{name: "QuickTime", head: []byte("\x00\x00\x00\x18ftypqt  ")},
		{name: "Matroska or WebM", head: []byte("\x1a\x45\xdf\xa3\x00\x00\x00\x00\x00\x00\x00\x00")},
		{name: "unlisted brand", head: []byte("\x00\x00\x00\x18ftypwebm"), want: domain.ErrUnsupportedSource},
		{name: "plain text", head: []byte("plain text!!"), want: domain.ErrUnsupportedSource},
		{name: "short EBML head", head: []byte("\x1a\x45\xdf\xa3\x00\x00\x00\x00\x00\x00\x00"), want: domain.ErrUnsupportedSource},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := domain.AcceptContainer(test.head); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestSelectLadder(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		width  int
		height int
		want   []domain.Rendition
	}{
		{name: "landscape 360p", width: 640, height: 360, want: []domain.Rendition{domain.Rendition360p}},
		{name: "landscape 720p", width: 1280, height: 720, want: []domain.Rendition{domain.Rendition360p, domain.Rendition720p}},
		{name: "landscape 1080p", width: 1920, height: 1080, want: []domain.Rendition{domain.Rendition360p, domain.Rendition720p, domain.Rendition1080p}},
		{name: "portrait", width: 1080, height: 1920, want: []domain.Rendition{domain.Rendition360p, domain.Rendition720p, domain.Rendition1080p}},
		{name: "below 360p", width: 320, height: 240, want: []domain.Rendition{domain.Rendition360p}},
		{name: "exactly 360p", width: 480, height: 360, want: []domain.Rendition{domain.Rendition360p}},
		{name: "4K", width: 3840, height: 2160, want: []domain.Rendition{domain.Rendition360p, domain.Rendition720p, domain.Rendition1080p}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := domain.SelectLadder(domain.SourceProbe{Width: test.width, Height: test.height})
			if !sameRenditions(got, test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
			if len(got) == 0 {
				t.Fatal("ladder must always contain at least one rung")
			}
			for _, rung := range got {
				if test.height >= 360 && rung.Height() > test.height {
					t.Errorf("rung %s upscales source height %d", rung, test.height)
				}
			}
		})
	}
}

func sameRenditions(got, want []domain.Rendition) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func durationPointer(value float64) *float64 { return &value }
