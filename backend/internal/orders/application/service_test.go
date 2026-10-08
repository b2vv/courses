package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"courses/internal/orders/application"
	"courses/internal/orders/domain"
	"courses/internal/orders/ports"
)

const (
	userID   = "40000000-0000-4000-8000-000000000001"
	adminID  = "40000000-0000-4000-8000-000000000002"
	courseID = "50000000-0000-4000-8000-000000000001"
	orderID  = "60000000-0000-4000-8000-000000000001"
)

var fixedNow = time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)

func TestStartPurchaseSnapshotsPublishedPaidCoursePrice(t *testing.T) {
	store := newFakeStore(t)
	store.course = ports.CourseFacts{Published: true, Price: money(t, "50000")}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

	order, err := service.StartPurchase(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID, "purchase-key")
	if err != nil {
		t.Fatalf("StartPurchase: %v", err)
	}
	if order.ID() != orderID || order.UserID() != userID || order.CourseID() != courseID || order.Status() != domain.OrderPending || order.Amount().MinorUnits() != 50000 {
		t.Fatalf("unexpected order: id=%q user=%q course=%q status=%q amount=%d", order.ID(), order.UserID(), order.CourseID(), order.Status(), order.Amount().MinorUnits())
	}
}

func TestStartPurchaseIsIdempotentAndRejectsOwnedOrUnpurchasableCourses(t *testing.T) {
	existing := pendingOrder(t)
	paidExisting, err := existing.MarkPaid(fixedNow.Add(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	failedExisting, err := existing.MarkFailed()
	if err != nil {
		t.Fatal(err)
	}
	otherCourseOrder, err := domain.NewOrder(orderID, userID, "50000000-0000-4000-8000-000000000002", *money(t, "50000"), "purchase-key", fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name       string
		configure  func(*fakeStore)
		want       error
		wantOrder  string
		wantCreate int
	}{
		{name: "same key pending returns existing order", configure: func(store *fakeStore) { store.idempotentOrder = &existing }, wantOrder: orderID},
		{name: "same key paid without membership returns paid order", configure: func(store *fakeStore) { store.idempotentOrder = &paidExisting }, wantOrder: orderID},
		// The production replay of a paid key: the owner check answers first (Pay12
		// accepts already_owned here rather than the order itself).
		{name: "same key after a completed purchase is already owned", configure: func(store *fakeStore) {
			store.idempotentOrder = &paidExisting
			store.membership = membership(t, domain.GrantAdmin, domain.GrantActive)
		}, want: domain.ErrCourseAlreadyOwned},
		{name: "same key failed returns failed order", configure: func(store *fakeStore) { store.idempotentOrder = &failedExisting }, wantOrder: orderID},
		// The same key raced for another course: the insert hands back the other
		// course's order, which must never be answered as this course's purchase.
		{name: "same key raced for another course is refused", configure: func(store *fakeStore) { store.createReturns = &otherCourseOrder }, want: domain.ErrOrderInvalid, wantCreate: 1},
		{name: "active grant blocks purchase", configure: func(store *fakeStore) { store.membership = membership(t, domain.GrantAdmin, domain.GrantActive) }, want: domain.ErrCourseAlreadyOwned},
		{name: "revoked membership blocks purchase", configure: func(store *fakeStore) { store.membership = membership(t, domain.GrantAdmin, domain.GrantRevoked) }, want: domain.ErrCourseAlreadyOwned},
		{name: "active self membership permits purchase", configure: func(store *fakeStore) { store.membership = membership(t, domain.GrantSelf, domain.GrantActive) }, wantCreate: 1},
		{name: "free course", configure: func(store *fakeStore) { store.course.Price = nil }, want: domain.ErrCourseNotPurchasable},
		{name: "unpublished course", configure: func(store *fakeStore) { store.course.Published = false }, want: domain.ErrCourseNotPurchasable},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			test.configure(store)
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
			order, err := service.StartPurchase(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID, "purchase-key")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if test.wantOrder != "" && order.ID() != test.wantOrder {
				t.Fatalf("order id=%q want=%q", order.ID(), test.wantOrder)
			}
			if store.createCalls != test.wantCreate {
				t.Fatalf("create calls=%d want=%d", store.createCalls, test.wantCreate)
			}
		})
	}
}

func TestPurchaseStatusIsScopedToActor(t *testing.T) {
	store := newFakeStore(t)
	existing := pendingOrder(t)
	store.order = &existing
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

	got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
	if err != nil {
		t.Fatalf("PurchaseStatus: %v", err)
	}
	if got.ID() != orderID || store.getUserID != userID {
		t.Fatalf("order=%q queried user=%q", got.ID(), store.getUserID)
	}
}

