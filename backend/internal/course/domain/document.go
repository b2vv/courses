package domain

import (
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"unicode/utf8"
)

const ErrInvalidLessonDocument Error = "invalid_lesson_document"

// String ceilings count Unicode code points, as in lesson authoring. The section
// ceiling includes row children; text is also bounded across the whole document.
const (
	CurrentLessonDocumentSchemaVersion = 1
	MaximumDocumentSections            = 200
	MaximumRowColumns                  = 3
	MaximumSectionID                   = 200
	MaximumImageAlt                    = 2000
	MaximumRichTextBlocks              = 200
	MaximumListItems                   = 200
	MaximumBlockSpans                  = 200
	MaximumSpanText                    = 20000
	MaximumSpanHref                    = 2048
	MaximumSpanMarks                   = 2
	MaximumDocumentText                = 200000
)

type SectionType string

const (
	SectionVideo SectionType = "video"
	SectionText  SectionType = "text"
	SectionImage SectionType = "image"
	SectionRow   SectionType = "row"
	SectionTask  SectionType = "task"
)

// SectionTypes is the complete section discriminator set.
var SectionTypes = []SectionType{SectionVideo, SectionText, SectionImage, SectionRow, SectionTask}

type BlockType string

const (
	BlockParagraph BlockType = "paragraph"
	BlockHeading   BlockType = "heading"
	BlockList      BlockType = "list"
)

// BlockTypes is the complete rich-text block discriminator set.
var BlockTypes = []BlockType{BlockParagraph, BlockHeading, BlockList}

type Mark string

const (
	MarkBold   Mark = "bold"
	MarkItalic Mark = "italic"
)

// Marks is the complete inline formatting mark set.
var Marks = []Mark{MarkBold, MarkItalic}

// Image content types an image section may carry. They mirror the image module's upload allowlist
// (internal/image/domain) — a browser cannot import a Go constant, and a module never imports
// another module's package, so the display contract repeats the three values here. The reader maps
// each to a serve extension to build /media/i/<imageId>.<ext>.
const (
	ImageContentTypeJPEG = "image/jpeg"
	ImageContentTypePNG  = "image/png"
	ImageContentTypeWebP = "image/webp"
)

// ImageContentTypes is the complete image-section content-type allowlist, in the order the
// contract publishes it. validImageContentType and the overlay enum both read it so neither drifts
// (TestDocumentEnumCensus, §16.22 follow-up, 2026-10-29).
var ImageContentTypes = []string{ImageContentTypeJPEG, ImageContentTypePNG, ImageContentTypeWebP}

const (
	LocaleUk = "uk"
	LocaleEn = "en"
)

// LessonDocumentLocales is the complete lesson-document locale set, in contract order; it is the
// one source shared by ValidateLessonDocument and the overlay enum (TestDocumentEnumCensus).
var LessonDocumentLocales = []string{LocaleUk, LocaleEn}

type LessonDocument struct {
	SchemaVersion int       `json:"schemaVersion"`
	Locale        string    `json:"locale"`
	Sections      []Section `json:"sections"`
}

type Section struct {
	ID          string      `json:"id"`
	Type        SectionType `json:"type"`
	Content     *RichText   `json:"content,omitempty"`
	ImageID     string      `json:"imageId,omitempty"`
	ContentType string      `json:"contentType,omitempty"`
	Alt         string      `json:"alt,omitempty"`
	Columns     []Section   `json:"columns,omitempty"`
}

type RichText struct {
	Blocks []Block `json:"blocks"`
}

type Block struct {
	Type    BlockType `json:"type"`
	Level   int       `json:"level,omitempty"`
	Ordered bool      `json:"ordered,omitempty"`
	Spans   []Span    `json:"spans,omitempty"`
	Items   [][]Span  `json:"items,omitempty"`
}

// MarshalJSON serializes "ordered" (true or false) only for list blocks, where
// it is a required discriminator the reader relies on, and omits it entirely for
// paragraph and heading blocks, which share this struct but whose strict reader
// schema forbids the key. A plain `json:"ordered"` tag cannot express this: it
// would emit the field for every block type; `omitempty` instead drops a bullet
// list's `ordered:false`, which the reader then rejects (audit H2 / SA02).
func (b Block) MarshalJSON() ([]byte, error) {
	wire := struct {
		Type    BlockType `json:"type"`
		Level   int       `json:"level,omitempty"`
		Ordered *bool     `json:"ordered,omitempty"`
		Spans   []Span    `json:"spans,omitempty"`
		Items   [][]Span  `json:"items,omitempty"`
	}{Type: b.Type, Level: b.Level, Spans: b.Spans, Items: b.Items}
	if b.Type == BlockList {
		wire.Ordered = &b.Ordered
	}
	return json.Marshal(wire)
}

// Text stays literal: HTML escaping belongs to the reader, not stored content.
type Span struct {
	Text  string `json:"text"`
	Marks []Mark `json:"marks,omitempty"`
	Href  string `json:"href,omitempty"`
}

