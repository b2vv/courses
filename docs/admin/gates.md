# Phase 4a gates

## GATE 2 — plan-critique (fable, fresh context)

Date: 2026-09-05. Target: `docs/admin/plan.md`. Context read: `docs/admin/spec.md`
(GATE 1), `.claude/standards.md`, `AGENTS.md`, `work/CTO-RESEARCH.md` §1/§6, plus
spot-checks of the cited code (migrations, seeds, Makefile/ci, boundaries test, identity
handler, UI kit, frontend-patterns, integration tests).

Method: four independent lenses, each a fresh subagent with a clean context (no fork):
AI-slop, logic holes, excess complexity, edge cases. **Lenses: 4 launched / 4 returned.**
Critic tier: `fable`; plan author's tier is unrecorded → treated as `opus` per
`multi-model-roles` §3, so independence holds. Blocking candidates were re-verified by the
synthesizer against the repository files, not taken on faith.

The plan's factual grounding is otherwise strong: every checkable citation held
(frontend-patterns §§1/2/5/6/7/8, private `authenticate`/`sameOrigin`, `boundaries_test.go`
permissiveness, seed conflict clauses, nginx zones and `client_max_body_size`, the i06d UI
debt, I06 п.8, the current `seed → seed-identities` gate order and the coherence of the
§5.1 reversal, the 49-row and 12-step counts). The findings below are what did not hold.

This critique is read-only by mandate: blocking fixes are recorded as **proposed plan
edits**, not applied. All 12 blocking findings therefore remain open.

### Blocking

**BLK-1 [edge] A maximal legal tree does not fit the plan's own 512 KiB body cap**
- where: §4 D-3 bounds ("200 × 2000 stays comfortably under the edge's `client_max_body_size 1m`"), rows B04/B05
- why: Ukrainian Cyrillic is 2 bytes/char in UTF-8 and JSON does not escape it. Lesson descriptions alone: 200 × 2000 × 2 = 800,000 B ≈ 781 KiB > 512 KiB. Add titles (≤200 chars each), the course `description` (schema allows 20,000 chars = ~40 KiB) and JSON structure, and a worst legal payload is ~1 MB — refused by the plan's own `MaxBytesReader` and brushing nginx's `1m`. A tree that passes every B04 limit fails B05. The plan's arithmetic assumed 1 byte/char. Verified against `backend/migrations/00001_courses.sql` (description CHECK 1–20000).
- proposed edit: redo the D-3 arithmetic at 2 bytes/char; either raise the cap (≥2 MiB `MaxBytesReader` **and** a named nginx `client_max_body_size` change — §5.10 "Nginx unchanged" then becomes false) or shrink the tree bounds so the worst case fits with margin. Add an acceptance row: a maximal-legal all-Cyrillic tree saves with 200.

**BLK-2 [logic] Nothing in the plan ever bumps `courses.updated_at`, so the optimistic lock can never trip**
- where: §4 D-3 (`WHERE id = $1 AND updated_at = $2`), R11, A02
- why: verified — `00001_courses.sql` gives `updated_at` only `DEFAULT now()`, there is no trigger, and migration §6 adds none. The plan never states that the tree write SETs `updated_at` (including when only children changed). Without that, writer 1 commits, `updated_at` stays put, and writer 2's stale token still matches — R11 becomes a silent lost update, the exact failure D-3 exists to prevent. Also unstated: `expectedUpdatedAt` JSON serialization precision, the I03 defect class (briefing §6) on the comparison path — B08 covers `publishedAt` only, and no row does GET → PUT-with-returned-`updatedAt` → 200.
- proposed edit: D-3 states every write ends with `SET …, updated_at = <application clock, microsecond-truncated> WHERE id = $1 AND updated_at = $2`, executed even when only children changed; `updatedAt` serialized at full microsecond precision both directions. Add a contract row: load tree → PUT with returned `updatedAt` → 200; repeat with the stale value → 409.

