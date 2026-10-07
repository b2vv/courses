package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/discussion/domain"
	"courses/internal/discussion/ports"
)

const reviewCourse = "10000000-0000-4000-8000-000000000001"
const reviewAuthor = "40000000-0000-4000-8000-000000000001"
const reviewOwner = "40000000-0000-4000-8000-000000000002"
const reviewID = "70000000-0000-4000-8000-000000000001"

type reviewFake struct {
	facts                    ports.ReviewFactsRecord
	factsErr                 error
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
	lesson                   ports.LessonFacts
	comment                  ports.CommentRecord
	inserted                 int
}

func (f *reviewFake) PublishedReviewCourse(context.Context, string) (string, error) {
	if f.publicID == "" {
		return "", domain.ErrNotFound
	}
	return f.publicID, nil
}
func (f *reviewFake) ReviewFacts(context.Context, string, string) (ports.ReviewFactsRecord, error) {
	return f.facts, f.factsErr
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
	return f.lesson, nil
}
func (f *reviewFake) List(context.Context, string, string, int) (ports.CommentPage, error) {
	return ports.CommentPage{}, nil
}
func (f *reviewFake) Get(context.Context, string, string) (ports.CommentRecord, error) {
	if f.comment.Comment.ID != "" {
		return f.comment, nil
	}
	return ports.CommentRecord{}, domain.ErrNotFound
}
func (f *reviewFake) InsertRoot(context.Context, domain.Comment) (ports.CommentRecord, error) {
	f.inserted++
	return ports.CommentRecord{}, nil
}
func (f *reviewFake) InsertReply(context.Context, domain.Comment) (ports.CommentRecord, error) {
	return ports.CommentRecord{}, nil
}
func (f *reviewFake) UpdateBody(context.Context, string, string, string, time.Time) (ports.CommentRecord, error) {
	return ports.CommentRecord{}, nil
}
func (f *reviewFake) Remove(context.Context, string, string, time.Time) error { return nil }
func (f *reviewFake) ReportSubject(context.Context, string, string) (string, access.Membership, bool, error) {
	return "", access.MembershipNone, false, nil
}
func (f *reviewFake) InsertParticipantReport(context.Context, domain.ParticipantReport) (string, error) {
	return "70000000-0000-4000-8000-0000000000dd", nil
}
func (f *reviewFake) ParticipantReportQueue(context.Context, string, int) (ports.ParticipantReportQueuePage, error) {
	return ports.ParticipantReportQueuePage{}, nil
}
func (f *reviewFake) ResolveParticipantReport(context.Context, string) error { return nil }

func newReviewFake() *reviewFake {
	r := ports.ReviewRecord{ID: reviewID, CourseID: reviewCourse, AuthorID: reviewAuthor, Stars: 4, Body: "this review body has enough characters", CreatedAt: time.Now()}
	return &reviewFake{facts: ports.ReviewFactsRecord{Facts: domain.ReviewFacts{Status: domain.Published, Membership: access.MembershipActive, HasProgress: true}, CompletedLessons: 1, TotalLessons: 4}, byID: map[string]ports.ReviewRecord{reviewID: r}, course: domain.Course{OwnerID: reviewOwner, Status: domain.Published}, lesson: ports.LessonFacts{Course: domain.Course{OwnerID: reviewOwner, Status: domain.Published, Membership: access.MembershipActive}}}
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
	}{{"anonymous", domain.ReviewFacts{Status: domain.Published, HasProgress: true}, Actor{}, ErrEmailNotVerified}, {"owner", domain.ReviewFacts{OwnerID: reviewOwner, Status: domain.Published, Membership: access.MembershipActive, HasProgress: true}, Actor{UserID: reviewOwner, EmailVerified: true}, domain.ErrReviewOwnerForbidden}, {"unavailable", domain.ReviewFacts{Status: domain.Archived, Membership: access.MembershipActive, HasProgress: true}, Actor{UserID: reviewAuthor, EmailVerified: true}, domain.ErrReviewNotAvailable}, {"purchase", domain.ReviewFacts{Status: domain.Published, Paid: true, Membership: access.MembershipActive, HasProgress: true}, Actor{UserID: reviewAuthor, EmailVerified: true}, domain.ErrReviewPurchaseRequired}, {"not started", domain.ReviewFacts{Status: domain.Published, Membership: access.MembershipActive}, Actor{UserID: reviewAuthor, EmailVerified: true}, domain.ErrReviewNotStarted}}
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

