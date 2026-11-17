package domain

import "testing"

func TestValidAudioLeaf(t *testing.T) {
	for leaf, want := range map[string]bool{
		"original": true,
		"audio-70000000-0000-4000-8000-000000000001": true,
		"":                false,
		"Original":        false,
		"audio-":          false,
		"../original":     false,
		"audio-../../etc": false,
		"audio-70000000-0000-4000-8000-00000000000A": false,
		"audio-00000000-0000-0000-0000-000000000000": false,
		"original/../x": false,
	} {
		if got := ValidAudioLeaf(leaf); got != want {
			t.Errorf("ValidAudioLeaf(%q)=%v, want %v", leaf, got, want)
		}
	}
}

func TestDraftStaleFollowsTrackLanguageAndMode(t *testing.T) {
	for _, test := range []struct {
		name                            string
		snapshot                        Language
		current                         Language
		previousMode, currentMode, want bool
	}{
		{"same tag ignores current case", "en-gb", "EN-GB", false, false, false},
		{"rename", "uk", "en", false, false, true},
		{"und can be assigned a language", "und", "en", false, false, false},
		{"und family can be reassigned", "und-latn", "fr", false, false, false},
		{"und prefix alone is not family", "under", "en", false, false, true},
		{"mode change", "en", "en", false, true, true},
		{"und still invalidates mode", "und", "en", true, false, true},
	} {
		draft := TranscriptDraft{TrackLanguage: test.snapshot, TrackMultilingual: test.previousMode}
		if got := draft.Stale(test.current, test.currentMode); got != test.want {
			t.Errorf("%s: Stale=%v, want %v", test.name, got, test.want)
		}
	}
}