func TestPurchaseStatusChecksProviderAndAppliesApprovedOnce(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	provider := &fakePaymentProvider{status: ports.PaymentStatus{Event: domain.PaymentEvent{Provider: "provider", EventID: "approved-1", OrderID: orderID, Type: domain.PaymentSucceeded, Amount: *money(t, "50000"), ReceivedAt: fixedNow}}}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, provider)
	if _, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID); err != nil {
		t.Fatal(err)
	}
	if provider.calls != 1 || store.appliedEvent.EventID != "approved-1" {
		t.Fatalf("provider calls=%d event=%q", provider.calls, store.appliedEvent.EventID)
	}
}

func TestPurchaseStatusIgnoresProviderFailure(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	provider := &fakePaymentProvider{err: errors.New("provider unavailable")}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, provider)
	got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
	if err != nil || got.ID() != orderID || got.Status() != domain.OrderPending {
		t.Fatalf("order=%+v error=%v", got, err)
	}
}

func TestCheckoutClaimsAnOrderOnce(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	provider := &fakePaymentProvider{}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, provider)

	if _, err := service.Checkout(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, orderID, ports.Buyer{}, ports.CheckoutURLs{}); err != nil {
		t.Fatalf("first checkout: %v", err)
	}
	if _, err := service.Checkout(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, orderID, ports.Buyer{}, ports.CheckoutURLs{}); !errors.Is(err, domain.ErrCheckoutAlreadyOpened) {
		t.Fatalf("second checkout error=%v want=%v", err, domain.ErrCheckoutAlreadyOpened)
	}
	if provider.checkoutCalls != 1 {
		t.Fatalf("provider checkout calls=%d want=1", provider.checkoutCalls)
	}
}

func TestPurchaseStatusThrottleIsSafeUnderConcurrentPolling(t *testing.T) {
	store := newFakeStore(t)
	order := pendingOrder(t)
	store.order = &order
	concurrentStore := &concurrentStatusStore{fakeStore: store}
	provider := &concurrentPaymentProvider{}
	service := application.New(concurrentStore, concurrentStore, concurrentStore, func() time.Time { return fixedNow }, func() string { return orderID }, provider)

	var wait sync.WaitGroup
	for range 32 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			got, err := service.PurchaseStatus(context.Background(), application.Actor{UserID: userID}, orderID)
			if err != nil || got.Status() != domain.OrderPending {
				t.Errorf("order=%q error=%v", got.Status(), err)
			}
		}()
	}
	wait.Wait()
}

func TestApplyPaymentEventDelegatesTheAtomicOperation(t *testing.T) {
	store := newFakeStore(t)
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, nil)
	event := domain.PaymentEvent{Provider: "provider", EventID: "event-1", OrderID: orderID, Type: domain.PaymentSucceeded, Amount: *money(t, "50000"), ReceivedAt: fixedNow}

	if err := service.ApplyPaymentEvent(context.Background(), event); err != nil {
		t.Fatalf("ApplyPaymentEvent: %v", err)
	}
	if store.appliedEvent.EventID != event.EventID {
		t.Fatalf("event=%q want=%q", store.appliedEvent.EventID, event.EventID)
	}
}

func TestGrantCourseRequiresAdministratorAndWritesAudit(t *testing.T) {
	for _, test := range []struct {
		name      string
		admin     bool
		targetID  string
		want      error
		wantCalls int
	}{
		{name: "administrator grants another user", admin: true, targetID: userID, wantCalls: 1},
		{name: "administrator cannot grant themselves", admin: true, targetID: adminID, want: domain.ErrGrantSelfForbidden},
		{name: "student is forbidden", targetID: userID, want: domain.ErrGrantForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, nil)
			grant, err := service.GrantCourse(context.Background(), application.Actor{UserID: adminID, Administrator: test.admin}, test.targetID, courseID, "support decision")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if store.createMembershipCalls != test.wantCalls {
				t.Fatalf("grant calls=%d want=%d", store.createMembershipCalls, test.wantCalls)
			}
			if test.want == nil && (grant.Source() != domain.GrantAdmin || grant.GrantedBy() == nil || *grant.GrantedBy() != adminID || store.membershipReason != "support decision") {
				t.Fatalf("unexpected grant or audit command: %#v", grant)
			}
		})
	}
}

