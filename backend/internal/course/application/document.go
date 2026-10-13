package application

import (
	"context"
	"errors"
	"time"

	"courses/internal/course/domain"
	"courses/internal/course/ports"
)

type Documents struct {
	repository ports.LessonDocumentRepository
	access     ports.LessonDocumentAccessReader
}

func NewDocuments(repository ports.LessonDocumentRepository, access ports.LessonDocumentAccessReader) *Documents {
	return &Documents{repository: repository, access: access}
}

func (documents *Documents) SaveDraft(ctx context.Context, actor domain.Actor, lessonID, locale string, body domain.LessonDocument, expectedUpdatedAt time.Time) (time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := documents.requireEditor(ctx, actor, lessonID); err != nil {
		return time.Time{}, err
	}
	if err := domain.ValidateLessonDocument(body); err != nil {
		return time.Time{}, err
	}
	updatedAt, err := documents.repository.SaveDraft(ctx, lessonID, locale, body, expectedUpdatedAt)
	return updatedAt, documentWriteError(err)
}

func (documents *Documents) PublishDocument(ctx context.Context, actor domain.Actor, lessonID, locale string, expectedUpdatedAt time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := documents.requireEditor(ctx, actor, lessonID); err != nil {
		return err
	}
	return documentWriteError(documents.repository.PublishDocument(ctx, lessonID, locale, expectedUpdatedAt))
}

func (documents *Documents) GetDraft(ctx context.Context, actor domain.Actor, lessonID, locale string) (domain.LessonDocument, time.Time, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := documents.requireEditor(ctx, actor, lessonID); err != nil {
		return domain.LessonDocument{}, time.Time{}, err
	}
	body, updatedAt, found, err := documents.repository.LoadDraft(ctx, lessonID, locale)
	if err != nil {
		return domain.LessonDocument{}, time.Time{}, err
	}
	if !found {
		return emptyLessonDocument(locale), time.Time{}, nil
	}
	return body, updatedAt, nil
}

func (documents *Documents) GetPublishedDocument(ctx context.Context, actor domain.Actor, lessonID, locale string) (domain.LessonDocument, error) {
	if !actor.EmailVerified {
		return domain.LessonDocument{}, domain.ErrEmailNotVerified
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	facts, err := documents.access.LoadLessonAccess(ctx, actor.UserID, lessonID)
	if err != nil {
		return domain.LessonDocument{}, err
	}
	// Gate before loading content: absence must not expose a different access signal.
	if !facts.CanConsumeBy(actor) {
		if facts.PurchaseRequiredBy(actor) {
			return domain.LessonDocument{}, domain.ErrPurchaseRequired
		}
		return domain.LessonDocument{}, domain.ErrForbidden
	}
	body, found, err := documents.repository.LoadPublished(ctx, lessonID, locale)
	if err != nil {
		return domain.LessonDocument{}, err
	}
	if !found {
		return emptyLessonDocument(locale), nil
	}
	return body, nil
}

func (documents *Documents) requireEditor(ctx context.Context, actor domain.Actor, lessonID string) error {
	facts, err := documents.access.LoadLessonAccess(ctx, actor.UserID, lessonID)
	if err != nil {
		return err
	}
	if !facts.CanBeEditedBy(actor) {
		return domain.ErrForbidden
	}
	return nil
}

func emptyLessonDocument(locale string) domain.LessonDocument {
	return domain.LessonDocument{SchemaVersion: domain.CurrentLessonDocumentSchemaVersion, Locale: locale, Sections: []domain.Section{}}
}

func documentWriteError(err error) error {
	if errors.Is(err, domain.ErrDocumentConflict) {
		return domain.ErrDocumentConflict
	}
	return err
}
