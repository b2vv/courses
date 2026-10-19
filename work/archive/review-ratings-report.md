# Review — ratings phase report §2, row by row (commit 5e3f60e, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-04. Checked against: `docs/ratings/plan.md` §11 (observable result
per row) and `docs/ratings/report.md` §2 (claimed evidence).

## Method, and what it does not cover

For each row I compared the cited test's **assertions** with the plan's observable result. I did not
count a test as evidence just because its name matched.

I launched three verifier subagents, one per ten rows. **All three failed before their first step**
(an API safeguard error, with no working file written): 3 launched, 0 returned. I then did all 30 rows
myself.

| Test group | How I checked it |
|---|---|
| e2e (`frontend/e2e/ratings.spec.ts` at 5e3f60e) | Read now |
| FE unit (`ratings.test.tsx`, `ReviewBlock.test.tsx`, `course-repository.test.ts`) | Read now |
| Backend unit / contract / integration | Read in full during my T01–T05b reviews this session. Since then, these files changed only by the `containers.Start` startup wrapper (`d5b8726`) and one added removed-review app test (`fc3d240`). The subtest bodies I rely on are unchanged; I checked `git log` per file. |

I did not re-run any suite. The gate numbers are Courses 1's.

## Verdict

**27 rows hold. 3 rows are weaker than claimed: Rt05, Rt26, Rt27.**

