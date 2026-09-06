//go:build integration

package tus_test

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	mediafilesystem "courses/internal/media/adapters/filesystem"
	mediahttp "courses/internal/media/adapters/http"
	mediapostgres "courses/internal/media/adapters/postgres"
	"courses/internal/media/adapters/tus"
	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/platform/httpserver"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	postgrescontainer "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	tusInstructorID = "43000000-0000-4000-8000-000000000001"
	tusCourseID     = "13000000-0000-4000-8000-000000000001"
	tusModuleID     = "63000000-0000-4000-8000-000000000001"
	tusLessonOne    = "73000000-0000-4000-8000-000000000001"
	tusAssetOne     = "83000000-0000-4000-8000-000000000001"
	tusAssetTwo     = "83000000-0000-4000-8000-000000000002"
	tusUnknownAsset = "83000000-0000-4000-8000-000000000099"
	tusVersion      = "1.0.0"
)

var tusNow = time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC)

type tusHarness struct {
	ctx        context.Context
	pool       *pgxpool.Pool
	repository *mediapostgres.Repository
}

func TestTusServerAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("tus_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(time.Minute)),
	)
	if err != nil {
		t.Fatalf("tus integration requires Docker: %v", err)
	}
	t.Cleanup(func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Errorf("terminate test container: %v", err)
		}
	})
	connection, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	database, err := sql.Open("pgx", connection)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	migrations := filepath.Join("..", "..", "..", "..", "migrations")
	provider, err := goose.NewProvider(goose.DialectPostgres, database, os.DirFS(migrations), goose.WithDisableGlobalRegistry(true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := provider.Up(ctx); err != nil {
		t.Fatal(err)
	}
	pool, err := pgxpool.New(ctx, connection)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	harness := &tusHarness{ctx: ctx, pool: pool, repository: mediapostgres.NewRepository(pool)}

	t.Run("B02 a chunk beyond the declared length is refused", harness.testDeclaredLengthCap)
	t.Run("B03 a crash-mid-chunk file is repaired to the column", harness.testCrashMidChunkRepair)
	t.Run("R10 unavailable tickets write nothing", harness.testUnavailableTickets)
	t.Run("R11 a wrong offset writes nothing", harness.testWrongOffset)
	t.Run("R29 concurrent PATCHes serialize without corruption", harness.testConcurrentPatch)
	t.Run("R30 DELETE frees the lesson immediately", harness.testDeleteFreesLesson)
	t.Run("R40 a chunked body cannot pass the server chunk cap", harness.testChunkedBodyCap)
	t.Run("three-verb tus surface and protocol headers", harness.testProtocolSurface)
}

func (h *tusHarness) testDeclaredLengthCap(t *testing.T) {
	h.reset(t)
	const declared = int64(1024 * 1024)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, declared, declared, tusNow.Add(time.Hour), nil)
	storage, handler := h.server(t)
	want := bytes.Repeat([]byte("a"), int(declared))
	h.writeScratch(t, storage, tusAssetOne, want)

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, declared, bytes.NewReader([]byte("x")), false)
	if response.Code != http.StatusRequestEntityTooLarge || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"request_too_large"`)) {
		t.Fatalf("PATCH status=%d body=%s", response.Code, response.Body.String())
	}
	assertScratch(t, storage, tusAssetOne, want)
	assertHeadOffset(t, handler, tusAssetOne, declared)
	h.assertReceived(t, tusAssetOne, declared)
}

func (h *tusHarness) testCrashMidChunkRepair(t *testing.T) {
	h.reset(t)
	prefix := []byte("already-committed-")
	crashed := []byte("uncommitted-chunk")
	remainder := []byte("replacement-tail")
	want := append(append([]byte(nil), prefix...), remainder...)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, int64(len(want)), int64(len(prefix)), tusNow.Add(time.Hour), nil)
	storage, handler := h.server(t)
	h.writeScratch(t, storage, tusAssetOne, append(append([]byte(nil), prefix...), crashed...))

	assertHeadOffset(t, handler, tusAssetOne, int64(len(prefix)))
	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, int64(len(prefix)), bytes.NewReader(remainder), false)
	if response.Code != http.StatusNoContent || response.Header().Get("Upload-Offset") != strconv.Itoa(len(want)) || response.Header().Get("Upload-Expires") != tusNow.Add(time.Hour).Format(http.TimeFormat) {
		t.Fatalf("PATCH status=%d offset=%q body=%s", response.Code, response.Header().Get("Upload-Offset"), response.Body.String())
	}
	assertScratch(t, storage, tusAssetOne, want)
	h.assertReceived(t, tusAssetOne, int64(len(want)))
}

func (h *tusHarness) testUnavailableTickets(t *testing.T) {
	h.reset(t)
	evictedAt := tusNow.Add(-time.Minute)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, 12, 0, tusNow.Add(time.Hour), &evictedAt)
	h.insertAsset(t, tusAssetTwo, tusLessonOne, domain.Created, 12, 0, tusNow.Add(-time.Second), nil)
	storage, handler := h.server(t)

	for _, id := range []string{tusUnknownAsset, tusAssetOne, tusAssetTwo} {
		response := serveTus(handler, http.MethodPatch, "/uploads/"+id, 0, bytes.NewReader([]byte("x")), false)
		if response.Code != http.StatusNotFound {
			t.Fatalf("id=%s status=%d body=%s", id, response.Code, response.Body.String())
		}
		if _, err := storage.Size(h.ctx, scratchKey(id)); !errors.Is(err, os.ErrNotExist) {
			t.Fatalf("id=%s scratch err=%v, want not-exist", id, err)
		}
	}
	h.assertReceived(t, tusAssetOne, 0)
	h.assertReceived(t, tusAssetTwo, 0)
}

func (h *tusHarness) testWrongOffset(t *testing.T) {
	h.reset(t)
	want := []byte("committed-data")
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, 128, int64(len(want)), tusNow.Add(time.Hour), nil)
	storage, handler := h.server(t)
	h.writeScratch(t, storage, tusAssetOne, want)

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader([]byte("loser")), false)
	if response.Code != http.StatusConflict {
		t.Fatalf("PATCH status=%d body=%s", response.Code, response.Body.String())
	}
	assertScratch(t, storage, tusAssetOne, want)
	assertHeadOffset(t, handler, tusAssetOne, int64(len(want)))
}

func (h *tusHarness) testConcurrentPatch(t *testing.T) {
	h.reset(t)
	prefix := []byte("committed-")
	chunk := []byte("one-chunk")
	want := append(append([]byte(nil), prefix...), chunk...)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, 128, int64(len(prefix)), tusNow.Add(time.Hour), nil)
	storage, handler := h.server(t)
	h.writeScratch(t, storage, tusAssetOne, prefix)

	start := make(chan struct{})
	results := make(chan *httptest.ResponseRecorder, 2)
	for range 2 {
		go func() {
			<-start
			results <- serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, int64(len(prefix)), bytes.NewReader(chunk), false)
		}()
	}
	close(start)
	accepted, conflicted := 0, 0
	for range 2 {
		response := <-results
		switch response.Code {
		case http.StatusNoContent:
			accepted++
		case http.StatusConflict:
			conflicted++
		default:
			t.Fatalf("concurrent PATCH status=%d body=%s", response.Code, response.Body.String())
		}
	}
	if accepted != 1 || conflicted != 1 {
		t.Fatalf("accepted=%d conflicted=%d", accepted, conflicted)
	}
	assertScratch(t, storage, tusAssetOne, want)
	h.assertReceived(t, tusAssetOne, int64(len(want)))
}

func (h *tusHarness) testDeleteFreesLesson(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, 12, 0, tusNow.Add(time.Hour), nil)
	_, handler := h.server(t)

	response := serveTus(handler, http.MethodDelete, "/uploads/"+tusAssetOne, 0, nil, false)
	if response.Code != http.StatusNoContent {
		t.Fatalf("DELETE status=%d body=%s", response.Code, response.Body.String())
	}
	response = serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader([]byte("x")), false)
	if response.Code != http.StatusNotFound {
		t.Fatalf("PATCH after DELETE status=%d body=%s", response.Code, response.Body.String())
	}
	var evicted bool
	if err := h.pool.QueryRow(h.ctx, `SELECT evicted_at IS NOT NULL FROM video_assets WHERE id=$1`, tusAssetOne).Scan(&evicted); err != nil || !evicted {
		t.Fatalf("evicted=%t err=%v", evicted, err)
	}

	body := bytes.NewBufferString(`{"declaredBytes":12,"originalName":"replacement.mp4"}`)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/lessons/"+tusLessonOne+"/video", body)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("mint after DELETE status=%d body=%s", response.Code, response.Body.String())
	}
}

func (h *tusHarness) testChunkedBodyCap(t *testing.T) {
	h.reset(t)
	const received = int64(1)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, 16*1024*1024, received, tusNow.Add(time.Hour), nil)
	storage, handler := h.server(t)
	want := []byte("a")
	h.writeScratch(t, storage, tusAssetOne, want)
	body := bytes.NewReader(bytes.Repeat([]byte("b"), 9*1024*1024))

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, received, body, true)
	if response.Code != http.StatusRequestEntityTooLarge || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"request_too_large"`)) {
		t.Fatalf("chunked PATCH status=%d body=%s", response.Code, response.Body.String())
	}
	assertScratch(t, storage, tusAssetOne, want)
	assertHeadOffset(t, handler, tusAssetOne, received)
	h.assertReceived(t, tusAssetOne, received)
}

