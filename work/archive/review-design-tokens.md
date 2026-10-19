# Review — storefront softening + two-density token split + visual-compare tool

Reviewer: **Courses 2**. Author/coder: Courses 1. Everything below is **uncommitted**
(review the working tree, `git diff` + untracked). Old scheme: C1 codes, C2 reviews.

РОБОЧИЙ ФАЙЛ: `work/review-design-tokens.md`
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — рядок `ДАЛІ:` з тим,
що збираєшся перевіряти. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінал = стислий переказ.

## What changed & why

Storefront looked "кострубато" / not soft vs the client's mockup
(https://courseforge-academy.strelchenko-dev.chatgpt.site/). I opened the mockup and
extracted its real tokens; the gap was **not colours** but: (1) system font vs **Inter**,
(2) heavy floating cards vs **flat** ones, (3) blue-tinted greys vs **neutral**. Fixed
that, then split the token layer into **shared / client / admin** ("one system, two
densities") on the owner's instruction.

### Scope (files)
- **Design softening**
  - `frontend/package.json` + `package-lock.json` — added `@fontsource-variable/inter`.
    NOTE: my local `npm install` (npm 11.6.2) PRUNED `@emnapi/core|runtime` (foreign-platform
    wasm optionals) from the lock → container `npm ci` failed. I restored HEAD's lock and
    spliced ONLY the fontsource entries back via a node script. `npm ci` now green in Docker.
  - `frontend/src/app/styles/tokens.css` — Inter font vars; neutral palette
    (`--color-text #111827`, `--color-muted #6b7280`, `--color-border #aeb4be`, new
    `--color-hairline`); primary `#5d4df4`; hero `#101827`; cover `#6d5dfc→#4029aa`;
    softer shadow scale.
  - `frontend/src/app/styles/main.css` — Inter `@import`; tighter display type
    (h1 −0.045em/lh1.03/700, h2 −0.025em/700); decorative dividers → hairline.
- **Two-density token split**
  - `tokens.css` `:root` — density knobs (`--card-padding/-radius/-border-color/-shadow/
    -shadow-hover`, `--surface-gap`); `.ui-card` now consumes them.
  - `tokens.client.css` NEW — `[data-surface='client']` airy.
  - `tokens.admin.css` NEW — `[data-surface='admin']` dense.
  - `main.css` — imports both; `.course-card` base border/shadow now from `.ui-card`
    density vars, hover uses `--card-shadow-hover`.
  - `src/routes/__root.tsx` — `data-surface` on `<main>` by pathname
    (`/admin`,`/teach`→admin, else client).
- **Tool**
  - `frontend/tools/visual-compare/` (compare.mjs + README + .gitignore) — Playwright
    side-by-side sheets (reference | ours), agent-judged, no pixel diff.

### Not my design scope (flag, don't fold in)
- `docs/foundation/screenshots/catalog-*.png` — modified BEFORE this work (pre-existing).
- `docs/foundation/screenshots/.catalog-desktop.png.swp` — vim swap junk, must NOT be committed.
- `src/app/providers/i18n.test.tsx` — earlier heading-query fix + prettier reformat.

## Gates already run (Courses 1)
format:check ✓ · lint ✓ (+architecture boundaries 126 files) · typecheck ✓ · rstest **508/508** ✓ ·
Docker frontend build ✓ · stack healthy, home/catalog HTTP 200.
Density mechanism verified by computed styles: same `.ui-card` → client 24px/#e6e8ec/no-shadow/r16,
admin 16px/#aeb4be/subtle-shadow/r12.

## Please scrutinise (my own suspicions)
1. **a11y contrast (WCAG 1.4.11, 3:1 for control boundaries).** `--color-border` went
   `#9ba9ba`→`#aeb4be` (lighter). Inputs/buttons/selects use it. Does the control border
   still pass 3:1 on white AND on dark? Hairline `#e6e8ec` is decorative (cards/dividers),
   not a control boundary — but confirm no CONTROL fell back to hairline.
2. **Dark mode.** Light palette changed; I added `--color-hairline` to both dark blocks but
   only screenshotted LIGHT. Storefront + a `.ui-card` in dark — nothing regressed?
3. **`.course-card` specificity.** Base border/shadow now come from `.ui-card` density vars.
   No double border, no lost hover, no leftover `--shadow-floating` reference?
4. **Lockfile integrity.** The hand-splice — is `package-lock.json` otherwise identical to
   HEAD except the fontsource add + kept @emnapi? Any other drift? (`git diff` the lock.)
5. **data-surface wiring.** header/footer sit OUTSIDE `<main>` → shared defaults. Intended?
   `/learn`→client, `/teach`→admin, `/login`→client — correct mapping? Any route missed?
6. **Bundle.** Inter Variable woff2 weight — acceptable? Any duplicate font loading?
7. **Tool.** compare.mjs bare-imports `playwright`; `out/` gitignored. Anything that would
   break in CI or leak large files into git?

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-09-30)

ДАЛІ: перевіряю по черзі 7 підозр, пріоритет — (1) a11y-контраст control-border 3:1 після
#9ba9ba→#aeb4be на white і dark (WCAG 1.4.11), (2) dark mode, (4) цілісність package-lock.
Спершу — дифи tokens.css / main.css / нові token-файли / __root.tsx, тоді рахую контрасти.

### F1 [BLOCKING a11y — WCAG 1.4.11] Light-mode control border 2.09:1 (<3:1)
ФАКТ: `.ui-input`/`.ui-button` база (tokens.css:97-98) = `border:1px solid var(--color-border)`;
`.ui-input` фон = `--color-background` #fff (tokens.css:182). `.ui-button--secondary` фон
`--color-surface` #f6f7f9 (майже білий) + border `--color-border`. Отже текст-інпути, secondary-
кнопки й `.ui-select-button` покладаються ЛИШЕ на бордер як межу контрола.
Контраст `--color-border #aeb4be` на #fff = **2.09:1**. WCAG 1.4.11 вимагає **3:1** для меж
контролів. Старий `#9ba9ba` = **2.39:1** — теж <3:1. Тобто дефект ПРЕ-ІСНУЮЧИЙ, але зміна
робить світлішим (2.39→2.09), тобто гірше саме там, де ти й підозрював.
DARK ok: `--color-border #6f839e` на dark-bg #111c2c = 4.42:1, на surface #1c2b3f = 3.69:1 — проходить,
і його не міняли. `.ui-button--primary` фон #5d4df4 — межа з фону, не з бордера, проходить.
РІШЕННЯ (фікс дешевий завдяки твоєму спліту decor/control): підняти САМЕ `--color-border` (контроли)
до L≤0.30, тобто ≈#767e8a або темніше (тоді ≥3:1 на білому); `--color-hairline #e6e8ec` лишити
світлим для карток/роздільників (декор, 1.4.11 на них не поширюється). Ідеально — окремий
`--color-control-border`. Перевір axe-правило non-text-contrast (часто не в дефолтному наборі → міг проскочити гейт).
ДАЛІ: #3 .course-card (подвійний бордер/hover/залишок --shadow-floating), #4 lockfile, #5 data-surface, нові token-файли, #7 тул.

### F2 [#3 course-card] — ЧИСТО
`.course-card ui-card` (CatalogView.tsx:83) — композиція. Border-width 1px від `.ui-card`
(--card-border-color=hairline); hover міняє border-color→--color-border + box-shadow→--card-shadow-hover;
додано border-color у transition. Подвійного бордера нема (border лише в .ui-card), hover не втрачено.
`--shadow-floating` не осиротів — легітимно вживає `.ui-card--floating` (tokens.css:455/458).

### F3 [#4 lockfile] — ЧИСТО
Diff lock: +12/−2. Єдина «видалена» стрічка — `"zod":"4.5.4"` без коми → з комою (fontsource
дописано в root deps після zod). fontsource-запис має version+resolved+integrity(sha512) → npm ci
верифікує. @emnapi присутній (23 посилання) — не спрунено, як і планувалось. npm ci зелений у Docker
(гейт) підтверджує sync package.json↔lock. Дрейфу нема.
ДАЛІ: #5 data-surface (__root.tsx), нові tokens.client/admin, #7 тул, dark mode візуально.

### F4 [staging risk] tokens.client.css / tokens.admin.css — untracked
ФАКТ: main.css вже імпортує `./tokens.client.css` і `./tokens.admin.css`; обидва файли `??` в git status.
Білд зелений (файли є на диску), але вони не потраплять у коміт, якщо не застейджити явно. Без них checkout гілки на іншій машині → @import resolve fail → build broken.
РІШЕННЯ: `git add frontend/src/app/styles/tokens.client.css frontend/src/app/styles/tokens.admin.css` разом з main.css.

### F5 [#5 data-surface] — ЧИСТО
ФАКТ: `data-surface` ставиться на `<main>` (__root.tsx:161). Усі `/admin/*` і `/teach/*` → 'admin'; усе інше → 'client'. Routes: admin.($.tsx|courses|index|users), teach.(courses|tsx). `.ui-card` вживається лише в CatalogView.tsx і HomeView.tsx — обидва всередині <main> → успадковують [data-surface='client']. header/footer поза <main> — не вживають --card-* токенів, shared defaults є коректними.

### F6 [#7 tool] — ЧИСТО (untracked, CI-safe)
ФАКТ: `frontend/tools/visual-compare/` — untracked (`??`); не потрапить в git без явного `git add`. Нема ризику для CI або lock-артефактів.

### F7 [dark mode] — ЧИСТО
ФАКТ: `--color-hairline` додано до обох dark-блоків (`@media prefers-color-scheme:dark :root:not([data-theme=light])` і `:root[data-theme=dark]`). `--card-shadow` в dark-блоках не потребує перевизначення — --card-shadow посилається на --shadow-card, той на --color-media-frame, що незмінний між темами.

### Summary
Blocker: **F1** (control border contrast 2.09:1 < 3:1 WCAG 1.4.11 — worsened from 2.39:1).
Pre-commit: **F4** (stage tokens.client/admin.css).
Clean: F2 (course-card), F3 (lockfile), F5 (data-surface), F6 (tool), F7 (dark mode).

### F4 [#5 data-surface] — ЧИСТО (1 дрібне спостереження)
surface='admin' для /admin,/teach, інакше client; на `<main>`. Усі admin/teach-роути
(admin.*, teach.*) під префіксами — жодного admin-surface поза ними (нема /instructor,/studio).
Knobs ре-байндяться на `[data-surface]`, каскадять у нащадків `<main>`; header/footer поза `<main>`
→ :root(airy) — навмисно. СПОСТЕРЕЖЕННЯ (nb): портальні поверхні (модалки/тости/діалоги) рендеряться
поза `<main>` → завжди client-щільність навіть в адмінці; якщо адмінський діалог має бути щільним —
не буде. Дешево лишити як є або підняти data-surface на кореневий контейнер layout.

### F5 [tokens.client/admin] — ЧИСТО
Обидва лише ре-байндять knobs, без нових кольорів/шрифтів. Усі вжиті vars існують: --radius-panel
(0.75rem), --radius-card(1rem), --space-4(1rem). client = :root-дефолти (airy, трохи надлишково,
але явно й нешкідливо); admin = щільно (space-4, radius-panel, --color-border, shadow-card/floating).

### F6 [#2 dark, #6 bundle] — ЧИСТО
Dark: --color-hairline #26364c додано в ОБИДВА темні блоки (@media prefers + [data-theme]).
Control-border dark #6f839e не мінявся, 3.69–4.42:1 — проходить. Токенно регресій нема (візуальний
smoke у dark усе ж раджу, бо ти знімав лише light).
Bundle: єдиний `@import '@fontsource-variable/inter'` (main.css), жодного Google Fonts / другого
Inter-імпорту. Один variable woff2 — прийнятно, подвійного завантаження нема.

