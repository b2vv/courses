package domain

import (
	"strings"
	"testing"
)

func id(last string) string { return "11111111-1111-4111-8111-1111111111" + last }

func segment(startMs, endMs int64, last string) Segment {
	return Segment{ID: id(last), StartMs: startMs, EndMs: endMs, Text: "текст"}
}

func TestLanguageNormalisationCannotDependOnTheWritePath(t *testing.T) {
	// The key is (lessonID, language). If `EN` and `en` normalised differently on the editor and
	// the import path, one transcript would become two rows and the reader would get a lottery.
	for _, raw := range []string{"en", "EN", " En ", "eN"} {
		got, err := NewLanguage(raw)
		if err != nil || got != "en" {
			t.Fatalf("NewLanguage(%q) = %q, %v; want \"en\", nil", raw, got, err)
		}
	}
	if got, err := NewLanguage("uk-UA"); err != nil || got != "uk-ua" {
		t.Fatalf("NewLanguage(\"uk-UA\") = %q, %v; want \"uk-ua\", nil", got, err)
	}
	for _, raw := range []string{"", "e", "english-", "-en", "en_US", "en US", strings.Repeat("a", 36)} {
		if _, err := NewLanguage(raw); err != ErrInvalidRequest {
			t.Fatalf("NewLanguage(%q) err = %v; want invalid_request", raw, err)
		}
	}
}

func TestDocumentLinkRefusesEverythingThatIsNotHTTP(t *testing.T) {
	// A3 is enforced where the library is entered, so a segment cannot hold a link that skipped
	// it. The /media/ case is A24: a transcript must not become a second way to name media bytes.
	for _, raw := range []string{
		"javascript:alert(1)", "data:text/html,<script>", "file:///etc/passwd",
		"http://host/media/p/abc/hls/x.m3u8", "https://host/media/p/abc/hls/x.m3u8",
		"http://", "not a url at all", "",
	} {
		if _, err := NewDocumentLink(DocumentLink{ID: id("01"), Label: "doc", URL: raw}); err == nil {
			t.Fatalf("NewDocumentLink accepted %q", raw)
		}
	}
	if _, err := NewDocumentLink(DocumentLink{ID: id("01"), Label: "doc", URL: "https://example.test/a"}); err != nil {
		t.Fatalf("NewDocumentLink refused a valid https link: %v", err)
	}
}

func TestSegmentBoundsAreNamedNotGeneric(t *testing.T) {
	over := strings.Repeat("я", MaximumTextRunes+1)
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: over}); err != ErrTextLimit {
		t.Fatalf("text over the ceiling err = %v; want transcript_text_limit", err)
	}
	exact := strings.Repeat("я", MaximumTextRunes)
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: exact}); err != nil {
		t.Fatalf("text exactly at the ceiling was refused: %v", err)
	}

	links := make([]string, 0, MaximumSegmentDocuments+1)
	for index := 0; index <= MaximumSegmentDocuments; index++ {
		links = append(links, id("0"+string(rune('1'+index))))
	}
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: "t", DocumentIDs: links}); err != ErrDocumentLimit {
		t.Fatalf("documents over the ceiling err = %v; want transcript_document_limit", err)
	}

	// The duplicate must lose before the ceiling, or the primary key answers with a 500 where a
	// named code was promised.
	same := id("01")
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: "t", DocumentIDs: []string{same, same}}); err != ErrInvalidRequest {
		t.Fatalf("duplicate document err = %v; want invalid_request", err)
	}
}

func TestSegmentTimesAndConfidence(t *testing.T) {
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: -1, EndMs: 5, Text: "t"}); err != ErrInvalidRequest {
		t.Fatalf("negative start was accepted")
	}
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: 5, EndMs: 4, Text: "t"}); err != ErrInvalidRequest {
		t.Fatalf("end before start was accepted")
	}
	// Zero length is allowed by A2 and is a state, not an error: such a segment is simply never
	// active. TestActiveSegmentIsExactlyOneOrNone pins the other half of that statement.
	if _, err := NewSegment(Segment{ID: id("01"), StartMs: 5, EndMs: 5, Text: "t"}); err != nil {
		t.Fatalf("zero-length segment was refused: %v", err)
	}
	for _, value := range []float64{-0.01, 1.01} {
		confidence := value
		if _, err := NewSegment(Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: "t", Confidence: &confidence}); err != ErrInvalidRequest {
			t.Fatalf("confidence %v was accepted", value)
		}
	}
	for _, value := range []float64{0, 0.5, 1} {
		confidence := value
		if _, err := NewSegment(Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: "t", Confidence: &confidence}); err != nil {
			t.Fatalf("confidence %v was refused: %v", value, err)
		}
	}
}

