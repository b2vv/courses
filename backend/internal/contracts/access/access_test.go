package access_test

import (
	"strconv"
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

func TestMembershipPredicates(t *testing.T) {
	t.Parallel()

	const owner = "owner"
	roles := []struct {
		name    string
		subject access.Subject
	}{
		{name: "stranger", subject: access.Subject{UserID: "stranger"}},
		{name: "owner", subject: access.Subject{UserID: owner}},
		{name: "administrator", subject: access.Subject{UserID: "administrator", Administrator: true}},
	}
	memberships := []access.Membership{
		access.MembershipNone,
		access.MembershipActive,
		access.MembershipBlocked,
		access.MembershipRevoked,
	}

	for _, membership := range memberships {
		for _, role := range roles {
			for _, preview := range []bool{false, true} {
				for _, published := range []bool{false, true} {
					name := string(membership)
					if name == "" {
						name = "none"
					}
					t.Run(name+"/"+role.name+"/preview="+strconv.FormatBool(preview)+"/published="+strconv.FormatBool(published), func(t *testing.T) {
						t.Parallel()

						course := access.Course{OwnerID: owner, Published: published, Membership: membership, Preview: preview}
						// A preview lesson opens content only while the course is published; membership
						// (active or blocked) opens it regardless of published state (archived member).
						wantConsume := role.subject.Administrator || role.subject.UserID == owner || (published && preview) || membership == access.MembershipActive || membership == access.MembershipBlocked
						wantCommunicate := role.subject.Administrator || role.subject.UserID == owner || membership == access.MembershipActive

						if got := access.CanConsume(role.subject, course); got != wantConsume {
							t.Fatalf("CanConsume() = %v, want %v", got, wantConsume)
						}
						if got := access.CanCommunicate(role.subject, course); got != wantCommunicate {
							t.Fatalf("CanCommunicate() = %v, want %v", got, wantCommunicate)
						}
					})
				}
			}
		}
	}
}

// SA09/M10: an anonymous subject must not match a course whose OwnerID is empty. The
// `subject.UserID != ""` guard on the owner branch is the only thing between "" == "" and access;
// CanView pins it, these pin it for CanConsume and CanCommunicate.
func TestEmptySubjectDoesNotMatchEmptyOwner(t *testing.T) {
	t.Parallel()

	for _, membership := range []access.Membership{access.MembershipNone, access.MembershipRevoked} {
		for _, published := range []bool{false, true} {
			course := access.Course{Published: published, Membership: membership}
			if access.CanConsume(access.Subject{}, course) {
				t.Fatalf("CanConsume(empty subject, empty owner, membership=%q, published=%v) = true, want false", membership, published)
			}
			if access.CanCommunicate(access.Subject{}, course) {
				t.Fatalf("CanCommunicate(empty subject, empty owner, membership=%q, published=%v) = true, want false", membership, published)
			}
		}
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
