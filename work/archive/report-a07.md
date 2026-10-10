# A07 report — public course program

## Changed

- `backend/sql/queries/courses.sql` — extended the published detail read to return course,
  module and lesson-title rows in deterministic module/lesson order; lesson descriptions
  and administrative fields are not selected.
- `backend/internal/course/domain/course.go`, `ports/repository.go` and
  `application/catalog.go` — added the title-only `CourseDetail` read model and carried it
  through the existing public detail seam without changing the catalog page model.
- `backend/internal/course/adapters/postgres/repository.go` — folded the ordered flat rows
  into an explicitly non-nil program and retained `course_not_found` for zero rows.
- `backend/internal/course/adapters/http/handler.go` — mapped only module and lesson titles
  into the detail response; the public list mapper remains unchanged.
- `backend/openapi/openapi.yaml` — changed only `GET /api/v1/courses/{slug}` success to
  `CourseDetail` and added the strict title-only program schemas. Generated output remains
  untouched for the reviewer.
- `backend/internal/course/application/catalog_test.go` and
  `backend/internal/course/adapters/http/admin_handler_test.go` — followed the narrowed
  repository return type in existing fakes.
- `backend/internal/course/adapters/postgres/repository_integration_test.go` — added
  out-of-insert-order program fixtures, asserted their public order, and proved programs
  attached to draft/archived fixtures remain unreachable.
- `backend/internal/course/adapters/http/handler_test.go` — added exact response-shape and
  ordering evidence for the published program, plus 404/no-leak evidence for draft and
  archived slugs; the existing OpenAPI response validator now exercises `CourseDetail`.
- `frontend/src/domain/course/course.ts`, `ports/CourseRepository.ts` and
  `application/catalog/catalog.ts` — separated detail-with-program from list items.
- `frontend/src/adapters/api/course-schema.ts`, `course-repository.ts` and
  `course-repository.test.ts` — validate strict title-only program nodes and map their order
  explicitly into the frontend domain.
- `frontend/src/presentation/storefront/CourseView.tsx`, `CatalogView.test.tsx` and
  `app/styles/main.css` — render a token-styled nested ordered list, with no links, and omit
  the entire program section when the program is empty.
- `frontend/src/app/providers/i18n.tsx` — added only the A07 program heading in Ukrainian
  and English; resource restructuring remains A08 work.
- `work/report-a07.md` — this report.

## Decisions

- One `LEFT JOIN` query serves present, empty and missing programs in one database snapshot.
  The module id is selected only as an adapter grouping key and never enters a public type
  or DTO. Stable order is `(module.position, module.id, lesson.position, lesson.id)`.
- `CourseDetail` embeds the existing course read model and adds
  `program: [{title, lessons: [{title}]}]`; `CoursePage` and `PublishedCourse` stay unchanged.
- Empty program and lesson collections are allocated as empty arrays, never serialized as
  `null`. `CourseView` renders no empty heading or broken list for an empty program.
- The rendered program uses native `section`, heading and nested `ol`/`li` elements. It has
  no playback route, lesson description, id, position, ownership field or new dependency.
- No admin surface, i18n bundle split, generated file, playback behavior or unrelated
  phase work was changed.

## Verified

- Local `gofmt -w` parsed and formatted the nine explicit Go files touched by the slice;
  final `gofmt -d` exited 0 with no output.
- Manual backend read-through followed the path SQL → PostgreSQL adapter → domain/port →
  application → HTTP DTO → OpenAPI. It confirmed the list schema remains `Course`, detail
  alone is `CourseDetail`, empty arrays stay arrays, and lesson descriptions never cross the
  query boundary.
- Manual frontend read-through followed OpenAPI-shaped input → Zod → adapter → domain →
  `CourseView`. It confirmed ordered mapping, strict program-node members, uk/en headings,
  native semantic lists, token-only colors/borders, no links, and no empty program section.
- Static acceptance scan found executable written evidence for the published ordered
  title-only response/UI and for inaccessible draft/archived programs. It also found no
  `program` addition to the list DTO/schema.
- Manual standards/spec review against `AGENTS.md`, `.claude/standards.md`, the backend and
  frontend architecture docs, brief A07 and plan D-6/P15/B17 found no remaining confirmed
  defect. The normal `code-review` skill workflow was denied because it requires Git
  fixed-point/diff commands and the brief says never touch Git; no workaround or diff-based
  subagent was used.
- Only the files listed above were written. No file inside a `generated/` directory was
  written. No filesystem write was denied.

## Not run

Environment denial: **No Docker: no generation, no compile, no Go/Rstest runs.** No
workaround was attempted.

- sqlc/OpenAPI/Orval generation and `make generate` — not run; the reviewer regenerates,
  and no generated file was hand-written.
- Go compilation, unit tests, HTTP contract tests and PostgreSQL integration tests — not
  run. The new cases are written evidence and are not represented as passing.
- Rstest component, domain and API-adapter tests — not run. The frontend changes received
  source read-through only, as required by the brief.
- OpenAPI document/request/response validation — not run because it is executed by the Go
  contract suite, whose compilation and execution are denied.
- `make format-check` — not run because its Go formatter is Docker-backed. Local `gofmt` is
  the only permitted formatting evidence.
- `make lint` — not run because its Go vet step uses Docker and compiles Go.
- `make typecheck` and `make test` — not run because compilation/tests are denied.
- `make integration` — not run because Docker/Testcontainers are unavailable.
- `make check-generated` — not run because generation is denied and the target also uses
  Git, which is separately forbidden. B17 is therefore not claimed as passed.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` and
  aggregate `make check` — not run because Docker is unavailable and these exceed A07's
  permitted evidence.
- No Git command was run; Git state was not inspected or changed, and no commit was created.

## Limitation and next task

The existing generated sqlc and Orval sources are intentionally stale. In particular, the
hand-written PostgreSQL adapter expects the generated `GetPublishedCourse` many-row result
after regeneration. A Docker-enabled reviewer must run generation first, then the ordered
format, lint, typecheck, test, integration, drift, build and browser gates, recording any
mismatch rather than treating this report's written tests as passed. A08 is the next logical
slice only after that review is green.

## Reviewer verification (measured in the reviewing session)

`make generate` exit 0; `make test` pass — frontend suite 0 failed tests, all Go
packages ok including the extended contract and catalog suites; `make typecheck`,
`make format-check`, `make lint` pass; `make integration` pass (12 packages).
`check-generated` green after commit. Provenance: implemented with Codex
(gpt-5.6-sol); reviewed and verified here.
