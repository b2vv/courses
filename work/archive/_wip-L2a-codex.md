# _wip L2a (Codex) — purchase-side membership edges

(Codex дописує тут після кожного кроку.)
ДАЛІ: звіряю чинні доменні тести, портові реалізації та settle-транзакцію перед TDD-зміною L2a.
ФАКТ: L1 вже має повний UNIQUE членства й стани self/blocked/revoked. Старе читання `ActiveGrants` втрачало blocked/revoked; старий settle робив лише DO NOTHING.
РІШЕННЯ: `CanStartOrder` приймає одне membership; тільки active+self дозволяє checkout. Settle відрізняє insert від upgrade за тим, чи RETURNING `id` дорівнює щойно згенерованому `$1`: upgrade повертає старий id і пише audit `upgrade`.
ДАЛІ: форматую зміни, запускаю доменні unit-тести та компілятор, потім виправляю лише L2a-помилки.
ФАКТ: `GOCACHE=/private/tmp/courses-l2a-gocache go test ./internal/orders/domain/...`, `go test ./internal/orders/...`, `go build ./...` і `go vet ./internal/orders/...` — green; `gofmt` і `git diff --check` чисті.
ФАКТ: Оновлено обидва фейки (`application/service_test.go`, `adapters/http/handler_test.go`): додано `Membership`, а `ActiveGrants` збережено для незачеплених викликів порту.
ФАКТ: Додані Docker integration-сценарії En02 (self-active upgrade in-place, `upgrade` audit, один рядок) та blocked/revoked/purchase-active (без мутації, `refund_required=true`). Docker-інтеграцію не запускав — це перевіряє Courses 1. Наявні Pay26 і settle-with-active-admin-grant не редагувались; їхній Docker-прогін лишається перевіркою Courses 1.
ДАЛІ: L2a завершено без коміту й без змін generated-файлів.
