# A04 report — authorization seam and protected reads

## Changed

- `backend/platform/httpserver/actor.go` and `actor_test.go` — added the transport-only
  `Actor`, `WithActor` and `ActorFromContext`, with absent, present and wrong-type context
  coverage.
- `backend/internal/identity/adapters/http/handler.go` and `handler_test.go` — exported
  `Authenticate` and `RequireSameOrigin` without changing login/logout behavior;
  authentication now carries both the existing identity user and the cross-module actor;
  added the authenticated, admin-only instructor list and its explicit `forbidden` mapping.
- `backend/internal/identity/domain/user.go` and `user_test.go` — added and covered
  `Role.CanAdministerAccounts()`, true only for `admin`.
- `backend/sql/queries/identity.sql` — added `ListInstructors`, limited to 200 active
  instructor/admin accounts and ordered by C-collated display name with ID as the
  deterministic tiebreak.
- `backend/internal/identity/adapters/postgres/repository.go` — mapped `ListInstructors`
  rows through the existing validated identity user mapping and collapsed persistence/data
  failures to `dependency_unavailable`.
- `backend/internal/identity/adapters/postgres/schema_integration_test.go` — supplied the
  existing repository to the expanded identity HTTP composition constructor.
- `backend/internal/course/adapters/http/handler.go` and `handler_test.go` — attached the
  existing authoring use cases to the course handler while preserving the public catalog
  registration and tests.
- `backend/internal/course/adapters/http/admin_handler.go` and `admin_handler_test.go` —
  added `RegisterAdmin`, strict list query parsing, explicit identity-role to course-role
  mapping, owner-scoped list and tree reads, DTO/error mapping, `Cache-Control: no-store`,
  and OpenAPI-validated contract cases for P07 and R01–R07 plus admin/instructor success,
  search parameters and wrong method.
- `backend/openapi/openapi.yaml` — documented the three GET operations, admin list/tree and
  instructor schemas, protected response headers, and the A04 error codes `forbidden` and
  `course_forbidden`.
- `backend/cmd/api/main.go` — wired the existing course repository into authoring, supplied
  the identity repository as the instructor reader, and passed the exported middlewares to
  `RegisterAdmin` exactly at the composition root.

## Decisions within A04

- The identity HTTP handler receives a narrow private `instructorLister` dependency directly
  from the composition root. No identity application use case, port or new domain model was
  added because the brief forbids domain/application changes beyond
  `CanAdministerAccounts()` and names only the handler, query, repository wiring and role
  predicate for this read.
- `ListInstructors` selects enough columns to reuse `domain.NewUser` validation internally;
  the HTTP DTO deliberately emits only `id`, `displayName` and `role`.
- Course role mapping is an explicit three-arm switch. The mapped actor immediately asks the
  existing course-domain `CanCreate` policy, so the role refusal precedes query/path
  validation; the application boundary retains its own policy check.
- `RegisterAdmin` accepts both middleware functions now, as D-1 requires. The same-origin
  argument stays unused until A05 introduces the first mutation; no speculative mutation
  registration was added.
- The admin list accepts only `limit`, `offset` and `status`, defaults to `limit=20` and
  `offset=0`, and uses the existing authoring `List` validation. The tree read rejects every
  query parameter.

## Verified

- Local `gofmt -w` parsed and formatted every changed Go file; a final `gofmt -d` pass is
  recorded below when empty.
- Static read-through against plan D-1, D-4c and D-5 confirmed: the course package imports no
  identity package in production code; authentication supplies a plain transport actor;
  unknown roles fail closed; instructors are admin-only; course ownership remains in the
  course domain/application; all protected responses are `no-store`; and no email is present
  in the instructor DTO.
- Acceptance read-through confirmed named contract cases for P07 and R01–R07. Anonymous,
  student, disabled and malformed-ID cases assert no course/instructor repository call;
  foreign ownership is checked after the one required load; P07 checks the exact instructor
  and admin fixture payload with exactly three members per object.
- Two-axis manual review against `.claude/standards.md`, `docs/architecture/backend.md`,
  `AGENTS.md` and `work/brief-a04.md` found no remaining confirmed standards or specification defect. The
  normal code-review skill workflow was not used because it requires Git commands and
  parallel review tooling unavailable in this session; the brief forbids Git literally.
- No file inside a `generated/` directory was written. No filesystem write was denied.

## Not run

Environment denial: **No Docker**. Per the brief, it denies sqlc/Orval generation and every
Go compile or test; no workaround was attempted.

- `make generate`, sqlc and Orval generation — not run. The reviewing session must generate
  `ListInstructors` and the OpenAPI clients; generated files were not hand-written.
- Go compilation, targeted `go test`, and contract tests — not run because the brief forbids
  Go compile/test in this environment. P07 and R01–R07 are written evidence, not claimed pass.
- `make format-check` — not run because its Go formatter runs through Docker. Local `gofmt`
  is the only permitted formatting evidence.
- `make lint` — not run because its Go vet portion runs through Docker and compiles Go.
- `make typecheck` — not run because it compiles Go.
- `make test` — not run because it compiles and runs Go tests.
- `make integration` — not run because Docker/Testcontainers are unavailable.
- `make check-generated` — not run because generation is denied and the target uses Git,
  which the brief separately forbids.
- `make build` — not run because it compiles Go and uses Docker.
- `make up`, `make seed-identities`, `make seed` and `make smoke` — not run because Docker is
  unavailable and they exceed A04's permitted evidence.
- `make e2e` — not run because Docker/Compose are unavailable.
- Aggregate `make check` — not run because its denied steps cannot execute.
- No Git command was run, no Git state was inspected or changed, and no commit was created.

## Limitation and next task

The existing identity generated package is intentionally stale until the Docker-enabled
reviewing session runs `make generate`. That session must confirm the expected sqlc
`ListInstructors` row symbol, validate the OpenAPI document and responses, and run the ordered
gates through `make integration`. A05 remains the next slice only after that review is green.

## Reviewer verification (measured in the reviewing session)

`make generate` exit 0 — sqlc `ListInstructors` and the Orval client with the new admin
models regenerated cleanly against the hand-written expectations. `make test` pass
(frontend suite + all Go packages with race; the new contract tests for P07/R01–R07
green), `make typecheck` pass, `make integration` pass (course 5.5s, identity 4.7s),
`make format-check` and `make lint` pass. `check-generated` green after commit.
Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified here.
