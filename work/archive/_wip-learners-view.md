# V-track діагностика — view 5 `#learners` (LearnerOverview)

**Гілка:** `feat/learners-view-designpass` (від main @403ea02). **Автор:** Courses 1. **Старт:** 2026-10-17.
Еталон методу: `_wip-courses-view.md` (view 2), `_wip-analytics-view.md` (view 4).

## ФАКТ: спільний view (важливо)
`presentation/teach/LearnerOverview.tsx` рендериться **ОБОМА** маршрутами:
`/admin/learners` (AdminShell → `[data-surface='admin']`) і `/teach/learners` (instructor surface).
БЕЗ surface-пропа. Будь-яка layout-зміна б'є обидві поверхні → копія/розмітка мусить пасувати обом,
або свідомо admin-scoped. Лінки ведуть на `/teach/courses/$courseId` завжди (не чіпати — behavior).

## ФАКТ: структурна відмінність даних (не лише layout)
- **proto** `fixtureLearners`: картка = **один** learner-курс пласко (name, email, 1 course, 1 progress, 1 status).
- **prod** `LearnerRow`: `{userId, displayName, lastActivityAt, courses: LearnerCourse[]}` — картка =
  **одна особа з N курсами** (вкладений `.learners-courses` список; status+progress **на курс**, не на особу).
- Немає поля `initials` — вивести з `displayName` клієнтом; `displayName===''` (видалений акаунт, Lo07) → fallback `ratings.formerAuthor`.
→ «Повна fidelity» = накласти **візуальну мову** proto на **багатшу** multi-course структуру prod, не зводячи до single.

## ФАКТ: proto `#learners` композиція (src.jsx case 'learners')
- `PageHeader` eyebrow **«LEARNING EXPERIENCE»** + title «Студенти» + **description** + **actions** = `Pill info` «N студентів».
- `filters('Search learners', ['Active','Blocked'])` — спільний admin filter-bar (icon-search + status select + reset).
- `.learner-grid` — **2-колонки** (`repeat(2,minmax(0,1fr))`, →1col ≤76rem) з `.learner-card`.
- `.learner-card` (ui-card, pad space-5, grid gap-5):
  - `.learner-header` (flex, align-start, space-between, wrap): `.person-cell` (`.admin-avatar` initials + name h2 + email small) + status-pill.
  - `.learner-enrollment` (grid gap-2): `.admin-eyebrow` «Зарахування» + `<a>` course-link (arrow icon inline-end).
  - `.learner-progress-label` (flex space-between): «Progress» + `<strong>` percent.
  - `.ui-progress` track (space-2) + filled span.
  - `.learner-meta` (flex space-between, **top border-rule**): clock-icon + «Last activity» + `<strong>` date.

## ФАКТ: prod `LearnerOverview` композиція (207 рядків)
- `<h1>` лише — **без eyebrow/description/count-pill/PageHeader**.
- `<search className="learners-filters">` form: ui-field search (без icon) + Select state + submit + optional clearCourse.
- `.learners-list` — **1 колонка** завжди (grid gap-4).
- `.learner-card`:
  - `.learner-header` (grid gap-2): h2 name + `<p>` «Last activity: <time>». **Без avatar/person-cell/status у header.**
  - `.learners-courses` — список `<li>` на курс: `.learner-enrollment` (Link + per-course status pill success/danger) + `ProgressBar` (лише aria-label, **без видимого percent-рядка**).
  - **Без `.learner-meta` footer** (last-activity у header).
- Load-more Button (prod має пагінацію; proto — ні).

## РОЗБІЖНОСТІ (layout-мова proto, якої бракує прод) — ранжовано
| # | proto | prod | обсяг | ризик |
|---|---|---|---|---|
| D1 | PageHeader (eyebrow+title+**description**+count-Pill) | bare `<h1>` | малий (markup+i18n) | низький; спільний view → нейтральна копія пасує admin+teach |
| D2 | `.learner-grid` 2-col (→1 ≤76rem) | 1-col завжди | малий (CSS) | низ.-сер.; prod-картки вищі (multi-course) — 2-col ок, перевірити що не тісно |
| D3 | header: avatar(initials)+person-cell + status-pill | h2+lastactivity, без avatar | середній | initials з displayName; status у prod **per-course**, не per-person |
| D4 | per-enrollment `.learner-progress-label` (label+percent strong) | ProgressBar лише aria | малий (markup) | низький |
| D5 | footer `.learner-meta` (clock+lastactivity+date, top-rule) | lastactivity `<p>` у header | малий (markup move) | низький; prod per-person → один meta-footer на картку |
| D6 | filters = спільний `.admin-filters` (icon-search+select+reset) | custom `.learners-filters` flex | середній | чіпає спільний filter-pattern (views 2/6) — межа слайсу |

