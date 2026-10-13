package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/course/application"
	"courses/internal/course/domain"
)

const (
	authoringAdminID      = "40000000-0000-4000-8000-000000000003"
	authoringInstructorID = "40000000-0000-4000-8000-000000000002"
	authoringOtherOwnerID = "40000000-0000-4000-8000-000000000004"
)

type authoringRepository struct {
	calls        []string
	list         func(context.Context, domain.Actor, domain.Query, domain.Status) (domain.AuthoringPage, error)
	load         func(context.Context, string) (domain.CourseTree, error)
	create       func(context.Context, domain.Actor, domain.CourseDraft, time.Time) (domain.CourseTree, error)
	save         func(context.Context, domain.Actor, domain.CourseTree, time.Time, time.Time) (domain.CourseTree, error)
	transition   func(context.Context, domain.Actor, string, domain.Status, domain.Status, time.Time, time.Time) (domain.CourseTree, error)
	deleteCourse func(context.Context, domain.Actor, string, time.Time) error
}

type grandfathererFunc func(context.Context, string, time.Time, int64, string, time.Time) error

func (fn grandfathererFunc) GrandfatherOnPriceSet(ctx context.Context, courseID string, expected time.Time, minor int64, currency string, updatedAt time.Time) error {
	return fn(ctx, courseID, expected, minor, currency, updatedAt)
}

func (repository *authoringRepository) List(ctx context.Context, actor domain.Actor, query domain.Query, status domain.Status) (domain.AuthoringPage, error) {
	repository.calls = append(repository.calls, "list")
	return repository.list(ctx, actor, query, status)
}

func (repository *authoringRepository) Load(ctx context.Context, id string) (domain.CourseTree, error) {
	repository.calls = append(repository.calls, "load")
	return repository.load(ctx, id)
}

func (repository *authoringRepository) Create(ctx context.Context, actor domain.Actor, draft domain.CourseDraft, now time.Time) (domain.CourseTree, error) {
	repository.calls = append(repository.calls, "create")
	return repository.create(ctx, actor, draft, now)
}

func (repository *authoringRepository) Save(ctx context.Context, actor domain.Actor, tree domain.CourseTree, expectedUpdatedAt, updatedAt time.Time) (domain.CourseTree, error) {
	repository.calls = append(repository.calls, "save")
	return repository.save(ctx, actor, tree, expectedUpdatedAt, updatedAt)
}

func (repository *authoringRepository) Transition(ctx context.Context, actor domain.Actor, id string, from, to domain.Status, expectedUpdatedAt, updatedAt time.Time) (domain.CourseTree, error) {
	repository.calls = append(repository.calls, "transition")
	return repository.transition(ctx, actor, id, from, to, expectedUpdatedAt, updatedAt)
}

func (repository *authoringRepository) Delete(ctx context.Context, actor domain.Actor, id string, expectedUpdatedAt time.Time) error {
	repository.calls = append(repository.calls, "delete")
	return repository.deleteCourse(ctx, actor, id, expectedUpdatedAt)
}

func TestAuthoringListBoundsTheCall(t *testing.T) {
	t.Parallel()

	repository := &authoringRepository{list: func(ctx context.Context, actor domain.Actor, query domain.Query, status domain.Status) (domain.AuthoringPage, error) {
		deadline, ok := ctx.Deadline()
		remaining := time.Until(deadline)
		if !ok || remaining <= 0 || remaining > 3*time.Second {
			t.Fatalf("unbounded deadline: %v, %v", deadline, ok)
		}
		if actor != adminActor() || query.Search != "" || query.Limit != 20 || query.Offset != 40 || query.Sort != domain.SortRecent || status != domain.Draft {
			t.Fatalf("wrong repository input: actor=%+v query=%+v status=%s", actor, query, status)
		}
		return domain.AuthoringPage{Items: []domain.AuthoringCourse{}, Limit: query.Limit, Offset: query.Offset, Total: 3}, nil
	}}

	page, err := newAuthoring(repository).List(context.Background(), adminActor(), application.AuthoringListInput{Limit: 20, Offset: 40, Status: domain.Draft})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 3 || page.Limit != 20 || page.Offset != 40 {
		t.Fatalf("wrong page: %+v", page)
	}
}

