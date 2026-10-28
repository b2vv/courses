package domain

import (
	"slices"
	"time"
)

const (
	ContentTypeJPEG = "image/jpeg"
	ContentTypePNG  = "image/png"
	ContentTypeWebP = "image/webp"

	SniffBytes               = 12
	MaximumImageBytes  int64 = 5 << 20
	MaximumImagePixels       = 40_000_000
)

// ContentTypes is the ordered set accepted by image validation.
var ContentTypes = []string{ContentTypeJPEG, ContentTypePNG, ContentTypeWebP}

type Image struct {
	ID          string    `json:"id"`
	LessonID    string    `json:"lessonId"`
	ContentType string    `json:"contentType"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	ByteSize    int64     `json:"byteSize"`
	CreatedBy   string    `json:"createdBy"`
	CreatedAt   time.Time `json:"createdAt"`
}

func DetectContentType(head []byte) (string, error) {
	if len(head) < SniffBytes {
		return "", ErrUnsupportedImage
	}
	contentType := ""
	if head[0] == 0xFF && head[1] == 0xD8 && head[2] == 0xFF {
		contentType = ContentTypeJPEG
	}
	if string(head[:8]) == "\x89PNG\r\n\x1a\n" {
		contentType = ContentTypePNG
	}
	if string(head[:4]) == "RIFF" && string(head[8:12]) == "WEBP" {
		contentType = ContentTypeWebP
	}
	if slices.Contains(ContentTypes, contentType) {
		return contentType, nil
	}
	return "", ErrUnsupportedImage
}

// ExtensionFor maps a validated content type to the file extension used on disk and in the serve
// URL (/media/i/<id>.<ext>), so nginx derives the Content-Type by extension under nosniff.
func ExtensionFor(contentType string) (string, error) {
	switch contentType {
	case ContentTypeJPEG:
		return "jpg", nil
	case ContentTypePNG:
		return "png", nil
	case ContentTypeWebP:
		return "webp", nil
	default:
		return "", ErrUnsupportedImage
	}
}

func AcceptDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return ErrInvalidImage
	}
	if int64(width)*int64(height) > MaximumImagePixels {
		return ErrImageTooManyPixels
	}
	return nil
}

func NewImage(id, lessonID, contentType, createdBy string, width, height int, byteSize int64, createdAt time.Time) (Image, error) {
	if id == "" || lessonID == "" || createdBy == "" {
		return Image{}, ErrInvalidImage
	}
	if !slices.Contains(ContentTypes, contentType) {
		return Image{}, ErrUnsupportedImage
	}
	if err := AcceptDimensions(width, height); err != nil {
		return Image{}, err
	}
	if byteSize <= 0 {
		return Image{}, ErrInvalidImage
	}
	if byteSize > MaximumImageBytes {
		return Image{}, ErrImageTooLarge
	}
	return Image{
		ID: id, LessonID: lessonID, ContentType: contentType,
		Width: width, Height: height, ByteSize: byteSize,
		CreatedBy: createdBy, CreatedAt: createdAt,
	}, nil
}
