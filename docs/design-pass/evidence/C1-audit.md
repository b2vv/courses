# C1 — home page: ui-audit (before the fix)

| | |
|---|---|
| Date | 2026-11-23 |
| Build | `main` @ `a9bc64a9` (#119). The running frontend image is from the #119 gate on `a5182cbf`; its `frontend/` tree is identical to `a9bc64a9` (`git diff a5182cbf a9bc64a9 -- frontend` is empty). |
| Role · task | Guest (and student for comparison). Task (inferred): **"understand what this place teaches and get to a course in one move"**. |
| States | `/` guest light and dark, student light, guest EN; `/?notice=verify-email` (EN, 390). |
| Viewports | 390×844, 768×1024, 1024×768, 1440×900, 390 with 200% text (simulated), 320×720. Student and EN at 390 and 1440 only. |
| Data | The page is static (`frontend/src/modules/catalog/ui/HomeView.tsx`). Read-only: only navigation and the keyboard were used. |

Boxes are CSS px, `x y w×h` from the top of the document. Captures and metrics live in the session scratchpad (`c1/shots`, `c1/cap.json`, `c1/kbd.mjs`). They are not committed.

## Problems

**#1 · Words · the course in the hero does not exist**
- sees: the large card says «Популярний курс · React Architecture Pro», and a card on top of it says «Ваш прогрес 68% · Модуль 7 з 10 · Domain boundaries». The catalog has no such course, the guest has no progress, and nothing marks the cards as an illustration.
- costs: the first thing a guest sees is a made-up course and a made-up figure. Looking for "React Architecture Pro" in the catalog finds nothing, and trust in the other claims drops.
- proof: `HomeView.tsx:36-50` hard-codes the strings. `.hero-art` is `aria-hidden`, so a screen reader skips it, but sighted users read it.
- rule: design-pass §4 "Label demo data as demo"; anti-slop (invented metrics).
- fix: show a real published course (for example the newest or best rated) with its real cover, or a neutral illustration with no names or figures.
- sure: yes. This is S0 C1-2.

**#2 · Stress · the progress card covers the course title**
- sees: «Ваш прогрес» lies over «React Architecture Pro» at every width.
- costs: the title is cut in half. At 200% text it cannot be read at all.
- proof: overlap of `.hero-preview-title` and `.hero-progress-card`: 147×19 (390), 327×13 (768), 182×3 (1024), 215×5 (1440), 113×16 (320), **109×165 (390/200%)**.
- rule: [visual] No overlapping content / WCAG 1.4.4.
- fix: if the cards stay, the progress card must not sit on text: either move it below the title or reserve space for it.
- sure: yes. This is S0 C1-1.

**#3 · Words · low contrast in the progress card**
- sees: the line «Модуль 7 з 10 · Domain boundaries» is pale grey, 11 px, on white.
- costs: hard to read; it fails the contrast minimum.
- proof: axe `color-contrast` (serious) on `.hero-progress-meta`, `#9ca3af` on `#ffffff` = **2.53:1** (needs 4.5:1), 11.2 px. It fails at every viewport in light and dark. axe cannot measure `.hero-preview-label` and `.hero-preview-title` on the gradient (`bgGradient`), so their contrast is **NOT_TESTED**.
- rule: WCAG 1.4.3 Contrast (Minimum).
- fix: if the card stays, darker text and at least 12–13 px. If the card goes (#1), the problem goes with it.
- sure: yes. This is S0 C1-3.

**#4 · Stress · 200% text breaks the hero**
- sees: at 390 with 200% text, the «↗» arrow of «Переглянути курси» sits outside the button. In the cards, words break mid-word («ПОПУЛЯР/НИЙ», «boundarie/s»), and «68%» covers «Ваш прогрес».
- costs: the main button looks broken, and the cards turn into a pile of letters.
- proof: `.primary-link` `scrollWidth > clientWidth` (390/200%). The capture shows the arrow to the right of the button edge and the broken words in the cards.
- rule: WCAG 1.4.4 Resize Text / 1.4.10 Reflow.
- fix: the arrow wraps with the text, or is hidden at large text. The cards either disappear at large text or keep whole words.
- sure: yes.

**#5 · Words · the illustration does not switch language**
- sees: with English selected, the page is in English, but the cards still say «Популярний курс», «Ваш прогрес», «Модуль 7 з 10».
- costs: a mixed-language page looks unfinished.
- proof: `artTexts` of guest-en at 390 and 1440. The strings in `HomeView.tsx:38-49` do not go through `t()`.
- rule: [copy] Consistent language.
- fix: through i18n, or no text in the illustration (see #1).
- sure: yes.

**#6 · Glance/Task · not one real course on the home page**
- sees: the home page has a slogan, two buttons, two facts and three features, but no course. The only way to the content is «Переглянути курси». The decorative block takes 352 px on a phone (y 720–1072) and about 430 px on a 768 tablet.
- costs: a guest does not see what the site actually teaches until the second page. The largest visual block is spent on a made-up course instead of a real one.
- proof: `.hero-art` `24 720 342×352` (390), `24 599 720×352` (768). `HomeView` has no data query.
- rule: [cog] Recognition over recall / [nav] Information scent.
- fix: replace the illustration with 2–3 real course cards (the same card as in the catalog), with «Переглянути всі курси» below them.
- sure: suspect. This changes the content of the home page, so it is the owner's call.

**#7 · Words · «▶ Як влаштоване навчання» promises a video**
- sees: the second button has a ▶ "play" icon, but it only jumps down the page to the three features. On a desktop those features are already visible under the hero, so the page moves 482 px.
- costs: the user waits for a video and gets a jump. After the jump, focus stays on the page body rather than on the section.
- proof: `href="#values"`. Enter at 1440: `scrollY` 0→482, `activeElement = BODY`. At 390: 0→1002.
- rule: [copy] Labels describe the result / Nielsen #2 Match with the real world.
- fix: drop the ▶ and make the button a link that names the destination, or remove it and leave «Переглянути курси» as the only action.
- sure: yes about the ▶. Removing the button is the owner's call.

**#8 · Squint/Twins · feature titles look like their descriptions**
- sees: in «Практика / Транскрипти / Обговорення» the title and the description have the same size and weight; only the colour differs.
- costs: the three features read as plain text rather than as headed items.
- proof: `.value-item h3` 16px / 400; `.value-item p` 16px / 400 (style rows from `cap.json`, light and dark).
- rule: [visual] Hierarchy by size/weight; [law] Similarity.
- fix: titles at a heavier weight (600–700), as on the other screens.
- sure: yes.

**#9 · Squint/Groups · the email notice runs into the hero**
- sees: after registration (`/?notice=verify-email`), «Ми надіслали лист. Підтвердіть пошту…» is plain white text right above the green pill. It has no frame and no icon, and there is no gap before the pill.
- costs: the important message (confirm your email) reads as part of the advertising text and is easy to skip.
- proof: `.hero-copy output` `24 115 282×46`, 16 px, `color #fff`, no class or role (390, EN).
- rule: [state] System status visible / [law] Common Region.
- fix: show it as a notice (the project's info status style, with an icon), above the hero or as its first block, with a gap.
- sure: yes.

**#10 · Words · one idea three times**
- sees: «Практичне навчання для розробників» (pill), «Підхід: Навчання через практику» (fact), «Практика» (first feature).
- costs: the screen repeats itself instead of telling more. The facts take 124 px on a phone.
- proof: `.hero-facts` `24 556 342×124` (390).
- rule: [cog] Reduce noise.
- fix: remove the «Підхід» fact, or replace it with a fact the page does not already state, such as the number of published courses.
- sure: suspect.

## Taste

- In light theme the hero stays dark (a dark band between a white header and a white feature strip). It looks deliberate, as a storefront accent. Not a defect.
- At 1440 there is about 95 px of empty space between the hero and the feature strip. Not a defect.

## Action inventory (guest, 390 and 1440)

```
skip «До основного вмісту» · go to content · own label · focus 3px · Tab #1 PASS
wordmark «CourseForge — головна» · home · aria-label · 152×44 (390) · Tab PASS
«Каталог» (header, 1440) · catalog · nav · 88×44 · Tab PASS
«Увійти» · sign in · own label · 78×44 · Tab PASS · NOT_TESTED activation (unsafe: sign-in flow)
«Меню» (390) · open menu · own label · 102×44 · Tab PASS (tested in G0)
UA / EN (1440) · language · group · 44×44 · Tab PASS (tested in S1/G0)
«Переглянути курси ↗» · catalog · own label · 342×51 / 233×51 · Tab PASS · Enter → /courses «Каталог курсів» PASS
«▶ Як влаштоване навчання» · features below · own label (misleading, #7) · 342×53 / 263×53 · Tab PASS · Enter → #values, focus lost to BODY
footer wordmark «CourseForge» · home · visible text · 186×44 · Tab PASS
footer «Каталог» · catalog · «Навчання» column · 57×44 · Tab PASS
hero cards · — (not an action, aria-hidden) · — · not focusable
```

The Tab order matches the visual order; after the footer, focus returns to the top of the page. The student version differs only in the header (G0); its metrics and axe results match the guest's.

## Not looked / NOT_TESTED

- **Contrast of the text on the gradient card** (`.hero-preview-label`, `.hero-preview-title`): NOT_TESTED, axe `incomplete` (`bgGradient`).
- **Real browser zoom:** NOT_TESTED, 200% is simulated.
- **The verify-email notice** was seen only in EN at 390.
- **Student, dark and EN at 768/1024/320/200%:** not looked. The page is static and the guest captures cover the layout.

## Owner answer

2026-11-23, owner: «Головна здається не дороблена так і залишилась макетом. І не співпадала з першочерговим макетом. Тому поки головну облиш».

The C1 fix is **deferred**. The home page is still a mock-up and does not match the original mock-up, so it is not reworked in the design pass. Finishing it is separate work, started when the owner decides. The ten problems above stay as candidates. #3 (contrast 2.53:1) is a failed WCAG 1.4.3 requirement, and #4 (200% text) fails 1.4.4/1.4.10. Both stay open until that work happens and are not accepted as limitations.
