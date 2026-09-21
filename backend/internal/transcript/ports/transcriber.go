package ports

import (
	"context"

	"courses/internal/transcript/domain"
)

// Recognition is what the engine reports back for one audio file.
type Recognition struct {
	// Language is what the model detected. It is the reason `auto` never needs to reach the
	// database: the select says "detect it", and this field is the answer (D8-8).
	Language domain.Language

	Segments []domain.Segment

	// SpeechSeconds is the audio left after voice-activity detection, and it is the signal that
	// decides `no_speech`. The spec first named `no_speech_prob` per segment; measurement
	// (§4.5, A01) showed that feature is degenerate on this model - silence, applause and
	// Ukrainian speech all land in one narrow band, and silence is sometimes HIGHER than
	// speech - while the model still invents confident text over non-speech. VAD separates
	// them completely, so the port carries the measurement rather than the model's own opinion.
	SpeechSeconds float64
}

// Transcriber is the seam PRODUCT-BRIEF.md:551 reserved. It has a real external dependency -
// the recognition engine, which is a separate process with its own weights - so it earns a port
// rather than being documented as a future one (AGENTS.md).
//
// The implementation runs a subprocess and reads structured output from its stdout (plan P-4).
// That shape is deliberate: the engine is Python and the worker is Go, and a subprocess boundary
// keeps the model's runtime out of the Go process entirely, exactly as ffmpeg is kept out of the
// media worker's.
type Transcriber interface {
	// Transcribe recognises the audio at `audioPath`. Cancellation through ctx must actually
	// stop the subprocess: a task that outlives its deadline holds a slot for tens of minutes
	// (D8-9), and a cancelled context that only abandons the pipe would leak the engine.
	//
	// `languageHint` is empty when the author chose `auto`. A non-empty hint is the author
	// overriding detection, not a suggestion.
	Transcribe(ctx context.Context, audioPath string, languageHint domain.Language) (Recognition, error)
}

// AudioSource extracts the audio the engine needs from the video the media module owns.
//
// It is a port rather than a direct call because the storage layout belongs to `media`: the key
// shape and the resolver are private to that module, and the precedent 7a set covers reading a
// NUMBER out of a neighbouring table, not knowing where that neighbour keeps its bytes (spec
// §9.2, corrected by D8-3).
type AudioSource interface {
	// Extract writes 16 kHz mono audio for one asset and returns the path to it, together with
	// a release that removes the file. An hour of speech is roughly 115 MB, so the caller must
	// be able to drop it promptly rather than wait for the task to end.
	//
	// It reports ErrNotFound when the asset's bytes are gone - which is an ordinary outcome,
	// not a fault: eviction removes the source of an older generation while recognition of it
	// may still be running.
	Extract(ctx context.Context, assetID string) (audioPath string, release func(), err error)
}
