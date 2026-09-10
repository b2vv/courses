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
	"sync"
	"testing"
	"time"

	mediafilesystem "courses/internal/media/adapters/filesystem"
	mediahttp "courses/internal/media/adapters/http"
	mediapostgres "courses/internal/media/adapters/postgres"
	"courses/internal/media/adapters/tus"
	"courses/internal/media/application"
	"courses/internal/media/domain"
	"courses/internal/media/ports"
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

type tusJob struct {
	assetID string
	attempt int
}

type tusQueue struct{ jobs []tusJob }

func (queue *tusQueue) Enqueue(_ context.Context, assetID string, attempt int) error {
	queue.jobs = append(queue.jobs, tusJob{assetID: assetID, attempt: attempt})
	return nil
}

type tusStorage struct {
	ports.MediaStorage
	free       *int64
	beforeHead func(context.Context) error
}

func (storage *tusStorage) Free(ctx context.Context) (int64, error) {
	if storage.free != nil {
		return *storage.free, nil
	}
	return storage.MediaStorage.Free(ctx)
}

func (storage *tusStorage) Head(ctx context.Context, key domain.StorageKey, limit int) ([]byte, error) {
	if storage.beforeHead != nil {
		if err := storage.beforeHead(ctx); err != nil {
			return nil, err
		}
	}
	return storage.MediaStorage.Head(ctx, key, limit)
}

func TestTusServerAgainstPostgreSQL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	container, err := postgrescontainer.Run(ctx, "postgres:17.5",
		postgrescontainer.WithDatabase("tus_test"),
		postgrescontainer.WithUsername("courses"),
		postgrescontainer.WithPassword("ephemeral-test-only"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(3*time.Minute)),
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

	t.Run("P03 HEAD then PATCH lands the source and queues one job", harness.testCompleteUpload)
	t.Run("B18 the first chunk rejects an unsupported container", harness.testUnsupportedContainer)
	t.Run("B35 an unsupported first chunk consumes the ticket before a waiting PATCH", harness.testConcurrentUnsupportedContainer)
	t.Run("B23 a full volume refuses the first byte", harness.testDiskGate)
	t.Run("B31 a post-rename crash is repaired without another append", harness.testPostRenameCrashRepair)
	t.Run("B34 a post-fsync crash replays the terminal chunk", harness.testPostFsyncCrashRepair)
	t.Run("B02 a chunk beyond the declared length is refused", harness.testDeclaredLengthCap)
	t.Run("B03 a crash-mid-chunk file is repaired to the column", harness.testCrashMidChunkRepair)
	t.Run("R10 unavailable tickets write nothing", harness.testUnavailableTickets)
	t.Run("R11 a wrong offset writes nothing", harness.testWrongOffset)
	t.Run("R29 concurrent PATCHes serialize without corruption", harness.testConcurrentPatch)
	t.Run("R30 DELETE frees the lesson immediately", harness.testDeleteFreesLesson)
	t.Run("R40 a chunked body cannot pass the server chunk cap", harness.testChunkedBodyCap)
	t.Run("three-verb tus surface and protocol headers", harness.testProtocolSurface)
}

func (h *tusHarness) testPostRenameCrashRepair(t *testing.T) {
	h.reset(t)
	source := []byte("\x00\x00\x00\x18ftypisomfinal-chunk!")
	const received = int64(12)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, int64(len(source)), received, tusNow.Add(time.Hour), nil)
	storage, queue, handler := h.server(t)
	if _, err := storage.Append(h.ctx, sourceKey(tusAssetOne), 0, bytes.NewReader(source)); err != nil {
		t.Fatal(err)
	}

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, received, bytes.NewReader(source[received:]), false)
	if response.Code != http.StatusNoContent || response.Header().Get("Upload-Offset") != strconv.Itoa(len(source)) {
		t.Fatalf("PATCH status=%d offset=%q body=%s", response.Code, response.Header().Get("Upload-Offset"), response.Body.String())
	}
	assertStoredBytes(t, storage, sourceKey(tusAssetOne), source)
	if _, err := storage.Size(h.ctx, scratchKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("scratch err=%v, want not-exist", err)
	}
	h.assertAsset(t, tusAssetOne, domain.Queued, int64(len(source)), "")
	if len(queue.jobs) != 1 || queue.jobs[0] != (tusJob{assetID: tusAssetOne, attempt: 0}) {
		t.Fatalf("jobs=%v", queue.jobs)
	}
}

