package application_test

import (
	"context"
	"errors"
	"testing"

	"courses/internal/contracts/access"
	"courses/internal/transcript/application"
	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
)

const readLessonID = "50000000-0000-4000-8000-000000000001"

type readRepository struct{ result ports.LessonTranscript }

func (repository readRepository) GetLessonTranscript(context.Context, string, string, domain.Language) (ports.LessonTranscript, error) {
	return repository.result, nil
}
func (readRepository) LessonCourse(context.Context, string) (domain.Course, error) {
	return domain.Course{}, errors.New("unexpected")
}
func (readRepository) CourseDocuments(context.Context, string) ([]domain.DocumentLink, error) {
	return nil, errors.New("unexpected")
}
func (readRepository) AddCourseDocument(context.Context, string, domain.DocumentLink) error {
	return errors.New("unexpected")
}
func (readRepository) RemoveCourseDocument(context.Context, string, string) error {
	return errors.New("unexpected")
}
func (readRepository) CourseOfID(context.Context, string) (domain.Course, error) {
	return domain.Course{}, errors.New("unexpected")
}
func (readRepository) SaveLessonTranscript(context.Context, ports.SaveTranscript) (string, error) {
	return "", errors.New("unexpected")
}

func TestReadingConsumesOnlyMembershipOrPreviewContent(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		actor  domain.Actor
		course domain.Course
		want   error
	}{
		{name: "active member consumes", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{Status: domain.Draft, Membership: access.MembershipActive}},
		{name: "blocked member consumes", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{Status: domain.Draft, Membership: access.MembershipBlocked}},
		{name: "revoked member denied from free content", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{Status: domain.Published, Membership: access.MembershipRevoked}, want: domain.ErrForbidden},
		{name: "free non member denied", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{Status: domain.Published}, want: domain.ErrForbidden},
		{name: "paid non member must purchase", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{Status: domain.Published, Paid: true}, want: domain.ErrPurchaseRequired},
		{name: "preview non member consumes", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{Status: domain.Published, Preview: true}},
		{name: "preview unverified is email denied", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}, course: domain.Course{Status: domain.Published, Preview: true}, want: domain.ErrEmailNotVerified},
		{name: "member unverified is email denied", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001"}, course: domain.Course{Status: domain.Draft, Membership: access.MembershipActive}, want: domain.ErrEmailNotVerified},
		{name: "owner consumes without membership", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", EmailVerified: true}, course: domain.Course{OwnerID: "40000000-0000-4000-8000-000000000001", Status: domain.Draft}},
		{name: "administrator consumes without membership", actor: domain.Actor{UserID: "40000000-0000-4000-8000-000000000001", Administrator: true, EmailVerified: true}, course: domain.Course{Status: domain.Draft}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := application.NewReading(readRepository{result: ports.LessonTranscript{Course: test.course}}).Get(context.Background(), test.actor, readLessonID, "")
			if !errors.Is(err, test.want) {
				t.Fatalf("Get() error = %v, want %v", err, test.want)
			}
		})
	}
}
