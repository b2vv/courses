# Codex brief — A04b: reset / change-password / email-change + Login remember-me + rate-limit

Репо `courses`, гілка `feat/accounts-selfservice`. A04a у дереві: `application/accounts.go` (Register/VerifyEmail/
ResendVerification), `account_tokens.go` (token helper), ports (UserWriter/TokenStore/RateLimiter/Mailer/
PasswordHasher/UserReader), Mailer-адаптер. Наявний `application/identity.go` має `Login` (12h), `SessionStore`.
Роль: реалізуєш A04b; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a04b-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Use-cases (application, на портах+домені, TDD на фейках як у accounts_test.go)

1. **RequestPasswordReset(ctx, email, ip, lang)**: UserReader lookup. Якщо акаунт є — InvalidateByPurpose
   (password_reset) → issue reset-токен (lifetime 1h, `PasswordResetTokenLifetime`) → best-effort SendPasswordReset.
   **Відповідь ЗАВЖДИ нейтральна** (не розкриває існування, plan G4/§4b). Rate-limit per-email+IP (RateLimiter.Allow),
   ключ із поданого email **незалежно від існування** (E3) — лічильник не оракул.
2. **ResetPassword(ctx, rawToken, newPassword)**: digest → Consume(password_reset) → `domain` валідація пароля →
   Hash → UpdatePassword → **DeleteIdentitySessionsForUser (усі сесії)** (G8/AC5). used/expired → доменна помилка.
3. **ChangePassword(ctx, userID, old, new, currentDigest)**: перевір стару (PasswordHasher/CredentialVerifier) →
   `domain.ValidatePasswordChange(old,new)` → Hash → UpdatePassword → **DeleteUserSessionsExceptCurrent** (G8/AC6).
   Невірна стара → помилка, пароль не змінено.
4. **RequestEmailChange(ctx, userID, newEmail, lang)**: `domain.ValidateEmailChange` → **pre-check унікальності**
   (UserReader: новий email не зайнятий) → issue email_change-токен з `new_email` → SendEmailChange на **нову** адресу.
5. **ConfirmEmailChange(ctx, rawToken)**: Consume(email_change) → **повторна перевірка унікальності new_email**
   (TOCTOU, Q4/AC7: між request і confirm адресу могли зайняти) → UpdateEmail + позначити verified (нова адреса
   стає email_verified). Зайнятий → `email_taken`.
6. **Login remember-me**: розший наявний `Login` (identity.go) прапором `rememberMe` у вхідному типі:
   `false` → 12h (як зараз), `true` → 3 дні (`RememberMeLifetime` константа). Rate-limit login per-email+IP
   (~5 невдалих/15хв; успішний вхід → RateLimiter.Reset ключа). Константа тривалостей — у domain/application.
   УВАГА: двокроковий 2FA — це A13; тут Login лишається однокроковим, лише remember-me + rate-limit.

## Правила
- Конституція: лише домен+порти; жодного SQL/SMTP в application. Нейтральність forgot — суворо.
  Rate-limit fail-open уже в адаптері (A03a). Без секретів/токенів у логах.
- TDD на фейках (patterns у accounts_test.go). Онови наявні Login-тести під новий прапор, не ламаючи контракт
  (rememberMe=false має давати стару поведінку/тривалість).
- НЕ комість; НЕ чіпай generated; НЕ запускай sqlc.

## Критерій готовності
- `go build ./...`, `go vet ./internal/identity/...`, `gofmt` чисті.
- `go test ./internal/identity/application/...` зелені: forgot (нейтрально для існуючого й ні; rate-limit),
  reset (happy/used/expired + усі сесії геть), change-password (happy + невірна стара + крім поточної),
  email-change (happy + зайнятий на request + зайнятий на confirm/TOCTOU), login (remember-me 12h/3дні + rate-limit/reset).
