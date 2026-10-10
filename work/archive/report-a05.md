# A05 report — create and full-tree write

## Changed

- `backend/internal/course/adapters/http/admin_handler.go` — registered authenticated,
  same-origin `POST /api/v1/admin/courses` and `PUT /api/v1/admin/courses/{id}`; added
  strict JSON request DTOs and decoding, the exact 512 KiB `http.MaxBytesReader` cap,
  required-field and duplicate-key checks, index-derived module/lesson positions, RFC 3339
  microsecond token parsing/formatting, Create/Save application calls, and the A05 business
  error mappings.
- `backend/internal/course/adapters/http/admin_handler_test.go` — extended the A04 harness
  with stateful authoring/catalog behavior and OpenAPI contract evidence for P02–P06,
  B04–B08, B12, R12, R13, R20–R22 and R29. Added focused evidence for empty trees,
  index-derived positions, course/node field bounds, malformed/duplicate/unknown/trailing
  JSON, non-JSON content types, unexpected query parameters and non-microsecond tokens.
- `backend/openapi/openapi.yaml` — documented both mutation operations, the create and
  full-tree write bodies, exact admin timestamp schemas, the 512 KiB/strict-body contract,
  mutation responses and all Phase 4 business error codes. Generated output was left
  untouched for the reviewer.

## Decisions within A05

- The same-origin middleware wraps authentication on both mutations, preserving D-1's rule
  that a foreign origin is refused first. The handler then maps the authenticated transport
  actor through the existing A04 seam before decoding the body.
- Request strings use pointers only to distinguish a missing required JSON member from a
  present empty value. `instructorId` remains optional at transport level because an
  instructor may omit it; the existing application policy forces self-ownership while an
  admin omission reaches the existing `invalid_owner` rule.
- A small standard-library token walk rejects duplicate object members at every nesting
  level before `json.Decoder.DisallowUnknownFields` maps the body. This matches the existing
  login decoder's strictness without adding a dependency or changing inward layers.
- Admin timestamps are response strings formatted with exactly six fractional digits and
  PUT accepts the same exact layout. Public catalog timestamps were deliberately left on
  their existing contract.
- The handler constructs only editable fields and derives every `Position` from its array
  index. Status, publication time and position have no request DTO member and are rejected
  as unknown properties before any application/repository call.
- No domain, application, port, PostgreSQL, nginx or generated file needed a change; the
  existing Create/Save seams express every required outcome.

## Verified

- Local `gofmt -w` parsed and formatted
  `admin_handler.go` and `admin_handler_test.go`; final `gofmt -d` exited 0 with no output.
- Static acceptance scan found named contract evidence for every owned row: P02–P06,
  B04–B08, B12, R12, R13, R20–R22 and R29.
- Read-through against D-3a/b/c/d/e/g, D-4b and D-5 confirmed: 40-module/120-lesson/1000
  lesson-description bounds stay inward; the handler mirrors the 512 KiB cap; body
  validation precedes Create/Save; positions are zero-based array indexes; tokens serialize
  and parse at microsecond precision; owner/slug/concurrency errors retain their explicit
  status/code mappings; create replay remains `slug_taken`.
- Manual standards/spec review found no remaining confirmed defect. The git-based
  `code-review` workflow was denied by the brief's literal **never touch git** rule, so the
  mandated read-through was used instead.
- Only the three implementation/contract files listed above and this report were written.
  No file inside any `generated/` directory was written. No filesystem write was denied.

## Not run

Environment denial: **No Docker: no generation, no compile, no tests.** No workaround was
attempted.

- OpenAPI/Orval generation and `make generate` — not run; the reviewer regenerates, and no
  generated file was hand-written.
- Go compilation, targeted `go test`, contract tests and race tests — not run. The new cases
  are written evidence and are not represented as passing.
- OpenAPI document/request/response validation — not run because it is exercised by the Go
  contract tests, whose compilation/execution is denied.
- `make format-check` — not run because its Go formatter is Docker-backed. Local `gofmt` is
  the only permitted formatting evidence.
- `make lint` — not run because its Go vet step uses Docker and compiles Go.
- `make typecheck` and `make test` — not run because compilation/tests are denied.
- `make integration` — not run because Docker/Testcontainers are unavailable.
- `make check-generated` — not run because generation is denied and the target also uses
  git, which is separately forbidden.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` and
  aggregate `make check` — not run because Docker is unavailable and these exceed A05's
  permitted evidence.
- No git command was run; git state was not inspected or changed, and no commit was created.

## Limitation and next task

The OpenAPI client is intentionally stale until the Docker-enabled reviewer runs generation
and the ordered verification gates. That reviewer must validate the document, compile and run
the A05 contract suite, and return any mismatch rather than treating written tests as passed.
A06 is the next logical slice only after that review is green; status and delete endpoints
remain unimplemented.

## Reviewer verification (measured in the reviewing session)

`make generate` exit 0 (Orval extended the admin models for the write paths);
`make test` pass — the course HTTP contract suite now 4.6s with all A05 rows green;
`make typecheck`, `make format-check`, `make lint` pass; `make integration` pass
(12 packages ok). `check-generated` green after commit. Provenance: implemented with
Codex (gpt-5.6-sol); reviewed and verified here.
