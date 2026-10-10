# admin-stats — жива нота (spec-flow)

РОБОЧИЙ ФАЙЛ цієї фази. Дописую ПІСЛЯ КОЖНОГО КРОКУ (append). Деталі — у репо; тут вказівники.

## План проходу (spec-flow)
0. [x] Маніфест прочитано. Поріг: ВИЩЕ (новий сабсистем + public API ×2 + крос-шар). spec-flow увімкнено.
1. [x] Setup: дім `docs/admin-stats/`, гілка — на S0 (feat/admin-reconstruction), жива нота — цей файл.
2. [x] spec.md (ЩО/ЧОМУ) — 12 acceptance, 6 out-of-scope, 5 відкритих.
3. [x] GATE 1 — grilling (2 раунди рішень) пройдено, власник підтвердив перехід.
4. [x] plan.md (ЯК) — архітектура, reuse, точний SQL per-метрика, side-effects/migration/rollback,
      конституція §10, acceptance-таблиця §6.1 (14 рядків, 0 ручних).
5. [~] GATE 2 — обидва RED, усі знахідки легітимні (critique звірив усі code-claims із репо — усі TRUE).
      critique: B1 (blocking — вимір «інструктор» без data-контракту) + N1..N5. defense: Q1 (parse-failure).
      ВСІ ФІКСИ ЗАСТОСОВАНО в plan.md:
      - B1 → вимір інструктора = плитка-лічильник byRole.instructor (hideable); leaderboard відкладено
        у analytics-tags-popularity-фазу. (потребує owner-confirm, бо вужче за §9.2).
      - Q1 → непарсабельна дата → invalid_range (§5 + acceptance 12b).
      - N1 COUNT FILTER zero-fill; N2 avgDuration nullable; N3 UTC date_trunc+cap формула;
        N4 кожна межа дефолтиться незалежно; N5 Donut викинуто, один TrendChart (line/bar).
      owner-confirm B1: **будувати leaderboard зараз** → plan оновлено (§3 Інструктори SQL join
      courses.instructor_id, §4, SL4, acceptance 9a/9b/9c). Re-verify надіслано обом рев'юерам
      (SendMessage), чекаю нові вердикти. Після GREEN → tasks.md.
   **GATE 2 GREEN обидва** (defense + critique re-verify; critique звірив новий leaderboard-SQL із
   міграціями 00003/00011). Дрібний косметичний борг: §3 рядки courses/grants/reports/media ще
   кажуть «COUNT GROUP BY» — керує N1-правило зверху, Codex робить FILTER. Далі: tasks.md.
6. [x] tasks.md (T1–T11) + work/tasks.md секція admin-stats. Планувальні доки закомічено `9cc7f72`
      (не запушено). GATE 1✔ GATE 2✔. Готово до implement.
7. [!] implement — **ЗАБЛОКОВАНО**: Codex usage limit вичерпано (reset 14.10 07:54). codex-rescue
      сабагент падав на Sonnet-safeguard (reasoning_extraction); codex exec напряму — «hit usage limit».
      Жодних code-змін не зроблено. Чекаю рішення власника: ждати до 14.10 / поповнити кредити /
      Claude реалізує цю сесію (відступ від «код через Codex»).
      РІШЕННЯ власника 2026-10-10: **Claude реалізує цю сесію** (Codex недоступний до 14.10).
      **T2 + T1 DONE `1cfb967`**: domain/ports/validation (unit+vet+gofmt clean) + OpenAPI 2 ендпоінти
      + Orval regen (check-generated green, build+typecheck ok). Планувальні доки — `9cc7f72`.
      ЗАЛИШИЛОСЬ: T3 snapshot repo · T4 leaderboard repo · T5 timeseries repo (всі потребують
      ізольованого Postgres + repo-тести) · T6 handler+wiring+flag · T7 engagement · T8 FE repo/queries
      · T9 StatTile+TrendChart · T10 /admin/analytics сторінка · T11 e2e+axe.
      Повний ordered `make check` — після T6 (handler) і фінально на T11.
8. [ ] verify (code-review на діф).
9. [ ] ship-report (дельта до плану).

