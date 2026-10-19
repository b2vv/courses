# Codex brief — L2c: course roster with progress (orders application + adapter, read model)

Репо `courses`, гілка `feat/enrollment`. План `docs/enrollment/plan.md` §7 (En08). L1/L2a/L2b у дереві
(35cf77b, a4cffc4, c5df464). Роль: реалізуєш L2c; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L2c-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`; якщо близько до ліміту — зупинись на межі фази з `RESUME`.
Фінал = стислий переказ.

## Межа L2c
ЛИШЕ read-model ростера в МОДУЛІ orders: порт + postgres-запит + application-метод + юніт/інтеграція. HTTP/фронт
(L6/L7) НЕ чіпай. Тільки читання — жодних write.

## Що це (En08 / G7)
Інструктор СВОГО курсу (або admin) бачить ростер: для кожного учасника — **ім'я, % пройденого, статус
(стан членства), дата запису**; **пагінація**; **email НЕ показувати**. 0 уроків → 0%; порожній ростер → порожньо.

## Контекст коду (звірено)
- Членство = `course_grants` (orders). Учасники ростера = `state IN ('active','blocked')` (поточні;
  `revoked` = знятий доступ, у ростер НЕ включаємо — познач це рішення в робочому файлі).
- **Прогрес-формула — мірор `discussion/adapters/postgres/review.go` (reviewFactsQuery, рядки 20-22):**
  completed = `(SELECT count(*) FROM course_modules m JOIN course_lessons l ON l.module_id=m.id JOIN
  lesson_progress p ON p.lesson_id=l.id WHERE m.course_id=$course AND p.user_id=<member> AND p.completed_at
  IS NOT NULL)`; total = `(SELECT count(*) FROM course_modules m JOIN course_lessons l ON l.module_id=m.id
  WHERE m.course_id=$course)`. `completed_at` семантику (≥95% тривалості) володіє playback — ЧИТАЙ колонку,
  НЕ реімплементуй поріг.
- Ім'я учасника — `users.display_name` (JOIN; НЕ селектуй email).
- Авторизація й курс-факти — `CourseForEnrollment` (status/free/instructor_id) уже є (L2b).
- **Keyset-пагінація — мірор `discussion/adapters/postgres/moderation.go`** (`encodeCursor`/`decodeCursor`,
  base64 RawURLEncoding від (timestamp,id); `domain.ErrInvalidCursor` на биті). Останню сторінку → **порожній
  NextCursor** (урок B-28: не null-костиль).

## Реалізація (TDD)
### 1. Порт (`ports/repository.go` + новий тип)
- `RosterEntry{UserID string; DisplayName string; State domain.GrantState; EnrolledAt time.Time; ProgressPercent int}`.
- `RosterPage{Entries []RosterEntry; NextCursor string}`.
- `GrantRepository.Roster(ctx, courseID, cursor string, limit int) (RosterPage, error)`.

### 2. Postgres (`adapters/postgres/`)
- Запит: `FROM course_grants g JOIN users u ON u.id=g.user_id WHERE g.course_id=$course AND g.state IN
  ('active','blocked')`, корельовані підзапити completed/total (мірор review.go), `progress_percent` рахуй у SQL
  **round-half-up, точно як review** (`discussion/domain/review.go:75` = `(Completed*100+Total/2)/Total`):
  `CASE WHEN total=0 THEN 0 ELSE (completed*100 + total/2)/total END` — НЕ truncating `completed*100/total`
  (2/3 → review показує 67, truncation дала б 66; та сама пара має давати те саме число). Інтеграція мусить
  мати кейс 2/3 → 67.
- **Keyset:** ORDER BY `g.granted_at DESC, g.user_id ASC`; курсор несе (granted_at, user_id); фільтр
  наступної сторінки `(g.granted_at, g.user_id) < (cursorStamp, cursorID)` у тому ж порядку. Вибирай `limit+1`
  рядків, зайвий → є наступна сторінка → `NextCursor = encodeCursor(last.granted_at, last.user_id)`, інакше "".
- Clamp limit у application (див. нижче); порожній cursor "" → перша сторінка.
- Мапінг стану в `domain.GrantState`. НЕ селектуй email (ревʼю перевірить SQL на відсутність `email`).

### 3. Application (`application/service.go`)
- `Roster(ctx, actor Actor, courseID, cursor string, limit int) (ports.RosterPage, error)`:
  `CourseForEnrollment`; authz: `actor.Administrator` АБО `actor.UserID == facts.InstructorID`, інакше
  `ErrEnrollmentForbidden` (як MembershipState, але БЕЗ self — ростер лише власнику/адміну).
  Clamp `limit` (default напр. 20, max 100, min 1). Делегуй у repo.Roster. (Percent уже порахований у SQL.)

## Інтеграційні тести (`adapters/postgres/..._integration_test.go`)
- En08 happy: курс із кількома учасниками (active+blocked) + різним прогресом → рядки з коректними name/state/
  granted_at/percent; **email у відповіді відсутній** (структура його не несе — досить, що порту нема поля).
- 0 уроків → percent=0 для всіх; учасник без прогресу → 0; повністю пройдений → 100.
- revoked-учасник **не** з'являється у ростері.
- Пагінація: limit=1 на 2+ учасниках → NextCursor непорожній; друга сторінка з курсором → решта; остання
  сторінка → NextCursor "".  Невалідний курсор → `ErrInvalidCursor` (переконайся, що такий код існує в
  orders/domain або додай).
- authz: не-власник не-admin → `ErrEnrollmentForbidden`.
- Порожній ростер (курс без учасників) → 0 рядків, NextCursor "".

## Application-юніти (фейки)
- authz (owner/admin/сторонній), clamp limit, проброс курсора/сторінок (фейк повертає сторінки).

## Правила
- Онови ВСІ реалізації порту й фейки (service_test, handler_test) — додасться метод Roster.
- sqlc НЕ генеруй (сирий SQL, як наявні). generated не чіпай. НЕ комість.
- gofmt; `go build ./...`, `go vet ./internal/orders/...`, `GOCACHE=/private/tmp/courses-l2c-gocache
  go test ./internal/orders/...` зелені. Інтеграцію (Docker) проганяє Courses 1.

## Критерій готовності
- Юніти зелені; build/vet/gofmt чисті. Робочий файл: рішення (revoked виключено; clamp-значення; де рахується
  percent), перелік інтеграційних сценаріїв, чи `ErrInvalidCursor` новий чи наявний.