func TestAuthoringRejectsRoleAndMalformedInputBeforeRepository(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		act  func(*application.Authoring) error
		want error
	}{
		{name: "student list", act: func(authoring *application.Authoring) error {
			_, err := authoring.List(context.Background(), domain.Actor{UserID: authoringInstructorID, Role: domain.ActorStudent}, application.AuthoringListInput{Limit: 20})
			return err
		}, want: domain.ErrForbidden},
		{name: "invalid list status", act: func(authoring *application.Authoring) error {
			_, err := authoring.List(context.Background(), adminActor(), application.AuthoringListInput{Limit: 20, Status: "unknown"})
			return err
		}, want: domain.ErrInvalidRequest},
		{name: "malformed load id", act: func(authoring *application.Authoring) error {
			_, err := authoring.Load(context.Background(), adminActor(), "not-a-uuid")
			return err
		}, want: domain.ErrInvalidRequest},
		{name: "zero save token", act: func(authoring *application.Authoring) error {
			_, err := authoring.Save(context.Background(), adminActor(), applicationTree(), time.Time{})
			return err
		}, want: domain.ErrInvalidRequest},
		{name: "invalid save price", act: func(authoring *application.Authoring) error {
			tree := applicationTree()
			tree.Price = &domain.Price{Minor: 0, Currency: "UAH"}
			tree.PriceSet = true
			_, err := authoring.Save(context.Background(), adminActor(), tree, tree.UpdatedAt)
			return err
		}, want: domain.ErrInvalidRequest},
		{name: "instructor reassignment", act: func(authoring *application.Authoring) error {
			tree := applicationTree()
			tree.InstructorID = authoringOtherOwnerID
			_, err := authoring.Save(context.Background(), instructorActor(), tree, tree.UpdatedAt)
			return err
		}, want: domain.ErrCourseForbidden},
		{name: "zero transition token", act: func(authoring *application.Authoring) error {
			_, err := authoring.Transition(context.Background(), adminActor(), applicationTree().ID, domain.Published, time.Time{})
			return err
		}, want: domain.ErrInvalidRequest},
		{name: "zero delete token", act: func(authoring *application.Authoring) error {
			return authoring.Delete(context.Background(), adminActor(), applicationTree().ID, time.Time{})
		}, want: domain.ErrInvalidRequest},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			repository := &authoringRepository{}
			if err := test.act(newAuthoring(repository)); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if len(repository.calls) != 0 {
				t.Fatalf("repository called: %v", repository.calls)
			}
		})
	}
}

func TestInstructorMayOmitSelfButCannotReassign(t *testing.T) {
	t.Parallel()

	actor := instructorActor()
	draft := applicationDraft()
	draft.InstructorID = ""
	repository := &authoringRepository{create: func(_ context.Context, gotActor domain.Actor, gotDraft domain.CourseDraft, _ time.Time) (domain.CourseTree, error) {
		if gotActor != actor || gotDraft.InstructorID != actor.UserID {
			t.Fatalf("owner not forced: actor=%+v draft=%+v", gotActor, gotDraft)
		}
		return applicationTree(), nil
	}}
	if _, err := newAuthoring(repository).Create(context.Background(), actor, draft); err != nil {
		t.Fatal(err)
	}

	stored := applicationTree()
	edit := stored
	edit.InstructorID = ""
	saveRepository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) { return stored, nil },
		save: func(_ context.Context, _ domain.Actor, tree domain.CourseTree, _, updatedAt time.Time) (domain.CourseTree, error) {
			if tree.InstructorID != actor.UserID {
				t.Fatalf("save owner not filled: %+v", tree)
			}
			tree.UpdatedAt = updatedAt
			return tree, nil
		},
	}
	if _, err := newAuthoring(saveRepository).Save(context.Background(), actor, edit, stored.UpdatedAt); err != nil {
		t.Fatal(err)
	}

	draft.InstructorID = authoringOtherOwnerID
	if _, err := newAuthoring(&authoringRepository{}).Create(context.Background(), actor, draft); !errors.Is(err, domain.ErrCourseForbidden) {
		t.Fatalf("got %v", err)
	}
}

