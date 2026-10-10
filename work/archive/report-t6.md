ДАЛІ: прочитати work/brief-t6.md, обов’язковий briefing репозиторію та інструкції skill implement перед виконанням T6.

- Журнал був відсутній; перша спроба append завершилася без змін. Файл створено для подальшої покрокової фіксації.

ФАКТ: прочитано `work/CTO-RESEARCH.md`, `work/brief-t6.md` і skill `implement`. T6 додає автентифіковану читальну модель навчання, інтеграційні докази B01/B27/R17/R18/R19/R21/R34 та один спільний набір випадків видимості для `course` і `playback`.

РІШЕННЯ: виконую TDD на наявних seams; завершальне рев’ю зроблю через skill `code-review`. Вимогу skill зробити commit не виконую, бо NOT-список T6 прямо забороняє торкатися git.

ДАЛІ: прочитати всі вказані в T6 контракти, план фази, стандарти, поточні задачі та звіт T5; перевірити наявність повного контракту до змін.

ФАКТ: первинний огляд підтвердив чинний публічний контракт `getPublishedCourse`: лише назви модулів/уроків, без id. `playback.Authorizer` застосовує `course.CanBeViewedBy(actor)` після перевірки ownership. Стандарти фіксують inline-plan поріг, але T6 уже має затверджені spec/plan і лежить нижче `spec-flow` за самим брифом.

ФАКТ: посилання брифу `sql/queries/media.sql:41-45` не існує від кореня репо; фактичне розташування треба знайти до висновку про прецедент. Вивід першого широкого читання був обрізаний, тому це ще не повне прочитання контрактів.

ДАЛІ: дочитати точні секції plan/spec без обрізання, знайти фактичний SQL-прецедент і актуальний рядок T6 у трекері.

ФАКТ: точні D-6, §1, §7, A18–A21 і потрібні рядки §8 прочитано. `work/tasks.md` має T6=`planned`, T1–T5=`done`. Фактичний SQL-прецедент — `backend/sql/queries/media.sql`; помилка шляху в brief/plan не приховує контракт і не змінює реалізацію.

РІШЕННЯ: T6 не змінює `getPublishedCourse`; новий endpoint належить `course` і читає `lesson_progress`/video state прямим SQL join. Видимість мусить залишитись двома production-реалізаціями, але одна таблиця тест-кейсів перевірятиме обидві, включно з archived owner (R34).

ДАЛІ: інвентаризувати домен/порти/HTTP/SQL/composition і schema, визначити найменший seam та точне правило `continueLessonId`; окремо перевірити потенційний четвертий випадок, про який попереджає brief.

ФАКТ: `course` уже має `CourseRepository`/`Catalog` і title-only DTO; `playback` має власні `Actor`, `CourseAccess` та `CanBeViewedBy`. SQL схеми містять потрібні `lesson_progress` і `video_assets`; окремої міграції T6 не потрібно. `backend/sql/queries/courses.sql` — sqlc-вхід, але generated output заборонено змінювати вручну.

ФАКТ: продовження не можна визначати лише як «перший незавершений»: brief окремо вимагає стан відео, а урок без відео/з processing/failed посеред курсу створює четвертий випадок — такий урок не може бути точкою запуску плеєра. Це неточність, яка може повести T7/T8 на непрогравний урок.

ДАЛІ: записати `## Discussion` з доказами та запропонованим правилом, потім дочитати точні auth/composition/test seams і перевірити, чи контракт дозволяє безпечно продовжити.

## Discussion — `continueLessonId` не визначений для непослідовного прогресу й стану відео

ФАКТ: `work/brief-t6.md:41-42` називає лише «немає прогресу / всі завершені / частина завершена», а `work/brief-t6.md:108-111` прямо визнає невизначені четверті випадки. `docs/playback/spec.md:57-59` обіцяє повернення туди, де користувач зупинився; водночас B02/A16 вимагають показувати урок без готового відео як стан, а не вилучати його з навчальної програми. `lesson_progress` зберігається по уроку (`backend/migrations/00006_playback.sql:6-14`), тому БД дозволяє завершений пізніший урок після незавершеного ранішого.

ПИТАННЯ: без правила реалізації можуть тихо розійтися між «останній оновлений», «перший незавершений» і «перший ready», що поведе користувача на різні уроки та може сховати обов’язковий no-video/processing/failed стан.

РІШЕННЯ: пропонована правка контракту — порядок програми є визначальним; `continueLessonId` дорівнює першому незавершеному уроку незалежно від стану відео; якщо прогресу немає — першому уроку; якщо всі уроки завершені — останньому уроку; якщо уроків немає — `null`. Завершений пізніший урок не перескакує незавершений раніший. Це зберігає B02/A16 як видимий стан і робить вибір детермінованим.

ФАКТ: канал brief для такої неточності — Discussion, не стоп. Тим часом реалізую саме запропоноване правило й запіню четверті випадки тестами; `docs/playback/*` не правлю через NOT-список.

ДАЛІ: дочитати auth/composition/test seams, сформувати RED-тести для доменної проєкції, HTTP-контракту й PostgreSQL integration.

ФАКТ: auth middleware вже передає `httpserver.Actor` у request context; `course.Handler.RegisterAdmin` приймає middleware через composition root, а `playback.Handler.Register` показує такий самий шаблон для автентифікованих public endpoints. `course.domain.Actor` уже містить user id і роль, тому новий identity/course port не потрібен.

РІШЕННЯ: мінімальний slice — нова `LearningRepository` у `course` ports, application `Learning`, прямий PostgreSQL reader, окрема HTTP registration через наявний `Authenticate`, OpenAPI operation і API composition wiring. Public catalog DTO/query лишаються byte-for-byte незмінними.

