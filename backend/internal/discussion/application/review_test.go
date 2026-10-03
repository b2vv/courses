package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
)

const reviewCourse = "10000000-0000-4000-8000-000000000001"
const reviewAuthor = "40000000-0000-4000-8000-000000000001"
const reviewOwner = "40000000-0000-4000-8000-000000000002"
const reviewID = "70000000-0000-4000-8000-000000000001"

type reviewFake struct {
	facts                    ports.ReviewFactsRecord
	own                      *ports.ReviewRecord
	byID                     map[string]ports.ReviewRecord
	publicID                 string
	summaryCount, summarySum int
	page                     ports.ReviewPage
	created, updated         int
	lastProgress             domain.ReviewProgress
	deleted                  bool
	reportErr                error
	dismisses                int
	removed                  int
	course                   domain.Course
}

func (f *reviewFake) PublishedReviewCourse(context.Context, string) (string, error) {
	if f.publicID == "" {
		return "", domain.ErrNotFound
	}
	return f.publicID, nil
}
func (f *reviewFake) ReviewFacts(context.Context, string, string) (ports.ReviewFactsRecord, error) {
	return f.facts, nil
}
func (f *reviewFake) GetReview(_ context.Context, id string) (ports.ReviewRecord, error) {
	r, ok := f.byID[id]
	if !ok {
		return ports.ReviewRecord{}, domain.ErrNotFound
	}
	return r, nil
}
func (f *reviewFake) CreateReview(_ context.Context, r ports.ReviewRecord) (ports.ReviewRecord, error) {
	f.created++
	f.own = &r
	f.facts.Exists = true
	return r, nil
}
func (f *reviewFake) UpdateReview(_ context.Context, _, _ string, c domain.ReviewContent, p domain.ReviewProgress, at time.Time) (ports.ReviewRecord, error) {
	f.updated++
	f.lastProgress = p
	r := *f.own
	r.Stars = c.Stars
	r.Body = c.Body
	r.CompletedLessons = p.Completed
	r.TotalLessons = p.Total
	r.EditedAt = &at
	f.own = &r
	return r, nil
}
func (f *reviewFake) DeleteReview(context.Context, string, string) error {
	if f.own == nil {
		return domain.ErrNotFound
	}
	if f.own.RemovedAt != nil {
		return domain.ErrReviewRemoved
	}
	f.deleted = true
	f.own = nil
	return nil
}
func (f *reviewFake) GetOwnReview(context.Context, string, string) (*ports.ReviewRecord, error) {
	return f.own, nil
}
func (f *reviewFake) ListReviews(context.Context, string, string, int) (ports.ReviewPage, error) {
	return f.page, nil
}
func (f *reviewFake) ReviewSummary(context.Context, string) (int, int, error) {
	return f.summaryCount, f.summarySum, nil
}
func (f *reviewFake) InsertReviewReport(context.Context, string, string, time.Time) error {
	return f.reportErr
}
func (f *reviewFake) DismissReviewReports(context.Context, string, string, time.Time) error {
	f.dismisses++
	return nil
}
func (f *reviewFake) RemoveReview(context.Context, string, string, string, *string, time.Time) error {
	f.removed++
	r := f.byID[reviewID]
	now := time.Now()
	r.RemovedAt = &now
	f.byID[reviewID] = r
	return nil
}
func (f *reviewFake) ListModerationQueue(context.Context, string, string, int) (ports.ModerationQueuePage, error) {
	return ports.ModerationQueuePage{}, nil
}
func (f *reviewFake) CourseFacts(context.Context, string) (ports.CourseFacts, error) {
	return ports.CourseFacts{Course: f.course, CourseID: reviewCourse}, nil
}
func (f *reviewFake) ListByCourse(context.Context, string, string, int) (ports.CommentPage, error) {
	return ports.CommentPage{}, nil
}
func (f *reviewFake) InsertReport(context.Context, domain.Report, time.Time) error    { return nil }
func (f *reviewFake) DismissReports(context.Context, string, string, time.Time) error { return nil }
func (f *reviewFake) ListQueue(context.Context, string, string, int) (ports.QueuePage, error) {
	return ports.QueuePage{}, nil
}
func (f *reviewFake) WriteModerationLog(context.Context, ports.ModerationLogEntry) error { return nil }
func (f *reviewFake) LessonFacts(context.Context, string, string) (ports.LessonFacts, error) {
	return ports.LessonFacts{}, nil
}
func (f *reviewFake) List(context.Context, string, string, int) (ports.CommentPage, error) {
	return ports.CommentPage{}, nil
}
func (f *reviewFake) Get(context.Context, string, string) (ports.CommentRecord, error) {
	return ports.CommentRecord{}, domain.ErrNotFound
}
func (f *reviewFake) InsertRoot(context.Context, domain.Comment) (ports.CommentRecord, error) {
	return ports.CommentRecord{}, nil
}
func (f *reviewFake) InsertReply(context.Context, domain.Comment) (ports.CommentRecord, error) {
	return ports.CommentRecord{}, nil
}
func (f *reviewFake) UpdateBody(context.Context, string, string, string, time.Time) (ports.CommentRecord, error) {
	return ports.CommentRecord{}, nil
}
func (f *reviewFake) Remove(context.Context, string, string, time.Time) error { return nil }

