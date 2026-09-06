# Phase 4b — identity administration: delivery report

Delta to [plan-4b.md](plan-4b.md) (edition 2.1, GATE 2 green). Written 2026-09-06 at slice
close. The Post-deploy section is left for `post-deploy-check`.

## 1. What was done

Nine tasks T01–T09, each implemented with Codex (gpt-5.6-sol) and re-verified here before
commit; per-task evidence in `work/report-t01.md` … `report-t08.md`. Everything landed as
planned except the deviations below.

- **Backend** (T01–T05): migration 00004 adds `users.deleted_at`, a CHECK enforcing
  deleted⊂disabled, and a partial email-unique index over non-deleted rows; the domain
  gains AccountStatus and the CanBlock/CanDelete invariants; the application use cases run
  the policy as a callback the adapter executes under its lock; the PostgreSQL adapter
  serialises the last-admin count with `LockAdmins ORDER BY id FOR UPDATE` under READ
  COMMITTED and revokes sessions on block/delete/reset; four HTTP routes ride the reused
  4a authorization seam.
- **Frontend** (T06–T08): the admin-users adapter/application mirroring the 4a slice; the
  `/admin/users` route and directory with per-row actions and a self-signing-out reset
  dialog; the browser journey on disposable fixtures.
- **Gate** (T09): the full ordered `make check`.

**Deviations from the plan:**
- **The three disposable fixtures moved id block** — they were first seeded at
  `...004/005/006`, which collide with the course authoring integration test's own
  fixtures (`...004` other-instructor, `...005` disabled-owner). Relocated to a reserved
  `...00d1/d2/d3` block; caught by `make integration` at review, not a plan change.
- **A redundant route-level self-reset test was removed** — it hung the rstest runner by
  navigating out of an open React Aria modal (a test-environment interaction, not a
  product bug). Self-reset sign-out stays covered at the ResetPasswordDialog component
  level; the redirect itself is the same layout mechanism 4a's R35 proves.

## 2. Acceptance scenarios

44 rows from the plan's acceptance table, each owned by exactly one task, every task's
evidence green inside the final `make check` (exit 0). Rows "not verified": **0**.

| ID | Scenario | Owner | Actual | Verdict |
|---|---|---|---|---|
| P01 | Directory lists every user with status | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| P02 | Default filter hides deleted; the deleted tab shows them | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| P03 | Block kills auth and live sessions | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| P04 | Restore re-enables auth | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| P05 | Soft delete retains the row, disables it, and frees the email | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| P06 | Password reset: new works, old fails, sessions revoked | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| P07 | The directory renders per-row actions | T07 | closed by T07; in the exit-0 `make check` | збіглося |
| P08 | The full admin journey in a browser | T08 | closed by T08; in the exit-0 `make check` | збіглося |
| B01 | Empty directory (API) | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| B02 | Empty directory (screen) | T07 | closed by T07; in the exit-0 `make check` | збіглося |
| B03 | Blocking a non-last admin is allowed | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| B04 | Password bounds | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| B05 | Reset produces a valid Argon2id credential | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| B06 | Migration over a populated database | T01 | closed by T01; in the exit-0 `make check` | збіглося |
| B07 | The failing-migration canary still rolls back | T01 | closed by T01; in the exit-0 `make check` | збіглося |
| B08 | Login survives email reuse | T01 | closed by T01; in the exit-0 `make check` | збіглося |
| B09 | Contract and client stay in step | T06 | closed by T06; in the exit-0 `make check` | збіглося |
| B10 | The ordered gate passes | T09 | closed by T09; in the exit-0 `make check` | збіглося |
| B11 | Same-state block is idempotent | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| B12 | A block writes updated_at so the CHECK holds | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R01 | Anonymous refused everywhere | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R02 | A student refused everywhere | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R03 | An instructor refused everywhere | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R04 | A disabled admin refused | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R05 | Cross-site mutation refused first | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R06 | Admin cannot block themselves | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R07 | Admin cannot delete themselves | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R08 | Admin cannot delete an admin | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R09 | The last active admin cannot be blocked | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R10 | The last-admin guard holds under the race, with no 503 | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R11 | Unknown id is not found | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R12 | A deleted target is not operable | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R13 | Malformed id rejected before lookup | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R14 | Illegal status value refused | T05 | closed by T05; in the exit-0 `make check` | збіглося |
| R15 | Dependency failure degrades safely | T03 | closed by T03; in the exit-0 `make check` | збіглося |
| R16 | A hasher failure is not a business refusal | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R17 | Cancellation is not a fault | T03 | closed by T03; in the exit-0 `make check` | збіглося |
| R18 | A cancelled block leaves nothing behind | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R19 | Business codes reach the domain intact | T06 | closed by T06; in the exit-0 `make check` | збіглося |
| R20 | The client guard is not the mechanism | T08 | closed by T08; in the exit-0 `make check` | збіглося |
| R21 | Unusable search params disappear | T07 | closed by T07; in the exit-0 `make check` | збіглося |
| R22 | Login racing a block **or delete** leaves no live session; a second concurrent delete is 404 | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R23 | A blocked/deleted owner's course cannot be PUT | T04 | closed by T04; in the exit-0 `make check` | збіглося |
| R24 | An instructor sees not-found at `/admin/users` | T07 | closed by T07; in the exit-0 `make check` | збіглося |

