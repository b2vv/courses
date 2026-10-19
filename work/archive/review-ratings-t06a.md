# Review — ratings T06a, frontend public side (commit c5468dc, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-03. Against: spec AC8/AC9, plan §6/§11 (Rt06–Rt08, Rt14, Rt28),
AGENTS (layering, validation, a11y), briefs `work/brief-codex-t06a*.md`.

## Verdict

**Approve. 0 blocking, 5 non-blocking.** Layering, schema validation, the screen-reader rating
sentence, Rt28 and report visibility all hold. Two of the non-blocking items are real defects, but
small: a «Показати ще» failure wipes the loaded list, and focus is lost after the last page loads. I
suggest fixing both in T06b.

What I ran: `rstest run` on `ratings.test.tsx`, `course-repository.test.ts` and
`format-rating.test.ts` → no failures, none skipped. I did not re-run the whole gate; Courses 1
reports it at exit 0, frontend 558/0.

## Focus — PASS

- **Layering.** The route (`routes/courses.$slug.tsx`) composes query options from
  `app/composition/catalog-queries.ts` (`reviews` infinite, `ownReview`, `report` mutation). There is
  no fetch in the route or in presentation. `CourseView`/`ReviewCard` receive data and callbacks
  only. The application (`catalog.ts`) validates the slug before calling the port. `catalog` comes
  from the route context, so queries stay request-scoped. Making the port's review methods required
  (they were optional) is the right fix.
- **External data validated.** `reviewSchema` checks: id uuid, stars int 1–5, non-negative
  integers, `totalLessons >= completedLessons` (refine), ISO datetimes with offset. The page has
  `items.max(20)` and a nullable cursor. Rating is `average 1..5, count int ≥3, nullable`, which
  mirrors the server threshold, so a premature average cannot arrive. `newCourse` re-checks
  rating/lessonCount in the domain.
- **Percent matches the backend.** `reviewPercent` computes
  `floor((c*100 + t/2)/t)`; the backend computes `(c*100 + floor(t/2))/t`. These differ only when
  `t` is odd, and in that case the fraction can never be exactly .5 (even numerator, odd product).
  Results are identical for all integer inputs.
- **A11y, rating summary.** The visible `★ 4,7 (3)` is `aria-hidden`. An sr-only sentence is read
  instead: «Рейтинг 4,7 з 5, 3 оцінки», with proper uk plurals. This avoids `aria-label` on a generic
  element. The average is one decimal in the locale (`Intl.NumberFormat`, min=max=1 fraction digit),
  so 5 renders as «5,0». Stars per review are `aria-hidden` and paired with sr «4 з 5» (plan §6).
- **Rt28.** The detail page is its own suspense query. The reviews query is non-suspense, and an
  error renders only inside the reviews section: `role="alert"`, a message and «Спробувати ще». The
  outline and description stay rendered.
- **Report visibility.** The button renders only when `signedIn && review.id !== ownReviewId`.
  Anonymous viewers never see it. The author is identified by review id from
  `GET /learn/courses/{id}/review`, not by display name (which can collide), which is the right
  call. A refusal by the server (`report_own_review`) still has its own message.
- **Live region on «Показати ще».** A polite `<output>` exists before the click (it mounts with the
  button) and announces «Завантажуємо відгуки». See №2 for the completion and focus side.

## Non-blocking

1. **A failed «Показати ще» replaces the whole list with the error (defect).** The route passes
   `reviewsError={reviews.isError}`. In TanStack Query v5, `isError` is also true when *fetching the
   next page* fails: `infiniteQueryObserver.js:38`, `isFetchNextPageError = isError && fetchDirection
   === "forward"`. So a next-page failure hides every review already shown. «Спробувати ще» then
   `refetch()`es all pages. Fix: show the section error only when no data has loaded
   (`isError && data === undefined`). For `isFetchNextPageError`, keep the items and show an inline
   error with retry next to «Показати ще» (`fetchNextPage`).
2. **Focus is lost after the last page (a11y defect, AGENTS keyboard/focus).** The button and the live
   region are inside `hasMore ? … : null`. When the last page arrives, the focused «Показати ще»
   unmounts and keyboard focus drops to `<body>`. Completion is also never announced: the region goes
   from «Завантажуємо…» to empty, then disappears. Fix: keep the live region outside the condition
   and announce the result (e.g. «Показано N відгуків»), and move focus to the first newly loaded
   review, or to the list. axe in T07 will not catch this; it needs a keyboard check in e2e.
3. **Own-review is cached across users.** The key `['course', courseId, 'own-review']` has no user
   in it, and `useSignOut` only sets the session to null (`use-sign-out.ts`). If user A signs out
   and user B signs in in the same tab, B can briefly see A's own-review id. The effect is that the
   report button is hidden on A's review (or shown on B's) until a refetch. It is cosmetic, since the
   server enforces the rule. Fix: put the user id in the key, or `removeQueries(['course'])`-scope it
   on sign-out. The other learn queries share this pre-existing pattern.
4. **Small UX gaps.**
   - The report button flashes on one's own review until `ownReview` loads.
   - `reportError` is not cleared on a later success, so «Скаргу надіслано» can show next to a stale
     error.
   - The pending state renders only the heading, with no loading text or `aria-busy`. Reviews are
     also not prefetched in the loader, so SSR HTML has an empty section.
5. **«змінено» has no date.** Plan Rt03 says «the page shows … an «змінено» date». T06a shows only
   the word. Either add `editedAt` (formatted like `createdAt`) or reword Rt03. This needs a decision,
   because it is a plan row.

## Out of scope, FYI

`d5b8726 fix(test): re-enable ryuk with a temporary retry for its create-to-start race` sits just
before this commit. It moves away from the «Ryuk disabled» decision that T05b recorded. I have not
reviewed it. If you want it reviewed against the source path in `work/research-ryuk.md` §7, send it
over.