func TestGrantCourseReinstatesOnlyRevokedMembership(t *testing.T) {
	for _, test := range []struct {
		name       string
		membership *domain.Grant
		want       error
		wantUpdate int
	}{
		{name: "revoked is reinstated", membership: membership(t, domain.GrantSelf, domain.GrantRevoked), wantUpdate: 1},
		{name: "active remains owned", membership: membership(t, domain.GrantSelf, domain.GrantActive), want: domain.ErrCourseAlreadyOwned},
		{name: "blocked remains owned", membership: membership(t, domain.GrantSelf, domain.GrantBlocked), want: domain.ErrCourseAlreadyOwned},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			store.membership = test.membership
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
			grant, err := service.GrantCourse(context.Background(), application.Actor{UserID: adminID, Administrator: true}, userID, courseID, "support decision")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
			if test.want == nil && (grant.State() != domain.GrantActive || grant.Source() != domain.GrantAdmin || store.updateMembershipCalls != test.wantUpdate || store.membershipAction != "grant") {
				t.Fatalf("grant=%#v update=%d action=%q", grant, store.updateMembershipCalls, store.membershipAction)
			}
		})
	}
}

// Lo18 (G8): an administrator is never a learner. Enrol, StartPurchase and GrantCourse-to-an-admin
// all refuse with the one named code before any membership or order is written.
func TestAdministratorIsNeverALearner(t *testing.T) {
	admin := application.Actor{UserID: adminID, Administrator: true, EmailVerified: true}
	t.Run("admin cannot enrol", func(t *testing.T) {
		store := newFakeStore(t)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Enroll(context.Background(), admin, courseID); !errors.Is(err, domain.ErrAdminNotLearner) {
			t.Fatalf("enrol error=%v want=%v", err, domain.ErrAdminNotLearner)
		}
		if store.createMembershipCalls != 0 {
			t.Fatalf("enrol wrote membership: calls=%d want=0", store.createMembershipCalls)
		}
	})
	t.Run("admin cannot start a purchase", func(t *testing.T) {
		store := newFakeStore(t)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.StartPurchase(context.Background(), admin, courseID, "idem-1"); !errors.Is(err, domain.ErrAdminNotLearner) {
			t.Fatalf("purchase error=%v want=%v", err, domain.ErrAdminNotLearner)
		}
		if store.createCalls != 0 {
			t.Fatalf("purchase wrote order: calls=%d want=0", store.createCalls)
		}
	})
	t.Run("an administrator cannot be granted a course", func(t *testing.T) {
		store := newFakeStore(t)
		store.recipientIsAdmin = true
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.GrantCourse(context.Background(), admin, userID, courseID, "support decision"); !errors.Is(err, domain.ErrAdminNotLearner) {
			t.Fatalf("grant error=%v want=%v", err, domain.ErrAdminNotLearner)
		}
		if store.createMembershipCalls != 0 {
			t.Fatalf("grant wrote membership: calls=%d want=0", store.createMembershipCalls)
		}
		// Role is checked before the reason: an admin recipient with an empty reason still fails
		// as admin_not_learner, not grant_invalid.
		if _, err := service.GrantCourse(context.Background(), admin, userID, courseID, ""); !errors.Is(err, domain.ErrAdminNotLearner) {
			t.Fatalf("grant with empty reason error=%v want=%v", err, domain.ErrAdminNotLearner)
		}
	})
}

