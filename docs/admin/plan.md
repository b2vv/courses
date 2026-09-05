# Admin course management — Phase 4a plan

Edition 2, 2026-09-05. Supersedes edition 1, which GATE 2 returned RED (12 blocking, 18
non-blocking — `docs/admin/gates.md`) and whose plan-defense left row 3 (race and replay)
open. Every blocking finding is closed at a named location below; §10 records the
disposition of all 18 non-blocking findings, addressed or explicitly retained.

Sources: `docs/admin/spec.md` (GATE 1 decisions, binding), `work/CTO-RESEARCH.md`
(base `c8d2324`), `work/PRODUCT-BRIEF.md` §21/§24,
`docs/reference/frontend-patterns.md`, `docs/identity/report.md` §I06 п.8.

Each task is 0.5–2 hours, owns named acceptance rows, and can produce its own verification
evidence **at landing time** — the property edition 1 asserted without checking (BLK-4,
BLK-11, BLK-12).

Scope fence: **slice 4a — course management only**. Slice 4b (identity administration:
blocking, restore, soft delete, admin password reset) is out of scope and gets its own spec
and plan. The single identity-side addition here is the read-only owner picker
(`GET /api/v1/admin/instructors`, D-4c); it manages nobody.

Mode: ponytail-full. Every deviation from the product brief's stack list is named with the
trigger that would revisit it (§5.7, §5.8).

---

## 1. Shape of the slice

Seven operations across four new paths, one extended public path, one migration, one new
authorization seam, two admin screens.

The central decision is that **a course is one aggregate written whole**. The editor loads
`{course, modules, lessons}` and saves the same shape back with `PUT`. That removes eight
endpoints (module and lesson create/update/delete, two reorder calls), removes the
delta-vs-full-write protocol question entirely (`frontend-patterns` §6), makes ordering an
array index rather than a column the client computes, and makes invalidation a single query
key. Its costs — a lost-update window, a payload budget, and the identity of the nodes
inside the payload — are paid for explicitly in D-3.

| Method | Path | Who | Answer |
|---|---|---|---|
| GET | `/api/v1/admin/courses` | admin: all; instructor: own | page of admin rows (id, slug, title, status, instructorId, updatedAt) |
| POST | `/api/v1/admin/courses` | admin, instructor | `201` created `draft`; server mints the course id |
| GET | `/api/v1/admin/courses/{id}` | admin; owning instructor | the whole tree: course fields, modules, lessons |
| PUT | `/api/v1/admin/courses/{id}` | admin; owning instructor | `200` with the stored tree; body carries `expectedUpdatedAt` |
| POST | `/api/v1/admin/courses/{id}/status` | admin; owning instructor | `200` after a legal transition; body carries `expectedUpdatedAt` |
| DELETE | `/api/v1/admin/courses/{id}` | admin; owning instructor | `204`; draft only, cascades |
| GET | `/api/v1/admin/instructors` | admin only | `[{id, displayName, role}]`, at most 200, unpaginated |

Four new path entries, seven operations, plus one extended path (NB-15).

Public delta: `GET /api/v1/courses/{slug}` gains `program` — module titles with their lesson
titles, in order, text only (spec Q6). The list response is unchanged.

Status transitions are exactly Q3: `draft→published`, `published→archived`,
`archived→published`. Everything else is `409 invalid_status_transition`. One endpoint
carries all three, because the refusal is one matrix in one place.

---

## 2. Placement (Розміщення)

Per `.claude/standards.md` §Розміщення. Every new file is named; extended files are marked
`(extend)` with what changes.

### Backend

| File | Contents |
|---|---|
| `backend/migrations/00003_authoring.sql` | `courses.instructor_id`, `course_modules`, `course_lessons`, indexes; down migration |
| `backend/internal/course/domain/authoring.go` | `Actor`, `ActorRole`, `Module`, `Lesson`, `CourseTree`; ownership/role policy; transition matrix; tree invariants and bounds; slug-mutability rule |
| `backend/internal/course/domain/authoring_test.go` | table tests for the above |
| `backend/internal/course/application/authoring.go` | `Authoring` use cases: `List`, `Load`, `Create`, `Save`, `Transition`, `Delete`; bounded deadlines; cancellation |
| `backend/internal/course/application/authoring_test.go` | fake-port table tests, dependency failure, cancellation |
| `backend/internal/course/ports/repository.go` (extend) | `AuthoringRepository` beside the existing read-only `CourseRepository` |
| `backend/internal/course/adapters/postgres/authoring.go` | tree read, transactional tree write, guarded transition, guarded delete, owner-scoped list, `23505` → `slug_taken` |
| `backend/internal/course/adapters/postgres/authoring_integration_test.go` | Testcontainers: tree upsert and delete-missing, ordering, cross-module moves, cascade, slug uniqueness under concurrency, optimistic conflict, delete/publish race, foreign node ids, timestamp round-trip, cancellation |
| `backend/internal/course/adapters/postgres/repository_integration_test.go` (extend) | **BLK-12**: apply `seed-identities.sql` before `seed.sql`; supply `instructor_id` in the two direct `INSERT`s (lines 158, 183); renumber the synthetic failing migration from `00003_failure.sql` to `00004_failure.sql` and add the real `00003_authoring.sql` to that subtest's `fstest.MapFS` |
| `backend/internal/course/adapters/http/admin_handler.go` | admin DTOs, decode and size limits, actor mapping, error mapping, `RegisterAdmin` |
| `backend/internal/course/adapters/http/admin_handler_test.go` | contract tests validated against `openapi.yaml`; the refusal matrix |
| `backend/platform/httpserver/actor.go` | `Actor{UserID, Role}`, `WithActor`, `ActorFromContext` |
| `backend/platform/httpserver/actor_test.go` | absent, present and wrong-type context reads |
| `backend/internal/identity/adapters/http/handler.go` (extend) | exported `Authenticate` and `RequireSameOrigin`; `GET /api/v1/admin/instructors` |
| `backend/internal/identity/domain/user.go` (extend) | `Role.CanAdministerAccounts()` (admin only) |
| `backend/sql/queries/courses.sql` (extend) | authoring queries; regenerates `adapters/postgres/generated` |
| `backend/sql/queries/identity.sql` (extend) | `ListInstructors` |
| `backend/openapi/openapi.yaml` (extend) | 4 new paths / 7 operations, extended `GET /courses/{slug}`, admin schemas, eight new error codes |
| `backend/sql/seed.sql` (extend) | `instructor_id` on every demo course; `LOCK TABLE courses IN SHARE ROW EXCLUSIVE MODE`; `ON CONFLICT (id) DO UPDATE` (§5.2) |
| `backend/cmd/api/main.go` (extend) | wires the two middlewares into `RegisterAdmin` |

No new Go module, no new sqlc target, no new platform concern beyond `actor.go`.

### Frontend

| File | Contents |
|---|---|
| `frontend/src/domain/course/course.ts` (extend) | branded `CourseId`/`ModuleId`/`LessonId`, `AdminCourse`, `CourseTree`, `AuthoringError`, transition, slug-mutability and tree validators |
| `frontend/src/ports/AdminCourseRepository.ts` | the authoring port |
| `frontend/src/application/admin/authoring.ts` | use cases over the port, returning `Result` |
| `frontend/src/adapters/api/admin-course-schema.ts` | zod schemas that brand ids while parsing |
| `frontend/src/adapters/api/admin-course-repository.ts` | Orval calls through `createApiTransport`; business-code mapping |
| `frontend/src/app/composition/admin.ts` | request-scoped dependency wiring |
| `frontend/src/app/composition/admin-queries.ts` | query options, mutations, the two staleness constants (NB-1), named invalidation helpers |
| `frontend/src/app/router/admin-search.ts` | never-throwing per-field search parsing for the list |
| `frontend/src/presentation/admin/CourseListView.tsx` | table, status tabs, pagination, empty and error states |
| `frontend/src/presentation/admin/CourseEditorView.tsx` | RHF tree editor, owner field, status controls, delete, conflict state |
| `frontend/src/routes/admin.tsx` (extend) | becomes the guarded layout rendering `<Outlet/>` |
| `frontend/src/routes/admin.index.tsx` | the course list route |
| `frontend/src/routes/admin.courses.$courseId.tsx` | the editor route |
| `frontend/src/app/providers/locales/uk.ts`, `en.ts` | the resource bundles, moved out of `i18n.tsx` |
| `frontend/src/app/providers/i18n.tsx` (extend) | keeps only `createI18n` and `I18nProvider` |
| `frontend/src/presentation/ui/data.tsx`, `states.tsx` (extend) | label props replacing hardcoded Ukrainian defaults (§5.5) |
| `frontend/e2e/auth.setup.ts` | one sign-in per role per run, writing `playwright/.auth/<role>.json` (D-7) |
| `frontend/playwright.config.ts` (extend) | a `setup` project the two browser projects depend on |
| `frontend/e2e/admin.spec.ts` | the browser journey against Compose |
| `frontend/e2e/route-guards.spec.ts` (extend) | adopts the shared storage state except where it must exercise the form |

