# T05 follow-up 2 — only two things left; finish both before stopping

Keep logging to `work/codex-t05.log`. Do not commit. Nothing else is in scope this round.

## A. HTTP contract tests — `backend/internal/discussion/adapters/http/handler_test.go`

Use the file's existing fake repository and `assertContract` helper. One test function per route is fine.
Each bullet is one assertion you must have:

`PUT /api/v1/learn/courses/{courseId}/review`
- stars 3 + body of exactly 20 runes after trim → 201, body matches the OpenAPI `Review` schema
- body 19 runes → 400 `review_invalid`; body 2000 → 201; body 2001 → 400 `review_invalid`
- stars 0 → 400; stars 6 → 400
- actor is the course owner → 403 `review_owner_forbidden`
- facts with no progress → 403 `review_not_started`
- paid course, not entitled → 403 `review_purchase_required`
- second PUT by the same author → 200 (update), the fake holds one review
- the author's review is removed → 409 `review_removed`
- missing session → 401 (as other learn routes); cross-origin → 403 (sameOrigin)

`DELETE /api/v1/learn/courses/{courseId}/review`
- author with a review → 204; removed review → 409 `review_removed`; no review → 404

`GET /api/v1/learn/courses/{courseId}/review`
- returns `{review: null, eligibility: "review_not_started"}` for a user without progress; a review +
  `"ok"` for an eligible author — both matching the schema

`GET /api/v1/courses/{slug}/reviews`
- no session cookie → 200 (public); 0 reviews → `items: []`, `rating: null`
- 25 reviews → 20 items + `nextCursor`; the cursor request returns the remaining 5; newest first
- 3 reviews (4, 5, 5) → `rating.average` is a JSON number equal to 4.7, `rating.count` 3
- unknown or archived course → 404

`POST /api/v1/learn/reviews/{reviewId}/reports` — 204; own review → 403 `report_own_review`; repeat → 204

`POST /api/v1/teach/reviews/{reviewId}/remove`
- admin → 204; admin again on the removed review → 204; course owner → 403 `review_moderation_forbidden`;
  student → 403

`POST /api/v1/teach/reviews/{reviewId}/dismiss-report` — owner → 204; unrelated student → 403

## B. Integration — `backend/internal/discussion/adapters/postgres/repository_integration_test.go`

A subtest that inserts into `discussion_moderation_log` (1) a row with both `comment_id` and `review_id`
and (2) a row with neither, and asserts each fails with a CHECK violation (SQLSTATE 23514). Use ids that
exist in the ratings fixture; leave no rows behind.

Run `go vet ./internal/discussion/...` and `go test ./internal/discussion/...` (the integration one needs
Docker — say if denied). Final answer: list every test name added; anything from A or B you did not add,
with the reason.
