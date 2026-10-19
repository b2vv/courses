# Codex brief — A03b: postgres-адаптери UserWriter/TokenStore + mapper + 23505

Репо `courses`, гілка `feat/accounts-selfservice`. A03a в дереві (порти `identity/ports/accounts.go`, SQL-запити).
Courses 1 уже прогнав `sqlc generate` — у `internal/identity/adapters/postgres/generated/identity.sql.go` вже є:
`CreateUser`, `SetUserEmailVerified`, `UpdateUserPassword`, `UpdateUserEmail`, `InsertAccountToken`,
`ConsumeAccountToken`, `InvalidateAccountTokensByPurpose`, `DeleteUserSessionsExceptCurrent`, і resolve-row
тепер несе `email_verified`. Роль: реалізуєш A03b; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a03b-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити (у `internal/identity/adapters/postgres`, патерн — наявний `repository.go`/`administration.go`)

1. **Реалізуй `ports.UserWriter`** поверх згенерованих `CreateUser`/`SetUserEmailVerified`/`UpdateUserPassword`/
   `UpdateUserEmail`. Явний маппінг generated-row → `domain.User` (включно з `EmailVerified`).
   - **CreateUser: мапни Postgres `unique_violation` (SQLSTATE 23505) на доменну помилку `email_taken`**
     (через `*pgconn.PgError` + `errors.As`, перевір `.Code == "23505"`; звір, чи є вже доменний код —
     якщо ні, додай `ErrEmailTaken Error = "email_taken"` у `domain/errors.go`). Інші помилки — обгортай.
2. **Реалізуй `ports.TokenStore`** поверх `InsertAccountToken` / `ConsumeAccountToken` / `InvalidateAccountTokensByPurpose`.
   - `Consume` вже атомарний у SQL (`consumed_at IS NULL AND expires_at>now() RETURNING`); на 0 рядків
     (`pgx.ErrNoRows`) повертай доменну помилку на кшталт `token_invalid`/`token_used` (звір стиль errors.go).
3. **Новий запит сесій**: `DeleteUserSessionsExceptCurrent` підключи до session-репозиторію (для change-password,
   G8). `ResetPassword` і далі юзає наявний delete-all; тут лише «крім поточної» (digest поточної сесії).
4. **Mapper EmailVerified у resolve**: онови маппінг у session-resolve/`Resolve`, щоб `domain.User.EmailVerified`
   наповнювався з нового стовпця (A03a лишив це на A03b). Переконайся, що наявні identity-тести лишаються зелені.
5. **Integration-тести** (`*_integration_test.go`, тег як у наявних repository-тестів, через `testsupport/containers`):
   - CreateUser happy + повторний той самий email → `email_taken` (НЕ raw pg error); race-подібний випадок теж.
   - Token: Insert → Consume (успіх, повертає user_id/new_email) → повторний Consume → помилка (вжитий);
     протермінований → помилка; InvalidateByPurpose робить попередні невживними.
   - DeleteUserSessionsExceptCurrent лишає рівно поточну.
   - email_verified round-trip: CreateUser(false) → SetUserEmailVerified → Resolve бачить true.

## Правила
- Конституція: generated лишається в адаптері, явний маппінг у домен; домен не імпортує адаптер. Без ORM.
- Без секретів у логах; помилки явні, обгорнуті; URL без query у логах.
- TDD де практично (репозиторні — integration). Якщо впираєшся в час — спершу дороби тести вже-написаного.
- НЕ комість. Якщо довелося ще раз торкнутись SQL — НЕ запускай sqlc сам, познач, що треба регенерувати
  (Courses 1 прожене), але в ідеалі A03b нових запитів не додає.

## Критерій готовності
- `go build ./...`, `go vet ./internal/identity/...`, `gofmt` чисті.
- `go test ./internal/identity/...` (unit) зелені; integration-тести репозиторію акаунтів проходять під
  `make integration`-харнесом (Courses 1 прожене).
- `check-generated` не червоніє (generated уже згенеровано Courses 1; A03b його не має змінювати знову).
