package domain

import (
	"encoding/hex"
	"strings"
	"time"
	"unicode/utf8"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrInvalidCourse         Error = "invalid_course"
	ErrInvalidQuery          Error = "invalid_query"
	ErrInvalidSlug           Error = "invalid_slug"
	ErrCourseNotFound        Error = "course_not_found"
	ErrDependencyUnavailable Error = "dependency_unavailable"
)

type Status string

const (
	Draft     Status = "draft"
	Published Status = "published"
	Archived  Status = "archived"
)

type Course struct {
	ID          string
	Slug        string
	Title       string
	Summary     string
	Description string
	Status      Status
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
	Price       *Price
	Paid        bool
}

type Price struct {
	Minor    int64
	Currency string
}

type ProgramLesson struct {
	Title string
}

type ProgramModule struct {
	Title   string
	Lessons []ProgramLesson
}

type CourseDetail struct {
	Course
	Program []ProgramModule
}

func NewCourse(course Course) (Course, error) {
	if !validID(course.ID) || ValidateSlug(course.Slug) != nil || !validText(course.Title, 200) || !validText(course.Summary, 500) || !validText(course.Description, 20000) {
		return Course{}, ErrInvalidCourse
	}
	if course.Status != Draft && course.Status != Published && course.Status != Archived {
		return Course{}, ErrInvalidCourse
	}
	if course.CreatedAt.IsZero() || course.UpdatedAt.Before(course.CreatedAt) {
		return Course{}, ErrInvalidCourse
	}
	if course.Status == Published && course.PublishedAt == nil || course.Status == Draft && course.PublishedAt != nil {
		return Course{}, ErrInvalidCourse
	}
	if course.PublishedAt != nil && (course.PublishedAt.Before(course.CreatedAt) || course.PublishedAt.After(course.UpdatedAt)) {
		return Course{}, ErrInvalidCourse
	}
	if course.Price != nil && (course.Price.Minor <= 0 || course.Price.Currency != "UAH") || course.Paid != (course.Price != nil) {
		return Course{}, ErrInvalidCourse
	}
	return course, nil
}

func validID(value string) bool {
	if len(value) != 36 || value[8] != '-' || value[13] != '-' || value[18] != '-' || value[23] != '-' || strings.ToLower(value) != value {
		return false
	}
	decoded, err := hex.DecodeString(strings.ReplaceAll(value, "-", ""))
	return err == nil && len(decoded) == 16 && value != "00000000-0000-0000-0000-000000000000"
}

func validText(value string, maximum int) bool {
	return utf8.ValidString(value) && !strings.ContainsRune(value, 0) && strings.TrimSpace(value) != "" && utf8.RuneCountInString(value) <= maximum
}

func ValidateSlug(slug string) error {
	if len(slug) == 0 || len(slug) > 120 || slug[0] == '-' || slug[len(slug)-1] == '-' {
		return ErrInvalidSlug
	}
	previousHyphen := false
	for _, char := range slug {
		if char == '-' {
			if previousHyphen {
				return ErrInvalidSlug
			}
			previousHyphen = true
		} else {
			if char < 'a' || char > 'z' {
				if char < '0' || char > '9' {
					return ErrInvalidSlug
				}
			}
			previousHyphen = false
		}
	}
	return nil
}