ДАЛІ: перевірити architecture boundary scanner і точні DB/query/testing helpers; потім додати RED-тести до production-коду.

ФАКТ: architecture scanner і наявні адаптери дозволяють тест-only cross-module contract та прямий PostgreSQL reader. Але до RED знайдено відсутній зовнішній контракт: єдиний новий endpoint — `GET /api/v1/learn/courses/{slug}` (`work/brief-t6.md:25`), тобто він не приймає lesson id взагалі. Водночас R18/R19 вимагають відкрити невідомий урок і відхилити не-UUID id до будь-якого DB lookup (`work/brief-t6.md:63-68`).

ФАКТ: намагатися довести R18/R19 через `POST /api/v1/playback/sessions` означало б змінювати або зараховувати T5 `play`, хоча NOT-список прямо каже «Не чіпати `play`/heartbeat: T5 їх закрив» (`work/brief-t6.md:81-83`). Вивести lesson id зі slug неможливо; response id не є request input, який можна валідувати до БД.

РІШЕННЯ: СТОП — бракує контракту адресації уроку для R18/R19. Production-код, OpenAPI та тести T6 не розпочато, щоб не вигадати path/query semantics і не порушити NOT-список.

ПИТАННЯ: де саме T6 має приймати `lessonId` для R18/R19: (1) окремий endpoint на кшталт `GET /api/v1/learn/courses/{slug}/lessons/{lessonId}`, (2) query-параметр нового course endpoint, чи (3) ці рядки треба вважати вже доказаними контрактом T5 `POST /api/v1/playback/sessions` без змін? Потрібні точні path/operationId і response semantics, якщо це новий endpoint.

ДАЛІ: після уточнення контракту — додати RED-тести й виконати T6; до того жодних припущень у публічний API не вносити.

ДАЛІ: перечитати виправлені D-6, R18/R19 і task T6, повторно звірити skill `implement`, після чого продовжити TDD без нового lesson-scoped endpoint.

ФАКТ: замовник зняв R18/R19 з T6 і прийняв запропоноване правило `continueLessonId`. Дві спроби append не змінили файл через неправильний patch-context/синтаксис; окремий no-op tool call теж не змінив стан. Фактичний EOF перевірено, журнал відновлено без втрати записів.

ФАКТ: виправлення підтверджено в джерелах: D-6 дослівно містить прийняте правило; R18/R19 переназначені T5/T7; `work/tasks.md` має T6=`in_progress`. Обов’язкові рядки T6 тепер B01, B27, R17, R21, R34 та O4; lesson-scoped endpoint не потрібен.

РІШЕННЯ: продовжую через TDD. Завершальну вимогу skill `code-review` виконаю, а commit знову не робитиму через явний NOT T6 «не торкатися git».

ДАЛІ: спроєктувати точні domain/port форми та RED-тести, використовуючи одну test-only таблицю видимості для обох production-предикатів.

## Discussion — шість media-станів не зведені до трьох learning-станів

ФАКТ: `work/brief-t6.md:29` дозволяє learning-відповіді лише `ready` / `processing` / `failed` / відсутність відео. Реальна state machine має `created`, `uploading`, `queued`, `processing`, `ready`, `failed` (`backend/internal/media/domain/asset.go:17-24`), а під час заміни старий ready asset може співіснувати з новішим незавершеним поколінням (план B40).

ПИТАННЯ: буквальний «latest generation» покаже `processing` або `failed`, хоча T5 ще може мінтити старий ready asset; буквальна передача стану порушить новий OpenAPI enum. Це дасть UI хибний непрогравний стан.

РІШЕННЯ: learning-проєкція спершу обирає найновіший неевікчений `ready`, якщо він існує; інакше найновіший неевікчений asset. `created`/`uploading`/`queued`/`processing` агрегуються в `processing`, `failed` і `ready` зберігаються, відсутній asset дає `null`. Так state відповідає фактичній здатності T5 стартувати playback і не приховує незавершене завантаження, коли ready ще немає.

ФАКТ: docs не правлю через NOT; вибір буде запінено PostgreSQL integration-тестом. Контракту достатньо, щоб продовжити через Discussion.

ДАЛІ: додати RED domain/application/shared-predicate тести, запустити адресні Go-тести й зафіксувати очікуваний RED.

ФАКТ: додано RED-тести доменної проєкції B01/B27, детермінованого `continueLessonId` з no-video та непослідовним completion, application visibility/validation і O4. O4 використовує рівно одну таблицю з усіма 9 комбінаціями status×actor та проганяє кожен рядок проти обох production-предикатів; R34 є окремо названим рядком.

ФАКТ: перед запуском помічено локальну test-fixture помилку: application test посилається на helpers з іншого Go package (`domain_test`), тому такий RED був би хибним. Production finding відсутній.

ДАЛІ: зробити application fixture self-contained, потім запустити адресні тести й зарахувати лише RED через відсутній production API.

ФАКТ: application fixture тепер self-contained; тестові пакети більше не залежать один від одного неявно.

ДАЛІ: запустити uncached domain/application tests із task-local `GOCACHE` і підтвердити очікуваний RED.

ФАКТ: RED підтверджено: `GOCACHE=/tmp/courses-t6-go-cache go test -count=1 ./internal/course/domain ./internal/course/application` exit 1 лише через відсутні `LearningCourse`/`LearningLesson`/video-state API та `application.NewLearning`; тестова інфраструктура не маскує причину.

