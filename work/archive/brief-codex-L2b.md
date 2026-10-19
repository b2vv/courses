# Codex brief — L2b: enrollment write operations (orders application + adapters)

Репо `courses`, гілка `feat/enrollment`. План `docs/enrollment/plan.md` §3/§4/§8 (T2). L1+L2a вже в дереві
(коміти 35cf77b, a4cffc4): домен має self/blocked/revoked + переходи Block/Unblock/Revoke/SelfEnroll;
`CanStartOrder(*Grant)`; `GrantRepository.Membership(user,course) (*Grant, error)`; один рядок на (user,course),
повна UNIQUE. Роль: реалізуєш L2b; коміт/генерація/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L2b-codex.md
Append ПІСЛЯ КОЖНОГО КРОКУ. Перший рядок `ДАЛІ:`. Якщо наближаєшся до ліміту часу — зупинись на межі ФАЗИ,
запиши в робочий файл «RESUME: фаза N, що лишилось», Courses 1 відновить тебе за id. Фінал = стислий переказ.

## Межа L2b
Write-операції членства в МОДУЛІ orders: Enroll(self), Block, Unblock, Revoke, MembershipState, і ревізія
admin-grant (revoked→active). Roster — це L2c, НЕ чіпай. HTTP/openapi/Orval — L6, НЕ чіпай (лише application
+ domain + postgres-adapter + ports + тести).

## Контекст коду (звірено)
- `application/service.go`: `Actor{UserID, Administrator, EmailVerified}`. Власник курсу = `course.instructor_id
  == actor.UserID`. `GrantCourse(actor,userID,courseID,reason)` — admin-only, self-forbidden, reason 1..500,
  робить `GrantWithAudit` (плоский INSERT, :135 → ErrCourseAlreadyOwned на будь-якому наявному рядку).
- `ports.CourseReader.CourseForPurchase` читає `courses` прямо в orders-адаптері (patterна — `repository.go:22`):
  SELECT status,title,price_minor,currency. Додамо аналогічний enrollment-read.
- `courses`: status ∈ {draft,published,archived}; price_minor NULL = **free**; instructor_id NOT NULL.
- `grant_audit(grant_id,actor_id,action,reason,payment_event_id,created_at)` — action вільний текст.
- Домен переходи (є): `Block` active→blocked, `Unblock` blocked→active, `Revoke` active|blocked→revoked,
  `SelfEnroll` → new active/self. Треба додати revoked→active (нижче).

## Фаза 1 — домен (`domain/grant.go` + `errors.go`) [TDD, юніти першими]
- Додай перехід **`Reinstate(source GrantSource, grantedBy *string) (Grant, error)`**: дозволено лише з
  `revoked`; → `active` з указаним source (self — для free re-enroll; admin — для admin re-grant), order_id
  лишається nil (інваріант purchase↔order тримається). Інакше `ErrGrantTransitionInvalid`.
- Нові коди в `errors.go`: `ErrCourseNotEnrollable = "course_not_enrollable"` (enroll на paid/draft/archived),
  `ErrEnrollmentForbidden = "enrollment_forbidden"` (block/unblock не власник-і-не-admin; revoke не admin).
- Юніт-таблиця: Reinstate з кожного стану (лише revoked→ok), обидва source; незмінність інших полів.

## Фаза 2 — порт+адаптер читання курсу для enrollment
- `ports.CourseReader`: додай `CourseForEnrollment(ctx, courseID) (EnrollmentFacts, error)`, де
  `EnrollmentFacts{Status string; Free bool; InstructorID string}`. (Не плутай із CourseForPurchase.)
- `adapters/postgres/repository.go`: реалізуй — SELECT status, (price_minor IS NULL) AS free, instructor_id
  FROM courses WHERE id=…; ErrNoRows → `ErrCourseNotEnrollable`.

## Фаза 3 — репозиторій: персист переходів (один рядок, UPSERT/UPDATE + audit, у tx)
Додай у `GrantRepository` (порт+impl) методи, кожен у транзакції з рядком `grant_audit`:
- `CreateMembership(ctx, grant domain.Grant, action, reason string, actorID *string) (domain.Grant, error)` —
  INSERT нового рядка (для enroll-з-none, admin-grant-з-none). На конфлікті UNIQUE → `ErrCourseAlreadyOwned`.
- `UpdateMembership(ctx, grant domain.Grant, action, reason string, actorID *string) (domain.Grant, error)` —
  UPDATE course_grants SET source=,order_id=,state=,granted_by= WHERE id=$grant.ID(); 0 рядків → `ErrGrantInvalid`.
