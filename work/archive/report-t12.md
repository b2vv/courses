ДАЛІ: Повністю прочитати work/brief-t12.md, усі матеріали з його §1, work/codex-working-model.md та інструкції скіла implement; після цього звірити контракт T12 з поточним кодом.
ФАКТ: work/brief-t12.md прочитано повністю. T12 вимагає client-only неблокувальну AV1-пробу рівно раз на завантаження сторінки, опційне enum-поле мінту, одну наперед засіяну CounterVec-метрику та вимірювальні тести; P12 без піднятого API доведеним оголошувати не можна.
РІШЕННЯ: Розвилку §2.3 не обходити: перевірити можливість явного required/optional поділу спільного decodePlaybackObject з тестами play і heartbeat. Docker-залежні гейти та T11-файли не запускати/не чіпати.
ДАЛІ: Прочитати всі джерела §1 брифа, робочу модель і SKILL.md implement, включно з точними ділянками плану, спеки, контракту, декодера, метрик, клієнтського мінту й стану T11.
ФАКТ: work/codex-working-model.md, AGENTS.md та implement/SKILL.md прочитано повністю. Робоча модель підтверджує: ця сесія пише код і локальні без-Docker перевірки, реценз-сесія запускає повний гейт і комітить; missing contract означає стоп, неточність із ризиком баґа — `## Discussion`.
РІШЕННЯ: Використати TDD на явних швах і code-review наприкінці. Інструкцію implement про commit не виконувати, бо прямі правила користувача й робочої моделі забороняють торкатися git і делегують commit реценз-сесії.
ДАЛІ: Прочитати tdd/SKILL.md, а тоді решту матеріалів §1 брифа та work/CTO-RESEARCH.md, який AGENTS.md вимагає читати першим для репозиторію.
ФАКТ: tdd/SKILL.md і work/CTO-RESEARCH.md прочитано. Погоджені брифом публічні шви для TDD: клієнтський PlaybackRepository/тіло createPlaybackSession, HTTP-контракт play/heartbeat і Prometheus registry/лічильник. Масовий вивід решти джерел був усічений інструментом, тому їх ще не вважаю прочитаними повністю.
РІШЕННЯ: Йти вертикальними red→green зрізами на цих уже визначених у брифі швах; не тестувати приватну реалізацію проби окремо, якщо поведінку можна побачити через repository.
ДАЛІ: Дочитати без усічення work/tasks.md, docs/playback/spec.md, docs/playback/_wip-t11-review.md і релевантні частини плану та OpenAPI; потім оглянути код і тести.
ФАКТ: Матеріали §1 прочитано: план P12 і пʼять питань, D-7 про рівно одну пробу на page load, A17/out-of-scope, PlaybackSessionRequest, Metrics, клієнтський play та повний _wip-t11-review. Також прочитано поточний Phase 6 tracker і код спільного декодера. P12 має два окремі спостережувані: поле в POST і приріст метрики на 1; AV1 ladder заборонена.
ФАКТ: QueryClient створюється всередині getRouter() на кожен SSR request і живе одне browser page load після гідратації. decodePlaybackObject зараз вимагає точну множину keys через len(seen)==len(fields), тому просте додавання av1Support зробить поле обовʼязковим.
РІШЕННЯ: Використати request-scoped QueryClient як носій однократності, а спільний декодер розділити на required та optional поля; heartbeat лишити з порожньою optional-мапою та додати регресійні тести його точної множини. Це проходить обидві явні розвилки брифа без Discussion.
ДАЛІ: Інвентаризувати production/test seams, генерацію Orval і композицію метрик, не читаючи й не змінюючи заборонені T11-файли; визначити найменший вертикальний red→green зріз.
ФАКТ: Інвентаризація знайшла три наявні публічні test seams: frontend adapter test перехоплює JSON мінту через MSW; backend handler_test проганяє справжній HTTP handler та OpenAPI validator; metrics_test читає exposition із приватного registry. API composition уже має один Metrics на executable.
ФАКТ: Просте кешування результату проби недостатнє: якби кожен повторний мінт ніс кешоване поле, сервер збільшував би метрику знову. Тому page-scoped носій мусить зберігати і результат, і стан «уже відправлено»; перший play бере значення, наступні мінти опускають optional field.
РІШЕННЯ: Обрати синхронний `MediaSource.isTypeSupported('video/mp4; codecs="av01.0.05M.08"')`: він прямо перевіряє потрібний decoder/container seam, не має promise, отже не може зависнути; відсутність API або throw мапиться в `probe_failed`. Закритий enum: `supported`, `unsupported`, `probe_failed`.
РІШЕННЯ: Назва серії — `courses_playback_av1_probe_total`, label `result`, наперед засіяні всі три enum-значення. Метрика інкрементується лише після успішного мінту і лише коли optional поле присутнє; старий клієнт без поля не створює події.
ДАЛІ: TDD slice 1 — додати frontend tests на supported/unsupported/probe_failed, одноразову пробу й одноразове поле мінту; запустити точкові тести й зафіксувати red до production-коду.
ФАКТ: TDD slice 1 red підтверджено: `npm test -- src/app/composition/av1-probe.test.ts` завершився exit 1 через відсутній модуль `./av1-probe` (0 тестів виконано).
ФАКТ: Мінімальна реалізація page-scoped носія додана в `frontend/src/app/composition/av1-probe.ts`: QueryClient тримає `{result, reported}`, `gcTime=Infinity` на час сторінки, повторний prime не питає API, повторний take не повертає поле; SSR guard повертає до звернення до MediaSource.
ФАКТ: TDD slice 1 green: та сама точкова команда — exit 0, 4/4 тести; окремо перевірені `supported`, `unsupported`, відсутній API та throw.
РІШЕННЯ: Carrier не є глобальним mutable store: його стан належить конкретному QueryClient, який `getRouter()` створює на SSR request/browser page load; `gcTime=Infinity` потрібен, щоб стан не зник через пʼять хвилин без observer.
ДАЛІ: TDD slice 2 — розширити adapter test: перший мінт несе конкретний enum, повторний мінт після споживання поля не несе; actual body має бути в діагностиці до assertion. Потім підключити carrier у client-only маршруті й adapter.
ФАКТ: TDD slice 2 red: adapter suite 10/11, фактичні тіла надруковано до assert — обидва були `{lessonId}`, очікувалось перше `{lessonId,av1Support:"supported"}`. Це точне спостереження відсутнього wire field.
ФАКТ: Після мінімальної зміни adapter callback slice 2 green: adapter suite 11/11; надруковані тіла — перше з `supported`, друге без optional поля.
ФАКТ: Carrier підключено в `createLearnDependencies(queryClient)`; client-only маршрут `/learn/$` має `ssr:false` і prime-ить пробу після валідного lesson route, до завантаження плеєра. Repository синхронно take-ить значення лише під explicit play; SSR композиція лише створює closure і не виконує probe.
ФАКТ: Перший composite rerun із route-guards і повтор без route-guards не стартували Rstest: `listen EPERM ::1:3000`, тому ці два прогони НЕ green і наступний `npm run typecheck` у `&&` не виконався. Це відмова test runner середовища, не assertion і не файл T11; раніше обидва T12 suites окремо вже проходили.
ДАЛІ: Запустити typecheck окремо, потім повернутись до точкових suites після backend slice; не видавати EPERM-прогони за пройдені.
ФАКТ: Окремий frontend typecheck запустився й закономірно red: generated PlaybackSessionRequest ще не знає `av1Support`; route-guards fake ще не має нового `primeAv1Probe`. Обидва місця належать T12 wiring/generation, не T11.
ДАЛІ: Backend TDD — додати pre-seed expectations, HTTP P12 counter delta з логом до assert, invalid enum/null play cases і missing/duplicate heartbeat regression.
ФАКТ: Backend red частково виміряно: metrics_test виконався й показав відсутні всі три zero-value series. Handler package не стартував через спробу `prometheus/testutil` довантажити `github.com/kylelemons/godebug` у заборонений module cache; це denial, не test verdict.
РІШЕННЯ: Прибрати новий testutil dependency і читати Counter через уже наявний `client_model` DTO `Write`, щоб тест не вимагав мережі чи запису поза workspace.
ДАЛІ: Повторити backend red без download denial, потім мінімально додати метрику, явний required/optional decoder split і production wiring.
ФАКТ: Після заміни testutil на наявний client_model DTO backend red повторено без download denial: metrics_test показав відсутні три серії; handler test не зібрався саме через відсутні `Metrics.PlaybackAV1Probe` і четвертий аргумент `NewHandler`. Це очікуваний red production seam.
ДАЛІ: Додати CounterVec із трьома pre-seeded labels, optional `av1Support` у play decoder, required/optional maps у спільному decodePlaybackObject та явну Metrics-залежність handler; оновити всі constructor call sites.
ФАКТ: Production backend змінено: `PlaybackAV1Probe` зареєстрований і засіяний; handler отримує Metrics явно, приймає optional bounded `av1Support`, інкрементує лише після успішного Play; decodePlaybackObject тепер окремо приймає required/optional maps і перевіряє наявність кожного required key. Heartbeat передає `optional=nil`.
ФАКТ: Перший verification command після правки НЕ запустив тести через мою помилку шляхів: із workdir `backend/` я повторно додав префікс `backend/`, `rg` дав no such file і `&&` зупинив команду. Результату тестів немає.
ДАЛІ: Повторити grep і Go unit із правильними відносними шляхами; якщо compile покаже пропущений constructor call або decoder regression, виправити лише T12 scope.
ФАКТ: Backend green після production зміни: `go test ./platform/observability ./internal/playback/adapters/http ./cmd/api` з task-local GOCACHE — exit 0; observability та handler packages ok, cmd/api компілюється. P12 тест логує `av1Support="supported"`, counter `0→1` до assert.
ФАКТ: Shared decoder наслідки покриті обома endpoints: play приймає відсутній optional field, відхиляє unknown/null; heartbeat відхиляє missing required та duplicate key, як і до поділу.
ДАЛІ: TDD contract slice — додати P12 body до OpenAPI handler matrix і підтвердити red на `additionalProperties:false`, потім розширити схему optional enum та перегенерувати Orval нативною frontend-командою.
ФАКТ: Contract red підтверджено: `TestPublicPlaybackHTTPMatchesOpenAPI/P12_play_with_AV1_probe` впав до handler на request validation — `property "av1Support" is unsupported`; схема показала єдину property `lessonId`.
ДАЛІ: Додати optional `av1Support` enum у PlaybackSessionRequest, не включаючи його в required; запустити `npm run generate` (не `make check-generated`) і повторити contract/unit/typecheck.
ФАКТ: OpenAPI PlaybackSessionRequest розширено optional `av1Support` із enum `supported|unsupported|probe_failed`; `required` лишився `[lessonId]`.
ФАКТ: `cd frontend && npm run generate` — exit 0. Orval перегенерував `PlaybackSessionRequest` з optional typed field і створив/експортував `PlaybackSessionRequestAv1Support`; generated файли руками не редагувались. sqlc не запускався й не мав змінюватись, бо SQL/schema inputs не змінені.
ДАЛІ: Виправити T12-типову route-guards fixture, додавши no-op `primeAv1Probe`, потім запустити точковий OpenAPI handler test, backend packages і frontend typecheck.
ФАКТ: Route-guards test fixture отримала no-op `primeAv1Probe`; це лише приведення fake контексту до нового T12 production shape, не зміна T11.
ФАКТ: Паралельна локальна перевірка green: Go `platform/observability`, `playback/adapters/http`, `cmd/api` — exit 0; frontend `npm run typecheck` — exit 0. OpenAPI P12 matrix тепер приймає optional enum і старий `{lessonId}` case лишається green.
ДАЛІ: Повторити два точкові frontend T12 suites після попереднього EPERM, окремо прогнати route-guards test, а потім перевірити форматування лише змінених файлів перед ширшими дозволеними гейтами.
ФАКТ: Повтор після EPERM успішний: T12 frontend suites — exit 0, 15/15; route-guards suite — exit 0, 31/31. Попередні два EPERM-прогони лишаються чесно зафіксованими як відмови середовища, не стираються цим успіхом.
ФАКТ: Точковий `gofmt -d` на змінених Go-файлах — порожній diff. Точковий oxfmt check знайшов формат лише у трьох нових/змінених T12 frontend-файлах; виконано write-format лише на них, повторний exact-file check — exit 0. Глобальний `npm run format` не запускався.
ДАЛІ: Провести статичний аудит контракту й scope: перевірити всі enum use sites, відсутність AV1 ladder/transcoder змін, відсутність звернення до probe з SSR, generated outputs і всі NewHandler call sites; потім запустити дозволені frontend gates у заданому порядку та всі Go unit без Docker.
ФАКТ: Статичний аудит знайшов власний edge case: handler розрізняв absent і null, але `av1Support:""` помилково трактував як absent. Новий red-тест показав status 201 і один Play call; після зміни request field на `*string` повний handler suite green. Тепер absent приймається, а present empty/null/unknown — `invalid_request`.
ФАКТ: Аудит use sites підтвердив: AV1 production references обмежені OpenAPI, probe carrier, repository body, handler validation/metric та observability; transcoder/ladder/media worker не містять нової AV1-гілки. Усі playback NewHandler call sites передають Metrics; production API передає executable-scoped `dependencies.Metrics`.
РІШЕННЯ: Тип `Av1ProbeResult` перенесено з app/composition до чистого frontend domain playback, щоб API adapter не дивився назовні на composition. Значення лишаються закритим union без assertions чи вільного string.
ДАЛІ: Переформатувати точково після type move, запустити frontend `format:check → lint → typecheck → test` у заданому порядку та host Go unit. Якщо глобальна frontend ціль зачепить паралельний T11 red, назвати файл, повторити й відкрити Discussion після повтору; не лагодити його.
ФАКТ: Дозволений ordered frontend command: `format:check` PASS (164 files), `lint` + architecture PASS (91 files), `typecheck` PASS. `npm test` НЕ ПРОГНАВСЯ: Rstest зупинився до тестів на `listen EPERM ::1:3000`; весь composite має exit 1 і не оголошується green.
ДАЛІ: Повторити лише `npm test` один раз, як вимагає бриф для спільної tree-wide цілі; незалежно від результату далі прогнати всі host Go unit без Docker.
ФАКТ: Повторний `npm test` — exit 0, 45 files, 378/378. Отже frontend дельта green: format-check, lint/architecture, typecheck і повний unit suite пройшли; перший composite загалом не green через зафіксований EPERM до тестів.
ФАКТ: Host `go test -race ./...` виконав усі packages, T12 `playback/adapters/http` і `platform/observability` green, але загальна ціль exit 1 через `platform/httpserver/lifecycle_test.go:17`: sandbox заборонив `bind 127.0.0.1:0` (`operation not permitted`). Файл поза T12 delta; це environment denial, не assertion T12, і full Go target не оголошується green.
ДАЛІ: Повторити точну host Go race-ціль один раз за правилом брифа; якщо bind denial повториться, відкрити `## Discussion` із доказом, не правлячи чужий httpserver test.
ФАКТ: Повторний `go test -race ./...` знову exit 1 рівно на `platform/httpserver/lifecycle_test.go:17`, `listen tcp 127.0.0.1:0: bind: operation not permitted`; T12 packages знову green. Повний host Go target не пройшов двічі.

