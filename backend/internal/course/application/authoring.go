package application

import (
	"context"
	"reflect"
	"time"

	"courses/internal/course/domain"
	"courses/internal/course/ports"
)

const authoringOperationTimeout = 3 * time.Second

type AuthoringClock func() time.Time

type Authoring struct {
	repository    ports.AuthoringRepository
	now           AuthoringClock
	grandfatherer ports.Grandfatherer
}

type AuthoringListInput struct {
	Limit  int
	Offset int
	Status domain.Status
	Search string
}

func NewAuthoring(repository ports.AuthoringRepository, now AuthoringClock, grandfatherer ...ports.Grandfatherer) *Authoring {
	var grants ports.Grandfatherer
	if len(grandfatherer) != 0 {
		grants = grandfatherer[0]
	}
	return &Authoring{repository: repository, now: now, grandfatherer: grants}
}

func (authoring *Authoring) List(ctx context.Context, actor domain.Actor, input AuthoringListInput) (domain.AuthoringPage, error) {
	if err := actor.CanCreate(); err != nil {
		return domain.AuthoringPage{}, err
	}
	query, err := domain.NewQuery(input.Limit, input.Offset, input.Search, string(domain.SortRecent))
	if err != nil {
		return domain.AuthoringPage{}, domain.ErrInvalidRequest
	}
	if err := domain.ValidateStatusFilter(input.Status); err != nil {
		return domain.AuthoringPage{}, err
	}

	callCtx, cancel := context.WithTimeout(ctx, authoringOperationTimeout)
	defer cancel()
	page, err := authoring.repository.List(callCtx, actor, query, input.Status)
	return page, operationError(ctx, err)
}

