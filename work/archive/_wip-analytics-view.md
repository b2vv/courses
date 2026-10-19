# V-track діагностика — view 4 `#analytics` (AnalyticsView)

**Гілка:** (створити від main). **Автор:** Courses 1. **Старт:** 2026-10-17.
Еталон методу: `_wip-courses-view.md` (view 2), `_wip-course-view.md` (view 3).

## ФАКТ: proto `#analytics` композиція (src.jsx case'analytics')
- `PageHeader` eyebrow **«PLATFORM STATISTICS»** + title «Аналітика» + **description** («Повна картина…»)
  + **actions** = `.ui-segmented` toggle **Last month / This month** (demo range).
- `.analytics-period`: clock-icon + діапазон дат + `Pill info` «Demo data» (pill зсунуто `margin-inline-start:auto`).
- 3× `.analytics-group`, кожна: `.group-heading` (**h2 + muted `<span>` підзаголовок** на flex-baseline) + `.stat-grid`
  (3-col, `repeat(3,minmax(0,1fr))`, mobile→auto-fit tile-min) з 3× `.ui-stat`:
  - **Platform** (Рух і результат): Revenue, Enrollments, Conversion + `.ui-card.ui-chart`:
    `.chart-heading` (eyebrow «Дохід» + h3 + `.chart-legend` summary-dot UAH); `.chart-layout`
    (grid `auto minmax(0,1fr)`: `.chart-y` 4 підписи осі 3000/2000/1000/0 + svg) з **area-fill**
    (`.chart-area` fill-opacity .09) + **4 grid-лінії** (`.chart-grid`) + line; `.chart-x` 5 підписів дат.
  - **People** (Навчання, що триває): Users, Active learners, Course completions.
  - **Content** (Від ідеї до уроку): Courses, Lessons, Drafts.
- `.ui-stat` (proto): `>div` (label + **icon-tile** на space-between) + `>strong` (**display-font**
  `clamp(1.5,,2rem)`, tabular-nums) + `>small` (trend, success-colored delta span).

## ФАКТ: prod `AnalyticsView.tsx` композиція (323 рядки) — БАГАТША за proto
- `section.admin-analytics` (grid gap-6).
- Заголовок: `.admin-list-heading` = `.eyebrow` + `<h1>` — **без description, без PageHeader-компонента**.
- Топ-рядок `.admin-analytics-grid` (4 StatTile: totalAccounts, published, revenue, openReports) — **без group-heading**.
- 8 `section.admin-analytics-group`, кожна `<h2>` (з **border-block-start rule** зверху, display-font) + `.admin-analytics-grid`:
  People (8), Content (3), Access (6), Engagement (3), Operations (5), Moderation (2),
  Instructors (**leaderboard `Table`**), Trend (**`SegmentedControl` metric-switcher + `TrendChart`**), Deferred (`EmptyState`).
- `StatTile` (prod `.ui-stat`): `.ui-stat-label.ui-mono` (зверху) + `.ui-stat-row>.ui-stat-value`
  (**mono-font** `clamp(1.75,4vw,2.5rem)`) + optional `.ui-stat-meta`. **Без icon-tile, без trend/delta** (крім опц. children `<Delta/>`).
- `TrendChart` (prod `.ui-chart`): `figure` + `figcaption.ui-chart-title` + svg (**одна baseline-вісь** + line/bars) + visually-hidden `<ul>`.
  **Немає: area-fill, grid-ліній, y-підписів, x-підписів, legend, card-heading.**

## РОЗБІЖНОСТІ (layout-мова proto, якої бракує прод) — ранжовано
| # | proto | prod | обсяг | ризик |
|---|---|---|---|---|
| D1 | `PageHeader` (eyebrow+title+**description**) | `.admin-list-heading` (eyebrow+h1, без опису) | малий (markup+i18n) | низький; як view 2/3 |
| D2 | `.group-heading` (h2 + muted підзаголовок, flex-baseline, **без top-rule**) | `h2` з `border-block-start` rule | малий (markup+CSS) | низький; підзаголовки = копірайт |
| D3 | `.ui-stat` display-font value + **icon-tile** у label-row + trend-delta | mono value, без icon, без trend | серед. | **СПІЛЬНИЙ примітив** StatTile — scope свідомо (gotcha S4/S2b/S3) |
| D4 | `.ui-chart` card: chart-heading+legend, chart-y/chart-x осі, **area+grid**, line | мінім. decorative line/bar, без осей/area/grid/legend | **великий** | **СПІЛЬНИЙ примітив** TrendChart; y/x-підписи = реальна шкала |

