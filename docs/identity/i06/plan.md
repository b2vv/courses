# I06 — plan

How the slice specified in [spec.md](spec.md) gets built. Placement follows
`.claude/standards.md` §Розміщення; the reuse survey below is checked against its reuse
index.

This file absorbs the plan section of the eventual report: the report will be the delta
against it, not a second plan.

## Reuse survey

The code graph was built for this repository (672 nodes, 6911 edges) but its MCP tools are
not live in this session, so the survey is against the manifest's reuse index and the code
itself.

### Reused unchanged

| Need | Reused | Why it fits |
|---|---|---|
| Read the viewer's identity | `app/composition/identity-queries.ts` — `current()`, `sessionQueryKey`, `SessionQueryError` | Already request-scoped and never-fresh, which is exactly what a guard needs |
| Sign out | `application/identity/session.ts` — `logOut` | Written in I05 with no consumer; this slice is its first |
| Role policy | `domain/identity/identity.ts` — `canAccessAdministration` | Already mirrors the Go policy and is unit-tested; the guard reads it rather than re-deriving |
| Safe return target | `domain/identity/identity.ts` — `safeReturnTo`, `app/router/login-search.ts` | Written and tested in I05 for exactly this case |
| Empty state | `presentation/ui` — `EmptyState` | `/learn` needs a title and a child link, which is its whole surface |
| SSR cookie forwarding | `adapters/api/request-headers.ts`, `adapters/api/transport.ts` | The root loader uses the same transport as every other adapter |
| Refusal mechanism | `notFound()` from the router, as already used by `app/router/catalog-search.ts` | Decision 2 asks for a 404; this is how this repository already produces one |

### Added

| Thing | Where | Why nothing existing fits |
|---|---|---|
| Root identity resolution | loader on `routes/__root.tsx` | No route resolves identity today; the header cannot render server-side without it |
| Guard for a protected route | `app/router/guards.ts` | Two routes need the same three-way decision (resolving / anonymous / role). Writing it twice is the duplication this session already removed once from the API adapters |
| Signed-in header | `presentation/identity/SessionMenu.tsx` | The root `Layout` currently hardcodes a "sign in" link; the signed-in variant has different content per role |
| Learning empty state | `presentation/learning/LearningHome.tsx` | New screen; composes the reused `EmptyState` |
| Administration shell | `presentation/admin/AdminShell.tsx` | New screen; deliberately minimal, and must not imply the instructor scoping that belongs to a later phase |
| Identity fixtures | `backend/sql/seed-identity.sql` | Kept out of `seed.sql` so the guard below can refuse identities without refusing course data |
| Seed environment guard | `Makefile` `seed` target | Decision 6 |

### Deleted

`presentation/LockedArea.tsx` and the `locked.*` translations. Its copy — "this area is not
available yet" — becomes false the moment the areas work, and every one of its three call
sites is replaced by a real screen or a guard. Leaving it would leave a lie reachable.

## Design

**Identity at the root.** `__root.tsx` gains a loader that resolves identity through
`context.identity.current()`. Per decision 1 it runs only when the request carries a session
cookie; the check is isomorphic, reading the request header on the server and
`document.cookie` on the client. Absent cookie resolves to `null` without a call.

**The guard.** One helper consuming route context, returning one of three outcomes rather
than a boolean, because decision 7 makes "not yet known" a distinct state from "anonymous":

- identity not resolved → the route suspends; it never decides on incomplete information
- anonymous → `redirect` to `/login` with `returnTo` set to the current address
- resolved, capability denied → `notFound()`
- resolved, capability allowed → render

`/learn` passes the "signed in" capability, `/admin` passes `canAccessAdministration`. The
capability is a predicate the caller supplies, so the guard has no knowledge of roles.

**The header.** `Layout` renders `SessionMenu`, which reads the same `['session']` query the
guard uses, so the two can never disagree. Anonymous keeps today's "sign in" link. Signed in
shows the display name, a sign-out control and — only when `canAccessAdministration` — an
administration link. Sign-out calls the reused use case, replaces the session query data
with `null`, and navigates to the catalog.

**Fixtures.** Three identities in a separate `seed-identity.sql`, `ON CONFLICT DO NOTHING`
so repeated runs are inert. Password hashes are literal Argon2id PHC strings generated once
with the repository's own parameters, so seeding needs no Go toolchain.

## Сайд-ефекти

| Effect | Assessment |
|---|---|
| Every SSR page with a session cookie now makes one extra internal call to `/api/v1/auth/me` | Measured at 0–2 ms on the internal network in I05. Anonymous traffic — the majority on a storefront — is unaffected by decision 1 |
| SSR HTML becomes viewer-specific | Already covered: `frontend/server.mjs:23` sets `Cache-Control: private, no-store` on all SSR responses, and the audit disposition records that this is why no `Vary: Cookie` is required |
| `route-guards.spec.ts` asserts behaviour that this slice deliberately removes | The spec is rewritten in the same task that removes `LockedArea`, not left failing |
| Seeded credentials exist in the repository | Local-only by the same convention as `.env.example`, and decision 6 adds a machine guard against a non-development environment |
| The header appears on every page, including the login screen | Intended: a signed-in visitor opening `/login` already sees the "already signed in" state from I05 |

## План міграції

No schema change: `users` and `user_sessions` already exist from I01. The migration is data
only, and additive.

1. `seed-identity.sql` is a new file; `make seed` applies both it and `seed.sql`.
2. Existing local databases are unaffected until `make seed` runs; running it twice is inert.
3. CI seeds a fresh database on every run, so no ordering problem arises between environments.
4. No production environment exists, so there is nothing to migrate forward.

## Rollback

| If | Then |
|---|---|
| The guard misbehaves after merge | Revert the slice commit. The routes return to `LockedArea` only if that revert includes its deletion — so `LockedArea` is deleted in the same commit as the guards land, keeping the revert atomic |
| The root loader proves too costly | Remove the loader alone; the header falls back to rendering after hydration, losing A04's no-flash guarantee but nothing else. This is a one-file change |
| Seeded identities reach an unwanted environment | `DELETE FROM users WHERE email LIKE '%@courseforge.local'`; sessions cascade. The seed guard makes this recovery rather than routine |

Every step is reversible with a revert or a single delete. Nothing in the slice writes data
a user could lose.

## Конституція

| Rule | Check |
|---|---|
| Domain has no framework, transport or database dependency | The guard lives in `app/router/`, not in domain; it reads the existing domain predicate |
| Application depends on domain and ports only | No application change in this slice |
| Handlers and routes translate and render, business policy lives inward | The capability predicate comes from `domain/identity`; routes pass it, they do not encode it |
| No `any`, no assertions disguising unsafe input | `returnTo` continues through the tested `safeReturnTo`; the guard's outcome is a discriminated union |
| No fetch in presentation or routes | The root loader calls the composition layer, as the catalog route already does |
| Add dependencies only when used | None added |
| Client guards are not the security boundary | Stated in the spec's Out of scope, and the administration shell exposes no data |
| Do not claim unimplemented capability | The administration shell says management arrives in a later phase; `/learn` shows an empty state rather than a fabricated list |

## Приймальні сценарії

Delegated to `acceptance-spec`; the table lands here before tasks are cut.
