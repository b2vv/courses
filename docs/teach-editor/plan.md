# /teach lesson-document editor — plan (ЯК)

**Phase:** teach-editor (owner-initiated, outside §24). **Date:** 2026-10-30.
**Threshold:** ABOVE (`.claude/standards.md` §Поріг) — cross-cutting, touches more than one layer of
both backend (new `/teach` document + image endpoints, two `RegisterTeach` mounts, one owner-scoped
course-tree read) and frontend (new `/teach` editor surface + navigation). spec-flow is the correct
vehicle; stated out loud.
**Branch:** `feat/teach-editor`, ff-merge into `main` (`.claude/standards.md` §Гілки й PR).
**Artifacts home:** `docs/teach-editor/` (spec.md, this plan, report.md); tasks in `work/tasks.md`.
**Review:** existing_session, Courses 2. GATE 2 reviewer-of-record = a FRESH Courses 2 session
(the prior `courses-96` hit its session-guard); auth-touching → money/auth scrutiny.
**Inputs:** `docs/teach-editor/spec.md` (AC1–AC7, GATE 1 closed), `_wip-recon.md`.

---

## 1. Approach (Option B, owner-chosen)

Add an owner-scoped `/api/v1/teach/...` path alias for the lesson-document editing capability that
the backend **already authorizes** for instructor-owners, give it an instructor-facing `/teach` UI,
and widen nothing else. No new authorization logic on either layer — teach handlers pass the real
actor into the **same** owner-scoped application paths the admin handlers already use.

Why Option B over Option A (frontend-only, calling `/admin/*`): product boundary — an instructor UI
must not depend on `/admin/*` paths, and the owner chose B. Cost is accepted: more code for behavior
that is functionally already authorized.

## 2. Pinned facts (code-verified this session — the plan rests on these)

- **P1 — backend already authorizes an instructor-owner on the document ops.** Document handler guard
  `adminActor` (`course/adapters/http/admin_handler.go:365`, misnamed) builds the real role and calls
  `CanCreate()` → admits `instructor || admin`, rejects students. Application `Documents`
  (`course/application/document.go`) runs every op through `requireEditor → facts.CanBeEditedBy(actor)
  = actor.CanManage(OwnerID)==nil` (`course/domain/document_access.go:20`, `authoring.go:135`) =
  **admin OR course-owner**. A non-owner instructor → `ErrCourseForbidden`. No new auth logic needed.
- **P2 — image upload already authorizes an instructor-owner.** `image/domain/access.go:11-12`
  `CanEditLesson(ownerID)` allows `actor.Administrator || (actor.UserID != "" && actor.UserID ==
  ownerID)`; `ownerID` is the lesson's course owner (`image/application/upload.go:31`
  `LoadLessonOwner`). An instructor-owner (Administrator=false, UserID==OwnerID) passes. The handler's
  `actor.Administrator = role=="admin"` (`image/adapters/http/handler.go:61`) does not block owners,
  because the match is by UserID.
- **P3 — admin mount pattern to mirror.** `document_handler.go:26-29 RegisterAdmin(mux, authenticate,
  requireSameOrigin)` mounts GET (no sameOrigin), PUT + POST/publish (sameOrigin). Image:
  `image/adapters/http/handler.go:38-39 RegisterAdmin` mounts `POST .../images` (sameOrigin). Orders
  and discussion already ship a `RegisterTeach(mux, authenticate, requireSameOrigin)` precedent
  (`cmd/api/main.go:218,225`).
- **P4 — reachability correction (spec §Lim1 was wrong).** Both `/admin` (`routes/admin.tsx:10`) and
  `/teach` (`routes/teach.tsx:10`) guard on the SAME `canAccessAdministration`, which admits
  `instructor || admin` (`identity.ts:79`, test `identity.test.ts:57`). The admin editor sub-route
  (`admin.courses.$courseId_.lessons.$lessonId.tsx`) adds only entity-id validation, no role guard.
  So a non-admin instructor can **already reach** the `/admin` editor route today. The gap is therefore
  **not** strict reachability — it is (a) no teach-side navigation to the editor and (b) the product
  boundary of instructors operating in the admin area. Option B stands on product grounds (owner).
