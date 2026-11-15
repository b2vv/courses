package ports

import (
	"context"
	"time"

	"courses/internal/transcript/domain"
)

// LessonDraft is one track's machine output with its attempt token and CURRENT metadata.
// The snapshot in Draft and the current metadata derive staleness without a stored flag.
// Only tracks of the latest Ready, non-evicted asset are returned, so an older video's draft
// cannot be offered after replacement. IsDefault keeps the legacy read usable until MS2c.
type LessonDraft struct {
	Draft             domain.TranscriptDraft
	UpdatedAt         string
	TrackLanguage     domain.Language
	TrackMultilingual bool
	IsDefault         bool
}

// PendingRecognition is one ready track of the latest Ready non-evicted asset with no draft.
// Renaming or changing multilingual does not create pending work (O7). It is the whole of the
// trigger (plan P-1): there is no event and no outbox in
// this repository, so "start recognition" is a QUESTION ABOUT STATE, asked repeatedly. A crash
// between publishing an asset and enqueueing a task therefore loses nothing - the next sweep
// asks again.
type PendingRecognition struct {
	LessonID   string
	AssetID    string
	Generation int32
	TrackID    string
	TaskMarker string

	// DurationSeconds is nil when the probe produced none. D8-9 makes the task timeout derive
	// from it, so an unknown duration is refused BEFORE the job starts rather than running
	// under a guessed ceiling.
	DurationSeconds *float64
}

// ClaimedDraft is what a worker gets when it takes a task: the draft it may write, and nothing
// else. Claiming is what makes a second delivery of the same task visible rather than silently
// successful (D8-6).
type ClaimedDraft struct {
	Claimed bool
	Draft   domain.TranscriptDraft

	// Token is the row's `updated_at` at the moment of the claim, and Complete must present it
	// back. Without it, recovery opens a hole that generation alone cannot close: a worker
	// declared dead, its draft requeued and claimed by a SECOND worker, and then the first one
	// - not dead, only slow - finishes and overwrites the second one's result. Both carry the
	// same (assetID, generation, trackID), so identity alone cannot distinguish them. The token distinguishes
	// two ATTEMPTS at the same generation, which is exactly what P-5 makes possible.
	Token string
}

// RecoveredDraft is a running draft whose P-3 timeout has elapsed, returned by Recover so the
// caller can fail it. Token is the row's updated_at at the moment of recovery — the same value
// Complete expects, so a concurrent requested re-Claim of this track is visible as
// ErrDraftSuperseded rather than silently overwritten.
type RecoveredDraft struct {
	Draft domain.TranscriptDraft
	Token string
}

type DraftRepository interface {
	// ListLessonDrafts returns one entry per current track that has a draft. Absence is a
	// state, not an error: a track simply may not have been recognised yet.
	ListLessonDrafts(ctx context.Context, lessonID string) ([]LessonDraft, error)

	// PendingRecognitions answers the trigger question (P-1). `limit` bounds one sweep so a
	// backlog cannot be pulled into memory whole.
	PendingRecognitions(ctx context.Context, limit int) ([]PendingRecognition, error)

	// Claim moves a track's draft to running in one INSERT SELECT against audio_tracks and
	// video_assets. Only the author's requested Regenerate may replace an existing draft (O7);
	// a duplicate automatic task visibly answers Claimed:false instead of overwriting output.
	//
	// The write MUST verify lesson, asset, generation and ready track membership together. A
	// CHECK cannot read another row, and video_assets.lesson_id is nullable (ON DELETE SET NULL),
	// so neither a CHECK nor a composite foreign key holds this pairing. Without verification
	// a foreign track or a right asset under a wrong generation would be storable (C21).
	// Wrong/foreign/missing tracks return ErrNotFound; a valid track of a non-latest Ready asset
	// returns ErrDraftSuperseded. This prevents delayed jobs from claiming an obsolete video.
	//
	// Language, multilingual and the immutable source leaf are read in that same statement:
	// the recorded snapshot must describe the bytes and mode this attempt will hear, even when
	// another author renames the track concurrently. Leaf is returned through the track join,
	// not duplicated on the draft row.
	Claim(ctx context.Context, lessonID, assetID string, generation int32, trackID string, requested bool, now time.Time) (ClaimedDraft, error)

	// Complete stores the terminal outcome - ready, failed or no_speech.
	//
	// Its refusals are NAMED, because the caller has to tell two opposite things apart and the
	// distinction is not visible in a bare error (D8-7):
	//
	//   - domain.ErrDraftSuperseded - the draft this job claimed is no longer the draft on the
	//     track row: another worker re-claimed it after recovery declared this one dead. The task
	//     is DONE, not failed: retrying would spend tens of minutes per
	//     attempt on a result nobody will read.
	//   - domain.ErrNotFound - the lesson, asset or track was deleted mid-flight, an ordinary authoring
	//     move, not a fault (C08). Also done, also without a retry.
	//   - anything else - transient, and the queue may retry.
	//
	// Without these names the application can only retry everything, which is exactly the three
	// attempts on a deleted lesson that D8-7 exists to refuse, or swallow everything, which
	// loses real failures.
	// `claimToken` is the value ClaimedDraft handed out; a completion that does not present the
	// current one is refused as superseded. The write is keyed by audio_track_id and token.
	Complete(ctx context.Context, draft domain.TranscriptDraft, claimToken string, now time.Time) error

	// Recover returns drafts abandoned in `running` whose per-row P-3 timeout has elapsed, so the
	// caller can fail them with a named cause (P-5). `now` is the reference time; the implementation
	// derives the per-row threshold from video_assets.duration_seconds via the P-3 formula.
	Recover(ctx context.Context, now time.Time, limit int) ([]RecoveredDraft, error)

	// LessonTrack returns a ready track of the latest Ready non-evicted asset, with duration,
	// or nil otherwise. Temporarily, an empty trackID selects its default ready track so the
	// legacy default-only Regenerate can work until MS2c supplies an explicit track ID.
	LessonTrack(ctx context.Context, lessonID, trackID string) (*PendingRecognition, error)
}
