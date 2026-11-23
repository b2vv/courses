# C3b: React Aria links through the router; course page states

| | |
|---|---|
| Date | 2026-11-23 |
| Branch | `design-pass-c3b`, root `cbad6649`. Plan `9138f963` → `d9640454` (rev. 5, accepted). Code `663b2179`. e2e `3811fab3`. Status-page failure `00dcdbbe`. Hyphenation `50f8af83` |
| Contract | `plan.md` §17, revision 5, accepted after the planning critique. Source: audit C3 #11 (`evidence/C3-audit.md`) |
| Build checked | frontend image `98c576944270` on Kiri with the e2e compose configuration (`environment.sh`) |
| Data | Seed course `go-backend-foundations` for the browser captures. e2e fixtures: three courses found by one search token (`catalog.spec.ts`) and a free course with one lesson (`enrollment.spec.ts`). Fixtures are deleted: the catalog ones in each test's `finally`, the enrollment ones in the spec's `afterAll` |

## Result against acceptance

| Item | Before | After | Status |
|---|---|---|---|
| C3b-1 (#11) | Every React Aria link loaded a new document: catalog ↔ course, pagination, «Почати навчання», the 404 way back | Guest: catalog pages, card title, card «Відкрити», «← Усі курси», Back, and «Перейти до каталогу» from the course 404 make 0 document loads after the first. Administrator: «Почати навчання» → `/learn/<slug>` makes 0. A Meta/Ctrl press and a middle press open a new tab, and the page and its URL stay, with 0 loads. Probe on image `d63ea8a7be61`, 10 fresh contexts each: React Aria «← Усі курси» modifier 10/10 and middle 10/10; a plain anchor (root 404) modifier 10/10 and middle 10/10. The e2e test asserts the tab. The home `#values` link scrolls with 0 loads | PASS |
| C3b-2 (R2) | — | Client → `/admin` («В адмінку»): exactly 1 document load. Editor «Перегляд» → course: 1. Inside admin («Учні»): 0. «На клієнт» → `/`: 1 | PASS |
| C3b-3 (R3, R5) | The course route failure was unreachable, and its texts were the catalog's | A client move to a course whose read answers 503 shows the course failure page. It has h1 «Не вдалося завантажити курс», a hint, the primary action «Спробувати ще раз» and the secondary «Перейти до каталогу». The retry is exactly 1 document load and recovers. An answer that fails validation gives the same page (unit; and seen live, when a mocked 404 without `requestId` was treated as invalid) | PASS |
| C3b-4 (R5) | Unreachable; the text said «Завантажуємо курси» | A slow course read (2 s) shows the status «Завантажуємо курс», then the course | PASS |
| C3b-5 | — | A client move to a missing course gives h1 «Курс не знайдено» and the tab title «Сторінку не знайдено — CourseForge». The same holds after a direct load: the server answers 404 with `<title>Сторінку не знайдено — CourseForge</title>` (curl) | PASS |
| C3b-6 (R4, R4a) | Focus was lost on `body` after every move inside the app (the TanStack header links too) | Enter on «← Усі курси» puts focus on `#main`, and the next Tab stays inside `#main`. Two pagination presses: «Наступна» from a middle page keeps focus. Onto the last page «Наступна» is gone and focus goes to `#main`; «Попередня» onto the first page also goes to `#main`. On a slow list (2 s), the pending form is `inert`: the field cannot be focused, and the status «Завантажуємо курси» is in the ARIA snapshot. Focus goes to `#main` when the list arrives. If the user moves to the header «Каталог» while the list loads, focus stays there | PASS |
| C3b-7 | — | The tab title follows each move: course → «<назва> — CourseForge», catalog → «Каталог курсів — CourseForge». The menu card closes (existing `onBeforeNavigate`) | PASS |
| C3b-8 (R6) | A paid order left a cached «not a member» | A move into `paid` invalidates the order's course membership (`['enrollment','membership',courseId]`, every viewer of that course) and `['enrollment','myMemberships']` once; another course's membership is untouched. A page opened on an already paid order does too. `paid` → `paid` does not repeat it (unit) | PASS |

## Found by the code review and fixed

A move inside the app keeps the query cache that a full page load used to drop. Lesson links in the learning outline used to load a document, and now they move inside the app. Two things followed from that.

- **The lesson player resumed from a cached position.** `LessonPlayer.tsx:168-172` resumes from the cached lesson's `position` and continues its `clientSeq`. Now every entry into `/learn/$` (the course and each lesson) and into `/learn` reads past the cache (`ensureRouteQuery(…, { fresh: true })`, `staleTime: 0`), with no exemption. Two exemptions were tried first: a 2 s window, then a history-state mark on the «continue» redirect. Each let a fast move or Back resume from a position read before the last heartbeat (review rounds 3 and 4). The cost is one more course read on the `/learn/<slug>` redirect.
- **The next lesson showed old completion and review eligibility.** A `nav` or `ended` heartbeat now invalidates the learning course and `['learn','review',courseId]` (`refreshAfterHeartbeat`). The next lesson, already mounted, then shows completion and the review form. Its player is not reset: `clientSeq` and the resume position are read once, at mount.
- **Tests.** `route-guards.test.tsx`: every lesson entry reads the course, including a real history Back, and a `nav` refresh reaches the mounted lesson. `learn-queries.test.ts`: which heartbeat reasons refresh. Mutation checks: with no fresh read on a lesson entry, with a redirect exemption, or with a 2 s window, the route test fails.
- **Not fixed (same kind as before C3b).** If lesson A's own `nav` heartbeat is still on its way when A is entered again, the fresh read can return the position from before that heartbeat. A full page load had the same race. This is not a beacon on unload: none exists in the code (review rounds 4 and 5). It is a new page's read racing an unsent or unfinished heartbeat. Review round 5 found no mechanism that makes it larger.

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

- **Unit:** 1585 passed, the whole frontend, on `165197ae` (1578 on `663b2179`).
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
- **New tab:** one earlier run had no tab after a Meta-click. The probe above then measured 40 of 40, and the e2e test asserts the tab again, for both modifier and middle presses (coverage review F2).
- **Ordered `make check`, attempt 1** on `c5bdd4bb`, 12:49–13:26 UTC: **GATE_EXIT=2**.
  - Unit 1585, Go `ok` lines 130, smoke PASS, B19 7/7.
  - e2e: 164 passed, 1 failed, 12 did not run.
  - The failure was the P03 row in `playback.spec.ts:1679`. It holds «Почати навчання» to one learning-course read, and the unconditional fresh read on the redirected lesson made it two.
  - Fixed in `f45bf7ea`: the redirect names its read (`dataUpdatedAt`) in history state, and the lesson reuses the cache only while it still holds exactly that read, not invalidated. Route tests: one read on the redirect; Back to the marked entry after another lesson's read reads again; after a `nav` invalidation it reads again. Each test fails when the mark or the comparison is removed.
  - Targeted rerun on image `d63ea8a7be61`, `-g "P02 P03|C3b"` with its dependency projects, after `make seed-identities`: 148 passed, B19 7/7.
- **Ordered `make check`, attempt 2:** pending.

## Independent review

| Stage | Reviewer | Verdict | Record |
|---|---|---|---|
| Plan rev. 1 | 4 lenses, separate Codex processes, + defense | 3 blocking + 2 defense gaps | `_wip-review-c3b-lens-*.md`, `_wip-review-c3b-defense.md` |
| Plan rev. 2 | logical-gaps, edge-cases, defense | 2 blocking; defense 0 | `*-r2.md` |
| Plan rev. 3 | the same three | 1 shared finding (the pending field disappears) | `*-r3.md` |
| Plan rev. 4 | the same three | 1 shared finding (`inert` hides the status) | `*-r4.md` |
| Plan rev. 5 | the same three | **ACCEPTED**, 0 gaps | `*-r5.md` |
| Code `8d864fd4` | Codex | BLOCK — P2 learning cache kept by moves inside the app (old resume position and heartbeat sequence); P2 full e2e not run | `_wip-review-c3b-code.md` |
| Fix `fe6f5b7c` | Codex | BLOCK — P2 the last `nav` heartbeat lands after the next lesson read the course; P2 review eligibility stale; P3 double read on `/learn/<slug>` | `-r2.md` |
| Fix `8fe63e72` | Codex | BLOCK — P2 the 2 s reuse window resumes a fast A→B→A from an old position | `-r3.md` |
| Fix `294ec245` | Codex | BLOCK — P2 the redirect mark is reused on Back within its window | `-r4.md` |
| Fix `165197ae` | Codex | **PASS** | `-r5.md` |

## Not done here

- The workspace → client move made by signing out (`use-sign-out.ts`, `navigate({ to: '/courses' })`) is a router call, not a link. It stays a move without a page load, as before C3b. R2 covers links only.
- C1 (home page) stays deferred. Its `#values` link was only checked for no page load.
