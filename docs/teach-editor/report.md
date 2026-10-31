# Teach editor verification

## TE1 — 2026-10-30 (Codex)

Scope: document teach alias only, based on `22aee5072e4adb6f9fe83a56c941805fad245a2a`
on `feat/teach-editor`. Existing unrelated WIP preserved. No commit or generation.
The existing application authorization and handler bodies remain unchanged.

Added `RegisterTeach`, composition-root registration, and three OpenAPI source bindings.
Added `TestTeachDocumentEditorAccess` (owner GET/PUT/publish 200; non-owner instructor
and student 403 `forbidden` on each), `TestTeachDocumentUnknownSibling` (history 404),
and `TestTeachDocumentAuthenticationAndOriginWiring` (anonymous 401 and existing
admin-style middleware stub: GET bypasses origin, PUT/POST refused before persistence).
Existing admin/learn tests retained byte-for-byte.

TDD: initial missing-method compile failure, then empty registration produced behavioral
RED (404 instead of 200/403/401); implemented registration produced GREEN:
`go test ./internal/course/adapters/http -run TestTeachDocument -count=1` exit 0.

Requested self-verification, from backend, in order:

| Command | Result |
| --- | --- |
| `gofmt -l internal/course cmd` | PASS, exit 0, empty output |
| `go vet ./internal/course/... ./cmd/...` | BLOCKED, exit 1: sandbox operation not permitted reading Go build cache |
| `go build ./cmd/api` | BLOCKED, exit 1: same cache access error |
| `go test ./internal/course/...` | BLOCKED, exit 1: setup failed on cache access; application/domain cached passes do not establish full suite success |

Blocked cache path: `/Users/strelia/Library/Caches/go-build/`.
Stopped per task instruction; no permission escalation or cache workaround attempted.
T8 unchanged-test integrity verified, but full rerun is BLOCKED. T10 proves mount wiring
only; real identity same-origin middleware/full-chain e2e remains for TE6.
OpenAPI generation, full ordered gate, mutation review and independent review NOT_RUN;
coordinator owns generation and subsequent integration/review. Next: resolve Go cache
access, rerun the requested checks, then coordinate TE2/TE3 and generation.

### Coordinator verification (Courses13, 2026-10-30) — impact gate GREEN

Ran the checks Codex's sandbox blocked, plus generation, from a full-access shell:
- `gofmt -l internal/course cmd` → empty (clean).
- `go vet ./internal/course/... ./cmd/...` → exit 0.
- `go build ./cmd/api` → ok (no output).
- `go test ./internal/course/...` → all `ok` (http ran fresh 9.66s with the new teach tests; application/domain/postgres cached pass).
- `make generate` → regenerated `backend/openapi/openapi.yaml` (both teach paths present) + Orval client (`frontend/src/adapters/api/generated/courses.ts`, `models/index.ts`, 4 new teach model files); staged; `git diff --exit-code` on generated paths clean + zero untracked → `check-generated` satisfied.
- frontend `typecheck` → clean; `format:check` → 373 files correct; `lint` (oxlint + architecture, 229 files) → clean.

**Correction applied before implementation (supersedes GATE 2 B1):** non-owner instructor on the
document endpoints returns `forbidden`, not `course_forbidden` — `requireEditor`
(`application/document.go:92`) returns `ErrForbidden` for every non-editor because `CanBeEditedBy`
collapses `CanManage`'s specific error to a boolean. Codex surfaced this by refusing to assert a code
that contradicts the existing passing `TestDocumentNonEditor`. spec AC2 / plan T5,T7 / tasks TE1-TE2
corrected; all teach endpoints → `forbidden` for a non-owner. The owner-enforcement property is
unchanged; this is an error-code-string fact.

Deferred to TE6 per plan: `make integration` and the full ordered `make check` incl. e2e (TE1 adds an
isolated mount with no DB/shared-state behavior change; admin/learn tests unchanged and green).
Independent review before merge still pending (mandatory-review-every-task).

## TE2 — 2026-10-30 (Codex)

