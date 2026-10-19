# Codex brief — A03a: шов EmailVerified + порти + SQL-запити + redis RateLimiter

Репо `courses`, гілка `feat/accounts-selfservice`. План `docs/accounts/plan.md` §2 (шов B2), §4b. A02 вже в дереві
(міграція 00014, `User.EmailVerified`, `account_policy.go`). Роль: реалізуєш A03a; коміт/генерація/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a03a-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

ВАЖЛИВО: A03a НЕ торкається `*/generated/*` і НЕ пише postgres-адаптерів на generated-типах (це A03b після
того, як Courses 1 прожене `make generate`). Тут лише: шов сесії, порти-інтерфейси, redis-адаптер, і SQL-запити
як ТЕКСТ у `sql/queries/identity.sql` (без виклику sqlc).

## Що зробити

1. **Шов `EmailVerified` (B2).** 
   - `platform/httpserver/actor.go`: додай поле `EmailVerified bool` у `Actor` (єдине місце конструювання —
     identity Authenticate; звір, що ніхто інший Actor не будує).
   - `identity/adapters/http/handler.go` (~:137): `Authenticate` кладе `user.EmailVerified` в `Actor`.
   - `SessionStore.Resolve` уже повертає `domain.User`; переконайся, що `User.EmailVerified` наповнюється — для
     цього онови запит у `sql/queries/identity.sql`, що читає користувача по сесії, щоб він SELECT-ив
     `users.email_verified` (генерацію НЕ запускай; лише текст запиту). Якщо мапінг рядок→domain.User у
     postgres-репозиторії посилається на поле — познач у робочому файлі, що маппер треба дочепити в A03b
     після генерації (не чіпай generated зараз).

2. **Порти** `backend/internal/identity/ports/` (лише інтерфейси, domain-safe):
   - `UserWriter`: `CreateUser(ctx, domain.User + password_hash) (domain.User, error)` (повертає з id),
     `SetEmailVerified(ctx, userID)`, `UpdatePassword(ctx, userID, hash)`, `UpdateEmail(ctx, userID, newEmail)`.
   - `TokenStore`: `Issue(ctx, domain.AccountToken, digest []byte)`, `Consume(ctx, purpose, digest) (…, error)`
     (атомарний consume — реалізація в A03b), `InvalidateByPurpose(ctx, userID, purpose)`.
   - `RateLimiter`: `Allow(ctx, key string) (bool, error)` + `Reset(ctx, key)` (для скидання при успіху).
   - `Mailer`: інтерфейс `SendVerification/SendPasswordReset/SendEmailChange(ctx, to, link/lang…)` —
     лише ІНТЕРФЕЙС; адаптер і шаблони — у A04.
   Імена/сигнатури узгодь із наявним стилем портів (`ports/authentication.go`, `ports/administration.go`).

3. **SQL-запити** — додай у `sql/queries/identity.sql` (назви у стилі наявних, `-- name: X :one/:exec/:many`):
   - CreateUser (з `email_verified`), SetUserEmailVerified, UpdateUserPassword, UpdateUserEmail.
   - InsertAccountToken, ConsumeAccountToken (АТОМАРНО: `UPDATE account_tokens SET consumed_at=now()
     WHERE token_digest=$1 AND purpose=$2 AND consumed_at IS NULL AND expires_at>now() RETURNING user_id, new_email`),
     InvalidateAccountTokensByPurpose (consumed_at=now() для активних того purpose).
   - DeleteUserSessionsExceptCurrent (`DELETE FROM user_sessions WHERE user_id=$1 AND token_digest<>$2`).
   - Запит resolve-користувача-по-сесії — додай `email_verified` у SELECT.
   НЕ запускай `sqlc generate` (Courses 1 зробить).

4. **redis RateLimiter** `backend/internal/identity/adapters/redis/` (патерн — `playback/adapters/redis`):
   реалізація `ports.RateLimiter` поверх `platform/redis`: INCR+EXPIRE вікно (fixed-window), `Allow` рахує й
   порівнює з лімітом, `Reset` видаляє ключ. **Fail-open** при помилці Redis із логом (plan §4b). Параметри
   ліміту (ліміт/вікно) конфігуровані. Integration-тест через `testsupport/containers.Start` (redis-контейнер).

## Правила
- Конституція: порти domain-safe; адаптери всередину; redis-клієнт не глобальний. Fail-open rate-limit — явно.
- TDD де можливо (redis — integration; шов — unit/контракт де є сенс).
- НЕ комість; НЕ чіпай generated; НЕ запускай sqlc. Якщо час вичерпується — дороби тести вже-написаного, познач залишок.

## Критерій готовності A03a
- `go build ./...` компілюється (окрім місць, що свідомо чекають generated у A03b — якщо такі є, вони мають
  бути ізольовані так, щоб решта збіркою не падала; інакше познач і лиши TODO-маркер у робочому файлі).
- `go vet`, `gofmt` чисті на нових пакетах.
- redis integration-тест RateLimiter проходить.
- `sql/queries/identity.sql` містить усі нові запити (текст), generated не чіпано.