func TestStoredOwnershipAndLifecyclePoliciesRunAfterLoad(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		status domain.Status
		act    func(*application.Authoring, domain.CourseTree) error
		want   error
	}{
		{name: "foreign load", act: func(authoring *application.Authoring, tree domain.CourseTree) error {
			_, err := authoring.Load(context.Background(), instructorActor(), tree.ID)
			return err
		}, want: domain.ErrCourseForbidden},
		{name: "foreign save", act: func(authoring *application.Authoring, tree domain.CourseTree) error {
			_, err := authoring.Save(context.Background(), instructorActor(), tree, tree.UpdatedAt)
			return err
		}, want: domain.ErrCourseForbidden},
		{name: "foreign transition", act: func(authoring *application.Authoring, tree domain.CourseTree) error {
			_, err := authoring.Transition(context.Background(), instructorActor(), tree.ID, domain.Published, tree.UpdatedAt)
			return err
		}, want: domain.ErrCourseForbidden},
		{name: "published delete", status: domain.Published, act: func(authoring *application.Authoring, tree domain.CourseTree) error {
			return authoring.Delete(context.Background(), adminActor(), tree.ID, tree.UpdatedAt)
		}, want: domain.ErrCourseNotDeletable},
		{name: "illegal transition", status: domain.Published, act: func(authoring *application.Authoring, tree domain.CourseTree) error {
			_, err := authoring.Transition(context.Background(), adminActor(), tree.ID, domain.Draft, tree.UpdatedAt)
			return err
		}, want: domain.ErrInvalidStatusTransition},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			stored := applicationTree()
			stored.InstructorID = authoringOtherOwnerID
			if test.status != "" {
				stored.Status = test.status
				if test.status == domain.Published {
					stored.PublishedAt = &stored.UpdatedAt
				}
			}
			repository := &authoringRepository{load: func(context.Context, string) (domain.CourseTree, error) { return stored, nil }}
			if err := test.act(newAuthoring(repository), applicationTree()); !errors.Is(err, test.want) {
				t.Fatalf("got %v, want %v", err, test.want)
			}
			if len(repository.calls) != 1 || repository.calls[0] != "load" {
				t.Fatalf("calls=%v", repository.calls)
			}
		})
	}
}

func TestPublishedSlugChangeIsRefusedAfterLoad(t *testing.T) {
	t.Parallel()

	stored := applicationTree()
	stored.Status = domain.Published
	stored.PublishedAt = &stored.UpdatedAt
	proposed := stored
	proposed.Slug = "renamed-course"
	repository := &authoringRepository{load: func(context.Context, string) (domain.CourseTree, error) { return stored, nil }}
	_, err := newAuthoring(repository).Save(context.Background(), adminActor(), proposed, stored.UpdatedAt)
	if !errors.Is(err, domain.ErrSlugImmutable) {
		t.Fatalf("got %v", err)
	}
	if len(repository.calls) != 1 || repository.calls[0] != "load" {
		t.Fatalf("calls=%v", repository.calls)
	}
}

func TestDuplicatePayloadIDsAreRefusedBeforeAnyRepositoryCall(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name   string
		change func(*domain.CourseTree)
	}{
		{name: "modules", change: func(tree *domain.CourseTree) {
			tree.Modules = append(tree.Modules, domain.Module{ID: tree.Modules[0].ID, Title: "Duplicate", Position: 1})
		}},
		{name: "lessons", change: func(tree *domain.CourseTree) {
			tree.Modules[0].Lessons = append(tree.Modules[0].Lessons, domain.Lesson{ID: tree.Modules[0].Lessons[0].ID, Title: "Duplicate", Position: 1})
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tree := applicationTree()
			test.change(&tree)
			repository := &authoringRepository{}
			_, err := newAuthoring(repository).Save(context.Background(), adminActor(), tree, tree.UpdatedAt)
			if !errors.Is(err, domain.ErrInvalidRequest) {
				t.Fatalf("got %v", err)
			}
			if len(repository.calls) != 0 {
				t.Fatalf("repository called: %v", repository.calls)
			}
		})
	}
}

