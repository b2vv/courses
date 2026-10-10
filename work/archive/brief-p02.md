# Brief P02 — три модулі переходять на контракт; AST-тест; захист від порожнього суб'єкта

Phase 9. Прочитай спершу: `docs/readiness/spec.md` (§3 D3, D4; §6 R04, R05),
`docs/adr/013-one-access-rule-as-a-contract.md`, `backend/internal/contracts/access/access.go` (P01),
`backend/platform/architecture/shared_rules_test.go` цілком — зокрема рядок
«draft, empty owner and empty actor: agreed, and unreachable by schema» і коментар над ним.

## Зроби

1. **Делегація (D3, R04).** Три методи стають мапінгом на контракт, без власної булевої логіки:
   - `backend/internal/course/domain/learning.go:80` — `LearningCourse.CanBeViewedBy`:
     `access.Subject{UserID: actor.UserID, Administrator: actor.Role == ActorAdmin}`,
     `access.Course{OwnerID: course.InstructorID, Published: course.Status == Published}`;
   - `backend/internal/playback/domain/authorization.go:33` — `VideoCourse.CanBeViewedBy`;
   - `backend/internal/transcript/domain/access.go:39` — `Course.CanBeViewedBy`.
   Тіло кожного — рівно один `return access.CanView(<мапінг>)`. Інлайн-мапінг полів в аргументах
   дозволений: це мапінг, а не правило. Порівняння статусу з «опубліковано» — частина мапінгу.
   Імена методів і типи доменів не змінюються; викликачі не змінюються.
2. **Захист від порожнього суб'єкта в контракті.** У `access.CanView` гілка власника спрацьовує лише
   за непорожнього `UserID`: `subject.UserID != "" && subject.UserID == course.OwnerID`. Коментар —
   одне речення: правило не довіряє транспорту наосліп, бо стане несучим для оплати. Додай рядок в
   `access_test.go`: порожній суб'єкт × неопублікований курс без власника → `false`.
3. **`shared_rules_test.go` (R05).** Таблиця лишається і проганяється через `CanBeViewedBy` кожного
   модуля. Тепер вона перевіряє **мапінг**, а не три логіки: онови коментар над тестом. Рядок
   «draft, empty owner and empty actor» переверни на `want: false` і перепиши коментар: раніше
   безпека жила лише в схемі, бо посилити одну копію означало розійтися з іншими; тепер правило
   одне, і захист у ньому (ADR 013). Решта рядків не змінюється.
4. **AST-тест R04** (`backend/platform/architecture/`, новий файл): знайди через `go/parser` усі
   методи з ім'ям `CanBeViewedBy` у `internal/**` (крім `_test.go`) і вимагай для кожного: тіло —
   рівно одна інструкція `return`, а її вираз — виклик `access.CanView` (селектор `CanView` на
   імпорті `courses/internal/contracts/access`). Інакше тест падає з назвою файлу. Доведи
   червоним: тимчасово поверни в одну копію інлайн-логіку `||`, переконайся, що тест падає,
   відкоти. Тест також вимагає, щоб таких методів було рівно три, — щоб четверта копія не
   з'явилась непоміченою. Опиши доказ у звіті.

## NOT

- Не чіпай SQL, адаптери, application, HTTP, фронтенд — це P03–P05.
- Не додавай у контракт полів для права на курс чи оплати.
- Git, `work/tasks.md`, `docs/` — ні.

## Якщо бачиш неточність

Якщо якась копія на таблиці дає вердикт, відмінний від контракту (крім рядка з порожнім суб'єктом,
який ти свідомо перевертаєш), це дефект, а не деталь рефакторингу: зупинись і опиши в
`## Discussion` з file:line.

## Перевірки

`gofmt`, `go vet ./...`, `go test ./...` (unit). Integration і e2e не запускаєш — їх жене
reviewing-сесія.

## Звіт

`work/report-p02.md`
