# A14 — the admin browser journey

## Delivered

- `frontend/e2e/admin.spec.ts` adds the P20 admin journey with the stored admin state:
  create a uniquely named course, add and save one module and lesson, publish it, observe it
  in the public catalog and public program, observe the published admin row, archive it, and
  observe its removal from the public catalog.
- The P21 instructor journey uses the stored instructor state, observes owner scoping against
  a fresh admin-owned draft, creates and edits an own course, proves the owner picker is absent,
  publishes and observes the own row, and receives a browser-rendered refusal plus HTTP 403 on
  the foreign editor route. Its published course is archived afterward.
- R33 uses the stored student state and `page.request`: the list and a valid-origin create are
  both refused with `403 forbidden`. An authorized create of the same slug then succeeds and is
  deleted, proving the refused request made no mutation.
- Every slug uses a per-run `e2e-admin-*` namespace. Draft fixtures are deleted; only archived
  journey courses remain, which persistent local volumes tolerate by namespace.
- The spec contains no login request. It reuses `admin.json`, `instructor.json`, and
  `student.json`, so A14 adds zero logins to B19.

Ponytail-full kept the change to the requested spec and this report. Existing storage-state,
request-tracking, accessible selector, and API-context mechanisms were reused; no application
code, dependency, or e2e-only product seam was added.

## Review

- Standards axis: clean after final review; no documented-standard breach or material Fowler
  smell remains.
- Specification axis initially found three literal gaps: `a14-*` instead of the required
  `e2e-admin-*` namespace, an undeleted foreign draft, and a seeded page-one assertion that
  would decay as persistent volumes accumulated courses. All three were corrected. Final
  re-review is clean for P20, P21, R33, cleanup, idempotence, and the zero-login constraint.

## Verification and denials

No denied or unrun check is represented as passed.

- `npm run format:check`: pass on 108 files after the final change.
- `npm run lint`: pass; Oxlint and the frontend architecture check are green.
- `make typecheck`: pass.
- Browser-free Playwright discovery:
  `playwright test e2e/admin.spec.ts --list` passes and lists the three A14 Chromium tests under
  the existing three setup tests and login-budget teardown (7 listed tests total).
- An earlier `npm test` run passed 243 tests before the final namespace and cleanup corrections.
  The final-code rerun was denied exactly with
  `listen EPERM: operation not permitted ::1:3000`. The bind address was not changed and the
  denial was not worked around, so the earlier result is not claimed as final-code evidence.
- Docker is denied by the brief. Therefore the Go halves of `make format-check`, `make lint`,
  and `make test`, plus `make integration`, `make build`, `make up`, `make seed-identities`,
  `make seed`, and `make smoke` were not run. The complete `make format-check`, `make lint`, and
  `make test` targets are consequently unrun.
- `make check-generated` was not run: it invokes Docker-backed generation, writes under
  `generated/`, and executes git checks, all forbidden here.
- Browser launch is denied by the brief. `make e2e`, the required second consecutive e2e run,
  the runtime P20/P21/R33 assertions, and B19's measured login print/assertion were not run.
- The full ordered `make check` was not run because it contains the denied Docker, generation,
  git, Compose, and browser steps above.
- Git was never invoked. No file inside a `generated/` directory was written, and
  `frontend/package-lock.json` was not modified.

## Limitation and next task

A14 exists as statically checked executable browser evidence, but the live Compose behavior is
unverified in this environment. The reviewer must run `make e2e` twice against Compose, confirm
P20/P21/R33 pass on both runs, and confirm B19 remains at no more than 7 total and 3 consecutive
host-originated logins. If green, A15 owns the full ordered gate and `docs/admin/report.md`.

## Reviewer verification (measured in the reviewing session)

`make format-check`, `make lint`, `make typecheck`, `make test` pass. e2e run twice
against live Compose: both 27/27 (24 prior + 3 admin journeys), each printing
`total=7, maximum consecutive=3`. The two consecutive passes also confirm the per-run
unique slug namespace is idempotent (the created course does not collide on rerun) and
the login budget is unregressed. Provenance: implemented with Codex (gpt-5.6-sol);
reviewed and verified here.