**BLK-3 [edge + logic — 2 lenses] Upsert-by-primary-key captures rows belonging to other courses**
- where: §4 D-3 bullet 1 ("the primary key rejects the rest")
- why: the sentence is false for the `DO UPDATE` arm. A payload carrying a module/lesson id that exists under a *different* course is not rejected — it is updated, re-parenting the foreign row into this course (and ripping future Phase 5 media out of the other aggregate). Ownership checks gate the course in the URL, not ids in the payload, so an instructor can mutate an aggregate they cannot access. No acceptance row covers a foreign-id payload.
- proposed edit: D-3 states the guard — `DO UPDATE … WHERE course_modules.course_id = <this course>` (lessons via membership in this course's modules) with an affected-rows check mapping a mismatch to `400`/`409` — plus a refusal row: a PUT reusing another course's id is refused and both courses read back unchanged.

**BLK-4 [edge + logic — 2 lenses] A04 owns refusal rows for endpoints that only exist after A05**
- where: §8 R01/R02/R03 ("each of the seven admin endpoints"), R12 ("POST, PUT and DELETE"), owned by A04; §9 sequencing note
- why: A04 registers only the three reads; create/save/status/delete land in A05, which depends on A04. An unregistered path answers the mux's 404/405, not 401/403 — so A04 cannot produce its owned evidence at landing time, contradicting "each task owns named acceptance rows and produces its own verification evidence" and hollowing the "refusals proven before mutations" sequencing claim.
- proposed edit: split ownership — A04 owns the matrix for the three reads (+R13/R14); A05 owns extending R01/R02/R03/R12 to the four mutations. Or A04 registers all seven routes behind the full middleware chain with mutation handlers stubbed — then the plan says so explicitly.

**BLK-5 [edge + logic — 2 lenses] "A retried save is idempotent" is contradicted by the plan's own concurrency token, and the conflict state's behavior is undefined**
- where: §4 D-3 bullet 1 vs the `expectedUpdatedAt` mechanism in the same section; §2 `AdminStates.tsx`; A10
- why: a retry after a lost response carries the now-stale `expectedUpdatedAt` and gets `409 course_modified` from its own successful write. A test or client built to the plan's sentence is wrong; the conflict copy tells the user someone else edited when nobody did, and nothing says the conflict state refetches so the user can see their save landed.
- proposed edit: correct the sentence (client-minted ids give stable keys and idempotent *content*, not an idempotent request). Specify: on `409 course_modified` the client refetches via the invalidation helper and presents the stored tree. Add a component row for the conflict state (A10).

**BLK-6 [edge] Slug mutability on PUT is never decided**
- where: §1 PUT row, D-3, D-6; B03 covers create only
- why: the plan never says whether PUT may change `slug`, how a slug collision on *update* maps, or what a rename does to a published course's public URL (the old URL 404s — counterfeiting Q8's "draft or archived" semantics). No acceptance row exercises PUT-with-changed-slug.
- proposed edit: decide in D-3/D-6 — slug immutable after publish (refused on change), or freely mutable with `409 slug_taken` on collision and an explicit note that the old public URL 404s. Add boundary rows for PUT slug change and PUT slug collision.

**BLK-7 [edge] Delete (and transition) can race a status change and violate Q5's invariant**
- where: D-5/R09 "draft only"; D-3 says nothing about delete/transition atomicity
- why: implemented as read-status-then-act, a `DELETE` racing a publish commits after the status check and hard-deletes a published course under READ COMMITTED — the product invariant Q5 protects. The status endpoint carries no `expectedUpdatedAt`, so nothing orders transitions against PUT/DELETE. R09/R10 are sequential and cannot catch it.
- proposed edit: one D-3 line — delete is a single guarded statement (`DELETE … WHERE id = $1 AND status = 'draft'`, zero rows disambiguated inside the same transaction) and transition legality is `UPDATE … WHERE status = $from`.

**BLK-8 [logic] Owner reassignment has no stated carrier and no owning happy-path row**
- where: §1 endpoint table vs D-5 "reassign owner" column vs §8
- why: D-5 promises "every cell proven by a contract test", yet no endpoint is described as carrying `instructorId` on update, D-3's tree never mentions it, and no row exercises a successful admin reassignment (R06 proves only the instructor refusal). Also undecided: validation of the new owner's role (S1/Q2) and the answer to reassignment onto a student.
- proposed edit: state that the PUT course section carries `instructorId`; admin may set any non-disabled instructor/admin (else `400`), an instructor must echo their own id. Add a happy-path row "admin reassigns owner → 200, course appears in the new owner's list" owned by A05.

**BLK-9 [edge] The shared editor fires the admin-only owner-picker query for instructors**
- where: D-4, R07, §2 `CourseEditorView` — never connected
- why: `GET /admin/instructors` answers `403` to instructors (R07), but the editor is shared by both roles and the plan never says the picker (and its `reference`-tier query) mounts only for admins. Unconditional, every instructor session performs a request that 403s — background noise at best, the editor's error state at worst; P16 trips on it late, in A11. Also undecided: the admin editor's behavior when the instructor list is `dependency_unavailable`.
- proposed edit: state in D-4/§2 that the picker renders and its query mounts only for admin actors; instructors see the owner read-only; an unavailable instructor list degrades to read-only owner without blocking save. Add a component-test row.

**BLK-10 [edge] The grown e2e suite's login-zone budget is unexamined**
- where: §5.9 (slug reuse only) and A11
- why: the briefing (§4) records the exact trap — e2e logs in from the host into the `login` zone (10 r/m, burst 5), a second run already 429s, and the named fix (shared `storageState`) carries the trigger "when e2e grows". This plan grows it: P15 + P16 + R16 add at least three sign-ins on top of the existing route-guards logins. A11 can be green alone and red in `make check`'s order, or flaky on the second local run.
- proposed edit: §5.9/A11 either counts the per-run logins and shows they fit the burst, or adopts shared `storageState`/API-seeded sessions for the admin spec — the recorded trigger has fired.

**BLK-11 [logic] P14 is owned by A07 but observable only after A09**
- where: §8 P14 ("open `/admin`", empty list, English) owned by A07; §9 A07 (deps: —, evidence `ui.test.tsx`)
- why: the row's action requires the admin list route, built in A09 which depends on A07. A UI-kit test proves the kit renders caller-supplied English labels — a different observable than the row states. The task DAG and the acceptance ownership contradict each other at this row.
- proposed edit: reword P14 to the kit-level observable A07 can produce, or move P14's ownership to A09 and give A07 a kit-level row.

