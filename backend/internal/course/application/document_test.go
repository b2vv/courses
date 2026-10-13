package application_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"courses/internal/contracts/access"
	"courses/internal/course/application"
	"courses/internal/course/domain"
)

const documentOwnerID = "11111111-1111-4111-8111-111111111111"
const documentStudentID = "22222222-2222-4222-8222-222222222222"
const documentLessonID = "33333333-3333-4333-8333-333333333333"

var documentOwner = domain.Actor{UserID: documentOwnerID, Role: domain.ActorInstructor, EmailVerified: true}
var documentStudent = domain.Actor{UserID: documentStudentID, Role: domain.ActorStudent, EmailVerified: true}

type documentAccessFake struct {
	facts   domain.LessonDocumentAccess
	byUser  map[string]domain.LessonDocumentAccess
	users   []string
	lessons []string
	err     error
}

func (f *documentAccessFake) LoadLessonAccess(_ context.Context, userID, lessonID string) (domain.LessonDocumentAccess, error) {
	f.users = append(f.users, userID)
	f.lessons = append(f.lessons, lessonID)
	if f.byUser != nil {
		return f.byUser[userID], f.err
	}
	return f.facts, f.err
}

type documentRepositoryFake struct {
	draft, published           domain.LessonDocument
	updatedAt                  time.Time
	draftFound, publishedFound bool
	calls                      []string
	err                        error
}

func (f *documentRepositoryFake) record(method, lessonID, locale string) {
	f.calls = append(f.calls, method+":"+lessonID+":"+locale)
}
func (f *documentRepositoryFake) SaveDraft(_ context.Context, lessonID, locale string, body domain.LessonDocument, expected time.Time) (time.Time, error) {
	f.record("save", lessonID, locale)
	if f.err != nil {
		return time.Time{}, f.err
	}
	if !expected.Equal(f.updatedAt) {
		return time.Time{}, fmt.Errorf("save: %w", domain.ErrDocumentConflict)
	}
	f.draft, f.draftFound = body, true
	f.updatedAt = f.updatedAt.Add(time.Second)
	return f.updatedAt, nil
}
func (f *documentRepositoryFake) PublishDocument(_ context.Context, lessonID, locale string, expected time.Time) error {
	f.record("publish", lessonID, locale)
	if f.err != nil {
		return f.err
	}
	if !expected.Equal(f.updatedAt) {
		return fmt.Errorf("publish: %w", domain.ErrDocumentConflict)
	}
	f.published, f.publishedFound = f.draft, f.draftFound
	return nil
}
func (f *documentRepositoryFake) LoadDraft(_ context.Context, lessonID, locale string) (domain.LessonDocument, time.Time, bool, error) {
	f.record("draft", lessonID, locale)
	return f.draft, f.updatedAt, f.draftFound, f.err
}
func (f *documentRepositoryFake) LoadPublished(_ context.Context, lessonID, locale string) (domain.LessonDocument, bool, error) {
	f.record("published", lessonID, locale)
	return f.published, f.publishedFound, f.err
}
func documentBody(locale, id string) domain.LessonDocument {
	return domain.LessonDocument{SchemaVersion: domain.CurrentLessonDocumentSchemaVersion, Locale: locale, Sections: []domain.Section{{ID: id, Type: domain.SectionVideo}}}
}
func documentFacts() domain.LessonDocumentAccess {
	return domain.LessonDocumentAccess{OwnerID: documentOwnerID, Status: domain.Published, Membership: access.MembershipActive}
}

func TestDocumentsPublishSeparatesDraft(t *testing.T) {
	for _, existing := range []bool{false, true} {
		t.Run(fmt.Sprintf("previous published=%t", existing), func(t *testing.T) {
			ctx := context.Background()
			old, draft := documentBody("uk", "old"), documentBody("uk", "new")
			repo := &documentRepositoryFake{published: old, publishedFound: existing}
			reader := &documentAccessFake{facts: documentFacts()}
			service := application.NewDocuments(repo, reader)
			token, err := service.SaveDraft(ctx, documentOwner, documentLessonID, "uk", draft, time.Time{})
			if err != nil || token.IsZero() {
				t.Fatalf("save = %v, %v", token, err)
			}
			gotDraft, gotToken, err := service.GetDraft(ctx, documentOwner, documentLessonID, "uk")
			if err != nil || !reflect.DeepEqual(gotDraft, draft) || !gotToken.Equal(token) {
				t.Fatalf("draft = %+v, %v, %v", gotDraft, gotToken, err)
			}
			before, err := service.GetPublishedDocument(ctx, documentStudent, documentLessonID, "uk")
			if err != nil {
				t.Fatal(err)
			}
			if existing {
				if !reflect.DeepEqual(before, old) {
					t.Fatalf("old publication changed: %+v", before)
				}
			} else {
				assertEmptyDocument(t, before, "uk")
			}
			if err := service.PublishDocument(ctx, documentOwner, documentLessonID, "uk", token); err != nil {
				t.Fatal(err)
			}
			after, err := service.GetPublishedDocument(ctx, documentStudent, documentLessonID, "uk")
			if err != nil || !reflect.DeepEqual(after, draft) {
				t.Fatalf("published = %+v, %v", after, err)
			}
			if !reflect.DeepEqual(reader.users, []string{documentOwnerID, documentOwnerID, documentStudentID, documentOwnerID, documentStudentID}) {
				t.Fatalf("viewer IDs = %v", reader.users)
			}
			for _, lesson := range reader.lessons {
				if lesson != documentLessonID {
					t.Fatalf("lesson = %s", lesson)
				}
			}
			wantCalls := []string{}
			for _, method := range []string{"save", "draft", "published", "publish", "published"} {
				wantCalls = append(wantCalls, method+":"+documentLessonID+":uk")
			}
			if !reflect.DeepEqual(repo.calls, wantCalls) {
				t.Fatalf("repository calls = %v", repo.calls)
			}
		})
	}
}

