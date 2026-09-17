package domain

import (
	"encoding/hex"
	"net/url"
	"strings"
	"unicode/utf8"
)

// The ceilings live here and nowhere else. The contract mirrors them and a pin holds the
// mirror equal (plan D-2); the read shape depends on MaximumSegments too, because the lesson
// is served in one response so that search cannot lie about a miss (plan D-4).
const (
	MaximumSegments = 5000

	// MaximumTextRunes is counted in RUNES. Every byte budget derived from it must multiply by
	// the worst case UTF-8 width, because a Ukrainian transcript is two bytes per rune and JSON
	// escaping can reach six.
	MaximumTextRunes = 1000

	MaximumSpeakerRunes = 100

	// MaximumSegmentDocuments and MaximumCourseDocuments are G6.
	MaximumSegmentDocuments = 3
	MaximumCourseDocuments  = 200

	MaximumLabelRunes = 200
	MaximumURLRunes   = 2000
)

// Language is a normalised BCP-47 tag. Normalisation lives in the domain precisely so that it
// cannot differ between write paths: the key is (lessonID, language), so `EN` reaching one path
// and `en` the other would split one transcript into two rows.
type Language string

func NewLanguage(raw string) (Language, error) {
	trimmed := strings.TrimSpace(raw)
	if !utf8.ValidString(trimmed) {
		return "", ErrInvalidRequest
	}
	lowered := strings.ToLower(trimmed)
	if !validLanguageTag(lowered) {
		return "", ErrInvalidRequest
	}
	return Language(lowered), nil
}

func validLanguageTag(value string) bool {
	if len(value) < 2 || len(value) > 35 {
		return false
	}
	for index, part := range strings.Split(value, "-") {
		if len(part) < 1 || len(part) > 8 {
			return false
		}
		for _, r := range part {
			if !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') {
				return false
			}
		}
		if index == 0 && (len(part) < 2 || part[0] < 'a' || part[0] > 'z') {
			return false
		}
	}
	return true
}

type DocumentLink struct {
	ID    string
	Label string
	URL   string
}

// NewDocumentLink enforces A3 at the boundary the library is entered through, so no segment can
// reference a link that skipped the rules. Only http(s) is allowed; the media tree is refused
// outright because a transcript must not become a second way to name media bytes (A24).
func NewDocumentLink(input DocumentLink) (DocumentLink, error) {
	if !validID(input.ID) || !validText(input.Label, 1, MaximumLabelRunes) || !validText(input.URL, 1, MaximumURLRunes) {
		return DocumentLink{}, ErrInvalidRequest
	}
	parsed, err := url.Parse(input.URL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return DocumentLink{}, ErrInvalidRequest
	}
	if strings.HasPrefix(parsed.Path, "/media/") {
		return DocumentLink{}, ErrInvalidRequest
	}
	return DocumentLink{ID: input.ID, Label: input.Label, URL: input.URL}, nil
}

type Segment struct {
	ID         string
	StartMs    int64
	EndMs      int64
	Text       string
	Speaker    string
	Confidence *float64
	Documents  []DocumentLink
}

// Active reports whether the segment covers the position. The interval is HALF-OPEN, so a
// position on a boundary belongs to exactly one segment; a zero-length segment is never active,
// which is a state rather than an error because ordered times allow StartMs == EndMs.
func (segment Segment) Active(positionMs int64) bool {
	return positionMs >= segment.StartMs && positionMs < segment.EndMs
}

