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
			course := access.Course{
				OwnerID:   owner,
				Published: test.Status == "published",
				Paid:      test.Paid,
				Entitled:  test.Granted,
				Preview:   test.Preview,
			}
			if got := access.CanView(subject, course); got != test.Want {
				t.Fatalf("CanView() = %v, want %v", got, test.Want)
			}
			if got := access.PurchaseRequired(subject, course); got != test.WantPurchase {
				t.Fatalf("PurchaseRequired() = %v, want %v", got, test.WantPurchase)
			}
		})
	}

	t.Run("empty subject, draft, empty owner", func(t *testing.T) {
		if got := access.CanView(access.Subject{}, access.Course{}); got {
			t.Fatal("CanView() = true, want false")
		}
	})
}

func TestCanStreamRequiresVerifiedEmailAfterViewAccess(t *testing.T) {
	t.Parallel()

	course := access.Course{OwnerID: "owner", Published: true}
	if access.CanStream(access.Subject{UserID: "student"}, course) {
		t.Fatal("CanStream() = true for an unverified viewer")
	}
	if !access.CanStream(access.Subject{UserID: "student", EmailVerified: true}, course) {
		t.Fatal("CanStream() = false for a verified viewer with view access")
	}
	if access.CanStream(access.Subject{UserID: "student", EmailVerified: true}, access.Course{OwnerID: "owner"}) {
		t.Fatal("CanStream() = true without view access")
	}
}

func TestPurchaseRequiredImpliesCannotView(t *testing.T) {
	t.Parallel()

	const owner = "owner"
	for _, test := range accesstest.TruthTable {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()

			subject, course := accessFacts(test, owner)
			if access.PurchaseRequired(subject, course) && access.CanView(subject, course) {
				t.Fatal("PurchaseRequired() = true and CanView() = true")
			}
		})
	}
}

func TestCanViewAndPurchaseRequiredUseTheSameCourseFacts(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name         string
		subject      access.Subject
		course       access.Course
		wantView     bool
		wantPurchase bool
	}{
		{name: "owner", subject: access.Subject{UserID: "owner"}, course: access.Course{OwnerID: "owner", Paid: true}, wantView: true},
		{name: "unpublished", subject: access.Subject{UserID: "stranger"}, course: access.Course{OwnerID: "owner", Paid: true}},
		{name: "free published", subject: access.Subject{UserID: "stranger"}, course: access.Course{OwnerID: "owner", Published: true}, wantView: true},
		{name: "paid published", subject: access.Subject{UserID: "stranger"}, course: access.Course{OwnerID: "owner", Published: true, Paid: true}, wantPurchase: true},
		{name: "entitled paid published", subject: access.Subject{UserID: "stranger"}, course: access.Course{OwnerID: "owner", Published: true, Paid: true, Entitled: true}, wantView: true},
		{name: "preview paid published", subject: access.Subject{UserID: "stranger"}, course: access.Course{OwnerID: "owner", Published: true, Paid: true, Preview: true}, wantView: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := access.CanView(test.subject, test.course); got != test.wantView {
				t.Fatalf("CanView() = %v, want %v", got, test.wantView)
			}
			if got := access.PurchaseRequired(test.subject, test.course); got != test.wantPurchase {
				t.Fatalf("PurchaseRequired() = %v, want %v", got, test.wantPurchase)
			}
		})
	}
}

func accessFacts(test accesstest.Case, owner string) (access.Subject, access.Course) {
	subject := access.Subject{UserID: owner}
	if test.Role == accesstest.Administrator {
		subject = access.Subject{UserID: "administrator", Administrator: true}
	} else if test.Role == accesstest.Stranger {
		subject = access.Subject{UserID: "stranger"}
	}
	return subject, access.Course{
		OwnerID:   owner,
		Published: test.Status == "published",
		Paid:      test.Paid,
		Entitled:  test.Granted,
		Preview:   test.Preview,
	}
}
