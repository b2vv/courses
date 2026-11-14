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
	TaskMarker string

	// Requested is true for the author's Regenerate and false for the pending sweep: only a
	// requested job may re-claim a same-generation draft that already heard the default track
	// (multi-audio S6, D-S6-2).
	Requested bool
}

type RecognitionQueue interface {
	Enqueue(ctx context.Context, job RecognitionJob, timeout time.Duration) error
}