Обидва пишуть `grant_audit(grant_id,actor_id,action,reason,created_at)`. (НЕ використовуй сирий ON CONFLICT тут —
застосунок уже прочитав Membership і обрав перехід; це робить рішення видимим і тестованим.)

## Фаза 4 — application use-cases (`application/service.go`) [juніти з фейками]
Усі операції спершу читають `Membership`; домен-перехід; персист. Ідемпотентність — нижче.
- **`Enroll(ctx, actor, courseID) (domain.Grant, error)`** (G2/G10/En01/En14):
  CourseForEnrollment: якщо `Status!='published'` АБО `!Free` → `ErrCourseNotEnrollable` (archived/draft/paid).
  Membership: none → `SelfEnroll`→`CreateMembership(action='enroll')`; revoked → `Reinstate(self)`→
  `UpdateMembership(action='enroll')` (вільний free re-enroll, L1#2 дозволяє саме для free); active → no-op
  поверни наявний; blocked → no-op поверни наявний (НЕ знімай мют самозаписом). Будь-який authenticated user —
  роль не змінюється (G10).
- **`Block(ctx, actor, userID, courseID) (domain.Grant, error)`** (G3/En05/En06):
  authz: `actor.Administrator` АБО `actor.UserID == CourseForEnrollment.InstructorID`, інакше
  `ErrEnrollmentForbidden`. Membership: active → `Block`→UpdateMembership(action='block'); blocked → no-op;
  none/revoked → `ErrGrantTransitionInvalid`.
- **`Unblock(ctx, actor, userID, courseID)`** (En06): authz як Block. blocked → `Unblock`→Update(action='unblock');
  active → no-op; none/revoked → `ErrGrantTransitionInvalid`.
- **`Revoke(ctx, actor, userID, courseID)`** (G4/En04): authz: **лише** `actor.Administrator`, інакше
  `ErrEnrollmentForbidden`. active|blocked → `Revoke`→Update(action='revoke'); revoked → **no-op** (L1#4
  ідемпотентність, повтор admin-revoke = той самий стан); none → `ErrGrantInvalid`.
- **`MembershipState(ctx, actor, userID, courseID) (domain.GrantState, bool, error)`**: authz: self
  (`actor.UserID==userID`) АБО admin АБО власник курсу, інакше `ErrEnrollmentForbidden`. Повертає
  (state, found). (Read-only; для L6/фронту «ви учасник».)
- **Ревізія `GrantCourse`** (admin re-grant, En10-суміжне): лишити admin-only+self-forbidden+reason. Замінити
  плоский insert на: Membership → none → CreateMembership(admin/active, action='grant'); revoked →
  `Reinstate(admin)`→UpdateMembership(action='grant'); active|blocked → `ErrCourseAlreadyOwned`. (Так admin
  відновлює revoked-доступ — єдиний шлях для paid revoked; free revoked ще й через self Enroll.)

## Фаза 5 — інтеграційні тести (`adapters/postgres/..._integration_test.go`)
- En01 Enroll free published → 1 рядок active/self, audit 'enroll'.
- En14 instructor/admin Enroll на ЧУЖОМУ free курсі → active/self (роль не змінилась).
- En15 Enroll коли вже active → без 2-го рядка (count=1), no-op.
- L1#2: free **revoked** → Enroll → active/self (той самий рядок, count=1); **paid** revoked → Enroll →
  `ErrCourseNotEnrollable` (paid не self-реактивується; лише admin).
- Enroll на archived/draft/paid → `ErrCourseNotEnrollable`.
- Block active→blocked (власник і admin); Block чужим не-власником → `ErrEnrollmentForbidden`; Unblock→active.
- Revoke admin active→revoked; revoke не-admin → forbidden; **re-revoke → no-op** (стан revoked, без помилки).
- admin GrantCourse на revoked → active (count=1, audit 'grant'); на active → ErrCourseAlreadyOwned.

## Правила
- Домен чистий. Онови ВСІ реалізації порту й фейки (компілятор знайде: service_test, handler_test).
- sqlc НЕ генеруй (сирі queries, як наявні). generated не чіпай. НЕ комість.
- gofmt; `go build ./...`, `go vet ./internal/orders/...`,
  `GOCACHE=/private/tmp/courses-l2b-gocache go test ./internal/orders/...` (юніти) зелені.
- Інтеграцію (Docker) проганяє Courses 1 — пиши тести й кажи, що перевірити.

## Критерій готовності
- Юніти (домен Reinstate-таблиця + application-authz з фейками) зелені; build/vet/gofmt чисті.
- У робочому файлі: які переходи куди лягли; як виражено ідемпотентність (no-op) кожної операції; перелік
  інтеграційних сценаріїв; чи щось лишилось (RESUME).
