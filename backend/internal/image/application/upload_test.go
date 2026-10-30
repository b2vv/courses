package application

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"courses/internal/image/domain"
)

func TestUpload(t *testing.T) {
	dependencyErr := errors.New("dependency failed")
	fixedTime := time.Date(2026, time.October, 12, 10, 0, 0, 0, time.UTC)
	newID := func() string { return "11111111-1111-4111-8111-111111111111" }
	now := func() time.Time { return fixedTime }
	for _, tt := range []struct {
		name        string
		actor       domain.Actor
		ownerID     string
		missing     bool
		authorErr   error
		contentType string
		width       int
		height      int
		data        []byte
		storageErr  error
		insertErr   error
		deleteErr   error
		cancelCtx   bool
		wantErr     error
		wantErrAlso error
		wantPut     int
		wantInsert  int
		wantDelete  int
	}{
		{name: "owner", actor: domain.Actor{UserID: "owner"}, wantPut: 1, wantInsert: 1},
		{name: "admin non-owner", actor: domain.Actor{UserID: "admin", Administrator: true}, wantPut: 1, wantInsert: 1},
		{name: "non-owner", actor: domain.Actor{UserID: "other"}, wantErr: domain.ErrForbidden},
		{name: "anonymous with empty owner", ownerID: "empty", wantErr: domain.ErrForbidden},
		{name: "missing lesson", missing: true, wantErr: domain.ErrLessonNotFound},
		{name: "authoring failure", authorErr: dependencyErr, wantErr: dependencyErr},
		{name: "oversized", actor: domain.Actor{UserID: "owner"}, data: make([]byte, domain.MaximumImageBytes+1), wantErr: domain.ErrImageTooLarge},
		{name: "unsupported", actor: domain.Actor{UserID: "owner"}, contentType: "image/gif", wantErr: domain.ErrUnsupportedImage},
		{name: "too many pixels", actor: domain.Actor{UserID: "owner"}, width: domain.MaximumImagePixels + 1, height: 1, wantErr: domain.ErrImageTooManyPixels},
		{name: "invalid dimensions", actor: domain.Actor{UserID: "owner"}, width: -1, wantErr: domain.ErrInvalidImage},
		{name: "empty data", actor: domain.Actor{UserID: "owner"}, data: []byte{}, wantErr: domain.ErrInvalidImage},
		{name: "storage failure", actor: domain.Actor{UserID: "owner"}, storageErr: dependencyErr, wantErr: dependencyErr, wantPut: 1},
		// Insert failed after a successful Put: the uploader compensates by deleting the just-stored
		// bytes so a failed upload does not orphan them (LB09a).
		{name: "insert failure deletes bytes", actor: domain.Actor{UserID: "owner"}, insertErr: dependencyErr, wantErr: dependencyErr, wantPut: 1, wantInsert: 1, wantDelete: 1},
		// When the compensating delete itself fails, both errors surface (joined) and neither masks
		// the other; the strictly smaller residual (bytes still orphaned) is the only cost.
		{name: "insert and delete failure surface both", actor: domain.Actor{UserID: "owner"}, insertErr: dependencyErr, deleteErr: errDeleteFailed, wantErr: dependencyErr, wantErrAlso: errDeleteFailed, wantPut: 1, wantInsert: 1, wantDelete: 1},
		// The compensating delete runs even when the insert failed because the request context was
		// cancelled (client disconnect after Put): it is detached, so it still removes the bytes
		// (LB09a residual N1).
		{name: "insert failure with cancelled request still deletes", actor: domain.Actor{UserID: "owner"}, insertErr: dependencyErr, cancelCtx: true, wantErr: dependencyErr, wantPut: 1, wantInsert: 1, wantDelete: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ownerID := "owner"
			if tt.ownerID == "empty" {
				ownerID = ""
			}
			contentType := tt.contentType
			if contentType == "" {
				contentType = domain.ContentTypePNG
			}
			width, height := tt.width, tt.height
			if width == 0 {
				width = 640
			}
			if height == 0 {
				height = 480
			}
			data := tt.data
			if data == nil {
				data = []byte("image bytes")
			}
			authoring := &fakeAuthoringReader{ownerID: ownerID, found: !tt.missing, err: tt.authorErr}
			storage := &fakeStorage{err: tt.storageErr, deleteErr: tt.deleteErr}
			repository := &fakeRepository{err: tt.insertErr, storage: storage, t: t}
			uploader := NewUploader(repository, storage, authoring, newID, now)
			ctx := context.Background()
			if tt.cancelCtx {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			got, err := uploader.Upload(ctx, tt.actor, "lesson", contentType, width, height, data)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Upload error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErrAlso != nil && !errors.Is(err, tt.wantErrAlso) {
				t.Fatalf("Upload error = %v, want it to also wrap %v", err, tt.wantErrAlso)
			}
			if authoring.calls != 1 || authoring.lessonID != "lesson" {
				t.Fatalf("authoring lookup = %+v", authoring)
			}
			if storage.calls != tt.wantPut || repository.calls != tt.wantInsert {
				t.Fatalf("Put/Insert calls = %d/%d, want %d/%d", storage.calls, repository.calls, tt.wantPut, tt.wantInsert)
			}
			if storage.deleteCalls != tt.wantDelete {
				t.Fatalf("Delete calls = %d, want %d", storage.deleteCalls, tt.wantDelete)
			}
			if tt.wantDelete == 1 {
				wantExt, err := domain.ExtensionFor(contentType)
				if err != nil {
					t.Fatal(err)
				}
				if storage.deletedID != newID() || storage.deletedExt != wantExt {
					t.Fatalf("Delete received id=%q ext=%q; want id=%q ext=%q", storage.deletedID, storage.deletedExt, newID(), wantExt)
				}
				// The compensating delete must run on a context detached from the request, so a
				// cancelled request does not make it a no-op (LB09a residual N1).
				if storage.deletedCtxErr != nil {
					t.Fatalf("Delete ran on a cancelled context (%v); want a detached, live context", storage.deletedCtxErr)
				}
			}
			want := domain.Image{ID: newID(), LessonID: "lesson", ContentType: contentType, Width: width, Height: height, ByteSize: int64(len(data)), CreatedBy: tt.actor.UserID, CreatedAt: fixedTime}
			if tt.wantErr != nil {
				if got != (domain.Image{}) {
					t.Fatalf("failed Upload returned image: %+v", got)
				}
			} else if got != want {
				t.Fatalf("Upload = %+v, want %+v", got, want)
			}
			if tt.wantPut == 1 {
				wantExt, err := domain.ExtensionFor(contentType)
				if err != nil {
					t.Fatal(err)
				}
				if storage.id != want.ID || storage.extension != wantExt || !bytes.Equal(storage.data, data) {
					t.Fatalf("storage received id=%q ext=%q bytes=%d; want id=%q ext=%q", storage.id, storage.extension, len(storage.data), want.ID, wantExt)
				}
			}
			if tt.wantInsert == 1 && repository.image != want {
				t.Fatalf("inserted image = %+v, want %+v", repository.image, want)
			}
		})
	}
}

