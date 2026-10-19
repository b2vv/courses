# Codex brief — L6a: orders HTTP endpoints (enroll/block/unblock/revoke/membership/roster)

Репо `courses`, гілка `feat/enrollment`. План §8 L6 (orders-частина). Усі application-методи вже в дереві
(L2b/L2c, коміти a4cffc4/c5df464/8c991ec). Роль: реалізуєш L6a-HTTP для orders; коміт/гейти/Orval/e2e — Courses 1.
**Межа: лише orders HTTP** (+ openapi paths + error-коди). Discussion report-endpoints — L6b (НЕ чіпай).

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L6a-codex.md — append; перший рядок `ДАЛІ:`;
близько ліміту → RESUME. Фінал = переказ.

## Патерн (мірор наявного `internal/orders/adapters/http/handler.go`)
Хендлер: `actorFromRequest` → decode path/body → `service.X(...)` → `httpserver.WriteJSON(status, dto)` або
`writeError`. Реєстрація — `Register`(authenticate, sameOrigin) і `RegisterAdmin`(…, administrator). Актор уже
несе UserID/Administrator/EmailVerified. Authz роблять APPLICATION-методи — хендлери тонкі.

## Application-методи (готові; сигнатури)
- `Enroll(ctx, actor, courseID) (domain.Grant, error)`
- `Block(ctx, actor, userID, courseID) (domain.Grant, error)`; `Unblock(...)`; `Revoke(...)`
- `MembershipState(ctx, actor, userID, courseID) (domain.GrantState, bool, error)`
- `Roster(ctx, actor, courseID, cursor string, limit int) (ports.RosterPage, error)`

## Маршрути (додати)
- `POST /api/v1/learn/courses/{courseId}/enroll` → Enroll (sameOrigin+authenticate). 201 з grant-DTO (як grantCourse DTO: state/source/… без зайвого). Самозапис — courseId з path, user = actor.
- `GET /api/v1/learn/courses/{courseId}/membership` → MembershipState(actor, actor.UserID, courseId) (authenticate). Повертає `{state: string|null, member: bool}` (found→member). (Self-перегляд свого членства; app дозволяє self.)
- `GET /api/v1/teach/courses/{courseId}/roster` → Roster(actor, courseId, cursor, limit) (authenticate; app перевіряє owner∨admin). Query `cursor`, `limit`. DTO `{entries:[{userId,displayName,state,enrolledAt,progressPercent}], nextCursor}`. Без email.
- `POST /api/v1/teach/courses/{courseId}/participants/{userId}/block` → Block (sameOrigin+authenticate). 200 grant-DTO.
- `POST /api/v1/teach/courses/{courseId}/participants/{userId}/unblock` → Unblock (sameOrigin+authenticate). 200.
- `POST /api/v1/admin/courses/{courseId}/participants/{userId}/revoke` → Revoke (sameOrigin+authenticate). 200. (Revoke app-перевіряє Administrator; НЕ став `administrator`-middleware, бо block/unblock — owner∨admin через ту саму teach-гілку; тримай authz у app, uniform.)
Додай нову reg-функцію якщо треба (напр. `RegisterTeach`) або впиши в Register/RegisterAdmin за наявним стилем композиційного кореня (звір `cmd/api/main.go`, як Register кличеться — онови виклик, якщо додаєш функцію).

## Error-мапінг (`writeError` switch) — додай
- `domain.ErrMembershipConflict` → **409** `membership_conflict` (Courses 2: конкурентний перехід = conflict).
- `domain.ErrCourseNotEnrollable` → **409** `course_not_enrollable` ("курс не можна записатись: paid/draft/archived").
- `domain.ErrEnrollmentForbidden` → **403** `enrollment_forbidden`.
(ErrCourseAlreadyOwned/ErrEmailNotVerified/ErrGrantInvalid уже мапляться.)

## OpenAPI (`backend/openapi/openapi.yaml`)
- Додай нові шляхи (6 вище) з параметрами/відповідями/DTO-схемами (membership, roster-page) у наявному стилі.
- Додай у error-enum: `membership_conflict`, `course_not_enrollable`, `enrollment_forbidden` (якщо ще нема).
- Orval НЕ генеруй (Courses 1 зробить `make generate` і звірить check-generated).

## Тести (`adapters/http/handler_test.go`, contract-validated проти openapi)
Кожен endpoint: happy + authz/error. Ключові:
- Enroll: active-member/owner → 201; paid/non-published → 409 course_not_enrollable; unverified → 403 email_not_verified; конкурентний (фейк повертає ErrMembershipConflict) → 409 membership_conflict.
- Block/Unblock: owner → 200; сторонній → 403 enrollment_forbidden; Revoke: admin → 200, не-admin → 403.
- MembershipState: self → {member:true,state}; stranger → 403 enrollment_forbidden (app); none → {member:false,state:null}.
- Roster: owner → page (без email-поля в DTO/схемі); stranger → 403; поганий cursor → 400 (ErrInvalidCursor).
(Усі коди мусять бути в openapi-enum, інакше contract-check впаде — як було в L4b.)

## Правила
- Тонкі хендлери; authz лишається в application. generated/sqlc НЕ чіпай (Orval — Courses 1). НЕ комість.
- gofmt; build; `go test ./internal/orders/...` зелені (contract-тести можуть вимагати openapi — додай коди/шляхи).
- Якщо додаєш reg-функцію — онови `cmd/api/main.go` виклик і переконайся `go build ./...` зелений.

## Критерій готовності
- 6 маршрутів + тонкі хендлери + DTO; error-мапінг (409 membership_conflict/course_not_enrollable, 403
  enrollment_forbidden); openapi шляхи+коди; handler-тести зелені (contract-validated); build. Робочий файл:
  які маршрути/DTO/коди додано, чи змінював main.go.
