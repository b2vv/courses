# Phase 4 living note — admin course management

Persist channel: repository file (no CLAUDE.local.md channel exists); flushed early and
often per spec-flow. Superseded entries stay, corrections append.

## 2026-09-05 — setup and spec

- Threshold called: Phase 4 crosses the spec-flow threshold (new subsystem, public API
  change, both stacks) — `.claude/standards.md` §Поріг.
- Branch `feat/admin-crud` from `main` = `e9edfd5`; artifacts home `docs/admin/`.
- `spec.md` written: intent only, 8 open questions carried to GATE 1 with
  recommendations. Standing decisions §24 of the product brief are treated as closed and
  not re-asked.
- GATE 1 (grilling) pending: user answers required on Q1–Q8 before any plan sentence.
- Delegation note: implementation is expected to go to Codex via detached briefs
  (working model), with the option of handing slices to peer session `courses-df`.

## 2026-09-05 — GATE 1, Q1 answered (scope widened)

User decision on Q1: **both roles mutate; instructor is ownership-scoped to own courses
now** — this amends brief §24.3 (ownership-scoped mutations arrive in Phase 4; the
instructor-workspace phase keeps students/progress/media views). Additionally named:
user administration (admin blocks/restores/deletes users, resets passwords; soft delete
with unique-email-among-non-deleted), per-course user blocking by instructors, and a
possible super-admin role (admin cannot delete an admin, only block). Placement of these
is being grilled next; per-course blocking has no substrate before enrollment exists.

## 2026-09-05 — GATE 1 complete, awaiting confirmation

All eight questions answered one-by-one (AskUserQuestion); decisions table lives in
`spec.md` §GATE 1 decisions, amendments 7-9 appended to the product brief §24. Scope
settled: Phase 4 = slice 4a (courses, ownership-scoped) + slice 4b (identity
administration, outlined). Next: user confirms shared understanding → plan.md.

## 2026-09-05 — plan edition 2 (GATE 2 RED → rewrite)

GATE 2 (plan-critique, fable, 4 lenses) returned RED on edition 1: 12 blocking, 18
non-blocking; plan-defense left row 3 (race and replay) open on `POST /admin/courses`.
`plan.md` rewritten as edition 2. `gates.md` untouched (read-only by mandate).