func TestDependencyFailurePassesThroughWithoutInternalDetail(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name string
		act  func(*application.Authoring) error
	}{
		{name: "list", act: func(authoring *application.Authoring) error {
			_, err := authoring.List(context.Background(), adminActor(), application.AuthoringListInput{Limit: 20})
			return err
		}},
		{name: "save", act: func(authoring *application.Authoring) error {
			tree := applicationTree()
			_, err := authoring.Save(context.Background(), adminActor(), tree, tree.UpdatedAt)
			return err
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			tree := applicationTree()
			repository := &authoringRepository{
				list: func(context.Context, domain.Actor, domain.Query, domain.Status) (domain.AuthoringPage, error) {
					return domain.AuthoringPage{}, domain.ErrDependencyUnavailable
				},
				load: func(context.Context, string) (domain.CourseTree, error) { return tree, nil },
				save: func(context.Context, domain.Actor, domain.CourseTree, time.Time, time.Time) (domain.CourseTree, error) {
					return domain.CourseTree{}, domain.ErrDependencyUnavailable
				},
			}
			err := test.act(newAuthoring(repository))
			if err != domain.ErrDependencyUnavailable || err.Error() != "dependency_unavailable" {
				t.Fatalf("dependency detail escaped: %v", err)
			}
		})
	}
}

func TestCancelledSaveIsReportedAsCancellation(t *testing.T) {
	t.Parallel()

	tree := applicationTree()
	ctx, cancel := context.WithCancel(context.Background())
	repository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) { return tree, nil },
		save: func(callCtx context.Context, _ domain.Actor, _ domain.CourseTree, _, _ time.Time) (domain.CourseTree, error) {
			cancel()
			<-callCtx.Done()
			return domain.CourseTree{}, domain.ErrDependencyUnavailable
		},
	}
	_, err := newAuthoring(repository).Save(ctx, adminActor(), tree, tree.UpdatedAt)
	if !errors.Is(err, context.Canceled) || errors.Is(err, domain.ErrDependencyUnavailable) {
		t.Fatalf("got %v", err)
	}
}

func TestWriteTokensAndMicrosecondClockReachRepository(t *testing.T) {
	t.Parallel()

	stored := applicationTree()
	edit := stored
	edit.Status = ""
	edit.CreatedAt = time.Time{}
	edit.UpdatedAt = time.Time{}
	edit.PublishedAt = nil
	expected := stored.UpdatedAt.Add(123 * time.Nanosecond)
	clockReading := expected.Add(-time.Second)
	wantUpdatedAt := expected.UTC().Truncate(time.Microsecond).Add(time.Microsecond)
	repository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) { return stored, nil },
		save: func(_ context.Context, _ domain.Actor, saved domain.CourseTree, gotExpected, gotUpdatedAt time.Time) (domain.CourseTree, error) {
			if gotExpected != expected {
				t.Fatalf("expected token changed: got %s want %s", gotExpected, expected)
			}
			if saved.Status != stored.Status || saved.CreatedAt != stored.CreatedAt || saved.PublishedAt != stored.PublishedAt {
				t.Fatalf("stored lifecycle changed: %+v", saved)
			}
			if gotUpdatedAt != wantUpdatedAt || gotUpdatedAt.Nanosecond()%int(time.Microsecond) != 0 {
				t.Fatalf("write timestamp=%s want=%s", gotUpdatedAt, wantUpdatedAt)
			}
			saved.UpdatedAt = gotUpdatedAt
			return saved, nil
		},
	}
	if _, err := application.NewAuthoring(repository, func() time.Time { return clockReading }).Save(context.Background(), adminActor(), edit, expected); err != nil {
		t.Fatal(err)
	}
}

