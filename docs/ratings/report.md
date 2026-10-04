# Phase ratings — delivery report

Delta to [plan.md](plan.md), [spec.md](spec.md) and [assumptions.md](assumptions.md). Written 2026-10-04 at
phase close from the slice commits T01–T07 and the ordered gate. Roles: Courses 1 planned, gated, fixed and
committed; Codex wrote the code from briefs (`work/brief-codex-t0*.md`); Courses 2 reviewed every slice
(`work/review-ratings-*.md`). Test mode only — there is no production server (owner, 2026-10-03).

## 1. What was done

A student who has started a course (and, for a paid course, holds a grant) rates it from the learning page:
1–5 stars and a 20–2000 character text; the review records how much of the course its author had completed
at the moment of writing. The author can edit (while the right holds) and delete; a review removed by an
admin cannot be replaced, and the author sees that it was removed. Anyone, signed in or not, sees on the catalog
card and on the course page `★ 4,7 (3)` once a course has three counted reviews, or «Новий курс» before, and
reads the reviews newest first with «Показати ще». Signed-in viewers report a review; it joins the existing
instructor moderation queue, where only an admin can remove it — the course owner can only dismiss reports.
Grant holders of an archived course still see its reviews on the learning page. An admin can buy a course from
its outline (they never see a locked lesson). The catalog card also shows the lesson count.

Backend: `internal/contracts/rating` (threshold + rounding, one pure function); reviews, reports and the
log in `discussion` (migration `00013_ratings.sql`); the catalog aggregate in `course` SQL (no summary table,
R6); no admin self-grant in `orders`. Frontend: shared review list, `RatingSummary`, `ReviewBlock`,
`ReviewModerationRow`, the radio-group primitive. e2e: `frontend/e2e/ratings.spec.ts`.

## 2. Row-by-row acceptance — 30 rows (Rt01–Rt30)

Levels: unit · contract (HTTP vs OpenAPI) · integration (real Postgres) · e2e (browser vs Compose). Every
row has an automated level; manual rows: 0.

