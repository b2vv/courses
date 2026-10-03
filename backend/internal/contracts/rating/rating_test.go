package rating

import "testing"

func TestSummary(t *testing.T) {
	for _, test := range []struct {
		name     string
		count    int
		starsSum int
		average  float64
		visible  bool
	}{
		{name: "zero", count: 0, starsSum: 0},
		{name: "one", count: 1, starsSum: 5},
		{name: "two", count: 2, starsSum: 9},
		{name: "threshold rounds 4.666 up", count: 3, starsSum: 14, average: 4.7, visible: true},
		{name: "exact half stays at one decimal", count: 4, starsSum: 18, average: 4.5, visible: true},
		{name: "five stars", count: 3, starsSum: 15, average: 5.0, visible: true},
		{name: "x.x5 rounds half up", count: 20, starsSum: 87, average: 4.4, visible: true},
		{name: "large", count: 1000000, starsSum: 4566667, average: 4.6, visible: true},
		{name: "negative count", count: -1, starsSum: 0},
		{name: "sum below minimum", count: 3, starsSum: 2},
		{name: "sum above maximum", count: 3, starsSum: 16},
	} {
		t.Run(test.name, func(t *testing.T) {
			average, visible := Summary(test.count, test.starsSum)
			if average != test.average || visible != test.visible {
				t.Fatalf("Summary(%d, %d) = (%v, %v), want (%v, %v)", test.count, test.starsSum, average, visible, test.average, test.visible)
			}
		})
	}
}
