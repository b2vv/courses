# Phase enrollment — delivery report

Course membership as the authoritative access fact: one row per `(user, course)` whose **state**
(`none → active → blocked → revoked`) is the power to consume content and communicate. Free content
now requires membership too. Branch `feat/enrollment`. Spec `docs/enrollment/spec.md` (G1–G11,
En01–En15); plan `docs/enrollment/plan.md` (+ dated amendments). Roles this phase: plan/gates/commits —
Courses 1; code — Codex; review — Courses 2.

## 1. What was done
- **L1** (`35cf77b`): domain grant→membership (source `+self`, state `+blocked`, transitions
  active↔blocked, *→revoked, invariants) + migration `00015` — CHECK widened, the partial
  active-only unique index DROPped for a full `UNIQUE(user_id,course_id)` (exactly one row per pair),
  `participant_reports` table (CHECK reporter≠target, partial unique dedup WHERE open), backfill of
  self-active for existing free-course users with progress∪discussion activity, seed self-active.
- **L2a/b/c** (`a4cffc4`,`c5df464`,`8c991ec`): orders application/adapters — CanStartOrder as
  UPSERT-transition (purchase on an existing row: no 2nd, no unblock, no revive-revoked→refund),
  Enroll(self, email-first)/Block/Unblock/Revoke/MembershipState/Roster(+progress), compare-and-set
  `ErrMembershipConflict`.
- **L3a/b** (`fc4ef4c`,`a696380`): `contracts/access` membership enum + `CanConsume`(active|blocked)
  and `CanCommunicate`(active) pure predicates; all 7 readers compute membership state in their own SQL.
- **L4a/b** (`7572da2`,`cfdc555`): content gate (CanConsume) in playback×2/transcript/course; comms gate
  (CanCommunicate) in discussion publish/report/review with distinct codes `enrollment_required`
  (non-member) / `enrollment_blocked` (blocked), email-first.
- **L5** (`5347545`): person-reports in discussion — domain (self-report rejected, dedup), application
  (4 ops, email-first, authz), membership read via discussion's OWN SQL over courses+course_grants
  (ADR-013, no orders import), keyset admin queue (no email), idempotent resolve.
- **L6a** (`c51db05`): orders HTTP — enroll/membership/roster + block/unblock/revoke; `membership_conflict`→409.
- **L6b** (`789f004`): discussion report HTTP (report-lecturer/report-student + admin queue/resolve,
  RegisterAdmin with administrator() middleware) **+ an L5 contract fix** (persisted-report id RETURNED).
- **L6c** (`61e4070`): `GET /api/v1/learn/memberships` — the «Моє навчання» list source (plan amendment).
- **L7a/b/c1/c2** (`92f618f`,`dbfc813`,`762161c`,`e92f8ce`): frontend — enroll CTA / «Ви учасник» / «Моє
  навчання» with blocked/revoked states and 409→refresh; instructor roster + block/unblock; report
  dialogs; admin participant-report queue + resolve + revoke.
- **L7d** (`2d21b99`, revoke-gate `b803007`): named comms-gate errors surfaced in the UI (discussion +
  review), the active-admin-nav contrast fix, and the report false-success fix.
- **L8** (`42cebce`): e2e `enrollment.spec.ts` + axe; existing specs updated for the deliberate `/learn`
  behavior change; `recovery.spec` scoped off the absolute published-course count.

## 2. Row-by-row acceptance — En01–En15
| En | Scenario | Proven |
|---|---|---|
| En01 | free «Записатися»→active(self); «ви учасник»; in «Моє навчання» | e2e En01 + L6c endpoint |
| En02 | paid purchase→active(purchase); upgrade self→purchase same row | integration (orders) |
| En03 | content to active\|blocked; non-member denied; free non-member — content closed, outline public | e2e En03 (outline 200, content 403) |
| En04 | admin revoke → content cut incl. media-bytes; row remains | e2e En04/En10 (nginx bytes 403) + En04/En15 |
| En05 | blocked: content yes (incl. bytes), communication `enrollment_blocked`; other courses intact | e2e En05/En08 (rendered message) |
| En06 | instructor block/unblock own; foreign refused; admin anywhere | integration + e2e roster |
| En07 | student→lecturer and lecturer→student reports in admin queue; self/dup refused | e2e En07 + L5 integration |
| En08 | roster: name/%/state/date; pagination; no email; 0 lessons→0%, empty state | e2e En05/En08 |
| En09 | archive preserves access; enroll on archived/unpublished refused (named guard) | integration |
| En10 | grandfather=active membership; preview lesson without enrolment | e2e En10 (preview bytes) + integration |
| En11 | communication = active + verified; non-member→required, blocked→blocked, unverified→email_not_verified | e2e En05 (message) + integration |
| En12 | a11y of new screens (keyboard/focus/44px/zoom/uk-en/axe) | e2e axe + Rstest |
| En13 | backfill: existing free user with progress OR comment/review keeps access after migration | integration — 00015 backfill on populated data (DownTo 14 → seed activity → UpTo 15): progress/comment/review each → self-active grant; paid & draft excluded; free-archived included; cross-source UNION + two-row dedup → one grant (course/postgres migration subtest, `make integration` green) |
| En14 | G10: instructor/admin enrols as a participant of another's course (membership active, role unchanged) | e2e En07/En14 |
| En15 | one row per (user,course): blocked/revoked «Записатися» → no 2nd active, no JOIN duplicates | e2e En04/En15 + integration |

