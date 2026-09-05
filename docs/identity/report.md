# Identity Phase 3 delivery report

Updated: 2026-09-05. Current implementation task: I06; the ordered completion gate awaits the
reviewing session.

## I01 — users, sessions and domain

1. **Implemented.** Added the Identity domain foundation for users, roles and time-bounded sessions, plus the PostgreSQL schema for credential records and opaque-session digests.
2. **Files and modules.** `backend/internal/identity/domain` owns stable errors, `User`, `Role`, `Session`, email normalization and capability policy. `backend/migrations/00002_identity.sql` owns `users` and `user_sessions`. `backend/internal/identity/adapters/postgres/schema_integration_test.go` verifies the schema against PostgreSQL 17.5.
3. **Architecture decisions.** Domain objects contain no password hashes, raw session tokens, pgx types or transport DTOs. Persistence stores an Argon2id PHC string beside the user record and only a unique 32-byte session-token digest. Session lifetime uses the half-open interval `created_at <= now < expires_at`. `student` can learn; `instructor` and `admin` can enter the administration shell.
4. **Verification commands.** Targeted TDD red/green runs used the pinned `golang:1.27.1-bookworm` toolchain. The final I01 gate ran `gofmt` cleanliness, `go vet ./internal/identity/domain`, `go test -race ./internal/identity/domain`, and `go test -tags=integration ./internal/course/adapters/postgres ./internal/identity/adapters/postgres`.
5. **Passed checks.** Domain tests pass with the race detector. Migration up/down/up, database constraints, unique digests and user-to-session cascade pass in Testcontainers. The existing Course repository integration suite also passes with migration 00002 present.
6. **Remaining limitations.** I01 does not authenticate credentials, create/revoke sessions, expose HTTP endpoints, seed demo identities or render authenticated frontend state. Those behaviors belong to I02–I06.
7. **Next smallest task.** I02 adds only the application ports/use cases and Argon2id/PostgreSQL adapters, with fake-port tests and transactional rotation integration evidence.
8. **User decision.** No product decision blocks I02. The GitHub remote remains unconfigured because external repository creation/push needs explicit approval for the exact destination.

## I02 — credential and session application slice

1. **Implemented.** Added login, authenticate and logout use cases; real Argon2id verification; sqlc Identity queries; and a PostgreSQL session adapter with opaque tokens, digest-only persistence, expiry cleanup, revocation and bounded collision retry.
2. **Files and modules.** `backend/internal/identity/application` owns orchestration and the 12-hour lifetime. `ports/authentication.go` defines only the credential and session dependencies used by those cases. `adapters/argon2id` owns bounded PHC parsing and constant-time digest comparison. `adapters/postgres/repository.go` owns credential lookup, user mapping and transactional session operations. Identity-generated sqlc code lives under the adapter.
3. **Architecture decisions.** Password hashes stay inside the credential adapter. A raw token crosses the session port only in an explicitly secret-bearing issuance result and never enters domain/generated persistence models. Login rotation confirms one transaction before returning a token; unique collisions retry the full transaction up to three times. An indeterminate commit fails closed without a cookie claim. Application and adapter boundaries both validate untrusted tokens.
4. **Verification commands.** TDD red/green runs covered application and Argon2id packages. The final backend gate ran `gofmt`, `go vet ./...`, `go test -race ./...`, and `go test -tags=integration ./internal/course/adapters/postgres ./internal/identity/adapters/postgres` in the pinned Go container.
5. **Passed checks.** All backend vet and race tests pass. Testcontainers verifies real Argon2id login, exactly one dummy verification for unknown users, disabled-user rejection, same-account/cross-account rotation, digest-only storage, one-collision recovery, three-collision rollback, concurrent distinct sessions, exact expiry cleanup, idempotent revoke, cancellation rollback and recovery. The Course integration suite remains green.
6. **Remaining limitations.** No HTTP/cookie/origin contract exists yet. Demo-user seed values are deferred until I06, where documented credentials and repeat/concurrency evidence are delivered together. A connection loss during PostgreSQL commit remains an intentionally reported indeterminate 503 outcome.
7. **Next smallest task.** I03 adds typed runtime security configuration, the three HTTP handlers, same-origin enforcement, cookie handling and authentication middleware without adding a role-restricted business endpoint.
8. **User decision.** No product decision blocks I03. GitHub remote creation/push still requires explicit authorization for the exact external destination.