func TestCommunicationGatePrecedence(t *testing.T) {
	t.Run("publish keeps email first then distinguishes membership", func(t *testing.T) {
		cases := []struct {
			name       string
			verified   bool
			membership access.Membership
			want       error
		}{
			{"unverified blocked", false, access.MembershipBlocked, ErrEmailNotVerified},
			{"verified blocked", true, access.MembershipBlocked, domain.ErrEnrollmentBlocked},
			{"verified none", true, access.MembershipNone, domain.ErrEnrollmentRequired},
			{"verified revoked", true, access.MembershipRevoked, domain.ErrEnrollmentRequired},
			{"verified active", true, access.MembershipActive, nil},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				f := newReviewFake()
				f.lesson.Course.Membership = tc.membership
				_, err := reviewService(f).Publish(context.Background(), Actor{UserID: reviewAuthor, EmailVerified: tc.verified}, PublishRequest{LessonID: "lesson", Body: "comment"})
				if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) {
					t.Fatalf("err=%v want=%v", err, tc.want)
				}
				if tc.want == nil && f.inserted != 1 {
					t.Fatalf("inserted=%d", f.inserted)
				}
			})
		}
	})

	t.Run("list consumes active and blocked membership", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			state access.Membership
			want  error
		}{{"none", access.MembershipNone, domain.ErrForbidden}, {"blocked", access.MembershipBlocked, nil}, {"active", access.MembershipActive, nil}} {
			t.Run(tc.name, func(t *testing.T) {
				f := newReviewFake()
				f.lesson.Course.Membership = tc.state
				_, err := reviewService(f).List(context.Background(), Actor{UserID: reviewAuthor}, "lesson", "")
				if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) {
					t.Fatalf("err=%v want=%v", err, tc.want)
				}
			})
		}
	})

	t.Run("report comment uses the same membership codes", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			state access.Membership
			want  error
		}{{"blocked", access.MembershipBlocked, domain.ErrEnrollmentBlocked}, {"revoked", access.MembershipRevoked, domain.ErrEnrollmentRequired}, {"active", access.MembershipActive, nil}} {
			t.Run(tc.name, func(t *testing.T) {
				f := newReviewFake()
				f.comment = ports.CommentRecord{Comment: domain.Comment{ID: "comment", AuthorID: "other"}, Course: domain.Course{OwnerID: reviewOwner, Status: domain.Published, Membership: tc.state}}
				err := reviewService(f).ReportComment(context.Background(), Actor{UserID: reviewAuthor, EmailVerified: true}, "comment")
				if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) {
					t.Fatalf("err=%v want=%v", err, tc.want)
				}
			})
		}
	})

	t.Run("review communication precedes review eligibility", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			actor    Actor
			state    access.Membership
			progress bool
			want     error
		}{
			{"blocked", Actor{UserID: reviewAuthor, EmailVerified: true}, access.MembershipBlocked, true, domain.ErrEnrollmentBlocked},
			{"none", Actor{UserID: reviewAuthor, EmailVerified: true}, access.MembershipNone, true, domain.ErrEnrollmentRequired},
			{"active progress", Actor{UserID: reviewAuthor, EmailVerified: true}, access.MembershipActive, true, nil},
			{"owner after active", Actor{UserID: reviewOwner, EmailVerified: true}, access.MembershipActive, true, domain.ErrReviewOwnerForbidden},
		} {
			t.Run(tc.name, func(t *testing.T) {
				f := newReviewFake()
				f.facts.Facts.OwnerID = reviewOwner
				f.facts.Facts.Membership = tc.state
				f.facts.Facts.HasProgress = tc.progress
				_, _, err := reviewService(f).PutReview(context.Background(), tc.actor, reviewCourse, 4, "this review body has enough characters")
				if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) {
					t.Fatalf("err=%v want=%v", err, tc.want)
				}
			})
		}
	})

	t.Run("report review loads the review course before applying the gate", func(t *testing.T) {
		for _, tc := range []struct {
			name  string
			state access.Membership
			want  error
		}{{"blocked", access.MembershipBlocked, domain.ErrEnrollmentBlocked}, {"none", access.MembershipNone, domain.ErrEnrollmentRequired}, {"active", access.MembershipActive, nil}} {
			t.Run(tc.name, func(t *testing.T) {
				f := newReviewFake()
				f.facts.Facts.Membership = tc.state
				err := reviewService(f).ReportReview(context.Background(), Actor{UserID: "reporter", EmailVerified: true}, reviewID)
				if !errors.Is(err, tc.want) || (tc.want == nil && err != nil) {
					t.Fatalf("err=%v want=%v", err, tc.want)
				}
			})
		}
	})
}
func TestPutReviewCreatesThenUpdatesRecapturedProgress(t *testing.T) {
	f := newReviewFake()
	s := reviewService(f)
	_, created, err := s.PutReview(context.Background(), Actor{UserID: reviewAuthor, EmailVerified: true}, reviewCourse, 4, "this review body has enough characters")
	if err != nil || !created || f.created != 1 {
		t.Fatalf("create %v %v %d", created, err, f.created)
	}
	f.facts.CompletedLessons, f.facts.TotalLessons = 4, 6
	_, created, err = s.PutReview(context.Background(), Actor{UserID: reviewAuthor, EmailVerified: true}, reviewCourse, 5, "another review body with enough characters")
	if err != nil || created || f.updated != 1 || f.lastProgress != (domain.ReviewProgress{Completed: 4, Total: 6}) {
		t.Fatalf("update created=%v err=%v progress=%+v", created, err, f.lastProgress)
	}
}

