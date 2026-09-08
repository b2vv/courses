ДАЛІ: Прочитати work/brief-t2.md, обов’язкові інструкції репозиторію та виконати перший визначений крок.
ФАКТ: Прочитано `work/brief-t2.md`, `work/CTO-RESEARCH.md` і скіл `implement`. Межі T2: лише `00006_playback`, атомарне з публікацією збереження тривалості та інтеграційні канарки; Docker-гейти не запускати; `backend/internal/playback/`, `docs/playback/*`, `work/tasks.md`, `generated/` і git не чіпати.
РІШЕННЯ: Дотримуватися окремого стоп-каналу для відсутнього контракту та додавати `## Discussion` лише для доведеної неточності, що спричинить баг.
ДАЛІ: Прочитати точні секції плану, стандарти, міграційні й транзакційні seams та наявні патерни канарок.
ФАКТ: Прочитано план, стандарти, поточні задачі, конвенції міграцій, код probe/AcceptSource, видалення уроків і знайдено наявні інтеграційні патерни B07/B08. `deleteMissingAuthoringLessons` справді видаляє лише відсутні у надісланому дереві уроки.
ФАКТ: Потрібний код публікації винесений у `backend/internal/media/adapters/postgres/publish.go`; попередній сукупний вивід був обрізаний, тому підпис `Publish` і точні D-5/B-рядки ще потребують окремого читання до будь-якої зміни.
ДАЛІ: Окремо прочитати D-5, §5/§6/§8, `Publish` порт та реалізацію, транскод-use-case і повні міграційні тести.
ФАКТ: Рецензент підтвердив стоп і дозволив розширити `AssetRepository.Publish` явно опційною, вже нормалізованою тривалістю. Додатковий інваріант: помилка publish-policy мусить відкотити тривалість разом з усією публікацією.
ФАКТ: Рецензент виявив пастку канарки: `repository_integration_test.go` зашиває міграції до `00005_media.sql`, а синтетична `00006_failure.sql` зіткнеться зі справжньою `00006_playback.sql`.
РІШЕННЯ: Додати доменну нормалізацію поруч із `SourceProbe.DurationSeconds`, передавати опційне значення через порт, персистити його в транзакційному `UPDATE`, оновити всі фейки; до MapFS додати `00006_playback.sql`, синтетичну поломку перейменувати на `00007_failure.sql`.
ДАЛІ: Прочитати TDD-інструкцію й точні невитягнуті секції плану, потім спочатку додати тести на нормалізацію, передачу та rollback.
ФАКТ: Прочитано TDD-інструкцію та точні D-5/§5/§6/§8. Узгоджені рецензентом seams: публічна доменна нормалізація `SourceProbe` → optional duration; application передає нормалізоване значення; repository `Publish` атомарно зберігає duration; міграційна канарка перевіряє заповнену БД і rollback.
РІШЕННЯ: Працювати вертикальними red→green циклами. Не реалізовувати T3-правила завершення/обрізання: B06/B41/B44 тут лише готуються схемою й даними.
ДАЛІ: Дочитати точні тестові harness-и й перевірити доменний `CONTEXT.md`; потім написати перший failing unit test нормалізації.
ФАКТ: `CONTEXT.md` у media-модулі відсутній. Наявні seams підтверджені кодом: `processingRepository` моделює application-рівень, `testPublishEvictsOldest` уже доводить rollback Phase 5 через справжній repository, а `mediaAssetColumns` навмисно не містить тривалості в доменній `VideoAsset`.
РІШЕННЯ: Не розширювати `VideoAsset`: тривалість потрібна як аргумент команди публікації й майбутнім playback/read SQL, але не поточним media-відповідям. Доменний API буде нормалізувати `SourceProbe.DurationSeconds` у `*float64`.
ДАЛІ: Додати failing unit test для `duration <= 0 → nil`, `duration > 0 → value`, запустити лише пакет domain і зафіксувати RED.
ФАКТ: `source_test.go` тестує публічний доменний API в package `domain_test`; новий тест природно додається поруч із `TestAcceptSource` без доступу до внутрішніх деталей.
ДАЛІ: Додати `TestSourceProbePersistedDuration` до `source_test.go` й запустити пакет domain для RED.
ФАКТ: Додано unit test `TestSourceProbePersistedDuration`: відʼємна й нульова тривалість очікуються відсутніми, додатна 61.25 — присутньою без зміни значення.
ДАЛІ: Запустити `go test ./internal/media/domain` і підтвердити RED через відсутній доменний метод.
ФАКТ: `go test ./internal/media/domain` не дійшов до компіляції: системний `/Users/strelia/Library/Caches/go-build` заборонений (`operation not permitted`). Це обмеження середовища, не результат тесту.
РІШЕННЯ: Наступні Go-команди запускати з `GOCACHE=/tmp/courses-go-build-cache`, як дозволяє бриф; у фінальному звіті назвати цей override.
ДАЛІ: Повторити той самий unit test із дозволеним GOCACHE й отримати справжній RED.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./internal/media/domain` — RED, exit 1: `SourceProbe.PersistedDurationSeconds undefined`. Тест чутливий до відсутньої поведінки.
ДАЛІ: Реалізувати мінімальний доменний метод `PersistedDurationSeconds`, що повертає nil для `<= 0` і вказівник на значення для `> 0`, потім повторити unit test.
ФАКТ: У `media/domain/source.go` додано `SourceProbe.PersistedDurationSeconds() *float64`; правило `<= 0 → nil` тепер ухвалюється до перетину repository-порту.
ДАЛІ: Повторити `go test ./internal/media/domain` із sandbox GOCACHE й підтвердити GREEN.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./internal/media/domain` — GREEN, exit 0.
ДАЛІ: Знайти найвужчий успішний transcode-сценарій і додати до нього доказ, що application передає в `Publish` нормалізовану додатну/нульову тривалість.
ФАКТ: `TestProcessClaimsRendersAndPublishes` є найвужчим успішним application-сценарієм; його shared `fakeRepository` визначено в `upload_test.go`, а default probe зараз має нульову тривалість.
РІШЕННЯ: Розширити тестовий fake новим optional duration аргументом, зафіксувати отримане значення і в наявному успішному сценарії очікувати 125.5 секунди. Окреме доменне unit-покриття вже доводить `0 → nil`; database B30 буде доведено integration-тестом.
ДАЛІ: Змінити лише тестовий контракт/очікування і запустити application package для RED перед production-кодом.
ФАКТ: Тестовий `fakeRepository.Publish` тепер очікує `*float64`; `TestProcessClaimsRendersAndPublishes` вимагає, щоб probe 125.5 дійшов до publish без втрати.
ДАЛІ: Запустити `go test ./internal/media/application` із sandbox GOCACHE та підтвердити RED на старому production-контракті/виклику.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./internal/media/application` — RED, exit 1: fake більше не реалізує старий `AssetRepository.Publish`; compiler прямо показує відсутній optional duration у порту.
ДАЛІ: Розширити порт `Publish` аргументом `*float64` і передати `probe.PersistedDurationSeconds()` з application-виклику; повторити application test.
ФАКТ: `AssetRepository.Publish` розширено параметром `durationSeconds *float64`; `Transcode.Process` передає `probe.PersistedDurationSeconds()` у той самий publish-виклик.
ДАЛІ: Повторити `go test ./internal/media/application` із sandbox GOCACHE й підтвердити GREEN application seam.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./internal/media/application` — GREEN, exit 0; application seam передає 125.5 як optional duration.
ДАЛІ: Додати integration-тест repository seam для трьох випадків: zero-normalized → SQL NULL; policy error → duration не лишається; успіх → positive numeric збережено. Оновити наявні integration-виклики новим аргументом і отримати compile RED до зміни адаптера.
ФАКТ: Додано integration-сценарій `B30 publish persists only usable duration atomically`: нуль після доменної нормалізації очікує SQL NULL; помилка policy очікує rollback до NULL; успіх очікує numeric 125.5. Наявні integration-виклики `Publish` отримали явне `nil`.
ДАЛІ: Скомпілювати integration-tagged postgres package без запуску тестів командою `go test -run '^$' -tags=integration ./internal/media/adapters/postgres` і підтвердити RED на старому адаптері.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test -run '^$' -tags=integration ./internal/media/adapters/postgres` — RED, exit 1: PostgreSQL `Repository.Publish` має старий підпис і не реалізує розширений порт; нові тестові виклики мають зайвий duration аргумент.
ДАЛІ: Додати optional duration до реалізації `Repository.Publish` і до її транзакційного `UPDATE`, не додаючи duration до `VideoAsset`/read model; повторити compile-only integration check.
ФАКТ: PostgreSQL `Repository.Publish` приймає `*float64` і записує `duration_seconds` у тому самому guarded `UPDATE`, що переводить асет у `ready`; commit лишився єдиним і відбувається після policy та eviction logic.
ДАЛІ: Повторити compile-only integration check; виправити лише залишкові контрактні місця, якщо compiler їх назве.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test -run '^$' -tags=integration ./internal/media/adapters/postgres` — GREEN compile-only, exit 0. Docker/testcontainers не стартували через `-run '^$'`; runtime integration ще не запускався.
ДАЛІ: Звірити в spec/plan точні SQL типи й nullability `lesson_progress`, перевірити конвенцію `lock_timeout`, потім додати міграцію та B07/B08 канарку.
ФАКТ: У репо немає наявної SQL-конвенції `lock_timeout`; plan/brief вимагають явний timeout, але не фіксують число. Попередня plan-критика називала це, однак фінальний plan закрито з вимогою лише явного bounded timeout і операційним відновленням.
РІШЕННЯ: Використати `SET LOCAL lock_timeout = '5s'` першим SQL statement у goose Up: він обмежує блокування обох кроків однієї транзакції й не змінює session після неї.
ДАЛІ: Точково перевірити spec на типи/nullability; якщо додаткового контракту немає, взяти мінімальну схему з природними типами (`uuid`, `numeric`, nullable `completed_at`, `bigint`, `timestamptz`) без спекулятивних індексів/constraints.
ФАКТ: `docs/playback/spec.md` не задає додаткових SQL типів/nullability; plan фіксує лише склад полів, composite PK, `client_seq bigint`, два cascade FK і `duration_seconds numeric NULL`.
РІШЕННЯ: Мінімальна схема: `user_id uuid`, `lesson_id uuid`, `position_seconds numeric NOT NULL`, `completed_at timestamptz NULL`, `client_seq bigint NOT NULL`, `updated_at timestamptz NOT NULL`; без defaults, CHECK-ів та додаткових індексів.
ДАЛІ: Прочитати точний контракт таблиці сесій і seed-дані, щоб populated-database canary містила всі названі класи рядків та порівнювала дані без хибного diff через нову nullable колонку.
ФАКТ: Сесійна таблиця зветься `user_sessions` і вимагає UUID, FK user, 32-byte digest та `expires_at > created_at`. `backend/sql/seed.sql` заповнює лише courses; populated canary має явно створити module, lesson, session і video_asset.
ДАЛІ: Створити `backend/migrations/00006_playback.sql` з timeout першим, мінімальною `lesson_progress`, однією nullable колонкою та test-only Down.
ФАКТ: Створено `00006_playback.sql`: `SET LOCAL lock_timeout='5s'` стоїть першим statement; `lesson_progress` має composite PK і два cascade FK без додаткового індексу; `video_assets` отримує лише nullable numeric duration; Down видаляє таблицю й колонку та позначений test-only.
ДАЛІ: Додати populated-database канарку з courses/lessons/video_assets/user_sessions, перевіркою структури й snapshot-ів; включити real `00006` у failing MapFS і перенумерувати synthetic failure на `00007`.
ФАКТ: Додано canary `playback migration preserves populated database`: перед UpTo(6) існують users/courses/modules/lessons/user_sessions/video_assets; після міграції перевіряються snapshots, `lesson_progress`, nullable numeric duration, bigint client_seq, два cascade FK та рівно один PK-індекс.
ФАКТ: Failing-migration MapFS тепер включає справжню `00006_playback.sql`, а синтетична помилка має номер `00007_failure.sql`, тому канарка більше не тестує застарілий набір і не має collision.
ДАЛІ: Скомпілювати course postgres integration package без запуску Docker-тесту; потім оновити всі залишкові test fakes/Publish calls за compiler/rg.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test -run '^$' -tags=integration ./internal/course/adapters/postgres` — GREEN compile-only, exit 0; Docker/testcontainers не запускались.
ДАЛІ: Перерахувати всі `AssetRepository.Publish` реалізації та виклики, оновити лише залишкові test fakes/calls, потім запустити всі недокерні Go-тести.
ФАКТ: Повний `rg` знайшов рівно одну неоновлену реалізацію `AssetRepository.Publish`: fake у `media/adapters/http/handler_test.go`. Решта збігів — storage-порт з іншим контрактом або вже оновлені repository-виклики.
ДАЛІ: Оновити HTTP fake optional duration аргументом, відформатувати змінені Go-файли й запустити недокерні Go-тести.
ФАКТ: Оновлено останній HTTP test fake; усі реалізації repository `Publish` тепер мають однаковий optional duration аргумент.
ДАЛІ: Запустити `gofmt` для змінених Go-файлів, потім перевірити формат diff-independent командою `make format-check` у належному порядку гейтів.
ФАКТ: `gofmt -w internal/media/domain/source.go internal/media/domain/source_test.go internal/media/ports/repository.go internal/media/application/transcode.go internal/media/application/transcode_test.go internal/media/application/upload_test.go internal/media/adapters/postgres/publish.go internal/media/adapters/postgres/repository_integration_test.go internal/media/adapters/http/handler_test.go internal/course/adapters/postgres/repository_integration_test.go` — exit 0.
ДАЛІ: Запустити перший дозволений ordered gate `make format-check`.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache make format-check` — exit 2: frontend `oxfmt --check .` пройшов (140 файлів), backend-підкрок не запустився через відсутній доступ до Docker socket. Gate загалом НЕ ПРОЙШОВ/не завершився; не представляти як pass.
ДАЛІ: Запустити наступний дозволений gate `make lint`; якщо його Docker-підкрок так само заблокований, зафіксувати окремо без обходу Docker integration.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache make lint` — exit 2: frontend oxlint + architecture boundaries пройшли (78 файлів), backend `go vet ./...` не запустився через Docker socket. Gate загалом НЕ ПРОЙШОВ/не завершився.
ДАЛІ: Запустити `make typecheck` як третій дозволений ordered gate.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache make typecheck` — exit 0; frontend TypeScript typecheck пройшов.
ДАЛІ: Запустити всі Go-тести без integration build tag командою `GOCACHE=/tmp/courses-go-build-cache go test ./...` з `backend/`.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./...` — exit 1: усі перелічені пакети пройшли, крім `courses/platform/httpserver`; `TestShutdownDrainsAnInFlightResponse` не зміг bind `127.0.0.1:0` (`operation not permitted`). Це sandbox-відмова, не assertion продукту; повний non-integration suite НЕ ПРОЙШОВ.
ДАЛІ: Повторити non-integration Go-тести явним набором пакетів без єдиного sandbox-заблокованого `platform/httpserver`, щоб довести всі доступні пакети; окремо не називати повний suite зеленим.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./cmd/... ./internal/... ./platform/architecture ./platform/config ./platform/observability ./platform/postgres ./platform/redis ./platform/runtime` — exit 0; усі доступні non-integration Go packages пройшли, `platform/httpserver` навмисно виключений лише через зафіксовану sandbox bind-відмову.
ДАЛІ: Виконати обовʼязкове self-review за `code-review`, не використовуючи git через NOT-список; перевірити змінені файли напряму, контрактні місця, SQL/канарки й потенційні runtime-помилки.
ФАКТ: Прочитано `code-review` skill. Він вимагає окремі Standards і Spec огляди паралельно; звичайний git fixed-point непридатний, бо brief прямо забороняє торкатися git.
РІШЕННЯ: Передати двом review-агентам явний список змінених файлів і джерела (`AGENTS.md`, `.claude/standards.md`, `work/brief-t2.md`, reviewer decision) без git-команд; результати звести окремо й виправити blocking знахідки.
ДАЛІ: Запустити паралельні Standards/Spec огляди відомого T2 change-set.
ФАКТ: Запущено два паралельні read-only огляди: Standards (`AGENTS.md` + `.claude/standards.md` + smell baseline) і Spec (`work/brief-t2.md` + reviewer decision + точні plan-рядки). Обом заборонено git і редагування.
ДАЛІ: Поки огляди працюють, локально перевірити pgx-кодування optional float64 у PostgreSQL numeric та цілісність міграційних тестів; потім звести findings.
ФАКТ: pgx/v5.10 `NumericCodec` підтримує `Float64Valuer`, а builtin wrappers перетворюють `float64` у цей контракт; optional `*float64` коректно кодується як numeric або SQL NULL без adapter-side нормалізації.