**BLK-12 [logic] Migration 00003 breaks the existing course repository integration test, and no task owns the fix**
- where: §2 backend table, §9 A01, §7 ("the only writer that would break is `seed.sql`")
- why: verified — `backend/internal/course/adapters/postgres/repository_integration_test.go` INSERTs into `courses` (lines 158, 183) without `instructor_id`, against a Testcontainers DB running all migrations. After A01's `NOT NULL`, those inserts fail and `make integration` — A01's own gate — goes red. The §7 sentence is false; §2 and A01 never name the file.
- proposed edit: add `repository_integration_test.go (extend: seed a user fixture, supply instructor_id)` to §2 and A01's scope; correct §7.

### Non-blocking

**NB-1 [complexity + ai-slop — 2 lenses] `staleness.ts` is a policy module with one consumer.** §2/§3/D-2c. Two constants imported by `admin-queries.ts` alone; the two existing staleness sites (`router.tsx` 60s, session 0) stay outside the "policy". Fold into named constants in `admin-queries.ts`; promote on a second consumer.

**NB-2 [complexity + ai-slop — 2 lenses] `InstructorRef`'s `kind: 'user'` discriminates nothing.** D-2b. One variant, one call site, and the wire shape is `{id, displayName, role}` anyway. Use a plain local type; add the discriminant when a second ref kind exists.

**NB-3 [complexity + ai-slop — 2 lenses] `CourseFormFields.tsx` shares shapes that share one field.** §2. Course (title/slug/description/owner), module (title), lesson (title+description) overlap only on title; the file is pre-committed conditionals. Extract repetition inside A10 when it appears.

**NB-4 [complexity] `AdminStates.tsx` duplicates `presentation/ui/states.tsx`.** §2. The kit already exports prop-driven `EmptyState`/`ErrorState`/`Skeleton`; the existing pattern is `t()` at the call site. Drop the file; the two views call the kit directly.

**NB-5 [complexity] Deferrable unique position constraints guard a case the write path precludes.** D-3/§6/B11. With a full-tree rewrite deriving positions server-side, a duplicate can only come from an adapter bug; B11 then tests PostgreSQL's documented deferral semantics. Either simplify to plain indexes + the domain invariant, or keep the constraint and name the trigger it insures against.

**NB-6 [edge] Position semantics undecided: array order or client-sent positions?** §1 vs D-3 vs §6. If DTOs carry `position`, duplicate/gapped payloads are undefined and deferred-constraint failures surface at COMMIT (mapped to 500 by naive error handling). State: request DTOs carry no `position`; server derives from array index. (Interlocks with NB-5.)

**NB-7 [complexity] A02 and A08 are layer-label cuts — the shape I06 п.8 warns against.** §9. Both own zero acceptance rows, their evidence environments are strict subsets of their consumers' (A03, A09), and the plan concedes they cannot land independently. Merge A02→A03 and A08→A09; same hours, every task owns rows.

**NB-8 [logic] R01's 401 for anonymous mutations contradicts D-1's check order unless tests send a valid Origin.** D-1 puts same-origin first (matching the live login handler), so an anonymous Origin-less POST answers `403 origin_forbidden`, not 401. Add "with a valid Origin" to the mutating rows' preconditions, or reorder and say so.

**NB-9 [logic] A04 edits `openapi.yaml`/`identity.sql` but `make generate` ownership sits in A05.** §9. Unregenerated output leaves trunk red at `check-generated` between A04 and A05 — against "main carries only verified work". State that A04 runs `make generate` for its own paths.

**NB-10 [edge] `slug_taken` must map from the unique-violation (23505), not a pre-check.** B03 is sequential; two concurrent creates pass any pre-check. One D-3 sentence makes B03's evidence race-safe by construction.

**NB-11 [edge] The course `summary` field is invisible to the plan.** §1/D-6/R19. `summary` is NOT NULL CHECK 1–500 (verified in 00001) and required by the public schema; the plan's field discussions name only title/slug/description. Name the full admin field set and state that DTOs carry the existing DB bounds; add one bound case.

**NB-12 [edge] A lesson id moving between modules in one PUT is undecided.** D-3. Depending on delete-missing ordering, a cross-module move either destroys the id stability D-3 preserves for Phase 5 or trips the deferred unique. Decide (support course-wide, or reject with 400) and add one integration row.

**NB-13 [edge] PUT's relationship to status is unstated.** D-3/D-4. (a) May an archived course's tree be edited? (b) Does the PUT schema exclude `status`? (c) Does a transition bump `updated_at` and thus invalidate outstanding `expectedUpdatedAt`? Three sentences plus one contract row. ((c) interlocks with BLK-2.)

**NB-14 [edge] The backfill's "instructor fixture wins" rests on an unacknowledged id tiebreaker, and it can pick a disabled account.** §6 step 2. All three fixtures share one transactional `now()`; `ORDER BY created_at, id` resolves by id (`…0002` < `…0003`), and there is no `disabled_at IS NULL` filter. Seed reconciliation repairs it, but the plan should name the tiebreak or filter explicitly so B09 doesn't encode an accident.

**NB-15 [ai-slop] "Seven paths" is an operation count.** §2 openapi row. The diff adds 4 paths (7 operations) and extends one. Say so, or an auditor hunts for seven path entries.

**NB-16 [ai-slop] "Pagination's three strings" undercounts.** §5.5. `Pagination` carries five Ukrainian strings (two buttons, page label, empty label, nav aria-label); `Table.emptyMessage`/`Skeleton.label` are optional-with-Ukrainian-default, so the real change is "defaults removed, props required". P14 as written would not catch a surviving aria-label.

