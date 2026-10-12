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
