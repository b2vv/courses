ДАЛІ: Створити журнал A02, звірити стан гілки та релевантні migration/domain/seed патерни перед реалізацією.
ФАКТ: Стартова перевірка 2026-10-05: файл журналу, вказаний у brief, був відсутній; створено цей append-only журнал. Перший інструмент уже виявив відсутність, тож ця примітка зберігає точний порядок подій.
ДАЛІ: Оглянути git-стан, план accounts, наявні міграції, seed та identity domain; після огляду дописати висновок.
ФАКТ: Гілка `feat/accounts-selfservice`; дерево вже містило не пов’язані зміни скриншотів і численні робочі файли — їх не чіпаю. План §2/§8 підтверджує A02. Наявний `validDisplayName` має межу 120, тому реєстраційна політика буде окремою з межею 100. Seed має лише INSERT-колонки та conflict-update, обидва треба змінити.
РІШЕННЯ: Попри інструкцію скіла `implement` коміт не створюватиму: явна вимога брифу та користувача «НЕ комість» має пріоритет.
ДАЛІ: Прочитати повні патерни domain, errors, міграцій і тести, потім написати таблиці тестів політик до реалізації.
ФАКТ: Domain використовує конструктори `New…`, `Error`-коди й half-open expiry (`now.Before(expiresAt)`). `NormalizeEmail` вже є відповідною email-політикою. Міграції — goose Up/Down; `user_sessions.token_digest` має `bytea NOT NULL UNIQUE CHECK (octet_length(...) = 32)`.
РІШЕННЯ: Додам нові незалежні конструктори/валідації без зміни `validDisplayName`: `NewRegistration`, `NewAccountToken`, `ValidatePasswordChange`, `ValidateEmailChange`; для токена — три typed purpose, дві тривалості й `UsableAt`.
ДАЛІ: Додати table-driven domain-тести політик (вони мають спершу впасти), після цього реалізувати мінімальний чистий домен.
ФАКТ: Додано table-driven тести для registration (email/password/name межі), account token (purpose/інваріанти/consumed/expiry) і password/email change. Перший TDD-прогін не дійшов до компіляції: sandbox не дозволяє стандартний Go build cache у `~/Library/Caches/go-build`.
РІШЕННЯ: Наступні Go-перевірки запускатиму з `GOCACHE=/private/tmp/courses-a02-go-build`, не змінюючи код чи конфігурацію репозиторію.
ДАЛІ: Підтвердити очікуване compile-fail у ізольованому cache, реалізувати policy API та явні error-коди, потім повторити domain-тести.
ФАКТ: Ізольований TDD-прогін дав очікуваний compile-fail для ще неіснуючих Registration/AccountToken/policy API. Реалізацію додано в чистий `identity/domain`; `go test ./internal/identity/domain/...` зелений. `User` тепер несе `EmailVerified`.
ФАКТ: Додано `00014_accounts.sql` із goose Up/Down: три user-колонки, `account_tokens` з SHA-256-length digest check, unique digest і user/purpose index, та `recovery_codes` з user index. Seed задає `email_verified=true` в INSERT і `ON CONFLICT DO UPDATE`; TOTP лишено тільки колонками.
ДАЛІ: Доповнити реєстр стабільних доменних error-кодів, виконати форматування й цільові build/vet/test перевірки; окремо перевірити diff на межі A02 та generated-файли.
ФАКТ: `gofmt`, `git diff --check`, domain test, `go vet ./internal/identity/...` та `go build ./...` пройшли з ізольованим Go cache. Diff обмежено A02-файлами; generated-код не змінено. Додано DB-check: `new_email` дозволений лише для `email_change`, а expiry строго після створення.
ФАКТ: Незалежне spec-рев’ю знайшло один P2: тести покривали лише невалідні краї, але не прийнятні значення 8/128 для пароля та 1/100 для імені.
ДАЛІ: Додати позитивні крайові рядки table-driven тесту, повторити цільові перевірки й дочекатися standards-рев’ю.
ФАКТ: Додано позитивні граничні рядки: пароль рівно 8/128 символів та ім’я рівно 1/100. Після цього `gofmt`, `git diff --check`, `go test ./internal/identity/domain/...`, `go vet ./internal/identity/...` і `go build ./...` знову зелені; generated-файлів серед змін немає.
ФАКТ: Standards-рев’ю — без знахідок. Spec-рев’ю: єдиний P2 про позитивні межі закрито доданими тестами; решта відповідає A02, scope creep не знайдено.
РІШЕННЯ: Інтеграційне застосування 00014 до чистої БД та повні гейти не запускав: brief явно відводить це Courses 1. Коміт не створено за явною вимогою користувача.
ДАЛІ: A02 готовий для Courses 1: перевірити міграцію на чистій БД через integration/up, виконати повні гейти та зробити коміт окремим кроком.
