package ports

import (
	"context"

	"courses/internal/transcript/domain"
)

// LessonTranscript is what one read returns: the text, the course facts the access rule needs,
// and the generation the player would pick. Staleness is not stored, so the repository hands
// back the two generations and the domain compares them.
type LessonTranscript struct {
	Found      bool
	Course     domain.Course
	Transcript domain.Transcript

	// ReadyGeneration is the generation of the asset the player would pick, and zero when the
	// lesson has no ready asset. Only a ready asset counts: a lesson whose first video is still
	// uploading, or whose transcode failed, must not read as stale.
	ReadyGeneration int32
}

type TranscriptRepository interface {
	// GetLessonTranscript returns the transcript of one lesson. An absent transcript is not an
	// error: a lesson simply may not have one yet, and the caller renders that as a state.
	// Language may be empty, in which case the repository picks the only one, or the first by
	// normalised code when several exist.
	GetLessonTranscript(ctx context.Context, lessonID string, language domain.Language) (LessonTranscript, error)
}
