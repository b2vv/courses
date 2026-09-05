# A13 — shared browser sessions

## Delivered

- `frontend/e2e/auth.setup.ts` signs in exactly once as student, instructor and admin through
  the real host-originated login endpoint, then writes
  `frontend/playwright/.auth/<role>.json`.
- The Playwright `setup` project runs before both browser projects. Its `login-budget`
  teardown runs after the dependent projects, prints the measured login total and asserts
  B19's amended `total <= 7` and `maximum consecutive <= 3` limits.
- The request ledger records the three completed setup POSTs and the actual browser request
  stream around the retained form and logout tests. A login increments the consecutive run;
  any observed non-login browser request resets it. The final full-run tally is 7 logins:
  three setup POSTs, two required login-form submissions and two disposable logout sessions.
- The two `returnTo`/SSR tests still use the real form and are counted. The route-guard test
  that signs out now provisions its own student session through `page.request` before its first
  navigation, so its logout cannot revoke `student.json` for another test.
- The keyboard sign-out test in `accessibility.spec.ts` likewise provisions and destroys its
  own student session. The read-only admin accessibility test remains on `admin.json`.
  `shell.spec.ts` and `recovery.spec.ts` contain no login and remain untouched.
- `/frontend/playwright/.auth/` is ignored. No rate limit, `server.mjs`, or admin browser journey
  was added.

Implementation choice: direct POST provisions state because it avoids coupling session setup
to login-page UI while exercising the same host-originated, rate-limited endpoint; the two
existing form tests retain browser coverage.

## Review

- Standards axis: no documented-standard violation. The initial ambiguous request-group naming
  was replaced with an explicit `login`/`other` request stream. The two pre-existing local
  `/auth/me` counters remain local; extracting a helper for two assertions would add more API
  than it removes under ponytail-full.
- Specification axis: the first implementation counted voluntary labels and did not prove real
  consecutive HTTP traffic. Review returned it; the final implementation records completed
  setup responses and observes browser requests independently of the form helper. Re-review
  found no remaining A13 gap or scope creep.
- Live-stack review then exposed a session-ownership collision: both logout tests consumed the
  same `student.json`, while logout revokes every presented session. Each destructive test now
  owns a disposable context session; shared storage state is reserved for read-only tests. The
  reviewer amended B19 from 6 to 7 because all seven logins are intrinsic to the retained
  scenarios.

## Verification and denials

No denied or unrun check is represented as passed.

- `npm run format:check`: pass after the final changes.
- `npm run lint`: pass; Oxlint and the frontend architecture check are green.
- `npm run typecheck`: pass.
- An earlier `npm test` run passed 243 tests before the final request-stream correction.
  The final rerun was denied exactly with `listen EPERM: operation not permitted ::1:3000`.
  The host/port was not changed and the denial was not worked around, so the earlier result is
  not claimed as final-code evidence.
- Docker is denied by the brief. Therefore the Go halves of `make format-check`, `make lint`
  and `make test`, plus `make integration`, `make check-generated`, `make build`, `make up`,
  both seeds and `make smoke` were not run.
- Browser launch and every Playwright run are denied by the brief's macOS Mach-port limit.
  `make e2e`, its required two consecutive runs, and B19's runtime print/assertion were not run.
- `make typecheck` and the full ordered `make check` targets were not run; only the direct
  frontend checks listed above were executed.
- Git was not invoked. No file inside a `generated/` directory was written.

## Limitation and next task

B19 exists as executable Playwright evidence but the disposable-session correction remains
unverified in this session. The reviewing session must run `make e2e` twice against Compose
with a cold login bucket and confirm both runs print exactly 7 total logins and at most 3
consecutive. Once that passes, A14 can add the admin and instructor browser journeys using the
stored role states.

## Reviewer verification (measured in the reviewing session)

Two review rounds. Round 1 (against the live stack): the shared student.json token is one
server session and logout revokes every presented session (1292b3a) — two logout tests
(route-guards ~142, accessibility ~92) killed the shared token for each other; reproduced
with curl (authentication_required). Also caught: the RHF lockfile broke Docker `npm ci`
(missing wasm optionals) — regenerated in-container. Round 2: the honest fix tallied 7
logins, over my plan-time ≤6; Codex stopped and reported the exact count rather than
forcing it. B19 amended to ≤7/≤3-consecutive (dated) — both returnTo and SSR login-flow
tests must use the form, two logout tests each need a disposable session (3+2+2=7).

Final fix: logout tests provision their own session via a context-level API login; read
tests keep shared state. **Validated empirically with two consecutive `make e2e` runs**
(the historical 429 trigger): both passed 24/24, each printing `total=7, maximum
consecutive=3`. Static gates (format-check, lint, typecheck, test) pass. Provenance:
implemented with Codex (gpt-5.6-sol); reviewed, corrected and verified here.
