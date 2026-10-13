# lesson-builder — implementation report

## LB01 — domain (2026-10-12, Codex)

Implemented `course/domain/document.go` and `document_test.go`: versioned lesson
documents, section/block/mark types, `ValidateLessonDocument`, and the domain
sentinel `ErrInvalidLessonDocument`. Imports are standard library only. Validation
does not mutate or HTML-escape content; escaping is the LB11 reader's responsibility.

Acceptance evidence: `TestLessonDocumentRows` covers #5 and #12 (1/3 columns,
0/4 rejected, nested rows and video/task children rejected);
`TestLessonDocumentStructure` covers #11 and video cardinality;
`TestLessonDocumentSpans` and `TestLessonDocumentRichText` cover the domain portion
of #13 (URL allowlist, malformed URLs, marks, heading levels).
`TestLessonDocumentEsoterica` accepts video + text + row[image|text] + task,
including bold and a valid link. Field-consistency and boundary tables additionally
cover required payloads, forbidden payloads, Unicode lengths and aggregate limits.

### Domain bounds decision

2026-10-12, Codex, under the owner's instruction to choose sane named bounds:
200 sections **including row children**, 1–3 columns, depth one (children can only
be text/image); IDs up to 200 Unicode code points, image alt up to 2,000,
200 blocks per rich-text value, 200 items per list, 200 spans per paragraph/heading
or list item, 20,000 code points per span, 2,048 per href, at most 2 mark entries.
Total span text plus image alt is capped at 200,000 code points per document.
These are domain ceilings, not measured infrastructure capacity. Check with
`TestLessonDocumentBounds` (exact limit accepted, limit + 1 rejected), the marks
table and row table. UTF-8 and NUL handling reuse the existing authoring helpers.
Empty rich text, paragraph/heading spans and list items are accepted; a list must
still contain at least one item. Fields belonging to another variant are rejected,
including non-nil empty slices; no UUID restriction is imposed on section/image IDs.

### Verification and remaining work

Requested sequence, run locally from `backend`, **exit 0**:

```sh
gofmt -w internal/course/domain/document.go internal/course/domain/document_test.go
go vet ./internal/course/domain/
go test ./internal/course/domain/
```

Domain package tests passed (`ok courses/internal/course/domain 0.766s`); vet passed.
No local Go version error; Docker GO_RUN fallback was unnecessary.
Full make gate, integration, frontend and e2e were not run, as instructed. These
unit results do not close HTTP or browser acceptance rows or the overall phase.
No commits made. Existing unrelated changes preserved.

Next task: LB02 migration/backfill using this validator. The owner's explicit LB01
schema uses paragraph `spans: [{text: ...}]`; the earlier direct paragraph `text`
example in plan.md's migration section is superseded by that schema. Persistence,
transport mapping/error codes, rendering and publication remain later tasks.

## LB02 — migration and backfill (2026-10-12, Codex)

Implemented `backend/migrations/00018_lesson_document.sql`: `lesson_documents`
with the `(lesson_id, locale)` primary key, cascading lesson FK, uk/en locale,
JSON-object checks and 1,048,576-byte `pg_column_size` ceilings on both bodies;
`lesson_images` with the lesson FK, `created_by REFERENCES users(id)` and lesson
index. SQL backfills only non-empty descriptions into identical draft/published
uk documents, with deterministic `<lesson_id>-s1` text sections and the LB01
`content.blocks[].spans[].text` shape.

The owner's LB02 instruction (2026-10-12) explicitly supersedes the stale Down
paragraph in plan.md: 00018 never alters `course_lessons.description`; its Down
only drops the two new tables. Column removal and its reverse belong to
00019/LB06. Down loses the new document/image metadata and is for disposable test
databases only. Check the SQL Up/Down and the preservation subtest below.