func TestLearnersScopesAuthorizesAndClampsLimit(t *testing.T) {
	instructor := application.Actor{UserID: userID}
	admin := application.Actor{UserID: adminID, Administrator: true}

	t.Run("admin scope requires an administrator", func(t *testing.T) {
		store := newFakeStore(t)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Learners(context.Background(), application.Actor{UserID: userID}, true, "", "", "", "", 0); !errors.Is(err, domain.ErrEnrollmentForbidden) {
			t.Fatalf("non-admin admin scope err=%v", err)
		}
	})

	t.Run("admin scope passes an administrator filter and clamps the limit", func(t *testing.T) {
		store := newFakeStore(t)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Learners(context.Background(), admin, true, "", "blocked", "ann", "", 500); err != nil {
			t.Fatalf("admin learners err=%v", err)
		}
		if !store.learnerFilter.Administrator || store.learnerFilter.ViewerID != "" || store.learnerFilter.State != "blocked" || store.learnerFilter.Query != "ann" || store.learnerLimit != 100 {
			t.Fatalf("admin filter=%#v limit=%d", store.learnerFilter, store.learnerLimit)
		}
	})

	t.Run("instructor scope views own courses and defaults the limit", func(t *testing.T) {
		store := newFakeStore(t)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Learners(context.Background(), instructor, false, "", "", "", "", 0); err != nil {
			t.Fatalf("instructor learners err=%v", err)
		}
		if store.learnerFilter.Administrator || store.learnerFilter.ViewerID != userID || store.learnerLimit != 20 {
			t.Fatalf("instructor filter=%#v limit=%d", store.learnerFilter, store.learnerLimit)
		}
	})

	t.Run("instructor filtering by an owned course passes it through", func(t *testing.T) {
		store := newFakeStore(t)
		store.enrollment = ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: userID}
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Learners(context.Background(), instructor, false, courseID, "", "", "", 20); err != nil {
			t.Fatalf("own course err=%v", err)
		}
		if store.learnerFilter.CourseID != courseID {
			t.Fatalf("course filter=%q", store.learnerFilter.CourseID)
		}
	})

	t.Run("instructor filtering by another course is refused", func(t *testing.T) {
		store := newFakeStore(t)
		store.enrollment = ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: "someone-else"}
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Learners(context.Background(), instructor, false, courseID, "", "", "", 20); !errors.Is(err, domain.ErrLearnerCourseForbidden) {
			t.Fatalf("foreign course err=%v", err)
		}
	})

	t.Run("instructor filtering by an unknown course is refused, not revealed", func(t *testing.T) {
		store := newFakeStore(t)
		store.enrollmentErr = domain.ErrCourseNotEnrollable
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.Learners(context.Background(), instructor, false, courseID, "", "", "", 20); !errors.Is(err, domain.ErrLearnerCourseForbidden) {
			t.Fatalf("unknown course err=%v", err)
		}
	})
}

// Lo12-Lo17: blocking from the queue resolves the author and course, then runs the same membership
// transition as the roster — one audit, an origin reason — and refuses unknown items, strangers and
// non-members by name.
func TestBlockAuthorFromQueueReusesTheMembershipTransition(t *testing.T) {
	admin := application.Actor{UserID: adminID, Administrator: true}

	t.Run("blocks the comment author with a moderation-origin reason", func(t *testing.T) {
		store := newFakeStore(t)
		store.authorUserID, store.authorCourseID = userID, courseID
		store.membership = membership(t, domain.GrantSelf, domain.GrantActive)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		grant, err := service.BlockAuthor(context.Background(), admin, "comment", "comment-1")
		if err != nil || grant.State() != domain.GrantBlocked {
			t.Fatalf("block author grant=%#v err=%v", grant, err)
		}
		if store.authorKind != "comment" || store.authorID != "comment-1" {
			t.Fatalf("AuthorOf called with kind=%q id=%q", store.authorKind, store.authorID)
		}
		if store.membershipAction != "block" || store.membershipReason != "moderation_queue:comment:comment-1" {
			t.Fatalf("audit action=%q reason=%q", store.membershipAction, store.membershipReason)
		}
	})

	t.Run("unblocks the review author", func(t *testing.T) {
		store := newFakeStore(t)
		store.authorUserID, store.authorCourseID = userID, courseID
		store.membership = membership(t, domain.GrantSelf, domain.GrantBlocked)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		grant, err := service.UnblockAuthor(context.Background(), admin, "review", "review-9")
		if err != nil || grant.State() != domain.GrantActive {
			t.Fatalf("unblock author grant=%#v err=%v", grant, err)
		}
		if store.membershipReason != "moderation_queue:review:review-9" {
			t.Fatalf("audit reason=%q", store.membershipReason)
		}
	})

	t.Run("an unknown item is refused by name", func(t *testing.T) {
		store := newFakeStore(t)
		store.authorErr = domain.ErrCommentNotFound
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.BlockAuthor(context.Background(), admin, "comment", "missing"); !errors.Is(err, domain.ErrCommentNotFound) {
			t.Fatalf("unknown comment err=%v", err)
		}
	})

	t.Run("a stranger to the course is forbidden (G6)", func(t *testing.T) {
		store := newFakeStore(t)
		store.authorUserID, store.authorCourseID = userID, courseID
		store.enrollment = ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: "another-instructor"}
		store.membership = membership(t, domain.GrantSelf, domain.GrantActive)
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.BlockAuthor(context.Background(), application.Actor{UserID: "stranger"}, "comment", "c"); !errors.Is(err, domain.ErrEnrollmentForbidden) {
			t.Fatalf("stranger block err=%v", err)
		}
	})

	t.Run("an author who is not a member cannot be blocked (G8 construction)", func(t *testing.T) {
		store := newFakeStore(t)
		store.authorUserID, store.authorCourseID = userID, courseID
		store.membership = nil
		service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
		if _, err := service.BlockAuthor(context.Background(), admin, "comment", "c"); !errors.Is(err, domain.ErrGrantTransitionInvalid) {
			t.Fatalf("non-member author err=%v", err)
		}
	})
}