Scope: image teach alias on `feat/teach-editor`, base `25d0502d20922d0b9351d72649403a28a86f2e52`.
Added `Handler.RegisterTeach`, API wiring, a matching OpenAPI source binding,
and expanded public-route census/fidelity expectations to include teach.
Handler bodies and application authorization unchanged; unrelated WIP preserved.

`TestTeachUpload` adds T4 owner 201 with exact upload response and T5 non-owner
lecturer 403 `forbidden`, using existing authenticated/passThrough/fakeUploader
harness and checking forwarded actor, lesson, dimensions, content type and bytes.
These are fake-backed HTTP contract tests, not independent proof of application ownership.
TDD: missing-method compile failure, then empty registration produced behavioral
RED (both 404, uploader never called). Real mount produced GREEN:
`go test ./internal/image/adapters/http -run 'TestTeachUpload|TestOpenAPIContributionCoversPublicRegistrations' -count=1`
exit 0.

Requested self-verification (backend working directory):

| Command | Result |
| --- | --- |
| `gofmt -l internal/image cmd` | PASS, exit 0, empty output |
| `go vet ./internal/image/... ./cmd/...` | BLOCKED, exit 1: sandbox operation not permitted reading `/Users/strelia/Library/Caches/go-build` |
| `go build ./cmd/api` | NOT_RUN: stopped on cache blocker as instructed |
| `go test ./internal/image/...` | NOT_RUN: stopped on cache blocker as instructed |

No cache workaround, escalation, commit or code generation. Coordinator owns
remaining checks and `make generate`; expanded fidelity expectations require the
new teach image path in generated YAML. Full gate, mutation review and independent
review NOT_RUN. Next: coordinator regenerates and verifies TE2 before integration.

## TE3 — 2026-10-30 (Codex), blocked at first TDD run

Base `25d0502d20922d0b9351d72649403a28a86f2e52`, branch `feat/teach-editor`.
User-scoped seam: HTTP GET `/api/v1/teach/courses/{id}` through the existing
admin identity/authenticate harness and real `application.Authoring` with fake repository.
Confirmed `authoring.Load` propagates `CanManage` errors: non-owner instructor
`course_forbidden`, student `forbidden`. Existing dirty TE2/WIP files preserved.

Added `TestTeachCourseTreeRead` with owner 200 draft tree including module/lesson IDs
and titles, non-owner 403 `course_forbidden`, student 403 `forbidden`.
Test uses the existing `adminReadHandler` and `serveAdminRead`; teach registration
is intentionally still absent pending behavioral RED. No production changes yet.

`go test ./internal/course/adapters/http -run '^TestTeachCourseTreeRead$' -count=1`
exited 1 during setup: sandbox denied `/Users/strelia/Library/Caches/go-build/`.
This is BLOCKED, not behavioral RED. Stopped as explicitly instructed; no escalation
or cache workaround. `gofmt -w` on the test file completed exit 0.
Requested `gofmt -l internal/course cmd`, `go vet ./internal/course/... ./cmd/...`,
`go build ./cmd/api`, and `go test ./internal/course/...`: NOT_RUN after blocker.

Next: coordinator establishes behavioral RED, then adds the single GET mount,
test-harness registration, API wiring and matching OpenAPI binding; inspects route
coverage expectations and reruns requested checks. Generation remains coordinator-owned.
No commit, generation, independent review, mutation review or full gate performed.

### TE3 coordinator completion (Courses13, 2026-10-30)

Codex delivered only the test; the implementation was completed in-tree by the coordinator
(a 4-edit mechanical change, avoiding a re-run that kept hitting the sandbox cache block):
- `admin_handler.go`: added `RegisterTeach(mux, authenticate)` mounting `GET /api/v1/teach/courses/{id}`
  → reuses `getAdminCourse` unchanged (adminActor admits instructors; authoring.Load is owner-scoped).
  Read-only, no requireSameOrigin.
- `cmd/api/main.go`: `courseHandler.RegisterTeach(mux, identityHandler.Authenticate)` after RegisterAdmin.
- `openapi_contract.go`: `GET /api/v1/teach/courses/{id}` binding mirroring the admin courses/{id} GET.
- `admin_handler_test.go`: mounted `RegisterTeach` in the `adminReadHandler` helper so the teach path
  is reachable in the test.
