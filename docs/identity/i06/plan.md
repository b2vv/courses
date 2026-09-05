# I06 — plan

Fourth edition. Three earlier ones were refused at GATE 2; [../_wip-i06.md](../_wip-i06.md)
holds all three verdicts, the router spike and the Query-layer measurements this design rests
on. Placement follows `.claude/standards.md` §Розміщення.

Edition 3 failed on the same class as editions 1 and 2 — framework behaviour asserted without
reading the framework — but one layer down: the router half was correct and the **Query** half
was not. Every mechanical claim below was executed or read in the installed package before it
was written, and the reading is cited so the next reader can repeat it rather than trust it.

## Measured facts this design rests on

### Router — `@tanstack/react-router` 1.170.32

| Fact | Consequence |
|---|---|
| A child's `beforeLoad` receives the root's `beforeLoad` context, root first | Identity resolves in the root's `beforeLoad`; the guard reads it from context |
| The root's `beforeLoad` re-runs on every client navigation; its **loader does not** | Identity lives in `beforeLoad`. In the loader it would freeze at document load |
| `redirect()` from a child `beforeLoad` lands on the target with `search` preserved | The `returnTo` round trip works from the guard |
| `notFound()` from a child `beforeLoad` sets `status="notFound"` | The 404 answer works from the guard |
| `beforeLoad` is skipped on the server for `ssr: false` matches (`load-server.js:159`) | `/learn` and `/admin` decide client-side |
| `__beforeLoadContext` dehydrates as `b` (`ssr-server.js:25`) and the client rebuilds the root context from it (`load-client.js:1200-1204`), starting the handoff load past the root | The guard's client-side decision needs **no** network call: the root's answer crossed in the HTML |

### Query — `@tanstack/query-core` 5.102.8

Read in `frontend/node_modules/@tanstack/query-core/build/modern/`. This is the layer edition 3
got wrong.

| Fact | Consequence |
|---|---|
| `ensureQueryData` returns cached data whenever it exists and consults `staleTime` only under `revalidateIfStale` — and even then returns the **stale** value while refetching in the background (`queryClient.js:78-81`). It is `@deprecated` in this version | It is a **read**, not a revalidation. Edition 3's "`staleTime: 0` forbids a stale answer" was false at this call site |
| `fetchQuery` tests `query.isStaleByTime(staleTime)` and fetches when stale (`queryClient.js`) | With `staleTime: 0` it always revalidates. This is the mechanism when freshness is actually required |
| `query.fetch` returns the in-flight retryer promise instead of starting a second request (`query.js:157-162`) | Concurrent reads deduplicate. Two matches asking at once cost one call |
| `setQueryData` rejects only `undefined` (`queryClient.js:92`) | `null` is a writable value, so "anonymous" can be **stored** rather than merely returned |
| `shouldFetchOnMount` refetches when data exists and is stale (`queryObserver.js:319-321`) | A header observer on `staleTime: 0` fires one `/auth/me` per document load unless `refetchOnMount` is set false |

### Postgres and Compose — executed, not read

| Fact | Consequence |
|---|---|
| With `id` as primary key and `email` unique (`00002_identity.sql:3-4`), `ON CONFLICT (email)` plus a fixed id and a changed address raises `duplicate key … "users_pkey"`, psql exit 3 under `ON_ERROR_STOP=1` | Reproduced on Postgres 17 in a temp table under `ROLLBACK`. The conflict target must be `id`, or an address edit aborts the entire seed including courses |
| `docker compose exec` does not forward host environment; a `$$VAR` inside the single-quoted `sh -c` of `Makefile:23` expands **in the container**, where it is empty | Measured live: the `\if` branch is skipped with `unrecognized value ""`, exit 0. CI would seed zero identities and report success |

## Reuse

**Reused unchanged:** `identity-queries.current()`, `sessionQueryKey` and its `staleTime: 0`;
`session.logOut` (written in I05, still without a consumer); `canAccessAdministration`;
`EmptyState`, **`ErrorState` with `retryHref`** (`presentation/ui/states.tsx:15-27`), `Button`
and `Link`; the shared API transport; **`SessionQueryError`** with its
`dependency_unavailable` code (`identity-queries.ts:5-12`); `notFound()`, already used by
`app/router/catalog-search.ts:77`; `createIsomorphicFn` request options from
`app/composition/identity.ts:7`.

