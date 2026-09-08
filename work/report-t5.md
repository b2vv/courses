ДАЛІ: прочитати інструкції implement, репозиторний бриф і work/brief-t5.md; перевірити контракти перед виконанням першого кроку.
ФАКТ: прочитано implement, work/CTO-RESEARCH.md і work/brief-t5.md. Бриф забороняє git, Docker-гейти, ручне редагування generated і зміни поза T5; це має пріоритет над загальним workflow скіла.
РІШЕННЯ: до реалізації перевірити всі названі контракти. Якщо контракту бракує — зупинитися; якщо неточність спричинить runtime-баґ — відкрити `## Discussion`.
ДАЛІ: прочитати D-1/D-3/D-4/D-5/D-8, цільові рядки §8, G3/G4/G6/G11/G12/G13, звіти T3/T4, стандарти й HTTP composition seam.
ФАКТ: цільові секції локалізовано. D-1/D-4 задають один Redis-ключ, безумовний Take та 204/409/410; D-5 — clamp, монотонність і межу seq; §8 поіменно містить усі рядки T5; T3 вже реалізував Sessions/ProgressWriter, T4 — authorize і єдиний smoke-скрипт.
ФАКТ: HTTP-шар playback зараз реєструє лише внутрішній authorize; public play/heartbeat ще відсутні. Наявні identity middleware дають потрібні Authenticate і RequireSameOrigin seams.
ДАЛІ: дочитати без обрізання D-8, повні звіти T3/T4, composition root, playback application/ports/adapters, OpenAPI і smoke-скрипт; окремо перевірити, чи `ready` гарантує наявність опублікованих файлів.
ФАКТ: D-8 підтверджує обов’язковий `lessonId` і точне HTTP-мапування 204/409/410/404/400. Звіти T3/T4 підтверджують, що runtime integration уже був зеленим у рецензента, а R02/R03/R05–R09 свідомо перенесені до T5.
ФАКТ: наявний `playback-authorize.sh` створює ready-рядок і HLS-файли окремими ручними діями, а ownership — напряму в Redis; T5 має перевести policy-сценарії на справжній `play`, не створюючи другого скрипта.
ДАЛІ: перевірити B39/C19 і контракт AV1-метрики, повний production seam Sessions/ProgressWriter, composition root, OpenAPI conventions та атомарність `ready` відносно filesystem publish.
ФАКТ: план містить P12 «проба AV1 доїжджає полем мінту», але §10 призначає його окремому T12 після T5; brief-t5 водночас вимагає AV1 у T5. У перевірених джерелах поки немає назви/типу поля запиту чи точної метрики.
ФАКТ: application `Sessions` уже має тонкі `Take` і `Heartbeat`; composition root поки лише валідовує playback config і монтує authorize. Для public endpoint треба підключити існуючі Redis writer, PostgreSQL progress/access adapters та config.
ДАЛІ: знайти точне тіло `POST /api/v1/playback/sessions`, контракт AV1 metric і готовий access query для play. Якщо хоча б один обов’язковий контракт відсутній — зупинити реалізацію без вигадування.

## Стоп — бракує контракту public `play`

ФАКТ: `work/brief-t5.md:24-47` визначає перевірки й response `play`, але не визначає request body. Зокрема, endpoint мусить знайти урок, ready-відео й збережений `clientSeq`, проте `lessonId` не названий ні в path, ні в query, ні в body. Без адреси уроку handler не може виконати контракт до `Take`; довільно вивести її неможливо.

ФАКТ: `work/brief-t5.md:46-47` вимагає покласти результат AV1-проби «полем цього ж тіла», але не задає назви поля, його типу/допустимих значень, поведінки при відсутності поля чи exact metric name/labels. `docs/playback/plan.md:497` повторює лише абстрактне «тіло ... несе результат проби», а `docs/playback/plan.md:624` і `work/tasks.md:165` відносять реалізацію проби до окремого T12 після T5. У репозиторії такого контракту або метрики немає.

ФАКТ: це не косметична невизначеність. Вигаданий `lessonId` transport shape зламає майбутній Orval-клієнт при іншому рішенні контракту; вигаданий boolean/enum AV1 field або labels створять несумісний public API чи unbounded/непридатну серію метрики.

РІШЕННЯ: канал «Стоп» застосовано до початку production-коду. `backend/`, `frontend/`, `infrastructure/tests/playback-authorize.sh`, OpenAPI і generated не змінювались; жоден гейт не запускався, бо реалізація заблокована відсутнім контрактом.