ДАЛІ: реалізувати domain model, `LearningRepository` port і application use case до GREEN, без transport/PostgreSQL деталей.

ФАКТ: наявний `course.domain.ValidateID` можна перевикористати; межа `clientSeq` у playback — `2^31-1`. Course domain не імпортуватиме playback: ту саму числову межу назвемо локально як частину learning read contract.

ДАЛІ: додати inward-only production типи та use case.

ФАКТ: реалізовано inward-only `LearningCourse`/module/lesson/video-state модель, валідацію read projection, детермінований continue selector, другий названий visibility predicate, справжній PostgreSQL `LearningRepository` port і application use case з валідацією/3s budget/policy denial.

ДАЛІ: форматувати й запустити адресні domain/application tests; виправляти лише фактичні GREEN-проблеми.

ФАКТ: domain tests GREEN. Application test упав не через production: fake repository був жорстко прив’язаний до viewer id, тому навмисний R34 owner-виклик повернув fixture error до policy. Це test-double defect; production-код не змінюю за ним.

ДАЛІ: зробити fake записувальним замість actor-specific guard, явно assert-ити переданий user/slug і повторити адресний прогін.

ФАКТ: fake repository тепер записує фактичні user/slug; тести окремо перевіряють viewer та owner lookup, не підміняючи visibility policy власним guard.

ДАЛІ: повторити uncached domain/application tests.

ФАКТ: domain/application GREEN: `go test -count=1 ./internal/course/domain ./internal/course/application` з `GOCACHE=/tmp/courses-t6-go-cache`, exit 0.

ДАЛІ: додати RED HTTP/OpenAPI tests для auth, exact learning JSON, B01/B27/R17 і R21; публічний handler лишити незміненим.

ФАКТ: `httpserver.WithActor` і наявний middleware injection достатні для transport tests. OpenAPI 3.0.3 використовує nullable fields та `no-store`; learning response може точно моделювати no-video/empty-course без нового transport primitive.

РІШЕННЯ: новий `LearningHandler` реєструватиметься окремо через `Authenticate`; existing public `Handler.Register` і `courseDetailDTO` не змінюються, що робить R21 структурно запіненим.

ДАЛІ: додати RED transport contract matrix та explicit anonymous/public-model assertions.

ФАКТ: додано RED transport tests: authenticated exact response+OpenAPI, B01 empty arrays/null continue, B27 concrete zero defaults, R17 forbidden, dependency/not-found/query/invalid-slug handling і explicit R21 assertion, що анонімний public lesson має тільки `title`.

ФАКТ: до запуску помічено зайвий test import із штучним sentinel; він не є частиною доказу й буде прибраний, щоб RED мав одну причину — відсутній learning transport/OpenAPI.

ДАЛІ: очистити test import, запустити HTTP package й зафіксувати RED.

ФАКТ: зайвий import/sentinel прибрано; transport test self-contained.

ДАЛІ: запустити uncached HTTP test package для RED.

ФАКТ: HTTP RED підтверджено: package build падає лише на відсутньому `coursehttp.NewLearningHandler`; existing public handler/tests компілюються до цієї точки.

ДАЛІ: додати thin learning transport і OpenAPI path/schemas, не редагуючи generated client та public catalog implementation.

ФАКТ: OpenAPI playback patterns підтверджують response refs і no-store contract. Learning GET матиме 200/400/401/403/404/500/503; окремий learning-forbidden response відрізнятиме policy denial від same-origin playback mutation.

ДАЛІ: реалізувати handler/DTO/error mapping і OpenAPI schema.

ФАКТ: додано окремий authenticated `LearningHandler`, exact DTO mapping із nullable `videoState`/`continueLessonId`, strict no-query behavior і policy/error mapping. OpenAPI містить `getLearningCourse`, learning course/module/lesson schemas і 200/400/401/403/404/500/503 responses. Existing `getPublishedCourse` implementation/schema не змінено; `generated/` не торкнуто.

ДАЛІ: форматувати й запустити HTTP/OpenAPI tests; перевірити exact contract errors.

ФАКТ: HTTP прогін упав лише в contract helper: тест навмисно подав invalid slug, але helper безумовно вимагав валідний OpenAPI request перед перевіркою 400 response. Handler повернув очікувану відмову й repository не викликався; production defect не доведено.

ДАЛІ: навчити helper розрізняти valid/invalid request, як existing catalog test, і повторити package tests.

ФАКТ: contract helper тепер валідовує request лише для syntactically valid cases, але завжди валідовує response; invalid slug/query можуть коректно довести 400 schema.

ДАЛІ: повторити uncached HTTP package tests.

ФАКТ: HTTP/OpenAPI package GREEN: `go test -count=1 ./internal/course/adapters/http` з task-local cache, exit 0. B01/B27/R17/R21 transport assertions проходять.

ДАЛІ: додати integration-tagged PostgreSQL RED для ordered projection, progress defaults, visibility/no-write, R34 і media-state selection/aggregation.

ФАКТ: додано focused PostgreSQL integration test: B01; B27 zero defaults; order; accepted continue rule; old-ready-vs-new-processing; failed/no-video projection; R17 403/no inserted progress; R34 archived owner; not-found classification.

ФАКТ: до runtime помічено fixture violation: `failed` asset потребує `failure_reason` за migration constraint. Виправляю fixture до RED compile, щоб майбутній Docker-прогін не падав на setup.

ДАЛІ: додати failure reason у fixture й виконати integration-tag compile-only; очікуваний RED — відсутній repository method.

ФАКТ: failed fixture тепер задовольняє `video_assets_failed_has_reason`; інші ready/pending constraints теж задані явно.

