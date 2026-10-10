# admin-stats — tasks

| | |
|---|---|
| Дата | 2026-10-10 |
| Спека/план | `spec.md`, `plan.md` (GATE 1 ✔, GATE 2 ✔ GREEN обидва рев'юери) |
| Реалізація | **Codex** (split impl/tests; `npm run format` per-slice; за потреби resume за id) |
| Гейт на слайс | ordered `make check` (format→lint→typecheck→test→integration→check-generated→build→up→seed→smoke→e2e) |
| Приймальні рядки | `plan.md §6.1` (1–14, 9a/9b/9c, 12b) |

Кожен таск: атомарний, явні файли, критерій «готово», посилання на приймальні рядки, [P]=паралельний.
Статус-трекер (planned→in_progress→done|blocked) — `work/tasks.md`, секція admin-stats.

## Backend

### T1 · SL1 · OpenAPI-контракт + re-gen `[залежить: —]`
- Файли: `backend/openapi/openapi.yaml` (+ 2 операції: `GET /admin/stats`, `GET /admin/stats/timeseries`;
  схеми Snapshot з `instructors[]`, `platform.singleInstructor`, `avgDurationSeconds` **nullable**),
  потім `make generate` → `frontend/src/adapters/api/generated/*`, sqlc (без змін).
- Готово: `make check-generated` зелений (нема дрейфу); обидві операції в згенерованому клієнті.
- Закриває (контракт): передумова для 1,3,9a/b/c,12b.

### T2 · SL2 · domain + ports + validation `[залежить: —]` `[P]`
- Файли: `backend/internal/adminstats/domain/*.go` (Snapshot/*, Metric, Bucket, TimeseriesPoint),
  `.../ports/stats_repository.go`, `.../domain/range.go` (cap `floor((to−from)/bucket)+1` в UTC;
  parse RFC3339 перед семантикою; коди `invalid_range`/`range_too_wide`/`unknown_metric`/`unknown_bucket`).
- Готово: unit-тести домену зелені — cap на межі 366 (рядок 7), from>to, parse-failure `from=banana` (12b),
  unknown metric/bucket (11), незалежний дефолт меж (N4).
- Закриває: 7, 11, 12, 12b (доменна частина).

### T3 · SL3 · snapshot repo (6 доменів) `[залежить: T2]`
- Файли: `backend/internal/adminstats/adapters/postgres/snapshot.go` + repo-тест на ізольованому Postgres.
- SQL: users (active/blocked/deleted/byRole/verified/twoFactor — **COUNT FILTER per key**, N1),
  courses (status), grants (state), orders (status + `revenueMinorUAH=SUM(amount_minor) FILTER paid`),
  reports (state), media (6-стан CASE→in-progress/ready/failed; `totalDurationSeconds`,
  `avgDurationSeconds` **nullable**, `withDurationCount`, `WHERE duration_seconds IS NOT NULL`).
- Готово: repo-тест — порожня БД дає нулі (не пропущені поля, рядок 5); null-duration (рядок 8);
  `active+blocked` узгоджено з identity-списком.
- Закриває: 1, 5, 8.

### T4 · SL4 · instructors leaderboard repo `[залежить: T2]`
- Файли: той самий пакет, `instructors.go` + repo-тест.
- SQL: `GROUP BY courses.instructor_id` з LEFT JOIN course_grants (active) і orders (paid); поля
  publishedCourses/activeEnrollments/revenueMinorUAH + displayName; топ-N (10) `revenue DESC, enroll DESC`.
- Готово: repo-тест — ≥2 інструктори впорядковані за revenue DESC (9b); інструктор без продажів = 0, не випав (9c).
- Закриває: 9b, 9c (дані).

### T5 · SL5 · timeseries repo (4 метрики) `[залежить: T2]`
- Файли: `timeseries.go` + repo-тест.
- SQL: `date_trunc(bucket, ts AT TIME ZONE 'UTC')`; signups=users.created_at, enrollments=grants.granted_at,
  orders=orders.created_at, revenue=orders.paid_at WHERE paid (SUM amount_minor).
- Готово: repo-тест — signups сума=к-сть у вікні (3); revenue за paid_at не created_at (4); порожнє вікно=`[]` (6).
- Закриває: 3, 4, 6 (дані).

### T6 · SL6 · http handler + wiring + flag `[залежить: T1,T2,T3,T4,T5]`
- Файли: `backend/internal/adminstats/adapters/http/handler.go` (+ `config.go` для
  `PLATFORM_SINGLE_INSTRUCTOR`, дефолт true), `application/service.go`, реєстрація в
  `backend/cmd/api/main.go` (`authenticate`→`RequireAdministrator`); `.env.example`/compose/CI env.
- Готово: contract-тести — 200 для admin; **403 `grant_forbidden`** student/instructor, 401 гість (10);
  payload містить `platform.singleInstructor`; помилки-коди (11/12/12b) на HTTP-рівні.
- Закриває: 1,3,9a,10,11,12,12b (HTTP).

### T7 · SL7 · engagement-slice (окремо, з виміром вартості) `[залежить: T3]`
- Файли: `engagement.go` + repo-тест; за потреби міграція `00016_*` (індекс) з up/down.
- Готово: активні учні (30д з `lesson_progress`), avg progress, completion — виміряно час запиту;
  якщо дорого/неточно — деградує до плейсхолдера (рядок 13), не вигадане число.
- Закриває: 13 (частково — залученість).

## Frontend

### T8 · SL8 · FE repo+schema+port+queries `[залежить: T1]`
- Файли: `adapters/api/admin-stats-schema.ts` (Zod, `avgDurationSeconds` nullable, `instructors[]`,
  `platform`), `adapters/api/admin-stats-repository.ts`, `ports/AdminStatsRepository.ts`,
  `app/composition/admin-stats-queries.ts`, додати в `app/composition/admin.ts`.
- Готово: typecheck/lint зелені; схема валідує реальний payload (fixture з контракту).

### T9 · SL9 · FE StatTile + TrendChart `[залежить: —]` `[P]`
- Файли: `presentation/ui/StatTile.tsx`, `presentation/ui/TrendChart.tsx` (inline-SVG line/bar;
  **без Donut**, N5), експорт у `presentation/ui/index.ts`; story/юніт.
- Готово: рендер line і bar; порожні дані → порожній стан; axe-чисто; клавіатура/фокус/44px/темна/200%.

### T10 · SL10 · `/admin/analytics` сторінка `[залежить: T8,T9]`
- Файли: `routes/admin.analytics.tsx`, nav-entry в `routes/__root.tsx`, icon у `presentation/admin/AdminNavIcon.tsx`
  (+ `AdminNavIconKey`), i18n `app/providers/locales/en.ts`+`uk.ts`; секції (вкл. leaderboard-таблицю),
  плейсхолдер-деградація, приховування виміру інструктора за `platform.singleInstructor`.
- Готово: екран рендерить реальні числа; картка без покриття → плейсхолдер (13); flag ховає інструктора (9a).
- Закриває: 2, 9a (UI).

### T11 · SL11 · e2e `[залежить: T6,T10]`
- Файли: `frontend/e2e/admin-analytics.spec.ts` (+ seed за потреби для ≥2 інструкторів).
- Готово: admin бачить числа й звірка `stats.active+blocked == users.list total` (14); не-admin 403 (10);
  тренд рендериться (2/3/6); leaderboard упорядкований за revenue (9b); single-instructor hide (9a); **axe 0**.
- Закриває: 2,3,6,9a,9b,10,13,14 (e2e) + A11/A12 гейти.

## Порядок виконання
1. **T1** (контракт) + **T2** (домен) — паралельно, перші.
2. **T3,T4,T5** — repo (після T2), можна паралельно між собою.
3. **T6** — handler (після T1–T5).
4. **T8** (після T1) + **T9** [P] — FE основи.
5. **T10** (після T8,T9).
6. **T7** — engagement (будь-коли після T3; окремий, ізольований ризик).
7. **T11** — e2e останнім (повний стек).

Коміти: малі логічні, per-task; кожен слайс → ordered `make check` → запис у `report.md`.