| # | Evidence (test) | Levels | Result |
|---|---|---|---|
| Rt01 | e2e `ratings.spec.ts` «Rt01 writes a free-course review…»; contract `TestT05ReviewRoutesContract/put boundaries…` | e2e, contract | proven |
| Rt02 | integration `TestDiscussionPostgres/ratings_repository/Rt02_facts_for_buyer_grant_and_progress` | integration | proven |
| Rt03 | e2e «Rt03 edits 4 stars to 2…»; FE `ReviewBlock.test` «submits… and opens edit prefilled» | e2e, unit | proven |
| Rt04 | e2e «…Rt04 deletes through the dialog then writes again»; integration «concurrent create and removed/delete semantics» (delete frees the slot) | e2e, integration | proven |
| Rt05 | e2e «Rt05 an admin buys a paid course through the offline WayForPay stub…»; FE `CourseLearningView.test` purchase-entry tests; unit `TestLearningCourseCanPurchaseBy` | e2e, unit | proven (stub, not live WayForPay — §6) |
| Rt06 | FE `ratings.test.tsx` «shows «Новий курс» below the threshold»; e2e «Rt06/Rt07/Rt08…»; unit `TestSummary` | unit, e2e | proven |
| Rt07 | integration course `Rt07_Rt09_catalog_aggregates…`; FE `ratings.test.tsx` Rt07; e2e | integration, unit, e2e | proven |
| Rt08 | FE `ratings.test.tsx` Rt08; e2e | unit, e2e | proven |
| Rt09 | integration discussion `Rt09_removed_review_is_omitted…` and course `Rt07_Rt09…`; e2e moderation scenario | integration, e2e | proven |
| Rt10 | unit `TestNewReviewContent` (19/20/2000/2001 after trim, spaces-only padding); contract `put boundaries…`; FE «counts runes after trimming» | unit, contract | proven |
| Rt11 | unit `TestNewReviewContent` (stars 0/1/5/6); contract `put boundaries…` | unit, contract | proven |
| Rt12 | integration `Rt12_frozen_progress_and_Rt13_edit_recaptures` | integration | proven |
| Rt13 | same subtest (edit recaptures 6/6) | integration | proven |
| Rt14 | e2e «Rt14 pages 20 + 20 + 5 by keyboard…» (focus stays on «Показати ще» on the intermediate page, moves to the 41st review after the last, live text); FE `ratings.test.tsx` Rt14; repository test for the `""` last-page cursor | e2e, unit | proven |
| Rt15 | unit `TestCanReview` (owner with grant + progress); app `TestPutReviewEligibilityOrder/owner`; contract `put boundaries…` (403 `review_owner_forbidden`) | unit, contract | proven |
| Rt16 | integration `Rt16_preview_only_refused` | integration | proven |
| Rt17 | integration `Rt17_admin_role_is_not_entitlement` | integration | proven |
| Rt18 | unit `TestCanReview`; app `TestPutReviewEligibilityOrder/not_started`; contract (403 `review_not_started`) | unit, contract | proven |
| Rt19 | contract `put boundaries…` (second PUT → 200, one review); integration «concurrent create…» (two concurrent creates: one succeeds, one `review_already_exists`) | contract, integration | proven |
| Rt20 | integration `Rt20_revocation_preserves_review` (grant revoked by SQL fixture — nothing in code revokes, §6); app `TestPutReviewRefusesEditAfterAccessIsLost`; FE «shows the reason instead of Edit…» | integration, unit | proven |
| Rt21 | FE `ReviewModerationRow.test` «offers removal only to an admin (G9)…»; contract (owner remove 403); e2e moderation scenario | unit, contract, e2e | proven |
| Rt22 | integration `Rt22_removal_logs_atomically` (forced log failure leaves the review); e2e admin removal | integration, e2e | proven |
| Rt23 | contract `own delete report remove dismiss` (delete after removal 409); integration «concurrent create…» (create after removal → `review_removed`) | contract, integration | proven |
| Rt24 | unit `TestGrantCourseRequiresAdministratorAndWritesAudit/administrator cannot grant themselves`; contract `TestPurchaseAndGrantHTTPMatchOpenAPI/admin cannot grant themselves`; integration orders «administrator cannot grant themselves or create a row» | unit, contract, integration | proven |
| Rt25 | integration `Rt25_own_report_refused_and_duplicate_idempotent`; contract (report own 403) | integration, contract | proven |
| Rt26 | integration `Rt26_archive_changes_review_reads_and_facts`; e2e «Rt26 Rt30…» (republish restores the card) | integration, e2e | proven |
| Rt27 | integration `Rt27_schema_has_only_review_tables_and_no_review_trigger` | integration | proven |
| Rt28 | FE `ratings.test.tsx` «keeps the program visible and offers a retry when reviews fail» and «shows the section error only when the first reviews page fails» | unit | proven |
| Rt29 | FE `ReviewBlock.test` «keeps failed text and names a removed review» | unit | proven |
| Rt30 | e2e «Rt26 Rt30…» (grant holder sees reviews on the learning page, public page 404); contract `TestT05ReviewRoutesGaps/learning list permits an archived course grant holder`, `…/learning list refuses a viewer without a right` | e2e, contract | proven |

Also verified beyond the rows: cross-origin review PUT refused (e2e — the handler harness stubs `sameOrigin`);
axe (`wcag2a`, `wcag2aa`, `wcag21aa`) on the course page with reviews, the learning form and the teach queue.

## 3. Gate

Ordered `make check` on the branch head (2026-10-04): format-check, lint (incl. the frontend architecture
check), typecheck, test, integration, check-generated, build, up, drain-queue, seed-identities, seed, smoke,
e2e — **exit 0**. Frontend unit **588 passed / 0 failed**; Go unit and integration: no `FAIL`; e2e **98 passed
of 98** in 11.6 min, all nine `ratings.spec.ts` scenarios among them. The run was made with the owner's other
Docker stacks (cassiopeia, cargotrans — 15 containers) paused by the owner's leave, because `cassiopeia-webapi`
held 6.6 of the Docker VM's 15.35 GiB and the gate ran 1.2 h with 503s and container-start failures; all 15
were restarted after the run (15/15).

