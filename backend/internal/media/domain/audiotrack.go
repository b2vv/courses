package domain

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaximumLanguageLength = 35
	MaximumLabelLength    = 64

	// AudioLeafPrefix names the per-track source leaf under source/<assetID>/.
	// An uploaded track stores its bytes at audio-<trackID>; an embedded track
	// has no separate source and reuses the video source leaf "original".
	AudioLeafPrefix   = "audio-"
	EmbeddedAudioLeaf = "original"
)

// AudioTrackKind separates the two lifecycles of an audio track (D3). An
// embedded track is a Ready-only row materialised from the video source's own
// audio stream when the video transcode completes; it never owns upload states.
// An uploaded track runs the full ticket→Ready/Failed arc of a VideoAsset,
// reusing the same State machinery, keyed on its own source leaf.
type AudioTrackKind string

const (
	EmbeddedTrack AudioTrackKind = "embedded"
	UploadedTrack AudioTrackKind = "uploaded"
)

// AudioTrackKinds is the ordered set used by validation and the HTTP contract.
var AudioTrackKinds = []AudioTrackKind{EmbeddedTrack, UploadedTrack}

// AudioTrack is per-track metadata; the upload, queue and transcode machinery
// is reused from VideoAsset rather than cloned (D3). Exactly one track of a
// Ready asset is the default, and a Ready asset always carries at least one
// Ready track (D4) — both invariants are pure functions below, enforced by the
// persistence layer inside a transaction.
type AudioTrack struct {
	ID           string
	VideoAssetID string
	Kind         AudioTrackKind
	Language     string
	Label        string
	IsDefault    bool
	State        State
	SourceLeaf   string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// NewUploadedTrack validates a freshly minted uploaded audio track. It begins in
// Created, is never default before it is Ready, and its bytes live at the
// audio-<trackID> leaf of the asset's source directory.
func NewUploadedTrack(track AudioTrack) (AudioTrack, error) {
	if err := validateLanguageLabel(track.Language, track.Label); err != nil {
		return AudioTrack{}, err
	}
	if !validID(track.ID) || !validID(track.VideoAssetID) || track.Kind != UploadedTrack ||
		track.State != Created || track.IsDefault || track.SourceLeaf != AudioLeafPrefix+track.ID ||
		track.CreatedAt.IsZero() || !track.UpdatedAt.Equal(track.CreatedAt) {
		return AudioTrack{}, ErrInvalidRequest
	}
	return track, nil
}

// NewEmbeddedTrack validates the derived Ready track materialised from a video
// source's own audio stream (D3). It has no upload arc: it is Ready from the
// start and points at the shared "original" source leaf. Unlike an uploaded
// track IsDefault is intentionally unconstrained here: an embedded track is the
// audio stream that takes the asset straight to Ready, so it is the first (and,
// until an uploaded track lands, only) default.
func NewEmbeddedTrack(track AudioTrack) (AudioTrack, error) {
	if err := validateLanguageLabel(track.Language, track.Label); err != nil {
		return AudioTrack{}, err
	}
	if !validID(track.ID) || !validID(track.VideoAssetID) || track.Kind != EmbeddedTrack ||
		track.State != Ready || track.SourceLeaf != EmbeddedAudioLeaf ||
		track.CreatedAt.IsZero() || !track.UpdatedAt.Equal(track.CreatedAt) {
		return AudioTrack{}, ErrInvalidRequest
	}
	return track, nil
}

// ValidateAudioTrackState reports whether a state is reachable for a kind. An
// embedded track is only ever Ready; an uploaded track walks the full arc but
// never enters AwaitingAudio, which is an asset-level waiting state.
func ValidateAudioTrackState(kind AudioTrackKind, state State) error {
	switch kind {
	case EmbeddedTrack:
		if state == Ready {
			return nil
		}
	case UploadedTrack:
		if state != AwaitingAudio && slices.Contains(States, state) {
			return nil
		}
	}
	return ErrInvalidStateTransition
}

// AcceptAudioTrackSource accepts an uploaded audio-track source file. Unlike the
// video source it carries no video stream, and — the point that distinguishes it
// from the video-only relaxation (D4) — an absent or unsupported audio codec is
// a business error: a file with no audio stream, or a non-audio payload behind an
// accepted container, never becomes a track.
func AcceptAudioTrackSource(probe SourceProbe) error {
	if probe.AudioCodec == CodecAAC || probe.AudioCodec == CodecMP3 {
		return nil
	}
	return ErrUnsupportedSource
}

// AssetMayBeReady reports whether the asset can hold Ready: the invariant is at
// least one Ready track (D4), so silent playback can never be published.
func AssetMayBeReady(tracks []AudioTrack) bool {
	for _, track := range tracks {
		if track.State == Ready {
			return true
		}
	}
	return false
}

// ElectDefault returns the ID of the track that should carry the default flag:
// the first Ready track in deterministic (CreatedAt, ID) order, or "" when no
// track is Ready yet.
func ElectDefault(tracks []AudioTrack) string {
	ready := readyTracksSorted(tracks)
	if len(ready) == 0 {
		return ""
	}
	return ready[0].ID
}

// ValidateDefaultInvariant checks the default flag against the asset state (D4):
// before the asset is Ready at most one track may be default; once it is Ready
// exactly one track is default and that track is itself Ready.
func ValidateDefaultInvariant(tracks []AudioTrack, assetState State) error {
	defaults := 0
	readyDefaults := 0
	for _, track := range tracks {
		if track.IsDefault {
			defaults++
			if track.State == Ready {
				readyDefaults++
			}
		}
	}
	if assetState == Ready {
		if defaults == 1 && readyDefaults == 1 {
			return nil
		}
		return ErrInvalidRequest
	}
	if defaults <= 1 {
		return nil
	}
	return ErrInvalidRequest
}

// ApplyDefault returns the track set with trackID as the sole default, moving the
// flag off every other track in one step (the atomic swap of D4). The target
// must exist and be Ready.
func ApplyDefault(tracks []AudioTrack, trackID string) ([]AudioTrack, error) {
	index := -1
	for i, track := range tracks {
		if track.ID == trackID {
			index = i
			break
		}
	}
	if index < 0 || tracks[index].State != Ready {
		return nil, ErrInvalidRequest
	}
	swapped := make([]AudioTrack, len(tracks))
	copy(swapped, tracks)
	for i := range swapped {
		swapped[i].IsDefault = i == index
	}
	return swapped, nil
}

// ValidateTrackDeletion refuses a deletion that would leave a Ready asset without
// a Ready track (D4). The track must exist; deleting it is allowed only when the
// asset is not Ready, or when at least one other Ready track would remain.
func ValidateTrackDeletion(tracks []AudioTrack, trackID string, assetState State) error {
	found := false
	readyOthers := 0
	for _, track := range tracks {
		if track.ID == trackID {
			found = true
			continue
		}
		if track.State == Ready {
			readyOthers++
		}
	}
	if !found {
		return ErrVideoNotFound
	}
	if assetState == Ready && readyOthers == 0 {
		return ErrInvalidRequest
	}
	return nil
}

func readyTracksSorted(tracks []AudioTrack) []AudioTrack {
	ready := make([]AudioTrack, 0, len(tracks))
	for _, track := range tracks {
		if track.State == Ready {
			ready = append(ready, track)
		}
	}
	slices.SortFunc(ready, func(a, b AudioTrack) int {
		if !a.CreatedAt.Equal(b.CreatedAt) {
			return a.CreatedAt.Compare(b.CreatedAt)
		}
		return strings.Compare(a.ID, b.ID)
	})
	return ready
}

func validateLanguageLabel(language, label string) error {
	if !validLanguage(language) || !validLabel(label) {
		return ErrInvalidRequest
	}
	return nil
}

func validLanguage(value string) bool {
	count := utf8.RuneCountInString(value)
	if count < 1 || count > MaximumLanguageLength {
		return false
	}
	for _, r := range value {
		letter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
		digit := r >= '0' && r <= '9'
		if !letter && !digit && r != '-' {
			return false
		}
	}
	return true
}

func validLabel(value string) bool {
	count := utf8.RuneCountInString(value)
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && count >= 1 && count <= MaximumLabelLength
}
