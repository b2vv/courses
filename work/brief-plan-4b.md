# Planning brief — Phase 4b plan.md (docs/admin/plan-4b.md)

You write `docs/admin/plan-4b.md` for slice 4b (identity administration). No code.
Ponytail-full: simplest thing that works; reuse the 4a seam and idioms; name every
deviation from the stack list with its trigger.

## Read first, in this order
1. `AGENTS.md`; `.claude/standards.md` (gate order, placement, reuse index, branches).
2. `docs/admin/spec-4b.md` — the approved spec WITH the GATE 1 decisions table. Binding.
3. `docs/admin/plan.md` (4a) — the shape/quality bar and the authorization seam you REUSE:
   §4 D-1 (httpserver.Actor, Authenticate, RequireSameOrigin), D-5 (error vocabulary
   style). 4b's endpoints ride the SAME seam.
4. `docs/admin/report.md` (4a) — what shipped; the lockfile and storageState traps.
5. `work/CTO-RESEARCH.md` — briefing (fact #1 now: server authz EXISTS).
6. `work/PRODUCT-BRIEF.md` §24.8/§24.10 — the 4b shape and additive-capability model.

## Verified facts you must not rediscover wrong (checked today)
- `backend/migrations/00002_identity.sql`: users has `disabled_at timestamptz` (nullable),
  `email text NOT NULL UNIQUE`, `role text CHECK IN (student,instructor,admin)`,
  `password_hash` Argon2id. 4b needs: a `deleted_at` column for soft delete, and the email
  UNIQUE must become "unique among non-deleted" — decide the migration (partial unique
  index `WHERE deleted_at IS NULL` vs a different shape) and state its cost.
- Authentication already refuses a disabled user and checks the session row
  (`internal/identity/domain`, `application/identity.go` Authenticate). Blocking =
  set disabled_at + delete the user's sessions in one tx (Q1). Existing session deletion
  query: `DeleteIdentitySession` (per-digest) — you likely need a delete-all-for-user
  query; name it.
- The admin authorization seam exists: `Role.CanAdministerAccounts()` (admin-only),
  `Authenticate`/`RequireSameOrigin` exported, `Actor` in platform/httpserver. 4b's
  handlers reuse them — do NOT build a second seam. The instructor-list endpoint
  (`GET /api/v1/admin/instructors`) already lives in the identity http handler as the
  first admin-account read; 4b extends this admin surface.
- Argon2id hashing lives in `internal/identity/adapters/argon2id`; the application creates
  credentials — password reset reuses that path, not a new hasher.
- sqlc identity queries: `backend/sql/queries/identity.sql`; generated under the identity
  postgres adapter. OpenAPI: `backend/openapi/openapi.yaml`; Orval regenerates; drift is a
  red gate. Frontend admin idioms from 4a: `frontend/src/app/composition/admin*.ts`,
  `presentation/admin/`, `routes/admin*.tsx`, `adapters/api/admin-course-*`.
- Last-admin invariant (Q5) and self-block refusal (S6) and admin-immutability (S6) are
  DOMAIN rules — put them in the identity domain policy with unit tests, enforced before/
  after the repository as data requires (mirror 4a's D-4a placement reasoning).
- Migration numbering: next is 00004 (00003 is authoring). Check for the synthetic
  failing-migration test canary numbering in identity/course integration tests before
  picking a number for any test fixture.

## The plan must contain (GATE 2 checks exactly this)
- Placement per manifest; every new file named; 4b lives in the identity module
  (domain/application/ports/adapters) + a frontend admin users slice.
- Reuse-first: reused (the 4a seam, Argon2id, the kit, admin composition idioms) vs new,
  and why existing did not fit where you add.
- Side effects · Migration plan · Rollback — three subsections. The email-uniqueness
  migration is the sharp one: state the exact index/constraint change and how existing
  rows and the seed survive it.
- Business error vocabulary for 4b (self-block, admin-immutable, last-admin, user-not-
  found, email-taken-among-active) — explicit codes, D-5 style.
- Acceptance table via acceptance-spec (read /Users/strelia/.claude/skills/acceptance-spec/SKILL.md):
  each row owned by exactly one task; cover Q1 session-death, Q5 last-admin, soft-delete
  email reuse, reset revoking sessions, the full non-admin refusal matrix.
- Task cut: 0.5–2h, cut around executable evidence and environment boundaries (the I06/4a
  lesson), dependencies explicit, each owning acceptance rows.
- Server enforcement design reusing the 4a seam; the refusal matrix; the transactional
  invariants (last-admin, self-block) with where they are checked.
- Scope fence: role change is NOT in 4b; enrollment/per-course blocking NOT in 4b.

Write in English, matching the repo plan style. Touch no other file, do not commit, do
not run Docker targets. End with a 10-line summary of the riskiest decisions.
