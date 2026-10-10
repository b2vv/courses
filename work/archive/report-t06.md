# T06 report — frontend identity administration adapter and application

## Files

- `frontend/src/domain/identity/administration.ts` and `administration.test.ts` — add
  `AdminUser`, `AccountStatus`, plain-string user ids, the administration result/error union,
  bounded directory queries, lowercase non-zero UUID validation, and the 1–1024 UTF-8-byte
  password rule.
- `frontend/src/ports/AdminUserRepository.ts` — adds the four-operation identity-administration
  port: list, set status, soft delete, and reset password.
- `frontend/src/application/admin/users.ts` and `users.test.ts` — add use cases that validate
  query, id, writable status, and password input before crossing the repository seam while
  returning named `AdministrationResult` outcomes unchanged.
- `frontend/src/adapters/api/admin-user-schema.ts` and `admin-user-schema.test.ts` — add strict
  Zod parsing for directory users and pages. UUIDs remain plain strings, and malformed ids,
  emails, names, roles, statuses, timestamps, members, pagination, and page sizes are refused.
- `frontend/src/adapters/api/admin-user-repository.ts` and
  `admin-user-repository.test.ts` — map the existing generated admin-user operations through
  `createApiTransport`, explicitly parse successful list data, serialize all three writes, and
  map every T05 status/business-code pair into the identity-administration union. The MSW table
  includes R19's `last_admin`, `cannot_target_self`, `forbidden`, and opaque 500 cases.
- `frontend/src/app/composition/admin-user-queries.ts` and
  `admin-user-queries.test.ts` — add stable list query options using imported
  `adminListStaleMs`, typed mutation options for status/delete/reset, and the named
  `invalidateAfterUserWrite` helper across every cached directory filter.
- `frontend/src/app/composition/admin.ts` — preserves the existing top-level course query API and
  extends it with the request-scoped nested `users` dependencies.
- `frontend/src/routes/route-guards.test.tsx` — extends the existing router fixture with the new
  nested user dependencies required by the composition type; no route behavior changed.
- `work/report-t06.md` — this report.

No route, screen, router-search parser, dependency, branded identity id, or generated file was
added or changed.

## Decisions

- User ids are validated lowercase, non-zero UUID strings but remain unbranded. The response
  schema reuses that validator so a parsed row can always be passed to a mutation use case.
- Block and restore stay one `setStatus` operation with an `active | blocked` input, matching the
  server endpoint and D-3. `deleted` is not a writable status and remains reachable only through
  delete.
- A directory response must echo the requested pagination, fit both the requested limit and
  reported total, match an explicit status filter, and contain no deleted row under the default
  filter. Generated response types never leave the adapter.
- Named business codes are accepted only at their T05 HTTP statuses. `internal_error`, unnamed
  5xx, and 429 become `dependency_unavailable`; malformed or mismatched responses become
  `invalid_response`.
- `createAdminDependencies` computes the request options once, builds both repositories for that
  router request, spreads the existing course dependencies unchanged, and exposes user
  administration at `admin.users` without a new root context key.
- Every user write invalidates the shared `['admin', 'users', 'list']` prefix. There is no detail
  resource or optimistic token to cache or invalidate.

## Written evidence

- R19: the MSW table covers all named T05 mappings and specifically proves `409 last_admin`,
  `409 cannot_target_self`, `403 forbidden`, and `500 internal_error` →
  `dependency_unavailable` through the application use cases.
- Malformed-response refusal: strict Zod cases cover every directory field, pagination, unknown
  members, the 50-row cap, and lowercase/non-zero ids; the repository converts malformed success
  data to `invalid_response`.
- Query-option shape: the list key contains status, search, limit, and offset; its stale time is
  the imported `adminListStaleMs`; all three write option factories invalidate active and deleted
  list variants through `invalidateAfterUserWrite`.
- Application/domain boundaries: tests pin query normalization, UTF-8 password bytes, ids,
  pre-port validation, and unchanged business refusals.

## Verification performed

- `cd frontend && npm run lint` passed: oxlint reported no warnings and the architecture check
  reported `Architecture boundaries checked (67 files).`
- `cd frontend && npm run typecheck` passed with `tsc --noEmit` after the final review fix.
- A two-axis named-file review was used because Git inspection is forbidden. Standards found no
  hard violation; its only low-confidence duplication note concerned deliberately local fake
  repositories in tests and was retained under ponytail-full. Spec review found one uppercase-id
  schema gap; the schema now reuses `validAdminUserId`, and the refusal test pins it. Its note
  about “four mutations” was not acted on: §1 defines four operations total (one list plus three
  writes), while D-3 deliberately combines block and restore in one status endpoint. Follow-up
  spec review confirmed no remaining blocking T06 issue.
- Manual inspection confirmed that only `admin-user-repository.ts` imports the generated client,
  and no generated model crosses the adapter boundary.
- No file inside a `generated/` directory was written. No filesystem write was denied.
- No Git command was run; Git state was not inspected or changed, and no commit was created.

## Not run

Environment denial: **Docker is denied; Rstest is denied because binding `::1:3000` returns
`EPERM`; oxfmt is denied.** No workaround was attempted, and no unrun check is represented as
passing.

- Targeted/new Rstest and `npm test` — not run. R19, schema, domain, application, and query-option
  tests are written for the reviewer.
- `npm run format:check` and `make format-check` — not run because oxfmt is denied; the make target
  also invokes Docker-backed Go formatting. Neighbor formatting was matched manually.
- `make lint` — not run because its Go vet half requires denied Docker. Only the frontend-local
  lint command reported above ran.
- `make typecheck` — not run; its underlying frontend command was run directly and passed.
- `make test` — not run because it includes denied Rstest and Docker-backed Go tests.
- `make integration` — not run because it requires Docker/Testcontainers.
- `make generate` and `make check-generated` — not run: generation would write inside
  `generated/`; the drift target also requires Docker and invokes Git. These actions are expressly
  forbidden by the brief. B09 therefore awaits the reviewer environment.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`, and
  aggregate `make check` — not run because they require Docker and/or denied test execution.

## Limitation and next task

A Docker-enabled reviewer must run the written Rstest evidence, regenerate/check Orval without
editing the handwritten T06 files, and run the ordered repository gates. T07 may begin only after
that review is green; it owns `/admin/users`, search parsing, views/actions, the reset-password
dialog, account-administrator guard, localization, and accessibility.

## Reviewer verification (measured in the reviewing session)

`make format` applied (oxfmt unrunnable in sandbox); format-check, lint, typecheck, test
(0 failed frontend tests including R19 mapping, B09 schema refusal, query-option shapes)
and check-generated pass. Mirrors the 4a admin-course slice. Provenance: implemented with
Codex (gpt-5.6-sol); reviewed and verified here.
