# Codex brief — A14: 2FA HTTP /auth/2fa/* + admin route + openapi + Login контракт

Репо `courses`, гілка `feat/accounts-selfservice`. A13 дав application 2FA use-cases (Enroll/Confirm/
двокроковий Login+Complete/Disable/Regenerate/AdminDisable) + postgres/redis сховища. Наявний identity HTTP —
`identity/adapters/http/` (handler + admin_handler). Роль: реалізуєш A14; генерація/коміт — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a14-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Маршрути (авторизовані, mutating → RequireSameOrigin; патерн наявних auth-маршрутів A06)
- `POST /api/v1/auth/2fa/enroll` → EnrollTwoFactor: повертає `secret` + `otpauthUri` (для QR). (Ще не enabled.)
- `POST /api/v1/auth/2fa/confirm` {code} → ConfirmTwoFactor: вмикає, повертає **recovery codes один раз**.
- `POST /api/v1/auth/2fa/complete` {pendingId, code} → CompleteTwoFactor: видає сесію-cookie (TOTP або recovery).
  Це НЕ авторизований маршрут (сесії ще нема) — лише валідний pending.
- `POST /api/v1/auth/2fa/disable` {password, code} → DisableTwoFactor.
- `POST /api/v1/auth/2fa/recovery-codes` {code?} → RegenerateRecoveryCodes (повертає нові один раз).
- `POST /api/v1/admin/users/{id}/2fa/disable` → AdminDisableTwoFactor (гейт адміна, патерн наявного admin reset).
- **Зміна контракту `POST /login`**: якщо 2FA увімкнено — відповідь `two_factor_required` + `pendingId`
  (НЕ ставити session-cookie); інакше як зараз. Онови відповідний DTO.

## Error→HTTP (стиль A06): invalid_totp/recovery_code_invalid → 401/422, twofactor_already_enabled → 409,
twofactor_not_enabled → 409, pending не знайдено/протерміновано → 401/410, rate-limited → 429.

## openapi + композиція
- Опиши всі нові ендпоінти, запити/відповіді (secret/otpauthUri, recoveryCodes[], two_factor_required+pendingId)
  у `openapi.yaml`. Orval НЕ запускай (Courses 1), generated не чіпай.
- Підключи 2FA use-cases у composition root (`cmd/api/main.go`): postgres TwoFactorStore/RecoveryCodeStore,
  redis PendingTwoFactorStore, rate-limiter — туди, де збирається Accounts/identity.

## Правила
- handler транслює транспорт; політики — в application (вже є). RequireSameOrigin на мутаціях. Без секретів/
  кодів/секрета TOTP у логах. Strict контракти.
- НЕ запускай sqlc/orval; openapi.yaml правиш, generated не чіпаєш. НЕ комість. Contract-тести HTTP.

## Критерій готовності
- `go build ./...`, `go vet`, `gofmt` чисті; `go test ./internal/identity/...` зелені (нові HTTP contract-тести:
  enroll повертає otpauth, confirm→recovery codes, login з 2FA→two_factor_required, complete видає сесію,
  disable, admin-disable гейт).
- `openapi.yaml` містить усі нові ендпоінти + зміну login; валідний YAML. (Courses 1 потім: Orval + повний гейт.)
