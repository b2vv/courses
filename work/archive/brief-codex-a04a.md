# Codex brief — A04a: Mailer-адаптер + шаблони + Register/VerifyEmail/ResendVerification

Репо `courses`, гілка `feat/accounts-selfservice`. A01–A03 у дереві: `platform/mail.Sender`, порти
`identity/ports/accounts.go` (UserWriter/TokenStore/RateLimiter/Mailer), postgres/redis адаптери, домен
`account_policy.go` (NewRegistration, AccountToken, lifetimes), `argon2id.Verifier.Hash`, `application.Clock`,
`SessionStore`. Роль: реалізуєш A04a; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a04a-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити

1. **Mailer-адаптер** `backend/internal/identity/adapters/mail/`: реалізує `ports.Mailer` поверх
   `platform/mail.Sender`. Рендерить листи uk/en через `text/template`/`html/template` (verification,
   password-reset, email-change). Мова — параметр (`lang`), дефолт uk. Лінк будує з публічного origin
   (звір, звідки брати — `identity/adapters/http/config.go` `publicOrigin`/`PUBLIC_ORIGIN`). Без секретів
   у логах/лінках у логах. Юніт-тести рендеру (обидві мови, екранування).

2. **Token-хелпер** (application або окремий пакет): генерує випадковий токен (`crypto/rand`, достатня
   ентропія), повертає (сирий токен для листа, sha-256 digest 32 байти для `TokenStore`). Expiry =
   `Clock()+lifetime` із доменних констант (`EmailVerificationTokenLifetime` 24h). Чистий, юніт-тест.

3. **Use-cases** у `backend/internal/identity/application/` (на портах+домені, TDD на фейках):
   - **Register(ctx, input, lang)**: `domain.NewRegistration` → `Hash` пароль → `UserWriter.CreateUser`
     (помилка `email_taken` з 23505 піднімається як є) → видати verification-токен (TokenStore.Issue) →
     **best-effort** `Mailer.SendVerification` (§4b: збій надсилання НЕ відкочує акаунт; поверни
     `verificationEmailSent:false`, не помилку) → **автологін**: `SessionStore.Rotate` видає сесію →
     поверни користувача+токен сесії+прапор. Rate-limit (signup) — ключем per-email+IP через RateLimiter
     (виклик додамо повноцінно в A04b разом з login; тут хоча б передбач точку).
   - **VerifyEmail(ctx, rawToken)**: digest → `TokenStore.Consume(email_verify)` → `UserWriter.SetEmailVerified`.
     Протермінований/вжитий → доменна помилка (як у TokenStore).
   - **ResendVerification(ctx, userID|email, lang)**: `TokenStore.InvalidateByPurpose(email_verify)` →
     видати новий → best-effort send. Якщо вже verified — нічого не слати (ідемпотентно/нейтрально).

## Правила
- Конституція: application залежить лише від домену й портів; Mailer/TokenStore/UserWriter — порти.
  Жодного прямого SQL/SMTP в application. Без секретів у логах; токен у лозі НЕ друкувати.
- TDD на фейках портів (є патерни у наявних `application/*_test.go`). Mailer-рендер — юніт.
- НЕ комість; НЕ чіпай generated; НЕ запускай sqlc. Час вичерпується — спершу тести вже-написаного, познач залишок.

## Критерій готовності A04a
- `go build ./...`, `go vet ./internal/identity/...`, `gofmt` чисті.
- `go test ./internal/identity/application/... ./internal/identity/adapters/mail/...` зелені (Register happy +
  email_taken + mail-fail-not-fatal; Verify happy+expired+used; Resend invalidates+idempotent; template render uk/en).
