# Identity administration — Phase 4b plan

Edition 2.1, 2026-09-06. Edition 2 passed GATE 2 round 2 GREEN (`docs/admin/gates-4b.md`:
plan-critique 0 blocking with the complexity lens delivered, plan-defense 5/5 with the race
verified in code). Edition 2.1 folds the twelve round-2 non-blocking findings (NB2-1 … NB2-12):
two that T01 trips over at its own `make integration` (NB2-4 harness/teardown, NB2-5 isolation
level), two structural simplifications (NB2-8 one status query, NB2-9 a schema CHECK for the
load-bearing invariant), and the rest one line each in §12. The five edition-1 blocking findings
and the defense race stay closed at the locations §12/§4 name; edition 2's history is below.

Edition 2 superseded edition 1, which GATE 2 round 1 returned RED (plan-critique 5 blocking / 22
non-blocking, plan-defense row 3 open); every one of those is closed at a named location, and §12
records the disposition of all 22 round-1 findings plus the 12 round-2 findings.

Sources: `docs/admin/spec-4b.md` (GATE 1 decisions Q1–Q5 + capability model, binding),
`docs/admin/plan.md` (4a — the seam and quality bar this reuses; the canary/BLK-12 lesson landed
in commit `fd90fc9`), `docs/admin/report.md` (4a — the lockfile and storageState traps),
`work/CTO-RESEARCH.md` (fact #1: server authorization exists), `work/PRODUCT-BRIEF.md`
§24.8/§24.10 (the 4b shape and the additive capability model). Every code claim was re-opened
against the files; where it and the briefing disagreed, the code won.

Mode: **ponytail-full** — the simplest thing that works, the 4a seam and idioms reused before
anything new, every deviation named with its revisit trigger. 4b is deliberately smaller than 4a:
no aggregate write, no owner picker, no public surface delta, no new runtime dependency, and — as
of this edition — **no new port and no branded id**, matching the module's own idioms.

**The one structural decision of edition 2.** Soft delete makes a user **deleted *and* disabled
in the same write** (`deleted_at` set, `disabled_at = COALESCE(disabled_at, $now)`); *deleted ⊂
disabled*. This single choice closes three of the five blocking findings at once: every existing
`disabled_at IS NULL` filter in **both** modules already refuses a deleted user (BLK-3 rollback,
BLK-4's two omitted queries), and the login `Rotate` lock already serialises against it (the
defense race). It is stated in §1 and §4 D-2 and carried through §5–§7.

**Scope.** Exactly `spec-4b.md` S1–S8: **block, restore, soft-delete, admin password reset**,
plus the directory that lists their targets. `spec-4b.md`'s out-of-scope list stands unchanged —
notably **role change** (Q3), **user creation / self-registration / self-service reset**,
**per-course / enrollment-scoped blocking**, an **audit-log UI**, and **any mailer**. 4b changes
no role and adds no public surface; it touches the course module only to *reason about* a
consequence of blocking an instructor (§5.h, BLK-5), not to change course behaviour.

---

## 1. Shape of the slice

Four operations on one new collection, all admin-only, all riding the **same** seam 4a built.

| Method | Path | Who | Answer |
|---|---|---|---|
| GET | `/api/v1/admin/users` | admin only | a page of directory rows (id, email, displayName, role, status, updatedAt), status-filtered, searchable, paged like the course list |
| POST | `/api/v1/admin/users/{id}/status` | admin only | `200`; body `{status: "active" \| "blocked"}` toggles `disabled_at`; block revokes the user's sessions in the same transaction (Q1); idempotent (§4 D-3) |
| DELETE | `/api/v1/admin/users/{id}` | admin only | `204`; sets `deleted_at` **and** `disabled_at`, revokes the user's sessions in the same transaction (Q1) |
| POST | `/api/v1/admin/users/{id}/password` | admin only | `204`; sets a fresh Argon2id credential and revokes the user's sessions (Q2) |

Four operations across three path templates. Block/restore is one `POST …/status` endpoint
carrying a two-state transition, mirroring the 4a course-status idiom (one refusal matrix in one
place). **Deleted is terminal**: it is reached only through `DELETE`, and **no 4b write ever
clears `deleted_at` or edits an email** — restore only clears `disabled_at`, and only for a
non-deleted (blocked) user. This is why edition 1's `email_taken` code is gone (BLK-1): no 4b
operation can add a second active row for one email, so the partial-unique-index violation it
mapped is unreachable.

**Status is a projection over two timestamps.**
`deleted_at IS NOT NULL → deleted`; else `disabled_at IS NOT NULL → blocked`; else `active`.
Because delete also sets `disabled_at`, a deleted row is *both* deleted and disabled, and the
projection gives `deleted_at` precedence so it still reads as **deleted**, not blocked. No
`status` column is added; the two timestamps are the state, so `domain.User` gains the second one
(`DeletedAt`, §2/T02, closing NB-2). The directory's **default filter shows only non-deleted users
(active + blocked); deleted users appear only under the explicit `deleted` filter** (Q4, NB-11).

**No optimistic token.** A block or delete is a single-row state flip; its only real concurrency
hazards — two admins racing the last-admin invariant, and a login in flight during a block — are
closed by the users-row / admin-set locks of §4 D-2, not by a client token. Trigger to add one:
an operation on this collection that carries more than a status flip (there is none in 4b).

---

## 2. Placement (Розміщення)

Per `.claude/standards.md` §Розміщення. 4b lives in the **identity** module plus a **frontend
admin users slice** that **extends** the 4a `admin-*` files rather than twinning them (NB-16–NB-18).
Extended files are marked `(extend)`; sibling `*_test.go` / `*.test.ts(x)` accompany every new
module.

### Backend — identity module

| File | Contents |
|---|---|
| `backend/migrations/00004_identity_admin.sql` | `users.deleted_at timestamptz` (nullable) + its CHECK; **drop the inline `users_email_key`**; `CREATE UNIQUE INDEX users_email_active_key … WHERE deleted_at IS NULL`; down migration (§6) |
| `backend/internal/identity/domain/user.go` (extend) | `DeletedAt *time.Time`; `NewUser` bound `deleted_at IS NULL OR deleted_at >= created_at`; `Active()` unchanged in effect (deleted rows are also disabled, so `DisabledAt == nil` is already false) — closes NB-2 |
| `backend/internal/identity/domain/administration.go` | `AccountStatus` projection (active/blocked/deleted) over the two timestamps; the invariant policy `CanBlock`/`CanDelete` — self-target (S6), admin-immutable delete (S6), last-admin block over a **supplied admin set** (Q5); password bounds |
| `backend/internal/identity/domain/administration_test.go` | table tests for the policy and the projection |
| `backend/internal/identity/domain/errors.go` (extend) | the four new business errors (§8) |
| `backend/internal/identity/application/administration.go` | `Administration` use cases `ListUsers`/`SetStatus`/`Delete`/`ResetPassword`; bounded deadlines and cancellation, reusing `operationTimeout`, `storeResolution`, `Clock` |
| `backend/internal/identity/application/administration_test.go` | fake-port tests: policy wiring, dependency failure, cancellation |
| `backend/internal/identity/ports/administration.go` | `AdministrationRepository` (list; transactional set-status, soft-delete, set-password — each returning the facts the domain needs). **No `PasswordHasher` port** — hashing stays consumer-side (§3, NB-19) |
| `backend/internal/identity/adapters/argon2id/verifier.go` (extend) | exported `Hash(password string) (string, error)` with a `crypto/rand` 16-byte salt, reusing the private `encode`; the package can currently only verify |
| `backend/internal/identity/adapters/postgres/repository.go` (extend) | **rename** the consumer-side `PasswordVerifier` interface (declared here, injected via `NewRepository`) to `Passwords` and add `Hash` — a `Verifier` that hashes is a name that lies (NB2-3); the administration methods hang off the existing `*Repository` (it already holds the pool and the verifier), so `AdministrationRepository` is satisfied with no new constructor |
| `backend/internal/identity/adapters/postgres/administration.go` | filtered/paged/escaped list; the block/restore/delete/reset transactions of §4 D-2; the `AdminUser` mapper |
| `backend/internal/identity/adapters/postgres/administration_integration_test.go` | Testcontainers: session death on block and delete, the last-admin lock under the R10 race, deleted-also-disabled, login-vs-block race, password round-trip, cancellation, hasher-failure via a fake verifier |
| `backend/internal/identity/adapters/postgres/schema_integration_test.go` (extend) | the migrated schema's new column, CHECK and partial index; the direct `insertUser` fixtures still satisfy them (all `deleted_at IS NULL`); **owns B08** — the email-reuse login evidence lives here because this harness has `NewRepository`/`VerifyCredentials`; its subtest **deletes the deleted row before this file's terminal reversibility `DownTo`**, so the 00004 Down never meets a reused email (NB2-4) |
| `backend/internal/identity/adapters/http/admin_handler.go` | 4b DTOs, decode/size limits, the admin-role gate via `CanAdministerAccounts()`, error mapping, route registration reusing `Authenticate`/`RequireSameOrigin` |
| `backend/internal/identity/adapters/http/admin_handler_test.go` | contract tests against `openapi.yaml`; the non-admin refusal matrix |
| `backend/internal/identity/adapters/http/handler.go` (extend) | `NewHandler` gains the `Administration` dependency; `Register` wires the four routes (mutations `RequireSameOrigin(Authenticate(…))`, the list `Authenticate(…)`) |
| `backend/sql/queries/identity.sql` (extend) | `ListAdminUsers` (search on email + display_name, escaped by a **duplicated** one-line `strings.NewReplacer("!","!!","%","!%","_","!_") + ESCAPE '!'` — the same shape as the course adapter, copied not imported, no cross-module dependency, NB2-3); `LockAdmins` (§4 D-2); `LockUserForStatus`; **one** `SetUserDisabledAt` with a **nullable** `disabled_at` parameter serving block (non-null `$clock`) and restore (NULL) — `RestoreUser` is collapsed into it, since block and restore differ only in the value written and idempotence already lives in Go over the locked `disabled_at` (NB2-8, `net: -1 query`); `SoftDeleteUser` (sets `deleted_at`, `disabled_at`, `updated_at`); `SetUserPasswordHash`; `DeleteIdentitySessionsForUser`; **and `deleted_at IS NULL` added to `GetIdentityCredential` only** (§5.a) |
| `backend/openapi/openapi.yaml` (extend) | 3 path templates / 4 operations, admin-user schemas, the four new error codes |
| `backend/sql/seed-identities.sql` (extend) | `DO UPDATE` also resets `deleted_at = NULL`; three disposable e2e fixtures `block-me@` / `delete-me@` / `reset-me@` no spec authenticates as (§5.f) |
| `backend/cmd/api/main.go` (extend) | constructs the `Administration` use case and passes it to `NewHandler` |

### Backend — course module (consequence only, no behaviour change)

| File | Contents |
|---|---|
| `backend/internal/course/adapters/postgres/repository_integration_test.go` (extend) | add the real `00004_identity_admin.sql` to the failing-migration `fstest.MapFS` and **renumber the canary `00004_failure.sql` → `00005_failure.sql`** (§5.b, the `fd90fc9`/BLK-12 pattern); assert **B06 only** on the already-populated harness — **B08 does not live here** (this subtest ends in `DownTo`, and the 00004 Down fails on a reused email; and the course harness has no identity repository to run a login through, NB2-4) |
| `backend/internal/course/adapters/postgres/authoring_integration_test.go` (extend) | one assertion for R23: a PUT on a course whose owner is blocked/deleted answers `invalid_owner` (§5.h, BLK-5) — an observation of existing 4a behaviour, not a change to it |

`courses.sql` is **not** edited: with *deleted ⊂ disabled*, `LockEligibleAuthoringOwner`'s existing
`disabled_at IS NULL` already refuses a deleted owner (§5.a). This keeps 4b's write surface in the
identity module and corrects edition 1's overreach.

### Frontend — the admin users slice (extends `admin-*`)

| File | Contents |
|---|---|
| `frontend/src/domain/identity/administration.ts` | `AdminUser`, `AccountStatus`, `AdministrationError`, password validator — **plain string ids** (NB-17) |
| `frontend/src/domain/identity/identity.ts` (extend) | `canAdministerAccounts` predicate (admin only) — the frontend twin of the domain rule (NB-4) |
| `frontend/src/ports/AdminUserRepository.ts` | the administration port |
| `frontend/src/application/admin/user-administration.ts` | use cases over the port, returning `Result` |
| `frontend/src/adapters/api/admin-user-schema.ts` | zod schemas (no branding) |
| `frontend/src/adapters/api/admin-user-repository.ts` | Orval calls through `createApiTransport`; business-code mapping (the `session-repository.ts` idiom) |
| `frontend/src/app/composition/admin.ts` (extend) | `createAdminDependencies` returns `{ ...courses, users }`, so no new `RootRouteContext` key is needed (NB-18) |
| `frontend/src/app/composition/admin-user-queries.ts` | query options, the four mutations, `invalidateAfterUserWrite`; **imports `adminListStaleMs`** from `admin-queries.ts` (the 4a D-2c trigger has fired — import, do not mint a third constant; NB-16) |
| `frontend/src/app/router/admin-search.ts` (extend) | add the user status enum and `q`; the never-throwing per-field parsing is already there (NB-18) |
| `frontend/src/presentation/admin/UserListView.tsx` | directory table, status tabs, search, pagination, per-row block/restore/delete, empty/error/degraded states |
| `frontend/src/presentation/admin/ResetPasswordDialog.tsx` | the one 4b form (a single password field) — React Hook Form, already installed by 4a; announces sign-out when the target is the caller (NB-7) |
| `frontend/src/routes/admin.users.index.tsx` | the directory route with a `beforeLoad` throwing `notFound()` for a non-admin, so an instructor gets not-found, not an error boundary (NB-4) |
| `frontend/src/presentation/identity/IdentityHeader.tsx` (extend) | an admin-only link to `/admin/users` — the entry point (NB-4) |
| `frontend/src/app/providers/locales/uk.ts`, `en.ts` (extend) | the users vocabulary; keys added only |
| `frontend/e2e/admin-users.spec.ts` | the browser journey on the disposable fixtures, reusing the 4a shared `storageState` |
| `frontend/e2e/route-guards.spec.ts` (extend) | the `/admin/users` guard and the direct-API non-admin refusal |

The `/admin` route is already a guarded layout with an `<Outlet/>` (4a A10); 4b adds a child route.

### Infrastructure

Nginx, Compose, `.env.example`, Prometheus, the Makefile/CI gate order and the media-worker are
**unchanged** — 4b adds no rate-limited path, no large body and no new gate.

---

## 3. Reuse-first

| Concern | Reused | Why nothing new |
|---|---|---|
| Server authorization (authenticate → role → same-origin) | `Handler.Authenticate`, `Handler.RequireSameOrigin`, `httpserver.Actor` | the 4a seam (D-1); 4b is its second consumer. A second seam is what the spec forbids |
| Admin-role predicate | `domain.Role.CanAdministerAccounts()` | already admin-only, already the `GET /admin/instructors` gate |
| Authenticated caller inside the handler | the existing `userKey{}` context value carrying the full `domain.User` | 4b handlers are in the identity http package and read the same value `me`/`listInstructors` read |
| Same-origin (CSRF) refusal | the same `RequireSameOrigin` | block/delete/reset are cookie-authenticated state changes |
| Password verify/**hash** placement | the **consumer-side** password interface in `adapters/postgres/repository.go`, injected via `NewRepository` | the module already keeps its password interface consumer-side and out of `ports/` (AGENTS.md: ports require an *external* dependency; argon2 is in-process). 4b widens it with `Hash` — renaming `PasswordVerifier` → `Passwords` since it no longer only verifies (NB2-3) — rather than inventing a `ports.PasswordHasher` (NB-19) |
| Argon2id parameters and encoding | `internal/identity/adapters/argon2id` | the exact `m=64MiB,t=3,p=2` cost and `$argon2id$…` encoding; 4b adds a random-salt entry point, it picks no new parameters |
| Session revoke primitive | the session table and the `DELETE … user_sessions` idiom; `Rotate`'s `tx.Begin → work → Commit` shape | 4b adds one query (`DeleteIdentitySessionsForUser`) and follows the transactional template |
| Bounded deadlines, cancellation, microsecond clock | `operationTimeout`, `storeResolution`, `Clock` | the I03 precision lesson and the 3s timeout are the module idiom |
| Contract-test / repository harnesses | `identity/adapters/http/handler_test.go`, `identity/adapters/postgres/schema_integration_test.go` | new paths and the migrated schema plug into existing validators |
| Frontend admin idioms | `admin-course-repository.ts`, `admin-queries.ts` (the `adminListStaleMs` constant), `admin-search.ts`, `CourseListView.tsx`, the required-label kit, the shared `storageState` | the users slice extends the course slice; a ≤50-row server-filtered table is exactly the course list |
| React Hook Form | 4a's installed dependency | the reset dialog is the second form family; still no `@hookform/resolvers` (the domain validator is the resolver) |

**What is genuinely new, and why the existing thing did not fit:**

- **`argon2id.Hash` — the first credential-writing code in the product.** `argon2id` today only
  *verifies*: `encode` is unexported and used solely for the dummy hash (`verifier.go:54`), and
  every real hash is seeded as a literal. Password reset (S5) is the first operation that writes a
  credential, so it needs `Hash` with a `crypto/rand` salt. This reuses the *parameters and
  encoding* but is a new *capability* — exposed on the existing consumer-side interface, not a new
  port.
- **`AdministrationRepository`** — the login/authentication ports are login/authentication ports;
  listing users by status and flipping status inside a session-revoking transaction are write
  concerns that do not belong on them. This *is* a real external dependency (Postgres), so it lives
  in `ports/` as `SessionStore`/`CredentialVerifier` do.
- **`domain/administration.go`** — `user.go` describes a valid user; the block/delete invariants
  are policy over a *set* of users, a distinct file in the same package, mirroring how 4a put
  `authoring.go` beside the catalog invariants.
- **`DeleteIdentitySessionsForUser`** — the existing session deletes are per-digest; block/delete/
  reset revoke *all* of one user's sessions.

---

## 4. Server enforcement design

### D-1 The seam is reused verbatim

4b registers from **inside** the identity http package, so it needs no cross-module wiring:

- Reads: `mux.Handle("GET /api/v1/admin/users", handler.Authenticate(http.HandlerFunc(handler.listUsers)))`.
- Mutations: `RequireSameOrigin(Authenticate(…))` for `POST …/status`, `DELETE …`, `POST …/password`.
- **Order per mutation** (4a NB-8): same-origin first, then authenticated, then role
  (`CanAdministerAccounts()` → non-admins get `403 forbidden`, the shape `listInstructors` already
  returns), then target exists, then the domain invariant. Every mutation acceptance row states a
  valid `Origin`; the foreign-`Origin` case is its own row and answers `403 origin_forbidden`
  before the cookie is read.
- The handler reads the caller's `domain.User` from `userKey{}` (the value `Authenticate` already
  stores). No policy in the handler beyond the role gate; the invariants live in the domain.
  (`httpserver.Actor` is written by `Authenticate` regardless of 4b; 4b does not touch it — so it
  is not listed as a 4b reuse, correcting an edition-1 decorative claim, NB-20.)

### D-2 The transactional invariants — the exact locks

The three refusals in S6/Q5 are domain rules over data that must be read transactionally, so they
follow 4a's D-4a placement: the adapter opens a transaction, takes the locks below, the **domain**
decides over the locked facts, and the adapter commits or rolls back. Two edition-1 defects are
fixed here: `SELECT count(*) … FOR UPDATE` is invalid PostgreSQL (BLK-2), and the "lock the other
admins" scheme deadlocks (BLK-2).

**Deleted ⊂ disabled — enforced by the schema, not just by the write (NB2-9).** Every `DELETE`
writes, in one statement, `deleted_at = $clock, disabled_at = COALESCE(disabled_at, $clock),
updated_at = $clock`, and the migration adds `CHECK (deleted_at IS NULL OR disabled_at IS NOT
NULL)` (§6), so the invariant is a **constraint**: a manual `UPDATE users SET disabled_at = NULL`
on a deleted row, or a future restore that lost its `deleted_at IS NULL` guard, is rejected by the
database rather than silently re-enabling a deleted account under the 4a-era filters. `Active()`
becomes `DisabledAt == nil && DeletedAt == nil` (two tokens; the 4a binary still keys on
`disabled_at`, which the write and now the CHECK guarantee). This is what makes every
`disabled_at IS NULL` filter in both modules refuse a deleted user (§5.a) and makes rollback safe
(§7, BLK-3). Restore only ever clears `disabled_at` on a **non-deleted** row
(`WHERE deleted_at IS NULL`), so a deleted user is never un-disabled.

**Every 4b write moves `updated_at` to the same `$clock`** it writes into
`disabled_at`/`deleted_at`/`password_hash`, so the `disabled_at BETWEEN created_at AND updated_at`
and `deleted_at >= created_at` CHECKs (`00002_identity.sql:24`, §6) hold; `$clock` is the
application `Clock`, µs-truncated (the I03 lesson). Recorded limitation: the API clock must be ≥
the DB clock that stamped `created_at`, for **both** `disabled_at` and `deleted_at` (NB-10,
NB2-10) — the reason B12 uses a real clock, not a fixed fake earlier than the DB's `now()`.

**Isolation level: READ COMMITTED (NB2-5).** All 4b write transactions use `pool.Begin` (READ
COMMITTED), matching the identity module's `Rotate`. The `last_admin` re-read that yields a 409 is
the READ COMMITTED re-evaluation of the locked row by the waiting transaction — **never**
`BeginTx(… RepeatableRead …)` as the *course* module uses for its reads (`authoring.go:76,124`):
copying that here would abort the loser with `40001`, which `persistenceError` maps to `503`, and
R10 would go red for a reason the plan did not name. This is a copy trap; the template to follow is
`Rotate` (`pool.Begin`), not the course adapter.

**Self-target (S6), decided before any lock.** The caller's id is known; `CanBlock`/`CanDelete`
refuse when target == caller → `cannot_target_self`. This one predicate covers **both** block-self
and delete-self, so there is no separate self-delete rule and the code is named for what it does,
not for "block" (NB-21).

**Admin-immutable delete (S6).** `DELETE` targeting an `admin` role is refused → `admin_immutable`,
**by anyone**. Role never changes in 4b (Q3), so role is read without a lock; a `DELETE` therefore
never needs the admin-set lock, and its target is always a non-admin. This subsumes "delete the
last admin".

**Last active admin (Q5) — block of an admin only.** Restore never removes an admin and delete
never targets one, so the last-admin rule is exercised only by **blocking an admin**. The adapter:

1. Reads the target's `role` **without a lock** (role is immutable). Missing / `deleted_at IS NOT
   NULL` → `user_not_found`.
2. **If the target is an admin:** `LockAdmins :many` —
   `SELECT id::text, disabled_at FROM users WHERE role = 'admin' AND deleted_at IS NULL ORDER BY id
   FOR UPDATE`. This locks the **whole** active-or-blocked admin set, target included, in a
   **deterministic id order**, so two concurrent admin-blocks acquire the lowest-id lock first and
   serialise — no deadlock (edition 1's "lock the *other* admins" deadlocked: Tx1 held B wanting A,
   Tx2 held A wanting B). The domain counts admins with `disabled_at IS NULL` **excluding the
   target**; zero → `CanBlock` refuses `last_admin`. The winner blocks; the loser, waking after the
   winner commits, re-reads the set (the just-blocked admin now has `disabled_at` set, so it drops
   out of the active count) and refuses `last_admin` — **a 409, never a 503** (no aborted
   transaction, no `40P01`). R10 asserts exactly this.
3. **If the target is not an admin:** `LockUserForStatus :one` —
   `SELECT id::text, disabled_at FROM users WHERE id = $1 AND deleted_at IS NULL FOR UPDATE` — locks
   the single target row (no admin-set lock; a non-admin cannot deadlock against the admin set,
   whose rows are disjoint).

**Restore always uses `LockUserForStatus` regardless of role; `LockAdmins` is taken only for a
block whose target is an admin (NB2-7).** A restore never removes an active admin, so it needs no
last-admin count and no set lock — the single-row `FOR UPDATE` gives it the locked `disabled_at`
the idempotence read needs and serialises it against `Rotate` like every other write.

**The users row is locked before sessions are touched (BLK-4 order), which serialises block/delete
against an in-flight login (the plan-defense race).** Login is two pool calls: `VerifyCredentials`
(a plain read) then `Rotate`, whose `LockActiveIdentityUser` takes
`… WHERE id = $1 AND disabled_at IS NULL FOR UPDATE` on the user row before creating the session.
A block/delete takes a `FOR UPDATE` lock on the **same** row (via `LockAdmins` for an admin target,
or `LockUserForStatus` otherwise) and then, *after* the `UPDATE`, runs
`DeleteIdentitySessionsForUser`. The two therefore serialise on that row:

- login commits first → `Rotate` created a session; the block's `UPDATE` + `DeleteIdentitySessions
  ForUser` (same tx, in that order) then delete it → **no live session survives** (Q1);
- block commits first → `Rotate`'s `LockActiveIdentityUser` now finds `disabled_at IS NOT NULL`,
  returns no row, and `Login` answers `invalid_credentials` → **no session is ever created**.

Either way the outcome is deterministic and Q1 holds. **`DELETE` follows the identical mechanism**
(it takes `LockUserForStatus` and the same UPDATE-then-revoke order), so login-vs-delete is the
same story, and a second concurrent `DELETE` re-evaluates `deleted_at IS NULL` to false and answers
`404`. R22 covers both block and delete, and the double-delete 404 (NB2-12). (This is the race §11
of edition 1 wrongly called impossible; §11 is now a pointer list, NB-15.)

**Block/delete revoke sessions in the same transaction (Q1)**, belt-and-braces with
`ResolveIdentitySession`'s `disabled_at IS NULL` filter, which already makes a surviving cookie
inert. **Password reset revokes all sessions (Q2)** in the same transaction as
`SetUserPasswordHash`, **in that order — `SetUserPasswordHash` (the row-locking UPDATE) first, then
`DeleteIdentitySessionsForUser`** (NB2-6), the same order and reasoning as block/delete: a `Rotate`
that slips in serialises on the row and either lands before the UPDATE (its session is then
revoked) or after (its `LockActiveIdentityUser` still passes, but the hash it verified is already
gone). Reset works on a **blocked** target too → `204`, the block stays, and there is simply
nothing to un-revoke (NB2-11); only a **deleted** target answers `404` (R12). **Accepted residual
(NB2-6):** `LockActiveIdentityUser` re-checks only `disabled_at`, not the credential, so a login
whose `VerifyCredentials` matched the *old* hash before the reset and whose `Rotate` waits on the
lock can still mint one session with the old password after the reset commits — a one-round-trip
TOCTOU by a party who already held the old password. Criterion 4 ("the old password stops working")
holds outside this window; the revisit trigger is "`Rotate` re-verifies the hash inside its
transaction".

**Same-state is idempotent (NB-6).** Block on an already-blocked user, or restore on an active
user, answers `200`, touches neither `disabled_at`, `updated_at` nor sessions, and **skips the
last-admin count** (so re-blocking a blocked admin cannot trip a spurious `last_admin`). The
adapter detects the no-op from the locked `disabled_at` it already holds. B11 asserts the
timestamp is unchanged.

### D-3 The refusal matrix

Every cell is server-enforced and contract-proven; nothing relies on a client guard (4a R33).
Mutation cells assume a valid `Origin`; a foreign one is refused first.

| Caller | list | status (block/restore) | delete | password |
|---|---|---|---|---|
| anonymous | 401 `authentication_required` | 401 | 401 | 401 |
| disabled account | 401 | 401 | 401 | 401 |
| student | 403 `forbidden` | 403 | 403 | 403 |
| instructor | 403 `forbidden` | 403 | 403 | 403 |
| admin, self | — | `cannot_target_self` (409) | `cannot_target_self` (409) | 204 (an admin may reset their own password — it signs them out, NB-7) |
| admin, another admin | row shown | 200 (a *sequential* call never hits the last admin — the caller is one; `last_admin` is the **race** outcome, R10) | `admin_immutable` (409) | 204 |
| admin, a student/instructor | row shown | 200 | 204 | 204 |
| admin, unknown or **deleted** id | — | `user_not_found` (404) | `user_not_found` (404) | `user_not_found` (404) |
| any, foreign `Origin` | — | `origin_forbidden` (403) | `origin_forbidden` (403) | `origin_forbidden` (403) |

An **instructor is refused on every 4b operation** (4b is admin-only, unlike 4a). A **deleted**
target answers `user_not_found` on every operation but still appears in the directory under the
`deleted` filter (listing is a read, not an operation). Malformed id → `400 invalid_request`
before any repository call (4a R07). An illegal `status` value → `400 invalid_request`; `deleted`
is reachable only via `DELETE` (R14). A sequential `last_admin` is **not** reachable through the
API — the caller is always one active admin, so blocking anyone else leaves ≥1; R09 therefore
exercises the invariant at the **adapter** level with a constructed single-admin state (NB-1).

---

## 5. Side effects (Сайд-ефекти)

### a. Only `GetIdentityCredential` needs a new filter (deleted ⊂ disabled does the rest)

Five queries gate on account status. With *deleted ⊂ disabled* (D-2; now a schema CHECK, §6), four
are already correct:

- `LockActiveIdentityUser` (login `Rotate`), `ResolveIdentitySession`, `ListInstructors`
  (owner picker), and the course module's `LockEligibleAuthoringOwner` all filter
  `disabled_at IS NULL`; a deleted row is disabled *by constraint*, so each already refuses it.
  **No `courses.sql` edit, no cross-module query change** (correcting edition 1's BLK-4 overreach).
- **`GetIdentityCredential` gains `AND deleted_at IS NULL`.** It is `:one` = pgx `QueryRow` +
  `Scan`, which on two matching rows returns **whichever the plan yields first, silently** (not an
  error — correcting edition 1's stated mechanism, NB-5). After an email is reused, one row is the
  deleted user and one the active user; without the filter, login could authenticate-attempt
  against the deleted (disabled) row and wrongly refuse the *active* user. The filter guarantees
  the active row is the one returned. This ships in the **same task** as the migration (T01).

### b. The migration collides with the course test canary

Next real migration is `00004`. The course `repository_integration_test.go` failing-migration
subtest builds an `fstest.MapFS` of `00001`+`00002`+`00003`+ a synthetic **`00004_failure.sql`**.
A real `00004` collides, so T01 adds the real `00004_identity_admin.sql` to that MapFS and
**renumbers the canary to `00005_failure.sql`** — the pattern 4a used for `00003`
(`docs/admin/plan.md` BLK-12; commit `fd90fc9`, NB-22). The identity `schema_integration_test.go`
reads via `os.DirFS`, so it picks up `00004` automatically.

### c. Argon2id gains a hashing entry point

`argon2id.Hash` is added beside `Verify`/`VerifyDummy`, reusing the private `encode` with a
`crypto/rand` 16-byte salt. The consumer-side interface in `adapters/postgres/repository.go`
(renamed `PasswordVerifier` → `Passwords`, NB2-3) is widened with `Hash`; `NewRepository` already
injects it, so no wiring changes. A round-trip unit test covers `Hash`→`Verify`.

### d. `NewHandler` signature grows

The identity http `Handler` gains an `Administration` dependency; `NewHandler` takes one more
argument and `main.go` constructs the use case. Existing identity tests pass a fake; their
behaviour assertions are unchanged.

### e. `seed-identities.sql` resets `deleted_at`; the partial-index caveat

`ON CONFLICT (id) DO UPDATE` already resets `disabled_at`; it must also reset `deleted_at = NULL`,
or a soft-deleted fixture stays deleted forever on a persistent dev volume. **Caveat (NB-9):** the
seed is one transaction, so if a *different* active row already holds `delete-me@` (reachable only
by a manual insert), the `DO UPDATE` raises `23505` on `users_email_active_key` and `make check`
fails wholesale; and the partial index permits *two deleted rows* to share an email, which the
directory would then show twice. Manual recovery: delete the stray row. CI is always fresh, so
this bites only the dev loop.

### f. e2e must not touch the shared fixtures, and re-run needs a re-seed

4b mutates accounts; the 4a shared `storageState` signs specs in as the three fixed fixtures.
Blocking/deleting/resetting one of them would revoke the shared session mid-run (the 4a report §3
trap, generalised). 4b seeds **three disposable fixtures** (`block-me@`, `delete-me@`, `reset-me@`,
all `student`, none authenticated by any spec); `admin-users.spec.ts` acts only on those, using the
shared admin session (which no 4b op touches — an admin cannot delete an admin, block themselves,
and the spec never resets the shared admin). **Zero new logins**, so the 4a D-7 budget is
untouched. **Re-run caveat (NB-13):** `make e2e` alone does not re-seed; after one run `delete-me@`
is deleted-and-terminal, so a bare second run fails at the delete step. `make check` orders
`seed-identities` before `e2e`, so only the dev loop is hit; the spec asserts each disposable's
pre-state in `beforeAll` with a message naming the target and telling the reader to
`make seed-identities`.

### g. Deleted/blocked instructors' courses stay published under their id

Public catalog queries never join `users` (verified `courses.sql:1–38`), so blocking or deleting
an instructor leaks nothing. Their courses **keep their state and owner**; reassignment is an admin
action, never automatic (NB-12).

### h. Blocking or deleting an instructor freezes editing of their courses (BLK-5) — accepted

`LockEligibleAuthoringOwner` (`disabled_at IS NULL`) runs on **every** course PUT, even when the
owner is unchanged (`authoring.go:206`); with *deleted ⊂ disabled* it refuses both blocked and
deleted owners. All five demo courses belong to the instructor fixture, so once that instructor is
blocked or deleted, **every PUT on their courses — an admin's title edit included — answers
`400 invalid_owner` until an admin reassigns the owner.** **Decision: this is the intended
stance** — a blocked/deleted owner's courses are frozen (published ones stay published) until
restore or reassignment; nothing is reassigned automatically. 4b does **not** weaken 4a's
eligibility lock (that would be an out-of-scope course-module behaviour change). Trigger to
revisit: if operators find frozen courses painful, a later slice lets a PUT that preserves the
current owner skip the eligibility check. R23 pins the current behaviour so a future change is a
deliberate one.

### i. Unchanged on purpose

Nginx (4b rides the existing `api` zone; the largest body is one password), Compose,
`.env.example`, Prometheus, the Makefile/CI gate **order**, the media-worker, the public catalog,
and SSR (admin routes are `ssr:false`; the SSR container has no `Origin` and could not call a 4b
mutation).

---

## 6. Migration plan (План міграції)

`backend/migrations/00004_identity_admin.sql`, goose, applied by `make up`/`make migrate` before
any seed.

**Up:**

1. `ALTER TABLE users ADD COLUMN deleted_at timestamptz;` — nullable; no backfill, every existing
   row is non-deleted.
2. `ALTER TABLE users ADD CONSTRAINT users_deleted_at_valid CHECK (deleted_at IS NULL OR deleted_at >= created_at);`
   — a delete sets `deleted_at`, `disabled_at` and `updated_at` to one `$clock`, so a `>= created_at`
   bound suffices and cannot fight the write; the existing `disabled_at BETWEEN created_at AND
   updated_at` CHECK is satisfied because that same `$clock` is written to `updated_at`.
3. `ALTER TABLE users ADD CONSTRAINT users_deleted_implies_disabled CHECK (deleted_at IS NULL OR disabled_at IS NOT NULL);`
   — **the load-bearing invariant made structural** (NB2-9): a deleted row must also be disabled,
   so *deleted ⊂ disabled* is enforced by the database, not only by `SoftDeleteUser`'s text. Every
   existing row has `deleted_at IS NULL`, so it passes trivially.
4. `ALTER TABLE users DROP CONSTRAINT users_email_key;` — the inline `email … UNIQUE` in
   `00002_identity.sql` created this implicitly-named constraint (referenced by no code;
   `uniqueCollision` names only session constraints).
5. `CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL;` —
   uniqueness now holds **only among non-deleted users** (S4).

**How existing rows and the seed survive it.** Every current row has `deleted_at IS NULL`, so the
partial index covers all of them and enforces exactly what the dropped `users_email_key` did; no
active pair shares an email, so index creation cannot fail. On a fresh volume / CI the table is
empty and steps 1–5 are trivially safe. Between the drop (step 4) and the index (step 5) there is a
sub-millisecond window with no email uniqueness, inside the migration's transaction, so no
concurrent writer sees it.

**Cost.** Drop-then-partial-index is the cheapest expression of "unique among non-deleted"; the
alternatives (a synthetic column, a computed uniqueness) add a column for nothing.

**Down:** drop `users_email_active_key`; drop both CHECKs; drop `deleted_at`; re-add
`ALTER TABLE users ADD CONSTRAINT users_email_key UNIQUE (email)`. The re-add **fails if any email
was reused** (a deleted and an active row share one), so the down migration is
destructive-by-construction and is exercised only in ephemeral Testcontainers databases, never
operationally — 4a's stance. **Consequence for tests (NB2-4):** any subtest that creates an
email-reuse pair (B08, P05) must **delete the deleted row before the harness's terminal `DownTo`**,
or the `DownTo` hits this re-add and fails on the `23505` the paragraph above predicts. B08 lives in
the identity `schema_integration_test.go` (which has a repository and a reversibility `DownTo`) and
cleans up before it; P05 lives in the identity `administration_integration_test.go`, which does
**not** run a 00004 `DownTo` at all.

**Clean bring-up (unchanged):** `make up` → `make seed-identities` → `make seed` → `make smoke` →
`make e2e`.

---

## 7. Rollback

- **Application rollback is the supported path.** Redeploy the previous API and frontend images and
  **leave migration 00004 in place**. The column and partial index are additive; the 4a binary
  reads `users` by explicit column list and never depends on `deleted_at`.
- **A rolled-back 4a binary stays safe against soft-deleted accounts — this is the point of *deleted
  ⊂ disabled* (BLK-3).** A 4b-deleted row has `deleted_at` set **and `disabled_at` set**. The 4a
  binary's `VerifyCredentials` accepts only on `matched && user.Active()`, and `Active()` is
  `DisabledAt == nil` (`user.go:59`), so the deleted user cannot log in; `ResolveIdentitySession`,
  `ListInstructors` and `LockEligibleAuthoringOwner` all filter `disabled_at IS NULL` and refuse it
  too. The `users_deleted_implies_disabled` CHECK (§6) makes this hold for *every* row a rolled-back
  binary could meet, not just those `SoftDeleteUser` wrote — so this argument rests on a constraint,
  not on a statement (NB2-9). **Without** the extra `disabled_at`, a rolled-back 4a binary (whose
  `GetIdentityCredential` has no `deleted_at` filter) would let every deleted user log in again — a
  security regression, not a lost capability. Edition 1 missed this; edition 2 closes it in the
  write and the schema.
- **The one residual data-shape hazard:** a rolled-back 4a `GetIdentityCredential` is `:one` with no
  `deleted_at` filter, so *after a real email reuse* it may `Scan` whichever of the two rows the
  plan yields first. Both are refused anyway (the deleted one by `!Active()`, the active one only if
  it, too, were somehow disabled), so the worst case is that a legitimately active reuser is denied
  login until the 4b binary (with the filter) is restored. Recorded so a reviewer weighs it.
- **Do not `goose down 00004` to recover from an application defect.** It drops `deleted_at` and its
  audit trail, and its re-added table-wide `UNIQUE` fails outright if any email was reused. A
  forward migration fixes a schema mistake; the down migration exists for tests. (It touches only
  `deleted_at` and the email index — `disabled_at` and block-state audit are untouched.)
- **Frontend rollback alone is inert as a security measure** (4a): removing `/admin/users` grants
  nothing; the API refusals stand on their own.

---

## 8. Business error vocabulary (D-5 style)

Server-enforced, mapped once in the handler's `writeError`, contract-proven. New codes in the
OpenAPI `error.code` enum:

| Code | HTTP | Meaning | Where decided |
|---|---|---|---|
| `cannot_target_self` | 409 | an admin tried to block or delete their own account (S6); one rule covering both | domain `CanBlock`/`CanDelete`, pre-lock (self is known from the caller) |
| `admin_immutable` | 409 | `DELETE` targeting an `admin` account (S6); no admin is deletable, role change is out of scope | domain `CanDelete`, over the unlocked role read (role is immutable) |
| `last_admin` | 409 | blocking the last active admin (Q5) — the **race** outcome, never a sequential one | domain `CanBlock`, over the `LockAdmins` set count |
| `user_not_found` | 404 | the target id is unknown or names a soft-deleted account, for status/delete/password | adapter, inside the transaction (missing row) |

Reused unchanged: `authentication_required` (401), `forbidden` (403, the non-admin refusal),
`origin_forbidden` (403), `invalid_request` (400), `dependency_unavailable` (503), `internal_error`
(500). A business refusal is never a 500 (spec criterion 3); an aborted transaction never surfaces
as a business code (BLK-2: `last_admin` is a 409 from a re-read, not a 503 from a `40P01`).

**`email_taken` is deliberately absent (BLK-1).** No 4b write clears `deleted_at` or edits an
email, so nothing can add a second active row for one email; the `23505` on `users_email_active_key`
is unreachable from any 4b endpoint. It is not in the enum, not in the adapter, and has no
acceptance row. (If admin-driven *undelete* is ever wanted, it is a new explicit operation and
reopens this — a Phase-later trigger, not 4b.)

---

## 9. Acceptance scenarios

Per `acceptance-spec`: three groups (all present), one observable result per row, one evidence
level per row, exactly one owning task, every owner able to produce its evidence **when that task
lands**. Mutation rows assume a valid `Origin` unless the row is about `Origin`. Manual rows:
**none**. **T02 owns no row — its evidence is its domain unit tests**; every other task owns at
least one (NB-20).

### Happy path

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| P01 | Directory lists every user with status | admin session; fixtures + disposables seeded | `GET /api/v1/admin/users` | `200`; each row carries id, email, displayName, role and a status of active/blocked/deleted derived from the timestamps | contract `admin_handler_test.go` | T05 |
| P02 | Default filter hides deleted; the deleted tab shows them | one deleted user exists | `GET /api/v1/admin/users`, then `?status=deleted` | the default omits the deleted user; the `deleted` filter returns exactly it | contract | T05 |
| P03 | Block kills auth and live sessions | active user with a live session | `POST /users/{id}/status {"status":"blocked"}` | `200`; a later `GET /auth/me` with that cookie answers `401` and the user's session rows are gone | integration | T04 |
| P04 | Restore re-enables auth | the blocked user from P03 | `POST /users/{id}/status {"status":"active"}` | `200`; the user logs in and authenticates again | integration | T04 |
| P05 | Soft delete retains the row, disables it, and frees the email | active user U, email E | `DELETE /users/{id}`, then insert a new active user with E | `204`; U's row persists with `deleted_at` **and `disabled_at`** set; a new active user with E is accepted | integration | T04 |
| P06 | Password reset: new works, old fails, sessions revoked | user with a known password and a live session | `POST /users/{id}/password {"password": new}` | `204`; login with `new` succeeds, with the old password fails, and the pre-reset session no longer authenticates | integration | T04 |
| P07 | The directory renders per-row actions | admin session, ≥1 user | open `/admin/users` | a table of email, name, role, status, each row offering block/restore (by state), delete and reset, localized | component `UserListView.test.tsx` | T07 |
| P08 | The full admin journey in a browser | Compose up; disposables seeded | sign in admin, block `block-me@`, restore it, reset `reset-me@`, delete `delete-me@` | the directory reflects each change; the deleted user appears only under the `deleted` filter | e2e `admin-users.spec.ts` | T08 |

### Boundaries

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| B01 | Empty directory (API) | a filter matching nobody | `GET /api/v1/admin/users?status=deleted` on a db with no deletes | `200 {items: [], total: 0}` | contract | T05 |
| B02 | Empty directory (screen) | the API returns no rows | open `/admin/users` on that filter | the localized empty state renders | component | T07 |
| B03 | Blocking a non-last admin is allowed | two active admins A and B, A acting | `POST /users/{B}/status {"status":"blocked"}` | `200`; B blocked, A active | integration | T04 |
| B04 | Password bounds | admin session | reset with an empty password, one over 1024 bytes, then exactly 1024 bytes | `400`, `400`, then `204`; 1–1024 bytes of any content is accepted, incl. whitespace-only and same-as-current (accepted by design, NB-8) | contract | T05 |
| B05 | Reset produces a valid Argon2id credential | admin session | reset then read the stored hash | it matches `$argon2id$v=19$m=65536,t=3,p=2$…` and verifies against the new password | integration | T04 |
| B06 | Migration over a populated database | 5 demo courses + 3 fixtures present | apply migration 00004 | `deleted_at` exists and is null on every row, `users_email_active_key` exists, `users_email_key` is gone, no row lost | integration `repository_integration_test.go` | T01 |
| B07 | The failing-migration canary still rolls back | migration 00004 committed | run the course integration suite | the subtest passes with the real `00004` applied and the canary renumbered to `00005` | `make integration` | T01 |
| B08 | Login survives email reuse | a soft-deleted user and an active user share one email | `POST /auth/login` as the active user | `200`; `GetIdentityCredential` returns the active row (its `deleted_at IS NULL` filter decides); the subtest deletes the deleted row before any `DownTo` | integration `schema_integration_test.go` (NB2-4) | T01 |
| B09 | Contract and client stay in step | all OpenAPI + query changes committed | `make check-generated` | sqlc and Orval output matches committed; exit 0 | `make check-generated` | T06 |
| B10 | The ordered gate passes | clean checkout | `make check` | every step exits 0 with the new surface | `make check` | T09 |
| B11 | Same-state block is idempotent | an already-blocked user | `POST /users/{id}/status {"status":"blocked"}` again | `200`; `disabled_at` and `updated_at` are unchanged and no session write occurs | integration | T04 |
| B12 | A block writes updated_at so the CHECK holds | a user seeded with the DB `now()` as `created_at`, blocked with a **real clock ≥ that `now()`** (not a fixed fake earlier than the DB, NB2-10) | `POST /users/{id}/status {"status":"blocked"}` | `200`; the row satisfies `disabled_at BETWEEN created_at AND updated_at` (no `23514`) | integration | T04 |

### Refusals

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| R01 | Anonymous refused everywhere | no cookie, valid `Origin` | list, status, delete, password | each `401 authentication_required`; no repository method called | contract | T05 |
| R02 | A student refused everywhere | student session | the same four | each `403 forbidden` | contract | T05 |
| R03 | An instructor refused everywhere | instructor session | the same four | each `403 forbidden` (4b is admin-only) | contract | T05 |
| R04 | A disabled admin refused | admin fixture with `disabled_at` set | list | `401 authentication_required` | contract | T05 |
| R05 | Cross-site mutation refused first | valid admin cookie, foreign `Origin` | status, delete, password | each `403 origin_forbidden`, no state change | contract | T05 |
| R06 | Admin cannot block themselves | admin session | `POST /users/{self}/status {"status":"blocked"}` | `409 cannot_target_self`; still active | contract | T05 |
| R07 | Admin cannot delete themselves | admin session | `DELETE /users/{self}` | `409 cannot_target_self` (the same self-target rule, refused before admin-immutable) | contract | T05 |
| R08 | Admin cannot delete an admin | admin A, admin B | `DELETE /users/{B}` | `409 admin_immutable`; B still reads back | contract | T05 |
| R09 | The last active admin cannot be blocked | a DB state with exactly one active admin (constructed) | block that admin through the adapter | `last_admin`; the admin stays active — exercised at the adapter, since a real API caller is always an extra active admin (NB-1) | integration | T04 |
| R10 | The last-admin guard holds under the race, with no 503 | two active admins, two concurrent blocks | issue both blocks concurrently | at most one succeeds; ≥1 active admin always remains; the loser answers `409 last_admin`, **never `503`** | integration | T04 |
| R11 | Unknown id is not found | admin session, unused uuid | status, delete, password | each `404 user_not_found` | contract | T05 |
| R12 | A deleted target is not operable | a soft-deleted user | status, delete, password | each `404 user_not_found`; the row still appears under the `deleted` filter | integration | T04 |
| R13 | Malformed id rejected before lookup | admin session | `DELETE /users/not-a-uuid` | `400 invalid_request`, no repository call | contract | T05 |
| R14 | Illegal status value refused | admin session | `POST /users/{id}/status {"status":"deleted"}` | `400 invalid_request`; deleted reachable only via `DELETE` | contract | T05 |
| R15 | Dependency failure degrades safely | the repository port returns dependency failure | `ListUsers`, `SetStatus` | `503 dependency_unavailable`; no SQL/credential/internal detail leaked | application fake | T03 |
| R16 | A hasher failure is not a business refusal | a fake verifier whose `Hash` errors | `ResetPassword` through the adapter | `dependency_unavailable`, distinct from a business refusal — no partial write | integration (fake verifier, no DB write) | T04 |
| R17 | Cancellation is not a fault | context cancelled mid-call | `SetStatus` | cancellation reported, distinct from dependency failure and refusal | application fake | T03 |
| R18 | A cancelled block leaves nothing behind | block cancelled mid-transaction | `SetStatus` then re-read | the user is still active and every session survives | integration | T04 |
| R19 | Business codes reach the domain intact | MSW answers `409 last_admin`, `409 cannot_target_self`, `403 forbidden`, `500` | call the matching use cases | the first three surface as named outcomes; the last as dependency-unavailable | adapter MSW | T06 |
| R20 | The client guard is not the mechanism | student signed in, in the browser | call `/api/v1/admin/users` via `page.request` | `403 forbidden`, though the SPA never renders the screen | e2e | T08 |
| R21 | Unusable search params disappear | admin session | open `/admin/users?status=nonsense&page=abc` | the directory renders with default filters, no error boundary, the URL keeps what was typed | route test `admin-search.test.ts` | T07 |
| R22 | Login racing a block **or delete** leaves no live session; a second concurrent delete is 404 | a login in flight while an admin blocks (and, separately, deletes) the same user; and two concurrent deletes | run each pair concurrently | for block and for delete: login is either refused (`invalid_credentials`, no session) or its session is revoked; no live session survives (Q1); of two concurrent deletes the second answers `404 user_not_found` (NB2-12) | integration | T04 |
| R23 | A blocked/deleted owner's course cannot be PUT | an instructor is blocked (or deleted); they own a course | `PUT` that course as admin | `400 invalid_owner` until an admin reassigns the owner — the accepted frozen-until-reassign stance (§5.h) | integration `authoring_integration_test.go` | T04 |
| R24 | An instructor sees not-found at `/admin/users` | instructor session | open `/admin/users` | the route's `beforeLoad` throws `notFound()`; the not-found screen renders, not an error boundary | route/component test | T07 |

Group counts: 8 happy, 12 boundary, 24 refusal = **44 rows**. Manual: 0. Every task except T02
owns ≥1 row; no row has two owners.

---

## 10. Task cut

Cut along **executable evidence and environment boundaries** (the I06/4a lesson). Each task is
0.5–2h, regenerates its own sqlc/Orval paths (4a NB-9), and names its dependencies.

| ID | Task | Est. | Depends on | Environment / evidence | Owns |
|---|---|---:|---|---|---|
| T01 | Migration 00004 (`deleted_at`, the two CHECKs incl. `deleted_implies_disabled`, drop `users_email_key`, partial index, down); `GetIdentityCredential` gains `deleted_at IS NULL`; the course canary fix (real `00004` in the MapFS, renumber to `00005`) + **B06 in the course harness**, **B08 in the identity `schema_integration_test.go`** (email-reuse pair cleaned before its `DownTo`); `seed-identities.sql` resets `deleted_at` and adds the three disposables | 2h | GATE 2 green | Testcontainers PostgreSQL (`make integration`) | B06, B07, B08 |
| T02 | Domain: `DeletedAt` on `user.go` + `NewUser` bound; `AccountStatus` projection; `CanBlock`/`CanDelete` (self-target, admin-immutable, last-admin over a supplied admin set); password bounds | 1h | — | Go unit tests (`make test`) | — (evidence: its unit tests) |
| T03 | `Administration` application use cases over fakes (`ListUsers`, `SetStatus`, `Delete`, `ResetPassword`), deadlines, cancellation, dependency failure | 1.25h | T02 | Go unit tests with fakes (`make test`) | R15, R17 |
| T04 | PostgreSQL administration adapter + sqlc queries (READ COMMITTED / `pool.Begin`): escaped list; `LockAdmins`/`LockUserForStatus` with the last-admin count + session revoke + updated_at; the single nullable-parameter `SetUserDisabledAt` (block+restore); soft-delete (deleted+disabled) + revoke; set-password (consumer-side `Hash`, UPDATE-then-revoke) + revoke; the AdminUser mapper; the login-vs-block/delete race; hasher-failure via a fake verifier; the cross-module blocked-owner assertion | 2h | T01, T03 | Testcontainers PostgreSQL (`make integration`) | P03–P06, B03, B05, B11, B12, R09, R10, R12, R16, R18, R22, R23 |
| T05 | HTTP surface: the four routes reusing `Authenticate`/`RequireSameOrigin`, the admin-role gate, DTOs and bounds, error mapping, OpenAPI paths/schemas/codes and generated output | 2h | T04 | Go contract tests against `openapi.yaml` (`make test`) | P01, P02, B01, B04, R01–R08, R11, R13, R14 |
| T06 | Frontend admin-users adapter and application: plain-string ids, zod schemas, repository with business-code mapping, use cases, `admin.ts` composition extension, `admin-user-queries.ts` importing `adminListStaleMs`, `invalidateAfterUserWrite`; regenerates Orval | 1.5h | T05 | Rstest with MSW (`make test`), `make check-generated` | B09, R19 |
| T07 | The `/admin/users` route and views: `UserListView`, status tabs, extended never-throwing search, per-row actions, `ResetPasswordDialog` (self-reset sign-out), empty/error/degraded states, the `canAdministerAccounts` guard + `notFound` for instructors, the header entry point, uk/en, a11y | 2h | T06 | Rstest route/component tests incl. axe | P07, B02, R21, R24 |
| T08 | The browser journey `admin-users.spec.ts` on the disposables using the shared `storageState`; the direct-API non-admin refusal; `route-guards.spec.ts` extended | 1h | T07 | Playwright against Compose (`make e2e`) | P08, R20 |
| T09 | Ordered gate and report: full `make check` from a clean checkout; the 4b report section with measured results, limitations and next task | 1h | T01–T08 | `make check` | B10 |

**Total: 13.75 h.** Sequencing notes:

- **T01 first** — the migration makes soft delete and email reuse possible; every later query and
  test runs against the migrated schema, and the canary/seed fixes land with it so trunk is never
  red at `make integration` between tasks.
- **T02 before T03/T04** — the invariants are pure domain, provable with unit tests; the
  application (T03, fakes) and the adapter (T04, real Postgres) both consume the policy. The
  *observable* refusals belong where they are observable: the last-admin and login-vs-block races
  (R09/R10/R22) to T04 (they need a real DB), the self/admin-immutable/not-found refusals (R06–R08,
  R11) to T05 (contract).
- **T03/T04 split by evidence environment**, not layer: T03 is everything provable with fakes in
  `make test`; T04 is everything that needs real PostgreSQL. `R16` (hasher failure) sits in T04
  because the consumer-side `Hash` lives in the adapter — a fake verifier proves it with no DB
  write. The frontend `T06`/`T07` split the same way (MSW vs a rendered route).
- **T05 before T06/T07** — the OpenAPI document must exist before Orval regenerates the client.
- **T08 last of the feature tasks** — it reuses the shared `storageState`, adds zero logins, and
  acts only on the disposables, so it cannot revoke a shared session mid-run.

---

## 11. Riskiest decisions — pointer list

Not a restatement; each points at the section that reasons it out (NB-15).

1. Soft delete writes **deleted *and* disabled** in one statement (§4 D-2) — the load-bearing
   choice; it closes rollback (§7) and the two omitted filters (§5.a) at once.
2. Email-uniqueness migration: drop `users_email_key`, add a partial unique index (§6); down is
   test-only.
3. Only `GetIdentityCredential` gains a `deleted_at` filter; the other four status-gating queries
   are correct by construction (§5.a).
4. Real `00004` collides with the course canary → renumber to `00005` (§5.b).
5. Last-admin is a **row-set lock in id order**, counted in the domain; a 409 under the race, never
   a 503 (§4 D-2, R10) — edition 1's `count(*) FOR UPDATE` was invalid and deadlocked.
6. Block/delete lock the users row before revoking sessions, serialising against login `Rotate`
   (§4 D-2, R22).
7. Password hashing is **consumer-side** (no `ports.PasswordHasher`), on the renamed `Passwords`
   interface, matching the module's own idiom (§3, NB-19/NB2-3).
8. `email_taken` is removed — unreachable, because restore never clears `deleted_at` (§8, BLK-1).
9. Blocking/deleting an instructor **freezes their courses** until reassign/restore — accepted,
   4a's lock unchanged (§5.h, R23).
10. e2e acts only on three disposable fixtures; `make e2e` alone is not re-runnable without
    `seed-identities` (§5.f).

---

## 12. Disposition of the non-blocking findings (22 round 1 + 12 round 2)

The repository tradition: addressed or explicitly retained, never silent.

### Round 1 (edition 1 → edition 2)

| # | Disposition | Where |
|---|---|---|
| NB-1 sequential `last_admin` unreachable; R09 as written is `cannot_target_self` | **Addressed** — D-3 cell says a sequential call never hits the last admin; R09 is now an adapter-level row with a constructed single-admin state | §4 D-3, R09 |
| NB-2 `AccountStatus` over two timestamps but `domain.User` has one | **Addressed** — `user.go (extend)` adds `DeletedAt` + bound; the `AdminUser` mapper is in T04 | §2, T02, T04 |
| NB-3 B06 has no owning harness | **Addressed** — B06 lands in the course `repository_integration_test.go` (migrates over 5 courses + 3 fixtures); B08 lands in the identity `schema_integration_test.go`, which has a repository (revised by NB2-4) | §2, B06, B08 |
| NB-4 the `/admin/users` guard admits instructors | **Addressed** — `identity.ts (extend)` adds `canAdministerAccounts`; the route `beforeLoad` throws `notFound()`; the header gains an admin-only entry point; R24 is the row | §2, R24 |
| NB-5 `:one` failure mode misstated | **Addressed** — "silently returns whichever row comes first", not "errors" | §5.a, §7 |
| NB-6 same-state block clobbers the audit timestamp | **Addressed** — status is idempotent: same-state → 200, no write, no revoke, no count; B11 asserts the timestamp | §4 D-2, B11 |
| NB-7 an admin resetting their own password logs themselves out | **Addressed, stated no dedicated row** (NB2-1) — Q2 revokes all sessions; the dialog announces the sign-out and lets the redirect run; stated in §4 D-2 (reset paragraph) and §2 `ResetPasswordDialog`, exercised in P08's reset step | §2, §4 D-2 |
| NB-8 password bounds under-specified | **Addressed** — 1–1024 bytes, any content; exact-1024 accepted; whitespace-only/same-as-current accepted by design; B04 covers all three | B04 |
| NB-9 seed vs partial index on a persistent volume | **Addressed** — both facts and the manual recovery stated | §5.e |
| NB-10 block must write `updated_at` for the CHECK; clock skew | **Addressed** — every 4b write sets `updated_at` to the same `$clock`; "API clock ≥ DB clock" recorded; B12 is the row | §4 D-2, B12 |
| NB-11 paging/search unspecified; default status | **Addressed** — default hides deleted (§1, P02 the filter row); page-past-end → `200 {items: [], total}` and `ListAdminUsers` searches email + display_name with a **duplicated** one-line replacer (stated in §2, NB2-3; no dedicated row per NB2-1) | §1, §2, P02 |
| NB-12 deleted/blocked instructor's courses stay published | **Addressed** — one sentence: courses keep their state and owner; reassignment is manual | §5.g |
| NB-13 `admin-users.spec.ts` rerunnable only after re-seed | **Addressed** — stated; the spec asserts the disposables' pre-state in `beforeAll` | §5.f |
| NB-14 a fabricated briefing quote | **Addressed** — the quote and the "contrary to the briefing" framing are gone; one factual sentence in §3; the §5.c/§11 duplicates removed | §3 |
| NB-15 §11 restates the plan; §11.9 asserts a false "only race" | **Addressed** — §11 is a risk→section pointer list; the false claim is gone; scope tightened to "exactly S1–S8" | §11, Scope |
| NB-16 a third staleness constant | **Addressed** — the 4a trigger has fired; `admin-user-queries.ts` imports `adminListStaleMs` | §2, §3 |
| NB-17 branded `UserId` for one id kind | **Addressed** — plain string ids; the branding trigger (a second identity id kind sharing call sites) named | §2 |
| NB-18 twin composition/search files; unlisted `__root.tsx` | **Addressed** — `admin.ts` composition and `admin-search.ts` are extended, so no new root-context key; no separate root | §2 |
| NB-19 `ports.PasswordHasher` vs the module's consumer-side idiom | **Addressed** — no port; the consumer-side interface is widened with `Hash` and renamed `PasswordVerifier` → `Passwords` (NB2-3); R16 moves to T04 | §2, §3, R16 |
| NB-20 T02 owns no row; "Actor reused unchanged" is decorative | **Addressed** — "every task except T02 owns a row; T02's evidence is its unit tests"; the Actor sentence is corrected in D-1 | §9, §4 D-1 |
| NB-21 self-delete rule redundant and misnamed | **Addressed** — one `cannot_target_self` rule covers block-self and delete-self; no separate self-delete clause; the code is renamed | §4 D-2, §8, R07 |
| NB-22 canary citation points at the wrong document | **Addressed** — cited as `plan.md` BLK-12 / commit `fd90fc9` | §5.b |

### Round 2 (edition 2 → edition 2.1) — all 12 addressed

| # | Disposition | Where |
|---|---|---|
| NB2-1 §12 over-claims rows for NB-7/NB-11 | **Addressed** — those two cells downgraded to "stated, no dedicated row"; the self-reset is exercised inside P08, the paging/escaping stated in §1/§2 | §12 NB-7, NB-11 |
| NB2-2 "closes three blocking" restated 5× | **Addressed** — kept once in the preamble and D-2; the §5.a/§7/§11 mentions reduced to a pointer; §13 (gate note) deleted now the gate is green | preamble, §4 D-2 |
| NB2-3 a `Verifier` that hashes; "the course replacer" implies a cross-module import | **Addressed** — the interface is renamed `Passwords`; the search escaping is a **duplicated** one-line `NewReplacer + ESCAPE '!'`, copied not imported | §2, §3 |
| NB2-4 email-reuse fixtures sit before a `DownTo` that fails on a reused email; B08 needs a repository | **Addressed** — B08 moved to the identity `schema_integration_test.go`; every reuse-pair subtest deletes the deleted row before its `DownTo`; the identity administration harness runs no 00004 `DownTo` | §2, §5.b/§6 Down, B08, T01 |
| NB2-5 R10's "never 503" silently depends on READ COMMITTED | **Addressed** — D-2 states all 4b writes use `pool.Begin` (READ COMMITTED) and warns against copying the course module's `RepeatableRead` (→ 40001/503) | §4 D-2 |
| NB2-6 reset lock/write order and the old-password TOCTOU unstated | **Addressed** — reset is UPDATE-then-revoke; the one-round-trip old-password TOCTOU recorded as an accepted limitation with its revisit trigger | §4 D-2 |
| NB2-7 restore of a blocked admin: lock choice ambiguous | **Addressed** — restore always uses `LockUserForStatus`; `LockAdmins` only for a block whose target is an admin | §4 D-2 |
| NB2-8 two status-write queries where one nullable-parameter UPDATE suffices | **Addressed** — collapsed to one `SetUserDisabledAt` with a nullable `disabled_at`; `RestoreUser` dropped (`net: -1 query`) | §2, T04 |
| NB2-9 the invariant rests on one COALESCE; a CHECK makes it structural | **Addressed** — `CHECK (deleted_at IS NULL OR disabled_at IS NOT NULL)` added in §6 step 3; `Active()` = `DisabledAt == nil && DeletedAt == nil`; §5.a/§7 now cite the constraint | §6, §4 D-2, §2 |
| NB2-10 B12's fixed clock fails under a clock earlier than the DB's | **Addressed** — B12 uses a real clock ≥ the DB `now()`; the "API clock ≥ DB clock" limitation extended to `deleted_at` | B12, §4 D-2 |
| NB2-11 reset on a blocked target allowed but unstated | **Addressed** — stated: reset on a blocked user → `204`, the block stays; only a deleted user → `404` | §4 D-2 (reset paragraph), §4 D-3 |
| NB2-12 login-vs-delete and delete-vs-delete unasserted | **Addressed** — R22 extended to block **and** delete, plus the double-delete `404` | §4 D-2, R22 |

**Retained: none.** All 22 round-1 and 12 round-2 findings are addressed. The complexity lens ran in
full on edition 2 (GREEN); edition 2.1 folds its two consolidations (NB2-8, NB2-9).
