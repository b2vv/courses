package domain_test

import (
	"courses/internal/contracts/access/accesstest"
	"testing"
	"time"

	coursedomain "courses/internal/course/domain"
	"courses/internal/testsupport/coursevisibility"
)

const (
	learningOwnerID  = "40000000-0000-4000-8000-000000000002"
	learningViewerID = "40000000-0000-4000-8000-000000000001"
	firstLessonID    = "70000000-0000-4000-8000-000000000001"
	secondLessonID   = "70000000-0000-4000-8000-000000000002"
	thirdLessonID    = "70000000-0000-4000-8000-000000000003"
)

func TestLearningCourseContinueLessonRule(t *testing.T) {
	for _, test := range []struct {
		name    string
		lessons []coursedomain.LearningLesson
		want    string
	}{
		{name: "B01 no lessons", lessons: nil, want: ""},
		{name: "no progress chooses first", lessons: []coursedomain.LearningLesson{
			learningLesson(firstLessonID, false, coursedomain.VideoReady),
			learningLesson(secondLessonID, false, coursedomain.VideoReady),
		}, want: firstLessonID},
		{name: "first incomplete wins despite no video and later completion", lessons: []coursedomain.LearningLesson{
			learningLesson(firstLessonID, true, coursedomain.VideoReady),
			learningLesson(secondLessonID, false, coursedomain.VideoAbsent),
			learningLesson(thirdLessonID, true, coursedomain.VideoReady),
		}, want: secondLessonID},
		{name: "all completed chooses last", lessons: []coursedomain.LearningLesson{
			learningLesson(firstLessonID, true, coursedomain.VideoReady),
			learningLesson(secondLessonID, true, coursedomain.VideoFailed),
		}, want: secondLessonID},
	} {
		t.Run(test.name, func(t *testing.T) {
			course, err := coursedomain.NewLearningCourse(learningCourse(test.lessons))
			if err != nil {
				t.Fatal(err)
			}
			if course.ContinueLessonID != test.want {
				t.Fatalf("continueLessonId=%q want %q", course.ContinueLessonID, test.want)
			}
			if len(test.lessons) == 0 && (course.Modules == nil || course.Modules[0].Lessons == nil) {
				t.Fatal("empty program must use non-nil arrays")
			}
		})
	}
}

func TestB27MissingProgressUsesConcreteZeroValues(t *testing.T) {
	lesson := learningLesson(firstLessonID, false, coursedomain.VideoProcessing)
	course, err := coursedomain.NewLearningCourse(learningCourse([]coursedomain.LearningLesson{lesson}))
	if err != nil {
		t.Fatal(err)
	}
	got := course.Modules[0].Lessons[0]
	if got.Position != 0 || got.Completed || got.ClientSequence != 0 {
		t.Fatalf("missing progress defaults=%+v", got)
	}
}

func TestLearningLessonPositionUsesReadyAssetDuration(t *testing.T) {
	durationSeconds := 480.0
	lesson := learningLesson(firstLessonID, false, coursedomain.VideoReady)
	lesson.Position = 2050

	got := coursedomain.NewLearningLesson(lesson, &durationSeconds)

	if got.Position != 480 {
		t.Fatalf("position=%v want 480", got.Position)
	}
}

func TestLearningLessonPositionIsUnchangedWithoutReadyDuration(t *testing.T) {
	durationSeconds := 480.0
	for _, test := range []struct {
		name            string
		state           coursedomain.LearningVideoState
		durationSeconds *float64
	}{
		{name: "no asset", state: coursedomain.VideoAbsent, durationSeconds: nil},
		{name: "processing asset with duration", state: coursedomain.VideoProcessing, durationSeconds: &durationSeconds},
		{name: "ready asset with null duration", state: coursedomain.VideoReady, durationSeconds: nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			lesson := learningLesson(firstLessonID, false, test.state)
			lesson.Position = 2050

			got := coursedomain.NewLearningLesson(lesson, test.durationSeconds)

			if got.Position != 2050 {
				t.Fatalf("position=%v want 2050", got.Position)
			}
		})
	}
}

func TestO4CourseVisibilityUsesSharedCases(t *testing.T) {
	for _, test := range coursevisibility.Cases() {
		t.Run(test.Name, func(t *testing.T) {
			course, err := coursedomain.NewLearningCourse(learningCourse(nil))
			if err != nil {
				t.Fatal(err)
			}
			course.Status = coursedomain.Status(test.Status)
			actor := coursedomain.Actor{UserID: learningViewerID, Role: coursedomain.ActorRole(test.Actor)}
			if test.Actor == coursevisibility.Owner {
				actor.UserID = learningOwnerID
				actor.Role = coursedomain.ActorInstructor
			}
			if got := course.CanBeViewedBy(actor); got != test.Want {
				t.Fatalf("course=%t want=%t", got, test.Want)
			}
		})
	}
}

// The outline of a published course is visible like a preview (owner decision
// 2026-10-01): the rule moves to the lesson, where locked equals the shared table's
// WantPurchase with that lesson's preview fact. Outline visibility is therefore
// Want || WantPurchase — visible content, or a purchase that would make it visible.
func TestLearningOutlineAndLessonLockFollowTheSharedTruthTable(t *testing.T) {
	for _, test := range accesstest.TruthTable {
		t.Run(test.Name, func(t *testing.T) {
			course, err := coursedomain.NewLearningCourse(learningCourse([]coursedomain.LearningLesson{
				{ID: "70000000-0000-4000-8000-000000000001", Title: "Lesson", IsPreview: test.Preview},
			}))
			if err != nil {
				t.Fatal(err)
			}
			course.Status = coursedomain.Status(test.Status)
			course.Paid = test.Paid
			course.Entitled = test.Granted
			actor := coursedomain.Actor{UserID: learningViewerID, Role: coursedomain.ActorStudent}
			switch test.Role {
			case accesstest.Administrator:
				actor.Role = coursedomain.ActorAdmin
			case accesstest.Owner:
				actor = coursedomain.Actor{UserID: learningOwnerID, Role: coursedomain.ActorInstructor}
			}
			if got, want := course.CanBeViewedBy(actor), test.Want || test.WantPurchase; got != want {
				t.Fatalf("outline visible=%t want=%t", got, want)
			}
			if got := course.PurchaseRequiredBy(actor, course.Modules[0].Lessons[0]); got != test.WantPurchase {
				t.Fatalf("lesson locked=%t want=%t", got, test.WantPurchase)
			}
		})
	}
}

func learningCourse(lessons []coursedomain.LearningLesson) coursedomain.LearningCourse {
	now := time.Date(2026, 9, 8, 12, 0, 0, 0, time.UTC)
	return coursedomain.LearningCourse{
		Course: coursedomain.Course{
			ID: "50000000-0000-4000-8000-000000000001", Slug: "learning-course",
			Title: "Learning course", Summary: "Learning summary", Description: "Learning description",
			Status: coursedomain.Published, CreatedAt: now, UpdatedAt: now, PublishedAt: &now,
		},
		InstructorID: learningOwnerID,
		Modules: []coursedomain.LearningModule{{
			ID: "60000000-0000-4000-8000-000000000001", Title: "Module", Lessons: lessons,
		}},
	}
}

func learningLesson(id string, completed bool, state coursedomain.LearningVideoState) coursedomain.LearningLesson {
	return coursedomain.LearningLesson{
		ID: id, Title: "Lesson", VideoState: state,
		Position: 0, Completed: completed, ClientSequence: 0,
	}
}
