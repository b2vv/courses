ДАЛІ: прочитати implement-інструкції, work/brief-t4.md та обов’язкові контракти; потім виконувати бриф покроково.

ФАКТ: work/report-t4.md був відсутній; створено новий файл звіту.

ФАКТ: прочитано `/Users/strelia/.agents/skills/implement/SKILL.md`, `work/CTO-RESEARCH.md` і `work/brief-t4.md` повністю. Бриф забороняє торкатися git, тому вимога implement-скілу про коміт тут не застосовується.

РІШЕННЯ: працювати лише в межах T4: integration-доказ `take → authorize`, доповнення наявного edge-скрипта і звіт; не створювати T5 API та не змінювати заборонені поверхні.

ДАЛІ: прочитати всі перелічені у §1 джерела й перевірити, чи контракт достатній для реалізації.

ФАКТ: проінвентаризовано обов’язкові джерела: `docs/playback/plan.md` — 690 рядків, звіти T1/T3 — 310/209, наявний edge-скрипт — 282; `backend/internal/playback/` містить application/domain/ports та Redis/PostgreSQL/HTTP adapters, зокрема патерн `application/sessions_integration_test.go`.

ДАЛІ: прочитати джерела повністю (для плану — обов’язкові рішення й рядки §8 разом із контекстом), не перевідкриваючи дві вже підтверджені рецензентом властивості.

ФАКТ: прочитано `AGENTS.md`, `.claude/standards.md` і `work/tasks.md`. T4 — локальний доказовий slice одного backend-шва та одного наявного infra-скрипта; він нижче порога `spec-flow`, тому виконання йде за inline-брифом. У трекері T4 досі `planned`, але бриф прямо забороняє змінювати `work/tasks.md`.

РІШЕННЯ: дотримуватися порядку доступних гейтів `format-check → lint → typecheck → test`; Docker-залежні наступні гейти назвати незапущеними, не видавати за пройдені.

ДАЛІ: прочитати D-2/D-3, §8 і суміжні acceptance-рядки плану, звіти T1/T3 та весь наявний playback-код.

ФАКТ: локалізовано обов’язкові місця плану: D-2 — рядки 135+, D-3 — 193+, §8 — 468+, B12–B14 — 516–518, R01–R13 — 559–571, R20 — 578, R29 — 587. План прямо відносить їх до T4 у рядку 616.

ДАЛІ: прочитати ці діапазони з повним контекстом і перевірити формулювання проти вже наявного runtime-коду.

ФАКТ: прочитано D-2/D-3 та контекст дизайну/ризиків/задач. Контракт достатній на рівні поведінки: 204 лише після cookie → Redis ownership (`playbackId` і `videoId`) → PostgreSQL access; 401 нормалізується в 403, dependency failure — 503; edge без authorization cache.

ФАКТ: великий діапазон плану був обрізаний виводом інструмента, тому §8 ще не вважається прочитаним повністю.

ДАЛІ: окремо прочитати §8 без обрізання, потім звіти T1/T3.

ФАКТ: §8 прочитано повністю. Для T4 план розміщує B12–B14/R01–R13/R20/R29 на smoke-рівні; application integration із `Take → Authorize` є додатковим наскрізним доказом B13, конкретизованим брифом. B12 вимагає `nginx -t`, health і порожній `/var/cache/nginx`; B13 — рівно три authorize на плейлист + два сегменти; B14 — повтор того самого URL анонімом → 403; R29 — однакові статус і байти тіла для аноніма та не-власника.

ДАЛІ: прочитати звіти T1 і T3, щоби перевикористати точні фікстури, метрики та відомі обмеження.

ФАКТ: прочитано `work/report-t1.md` і `work/report-t3.md`. T1 уже поставив production `authorize`, edge без кеша, self-cleaning `playback-authorize.sh`, counters і пройдену рецензентом Docker-верифікацію. T3 уже поставив `Take`/`Renew`, спільний JSON `storedOwnership` і application orchestration; рецензент підтвердив повний `make check`.

