# _wip L3a (Codex) — membership predicates in contracts/access

(Codex дописує тут після кожного кроку.)

ДАЛІ: додати спершу table-driven юніт-матрицю L3a для CanConsume/CanCommunicate, потім мінімальну реалізацію лише в contracts/access.
2026-10-06 — Прочитано CTO briefing, `brief-codex-L3a.md` і enrollment plan §2. Межа підтверджена: лише `backend/internal/contracts/access` та його юніти; `CanView`, `CanStream`, `PurchaseRequired`, TruthTable і читачі не змінюються. Робоче дерево вже брудне сторонніми файлами; їх не чіпаю.
2026-10-06 — TDD: додано table-driven матрицю 4 membership-стани × stranger/owner/admin × preview false/true; вона перевіряє обидва нові предикати, включно з blocked, revoked, none+preview і stranger+active.
2026-10-06 — Перший `go test ./internal/contracts/access/...` не стартував через sandbox-доступ до стандартного Go build cache (`~/Library/Caches/go-build`); це інфраструктурна, не тестова помилка. Наступний прогін використовує ізольований кеш у `/private/tmp`.
2026-10-06 — TDD red підтверджено з ізольованим GOCACHE: нові тип, поле й предикати ще не існували. Додано тільки в `contracts/access`: Membership (`none|active|blocked|revoked`), адитивне поле `Course.Membership`, чисті `CanConsume` і `CanCommunicate`; наявні access-предикати не редаговано.
2026-10-06 — Застосовано `gofmt` до рівно двох файлів L3a: `access.go` та `access_test.go`.
2026-10-06 — `GOCACHE=/private/tmp/courses-l3a-gocache go test ./internal/contracts/access/...` PASS: `courses/internal/contracts/access` green; `accesstest` без власних тестів.
2026-10-06 — `GOCACHE=/private/tmp/courses-l3a-gocache go build ./...` PASS.
2026-10-06 — `GOCACHE=/private/tmp/courses-l3a-gocache go test ./platform/architecture/...` PASS. `rg` підтвердив рівно 4 декларації `CanBeViewedBy` (course, playback, transcript, discussion); TruthTable і всі наявні `CanView`/`CanStream`/`PurchaseRequired` рядки не чіпалися.
2026-10-06 — Переглянуто L3a-диф відносно HEAD: тільки `access.go`, `access_test.go` і цей append-only журнал; scope creep, generated-файлів, читачів, модулів, SQL і гейтингу немає. `git diff --check` чистий. L3a готовий; коміт не створювався за брифом.
