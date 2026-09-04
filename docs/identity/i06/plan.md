# I06 — plan

Third edition. Two earlier ones were refused at GATE 2; [../_wip-i06.md](../_wip-i06.md)
holds the verdicts and the router spike whose measurements this design rests on. Placement
follows `.claude/standards.md` §Розміщення.

## Measured facts this design rests on

Run against `@tanstack/react-router` 1.170.32 in this repository, not reasoned about:

| Fact | Consequence |
|---|---|
| A child's `beforeLoad` receives the root's `beforeLoad` context, root first | Identity resolves in the root's `beforeLoad` and the guard reads it from context |
| The root's `beforeLoad` re-runs on every client navigation; its **loader does not** | Identity must live in `beforeLoad`. In the loader it would freeze at document load and bounce a just-signed-in visitor back to `/login` |
| `redirect()` from a child `beforeLoad` lands on the target with `search` preserved | The `returnTo` round trip works from the guard |
| `notFound()` from a child `beforeLoad` sets `status="notFound"` | The 404 answer works from the guard |
| `beforeLoad` is skipped on the server for `ssr: false` matches (`load-server.js:159`) | `/learn` and `/admin` decide client-side. The server still resolves identity in the root and dehydrates it, so nothing waits on a fetch |

## Reuse

**Reused unchanged:** `identity-queries.current()` and `sessionQueryKey`; `session.logOut`
(written in I05, still without a consumer); `canAccessAdministration`; `EmptyState`,
`Button` and `Link` from `presentation/ui`; the shared API transport; `notFound()`, already
used by `app/router/catalog-search.ts`; `createIsomorphicFn` request options from
`app/composition/identity.ts`.

**New to this slice:** `redirect()` — the second edition claimed it as reused; `grep` finds
no router `redirect()` anywhere in `frontend/src`.

**Changed:** `safeReturnTo`; `__root.tsx` gains a `beforeLoad` and its `Layout` the
signed-in header; two layout routes gain guards; `Makefile` `seed`; `frontend/server.mjs`
gains `Vary`; `.github/workflows/ci.yml` sets `SEED_IDENTITIES=1`; `Makefile` `check`.

**Deleted:** `presentation/LockedArea.tsx` and the `locked.*` resources — **four** call
sites: `routes/{learn,learn.$,admin,admin.$}.tsx`.

## Design

### Identity

`__root.tsx` gains a `beforeLoad` that returns `{ identity }` into context, where

`identity: { status: 'anonymous' } | { status: 'signed-in', user } | { status: 'unavailable' }`

It never throws: `__root.tsx:35` carries `errorComponent`, so a thrown identity fault would
replace every public page with a failure screen. It resolves through
`context.queryClient.ensureQueryData(context.identity.current())` so the Query cache is the
single store — the header, the guard and `login.tsx` all read `['session']` rather than
three different values.

It calls the API only when the request carries the **session** cookie by name. Checking for
a `Cookie` header would be useless: `LanguageSwitcher.tsx:8` writes `courseforge_locale`
with a year's `Max-Age`, so any visitor who once switched language carries one forever. On
the client there is nothing to inspect — the cookie is `HttpOnly`
(`identity/adapters/http/handler.go:130`) — so the client always asks.

`login.tsx:12` already resolves identity in its own loader. It drops that loader and reads
the root's value, or the route resolves identity twice per SSR render.

### Guard

`/learn` and `/admin` each become a **layout route** carrying the guard, the
`errorComponent` and the `pendingComponent`, with the existing leaf and splat routes as
children. Two files instead of four copies of the same three lines, and G06's nested
addresses then hold by construction.

| Context identity | Guard |
|---|---|
| `signed-in`, capability allowed | returns |
| `signed-in`, capability denied | `throw notFound()` |
| `anonymous` | `throw redirect({ to: '/login', search: { returnTo } })` |
| `unavailable` | throws a typed failure into the layout's own `errorComponent`, which renders the localized retryable state `docs/identity/spec.md:50` requires |

Two named guards — `requireSignedIn` and `requireAdministration` — not one taking a
predicate. Two call sites do not justify a parameter.

`returnTo` is the current path **including its search string**; a deep link to
`/learn/go/intro?lesson=3` must survive, which the second edition's design silently dropped.

### `safeReturnTo`

`docs/identity/spec.md:52` requires only normalized internal protected paths. The shipped
version accepts `/login`, `/courses`, `%0d%0a` forms and dot segments. Replaced by:
an allowlist `^/(learn|admin)(/[A-Za-z0-9._~-]+)*$` on the **path**, an explicit rejection of
any `.` or `..` segment, and a separately validated optional search string. No `%` can occur
in the path, so encoded slashes and encoded CRLF cannot.

### Header

`Layout` reads `['session']` through `useQuery` and passes props to `SessionMenu`, which
stays prop-driven like `LoginView`. Name as text, a sign-out `Button`, and a `Link` to
`/admin` only when `canAccessAdministration`. Sign-out calls the reused use case, sets the
query to `null`, and navigates to the catalog.