**NB-17 [ai-slop] "Total ≈ 19 h" contradicts the plan's own column.** §9. The Est. column sums to 19.5 h. Write the computed number.

**NB-18 [ai-slop] "~450 lines" is a forecast dressed as measurement.** §5.6. `i18n.tsx` is 289 lines today; the projection is invented and does no work the genuine frontend-patterns §8 citation doesn't already do. State "289 today, growing by the admin vocabulary".

### Cross-lens observations

- The densest defect cluster is D-3: five findings (BLK-1/2/3/5, NB-5/6/12/13 adjacent) land on the aggregate-write section. The aggregate PUT itself is the plan's best simplification — but its mechanism paragraphs are under-specified exactly where correctness lives (who bumps the token, who guards foreign ids, what the payload's position/status/slug/instructorId members mean).
- The task-cut findings (BLK-4/11/12, NB-7/9) share one root: ownership of evidence was assigned before checking what each task can actually observe at landing time — the same I06 п.8 failure mode the plan cites as its guiding lesson.
- D-4's open question for the user (instructor publish/archive/restore) is correctly flagged by the plan itself and stands; no lens found it decided elsewhere.

### Verdict

- Lenses: 4 launched / 4 returned. Every lens produced findings; none needed a rerun.
- Findings after dedup: **12 blocking, 18 non-blocking** (three findings were seen independently by two lenses each: BLK-3, BLK-4, BLK-5; likewise NB-1/2/3).
- Blocking findings applied to the plan: 0 — this critique runs read-only by mandate; each carries its proposed edit above and awaits the plan author.
- **Open blocking: 12 → GATE 2 is RED.** The plan is not ready for plan-defense until the D-3 mechanism cluster, the task-ownership contradictions, and the body-size arithmetic are fixed in `plan.md`.


## GATE 2 — plan-defense (sonnet, fresh context) — reconstructed record

The defense agent's write was lost to a concurrent append race (the critique section
overwrote the file); its verdict is reconstructed here from the agent's returned report,
by the orchestrator. Target: plan edition 1.

Spot-checks all confirmed accurate (identity handler privacy, api zone 10r/s burst 20,
migration CHECK style, seed ON CONFLICT clauses, seed-before-seed-identities order,
boundaries_test permitting application→identity/domain, transport failure codes).

| Row | Verdict on edition 1 |
|---|---|
| 1 Hostile input | closed (D-3, B04/B05, R14, R19) |
| 2 Dependency failure | closed, narrowly evidenced — F1: proven for one read endpoint only; no frontend ErrorState row |
| 3 Race and replay | **OPEN** — F2: POST /admin/courses had no duplicate/retry treatment |
| 4 Empty and boundary data | closed — F3: "duplicate ids within payload are rejected" had no acceptance row |
| 5 Rollback | closed (§7 incl. named irreversible cost of goose down) |

Verdict on edition 1: RED (row 3 open). Rows 2/4 notes F1, F3 passed to the edition 2
re-verification round.


## GATE 2 round 2 — plan-critique verification (fable, fresh context)

Date: 2026-09-05. Target: `docs/admin/plan.md` edition 2. Method: adversarial re-verification
with a clean context — every BLK closure opened at its named location and checked for
substance, not phrasing; every code fact the plan asserts spot-checked against the repository;
fresh-eyes sweep of the changed sections (task cut, new acceptance rows, D-3b–g, D-4a–c, D-7,
§10 dispositions).

### Code facts spot-checked — all hold

- `repository_integration_test.go`: the two direct `INSERT INTO courses` without
  `instructor_id` are at **exactly lines 158 and 183** as cited; the failing-migration subtest
  builds `fstest.MapFS` with a synthetic `00003_failure.sql` at line 89; `backend/migrations/`
  holds only 00001/00002, so the collision with the planned `00003_authoring.sql` is real and
  the renumber-to-00004 fix is correct. The concurrent double seed (lines 109–116) is real, so
  §5.2's `LOCK TABLE` reasoning for `DO UPDATE` is sound.
