ДАЛІ: прочитати work/brief-b07.md цілком, усі перелічені в його §1 матеріали та work/codex-working-model.md, після чого звірити контракт B07 з поточним кодом.
ФАКТ: work/brief-b07.md і work/codex-working-model.md прочитані повністю. Бриф вимагає доменне звуження позиції лише за тривалістю обраного ready-ассета, без збереження тривалості в LearningLesson, без змін запису/позначки completed та без git; integration-тест має зафіксувати одночасно збережені 2050, ready-тривалість 480 і результат читання 480.
РІШЕННЯ: вимогу скіла implement про коміт відхилено на користь прямої заборони брифа; git не використовуватиму.
ДАЛІ: прочитати кожне джерело з §1, включно з доменною моделлю LearningLesson та наявними integration-фікстурами, і перевірити, чи наявний контракт дозволяє передати тривалість як незбережуваний вхід.
ФАКТ: первинне читання джерел підтвердило B07 на рядку 595, наявний LATERAL-вибір ассета й фікстуру одного integration-тесту; великий пакет виводу обрізано, а вказаний у брифі шлях LessonPlayer.tsx не збігся з деревом. Це ще не повне прочитання §1.
ДАЛІ: дочитати окремими вузькими запитами блок «Пʼять питань», повні цільові Go-файли, міграцію, AGENTS.md і handoff; знайти фактичний шлях LessonPlayer.tsx.
ФАКТ: рядок B07 вимагає integration-доказу «2050 збережено, ready-ассет 480, читання ≤480». Пʼять питань вимагають назвати точний доказ, production-походження артефакту, виключити випадково зелений результат, привʼязати доказ саме до B07 і сформулювати мутацію, що червонить тест. Код підтвердив: запис бере тільки ready-ассет і обрізає в playback/domain; читання пріоритезує ready, але за його відсутності обирає інший стан; LearningLesson тривалості не має; duration_seconds nullable. Фактичний шлях плеєра — frontend/src/presentation/learn/LessonPlayer.tsx.
РІШЕННЯ: контракт не бракує: доменний конструктор може прийняти окремий вхід із LearningLesson та необовʼязковою тривалістю, повертаючи незмінний LearningLesson. Реалізаційний шов треба уточнити після читання наявних доменних тестів і всіх викликів NewLearningCourse.
ДАЛІ: прочитати AGENTS.md і work/handoff-phase6.md окремо без обрізання, потім знайти всі виклики/тести NewLearningCourse та прочитати фактичний LessonPlayer.
ФАКТ: AGENTS.md і work/handoff-phase6.md прочитані повністю. Вони підтверджують внутрішнє розміщення бізнес-політики, integration-рівень B07, заборону оголошувати недоступний Docker-гейт пройденим та обовʼязок ревʼю-сесії перевірити й закомітити. Handoff окремо попереджає перевіряти середовище запуску самого доказу.
РІШЕННЯ: жодної неточності, що вимагала б `## Discussion`, у прочитаному контракті B07 не знайдено; заборонені паралельні шляхи лишаються поза слайсом.
ДАЛІ: прочитати обовʼязковий repository briefing work/CTO-RESEARCH.md, потім усі виклики/тести NewLearningCourse і плеєр.
ФАКТ: work/CTO-RESEARCH.md прочитано повністю. Він підтверджує B07 як невиконану роботу integration-рівня, модульні межі та недоступність Docker-гейтів у цій сесії; суперечностей із кодом, релевантних B07, не виявлено.
ДАЛІ: інвентаризувати всі виклики NewLearningCourse, наявні доменні тести навчальної моделі, точне відновлення позиції у LessonPlayer та форму API-мапінгу, щоб спроєктувати мінімальний незмінний зовнішній контракт.
ФАКТ: NewLearningCourse має три групи викликів: production PostgreSQL-адаптер і доменні/HTTP-тести. LessonPlayer відновлює незавершений урок безпосередньо з lesson.position; OpenAPI LearningLesson містить лише id/title/videoState/position/completed/clientSeq. Отже виправлення на серверному читанні не потребує змін API чи фронтенду.
РІШЕННЯ: обрано варіант 2 з брифа. Додам доменний конструктор LearningLesson, який приймає тривалість окремим незбережуваним аргументом і звужує лише коли VideoState == ready; LearningLesson та NewLearningCourse лишаться структурно й сигнатурно незмінними. Адаптер лише сканує сирі state/position/duration і передає їх домену. Це дешевше за спільний домен і не переносить політику в SQL.
ДАЛІ: прочитати інструкції tdd, повний learning_test.go і спосіб декодування nullable numeric у playback-адаптері; тоді написати спершу доменні failing-тести для ready/non-ready/NULL та integration-сценарій із явною перевіркою збережених 2050.
ФАКТ: tdd-інструкції прочитані; погоджені брифом шви — публічний доменний конструктор для правила та PostgreSQL repository integration для B07. Playback уже декодує nullable numeric через pgtype.Numeric.Float64Value(), що можна повторити як технічне мапування без політики. OpenAPI підтверджує відсутність duration у LearningLesson.
РІШЕННЯ: працюю вертикально: спочатку один доменний red/green для ready 2050→480; далі edge-випадки non-ready/NULL; після цього repository integration, який окремим SELECT перевіряє, що БД досі містить 2050, а GetLearningCourse віддає 480. Побічна SQL-перевірка тут потрібна не як реалізаційна деталь, а як явна передумова «застряглого стану» з приймального рядка.
ДАЛІ: перевірити, чи є CONTEXT.md та релевантні ADR-вказівки, після чого додати перший failing доменний тест без production-коду.
ФАКТ: CONTEXT.md у репозиторії не знайдено. ADR 011 і 012 прочитані; вони не змінюють правило B07. Першу спробу прочитати ADR 010 зроблено за неточним іменем файла; команда явно відмовила `No such file or directory`, обхід не застосовувався.
ДАЛІ: прочитати знайдений фактичний docs/adr/010-access-predicate-before-enrollment.md, потім почати red-крок.
ФАКТ: ADR 010 прочитано; він стосується видимості курсу й не змінює B07 або обраний доменний шов.
ДАЛІ: додати один доменний тест public-конструктора: ready-урок зі збереженою позицією 2050 і тривалістю 480 має будуватися з Position 480; production-код поки не змінювати.
ФАКТ: додано перший TDD-тест TestLearningLessonPositionUsesReadyAssetDuration; він фіксує незалежні літерали 2050, 480 і очікувану позицію 480 через публічний доменний шов.
ДАЛІ: запустити лише цей доменний тест і зафіксувати точний red до додавання конструктора.
ФАКТ: перший запуск тесту не дійшов до компіляції: sandbox відмовив у записі до /Users/strelia/Library/Caches/go-build (`operation not permitted`), exit 1. Це відмова середовища, не red продуктового тесту.
РІШЕННЯ: повторюю той самий тест із GOCACHE у дозволеному /tmp; код і обсяг тесту не змінюються.
ДАЛІ: отримати змістовний compile/test red з локальним кешем збірки.
ФАКТ: змістовний red отримано: `internal/course/domain/learning_test.go:72:22: undefined: coursedomain.NewLearningLesson`; `go test` exit 1. Тест червоний саме через відсутню доменну поведінку.
ДАЛІ: додати мінімальний доменний конструктор, який для ready і ненульової тривалості звужує 2050 до 480, та повторити лише цей тест.
ФАКТ: додано NewLearningLesson: він змінює лише Position і лише якщо стан ready, duration передано та Position перевищує його. Duration не зберігається. Коментарі в course/read і playback/write явно позначають свідоме дублювання одного обмеження на двох межах.
ДАЛІ: повторити точковий ready-тест (green), потім додати наступним циклом випадки non-ready і NULL.
ФАКТ: точковий ready-тест зелений: `ok courses/internal/course/domain`, exit 0. Числовий результат доменного правила — 480.
ДАЛІ: додати доменний table-test, що позиція 2050 не змінюється для відсутнього ассета, processing-ассета з duration 480 і ready-ассета з NULL duration.
ФАКТ: додано table-test трьох негативних умов. Він окремо захищає блокуючу вимогу: наявна duration=480 у processing-ассета не має звужувати позицію 2050.
ДАЛІ: запустити обидва доменні тести правила одним точковим прогоном.
ФАКТ: обидва доменні тести правила зелені (`go test ./internal/course/domain -run TestLearningLessonPosition`, exit 0): ready 2050→480; no asset, processing+480 і ready+NULL лишають 2050.
ДАЛІ: уточнити прийнятий у course/postgres спосіб мапити помилки nullable numeric, потім розширити SELECT/Scan сирою duration та викликати доменний конструктор без зміни LearningLesson.
ФАКТ: у course-модулі іншого pgtype.Numeric-мапінгу немає; playback використовує Float64Value і повертає dependency unavailable на помилці. Треба звірити локальний persistenceError перед реалізацією.
ДАЛІ: прочитати course/adapters/postgres/repository.go:89 і застосувати його чинну політику помилок.
ФАКТ: course persistenceError мапить непередбачені локальні помилки декодування в ErrDependencyUnavailable; це сумісно з playback-підходом.
ДАЛІ: перед production-правкою додати окрему B07 integration-фікстуру з трьома уроками: replacement ready 480 + stored 2050, processing 480 + stored 2050, ready NULL + stored 2050. B07-тест окремо прочитає lesson_progress і доведе, що 2050 не переписано.
ФАКТ: інвентаризація знайшла створення video_assets у backend/migrations/00005_media.sql; спроба відкрити 00004_media.sql явно відмовила через неправильне імʼя. Фікстуру ще не змінено.
ДАЛІ: прочитати точну схему 00005_media.sql, щоб integration-фікстура дотрималася всіх CHECK-обмежень.
ФАКТ: схема video_assets прочитана. Для старого ready-ассета використаю evicted_at після created_at; кожен ready матиме непорожній renditions, а processing — порожній. Унікальні generation і pending-індекс не конфліктують.
ДАЛІ: додати integration-тести й фікстури до learning_integration_test.go, не змінюючи production-код.
ФАКТ: integration-фікстуру додано окремим курсом. B07 має старий evicted ready-ассет duration 2050, новий ready generation 2 duration 480 і незавершений lesson_progress position 2050; тест читає БД до й після repository-read та вимагає 2050 в обох, а в моделі — рівно 480. Окремі integration-підтести вимагають 2050 для processing+480 і ready+NULL.
РІШЕННЯ: integration-тест навмисно не запущено в цій сесії: його середовище створює Testcontainers/Docker, що прямо заборонено §4. До production-правки очікувана помилка B07 — `short replacement` містить Position 2050 замість 480; це числове очікування буде передано рецензенту, але не видається за виміряний тут red.
ДАЛІ: змінити learning SELECT так, щоб LATERAL повертав raw duration_seconds, декодувати nullable numeric в адаптері й передати його NewLearningLesson; стан ready перевіряє лише домен.
ФАКТ: learning SELECT тепер повертає state і duration_seconds того самого current_video; адаптер сканує pgtype.Numeric, перетворює лише технічну nullable-форму та передає pointer у NewLearningLesson. Рішення «чи звужувати» лишилося в course/domain і залежить від уже спроєктованого VideoState == ready.
РІШЕННЯ: відсутній ассет і NULL duration обидва дають invalid pgtype.Numeric → nil; не-ready з ненульовою duration передається сирим, але домен її ігнорує. Це явно розділяє технічне мапування від політики.
ДАЛІ: застосувати gofmt до чотирьох змінених Go-файлів, потім прогнати точкові доменні тести та compile-only integration-пакет без запуску Docker-тесту.
ФАКТ: gofmt успішно застосовано (exit 0) до пʼяти, а не чотирьох, змінених Go-файлів: domain implementation/test, postgres implementation/integration-test і playback-коментар. Попередній підрахунок у `ДАЛІ` був опискою, не зміною обсягу.
ДАЛІ: повторити точкові доменні тести після форматування.
ФАКТ: точкові доменні тести після gofmt зелені (`ok courses/internal/course/domain`, exit 0).
ДАЛІ: скомпілювати integration-tagged course/postgres пакет із `-run '^$'`, не запускаючи Testcontainers; це перевіряє синтаксис і типи нового integration-тесту, але не є доказом B07.
ФАКТ: integration-tagged course/postgres пакет скомпільовано (`ok ... [no tests to run]`, exit 0). Жоден integration-тест не виконувався, Docker не запускався, B07 цим не доведено.
ДАЛІ: перечитати повні змінені секції SQL/Scan, доменного правила й integration-фікстури після gofmt; перевірити порядок колонок, NULL-поведінку, порядок уроків і те, що completed не змінюється.
ФАКТ: порядок SELECT і Scan збігається; current_video.duration_seconds належить тому самому LATERAL-рядку, що й state; completed лише копіюється; B07-уроки впорядковані позиціями 0/1/2. Виявлено лише випадкові таби всередині raw SQL-рядків і відсутність захисту від індексації при неочікуваній формі B07-курсу — це погіршує діагностику тесту, хоч runtime SQL не змінює.
РІШЕННЯ: виправити формат raw SQL і додати перевірку `1 module × 3 lessons` перед індексацією в кожному B07-підтесті; продуктову логіку не змінювати.
ДАЛІ: внести ці тестові/читабельні правки та знову gofmt.
ФАКТ: raw SELECT вирівняно; кожен із трьох нових integration-підтестів тепер перевіряє форму `1 module × 3 lessons` до індексації. Після цього gofmt для двох зачеплених файлів завершився exit 0.
ДАЛІ: виконати mutation proof: тимчасово прибрати звуження з NewLearningLesson, запустити ready-тест і зафіксувати змістовне падіння 2050 проти 480; потім негайно відновити production-правило.
ФАКТ: для mutation proof тимчасово прибрано тіло звуження, сигнатуру й тест лишено незмінними.
ДАЛІ: запустити лише TestLearningLessonPositionUsesReadyAssetDuration і зняти точний текст падіння.
ФАКТ: mutation proof червоний, exit 1: `learning_test.go:75: position=2050 want 480`. Отже тест відрізняє дефектну поведінку від виправленої числом, а не зеленіє випадково.
ДАЛІ: негайно відновити доменне звуження лише для ready+known duration і повторити той самий тест до green.
ФАКТ: production-правило відновлено в точній попередній формі: VideoReady && duration != nil && Position > duration.
ДАЛІ: повторити mutation-tested ready-тест після відновлення.
ФАКТ: той самий ready-тест після відновлення зелений (`ok courses/internal/course/domain`, exit 0). Red 2050→green 480 виміряно.
ДАЛІ: перевірити контрактну пастку: LearningLesson/OpenAPI/generated не отримали duration; зафіксувати часові мітки OpenAPI/generated та пошуком підтвердити відсутність поля.
ФАКТ: пошук durationSeconds/duration_seconds/DurationSeconds у course LearningLesson, backend/openapi/openapi.yaml і frontend/src/adapters/api/generated знайшов лише параметр доменного конструктора; у схемі та generated-моделі поля немає. generated LearningLesson має mtime 2026-09-08, тобто ця сесія його не змінювала; openapi має mtime 2026-09-12 17:45:53 і потребує порівняння з початком цієї сесії.
ДАЛІ: порівняти birth/mtime робочого звіту з mtime openapi, щоб документально підтвердити, що openapi не змінювався після старту B07.
ФАКТ: report-b07.md створено 2026-09-12 23:38:13; openapi.yaml востаннє змінено 17:45:53, generated LearningLesson — 2026-09-08 12:20:33. Отже обидва передують цій сесії й B07 їх не змінював; зовнішній контракт незмінний.
ДАЛІ: запустити дозволений брифом пакетний тест `go test ./internal/course/... ./internal/playback/...` з локальним GOCACHE у /tmp.
ФАКТ: `go test ./internal/course/... ./internal/playback/...` завершився exit 0; усі course/playback unit-пакети зелені. Integration-tagged TestLearningReaderAgainstPostgreSQL сюди не входить і не видається за пройдений.
ДАЛІ: запустити `go vet ./internal/course/... ./internal/playback/...` з тим самим дозволеним кешем.
ФАКТ: `go vet ./internal/course/... ./internal/playback/...` завершився exit 0 без діагностик.
ДАЛІ: перевірити gofmt через `gofmt -d` на всіх пʼяти змінених Go-файлах; порожній diff буде фінальним форматним доказом.
ФАКТ: `gofmt -d` завершився exit 0 з порожнім виводом; усі пʼять Go-файлів відформатовані.
ДАЛІ: повторно скомпілювати integration-tagged course/postgres пакет після фінальних тестових правок, з `-run '^$'` і без запуску Docker.
ФАКТ: фінальна compile-only перевірка integration-tagged course/postgres пакета зелена (`ok ... [no tests to run]`, exit 0). Вона перевіряє компіляцію, не поведінку PostgreSQL.
ДАЛІ: прочитати code-review skill і провести саморевʼю явних пʼяти файлів проти брифа, архітектурних стандартів і acceptance-умов без використання git.
ФАКТ: code-review skill прочитано. Його стандартний fixed-point/git процес конфліктує з абсолютною забороною §4, тому git-команди й commit-list не виконуються; специфіка явно задана work/brief-b07.md, стандарти — AGENTS.md і work/codex-working-model.md.
РІШЕННЯ: виконати дві обовʼязкові осі ревʼю паралельними read-only агентами за явним переліком змінених файлів; їм заборонено git, редагування й доступ до паралельно зайнятих шляхів. Відсутність fixed point названа, а не прихована.
ДАЛІ: отримати незалежні висновки Standards і Spec; потім виправити лише підтверджені знахідки в межах B07.
ФАКТ: незалежний Standards-review не знайшов порушень чи суттєвих smell; свідоме дублювання read/write bound виправдане брифом і звʼязане коментарями. Незалежний Spec-review не знайшов missing/partial, scope creep або wrong implementation: ready-only умова, незбережуваний duration, незмінний completed, stored-before/after=2050 і read=480 відповідають контракту.
РІШЕННЯ: підтверджених ревʼю-знахідок для виправлення немає; `## Discussion` не відкривається, бо неточності, що призводила б до баґа, не виявлено.
ДАЛІ: провести фінальну інвентаризацію торкнутих шляхів без git, явно перевірити, що заборонені файли не новіші за старт сесії, і дописати перелік запущених/незапущених гейтів та числове очікування рецензенту.
ФАКТ: торкнуто лише `backend/internal/course/domain/learning.go`, `backend/internal/course/domain/learning_test.go`, `backend/internal/course/adapters/postgres/learning.go`, `backend/internal/course/adapters/postgres/learning_integration_test.go`, `backend/internal/playback/domain/progress.go` (лише коментар) і цей робочий звіт. Git-команд не виконувалось; `infrastructure/`, `backend/platform/architecture/`, `docs/playback/report.md`, `work/tasks.md`, `backend/openapi/openapi.yaml` та `frontend/src/adapters/api/generated/` не редагувались.
ФАКТ: поведінка трьох відсутніх тривалостей: (1) ассета немає — duration=nil, позиція без змін; (2) обраний ассет не ready, навіть із duration=480 — домен ігнорує межу, позиція 2050; (3) ready-ассет має NULL duration — duration=nil, позиція 2050. Лише ready+known duration звужує 2050 до 480.
ФАКТ: запущено: доменний red/green і mutation proof; `go test ./internal/course/... ./internal/playback/...` exit 0; `go vet ./internal/course/... ./internal/playback/...` exit 0; `gofmt -w` і фінальний `gofmt -d` без diff; двічі compile-only `go test -tags=integration ./internal/course/adapters/postgres -run '^$'`, exit 0, без виконаних тестів. Перший тестовий запуск був відхилений sandbox через системний Go cache; повтори з `GOCACHE=/tmp/courses-b07-go-cache` пройшли. Review Standards/Spec — 0 знахідок.
ФАКТ: не запущено й не оголошено пройденими: сам PostgreSQL/Testcontainers integration-тест і `make integration` (немає дозволеного Docker); `make build`, `make check-generated`, `make up`, `seed`, `smoke`, `make e2e`, повний `make check`; також не запускався ordered make-гейт (`make format-check`, `make lint`, `make typecheck`, `make test` як його окремі сходинки), бо §4 віддає повну перевірку реценз-сесії.
РІШЕННЯ: B07 реалізовано, але не позначається доведеним на integration-рівні в цій сесії. Числове очікування рецензенту: без правки repository-read повертає 2050; з правкою — 480; `lesson_progress.position_seconds` до і після читання лишається 2050. Mutation proof фактично впав з `position=2050 want 480`.
ПИТАННЯ: немає. Контракту вистачило; підстав для `## Discussion` не виявлено.
ДАЛІ: реценз-сесії запустити справжній integration-тест/упорядкований gate, підтвердити B07 на PostgreSQL і лише після цього виконати git-коміт та оновлення фазових артефактів у своїй зоні відповідальності.