### F7 [#7 тул] — ЧИСТО
compare.mjs bare-import playwright (dev-утиліта, не CI); `out/` + `out/raw/` під .gitignore тула →
великі PNG не потечуть. Гейти (lint/typecheck) зелені з tools/ у дереві → tsconfig/lint не спотикаються.
REF_URL — публічний мокап, не секрет. При `git add frontend/tools/` вкладений .gitignore виключить out/.

## ПІДСУМОК рев'ю
1 BLOCKING: **F1** — light-mode control-border 2.09:1 (<3:1, WCAG 1.4.11); пре-існуюче, але зміна
#9ba9ba→#aeb4be погіршує. Фікс: підняти лише --color-border (контроли) до ≥3:1 (≈#767e8a або темніше),
hairline лишити. Перевір axe non-text-contrast (міг проскочити гейт).
Решта (F2–F7): dark, course-card, lockfile, data-surface, token-файли, bundle, тул — ЧИСТО.
Out-of-scope не чіпав: docs/foundation/screenshots/*, .swp, i18n.test.tsx.

## Resolution — Courses 1 (after C2 verdict)

ДАЛІ: закрити BLOCKING F1, решту (F2–F7 clean) лишити як є.

- **F1 FIXED.** Raised control-grade `--color-border` `#aeb4be`→**`#7d828c`** in light `:root`
  (kept `--color-hairline #e6e8ec` light for decorative card/divider outlines — 1.4.11 does
  not apply there). No separate `--color-control-border` needed: the hairline split already
  gives a two-tier system (hairline = decorative, border = control-grade).
  - Verified computed on `/courses`, `.ui-input` border vs its bg:
    **light 3.86:1**, **dark 4.42:1** (both ≥3:1). Dark control border `#6f839e` unchanged.
  - Dark-mode visual smoke captured (C2 rec #2): storefront + cards + inputs clean, no
    regression; hairline dividers `#26364c` correct.
  - format:check ✓ after the edit.
- F2–F7: accepted as clean, no change. NB (C2 #5): portal modals/toasts render outside
  `<main>` → always client density; acceptable for now (no dense admin dialog exists yet;
  revisit at D-d if an admin dialog must be dense).

Ready to commit as three slices (design / two-density split / tool) pending owner go-ahead.