func TestEnrollmentWritesTransitionsAndPreservesNoOps(t *testing.T) {
	for _, test := range []struct {
		name       string
		membership *domain.Grant
		call       func(*application.Service) (domain.Grant, error)
		wantState  domain.GrantState
		wantAction string
		wantCreate int
		wantUpdate int
		wantErr    error
	}{
		{name: "enroll creates self membership", call: func(s *application.Service) (domain.Grant, error) {
			return s.Enroll(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID)
		}, wantState: domain.GrantActive, wantAction: "enroll", wantCreate: 1},
		{name: "enroll active is no op", membership: membership(t, domain.GrantSelf, domain.GrantActive), call: func(s *application.Service) (domain.Grant, error) {
			return s.Enroll(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID)
		}, wantState: domain.GrantActive},
		{name: "enroll blocked is no op", membership: membership(t, domain.GrantSelf, domain.GrantBlocked), call: func(s *application.Service) (domain.Grant, error) {
			return s.Enroll(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID)
		}, wantState: domain.GrantBlocked},
		{name: "enroll reinstates revoked self membership", membership: membership(t, domain.GrantAdmin, domain.GrantRevoked), call: func(s *application.Service) (domain.Grant, error) {
			return s.Enroll(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID)
		}, wantState: domain.GrantActive, wantAction: "enroll", wantUpdate: 1},
		{name: "owner blocks active", membership: membership(t, domain.GrantSelf, domain.GrantActive), call: func(s *application.Service) (domain.Grant, error) {
			return s.Block(context.Background(), application.Actor{UserID: adminID}, userID, courseID)
		}, wantState: domain.GrantBlocked, wantAction: "block", wantUpdate: 1},
		{name: "blocked block is no op", membership: membership(t, domain.GrantSelf, domain.GrantBlocked), call: func(s *application.Service) (domain.Grant, error) {
			return s.Block(context.Background(), application.Actor{UserID: adminID}, userID, courseID)
		}, wantState: domain.GrantBlocked},
		{name: "owner unblocks", membership: membership(t, domain.GrantSelf, domain.GrantBlocked), call: func(s *application.Service) (domain.Grant, error) {
			return s.Unblock(context.Background(), application.Actor{UserID: adminID}, userID, courseID)
		}, wantState: domain.GrantActive, wantAction: "unblock", wantUpdate: 1},
		{name: "admin revokes", membership: membership(t, domain.GrantSelf, domain.GrantActive), call: func(s *application.Service) (domain.Grant, error) {
			return s.Revoke(context.Background(), application.Actor{UserID: adminID, Administrator: true}, userID, courseID)
		}, wantState: domain.GrantRevoked, wantAction: "revoke", wantUpdate: 1},
		{name: "revoked revoke is no op", membership: membership(t, domain.GrantSelf, domain.GrantRevoked), call: func(s *application.Service) (domain.Grant, error) {
			return s.Revoke(context.Background(), application.Actor{UserID: adminID, Administrator: true}, userID, courseID)
		}, wantState: domain.GrantRevoked},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			store.membership = test.membership
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return "70000000-0000-4000-8000-000000000001" }, nil)
			grant, err := test.call(service)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v want=%v", err, test.wantErr)
			}
			if test.wantErr == nil && grant.State() != test.wantState {
				t.Fatalf("state=%q want=%q", grant.State(), test.wantState)
			}
			if store.membershipAction != test.wantAction || store.createMembershipCalls != test.wantCreate || store.updateMembershipCalls != test.wantUpdate {
				t.Fatalf("action=%q create=%d update=%d", store.membershipAction, store.createMembershipCalls, store.updateMembershipCalls)
			}
		})
	}
}