РІШЕННЯ: production-код не розширювати без фактичної прогалини. Нову цінність T4 шукати тільки в (1) combined real Redis+PostgreSQL integration test `Take → Authorize` та (2) розширенні існуючого smoke-скрипта на B12–B14/R29.

ДАЛІ: прочитати весь `backend/internal/playback/`, nginx і наявний скрипт; звірити сигнатури, testcontainers harness та фактичні counters перед тестовими змінами.

ФАКТ: фактичний scope читання — 2 631 рядок: playback-код 2 207, `playback-authorize.sh` 282, nginx 142. Combined application integration pattern уже займає 182 рядки й буде базою, як вимагає бриф.

ДАЛІ: прочитати domain/ports/application повністю, потім adapters і edge-файли окремими порціями без обрізання.

ФАКТ: domain/ports/application прочитано повністю. `Sessions.Take(ctx,userID,Ownership)` приймає TTL із constructor; `Authorizer.Authorize(ctx,Actor,playbackID,videoID)` читає Redis, звіряє обидва id, далі читає PostgreSQL policy. Наявний `sessions_integration_test.go` уже підіймає Redis 7.4 + PostgreSQL 17.5 та застосовує реальні міграції, тож контрактів для combined тесту вистачає — стоп не потрібен.

РІШЕННЯ: додати сценарії T4 в наявний `application/sessions_integration_test.go`, а не створювати новий harness/файл; сконструювати `Authorizer` з production Redis reader і PostgreSQL access reader на тих самих контейнерах.

ДАЛІ: прочитати adapters повністю та перевірити constructor signatures/error mapping/TTL semantics.

ФАКТ: HTTP/PostgreSQL/Redis adapters прочитано повністю. Production constructors сумісні з combined тестом: `NewOwnershipWriter`, `NewOwnershipReader`, `NewVideoAccessReader(pool,counter)`. Redis `Take` використовує реальний expiration; missing key мапиться в `ErrForbidden`. PostgreSQL reader перевіряє exact non-evicted video та course visibility.

РІШЕННЯ: TTL-сценарій доводити коротким реальним expiration із bounded polling до зникнення ключа, а тоді викликати `Authorize`; не підміняти час Redis і не видаляти ключ руками.

ДАЛІ: прочитати `infrastructure/tests/playback-authorize.sh` і `nginx.conf` повністю та перевірити, чи §2.2 брифу точний відносно фактичного stack fixture.

ФАКТ: edge-файли прочитано повністю. Production nginx уже має безкешовий `auth_request`, deny/unavailable mapping, media limiter і deny-all fallback — конфіг змінювати не треба. Наявний скрипт уже робить `nginx -t`, перевіряє порожній cache dir, створює fixture, міряє counters і доводить R01/R10–R13/R20, але ще не має точних B13 (3 конкретні requests), явного B14 та byte-for-byte R29.

РІШЕННЯ: `nginx.conf` не змінювати. У тому самому Node-блоці скрипта додати ізольовані counter deltas для B13, повтор exact URL без cookie для B14, другу автентифіковану run-unique identity для R29 і порівняння `Uint8Array` response bodies.

ДАЛІ: перевірити healthcheck nginx і схему session/password fixtures; якщо брифове очікування B12 неможливо спостерігати наявним контрактом — зупинитися.

ФАКТ: `docker-compose.yml:147-153` задає nginx healthcheck і tmpfs `/var/cache/nginx`; B12 повністю спостережуваний. Наявний Argon2id hash для пароля `courseforge-admin-demo` придатний для другої run-unique identity. Відсутнього контракту немає.

РІШЕННЯ: B12 додатково assert-нути через Compose container id + Docker health status `healthy`; другу identity зробити `student`, без ownership key, і видаляти в наявному EXIT cleanup.

