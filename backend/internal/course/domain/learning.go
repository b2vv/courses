package domain

import (
	"math"

	"courses/internal/contracts/access"
)

const MaximumLearningClientSequence int64 = 1<<31 - 1

type LearningVideoState string

const (
	VideoAbsent     LearningVideoState = ""
	VideoProcessing LearningVideoState = "processing"
	VideoReady      LearningVideoState = "ready"
	VideoFailed     LearningVideoState = "failed"
)

type LearningLesson struct {
	ID             string
	Title          string
	VideoState     LearningVideoState
	Position       float64
	Completed      bool
	ClientSequence int64
	// IsPreview is the lesson's own preview fact (ADR 014: preview is lesson-level).
	IsPreview bool
	// Locked is computed for the viewer by GetCourse: the lesson needs a purchase.
	Locked bool
}

// NewLearningLesson applies the playback position bound at the read boundary. ApplyProgress
// deliberately repeats the same bound at the write boundary; both use only a ready asset.
func NewLearningLesson(input LearningLesson, durationSeconds *float64) LearningLesson {
	if input.VideoState == VideoReady && durationSeconds != nil && input.Position > *durationSeconds {
		input.Position = *durationSeconds
	}
	return input
}

type LearningModule struct {
	ID      string
	Title   string
	Lessons []LearningLesson
}

type LearningCourse struct {
	Course
	InstructorID     string
	Modules          []LearningModule
	ContinueLessonID string
	Paid             bool
	Entitled         bool
	Membership       access.Membership
}

func NewLearningCourse(input LearningCourse) (LearningCourse, error) {
	course, err := NewCourse(input.Course)
	if err != nil || !validID(input.InstructorID) || len(input.Modules) > MaximumModules {
		return LearningCourse{}, ErrInvalidCourse
	}

	modules := make([]LearningModule, 0, len(input.Modules))
	seen := make(map[string]struct{}, len(input.Modules))
	lessonCount := 0
	for _, module := range input.Modules {
		if !validLearningNode(module.ID, module.Title, seen) {
			return LearningCourse{}, ErrInvalidCourse
		}
		lessons := make([]LearningLesson, 0, len(module.Lessons))
		for _, lesson := range module.Lessons {
			lessonCount++
			if lessonCount > MaximumLessons || !validLearningNode(lesson.ID, lesson.Title, seen) ||
				!lesson.VideoState.Valid() || math.IsNaN(lesson.Position) || math.IsInf(lesson.Position, 0) || lesson.Position < 0 ||
				lesson.ClientSequence < 0 || lesson.ClientSequence > MaximumLearningClientSequence {
				return LearningCourse{}, ErrInvalidCourse
			}
			lessons = append(lessons, lesson)
		}
		modules = append(modules, LearningModule{ID: module.ID, Title: module.Title, Lessons: lessons})
	}
	course.LessonCount = lessonCount

	result := LearningCourse{
		Course: course, InstructorID: input.InstructorID, Modules: modules,
		Paid: input.Paid, Entitled: input.Entitled, Membership: input.Membership,
	}
	result.ContinueLessonID = continueLessonID(modules)
	return result, nil
}

// CanBeViewedBy decides the outline. A published course's structure is visible like a
// preview (owner decision 2026-10-01, docs/payment/assumptions.md): titles are public on
// the storefront already, and content stays gated per lesson. Hence preview is true here.
func (course LearningCourse) CanBeViewedBy(actor Actor) bool {
	return access.CanView(access.Subject{UserID: actor.UserID, Administrator: actor.Role == ActorAdmin}, access.Course{OwnerID: course.InstructorID, Published: course.Status == Published, Paid: course.Paid, Entitled: course.Entitled, Membership: course.Membership, Preview: true})
}

// PurchaseRequiredBy reports whether lesson needs a purchase for actor, with that
// lesson's own preview fact.
func (course LearningCourse) PurchaseRequiredBy(actor Actor, lesson LearningLesson) bool {
	return access.PurchaseRequired(access.Subject{UserID: actor.UserID, Administrator: actor.Role == ActorAdmin}, access.Course{OwnerID: course.InstructorID, Published: course.Status == Published, Paid: course.Paid, Entitled: course.Entitled, Membership: course.Membership, Preview: lesson.IsPreview})
}

// CanPurchaseBy controls whether the learning outline presents a purchase entry. An administrator
// is never a learner (G8): they view the course as exempt but cannot buy it, so they are excluded
// here just as the owner is. StartPurchase refuses an administrator at the service too.
func (course LearningCourse) CanPurchaseBy(actor Actor) bool {
	return course.Status == Published && course.Paid && !course.Entitled && actor.UserID != course.InstructorID && actor.Role != ActorAdmin
}

func (state LearningVideoState) Valid() bool {
	return state == VideoAbsent || state == VideoProcessing || state == VideoReady || state == VideoFailed
}

func validLearningNode(id, title string, seen map[string]struct{}) bool {
	if !validID(id) || !validText(title, 200) {
		return false
	}
	if _, exists := seen[id]; exists {
		return false
	}
	seen[id] = struct{}{}
	return true
}

func continueLessonID(modules []LearningModule) string {
	last := ""
	for _, module := range modules {
		for _, lesson := range module.Lessons {
			last = lesson.ID
			if !lesson.Completed {
				return lesson.ID
			}
		}
	}
	return last
}
