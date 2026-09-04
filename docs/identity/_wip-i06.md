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
| 3 GATE 1 | `grilling` in progress; 3 of 5 open questions closed | user |
| 4 plan | — | — |
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
3. **Anonymous visitors are treated differently per route.** `/learn` sends them to login
   with `returnTo`, because a member area is not a secret and the header already offers
   "sign in". `/admin` answers 404 for everyone who is not permitted, anonymous included,
   so an administrator signs in first and navigates afterwards. A06 is therefore exercised
   on `/learn`.

## Delegation failures

None yet. Any role executed inline instead of delegated is recorded here, because a silent
substitution makes the cycle unrepeatable.
