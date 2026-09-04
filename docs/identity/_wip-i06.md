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
| 4 plan | `i06/plan.md` written: reuse survey, design, side effects, migration, rollback, constitution. Acceptance table delegated to `acceptance-spec` and integrated: 36 rows, two corrected by the author, one reported gap discarded as false, one accepted as a real defect | — |
| 5 GATE 2 | **RED.** `plan-critique`: 11 blocking, 19 non-blocking. `plan-defense`: 0 covered, 5 holes. Constitution: 1 violation, closed. Tasks are not cut | fresh delegates |
| 6 tasks | — | — |
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