Dropped from edition 1: `staleness.ts` (NB-1), `CourseFormFields.tsx` (NB-3),
`AdminStates.tsx` (NB-4). Sibling `*.test.ts(x)` files accompany every new frontend module
and are not listed individually.

### Infrastructure and docs

`Makefile`, `.github/workflows/ci.yml`, `README.md`, `.claude/standards.md` §Машинні гейти
(gate reorder, §5.1), `docs/admin/report.md` (new), `work/tasks.md` (orchestrator-owned).

Nginx, Compose, Prometheus and `.env.example` are unchanged (§5.10), and the D-3d budget is
what keeps that true.

---

## 3. Reuse-first

| Concern | Reused | Why nothing new |
|---|---|---|
| HTTP call to our own API | `adapters/api/transport.ts` | `TransportOutcome` already carries the four pre-domain failure codes; the admin adapter widens them into `AuthoringError` exactly as `session-repository.ts` does |
| SSR cookie forwarding, per-request headers | `adapters/api/request-headers.ts` | admin routes are `ssr: false`, so this serves only the public program read |
| Route-loader cancellation | `app/composition/catalog-queries.ts` `ensureRouteQuery` | the list route is a loader consumer like the catalog; the helper is already generic |
| Route protection, session-loss redirect | `app/composition/require-identity.ts` | `/admin` keeps `requireIdentity` and `useRedirectWhenSessionEnds`; the layout restructure does not touch the guard |
| Table, Pagination, Tabs, TextField, TextArea, Select, Dialog, EmptyState, ErrorState, Skeleton | `presentation/ui/` | a ≤50-row page needs no sorting engine, virtualization or row selection; the status filter is a `Tabs` bound to a search param; the kit's prop-driven states are the admin screens' states (NB-4) |
| HTTP envelope, request id, JSON and error writers | `platform/httpserver` | `WriteJSON`/`WriteError` already emit the contract's error envelope |
| Authentication, session cookie, disabled-account handling | identity HTTP middleware | it exists and is proven; only its **visibility** changes |
| Same-origin (CSRF) refusal | the same handler's `sameOrigin` check | admin mutations are cookie-authenticated state changes and need the identical rule; a second implementation would drift |
| Contract-test harness (kin-openapi, `openapi3filter`) | `internal/*/adapters/http/handler_test.go` | it validates responses against the published document; new paths plug in |
| Repository harness (Testcontainers, goose provider) | `course/adapters/postgres/repository_integration_test.go`, `identity/.../schema_integration_test.go` | already boots isolated PostgreSQL and applies migrations |
| Seed idiom (one transaction, `LOCK TABLE`, reconcile-to-file) | `backend/sql/seed-identities.sql` | `seed.sql` adopts it verbatim (§5.2) |
| MSW, Rstest, Playwright | existing configs | no test-runner change |

**What is new, and why the existing thing did not fit:**

- `ports.AuthoringRepository` — `CourseRepository` is a two-method read port over
  *published* courses. Authoring reads drafts and archives, writes a tree inside one
  transaction, and scopes by owner. Widening the read port would push write concerns into
  the catalog; it stays untouched.
- `platform/httpserver.Actor` — nothing today carries an authenticated caller across a
  module boundary, because nothing today has two modules in one request path (D-1).
- `domain/authoring.go` — the existing `Course` invariants describe a published catalog
  row; ownership, transitions, slug mutability and tree structure are new rules in the same
  package.
- React Hook Form — the brief's §24.6 trigger fires here: the first admin forms.

---

## 4. Decisions

### D-1 The authorization seam (server side)

Authentication is reused; **authorization is new**, per briefing architectural fact #1. The
first protected business endpoint has nothing to inherit, and the client guards are never
the mechanism.

1. `platform/httpserver` gains `Actor{UserID, Role string}` with `WithActor` and
   `ActorFromContext` — two primitives, no domain type.
2. The identity adapter's private `authenticate` becomes exported `Authenticate` and, beside
   the `domain.User` it already stores for `/auth/me`, writes `httpserver.Actor`.
   `RequireSameOrigin` is extracted from the existing `sameOrigin` check and exported.
3. The composition root passes both as plain `func(http.Handler) http.Handler`:
   `courseHandler.RegisterAdmin(mux, identityHandler.Authenticate, identityHandler.RequireSameOrigin)`.
   **The course module imports nothing from the identity module.** `boundaries_test.go`
   would have permitted `course/application` to import `identity/domain`, so this is a
   choice, not a constraint: a shared kernel of role strings is cheaper to avoid now than to
   unpick later.
4. The course handler maps `httpserver.Actor` to `coursedomain.Actor` explicitly; an
   unrecognised role string is a refusal, not a default. This duplicates three role
   constants across two domains. Accepted, with the trigger: **when a third module needs
   roles, promote them to a shared kernel** rather than duplicating twice more.
5. The policy lives in `course/domain`: `CanManage`, `CanCreate`, `CanAssignOwner`,
   `CanTransition`. The application asks the domain; the handler only translates. No policy
   in a handler, none in SQL.

**Order of checks, and what each stage answers** (NB-8). Same-origin runs **first** on every
mutation, matching the live login handler, so an anonymous cross-site `POST` answers
`403 origin_forbidden` rather than `401`. Then: authenticated → role → course exists →
ownership → domain legality (transition, slug mutability, node membership). Every acceptance
row that exercises a mutation therefore states a valid `Origin` in its precondition.

### D-2 The four `frontend-patterns` decisions

| # | Decision | Rationale (one line) |
|---|---|---|
| D-2a | **Branded ids: adopt** — `CourseId`, `ModuleId`, `LessonId` defined in `domain/course/course.ts`, branded while parsing in `admin-course-schema.ts`, with one documented trust-boundary escape hatch | three string ids pass through the same editor call sites at once, which is the shape §2 says branding protects, and the cost is one `unique symbol` per aggregate |
| D-2b | **`EntityRef`: not adopted.** The owner picker uses a plain local `InstructorOption = {id: string; displayName: string; role: 'instructor' \| 'admin'}` | one variant discriminates nothing and the wire shape is already flat (NB-2); trigger to adopt: a second cross-slice reference kind appears |
| D-2c | **Staleness: two named constants in `admin-queries.ts`**, `adminListStaleMs = 30_000` and `instructorListStaleMs = 300_000`; the session query keeps its documented `staleTime: 0` | naming beats scattering numbers, but a module with one consumer is a policy in name only (NB-1); trigger to promote to a shared module: a second slice imports either constant |
| D-2d | **Write protocol: full-resource PUT of the course tree**, with `expectedUpdatedAt`; no delta protocol, no CQRS leaf helpers | §6 says the question disappears if the API is a full-resource PUT, and it collapses eight endpoints into one; the price is D-3b–D-3d |

Two smaller borrowings, taken explicitly:

- §5 **named invalidation helpers** — `invalidateAfterCourseWrite(client, courseId)` and
  `invalidateAfterCourseDelete(client, courseId)` live beside the resource in
  `admin-queries.ts`. One write invalidates the list and the tree.
- §7 **never throw inside `validateSearch`** — `admin-search.ts` does safeParse per field and
  drops unusable values. Admin routes are `ssr: false`, so a throw paints the router's error
  boundary in the browser. This is the opposite of `catalog-search.ts`, which may throw only
  because its route lane answers `notFound()` with a real 404.
- §1 (`HttpServiceError.method`) is **not** adopted: the admin adapter has one function per
  operation, so it already knows whether a `404` came from a read or a write.

### D-3 The aggregate write

#### D-3a Payload shape and the full admin field set (NB-11, NB-6, NB-13b)

The PUT body is `{expectedUpdatedAt, course: {title, slug, summary, description, instructorId}, modules: [{id, title, lessons: [{id, title, description}]}]}`.

