package domain_test

import (
	"errors"
	"testing"

	"courses/internal/media/domain"
)

func TestActorCanEditLesson(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		actor domain.Actor
		owner string
		want  error
	}{
		{name: "administrator edits any lesson", actor: domain.Actor{UserID: "admin", Administrator: true}, owner: "owner"},
		{name: "owner edits own lesson", actor: domain.Actor{UserID: "owner"}, owner: "owner"},
		{name: "non-owner is refused", actor: domain.Actor{UserID: "other"}, owner: "owner", want: domain.ErrCourseForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.actor.CanEditLesson(test.owner); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}
