package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/contracts/access/accesstest"
)

func TestNewCommentValidatesBodyByUnicodeRunesAndTrims(t *testing.T) {
	tests := []struct {
		name string
		body string
		want string
		err  error
	}{
		{name: "empty", body: "   ", err: ErrBodyInvalid},
		{name: "one rune", body: " x ", want: "x"},
		{name: "two thousand runes", body: strings.Repeat("ї", 2000), want: strings.Repeat("ї", 2000)},
		{name: "two thousand and one runes", body: strings.Repeat("ї", 2001), err: ErrBodyInvalid},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			comment, err := NewComment(Comment{LessonID: "lesson", AuthorID: "author", Body: test.body}, nil, nil)
			if !errors.Is(err, test.err) {
				t.Fatalf("error=%v want=%v", err, test.err)
			}
			if test.err == nil && comment.Body != test.want {
				t.Fatalf("body=%q want=%q", comment.Body, test.want)
			}
		})
	}
}

func TestNewCommentValidatesPositionInMilliseconds(t *testing.T) {
	duration := func(seconds float64) *float64 { return &seconds }
	position := func(milliseconds int64) *int64 { return &milliseconds }
	for _, test := range []struct {
		name     string
		position *int64
		duration *float64
		wantErr  error
	}{
		{name: "root without label", duration: nil},
		{name: "zero", position: position(0), duration: duration(0)},
		{name: "exact end", position: position(100000), duration: duration(100)},
		{name: "fractional exact end", position: position(1234), duration: duration(1.234)},
		{name: "fractional end plus one", position: position(1235), duration: duration(1.234), wantErr: ErrPositionInvalid},
		{name: "negative", position: position(-1), duration: duration(10), wantErr: ErrPositionInvalid},
		{name: "missing duration", position: position(1), duration: nil, wantErr: ErrPositionInvalid},
		{name: "reply cannot have label", position: position(1), duration: duration(10), wantErr: ErrPositionInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			parent := (*Comment)(nil)
			if test.name == "reply cannot have label" {
				parent = &Comment{ID: "parent", LessonID: "lesson", Body: "root"}
			}
			_, err := NewComment(Comment{LessonID: "lesson", AuthorID: "author", Body: "body", ParentID: parentID(parent), PositionMs: test.position}, parent, test.duration)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v want=%v", err, test.wantErr)
			}
		})
	}
}

func TestNewCommentValidatesFlatReplyParentAndCapacity(t *testing.T) {
	root := &Comment{ID: "root", LessonID: "lesson", Body: "root"}
	validReply := Comment{LessonID: "lesson", AuthorID: "author", Body: "reply", ParentID: stringPointer("root")}
	if _, err := NewComment(validReply, root, nil); err != nil {
		t.Fatalf("valid reply: %v", err)
	}
	for _, test := range []struct {
		name   string
		parent *Comment
		input  Comment
		want   error
	}{
		{name: "unknown parent", input: validReply, want: ErrParentInvalid},
		{name: "different lesson", parent: &Comment{ID: "root", LessonID: "other", Body: "root"}, input: validReply, want: ErrParentInvalid},
		{name: "nested parent", parent: &Comment{ID: "root", LessonID: "lesson", ParentID: stringPointer("another"), Body: "root"}, input: validReply, want: ErrParentInvalid},
		{name: "removed parent", parent: &Comment{ID: "root", LessonID: "lesson", Body: "root", RemovedAt: timePointer(time.Now())}, input: validReply, want: ErrParentInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewComment(test.input, test.parent, nil)
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
		})
	}
	if err := ValidateReplyCount(199); err != nil {
		t.Fatalf("199 replies: %v", err)
	}
	if !errors.Is(ValidateReplyCount(200), ErrThreadFull) {
		t.Fatal("200 replies should fill the thread")
	}
}

func TestCanBeViewedByMatchesAccessTruthTable(t *testing.T) {
	const owner = "40000000-0000-4000-8000-000000000002"
	const stranger = "40000000-0000-4000-8000-000000000009"
	for _, test := range accesstest.TruthTable {
		t.Run(test.Name, func(t *testing.T) {
			actorID := stranger
			if test.Role == accesstest.Owner {
				actorID = owner
			}
			allowed := Course{
				OwnerID:  owner,
				Status:   Status(test.Status),
				Paid:     test.Paid,
				Entitled: test.Granted,
				Preview:  test.Preview,
			}.CanBeViewedBy(Actor{UserID: actorID, Administrator: test.Role == accesstest.Administrator})
			if allowed != test.Want {
				t.Fatalf("allowed=%v want=%v", allowed, test.Want)
			}
		})
	}
}

func TestCommentMutationPermissions(t *testing.T) {
	comment := Comment{AuthorID: "author"}
	course := Course{OwnerID: "owner"}
	for _, test := range []struct {
		name      string
		actor     Actor
		canEdit   bool
		canRemove bool
	}{
		{name: "author", actor: Actor{UserID: "author"}, canEdit: true, canRemove: true},
		{name: "student", actor: Actor{UserID: "student"}},
		{name: "admin", actor: Actor{UserID: "admin", Administrator: true}, canRemove: true},
		{name: "course owner", actor: Actor{UserID: "owner"}, canRemove: true},
		{name: "anonymous", actor: Actor{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := comment.CanBeEditedBy(test.actor); got != test.canEdit {
				t.Fatalf("edit=%v want=%v", got, test.canEdit)
			}
			if got := comment.CanBeRemovedBy(test.actor, course); got != test.canRemove {
				t.Fatalf("remove=%v want=%v", got, test.canRemove)
			}
		})
	}
}

func parentID(comment *Comment) *string {
	if comment == nil {
		return nil
	}
	return stringPointer(comment.ID)
}

func stringPointer(value string) *string     { return &value }
func timePointer(value time.Time) *time.Time { return &value }
