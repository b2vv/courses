# C2 — catalog: ui-audit (before the fix)

| | |
|---|---|
| Date | 2026-11-23 |
| Build | `main` @ `a728a564` (`frontend/` identical to `a9bc64a9`, #119). The running frontend image is from the #119 gate on `a5182cbf`. |
| Role · task | Guest. Task (inferred): **"find a course that suits me and open it"**. |
| States | data (`/courses`) light and dark; search with no results (`?q=zzzqqq`); second page (`?limit=2&offset=2`); page past the end (`?limit=2&offset=40`); error (client-side navigation with the API blocked); loading (API delayed 8 s); EN. |
| Viewports | 390×844, 768×1024, 1024×768, 1440×900, 390 with 200% text (simulated), 320×720. The data state at all six; the other states at 390 and 1440. |
| Data | Live seed: 3 published courses, all free, each with `lessonCount: 0` and `rating: null` (`GET /api/v1/courses`). Read-only: only navigation, search and sort were used. |

Boxes are CSS px, `x y w×h` from the top of the document. Captures and metrics live in the session scratchpad (`c2/shots`, `c2/cap.json`, `c2/kbd.mjs`). They are not committed.

## Problems

**#1 · Task · «← Попередня» past the end leads to another empty page**
- sees: on a page past the end of the list (`offset=40`), the page shows «На цій сторінці немає курсів» and «← Попередня». That link goes to `offset=38`, which is empty too, and so on. You cannot get back to the courses by pressing «Попередня».
- costs: a dead end. An old link or a changed filter leaves the user on empty pages.
- proof: click «← Попередня» on `?limit=2&offset=40` → `?…&offset=38`, 0 cards, the same empty state. Code: `CatalogView.tsx:56` `offset - limit`.
- rule: [state] Error states actionable / Nielsen #3 User control.
- fix: past the end, «Попередня» leads to the last page that has data (`floor((total-1)/limit)*limit`), or is not shown, leaving only the reset.
- sure: yes.

**#2 · Words · the visible «Відкрити» button is named «Детальніше: …» for assistive tech**
- sees: the card button says «Відкрити ↗» (or «Купити» for a paid course), but its accessible name is «Детальніше: Go: фундамент надійного бекенду».
- costs: a voice-control user says "click Відкрити" and nothing happens. A screen reader announces a word that is not on the screen.
- proof: Tab stop `Детальніше: Go: фундамент надійного бекенду | A | 126x44` (1440). The visible text is `Відкрити ↗`. Code: `CatalogView.tsx:104` `aria-label={t('catalog.details', …)}`.
- rule: WCAG 2.5.3 Label in Name (A).
- fix: the accessible name starts with the visible word: «Відкрити: <назва>» / «Купити: <назва>».
- sure: yes.

**#3 · Stress · 200% text: the sort list goes past the form edge**
- sees: at 390 with 200% text, the «Спочатку нові» list sticks 20 px past the search block, and the field placeholder is cut («Назва або тем»).
- costs: the control looks broken. The search field is only 260 px wide, because the form adds its own padding inside the page padding.
- proof: `.catalog-search` `32 821 326×508` (right edge 358) vs `.sort-field` `64 1042 314×135` (right edge 378), at 390/200%. The capture shows the cut placeholder.
- rule: WCAG 1.4.10 Reflow / 1.4.4 Resize Text.
- fix: the list takes the form's width (`max-inline-size: 100%`), and the form drops its inner padding at narrow widths.
- sure: yes.

**#4 · Squint/Lines · the price, rating and lesson count each take their own row on a phone**
- sees: under each description come «Безкоштовно», «Новий курс», «0 уроків» and the «Відкрити» button, each on its own row. The card footer is 206 px tall, and each card is 505 px.
- costs: on a phone only one card fits per screen. Comparing courses means scrolling a long way.
- proof: `.course-card-footer` `17 840 356×206` (390). The card is `16 542 358×505`.
- rule: [law] Proximity / [cog] Chunking.
- fix: one meta row «Безкоштовно · 0 уроків · Новий курс» and the button below it. That makes the card about 120 px shorter.
- sure: yes.

**#5 · Words · «Практичний курс» on every card**
- sees: every card has the same caption «ПРАКТИЧНИЙ КУРС» above the description. The page heading also says «НАВЧАННЯ ЧЕРЕЗ ПРАКТИКУ».
- costs: a line that tells nothing and does not separate one course from another. It takes 20 px on every card.
- proof: `.course-card-copy .eyebrow` is the same on all 3 cards (`t('catalog.kind')`, a constant).
- rule: [cog] Reduce noise / [copy] Meaningful labels.
- fix: remove it, or replace it with something real (the level or topic, when the course has one).
- sure: yes.

**#6 · Words · a published course with «0 уроків»**
- sees: all three courses show «0 уроків» and «Новий курс». A guest opens a course and finds nothing to watch.
- costs: the catalog offers courses that have no lessons; trust drops on the first click.
- proof: `GET /api/v1/courses` → `lessonCount: 0` for all 3 (seed). The UI shows the number as it is.
- rule: [state] Honest empty states.
- fix: an owner decision. Either the catalog does not show courses without published lessons, or the card says «Уроки готуються» instead of «0 уроків».
- sure: suspect. This depends on the seed data and on the publishing rule; it is not a UI defect.

**#7 · Words · the empty-state recovery says one thing and does another**
- sees: when a search finds nothing, the page says «Курсів не знайдено» with the link «На першу сторінку», which actually clears the search. Below it the pagination says «0 із 0». Past the end of the list there are three exits at once: «На першу сторінку», «0 із 3» and «← Попередня» (#1).
- costs: the user wants to clear the search but is offered "first page". Several exits compete with each other.
- proof: «На першу сторінку» → `/courses` (field empty, 3 cards). `.catalog-pagination` «0 із 0» under the empty state (390, `?q=zzzqqq`).
- rule: [copy] Labels describe the result; [cog] One clear way out.
- fix: for a search: «Скинути пошук». Past the end: «До останньої сторінки». No pagination when there are no results.
- sure: yes.

**#8 · Twins · the empty-state recovery is a link, the error recovery is a button**
- sees: «Спробувати ще раз» (error) is a filled button, as decided in G0. «На першу сторінку» (empty) is an underlined text link.
- costs: one recovery action looks different in two places.
- proof: `.ui-state > .ui-button` (error) vs `.ui-state a` with no class (empty), 390 and 1440.
- rule: [consistency] Same role, same treatment.
- fix: the empty-state recovery is drawn as a secondary button.
- sure: suspect. A link may be enough for an empty result; the owner decides.

**#9 · Groups/Twins · the course counter is far from the heading and is repeated**
- sees: on a desktop the «3 курси» pill sits on the right edge (x 1219), away from the «Каталог курсів» heading. Below the list the same number appears again as «1–3 із 3». Past the end of the list the pill says «3 курси», but the page shows none.
- costs: the number is easy to miss at the top, and the page states it twice.
- proof: `.catalog-count` `1219 319 77×28` (1440). `.catalog-pagination` «1–3 із 3».
- rule: [law] Proximity.
- fix: one counter next to the heading or above the list («3 курси»), with the range only when there is more than one page.
- sure: yes.

**#10 · Stress · during loading and after an error, the page loses its heading and search**
- sees: while loading there are only five grey lines, with no «Каталог курсів» heading, no search and no card outlines. After an error there is only the error box: the page has no `h1` and no search, so you cannot change the query.
- costs: during loading the page jumps from lines to cards. After an error the user does not see where they are, and the only action is «Спробувати ще раз».
- proof: loading: `h1 = []`, `.ui-skeleton` `16 97 358×144` (390). Error: `h1 = []`, state `144 165 1152×174` (1440).
- rule: [state] Skeleton matches layout / WCAG 2.4.6 Headings (best practice).
- fix: the heading and search stay in place, and only the list area shows the skeleton (card shapes) or the error.
- sure: yes.

**#11 · Task · changing the sort does nothing until «Знайти»**
- sees: choosing «За назвою» changes nothing until you press «Знайти».
- costs: the user changes the sort, sees the same list and decides sorting does not work.
- proof: `selectOption('title')` → URL and order unchanged; after «Знайти» → `?sort=title`.
- rule: [forms] Immediate feedback for filters.
- fix: apply the sort on change (the search query stays until «Знайти»).
- sure: suspect. A form that submits on «Знайти» is a deliberate pattern that works without JS.

## Taste

- The three features («Практика / Транскрипти / Обговорення») repeat at the bottom of the catalog, as on the home page.
- On a desktop the grid has 4 columns, so with 3 courses the right column is empty. With more data this is not noticeable.

## Action inventory (guest, 1440; 390 the same with «Меню»)

```
skip «До основного вмісту» · go to content · own label · focus 3px · Tab #1 PASS
header (wordmark, «Каталог», «Увійти», UA/EN) · checked in G0 · Tab PASS
search field «Знайти курс» · query · label above the field · 869×44 · Tab PASS · Enter → ?q=React, 1 card PASS; focus → BODY afterwards
«Сортування» · list order · label · 168×44 · Tab PASS · change without «Знайти» does nothing (#11)
«Знайти» · submit the form · own label · 83×44 · Tab PASS · → ?sort=title PASS
card title (link) · open the course · in the card · 214×46 · Tab PASS
«Відкрити ↗» · open the course · in the card footer · 126×44 · Tab PASS · → /courses/go-backend-foundations PASS · name ≠ visible text (#2)
«Наступна →» · next page · pagination · Click → offset=2 «3–3 із 3» PASS; focus → BODY
«← Попередня» · previous page · pagination · past the end → empty page again (#1)
«На першу сторінку» (empty) · clear the search · empty state · → /courses, 3 cards PASS; label (#7)
«Спробувати ще раз» (error) · reload · error state · button · NOT_TESTED activation (API blocked in the test)
footer wordmark, «Каталог» · checked in G0 · Tab PASS
```

The Tab order matches the visual order. Each card has two stops (the title and the button) that lead to the same address.

## Not looked / NOT_TESTED

- **Paid course card** («1 890 ₴ · Купити»): NOT_TESTED, because the current seed has no paid published course. It was seen in S0 (C3-1).
- **Card with a rating:** NOT_TESTED, because no course has a rating.
- **Screen reader announcement** of the result count after a search: NOT_TESTED, with no screen reader in the session.
- **Contrast on the gradient card covers:** axe passed with no `incomplete` items, but axe does not measure gradients reliably (gotcha 2026-10-10), so the covers are treated as NOT_TESTED.
- **Real browser zoom:** NOT_TESTED, 200% is simulated.
- **Student:** not looked. The catalog for a student differs only in the header (G0).

## Owner answers

2026-11-23, owner: «Та все треба fix · 6 я б зробив скоро буде але з поправкою якщо пройшов n період тоді ховати · 7 не може бути сторінок, або активних сторінок якщо нема результату». In answer to the follow-up question: N = **30 days**, and the hiding goes into **its own slice** (backend, seed, e2e), right after C2.

| # | Answer | Date |
|---|---|---|
| 1–5, 8–11 | fix (all of them, including suspects #8 and #11) | 2026-11-23 |
| 6 | a course with no lessons shows «Скоро буде»; once 30 days have passed since publication it is hidden from the catalog. The hiding is a separate slice. | 2026-11-23 |
| 7 | no pagination and no page links when there are no results | 2026-11-23 |