ДАЛІ: спроєктувати мінімальні тестові правки, перевірити точні line anchors, потім почати TDD із combined integration test.

ФАКТ: прочитано `tdd/SKILL.md`. Погоджений самим брифом seam — публічний application ланцюг `Sessions.Take → Authorizer.Authorize` через production Redis/PostgreSQL adapters; edge seam — спостережувана HTTP-поведінка піднятого stack. Тести не торкатимуться приватних деталей.

ДАЛІ: прочитати обов’язкові TDD references `tests.md` і `mocking.md`, перевірити `CONTEXT.md`, після чого внести перший RED integration slice.

ФАКТ: `tdd/tests.md` і `tdd/mocking.md` прочитано; `CONTEXT.md` відсутній. Для T4 використовуються справжні Redis/PostgreSQL, без mock внутрішніх collaborators; очікування — літерали зі сценаріїв брифу.

ДАЛІ: додати перший combined integration сценарій `Take → Authorize` у наявний harness і compile-check-нути його під `integration` tag (runtime неможливий без Docker).

ФАКТ: у `application/sessions_integration_test.go` додано перший B13 tracer bullet через production `OwnershipWriter` → Redis → `OwnershipReader` → `Authorizer` → PostgreSQL `VideoAccessReader`. Нового production-коду немає.

ФАКТ: `GOCACHE=/tmp/courses-t4-go-cache go test -run '^$' -tags=integration ./internal/playback/application` — compile-only PASS (`[no tests to run]`). Runtime RED/GREEN неможливий без Docker; це не runtime proof.

ДАЛІ: розширити цей самий combined test трьома відмовами: takeover старого id, чужий video id, реальне TTL expiry; повторити compile-check.

ФАКТ: combined application integration test тепер містить усі чотири сценарії брифу: allow після `Take`, deny старого playback після другого `Take`, deny foreign video та deny після реального 250ms Redis TTL. TTL-доказ спершу підтверджує allow, потім bounded polling-ом чекає denial через той самий public `Authorize` seam.

ФАКТ: повторний integration-tag compile-only check — PASS. Runtime testcontainers не запускався й не називається пройденим.

ДАЛІ: доповнити наявний edge-скрипт B12 health assertion і cleanup другої identity; потім syntax-check shell.

ФАКТ: `playback-authorize.sh` тепер для B12 вимагає наявний nginx container id, Docker health=`healthy`, `nginx -t` і порожній `/var/cache/nginx`. Додано run-unique outsider fixture та FK-safe cleanup обох test identities; другого скрипта не створено.

ФАКТ: `sh -n infrastructure/tests/playback-authorize.sh` — PASS після shell-правок.

ДАЛІ: у наявний Node runtime block додати B13 exact-three counter proof, B14 same-URL anonymous denial і R29 equal-status/equal-bytes proof; logout обох sessions.

ФАКТ: edge Node block розширено: B13 міряє delta counter навколо рівно трьох успішних requests; B14 повторює щойно дозволений segment URL без cookie; R29 тим самим URL порівнює anonymous та authenticated non-owner за status і кожним байтом `Uint8Array`. Обидві run-unique sessions завершуються logout.

ФАКТ: `sh -n infrastructure/tests/playback-authorize.sh` — PASS після повного edge-розширення.

ДАЛІ: окремо перевірити синтаксис embedded Node після heredoc-unescape; потім статично рев’ювати shell runtime sequencing і limiter budget.

ФАКТ: embedded Node після моделювання heredoc-unescape пройшов `node --check --input-type=module`, exit 0. Це синтаксичний, не stack-runtime доказ.

ДАЛІ: перечитати актуальні diff-фрагменти без git, перевірити cleanup, health assertion, counter isolation, cookie/session lifecycle та pacing; за наявності баготворної неточності відкрити `## Discussion` до продовження.

ФАКТ: актуальні змінені фрагменти перечитано з номерами рядків без git. B13 counter window містить лише три послідовні media requests і лишається далеко нижче media limiter; B14/R29 використовують exact same segment URL; no-owner має валідну окрему session; production nginx не змінено.

