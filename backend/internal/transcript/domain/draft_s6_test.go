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

func TestDraftStaleFollowsGenerationAndHeardLeaf(t *testing.T) {
	const uploaded = "audio-70000000-0000-4000-8000-000000000001"
	legacy := TranscriptDraft{Generation: 2}
	heardUpload := TranscriptDraft{Generation: 2, SourceLeaf: uploaded}
	for _, test := range []struct {
		name        string
		draft       TranscriptDraft
		generation  int32
		defaultLeaf string
		want        bool
	}{
		{"no ready asset never reads stale", legacy, 0, uploaded, false},
		{"another generation", legacy, 3, OriginalAudioLeaf, true},
		{"a pre-S6 draft heard the original", legacy, 2, OriginalAudioLeaf, false},
		{"a pre-S6 draft after the default moved to an upload", legacy, 2, uploaded, true},
		{"heard the current default", heardUpload, 2, uploaded, false},
		{"default moved back to the original", heardUpload, 2, OriginalAudioLeaf, true},
		{"no default known reads by generation only", heardUpload, 2, "", false},
	} {
		if got := test.draft.Stale(test.generation, test.defaultLeaf); got != test.want {
			t.Errorf("%s: Stale=%v, want %v", test.name, got, test.want)
		}
	}
}
