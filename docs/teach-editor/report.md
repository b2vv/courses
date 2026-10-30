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

Deferred to TE6: `make integration` + full ordered `make check` incl. e2e. Independent review of the
backend batch before merge still pending (mandatory-review-every-task). Batch cadence: 3 PRs (owner,
2026-10-30) — this is PR #1 (backend TE1-TE3).
