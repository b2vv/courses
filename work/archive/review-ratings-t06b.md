# Review — ratings T06b (commit c9bbd2b, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-04. Checked against: spec AC1/AC4/AC6/AC7/AC10, plan §6/§7
(Rt29, Rt30, admin buy entry), AGENTS (validate external data, a11y), T06a carry-over
(`work/_t06b-carryover.md`).

## Verdict

**Changes requested: 1 blocking, 9 non-blocking.**
- The blocking item violates AGENTS: the teach queue renders review rows that were never validated.
  It is small to fix.
- Of the T06a fixes, #1 (section error vs. next-page error) is correct. #2 (focus) is only half met:
  in a real browser focus still lands on `<body>` during an intermediate «Показати ще».

What I ran: rstest on `ReviewBlock`, `ratings`, `ReviewModerationRow`, `teach-repository` and
`playback-repository` tests → 0 failed files, 0 failed tests. I did not re-run the whole gate;
Courses 1 reports exit 0, 578/0.

## Q1 — Does any other strict schema have the `playback-schema` gap? — No other *strict* gap, but one unvalidated read (blocking)

Method: list every response DTO field that changed on the branch
(`git diff main...HEAD` over `backend/internal/*/adapters/http/*.go` + `openapi.yaml`), then find each
frontend reader.

Changed response shapes:
- `courseDTO`: +`rating`, +`lessonCount`. Used by the catalog list, course detail and learning
  course.
- `learningCourseDTO`: +`canPurchase`.
- `queueItemDTO`: `kind`, a nullable `comment`, `review`.
- New review DTOs: public page, own review.
- New error codes (orders, discussion).

Readers, one by one:

| Response | Frontend reader | Status |
|---|---|---|
| catalog list / `CourseDetail` | `course-schema.ts` `courseSchema`/`courseDetailSchema`: plain `z.object` (strips extras), `rating` + `lessonCount` required | ✔ no gap |
| learning course | `playback-schema.ts` (strict) | ✔ fixed in this commit: `rating` (strict, nullable), `lessonCount`, `canPurchase` |
| public / learning review page, own review | `publicReviewPageSchema`, `ownReviewSchema`: non-strict, stars/int/refine | ✔ |
| error envelopes | `errorSchema` / playback error: `code: z.string()` | ✔ new codes pass |
| admin / grant / purchase / transcript / media strict schemas | their DTOs did **not** change on this branch (the only orders change is a new error code) | ✔ untouched |
| **teach queue item** | `teach-repository.ts` `queuePage`, a hand-written type guard | ✘ **see B1** |

### B1 (blocking) — teach queue review rows are rendered without validation

`queuePage` (`teach-repository.ts:79-83`) accepts
`item.kind === 'review' && item.comment === null && record(item.review)`. That only checks that
`review` is an object. The type guard still asserts `ModerationQueuePage`, so `review.stars`, `body`,
`authorDisplayName` and `id` reach `ReviewModerationRow` unchecked. The comment above it ("T06 renders
review queue rows. Preserve the existing teach surface until then") is now stale: T06b *does* render
them.

This is exactly what AGENTS forbids: "no assertions that disguise unsafe input. Validate external
data." It also has a concrete failure mode. `'★'.repeat(item.review.stars)` throws `RangeError` for a
negative or infinite `stars`, which crashes the teach page render. For a missing `stars` it prints
nothing.

Fix: validate the review like the comment branch does:
- `uuid(id)`;
- `stars` an integer 1–5;
- `body` a string;
- `authorDisplayName` a string or null;
- `date(createdAt)`;
- `editedAt` null or a date;
- `completedLessons`/`totalLessons` (if present) as in `reviewSchema`.

Or reuse `reviewSchema` from `course-schema.ts`. Add one teach-repository test with a malformed review
row → `invalid_response`.

## Q2 — T06a #1 and #2 against TanStack v5 and the browser

- **#1 — correct.**
  - `Reviews` now shows the section error only when `error && reviews === undefined`.
  - Both routes pass `isFetchNextPageError` (TanStack v5: `isError && fetchDirection === "forward"`,
    `infiniteQueryObserver.js:38`).
  - A failed next page keeps the items and shows an inline `role="alert"` with a retry that calls
    `fetchNextPage()`.
  - Tests cover both cases.
- **#2 — half met. Focus still lands on `<body>` on intermediate pages in a real browser.**
  - The live region is now always mounted and announces «завантажено N» / «Усі відгуки показано». ✔
  - After the **last** page, the first new review (`tabIndex=-1`) gets focus. ✔
  - But «Показати ще» is `isDisabled={isLoadingMore}`. Our `Button` is `Aria.Button`, and its
    `useButton` renders a **native `disabled` attribute** (seen in the built bundle: `disabled:
    isDisabled` for `elementType === 'button'`). By the HTML focus-fixup rule, a focused element that
    becomes disabled loses focus to the body.
  - On the **last** page the effect brings focus back. On an **intermediate** page `hasMore` stays
    true, the effect does nothing, and focus stays on `<body>`.
  - The unit test passes only because jsdom/happy-dom don't run focus fixup ("a fake proves the
    fake").
  - Fix: use React Aria's `isPending={isLoadingMore}` (it keeps the button focusable and announces
    pending), or restore focus to the button after the page arrives. Verify with a keyboard
    Playwright step in T07: press «Показати ще», then assert `document.activeElement` is not `body`.
  - The same applies to «Поскаржитись»: `isDisabled={reported || isReporting}` drops focus to body on
    press, and permanently after success.

