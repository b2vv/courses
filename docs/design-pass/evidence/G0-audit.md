# G0 — client shell: ui-audit (before the fix)

| | |
|---|---|
| Date | 2026-11-22 |
| Build | `main` @ `78542a14`. The frontend image is from the #117 gate (`8be1c3d1…`). |
| Role · task | Guest and student. Task (inferred): **"get to the content and the main sections in one move"**. The header shows where I am and lets me sign in or out and switch language; the 404 and error pages return me to work. |
| States | Guest `/courses`; student `/learn`, `/courses`; 404 `/no-such-page-dp` for guest and student; error: client-side navigation to `/courses` with the API blocked in the test browser. |
| Viewports | 390×844, 768×1024, 1024×768, 1440×900, 390 with 200% text (simulated), 320. |
| Data | Seed plus fixture accounts. Read-only: only navigation and the keyboard were used. |

Boxes are CSS px, `x y w×h` from the top of the document.

## Problems

**#1 · Stress · phone header**
- sees: the student's header takes 258 px on a phone (a third of the screen), 310 px at 320, and **699 px at 200% text**, which is 83% of the screen. The guest's takes 206 and 357. The content starts below the first screen.
- costs: on a phone with large text the user scrolls past the menu just to reach the page; the content is not visible.
- proof: `.site-header` `0 0 390×258` (student 390), `0 0 390×699` (student 390/200%). Desktop and 1024: 85.
- rule: [mobile] Resize text 200% / [cog] Progressive disclosure.
- fix: on a narrow screen, one row with the wordmark and a «Меню» button; navigation, account and language go into a disclosure panel.
- sure: yes. This is S0 S-2.

**#2 · Words/Stress · wordmark at 200% text**
- sees: «CourseForge» does not fit at 390 and 320 with 200% text; the text is cut by 19 px.
- costs: the visible product name is lost.
- proof: `a.wordmark` `32 32 326×88`, `scrollWidth − clientWidth = 19`.
- rule: WCAG 1.4.4 Resize Text.
- fix: in the compact header the wordmark text gets the whole row except the menu button; at narrow width with large text, a smaller `clamp()` size.
- sure: yes. This is S1-audit #1.

**#3 · Twins · the student's name in the header**
- sees: «Студентка Курсу» is a grey pill with a border, 142×44, the same size and shape as the «Акаунт» and «Вийти» buttons next to it, but nothing happens when you click it.
- costs: it reads as a button; the user clicks and gets no response.
- proof: `span.header-identity-name` `16 137 142×44`, `border-radius: 999px`, `cursor: auto`, not focusable (the Tab walk skips it).
- rule: [consistency] Distinguishable roles.
- fix: show the name as plain text, in a caption and without a frame, or as the label of the account section in the menu.
- sure: yes.

**#4 · Groups/Lines · header on a 768 tablet**
- sees: for the student, «Студентка Курсу / Акаунт / Вийти / UA EN» moves to a second row on the right (109 px), separate from the logo and navigation.
- costs: two rows on a tablet. The account actions sit away from the navigation, and the row does not read as one header.
- proof: `.site-header` `0 0 768×109` (student), 85 for the guest.
- rule: [law] Proximity.
- fix: the compact header from #1 applies up to the width where the full row fits.
- sure: yes.

**#5 · Twins · two footers**
- sees: on the home page, the catalog and the catalog error page the footer is large (logo, tagline, «Навчання» column). On the 404 page and on `/learn` it is one plain line.
- costs: the same site gets two different finishes. On the 404 page the footer leads nowhere.
- proof: `footer.site-footer--storefront` `0 1460 1440×256` vs `footer.site-footer` `144 818 1152×82`.
- rule: [consistency] Same role, same treatment.
- fix: one footer on every client page.
- sure: suspect: the two footers may be intentional (a storefront and a learning area). This is S0 S-6.