**Напрям:** D1/D2 — чистий bounded markup/CSS у AnalyticsView (як view 2/3). D3/D4 чіпають **спільні примітиви**
(StatTile/TrendChart вживані поза admin) — тут повторюваний клас admin-restyle (admin-специфіку тягнуть у
глобальний примітив). Будь-яка візуальна зміна примітива — або `[data-surface='admin']`-scope, або свідома
глобальна редефініція, НАЗВАНА так (не «scoped»).

## РІШЕННЯ scope (2026-10-17, Courses 1, вибір власника «C»)
Повна fidelity: **D1+D2+D3 (admin-scoped) + D4-full**. Закриває альтернативи A (лише layout) і
B (без контракт-зміни chart). Спосіб перевірки: нижче — браузер desktop+mobile + e2e axe@200%.

## РЕАЛІЗОВАНО
- **D1** — `.admin-list-heading` → `PageHeader` (eyebrow+title+**description** `adminAnalytics.lede`).
- **D2** — кожна з 9 груп обгорнута в `.group-heading` (h2 display-font + muted-підзаголовок, baseline-flex,
  БЕЗ top-rule); топ-рядок дістав групу «Платформа/Коротко»; нові i18n `sectionPlatform`+8×`sub*` (en+uk).
  CSS `.group-heading` у tokens.css. Підзаголовок — `--color-muted` (не faint: gotcha контрасту).
- **D3** — `[data-surface='admin'] .ui-stat-value` → display-font `clamp(1.5,,2rem)` tabular-nums
  (tokens.admin.css). Icon-tile НЕ додано (30+ метрик, потребує мапінгу). Base-примітив лишається mono.
- **D4-full** — `TrendChart` переписано: `.ui-chart-heading` (eyebrow+h3-title+legend-dot), `.ui-chart-layout`
  (chart-y 4 підписи осі зі шкали `niceCeil` + plot + chart-x sampled-дати), area-fill + grid-лінії + line/bars.
  Нові ОПЦІЙНІ пропси `eyebrow`/`unit` (контракт не ламає наявні тести). `.ui-chart` card-treatment —
  admin-scoped (tokens.admin.css). Mobile-редукції + `.ui-chart-x{overflow:hidden}` backstop від overflow@200%.
  AnalyticsView TrendSection: `formatValue` (копійки→грн для revenue, лічильник інакше) + `unit` для revenue.
- Контракт-зміна blast radius: лише AnalyticsView (консюмер) + ui.test.tsx + AdminKit.stories — усі зелені.

## ВЕРИФІКАЦІЯ
- format/lint/typecheck ✓; frontend unit **772 passed** (вкл. ui.test TrendChart line/bar/empty/axe).
- Браузер (MCP_DOCKER, admin-сесія, frontend ребілджено): desktop 1440 — PageHeader+опис, 8 group-headings
  з підзаголовками, raised stat-cards display-font, chart-card з y-віссю/segmented; mobile 390 — чистий
  single-column reflow; **390+200%zoom: `overflowed:false`** (docScroll=client=390). Скрини: `/tmp/analytics-{desktop,mobile}.png`.
- Повний ордерний `make check`: **exit=0** (2026-10-17). e2e **121 passed (15.1m)**, вкл. обидва
  `admin-analytics.spec.ts` (Rt-analytics admin reads + axe wcag2a/aa/21aa @390×200% + overflow:false;
  non-admin 403). Лог: `frontend/test-logs/api-nginx-20261017T123800Z.log`.

## ОБМЕЖЕННЯ
- Chart при **1 точці даних** (поточний seed — лише сьогодні) малює осі+grid, але не лінію (лінію з однієї
  точки не провести; `.ui-chart-line` рендериться при ≥2 точках — підтверджено ui.test). Не дефект лейауту,
  артефакт даних; реальне багатоденне вікно дає лінію. Маркери точок не додавав (`preserveAspectRatio=none`
  спотворив би circle в ellipse).
- Icon-tile у stat-tiles (proto має) не перенесено — потребує мапінгу іконки на 30+ метрик; відкладено.

## Наступне
- Дочекатись `make check`; записати e2e-рядок. Відкрити PR (від main).
