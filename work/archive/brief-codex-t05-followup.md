# T05 follow-up from Courses 1 — finish the tests (the slice is not done without them)

Keep logging to `work/codex-t05.log`. Do not commit. You reported the tests as not written; the
acceptance rows have no proof until they exist. Write them now, against the code already in the tree,
and fix the code where a test exposes a defect (say so in the log).

1. **Application unit tests** (`discussion/application`, fake repository like the existing tests): each
   refusal in plan §4 order (anonymous, owner Rt15, not available, purchase required, not started Rt18);
   PutReview creates then updates with progress recaptured (G6) and returns created vs updated; repeat PUT
   keeps one review (Rt19); DeleteReview without eligibility (works after losing access), removed →
   `review_removed`; ReportReview own → `report_own_review`; RemoveReview: owner → `review_moderation_forbidden`
   (Rt21), admin removes, already removed → no-op success, missing → not found; DismissReviewReports
   owner ∨ admin; PublicReviews on an archived course → not found, rating nil below 3.
2. **HTTP contract tests** (`handler_test.go`, `assertContract` against OpenAPI) for every new route and
   the changed queue: success shapes; Rt10 (19/20/2000/2001 runes after trim → 400/201/201/400); Rt11
   (stars 0/1/5/6 → 400/201/201/400); Rt15 403; Rt18 403; Rt19 (second PUT → 200); Rt21 owner remove 403;
   Rt23 (create and delete after removal → 409 `review_removed`); Rt14 (public list 0 items; 25 → 20 +
   cursor → 5, newest first); public list without a session cookie succeeds; archived course → 404;
   `rating.average` is a number; the queue returns both kinds.
3. **Integration** (T03 review №1): moderation-log rows with both targets and with neither are rejected by
   the one-target CHECK. Fixture rules: Argon2id fixture hash for users; one statement per parameterised
   `Exec`; restore any shared row you change in `t.Cleanup`; scope counts to your own rows.

Run what you can (`go vet`, `go test ./internal/discussion/...`, frontend `typecheck`); say what you could
not. Final answer: tests added (names), defects found and fixed, what you ran, what you could not.
