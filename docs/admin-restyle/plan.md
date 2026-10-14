# Admin restyle — план перекладу прототипу (зріз 1 → код)

Прототип від плагіна: `admin-restyle` (React + React Aria + plain CSS). Локально:
`scratchpad/admin-proto/admin-restyle/` (src.jsx, dist/styles.css, fixtures.js, README).
Це **референс цільового вигляду**; у прод іде перекладеним на наш код і токени.

## Вердикт придатності (статичний аналіз)

Прототип **чистий** — плагін ретельно виконав brief:
- **0 паралельних токенів** (нема `--type-*/--size-*/--weight-*/--leading-*/--tracking-*/--duration-*/--layer-*`),
  нема `--size-zero`/`--color-transparent`. Токени — точно наші (`#0c0d12`, бренд `#5273f2`, лайм-акцент `#b9ff4a`).
- Ваги **650/700/750** ✓. `--btn-primary-bg` — наш градієнт, не `color-mix`-ремікс ✓.
- **`aria-disabled` + no-op** замість native `disabled`/`isDisabled` (9 використань, native нема) — уникає нашого
  focus-drop gotcha ✓.
- Icon-дії через `IconButton name+label` + `.ui-tooltip` (aria-label 12×) — **дельта A (доступні імена) виконана** ✓.
- Контраст-landmine (дельта B) адресовано: primary-текст великий/напівжирний (1.1875rem/750 → шлях 3:1), solid `#5273f2`
  не під normal-текст; README сам фіксує 3.25:1/4.50:1 по стопах градієнта ✓.
- Один labeled nav + mobile drawer (rail прибрано) ✓; `.ui-card` elevation ✓; статуси через `.ui-pill` ✓;
  форми label-above + blur-валідація + error-summary + preserve-input ✓.
- README чесний: демо-дані, без прод-shell/API, «browser checks for 375px / 200% zoom / forced colors still required».

**Браузерна перевірка прототипу — НЕ виконана** (renderer завис при рендері на 8098). Лишається зробити: 375px,
200% text-zoom, forced-colors — перед тим як покладатися на вигляд.

## Translation-дельти (рішення перед кодом)

1. **Нові `--space-*` токени.** Прототип ввів і ЗАДОКУМЕНТУВАВ: `--space-admin-sidebar:15.5rem`,
   `--space-admin-canvas:76rem`, `--space-dialog:36rem`, `--space-prose-measure:65ch`, `--space-prose-min:12rem`,
   `--space-tile-min:9rem`, `--space-icon:1.25rem`, `--space-border:1px`, `--space-focus-ring/offset:3px`,
   `--space-callout-rule:.25rem` + `--color-modal-backdrop`, `--color-forced-*`.
   → Корисні контейнерні (sidebar/canvas/dialog/prose-measure/prose-min/tile-min) — завести в `tokens.css`/`tokens.admin.css`
   як названі токени. **Перейменувати мисномери:** `--space-border` (ширина межі ≠ spacing) → інлайн `1px` або `--border-width`;
   `--space-focus-ring/offset` **дублюють** наш наявний focus-outline (`3px`, tokens.css) — не заводити, вживати наш;
   `--space-callout-rule` → лишити як є або `--border-accent`.
2. **Primary-кнопка: контраст vs наша шкала.** Прототип збільшив primary-ТЕКСТ до 1.1875rem/750, щоб пройти
   large-text 3:1. Наші реальні кнопки — `.8125rem/650`. **Рішення (рекомендую):** НЕ роздувати текст, а **поглибити
   заливку** — наш уже вирішений gotcha: `color-mix(in srgb, var(--color-primary) 86%, #000)` = `#4763d0` = **5.30:1**
   для normal-тексту (axe-перевірено раніше). Так тримаємо нашу button-шкалу І проходимо AA. ⚠️ Градієнт axe не
   міряє — білий на `--btn-primary-bg` верифікувати вручну. *(Відкрите рішення власника/Courses 2.)*
