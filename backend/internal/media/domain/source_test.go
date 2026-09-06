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
