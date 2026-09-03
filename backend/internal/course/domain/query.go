package domain

import (
	"strings"
	"unicode/utf8"
)

type Sort string

const (
	SortRecent Sort = "recent"
	SortTitle  Sort = "title"
)

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

func NewQuery(limit, offset int, search, sort string) (Query, error) {
	search = strings.TrimSpace(search)
	if limit < 1 || limit > 50 || offset < 0 || offset > 10000 || !utf8.ValidString(search) || strings.ContainsRune(search, 0) || utf8.RuneCountInString(search) > 100 || sort != string(SortRecent) && sort != string(SortTitle) {
		return Query{}, ErrInvalidQuery
	}
	return Query{Limit: limit, Offset: offset, Search: search, Sort: Sort(sort)}, nil
}
