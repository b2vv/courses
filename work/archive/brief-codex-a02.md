# Codex brief — A02: міграція 00014 + seed + доменні політики акаунтів

Репо `courses`, гілка `feat/accounts-selfservice`. План — `docs/accounts/plan.md` (§2, §5), spec — AC1, AC12, AC20.
Роль: ти реалізуєш A02; коміт і гейти — Courses 1; рев'ю — Courses 2.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a02-codex.md
Append після КОЖНОГО кроку. Перший запис — ДО першого інструменту: рядок `ДАЛІ:`. Далі ФАКТ/РІШЕННЯ/ДАЛІ/ПИТАННЯ.
Фінал = стислий переказ файлу.

## Що зробити

### 1. Міграція `backend/migrations/00014_accounts.sql` (up+down, як наявні 00013 etc.)
UP:
- `ALTER TABLE users ADD COLUMN email_verified boolean NOT NULL DEFAULT false`,
  `ADD COLUMN totp_secret text NULL`, `ADD COLUMN totp_enabled boolean NOT NULL DEFAULT false`.
- `CREATE TABLE account_tokens`: `id uuid PK`, `user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`,
  `purpose text NOT NULL CHECK (purpose IN ('email_verify','password_reset','email_change'))`,
  `token_digest bytea NOT NULL` (sha-256 digest, як сесії зберігають digest — звір `user_sessions`),
  `new_email text NULL` (лише для email_change), `expires_at timestamptz NOT NULL`,
  `consumed_at timestamptz NULL`, `created_at timestamptz NOT NULL DEFAULT now()`.
  Індекси: унікальний на `token_digest`; на `(user_id, purpose)`.
- `CREATE TABLE recovery_codes`: `id uuid PK`, `user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE`,
  `code_hash text NOT NULL`, `consumed_at timestamptz NULL`, `created_at timestamptz NOT NULL DEFAULT now()`;
  індекс на `(user_id)`.
DOWN: дзеркально — `DROP TABLE recovery_codes; DROP TABLE account_tokens;
  ALTER TABLE users DROP COLUMN totp_enabled, DROP COLUMN totp_secret, DROP COLUMN email_verified;`
Дотримуйся наявного стилю міграцій (коментарі, назви констрейнтів, canary якщо є патерн).

### 2. Seed — `backend/sql/seed-identities.sql`
Додай `email_verified` **у ОБИДВІ частини**: у список колонок INSERT зі значенням `true` для всіх 6 рядків,
**і** у `ON CONFLICT … DO UPDATE SET` (щоб повторний сід теж ставив true). Без цього наявні e2e (verified-gated
контент) впадуть. `totp_enabled` лишається false за дефолтом (нічого не додавай).

### 3. Домен `backend/internal/identity/domain` (чистий, без I/O, без зовнішніх пакетів)
- `User.EmailVerified bool` — додай поле у domain `User` (воно нестиме гейт у пізніших тасках).
- Політики (чисті функції/типи з таблицями юніт-тестів **спершу**, TDD):
  - **Registration**: валідація email (формат), пароль **8–128** символів, `display_name` **1–100** (УВАГА:
    наявний `validDisplayName` допускає ≤120 — нова реєстрація ріже саме на 100; не чіпай наявну функцію,
    зроби політику реєстрації окремо або параметризуй). Явні доменні помилки (`weak_password`,
    невалідний email/ім'я) у стилі наявних `domain/errors.go`.
  - **AccountToken**: доменне представлення одноразового токена (вид purpose, строк, перевірка
    `не протермінований і не спожитий`). Строки з плану: verify 24 год, reset 1 год — визнач як доменні
    константи/параметри (саме обчислення expiry за `Clock` лишиться в application; домен дає політику).
  - **ChangePassword**: нова≠стара; потребує валідну стару (перевірку робить application через verifier;
    домен дає правило порівняння й політику пароля 8–128).
  - **ChangeEmail**: валідація нового email; доменна політика (унікальність/верифікація — в application/adapter).
- **TOTP/2FA-логіку НЕ роби** — це A12. Лише колонки в міграції (вище) вже є; домен TwoFactor — пізніше.

## Правила
- Конституція: домен — тільки stdlib + `internal/contracts/*`; жодного framework/DB/transport. Юніт-тести
  на межах домену. Генерацію sqlc НЕ чіпай (нові запити — у A03); A02 не має змінювати `*/generated/*`.
- TDD: тести → реалізація. Якщо впираєшся в ліміт часу — спершу дороби тести вже-написаного; познач залишок.
- НЕ комість — лиши зміни в робочому дереві для Courses 1. Не чіпай незв'язані файли/скриншоти.

## Критерій готовності
- `go build ./...`, `go vet ./internal/identity/...`, `gofmt` — чисті.
- `go test ./internal/identity/domain/...` — зелені (нові політики покриті таблицями).
- Міграція 00014 up+down синтаксично валідна; застосовується на чистій БД (перевірить Courses 1 через
  `make integration`/`make up`).
- `check-generated` не має червоніти (A02 не торкається generated).