func (h *tusHarness) testProtocolSurface(t *testing.T) {
	h.reset(t)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, 12, 1, tusNow.Add(time.Hour), nil)
	storage, handler := h.server(t)
	h.writeScratch(t, storage, tusAssetOne, []byte("a"))

	head := serveTus(handler, http.MethodHead, "/uploads/"+tusAssetOne, 0, nil, false)
	if head.Code != http.StatusOK || head.Header().Get("Upload-Offset") != "1" || head.Header().Get("Upload-Length") != "12" || head.Header().Get("Upload-Expires") != tusNow.Add(time.Hour).Format(http.TimeFormat) {
		t.Fatalf("HEAD status=%d headers=%v", head.Code, head.Header())
	}
	request := httptest.NewRequest(http.MethodHead, "/uploads/"+tusAssetOne, nil)
	missingVersion := httptest.NewRecorder()
	handler.ServeHTTP(missingVersion, request)
	if missingVersion.Code != http.StatusPreconditionFailed {
		t.Fatalf("missing version status=%d", missingVersion.Code)
	}

	wrongType := httptest.NewRequest(http.MethodPatch, "/uploads/"+tusAssetOne, bytes.NewReader([]byte("x")))
	wrongType.Header.Set("Tus-Resumable", tusVersion)
	wrongType.Header.Set("Upload-Offset", "1")
	wrongType.Header.Set("Content-Type", "application/octet-stream")
	wrongTypeResponse := httptest.NewRecorder()
	handler.ServeHTTP(wrongTypeResponse, wrongType)
	if wrongTypeResponse.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("wrong content type status=%d", wrongTypeResponse.Code)
	}
	for _, value := range []string{"", "+1", "-0", "1.0"} {
		request := httptest.NewRequest(http.MethodPatch, "/uploads/"+tusAssetOne, bytes.NewReader([]byte("x")))
		request.Header.Set("Tus-Resumable", tusVersion)
		request.Header.Set("Content-Type", "application/offset+octet-stream")
		request.Header.Set("Upload-Offset", value)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("Upload-Offset %q status=%d", value, response.Code)
		}
	}

	for _, method := range []string{http.MethodPost, http.MethodOptions} {
		response := serveTus(handler, method, "/uploads/"+tusAssetOne, 0, nil, false)
		if response.Code != http.StatusMethodNotAllowed {
			t.Fatalf("%s status=%d body=%s", method, response.Code, response.Body.String())
		}
	}
}

