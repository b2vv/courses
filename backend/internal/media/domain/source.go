package domain

const (
	CodecH264 = "h264"
	CodecHEVC = "hevc"
	CodecAAC  = "aac"
	CodecMP3  = "mp3"

	ContainerSniffBytes = 12
)

type SourceProbe struct {
	Width      int
	Height     int
	VideoCodec string
	AudioCodec string
	// DurationSeconds is what the thumbnail's seek position depends on: D-11 seeks to
	// one second, or to zero when the source is shorter than that.
	DurationSeconds float64
}

// UsableDurationSeconds returns no duration when the probe did not produce a
// positive value.
func (probe SourceProbe) UsableDurationSeconds() *float64 {
	if probe.DurationSeconds <= 0 {
		return nil
	}
	duration := probe.DurationSeconds
	return &duration
}

func AcceptContainer(head []byte) error {
	if len(head) < ContainerSniffBytes {
		return ErrUnsupportedSource
	}
	if string(head[:4]) == "\x1a\x45\xdf\xa3" {
		return nil
	}
	if string(head[4:8]) != "ftyp" {
		return ErrUnsupportedSource
	}
	switch string(head[8:12]) {
	case "isom", "iso2", "iso4", "iso5", "iso6", "mp41", "mp42", "avc1", "mmp4", "dash", "M4V ", "qt  ":
		return nil
	default:
		return ErrUnsupportedSource
	}
}

func AcceptSource(probe SourceProbe) error {
	videoAccepted := probe.VideoCodec == CodecH264 || probe.VideoCodec == CodecHEVC
	audioAccepted := probe.AudioCodec == CodecAAC || probe.AudioCodec == CodecMP3
	if probe.Width <= 0 || probe.Height <= 0 || !videoAccepted || !audioAccepted {
		return ErrUnsupportedSource
	}
	return nil
}

type Rendition string

const (
	Rendition360p  Rendition = "360p"
	Rendition720p  Rendition = "720p"
	Rendition1080p Rendition = "1080p"
)

func (rendition Rendition) Height() int {
	switch rendition {
	case Rendition360p:
		return 360
	case Rendition720p:
		return 720
	case Rendition1080p:
		return 1080
	default:
		return 0
	}
}

func SelectLadder(probe SourceProbe) []Rendition {
	ladder := []Rendition{Rendition360p}
	if probe.Height >= 720 {
		ladder = append(ladder, Rendition720p)
	}
	if probe.Height >= 1080 {
		ladder = append(ladder, Rendition1080p)
	}
	return ladder
}