- **P5 — teach surface has no course-tree / lesson-list source.** `teach-queries.ts` exposes only
  `courseComments` + `moderationQueue`; `teach.courses.$courseId.tsx` loads comments + moderation +
  enrollment roster. There is no owner-scoped authoring course-tree (with unpublished lessons) on the
  teach side. Navigation to the editor is a genuinely new build (owner accepted "NEW" at GATE 1).

## 3. Reuse-first analysis (`.claude/standards.md` §Reuse index + code graph)

### Backend — reuse

| Concern | Reused as-is | Where |
|---|---|---|
| Document get/save/publish handlers | `getDraft` / `saveDraft` / `publish` methods | `course/adapters/http/document_handler.go:46-137` |
| Owner-scoped document authorization | `Documents` + `requireEditor` + `CanBeEditedBy` | `course/application/document.go`, `course/domain/document_access.go` |
| Instructor-admitting handler guard | `adminActor` | `course/adapters/http/admin_handler.go:365` |
| Image upload handler + owner auth | `upload` + `CanEditLesson` | `image/adapters/http/handler.go`, `image/domain/access.go` |
| HTTP server, error envelope, request id, sameOrigin middleware | reused | `backend/platform/httpserver/`, identity `RequireSameOrigin` |
| OpenAPI route-binding registration | extend existing contract files | `course/adapters/http/openapi_contract.go`, `image/adapters/http/openapi_contract.go` |

### Backend — new

| New | Why existing did not fit |
|---|---|
| `DocumentHandler.RegisterTeach(mux, authenticate, requireSameOrigin)` mounting GET/PUT/POST-publish under `/api/v1/teach/lessons/{lessonId}/document[/publish]` | `RegisterAdmin` mounts the `/admin` prefix; a second mount is the minimal way to expose the SAME handlers under `/teach` without touching `/admin`. No handler-body change. |
| `Handler.RegisterTeach(...)` (image) mounting `POST /api/v1/teach/lessons/{id}/images` | same — reuse `upload`, new prefix only. |
| One owner-scoped course-tree read on the teach side (**§6 decision**) to feed lesson-list navigation + lesson meta | P5: no teach course-tree exists; `admin.detail` calls `/admin/courses/{id}` and reusing it reintroduces the `/admin` dependency Option B exists to remove. |
| Wiring in `cmd/api/main.go` (two `RegisterTeach` calls after the existing `RegisterAdmin` lines; one course-tree teach mount) | composition root is the only place mounts are wired. |

### Frontend — reuse

| Concern | Reused as-is | Where |
|---|---|---|
| The editor component (Tiptap, lazy-loaded) | `LessonDocumentEditor` + `LessonDocumentEditorLoading` | `modules/document/ui/` |
| Document domain types + error codes | `LessonDocument`, `DocumentAuthoringErrorCode`, `ImageUploadErrorCode`, `UploadedImage` | `modules/document/domain/document.ts` |
| Repository port | `DocumentAuthoringRepository` | `modules/document/ports/` |
| Query/mutation factory shape | `createAdminDocumentQueries` (to be generalized — see new) | `modules/document/application/admin-document-queries.ts` |
| HTTP transport, request headers, SSR cookie forwarding | reused | `adapters/api/transport.ts`, `request-headers.ts` |
| Orval-generated client | regenerated to include `/teach/*` | generated adapters |

### Frontend — new

| New | Why existing did not fit |
|---|---|
| Teach `DocumentAuthoringRepository` adapter calling `/api/v1/teach/...` via Orval | the admin repository targets `/admin/*` endpoints; the teach UI must hit `/teach/*` (Option B). |
| Generalize the one factory by (repository + query-key prefix) so `admin`/`teach` share it — **decided: parameterize, do not copy** (NB1, 2026-10-30) | the current factory hard-codes the `['admin', …]` query key; a copy would drift. Parameterizing keeps the admin key byte-identical (prefix `'admin'`) while adding a `'teach'` prefix — DRY without regressing admin (constitution §7). |
| Teach composition root exposing `documents` + the teach course-tree query | request-scoped wiring lives in `app/composition/`. |
| Route `teach.courses.$courseId_.lessons.$lessonId.tsx` (editor) mirroring the admin editor route | new file; under the `/teach` layout guard, lazy-loads `LessonDocumentEditor`, passes the teach `documents`. |
| Lesson-list entry on `teach.courses.$courseId.tsx` (course → lessons → editor) | P5: no lesson list on the teach course view today. |