func (h *tusHarness) server(t *testing.T) (*mediafilesystem.Storage, http.Handler) {
	t.Helper()
	storage, err := mediafilesystem.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := storage.Close(); err != nil {
			t.Errorf("close storage: %v", err)
		}
	})
	upload := application.NewUpload(h.repository, storage, nil, func() time.Time { return tusNow })
	tickets := application.NewTickets(h.repository, nil, func() time.Time { return tusNow })
	uploadServer := tus.NewServer(upload)
	mediaHandler := mediahttp.NewHandler(tickets)
	return storage, httpserver.NewHandler(httpserver.Options{Service: "upload", Register: func(mux *http.ServeMux) {
		uploadServer.Register(mux)
		mediaHandler.RegisterAdmin(mux, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: tusInstructorID, Role: "admin"})))
			})
		}, func(next http.Handler) http.Handler { return next })
	}})
}

func (h *tusHarness) reset(t *testing.T) {
	t.Helper()
	statements := []struct {
		sql       string
		arguments []any
	}{
		{`TRUNCATE video_assets, course_lessons, course_modules, courses, users CASCADE`, nil},
		{`INSERT INTO users(id,email,display_name,role,password_hash,created_at,updated_at) VALUES
		($1,'tus@courseforge.local','Tus Instructor','instructor','$argon2id$v=19$m=65536,t=3,p=2$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA',$2,$2)`, []any{tusInstructorID, tusNow}},
		{`INSERT INTO courses(id,instructor_id,slug,title,summary,description,status,created_at,updated_at) VALUES
		($1,$2,'tus-course','Tus course','Tus summary','Tus description','draft',$3,$3)`, []any{tusCourseID, tusInstructorID, tusNow}},
		{`INSERT INTO course_modules(id,course_id,title,position,created_at,updated_at) VALUES
		($1,$2,'Tus module',0,$3,$3)`, []any{tusModuleID, tusCourseID, tusNow}},
		{`INSERT INTO course_lessons(id,module_id,title,description,position,created_at,updated_at) VALUES
		($1,$2,'Lesson one','',0,$3,$3)`, []any{tusLessonOne, tusModuleID, tusNow}},
	}
	for _, statement := range statements {
		if _, err := h.pool.Exec(h.ctx, statement.sql, statement.arguments...); err != nil {
			t.Fatal(err)
		}
	}
}