ДАЛІ: підтвердити FK cleanup semantics для session rows і перевірити shell/Node форматування; після цього виконати targeted unit/compile verification.

ФАКТ: `backend/migrations/00002_identity.sql:29` задає `user_sessions.user_id REFERENCES users(id) ON DELETE CASCADE`, тому EXIT cleanup прибирає sessions навіть якщо Node block впаде до logout. Cleanup другої identity не залишає зовнішнього стану.

ДАЛІ: застосувати `gofmt` до зміненого Go integration test і повторити статичні синтаксичні перевірки.

ФАКТ: `gofmt -w backend/internal/playback/application/sessions_integration_test.go` — exit 0.

ДАЛІ: повторити integration compile-check, `sh -n` та embedded Node parse на відформатованому change-set.

ФАКТ: фінальний на цьому етапі `go test -run '^$' -tags=integration ./internal/playback/application` з `GOCACHE=/tmp/courses-t4-go-cache` — compile-only PASS; runtime не виконувався.

ДАЛІ: повторити shell parser check.

ФАКТ: фінальний `sh -n infrastructure/tests/playback-authorize.sh` — PASS.

ДАЛІ: повторити embedded Node parser check.

ФАКТ: фінальний embedded Node `--check` після heredoc-unescape — PASS.

ДАЛІ: виконати targeted playback unit suite та architecture boundary без integration runtime; потім пройти дозволені ordered make-gates.

ФАКТ: `GOCACHE=/tmp/courses-t4-go-cache go test ./internal/playback/... ./platform/architecture` — PASS для всіх playback packages і backend architecture boundary.

ДАЛІ: запустити ordered `make format-check`; якщо середовище всупереч брифу звернеться до Docker, назвати точний незапущений підкрок.

ФАКТ: `make format-check` — НЕ ЗАВЕРШИВСЯ, exit 2: frontend oxfmt PASS (140 files), backend Docker `gofmt` підкрок не стартував через permission denied до Docker socket. Gate не green.

ДАЛІ: запустити ordered `make lint` і так само відокремити виконану frontend-половину від Docker blocker.

ФАКТ: `make lint` — НЕ ЗАВЕРШИВСЯ, exit 2: frontend oxlint + architecture PASS (78 files), backend Docker `go vet` не стартував через permission denied до Docker socket. Gate не green.

ДАЛІ: запустити ordered `make typecheck`.

ФАКТ: `make typecheck` — PASS, exit 0.

ДАЛІ: запустити ordered `make test`; якщо Docker/loopback недоступні, зафіксувати відмову без підміни її partial checks.

ФАКТ: `make test` — НЕ ЗАВЕРШИВСЯ, exit 2: frontend Rstest не зміг bind `::1:3000` (`listen EPERM`); backend Docker test підкрок не був досягнутий. Gate не green.

ДАЛІ: виконати дозволені host Go checks (`gofmt -l`, `go vet`, повний non-integration `go test`) і окремо зберегти targeted PASS; це додаткові докази, не заміна make-гейтів.

ФАКТ: host `gofmt -l .` у `backend/` — exit 0 без виводу.

ДАЛІ: виконати host `go vet ./...` із task-local `GOCACHE`.

ФАКТ: host `GOCACHE=/tmp/courses-t4-go-cache go vet ./...` — PASS.

ДАЛІ: виконати повний host non-integration `go test ./...`; очікуваний sandbox loopback blocker назвати точно, якщо повториться.

ФАКТ: host `GOCACHE=/tmp/courses-t4-go-cache go test ./...` — НЕ ПОВНИЙ PASS, exit 1: усі playback та інші виконані packages green; єдина відмова — `platform/httpserver TestShutdownDrainsAnInFlightResponse`, sandbox заборонив bind `127.0.0.1:0`. Це не дефект T4.