## 3. Gate
Ordered `make check` (format-check → lint → typecheck → test[-race] → integration → check-generated →
build → up → reseed → smoke → e2e): **GREEN end-to-end, exit 0** — run to the last step (e2e 107 passed,
13.4m), no FAIL at any stage. The first attempt tripped `format-check` on 13 L7/L8 files Codex produced
without oxfmt; committed the format-only fix (`ee8719c`) and the re-run is the green above. Frontend
Rstest 638; Go integration green; `check-generated` clean.

## 4. Decisions taken this phase (artefacts)
- Membership = one authoritative row per (user,course), state = access power; full `UNIQUE(user,course)`;
  all operations are UPSERT transitions. Free content requires membership (plan §0/§3, G1/G2/G9).
- «Моє навчання» needs a list source → added `GET /learn/memberships` (L6c, plan amendment 2026-10-07),
  bounded, no pagination.
- `member` boolean dropped from the frontend Membership type (state is authoritative; it was true for
  revoked and caused the L7a bug).
- Revoke control gated to on_student reports (on_lecturer revoke is a dead `grant_invalid`).
- `recovery.spec` «retains seeded courses» scoped to the seeded slugs, not an absolute total — a
  globally-counted thing read by scope, per AGENTS; robust to any other spec's leak.
- **L5 resolve is idempotent** (Courses 2 L5 non-blocking, recorded not folded): resolving an unknown
  or already-resolved reportID is a no-op returning nil, not `ErrNotFound`. Mechanism: a mistyped id is
  silent — an admin gets no error. Accepted as consistent with moderation resolve; the arbiter stays
  `WHERE state='open'`, so a resolved report never blocks a fresh one.
- **L6a keeps authz in the application for orders enroll/teach routes** (not middleware), the uniform
  rule for that module: `Revoke`'s admin check lives in `application.Revoke` (→403, test-proven), unlike
  the older `grantCourse` which uses `administrator()` middleware. L6b's categorically-admin routes
  (queue/resolve) do get `administrator()` middleware (defense-in-depth), the deliberate asymmetry.
- The Codex stop/discuss channel paid for itself four times: a real L5 persisted-id contract gap (L6b);
  two reviewer briefing errors about a 404 (block/unblock has none; the report endpoints do); and one
  false alarm (the queue's bad-cursor code is `comment_cursor_invalid`, which is in the enum).

## 5. Defects found and fixed during the phase
- **L7a, real authz-display bug** (caught by Rstest behind Codex's local-green): CourseView branched on
  the backend `member` flag, true for revoked, so a revoked user saw «Ви учасник»; fixed to branch on
  state.
- **L6c, migration CHECK** (caught by the real Docker integration the sandbox can't run): the En01
  fixture gave a draft course a non-null `published_at`, violating `courses_check3`; the production
  query was correct.
- **L8 a11y** (caught by e2e axe): the active admin-nav link «Скарги учасників» sat at 4.39:1 contrast →
  12.77:1.
- **L8 report false-success** (caught while writing e2e): `learn.$.tsx` swallowed a comment-report
  rejection, so a 403 showed «Скаргу надіслано»; the rejection now renders the named error.
- Behind-gate spec errors the full e2e exposed (test bugs, not product): asserting the public outline
  endpoint as gated (En03/En04), a non-existent lesson-title heading (En05), stale `/learn`→`/courses`
  redirect assertions (route-guards), a wrong roster path/authz/shape (En04/En15).
- **L9 proof-gap** (caught by Courses 2 at the report gate, the phase-closing rule in action): En13 was
  tagged proven at integration, but the only backfill coverage ran 00015 on an empty DB (no-op) and
  seed self-active is a seed INSERT, not the backfill SELECT — the row was claimed by a same-named path,
  not an actual test. Closed with a DownTo(14)→seed activity→UpTo(15) migration subtest exercising the
  real 3-way UNION, the price/status filter, and cross-source dedup on populated data.

## 6. Limitations (named, not claimed away)
- **Paid re-enrolment after revoke is refused** (`course_already_owned`) — enrollment is senior to grant;
  re-purchase does not restore access (plan §4, behavioural change). In the UI a paid-revoked learner
  sees the purchase path that then refuses — the L4a dead-end surfacing, not an L7 defect.
- **media-bytes, preview** are proven at the nginx `auth_request` level by e2e (member 200 / revoked &
  non-member 403 / published-preview non-member 200); **backfill** (En13) and **grandfather** (En10
  price-set) are proven at integration only — a fresh e2e seed has no pre-enrolment history to convert,
  and grandfather is backend-only orders logic with no new enforcement surface.
- **«Моє навчання» has no pagination** — a learner's own membership count is bounded by the courses they
  enrol in (dozens), well under any page size; keyset to be added only if that stops holding.
- **Per-course author blocking and a learning (student/progress) overview** remain out of the instructor
  workspace (deferred with enrollment, `docs/instructor/report.md` §5) — now partially served by the
  roster; a cross-course learning overview is still not built.
- **e2e DB persistence**: the Compose e2e DB persists across runs, so an interrupted run leaks
  published fixtures that absolute-count assertions trip over. Mitigated by scoping `recovery.spec`;
  the durable option is a fresh-DB default for the authoritative run (noted for the backlog).

## 7. Next
The phase roadmap is complete with enrollment. No roadmap phase remains open. Production deployment was
deferred by the owner (2026-10-03, test mode) and is a human decision, not a queued phase. PR to `main`
after Courses 2's L8 row-by-row sign-off.
