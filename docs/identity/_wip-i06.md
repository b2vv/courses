# Living note — I06

Persist channel: none. `CLAUDE.local.md` does not exist in this repository, so this file is
the fallback the manifest prescribes. Flushed as decisions land, not at the end.

## Cycle state

| Step | State | Who decided |
|---|---|---|
| 0 manifest | read; refreshed first because it still described bootstrap (`8b5fc36`) | — |
| threshold | **not crossed** — I06 changes one backend layer (seed) only, so "more than one layer of both" fails. Full cycle runs on the user's explicit invocation, which the constitution allows to override ceremony | user |
| 1 setup | branch `feat/identity-foundation`; artifacts `docs/identity/i06/`; living note here | — |
| 2 specify | `i06/spec.md` written | — |
| 3 GATE 1 | `grilling` complete; all 5 questions closed, decisions in `i06/spec.md` §Рішення GATE 1. Awaiting the user's confirmation of shared understanding before the plan | user |
| 4 plan (3rd) | `i06/plan.md` rewritten twice. Edition 3 rests on the measured router spike; acceptance table rebuilt to 47 rows | — |
| 5 GATE 2 (1st) | **RED.** `plan-critique`: 11 blocking, 19 non-blocking. `plan-defense`: 0 covered, 5 holes. Constitution: 1 violation, closed | fresh delegates |
| 5 GATE 2 (2nd) | **RED.** `plan-critique`: 10 blocking, 17 non-blocking. `plan-defense`: 2 covered, 3 holes. Decision taken: spike the router before drafting again | fresh delegates + user |
| 5 GATE 2 (3rd) | **RED.** Four lenses, `4 запущено / 4 повернулось`; deduplicated to 8 blocking themes. Constitution inline: 1 blocking. `plan-defense` **not run** — it would judge a text that is about to change. Two forks put to the user | fresh delegates |
| 6 tasks | cut into I06a–I06h at edition 3; the cut itself is now a blocking finding (evidence per task) and is recut with edition 4 | — |
| 7 implement | — | — |
| 8 verify | — | — |
| 9 wrap | — | — |

## Decisions

- Slice artifacts live in `docs/identity/i06/`. The manifest names a directory per phase,
  not per slice; the phase directory already holds Phase 3's own `spec.md` and `plan.md`,
  so a slice subdirectory avoids two files with the same name meaning different scopes.

### GATE 1 answers

1. **SSR resolves identity only when the request carries a session cookie.** No cookie
   means anonymous by definition, so the call is skipped and public storefront traffic
   pays nothing. Closes open question 5 as well: the anonymous header is what the root
   already renders.
2. **A role refusal is a 404, not a refusal screen.** The user's reason: this is not a
   security claim but keeping the administration area out of sight of casual probing, so a
   wrong path looks like a wrong path. `notFound()` is already this repository's answer for
   an unusable address, so the mechanism is reused rather than invented. The real boundary
   stays server-side in Phase 4.
3. **REVISED.** Both protected addresses send an anonymous visitor to login with
   `returnTo`, and the role is checked after signing in. Superseded the answer below once
   the hiding was measured: `/admin` appears verbatim in the built client bundle, so it was
   never concealed from anyone with developer tools, and a login screen leaks nothing the
   bundle does not. A signed-in student still gets 404, which is the case the decision was
   actually protecting. Original answer kept for the record:

   ~~**Anonymous visitors are treated differently per route.** `/learn` sends them to login
   with `returnTo`, because a member area is not a secret and the header already offers
   "sign in". `/admin` answers 404 for everyone who is not permitted, anonymous included,
   so an administrator signs in first and navigates afterwards. A06 is therefore exercised
   on `/learn`.~~

4. **`/learn` shows an honest empty state**, reusing `EmptyState` from the UI kit.
5. **Header carries the name as text, a sign-out control, and an administration link only
   for permitted roles.** A menu was rejected: two items do not justify a popover and its
   focus management, and `Menu` already exists in the kit for when Phase 4 adds entries.
6. **`make seed` refuses unless `APP_ENV` is `development`**, and the credentials go into
   the README. A machine guard rather than an understanding, because the fixture creates an
   administrator with a known password.

