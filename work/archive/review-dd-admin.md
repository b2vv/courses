# Review — D-d: admin on NorthWindLab stylistics

Reviewer: **Courses 2**. Author: Courses 1 (design lane, direct). Uncommitted —
review the working tree diff on `feat/payment` (base = 1fe6cef, the visual-compare
commit). This builds on the two-density split (311ceb5) already reviewed by you.

РОБОЧИЙ ФАЙЛ: `work/review-dd-admin.md`
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append). Перший рядок — `ДАЛІ:`. Далі ФАКТ/РІШЕННЯ/
ПИТАННЯ. Фінал = стислий переказ.

## What & why
Owner: "D-d адмінка має бути на стилістиці https://northwindlab.com/". I opened
northwind, extracted its real tokens, and re-skinned the admin surface: near-black
editorial dark (#070a13 / card #0d121c / hairline #242932), teal accent, Space
Grotesk display, thin borders, tinted status pills. Owner decisions: admin follows
**light/dark** (not dark-only), start with the **dashboard anchor** then roll out.

## Scope (files, all frontend)
- **package.json / package-lock.json** — added `@fontsource-variable/space-grotesk`.
  Same lock hazard as Inter: npm 11.6.2 prunes `@emnapi/core|runtime` (foreign-platform
  wasm optionals) → container `npm ci` fails. Restored HEAD lock and spliced ONLY the
  space-grotesk entry. `npm ci --dry-run` exit 0; @emnapi kept (2 node entries).
- **tokens.admin.css** — the admin scope now re-binds palette + accent + display font
  in BOTH themes, on top of the density knobs:
  - display: Space Grotesk; density unchanged (dense).
  - light (derived editorial-light): bg #f7f9fa, surface #eef2f4, text #0f1720,
    control border #7d828c, hairline #dfe4e8, teal primary **#0b6b78** (deep, so teal
    TEXT clears 4.5:1 on both bg and card surface), on-primary #fff.
  - dark (faithful northwind): bg #070a13, surface #0d121c, text #dae0e7, muted
    #8b95a3, control border **#656e7b** (>=3:1 on both bg and surface), hairline
    #242932, teal primary #20b2c5, on-primary #071018.
- **main.css** — space-grotesk @import; `.app-shell` (wraps header+main+footer, carries
  data-surface, paints --color-background across the viewport so body behind never shows
  the other surface's colour); admin editorial block: teal `.eyebrow`, tighter Space
  Grotesk heading, `.admin-list-results` as an editorial card, `.ui-table` hairline rows +
  muted uppercase header, tinted status pills, tinted (not solid) `.ui-button--danger`.
- **__root.tsx** — moved `data-surface` from `<main>` onto the new `.app-shell` wrapper.

## Verified (Courses 1)
- format ✓ · lint ✓ (+architecture 126) · typecheck ✓ · rstest **508/508** ✓.
- a11y (computed + calc): LIGHT teal-text 5.5–5.9:1, control border 3.4–3.7:1, btn text
  4.7:1; DARK teal-text 7.8:1, control border 3.6–3.8:1, text 14.9:1, muted 6.5:1,
  danger-tint ink >=4.8:1 both themes.
- Screens (admin login + instructor login, Playwright, light+dark): admin course-list ✓,
  admin users ✓ (danger buttons toned). Course-detail rendered its ERROR state (the
  course id from the list link failed to load — a data/loading thing, NOT this diff;
  theme inherited correctly on the error state). /teach for instructor@ is EMPTY (that
  account owns no courses/comments) — chrome inherited; content not verifiable on this data.

## Please scrutinise
1. a11y in BOTH admin themes — did any control/link fall below threshold? Especially the
   deep-teal-on-light and the dark control border #656e7b on surface #0d121c (I got 3.08
   at #5a636f, bumped to #656e7b for margin).
2. Theme cascade specificity: `[data-surface='admin']` (light) vs `:root[data-theme='dark']
   [data-surface='admin']` and the `@media dark :root:not([data-theme='light']) …` — does
   admin re-theme correctly under system-dark, forced-light, forced-dark? Client unaffected?
3. `.app-shell` refactor — header/footer now inside it; skip-link still first focus; no
   layout/scroll regression; footer still bottom (flex column, main flex:1).
4. Lockfile — space-grotesk added, @emnapi intact, nothing else drifted.
5. The course-detail error state — is it a pre-existing/data issue or did my diff touch a
   loader? (I did not, but please confirm from the diff.)

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-09-30)

ДАЛІ: перевіряю 5 підозр. Пріоритет — (1) a11y обох admin-тем (перерахунок контрастів:
deep-teal #0b6b78 на light bg/surface як ТЕКСТ 4.5:1; dark control-border #656e7b на #070a13/#0d121c
3:1), (2) каскад теми під system-dark/forced-light/forced-dark + клієнт не зачеплено, (3) .app-shell
(skip-link перший фокус, футер знизу), (4) lockfile space-grotesk+@emnapi, (5) course-detail error —
чи диф чіпав лоадер. Спершу дифи tokens.admin.css / main.css / __root.tsx, тоді рахую.

