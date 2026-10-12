package application

import (
	"context"
	"errors"
	"testing"

	"courses/internal/contracts/access"
	"courses/internal/image/domain"
)

func TestAuthorize(t *testing.T) {
	verified := domain.Actor{UserID: "viewer", EmailVerified: true}
	active := domain.ImageAccess{Membership: access.MembershipActive}
	readerErr := errors.New("reader failed")
	for _, tt := range []struct {
		name    string
		actor   domain.Actor
		course  domain.ImageAccess
		readErr error
		wantErr error
	}{
		{name: "active member", actor: verified, course: active},
		{name: "unverified member", actor: domain.Actor{UserID: "viewer"}, course: active, wantErr: domain.ErrEmailNotVerified},
		{name: "email checked before consumption", actor: domain.Actor{UserID: "viewer"}, wantErr: domain.ErrEmailNotVerified},
		{name: "not consumable", actor: verified, wantErr: domain.ErrForbidden},
		{name: "missing image", actor: verified, readErr: domain.ErrImageNotFound, wantErr: domain.ErrImageNotFound},
		{name: "missing image before email check", readErr: domain.ErrImageNotFound, wantErr: domain.ErrImageNotFound},
		{name: "reader failure", actor: verified, readErr: readerErr, wantErr: readerErr},
		{name: "owner", actor: verified, course: domain.ImageAccess{OwnerID: "viewer"}},
		{name: "admin", actor: domain.Actor{UserID: "admin", Administrator: true, EmailVerified: true}},
		{name: "unverified owner", actor: domain.Actor{UserID: "viewer"}, course: domain.ImageAccess{OwnerID: "viewer"}, wantErr: domain.ErrEmailNotVerified},
		{name: "unverified admin", actor: domain.Actor{UserID: "admin", Administrator: true}, wantErr: domain.ErrEmailNotVerified},
		{name: "blocked member can consume", actor: verified, course: domain.ImageAccess{Membership: access.MembershipBlocked}},
		{name: "revoked member", actor: verified, course: domain.ImageAccess{Membership: access.MembershipRevoked}, wantErr: domain.ErrForbidden},
		{name: "published preview", actor: verified, course: domain.ImageAccess{Published: true, Paid: true, Preview: true}},
		{name: "unpublished preview", actor: verified, course: domain.ImageAccess{Preview: true}, wantErr: domain.ErrForbidden},
		{name: "entitlement alone is insufficient", actor: verified, course: domain.ImageAccess{Published: true, Paid: true, Entitled: true}, wantErr: domain.ErrForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reader := &fakeAccessReader{course: tt.course, err: tt.readErr}
			err := NewAuthorizer(reader).Authorize(context.Background(), tt.actor, "image")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Authorize error = %v, want %v", err, tt.wantErr)
			}
			if reader.calls != 1 || reader.userID != tt.actor.UserID || reader.imageID != "image" {
				t.Fatalf("access lookup = %+v", reader)
			}
		})
	}
}

type fakeAccessReader struct {
	course          domain.ImageAccess
	err             error
	userID, imageID string
	calls           int
}

func (f *fakeAccessReader) LoadImageAccess(_ context.Context, userID, imageID string) (domain.ImageAccess, error) {
	f.calls++
	f.userID, f.imageID = userID, imageID
	return f.course, f.err
}
