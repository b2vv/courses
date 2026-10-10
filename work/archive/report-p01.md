# P01 — звіт

Дата: 2026-09-25

## Зроблено

- Додано `backend/internal/contracts/access` з `Subject`, `Course` та чистим предикатом `CanView`.
- Додано unit-тест із 9 випадками `{адмін, власник, сторонній} × {чернетка, опубліковано, архів}`. Для чернетки й архіву в тесті явно використано `Published: false`; усі суб'єкти автентифіковані.
- Оновлено `backend/platform/architecture/boundaries_test.go`: домен може імпортувати лише `courses/internal/contracts/…` додатково до стандартної бібліотеки. Правило для application не змінювалося.
- Додано architecture-тест, який перевіряє, що production-файли під `internal/contracts/…` імпортують лише стандартну бібліотеку.
- Три модулі, `shared_rules_test.go`, фронтенд, міграції, OpenAPI, `work/tasks.md` і `docs/` не змінювалися цим завданням.

## Перевірки

- `gofmt -l` для змінених Go-файлів: пройшов, вивід порожній.
- `go vet ./...`: пройшов, exit 0.
- `go test ./internal/contracts/... ./platform/architecture/...`: пройшов; обидва пакети green.
- R02 red-доказ: тимчасовий blank-import `courses/internal/course/domain` у production-файл контракту спричинив падіння `TestContractsImportOnlyTheStandardLibrary` з повідомленням про non-standard package. Імпорт одразу прибрано; фінальний тест знову green.

Integration, e2e та ordered gate не запускалися: вони не входять до перевірок P01 і потребують середовища поза цим робочим циклом.

## R01–R03

- R01: виконано — 9/9 truth-table випадків проходять через `access.CanView`.
- R02: виконано — architecture guard на імпорти контрактів додано й перевірено green/red/green.
- R03: виконано — domain boundary дозволяє `courses/internal/contracts/…`; application rule лишився без змін.

## Scope

Три наявні копії правила не переводилися на контракт; це P02.

Стан Git не змінювався: commit, branch, reset та інші write-операції з Git не виконувалися.