func TestPutReviewRefusesEditAfterAccessIsLost(t *testing.T) {
	f := newReviewFake()
	stored := f.byID[reviewID]
	f.own = &stored
	f.facts.Exists = true
	f.facts.Facts = domain.ReviewFacts{Status: domain.Published, Paid: true, Membership: access.MembershipActive, HasProgress: true}

	_, _, err := reviewService(f).PutReview(context.Background(), Actor{UserID: reviewAuthor, EmailVerified: true}, reviewCourse, 5, "another review body with enough characters")
	if !errors.Is(err, domain.ErrReviewPurchaseRequired) {
		t.Fatalf("err=%v", err)
	}
	if f.updated != 0 || f.own.Stars != stored.Stars || f.own.Body != stored.Body {
		t.Fatalf("stored review changed: %+v", f.own)
	}
}

func TestLearningReviews(t *testing.T) {
	t.Run("grant holder can list archived course reviews", func(t *testing.T) {
		f := newReviewFake()
		f.facts.Facts = domain.ReviewFacts{Status: domain.Archived, Paid: true, Entitled: true, Membership: access.MembershipActive}
		f.summaryCount, f.summarySum = 3, 14
		f.page = ports.ReviewPage{Items: []ports.ReviewRecord{f.byID[reviewID]}, NextCursor: "next-page"}

		page, err := reviewService(f).LearningReviews(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse, "")
		if err != nil || page.Rating == nil || page.Rating.Average != 4.7 || page.Rating.Count != 3 || len(page.Items) != 1 || page.NextCursor != "next-page" {
			t.Fatalf("page=%+v err=%v", page, err)
		}
	})
	t.Run("without a viewing right is forbidden", func(t *testing.T) {
		f := newReviewFake()
		f.facts.Facts = domain.ReviewFacts{Status: domain.Archived, Paid: true}

		_, err := reviewService(f).LearningReviews(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse, "")
		if !errors.Is(err, domain.ErrForbidden) {
			t.Fatalf("err=%v", err)
		}
	})
	t.Run("unknown course is not found", func(t *testing.T) {
		f := newReviewFake()
		f.factsErr = domain.ErrNotFound

		_, err := reviewService(f).LearningReviews(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse, "")
		if !errors.Is(err, domain.ErrNotFound) {
			t.Fatalf("err=%v", err)
		}
	})
}
func TestReviewDeleteReportRemoveAndPublic(t *testing.T) {
	f := newReviewFake()
	s := reviewService(f)
	stored := f.byID[reviewID]
	f.own = &stored
	f.facts.Facts = domain.ReviewFacts{Status: domain.Archived, Membership: access.MembershipActive}
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
	if err := s.ReportReview(context.Background(), Actor{UserID: reviewAuthor, EmailVerified: true}, reviewID); !errors.Is(err, domain.ErrReportOwnReview) {
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

func TestRemoveReviewIsNoopWhenAlreadyRemoved(t *testing.T) {
	f := newReviewFake()
	s := reviewService(f)
	if err := s.RemoveReview(context.Background(), Actor{UserID: reviewOwner}, reviewID, "reason"); !errors.Is(err, domain.ErrReviewModerationForbidden) {
		t.Fatal(err)
	}
	admin := Actor{UserID: "admin", Administrator: true}
	if err := s.RemoveReview(context.Background(), admin, reviewID, "reason"); err != nil {
		t.Fatal(err)
	}
	if err := s.RemoveReview(context.Background(), admin, reviewID, "reason"); err != nil {
		t.Fatal(err)
	}
	if f.removed != 1 {
		t.Fatalf("repository remove calls=%d want 1", f.removed)
	}
}

func TestOwnReviewRemovedByModerationIsNotShownAsLive(t *testing.T) {
	f := newReviewFake()
	removedAt := time.Now()
	removed := ports.ReviewRecord{ID: reviewID, CourseID: reviewCourse, AuthorID: reviewAuthor, Stars: 2, Body: "this review body has enough characters", CreatedAt: removedAt, RemovedAt: &removedAt}
	f.own = &removed
	got, err := reviewService(f).OwnReview(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse)
	if err != nil || got.Review != nil || got.Eligibility != domain.ErrReviewRemoved.Error() {
		t.Fatalf("own review after removal = %+v, %v; want no review and eligibility %q", got, err, domain.ErrReviewRemoved.Error())
	}
	live := removed
	live.RemovedAt = nil
	f.own = &live
	got, err = reviewService(f).OwnReview(context.Background(), Actor{UserID: reviewAuthor}, reviewCourse)
	if err != nil || got.Review == nil || got.Eligibility != "ok" {
		t.Fatalf("live own review = %+v, %v", got, err)
	}
}