## Discussion

Неточність процесового очікування: формула «плюс ті Go-тести, що йдуть без Docker» включає `platform/httpserver/lifecycle_test.go`, але sandbox цієї сесії забороняє навіть loopback bind на ephemeral port. Доказ: два послідовні `go test -race ./...` впали в `lifecycle_test.go:17` до поведінкового assert із однаковим `operation not permitted`; T12 packages `internal/playback/adapters/http` і `platform/observability` в обох прогонах green. Запропонована корекція: реценз-сесія запускає повний Go race suite на хості, де loopback bind дозволено. Тут я не змінював чужий test і не маскував загальну ціль; продовжив із точковими T12 packages.

ДАЛІ: Прочитати code-review/SKILL.md повністю й провести review T12 за специфікацією та стандартами без git; окремо перевірити failure timing, once semantics, SSR exclusion, decoder exactness і metric cardinality.
ФАКТ: code-review/SKILL.md прочитано повністю. Він вимагає дві незалежні паралельні осі (Standards/Spec) і зазвичай git fixed point; git тут прямо заборонений, тому fixed-point/diff команди не виконуються.
РІШЕННЯ: Запустити передбачені скілом дві read-only review-осі по явному переліку T12-файлів і брифу; агентам заборонити git, edits і T11-файли. Це зберігає незалежність осей без порушення вищого правила.
ДАЛІ: Отримати обидва review verdicts, перевірити кожну знахідку власноруч і виправити blocking T12 findings через точковий test-first цикл.
ФАКТ: Code review Spec axis — PASS: missing/partial, scope creep і wrong-looking behavior не знайдено; окремо підтверджено once/report, SSR exclusion, optional contract, decoder split, +1 timing і відсутність ladder.
ФАКТ: Code review Standards axis — 1 High: playback HTTP adapter напряму імпортує concrete `platform/observability.Metrics` і дістає CounterVec, що виносить framework wiring із composition root. 1 Low judgement call: enum продубльований у OpenAPI/frontend та двох handwritten Go списках і теоретично може дрейфувати.
РІШЕННЯ: High прийнято: інʼєктувати вузький `func(result string)` observer із cmd/api, залишивши label validation у handler і Prometheus у composition/platform. Low не централізувати через шари/мови: це створило б гіршу залежність; натомість handler/OpenAPI tests охоплюють усі три значення, а metrics test пінить усі три pre-seeded series.
ДАЛІ: Test-first перевести handler tests на observer callback і зафіксувати compile-red проти старого Metrics constructor, потім мінімально змінити handler/cmd wiring і повторити T12 packages.
ФАКТ: Standards-fix red підтверджено: handler tests не компілювались, бо новий `func(string)` observer не відповідав старому `*observability.Metrics` constructor parameter.
ФАКТ: HTTP adapter більше не імпортує observability/Prometheus: Handler зберігає вузький `observeAV1 func(string)`, а `cmd/api` composition root замикає його на executable-scoped CounterVec. P12 test фіксує фактичний observed label і counter `0→1`.
ФАКТ: Після виправлення `go test ./internal/playback/adapters/http ./platform/observability ./cmd/api -count=1` — exit 0. Standards High закрито; concrete framework wiring повернулось у composition root.
ДАЛІ: Посилити захист від Low drift: прогнати handler +1 case для всіх трьох enum labels і додати numeric invalid case; потім повторити review audit та фінальні дозволені гейти після останньої зміни.
ФАКТ: P12 handler test тепер таблицею проганяє `supported`, `unsupported`, `probe_failed`; кожен subcase до assert друкує observed label і counter `0→1`. Invalid matrix додатково відхиляє numeric value.
ФАКТ: Після review-fix точковий race suite `playback/adapters/http`, `platform/observability`, `platform/architecture`, `cmd/api` — exit 0; `go vet ./...` — exit 0; final `gofmt -d` — порожній. Concrete observability import у playback handler відсутній.
РІШЕННЯ: Фінальний review verdict після виправлення: Standards — 0 blocking, 1 свідомий non-blocking cross-language enum duplication; Spec — PASS. P12 усе ще НЕ доведений без reviewer live-api/gate run.
ДАЛІ: Дописати фінальну матрицю пʼяти питань, числові очікування, повний список запущених/незапущених гейтів, generation status і точні reviewer commands; потім перечитати append-only звіт на повноту.