Accessibility, which the first edition omitted entirely: the sign-out control is a real
button in tab order with the kit's visible focus ring and a ≥44px target; the administration
link sits in the existing `<nav>`; the name is text; both locales carry every string; no
colour is hardcoded.

### Screens

`/learn` and `/admin` compose `EmptyState` in their route files, as `courses.index.tsx`
composes `CatalogView`.

### Fixtures

Three identities appended to `backend/sql/seed.sql`.

`ON CONFLICT (email) DO UPDATE` sets role, display name and hash from the file, so editing a
fixture password takes effect. Fixture **ids are fixed constants**, so a primary-key
collision cannot arise from a differing email; `ON_ERROR_STOP=1` would otherwise abort the
whole seed including courses.

The identity section is gated on a psql variable, not on a shell variable inside the
container: `docker compose exec` does not forward host environment, so the second edition's
`SEED_IDENTITIES=1 make seed` would have reached nothing. The `seed` recipe passes
`-v identities="$$SEED_IDENTITIES"` and the SQL branches with `\if`. CI sets the variable
explicitly.

`APP_ENV` cannot carry this gate: the `Makefile` never loads `.env`,
`docker-compose.yml:61` defaults it to `development`, and `.github/workflows/ci.yml:31` runs
`make seed` with it unset. The slice specification's decision 6 is amended accordingly — the
user's intent, keeping fixtures out of a shared environment, is unchanged; only the
mechanism is.

### `Vary` and `check`

`server.mjs` **appends** `Vary: Cookie` rather than setting it, so a later locale variation
is not clobbered. `docs/identity/spec.md:41` also expects variation on the locale inputs,
which today is undeclared; declaring it is out of this slice and recorded as such.

`check` becomes the manifest's ordered sequence — `format-check lint typecheck test
integration check-generated build up seed smoke e2e` — so C03 is one command rather than a
prose list.

## Сайд-ефекти

| Effect | Assessment |
|---|---|
| One `/auth/me` per SSR render for a viewer carrying the session cookie | Internal call. This slice measures it and records the number; no repository-recorded figure exists, and the first edition cited one that had never been written down |
| One identity revalidation **per client navigation**, not per document load | Measured: root `beforeLoad` re-runs on every navigation and `staleTime: 0` forbids a stale answer. Query deduplicates concurrent reads |
| A hung API stalls SSR for the transport deadline before `unavailable` is returned | `transport.ts:26` is 5000 ms, which is too long to block a storefront render. The root resolve passes a shorter deadline; the number is set in the task and recorded |
| SSR HTML becomes viewer-specific | `private, no-store` already; `Vary: Cookie` appended |
| Seeded credentials in the repository | Local-only, behind the psql gate |

## План міграції

`make seed` applies one file whose identity section is gated. No schema change. Re-running
reconciles fixtures to the file.

## Rollback

The slice lands as one squashed commit — `LockedArea` is deleted in it and the routes point
at replacements, so a partial revert dangles imports. `AGENTS.md:9` asks for small logical
commits; this slice is one logical change.

| Change | Undo | Lost |
|---|---|---|
| Routes, header, root `beforeLoad`, `safeReturnTo`, `Vary`, `check`, `ci.yml` | `git revert` of the squashed commit | nothing |
| Seeded rows | `DELETE FROM users WHERE email IN (…the three fixture addresses…)`; sessions cascade (`migrations/00002_identity.sql:29`) | the fixtures |
| Contract amendments (`b5a0cd7` and decision 6) | **Not reverted by the above** — they are separate commits | Reverting the code leaves a contract describing a 404 answer and a reconciling seed that no longer exist. Either revert those commits too, or accept the contract running ahead of the code and say so |

## Конституція

Domain stays import-free; `safeReturnTo` remains so. Application untouched. Policy inward:
the guards call a domain predicate. No `any`; identity is a discriminated union. No `fetch`
in presentation or routes — the root `beforeLoad` calls the composition layer, and request
options come from `app/composition/identity.ts` rather than being recomputed in a route file
(`AGENTS.md:15`). No dependency added. Accessibility addressed above.

## Оцінка

`docs/identity/plan.md:3` caps a task at 2h, so a realistic I06 is not one row. Sized
honestly at **≈12h across eight tasks**, cut in the task tracker rather than hidden in one
number. The dominant cost is that `frontend/src` has **no route-level test at all**, no SSR
assertion and no seed integration test, and the acceptance table needs all three. Playwright's
`request` API covers the raw-HTML and parallel-SSR rows without a new harness, which is why
the figure is twelve and not twenty.

## Приймальні сценарії

[acceptance.md](acceptance.md), amended for this edition: the no-flash row lost in the second
rebuild is restored, G07/G08 corrected for `ssr: false`, G10 made writable, and rows added
for a deep link carrying a search string and for the identity revalidation cadence.
