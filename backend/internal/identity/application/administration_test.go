package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/identity/application"
	"courses/internal/identity/domain"
	"courses/internal/identity/ports"
)

type administrationRepository struct {
	list        func(context.Context, ports.AdministrationQuery) ([]domain.User, int64, error)
	setStatus   func(context.Context, string, domain.AccountStatus, time.Time, ports.StatusPolicy) error
	delete      func(context.Context, string, time.Time, ports.DeletePolicy) error
	setPassword func(context.Context, string, string, time.Time) error
}

func (repository *administrationRepository) List(ctx context.Context, query ports.AdministrationQuery) ([]domain.User, int64, error) {
	return repository.list(ctx, query)
}

func (repository *administrationRepository) SetStatus(ctx context.Context, id string, status domain.AccountStatus, now time.Time, policy ports.StatusPolicy) error {
	return repository.setStatus(ctx, id, status, now, policy)
}

func (repository *administrationRepository) Delete(ctx context.Context, id string, now time.Time, policy ports.DeletePolicy) error {
	return repository.delete(ctx, id, now, policy)
}

func (repository *administrationRepository) SetPassword(ctx context.Context, id, password string, now time.Time) error {
	return repository.setPassword(ctx, id, password, now)
}

func TestListUsersForwardsTheDirectoryQueryAndValidatesStoredUsers(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	wantUser := administrationUser("40000000-0000-4000-8000-000000000001", domain.Student, now)
	wantQuery := application.ListUsersInput{Limit: 20, Offset: 40, Search: "student", Status: domain.AccountBlocked}
	repository := &administrationRepository{list: func(ctx context.Context, query ports.AdministrationQuery) ([]domain.User, int64, error) {
		assertBoundedDeadline(t, ctx)
		if query.Limit != wantQuery.Limit || query.Offset != wantQuery.Offset || query.Search != wantQuery.Search || query.Status != wantQuery.Status {
			t.Fatalf("query=%+v", query)
		}
		return []domain.User{wantUser}, 3, nil
	}}

	page, err := application.NewAdministration(repository, func() time.Time { return now }).ListUsers(context.Background(), wantQuery)
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 || page.Items[0] != wantUser || page.Limit != 20 || page.Offset != 40 || page.Total != 3 {
		t.Fatalf("page=%+v", page)
	}
}

func TestSetStatusRunsDomainPolicyOverLockedFacts(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 6, 12, 0, 0, 123456789, time.FixedZone("test", 2*60*60))
	wantNow := now.UTC().Truncate(time.Microsecond)
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}
	target := administrationUser("40000000-0000-4000-8000-000000000002", domain.Admin, wantNow)
	repository := &administrationRepository{setStatus: func(ctx context.Context, id string, status domain.AccountStatus, at time.Time, policy ports.StatusPolicy) error {
		assertBoundedDeadline(t, ctx)
		if id != target.ID || status != domain.AccountBlocked || !at.Equal(wantNow) {
			t.Fatalf("id=%q status=%q at=%v", id, status, at)
		}
		return policy(target, 0)
	}}

	err := application.NewAdministration(repository, func() time.Time { return now }).SetStatus(context.Background(), actor, target.ID, domain.AccountBlocked)
	if !errors.Is(err, domain.ErrLastAdmin) {
		t.Fatalf("got %v", err)
	}
}

func TestSetStatusRefusesSelfBeforeRepository(t *testing.T) {
	t.Parallel()

	called := false
	repository := &administrationRepository{setStatus: func(context.Context, string, domain.AccountStatus, time.Time, ports.StatusPolicy) error {
		called = true
		return nil
	}}
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}

	err := application.NewAdministration(repository, time.Now).SetStatus(context.Background(), actor, actor.UserID, domain.AccountBlocked)
	if !errors.Is(err, domain.ErrCannotTargetSelf) {
		t.Fatalf("got %v", err)
	}
	if called {
		t.Fatal("repository called for self-block")
	}
}

func TestRestoreDoesNotApplyTheLastAdminBlockPolicy(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}
	target := administrationUser("40000000-0000-4000-8000-000000000002", domain.Admin, now)
	disabledAt := now
	target.DisabledAt = &disabledAt
	repository := &administrationRepository{setStatus: func(_ context.Context, _ string, status domain.AccountStatus, _ time.Time, policy ports.StatusPolicy) error {
		if status != domain.AccountActive {
			t.Fatalf("status=%q", status)
		}
		return policy(target, 0)
	}}

	err := application.NewAdministration(repository, func() time.Time { return now }).SetStatus(context.Background(), actor, target.ID, domain.AccountActive)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
}