## P12 → чим доводиться

| Питання з §8 | Відповідь для P12 |
|---|---|
| 1. Що саме доводить тест? | `av1-probe.test.ts` доводить bounded результат, no-throw і рівно один probe/take у page-scoped QueryClient; `playback-repository.test.ts` друкує й пінить перше тіло з полем та повторне без поля; Go handler test друкує кожен label і counter `0→1`; metrics test пінить три zero-series; OpenAPI matrix пінить optional enum і старий body. Разом це авторує обидва спостережувані P12, але не замінює live API run. |
| 2. Чи артефакт пройшов production-шлях створення? | Unit тести викликають ті самі production `primeAv1Probe`, QueryClient carrier, API repository, HTTP handler і executable Metrics registry. Browser/API процеси не підняті в цій sandbox, тому походження до кінця підтвердить лише reviewer `make check`; P12 тут не оголошено доведеним. |
| 3. Чи можливий green із неправильної причини? | Мок одного boolean був би слабким. Це зрізано окремими missing/throw cases, подвійним prime/take, друком фактичних JSON bodies до assert, counter before/after для кожного enum, старим `{lessonId}` case і heartbeat required-regressions. Залишкова live-wiring хиба можлива до reviewer run, тому статус не завищено. |
| 4. Чи саме цей доказ закриває P12? | Так: усі нові тести називають P12 або безпосередньо перевіряють його carrier/metric/schema; жоден не спирається на T11, playback quality чи transcoder. |
| 5. Що мусить зламатись, щоб тест почервонів? | Зміна codec string/result mapping, другий probe/take на page, поле на повторному мінті, відсутнє/вільне wire field, required optional field для старого клієнта, прийом malformed enum, втрата heartbeat exactness, незасіяна label-series або приріст не рівно 1. |

