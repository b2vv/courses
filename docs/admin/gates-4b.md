# Phase 4b — gate log

## GATE 2 — plan-defense (sonnet, fresh context)

Target: `docs/admin/plan-4b.md`. Method: `plan-defense` skill, five fixed questions, each
closed only by a named location in the plan itself. Code claims spot-checked against
`backend/migrations/00002_identity.sql`, `backend/internal/identity/adapters/argon2id/verifier.go`,
`backend/sql/queries/identity.sql`, `backend/internal/identity/adapters/http/handler.go`,
`backend/internal/identity/application/identity.go`, `backend/internal/identity/adapters/postgres/repository.go`,
and `backend/internal/identity/domain/user.go`.

### Code-claim spot checks (all confirmed accurate)

- `00002_identity.sql` creates `users` with `email text NOT NULL UNIQUE`, an inline column
  constraint. Postgres names this `users_email_key` by default — plan-4b.md §6 step 3's
  `DROP CONSTRAINT users_email_key` is correct.
- `argon2id/verifier.go` confirms: `encode` is an unexported function used only by
  `NewVerifier` (the dummy hash) and by the package's own tests; there is no exported
  `Hash`/credential-writing entry point. Plan-4b.md §3 "What is new" (line 155-164) and §11
  item 1's claim that "the app never hashed" and needs a new `PasswordHasher` port is
  accurate.
- `identity.sql` confirms `GetIdentityCredential` (`:one`, `WHERE email = $1`, no
  `disabled_at`/`deleted_at` filter), `ResolveIdentitySession` (already filters
  `disabled_at IS NULL`), and `ListInstructors` (already filters `disabled_at IS NULL`) —
  matching plan-4b.md §5.a exactly, including that `GetIdentityCredential` currently has no
  filter at all and would break first under email reuse.
- `handler.go` confirms `Authenticate`, `RequireSameOrigin`, `userKey{}`,
  `CanAdministerAccounts()` and the existing `GET /admin/instructors` route exist exactly as
  plan-4b.md §4 D-1 describes, including the outer-`RequireSameOrigin`/inner-`Authenticate`
  nesting order for mutations.
- `postgres/repository.go`'s `Rotate` is a real transactional pattern (`tx.Begin` → work →
  `tx.Commit`/`rollback`) matching the plan's citation of it as the transactional template.
- All internal citations checked (`NB-8`, `NB-9`, `NB-10`, `D-1`, `D-2`, `D-4a`, `D-7`,
  `R33`, `BLK-12`, `I03`, `I06`, `A10`) resolve to real sections of `docs/admin/plan.md` /
  `docs/admin/report.md` — no invented cross-references found.

### One material finding surfaced during the code check (bears on race/replay, below)

`postgres/repository.go`'s `Rotate` (the login session-creation path, pre-existing 4a/
foundation code) calls `LockActiveIdentityUser` — `SELECT id FROM users WHERE id = $1 AND
disabled_at IS NULL FOR UPDATE` — **inside the same transaction** that creates the new
session, and `application/identity.go`'s `Login` checks `user.Active()` from a credential
read that happens **before** that transaction, in a separate round trip. This means a login
in flight when an admin blocks the same user is a genuine race window that plan-4b.md never
names or reasons about, even though the specific text of §11 item 9 asserts "the only real
race is the last-admin count" (see row 3 below).

### Table

| # | Question | Closed by | Verdict |
|---|---|---|---|
| 1 | Hostile input | plan-4b.md §4 D-3 refusal matrix + acceptance rows R13 (malformed id → `400 invalid_request` before any repository call), R14 (illegal `status` enum value → `400 invalid_request`, deleted reachable only via `DELETE`), B04 (empty password and over-bound password → `400 invalid_request`, credential unchanged); handler-level "decode and size limits" named in §2 Backend table row for `admin_handler.go` | closed |
| 2 | Dependency failure | plan-4b.md §9 Refusals R15 (repository port dependency failure on `ListUsers`/`SetStatus` → `503 dependency_unavailable`, no SQL/credential/internal detail leaked), R16 (`PasswordHasher` failure surfaces as dependency-unavailable, not a business refusal), R17 (context cancellation is reported distinctly from both); owned by T03 against fakes | closed |
| 3 | Race / replay | **partially closed.** (a) Concurrent block of the last two admins: closed by §4 D-2 "Last active admin (Q5)" — `CountOtherActiveAdmins` under `FOR UPDATE` inside the same transaction as the status flip, plus acceptance row R10 ("two concurrent blocks… at most one succeeds… loser answers `409 last_admin`"). (b) Restore hitting a now-taken email: closed by §4 D-2 "Password reset…"/email_taken paragraph and §8's `email_taken` row — mapped from Postgres `23505` on `users_email_active_key`, explicitly "never a pre-check" (NB-10 reuse), which is atomic by construction under concurrency, plus acceptance row B02. (c) **Block racing an in-flight login is not answered anywhere in the plan.** No line in §4 D-2 (the transactional-invariants section, which is exactly where this belongs) or §9 mentions the login/session-creation path at all, and §11 item 9 states outright "the only real race is the last-admin count" — a claim the plan does not qualify against the login path. This is not a hypothetical the plan implicitly waves off in scope language (contrast with the explicit "poza scoupom" phrasing used elsewhere for role change, e.g. §"Scope fence"); it is a race the plan simply never raises, so per plan-defense discipline it is unanswered regardless of what the actual code does. (Spot-checking the code as a matter of intellectual honesty, not as a way to close the row: `Rotate`'s `LockActiveIdentityUser … FOR UPDATE` and the plan's own `SetStatus`/`DeleteIdentitySessionsForUser` transaction would very likely serialize correctly against each other through ordinary Postgres row locking — but the plan text asserts and reasons about none of this, so it does not count as closed.) | **hole** |
| 4 | Empty and boundary data | plan-4b.md §9 Boundaries B01 (`status=deleted` matching nobody → `200 {items: [], total: 0}` + localized empty state, owned T05/T07), B04 (password length boundaries), B08 (duplicate/shared email across a deleted and an active row — the core boundary this slice introduces), and Refusals R09 ("exactly one active admin remains" — the single-element boundary for the last-admin invariant) | closed |
| 5 | Rollback | plan-4b.md §7 Rollback, read together with §6 Migration plan's "Down:" paragraph (lines ~372-377). §7 names the supported path (redeploy previous images, leave migration 00004 in place), names the one data-shape hazard a rolled-back 4a binary is exposed to (a `GetIdentityCredential` `:one` query erroring on a reused email — but only after a real reuse has happened), and explicitly forbids `goose down 00004` as an operational recovery step because it is "destructive-by-construction": it drops `deleted_at` and its audit trail, and its re-added table-wide `UNIQUE` **fails outright** if any email has been reused — i.e. the plan states the irreversible/destructive property of the down path directly, not just its existence | closed (one wording nit, non-blocking: §7's phrase "drops `deleted_at` and every soft-delete/block audit trail it carries" reads as if the down migration also destroys block-state audit; the down migration only touches `deleted_at`/the partial index, not `disabled_at`, so "block" audit is not actually at risk from this specific migration — cosmetic, does not change the verdict) |

### Rows closed / open

Closed: 1 (hostile input), 2 (dependency failure), 4 (empty/boundary data), 5 (rollback).
Open: 3 (race/replay) — specifically the block-racing-an-in-flight-login sub-case, which
the plan never names, reasons about, or scopes out, despite §11 item 9 asserting there is
only one real race in the slice.

**Verdict: RED** — one row (race/replay) open. Recommended fix before this leaves GATE 2:
add a line to §4 D-2 (or a new D-2 bullet) that either (a) explicitly reasons about the
login/`Rotate` path against a concurrent block/delete — pointing at
`LockActiveIdentityUser`'s `FOR UPDATE` as the serializing mechanism — and adds an
acceptance row exercising it, or (b) explicitly scopes it out with a named reason if the
authors judge the existing `Rotate` lock sufficient and out of 4b's own surface to test.

## GATE 2 — plan-critique (fable, fresh context)

Target: `docs/admin/plan-4b.md` (edition 1, 2026-09-06). Method: `plan-critique` skill —
four independent lenses, each a fresh `general-purpose` subagent with the plan, the spec,
`AGENTS.md`, `.claude/standards.md` and the 4a plan, blind to the other lenses; consolidation
and every code-fact spot-check below done by the critic itself in a clean context.