None of the three is a product defect I can demonstrate. Each is a part of the row's observable result
that no test asserts, at least not at the level the report claims. Per AGENTS ("a phase is closed when
its rows are"), they should be either fixed (each is a small test) or re-stated as partial in §2,
**before the PR**.

## Rows weaker than claimed

| Row | Plan's observable result | What is actually asserted | Gap |
|---|---|---|---|
| **Rt05** | admin sees «Купити курс»; after a stub purchase and progress, **a review returns 201** | e2e: «Купити курс» visible, order `data-status="paid"`, then `writeReviewInUI(...)`. **Nothing after it.** `writeReviewInUI` clicks «Зберегти» and asserts nothing about the outcome. | The end of the row is unproven. If the save failed (a refusal or an error alert), the test still passes. Unit `TestLearningCourseCanPurchaseBy` and the FE entry tests prove only the button. Fix: after the save, assert the review is shown (own block or course page), or wait for the PUT response 201. |
| **Rt26** | archived course + grant holder: reviews visible; **edit 200**; after re-publish the card shows the rating | Integration `Rt26_archive_changes_review_reads_and_facts`: list and summary unchanged, facts `Entitled`. e2e «Rt26 Rt30…»: learning page shows the reviews, public list 404, card `★ 5,0 (3)` after re-publish. | **"edit 200" on an archived course is asserted nowhere** at integration or e2e. Only the domain unit `TestCanReview/archived with grant` (allowed) covers it. Fix: in the e2e, edit the grant holder's review on the archived course (UI or PUT) and assert success, or add a contract/integration PUT → 200 on archived facts. |
| **Rt27** | no summary table, counter column or trigger for reviews | Integration `Rt27_schema_has_only_review_tables_and_no_review_trigger`: counts tables **named exactly** `discussion_reviews`/`discussion_review_reports` (= 2), `courses` columns `ILIKE '%rating%'` (= 0), non-internal triggers on `discussion_reviews` (= 0). | Narrower than the claim. A `course_review_stats` table, a `review_count` column on `courses`, a trigger on `courses`, or a materialized view would all pass. Fix: assert that the migration adds exactly the expected relations (compare `pg_class` before/after 00013, or list all tables/matviews referencing `discussion_reviews`), and widen the column check to `%review%`. |

## Rows that hold, with notes where the evidence is thinner than it reads

| Row | Verdict | Note |
|---|---|---|
| Rt01 | holds | e2e writes through the UI and asserts the body and «пройдено 0% на момент відгуку» **on the course page**. The 201 is at contract. |
| Rt02 | holds | Integration facts: Entitled, HasProgress, Paid (real SQL). |
| Rt03 | holds | e2e: `2 з 5` checked, then «2 з 5» and «змінено» on the course page. |
| Rt04 | holds | e2e: the original is absent after delete, and the rewrite is visible. "Count drops by 1" is implied by absence on a one-review course, not counted. Integration: delete frees the slot. |
| Rt06/07/08 | holds | e2e asserts «Новий курс» at 2 and `★ 4,7 (3)` at 3, on both the card and the course page. The reviews are **seeded by SQL**; aggregation and rendering take the production path. FE units and `TestSummary` add the boundaries. Integration (course) asserts 4.7/3. |
| Rt09 | holds | Integration: summary 2/7 and list without the removed review. e2e: the removed review leaves the page and the card returns to «Новий курс». The e2e clicks «Зняти» and navigates without waiting for the removal response, which is a possible flake source. |
| Rt10/11 | holds | Unit `TestNewReviewContent` (19/20/2000/2001 after trim, stars 0/1/5/6) plus contract `put boundaries` with exact codes. FE rune count. |
| Rt12 | holds (freeze) | Stored 2/4 is unchanged after +2 lessons and completion. The capture at write is a **hardcoded 2/4 passed to `CreateReview`**, not computed from real facts. "Capture from facts at write" is proven only by app units (fake facts) plus Rt13's real facts SQL, and on the real path only at 0% (Rt01). Acceptable for "frozen"; worth knowing. |
| Rt13 | holds | Real `ReviewFacts` give 6/6 → `UpdateReview` stores 6/6, `edited_at` set, same id. |
| Rt14 | holds | e2e: 45 seeded, 20 newest first, `toBeFocused` on «Показати ще» after 40, `items.nth(40)` focused after 45, live «Усі відгуки показано». FE unit plus a repository test for the `""` last-page cursor. |
| Rt15/18 | holds | Unit, app and contract, with the exact 403 codes. |
| Rt16/17 | holds | Integration: real facts + `CanReview` → `review_purchase_required`. Contract also has a generic `review_purchase_required` 403 (the report under-claims; the plan asks for contract). |
| Rt19 | holds | Contract: a second PUT returns 200 (update). Integration: two concurrent creates → 1 success + 1 `review_already_exists`; Rt13 keeps the same id on update. |
| Rt20 | holds | Integration: after a SQL-revoked grant, the review stays in list and summary and facts are not entitled. The 403 on edit is asserted only at app unit with a fake (`TestPutReviewRefusesEditAfterAccessIsLost`). The FE shows the reason. The revocation is by SQL fixture, as the report says (nothing in code revokes). |
| Rt21 | holds | e2e: the owner's queue row has **no «Зняти»**. Contract: owner remove → 403 `review_moderation_forbidden`. FE row test. |
| Rt22 | holds | Integration: a forced log failure leaves the review unremoved with 0 log rows. One log row with `review_id` and the actor. e2e removal hides the review. |
| Rt23 | holds | Integration: create after removal and delete after removal → `review_removed`. Contract: delete after removal → 409. The 2026-10-04 own-review decision (`review: null`) does not change these write results; `PutReview` still returns `review_removed`. |
| Rt24 | holds | Unit (store untouched), contract (403 + code), integration (before == after). |
| Rt25 | holds | Integration: own report refused, duplicate → one row. Contract: own report → 403 code. |
| Rt28 | holds | FE: program heading visible, `role=alert` text, «Спробувати ще» calls retry. Plus the section-error-only-without-data test. |
| Rt29 | holds | FE: a 503 on save keeps the typed body and shows «Не вдалося зберегти». The test name also says "names a removed review", but the body does not assert that message. The name overclaims; the row (503) holds. |
| Rt30 | holds | e2e: grant holder on an archived course sees the reviews on the learning page; public list → 404. Contract: permitted/refused. |

## The T07 / T06b-fixes notes in 5e3f60e — verified

- **Rt14** seeds 45 and asserts `toBeFocused` on the button between pages and on `items.nth(40)` after
  the last. This closes my T07 gap.
- **`afterAll`** now archives first. It re-reads `status`/`updatedAt` per course and skips
  already-archived courses. It deletes the namespaced reviewers in `finally`. Both T07 robustness notes
  are closed.
- **Delete error** `role="alert"` now renders **inside** the Dialog (`ReviewBlock.tsx:158`).
- **Invalidation** after put/delete now also covers `['learn','course']`, `['catalog']` and
  `['course']`.

## One hygiene question (non-blocking)

5e3f60e also commits four `docs/foundation/screenshots/*.png`. These were already modified in the
working tree before this phase's commits: `catalog-desktop.png` 988 KB → 185 KB, `course-desktop.png`
88 KB → 108 KB. If they were regenerated on purpose because the cards now show ratings, fine; say so in
the commit or report. If they are a by-product of a local e2e run, they don't belong in the ratings
report commit.

## Recommendation

Do not open the PR with §2 saying "30/30 proven". Either:
- **(a)** add the three small assertions (Rt05 save outcome, Rt26 edit on archived, Rt27 schema
  diff), re-run the affected levels, and keep 30/30; or
- **(b)** mark Rt05, Rt26 and Rt27 as partial in §2, naming the missing part.

**(a)** is cheaper than the argument.
