# C3 — course page: ui-audit (before the fix)

| | |
|---|---|
| Date | 2026-11-23 |
| Build | `main` @ `0348c4d1` (#121). The running images are from the #120 gate on `1f94eabc`; `git diff 6c74f867 0348c4d1 -- frontend/src` changes only a generated API description string and a unit test. |
| Role · task | Guest and enrolled student. Task (inferred): **"decide whether this course is for me, and start it"**. |
| States | `/courses/go-backend-foundations`: guest and student, light and dark, EN; `/courses/does-not-exist` (404); reviews failed to load (API aborted in the test browser); guest pressing «Записатися». |
| Viewports | 390×844, 768×1024, 1024×768, 1440×900, 390 with 200% text (simulated), 320×720 for guest and student; the other states at 390 and 1440. |
| Data | Live seed: 3 free published courses, each with 0 lessons and no rating. Read-only, except one guest «Записатися» press, which the server refused with 401 and which changed nothing. |

Boxes are CSS px, `x y w×h` from the top of the document. The captures and metrics are in the session scratchpad (`c3/shots`, `c3/cap.json`, `c3/kbd.mjs`, `c3/guest.mjs`) and are not committed.

## Problems

**#1 · Task · a guest who presses «Записатися» gets "try again", with no way to sign in**
- sees: a guest on the course page sees «Готові почати?» and the button «Записатися». Pressing it shows «Не вдалося виконати дію. Спробуйте ще раз.». Trying again gives the same message. Nothing tells the guest to sign in, and nothing leads to the sign-in page.
- costs: the main action of the page is a dead end for every guest, which is exactly the visitor the page is meant to convert.
- proof: the press sent `POST /api/v1/learn/courses/…/enroll` → **401**; the URL stayed the same; the alert in `.availability [role=alert]` showed the message above (390, guest). Code: `CourseView.tsx:141` shows the enrol button to anyone who is not a member, and `enrollErrorKey` maps the 401 to `generic`.
- rule: [state] Error states actionable / Nielsen #9 Help users recover from errors.
- fix: a guest sees «Увійти, щоб записатися» (to `/login` with a return to this course), or the button leads to sign-in and comes back to enrol.
- sure: yes.

**#2 · Words/Task · a course with no lessons promises lessons**
- sees:
  - the catalog card says «Скоро буде», but the course page does not;
  - the sidebar says «Відкриємо перший незавершений урок за програмою курсу» and offers «Записатися» or «Почати навчання»;
  - the page has no programme and does not say why.
  - For an enrolled student, «Почати навчання» leads to `/learn/go-backend-foundations`, which says «Курси з’являться тут після відкриття навчального доступу». The student does have access, and that page has no h1.
- costs: the page promises a lesson that does not exist. The student follows it and is told the wrong reason, as if their access were missing.
- proof: `lessonCount: 0` for all 3 seed courses (`GET /api/v1/courses`). The programme section is hidden when it is empty (`CourseView.tsx:68`). After «Почати навчання» the page has `h1 = []` and the text above (1440, student).
- rule: [state] Honest empty states / Nielsen #2 Match with the real world.
- fix: a course with no lessons says «Скоро буде» (the same as the card) instead of the programme. It offers no "start learning". Whether a course with no lessons can be enrolled in at all is the owner's call.
- sure: yes. The `/learn` text belongs to L1/L2, but this page is the way into it.

**#3 · Twins/Squint · for a student, «Поскаржитися на викладача» outweighs «Почати навчання»**
- sees: «Почати навчання» is a plain underlined link. Right under it, «Поскаржитися на викладача» is an outlined bold button, twice as wide.
- costs: the heaviest thing in the student's box is the complaint, not the next step in learning.
- proof: «Почати навчання» `958 809 135×44`, 16px/400, underline, no border; «Поскаржитися на викладача» `958 853 258×44`, 16px/650, 1px border (1440, student). Known from S0 as C3-2.
- rule: [visual] Primary action most prominent / [consistency] Same role, same treatment (the guest's «Записатися» is a primary button).
- fix: «Почати навчання» becomes the primary button. The complaint becomes a quiet text action below it, set apart from it.
- sure: yes.

**#4 · Words · the page never shows the price; a paid course has no "buy"**
- sees: the catalog card shows «Безкоштовно» (or the price with «Купити»), and the course page shows neither. For a paid course that the visitor has not bought, the code shows «Почати навчання» as a link to `/learn/…`, with no price and no buy action.
- costs: on the page where people decide, the cost is missing. A buyer has no way to buy from it.
- proof: `CourseView.tsx:141-172`: there is no `course.price` output, and a paid non-member falls into the last branch («Почати навчання»). Free course: no price text anywhere in `main` (1440, 390). The paid case is **NOT_TESTED live**: the seed has no paid course. Known from S0 as C3-1.
- rule: [cog] Recognition over recall / [nav] Information scent.
- fix: the sidebar shows the price («Безкоштовно» or «1 890 ₴»). A paid non-member gets «Купити» to the existing checkout.
- sure: yes from the code. The live paid view is NOT_TESTED.

**#5 · Squint/Stress · on a phone the action is below the first screen**
- sees: on a 390 phone, the dark header block takes the top of the screen. Then come the description and the reviews. «Записатися» is only at y 895, below the 844 px screen.
- costs: a phone visitor does not see what to do without scrolling. On a phone the sidebar box becomes the last block of the page.
- proof: the header block's lede ends at y 416 (`p.lede` `36 324 319×92`); `aside.availability` `16 796 358×168`; «Записатися» `41 895 121×44`; viewport 844 (390, guest). At 390/200% the box is near the end of the 3781 px page (capture `guest-p200`).
- rule: [visual] Primary action visible / [law] Serial position.
- fix: on narrow screens the sidebar box (price + action) comes right after the header block, before «Про курс».
- sure: yes.

**#6 · Groups · «Про курс» and «Відгуки» stick together**
- sees: the «Відгуки» heading starts right where the description ends, with no gap. The page reads as one long text.
- costs: the sections are hard to tell apart, and reviews look like part of the description.
- proof: `section` «Про курс» ends at y 877, and the «Відгуки» heading starts at y 877 (1440). On 390: 712 = 712. Known from S0 as C3-4.
- rule: [law] Proximity / Common Region.
- fix: a section gap (the same as between the hero and the body).
- sure: yes.

**#7 · Twins/Words · the header block repeats what C2 removed from the cards**
- sees: «ПРАКТИЧНИЙ КУРС» above every course title, and «Новий курс» in white text where a rating would be. The catalog card no longer shows either (C2 #5, #6).
- costs: the same course is described differently in the catalog and on its own page. «Новий курс» looks like a label rather than "no ratings yet".
- proof: `.detail-heading .eyebrow` = `t('catalog.kind')`, a constant; `p` «Новий курс» `208 529 1024×26` (1440).
- rule: [consistency] Same thing, same words / [cog] Reduce noise.
- fix: remove the constant eyebrow. Show the rating only when there is one, the same as on the card.
- sure: yes.

**#8 · Twins · the course 404 looks different from the site 404**
- sees: «Курс не знайдено» is a small h2 with no h1, and the way out is an underlined text link «Перейти до каталогу». The site 404 (G0) has an h1 and a button.
- costs: two "not found" pages that look different. Screen-reader users get no page heading.
- proof: `/courses/does-not-exist` → 404, `h1 = []`, h2 `144 191 1152×30`, link `144 259 164×44` 16px/400 underline (1440). Known as C3-3 and S-6.
- rule: WCAG 2.4.6 Headings and Labels (best practice) / [consistency] Same role, same treatment.
- fix: use the same StatusPage as G0, with h1 «Курс не знайдено» and the button «До каталогу».
- sure: yes.

**#9 · Stress · 200% text breaks words in the button and the hint**
- sees: at 390 with 200% text, the button reads «Записатис / я» and the hint «незавершени / й», split across lines in the middle of a word.
- costs: the main button looks broken.
- proof: capture `guest-p200`: the button is 2 lines high and the hint is broken mid-word. This is the same class as the gotcha of 2026-11-16 (`overflow-wrap:anywhere`).
- rule: WCAG 1.4.4 Resize Text.
- fix: keep whole words in buttons, and let a long word break only with a hyphen (the TYPE rule from D5).
- sure: yes.

**#10 · Words · «Спробувати ще» next to «Спробувати ще раз»**
- sees: when the reviews fail to load, the button says «Спробувати ще». The page-level error and the other retries say «Спробувати ще раз».
- costs: one action has two names.
- proof: reviews error state `1440`: «Не вдалося завантажити відгуки» + «Спробувати ще». Locales `uk.ts:42,341,542` «Спробувати ще раз» vs `uk.ts:299,366` «Спробувати ще».
- rule: [copy] One name per action.
- fix: «Спробувати ще раз» everywhere.
- sure: yes.

**#11 · Task · every move between the catalog and a course reloads the whole page**
- sees: clicking a course card, «← Усі курси» or the card button loads a new document from the server instead of switching inside the app.
- costs:
  - each move refetches and re-renders everything, so it is slower;
  - the course page's own loading and error states cannot be reached by navigation.
  This is not limited to this page: links drawn with the React Aria `Link` are not connected to the router.
- proof: `request.resourceType() === 'document'` for `/courses/go-backend-foundations` (title click and card button) and for `/courses` («← Усі курси»). No `RouterProvider`/`useHref` in `frontend/src`.
- rule: no rule (performance/consistency).
- fix: connect React Aria to the TanStack router (`RouterProvider` with `navigate`/`useHref`) once, for the whole app. That is a separate shared change with full e2e, not part of C3.
- sure: suspect. It may be deliberate (G0 used plain anchors for recovery on purpose), and it touches every screen.

## Taste

- On a desktop, the dark header block is about 515 px tall in the capture, and its right half is empty.
- A grey rule separates the header block from the body, although the header block already has its own frame.

## Action inventory

Guest, 1440 (390 the same, with «Меню»):

```
skip «До основного вмісту» · go to content · own label · focus 3px · Tab #1 PASS
header (wordmark, «Каталог», «Увійти», UA/EN) · checked in G0 · Tab PASS
«← Усі курси» · back to the catalog · in the header block · 91×44 · Tab PASS · → /courses (full page load, #11)
«Записатися» · enrol in this course · sidebar «Готові почати?» · 121×44 · Tab PASS · → 401, «Спробуйте ще раз» (#1)
footer wordmark, «Каталог» · checked in G0 · Tab PASS
```

Enrolled student, 1440:

```
header («Моє навчання», «Акаунт», «Вийти») · checked in G0 · Tab PASS
«← Усі курси» · as for the guest · Tab PASS
«Почати навчання» · open learning · sidebar «Ви учасник» · 135×44 · Tab PASS · → /learn/go-backend-foundations: «Курси з’являться тут…», no h1 (#2)
«Поскаржитися на викладача» · report the lecturer · sidebar «Ви учасник» · 258×44 · Tab PASS · opens the dialog «Причина скарги / Надіслати скаргу / Скасувати»; Escape closes it and focus returns to the button. «Надіслати скаргу»: NOT_TESTED (unsafe)
```

Reviews error: «Спробувати ще» · reload the reviews · under «Відгуки» · NOT_TESTED activation (API blocked in the test).

The Tab order matches the visual order. Focus is visible (3 px) on every stop. After the footer, focus goes to the page body before it wraps to the top.

## Not looked / NOT_TESTED

- **Paid course** (price, «Купити», the bought state): NOT_TESTED, because the seed has no paid course. #4 rests on the code and the S0 capture.
- **Course with lessons and a programme, and course with a rating and reviews:** NOT_TESTED, because no published course has lessons or reviews. The programme layout (C3-4 in S0) is therefore not rechecked.
- **Course owner (instructor) and administrator:** NOT_TESTED. The stored instructor session had expired (the page showed the guest header). From the code, an instructor who is not a member would see «Записатися» on their own course (`CourseView.tsx:141` excludes only administrators).
- **Blocked and revoked members:** NOT_TESTED, because there is no fixture.
- **Page-level loading and error of the course route:** NOT_TESTED, because they cannot be reached by navigation (#11). Only the server-rendered page and the reviews error were seen.
- **Contrast:** axe (`wcag2a/aa/21aa/22aa`) found 0 violations and 0 contrast `incomplete` items in all 26 captures, including the light lede on the dark header block.
- **Real browser zoom:** NOT_TESTED, because 200% is simulated.
- **Screen reader:** NOT_TESTED.