**User decision folded in as decided:** an instructor may publish, archive and restore
courses they own (edition 1's D-4 open question, answered yes). §10's open-question section
is gone; the decision is D-4a.

**Blocking findings closed, by location:** BLK-1 → D-3d (arithmetic redone at 2 B/char;
bounds cut to 40 modules / 120 lessons / 1000-char description; worst case 348 KiB under a
512 KiB cap, so §5.10's "nginx unchanged" stays true). BLK-2 → D-3b (every accepted write
bumps `updated_at` with a microsecond-truncated application clock, even when only children
changed). BLK-3 → D-3c (upsert guarded by `course_id` / module membership plus an
affected-rows check; foreign node ids answer 400). BLK-4 → §9 (A04 owns refusals only for
the three paths it registers; R08–R11 move to A06). BLK-5 → D-3g (the idempotency sentence
corrected; conflict refetches and says the save may already have landed). BLK-6 → D-3e
(slug mutable while never published, immutable after; `409 slug_immutable`). BLK-7 → D-3f
(delete and transition are single guarded statements; the status request carries
`expectedUpdatedAt`). BLK-8 → D-4b (PUT's `instructorId` is the carrier; `400 invalid_owner`
for an ineligible owner; P06 is the happy row). BLK-9 → D-4c (picker mounts for admins only;
an unavailable list degrades to read-only owner without blocking save). BLK-10 → D-7
(shared `storageState`; the recorded trigger has fired; ≤6 logins per run, B19 measures it).
BLK-11 → P17 owned by A08 at kit level, P18 owned by A10 at screen level. BLK-12 → §2, §5.3,
§7, A01 — and two hazards the critique had not named: the test executes `seed.sql` against a
database with no users, and its synthetic `00003_failure.sql` now collides with the real
migration (renumbered to `00004`, real 00003 added to the MapFS). Defense row 3 → D-3g
(create is not idempotent; the unique slug is the replay guard, `409 slug_taken` on replay,
submit disabled in flight; R29/R30 are the rows).

**Other decisions now explicit:** slug mutability (D-3e), owner-reassignment carrier (D-4b),
delete-vs-transition race under Q5 (D-3f), owner picker placement (D-4c), e2e login budget
(D-7).

**Non-blocking:** 17 of 18 addressed (§10 table). **Retained: NB-7 only**, and only its
literal-merge remedy — A02/A08 stay as separate tasks but are re-cut on evidence environment
(fakes versus real PostgreSQL; MSW versus a rendered route) and now own rows, because merging
them would produce 3.5-hour tasks against this phase's two-hour ceiling.

Shape of edition 2: 76 acceptance rows (21 happy / 19 boundary / 36 refusal), 0 manual, every
row owned by exactly one of 14 tasks, every task owning at least one row; 22.5 h computed from
the estimate column. Files dropped versus edition 1: `staleness.ts`, `CourseFormFields.tsx`,
`AdminStates.tsx`. Files added: `e2e/auth.setup.ts` and the Playwright `setup` project.

## 2026-09-05 — plan edition 2.1 (GATE 2 round 2: critique GREEN, defense rows 2/4 open)

Round-2 verification passed the critique (all 12 blocking closed, no new blocking) and closed
defense rows 1, 3 and 5; rows 2 and 4 stayed open on the same F1/F3 evidence gaps, now stated
precisely: no acceptance row exercised `dependency_unavailable` on the **primary** admin reads
(course list, course tree) with the user seeing the kit's `ErrorState` — R24/R27/R32 covered the
application layer, the adapter mapping and the owner-picker only — and D-3c's
duplicate-ids-within-one-payload rule lived in prose with no row (B06 tests bounds, R19 tests
cross-course theft; both are different rules). Edition 2.1 adds four rows: **R37** (list read 503
→ `ErrorState` with a working retry) and **R38** (editor tree read 503 → same, and no empty form
pretending the course has no content), both owned by a new task **A12**; **R39** (two nodes sharing
one client-minted id → invalid-tree outcome with the repository never called, which is what "before
any statement runs" means) owned by **A02**, whose estimate rises 1.5 h → 1.75 h; and **R40** (the
editor's Save cannot be double-submitted), the gap the defense noted beside row 3 — taken rather
than retained, because it costs one component test on the harness R28/R30 already build. A12 exists
because A10 and A11 were already at the two-hour ceiling: instead of pretending the extra tests fit,
the degradation path became its own 0.75 h task cut around one class of executable evidence (a
primary read answering 503, and a mutation still in flight) rather than around a screen, landing
after both screens exist because it asserts on both. The three e2e/gate tasks renumber A12–A14 →
A13–A15, and their owned rows (B19; P20/P21/R33; B18) follow. New shape, verified mechanically
against the file: **80 rows** (21 happy / 19 boundary / 40 refusal), 0 manual, each owned by exactly
one of **15 tasks**, every task owning at least one row, **23.5 h** summed from the estimate column.
Prose now points at the new rows from D-3c and D-3g so no rule is asserted without an owner.


## 2026-09-05 — GATE 2 GREEN on edition 2.1

Round 1 (edition 1): critique 12 blocking / 18 non-blocking; defense row 3 open — RED.
Edition 2 closed all 12 + row 3; round 2 critique verification GREEN; defense left rows
2/4 open (F1/F3 evidence gaps) — edition 2.1 added R37–R40 and task A12; defense 5/5
GREEN. Constitution inline PASS. D-4 user decision folded: instructors publish/archive/
restore their own courses. Traps recorded: concurrent appends to gates.md lost the first
defense section (reconstructed from the agent report — appenders must never rewrite the
file); one non-blocking nit open for implementation: §2's CourseEditorView row omits
"error state" though R38 owns it. Next: tasks to work/tasks.md, Codex on A01.

## 2026-09-05 — A11 review round

Codex could not run Rstest; first real run happened at review. Six failures, three
causes, all fixed at review: (1) four editor tests queried 'Course title' while the
en bundle renders 'Title' (uk 'Назва') — tests aligned to the bundles; (2) the locale
alignment test compared raw keys and broke on Ukrainian plural forms _few/_many —
now compares plural-suffix-normalized key sets; (3) R35 promised a visible alert AND
an immediate redirect — mutually exclusive by timing; the row is amended to the
redirect-only observable, consistent with identity E10's established behavior.

## 2026-09-05 — A13 review round (two real traps caught at the live gate)

1. **Lockfile trap.** RHF was added via host `npm install`, which produced a lock the
   Docker image's `npm ci` rejected (missing optional wasm deps @emnapi/core, @emnapi/
   runtime pulled by @rspack/binding-wasm32-wasi and @tailwindcss/oxide-wasm32-wasi).
   `make build` failed at `npm ci`. Fix: regenerate the lock with the image's node via
   `docker run node:24-bookworm npm install --package-lock-only`. Rule for next time:
   after adding a frontend dep, regenerate the lock in-container, never trust a host
   `npm install` lock for `npm ci`.
2. **Shared-storageState vs logout collision.** A13's first cut put logout tests on the
   shared per-role storageState. One server session per role token; logout revokes every
   presented session (1292b3a), so the first logout test killed the shared token for the
   rest — reproduced with curl (authentication_required). Fix: logout tests provision a
   disposable session each via a context-level API login; read-only authenticated tests
   keep the shared state. Rule: shared storageState is for session-preserving tests only;
   any session-destroying test owns its session.

## 2026-09-06 — Phase 4b GATE 1 complete

Slice 4b (identity administration) specified in docs/admin/spec-4b.md. Five questions +
one capability-model fork answered: sessions die immediately on block/delete (same tx);
password reset revokes sessions; role change out of 4b; deleted users shown under a
status filter; last-active-admin protected by a domain invariant; capabilities are
additive (brief §24.10). Next: plan-4b.md → GATE 2 → tasks → Codex.

## 2026-09-06 — plan-4b edition 2 (GATE 2 RED → rewrite)

GATE 2 on edition 1 returned RED (`gates-4b.md`): plan-critique (fable, 3 of 4 lenses —
the complexity lens was launched but never delivered) found 5 blocking / 22 non-blocking;
plan-defense (sonnet) left row 3 (race/replay) open. `plan-4b.md` rewritten as edition 2.
`gates-4b.md` untouched (read-only by mandate).

**One structural decision closes most of it: soft delete now writes the row *deleted AND
disabled* in the same statement (`deleted_at` set, `disabled_at = COALESCE(disabled_at,
$now)`); deleted ⊂ disabled.** This closes BLK-3, both omitted-query halves of BLK-4 and
the LockActiveIdentityUser half of defense row 3 with no course-module query change.

**Blocking closed, by location:**
- BLK-1 (`email_taken`/B02 dead — restore never clears `deleted_at`) → removed the code
  from §2/§8, the B02 row, the T04 mapping and risk #10; kept the deleted→404 rule (§1,
  §4 D-3, §8). The partial index stays (S4 email reuse) but no 4b op can violate it.