**Reused, found this round:** **`getCookie`** from `@tanstack/react-start/server`
(`start-server-core/dist/esm/request-response.js:150`), which parses the request's cookies
through h3. The cookie-presence check is therefore not new code, and no hand-rolled
`includes('courseforge_session=')` is written — that form gives a false positive on
`xcourseforge_session=` and, worse, a false negative on an unusual header, which would render
an anonymous header server-side and a signed-in one after hydration: exactly the flash A04
and G21a forbid.

**Reused, correcting edition 3:** the repository **does** have a route-level test harness.
`app/composition/catalog-queries.test.tsx:40-68` builds
`createRootRoute`/`createRoute`/`createRouter`/`createMemoryHistory`, runs `router.load()`,
navigates and asserts `router.state.location.pathname`;
`presentation/storefront/CatalogView.test.tsx:172-184` renders `RouterProvider`;
`e2e/shell.spec.ts` asserts SSR output with JavaScript disabled; `e2e/recovery.spec.ts:46`
already uses Playwright's `request` fixture; `schema_integration_test.go:30-53` is a working
Testcontainers + goose harness. What this slice lacks is test **files**, not test **surfaces**.

**New to this slice:** `redirect()` — `grep` finds no router `redirect()` anywhere in
`frontend/src`.

**Changed:** `safeReturnTo`; `__root.tsx` gains a `beforeLoad` and its `Layout` the signed-in
header; `app/composition/identity.ts` gains the cookie check and an SSR deadline;
`routes/{learn,admin}.tsx` gain guards and real screens; `routes/{learn,admin}.$.tsx` lose
`LockedArea`; `Makefile` gains `seed-identities` and rewrites `check`;
`frontend/server.mjs` gains `Vary`; `.github/workflows/ci.yml` calls the new target.

**Deleted:** `presentation/LockedArea.tsx` and the `locked.*` resources — **four** call sites:
`routes/{learn,learn.$,admin,admin.$}.tsx`.

## Design

### Identity in the root

`__root.tsx` gains a `beforeLoad` returning `{ identity }` into context, where

`identity: { status: 'anonymous' } | { status: 'signed-in', user } | { status: 'unavailable' }`

It never throws — `__root.tsx:35` carries `errorComponent`, so a thrown identity fault would
replace every public page with a failure screen. Every failure mode already arrives as a
`SessionQueryError` through `unwrap` (empty body, unknown role, missing `displayName`, 401,
500, timeout, client abort), and `retry: false` (`router.tsx:14`) means no retry delay hides
inside it; the `beforeLoad` catches and maps rather than propagating.

It reads through **`ensureQueryData`**, which is a cheap cache read. That is deliberate and
its limit is stated rather than disguised: the root does **not** revalidate. Freshness is
bought where it is needed, in the guard, not on every navigation in the product — see
§Freshness.

**Cookie check.** On the server the root asks the API only when the request actually carries
the session cookie, via `getCookie('courseforge_session')` inside the `.server()` branch of
`createIsomorphicFn` in `app/composition/identity.ts`. Checking for *a* `Cookie` header would
be useless: `LanguageSwitcher.tsx:8` writes `courseforge_locale` with a year's `Max-Age`, so
any visitor who once switched language carries one forever. On the client there is nothing to
inspect — the cookie is `HttpOnly` (`identity/adapters/http/handler.go:130`) — so the client
always asks. The cookie **name** is a backend constant
(`backend/internal/identity/adapters/http/config.go:11`) and appears in `frontend/src` today
only inside tests; this slice introduces it into frontend production code, in composition,
and that crossing is named here rather than discovered later.

**The no-cookie branch writes the cache.** It calls `setQueryData(sessionQueryKey, null)`
rather than merely returning `anonymous`. Without it, `login.tsx:47`'s `useSuspenseQuery`
meets an empty cache during SSR, suspends, and fetches `/auth/me` on the server for a visitor
who has no cookie — breaking decision 1 on the one route where it matters most. `null` is
writable (`queryClient.js:92` refuses only `undefined`) and dehydrates as a success state.
With it, "everything reads `['session']`" becomes true instead of aspirational.

**SSR deadline.** `transport.ts:26` is 5000 ms, too long to block a storefront render on an
identity lookup. The `.server()` branch of `requestOptions` passes `timeoutMs: 1000`. Because
the `.client()` branch does not, `logIn` and `logOut` — which only ever run from the browser —
keep the default. I06a records the observed SSR call duration beside this number, so the
figure stops being an assumption; no such figure exists anywhere in this repository today and
edition 1 cited one that had never been written down.