func TestActiveSegmentIsExactlyOneOrNone(t *testing.T) {
	// Overlap is not a defect: WebVTT and SRT both allow two speakers at once, so a valid file
	// produces segments whose intervals contain the same position. The interval is half-open and
	// the earliest wins, so the answer is always one segment or none — never two highlighted.
	transcript, err := NewTranscript(Transcript{
		LessonID: id("aa"), Language: "uk",
		Segments: []Segment{segment(0, 1000, "01"), segment(500, 1500, "02"), segment(2000, 2000, "03")},
	})
	if err != nil {
		t.Fatalf("NewTranscript: %v", err)
	}
	for _, test := range []struct {
		position int64
		want     int
	}{
		{position: 0, want: 0},
		{position: 999, want: 0},
		{position: 500, want: 0},  // overlap: the earlier segment wins
		{position: 1000, want: 1}, // half-open: the boundary belongs to the next one only
		{position: 1499, want: 1},
		{position: 1500, want: -1}, // the gap highlights nothing
		{position: 2000, want: -1}, // the zero-length segment is never active
	} {
		if got := transcript.ActiveSegment(test.position); got != test.want {
			t.Fatalf("ActiveSegment(%d) = %d; want %d", test.position, got, test.want)
		}
	}
}

func TestSegmentCeilingIsNamedAndOwnedHere(t *testing.T) {
	segments := make([]Segment, 0, MaximumSegments+1)
	for index := 0; index <= MaximumSegments; index++ {
		segments = append(segments, Segment{ID: id("01"), StartMs: 0, EndMs: 1, Text: "t"})
	}
	if _, err := NewTranscript(Transcript{LessonID: id("aa"), Language: "uk", Segments: segments}); err != ErrSegmentLimit {
		t.Fatalf("over the segment ceiling err = %v; want transcript_segment_limit", err)
	}
}

func TestStalenessIsDerivedAndNeverFiresWithoutAReadyAsset(t *testing.T) {
	// The whole point of deriving staleness instead of storing it: a flag cannot drift from
	// reality when it is computed from reality. These rows are the ones a stored flag got wrong
	// in the first edition of the plan.
	for _, test := range []struct {
		name      string
		authored  int32
		readyNow  int32
		wantStale bool
	}{
		{name: "written before any video, nothing ready yet", authored: 0, readyNow: 0, wantStale: false},
		{name: "written before any video, first video published", authored: 0, readyNow: 1, wantStale: false},
		{name: "written against the video that is still current", authored: 2, readyNow: 2, wantStale: false},
		{name: "video replaced", authored: 1, readyNow: 2, wantStale: true},
		{name: "transcode failed, nothing ready", authored: 1, readyNow: 0, wantStale: false},
	} {
		transcript := Transcript{AuthoredAgainstGeneration: test.authored}
		if got := transcript.Stale(test.readyNow); got != test.wantStale {
			t.Fatalf("%s: Stale(%d) with authored=%d = %v; want %v", test.name, test.readyNow, test.authored, got, test.wantStale)
		}
	}
}

func TestTranscriptRefusesDuplicateSegmentIDs(t *testing.T) {
	_, err := NewTranscript(Transcript{
		LessonID: id("aa"), Language: "uk",
		Segments: []Segment{segment(0, 1, "01"), segment(1, 2, "01")},
	})
	if err != ErrInvalidRequest {
		t.Fatalf("duplicate segment id err = %v; want invalid_request", err)
	}
}

func TestWritingIsNotTheSameRuleAsReading(t *testing.T) {
	// A published course is readable by anyone signed in. If writing shared that rule, every
	// student could rewrite the lecture, so this pair is pinned as a pair.
	published := Course{OwnerID: id("aa"), Status: Published}
	student := Actor{UserID: id("bb")}
	owner := Actor{UserID: id("aa")}
	admin := Actor{UserID: id("cc"), Administrator: true}

	if !published.CanBeViewedBy(student) {
		t.Fatal("a signed-in student cannot read a published course")
	}
	if published.CanBeEditedBy(student) {
		t.Fatal("a student can write the transcript of a published course")
	}
	if !published.CanBeEditedBy(owner) || !published.CanBeEditedBy(admin) {
		t.Fatal("the owner or an admin cannot write")
	}
	// An empty actor must not read as the owner of an ownerless course on the WRITE path. The
	// read path records that it does and relies on the schema; writing cannot afford that.
	if (Course{}).CanBeEditedBy(Actor{}) {
		t.Fatal("an anonymous actor writes an ownerless course")
	}
}
