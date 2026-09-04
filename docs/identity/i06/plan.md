# I06 — plan

How the slice specified in [spec.md](spec.md) gets built. Placement follows
`.claude/standards.md` §Розміщення; the reuse survey below is checked against its reuse
index.

This file absorbs the plan section of the eventual report: the report will be the delta
against it, not a second plan.

## Reuse survey

The code graph was built for this repository (672 nodes, 6911 edges) but its MCP tools are
not live in this session, so the survey is against the manifest's reuse index and the code
itself.

### Reused unchanged

| Need | Reused | Why it fits |
|---|---|---|
| Read the viewer's identity | `app/composition/identity-queries.ts` — `current()`, `sessionQueryKey`, `SessionQueryError` | Already request-scoped and never-fresh, which is exactly what a guard needs |
| Sign out | `application/identity/session.ts` — `logOut` | Written in I05 with no consumer; this slice is its first |
| Role policy | `domain/identity/identity.ts` — `canAccessAdministration` | Already mirrors the Go policy and is unit-tested; the guard reads it rather than re-deriving |
| Safe return target | `domain/identity/identity.ts` — `safeReturnTo`, `app/router/login-search.ts` | Written and tested in I05 for exactly this case |
| Empty state | `presentation/ui` — `EmptyState` | `/learn` needs a title and a child link, which is its whole surface |
| SSR cookie forwarding | `adapters/api/request-headers.ts`, `adapters/api/transport.ts` | The root loader uses the same transport as every other adapter |
| Refusal mechanism | `notFound()` from the router, as already used by `app/router/catalog-search.ts` | Decision 2 asks for a 404; this is how this repository already produces one |

### Added

| Thing | Where | Why nothing existing fits |
|---|---|---|
| Root identity resolution | loader on `routes/__root.tsx` | No route resolves identity today; the header cannot render server-side without it |
| Guard for a protected route | `app/router/guards.ts` | Two routes need the same three-way decision (resolving / anonymous / role). Writing it twice is the duplication this session already removed once from the API adapters |
| Signed-in header | `presentation/identity/SessionMenu.tsx` | The root `Layout` currently hardcodes a "sign in" link; the signed-in variant has different content per role |
| Learning empty state | `presentation/learning/LearningHome.tsx` | New screen; composes the reused `EmptyState` |
| Administration shell | `presentation/admin/AdminShell.tsx` | New screen; deliberately minimal, and must not imply the instructor scoping that belongs to a later phase |
| Identity fixtures | `backend/sql/seed-identity.sql` | Kept out of `seed.sql` so the guard below can refuse identities without refusing course data |
| Seed environment guard | `Makefile` `seed` target | Decision 6 |

### Deleted

`presentation/LockedArea.tsx` and the `locked.*` translations. Its copy — "this area is not
available yet" — becomes false the moment the areas work, and every one of its three call
sites is replaced by a real screen or a guard. Leaving it would leave a lie reachable.

## Design

**Identity at the root.** `__root.tsx` gains a loader that resolves identity through
`context.identity.current()`. Per decision 1 it runs only when the request carries a session
cookie; the check is isomorphic, reading the request header on the server and
`document.cookie` on the client. Absent cookie resolves to `null` without a call.

**The guard.** One helper consuming route context, returning one of three outcomes rather
than a boolean, because decision 7 makes "not yet known" a distinct state from "anonymous":

- identity not resolved → the route suspends; it never decides on incomplete information
- anonymous → `redirect` to `/login` with `returnTo` set to the current address
- resolved, capability denied → `notFound()`
- resolved, capability allowed → render

`/learn` passes the "signed in" capability, `/admin` passes `canAccessAdministration`. The
capability is a predicate the caller supplies, so the guard has no knowledge of roles.

**The header.** `Layout` renders `SessionMenu`, which reads the same `['session']` query the
guard uses, so the two can never disagree. Anonymous keeps today's "sign in" link. Signed in
shows the display name, a sign-out control and — only when `canAccessAdministration` — an
administration link. Sign-out calls the reused use case, replaces the session query data
with `null`, and navigates to the catalog.

**Fixtures.** Three identities in a separate `seed-identity.sql`, `ON CONFLICT DO NOTHING`
so repeated runs are inert. Password hashes are literal Argon2id PHC strings generated once
with the repository's own parameters, so seeding needs no Go toolchain.

## Сайд-ефекти