ПИТАННЯ: зафіксуйте, будь ласка, точний request schema для `POST /api/v1/playback/sessions` (щонайменше спосіб передати `lessonId`; назву/тип/requiredness AV1-поля) і точний контракт метрики (ім’я та bounded labels). Також підтвердьте, чи T5 всупереч `docs/playback/plan.md:624`/`work/tasks.md:165` має реалізувати T12 зараз.

ДАЛІ: після появи контракту реалізувати T5 тонким HTTP transport, OpenAPI + generation, composition wiring, доповненням єдиного smoke-скрипта та доступними host-гейтами; Docker-гейти лишити рецензентові згідно з брифом.
ФАКТ: виправлений `work/brief-t5.md` перечитано. Стоп знято: `play` приймає рівно `{ "lessonId": "<uuid>" }`; `videoId` сервер знаходить сам; AV1 повністю поза T5 і лишається T12.
РІШЕННЯ: попередній блок `## Стоп` є історичним append-only записом і замінений цим рішенням. Жодного AV1-поля чи метрики в T5 не буде.
РІШЕННЯ: робота нижче порога `spec-flow`, як прямо визначено брифом. Застосовую TDD на HTTP/application seams; git-коміт зі skill workflow не робиться через явний NOT-список.
ДАЛІ: прочитати точні playback domain/error/adapters, SQL schema та HTTP/OpenAPI test conventions; спроєктувати мінімальний play lookup без читальної моделі T6.
ФАКТ: наявний `Sessions` уже правильно поновлює ownership перед progress і не пише при missing/taken-over; `ProgressWriter` сам мапить зниклий урок у `ErrLessonNotFound`, робить server-side clamp і застосовує domain monotonicity.
ФАКТ: для `play` бракує лише реального зовнішнього seam: PostgreSQL lookup за `lessonId`, що поверне current ready `videoId`, course owner/status і збережений `clientSeq`; це порт на фактичну БД-залежність, не читальна модель T6.
РІШЕННЯ: lookup має відрізняти зниклий урок від відсутнього ready-відео та policy denial до Redis `Take`; transport лишається відповідальним лише за strict JSON/UUID, actor mapping, id mint, response і HTTP errors.
ДАЛІ: звірити acceptance для not-ready та точні error codes; потім написати failing application/HTTP/PostgreSQL contract tests перед production-кодом.
ФАКТ: `docs/playback/plan.md:442` називає `video_not_ready`; play-contract дозволяє 403/404, тож зниклий lesson мапиться 404 `lesson_not_found`, а відсутність current ready asset — 404 `video_not_ready`; policy denial — 403 `playback_forbidden`.
ФАКТ: план окремо має B28: DB asset `ready` при порожньому каталозі все одно дає mint, а 404 виникає лише на edge. Це означає, що сам state не є абсолютним доказом наявності байтів; бриф прямо просить перевірити цей ризик через Discussion.
ДАЛІ: перевірити production publish ordering і storage failure semantics. Якщо `ready` може існувати без файлів у штатній/післяаварійній системі, відкрити `## Discussion` до реалізації, але продовжити за явним B28.

## Discussion

### D1 — `ready` не є перевіркою наявності HLS-файлів у момент `play`

ФАКТ: штатна публікація спочатку переносить HLS у filesystem (`backend/internal/media/application/transcode.go:120-129`) і лише потім у тій самій PostgreSQL-транзакції ставить `state='ready'` (`backend/internal/media/adapters/postgres/publish.go:69-74`). Це захищає нормальний publish path, але DB і filesystem не мають спільної транзакції, а API за конституцією не має media mount. Після видалення/втрати volume або ручного `ready`-рядка БД лишається `ready`, хоча `master.m3u8` відсутній. Сам план уже визнає це в B28 (`docs/playback/plan.md:534`).

ПИТАННЯ: формулювання brief «відео в стані ready» не може гарантувати playable bytes. У runtime `play` законно змінтить ownership і поверне `mediaPath`, а перший edge-запит дасть 404. Пропонована правка документації: назвати `ready` DB-передумовою мінту, не filesystem-гарантією, і лишити missing-bytes recovery рядку B28/T7. Додавати filesystem check у Go API не можна: це потребувало б забороненого media mount і створило TOCTOU.

РІШЕННЯ: T5 реалізує перевірку current DB `ready` буквально й не обходить архітектуру. B28 не заявляється закритим у T5, бо стан помилки належить плеєру T7.

