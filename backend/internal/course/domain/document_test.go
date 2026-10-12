package domain_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"courses/internal/course/domain"
)

func documentWith(sections ...domain.Section) domain.LessonDocument {
	return domain.LessonDocument{SchemaVersion: domain.CurrentLessonDocumentSchemaVersion, Locale: "uk", Sections: sections}
}

func textSection() domain.Section {
	return domain.Section{ID: "text", Type: domain.SectionText, Content: &domain.RichText{Blocks: []domain.Block{{Type: domain.BlockParagraph, Spans: []domain.Span{{Text: "Практика"}}}}}}
}

func imageSection() domain.Section {
	return domain.Section{ID: "image", Type: domain.SectionImage, ImageID: "lesson-image", Alt: "Зображення практики"}
}

func checkDocument(t *testing.T, document domain.LessonDocument, valid bool) {
	t.Helper()
	err := domain.ValidateLessonDocument(document)
	if valid && err != nil {
		t.Fatalf("valid document rejected: %v", err)
	}
	if !valid && !errors.Is(err, domain.ErrInvalidLessonDocument) {
		t.Fatalf("got %v, want ErrInvalidLessonDocument", err)
	}
}

func TestLessonDocumentStructure(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name     string
		document domain.LessonDocument
		valid    bool
	}{
		{"empty", documentWith(), true},
		{"English", domain.LessonDocument{SchemaVersion: 1, Locale: "en"}, true},
		{"missing version", domain.LessonDocument{Locale: "uk"}, false},
		{"future version", domain.LessonDocument{SchemaVersion: 2, Locale: "uk"}, false},
		{"missing locale", domain.LessonDocument{SchemaVersion: 1}, false},
		{"unknown locale", domain.LessonDocument{SchemaVersion: 1, Locale: "fr"}, false},
		{"one video", documentWith(domain.Section{ID: "v", Type: domain.SectionVideo}), true},
		{"two videos", documentWith(domain.Section{ID: "v1", Type: domain.SectionVideo}, domain.Section{ID: "v2", Type: domain.SectionVideo}), false},
		{"unknown section AC11", documentWith(domain.Section{ID: "x", Type: "iframe"}), false},
		{"missing section type", documentWith(domain.Section{ID: "x"}), false},
		{"text and image without video", documentWith(textSection(), imageSection()), true},
	} {
		t.Run(test.name, func(t *testing.T) { checkDocument(t, test.document, test.valid) })
	}
}

// Acceptance rows 5 and 12: rows allow only one level of text/image columns.
func TestLessonDocumentRows(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		columns []domain.Section
		valid   bool
	}{
		{"one", []domain.Section{textSection()}, true},
		{"two", []domain.Section{imageSection(), textSection()}, true},
		{"three", []domain.Section{textSection(), imageSection(), {ID: "third", Type: domain.SectionText, Content: &domain.RichText{}}}, true},
		{"zero", nil, false},
		{"four", []domain.Section{textSection(), textSection(), textSection(), textSection()}, false},
		{"nested row", []domain.Section{{ID: "nested", Type: domain.SectionRow, Columns: []domain.Section{textSection()}}}, false},
		{"video", []domain.Section{{ID: "v", Type: domain.SectionVideo}}, false},
		{"task", []domain.Section{{ID: "task", Type: domain.SectionTask, Content: &domain.RichText{}}}, false},
		{"unknown", []domain.Section{{ID: "x", Type: "iframe"}}, false},
		{"invalid child", []domain.Section{{ID: "image", Type: domain.SectionImage, ImageID: "i"}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkDocument(t, documentWith(domain.Section{ID: "row", Type: domain.SectionRow, Columns: test.columns}), test.valid)
		})
	}
	// Even a cyclic in-memory input is refused before descending into a second row.
	row := domain.Section{ID: "cycle", Type: domain.SectionRow, Columns: make([]domain.Section, 1)}
	row.Columns[0] = row
	checkDocument(t, documentWith(row), false)
}

func TestLessonDocumentSectionFields(t *testing.T) {
	t.Parallel()
	for _, kind := range []domain.SectionType{domain.SectionVideo, domain.SectionText, domain.SectionTask, domain.SectionImage, domain.SectionRow} {
		valid := domain.Section{ID: "section", Type: kind}
		switch kind {
		case domain.SectionText, domain.SectionTask:
			valid.Content = &domain.RichText{}
		case domain.SectionImage:
			valid.ImageID, valid.Alt = "image", "Alt"
		case domain.SectionRow:
			valid.Columns = []domain.Section{textSection()}
		}
		t.Run(string(kind), func(t *testing.T) {
			checkDocument(t, documentWith(valid), true)
			for _, test := range []struct {
				name   string
				change func(*domain.Section)
			}{
				{"missing id", func(s *domain.Section) { s.ID = "" }},
				{"blank id", func(s *domain.Section) { s.ID = " " }},
				{"long id", func(s *domain.Section) { s.ID = strings.Repeat("i", domain.MaximumSectionID+1) }},
				{"content", func(s *domain.Section) {
					if s.Content == nil {
						s.Content = &domain.RichText{}
					} else {
						s.Content = nil
					}
				}},
				{"image id", func(s *domain.Section) {
					if s.ImageID == "" {
						s.ImageID = "image"
					} else {
						s.ImageID = ""
					}
				}},
				{"alt", func(s *domain.Section) {
					if s.Alt == "" {
						s.Alt = "alt"
					} else {
						s.Alt = ""
					}
				}},
				{"columns", func(s *domain.Section) { s.Columns = []domain.Section{} }},
			} {
				t.Run(test.name, func(t *testing.T) { s := valid; test.change(&s); checkDocument(t, documentWith(s), false) })
			}
		})
	}
}

