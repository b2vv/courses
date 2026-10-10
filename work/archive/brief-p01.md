# Brief P01 — контракт доступу `internal/contracts/access` і межі

Phase 9. Прочитай спершу: `docs/readiness/spec.md` (§3 D1, D2, D4 — перший пункт; §6 R01–R03),
`docs/adr/013-one-access-rule-as-a-contract.md`, `AGENTS.md` §Architecture,
`backend/platform/architecture/boundaries_test.go`, `backend/platform/architecture/shared_rules_test.go`
і три наявні копії правила: `backend/internal/course/domain/learning.go:80`,
`backend/internal/playback/domain/authorization.go:33`, `backend/internal/transcript/domain/access.go:39`.

## Зроби

1. **Пакет `backend/internal/contracts/access`** (один файл коду плюс тест):
   - `Subject` — `UserID string`, `Administrator bool`;
   - `Course` — `OwnerID string`, `Published bool`;
   - `func CanView(subject Subject, course Course) bool` — рівно правило трьох копій:
     адмін, або власник, або опубліковано. Жодного поля чи клаузи для права на курс чи оплати
     (ADR 013: без споживача не додаємо);
   - короткий коментар пакета: це контракт між модулями, лише чисті предикати, модулі мапляться в
     нього; за подробицями — ADR 013.
2. **Unit-тест R01** (`access_test.go`): таблиця з 9 випадків
   `{адмін, власник, сторонній} × {чернетка, опубліковано, архів}`. Контракт знає лише `Published`,
   тож «чернетка» й «архів» обидва — `Published: false`: назви випадки так, щоб це було видно.
   Суб'єкти автентифіковані; «сторонній» — не анонім. Очікування: адмін — так у всіх трьох;
   власник — так у всіх трьох; сторонній — так лише для «опубліковано».
3. **R03 — межа домену.** `boundaries_test.go`: домен може імпортувати, крім стандартної бібліотеки,
   **лише** пакети з префіксом `courses/internal/contracts/`. Правило для application не змінюй.
4. **R02 — межа контрактів.** Новий architecture-тест поряд: будь-який `.go` (не `_test.go`) у
   `internal/contracts/…` імпортує лише стандартну бібліотеку. Жодного `courses/…`, жодного
   зовнішнього пакета. Доведи червоним: тимчасово додай заборонений імпорт, переконайся, що тест
   падає, прибери. Опиши це в звіті.

## NOT

- Не переводь три модулі на контракт — це P02. Копії правила лишаються як є.
- Не чіпай `shared_rules_test.go` — це P02.
- Не чіпай фронтенд, міграції, `openapi.yaml`, `work/tasks.md`, `docs/`. Git не чіпай.

## Перевірки

`gofmt`, `go vet ./...`, `go test ./internal/contracts/... ./platform/architecture/...`. Результат
кожної — чесно.

## Якщо бачиш неточність, що дасть баг

`## Discussion` у звіті: що не так, file:line, що станеться, пропозиція.

## Звіт

`work/report-p01.md`