ДАЛІ: написати failing tests для play eligibility/ordering, strict public HTTP contracts і 204/409/410/404/400 heartbeat mapping.
ФАКТ: RED application-тест додано: missing lesson, video-not-ready, draft denial, published/owner/admin allow; кожна відмова перевіряє, що ownership не забрано, кожен allow — точний server-selected ownership і `clientSeq`.
ФАКТ: `GOCACHE=/tmp/courses-t5-go-cache go test ./internal/playback/application` очікувано FAIL compile: відсутні `LessonPlayback`, `ErrVideoNotReady`, новий reader constructor seam і `Sessions.Play`.
ДАЛІ: реалізувати мінімальні domain/port/application контракти й довести GREEN application package.
ФАКТ: додано `LessonPlayback`, `ErrVideoNotReady`, зовнішній `LessonPlaybackReader` port і `Sessions.Play`: lookup → policy → unconditional Take. Жодної transport/DB деталі в application/domain немає.
ФАКТ: inventory `NewSessions` показав лише unit/integration playback callers; production composition ще не створював Sessions до T5.
ДАЛІ: оновити T3 integration constructors, форматувати й прогнати application unit + integration compile-only.
ФАКТ: T3 integration constructors оновлено лише новим nil reader seam; їхня поведінка heartbeat/authorize не змінена.
ФАКТ: `gofmt` виконано; `GOCACHE=/tmp/courses-t5-go-cache go test ./internal/playback/application` — PASS; `go test -run '^$' -tags=integration ./internal/playback/application` — compile-only PASS.
ДАЛІ: додати RED-тести PostgreSQL lookup: current highest-generation ready asset, saved/missing clientSeq, lesson-not-found, video-not-ready, corrupt/dependency classification.
ФАКТ: RED PostgreSQL adapter-тест додано для server-selected ready asset, course policy data, saved `clientSeq`, exact SQL predicates та error classification.
ФАКТ: `go test ./internal/playback/adapters/postgres` — очікуваний FAIL: `NewLessonPlaybackReader` ще відсутній. Тест також виявив локальну помилку фікстури — package constants не існували; її виправляю до GREEN, це не production finding.
ДАЛІ: додати self-contained UUID constants у тест і реалізувати `LessonPlaybackReader` без ORM/sqlc/generated змін.
ФАКТ: `LessonPlaybackReader` реалізовано одним PostgreSQL query: lesson/course row + highest-generation non-evicted ready asset + user-scoped progress seq; missing lesson, missing ready video, dependency/corrupt row розрізняються.
ФАКТ: `gofmt` і `GOCACHE=/tmp/courses-t5-go-cache go test ./internal/playback/adapters/postgres` — PASS.
ДАЛІ: додати integration-tagged доказ query проти isolated PostgreSQL і compile-only прогін; runtime лишити Docker-рецензентові.
ФАКТ: PostgreSQL integration proof розширено: no-progress → `clientSeq=0`, новий ready generation витісняє старий у server selection, saved seq=12 повертається, no ready → `video_not_ready`, deleted lesson → `lesson_not_found`.
ФАКТ: `gofmt` і `go test -run '^$' -tags=integration ./internal/playback/adapters/postgres` — compile-only PASS; runtime не запускався через Docker-обмеження.
ДАЛІ: написати HTTP RED для strict `{lessonId}` mint response/server path та heartbeat status/code matrix, authentication+same-origin wrappers і OpenAPI validation.
ФАКТ: HTTP RED додано: 201 mint із 32-hex `playbackId`, server-built `mediaPath`, config cadence і saved seq; strict one-field play body; play failure matrix без 409; heartbeat 204/409/410/404/503 та invalid-body rejection before application.
ФАКТ: перший 10s package invocation не встиг повернути output; повторний `go test ./internal/playback/adapters/http` очікувано FAIL compile лише на ще не реалізованих `NewHandler`/`Register` signatures.
ДАЛІ: реалізувати thin public HTTP transport з bounded strict JSON, UUID/playback-id validation, random 128-bit mint і exact error envelopes; зберегти authorize semantics.
ФАКТ: public HTTP transport реалізовано: обидва POST routes обгорнуті authenticate + same-origin; тіла обмежені 8 KiB, strict JSON з required/unknown/duplicate/trailing rejection; id валідовані до application.
ФАКТ: `play` мінтить 128-bit lowercase hex через `crypto/rand`, передає лише lesson до application, повертає 201 із повним server-built mediaPath/config/seq і не має 409 mapping. Heartbeat зберігає distinct 409/410 та 204 без body.
ФАКТ: `gofmt` і `GOCACHE=/tmp/courses-t5-go-cache go test ./internal/playback/adapters/http` — PASS.
ДАЛІ: додати OpenAPI paths/schemas/responses і contract validation test; не редагувати generated вручну.
ФАКТ: OpenAPI conventions перевірено: `additionalProperties:false`, shared no-store/error envelopes й kin-openapi request/response validation у handler tests. Поточний global error enum ще не містить playback public codes.
ДАЛІ: дописати два public paths, bounded request/response schemas та playback error codes; потім додати focused kin-openapi validation для фактичних HTTP відповідей.
ФАКТ: OpenAPI доповнено `createPlaybackSession` і `heartbeatPlayback`, exact one-field play schema, bounded heartbeat schema, full mint response та named 403/404/409/410 responses. AV1 поля немає.
ФАКТ: focused kin-openapi test валідовує реальні 201/204/409/410 request+response pairs; `gofmt` і `go test ./internal/playback/adapters/http` — PASS.
ДАЛІ: підключити playback reader/writer/progress/config у API composition root і оновити compile callers; потім прогнати всі playback unit та architecture checks.
ФАКТ: API composition root тепер створює PostgreSQL lesson/progress adapters, Redis ownership writer, `Sessions` із config TTL і монтує public routes через identity Authenticate/RequireSameOrigin. SSR/Go media mount не додано.
ФАКТ: `gofmt` і `GOCACHE=/tmp/courses-t5-go-cache go test ./internal/playback/... ./platform/architecture ./cmd/api` — PASS.
ДАЛІ: довести public policy/heartbeat integration проти real Redis+PostgreSQL у наявному combined integration suite; compile-only у цьому середовищі.
ФАКТ: combined integration suite тепер поіменно доводить B17 fresh repeated play + saved seq, старий id denied/new id allowed; R15/R16 no progress and no foreign TTL renewal; B44 120→100 clamp; R32 deleted lesson→not found.
ФАКТ: `gofmt` і `go test -run '^$' -tags=integration ./internal/playback/application` — compile-only PASS; real Redis/PostgreSQL не запускались.
ДАЛІ: перечитати повний єдиний `playback-authorize.sh` і Compose cadence, потім додати real-mint policy fixtures R02/R03/R05–R09/R27 із pacing та cleanup.
ФАКТ: скрипт повністю перечитано. Він уже має paced 400-request batches нижче media 120r/s і відновлює зупинені dependencies, але primary fixture сидів у seeded course та ownership писав напряму в Redis.
РІШЕННЯ: зробити fixture self-contained із run-unique course; додати host-cookie real `play` flows перед наявними measurements. R05 чекатиме фактичний `ownershipTtlSeconds+1`, а не штучно скорочуватиме Redis TTL; R07 додатково перевірить matching ownership record до inaccessible-course request.
ДАЛІ: доповнити той самий скрипт, cleanup і фінальний PASS list; edge config не змінювати.
ФАКТ: єдиний `playback-authorize.sh` доповнено self-contained course, двома host cookie jars і real `POST /playback/sessions` flows для R02/R03/R05/R06/R07/R27; R07 перевіряє точний Redis ownership match до condition-3 denial; R08/R09 покривають закриті/обхідні paths. Cleanup видаляє обидва ownership keys, bytes, course tree і users.
ФАКТ: R05 читає TTL із server response і мовчить `TTL+1`; policy-запити послідовні, а наявний 400-request test лишив pacing 200ms між batches. `sh -n` — PASS.
ФАКТ: naive extraction embedded Node для `node --check` впала на shell-escaped backticks; це артефакт extraction (`\`` лишається у файлі), не syntax proof runtime heredoc. Потрібен check після моделювання shell unescape.
ДАЛІ: перевірити embedded Node після unescape, shell helper semantics і static NOT-list; runtime скрипт не запускати без Docker.
ФАКТ: embedded Node після моделювання heredoc unescape — syntax PASS; shell policy markers/inventory підтверджує всі R02/R03/R05–R09/R27, cleanup і pacing. Runtime скрипт НЕ ЗАПУСКАВСЯ: Docker недоступний за брифом.
ДАЛІ: додати/перевірити API middleware order tests, потім прогнати playback unit, integration compile-only і OpenAPI tests разом.
ФАКТ: route test поіменно пінить `RequireSameOrigin → Authenticate` для обох public mutations; внутрішній authorize лишається на своєму окремому wrapper path.
ФАКТ: `go test ./internal/playback/... ./platform/architecture ./cmd/api` — PASS; `go test -run '^$' -tags=integration ./internal/playback/...` — compile-only PASS.
ДАЛІ: перевірити generation workflow і спробувати `make generate` лише якщо він не потребує Docker; якщо потребує — назвати точні очікувані generated artifacts і не писати їх руками.
ФАКТ: `Makefile:48-50` показує, що `make generate` починається з Dockerized sqlc, отже за брифом його НЕ ЗАПУСКАЮ й generated руками не правлю. Backend sqlc inputs не змінені, тому його output має лишитися byte-identical.
ФАКТ: reviewer `make generate` має додати в `frontend/src/adapters/api/generated/courses.ts` operations `createPlaybackSession`/`heartbeatPlayback`; створити model-файли для `PlaybackSessionRequest`, `PlaybackSession`, `PlaybackHeartbeatRequest`, `PlaybackForbiddenResponse`, `PlaybackNotFoundResponse`, `PlaybackTakenOverResponse`, `PlaybackExpiredResponse`; оновити `models/index.ts` і error-code enum.
ДАЛІ: виконати static scope/smell audit, потім доступні ordered gates `make format-check`, `make lint`, `make typecheck`, host Go tests; Docker-залежні результати не називати green.
ФАКТ: static audit — `gofmt -l` порожній, `sh -n` PASS; у T5 playback-коді немає TODO/FIXME/AV1/MediaSource/generated edits або залежності на `internal/media`. Збіги `any` — наявні Go interface values/англійська проза, не TypeScript escape hatch.
ДАЛІ: ordered gate 1 — `make format-check`; зафіксувати frontend результат і Docker-блокер backend окремо.
ФАКТ: `make format-check` — НЕ ЗАВЕРШИВСЯ: frontend oxfmt PASS (140 files); backend Dockerized gofmt не стартував через permission denied на Docker socket. Host `gofmt -l` уже порожній, але make-gate не названо green.
ДАЛІ: ordered gate 2 — `make lint`; зафіксувати frontend/architecture та Dockerized Go vet окремо.
ФАКТ: `make lint` — НЕ ЗАВЕРШИВСЯ: frontend oxlint + architecture PASS (78 files); backend Dockerized `go vet ./...` не стартував через Docker socket permission. Gate не green.
ДАЛІ: ordered gate 3 — `make typecheck`.
ФАКТ: `make typecheck` — PASS, exit 0.
ДАЛІ: ordered host Go verification без testcontainers: `go vet ./...`, `go test ./...`, targeted integration compile-only.
РІШЕННЯ: перед host Go verification виконати ordered `make test`, щоб виміряти frontend half; попередній `ДАЛІ` уточнено цим append-only записом.
ДАЛІ: ordered gate 4 — `make test`.
ФАКТ: `make test` — НЕ ЗАВЕРШИВСЯ: frontend Rstest не стартував через sandbox `listen EPERM ::1:3000`; Dockerized backend step не досягнуто. Gate не green.
ДАЛІ: host `go vet ./...` і `go test ./...`, потім playback integration-tag compile-only; це не заміна заблокованого make-gate.
ФАКТ: host `go vet ./...` — PASS. Host `go test ./...` — усі виконані packages, включно з playback, PASS; загальний exit 1 лише через sandbox bind `127.0.0.1:0` у `platform/httpserver/TestShutdownDrainsAnInFlightResponse`.
ФАКТ: через `&&` integration compile-only після загального Go test не стартував; це не пройдений крок.
ДАЛІ: окремо запустити playback integration-tag compile-only.
ФАКТ: `GOCACHE=/tmp/courses-t5-go-cache go test -run '^$' -tags=integration ./internal/playback/...` — compile-only PASS для всіх playback packages. Runtime testcontainers НЕ ЗАПУСКАЛИСЯ.
ДАЛІ: провести власне code review за обов’язковим implement workflow, але без git і без делегації: прочитати `code-review/SKILL.md`, визначити explicit T5 file set, звірити Standards і Spec axes.
ФАКТ: `code-review/SKILL.md` прочитано повністю. Він вимагає дві незалежні паралельні осі; це є явним skill-дозволом на sub-agents. Git fixed-point неможливий через T5 NOT-list, тому review працює по explicit file set без git.
РІШЕННЯ: file set: `backend/internal/playback/{domain/authorization.go,domain/errors.go,ports/sessions.go,application/sessions.go,application/sessions_test.go,application/sessions_integration_test.go,adapters/postgres/playback.go,adapters/postgres/playback_test.go,adapters/postgres/access_integration_test.go,adapters/http/handler.go,adapters/http/handler_test.go}`, `backend/cmd/api/main.go`, `backend/openapi/openapi.yaml`, `infrastructure/tests/playback-authorize.sh`, `work/report-t5.md`.
ДАЛІ: запустити паралельно Standards і Spec read-only reviews; заборонити git та редагування.
ФАКТ: дві незалежні review-осі запущено read-only без git/редагування.
ФАКТ: локальна Spec-вісь знайшла дефект: shared error mapper технічно дозволяв `play` відповісти 409, якщо service поверне takeover. Production `Sessions.Play` цього не генерує, але абсолютний public contract не був transport-level invariant.
РІШЕННЯ: додано regression case; `play` тепер мапить неможливі takeover/expired у 500 `internal_error`, тому 409/410 існують лише на heartbeat. `go test ./internal/playback/adapters/http` — PASS.
ДАЛІ: отримати незалежні Standards/Spec verdicts, виправити лише підтверджені findings і повторити touched checks.
ФАКТ: незалежний Spec-review підтвердив один High finding: R09 використовував випадковий неактивний `$playback_id`, тому 403 міг доводити лише відсутність ownership, а не traversal-захист для чинної сесії. Також reviewer повторно знайшов play→409 ризик, уже усунутий попереднім endpoint-specific mapping, і позначив відсутній generated client як чесно задокументовану Docker-залежну partial delivery.
ФАКТ: незалежний Standards-review не знайшов hard violations. Його transport-name finding усунуто локальним `errInvalidRequest`; пропозиції прибрати T3 `Sessions.Take` або винести два малих strict decoders у спільний helper відхилено як scope/architecture regressions: `Take` — чинний T3 application seam, а transport validation має модульно різні контракти.
ДАЛІ: виправити R09 так, щоб усі traversal probes несли server-minted active admin playback ID, потім перевірити shell/embedded Node і HTTP package після локального error-name/query hardening.
ФАКТ: R09 виправлено: скрипт дістає `playbackId` із того самого успішного `admin_play`, чий `mediaPath` перед цим доводить active ownership, і використовує цей ID у plain/encoded/double-encoded traversal probes. Випадковий legacy ID у цих доказах більше не бере участі.
ДАЛІ: виконати `sh -n`, syntax-check embedded Node після shell-unescape, `gofmt` і точкові Go-тести для змінених playback packages.
ФАКТ: `gofmt` і точкові `go test ./internal/playback/adapters/http ./internal/playback/application ./internal/playback/adapters/postgres ./platform/architecture ./cmd/api` — PASS після query rejection, endpoint-specific play mapping і transport-local invalid-request error.
ФАКТ: shell checks у тому самому invocation не стартували через помилковий cwd `backend` (`infrastructure/tests/...` не знайдено). Це не результат скрипта; повторюю з repository root.
ДАЛІ: повторити обидві static script-перевірки з кореня репозиторію та окремо перевірити, що R09 path справді містить extracted active ID.
ФАКТ: `sh -n infrastructure/tests/playback-authorize.sh` — PASS. Static inventory показує `admin_playback_id` extracted із `admin_play` і всі три R09 paths використовують саме його.
ФАКТ: попередня команда embedded-Node extraction повернула exit 0, але pattern `node <<NODE` нічого не знайшов і перевірила порожній stdin; це не зараховано як proof.
ДАЛІ: знайти фактичні heredoc delimiters і повторити непорожній Node syntax-check, який fail-иться при нульовому extraction.
ФАКТ: фактичний block починається `docker compose exec -T frontend node --input-type=module <<JS` і завершується `JS`; попередній delimiter `NODE` був неправильний.
ДАЛІ: витягти JS-блок за фактичними delimiters, перевірити що він непорожній, змоделювати shell unescape й запустити `node --check`.
ФАКТ: extraction guard підтвердив непорожній JS, але host `node --check` впав на top-level `import`, бо invocation не повторив runtime `--input-type=module`. Це помилка check-команди, не скрипта; результат не зараховано.
ДАЛІ: повторити той самий guarded extraction із `node --input-type=module --check`.
ФАКТ: guarded non-empty embedded JS extraction + shell-unescape + `node --input-type=module --check` — PASS. Разом із `sh -n` static smoke syntax green; runtime досі не запускався без Docker.
ДАЛІ: перечитати фінальні §7–§9 брифу й скласти повну матрицю acceptance→proof, commands, blocked gates та generation delta без завищення статусів.
ФАКТ: актуальний бриф перечитано повністю (фінальні вимоги містяться в §4–§6): матриця має охопити B17/B31/B41/B42/B44, R02/R03/R05–R09/R15/R16/R22/R27/R32; Docker/runtime/generation gates треба назвати незапущеними; AV1/T12 і B39 не заявляються.
ДАЛІ: зібрати точні test/script anchors для кожного рядка матриці та перевірити, що кожна заява прив’язана до фактичного assertion, а не лише до назви тесту.
ФАКТ: inventory знайшов усі required anchors: B17/B31 і R15/R16/B44/R32 у application integration suite; B41/B42/B44 у PostgreSQL progress integration; R22 у HTTP 204 case з position=120; R02/R03/R05–R09/R27 у єдиному smoke-скрипті. Smoke anchors є static-only до reviewer runtime.
ДАЛІ: прочитати тіла цих тестів навколо assertions, уточнити B31/R22 докази й лише потім записати фінальну таблицю.
ФАКТ: B17/B31 один наскрізний integration case: перший `play`, другий fresh ID без ownership, старий ID стає forbidden, новий дозволений, saved seq незмінний. R22 transport case приймає position=120 з 204; B44 real PostgreSQL case доводить 120→duration 100, а existing progress suite також пінить B41/B42.
РІШЕННЯ: перейменувати B17 case на `B17 B31 ...`, не змінюючи поведінки, щоб required proof був discoverable за обома row IDs.
ДАЛІ: оновити лише ім’я integration test, форматувати й повторити application unit + integration compile-only.
ФАКТ: integration test label тепер явно містить `B17 B31`; assertions не змінені.
ДАЛІ: виконати `gofmt`, application unit test без cache і integration-tag compile-only.
ФАКТ: `gofmt`; `go test -count=1 ./internal/playback/application` — PASS; `go test -count=1 -run '^$' -tags=integration ./internal/playback/application` — compile-only PASS.
ДАЛІ: звірити Make targets, щоб фінальні reviewer commands зберігали конституційний порядок і реально піднімали dependencies для integration/smoke.
ФАКТ: Makefile підтверджує: `generate` починається Dockerized sqlc; `check-generated` повторно запускає generate й використовує git; `up` будує та чекає Compose; повний нормативний порядок гейтів — format-check→lint→typecheck→test→integration→build→e2e.
ДАЛІ: дописати фінальну evidence matrix, точні reviewer commands, generation delta та незапущені/не-green gates.