`login.tsx:12` drops its own loader and reads the root's value, or the route resolves identity
twice per SSR render.

### Freshness — where it is bought, and what it costs

Decided with the user this round, after the cost of the alternative was measured rather than
assumed.

| Where | Mechanism | Cost |
|---|---|---|
| Root, every navigation | `ensureQueryData` | zero requests |
| `/learn` and `/admin` guard | `fetchQuery` on `identity.current()` | one `/auth/me`, blocking, **only when entering a protected address** |

`fetchQuery` with `staleTime: 0` always revalidates, and `query.js:157-162` deduplicates a
concurrent read, so two matches asking together cost one call.

The rejected alternative was `fetchQuery` in the **root**, which would make every in-app
navigation — catalog browsing included, and anonymous visitors included, since an `HttpOnly`
cookie leaves the client unable to rule itself out — wait on a round trip. The slice's own
scenario 6 (`i06/spec.md`) asks that "the **next navigation to a protected address** behaves
exactly like the anonymous case", which is precisely the boundary chosen: the guarantee is
bought exactly where the specification asks for it and nowhere else.

The header therefore reads `['session']` with `refetchOnMount: false`. Without that, the
dehydrated value is immediately stale (`staleTime: 0`) and `shouldFetchOnMount`
(`queryObserver.js:319-321`) fires one `/auth/me` per document load, which G07 asserts must
not happen. `refetchOnWindowFocus` stays at its default, because G12 depends on it.

### Guard

`/learn` and `/admin` carry the guard in their **own** `beforeLoad`. No new route files: per
the spike, a parent's `beforeLoad` already covers its nested addresses, and `learn.tsx`
renders no `Outlet` today, so `/learn/go/intro` shows the parent's screen — which is the
correct answer while no lesson content exists. `learn.$.tsx` and `admin.$.tsx` keep
`ssr: false` and lose their `LockedArea` component.

One guard, `requireIdentity(context, location)`, which revalidates and returns the
`AuthenticatedUser`:

| Result | Guard |
|---|---|
| `signed-in` | returns the user |
| `anonymous` | `throw redirect({ to: '/login', search: { returnTo } })` |
| `unavailable` | `throw new SessionQueryError({ code: 'dependency_unavailable' })` into the route's own `errorComponent`, which renders `ErrorState` with a `retryHref` — the localized retryable state `docs/identity/spec.md:50` requires |

The role check is one line in `admin.tsx`, not a second guard:
`if (!canAccessAdministration(user)) throw notFound()`. Edition 3 proposed two named guards
whose `anonymous` and `unavailable` branches were identical; the difference was always a
single domain predicate, and the predicate stays in the domain while the policy stays in the
route.

`returnTo` is the current path **including its search string**; a deep link to
`/learn/go/intro?lesson=3` must survive.

**A mounted route whose session dies.** `beforeLoad` does not re-run when the cache changes,
so a guard alone cannot answer G12: an open `/admin` whose identity turns `null` on window
focus would sit on screen indefinitely. The layout component subscribes to `['session']` and,
on a `signed-in → null` transition, navigates to `/login` with the same `returnTo`. This is
the one mechanism the guard cannot provide, and edition 3 had neither the mechanism nor an
owning task for the row.

### `safeReturnTo`

A **narrowing of the shipped function**, not a replacement. `identity.ts:66-78` already
rejects absolute, protocol-relative and backslash forms, `%2f` prefixes and control
characters; `docs/identity/spec.md:52` asks additionally that only normalized internal
protected paths survive. Two additions to the existing body:

1. a path prefix test `^/(learn|admin)(/|$)`, which also rejects `/login` and `/courses` and
   so closes the defect this slice inherited from I05;
2. rejection of any `.` or `..` segment.

A single trailing slash is normalized away **before** the test and the normalized path is what
is returned. Without that, `/admin/` fails the test and falls back to `/learn` — which is what
G15b explicitly forbids, and edition 3's allowlist regex `^/(learn|admin)(/[A-Za-z0-9._~-]+)*$`
did exactly that: plan and acceptance table contradicted each other inside one line. The search
string keeps the existing control-character check and gains nothing else: once the path prefix
is fixed, a query string cannot change origin.

### Header

