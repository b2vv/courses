# Brief — B-L9.1: integration test for the 00015 enrollment backfill (En13)

РОБОЧИЙ ФАЙЛ: `work/_wip-L9-backfill-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`.
Docker/integration ти НЕ ганяєш — пишеш тест, Courses 1 проганяє `make integration`.

## Контекст / чому (Courses 2 BLOCKING)
En13 (backfill) позначено «proven at integration», але НІЩО не проганяє логіку backfill'у міграції
`00015`. Orders-integration ганяє міграції на ПОРОЖНІЙ БД → `INSERT…SELECT…UNION` backfill нічого не
знаходить (no-op); seed self-active — це seed INSERT, не backfill. Backfill — one-shot prod-міграція:
якщо UNION/фільтр хибний, наявні free-користувачі (progress/comment/review, без членства) НЕ отримають
членство → втратять доступ до free-контенту. Треба справжній тест на populated-даних.

Backfill SQL (UP, `backend/migrations/00015_enrollment.sql`): для кожного (user,course), де user має
рядок у `lesson_progress` АБО `discussion_comments` АБО `discussion_reviews` на курсі, і курс
`price_minor IS NULL AND status IN ('published','archived')` → INSERT `course_grants` source='self'
state='active', `ON CONFLICT (user_id,course_id) DO NOTHING`.

## Завдання — один subtest у наявному harness
Файл: `backend/internal/course/adapters/postgres/repository_integration_test.go`. Дзеркаль субтест
**«playback migration preserves populated database»** (~рядок 276): `provider.DownTo(ctx, 14)` → `pool.Exec`
inserts → `provider.UpTo(ctx, 15)` → assert → `provider.Up(ctx)` (відновити). Використай наявні `pool`,
`db`, `ctx`, `provider`. Назви субтест «enrollment migration backfills free-course activity into self-active membership».

На DownTo(14) (до enrollment-міграції; `course_grants` ще має старий CHECK без 'self'/'blocked' — тому
вставляй АКТИВНІСТЬ, НЕ grants). Встав фікстури з унікальними UUID (не чіпай сідовані), ТОЧНІ шейпи звір
за реальними схемами міграцій/наявними вставками в цьому файлі:
- FREE PUBLISHED курс (`price_minor IS NULL`, `status='published'`) + module + lesson.
- Окремі membership-less users, по одному на кожну гілку UNION, щоб довести КОЖНУ:
  - user A: рядок у `lesson_progress` на уроці курсу;
  - user B: рядок у `discussion_comments` на уроці курсу;
  - user C: рядок у `discussion_reviews` на курсі.
- Негативи (НЕ мають бекфілитись):
  - user D: активність на **PAID** курсі (`price_minor` НЕ NULL, published) → немає grant;
  - user E: активність на **DRAFT** курсі (`status='draft'`) → немає grant;
- Позитив-архів (МАЄ, En09): user F: активність на FREE **ARCHIVED** курсі → отримує grant.
- ON CONFLICT: user G, що вже має `course_grants` рядок на free-курсі (встав його ПІСЛЯ UpTo? ні —
  course_grants старий CHECK до 15). Простіше: user G має активність І на UpTo(15) backfill спробує
  вставити, але ти спершу (після UpTo, коли CHECK вже широкий) вставив би дубль… — АБО доведи ON CONFLICT
  інакше: user A має АКТИВНІСТЬ на ДВОХ уроках того самого курсу (дві progress-рядки) → UNION+ON CONFLICT
  дає РІВНО один grant (не два). Це чисто доводить дедуп без маніпуляцій зі старим CHECK.

Після `UpTo(ctx, 15)` assert через `db.QueryRowContext`/`pool`:
- A, B, C мають рівно один `course_grants` рядок, `source='self'`, `state='active'` (кожен).
- D, E — НЕ мають рядка (`count=0`).
- F — має self/active (архів).
- A — рівно ОДИН рядок попри дві активності (ON CONFLICT dedup).
Наприкінці `provider.Up(ctx)` (відновити повну схему, як інші субтести).

## Межі
Лише цей тест-файл (+ нічого в продакшені). НЕ чіпати міграцію/бекенд-логіку. НЕ комітити. git/tasks/report — ні.
Якщо під час написання помітиш, що backfill-SQL реально хибний (напр. гілка UNION не спрацює) —
`## Discussion` у WIP, НЕ міняй міграцію сам.

## Перевірка (у WIP)
`cd backend && GOCACHE=/tmp/... go vet ./internal/course/...` і компіляція тесту
(`go test -tags=integration -run '^$' ./internal/course/adapters/postgres`). Реальний `make integration`
(Docker) — Courses 1.
