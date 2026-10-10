# A06 report — lifecycle endpoints and mutation refusals

## Changed

- `backend/internal/course/adapters/http/admin_handler.go` — registered same-origin,
  authenticated `POST /api/v1/admin/courses/{id}/status` and
  `DELETE /api/v1/admin/courses/{id}` adjacent to the A05 mutations; added strict JSON
  request DTOs, exact RFC 3339 microsecond token parsing, calls to the existing Transition
  and Delete use cases, tree/204 responses, and transport mappings for
  `invalid_status_transition` and `course_not_deletable`.
- `backend/internal/course/adapters/http/admin_handler_test.go` — made the existing fake
  repository stateful for transition/delete and added OpenAPI contract cases for every A06
  row: P11–P14, B01, R08–R11, R14, R15 and R23. The refusal matrix covers create, tree save,
  status and delete; ownership/origin refusals assert no state change, and R11 performs a
  following HTTP read. A successful draft delete also exercises the new `204` contract.
- `backend/openapi/openapi.yaml` — documented both lifecycle operations, their strict
  request bodies and concurrency token, response sets, ownership/same-origin behavior, and
  the status/delete request schemas. Generated output remains untouched for the reviewer.
- `work/report-a06.md` — this report.

## Decisions

- Both endpoints use the established A05 order: same-origin middleware wraps
  authentication, actor mapping happens before request decoding, query parameters are
  rejected, and strict JSON decoding rejects duplicate/unknown members and bodies above the
  existing 512 KiB ceiling.
- Status carries `{status, expectedUpdatedAt}`. The OpenAPI target enum includes `draft` so
  a syntactically valid but illegal `published -> draft` request can reach the domain policy
  and produce R14's required `409 invalid_status_transition` instead of becoming a schema
  rejection.
- Delete carries `{expectedUpdatedAt}` in a strict JSON body, matching the existing
  application and repository signatures. Success returns `204` with no response body.
- The handler contains no lifecycle policy: instructor ownership, legal transitions,
  draft-only deletion, optimistic locking and in-transaction disambiguation stay in the
  existing domain/application/PostgreSQL seams.
- No frontend, domain, application, port, PostgreSQL, nginx, A05 handler behavior or
  generated file was changed.

## Verified

- Local `gofmt -w` parsed and formatted `admin_handler.go` and
  `admin_handler_test.go`; final `gofmt -d` exited 0 with no output.
- Static acceptance scan found named evidence for P11, P12, P13, P14, B01, R08, R09, R10,
  R11, R14, R15 and R23.
- Manual request/response read-through aligned both handlers with their OpenAPI operations:
  path IDs, strict body members, six-digit timestamp carriers, `200` tree and `204` empty
  success responses, `no-store`, and the 400/401/403/404/409/500/503 response sets.
- Manual standards/spec review against `AGENTS.md`, brief A06 and plan D-3f/D-4a/D-5 found
  no remaining confirmed defect. The normal `code-review` skill workflow was denied because
  it requires Git fixed-point/diff commands and the brief says never touch Git; no workaround
  or diff-based subagent was used.
- Only the four files listed above were written. No file inside a `generated/` directory was
  written. No filesystem write was denied.

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
  Git, which is separately forbidden.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` and
  aggregate `make check` — not run because Docker is unavailable and these exceed A06's
  permitted evidence.
- No Git command was run; Git state was not inspected or changed, and no commit was created.

## Limitation and next task

The generated OpenAPI client is intentionally stale until a Docker-enabled reviewer runs
generation and the ordered verification gates. That reviewer must validate the OpenAPI
document, compile and run the A06 contract suite, regenerate the client, and report any
mismatch rather than treating these written tests as passed. A07 is the next logical slice
only after that review is green.

## Reviewer verification (measured in the reviewing session)

`make generate` exit 0; `make test` pass (course HTTP contract suite 7.6s with the
lifecycle rows green); `make typecheck`, `make format-check`, `make lint` pass;
`make integration` pass (12 packages). `check-generated` green after commit.
Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
