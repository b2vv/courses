# Review — T07 e2e `frontend/e2e/ratings.spec.ts` (landed in fc3d240; 4bffd0a = tasks.md)

Reviewer: Courses 2. Date: 2026-10-04. Checked against: plan §11 e2e column, AGENTS (counted-thing
cleanup; evidence on the production path).

## Verdict

**Approve. 0 blocking, 4 non-blocking.** The spec exercises the production UI/HTTP path. It covers
Rt01, Rt03–Rt09, Rt14, Rt21, Rt22, Rt26, Rt30, cross-origin PUT → 403 (which closes the T05 carry-over
gap) and axe on three screens. It cleans up what it counts.

Your first question is answered "yes, for the last page" only. The spec does not guard the
intermediate-page focus loss, which was the actual defect.

## Q1 — Would the e2e fail if focus dropped to `<body>`? Only on the last page

- **What Rt14 does.** It seeds **25** reviews and presses Enter on «Показати ще» once. That single
  click loads the **last** page: the button unmounts, and the effect focuses the first new review. If
  that effect broke, focus would sit on `<body>` and
  `expect.poll(activeElement.tagName).not.toBe('BODY')` would fail. So **the last-page path is
  guarded**.
- **What it misses.** The defect I flagged in T06b was the **intermediate** page. There, the
  `disabled` focused button dropped focus while `hasMore` stayed true, and nothing restored it. With 25
  reviews there is no intermediate page. If someone reverted `isPending` → `isDisabled`, focus would
  drop during loading, the last-page effect would restore it, and **Rt14 would still pass**.
- **Fix.**
  1. Seed **45** reviews (three pages).
  2. Press Enter, wait for 40 items, then `await expect(more).toBeFocused()`.
  3. Press Enter again, wait for 45, then
     `await expect(list.getByRole('listitem').nth(40)).toBeFocused()`.

  This guards both paths with exact targets.
- **Assertion strength.** `not.toBe('BODY')` on `tagName` is weak: it also passes when `activeElement`
  is `null`/`undefined`, or when focus is on any unrelated element. Prefer `toBeFocused()` on the
  expected element. (Chromium applies the HTML focus-fixup rule, which is why this needs a real
  browser rather than jsdom.)

## Q2 — Cleanup in `afterAll`: correct in scope, two robustness gaps

What it does right:
- `DELETE FROM users WHERE email LIKE '<runNamespace>-%@example.test'` removes every reviewer the run
  inserted. Their reviews cascade (`author_id … ON DELETE CASCADE`), and so do the reports on those
  reviews (`review_id … CASCADE`). The reporters in the spec are fixture users reporting namespaced
  reviews, so `reporter_id` (no cascade) never blocks the delete. This matches the comment about the
  admin-users P08 precedent.
- Every course goes into `published` right after its publish call (`createCourse`, line 105), so there
  is no leak window. Rt26 replaces its entry with the republished token (lines 412-414), so `afterAll`
  archives it with a current `updatedAt`.

Gaps (non-blocking):
1. **Order and isolation.** The SQL delete runs first, then the archive loop. If the delete throws
   (FK, psql, Docker hiccup), the archive loop is skipped. The courses then stay **published** and
   break the catalog counts in `catalog.spec`/`recovery.spec` (`pagination.total: 3`) for the next spec
   and run. Run the archive loop first, or put each step in its own try/finally so one failure cannot
   skip the other.
2. **A stale token after a mid-test failure.** If Rt26 fails between archiving and republishing, its
   `published` entry still holds the pre-archive `updatedAt`. `afterAll` then gets 409
   `course_modified` and adds a secondary failure on top of the real one. Make the cleanup tolerant:
   re-read the course status/`updatedAt` before archiving, or treat "already archived" as done.
3. **Rows that stay behind by design.** These are named here so they are not forgotten:
   - fixture-user reviews and progress on archived run courses;
   - admin-issued grants to the fixture student;
   - Rt05's admin order/payment rows from the stub.

   None of them appears in a counted view today: archived courses are out of the catalog, and grant
   lists are per course.
4. **Archived run courses accumulate** (about eight per run) in the admin course list. I could not
   confirm, within this review, how that list is ordered and paged. If it is paged and not
   newest-first, it can displace fixture rows that `admin.spec` finds by title, like the users/P08
   precedent. Worth one check.

## Also verified

- Fixtures go through the production path where it exists. SQL is used only for rows with no public
  creation API (25 authors) and for progress, and the spec says so.
- `accessibility.spec`'s learning mock gained `rating`/`lessonCount`/`canPurchase`, matching the now
  strict `playback-schema`.
- The gate is reported as a full ordered `make check` with e2e 98/98 (run with the other projects
  paused). I did not re-run e2e myself.
