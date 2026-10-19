# S6 — reports + refunds restyle · Codex brief + working file

РОБОЧИЙ ФАЙЛ. Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого
інструменту: рядок `ДАЛІ:`. Далі ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінальна відповідь = переказ файлу.

## Завдання
Останній зріз фази admin-restyle. Дві адмін-черги — скарги учасників і повернення — сьогодні
функціональні, але сирі: `<article>` без класу, дамп рядків «Label: value», стан плоским `<p>`, дії
відкинуті у фіксовану колонку праворуч. Перетворити кожен елемент на структуровану `.ui-card` панель.

## Файли (і лише вони + CSS)
- `frontend/src/presentation/admin/ParticipantReportQueue.tsx`
- `frontend/src/presentation/admin/RefundQueue.tsx`
- `frontend/src/app/styles/main.css` — НОВИЙ page-scoped CSS (admin/queue-префікс).
Unit-тестів у цих компонентів НЕМА; рестайл візуальний — нових не треба. Домен/роут/ports не чіпати.

## Поточний стан (факти, звірено 2026-10-15)
Обидві черги мають ІДЕНТИЧНУ структуру-болячку:
- `<section className="admin-user-list">` → heading (eyebrow+h1) → loading/error/actionError → empty
  (`<EmptyState title=.../>`) → `<ul className="teach-section">` (МІСНЕЙМ) → `<li>` → `<article>` (БЕЗ класу):
  - `<h2>` (reports: kind; refunds: reason);
  - дамп `<p>`: reports — reporter/target/course/reason/state; refunds — course(courseId)/user(userId)/ordered/charged;
  - reports має ще плоский `<p>{state}</p>`;
  - `<time>`;
  - `<div className="admin-user-actions">` з діями: reports — Resolve + умовний Revoke-`Dialog` (лише `kind==='on_student'`);
    refunds — умовний Revoke-`Dialog` (лише `membershipState active|blocked`) + Resolve.
- Reports-черга має пагінацію (Load more); refunds — ні.

## Доменні типи (точні — для pill-мапи, уникнути неповного enum)
- `ParticipantReportState = 'open' | 'resolved'` (повний). `ParticipantReportKind = 'on_lecturer' | 'on_student'`.
- `RefundQueueItem`: НЕ має status-поля. Є `reason: 'settlement_rejected'|'grant_conflict'|'unknown'`,
  `membershipState: 'active'|'blocked'|'revoked' | null`, orderAmountMinor, chargedAmountMinor|null, receivedAt|null.

## Дизайн-ціль (proto `proto/src.jsx:reports/refunds`, `proto/styles.css`) — ПЕРЕКЛАСТИ на наш DOM
- **Список:** `<ul>` з новим page-класом (напр. `admin-queue-list`), grid gap. ЗАМІНИТИ `className="teach-section"`
  (НЕ стилізувати teach-section — спільний із CourseRoster/comments/moderation).
- **Картка:** `<article class="ui-card admin-queue-card">` (СПОЖИВАННЯ `.ui-card`, не редефініція), padding.
- **Heading-зона** `admin-queue-heading`: flex wrap gap — `.ui-icon-tile` (СПОЖИВАННЯ, inline-svg child + `aria-hidden`,
  як CatalogView) + колонка eyebrow(напр. id/orderId · дата) та `<h2>` + **статус-pill** праворуч.
- **Структуровані деталі** `admin-queue-details` (ЗАМІСТЬ дампу `<p>`): grid label/strong пар
  (`<div><span>Label</span><strong>value</strong></div>`), border-top hairline, padding-block; narrow → 1 колонка.
  Reports: reporter / course / target. Refunds: course + amount (`admin-refund-summary`/`admin-refund-amount`
  tabular-nums) + charged.
- **Статус-pill** (`.ui-pill ui-pill--{variant}`, СПОЖИВАННЯ):
  - reports `state`: `open`→warning, `resolved`→success (повний, 2 значення).
  - refunds `membershipState`: `active`→success, `blocked`→warning, `revoked`→neutral, `null`→БЕЗ pill
    (повний розбір усіх 4 випадків; не впасти як S5-revoked). Текст — наявний i18n-лейбл стану; не color-only.
- **Дії** `admin-queue-actions` (ЗАМІСТЬ `admin-user-actions`): flex, border-top hairline, padding-top →
  дії ПРИВ'ЯЗАНІ до картки знизу, НЕ фіксована колонка праворуч. Кнопки рівномірно (flex:1 або wrap).
  Зберегти ТОЧНО наявні дії + gating + handlers (resolve/revoke Dialog). Нічого з proto-only НЕ додавати
  (без QueueFilter, Details, example-tools — то фікстури прототипу).
- **Empty-state структурований (icon+рамка):** page-scoped панель (напр. `admin-queue-empty`) — `.ui-card` +
  `.ui-icon-tile`(svg) + `<h2>`(title) + `<p>`(hint). ЗАМІНИТИ голий `<EmptyState title=.../>`. НЕ міняти
  спільний компонент `EmptyState`/`.ui-state`. Якщо треба i18n-hint — додай ключ у наявний
  `enrollment.reports.*`/`enrollment.refunds.*` namespace (звір існування, не вигадуй).

## ЖОРСТКІ обмеження (пастки фази — не повтори)
- **НЕ чіпати базу `.admin-user-actions`** (main.css:911) — UserListView залежить від її `display:flex`+`gap`
  (правило :807 їх НЕ передекларовує). Дати чергам ВЛАСНИЙ клас дій; `.admin-user-actions` лишити як є.
