# Codex brief — ratings T06a: catalog card + course-page reviews (public side)

Repository /Users/strelia/projects/courses, branch `feat/ratings`, frontend in `frontend/`. Read first (they
bind): `AGENTS.md` (TypeScript rules, design tokens, a11y), `docs/architecture/frontend.md`,
`docs/ratings/spec.md` (AC8, AC9, AC11, G7, R4), `docs/ratings/plan.md` §6 (frontend) and §11 rows Rt06–Rt08,
Rt14, Rt25, Rt28, `work/_t05-carryover.md` (format the average to one decimal on the client). The backend
and the Orval client are done: `frontend/src/adapters/api/generated/` has `Rating`, `Review`,
`PublicReviewPage`, the catalog `Course.rating`/`lessonCount`, and the public reviews + report operations.
Log each step to `work/codex-t06a.log` (first line before the first tool: `NEXT:`). Do **not** commit.

Scope is the **public side only**: catalog card and course page. Not in this slice (T06b): the learning-page
form, own review, the archived learning list, the admin buy entry, the teach queue.

Follow the existing layering exactly — look at how the catalog already flows before writing:
`adapters/api/course-repository.ts` + `course-schema.ts` (validate external data with the schema; no `any`,
no assertions that disguise input) → `domain/course` → `app/composition/catalog.ts` +
`catalog-queries.ts` (Query, request-scoped) → `presentation/storefront/CatalogView.tsx`,
`CourseView.tsx`; primitives from `presentation/ui/` (React Aria); formatting like `presentation/format-price.ts`
(`Intl`). No fetch in presentation or routes; no server state in Zustand. Write tests first (rstest, the repo's
existing test style next to each file).

## 1. Domain + adapter
- Course (catalog and course page) gains `rating: { average: number; count: number } | null` and
  `lessonCount: number`, validated in the schema (average 1–5, count ≥ 3, lessonCount ≥ 0).
- Review read model: `{ id, authorDisplayName, stars (1–5), body, completedLessons, totalLessons, createdAt,
  editedAt | null }`, validated; `percent` = round half-up of completed·100/total, 0 when total is 0 (same
  rule as the backend `ReviewProgress.Percent`).
- Repository methods: public reviews page by slug + cursor; report a review (POST
  `/learn/reviews/{id}/reports`). Map error codes to domain errors the UI can name (`report_own_review`,
  `authentication_required`, `dependency_unavailable`, not found).

## 2. Catalog card (`CatalogView.tsx`)
- `rating` present → `★ 4.7 (3)`: the average always with **one decimal** in the current locale
  (`Intl.NumberFormat`, min = max fraction digits 1 → «5,0» in uk, «5.0» in en); accessible name
  «Рейтинг 4,7 з 5, 3 оцінки» (uk) / «Rated 4.7 out of 5, 3 ratings» (en).
- `rating: null` → «Новий курс» / «New course».
- Lesson count with correct Ukrainian plural (1 урок, 2–4 уроки, 5+ уроків; use `Intl.PluralRules`), English
  «1 lesson / N lessons».
- No duration (decided — `docs/ratings/assumptions.md`).

## 3. Course page (`CourseView.tsx`)
- The same rating summary (or «Новий курс») near the title.
- A «Відгуки» section below the program: infinite query over the public list (20 per page, newest first),
  «Показати ще» when there is a next cursor; empty state «Ще немає відгуків»; **error state** with
  «Спробувати ще» that refetches only the reviews while the rest of the page stays rendered (Rt28).
- Each review: author display name, stars shown visually and as text for assistive tech («4 з 5»), the body
  (preserve line breaks, no HTML), «пройдено N% на момент відгуку», the date (`format-timestamp.ts`), and
  «змінено» when `editedAt` is set.
- A «Поскаржитись» button for a signed-in viewer who is not the author (use the existing identity queries
  to know who is signed in); after success it shows «Скаргу надіслано» and disables; `report_own_review` is not
  reachable from the UI (button hidden for the author); anonymous → no button.
- i18n: every string in `app/providers/locales/uk.ts` and `en.ts`.
- a11y per AGENTS: keyboard, visible focus, ≥ 44 px targets, headings in order, light/dark tokens, 200 % zoom;
  the reviews list is a list with an accessible label; «Показати ще» announces new items (polite live region)
  without moving focus.

## Tests (required — do not hand back without them)
- schema: valid / invalid rating, review, lessonCount (rejects average 6, count 2, negative lessons).
- formatting: 5 → «5,0» (uk) and «5.0» (en); 4.666 never shown (server sends 4.7); plural 1/2/5/21/22/25/11.
- `percent`: 0/4 → 0, 1/3 → 33, 2/3 → 67, 1/8 → 13, 4/4 → 100, 0/0 → 0.
- CatalogView: rating null → «Новий курс»; rating → text + accessible name; lesson count.
- CourseView reviews: empty state; 20 items + «Показати ще» loads more (Rt14); error state with retry while the
  program stays visible (Rt28); report button visible only for a signed-in non-author; report success state.

## Verify
`cd frontend && npm run typecheck && npm run lint && npm test` (and `npm run format:check`). The repo's
architecture check runs inside `lint`. If a command is denied in your sandbox (e.g. binding a port), say so —
Courses 1 runs the full gate. Final answer: files changed, tests added (names), what you ran with results,
what you could not.