ДАЛІ: форматувати integration test і запустити `-tags=integration -run '^$'` compile-only для course postgres.

ФАКТ: integration RED compile-only підтверджено: exit 1 лише через відсутній `Repository.GetLearningCourse`; testcontainers/runtime не запускались.

ДАЛІ: реалізувати один read-only PostgreSQL query із прямими joins, deterministic ordering, user-scoped LEFT JOIN progress і media projection rule з Discussion.

ФАКТ: PostgreSQL reader реалізовано одним read-only SELECT: course→modules→lessons, lateral current-video selection, caller-scoped progress defaults і curriculum ordering. Він не імпортує `media`/`playback`, не створює міжмодульного порту й не пише в БД.

ФАКТ: статичне перечитання до compile знайшло пропущений stdlib `time` import у локальному mapper; це механічна compile-помилка, не контрактна знахідка.

ДАЛІ: додати import, форматувати, повторити integration compile-only та всі course package tests.

ФАКТ: stdlib `time` import додано; production mapping тепер повністю типізований.

ДАЛІ: виконати gofmt, integration-tag compile-only і uncached course package tests.

ФАКТ: PostgreSQL integration suite compile-only GREEN; всі `./internal/course/...` host tests GREEN. Docker/testcontainers runtime не запускався, тож SQL поведінка ще reviewer-only proof.

ДАЛІ: підключити learning use case/handler у API composition root, потім перевірити API package й architecture boundary.

ФАКТ: API composition root тепер створює request-independent course repository, `Learning` use case і окремий handler; route обгорнутий чинним session `Authenticate`. Нових global mutable dependencies немає.

ДАЛІ: форматувати composition root і запустити API/architecture tests та всі touched packages.

ФАКТ: API composition, architecture і всі touched host packages GREEN; integration-tagged course postgres compile-only також GREEN.

РІШЕННЯ: буквальне формулювання O4 каже «одна структура даних, спільна для обох тестів». Поточний один cross-package test сильний, але це все ж один тест. Щоб виконати brief дослівно, винесу нейтральні 9 cases у test-support provider і запущу їх окремо в `course/domain` та `playback/domain` tests.

ДАЛІ: замінити combined O4 test двома package-local tests, що імпортують один case provider; повторити обидва domain packages.

ФАКТ: перший O4 refactor patch не застосувався через mismatch поля `test.Name` проти фактичного `test.name`; змін не було. Точний блок перечитано.

ДАЛІ: застосувати вузький patch за фактичним контекстом.

ФАКТ: O4 тепер має один provider `internal/testsupport/coursevisibility.Cases()` з 9 cases і два package-local tests — `course/domain` та `playback/domain`. Обидва independently маплять нейтральні actor/status values у власні production типи; R34 archived owner входить у єдину таблицю.

ДАЛІ: форматувати й запустити обидва domain packages, test-support package та architecture gate.

ФАКТ: обидва O4 domain tests, shared provider package й backend architecture test GREEN. Дублювання policy тепер ламається голосно від одного набору cases.

ДАЛІ: перечитати Make targets і виконати дозволені гейти строго у порядку `format-check` → `lint` → `typecheck` → host Go tests; Docker-залежні не запускати.

ФАКТ: Makefile підтверджує, що `format-check` і `lint` мають frontend host step та backend Docker step; `typecheck` повністю host. `test` містить Docker backend, тому замість нього brief дозволяє host Go tests без testcontainers. `integration`/`check-generated`/`generate`/`build`/`up`/`e2e` не запускаються.

ДАЛІ: запустити `make format-check` і чесно розділити frontend результат від Docker-відмови.

ФАКТ: `make format-check` НЕ GREEN, exit 2: frontend `oxfmt --check` PASS (140 files), backend Docker gofmt не стартував — permission denied on Docker socket. Це очікувана named environment refusal, не обхід.

ДАЛІ: виконати host `gofmt -l` як додатковий доступний доказ, потім перейти до `make lint` у нормативному порядку.

ФАКТ: host `test -z "$(gofmt -l .)"` у backend PASS, exit 0. Це додатковий proof, не заміна не-green make gate.

ДАЛІ: запустити `make lint` і окремо записати frontend/architecture та Docker Go-vet результати.

ФАКТ: `make lint` НЕ GREEN, exit 2: frontend oxlint + architecture PASS (78 files); Dockerized `go vet` не стартував через permission denied on Docker socket.

ДАЛІ: запустити host `go vet ./...` з task-local `GOCACHE`, потім `make typecheck`.

ФАКТ: host `GOCACHE=/tmp/courses-t6-go-cache go vet ./...` PASS, exit 0; task-local cache використано через sandbox/system-cache обмеження.

ДАЛІ: запустити `make typecheck`.

ФАКТ: `make typecheck` PASS, exit 0 (`tsc --noEmit`). Generated client ще старий за задумом brief, але чинний frontend typecheck не залежить від нової operation до reviewer generation.

ДАЛІ: запустити повний uncached host `go test ./...` без integration tag/testcontainers; named sandbox failures не маскувати.

ФАКТ: повний host `go test -count=1 ./...` НЕ GREEN, exit 1: усі course/identity/media/playback packages і architecture PASS; єдина відмова — existing `platform/httpserver.TestShutdownDrainsAnInFlightResponse` не може bind `127.0.0.1:0` у sandbox (`operation not permitted`). Це середовищна відмова, не T6 test failure; її не обходжу.

ДАЛІ: виконати завершальний review через обов’язковий skill `code-review`, спочатку прочитавши його повністю; git не використовувати, review scope скласти з відомого T6 file set.