func ValidateLessonDocument(document LessonDocument) error {
	if document.SchemaVersion != CurrentLessonDocumentSchemaVersion || !slices.Contains(LessonDocumentLocales, document.Locale) {
		return ErrInvalidLessonDocument
	}
	sectionCount, videoCount, textCount := len(document.Sections), 0, 0
	if sectionCount > MaximumDocumentSections {
		return ErrInvalidLessonDocument
	}
	for i, section := range document.Sections {
		sectionCount += len(section.Columns)
		if section.Type == SectionVideo {
			videoCount++
		}
		if sectionCount > MaximumDocumentSections || videoCount > 1 {
			return ErrInvalidLessonDocument
		}
		if err := validateDocumentSection(section, false, &textCount); err != nil {
			return fmt.Errorf("section %d: %w", i, err)
		}
	}
	return nil
}

func validateDocumentSection(section Section, inRow bool, textCount *int) error {
	if !slices.Contains(SectionTypes, section.Type) || !validText(section.ID, MaximumSectionID) || inRow && section.Type != SectionText && section.Type != SectionImage {
		return ErrInvalidLessonDocument
	}
	switch section.Type {
	case SectionVideo:
		if section.Content != nil || section.ImageID != "" || section.ContentType != "" || section.Alt != "" || section.Columns != nil {
			return ErrInvalidLessonDocument
		}
	case SectionText, SectionTask:
		if section.Content == nil || section.ImageID != "" || section.ContentType != "" || section.Alt != "" || section.Columns != nil {
			return ErrInvalidLessonDocument
		}
		return validateRichText(*section.Content, textCount)
	case SectionImage:
		// The reader derives the serve extension from contentType to build /media/i/<id>.<ext>,
		// so an image section carries exactly one content type from the fixed allowlist.
		if !validText(section.ImageID, MaximumSectionID) || !validImageContentType(section.ContentType) || !validText(section.Alt, MaximumImageAlt) || section.Content != nil || section.Columns != nil {
			return ErrInvalidLessonDocument
		}
		*textCount += utf8.RuneCountInString(section.Alt)
	case SectionRow:
		if section.Content != nil || section.ImageID != "" || section.ContentType != "" || section.Alt != "" || len(section.Columns) < 1 || len(section.Columns) > MaximumRowColumns {
			return ErrInvalidLessonDocument
		}
		for i, column := range section.Columns {
			if err := validateDocumentSection(column, true, textCount); err != nil {
				return fmt.Errorf("column %d: %w", i, err)
			}
		}
	default:
		return ErrInvalidLessonDocument
	}
	if *textCount > MaximumDocumentText {
		return ErrInvalidLessonDocument
	}
	return nil
}

func validateRichText(content RichText, textCount *int) error {
	if len(content.Blocks) > MaximumRichTextBlocks {
		return ErrInvalidLessonDocument
	}
	for i, block := range content.Blocks {
		if err := validateDocumentBlock(block, textCount); err != nil {
			return fmt.Errorf("block %d: %w", i, err)
		}
	}
	return nil
}

func validateDocumentBlock(block Block, textCount *int) error {
	if !slices.Contains(BlockTypes, block.Type) {
		return ErrInvalidLessonDocument
	}
	switch block.Type {
	case BlockParagraph, BlockHeading:
		if block.Ordered || block.Items != nil || block.Type == BlockParagraph && block.Level != 0 || block.Type == BlockHeading && block.Level != 2 && block.Level != 3 {
			return ErrInvalidLessonDocument
		}
		return validateDocumentSpans(block.Spans, textCount)
	case BlockList:
		if block.Level != 0 || block.Spans != nil || len(block.Items) < 1 || len(block.Items) > MaximumListItems {
			return ErrInvalidLessonDocument
		}
		for _, item := range block.Items {
			if err := validateDocumentSpans(item, textCount); err != nil {
				return err
			}
		}
	default:
		return ErrInvalidLessonDocument
	}
	return nil
}

func validateDocumentSpans(spans []Span, textCount *int) error {
	if len(spans) > MaximumBlockSpans {
		return ErrInvalidLessonDocument
	}
	for _, span := range spans {
		if !validOptionalText(span.Text, MaximumSpanText) || len(span.Marks) > MaximumSpanMarks || !validDocumentHref(span.Href) {
			return ErrInvalidLessonDocument
		}
		for _, mark := range span.Marks {
			if !slices.Contains(Marks, mark) {
				return ErrInvalidLessonDocument
			}
		}
		*textCount += utf8.RuneCountInString(span.Text)
		if *textCount > MaximumDocumentText {
			return ErrInvalidLessonDocument
		}
	}
	return nil
}

func validDocumentHref(href string) bool {
	if href == "" {
		return true
	}
	if !validOptionalText(href, MaximumSpanHref) {
		return false
	}
	parsed, err := url.Parse(href)
	return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != ""
}

func validImageContentType(contentType string) bool {
	return slices.Contains(ImageContentTypes, contentType)
}