Added one `lesson-document migration preserves populated database` integration
subtest in `repository_integration_test.go`: DownTo(17), seed a course/module and
lessons with non-empty/empty descriptions, Up, read both bodies, unmarshal into
`domain.LessonDocument`, validate, assert the exact section/paragraph/span shape,
preserved Ukrainian text and equal draft/published bytes. It also checks that
the old description remains unchanged and the empty description has no document.
The fixture course is removed with cleanup, cascading to its children.

### Verification and remaining work

Commands run from `backend`, in this order:

```sh
gofmt -w internal/course/adapters/postgres/repository_integration_test.go
go build ./...
go vet ./internal/course/...
GOCACHE=/private/tmp/courses-lb02-go-build go build ./...
GOCACHE=/private/tmp/courses-lb02-go-build go vet ./internal/course/...
```

gofmt passed. The initial build and vet exited 1 because the sandbox refused
access to `/Users/strelia/Library/Caches/go-build` (`operation not permitted`).
Both retries using the writable temporary cache passed (exit 0). `git diff --check`
also passed. No SQL file was passed to gofmt.

Integration tests, Docker/testcontainers, the full make gate and e2e were
deliberately not run, per the owner's instruction: the sandbox cannot bind the
required ports/containers (prior `listen EPERM`). The new subtest is behind the
`integration` build tag; the requested build/vet commands do not compile or run
it. Backfill execution and acceptance scenario #7 remain unverified here.

Human next step: run `make integration` from the repository root, or from
`backend` run:

```sh
go test -tags=integration ./internal/course/adapters/postgres -run '^TestRepositoryAgainstPostgreSQL$/^lesson-document_migration_preserves_populated_database$' -count=1
```

Next implementation task: LB03 (ports/application); render parity remains LB11,
and column removal remains LB06. No commits made; unrelated work preserved.

## LB03 — document ports and application (2026-10-12, Codex)

Implemented `domain/document_access.go`, the two ports in `ports/repository.go`,
`application/document.go` (`Documents` / `NewDocuments`) and table-driven tests in
`application/document_test.go`. Application imports only standard library and its
own domain/ports. No PostgreSQL adapter, HTTP handler or composition wiring added.

All four methods pass `actor.UserID` to the access reader. The domain maps the full
owner/status/paid/entitled/membership/lesson-preview facts into `access.CanConsume`
and `access.PurchaseRequired`. Editing reuses `Actor.CanManage` exactly, including
its valid actor ID and instructor/admin role checks; denied edits return
`ErrForbidden`. Published reads check email before calling either port, then check
access before loading content, including when content is absent. Missing content
returns a current-schema document with the requested locale and an empty sections
array; missing drafts also return a zero updated-at token.

The repository contract names `domain.ErrDocumentConflict` for optimistic-lock
misses; application normalizes wrapped instances to that sentinel. Existing
`ErrForbidden` / `ErrPurchaseRequired` are reused; `ErrEmailNotVerified` and
`ErrDocumentConflict` are added in the domain using the existing `Error` style.
Save validates with LB01's `ValidateLessonDocument` before persistence.

Unit evidence:
- `TestDocumentsPublishSeparatesDraft`: save leaves the published document absent
  or unchanged until publish; editor gets draft and updated-at; all four methods
  forward viewer/lesson IDs and the repository receives the locale.
