# S5 — learners restyle · Codex brief + working file

РОБОЧИЙ ФАЙЛ. Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого
інструменту: рядок `ДАЛІ:` з тим, що збираєшся робити. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ цього файлу.

## Завдання
Візуальний рестайл екрана learners (LEARNING EXPERIENCE) за admin design-system restyle, зріз S5.
Екран сьогодні функціональний, але сирий: класи `learners-list` / `learners-courses` /
`learners-filters` / `teach-section` НЕ мають жодного CSS — рядки злипаються, статус-бейдж
впритул до назви курсу, прогрес виглядає волосиною в тісному контексті.

## Файли (і лише вони)
- `frontend/src/presentation/teach/LearnerOverview.tsx` — компонент (той самий для `/admin/learners`
  і `/teach/learners`). Додати класи / мінімальні зонні обгортки. НЕ міняти модель даних і не
  ламати рольові/текстові селектори (тести асертять за `getByRole`/`getByText`, не за класами).
- `frontend/src/app/styles/main.css` — НОВИЙ page-scoped CSS для learners. Додавати в кінець або
  біля інших teach/admin-правил.
- За потреби `frontend/src/presentation/teach/LearnerOverview.test.tsx` — лише якщо додаєш новий
  спостережуваний селектор; наявні тести мають лишитися зеленими без правок.

## Поточний стан (факти, звірено 2026-10-14)
- Компонент рендерить:
  - `<section class="teach-section" aria-labelledby="teach-learners-heading">` → `<h1>`.
  - `<search class="learners-filters">` з формою (пошук `ui-input`, `Select` стану, кнопки).
  - `<ul class="learners-list">` → `<li>` → `<article>` (БЕЗ класу) на кожного учня:
    - `<h2>` ім'я; `<p>` «Остання активність» + `<time>`/«ніколи»;
    - `<ul class="learners-courses">` → `<li>` на КОЖЕН курс учня:
      `<Link>` назва курсу · `<span>` стан (`active`/`blocked`) · `<ProgressBar>` (рендерить
      `.ui-progress`/`.ui-progress-label`/`.ui-progress-track`/`.ui-progress-fill`).
  - Кнопка «Показати ще» (`isPending`).
- `LearnerRow`: `userId`, `displayName`, `lastActivityAt: string|null`, `courses: LearnerCourse[]`.
  `LearnerCourse`: `courseId`, `slug`, `title`, `state: 'active'|'blocked'`, `progressPercent: number`,
  `enrolledAt`. **УВАГА: на одного учня — МАСИВ курсів** (на відміну від proto, де 1 курс на картку).
  Переноси ВІЗУАЛЬНІ правила proto на наявний DOM; НЕ згортай до одного курсу.

## Дизайн-ціль (proto `docs/admin-restyle/proto/src.jsx:31`, `proto/styles.css`)
Переклади ЦІ правила на наш DOM (proto не можна копіювати 1:1 — інша структура):
- Кожен учень = **картка** (фон-заливка + hairline + тінь — як `.ui-card`). Можна додати класи
  `ui-card learner-card` на `<article>` (споживання `.ui-card`, НЕ редефініція).
- **Розділені зони** в картці, з вертикальним ритмом між ними (не злипаються):
  (1) header — ім'я + остання активність; (2) курси.
- Для КОЖНОГО курсу (елемент `learners-courses li`): **назва-лінк і статус-бейдж — окремі зони з
  gap** (фікс «…Active» впритул). Статус-бейдж — через СПОЖИВАННЯ `.ui-pill` з семантичним
  модифікатором (`active`→success, `blocked`→danger), як у `UserListView`
  (`ui-pill ui-pill--success`/`--danger`). Прогрес під ними з видимим треком.
- Рядки-курси **розділені дільниками** (border-block-start hairline між елементами), не злипаються.
- Прогрес — видимий ТРЕК, не волосина, у контексті картки.

## MUST (тикет S5)
1. Розділені зони enrollment/статус; фікс «…Active» впритул (додати gap + структуру).
2. Рядки-картки з дільниками (не злипаються).
3. `.ui-progress` трек замість волосини (видимий трек у контексті learners).

## ЖОРСТКІ обмеження (пастки цієї фази — не повтори)
- **Увесь новий CSS — PAGE-SCOPED під learners-селектор** (`.teach-section`, `.learners-list`,
  `.learner-card`, `.learners-courses` тощо). **ЗАБОРОНЕНО** міняти спільні примітиви/токени:
  `.ui-progress*`, `.ui-card` (можна лише СПОЖИВАТИ клас), `.ui-pill*`, `.ui-meter*`, `.ui-input`,
  будь-що в `tokens.css`. Це S5 (не S0–S2) → жодної редефініції глобального примітива. Повторювана
  пастка фази: admin/page-специфіку вставляють у глобальний примітив і звуть «scoped» — НЕ роби так.
