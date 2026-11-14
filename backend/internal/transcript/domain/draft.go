package domain

import (
	"slices"
	"strings"
)

// A draft is what the machine produces and the author corrects. It is deliberately NOT a
// transcript: the transcript is what a student reads, and D8-10 keeps publication on the
// existing write path, so nothing here needs a promotion step of its own.
//
// The states are the worker's, not the product's. `no_speech` is one of them rather than a
// `ready` draft with no lines, because C13 asks that silence produce a NAMED result - an empty
// `ready` draft is indistinguishable from work nobody did.
type DraftState string

const (
	DraftQueued   DraftState = "queued"
	DraftRunning  DraftState = "running"
	DraftReady    DraftState = "ready"
	DraftFailed   DraftState = "failed"
	DraftNoSpeech DraftState = "no_speech"
)

// DraftStates is the ordered set used by validation and the HTTP contract.
var DraftStates = []DraftState{DraftQueued, DraftRunning, DraftReady, DraftFailed, DraftNoSpeech}

// MaximumFailureRunes bounds the named cause carried back to the author. It is short on
// purpose: the field holds a reason, not a stack trace, and the migration mirrors the number.
const MaximumFailureRunes = 200

func validDraftState(state DraftState) bool {
	return slices.Contains(DraftStates, state)
}

// Terminal answers whether the worker is done with this draft. Recovery (plan P-5) reads it to
// tell a draft abandoned mid-flight from one that simply finished.
func (state DraftState) Terminal() bool {
	return state == DraftReady || state == DraftFailed || state == DraftNoSpeech
}

type TranscriptDraft struct {
	LessonID string

	// AssetID and Generation name the video this text came from. Staleness is DERIVED from the
	// generation, exactly as the transcript derives it (D8-4), so there is no stored flag to
	// drift. Unlike the transcript's zero, a draft ALWAYS has an asset: nothing makes one
	// except a ready asset.
	AssetID    string
	Generation int32

	State DraftState

	// Language is absent until recognition reports one, so an unfinished or speechless draft
	// carries none.
	Language Language

	Segments []Segment

	// Failure is the named cause, and only a failed draft has one.
	Failure string

	// SourceLeaf names the audio the draft was recognised from: the default track's leaf at claim
	// time (multi-audio S6, D-S6-1). Empty is a draft made before S6, which always heard
	// "original" (HeardLeaf).
	SourceLeaf string
}

// OriginalAudioLeaf is the leaf of a video's own sound: the embedded track and every pre-S6 draft.
const OriginalAudioLeaf = "original"

// ValidAudioLeaf mirrors the leaf shape of `audio_tracks.source_leaf` (00025) and the
// `asr_drafts.source_leaf` CHECK (00029): "original" or "audio-<lowercase uuid>". The extraction
// adapter joins it into a path, so nothing else may pass.
func ValidAudioLeaf(leaf string) bool {
	if leaf == OriginalAudioLeaf {
		return true
	}
	id, found := strings.CutPrefix(leaf, "audio-")
	return found && validID(id)
}

// HeardLeaf is the leaf the draft's text came from; a pre-S6 draft heard the original.
func (draft TranscriptDraft) HeardLeaf() string {
	if draft.SourceLeaf == "" {
		return OriginalAudioLeaf
	}
	return draft.SourceLeaf
}

// NewTranscriptDraft validates with the SAME constructors the transcript uses (D8-12). That is the point
// of the decision: a looped model output has to be refused here, at the boundary of the module,
// rather than an hour later when the author presses save and discovers the text will not go in.
func NewTranscriptDraft(input TranscriptDraft) (TranscriptDraft, error) {
	if !validID(input.LessonID) || !validID(input.AssetID) {
		return TranscriptDraft{}, ErrInvalidRequest
	}
	if input.Generation < 1 {
		return TranscriptDraft{}, ErrInvalidRequest
	}
	if !validDraftState(input.State) {
		return TranscriptDraft{}, ErrInvalidRequest
	}
	// validText is the transcript's own check, reused rather than re-derived: besides the rune
	// ceiling it refuses a NUL, which is valid UTF-8 and which PostgreSQL `text` cannot store.
	// Without it a cause carrying one would fail at the INSERT, and `Complete` would report that
	// as transient - three pointless retries of a job that can never succeed.
	if input.Failure != "" && !validText(input.Failure, 1, MaximumFailureRunes) {
		return TranscriptDraft{}, ErrInvalidRequest
	}

	// `failed` and a cause imply each other. Either alone is a draft that cannot be explained
	// to the author, or a cause nobody will ever show.
	if (input.State == DraftFailed) != (input.Failure != "") {
		return TranscriptDraft{}, ErrInvalidRequest
	}

	// The NORMALISED tag is what gets stored, and what the `auto` guard is applied to. Keeping
	// the raw value would have let `AUTO` past the guard - it is not equal to `auto` - and would
	// have stored ` EN `, which the column's own CHECK (`language = lower(language)`) then
	// refuses at the INSERT. The transcript makes normalisation a domain property for the same
	// reason: so it cannot differ between write paths.
	language := input.Language
	if language != "" {
		normalised, err := NewLanguage(string(language))
		if err != nil {
			return TranscriptDraft{}, err
		}
		// `auto` is the select's word for "detect it" and it satisfies the BCP-47 shape, so the
		// constructor alone lets it through - and the read orders by language, where it sorts
		// ahead of `en` and `uk`. D8-8 keeps it a UI state; the migration carries the same
		// refusal.
		if normalised == "auto" {
			return TranscriptDraft{}, ErrInvalidRequest
		}
		language = normalised
	}

	if len(input.Segments) > MaximumSegments {
		return TranscriptDraft{}, ErrSegmentLimit
	}
	segments := make([]Segment, 0, len(input.Segments))
	seen := make(map[string]struct{}, len(input.Segments))
	for _, raw := range input.Segments {
		if _, duplicate := seen[raw.ID]; duplicate {
			return TranscriptDraft{}, ErrInvalidRequest
		}
		seen[raw.ID] = struct{}{}
		segment, err := NewSegment(raw)
		if err != nil {
			return TranscriptDraft{}, err
		}
		segments = append(segments, segment)
	}

	// A ready draft has text and a language. Without this, recognition that produced nothing
	// would land as `ready` with an empty list, which is exactly the shape C13 refuses.
	if input.State == DraftReady && (len(segments) == 0 || language == "") {
		return TranscriptDraft{}, ErrInvalidRequest
	}

	return TranscriptDraft{
		LessonID:   input.LessonID,
		AssetID:    input.AssetID,
		Generation: input.Generation,
		State:      input.State,
		Language:   language,
		Segments:   segments,
		Failure:    input.Failure,
		SourceLeaf: input.SourceLeaf,
	}, nil
}

// Stale answers whether the draft was made from a different video than the one the player would
// pick now, or from a different audio track than its current default (multi-audio S6, D-S6-6).
// The generation rule is the transcript's (D8-4): only a READY asset counts, and a lesson with
// none must not read as stale. An empty defaultLeaf (no default known) never makes it stale.
func (draft TranscriptDraft) Stale(readyGeneration int32, defaultLeaf string) bool {
	if readyGeneration == 0 {
		return false
	}
	if draft.Generation != readyGeneration {
		return true
	}
	return defaultLeaf != "" && draft.HeardLeaf() != defaultLeaf
}