## Discussion

ФАКТ: `work/brief-t2.md` §2.4 називає рядки plan §8 **B07/B08 канаркою**, але `docs/playback/plan.md:489-490` визначає B07 як обрізання позиції після заміни на коротше відео, а B08 — збереження `completed_at` після заміни. Міграційна канарка є лише прозою §5 (`plan.md:389`) і не має власних B-ID.
ПИТАННЯ: Якщо T2 позначить B07/B08 закритими migration-canary тестом, T5/T9 можуть не реалізувати runtime-поведінку заміни: позиція лишиться понад нову тривалість або позначка completed буде втрачена/перерахована. Пропонована правка: дати migration populated/rollback canaries окремі acceptance IDs (або O3 subrows), а B07/B08 лишити власникам progress/read-model.
РІШЕННЯ: Документи не редагував за NOT-списком. Канарки реалізував як окремі докази вимоги §5/brief §2.3; у фінальній таблиці не видаватиму їх за доказ runtime B07/B08.
ДАЛІ: Посилити failing-migration canary перевіркою, що успішна реальна `00006` лишилася застосованою після rollback синтетичної `00007`; потім дочекатися двох review-звітів.
ФАКТ: Failing-migration canary тепер після відмови `00007_failure.sql` перевіряє не лише відсутність частково створеної `migration_canary`, а й збереження успішної `lesson_progress` та `video_assets.duration_seconds` із `00006`.
ДАЛІ: Відформатувати змінений integration test і повторити compile-only course integration check.
ФАКТ: `gofmt -w internal/course/adapters/postgres/repository_integration_test.go && GOCACHE=/tmp/courses-go-build-cache go test -run '^$' -tags=integration ./internal/course/adapters/postgres` — exit 0; integration package компілюється, runtime Docker-тест не запускався.
ДАЛІ: Отримати Standards і Spec review findings та виправити всі підтверджені проблеми в межах T2.
ФАКТ: Стоп-умова брифу підтвердилась. `backend/internal/media/ports/repository.go:30-31` передає в `Publish` лише `id`, `token`, `renditions`, `now` і policy; тривалості в контракті немає. `backend/internal/media/application/transcode.go:120` має `probe.DurationSeconds`, але викликати з ним порт не може. `backend/internal/media/adapters/postgres/publish.go:68-72` в атомарному `UPDATE video_assets` записує стан, claim, failure reason, renditions і updated_at — не duration. Отже вимога «тривалість тією ж транзакцією» потребує зміни підпису порту; окремий запис після `Publish` порушив би атомарність.
РІШЕННЯ: Зупинено роботу до зміни/уточнення контракту, як прямо вимагає §5 брифу. Жодних змін у міграції, production-коді чи тестах не зроблено.
ПИТАННЯ: Підтвердіть розширення `AssetRepository.Publish` параметром тривалості (з нормалізацією `duration <= 0` до `NULL` у application/adapter seam) або надайте інший дозволений контракт для атомарного запису.
ФАКТ: Хронологічна поправка: попередні три рядки були первинним стоп-звітом, який фізично лишився в кінці файлу через вставки за anchor; рецензент після нього явно дозволив розширення `Publish`, і реалізація, міграція та тести вже виконані згідно з записами вище. Стоп більше не активний.
ФАКТ: Standards review знайшов hard violation: `PersistedDurationSeconds` заносить repository/persistence термін у чистий domain. Spec review знайшов partial O3: немає canary, що реально тримає конфліктний lock, отримує timeout/rollback, звільняє lock і успішно повторює migration. Scope creep і порушень NOT-списку не знайдено.
РІШЕННЯ: Перейменувати domain API на `UsableDurationSeconds` із предметним коментарем; додати integration lock-timeout/recovery canary. Повторення snapshot loops лишити як non-blocking judgement call: виділення helper розширило б рефакторинг поза необхідною T2 поведінкою.
ДАЛІ: Перейменувати доменний API та всі його використання, відформатувати й повторити domain/application compile-tests.
ФАКТ: `PersistedDurationSeconds` перейменовано на предметний `UsableDurationSeconds`; коментар більше не згадує repository, port чи persistence. Оновлено unit/application/integration використання.
ДАЛІ: Відформатувати чотири змінені Go-файли й окремо запустити domain/application tests.
ФАКТ: `gofmt -w internal/media/domain/source.go internal/media/domain/source_test.go internal/media/application/transcode.go internal/media/adapters/postgres/repository_integration_test.go` — exit 0.
ДАЛІ: Запустити `go test ./internal/media/domain ./internal/media/application` із sandbox GOCACHE.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./internal/media/domain ./internal/media/application` — exit 0; domain boundary rename не змінив поведінку.
ДАЛІ: Додати integration canary, що тримає `ROW EXCLUSIVE` lock на `video_assets`, очікує явний lock-timeout і повний rollback `00006`, звільняє lock та повторно успішно застосовує міграцію.
ФАКТ: Додано lock-contention canary: real `00006` запускається під `ROW EXCLUSIVE` lock, мусить повернути помилку з `lock timeout`, не лишити ані таблицю, ані колонку; після rollback блокувальної транзакції той самий `UpTo(6)` мусить застосувати обидва кроки.
ДАЛІ: Відформатувати course integration test і compile-check його з integration tag без runtime Docker.
ФАКТ: `gofmt -w internal/course/adapters/postgres/repository_integration_test.go` — exit 0.
ДАЛІ: Запустити `go test -run '^$' -tags=integration ./internal/course/adapters/postgres` із sandbox GOCACHE.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test -run '^$' -tags=integration ./internal/course/adapters/postgres` — exit 0; lock-timeout canary компілюється, але не запускався без Docker.
ДАЛІ: Повторити фінальний ordered verification після review-fixes: `make format-check`, `make lint`, `make typecheck`, доступні Go-тести; окремо compile-check обох змінених integration packages.
ФАКТ: Фінальний `GOCACHE=/tmp/courses-go-build-cache make format-check` — exit 2: frontend format check знову пройшов; backend Docker/gofmt підкрок знову не запустився через Docker socket. Gate НЕ ПРОЙШОВ/не завершився.
ДАЛІ: Запустити фінальний `make lint`.
ФАКТ: Фінальний `GOCACHE=/tmp/courses-go-build-cache make lint` — exit 2: frontend lint і architecture check знову пройшли; backend Docker/go-vet підкрок не запустився через Docker socket. Gate НЕ ПРОЙШОВ/не завершився.
ДАЛІ: Запустити фінальний `make typecheck`.
ФАКТ: Фінальний `GOCACHE=/tmp/courses-go-build-cache make typecheck` — exit 0.
ДАЛІ: Запустити фінальний доступний non-integration Go suite без sandbox-заблокованого `platform/httpserver`.
ФАКТ: Фінальний `GOCACHE=/tmp/courses-go-build-cache go test ./cmd/... ./internal/... ./platform/architecture ./platform/config ./platform/observability ./platform/postgres ./platform/redis ./platform/runtime` — exit 0.
ДАЛІ: Compile-check обох змінених integration packages з `-run '^$'`, не запускаючи Docker/runtime tests.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test -run '^$' -tags=integration ./internal/course/adapters/postgres` і аналогічна команда для `./internal/media/adapters/postgres` — обидві exit 0 `[no tests to run]`; це лише compile-check, не runtime pass.
ДАЛІ: Звірити Makefile-таргети для точного переліку незапущених Docker-гейтів, після чого дописати фінальні canary-команди й таблицю доказів §8.
ФАКТ: Makefile підтверджує: backend частини `format-check`, `lint`, `test` запускаються через Docker; `integration`, `build`, `up`, `migrate` також Docker/Compose; `e2e` потребує піднятого Compose. `check-generated` ще й торкається git, що заборонено NOT-списком.
ДАЛІ: Як додатковий, нееквівалентний Docker-гейтам доказ запустити host `gofmt -l .` і `go vet ./...`; результати маркувати саме як host checks, не як green `make format-check`/`make lint`.
ФАКТ: Host `gofmt -l .` — exit 0 без виводу; host `GOCACHE=/tmp/courses-go-build-cache go vet ./...` — exit 0. Це додаткові перевірки, не заміна Docker-підкроків Makefile.
ФАКТ: Повторний review після виправлень: hard domain-boundary finding закрито перейменуванням на `UsableDurationSeconds`; missing lock-timeout/recovery scenario закрито integration canary. Non-blocking дублювання test snapshot loops свідомо не рефакторилось поза T2.

## Команди рецензента для runtime-канарок

Запускати з `backend/` у середовищі з Docker:

```sh
GOCACHE=/tmp/courses-go-build-cache go test -count=1 -tags=integration -run='^TestRepositoryAgainstPostgreSQL$/^(playback migration preserves populated database|playback migration lock timeout rolls back and retry succeeds|failed migration rolls back its complete transaction)$' ./internal/course/adapters/postgres
GOCACHE=/tmp/courses-go-build-cache go test -count=1 -tags=integration -run='^TestMediaRepositoryAgainstPostgreSQL$/^B30 publish persists only usable duration atomically$' ./internal/media/adapters/postgres
```

Перша команда доводить populated migration, реальний lock-timeout + rollback + retry та rollback наступної хибної `00007`. Друга доводить SQL NULL для нуля, rollback duration при publish-policy error і positive duration на успіху.

## Гейти

| Перевірка | Результат |
|---|---|
| `make format-check` | НЕ ЗАВЕРШИВСЯ: frontend pass; backend Docker-крок не запущено через socket permission |
| `make lint` | НЕ ЗАВЕРШИВСЯ: frontend lint/architecture pass; backend Docker `go vet` не запущено |
| `make typecheck` | PASS, exit 0 |
| доступні host Go tests без `platform/httpserver` | PASS, exit 0 |
| повний host `go test ./...` | НЕ ПРОЙШОВ: лише `platform/httpserver` bind `127.0.0.1:0` заборонений sandbox |
| integration-tagged course/media packages | лише compile-check PASS з `-run '^$'`; runtime НЕ ЗАПУСКАВСЯ |
| host `gofmt -l .`, host `go vet ./...` | PASS, exit 0; не замінюють pinned Docker gates |

Не запускались через обмеження брифу/середовища: `make test` (Docker backend), `make integration`, жива міграція через `make migrate`, `make build`, `make up`, `make e2e`; також `make check-generated` не запускався, бо T2 не чіпає generated і NOT-список забороняє git.

## Рядок §8 → чим доводиться

| Рядок | Доказ у T2 | Статус |
|---|---|---|
| B06 | nullable `video_assets.duration_seconds`; `UsableDurationSeconds`; integration zero→NULL | підготовлено; правило `completed_at`/необрізання реалізує T3 |
| B07 | migration populated-data canary є окремим доказом §5, але plan B07 — обрізання після заміни на коротше | НЕ закрито T2; див. Discussion |
| B08 | failing-migration rollback canary є окремим доказом §5, але plan B08 — збереження completed-позначки після заміни | НЕ закрито T2; див. Discussion |
| B30 | `TestSourceProbeUsableDuration`; `TestProcessClaimsRendersAndPublishes`; `B30 publish persists only usable duration atomically` | реалізовано; runtime integration чекає рецензента |
| B41 | duration schema/data seam | лише підготовлено; правило зростання позиції — T3/T5 |
| B44 | duration schema/data seam | лише підготовлено; обрізання позиції — T3/T5 |

ФАКТ: T2 завершено в дозволеному обсязі: міграція, атомарний duration write, нормалізація нуля, populated/rollback/lock-timeout canaries та test-fake contract updates. `backend/internal/playback/`, `docs/playback/*`, `work/tasks.md`, `generated/` і git не змінювались.
ДАЛІ: Рецензент запускає дві runtime integration-команди вище та повний Docker-порядок гейтів у дозволеному середовищі.

---

## Верифікація рецензентом (2026-09-08)

### Гейти

**Повний `make check` — чистий з першого прогону**, без розбиття на частини:
334 фронт-тести, `integration` без флаків, e2e **35/35**, жодного `make: ***`.

### Канарки — доведені поіменно, не «пакет зелений»

`go test -v -run TestRepositoryAgainstPostgreSQL ./internal/course/adapters/postgres`:

```
=== RUN   .../playback_migration_preserves_populated_database          -> B45
=== RUN   .../playback_migration_lock_timeout_rolls_back_and_retry_succeeds -> B47 (O3)
=== RUN   .../failed_migration_rolls_back_its_complete_transaction     -> B46
ok  courses/internal/course/adapters/postgres  8.888s
```

Перевірено саме поіменно: зелений пакет **не доводить**, що новий підтест виконався, а не
відсутній. Це та сама дисципліна, якої фаза навчила на кеші й на скрипті заміру.

### Discussion — прийнято, і дефект був мій

Бриф §2.4 називав **B07/B08** канарками міграції. Це нумерація **Phase 5**; у плані Phase 6
B07 — обрізання позиції після заміни на коротше відео, B08 — збереження позначки
(`plan.md:489-490`). Якби T2 записав їх закритими, T5 і T9 могли б не реалізувати поведінку
заміни взагалі.

Виконавець відмовився видавати канарку за їх доказ. Канарки отримали **власні
номери B45–B47**, а O3 перепризначено з «таск міграції» на **T2, рядок B47 — закрито**.

### Пастка канарки — віддана заздалегідь і посилена виконавцем

Рецензент знайшов до рев'ю, що `repository_integration_test.go` тримає список міграцій,
зашитий поіменно до `00005`, і синтетичну поломку під номером `00006_failure.sql` — тобто
поява справжньої `00006` зробила б канарку сліпою до неї **й** дала б колізію імен.
Передано резюмом разом із відповіддю на стоп.

Виконавець не лише перейменував поломку на `00007_failure.sql`, а **додав перевірку, якої не
просили**: після відмови синтетичної міграції доводиться, що успішна `00006` лишилась
застосованою — `lesson_progress` і `duration_seconds` на місці.

### Стоп

Спрацював правильно і вдруге за фазу зберіг шов. `AssetRepository.Publish` не приймав
тривалості; окремий запис після нього порушив би атомарність. Виконавець показав три рядки
доказу й зупинився замість того, щоб зробити «працююче». Контракт розширено з дозволу
рецензента; нормалізація `duration <= 0 → NULL` живе **в домені**
(`media/domain/source.go: UsableDurationSeconds`), а не в адаптері.

**Власне рев'ю виконавця** окремо знайшло, що перша назва методу — `PersistedDurationSeconds` —
тягла репозиторний термін у чистий домен, і перейменувала його. Знахідка не з брифу.