**#6 · Twins · «Спробувати ще раз»**
- sees: when the catalog fails to load, «Спробувати ще раз» is a text link. On the general error page (root PageError) the same action is a filled button.
- costs: one action looks different in two places.
- proof: catalog error at `/courses` (API blocked), a link under the text; `__root.tsx` PageError `ui-button ui-button--primary`.
- rule: [nav] Consistent identification / WCAG 3.2.4.
- fix: one control for retry (a button).
- sure: yes.

**#7 · Squint · 404 page**
- sees: a large «Сторінку не знайдено» and below it only a small link «Перейти до каталогу». For a signed-in student there is no route to «Моє навчання».
- costs: the only way back is barely visible, and the student is sent to the catalog instead of their courses.
- proof: 390: link `164×44` at y 325; desktop: link 11 px tall in the shot.
- rule: [state] Error states actionable.
- fix: the recovery action as a button, plus «Моє навчання» for a signed-in user.
- sure: yes.

**#8 · Words · storefront footer for a guest**
- sees: the footer always has «Моє навчання», even for a guest without an account.
- costs: a guest follows it and lands on the sign-in form.
- proof: `footer.site-footer--storefront nav a` «Моє навчання», guest `/courses`.
- rule: [copy] Predictability.
- fix: show «Моє навчання» only to a signed-in user, as the header does.
- sure: suspect.

## Taste

The 404 page at 1440 is mostly empty space (a heading and a link at the top of a blank page). Not a defect.

## Action inventory (guest and student, 404 page, 390)

```
skip «До основного вмісту» · go to content · own label · focus visible 3px · Tab #1 PASS
wordmark «CourseForge — головна» · home · aria-label · 180×44, focus 3px · Tab PASS
«Каталог» · catalog · nav · 88×44 · Tab PASS (not activated)
«Моє навчання» (student) · my courses · nav · 137×44 · Tab PASS (not activated)
«Увійти» (guest) · sign in · own label · 78×44 · Tab PASS · NOT_TESTED (unsafe: leads into the sign-in flow)
«Акаунт» (student) · account settings · own label · 83×44 · Tab PASS (not activated)
«Вийти» (student) · sign out · own label · 74×44 · Tab PASS · NOT_TESTED (unsafe: ends the fixture session)
UA / EN · language · group · 44×44 · Tab PASS (activated in S1: title and texts switch)
«Перейти до каталогу» · return from 404 · own label · 164×44 · Tab PASS
«Студентка Курсу» · — (not an action) · — · not focusable · see #3
footer links «Каталог», «Моє навчання» · navigation · «Навчання» column · NOT_TESTED keyboard (not reached in this walk)
```

## Not looked / NOT_TESTED

- **Contrast:** NOT_TESTED in this sweep. The colours are the same as S1, where axe found no violations on the client pages.
- **Root PageError (`__root.tsx`):** not reached. The catalog failure is handled by its own route boundary (#6). The root error appears when a route without its own boundary fails; there is no safe trigger here.
- **Email verification banner and sign-out failure toast:** not looked. Fixture accounts are verified, and the toast needs a failing sign-out.
- **Client dark theme:** not looked in this sweep.
- **Real browser zoom:** NOT_TESTED, 200% is simulated.

## Owner answers

| # | Answer | Date |
|---|---|---|
| 1 | yes — compact header on a narrow screen | 2026-11-22 |
| 2 | the logo should be smaller | 2026-11-22 |
| 3 | account and sign-out need a different style | 2026-11-22 |
| 4 | the menu opens as a card fixed to the right side with a 32 px margin, not a sliding sheet | 2026-11-22 |
| 5 | bug — one footer | 2026-11-22 |
| 6 | one view, with a retry button | 2026-11-22 |
| 7 | restyle the 404 page and give it a better recovery | 2026-11-22 |
| 8 | bug | 2026-11-22 |

Owner's words (2026-11-22): «1 yes · 2 logo be smaller · 3 need another style for acc and logout · 4 maybe we use drawer card? Not slider card fixed by right side with margin 32px? · 5 bug · 6 need one view, with button try · 7 change style, not good ui and update case by good ux · 8 this bug». Reading of #4 (author): a dialog card anchored to the top right, inset 32 px, sized to its content; not a full-height panel that slides in.
