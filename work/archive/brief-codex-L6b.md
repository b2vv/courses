# Brief — L6b: discussion person-report HTTP endpoints

РОБОЧИЙ ФАЙЛ: `work/_wip-L6b-codex.md`
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого інструменту:
рядок `ДАЛІ:` з тим, що збираєшся перевіряти. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ файлу.

## Контекст
Фаза `enrollment`, гілка `feat/enrollment`. L5 (person-reports backend у discussion) ВЖЕ
закомічено (`5347545`) і заморожено: домен/application/ports/postgres НЕ чіпати. L6a (orders
HTTP) закомічено (`c51db05`). Твоє завдання — лише **HTTP-шар discussion** для вже готових
application-операцій + OpenAPI + Orval + contract-тести.

Application-методи (вже є, сигнатури фіксовані — `backend/internal/discussion/application/application.go`):
- `ReportLecturer(ctx, actor Actor, courseID, reason string) (domain.ParticipantReport, error)`
- `ReportStudent(ctx, actor Actor, courseID, targetID, reason string) (domain.ParticipantReport, error)`
- `ReportQueue(ctx, actor Actor, cursor string) (ParticipantReportQueuePage, error)`
- `ResolveReport(ctx, actor Actor, reportID string) error`
Типи сторінки: `application.ParticipantReportQueuePage{ Items []ParticipantReportQueueItem; NextCursor string }`,
`ParticipantReportQueueItem{ ReporterName *string; TargetName *string; CourseTitle string; Report domain.ParticipantReport }`.
Доменні коди помилок (вже є — `discussion/domain/errors.go`): `ErrReportSelf`="report_own_person",
`ErrReportInvalid`="report_invalid", `ErrReportForbidden`="report_forbidden",
`ErrReportDuplicate`="report_already_open"; плюс `application.ErrEmailNotVerified`="email_not_verified",
`domain.ErrNotFound` (course не знайдено).

## Маршрути (net/http ServeMux, дзеркаль наявні у `discussion/adapters/http/handler.go`)
Усі хендлери — ТОНКІ: дістати actor, зчитати transport-вхід, викликати application, відрендерити;
authz лишається в application.

1. `POST /api/v1/learn/courses/{courseId}/report-lecturer` — у `Register(...)`,
   `sameOrigin(authenticate(...))`. Body `{ "reason": string }`. Викликає `ReportLecturer`.
2. `POST /api/v1/teach/courses/{courseId}/participants/{userId}/report` — у `RegisterTeach(...)`,
   `sameOrigin(authenticate(...))`. Body `{ "reason": string }`. Викликає `ReportStudent(courseID, userId, reason)`.
3. `GET /api/v1/admin/participant-reports?cursor=` — у **новому** `RegisterAdmin(mux, authenticate, sameOrigin, administrator)`,
   обгортка `authenticate(administrator(...))` (defense-in-depth, дзеркаль orders `RegisterAdmin`/grantCourse;
   application теж перевіряє admin). Повертає сторінку черги.
4. `POST /api/v1/admin/participant-reports/{reportId}/resolve` — у тому ж `RegisterAdmin`,
   `sameOrigin(authenticate(administrator(...)))`. Викликає `ResolveReport`. 204/200 на успіх (idempotent).

Додай `RegisterAdmin` до discussion `Handler` і виклич його в `backend/cmd/api/main.go` там, де
реєструються інші discussion-групи, передавши наявний `administrator` middleware (глянь, як
orders-handler отримує його в main.go).

## Відповіді / DTO
- Report-створення (1,2): на успіх 201 з DTO репорту (`id`, `courseId`, `reporterId`, `targetId`,
  `kind`, `reason`, `state`, `createdAt`). Дзеркаль стиль наявних review/comment DTO.
- Черга (3): `{ "items": [ { report поля + reporterName?:string|null, targetName?:string|null,
  courseTitle } ], "nextCursor": string|null }`. **Email НЕ включати.** `nextCursor` — через
  `nextCursorOrNull` (null на порожньому, патерн B-28, уже є в handler.go).
- Resolve (4): 204 No Content.