**Lenses: 4 launched / 3 returned to this context** — AI-slop (7 findings), logic holes
(9 findings), edge cases (12 findings). The **excess-complexity** lens was launched, nudged
once, and the coordinator reports it returned, but its result was never delivered to this
context, so it is recorded here as **not received** rather than clean (skill §"Лінза, що не
повернулась"). Its ground is partly covered by the critic's own reuse-first checks and by the
AI-slop lens's structural findings (NB-16–NB-21: staleness constant, branded id, twin
composition/search files, `PasswordHasher` port, redundant self-delete rule); anything only a
dedicated complexity read would catch stays unreviewed and is named in the verdict.
Duplicates across lenses are collapsed into one finding with the count of lenses that saw it;
blocking first.

### Code-fact spot checks (opened, not trusted)

| Plan claim | Verdict | Where |
|---|---|---|
| `users.email` is an inline `UNIQUE` (constraint `users_email_key`); `disabled_at` CHECK is `BETWEEN created_at AND updated_at` | **holds** | `backend/migrations/00002_identity.sql:4,22–24` |
| `users_email_key` is referenced by no code; `uniqueCollision` names only session constraints | **holds** | `backend/internal/identity/adapters/postgres/repository.go:236–239`; grep over `backend/**` finds no other reference |
| `argon2id` can only verify — `encode` is unexported, used only for the dummy hash; no hashing entry point anywhere | **holds** | `backend/internal/identity/adapters/argon2id/verifier.go:26,54–57`; `main.go:44` injects `passwords.NewVerifier()` only |
| The 4a seam: exported `Authenticate`/`RequireSameOrigin`, `userKey{}`, `CanAdministerAccounts()` gating `GET /admin/instructors`; `RequireSameOrigin(Authenticate(h))` runs same-origin first | **holds** | `backend/internal/identity/adapters/http/handler.go:60–64,111–140,142–150`; `backend/cmd/api/main.go:53` |
| Three identity queries need `deleted_at IS NULL` | **incomplete** — five queries gate on account status: `GetIdentityCredential`, `LockActiveIdentityUser` (used by `Rotate` at login), `ResolveIdentitySession`, `ListInstructors` (all `backend/sql/queries/identity.sql`) and the course module's `LockEligibleAuthoringOwner` (`backend/sql/queries/courses.sql:105–111`, run on every course create/PUT at `authoring.go:154,206`) | see BLK-4, BLK-5 |
| `GetIdentityCredential` is `:one` and "would error at login" on two rows | **false mechanism** — sqlc `:one` is pgx `QueryRow`, which silently returns the first matching row (`generated/identity.sql.go:108–109`); the remedy is right, the stated failure is not | see NB-5 |
| The course canary `00004_failure.sql` collides with a real `00004` | **holds** | `backend/internal/course/adapters/postgres/repository_integration_test.go:171–175` (MapFS of 00001–00003 + synthetic 00004); the identity schema test uses `os.DirFS` (`schema_integration_test.go:58`) as the plan says. The citation is `plan.md` BLK-12 / commit `fd90fc9`, not `report.md` §1 |
| `seed-identities.sql` `ON CONFLICT (id) DO UPDATE` resets `disabled_at` | **holds** | `backend/sql/seed-identities.sql:8–14` |
| `domain.User` carries "the two timestamps" | **false** — `domain.User` has `DisabledAt` only; `Active()` is `DisabledAt == nil`; `mapUser` takes seven columns | `backend/internal/identity/domain/user.go:35–43,59–61`; `repository.go:217–227` — see NB-2 |
| `/admin` layout guard covers `/admin/users` | **partly** — the guard is `canAccessAdministration` (instructor **or** admin); the frontend domain has no `canAdministerAccounts` | `frontend/src/routes/admin.tsx:13`; `frontend/src/domain/identity/identity.ts` — see NB-4 |
| `SELECT count(*) … FOR UPDATE` | **not executable** — PostgreSQL forbids `FOR UPDATE` with aggregate functions | see BLK-2 |

### Blocking findings

```
[logic ×2 + edge + critic] BLK-1  `email_taken` / B02 contradict R12, D-3 and §1 — the code can never fire
де:      §1 lines 45–47 (deleted is terminal, only via DELETE); D-3 note lines 255–258; §8 line 417
         (deleted target → 404 on status) — versus §2 lines 84–85, §8 line 418 and 426–429, B02
         line 458, T04 line 511, risk #10 line 552
чому:    The same request — POST /users/{U}/status {"status":"active"} with U soft-deleted — is
         asserted to answer 404 user_not_found (R12) and 409 email_taken (B02). Spec S3 defines
         restore as un-blocking; a blocked row has deleted_at IS NULL and is already inside
         users_email_active_key, so flipping disabled_at cannot add a row to that index. No 4b
         write sets deleted_at back to NULL or edits email, so 23505 on the partial index is
         unreachable from any 4b endpoint: B02 has no producible evidence, T04 carries a dead
         mapping and the OpenAPI enum an unfirable code.
що зробити: Keep the 404 rule (matches §1, S3, R12). Delete email_taken from §2 (lines 84–85),
         §8 (line 418 and the paragraph 426–429), B02, T04's Owns list and risk #10. If undelete
         is wanted instead, add it as an explicit fifth operation and drop "terminal" from §1.
вага:    blocking
```

```
[logic + edge + critic] BLK-2  `CountOtherActiveAdmins … FOR UPDATE` is not valid SQL, and the R10 race deadlocks — the loser gets 503, not 409
де:      §2 line 90; §4 D-2 lines 218–225; R10 line 481; risk #5 line 547
чому:    PostgreSQL rejects `SELECT count(*) … FOR UPDATE` ("FOR UPDATE is not allowed with
         aggregate functions"); sqlc cannot generate the query as named. Rewritten as a row lock
         over "the admins other than the target", the scheme still does not serialise: Tx1
         (A blocks B) locks A's row, Tx2 (B blocks A) locks B's row; Tx1's UPDATE … WHERE id=B
         waits on Tx2, Tx2's UPDATE … WHERE id=A waits on Tx1 — deadlock, one side aborted with
         40P01, which persistenceError (repository.go:241–243) maps to ErrDependencyUnavailable
         → 503 dependency_unavailable (handler.go:259–260). D-2's "one waits, re-reads a count
         of zero, and is refused" and R10's "the loser answers 409 last_admin" do not follow;
         R10 is flaky by interleaving.
що зробити: Replace the query with a row-returning lock over the WHOLE admin set in a
         deterministic order — `LockAdmins :many — SELECT id, disabled_at FROM users WHERE
         role='admin' AND deleted_at IS NULL ORDER BY id FOR UPDATE` — and count the other
         active admins in the domain. State in D-2 that the target row is inside the locked set
         and that the lock is taken before the UPDATE; the second transaction then waits on the
         first, re-evaluates (the just-blocked row drops out) and refuses with last_admin. Add
         "no 503 under the race" to R10.
вага:    blocking
```

```
[logic + edge + critic] BLK-3  §7 rollback resurrects every soft-deleted account
де:      §7 lines 386–396 ("never depends on deleted_at", "the only data-shape hazard")
чому:    A 4b-deleted row has deleted_at set, disabled_at NULL and password_hash intact. The
         4a binary's GetIdentityCredential (identity.sql:1–4) has no deleted_at filter,
         VerifyCredentials accepts on matched && user.Active() (repository.go:62), and Active()
         is DisabledAt == nil (user.go:59–61). On the plan's own "supported rollback path" every
         deleted user logs in again, ResolveIdentitySession resolves them, ListInstructors
         offers a deleted instructor in the owner picker and LockEligibleAuthoringOwner accepts
         them as a course owner. That is a security regression, not a lost capability.
що зробити: Make DELETE also write `disabled_at = COALESCE(disabled_at, $now)` beside
         deleted_at (deleted ⊂ disabled; the projection already gives deleted_at precedence and
         deleted is terminal, so restore never clears it). State this in §1/§4 D-2 and rewrite
         §7: with it, every existing disabled_at IS NULL filter in both modules refuses a deleted
         user under both binaries and rollback is genuinely safe. Add a T04 row "after DELETE,
         login with the old password fails against both the 4b and the 4a credential query".
         If rejected, §7 must say rollback requires blocking every deleted_at IS NOT NULL row
         first.
вага:    blocking
```