- BLK-2 (`count(*) FOR UPDATE` invalid; "lock other admins" deadlocks → 503) → §4 D-2:
  `LockAdmins :many` locks the whole active-admin set (target included) `ORDER BY id FOR
  UPDATE`, count in the domain, loser re-reads and gets 409 `last_admin` not 503; R10 adds
  "never 503".
- BLK-3 (rollback resurrects soft-deleted accounts) → deleted ⊂ disabled makes the 4a
  binary's `Active()` refuse them; §7 rewritten on that basis; T04 asserts old-password
  login fails under both credential queries (P05).
- BLK-4 (only 3 of 5 status-gating queries) → deleted ⊂ disabled makes
  `LockActiveIdentityUser`, `ResolveIdentitySession`, `ListInstructors` and the course
  module's `LockEligibleAuthoringOwner` correct by construction; only
  `GetIdentityCredential` gains `deleted_at IS NULL` (§5.a). Intra-tx order fixed (users
  row locked before sessions); R22 is the login-vs-block race row.
- BLK-5 (blocking an instructor bricks their course PUTs, `invalid_owner`) → §5.h states
  and **accepts** the frozen-until-reassign stance (4a's eligibility lock unchanged); R23
  pins it via the course authoring integration test.

**Defense row 3 (block vs in-flight login)** → §4 D-2 reasons about `Rotate`'s
`LockActiveIdentityUser FOR UPDATE` as the serialising lock; R22 exercises it; §11's false
"only real race is the last-admin count" is gone (§11 is now a pointer list).

**Non-blocking: all 22 addressed, none retained** (§12 table). Notable idiom fixes:
dropped `ports.PasswordHasher` for the module's consumer-side `PasswordVerifier` widened
with `Hash` (NB-19); plain-string ids, no branding (NB-17); extend `admin.ts`/
`admin-search.ts` rather than twin them, import `adminListStaleMs` (NB-16/18); frontend
`/admin/users` guard refuses instructors via `canAdministerAccounts` + `notFound()`
(NB-4); removed the fabricated briefing quote (NB-14); `DeletedAt` on `domain.User` +
mapper (NB-2); same-state block idempotent (NB-6); block writes `updated_at` for the CHECK
(NB-10); canary cited as plan.md BLK-12 / fd90fc9 (NB-22).

Shape of edition 2: 44 acceptance rows (8 happy / 12 boundary / 24 refusal), 0 manual,
every row owned by exactly one of 9 tasks, every task except T02 (unit-only) owning ≥1;
13.75 h computed. §13 flags that the GATE 2 re-run must actually deliver the complexity
lens (not delivered on edition 1).

## 2026-09-06 — plan-4b edition 2.1 (GATE 2 round 2 GREEN → fold round-2 non-blocking)

GATE 2 round 2 (`gates-4b.md`) returned GREEN: plan-critique (fable, 4 of 4 lenses — the
complexity lens delivered this round) found 0 blocking / 12 non-blocking (NB2-1 … NB2-12);
plan-defense re-ran 5/5 with the block-vs-login race verified in code. `plan-4b.md` bumped
to edition 2.1; `gates-4b.md` untouched.

**Two folded because T01 trips over them at its own `make integration`:**
- NB2-4 (email-reuse fixtures sit before a `DownTo` that fails on a reused email; B08 needs
  a repository) → B08 re-homed to the identity `schema_integration_test.go` (has
  `NewRepository`/`VerifyCredentials`); P05 stays in the identity administration harness,
  which runs no 00004 `DownTo`; every reuse-pair subtest deletes the deleted row before any
  `DownTo`. Stated in §2, §5.b/§6 Down, the B08 row and T01. B06 stays in the course harness.
- NB2-5 (R10's "never 503" silently assumes READ COMMITTED) → §4 D-2 states all 4b writes use
  `pool.Begin` (READ COMMITTED, the identity `Rotate` template) and warns against copying the
  course module's `BeginTx(RepeatableRead)`, which would abort the loser 40001 → 503.

**Two structural simplifications (complexity lens):**
- NB2-8 → collapsed `SetUserDisabledAt` + `RestoreUser` into one nullable-parameter UPDATE
  (`net: -1 query`); idempotence still decided in Go over the locked `disabled_at`.
- NB2-9 → added `CHECK (deleted_at IS NULL OR disabled_at IS NOT NULL)` (00004 step 3) and
  `Active() = DisabledAt == nil && DeletedAt == nil`, so *deleted ⊂ disabled* is enforced by
  the schema, not one COALESCE; §5.a/§7 now cite the constraint.

**The other eight, one line each in §12 (all Addressed, none retained):** NB2-1 (downgraded the
over-claimed §12 cells for NB-7/NB-11 to "stated, no dedicated row"); NB2-2 (restatement reduced
to a pointer; deleted §13 gate note now the gate is green); NB2-3 (renamed `PasswordVerifier` →
`Passwords`; search escaping is a duplicated one-line replacer, not a cross-module import);
NB2-6 (reset is UPDATE-then-revoke; old-password TOCTOU recorded as an accepted limitation);
NB2-7 (restore always uses `LockUserForStatus`; `LockAdmins` only for a block of an admin);
NB2-10 (B12 uses a real clock ≥ DB now(); limitation extended to `deleted_at`); NB2-11 (reset on
a blocked target → 204, block stays; deleted → 404); NB2-12 (R22 extended to block **and**
delete, plus double-delete 404).

Shape unchanged in spirit: 44 acceptance rows, 0 manual, every row owned by one of 9 tasks,
T02 unit-only; the query surface is one shorter and the migration one CHECK longer
(`net: -1 query, +1 constraint`). §13 (gate note) deleted. Reply to coordinator carried the
re-homed harness per row and the CHECK's exact text.
