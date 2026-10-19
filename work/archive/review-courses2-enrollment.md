# Code review (Courses 2) — enrollment, slice by slice

Reviewer: Courses 2 (Opus 4.8). Skills applied per slice: `code-review` (vs spec/plan/standards/diff) and
`pattern-architecture-review` (GoF/KISS/DRY/SOLID/Clean Arch). Read-only; nothing committed. Subagent
fan-out was avoided deliberately — three verifier subagents died on the model safeguard
(`reasoning_extraction`) earlier this session, so each slice is read inline.

---

## L1 (T1) — domain grant→membership + migration 00015 + backfill + seed — **CHANGES REQUESTED**

Commit: uncommitted working tree on `feat/enrollment` (as of 2026-10-06). Files: `migrations/00015_enrollment.sql`,
`internal/orders/domain/{grant.go,errors.go,grant_test.go}`, `sql/seed.sql`.

**Verdict: CHANGES — 1 blocking, 4 non-blocking.** The domain model and the migration internals are
right; the migration breaks three *unchanged* call sites in the orders adapter that the slice does not
touch and T2 has not yet rewritten.

### Blocking

**B-L1.1 — the migration breaks the orders purchase/grandfather/grant paths (shared persisted state, outside the diff).**
The migration drops the partial index `course_grants_one_active_per_user_course` and replaces it with a
full `UNIQUE (user_id, course_id)` (`00015_enrollment.sql` UP). Three sites in
`internal/orders/adapters/postgres/repository.go` — none touched by L1 — depend on the old index:

- `:133-135` `GrantWithAudit`: a plain `INSERT` whose unique-violation is mapped to
  `ErrCourseAlreadyOwned` via `isUnique(err, "course_grants_one_active_per_user_course")`. After the
  migration the violation carries the new name `course_grants_one_per_user_course`, so the branch is
  dead: an admin re-grant onto an existing row falls through to a generic `persistenceError` instead of
  `course_already_owned`. The semantics also change — the plain INSERT now conflicts against *any* row,
  not only an active one.
- `:176-182` grandfather and `:251-252` paid-event settle: both use
  `ON CONFLICT (user_id,course_id) WHERE state='active' DO NOTHING`. That arbiter is inferred from the
  partial index the migration drops. With no partial unique matching the predicate, Postgres raises
  `42P10` *"there is no unique or exclusion constraint matching the ON CONFLICT specification"* — a
  statement-level failure. **Purchase settlement and grandfathering break outright.**

This is the AGENTS lesson in the flesh: *"a change to shared or persisted state is checked with the full
e2e set, not a targeted one; review of the diff misses breaks outside the diff."* As it stands, L1's own
ordered `make check` (orders integration + payment e2e) will be red, and merging the migration ahead of
the T2 rewrite would break production purchase.

Resolution (either):
- fold the T2 rewrite of these three sites (UPSERT / new constraint name) into the **same** commit as
  the migration, so the slice is gate-green; or
- if L1 must land first, update the three sites in L1 too (at minimum the `isUnique` name and the two
  `ON CONFLICT` arbiters), with an orders integration run.
Do not gate or merge the migration while `repository.go` still names the old index.

### Non-blocking