## I03 — HTTP transport, cookies and runtime security configuration

1. **Implemented.** Added the three Identity HTTP endpoints (`POST /api/v1/auth/login`, `GET /api/v1/auth/me`, `POST /api/v1/auth/logout`), same-origin enforcement, an authentication middleware, opaque session cookies and typed runtime security configuration read from the environment at startup.
2. **Files and modules.** `backend/internal/identity/adapters/http/handler.go` owns transport translation, strict JSON decoding and cookie writing. `config.go` owns `APP_ENV`, `PUBLIC_ORIGIN` and `COOKIE_SECURE` parsing with canonical origin comparison. `backend/cmd/api/main.go` wires the handler beside the Course handler and fails startup on invalid configuration. `docker-compose.yml` and `.env.example` carry the three variables.
3. **Architecture decisions.** Configuration is validated once at the composition root, so a production environment cannot start with insecure cookies or a non-HTTPS origin, and an insecure cookie requires a loopback origin. Login and logout require exactly one `Origin` header matching the canonical public origin; `null` is rejected. The login body accepts only `email` and `password`, rejects duplicate and unknown keys, trailing content and bodies over 8 KiB. Every Identity response sets `Cache-Control: no-store`. Session cookies are `HttpOnly`, `SameSite=Lax` and `Secure` per configuration. A duplicate session cookie is treated as a malformed request rather than silently choosing one value.
4. **Verification commands.** Full ordered gate sequence on the pinned toolchain: `make format-check`, `make lint`, `make typecheck`, `make test`, `make integration`, `make check-generated`, `make build`, `make up`, `make seed`, `make smoke`, `make e2e`. Live transport behavior was then exercised with `curl` against Nginx on `http://localhost:8080`.
5. **Passed checks.** All gates pass: 53 Rstest tests, Go `vet` and `-race` across all packages, Testcontainers integration including the Course suite, zero sqlc/Orval drift, four production images built, infrastructure smoke PASS and 14 Playwright tests. Live evidence: missing Origin and foreign Origin return 403 `origin_forbidden`; a non-JSON content type, an unknown body field and a duplicate key return 400 `invalid_request`; unknown and wrong-password logins return 401 `invalid_credentials` with no `Set-Cookie`; a correct login returns 200 with an `HttpOnly; SameSite=Lax; Max-Age=43200` cookie; `GET /me` returns the user with the cookie and 401 without it; presenting the previous cookie on re-login leaves exactly one session row and invalidates the old cookie; logout returns 204, clears the cookie, removes the row and a replayed cookie returns 401.
6. **Defect found and fixed during live verification.** The first correct-credentials login returned 503 `dependency_unavailable` while still committing a session row. `backend/cmd/api/main.go` injects `time.Now`, whose nanosecond precision cannot be represented by a PostgreSQL `timestamptz`; the value read back through `RETURNING` was truncated to microseconds, so the post-issuance invariant `session.CreatedAt.Equal(now)` in `application.Login` never held. Login was therefore broken for every real clock and leaked an orphan session row on each attempt. Every existing test used whole-second `time.Date(...)` values, which survive the round trip unchanged, so a fully green suite could not observe it. Fix: `application.Identity.timestamp` normalizes each clock reading to `storeResolution` (one microsecond) before it enters the round trip. Guards added: `TestLoginIssuesASessionWhenTheClockCarriesSubMicrosecondPrecision` at the application boundary, verified red against the previous code and green after the fix, and a PostgreSQL integration case pinning that session timestamps round-trip at microsecond resolution and no finer.
7. **Remaining limitations.** No OpenAPI operations, Orval client, Nginx auth edge rule, frontend session adapter, login screen or route guards exist yet; those are I04–I06. No role-restricted business endpoint consumes the authenticated user. The database still seeds no identities, so the live evidence above used a temporary local user that was removed afterwards; documented demo credentials remain I06 work. An indeterminate PostgreSQL commit is still reported as 503, and a session committed immediately before an unrelated invariant failure would still be orphaned.
8. **Next smallest task.** I04 publishes the three operations in `backend/openapi/openapi.yaml`, regenerates the Orval client without drift and adds the Compose/Nginx auth edge configuration.
9. **User decision.** No product decision blocks I04. The GitHub remote is now `git@github.com:b2vv/courses.git`; `main` and `feat/identity-foundation` are published.

## I04 — published contract, generated client and the authentication edge

