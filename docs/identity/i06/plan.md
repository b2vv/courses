# I06 — plan

Second draft. The first was refused at GATE 2 — 11 blocking findings, 5 defence holes out of
5, and one constitution violation; [../_wip-i06.md](../_wip-i06.md) records what was wrong.
Placement follows `.claude/standards.md` §Розміщення.

## Reuse

**Reused unchanged:** `identity-queries.current()` and `sessionQueryKey`; `session.logOut`
(written in I05, still without a consumer); `canAccessAdministration`; `EmptyState`,
`Button` and `Link` from `presentation/ui`; the shared API transport; `notFound()` and
`redirect()` as `app/router/catalog-search.ts` already uses them.

**Changed:** `safeReturnTo` narrowed; `__root.tsx` gains a loader and its `Layout` the
signed-in header; four route files get a guard; `Makefile` `seed` gains a confirmation gate;
`frontend/server.mjs` adds `Vary: Cookie`.

**Deleted:** `presentation/LockedArea.tsx` and the `locked.*` resources. It has **four**
call sites — `routes/{learn,learn.$,admin,admin.$}.tsx` — all replaced here. The first draft
said three and never mentioned the splat routes, which G25 depends on.

**Added:** `app/router/guards.ts`; `presentation/identity/SessionMenu.tsx`; an identity
section in `backend/sql/seed.sql`.

## Design

### Identity at the root — server-side only, never throwing

`__root.tsx` gains a loader returning a discriminated value:

`{ status: 'anonymous' } | { status: 'signed-in', user } | { status: 'unavailable' }`

Throwing was the first draft's worst error. `__root.tsx:35` carries `errorComponent`, so an
identity outage would have replaced every public page — `/`, `/courses`, `/courses/$slug` —
with a failure screen. Catching at the root keeps a `/auth/me` fault out of the storefront.

**Decision 1 applies to SSR only.** The cookie is `HttpOnly`
(`identity/adapters/http/handler.go:130`), so `document.cookie` cannot see it and the first
draft's "isomorphic" check was unimplementable. On the server the loader reads the request
header and skips the call when absent. In the browser there is nothing to inspect, so it
asks.

### Guard

`app/router/guards.ts`, called from `beforeLoad` in all four protected routes, which keep
`ssr: false`:

| Root loader said | Guard does |
|---|---|
| `signed-in`, capability allowed | returns; the route renders |
| `signed-in`, capability denied | `throw notFound()` |
| `anonymous` | `throw redirect({ to: '/login', search: { returnTo } })` |
| `unavailable` | throws a typed failure caught by the **route's own** `errorComponent`, rendering the localized retryable state that `docs/identity/spec.md:50` requires |

The capability is a predicate the caller passes, so the guard knows nothing about roles:
`/learn` passes "is signed in", `/admin` passes `canAccessAdministration`. There is no
"pending" outcome — `beforeLoad` runs after the loader resolves, so that state is
unreachable. The first draft invented it and cited a "decision 7" that does not exist.

### `safeReturnTo`, narrowed

`docs/identity/spec.md:52` requires only **normalized internal protected** paths. The
shipped version accepts `/login`, `/courses`, percent-encoded control characters and dot
segments. Replaced by one strict allowlist:

`/^\/(learn|admin)(?:\/[A-Za-z0-9._~-]+)*$/`, plus a rejection of any segment that is `.`
or `..`.

No percent sign can appear, so encoded slashes and encoded CRLF cannot; `/login` and
`/courses` are not in the allowlist. Everything else falls back to `/learn`.

### Header

`Layout` reads the root loader's value and passes props to `SessionMenu`, which stays
prop-driven like every other presentation component here (`LoginView` takes `user`). It
renders the display name as text, a sign-out `Button`, and a `Link` to `/admin` only when
`canAccessAdministration`. Sign-out calls the reused use case, sets the session query to
`null`, and navigates to the catalog.

Accessibility — the rule the first draft failed: the sign-out control is a real button in
tab order with the kit's visible focus ring and a ≥44px target; the administration link sits
in the existing `<nav>` landmark; the display name is text, not a control; both locales
carry every string; no colour is hardcoded.

### Screens

`/learn` and `/admin` compose `EmptyState` directly in their route files, as
`courses.index.tsx` composes `CatalogView`. The first draft proposed two wrapper components
holding one JSX element each.

### Fixtures

Three identities appended to `backend/sql/seed.sql` — one file, not two: that file already
carries `draft-course` and `archived-course`, which have no more business outside
development than the fixtures do.

`ON CONFLICT (email) DO UPDATE` sets role, display name and hash from the file, so the seed
is the source of truth. `DO NOTHING` would keep an old credential while reporting success,
and the README would then document a password that returns 401. `docs/identity/plan.md` B09
was amended to match.