## Рядок §8 → чим доводиться

| Рядок | Доказ | Стан тут |
|---|---|---|
| B17 | `TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL/B17_B31...`: повторний `Play` дає fresh ID, зберігає `clientSeq`, старий ID заборонений, новий дозволений | integration compile PASS; runtime reviewer |
| B31 | той самий case: caller без поточного ownership безумовно забирає слот другим server-minted ID | integration compile PASS; runtime reviewer |
| B41 | `TestProgressWriterAgainstPostgreSQL/B41...`: restore після заміни duration clamp-иться і не завершує урок без growth | integration compile PASS; runtime reviewer |
| B42 | `TestProgressWriterAgainstPostgreSQL/B42...`: `ended` на duration ставить completed | integration compile PASS; runtime reviewer |
| B44 | application integration `120→100` і PostgreSQL progress case `120→90`; transport приймає 120, не повертає 400 | unit transport PASS; integration compile PASS; runtime reviewer |
| R02 | existing single smoke: real student mint → course draft → playlist 403 | static syntax PASS; runtime reviewer |
| R03 | real student mint → course archived → playlist 403 | static syntax PASS; runtime reviewer |
| R05 | real mint, sleep `ownershipTtlSeconds+1`, перша спроба 403 | static syntax PASS; runtime reviewer |
| R06 | admin-owned real playback path від authenticated outsider → 403 | static syntax PASS; runtime reviewer |
| R07 | real ownership JSON exact-match перевірено в Redis, потім той самий draft media path → 403 | static syntax PASS; runtime reviewer |
| R08 | source/thumbnails/temporary paths → exact 403 | static syntax PASS; runtime reviewer |
| R09 | plain/encoded/double-encoded traversal із чинним server-minted admin ID → 403 або 404 через `curl --path-as-is` | static syntax + active-ID audit PASS; runtime reviewer |
| R15 | taken-over heartbeat лишає PostgreSQL progress незмінним | unit no-write PASS; integration compile PASS; runtime reviewer |
| R16 | той самий case порівнює Redis PTTL до/після і забороняє renewal чужого ownership | integration compile PASS; runtime reviewer |
| R22 | HTTP test передає position=120 і одержує 204; B44 доводить server clamp | host unit PASS; integration runtime reviewer |
| R27 | real mint і segment 200 до logout; наступний той самий segment 403 | static syntax PASS; runtime reviewer |
| R32 | deleted lesson → application `ErrLessonNotFound`; HTTP mapper → 404 `lesson_not_found` | host mapper PASS; integration compile PASS; runtime reviewer |

