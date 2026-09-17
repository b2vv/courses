# Phase 4a — admin course management: delivery report

Delta to [plan.md](plan.md) (edition 2.1, GATE 2 green). Written 2026-09-05 at slice
close. One report per feature; the Post-deploy section is left for `post-deploy-check`.

## 1. What was done

Fifteen tasks A01–A15, each implemented with Codex (gpt-5.6-sol) and re-verified here
before commit; per-task evidence in `work/report-a01.md` … `report-a14.md`. Everything
landed as planned except the deviations below.

- **Backend** (A01–A07): migration 00003 adds `courses.instructor_id` (nullable add →
  eligible-owner backfill → NOT NULL, failing loudly on a users-less volume),
  `course_modules`/`course_lessons`; the gate order reversed `seed-identities` before
  `seed` across Makefile/CI/README/manifest; authoring domain policy, port and use
  cases; the PostgreSQL adapter with the guarded full-tree write; the authorization seam
  (`httpserver.Actor`, exported `Authenticate`/`RequireSameOrigin`) and the seven admin
  operations with the full refusal matrix; the public course program.
- **Frontend** (A08–A12): the kit i18n debt paid (required labels, locale split); the
  authoring adapter/application (branded ids, business-code mapping); the guarded shell
  layout and course list; the RHF course editor; degradation and in-flight guards.
- **E2E + gate** (A13–A15): shared per-role browser sessions with a measured login
  budget; the admin/instructor/refusal browser journeys; the final ordered gate.

**Deviations from the plan:**
- **B19 budget ≤6 → ≤7** (amended in plan.md, dated). Two logout tests each need an
  independent session because logout revokes every presented session (`1292b3a`); the
  true minimum is 3 role setups + 2 login-flow tests + 2 disposable logout sessions = 7.
  Validated empirically by two consecutive e2e runs, the historical 429 trigger.
- **R35 observable amended** (dated): "visible alert + immediate redirect" is
  unobservable by timing; the row is the redirect-only behavior, consistent with
  identity E10.
- **A12 existed by design** (planned): the degradation path is its own small task because
  A10/A11 were at the 2h ceiling.

## 2. Acceptance scenarios

80 rows from the plan's acceptance table. Each is owned by exactly one task, and every
task's evidence ran green inside the final `make check` (exit 0, twelve steps). Rows
"not verified": **0**.

