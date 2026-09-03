package domain_test

import (
	"errors"
	"strings"
	"testing"

	"courses/internal/course/domain"
)

func TestCatalogQueryBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name          string
		limit, offset int
		search, sort  string
		valid         bool
	}{
		{"minimum", 1, 0, "", "recent", true},
		{"maximum", 50, 10000, strings.Repeat("ї", 100), "title", true},
		{"trimmed Unicode", 20, 0, "  Go  ", "recent", true},
		{"zero limit", 0, 0, "", "recent", false},
		{"too many", 51, 0, "", "recent", false},
		{"negative offset", 20, -1, "", "recent", false},
		{"huge offset", 20, 10001, "", "recent", false},
		{"long Unicode", 20, 0, strings.Repeat("ї", 101), "recent", false},
		{"unknown sort", 20, 0, "", "popular", false},
		{"invalid UTF8", 20, 0, string([]byte{0xff}), "recent", false},
		{"NUL text", 20, 0, "Go\x00", "recent", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			query, err := domain.NewQuery(tc.limit, tc.offset, tc.search, tc.sort)
			if tc.valid {
				if err != nil {
					t.Fatal(err)
				}
				if query.Search != strings.TrimSpace(tc.search) {
					t.Fatalf("query not normalized: %+v", query)
				}
			} else if !errors.Is(err, domain.ErrInvalidQuery) {
				t.Fatalf("expected invalid query, got %v", err)
			}
		})
	}
}

func TestCanonicalSlugBoundaries(t *testing.T) {
	for _, tc := range []struct {
		slug  string
		valid bool
	}{
		{"a", true}, {"course-123", true}, {strings.Repeat("a", 120), true},
		{"", false}, {"-course", false}, {"course-", false}, {"course--one", false}, {"Uppercase", false}, {"курс", false}, {"course.name", false}, {strings.Repeat("a", 121), false},
	} {
		err := domain.ValidateSlug(tc.slug)
		if (err == nil) != tc.valid {
			t.Errorf("slug %q validity=%t err=%v", tc.slug, tc.valid, err)
		}
	}
}
