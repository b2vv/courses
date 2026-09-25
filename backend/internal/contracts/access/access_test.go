package access_test

import (
	"testing"

	"courses/internal/contracts/access"
	"courses/internal/contracts/access/accesstest"
)

func TestCanView(t *testing.T) {
	t.Parallel()

	const owner = "owner"
	for _, test := range accesstest.TruthTable {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			subject := access.Subject{UserID: owner}
			if test.Role == accesstest.Administrator {
				subject = access.Subject{UserID: "administrator", Administrator: true}
			} else if test.Role == accesstest.Stranger {
				subject = access.Subject{UserID: "stranger"}
			}
			course := access.Course{OwnerID: owner, Published: test.Status == "published"}
			if got := access.CanView(subject, course); got != test.Want {
				t.Fatalf("CanView() = %v, want %v", got, test.Want)
			}
		})
	}

	t.Run("empty subject, draft, empty owner", func(t *testing.T) {
		if got := access.CanView(access.Subject{}, access.Course{}); got {
			t.Fatal("CanView() = true, want false")
		}
	})
}