| ID | Scenario | Owner | Actual | Verdict |
|---|---|---|---|---|
| P01 | Every course has an owner | A01 | closed by A01 evidence; in the exit-0 `make check` | збіглося |
| P02 | Admin creates a draft | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| P03 | A draft is invisible publicly | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| P04 | Instructor owns what they create | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| P05 | Admin picks the owner at create | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| P06 | Admin reassigns the owner | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| P07 | Owner candidates are listable | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| P08 | A tree is saved whole | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| P09 | Reordering persists | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| P10 | A lesson moves between modules | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| P11 | Publishing makes a course public | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| P12 | Archiving withdraws it | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| P13 | Restoring republishes it | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| P14 | An instructor publishes their own course | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| P15 | The public page shows the program | A07 | closed by A07 evidence; in the exit-0 `make check` | збіглося |
| P16 | Admin list renders | A10 | closed by A10 evidence; in the exit-0 `make check` | збіглося |
| P17 | The kit carries no Ukrainian of its own | A08 | closed by A08 evidence; in the exit-0 `make check` | збіглося |
| P18 | The admin screen is localized | A10 | closed by A10 evidence; in the exit-0 `make check` | збіглося |
| P19 | The editor saves a tree | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| P20 | The full admin journey in a browser | A14 | closed by A14 evidence; in the exit-0 `make check` | збіглося |
| P21 | An instructor manages their own course | A14 | closed by A14 evidence; in the exit-0 `make check` | збіглося |
| B01 | An empty course may be published | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| B02 | Empty admin list | A10 | closed by A10 evidence; in the exit-0 `make check` | збіглося |
| B03 | Slug uniqueness survives concurrency | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| B04 | A slug is mutable before publication | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| B05 | A maximal legal tree saves | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| B06 | Tree bounds are refused | A05 | closed by A05 evidence; in the exit-0 `make check`. **Amended 2026-09-17 (backlog B-3):** the `121 lessons` case now expects `400 lesson_limit`, the other two stay `400 invalid_request` — dated in `plan.md` §8, evidence `work/report-b3.md`. The proof does not weaken; it splits into three cases carrying two codes | збіглося |
| B07 | Oversized bodies are refused | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| B08 | Summary bounds | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| B09 | A deep tree stays operable | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| B10 | Pagination boundary | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| B11 | Timestamps survive the round trip | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| B12 | The token round-trips | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| B13 | Migration over a populated database | A01 | closed by A01 evidence; in the exit-0 `make check` | збіглося |
| B14 | Migration over courses without users | A01 | closed by A01 evidence; in the exit-0 `make check` | збіглося |
| B15 | Down migration | A01 | closed by A01 evidence; in the exit-0 `make check` | збіглося |
| B16 | The existing catalog suite stays green | A01 | closed by A01 evidence; in the exit-0 `make check` | збіглося |
| B17 | Contract and client stay in step | A07 | closed by A07 evidence; in the exit-0 `make check` | збіглося |
| B18 | The ordered gate passes | A15 | closed by A15 evidence; in the exit-0 `make check` | збіглося |
| B19 | The e2e login budget fits the zone | A13 | closed by A13 evidence; in the exit-0 `make check` | збіглося |
| R01 | Anonymous is refused on every read | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R02 | A student is refused on every read | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R03 | A disabled account is refused | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R04 | A foreign instructor cannot read | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R05 | An instructor cannot list owner candidates | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R06 | Unknown ids are not found | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R07 | Malformed ids are rejected before lookup | A04 | closed by A04 evidence; in the exit-0 `make check` | збіглося |
| R08 | Anonymous is refused on every mutation | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R09 | A student is refused on every mutation | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R10 | Cross-site mutation is refused first | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R11 | A foreign instructor cannot write | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R12 | An instructor cannot reassign ownership | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| R13 | An ineligible owner is refused | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| R14 | Illegal transitions are refused | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R15 | A published course cannot be deleted | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R16 | Delete cannot race a publish | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| R17 | A draft deletes with its tree | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| R18 | A concurrent save is refused, not merged | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| R19 | A payload cannot steal another course's nodes | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| R20 | A published course's slug is immutable | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| R21 | A rename onto a taken slug is refused | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| R22 | Lifecycle members are rejected by the write schema | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| R23 | A transition retires outstanding tokens | A06 | closed by A06 evidence; in the exit-0 `make check` | збіглося |
| R24 | A dependency failure degrades safely | A02 | closed by A02 evidence; in the exit-0 `make check` | збіглося |
| R25 | Cancellation is not a fault | A02 | closed by A02 evidence; in the exit-0 `make check` | збіглося |
| R26 | A cancelled save leaves nothing behind | A03 | closed by A03 evidence; in the exit-0 `make check` | збіглося |
| R27 | Business codes reach the domain intact | A09 | closed by A09 evidence; in the exit-0 `make check` | збіглося |
| R28 | The conflict state shows the stored tree | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| R29 | A replayed create makes at most one course | A05 | closed by A05 evidence; in the exit-0 `make check` | збіглося |
| R30 | Double submission is impossible | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| R31 | The owner picker is admin-only | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| R32 | An unavailable picker does not block saving | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| R33 | The client guard is not the mechanism | A14 | closed by A14 evidence; in the exit-0 `make check` | збіглося |
| R34 | Unusable admin search params disappear | A10 | closed by A10 evidence; in the exit-0 `make check` | збіглося |
| R35 | A session that ends mid-edit is handled | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| R36 | An invalid form is refused before it is sent | A11 | closed by A11 evidence; in the exit-0 `make check` | збіглося |
| R37 | The list survives an unavailable primary read | A12 | closed by A12 evidence; in the exit-0 `make check` | збіглося |
| R38 | The editor survives an unavailable primary read | A12 | closed by A12 evidence; in the exit-0 `make check` | збіглося |
| R39 | Duplicate node ids in one payload are refused before any statement | A02 | closed by A02 evidence; in the exit-0 `make check` | збіглося |
| R40 | The editor's Save cannot be double-submitted | A12 | closed by A12 evidence; in the exit-0 `make check` | збіглося |