ФАКТ: skill `code-review` прочитано повністю. Його штатний fixed-point/diff процес потребує git, що прямо заборонено T6. User scope overrides skill ceremony за `.claude/standards.md`; git не викликаю.

РІШЕННЯ: зберігаю дві незалежні review-осі skill через два read-only sub-agents — Standards і Spec — але даю їм відомий T6 file set замість git diff та прямо забороняю git/редагування. Це fallback через NOT, не прихована зміна scope.

ДАЛІ: запустити паралельні read-only Standards/Spec reviews і паралельно виконати локальний статичний audit контракту/NOT-списку.

ФАКТ: локальний статичний audit підтвердив: course production не імпортує `media`/`playback`; joins живуть лише в postgres adapter; transport лише перекладає context/path/DTO/errors; public title-only program незмінний; єдиний `/learn/` endpoint course-scoped; generated files не редагувались у T6.

РІШЕННЯ: `mapCourseText` має зайву локальну pointer-сходинку, але це judgement-call readability, не дефект. Спростити її до прямого `storedPublishedAt.Valid` branch перед фінальними тестами.

ДАЛІ: виконати малий cleanup mapper, повторити touched tests і дочекатися двох незалежних review verdicts.

ФАКТ: PostgreSQL course mapper спрощено до одного nullable `*time.Time` branch; поведінка незмінна, зайву проміжну `*pgtype.Timestamptz` прибрано.

ДАЛІ: gofmt + uncached touched tests + integration compile-only після cleanup.

ФАКТ: post-cleanup touched suite GREEN: course domain/application/http/postgres, playback domain, architecture, API composition; course PostgreSQL integration compile-only GREEN.

ДАЛІ: отримати обидва незалежні review verdicts; blocking findings виправити й перевірити, non-blocking назвати.

ФАКТ: перше очікування review (мінімум 10s, запитаний timeout був clamp-нутий) завершилось без verdict; агенти ще працюють.

ДАЛІ: поки review триває, сформувати required evidence matrix, generation delta й список незапущених гейтів за фактичними test anchors.

ФАКТ: exact anchors зібрано: B01 domain+HTTP+integration; B27 domain+HTTP+integration; R17 HTTP+real-DB no-write integration; R21 explicit public response shape; R34 shared policy cases+application+integration; O4 один provider і два domain tests. Integration anchors поки compile-only, не runtime.

ДАЛІ: дочекатися review verdicts, потім внести blocking fixes перед фінальною evidence matrix.

ФАКТ: Standards review повернув 1 hard finding: `docs/architecture/backend.md` вимагає sqlc для course PostgreSQL, а T6 reader зараз має raw SQL. Judgment calls: course-row data clump, stringly test-support cases, повторені fixtures. Deliberate predicate duplication не визнано smell.

РІШЕННЯ: hard finding перевіряю проти повного standard і NOT «не писати generated/». Якщо sqlc обов’язковий, не маскуватиму конфлікт raw query; знайду спосіб лишити reviewer generation без ручного generated edit або зафіксую named limitation.

ДАЛІ: прочитати `docs/architecture/backend.md` повністю й отримати Spec review.

ФАКТ: backend architecture описує course PostgreSQL як sqlc-generated, але не формулює заборони raw SQL; AGENTS вимагає generated models лишати в adapters, а T6 прямо забороняє ручний `generated/` і відкладає `make generate`. Standards finding тому потребує judgment, а не автоматичного rewrite у non-compiling стан. Existing media direct-query precedent і brief підтримують поточний adapter-local SQL.

ФАКТ: Spec review повернув High: поточний R17 test є хибнопозитивом — read-only learning lookup не може довести, що progress write недоступного уроку не створює рядка. Medium: R34 доведено policy/read, але не segment 200 end-to-end.

РІШЕННЯ: R17 High приймаю і виправлю через наявний playback write path без нового endpoint. R34 medium не розширюю до edge/e2e: brief §2.2 призначає R34 саме shared predicate table, забороняє edge і Docker; full segment behavior лишається e2e-рівнем плану, тут чесно позначу як незапущений runtime proof.

ДАЛІ: прочитати existing Sessions/Progress integration seams і додати один справжній R17 write attempt, не змінюючи play/heartbeat production.

## Discussion — R17 суперечить NOT «не чіпати heartbeat» і відкриває реальний write bug

ФАКТ: `backend/internal/playback/application/sessions.go:38-54` у чинному коді heartbeat перевіряє лише Redis ownership, після чого без visibility check викликає `progress.Save`. Тому сесія, мінтована на published курсі, після archive/draft отримує 403 на медіабайти, але heartbeat усе ще може записувати `lesson_progress`. Поточні R15/R16/R31/R33 доводять ownership-відмови, не course-policy denial.

ПИТАННЯ: `work/brief-t6.md:67` вимагає R17 `403; рядка не з’явилось`, але NOT `work/brief-t6.md:82` каже не чіпати heartbeat. Без production policy check справжній write-path тест закономірно впаде; read-only test був хибнопозитивом саме з тих, про які попереджає brief.

РІШЕННЯ: виправлення — перед Redis renewal/write `Sessions.Heartbeat` перевикористовує вже наявний `LessonPlaybackReader`, отримує actor (не лише userID) і застосовує той самий `VideoCourse.CanBeViewedBy`; нового порту немає. HTTP повертає наявний `playback_forbidden` 403. Це мінімальна необхідна правка heartbeat production для явного R17, зафіксована через Discussion; інші T5 semantics не змінюються.