| Effect | Assessment |
|---|---|
| Every SSR page with a session cookie now makes one extra internal call to `/api/v1/auth/me` | Measured at 0–2 ms on the internal network in I05. Anonymous traffic — the majority on a storefront — is unaffected by decision 1 |
| SSR HTML becomes viewer-specific | Already covered: `frontend/server.mjs:23` sets `Cache-Control: private, no-store` on all SSR responses, and the audit disposition records that this is why no `Vary: Cookie` is required |
| `route-guards.spec.ts` asserts behaviour that this slice deliberately removes | The spec is rewritten in the same task that removes `LockedArea`, not left failing |
| Seeded credentials exist in the repository | Local-only by the same convention as `.env.example`, and decision 6 adds a machine guard against a non-development environment |
| The header appears on every page, including the login screen | Intended: a signed-in visitor opening `/login` already sees the "already signed in" state from I05 |

## План міграції

No schema change: `users` and `user_sessions` already exist from I01. The migration is data
only, and additive.

1. `seed-identity.sql` is a new file; `make seed` applies both it and `seed.sql`.
2. Existing local databases are unaffected until `make seed` runs; running it twice is inert.
3. CI seeds a fresh database on every run, so no ordering problem arises between environments.
4. No production environment exists, so there is nothing to migrate forward.

## Rollback

| If | Then |
|---|---|
| The guard misbehaves after merge | Revert the slice commit. The routes return to `LockedArea` only if that revert includes its deletion — so `LockedArea` is deleted in the same commit as the guards land, keeping the revert atomic |
| The root loader proves too costly | Remove the loader alone; the header falls back to rendering after hydration, losing A04's no-flash guarantee but nothing else. This is a one-file change |
| Seeded identities reach an unwanted environment | `DELETE FROM users WHERE email LIKE '%@courseforge.local'`; sessions cascade. The seed guard makes this recovery rather than routine |

Every step is reversible with a revert or a single delete. Nothing in the slice writes data
a user could lose.

## Конституція

| Rule | Check |
|---|---|
| Domain has no framework, transport or database dependency | The guard lives in `app/router/`, not in domain; it reads the existing domain predicate |
| Application depends on domain and ports only | No application change in this slice |
| Handlers and routes translate and render, business policy lives inward | The capability predicate comes from `domain/identity`; routes pass it, they do not encode it |
| No `any`, no assertions disguising unsafe input | `returnTo` continues through the tested `safeReturnTo`; the guard's outcome is a discriminated union |
| No fetch in presentation or routes | The root loader calls the composition layer, as the catalog route already does |
| Add dependencies only when used | None added |
| Client guards are not the security boundary | Stated in the spec's Out of scope, and the administration shell exposes no data |
| Do not claim unimplemented capability | The administration shell says management arrives in a later phase; `/learn` shows an empty state rather than a fabricated list |
| Semantic tokens, keyboard navigation, visible focus, >=44px controls, light and dark, 200% text zoom | **Initially violated — the first draft of this plan and its acceptance table said nothing about any of it for three new components.** Closed by the section below and by rows G37–G39 |

### Accessibility, added after the constitution check failed

The first pass through the manifest's constitution found this rule unaddressed. Three new
surfaces — the signed-in header with its sign-out control, the learning empty state and the
administration shell — had no accessibility requirement in the plan and no row in the
acceptance table, while `frontend/e2e/accessibility.spec.ts` already holds the rest of the
product to keyboard navigation, dark theme and 200% zoom.

What the slice must therefore satisfy:

- The sign-out control is a real button reachable in tab order, with the visible focus ring
  the UI kit already provides, and a target of at least 44px. It reuses `Button` from
  `presentation/ui`, which is React Aria and gives focus behaviour for free.
- The administration link is a `Link` from the same kit, in the same navigation landmark as
  the existing catalog link, so it inherits that landmark's semantics.
- The display name is text inside the existing `<nav>`, not a control, and is not
  announced as interactive.
- Every new string carries both locales, and no new colour is hardcoded: the components use
  the existing semantic tokens, so light and dark follow without per-component work.
- The header must survive 200% text zoom without the navigation overlapping, which is the
  bar `accessibility.spec.ts` already applies to the catalog.

## Приймальні сценарії

Delegated to `acceptance-spec`. The table is its output with two rows corrected by the
author, marked below, and one reported gap discarded as false.

The 36 rows collapse onto a small number of surfaces: a rewritten
`frontend/e2e/route-guards.spec.ts`, route and header component tests under
`frontend/src/**`, the Go HTTP contract test, a seed integration test, and `make smoke`.