func TestEnrollmentAuthorizationAndCourseEligibility(t *testing.T) {
	for _, test := range []struct {
		name       string
		enrollment ports.EnrollmentFacts
		membership *domain.Grant
		call       func(*application.Service) error
		want       error
	}{
		{name: "unverified cannot enroll", enrollment: ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: adminID}, call: func(s *application.Service) error {
			_, err := s.Enroll(context.Background(), application.Actor{UserID: userID}, courseID)
			return err
		}, want: domain.ErrEmailNotVerified},
		{name: "draft cannot enroll", enrollment: ports.EnrollmentFacts{Status: "draft", Free: true, InstructorID: adminID}, call: func(s *application.Service) error {
			_, err := s.Enroll(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID)
			return err
		}, want: domain.ErrCourseNotEnrollable},
		{name: "paid cannot enroll", enrollment: ports.EnrollmentFacts{Status: "published", InstructorID: adminID}, call: func(s *application.Service) error {
			_, err := s.Enroll(context.Background(), application.Actor{UserID: userID, EmailVerified: true}, courseID)
			return err
		}, want: domain.ErrCourseNotEnrollable},
		{name: "stranger cannot block", membership: membership(t, domain.GrantSelf, domain.GrantActive), call: func(s *application.Service) error {
			_, err := s.Block(context.Background(), application.Actor{UserID: "other"}, userID, courseID)
			return err
		}, want: domain.ErrEnrollmentForbidden},
		{name: "owner cannot revoke", membership: membership(t, domain.GrantSelf, domain.GrantActive), call: func(s *application.Service) error {
			_, err := s.Revoke(context.Background(), application.Actor{UserID: adminID}, userID, courseID)
			return err
		}, want: domain.ErrEnrollmentForbidden},
		{name: "block missing membership is invalid transition", call: func(s *application.Service) error {
			_, err := s.Block(context.Background(), application.Actor{UserID: adminID}, userID, courseID)
			return err
		}, want: domain.ErrGrantTransitionInvalid},
		{name: "revoke missing membership is invalid grant", call: func(s *application.Service) error {
			_, err := s.Revoke(context.Background(), application.Actor{UserID: adminID, Administrator: true}, userID, courseID)
			return err
		}, want: domain.ErrGrantInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			store.enrollment = test.enrollment
			store.membership = test.membership
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
			if err := test.call(service); !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
		})
	}
}

func TestMembershipStateAllowsSelfAdministratorAndInstructor(t *testing.T) {
	store := newFakeStore(t)
	store.membership = membership(t, domain.GrantSelf, domain.GrantBlocked)
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
	for _, actor := range []application.Actor{{UserID: userID}, {UserID: adminID}, {UserID: "other", Administrator: true}} {
		state, found, err := service.MembershipState(context.Background(), actor, userID, courseID)
		if err != nil || !found || state != domain.GrantBlocked {
			t.Fatalf("actor=%#v state=%q found=%v error=%v", actor, state, found, err)
		}
	}
	if _, _, err := service.MembershipState(context.Background(), application.Actor{UserID: "other"}, userID, courseID); !errors.Is(err, domain.ErrEnrollmentForbidden) {
		t.Fatalf("stranger error=%v", err)
	}
}

func TestMyMembershipsDelegatesForActor(t *testing.T) {
	store := newFakeStore(t)
	store.memberships = []ports.MembershipCard{{CourseID: courseID, Slug: "course", Title: "Course", State: domain.GrantActive}}
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

	cards, err := service.MyMemberships(context.Background(), application.Actor{UserID: userID})
	if err != nil {
		t.Fatalf("MyMemberships: %v", err)
	}
	if store.membershipsUserID != userID || len(cards) != 1 || cards[0].CourseID != courseID {
		t.Fatalf("user=%q cards=%#v", store.membershipsUserID, cards)
	}
}

func TestRosterAllowsInstructorAndAdministratorAndClampsLimit(t *testing.T) {
	page := ports.RosterPage{Entries: []ports.RosterEntry{{UserID: userID, DisplayName: "Student", State: domain.GrantActive, ProgressPercent: 50}}, NextCursor: "next"}
	for _, test := range []struct {
		name      string
		actor     application.Actor
		limit     int
		wantLimit int
		wantErr   error
	}{
		{name: "owner gets default limit", actor: application.Actor{UserID: adminID}, wantLimit: 20},
		{name: "administrator gets minimum limit", actor: application.Actor{UserID: "other", Administrator: true}, limit: -1, wantLimit: 20},
		{name: "limit one", actor: application.Actor{UserID: adminID}, limit: 1, wantLimit: 1},
		{name: "limit is capped", actor: application.Actor{UserID: adminID}, limit: 101, wantLimit: 100},
		{name: "stranger is forbidden", actor: application.Actor{UserID: "other"}, limit: 10, wantErr: domain.ErrEnrollmentForbidden},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := newFakeStore(t)
			store.rosterPage = page
			service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)
			got, err := service.Roster(context.Background(), test.actor, courseID, "cursor", test.limit)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v want=%v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}
			if store.rosterCourseID != courseID || store.rosterCursor != "cursor" || store.rosterLimit != test.wantLimit {
				t.Fatalf("repository arguments course=%q cursor=%q limit=%d", store.rosterCourseID, store.rosterCursor, store.rosterLimit)
			}
			if got.NextCursor != page.NextCursor || len(got.Entries) != 1 || got.Entries[0].DisplayName != "Student" {
				t.Fatalf("page=%#v", got)
			}
		})
	}
}