- Якщо трек прогресу треба «потовстити» в контексті learners — роби це page-scoped override
  (`.learner-card .ui-progress-track { … }`), НЕ чіпаючи базовий `.ui-progress-track` у tokens.css.
- Семантичні design-токени (`--color-*`, `--space-*`, `--shadow-*`), без хардкод-кольорів/px-магії.
- A11y/responsive: клавіатурна навігація, видимий фокус, контроли ≥44px, light/dark, 200% text-zoom
  без h-scroll, 375px без зламів. Lo11 e2e вже перевіряє це на `/teach/learners` — не зламай.
- `exactOptionalPropertyTypes` (conditional spread, не `key: undefined`).
- Рольові/текстові селектори тестів недоторкані: `getByRole('heading')`, `getByRole('link')`,
  `getByRole('progressbar', {name})`, `getByText` стану, empty/error тексти, кнопка «Показати ще».
- Статус-бейдж: accessible-текст стану має лишитися читабельним (бейдж — це span із текстом стану,
  як зараз; просто з класом `.ui-pill ui-pill--{variant}`). Не ховай текст у `aria-hidden`.

## Self-verify (ти, Codex)
Після імплементації ОБОВ'ЯЗКОВО:
1. `cd frontend && npm run format` (oxfmt — інакше format-check завалить гейт).
2. `npx tsc --noEmit` — 0 помилок.
3. `npx oxlint --deny-warnings` на змінених файлах (і `node scripts/check-architecture.mjs`).
4. `npx rstest run src/presentation/teach/LearnerOverview.test.tsx` — усі зелені.
НЕ запускай `make check`/e2e/docker — повний ordered-гейт прожене Courses 1 (я) окремо.

## Поза скоупом
- Не чіпати route-файл `admin.learners.tsx`, домен, застосунок, API, інші екрани.
- Не додавати залежності. Не вводити нові токени в tokens.css.

---

## Журнал (Codex дописує нижче)
ДАЛІ: Прочитаю повністю бриф S5, план і поточний компонент; додам лише learners-scoped оформлення та виконаю чотири self-verify кроки без make/e2e/docker.
ФАКТ: Бриф S5 прочитано повністю. Перевірено план, поточні тести й стан дерева; наявні сторонні зміни зберігаю. S5 обмежений компонентом, page-scoped CSS і цим журналом.
ДАЛІ: Звірю наявні токени й примітиви та додам зонні обгортки без зміни поведінки.
ФАКТ: Підтверджено відсутність learners-CSS; прототип має іншу структуру курсів. Наявні ui-card/ui-pill споживатиму без редефініції, трек зміню лише через .learner-card. Нових поведінкових селекторів не потрібно, тести не редагую.
ФАКТ: Додано ui-card learner-card, header учня та learner-enrollment для кожного курсу; active/blocked споживають success/danger pill із незміненим текстом. Новий CSS лише learners-scoped: відступи, дільники, flex-wrap, переноси довгих назв, мінімальна висота посилань із control-min-size, трек space-3. Модель, handlers, рольові/текстові селектори й спільні токени не змінено.
ДАЛІ: Запускаю format, TypeScript, oxlint + architecture та таргетний rstest у порядку брифа.
ФАКТ: Перший виклик npm run format помилково зроблено з кореня репо: ENOENT package.json, форматування не виконалося. Повторюю з frontend.
ФАКТ: npm run format у frontend — PASS (exit 0, 341 файлів).
ФАКТ: npx tsc --noEmit — PASS (exit 0, без помилок).
ФАКТ: npx oxlint --deny-warnings на змінених шляхах — PASS (exit 0); CSS форматовано oxfmt, oxlint перевіряє TSX.
ФАКТ: node scripts/check-architecture.mjs — PASS (exit 0, 208 файлів).
ФАКТ: Таргетний rstest у sandbox не стартував: listen EPERM ::1:3000. Це блокування локального сокета, не падіння тесту; повторюю ту саму команду з розширеним дозволом.
ФАКТ: Таргетний rstest після дозволу локального сокета — PASS: 8/8, без skipped, тестовий файл незмінений.
ДАЛІ: Перевіряю фінальний diff та межі змін; браузерні 375px/200%/light-dark і повний гейт лишаються перевіркою колеги, тут їх не заявляю пройденими.
ФАКТ: Фінальний diff перевірено: лише LearnerOverview.tsx і доданий learners-scoped блок main.css; форматування не внесло інших змін у frontend. git diff --check — PASS. Початкові сторонні зміни скриншотів не чіпав. Усі 3 MUST реалізовано; self-verify завершено. make/e2e/docker не запускав. Повний гейт і браузерний доказ залишаються колезі.
