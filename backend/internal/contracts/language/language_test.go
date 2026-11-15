package language

import (
	"strings"
	"testing"
)

// The table the frontend mirrors in frontend/src/modules/media/domain/audio-track.test.ts.
func TestValidTag(t *testing.T) {
	for _, value := range []string{"uk", "en-us", "pt-br", "und", "und-x-original", "zh-hant-tw", "de-1996", strings.Repeat("a", 8) + "-" + strings.Repeat("b", 8) + "-" + strings.Repeat("c", 8) + "-" + strings.Repeat("d", 8)} {
		if !ValidTag(value) {
			t.Errorf("ValidTag(%q) = false, want true", value)
		}
	}
	for _, value := range []string{"", "e", "1x", "-en", "en-", "a--b", "und-", strings.Repeat("x", 35), strings.Repeat("x", 36), "uk_ua", "укр", "en us", "EN", " uk", "abcdefghi"} {
		if ValidTag(value) {
			t.Errorf("ValidTag(%q) = true, want false", value)
		}
	}
}