- The **complete** admin field set for a course is `title`, `slug`, `summary`,
  `description`, `instructorId`. `summary` is `NOT NULL CHECK 1–500` in
  `00001_courses.sql` and required by the public schema; edition 1 never mentioned it. Every
  DTO bound equals the database CHECK it maps to: title 1–200, summary 1–500, description
  1–20000, slug ≤120 matching `^[a-z0-9]+(-[a-z0-9]+)*$`.
- **Request DTOs carry no `position`.** The server derives order from array index, writing
  `0..n-1`. A client cannot send a duplicate or gapped position because it cannot send one at
  all. Response DTOs likewise omit it; order is array order.
- The PUT schema **excludes** `status` and `publishedAt` (`additionalProperties: false`
  rejects them with `400 invalid_request`). Lifecycle has its own endpoint.
- A `published` or `archived` course's tree **may** be edited: spec S5 makes edits to a
  published course immediately public, and an archived course must be fixable before it is
  restored.

#### D-3b The concurrency token, and who bumps it (BLK-2)

`courses.updated_at` has only `DEFAULT now()` in `00001_courses.sql`; there is no trigger and
migration 00003 adds none. Therefore, explicitly:

- **Every accepted write ends with**
  `UPDATE courses SET <fields>, updated_at = $clock WHERE id = $1 AND updated_at = $2`,
  **executed even when only modules or lessons changed**. Without that statement the token
  never moves and the optimistic lock silently permits the lost update it exists to prevent.
- `$clock` is the application clock, **truncated to microseconds** before it leaves the
  application — the I03 defect class (briefing §6): `time.Now()` nanoseconds do not survive a
  `timestamptz` round trip, and a token that changes on read is worse than none.
- `updatedAt` and `expectedUpdatedAt` are serialized RFC 3339 with microsecond precision in
  both directions, so the value a client received compares equal to the value it sends back.
- Zero rows affected means another writer moved first: `409 course_modified`.
- A **status transition also bumps `updated_at`** (D-3f) and therefore invalidates
  outstanding tokens (NB-13c). The client refetches through `invalidateAfterCourseWrite`
  after every status change, so the editor never holds a token the server has retired.
- The `CHECK (published_at BETWEEN created_at AND updated_at)` in `00001_courses.sql` stays
  satisfied because `updated_at` only ever moves forward.

#### D-3c Node identity: the payload cannot re-parent another course's rows (BLK-3)

Ids for new modules and lessons are minted by the client (UUIDv4) so the form has stable keys
before saving. Edition 1 claimed "the primary key rejects the rest"; that is false for the
`DO UPDATE` arm, which would happily re-parent a row belonging to another course — including
one the caller may not access.

- Modules: `INSERT … ON CONFLICT (id) DO UPDATE SET … WHERE course_modules.course_id = $courseId`.
- Lessons: the same, with `WHERE course_lessons.module_id = ANY($modulesOfThisCourse)`.
- The adapter compares affected rows to payload rows inside the transaction; a shortfall
  means the payload named a node that is not part of this course, and the whole transaction
  rolls back with `400 invalid_request`. `400` rather than `403`/`404` because it discloses
  nothing about the other course — the caller's payload is simply wrong.
- Ids are also validated for UUID format, non-nil value and uniqueness within the payload
  **before any statement runs** — two nodes sharing one client-minted id are refused by the use
  case with no repository call at all, which is a different rule from the cross-course guard above
  and has its own row (R39). Edition 2 stated this rule in prose only, which is how plan-defense
  row 4 stayed open.
- Rows are **upserted, never deleted-and-reinserted**: lesson ids must survive because Phase 5
  will hang media off them.
- A lesson **may move between modules of the same course** in one PUT (NB-12): its
  `module_id` is part of the upsert, the guard above keeps the move inside this course, and
  delete-missing runs after the upserts so a moved id is never treated as absent.

#### D-3d The size budget, recomputed at 2 bytes per character (BLK-1)

Edition 1 assumed one byte per character. Ukrainian Cyrillic is two bytes in UTF-8 and JSON
does not escape it, so its bounds produced a ~1 MB worst case that its own 512 KiB cap would
refuse: a tree passing every B04 limit failed B05. Revised bounds, with the arithmetic:

| Part | Bound | Worst case (bytes) |
|---|---|---|
| course title, summary, description | 200 + 500 + 20000 chars | 41 400 |
| modules | ≤ 40 × title 200 chars | 16 000 |
| lessons | ≤ 120 × (title 200 + description 1000 chars) | 288 000 |
| ids, keys, structure | ~160 nodes × ~70 B | 11 200 |
| **total** | | **≈ 356 600 B ≈ 348 KiB** |

Cap: `http.MaxBytesReader` at **512 KiB**, as `login` already does at 8 KiB — 32 % headroom
over the worst legal tree. Nginx's `client_max_body_size 1m` is untouched and §5.10 stays
true. Lesson description drops from 2000 to **1000** characters and lessons per course from
200 to **120**; modules from 50 to **40**. Trigger to revisit: a real course that legitimately
needs more than 120 lessons — at which point the cap, the nginx directive and this table move
together, not separately.

#### D-3e Slug mutability (BLK-6)

**Decided: a slug is mutable while the course has never been published, and immutable once it
has.** Rationale: a public URL that has ever resolved must keep resolving, so Q8's 404 always
means "draft or archived" and never "renamed".

- "Never published" is exactly `published_at IS NULL`, which the schema already guarantees
  for drafts (`CHECK (status <> 'draft' OR published_at IS NULL)`) and which survives archive.
- A slug change on a course with `published_at IS NOT NULL` is `409 slug_immutable`.
- A legal rename that collides is `409 slug_taken`, mapped from PostgreSQL's `23505` unique
  violation (NB-10) — never from a pre-check, which two concurrent writers would both pass.
- The same `23505` mapping serves create (B03).

#### D-3f Delete and transition are single guarded statements (BLK-7)

Read-status-then-act is a race under READ COMMITTED: a `DELETE` can commit after its status
check and hard-delete a course another request just published, violating Q5.

- Delete: `DELETE FROM courses WHERE id = $1 AND status = 'draft'` (plus
  `AND instructor_id = $actor` when the actor is an instructor). Zero rows are disambiguated
  by a re-read **inside the same transaction**: no row → `404 course_not_found`; a row with
  another status → `409 course_not_deletable`; a row owned by another → `403 course_forbidden`.
- Transition: `UPDATE courses SET status = $to, published_at = …, updated_at = $clock
  WHERE id = $1 AND status = $from AND updated_at = $expected`. Zero rows re-read in the same
  transaction answer `404`, `409 invalid_status_transition` or `409 course_modified`.
- The status request therefore carries `expectedUpdatedAt` like the tree write, which is what
  orders transitions against PUT and DELETE at all.
- `published_at` on transition: publish sets it to `$clock`; archive leaves it; restore leaves
  it, and sets it to `$clock` if it is somehow null, so the `status <> 'published' OR
  published_at IS NOT NULL` CHECK can never be the thing that fails.

#### D-3g Replay and retry (BLK-5, plan-defense row 3)

Edition 1 claimed "a retried save is idempotent". That is wrong, and it contradicted the
token in the same section.

- **Tree write.** Client-minted ids make the *content* stable and give the form stable keys;
  they do not make the *request* idempotent. A retry after a lost response carries a token the
  first attempt already retired, so it answers `409 course_modified` — from the caller's own
  successful write.
- **The conflict state is therefore not "someone else edited".** On `409 course_modified` the
  client calls `invalidateAfterCourseWrite`, refetches, and presents the stored tree with copy
  that says the course changed since it was loaded and that this may have been the user's own
  save. The user compares and re-saves if needed; nothing is silently discarded.
- **Create has no token, so the unique slug is its replay guard.** A replayed
  `POST /admin/courses` with the same slug answers `409 slug_taken`, mapped from `23505`, so
  at most one course per slug exists no matter how many times the request arrives. Create is
  not idempotent and this plan does not claim it is: on `409 slug_taken` the create dialog
  refetches the list and offers the existing course, rather than asking the user to retry
  blindly.
- **Double submission is prevented at the source**: the create and save buttons are disabled
  while their mutation is in flight (`isSubmitting`), so the ordinary double-click never
  becomes a replay — R30 for the create dialog and R40 for the editor's Save, one row each because
  they are two buttons on two screens.
