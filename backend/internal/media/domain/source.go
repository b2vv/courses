package domain

const (
	CodecH264 = "h264"
	CodecHEVC = "hevc"
	CodecAAC  = "aac"
	CodecMP3  = "mp3"
)

type SourceProbe struct {
	Width      int
	Height     int
	VideoCodec string
	AudioCodec string
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