Derived without asking, flagged for objection: two fixtures rather than three. B01's
instructor row is covered by the domain policy test, so a third seeded credential would add
no evidence.

### Defect the delegation found

`docs/identity/spec.md:52` requires `returnTo` to reject recursive `/login` targets and
non-protected paths. `safeReturnTo`, shipped in I05, rejects neither. The phase security
contract has been stricter than the code since I05 landed, and no test caught it because
the tests were written from the implementation rather than from the contract. Narrowing it
is a task in this slice; `G15` encodes both cases.

## Delegation failures

**Code graph consulted by index, not by tool.** `.claude/standards.md` §Розміщення carries a
reuse index and the graph itself is built (672 nodes, 6911 edges), but the
`code-review-graph` MCP tools are not live in this session — the server was registered
today and Claude Code reads MCP configuration at startup. The reuse survey was therefore
made against the manifest index and the code. Recorded because a later run with the tools
live may find reuse this pass missed. Any role executed inline instead of delegated is recorded here, because a silent
substitution makes the cycle unrepeatable.

## GATE 2 — red, and why

Verified every load-bearing finding against the code before accepting it. All held.

### Two citation failures by the plan's author

1. **Fabricated.** The plan struck `Vary: Cookie` from G19 claiming
   `docs/security/audit-2026-09-03.md` records that `private, no-store` makes it
   unnecessary. That document contains no occurrence of `Vary` at all, and the phase
   contract *requires* it twice (`docs/identity/spec.md:41`, `:49`).
2. **Unsupported.** "Measured at 0–2 ms on the internal network in I05" was the sole
   justification for the extra SSR call. The figure was seen in container logs during I05
   debugging but was never recorded as evidence anywhere in the repository, and the plan
   cited it as though it had been.

Both times the author overrode a delegate's row and was wrong both times.

### The design error that invalidates decision 1's implementation

The plan proposed an isomorphic cookie check "reading the request header on the server and
`document.cookie` on the client". The session cookie is `HttpOnly`
(`backend/internal/identity/adapters/http/handler.go:130`), so `document.cookie` can never
see it. Every client-side evaluation returns "anonymous" for exactly the signed-in viewers
the slice exists to serve. Decision 1 is implementable server-side only.

### The verification design error

G09, G10 and G11 delay or fault `/auth/me` with Playwright and observe a full page load.
Under a root loader that call happens SSR-container-to-API over the internal network;
`page.route` intercepts browser requests only. The three rows written to prove the single
hardest property of the slice cannot be written as specified.

### Contract contradictions the plan did not schedule

- 404-for-a-student contradicts `docs/identity/spec.md:45`, `:50`, `:51` and
  `docs/identity/plan.md:26`, all of which require a localized forbidden state.
- `ON CONFLICT DO NOTHING` repeals B09 (`docs/identity/plan.md:34`), which requires
  conflicting rows to fail loudly without partial fixtures, verified concurrently.
- The `APP_ENV` seed guard breaks CI (`.github/workflows/ci.yml:31` runs `make seed` with
  the variable unset) and does not defend its own threat model
  (`docker-compose.yml:61` defaults it to `development`).

## Second plan draft

Rewritten rather than patched. Every blocking finding is answered in the plan text and the
reasoning for the change is kept beside it, so the next reader sees why the first shape was
wrong rather than only the second shape.

Three decisions were taken with the user to unblock it:

1. The phase contract is amended to the 404 answer, in `docs/identity/spec.md` and B01.
2. The seed becomes the source of truth via `ON CONFLICT DO UPDATE`; B09 amended.
3. The seed gate moves from `APP_ENV` to an explicit `SEED_IDENTITIES=1`, because `APP_ENV`
   both breaks CI and defaults to `development` in Compose.

Estimate corrected in the phase task row: 1.5h → 4h.

## Found by the author while the second GATE 2 ran

**The second draft's guard cannot work as written.** It resolves identity in the root
*loader* and guards in each protected route's `beforeLoad`. In `router-core`,
`executeServerLane` (`load-server.js:415`) runs `contextualize` — the whole `beforeLoad`
lane, top-down across every match — to completion first, and only then the loader phase. A
child's `beforeLoad` therefore cannot see the root loader's data.

