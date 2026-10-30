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
