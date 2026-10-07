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
	ErrEmailNotVerified = errors.New("email_not_verified")
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
	Kind        string
	Comment     *ModerationComment
	Review      *Review
	OpenReports int
}

type QueuePage struct {
	Items      []QueueItem
	NextCursor string
}

type Rating struct {
	Average float64
	Count   int
}
type Review struct {
	ID                string
	AuthorDisplayName *string
	Stars             int
	Body              string
	CompletedLessons  int
	TotalLessons      int
	CreatedAt         time.Time
	EditedAt          *time.Time
}
type PublicReviewPage struct {
	Rating     *Rating
	Items      []Review
	NextCursor string
}
type OwnReview struct {
	Review      *Review
	Eligibility string
}

type ParticipantReportQueueItem struct {
	ReporterName *string
	TargetName   *string
	CourseTitle  string
	Report       domain.ParticipantReport
}

type ParticipantReportQueuePage struct {
	Items      []ParticipantReportQueueItem
	NextCursor string
}

func (service *Service) ReportLecturer(ctx context.Context, actor Actor, courseID, reason string) (domain.ParticipantReport, error) {
	if !actor.EmailVerified {
		return domain.ParticipantReport{}, ErrEmailNotVerified
	}
	ownerID, membership, courseFound, err := service.repository.ReportSubject(ctx, courseID, actor.UserID)
	if err != nil {
		return domain.ParticipantReport{}, err
	}
	if !courseFound {
		return domain.ParticipantReport{}, domain.ErrNotFound
	}
	if actor.UserID == ownerID {
		return domain.ParticipantReport{}, domain.ErrReportSelf
	}
	if !domain.IsParticipant(membership) {
		return domain.ParticipantReport{}, domain.ErrReportForbidden
	}
	report, err := domain.NewParticipantReport("", courseID, actor.UserID, ownerID, domain.ReportOnLecturer, reason, service.now())
	if err != nil {
		return domain.ParticipantReport{}, err
	}
	id, err := service.repository.InsertParticipantReport(ctx, report)
	if err != nil {
		return domain.ParticipantReport{}, err
	}
	report.ID = id
	return report, nil
}

func (service *Service) ReportStudent(ctx context.Context, actor Actor, courseID, targetID, reason string) (domain.ParticipantReport, error) {
	if !actor.EmailVerified {
		return domain.ParticipantReport{}, ErrEmailNotVerified
	}
	ownerID, membership, courseFound, err := service.repository.ReportSubject(ctx, courseID, targetID)
	if err != nil {
		return domain.ParticipantReport{}, err
	}
	if !courseFound {
		return domain.ParticipantReport{}, domain.ErrNotFound
	}
	if !actor.Administrator && actor.UserID != ownerID {
		return domain.ParticipantReport{}, domain.ErrReportForbidden
	}
	if !domain.IsParticipant(membership) {
		return domain.ParticipantReport{}, domain.ErrReportForbidden
	}
	report, err := domain.NewParticipantReport("", courseID, actor.UserID, targetID, domain.ReportOnStudent, reason, service.now())
	if err != nil {
		return domain.ParticipantReport{}, err
	}
	id, err := service.repository.InsertParticipantReport(ctx, report)
	if err != nil {
		return domain.ParticipantReport{}, err
	}
	report.ID = id
	return report, nil
}

func (service *Service) ReportQueue(ctx context.Context, actor Actor, cursor string) (ParticipantReportQueuePage, error) {
	if !actor.Administrator {
		return ParticipantReportQueuePage{}, domain.ErrReportForbidden
	}
	page, err := service.repository.ParticipantReportQueue(ctx, cursor, 20)
	if err != nil {
		return ParticipantReportQueuePage{}, err
	}
	result := ParticipantReportQueuePage{NextCursor: page.NextCursor, Items: make([]ParticipantReportQueueItem, len(page.Items))}
	for i, item := range page.Items {
		result.Items[i] = ParticipantReportQueueItem{Report: item.Report, ReporterName: item.ReporterName, TargetName: item.TargetName, CourseTitle: item.CourseTitle}
	}
	return result, nil
}

func (service *Service) ResolveReport(ctx context.Context, actor Actor, reportID string) error {
	if !actor.Administrator {
		return domain.ErrReportForbidden
	}
	return service.repository.ResolveParticipantReport(ctx, reportID)
}

