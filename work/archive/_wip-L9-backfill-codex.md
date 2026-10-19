ДАЛІ:

## Крок 1 — бриф прочитано
Прочитано CTO-RESEARCH, brief-codex-L9-backfill-test і codex-working-model. Межі: один subtest у repository_integration_test.go; лише цей тест і append-only WIP. Git, Docker, міграції, generated, tasks та phase report не змінюю. Далі — звірка реальних схем і harness.

## Крок 2 — схеми й harness звірено
Прочитано enrollment/plan.md і work/tasks.md (лише читання), весь цільовий тест, міграції 00001–00004, 00006, 00009–00015 за потрібними таблицями. DownTo(14) → pool.Exec → UpTo(15) → assertions → Up відповідає playback-subtest. Users/modules/lessons/reviews/courses звірені також із наявними INSERT у тесті; progress/comments INSERT у цьому файлі немає, їхні шейпи звірені з CREATE TABLE 00006/00009. Paid: price_minor + currency; published: published_at між created_at/updated_at; progress: position_seconds/client_seq/updated_at; review: stars/body/completed_lessons/total_lessons/created_at. Додам окремого fixture-owner і cleanup grants → courses (cascade activity) → users до наступного subtest.

## Discussion
2026-10-07, Codex: бриф називає дві progress-активності доказом ON CONFLICT, але 00015_enrollment.sql:37–58 UNION уже усуває дубль (user,course) до INSERT; видалення ON CONFLICT не змусить такий тест упасти. Це доказ дедуплікації результату, не конфлікту з наявним grant. Реального дефекту backfill SQL не виявлено. Реалізую дозволений брифом варіант A з двома уроками, без вставок grants до міграції. Для окремого доказу conflict-skip потрібен старий допустимий admin/grandfather grant перед UpTo(15) та assertion збереження рядка; залишаю це на рішення Courses 1.

## Крок 3 — додано один subtest
A/B/C — незалежні UNION-гілки, A має два progress-рядки на одному курсі; D paid та E draft — count=0; F free archived — self/active. Усі перевірки scoped на конкретну пару user/course, позитиви рахують усі рядки й self/active окремо. Нові UUID, окремий instructor, cleanup власних grants/courses/users; seeded-рядки не змінюються. Далі — gofmt, go vet і compile-only integration-tagged test.

## Крок 4 — формат і scope перевірено
gofmt виконано; gofmt -l порожній. Порівняння з копією файлу до правки (без git): рівно один доданий t.Run, видалення цього блоку повертає байт-у-байт початковий файл. go vet і compile-only test запущені з GOCACHE=/tmp/courses-L9-backfill-go-cache, результати ще очікуються.

## Крок 5 — верифікація завершена
- PASS: `cd backend && GOCACHE=/tmp/courses-L9-backfill-go-cache go vet ./internal/course/...` — exit 0, без діагностик.
- PASS (лише компіляція): `cd backend && GOCACHE=/tmp/courses-L9-backfill-go-cache go test -tags=integration -run '^$' ./internal/course/adapters/postgres` — exit 0; `ok courses/internal/course/adapters/postgres 1.347s [no tests to run]`.
- Integration **NOT RUN**: Docker / `make integration` заборонено умовами цього завдання; SQL і міграція на PostgreSQL тут не виконувалися. Реальний прогін — Courses 1.
- Зміни: лише один subtest у `backend/internal/course/adapters/postgres/repository_integration_test.go` і цей append-only WIP. Git не викликався; tasks, docs/report, generated, migrations і backend-logic не змінювались.
- ДАЛІ: Courses 1 запускає реальний `make integration` та читає Discussion про межу доказу UNION-дедуплікації.