ДАЛІ: додати RED unit/integration R17 через реальний heartbeat, потім мінімально змінити application/transport signature та повторити всі playback tests.

ФАКТ: existing unit/integration seams підтверджують defect: `Sessions` уже має `LessonPlaybackReader`, але heartbeat його не викликає; integration harness має real Redis/PostgreSQL і може довести відсутність row після archive.

ДАЛІ: додати RED policy-before-renew unit test і real R17 subtest до existing integration harness; спершу прогнати unit та integration compile-only без production change.

ФАКТ: R17 можна вставити після B44 і до R32, відновивши course status у `published`; це не змінить ordering/fixtures наступного test і дозволить exact row-count assertion.

ДАЛІ: внести лише test changes.

ФАКТ: додано RED unit, який вимагає course visibility lookup до Redis renewal/progress, і real Redis/PostgreSQL integration subtest, який архівує курс, видаляє progress, шле heartbeat, очікує forbidden та count=0, після чого відновлює published для R32.

ДАЛІ: запустити uncached application unit; integration лише compile-only через Docker restriction.

ФАКТ: R17 unit очікувано RED: heartbeat повернув nil і записав би progress замість forbidden. Integration-tag compile-only GREEN. Composite shell повернув 0 через успіх останньої команди попри unit failure, тому загальний exit code НЕ зараховано; надалі verification composite використовує `set -eu`.

ДАЛІ: змінити heartbeat service signature на actor, виконати existing playback lookup+predicate до renewal, оновити thin HTTP adapter/test callers і довести GREEN.

ФАКТ: heartbeat тепер приймає domain.Actor, завантажує lesson/course через existing reader, застосовує `CanBeViewedBy` до Redis renewal і progress save. HTTP передає вже валідований actor; unit/integration callers і fake оновлені. Жодного нового порту або endpoint немає.

ДАЛІ: audit усіх Heartbeat call sites, форматування, application+HTTP unit GREEN та integration compile-only із `set -eu`.

ФАКТ: call-site audit знайшов лише очікувані production interface/handler та оновлені unit/integration calls; старої userID-only сигнатури не лишилось.

ДАЛІ: запустити строгий composite verification із `set -eu`.

ФАКТ: strict post-R17 composite GREEN: playback application/HTTP/domain unit tests, playback application integration compile-only, all course packages, architecture і API composition. R17 unit тепер доводить policy before renewal/write; runtime integration лишається reviewer-only.

ДАЛІ: попросити обидві review-осі перевірити fix: Spec — R17 false-positive closure; Standards — heartbeat delta та чи raw SQL є hard rule з урахуванням brief/NOT/generated constraints.

ФАКТ: обом review-агентам надіслано follow-up із post-R17 file scope та explicit no-git/no-edit. Verdicts очікуються.

ДАЛІ: локально перевірити transport mapping R17→403 `playback_forbidden` і додати regression row, якщо heartbeat-specific case відсутній.

ФАКТ: mapper уже повертає 403 `playback_forbidden`, але heartbeat OpenAPI 403 посилається лише на `OriginForbidden`, а contract matrix не має policy-forbidden heartbeat case. Після R17 production fix це стало неточним public contract.

РІШЕННЯ: змінити heartbeat 403 ref на `PlaybackForbidden` (опис покриває Origin або course policy) і додати heartbeat-forbidden до OpenAPI response matrix; endpoint/body semantics не змінюються.

ДАЛІ: внести contract/test fix і повторити playback HTTP tests.

ФАКТ: heartbeat OpenAPI 403 ref і contract matrix оновлено для R17 `playback_forbidden`.

ДАЛІ: перевірити, що patch змінив саме heartbeat path, потім gofmt/test OpenAPI matrix.

ФАКТ: перевірка знайшла, що broad YAML patch не змінив heartbeat ref: він лишився `OriginForbidden`; попередній запис про оновлення був передчасним. Spec follow-up також знайшов хибну R17 fixture: actor був одночасно course owner, тому archived доступ дозволений і runtime test впаде.

РІШЕННЯ: виправити обидва до наступного verdict: exact heartbeat YAML context; окремий student fixture/ownership/row count для R17, owner лишається instructor. Course status restore зробити cleanup-safe.

ДАЛІ: внести exact fixes, потім HTTP unit + integration compile-only.

ФАКТ: exact heartbeat OpenAPI ref тепер `PlaybackForbidden`; R17 integration actor — окремий student, course owner лишився instructor, row count user-scoped, status restore виконується через cleanup. Spec review підтвердив, що додатковий R34 e2e у T6 не потрібен і відкликав Medium.

ФАКТ: Standards follow-up відкликав raw-SQL hard finding: це consistency tradeoff, не заборона; hard findings = 0. Material smell: security-sensitive load+visibility guard дублюється в Play/Heartbeat; course-row scalar clump у mapper.

РІШЕННЯ: усунути обидва non-blocking smells до фіналу: один private `loadViewablePlayback` у application; private adapter row struct для course mapping. Спочатку перевірити fixture/contract fix, потім refactor з тестами.

ДАЛІ: запустити playback HTTP unit та application integration compile-only після exact fixes.

ФАКТ: exact fixes GREEN: playback application+HTTP unit PASS, application integration compile-only PASS, inspected OpenAPI heartbeat 403 points to `PlaybackForbidden`.

ДАЛІ: extract shared application guard і adapter course-row struct, потім rerun touched suites.

ФАКТ: `Play` і `Heartbeat` тепер викликають один private `loadViewablePlayback`, тож security policy не може дрейфувати між ними. PostgreSQL mapper приймає `learningCourseRow` замість дев’яти scalar parameters.

