package ports

import (
	"context"
	"time"

	"courses/internal/transcript/domain"
)

type RecognitionJob struct {
	LessonID   string
	AssetID    string
	Generation int32
	Language   domain.Language
}

type RecognitionQueue interface {
	Enqueue(ctx context.Context, job RecognitionJob, timeout time.Duration) error
}