type documentOperation struct {
	name string
	call func(*application.Documents, domain.Actor) error
}

func documentOperations() []documentOperation {
	return []documentOperation{
		{"save", func(s *application.Documents, actor domain.Actor) error {
			_, err := s.SaveDraft(context.Background(), actor, documentLessonID, "uk", documentBody("uk", "new"), time.Time{})
			return err
		}},
		{"publish", func(s *application.Documents, actor domain.Actor) error {
			return s.PublishDocument(context.Background(), actor, documentLessonID, "uk", time.Time{})
		}},
		{"draft", func(s *application.Documents, actor domain.Actor) error {
			_, _, err := s.GetDraft(context.Background(), actor, documentLessonID, "uk")
			return err
		}},
		{"published", func(s *application.Documents, actor domain.Actor) error {
			_, err := s.GetPublishedDocument(context.Background(), actor, documentLessonID, "uk")
			return err
		}},
	}
}
func TestDocumentsEditorGate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		actor   domain.Actor
		allowed bool
	}{
		{"owner", documentOwner, true},
		{"admin", domain.Actor{UserID: documentStudentID, Role: domain.ActorAdmin}, true},
		{"other instructor", domain.Actor{UserID: documentStudentID, Role: domain.ActorInstructor}, false},
		{"student owner", domain.Actor{UserID: documentOwnerID, Role: domain.ActorStudent}, false},
		{"invalid actor", domain.Actor{Role: domain.ActorAdmin}, false},
	} {
		for _, op := range documentOperations()[:3] {
			t.Run(tc.name+"/"+op.name, func(t *testing.T) {
				repo := &documentRepositoryFake{}
				reader := &documentAccessFake{facts: documentFacts()}
				err := op.call(application.NewDocuments(repo, reader), tc.actor)
				if tc.allowed {
					if err != nil || len(repo.calls) != 1 {
						t.Fatalf("allowed: %v, %v", err, repo.calls)
					}
				} else if !errors.Is(err, domain.ErrForbidden) || len(repo.calls) != 0 {
					t.Fatalf("denied: %v, %v", err, repo.calls)
				}
				if !reflect.DeepEqual(reader.users, []string{tc.actor.UserID}) {
					t.Fatalf("viewer = %v", reader.users)
				}
			})
		}
	}
}

func TestDocumentsReadGate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		actor domain.Actor
		facts domain.LessonDocumentAccess
		want  error
	}{
		{"unverified", domain.Actor{UserID: documentStudentID, Role: domain.ActorStudent}, documentFacts(), domain.ErrEmailNotVerified},
		{"paid nonmember", documentStudent, domain.LessonDocumentAccess{Status: domain.Published, Paid: true}, domain.ErrPurchaseRequired},
		{"free nonmember", documentStudent, domain.LessonDocumentAccess{Status: domain.Published}, domain.ErrForbidden},
		{"paid preview", documentStudent, domain.LessonDocumentAccess{Status: domain.Published, Paid: true, Preview: true}, nil},
		{"draft preview", documentStudent, domain.LessonDocumentAccess{Status: domain.Draft, Paid: true, Preview: true}, domain.ErrForbidden},
		{"archived member", documentStudent, domain.LessonDocumentAccess{Status: domain.Archived, Membership: access.MembershipActive}, nil},
		{"blocked member", documentStudent, domain.LessonDocumentAccess{Paid: true, Membership: access.MembershipBlocked}, nil},
		{"revoked member", documentStudent, domain.LessonDocumentAccess{Status: domain.Published, Membership: access.MembershipRevoked}, domain.ErrForbidden},
		{"entitled without membership", documentStudent, domain.LessonDocumentAccess{Status: domain.Published, Paid: true, Entitled: true}, domain.ErrForbidden},
		{"owner", documentOwner, domain.LessonDocumentAccess{OwnerID: documentOwnerID, Status: domain.Draft}, nil},
		{"admin", domain.Actor{UserID: documentStudentID, Role: domain.ActorAdmin, EmailVerified: true}, domain.LessonDocumentAccess{Status: domain.Draft}, nil},
	} {
		for _, found := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/found=%t", tc.name, found), func(t *testing.T) {
				body := documentBody("uk", "published")
				repo := &documentRepositoryFake{published: body, publishedFound: found, draft: documentBody("uk", "secret"), draftFound: true}
				reader := &documentAccessFake{facts: tc.facts}
				got, err := application.NewDocuments(repo, reader).GetPublishedDocument(context.Background(), tc.actor, documentLessonID, "uk")
				if !errors.Is(err, tc.want) {
					t.Fatalf("error = %v, want %v", err, tc.want)
				}
				if tc.want != nil {
					if len(repo.calls) != 0 {
						t.Fatal("read content before gate")
					}
				} else {
					if found {
						if !reflect.DeepEqual(got, body) {
							t.Fatalf("body = %+v", got)
						}
					} else {
						assertEmptyDocument(t, got, "uk")
					}
				}
				expectedReads := 1
				if !tc.actor.EmailVerified {
					expectedReads = 0
				}
				if len(reader.users) != expectedReads {
					t.Fatalf("access reads = %v", reader.users)
				}
			})
		}
	}
}

