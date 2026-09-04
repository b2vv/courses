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
| 5 GATE 2 | — | — |
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
