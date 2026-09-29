package domain

import (
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	MaximumBodyRunes  = 2000
	MaximumReplyCount = 200
)

var (
	ErrBodyInvalid     = errors.New("comment_body_invalid")
	ErrPositionInvalid = errors.New("comment_position_invalid")
	ErrParentInvalid   = errors.New("comment_parent_invalid")
	ErrThreadFull      = errors.New("comment_thread_full")
	ErrInvalidCursor   = errors.New("comment_cursor_invalid")
)

type Comment struct {
	ID         string
	LessonID   string
	AuthorID   string
	ParentID   *string
	Body       string
	PositionMs *int64
	CreatedAt  time.Time
	EditedAt   *time.Time
	RemovedAt  *time.Time
	RemovedBy  *string
}

// NewComment validates a root or a flat reply. parent is required exactly when ParentID is set;
// the application supplies it from its own lesson-scoped read before the domain is called.
func NewComment(input Comment, parent *Comment, durationSeconds *float64) (Comment, error) {
	body := strings.TrimSpace(input.Body)
	if !utf8.ValidString(body) || utf8.RuneCountInString(body) < 1 || utf8.RuneCountInString(body) > MaximumBodyRunes {
		return Comment{}, ErrBodyInvalid
	}
	if input.ParentID == nil {
		if parent != nil {
			return Comment{}, ErrParentInvalid
		}
		if input.PositionMs != nil && !validPosition(*input.PositionMs, durationSeconds) {
			return Comment{}, ErrPositionInvalid
		}
	} else {
		if parent == nil || parent.ID != *input.ParentID || parent.LessonID != input.LessonID || parent.ParentID != nil || parent.RemovedAt != nil {
			return Comment{}, ErrParentInvalid
		}
		if input.PositionMs != nil {
			return Comment{}, ErrPositionInvalid
		}
	}
	input.Body = body
	return input, nil
}

func validPosition(positionMs int64, durationSeconds *float64) bool {
	return positionMs >= 0 && durationSeconds != nil && float64(positionMs) <= *durationSeconds*1000
}

func ValidateReplyCount(nonRemovedReplies int) error {
	if nonRemovedReplies >= MaximumReplyCount {
		return ErrThreadFull
	}
	return nil
}

func (comment Comment) CanBeEditedBy(actor Actor) bool {
	return actor.UserID != "" && actor.UserID == comment.AuthorID
}

func (comment Comment) CanBeRemovedBy(actor Actor, course Course) bool {
	if actor.UserID == "" {
		return false
	}
	return comment.CanBeEditedBy(actor) || actor.Administrator || actor.UserID == course.OwnerID
}