## Факти (число + метод)
- 2026-10-10: admin-ендпоінти РОЗКИДАНІ по модулях, окремого admin-модуля НЕМА.
  `/api/v1/admin/courses` → `backend/internal/course/adapters/http/admin_handler.go:128`.
  Наслідок: крос-модульний stats-агрегатор = ключова архітектурна розвилка (модуль не імпортує
  модуль; contracts без I/O). Вирішити на кроці 4 (plan).

## Рішення
- 2026-10-10 власник: аналітика-мінфаза через spec-flow; revenue включити зараз; знімок+тренди (B+C).
  Зафіксовано в `docs/design/admin-analytics-plan.md §8.1`.

## Факти (крок 3, від 2 Explore-агентів, 2026-10-10)
Backend:
- created_at: users✓, orders✓(+paid_at), courses✓(+updated/published_at), participant_reports✓,
  video_assets✓(no default, app-set). course_grants — НЕ created_at, а `granted_at`✓
  (00011_payment.sql:35). Усі домени бакетуються. (grants бакет = granted_at.)
- **Admin-авторизація:** строгий admin-only гейт = `identityHandler.RequireAdministrator`
  (403 "grant_forbidden" non-admin; identity handler.go:518). course `adminActor` НЕ admin-only
  (instructor проходить) — НЕ використовувати для stats. Реєстрація: wrap authenticate →
  requireSameOrigin(мутації) → RequireAdministrator (precedent: orders/discussion RegisterAdmin, main.go:158/165).
- **Модулі-джерела:** НЕМА enrollment-модуля. users→identity(sqlc). course_grants/roster/learners→
  **orders**-модуль (hand pgx, має GROUP BY/count). orders→orders (лише lookup, агрегату НЕМА — net-new SQL).
  courses→course(sqlc, має count). participant_reports→**discussion** (cursor, без count). media→media(sqlc, state+created_at).
- **Orders enum РЕАЛЬНИЙ: 'pending','paid','failed'** (00011:19; domain order.go:8) — НЕ
  approved/declined/refunded як припускав spec draft. Колонки: amount_minor bigint, currency 'UAH',
  status, created_at, paid_at. revenue = Σ amount_minor WHERE status='paid'. → ВИПРАВИТИ spec A1.
- **Крос-модуль:** жоден хендлер не тримає кілька repo. Precedent: (i) repo одного модуля читає
  чужі таблиці raw SQL (orders repo join users/courses/discussion); (ii) bridge-структ у main.go.
  Новий агрегатор: dependencies.Database(*pgxpool.Pool) → NewRepository → register під RequireAdministrator.
- **Config-flag:** НЕМА жодного feature-flag/central Config. singleInstructor = повністю нове
  (env у ConfigFromEnvironment модуля + SSR→клієнт).

Frontend:
- AdminShell рендериться в `__root.tsx:143` (не в routes), navSections — літерал-масив там (admin гілка :157).
  Додати: route `admin.analytics.tsx` (копія admin.users.index.tsx), nav-entry в __root, icon в
  AdminNavIcon.tsx, i18n en.ts/uk.ts. Route tree генерується авто.
- **ui-stat КОМПОНЕНТА НЕМА** (лише CSS-класи .ui-stat у Storybook-сторі; CSS існує). Примітиви:
  ProgressBar, LoadMeter, Delta (data.tsx). KPI-tile — новий.
- **Чартів НЕМА, чарт-ліби НЕМА.** Hand-built inline SVG (sparkline/donut/bar — свої).
- **API: Orval** з `backend/openapi/openapi.yaml` → `make generate` → **check-generated гейт** (CI падає
  на дрейфі). Потім hand repo-adapter + Zod-схема + port. Нові: admin-stats-schema.ts,
  admin-stats-repository.ts, ports/AdminStatsRepository, composition/admin-stats-queries.ts, в admin.ts.
- single-instructor на фронті — повністю нове (SSR config до клієнта немає).

## Рішення (крок 3 GATE 1, round 1, власник 2026-10-10)
- Гілка: продовжити `feat/admin-reconstruction` (UI на S0).
- Обсяг: усі 6 доменів (повний розділ).
- Деградація: плейсхолдер-картка «потребує stats-ендпоінта».

