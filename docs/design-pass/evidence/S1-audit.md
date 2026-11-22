# S1 — ui-audit + interface-review (diff mode)

| | |
|---|---|
| Date | 2026-11-22 |
| Artifact | design-pass S1 (#116), `main` @ `ab29bace`; base for the diff is `254c47be` |
| Build | Kiri, `http://localhost:8080`. The frontend container runs image `bf831020…`, the image from gate 4. Its frontend tree matches `dc0a8021` (`git diff --stat` is empty). |
| Data | Seed plus fixture accounts `frontend/e2e/auth.ts`. Read-only: nothing that saves or sends was activated. |
| Captures | 27 surfaces (`s1/surfaces.json`) × phone 390, phone 390 with 200% text, 320, tablet 768, desktop 1440, desktop with 200% text. That is 162 full-page captures, plus point checks at 320 with 200% text, keyboard and element boxes. Scripts and captures are in the session scratchpad, not the repository. |
| Role · task | Several roles (guest, student, admin, instructor). The S1 task is inferred: **"the shared primitives do not break text, the face or the tab title on any surface"**. |
| Lenses | ui-audit (eight looks), anti-slop-design (visual), ux-review (usability/WCAG-UX). Merged: one defect is one finding. |

200% text is simulated with `html{font-size:200%}`; real browser zoom was not tested. Boxes are CSS px, `x y w×h`, measured from the top of the document.

## S1 regressions (inside the diff)

**#1 · Major · Words/Stress · wordmark in the header, 320 px with 200% text**
- sees: the brand name is cut off to «CourseFor».
- costs: the visible name of the product is lost on a phone with large text. The accessible name «CourseForge — головна» survives.
- proof: `a.wordmark` box `32 32 256×88`. `scrollWidth` 345 > `clientWidth` 256. S1 removed `anywhere`, which worsened the clipping by about 25 px (Codex round 2). Capture `hdr-320-200.png`.
- rule: [mobile] No 2D scroll (Reflow) / WCAG 1.4.4 Resize Text, "without loss of content".
- fix: in S-2, let the wordmark text shrink or move under the icon at narrow width with large text, for example a `clamp()` font size for `.wordmark` or a `flex-wrap` header.
- sure: yes · lens: ux-review.

**#2 · Minor · Words · tab title on a missing course (C3-404)**
- sees: the page says «Курс не знайдено», but the tab reads «Курс — CourseForge».
- costs: tabs and screen readers present the page as a valid course.
- proof: C3-404 `document.title` vs the heading `main h2` `16 264 358×30`.
- rule: WCAG 2.4.2 Page Titled.
- fix: the `courses.$slug` head on notFound/empty `loaderData` should return `pageTitle.notFound`, not the generic `course`.
- sure: yes.

**#3 · Minor · Words · the title of the course outline `/learn/<slug>`**
- sees: the tab says «Навчання» and does not name the course.
- costs: with several open tabs, courses cannot be told apart.
- proof: L2-outline title. S1 evidence records this as a limitation, because `/learn` also matches `learn.$`.
- rule: WCAG 2.4.2.
- fix: `namedPageTitle(course.title)` in the outline route, matching C3/D2.
- sure: yes.

**#4 · Minor · Twins · titles of the two course editors**
- sees: the admin editor (D2) has the course name as its title. The instructor editor (T1) says «Редактор курсу», and its h1 is «Робоче місце інструктора».
- costs: one job gets two naming schemes. The T1 tab does not say which course is being edited.
- proof: the title/h1 table in the probe (`h1.mjs`).
- rule: [consistency] Same role, same treatment.
- fix: T1 should use `namedPageTitle` like D2.
- sure: suspect. It may be intentional: `/teach` and `/admin` are a product boundary.

**Positive changes from the diff, confirmed:**
- Courier/mono are gone. The KPI and price figures are Inter with tabular figures.
- All 27 surfaces have their own title. Switching language from the keyboard (Space on EN) changes the title on the spot: «Каталог курсів» → «Course catalog».
- The wordmark is 44 px tall, and its focus is visible.
- Horizontal overflow is 0 on all 162 captures.

## Pre-existing debt (outside the diff, kept separate)

**#5 · Major · Stress · h1/h2 split mid-word at 200% on a phone**
- sees: «Навичк|и», «працю|ють», «product|ion» (C1); «фунд|амент», «надій|ного» (C3); «Налашту|вання» (A7).
- costs: the heading is hard to read exactly where the user has enlarged the text.
- proof: 75 breaks in `probe.json`. The base `metrics-before.json` has the same words. On C3 the h1 box is `80 726 230×554`: the hero padding in `rem` doubles to about 64 px per side and leaves a 230 px column.
- rule: WCAG 1.4.4.
- fix (per-screen slices): set hero/card padding with `clamp()`/px rather than `rem`; cap the h1 size by the container (`cqi`); consider `hyphens:auto` with `lang="uk"`.
- sure: yes.

**#6 · Minor · Words · the «Записатися» button, C3 at 200%**
- sees: the label breaks as «Записатис|я».
- costs: the primary action of the page reads as broken.
- proof: `.availability button` `81 3724 228×136`.
- rule: WCAG 1.4.4.
- fix: in the C3 slice, a wider column for the card or `white-space:nowrap` on short CTA labels.
- sure: yes. Already listed in S1 as a remainder.

**#7 · Minor · Groups · C3 "Про курс" → "Відгуки"**
- sees: the «Відгуки» heading sits right under the course text with no gap, on both phone and desktop.
- costs: two different sections read as one block.
- proof: h2 «Відгуки» top = 880, which is also the bottom of the previous section (`144 880 725×32`, 1440).
- rule: [law] Proximity.
- fix: a section gap in `.detail-body`.
- sure: yes.

**#8 · Minor · Stress · D1 course status filter**
- sees: «В архіві» is cut at the edge at 390. At 200%, only 2 of the 4 options are visible, and nothing shows that it scrolls.
- costs: the admin does not see that an "archived" filter exists.
- proof: `.ui-segmented` `16 608 358×54`. `scrollWidth` 362 > 356, `overflow-x:auto`.
- rule: [mobile] Reflow / [nav] hidden options.
- fix: wrap or stack the segmented control at narrow width, or show a visible scroll cue.
- sure: yes.

**#9 · Minor · anti-slop + Lines · C1 hero, the "Популярний курс" card**
- sees: the title «React Architecture Pro» is half covered by the «Ваш прогрес 68%» card. The card shows a course and a progress value that the catalog does not have.
- costs: decoration with invented data, which also clips its own text.
- proof: `.hero-preview-card` `80 897 292×238` (390).
- rule: anti-slop: decorative fake UI / demo data without a label; ux-review: honest content.
- fix: in the C1 slice, take the data from a real course, or remove the overlap and mark it as an illustration.
- sure: suspect. The owner may want the card as decoration.

**#10 · Minor · Glance · L2 for an enrolled student**
- sees: on the page of a course the student is enrolled in, it says «Курси з'являться тут після відкриття навчального доступу».
- costs: the student concludes they have no access.
- proof: L2-outline, `main h2` `16 316 358×30`. The seed course has no lessons.
- rule: [state] Empty state explains the real reason.
- fix: an empty state for "the course has no lessons yet".
- sure: suspect. This may be a seed artefact.

**#11 · Minor · structure · pages with no h1**
- sees: C3-404, L2-outline and T0 have no `h1`. T1's h1 is the workspace, not the course.
- costs: screen-reader heading navigation has no entry point.
- proof: the h1 column in `h1.mjs` is empty.
- rule: WCAG 1.3.1 / 2.4.6.
- fix: make the page heading an `h1`.
- sure: yes.

**#12 · taste · D5 KPI digits**
- sees: «1» in Space Grotesk tabular figures has a flag and a foot, and reads like «ɿ».
- source: the `[data-surface='admin'] .ui-stat-value` rule existed before S1.
- sure: suspect / taste.

**Known, not duplicated here:** the phone header takes 3–4 rows (S-2).

## Action inventory (S1-changed actions, guest `/courses`, 1440)

```
skip link «До основного вмісту» · go to content · own label · focus visible (outline 3px) · Tab #1 PASS
wordmark «CourseForge — головна» · go to home · aria-label · 44 px tall; focus outline 3px · Tab #2 PASS
«Каталог» · catalog · nav label · not activated (navigation) · Tab #3 PASS
«Увійти» · login · own label · NOT_TESTED (unsafe: leads into the login flow) · Tab #4 PASS
UA / EN · interface language · group label · Space on EN → title and texts in EN, focus stays on EN · Tab #5/#6 PASS
«Спробувати ще раз» (PageError/LoginFailure) · retry · own label · NOT_TESTED: no safe way to trigger a loader error in this sweep (S1 evidence: 182×44) · NOT_TESTED
```

## Gate (diff mode)

- **S1 diff:** blocker 0, **major 1 (#1)**. #1 is recorded as a known P3 (Codex) and handed to S-2, but there is no dated limitation accepted by the owner. **Until the owner accepts it or S-2 closes it, S1 is not READY at the interface gate.**
- **Pre-existing debt:** major 1 (#5). This does not block S1, because the base has the same 75/103 breaks. It goes to the per-screen slices.

## Not looked / NOT_TESTED

- Tablet landscape (1024): not captured.
- Client dark theme: not captured. Admin is dark-only and was captured.
- Real browser zoom and Firefox text zoom: NOT_TESTED, simulated only.
- Contrast: NOT_TESTED. S1 did not change colours, and `measureUI contrast` did not run.
- Lesson, editor and transcript screens (D3/T2/L2-document, M3): not looked. The seed has no lessons.
- The retry button in a real error state: NOT_TESTED.
- Keyboard: only the header and language switcher on `/courses`. The rest is not looked.

## Owner answers

Answer by number: `1 yes`, `4 intended`, `9 no`, …

| # | Answer | Date |
|---|---|---|