1. **Backfill is scoped to `status='published'`, which the plan did not ask for.** Plan §5 backfills
   "self-active membership for each (user,course) on a **free** course" with no status filter; the
   migration adds `AND courses.status = 'published'`. A user with historical progress/comments on a free
   course that is **archived at migration time** gets no membership, so if that course is re-published
   they must re-enroll (free re-enroll is self-service, so no hard lockout — but it is a behaviour
   change the plan's "nobody locked out, B5 union" did not state). Confirm the status filter is intended,
   or drop it.
2. **The `revoked → active` edge is not a guarded domain method.** `Block`/`Unblock`/`Revoke` are
   guarded transitions on `Grant`, but re-enrolment / admin re-grant (`revoked→active`, plan §4) is not
   modelled; `SelfEnroll` just builds a fresh active grant. T2 will do it via UPSERT, which means the
   important guard — *paid* revoked must not be self-reactivated (plan §4: "для paid revoke знімає лише
   адмін"), only free — will live in the application, not the domain. Ensure T2's review covers that
   guard with a test; consider a domain `Reactivate(bySelf bool, paid bool)` for symmetry and
   testability.
3. **`participant_reports` FKs have no `ON DELETE`.** `course_id`→`courses(id)` and
   `reporter_id`/`target_id`→`users(id)` default to RESTRICT. Users are soft-deleted so that is fine,
   but a course that has reports cannot be hard-deleted (there is a `course_not_deletable` path). Decide
   the policy before reports are wired (L8), and compare to `discussion_reviews` which used
   `ON DELETE CASCADE`.
4. **Re-revoke is not idempotent.** `Revoke()` on an already-`revoked` grant returns
   `ErrGrantTransitionInvalid` (correct as a pure transition), so the application must treat a repeat
   admin-revoke as a no-op rather than surfacing an error. Note for T2.

### What is correct (verified, not assumed)

- Domain transitions `Block`(active→blocked), `Unblock`(blocked→active), `Revoke`(active|blocked→revoked)
  all guard the source state and return `ErrGrantTransitionInvalid` otherwise (`grant.go:55-80`). The
  unit table `TestGrantTransitionsFollowMembershipStateMachine` covers every edge and every rejected
  edge. `SelfEnroll` builds `source=self,state=active,order=nil`. `NewGrant` validates the widened
  source/state sets and keeps the `purchase ⇔ hasOrder` invariant.
- `TestActiveGrantOfEverySourceBlocksAnotherOrder` now includes `self`, so an active self membership
  blocks a new order at the domain level (anti-double-order). **Note:** this still only exercises the
  *active* case — the plan's stronger rule (blocked/revoked rows must also stop a 2nd order, no
  mute-bypass by paying) is T2 and must be tested there.
- Migration internals are self-consistent: the full `UNIQUE(user_id,course_id)` is added before the
  backfill, and the backfill + seed both use the matching full `ON CONFLICT (user_id,course_id)` arbiter.
  The DOWN is ordered correctly — convert `blocked→revoked` and delete `source='self'` *before*
  narrowing the CHECKs and restoring the partial index, so no constraint fails on the way down.
- `participant_reports` matches plan §5 (anti-self-report CHECK, partial-unique open dedup).
- Architecture: transitions are pure domain methods with a domain error; no framework/DB leakage; the
  guarded-method state machine is the right amount of structure (no over-engineered State pattern). The
  only asymmetry is non-blocking #2.

---

## Scope 1 — accounts (merged, retroactive): NOT YET REVIEWED

166 files / ~9.6k lines. Starting after L1 so Courses 1 is unblocked on enrollment first. Findings will
go to `work/review-courses2-accounts.md`.

---

## L2a (T2, purchase-side) — CanStartOrder + settle UPSERT + Membership — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-06), 8 files under `internal/orders`.

**Verdict: APPROVE — 0 blocking, 3 non-blocking.** All four points Courses 1 asked about check out
against the code. Full `make integration` (orders+course+playback) was run green — the L1 lesson about
the full suite was applied.

### Answers to the review asks (verified against code)

- **`CanStartOrder(*Grant)` vs plan §4 — correct.** `membership == nil || (state==active && source==self)
  → ok; else course_already_owned` (`grant.go`). `&&` binds tighter than `||`, so it is `nil ||
  (active ∧ self)`. The unit table covers all 12 state×source combinations plus the `nil` case
  (`grant_test.go:96-130`): active-self → ok; active purchase/admin/grandfather, every blocked, every
  revoked → owned. Matches §4 exactly (self-active is purchasable as an upgrade; any other row, including
  revoked, blocks the front door).
- **Settle UPSERT WHERE-branch — correct.** `ON CONFLICT (user_id,course_id) DO UPDATE SET
  source='purchase', order_id=EXCLUDED.order_id WHERE course_grants.source='self' AND state='active'`.
  No row → plain INSERT (grant). Self-active row → upgrade in place (same id, source→purchase, order
  attached). Any other existing state (purchase/admin/grandfather active, any blocked, any revoked) →
  `WHERE` false → 0 rows → `pgx.ErrNoRows`. Nothing else is ever upgraded, so a paid event cannot
  unblock a mute or resurrect a revoke.
- **Money-slipped safety net is intact.** The order is set `status='paid'` before the grant upsert, but
  the `ErrNoRows` branch sets `refund_required=true` and warns, in the *same transaction*
  (`repository.go:284-289`). So a purchase that lands on a conflicting membership leaves the order paid
  **and** flagged for refund — no paid-without-grant-without-refund hole. `CanStartOrder` is the front
  door; this is the race safety net (membership changed between order-start and settle). Both are needed
  and both are right.
- **insert-vs-upgrade via `insertedGrantID != grantID` — correct in prod.** `newUUID()`
  (`cmd/api/main.go:191`) formats with `%x` → lowercase canonical, and Postgres `id::text` returns
  lowercase canonical, so on a fresh INSERT `insertedGrantID == grantID` exactly, and an upgrade returns
  the pre-existing (different) id → `action='upgrade'`. See non-blocking #1 for the fragility.
- **`Membership` NewGrant-fail → `ErrDependencyUnavailable` — defensible.** A `NewGrant` failure here
  means the stored row violates a domain invariant the DB CHECKs are supposed to prevent (a "can't
  happen" / corruption), so a 503 "backend is in a bad state" is reasonable rather than 404/400. Fine.

### Non-blocking

1. **`insertedGrantID != grantID` couples to the lowercase-canonical id invariant.** It is correct today
   (verified `newUUID`), but it is a string comparison between a client-supplied id and a
   DB-round-tripped `id::text`; if an id provider ever emitted uppercase/non-canonical, a fresh insert
   would be mislabelled `upgrade`. Only the audit action label is at stake (the grant row is correct
   either way). The robust idiom is `RETURNING id::text, (xmax = 0) AS inserted` — `xmax = 0` is true
   for the inserted tuple, false for the updated one — which needs no assumption about id formatting.
2. **`ActiveGrants` is now dead code.** After `StartPurchase` moved to `Membership`, `ActiveGrants` has
   no production caller — only its port declaration (`ports/repository.go:34`), its impl
   (`repository.go:100`) and the fakes remain. Remove it (and from the fakes), or note why it is kept
   for a later slice. YAGNI / dead flexibility.
3. **`found`-bool + pointer dance in `StartPurchase`** (`service.go`): `Membership` returns
   `(Grant, bool, error)`, then the service rebuilds a `*Grant` from the bool. Minor; a repository that
   returned `(*Grant, error)` directly would remove the conversion. Cosmetic.

### Architecture (clean)

Pure domain predicate (`CanStartOrder`), port method (`Membership`), UPSERT in the adapter — layering
intact. The `*Grant`-as-optional membership is idiomatic. No GoF misuse; the only tidy-up is the dead
`ActiveGrants` (#2).

---

## L2b (T2, state application + authz) — enroll/block/unblock/revoke/reinstate — **CHANGES REQUESTED**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-06), 9 files `internal/orders` + plan §4.

**Verdict: CHANGES — 1 blocking, 3 non-blocking.** Every rule Courses 1 asked about is correct
(authz, the free/paid/admin reinstate guard, the no-op transitions, the new email gate). The blocking
item is the one I flagged ahead of the slice: the read-modify-write transition is not concurrency-safe.

### Blocking

**B-L2b.1 — membership transitions lose updates under concurrency (unguarded compare-and-set).**
Every transition (GrantCourse, Enroll, `transitionMembership`) does: (1) `grants.Membership()` — a read
on `repository.pool` in its own transaction, no `FOR UPDATE`; (2) a domain decision; (3)
`UpdateMembership` → `UPDATE course_grants SET … WHERE id=$1` in a *separate* transaction, with **no
expected-state predicate** (`repository.go` UpdateMembership). The read and the write are two
transactions with no lock and no optimistic guard.

Under the pool's READ COMMITTED isolation, two conflicting transitions on the same (user,course) row
serialize on the row lock but the second still **blindly overwrites** — a lost update. Concrete: admin
`Revoke` (active→revoked) concurrent with owner `Block` (active→blocked). Both read `active`; T1 commits
`revoked`; T2's `UPDATE … WHERE id` then re-reads the committed row and sets `blocked` — the row ends
`blocked`, the **revoke is silently lost**, yet `Revoke` returned success and wrote a `revoke` audit
row. Because `CanConsume` grants content to `state ∈ {active, blocked}`, a lost revoke leaves the user
with the content access the admin meant to remove — an authorization-integrity failure, and the audit
log disagrees with the row. The same shape loses `Unblock` vs `Revoke` on a `blocked` row.

This undermines the phase thesis (§0/§3: the single row's state *is* the access authority). The window
is narrow (two conflicting transitions interleaving on the same pair), so this is low-frequency, but
the consequence is a security-relevant state corruption, so I rate it blocking.

Fix (smallest that keeps the explicit, testable UPDATE): make the UPDATE a compare-and-set — thread the
expected prior state and use `UPDATE … SET … WHERE id=$1 AND state=$expectedOld`, treating
`RowsAffected()==0` as a distinct concurrency-conflict error (not the current `grant_invalid`). The
domain transition already knows the `from` state, so this is a small change. (Alternatively, read the
row `FOR UPDATE` inside the same transaction as the write; heavier, needs a transactional transition
method.) Add an integration test with two concurrent conflicting transitions asserting exactly one wins
and the other reports a conflict.

### Non-blocking

1. **The no-op check is itself TOCTOU (benign).** `transitionMembership` reads state and returns a
   no-op if already in the target state, but two concurrent `Block`s both read `active`, both pass the
   check, both UPDATE to `blocked`, both write a `block` audit row — same final state, duplicate audit.
   Fixing B-L2b.1 with a compare-and-set naturally collapses the second to a conflict/no-op.
2. **Concurrent `Enroll` loser gets `course_already_owned`, not the no-op the contract implies.** Two
   concurrent `none→create` enrolls: one INSERTs, the other hits the full unique →
   `ErrCourseAlreadyOwned`, even though `Enroll` on an existing active membership is documented as a
   no-op returning the membership. Consider mapping the unique violation on the self-enroll path to a
   re-read + no-op return. Minor.
3. **`UpdateMembership` 0-rows → `ErrGrantInvalid`** is a generic label for "row vanished"; once you add
   the expected-state predicate (B-L2b.1), 0 rows will more often mean a concurrency conflict — give it
   its own error so the two causes are distinguishable.

### Verified correct (not assumed)

- **authz** (`transitionMembership`): Block/Unblock → admin ∨ course owner (`actor.UserID ==
  instructor_id`); Revoke → admin only (`administratorOnly=true`). `MembershipState` → self ∨ admin ∨
  owner. Matches plan §4.
- **Reinstate guard (L1 #2), free vs paid vs admin:** `Enroll` requires `facts.Free && published`, so a
  **paid** revoked membership returns `course_not_enrollable` before any state read; only admin
  `GrantCourse` (no free gate) reinstates a paid revoked via `Reinstate(admin)`; a **free** revoked
  membership self-reinstates via `Reinstate(self)`. Exactly §4.
- **Reinstate domain method:** guards `state != revoked → ErrGrantTransitionInvalid`; builds the active
  grant with `orderID=nil`, so `Reinstate(purchase, …)` is structurally impossible (NewGrant's
  purchase⇔order invariant) — reinstatement is only self/admin, never purchase. Correct per §4.
- **no-op (L1 #4):** re-block/re-unblock/re-revoke return the current record with no write and no audit
  (at read time; see non-blocking #1 for the concurrent caveat).
- **email gate:** `Enroll` returns `ErrEmailNotVerified` first, before the course read — matches the
  new owner decision (enroll creates membership, so it mirrors StartPurchase's verified requirement).
- **GrantCourse revision:** none→create admin; revoked→Reinstate(admin); active|blocked→already_owned.
  Correct.
- Architecture: shared `transitionMembership` + method-expressions (`domain.Grant.Block` etc.) is a
  clean, DRY dispatch; the explicit Create/Update split keeps transitions visible. No GoF misuse. The
  only defect is the concurrency gap (B-L2b.1), which is a correctness issue, not a structure one.

### L2b re-check (B-L2b.1 fix) — **APPROVE**

Verified the compare-and-set fix against the code:
- `UpdateMembership(grant, expected, …)` → `UPDATE … WHERE id=$1 AND state=$expected`;
  `RowsAffected()!=1 → domain.ErrMembershipConflict` (its own code, not `grant_invalid`). The
  lost-update window is documented in a code comment.
- All three callers pass the correct prior state: `transitionMembership` → `membership.State()`
  (service.go:267); Enroll-reinstate and GrantCourse-reinstate → `GrantRevoked` (service.go:207,167),
  which equals the read state because both are guarded by a `state == revoked` check first.
- The deterministic integration test `L2b compare-and-set rejects a stale concurrent transition`
  reproduces the exact window: one `snapshot` read as active, a winning `Block(expected=active)`, then
  a stale `Revoke(expected=active)` on the now-blocked row → `ErrMembershipConflict`; asserts final
  state `blocked` **and** `audits==1`. That proves exactly-one-wins, no blind overwrite, and no phantom
  audit — a cleaner proof than a goroutine race, which could miss the window.
- Non-blocking #1 (no-op TOCTOU double-audit) collapses via the CAS — the `audits==1` assertion proves
  the conflicting attempt writes no audit row. #2 (Enroll create-loser) now re-reads and returns the
  existing row (documented no-op), consistent with no-op-on-existing. #3 (0-rows own error) done.

All three blocking/non-blocking items I raised are resolved. **L2b: APPROVE**, conditional only on the
full `make integration` run Courses 1 has in flight.

Forward note (not L2b): when the enroll/block/unblock HTTP endpoints land, map `membership_conflict`
to 409 so a losing concurrent transition surfaces as a conflict, not a 400/500.

---

## L2c (T2, roster + progress) — Roster read + keyset + authz — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-06), 7 files `internal/orders`.

**Verdict: APPROVE — 0 blocking, 1 non-blocking.** All five focus areas verified against the code and
the integration test, which exercises each on real Postgres.

### Verified (evidence)

- **percent = review parity.** `CASE WHEN total=0 THEN 0 ELSE (completed*100 + total/2)/total END`
  (`repository.go` Roster) — the round-half-up my L2c note asked for, identical to
  `discussion/domain/review.go:75`. completed/total are CROSS JOIN LATERAL aggregates
  (`lesson_progress.completed_at IS NOT NULL`; `course_lessons` via `course_modules`), mirroring
  `review.go`, so no row multiplication. En08 asserts the exact regression case: a blocked member at
  2/3 → **67** (truncation would give 66), plus 3/3 → 100 and 0-lessons → 0.
- **keyset / last page.** `ORDER BY g.granted_at DESC, g.user_id ASC`; predicate
  `$2 IS NULL OR granted_at < $2 OR (granted_at = $2 AND user_id > $3)` — exactly the strict-after set
  for that mixed DESC/ASC order. `limit+1` fetch; `len > limit → NextCursor = encode(last)`, else `""`
  (B-28). The cursor's user_id is validated as a uuid on decode (`pgtype.UUID.Scan`) → bad cursor →
  `ErrInvalidCursor`. En08 walks three members across three keyset pages (1+1+1) and asserts the final
  page's `NextCursor == ""`; the empty-course case returns 0 entries and `""`.
- **authz owner∨admin, no self.** `!Administrator && UserID != InstructorID → ErrEnrollmentForbidden`.
  The unit table covers owner, admin, a stranger (forbidden), and En08 additionally asserts a member
  acting on their own course's roster is forbidden (no self). Limit clamp: `<=0 → 20`, `>100 → 100`,
  tested at -1/1/101.
- **revoked excluded.** `WHERE g.state IN ('active','blocked')`. En08 seeds a revoked member and
  asserts it appears in no page. `scanRosterEntry` also defensively rejects any other scanned state →
  `ErrDependencyUnavailable`.
- **email absence.** Not in `ports.RosterEntry`, not selected in the SQL. Confirmed.

### Non-blocking

1. **Soft-deleted users still appear in the roster with their real display_name.** The SQL inner-joins
   `users u ON u.id = g.user_id` with no `u.deleted_at IS NULL` filter, and a membership row is not
   removed when a user soft-deletes. So a since-deleted but still-enrolled account shows in the owner's
   roster under its real name. No crash (`display_name` is `NOT NULL`), and the view is owner/admin
   only — but the review list (K-phase) deliberately filtered `deleted_at IS NULL` and showed «Колишній
   учасник» for deleted authors. Decide whether the roster should match that (filter or anonymize
   deleted users), for consistency and to avoid surfacing a deleted account's name. Low severity.

### Architecture

Cross-module read of `lesson_progress`/`course_lessons` by direct SQL (no module import) is the
ADR-013/014 precedent (course reads `discussion_reviews`; four modules read `course_grants`). LATERAL
aggregates, keyset mirroring `moderation.go`, clamp in the application, defensive state scan — all
consistent with the codebase. No GoF misuse, no over-engineering. L2 (purchase + state + roster) is
complete; next is L3 (contracts/access predicates + the module JOIN updates), which is the
cross-cutting one — I'll want the full e2e there.

---

## L3a (contracts/access predicates) — CanConsume / CanCommunicate — **CHANGES REQUESTED**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-06); diff = `access.go` + `access_test.go`.

**Verdict: CHANGES — 1 blocking (predicate composition), 0 non-blocking.** Additive, nothing existing
touched (CanView/CanStream/accessExempt/PurchaseRequired/Entitled and the AST + truth-table all intact
— confirmed by `go test ./platform/architecture/...` green). CanCommunicate is correct. The one issue
is CanConsume's preview disjunct, and it is best fixed here, before L3b wires 7 readers onto it.

### Blocking

**B-L3a.1 — `CanConsume`'s `|| course.Preview` has no `Published` guard → widens access vs the old gate.**
```
CanConsume = admin ∨ owner ∨ course.Preview ∨ Membership∈{active,blocked}
```
The old content gate is `CanStream = CanView && EmailVerified`, and
`CanView = accessExempt ∨ (Published ∧ (!Paid ∨ Preview))` — so a **stranger** on a preview lesson gets
access only when the course is **Published**. `CanConsume`'s bare `|| course.Preview` drops that guard,
so a preview lesson of a **draft or archived** course becomes consumable by anyone. The unit test
encodes this as intended: every case builds `Course{…, Preview: preview}` with `Published` left false,
and asserts `preview=true ⇒ CanConsume=true`. When L4 flips the content gate from `CanStream` to
`CanConsume`, this is a silent authorization widening — preview bytes of unpublished courses become
reachable — propagated to all 7 readers, and sharpest on the media-bytes `access.go` (`auth_request`)
path if it gates only on `CanConsume`.

Fix: guard the preview disjunct with Published, matching `CanView`:
```
CanConsume = admin ∨ owner ∨ (course.Published ∧ course.Preview) ∨ Membership∈{active,blocked}
```
The **membership** disjunct stays Published-independent on purpose — that matches the old
Entitled-via-accessExempt behaviour, so an active/blocked member of an **archived** course still
consumes (G3). Only the stranger-preview-of-unpublished case changes, and that is exactly the widening
to prevent. Add a `Published` dimension to the test matrix and assert
`CanConsume(stranger, {Published:false, Preview:true}) == false` (and `{Published:true, Preview:true}
== true`).

### Confirmed (answers to the two explicit asks)

- **"Preview makes CanConsume true regardless of membership"** — yes for a **published** course's
  preview lesson (that is public, same as CanView). But NOT regardless of `Published`: see B-L3a.1.
  With the `Published ∧ Preview` fix, preview still grants consume independent of membership on a
  published course, which is the intent.
- **Email stays a separate L4 gate** — agreed. Keeping `email_not_verified` distinct from
  `enrollment_required`/`enrollment_blocked` needs email out of these predicates; an application-layer
  check in L4 is the right seam.
- **CanCommunicate** = `admin ∨ owner ∨ Membership==active` — correct: blocked=false, revoked=false,
  none=false, active=true. Matches plan §4. The `UserID != "" && == OwnerID` owner check guards the
  anonymous-empty-owner case. No Published guard needed (owner/admin on their own draft is fine;
  a non-member can't communicate regardless).

Fix B-L3a.1 and I'll re-check just the predicate + the added Published test rows, then you commit and
wire L3b.

### L3a re-check (B-L3a.1 fix) — **APPROVE**

`CanConsume = admin ∨ owner ∨ (course.Published && course.Preview) ∨ Membership∈{active,blocked}`.
The preview disjunct is now Published-guarded (mirrors `CanView`'s `Published && (!Paid||Preview)`), so a
preview lesson of a draft/archived course is no longer consumable by a stranger; the membership
disjunct stays Published-independent on purpose (archived member / G3). The test matrix gained a
`published` dimension (membership×role×preview×published) and its computed `wantConsume` uses
`(published && preview)`, so the widening cell — none/stranger/preview=true/published=false → false — is
asserted, along with its published=true → true counterpart. Architecture test green (CanView/CanStream/
Entitled + AST + truth-table untouched). B-L3a.1 resolved. **L3a: APPROVE** — commit and wire L3b.

(L3b reminder, already in your brief: each of the 7 readers must derive Entitled as an explicit
`state='active'` — not a surviving `id IS NOT NULL` — and the blocked+revoked integration fixture must
assert Entitled=false on every reader, since today's e2e has no non-active members to catch a miss.)

---

## L3b (7 readers carry membership state; no gating) — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-06); 17 files across course/playback/
transcript/discussion.

**Verdict: APPROVE — 0 blocking, 0 non-blocking.** Both guards I insisted on are present and verified
on the production path. This was the cross-cutting slice; it holds.

### Guard (1) — explicit `state='active'`, no surviving proxy — VERIFIED on all 7 readers

- `grep "state='active'"` over non-test reader SQL in all four modules → **empty**: no reader JOIN
  still filters on active, so the full membership state is read (needed for L4).
- `grep "grants.id IS NOT NULL"` non-test → **empty**: the old active-proxy is gone everywhere.
- Every reader derives `Entitled = membership.Valid && membership.String == "active"` via a shared
  `membershipState(pgtype.Text)→access.Membership` helper: course learning, playback media-bytes +
  session, transcript, discussion lessonFacts + commentQuery + reviewFacts. The reviewFacts EXISTS form
  became `LEFT JOIN course_grants g ON g.course_id AND g.user_id` + `SELECT g.state`; under the full
  `UNIQUE(user,course)` that LEFT JOIN returns at most one row, and completed/total are correlated
  subqueries, so there is no review-row fan-out and the progress counts are unaffected.

### Guard (2) — blocked+revoked fixtures assert `Entitled=false` per reader — VERIFIED

Integration subtests loop `{active, blocked, revoked}` and assert `Entitled == (want==active)` and the
right `Membership`, on each of the seven reader paths (course learning_integration_test; playback
access_integration_test:203 bytes + :207 session; transcript repository_integration_test:217;
discussion membership subtest :299 lessonFacts, :303 commentQuery, :327 reviewFacts). This is the fixture
I insisted on — it runs on real Postgres, so it catches a missed conversion that today's e2e (no
non-active members) cannot.

### No behaviour change (true plumbing)

- `CanConsume`/`CanCommunicate` have **no production caller** (grep empty outside their definition) —
  gating is deferred to L4, so L3b changes no behaviour. `Entitled` still means `state=='active'`, so
  every existing gate (CanView/CanStream/CanBeViewedBy/PurchaseRequired) sees an identical value.
- The 4 `CanBeViewedBy` bodies are unchanged — each `return access.CanView(...)` with only a `Membership`
  field added to the Course literal (CanView ignores Membership). `platform/architecture` (AST +
  truth-table) green confirms CanView/CanStream/Entitled and the 4 CanBeViewedBy are untouched.
- Bypass-consumers `review.go CanReview` and course `CanPurchaseBy` were not touched and remain correct
  on `Entitled=active`.
- The two Codex test-defect fixes (shared-lesson comment counted by K09; shared metric counter) are
  fixture-isolation, not conversion changes — the "whoever adds a globally-counted thing isolates it"
  rule, correctly applied.

### L4 forward note

L4 flips the content gate to `CanConsume` and the communication gate to `CanCommunicate` — a real
behaviour change (free content starts requiring membership; blocked loses communication). That's where
the L3a preview-Published fix matters (it's in now) and where the email gate lands as a separate code.
Run the full e2e, and seed the e2e fixtures with at least one blocked and one revoked member so the new
denials are actually exercised (same reason as guard 2 — a green e2e with only active members wouldn't
prove the gate denies).

---

## L4a (content gate → CanConsume + email; first behaviour change) — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-06); playback (authorize + sessions) +
transcript read + e2e/seed fixtures.

**Verdict: APPROVE — 0 blocking, 3 non-blocking.** The gate flip is correct on all three content sites,
the email split is necessary and right, and guard-rail #1 held — every red e2e was fixed by enrolling a
fixture or seeding a non-member, not by weakening the predicate. Full integration + full e2e (101)
green.

### Verified (evidence)

- **Gate conversion, all 3 sites identical and correct** (`authorize.go`, `sessions.go:74`,
  `transcript/read.go:60`):
  ```
  if !CanConsumeBy(actor) { if PurchaseRequiredBy(actor) { return ErrPurchaseRequired }; return ErrForbidden }
  if !actor.EmailVerified { return ErrEmailNotVerified }
  ```
  Denial matrix (domain + application tests): active/blocked → consume; revoked/free-non-member →
  forbidden; paid-non-member → purchase; preview(Published) → consume; owner/admin → consume;
  member-unverified and blocked-unverified → email. Matches CanConsume + the content rule.
- **Email split is necessary and correct.** Keeping `CanStreamBy` (= CanView && email) would deny a
  blocked **paid** member even when verified, because CanView is false for them (Entitled=state=='active'
  → blocked false; paid, no preview). Splitting email into a pure `if !actor.EmailVerified` lets a
  verified blocked member consume (the whole point of blocked = muted-but-watching). Right call.
- **Guard-rail #1 upheld.** The only gate code is the clean `CanBeViewedBy→CanConsumeBy` + email swap;
  the e2e/seed diffs are fixtures: `seed-identities.sql` adds `outsider@courseforge.local` (verified,
  deliberately never enrolled) as the non-member fixture; playback/transcript specs enrol their student
  via the admin-grant endpoint; R04→En09 (enrolled member plays an archived course; the "cut" = admin
  revoke deferred to L6, noted in a comment); R34 uses the seeded `outsider`. No predicate was loosened
  and no gate was special-cased to turn a red spec green.
- Api-zone 429 handling in the ratings afterAll is a retry, and recovery:102 (which intentionally tests
  the 429) confirms the zone was NOT relaxed — correct (relaxing it to fix leftovers would have broken
  the deliberate rate-limit test).

### Non-blocking

1. **Content checks email AFTER membership/purchase, so an unverified NON-member gets
   forbidden/purchase, not `email_not_verified` — and this will be inconsistent with the comms gate
   (email-first).** Old behaviour: an unverified user on free content hit `CanStreamBy`
   (=CanView&&email) → `email_not_verified`. New: `CanConsumeBy`=false (not a member) → `ErrForbidden`
   first. Since **Enroll itself requires `email_verified`** (L2b), an unverified non-member is now told
   "enroll" → clicks enroll → gets `email_not_verified` — a two-step discovery, where email-first would
   be one step, and L4b's comms gate is email-first. Recommend moving `if !actor.EmailVerified` ABOVE
   the CanConsume gate on content too, so email-first is consistent everywhere and actionable; or record
   the asymmetry as deliberate. The free-non-member tests all use `EmailVerified:true`, so the current
   order is unpinned — add a free-non-member-**unverified** case asserting whichever order you choose.
2. **revoked PAID member → `ErrPurchaseRequired` on content, but `StartPurchase` then refuses with
   `course_already_owned`** (CanStartOrder blocks a revoked row) — a dead-end: the content gate says
   "purchase", the purchase path says "already owned". Edge (admin-revoked paid member); the denial
   itself is correct. Consider surfacing a clearer reason for a revoked member, or note it.
3. **`access.CanStream` is now dead code** — no caller after the `CanStreamBy` wrappers were removed
   (only its own definition remains). Remove it (like `ActiveGrants` in L2a), or note why it's kept.

### L4b forward

Comms gate (CanCommunicate + email, order email→required→blocked): keep the email-first ordering you
stated, and reconcile with non-blocking #1 — ideally both gates check email first. Seed a blocked and a
revoked member into the comms e2e so `enrollment_blocked` and the active-only rule are actually
exercised (same reason as the content fixtures). I'll check the deny paths and the precedence asserts
(unverified+blocked → email; verified blocked → enrollment_blocked; verified non-member →
enrollment_required).

---

## L4b (communication gate → CanCommunicate + email) — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-07); discussion + OpenAPI + frontend.

**Verdict: APPROVE — 0 blocking, 1 non-blocking.** Every checklist item verified; the one finding is a
read-gate inconsistency against plan §4 (restrictive direction, not a security hole).

### Verified (evidence)

- **email-first in all 4 write paths.** `Publish` (:146), `ReportComment` (:246), `PutReview` (:369),
  `ReportReview` (:411) each check `if !actor.EmailVerified → ErrEmailNotVerified` **before** loading
  facts and calling `requireCommunication`. Consistent with the L4a content gate.
- **`requireCommunication` on all writes** = `CanCommunicateBy` (admin∨owner∨active) → ok; else
  `CommunicationBlocked()` (Membership==blocked) → `ErrEnrollmentBlocked`; else → `ErrEnrollmentRequired`.
  Order email → blocked → required matches the plan.
- **CanReview bypass fixed (real bug closed).** `PutReview` runs `requireCommunication` **before**
  `CanReview` (:377 before :381), so a blocked member with progress on a free course is stopped with
  `enrollment_blocked` and can no longer write a review — the pre-L4b hole. Owner still falls through
  communication (owner can communicate) to `CanReview` → `review_owner_forbidden`; an admin non-member
  passes communication (admin-exempt) but `CanReview` still denies (G2: role isn't a review right) —
  both gates correctly applied in order.
- **ReportReview now gated** (it had none before): email → `GetReview` → `ReviewFacts` → course literal
  **with Membership** → `requireCommunication`. A non-member reporting a review now gets
  `enrollment_required`; this is a tightening, consistent with the plan putting ReportReview under
  CanCommunicate.
- **Three precedence asserts explicit** (`review_test.go`): `unverified blocked → ErrEmailNotVerified`
  (:182), `verified blocked → ErrEnrollmentBlocked` (:222), `verified none → ErrEnrollmentRequired`
  (:223), active → ok (:224), owner-after-active → review_owner_forbidden (:225). Not inferred.
- **Integration proves blocked/revoked on the real path** (not inference): active Publish ok;
  blocked→enrollment_blocked; revoked→enrollment_required; blocked List reads; non-member List→forbidden,
  seeded by direct grant-insert + the application service.
- **Guard-rail #1 upheld:** e2e breaks fixed by enrolling fixtures (discussion-commenter, ratings
  seedProgress now grants, instructor publishCourse grants the reporter, Rt09 reporter-grant), not by
  loosening a gate. OpenAPI gains enrollment_required/blocked (+2 enum), Orval regenerated,
  check-generated clean. `make integration` + `make e2e-local` (101) green.

### Non-blocking

1. **`LearningReviews` reads via `CanBeViewedBy` (CanView), not `CanConsume` — inconsistent with the
   comment List and with plan §4.** The comment list (:122) was converted to `CanConsumeBy` (a blocked
   member reads comments), but the learning-scoped review list `LearningReviews` (:317) still gates on
   `CanBeViewedBy`, and its course literal (:315) doesn't even carry `Membership`. Consequence: a blocked
   member of a **paid or archived** course can read comments and consume content but is refused the
   review list (CanView is false for a blocked member there). Plan §4 is explicit — "перегляд
   коментарів/**відгуків** — це споживання (CanConsume)". The direction is restrictive (denies, no
   security widening), so non-blocking, but it's an incomplete conversion in the slice that owns
   "reading discussion = CanConsume". Fix: thread `Membership: facts.Facts.Membership` into the
   `LearningReviews` course literal and gate on `CanConsumeBy` (2 lines), and add a "blocked member reads
   the review list" integration assertion. (`PublicReviews` correctly stays CanView — that's the
   anonymous public catalog path.) Fix in L4b or record why LearningReviews deliberately stays CanView.

---

## L5 (person-reports, backend) — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-07); discussion domain + adapter +
application + tests.

**Verdict: APPROVE — 0 blocking, 1 trivial non-blocking.** All three focus points hold.

### Verified (evidence)

- **Self-report & validation (domain).** `NewParticipantReport` rejects `reporterID==targetID` →
  `ErrReportSelf` first, then empty ids / kind∉{on_lecturer,on_student} / reason(trim+utf8+1..500) →
  `ErrReportInvalid`. `IsParticipant = active||blocked` (G6: a blocked member may still report).
- **Membership read via discussion's own SQL — ADR-013 held.** `ReportSubject` =
  `SELECT c.instructor_id, g.state FROM courses c LEFT JOIN course_grants g ON g.course_id=c.id AND
  g.user_id=$2 WHERE c.id=$1` — reads the owner and the user's membership state directly, no
  `internal/orders` import (grep over discussion non-test → clean). No `state='active'` filter, so the
  real state (incl. blocked) is read; full UNIQUE → one row, no fan-out; missing grant → `none`.
- **Dedup via the partial unique.** `InsertParticipantReport` uses
  `ON CONFLICT (course_id,reporter_id,target_id) WHERE state='open' DO NOTHING`, matching the
  `participant_reports_one_open_per_pair` partial index from 00015; `RowsAffected()==0 →
  report_already_open`. Correct, and because the arbiter is `WHERE state='open'`, a **resolved** report
  does not block a fresh report of the same pair — the intended behaviour. (This is the same
  partial-arbiter shape that bit L1; here the index exists and the predicate matches, so it's sound.)
- **Authz.** `ReportLecturer`: email-first; actor==owner → `ErrReportSelf`; non-participant →
  `report_forbidden`; reports the owner as `on_lecturer`. `ReportStudent`: email-first; `!Administrator
  && actor!=owner → report_forbidden` (admin∨owner); target must be a participant; reporting the owner
  via this path is refused (owner has no membership → not a participant). `ReportQueue` / `ResolveReport`:
  **admin-only** (`!Administrator → report_forbidden`) — the focus item, confirmed.
- **Queue read:** keyset `(created_at DESC, id DESC)`, open-only, NULL-guarded first page;
  reporter/target display names via `LEFT JOIN users … deleted_at IS NULL` (deleted → anonymized, like
  the roster); **email not selected**. `ResolveParticipantReport` is idempotent
  (`WHERE id=$1 AND state='open'`).
- **Integration on the production path** (isolated UUID fixtures, not the shared d1/b1): participant →
  on_lecturer; blocked → reports; non-member → forbidden; duplicate → report_already_open; owner →
  on_student on a member, non-member target → forbidden; admin Queue sees open not resolved; Resolve
  hides; non-admin Queue/Resolve → forbidden. `make integration` green.

### Non-blocking (trivial)

1. `ResolveParticipantReport` returns nil for an unknown reportID as well as an already-resolved one
   (`WHERE id AND state='open'` → 0 rows, ignored). Idempotent-by-design is fine for a moderation
   resolve, but a genuinely missing id silently succeeds; if you want admins to notice a bad id,
   distinguish 0-rows → `ErrNotFound`. Optional.

---

## L6a (orders HTTP — wires the L2 methods) — **APPROVE**

Commit: uncommitted working tree on `feat/enrollment` (2026-10-07); orders/adapters/http + openapi +
main.go + generated Orval.

**Verdict: APPROVE — 0 blocking, 1 non-blocking.** The six routes are wired with correct middleware,
the error codes map as intended (including my membership_conflict→409 note), the roster carries no
email, and handlers are thin with authz left in the application.

### Verified (evidence)

- **Routes + middleware:** `POST /learn/courses/{id}/enroll`, `/teach/.../block`, `/unblock`,
  `/admin/.../revoke` are `sameOrigin(authenticate(...))` (writes, CSRF-guarded); `GET
  /learn/courses/{id}/membership` and `GET /teach/courses/{id}/roster` are `authenticate` only (reads).
  Handlers extract actor + path params, call the service, render — authz (owner∨admin / admin-only)
  stays in the application (Block/Unblock/Revoke/Roster), as designed.
- **Error mapping (writeError):** `ErrMembershipConflict → 409 membership_conflict` (the concurrent-
  transition code I asked for), `ErrCourseNotEnrollable → 409`, `ErrEnrollmentForbidden → 403`. Added to
  the OpenAPI error enum; Orval regenerated (additive client fns + membership/roster models),
  check-generated clean, frontend `tsc` clean.
- **No email in the roster DTO:** `rosterEntryDTO` = {UserID, DisplayName, State, EnrolledAt,
  ProgressPercent} — no email field, and the mapping (`handler.go:148`) doesn't select one; the
  membership DTO is {state (nullable), member}. `assertContract` validates every response against the
  OpenAPI schema.
- **Contract tests** cover the surface: enroll (verified) success; enroll unverified → 403
  email_not_verified; paid course enroll → 409 course_not_enrollable; GET membership for a member AND
  for none (nullable state, both contract-validated); roster entry shape (no email). `make integration`
  green; orders unit + frontend typecheck + check-generated clean. e2e untouched (new endpoints, no
  existing spec calls them; browser flows are L7/L8).

### Non-blocking

1. **`/admin/.../revoke` is not behind the `administrator()` middleware** that `grantCourse` uses — it
   relies solely on `application.Revoke`'s admin check (`administratorOnly=true → ErrEnrollmentForbidden`
   → 403). Correctly enforced (not a hole), but asymmetric with `grantCourse` and with the `/admin/`
   path convention. Either wrap revoke in `administrator()` for edge-level defense-in-depth + consistency,
   or treat "authz in application" as the uniform rule (then grantCourse's middleware is the outlier).
   Minor.

---

## L6b (discussion person-report HTTP) — **APPROVE**

Commit: 789f004 on `feat/enrollment`. discussion http + a narrow fix to L5's repository/application +
OpenAPI/Orval.

**Verdict: APPROVE — 0 blocking, 0 new non-blocking.** The dedup refactor Courses 1 flagged is correct,
the routes/middleware/codes are right, and no DTO carries email.

### Re-checked: the L5 touch (the explicit ask)

- `InsertParticipantReport` now `(string, error)`:
  `INSERT … VALUES(COALESCE(NULLIF($1,'')::uuid, gen_random_uuid()), …) ON CONFLICT (course_id,
  reporter_id,target_id) WHERE state='open' DO NOTHING RETURNING id::text`. On a conflict the DO NOTHING
  suppresses the row → `QueryRow().Scan` → `pgx.ErrNoRows` → `ErrReportDuplicate`. `ErrNoRows ⟺ conflict`
  (RETURNING always yields the row otherwise), so no false-positive dedup. `ErrReportDuplicate` is still
  the string `report_already_open` (errors.go:26) → unchanged contract. A **resolved** report still does
  not block a fresh one (the arbiter is `WHERE state='open'`). The dedup path is sound after the
  RowsAffected→ErrNoRows swap.
- The application now does `id, err := InsertParticipantReport(...); report.ID = id` in both
  `ReportLecturer` and `ReportStudent`, so the 201 carries the real id. Mirrors `CreateReview`. Fakes +
  the application test (asserts the returned UUID) updated.

### HTTP

- Routes + middleware: `POST /learn/.../report-lecturer`, `POST /teach/.../participants/{uid}/report`
  are `sameOrigin(authenticate)` (authz admin∨owner / participant in the application). `RegisterAdmin`:
  `GET /admin/participant-reports` = `authenticate(administrator(...))`, `POST
  /admin/participant-reports/{id}/resolve` = `sameOrigin(authenticate(administrator(...)))`. So the
  categorically-admin routes carry the `administrator()` middleware **and** the application still checks
  `!Administrator → report_forbidden` (L5) — genuine defense-in-depth, not middleware-only. This is the
  resolution of my L6a asymmetry note, applied correctly for L6b.
- Error mapping: `ErrReportForbidden`/`ErrReportSelf` → 403; `ErrReportDuplicate` → 409;
  `ErrReportInvalid` → 400; course-not-found → 404; email_not_verified → 403. All present.
- DTOs carry **no email**: `participantReportDTO` {id,courseId,reporterId,targetId,kind,reason,state,
  createdAt}; the queue item embeds it + nullable `reporterName`/`targetName` (deleted users → null, the
  anonymization pattern) + `courseTitle`. OpenAPI queue item is a standalone strict object (so the
  queue-only fields pass `additionalProperties:false`); check-generated + frontend tsc green.

### Owning a miss (honest note)

The latent `report.ID == ""` that Codex caught at L6b **originated in L5, which I approved** (5347545):
the application passed `NewParticipantReport("", …)` and the repo generated the id in SQL
(`gen_random_uuid`) but returned only `error`, so the returned report's ID stayed empty. It was latent
at L5 (backend-only, no consumer read `report.ID`) and only bit when L6b's HTTP returns the DTO. I
reviewed that Insert for dedup/authz/ADR-013 but did not flag "the generated id is discarded → a future
HTTP 201 will carry an empty id." The forward-looking check — *when a write generates a value in SQL
(DEFAULT / gen_random_uuid) that a caller will later read, confirm the method returns it* — is the one I
should have applied. Noted for later slices.

L6 (a+b) complete pending commit of this verdict. Next is L7 (frontend), which Courses 1 is holding
until this sign-off.

---

## L6c (GET /learn/memberships — «Моє навчання» source) — **APPROVE**

Commit: 61e4070 on `feat/enrollment`. orders: port MyMemberships + MembershipCard, thin app/HTTP,
OpenAPI + Orval. A plan amendment (docs/enrollment/plan.md §Amendment 2026-10-07).

**Verdict: APPROVE — 0 blocking, 0 non-blocking.**

- **Own-data read, actor-scoped — no cross-user leak.** Handler takes the authenticated `actor`
  (`actorFromRequest`), application does `grants.MyMemberships(ctx, actor.UserID)`, SQL is
  `WHERE g.user_id = $1`. The user id is never a param/query. `emptyQuery` guard rejects any query
  string → 400; `authenticate` only (no sameOrigin needed for a GET, no email gate — reading your own
  enrolment list shouldn't require verification). `noStore`.
- **Progress parity.** The percent is the roster formula verbatim —
  `CASE WHEN total=0 THEN 0 ELSE (completed*100 + total/2)/total END`, completed/total via CROSS JOIN
  LATERAL — so «Моє навчання» matches the roster and the review percentages, no truncation.
- **State + publish semantics.** `WHERE g.state IN ('active','blocked')` (revoked excluded, consistent
  with the roster); plain `JOIN courses` with no publish filter, so an archived course you're a member
  of still appears (En09). `ORDER BY granted_at DESC LIMIT 200`, bounded no-pagination — a recorded
  amendment decision (own-membership count is bounded).
- **No email** in `membershipCardDTO` {courseId, slug, title, state, progressPercent, enrolledAt}.
- Contract tests validate the populated page (progressPercent, no email) and the empty page against
  OpenAPI; an application unit test asserts actor-delegation. The real Docker integration caught a
  fixture defect (a draft course seeded with a non-null published_at, violating courses_check3) that
  Codex's local-green run missed — the production query was correct; good gate discipline. make
  integration GATE_EXIT=0, check-generated + tsc green.

---

## L7a (learn-side enrollment frontend) — **APPROVE**

Commit: 92f618f on `feat/enrollment`. Greenfield enrollment FE layer + CourseView CTA + «Моє навчання».

**Verdict: APPROVE — 0 blocking, 2 non-blocking.** All four focus points hold; I agree with the
`member`-vs-`state` fix.

### Verified

- **CTA branches on state, not the `member` flag (the fix) — I agree revoked = non-member.**
  `activeMember = state==='active' || state==='blocked'`; `blocked = state==='blocked'`;
  `canEnroll = !activeMember && course.price===null`. So: active → «Ви учасник» + open; blocked → «Ви
  учасник» + blocked note (blocked still consumes content, consistent with CanConsume); revoked →
  activeMember=false → «Записатися» for a free course. Branching on the backend `member` boolean (true
  whenever a row exists, incl. revoked) was the real bug — a revoked user saw «Ви учасник». Correct to
  read revoked as a non-member: access was removed, and a free self-re-enrol is the right next step
  (matches L2b).
- **409 → refresh/retry, not a dead error.** The enroll mutation `onSuccess` invalidates
  membership+myMemberships+course; `onError` on `membership_conflict` invalidates the membership query
  so the UI re-reads the concurrent state, and the component shows a `role="alert"` retryable message.
  Matches my membership_conflict→409 forward intent.
- **blocked/revoked states** handled (above). **«Моє навчання»** at `/learn` (redirect removed, splat
  player untouched) renders off `getMyCourseMemberships` with a state pill + ProgressBar + empty state;
  MembershipCard.state is `active|blocked` only (revoked excluded by the L6c query), so the list never
  shows a revoked course. **no fetch in presentation** — adapter(transport)/queries/presentation
  separation; the architecture lint over 170 files is green.
- **a11y:** blocked note is `<output>` (role=status), enroll error is `role="alert"`, controls use the
  shared 44px Button, i18n uk/en namespace + error-code messages. (Full keyboard/focus is the L8 e2e.)
- The Rstest gate caught the member-flag bug behind Codex's locally-green run (sandbox can't run
  Rstest) — a real authorization-display defect, and the CTA-transition test already encoded the
  correct expectation. Good gate discipline.

### Non-blocking

1. **`member: boolean` on the `Membership` domain type is a footgun** — it's true for revoked, which is
   exactly what caused the fixed bug. The component now correctly ignores it and branches on `state`.
   Consider dropping `member` from the domain type (state ∈ {active,blocked,revoked,null} is
   authoritative) so no future code can branch on the misleading flag. Defensive cleanup.
2. **Paid-revoked inherits the L4a dead-end.** A paid-course revoked user is not activeMember and not
   canEnroll (price≠null), so they see the purchase path, but a re-purchase after revoke is refused at
   the backend (`course_already_owned`, CanStartOrder). That's the recorded L4a limitation surfacing in
   the UI, not an L7a defect — just cross-referencing it so it's tracked when that limitation is
   revisited.

---

## L7b (instructor roster + block/unblock) — **CHANGES REQUESTED**

Commit: dbfc813 on `feat/enrollment`. enrollment feature + presentation/teach/CourseRoster.

**Verdict: CHANGES — 2 blocking, 0 non-blocking.** The data layer and states are right; both blockers
are small front-end fixes (one a11y, one 409-handling).

### Blocking

**B-L7b.1 — focus drops to `<body>` on two controls (`isDisabled` on the active control).** Same class
as the L4a/Rt14 fix.
- Row action (`CourseRoster.tsx:83`): `isDisabled={isPending}` with a single list-wide `isPending`
  flag — one block/unblock disables **every** row button, and the clicked (focused) button gets the
  native `disabled` attribute, so keyboard focus drops to `<body>` and the roster loses the user's
  place. Fix: per-row pending (track the acted userId) + `isPending={pendingUserId===entry.userId}` so
  only the acted button shows busy and it stays focusable.
- Load-more (`CourseRoster.tsx:98`): `isDisabled={isFetchingNextPage}` — this is the **same «Показати
  ще» focus-drop we fixed in ratings (L4a) with `isPending`**, reintroduced here. Fix: `isPending={
  isFetchingNextPage}` (and, as in Rt14, move focus to the first newly-loaded row). Courses 1 flagged
  the row button; the load-more instance is the second one.

**B-L7b.2 — block/unblock do not refresh the roster on a 409 (`membership_conflict`).** The `enroll`
mutation has `onError` → invalidate-membership on `membership_conflict`, but `block`/`unblock`
(`enrollment-queries.ts:60-73`) have only `onSuccess` → invalidate-roster; there is no `onError`. So a
concurrent-transition 409 (e.g. an admin revokes while the owner blocks — the CAS yields
`ErrMembershipConflict`) surfaces a `role="alert"` but leaves the **stale** roster showing the old
state, misleading the owner. This misses the "409 → refresh" behaviour baked into the brief. Fix:
add `onError` to block/unblock → if `membership_conflict`, invalidate the roster key (mirror `enroll`),
so the list re-reads the real current state.

### Verified correct

- **Keyset, not offset:** `infiniteQueryOptions` with `getNextPageParam: page.nextCursor ?? undefined`,
  queryFn passes `pageParam` (cursor) to `enrollment.roster` — the L6a keyset page. «Показати ще» is
  cursor-driven.
- **States:** ProgressBar (0 lessons → 0% from the backend CASE), state badge, `EmptyState`, query-error
  `role="alert"` + retry. block shown for active, unblock for blocked.
- **Authz:** block/unblock are backend-enforced (application Block/Unblock = admin∨owner → 403
  enrollment_forbidden); the roster renders on the owner∨admin teach route. The buttons don't need a
  client role-gate (backend is the authority).
- **L7a non-blocking #1 folded well:** `member` dropped from the `Membership` domain type (state ∈
  {active,blocked,revoked,null} authoritative), with a comment documenting the revoked-as-member
  footgun; the adapter validates `state` alone and ignores any extra `member` field. tsc confirms
  nothing branches on `member`.
- The brief-correction (removing the non-existent 404 map for block/unblock) is right — the generated
  client declares 403/409 only.

Fix B-L7b.1 (both buttons) and B-L7b.2; I'll re-check just those. The rest is commit-ready.

---

## L7b re-check + L7c1 (report forms) — **APPROVE**

Commit: 762161c on `feat/enrollment` (one commit: the B-L7b fixes + L7c1 landed in the same files).

**Verdict: APPROVE — 0 blocking, 0 non-blocking.** Both L7b blockers are resolved correctly and L7c1 is
clean.

### B-L7b re-check

- **B-L7b.1 (focus) — fixed on both controls.** Row action:
  `isPending={isPending && actedUserId === entry.userId}` (per-row, not list-wide `isDisabled`);
  `onPress` sets `actedUserId` then calls onBlock/onUnblock, so only the acted button is busy and stays
  focusable; the busy condition self-clears when `isPending` falls (no reset effect — correctly removed
  after oxlint flagged set-state-in-effect). Load-more: `isPending={isFetchingNextPage}` + focus moves
  to the first newly-loaded row (`ref`+`tabIndex=-1`, effect on `!hasNextPage`), mirroring Rt14. My
  forward note is satisfied: the report row action uses `ReportDialog`'s own per-instance pending, so all
  three row actions (block/unblock/report) are per-row and none disables the others.
- **B-L7b.2 (409 refresh) — fixed on both.** `block` and `unblock` mutations gained
  `onError → if membership_conflict → invalidate the roster key`; a non-conflict error leaves the cache
  untouched and surfaces the alert. New test asserts both paths.

### L7c1

- **ReportDialog:** reason trimmed, `valid = 1..500`, and the form `onSubmit` guards `if (!valid)
  return` (client-side, plus `isInvalid`/`errorMessage` on the field) — not relying on the backend alone.
  Submit is `isPending={pending}` (per-instance). Success → `role="status"`; every backend code
  (report_own_person / report_forbidden / report_already_open / report_invalid / email_not_verified /
  course_not_found) maps through the error switch to a `role="alert"` message — sane, not dead errors —
  and the alert renders inside the dialog.
- **CTA gating:** «Поскаржитися на викладача» renders only in the `activeMember` branch (state ∈
  {active, blocked}), so a member — including a blocked one (G6) — can report the lecturer and a
  non-member (revoked/none) cannot. The per-row «Поскаржитися» sits on the owner∨admin teach roster.
  email-first is the backend's (L5/L6b); the dialog surfaces `email_not_verified` if returned.
- **404 self-correction confirmed:** `course_not_found` added to the enrollment error union + mapped
  (report endpoints declare 404 — ReportLecturer/Student → ErrNotFound); block/unblock stay 403/409.

Gate: typecheck + lint/architecture + Rstest (629) green. Both my L7b items closed; L7c1 approved.
Next is L7c2 (admin participant-report queue + resolve + revoke).

---

## L7c2 (admin participant-report queue + resolve + revoke) — **APPROVE** · L7 complete

Commit: e92f8ce on `feat/enrollment`.

**Verdict: APPROVE — 0 blocking, 1 non-blocking.**

### Verified

- **No email; deleted anonymized.** Rows render kind, reporterName/targetName (`?? formerAuthor` for
  deleted), courseTitle, reason, createdAt, state — no email.
- **Per-row pending:** `isPending = actedReportId === item.id && (resolvePending || revokePending)` on
  both resolve and revoke buttons (isPending, not isDisabled) — only the acted row is busy and stays
  focusable. Load-more `isPending={isFetchingNextPage}` + focus moves to the first newly-loaded row.
  Same correct pattern as the fixed L7b.
- **Open-only keyset** infinite query; **revoke behind a confirm Dialog**; resolve is idempotent
  (backend); resolve/revoke are admin-only (backend + the admin-only route/nav). query-error and
  action-error are `role="alert"` with retry; `EmptyState`.
- **409 → refresh on both:** resolve invalidates the queue; revoke invalidates queue+roster on success,
  and `onError → membership_conflict → invalidate queue+roster`. So a concurrent conflict re-reads, not
  a dead error.

### Non-blocking

1. **«Відкликати доступ» is offered on `on_lecturer` reports, where it can't apply.** `revoke(item)`
   calls `onRevoke(item.courseId, item.targetId)`, and for an `on_lecturer` report the target is the
   course owner, who has no membership row → `application.Revoke` → `Membership==nil && action=='revoke'`
   → `ErrGrantInvalid`. So on a lecturer report the revoke button (after its confirm dialog) always
   fails with a grant_invalid alert — a confusing dead action. The backend safely refuses (no wrong
   state change), so it's non-blocking, but revoke only makes sense for `on_student` reports (the target
   is a revokable member). Gate the revoke Dialog on `item.kind === 'on_student'`; on an `on_lecturer`
   report only «Вирішити» applies. (Also confirm `grant_invalid` is in the enrollment error union if you
   keep the button, so the alert isn't a raw key.)

L7 (a + b + c1 + c2) complete. Next is L8 (e2e enrollment.spec + axe) — the real-browser gate where
behind-the-gate defects surface. My L8 focus: the deny paths actually exercised with seeded
blocked/revoked members (not happy-path only), media-bytes-blocked, backfill/single-row/grandfather/
preview/G10, axe on the new screens — and, per the L4 guard-rail, that no e2e red was turned green by
weakening a gate rather than seeding a membership.

---

## L8 (e2e enrollment.spec + axe) — **APPROVE** · enrollment implementation review (L1–L8) complete

Commits: b803007 (L7c2 revoke-gate on_student), 2d21b99 (L7d named comms errors + queue contrast +
learn report-success fix), 42cebce (enrollment.spec + route-guards + recovery:83 scoping). Authoritative
clean-DB e2e: 107 passed, 0 failed.

**Verdict: APPROVE — 0 blocking, 0 non-blocking.** Every focus point is proven on the production path,
and the L4 guard-rail held — the spec-fixes corrected mis-targeted assertions and kept the deny-paths;
no gate was weakened.

### Row-by-row (verified against the committed spec)

- **En04/En10 — media bytes at nginx (the crux):** mints a real playback session, walks the actual
  `/media/p/<id>/` master→rendition→segment through nginx (not the Go handler): enrolled student →
  segment **200** with bytes; after admin revoke → same segment **403** (nginx auth_request); after
  marking the lesson preview → same segment **200** again (Published&&Preview opens it to a non-member).
  It even isolates that the 403 was the membership gate, not a session-TTL artifact, by re-enabling via
  preview on the *same* segment. This is the production-path proof I asked for — a revoked member's
  bytes are cut by nginx, and the L3a preview-Published fix is exercised at the byte level.
- **En05/En08 — comms message (the L7d requirement):** blocked member retains content; a comment POST
  returns the code `enrollment_blocked` (line 282) AND the discussion panel renders «Вас заблоковано в
  цьому курсі» (line 285) — the user-visible named message, not only the network code, exactly what
  L7d had to add. Roster toggles active↔blocked via the UI.
- **En03:** outline/«Програма курсу» stays public (200) while a non-member's lesson content
  (`/learn/lessons/{id}/comments`) → 403 — the corrected assertion keeps the content deny-path, not just
  outline=200.
- **En04/En15 — single row:** after revoke + re-enroll, the roster filtered to the student has exactly
  one active entry (`toHaveLength(1)`, state active) — the one-row-per-(user,course) invariant proven,
  and scoped to the user rather than an absolute count.
- **En07/En14:** participant report → admin queue → resolve (report count → 0); instructor self-enrolls
  in another course; `wcagViolations(queue) == []` confirms the queue color-contrast fix (4.39→12.77:1).
- axe (wcag2a/2aa/21aa), 44px, 200% zoom across the new screens; `recovery:83` now scoped to the seeded
  slugs (the AGENTS "scope the selection" point), not `total==3`.

### Guard-rail held

No red was turned green by weakening a gate: En03/En04 keep the content-endpoint 403, En04 proves the
nginx bytes-cut, En05 keeps the comms 403 + message. The spec-fixes (outline-vs-content endpoint, the
non-existent heading, stale /learn redirects) corrected the tests onto the real gates. `CanConsume`/
`CanCommunicate` were not relaxed.

**Enrollment L1–L8 are all APPROVE.** Remaining: L9 (report + the authoritative ordered `make check` +
PR). For the L9 sign-off I'll check the report's row-by-row matches what is actually proven (a row is
closed when proven, not when a same-named test exists), the ordered `make check` is green end-to-end,
and the named limitations are recorded: paid-revoke purchase dead-end (L4a), the e2e DB-persistence /
recovery-count fragility + recovery:83 scoping, the L5 resolve-unknown-id idempotence decision, and the
L6a revoke authz-in-application decision.

---

## L9 (report + ordered gate + PR) — **CHANGES REQUESTED**

Commit: report fcb0dec; ordered `make check` reported green end-to-end (exit 0, e2e 107).

**Verdict: CHANGES — 1 blocking.** The gate, decisions (§4) and the L1–L8 rows hold, but one §2 row is
tagged at a proof level that does not exist — the exact Phase-8 lesson the report itself cites.

### Blocking

**B-L9.1 — En13 (backfill) is reported "integration (migration on seeded DB)" but no test proves it.**
§2 tags En13 ("existing free user with progress OR comment/review keeps access after migration") as
integration-proven. There is no such test:
- The only "backfill" integration test is `course/…:80 "authoring migration backfill failure and
  rollback"` — that's the authoring/grandfather price-set backfill from an earlier phase, unrelated to
  00015.
- The orders integration setup runs `provider.Up()` over all migrations, but on an **empty** DB — so
  00015's `INSERT … SELECT … UNION` backfill finds no pre-existing activity and is a no-op. The seed-fix
  (demo student self-active in seed.sql) is a seed INSERT, not the backfill SELECT. So the backfill's
  actual logic — the three-way UNION (progress / comment / review), the `price_minor IS NULL AND
  status='published'` filter (the status filter I flagged in L1), and `ON CONFLICT DO NOTHING` — is
  exercised by nothing. The grep match on "00015" in the orders test is a false positive (UUID
  constants `…000000000151`).

Why it matters: this is a **one-shot data migration** — it runs once on the populated production DB and
can't be re-tested after deploy. If the UNION or the filter is wrong, existing free-course consumers
(progress/comments/reviews, no membership row) are NOT backfilled → locked out of free content they
previously had — the precise lockout the backfill exists to prevent. A correct-looking migration SQL
(which I reviewed in L1) is not a proven row; per the phase-closing rule, En13 can't be marked proven
without a test on populated data.

Fix: add an integration test that exercises the real 00015 backfill on populated data — the cleanest is
the DownTo/Up pattern already used by the course migration-preservation test: `DownTo(14)`, seed a free
published course + a membership-less user with (one subtest each) a `lesson_progress` row / a comment /
a review, `Up()` to 15, and assert the user now has a `self`/`active` grant. Add the negatives too: a
user whose only activity is on an **archived** or **paid** course is NOT backfilled, and an existing
member is not duplicated (ON CONFLICT). Then re-tag En13 as integration-proven honestly.

### Otherwise sound

- Ordered `make check` green end-to-end (format→lint→typecheck→test -race→integration→check-generated→
  build→up→reseed→smoke→e2e, e2e 107, exit 0) — the full pipeline, not a subset.
- §2's other integration tags are honest: En02 (settle upgrade — verified in L2a), En09 (enroll on
  archived/draft → `ErrCourseNotEnrollable`, integration 669/673 + HTTP 409 + unit) are really proven at
  the level claimed; the e2e rows (En03/04/05/10/15 etc.) I verified in L8.
- §4 records the two decisions I named (L5 resolve-unknown-id idempotence; L6a authz-in-application with
  the deliberate L6b-middleware asymmetry) plus the model/amendment/fixes.

Close B-L9.1 (one integration test) and re-tag En13; then the phase is honestly 15/15 and ready for the
PR. I'll re-check just that test.

### L9 re-check (B-L9.1 fix) — **APPROVE** · enrollment phase signed off

Commit: test 57dd18d, re-tag 5626ba5.

B-L9.1 resolved. The new subtest `course/postgres "enrollment migration backfills free-course activity
into self-active membership"` exercises the REAL 00015 backfill on populated data: `DownTo(14)` → seed
(at v14) a free-published + paid + draft + free-archived course and 7 membership-less users with a
`lesson_progress` row / comment / review (incl. user1 with two progress rows AND a comment on the free
course) → `UpTo(15)` → assert, per user, `count(*)` and `count FILTER (source='self' AND state='active')`
both equal the expected:
- A progress×2 + comment → 1 (cross-source UNION dedup — my suggested case, added);
- B comment-only → 1; C review-only → 1 (each UNION arm);
- D paid → 0; E draft → 0 (the `price_minor IS NULL AND status IN ('published','archived')` filter);
- F free-archived → 1 (the L1 NB#1 widening).
The assertion is strict (`count != grants || selfActive != grants`), so the excluded users must have NO
grant, not merely no self-active one; `t.Cleanup` + a final `Up()` restore the shared schema/counts.
`make integration` GATE_EXIT=0 (course/postgres 20.3s) — it runs on the production path (real Postgres +
goose), not an empty DB. En13 re-tagged honestly; §5 records the proof-gap. En13 is now genuinely proven.

**Enrollment L1–L9: APPROVE.** The phase is honestly 15/15 — every row proven at the level the report
claims; the ordered `make check` is green end-to-end; §4 decisions and §6 limitations carry
magnitude/mechanism. Clear to open the PR (merge is the owner's call). This closes the enrollment
code review; the retroactive accounts review (work/review-courses2-accounts.md) remains APPROVE.