## Рішення (крок 3 GATE 1, round 2, власник 2026-10-10)
- Orders: {pending,paid,failed}; revenue=SUM(amount_minor WHERE paid); revenue-timeseries за paid_at.
- single-instructor: БУДУВАТИ прапорець зараз (env→SSR→клієнт, дефолт true).
- Агрегатор: новий dedicated `backend/internal/adminstats`-модуль, raw SQL крос-таблиці, під RequireAdministrator.

## Відкрите (plan-рівень, не блокує GATE 1)
- Ліміт timeseries-вікна (cap бакетів) — у plan.
- Engagement-означення (активний учень / avg progress / completion) — явні припущення в plan.
- Джерело single-instructor-конфігу (env-змінна/дефолт/SSR-канал) — у plan.

## Факти — теги/популярність/довжина (запит власника під час GATE 1)
- Теги курсу/відео: НЕ існують (grep) → нова продуктова фіча, не аналітика.
- `video_assets.duration_seconds` ІСНУЄ (00006, nullable) → довжина відео має джерело.
- Популярність: сигнали є всі (grants, orders paid, discussion_reviews.stars, lesson_progress).
- Оформлено окремо: `docs/design/admin-analytics-tags-popularity.md` (аналіз + фазування).
  ВІДКЛАДЕНО за рішенням власника — НЕ в admin-stats. spec out-of-scope оновлено.

## Журнал
- крок 0–1: маніфест, поріг, дім, жива нота.
- крок 2: spec.md.
- крок 3 GATE 1: 2 Explore-агенти (факти); round 1 + round 2 власника; spec виправлено
  (orders enum, RequireAdministrator, single-instructor у скоуп, питання вирішено).
- GATE 1 (триває): власник попросив аналіз популярності/тегів/довжини → написано deferred-док,
  межу зафіксовано в spec out-of-scope.
- GATE 1 ЗАКРИТО (власник): базова довжина відео (total/avg) → у домен «Операції» поточної фази
  (A6b); watch-through/гістограма відкладені. Перехід до plan.md підтверджено.
- Рішення власника: **код — через Codex** (крок implement; split impl/tests, npm run format per-slice).
  Я пишу plan.md/tasks; реалізацію делегую Codex. Далі: plan.md.
- T1/T2 реалізовано (`1cfb967`), трекери `9193e34`. Codex недоступний до 14.10 → код пише Claude.
- **T3 (snapshot repo) готово** (Claude): `adapters/postgres/snapshot.go` (6 доменів, COUNT FILTER
  per-key zero-fill N1; media 6→3 CASE; revenue `SUM(amount_minor) FILTER paid ::bigint`;
  duration total `COALESCE(SUM,0)`, avg `AVG FILTER (duration IS NOT NULL)`→nullable, withDurationCount)
  + `snapshot_integration_test.go`. Instructors лишено порожнім (T4 допише Snapshot()).
  Додав `domain.ErrDependencyUnavailable` (read-path фейл). Верифікація: gofmt/build/`go vet ./...`
  чисто; repo-тест 2/2 на ізольованому Postgres — порожня БД=нулі+avg nil (рядок 5/8-null),
  seed рахує всі 6 доменів, `active+blocked == COUNT non-deleted` (A3, рядок 14-прекурсор).
  Повний ordered `make check` — за планом ПІСЛЯ T6. Далі: **T4** (instructors leaderboard).
- **T4 (instructors leaderboard) готово** (Claude): `adapters/postgres/instructors.go` —
  `loadInstructors` будує leaderboard: base `SELECT DISTINCT instructor_id FROM courses`,
  кожна метрика (publishedCourses/activeEnrollments/revenue) — **окремий pre-aggregated subquery
  LEFT JOIN на instructor_id** (щоб 3 незалежні one-to-many не давали fan-out), zero-fill COALESCE,
  `ORDER BY revenue DESC, enroll DESC, id`, LIMIT 10. `Snapshot()` тепер кличе loadInstructors.
  Тест `instructors_integration_test.go`: 3 інструктори, A(10000)>B(3000)>C(0); інструктор без
  продажів присутній із 0 (9c), не випадає; читання через `Snapshot().Instructors` (перевіряє wiring).
  Контейнер-бойлерплейт винесено у спільний `newTestPool(t,ctx)`. Верифікація: gofmt/build/vet
  чисто; обидва repo-тести зелені (leaderboard + snapshot). `4d3a141`→наступний коміт. Далі: **T5** (timeseries).