**Напрям:** серце fidelity-гапу — **картка** (D1–D5). D6 (спільний filter-bar) чіпає патерн поза цим view
(courses/users) — тримати за межею слайсу або мінімальний полиш. Візуальну мову proto накладаю на multi-course
структуру prod (не зводжу до single). Копія нейтральна (пасує admin+teach). Status — per-course (як у prod).

## РІШЕННЯ scope (2026-10-17, Courses 1, вибір власника «C»)
Повна card-fidelity **D1–D5 + D6** (filter-bar під спільний `.admin-filters`). Закриває альтернативи
A (лише PageHeader+grid) і B (без filter-bar). Копія нейтральна (admin+teach), status per-course (як prod).
Спосіб перевірки: браузер desktop+mobile обох маршрутів + axe@200% + ордерний make check.

## РЕАЛІЗОВАНО
- **D1** — `<h1>` → `PageHeader` (eyebrow `teach.learners.eyebrow` + title + description `lede`). Додав опційний
  `titleId` у PageHeader (layout.tsx), щоб секція лишалась `aria-labelledby="teach-learners-heading"`
  (e2e learning.spec:161 шукає цей id; lang-незалежно). Analytics PageHeader не чіпає.
- **D2** — `.learners-list` → `.learner-grid` 2-col (`repeat(2,minmax(0,1fr))`, →1col `@media max-width:76rem`).
- **D3** — header: `.person-cell` (`.admin-avatar` initials з displayName + h2). `initialsOf()` helper;
  `displayName===''` → `ratings.formerAuthor` (Lo07). Status лишається **per-course** (pill у course-рядку, як prod).
- **D4** — вже закрито компонентом `<ProgressBar>` (рендерить `.ui-progress-label` label+valueText);
  лейбл змінив з `enrollment.progress` на новий `progressLabel`=«Прогрес» (proto: label+value окремо).
- **D5** — `.learner-meta` footer (ClockGlyph + «Остання активність» + дата/`neverActive`, top-rule); винесено з header.
- **D6** — filter `<search className="admin-filter-bar">` + `.admin-user-search` форма + state `Select` (сиблінг) +
  optional clearCourse — вирівняно під спільний прод-патерн (views 2/6). Прибрав мертві `.learners-filters`/`.learners-list` CSS.
- **count-pill ОПУЩЕНО** (D1 actions): `LearnerPage` не має total (лише entries+nextCursor) → «N студентів»
  брехало б при hasNextPage. Чесно, як analytics. Нові inline-glyph (ArrowGlyph/ClockGlyph) — нема `Icon`-компонента.
- Копія нейтральна (admin+teach). Blast radius: LearnerOverview + layout.tsx(PageHeader titleId) + 2× locale + LearnerOverview.test.

## ВЕРИФІКАЦІЯ (проміжна)
- format/lint(арх.межі 210)/typecheck = 0; unit **LearnerOverview 7/7** (оновив Lo10-асерт: progressbar name=«Прогрес» + aria-valuenow=42).
- Браузер (Playwright host→:8080, admin, фронт ребілджено `make up`): desktop 1440 — PageHeader+опис, filter-bar,
  **2-col grid (gridCols «564px 564px»)**, avatar+ім'я, «КУРСИ» eyebrow, per-course link+status-pill, «Прогрес»+%, meta-footer.
  mobile 390 — чистий single-column reflow. **390+200%zoom: overflow=false** (scrollW=innerW=390). Скрини: scratchpad.
- Ордерний `make check`: **біжить** (лог frontend/test-logs/learners-gate-*.log).

## ОБМЕЖЕННЯ
- Seed-резидуальність: одна особа («Студентка Курсу») зарахована у ~34 e2e-курси → дуже висока картка в 2-col
  grid (lopsided). Артефакт накопичених e2e-даних, НЕ дефект лейауту; реальна особа має кілька курсів. Картка
  коректно рендерить N курсів списком із роздільниками.
- Teach-поверхня (`/teach/learners`): не знімав візуально (нема plaintext instructor-пароля локально);
  функціонально+a11y її покриває e2e learning.spec Lo11 (axe light+dark + keyboard + 44px на teach-маршруті).

## Наступне
- Дочекатись `make check` exit; записати e2e-рядок. Відкрити PR (від main). Перенести summary у Obsidian.