## Non-blocking (defects first)

1. **Literal «змінено {{date}}» on one's own edited review.** `ReviewBlock` calls `t('ratings.edited')`
   without `date`, while the locale is now `'змінено {{date}}'`. I checked i18next with this exact
   resource: it returns the placeholder verbatim. In the same block, `<time>` prints the raw ISO
   `createdAt` instead of `formatDate`.
2. **Dead «Редагувати» after losing access (G4).** «Редагувати» renders whenever `own.review !== null`,
   but the form renders only when `eligibility === 'ok'`. The refusal sentence renders only when
   `own.review === null`. So an author who lost access clicks Edit, nothing happens, and nothing
   explains why. Fix: hide Edit when `eligibility !== 'ok'` and show the refusal sentence under the
   existing review. Delete stays available (AC6).
3. **A removed review is invisible as removed to its author (gap in the T05 contract).**
   `GET /learn/courses/{id}/review` returns the author's row even after admin removal
   (`GetOwnReview` has no `removed_at` filter), and `reviewDTO` has no `removed` field. The author
   therefore sees a moderated-away review as live, with Edit/Delete. Save then fails with
   `review_removed` (handled), and Delete fails silently (see 4). AC7/G10 imply the author should be
   told. Options: add `removedAt` (or `removed: true`) to the own-review DTO and render «Відгук знято
   модератором» (the locale already has `eligibility.review_removed`), or have `OwnReview` set
   `eligibility = review_removed`. This needs a backend change, so it is a decision for you.
4. **Delete failures are swallowed.** `onPress={() => void onDelete()}` has no catch. A 409
   `review_removed` or a 503 leaves the dialog open with no message. Rt29's "failure keeps state and
   says so" applies to writes in general.
5. **Own-review request failure is not handled (plan §7).** The plan says: "the own-review request
   fails → the review block shows the same error state". `ReviewBlock` treats `own === undefined` as
   loading forever (bare heading), and the route doesn't pass the error. Add an error branch with
   retry.
6. **The counter doesn't match the server's rule.** The client uses `body.trim().length` (UTF-16 code
   units); the server counts runes after trim (`utf8.RuneCountInString`).
   - Ten emoji (20 code units, 10 runes) pass the client and get 400 `review_invalid`, shown as a
     generic failure.
   - Use `[...body.trim()].length` for both the minimum and the maximum, and show the counter on the
     trimmed code-point count.
   - The counter `<output aria-live="polite">` also re-announces on every keystroke, which is noisy.
     Make it non-live and announce only near or over the limit (or tie it to `aria-describedby`).
7. **Hard-coded Ukrainian in the star choices.** `` `${value} з 5` `` in `ReviewBlock`: the en locale
   shows «з 5». Use `t('ratings.stars', { count: value })` (the same regression class Courses 1 fixed in
   T06a).
8. **Duplicate star announcement in one's own review.** `'★'.repeat(stars)` is not `aria-hidden` there
   (the shared list fixed this in T06a), so screen readers hear the stars and then «4 з 5».
9. **Buy entry condition and feedback.** Plan §6 says «Купити курс» shows when `canPurchase` **and no
   lesson is locked for the viewer**. `CourseLearningView` shows it on `canPurchase` alone, so a regular
   student also gets it next to `LockedLessonPanel`'s buy, which duplicates the entry. `onBuy` also has
   no pending or error feedback (double press → `checkout_already_opened`).

## Also fine

- The own-review query key now includes the user id. This closes T06a #3.
- The report error clears on success and the button shows a pending state. This closes part of T06a #4.
- «змінено» with a date in the shared list. This closes T06a #5 there; not in `ReviewBlock`, see 1.
- `learn.reviews(course.id)` uses the learning-scoped endpoint, so archived courses work for grant
  holders. The Rt30 test is present.
- `ReviewModerationRow`: «Зняти» only for an admin (G9); the owner gets dismiss only.
- Rt29: a failed save keeps the typed text and shows `role="alert"`, and `review_removed` is named.
  Tested.
- Rt30 and the shared `Reviews` component: one list implementation for both pages.