## 4. Side-effects

- **New routable surface.** Three document ops + one image upload + one course-tree read become
  reachable under `/teach/*`. AC3 requires that **no sibling path** under the new prefix becomes
  reachable and that `/admin` + `/learn` document endpoints are byte-unchanged. Enforced by: mounting
  ONLY the named routes; a contract-census test asserting the teach prefix exposes exactly these paths;
  re-running the existing admin/learn contract tests unchanged.
- **CSRF / sameOrigin (AC7).** Each `RegisterTeach` must keep `requireSameOrigin` on exactly the
  mutations (`PUT …/document`, `POST …/document/publish`, `POST …/images`) and leave the GETs without
  it — mirroring the admin mounts and the orders/discussion `RegisterTeach` precedent. A dropped
  `requireSameOrigin` opens CSRF on teach mutations. Guarded by an explicit mount test asserting the
  middleware chain per route.
- **OpenAPI contract drift (AC5).** Adding routes changes the generated contract; `make check-generated`
  fails on sqlc/Orval drift, so the regenerated OpenAPI + Orval client must be committed in the same
  slice. Generated-authoritative: hand-edit the contract source, regenerate, never the reverse.
- **Shared state + e2e (AGENTS.md "lessons from closing phases").** This is an auth + shared-state
  change; it is verified with the FULL e2e set, not a targeted one. New e2e: owner-instructor
  edits+publishes own lesson; non-owner instructor refused.
- **Image bytes / nginx.** No change to the media/image byte path or nginx `auth_request`; the teach
  image upload reuses the same `upload` application and storage. No new byte-reachability surface.
- **Concurrency & re-entry (editing feature — concurrent save, double-publish, two tabs on one lesson).**
  The defense already exists in the reused-as-is `saveDraft` / `publish`: optimistic concurrency on
  `expectedUpdatedAt`, so a stale write returns `ErrDocumentConflict` → HTTP 409 `document_conflict`
  (`course/adapters/http/document_handler.go:38,85-93,117-122,185-186`). The teach mount reuses these
  handler bodies unchanged, so the teach editor inherits the same 409 behavior; the UI reuses the admin
  editor's existing conflict handling. No new concurrency logic. Image upload is idempotent-safe (each
  upload is a fresh image id; LB09a/N1 already compensate orphaned bytes on insert failure).

## 5. Migration plan

- **Database:** none. No schema change — no new table, column, or migration. Ownership and documents
  already exist; this phase adds transport + UI only.
- **Data backfill:** none.
- **Generated artifacts:** regenerate OpenAPI + Orval (`make check-generated` gate). Deterministic.
- **Deploy ordering:** backend endpoints ship before/with the frontend that calls them (same branch,
  same merge); the frontend adapter fails classified (not crash) if an endpoint 404s, because transport
  classifies failures. No separate rollout step — test mode, no prod server (owner decision 2026-10-03).

## 6. Design decision (resolved) — teach lesson-list / course-tree data source

**Decision (owner, 2026-10-30): option (a).** Verify by: the lesson-list on the teach course view and
the editor title block are fed by a new owner-scoped `GET /api/v1/teach/courses/{courseId}`, with no
`/admin/*` call from the teach UI.