Correct shape: resolve identity in the root's **`beforeLoad`** and return it as context.
`load-server.js:187-190` sets `match.context = {...context, ...beforeLoadContext}` and
passes it down, which is exactly what context is for.

**Second consequence, from `load-server.js:159`:** `if (match.ssr === false || !route.options.beforeLoad) continue;`
— on the server, `beforeLoad` is skipped for `ssr: false` routes. `/learn` and `/admin` are
both `ssr: false`, so their guard runs on the client only. Identity is still resolved and
dehydrated server-side by the root, so there is no fetch to wait for, but the guard's
decision happens after hydration. G07's wording overstates this and needs correcting.

Not patched immediately: the critique and defence agents were reading the document at the
time, and editing under them would produce a verdict on a version that no longer exists.
Folded into one revision once they return.

## GATE 2, second round — red again

`plan-critique`: 10 blocking, 17 non-blocking. `plan-defense`: 2 covered, 3 holes.

Worst three:

1. **Sign-in loops.** `load-client.js:330` — the root match is retained and its loader does
   not re-run on a client-side navigation. `login.tsx:52-57` navigates after a successful
   sign-in, the root value still says `anonymous`, and `/learn`'s guard bounces back to
   `/login`. The design breaks the phase's main scenario.
2. **A requirement was lost while rebuilding the table.** Edition 1 covered the no-flash
   rule (A04); edition 2's table contains no row mentioning it. Rebuilding to close holes
   dropped coverage that already existed.
3. **The author's own contract amendment is self-contradictory.** B09 now says both "the
   seed file is the source of truth" and "conflicting rows fail loudly", which
   `ON CONFLICT DO UPDATE` can never satisfy. Half a sentence was rewritten.

Also: `SEED_IDENTITIES` cannot reach the container (`docker compose exec` does not forward
host environment); `i06/spec.md:104` still fixes decision 6 on `APP_ENV`; `redirect()` was
claimed as reused and is new; `/login` would resolve identity twice per SSR render;
`work/tasks.md:30` still reads 1.5h.

### The pattern, named

Both editions failed on the same class: **framework behaviour asserted without reading the
framework.** `document.cookie` against an `HttpOnly` cookie; the `beforeLoad`/loader
ordering; now root-match retention. Plausible every time, wrong every time. Alongside it,
false reuse claims (`redirect()`, "three call sites") and one fabricated citation (`Vary`).

A third draft written the same way fails on the fourth such assumption.

### Decision: verify the mechanism before drafting again

Agreed with the user. A throwaway spike against the installed router answers, with running
code rather than reasoning: whether a child's `beforeLoad` receives the root's `beforeLoad`
context; whether root `beforeLoad` and root `loader` re-run on client navigation; what a
`redirect()` and a `notFound()` thrown from a child `beforeLoad` actually produce. The
answers go here, then the third draft is written against them.

The 4h estimate is also unsound: the repository has no route-level tests, no SSR harness and
no seed integration test, and the table requires building six test surfaces from nothing.

## Spike results — measured, not reasoned

Throwaway tests against `@tanstack/react-router` 1.170.32 in this repository's own harness,
run and deleted. Four questions, four answers.

| # | Question | Answer |
|---|---|---|
| Q1 | Does a child's `beforeLoad` receive the root's `beforeLoad` context? | **Yes**, and the root runs first. `guarded.beforeLoad sees context.identity=anonymous`, ordered after `root.beforeLoad` |
| Q2 | Does the root re-run on a plain client-side navigation? | **`beforeLoad` yes, `loader` no.** Log across a navigation: `root.beforeLoad=anonymous`, `root.loader=anonymous`, *(identity changes)*, `root.beforeLoad=signed-in`, `learn.guard sees=signed-in`, landed on `/learn` |
| Q3 | What does `redirect()` from a child `beforeLoad` produce? | Lands on the target with search preserved: `pathname=/`, `search={"returnTo":"/guarded"}` |
| Q4 | What does `notFound()` from a child `beforeLoad` produce? | The match takes `status="notFound"`; `isNotFound`/`isRedirect` helpers identify both throws |

### What this settles