```
[logic + edge + critic] BLK-4  §5.a counts three status-gating queries; there are five, and the two omitted ones let a deleted user own courses and rotate sessions
де:      §2 line 90; §5.a lines 266–281; risk #3 line 545; scope claim §2 line 66 ("entirely in the identity module")
чому:    Not listed: LockActiveIdentityUser (identity.sql:6–10; Rotate at repository.go:79 —
         login is two pool calls, so VerifyCredentials → admin DELETE → Rotate creates a session
         row for a deleted user, contradicting Q1 "sessions gone") and the course module's
         LockEligibleAuthoringOwner (courses.sql:105–111, run on every course create and PUT,
         authoring.go:154,206 — an admin can assign or keep a soft-deleted instructor as owner;
         only the picker hides them). "No window where login is undefined" is incomplete, and
         one of the missing edits is cross-module.
що зробити: Either list all five queries in §2/§5.a/T01 (naming the courses.sql edit and its
         `make generate` path, and correcting the "entirely in the identity module" sentence),
         or adopt BLK-3's "deleted ⇒ also disabled" and state in §5.a that the remaining
         disabled_at filters are then correct by construction, leaving only
         GetIdentityCredential to gain deleted_at IS NULL for the two-rows-per-email case.
         Also fix the intra-transaction order in D-2: lock/UPDATE the users row first, then
         DeleteIdentitySessionsForUser, so Rotate's FOR UPDATE serialises against it; add a
         T04 row "login racing a block/delete leaves no session row".
вага:    blocking
```

```
[edge + critic] BLK-5  Blocking (or deleting) an instructor silently bricks editing of every course they own
де:      §5.a lines 276–277; §5.g line 334; scope fence line 15 ("no public surface delta", nothing on the course module)
чому:    LockEligibleAuthoringOwner (disabled_at IS NULL) runs on every course PUT even when
         the owner is unchanged (authoring.go:206). All five demo courses belong to fixture
         …0002 (backend/sql/seed.sql); once that instructor is blocked, every PUT on their
         courses — an admin's title edit included — answers 400 invalid_owner until an admin
         reassigns the owner. Conversely a deleted owner (disabled_at NULL today) keeps passing
         and CourseEditorView falls back to the raw id for the owner name. The plan states
         neither outcome, and an implementer hits it first in e2e (P08 blocks a student, so the
         plan's own journey never sees it).
що зробити: Add a §5 side-effect row for the course module: "blocking or deleting an
         instructor makes their courses uneditable (invalid_owner) until an admin reassigns the
         owner; nothing is reassigned automatically; their published courses stay published"
         — and decide whether that is the intended product stance or whether an admin's PUT
         that does not change the owner must skip the eligibility lock. Add a T04 row "PUT on
         a course whose owner is blocked/deleted answers invalid_owner (or succeeds, per the
         decision)".
вага:    blocking
```

### Non-blocking findings

```
[logic] NB-1  Sequential last_admin is unreachable through the API; R09 as written answers cannot_block_self
де:      D-2 lines 218–225; D-3 line 250; R09 line 480
чому:    The caller must be an active admin, and CountOtherActiveAdmins excludes only the
         target, so the caller is always counted: blocking another admin always sees ≥1 and
         never trips last_admin; blocking oneself is refused earlier. R09's precondition
         "exactly one active admin remains" plus an admin caller implies caller == target.
         last_admin is a race-only outcome (R10).
що зробити: D-3 cell → "200 (a sequential call can never target the last active admin — the
         caller is one); last_admin is the race outcome, see R10". Rewrite R09 as an adapter-
         level row with an explicit artificial actor, or fold it into R10.
вага:    non-blocking
```

```
[logic + critic] NB-2  AccountStatus is "derived from the two timestamps" but domain.User has one; no file adds DeletedAt or maps it
де:      §2 line 77; §1 lines 49–52; T02 line 509 — no `user.go (extend)` row, no mapper change
чому:    domain.User (user.go:35–43) has DisabledAt only; NewUser validates only that bound;
         mapUser (repository.go:217–227) takes seven columns. T02 cannot write a projection over
         a field nobody adds, and ListAdminUsers has no mapper to a domain entity.
що зробити: Add `domain/user.go (extend)` (DeletedAt *time.Time, NewUser bound
         deleted_at >= created_at, Active() false when deleted) to §2 and T02; add the mapper
         for the ListAdminUsers row; note that the untouched login/resolve/instructor queries
         never return a deleted row so their nil DeletedAt is correct.
вага:    non-blocking
```

```
[logic + critic] NB-3  B06 "migration over a populated database" has no owning harness
де:      B06 line 462; §2 line 86
чому:    The identity schema test applies all migrations over an empty users table before
         seeding (schema_integration_test.go:62,148). The only harness that migrates over "5
         courses + 3 fixtures" is the course subtest at repository_integration_test.go:71–91,
         which asserts authoring backfill facts only.
що зробити: Name the file: either add B06's assertions to the course subtest (and list
         `repository_integration_test.go (extend)` for more than the canary renumber) or move
         them to the identity schema test with a DownTo(3) → seed → Up precondition.
вага:    non-blocking
```

```
[logic + critic] NB-4  "The /admin/users guard" names a guard the plan says it does not add; the layout admits instructors
де:      §2 lines 121, 123–125; T08 line 515
чому:    admin.tsx:13 gates with canAccessAdministration (instructor or admin). An instructor
         reaches /admin/users, the list answers 403 and the route renders an error boundary
         rather than the not-found a student gets. The frontend domain has no
         canAdministerAccounts to call.
що зробити: Either add `domain/identity/identity.ts (extend)` with canAdministerAccounts and a
         beforeLoad in admin.users.index.tsx throwing notFound() for non-admins, or drop the
         phrase and add a row for what an instructor sees. Also name the entry point: the
         IdentityHeader link goes to /admin only; how an admin reaches /admin/users is unstated.
вага:    non-blocking
```

```
[logic + critic] NB-5  §5.a/§7 misstate the :one failure mode
де:      §5.a lines 271–273; §7 lines 391–394; risk #3 line 545
чому:    sqlc :one is QueryRow + Scan: with two matching rows it returns the first silently.
         Without the filter a reused email authenticates whichever row comes first — possibly
         the deleted one — rather than erroring.
що зробити: Replace "error at login" with "silently authenticate against whichever row comes
         first" in §5.a and §7.
вага:    non-blocking
```

```
[edge + critic] NB-6  Same-state block/restore is undefined; an unconditional write clobbers the retained audit timestamp
де:      D-3 lines 243–253; §8 line 417
чому:    Block on an already-blocked user with `SET disabled_at = $now` overwrites the timestamp
         the spec retains for a later audit UI, re-runs the revoke and the last-admin count —
         where a blocked admin target is excluded and inactive, so a spurious 409 last_admin is
         possible. Restore on an active user: 200 or 409 — unstated.
що зробити: State in D-3/§8 that status is idempotent: a same-state transition answers 200 and
         touches neither disabled_at, updated_at nor sessions, and skips the last-admin count;
         add a contract row (T05) and a timestamp assertion (T04).
вага:    non-blocking
```

```
[edge + critic] NB-7  An admin resetting their own password logs themselves out mid-request
де:      D-3 line 249; §5.f lines 324–326
чому:    Q2 revokes every session of the target; the target is the caller, so the 204 is
         followed by invalidateAfterUserWrite refetching with a dead cookie → 401 →
         useRedirectWhenSessionEnds (require-identity.ts:30–45) navigates to /login. No §9 row
         covers it. e2e is safe (the spec never resets the shared admin).
що зробити: Decide and state: keep the calling session (revoke all but the presented token) or
         have the dialog announce the sign-out and navigate to /login itself; add a T07 row.
вага:    non-blocking
```

```
[edge] NB-8  Password bounds "reused from login" are under-specified
де:      B04 line 460; §2 line 77
чому:    The login bounds are only len == 0 || len > 1024 (application/identity.go:48) under the
         8 KiB body cap. Exactly-1024, whitespace-only and same-as-current are all accepted and
         untested; B04 covers two cases.
що зробити: State the bounds explicitly (1–1024 bytes, any content), add the exact-1024
         acceptance to B04, and record whitespace-only / same-as-current as accepted by design
         or name a minimum as a decision.
вага:    non-blocking
```

