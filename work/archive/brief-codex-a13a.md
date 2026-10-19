# Codex brief — A13a: 2FA application + порти + SQL-текст + redis pending (без generated)

Репо `courses`, гілка `feat/accounts-selfservice`. A12 дав домен `two_factor.go` (TOTP verify, otpauth,
recovery codes, enroll/disable політики). Колонки `users.totp_secret/totp_enabled` + таблиця `recovery_codes`
вже в 00014. Наявне: `application/identity.go` Login (A04b remember-me+rate-limit), `SessionStore`, Redis
RateLimiter, argon2. Роль: реалізуєш A13a; генерацію/коміт — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a13a-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

ВАЖЛИВО: A13a НЕ чіпає `*/generated/*`, НЕ запускає sqlc, НЕ пише postgres-адаптерів на generated-типах (A13b).
Тут: порти, SQL-текст у `sql/queries/identity.sql`, redis pending-store, application use-cases (на фейках портів).

## 1. SQL-запити (текст у `sql/queries/identity.sql`, стиль наявних)
- `GetUserTwoFactor` (totp_secret, totp_enabled по user_id).
- `SetUserTotpSecret` (запис секрета, enabled лишається false — стан «очікує підтвердження»).
- `EnableUserTotp`, `DisableUserTotp` (DisableUserTotp: totp_secret=NULL, totp_enabled=false).
- `InsertRecoveryCodes` (батч; code_hash = SHA-256 digest коду — детермінований, щоб consume був атомарним).
- `ConsumeRecoveryCode` (АТОМАРНО: `UPDATE recovery_codes SET consumed_at=now() WHERE user_id=$1 AND
  code_hash=$2 AND consumed_at IS NULL RETURNING id`).
- `DeleteRecoveryCodesForUser` (для регенерації/вимкнення).

## 2. Порти (`identity/ports`)
- `TwoFactorStore`: GetTwoFactor, SetTotpSecret, EnableTotp, DisableTotp.
- `RecoveryCodeStore`: InsertCodes([]digest), ConsumeCode(userID, digest) (bool/err), DeleteCodes(userID).
- `PendingTwoFactorStore`: Put(ctx, pendingID, data{userID, rememberMe}, ttl), Take(ctx, pendingID) (one-time).

## 3. redis PendingTwoFactorStore (`identity/adapters/redis`, патерн RateLimiter/playback redis)
- Зберігає pending-логін ~5 хв, прив'язаний до user (і IP як частина перевірки в application), **одноразовий**
  (Take видаляє). Несе прапор rememberMe. Integration-тест (redis testcontainer).

## 4. Application use-cases (на фейках портів, TDD)
- **EnrollTwoFactor(ctx, userID)**: якщо вже enabled → помилка; інакше генерує секрет (домен) → SetTotpSecret →
  повертає secret + otpauth URI (для QR). Ще НЕ enabled.
- **ConfirmTwoFactor(ctx, userID, code)**: домен Verify(secret, code) → EnableTotp → генерує 10 recovery codes
  (домен), хешує SHA-256, InsertCodes → повертає plaintext-коди **один раз**.
- **Login стає двокроковим**: якщо `totp_enabled` — НЕ видавати сесію; покласти pending у PendingTwoFactorStore
  (TTL ~5хв, rememberMe), повернути `two_factor_required` + pendingID. Якщо вимкнено — як зараз.
- **CompleteTwoFactor(ctx, pendingID, code)**: Take pending (one-time) → перевір TOTP **або** ConsumeRecoveryCode
  (атомарно) → видати сесію (remember-me з pending). Rate-limit спроб (RateLimiter, ключ per-user+pending, 5 спроб).
- **DisableTwoFactor(ctx, userID, password, code)**: перевір пароль (PasswordHasher) + валідний TOTP/recovery →
  DisableTotp + DeleteCodes.
- **RegenerateRecoveryCodes(ctx, userID, code?)**: DeleteCodes → нові 10 → повернути раз.
- **AdminDisableTwoFactor(ctx, adminActor, targetUserID)**: лише адмін → DisableTotp + DeleteCodes; структурований
  лог-івент (без секретів), не DB-audit.

## Правила
- Конституція: application лише домен+порти; TOTP/коди — домен; сховища — порти. Без секретів/кодів у логах.
- НЕ чіпай generated, НЕ запускай sqlc, НЕ пиши postgres-адаптер (A13b). НЕ комість. TDD; час — спершу тести.

## Критерій готовності
- `go build ./...`, `go vet`, `gofmt` чисті (окрім ізольованих TODO для A13b-persist, якщо є — познач).
- `go test ./internal/identity/application/... ./internal/identity/adapters/redis/...` зелені:
  enroll/confirm, двокроковий login→complete (TOTP і recovery), disable, regenerate, admin-disable, rate-limit, pending one-time.
- SQL-текст усіх запитів у identity.sql; порти визначені; redis pending-store з integration-тестом.
