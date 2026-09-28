package application

import (
	"context"
	"errors"
	"time"

	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
)

var (
	ErrCommentForbidden = errors.New("comment_forbidden")
	ErrCommentRemoved   = errors.New("comment_removed")
)

type Service struct {
	repository ports.Repository
	now        func() time.Time
}

func New(repository ports.Repository, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{repository: repository, now: now}
}

type Actor = domain.Actor

type Page struct {
	Threads    []Thread
	NextCursor string
}

type Thread struct {
	Root    Comment
	Replies []Comment
}

type Comment struct {
	ID         string
	Author     *string
	Body       *string
	PositionMs *int64
	CreatedAt  time.Time
	EditedAt   *time.Time
	Removed    bool
	CanEdit    bool
	CanRemove  bool
}

type PublishRequest struct {
	LessonID   string
	Body       string
	PositionMs *int64
	ParentID   *string
}

func (service *Service) List(ctx context.Context, actor Actor, lessonID, cursor string) (Page, error) {
	facts, err := service.repository.LessonFacts(ctx, lessonID)
	if err != nil {
		return Page{}, err
	}
	if !facts.Course.CanBeViewedBy(actor) {
		return Page{}, domain.ErrForbidden
	}
	page, err := service.repository.List(ctx, lessonID, cursor, 50)
	if err != nil {
		return Page{}, err
	}
	result := Page{NextCursor: page.NextCursor, Threads: make([]Thread, len(page.Threads))}
	for i, thread := range page.Threads {
		result.Threads[i] = Thread{
			Root:    service.present(thread.Root, actor),
			Replies: make([]Comment, len(thread.Replies)),
		}
		for j, reply := range thread.Replies {
			result.Threads[i].Replies[j] = service.present(reply, actor)
		}
	}
	return result, nil
}

func (service *Service) Publish(ctx context.Context, actor Actor, request PublishRequest) (Comment, error) {
	facts, err := service.repository.LessonFacts(ctx, request.LessonID)
	if err != nil {
		return Comment{}, err
	}
	if !facts.Course.CanBeViewedBy(actor) {
		return Comment{}, domain.ErrForbidden
	}
	var parent *domain.Comment
	if request.ParentID != nil {
		parentRecord, getErr := service.repository.Get(ctx, *request.ParentID)
		if getErr == nil {
			parentComment := parentRecord.Comment
			parent = &parentComment
		} else if errors.Is(getErr, domain.ErrNotFound) {
			return Comment{}, domain.ErrParentInvalid
		} else {
			return Comment{}, getErr
		}
	}
	comment, err := domain.NewComment(domain.Comment{
		LessonID: request.LessonID, AuthorID: actor.UserID, Body: request.Body,
		PositionMs: request.PositionMs, ParentID: request.ParentID,
	}, parent, facts.DurationSecs)
	if err != nil {
		return Comment{}, err
	}
	var record ports.CommentRecord
	if request.ParentID == nil {
		record, err = service.repository.InsertRoot(ctx, comment)
	} else {
		record, err = service.repository.InsertReply(ctx, comment)
	}
	if err != nil {
		return Comment{}, err
	}
	if record.Comment.ID != "" {
		fresh, getErr := service.repository.Get(ctx, record.Comment.ID)
		if getErr != nil {
			return Comment{}, getErr
		}
		record = fresh
	}
	if record.Course.OwnerID == "" {
		record.Course = facts.Course
	}
	return service.present(record, actor), nil
}

func (service *Service) Update(ctx context.Context, actor Actor, commentID, body string) (Comment, error) {
	record, err := service.repository.Get(ctx, commentID)
	if err != nil {
		return Comment{}, err
	}
	if !record.Comment.CanBeEditedBy(actor) {
		return Comment{}, ErrCommentForbidden
	}
	if record.Comment.RemovedAt != nil {
		return Comment{}, ErrCommentRemoved
	}
	validated, err := domain.NewComment(domain.Comment{LessonID: record.Comment.LessonID, Body: body}, nil, nil)
	if err != nil {
		return Comment{}, err
	}
	updated, err := service.repository.UpdateBody(ctx, commentID, actor.UserID, validated.Body, service.now())
	if err != nil {
		return Comment{}, err
	}
	if updated.Course.OwnerID == "" {
		updated.Course = record.Course
	}
	return service.present(updated, actor), nil
}

func (service *Service) Remove(ctx context.Context, actor Actor, commentID string) error {
	record, err := service.repository.Get(ctx, commentID)
	if err != nil {
		return err
	}
	if !record.Comment.CanBeRemovedBy(actor, record.Course) {
		return ErrCommentForbidden
	}
	if record.Comment.RemovedAt != nil {
		return nil
	}
	return service.repository.Remove(ctx, commentID, service.now())
}

func (service *Service) present(record ports.CommentRecord, actor Actor) Comment {
	result := Comment{
		ID: record.Comment.ID, Author: record.DisplayName, PositionMs: record.Comment.PositionMs,
		CreatedAt: record.Comment.CreatedAt, EditedAt: record.Comment.EditedAt,
		Removed:   record.Comment.RemovedAt != nil,
		CanEdit:   record.Comment.RemovedAt == nil && record.Comment.CanBeEditedBy(actor),
		CanRemove: record.Comment.RemovedAt == nil && record.Comment.CanBeRemovedBy(actor, record.Course),
	}
	if !result.Removed {
		body := record.Comment.Body
		result.Body = &body
	}
	return result
}