```
[edge] NB-9  Seed vs partial index on a persistent volume; duplicate deleted emails
де:      §5.e lines 308–314; §6 lines 358–364
чому:    The seed's DO UPDATE … deleted_at = NULL raises 23505 on users_email_active_key if a
         different active row holds delete-me@ — reachable only by a manual insert, but the seed
         is one transaction so make check fails wholesale. Two deleted rows with one email are
         allowed by the index; the directory will show both.
що зробити: State both facts and the manual recovery in §5.e.
вага:    non-blocking
```

```
[edge] NB-10  Block must write updated_at or the disabled_at CHECK rejects it; clock skew
де:      §6 lines 348–351 (delete is specified); §2 line 84 SetUserDisabledAt (block is not)
чому:    00002_identity.sql:24: disabled_at BETWEEN created_at AND updated_at. `SET disabled_at
         = $now` alone is 23514 → 503. Seeded rows get created_at from the DB clock while the
         API writes its own µs-truncated Clock; an API clock behind the DB clock violates
         updated_at >= created_at, and the fixtures' fixed clocks hide it.
що зробити: State that every 4b write sets updated_at to the same clock value it writes into
         disabled_at/deleted_at/password_hash; add a T04 assertion on a row seeded with now();
         record "API clock ≥ DB clock" as a limitation.
вага:    non-blocking
```

```
[edge + critic] NB-11  Paging and search are unspecified: default status, page past the end, ILIKE metacharacters
де:      §1 line 37; §2 line 115; R21 line 492
чому:    The course adapter escapes search with NewReplacer("!","!!","%","!%","_","!_") +
         ESCAPE '!' (course repository.go:33); the plan names neither the searched columns nor
         the escaping. The course list has no default status, so the directory would show
         deleted rows by default while Q4 puts them "under a deleted filter".
що зробити: State the default filter, that page past the end is 200 with items: [] and the
         true total, that ListAdminUsers searches email and display_name with the course
         replacer; add a contract row "?query=% matches only literal %" (T05).
вага:    non-blocking
```

```
[edge] NB-12  Deleted or blocked instructor's courses stay published under their id — unstated stance
де:      scope fence line 16; §5.g line 334
чому:    Public queries never join users, so nothing leaks (verified courses.sql:1–38), but all
         five demo courses are the instructor fixture's and a reviewer will ask.
що зробити: One sentence in §5.g: courses keep their state and owner; reassignment is an admin
         action, never automatic.
вага:    non-blocking
```

```
[edge] NB-13  admin-users.spec.ts is rerunnable only after make seed-identities
де:      §5.f lines 316–327; §6 lines 379–380
чому:    `make e2e` alone does not re-seed (Makefile:44); after one P08 run delete-me@ is deleted
         and terminal (R12 → 404), so a second bare run fails at the delete step. `make check`
         orders seed-identities before e2e, so only the dev loop is hit.
що зробити: State it in §5.f and have the spec assert the disposables' pre-state in beforeAll
         with a message naming the target.
вага:    non-blocking
```

```
[AI-slop] NB-14  A briefing quote that does not exist, refuted four times
де:      lines 8–11, 83, 155–164, 295–300, 543
чому:    work/CTO-RESEARCH.md contains no "the application creates credentials — password reset
         reuses that path"; its only credential lines are 98 and 141. The plan manufactures a
         wrong source to be right against, then repeats the correction in §2, §3, §5.c and §11.
що зробити: Delete the quoted claim and the "contrary to the briefing" framing; keep one
         sentence in §3: "argon2id can only verify (encode is unexported, verifier.go:54); 4b
         adds Hash with a crypto/rand salt." Drop §5.c and §11.1 as duplicates.
вага:    non-blocking
```

```
[AI-slop] NB-15  §11 "Ten riskiest decisions" and the scope fence restate the plan and the spec
де:      lines 541–552; lines 18–25
чому:    Every §11 item maps 1:1 to an earlier paragraph; the 4a plan has no such section.
         Restating is where drift starts — item 9's "the only real race is the last-admin count"
         exists only in the summary and is exactly what plan-defense row 3 flagged.
що зробити: Reduce §11 to a "risk → section" pointer list or delete it; replace the scope fence
         with "Scope is exactly spec-4b.md S1–S8; its out-of-scope list stands unchanged."
вага:    non-blocking
```

```
[AI-slop] NB-16  A third staleness constant, ignoring the trigger 4a set for exactly this case
де:      line 114; T06 line 513
чому:    4a D-2c (plan.md line 205) named the trigger "a second slice imports either constant".
         4b is that slice and mints a new number instead of importing adminListStaleMs
         (admin-queries.ts:15).
що зробити: Line 114 → import adminListStaleMs (trigger fired; promote or import, no new
         constant); add the row to §3.
вага:    non-blocking
```

```
[AI-slop] NB-17  Branded UserId for a single id kind, against the 4a rationale for branding
де:      lines 108, 111; T06 line 513
чому:    4a D-2a branded because three id kinds share editor call sites. 4b has one. The
         existing AuthenticatedUser.id (identity.ts:4) is a plain string, so the one comparison
         4b needs (caller vs row for the self-block affordance) goes through the escape hatch.
що зробити: Plain string ids in AdminUser; parse without branding; name the trigger (a second
         identity id kind sharing call sites).
вага:    non-blocking
```

```
[AI-slop] NB-18  Twin composition/search files where an extension would do; __root.tsx (extend) is unlisted
де:      lines 113, 115, 123–125
чому:    admin.ts is nine lines; a second root forces a new key on RootRouteContext
         (__root.tsx:21–22), a file the plan omits — T06 finds it by the typecheck failing.
         admin-users-search.ts re-creates pageSize/maximumPage/pageSchema/echo of admin-search.ts
         with a different enum.
що зробити: Extend createAdminDependencies to return { ...courses, users }; extend
         admin-search.ts with the user status enum and q. If a separate root is kept, list
         __root.tsx (extend).
вага:    non-blocking
```

```
[AI-slop + critic] NB-19  ports.PasswordHasher mirrors "port per thing" rather than the module's own idiom
де:      line 82; lines 155–164; R16 line 487; T03 line 510
чому:    The identity module keeps its password interface consumer-side: PasswordVerifier is
         declared in adapters/postgres/repository.go:24–27 and injected via NewRepository, not
         in ports/. AGENTS.md: "Ports require an actual external dependency" — argon2 is an
         in-process library. The plan claims reuse-first and introduces a second placement,
         costing a port file, a constructor argument, a fake and an application-layer row for
         a computation the application never sees today.
що зробити: Either drop PasswordHasher — ResetPassword(ctx, id, plaintext) hashes inside the
         postgres adapter through the existing consumer-side interface widened with Hash; R16
         moves to T04 — or keep the port and state in §3 why the use case must hold the hash.
вага:    non-blocking
```

```
[AI-slop] NB-20  Claims the plan's own tables contradict
де:      lines 494–496 vs line 509; lines 101–102
чому:    T02 owns no acceptance row, so "every task owns at least one row" is false. Actor is
         written by Authenticate whether or not 4b exists; D-1 reads userKey{} and never
         touches Actor, so "reused unchanged" is decorative.
що зробити: "Every task except T02 owns a row; T02's evidence is its unit tests"; delete or
         reword the Actor sentence.
вага:    non-blocking
```

```
[critic] NB-21  Self-delete refusal is redundant with admin-immutable and misnamed
де:      D-2 lines 211–212; D-3 line 249; R07 line 478; §8 line 414
чому:    The caller is always an admin, so DELETE self already answers admin_immutable; the
         extra pre-transaction rule adds a second code path and a row proving a subset. If kept,
         `cannot_block_self` on a DELETE is a misleading code name.
що зробити: Drop the self-delete clause and R07, or rename the code to cannot_target_self and
         say it is a deliberate earlier refusal.
вага:    non-blocking
```

```
[critic] NB-22  Canary citation points to the wrong document
де:      §5.b line 292 ("report §1, BLK-12")
чому:    BLK-12 is a plan.md finding (edition 2 header, line 3–5); the renumber landed in commit
         fd90fc9. report.md §1 does not carry it.
що зробити: Cite plan.md BLK-12 / fd90fc9.
вага:    non-blocking
```

### Disposition of blocking findings (each carried to a named plan edit)