func TestDeleteRunsDomainPolicyAndRefusesSelfBeforeRepository(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}
	admin := administrationUser("40000000-0000-4000-8000-000000000002", domain.Admin, now)
	student := administrationUser("40000000-0000-4000-8000-000000000003", domain.Student, now)
	called := 0
	repository := &administrationRepository{delete: func(ctx context.Context, id string, at time.Time, policy ports.DeletePolicy) error {
		called++
		assertBoundedDeadline(t, ctx)
		if !at.Equal(now) {
			t.Fatalf("id=%q at=%v", id, at)
		}
		if id == admin.ID {
			return policy(admin)
		}
		return policy(student)
	}}
	administration := application.NewAdministration(repository, func() time.Time { return now })

	if err := administration.Delete(context.Background(), actor, admin.ID); !errors.Is(err, domain.ErrAdminImmutable) {
		t.Fatalf("delete admin: %v", err)
	}
	if err := administration.Delete(context.Background(), actor, actor.UserID); !errors.Is(err, domain.ErrCannotTargetSelf) {
		t.Fatalf("delete self: %v", err)
	}
	if err := administration.Delete(context.Background(), actor, student.ID); err != nil {
		t.Fatalf("delete student: %v", err)
	}
	if called != 2 {
		t.Fatalf("repository calls=%d", called)
	}
}

func TestResetPasswordUsesTheRepositoryConsumerHashPath(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 6, 12, 0, 0, 987654321, time.FixedZone("test", -3*60*60))
	wantNow := now.UTC().Truncate(time.Microsecond)
	wantID := "40000000-0000-4000-8000-000000000002"
	repository := &administrationRepository{setPassword: func(ctx context.Context, id, password string, at time.Time) error {
		assertBoundedDeadline(t, ctx)
		if id != wantID || password != "new password" || !at.Equal(wantNow) {
			t.Fatalf("id=%q password=%q at=%v", id, password, at)
		}
		return nil
	}}

	err := application.NewAdministration(repository, func() time.Time { return now }).ResetPassword(context.Background(), wantID, "new password")
	if err != nil {
		t.Fatal(err)
	}
}

func TestResetPasswordRejectsInvalidBoundsBeforeRepository(t *testing.T) {
	t.Parallel()

	called := false
	repository := &administrationRepository{setPassword: func(context.Context, string, string, time.Time) error {
		called = true
		return nil
	}}
	err := application.NewAdministration(repository, time.Now).ResetPassword(context.Background(), "40000000-0000-4000-8000-000000000002", "")
	if !errors.Is(err, domain.ErrInvalidUser) {
		t.Fatalf("got %v", err)
	}
	if called {
		t.Fatal("repository called for invalid password")
	}
}

func TestDependencyFailurePassesThroughWithoutInternalDetail(t *testing.T) {
	t.Parallel()

	repository := &administrationRepository{
		list: func(context.Context, ports.AdministrationQuery) ([]domain.User, int64, error) {
			return nil, 0, domain.ErrDependencyUnavailable
		},
		setStatus: func(context.Context, string, domain.AccountStatus, time.Time, ports.StatusPolicy) error {
			return domain.ErrDependencyUnavailable
		},
	}
	administration := application.NewAdministration(repository, time.Now)
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}

	if _, err := administration.ListUsers(context.Background(), application.ListUsersInput{Limit: 20}); err != domain.ErrDependencyUnavailable || err.Error() != "dependency_unavailable" {
		t.Fatalf("list detail escaped: %v", err)
	}
	if err := administration.SetStatus(context.Background(), actor, "40000000-0000-4000-8000-000000000002", domain.AccountBlocked); err != domain.ErrDependencyUnavailable || err.Error() != "dependency_unavailable" {
		t.Fatalf("status detail escaped: %v", err)
	}
}

func TestCancelledSetStatusIsReportedAsCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	repository := &administrationRepository{setStatus: func(callCtx context.Context, _ string, _ domain.AccountStatus, _ time.Time, _ ports.StatusPolicy) error {
		cancel()
		<-callCtx.Done()
		return domain.ErrDependencyUnavailable
	}}
	actor := domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}

	err := application.NewAdministration(repository, time.Now).SetStatus(ctx, actor, "40000000-0000-4000-8000-000000000002", domain.AccountBlocked)
	if !errors.Is(err, context.Canceled) || errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func assertBoundedDeadline(t *testing.T, ctx context.Context) {
	t.Helper()
	deadline, ok := ctx.Deadline()
	remaining := time.Until(deadline)
	if !ok || remaining <= 0 || remaining > 3*time.Second {
		t.Fatalf("deadline ok=%v remaining=%v", ok, remaining)
	}
}

func administrationUser(id string, role domain.Role, now time.Time) domain.User {
	return domain.User{ID: id, Email: "user-" + id[len(id)-1:] + "@courseforge.local", DisplayName: "Identity User", Role: role, CreatedAt: now, UpdatedAt: now}
}