## Команди для рецензента

Спершу materialize generated client і перевірити delta:

```sh
make generate
make check-generated
```

Потім нормативні гейти у фіксованому порядку:

```sh
make format-check
make lint
make typecheck
make test
make integration
make build
make up
sh infrastructure/tests/playback-authorize.sh
make e2e
```

Для адресного runtime proof T5 проти real PostgreSQL/Redis після доступного Docker:

```sh
cd backend
go test -count=1 -tags=integration -run '^TestVideoAccessReaderAgainstPostgreSQL$' ./internal/playback/adapters/postgres
go test -count=1 -tags=integration -run '^TestProgressWriterAgainstPostgreSQL$' ./internal/playback/adapters/postgres
go test -count=1 -tags=integration -run '^TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL$' ./internal/playback/application
cd ..
```

## Що має дати `make generate`

- Backend sqlc output byte-identical: schema/query inputs не змінювались.
- `frontend/src/adapters/api/generated/courses.ts`: operations `createPlaybackSession` і `heartbeatPlayback`.
- Generated models для `PlaybackSessionRequest`, `PlaybackSession`, `PlaybackHeartbeatRequest`, `PlaybackForbiddenResponse`, `PlaybackNotFoundResponse`, `PlaybackTakenOverResponse`, `PlaybackExpiredResponse`.
- Оновлені generated `models/index.ts` та error-code enum із playback codes.
- Жодного AV1 поля: це T12.