| # | Plan edit | Status |
|---|---|---|
| BLK-1 | Remove `email_taken`, B02, T04's `23505` mapping, risk #10; keep the deleted → 404 rule (§1, D-3, §8, R12) | open — plan not yet edited |
| BLK-2 | §2/§4 D-2: replace `CountOtherActiveAdmins (FOR UPDATE)` with a deterministic-order row lock over the whole admin set, target included, taken before the UPDATE; count in the domain; R10 adds "no 503" | open |
| BLK-3 | §1/§4 D-2: DELETE also sets `disabled_at`; §7 rewritten on that basis; new T04 row | open |
| BLK-4 | §5.a/T01: all five status-gating queries named (incl. the cross-module `LockEligibleAuthoringOwner`) or "deleted ⇒ disabled" adopted and stated; D-2 fixes the intra-transaction order; new T04 race row | open |
| BLK-5 | §5: course-module side-effect row and a decision on PUT-with-unchanged-owner under a blocked/deleted owner; new T04 row | open |

None of the blocking findings is accepted as-is; every one names the line that must change.
BLK-3 and BLK-4 share one structural fix (soft-delete also disables), which also closes the
`LockActiveIdentityUser` half of the plan-defense row-3 hole without a course-module edit.

### Verdict

**Blocking: 5 open. Non-blocking: 22.** Lenses returned: 3 of 4 (complexity not received).

**RED.** Five blocking findings stand, all with a named plan edit; three of them (BLK-2,
BLK-3, BLK-4) sit exactly on the invariants the gate was asked to weigh — the last-admin race
as described cannot be written in SQL and deadlocks, the rollback path re-enables deleted
accounts, and the soft-delete filter list misses two status-gating queries, one of them in the
course module. BLK-1 removes a dead code path (`email_taken`) that contradicts the plan's own
404 rule; BLK-5 names a course-module consequence of blocking an instructor that the plan does
not mention. The email-uniqueness migration itself (drop `users_email_key`, partial unique
index) is correct against existing rows and the seed and needs no change. Re-run GATE 2 on
edition 2 after the edits, with the complexity lens actually delivered.

## GATE 2 round 2 — plan-critique verification (fable, fresh context)

Target: `docs/admin/plan-4b.md` **edition 2** (2026-09-06). Method: `plan-critique` skill, all
four lenses run **inline by the critic in this one clean context** (no sub-agents, per the
edition-1 lesson: the complexity lens was lost to a coordination deadlock). **Lenses: 4 launched
/ 4 returned** — AI-slop, logic holes, excess complexity, edge cases — each recorded below with
≥1 finding or an explicit "clean, because". Every closure and every new claim was checked by
opening the file, not by trusting the plan.

### Part 1 — closure of the edition-1 blocking findings and the defense race hole

Files opened for this part: `backend/sql/queries/identity.sql` (all six queries),
`backend/sql/queries/courses.sql:105–111` (`LockEligibleAuthoringOwner`),
`backend/internal/identity/adapters/postgres/repository.go` (`VerifyCredentials`, `Rotate`,
`Resolve`, `ListInstructors`, `persistenceError`), `backend/internal/identity/application/identity.go`
(`Login`, `Authenticate`), `backend/internal/identity/domain/user.go` (`Active`, `NewUser`),
`backend/internal/course/adapters/postgres/authoring.go:140–210` (`Create`, `Save`),
`backend/migrations/00002_identity.sql`, `backend/sql/seed-identities.sql`, both integration
harnesses, `Makefile`, `git log fd90fc9`, `work/PRODUCT-BRIEF.md` §24 items 8/10.

**The load-bearing claim — "deleted ⊂ disabled" makes the four untouched filters correct — holds.**
Verified query by query:

| Query | Filter in the file | Under a 4b DELETE (`deleted_at=$c, disabled_at=COALESCE(disabled_at,$c), updated_at=$c`) |
|---|---|---|
| `LockActiveIdentityUser` (`identity.sql:6–10`, used by `Rotate` at `repository.go:79`) | `disabled_at IS NULL FOR UPDATE` | no row → `ErrInvalidCredentials`; no session created |
| `ResolveIdentitySession` (`identity.sql:32–46`) | `users.disabled_at IS NULL` | no row → `ErrAuthenticationRequired` |
| `ListInstructors` (`identity.sql:48–54`) | `disabled_at IS NULL` | deleted instructor absent from the picker |
| `LockEligibleAuthoringOwner` (`courses.sql:105–111`, run at `authoring.go:154,206`) | `disabled_at IS NULL … FOR SHARE` | no row → `ErrInvalidOwner` (→ `400 invalid_owner`, `course/adapters/http/admin_handler.go:535–536`) |
| `GetIdentityCredential` (`identity.sql:1–4`) | **none today**; plan adds `deleted_at IS NULL` (§5.a, T01) | with the filter: the active reuser's row; without it (4a binary): whichever row comes first, and the deleted one is refused by `!user.Active()` (`repository.go:62`, `user.go:59–61`) |

The domain side also holds: `Active()` is `DisabledAt == nil` (`user.go:59`), `Login` re-checks
`user.Active()` (`identity.go:56`), `Authenticate` re-checks it (`identity.go:84`). A deleted row
carries a non-null `disabled_at`, so every one of these refuses it under **both** the 4b and the
4a binary. The existing CHECK `disabled_at BETWEEN created_at AND updated_at`
(`00002_identity.sql:24`) is satisfied because the same `$clock` goes to `updated_at`, and a
delete-after-block keeps the earlier `disabled_at` via `COALESCE`, which is ≤ the new `updated_at`.

