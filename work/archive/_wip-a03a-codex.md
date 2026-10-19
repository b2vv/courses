ДАЛІ: додати та перевірити шов EmailVerified у HTTP Actor; SQL-генерацію не запускати.

2026-10-05 — Крок 1: прочитано `work/brief-codex-a03a.md`, `docs/accounts/plan.md` §2/§4b та наявні identity/Redis патерни. WIP-файл створено, бо його не було в дереві. Виявлено: `domain.User.EmailVerified` уже існує; `ResolveIdentitySession` і generated-мапер його ще не несуть. У A03a змінюється лише текст SQL та Actor; мапер лишається для A03b після `sqlc generate`.

2026-10-05 — Крок 2: додано `httpserver.Actor.EmailVerified`; `identity` Authenticate переносить факт із `domain.User`. Додано HTTP seam-тест, який перевіряє Actor, переданий наступному handler. Інших конструкторів `httpserver.Actor` у production-коді немає.

2026-10-05 — Крок 3: додано domain-safe інтерфейси `UserWriter`, `TokenStore`, `RateLimiter`, `Mailer` у `identity/ports`. Для TokenStore consume повертає лише факти, які повертає передбачений атомарний SQL (`userID`, `newEmail`). За TDD додано integration-тест публічного Redis seam (`Allow`/`Reset`) для fixed-window; реалізації ще немає, отже це очікуваний red-крок.

2026-10-05 — Крок 4: red-крок підтверджено (`NewRateLimiter` був відсутній; запуск із тимчасовим `GOCACHE`, бо системний cache закритий sandbox). Додано Redis adapter: `INCR`, `EXPIRE` лише на першому запиті fixed-window, порівняння з конфігурованим лімітом, `DEL` у Reset. Помилки Redis fail-open і структуровано логуються; `Allow`/`Reset` не переносять outage у login/mail потік.

2026-10-05 — Крок 5: у `sql/queries/identity.sql` додано текстові Create/verify/password/email/token/session-except-current запити; `ConsumeAccountToken` — атомарний `UPDATE … consumed_at=now() … RETURNING user_id,new_email`. `ResolveIdentitySession` тепер SELECT-ить `users.email_verified`. `sqlc generate` навмисно НЕ запускався й generated-файли не змінювалися. A03b має після генерації розширити `mapResolved`/спільний mapper полем generated row `EmailVerified`.

2026-10-05 — Крок 6: `gofmt` застосовано до всіх нових/змінених Go-файлів. Зелені: `go test ./internal/identity/adapters/http ./internal/identity/ports ./internal/identity/adapters/redis`, `go vet` для змінених пакетів, `go build ./...` (усі з тимчасовим writable `GOCACHE`). Не пройдено лише sandbox-обмежені інтеграційні перевірки: Redis test не має доступу до Docker socket; `platform/httpserver` test не має права bind loopback socket. `git diff --check` зелений; generated не змінені.

ДАЛІ: переглянути A03a diff, не змінюючи generated і не комітячи.

2026-10-05 — Крок 7: final review пройдено вручну для незакоміченого diff (формальний `code-review` вимагає базовий commit-to-HEAD diff, а A03a навмисно лишається uncommitted). Знахідок немає: портові типи не імпортують адаптерів; Redis state не глобальний; fail-open обмежено outage Redis; SQL consume має умову `consumed_at IS NULL` і expiry в тому ж UPDATE; `git diff --check` чистий. Підтверджено, що generated не змінені і sqlc не запускався.

ДАЛІ: A03a готова до передачі Courses 1; A03b має згенерувати sqlc та підключити `EmailVerified` у postgres mapper.
