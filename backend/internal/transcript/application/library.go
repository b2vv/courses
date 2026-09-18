package application

import (
	"context"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

type Library struct{ repository ports.TranscriptRepository }

func NewLibrary(repository ports.TranscriptRepository) *Library {
	return &Library{repository: repository}
}

func (library *Library) course(ctx context.Context, actor domain.Actor, courseID string) (domain.Course, error) {
	if err := domain.ValidateID(courseID); err != nil {
		return domain.Course{}, err
	}
	if err := domain.ValidateID(actor.UserID); err != nil {
		return domain.Course{}, domain.ErrForbidden
	}
	return library.repository.CourseOfID(ctx, courseID)
}

// List is readable by whoever may edit, not by whoever may watch: the library is an authoring
// surface, and a student never picks documents.
func (library *Library) List(ctx context.Context, actor domain.Actor, courseID string) ([]domain.DocumentLink, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	course, err := library.course(ctx, actor, courseID)
	if err != nil {
		return nil, err
	}
	if !course.CanBeEditedBy(actor) {
		return nil, domain.ErrForbidden
	}
	return library.repository.CourseDocuments(ctx, courseID)
}

func (library *Library) Add(ctx context.Context, actor domain.Actor, courseID string, input domain.DocumentLink) error {
	// Validated before any authorisation round trip and, more importantly, before the database:
	// A3 lives at the entry to the library precisely so that no segment can reference a link that
	// skipped it.
	link, err := domain.NewDocumentLink(input)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	course, err := library.course(ctx, actor, courseID)
	if err != nil {
		return err
	}
	if !course.CanBeEditedBy(actor) {
		return domain.ErrForbidden
	}
	return library.repository.AddCourseDocument(ctx, courseID, link)
}

func (library *Library) Remove(ctx context.Context, actor domain.Actor, courseID, documentID string) error {
	if err := domain.ValidateID(documentID); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	course, err := library.course(ctx, actor, courseID)
	if err != nil {
		return err
	}
	if !course.CanBeEditedBy(actor) {
		return domain.ErrForbidden
	}
	return library.repository.RemoveCourseDocument(ctx, courseID, documentID)
}
