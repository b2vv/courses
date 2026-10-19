ДАЛІ: Оглянути A06-специфікацію, identity HTTP/application interfaces та наявні контрактні тести; реалізувати маршрути через існуючі use-case без дублювання політик.

2026-10-05 — Крок 1: Прочитано `work/CTO-RESEARCH.md`, `work/brief-codex-a06.md`, AGENTS.md та інструкцію implement. Підтверджено гілку `feat/accounts-selfservice`; дерево вже містить не пов’язані screenshots і WIP/brief-файли, їх не чіпаю. `work/_wip-a06-codex.md` був відсутній, створено append-only. Генератори sqlc/Orval не запускатиму й generated-файли не змінюватиму; коміт заборонений брифом.

2026-10-05 — Крок 2: Додано тонкі `/api/v1/auth/*` handlers над готовими `Accounts` use-case: register (201+HttpOnly session), verify/resend, forgot/reset/change password, email change/confirm; усі мутації за `RequireSameOrigin`, JSON строго обмежений 8 KiB, з відмовою duplicate/unknown/trailing полів. Login приймає `rememberMe`; `me` і login DTO несуть `emailVerified`. Додано явний error→HTTP мапінг для account-кодів і мову листа з `Accept-Language`.

2026-10-05 — Крок 3: У `cmd/api` зібрано `Accounts` з існуючих Postgres/Argon2/Redis/SMTP Mailer адаптерів. Додано adapter-read `FindUser` (ID або email) у Postgres repository, бо його порт уже вимагав A04, але runtime-реалізація була відсутня; generated-файли не змінювалися. OpenAPI доповнено всіма маршрутами/request schemas/error codes і обов’язковим `AuthenticatedUser.emailVerified`.

2026-10-05 — Крок 4: Додано HTTP contract coverage для register auto-login-cookie+emailVerified, forgot neutral 200, `email_taken`=409 та `token_used`=410. Прогнано `gofmt`, `go test ./internal/identity/...`, `go vet ./internal/identity/...`, `go build ./...`, `git diff --check` з ізольованим GOCACHE — зелено. Orval/sqlc не запускалися, `generated/` не торкано, коміт не створювався.

2026-10-05 — Крок 5: Додано contract assertion, що `rememberMe: true` доходить до `application.LoginInput`; повторний `go test ./internal/identity/adapters/http` і `git diff --check` зелені.

ДАЛІ: A06 готовий для перегляду Courses 1; наступний власник запускає генерацію Orval/sqlc лише за своїм процесом і повний repository gate.
