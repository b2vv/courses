# Report — I06a

Root identity resolution. Implemented by Codex (`gpt-5.6-sol`, session `01a06f06`); reviewed,
verified and committed from the Claude session that planned the slice.

## What was implemented

`resolveRootIdentity` resolves the viewer's identity once per navigation and returns it as
router context.

| File | Change |
|---|---|
| `frontend/src/app/composition/root-identity.ts` | **new** — the resolver and the `RootIdentity` discriminated union |
| `frontend/src/app/composition/identity.ts` | cookie presence via `getCookie`, `timeoutMs` on the server branch only, `shouldLoadCurrentUser` exposed |
| `frontend/src/routes/__root.tsx` | `beforeLoad: resolveRootIdentity`; context type composed from `RootIdentityContext` |
| `frontend/src/router.tsx` | context key renamed (see "Deviations") |
| `frontend/src/routes/login.tsx` | drops its own identity loader and `useSuspenseQuery`; reads the root's value |
| `frontend/rstest.config.ts` | Rspack alias so tests can import the server barrel (see "Deliberate simplification") |
| `frontend/src/app/composition/root-identity.test.tsx` | **new** — five tests |

The plan's four load-bearing properties all hold, and each was checked against the code rather
than taken from the report:

- Identity resolves through `ensureQueryData`, **not** `fetchQuery`. Revalidation belongs to the
  guards in I06b; the plan explains why, and the implementation did not "fix" it.
- The server calls the API only when the request carries the session cookie, checked with
  `getCookie('courseforge_session')` inside the `.server()` branch. No hand-rolled cookie parser.
- `timeoutMs: 1000` is set on the `.server()` branch only, so `logIn` and `logOut` — which only
  run from the browser — keep the transport default of 5000.
- The no-cookie branch writes `setQueryData(sessionQueryKey, null)`, so `/login` no longer meets
  an empty cache during SSR.
- The resolver never throws: every failure is caught and mapped to `unavailable`.

## Acceptance rows covered

I06a owns G10, G15, G22, G23. All four are covered by `root-identity.test.tsx`, five tests:

| Row | Test |
|---|---|
| G10 | `keeps the root route successful when identity returns 503` |
| G10 | `limits a stalled SSR identity read to the one-second server deadline` |
| G15 | `renders the public route anonymously when the API returns an unknown role` |
| G22 | `makes no identity request without the session cookie and caches anonymous` |
| G23 | `keeps concurrent SSR identities isolated by request and response` |

The tests drive the **real** path: they build the actual router through `getRouter()` and
establish a server request context with `requestHandler`, so `getCookie` runs against a real
`Request` carrying real cookies. G22 therefore proves the cookie check itself, not a mock of it —
which is what makes it able to catch a hand-rolled parser regression.

## Deviations from the plan, and why

**The resolver lives in `app/composition/root-identity.ts`, not inline in `__root.tsx`.** The
plan said "`__root.tsx` gains a `beforeLoad`", which it does — it references the function.
Putting the function in composition matches `.claude/standards.md` §Розміщення and is what makes
it testable without standing up a route. Accepted.

**The router context key `identity` is renamed `identityDependencies`.** Not foreseen by the
plan and forced by it: `beforeLoad` returns `{ identity }`, which merges into context and would
have collided with the existing dependency bundle under the same name. **I06b, I06d and I06e read
this context** — they must use `identityDependencies` for the bundle and `identity` for the
resolved value.

## Deliberate simplification

`frontend/rstest.config.ts` aliases `@tanstack/react-start/server$` to
`@tanstack/start-server-core/request-response` for **all** tests, because the public barrel pulls
virtual production entries Rstest cannot resolve. It is load-bearing: without it the test cannot
import `requestHandler`, and without `requestHandler` it cannot exercise the real cookie path.

Ceiling and trigger are marked in the file with a `ponytail:` comment: a test wanting some other
export from that barrel will silently get a module without it. Scope the alias per test project
when a second test needs one.

## Verification

Run from this session, not taken from the implementer's word.

| Gate | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass (oxlint, frontend architecture check, gofmt, go vet) |
| `make typecheck` | pass |
| `make test` | pass — frontend 143 tests / 17 files / 0 failures; all Go packages ok |
| `make integration` | pass — Testcontainers Postgres, identity and course repository suites green |

Frontend test count rose from 112 after I05 to 143. Docker was available, so no gate was skipped.

## Limitations

- Client guards are navigation convenience, never the security boundary. Nothing in I06a is
  evidence of authorization; server-side role enforcement remains Phase 4.
- `shouldLoadCurrentUser` is unconditionally `true` on the client, by design: the session cookie
  is `HttpOnly`, so the browser cannot rule itself out and must ask.
- No revalidation happens on ordinary client navigation. That is deliberate and is the whole
  point of the freshness decision in the plan; the guards in I06b buy freshness at the protected
  boundary.

## Next task

**I06b** — guards on `learn.tsx` and `admin.tsx`, one `requireIdentity` using `fetchQuery`, the
role check inline via `notFound()`, `errorComponent` on `ErrorState`, the `['session']`
subscription for G12, and the deletion of `LockedArea`. It owns twelve acceptance rows and
depends on this task.
