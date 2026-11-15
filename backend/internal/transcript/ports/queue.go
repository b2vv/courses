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
	TrackID    string
	Language   domain.Language
	TaskMarker string

	// Requested is true for the author's Regenerate and false for the pending sweep: only a
	// requested job may replace an existing draft of this track (O7).
	Requested bool
}

type RecognitionQueue interface {
	Enqueue(ctx context.Context, job RecognitionJob, timeout time.Duration) error
}