## Мапінг помилок — розшир `statusFor` і `writeError` у handler.go
`statusFor`:
- `ErrReportForbidden` → 403 (додати до forbidden-кейсу поряд із `ErrReportOwnReview`).
- `ErrReportSelf` → 403 (дзеркаль `ErrReportOwnReview`).
- `ErrReportDuplicate` → 409 (поряд із conflict-кейсом).
- `ErrReportInvalid` → 400 (поряд із bad-request-кейсом).
- `domain.ErrNotFound` → у хендлерах report map explicitly 404 (дзеркаль наявні
  `course_not_found`/`comment_not_found` writeError-патерни), НЕ через statusFor default.
`writeError`: додай людські message для `report_own_person`, `report_invalid`, `report_forbidden`,
`report_already_open` (стиль наявних кейсів). `email_not_verified` вже мапиться (403 через statusFor).
**Порядок:** email-first уже забезпечено в application (verified → потім факти) — у HTTP нічого
додаткового не роби, просто пропагуй код.

## OpenAPI + Orval
- Додай 4 paths + request/response schemas у `backend/openapi/openapi.yaml`; додай нові error-коди
  до enum `ErrorResponse.error.code`: `report_own_person`, `report_invalid`, `report_forbidden`,
  `report_already_open` (якщо якогось немає). Roster/членство L6a enum уже додані — не дублюй.
- Регенеруй Orval: `make generate` (або та ціль, що генерує `frontend/src/adapters/api/generated/`).
  Переконайся `make check-generated` зелений **зі staged generated** (git add перед перевіркою —
  ціль порівнює з індексом).

## Тести
- Contract-validated HTTP-тести (`discussion/adapters/http/handler_test.go` або новий
  `participant_report_handler_test.go`) на всі 4 маршрути: happy + ключові error (report-lecturer
  не-член → 403 report_forbidden; self → 403 report_own_person; дубль → 409 report_already_open;
  unverified → 403 email_not_verified; course not found → 404; queue без admin → 403; resolve без
  admin → 403; roster/queue DTO БЕЗ email). Використай `assertContract`/наявний contract-хелпер,
  як у review/comment тестах. fakeRepository вже має інертні методи — заміни на керовані стаби.
- НЕ змінюй L5 application/domain/postgres логіку. Якщо треба — лише розширити fake у тестах.

## Межі
- НЕ чіпати: домен/application/ports/postgres L5; orders; будь-який інший модуль.
- НЕ комітити, НЕ пушити. Коміт і фінальні гейти — Courses 1.
- generated/sqlc тільки через `make generate`, не руками.

## Перевірка перед здачею (прожени, впиши результати у WIP)
`go build ./...`; `go test ./internal/discussion/...`; `go test ./platform/architecture/...`
(constitution: модуль не імпортує модуль, generated в адаптерах); `git diff --check`;
`make check-generated` (зі staged generated); `cd frontend && tsc` (strict, generated типи).
Реальний `make integration` лишаєш Courses 1 (Docker у sandbox недоступний).

## Рішення щодо контракту ID (Courses 1, 2026-10-07) — авторизовано через discussion-канал
Codex коректно зупинився: `InsertParticipantReport` генерує UUID у SQL і відкидає його, тож
повернений `report.ID` порожній і HTTP не може чесно віддати `id`. Реальний латентний дефект L5.
**Фікс (мінімальний, дзеркало `CreateReview`, що повертає persisted-запис):**
- Порт: `InsertParticipantReport(ctx, report) (string, error)` — повертає persisted `id`.
- Postgres: `INSERT … ON CONFLICT … WHERE state='open' DO NOTHING RETURNING id`; через `QueryRow`;
  на конфлікті рядок не повертається → `pgx.ErrNoRows` → `domain.ErrReportDuplicate` (семантика збережена).
- Application `ReportLecturer`/`ReportStudent`: `id, err := repo.Insert(...)`; `report.ID = id`; повернути report.
- Fakes (handler_test.go, review_test.go, будь-який integration fake) → повертають стаб-id ("" більше не валідний для happy-path; у тестах поверни валідний UUID-стаб).
- Покриття: integration — повернений id є валідним UUID і дорівнює рядку в черзі; unit HTTP — 201 DTO несе цей id.
Дозволено чіпати L5 application/ports/postgres **лише для цього фікса**; решта L5-логіки заморожена.
Далі — реалізуй L6b HTTP/OpenAPI/Orval/тести як у брифінгу. git/tasks/report/generated-руками — не чіпати.
