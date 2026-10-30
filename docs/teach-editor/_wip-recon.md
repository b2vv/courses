# /teach lesson-document editor — recon (pre-spec)

**Date:** 2026-10-30. **Driver:** AGENDA #2 (owner sequence 1→3→N1→2). Source: `docs/lesson-builder/report.md` §Lim1.

## Recon finding (code-verified) — backend already authorizes instructor-owners
- Document endpoints (`course/adapters/http/document_handler.go`): `GET/PUT/POST /api/v1/admin/lessons/{id}/document[/publish]`, mounted via `RegisterAdmin(authenticate, requireSameOrigin)` — **no `RequireAdministrator`** (cf. `cmd/api/main.go:222`).
- Handler guard `adminActor` (`admin_handler.go:365`) is **misnamed**: it builds the real role and calls `CanCreate()` → admits `ActorInstructor || ActorAdmin`, rejects students. So an **instructor** already passes the handler.
- Application `Documents` (`application/document.go`): every write/read-draft op calls `requireEditor → facts.CanBeEditedBy(actor)` = `actor.CanManage(OwnerID) == nil` (`domain/document_access.go:20`, `authoring.go:135`) = **admin OR course-owner**; a non-owner instructor → `ErrCourseForbidden`.
- **Conclusion:** an instructor-owner calling the existing endpoints already succeeds for their own lessons and is refused (course_forbidden) for others. Per-endpoint owner enforcement already exists (satisfies reviewer's PR-#91-style concern). **The gap is frontend only.**

## Frontend gap
- Editor route `admin.courses.$courseId_.lessons.$lessonId.tsx` sits under `/admin` layout whose `beforeLoad` guard `canAccessAdministration(user)` → `notFound()` for non-admins. Uses `admin.detail(...)` + `admin.documents` composition root (calls the /admin document endpoints).
- `/teach` surface already exists (`routes/teach.tsx`, `teach.courses.$courseId.tsx`, `teach.learners.tsx`; `modules/teach/ui/*`). Orders has a `RegisterTeach` backend precedent (`cmd/api/main.go:218`).

## Fork (owner/architecture decision — drives scope AND process)
- **Option A — frontend-only.** New `/teach` editor route reusing `LessonDocumentEditor`, calling the existing (already instructor-authorized) `/admin/lessons/{id}/document*` endpoints via a teach composition root. No backend change. **Below spec-flow threshold** → inline plan + targeted verification. Smallest. Smell: an instructor UI depends on `/admin/*` paths.
- **Option B — teach backend path alias.** Add `/api/v1/teach/lessons/{id}/document*` + `RegisterTeach` (mirror orders), teach UI calls those; handlers pass the real actor into the same owner-scoped application path. **Cross-layer (backend+frontend) → above threshold → spec-flow.** Cleaner API separation, matches precedent; reviewer's "don't widen siblings" concern applies to the new mount. More code for functionally-already-authorized behavior.

## GATE-1 findings (2026-10-30)
- Owner chose **Option B** (backend `/teach` alias). spec.md written.
- `canAccessAdministration` (`identity.ts:79`) = `instructor || admin`; **both** `/admin` (admin.tsx:10)
  and `/teach` (teach.tsx:10) guard on it. So an instructor may already reach the `/admin` route tree —
  the exact current reachability of the editor sub-route for a non-admin instructor must be pinned at
  plan/reuse analysis (report §Lim1 implied admin-only; the guard says instructor+admin). Option B stands
  regardless on product grounds (instructors should not operate in the admin area).
- `teach.courses.$courseId.tsx` shows roster + moderation + comments, **no lesson list/outline** → a
  teach navigation entry (course → lessons → editor) is **new build**, not reuse.

## Correction to earlier framing
Earlier this session I told the owner/peer the feature "touches backend + frontend". Recon shows the backend already authorizes instructor-owners; it is cross-layer only if we choose Option B. Surfaced to owner before writing spec.md (no wasted ceremony yet).
