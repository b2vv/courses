// Package rating contains pure cross-module rating display rules.
package rating

const MinimumVisibleCount = 3

// Summary decides whether a rating is shown and returns its one-decimal,
// half-up average. count is the number of counted reviews and starsSum is the
// sum of their 1..5 star values.
func Summary(count int, starsSum int) (average float64, visible bool) {
	if count < 0 || starsSum < count || starsSum > 5*count {
		return 0, false
	}
	if count < MinimumVisibleCount {
		return 0, false
	}
	tenths := (starsSum*20 + count) / (2 * count)
	return float64(tenths) / 10, true
}
