# S0.1b-3b — const-set single-source (§16.9 Q3 + T2) — WORK FILE

Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append). Перший запис — `ДАЛІ:` до першого інструменту.
ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінал = стислий переказ. Base: s0.0-spike, HEAD у brief.

## Журнал
ДАЛІ: прочитати briefing, §16.9 та локальний workflow; звірити branch/HEAD і реалізацію const-set перед змінами.
ФАКТ: прочитано briefing та AI Factory adapter; звіряю snapshot і точний план зрізу.
ФАКТ: HEAD 671a442; наявні чужі WIP/plan/tasks зміни зберігаю; чинний план — s0-openapi-gen-plan.md. Hook конфіг знайдено.
ФАКТ: s0.0-spike @671a442 підтверджено. NewCourse дублює повний status-list; Block validator має семантичний switch, який треба зберегти fail-closed.
ФАКТ: §16.9 прочитано: видалення overlay-дублікатів + fidelity + byte-identical в одному коміті. Генератор має Enums і FieldRefs, дозволяючи зберегти published-only окремо.
РІШЕННЯ: small, READY (Codex, 2026-10-27); acceptance — три авторитетні набори, census та fidelity; review передає власник Courses 2. Impact: domain query/course/authoring/document, course HTTP contribution, overlay, contract tests. BlockType досі не подавався production contribution (document route ще не генерується): реєструю Enums і перевіряю scalar emission без розширення route scope.
ДАЛІ: використати typed domain.Status у двох DTO, наявний Enums registry та typed→string mapping у contribution; BlockTypes membership guard збереже семантичний switch. Census через go/types охопить const-декларації, public sorts матимуть явні admin-only виключення.
ФАКТ: додано PublicCatalogSorts/CourseStatuses/BlockTypes; NewQuery, NewCourse, status filter і block membership споживають набори. DTO status типізовано без зміни JSON; contribution Enums живиться domain. Видалено обидва full-status overlay списки, published-only збережено.
ФАКТ: додано typed-const census для трьох наборів із синтетичною новою const-декларацією; equality emitted↔domain для status request/response, sort, BlockType; окремий published-only assertion та mutation-тест єдиного PublicCatalogSorts для NewQuery/contribution.
ДАЛІ: компіляційна перевірка нових тестів і fidelity, потім точні ordered gates.
ФАКТ: перший targeted запуск із root exit 1: go: cannot find main module (go.mod у backend). Повторюю з backend.
ФАКТ: targeted запуск з backend заблокований sandbox: Go build cache operation not permitted, exit 1; повтор із доступом до cache.
ФАКТ: targeted exit 1 — go/packages не є наявною test-залежністю (missing go.sum). Додаткової залежності не вводжу: census читає AST production const-блоків і типізує їх stdlib go/types.
ФАКТ: census охоплює всі non-test Go-файли domain (не лише три поточні), щоб новий const у новому файлі теж був видимий.
ФАКТ: targeted run exit 1: census і PublicSort PASS; fidelity PASS (109 error codes, два генерування byte-identical); scalar BlockType probe FAIL через непотрібний NullablePrice→Price ref у повному Components registry. Пробу звужено до production Enums registry (без unrelated Components).

ФАКТ: targeted tests exit=0
```text
=== RUN   TestDomainConstSetCensus
=== RUN   TestDomainConstSetCensus/Sort
=== RUN   TestDomainConstSetCensus/Status
=== RUN   TestDomainConstSetCensus/BlockType
--- PASS: TestDomainConstSetCensus (0.01s)
    --- PASS: TestDomainConstSetCensus/Sort (0.00s)
    --- PASS: TestDomainConstSetCensus/Status (0.00s)
    --- PASS: TestDomainConstSetCensus/BlockType (0.00s)
=== RUN   TestConstSetsEmittedAndPublishedNarrowing
--- PASS: TestConstSetsEmittedAndPublishedNarrowing (0.10s)
=== RUN   TestPublicSortSingleSource
--- PASS: TestPublicSortSingleSource (0.00s)
=== RUN   TestCatalogOpenAPIFidelity
    openapi_contract_test.go:55: error enum ref-closure mutation detected; 109 emitted codes compared
    openapi_contract_test.go:84: catalog and status request + complete referenced component closures equal hand spec; two generations byte-identical
--- PASS: TestCatalogOpenAPIFidelity (0.28s)
PASS
ok  	courses/internal/course/adapters/http	1.903s

```

ДАЛІ: gate `gofmt -l .` (cwd backend).
ФАКТ: `gofmt -l .` exit=0
```text
(empty output)
```

ДАЛІ: gate `go vet ./...` (cwd backend).
ФАКТ: `go vet ./...` exit=0
```text
(empty output)
```

ДАЛІ: gate `go build ./...` (cwd backend).
ФАКТ: `go build ./...` exit=0
```text
(empty output)
```