func NewSegment(input Segment) (Segment, error) {
	if !validID(input.ID) {
		return Segment{}, ErrInvalidRequest
	}
	if input.StartMs < 0 || input.EndMs < input.StartMs {
		return Segment{}, ErrInvalidRequest
	}
	if !validText(input.Text, 1, MaximumTextRunes) {
		if utf8.ValidString(input.Text) && utf8.RuneCountInString(input.Text) > MaximumTextRunes {
			return Segment{}, ErrTextLimit
		}
		return Segment{}, ErrInvalidRequest
	}
	if input.Speaker != "" && !validText(input.Speaker, 1, MaximumSpeakerRunes) {
		return Segment{}, ErrInvalidRequest
	}
	if input.Confidence != nil && (*input.Confidence < 0 || *input.Confidence > 1 || *input.Confidence != *input.Confidence) {
		return Segment{}, ErrInvalidRequest
	}
	if len(input.Documents) > MaximumSegmentDocuments {
		return Segment{}, ErrDocumentLimit
	}

	documents := make([]DocumentLink, 0, len(input.Documents))
	seen := make(map[string]struct{}, len(input.Documents))
	for _, link := range input.Documents {
		// The duplicate is refused before the ceiling so that the same document twice never
		// reaches the primary key and surfaces as a 500 instead of a named code.
		if _, duplicate := seen[link.ID]; duplicate {
			return Segment{}, ErrInvalidRequest
		}
		seen[link.ID] = struct{}{}
		valid, err := NewDocumentLink(link)
		if err != nil {
			return Segment{}, err
		}
		documents = append(documents, valid)
	}
	return Segment{
		ID: input.ID, StartMs: input.StartMs, EndMs: input.EndMs,
		Text: input.Text, Speaker: input.Speaker, Confidence: input.Confidence,
		Documents: documents,
	}, nil
}

type Transcript struct {
	LessonID string
	Language Language
	Segments []Segment

	// AuthoredAgainstGeneration is the generation of the ready asset the author wrote against.
	// Staleness is DERIVED by comparing it with the generation the player would pick, so there
	// is no second write to fall out of step with reality (plan D-3). Zero means the lesson had
	// no ready asset when the text was written, and such a transcript is never stale.
	AuthoredAgainstGeneration int32
}

func NewTranscript(input Transcript) (Transcript, error) {
	if !validID(input.LessonID) || input.Language == "" {
		return Transcript{}, ErrInvalidRequest
	}
	if _, err := NewLanguage(string(input.Language)); err != nil {
		return Transcript{}, err
	}
	if len(input.Segments) > MaximumSegments {
		return Transcript{}, ErrSegmentLimit
	}
	if input.AuthoredAgainstGeneration < 0 {
		return Transcript{}, ErrInvalidRequest
	}

	segments := make([]Segment, 0, len(input.Segments))
	seen := make(map[string]struct{}, len(input.Segments))
	for _, raw := range input.Segments {
		if _, duplicate := seen[raw.ID]; duplicate {
			return Transcript{}, ErrInvalidRequest
		}
		seen[raw.ID] = struct{}{}
		segment, err := NewSegment(raw)
		if err != nil {
			return Transcript{}, err
		}
		segments = append(segments, segment)
	}
	return Transcript{
		LessonID: input.LessonID, Language: input.Language,
		Segments: segments, AuthoredAgainstGeneration: input.AuthoredAgainstGeneration,
	}, nil
}

// Stale answers whether the text was written against a different video than the one the player
// would pick now. Only a READY asset counts, which is why the caller passes the generation of
// the ready asset and zero when there is none: a lesson without ready video, and a lesson whose
// first video is still uploading, must not read as stale.
func (transcript Transcript) Stale(readyGeneration int32) bool {
	if transcript.AuthoredAgainstGeneration == 0 || readyGeneration == 0 {
		return false
	}
	return transcript.AuthoredAgainstGeneration != readyGeneration
}

// ActiveSegment returns the index of the segment covering the position, or -1. Several segments
// may cover one position because WebVTT and SRT both allow overlapping cues, so the tie is
// broken by order — the earliest wins — and the answer is always exactly one segment or none.
func (transcript Transcript) ActiveSegment(positionMs int64) int {
	for index, segment := range transcript.Segments {
		if segment.Active(positionMs) {
			return index
		}
	}
	return -1
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}

func validText(value string, minimum, maximum int) bool {
	if !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
		return false
	}
	count := utf8.RuneCountInString(value)
	return count >= minimum && count <= maximum && strings.TrimSpace(value) != ""
}