**One change resolves two blocking findings.** Moving identity from the root *loader* to the
root *beforeLoad* fixes both the ordering problem (a child's `beforeLoad` cannot read loader
data) and the sign-in loop (the loader is retained across client navigation, `beforeLoad` is
not). The critique's loop finding was correct about the design as written and dissolves
under the corrected placement — which is why it was worth measuring rather than arguing.

**The guard mechanism is confirmed workable**: `redirect()` and `notFound()` both behave as
the design needs, from exactly the place the design needs them.

**New cost, now measured rather than assumed:** root `beforeLoad` runs on *every* client
navigation. With `staleTime: 0` that is an identity revalidation per navigation, not per
document load. The third draft states this instead of the two different wrong numbers the
second draft gave.

## GATE 2, third round — red again

`4 запущено / 4 повернулось`; no lens went missing. Every load-bearing finding was
re-verified against the code before being accepted, and two lenses verified against running
systems rather than reading: one measured `\if` on the live Compose container, the other
reproduced the seed collision on Postgres 17 in a temp table under `ROLLBACK`.

Counts by lens: AI-slop 3/4, overcomplexity 4/3, logic gaps 7/5, edge cases 5/6
(blocking/non-blocking). Deduplicated to **eight blocking themes**. Constitution check ran
inline: 1 blocking, 4 non-blocking.

### The fourth instance of the named pattern

`_wip-i06.md` §"The pattern, named" predicted that a third draft written the same way would
fail on the fourth framework assumption. It did, and **all four lenses found the same one
independently**:

`ensureQueryData` does not revalidate. `query-core/build/modern/queryClient.js:78-81`:

```
const cachedData = query.state.data;
if (cachedData === void 0) return this.fetchQuery(options);
if (options.revalidateIfStale && …) this.prefetchQuery(defaultedOptions);
return Promise.resolve(cachedData);
```

`staleTime` is consulted only under `revalidateIfStale`, and even then the stale value is
returned while the refetch runs in the background. The plan's "`staleTime: 0` forbids a
stale answer" is false at this call site — the constant is real (`identity-queries.ts:32`),
the call simply never asks it. The method is additionally marked `@deprecated` in the
installed version.

**Why the spike did not catch it.** Q2 measured whether the root's `beforeLoad` *re-runs*
across a client navigation. It does. What was never measured is whether the *fetch* re-runs
inside it — a different question, answered by Query rather than by the router. The spike
was right and the conclusion drawn from it was wider than the measurement.

**Narrower diagnosis than the lenses stated.** Three of them wrote that the sign-in scenario
breaks. It does not: `login.tsx:53-56` already writes the cache with `setQueryData`, so the
guard reads a signed-in viewer straight after login. What breaks is *staleness detection* —
a revoked or expired session (G11, G12, G22a), and G08, whose precondition leaves the cache
populated so the deferred handler is never called and the row tests nothing.

### The other seven blocking themes

1. **"`frontend/src` has no route-level test at all" is false**, and the 12h estimate rests
   on it. `catalog-queries.test.tsx:40-68` already builds
   `createRootRoute`/`createRoute`/`createRouter`/`createMemoryHistory`, runs `router.load()`,
   navigates and asserts `router.state.location.pathname`; `CatalogView.test.tsx:172-184`
   renders `RouterProvider`; `e2e/shell.spec.ts` asserts SSR output with JavaScript disabled;
   `recovery.spec.ts:46` already uses the `request` fixture; `schema_integration_test.go:30-53`
   is a working Testcontainers+goose harness. What is missing is test *files*, not test
   *surfaces*. "twelve and not twenty" cites a figure that was never measured.
2. **I06g holds every proof, so I06a–I06f produce none.** `docs/identity/plan.md:3` requires
   each task to produce its own verification evidence. Six consecutive tasks without it
   breaks the phase rule and makes the "closes Gxx" column unverifiable until the end.
3. **The `safeReturnTo` allowlist contradicts its own acceptance row.**
   `^/(learn|admin)(/[A-Za-z0-9._~-]+)*$` cannot match `/admin/` — the group requires at
   least one character after each slash — so it falls back to `/learn`, which is exactly what
   G15b forbids. Plan and table disagree inside one line, with no router involved.