ДАЛІ: перевірити integration-tag compile для всього playback й підготувати code review згідно implement-скілу, не використовуючи git.

ФАКТ: `GOCACHE=/tmp/courses-t4-go-cache go test -run '^$' -tags=integration ./internal/playback/...` — compile-only PASS для всіх playback packages; runtime testcontainers не запускався.

ДАЛІ: прочитати `code-review/SKILL.md` повністю й виконати його настільки, наскільки дозволяє брифова заборона git; review не має змінювати scope.

ФАКТ: `code-review/SKILL.md` прочитано повністю. Його звичайний fixed-point/diff процес неможливий, бо T4 NOT-список забороняє будь-який git touch. За прецедентом T1/T3 review виконується двома незалежними read-only осями по явному T4 file set: `sessions_integration_test.go`, `playback-authorize.sh`, `work/report-t4.md`.

РІШЕННЯ: Standards axis звіряє `AGENTS.md`, `.claude/standards.md` і smell baseline; Spec axis — `work/brief-t4.md`, D-2/D-3 та §8. Обом заборонено git і редагування.

ДАЛІ: запустити обидві незалежні review-осі паралельно; локально паралельно скласти acceptance/gate таблиці звіту.

ФАКТ: запущено два незалежні read-only reviews без git і без редагування: Standards та Spec. Їм передано точний T4 file set, стандарти/контракти й вимогу назвати runtime consequence кожної знахідки.

ДАЛІ: поки reviews працюють, дописати обов’язкову таблицю «рядок §8 → чим доводиться», точні reviewer-команди та попередній перелік незапущених Docker-гейтів.

## Рядок §8 → чим доводиться

| Рядок | Доказ T4 | Статус у цьому середовищі |
|---|---|---|
| B12 | `infrastructure/tests/playback-authorize.sh`: container `nginx -t`, Compose container id, Docker health=`healthy`, порожній `/var/cache/nginx` | `sh -n` PASS; stack runtime НЕ ЗАПУСКАВСЯ |
| B13 | `TestSessionsOwnershipFailuresCannotWriteProgress`: production `Take → OwnershipReader → Authorizer → VideoAccessReader` для allow, takeover, foreign video й TTL expiry; той самий edge-скрипт міряє рівно 3 authorize calls на master playlist + 2 segments | integration-tag compile PASS; Redis/PostgreSQL та stack runtime НЕ ЗАПУСКАЛИСЯ |
| B14 | edge-скрипт: після успішного B13 segment request анонім повторює exact same URL і мусить отримати 403 | embedded Node parse PASS; stack runtime НЕ ЗАПУСКАВСЯ |
| R29 | edge-скрипт: анонім і authenticated run-unique non-owner просять exact same segment; assert однакового 403 і byte-for-byte рівності `Uint8Array` body | embedded Node parse PASS; stack runtime НЕ ЗАПУСКАВСЯ |

ФАКТ: суміжні pre-existing proofs R01/R10–R13/R20 та M1–M4/B37/B38 збережено в тому самому скрипті без переписування; T4 лише доповнює його. Production реалізація вже була повною: `application/authorize.go:17-31`, `application/sessions.go:21-23`, `adapters/redis/ownership.go:22-46`, `ownership_writer.go:34-51`, `adapters/postgres/access.go:29-48`, `infrastructure/nginx/nginx.conf:55-91`.

РІШЕННЯ: production-код не змінювався. Змінено лише integration proof `backend/internal/playback/application/sessions_integration_test.go`, наявний `infrastructure/tests/playback-authorize.sh` і цей append-only звіт; жодних абстракцій/прапорів/підготовки до T5 не додано.

## Команди рецензента

З кореня репозиторію в Docker-enabled середовищі:

```sh
cd backend
GOCACHE=/tmp/courses-t4-go-cache go test -count=1 -tags=integration -run '^TestSessionsOwnershipFailuresCannotWriteProgress$' ./internal/playback/application
cd ..
make up
make seed-identities
make seed
sh infrastructure/tests/playback-authorize.sh
```