3. **Нові helper-класи** прототипу (`.admin-brand`, `.admin-nav-item`, `.admin-nav-label`, `.sidebar-reference`,
   `.sidebar-footer`, `.admin-icon-button`, `.card-heading`, `.pill-dot`, `.ui-error-summary`, `.ui-tooltip`, `.ui-state`,
   `.lesson-editor-*`) → додати в `main.css` під `[data-surface='admin']`, змапивши на наші наявні де є.
4. **Дані/плумбінг:** fixtures/hash-routing/`localStorage`(lang)/CSV-blob/`setTimeout(450)` — ПРОТОТИПНІ. У прод:
   TanStack Query (save/publish/upload, конфлікт-token), Router search-params, i18n, доменні типи. Не переносити плумбінг.
5. **Не ставити власний shell у прод** — перекласти правила на наш `AdminShell.tsx`, не замінювати DOM.

## Розбивка по сторінках (8) → наші цілі

| # | Прототип (page) | Eyebrow / що показує | Наш таргет (route · view) | Що перекласти (pain-фікс) |
|---|---|---|---|---|
| 1 | `courses` | CONTENT MANAGEMENT · таблиця курсів + create-dialog | `admin.index.tsx` · `CourseListView` | DataTable: elevation, sticky inline-end actions, stacked-cards <52rem, compact icon-дії, human-readable 1-ша кол., статус-pill |
| 2 | `analytics` | PLATFORM STATISTICS · 3 групи stat-тайлів + chart + range | `admin.analytics.tsx` · `AnalyticsView` | `.ui-stat`/`.ui-card` elevation (тайли не зливаються), group-heading між групами, фікс letter-spacing грошей, `.ui-segmented` range |
| 3 | `course` | COURSE WORKSPACE · форма-редактор (CourseForm/VideoForm) | `admin.courses.$courseId.tsx` · `CourseEditorView` | `.ui-card` групування, label-above + blur-валідація + error-summary, destructive-confirm (delete/revoke), grants-панель |
| 4 | `lesson` | LESSON DOCUMENT · секції-картки, toolbar, preview | `admin.courses.$courseId_.lessons.$lessonId.tsx` · `LessonDocumentEditor` | **За editor-рішенням** (`decision-editor-redesign.md`): візуальна мова прототипу на наших токенах; `aria-disabled` no-op, icon+tooltip, future-handle `aria-hidden` |
| 5 | `learners` | LEARNING EXPERIENCE · список-картки учнів | `admin.learners.tsx` · `LearnerOverview` | розділені зони enrollment/статус (фікс «…Active» впритул), рядки-картки з дільниками, `.ui-progress`/`.ui-meter` замість волосини |
| 6 | `users` | ACCOUNT ADMINISTRATION · таблиця (еталон) | `admin.users.index.tsx` · `UserListView` | compact + sticky row-actions; решта вже охайна — тримати як референс DataTable |
| 7 | `reports` | SAFETY & TRUST · структурована черга | `admin.participant-reports.tsx` · `ParticipantReportQueue` | кожен репорт = `.ui-card` панель (не «Label: value» дамп), дії прив'язані, «Open»=`.ui-pill` badge, QueueFilter |
| 8 | `refunds` | PAYMENTS & REFUNDS · структурована черга | `admin.refunds.tsx` · `RefundQueue` | структурований empty-state (icon+рамка) + populated-черга (рядки + resolve/revoke + статус) |
| — | shell | один labeled nav + mobile drawer + topbar | `AdminShell.tsx` · `main.css:2042-2242` · `tokens.admin.css` | ОДИН нав (прибрати декор `.admin-rail`), narrow = drawer (не їсть контент), зберегти eyebrow+H1 і правий хедер |

## Зрізи (порядок) — спільні примітиви першими, екрани успадковують

