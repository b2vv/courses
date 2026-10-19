# Codex brief — L5: person-reports (discussion) — domain + application + repository

Репо `courses`, гілка `feat/enrollment`. Spec G6/En07; план §8 L5. Таблиця `participant_reports` вже існує
(міграція 00015): `id, course_id, reporter_id, target_id, kind CHECK IN(on_lecturer,on_student), reason,
state CHECK IN(open,resolved) DEFAULT open, created_at, CHECK(reporter_id<>target_id)`, partial-unique
`participant_reports_one_open_per_pair` (course_id,reporter_id,target_id) WHERE state='open'. Роль: реалізуєш
L5-backend; HTTP — L6 (НЕ роби). Коміт/гейти/e2e — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L5-codex.md — append; перший рядок `ДАЛІ:`; близько
ліміту → RESUME. Фінал = переказ.

## Семантика (G6/G3/En07)
Скарги на ОСОБУ, окремо від модерації коментарів. Ескалація в адмін-чергу (над інструктором).
- **on_lecturer**: учасник курсу (член — active АБО blocked, G6 «доступно й заблокованому») скаржиться на
  ЛЕКТОРА (власник курсу = instructor_id). target = owner. reporter мусить бути членом; reporter ≠ owner.
- **on_student**: ЛЕКТОР (власник; або admin) скаржиться на учня (учасник-член). reporter = owner (або admin);
  target мусить бути членом; target ≠ reporter.
- **Адмін-черга**: список open-скарг + резолв — ТІЛЬКИ admin (G4/G6).
- **En07 відмови**: self-report (reporter==target) → відмова; дубль (другий open на ту саму трійку) → відмова.
- Членство читаємо ВЛАСНИМ SQL discussion (course_grants + courses.instructor_id), **БЕЗ імпорту orders**
  (ADR-013). `contracts/access` уже імпортується в discussion — використовуй `access.Membership`.

## 1. Домен `discussion/domain`
- Новий тип `ParticipantReport{ID, CourseID, ReporterID, TargetID, Kind, Reason, State, CreatedAt}` +
  консти `ReportOnLecturer="on_lecturer"`, `ReportOnStudent="on_student"`, `ReportOpen/Resolved`.
- `NewParticipantReport(id, courseID, reporterID, targetID string, kind ReportKind, reason string, now time.Time)
  (ParticipantReport, error)`: інваріанти — reporterID≠targetID (інакше `ErrReportSelf`); kind ∈ {on_lecturer,
  on_student}; reason TrimSpace + utf8 + 1..500 (як RemoveReview), інакше `ErrReportInvalid`; усі id непорожні.
- Нові коди в `discussion/domain/errors.go`: `ErrReportSelf="report_own_person"`, `ErrReportInvalid="report_invalid"`,
  `ErrReportForbidden="report_forbidden"`, `ErrReportDuplicate="report_already_open"`.
- Чистий «учасник» предикат: member = `membership == active || membership == blocked` (G6). Можеш додати
  маленький хелпер у домені або перевіряти в application.

## 2. Порти + репозиторій `discussion/ports` + `adapters/postgres`
Новий інтерфейс (або розширення): `ParticipantReportRepository`:
- `ReportSubject(ctx, courseID, userID string) (ownerID string, membership access.Membership, courseFound bool, err error)`
  — читає `courses.instructor_id` + LEFT JOIN `course_grants` (user,course) → state→Membership (none якщо нема).
  Курс не знайдено → courseFound=false.
- `InsertParticipantReport(ctx, ParticipantReport) error` — INSERT; на partial-unique-конфлікт (open-дубль) →
  `domain.ErrReportDuplicate` (ON CONFLICT ... DO NOTHING + 0 рядків → дубль).