type fakeAuthoringReader struct {
	ownerID, lessonID string
	found             bool
	err               error
	calls             int
}

func (f *fakeAuthoringReader) LoadLessonOwner(_ context.Context, lessonID string) (string, bool, error) {
	f.calls++
	f.lessonID = lessonID
	return f.ownerID, f.found, f.err
}

var errDeleteFailed = errors.New("delete failed")

type fakeStorage struct {
	id            string
	extension     string
	data          []byte
	err           error
	calls         int
	deleteErr     error
	deletedID     string
	deletedExt    string
	deleteCalls   int
	deletedCtxErr error
}

func (f *fakeStorage) Put(_ context.Context, id, extension string, r io.Reader) (int64, error) {
	f.calls++
	f.id = id
	f.extension = extension
	data, err := io.ReadAll(r)
	if err != nil {
		return 0, err
	}
	f.data = data
	return int64(len(data)), f.err
}

func (f *fakeStorage) Delete(ctx context.Context, id, extension string) error {
	f.deleteCalls++
	f.deletedID = id
	f.deletedExt = extension
	f.deletedCtxErr = ctx.Err()
	return f.deleteErr
}

type fakeRepository struct {
	image   domain.Image
	err     error
	calls   int
	storage *fakeStorage
	t       *testing.T
}

func (f *fakeRepository) Insert(_ context.Context, image domain.Image) error {
	f.t.Helper()
	if f.storage.calls != 1 || f.storage.err != nil {
		f.t.Fatal("Insert called before successful Put")
	}
	f.calls++
	f.image = image
	return f.err
}

func (f *fakeRepository) FindByID(context.Context, string) (domain.Image, error) {
	f.t.Fatal("Upload must not call FindByID")
	return domain.Image{}, domain.ErrImageNotFound
}