| # | Edition-1 finding | Closed where in edition 2 | Substance check | Verdict |
|---|---|---|---|---|
| BLK-1 | `email_taken`/B02 contradicted R12/D-3/§1 | §1 lines 47–53 ("deleted is terminal… no 4b write ever clears `deleted_at` or edits an email"); §8 lines 477–481 (code "deliberately absent"); D-3 deleted target → 404; R12 | `grep email_taken plan-4b.md` finds only the two "it is gone" sentences and §11.8; no acceptance row, no enum entry, no T04 mapping. Restore is `WHERE deleted_at IS NULL` (D-2 line 209), so `23505` on the partial index is unreachable from any 4b endpoint | **closed** |
| BLK-2 | `count(*) FOR UPDATE` invalid; "lock the others" deadlocks; loser 503 | §4 D-2 lines 228–246 (`LockAdmins :many … ORDER BY id FOR UPDATE`, whole active-or-blocked admin set, target included, count in the domain), R10 line 536 ("never 503"), §8 line 475 | (a) Valid SQL: row-returning `SELECT … FOR UPDATE` with `ORDER BY`. (b) Deadlock-free: PostgreSQL applies `ORDER BY` first and then the LockRows node, so both transactions acquire the admin rows in ascending `id`; both take `LockAdmins` **before** any other lock (the role read at step 1 is unlocked); non-admin targets use a single disjoint row (`LockUserForStatus`); `Rotate` holds exactly one users row and waits on nothing else; the course `Save` path takes `courses` then `users FOR SHARE` and the block path never touches `courses` — no wait cycle anywhere. (c) Loser gets 409: under READ COMMITTED (identity uses `pool.Begin`, `repository.go:74`; the 4b template is `Rotate`) the loser's `FOR UPDATE` waits, then re-evaluates the winner's committed row version (EvalPlanQual) — `role='admin' AND deleted_at IS NULL` still matches, and the returned row carries the new `disabled_at`, so the domain count excluding the loser's target is 0 → `last_admin`. No `40P01`, no `40001`. Also holds if the two calls happen to run sequentially, so R10 is not interleaving-flaky. One caveat is recorded as NB-2-2 below (the isolation level is relied on, not stated) | **closed** |
| BLK-3 | §7 rollback resurrected deleted accounts | §1 line 44 (DELETE "sets `deleted_at` **and** `disabled_at`"), §4 D-2 lines 205–210, §7 lines 438–445, P05 (asserts both timestamps) | Verified against the 4a code paths in the table above: with `disabled_at` set, `VerifyCredentials` refuses, `Rotate` refuses, `Resolve` refuses, the picker hides, the owner lock refuses. Restore cannot clear it (`WHERE deleted_at IS NULL`). The only write that resets `deleted_at`/`disabled_at` is the dev seed's `ON CONFLICT (id) DO UPDATE`, fixture ids only, stated in §5.e. The residual `:one` hazard after a real reuse is stated honestly in §7 and is a denial, not an admission (the deleted row is refused by `!Active()`) | **closed** |
| BLK-4 | five status-gating queries, not three; login could mint a session for a deleted user; intra-tx order | §5.a lines 304–317 (all five named; four correct by construction, one edited); §2 lines 109–111 (no `courses.sql` edit, with the reason); D-2 lines 248–262 (lock the users row **before** `DeleteIdentitySessionsForUser`); R22 | Both previously omitted queries opened and confirmed to filter `disabled_at IS NULL`. The `Rotate` interleavings in D-2 are correct for both orders (see the defense row below). The "entirely in the identity module" overreach is corrected — the course module gets one *test* file edit (R23), no query change | **closed** |
| BLK-5 | blocking an instructor silently bricks their courses | §5.h lines 371–383 (explicit decision: frozen-until-reassign/restore is the stance; 4a's lock unchanged; revisit trigger named), §5.g (published stays published; no auto-reassign), R23 (integration row in `authoring_integration_test.go`), §2 course table row | Mechanism verified: `Save` runs `LockEligibleAuthoringOwner` on `tree.InstructorID` even when unchanged (`authoring.go:206`), refused → `ErrInvalidOwner` → `400 invalid_owner`. The plan now *states* the outcome and *pins* it. The decision is a product stance, not a defect | **closed** |
| Defense row 3 | block racing an in-flight login unnamed | §1 lines 63–66, §4 D-2 lines 248–262, R22 | Login = `VerifyCredentials` (plain pool read) then `Rotate` (tx: `LockActiveIdentityUser … FOR UPDATE` → session insert). Block = tx: role read → `LockUserForStatus`/`LockAdmins` `FOR UPDATE` on the same row → `UPDATE` → `DeleteIdentitySessionsForUser`. Same row lock on both sides; each side takes it before touching `user_sessions`. Login-first: the block's per-statement snapshot (READ COMMITTED) sees the committed session row and deletes it. Block-first: `Rotate`'s `FOR UPDATE` re-evaluates the updated row, `disabled_at IS NULL` now false → `ErrNoRows` → `invalid_credentials`. Deterministic, Q1 holds; DELETE uses the identical mechanism | **closed** |

**All five blocking findings and the defense race hole are closed by location and by substance.**

### Part 2 — the four lenses on edition 2 (new defects the revision introduced)

#### Lens: AI-slop

```
[AI-slop] NB2-1  §12 marks NB-7 and NB-11 "Addressed" at placements the body does not contain
де:      §12 rows NB-7 ("noted in D-3 and §2"), NB-11 ("page-past-end is 200 {items: [], total}… §1, §2, P02")
чому:    Page-past-the-end and the `%`/`_` search-escaping contract row exist only inside the §12
         disposition cell — no §1/§2 sentence and no §9 row says them; NB-7's requested T07 row for
         the self-reset sign-out is not in §9 either (D-3 mentions it, no evidence row). The
         disposition table over-claims; an implementer reading §9 alone never tests either.
що зробити: Add two contract rows under T05 ("page past the end → 200, items: [], true total";
         "?q=% matches only a literal %") and one T07 row ("self-reset announces sign-out and lands
         on /login"); or downgrade those §12 cells to "stated, no row".
вага:    non-blocking
```

```
[AI-slop] NB2-2  "Closes three of the five blocking findings" is restated five times; §13 addresses the reviewer
де:      preamble lines 20–25; D-2 line 205–210; §5.a; §7 line 438; §11.1; §12
чому:    The same sentence recurs in five sections; §13 "Gate note" is text for the gate, not for
         the implementer, and will read as noise once the gate is green. Restating is where drift
         starts (edition 1's §11 was exactly this).
що зробити: Keep the preamble sentence and the D-2 mechanism; reduce the other three to "(deleted
         ⊂ disabled, D-2)". Delete §13 after this round.
вага:    non-blocking
```

```
[AI-slop] NB2-3  The widened interface keeps the name `PasswordVerifier` while gaining `Hash`; "the course replacer" implies a cross-module import
де:      §2 line 90; §3 line 151; §2 line 97 ("escaped search"), §12 NB-11 ("with the course replacer")
чому:    A `Verifier` that hashes is a name that lies to the next reader. "Reuse the course
         replacer" either means importing `course/adapters/postgres` from `identity/adapters/postgres`
         (a module coupling the architecture forbids) or copying one `strings.NewReplacer` line — the
         plan does not say which.
що зробити: Name the widened interface `Passwords` (or `PasswordHasherVerifier`); say "duplicate
         the one-line replacer + ESCAPE '!'" or hoist it to `platform/`.
вага:    non-blocking
```

Cited facts re-opened for this lens, all found accurate: `authoring.go:206`, `verifier.go:54`
(`encode`), `user.go:59`, `00002_identity.sql:24`, `courses.sql:1–38` (no `users` join),
commit `fd90fc9`, `Makefile:44–45` (`e2e` does not re-seed) and `:54` (`check` orders
`seed-identities` before `e2e`), `PRODUCT-BRIEF.md` §24 items 8 and 10, `plan.md` BLK-12,
`admin.tsx:13` (`canAccessAdministration`), `__root.tsx:21–22` (`admin: AdminDependencies` derived
from `ReturnType`, so `{ ...courses, users }` needs no new key — NB-18 closure holds). No
fabricated citation found.

#### Lens: logic holes

```
[logic] NB2-4  The email-reuse fixtures (B08, P05) sit in harnesses that end with `DownTo`, and §6 says that down fails on a reused email
де:      §2 course-test row (B06/B08 "on the already-populated harness"), T01 "B06/B08 assertions";
         §6 Down paragraph lines 422–426; §12 NB-3
чому:    `repository_integration_test.go:71–110` is one subtest: DownTo(0) → UpTo(2) → seed →
         Up → assertions → **DownTo(2)**. The identity `schema_integration_test.go:536` likewise
         ends with a "down and up is reversible" subtest (DownTo(0)) on the same container. Once
         B08/P05 leave a deleted row and an active row sharing one email, the 00004 Down's
         `ADD CONSTRAINT users_email_key UNIQUE (email)` raises 23505 and the whole subtest fails —
         the plan's own §6 predicts exactly this. Also, B08's stated evidence ("POST /auth/login…
         GetIdentityCredential returns the active row") cannot be produced in the course harness,
         which has no identity repository; it belongs in the identity schema test.
що зробити: Move B08 to `schema_integration_test.go (extend)` (it has `NewRepository` and
         `VerifyCredentials`); state in T01 and in the down paragraph of §6: "every test that
         creates an email-reuse pair deletes the deleted row before the harness's `DownTo`". Keep
         B06 in the course harness.
вага:    non-blocking (fails loudly at T01's own `make integration`; the fix is one DELETE)
```

```
[logic] NB2-5  R10's "never 503" silently depends on READ COMMITTED; the plan never states the isolation level
де:      §4 D-2 step 2 lines 239–242; §8 line 475; R10
чому:    The loser's `FOR UPDATE` re-reads the winner's committed row only under READ COMMITTED.
         The course module opens `BeginTx(… RepeatableRead …)` for its read paths
         (`authoring.go:76,124`); an implementer who copies that for the status transaction gets
         40001 on the loser → `persistenceError` → 503, and R10 goes red for a reason the plan
         does not name. The plan's cited template (`Rotate`, `pool.Begin`) is right by accident of
         reference, not by statement.
що зробити: One sentence in D-2: "all 4b write transactions use `pool.Begin` (READ COMMITTED);
         the re-read that yields `last_admin` is the READ COMMITTED re-evaluation of the locked
         row — never REPEATABLE READ, which would turn the loser into a 40001/503."
вага:    non-blocking
```

```
[logic] NB2-6  Password reset: lock/write order and the old-password race are unstated; Q2 has a window the plan does not name
де:      §4 D-2 lines 265–267 ("in the same transaction as SetUserPasswordHash"); T04 "set-password
         + revoke"; §1 line 45
чому:    D-2 states the users-row-before-sessions order for block/delete only. For reset, if the
         adapter deletes sessions first and updates the hash second, a `Rotate` that slips between
         the two leaves a live session created with the OLD password after the reset commits. Even
         with the right order (UPDATE first — the UPDATE is the row lock — then DELETE), a login
         whose `VerifyCredentials` matched the old hash *before* the reset and whose `Rotate` waits
         on the row will still mint a session afterwards, because `LockActiveIdentityUser`
         re-checks only `disabled_at`, not the credential. The window is one pool round trip and
         the attacker already held the old password, but "the old password stops working"
         (criterion 4) is not strictly true inside it.
що зробити: State the order for reset (UPDATE users … then DeleteIdentitySessionsForUser, same
         reasoning as D-2 for block); record the residual old-password TOCTOU as an accepted
         limitation with the revisit trigger "Rotate re-verifies the hash inside its transaction".
вага:    non-blocking
```

```
[logic] NB2-7  Restore of a blocked *admin*: D-2 leaves the lock choice ambiguous
де:      §4 D-2 step 2 heading "Last active admin — block of an admin only" vs step 3 "if the target is not an admin"
чому:    A restore on an admin target matches neither branch as written. Either lock is safe (the
         admin set or the single row — both are disjoint from or ordered with every other path),
         but the implementer must guess, and the idempotence read (NB-6) needs *some* locked
         `disabled_at`.
що зробити: One clause: "restore always uses `LockUserForStatus` regardless of role; `LockAdmins`
         is taken only for a block whose target is an admin".
вага:    non-blocking
```

#### Lens: excess complexity

This lens was run in full on edition 2, with the reuse-first question asked of every new
mechanism: "is there an existing thing, a stdlib thing, or a smaller thing?"

Checked and found **proportionate** (no finding): the `AdministrationRepository` port (a real
external dependency, four methods, mirrors `SessionStore`); `domain/administration.go` (pure
policy over a supplied set, mirrors 4a's `authoring.go`); the role-branch between `LockAdmins`
and `LockUserForStatus` (the branch is what makes the scheme deadlock-free; the alternative —
always lock the admin set — serialises every block behind every other for nothing); counting in
the domain over the locked rows rather than a second SQL count (one query, testable); the
frontend file set (eight files, but each is the layer the standards require and each mirrors an
existing 4a twin — no new pattern); three disposable fixtures (needed, as the 4a storageState trap
shows); the consumer-side `Hash` in place of a port (a genuine simplification versus edition 1);
`{ ...courses, users }` in place of a second composition root; importing `adminListStaleMs`;
plain-string ids. The edition-2 simplifications the plan claims (no port, no branded id,
extend-not-twin, consumer-side hashing) are real, verified against `admin.ts`, `__root.tsx`,
`identity.ts` and `repository.go:24–27`.

```
[complexity] NB2-8  Two status-write queries where one nullable-parameter UPDATE suffices
де:      §2 line 97 (`SetUserDisabledAt`; `RestoreUser`)
чому:    Block and restore differ only in the value written to `disabled_at`. One query —
         `UPDATE users SET disabled_at = sqlc.narg(disabled_at), updated_at = $clock WHERE id = $1
         AND deleted_at IS NULL` — serves both, and the idempotence decision already lives in Go
         over the locked `disabled_at`. Two queries is two generated functions, two names and two
         places to forget `updated_at`.
що зробити: Collapse to one `SetUserDisabledAt` with a nullable parameter; drop `RestoreUser`.
вага:    non-blocking
```

```
[complexity] NB2-9  The load-bearing invariant is carried by one COALESCE in one statement; a one-line CHECK makes it structural and shortens the argument
де:      §4 D-2 "Deleted ⊂ disabled" lines 205–210; §2 `user.go (extend)` "Active() unchanged in effect"; §6 Up
чому:    Every rollback and filter argument in §5.a/§7 rests on "the write does both". That is
         enforced only by the text of `SoftDeleteUser`; a manual `UPDATE users SET disabled_at =
         NULL` on a deleted row, or a future restore that loses its `deleted_at IS NULL` guard,
         silently re-enables a deleted account under the 4a-era filters. Making it a constraint
         costs one line and *removes* a paragraph of reasoning: `CHECK (deleted_at IS NULL OR
         disabled_at IS NOT NULL)` in 00004, and `Active()` = `DisabledAt == nil && DeletedAt ==
         nil` in the domain (two tokens; the 4a binary still keys on `disabled_at`, which the
         write guarantees).
що зробити: Add the CHECK to §6 step 2 and the `Active()` tightening to §2/T02; §7 can then
         cite the constraint rather than the statement.
вага:    non-blocking (correct as planned; this is hardening that simplifies)
```

Net: the complexity lens finds edition 2 **smaller than edition 1** and no mechanism without a
client; two consolidations above, `net: -1 query, +1 constraint`.

#### Lens: edge cases

```
[edge] NB2-10  B12's "seeded with now()" fixture fails under a fixed test clock earlier than the DB clock
де:      B12 line 521; §4 D-2 lines 212–216 (limitation "API clock ≥ DB clock")
чому:    Identity tests use fixed fake clocks. A row whose `created_at` is DB `now()` (2026-09)
         and a block whose `$clock` is a 2026-01 fake gives `updated_at < created_at` → 23514 →
         503, so B12 proves the CHECK by failing. The same limitation now also covers the new
         `deleted_at >= created_at` CHECK, which §4 states only for `disabled_at`.
що зробити: B12 must use a real clock (or a fake ≥ the DB's `now()` read in the test); extend the
         limitation sentence to `deleted_at`.
вага:    non-blocking
```

```
[edge] NB2-11  Password reset on a *blocked* target is allowed but unstated
де:      §4 D-3 row "admin, a student/instructor → 204"; §8
чому:    Nothing in D-2/D-3 says whether reset requires an active target. 204 with the block
         intact is the sensible answer (the credential changes, the block stays, nothing to revoke)
         but it is a decision, and an implementer might add `disabled_at IS NULL` to
         `SetUserPasswordHash` and produce a 404 the matrix does not list.
що зробити: One cell/row: "reset on a blocked user → 204; the block stays; on a deleted user →
         404 (R12)".
вага:    non-blocking
```

```
[edge] NB2-12  Login racing a DELETE, and DELETE racing DELETE, are covered by the same mechanism but only block has a row
де:      R22 (block only); D-2 lines 248–262
чому:    DELETE takes `LockUserForStatus` and the same order, so login-vs-delete is identical and a
         second concurrent DELETE re-evaluates `deleted_at IS NULL` to false → 404. Both follow;
         neither is asserted. With deleted ⊂ disabled being *the* decision of this edition, the
         delete half deserves its own evidence.
що зробити: Extend R22 to "block or delete", or add "R22b login racing a delete leaves no
         session; a second delete answers 404".
вага:    non-blocking
```

Edge cases checked and found **covered** by the plan text: same-state block/restore (B11);
delete-after-block keeps the earlier `disabled_at` (COALESCE, D-2); delete twice (R12);
two deleted rows sharing an email (§5.e); `make e2e` re-run after P08 (§5.f); the seed
resurrecting a deleted fixture on a persistent volume (§5.e); a deleted user's surviving cookie
(`Resolve` filters `disabled_at`, and the rows were deleted in-tx anyway); empty admin set at
the adapter (count 0 → `last_admin`); a just-blocked admin's in-flight request completing (a
per-request auth TOCTOU inherent to session auth, invariant still ≥1 active admin).

### Part 3 — consolidation

**Lenses: 4 launched / 4 returned.** AI-slop 3, logic 4, complexity 2 (+ an explicit clean list),
edge 3. No cross-lens duplicates to collapse; NB2-4 was seen from both logic (harness order) and
AI-slop (B08's evidence placement) and is recorded once.

**Blocking findings from edition 1:** BLK-1 closed, BLK-2 closed, BLK-3 closed, BLK-4 closed,
BLK-5 closed; the plan-defense race hole closed. Each verified in the code, not the prose.

**New blocking findings in edition 2: 0.** New non-blocking: 12 (NB2-1 … NB2-12). Two of them
should land before T01 starts because T01 trips over them at `make integration`: NB2-4 (delete
the reuse pair before `DownTo`; B08 to the identity harness) and NB2-5 (state READ COMMITTED).
NB2-9 is the recommended one-line hardening of the edition's load-bearing decision.

**Verdict: GREEN** — 0 blocking open; edition 2 is ready for `plan-defense`, with the two
pre-T01 non-blocking edits (NB2-4, NB2-5) applied first.

## GATE 2 round 2 — plan-defense (sonnet, fresh context)

Target: `docs/admin/plan-4b.md` — **edition 2**. Method: `plan-defense` skill, five fixed
questions re-run fresh (not inherited from round 1), each closed only by a named location
in the plan text itself; reasoning in this log does not close a row on its own.

### Code-claim spot check (the specific claim edition 2 rests row 3 on)

Claim under test: plan-4b.md §4 D-2 (paragraph "The users row is locked before sessions are
touched", lines 248–262) asserts that `Rotate`'s `LockActiveIdentityUser` is taken **inside
the session-creating transaction, before the session is created**, and that a concurrent
block/delete taking a `FOR UPDATE` lock on the same physical row therefore serialises
against it.

Opened `backend/internal/identity/adapters/postgres/repository.go` and
`backend/sql/queries/identity.sql`:

- `Rotate` (`repository.go:68–118`) does `tx, err := repository.pool.Begin(ctx)` (line 74),
  builds `queries := generated.New(tx)` (line 78), and its **first** statement in the loop
  body is `queries.LockActiveIdentityUser(ctx, userID)` (line 79) — before
  `DeleteIdentitySession` (line 88, presented-token cleanup) and before
  `CreateIdentitySession` (line 93). The transaction is committed only at line 112, after
  the session row is created. So the claim "the lock is taken before the session is
  created, in the same transaction" is **accurate**, not just plausible.
- `LockActiveIdentityUser` (`identity.sql:6–10`) reads `SELECT id::text AS id FROM users
  WHERE id = $1 AND disabled_at IS NULL FOR UPDATE` — a real `FOR UPDATE` row lock on the
  `users` row, gated on `disabled_at IS NULL`, exactly as quoted in the plan.
- On `pgx.ErrNoRows` from that lock read, `Rotate` rolls back and returns
  `domain.ErrInvalidCredentials` (lines 79–81) — matching the plan's "block commits first →
  … `Login` answers `invalid_credentials`" branch.
- The plan's own new queries (`LockAdmins`, `LockUserForStatus`,
  `DeleteIdentitySessionsForUser`) do not exist yet in `identity.sql` or anywhere in
  `backend/internal/identity` (grepped, no hits) — expected, since T04 has not landed; there
  is nothing in the current codebase that contradicts the plan's description of them.
- Postgres row-level `FOR UPDATE` locking is per-tuple, not per-WHERE-predicate-match at
  lock-acquisition time under contention: when a second transaction waits on a row a first
  transaction holds, on wake it re-evaluates the WHERE predicate against the now-committed
  row version (EvalPlanQual). That is exactly the mechanism the plan's two bullets
  (`login commits first…` / `block commits first…`, lines 256–259) describe, and it is
  standard, well-defined Postgres behaviour, not a stretch.

**Verdict on the spot check: the code claim holds.** `Rotate`'s `LockActiveIdentityUser` is
genuinely the serialising lock the plan says it is.

Also re-verified two supporting citations edition 2 leans on:
- `backend/migrations/00002_identity.sql:24` — `CHECK (disabled_at IS NULL OR disabled_at
  BETWEEN created_at AND updated_at)` — matches the line number cited at §6 step 2 / §4 D-2.
- `backend/internal/course/adapters/postgres/authoring.go:206` — `Save` (the course PUT
  path) calls `queries.LockEligibleAuthoringOwner(ctx, tree.InstructorID)` unconditionally,
  confirming §5.h's claim that the eligibility lock runs even when the owner is unchanged.

### Table

| # | Question | Closed by | Verdict |
|---|---|---|---|
| 1 | Hostile input | plan-4b.md §4 D-3 refusal matrix + R13 (malformed id → `400 invalid_request` before any repository call), R14 (illegal `status` value → `400 invalid_request`; `deleted` reachable only via `DELETE`), B04 (empty / >1024-byte / exactly-1024-byte password → `400`, `400`, `204`); `admin_handler.go`'s "DTOs, decode/size limits" row in §2 | closed |
| 2 | Dependency failure | plan-4b.md §9 R15 (`ListUsers`/`SetStatus` dependency failure → `503 dependency_unavailable`, no internal detail leaked), R16 (a hasher failure via a fake verifier → `dependency_unavailable`, distinct from a business refusal, no partial write), R17 (cancellation reported distinctly from dependency failure and refusal) | closed |
| 3 | Race / replay | **closed.** §4 D-2, the paragraph "The users row is locked before sessions are touched (BLK-4 order), which serialises block/delete against an in-flight login (the plan-defense race)" (lines 248–262), names the exact mechanism — `Rotate`'s `LockActiveIdentityUser … FOR UPDATE` taken before session creation vs. a block/delete's `FOR UPDATE` on the same row via `LockAdmins`/`LockUserForStatus`, followed in the same transaction by `DeleteIdentitySessionsForUser` — and reasons out both interleavings by name ("login commits first…", "block commits first…"), each with a deterministic, non-503 outcome. R22 (§9, line 548) is the owning acceptance row ("Login racing a block leaves no live session… no live session survives (Q1)"), owned by T04. §11 item 6 (line 606–608) points back at this same section rather than re-asserting a conclusion, closing the round-1 complaint that the summary overclaimed. The concurrent-last-two-admins sub-case remains separately closed by the `LockAdmins` deterministic-order lock (§4 D-2, R10), and the round-1 "restore hitting a now-taken email" sub-case is now moot rather than silently dropped: restore never touches `deleted_at` or `email` in edition 2 (§1, §4 D-2), so no code path exists for it to race against. The code-claim spot check above (Rotate's lock is real, taken pre-commit, inside the same transaction) confirms the mechanism the plan cites actually exists as described. | closed |
| 4 | Empty and boundary data | plan-4b.md §9 B01 (`status=deleted` matching nobody → `200 {items: [], total: 0}`), B04 (password length boundaries incl. exactly 1024 bytes), B12 (a row seeded with `now()` as `created_at`, so the `disabled_at BETWEEN` CHECK is exercised at its edge), P05/B08 (email reuse after a soft delete — the boundary this slice's schema change introduces), R09 (exactly one active admin remains — the single-element boundary for the last-admin invariant, now correctly scoped to the adapter since it is unreachable through the API) | closed |
| 5 | Rollback | plan-4b.md §7 Rollback (lines 434–457), read with §6's "Down:" paragraph. Names the supported path (redeploy previous images, leave migration 00004 in place), reasons out why a rolled-back 4a binary stays safe specifically **because** `DELETE` also sets `disabled_at` (deleted ⊂ disabled) — the fix edition 1 was missing — names the one residual data-shape hazard (a `:one` `GetIdentityCredential` on a reused email under the old binary, worst case a denied login, not a security regression), and explicitly forbids `goose down 00004` as an operational recovery step, naming why it is destructive (drops the audit column; the re-added table-wide `UNIQUE` fails outright on any reused email) | closed |

### Rows closed / open

Closed: 1, 2, 3, 4, 5. Open: none.

**Verdict: GREEN** — all five rows closed at a named plan location; the race/replay row's
closing claim (Rotate's `LockActiveIdentityUser` as the serialising lock) was independently
spot-checked against `backend/internal/identity/adapters/postgres/repository.go` and
`backend/sql/queries/identity.sql` and confirmed accurate.


## GATE 2 — constitution (inline, orchestrator, per .claude/standards.md §Конституція)

Checked plan-4b edition 2.1 against AGENTS.md architecture and TypeScript rules: 4b lives
in the identity module (domain/application/ports/adapters) plus a frontend admin-users
slice mirroring admin-course-*; domain stays stdlib-only, the invariants (self-block,
admin-immutable, last-admin) proven with unit tests; the escaped-search replacer is a
one-line duplication, NOT a cross-module import (NB2-3); generated sqlc/Orval confined to
adapters with explicit mapping; the admin authorization seam is reused verbatim, no second
seam; READ COMMITTED (pool.Begin) matches the identity module; no forbidden dependency
added (consumer-side Hash on the module's Passwords verifier, no new hasher lib); explicit
business error codes, email_taken deliberately absent with reasoning. Violations: 0.
**Constitution: PASS.**

## GATE 2 final verdict (Phase 4b)

Round 1: critique RED (5 blocking BLK-1..5) + defense RED (race hole). Edition 2 closed
all six. Round 2: critique verification GREEN (all closed by code inspection, complexity
lens delivered, 0 new blocking, 12 non-blocking) + defense GREEN (5/5, race verified in
Rotate). Edition 2.1 folded the two T01-tripping non-blockings (NB2-4 harness re-home,
NB2-5 isolation level) plus two structural wins (NB2-8 single UPDATE, NB2-9 CHECK
constraint). Constitution PASS. **GATE 2 GREEN on plan-4b edition 2.1.** Verdicts by:
fable fresh-context verifier, sonnet fresh-context examiner, orchestrator (constitution).
Plan author: delegated opus.