func TestGrandfatherOnPriceSetUsesOneRepositoryOperation(t *testing.T) {
	store := newFakeStore(t)
	service := application.New(store, store, store, func() time.Time { return fixedNow }, func() string { return orderID }, nil)

	updatedAt := fixedNow.Add(time.Microsecond)
	err := service.GrandfatherOnPriceSet(context.Background(), courseID, fixedNow, 50000, "UAH", updatedAt)
	if err != nil {
		t.Fatalf("GrandfatherOnPriceSet: %v", err)
	}
	if store.grandfatherCourse != courseID || !store.grandfatherExpected.Equal(fixedNow) || !store.grandfatherUpdated.Equal(updatedAt) || store.grandfatherPrice.MinorUnits() != 50000 {
		t.Fatalf("course=%q expected=%v updated=%v price=%d", store.grandfatherCourse, store.grandfatherExpected, store.grandfatherUpdated, store.grandfatherPrice.MinorUnits())
	}
}

func newFakeStore(t *testing.T) *fakeStore {
	t.Helper()
	return &fakeStore{course: ports.CourseFacts{Published: true, Price: money(t, "50000")}}
}

type fakeStore struct {
	course                ports.CourseFacts
	enrollment            ports.EnrollmentFacts
	idempotentOrder       *domain.Order
	createReturns         *domain.Order
	membership            *domain.Grant
	createCalls           int
	checkoutCalls         int
	order                 *domain.Order
	getUserID             string
	appliedEvent          domain.PaymentEvent
	applyErr              error
	createMembershipCalls int
	updateMembershipCalls int
	membershipAction      string
	membershipReason      string
	grandfatherCourse     string
	grandfatherExpected   time.Time
	grandfatherUpdated    time.Time
	grandfatherPrice      domain.Money
	rosterPage            ports.RosterPage
	rosterCourseID        string
	rosterCursor          string
	rosterLimit           int
	memberships           []ports.MembershipCard
	membershipsUserID     string
	recipientIsAdmin      bool
	learnerPage           ports.LearnerPage
	learnerErr            error
	learnerFilter         ports.LearnerFilter
	learnerCursor         string
	learnerLimit          int
	enrollmentErr         error
	authorUserID          string
	authorCourseID        string
	authorErr             error
	authorKind            string
	authorID              string
}

type fakePaymentProvider struct {
	status        ports.PaymentStatus
	err           error
	calls         int
	checkoutCalls int
}

type concurrentStatusStore struct {
	*fakeStore
	mu sync.Mutex
}

func (store *concurrentStatusStore) Get(_ context.Context, _ string, _ string) (domain.Order, error) {
	store.mu.Lock()
	defer store.mu.Unlock()
	if store.order == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.order, nil
}

type concurrentPaymentProvider struct {
	mu sync.Mutex
}

func (provider *concurrentPaymentProvider) Checkout(context.Context, domain.Order, ports.Buyer, ports.CheckoutURLs) (ports.CheckoutForm, error) {
	return ports.CheckoutForm{}, nil
}

func (provider *concurrentPaymentProvider) OrderStatus(context.Context, string) (ports.PaymentStatus, error) {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return ports.PaymentStatus{}, nil
}

func (provider *fakePaymentProvider) Checkout(context.Context, domain.Order, ports.Buyer, ports.CheckoutURLs) (ports.CheckoutForm, error) {
	provider.checkoutCalls++
	return ports.CheckoutForm{}, nil
}
func (provider *fakePaymentProvider) OrderStatus(context.Context, string) (ports.PaymentStatus, error) {
	provider.calls++
	return provider.status, provider.err
}

func (store *fakeStore) CourseForPurchase(context.Context, string) (ports.CourseFacts, error) {
	return store.course, nil
}