ДАЛІ: gate `go test -count=1 ./...` (cwd backend).
ФАКТ: AI Factory checkpoint OK: 5f5782c85e26ce6251b42c64f5b19b99-1-f338d1ddc3a0. Targeted тести PASS exit 0; ordered gates запущено.
ФАКТ: diff перевірено; openapi.yaml і openapi.generated.yaml не змінені; лише full-status overlay структури вилучено. gofmt/vet/build exit 0, повний go test ще виконується.
ФАКТ: `go test -count=1 ./...` exit=0
```text
?   	courses/cmd/api	[no test files]
ok  	courses/cmd/asr-worker	1.602s
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]
ok  	courses/internal/adminstats/adapters/http	2.482s
?   	courses/internal/adminstats/adapters/postgres	[no test files]
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	1.546s
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	2.246s
ok  	courses/internal/clientlog/domain	2.119s
ok  	courses/internal/contracts/access	2.367s
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	2.682s
ok  	courses/internal/course/adapters/http	16.183s
ok  	courses/internal/course/adapters/postgres	2.676s
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	5.848s
ok  	courses/internal/course/domain	2.891s
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	15.438s
?   	courses/internal/discussion/adapters/postgres	[no test files]
ok  	courses/internal/discussion/application	3.480s
ok  	courses/internal/discussion/domain	3.779s
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	4.113s
ok  	courses/internal/identity/adapters/http	15.270s
ok  	courses/internal/identity/adapters/mail	4.252s
?   	courses/internal/identity/adapters/postgres	[no test files]
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	4.490s
ok  	courses/internal/identity/application	4.516s
ok  	courses/internal/identity/domain	3.981s
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	3.664s
ok  	courses/internal/image/adapters/http	3.996s
?   	courses/internal/image/adapters/postgres	[no test files]
ok  	courses/internal/image/application	3.370s
ok  	courses/internal/image/domain	2.910s
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	5.151s
ok  	courses/internal/media/adapters/ffmpeg	3.735s
ok  	courses/internal/media/adapters/filesystem	4.123s
ok  	courses/internal/media/adapters/http	13.933s
?   	courses/internal/media/adapters/postgres	[no test files]
?   	courses/internal/media/adapters/postgres/generated	[no test files]
?   	courses/internal/media/adapters/tus	[no test files]
ok  	courses/internal/media/application	3.923s
ok  	courses/internal/media/domain	3.605s
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	9.983s
?   	courses/internal/orders/adapters/postgres	[no test files]
ok  	courses/internal/orders/adapters/wayforpay	1.166s
ok  	courses/internal/orders/application	1.338s
ok  	courses/internal/orders/domain	1.773s
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	6.068s
ok  	courses/internal/playback/adapters/http	6.121s
ok  	courses/internal/playback/adapters/postgres	4.204s
ok  	courses/internal/playback/adapters/redis	4.086s
ok  	courses/internal/playback/application	3.864s
ok  	courses/internal/playback/domain	4.137s
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	2.938s
ok  	courses/internal/testsupport/coursevisibility	1.873s
ok  	courses/internal/transcript/adapters/asynq	1.381s
ok  	courses/internal/transcript/adapters/http	2.290s
ok  	courses/internal/transcript/adapters/media	2.086s
?   	courses/internal/transcript/adapters/postgres	[no test files]
ok  	courses/internal/transcript/adapters/whisper	7.682s
ok  	courses/internal/transcript/application	3.235s
ok  	courses/internal/transcript/domain	2.174s
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	1.584s
ok  	courses/platform/config	2.986s
ok  	courses/platform/httpserver	3.944s
ok  	courses/platform/jsonbody	3.837s
ok  	courses/platform/mail	3.722s
ok  	courses/platform/observability	4.516s
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]
```
ДАЛІ: очікую завершення повного Go gate; перевіряю census на неповні припущення про форму const-декларацій.
РІШЕННЯ: census додатково типізує всі const-декларації та scalar types, включно з inferred conversion/alias; не фільтрує лише явно typed const, щоб уникнути тихого пропуску. Непідтримана залежність дає typecheck FAIL. Після цього test-only delta ordered gates буде повторено.

ДАЛІ: final gate `gofmt -l .` (cwd backend, census delta).
ФАКТ: `gofmt -l .` exit=0
```text
(empty output)
```

ДАЛІ: final gate `go vet ./...` (cwd backend, census delta).
ФАКТ: `go vet ./...` exit=0
```text
(empty output)
```

ДАЛІ: final gate `go build ./...` (cwd backend, census delta).
ФАКТ: перший повний ordered run завершено exit 0; остаточний census delta перевіряється другим ordered run. Незалежний review не заявляється: його виконає Courses 2 через власника. E2E не запускався за явною вказівкою.
ФАКТ: `go build ./...` exit=0
```text
(empty output)
```

