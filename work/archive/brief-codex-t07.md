# Codex brief — ratings T07: browser e2e `frontend/e2e/ratings.spec.ts` + axe

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first (they bind): `AGENTS.md`
(the Decisions section: whoever adds something globally counted cleans up after itself; readers scope their
selection), `docs/ratings/spec.md`, `docs/ratings/plan.md` §11 (the e2e column), and the existing e2e as the
pattern — especially `frontend/e2e/purchase-wayforpay.spec.ts` (`createPaidCourse`, run-scoped namespace,
`afterAll` archiving every course it published, axe helper, offline WayForPay stub), `purchase.spec.ts`
(admin grant), `discussion.spec.ts` / `instructor.spec.ts` (moderation queue), `auth.ts` (storage states for
student / instructor / admin), `recovery.spec.ts:91` and `catalog.spec.ts:52` (they count published courses —
yours must not change those counts after your run). Log each step to `work/codex-t07.log` (first line before the
first tool: `NEXT:`). Do **not** commit. You cannot run the browser suite (Docker); write it so Courses 1 can run
it — every test independent, run-scoped data, no reliance on test order across files.

**Write the spec only — one file plus small helpers if needed. Do not change production code.** If a scenario
cannot be driven through the UI/API as it stands, write it as `test.fixme` with one line saying why, and list it
in your final answer.

## Data each test needs (create it run-scoped, clean it up)
- Courses via the admin API as `createPaidCourse` does (free = no price). Lessons with progress: find how
  existing e2e give a student progress (playback heartbeat/progress API, or the SQL path other specs use via
  `docker compose exec`); a lesson counts as **completed** only at ≥95 % of a ready video, so for the
  percentage use whatever existing fixture videos/progress paths allow, and keep the asserted percent consistent.
- A paid course's student right: admin grant API (as `purchase.spec.ts`).
- `afterAll`: archive every course this file published (reviews stay but archived courses are not counted).

## Scenarios (plan rows) — one `test` each
- Rt01: a student with progress on a free course writes 4★ + a 25-char text on the learning page → it appears on
  the course page with «пройдено N% на момент відгуку».
- Rt03: edit 4★→2★ → course page shows 2 stars and «змінено <date>».
- Rt04: delete through the dialog → gone, count drops; write again → appears.
- Rt05: an admin on a paid course without a grant sees «Купити курс» on the learning page; complete the purchase
  with the offline WayForPay stub exactly as `purchase-wayforpay.spec.ts`; with progress the admin can review.
- Rt06/Rt07/Rt08: a course with 2 reviews shows «Новий курс» on the catalog card and the course page; after a 3rd
  (4,5,5) both show `★ 4,7 (3)` — select the card **by this run's course title**, never by position.
- Rt09/Rt22: an admin removes one of the 3 in the teach queue → the review leaves the course page and the card
  falls back to «Новий курс»; the course owner (instructor) sees no «Зняти» (Rt21).
- Rt14: 25 reviews (create via API with 25 run-scoped students, or the fastest existing path) → 20 shown,
  «Показати ще» loads 5 more, newest first; **keyboard**: press «Показати ще» with Enter until the end and assert
  `document.activeElement` is not `body` after the last page and the polite live region text says all shown.
- Rt26/Rt30: archive a course with reviews → a grant holder still sees the reviews on the learning page; the public
  course page is 404; re-publish → the card shows the rating again.
- Report: a signed-in non-author reports → «Скаргу надіслано»; the item appears in the owner's teach queue.
- Cross-origin (left from T05): a review PUT with a foreign `Origin` header is refused (403) — via
  `request.put` with that header.
- a11y: axe (`wcag2a`, `wcag2aa`, `wcag21aa`) on the course page with reviews, the learning-page review block
  with the form open, and the teach queue with a review item — no violations.

## Verify
`cd frontend && npm run typecheck && npm run lint && npm run format:check` (the e2e folder is type-checked).
Final answer: the test names, any `test.fixme` with reasons, what you ran.
