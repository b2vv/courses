# C01–C02 — Course backend evidence

Date: 2026-09-03. Phase 1 gate was passed before starting this work (parent commit `5608a03`). Scope is the public Course slice; no authentication, administration, media processing or payments are claimed.

## 1. Implemented behavior

`GET /api/v1/courses` lists only published courses with normalized search, stable sorting and `{items,pagination:{limit,offset,total}}`. `GET /api/v1/courses/{slug}` returns a published course; draft, archived and unknown slugs all return the same safe 404.

Queries default to limit 20, offset 0, empty search and recent sort. Application validation enforces limit 1–50, offset 0–10000, trimmed search at most 100 Unicode code points, and sort `recent`/`title`. Invalid UTF-8 and NUL search input are rejected before PostgreSQL. Slugs use lowercase ASCII letters/digits separated by single hyphens, at most 120 characters. Repeated, unknown, malformed or empty numeric query parameters receive 400. Reads have a three-second application deadline and preserve caller cancellation.

## 2. Files and modules

- `backend/internal/course/domain`: validated Course, status, query/page values and explicit error codes; standard-library dependencies only.
- `backend/internal/course/application`: ListPublishedCourses/GetPublishedCourse with CourseRepository injection and bounded read contexts.
- `backend/internal/course/ports/repository.go`: the currently used persistence contract.
- `backend/internal/course/adapters/postgres`: pgx repository and explicit generated-model-to-domain mapping.
- `backend/internal/course/adapters/postgres/generated`: actual sqlc v1.30.0 output from `backend/sql/queries/courses.sql` and `backend/sqlc.yaml`.
- `backend/migrations/00001_courses.sql`: Course table, unique/canonical slug, required text bounds, status/publication/timestamp constraints and partial published-list indexes.
- `backend/sql/seed.sql`: explicit transaction containing three published Ukrainian demo courses, one draft and one archived record, stable UUIDs and `ON CONFLICT DO NOTHING`.
- `backend/internal/course/adapters/http`: transport parsing, DTO mapping, standardized errors and HTTP/OpenAPI contract tests.
- `backend/openapi/openapi.yaml`: OpenAPI 3.0.3 with `listPublishedCourses` and `getPublishedCourse`, pagination, Request ID headers, errors and a reserved future Idempotency-Key component.
- `backend/cmd/api/main.go`: composition of PostgreSQL repository, use cases and HTTP handlers.

## 3. Decisions

Course contains only id, slug, title, summary, description, status, createdAt, updatedAt and nullable publishedAt. No invented price, instructor, duration or lesson-count data was added. Text limits are title 200, summary 500 and description 20000 Unicode code points.

List and count use one read-only repeatable-read PostgreSQL transaction. Recent order is published_at DESC/id ASC; title order uses explicit `COLLATE "C"` then id ASC. Search is a case-insensitive literal substring of title or summary. `!` is the SQL LIKE escape character; the adapter escapes `!`, `%` and `_`, so source backslashes remain ordinary text.

Generated database types remain in the adapter. HTTP DTOs are separate from domain entities. Database availability/timeout classes map to `dependency_unavailable`; unexpected SQL/domain mapping failures retain internal errors and are returned as safe HTTP 500. No raw database messages, SQL arguments or credentials enter HTTP responses.

The HTTP contract is kept in the backend, validated against observed handler responses with kin-openapi v0.149.0. The reserved Idempotency-Key component documents future payment work and is not attached to read-only operations. No OpenAPI serving endpoint was added; generation reads the checked-in specification directly.

## 4. Commands executed

Go ran in the pinned `golang:1.27.1-bookworm` container with the backend mounted at `/work`, plus named module/build caches.

```sh
gofmt -w internal cmd
go vet ./...
go test -race ./...
go test -tags=integration ./...
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
docker build --target api -t courses-api:course backend
```

The integration runner additionally mounts `/var/run/docker.sock` and uses `TESTCONTAINERS_HOST_OVERRIDE=host.docker.internal` on Docker Desktop. Tests fail when Docker cannot start the disposable database; there is no skip or fake repository-success fallback. They use Testcontainers v0.44.0 and PostgreSQL 17.5. `go mod tidy` pinned the actual dependency graph.

## 5. Passed verification

- TDD red/green cycles at the domain, application, repository and HTTP seams; early tests failed before their implementation existed.
- Domain/application table tests: publication states, canonical slug boundaries, query bounds/Unicode/trim, defaults at HTTP boundary, explicit not-found propagation, caller cancellation and an independently bounded slow dependency.
- Real PostgreSQL integration: empty page, migration up/down/up, atomic rollback of an intentionally failing migration, concurrent/repeated seed, hidden draft/archived rows, generated mapping, constraints, 1/50/51 records, offset 10000, stable ties in both sort modes, literal `%`/`_`/`!`/backslash search, case-insensitive search, cancellation while waiting on a database lock and successful reuse afterward.
- HTTP/OpenAPI: specification validation, valid request validation and every observed response validated against the schema; success/detail/empty/boundary cases; 400 malformed/repeated/unknown/empty/oversized/NUL inputs; identical hidden/unknown 404; safe 503/500; request-ID correlation; exact default and requested pagination.
- `go vet ./...`, `go test -race ./...` and the complete tagged integration suite passed, including the parent's architecture boundary tests.
- sqlc regeneration succeeded and generated-file checksums were unchanged.
- Docker API image built successfully; the Docker build stage compiles API, worker and migration executables.

All destructive migration checks ran only in disposable Testcontainers databases, which were terminated by test cleanup. No retained application database was reset or seeded by this subtask.

## 6. Remaining limits

Full Course frontend/SSR/browser and Compose end-to-end verification belong to C03–C05. This record does not mark them passed. Substring search is appropriate for the foundation catalog; no large-catalog search index or full-text engine is claimed. No public auth/admin/media/order endpoints were added. The production build was verified before the final small NUL-input guard; final unit/contract checks include that guard and the overall Compose build gate will rebuild it.

## 7. Next smallest task

Generate Orval from `backend/openapi/openapi.yaml`, map generated DTOs through the frontend API adapter, then rebuild Compose API/migrate and explicitly run `make seed` for the catalog browser tests.

## 8. User decision

None required. All decisions stay within the approved Course slice and keep future phases out of the implementation.
