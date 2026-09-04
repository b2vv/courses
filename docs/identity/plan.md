# Identity Phase 3 plan

Status: revision under independent critique; implementation starts only after plan-critique and plan-defense are green. Each task is 0.5–2 hours and produces its own verification evidence.

## Small tasks and dependencies

| ID | Task | Estimate | Depends on | Result | Minimum gate |
|---|---|---:|---|---|---|
| I01 | Users/sessions migration and Identity domain | 1.5h | green plan gates | schema plus User/Role/Session invariants and role policy | domain/migration tests |
| I02 | Credential and session application slice | 2h | I01 | ports, login/authenticate/logout use cases, Argon2id and PostgreSQL adapters | table tests, dummy-verification call path, Testcontainers rotation/expiry |
| I03 | HTTP auth contract and runtime configuration | 2h | I02 | handlers, same-origin check, cookies, authentication middleware, typed startup config | HTTP/config matrix, role-policy tests, no-secret logging |
| I04 | Contract generation and edge infrastructure | 1h | I03 | OpenAPI/Orval, Compose/env wiring, login-specific Nginx limit | generated drift; Nginx threshold smoke |
| I05 | Frontend session slice and login | 2h | I04 | domain/application/port/adapter, reuse of request-scoped Query/API forwarding, RHF login, localized uk/en states | MSW, component, SSR identity/isolation tests |
| I06 | Client route guards and completion gate | 4h | I05 | client-only `/learn` authenticated; `/admin` instructor/admin; header/logout; demo fixtures | ordered CI, Compose login/logout/role/expiry E2E |

## Acceptance scenarios

| ID | Scenario | Preconditions | Action | Observable result | Evidence |
|---|---|---|---|---|---|
| A01 | Student login | active student fixture | submit correct credentials | 200, HttpOnly cookie, `/learn` opens | HTTP + Playwright |
| A02 | Admin login | active admin fixture | submit correct credentials | admin identity and `/admin` opens | HTTP + Playwright |
| A03 | Invalid credentials | wrong/unknown/syntax-invalid email | submit login | same localized UI and `invalid_credentials`; no enumeration | application/HTTP tests |
| A04 | Current user SSR | valid session cookie | reload an SSR catalog URL | signed-in header exists in raw HTML and hydrates without identity flash | HTTP/raw HTML + Playwright |
| A05 | Logout | valid or absent session | submit logout | 204, cookie expired, protected route redirects to login | HTTP + Playwright |
| A06 | Return URL | anonymous user opens protected route | authenticate | safe internal `returnTo` restored | route/component tests |
| B01 | Roles | student/instructor/admin | open client-only `/admin` | student receives the not-found answer; instructor/admin allowed shell; domain policy expresses the same capability matrix | policy + Playwright |
| B02 | Expiry boundary | rows before/equal/after one injected `now` | call `/me` | only `expires_at > now` authenticates; stale matching row is removed; cleanup failure is 503 | integration/HTTP tests |
| B03 | Concurrent login | same user logs in concurrently | create sessions | every successful response has a distinct valid token and committed row | integration test |
| B04 | Cookie flags | successful login/logout | inspect Set-Cookie | HttpOnly, SameSite=Lax, Path=/, bounded Max-Age; deletion has same scope and Max-Age=-1/past Expires; Secure by config | HTTP table test |
| B05 | Unicode/display | valid bounded profile | call `/me` | JSON and localized UI preserve text | adapter/component tests |
| B06 | Atomic fixation rotation | browser has same-account, cross-account or attacker-chosen cookie | successful login | after a confirmed commit a fresh 43-char token works and any valid presented token is revoked; chosen invalid token is never stored; failed insertion leaves the prior session intact; indeterminate commit returns 503 and sends no cookie | application/integration tests |
| B07 | At-rest token | session created | inspect ephemeral DB | only a unique 32-byte SHA-256 digest exists; raw token is absent | integration test |
| B08 | Digest collision | token generator repeats an existing digest | create session | each collision rolls back the full rotation and retries in a fresh transaction at most 3 times; exhaustion leaves the prior session intact | adapter/integration tests |
| B09 | Seed repeatability | empty, already-correct or conflicting fixture state | run local seed once, repeatedly and concurrently | one atomic exact student/instructor/admin fixture set, the seed file being the source of truth so a changed fixture is applied rather than ignored; correct repeats succeed; conflicting rows fail loudly without partial fixtures | seed integration test |
| E01 | Malformed body | body over 8 KiB; broken/missing/wrong-type/unknown/duplicate fields; email over 254 bytes or password over 1024 bytes | login | bounded `400 invalid_request`; syntax-invalid but bounded email remains `401 invalid_credentials`; no repository call for transport failures | HTTP tests |
| E02 | CSRF/login CSRF | foreign/missing/multiple/null/malformed origin, including default-port boundaries | login/logout | canonical scheme/host/effective-port mismatch returns `403 origin_forbidden`, with no session mutation | HTTP/config tests |
| E03 | Dependency failure | valid-format cookie or login plus PostgreSQL unavailable | login or `/me` | safe `503`; missing/malformed cookie remains 401 without DB lookup; no SQL/secret in response | fake port + recovery test |
| E04 | Tampered token | random/malformed or duplicate session cookie | `/me` | `401 authentication_required` without repository call | HTTP tests |
| E05 | Session isolation | concurrent SSR users A/B | load SSR routes | no cookie, Query cache or user data crosses requests; raw HTML varies on Cookie | frontend integration test |
| E06 | Rate limit | comparable login and ordinary API bursts | repeated requests | login reaches 429 at a lower threshold; neither path logs query/credentials | infrastructure smoke |
| E07 | Open redirect | external/protocol-relative/encoded/backslash/control/login-loop returnTo | login succeeds | falls back to `/learn`; normalized protected paths survive | frontend application tests |
| E08 | Cancellation | cancel an in-flight session mutation | application returns context cancellation and transaction leaves no partial rotation | integration test |
| E09 | Logout store failure | valid cookie, PostgreSQL unavailable | logout | cookie expires, safe 503; frontend clears private Query identity and shows localized uncertain-revocation error | HTTP/component/recovery test |
| E10 | Cache/history and auth availability | authenticated identity, `/me` 401 or 503, then logout/back | navigate or reload | 401 redirects, 503 renders retry without redirect, every expired-cookie logout clears private Query data, and history never shows stale user UI | HTTP/component/Playwright |
| E11 | Startup configuration | production/development origin-cookie matrix | start API | loopback insecure dev and HTTPS secure production start; insecure non-loopback, HTTP production, origin mismatch and invalid values fail before listening | config table tests + Compose smoke |
| E12 | Concurrent logout | two requests share one valid token | submit simultaneously | both return 204 with expired cookies and no session remains | integration/HTTP test |

