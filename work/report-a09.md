# A09 report — frontend admin adapter and application

## Files

- `frontend/src/domain/course/course.ts` — added branded course/module/lesson ids, admin
  list/tree/write types, the `AuthoringError` result, client-minted node ids, bounded tree,
  transition, slug-mutability, pagination and microsecond timestamp validators.
- `frontend/src/domain/course/course.test.ts` — added authoring rule evidence for tree bounds,
  duplicate ids, lifecycle timestamps, transitions, permanent slugs and list pagination.
- `frontend/src/ports/AdminCourseRepository.ts` — added the seven-operation authoring port.
- `frontend/src/application/admin/authoring.ts` and `authoring.test.ts` — added use cases that
  validate user-authored values before calling the port and carry the loaded concurrency token
  into full-resource writes.
- `frontend/src/adapters/api/admin-course-schema.ts` and
  `admin-course-schema.test.ts` — added strict Zod parsing for list, tree and instructor
  responses; ids are branded only after UUID validation, and malformed payloads are refused.
- `frontend/src/adapters/api/admin-course-repository.ts` and
  `admin-course-repository.test.ts` — added explicit domain mapping over every generated admin
  operation through `createApiTransport`; MSW evidence covers every D-5 business code,
  including R27's `course_modified`, `slug_taken`, `forbidden` and 500 mapping.
- `frontend/src/app/composition/admin-queries.ts` and `admin-queries.test.ts` — added query and
  mutation options, stable keys, the two named staleness constants, named write/delete
  invalidation helpers and conflict-triggered tree refetch evidence.
- `frontend/src/app/composition/admin.ts` — added request-scoped admin dependency wiring,
  mirroring catalog/identity composition.
- `work/report-a09.md` — this report.

## Decisions

- `CourseId`, `ModuleId` and `LessonId` share one private brand mechanism and one documented
  `asBrandedId` trust-boundary escape hatch. Zod validates before branding; new module and
  lesson ids use the native `crypto.randomUUID()`. Instructor options stay a plain local shape,
  with no `EntityRef`.
- The existing catalog `Result` gained an error type parameter with `CatalogError` as its
  default. Authoring therefore reuses the same result shape without changing catalog callers.
- Save is one full-tree PUT carrying the exact loaded `updatedAt` as `expectedUpdatedAt`.
  Lifecycle and delete use the same token; no delta/CQRS helpers were introduced.
- The adapter accepts a named code only at its documented D-5 status. `internal_error` and
  unnamed 5xx responses become `dependency_unavailable`; malformed success data becomes
  `invalid_response`. Generated models never cross the adapter boundary.
- `adminListStaleMs` is 30 seconds and `instructorListStaleMs` is 5 minutes. A write invalidates
  list and tree; delete removes the tree and invalidates the list. `course_modified` calls
  `invalidateAfterCourseWrite`, which refetches an active editor query.
- No router, route, screen, search parser, RHF, Zustand or dependency was added.

## Verified / how

- `cd frontend && npm run lint` passed: oxlint reported no warnings and the architecture check
  reported `Architecture boundaries checked (56 files).`
- `cd frontend && npm run typecheck` passed with `tsc --noEmit`.
- Manual read-through followed domain → port → application → generated client adapter →
  request-scoped composition. It checked full-PUT serialization, exact status/code mapping,
  response-id/status checks, strict response parsing and named invalidation coverage.
- Independent standards/spec review over the named A09 files found no spec or hard standards
  defect. Its one low repeated-switch test smell was removed before the final static checks.
- Formatting was matched by eye against neighboring TypeScript files: two spaces, single
  quotes, trailing commas and no line over 100 characters.
- No file inside a `generated/` directory was written. No filesystem write was denied.
- No Git command was run; Git state was not inspected or changed, and no commit was created.

## Not run

Environment denial: **Docker is unavailable; local Rstest execution is denied because its port
bind is refused by the sandbox; oxfmt execution is denied.** No workaround was attempted.

- `npm test` / targeted Rstest — not run. The R27, schema, domain, application and query tests
  are written for the reviewer and are not represented as passing.
- `npm run format:check` and `make format-check` — not run because oxfmt is denied; the make
  target also invokes Docker-backed Go formatting.
- `make lint` — not run because its Go vet half requires Docker. Only the frontend-local lint
  command reported above ran.
- `make typecheck` — not run. Its frontend command was run directly and passed as reported.
- `make test` — not run because it includes both denied Rstest and Docker-backed Go tests.
- `make integration` — not run because it requires Docker/Testcontainers.
- `make generate` and `make check-generated` — not run: generation would write inside
  `generated/`, and the drift target also invokes Git. Both are expressly forbidden here.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke` and `make e2e` —
  not run because they require Docker and/or denied test execution.
- Aggregate `make check` — not run; its denied/unavailable constituent gates are listed above.

## Limitation and next task

A Docker-enabled reviewer must run the written Rstest evidence and the ordered repository gates.
A10 may begin only after that review is green; it owns the admin list route and screens.

## Reviewer verification (measured in the reviewing session)

`make format` applied (oxfmt was unrunnable in the sandbox; formatting was the only
gap), then `make format-check`, `make lint` (oxlint --deny-warnings + architecture
check), `make typecheck` and `make test` all pass — 0 failed frontend tests including
the new R27 mapping, schema and query-option suites. Provenance: implemented with
Codex (gpt-5.6-sol); reviewed and verified here.