### F1 [a11y обидві теми] — ПРОХОДЯТЬ (перерахунок), 1 nb на danger-бордер
LIGHT: teal-текст #0b6b78 на bg/surface 5.87/5.5 (≥4.5✓); control #7d828c 3.65/3.42 (≥3✓);
btn #fff/#0b6b78 6.2; body 17.1; muted #5b6675 5.52 (≥4.5✓); hairline 1.21 (декор).
DARK: control #656e7b 3.83/3.63 (≥3✓); teal #20b2c5 7.76/7.35; btn #071018/#20b2c5 7.51; text 14.9;
muted 6.52; hairline 1.35 (декор).
Тінтовані піли (14% currentColor wash): LIGHT ink 4.74/4.81/4.92 (тонко, але ≥4.5✓); DARK 8.05/9.01/8.62.
[nb] danger-БУТТОН бордер = color-mix(danger 40%) = 1.98:1 (light) / 2.64:1 (dark) на surface — <3:1.
Фон-tint (14%) теж субтильний. Тобто МЕЖА danger-кнопки не дає 1.4.11 3:1; ідентифікація тримається
на danger-кольоровому ТЕКСТІ (4.74/8.05 — проходить). Захищувано (текст-ідентифікований контрол), але
danger-кнопка так виглядає майже як status-піл (той самий 14% wash) — не-blocking, але якщо хочеш
запас/чіткість «це клікабельно» — підняти бордер до ≥3:1 (≈60-70% wash або суцільний). Твій вибір.
ДАЛІ: #2 каскад теми (звірка з base-патерном), #3 app-shell/skip-link, #4 lock, #5 course loader.

### F2 [#2 каскад теми] — ЧИСТО (всі 4 режими)
Admin дзеркалить base-патерн, scoped під `[data-surface='admin']`:
- base light `[data-surface='admin']` палітра = (0,1,0);
- `@media dark {:root:not([data-theme='light']) [data-surface='admin']}` = (0,3,0);
- `:root[data-theme='dark'] [data-surface='admin']` = (0,3,0).
Перевірка режимів: system-dark(no force)→dark(0,3,0)>light(0,1,0)✓; forced-light→:not виключає dark-media,
:root[dark] не матчить→light✓; forced-dark(sys-light)→media off, :root[dark] матчить→dark✓; system-light→light✓.
Вища специфічність dark-блоків б'є light незалежно від порядку джерела. Base :root dark-палітра
(client) на <html>, admin — на глибшому .app-shell → для admin-піддерева виграє ближчий елемент.
КЛІЄНТ не зачеплено: усі admin-правила під [data-surface='admin']; client має data-surface='client'.

### F3 [#3 .app-shell] — ЧИСТО
__root: skip-link ПЕРШИЙ дочірній у `.app-shell`, `<main id="main">` ціль #main ціла, data-surface
перенесено з <main> на wrapper (main інеритить). `.app-shell` flex-column min-100dvh, main flex:1 0 auto
→ футер знизу (sticky-footer), без scroll-регресій; app-shell фарбує --color-background на весь viewport,
тіло іншою темою не просвічує (як і задумано). Header/footer тепер flex-items (natural height) — ok.

### F4 [#4 lockfile] — ЧИСТО
+12/−2: space-grotesk з version+resolved+integrity(sha512); @emnapi цілий (23 посилання); єдина
«видалена» — кома на inter (space-grotesk дописано після). npm ci верифікує. Дрейфу нема.

### F5 [#5 course-detail error] — ПІДТВЕРДЖЕНО пре-існуючим
Диф чіпає лише css + __root.tsx + package (жодного loader/query/data). Error-стан деталі курсу —
дані/loading, НЕ цей диф. Тему на error-стані інеритнуто коректно (як ти й спостерігав).

## ПІДСУМОК: 0 blocking, 1 non-blocking, commit-ready
nb: danger-button бордер <3:1 (1.98 light / 2.64 dark) — ідентифікація на danger-тексті (проходить),
захищувано; опційно підняти бордер до ≥3:1 для чіткості «клікабельно» vs status-піл. Твій вибір.
Усе інше (a11y обох тем, каскад 4 режими, app-shell, lockfile, course-error) — ЧИСТО.
Out-of-scope не чіпав: screenshots/*, work/review-design-tokens.md (мій попередній файл).

## Resolution — Courses 1 (after C2 verdict: 0 blocking, 1 non-blocking)
ДАЛІ: закрити non-blocking (danger button reads like a pill / border <3:1).
- Applied: `[data-surface='admin'] .ui-button--danger` border → SOLID `var(--color-danger)`
  (was color-mix 40%). Border vs surface now 5.95:1 light / 10.55:1 dark (>=3:1) → the
  boundary itself satisfies 1.4.11, and the outlined look is clearly a button, distinct
  from the borderless status pills that share the 14% wash. format:check ✓, re-shot users
  (dark) confirms.
- F1–F5 accepted clean. Course-detail error state noted as pre-existing/data (out of D-d).
Committing two slices: (1) font, (2) D-d admin.