## Decisions

- PostgreSQL is the session source of truth for this slice and avoids a second consistency model. Redis remains available for later distributed attempt limiting/jobs; Nginx provides the current external login throttle. Phase 3 has no login-session cap; the separate concurrent-playback limit belongs to Phase 6.
- Opaque server-side sessions are simpler to revoke than JWTs and keep permissions current. The browser receives only a random token; its digest is indexed in PostgreSQL.
- Login rotation is a single transaction. Unique-digest collision aborts PostgreSQL's transaction, so each bounded retry starts a fresh transaction and repeats revocation plus insertion; the HTTP layer emits a cookie only after the repository reports a confirmed commit. A connection failure during commit is an indeterminate database outcome, so the adapter returns dependency failure without claiming that the old session survived.
- A 12-hour fixed session avoids a speculative refresh-token protocol. Rolling renewal/device UI can be added with a real requirement.
- Demo users belong to the explicit local seed and carry clearly public fixture credentials. Production setup must use a separate bootstrap path before deployment.
- Mutation Origin validation uses configured `PUBLIC_ORIGIN`. A typed composition-root config parses `APP_ENV`, `PUBLIC_ORIGIN` and `COOKIE_SECURE` before opening the listener. `APP_ENV=production` requires one canonical HTTPS origin and `COOKIE_SECURE=true`; any `COOKIE_SECURE=false` configuration is rejected unless the origin hostname is `localhost`, `127.0.0.1` or `::1`. `.env.example` and Compose own the development values. Tests cover valid HTTPS, loopback development, non-loopback insecure origin, mismatched/missing request origin and invalid startup combinations.
- No role-restricted business API exists in Phase 3. `/auth/me` exercises authentication middleware and the domain role policy receives unit coverage. A handler-level authorization wrapper is added when Phase 4 introduces the first admin handler. The Done report must not claim protected admin data before then.
- Identity Query options reuse the foundation's request-scoped `QueryClient`, abort signal, cookie forwarding and API-origin selection. If sharing requires extraction, it is a small common adapter helper rather than a second Identity-specific SSR/cache mechanism.
- The local identity seed is one transaction. It inserts the three documented fixtures and reconciles an existing row to the file's values, so editing a fixture password takes effect on the next run instead of silently keeping the old credential. It is verified for first, repeated and concurrent execution.

## Rollback

Stop new binaries, revoke every server session in one audited `DELETE FROM user_sessions` transaction, then restore previous images/config and retain user data. The old foundation ignores the now-inert cookies. A later Phase 3 rollout cannot reactivate them because their rows were deleted. Prefer a forward migration for user/schema corrections. The down migration is exercised only in ephemeral Testcontainers databases. Removing Phase 3 frontend guards returns the foundation fail-closed placeholders; it never grants API access.
