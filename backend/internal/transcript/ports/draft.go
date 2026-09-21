package ports

import (
	"context"
	"time"

	"courses/internal/transcript/domain"
)

// LessonDraft is one read of the machine-made draft. It carries the same two pieces the
// transcript read carries and for the same reasons: the lock token, without which the author
// cannot close the editing loop, and the generation the player would pick, because staleness is
// derived rather than stored (D8-4).
type LessonDraft struct {
	Found bool
	Draft domain.TranscriptDraft

	// UpdatedAt is the optimistic-lock token. The draft has two writers - the author and
	// recognition - so it needs the same guard the transcript has, and for a sharper reason:
	// regeneration moves the token while the author is typing into it.
	UpdatedAt string

	ReadyGeneration int32
}

// PendingRecognition is one lesson that has a ready asset and no draft for that asset's
// generation. It is the whole of the trigger (plan P-1): there is no event and no outbox in
// this repository, so "start recognition" is a QUESTION ABOUT STATE, asked repeatedly. A crash
// between publishing an asset and enqueueing a task therefore loses nothing - the next sweep
// asks again.
type PendingRecognition struct {
	LessonID   string
	AssetID    string
	Generation int32

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
	// same (assetID, generation), so ErrDraftSuperseded never fires. The token distinguishes
	// two ATTEMPTS at the same generation, which is exactly what P-5 makes possible.
	Token string
}

type DraftRepository interface {
	// GetLessonDraft returns the draft of one lesson. Absence is a state, not an error: a
	// lesson simply may not have been recognised yet.
	GetLessonDraft(ctx context.Context, lessonID string) (LessonDraft, error)

	// PendingRecognitions answers the trigger question (P-1). `limit` bounds one sweep so a
	// backlog cannot be pulled into memory whole.
	PendingRecognitions(ctx context.Context, limit int) ([]PendingRecognition, error)

	// Claim moves a draft to `running` for exactly one caller. A second task for the same
	// (assetID, generation) finds it already claimed and answers so - the conflict is VISIBLE,
	// unlike the media queue's TaskID collision, which returns a silent success and would make
	// the author's "regenerate" button do nothing (D8-6).
	//
	// The implementation MUST verify the whole triple - that the asset belongs to the lesson AND
	// that the generation is that asset's own - in one statement against `video_assets`. Neither
	// a foreign key nor a CHECK can hold it here: a CHECK cannot see other rows, and a composite
	// key into `video_assets (lesson_id, id)` would rest on a nullable column that
	// `ON DELETE SET NULL` empties. So the write path is the only place the pairing can live,
	// and a draft naming another lesson's asset, or the right asset under the wrong generation,
	// is otherwise perfectly storable (C21).
	//
	// A wrong generation is not cosmetic: `Stale` compares it against the generation the player
	// would pick, so a draft claiming one it does not have reports the wrong answer, and the
	// idempotency key (assetID, generation) stops naming the work it is supposed to name.
	Claim(ctx context.Context, lessonID, assetID string, generation int32, now time.Time) (ClaimedDraft, error)

	// SaveDraft stores the author's corrections and returns the new token.
	//
	// A8b is why this exists: the owner decided the draft is saved on the server so work can
	// continue later, from another device, after a closed browser. Without it, `UpdatedAt` above
	// would be a token that nothing consumes - a lock announced and never taken - and the author's
	// only way to keep text would be to publish it as a transcript, which is the OTHER save.
	//
	// It refuses with domain.ErrTranscriptModified when the token is stale, which is how the
	// author learns that regeneration replaced the text under them rather than silently losing
	// the edit (D8-1).
	SaveDraft(ctx context.Context, draft domain.TranscriptDraft, expectedUpdatedAt string, now time.Time) (string, error)

	// Complete stores the terminal outcome - ready, failed or no_speech.
	//
	// Its refusals are NAMED, because the caller has to tell two opposite things apart and the
	// distinction is not visible in a bare error (D8-7):
	//
	//   - domain.ErrDraftSuperseded - the draft this job claimed is no longer the draft on the
	//     row. Two different causes, one answer: the video was replaced while recognition ran
	//     (C09), or another worker re-claimed the same generation after recovery declared this
	//     one dead. The task is DONE, not failed: retrying would spend tens of minutes per
	//     attempt on a result nobody will read.
	//   - domain.ErrNotFound - the lesson was deleted mid-flight, which is an ordinary authoring
	//     move, not a fault (C08). Also done, also without a retry.
	//   - anything else - transient, and the queue may retry.
	//
	// Without these names the application can only retry everything, which is exactly the three
	// attempts on a deleted lesson that D8-7 exists to refuse, or swallow everything, which
	// loses real failures.
	// `claimToken` is the value ClaimedDraft handed out; a completion that does not present the
	// current one is refused as superseded.
	Complete(ctx context.Context, draft domain.TranscriptDraft, claimToken string, now time.Time) error

	// Recover returns drafts abandoned in `running` by a worker that died, so the caller can
	// requeue them or fail them with a named cause (P-5). It mirrors media's recovery, which is
	// state in the database rather than a signal the process must survive to receive: a deploy
	// mid-task is the ordinary case, not the exception, when a task runs for tens of minutes.
	Recover(ctx context.Context, olderThan time.Time, limit int) ([]domain.TranscriptDraft, error)
}
