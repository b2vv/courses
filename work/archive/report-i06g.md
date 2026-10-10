# Report — I06g

Browser evidence against a running Compose stack. Written by Codex (`gpt-5.6-sol`); **none of it
was executed by the implementer** — see below — so every assertion here was run, and two were
fixed, from the reviewing session.

## The implementer could not run a single browser test

Chromium would not launch in its sandbox: *"blocked before assertions by macOS Mach-port
permission"*. The preparation that made the previous tasks verifiable — bringing the stack up and
seeding it beforehand, on the theory that Playwright only needs an outbound connection to
`localhost:8080` — was not enough. The block is not the network, it is the browser process itself.

The consequence is that this task's specs were authored blind, and two of them were wrong in ways
only a run could reveal. Both are fixed here.

## What was implemented

| File | Change |
|---|---|
| `frontend/e2e/route-guards.spec.ts` | rewritten — it still asserted the Phase 1 placeholder that I06b deleted, so `make e2e` was red before this task |
| `frontend/e2e/accessibility.spec.ts` | keyboard reach of the sign-out control; 200% zoom in dark theme with target sizes |
| `frontend/e2e/recovery.spec.ts` | G10a — the static public page survives with the API stopped |
| `frontend/server.mjs` | `Vary: Cookie` appended beside the existing `Cache-Control` |

## One production line moved here from I06h, deliberately

`server.mjs` had no `Vary` header at all. G21 asserts one, and G21 is owned by this task — so the
line that makes it true had to land with the test that proves it. The plan assigned it to I06h;
that ordering would have left a test that could not go green until a later task, which is a red
gate rather than evidence. The brief for this task superseded the plan on that point.

`docs/identity/spec.md:41` also expects variation on the locale inputs. That stays undeclared and
out of this slice.

## Two defects found by running what could not be run

**1. The G07 dehydration assertion could never have matched.** It demanded
`data … displayName:"Адміністратор Курсу" … queryKey … ["session"]` in that order. The real
payload is:

```
dehydratedData:$R[17]={query:…initial:…state:$R[21]={data:$R[15],…,status:"success"},queryKey:$R[22]=["session"]…
```

The serializer **deduplicates by reference**: the user object is emitted once elsewhere and the
entry holds `data:$R[15]`, a pointer. The display name is therefore never between `data:` and
`queryKey:`, and no amount of widening the gaps would have helped.

Replaced with an assertion against the shape that actually exists — a dehydrated `["session"]`
query in `status:"success"` — with the display name asserted separately on the document, where it
genuinely appears. Both halves of the row still hold: the identity crosses in the HTML, and the
client does not fetch to learn it.

**2. `Vary` was correct in the source and absent from the running stack.** The container had been
built before the change. Not a code defect, but a procedural one worth recording: a change to
`server.mjs` needs `docker compose up -d --build frontend` before `make e2e` means anything.

## Acceptance rows

| Row | Where |
|---|---|
| G07 | `route-guards.spec.ts` — raw HTML carries the identity and the dehydrated session; after hydration exactly **one** `/auth/me`, the guard's boundary revalidation |
| G21 | same test — `Cache-Control: private, no-store` **and** `Vary: Cookie`, name and sign-out control in raw HTML |
| G21a | `MutationObserver` installed through `addInitScript` **before** hydration; no anonymous frame is ever observed |
| G22a | three public transitions cost **zero** `/auth/me`; entering `/learn` costs one |
| G17, G18 | sign-in from `/admin` as admin lands on the shell; as student lands on `/admin` and is not-found there, without bouncing to `/login` |
| G27 | Back after sign-out shows no name and no protected content |
| G28 | `/login?returnTo=%2Fadmin` with a live session offers a control that continues to `/admin` |
| G03, G40 browser halves | student on `/learn`, admin on `/admin` |
| G29, G30 browser halves | `accessibility.spec.ts` |
| G10a, G10b | `recovery.spec.ts`, **run here** — the implementer could not stop a container |

## Verification

Run entirely from the reviewing session.

| Gate | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass |
| `make typecheck` | pass |
| `make test` | pass — frontend 175 tests / 19 files / 0 failures; all Go packages ok |
| `make integration` | pass — 12 Go packages, 0 failures |
| `make e2e` | pass — **20 tests**, 16 chromium and 4 recovery |

## A property of the suite worth knowing before it wastes someone's afternoon

**`make e2e` fails when run twice in quick succession**, and the failure looks like a product bug:
`expect(received).toBe(200) … Received: 429`.

It is the login rate limit doing its job. Several specs sign in through the real endpoint, and
`infrastructure/tests/smoke.sh` deliberately asserts that thirty login attempts produce a 429 —
the limit is a security control this repository verifies on purpose, not an accident.

The contract is unaffected: G41 asks for the ordered sequence on a **clean checkout**, and from a
fresh stack the suite passes, which was confirmed twice. `docker compose restart nginx` resets the
limiter state if a rerun is needed locally.

Not fixed, and deliberately so. The two ways to make the suite immune are to raise the limit,
which weakens a control the repository tests for, or to share an authenticated `storageState`
across specs, which is a redesign of the suite for a problem that only appears when a human runs
it twice. Recorded instead, with its trigger: **if the suite ever grows enough sign-ins to trip
the limit within a single run, share the storage state rather than touching the limit.**

## Notes

Re-running `make e2e` regenerates `docs/foundation/screenshots/*.png` as a side effect of
`accessibility.spec.ts` and `recovery.spec.ts`. The byte-level churn carries no new information
and was reverted rather than committed.

## Limitations

Nothing here is authorization evidence. These rows prove navigation behaviour and SSR output;
server-side role enforcement remains Phase 4, and the slice specification says so explicitly.