func (h *tusHarness) testPostFsyncCrashRepair(t *testing.T) {
	h.reset(t)
	source := []byte("\x00\x00\x00\x18ftypisomfinal-chunk!")
	const received = int64(12)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, int64(len(source)), received, tusNow.Add(time.Hour), nil)
	storage, queue, handler := h.server(t)
	h.writeScratch(t, storage, tusAssetOne, source)

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, received, bytes.NewReader(source[received:]), false)
	if response.Code != http.StatusNoContent || response.Header().Get("Upload-Offset") != strconv.Itoa(len(source)) {
		t.Fatalf("PATCH status=%d offset=%q body=%s", response.Code, response.Header().Get("Upload-Offset"), response.Body.String())
	}
	assertStoredBytes(t, storage, sourceKey(tusAssetOne), source)
	if _, err := storage.Size(h.ctx, scratchKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("scratch err=%v, want not-exist", err)
	}
	h.assertAsset(t, tusAssetOne, domain.Queued, int64(len(source)), "")
	if len(queue.jobs) != 1 || queue.jobs[0] != (tusJob{assetID: tusAssetOne, attempt: 0}) {
		t.Fatalf("jobs=%v", queue.jobs)
	}
}

func (h *tusHarness) testUnsupportedContainer(t *testing.T) {
	h.reset(t)
	source := []byte("plain text!!")
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, int64(len(source)), 0, tusNow.Add(time.Hour), nil)
	storage, queue, handler := h.server(t)

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader(source), false)
	if response.Code != http.StatusBadRequest || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"unsupported_media"`)) {
		t.Fatalf("PATCH status=%d body=%s", response.Code, response.Body.String())
	}
	h.assertAsset(t, tusAssetOne, domain.Failed, 0, domain.UnsupportedSource)
	if len(queue.jobs) != 0 {
		t.Fatalf("jobs=%v", queue.jobs)
	}
	if _, err := storage.Size(h.ctx, sourceKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source err=%v, want not-exist", err)
	}
}

func (h *tusHarness) testConcurrentUnsupportedContainer(t *testing.T) {
	h.reset(t)
	invalid := []byte("plain text!!")
	valid := []byte("\x00\x00\x00\x18ftypisom")
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, int64(len(valid)), 0, tusNow.Add(time.Hour), nil)
	enteredPolicy := make(chan struct{})
	releasePolicy := make(chan struct{})
	var enterOnce sync.Once
	var releaseOnce sync.Once
	t.Cleanup(func() { releaseOnce.Do(func() { close(releasePolicy) }) })
	storage, queue, handler := h.serverConfigured(t, func(storage *tusStorage) {
		storage.beforeHead = func(ctx context.Context) error {
			var waitErr error
			enterOnce.Do(func() {
				close(enteredPolicy)
				select {
				case <-releasePolicy:
				case <-ctx.Done():
					waitErr = ctx.Err()
				}
			})
			return waitErr
		}
	})

	invalidResult := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		invalidResult <- serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader(invalid), false)
	}()
	select {
	case <-enteredPolicy:
	case response := <-invalidResult:
		t.Fatalf("invalid PATCH completed before entering its policy barrier: status=%d body=%s", response.Code, response.Body.String())
	case <-time.After(5 * time.Second):
		t.Fatal("invalid PATCH never entered its policy barrier")
	}

	validResult := make(chan *httptest.ResponseRecorder, 1)
	go func() {
		validResult <- serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader(valid), false)
	}()
	h.waitForBlockedAppend(t)
	select {
	case response := <-validResult:
		t.Fatalf("valid PATCH completed before invalid policy was released: status=%d body=%s", response.Code, response.Body.String())
	default:
	}
	releaseOnce.Do(func() { close(releasePolicy) })

	invalidResponse := <-invalidResult
	if invalidResponse.Code != http.StatusBadRequest || !bytes.Contains(invalidResponse.Body.Bytes(), []byte(`"code":"unsupported_media"`)) {
		t.Fatalf("invalid PATCH status=%d body=%s", invalidResponse.Code, invalidResponse.Body.String())
	}
	validResponse := <-validResult
	if validResponse.Code == http.StatusNoContent {
		t.Fatalf("valid PATCH advanced the failed ticket: status=%d body=%s", validResponse.Code, validResponse.Body.String())
	}
	h.assertAsset(t, tusAssetOne, domain.Failed, 0, domain.UnsupportedSource)
	if len(queue.jobs) != 0 {
		t.Fatalf("jobs=%v", queue.jobs)
	}
	if _, err := storage.Size(h.ctx, sourceKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("source err=%v, want not-exist", err)
	}
}

func (h *tusHarness) testDiskGate(t *testing.T) {
	h.reset(t)
	source := []byte("\x00\x00\x00\x18ftypisom")
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, int64(len(source)), 0, tusNow.Add(time.Hour), nil)
	storage, queue, handler := h.server(t, 3*int64(len(source))-1)

	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader(source), false)
	if response.Code != http.StatusInsufficientStorage || !bytes.Contains(response.Body.Bytes(), []byte(`"code":"insufficient_storage"`)) {
		t.Fatalf("PATCH status=%d body=%s", response.Code, response.Body.String())
	}
	h.assertAsset(t, tusAssetOne, domain.Created, 0, "")
	if len(queue.jobs) != 0 {
		t.Fatalf("jobs=%v", queue.jobs)
	}
	if _, err := storage.Size(h.ctx, scratchKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("scratch err=%v, want not-exist", err)
	}
}

func (h *tusHarness) testCompleteUpload(t *testing.T) {
	h.reset(t)
	source := []byte("\x00\x00\x00\x18ftypisom")
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, int64(len(source)), 0, tusNow.Add(time.Hour), nil)
	storage, queue, handler := h.server(t)

	head := serveTus(handler, http.MethodHead, "/uploads/"+tusAssetOne, 0, nil, false)
	if head.Code != http.StatusOK || head.Header().Get("Upload-Offset") != "0" {
		t.Fatalf("HEAD status=%d headers=%v body=%s", head.Code, head.Header(), head.Body.String())
	}
	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, 0, bytes.NewReader(source), false)
	if response.Code != http.StatusNoContent || response.Header().Get("Upload-Offset") != strconv.Itoa(len(source)) {
		t.Fatalf("PATCH status=%d offset=%q body=%s", response.Code, response.Header().Get("Upload-Offset"), response.Body.String())
	}
	assertStoredBytes(t, storage, sourceKey(tusAssetOne), source)
	if _, err := storage.Size(h.ctx, scratchKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("scratch err=%v, want not-exist", err)
	}
	h.assertAsset(t, tusAssetOne, domain.Queued, int64(len(source)), "")
	if len(queue.jobs) != 1 || queue.jobs[0] != (tusJob{assetID: tusAssetOne, attempt: 0}) {
		t.Fatalf("jobs=%v", queue.jobs)
	}
}

func (h *tusHarness) testDeclaredLengthCap(t *testing.T) {
	h.reset(t)
	const declared = int64(1024 * 1024)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Uploading, declared, declared, tusNow.Add(time.Hour), nil)
	storage, _, handler := h.server(t)
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
	storage, queue, handler := h.server(t)
	h.writeScratch(t, storage, tusAssetOne, append(append([]byte(nil), prefix...), crashed...))

	assertHeadOffset(t, handler, tusAssetOne, int64(len(prefix)))
	response := serveTus(handler, http.MethodPatch, "/uploads/"+tusAssetOne, int64(len(prefix)), bytes.NewReader(remainder), false)
	if response.Code != http.StatusNoContent || response.Header().Get("Upload-Offset") != strconv.Itoa(len(want)) || response.Header().Get("Upload-Expires") != tusNow.Add(time.Hour).Format(http.TimeFormat) {
		t.Fatalf("PATCH status=%d offset=%q body=%s", response.Code, response.Header().Get("Upload-Offset"), response.Body.String())
	}
	assertStoredBytes(t, storage, sourceKey(tusAssetOne), want)
	if _, err := storage.Size(h.ctx, scratchKey(tusAssetOne)); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("scratch err=%v, want not-exist", err)
	}
	h.assertReceived(t, tusAssetOne, int64(len(want)))
	if len(queue.jobs) != 1 {
		t.Fatalf("jobs=%v", queue.jobs)
	}
}

func (h *tusHarness) testUnavailableTickets(t *testing.T) {
	h.reset(t)
	evictedAt := tusNow.Add(-time.Minute)
	h.insertAsset(t, tusAssetOne, tusLessonOne, domain.Created, 12, 0, tusNow.Add(time.Hour), &evictedAt)
	h.insertAsset(t, tusAssetTwo, tusLessonOne, domain.Created, 12, 0, tusNow.Add(-time.Second), nil)
	storage, _, handler := h.server(t)

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
	storage, _, handler := h.server(t)
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
	storage, _, handler := h.server(t)
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
	_, _, handler := h.server(t)

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
	storage, _, handler := h.server(t)
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
	storage, _, handler := h.server(t)
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

func (h *tusHarness) server(t *testing.T, free ...int64) (*mediafilesystem.Storage, *tusQueue, http.Handler) {
	return h.serverConfigured(t, func(storage *tusStorage) {
		if len(free) == 1 {
			storage.free = &free[0]
		}
	})
}

func (h *tusHarness) serverConfigured(t *testing.T, configure func(*tusStorage)) (*mediafilesystem.Storage, *tusQueue, http.Handler) {
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
	queue := &tusQueue{}
	mediaStorage := &tusStorage{MediaStorage: storage}
	configure(mediaStorage)
	upload := application.NewUpload(h.repository, mediaStorage, queue, func() time.Time { return tusNow })
	tickets := application.NewTickets(h.repository, nil, func() time.Time { return tusNow })
	uploadServer := tus.NewServer(upload)
	mediaHandler := mediahttp.NewHandler(tickets)
	return storage, queue, httpserver.NewHandler(httpserver.Options{Service: "upload", Register: func(mux *http.ServeMux) {
		uploadServer.Register(mux)
		mediaHandler.RegisterAdmin(mux, func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				next.ServeHTTP(w, r.WithContext(httpserver.WithActor(r.Context(), httpserver.Actor{UserID: tusInstructorID, Role: "admin"})))
			})
		}, func(next http.Handler) http.Handler { return next })
	}})
}

func (h *tusHarness) waitForBlockedAppend(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(h.ctx, 5*time.Second)
	defer cancel()
	for {
		var waiting bool
		err := h.pool.QueryRow(ctx, `SELECT EXISTS (
			SELECT 1 FROM pg_stat_activity
			WHERE datname=current_database() AND pid<>pg_backend_pid()
				AND wait_event_type='Lock' AND query LIKE '%FOR UPDATE OF assets%'
		)`).Scan(&waiting)
		if err != nil {
			t.Fatal(err)
		}
		if waiting {
			return
		}
		select {
		case <-ctx.Done():
			t.Fatal("valid PATCH never waited on the invalid PATCH row lock")
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func (h *tusHarness) assertAsset(t *testing.T, id string, state domain.State, received int64, reason domain.FailureReason) {
	t.Helper()
	var gotState string
	var gotReceived int64
	var gotReason *string
	if err := h.pool.QueryRow(h.ctx, `SELECT state, received_bytes, failure_reason FROM video_assets WHERE id=$1`, id).Scan(&gotState, &gotReceived, &gotReason); err != nil {
		t.Fatal(err)
	}
	if domain.State(gotState) != state || gotReceived != received || reason == "" && gotReason != nil || reason != "" && (gotReason == nil || *gotReason != string(reason)) {
		t.Fatalf("asset state=%s received=%d reason=%v, want state=%s received=%d reason=%s", gotState, gotReceived, gotReason, state, received, reason)
	}
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
	assertStoredBytes(t, storage, scratchKey(id), want)
}

func assertStoredBytes(t *testing.T, storage *mediafilesystem.Storage, key domain.StorageKey, want []byte) {
	t.Helper()
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

func sourceKey(id string) domain.StorageKey {
	return domain.StorageKey{Kind: domain.StorageSource, ID: id, Leaf: "original"}
}