ДАЛІ: gofmt і strict touched verification, включно з integration compile-only.

ФАКТ: post-review-fix strict suite GREEN: all course packages, playback application/HTTP/domain, architecture, API composition, both integration suites compile-only, full host `go vet ./...`.

ДАЛІ: отримати final delta-review після distinct-student R17 fixture та smell refactors; потім повторити дозволені gates, бо production змінився після першого gate pass/fail set.

ФАКТ: final delta-review запущено на обох осях. До verdict повторюю доступні гейти в нормативному порядку після останньої production зміни.

ДАЛІ: повторити `make format-check`.

ФАКТ: фінальний `make format-check` НЕ GREEN, exit 2: frontend format PASS (140 files); backend Docker gofmt знову не стартував через permission denied on Docker socket.

ФАКТ: final reviews завершені: Standards — 0 hard findings, 0 material smells; Spec — 0 findings. R17 distinct-student write proof і R34 shared-table scope підтверджені.

ДАЛІ: повторити host gofmt proof, потім `make lint`.

ФАКТ: фінальний host gofmt proof PASS, exit 0.

ДАЛІ: повторити `make lint`.

ФАКТ: фінальний `make lint` НЕ GREEN, exit 2: frontend oxlint+architecture PASS (78 files); Dockerized Go vet не стартував через permission denied.

ДАЛІ: повторити host Go vet, потім `make typecheck`.

ФАКТ: фінальний host `go vet ./...` PASS з task-local cache.

ДАЛІ: повторити `make typecheck`.

ФАКТ: фінальний `make typecheck` PASS, exit 0.

ДАЛІ: виконати фінальний uncached host `go test ./...`; очікувану sandbox bind-відмову не вважати PASS.

ФАКТ: фінальний full host `go test -count=1 ./...` НЕ GREEN, exit 1: усі business/adapters/application/domain/architecture packages PASS, включно з T6 і playback R17; єдина відмова — existing `platform/httpserver` lifecycle test cannot bind `127.0.0.1:0` у sandbox (`operation not permitted`).

ДАЛІ: виконати `-race` лише для touched packages як доступний еквівалент backend частини make test, потім скласти фінальний звіт/матрицю.

ФАКТ: перший touched `-race` composite не завершився в 30s yield; вивід підтвердив course domain/application PASS, але session id не було збережено оркестраційним wrapper, тому решту не зараховую. Це незавершений proof, не failure.

ДАЛІ: повторити race меншими package groups до terminal exit, потім integration compile-only окремо.

ФАКТ: course HTTP/PostgreSQL adapter `go test -race -count=1` PASS, terminal exit 0.

ДАЛІ: race playback domain/application/HTTP.

ФАКТ: playback domain/application/HTTP `go test -race -count=1` PASS, exit 0; R17 unit входить у цей прогін.

ДАЛІ: race course domain/application + architecture/API composition, потім integration compile-only.

ФАКТ: course domain/application, architecture й API composition race run PASS, exit 0.

ДАЛІ: фінальний integration-tag compile-only для course postgres і playback application.

ФАКТ: обидві integration suites compile-only PASS; жоден testcontainers runtime не запускався.

## Рядок §8 → чим доводиться

| Рядок | Доказ | Стан тут |
|---|---|---|
| B01 | `course/domain.TestLearningCourseContinueLessonRule/B01_no_lessons`; `course/http.TestB01LearningCourseWithoutLessonsUsesEmptyArraysAndNullContinue`; `course/postgres.TestLearningReaderAgainstPostgreSQL/B01...` | domain/HTTP race PASS; integration compile PASS, runtime reviewer |
| B27 | `course/domain.TestB27MissingProgressUsesConcreteZeroValues`; exact HTTP zero-value assertion; PostgreSQL fixture без progress row перевіряє `position=0`, `completed=false`, `clientSeq=0` | domain/HTTP race PASS; integration compile PASS, runtime reviewer |
| R17 | `playback/application.TestR17HeartbeatChecksCourseVisibilityBeforeRenewalAndProgress`; real Redis/PostgreSQL `R17 inaccessible course heartbeat...` із distinct student, 403/no row; heartbeat OpenAPI 403 matrix | unit/HTTP race PASS; integration compile PASS, runtime reviewer |
| R21 | `course/http.TestR21AnonymousPublicCourseProgramStillExposesNoLessonIDs`: кожен public lesson має рівно поле `title`; existing `getPublishedCourse` code/schema не змінені | race PASS |
| R34 | єдиний shared case `R34 archived owner` проганяють обидва O4 tests; application owner-read test; PostgreSQL archived-owner read | shared unit/race PASS; integration compile PASS, runtime reviewer; edge/e2e поза дозволеним T6 scope |
| O4 | `internal/testsupport/coursevisibility.Cases()` — одна 9-row структура `(published/draft/archived) × (student/owner/admin)`; `course/domain.TestO4CourseVisibilityUsesSharedCases` і `playback/domain.TestO4PlaybackVisibilityUsesSharedCourseCases` | обидва race PASS; divergence валить build |

## Правило `continueLessonId`

РІШЕННЯ: порядок програми визначальний. `continueLessonId` — перший незавершений урок незалежно від стану відео; якщо progress немає — перший урок; якщо всі уроки завершені — останній; якщо уроків немає — `null`. Завершений пізніший урок не перескакує незавершений раніший. No-video/processing/failed урок лишається видимим станом B02/A16, а не відфільтровується як «не ready».

## Що має дати `make generate`

