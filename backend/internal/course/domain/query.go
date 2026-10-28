package domain

import (
	"slices"
	"strings"
	"unicode/utf8"
)

type Sort string

const (
	SortRecent   Sort = "recent"
	SortTitle    Sort = "title"
	SortStudents Sort = "students"
	SortPrice    Sort = "price"
)

// PublicCatalogSorts is the authoritative public allowlist; admin-only sorts stay separate.
var PublicCatalogSorts = []Sort{SortRecent, SortTitle}

// AdminCourseSorts is the authoritative ordered admin allowlist, including both admin-only sorts.
var AdminCourseSorts = []Sort{SortRecent, SortTitle, SortStudents, SortPrice}

type Query struct {
	Limit  int
	Offset int
	Search string
	Sort   Sort
}
type Page struct {
	Items  []Course
	Limit  int
	Offset int
	Total  int64
}

// NewQuery validates a public catalog list query. The catalog offers two sorts.
func NewQuery(limit, offset int, search, sort string) (Query, error) {
	return newQuery(limit, offset, search, sort, PublicCatalogSorts...)
}

// NewAuthoringQuery validates an admin list query. Admin offers two sorts the public
// catalog does not — by active-student count and by price — so they live here, not in
// NewQuery, and a catalog request for them stays an invalid_query.
func NewAuthoringQuery(limit, offset int, search, sort string) (Query, error) {
	return newQuery(limit, offset, search, sort, AdminCourseSorts...)
}

func newQuery(limit, offset int, search, sort string, allowed ...Sort) (Query, error) {
	search = strings.TrimSpace(search)
	if limit < 1 || limit > 50 || offset < 0 || offset > 10000 || !utf8.ValidString(search) || strings.ContainsRune(search, 0) || utf8.RuneCountInString(search) > 100 || !slices.Contains(allowed, Sort(sort)) {
		return Query{}, ErrInvalidQuery
	}
	return Query{Limit: limit, Offset: offset, Search: search, Sort: Sort(sort)}, nil
}