- No `Idempotency-Key` header is introduced. The reserved parameter in `openapi.yaml` stays
  reserved for payment mutations; a slug-unique create plus a token-guarded update covers the
  replay cases this slice actually has. Trigger to revisit: an operation that is neither
  slug-unique nor token-guarded — the first of those arrives with Phase 8.

### D-4 Who may do what

#### D-4a Role powers (decided by the user, 2026-09-05)

**An instructor may publish, archive and restore courses they own.** Edition 1 raised this as
its one open question; the user answered yes. Admin-only powers are exactly two: assigning or
reassigning an owner (D-4b), and acting on a course the actor does not own.

#### D-4b Owner assignment and reassignment (BLK-8)

- The **carrier is the PUT course section's `instructorId`** — there is no separate
  reassignment endpoint.
- An **admin** may set any user whose role is `instructor` or `admin` and whose `disabled_at`
  is null. Anything else — a student, a disabled account, an unknown id — is
  `400 invalid_owner`, checked in the same transaction as the write so it cannot race a
  role change in 4b.
- An **instructor** must echo their own id or omit the field; a different id is
  `403 course_forbidden` (R12).
- On create, an instructor's course is owned by the instructor; an admin may name the owner.
- A successful reassignment moves the course between owners' lists, which is what P06 observes.

#### D-4c The owner picker mounts only for admins (BLK-9)

`GET /api/v1/admin/instructors` answers `403 forbidden` to instructors (R05), and the editor
is shared by both roles. Therefore:

- The picker **renders and its query mounts only when the actor's role is `admin`**. An
  instructor sees the owner as read-only text (their own display name) and issues no request,
  so no instructor session ever produces a background 403.
- If the instructor list answers `dependency_unavailable`, the admin editor degrades to the
  same read-only owner field and **saving stays possible** — the PUT echoes the current owner.
  A picker outage must not block editing a course.
- The endpoint exists in 4a because spec S1 requires an owner chosen *from users with the
  instructor or admin role*, and 4b ships after 4a. It is admin-only, read-only, returns at
  most 200 non-disabled `instructor`/`admin` accounts as `{id, displayName, role}`, and
  exposes no email, hash or status. It is gated by a new `Role.CanAdministerAccounts()` in
  identity's domain — the predicate 4b reuses.

### D-5 Refusal matrix and business error codes

Every cell is server-enforced and proven by a contract test. Nothing here relies on a client
guard. Mutation cells assume a valid `Origin`; a foreign one is refused first (D-1, NB-8).

| Caller | list | read one | instructors | create | save tree | reassign owner | status | delete |
|---|---|---|---|---|---|---|---|---|
| anonymous | 401 `authentication_required` | 401 | 401 | 401 | 401 | 401 | 401 | 401 |
| student | 403 `forbidden` | 403 | 403 | 403 | 403 | 403 | 403 | 403 |
| disabled account | 401 | 401 | 401 | 401 | 401 | 401 | 401 | 401 |
| instructor, own course | own rows only | 200 | 403 `forbidden` | 201, owner forced to self | 200 | 403 `course_forbidden` | 200 | 204 if draft |
| instructor, foreign course | not listed | 403 `course_forbidden` | 403 | — | 403 | 403 | 403 | 403 |
| admin | all rows | 200 | 200 | 201, owner chosen | 200 | 200 | 200 | 204 if draft |
| any, foreign `Origin` | — | — | — | 403 `origin_forbidden` | 403 | 403 | 403 | 403 |

New codes in the OpenAPI `error.code` enum: `forbidden`, `course_forbidden`, `slug_taken`,
`slug_immutable`, `invalid_owner`, `invalid_status_transition`, `course_not_deletable`,
`course_modified`. Reused unchanged: `authentication_required`, `origin_forbidden`,
`invalid_request`, `course_not_found`, `dependency_unavailable`, `internal_error`.

Unknown id, for any caller past the role gate: `404 course_not_found`. Malformed id:
`400 invalid_request`, before any repository call. An illegal transition is `409`, never a
500 (acceptance criterion 3).

Accepted disclosure: answering a non-owning instructor with `403 course_forbidden` rather than
`404` tells them a course id exists. Spec S6 asks for a distinct ownership refusal; the
surface is staff-only and ids are unguessable UUIDs. Recorded so a later review finds the
reasoning rather than the symptom.

### D-6 Public surface delta

`GET /api/v1/courses/{slug}` answers a new `CourseDetail`: every existing `Course` field plus
`program: [{title, lessons: [{title}]}]`, ordered, titles only. The list keeps returning
`Course`, so pages never carry programs they do not render. Drafts and archived courses still
answer 404 (Q8), so a program is visible only for a published course. `instructorId` is **not**
added to any public schema — ownership is an administrative fact. A slug never changes after
publication (D-3e), so no public URL this response advertises can later 404 through a rename.

Frontend: `domain/course/course.ts` gains the program type, `course-schema.ts` parses it, and
`CourseView` renders it as an ordered list under the description. It is SSR content: with the
D-3d bounds a maximal program adds roughly 8 KB of HTML, which is why it carries titles only.

### D-7 The e2e session budget (BLK-10)

The briefing (§4) records the trap exactly: e2e signs in from the host into the `login` zone
(`10r/m`, `burst=5 nodelay`, keyed by `$binary_remote_addr`), a second consecutive run already
429s, and the recorded fix — shared `storageState` — carried the trigger "when e2e grows".
**This slice grows it, so the trigger has fired and the fix lands here, not later.**

- `e2e/auth.setup.ts` becomes a Playwright `setup` project that signs in **once per role per
  run** (3 requests, back-to-back, inside `burst=5`) and writes `playwright/.auth/<role>.json`.
- `admin.spec.ts` and the parts of `route-guards.spec.ts` that only need an authenticated
  session use `test.use({ storageState })` and perform **zero** sign-ins.
- The specs that genuinely exercise the login form keep doing so: at most 3 form submissions
  per run remain.
- Budget: ≤ 6 host-originated `POST /api/v1/auth/login` per full `make e2e`, of which at most 3
  are consecutive; the run exceeds 60 s, so the sustained rate stays under `10r/m` and no
  burst exceeds 5. A13 measures the actual count rather than asserting this (B19).
- This also makes a second consecutive local run pass, which is the failure the briefing
  predicted and edition 1 left unexamined.

---

## 5. Side effects (Сайд-ефекти)

1. **Gate order reverses.** Ownership makes courses depend on users, so `seed-identities` must
   run **before** `seed`. Four files change together or the gate is a lie: `Makefile` (`check`
   target), `.github/workflows/ci.yml` (`Seed and smoke` step), `README.md` (quick start and
   the CI-order sentence) and `.claude/standards.md` §Машинні гейти. New order:
   `… build → up → seed-identities → seed → smoke → e2e`.
2. **`backend/sql/seed.sql` changes three ways.** Every demo course gains `instructor_id`
   pointing at the instructor fixture's fixed UUID; the conflict clause becomes
   `ON CONFLICT (id) DO UPDATE` so an edited fixture is applied rather than silently ignored on
   a persistent dev volume; and it gains
   `LOCK TABLE courses IN SHARE ROW EXCLUSIVE MODE` — `repository_integration_test.go` runs the
   seed **twice concurrently**, which `DO NOTHING` tolerated and `DO UPDATE` would not.
3. **The existing course repository integration test breaks in three ways** (BLK-12), all fixed
   inside A01: its two direct `INSERT`s (lines 158, 183) omit `instructor_id`; it executes
   `seed.sql` against a database that has never seeded users, so the new foreign key fails; and
   its "failed migration rolls back" subtest builds an `fstest.MapFS` containing a synthetic
   `00003_failure.sql`, whose version now collides with the real `00003_authoring.sql`. The fix
   is to seed identities first, supply `instructor_id`, add the real 00003 to that MapFS and
   renumber the canary to `00004_failure.sql`.
4. **The identity HTTP adapter's surface grows.** `authenticate` → `Authenticate`,
   `sameOrigin` → `RequireSameOrigin`, plus one new route. Behaviour does not change, only
   visibility and the extra `Actor`; the existing identity contract tests must pass unchanged.
5. **`/admin` becomes a layout.** `admin.tsx` keeps the guard and renders `<Outlet/>`; its
   current `EmptyState` body moves aside for `admin.index.tsx`. `admin.$.tsx` stays as the
   unknown-address sink. `route-guards.test.tsx` and `e2e/route-guards.spec.ts` assert on that
   screen and are updated in the task that moves it (A10).