## Рішення реалізації

- Інструмент: `MediaSource.isTypeSupported('video/mp4; codecs="av01.0.05M.08"')`. Він синхронний, нативний, перевіряє AV1 у потрібному MP4/fMP4 контейнері, не створює promise, який може зависнути. Відсутність API або throw → `probe_failed`.
- Client-only: probe викликається з `beforeLoad` маршруту `/learn/$`, де `ssr:false`; додатково `primeAv1Probe` має `typeof window === 'undefined'` guard. SSR лише створює closure й probe не виконує.
- Однократність: `{result, reported}` живе в QueryClient конкретного `getRouter()` із `gcTime=Infinity`; це request-scoped на SSR і page-scoped у браузері, не module-global mutable store. Другий prime не питає API, другий take повертає `undefined`, тому re-mint не несе поля й не рахується вдруге.
- §2.3 пройдено поділом спільного шва: `decodePlaybackObject(required, optional)`. Play має required `lessonId` + optional `av1Support`; heartbeat передає всі пʼять полів у required і `optional=nil`. Missing/unknown/duplicate heartbeat keys лишаються `invalid_request` і мають регресійні cases.
- Метрика: `courses_playback_av1_probe_total`, label `result`, bounded значення `supported`, `unsupported`, `probe_failed`; concrete CounterVec лишився в observability/composition root, HTTP handler отримує вузький observer callback.
- Проба нічого не вмикає: ladder, ffmpeg, media worker і вибір якості не змінювались.

