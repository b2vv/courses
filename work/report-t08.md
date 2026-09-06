# T08 report — the admin-users browser journey

## Files

- `frontend/e2e/admin-users.spec.ts` — adds P08 against the three disposable identity fixtures
  with the shared admin `storageState`, a fail-fast preflight for each fixture, the complete
  block → restore → password reset → soft-delete journey, and the deleted-filter assertion.
- `frontend/e2e/route-guards.spec.ts` — adds the stored-instructor `/admin/users` not-found
  journey and R20's stored-student direct API refusal, including a post-refusal admin read that
  proves the attempted status mutation changed nothing.
- `work/report-t08.md` — this report.

## Decisions

- All browser actors reuse `admin.json`, `instructor.json`, or `student.json`. The new tests make
  zero login requests, so B19's seven-login budget is unchanged. The optional blocked-user login
  assertion was omitted because it would violate the brief's stricter zero-extra-login rule.
- P08 acts only on `...d1`, `...d2`, and `...d3`; none of the shared `...001`, `...002`, or
  `...003` fixtures is mutated. A stored-admin `beforeAll` reads all three disposables before the
  first mutation and names the stale email plus `make seed-identities` if any is missing,
  non-student, or non-active.
- `delete-me` is consumed once per seed. The spec records that terminal behavior beside the
  delete step; cross-run idempotency remains the reviewer's reseed responsibility exactly as the
  brief states.
- R20 covers both surfaces named by the plan and brief: the student receives `403 forbidden` on
  the directory GET and on a status POST. A stored-admin read then confirms `reset-me` is still
  active. The instructor browser guard additionally proves the route renders not-found before a
  directory request is sent.
- No app code, dependency, login flow, rate-limit configuration, `server.mjs`, generated file, or
  lockfile was changed.

## Written evidence

- P08: the admin sees the directory, blocks and restores `block-me`, resets `reset-me` with a
  unique-per-run password, deletes `delete-me`, sees it disappear from the default list, then
  sees it as deleted with no actions under the deleted filter.
- R20: the stored student session receives the named server refusal for both read and mutation;
  the subsequent stored-admin read proves no status mutation occurred.
- Route guard: the stored instructor session sees the not-found shell at `/admin/users`, and the
  test counts zero requests to the directory endpoint.
- Fixture freshness: all three disposable rows must be active students before the journey starts,
  so a stale local seed fails before partial mutation.

## Verification performed

- `cd frontend && npm run format:check` — passed; all 124 matched files formatted.
- `cd frontend && npm run lint` — passed with no warnings; architecture boundaries checked 70
  files.
- `cd frontend && npm run typecheck` — passed with `tsc --noEmit`.
- `cd frontend && npm test` — passed: 32 files, 293 tests, 0 failed/skipped/todo, no snapshot
  changes.
- `cd frontend && playwright test --list` — passed: 30 tests in 9 files; it discovered P08, R20,
  and the stored-instructor guard. The setup project still contains exactly the three shared-role
  login requests; the T08 tests add none.
- Two-axis named-file review was used because Git inspection is forbidden. The first spec review
  found the missing §5.f disposable preflight; it was added. The final standards and spec reviews
  both passed with no remaining findings.
- No Git command was run. Git state was not inspected or changed, and no commit was created.
- No file inside a `generated/` directory was written. `frontend/package-lock.json` was not
  edited. No filesystem write was denied.

## Denial and checks not run

Environment denial: **Docker/browser execution is unavailable here.** It was not attempted and
no workaround was used. The reviewer must run e2e twice, running `make seed-identities` between
the two executions for the 429 check and fresh `delete-me` fixture.

- `make format-check`, `make lint`, `make typecheck`, and `make test` — not run because their
  repository-wide forms include Docker-backed work; the permitted frontend equivalents passed
  as reported above.
- `make integration` — not run because Docker/Testcontainers are unavailable.
- `make check-generated` — not run because it invokes forbidden Git inspection and generation
  that may write inside `generated/` directories.
- `make build`, `make up`, `make seed-identities`, `make seed`, and `make smoke` — not run because
  the ordered repository gate cannot continue past the denied Docker integration step.
- `make e2e` and direct Playwright execution — not run because Docker/browser execution is
  unavailable. Discovery only was run; P08 and R20 are authored and typechecked, not represented
  as passing browser evidence.
- Aggregate `make check` — not run because it includes the denied checks above.

## Limitation and next task

P08 and R20 require the reviewer's real Compose/browser run before they are measured evidence.
After that review passes twice with the required reseed, T09 owns the clean-checkout ordered gate
and the Phase 4b final report update.

## Reviewer verification (measured in the reviewing session)

Stack rebuilt on the 4b image (migration 00004 applied, three disposable fixtures
seeded), then e2e run twice with a `make seed-identities` reseed between (delete-me is
terminal per 4b). Both runs 30/30 (27 prior + 3 admin-users journeys), each printing
`total=7, maximum consecutive=3` — the admin-users journey and the direct-API refusal add
zero logins, and the double run confirms the 429 non-regression. format-check, lint,
typecheck pass. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified
here.