Refusal codes (code-verified, differ from the document editor path): non-owner instructor →
`course_forbidden` (`authoring.Load` returns `CanManage`'s `ErrCourseForbidden` directly, authoring.go:78),
student → `forbidden` (`CanCreate` fails first). `TestTeachCourseTreeRead` asserts all three.

### Backend batch (TE1+TE2+TE3) verification — impact gate GREEN (Courses13, 2026-10-30)

From a full-access shell, after `make generate`:
- `gofmt -l internal/course internal/image cmd` → empty.
- `go vet ./internal/course/... ./internal/image/... ./cmd/...` → exit 0.
- `go build ./cmd/api` → ok.
- `go test ./internal/course/... ./internal/image/...` → all `ok` (course http 10.1s with the new teach
  document + course-tree tests; image http 3.28s with the teach upload test and `TestImageOpenAPIFidelity`
  now green post-generate).
- `make generate` → openapi.yaml carries all teach paths (document ×3, image, course); Orval client
  regenerated; staged; `git diff --exit-code` on generated clean + zero untracked → check-generated satisfied.
- frontend `typecheck` / `format:check` (373 files) / `lint` (oxlint + architecture 229 files) → all clean.

Deferred to TE6: `make integration` + full ordered `make check` incl. e2e. Batch cadence: 3 PRs (owner,
2026-10-30) — this is PR #1 (backend TE1-TE3), opened as GitHub PR #93.

### Independent review of PR #93 — PASS (independent_subagent, 2026-10-30)

Fresh-context reviewer, read-only, own verdict from the code (review_mode independent_subagent). **PASS,
zero blocking.** Independently ran `go build ./cmd/api` (exit 0) and `go test ./internal/course/...
./internal/image/...` (all ok). Verified against the real handlers: RegisterTeach methods are exact
mirrors of RegisterAdmin (same handlers, no authz/body change); AC7 same-origin parity exact (requireSameOrigin
on PUT document / POST publish / POST images, absent on both GETs); AC3 no sibling widening + admin/learn
unchanged; refusal codes correct per handler (document+image → `forbidden`; course-tree non-owner →
`course_forbidden`, student → `forbidden`); tests assert real observable behavior; generated artifacts
consistent; the image OpenAPI-coverage AST test was strengthened (asserts both RegisterAdmin+RegisterTeach,
len==2, fail-closed), not weakened.

Two non-blocking notes, accepted (not regressions): (1) the image teach test drives authorization via an
injected service error rather than a real access decision, and has no explicit student case — equivalent
because the image handler does no role-gate itself (identical to the admin path) and the real owner check
lives in the image domain (tested there); worth keeping in mind on a future change. (2) the course module
lacks an AST coverage test like the image module's — an asymmetry, not a gap (the teach course GET is
covered by a contract test + the OpenAPI fidelity tests). Merge of PR #93 awaits owner confirmation
(feature PR, not an A-batch self-merge).

## TE4 — 2026-10-30 (Codex), pre-code checkpoint

READY for the user-requested document data-layer slice. Owner/writer: Codex; root
`/Users/strelia/projects/courses`, branch `feat/teach-editor`, base
`e82aa4052e3b276e9962bc11dde7762285ec49d1`. Small implementation slice of the
approved phase plan (§11); no new planning decision or additional critique needed.
Acceptance: shared parameterized repository/query factories, unchanged admin API/keys,
teach operations and request-scoped composition; query and wire tests. Course-tree
and editor UI are outside this specific user brief. Existing generated-client staged
changes are inputs; unrelated .gitignore/security WIP is preserved. No generation/commit.
Inspected admin adapter, query factory/tests, document barrel, admin/teach composition,
teach queries and generated operations by direct file reads (graph refresh unnecessary).
Shared authoring factory stays internal because the barrel already exports a learner
`createDocumentQueries`. TE1–TE3 and generated teach functions are present.
Pinned Factory release verification: PASS, 447 files. Verification commands authorized
by the task; on sandbox tooling/cache failure stop and hand back to coordinator.
Review delivery: coordinator/manual handoff, exact independent reviewer pending; no
automatic spawn/contact. Author checks are not independent review or phase closure.