ДАЛІ: final gate `go test -count=1 ./...` (cwd backend, census delta).
РІШЕННЯ: 2026-10-27, Codex, S0.1b-3b — повні набори status і BlockType є domain-owned; public sort subset має явні admin-only виключення SortStudents/SortPrice у census. Перевірка: TestDomainConstSetCensus + TestConstSetsEmittedAndPublishedNarrowing + TestPublicSortSingleSource.
ФАКТ: BlockType production Enums registry підключено й scalar emission перевірено; LessonDocument/Block routes поки не входять у generated artifact (поточний catalog/request scope). Семантичний block switch з default reject збережено: membership не замінює перевірки полів paragraph/heading/list.
ФАКТ: hand SHA256=9106bb8adc4fcdee02be4968f0c11e20f6c62345620e4a8719ef30fce8151b62; generated SHA256=9351ef96eb00a93de4e35f0ffd1ac152fe2c756a7be92ee3bf92661c585f8bf1. Обидва файли незмінні щодо HEAD; git diff --check exit 0.
ДАЛІ: після фінального gate зробити scoped commit із правдивими Codex trailers; чужі WIP/plan/tasks не включати. Передати власнику для Courses 2, без merge/push.
ФАКТ: `go test -count=1 ./...` exit=0
```text
?   	courses/cmd/api	[no test files]
ok  	courses/cmd/asr-worker	1.807s
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]
ok  	courses/internal/adminstats/adapters/http	2.642s
?   	courses/internal/adminstats/adapters/postgres	[no test files]
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	1.250s
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	1.972s
ok  	courses/internal/clientlog/domain	1.570s
ok  	courses/internal/contracts/access	1.683s
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	1.133s
ok  	courses/internal/course/adapters/http	15.323s
ok  	courses/internal/course/adapters/postgres	1.588s
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	4.749s
ok  	courses/internal/course/domain	2.003s
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	14.169s
?   	courses/internal/discussion/adapters/postgres	[no test files]
ok  	courses/internal/discussion/application	2.066s
ok  	courses/internal/discussion/domain	2.162s
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	3.325s
ok  	courses/internal/identity/adapters/http	12.497s
ok  	courses/internal/identity/adapters/mail	3.346s
?   	courses/internal/identity/adapters/postgres	[no test files]
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	3.257s
ok  	courses/internal/identity/application	2.945s
ok  	courses/internal/identity/domain	2.858s
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	2.588s
ok  	courses/internal/image/adapters/http	2.818s
?   	courses/internal/image/adapters/postgres	[no test files]
ok  	courses/internal/image/application	2.448s
ok  	courses/internal/image/domain	1.449s
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	2.245s
ok  	courses/internal/media/adapters/ffmpeg	1.917s
ok  	courses/internal/media/adapters/filesystem	2.115s
ok  	courses/internal/media/adapters/http	11.360s
?   	courses/internal/media/adapters/postgres	[no test files]
?   	courses/internal/media/adapters/postgres/generated	[no test files]
?   	courses/internal/media/adapters/tus	[no test files]
ok  	courses/internal/media/application	2.437s
ok  	courses/internal/media/domain	2.620s
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	11.692s
?   	courses/internal/orders/adapters/postgres	[no test files]
ok  	courses/internal/orders/adapters/wayforpay	3.389s
ok  	courses/internal/orders/application	3.349s
ok  	courses/internal/orders/domain	2.598s
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	5.715s
ok  	courses/internal/playback/adapters/http	5.921s
ok  	courses/internal/playback/adapters/postgres	3.690s
ok  	courses/internal/playback/adapters/redis	3.920s
ok  	courses/internal/playback/application	2.609s
ok  	courses/internal/playback/domain	2.611s
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	3.315s
ok  	courses/internal/testsupport/coursevisibility	1.713s
ok  	courses/internal/transcript/adapters/asynq	1.792s
ok  	courses/internal/transcript/adapters/http	3.125s
ok  	courses/internal/transcript/adapters/media	0.933s
?   	courses/internal/transcript/adapters/postgres	[no test files]
ok  	courses/internal/transcript/adapters/whisper	2.260s
ok  	courses/internal/transcript/application	2.548s
ok  	courses/internal/transcript/domain	1.714s
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	2.513s
ok  	courses/platform/config	2.615s
ok  	courses/platform/httpserver	3.304s
ok  	courses/platform/jsonbody	2.817s
ok  	courses/platform/mail	2.730s
ok  	courses/platform/observability	3.203s
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]
```
ФАКТ: фінальний ordered gate PASS: gofmt -l . → go vet ./... → go build ./... → go test -count=1 ./..., кожен exit 0 (повний вивід вище). T2 реалізовано; готово до незалежного review Courses 2.
ФАКТ: реалізація та всі gate outputs закомічені як a269237 на s0.0-spike; 10 scoped files. Чужі WIP/plan/tasks виключено. Merge/push/e2e не виконувалися.
ФАКТ: фінальний AI Factory checkpoint OK: 5f5782c85e26ce6251b42c64f5b19b99-2-35781166bbcc. Status перевірено: поза журналом лишилися тільки п’ять попередніх чужих docs/tasks змін.
ДАЛІ: зберегти цей фінальний запис окремим docs commit; власник перевіряє a269237 і передає Courses 2. Незалежний verdict — PENDING, технічні гейти — PASS; відкритих контрактних суперечностей не виявлено.