- **НЕ стилізувати `.teach-section`** (спільний теач-клас) — замінити на новий queue-клас у розмітці.
- **НЕ редефінити й НЕ міняти** спільні примітиви/токени: `.ui-card`, `.ui-pill*`, `.ui-icon-tile`,
  `.ui-state`, `EmptyState`, будь-що в `tokens.css`/`tokens.admin.css`. Лише СПОЖИВАТИ класи. Весь новий
  CSS — page-scoped (admin/queue-префікс). S6 не S0–S2 → жодної глобальної редефініції.
- Усі вжиті CSS-токени — звір grep'ом, що ВИЗНАЧЕНІ (undefined var гейт не ловить — пастка S3).
- **E2e-safe (критично):** зберегти `<article>` елемент (e2e `locator('article').filter({hasText})`),
  усі тексти полів усередині article (reporter/target/course/reason/amount — hasText фільтрує), лейбли
  кнопок дослівно (`Вирішити`/resolve, `Відхилити`/revoke), Dialog trigger/confirm, heading-тексти.
  Нічого не ховати, що e2e шукає. Таргетні: enrollment.spec En07/En14 (/admin/participant-reports),
  refund.spec Rf-e2e.
- `exactOptionalPropertyTypes` (conditional spread). Семантичні токени. A11y: 44px контроли, видимий
  фокус, light/dark, 200% zoom без h-scroll, 375px без зламів.

## Self-verify (ти, Codex)
1. `cd frontend && npm run format`; 2. `npx tsc --noEmit`; 3. `npx oxlint --deny-warnings` +
`node scripts/check-architecture.mjs`; 4. `npx rstest run` на будь-яких наявних тестах, що імпортують ці
компоненти (якщо є) — усі зелені. НЕ запускай make/e2e/docker — повний гейт прожене колега.

## Поза скоупом
Домен/роут/ports/API/інші екрани; нові залежності; нові токени; proto-only фічі (filter/details).

---

## Журнал (Codex дописує нижче)
ДАЛІ: Прочитати повний бриф S6, план і поточні компоненти; виконати лише рестайл двох черг та дозволені self-verify перевірки.
ФАКТ: Бриф S6 прочитано повністю; обидві черги відповідають описаному DOM. Наявні сторонні зміни не чіпаю. Зберігаю article, поля, пагінацію та всі обробники/умови Dialog; змінюю лише представлення.
ДАЛІ: Звірити локалізацію, CSS-токени та наявні тести; додати локальні класи карток, станів і порожніх панелей.
ФАКТ: Тестів із імпортами ParticipantReportQueue/RefundQueue немає. Локалі reports/refunds існують, emptyHint у них відсутній; додам лише ці два ключі в uk/en (дозволений брифом виняток). Базові ui-card/ui-pill/ui-icon-tile та спільні teach/admin-user-actions залишаються незмінними.
ФАКТ: Розмітку обох черг перетворено на ui-card: SVG-іконка, ID/дата, heading і текстовий pill, структуровані поля, нижні дії; empty-панелі мають іконку, заголовок і hint. Обробники та gating не змінені.
РІШЕННЯ (Codex, 2026-10-15): У локалях немає тексту стану revoked (перевірено пошуком revoked/відкли); додано enrollment.refunds.revoked у uk/en, active/blocked використовують наявні roster.state. Повноту мап перевіряє satisfies Record з доменними типами; null явно без pill.
ДАЛІ: Додати page-scoped responsive CSS і виконати self-verify.
ФАКТ: Додано лише admin-queue/admin-refund CSS: grid-картки, hairline-розділювачі, wrap дій, 44px мінімум кнопок, одна колонка на narrow, перенесення довгих ID/текстів. Усі 8 використаних токенів визначені в tokens*.css (автоматично звірено, undefined: []).
ДАЛІ: npm run format → tsc --noEmit → oxlint + check-architecture; перевірити diff і локалізаційні тести.
ФАКТ: Перший виклик npm run format помилково зроблено з кореня репо: ENOENT package.json, перевірка не стартувала. Повторюю з frontend.
ФАКТ: npm run format з frontend — PASS (341 файл).
ФАКТ: Diff перевірено: усі старі значення полів збережені всередині article; handlers, Dialog і пагінація незмінні. Форматування не зачепило сторонні вихідні файли; main.css має лише новий page-scoped блок.
ФАКТ: npx tsc --noEmit — PASS (exit 0).
ФАКТ: npx oxlint --deny-warnings і node scripts/check-architecture.mjs — PASS; архітектурні межі перевірено для 208 файлів.
ДАЛІ: Хоча тестів-імпортерів черг немає, запущу наявний i18n.test.tsx через додані переклади.
ФАКТ: rstest i18n у sandbox не стартував: listen EPERM ::1:3000. Це обмеження відкриття локального порту, не падіння тесту; повторюю той самий таргет з розширеним дозволом.
ФАКТ: Повторний npx rstest run src/app/providers/i18n.test.tsx — PASS, 4/4, без skipped. Нових тестів не додано.
ФАКТ: git diff --check — PASS. S6 реалізовано в двох компонентах + новому page-scoped CSS; у uk/en лише emptyHint×2 і відсутній revoked-лейбл. Спільні примітиви/токени, домен/роути/порти та сторонні зміни не редагувались.
ФАКТ: make/e2e/docker не запускалися. 375px, 200% zoom, light/dark та браузерні сценарії не заявляю перевіреними: виконано CSS-підготовку і статичні перевірки, браузерний доказ ще потрібен.
ДАЛІ: Колезі виконати браузерну перевірку обох черг (empty/populated, усі стани, Dialog, narrow/zoom) і повний гейт, зокрема En07/En14 та Rf-e2e.