- `TestDocumentsStaleWrites`: stale save and publish return `ErrDocumentConflict`
  without replacing either body (#18, application proof only).
- `TestDocumentsEditorGate`, `TestDocumentsReadGate` and
  `TestDocumentsAccessIsPerViewer`: owner/admin edit permission; non-owner refusal;
  email checked before access; paid/free nonmembers; published/draft preview;
  active/blocked/revoked membership; archived member; entitlement alone does not
  grant consumption; owner/admin reading; gates apply with and without content.
- `TestDocumentsInvalidDraft`, `TestDocumentsMissing`, `TestDocumentsPortErrors`:
  invalid section never persists; absent uk/en documents are valid empty bodies;
  failures propagate and access failures never reach the content repository.

Verification, in order (all exit 0):

```sh
cd backend
gofmt -w internal/course/application/document.go internal/course/application/document_test.go internal/course/ports/repository.go internal/course/domain/document_access.go
GOCACHE=/private/tmp/courses-lb03-go-build go vet ./internal/course/...
GOCACHE=/private/tmp/courses-lb03-go-build go test ./internal/course/application/ ./internal/course/domain/
```

Application tests passed in 3.849s; domain tests passed in 1.194s. A writable
GOCACHE was used for sandbox compatibility. `git diff --check` also passed.
No full gate, Docker, integration or e2e run; no commits made.

Next: LB04, implementing and proving atomic publication, optimistic locking and
per-viewer access joins against PostgreSQL. These unit tests use fakes and do not
prove SQL behavior. Locale coexistence/fallback in plan scenario #8 remains
repository-level evidence; HTTP and browser acceptance remain LB05/LB11/LB12.

## LB04 — PostgreSQL document adapter (2026-10-12, Codex)

Implemented `course/adapters/postgres/document.go` on the existing `Repository`
and pool, with compile-time assertions for both exact LB03 ports. Raw pgx and
encoding/json persist the domain document; no generated models or schema changes.
First save inserts an empty valid published document (`sections: []`) beside the
draft. Subsequent saves change only the draft under the updated-at guard. Publish
copies draft to published and advances updated-at in one guarded UPDATE. Both
write paths map no returned row to `ErrDocumentConflict`. Missing reads return
found=false without an error; malformed stored JSON returns dependency unavailable.

The access reader joins lesson → module → course and left-joins the viewer's grant.
It mirrors learning.go's nullable price/currency consistency check, active-only
Entitled flag and membership() mapping, and carries the lesson's preview flag.
Added `domain.ErrLessonNotFound` for absent lesson access facts. No contract
ambiguity was found; no assumptions entry was needed.

Implementation decision (2026-10-12, Codex): use the requested single-statement
upsert. PostgreSQL's [INSERT documentation](https://www.postgresql.org/docs/current/sql-insert.html)
explicitly states that a conflicting row failing the DO UPDATE WHERE condition
is not returned. Thus pgx.ErrNoRows identifies the stale guard without a separate
existence query. Check through the integration test's zero/stale token cases;
documentation confirmation is not a claim of local database execution.

Added `document_integration_test.go` with the existing testcontainers/goose setup
and an isolated database removed on container cleanup. Tests cover first-save
empty publication, JSON/token round-trip, matching and stale saves, guarded
publication and unchanged content/tokens after conflicts, private edits after a
publish, missing reads/publish, and separate uk/en rows. Access assertions cover
owner edit permission, paid nonmember refusal, active/blocked/revoked membership,
published preview without membership, free draft preview refusal and missing
lesson sentinel. These are pending PostgreSQL evidence for #8/#18, not closed
acceptance rows.

Verification (all exit 0), from `backend`:

```sh
gofmt -w internal/course/adapters/postgres/document.go internal/course/adapters/postgres/document_integration_test.go internal/course/domain/document_access.go
GOCACHE=/private/tmp/courses-lb03-go-build go build ./...
gofmt -w internal/course/adapters/postgres/document.go internal/course/adapters/postgres/document_integration_test.go
GOCACHE=/private/tmp/courses-lb03-go-build go vet ./internal/course/...
GOCACHE=/private/tmp/courses-lb03-go-build go test -c -tags=integration -o /private/tmp/courses-lb04-postgres.test ./internal/course/adapters/postgres
```

Used the existing writable temporary Go cache for sandbox compatibility.
`go test -c` compiled the integration suite only; it did not execute tests or
start Docker. `git diff --check` passed. Docker/integration execution, the full
ordered make gate and e2e were deliberately not run per the owner's instruction
(prior container bind listen EPERM). No commits made; unrelated work preserved.

Human next step: `make integration`, or from `backend`:

```sh
go test -tags=integration ./internal/course/adapters/postgres -run '^TestLessonDocumentRepositoryAgainstPostgreSQL$' -count=1
```

LB04 implementation is ready for that runtime verification. Next implementation
slice is LB05 (HTTP/OpenAPI/contract tests); the complete phase gate remains LB15.

## LB05 — document HTTP contract (2026-10-12, Codex)

Implemented `DocumentHandler` and API composition wiring over the existing
`Documents` service and PostgreSQL repository. Four operations were added:

- `GET /api/v1/admin/lessons/{lessonId}/document`: draft plus nullable updatedAt;
  a missing draft supplies an empty valid document and null first-write token.
- `PUT /api/v1/admin/lessons/{lessonId}/document`: save the domain document directly,
  using its locale and an optional/null first-write token; returns updatedAt.
- `POST /api/v1/admin/lessons/{lessonId}/document/publish`: required microsecond
  token, optional uk/en locale, empty 200 response.
- `GET /api/v1/learn/lessons/{lessonId}/document`: published domain document directly;
  absent publication remains an empty valid document with 200.

All handlers set no-store, reuse the admin/learning actor helpers and delegate
editor, email-first, consumption and validation policies to the service. Mutations
use RequireSameOrigin outside Authenticate, matching admin course wiring. The
admin JSON decoder now delegates to a limit-parameterized helper: existing admin
writes retain 512 KiB, document saves allow 2 MiB, publication allows 1 KiB.
Duplicate object keys, unknown fields (including nested document fields), trailing
bytes and invalid content types remain rejected. Query locales default to uk;
unknown, duplicated or invalid query keys/values are refused. PUT has no query.

Publication advances updated_at but the port returns no timestamp. **The editor
must re-GET the draft after publish to obtain the fresh lock token.** The port is
unchanged, as explicitly requested by the owner on 2026-10-12; check the empty
publish response and token forwarding in `TestDocumentDraftRoundTrip`.

OpenAPI adds `getAdminLessonDocument`, `saveAdminLessonDocument`,
`publishAdminLessonDocument`, `getLearnLessonDocument`, and the LessonDocument,
LessonSection (recursive columns), LessonRichText, LessonBlock, LessonSpan,
LessonDocumentDraft, LessonDocumentUpdated, SaveLessonDocumentRequest and
PublishLessonDocumentRequest schemas. It adds a document-conflict response and
the invalid_lesson_document/document_conflict error codes. Adjacent admin/learn
operations have no tags, so these follow that convention. Optional fields mirror
Go omitempty. Required slice fields permit null because nil sections/blocks/list
items are accepted by the existing domain and encoded as null; absent publications
still specifically return sections: []. Recursive nullable schemas generate
usable Orval types and pass request/response schema checks.

### Contract evidence

The owner's LB05 instruction supersedes the plan's older error-code wording and
row-label assignments; the exact observed behaviors below identify the evidence
without claiming every similarly numbered phase row closed:

- `TestDocumentInvalidDraftNeverPersists`: unknown type, video carrying content
  and nested row return 400 invalid_lesson_document without persistence (plan #11/#12).
- `TestDocumentNonEditor`: non-owner instructor draft GET/save/publish return 403
  forbidden before the content repository.
- `TestDocumentPublishedAccess`: paid nonmember returns course_purchase_required;
  active member gets only published content, verified preview is allowed; unverified
  email is denied before access lookup; absent publication is a valid en empty
  document with 200. Covers the requested access/absence/email behaviors; image
  serving and browser draft visibility remain LB09–12, not proven here.
- `TestDocumentConflicts`: stale save and publish map wrapped repository conflicts
  to 409 document_conflict, with exact locale/token forwarding (#18).
- `TestDocumentDraftRoundTrip`: null initial token, absent/null PUT token, draft
  document/token reads, PUT token response and GET→PUT→publish token forwarding.
- `TestLearningOutlineHasNoDocumentBody`: structural guard on learningLessonDTO
  forbids document/sections/body (AC7 / plan #14), using the requested lighter option.
- Transport tables cover query, JSON and timestamp rejection, nested unknown and
  duplicate fields, body limits, all error mappings, authentication and mutation
  origin wiring. A body above the old admin limit succeeds under the new limit.
- `TestDocumentHTTPMatchesOpenAPI` validates all four operations through the real
  HTTP wrapper (including request IDs); `TestDocumentRecursiveSchemaRoundTrip`
  checks rich text, recursive columns, images/tasks/video and nullable collections.

These tests use in-package port fakes and the real application service, HTTP
handlers and platform wrapper. They prove transport/service composition, not
PostgreSQL concurrency or deployed nginx behavior.

### Verification and remaining work

From `backend`, using the existing writable cache, ran this ordered sequence twice:

```sh
gofmt -w internal/course/adapters/http/document_handler.go internal/course/adapters/http/document_handler_test.go internal/course/adapters/http/admin_handler.go cmd/api/main.go
GOCACHE=/private/tmp/courses-lb03-go-build go build ./...
GOCACHE=/private/tmp/courses-lb03-go-build go vet ./internal/course/...
GOCACHE=/private/tmp/courses-lb03-go-build go test ./internal/course/adapters/http/
```

First run: formatting/build/vet passed; HTTP tests failed because the new contract
fixture manually supplied a request ID outside the OpenAPI hex pattern. Replaced
that fixture with `httpserver.NewHandler`, which generates real request IDs, and
added recursive-schema/authentication/origin checks. Second complete sequence:
**all exit 0**, HTTP package `ok` (6.168s).

From `frontend`:

```sh
npm run generate && npm run typecheck
```

**Exit 0**: Orval regenerated the client/models and TypeScript passed.

Offline sqlc attempt, from `backend`:

```sh
GOCACHE=/private/tmp/courses-lb03-go-build GOPROXY=off GOSUMDB=off go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate
```

**Exit 1 / regeneration skipped:** module lookup disabled by GOPROXY=off while
loading sqlc module deprecation metadata. No SQL queries changed and no sqlc output
changed. Root `git diff --check` passed.

**Human-side verification remains `make check-generated` and the full ordered
`make check` gate, including the full e2e suite.** Docker, integration, listeners
and e2e were deliberately not run under the owner's sandbox-safe instruction.
LB05 implementation and local contract verification are complete; the phase is
not closed. Next implementation slice: LB06 Description removal. No commits made;
existing unrelated screenshots and archive files preserved.

### LB05 — session verification (2026-10-12, Courses 1)

Closed the sqlc-offline gap Codex flagged. From the repo root, after staging the
LB05 paths (explicit paths only — no stray screenshots), ran the full generator:

```sh
make check-generated   # sqlc via Docker (golang:1.27.1) + orval, then git diff --exit-code
```

**CG_EXIT=0** — sqlc regenerated with **no drift** and the Orval client/models match
the staged output. Plus `go build ./...`, `go vet ./internal/course/...` (clean) and
`go test ./internal/course/adapters/http/` (**ok 6.223s**) re-run in-session.

LB05 is **done** at the slice level. The ordered `make check` + full e2e under enforcing
CSP is the phase gate (LB15), not a per-slice run; LB05 adds no migration or SQL query,
so `make integration` is unchanged by it.


## LB06 — remove the flat lesson description (2026-10-12, Codex)

Implementation complete on `feat/lesson-builder`; verification remains pending outside
this sandbox. No commits made. Owner's pre-launch “cut freely” decision (2026-10-08,
reconfirmed in this task) is implemented without a compatibility shim. Only the lesson
field is removed. `courses.description`, course Summary, CourseDraft/CourseTree
Description, course API descriptions and course form fields/fixtures are preserved.

Migration 00019 drops `course_lessons.description` after 00018's document backfill.
Both directions use `SET LOCAL lock_timeout = '5s'`. **Named downgrade loss:** Down
adds a NOT NULL text column with default `''` and the original 0–1000-character CHECK;
it does not reconstruct flat text from lesson_documents. The Down block says this
explicitly. Check via the extended lesson-document migration-preservation subtest:
DownTo(17) → populated seed → Up (now 19) → validated/preserved document + absent
column → DownTo(18) → empty restored column + unchanged published document → Up.
That integration test was edited but NOT run here.

### Completed checklist

- Removed the Lesson field and ceiling/validation, SQL lesson projection, lesson
  upsert column/value/update and aligned arguments, tree mapper, admin lesson DTOs,
  request nil-check/mapping, and AdminLesson OpenAPI property/required entry.
- Kept `validOptionalText`: document.go still uses it for rich-text spans and links.
  Removed only the obsolete lesson-bound tests and their fixture/helper fields;
  retained course bounds and HTTP request-body limit coverage.
- Regenerated Orval. Removed the lesson field from the API mapper, authoring domain
  type/validation, Zod schema, form type/control/default, and en/uk label.
  The admin lesson schema is actually in `admin-course-schema.ts`; inspected
  `course-schema.ts` and left its course description unchanged. Inspected
  `admin-queries.test.ts` and `course-schema.test.ts`: only course descriptions,
  so neither needed edits.
- Updated lesson fixtures in unit tests, shared backend integration tests, browser
  journeys and the playback authorization shell fixture. These other modules insert
  into the current schema and would otherwise reference the dropped column.
  The three lesson INSERT fixtures below migration 19 in course's
  `repository_integration_test.go` remain unchanged. Historical migrations 00003
  and 00018 also remain unchanged.
- Reviewed remaining lesson-description references: historical migration/backfill
  and downgrade assertions, plus intentionally stale generated sqlc output. No
  hand-written production mapper reads a lesson Description field.

Files completed (all paths relative to repository root):

- `backend/internal/adminstats/adapters/postgres/engagement_integration_test.go`
- `backend/internal/course/adapters/http/admin_handler.go`
- `backend/internal/course/adapters/http/admin_handler_test.go`
- `backend/internal/course/adapters/http/contract_ceilings_test.go`
- `backend/internal/course/adapters/postgres/authoring.go`
- `backend/internal/course/adapters/postgres/authoring_integration_test.go`
- `backend/internal/course/adapters/postgres/catalog_measure_test.go`
- `backend/internal/course/adapters/postgres/document_integration_test.go`
- `backend/internal/course/adapters/postgres/learning_integration_test.go`
- `backend/internal/course/adapters/postgres/repository_integration_test.go`
- `backend/internal/course/application/authoring_test.go`
- `backend/internal/course/domain/authoring.go`
- `backend/internal/course/domain/authoring_test.go`
- `backend/internal/discussion/adapters/postgres/repository_integration_test.go`
- `backend/internal/media/adapters/postgres/repository_integration_test.go`
- `backend/internal/media/adapters/tus/server_test.go`
- `backend/internal/orders/adapters/postgres/repository_integration_test.go`
- `backend/internal/playback/adapters/postgres/access_integration_test.go`
- `backend/internal/playback/adapters/postgres/progress_integration_test.go`
- `backend/internal/playback/application/sessions_integration_test.go`
- `backend/internal/transcript/adapters/postgres/draft_integration_test.go`
- `backend/internal/transcript/adapters/postgres/repository_integration_test.go`
- `backend/migrations/00019_drop_lesson_description.sql`
- `backend/openapi/openapi.yaml`
- `backend/sql/queries/courses.sql`
- `frontend/e2e/accounts.spec.ts`
- `frontend/e2e/admin.spec.ts`
- `frontend/e2e/asr.spec.ts`
- `frontend/e2e/discussion.spec.ts`
- `frontend/e2e/enrollment.spec.ts`
- `frontend/e2e/instructor.spec.ts`
- `frontend/e2e/learn-access.spec.ts`
- `frontend/e2e/learning.spec.ts`
- `frontend/e2e/media.spec.ts`
- `frontend/e2e/purchase-wayforpay.spec.ts`
- `frontend/e2e/purchase.spec.ts`
- `frontend/e2e/ratings.spec.ts`
- `frontend/e2e/refund.spec.ts`
- `frontend/e2e/transcript.spec.ts`
- `frontend/src/adapters/api/admin-course-repository.test.ts`
- `frontend/src/adapters/api/admin-course-repository.ts`
- `frontend/src/adapters/api/admin-course-schema.test.ts`
- `frontend/src/adapters/api/admin-course-schema.ts`
- `frontend/src/adapters/api/generated/courses.ts`
- `frontend/src/adapters/api/generated/models/adminLesson.ts`
- `frontend/src/app/providers/locales/en.ts`
- `frontend/src/app/providers/locales/uk.ts`
- `frontend/src/domain/course/course.test.ts`
- `frontend/src/domain/course/course.ts`
- `frontend/src/presentation/admin/CourseEditorView.test.tsx`
- `frontend/src/presentation/admin/CourseEditorView.tsx`
- `frontend/src/presentation/admin/LessonVideoField.test.tsx`
- `infrastructure/tests/playback-authorize.sh`

### Commands and results

From repository root (run after Go edits; all edited Go files formatted):

```sh
git diff --name-only -- '*.go' | xargs gofmt -w
cd backend
GOCACHE=/private/tmp/courses-lb03-go-build go build ./...
GOCACHE=/private/tmp/courses-lb03-go-build go vet ./internal/course/...
GOCACHE=/private/tmp/courses-lb03-go-build go test ./internal/course/domain/ ./internal/course/application/ ./internal/course/adapters/http/
```

**PASS:** formatting, build and vet exit 0; domain, application and HTTP tests pass.
The final run used cached domain/application results and HTTP passed in 5.875s.
The writable cache location avoids the restricted default cache. This build uses
stale sqlc output; it is not evidence that generated SQL works after migration 19.

From `frontend`:

```sh
npm run generate && npm run typecheck && npm test > /private/tmp/courses-lb06-frontend-test.log 2>&1
npm run typecheck
```

First generation passed; first typecheck failed (exit 2) on a residual lesson
`description` in CourseFormValues. Removed it and reran the chain: generation and
typecheck passed; **npm test FAILED before executing tests**, because Rstest attempted
`listen ::1:3000` and received `EPERM`. No listener workaround or configuration change
was attempted. The final standalone typecheck passed (exit 0).

Formatted the 25 modified hand-written TS/TSX paths with the existing formatter
(`./node_modules/.bin/oxfmt <changed TS/TSX paths>`, cwd frontend; generated output
excluded): **PASS**. Root `git diff --check`: **PASS**.

### Required human handoff / limitations

**Backend sqlc output could not be regenerated in-sandbox and is intentionally stale.**
In particular the old generated ListAuthoringLessons SELECT still names the dropped
column; running the migrated application before regenerating it will fail that query.
Hand-written Go compiles without reading the stale generated Description field.

The human must run from the repository root:

```sh
make generate          # includes sqlc via Docker; Orval already regenerated here
make check-generated
make integration       # full migration-preservation + authoring + shared fixture coverage
(cd backend && go build ./...)  # fresh build against regenerated sqlc
(cd frontend && npm test)      # rerun where Rstest can bind its local port
```

Docker, integration, sqlc generation, check-generated and e2e were not run here,
per the task's sandbox constraints; none is claimed passed. Full ordered
`make check` and the full e2e suite remain the LB15 phase gate. LB06 is not marked
fully verified until the human handoff passes. Next implementation slice is LB07.
Existing unrelated screenshot modifications and untracked archive files were preserved.

### LB06 — session verification (2026-10-12, Courses 1)

Ran the human-side gates Codex could not run in its sandbox, from the repo root:

```sh
make generate        # sqlc via Docker (drops lessons.description) + Orval → GEN_EXIT=0
make check-generated # regenerate + diff → CG_EXIT=0 (no drift)
(cd backend && go build ./... && go vet ./internal/course/...)  # EXIT 0 vs regenerated sqlc
make integration     # INTEG_EXIT=0; courses/internal/course/adapters/postgres ok 23.258s
(cd frontend && npm test)   # all passed; typecheck green
```

The integration run exercised migration **00019** (DROP COLUMN course_lessons.description),
every `*migration preserves populated database` subtest (incl. the LB02 lesson-document
backfill at DownTo 17 → Up through 18/19), authoring write/load against the dropped
column, and the cross-module lesson fixtures Codex updated (adminstats/discussion/media/
orders/playback/transcript). Guardrail held: `courses.description` and course summaries
untouched; no hand-written reference to the lesson description remains.

Blast radius note: LB06 touched **every** e2e spec (lesson-creation payloads dropped the
`description` field, now rejected by DisallowUnknownFields). e2e was NOT run here — it is
the LB15 phase gate (full ordered `make check` + e2e under enforcing CSP), which must run
the whole suite because this slice changes shared/persisted schema.

### LB11 — learner lesson-document reader (2026-10-13, Courses 1, direct)

Implemented directly in the main session (not via Codex): a contained, read-only
frontend slice with full context already loaded. Committed as two slices.

**Owner decision (AskUserQuestion, 2026-10-13).** A rendered image needs the serve
extension (`/media/i/<id>.<ext>`; nginx matches that location only with an extension),
but the stored section carried only `imageId`, and the content type lives in the image
module's `lesson_images` table, which the course module may not read. Chosen: the
document carries `contentType` on image sections. Rejected: a full URL baked into the
persisted jsonb (prefix change would need a document migration) and a backend resolve at
read time (course reading the image module's table breaks module boundaries). No data
migration: no image section has been stored yet. Artefact: `assumptions.md` Журнал LB11.

**LB11a `5b942fc` (backend).** `domain.Section.ContentType` with validation: an image
section must carry exactly one content type from the fixed `image/jpeg|png|webp` allowlist,
every other section type must not carry one. OpenAPI `LessonSection.contentType` (enum),
Orval regenerated. Verified: `go test ./internal/course/domain/ ./internal/course/adapters/http/` ok,
`make check-generated` CG_EXIT=0 (no sqlc/Orval drift), frontend `tsc --noEmit` 0.

**LB11b `5136c5e` (frontend).** `LessonDocumentView` renders the published body as
semantic HTML with no editor dependency: text/task rich text (paragraph, h2/h3 heading,
ordered/unordered list, bold/italic, http(s) link), image `<figure><img>`, and a depth-1
`row` of text/image columns with a responsive CSS preset (stacks at narrow width and 200%
zoom). A `video` section is played by the existing lesson player, not the reader, so a
document that is only a video slot — or an absent publication — renders nothing (scenario
4). The reader builds `/media/i/<imageId>.<ext>` from the section content type (mirror of
`internal/image/domain.ExtensionFor`). The zod schema is the outermost mirror of the
contract and refuses what the server should never produce before it reaches the DOM —
above all a non-http(s) href (`javascript:` refused) and a non-UUID image id (a URL-path
injection). The port/adapter/query mirror the transcript: request-scoped repository,
non-suspending in the lesson route (a reader failure never stops playback), slotted into
`OpenLesson` right after the player.

Verified: `tsc --noEmit` 0; `oxfmt` + `oxlint --deny-warnings` + architecture boundaries 0;
frontend suite **702** tests pass (14 new: component render of every section type + URL
extension + empty/video-only → null + migration parity; domain helpers; schema accept/reject).

**Limitation — e2e render carried to LB15.** Scenarios 2/3/7 (a learner reading a published
document, including the migrated flat description) are e2e-render rows. They are deferred to
the LB15 phase gate: no editor (LB12) exists yet to author a document, and the e2e seed has
no `lesson_documents` row, so the phase e2e (lesson-builder.spec) will seed or author a
published document and assert rendering there, alongside the full ordered `make check`. This
matches the plan's "full e2e at LB15". The reader's client-side logic is covered by the unit
suite; what LB15 adds is the real `/media/i/<id>.<ext>` fetch and the gated document endpoint
over the production path.