| ID | Сценарій | Передумова | Дія | Очікуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| G01 | Анонім на `/learn` (розряджає A06 у частині `/learn`) | Cookie відсутній; фікстури засіяні | Ввести `/learn` в адресний рядок | URL стає `/login?returnTo=%2Flearn`; видно форму входу | Playwright; компонентний тест роута з MSW `401` |
| G02 | Анонім на `/admin`; підтверджує перегляд рішення 3 | Cookie відсутній | Відкрити `/admin` із закладки | URL стає `/login?returnTo=%2Fadmin`, статус 200, форма входу видима; 404 немає | Playwright; компонентний тест роута |
| G03 | Студент на `/learn` (розряджає A01) | Активна сесія `student` | Відкрити `/learn` | Порожній стан `EmptyState` з посиланням у каталог; ні редиректу, ні 404 | Playwright; компонентний тест |
| G04 | Студент на `/admin`; відмова за роллю = 404 (рішення 2, розряджає B01) | Активна сесія `student` | Відкрити `/admin` | Той самий екран «не знайдено», що й для довільної неіснуючої адреси; жодного натяку на існування адмінки | Playwright; компонентний тест; доменний тест `canAccessAdministration` |
| G05 | Інструктор на `/learn`; `/learn` про зарахування, не про роль (рішення 4) | Активна сесія `instructor` | Відкрити `/learn` | Той самий порожній стан, що й у студента; жодних «моїх викладацьких курсів» | Компонентний тест; Playwright |
| G06 | Інструктор на `/admin` (розряджає B01) | Активна сесія `instructor` | Відкрити `/admin` | Оболонка адміністрування з явним текстом, що керування зʼявиться пізніше; 404 немає | Playwright; компонентний тест |
| G07 | Адміністратор на `/learn` | Активна сесія `admin` | Відкрити `/learn` | Той самий порожній стан | Компонентний тест |
| G08 | Адміністратор на `/admin` (розряджає A02, B01) | Активна сесія `admin` | Відкрити `/admin` | Оболонка видима; URL лишається `/admin` | Playwright; компонентний тест |
| G09 | Гонка резолву на набраній адресі (рішення 7; дзеркало A04) | Сесія `admin`; `/auth/me` відповідає із затримкою ~1 с | Повне завантаження `/admin` | Поки ідентичність невідома — стан очікування; далі оболонка. Жодного кадру з 404 і жодного редиректу на `/login` | Playwright із затримкою на `/auth/me` і DOM-спостерігачем до гідратації; компонентний тест |
| G10 | Гонка резолву на `/learn` | Сесія `student`; `/auth/me` із затримкою | Повне завантаження `/learn` | Ані редиректу, ані 404 до відповіді; далі порожній стан | Playwright; компонентний тест |
| G11 | `/auth/me` віддає 503 на захищеній адресі | Сесія є, сховище недоступне | Відкрити `/admin` | Локалізований повторюваний стан помилки; ні редиректу, ні 404; повтор після відновлення відкриває оболонку | Компонентний тест з MSW `503`; Playwright `recovery.spec.ts` |
| G12 | Сесія завершилась або відкликана при відкритому табі (сценарій 6) | Ідентичність у кеші; сесію видалено на сервері | Перейти на `/admin` усередині SPA | Поведінка тотожна анонімній: `/login?returnTo=%2Fadmin`; ні зламаного екрана, ні застарілого імені в шапці | Playwright; компонентний тест з MSW `401` після `200` |
| G13 | Повний оберт `returnTo` для дозволеної ролі (розряджає A06, C01) | Cookie відсутній | `/admin` → логін як `admin` | 200 і `Set-Cookie`; браузер на `/admin` з видимою оболонкою, не на `/learn` | Playwright; тест роута `/login` |
| G14 | Оберт `returnTo` для забороненої ролі — роль перевіряється ПІСЛЯ входу (рішення 3) | Cookie відсутній | `/admin` → логін як `student` | Логін успішний, URL стає `/admin`, на ньому 404; назад на `/login` не повертає | Playwright; тести роутів |
| G15 | Ворожий і непридатний `returnTo` (розряджає C01) | Cookie відсутній | Логін із `returnTo`: `https://evil.example/`, `//evil.example`, `/\evil.example`, `%2f%2fevil.example`, `/admin%0d%0aSet-Cookie:x`, **`/login`**, **`/courses`**, порожній, відсутній | Після входу URL завжди `/learn` | Табличний юніт-тест `safeReturnTo`; компонентний тест `/login`; один ворожий кейс у Playwright |
| G16 | Шапка для аноніма не змінилася (рішення 1) | Cookie відсутній | Відкрити `/` | «Увійти» в шапці; ні імені, ні виходу, ні посилання на адміністрування | Rstest; Playwright `shell.spec.ts` |
| G17 | Шапка для студента (рішення 5, розряджає A04) | Сесія `student` | Відкрити `/` | Видно `displayName` і «Вийти»; посилання на адміністрування немає; «Увійти» немає | Компонентний тест; Playwright |
| G18 | Шапка для інструктора й адміністратора (рішення 5) | Сесія `instructor`, потім `admin` | Відкрити `/`, натиснути посилання адміністрування | В обох ролей імʼя, «Вийти» і посилання; клік веде на `/admin` і відкриває оболонку | Компонентний тест по ролях; Playwright |
| G19 | SSR рендерить ідентичність без миготіння (розряджає A04) — **скориговано автором** | Валідний cookie | Перезавантажити `/courses` | Сире HTML уже містить `displayName` і контроль виходу; відповідь має `Cache-Control: private, no-store`; спостерігач, встановлений до гідратації, не фіксує кадру з анонімною шапкою | Інтеграційний тест сирого HTML; Playwright з раннім спостерігачем |
| G20 | SSR не резолвить ідентичність без cookie (рішення 1) | Cookie відсутній | Запросити `/courses` | Анонімна шапка; у логах API за час запиту немає звернення до `/auth/me` | Інтеграційний тест SSR із лічильником; перевірка мережі |
| G21 | Ізоляція ідентичностей між SSR-запитами | Дві різні валідні сесії | Одночасно запросити `/courses` з cookie A і B | Кожна відповідь містить своє імʼя; чуже не потрапляє | Інтеграційний тест SSR (паралельні запити) |
| G22 | Вихід (розряджає A05) | Сесія будь-якої ролі, користувач на `/learn` | Натиснути «Вийти» | `POST /auth/logout` → 204, cookie протермінований; браузер на публічній сторінці; шапка анонімна; `/learn` знову веде на `/login?returnTo=%2Flearn` | Playwright; Go HTTP-тест cookie; компонентний тест |
| G23 | Вихід при недоступному сховищі | Сесія; PostgreSQL недоступний | Натиснути «Вийти» | 503, cookie все одно протермінований, локалізоване повідомлення; ідентичність зникла з кешу, `/learn` замкнений | Компонентний тест з MSW `503`; Go HTTP-тест; `recovery.spec.ts` |
| G24 | Історія браузера після виходу | Вийшов зі сторінки `/learn` | Натиснути «Назад» | Жодна сторінка з історії не показує імʼя чи вміст `/learn`; захищена адреса веде на `/login` | Playwright |
| G25 | Вкладені захищені адреси підпорядковані тій самій матриці | Cookie відсутній, потім сесія `student` | Відкрити `/learn/go/intro` і `/admin/courses` | Анонім → `/login?returnTo=` з відповідним шляхом; студент → `/learn/go/intro` відкривається, `/admin/courses` дає 404 | Playwright (переписаний `route-guards.spec.ts`); компонентні тести splat-роутів |
| G26 | Вхід через `/login`, коли вже увійшли | Сесія `admin` | Відкрити `/login?returnTo=%2Fadmin` | Показано, що користувач уже увійшов; контроль продовження веде на `/admin` | Компонентний тест; Playwright |
| G27 | Локалізація станів гарда | `uk`, потім `en` | Порожній стан, 404 як студент, шапка в обох станах | В обох мовах усе перекладено; жодного ключа-заглушки в DOM | Rstest з обома локалями |
| G28 | Три фікстури входять і мають правильні ролі (розряджає A01, A02, B01) | Виконано `make seed` | `POST /auth/login` для трьох ролей | Кожен — 200, `Set-Cookie`, DTO з очікуваною роллю; жодного поля з паролем чи токеном | Go контрактний тест; інтеграційний тест сіду |
| G29 | Ідемпотентність сіду (розряджає C02) | Порожня БД | `make seed` двічі | Обидва рази код 0; `count(*) FROM users` = 3 після кожного; рядки ідентичні | Інтеграційний тест на Testcontainers |
| G30 | Наявний конфліктний рядок не породжує часткових фікстур — **скориговано автором** | У БД користувач із email фікстури, але іншою роллю | `make seed` | Команда завершується кодом 0, наявний рядок не змінюється, `count(*)` не зростає; часткових фікстур не лишається | Інтеграційний тест сіду |
| G31 | Відмова сіду поза розробкою (рішення 6, розряджає C02) | `APP_ENV` не `development` (зокрема порожнє й `Development`) | `make seed` | Ненульовий код, повідомлення називає причину; `count(*) FROM users` = 0 | Тест у `infrastructure/tests/` з переозначеним `APP_ENV` |
| G32 | Облікові дані фікстур задокументовані як локальні (розряджає C02) | Чистий checkout | Прочитати README біля `make seed` | README містить три email-и й паролі та застереження про локальність; жодного пароля в `.env.example` чи Compose | Перевірка README; рев'ю |
| G33 | Сід не ламає інфраструктурні пороги | Compose піднято, сід виконано | `make smoke` | Пороги API/логіну/SSR лишаються; бурст із 30 спроб логіну все одно дає 429 | `make smoke` |
| G34 | Плейсхолдер Phase 1 більше недосяжний | Будь-який відвідувач | Пройти шість клітин матриці й вкладені адреси | «Розділ поки недоступний» і «Перейти до входу» не рендеряться ніде; переписаний spec не містить цього твердження | Playwright; grep по `LockedArea` |
| G35 | Оболонка `/admin` нічого не вигадує (межа слайсу) | Сесія `admin` | Відкрити `/admin` | Немає списків курсів, студентів, відео, прогресу чи коментарів; є явне повідомлення про пізнішу фазу | Компонентний тест; Playwright |
| G37 | Клавіатурна досяжність контролів шапки (конституція AGENTS.md) | Сесія `student` | Пройти табом від початку сторінки | Контроль «Вийти» отримує фокус із видимим кільцем і активується з клавіатури; порядок табуляції не стрибає через імʼя користувача | Rstest із `user-event`; Playwright `accessibility.spec.ts` |
| G38 | Шапка витримує 200% зуму й темну тему | Сесія `admin`, темна тема, зум 200% | Відкрити `/` | Навігація не накладається, імʼя не обрізає контроли, посилання адміністрування лишається доступним; жодного захардкодженого кольору | Playwright `accessibility.spec.ts` (той самий кейс, що для каталогу) |
| G39 | Розмір цілі контролів не менший за 44px | Будь-яка сесія | Виміряти «Вийти» й посилання адміністрування | Обидва не менші за 44px за меншим виміром | Playwright `accessibility.spec.ts` |
| G36 | Впорядкований гейт завершення фази (розряджає C03) | Чистий checkout, підняте Compose | Виконати всю послідовність гейтів | Кожен крок — код 0 у порядку; жоден не потребує ручного створення користувача; звіт фази посилається на цей прогін | Прогін гейта; журнал у `docs/identity/report.md` |