Повний ordered gate після runtime proofs:

```sh
make format-check
make lint
make typecheck
make test
make integration
make check-generated
make build
make up
make seed-identities
make seed
make smoke
make e2e
```

## Незапущені та незавершені гейти

| Перевірка | Результат |
|---|---|
| `make format-check` | НЕ ЗАВЕРШИВСЯ: frontend PASS; backend Docker підкрок не стартував |
| `make lint` | НЕ ЗАВЕРШИВСЯ: frontend lint/architecture PASS; backend Docker підкрок не стартував |
| `make typecheck` | PASS |
| `make test` | НЕ ЗАВЕРШИВСЯ: Rstest bind `::1:3000` → EPERM; backend підкрок не досягнуто |
| playback unit + architecture | PASS |
| playback integration tag | compile-only PASS; runtime НЕ ЗАПУСКАВСЯ |
| `make integration`, `make check-generated`, `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`, `make check` | НЕ ЗАПУСКАЛИСЯ; Docker недоступний і бриф забороняє називати їх пройденими |
| `sh infrastructure/tests/playback-authorize.sh` | НЕ ЗАПУСКАВСЯ; потребує піднятого Compose stack |

ДАЛІ: отримати незалежні Standards/Spec verdicts; виправити лише підтверджені T4 findings, повторити доступні checks і завершити звіт.

ФАКТ: Standards review — hard violations 0. Judgement call: назва `TestSessionsOwnershipFailuresCannotWriteProgress` стала оманливо вузькою після додавання authorization subtests; також відзначено малий дубль двох login fixture flows у shell.

РІШЕННЯ: назву suite слід розширити, щоб reviewer міг адресно запускати combined proof. Дві явні login fixture послідовності поки не абстрагувати: вони короткі, мають різні assertion labels і helper заради двох викликів додасть зайву непрозорість.

ДАЛІ: дочекатися Spec review; потім одним циклом застосувати підтверджені правки.

ФАКТ: Spec review — FAIL через конфлікт плану з executable brief і один flake-ризик тесту; NOT-list scope creep не знайдено.

## Discussion

### D1 — рядок задачі T4 обіцяє докази, яких бриф не замовляє, а один із них до T4 неможливий

ФАКТ: `docs/playback/plan.md:616` приписує T4 усі `R01–R13, R20, R29`. Натомість `work/brief-t4.md:45-57` наказує додати на edge лише B12–B14/R29, а `work/brief-t4.md:76` прямо забороняє `play`/heartbeat. `docs/playback/plan.md:562` визначає R04 як e2e «грає понад 30 с», що фізично потребує плеєра T7. Наявний `playback-authorize.sh` доводить R01/R10–R13/R20 і новий R29, але не R02/R03/R05–R09 на smoke-рівні. Якщо закрити T4 за рядком 616, фаза отримає хибне покриття policy/path regressions; якщо буквально будувати R04, буде порушено NOT-список і створено T5/T7 передчасно.

ПИТАННЯ: запропонована правка плану — забрати R04 з T4 й лишити його T7; у T4 назвати фактичний новий scope B12–B14/R29 + combined application proof, а pre-existing R01/R10–R13/R20 позначити preserved, не новими доказами. R02/R03/R05–R09 або отримують окремий названий smoke slice, або не вважаються закритими T4.

РІШЕННЯ: `docs/playback/*` не редагую через NOT-список і не будую player/API. У таблиці доказів не приписую T4 непроведені smoke-сценарії; рецензент отримує точні команди лише для реально доданих доказів.

### D2 — «чужий videoId» із брифу не є доказом R07 condition 3