- **T5 (timeseries repo) готово** (Claude): `adapters/postgres/timeseries.go` — `Timeseries(...)` реалізує порт.
  `date_trunc($1, <timecol> AT TIME ZONE 'UTC')` (N3 UTC-бакети), half-open `[from,to)`,
  `value::double precision`. Per-metric: signups=users.created_at, enrollments=grants.granted_at,
  orders=orders.created_at, revenue=orders.paid_at `WHERE status='paid'` (SUM amount_minor). Ідентифікатори
  фіксовані (не інпут), параметри лише bucket-unit+вікно. Порожнє вікно → **non-nil `[]`** (не null).
  `Repository` тепер повністю реалізує `ports.StatsRepository` — додано `var _ ports.StatsRepository`.
  Тест `timeseries_integration_test.go` (3 сабтести): signups сума==`COUNT` у вікні [3] (звірка з таблицею,
  robust до revenue-фікстур); revenue 1 бакет за paid_at=7777, той самий ордер у orders-метриці (за created_at)
  поза вікном=0 [4]; порожнє вікно non-nil len0 [6]. Верифікація: gofmt/build/vet чисто; усі 3 adminstats
  repo-тести зелені. Далі: **T6** (handler+wiring+flag → **повний ordered make check ПІСЛЯ T6**).
- **T6 (handler+service+config+wiring+env) готово** (Claude): `adapters/http/handler.go` (2 GET-ендпоінти
  за `authenticate→RequireAdministrator`, no-store, DTO точно за openapi, instructors→non-nil `[]`,
  avg→`*float64` null), `adapters/http/config.go` (`PLATFORM_SINGLE_INSTRUCTOR` дефолт true,
  empty→fallback), `application/service.go` (Snapshot стемпить Platform.SingleInstructor; Timeseries
  ParseMetric→ParseBucket→ResolveRange перед repo). Wiring `cmd/api/main.go` (config+repo+service+
  RegisterAdmin). Env у `.env.example`+`docker-compose.yml`.
  - **Знахідка/дефект:** T1 додав операції+схеми, але **НЕ** додав коди `invalid_range/range_too_wide/
    unknown_metric/unknown_bucket` у `ErrorResponse.code` enum → contract-валідатор відхиляв би тіло.
    Додав 4 коди в enum + `make generate` → `errorResponseErrorCode.ts` +4 (check-generated 0 drift).
  - **РІШЕННЯ (2026-10-10, Claude):** адмін-гейт живе ЛИШЕ в middleware `RequireAdministrator`
    (єдина точка; прод-шлях і contract-тест через нього). Service НЕ приймає actor і не дублює
    перевірку — прибирає недосяжну гілку; спостережувана поведінка (401/403) незмінна й покрита
    contract-тестом. Закриває альтернативу plan §1 «service re-checks admin via actor».
  - contract-тест `handler_test.go`: 200 admin+no-store+singleInstructor+revenue; порожній leaderboard→`[]`,
    nil avg→null; авторизація guest401/student403/instructor403/admin200 [10]; timeseries 200+parsed
    metric/bucket; порожня серія→`[]`; range-помилки invalid_range/range_too_wide/unknown_metric/
    unknown_bucket [11/12/12b] (unknown-* out-of-contract request → assertContract skip, бо прод не
    openapi-валідує запити; kin-openapi таки чекає date-time формат).
  - **ВЕРИФІКАЦІЯ: повний ordered `make check` зелений** — format/lint/typecheck, test (FE 653, BE усе),
    integration EXIT=0 (adminstats repo 8.1s), check-generated 0 drift, build усі образи,
    **e2e 112 passed (12.7m, 0 failed/0 did-not-run)**. Бекенд фази завершено. Далі: **T7** (engagement) / FE T8–T11.