- **S0 — токени + shell.** Звести нові `--space-*` (дельта 1), поглибити primary-fill (дельта 2), `AdminShell`: один нав +
  mobile drawer. Гейт: shell на desktop+375px+200%zoom, forced-colors.
- **S1 — DataTable (спільний).** elevation + sticky inline-end actions + stacked-cards <52rem + compact icon-дії
  (aria-label+tooltip). Споживачі одразу: `courses`(#1), `users`(#6). ⚠️ крос-зрізна зміна спільної таблиці → **повний e2e**.
- **S2 — form/card (спільний).** `.ui-card` elevation скрізь, статус-pill, label-above+blur-валідація+error-summary.
  Споживач: `course`(#3).
- **S3 — lesson editor** (#4) — за editor-рішенням (перший «видимий» зріз, як вирішено).
- **S4 — analytics** (#2): тайли-elevation + групи + money letter-spacing.
- **S5 — learners** (#5): картки-рядки + spacing + progress.
- **S6 — reports+refunds** (#7,#8): панель-черги + empty/populated.

## Гейт і процес (кожен зріз)

- Рев'ю: `interface-review` (intake-чекліст Courses 2 + ux-review) на скрині зрізу → далі код.
- `make format-check → lint → typecheck → test → integration → check-generated → build → up → seed → e2e` (ordered).
- **Крос-зрізна зміна спільного CSS/примітиву (S0–S2) → ПОВНИЙ e2e**, не таргетний.
- **S3 (lesson-editor) теж під ПОВНИЙ e2e, якщо чіпає reader-спільний селектор** `.lesson-prose` /
  `:is(.lesson-document,.lesson-editor-input)` (`main.css:1361+`) — editor↔reader coherence майже напевно це робить;
  інакше в тикеті S3 явно довести, що жодне reader-shared правило не зачеплено (тоді таргетний e2e ок). *(Courses 2)*
- **Browser-verify кожного зрізу — на реальному коді** (прод-шлях), не на прототипі.
- ⚠️ **CI на білінгу** (job `verify` не стартує) — зелений CI не покладатися; хто мерджить, кладе локальний run-id у звіт.
- Код — Courses 1 через Codex (тикети = зрізи S0–S6, MUST = дельти + pain-фікс сторінки). Рев'ю зрізів — Courses 2.

## Рішення (прийнято власником 2026-10-14, «Го»)

1. **Primary-кнопка = поглибити заливку** `color-mix(in srgb, var(--color-primary) 86%, #000)` = `#4763d0`
   = 5.30:1 для normal-тексту (наш уже вирішений gotcha). Тримаємо нашу button-шкалу `.8125rem/650`; НЕ
   роздуваємо primary-текст. Градієнт на кнопці — білий текст верифікувати вручну (axe пропускає).
   *(Провенанс: власник підтвердив напряму в цьому вікні 2026-10-14, попри паралельний large-text-сигнал з іншого
   вікна; fill — фінал. Альтернатива large-text/яскравий #5273f2 з guardrail «усі admin-primary ≥large-text» — закрита.)*
2. **Нові `--space-*` токени:** завести контейнерні (`--space-admin-sidebar/-canvas/-dialog/-prose-measure/
   -prose-min/-tile-min/-icon`) у `tokens.css`/`tokens.admin.css`; мисномери прибрати — `--space-border`→`1px` інлайн,
   `--space-focus-ring/-offset`→наш наявний `3px` focus-outline, `--space-callout-rule`→лишити як accent-rule.
3. **Browser-verify — на ПРОД-зрізі, не на прототипі** (Courses 2): renderer завис на прототипі, і це ок —
   доводити 375/200%/forced-colors на прототипі = доводити фейк (докази рахуються, коли пройшли прод-шлях). Кожен
   зріз верифікується в браузері на реальному коді за своїм гейтом (shell→S0, таблиці 375px→S1, …). Прототип —
   референс вигляду, не предмет гейта.

Тикети Codex по зрізах — `docs/admin-restyle/tickets.md`.
