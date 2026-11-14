// Package language holds the one grammar a language tag must satisfy wherever a tag
// names subtitles: a transcript is keyed by it and an audio track is matched to its
// transcript by it (multi-subtitles D1/D2), so the two modules cannot be allowed to
// disagree about what a tag is. Normalisation stays with each module.
package language

import "strings"

const (
	MinimumTagLength = 2
	MaximumTagLength = 35
)

// ValidTag reports whether an already lowercased value is a BCP-47-shaped tag:
// 2–35 characters, hyphen-separated subtags of 1–8 [a-z0-9], the first subtag at
// least two characters and starting with a letter. It does not trim or lowercase;
// callers decide their own normalisation.
func ValidTag(lowered string) bool {
	if len(lowered) < MinimumTagLength || len(lowered) > MaximumTagLength {
		return false
	}
	for index, part := range strings.Split(lowered, "-") {
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