func TestLessonDocumentRichText(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		block domain.Block
		valid bool
	}{
		{"empty paragraph", domain.Block{Type: domain.BlockParagraph}, true},
		{"empty h2", domain.Block{Type: domain.BlockHeading, Level: 2}, true},
		{"empty h3", domain.Block{Type: domain.BlockHeading, Level: 3}, true},
		{"h4 AC13", domain.Block{Type: domain.BlockHeading, Level: 4}, false},
		{"missing heading level", domain.Block{Type: domain.BlockHeading}, false},
		{"unknown block", domain.Block{Type: "script"}, false},
		{"unordered list", domain.Block{Type: domain.BlockList, Items: [][]domain.Span{{{Text: "item"}}}}, true},
		{"ordered list", domain.Block{Type: domain.BlockList, Ordered: true, Items: [][]domain.Span{{{Text: "item"}}}}, true},
		{"empty list item", domain.Block{Type: domain.BlockList, Items: [][]domain.Span{nil}}, true},
		{"no list items", domain.Block{Type: domain.BlockList}, false},
		{"paragraph level", domain.Block{Type: domain.BlockParagraph, Level: 2}, false},
		{"paragraph ordered", domain.Block{Type: domain.BlockParagraph, Ordered: true}, false},
		{"paragraph items", domain.Block{Type: domain.BlockParagraph, Items: [][]domain.Span{}}, false},
		{"heading items", domain.Block{Type: domain.BlockHeading, Level: 2, Items: [][]domain.Span{}}, false},
		{"heading ordered", domain.Block{Type: domain.BlockHeading, Level: 2, Ordered: true}, false},
		{"list level", domain.Block{Type: domain.BlockList, Level: 2, Items: [][]domain.Span{nil}}, false},
		{"list spans", domain.Block{Type: domain.BlockList, Spans: []domain.Span{}, Items: [][]domain.Span{nil}}, false},
		{"invalid list span", domain.Block{Type: domain.BlockList, Items: [][]domain.Span{{{Text: "x", Href: "javascript:alert(1)"}}}}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := textSection()
			s.Content.Blocks = []domain.Block{test.block}
			checkDocument(t, documentWith(s), test.valid)
		})
	}
}

func TestLessonDocumentSpans(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		span  domain.Span
		valid bool
	}{
		{"empty", domain.Span{}, true},
		{"whitespace", domain.Span{Text: " \n "}, true},
		{"literal HTML", domain.Span{Text: `<script>alert(1)</script> & <b>literal</b>`}, true},
		{"both marks", domain.Span{Text: "x", Marks: []domain.Mark{domain.MarkBold, domain.MarkItalic}}, true},
		{"unknown mark", domain.Span{Text: "x", Marks: []domain.Mark{"underline"}}, false},
		{"too many marks", domain.Span{Marks: []domain.Mark{domain.MarkBold, domain.MarkBold, domain.MarkBold}}, false},
		{"invalid UTF8", domain.Span{Text: string([]byte{0xff})}, false},
		{"NUL", domain.Span{Text: "a\x00b"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			s := textSection()
			s.Content.Blocks[0].Spans = []domain.Span{test.span}
			d := documentWith(s)
			checkDocument(t, d, test.valid)
			if !reflect.DeepEqual(d.Sections[0].Content.Blocks[0].Spans[0], test.span) {
				t.Fatal("validation changed span")
			}
		})
	}
	// Acceptance row 13 covers URL rejection, not HTML escaping (LB11).
	for _, test := range []struct {
		href  string
		valid bool
	}{
		{"", true}, {"https://ex.com", true}, {"http://ex.com/a?q=x#part", true},
		{"javascript:alert(1)", false}, {"data:text/html,hello", false}, {"/x", false},
		{"//ex.com", false}, {"ftp://ex.com", false}, {"https:ex.com", false},
		{"https://", false}, {"https://:443", false}, {"https://exa mple.com", false},
		{"https://ex.com/%zz", false}, {"https://ex.com:bad", false}, {"https://[::1", false},
		{" https://ex.com", false}, {"https://ex.com\n", false},
	} {
		t.Run("href_"+test.href, func(t *testing.T) {
			s := textSection()
			s.Content.Blocks[0].Spans[0].Href = test.href
			checkDocument(t, documentWith(s), test.valid)
		})
	}
}