func newReviewFake() *reviewFake {
	r := ports.ReviewRecord{ID: reviewID, CourseID: reviewCourse, AuthorID: reviewAuthor, Stars: 4, Body: "this review body has enough characters", CreatedAt: time.Now()}
	return &reviewFake{facts: ports.ReviewFactsRecord{Facts: domain.ReviewFacts{Status: domain.Published, HasProgress: true}, CompletedLessons: 1, TotalLessons: 4}, byID: map[string]ports.ReviewRecord{reviewID: r}, course: domain.Course{OwnerID: reviewOwner, Status: domain.Published}}
}
func reviewService(f *reviewFake) *Service {
	return New(f, func() time.Time { return time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC) })
}

func TestPutReviewEligibilityOrder(t *testing.T) {
	cases := []struct {
		name  string
		facts domain.ReviewFacts
		actor Actor
		want  error
	}{{"anonymous", domain.ReviewFacts{Status: domain.Published, HasProgress: true}, Actor{}, domain.ErrForbidden}, {"owner", domain.ReviewFacts{OwnerID: reviewOwner, Status: domain.Published, HasProgress: true}, Actor{UserID: reviewOwner}, domain.ErrReviewOwnerForbidden}, {"unavailable", domain.ReviewFacts{Status: domain.Archived, HasProgress: true}, Actor{UserID: reviewAuthor}, domain.ErrReviewNotAvailable}, {"purchase", domain.ReviewFacts{Status: domain.Published, Paid: true, HasProgress: true}, Actor{UserID: reviewAuthor}, domain.ErrReviewPurchaseRequired}, {"not started", domain.ReviewFacts{Status: domain.Published}, Actor{UserID: reviewAuthor}, domain.ErrReviewNotStarted}}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := newReviewFake()
			f.facts.Facts = c.facts
			_, _, err := reviewService(f).PutReview(context.Background(), c.actor, reviewCourse, 4, "this review body has enough characters")
			if !errors.Is(err, c.want) {
				t.Fatalf("err=%v want %v", err, c.want)
			}
		})
	}
}
func TestPutReviewCreatesThenUpdatesRecapturedProgress(t *testing.T) {
	f := newReviewFake()
	s := reviewService(f)
	_, created, err := s.PutReview(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse, 4, "this review body has enough characters")
	if err != nil || !created || f.created != 1 {
		t.Fatalf("create %v %v %d", created, err, f.created)
	}
	f.facts.CompletedLessons, f.facts.TotalLessons = 4, 6
	_, created, err = s.PutReview(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse, 5, "another review body with enough characters")
	if err != nil || created || f.updated != 1 || f.lastProgress != (domain.ReviewProgress{Completed: 4, Total: 6}) {
		t.Fatalf("update created=%v err=%v progress=%+v", created, err, f.lastProgress)
	}
}
func TestReviewDeleteReportRemoveAndPublic(t *testing.T) {
	f := newReviewFake()
	s := reviewService(f)
	stored := f.byID[reviewID]
	f.own = &stored
	f.facts.Facts = domain.ReviewFacts{Status: domain.Archived}
	if err := s.DeleteReview(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse); err != nil || !f.deleted {
		t.Fatalf("delete=%v", err)
	}
	now := time.Now()
	r := f.byID[reviewID]
	r.RemovedAt = &now
	f.own = &r
	if err := s.DeleteReview(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse); !errors.Is(err, domain.ErrReviewRemoved) {
		t.Fatal(err)
	}
	f.reportErr = domain.ErrReportOwnReview
	if err := s.ReportReview(context.Background(), Actor{UserID: reviewAuthor}, reviewID); !errors.Is(err, domain.ErrReportOwnReview) {
		t.Fatal(err)
	}
	if err := s.RemoveReview(context.Background(), Actor{UserID: reviewOwner}, reviewID, "reason"); !errors.Is(err, domain.ErrReviewModerationForbidden) {
		t.Fatal(err)
	}
	if err := s.RemoveReview(context.Background(), Actor{UserID: "admin", Administrator: true}, reviewID, "reason"); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveReview(context.Background(), Actor{UserID: "admin", Administrator: true}, reviewID, "reason"); err != nil {
		t.Fatal(err)
	}
	f.publicID = reviewCourse
	f.summaryCount, f.summarySum = 2, 9
	page, err := s.PublicReviews(context.Background(), "slug", "")
	if err != nil || page.Rating != nil {
		t.Fatalf("public=%+v err=%v", page, err)
	}
	f.publicID = ""
	if _, err := s.PublicReviews(context.Background(), "archived", ""); !errors.Is(err, domain.ErrNotFound) {
		t.Fatal(err)
	}
}