### TE4 implementation and verification

Implemented a shared adapter parameterized by the four generated operations and label
prefix, and an internal generic query factory parameterized by admin/teach. Admin
public functions, literal query-key helper, refusal maps, mapping and parsing remain
unchanged. Teach composition exposes request-scoped `documents` using the same options
as the existing teach repository. Updated the route-guard fixture for this added dependency.

Added `teach-document-queries.test.ts` (7 cases: mirrored admin behavior plus literal
teach keys and forwarding all four operations) and `teach-document-repository.test.ts`
(2 node/MSW wire cases: generated teach GET/PUT/publish/multipart requests and named
403 refusals for all four operations). Run by the coordinator 2026-10-30: PASS (see table).

Requested checks, frontend working directory, against the uncommitted TE4 delta at
`e82aa4052e3b276e9962bc11dde7762285ec49d1`:

| Command | Result |
| --- | --- |
| `npm run typecheck` | PASS, exit 0 after adding the missing teach document dependency to the route-guard fixture |
| `npm run format` | PASS, exit 0; 375 files |
| `npm run format:check` | PASS, exit 0 |
| `npm run lint` | PASS, exit 0; architecture checked 229 files |
| `npm run test` | PASS, exit 0 (coordinator rerun 2026-10-30, non-sandbox listener bind): `src/modules/document` + `src/routes/route-guards.test.tsx` → 100 tests, 0 failed, 0 skipped. Earlier sandbox attempt was BLOCKED, exit 1 before test execution: `listen EPERM: operation not permitted ::1:3000` |

An initial typecheck invocation accidentally ran at repository root and failed ENOENT
(no package.json); corrected to frontend. The first frontend typecheck identified the
route-guard fixture dependency, now fixed. No sandbox workaround or escalation used.
`git diff --exit-code` confirms admin composition and admin query tests unchanged;
their execution is now unblocked and green (coordinator rerun above). No build, e2e or
independent review performed yet. TE4 committed with the overlay/generated delta; next:
independent review before PR #2, and TE5/TE6 remain separate slices.

## TE5 — 2026-10-30 (Codex impl, coordinator verify)

Frontend teach editor route + nav. Delivered:
- New route `frontend/src/routes/teach.courses.$courseId_.lessons.$lessonId.tsx` — a near-exact
  mirror of the admin editor route (`admin.courses.$courseId_.lessons.$lessonId.tsx`): lazy
  `LessonDocumentEditor`, `ssr:false`, `resolveLessonMeta` from the course tree with a benign
  null-on-miss title block; differs only in `teach` context (`teach.detail`, `teach.documents`)
  and the `/teach` path. No `/admin/*` call from the teach flow.
- `frontend/src/app/composition/teach.ts`: new `detail(id)` query (key `['teach','courses','tree',id]`)
  over `getTeachCourse` (owner-scoped, from TE3), reusing the admin authoring mappers.
- `frontend/src/modules/admin/adapters/admin-course-repository.ts` + barrel: `wireError`/`courseTree`
  renamed to exported `mapAuthoringError`/`mapAuthoringCourseTree` and reused by teach — pure
  extract+rename, admin runtime unchanged (same logic, same call sites).
- `frontend/src/modules/document/application/admin-document-queries.ts`: the internal
  `createDocumentQueries` key-prefix generic widened to a `'admin' | 'teach'` union so both wrappers
  satisfy one editor props contract. Runtime keys/endpoints/invalidation byte-identical (admin still
  `['admin',...]`).
- Lesson-list on the teach course view extracted to a tested presentational component
  `frontend/src/modules/teach/ui/TeachLessonList.tsx` (the project tests components, not routes):
  module-grouped lesson links into the editor route, empty modules hidden, `EmptyState` (T13) when
  zero lessons across all modules (incl. zero modules). The route keeps pending/error/refetch.
- i18n keys `teach.lessons.empty` / `teach.lessons.loadError` added (uk/en).

Coordinator fix: the R24 route-tree census (`route-guards.test.tsx`) did not list the new editor
route — added it (globally-counted artefact: whoever adds one cleans up the census).