func (service *Service) List(ctx context.Context, actor Actor, lessonID, cursor string) (Page, error) {
	facts, err := service.repository.LessonFacts(ctx, actor.UserID, lessonID)
	if err != nil {
		return Page{}, err
	}
	if !facts.Course.CanConsumeBy(actor) {
		if facts.Course.PurchaseRequiredBy(actor) {
			return Page{}, domain.ErrPurchaseRequired
		}
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
	if !actor.EmailVerified {
		return Comment{}, ErrEmailNotVerified
	}
	facts, err := service.repository.LessonFacts(ctx, actor.UserID, request.LessonID)
	if err != nil {
		return Comment{}, err
	}
	if err := requireCommunication(facts.Course, actor); err != nil {
		return Comment{}, err
	}
	var parent *domain.Comment
	if request.ParentID != nil {
		parentRecord, getErr := service.repository.Get(ctx, actor.UserID, *request.ParentID)
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
		fresh, getErr := service.repository.Get(ctx, actor.UserID, record.Comment.ID)
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
	record, err := service.repository.Get(ctx, actor.UserID, commentID)
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
	record, err := service.repository.Get(ctx, actor.UserID, commentID)
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
	if !actor.EmailVerified {
		return ErrEmailNotVerified
	}
	record, err := service.repository.Get(ctx, actor.UserID, commentID)
	if err != nil {
		return err
	}
	if err := requireCommunication(record.Course, actor); err != nil {
		return err
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
	page, err := service.repository.ListModerationQueue(ctx, courseID, cursor, 50)
	if err != nil {
		return QueuePage{}, err
	}
	result := QueuePage{NextCursor: page.NextCursor, Items: make([]QueueItem, len(page.Items))}
	for i, item := range page.Items {
		result.Items[i] = QueueItem{Kind: item.Kind, OpenReports: item.OpenReports}
		if item.Kind == "review" {
			review := service.presentReview(ports.ReviewRecord{ID: item.ID, DisplayName: item.DisplayName, Stars: *item.Stars, Body: item.Body, CreatedAt: item.CreatedAt})
			result.Items[i].Review = &review
		} else {
			comment := ModerationComment{ID: item.ID, LessonID: dereference(item.LessonID), LessonTitle: dereference(item.LessonTitle), Author: item.DisplayName, Body: &item.Body, CreatedAt: item.CreatedAt}
			result.Items[i].Comment = &comment
		}
	}
	return result, nil
}

func (service *Service) PublicReviews(ctx context.Context, slug, cursor string) (PublicReviewPage, error) {
	courseID, err := service.repository.PublishedReviewCourse(ctx, slug)
	if err != nil {
		return PublicReviewPage{}, err
	}
	return service.listReviews(ctx, courseID, cursor)
}

func (service *Service) LearningReviews(ctx context.Context, actor Actor, courseID, cursor string) (PublicReviewPage, error) {
	facts, err := service.repository.ReviewFacts(ctx, courseID, actor.UserID)
	if err != nil {
		return PublicReviewPage{}, err
	}
	// Reading the in-course review list is consumption (plan §4): a member — active or blocked —
	// reads it like lesson comments and content; a non-member does not. (PublicReviews stays CanView.)
	course := domain.Course{OwnerID: facts.Facts.OwnerID, Status: facts.Facts.Status, Paid: facts.Facts.Paid, Entitled: facts.Facts.Entitled, Membership: facts.Facts.Membership}
	if !course.CanConsumeBy(actor) {
		return PublicReviewPage{}, domain.ErrForbidden
	}
	return service.listReviews(ctx, courseID, cursor)
}

func (service *Service) listReviews(ctx context.Context, courseID, cursor string) (PublicReviewPage, error) {
	count, sum, err := service.repository.ReviewSummary(ctx, courseID)
	if err != nil {
		return PublicReviewPage{}, err
	}
	page, err := service.repository.ListReviews(ctx, courseID, cursor, 20)
	if err != nil {
		return PublicReviewPage{}, err
	}
	result := PublicReviewPage{Items: make([]Review, len(page.Items)), NextCursor: page.NextCursor}
	if summary := domain.NewRatingSummary(count, sum); summary != nil {
		result.Rating = &Rating{Average: summary.Average, Count: summary.Count}
	}
	for i, item := range page.Items {
		result.Items[i] = service.presentReview(item)
	}
	return result, nil
}

func (service *Service) OwnReview(ctx context.Context, actor Actor, courseID string) (OwnReview, error) {
	facts, err := service.repository.ReviewFacts(ctx, courseID, actor.UserID)
	if err != nil {
		return OwnReview{}, err
	}
	review, err := service.repository.GetOwnReview(ctx, courseID, actor.UserID)
	if err != nil {
		return OwnReview{}, err
	}
	result := OwnReview{Eligibility: "ok"}
	if err := domain.CanReview(facts.Facts, actor); err != nil {
		result.Eligibility = err.Error()
	}
	if review != nil && review.RemovedAt != nil {
		// A review removed by moderation is not the author's live review: they see why, and cannot write a
		// replacement (AC7). Decision 2026-10-04, docs/ratings/assumptions.md.
		result.Eligibility = domain.ErrReviewRemoved.Error()
		return result, nil
	}
	if review != nil {
		rendered := service.presentReview(*review)
		result.Review = &rendered
	}
	return result, nil
}

func (service *Service) PutReview(ctx context.Context, actor Actor, courseID string, stars int, body string) (Review, bool, error) {
	if !actor.EmailVerified {
		return Review{}, false, ErrEmailNotVerified
	}
	facts, err := service.repository.ReviewFacts(ctx, courseID, actor.UserID)
	if err != nil {
		return Review{}, false, err
	}
	course := domain.Course{OwnerID: facts.Facts.OwnerID, Status: facts.Facts.Status, Paid: facts.Facts.Paid, Entitled: facts.Facts.Entitled, Membership: facts.Facts.Membership}
	if err := requireCommunication(course, actor); err != nil {
		return Review{}, false, err
	}
	if err := domain.CanReview(facts.Facts, actor); err != nil {
		return Review{}, false, err
	}
	content, err := domain.NewReviewContent(stars, body)
	if err != nil {
		return Review{}, false, err
	}
	progress, err := domain.NewReviewProgress(facts.CompletedLessons, facts.TotalLessons)
	if err != nil {
		return Review{}, false, err
	}
	if facts.Removed {
		return Review{}, false, domain.ErrReviewRemoved
	}
	var stored ports.ReviewRecord
	created := !facts.Exists
	if created {
		stored, err = service.repository.CreateReview(ctx, ports.ReviewRecord{CourseID: courseID, AuthorID: actor.UserID, Stars: content.Stars, Body: content.Body, CompletedLessons: progress.Completed, TotalLessons: progress.Total, CreatedAt: service.now()})
	} else {
		stored, err = service.repository.UpdateReview(ctx, courseID, actor.UserID, content, progress, service.now())
	}
	if err != nil {
		return Review{}, false, err
	}
	return service.presentReview(stored), created, nil
}

func (service *Service) DeleteReview(ctx context.Context, actor Actor, courseID string) error {
	return service.repository.DeleteReview(ctx, courseID, actor.UserID)
}
func (service *Service) ReportReview(ctx context.Context, actor Actor, reviewID string) error {
	if !actor.EmailVerified {
		return ErrEmailNotVerified
	}
	review, err := service.repository.GetReview(ctx, reviewID)
	if err != nil {
		return err
	}
	facts, err := service.repository.ReviewFacts(ctx, review.CourseID, actor.UserID)
	if err != nil {
		return err
	}
	course := domain.Course{OwnerID: facts.Facts.OwnerID, Status: facts.Facts.Status, Paid: facts.Facts.Paid, Entitled: facts.Facts.Entitled, Membership: facts.Facts.Membership}
	if err := requireCommunication(course, actor); err != nil {
		return err
	}
	return service.repository.InsertReviewReport(ctx, reviewID, actor.UserID, service.now())
}

func requireCommunication(course domain.Course, actor Actor) error {
	if course.CanCommunicateBy(actor) {
		return nil
	}
	if course.CommunicationBlocked() {
		return domain.ErrEnrollmentBlocked
	}
	return domain.ErrEnrollmentRequired
}

func (service *Service) RemoveReview(ctx context.Context, actor Actor, reviewID, reason string) error {
	if !actor.Administrator {
		return domain.ErrReviewModerationForbidden
	}
	review, err := service.repository.GetReview(ctx, reviewID)
	if err != nil {
		return err
	}
	if review.RemovedAt != nil {
		return nil
	}
	reason = strings.TrimSpace(reason)
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 1 || utf8.RuneCountInString(reason) > 500 {
		return domain.ErrInvalidRequest
	}
	return service.repository.RemoveReview(ctx, reviewID, actor.UserID, "admin", &reason, service.now())
}

func (service *Service) DismissReviewReports(ctx context.Context, actor Actor, reviewID, reason string) error {
	review, err := service.repository.GetReview(ctx, reviewID)
	if err != nil {
		return err
	}
	if err := service.requireModerationAccess(ctx, actor, review.CourseID); err != nil {
		return err
	}
	reason = strings.TrimSpace(reason)
	if !utf8.ValidString(reason) || utf8.RuneCountInString(reason) < 1 || utf8.RuneCountInString(reason) > 500 {
		return domain.ErrInvalidRequest
	}
	if err := service.repository.DismissReviewReports(ctx, reviewID, actor.UserID, service.now()); err != nil {
		return err
	}
	return service.repository.WriteModerationLog(ctx, ports.ModerationLogEntry{CourseID: review.CourseID, ReviewID: reviewID, ActorID: actor.UserID, ActorRole: actorRole(actor), Action: "dismiss", Reason: &reason, CreatedAt: service.now()})
}

func (service *Service) presentReview(review ports.ReviewRecord) Review {
	return Review{ID: review.ID, AuthorDisplayName: review.DisplayName, Stars: review.Stars, Body: review.Body, CompletedLessons: review.CompletedLessons, TotalLessons: review.TotalLessons, CreatedAt: review.CreatedAt, EditedAt: review.EditedAt}
}
func dereference(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (service *Service) DismissReports(ctx context.Context, actor Actor, commentID, reason string) error {
	record, err := service.repository.Get(ctx, actor.UserID, commentID)
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
