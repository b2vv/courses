# Codex brief — A12: домен 2FA (TOTP RFC6238 + резервні коди)

Репо `courses`, гілка `feat/accounts-selfservice`. Міграція 00014 (A02) ВЖЕ має `users.totp_secret`,
`users.totp_enabled`, таблицю `recovery_codes` — НОВОЇ міграції не треба. A12 — лише ЧИСТИЙ домен у
`backend/internal/identity/domain` (без I/O, без adapters, без sqlc). Роль: реалізуєш A12; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a12-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити (TDD, таблиці тестів спершу)
1. **TOTP (RFC 6238)** — stdlib `crypto/hmac`+`crypto/sha1`+`encoding/base32`+`encoding/binary`:
   - Генерація секрета (достатня ентропія, base32 без padding для otpauth).
   - `Verify(secret, code string, now time.Time) bool` — крок 30с, 6 цифр, **вікно ±1 крок** (prev/curr/next),
     constant-time порівняння коду.
   - Побудова `otpauth://totp/...` URI (issuer «CourseForge» або як у конфізі, account=email) — чистий хелпер
     (QR малює фронт, не домен).
2. **Резервні коди**: генерація набору (напр. 10) випадкових кодів (формат читабельний), повернення plaintext
   (показати користувачу раз) + їх для хешування (хеш робить adapter/argon2 в A13 — домен дає генерацію формату
   й політику одноразовості/кількості, не сам хеш, якщо хеш потребує I/O; але якщо хеш — чистий, можна тут).
3. **Політики**:
   - Enroll: 2FA вмикається лише після підтвердження валідним TOTP-кодом (домен дає перевірку, що код валідний
     для секрета перед enable).
   - Disable: вимагає валідний TOTP **або** резервний код (+ пароль — перевірка пароля в application).
   - Доменні помилки (invalid_totp, twofactor_already_enabled/not_enabled, recovery_code_invalid) у стилі `errors.go`.

## Правила
- Чистий домен: stdlib + `contracts/*` лише; жодного framework/DB. Юніт-таблиці на межах.
- **НЕ** додавай міграцій (колонки вже є), НЕ пиши adapters/application (A13), НЕ запускай sqlc, generated не чіпай.
- НЕ комість. TDD; час вичерпується — спершу тести вже-написаного.

## Критерій готовності
- `go build ./...`, `go vet ./internal/identity/domain/...`, `gofmt` чисті.
- `go test ./internal/identity/domain/...` зелені: TOTP verify (валідний curr, ±1 вікно, невалідний, поза вікном,
  constant-time не ламає логіку), otpauth URI формат, резервні коди (унікальність/кількість/формат), enroll/disable політики.
- Домен лишається без імпортів поза stdlib/contracts.
