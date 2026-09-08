package application

import (
	"context"
	"time"

	"courses/internal/course/domain"
	"courses/internal/course/ports"
)

type Learning struct{ repository ports.LearningRepository }

func NewLearning(repository ports.LearningRepository) *Learning {
	return &Learning{repository: repository}
}

func (learning *Learning) GetCourse(ctx context.Context, actor domain.Actor, slug string) (domain.LearningCourse, error) {
	if err := domain.ValidateSlug(slug); err != nil {
		return domain.LearningCourse{}, err
	}
	if err := domain.ValidateID(actor.UserID); err != nil || !actor.Role.Valid() {
		return domain.LearningCourse{}, domain.ErrForbidden
	}
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	course, err := learning.repository.GetLearningCourse(ctx, actor.UserID, slug)
	if err != nil {
		return domain.LearningCourse{}, err
	}
	if !course.CanBeViewedBy(actor) {
		return domain.LearningCourse{}, domain.ErrCourseForbidden
	}
	return course, nil
}
