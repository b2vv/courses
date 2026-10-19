# Codex brief — A06: HTTP-маршрути /api/v1/auth/* + openapi + me DTO

Репо `courses`, гілка `feat/accounts-selfservice`. A01–A05 у дереві: application use-cases (Register/VerifyEmail/
ResendVerification/RequestPasswordReset/ResetPassword/ChangePassword/RequestEmailChange/ConfirmEmailChange,
Login remember-me), Mailer, access-gate. Наявний identity HTTP — `identity/adapters/http/handler.go`
(login/me/logout, cookie HttpOnly+Secure+SameSite, RequireSameOrigin). Роль: реалізуєш A06; коміт+генерація — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a06-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити (у `identity/adapters/http`, патерн наявного handler.go/admin_handler.go)

1. **Маршрути під `/api/v1/auth/`** (mutating → `RequireSameOrigin`, тіла з лімітом, без кешу — як наявні):
   - `POST /register` → Register: 201 + ставить session-cookie (автологін); тіло {email,password,displayName};
     заголовок мови (Accept-Language) → lang листа.
   - `POST /verify-email` {token} → VerifyEmail.
   - `POST /resend-verification` (авторизований або {email}) → ResendVerification.
   - `POST /password/forgot` {email} → RequestPasswordReset (ЗАВЖДИ 200 нейтрально).
   - `POST /password/reset` {token,password} → ResetPassword.
   - `POST /password/change` (авторизований) {oldPassword,newPassword} → ChangePassword (поточний session-digest
     із cookie → крім поточної).
   - `POST /email/change` (авторизований) {newEmail} → RequestEmailChange.
   - `POST /email/change/confirm` {token} → ConfirmEmailChange.
   - `POST /login`: розший наявний — приймай `rememberMe` (cookie 12h/3дні); rate-limit уже в use-case.
2. **`GET /auth/me` DTO += `email_verified`** (N1: банер фронту на нього спирається).
3. **Error→HTTP мапінг** (явні коди, звір стиль `platform/httpserver` envelope): `email_taken`→409,
   `email_not_verified`→403, `weak_password`/валідація→422 (або як прийнято в репо), токен used/expired→410/400,
   `invalid_credentials`→401, rate-limited→429. Без секретів/токенів у логах; URL без query.
4. **openapi.yaml**: опиши всі нові ендпоінти, запити/відповіді, коди помилок, `AuthenticatedUser.emailVerified`.
   (Orval-генерацію Courses 1 прожене — ти лише правиш `openapi.yaml`; НЕ запускай npm/orval і НЕ чіпай
   `frontend/src/adapters/api/generated`.)
5. **Композиція**: підключи use-cases (NewAccounts + Mailer+RateLimiter+PasswordHasher+UserReader/Writer+TokenStore)
   у composition root (`cmd/api/main.go` / identity wiring), щоб маршрути були живі.
6. **Contract-тести** HTTP (патерн наявних handler_test.go): happy + ключові відмови кожного маршруту,
   нейтральність forgot, автологін-cookie на register, me.emailVerified.

## Правила
- Конституція: handler транслює транспорт і рендерить; бізнес-політики — в use-case (вже є). Валідація вводу
  на межі. Без fetch/SQL у handler. RequireSameOrigin на мутаціях.
- НЕ запускай sqlc/orval; openapi.yaml правиш, generated НЕ чіпаєш (Courses 1 згенерує). НЕ комість.
- TDD/contract-тести. Час вичерпується — спершу тести вже-написаного, познач залишок.

## Критерій готовності
- `go build ./...`, `go vet`, `gofmt` чисті.
- `go test ./internal/identity/...` зелені (нові contract-тести HTTP).
- `openapi.yaml` містить усі нові ендпоінти + emailVerified; валідний YAML.
- (Courses 1 потім: `make generate` для Orval → `check-generated`, повний гейт.)
