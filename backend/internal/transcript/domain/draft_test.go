package domain

import (
	"fmt"
	"strings"
	"testing"
)

func draft(mutate func(*TranscriptDraft)) TranscriptDraft {
	base := TranscriptDraft{
		LessonID:   id("01"),
		AssetID:    id("02"),
		Generation: 1,
		State:      DraftReady,
		Language:   "uk",
		Segments:   []Segment{segment(0, 1000, "03")},
	}
	if mutate != nil {
		mutate(&base)
	}
	return base
}

func TestDraftValidatesWithTheSameConstructorsAsTheTranscript(t *testing.T) {
	// D8-12 is the point of this test: a looped model output has to be refused at the boundary
	// of the module, not an hour later when the author presses save. If the draft grew its own
	// rules, the text would pass here and fail there - and the author would have spent the hour.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.Segments = []Segment{{ID: id("03"), StartMs: 0, EndMs: 1000, Text: strings.Repeat("я", MaximumTextRunes+1)}}
	})); err != ErrTextLimit {
		t.Fatalf("a line over the rune ceiling must fail with the transcript's own error, got %v", err)
	}

	over := make([]Segment, MaximumSegments+1)
	for index := range over {
		over[index] = segment(int64(index), int64(index)+1, "03")
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Segments = over })); err != ErrSegmentLimit {
		t.Fatalf("segment ceiling must be the transcript's, got %v", err)
	}

	// End before start is the transcript's rule too, and a model that emits a backwards span is
	// exactly the kind of output this boundary exists to stop.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.Segments = []Segment{{ID: id("03"), StartMs: 900, EndMs: 100, Text: "текст"}}
	})); err == nil {
		t.Fatal("a backwards span must not enter the module")
	}

	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.Segments = []Segment{segment(0, 10, "03"), segment(10, 20, "03")}
	})); err != ErrInvalidRequest {
		t.Fatal("a repeated segment id must be refused, as it is for a transcript")
	}
}

func TestAutoNeverBecomesALanguage(t *testing.T) {
	// `auto` satisfies the BCP-47 shape, so NewLanguage alone lets it through - and the read
	// orders by language, where "auto" sorts ahead of "en" and "uk". D8-8 keeps it a UI state;
	// this is the backstop, and the migration carries the same refusal.
	if _, err := NewLanguage("auto"); err != nil {
		t.Fatal("precondition: NewLanguage accepts \"auto\", which is why this backstop exists")
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Language = "auto" })); err != ErrInvalidRequest {
		t.Fatalf("auto must not reach the language field, got %v", err)
	}
	// The hint travelling as "absent" is the legitimate shape, and only for a draft that has
	// nothing to report yet.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments = DraftQueued, "", nil
	})); err != nil {
		t.Fatalf("a queued draft has no language yet, got %v", err)
	}
}

func TestSilenceCannotLookLikeFinishedWork(t *testing.T) {
	// C13: an empty `ready` draft is indistinguishable from work nobody did, which is why
	// `no_speech` is a state rather than an outcome the reader has to infer.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Segments = nil })); err != ErrInvalidRequest {
		t.Fatal("ready with no lines must be refused")
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Language = "" })); err != ErrInvalidRequest {
		t.Fatal("ready with no language must be refused")
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments = DraftNoSpeech, "", nil
	})); err != nil {
		t.Fatalf("no_speech is a legitimate terminal state with nothing in it, got %v", err)
	}
}

func TestAFailureAndItsCauseImplyEachOther(t *testing.T) {
	// Either alone is a draft that cannot be explained to the author, or a cause nobody shows.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments = DraftFailed, "", nil
	})); err != ErrInvalidRequest {
		t.Fatal("failed without a cause must be refused")
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Failure = "щось зламалось" })); err != ErrInvalidRequest {
		t.Fatal("a cause on a non-failed draft must be refused")
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments, d.Failure = DraftFailed, "", nil, "engine_unavailable"
	})); err != nil {
		t.Fatalf("failed with a cause is the legitimate shape, got %v", err)
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments, d.Failure = DraftFailed, "", nil, strings.Repeat("я", MaximumFailureRunes+1)
	})); err != ErrInvalidRequest {
		t.Fatal("a cause over the ceiling must be refused; the field holds a reason, not a trace")
	}
}

