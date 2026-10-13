package domain

import (
	"errors"
	"testing"
	"time"
)

func TestDetectContentType(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		head []byte
		want string
		err  error
	}{
		{name: "JPEG", head: []byte("\xff\xd8\xff\x00\x00\x00\x00\x00\x00\x00\x00\x00"), want: ContentTypeJPEG},
		{name: "PNG", head: []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\x00"), want: ContentTypePNG},
		{name: "WebP", head: []byte("RIFF\x10\x00\x00\x00WEBPVP8 "), want: ContentTypeWebP},
		{name: "nil", err: ErrUnsupportedImage},
		{name: "short JPEG", head: []byte("\xff\xd8\xff\x00\x00\x00\x00\x00\x00\x00\x00"), err: ErrUnsupportedImage},
		{name: "short PNG", head: []byte("\x89PNG\r\n\x1a\n"), err: ErrUnsupportedImage},
		{name: "unknown", head: []byte("unknown type"), err: ErrUnsupportedImage},
		{name: "other RIFF", head: []byte("RIFF\x10\x00\x00\x00WAVE"), err: ErrUnsupportedImage},
		{name: "invalid PNG signature", head: []byte("\x89PNG\r\n\x1aX\x00\x00\x00\x00"), err: ErrUnsupportedImage},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := DetectContentType(test.head)
			if got != test.want || !errors.Is(err, test.err) {
				t.Fatalf("got (%q, %v), want (%q, %v)", got, err, test.want, test.err)
			}
		})
	}
}

func TestAcceptDimensions(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name          string
		width, height int
		want          error
	}{
		{name: "zero width", width: 0, height: 1, want: ErrInvalidImage},
		{name: "zero height", width: 1, height: 0, want: ErrInvalidImage},
		{name: "negative width", width: -1, height: 1, want: ErrInvalidImage},
		{name: "negative height", width: 1, height: -1, want: ErrInvalidImage},
		{name: "small", width: 1, height: 1},
		{name: "at budget", width: 8000, height: 5000},
		{name: "over budget", width: 8000, height: 5001, want: ErrImageTooManyPixels},
		{name: "one pixel over budget", width: MaximumImagePixels + 1, height: 1, want: ErrImageTooManyPixels},
		{name: "int32 overflow", width: 60000, height: 60000, want: ErrImageTooManyPixels},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := AcceptDimensions(test.width, test.height); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestNewImage(t *testing.T) {
	t.Parallel()

	valid := Image{
		ID: "image-id", LessonID: "lesson-id", ContentType: ContentTypeJPEG,
		Width: 1920, Height: 1080, ByteSize: 1024,
		CreatedBy: "author-id", CreatedAt: time.Date(2026, time.October, 12, 10, 0, 0, 0, time.UTC),
	}
	for _, test := range []struct {
		name   string
		change func(*Image)
		want   error
	}{
		{name: "happy path"},
		{name: "PNG", change: func(i *Image) { i.ContentType = ContentTypePNG }},
		{name: "WebP", change: func(i *Image) { i.ContentType = ContentTypeWebP }},
		{name: "empty id", change: func(i *Image) { i.ID = "" }, want: ErrInvalidImage},
		{name: "empty lesson", change: func(i *Image) { i.LessonID = "" }, want: ErrInvalidImage},
		{name: "empty creator", change: func(i *Image) { i.CreatedBy = "" }, want: ErrInvalidImage},
		{name: "GIF", change: func(i *Image) { i.ContentType = "image/gif" }, want: ErrUnsupportedImage},
		{name: "zero bytes", change: func(i *Image) { i.ByteSize = 0 }, want: ErrInvalidImage},
		{name: "negative bytes", change: func(i *Image) { i.ByteSize = -1 }, want: ErrInvalidImage},
		{name: "too large", change: func(i *Image) { i.ByteSize = MaximumImageBytes + 1 }, want: ErrImageTooLarge},
		{name: "maximum bytes", change: func(i *Image) { i.ByteSize = MaximumImageBytes }},
		{name: "invalid dimensions", change: func(i *Image) { i.Width = 0 }, want: ErrInvalidImage},
		{name: "oversized dimensions", change: func(i *Image) { i.Width, i.Height = 8000, 5001 }, want: ErrImageTooManyPixels},
		{name: "identity before type", change: func(i *Image) { i.ID, i.ContentType = "", "image/gif" }, want: ErrInvalidImage},
		{name: "type before dimensions", change: func(i *Image) { i.ContentType, i.Width = "image/gif", 0 }, want: ErrUnsupportedImage},
		{name: "dimensions before size", change: func(i *Image) { i.Width, i.Height, i.ByteSize = 8000, 5001, 0 }, want: ErrImageTooManyPixels},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			input := valid
			if test.change != nil {
				test.change(&input)
			}
			got, err := NewImage(input.ID, input.LessonID, input.ContentType, input.CreatedBy, input.Width, input.Height, input.ByteSize, input.CreatedAt)
			if !errors.Is(err, test.want) {
				t.Fatalf("got error %v, want %v", err, test.want)
			}
			if test.want != nil {
				if got != (Image{}) {
					t.Fatalf("invalid input returned image %+v", got)
				}
				return
			}
			if got != input {
				t.Fatalf("got %+v, want %+v", got, input)
			}
		})
	}
}