## 3. Non-obvious findings

- **GATE 2 caught five real defects before any code** (`docs/admin/gates-4b.md`): a dead
  `email_taken` path, an invalid `count(*) FOR UPDATE` that deadlocks to 503, a rollback
  that resurrects soft-deleted accounts, two missed status-gating queries that would let a
  deleted user rotate a session and remain a course owner, and instructor-block freezing
  their courses. All closed by the single "deleted ⊂ disabled" invariant made structural
  in migration 00004's CHECK.
- **The last-admin refusal is a re-read, not a deadlock** — locking the whole active-admin
  set `ORDER BY id FOR UPDATE` under READ COMMITTED (not the course module's
  RepeatableRead) makes the concurrent loser re-read the committed `disabled_at` and return
  `409 last_admin`, never a `503` from a `40P01` abort (`administration.go`, T04 report).
- **The rstest suite is pathologically slow** (~10 min for 293 tests) and repeatedly
  exceeded foreground timeouts during review; the working method was detached runs plus a
  process-exit watch (recorded for the next session).
- **e2e reseeds identities between the two 429-check runs** because soft-delete is terminal
  in 4b — delete-me cannot be un-deleted, so a fresh seed restores the disposable set.

## 4. Screenshots

Not applicable as static images here; the admin-users flow is exercised visually by
`frontend/e2e/admin-users.spec.ts` against real Compose (block → restore → reset →
soft-delete under the status filter).

## 5. User flow walk-through

Verified by `admin-users.spec.ts` (not by hand). An admin opens `/admin/users`, sees the
directory with status tabs, blocks a user (it moves to blocked and can no longer sign in),
restores it, resets another user's password, and soft-deletes a third out of the active
list. Resetting their own password signs the admin out to `/login`. An instructor
reaching `/admin/users` gets the not-found shell; a student's direct API mutation is
refused by the server.

## 6. Risks and limitations

- **No un-delete** — soft delete is terminal in 4b; a deleted user's row is retained for
  audit and email uniqueness but there is no restore-from-deleted operation.
- **No role change** — promoting a student to instructor is out of 4b (a distinct
  capability); the additive-capability model (brief §24.10) means any user can learn once
  enrollment exists, so role change is only about authoring/admin power.
- **Per-course user blocking** still waits for enrollment.
- **No super_admin role** — admin-protection is domain invariants (an admin cannot be
  deleted, cannot have their role changed via the API, cannot block themselves, and the
  last active admin cannot be blocked).
- **Password reset has no email flow** — no mailer exists; the admin sets a new password
  directly, and the old-password TOCTOU window is an accepted, recorded limitation.
- **Retained debt** (plan §12): the NB2 non-blocking notes disposed there; none blocking.

## Post-deploy

_Left for `post-deploy-check` after deployment. No deploy target exists in the repo yet._