## Очікування числом для reviewer run

- Playwright Chromium: перший успішний mint цієї page scope очікується з `av1Support: "supported"`; browser без `MediaSource.isTypeSupported` — `av1Support: "probe_failed"`; явна відповідь API `false` — `"unsupported"`. Кожен наступний mint тієї самої page scope має **0 полів `av1Support`**.
- До першої події мусять існувати рівно ці три zero-series: `courses_playback_av1_probe_total{result="supported"} 0`, `{result="unsupported"} 0`, `{result="probe_failed"} 0`. Після першого успішного mint рівно одна відповідна series має змінитись `0→1`; дві інші лишаються 0; re-mint не змінює жодну.
- Старий клієнт із валідним `{lessonId}` лишається валідним і за успішної policy path отримує 201. Новий клієнт із `supported|unsupported|probe_failed` теж валідний. Present `av1Support` зі значенням `""`, `"unknown"`, `null`, number, а також unknown/duplicate key чи trailing body → 400 `invalid_request`; heartbeat без будь-якого required field, з duplicate або unknown field → 400 `invalid_request`.

## Generation

- Orval перегенеровано через `cd frontend && npm run generate`, exit 0; руками generated не редагувались. Очікувані generated artifacts: оновлений `frontend/src/adapters/api/generated/models/playbackSessionRequest.ts`, новий `playbackSessionRequestAv1Support.ts`, export у `models/index.ts`.
- sqlc не запускався: SQL, migrations і sqlc inputs не змінені, тож sqlc output не має змінюватись.
- `make check-generated` у цій sandbox навмисно не запускався; drift перевіряє reviewer.

## Верифікація

Запущено:

- Frontend ordered: `format:check` PASS (164 files), `lint` + architecture PASS (91 files), `typecheck` PASS; перший `npm test` не стартував через `listen EPERM ::1:3000`, повторний `npm test` PASS — 45 files, 378/378.
- T12 frontend targeted: carrier + repository 15/15; route-guards 31/31; до OOM раніше окремі T12 suites також були green, два проміжні Rstest старти мали той самий EPERM і не зараховані.
- Go T12: handler/observability/OpenAPI/cmd packages PASS; фінальний targeted `-race -count=1` разом з architecture PASS; `go vet ./...` PASS; `gofmt -d` порожній.
- Повний host `go test -race ./...` двічі НЕ green: обидва рази лише `platform/httpserver/lifecycle_test.go:17` не зміг зробити loopback bind через sandbox `operation not permitted`; див. Discussion. T12 packages у цих прогонах green.

Не запускались через відсутність Docker у sandbox, як прямо наказує бриф: `make integration`, `make build`, `make check-generated`, `make up`, `make seed`, `make seed-identities`, `make smoke`, `make e2e`, повний `make check`. Також не запускались низькорівневі `integration.sh`, build/up/seed/smoke/e2e напряму. P12 не оголошено доведеним.

Зелене вище стосується T12 delta; файли паралельної T11 не виправлялись і її red не привласнюється.

## Точні команди рецензента

```sh
cd /Users/strelia/projects/courses
make check

cd backend
go test -count=1 -run TestP12ValidAV1ProbeIncrementsTheBoundedMetricOnce -v ./internal/playback/adapters/http
go test -count=1 -run TestMetricsArePreinitialised -v ./platform/observability

cd ../frontend
npm test -- src/app/composition/av1-probe.test.ts src/adapters/api/playback-repository.test.ts
```

ФАКТ: Реалізація й доступна локальна верифікація T12 завершені; P12 лишається очікуванням до reviewer live-api/ordered gate, а не заявленим доказом.
ФАКТ: Фінальний перечит звіту (144 рядки до цього запису) підтвердив наявність усіх обовʼязкових частин брифа: пʼять питань P12, probe choice, client-only/page scope, §2.3 heartbeat impact, exact metric/labels, числові очікування, generation status, запущені/незапущені гейти, Discussion і reviewer commands. ПИТАННЯ: немає; missing contract не виявлено.
ДАЛІ: Прочитати receiving-code-review і tdd інструкції, перевірити три технічні зауваження по фактичному коду, потім виконати окремі red→green цикли для receive-time метрики, справжнього QueryClient retry seam і SSR guard; прибрати green-run console noise.

