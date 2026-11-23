# C3b: React Aria links through the router; course page states

| | |
|---|---|
| Date | 2026-11-23 |
| Branch | `design-pass-c3b`, root `cbad6649`. Plan `9138f963` → `d9640454` (rev. 5, accepted). Code `663b2179`. e2e `3811fab3`. Status-page failure `00dcdbbe`. Hyphenation `50f8af83` |
| Contract | `plan.md` §17, revision 5, accepted after the planning critique. Source: audit C3 #11 (`evidence/C3-audit.md`) |
| Build checked | frontend image `98c576944270` on Kiri with the e2e compose configuration (`environment.sh`) |
| Data | Seed course `go-backend-foundations` for the browser captures. e2e fixtures: three courses found by one search token (`catalog.spec.ts`) and a free course with one lesson (`enrollment.spec.ts`). All fixtures are deleted after each test |

## Result against acceptance

| Item | Before | After | Status |
|---|---|---|---|
| C3b-1 (#11) | Every React Aria link loaded a new document: catalog ↔ course, pagination, «Почати навчання», the 404 way back | Guest: catalog pages, card title, card «Відкрити», «← Усі курси», Back, and «Перейти до каталогу» from the course 404 make 0 document loads after the first. Administrator: «Почати навчання» → `/learn/<slug>` makes 0. A Meta/Ctrl press is not taken by the app: the page and its URL stay, with 0 loads. The home `#values` link scrolls with 0 loads | PASS |
| C3b-2 (R2) | — | Client → `/admin` («В адмінку»): exactly 1 document load. Editor «Перегляд» → course: 1. Inside admin («Учні»): 0. «На клієнт» → `/`: 1 | PASS |
| C3b-3 (R3, R5) | The course route failure was unreachable, and its texts were the catalog's | A client move to a course whose read answers 503 shows the course failure page. It has h1 «Не вдалося завантажити курс», a hint, the primary action «Спробувати ще раз» and the secondary «Перейти до каталогу». The retry is exactly 1 document load and recovers. An answer that fails validation gives the same page (unit; and seen live, when a mocked 404 without `requestId` was treated as invalid) | PASS |
| C3b-4 (R5) | Unreachable; the text said «Завантажуємо курси» | A slow course read (2 s) shows the status «Завантажуємо курс», then the course | PASS |
| C3b-5 | — | A client move to a missing course gives h1 «Курс не знайдено» and the tab title «Сторінку не знайдено — CourseForge». The same holds after a direct load: the server answers 404 with `<title>Сторінку не знайдено — CourseForge</title>` (curl) | PASS |
| C3b-6 (R4, R4a) | Focus was lost on `body` after every move inside the app (the TanStack header links too) | Enter on «← Усі курси» puts focus on `#main`, and the next Tab stays inside `#main`. Two pagination presses: «Наступна» from a middle page keeps focus. Onto the last page «Наступна» is gone and focus goes to `#main`; «Попередня» onto the first page also goes to `#main`. On a slow list (2 s), the pending form is `inert`: the field cannot be focused, and the status «Завантажуємо курси» is in the ARIA snapshot. Focus goes to `#main` when the list arrives. If the user moves to the header «Каталог» while the list loads, focus stays there | PASS |
| C3b-7 | — | The tab title follows each move: course → «<назва> — CourseForge», catalog → «Каталог курсів — CourseForge». The menu card closes (existing `onBeforeNavigate`) | PASS |
| C3b-8 (R6) | A paid order left a cached «not a member» | A move into `paid` invalidates `['enrollment','membership']` and `['enrollment','myMemberships']` once. A page opened on an already paid order does too. `paid` → `paid` does not repeat it (unit) | PASS |

## Found and fixed during verification (not in the plan)

- **The failed course page was an inline block with an h2.** The course 404 and the site error page are a `StatusPage` with an h1. The failure now uses the same page, with a retry and a way to the catalog (`00dcdbbe`).
- **At 390/200 % the status page heading broke inside a word** («завантажи / ти»). The status page now uses the C3 #9 / TYPE rule: under enlarged text a word breaks at a syllable with a hyphen («заванта- / жити»). At normal size nothing changes (`50f8af83`). This applies to every `StatusPage`, including the site 404.
- **React keeps the previous catalog page mounted but hidden while the pending frame shows.** So there are two search forms in the DOM during a slow move. The e2e test addresses the pending one by `[inert]`.

## Browser

38 captures on image `98c576944270`:
- course loading, course failure, course 404, catalog pending and site 404, each at 390, 768, 1024, 1440, 390 with 200 % text, and 320;
- dark mode for loading, failure and pending, at 390 and 1440;
- EN failure at 390 and 1440.

axe (`wcag2a/aa/21aa/22aa`) finds 0 violations and 0 contrast `incomplete`. Horizontal overflow is 0 and no text is clipped. Each state was reached by a move inside the app (0 document loads). In the catalog pending state the search field, sort and «Знайти» are `inert`; the h1 and the status are not.

Keyboard: real key presses in Chromium (Playwright `keyboard.press`), see C3b-6. Screen reader: NOT_TESTED. Real browser zoom: NOT_TESTED (200 % is simulated by the root font size).

## Tests

- **Unit:** 1578 passed, the whole frontend, on `663b2179`.
  - `aria-router.test.ts`: the surface of each path, `reloadDocument` across surfaces and for a retry, search and hash kept.
  - `route-guards.test.tsx`, «moves inside the app»: a React Aria link moves through the router and focus goes to `#main`; focus kept on a surviving control; `body` → `#main`; two moves in a row give one focus; the course failure for `dependency_unavailable` and `invalid_response`; the course loading state; the course 404 heading, tab title and way back.
  - `CatalogView.test.tsx`: the course loading text; only the pending form is `inert`; the retry asks for `{ reloadDocument: true }`.
  - `purchase-queries.test.ts`: R6.
- **Mutation check:** with the root layout of `d9640454`, 4 of the 8 route tests fail: the router move, `body` → `#main`, one focus for two moves, and the 404 way back.
- **e2e, live stack:**
  - `catalog.spec.ts`: 4 new C3b tests, as guest; fixtures with three courses.
  - `enrollment.spec.ts`: 1 new admin test, using the saved session.
  - Both files together: 26/26. The four C3b catalog tests passed 3 times in a row after the modifier-press fix.
  - Login budget: unchanged (no new form login; B19 3 in the targeted run).
- **Not asserted:** whether headless Chromium opens a new tab on a Meta-click. It did in two of three runs. The app's part is asserted: it does not take the press.
- **Ordered `make check`:** not run yet.

## Independent review

| Stage | Reviewer | Verdict | Record |
|---|---|---|---|
| Plan rev. 1 | 4 lenses, separate Codex processes, + defense | 3 blocking + 2 defense gaps | `_wip-review-c3b-lens-*.md`, `_wip-review-c3b-defense.md` |
| Plan rev. 2 | logical-gaps, edge-cases, defense | 2 blocking; defense 0 | `*-r2.md` |
| Plan rev. 3 | the same three | 1 shared finding (the pending field disappears) | `*-r3.md` |
| Plan rev. 4 | the same three | 1 shared finding (`inert` hides the status) | `*-r4.md` |
| Plan rev. 5 | the same three | **ACCEPTED**, 0 gaps | `*-r5.md` |
| Code | Codex | pending | — |

## Not done here

- The workspace → client move made by signing out (`use-sign-out.ts`, `navigate({ to: '/courses' })`) is a router call, not a link. It stays a move without a page load, as before C3b. R2 covers links only.
- C1 (home page) stays deferred. Its `#values` link was only checked for no page load.
