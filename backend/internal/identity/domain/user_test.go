package domain_test

import (
	"errors"
	"testing"
	"time"

	"courses/internal/identity/domain"
)

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name  string
		input string
		want  string
	}{
		{name: "trims and lowercases", input: "  Student@CourseForge.Local  ", want: "student@courseforge.local"},
		{name: "unicode local part", input: "Викладач@CourseForge.Local", want: "викладач@courseforge.local"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := domain.NormalizeEmail(test.input)
			if err != nil {
				t.Fatal(err)
			}
			if got != test.want {
				t.Fatalf("got %q, want %q", got, test.want)
			}
		})
	}

	for _, input := range []string{
		"",
		"a@",
		"@b",
		"a@@b",
		"a b@example.test",
		"a\n@example.test",
		string([]byte{0xff, '@', 'b'}),
		"a@" + repeat("b", 253),
	} {
		t.Run("rejects "+input, func(t *testing.T) {
			t.Parallel()
			if _, err := domain.NormalizeEmail(input); !errors.Is(err, domain.ErrInvalidEmail) {
				t.Fatalf("got %v", err)
			}
		})
	}
}

func TestUserInvariantsAndRolePolicy(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 9, 3, 12, 0, 0, 0, time.UTC)
	valid := domain.User{
		ID:          "40000000-0000-4000-8000-000000000001",
		Email:       "student@courseforge.local",
		DisplayName: "Студентка Курсу",
		Role:        domain.Student,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if _, err := domain.NewUser(valid); err != nil {
		t.Fatal(err)
	}

	for _, test := range []struct {
		name   string
		change func(*domain.User)
	}{
		{name: "invalid id", change: func(user *domain.User) { user.ID = "invalid" }},
		{name: "noncanonical email", change: func(user *domain.User) { user.Email = " Student@CourseForge.Local " }},
		{name: "blank display name", change: func(user *domain.User) { user.DisplayName = " " }},
		{name: "untrimmed display name", change: func(user *domain.User) { user.DisplayName = " Student " }},
		{name: "long display name", change: func(user *domain.User) { user.DisplayName = repeat("і", 121) }},
		{name: "invalid role", change: func(user *domain.User) { user.Role = "owner" }},
		{name: "missing created timestamp", change: func(user *domain.User) { user.CreatedAt = time.Time{} }},
		{name: "update before create", change: func(user *domain.User) { user.UpdatedAt = now.Add(-time.Second) }},
		{name: "disabled before create", change: func(user *domain.User) { disabled := now.Add(-time.Second); user.DisabledAt = &disabled }},
		{name: "disabled after update", change: func(user *domain.User) { disabled := now.Add(time.Second); user.DisabledAt = &disabled }},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			user := valid
			test.change(&user)
			if _, err := domain.NewUser(user); !errors.Is(err, domain.ErrInvalidUser) {
				t.Fatalf("got %v", err)
			}
		})
	}

	for _, test := range []struct {
		role            domain.Role
		learn           bool
		adminArea       bool
		administerUsers bool
	}{
		{role: domain.Student, learn: true},
		{role: domain.Instructor, learn: true, adminArea: true},
		{role: domain.Admin, learn: true, adminArea: true, administerUsers: true},
		{role: domain.Role("owner")},
	} {
		if got := test.role.CanLearn(); got != test.learn {
			t.Errorf("%s CanLearn=%v", test.role, got)
		}
		if got := test.role.CanAccessAdministration(); got != test.adminArea {
			t.Errorf("%s CanAccessAdministration=%v", test.role, got)
		}
		if got := test.role.CanAdministerAccounts(); got != test.administerUsers {
			t.Errorf("%s CanAdministerAccounts=%v", test.role, got)
		}
	}
}

func repeat(value string, count int) string {
	result := ""
	for range count {
		result += value
	}
	return result
}
