package domain

import (
	"errors"
	"testing"
	"time"
)

func TestNewGrantEnforcesSourceOrderInvariantAndKnownState(t *testing.T) {
	orderID := "order-1"
	emptyOrderID := ""
	grantedAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		source  GrantSource
		state   GrantState
		orderID *string
		wantErr error
	}{
		{name: "purchase has order", source: GrantPurchase, state: GrantActive, orderID: &orderID},
		{name: "admin has no order", source: GrantAdmin, state: GrantActive},
		{name: "grandfather has no order", source: GrantGrandfather, state: GrantRevoked},
		{name: "self has no order", source: GrantSelf, state: GrantBlocked},
		{name: "purchase missing order", source: GrantPurchase, state: GrantActive, wantErr: ErrGrantInvalid},
		{name: "purchase empty order", source: GrantPurchase, state: GrantActive, orderID: &emptyOrderID, wantErr: ErrGrantInvalid},
		{name: "admin unexpectedly has order", source: GrantAdmin, state: GrantActive, orderID: &orderID, wantErr: ErrGrantInvalid},
		{name: "admin has present empty order", source: GrantAdmin, state: GrantActive, orderID: &emptyOrderID, wantErr: ErrGrantInvalid},
		{name: "grandfather unexpectedly has order", source: GrantGrandfather, state: GrantActive, orderID: &orderID, wantErr: ErrGrantInvalid},
		{name: "grandfather has present empty order", source: GrantGrandfather, state: GrantActive, orderID: &emptyOrderID, wantErr: ErrGrantInvalid},
		{name: "unknown source", source: GrantSource("gift"), state: GrantActive, wantErr: ErrGrantInvalid},
		{name: "unknown state", source: GrantAdmin, state: GrantState("expired"), wantErr: ErrGrantInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewGrant("grant-1", "user-1", "course-1", test.source, test.orderID, test.state, grantedAt, nil)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v want=%v", err, test.wantErr)
			}
		})
	}
}

func TestGrantTransitionsFollowMembershipStateMachine(t *testing.T) {
	grantedAt := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	active, err := NewGrant("grant-1", "user-1", "course-1", GrantSelf, nil, GrantActive, grantedAt, nil)
	if err != nil {
		t.Fatalf("NewGrant: %v", err)
	}
	blocked, err := NewGrant("grant-2", "user-1", "course-1", GrantSelf, nil, GrantBlocked, grantedAt, nil)
	if err != nil {
		t.Fatalf("NewGrant blocked: %v", err)
	}
	revoked, err := NewGrant("grant-3", "user-1", "course-1", GrantSelf, nil, GrantRevoked, grantedAt, nil)
	if err != nil {
		t.Fatalf("NewGrant revoked: %v", err)
	}
	for _, test := range []struct {
		name    string
		grant   Grant
		call    func(Grant) (Grant, error)
		want    GrantState
		wantErr error
	}{
		{name: "active blocks", grant: active, call: Grant.Block, want: GrantBlocked},
		{name: "active revokes", grant: active, call: Grant.Revoke, want: GrantRevoked},
		{name: "active cannot unblock", grant: active, call: Grant.Unblock, wantErr: ErrGrantTransitionInvalid},
		{name: "blocked unblocks", grant: blocked, call: Grant.Unblock, want: GrantActive},
		{name: "blocked revokes", grant: blocked, call: Grant.Revoke, want: GrantRevoked},
		{name: "blocked cannot block", grant: blocked, call: Grant.Block, wantErr: ErrGrantTransitionInvalid},
		{name: "revoked cannot block", grant: revoked, call: Grant.Block, wantErr: ErrGrantTransitionInvalid},
		{name: "revoked cannot unblock", grant: revoked, call: Grant.Unblock, wantErr: ErrGrantTransitionInvalid},
		{name: "revoked cannot revoke", grant: revoked, call: Grant.Revoke, wantErr: ErrGrantTransitionInvalid},
	} {
		t.Run(test.name, func(t *testing.T) {
			next, err := test.call(test.grant)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error=%v want=%v", err, test.wantErr)
			}
			if test.wantErr == nil && next.State() != test.want {
				t.Fatalf("state=%v want=%v", next.State(), test.want)
			}
		})
	}
}

func TestSelfEnrollCreatesAnActiveSelfMembership(t *testing.T) {
	grantedAt := time.Date(2026, 10, 6, 9, 0, 0, 0, time.UTC)
	grant, err := SelfEnroll("grant-1", "user-1", "course-1", grantedAt)
	if err != nil {
		t.Fatalf("SelfEnroll: %v", err)
	}
	if grant.Source() != GrantSelf || grant.State() != GrantActive || grant.OrderID() != nil {
		t.Fatalf("grant = source %q, state %q, order %v; want self, active, nil", grant.Source(), grant.State(), grant.OrderID())
	}
}

func TestActiveGrantOfEverySourceBlocksAnotherOrder(t *testing.T) {
	orderID := "order-1"
	grantedAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name    string
		source  GrantSource
		orderID *string
	}{
		{name: "purchase", source: GrantPurchase, orderID: &orderID},
		{name: "admin", source: GrantAdmin},
		{name: "grandfather", source: GrantGrandfather},
		{name: "self", source: GrantSelf},
	} {
		t.Run(test.name, func(t *testing.T) {
			grant, err := NewGrant("grant-1", "user-1", "course-1", test.source, test.orderID, GrantActive, grantedAt, nil)
			if err != nil {
				t.Fatalf("NewGrant: %v", err)
			}
			if err := CanStartOrder("user-1", "course-1", []Grant{grant}); !errors.Is(err, ErrCourseAlreadyOwned) {
				t.Fatalf("error=%v want=%v", err, ErrCourseAlreadyOwned)
			}
		})
	}
}

func TestRevokedOrOtherCourseGrantDoesNotBlockOrder(t *testing.T) {
	grantedAt := time.Date(2026, 9, 30, 9, 0, 0, 0, time.UTC)
	revoked, err := NewGrant("grant-1", "user-1", "course-1", GrantAdmin, nil, GrantRevoked, grantedAt, nil)
	if err != nil {
		t.Fatalf("NewGrant revoked: %v", err)
	}
	otherCourse, err := NewGrant("grant-2", "user-1", "course-2", GrantAdmin, nil, GrantActive, grantedAt, nil)
	if err != nil {
		t.Fatalf("NewGrant other course: %v", err)
	}
	otherUser, err := NewGrant("grant-3", "user-2", "course-1", GrantAdmin, nil, GrantActive, grantedAt, nil)
	if err != nil {
		t.Fatalf("NewGrant other user: %v", err)
	}
	if err := CanStartOrder("user-1", "course-1", []Grant{revoked, otherCourse, otherUser}); err != nil {
		t.Fatalf("CanStartOrder: %v", err)
	}
}
