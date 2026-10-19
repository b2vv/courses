ДАЛІ: A13a готовий до передачі без коміту; наступний зріз A13b має реалізувати PostgreSQL-адаптери й composition wiring.

2026-10-05 — Прочитано AGENTS.md, CTO briefing і `work/brief-codex-a13a.md`; підтверджено гілку `feat/accounts-selfservice`. Межі: без generated/sqlc/Postgres-адаптера/коміту. Наявний `PasswordHasher` має лише `Hash`; для Disable потрібен окремий вузький verify-порт.

2026-10-05 — TDD: додано початкові тести Confirm/Complete/Admin-disable та порти й application `TwoFactor`; перший запуск червоний тільки через дефект тестового RFC 6238 генератора (`invalid_totp`), реалізаційний контракт ще не оцінювався.

2026-10-05 — Виправлено тестовий HMAC-вектор, додано SQL-текст усіх семи запитів і Redis `PendingTwoFactorStore` з integration-тестом. Наступний тест знайшов реальний дефект: спільний валідатор сесії жорстко вимагав 12h, тому Complete для remember-me повертав `dependency_unavailable`; замінено на lifetime-обізнаний валідатор.

2026-10-05 — Звичайні `go test ./internal/identity/application/... ./internal/identity/adapters/redis/...` зелені. Integration Redis-тести не можуть стартувати testcontainer: sandbox відмовляє Docker socket (`permission denied`), і так само падає наявний `TestRateLimiterFixedWindowAgainstRedis`; це не позначено як pass. Додано TDD-кейси Disable, Regenerate та limiter; невдала перевірка повертає pending лише для наступної обмеженої спроби, тоді як Redis `Take` лишається атомарним delete.

2026-10-05 — Додано recovery-path до Complete та виправлено Redis decode-помилку для порожнього user ID. `go test ./internal/identity/application/... ./internal/identity/adapters/redis/...`, `go build ./...` і `go vet ./...` зелені (з `GOCACHE=/private/tmp/courses-go-cache`, бо стандартний cache sandbox-недоступний). Помилковий `make format-check` з `backend/` не є перевіркою: ціль є лише в корені репо.

2026-10-05 — Кореневий `make format-check`: frontend `oxfmt --check` зелений; Go `gofmt -l` не запущено, бо Makefile запускає Docker і sandbox відмовляє socket. Це той самий зовнішній блокер, що й integration Redis. Власний `gofmt` на всіх A13a Go-файлах уже виконано.

2026-10-05 — Додано TDD-кейси Enroll і позитивного Admin-disable. Фінальний повтор: application/Redis unit tests, `go build ./...`, `go vet ./...` — зелені; `gofmt -d` для всіх змінених Go-файлів дав 0 bytes, `git diff --check` чистий. Формальний code-review workflow не запускався: він потребує фіксованої git-точки й diff `base...HEAD`, а A13a навмисно лишено незафіксованим; натомість виконано локальну spec/diff перевірку.