- `ParticipantReportQueue(ctx, cursor string, limit int) (ParticipantReportQueuePage, error)` — лише open,
  keyset `ORDER BY created_at DESC, id DESC` (мірор moderation.go `encodeCursor`/limit+1); JOIN users для
  reporter/target `display_name` (НЕ email), courses для title. Елемент: report-поля + reporterName + targetName
  + courseTitle. Остання сторінка → порожній NextCursor.
- `ResolveParticipantReport(ctx, reportID string) error` — `UPDATE participant_reports SET state='resolved'
  WHERE id=$1 AND state='open'`; 0 рядків (вже resolved або нема) → трактуй як no-op (idempotent resolve;
  якщо id не існує взагалі — можна `domain.ErrNotFound`, обери й познач).

## 3. Application `discussion/application`
- `ReportLecturer(ctx, actor Actor, courseID, reason string) (domain.ParticipantReport, error)`:
  ReportSubject(courseID, actor.UserID) → ownerID, actorMembership. Якщо !courseFound → ErrNotFound.
  Якщо actor.UserID==ownerID → ErrReportSelf (лектор не скаржиться сам на себе як на лектора). Якщо actor НЕ
  член (membership не active/blocked) → ErrReportForbidden. NewParticipantReport(..., target=ownerID,
  kind=on_lecturer, reason) → InsertParticipantReport. (email_verified гейт — як на комунікації? Report — це
  комунікація-суміжне; постав `if !actor.EmailVerified` першим, як у write-paths L4b, для узгодженості.)
- `ReportStudent(ctx, actor Actor, courseID, targetID, reason string) (domain.ParticipantReport, error)`:
  ReportSubject(courseID, targetID) → ownerID, targetMembership. authz: actor.Administrator АБО
  actor.UserID==ownerID, інакше ErrReportForbidden. target має бути членом (active|blocked), інакше
  ErrReportForbidden (нема на кого — не учасник). target≠actor (NewParticipantReport дасть ErrReportSelf).
  kind=on_student, target=targetID. email-first.
- `ReportQueue(ctx, actor Actor, cursor string) (ParticipantReportQueuePage, error)`: лише `actor.Administrator`,
  інакше ErrReportForbidden. Делегуй repo.ParticipantReportQueue (clamp limit напр. default 20/max 100).
- `ResolveReport(ctx, actor Actor, reportID string) error`: лише admin, інакше ErrReportForbidden. Делегуй.

## 4. Тести
- Домен unit: NewParticipantReport — self-report, bad kind, reason-межі, happy.
- Application unit (fakes): ReportLecturer (member ok / blocked ok / non-member forbidden / self(owner) forbidden /
  unverified→email); ReportStudent (owner ok / admin ok / stranger forbidden / target-non-member forbidden);
  ReportQueue + ResolveReport (admin-only forbidden для не-admin). Дубль → ErrReportDuplicate (через фейк, що
  повертає його).
- Integration: сідай курс+власник+учасників (active/blocked/none); ReportLecturer учасником → рядок on_lecturer
  target=owner; дубль → ErrReportDuplicate; blocked-учасник теж може; non-member → forbidden; ReportStudent
  власником на члена → on_student; на не-члена → forbidden; Queue(admin) бачить open, не бачить resolved;
  Resolve(admin) → зникає з черги; Queue/Resolve не-admin → forbidden. Імен email у виводі немає.

## Правила
- discussion НЕ імпортує orders (членство — власним SQL). contracts/access — ок. generated/sqlc не чіпай.
  Фікстури ІЗОЛЬОВАНІ (власні course/users; НЕ спільний lesson b1 / course d1 — K09!). НЕ комість.
- gofmt; build; `go test ./internal/discussion/...` + architecture зелені. Integration — Courses 1.

## Критерій готовності
- Домен+application+repository+тести для 4 операцій; self-report і дубль відхиляються; адмін-черга лише admin;
  email у виводі немає; членство власним SQL. Робочий файл: рішення (resolve-idempotency; дубль-детект;
  які ізольовані фікстури).