func (store *fakeStore) CourseForEnrollment(context.Context, string) (ports.EnrollmentFacts, error) {
	if store.enrollmentErr != nil {
		return ports.EnrollmentFacts{}, store.enrollmentErr
	}
	if store.enrollment.Status == "" {
		return ports.EnrollmentFacts{Status: "published", Free: true, InstructorID: adminID}, nil
	}
	return store.enrollment, nil
}

func (store *fakeStore) FindByIdempotencyKey(context.Context, string, string) (domain.Order, error) {
	if store.idempotentOrder == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.idempotentOrder, nil
}

func (store *fakeStore) Create(_ context.Context, order domain.Order) (domain.Order, error) {
	store.createCalls++
	// A unique (user, key) conflict makes the repository hand back the row that won
	// the race; createReturns models that row.
	if store.createReturns != nil {
		return *store.createReturns, nil
	}
	return order, nil
}

func (store *fakeStore) Get(_ context.Context, userID, _ string) (domain.Order, error) {
	store.getUserID = userID
	if store.order == nil {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	return *store.order, nil
}

func (store *fakeStore) MarkCheckoutOpened(_ context.Context, userID, orderID string) (domain.Order, error) {
	store.getUserID = userID
	if store.order == nil || store.order.ID() != orderID {
		return domain.Order{}, domain.ErrOrderNotFound
	}
	if store.checkoutCalls > 0 {
		return domain.Order{}, domain.ErrCheckoutAlreadyOpened
	}
	store.checkoutCalls++
	return *store.order, nil
}

func (store *fakeStore) Membership(context.Context, string, string) (*domain.Grant, error) {
	return store.membership, nil
}

func (store *fakeStore) IsAdministrator(context.Context, string) (bool, error) {
	return store.recipientIsAdmin, nil
}

func (store *fakeStore) Learners(_ context.Context, filter ports.LearnerFilter, cursor string, limit int) (ports.LearnerPage, error) {
	store.learnerFilter = filter
	store.learnerCursor = cursor
	store.learnerLimit = limit
	return store.learnerPage, store.learnerErr
}

func (store *fakeStore) AuthorOf(_ context.Context, kind, id string) (string, string, error) {
	store.authorKind = kind
	store.authorID = id
	return store.authorUserID, store.authorCourseID, store.authorErr
}

func (store *fakeStore) MyMemberships(_ context.Context, userID string) ([]ports.MembershipCard, error) {
	store.membershipsUserID = userID
	return store.memberships, nil
}

func (store *fakeStore) Roster(_ context.Context, courseID, cursor string, limit int) (ports.RosterPage, error) {
	store.rosterCourseID = courseID
	store.rosterCursor = cursor
	store.rosterLimit = limit
	return store.rosterPage, nil
}

func membership(t *testing.T, source domain.GrantSource, state domain.GrantState) *domain.Grant {
	t.Helper()
	var order *string
	if source == domain.GrantPurchase {
		value := orderID
		order = &value
	}
	grant, err := domain.NewGrant("grant-1", userID, courseID, source, order, state, fixedNow, nil)
	if err != nil {
		t.Fatal(err)
	}
	return &grant
}

func (store *fakeStore) CreateMembership(_ context.Context, grant domain.Grant, action, reason string, _ *string) (domain.Grant, error) {
	store.createMembershipCalls++
	store.membershipAction = action
	store.membershipReason = reason
	store.membership = &grant
	return grant, nil
}

func (store *fakeStore) UpdateMembership(_ context.Context, grant domain.Grant, _ domain.GrantState, action, reason string, _ *string) (domain.Grant, error) {
	store.updateMembershipCalls++
	store.membershipAction = action
	store.membershipReason = reason
	store.membership = &grant
	return grant, nil
}

func (store *fakeStore) GrandfatherOnPriceSet(_ context.Context, courseID string, expected time.Time, price domain.Money, updated time.Time) error {
	store.grandfatherCourse = courseID
	store.grandfatherExpected = expected
	store.grandfatherPrice = price
	store.grandfatherUpdated = updated
	return nil
}

func (store *fakeStore) Apply(_ context.Context, event domain.PaymentEvent, _ func() string) error {
	store.appliedEvent = event
	return store.applyErr
}

func money(t *testing.T, amount string) *domain.Money {
	t.Helper()
	value, err := domain.NewMoney(amount, "UAH")
	if err != nil {
		t.Fatal(err)
	}
	return &value
}

func pendingOrder(t *testing.T) domain.Order {
	t.Helper()
	order, err := domain.NewOrder(orderID, userID, courseID, *money(t, "50000"), "purchase-key", fixedNow)
	if err != nil {
		t.Fatal(err)
	}
	return order
}