- Backend sqlc output має лишитися byte-identical: migrations і `backend/sql/queries/*.sql` у T6 не змінювались; learning reader навмисно adapter-local за direct-join прецедентом brief.
- Orval має додати operation `getLearningCourse` для `GET /api/v1/learn/courses/{slug}` у `frontend/src/adapters/api/generated/courses.ts`.
- Orval має створити/експортувати generated models для `LearningCourse`, `LearningModule`, `LearningLesson` і nullable enum `processing | ready | failed`, оновити `models/index.ts` та response typings 200/400/401/403/404/500/503.
- Нових error codes немає: `invalid_query`, `invalid_slug`, `authentication_required`, `course_forbidden`, `course_not_found`, `dependency_unavailable`, `internal_error` уже входять у загальний enum.
- Generated файли тут не редагувались; materialization і `check-generated` лишені рецензентові рівно за brief.

## Фінальна верифікація і незапущені гейти

- `make format-check` — НЕ GREEN: frontend format PASS; Docker backend gofmt не стартував (`permission denied` on Docker socket). Host `gofmt -l` clean.
- `make lint` — НЕ GREEN: frontend oxlint+architecture PASS; Dockerized Go vet не стартував. Host `go vet ./...` PASS із `GOCACHE=/tmp/courses-t6-go-cache`.
- `make typecheck` — PASS.
- Touched Go packages з `-race` — PASS: course domain/application/HTTP/PostgreSQL adapter, playback domain/application/HTTP, architecture, API composition.
- Full host `go test -count=1 ./...` — НЕ GREEN лише через sandbox bind refusal в existing `platform/httpserver.TestShutdownDrainsAnInFlightResponse`; усі інші packages PASS.
- Integration-tag compile-only — PASS для course PostgreSQL reader і playback R17 application suite.
- НЕ ЗАПУСКАЛИСЯ через Docker restriction: `make test` (Dockerized backend half), `make integration`, `make generate`, `make check-generated`, `make build`, `make up`, `make smoke`, `make e2e`, runtime testcontainers integration та Compose/HTTP flows.

## Review verdict

- Standards: 0 hard findings, 0 material smells після shared `loadViewablePlayback` і `learningCourseRow` cleanup. Raw adapter SQL — consistency tradeoff, не hard rule; він узгоджений із brief direct-join precedent і NOT щодо generated.
- Spec: 0 findings після виправлення R17 false-positive. Distinct student real-DB fixture, policy-before-renew/write і 403 OpenAPI mapping перевірені статично/compile-only; runtime лишений рецензентові. R34 достатньо закритий shared predicate table у межах T6.
- Git і `generated/` не торкались за NOT-списком; commit не створено.

РІШЕННЯ: T6 implementation і вся доступна локальна verification завершені. Runtime SQL/R17 integration та generated client лишаються явно незапущеними reviewer steps, не видаються за PASS.

---

## Верифікація рецензентом (2026-09-08)

### Гейти

**Повний `make check` — чистий**: 334 фронт-тести, e2e **35/35**. Порядок цього разу був
правильний одразу — регенерувати, закомітити, прогнати (урок T5).

### O4 — спільна таблиця предиката справді спільна

`backend/internal/testsupport/coursevisibility/cases.go` — **девʼять випадків**
(`published`/`draft`/`archived` × студент/власник/адмін), і **обидва** тести ітерують той
самий виклик `Cases()`:

```
course/domain/learning_test.go:68        for _, test := range coursevisibility.Cases()
playback/domain/authorization_test.go:16 for _, test := range coursevisibility.Cases()
```

Це саме те, чого вимагав бриф: **одна структура даних**, а не два схожі списки. Розійтися
реалізації тепер не можуть тихо — щоб змінити набір, треба змінити спільний файл, і це
одразу бачать обидві сторони. **O4 закрито.**

Окремо визнаю: перші два греп-заходи цієї таблиці **не знайшли**, і я був близький до
висновку, що її немає. Знайшлась пошуком по каталогах, а не по вмісту. Висновок «немає» з
невдалого пошуку — та сама помилка, що вже коштувала фазі чотирьох хибних причин.

### R17 — виправлення доведено поіменно

```
TestR17HeartbeatChecksCourseVisibilityBeforeRenewalAndProgress
TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL/R17_inaccessible_course_heartbeat_is_forbidden_without_progress
TestLearningHTTPBoundariesAndDenials/R17_unavailable_course
```

Три рівні: unit на політику перед поновленням, integration проти живих Redis + PostgreSQL із
перевіркою **відсутності рядка**, і HTTP-межа.

### Discussion — три, усі прийнято

**D3 — дефект у зданому коді T5, перший такий за фазу.** Heartbeat перевіряв лише володіння
в Redis і писав прогрес **без перевірки видимості курсу**
(`playback/application/sessions.go:38-54`). Сесія, змінтована на опублікованому курсі, після
архівації діставала 403 на **байти**, але **продовжувала писати** в `lesson_progress`. Жоден
наявний рядок цього не ловив: R15/R16/R31/R33 доводять відмови **володіння**, не політики.

Мій NOT-список забороняв чіпати heartbeat — і за буквою лишав два виходи: пропустити R17 або
написати read-only тест, тобто **хибнопозитив рівно того ґатунку, про який попереджає цей же
бриф**. Виконавець не зробив ні того, ні того.

**Урок для наступних брифів:** «не чіпати X» треба писати як «не розширювати X **новою
функціональністю**». Заборона, сформульована як абсолют, ледь не сховала дефект.

**D1 — `continueLessonId`** отримав детерміноване правило (прийнято дослівно, записано в план).
**D2 — шість media-станів** зведені до трьох learning-станів явно, а не за замовчуванням.