`Layout` reads `['session']` and passes props to a prop-driven component, like `LoginView`.
Name as text, a sign-out `Button`, and a `Link` to `/admin` only when
`canAccessAdministration`. Sign-out calls the reused use case, sets the query to `null`, and
navigates to the catalog. The component is **not** called `SessionMenu`: `Menu` exists in the
kit and GATE 1 decision 5 explicitly rejected a menu, so the name would contradict the
decision it implements.

Accessibility: the sign-out control is a real button in tab order with the kit's visible focus
ring and a ≥44px target; the administration link sits in the existing `<nav>`; the name is
text; both locales carry every string; no colour is hardcoded.

### Screens

`/learn` and `/admin` compose `EmptyState` in their route files, as `courses.index.tsx`
composes `CatalogView`.

### Fixtures

Three identities in a **separate file**, `backend/sql/seed-identities.sql`, applied by a
separate target, `make seed-identities`. Decided with the user this round; it is the third
mechanism for this gate and the first that is not a variable.

`APP_ENV` failed because the `Makefile` never loads `.env`, `docker-compose.yml:61` defaults
it to `development`, and CI runs `make seed` with it unset. `SEED_IDENTITIES` through a psql
`\if` fails one level lower: `Makefile:23` keeps the whole `psql` call inside single quotes
within `sh -c`, deliberately expanding `$$POSTGRES_USER` **in the container**, and the natural
edit expands the gate variable there too — where `docker compose exec` never forwarded it.
Measured: the branch is silently skipped and the command exits 0, so CI would seed nothing and
report success.

A second file removes the whole class. There is no variable, no `-v`, no `\if` and no stderr
noise; G36 holds by construction, because a target that is not invoked does not run; and the
existing Go harness applies the file through `pool.Exec`, which cannot execute a psql
metacommand. Decision 6's "explicit confirmation" becomes the target name rather than an
environment variable — the user's intent, keeping fixtures out of a shared environment, is
unchanged.

The upsert conflicts on **`id`**, not on email:

```sql
ON CONFLICT (id) DO UPDATE SET
  email = EXCLUDED.email, display_name = EXCLUDED.display_name,
  role = EXCLUDED.role, password_hash = EXCLUDED.password_hash,
  disabled_at = NULL, updated_at = now()
```

Edition 3 conflicted on email with fixed ids and claimed that fixed ids made a primary-key
collision impossible. The opposite is true and was reproduced: editing an address in the file
produces `{id: X, email: new}` against `{id: X, email: old}`, no email conflict, and a
`users_pkey` violation that aborts the whole seed under `ON_ERROR_STOP=1`. Conflicting on `id`
makes the file authoritative for every field including the address, which is what the amended
B09 actually promises. `disabled_at = NULL` is part of it: without it, a fixture disabled
during testing stays disabled through a reseed and its login answers 401 with nothing to
explain why.

The transaction opens with `LOCK TABLE users IN SHARE ROW EXCLUSIVE MODE`, so two concurrent
seeds serialize instead of racing into a `users_pkey` violation between the conflict check and
the insert. B09 asks that concurrency leave no partial fixtures; without the lock that holds,
but "both runs succeed" becomes a rare flake.

A row inserted by hand carrying a fixture's address under a different id still fails loudly.
That is correct and stays uncaught by design: it is a real collision in a developer's database,
not a fixture concern.

### `Vary` and `check`

`server.mjs:23` sets `Cache-Control: private, no-store` on every SSR response and sets no
`Vary` at all. `Vary: Cookie` is added with `append`, which here is identical in effect to
`set`; the reason is simply that a later locale variation is expected — `docs/identity/spec.md:41`
expects variation on the locale inputs too, and declaring that is out of this slice and
recorded as such. Edition 3 justified `append` by a clobbering that could not occur.

`check` becomes `format-check lint typecheck test integration check-generated build up seed
seed-identities smoke e2e` — the manifest's ordered sequence with the new target in it. Two
consequences are stated rather than left to be discovered. First, `check` stops being a
Compose-free local command; `AGENTS.md:29`'s seven targets keep their relative order, so this
is an extension, not a divergence. Second, `seed-identities` **must** be in the list: edition 3
put the gate variable only in CI, so `make check` would have seeded no fixtures and every
sign-in row in `e2e` would have failed while C03 claimed "one command".

`up` already builds (`Makefile:9-10`), so the sequence builds twice. That is inherited from the
manifest and CI rather than introduced here, and it is accepted knowingly.

