# Admin restyle — Codex-тикети (зрізи)

Виконавець — Codex (Courses 1 драйвить), рев'ю зрізів — Courses 2. Контекст: `docs/admin-restyle/plan.md`
(розбивка сторінок + translation-дельти + прийняті рішення) і `docs/admin-restyle/README.md` (pain-список).
Референс вигляду: `docs/admin-restyle/proto/` (НЕ копіювати код/плумбінг — лише правила вигляду; див. `proto/_ORIGIN.md`
— §Contrast еталона РОЗХОДИТЬСЯ з рішенням 1: у прод primary = поглиблена заливка `#4763d0`, не large-text).

## Спільні обмеження (кожен тикет)

- React Aria + plain CSS на НАШИХ семантичних токенах; no Tailwind/Radix/CSS-in-JS. Класи — наші `.admin-*`/`.ui-*`.
- Токени: **без паралельного шару**; нові значення — названими токенами (див. рішення 2 плану). Бренд `#5273f2`
  + лайм-акцент; ваги 650/700/750; `--btn-primary-*` наш (primary-ГРАДІЄНТ поглиблено — обидві стопи ≥4.5:1 білому
  для normal-тексту, глянець D3 збережено; рішення 1 — НЕ плаский fill).
- A11y (гейт): ≥44px, видимий focus, 200% text-zoom + reflow без h-скролу, forced-colors, WCAG AA. **Ніколи**
  native `disabled`/`isDisabled` на нав/reorder — `aria-disabled` + no-op. Icon-дія = `aria-label` + tooltip.
- Логічні CSS-властивості. Плумбінг — реальний: TanStack Query / Router search-params / i18n / доменні типи
  (НЕ fixtures/hash/localStorage/CSV-blob/`setTimeout` з прототипу).
- Per-slice: `npm run format` (Codex пропускає oxfmt — format-check 1-й крок гейта). Гейт — ordered `make check`.
- **Browser-verify на РЕАЛЬНОМУ коді зрізу** (не на прототипі): desktop + 375px + 200% zoom + forced-colors.
- Великий зріз — ділити impl / тести окремими Codex-прогонами (resume за id).

---

## S0 — токени + shell (фундамент) · ПОВНИЙ e2e

**Файли:** `frontend/src/app/styles/tokens.css`, `tokens.admin.css`, `main.css` (shell `2042-2242`),
`frontend/src/presentation/admin/AdminShell.tsx`, `AdminNavIcon.tsx`, `frontend/src/routes/__root.tsx`.
**MUST:**
1. Нові контейнерні токени в `tokens.css`/`tokens.admin.css`: `--space-admin-sidebar:15.5rem`, `--space-admin-canvas:76rem`,
   `--space-dialog:36rem`, `--space-prose-measure:65ch`, `--space-prose-min:12rem`, `--space-tile-min:9rem`,
   `--space-icon:1.25rem`. Мисномери НЕ заводити (`--space-border`→`1px` інлайн; focus-ring→наш `3px`; callout→accent-rule).
2. Primary-ГРАДІЄНТ (рішення 1): поглибити `--btn-primary-bg` + `--btn-primary-bg-hover` так, щоб БІЛИЙ на КОЖНІЙ стопі
   (верх і низ, rest+hover) давав ≥4.5:1 для normal-тексту. Зараз верх `#6986ff`=3.26:1 (fail), низ `#4c6bf0`=4.51:1.
   Метод у системі: стопи з `--color-primary` через `color-mix(…+#000)` (напр. верх ~92%, низ = наявний
   `--color-primary-pressed` 86% `#4763d0`/5.30:1) — НЕ вводити паралельні захардкоджені hex. НЕ сплющувати в пласку
   заливку (глянець D3: highlight+glow зберегти). НЕ роздувати текст (лишаємо `.8125rem/650`). Той самий градієнт живить
   `.ui-segment[data-selected]` — поправиться разом. ⚠️ axe НЕ міряє градієнти → кожну стопу верифікувати вручну в браузер-кроці.
3. **Один нав:** прибрати декоративний `.admin-rail` (`aria-hidden` icon-дубль, main.css:2048) — лишити labeled `.admin-sidebar`.
4. **Narrow (≤52rem): mobile drawer**, не повноширинний блок над контентом (зараз main.css:2236 тисне контент). Toggle-кнопка в topbar.
5. Зберегти: eyebrow caps-лейбл + H1, правий хедер (Administration · user · UA/EN · Sign out).
**Acceptance:** shell рендериться desktop+375px(drawer)+200%zoom(no h-scroll)+forced-colors; один нав; primary AA;
ordered `make check` + **повний e2e** зелені локально (run-id у звіт; CI на білінгу).

