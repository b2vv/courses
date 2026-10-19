# Re-review — T06b fixes (commit fc3d240, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-04. Checked against: `work/review-ratings-t06b.md`, B1, and
non-blocking №1–№9.

## Verdict

**Approve. 0 blocking, 4 non-blocking.** B1 and all nine non-blocking items are closed. The three
defects that only e2e exposed are fixed correctly.

## Closed

- **B1.** `queuePage` now validates a review row with `reviewSchema.safeParse` (`teach-repository.ts`),
  and there is a malformed-row test. A real backend queue row passes the schema: `presentReview`
  emits every field, with `completedLessons`/`totalLessons` = 0 and `editedAt` = null.
- **Focus.** «Показати ще» and «Поскаржитись» use `isPending`, not `isDisabled`
  (`CourseView.tsx:177-179`, `:237`), so they stay focusable while pending.
- **№1.** `formatDate` is used for `createdAt`/`editedAt`, and «змінено {date}» gets its date.
- **№2.** Edit renders only when `refusal === undefined`. Otherwise the reason is shown (G4).
- **№3.** This is decided as `review: null` + `eligibility: review_removed`. The backend
  (`OwnReview`) sets it when `RemovedAt != nil`, it is recorded in `assumptions.md`, and there is a
  unit test.
- **№4.** Delete failures are caught and shown.
- **№5.** An own-review request failure has its own branch (`ownError`).
- **№6.** The length counts code points after trim (`[...text.trim()].length`). The live region
  announces only when the text is too long.
- **№7.** The star choices use `t('ratings.stars')`.
- **№8.** The own review's stars are `aria-hidden` + sr-only.
- **№9.** The buy entry shows only when `canPurchase && !hasLocked`.
- **E2E-only defects.**
  - The last-page `""` cursor is mapped to `null` in both repository paths, so «Показати ще» no
    longer stays forever.
  - `RadioGroup` now uses `.ui-choice` (44 px, focus ring), and the label is the click target.
  - put/delete now also invalidate the learning list.

## Non-blocking

1. **The delete error is rendered outside the modal Dialog.** `remove()` catches and sets `error`,
   but `<p role="alert">` sits in the article, behind the still-open modal. React Aria's modal hides
   outside content from assistive tech, so a screen-reader user does not hear the failure. Render the
   error inside the dialog, or close the dialog on failure before showing it.
2. **The rating summary goes stale after a write.** `invalidateReviews` refreshes the own review and
   the learning list. It does not refresh the learning course (its `rating` summary) or the catalog
   caches (`['course', slug]`, `['course', slug, 'reviews']`). After writing the third review, the
   learning page summary still says «Новий курс» until the next refetch.
3. **The `""` cursor should be fixed at the boundary too.** The frontend mapping works. But the
   contract says `nextCursor: string`, and the API uses `""` to mean "no more". `null`, plus
   `nullable` in OpenAPI, would stop the next client from repeating the bug. Note: I missed this in
   T06a. The T06a contract test used a non-empty `"next-page"` cursor, and only the real-API e2e
   exposed it.
4. **Commit hygiene.** `ratings.spec.ts` (471 lines) landed in this fix commit. `4bffd0a` «T07 browser
   e2e ratings.spec.ts with axe» changes only `work/tasks.md`. The history attributes the spec to the
   wrong commit, which goes against AGENTS' "small logical commits". Worth a note in the T07 report.
   Rewriting history is not worth it.
