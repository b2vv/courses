package ports

import "context"

type Job struct {
	AssetID string
	// TrackID is empty for a video transcode job and set to an uploaded audio
	// track's id for an audio-track transcode job (the payload discriminator, D1).
	TrackID    string
	Attempt    int
	RetryCount int
	MaxRetry   int
}

type TranscodingQueue interface {
	Enqueue(context.Context, string, int) error
	// EnqueueAudioTrack schedules the transcode of one uploaded audio track; it
	// reuses the same queue as the video job, discriminated by the track id.
	EnqueueAudioTrack(ctx context.Context, assetID, trackID string, attempt int) error
}