func TestFirstPriceSetUsesAtomicGrandfatherOperation(t *testing.T) {
	stored := applicationTree()
	edit := stored
	edit.Price = &domain.Price{Minor: 50000, Currency: "UAH"}
	edit.PriceSet = true
	loaded := stored
	loaded.Price = edit.Price
	loaded.UpdatedAt = stored.UpdatedAt.Add(time.Microsecond)
	loadCalls := 0
	repository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) {
			loadCalls++
			if loadCalls == 1 {
				return stored, nil
			}
			return loaded, nil
		},
		save: func(_ context.Context, _ domain.Actor, tree domain.CourseTree, _, _ time.Time) (domain.CourseTree, error) {
			t.Fatalf("first price must not use the non-atomic course save: %+v", tree)
			return domain.CourseTree{}, nil
		},
	}
	called := false
	grandfatherer := grandfathererFunc(func(_ context.Context, courseID string, expected time.Time, minor int64, currency string, updatedAt time.Time) error {
		called = true
		if courseID != stored.ID || expected != stored.UpdatedAt || minor != 50000 || currency != "UAH" || !updatedAt.After(expected) {
			t.Fatalf("atomic price args: id=%s expected=%s minor=%d currency=%s updated=%s", courseID, expected, minor, currency, updatedAt)
		}
		return nil
	})
	saved, err := application.NewAuthoring(repository, func() time.Time { return loaded.UpdatedAt }, grandfatherer).Save(context.Background(), adminActor(), edit, stored.UpdatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if !called || saved.Price == nil || saved.Price.Minor != 50000 || loadCalls != 2 {
		t.Fatalf("saved=%+v called=%v loads=%d", saved, called, loadCalls)
	}
}

// withPreviewFlags returns a deep copy of modules with every lesson's preview flag
// explicitly specified, as the pricing editor sends it; flip changes the value.
func withPreviewFlags(modules []domain.Module, flip bool) []domain.Module {
	copied := make([]domain.Module, len(modules))
	for i, module := range modules {
		copied[i] = module
		copied[i].Lessons = make([]domain.Lesson, len(module.Lessons))
		for j, lesson := range module.Lessons {
			lesson.PreviewSet = true
			if flip {
				lesson.IsPreview = !lesson.IsPreview
			}
			copied[i].Lessons[j] = lesson
		}
	}
	return copied
}

func TestFirstPriceSetAcceptsUnchangedPreviewsSpecifiedByTheEditor(t *testing.T) {
	stored := applicationTree()
	if len(stored.Modules) == 0 || len(stored.Modules[0].Lessons) == 0 {
		t.Fatal("fixture must carry at least one lesson")
	}
	edit := stored
	edit.Modules = withPreviewFlags(stored.Modules, false)
	edit.Price = &domain.Price{Minor: 50000, Currency: "UAH"}
	edit.PriceSet = true
	loaded := stored
	loaded.Price = edit.Price
	loaded.UpdatedAt = stored.UpdatedAt.Add(time.Microsecond)
	loadCalls := 0
	repository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) {
			loadCalls++
			if loadCalls == 1 {
				return stored, nil
			}
			return loaded, nil
		},
		save: func(_ context.Context, _ domain.Actor, tree domain.CourseTree, _, _ time.Time) (domain.CourseTree, error) {
			t.Fatalf("first price must not use the non-atomic course save: %+v", tree)
			return domain.CourseTree{}, nil
		},
	}
	called := false
	grandfatherer := grandfathererFunc(func(context.Context, string, time.Time, int64, string, time.Time) error {
		called = true
		return nil
	})
	saved, err := application.NewAuthoring(repository, func() time.Time { return loaded.UpdatedAt }, grandfatherer).Save(context.Background(), adminActor(), edit, stored.UpdatedAt)
	if err != nil {
		t.Fatalf("an editor that states the stored previews must still be able to set the first price: %v", err)
	}
	if !called || saved.Price == nil || saved.Price.Minor != 50000 {
		t.Fatalf("saved=%+v called=%v", saved, called)
	}
}