## Сайд-ефекти

| Effect | Assessment |
|---|---|
| One `/auth/me` per SSR render for a viewer carrying the session cookie | Internal call, capped at 1000 ms. I06a records the observed duration; no repository-recorded figure exists and edition 1 cited one that had never been written down |
| One `/auth/me`, blocking, on entering `/learn` or `/admin` | The price of scenario 6, bought only at the boundary that scenario names. Concurrent reads deduplicate (`query.js:157-162`) |
| **No** identity request on ordinary client navigation, and none after hydration | The root reads cache; the header sets `refetchOnMount: false`. A stale identity is therefore possible on public pages between navigations, and is corrected on window focus and at every protected boundary |
| A hung API stalls SSR for 1000 ms before `unavailable` is returned | Down from the transport default of 5000 ms; the client-side branch is untouched |
| SSR HTML becomes viewer-specific | `private, no-store` already; `Vary: Cookie` added |
| Seeded credentials in the repository | Local-only, behind a target that must be invoked by name |

## План міграції

No schema change. `make seed` applies courses as before; `make seed-identities` applies
fixtures and reconciles them to the file on every run.

## Rollback

The slice lands as **several commits**, one per task, not as one squash. Edition 3 argued for
a single squashed commit because "this slice is one logical change" — true when I06 was one
tracker row, false across eight tasks, two of which declare no dependencies at all.
`AGENTS.md:9` asks for small logical commits and the premise that overrode it has since been
retired by the slice's own task cut.

| Change | Undo | Note |
|---|---|---|
| `safeReturnTo` narrowing (I06c) | independent revert | touches nothing else |
| Fixtures, target, CI (I06f) | independent revert | touches nothing else |
| Guards, screens and the `LockedArea` deletion (I06b, I06e) | revert together | the routes point at the replacements, so a partial revert dangles imports |
| Root `beforeLoad`, composition, header (I06a, I06d) | revert together | the header reads what the root writes |
| Seeded rows | `DELETE FROM users WHERE email IN (…)`; sessions cascade (`00002_identity.sql:29`) | — |
| Contract amendments (`b5a0cd7`, decision 6) | **not covered by the above** | separate commits. Reverting the code leaves a contract describing a 404 answer and a reconciling seed that no longer exist. Either revert those too, or accept the contract running ahead of the code and say so |

## Конституція

Domain stays import-free: `domain/identity/identity.ts:1` has no imports and `safeReturnTo`
(`:66`) is narrowed with standard language only. Application untouched. Policy inward: the
guard calls a domain predicate. No `any`; identity is a discriminated union. No `fetch` in
presentation or routes — the cookie check and the SSR deadline live in
`app/composition/identity.ts`, not in a route file (`AGENTS.md:15`). Query caches stay request
scoped: `QueryClient` is constructed inside `getRouter()` (`router.tsx:12`), which is what
makes G23 writable (`AGENTS.md:18`). No dependency added. Accessibility addressed above.

Fixture passwords are committed. `AGENTS.md:8` forbids committing secrets; these are demo
credentials for a local database behind a target that must be named, documented in the README,
and fixed by the user at GATE 1. Recorded as knowingly accepted rather than passed over.

`AGENTS.md:9` also requires this phase's verification, limitations and next task to be recorded
in `docs/identity/report.md`. That is I06h's closing step, and the task row says so.

## Оцінка

`docs/identity/plan.md:3` caps a task at 2h **and requires each task to produce its own
verification evidence**. Edition 3 collected every test file into one trailing task, which
left six consecutive tasks with nothing to show and made their "closes Gxx" columns
unverifiable until the end. The tests move to the tasks that own the behaviour.

Sized at **≈11.5h across eight tasks**. Edition 3 said twelve and justified it with "`frontend/src`
has no route-level test at all" — false, as §Reuse now records — and with a comparison to a
figure of twenty that was never measured. The remaining cost is real but smaller: test files on
harnesses that already exist, plus one genuinely new integration test for the seed.

## Приймальні сценарії

[acceptance.md](acceptance.md), amended for this edition: G07 rewritten now that the header's
refetch is understood, G08 rewritten for what `pendingComponent` actually does, G11/G22a
rewritten to the protected boundary, G12 given the mechanism it lacked, G15b reconciled with
the trailing-slash rule, G34 extended to a changed address and a disabled fixture, G36
rewritten for the target, and the seven rows that had no owning task assigned to one.
