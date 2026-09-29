package application

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

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

type ModerationComment struct {
	ID          string
	LessonID    string
	LessonTitle string
	ParentID    *string
	Author      *string
	Body        *string
	PositionMs  *int64
	CreatedAt   time.Time
	EditedAt    *time.Time
	Removed     bool
}

type CourseCommentPage struct {
	Comments   []ModerationComment
	NextCursor string
}

type QueueItem struct {
	Comment     ModerationComment
	OpenReports int
}

type QueuePage struct {
	Items      []QueueItem
	NextCursor string
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
	if err := service.repository.Remove(ctx, commentID, actor.UserID, service.now()); err != nil {
		return err
	}
	if domain.IsModeration(actor, record.Comment) {
		return service.repository.WriteModerationLog(ctx, ports.ModerationLogEntry{
			CourseID: record.CourseID, CommentID: commentID, ActorID: actor.UserID,
			ActorRole: actorRole(actor), Action: "remove", CreatedAt: service.now(),
		})
	}
	return nil
}

func (service *Service) ReportComment(ctx context.Context, actor Actor, commentID string) error {
	record, err := service.repository.Get(ctx, commentID)
	if err != nil {
		return err
	}
	if !record.Course.CanBeViewedBy(actor) {
		return domain.ErrForbidden
	}
	if record.Comment.RemovedAt != nil {
		return nil
	}
	report, err := domain.NewReport(domain.Report{CommentID: commentID, ReporterID: actor.UserID}, record.Comment)
	if err != nil {
		return err
	}
	return service.repository.InsertReport(ctx, report, service.now())
}

func (service *Service) ListCourseComments(ctx context.Context, actor Actor, courseID, cursor string) (CourseCommentPage, error) {
	if err := service.requireModerationAccess(ctx, actor, courseID); err != nil {
		return CourseCommentPage{}, err
	}
	page, err := service.repository.ListByCourse(ctx, courseID, cursor, 50)
	if err != nil {
		return CourseCommentPage{}, err
	}
	result := CourseCommentPage{NextCursor: page.NextCursor, Comments: make([]ModerationComment, len(page.Comments))}
	for i, record := range page.Comments {
		result.Comments[i] = service.presentModeration(record)
	}
	return result, nil
}

func (service *Service) Queue(ctx context.Context, actor Actor, courseID, cursor string) (QueuePage, error) {
	if err := service.requireModerationAccess(ctx, actor, courseID); err != nil {
		return QueuePage{}, err
	}
	page, err := service.repository.ListQueue(ctx, courseID, cursor, 50)
	if err != nil {
		return QueuePage{}, err
	}
	result := QueuePage{NextCursor: page.NextCursor, Items: make([]QueueItem, len(page.Items))}
	for i, item := range page.Items {
		result.Items[i] = QueueItem{Comment: service.presentModeration(item.ModerationCommentRecord), OpenReports: item.OpenReports}
	}
	return result, nil
}

func (service *Service) DismissReports(ctx context.Context, actor Actor, commentID, reason string) error {
	record, err := service.repository.Get(ctx, commentID)
	if err != nil {
		return err
	}
	if !record.Course.CanBeModeratedBy(actor) {
		return domain.ErrForbidden
	}
	reason = strings.TrimSpace(reason)
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 1 || utf8.RuneCountInString(reason) > 500 {
		return domain.ErrInvalidRequest
	}
	if err := service.repository.DismissReports(ctx, commentID, actor.UserID, service.now()); err != nil {
		return err
	}
	return service.repository.WriteModerationLog(ctx, ports.ModerationLogEntry{
		CourseID: record.CourseID, CommentID: commentID, ActorID: actor.UserID,
		ActorRole: actorRole(actor), Action: "dismiss", Reason: &reason, CreatedAt: service.now(),
	})
}

func (service *Service) requireModerationAccess(ctx context.Context, actor Actor, courseID string) error {
	facts, err := service.repository.CourseFacts(ctx, courseID)
	if err != nil {
		return err
	}
	if !facts.Course.CanBeModeratedBy(actor) {
		return domain.ErrForbidden
	}
	return nil
}

func actorRole(actor Actor) string {
	if actor.Administrator {
		return "admin"
	}
	return "instructor"
}

func (service *Service) presentModeration(record ports.ModerationCommentRecord) ModerationComment {
	result := ModerationComment{ID: record.Comment.ID, LessonID: record.Comment.LessonID, LessonTitle: record.LessonTitle, ParentID: record.Comment.ParentID,
		Author: record.DisplayName, PositionMs: record.Comment.PositionMs, CreatedAt: record.Comment.CreatedAt,
		EditedAt: record.Comment.EditedAt, Removed: record.Comment.RemovedAt != nil}
	if !result.Removed {
		body := record.Comment.Body
		result.Body = &body
	}
	return result
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