ФАКТ: `work/brief-t4.md:39` вимагає `Take`, потім authorize з чужим `videoId`; `application/authorize.go:24-26` закономірно відмовляє на condition 2, до PostgreSQL не доходить. `docs/playback/plan.md:565` R07 натомість вимагає matching ownership для відео недоступного курсу, тобто виконання condition 3. Якщо назвати brief-сценарій доказом R07, регресія course visibility query лишиться зеленою.

ПИТАННЯ: запропонована правка — лишити brief-сценарій як окремий B13 ownership-binding proof і не маркувати ним R07; для R07 потрібна окрема fixture з matching ownership та inaccessible course на smoke-рівні.

РІШЕННЯ: поточний тест зберігає буквальний сценарій брифу й названий лише B13; R07 не заявляється закритим.

ДАЛІ: виправити review findings у дозволеному scope: перейменувати combined suite й збільшити TTL window; повторити compile/static checks.

ФАКТ: review fix застосовано: suite перейменовано на `TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL`; TTL збільшено з ранніх 250ms до 2s, bounded expiry deadline — до 10s. Це зменшує scheduler-flake до виконання pre-expiry assertion.

РІШЕННЯ: ранні записи про назву suite, 250ms TTL і reviewer-команду є історичними та замінені цим append. Актуальна runtime-команда:

```sh
cd backend
GOCACHE=/tmp/courses-t4-go-cache go test -count=1 -tags=integration -run '^TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL$' ./internal/playback/application
```

ДАЛІ: виконати `gofmt`, integration compile-check, shell і embedded Node checks після review fix.

ФАКТ: post-review `gofmt -w backend/internal/playback/application/sessions_integration_test.go` — exit 0.

ДАЛІ: post-review integration-tag compile-check актуального suite.

ФАКТ: post-review integration-tag compile-only check — PASS, `[no tests to run]`; runtime Docker proof досі не запускався.

ДАЛІ: post-review `sh -n`.

ФАКТ: post-review `sh -n infrastructure/tests/playback-authorize.sh` — PASS.

ДАЛІ: post-review embedded Node parse.

ФАКТ: post-review embedded Node parser check — PASS.

ДАЛІ: провести короткий повторний Standards/Spec review саме review-fix і Discussion disposition; потім фінально повторити дозволені ordered checks, які могли зачепити зміни.

ФАКТ: обидві незалежні review-осі запущено повторно без git/редагування. Standards перевіряє закриття naming finding; Spec — повноту саме executable brief та коректність `## Discussion`, не підміняючи документну неточність scope creep-ом.

ДАЛІ: повторити targeted playback suite й `make typecheck`; інші make-gates не можуть змінити відомий Docker/loopback verdict, але фінальний static стан буде переміряно.

ФАКТ: post-review targeted `go test ./internal/playback/... ./platform/architecture` з task-local `GOCACHE` — PASS.

ДАЛІ: повторити фінальний `make typecheck`.

ФАКТ: фінальний `make typecheck` — PASS, exit 0.

## Code review — фінальний verdict

### Standards

PASS — hard violations 0. Naming finding закрито перейменуванням combined suite. Лишається один non-blocking judgement call: два короткі явні login flows у shell дублюють форму; helper не додано, бо для двох fixture-ролей він збільшив би непрозорість без нового capability.

### Spec

PASS — усі executable вимоги `work/brief-t4.md` реалізовано; TTL flake закрито. `## Discussion` D1/D2 правильно лишає plan/brief mismatch нерозв’язаною документною проблемою без неправдивого закриття R02–R09/R04/R07 і без виходу в T5/T7.

ПІДСУМОК REVIEW: Standards 0 blocking findings; Spec 0 implementation findings. Документна неточність лишається явно відкритою в Discussion.

ДАЛІ: перевірити executable bit на єдиному edge-скрипті, фінальний `gofmt -l`/`go vet`, після чого завершити звіт без git і без Docker claims.

ФАКТ: `infrastructure/tests/playback-authorize.sh` зберіг executable bit; другого скрипта не створено.

ДАЛІ: фінальний host `gofmt -l .`.