## 3. Non-obvious findings

- **Lockfile vs Docker `npm ci`** — a host `npm install` after adding react-hook-form
  produced a lock the frontend image's `npm ci` rejected (missing optional wasm deps
  `@emnapi/core`/`@emnapi/runtime` pulled by `@rspack/binding-wasm32-wasi` and
  `@tailwindcss/oxide-wasm32-wasi`). `make build` failed at `npm ci`. Fixed by
  regenerating the lock in-container (`docker run node:24-bookworm npm install
  --package-lock-only`). Recorded in `docs/admin/_wip-phase4.md`.
- **Shared storageState vs logout** — the shared per-role session token is one server
  session; logout revokes every presented one (`1292b3a`), so a logout test kills the
  shared token for the rest. Reproduced with curl. Fix: logout tests provision disposable
  sessions (`frontend/e2e/route-guards.spec.ts`, `accessibility.spec.ts`).
- **The catalog mapper decoupling** — migration 00003 adds a column to sqlc's `Course`
  model, so the nine-column public projections generate query-specific row types; the
  mapper now consumes scalar projections so those queries survive a rollback of 00003
  (`backend/internal/course/adapters/postgres/repository.go`, A03 report).
- **The optimistic lock had to be made to bite** — nothing bumped `courses.updated_at`
  (only `DEFAULT now()`); the guarded write sets it explicitly with a microsecond clock,
  or R11/R18 would have been silent lost updates (`D-3b`; caught at GATE 2, not runtime).

## 4. Screenshots

Not applicable as static images here, but the admin flow is exercised visually by
`frontend/e2e/admin.spec.ts` against real Compose (create → structure → publish →
public program → archive). The public program render is covered by `CourseView` tests.

## 5. User flow walk-through

Verified by the `admin.spec.ts` e2e run (not by hand). An admin signs in, opens
`/admin`, sees the owner-scoped course list with status tabs, clicks Create course,
fills title/slug/summary/description and picks an owner, saves a draft, adds a module and
a lesson and saves the tree, presses Publish — the course then appears in the public
catalog and its page shows the program — and later Archive removes it from the catalog.
An instructor signs in, edits a course they own, and is refused on one they do not. A
student's direct API mutation is refused by the server.

## 6. Risks and limitations

- **Slice 4b (identity administration) is not built** — block/restore/soft-delete/
  password reset remain outlined in `spec.md`; `disabled_at` still has no operator.
- **Per-course user blocking** waits for enrollment (no substrate yet).
- **No super_admin role** — admin-protection invariants live in the domain; a real admin
  hierarchy is a later decision.
- **Retained debt** (plan §10): NB-7's literal task-merge was declined (would breach the
  2h ceiling); the two layer-split task pairs own acceptance rows instead. Five
  non-blocking notes from GATE 2 round 2 (N1–N5) remain as recorded, none blocking.
- **Lesson content is title + plain text only** — rich text (TipTap) and media fields
  arrive with their real consumers (Phase 5/7).
- **Client guards remain navigation convenience**; the server handlers are the
  authorization mechanism, proven by the refusal matrix and R33.

## Post-deploy

_Left for `post-deploy-check` after deployment. No deploy target exists in the repo yet._