The identity section runs only when `SEED_IDENTITIES=1`. `APP_ENV` cannot carry this: the
`Makefile` never loads `.env`, `docker-compose.yml:61` defaults the variable to
`development`, and `.github/workflows/ci.yml:31` runs `make seed` with it unset — a guard on
it would both break CI and fail to defend the case it exists for. CI sets `SEED_IDENTITIES=1`
explicitly.

## Verification of the resolve race

The first draft's G09–G11 delayed `/auth/me` with Playwright and watched a full page load.
That call goes SSR-container-to-API over the internal network; `page.route` sees browser
traffic only, so those rows were unwritable.

With a root loader the race mostly dissolves: on a full page load identity is decided
**before** the HTML is sent, so there is no anonymous frame to observe. What remains is
verified where it happens:

- **Full page load** — assert the served HTML already carries the right answer. Raw-HTML
  test, no timing.
- **Client-side navigation** — a route test with a deferred MSW handler, which is where a
  browser-side wait is real.
- **Outage** — Playwright with the API container stopped, as `recovery.spec.ts` already
  does, needing no request interception.

## Сайд-ефекти

| Effect | Assessment |
|---|---|
| One `/auth/me` per SSR render for viewers carrying a cookie | Internal call. No repository-recorded latency exists; the first draft cited a figure that had never been written down. This slice measures it and records the number in the phase report |
| One `/auth/me` per page load from the browser, **including anonymous viewers** | Consequence of `staleTime: 0` and the `HttpOnly` cookie. Accepted: identity must revalidate, and the alternative is a stale header. Decision 1's "public traffic pays nothing" holds for SSR only |
| SSR HTML becomes viewer-specific | `server.mjs` keeps `Cache-Control: private, no-store` **and** gains `Vary: Cookie`, required by `docs/identity/spec.md:41` and `:49`. The first draft struck it citing a document that says nothing of the kind |
| A mounted protected route whose identity flips to `null` on refocus | `refetchOnWindowFocus` is at its default. The guard runs at navigation; a mounted route re-reads the query and redirects when it becomes anonymous |
| `route-guards.spec.ts` asserts behaviour this slice removes | Rewritten in the same task that deletes `LockedArea` |
| Seeded credentials live in the repository | Local-only, gated behind `SEED_IDENTITIES=1` |

## План міграції

`make seed` applies one file whose identity section is gated by `SEED_IDENTITIES=1`. No
schema change. Re-running reconciles fixtures to the file.

## Rollback

The slice lands as **one squashed commit**: `LockedArea` is deleted in it and four routes
point at the replacement, so a partial revert leaves dangling imports. The first draft said
"revert the slice commit" without reconciling that against `AGENTS.md:9`, which asks for
small logical commits — the reconciliation is that this slice is one logical change.

| Change | Undo | Lost |
|---|---|---|
| Routes, header, root loader, `Vary`, `check` target | `git revert` of the squashed commit | nothing |
| `safeReturnTo` narrowing | same commit; reverting widens it back to the I05 behaviour, which is the pre-existing contract violation | nothing |
| Seeded rows | `DELETE FROM users WHERE email IN (…the three fixture addresses…)`; `user_sessions` cascades (`migrations/00002_identity.sql:29`) | the fixtures only |

The fixture addresses are fixed in `backend/sql/seed.sql` and repeated in the README, so the
delete predicate is checkable rather than guessed.

## Конституція

Walked rule by rule against AGENTS.md. Domain keeps no framework or transport dependency —
the narrowed `safeReturnTo` stays import-free. Application untouched. Policy stays inward:
routes pass a domain predicate. No `any`; the loader's value is a discriminated union. No
`fetch` in presentation or routes. No dependency added. Accessibility addressed above, which
is the rule the first draft failed.

## Оцінка

`docs/identity/plan.md` budgets I06 at 1.5h. That predates the narrowed `safeReturnTo` and
its tests, a rewritten Playwright spec, a raw-HTML test, four route rewrites and the seed
reconciliation. **Revised to 4h**, and the phase task row is updated rather than left
looking on schedule.

## Приймальні сценарії

Rebuilt in [acceptance.md](acceptance.md): G09–G11 replaced per the section above, G19
restored to require `Vary: Cookie`, G29–G31 rewritten for `DO UPDATE` and `SEED_IDENTITIES`,
G37–G39 kept from the constitution fix, and rows added for a garbage cookie, duplicate
cookies, an unrecognised role, display-name bounds, repeated sign-out and dot-segment
`returnTo` values — every hole the defence named.