1. **Implemented.** Published the three Identity operations in the OpenAPI contract, regenerated the Orval client from it, and gave Nginx an authentication edge: a dedicated rate-limit zone for login and, closing audit MEDIUM-1, a zone for the SSR page routes that previously had none.
2. **Files and modules.** `backend/openapi/openapi.yaml` gains `/api/v1/auth/login`, `/api/v1/auth/me` and `/api/v1/auth/logout`, the `LoginRequest` and `AuthenticatedUser` schemas, `NoStore` and `SessionCookie` response headers, the `InvalidRequest`, `Unauthenticated` and `OriginForbidden` responses, and four new error codes. `frontend/src/adapters/api/generated` gains `login`, `getCurrentUser`, `logout` and their models. `infrastructure/nginx/nginx.conf` adds the `login` and `web` zones. `infrastructure/tests/smoke.sh` asserts all three limits. `Makefile` reloads Nginx after `up` and `dev`.
3. **Architecture decisions.** Login gets an exact-match location ahead of the `/api/` prefix so it carries its own budget: `10r/m` with `burst=5` against `10r/s` with `burst=20` for the rest of the API. The page routes get `50r/s` with `burst=100`, high enough that a normal page load and the browser suite never trip it and low enough that the catalog stops being an unmetered path into PostgreSQL. The contract declares `Cache-Control: no-store` and the session cookie as required response headers, so a regression that stops setting either fails the contract test rather than only a hand-written assertion.
4. **Verification commands.** `make format-check`, `make lint`, `make typecheck`, `make test`, `make integration`, `make check-generated`, `make build`, `make up`, `make seed`, `make smoke`, `make e2e`, then live `curl` through Nginx.
5. **Passed checks.** All gates pass, including 14 Playwright tests with the new SSR limit active and zero generated drift after regeneration. `TestIdentityHTTPMatchesOpenAPI` validates twelve request and response pairs against the published document, covering success, origin refusal, malformed bodies, wrong credentials, missing and revoked sessions, idempotent logout, dependency failure and an unexpected error whose internal text never reaches the body. Live evidence through the new exact-match location: login 200, `me` 200, logout 204, replayed cookie 401. Rate limits measured from one source: 250 requests to `/` returned 102 successes and 148 refusals; 30 login attempts returned 6 through and 24 refused.
6. **Defect found during verification.** The Nginx configuration is a bind mount, so `make up` left an unchanged Nginx container serving the configuration it had loaded at start. A first measurement of the new zones therefore reported them as ineffective while the file on disk already contained them. `make up` and `make dev` now reload Nginx, and the zone measurements above were taken after a forced recreate.
7. **Remaining limitations.** No frontend consumes the generated auth client yet, and no route guard or login screen exists; that is I05 and I06. Two constraints must be honored there: a server-side call carries no `Origin` header, so login and logout are browser-only operations unless the SSR adapter sets one explicitly, and the generated client does not set `credentials`, which is correct for same-origin browser calls but means SSR must forward the session cookie itself, as the Course adapter already does. The database still seeds no identities.
8. **Next smallest task.** I05 adds the frontend session adapter and the localized login screen against the generated client.
9. **User decision.** No product decision blocks I05.

## I05 — frontend session slice and localized login