6. **UI-kit debt trigger fires** (briefing §6, `work/report-i06d.md`). The real change is
   "Ukrainian defaults removed, labels required" (NB-16): `Pagination` carries **five**
   Ukrainian strings — two buttons, the page label, the empty label and the `nav` aria-label —
   while `Table.emptyMessage` and `Skeleton.label` are optional-with-Ukrainian-default. All
   become required props supplied by callers through `t()`. The kit stays i18n-free, which is
   why props rather than a `useTranslation` call inside it. Every existing caller is updated in
   the same task.
7. **Translation resources split** into `app/providers/locales/{uk,en}.ts`; keys unchanged.
   `i18n.tsx` is **289 lines today** and grows by the admin vocabulary (NB-18);
   `frontend-patterns` §8 already flagged one file per two locales as not scaling through
   Phase 4. `i18n.test.tsx` imports `resources` and follows the move.
8. **One new runtime dependency: `react-hook-form`.** Brief §24.6 named this exact trigger.
   `@hookform/resolvers` is **not** installed: RHF accepts a plain resolver function and the
   domain validators already return the union we need. Trigger to revisit: a second form family
   needing schema-driven resolution across many fields.
   **TanStack Table and TanStack Virtual are not installed**: the list is ≤50 server-filtered
   rows and the editor renders ≤40 modules. Trigger: client-side multi-column sort or row
   selection on the list, or more than ~500 simultaneous fields.
9. **`make e2e` mutates shared state.** Playwright runs against Compose with
   `fullyParallel: false`, and local volumes persist (`make down` keeps them). A published or
   archived course cannot be deleted (Q5), so a fixed slug would fail the second run with
   `slug_taken`. `admin.spec.ts` therefore mints a per-run slug in an `e2e-admin-*` namespace
   that no catalog assertion matches, and deletes what it may (drafts). Local volumes accumulate
   a few archived e2e courses; CI is always fresh. Named rather than fixed, because the
   alternative — an e2e-only delete path for published courses — would weaken a product
   invariant to serve a test. The **login**-side of the same problem is fixed properly, in D-7.
10. **Unchanged on purpose:** Nginx (admin rides the existing `api` zone at 10 r/s burst 20; a
    tree save is one request; `client_max_body_size 1m` survives D-3d's 512 KiB cap with 2×
    margin), Compose, `.env.example`, Prometheus, the media-worker, the `login` zone rate, and
    the public catalog list response.