Earlier full runs this phase, recorded rather than hidden: they surfaced the defects in §5; one failed only on
B39 (B-23); two failed on memory pressure from the other stacks; one on a reviewer-email collision introduced by
widening Rt14 (fixed — disjoint index ranges per scenario).

## 4. Decisions taken this phase (artefacts)

All in `assumptions.md` or `spec.md` with a date, an author and a check: GATE 1 G1–G10 (owner, 2026-10-03);
R6 no summary table, and its trigger made absolute — catalog page median ≤ 5 ms (owner, 2026-10-03; the
">2×" ratio measured nothing against a 16 µs baseline); no duration on the card; archived-course list via the
learning API (owner, AC10); a removed review reads as `review_removed` for its author (Courses 1); production
deployment deferred (owner, `work/AGENDA.md`); the testcontainers reaper race workaround (owner,
backlog B-27).

R6 measured (`catalog_measure_test.go`, tag `measure`): 100 courses, 20 000 reviews Zipf top 2 000, `VACUUM
ANALYZE`, median of 5 — **1.306 ms / 1.314 ms** with the aggregate (16 µs without) — within the 5 ms budget.

## 5. Defects found and fixed during the phase

Each passed whatever stood before it and was caught later:
- **Learning-course schema rejected every real course** (`.strict()` without `rating`/`lessonCount`, which the
  API sends since T04) — found by a unit test whose stale fixture had hidden it.
- **`""` last-page cursor** kept «Показати ще» forever — found by e2e (unit fixtures used `null`).
- **Star radio group had no CSS**; its group intercepted clicks — stars could not be picked with a mouse —
  found by e2e.
- **Stale learning list** after a review write/delete (only the own review was invalidated) — e2e.
- **Teach queue review rows unvalidated** (`'★'.repeat` throws on a negative count) — Courses 2 review.
- **Focus dropped to `<body>`** from a natively disabled focused button — Courses 2 review; Rt14 now proves the
  intermediate page.
- **Two constitution violations** (application imported `contracts/rating` and `google/uuid`) — architecture test.
- **Contract gaps:** removed own review shown as live; teach adapter named no review codes; `canPurchase` button
  duplicated beside the locked lesson.
- **Shared state:** the e2e left reviewers that pushed `admin-users` P08's fixtures off the paged list, and
  courses that broke `recovery.spec`'s total — the spec now cleans up (AGENTS rule).
- **Integration gate:** the testcontainers reaper create→start race (upstream #3867/#3918) — TEMPORARY retry,
  proven by a deterministic real-path test (B-27).

## 6. Limitations (named, not claimed away)

- **A lesson without a ready video can never be completed**, so it caps every review's percentage below 100 %.
  Deliberate: the percentage describes the course as published (plan §3).
- **No duration on the card** (decided; a true total needs a per-lesson lookup on the hottest read).
- **Losing access keeps the review** (G4) — and nothing in code revokes a grant today (refunds are ignored,
  `docs/payment/report.md` §6), so Rt20 uses an SQL fixture for the revocation.
- **Rt05 uses the offline WayForPay stub**; the live provider path was proven in the payment phase.
- **The API still sends `""` as the last-page cursor**; the frontend maps it at the adapter boundary. Making it
  `null` and nullable in OpenAPI is backlog (`work/backlog.md`).
- **Archived e2e courses accumulate** (~8 per full run) in the admin course list; checked harmless today — the
  list is newest-first and `admin.spec` opens its course by URL, not by position.
- **R6 is measured at 20 000 reviews**; the 5 ms budget is the revisit trigger, not a load test.
- **B39 (`playback.spec`) failed once** in this phase's gate runs (B-23, no diagnosis; unrelated to reviews).
- **The reaper retry is TEMPORARY** (B-27) until a testcontainers-go release contains #3868.

## 7. Next

Merge the phase PR into `main`. The agenda's next phase is chosen by the owner (`work/AGENDA.md`).