1. **Implemented.** Added the client-side identity slice end to end - domain, port, application, API adapter, composition and query boundary - and replaced the login placeholder with a working, localized login screen that restores a safe return target.
2. **Files and modules.** `frontend/src/domain/identity/identity.ts` owns the user model, the capability policy, credential normalization and return-target narrowing. `ports/SessionRepository.ts` names only the three operations the screen needs. `application/identity/session.ts` validates before touching the port. `adapters/api/session-repository.ts` maps transport to domain and `adapters/api/request-headers.ts` builds per-request headers for both API adapters. `app/composition/identity.ts` and `identity-queries.ts` wire request-scoped dependencies and the `['session']` query. `app/router/login-search.ts` narrows `returnTo`. `presentation/identity/LoginView.tsx` renders the screen and `routes/login.tsx` wires it.
3. **Architecture decisions.** An anonymous viewer is a success carrying `null`, not an error, so a public page never renders a failure state for the ordinary case; only transport faults reject. A syntactically impossible email is refused as `invalid_credentials` in the domain, so the client cannot become an account-enumeration oracle even before a request is sent. The identity query is never considered fresh, so every mount revalidates it; collection stays at the default because dropping the entry immediately would race SSR dehydration. `returnTo` is narrowed to an internal path in the domain and re-checked at use, and an unusable value produces no search key at all. Login and logout stay browser-only, as I04 recorded: a server-side call sends no `Origin` header and would be refused.
4. **Deviation from the plan.** The plan named React Hook Form for this screen. It is not implemented with one: the form has two fields and one submit, React Aria already supplies labelling, validation state and focus behavior, and the repository standard adds a dependency only when something needs it. No dependency was added for I05.
5. **Verification commands.** `make format-check`, `make lint`, `make typecheck`, `make test`, `make integration`, `make check-generated`, `make build`, `make up`, `make seed`, `make smoke`, `make e2e`, then live `curl` against SSR through Nginx.
6. **Passed checks.** All gates pass. The frontend suite grew from 53 to 112 tests across 14 files: domain and application boundaries, MSW transport mapping for every status the contract names, the shared header builder, the login URL contract, the query boundary, and the screen itself in both locales including keyboard-only submission and the password never rendering as readable text. Live SSR evidence through Nginx, repeated six times with identical results: a request carrying a valid session cookie renders the signed-in state and the display name inside the raw HTML, and the same URL without a cookie renders the login form. `Cache-Control: private, no-store` covers both.
7. **Defect found and fixed during verification.** A plain `/login` answered `307` to `/login?returnTo=%2Flearn`, because the search validator always produced a default and the router rewrote the address to match it. Every visitor paid a redirect and carried a meaningless parameter. The validator now emits no key when no usable target was supplied, and the route resolves the destination at use instead. `validateLoginSearch` is exported separately from the router-typed entry point so the contract is testable against untrusted input, matching the catalog pattern.
8. **Verification notes worth keeping.** MSW in this setup does not expose request headers to a handler, so cookie forwarding cannot be asserted there; it is asserted directly against the header builder and proven live end to end instead. Separately, `/healthz` is answered by Nginx itself, so waiting on it says nothing about whether the frontend is ready - an early measurement against a still-starting frontend produced two misleading results before this was noticed. Use `docker compose up -d --wait <service>`.
9. **Remaining limitations.** No header shows the signed-in user and no route guard exists yet, so `/learn` and `/admin` remain the Phase 1 locked placeholders; that is I06 together with logout in the header, demo fixtures and the Compose login/logout/role/expiry browser evidence. The database still seeds no identities, so the live evidence above used a temporary local user that was removed afterwards. Acceptance rows A01, A02, A04, A05 and B01 stay open until I06 supplies fixtures and Playwright coverage; A03 and A06 are covered here.
10. **Next smallest task.** I06 adds the authenticated header with logout, client route guards for `/learn` and `/admin`, demo identity fixtures and the ordered completion gate.
11. **User decision.** No product decision blocks I06.

## I06 — protected navigation, signed-in shell and completion gate

1. **Implemented.** Phase 3 now has a complete visitor journey: a visitor can sign in, return to
   the protected address they asked for, see their identity in the header, sign out, and be
   refused an area their role may not use. `/learn` and `/admin` show honest empty shells rather
   than the Phase 1 locked placeholder, and three explicit local fixtures make the journey
   repeatable without hand-written database rows.
2. **Files and modules.** The root route resolves and dehydrates request-scoped identity; the
   protected routes revalidate through one shared guard; the prop-driven header owns identity
   display and logout; the existing domain policy owns administration capability; and
   `backend/sql/seed-identities.sql` owns the three local fixtures. `Makefile` now makes `check`
   the manifest sequence: `format-check`, `lint`, `typecheck`, `test`, `integration`,
   `check-generated`, `build`, `up`, `seed`, `seed-identities`, `smoke`, `e2e`.
3. **Acceptance disposition.** I06 discharges phase rows A01, A02, A04, A05 and B01, plus the
   slice's C01, C02 and C03. The behavior and focused evidence were delivered across I06a-I06g;
   G38 and G41 are the final ordered-gate confirmation owned by I06h and remain explicitly
   awaiting execution in the table below.