11. **SSR crosses none of this.** Admin routes are `ssr: false`, and the SSR container reaches
    the API directly without an `Origin` header (architectural fact #2), so it could not call a
    mutation even if a route tried. The only SSR-visible change is the public program.
    Reachable SSR statuses stay 200/404/3xx/500.

---

## 6. Migration plan (План міграції)

`backend/migrations/00003_authoring.sql`, applied by `make migrate` / `make up` (goose),
before any seed.

**Up:**

1. `ALTER TABLE courses ADD COLUMN instructor_id uuid REFERENCES users(id) ON DELETE RESTRICT;`
   — nullable at first. `RESTRICT`, not `CASCADE`: deleting a user must never delete their
   courses (4b soft-deletes users precisely so this cannot happen).
2. Backfill (NB-14 — the tiebreak and the filter are deliberate, not incidental):
   `UPDATE courses SET instructor_id = (SELECT id FROM users WHERE role IN ('admin','instructor')
   AND disabled_at IS NULL ORDER BY created_at, id LIMIT 1) WHERE instructor_id IS NULL;`
   All three fixtures share one transactional `now()`, so `created_at` ties and **`id` decides**
   — `…0002` (instructor) beats `…0003` (admin). `disabled_at IS NULL` keeps a blocked account
   from inheriting a catalog. B13 asserts the resulting owner is non-disabled.
3. `ALTER TABLE courses ALTER COLUMN instructor_id SET NOT NULL;`
4. `CREATE INDEX courses_by_instructor ON courses (instructor_id, updated_at DESC, id);`
5. `course_modules` (id non-nil uuid PK, `course_id` → `courses` ON DELETE CASCADE, title
   1–200 non-blank, `position integer NOT NULL CHECK (position >= 0)`, timestamps with
   `updated_at >= created_at`), plus `CREATE INDEX course_modules_ordered ON course_modules
   (course_id, position, id);`
6. `course_lessons` (id PK, `module_id` → `course_modules` ON DELETE CASCADE, title 1–200,
   description 0–1000, position, timestamps), plus
   `CREATE INDEX course_lessons_ordered ON course_lessons (module_id, position, id);`

**Ordinary indexes, not deferrable unique constraints** (NB-5, NB-6). Positions are derived
server-side from array index and rewritten wholesale, so a duplicate can only come from an
adapter bug; a deferrable unique constraint would insure against that bug at the cost of a
COMMIT-time failure mode that naive error handling turns into a 500. Contiguity stays a domain
invariant with unit coverage, and reads order by `(position, id)` so even a bug degrades to a
stable order rather than an error. Trigger to add the constraint: a second writer of these
tables that does not rewrite the whole level.

**Why this backfill strategy, and what it costs.** A bare `ADD COLUMN … NOT NULL` fails on any
non-empty `courses` table, and local databases *are* non-empty: `make down` keeps volumes and
`seed.sql` inserts five demo courses with fixed UUIDs. There is no production and no deploy
target (briefing §4), so the only populated databases are developer volumes holding seed data.

- Fresh volume, and CI: `courses` is empty when the migration runs, steps 2–3 are no-ops, and
  the column is `NOT NULL` from the first row.
- Developer volume that has run both seeds: the backfill finds the instructor fixture, step 3
  succeeds, and the next `make seed` reconciles the demo courses to the file's explicit owner.
- Developer volume with courses but **no users** (last touched at Phase 2): step 3 fails loudly
  with "column contains null values" and `make up` stops. Recovery is documented in the
  migration header and the report: `docker compose down -v && make up && make seed-identities
  && make seed`. **This plan deliberately does not delete ownerless courses to make the
  migration pass** — a migration that silently drops rows is a worse default than one that
  stops, and no data of value exists here.

**Down:** drop `course_lessons`, `course_modules`, the index and the column, in that order.
It is destructive of authored content by construction, so it is exercised only in ephemeral
Testcontainers databases, never as an operational step (§7).

**Clean bring-up order:** `make up` (migrate) → `make seed-identities` → `make seed` →
`make smoke` → `make e2e`.

---

## 7. Rollback

- **Application rollback is the supported path.** Redeploy the previous API and frontend images
  and **leave migration 00003 in place**. The new column and tables are additive: the Phase 3
  binary reads `courses` by explicit column list, never `SELECT *`.
- **What would break on a Phase 3 binary is every writer of `courses`** — corrected from edition
  1, which named only `seed.sql` (BLK-12). Those writers are `backend/sql/seed.sql` and the
  fixtures inside `repository_integration_test.go`; both are repository files that revert with
  the images, and there is no third writer, because Phase 3 application code never inserts a
  course.
- **Do not run `goose down 00003` to recover from an application defect.** It drops every module,
  lesson and ownership assignment. A forward migration fixes schema mistakes; the down migration
  exists for tests.
- **Data rollback for a bad edit** is out of scope for 4a — there is no versioning (spec S5). The
  mitigations are that publishing is a separate deliberate action, archiving is reversible (Q3),
  hard delete is impossible once published (Q5), and a conflicting save is refused rather than
  merged (D-3b).
- **Frontend rollback alone is inert as a security measure**: removing the admin screens returns
  the Phase 3 shell and grants nothing, exactly as removing the Phase 3 guards granted nothing.
  The API refusals stand on their own, which is the point of D-1.
- **Gate-order rollback**: the four files of §5.1 revert together; reverting `Makefile` alone
  leaves CI seeding in an order the schema rejects.

---

## 8. Acceptance scenarios

Per `acceptance-spec`: three groups, one observable result per row, one evidence level per row,
exactly one owning task, and every owner able to produce its evidence **when that task lands**.
Rows that stay manual: **none**.

Mutation rows assume a valid `Origin` header unless the row is about `Origin` (D-1, NB-8).

### Happy path

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| P01 | Every course has an owner | fresh database, migrations applied | insert a course without `instructor_id` | PostgreSQL refuses with a not-null violation; no row is created | integration `authoring_integration_test.go` | A01 |
| P02 | Admin creates a draft | admin session, unused slug | `POST /api/v1/admin/courses` | `201`, status `draft`, `publishedAt` null, server-minted uuid | contract `admin_handler_test.go` | A05 |
| P03 | A draft is invisible publicly | the course from P02 | `GET /api/v1/courses/{slug}` | `404 course_not_found`, and it is absent from `GET /api/v1/courses` | contract | A05 |
| P04 | Instructor owns what they create | instructor session | `POST` with no `instructorId` | `201` with `instructorId` equal to the caller | contract | A05 |
| P05 | Admin picks the owner at create | admin session, instructor fixture | `POST` with the instructor's `instructorId` | `201` with that owner; the course appears in the instructor's own list | contract | A05 |
| P06 | Admin reassigns the owner | course owned by instructor A, admin session | `PUT` with instructor B's `instructorId` | `200`; the course now appears in B's list and no longer in A's | contract | A05 |
| P07 | Owner candidates are listable | admin session, three fixtures seeded | `GET /api/v1/admin/instructors` | `200` with exactly the instructor and admin fixtures as `{id, displayName, role}`; no email member present | contract | A04 |
| P08 | A tree is saved whole | draft course exists | `PUT` 2 modules with 3 lessons, client-minted ids | `200`; a following `GET` returns the same ids, titles and order | integration | A03 |
| P09 | Reordering persists | course with 3 modules | `PUT` with the module array reversed | `GET` returns them reversed, positions `0,1,2` | integration | A03 |
| P10 | A lesson moves between modules | course with 2 modules, lesson L in the first | `PUT` with L in the second module's array | `GET` shows L under the second module with the same id; nothing else changed | integration | A03 |
| P11 | Publishing makes a course public | draft with valid fields | `POST /status {"status":"published"}` | `200` status `published`, non-null `publishedAt`; `GET /api/v1/courses/{slug}` answers `200` | contract | A06 |
| P12 | Archiving withdraws it | published course | `POST /status {"status":"archived"}` | `200` status `archived`; the public URL answers `404` and it leaves the catalog list | contract | A06 |
| P13 | Restoring republishes it | archived course | `POST /status {"status":"published"}` | `200` status `published`, `publishedAt` still non-null; the public page answers `200` | contract | A06 |
| P14 | An instructor publishes their own course | instructor session, own draft | `POST /status {"status":"published"}` | `200` status `published` (D-4a) | contract | A06 |
| P15 | The public page shows the program | published course with 2 modules and 3 lessons | `GET /api/v1/courses/{slug}` | `program` carries module and lesson titles in order, with no description, id or position member | contract | A07 |
| P16 | Admin list renders | admin session, ≥1 course | open `/admin` | a table of title, slug, status, owner and updated time, each row linking to its editor | component `CourseListView.test.tsx` | A10 |
| P17 | The kit carries no Ukrainian of its own | none | render `Pagination`, `Table` with no rows and `Skeleton` with English labels supplied | all five `Pagination` strings, the table's empty text and the skeleton label render the supplied English, and the rendered output — including the `nav` aria-label — contains no Cyrillic character | component `presentation/ui/ui.test.tsx` | A08 |
| P18 | The admin screen is localized | locale `en`, empty course list | open `/admin` | the empty state and both pagination buttons render English text | component `CourseListView.test.tsx` | A10 |
| P19 | The editor saves a tree | admin session, course loaded | add a module and a lesson, press Save | one `PUT` is sent carrying both new nodes with client-minted uuid ids and the loaded `expectedUpdatedAt`, and the saved tree is rendered back | component `CourseEditorView.test.tsx` | A11 |
| P20 | The full admin journey in a browser | Compose up, seeded fixtures | sign in as admin, create, add a module and lesson, save, publish, open the public URL | the public page shows the module and lesson titles and the admin row shows `published` | e2e `e2e/admin.spec.ts` | A14 |
| P21 | An instructor manages their own course | Compose up, instructor fixture | sign in as instructor, create, edit, publish an own course | the list shows only own courses, no owner picker is present, and every action succeeds | e2e | A14 |

### Boundaries

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| B01 | An empty course may be published | draft with valid fields, no modules | publish | `200` status `published` (Q4) | contract | A06 |
| B02 | Empty admin list | instructor with no courses | `GET /api/v1/admin/courses` | `200` with `items: []` and `total: 0`; the screen shows the localized empty state | component | A10 |
| B03 | Slug uniqueness survives concurrency | slug `s` exists | two concurrent creates with slug `s` | exactly one `201`; the other is `409 slug_taken` mapped from `23505`, and one course exists | integration | A03 |
| B04 | A slug is mutable before publication | never-published draft | `PUT` with a new free slug | `200`; a later publish makes the new public URL resolve | contract | A05 |
| B05 | A maximal legal tree saves | admin session | `PUT` 40 modules and 120 lessons, every text field at its bound and entirely Cyrillic | `200`, and the request body measured in the test is under 512 KiB | contract | A05 |
| B06 | Tree bounds are refused | admin session | `PUT` with 41 modules, or 121 lessons, or a 1001-character lesson description | `400 invalid_request`; a following `GET` shows the tree unchanged | contract | A05 |
| B07 | Oversized bodies are refused | admin session | `PUT` a body over 512 KiB | `400 invalid_request` and a bounded response | contract | A05 |
| B08 | Summary bounds | admin session | `PUT` with a 500-character summary, then a 501-character one | `200`, then `400 invalid_request` | contract | A05 |
| B09 | A deep tree stays operable | 40 modules × 3 lessons in the editor | tab through the form | every module heading and lesson title field is reachable by keyboard and axe reports no violations | component | A11 |
| B10 | Pagination boundary | 51 courses, `limit=50` | request the second page | 1 item with `total: 51` | integration | A03 |
| B11 | Timestamps survive the round trip | publish a course, read it back from PostgreSQL | publish then `GET` | `publishedAt` matches the stored value to the microsecond and `published_at <= updated_at` holds | integration | A03 |
| B12 | The token round-trips | admin session | `GET` a course, `PUT` with the returned `updatedAt` | `200`, and the response's new `updatedAt` differs from the sent one — including when only a lesson title changed | contract | A05 |
| B13 | Migration over a populated database | 5 demo courses and 3 identity fixtures present, one of them disabled | apply migration 00003 | every course carries a non-null `instructor_id`, the owner is the non-disabled instructor fixture, and the column is `NOT NULL` | integration | A01 |
| B14 | Migration over courses without users | courses present, `users` empty | apply migration 00003 | the migration fails with a not-null violation, leaves the previous schema in place and deletes no course | integration | A01 |
| B15 | Down migration | migrated database | `goose down` one step | `course_lessons`, `course_modules` and `instructor_id` are gone and the catalog queries still execute | integration | A01 |
| B16 | The existing catalog suite stays green | migration 00003 committed | `make integration` | `repository_integration_test.go` passes, including its concurrent double seed and its failing-migration subtest | `make integration` | A01 |
| B17 | Contract and client stay in step | all OpenAPI changes committed | `make check-generated` | sqlc and Orval output matches the committed generated files; exit 0 | `make check-generated` | A07 |
| B18 | The ordered gate passes | clean checkout | `make check` | twelve steps exit 0, with `seed-identities` before `seed` | `make check` | A15 |
| B19 | The e2e login budget fits the zone | Compose up, `login` bucket cold | run `make e2e` twice in a row | both runs pass; the counted host-originated `POST /auth/login` per run is at most 6, with at most 3 consecutive | `make e2e` + the count printed by the setup project | A13 |

### Refusals

| # | Scenario | Precondition | Action | Observable result | Evidence | Owner |
|---|---|---|---|---|---|---|
| R01 | Anonymous is refused on every read | no session cookie | `GET` list, `GET` one, `GET` instructors | each answers `401 authentication_required` and no repository method is called | contract | A04 |
| R02 | A student is refused on every read | student session | the same three reads | each answers `403 forbidden` | contract | A04 |
| R03 | A disabled account is refused | instructor fixture with `disabled_at` set | `GET` list | `401 authentication_required`, identical to an unknown session | contract | A04 |
| R04 | A foreign instructor cannot read | instructor B, course owned by A | `GET /api/v1/admin/courses/{id}` | `403 course_forbidden` | contract | A04 |
| R05 | An instructor cannot list owner candidates | instructor session | `GET /api/v1/admin/instructors` | `403 forbidden` | contract | A04 |
| R06 | Unknown ids are not found | admin session, unused uuid | `GET /api/v1/admin/courses/{uuid}` | `404 course_not_found` | contract | A04 |
| R07 | Malformed ids are rejected before lookup | admin session | `GET /api/v1/admin/courses/not-a-uuid` | `400 invalid_request` and no repository call | contract | A04 |
| R08 | Anonymous is refused on every mutation | no cookie, valid `Origin` | `POST` create, `PUT`, `POST /status`, `DELETE` | each answers `401 authentication_required` | contract | A06 |
| R09 | A student is refused on every mutation | student session, valid `Origin` | the same four mutations | each answers `403 forbidden` | contract | A06 |
| R10 | Cross-site mutation is refused first | valid session cookie, foreign `Origin` | the same four mutations | each answers `403 origin_forbidden` with no state change, whatever the caller's role | contract | A06 |
| R11 | A foreign instructor cannot write | instructor B, course owned by A | `PUT`, `POST /status`, `DELETE` | each answers `403 course_forbidden` and a following read shows the course unchanged | contract | A06 |
| R12 | An instructor cannot reassign ownership | instructor owns the course | `PUT` with a different `instructorId` | `403 course_forbidden`; a following read shows the original owner | contract | A05 |
| R13 | An ineligible owner is refused | admin session | `PUT` with a student's id, then a disabled instructor's id, then an unknown uuid | each answers `400 invalid_owner`; the owner is unchanged | contract | A05 |
| R14 | Illegal transitions are refused | published course | `POST /status {"status":"draft"}` | `409 invalid_status_transition` — never a 500 — and the status is unchanged | contract | A06 |
| R15 | A published course cannot be deleted | published, and separately archived, course | `DELETE` | `409 course_not_deletable` and the course still reads back | contract | A06 |
| R16 | Delete cannot race a publish | draft course; a publish commits between the delete's read and its write | `DELETE` issued concurrently with the publish | either the delete removes a still-draft course or it answers `409 course_not_deletable`; a published course is never deleted | integration | A03 |
| R17 | A draft deletes with its tree | draft with 2 modules and 3 lessons | `DELETE` | `204`; the course, its modules and its lessons are all gone | integration | A03 |
| R18 | A concurrent save is refused, not merged | two clients loaded the same tree | the second `PUT` carries the stale `expectedUpdatedAt` | `409 course_modified`; a read returns the first writer's tree intact | integration | A03 |
| R19 | A payload cannot steal another course's nodes | courses X and Y, Y owning module M and lesson L | `PUT` on X carrying M's id, then L's id | each answers `400 invalid_request`; both X and Y read back unchanged and M is still parented to Y | integration | A03 |
| R20 | A published course's slug is immutable | published, and separately archived, course | `PUT` with a changed slug | `409 slug_immutable`; the slug is unchanged and the public URL still resolves | contract | A05 |
| R21 | A rename onto a taken slug is refused | never-published draft, slug `s` taken elsewhere | `PUT` with slug `s` | `409 slug_taken`; the course keeps its slug | contract | A05 |
| R22 | Lifecycle members are rejected by the write schema | admin session | `PUT` carrying `status`, then `publishedAt`, then `position` | each answers `400 invalid_request` | contract | A05 |
| R23 | A transition retires outstanding tokens | admin session, tree loaded | publish the course, then `PUT` with the pre-publish `expectedUpdatedAt` | `409 course_modified` | contract | A06 |
| R24 | A dependency failure degrades safely | the repository port returns dependency failure | `List` and `Save` | `503 dependency_unavailable`; no SQL, credential or internal detail appears in the result | application test with a fake port | A02 |
| R25 | Cancellation is not a fault | the caller's context is cancelled mid-call | `Save` | the use case reports cancellation, distinct from dependency failure and from a business refusal | application test with a fake port | A02 |
| R26 | A cancelled save leaves nothing behind | tree write cancelled mid-transaction | `Save` then re-read | the stored tree is exactly the pre-write tree; no partial module or lesson exists | integration | A03 |
| R27 | Business codes reach the domain intact | MSW answers `409 course_modified`, then `409 slug_taken`, then `403 forbidden`, then `500` | call the corresponding use cases | the first three surface as their named domain outcomes and the last as `dependency_unavailable`; none collapses into a generic failure | adapter test with MSW | A09 |
| R28 | The conflict state shows the stored tree | editor open, the server answers `409 course_modified` | press Save | the tree is refetched and rendered, and the message says the course changed since it was loaded and the save may already have landed — it does not assert another person edited it | component | A11 |
| R29 | A replayed create makes at most one course | admin session, slug `s` free | send the same create twice | the first answers `201`, the second `409 slug_taken`, and exactly one course with slug `s` exists | contract | A05 |
| R30 | Double submission is impossible | create dialog open, mutation in flight | press Create twice quickly | exactly one request is sent and the button is disabled between them | component | A11 |
| R31 | The owner picker is admin-only | instructor session, editor open | inspect the editor | the owner is read-only text and no request to `/admin/instructors` is made | component | A11 |
| R32 | An unavailable picker does not block saving | admin session, `/admin/instructors` answers 503 | open the editor and press Save | the owner field renders read-only with the current owner and the save succeeds | component | A11 |
| R33 | The client guard is not the mechanism | student signed in, in the browser | call `/api/v1/admin/courses` directly through `page.request` | `403 forbidden`, even though the SPA never renders the screen | e2e | A14 |
| R34 | Unusable admin search params disappear | admin session | open `/admin?status=nonsense&page=abc&x=1` | the list renders with default filters, no error boundary appears, and the URL keeps what the user typed | route test `admin-search.test.ts` | A10 |
| R35 | A session that ends mid-edit is handled | editor open, session revoked server-side | press Save | the shell redirects to `/login` carrying the editor address as `returnTo` (amended 2026-09-05 at review: the transient editor alert is unobservable by design — the immediate redirect is the product behavior established by identity E10) | component | A11 |
| R36 | An invalid form is refused before it is sent | editor open | save with a blank title and a 300-character slug | field-level messages appear and no request is sent | component | A11 |
| R37 | The list survives an unavailable primary read | admin session, `GET /api/v1/admin/courses` answers `503 dependency_unavailable` | open `/admin` | the kit's `ErrorState` renders with a working retry control, and the retry re-issues the list request | component `CourseListView.test.tsx` | A12 |
| R38 | The editor survives an unavailable primary read | admin session, `GET /api/v1/admin/courses/{id}` answers `503 dependency_unavailable` | open the editor address | the kit's `ErrorState` renders with a working retry control; no empty form is shown as if the course had no content | component `CourseEditorView.test.tsx` | A12 |
| R39 | Duplicate node ids in one payload are refused before any statement | admin session | `Save` a tree whose two modules share one client-minted id, then one whose two lessons do | the use case returns the invalid-tree outcome and the repository is never called — the guard runs before the transaction opens (D-3c) | application test with a fake port | A02 |
| R40 | The editor's Save cannot be double-submitted | editor open, a save in flight | press Save twice quickly | exactly one `PUT` is sent and the button is disabled between the presses | component `CourseEditorView.test.tsx` | A12 |

Row count: 21 happy, 19 boundary, 40 refusal = **80**. Manual rows: 0. Every task in §9 owns at
least one row; no row has two owners.

---

## 9. Task cut

Cut along **executable evidence and environment boundaries**, per `docs/identity/report.md`
§I06 п.8. Edition 1's cut failed its own rule three times (BLK-4, BLK-11, BLK-12): ownership was
assigned before checking what each task could observe at landing time. Edition 2 states the
environment for every task and verifies that its rows are observable there.

Each backend task that edits `openapi.yaml` or a `.sql` query **runs `make generate` for its own
paths and commits the output** (NB-9), so trunk is never red at `check-generated` between tasks.

| ID | Task | Est. | Depends on | Environment / evidence | Owns |
|---|---|---:|---|---|---|
| A01 | Migration 00003; gate reorder across `Makefile`/`ci.yml`/`README`/`standards.md`; `seed.sql` owner, lock and conflict clause; the three fixes to `repository_integration_test.go` | 2h | GATE 2 green | Testcontainers PostgreSQL (`make integration`) + `make -n check` | P01, B13, B14, B15, B16 |
| A02 | Authoring domain rules, ports and application use cases, proven against fakes: policy, transitions, slug mutability, tree invariants and bounds, payload id validation, deadlines, cancellation | 1.75h | A01 | Go unit tests, stdlib and fakes only (`make test`) | R24, R25, R39 |
| A03 | PostgreSQL authoring adapter and sqlc queries: tree read, transactional write with the guards of D-3b/D-3c, guarded delete and transition, owner-scoped list, `23505` mapping | 2h | A02 | Testcontainers PostgreSQL (`make integration`) | P08, P09, P10, B03, B10, B11, R16, R17, R18, R19, R26 |
| A04 | The authorization seam and the three protected reads: `httpserver.Actor`, exported `Authenticate`/`RequireSameOrigin`, composition wiring, `GET` list, `GET` one, `GET /admin/instructors`, their OpenAPI paths and generated output | 2h | A03 | Go contract tests against `openapi.yaml` (`make test`) | P07, R01–R07 |
| A05 | Create and tree write: `POST /admin/courses`, `PUT /admin/courses/{id}`, size and field bounds, slug mutability, owner assignment, replay behaviour, their OpenAPI and generated output | 2h | A04 | Go contract tests | P02–P06, B04, B05, B06, B07, B08, B12, R12, R13, R20, R21, R22, R29 |
| A06 | Lifecycle: `POST /status`, `DELETE`, the mutation half of the refusal matrix, their OpenAPI and generated output | 1.5h | A05 | Go contract tests | P11, P12, P13, P14, B01, R08, R09, R10, R11, R14, R15, R23 |
| A07 | Public program: repository read, `CourseDetail` schema, frontend domain/schema/`CourseView` rendering, final drift check | 1.5h | A06 | Go contract test + Rstest component test + `make check-generated` | P15, B17 |
| A08 | UI-kit i18n debt and locale split: required label props on `Table`/`Pagination`/`Skeleton`, every existing caller updated, resources moved to `locales/{uk,en}.ts` with the admin vocabulary added | 1h | — | Rstest (`make test`) | P17 |
| A09 | Frontend admin adapter and application: branded ids and tree types, port, use cases, zod schemas, repository, composition and invalidation helpers | 1.5h | A06, A08 | Rstest with MSW (`make test`) | R27 |
| A10 | Admin list route and shell restructure: `admin.tsx` → layout with `Outlet`, `admin.index.tsx`, never-throwing search parsing, status tabs, pagination, empty and error states, uk/en, existing guard tests updated | 2h | A09 | Rstest route and component tests | P16, P18, B02, R34 |
| A11 | Course editor route with React Hook Form: create dialog, tree editing and reorder, delete, status controls, owner field, conflict and refusal states, accessibility | 2h | A10 | Rstest component tests including axe | P19, B09, R28, R30, R31, R32, R35, R36 |
| A12 | Degradation and in-flight guards on both admin screens: the primary-read failure path rendering the kit's `ErrorState` with a working retry on the list and on the editor, and the editor Save's in-flight guard | 0.75h | A11 | Rstest component tests with MSW forcing 503 | R37, R38, R40 |
| A13 | Shared browser sessions: `auth.setup.ts`, the `setup` project, `route-guards.spec.ts` converted where it does not need the form, and the per-run login count printed and asserted | 1h | A12 | Playwright against Compose (`make e2e`) | B19 |
| A14 | The admin browser journey: `admin.spec.ts` for the admin and instructor journeys and the direct-API refusal, per-run slug namespace | 1.5h | A13 | Playwright against Compose (`make e2e`) | P20, P21, R33 |
| A15 | Ordered gate and report: full `make check` from a clean checkout, `docs/admin/report.md` with measured results, limitations and next task | 1h | A01–A14 | `make check`, twelve steps | B18 |

**Total: 23.5 h** — the arithmetic of the column above (six tasks at 2 h, one at 1.75 h, four at
1.5 h, three at 1 h, one at 0.75 h). Edition 1 wrote "≈19 h" over a column summing to 19.5 h;
NB-17 is the reason this number is computed rather than estimated.

Sequencing notes:

- A01 lands first because the gate reorder changes what a green `make check` means, and because
  every later Go task runs against the migrated schema.
- **A04 before A05 and A06** — and A04 owns the refusal rows only for the paths it registers.
  Edition 1 gave A04 rows covering "each of the seven admin endpoints", four of which A05 had not
  yet created; an unregistered path answers the mux's 404, not 401/403, so A04 could not have
  produced its own evidence (BLK-4). The mutation half of the matrix (R08–R11) now lands with the
  lifecycle task that completes the mutation surface.
- A02 and A03 are one layer split by **evidence environment**, not by layer name: A02 is
  everything provable with fakes in `make test`, A03 is everything that needs real PostgreSQL.
  A09 and A10 split the same way (MSW versus a rendered route). Both pairs now own rows, which is
  what NB-7 was really asking for; a literal merge would produce 3.5-hour tasks, over this
  phase's two-hour ceiling.
- A08 has no dependency and can land any time before A09; it is placed early so the admin screens
  are never written against Ukrainian-defaulted kit components.
- **A12 exists because A10 and A11 are already at the two-hour ceiling.** Edition 2 named the
  primary-read failure path in prose (§2, §3, §9 all said "empty and error states") without a row
  that observes it, which is how GATE 2 round 2 left defence row 2 open. Rather than pretend the
  extra tests fit inside two full tasks, the degradation path becomes its own small task, cut the
  way the rest of this plan is cut — around one class of executable evidence (a primary read
  answering 503, and a mutation still in flight) rather than around a screen. It lands after both
  screens exist, because it asserts on both.
- A13 before A14, so the admin spec is written against shared storage state from the start rather
  than converted afterwards.

---

## 10. Disposition of the 18 non-blocking findings

The repository's ledger tradition: addressed or explicitly retained, never silent.

| # | Disposition | Where |
|---|---|---|
| NB-1 `staleness.ts` has one consumer | **Addressed** — module dropped; two named constants in `admin-queries.ts`, with the promotion trigger | D-2c, §2 |
| NB-2 `InstructorRef.kind` discriminates nothing | **Addressed** — `EntityRef` not adopted; plain `InstructorOption` | D-2b |
| NB-3 `CourseFormFields.tsx` shares one field | **Addressed** — file dropped; extract inside A11 if repetition appears | §2 |
| NB-4 `AdminStates.tsx` duplicates the kit | **Addressed** — file dropped; views call the kit with `t()` | §2, §3 |
| NB-5 deferrable unique guards a precluded case | **Addressed** — plain ordered indexes; contiguity is a domain invariant; the trigger to add the constraint is named | §6 |
| NB-6 position semantics undecided | **Addressed** — request and response DTOs carry no `position`; the server derives it from array index | D-3a |
| NB-7 A02/A08 are layer-label cuts | **Retained in part** — the two tasks stay but are re-cut on evidence environment and now own rows; a literal merge would exceed the two-hour ceiling | §9 sequencing notes |
| NB-8 R01's 401 contradicts the check order | **Addressed** — same-origin runs first and is stated; mutation rows carry "valid `Origin`" preconditions; R10 owns the foreign-origin case | D-1, §8 preamble |
| NB-9 `make generate` ownership | **Addressed** — every backend task regenerates its own paths | §9 preamble |
| NB-10 `slug_taken` from `23505` | **Addressed** — mapped from the unique violation, never a pre-check; B03 is a concurrent row | D-3e, B03 |
| NB-11 `summary` invisible | **Addressed** — the full admin field set is named with DB-matching bounds; B08 is the bound row | D-3a, B08 |
| NB-12 cross-module lesson move | **Addressed** — supported course-wide, with delete-missing ordered after the upserts; P10 is the row | D-3c, P10 |
| NB-13 PUT versus status | **Addressed** — (a) published and archived trees are editable, (b) the schema excludes `status`/`publishedAt`, (c) a transition bumps the token; R22 and R23 are the rows | D-3a, D-3b |
| NB-14 backfill tiebreak and disabled accounts | **Addressed** — `disabled_at IS NULL` filter and the id tiebreak stated; B13 asserts a non-disabled owner | §6 step 2 |
| NB-15 "seven paths" is an operation count | **Addressed** — "4 new paths, 7 operations, plus one extended path" | §1, §2 |
| NB-16 "three strings" undercounts | **Addressed** — five `Pagination` strings named plus the two defaulted props; P17 asserts the aria-label too | §5.6, P17 |
| NB-17 total hours contradicts the column | **Addressed** — 23.5 h, the computed sum | §9 |
| NB-18 "~450 lines" is invented | **Addressed** — "289 lines today, growing by the admin vocabulary" | §5.7 |

Retained: **NB-7 only**, and only its literal-merge remedy — the finding's substance (every task
owns rows) is satisfied.
