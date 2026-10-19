# T06b carry-over — from the T06a review (Courses 2, work/review-ratings-t06a.md)
Apply in the shared review list component T06b extracts (used by the course page and the learning page):
1. DEFECT: show the section error only when there is no data; `isFetchNextPageError` (TanStack v5 sets
   `isError` for a failed next page too) keeps the shown items and adds an inline retry that refetches only the
   next page. Test: a failed «Показати ще» leaves the 20 items visible + inline retry.
2. DEFECT: keep the polite live region mounted outside `hasMore ? … : null`; announce «Завантажено ще N
   відгуків» / «Усі відгуки показано»; when the last page arrives and the button unmounts, move focus to the
   first newly added review (or the list). Test: focus is not on <body> after the last page; announcement text.
3. Own-review query key includes the signed-in user id (A→B switch in one tab).
4. Clear `reportError` on a later success; the report button has a pending label/`aria-busy` while sending.
5. «змінено» shows the edit date («змінено 2 жовт.») — plan Rt03 says date.