func (authoring *Authoring) Load(ctx context.Context, actor domain.Actor, id string) (domain.CourseTree, error) {
	if err := actor.CanCreate(); err != nil {
		return domain.CourseTree{}, err
	}
	if err := domain.ValidateID(id); err != nil {
		return domain.CourseTree{}, err
	}

	callCtx, cancel := context.WithTimeout(ctx, authoringOperationTimeout)
	defer cancel()
	tree, err := authoring.repository.Load(callCtx, id)
	if err != nil {
		return domain.CourseTree{}, operationError(ctx, err)
	}
	if err := domain.ValidateCourseTree(tree); err != nil {
		return domain.CourseTree{}, domain.ErrDependencyUnavailable
	}
	if err := actor.CanManage(tree.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	return tree, nil
}

func (authoring *Authoring) Create(ctx context.Context, actor domain.Actor, draft domain.CourseDraft) (domain.CourseTree, error) {
	if err := actor.CanCreate(); err != nil {
		return domain.CourseTree{}, err
	}
	if err := actor.CanAssignOwner(draft.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	if actor.Role == domain.ActorInstructor {
		draft.InstructorID = actor.UserID
	}
	if err := domain.ValidateCourseDraft(draft); err != nil {
		return domain.CourseTree{}, err
	}

	callCtx, cancel := context.WithTimeout(ctx, authoringOperationTimeout)
	defer cancel()
	tree, err := authoring.repository.Create(callCtx, actor, draft, authoring.timestamp())
	if err != nil {
		return domain.CourseTree{}, operationError(ctx, err)
	}
	return authoring.validateStoredTree(tree)
}

func (authoring *Authoring) Save(ctx context.Context, actor domain.Actor, tree domain.CourseTree, expectedUpdatedAt time.Time) (domain.CourseTree, error) {
	if err := actor.CanCreate(); err != nil {
		return domain.CourseTree{}, err
	}
	if err := actor.CanAssignOwner(tree.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	if actor.Role == domain.ActorInstructor && tree.InstructorID == "" {
		tree.InstructorID = actor.UserID
	}
	if err := domain.ValidateCourseTreeEdit(tree); err != nil {
		return domain.CourseTree{}, err
	}
	if expectedUpdatedAt.IsZero() {
		return domain.CourseTree{}, domain.ErrInvalidRequest
	}

	callCtx, cancel := context.WithTimeout(ctx, authoringOperationTimeout)
	defer cancel()
	stored, err := authoring.repository.Load(callCtx, tree.ID)
	if err != nil {
		return domain.CourseTree{}, operationError(ctx, err)
	}
	if _, err := authoring.validateStoredTree(stored); err != nil {
		return domain.CourseTree{}, err
	}
	if err := actor.CanManage(stored.InstructorID); err != nil {
		return domain.CourseTree{}, err
	}
	original := stored
	if err := domain.ValidateSlugChange(stored.Slug, stored.PublishedAt, tree.Slug); err != nil {
		return domain.CourseTree{}, err
	}
	stored.Slug = tree.Slug
	stored.Title = tree.Title
	stored.Summary = tree.Summary
	stored.Description = tree.Description
	stored.InstructorID = tree.InstructorID
	if !tree.PriceSet {
		tree.Price = stored.Price
	}
	preserveUnspecifiedPreviews(stored.Modules, tree.Modules)
	firstPriceSet := stored.Price == nil && tree.Price != nil
	requestedPrice := tree.Price
	if !firstPriceSet {
		stored.Price = requestedPrice
	}
	stored.Modules = tree.Modules

	var saved domain.CourseTree
	if firstPriceSet {
		storedForCompare := stored
		storedForCompare.Modules = clearPreviewSet(stored.Modules)
		if !reflect.DeepEqual(original, storedForCompare) {
			return domain.CourseTree{}, domain.ErrInvalidRequest
		}
		saved = original
	} else {
		saved, err = authoring.repository.Save(callCtx, actor, stored, expectedUpdatedAt, authoring.timestampAfter(expectedUpdatedAt))
		if err != nil {
			return domain.CourseTree{}, operationError(ctx, err)
		}
	}
	if firstPriceSet {
		if authoring.grandfatherer == nil {
			return domain.CourseTree{}, domain.ErrDependencyUnavailable
		}
		priceUpdatedAt := authoring.timestampAfter(saved.UpdatedAt)
		if err := authoring.grandfatherer.GrandfatherOnPriceSet(callCtx, saved.ID, saved.UpdatedAt, requestedPrice.Minor, requestedPrice.Currency, priceUpdatedAt); err != nil {
			return domain.CourseTree{}, operationError(ctx, err)
		}
		saved, err = authoring.repository.Load(callCtx, saved.ID)
		if err != nil {
			return domain.CourseTree{}, operationError(ctx, err)
		}
	}
	return authoring.validateStoredTree(saved)
}

func preserveUnspecifiedPreviews(stored, proposed []domain.Module) {
	previews := make(map[string]bool)
	for _, module := range stored {
		for _, lesson := range module.Lessons {
			previews[lesson.ID] = lesson.IsPreview
		}
	}
	for moduleIndex := range proposed {
		for lessonIndex := range proposed[moduleIndex].Lessons {
			lesson := &proposed[moduleIndex].Lessons[lessonIndex]
			if !lesson.PreviewSet {
				lesson.IsPreview = previews[lesson.ID]
			}
		}
	}
}

func clearPreviewSet(modules []domain.Module) []domain.Module {
	result := make([]domain.Module, len(modules))
	for i, module := range modules {
		result[i] = module
		result[i].Lessons = make([]domain.Lesson, len(module.Lessons))
		for j, lesson := range module.Lessons {
			lesson.PreviewSet = false
			result[i].Lessons[j] = lesson
		}
	}
	return result
}

func (authoring *Authoring) Transition(ctx context.Context, actor domain.Actor, id string, to domain.Status, expectedUpdatedAt time.Time) (domain.CourseTree, error) {
	if err := actor.CanCreate(); err != nil {
		return domain.CourseTree{}, err
	}
	if err := domain.ValidateID(id); err != nil || expectedUpdatedAt.IsZero() {
		return domain.CourseTree{}, domain.ErrInvalidRequest
	}

	callCtx, cancel := context.WithTimeout(ctx, authoringOperationTimeout)
	defer cancel()
	stored, err := authoring.repository.Load(callCtx, id)
	if err != nil {
		return domain.CourseTree{}, operationError(ctx, err)
	}
	if _, err := authoring.validateStoredTree(stored); err != nil {
		return domain.CourseTree{}, err
	}
	if err := actor.CanTransition(stored.InstructorID, stored.Status, to); err != nil {
		return domain.CourseTree{}, err
	}

	transitioned, err := authoring.repository.Transition(callCtx, actor, id, stored.Status, to, expectedUpdatedAt, authoring.timestampAfter(expectedUpdatedAt))
	if err != nil {
		return domain.CourseTree{}, operationError(ctx, err)
	}
	return authoring.validateStoredTree(transitioned)
}

func (authoring *Authoring) Delete(ctx context.Context, actor domain.Actor, id string, expectedUpdatedAt time.Time) error {
	if err := actor.CanCreate(); err != nil {
		return err
	}
	if err := domain.ValidateID(id); err != nil || expectedUpdatedAt.IsZero() {
		return domain.ErrInvalidRequest
	}

	callCtx, cancel := context.WithTimeout(ctx, authoringOperationTimeout)
	defer cancel()
	stored, err := authoring.repository.Load(callCtx, id)
	if err != nil {
		return operationError(ctx, err)
	}
	if _, err := authoring.validateStoredTree(stored); err != nil {
		return err
	}
	if err := actor.CanManage(stored.InstructorID); err != nil {
		return err
	}
	if err := domain.ValidateDelete(stored.Status); err != nil {
		return err
	}
	return operationError(ctx, authoring.repository.Delete(callCtx, actor, id, expectedUpdatedAt))
}

func (authoring *Authoring) validateStoredTree(tree domain.CourseTree) (domain.CourseTree, error) {
	if err := domain.ValidateCourseTree(tree); err != nil {
		return domain.CourseTree{}, domain.ErrDependencyUnavailable
	}
	return tree, nil
}

func (authoring *Authoring) timestamp() time.Time {
	return authoring.now().UTC().Truncate(time.Microsecond)
}

func (authoring *Authoring) timestampAfter(expected time.Time) time.Time {
	timestamp := authoring.timestamp()
	if !timestamp.After(expected) {
		return expected.UTC().Truncate(time.Microsecond).Add(time.Microsecond)
	}
	return timestamp
}

func operationError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	return err
}