4. **"A primary-key collision cannot arise from a differing email" is backwards.**
   `00002_identity.sql:3-4`: `id` is the primary key, `email` is unique. With fixed ids and
   `ON CONFLICT (email)`, editing a fixture's address in the file produces `{id: X, email:
   new}` against `{id: X, email: old}`: no email conflict, a primary-key violation instead.
   Reproduced on Postgres 17 — `duplicate key value violates unique constraint`, psql exit 3
   under `ON_ERROR_STOP=1`, which aborts the whole seed including courses. Fixed ids are what
   makes this collision possible, not what prevents it. The amended B09 ("a changed fixture is
   applied rather than ignored") therefore does not hold for `email`.
5. **The seed recipe expands the gate variable in the wrong place.** `Makefile:23` keeps the
   whole `psql` call inside single quotes within `sh -c`, deliberately expanding
   `$$POSTGRES_USER` *in the container*. Appending `-v identities="$$SEED_IDENTITIES"` to the
   same line expands it there too, where `docker compose exec` never forwarded it —
   reproducing, one level down, the exact failure the third edition diagnosed in the second.
   Measured live: the branch is skipped with `unrecognized value "" for "\if expression"`,
   exit 0. CI would seed zero identities: G32 red, G36 **falsely green**.
6. **`make check` does not carry the variable**, so C03's "one command" is false: `check`
   runs `seed` without `SEED_IDENTITIES`, no fixtures exist, and every sign-in row in `e2e`
   fails. G41's "each step exit 0 on a clean checkout" is unreachable as written.
7. **G07 is unwritable as specified.** The header reads `['session']` through `useQuery` with
   `staleTime: 0`; dehydrated data is immediately stale, so `shouldFetchOnMount`
   (`queryObserver.js:319-321`) fires one client `/auth/me` per document load. A network
   counter asserting "no request at all after hydration" will see it, even though the guard
   itself asked for nothing.
8. **G12 has no mechanism and no owner.** The guard lives only in `beforeLoad`, which does
   not re-run when the cache changes. A mounted `/admin` whose identity turns `null` on
   window focus stays on screen. No plan step subscribes to `['session']` and navigates, and
   no task in `work/tasks.md` is assigned the row.

### Constitution check — inline, against AGENTS.md

**Blocking.** The single squashed commit (§Rollback) was justified by "this slice is one
logical change". That was true when I06 was one tracker row; it is not true across eight
tasks, two of which (`I06c`, `I06f`) declare no dependencies at all. `AGENTS.md:9` asks for
small logical commits, and the plan overrides it on a premise its own task cut retired.

**Non-blocking.** The cookie-by-name check has no owning module named, and
`courseforge_session` appears in `frontend/src` only inside tests — the constant lives in
`backend/internal/identity/adapters/http/config.go:11`. (The edge-cases lens supplied the
answer: `getCookie` is already exported from `@tanstack/react-start/server`,
`start-server-core/dist/esm/request-response.js:150`, so this is reuse, not new code.)
`make check` silently becomes a Compose-dependent target. The `append`-not-`set` reasoning for
`Vary` defends a case that does not exist — `server.mjs:23` sets `Cache-Control` and no `Vary`
header is present at all. `AGENTS.md:9` also requires a `docs/identity/report.md` entry;
`acceptance.md` G41 carries it, `work/tasks.md` I06h does not.

**Verified clean:** the domain stays import-free (`domain/identity/identity.ts:1`, with
`safeReturnTo` at `:66`); the Query cache is request-scoped — `QueryClient` is constructed
inside `getRouter()` (`router.tsx:12`), so `AGENTS.md:18` holds and G23 is writable.

### What the round also settled, so it is not reopened

The **router half of the design is sound**. The logic-gaps lens traced the full chain: the
root's `beforeLoad` runs on the server (`load-server.js:159` skips only the `ssr:false`
child), `__beforeLoadContext` is dehydrated as `b` (`ssr-server.js:25`), the client rebuilds
the root context from it (`load-client.js:1200-1204`), and the handoff load starts past the
root. `redirect()` and `notFound()` from a child `beforeLoad` behave as the design needs.
The failure this round is entirely in the **Query layer**, not the router layer — which is
why moving identity into the root `beforeLoad` stays correct.

The edge-cases lens separately cleared the "never throws" property end to end: empty body,
unknown role, missing `displayName`, 401, 500, timeout and client abort all surface as
`SessionQueryError` through `unwrap`, with `retry: false` (`router.tsx:14`).

### Two forks that need the user before a fourth edition

1. **`fetchQuery` costs a blocking round trip per client navigation.** It is the mechanism
   that makes G11/G12/G22a true, but the root's `beforeLoad` runs first and serially, so
   every in-app navigation waits on `/auth/me` — for anonymous visitors too, since the cookie
   is `HttpOnly` and the client can never rule itself out. The alternative is to keep the
   cheap read and drop the staleness rows. This is the third time the mechanism has moved,
   so it is put to the user rather than chosen quietly.
2. **The seed gate mechanism, for the third time.** `APP_ENV` failed, `SEED_IDENTITIES`
   through `\if` fails at the quoting boundary. The overcomplexity lens proposes a second
   file (`backend/sql/seed-identities.sql`) plus a second make target: no variable, no `-v`,
   no `\if`, no stderr noise; G36 holds by construction; the existing Go harness can apply it
   through `pool.Exec`, which it cannot do with a psql metacommand. Decision 6's "explicit
   confirmation" becomes a target name instead of an environment variable — the user's intent
   is unchanged, the mechanism is not.

## Fourth plan draft

Rewritten, not patched, for the third time. All eight blocking themes and the constitution
blocking are answered in the plan text, and the reasoning for each change is kept beside it.
**Zero blocking left open.**

Two decisions were taken with the user to unblock it. Both are mechanism choices whose
alternatives were measured rather than argued.

1. **Freshness is bought at the protected boundary, not on every navigation.** The root keeps
   `ensureQueryData` — a cache read costing nothing — and the `/learn` and `/admin` guards call
   `fetchQuery`, which with `staleTime: 0` always revalidates and deduplicates concurrent reads
   (`query.js:157-162`). The rejected alternative, `fetchQuery` in the root, would block every
   in-app navigation on a round trip, anonymous visitors included, because an `HttpOnly` cookie
   leaves the client unable to rule itself out. The slice's own scenario 6 asks about "the next
   navigation **to a protected address**", so the guarantee is bought exactly at the boundary
   the specification names. The header takes `refetchOnMount: false` — without it the dehydrated
   value is instantly stale and fires one `/auth/me` per document load — while
   `refetchOnWindowFocus` stays default, because G12 depends on it.
2. **The seed gate is a second file and a second target, not a third variable.**
   `backend/sql/seed-identities.sql` plus `make seed-identities`. No variable, no `-v`, no `\if`,
   no stderr noise. G36 holds by construction, since a target that is not invoked cannot seed;
   the existing Go harness applies plain SQL through `pool.Exec`, which it cannot do with a psql
   metacommand; and `ci.yml` changes one line instead of gaining an environment variable.
   Decision 6's "explicit confirmation" becomes the target name. `i06/spec.md` decision 6 is
   amended accordingly and now records all three mechanisms and why the first two failed.

Three further corrections worth naming, because each was a claim rather than an oversight:

- **The upsert conflicts on `id`, not on `email`,** and sets `disabled_at = NULL`. Edition 3
  had it backwards and would have aborted the whole seed on an address edit.
- **`safeReturnTo` is narrowed, not replaced,** and a single trailing slash is normalized away
  before the prefix test. Edition 3's allowlist regex contradicted its own acceptance row G15b.
- **The slice lands as several commits, one per task.** Edition 3's squash was justified by "one
  logical change", a premise its own eight-task cut had already retired (`AGENTS.md:9`).

The task table is recut: tests move to the tasks that own the behaviour, because
`docs/identity/plan.md:3` requires each task to produce its own evidence and edition 3 left six
consecutive tasks with none. The column is renamed `owns`. All 49 acceptance rows now have an
owning task; edition 3 left G10a, G10b, G12, G13, G14, G22 and G23 with none. The estimate falls
from ≈12h to ≈11.5h, and the false premise it rested on — "no route-level test at all" — is
corrected in both the plan and the tracker.