## S1 — DataTable (спільний) · ПОВНИЙ e2e
**Файли:** `frontend/src/presentation/ui/data.tsx` (Table), `main.css` (.ui-table overrides, .admin-user-actions 845),
`tokens.css` (.ui-table 781-818). Споживачі: `CourseListView`(#1 courses), `UserListView`(#6 users).
**MUST:** elevation (таблиця в `.ui-card`); **sticky inline-end actions** (`position:sticky; inset-inline-end:0` + фон+hairline);
**narrow ≤52rem: stacked labeled cards** (або scroll зі sticky identifier+actions) — без mid-word wrap; compact icon-дії
(`aria-label`+tooltip) замість 3 стекованих кнопок; human-readable 1-ша колонка; right-align чисел; видимий стан фільтра/сорту.
**Acceptance:** courses+users на 375px не ламаються, дії завжди досяжні; `make check` + **повний e2e**.

## S2 — form/card (спільний) · ПОВНИЙ e2e (зачіпає спільні .ui-field/.ui-card)
**Файли:** `tokens.css` (.ui-card, .ui-field, .ui-pill), `main.css` (admin .admin-*-status 659-773 → на .ui-pill),
`frontend/src/presentation/ui/controls.tsx`. Споживач: `CourseEditorView`(#3).
**MUST:** `.ui-card` elevation скрізь на admin; статуси → `.ui-pill` badge (прибрати саморобні .admin-*-status);
форм-патерни: label-above, blur-валідація, error-summary з лінками, preserve-input, destructive-confirm. Interaction-states
(rest/hover/active/focus/disabled/loading) на контролях.
**Acceptance:** course-editor форма + будь-який статус-pill; `make check` + **повний e2e**.

## S3 — lesson editor (#4) · e2e: ПОВНИЙ, якщо чіпає reader-shared
**Файли:** `LessonDocumentEditor.tsx`, `LessonDocumentRichText.tsx`, `main.css` (.lesson-editor-* 1455-1667).
За `docs/lesson-builder/decision-editor-redesign.md` (adopt-with-changes). ⚠️ **Reader-spільний селектор** `.lesson-prose` /
`:is(.lesson-document,.lesson-editor-input)` (main.css:1361+) — якщо S3 його чіпає (coherence SHOULD майже напевно) → **повний
e2e** (reader-регресія інакше прослизне); інакше довести в тикеті, що reader-shared не зачеплено → таргетний e2e.
**MUST:** візуальна мова прототипу на наших токенах; `aria-disabled`+no-op на move (НЕ isDisabled); icon+tooltip; future-handle
`aria-hidden` без натяку на drag (drag — окрема фаза).
**Acceptance:** editor + student-reader (playback.spec) зелені; відповідний e2e.

## S4 — analytics (#2) · таргетний e2e (admin.analytics не спільний)
**Файли:** `admin.analytics.tsx` · `AnalyticsView`, `tokens.css` (.ui-stat/.ui-chart 521-583, .admin-analytics 584-605).
**MUST:** stat-тайли `.ui-card`/`.ui-stat` elevation (не зливаються); group-heading між 3 групами; фікс letter-spacing грошей
(`.ui-stat-value` / money); `.ui-segmented` range; `.ui-chart` trend.
**Acceptance:** analytics desktop+375px+200%; `make check` + таргетний analytics e2e.

## S5 — learners (#5) · таргетний e2e
**Файли:** `admin.learners.tsx` · `LearnerOverview`, `main.css` (learners-list/learners-courses).
**MUST:** розділені зони enrollment/статус (фікс «…Active» впритул — додати gap); рядки-картки з дільниками (не злипаються);
`.ui-progress`/`.ui-meter` трек замість волосини.
**Acceptance:** learners desktop+375px; `make check` + таргетний learners e2e.

## S6 — reports + refunds (#7,#8) · таргетний e2e
**Файли:** `admin.participant-reports.tsx` · `ParticipantReportQueue`, `admin.refunds.tsx` · `RefundQueue`,
`main.css` (.admin-user-actions 845 fixed-width — прибрати на користь компактних).
**MUST:** кожен репорт/рефанд = `.ui-card` панель зі структурою (не «Label: value» дамп); дії прив'язані (не відкинуті вправо);
статус/«Open» = `.ui-pill` badge; структурований empty-state (icon+рамка) + populated-черга.
**Acceptance:** reports+refunds (populated+empty) desktop+375px; `make check` + таргетний e2e обох черг.

---

**Порядок:** S0 → S1 → S2 → S3 → S4 → S5 → S6. Кожен: Codex impl (+тести окремо за потреби) → `npm run format` →
ordered `make check` (+ повний e2e де вказано) локально → browser-verify на реальному коді → рев'ю Courses 2 → коміт/PR.
