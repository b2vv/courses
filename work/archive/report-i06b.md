# Report — I06b

Client route guards and protected-area shells.

## What was implemented

- Added one shared `requireIdentity(context, location)` guard. It revalidates the existing
  session Query with `fetchQuery`, returns the authenticated user, redirects anonymous viewers
  to `/login` with the complete path and search string, and converts dependency failures into a
  retryable `SessionQueryError`.
- Added the guard to `/learn` and `/admin`; the administration route applies the existing
  `canAccessAdministration` policy inline and throws `notFound()` for students.
- Added shared localized pending and retryable error states, plus basic localized `EmptyState`
  shells for both protected roots. The header, logout and administration link remain I06d.
- Added a mounted-session subscription with `refetchOnMount: false`. Window-focus revalidation
  that changes the session to `null` redirects with the protected address captured before the
  redirect starts.
- Removed `LockedArea`, all four route call sites, and the `locked.*` translation resources.
  Splat routes remain client-only and inherit their parent guard and screen.
- Strengthened the existing request-header test so two session cookies are demonstrably forwarded
  verbatim for the backend to reject.

## Acceptance rows covered

| Row | Evidence |
|---|---|
| G01 | `G01/G02/G13 redirects an anonymous visit to /learn` |
| G02 | `G01/G02/G13 redirects an anonymous visit to /admin` |
| G04 | `G04 gives a student the same not-found answer for /admin as an unknown address`; existing domain policy test |
| G05 | `G05 opens /admin for the instructor role` and `…admin role` |
| G06 | `G06 redirects anonymous nested address …` including preserved search; `G06 applies the parent role matrix to signed-in nested addresses` |
| G08 | `G08 shows the route pending state while protected navigation revalidates` |
| G09 | `G09 renders a localized retryable route error when identity is unavailable`, including successful route invalidation after recovery |
| G11 | `G11 revalidates a cached identity and redirects when the session has ended` |
| G12 | `G12 leaves a mounted protected route when focus revalidation ends the session` |
| G13 | The G01/G02 route matrix covers the frontend `401` result; existing `TestAuthenticateAndLogoutRejectMalformedTokensWithoutStoreLookup` covers rejection before storage |
| G14 | `forwards the request cookie verbatim when one was given` now uses two session cookies; existing `TestDuplicateSessionCookieFailsClosed` covers backend rejection |
| G39 | `rg -n "LockedArea|locked\\." frontend/src --glob '!routeTree.gen.ts'` returned no matches |

The route suite has 12 tests in `frontend/src/routes/route-guards.test.tsx` and uses the real
generated route tree, memory history, the existing API session adapter, Query client and MSW.

## Verification

Run in the required order from `/Users/strelia/projects/courses`:

| Step | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass (oxlint, architecture check, gofmt, go vet) |
| `make typecheck` | pass |
| `make test` | pass — frontend 155 tests / 18 files / 0 failures; all Go packages ok |
| `make integration` | pass — Testcontainers Postgres, identity and course suites green |

The implementer could not run the Docker-backed gates or `make test` unmodified: its sandbox
denied the Docker socket and a localhost port probe. Every row above was therefore re-run
unsandboxed from the reviewing session, and the counts match what the implementer reported.

Supplemental checks:

- `npm run format:check`, `npm run lint`, and `npm run typecheck`: pass.
- Full `npm test` with a temporary process-only shim for Rstest's localhost port probe: pass,
  155 tests / 18 files / 0 failures. The shim did not alter application or test behavior and was
  removed immediately after the run.
- `git diff --check`: pass.

`make e2e` was not run: the brief assigns Compose browser evidence to I06g after I06f fixtures.
`make build` was not requested for I06b and was not run.

## Review

The required two-axis review found no production-code standards issue. Its one standards
suggestion was to omit the duplicate-cookie test change; it remains because G14 is explicitly
owned by I06b and the concrete duplicate is the observable frontend half of that row. The spec
review found that G09 did not initially prove recovery; the test now restores the API response,
invalidates the route and observes the administration shell.

## Carried forward to I06d

`useRedirectWhenSessionEnds` sets `refetchOnMount: false` on its `['session']` observer. **I06d's
header observer must set it too.** Without it, the dehydrated value is instantly stale under
`staleTime: 0` and `shouldFetchOnMount` fires one `/auth/me` per document load, which is exactly
what G07 asserts must not happen. The tracker row for I06d now says so.

The pending and failure states reuse the `login.*` translation keys rather than introducing new
ones. Correct as reuse, slightly off as naming; I06e may revisit the copy.

## Deviations and limitations

- No behavior deviates from the fourth-edition plan. `safeReturnTo`, the signed-in header, logout,
  fixtures and browser tests were deliberately left to I06c, I06d, I06f and I06g.
- Client guards are navigation convenience, not an authorization boundary; server-side role
  enforcement remains Phase 4.
- A focused commit could not be created in this environment: Git cannot create `.git/index.lock`
  because the workspace exposes `.git` read-only. The supplied untracked `work/brief-i06b.md` was
  preserved unchanged, so the tree is not clean.

## Next task

After committing I06b in a Git-writable environment and rerunning the unmodified Docker-backed
gates, I06c can narrow `safeReturnTo` as planned.
