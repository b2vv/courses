# T06b follow-up 2 of 2 — tests only (no production changes unless a test exposes a defect; then say so)

Keep logging to `work/codex-t06b.log`. Do not commit. You cannot run `npm test` here; write the tests anyway,
in the style of the neighbouring test files. Write exactly these; stop only when all exist.

A. `adapters/api` (msw, like `course-repository.test.ts`):
   1. put own review → 201 created and 200 updated map to the review + created flag;
   2. put refusals map each code: `review_owner_forbidden`, `review_purchase_required`, `review_not_started`,
      `review_not_available`, `review_invalid`, `review_removed`;
   3. delete → 204; 409 `review_removed`; 404 `review_not_found`;
   4. learning reviews list maps items + cursor;
   5. teach remove/dismiss review → 204 and 403 `review_moderation_forbidden`; queue review item parses.
B. learning review block (`presentation/learn/ReviewBlock` tests):
   6. no star selected → submit disabled; 19 chars → disabled with the minimum shown; 20 → enabled;
      2000 → enabled; 2001 → disabled; counter «N / 2000» in a polite live region;
   7. submit success shows the review; edit opens the form prefilled;
   8. delete goes through the dialog (no browser confirm) and calls the mutation;
   9. each not-eligible code shows its sentence and no form;
   10. a write failure keeps the typed text and shows the message (Rt29); `review_removed` shows
       «Відгук знято модератором» and no form;
   11. the star radio group has an accessible name and each radio is «N з 5».
C. shared review list (the component used by both pages):
   12. a failed next page keeps the shown items and shows the inline retry; an initial failure shows the
       section error;
   13. the live region announces added / all loaded; after the last page focus is on the first new review,
       not on <body>;
   14. learning list renders for an archived course viewer (Rt30).
D. 15. admin buy entry: shown when `canPurchase` is true, hidden when false; pressing it calls start (spy).
E. teach queue:
   16. a review item renders stars (sr-only «N з 5»), body, report count;
   17. «Зняти» visible for an admin and absent for the course owner; dismiss calls its mutation.

Run `npm run typecheck && npm run lint && npm run format:check`. Final answer: every test name per file.