4. **Ordered-gate evidence.** No result or count in this table is inferred from an earlier,
   partial or differently ordered run. The reviewing session must run it from a clean checkout
   and replace each awaiting marker with the observed result.

   | Step | Result |
   |---|---|
   | `make format-check` | pass |
   | `make lint` | pass — oxlint, frontend architecture check, gofmt, go vet |
   | `make typecheck` | pass |
   | `make test` | pass — frontend 175 tests / 19 files / 0 failures; all Go packages ok |
   | `make integration` | pass — Testcontainers PostgreSQL; identity and course suites green |
   | `make check-generated` | pass — no sqlc or Orval drift |
   | `make build` | pass |
   | `make up` | pass — all six containers healthy |
   | `make seed` | pass |
   | `make seed-identities` | pass — three fixtures reconciled |
   | `make smoke` (G38) | pass — "readiness, proxy, private media/metrics, worker status and API, login and SSR rate limits"; the seeded fixtures do not relax any threshold |
   | `make e2e` | pass — 20 tests, 16 chromium and 4 recovery |
   | Complete `make check` in manifest order (G41) | **pass, exit 0** — twelve steps, no manual user creation at any point |

5. **I06h task-local verification.** The requested command stopped at the denied Docker half of
   `make format-check`; the remaining targets were then invoked individually in the same order.
   These are partial task-local observations, not substitutes for the clean-checkout ordered gate
   above.

   | Check | Observed result |
   |---|---|
   | `make -n check` | pass — emits all twelve targets in manifest order, with `seed-identities` before `smoke` |
   | `make format-check` | blocked — frontend formatting passed on 84 files; Docker denied the Go formatting half |
   | `make lint` | blocked — frontend lint and the 48-file architecture check passed; Docker denied `go vet` |
   | `make typecheck` | pass — `tsc --noEmit` |
   | `make test` | blocked before collection — Rstest could not bind `::1:3000` (`EPERM`); the Go half was not reached |
   | `git diff --check` | pass |

   No target was edited to work around the sandbox. `integration` and every later gate were not
   run here.
6. **Known gate interaction.** `smoke` deliberately sends thirty concurrent login attempts.
   Nginx meters login at `10r/m` with `burst=5 nodelay`, so one source can leave its bucket at the
   ceiling and it drains at roughly one request every six seconds. The key is
   `$binary_remote_addr`, however: smoke reaches Nginx from the `frontend` container, while
   Playwright reaches `localhost:8080` from the host. Standard Compose networking presents those
   as different source addresses, so the prediction is that smoke's hot bucket does **not** carry
   into `e2e`. The ordered run must still settle that assumption; an environment that presents
   both paths under one address can produce a legitimate 429. If observed, the correction belongs
   in the ordering or the specs' sign-in count, never in the rate limit. Separately, `check` is no
   longer Compose-free, and `up` rebuilds after `build`, so the sequence knowingly builds twice;
   both consequences come from the manifest and CI order.
7. **Limitations.** Client guards and the header are navigation convenience, not an
   authorization boundary. Server-side role enforcement begins with the first protected Phase 4
   business endpoint. The `__Host-` session-cookie prefix and browser security headers remain
   deferred until TLS under S02 and S01 respectively. The administration and learning routes are
   shells only: there is no admin CRUD, enrollment, playback, upload, progress or instructor
   workspace in this phase.
8. **Planning evidence for the next slice.** The task cut did not survive execution. I06b had to
   build I06e's screens, so I06e shrank to two regression tests; I06g took the production `Vary`
   line from I06h so its own browser assertion could pass; and G37 moved from a Go integration
   test that could not see the repository root into `integration.sh`. Four of eight tasks moved
   work across their written boundaries. The next planner should cut around executable evidence
   and environment boundaries, not around file or UI labels that cannot land independently.
9. **Next phase.** Phase 4 specifies and builds server-authorized administration CRUD for courses,
   modules and lessons, including course ownership from its first migration. It must not infer
   authorization from these client guards. S01 and S02 remain blocked on TLS rather than becoming
   Phase 4 scope.
10. **User decision.** No new product decision was taken in I06h. The ordered gate and its measured
   values are deliberately left to the reviewing session with Docker, Compose and a browser.

## Catalog URL contract — two defects fixed after I05

Found while reading two reference codebases for Phase 4, both in Phase 2 code that had
been green since `4d73072`.

1. **`/courses` answered 307.** `validateCatalogSearch` always emitted all four
   parameters, so the router rebuilt a canonical URL that differed from the request and
   redirected to `/courses?q=&sort=recent&limit=20&offset=0`. Every catalog visit paid a
   redirect. This is the same defect fixed for `/login` in I05; the twin in existing code
   was not checked at the time. The rule is in `router-core` `load-server.js:530`: the
   router compares `next.publicHref` with the canonical location built with
   `_includeValidateSearch` and redirects on any difference.