func TestDocumentsAccessIsPerViewer(t *testing.T) {
	member, nonmember := documentStudent, documentOwner
	memberFacts := documentFacts()
	memberFacts.Paid = true
	nonmemberFacts := memberFacts
	nonmemberFacts.Membership = access.MembershipNone
	nonmemberFacts.OwnerID = "another-owner"
	reader := &documentAccessFake{byUser: map[string]domain.LessonDocumentAccess{member.UserID: memberFacts, nonmember.UserID: nonmemberFacts}}
	service := application.NewDocuments(&documentRepositoryFake{}, reader)
	for _, tc := range []struct {
		actor domain.Actor
		want  error
	}{{member, nil}, {nonmember, domain.ErrPurchaseRequired}} {
		_, err := service.GetPublishedDocument(context.Background(), tc.actor, documentLessonID, "uk")
		if !errors.Is(err, tc.want) {
			t.Fatalf("viewer %s: %v", tc.actor.UserID, err)
		}
	}
}

func TestDocumentsStaleWrites(t *testing.T) {
	for _, op := range documentOperations()[:2] {
		t.Run(op.name, func(t *testing.T) {
			original := documentBody("uk", "original")
			repo := &documentRepositoryFake{draft: original, published: original, draftFound: true, publishedFound: true, updatedAt: time.Now()}
			err := op.call(application.NewDocuments(repo, &documentAccessFake{facts: documentFacts()}), documentOwner)
			if err != domain.ErrDocumentConflict {
				t.Fatalf("error = %v", err)
			}
			if !reflect.DeepEqual(repo.draft, original) || !reflect.DeepEqual(repo.published, original) {
				t.Fatal("stale write changed body")
			}
		})
	}
}

func TestDocumentsInvalidDraft(t *testing.T) {
	repo := &documentRepositoryFake{}
	body := documentBody("uk", "invalid")
	body.Sections[0].Type = "unknown"
	_, err := application.NewDocuments(repo, &documentAccessFake{facts: documentFacts()}).SaveDraft(context.Background(), documentOwner, documentLessonID, "uk", body, time.Time{})
	if !errors.Is(err, domain.ErrInvalidLessonDocument) || len(repo.calls) != 0 {
		t.Fatalf("invalid save: %v, %v", err, repo.calls)
	}
}

func TestDocumentsMissing(t *testing.T) {
	for _, locale := range []string{"uk", "en"} {
		t.Run(locale, func(t *testing.T) {
			service := application.NewDocuments(&documentRepositoryFake{}, &documentAccessFake{facts: documentFacts()})
			draft, token, err := service.GetDraft(context.Background(), documentOwner, documentLessonID, locale)
			if err != nil || !token.IsZero() {
				t.Fatalf("missing draft: %v, %v", token, err)
			}
			assertEmptyDocument(t, draft, locale)
			published, err := service.GetPublishedDocument(context.Background(), documentStudent, documentLessonID, locale)
			if err != nil {
				t.Fatal(err)
			}
			assertEmptyDocument(t, published, locale)
		})
	}
}

func TestDocumentsPortErrors(t *testing.T) {
	failure := errors.New("dependency failed")
	for _, op := range documentOperations() {
		for _, failAccess := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/access=%t", op.name, failAccess), func(t *testing.T) {
				repo := &documentRepositoryFake{}
				reader := &documentAccessFake{facts: documentFacts()}
				if failAccess {
					reader.err = failure
				} else {
					repo.err = failure
				}
				err := op.call(application.NewDocuments(repo, reader), documentOwner)
				if !errors.Is(err, failure) {
					t.Fatalf("error = %v", err)
				}
				if failAccess && len(repo.calls) != 0 {
					t.Fatal("access failure reached document repository")
				}
			})
		}
	}
}
func assertEmptyDocument(t *testing.T, body domain.LessonDocument, locale string) {
	t.Helper()
	if body.Locale != locale || body.Sections == nil || len(body.Sections) != 0 || domain.ValidateLessonDocument(body) != nil {
		t.Fatalf("invalid empty document: %+v", body)
	}
}