| Command | Result |
| --- | --- |
| `npm run format:check` | PASS, exit 0 |
| `npm run lint` | PASS, exit 0; architecture 231 files |
| `npm run typecheck` | PASS, exit 0 |
| `npm run test` (full) | PASS, exit 0; 95 files, 841 tests, 0 failed, 0 skipped (incl. 4 new `TeachLessonList` cases + R24 census) |

Note (Codex, honest scope): T13's empty-draft **editor** half is already covered by the existing
`LessonDocumentEditor.test.tsx` (zero sections → EmptyState); this slice adds the empty **lesson-list**
component test. No build/integration/e2e run here (e2e is TE6). Independent review of the full PR #2
delta (TE4+TE5) is the next gate before merge.

## TE6 — 2026-10-31 (e2e + full ordered gate; shared publish-path bug fixed)

New e2e `frontend/e2e/teach-editor.spec.ts` (two scenarios, mirror of `lesson-builder.spec.ts`):
- **T9** owner opens a lesson from `/teach`, edits and publishes it, and a member reads the
  published content — drives the real UI Publish, asserts the "Урок опубліковано" toast and zero
  `role="alert"`, and checks no `/admin/*` request leaks from the teach flow.
- **T5/T6** a non-owner instructor is refused by every teach write/read op (exact status + code,
  `forbidden` on document/image ops), and a student gets `course_forbidden` on the teach course tree.

**Defect found behind the green gate (pre-existing, shared admin+teach publish path).** T9 was RED
on first run and surfaced a real bug, not a test flaw: the publish contract returns **200 with an
empty body** (intentional — `openapi_contract.go:110,113` declare `Response{Empty:true}` for both the
admin and teach publish; `document_handler.go:132` writes a bare `WriteHeader(200)`; the OpenAPI
description says "re-read the draft for the new token"; generated `publishAdminLessonDocumentResponse200.data`
is `void`), but the frontend repository `publish` parsed `lessonDocumentUpdatedSchema` on that empty
body, failed, and returned `invalid_request` → the editor showed a **false "save failed" error after a
successful publish**. Admin's own lesson-builder e2e never caught it (it does not assert the publish
success toast). The misread was shared by `/admin` (merged, live) and the new `/teach` surface.

Fix (frontend-only; the backend contract was deliberate, so it was left untouched — verdict of an
independent pre-code critic, `scratchpad/te6-critic-b1.md`, confirmed by code):
- `DocumentAuthoringRepository.publish` now returns `DocumentAuthoringResult<void>`; the repository
  stops parsing the empty body and returns success on 200 (`admin-document-repository.ts`); the
  publish mutation's `mutationFn` returns void (`admin-document-queries.ts`), its invalidation of the
  draft + learner readers unchanged.
- `LessonDocumentEditor.onPublish` no longer reads a token from the publish response; it refreshes the
  optimistic-lock token from the authoritative re-read the publish `onSuccess` invalidation performs
  (`client.getQueryData(documents.document(lessonId).queryKey)`), **keeping the editor overlay** so
  edits made during the save→publish round-trip survive — symmetric with `onSave`, rather than dropping
  the overlay (which would have silently discarded in-flight edits; the omitted guard the critic flagged).
- Four unit mocks that returned a fabricated `{updatedAt}` body for publish (which the server never
  sends) were corrected to the real empty response (`teach-document-repository.test.ts`,
  `LessonDocumentEditor.test.tsx`, admin/teach `*-document-queries.test.ts`). The green-but-fake mock
  was why the bug hid behind unit tests — recorded as a gotcha.

| Command (full ordered `make check`) | Result |
| --- | --- |
| format-check, lint (arch 231), typecheck | PASS, exit 0 |
| test (frontend rstest + Go), integration, check-generated, build | PASS, exit 0 |
| e2e (full, not targeted — shared/persisted state) | **129 passed**, 18.5m; incl. `teach-editor.spec.ts` T9 (4.8s) and T5/T6 (1.1s) |
| GATE_EXIT (read from log, not the hook echo) | **0** |

Verified end-to-end through the production path: T9, RED before the fix, is GREEN after. Independent
review of PR #3 is the next gate before the owner merges and the phase closes.