## Незапущені та не-green гейти

- НЕ ЗАПУСКАЛИСЯ через Docker restriction брифу: `make generate`, `make check-generated`, `make integration`, `make build`, `make up`, `make e2e`, runtime integration tests і `sh infrastructure/tests/playback-authorize.sh`.
- `make format-check` не green: frontend PASS, Dockerized backend formatter не стартував. Host `gofmt -l` порожній.
- `make lint` не green: frontend lint/architecture PASS, Dockerized Go vet не стартував. Host `go vet ./...` PASS.
- `make test` не green: frontend Rstest не зміг bind `::1:3000` у sandbox; backend make-step не досягнуто. Host package tests виконані окремо; повний `go test ./...` мав єдину sandbox bind-відмову в `platform/httpserver`.
- `make typecheck` — PASS.

## Review verdict

- Standards: hard violations не знайдено. Transport-local naming виправлено; T3 `Sessions.Take` і module-local strict decoders свідомо лишені без нової speculative abstraction.
- Spec: High R09 false-positive виправлено active server-minted ID; play→409 invariant уже закрито endpoint-specific mapper + regression test. Єдина known partial delivery — generated client до reviewer `make generate`.
- Відкрита продуктова неточність D1 лишається в `## Discussion`: DB `ready` не є filesystem-гарантією; реалізація дотримується B28 та не додає заборонений media mount.

ДАЛІ: виконати фінальний uncached touched-package test, integration compile-only, format/shell/scope audit; зафіксувати exact result після цієї фінальної матриці.
ФАКТ: перший фінальний composite invocation не green: Go commands не мали task-local `GOCACHE` і впали на sandbox read `~/Library/Caches/go-build`; shell продовжився без `set -e`; scope grep хибно зловив давно наявні media composition imports у `cmd/api/main.go`. Це помилки verification-команди, не code failures; жоден composite verdict не зараховано.
ДАЛІ: повторити з `set -eu`, `GOCACHE=/tmp/courses-t5-go-cache` і точним audit: AV1/MediaSource у T5 production artifacts та `internal/media` лише всередині нового playback модуля.
ФАКТ: фінальний composite PASS: `gofmt` clean; task-local-cache `go vet` PASS; uncached `go test` PASS для всіх playback packages, architecture і API composition; integration-tag compile-only PASS для всіх playback packages; `sh -n` і guarded embedded ES-module syntax PASS; AV1/MediaSource відсутні; playback модуль не залежить від `internal/media`.
РІШЕННЯ: T5 implementation і доступна локальна verification завершені. Залишок не маскується: reviewer мусить materialize generated client і виконати Docker/runtime gates та smoke за командами вище.