func TestFirstPriceSetStillRefusesAPreviewChange(t *testing.T) {
	stored := applicationTree()
	edit := stored
	edit.Modules = withPreviewFlags(stored.Modules, true)
	edit.Price = &domain.Price{Minor: 50000, Currency: "UAH"}
	edit.PriceSet = true
	repository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) { return stored, nil },
		save: func(_ context.Context, _ domain.Actor, tree domain.CourseTree, _, _ time.Time) (domain.CourseTree, error) {
			t.Fatalf("a refused first price must not save: %+v", tree)
			return domain.CourseTree{}, nil
		},
	}
	grandfatherer := grandfathererFunc(func(context.Context, string, time.Time, int64, string, time.Time) error {
		t.Fatal("a refused first price must not grandfather")
		return nil
	})
	_, err := application.NewAuthoring(repository, func() time.Time { return stored.UpdatedAt.Add(time.Second) }, grandfatherer).Save(context.Background(), adminActor(), edit, stored.UpdatedAt)
	if !errors.Is(err, domain.ErrInvalidRequest) {
		t.Fatalf("first price with a preview change: want ErrInvalidRequest, got %v", err)
	}
}

func TestTransitionAndDeleteCarryExpectedToken(t *testing.T) {
	t.Parallel()

	tree := applicationTree()
	expected := tree.UpdatedAt
	repository := &authoringRepository{
		load: func(context.Context, string) (domain.CourseTree, error) { return tree, nil },
		transition: func(_ context.Context, actor domain.Actor, id string, from, to domain.Status, gotExpected, updatedAt time.Time) (domain.CourseTree, error) {
			if actor != adminActor() || id != tree.ID || from != domain.Draft || to != domain.Published || gotExpected != expected || !updatedAt.After(expected) {
				t.Fatalf("transition args: actor=%+v id=%s %s->%s expected=%s updated=%s", actor, id, from, to, gotExpected, updatedAt)
			}
			result := tree
			result.Status = domain.Published
			result.UpdatedAt = updatedAt
			result.PublishedAt = &updatedAt
			return result, nil
		},
		deleteCourse: func(_ context.Context, actor domain.Actor, id string, gotExpected time.Time) error {
			if actor != adminActor() || id != tree.ID || gotExpected != expected {
				t.Fatalf("delete args: actor=%+v id=%s expected=%s", actor, id, gotExpected)
			}
			return nil
		},
	}
	authoring := newAuthoring(repository)
	if _, err := authoring.Transition(context.Background(), adminActor(), tree.ID, domain.Published, expected); err != nil {
		t.Fatal(err)
	}
	if err := authoring.Delete(context.Background(), adminActor(), tree.ID, expected); err != nil {
		t.Fatal(err)
	}
}

func newAuthoring(repository *authoringRepository) *application.Authoring {
	return application.NewAuthoring(repository, func() time.Time {
		return time.Date(2026, 9, 5, 12, 0, 1, 123456789, time.UTC)
	})
}

func adminActor() domain.Actor {
	return domain.Actor{UserID: authoringAdminID, Role: domain.ActorAdmin}
}

func instructorActor() domain.Actor {
	return domain.Actor{UserID: authoringInstructorID, Role: domain.ActorInstructor}
}

func applicationDraft() domain.CourseDraft {
	return domain.CourseDraft{Slug: "course", Title: "Course", Summary: "Summary", Description: "Description", InstructorID: authoringInstructorID}
}

func applicationTree() domain.CourseTree {
	now := time.Date(2026, 9, 5, 12, 0, 0, 0, time.UTC)
	return domain.CourseTree{
		ID:           "10000000-0000-4000-8000-000000000001",
		Slug:         "course",
		Title:        "Course",
		Summary:      "Summary",
		Description:  "Description",
		Status:       domain.Draft,
		InstructorID: authoringInstructorID,
		CreatedAt:    now,
		UpdatedAt:    now,
		Modules: []domain.Module{{
			ID:       "20000000-0000-4000-8000-000000000001",
			Title:    "Module",
			Position: 0,
			Lessons: []domain.Lesson{{
				ID:       "30000000-0000-4000-8000-000000000001",
				Title:    "Lesson",
				Position: 0,
			}},
		}},
	}
}
