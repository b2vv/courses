# S0.1b-sec — error boundary normalization (§16.13) — WORK FILE

Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — `ДАЛІ:` до першого інструменту.
ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінал = стислий переказ. Base: branch s0.0-spike, HEAD у brief.

## Журнал
ДАЛІ: Прочитати briefing, локальний workflow і §16.13; перевірити branch/HEAD та провести factual audit кожної HTTP error-обгортки до змін. Потім мінімально нормалізувати unknown → dependency_unavailable/503, перевірити повними заданими гейтами без e2e/merge/push.
ФАКТ: Прочитано briefing та workflow/adapter; перевіряю поточний snapshot, рішення власника і повний inventory обгорток.
ФАКТ: Snapshot s0.0-spike @0149d9c67ed9cbb115e44ffe62311ecd77c44be3 підтверджено. Уже існують чужі WIP: _wip-s0.1b2-request.md, s0-openapi-gen-plan.md, work/tasks.md, _wip-s0.1b3a-errorcode.md; не редагую і не комічу їх.
ДАЛІ: Читаю тіла всіх HTTP error writers, включно з transcript/clientlog/platform, щоб відокремити explicit errors.Is мапінги від fallback.
ФАКТ: Factual audit ДО product-змін, baseline 0149d9c (усі шляхи нижче від backend/):
- internal/course/adapters/http/handler.go:181 writeError — unknown → internal_error/500, сирого рядка немає; статус не відповідає новому правилу 503.
- course/admin_handler.go:539 writeAdminError — unknown → internal_error/500, без raw.
- course/document_handler.go:167 writeDocumentError — unknown → internal_error/500, без raw.
- course/learning_handler.go:110 writeLearningError — explicit purchase/forbidden збережені; unknown делегує course.writeError → internal_error/500.
- adminstats/adapters/http/handler.go:69 writeError — unknown → internal_error/500, без raw.
- discussion/adapters/http/handler.go:642 writeListError — ErrNotFound → lesson_not_found/404; решта → statusFor + writeError, unknown raw/400.
- discussion:650 writePublishError — ErrNotFound → lesson_not_found/404; unknown raw/400.
- discussion:658 writeMutationError — ErrNotFound → comment_not_found/404; unknown raw/400.
- discussion:666 writeReportError — ErrNotFound → comment_not_found/404; unknown raw/400.
- discussion:674 writeTeachListError — ErrNotFound → course_not_found/404; unknown raw/400.
- discussion:681 writeReviewError — ErrNotFound → переданий notFound/404; unknown raw/400.
- discussion:688 writeParticipantReportError — ErrNotFound → course_not_found/404; unknown raw/400.
- discussion:696 statusFor — errors.Is групи 403/409/503/400, default 400.
- discussion:711 writeError — безумовний code := err.Error(); code-switch формує message, origin_forbidden перевизначає статус на 403. Немає перевірки identity для code, wrapped recognized errors також віддають wrapper text.
- identity/adapters/http/handler.go:791 writeError — unknown без override → internal_error/500; explicit override origin_forbidden/invalid_request/forbidden/grant_forbidden є transport-мапінгами без errors.Is. Немає raw.
- orders/adapters/http/handler.go:785 writeError — unknown → internal_error/500, без raw.
- media/adapters/http/handler.go:237 writeMediaError — unknown → internal_error/500, без raw. Уточнення до контексту: dependency_unavailable/503 тут є лише explicit recognized case.
- playback/adapters/http/handler.go:304 writePlayError — recognized TakenOver/SessionExpired НАВМИСНО internal_error/500; unknown делегує writePlaybackError.
- playback:312 writePlaybackError — unknown → internal_error/500; explicit TakenOver/SessionExpired → 409/410.
- playback:167 writeAuthorizationError — приймає status, не error; статичні playback_forbidden або dependency_unavailable, не raw.
- image/adapters/http/handler.go:105 writeUploadError — unknown уже dependency_unavailable/503, відповідає вимозі.
- transcript/adapters/http/handler.go:119 writeTranscriptError — unknown → invalid_request/400; немає raw. ВАЖЛИВО: domain.ErrInvalidRequest також обслуговується ЛИШЕ default, а get:64/70 явно передає цю бізнес-помилку.
- clientlog/adapters/http/handler.go ingest — окремої writeError-обгортки немає; decode нормалізує parse-помилки до ErrInvalidEvent, NewClientEvent повертає доменну валідацію; статичний invalid_client_event/400 (точне string(domain.ErrInvalidEvent)), немає raw чи інфра-залежності на цьому шляху.
- platform/httpserver/server.go:55 WriteError — serializer status/code/message; не приймає error і не має unknown/default класифікації. Platform callers використовують статичні коди; panic recovery internal_error/500 є окремою panic-гілкою, не errors.Is error boundary.
РІШЕННЯ: STOP до product-змін за прямою умовою користувача. Transcript.ErrInvalidRequest є реальною бізнес-помилкою на default; буквальна заміна unmatched на 503 змінить її existing invalid_request/400. Потрібне явне рішення, чи додати errors.Is(ErrInvalidRequest) → existing invalid_request/400 перед новим fallback. Не перекласифіковую тихо.
ПИТАННЯ: Чи зберігаємо implicit бізнес-мапінг transcript.ErrInvalidRequest через новий explicit errors.Is case (invalid_request/400), а також наявні transport-мапінги identity override і discussion origin/notFound як явні безпечні статичні відповіді? Після цього unknown лишатиметься dependency_unavailable/503.
ФАКТ: Product-файлів не змінював; тести/gates NOT_RUN через STOP до реалізації; e2e/merge/push не запускав. Тимчасових cmd або spike-файлів не створював. Незалежний security-review залишається за Courses 2 після writer-verify Courses 13.
ФАКТ: AI Factory checkpoint exit=0, status=OK, checkpoint_id=727558409d2c2c694ae20b4910405d35-1-f9dfe663be0d; current native session 01a17ee7-7da0-7562-9c44-86ad0a6b96ae, exact Courses policy row підтверджено. Transcript не читав.
ФАКТ: Існуючі course/adapters/http/s0_spike_test.go та s0_spike_errcode_test.go належать baseline, мною не створені — збережено.
ДАЛІ: Дочекатися відповіді на STOP-питання; далі explicit mapping + fallback normalization, регресійні тести кожної обгортки, повні ordered gates з реальним виводом і scoped commits. Зараз commit не створено.
ФАКТ (уточнення audit clientlog): domain.ErrInvalidEvent має точне значення invalid_request, отже відповідь invalid_request/400; попередня назва invalid_client_event була неточною. Джерело: backend/internal/clientlog/domain/errors.go:13.
РІШЕННЯ: Writer погодив explicit transcript.ErrInvalidRequest → invalid_request/400 та збереження static transport mappings; застосувати це правило до всіх default-reliant бізнес-помилок. ДАЛІ: звірити snapshot і default/callers, реалізувати нормалізацію та повні backend gates.
ФАКТ: Snapshot без нових product-змін; owner explicit mapping дозвіл чинний. ДАЛІ: зберегти transcript validation default як explicit case; інші internal_error/500 defaults не мають окремого осмисленого business code; playback writePlayError explicit 500-case лишається незмінним.
ФАКТ: Discussion має 32 domain/application sentinel-и; code раніше брався з самого err, message-switch default зберігав Request is not valid. Static notFound створював errors.New тільки для передачі коду. План: залишити статичне rendering в окремому writeErrorCode, а error boundary перевести на explicit errors.Is → canonical literal; unknown примусово 503. Origin mapping залишається у static renderer.
РІШЕННЯ: Pre-code READY за owner §16.13 + поточним уточненням; existing security slice з inline implementation plan, один HTTP adapter layer, без нового spec-flow. Незалежне review передається Courses 2 через writer, не self-PASS. File/rg impact fallback охоплює всі internal/*/adapters/http та platform/httpserver; graph refresh не потрібен.
ФАКТ: Змінено fallback course×3 (learning делегує), adminstats, identity, orders, media, playback, transcript на dependency_unavailable/503. Explicit playback writePlayError internal_error/500 лишився незмінний. Image fallback уже правильний, не змінено.
ФАКТ: Винесено transcript.ErrInvalidRequest → invalid_request/400 (message також збережено). Discussion.ErrNotFound отримав explicit statusFor 400 для старого direct-helper mapping; contextual wrappers зберігають свої 404. У discussion всі domain/application sentinel-и отримали explicit errors.Is→canonical code; recognized статуси збережено. Static notFound render викликається напряму після errors.Is(ErrNotFound), без синтетичного errors.New. Origin renderer 403 збережено.
ДАЛІ: Додати per-wrapper unknown/wrapped/spoof regressions і recognized mapping coverage; перевірити diff до ordered gates.
ФАКТ: Додано error_boundary_test.go у 9 модулях: кожна wrapper отримує plain/wrapped unknown та errors.New з відомим текстом (не sentinel); canonical business cases перевіряються plain і wrapped. Expected status/code зафіксовано з baseline, не обчислюється production mapper-ом.
ФАКТ: Додано regression статичних identity overrides та discussion origin/course/review notFound; recognized discussion cases перевіряються через один спільний mapper-path, notFound та unknown через кожну wrapper без зайвого дублювання. gofmt -w виконано.

ДАЛІ: GATE gofmt -l .; backend Go commands in golang:1.27.3-bookworm, /work=backend, standard Makefile cache volumes.
```text

```
ФАКТ: gofmt -l . EXIT=0

ДАЛІ: GATE go vet ./...; backend Go commands in golang:1.27.3-bookworm, /work=backend, standard Makefile cache volumes.
```text

```
ФАКТ: go vet ./... EXIT=0

ДАЛІ: GATE go build ./...; backend Go commands in golang:1.27.3-bookworm, /work=backend, standard Makefile cache volumes.
```text

```
ФАКТ: go build ./... EXIT=0

ДАЛІ: GATE make test (includes full go test -race ./...); backend Go commands in golang:1.27.3-bookworm, /work=backend, standard Makefile cache volumes.
```text
cd frontend && npm test

> @courses/frontend@0.1.0 test
> rstest run

ФАКТ: Post-change audit: усі audited unknown/default error branches тепер 503/dependency_unavailable; image вже був правильним; learning/play делегують до нормалізованих helpers. Raw err.Error у production HTTP adapters відсутній. Static platform responses/recovery лишено: panic не є error-wrapper. Identity two_factor_required обробляється в login окремо як 200; transcript ErrDraftSuperseded обробляється worker application, не HTTP validation; інших осмислених бізнес-кодів на HTTP default не знайдено. У решті модулів default був generic internal_error/500, не business code. git diff --check EXIT=0.
---
tool: "@rstest/core@0.12.4"
timestamp: "2026-10-27T19:53:05.332Z"
runtime: {"node":"v24.11.1","platform":"darwin","cwd":"/Users/strelia/projects/courses/frontend"}
---

# Rstest Test Execution Report

## Summary

```json
{
  "status": "pass",
  "counts": {
    "testFiles": 92,
    "failedFiles": 0,
    "tests": 824,
    "failedTests": 0,
    "passedTests": 824,
    "skippedTests": 0,
    "todoTests": 0
  },
  "durationMs": {
    "total": 10083,
    "build": 5,
    "tests": 10078
  },
  "snapshot": {
    "added": 0,
    "updated": 0,
    "unmatched": 0,
    "removed": 0,
    "unchecked": 0
  }
}
```

## Failures

No test failures reported.

Note: all tests passed. Lists omitted for brevity.
docker run --rm -e GOTOOLCHAIN=local -v "/Users/strelia/projects/courses/backend:/work" -w /work -v courses-gomod:/go/pkg/mod -v courses-gobuild:/root/.cache/go-build golang:1.27.3-bookworm go test -race ./...
?   	courses/cmd/api	[no test files]
ok  	courses/cmd/asr-worker	(cached)
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]
ok  	courses/internal/adminstats/adapters/http	7.247s
?   	courses/internal/adminstats/adapters/postgres	[no test files]
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	(cached)
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	(cached)
ok  	courses/internal/clientlog/domain	(cached)
ok  	courses/internal/contracts/access	(cached)
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	(cached)
ФАКТ: Mechanical preservation audit EXIT=0: усі 9 не-discussion змінених handler-файлів дорівнюють baseline після лише погодженого fallback delta (і explicit transcript.ErrInvalidRequest); усі старі recognized cases byte-identical. Discussion static message/status renderer byte-identical після відокремлення від error classifier. Повний test gate ще виконується; цей audit-запис може бути між рядками його live output.
ФАКТ: AI Factory implementation checkpoint OK: 727558409d2c2c694ae20b4910405d35-2-ce2649a97422, exit=0.
--- FAIL: TestDocumentErrorMapping (0.00s)
    --- FAIL: TestDocumentErrorMapping/internal_error (0.00s)
        document_handler_test.go:356: status=503 want=500 body={"error":{"code":"dependency_unavailable","message":"A dependency is unavailable","requestId":"949969a18cd08cd7b06938f41031727a"}}
--- FAIL: TestCatalogHTTPMatchesOpenAPI (0.57s)
    --- FAIL: TestCatalogHTTPMatchesOpenAPI/unexpected (0.00s)
        handler_test.go:111: status 503 body={"error":{"code":"dependency_unavailable","message":"A dependency is unavailable","requestId":"f85f128a66d49ac1f99a2271fde2d1f7"}}
FAIL
FAIL	courses/internal/course/adapters/http	51.495s
ok  	courses/internal/course/adapters/postgres	(cached)
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	(cached)
ok  	courses/internal/course/domain	(cached)
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	53.711s
?   	courses/internal/discussion/adapters/postgres	[no test files]
ok  	courses/internal/discussion/application	(cached)
ok  	courses/internal/discussion/domain	(cached)
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	(cached)
--- FAIL: TestAdminUsersBusinessErrorMapping (6.49s)
    --- FAIL: TestAdminUsersBusinessErrorMapping/unexpected_failure (0.59s)
        admin_handler_test.go:275: status=503 calls=1 body={"error":{"code":"dependency_unavailable","message":"A dependency is unavailable","requestId":"4a4938b39d27e5e123e7256ee9c064b2"}}
--- FAIL: TestIdentityHTTPMatchesOpenAPI (0.63s)
    --- FAIL: TestIdentityHTTPMatchesOpenAPI/login_with_an_unexpected_failure (0.00s)
        handler_test.go:497: status 503 body={"error":{"code":"dependency_unavailable","message":"A dependency is unavailable","requestId":"6c4d8f87c6a117dad7bf27978496ccf4"}}
FAIL
FAIL	courses/internal/identity/adapters/http	37.245s
ok  	courses/internal/identity/adapters/mail	(cached)
?   	courses/internal/identity/adapters/postgres	[no test files]
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	(cached)
ok  	courses/internal/identity/application	(cached)
ok  	courses/internal/identity/domain	(cached)
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	(cached)
ok  	courses/internal/image/adapters/http	1.187s
?   	courses/internal/image/adapters/postgres	[no test files]
ok  	courses/internal/image/application	(cached)
ok  	courses/internal/image/domain	(cached)
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	(cached)
ok  	courses/internal/media/adapters/ffmpeg	(cached)
ok  	courses/internal/media/adapters/filesystem	(cached)
ok  	courses/internal/media/adapters/http	29.834s
?   	courses/internal/media/adapters/postgres	[no test files]
?   	courses/internal/media/adapters/postgres/generated	[no test files]
?   	courses/internal/media/adapters/tus	[no test files]
ok  	courses/internal/media/application	(cached)
ok  	courses/internal/media/domain	(cached)
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	31.002s
?   	courses/internal/orders/adapters/postgres	[no test files]
ok  	courses/internal/orders/adapters/wayforpay	(cached)
ok  	courses/internal/orders/application	(cached)
ok  	courses/internal/orders/domain	(cached)
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	1.222s
ok  	courses/internal/playback/adapters/http	1.798s
ok  	courses/internal/playback/adapters/postgres	(cached)
ok  	courses/internal/playback/adapters/redis	(cached)
ok  	courses/internal/playback/application	(cached)
ok  	courses/internal/playback/domain	(cached)
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	1.229s
ok  	courses/internal/testsupport/coursevisibility	(cached)
ok  	courses/internal/transcript/adapters/asynq	(cached)
ok  	courses/internal/transcript/adapters/http	6.776s
ok  	courses/internal/transcript/adapters/media	(cached)
?   	courses/internal/transcript/adapters/postgres	[no test files]
ok  	courses/internal/transcript/adapters/whisper	(cached)
ok  	courses/internal/transcript/application	(cached)
ok  	courses/internal/transcript/domain	(cached)
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	1.576s
ok  	courses/platform/config	(cached)
ok  	courses/platform/httpserver	(cached)
ok  	courses/platform/jsonbody	(cached)
ok  	courses/platform/mail	(cached)
ok  	courses/platform/observability	(cached)
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]
FAIL
make: *** [test] Error 1

```
ФАКТ: make test (includes full go test -race ./...) EXIT=2
ФАКТ: Перший повний make test EXIT=2. Знайдено legacy assertions для unknown-помилок у course і identity, що очікують старий 500/internal_error; це саме authorized behavior delta. ДАЛІ: звірити fixtures, змінити лише unknown expectations на 503/dependency_unavailable; повторити ordered gates. make integration ще NOT_RUN через попередній red.
ФАКТ: Оновлено рівно 4 fixture expectations: course catalog unexpected/private SQL secret, course document errors.New(unexpected), identity admin users private database detail, identity login private argon secret. Це не sentinel-и й не зміна business mapping. Assertions/status/code coverage збережено; openapi.yaml не змінено.

ДАЛІ: RUN 2 GATE gofmt -l . (same Docker environment as run 1).
```text

```
ФАКТ: RUN 2 gofmt -l . EXIT=0

ДАЛІ: RUN 2 GATE go vet ./... (same Docker environment as run 1).
```text

```
ФАКТ: RUN 2 go vet ./... EXIT=0

ДАЛІ: RUN 2 GATE go build ./... (same Docker environment as run 1).
```text

```
ФАКТ: RUN 2 go build ./... EXIT=0

ДАЛІ: RUN 2 GATE make test (full go test -race ./...) (same Docker environment as run 1).
```text
cd frontend && npm test

> @courses/frontend@0.1.0 test
> rstest run

ФАКТ (уточнення census): discussion classifier містить 29 domain/application sentinel-ів (не 32 у попередньому записі): 5 comment.go + 21 errors.go + 3 application.go. Regression matrix: 28 загальних mappings + 7 contextual ErrNotFound; unknown перевірено на всіх 8 error wrappers.
---
tool: "@rstest/core@0.12.4"
timestamp: "2026-10-27T19:54:41.789Z"
runtime: {"node":"v24.11.1","platform":"darwin","cwd":"/Users/strelia/projects/courses/frontend"}
---

# Rstest Test Execution Report

## Summary

```json
{
  "status": "pass",
  "counts": {
    "testFiles": 92,
    "failedFiles": 0,
    "tests": 824,
    "failedTests": 0,
    "passedTests": 824,
    "skippedTests": 0,
    "todoTests": 0
  },
  "durationMs": {
    "total": 9209,
    "build": 5,
    "tests": 9204
  },
  "snapshot": {
    "added": 0,
    "updated": 0,
    "unmatched": 0,
    "removed": 0,
    "unchecked": 0
  }
}
```

## Failures

No test failures reported.

Note: all tests passed. Lists omitted for brevity.
docker run --rm -e GOTOOLCHAIN=local -v "/Users/strelia/projects/courses/backend:/work" -w /work -v courses-gomod:/go/pkg/mod -v courses-gobuild:/root/.cache/go-build golang:1.27.3-bookworm go test -race ./...
?   	courses/cmd/api	[no test files]
ok  	courses/cmd/asr-worker	(cached)
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]
ok  	courses/internal/adminstats/adapters/http	(cached)
?   	courses/internal/adminstats/adapters/postgres	[no test files]
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	(cached)
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	(cached)
ok  	courses/internal/clientlog/domain	(cached)
ok  	courses/internal/contracts/access	(cached)
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	(cached)
ФАКТ: Додатковий full boundary audit: playback.authorize unknown вже викликає writeAuthorizationError(503), image.authorize default уже dependency_unavailable/503; обидва залишено. Єдиний runtime StatusInternalServerError у module HTTP handlers після змін — explicit playback writePlayError для TakenOver/SessionExpired, збережений і покритий тестом. Discussion wrapped recognized errors тепер емітять canonical code без приватного wrapper text; це required normalization, sentinel code/status незмінні. Course openapi_contract.go досі містить declared 500 response — enum/parser/spec cleanup поза цим scope, не змінено.
ok  	courses/internal/course/adapters/http	51.123s
ok  	courses/internal/course/adapters/postgres	(cached)
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	(cached)
ok  	courses/internal/course/domain	(cached)
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	(cached)
?   	courses/internal/discussion/adapters/postgres	[no test files]
ok  	courses/internal/discussion/application	(cached)
ok  	courses/internal/discussion/domain	(cached)
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	(cached)
ok  	courses/internal/identity/adapters/http	36.709s
ok  	courses/internal/identity/adapters/mail	(cached)
?   	courses/internal/identity/adapters/postgres	[no test files]
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	(cached)
ok  	courses/internal/identity/application	(cached)
ok  	courses/internal/identity/domain	(cached)
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	(cached)
ok  	courses/internal/image/adapters/http	(cached)
?   	courses/internal/image/adapters/postgres	[no test files]
ok  	courses/internal/image/application	(cached)
ok  	courses/internal/image/domain	(cached)
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	(cached)
ok  	courses/internal/media/adapters/ffmpeg	(cached)
ok  	courses/internal/media/adapters/filesystem	(cached)
ok  	courses/internal/media/adapters/http	(cached)
?   	courses/internal/media/adapters/postgres	[no test files]
?   	courses/internal/media/adapters/postgres/generated	[no test files]
?   	courses/internal/media/adapters/tus	[no test files]
ok  	courses/internal/media/application	(cached)
ok  	courses/internal/media/domain	(cached)
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	(cached)
?   	courses/internal/orders/adapters/postgres	[no test files]
ok  	courses/internal/orders/adapters/wayforpay	(cached)
ok  	courses/internal/orders/application	(cached)
ok  	courses/internal/orders/domain	(cached)
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	(cached)
ok  	courses/internal/playback/adapters/http	(cached)
ok  	courses/internal/playback/adapters/postgres	(cached)
ok  	courses/internal/playback/adapters/redis	(cached)
ok  	courses/internal/playback/application	(cached)
ok  	courses/internal/playback/domain	(cached)
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	1.133s
ok  	courses/internal/testsupport/coursevisibility	(cached)
ok  	courses/internal/transcript/adapters/asynq	(cached)
ok  	courses/internal/transcript/adapters/http	(cached)
ok  	courses/internal/transcript/adapters/media	(cached)
?   	courses/internal/transcript/adapters/postgres	[no test files]
ok  	courses/internal/transcript/adapters/whisper	(cached)
ok  	courses/internal/transcript/application	(cached)
ok  	courses/internal/transcript/domain	(cached)
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	1.363s
ok  	courses/platform/config	(cached)
ok  	courses/platform/httpserver	(cached)
ok  	courses/platform/jsonbody	(cached)
ok  	courses/platform/mail	(cached)
ok  	courses/platform/observability	(cached)
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]

```
ФАКТ: RUN 2 make test (full go test -race ./...) EXIT=0

ДАЛІ: RUN 2 GATE make integration (same Docker environment as run 1).
```text
sh infrastructure/tests/integration.sh
?   	courses/cmd/api	[no test files]
ФАКТ: Перед delivery перевірено git status: лише 14 змінених backend-файлів і 9 нових module regression-файлів належать цьому зрізу; чужі docs/tasks WIP лишаються окремо. ДАЛІ: після завершення ordered gates зробити scoped implementation commit і journal-only delivery commit з наявними trailers.
ok  	courses/cmd/asr-worker	0.015s
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]
ok  	courses/internal/adminstats/adapters/http	1.419s
ok  	courses/internal/adminstats/adapters/postgres	15.361s
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	0.010s
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	0.007s
ok  	courses/internal/clientlog/domain	0.008s
ok  	courses/internal/contracts/access	0.006s
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	0.002s
ok  	courses/internal/course/adapters/http	16.961s
ok  	courses/internal/course/adapters/postgres	23.029s
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	3.027s
ok  	courses/internal/course/domain	0.010s
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	15.516s
ok  	courses/internal/discussion/adapters/postgres	6.602s
ok  	courses/internal/discussion/application	0.008s
ok  	courses/internal/discussion/domain	0.010s
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	1.256s
ok  	courses/internal/identity/adapters/http	11.387s
ok  	courses/internal/identity/adapters/mail	0.005s
ok  	courses/internal/identity/adapters/postgres	17.512s
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	2.463s
ok  	courses/internal/identity/application	0.045s
ok  	courses/internal/identity/domain	0.017s
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	0.060s
ok  	courses/internal/image/adapters/http	0.046s
ok  	courses/internal/image/adapters/postgres	6.339s
ok  	courses/internal/image/application	0.011s
ok  	courses/internal/image/domain	0.004s
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	0.013s
ok  	courses/internal/media/adapters/ffmpeg	0.018s
ok  	courses/internal/media/adapters/filesystem	0.095s
ok  	courses/internal/media/adapters/http	9.511s
ok  	courses/internal/media/adapters/postgres	3.989s
?   	courses/internal/media/adapters/postgres/generated	[no test files]
ok  	courses/internal/media/adapters/tus	3.140s
ok  	courses/internal/media/application	0.003s
ok  	courses/internal/media/domain	0.006s
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	8.785s
ok  	courses/internal/orders/adapters/postgres	4.115s
ok  	courses/internal/orders/adapters/wayforpay	0.005s
ok  	courses/internal/orders/application	0.006s
ok  	courses/internal/orders/domain	0.006s
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	0.088s
ok  	courses/internal/playback/adapters/http	0.226s
ok  	courses/internal/playback/adapters/postgres	6.109s
ok  	courses/internal/playback/adapters/redis	1.027s
ok  	courses/internal/playback/application	5.678s
ok  	courses/internal/playback/domain	0.051s
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	3.997s
ok  	courses/internal/testsupport/coursevisibility	0.004s
ok  	courses/internal/transcript/adapters/asynq	4.300s
ok  	courses/internal/transcript/adapters/http	1.481s
ok  	courses/internal/transcript/adapters/media	0.027s
ok  	courses/internal/transcript/adapters/postgres	4.499s
ok  	courses/internal/transcript/adapters/whisper	5.118s
ok  	courses/internal/transcript/application	0.008s
ok  	courses/internal/transcript/domain	0.011s
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	0.643s
ok  	courses/platform/config	0.004s
ok  	courses/platform/httpserver	0.048s
ok  	courses/platform/jsonbody	0.003s
ok  	courses/platform/mail	1.037s
ok  	courses/platform/observability	0.009s
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]
integration log: /Users/strelia/projects/courses/backend/test-results/integration.log

```
ФАКТ: RUN 2 make integration EXIT=0
ФАКТ: Implementation commit 4aa15a8 — 23 scoped backend files; branch s0.0-spike, no merge/push. Recognized regressions + unknown tests і 4 authorized legacy fixture deltas включено; чужі WIP не staged. Повний make test RUN 2 EXIT=0; make integration ще виконується.
ФАКТ: Уточнення попереднього live-status: make integration уже завершився EXIT=0, повний output вище; Docker/Ryuk не блокували цей run. Backend integration detailed artifact: backend/test-results/integration.log (gitignored), output збережено тут.
ФАКТ: Фінальний ordered RUN 2 на code snapshot 4aa15a8: gofmt -l . (empty output) EXIT=0 → go vet ./... EXIT=0 → go build ./... EXIT=0 → make test (frontend + FULL go test -race ./...) EXIT=0 → make integration EXIT=0. Перший red збережено, не приховано. Таргетних Go-only test runs не було.
ФАКТ: Final scope audit: unknown paths у course (catalog/admin/document/learning), adminstats, discussion (8 wrappers), identity, orders, media, playback (play/playback), transcript нормалізовані; image upload/authorization та playback authorization вже відповідали правилу. Clientlog validation повертає лише static invalid_request; platform serializer/panic/static transport не змінювались. У module production HTTP err.Error() відсутній; business cases збережено, explicit transcript invalid_request/400 додано. Нові tests у всіх 9 error-wrapper модулях.
ФАКТ: openapi.yaml, enum/parser, бізнес-логіку, чужі docs/tasks WIP не змінював. Нових s0-spike або cmd-файлів не створював; baseline spike-tests збережено. No merge/push. make e2e NOT_RUN за прямою вказівкою writer.
ДАЛІ: Courses 13 — власний повний e2e/writer verification на 4aa15a8; Courses 2 — незалежний security-sensitive review baseline 0149d9c → 4aa15a8, звернути увагу на unknown 500/400→503, canonical wrapped discussion codes і незмінні static mappings. Review/e2e PENDING; загального phase PASS не заявляю.
ФАКТ: Final AI Factory checkpoint OK, exit=0: 727558409d2c2c694ae20b4910405d35-3-997d82b7a02b; native Codex session 01a17ee7-7da0-7562-9c44-86ad0a6b96ae. Scope/gates/evidence зафіксовані.
ДАЛІ: Journal-only delivery commit цього append-only файлу з тими самими trailers; product snapshot 4aa15a8 незмінний, повторний product gate не потрібен. Наступний власник — Courses 13.

## WRITER REVIEW FINDING (Courses 13, 2026-10-27) — перед гейтами/рев'ю Courses 2
Діф 4aa15a8 перевірено. ЗАУВАЖЕННЯ:
- Витік (code := err.Error()) був ЛИШЕ в discussion. Решта 8 модулів (adminstats/course/admin/document/identity/media/orders/playback) ВЖЕ мали безпечний default `internal_error`/500 — не текли.
- Codex змінив їх default `internal_error`/500 → `dependency_unavailable`/503. Семантично спірно: невідома/неочікувана помилка (баг/паніка) чесно = internal_error/500, а не dependency_unavailable/503 (специфічна заява «downstream лежить, ретрайни»).
- Неконсистентність: playback writePlayError:306 ЛИШИВСЯ internal_error/500, тоді як writePlaybackError:330 став 503. media/domain/errors.go:20 має ErrInternal="internal_error".
- discussion-фікс (явний switch кодів + default) — коректний і потрібний; його лишаємо.
РЕКОМЕНДАЦІЯ: unknown → internal_error/500 КОНСИСТЕНТНО (повернути 8 модулів до оригінального default; фіксувати лише discussion-витік через маршрут unknown→internal_error/500; вирівняти playback). internal_error і dependency_unavailable лишаються живими. Ескальовано власнику (реверс §16.13 wording + клієнт-видима семантика). Гейти/e2e НЕ ганяв до рішення — діф може зміститись.
РІШЕННЯ: Writer передав нове рішення власника 2026-10-27: security slice має зберегти baseline business mappings; discussion unknown → internal_error/500 замість raw/400, попередні зміни 500→503 в інших модулях відкликати. ДАЛІ: звірити baseline/current та неоднозначність image/transcript перед новими повними gates.
ФАКТ: Review recheck baseline 0149d9c: image.writeUploadError default = dependency_unavailable/503 (Image upload is temporarily unavailable); transcript.writeTranscriptError default = invalid_request/400 (Request is not valid). Решта перелічених generic defaults були internal_error/500. Поточний HEAD 5395257; знайдено 9 error_boundary_test.go, не 10.
ПИТАННЯ: Новий інваріант «єдина поведінкова зміна — discussion unknown» суперечить пункту «ВСІ error_boundary_test.go unknown → internal_error/500» для image/transcript. Чи зберігати їхні baseline fallback-и (image 503, transcript 400), чи санкціоновано також змінити їхні unknown на 500, з explicit recognized cases (зокрема transcript.ErrInvalidRequest=400, image.ErrDependencyUnavailable=503)?
РІШЕННЯ: До відповіді product-змін не роблю, щоб не порушити один із двох явних інваріантів. Застосовано receiving-code-review (/Users/strelia/.agents/skills/receiving-code-review/SKILL.md): «IF any item is unclear: STOP - do not implement anything yet». Це уточнення контракту image/transcript, а не повторний запит дозволу на погоджені правки discussion/restoration. Gates NOT_RUN у цьому review-followup; попередній green не покриває нове рішення.
ФАКТ: Review-followup checkpoint OK, exit=0: 727558409d2c2c694ae20b4910405d35-4-d5b6d4bc5e74. ДАЛІ: відповідь власника щодо baseline-vs-uniform fallback для image/transcript; product diff цього followup порожній.
РІШЕННЯ: Writer зняв неоднозначність: єдина production behavior delta проти 0149d9c — discussion; image unknown 503 та transcript unknown 400 зберегти. ДАЛІ: відновити інші handlers і чотири legacy test expectations до baseline; discussion unknown internal_error/500; regression pgx; ordered gates.
ФАКТ: Відновлено 9 non-discussion handler-файлів і 4 старі test-файли байтами 0149d9c; перед записом перевірено current==HEAD (жодних чужих правок). Image не змінено. Нові per-module tests тепер assert baseline unknown: image 503/dependency_unavailable, transcript 400/invalid_request, решта 500/internal_error.
ФАКТ: Discussion explicit errors.Is classifier та static renderer збережено; statusFor unknown=500, writeError default code=internal_error/status=500; explicit ErrDependencyUnavailable=503 незмінний.
ФАКТ: Додано real pgx ConnectError regression: pgconn.ConnectConfig з контрольованим DialFunc (без мережі) формує справжній ConnectError; два різні імені БД + user у wrapped err перевіряються як передумова, response мусить бути 500/internal_error і не містити приватних даних. Локальний pgx v5.10.0 source перевірено в Makefile Docker cache.
ДАЛІ: Перевірити exact baseline invariant, потім ordered full gates; будь-який business-test failure → STOP без зміни expected business mapping.

ФАКТ: Exact baseline invariant PASS: git diff 0149d9c для восьми non-discussion module trees містить лише нові error_boundary_test.go; жодних production/legacy-test deltas.

ДАЛІ: RUN 3 GATE gofmt -l .; Docker golang:1.27.3-bookworm /work=backend, Makefile cache volumes.
```text

```
ФАКТ: RUN 3 gofmt -l . EXIT=0

ДАЛІ: RUN 3 GATE go vet ./...; Docker golang:1.27.3-bookworm /work=backend, Makefile cache volumes.
```text

```
ФАКТ: RUN 3 go vet ./... EXIT=0

ДАЛІ: RUN 3 GATE go build ./...; Docker golang:1.27.3-bookworm /work=backend, Makefile cache volumes.
```text

```
ФАКТ: RUN 3 go build ./... EXIT=0

ДАЛІ: RUN 3 GATE go test -race ./...; Docker golang:1.27.3-bookworm /work=backend, Makefile cache volumes.
```text
?   	courses/cmd/api	[no test files]
ok  	courses/cmd/asr-worker	(cached)
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]
ФАКТ: git diff --check EXIT=0. Production behavior delta від baseline обмежена discussion; розпізнані domain/application коди та static transport mappings збережені. Старі записи про загальну 503-нормалізацію є історією відкликаного рішення, не фінальним контрактом.
ok  	courses/internal/adminstats/adapters/http	7.677s
?   	courses/internal/adminstats/adapters/postgres	[no test files]
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	(cached)
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	(cached)
ok  	courses/internal/clientlog/domain	(cached)
ok  	courses/internal/contracts/access	(cached)
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	(cached)
ФАКТ: Discussion-only implementation checkpoint OK, exit=0: 727558409d2c2c694ae20b4910405d35-5-0cde31d64b7a. Повний race gate ще працює; completion не заявляється до його виходу та integration.
ok  	courses/internal/course/adapters/http	52.782s
ok  	courses/internal/course/adapters/postgres	(cached)
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	(cached)
ok  	courses/internal/course/domain	(cached)
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	54.547s
?   	courses/internal/discussion/adapters/postgres	[no test files]
ok  	courses/internal/discussion/application	(cached)
ok  	courses/internal/discussion/domain	(cached)
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	(cached)
ok  	courses/internal/identity/adapters/http	38.623s
ok  	courses/internal/identity/adapters/mail	(cached)
?   	courses/internal/identity/adapters/postgres	[no test files]
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	(cached)
ok  	courses/internal/identity/application	(cached)
ok  	courses/internal/identity/domain	(cached)
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	(cached)
ok  	courses/internal/image/adapters/http	(cached)
?   	courses/internal/image/adapters/postgres	[no test files]
ok  	courses/internal/image/application	(cached)
ok  	courses/internal/image/domain	(cached)
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	(cached)
ok  	courses/internal/media/adapters/ffmpeg	(cached)
ok  	courses/internal/media/adapters/filesystem	(cached)
ok  	courses/internal/media/adapters/http	30.176s
?   	courses/internal/media/adapters/postgres	[no test files]
?   	courses/internal/media/adapters/postgres/generated	[no test files]
?   	courses/internal/media/adapters/tus	[no test files]
ok  	courses/internal/media/application	(cached)
ok  	courses/internal/media/domain	(cached)
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	31.513s
?   	courses/internal/orders/adapters/postgres	[no test files]
ok  	courses/internal/orders/adapters/wayforpay	(cached)
ok  	courses/internal/orders/application	(cached)
ok  	courses/internal/orders/domain	(cached)
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	(cached)
ok  	courses/internal/playback/adapters/http	1.887s
ok  	courses/internal/playback/adapters/postgres	(cached)
ok  	courses/internal/playback/adapters/redis	(cached)
ok  	courses/internal/playback/application	(cached)
ok  	courses/internal/playback/domain	(cached)
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	1.317s
ok  	courses/internal/testsupport/coursevisibility	(cached)
ok  	courses/internal/transcript/adapters/asynq	(cached)
ok  	courses/internal/transcript/adapters/http	7.104s
ok  	courses/internal/transcript/adapters/media	(cached)
?   	courses/internal/transcript/adapters/postgres	[no test files]
ok  	courses/internal/transcript/adapters/whisper	(cached)
ok  	courses/internal/transcript/application	(cached)
ok  	courses/internal/transcript/domain	(cached)
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	1.713s
ok  	courses/platform/config	(cached)
ok  	courses/platform/httpserver	(cached)
ok  	courses/platform/jsonbody	(cached)
ok  	courses/platform/mail	(cached)
ok  	courses/platform/observability	(cached)
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]

```
ФАКТ: RUN 3 go test -race ./... EXIT=0

ДАЛІ: RUN 3 GATE make integration; Docker golang:1.27.3-bookworm /work=backend, Makefile cache volumes.
```text
sh infrastructure/tests/integration.sh
?   	courses/cmd/api	[no test files]
ok  	courses/cmd/asr-worker	0.020s
?   	courses/cmd/media-worker	[no test files]
?   	courses/cmd/migrate	[no test files]
?   	courses/cmd/upload	[no test files]

ФАКТ: Final source audit assertions EXIT=0: повний backend diff 0149d9c має рівно один non-test файл — discussion/adapters/http/handler.go; static renderer byte-identical baseline, code := err.Error() вилучено.
ok  	courses/internal/adminstats/adapters/http	1.309s
ok  	courses/internal/adminstats/adapters/postgres	13.672s
?   	courses/internal/adminstats/application	[no test files]
ok  	courses/internal/adminstats/domain	0.007s
?   	courses/internal/adminstats/ports	[no test files]
ok  	courses/internal/clientlog/adapters/http	0.006s
ok  	courses/internal/clientlog/domain	0.005s
ok  	courses/internal/contracts/access	0.002s
?   	courses/internal/contracts/access/accesstest	[no test files]
ok  	courses/internal/contracts/rating	0.011s
ok  	courses/internal/course/adapters/http	13.232s

ФАКТ: Discussion-only correction commit d2ad7dd (22 scoped backend files), no merge/push; gofmt/vet/build/full go test -race ./... RUN 3 EXIT=0. make integration ще in progress. Чужі docs/tasks не staged.
ok  	courses/internal/course/adapters/postgres	21.462s
?   	courses/internal/course/adapters/postgres/generated	[no test files]
ok  	courses/internal/course/application	3.027s
ok  	courses/internal/course/domain	0.007s
?   	courses/internal/course/ports	[no test files]
ok  	courses/internal/discussion/adapters/http	12.963s
ok  	courses/internal/discussion/adapters/postgres	5.056s
ok  	courses/internal/discussion/application	0.002s
ok  	courses/internal/discussion/domain	0.012s
?   	courses/internal/discussion/ports	[no test files]
ok  	courses/internal/identity/adapters/argon2id	1.102s
ok  	courses/internal/identity/adapters/http	9.938s
ok  	courses/internal/identity/adapters/mail	0.006s
ok  	courses/internal/identity/adapters/postgres	15.913s
?   	courses/internal/identity/adapters/postgres/generated	[no test files]
ok  	courses/internal/identity/adapters/redis	1.651s
ok  	courses/internal/identity/application	0.009s
ok  	courses/internal/identity/domain	0.008s
?   	courses/internal/identity/ports	[no test files]
ok  	courses/internal/image/adapters/filesystem	0.033s
ok  	courses/internal/image/adapters/http	0.033s
ok  	courses/internal/image/adapters/postgres	5.201s
ok  	courses/internal/image/application	0.006s
ok  	courses/internal/image/domain	0.003s
?   	courses/internal/image/ports	[no test files]
ok  	courses/internal/media/adapters/asynq	0.019s
ok  	courses/internal/media/adapters/ffmpeg	0.009s
ok  	courses/internal/media/adapters/filesystem	0.034s
ok  	courses/internal/media/adapters/http	7.338s
ok  	courses/internal/media/adapters/postgres	3.814s
?   	courses/internal/media/adapters/postgres/generated	[no test files]
ok  	courses/internal/media/adapters/tus	4.097s
ok  	courses/internal/media/application	0.018s
ok  	courses/internal/media/domain	0.004s
?   	courses/internal/media/ports	[no test files]
ok  	courses/internal/orders/adapters/http	6.834s
ok  	courses/internal/orders/adapters/postgres	3.819s
ok  	courses/internal/orders/adapters/wayforpay	0.004s
ok  	courses/internal/orders/application	0.003s
ok  	courses/internal/orders/domain	0.004s
?   	courses/internal/orders/ports	[no test files]
ok  	courses/internal/platform/openapigen	0.062s
ok  	courses/internal/playback/adapters/http	0.146s
ok  	courses/internal/playback/adapters/postgres	5.067s
ok  	courses/internal/playback/adapters/redis	0.915s
ok  	courses/internal/playback/application	5.499s
ok  	courses/internal/playback/domain	0.004s
?   	courses/internal/playback/ports	[no test files]
ok  	courses/internal/testsupport/containers	3.641s
ok  	courses/internal/testsupport/coursevisibility	0.008s
ok  	courses/internal/transcript/adapters/asynq	3.381s
ok  	courses/internal/transcript/adapters/http	1.231s
ok  	courses/internal/transcript/adapters/media	0.011s
ok  	courses/internal/transcript/adapters/postgres	4.447s
ok  	courses/internal/transcript/adapters/whisper	5.116s
ok  	courses/internal/transcript/application	0.007s
ok  	courses/internal/transcript/domain	0.008s
?   	courses/internal/transcript/ports	[no test files]
ok  	courses/platform/architecture	0.854s
ok  	courses/platform/config	0.008s
ok  	courses/platform/httpserver	0.049s
ok  	courses/platform/jsonbody	0.028s
ok  	courses/platform/mail	1.142s
ok  	courses/platform/observability	0.011s
?   	courses/platform/postgres	[no test files]
?   	courses/platform/redis	[no test files]
?   	courses/platform/runtime	[no test files]
integration log: /Users/strelia/projects/courses/backend/test-results/integration.log

```
ФАКТ: RUN 3 make integration EXIT=0
ФАКТ: FINAL ACCEPTANCE (остаточне рішення writer): baseline 0149d9c → correction d2ad7dd. Єдиний production-file diff — backend/internal/discussion/adapters/http/handler.go. Усі інші handlers БАЙТОВО baseline; чотири legacy tests також відновлено. Нові baseline regression tests лишаються у 9 модулях.
ФАКТ: Discussion unknown plain/wrapped/spoof → internal_error/500; explicit business errors зберігають canonical code/status, ErrDependencyUnavailable → dependency_unavailable/503; static notFound/origin renderer byte-identical. Real pgx ConnectError з двома DB names та user не витікає в response (передумова приватного тексту перевірена тестом). Image unknown лишається dependency_unavailable/503, transcript unknown invalid_request/400, інші generic fallbacks internal_error/500; обидві playback wrappers consistent на unknown/500.
ФАКТ: RUN 3 ordered gates на product snapshot d2ad7dd: gofmt -l . (порожній output) EXIT=0 → go vet ./... EXIT=0 → go build ./... EXIT=0 → FULL go test -race ./... EXIT=0 → make integration EXIT=0. Реальний вивід кожної команди вище. Бізнес-тести не червоніли; Docker/Ryuk не блокували integration. Старий RUN 1 red стосувався відкликаного cross-module 503-рішення; не фінального коду.
ФАКТ: make e2e NOT_RUN за дорученням. OpenAPI/parser не змінено; tmp cmd/spike не створювались; чужі WIP збережено; no merge/push. Попередні записи про all-module normalization superseded цим рішенням.
ДАЛІ: Courses 13 writer/e2e та Courses 2 security-sensitive review baseline 0149d9c → d2ad7dd. Незалежний review/e2e залишаються PENDING, загальний phase PASS не заявляється. Зберегти journal-only delivery commit з тими самими trailers.
ФАКТ: Final discussion-only checkpoint OK, exit=0: 727558409d2c2c694ae20b4910405d35-6-10a291235020; native session 01a17ee7-7da0-7562-9c44-86ad0a6b96ae. Product snapshot d2ad7dd незмінний після gates.
