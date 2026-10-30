# /teach lesson-document editor — spec (ЩО / ЧОМУ)

**Phase:** teach-editor (owner-initiated, outside §24). **Date:** 2026-10-30.
**Threshold:** ABOVE spec-flow threshold — cross-cutting, touches backend (new `/teach` document
endpoints + `RegisterTeach`) and frontend (new `/teach` editor surface). Owner chose Option B
(backend `/teach` alias) over the frontend-only Option A; see `_wip-recon.md`.
**Review:** existing_session, reviewer `courses-96 [557618]` (mandatory; auth-touching → money/auth
scrutiny).

## Проблема
Instructors who own a course cannot edit their own lessons' documents without administrator access.
The lesson-document editor (reader + Tiptap editor) exists only under `/admin`, whose route guard
(`canAccessAdministration`) admits only admins. The **backend already authorizes an instructor-owner**
on the document endpoints — the handler admits instructors and the application scopes every op to
admin-OR-course-owner (`requireEditor → CanBeEditedBy → CanManage(OwnerID)`) — but there is no
instructor-facing API path or UI surface, so the capability is unreachable for a non-admin owner
(`docs/lesson-builder/report.md` §Lim1).

## Користувацькі сценарії
1. **Owner edits own lesson.** An instructor opens a course they own under `/teach`, navigates to one
   of its lessons, edits the lesson document (sections, rich text, images), saves a draft, and
   publishes — without administrator rights.
2. **Non-owner refused.** The same instructor cannot read or edit the document of a lesson in a course
   they do not own: the teach endpoints answer `course_forbidden` (403), independently per endpoint.
3. **Admin unchanged.** An administrator keeps editing any lesson's document via `/admin`, byte-for-byte
   unchanged.
4. **Learner unchanged.** Students keep the read-only published path (`/learn/lessons/{id}/document`);
   no editor is exposed to them.

## Acceptance criteria
- **AC1** New teach-reachable endpoints `GET/PUT/POST /api/v1/teach/lessons/{lessonId}/document[/publish]`
  serve draft-get / draft-save / publish to an authenticated instructor, routed through the **same
  owner-scoped application path** (`Documents` + `requireEditor`). No new authorization logic.
- **AC2** Each teach endpoint **independently** enforces owner-of-THIS-lesson (not merely
  role=instructor): a non-owner instructor → `course_forbidden`; a student → `forbidden`.
- **AC3** Mounting the teach endpoints (`RegisterTeach`) widens **nothing** beyond exactly those three
  document ops — no sibling path under the new prefix becomes reachable, and the `/admin` and `/learn`
  document endpoints are byte-unchanged.
- **AC4** A `/teach` UI surface lets an owner-instructor open and edit their own lessons' documents,
  reusing the existing `LessonDocumentEditor`; there is no navigation or data path to another owner's
  content.
- **AC5** OpenAPI regenerates deterministically (generated-authoritative contract); the new endpoints
  appear in the contract; frontend reads them through Orval.
- **AC6** Full ordered `make check` green including e2e (auth + shared-state change): an owner-instructor
  e2e edits+publishes their own lesson; a non-owner instructor is refused.
- **AC7 (CSRF/sameOrigin parity — reviewer must-check).** The teach `RegisterTeach` mount keeps
  `requireSameOrigin` on exactly the mutations — `PUT …/document` (saveDraft), `POST …/document/publish`,
  and `POST …/images` — and leaves `GET` without it, matching the admin mount and the orders/discussion
  `RegisterTeach` precedent. Dropping it would open CSRF on /teach mutations that /admin does not have.
  Parity, not new logic, but easy to miss when copying a mount.

## Out of scope
- Creating / deleting / reordering lessons or modules from `/teach` (document editing only).
- Second-locale editing UI (single-locale, as today; `report.md` §Lim2).
- Image orphaned-bytes sweeper (closed: LB09a + N1).
- Multicurrency; any change to the `/admin` editor surface itself.

## GATE 1 — resolved (owner, 2026-10-30)
1. **Ownership model:** course owner (`CanManage` = admin OR `course.instructor`); no co-instructor /
   shared-ownership concept in v1. Stated as assumption, owner did not object.
2. **Publish from /teach: INCLUDED.** Owner-instructor saves draft AND publishes — full editor cycle
   (3 document ops: get/save/publish). Domain already permits owner publish.
3. **Images from /teach: INCLUDED.** Add `POST /api/v1/teach/lessons/{id}/images` (image `CanEditLesson`
   already admits owner); the teach editor is fully functional with image sections. +1 teach endpoint.
4. **Navigation entry: NEW build.** `teach.courses.$courseId.tsx` today shows roster + moderation +
   comments, no lesson list; a teach course→lessons→editor entry is new, not reuse.

GATE 1 verdict: shared understanding reached with the owner inline (grilling is user-facing; run with
the owner, not a subagent). Scope confirmed; proceed to plan.

## Open items carried to plan
- **Current /admin-editor reachability for a non-admin instructor** must be pinned in plan/reuse analysis
  (both `/admin` and `/teach` guard on `canAccessAdministration` = instructor+admin; report §Lim1 implied
  admin-only). Option B stands regardless on product grounds.
- **AC7 CSRF/sameOrigin parity** is a plan constraint on the mount.