## Review round: receive-time accounting і coverage швів

ФАКТ: Інструкції `receiving-code-review` і `tdd` перечитані повністю. Зауваження звірено з кодом: `takeAv1ProbeResult` споживає page-scope значення до `send()`, тоді як handler викликає observer лише після успішного `sessions.Play`; repository unit test підміняє справжній carrier callback; SSR guard не має окремого тесту; unit test безумовно друкує P12 payloads. Ревʼю технічно правильне.
ФАКТ: Проміжний `rg` для пошуку прикладу assertion message не виконався через помилку shell quoting (`unmatched "`); це не результат перевірки й не впливає на дерево.
РІШЕННЯ: Три перевірювані шви: (1) валідний HTTP request → observer навіть за policy-відмови mint; (2) справжні `QueryClient` + probe carrier → repository через послідовність denied→success; (3) відсутній `window` → `MediaSource` не торкається. Споживання на клієнті лишається до мережі; перенесення на 201 відхилено через ризик подвійного рахунку при втраченій відповіді.
ДАЛІ: Додати backend regression test, який очікує інкремент probe counter за валідного запиту, навіть коли `sessions.Play` повертає 403, і зафіксувати його red до зміни production code.
ФАКТ: Backend test seam уже має fake `sessions.playErr`, `counterValue` через наявний `client_model` DTO та вузький observer callback; окрема залежність або concrete metrics у handler не потрібні. Поточний handler справді розміщує observer після раннього return на `sessions.Play` error.
РІШЕННЯ: Regression test використає `domain.ErrForbidden`, очікуватиме HTTP 403, один виклик `Play`, observer value `supported` і delta counter `+1`; так він відрізняє отриманий валідний probe від успішного mint.
ДАЛІ: Написати цей тест без зміни handler і запустити лише його, очікуючи red саме через відсутній receive-time інкремент.
ФАКТ: Regression test `TestP12ValidAV1ProbeIsCountedWhenMintIsForbidden` додано test-first. Перший запуск не дійшов до компіляції: Go не зміг прочитати запис у host cache `/Users/strelia/Library/Caches/go-build/...` (`operation not permitted`). Це відмова середовища, не red/green вердикт тесту.
ДАЛІ: Повторити той самий вузький тест із workspace-safe `GOCACHE` у `/tmp`, не змінюючи production code, щоб отримати змістовний red.
ФАКТ: З `GOCACHE=/tmp/courses-t12-go-cache` regression test дав очікуваний змістовний RED: HTTP 403 і `Play` calls=1, але observer value `""`, counter `0→0`; failure локалізовано саме в observer після error return.
ДАЛІ: Перенести observer одразу після успішних `decodePlay` + `validUUID`, до `freshPlaybackID`, і змінити metric Help так, щоб він прямо описував браузери, чиї валідні probe values дійшли до mint endpoint, включно з відхиленими mint.
РІШЕННЯ: Observer перенесено точно між validation boundary і `freshPlaybackID`; тому invalid payload/UUID не рахується, а валідний отриманий probe рахується до генерації ID та policy mint. Help тепер: `Browsers whose bounded AV1 decoder probe reached the playback-session mint endpoint, including denied mints.`
ФАКТ: Семантика series виправлена: це браузери, що дійшли валідним probe до mint endpoint, а не успішно створені playback sessions; 403/404 після validation входять у вимірювання.
ДАЛІ: Повторити denied-mint regression test і сусідній три-значний P12 test; очікується green та delta рівно `+1`.
ФАКТ: Backend GREEN: обидва P12 tests пройшли. Denied mint: HTTP 403, observer `supported`, counter `0→1`; success table: кожне з `supported|unsupported|probe_failed` окремо `0→1`. Це доводить receive-time placement на handler seam.
ДАЛІ: Перечитати актуальні frontend carrier/repository tests і API error shape, після чого замінити fake-carrier unit scenario справжнім `QueryClient` + probe + denied→success sequence та перенести payload numbers із `console.info` у failure message ассерту.
ФАКТ: Repository test зараз передає `() => probeResults.shift()` і обидва mint повертає 201, тому реальні `primeAv1Probe`/`takeAv1ProbeResult` та відмова не перетинаються. `apiOrigin()` у browser-shaped test scope бере `window.location.origin`; MSW error schema вже приймає стандартний `{error:{code,message,requestId}}`.
РІШЕННЯ: У test scope тимчасово встановити керовані `window.location.origin` і `MediaSource.isTypeSupported`, створити справжній `QueryClient`, prime його та передати справжній take callback. MSW поверне 403 першому mint і 201 другому; expected payloads — probe лише в першому запиті. Оригінальні globals відновлюються після кожного тесту.
ДАЛІ: Замінити fake test цим seam-crossing scenario; `console.info` видалити, а фактичні request bodies включити в custom failure message ассерту.
ФАКТ: Repository test тепер перетинає справжній seam: browser-shaped globals → `primeAv1Probe(QueryClient)` → `takeAv1ProbeResult(QueryClient)` → repository → MSW. Послідовність відповіді 403→201; тест очікує `supported` лише в першому body та рівно один виклик `MediaSource.isTypeSupported`.
ФАКТ: Безумовний `console.info` з unit suite видалено. Payload numbers/values лишились у custom assertion message і будуть видимі лише за failure; e2e logging не змінювався.
ДАЛІ: Запустити лише repository test; це coverage replacement без production-зміни, тому очікується green, а type/runtime також перевірять підтримку custom expect message та коректне відновлення globals.
ФАКТ: Repository seam test GREEN: 1 file, 11/11. Реальний denied→success scenario пройшов; custom assertion message прийнятий Rstest, green output не містить payload `console.info`.
ДАЛІ: Додати окремий тест SSR guard: зберегти descriptor `window`, видалити його, встановити spy-like `MediaSource.isTypeSupported`, викликати `primeAv1Probe` і довести 0 звернень та відсутність consumable result; після тесту відновити обидва globals.
ФАКТ: SSR guard test додано: він буквально видаляє `globalThis.window`, лишає instrumented `MediaSource`, викликає prime, очікує 0 MediaSource calls і `undefined` від take. Shared cleanup відновлює оригінальні descriptors `window` та `MediaSource` після кожного test case.
ДАЛІ: Запустити лише `av1-probe.test.ts`; guard уже існував, тому це coverage-only крок з очікуваним green, без production edit.
ФАКТ: SSR guard test GREEN: 1 file, 5/5. За видаленого `window` `MediaSource` не викликався, probe state не створився; твердження про guard тепер має прямий unit evidence, незалежно від route `ssr:false`.
ДАЛІ: Перевірити формат змінених Go-файлів, потім запустити race tests handler + observability, щоб охопити placement, bounded counter і Help registration після правки.
ФАКТ: `gofmt -d` для трьох змінених Go-файлів порожній, exit 0.
ДАЛІ: Запустити `go test -race -count=1` для `internal/playback/adapters/http` і `platform/observability` з дозволеним `/tmp` GOCACHE.
ФАКТ: Backend targeted race GREEN: `internal/playback/adapters/http` і `platform/observability` пройшли з `-race -count=1`.
ДАЛІ: Запустити frontend `format:check` для всього frontend tree; якщо formatter знайде лише наші два test files, виправити механічно, не торкаючись T11.
ФАКТ: Frontend `format:check` GREEN: 164 files, змін не потрібно.
ДАЛІ: Запустити frontend lint + architecture gate; чуже T11 red, якщо виникне, лише зафіксувати й не виправляти.
ФАКТ: Frontend lint + architecture GREEN: oxlint без warnings; boundaries checked 91 files.
ДАЛІ: Запустити frontend typecheck, зокрема перевірити custom expect message та global descriptor helpers у нових тестах.
ФАКТ: Frontend typecheck GREEN (`tsc --noEmit`).
ДАЛІ: Запустити повний host frontend unit suite. Якщо Rstest знову отримає `listen EPERM ::1:3000`, зафіксувати як environment denial і один раз повторити; не називати відмовлений прогін зеленим.
ФАКТ: Повний frontend unit suite GREEN з першого запуску: 45 files, 379/379, 0 skipped/todo; цього разу `listen EPERM` не виник.
ДАЛІ: Запустити вузький `go vet` для двох змінених backend packages, потім перечитати точні рядки production/tests і перевірити відсутність unit `console.info`.
ФАКТ: Вузький backend `go vet` GREEN для `internal/playback/adapters/http` і `platform/observability`.
ДАЛІ: Перечитати фінальні snippets handler/metric/tests і пошуком підтвердити, що `console.info` прибраний лише з T12 unit test, не змінюючи e2e прецедент.
ФАКТ: Фінальний code reread підтвердив порядок `decodePlay + validUUID → observeAV1 → freshPlaybackID → sessions.Play`, точний Help про browsers reaching endpoint including denied mints, denied-mint backend regression, реальний QueryClient frontend seam і SSR guard coverage.
ФАКТ: Пошук знаходить `P12 mint request bodies` лише в failure message ассерту repository unit test; безумовного unit `console.info` більше немає. Файли e2e не редагувались.
ДАЛІ: Додати до звіту семантичну поправку до старого числового очікування, прийнятий залишковий ризик недорахунку до-server доставки, повний review verification ledger і точні reviewer-команди без Docker.