func TestLessonDocumentBounds(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name    string
		maximum int
		build   func(int) domain.LessonDocument
	}{
		{"sections", domain.MaximumDocumentSections, func(n int) domain.LessonDocument {
			d := documentWith()
			for i := range n {
				s := textSection()
				s.ID = fmt.Sprint(i)
				d.Sections = append(d.Sections, s)
			}
			return d
		}},
		{"sections including row children", domain.MaximumDocumentSections, func(n int) domain.LessonDocument {
			d := documentWith(domain.Section{ID: "row", Type: domain.SectionRow, Columns: []domain.Section{textSection()}})
			for i := 2; i < n; i++ {
				s := textSection()
				s.ID = fmt.Sprint(i)
				d.Sections = append(d.Sections, s)
			}
			return d
		}},
		{"section ID", domain.MaximumSectionID, func(n int) domain.LessonDocument {
			s := textSection()
			s.ID = strings.Repeat("ї", n)
			return documentWith(s)
		}},
		{"image ID", domain.MaximumSectionID, func(n int) domain.LessonDocument {
			s := imageSection()
			s.ImageID = strings.Repeat("i", n)
			return documentWith(s)
		}},
		{"alt", domain.MaximumImageAlt, func(n int) domain.LessonDocument {
			s := imageSection()
			s.Alt = strings.Repeat("ї", n)
			return documentWith(s)
		}},
		{"blocks", domain.MaximumRichTextBlocks, func(n int) domain.LessonDocument {
			s := textSection()
			s.Content.Blocks = make([]domain.Block, n)
			for i := range s.Content.Blocks {
				s.Content.Blocks[i].Type = domain.BlockParagraph
			}
			return documentWith(s)
		}},
		{"list items", domain.MaximumListItems, func(n int) domain.LessonDocument {
			s := textSection()
			s.Content.Blocks = []domain.Block{{Type: domain.BlockList, Items: make([][]domain.Span, n)}}
			return documentWith(s)
		}},
		{"spans", domain.MaximumBlockSpans, func(n int) domain.LessonDocument {
			s := textSection()
			s.Content.Blocks[0].Spans = make([]domain.Span, n)
			return documentWith(s)
		}},
		{"list item spans", domain.MaximumBlockSpans, func(n int) domain.LessonDocument {
			s := textSection()
			s.Content.Blocks = []domain.Block{{Type: domain.BlockList, Items: [][]domain.Span{make([]domain.Span, n)}}}
			return documentWith(s)
		}},
		{"span text", domain.MaximumSpanText, func(n int) domain.LessonDocument {
			s := textSection()
			s.Content.Blocks[0].Spans[0].Text = strings.Repeat("ї", n)
			return documentWith(s)
		}},
		{"href", domain.MaximumSpanHref, func(n int) domain.LessonDocument {
			s := textSection()
			prefix := "https://ex.com/"
			s.Content.Blocks[0].Spans[0].Href = prefix + strings.Repeat("a", n-len(prefix))
			return documentWith(s)
		}},
		{"document text across rows and tasks", domain.MaximumDocumentText, func(n int) domain.LessonDocument {
			d := documentWith(domain.Section{ID: "row", Type: domain.SectionRow, Columns: []domain.Section{{ID: "image", Type: domain.SectionImage, ImageID: "i", Alt: "ї"}}})
			remaining := n - 1
			for remaining > 0 {
				size := min(remaining, domain.MaximumSpanText)
				s := textSection()
				s.ID = fmt.Sprint(remaining)
				s.Type = domain.SectionTask
				s.Content.Blocks[0].Spans[0].Text = strings.Repeat("ї", size)
				d.Sections = append(d.Sections, s)
				remaining -= size
			}
			return d
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			checkDocument(t, test.build(test.maximum), true)
			checkDocument(t, test.build(test.maximum+1), false)
		})
	}
}

func TestLessonDocumentEsoterica(t *testing.T) {
	t.Parallel()
	text := textSection()
	text.Content.Blocks = []domain.Block{
		{Type: domain.BlockHeading, Level: 2, Spans: []domain.Span{{Text: "Практика", Marks: []domain.Mark{domain.MarkBold}}}},
		{Type: domain.BlockParagraph, Spans: []domain.Span{{Text: "Замовляння", Marks: []domain.Mark{domain.MarkItalic}}}},
		{Type: domain.BlockList, Ordered: true, Items: [][]domain.Span{{{Text: "Перший крок"}}, {{Text: "Другий крок"}}}},
	}
	task := domain.Section{ID: "task", Type: domain.SectionTask, Content: &domain.RichText{Blocks: []domain.Block{{Type: domain.BlockParagraph, Spans: []domain.Span{{Text: "Виконайте", Marks: []domain.Mark{domain.MarkBold}, Href: "https://ex.com"}}}}}}
	column := textSection()
	column.ID = "explanation"
	checkDocument(t, documentWith(domain.Section{ID: "video", Type: domain.SectionVideo}, text, domain.Section{ID: "row", Type: domain.SectionRow, Columns: []domain.Section{imageSection(), column}}, task), true)
}