func TestADraftAlwaysNamesTheVideoItCameFrom(t *testing.T) {
	// Unlike a transcript, which may be typed before any video exists, nothing makes a draft
	// except a ready asset - so a missing asset or generation is a defect, not a state.
	for name, mutate := range map[string]func(*TranscriptDraft){
		"no asset":        func(d *TranscriptDraft) { d.AssetID = "" },
		"zero generation": func(d *TranscriptDraft) { d.Generation = 0 },
		"no lesson":       func(d *TranscriptDraft) { d.LessonID = "" },
		"unknown state":   func(d *TranscriptDraft) { d.State = "paused" },
	} {
		if _, err := NewTranscriptDraft(draft(mutate)); err != ErrInvalidRequest {
			t.Fatalf("%s must be refused, got %v", name, err)
		}
	}
}

func TestStalenessIsDerivedAndAgreesWithTheTranscriptRule(t *testing.T) {
	// D8-4: the same rule as Transcript.Stale, so a draft and a transcript of the same lesson
	// cannot disagree about which one is out of date.
	made, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Generation = 2 }))
	if err != nil {
		t.Fatal(err)
	}
	if made.Stale(2) {
		t.Fatal("the generation the player would pick is this one; not stale")
	}
	if !made.Stale(3) {
		t.Fatal("the video was replaced; stale")
	}
	// No ready asset at all must not read as stale — the same exception the transcript makes,
	// so that a lesson whose transcode failed does not mark every draft out of date.
	if made.Stale(0) {
		t.Fatal("a lesson with no ready asset must not read as stale")
	}
	if (Transcript{AuthoredAgainstGeneration: 2}).Stale(0) != made.Stale(0) {
		t.Fatal("the draft and the transcript must answer the no-ready-asset case the same way")
	}
}

func TestTerminalNamesExactlyTheStatesTheWorkerIsDoneWith(t *testing.T) {
	// Recovery reads this to tell a draft abandoned mid-flight from one that finished (P-5).
	for state, want := range map[DraftState]bool{
		DraftQueued: false, DraftRunning: false,
		DraftReady: true, DraftFailed: true, DraftNoSpeech: true,
	} {
		if state.Terminal() != want {
			t.Fatalf("%s.Terminal() = %v; want %v", state, state.Terminal(), want)
		}
	}
}

func TestTheCeilingsAreInclusiveOnBothSidesOfTheBoundary(t *testing.T) {
	// Codex found this gap by asking for a mutation the suite would NOT catch: the earlier tests
	// checked 5001 segments and a one-line draft, never exactly 5000. Turning `>` into `>=` would
	// therefore stay green while the domain started refusing a count the column's own
	// `jsonb_array_length <= 5000` accepts - the two would disagree, and only in production.
	exact := make([]Segment, MaximumSegments)
	for index := range exact {
		// Distinct ids, because a repeat is refused before the ceiling is ever reached. The last
		// four hex digits carry the index, which keeps every id a valid UUID.
		exact[index] = Segment{
			ID:      fmt.Sprintf("11111111-1111-4111-8111-11111111%04x", index),
			StartMs: int64(index), EndMs: int64(index) + 1, Text: "текст",
		}
	}
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Segments = exact })); err != nil {
		t.Fatalf("exactly MaximumSegments must be accepted, got %v", err)
	}

	// The same boundary for the failure cause: exactly the ceiling is legitimate, and the
	// migration's `char_length(failure) BETWEEN 1 AND 200` agrees.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments = DraftFailed, "", nil
		d.Failure = strings.Repeat("я", MaximumFailureRunes)
	})); err != nil {
		t.Fatalf("exactly MaximumFailureRunes must be accepted, got %v", err)
	}
}

func TestLanguageIsStoredNormalised(t *testing.T) {
	// Codex found this one: the constructor's result was discarded and the raw value kept, so
	// ` EN ` reached the column whose CHECK is `language = lower(language)` and the INSERT failed
	// - and `AUTO`, not being equal to `auto`, walked straight past the guard that exists to stop
	// exactly it (D8-8).
	made, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Language = " EN " }))
	if err != nil || made.Language != "en" {
		t.Fatalf("language must be stored normalised; got %q, %v", made.Language, err)
	}
	for _, raw := range []Language{"auto", "AUTO", " Auto "} {
		if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) { d.Language = raw })); err != ErrInvalidRequest {
			t.Fatalf("%q must be refused whatever its case, got %v", raw, err)
		}
	}
}

func TestACauseCannotCarryWhatPostgresCannotStore(t *testing.T) {
	// A NUL is valid UTF-8 and PostgreSQL `text` cannot hold it. Letting one through would fail
	// at the INSERT, which `Complete` reports as transient - three retries of a job that can
	// never succeed.
	if _, err := NewTranscriptDraft(draft(func(d *TranscriptDraft) {
		d.State, d.Language, d.Segments, d.Failure = DraftFailed, "", nil, "engine\x00crashed"
	})); err != ErrInvalidRequest {
		t.Fatalf("a NUL in the cause must be refused, got %v", err)
	}
}