### Author's corrections to the delegated table

- **G19** demanded `Vary: Cookie` on the SSR response. Removed: `frontend/server.mjs:23`
  sets `Cache-Control: private, no-store`, and `docs/security/audit-2026-09-03.md` records
  that this is precisely why `Vary: Cookie` is not required. Asserting it would fail against
  a design that is already correct.
- **G30** demanded a non-zero exit when a conflicting row exists. Softened: the fixture
  insert is `ON CONFLICT DO NOTHING`, which cannot fail on a conflict. Detecting
  "same email, different role" would need a separate check for a case that arises only from
  hand-edited data, and the property that matters — no partial fixtures — holds as written.

### Reported gap discarded

The delegate reported `backend/internal/identity/adapters/http/` as untracked, which would
break a clean-checkout gate. It is tracked: `git ls-files` lists all four files, committed
in `e8cadab`. The delegate read the stale session-start status.

### Gap accepted, and it is a defect this slice must fix

`docs/identity/spec.md:52` requires `returnTo` to accept **only normalized internal
protected paths**, and to reject external URLs, protocol-relative and encoded forms,
control characters, **recursive `/login` targets** and **non-protected paths**.

`safeReturnTo` as shipped in I05 implements the first four and not the last two:
`safeReturnTo('/login')` returns `/login`, and `safeReturnTo('/courses')` returns
`/courses`. The phase security contract is stricter than the code that claims to satisfy
it. G15 encodes both missing cases, and narrowing `safeReturnTo` is a task in this slice.

The other two reported gaps are the slice's own work, not defects: `make seed` has no
`APP_ENV` guard yet, and the `check` target lists neither `check-generated` nor `smoke`,
so C03's ordered gate is not yet expressed as one command.