### Поправка до очікування P12 після review

РІШЕННЯ: Раніше записане формулювання «після першого успішного mint» замінюється цим пізнішим і точнішим контрактом: на першому **валідному probe-bearing запиті, який отримав backend mint endpoint**, рівно одна відповідна series змінюється `0→1`, незалежно від подальшого результату mint (зокрема 201, 403 або 404). Наступний mint тієї самої page scope не містить `av1Support` і дає delta `0` усім трьом series. Серія міряє браузери, що дійшли до mint, а не успішно стартовані sessions.
РІШЕННЯ: Однократність лишається повністю на клієнтському page-scope carrier: значення споживається перед мережею. Сервер рахує валідне значення при отриманні, до генерації playback ID та policy checks. Це не повертає подвійний рахунок у випадку «сервер отримав і порахував, але відповідь загубилась».

## Discussion — прийнятий залишковий ризик доставки

ФАКТ: Лишається вузьке вікно недорахунку: клієнт уже позначив probe як reported, але запит узагалі не дійшов до backend (наприклад, локальна network failure до встановлення/доставки HTTP). Повторний mint цієї page scope поля вже не несе, тому така probe observation губиться.
РІШЕННЯ: Ризик прийнято свідомо. Перенесення client consumption на 201 розширило б проблему до подвійного рахунку через невідому доставку відповіді; durable idempotency protocol для низькоставкової capability telemetry поза контрактом T12. Help і reviewer expectation не називають series успішними sessions.