2. **An unusable parameter answered 500.** `?limit=abc`, a repeated parameter, an
   out-of-range offset and an unknown sort all claimed a server fault for a malformed
   address, which also counts them as 5xx in metrics.

**What was measured before choosing a fix.** `z.catch()` cannot solve either problem: it
injects the key even when the parameter was absent, which makes the redirect
unconditional, and it cannot see the range rules at all because those live in
`makeCatalogQuery` rather than in the schema. A route lane answers any thrown value with
a hardcoded 500 (`load-server.js:303`) and offers no hook; `setResponseStatus(400)` from
`@tanstack/react-start/server` was tried in the loader and left the response at 200. A
400 is therefore not reachable for this case in this router version. `notFound()` is, and
404 is defensible for an address that does not identify a valid catalog state.

**Shape of the fix.** `validateCatalogSearch` echoes the four known parameters verbatim,
drops everything else, and neither defaults nor refuses — its only job is to leave the
URL alone. `catalogQuery` applies defaults once and refuses with `notFound()`. The route
calls it in `beforeLoad`, inside the lane, because a `notFound()` thrown from
`validateSearch` is not classified as a not-found and still produces 500.

**Verified live.** `/courses`, `?q=Go`, `?offset=20` and `?sort=title&limit=10` all
answer 200 with zero redirects. `?limit=abc`, `?limit=1&limit=2`, `?offset=-1`,
`?limit=51`, `?sort=popular`, `?q=%00` and `?offset=10001` answer 404 with the
not-found screen.

**Behaviour change to note.** An unknown parameter such as `?role=admin` is now dropped
and the catalog renders, where it was previously refused. The backend contract still
rejects unknown query parameters, so the API surface is unchanged; only the page URL is
now tolerant, which keeps shared links carrying tracking parameters working.

## Security review finding — logout failed open on a duplicated session cookie

An adversarial review of the branch found no HIGH or MEDIUM issues, and one
sub-threshold divergence between the specification and the code, fixed here.

`handler.go` treated a duplicated `courseforge_session` cookie the same way in all
three handlers, but the consequence differed. Login and authentication refuse the
request without touching anything, which is fail-closed. Logout expired the browser
cookie, answered `204` and returned **without revoking**, leaving the presented session
valid on the server for up to twelve hours while telling the browser it had signed out.
The specification claims duplicate session cookies "fail closed before mutation"; here
they failed open, and the response could not be distinguished from a real logout.

Exploitation needs an attacker who can already set a second cookie of that name, which
nothing on this branch enables, so severity is low. The divergence is the defect: a
document that promises a guarantee the code does not keep is worse than an absent
guarantee.

**Fix.** `sessionCookies` returns every session cookie the request carried and each
handler states its own policy. Login and authentication keep refusing more than one.
Logout revokes all of them, because refusing would leave the session live and that is
the outcome the endpoint exists to prevent. A failure to revoke any one of them still
reports `503` rather than claiming success. The contract is unchanged: logout still
answers `204` and remains idempotent, so no OpenAPI or client regeneration was needed.

**Guards.** Two tests: one asserts both presented tokens reach revocation and the
browser cookie is cleared, the other asserts that a store failure during a multi-token
logout reports `503` instead of `204`.

**Not affected: multiple devices.** Rotation deletes only the token the request
presented (`repository.go:86`, `DeleteIdentitySession` matches a single digest), never
every session of the user. Verified live with two independent cookie jars: both sign in
and both read `/auth/me`; a logout on one leaves the other authenticated and one row in
`user_sessions`; re-login on either leaves the other untouched at two rows. The fix
revokes every cookie *in one request*, which is one browser, not every device.

**The risk this task was told to surface did not materialize.** `smoke` deliberately spends thirty
login attempts proving the limiter answers `429`, and since I06g several browser specs sign in
through the real endpoint immediately afterwards. Nobody had run those two steps back to back
before. They pass: the limiter's window recovers between them, and `e2e` authenticates normally.
Recorded because the failure, had it happened, would have looked exactly like a product defect and
is not one — `work/report-i06g.md` describes the same `429` arriving from a different cause when
the suite is run twice by hand.
