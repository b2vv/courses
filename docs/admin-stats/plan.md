# admin-stats — plan (ЯК)

| | |
|---|---|
| Дата | 2026-10-10 |
| Фаза | admin-stats |
| Спека | `docs/admin-stats/spec.md` (GATE 1 пройдено 2026-10-10) |
| Поріг | ВИЩЕ (новий сабсистем + 2 public endpoints + крос-шар) |
| Гілка | `feat/admin-reconstruction` (на S0) |
| Код | реалізація — через **Codex** (рішення власника); plan/tasks — Courses |

Поглинає план-секцію майбутнього звіту. Side-effects / migration / rollback — §7–9 (обов'язкові для GATE 2).

## 1. Архітектура й розміщення

### 1.1. Backend — новий модуль `backend/internal/adminstats`
За `## Розміщення` маніфесту (`backend/internal/<module>/{domain,application,ports,adapters}`).

- **domain/** — типи-знімки (`Snapshot`, `UsersStat`, `CoursesStat`, `GrantsStat`, `OrdersStat`,
  `ReportsStat`, `MediaStat`) + `TimeseriesPoint`, `Metric` (enum), `Bucket` (enum). Stdlib + `contracts`
  тільки. Валідація діапазону/метрики/бакета — чисті функції домену.
- **ports/** — `StatsRepository` інтерфейс (`Snapshot(ctx) (domain.Snapshot, error)`,
  `Timeseries(ctx, Metric, Bucket, from, to) ([]TimeseriesPoint, error)`). Порт потрібен (є реальна
  зовнішня залежність — БД), не спекулятивний.
- **application/** — `Service` (admin-only політика через actor; виклик repo; застосування
  cap-перевірки діапазону з домену).
- **adapters/postgres/** — `Repository` над `*pgxpool.Pool` (shared, з `runtime.Dependencies.Database`).
  **Read-only** `COUNT(*) … GROUP BY` / `SUM` / `date_trunc` raw SQL по чужих таблицях
  (precedent: orders-repo джойнить users/courses/discussion — §precedent нижче). **Не sqlc** (sqlc.yaml
  генерує лише для course/identity/media; adminstats — hand pgx, як orders/discussion/grants).
- **adapters/http/** — `Handler` реєструє:
  - `GET /api/v1/admin/stats`
  - `GET /api/v1/admin/stats/timeseries`
  обидва обгорнуті `authenticate` → `RequireAdministrator` (строгий admin-only; НЕ course-`adminActor`).
  `adminHeaders`-стиль: `Cache-Control: no-store`. GET — без `requireSameOrigin`.
- **Wiring** `backend/cmd/api/main.go`: `adminstatsRepo := adminstatspostgres.NewRepository(deps.Database)`
  → `adminstatsApp := adminstatsapp.New(adminstatsRepo)` → `adminstatsHandler.Register(mux,
  identityHandler.Authenticate, identityHandler.RequireAdministrator)`.

### 1.2. Чому новий модуль, а не розширення існуючого (reuse-first обґрунтування)
Правило AGENTS.md: модуль не імпортує інший модуль; `contracts` без I/O. Агрегат читає 6 таблиць
5 модулів. Варіант «порт у кожному модулі + composition-root» (власник відхилив round 2) = 6 нових
портів + wiring. Обраний варіант (dedicated `adminstats` з raw-SQL repo) спирається на **наявний
precedent**: `internal/orders/adapters/postgres/repository.go` вже джойнить users/courses/discussion
таблиці raw SQL (Roster/Learners). adminstats робить те саме read-only — не вводить нового патерну.
Межа чиста: один дім для stats, один admin-gate, один round-trip.

### 1.3. Frontend — `/admin/analytics`
- **Route:** `frontend/src/routes/admin.analytics.tsx` (копія форми `admin.users.index.tsx`):
  `createFileRoute('/admin/analytics')`, `ssr:false`, `beforeLoad` role-check, `loader =
  ensureRouteQuery(context.queryClient, context.admin.stats.snapshot(), signal)`, component/pending/error.
- **Nav:** додати entry в admin `navSections`-літерал `routes/__root.tsx` (секція КУРСИ/ОГЛЯД), icon-key
  в `AdminNavIcon.tsx` (+ `AdminNavIconKey`), i18n-ключі `locales/en.ts` + `uk.ts`.
- **API-клієнт (Orval-first):** додати `/admin/stats` і `/admin/stats/timeseries` у
  `backend/openapi/openapi.yaml` → `make generate` (orval+sqlc) → `check-generated` зелений. Потім
  `adapters/api/admin-stats-schema.ts` (Zod), `adapters/api/admin-stats-repository.ts`
  (`createApiAdminStatsRepository` через `createApiTransport`), `ports/AdminStatsRepository`.
- **Query boundary:** новий `app/composition/admin-stats-queries.ts` (`createAdminStatsQueries` →
  `snapshot()`, `timeseries(range)` як `queryOptions`), додати в `createAdminDependencies()` (`composition/admin.ts`).
- **UI-примітиви (нові, hand-built):** `StatTile` (KPI-плитка; `.ui-stat` CSS вже є) + **один**
  параметризований inline-SVG `TrendChart` (режим line/bar під sparkline і бар; чарт-ліби нема й
  не додаємо — AGENTS.md). **N5 (GATE 2):** `Donut` ВИКИНУТО — жоден acceptance-рядок його не вимагає;
  розподіли показуємо наявними `LoadMeter`/бар-рядками. Нові примітиви — лише коли секція реально
  потребує. Токени семантичні, клавіатура/фокус/44px/світла-темна/200% — як решта admin.
- **Секції §4 плану аналітики:** Люди / Контент / Доступ+продажі / Залученість / Модерація / Операції.
  Картка без покриття → плейсхолдер (рішення власника).

## 2. Reuse-first (переюзуємо / нове / чому)

| Переюзуємо | Чому підходить |
|---|---|
| `identityHandler.RequireAdministrator` | єдиний строгий admin-only middleware (403 grant_forbidden) |
| `runtime.Dependencies.Database (*pgxpool.Pool)` | спільний пул, як у кожного модуля |
| raw-pgx-pattern orders/discussion-repo | precedent крос-таблиць read; adminstats не вводить нового |
| `httpserver.WriteJSON`, error-envelope, request-id | стандарт транспорту |
| FE `createApiTransport`, Orval-generated client, Zod-репо-патерн | як `admin-user-repository.ts` |
| FE `ensureRouteQuery`, Query-boundary, `createAdminDependencies` | як решта admin-queries |
| `ProgressBar`/`LoadMeter`/`Delta`, `.ui-stat` CSS | базис для tile/метрів |

| Нове | Чому існуюче не підійшло |
|---|---|
| модуль `adminstats` (domain/ports/app/adapters) | stats-агрегату нема ніде; жоден модуль не читає всі 6 |
| 2 OpenAPI-операції + Orval re-gen | нових ендпоінтів нема |
| FE `StatTile` + один inline-SVG `TrendChart` (line/bar) | чарт-компонентів і чарт-ліб нема зовсім (Donut викинуто, N5) |
| single-instructor прапорець | жодного feature-flag/platform-config нема |

## 3. Точні означення метрик (SQL-рівень, звірено зі схемою)

**N1 (zero-fill, GATE 2):** усі статус/роль/стан-лічильники рахувати
`COUNT(*) FILTER (WHERE … = '<key>')` **на кожен ключ** (не голий `GROUP BY`, що пропускає ключ із
0 рядків). Відсутній ключ = `0`, не брак поля. Інакше порожня платформа (сценарій 5) дає пропущені
поля замість нулів.

**Люди** (`users`; `deleted_at`/`disabled_at` керують статусом — identity.sql.go:66-69):
- active = `COUNT WHERE deleted_at IS NULL AND disabled_at IS NULL`
- blocked = `COUNT WHERE deleted_at IS NULL AND disabled_at IS NOT NULL`
- deleted = `COUNT WHERE deleted_at IS NOT NULL`
- byRole student/instructor/admin = `COUNT GROUP BY role WHERE deleted_at IS NULL` (role enum точний)
- verified = `COUNT WHERE email_verified AND deleted_at IS NULL` (00014)
- twoFactor = `COUNT WHERE totp_enabled AND deleted_at IS NULL` (00014)
- *«усього акаунтів» = active+blocked* (deleted виключено — дефолт списку, analytics-plan §5).

**Контент** (`courses`): draft/published/archived = `COUNT GROUP BY status`.

**Доступ+продажі:**
- grants (`course_grants`, post-00015): active/blocked/revoked = `COUNT GROUP BY state`.
- orders (`orders`): pending/paid/failed = `COUNT GROUP BY status`;
  `revenueMinorUAH = COALESCE(SUM(amount_minor),0) WHERE status='paid'`.

**Модерація** (`participant_reports`): open/resolved = `COUNT GROUP BY state`.

**Операції** (`video_assets`, реальний enum 6 станів):
- in-progress = `created|uploading|queued|processing`, ready = `ready`, failed = `failed`
  (`COUNT` з CASE-мапінгом 6→3).
- **Довжина (A6b):** `totalDurationSeconds = COALESCE(SUM(duration_seconds),0)`,
  `avgDurationSeconds = AVG(duration_seconds) WHERE duration_seconds IS NOT NULL`,
  `withDurationCount = COUNT WHERE duration_seconds IS NOT NULL` для чесного підпису покриття.
  **N2 (GATE 2):** `avgDurationSeconds` **nullable у контракті** — `AVG` по 0 рядках = NULL (не 0,
  бо 0 оманливий). OpenAPI/Zod `nullable`; FE при `withDurationCount=0` показує «—/нема покриття».

**Timeseries** (`date_trunc(bucket, ts)`):
| metric | таблиця.часова колонка | value |
|---|---|---|
| signups | `users.created_at` | COUNT |
| enrollments | `course_grants.granted_at` | COUNT |
| orders | `orders.created_at` | COUNT |
| revenue | `orders.paid_at` WHERE `status='paid'` | SUM(amount_minor) |

**Залученість — явні ПРИПУЩЕННЯ (потребують підтвердження GATE 2 / власника):**
- *активний учень* = distinct user з `active`-grant і активністю в `lesson_progress` за останні 30 днів
  (використати timestamp lesson_progress; звірити назву колонки на impl).
- *середній прогрес* = AVG прогресу по `active`-grants (progress рахує orders-repo Learners; для
  платформенного avg — окремий агрегат; дорожчий запит, окремий slice).
- *завершення* = COUNT (user,course) із повним проходженням (усі уроки/прогрес=100%).
  ⚠️ Ці три — найдорожчі; якщо на impl виявляться важкими/неточними — деградують до плейсхолдера
  (чесно), не до вигаданого числа. Зафіксовано як ризик §7.

**Інструктори — leaderboard (B1, join `courses.instructor_id`, індекс `courses_by_instructor`):**
Поле snapshot `instructors: [{instructorId, displayName, publishedCourses, activeEnrollments,
revenueMinorUAH}]`, топ-N (дефолт 10) ранг за `revenueMinorUAH DESC, activeEnrollments DESC`.
- `publishedCourses` = `COUNT(courses) WHERE instructor_id=… AND status='published'`.
- `activeEnrollments` = `COUNT(course_grants g JOIN courses c ON c.id=g.course_id) WHERE
  c.instructor_id=… AND g.state='active'`.
- `revenueMinorUAH` = `COALESCE(SUM(o.amount_minor),0) FROM orders o JOIN courses c ON c.id=o.course_id
  WHERE c.instructor_id=… AND o.status='paid'`.
- `displayName` з `users` (role instructor/admin). Один агрегатний запит `GROUP BY instructor_id`
  з LEFT JOIN-лічильниками (щоб інструктор без продажів давав 0, не випадав — пор. N1). Ховається
  прапорцем `singleInstructor` (§4).

## 4. single-instructor прапорець (механізм)

Власник: будувати зараз. **Reuse-first мінімум замість повного SSR-config-plumbing:**
- Backend: env-змінна `PLATFORM_SINGLE_INSTRUCTOR` (дефолт `true`), читається у новому
  `adminstats` config (`ConfigFromEnvironment`, патерн identity/playback config.go).
- Доставка клієнту: прапорець їде **в payload відповіді `/admin/stats`** як `platform:{singleInstructor}`.
  Споживач зараз лише `/admin/analytics`, яка цей ендпоінт і так тягне → **0 нового SSR-plumbing**.
  (Якщо згодом знадобиться глобально — промоутнути в SSR-context окремо; зараз YAGNI.)
- FE: analytics приховує вимір «інструктор» коли `singleInstructor=true`.

**B1 — контракт виміру «інструктор» (GATE 2, рішення власника 2026-10-10: будувати leaderboard зараз):**
Вимір інструктора у цій фазі = плитка-лічильник `users.byRole.instructor` + **per-instructor
leaderboard** (§3 «Інструктори»). Прапорець `singleInstructor` ховає обидва. Джерело join —
`courses.instructor_id` (NOT NULL FK users, індекс `courses_by_instructor`, 00003:6/21). Закриває
B1 реальним data-контрактом (не порожня секція).

## 5. Ліміт timeseries-вікна та валідація інпуту
- Дефолт: `from=now-30d`, `to=now`, `bucket=day`. **N4 (GATE 2):** кожна межа дефолтиться
  НЕЗАЛЕЖНО — лише `from` → `to=now`; лише `to` → `from=to-30d`. Одностороннього «invalid» нема.
- Cap: відмова `range_too_wide`, якщо к-сть бакетів > **366** (day ≤ ~1р, week ≤ ~7р).
- **N3 (GATE 2) — UTC скрізь:** і SQL `date_trunc` (`date_trunc('day', ts AT TIME ZONE 'UTC')`), і
  доменна cap-формула рахують бакети в UTC за однією формулою `floor((to−from)/bucket)+1`, щоб межа
  рівно на cap (сценарій 7) не плавала через DST (365/367).
- **Коди помилок (явні бізнес-коди, не 500, не мовчазний 0):**
  - `from>to` або бакетів > cap → `invalid_range` / `range_too_wide` (4xx).
  - невідома `metric` → `unknown_metric`; невідомий `bucket` → `unknown_bucket`.
  - **Q1 (GATE 2 defense) — непарсабельна дата:** `from`/`to`, що не парситься як RFC3339-дата,
    → `invalid_range` (4xx), з повідомленням про поле; парситься до валідації семантики. Чиста
    доменна/транспортна перевірка, тестована (unit + contract row).

## 6. Приймальні сценарії
Делеговано `acceptance-spec`. Три групи (happy / межі / відмови) — усі представлені.

### 6.1. Таблиця

| # | Сценарій | Передумова | Дія | Очікуваний спостережуваний результат | Чим перевіряється |
|---|---|---|---|---|---|
| **Happy path** |
| 1 | Знімок як admin | seed із даними; actor=admin | `GET /api/v1/admin/stats` | 200 JSON: `users/courses/grants/orders/reports/media` присутні; `orders.revenueMinorUAH` ≥ 0 — число | contract `adminstats_handler_test.go` |
| 2 | Екран аналітики | admin залогінений, seed | відкрити `/admin/analytics` | видно hero-плитки з числами (усього акаунтів, курсів, дохід, відкриті скарги) і секції Люди/Контент/Продажі/Модерація/Операції з реальними числами | e2e `admin-analytics.spec.ts` |
| 3 | Тренд signups | seed з реєстраціями в діапазоні | `GET …/timeseries?metric=signups&bucket=day&from&to` | 200 масив `[{bucketStart,value}]`, сума value = к-сть users.created_at у вікні | contract + repo `adminstats_timeseries_test.go` |
| 4 | Revenue за paid_at | 1 paid-order із `paid_at` у вікні, `created_at` поза ним | `…?metric=revenue&bucket=day…` | бакет потрапляє за `paid_at`, не за `created_at`; value = amount_minor | repo test |
| **Межі** |
| 5 | Порожня платформа | БД без курсів/замовлень/скарг | `GET /admin/stats` + екран | усі лічильники = 0, revenue = 0; екран рендериться без помилки, картки показують 0 / порожній стан | contract + e2e |
| 6 | Порожнє вікно тренду | нема рядків у діапазоні | timeseries у «тихому» вікні | `[]` (порожній масив); чарт показує порожній стан, не падає | contract + e2e |
| 7 | Максимальне вікно | діапазон рівно на cap (366 day-бакетів) | timeseries на межі | 200, повертає бакети (не `range_too_wide`) | unit (доменна cap-функція) + contract |
| 8 | Відео без тривалості | частина `video_assets.duration_seconds IS NULL` (до публікації) | `GET /admin/stats` | `avgDurationSeconds` рахується лише по non-null; `withDurationCount` < усіх; підпис покриття видно на екрані | repo + e2e |
| 9a | single-instructor ON | env `PLATFORM_SINGLE_INSTRUCTOR=true` | відкрити `/admin/analytics` | плитка-лічильник «Інструктори» і секція-leaderboard ВІДСУТНІ; payload `platform.singleInstructor=true` | e2e + contract |
| 9b | single-instructor OFF | env `=false`; seed з ≥2 інструкторами, їх курсами, paid-замовленнями | відкрити `/admin/analytics` | секція-leaderboard ПРИСУТНЯ; рядки впорядковані за revenue DESC; числа курсів/enrollments/доходу = реальні агрегати | e2e + contract |
| 9c | інструктор без продажів | OFF; інструктор із 0 paid-замовлень | завантажити `/admin/stats` | цей інструктор у leaderboard з `revenueMinorUAH=0`/`activeEnrollments=0` (LEFT JOIN, не випадає) | contract |
| **Відмови** |
| 10 | Не-admin відмова | actor=student або instructor або гість | `GET /admin/stats` / `…/timeseries` | 403 код `grant_forbidden` (гість → 401 `authentication_required`); тіло без даних | contract |
| 11 | Невідома метрика | admin | `…/timeseries?metric=bogus` | явна бізнес-помилка (4xx, код `unknown_metric`), не 500, не порожній «0» | contract + unit |
| 12 | Кривий діапазон (семантика) | admin | `from>to` або вікно > cap | бізнес-код `invalid_range` / `range_too_wide` (4xx), не 500 | contract + unit |
| 12b | Непарсабельна дата (Q1) | admin | `…?from=banana` (не RFC3339) | бізнес-код `invalid_range` (4xx) з полем; не 500, не мовчазний 0 | contract + unit |
| 13 | Чесна деградація | метрика залученості не покрита цим слайсом | відкрити секцію «Залученість» | картка показує плейсхолдер «потребує stats-ендпоінта»; жодного вигаданого числа в DOM | e2e |
| 14 | Узгодженість із списком (A3) | seed; admin | завантажити `/admin/analytics` і `/admin/users` | `users.active + users.blocked` зі stats = `pagination.total` з `users.list` без фільтра | e2e (обидва запити в одному тесті) |

**Ручних рядків: 0** — усі покриті contract/repo/unit/e2e. (A11 axe-перевірка й A12 ordered `make check` —
гейти, не окремі рядки: axe вбудовано в e2e `admin-analytics.spec.ts`.)

## 7. Сайд-ефекти (обов'язково для GATE 2)
- **Чисте читання:** adminstats лише SELECT; жодних записів, тригерів, стану. Нові індекси —
  тільки якщо профіль покаже потребу (існуючі статусні індекси вже є: courses_published_recent,
  discussion_reviews_counted тощо); спершу без нових індексів, виміряти.
- **Навантаження:** знімок = ~6 легких COUNT/GROUP BY на індексованих колонках, один round-trip;
  no-store (без кешу цієї фази). **Leaderboard** = 1 GROUP BY по `instructor_id` з LEFT JOIN
  grants/orders (індекс `courses_by_instructor`); топ-N обмежує вивід — виміряти на seed.
  Engagement-avg — найдорожчий запит; ізолювати в окремий slice, виміряти, за потреби обмежити/деградувати.
- **Shared/persisted стан:** нові колонки/таблиці НЕ додаються → ризик зламати чужий лічильник
  низький, але **повний `make check` з усім e2e** (не таргетний) — зміни зачіпають shared schema-read.
- **OpenAPI/generated:** re-gen змінює `generated/courses.ts` — закомітити разом, інакше check-generated.
- **i18n:** нові ключі у двох локалях (інакше впаде рендер/лінт).

## 8. План міграції
- **Схема:** міграцій НЕ потрібно (усі колонки існують; звірено §3). Якщо профіль engagement
  вимагатиме індексу — окрема `00016_*`-міграція з up/down, проганяється на populated seed
  (пам'ять: seed.sql має вантажитись на downgraded-схему — нові INSERT guard'ити to_regclass;
  тут INSERT нема, лише, можливо, CREATE INDEX).
- **Дані:** backfill не потрібен (читання наявного).
- **Env:** додати `PLATFORM_SINGLE_INSTRUCTOR` у `.env.example` / compose / CI з дефолтом true.

## 9. Rollback
- Backend: новий модуль + роут ізольовані; відкат = прибрати реєстрацію в main.go (решта не зачеплена).
- Frontend: новий роут/нав/компоненти; відкат = прибрати nav-entry + route (інші екрани не залежать).
- OpenAPI: revert spec + re-gen.
- Жодних незворотних змін (нема міграцій/записів) → відкат чистий `git revert` слайсу.

## 10. Конституція (маніфест §Конституція)
- AGENTS.md шари: domain(stdlib+contracts) ← application(domain+ports) ← adapters(pgx/Orval у адаптері,
  мапа в домен). ✔ дотримано розкладкою §1.
- Модуль не імпортує модуль; cross-module — raw SQL у власному adapter (precedent). ✔
- Нема глобального мутабельного клієнта; FE Query/SSR request-scoped. ✔
- TS strict, без any; зовнішні дані (stats payload) валідуються Zod. ✔
- Nginx/медіа-байти не зачіпаються. ✔
- Порушень не виявлено.

## 11. Слайси (прев'ю; деталі в tasks.md). Мітки SL# — щоб не плутати з acceptance A#/знахідкою B1.
1. **SL1** OpenAPI-контракт обох ендпоінтів (+ `instructors` leaderboard, `platform.singleInstructor`, nullable avg) + re-gen (контракт-first, FE/BE паралеляться).
2. **SL2** backend domain+ports+validation (cap/metric/bucket/parse-failure, UTC-формула) — юніт-тести.
3. **SL3** backend adapters/postgres snapshot (users/courses/grants/orders/reports/media+duration, COUNT FILTER zero-fill) — repo-тести на ізольованому Postgres.
4. **SL4** backend adapters/postgres **instructors leaderboard** (join `courses.instructor_id`, LEFT JOIN лічильники, топ-N за revenue) — repo-тести (вкл. інструктор без продажів = 0).
5. **SL5** backend adapters/postgres timeseries (4 метрики, revenue за `paid_at`) — repo-тести.
6. **SL6** backend http handler + wiring + RequireAdministrator + single-instructor env у payload — contract-тести.
7. **SL7** engagement-slice (активні/avg progress/completion) — окремо, з виміром вартості; деградація за потреби.
8. **SL8** FE repo+schema+port+queries (snapshot+timeseries), Zod nullable avg.
9. **SL9** FE `StatTile` + один `TrendChart` (line/bar; Donut викинуто, N5) — юніт/стор.
10. **SL10** FE `/admin/analytics` route+nav+icon+i18n, секції (вкл. leaderboard-таблицю), плейсхолдер-деградація, single-instructor hide.
11. **SL11** e2e: admin бачить реальні числа (звірка з наявними списками, сценарій 14), не-admin 403, тренд рендериться, leaderboard упорядкований, axe 0.

Кожен слайс: реалізація (**Codex**, split impl/tests, `npm run format` per-slice) → ordered `make check` → запис у report.