AC4 ("owner opens and edits their own lessons") requires listing a course's lessons on the teach side,
plus lesson meta (title/module) for the editor title block (the admin route resolves this from
`admin.detail`'s course tree). P5: no such source exists on teach. Options considered:

- **(a) Add an owner-scoped teach course-tree read** `GET /api/v1/teach/courses/{courseId}` via
  `courseHandler.RegisterTeach`, returning the authoring tree (incl. unpublished lessons) scoped to the
  course owner (reusing the course module's existing owner-scoped read + `adminActor`-style guard; the
  implementing task must verify the course detail read is owner-scoped, not merely role-gated).
  Consistent with Option B's "no `/admin` dependency". One new read endpoint beyond the spec's AC1–AC3
  document ops; owner already accepted navigation as a new build at GATE 1.
- **(b) Reuse `/admin/courses/{id}` from the teach UI.** No new endpoint, but reintroduces the exact
  `/admin/*` dependency Option B was chosen to remove. **Rejected (owner, 2026-10-30)** — product boundary
  wins over minimal surface.

This is the one genuinely-new backend read beyond the mount aliases; it adds `courseHandler.RegisterTeach`
and a teach course-tree query to the frontend §3 "new" lists.

## 7. Constitution check (`.claude/standards.md` §Конституція → AGENTS.md architecture + TS rules)

- **Dependency direction (backend).** New code is adapters (handlers/mounts) pointing inward to the
  existing application `Documents` / image `Uploader`. No domain/application change. Machine-checked by
  `backend/platform/architecture/boundaries_test.go`.
- **Dependency direction (frontend).** New repository adapter in `modules/document/adapters` (or a teach
  adapter), composition in `app/composition/`, route in `routes/`. No fetch in presentation/routes; the
  editor stays presentation, data comes through the composition root. Machine-checked by
  `frontend/scripts/check-architecture.mjs`.
- **No new cross-module import.** The image endpoint stays in the image module; the document endpoints
  in the course module. Teach UI composes repositories; it does not import another module's internals.
- **TypeScript strict.** No `any`, no unsafe assertions; external data (endpoint responses) validated
  through the existing Orval/types + the repository's result typing. Reuse the discriminated-union
  result (`{ ok, value | error }`) the admin queries already use.
- **Generated models stay in adapters** and map to domain types (sqlc/Orval). AC5.
- **Rollback (constitution: a closed alternative is a dated artefact).** Rollback = revert the feature
  branch merge; no data migration to unwind, so revert is clean. Recorded here, dated 2026-10-30.

## 8. Rollback

Single feature branch `feat/teach-editor`; revert the merge commit. No DB migration, no backfill, no
persisted-state change → revert restores the prior tree exactly. The `/admin` and `/learn` surfaces are
byte-unchanged by construction (AC3), so a revert cannot regress them.

## 9. Acceptance-spec table (observable behavior → verified-by; maps to spec AC1–AC7)

Derived inline from spec.md user scenarios + AC (acceptance-spec role; the independent gates in §10 are
delegated, this table is not a gate).

| # | Scenario | Precondition | User action | Expected result | Verified by | AC |
|---|---|---|---|---|---|---|
| T1 | Owner gets own draft | Instructor owns course C, lesson L in C | `GET /api/v1/teach/lessons/L/document` | 200, draft document body | contract test + e2e | AC1 |
| T2 | Owner saves draft | as T1 | `PUT /api/v1/teach/lessons/L/document` (sameOrigin) | 200, updatedAt advanced | contract test + e2e | AC1, AC7 |
| T3 | Owner publishes | as T1 | `POST /api/v1/teach/lessons/L/document/publish` (sameOrigin) | 200/204; published path reflects it | contract test + e2e | AC1, AC7 |
| T4 | Owner uploads image | as T1 | `POST /api/v1/teach/lessons/L/images` (sameOrigin) | 201 image created, owner-scoped | contract test + e2e | AC1(image), AC7 |
| T5 | Non-owner instructor refused (all endpoints) | Instructor does NOT own course of lesson L | each teach op on L (GET/PUT/POST-publish document + POST images) | `forbidden` (403) on every endpoint — `requireEditor`→`ErrForbidden` (document) and `CanEditLesson`→`ErrForbidden` (image); NOT `course_forbidden` (code-verified, no handler change) | contract test (each endpoint) + e2e | AC2 |
| T6 | Student refused | role=student | any teach op | `forbidden` (403) | contract test | AC2 |
| T7 | No sibling widening | teach mount active | request an UNKNOWN (non-mounted) sibling path under `/teach/lessons/{id}/…` (not DELETE on a mounted path — Go ServeMux returns 405 there, not 404) | 404 (not reachable) | contract-census test | AC3 |
| T8 | Admin + learn unchanged | — | existing admin/learn document suites | byte-for-byte pass, unchanged | existing contract tests re-run | AC3 |
| T9 | Owner edits via UI | owner-instructor signed in | open `/teach` course → lesson → edit → save → publish | editor works; no path to another owner's content | e2e | AC4 |
| T10 | sameOrigin parity | teach mount active | inspect middleware chain per route | sameOrigin on PUT/POST only, absent on GET | mount test | AC7 |
| T11 | Contract regenerates | endpoints added | `make check-generated` | no drift; teach routes in OpenAPI; Orval client updated | gate | AC5 |
| T12 | Full ordered gate green | all slices merged | `make check` (through e2e) | green incl. new e2e | gate | AC6 |
| T13 | Empty states render | owner with a course of 0 lessons, or a lesson with no draft | open teach course / lesson | empty lesson list and empty-draft editor render benignly (no crash/invention), reusing admin editor empty-state handling | component test | AC4 |

## 10. Verification (machine gates, `.claude/standards.md` §Машинні гейти)

Ordered: `format-check → lint → typecheck → test → integration → check-generated → build → up →
seed-identities → seed → smoke → e2e`. `lint` includes the frontend architecture check;
`check-generated` catches OpenAPI/Orval drift. Measured results → `docs/teach-editor/report.md`.
Per-slice: `npm run format` on touched frontend (gotcha: Codex/format skip trips format-check first).

## 11. GATE 2 — critique & defense — **PASS** (2026-10-30)

**Verdict: PASS.** Reviewer: independent fresh-context subagent (review_mode `independent_subagent`,
owner-authorized 2026-10-30; `gate-reviewer` agent avoided — known to trip a safeguard on
critique/defense). Round 1 = RED on two closable items (B1 per-endpoint refusal codes; defense Q3
race/re-entry) — both verified against the real handlers and closed by plan/spec edits (no code
change), committed `7164f31`. Round 2 (delta) = PASS: zero blocking, zero unclosed defense rows, all
five pinned facts (P1–P5) confirmed, both auth must-checks (AC7, AC3) sound. §6 option (a) reuse
confirmed owner-scoped (`authoring.go:78` `CanManage(InstructorID)`). Cleared to tasks.

**Post-gate correction (2026-10-30, during TE1 implementation).** B1's resolution was itself partly
wrong and is superseded: the GATE 2 reviewer (and the earlier B1 check) traced `CanManage →
ErrCourseForbidden`, but `CanBeEditedBy` collapses that to a boolean and `requireEditor`
(`application/document.go:92`) returns `domain.ErrForbidden` for every non-editor — so the document
editor endpoints emit **`forbidden`**, not `course_forbidden`, for a non-owner instructor (proven by the
existing passing test `TestDocumentNonEditor`, which Codex surfaced when refusing to assert a false code).
AC2/T5 corrected: all endpoints → `forbidden`. This is an error-code-string fact only; the
owner-enforcement security property and the plan's design/reuse/side-effects/rollback are unchanged, so
it is handled as a dated correction (like B1) rather than a re-gate. The per-slice independent review
before merge verifies the actual test assertions against the handlers.

Three checks, all mandatory (two delegated to a FRESH subagent, one inline):

- **plan-critique** on this plan — red if any blocking item open or a lens didn't return.
- **plan-defense** on this plan — red if any of the 5 questions has an unclosed row.
- **Constitution** (inline, §7 above) — red on any violation.

Verdict rendered by a fresh-context reviewer, not the author. Red gate stops the move to tasks. Reviewer
of record: **`courses-96 [557618]`** (owner-authorized, 2026-10-30). Package = this plan + spec.md +
`_wip-recon.md`, reviewed read-only via `git show` at `613b50d`+ on the shared tree. §6 is resolved (a).
Code begins only after GATE 2 PASS, on branch `feat/teach-editor`.
