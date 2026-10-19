ДАЛІ: Передати A14 без коміту; Courses 1 має згенерувати клієнт після прийняття OpenAPI-зміни й прогнати повний gate.

2026-10-05 — A14 розпочато: прочитано CTO briefing, task brief і протокол implement. Робочий файл був відсутній, створено за вимогою брифу; дерево вже містило не пов’язані untracked work-файли, їх не чіпаю.

2026-10-05 — Оглянуто A13 application/ports/adapters та існуючі identity HTTP-патерни. Додано авторизовані 2FA маршрути, неавторизований complete, адмінське вимкнення, strict JSON-декодування, cookies лише після успішного complete, а також login-відповідь `two_factor_required` з `pendingId` без cookie. Додано error→HTTP для 2FA.

2026-10-05 — У composition root підключено Postgres identity repository як TwoFactor/Recovery/Password store, Redis pending store і наявний rate limiter. До PostgreSQL adapter додано вузький VerifyPassword запит за user ID (без generated/sqlc змін). OpenAPI вручну доповнено маршрутами, DTO, login oneOf та error codes; Orval/sqlc не запускались.

2026-10-05 — Додано HTTP contract-тест маршруту: enroll (otpauth), confirm/recovery codes, login 2FA challenge без cookie, complete з cookie, disable, regeneration і admin-disable/gate. `gofmt`, `go test ./internal/identity/...`, `go build ./...`, `go vet ./...` пройшли з GOCACHE=/private/tmp/courses-a14-gocache через sandbox-заборону системного Go cache.

2026-10-05 — Фінальний перегляд: `git diff --check` чистий; OpenAPI завантажується й HTTP OpenAPI contract-тести зелені. Під час огляду виправлено розміщення admin 2FA path, щоб він не потрапив усередину register response. Generated-файли не змінено, Orval/sqlc не запускалися, коміт не створювався.
