package access_test

import (
	"testing"

	"courses/internal/contracts/access"
)

func TestCanView(t *testing.T) {
	t.Parallel()

	const (
		owner    = "owner"
		stranger = "stranger"
	)

	for _, test := range []struct {
		name    string
		subject access.Subject
		course  access.Course
		want    bool
	}{
		{name: "admin draft", subject: access.Subject{UserID: stranger, Administrator: true}, course: access.Course{OwnerID: owner}, want: true},
		{name: "admin published", subject: access.Subject{UserID: stranger, Administrator: true}, course: access.Course{OwnerID: owner, Published: true}, want: true},
		{name: "admin archived (Published false)", subject: access.Subject{UserID: stranger, Administrator: true}, course: access.Course{OwnerID: owner}, want: true},
		{name: "owner draft", subject: access.Subject{UserID: owner}, course: access.Course{OwnerID: owner}, want: true},
		{name: "owner published", subject: access.Subject{UserID: owner}, course: access.Course{OwnerID: owner, Published: true}, want: true},
		{name: "owner archived (Published false)", subject: access.Subject{UserID: owner}, course: access.Course{OwnerID: owner}, want: true},
		{name: "stranger draft", subject: access.Subject{UserID: stranger}, course: access.Course{OwnerID: owner}, want: false},
		{name: "stranger published", subject: access.Subject{UserID: stranger}, course: access.Course{OwnerID: owner, Published: true}, want: true},
		{name: "stranger archived (Published false)", subject: access.Subject{UserID: stranger}, course: access.Course{OwnerID: owner}, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := access.CanView(test.subject, test.course); got != test.want {
				t.Fatalf("CanView() = %v, want %v", got, test.want)
			}
		})
	}
}
