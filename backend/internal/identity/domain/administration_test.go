package domain_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"courses/internal/identity/domain"
)

func TestAccountStatusProjection(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

	for _, test := range []struct {
		name       string
		user       domain.User
		wantStatus domain.AccountStatus
		wantActive bool
	}{
		{name: "active", user: domain.User{}, wantStatus: domain.AccountActive, wantActive: true},
		{name: "blocked", user: domain.User{DisabledAt: &now}, wantStatus: domain.AccountBlocked},
		{name: "deleted takes precedence", user: domain.User{DisabledAt: &now, DeletedAt: &now}, wantStatus: domain.AccountDeleted},
		{name: "deleted without disabled is inactive", user: domain.User{DeletedAt: &now}, wantStatus: domain.AccountDeleted},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := test.user.AccountStatus(); got != test.wantStatus {
				t.Fatalf("status=%q, want %q", got, test.wantStatus)
			}
			if got := test.user.Active(); got != test.wantActive {
				t.Fatalf("active=%v, want %v", got, test.wantActive)
			}
		})
	}
}

func TestAccountAdministrationPolicy(t *testing.T) {
	t.Parallel()
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000003"}
	student := domain.User{ID: "40000000-0000-4000-8000-000000000001", Role: domain.Student}
	admin := domain.User{ID: "40000000-0000-4000-8000-000000000004", Role: domain.Admin}
	blockedAt := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	blockedAdmin := admin
	blockedAdmin.DisabledAt = &blockedAt
	self := admin
	self.ID = actor.UserID

	for _, test := range []struct {
		name string
		act  func() error
		want error
	}{
		{name: "block non-admin", act: func() error { return domain.CanBlock(actor, student, 0) }},
		{name: "block admin with one other active admin", act: func() error { return domain.CanBlock(actor, admin, 1) }},
		{name: "block last active admin", act: func() error { return domain.CanBlock(actor, admin, 0) }, want: domain.ErrLastAdmin},
		{name: "block already-blocked admin", act: func() error { return domain.CanBlock(actor, blockedAdmin, 0) }},
		{name: "block self", act: func() error { return domain.CanBlock(actor, self, 1) }, want: domain.ErrCannotTargetSelf},
		{name: "delete non-admin", act: func() error { return domain.CanDelete(actor, student) }},
		{name: "delete admin", act: func() error { return domain.CanDelete(actor, admin) }, want: domain.ErrAdminImmutable},
		{name: "delete self before admin check", act: func() error { return domain.CanDelete(actor, self) }, want: domain.ErrCannotTargetSelf},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := test.act(); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
		})
	}
}

func TestPasswordBounds(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name     string
		password string
		want     bool
	}{
		{name: "empty", password: ""},
		{name: "one byte", password: "p", want: true},
		{name: "whitespace only", password: " ", want: true},
		{name: "exactly 1024 bytes", password: strings.Repeat("p", 1024), want: true},
		{name: "over 1024 bytes", password: strings.Repeat("p", 1025)},
		{name: "multibyte exactly 1024 bytes", password: strings.Repeat("і", 512), want: true},
		{name: "multibyte over 1024 bytes", password: strings.Repeat("і", 513)},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := domain.ValidPassword(test.password); got != test.want {
				t.Fatalf("got %v, want %v", got, test.want)
			}
		})
	}
}
