package ports

import (
	"context"

	"courses/internal/media/domain"
)

// AudioRendition is one Ready audio track as the master manifest advertises it
// (D7). The adapter derives the playlist URI from Embedded/TrackID so the HLS
// storage layout stays inside the adapter: an embedded track reuses the video
// render's "audio/" variant, an uploaded track its own "audio-<trackID>/".
type AudioRendition struct {
	TrackID  string
	Name     string
	Language string
	Default  bool
	Embedded bool
}

type Transcoder interface {
	Probe(context.Context, domain.StorageKey) (domain.SourceProbe, error)
	// Render transcodes the video source. hasAudio selects the audio-carrying
	// render (embedded "audio/" variant) or the video-only render (no audio map,
	// no audio group — D4).
	Render(context.Context, domain.StorageKey, domain.StorageKey, []domain.Rendition, bool) error
	// RenderAudioTrack packages one uploaded audio-track source file into a single
	// fmp4 HLS audio variant written flat into the work key.
	RenderAudioTrack(context.Context, domain.StorageKey, domain.StorageKey) error
	Thumbnail(context.Context, domain.StorageKey, domain.StorageKey, domain.SourceProbe) error
	// RegenerateMaster rewrites the published master.m3u8 of the asset from the
	// given Ready audio set, atomically (temp+rename), preserving the ffmpeg video
	// variant lines verbatim (D7). Callers serialise it per asset.
	RegenerateMaster(context.Context, string, []AudioRendition) error
}

// MasterWriter is the one Transcoder capability the API's audio-track authoring
// needs: rewriting an asset's master from its Ready audio set (D7). The API never
// probes or renders, so it depends on this narrow port rather than Transcoder.
type MasterWriter interface {
	RegenerateMaster(context.Context, string, []AudioRendition) error
}
