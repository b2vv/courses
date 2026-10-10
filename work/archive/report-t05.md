# T05 report — HTTP surface for identity administration

## Files

- `backend/internal/identity/adapters/http/admin_handler.go` — adds the admin-user DTOs,
  bounded list-query parsing, strict flat-object JSON decoding, UUID validation, the
  admin-role gate, and handlers for list, block/restore, soft delete, and password reset.
- `backend/internal/identity/adapters/http/admin_handler_test.go` — adds kin-openapi
  contract evidence for P01, P02, B01, B04, R01–R08, R11, R13, and R14, plus strict
  body/query bounds and the complete business/dependency/internal error mapping.
- `backend/internal/identity/adapters/http/handler.go` — injects the existing
  `Administration` use-case seam, registers the four routes through the existing
  `Authenticate`/`RequireSameOrigin` middleware, reuses one account-administrator gate
  for users and instructors, and maps the four new business codes.
- `backend/cmd/api/main.go` — constructs `application.Administration` with the existing
  identity repository and clock and passes it to the identity handler.
- `backend/internal/identity/adapters/http/handler_test.go`,
  `backend/internal/identity/adapters/postgres/schema_integration_test.go`, and
  `backend/internal/course/adapters/http/admin_handler_test.go` — update existing
  handler construction for the new dependency; tests that do not register/use the
  admin-user surface pass `nil`.
- `backend/openapi/openapi.yaml` — documents the three path templates/four operations,
  list filters and pagination, request/response schemas, no-store headers, and
  `cannot_target_self`, `admin_immutable`, `last_admin`, and `user_not_found`.
- `work/report-t05.md` — this report.

No frontend, domain, application, port, PostgreSQL adapter, SQL, migration, or dependency
was changed. No file under a `generated/` directory was written.

## HTTP behavior

- `GET /api/v1/admin/users` defaults to limit 20, offset 0, and the repository's
  non-deleted view. `limit` is bounded to 1–50, `offset` to 0–10000, `q` is trimmed and
  limited to 100 valid Unicode code points with NUL refused, and status accepts only
  `active`, `blocked`, or `deleted`. Unknown, repeated, or malformed query parameters
  return `400 invalid_request`. Literal search escaping remains in the existing T04
  PostgreSQL adapter.
- List rows expose only id, email, displayName, role, projected status, and a UTC
  microsecond `updatedAt`; the page uses the existing `{items, pagination}` shape and
  always emits `items: []` for an empty result. Every response is `Cache-Control:
  no-store`.
- Status and password bodies reuse the identity package's 8 KiB cap. They require
  `application/json`, exactly one expected member, a JSON string value, and no duplicate,
  unknown, trailing, malformed, oversized, or invalid-UTF-8 content. Status accepts only
  active/blocked. Password accepts any content of 1–1024 UTF-8 bytes, including
  whitespace-only and same-as-current values.
- Mutation order is the reused seam verbatim: same-origin first, then authentication,
  then `CanAdministerAccounts`, then UUID/query/body validation, then the application use
  case. A malformed UUID is therefore refused before any administration call.
- Status returns 200; delete and reset return 204. Unknown/deleted targets return 404;
  self-target, immutable-admin, and last-admin refusals return 409; non-admins return 403;
  anonymous/disabled callers return 401; dependency and unexpected errors remain 503/500
  with safe envelopes.

## Written evidence

- P01/P02/B01: DTO/status projection, default versus deleted filter input, pagination,
  literal `%_!` search input, and an explicit empty array/zero total.
- B04: empty and over-1024-byte refusal; exact 1024-byte ASCII and multibyte acceptance;
  multibyte overflow refusal; whitespace-only and same-as-current acceptance.
- R01–R05: anonymous, student, instructor, disabled-admin, and cross-origin matrices;
  the administration use case is not reached on middleware/role refusals.
- R06–R08/R11: self block/delete, immutable-admin delete, and unknown targets across all
  three mutations map to their exact codes. The same table also covers `last_admin`, 503,
  and opaque 500 mapping.
- R13/R14: malformed UUID and illegal `deleted` status return `invalid_request` before an
  administration call.
- The brief's extra reset wording is covered explicitly: a blocked target reaches reset
  and returns 204, while a deleted target's `user_not_found` returns 404.
- Strict transport evidence covers unknown/duplicate members, malformed/trailing JSON,
  invalid UTF-8, oversized bodies, invalid list bounds, repeated/unknown query members,
  and NUL/overlong search.

## Verification performed

- Ran `gofmt -w` on every touched Go file, then `gofmt -d`; the final output was empty.
- Read every `NewHandler` call site after the signature change; production supplies the
  real `Administration`, while unrelated existing harnesses supply `nil` and never call
  the new routes.
- Read the route nesting against plan §4 D-1 and confirmed `RequireSameOrigin` is outermost
  on every mutation and the list uses `Authenticate` only.
- Read the four handler calls against `application.Administration`; no missing seam or
  non-HTTP adapter change was required.
- Read the OpenAPI operations, schemas, response references, error enum, and no-store
  headers against each handler outcome.
- The required two-axis review was adapted to the brief's no-Git rule: reviewers inspected
  the named files directly. Standards found no documented violation; its production
  duplication note was removed by sharing the admin gate. Spec review's invalid-UTF-8 and
  B04 gaps were fixed. Its requested post-refusal database readback was not added to a
  fake HTTP contract because that state invariant belongs to the already-delivered
  application/PostgreSQL seams; the HTTP evidence proves the exact refusal and no extra
  administration call.

## Denials and checks not run

The brief explicitly denies Docker, generation, Go compilation, tests, and every Git
operation. None was run or worked around, and no skipped check is represented as passed:

- `make format-check`, `make lint`, and `make typecheck` — not run; their backend portions
  require denied Docker/Go compilation. Direct `gofmt` is the only executed formatter.
- `make test` — not run; Go compilation/tests are denied. The kin-openapi cases are written
  evidence awaiting the reviewer environment.
- `make integration` — not run; Docker/Testcontainers are denied.
- `make generate` — not run; generation and generated-directory writes belong to the
  reviewer. No sqlc or Orval output was written.
- `make check-generated` — not run; it requires generation and invokes Git, both forbidden.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`,
  and aggregate `make check` — not run; they require denied compilation or Docker and
  include work outside T05.
- Commit/diff/status/log operations — not run because the brief says no Git, overriding the
  implementation and code-review skills' normal Git steps.

No write was denied, so the brief's exact-text stop condition did not trigger. No Git
command was run.

## Limitation and next task

The reviewer must regenerate Orval from the changed OpenAPI document and compile/run all
written contract tests in the Docker-enabled review environment. T06 is next: the
frontend admin-users domain/application/API adapter, composition, queries, and generated
Orval client.

## Reviewer verification (measured in the reviewing session)

`make generate` (four admin/users routes + schemas in OpenAPI and Orval); `make test`
identity HTTP contract suite 9s green with the full refusal matrix and error mapping;
`make integration` 12 packages ok; format-check, lint, typecheck, check-generated pass.
Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