func (h *tusHarness) insertAsset(t *testing.T, id, lessonID string, state domain.State, declared, received int64, expiresAt time.Time, evictedAt *time.Time) {
	t.Helper()
	if _, err := h.pool.Exec(h.ctx, `INSERT INTO video_assets(
		id,lesson_id,generation,state,declared_bytes,received_bytes,original_name,
		expires_at,evicted_at,created_at,updated_at)
		VALUES($1,$2::text::uuid,(SELECT coalesce(max(generation),0)+1 FROM video_assets WHERE lesson_id=$2::text::uuid),
			$3,$4,$5,'fixture.mp4',$6,$7,$8,$8)`,
		id, lessonID, string(state), declared, received, expiresAt, evictedAt, tusNow.Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
}

func (h *tusHarness) writeScratch(t *testing.T, storage *mediafilesystem.Storage, id string, value []byte) {
	t.Helper()
	if _, err := storage.Append(h.ctx, scratchKey(id), 0, bytes.NewReader(value)); err != nil {
		t.Fatal(err)
	}
}

func (h *tusHarness) assertReceived(t *testing.T, id string, want int64) {
	t.Helper()
	var received int64
	if err := h.pool.QueryRow(h.ctx, `SELECT received_bytes FROM video_assets WHERE id=$1`, id).Scan(&received); err != nil {
		t.Fatal(err)
	}
	if received != want {
		t.Fatalf("received_bytes=%d want=%d", received, want)
	}
}

func serveTus(handler http.Handler, method, path string, offset int64, body io.Reader, chunked bool) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, body)
	request.Header.Set("Tus-Resumable", tusVersion)
	if method == http.MethodPatch {
		request.Header.Set("Content-Type", "application/offset+octet-stream")
		request.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
	}
	if chunked {
		request.ContentLength = -1
		request.TransferEncoding = []string{"chunked"}
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)
	return response
}

func assertHeadOffset(t *testing.T, handler http.Handler, id string, want int64) {
	t.Helper()
	response := serveTus(handler, http.MethodHead, "/uploads/"+id, 0, nil, false)
	if response.Code != http.StatusOK || response.Header().Get("Tus-Resumable") != tusVersion || response.Header().Get("Upload-Offset") != strconv.FormatInt(want, 10) {
		t.Fatalf("HEAD status=%d headers=%v body=%s", response.Code, response.Header(), response.Body.String())
	}
}

func assertScratch(t *testing.T, storage *mediafilesystem.Storage, id string, want []byte) {
	t.Helper()
	key := scratchKey(id)
	size, err := storage.Size(context.Background(), key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := storage.Head(context.Background(), key, len(want)+1)
	if err != nil {
		t.Fatal(err)
	}
	if size != int64(len(want)) || !bytes.Equal(got, want) {
		t.Fatalf("scratch size=%d bytes=%q want=%q", size, got, want)
	}
}

func scratchKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageTemporary, ID: id, Leaf: "upload"}
}