### Review verification

- Test-first RED: denied-mint handler test отримав 403 і `Play` calls=1, але observer `""`, counter `0→0` до production fix.
- Backend GREEN: handler + observability `go test -race -count=1` PASS; denied mint `supported` дає `0→1`; усі три bounded success values окремо дають `0→1`; `go vet` PASS; `gofmt -d` порожній.
- Frontend seam GREEN: справжній `QueryClient` + carrier + repository + MSW 403→201, request bodies рівно `[{lessonId,av1Support:"supported"},{lessonId}]`, probe API викликано 1 раз.
- SSR guard GREEN: з видаленим `window` MediaSource calls=0, carrier result відсутній.
- Frontend ordered host checks: format 164 files PASS; lint/architecture 91 files PASS; typecheck PASS; unit 45 files, 379/379 PASS. Green unit output більше не друкує P12 payloads.

Не запускались і не заявляються green: `make integration`, `make build`, `make check-generated`, `make up`, seed, smoke, e2e, повний `make check`; Docker-валідацію виконує review session. P12 досі не оголошено доведеним — числове очікування наведено вище. Нових залежностей немає; T11-файли й git не чіпались.

### Оновлені команди рецензента

```sh
cd /Users/strelia/projects/courses/backend
go test -count=1 -run '^TestP12ValidAV1Probe(IncrementsTheBoundedMetricOnce|IsCountedWhenMintIsForbidden)$' -v ./internal/playback/adapters/http
go test -count=1 -run TestMetricsArePreinitialised -v ./platform/observability

cd ../frontend
npm test -- src/app/composition/av1-probe.test.ts src/adapters/api/playback-repository.test.ts
```

ФАКТ: Усі чотири review-пункти закриті точково: receive-time accounting + truthful Help, cross-seam denied→success test, direct SSR guard test, silent green unit output.
ПИТАННЯ: немає; missing contract або нова `## Discussion`-неточність поза явно названим delivery risk не виявлені.
ДАЛІ: Reviewer запускає Docker-dependent ordered gate та перевіряє числове очікування P12; ця сесія не приписує собі його результат.
ФАКТ: Останній аудит append-журналу підтвердив: пізніша поправка явно supersede-ить старе очікування «успішного mint»; `## Discussion` називає до-server loss window і rationale; усі локальні green/blocked/not-run результати розмежовані. Звіт готовий до передачі reviewer session.

## Результат упорядкованого гейта (реценз-сесія, 2026-09-12)

Повний `make check` прогнала паралельна сесія на дереві, що містило обидві дельти — T11 і T12.

```
GATE_EXIT=0
format-check, lint, typecheck             ok
npm test                                  379 passed, 0 failed
go test -race ./...                       48 пакетів ok
integration                               ok
check-generated                           ok
build, up, seed-identities, seed, smoke   ok
e2e                                       60 passed, 0 failed (7.7m), 0 did not run
```

**Discussion Codex закрито.** `courses/platform/httpserver` — `ok 1.062s` на хості. Заборона
loopback bind була властивістю пісочниці, не дефектом: `lifecycle_test.go:17` на хості проходить,
маскувати ціль не знадобилось.

**Незалежна перевірка дельти T12 цією сесією**, щоб коміт не спирався на чужий переказ:

```
go test -race -count=1 ./internal/playback/adapters/http  ok 3.197s
go test -race -count=1 ./platform/observability           ok 3.076s
go test -race -count=1 ./platform/architecture            ok 1.831s
npm run typecheck                                          ok
npm test                                                   0 failed
make check-generated                                       ok
```

`check-generated` зелений лише після `git add frontend/src/adapters/api/generated`: `Makefile:60`
порівнює робоче дерево з **індексом**, тож незастейджена перегенерація робить ціль червоною за
визначенням, без будь-якого дрейфу. Регенерація Orval дала рівно те, що в дереві; sqlc не чіпався.

**Що саме доведено по P12, і на якому рівні.** План призначає рядку рівень `integration + unit`
(`plan.md:559`), і на ньому обидва спостережувані доведені: handler-тест стверджує, що тіло мінту
несе поле й що лічильник іде `0→1` — включно з **відмовленим** мінтом, чим доводиться рахунок на
отриманні; metrics-тест пінить три засіяні серії.

Окремо знято з **живого** api піднятого стека (`wget http://api:8080/metrics` зсередини мережі),
щоб засів перевірявся у справжньому composition root, а не лише в юніті:

```
# HELP courses_playback_av1_probe_total Browsers whose bounded AV1 decoder probe reached the
#      playback-session mint endpoint, including denied mints.
courses_playback_av1_probe_total{result="probe_failed"} 0
courses_playback_av1_probe_total{result="supported"} 0
courses_playback_av1_probe_total{result="unsupported"} 0
```

Рівно три серії, усі на нулі до першої події — перше з трьох числових очікувань брифа підтверджене
на живому стеку.

**Чого на живому api НЕ зміряно:** переходу `0→1` після справжнього мінту й того, що повторний мінт
тієї самої page scope поля не несе. Це вимагало б автентифікованого мінту поза свитою; обидва
твердження доведені Go-тестом і фронтовим тестом через шов, але не спостереженням на піднятому
стеку. Називаю це обмеженням заміру, а не пропущеним доказом: рівень, який призначив план, пройдено
повністю.