ФАКТ: фінальний host `gofmt -l .` — exit 0 без виводу.

ДАЛІ: фінальний host `go vet ./...`.

ФАКТ: фінальний host `GOCACHE=/tmp/courses-t4-go-cache go vet ./...` — PASS.

ФАКТ: executable scope T4 завершено: combined real-dependency integration proof написано; наявний edge-скрипт доповнено B12/B13/B14/R29; production-код не змінювався, бо `Take`, `Authorize` і nginx seam уже реалізують контракт у місцях, названих вище.

РІШЕННЯ: NOT-список дотримано: не додано `play`/heartbeat, не змінено reader model/player/migrations/`internal/media`/nginx config/`docs/playback/*`/`work/tasks.md`/`generated/`; git-команд і коміту не було. Docker/Compose runtime не запускався й не заявлений пройденим.

ДАЛІ: рецензент запускає актуальну integration-команду та `playback-authorize.sh` з розділу «Команди рецензента», потім повний ordered gate; власник плану розв’язує Discussion D1/D2 до того, як позначати весь рядок T4 (`R01–R13`) закритим.

---

## Верифікація рецензентом (2026-09-08)

### Гейти

**Повний `make check` — чистий з першого прогону**: 334 фронт-тести, e2e **35/35**.
Виконавець чесно записав, що `format-check`, `lint` і `test` у нього не завершились через
Docker/loopback обмеження пісочниці — тут вони пройшли.

### Нові докази, прогнані рецензентом

```
B12 nginx_config=valid container=healthy cache_entries=0
B13 media_requests=3 authorize_calls=3
B14 allowed_then_anonymous_same_url=403
R29 status=403 bodies_equal=true bytes=146
```

- **B13** — кожен медіазапит доходить до `authorize`: три запити, три звернення. Склеювання
  немає, бо кеша немає.
- **B14** — анонім на URL, який щойно спрацював для іншого, дістає 403. Це рівно та дірка,
  яку створював кеш ред. 4 і яку закрило його зняття.
- **R29** — відмови **побайтово однакові** (146 байт) для аноніма й не-власника. A7
  доведено порівнянням байтів, а не твердженням.

Збережені з T1 і досі зелені: M1–M4, B37, B38, R01, R10–R13, R20.

### Discussion — обидві прийнято, обидві дефекти рецензента

**D1 — хибне покриття в таблиці тасків.** Рядок T4 приписував йому `R01–R13, R20, R29`.
R04 («архівація зупиняє перегляд понад 30 с») фізично потребує плеєра T7, а R02/R03/R05–R09
наявний скрипт не доводить. Закрити T4 «за рядком» означало б записати policy-регресії
покритими — і надалі їх ніхто б не перевіряв, бо в таблиці стоїть галочка.

Виправлено: T4 отримав **фактичний** scope (B12–B14, R29 нові; R01/R10–R13/R20 **збережені
з T1**, не нові); R04 переїхав у **T7**; R02/R03/R05–R09 — у **T5**, де справжній мінт дає
фікстури, яких T4 не мав.

**D2 — тест, що проходив би з неправильної причини.** Сценарій «чужий `videoId`» із брифу
відмовляє на **умові 2** (`application/authorize.go:24`) і до PostgreSQL не доходить. Отже
він **не доводить R07**, який вимагає збігу володіння при **недоступному курсі**, тобто
виконання умови 3. Позначити його R07 означало б лишити регресію запиту видимості курсу
назавжди зеленою.

Виконавець лишив сценарій як окремий доказ привʼязки володіння (B13) і **відмовився**
заявляти R07 закритим. У плані додано явне застереження до рядка R07.

### Код

**Продакшн-код і `nginx.conf` не змінювалися жодним рядком** — бриф це дозволяв, і виконавець
скористався дозволом замість вигадувати роботу під назву таска. Змінено три файли:
інтеграційний тест, скрипт доказів і план (правка рецензента).