- `seed.sql`: five fixed-UUID demo courses, `ON CONFLICT DO NOTHING`, no lock, no
  `instructor_id` — the "changes three ways" description is accurate, and the FK hazard of
  seeding courses before users is genuine (gate order reversal §5.1 is necessary; the current
  `Makefile` `check` target runs `seed` before `seed-identities` in a 12-step chain, matching
  B18's count).
- `00001_courses.sql`: every D-3a bound equals its CHECK (title 1–200, summary NOT NULL 1–500,
  description 1–20000, slug ≤120 with the exact pattern); `updated_at` has only
  `DEFAULT now()` and no trigger — the BLK-2 premise is correct as restated.
- `seed-identities.sql`: fixtures are …0001 student / …0002 instructor / …0003 admin, inserted
  in one transaction with `created_at` defaulting to a single `now()` — the §6 backfill
  tiebreak (`ORDER BY created_at, id` → instructor wins) is fact, not hope; the reconcile
  clause resets `disabled_at = NULL`, so B13 must set a disabled account inside the test, which
  its precondition permits.
- `presentation/ui/data.tsx` / `states.tsx`: `Pagination` carries exactly five Ukrainian
  strings (Назад, Далі, the page label, "Немає сторінок", the `nav` aria-label
  "Сторінки результатів"); `Table.emptyMessage` defaults to "Ще немає записів" and
  `Skeleton.label` to "Завантаження" — §5.6 and P17 describe the debt precisely.
- `i18n.tsx` is 289 lines; `nginx.conf` has `api` 10r/s burst 20, `login` 10r/m burst 5,
  `client_max_body_size 1m`; identity's `authenticate`/`sameOrigin` are private today;
  `handler_test.go` (course) is a kin-openapi contract harness over a **fake repository** —
  relevant to N2 below.

### Per-finding verification

**BLK-1 — CLOSED** (D-3d, B05/B06/B07). The arithmetic is redone at 2 B/char and is correct:
41,400 + 16,000 + 288,000 + 11,200 = 356,600 B ≈ 348 KiB, 32 % headroom under 524,288 B. The
shrunk bounds (40 modules, 120 lessons, lesson description 1000) are carried consistently into
§6's CHECKs and B05/B06. Nginx untouched claim now true. A residual at >2 B/char is N1 below —
non-blocking because the outcome is B07's defined bounded refusal, not the silent
pass-bounds-fail-cap contradiction of edition 1.

**BLK-2 — CLOSED** (D-3b, B12, R18, R23). The mechanism is now stated, not implied: an
explicit `SET …, updated_at = $clock WHERE id = $1 AND updated_at = $2` on every accepted
write **including children-only changes**, microsecond truncation on the application clock,
RFC 3339 microsecond serialization both directions, zero rows → `409`. The GET→PUT→200 round
trip (B12) and stale→409 (R18, integration) rows exist; R23 closes the NB-13c interlock
(transition retires tokens). Evidence-tier nit recorded as N2.

**BLK-3 — CLOSED** (D-3c, R19). Substantively verified: `DO UPDATE … WHERE
course_modules.course_id = $courseId` makes a foreign-row conflict count as zero affected rows
(PostgreSQL does not update when the conflict-action WHERE is false), the affected-rows-vs-
payload-rows comparison catches the shortfall inside the transaction, and the whole write
rolls back with 400. The lesson guard via membership in this course's modules correctly
permits the supported cross-module move while blocking foreign capture; the INSERT arm cannot
capture by construction. R19 asserts both courses read back unchanged.

**BLK-4 — CLOSED** (§8 ownership, §9 sequencing note). A04 now owns only what A04 registers:
P07 + R01–R07, all on the three reads. The mutation half of the matrix (R08–R11) lands with
A06, which completes the mutation surface, so every owner can produce its evidence at landing
time. Checked row-by-row: no A04 row touches an unregistered path.

**BLK-5 — CLOSED** (D-3g, R28, R29, R30). The false idempotency sentence is retracted and
replaced with the correct model: stable content via client-minted ids, non-idempotent request,
self-conflict on retry, refetch-and-present conflict state whose copy does not assert another
person edited (R28). Create replay is guarded by the slug unique (R29) and double submission
at the source (R30). This also closes plan-defense row 3.

**BLK-6 — CLOSED** (D-3e, B04, R20, R21, D-6). Decided: mutable until ever-published
(`published_at IS NULL`, which the schema CHECKs guarantee and which survives archive —
verified against 00001), then `409 slug_immutable`; collision on legal rename is `slug_taken`
from 23505. Rows cover mutable, immutable and collision; D-6 states no advertised public URL
can 404 through a rename, preserving Q8's semantics.

**BLK-7 — CLOSED** (D-3f, R16). Delete and transition are single guarded statements with
in-transaction zero-row disambiguation; the status endpoint now carries `expectedUpdatedAt`
(also in §1's table), which is what orders transitions against PUT/DELETE. R16 is an
integration race row owned by A03. The `published_at` handling keeps every schema CHECK
satisfiable.

**BLK-8 — CLOSED** (D-4b, P06, R12, R13). Carrier named (PUT `instructorId`), eligibility
decided (non-disabled instructor/admin, else `400 invalid_owner`, checked in the write's
transaction), instructor echo rule kept, and the missing happy path exists: P06 (admin
reassigns → course moves between owners' lists), owned by A05, plus R13 for the three
ineligible-owner shapes.

**BLK-9 — CLOSED** (D-4c, R31, R32). The picker renders and its query mounts only for admin
actors; an instructor issues zero requests (R31 asserts no request is made); a 503 picker
degrades to read-only owner without blocking save (R32). Both are component rows owned by A11.

**BLK-10 — CLOSED** (D-7, B19, A12). The recorded trigger is acknowledged as fired and the
recorded fix (shared `storageState`) lands in this slice: a setup project signs in once per
role, the budget is counted (≤6 per run, ≤3 consecutive, inside burst=5 of the verified
10r/m zone), and B19 measures the actual count across two consecutive runs rather than
asserting it. A12 exists to own it.

**BLK-11 — CLOSED** (P17/P18, A08/A10). Edition 1's P14 is split into the kit-level observable
(P17, owned by A08, `ui.test.tsx`, asserting all five Pagination strings, the table empty text,
the skeleton label and no Cyrillic in the output — matching the verified kit inventory) and the
screen-level observable (P18, owned by A10, which builds the route). The DAG and ownership now
agree.

**BLK-12 — CLOSED** (§2 row, §5.3, §7, A01, B16). The plan names the file, the exact INSERT
lines (verified), all **three** breakages including the migration-number collision edition 1's
critique itself had not caught (verified: the canary is `00003_failure.sql` in a MapFS that
also lacks the new real migration), and the fix is inside A01 whose own gate is
`make integration`. §7 is corrected from "only seed.sql" to "every writer of courses", and the
claim that no third writer exists holds — Phase 3 application code never inserts a course.

### New findings (edition 2 sweep) — none blocking

**N1 [edge, non-blocking] The 2 B/char budget is a Cyrillic worst case, not a legal worst
case.** D-3d. The CHECKs bound characters, not bytes: an all-CJK maximal tree is ~3 B/char
(≈ 517 KiB) and an all-astral one ~4 B/char (≈ 686 KiB) — both pass every field bound and
exceed the 512 KiB cap. Unlike BLK-1 this is not a contradiction: B07 defines the refusal
(400, bounded response) and nginx's 1m still covers the cap, so the failure mode is a safe,
explicit refusal of pathological content. Proposed edit: one D-3d sentence naming the
residual — the bounds guarantee fit only at ≤2 B/char; beyond that the byte cap is the guard.

**N2 [logic, non-blocking] B12's "including when only a lesson title changed" is a
repository-tier observable carrying contract-tier evidence.** The contract harness (verified)
runs the handler over a fake repository, so B12 proves serialization precision and handler
behaviour but can pass even if the PostgreSQL adapter forgets the children-only bump — the
exact BLK-2 mechanism. R18 (integration) catches it only if its first write is children-only,
which is unspecified. Proposed edit: pin R18's first write to a children-only change, or add
"children-only write bumps `updated_at`" to A03's integration list in §2.

**N3 [ai-slop, non-blocking] D-6's "a maximal program adds roughly 8 KB of HTML" is out by
~8×.** With D-3d bounds a maximal program is 40×200 + 120×200 = 32,000 chars ≈ 64 KB in
Cyrillic UTF-8 before markup. The titles-only decision survives — the correct number argues
for it more strongly — but this is the NB-18 genre recurring in a plan burned twice on
arithmetic. Proposed edit: write the computed figure.

**N4 [ai-slop, non-blocking] D-5's "Every cell … proven by a contract test" overstates the
disabled-account row.** Only R03 (GET list) exercises it; the other seven cells ride the
uniform middleware. Defensible via the representative cell, but the sentence should say that
rather than "every cell".

**N5 [edge, non-blocking, minor] `GET /admin/instructors` at 201 eligible accounts is
undefined.** "At most 200, unpaginated" does not say which 200 or whether truncation is
signalled. Fixture-scale data today; one sentence (deterministic order, silent cap, trigger
to paginate) would do.

### §10 dispositions spot-check

All 17 "addressed" dispositions verified at their cited locations (NB-16's five strings and
NB-18's 289 lines verified against the actual files; NB-17's 22.5 h recomputed from the Est.
column: 6×2 + 5×1.5 + 3×1 = 22.5). NB-7's retained-with-substitute stands up: A02/A03 and
A09/A10 are cut on evidence environment (fakes vs Testcontainers, MSW vs rendered route), every
task now owns ≥1 row (checked task-by-task), no row has two owners, and the row count (21 + 19
+ 36 = 76) is accurate. The literal-merge refusal is justified by the 2 h ceiling (A02+A03
would be 3.5 h).

### Verdict

- Blocking findings closed: **12 of 12** — BLK-1 ✓, BLK-2 ✓, BLK-3 ✓, BLK-4 ✓, BLK-5 ✓,
  BLK-6 ✓, BLK-7 ✓, BLK-8 ✓, BLK-9 ✓, BLK-10 ✓, BLK-11 ✓, BLK-12 ✓. Every closure is
  substantive at its named location; none is a rephrase.
- New blocking findings: **0**. New non-blocking: 5 (N1–N5), recommended for a light editorial
  pass or the report's limitations section — none forces rework.
- **GATE 2 round 2: GREEN.** Edition 2 is ready for plan-defense re-verification (rows 2/4
  notes F1/F3 from the reconstructed defense record remain to be re-checked there; F3 appears
  closed by R19's acceptance row and F1 partially by R24/R32, but that verdict belongs to the
  defense gate, not this one).

## GATE 2 round 2 — plan-defense (sonnet, fresh context)

Date: 2026-09-05. Target: `docs/admin/plan.md`, **edition 2**. Context read: `docs/admin/spec.md`
(GATE 1), `.claude/standards.md`, `AGENTS.md`, the edition-1 reconstructed defense record and
GATE 2 round-1 critique above. Method: `plan-defense` skill, five fixed questions, run fresh
against edition 2 with named re-checks of row 3, F1 and F3.

**Spot-checks performed against the repository** (not taken on faith):

- `backend/migrations/00001_courses.sql` — confirmed `updated_at timestamptz NOT NULL DEFAULT now()`
  with no trigger, and the four `CHECK` constraints the plan cites (D-3b, D-3f). Only
  `00001_courses.sql` and `00002_identity.sql` exist; `00003_authoring.sql` is not yet created,
  consistent with this being a pre-implementation plan.
- `backend/internal/course/adapters/postgres/repository_integration_test.go` — confirmed the two
  bare `INSERT INTO courses(...)` at **lines 158 and 183** omit `instructor_id`, and the synthetic
  `"00003_failure.sql"` key exists in the `fstest.MapFS` at line 89. §2/§5.3's fix (renumber to
  `00004_failure.sql`, add the real `00003`, supply `instructor_id`) is grounded in the actual file,
  not invented.
- `backend/platform/architecture/boundaries_test.go` — confirmed the `allowed` predicate is
  `application && strings.HasPrefix(name, "courses/internal/") && (contains "/domain" or "/ports")`,
  which permits `course/application` to import `courses/internal/identity/domain`. D-1 point 3's
  claim ("boundaries_test.go would have permitted this; a choice, not a constraint") holds exactly.
- `backend/internal/identity/adapters/http/handler.go` — confirmed private `func (handler *Handler)
  authenticate` (line 107) and `func (handler *Handler) sameOrigin` (line 124), called at lines
  57/92. D-1 point 2's "private → exported" claim is accurate.
- `infrastructure/nginx/nginx.conf` — confirmed `zone=login` at `10r/m burst=5 nodelay` and
  `zone=api` at `10r/s burst=20 nodelay`, `client_max_body_size 1m`. D-7's login-zone arithmetic and
  §5.10's "unchanged, 2× margin over the 512 KiB cap" both check out against the real file.

All five spot-checked citations held. No fabricated line number or config value found.

### The five questions

| # | Question | Closed by | Verdict |
|---|---|---|---|
| 1 | Hostile input | §4 D-3a (field bounds matching DB `CHECK`s), D-3d (size budget), D-5 (refusal matrix); rows R06/R07 (malformed/unknown id), R13 (ineligible owner), R22 (lifecycle members rejected by schema), B06/B07/B08 (tree/body/summary bounds), R19 (foreign node ids), R36 (client-side field validation before send) | **closed** |
| 2 | Dependency failure | §4 D-4c (picker degrades to read-only, save still possible) + R32 (component); application layer R24 (`List`/`Save` → 503, fake port); adapter layer R27 (MSW codes reach the domain intact) — **but no acceptance row exercises the primary list/tree-read failure surface** | **open — see F1 re-check below** |
| 3 | Race and replay | §4 D-3g (tree-write retry gets its own stale token back → 409, not merged; create replay guarded by unique slug → `409 slug_taken`; buttons disabled in flight); rows R29 (replayed create → one course), R18 (concurrent save refused not merged), R23 (transition retires tokens), R28 (conflict copy does not assert another person edited), R30 (double-submit on create), R16 (delete cannot race publish), B03 (slug uniqueness under concurrency) | **closed — see row-3 re-check below** |
| 4 | Empty and boundary data | B02 (empty list), B05/B06 (max tree, bounds exceeded), B10 (pagination boundary), R19 (foreign-course node ids) all covered — **but the payload-internal-duplicate-id rule stated in D-3c has no acceptance row** | **open — see F3 re-check below** |
| 5 | Rollback | §7 (application rollback via redeploy, migration left in place and additive; named list of every writer of `courses` that would break, corrected from edition 1's BLK-12; explicit "do not `goose down 00003`" with the irreversible cost named; data-rollback for a bad edit explicitly out of scope for 4a with named mitigations; gate-order rollback) | **closed** |

### Row 3 re-check (edition 1 was OPEN here)

Edition 1's F2 — "`POST /admin/courses` had no duplicate/retry treatment" — is fixed: D-3g states
the unique-slug guard and R29 is a dedicated row (`send the same create twice` → one `201`, one
`409 slug_taken`, exactly one course). The tree-write retry case (a lost response replayed) is
handled by the same `expectedUpdatedAt` mechanism already used for genuine concurrent edits (R18),
and D-3g's second bullet explicitly redefines the conflict-state copy so it does not claim "someone
else edited" when the retry was the caller's own successful write — R28 is the component row for
that copy. Double submission at the UI is named as prevented by disabling the button while the
mutation is in flight; R30 is the dedicated row, though **only for the create dialog** — the save
button's `isSubmitting` guard (also asserted in prose, D-3g bullet 4) has no equivalent named row
in the editor's acceptance rows (A11 owns P19/B09/R28/R30/R31/R32/R35/R36, none of which is "press
Save twice quickly"). This is a narrower instance of the same claim-without-row pattern as F1/F3,
but it is not the row-3 finding edition 1 raised (F2, about create), and the mechanism (RHF
`isSubmitting`) is identical to the one R30 already exercises. Verdict: **row 3 closed**; the
save-button double-submit gap is noted but does not reopen the row edition 1 flagged.

### F1 re-check (dependency failure — narrowly evidenced, no frontend ErrorState row)

Still true in edition 2. `§2` names `CourseListView.tsx` as owning "empty **and error** states" and
`§9` A10's task description repeats "empty and error states," and `§3`'s reuse table asserts "the
kit's prop-driven states are the admin screens' states." None of A10's owned rows (P16, P18, B02,
R34) exercises a `dependency_unavailable`/503 response from `GET /admin/courses` and checks that
`ErrorState` renders. The same gap exists for the editor's tree read (`GET /admin/courses/{id}`) —
no row in A11's list (P19, B09, R28, R30, R31, R32, R35, R36) covers that GET failing. R24 proves
the backend degrades to 503 without leaking detail (application layer, fake port); R27 proves the
frontend adapter maps HTTP failure codes to domain outcomes without collapsing them (MSW, adapter
layer); R32 proves one specific, narrower case (the owner-picker's own dependency, not the primary
resource). None of the three reaches "the user looking at `/admin` or the editor sees the
`ErrorState` component when the primary read fails." F1 is **not closed** by edition 2.

### F3 re-check (duplicate-ids-within-payload — no acceptance row)

Still true in edition 2. D-3c states the rule in prose: "Ids are also validated for UUID format,
non-nil value and **uniqueness within the payload** before any statement runs." Searching all 76
acceptance rows (21 happy + 19 boundary + 36 refusal, §8) for a case that sends two modules or two
lessons sharing the same client-minted id in one `PUT` finds none. B06 tests count/length bounds
(41 modules, 121 lessons, an oversized description), not duplicated ids; R19 tests an id that
belongs to a *different* course (the cross-course-theft guard, D-3c's other rule), not a duplicate
within the *same* payload. The rule the plan states is real and specific (checked "before any
statement runs," implying a pre-transaction guard distinct from the `ON CONFLICT` upsert path), but
it is asserted, not proven at landing time — the same pattern GATE 2 round 1 flagged repeatedly
(BLK-4/11/12) applied here to a single validation rule. F3 is **not closed** by edition 2.

### Verdict

closed: 1, 3, 5 — open: 2, 4 (F1: no frontend ErrorState acceptance row for a primary-resource
dependency failure; F3: no acceptance row for duplicate ids within one payload). **3/5 closed, 2
open → GATE 2 round 2 is RED.**

### Verdict update — edition 2.1

Re-read the changed spans of `docs/admin/plan.md`: D-3c (lines ~271-276), D-3g (lines ~353-356),
the §8 acceptance table (R37-R40, row-count line), and §9 (A02's row bumped to 1.75h, new A12,
renumbered A13-A15 with dependencies shifted one step). Verified directly in the file, not taken
on the author's word:

- **R37** (`admin session, GET /api/v1/admin/courses answers 503` → `ErrorState` with a working
  retry that re-issues the request, `CourseListView.test.tsx`) and **R38** (same for
  `GET /admin/courses/{id}`, plus the explicit "no empty form is shown as if the course had no
  content" clause — closing exactly the failure-reads-as-legitimate-empty-state trap) both exist
  as single rows, both owned **only** by the new **A12** ("Degradation and in-flight guards on
  both admin screens," 0.75h, depends on A11, evidence "Rstest component tests with MSW forcing
  503"). This is observable, component-level evidence, not prose — it closes row 2/F1.
- **R39** (a tree whose two modules — then two lessons — share one client-minted id →
  the use case returns the invalid-tree outcome, **repository never called**) exists as a single
  row owned **only** by **A02**, and A02's own row list now includes it (`R24, R25, R39`) alongside
  "payload id validation" added to A02's task description. D-3c now cites R39 by name at the exact
  point it states the rule, and adds the honest line "Edition 2 stated this rule in prose only,
  which is how plan-defense row 4 stayed open" — closing row 4/F3.
- **R40** (editor Save pressed twice quickly → exactly one `PUT`, button disabled between presses)
  exists as a single row, owned only by A12, and D-3g now explains why it and R30 are two separate
  rows ("two buttons on two screens"). This closes the asymmetry noted in the round-2 report,
  though it was not one of the two rows this session held open.
- Ownership is clean: grep confirms each of R37/R38/R39/R40 appears exactly once as a row
  definition and exactly once in a task's "Owns" column (R39's and R40's other appearances are
  narrative cross-references in D-3c/D-3g, not a second ownership claim). The row-count line
  (`21 happy, 19 boundary, 40 refusal = 80`) and the total-hours line (`23.5h`) both recompute
  correctly against the new rows and A02's revised estimate.
- One residual, non-blocking inconsistency: §2's file table still describes `CourseEditorView.tsx`
  as owning only "conflict state," not "error state," even though R38 now requires it to render
  `ErrorState` on a failed read. The acceptance row is what plan-defense's own skill treats as
  authoritative closure, and R38 names the file, the behavior and the owner unambiguously, so this
  is a documentation nit for the author to tidy, not a reopened hole.

Both previously-open rows are now closed by genuine acceptance evidence with a single owner each.
No new rows were required for the previously-closed rows (1, 3, 5), which edition 2.1 left
otherwise untouched.

**Final verdict, edition 2.1: closed 1, 2, 3, 4, 5 — 5/5 closed. GATE 2 round 2 is GREEN.**


## GATE 2 — constitution (inline, orchestrator, per .claude/standards.md §Конституція)

Checked plan edition 2.1 against AGENTS.md architecture and TypeScript rules: domain
stdlib-only (authoring domain proven with fakes); application = domain+ports; generated
sqlc/Orval confined to adapters with explicit mapping (`admin-course-schema.ts`);
`AuthoringRepository` port backed by a real external dependency; composition request-scoped,
no global mutable clients; no fetch in presentation/routes; URL state via search params;
no forbidden dependencies (RHF enters on the recorded §24.6 trigger; TanStack Table
deliberately not added — the kit's `Table` is the consumer); Nginx media rule untouched;
gate reorder updates Makefile, ci.yml, README and the manifest together (A01); explicit
business error codes. Violations: 0. **Constitution: PASS.**

## GATE 2 final verdict

Round 2: plan-critique verification GREEN (12/12 closed, 0 new blocking) · plan-defense
GREEN (5/5 after edition 2.1) · constitution PASS. **GATE 2 GREEN on plan edition 2.1.**
Verdicts by: fable fresh-context verifier, sonnet fresh-context examiner, orchestrator
(constitution only). Plan author: delegated opus.
