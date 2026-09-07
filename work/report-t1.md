ДАЛІ: Прочитати work/brief-t1.md, обов’язковий репозиторний контекст і контракт реалізації; після кожного кроку доповнювати цей звіт.

КРОК 1: Прочитано `work/CTO-RESEARCH.md`. Зафіксовано: Phases 1–5 закриті; медіабайти віддає лише Nginx; `/media/*` зараз безумовно 403; Go/SSR не можуть проксіювати медіабайти; SSR звертається до API напряму повз Nginx; поточна робота має спиратися на фактичний код і первинні контракти.

КРОК 2: Прочитано скіл `implement`. Він вимагає TDD на погоджених швах, регулярні точкові перевірки й typecheck, повний suite наприкінці, окреме code review та коміт поточної роботи.

КРОК 3: Прочитано `work/brief-t1.md`. Межі: реалізувати лише `GET /internal/playback/authorize`, мінімальні playback domain/application/ports/adapters, production edge-конфіг і запускабельний вимір; не робити `play`, heartbeat, progress, read model, player чи міграцію. Docker-гейти та M1–M4 у цій пісочниці не запускати, а дати точні команди. Брифова заборона «не торкатися git» має вищий пріоритет за загальну вимогу скіла `implement` зробити коміт, тому git-команд не виконуватиму.

ДАЛІ: Прочитати всі названі в брифі контракти та наявні серверні/Redis seams; зупинитися, якщо потрібного контракту бракує.

КРОК 4: Прочитано `tdd/SKILL.md` та його обов’язкові `tests.md`/`mocking.md`; `CONTEXT.md` у репозиторії не знайдено. Погоджений брифом seam — публічна поведінка `GET /internal/playback/authorize`; тести мають іти вертикальними red→green зрізами, а зовнішні Redis/Postgres залежності дозволено підміняти лише на їхніх портах.

КРОК 5: Знайдено точні межі обов’язкових розділів у playback-документах: plan D-2/D-3/D-4, acceptance B37/B38 та R01/R10–R13/R20, §11 O2; spec §3; spike §1/§8. Далі читаю їхній повний текст, а не покладаюсь на збіги пошуку.

КРОК 6: Прочитано `docs/playback/plan.md` §3 D-2/D-3/D-4, §8 і §11. Несучий порядок підтверджено: Authenticate → Redis owner match → Postgres access; недоступність залежностей = 503; edge не кешує авторизацію; криві path-id відсікаються до Go.

## Discussion

### D1 — коментар D-2 суперечить самому конфігу й брифу

ФАКТ: `docs/playback/plan.md:160-161` показує правильне `error_page 500 502 504 = @unavailable;`, але коментар на тому самому рядку каже «503 у списку — самопосилання». Насправді 503 у списку немає, і саме його додавання створило б цикл internal redirect. `work/brief-t1.md` §2.2 явно вимагає 503 не додавати.

РІШЕННЯ: У production-конфіг перенесу правильний список `500 502 504`, без 503. Запропонована поправка плану: змінити коментар на «503 у список не додавати — це самопосилання». Сам `docs/playback/plan.md` не редагую через NOT-список брифу.

КРОК 7: Прочитано `docs/playback/spec.md` §3 «Доступ» (A1–A10). Для T1 релевантні: жодних байтів аноніму/чужому, admin/owner override для неопублікованого курсу, перевірка кожного відеозапиту на edge, однакове нерозкривальне 403 для бізнес-відмов і fail-closed при недоступності Go/Redis/Postgres.

КРОК 8: Повністю прочитано `docs/playback/spike-edge.md` і `spike-edge.conf`. Переважливі виміряні факти: quoted `{32}` у rewrite; dynamic upstream через `set`+resolver; 401/403→403, 500/502/504→503; таймаути 2s/5s дали 400/400 дозволів на заглушці; криві id не звертаються до Go; без кеша кожен із чотирьох запитів звернувся до auth upstream окремо.

КРОК 9: Прочитано фактичні `AGENTS.md` і `.claude/standards.md`. T1 нижче порогу `spec-flow`; робота є малим backend+edge зрізом. Домен лишається stdlib-only, application залежить лише від domain/ports, адаптери дивляться всередину, wiring — у composition root. Невиконані гейти не можна називати пройденими.

КРОК 10: Прочитано `work/tasks.md`, plan §1–§2 і `docs/architecture/backend.md`. `work/tasks.md` уже позначає T1 `in_progress`; playback володіє записом/ownership, а доступ до навчальної поверхні не слід моделювати портом між модулями. Наявні identity Authenticate, platform Redis/httpserver є обов’язковими reuse seams.

### D2 — T1 і T4 дублюють ownership `authorize`

ФАКТ: `work/tasks.md` рядок T1 вимагає реальний `authorize` з Redis/Postgres, а T4 знову названо «`authorize` + edge-конфіг» і залежить від T1/T3. Якщо T4 реалізовувати буквально пізніше, він або дублює вже поставлений endpoint, або спонукає переробити контракт після виміру, знецінюючи M1/M2.

РІШЕННЯ: Виконую конкретний `work/brief-t1.md`: T1 ставить мінімальний production `authorize` та edge-конфіг, потрібні для справжнього виміру; не додаю T3-функції play/heartbeat/progress. Запропонована поправка трекера/плану: T4 має розширити вже наявний seam після T3, а не повторно створювати його. `work/tasks.md` не редагую через NOT-список.

КРОК 11: Перевірено наявні identity/Redis/httpserver/composition seams. `Handler.Authenticate` експортований, кладе `httpserver.Actor` у context, повертає 401 для відсутньої/недійсної сесії й 503 для збою identity Postgres. `runtime.Dependencies` уже дає `*pgxpool.Pool` і Redis client. Умова доступу до відео може бути реалізована playback Postgres-адаптером через наявні `video_assets → course_lessons → course_modules → courses`; змін у чужих модулях або generated-коді не потрібно, тож відсутнього контракту наразі немає.

КРОК 12: Прочитано production `nginx.conf`, Compose і наявні infrastructure tests. `/internal/*` зараз потрапляє у frontend через catch-all, отже R10 вже дає зовнішній 404; `/_media_auth` має бути `internal`. Наявні temp-шляхи збережено, resolver уже присутній, медіатом змонтований `ro`. Новий smoke/benchmark script має розширити `infrastructure/tests/` без зміни Docker topology.

КРОК 13: Прочитано наявні application/HTTP/Postgres тестові патерни, boundary test, Go dependencies і метрики. Обрані seams: application `Authorizer.Authorize` спостерігає строгий порядок двох портів після identity; HTTP endpoint спостерігає 204/403/503; Postgres adapter повертає лише course visibility/owner, а домен вирішує право. Для M2 буде окремий Prometheus counter саме викликів playback Postgres access check, щоб benchmark міг довести дельту 0 на випадкових playbackId.

РІШЕННЯ: TDD починаю з головної властивості M2 — після валідної identity випадковий playbackId робить один ownership read і нуль access reads. Потім вертикально додам allow/deny/dependency cases і HTTP transport.

КРОК 14 (RED, спроба 1): Додано перший application-тест `TestRandomPlaybackIDStopsAfterOwnershipRead`. `go test ./internal/playback/application` не дістався компіляції: sandbox заборонив системний Go build cache (`~/Library/Caches/go-build`). Це не доказ red; повторюю з task-local `GOCACHE` у `/tmp`.

КРОК 15 (RED): З `GOCACHE=/tmp/courses-t1-go-cache` тест упав очікувано: playback domain/application ще не існували.

КРОК 16 (GREEN): Додано мінімальні playback domain types/errors, два реальні порти (`OwnershipReader`, `VideoAccessReader`) і `Authorizer`, що припиняє виконання після mismatch. `go test ./internal/playback/application` — PASS, 0.860s. Головна властивість M2 уже піниться тестом: 1 ownership read, 0 access/Postgres reads.

КРОК 17 (RED): Додано behavior test дозволу студентові для published course з власним playbackId/videoId. Тест упав очікувано: поточний мінімальний authorizer повернув `forbidden`.

КРОК 18 (GREEN): Домен отримав правило `VideoCourse.CanBeViewedBy`: published доступний автентифікованому актору, admin/owner — незалежно від статусу. Application застосовує правило лише після успішного Redis match і access read. Точковий application suite — PASS, 0.839s.

КРОК 19 (RED): Додано adapter test для точного Redis key `playback:owner:{userId}` і JSON-контракту `{playbackId,videoId,lessonId}`. `go test ./internal/playback/adapters/redis` упав очікувано: non-test adapter code ще відсутній.

КРОК 20 (GREEN): Реалізовано Redis ownership reader; він читає один user-scoped key і мапить JSON у playback domain. Точковий Redis suite — PASS, 0.743s.

КРОК 21 (RED): Додано тест missing/expired ownership як бізнес-відмови. Він упав: перша версія помилково мапила `redis.Nil` у 503/dependency unavailable замість 403/forbidden.

КРОК 22 (GREEN): `redis.Nil` тепер мапиться у `ErrForbidden`; transport/corrupt-value failures лишаються `ErrDependencyUnavailable`. Redis adapter suite — PASS, 0.690s.

КРОК 23 (RED): Додано Postgres adapter behavior test: точний video id, join `video_assets → lesson → module → course`, явне мапування owner/status і інкремент M2 counter. Тест упав очікувано: adapter code ще відсутній.

КРОК 24 (GREEN): Реалізовано playback Postgres access reader без sqlc/generated змін. Він інкрементує окремий counter рівно перед query, повертає owner/status для доменного правила, no rows → forbidden, query failure → dependency unavailable. Postgres adapter unit suite — PASS, 0.657s.

КРОК 25 (RED): Додано HTTP contract test для `GET /internal/playback/authorize`: actor і два заголовки мапляться в application, успіх повертає порожній 204. Тест упав очікувано: HTTP adapter ще відсутній.

КРОК 26 (GREEN): Реалізовано internal HTTP adapter і реєстрацію поверх наявного `Handler.Authenticate`; allowed → порожній 204, domain denial → порожній 403, будь-яка неможливість перевірки → порожній 503, `Cache-Control: no-store`. Точковий HTTP suite — PASS, 1.300s.

### D3 — `authorize` має сам нормалізувати 401 у 403

ФАКТ: `identity/adapters/http/handler.go:123-139` повертає 401 до того, як playback handler отримає control, тоді як `work/brief-t1.md` §2.1 визначає контракт самого endpoint як 204/403/503. Якщо покласти нормалізацію лише на edge, прямий внутрішній контракт endpoint дрейфує й дві відмови лишаються різними до Nginx.

РІШЕННЯ: Playback route обгортає наявний `Authenticate` вузьким response-writer adapter, який змінює лише 401→403; 503 та інші статуси не чіпає. Тому endpoint сам дає 403, а `error_page 401 403 = @denied` на edge лишається другим поясом, як підказує бриф. Чужий identity layer не змінено.

КРОК 27 (RED): Розширено metrics contract test нульовою серією `courses_playback_authorization_database_checks_total 0`, потрібною для числового доказу M2. Тест упав очікувано: counter ще не зареєстрований.

КРОК 28 (GREEN): Додано process-scoped Prometheus counter без global registry; він зареєстрований і preinitialized. `go test ./platform/observability` — PASS, 1.283s. Benchmark зможе порівняти counter до/після 400 випадкових playbackId.

КРОК 29: Playback composition підключено в `cmd/api`: наявні request-scoped process dependencies передані Redis/Postgres adapters, далі `Authorizer`, HTTP handler і existing identity `Authenticate`. `go test ./cmd/api ./internal/playback/...` — PASS; усі playback packages компілюються й точкові suites зелені.

КРОК 30: Розширено HTTP matrix на 401→403, identity 503, domain 403, dependency/unknown failure 503 і відсутність виклику application при провалі identity. Перший прогін не скомпілювався через test-local constants, які помилково лишились тільки в application test package; виправляю локально, це не продуктова знахідка.

КРОК 31: HTTP failure matrix виправлено; `go test ./internal/playback/adapters/http` — PASS, 1.218s. Контракт endpoint тепер тестом обмежений статусами 204/403/503, а identity denial не запускає application.

КРОК 32: Додано повну application policy/order matrix: student+archived deny, owner+archived allow, admin+draft allow, Redis failure не торкає access port, Postgres failure лишається unavailable. `go test ./internal/playback/application` — PASS, 0.707s.

КРОК 33 (RED): Додано Redis corruption test. Invalid JSON уже давав 503, але syntactically valid запис без `lessonId` помилково приймався; тест упав. Такий стан не є бізнес-відмовою — це порушення storage contract і має бути 503.

КРОК 34 (GREEN): Redis adapter тепер відхиляє incomplete ownership records як dependency unavailable. Redis suite — PASS, 1.036s.

КРОК 35 (RED): Додано Postgres classification matrix. Absent video → 403 і query error → 503 вже правильні; невідомий course status помилково допускався як валідний row (і міг бути дозволений owner/admin). Тест упав на corrupt status; виправляю fail-closed.

КРОК 36 (GREEN): Postgres adapter тепер приймає лише `draft|published|archived` і non-empty owner; corrupt row → dependency unavailable. Postgres adapter suite — PASS, 0.668s.

КРОК 37: Перенесено виміряний edge block у production `nginx.conf`: media zone, lazy maps, guards до Go, безкешовий `auth_request`, правильні error mappings, 2s/5s, quoted rewrite, dynamic API upstream і deny-all fallback. Наявні temp paths не дублювались.

КРОК 38 (перевірка не відбулась): локальний `nginx -t` існує, але sandbox зупинив його на доступі до Homebrew error log/sysctl ще до валідного config verdict. Це не PASS; спробую із явним stderr/prefix, а production `nginx -t` все одно лишається Docker-гейтом B37.

КРОК 39 (перевірка не відбулась): повторний local `nginx -t -e stderr -p /tmp/...` також зупинив sandbox (`sysctlbyname(net.inet.tcp.sendspace)`, exit 1) без повідомлення про parse success/failure. B37 не позначаю пройденим.

КРОК 40: Прочитано smoke/seed і точні schema constraints для запускабельного benchmark fixture. Скрипт може детерміновано створити module/lesson/ready video row у вже seeded published course, тимчасові HLS-файли в Docker volume та user-scoped Redis ownership без міграції чи `play` endpoint.

КРОК 41: Додано executable `infrastructure/tests/playback-authorize.sh`. Він запускає `nginx -t`, перевіряє порожній cache dir; upsert-ить ізольовані module/lesson/ready asset rows; створює runtime-only HLS fixture; напряму кладе ownership у Redis; виконує M1/M2 по 400 запитів із concurrency 20; друкує p50/p95/p99/RPS/status counts, M2 Postgres delta, M3 false-denial count і M4 authorize count; перевіряє R01/R10/R20 та відмови Go/Redis/Postgres R11–R13 із відновленням сервісів у trap. `sh -n` — PASS. Сам Docker-скрипт не запускався.

КРОК 42: Усі змінені Go-файли відформатовано `gofmt`; команда завершилась exit 0.

КРОК 43: Перегнано targeted backend verification разом з architecture boundary: `go test ./platform/architecture ./internal/playback/... ./platform/observability ./cmd/api` з sandbox-local cache — PASS. Playback domain/application dependency direction машинно підтверджено.

КРОК 44 (format-check, локальні половини): `cd frontend && npm run format:check` — PASS, 140 файлів; `test -z "$(gofmt -l .)"` у backend — PASS. Сам `make format-check` не запускався, бо його Go-команда загорнута в Docker.

КРОК 45 (lint, локальні половини): `cd frontend && npm run lint` — PASS, frontend architecture checked 78 files; `go vet ./...` з local GOCACHE — PASS. Сам `make lint` не запускався через Docker wrapper Go-команди.

КРОК 46 (typecheck): `cd frontend && npm run typecheck` (`tsc --noEmit`) — PASS. Це повний non-Docker зміст `make typecheck`.

КРОК 47 (test, frontend): `cd frontend && npm test` — PASS: 38 files, 334/334 tests, 0 skipped/todo, 7.881s.

КРОК 48 (test, Go — незавершений доказ): `go test -race ./...` перевищив 30s yield; orchestration не зберіг session id у виводі. Результат не зараховую й повторюю команду окремо до завершення.

КРОК 49 (test, Go — sandbox blocker): повторний `go test -race ./...` завершився FAIL лише в наявному `platform/httpserver.TestShutdownDrainsAnInFlightResponse`: sandbox заборонив `listen tcp 127.0.0.1:0` (`bind: operation not permitted`). Усі показані до цього packages, включно з playback, PASS/cached. Повний Go suite не позначаю пройденим; запускаю всі packages, крім заблокованого loopback package, окремо.

КРОК 50 (test, Go partial): `go test -race` для всіх Go packages, крім sandbox-blocked `courses/platform/httpserver`, — PASS. Це частковий доказ; `make test` і повний `go test -race ./...` не пройшли як гейт через відсутність Docker/loopback permission відповідно.

КРОК 51: Прочитано скіл `code-review`. Його звичайна процедура потребує git fixed point/diff, але T1 NOT-список забороняє торкатися git. Тому review виконується двома незалежними осями (Standards/Spec) по явно названому поточному набору T1-файлів, без git-команд і без commit ceremony.

КРОК 52: Запущено два незалежні review-канали Standards і Spec, як вимагає `code-review`; обом явно заборонено git і редагування. Паралельно перевірено наявність `shellcheck`: binary відсутній, тому shellcheck не запускався; `sh -n` із кроку 41 лишається єдиним локальним shell parser proof.

КРОК 53: Окремо витягнуто embedded Node module з benchmark heredoc з моделюванням shell-unescape; `node --check --input-type=module` — PASS. Це синтаксичний доказ JS, не runtime/Docker proof.

КРОК 54: Обидва незалежні review-канали ще виконуються; продовжую локальну перевірку й не вважаю review завершеним до їхніх окремих висновків.

КРОК 55 (review refactor): Перейменовано застарілий `TestTranscodingMetricsArePreinitialised` на `TestMetricsArePreinitialised`, бо тест тепер також пінує playback metric. `gofmt` і `go test ./platform/observability` — PASS, 1.175s.

КРОК 56: Review-канали ще не повернули фінальних висновків після наступного очікування; роботу не завершую перед їх агрегацією.

КРОК 57: Перевірено статус агентів: обидві осі review активні, жодна не зависла й не завершилася.

КРОК 58 (typecheck target): `make typecheck` — PASS, exit 0. На відміну від format/lint/test, цей target не потребує Docker, тому виконаний саме через Makefile.

## Команди виміру M1–M4 для рецензента

ФАКТ: M1–M4 у цій пісочниці **не запускались**, бо Docker недоступний. Точний чистий прогін із кореня репозиторію:

```sh
make up
make seed-identities
make seed
sh infrastructure/tests/playback-authorize.sh
```

Скрипт друкує:

- `M1 {requests:400, concurrency:20, p50_ms, p95_ms, p99_ms, rps, statuses}` для allow;
- `M2 {requests:400, concurrency:20, p50_ms, p95_ms, p99_ms, rps, statuses, postgres_queries}` для випадкових syntactically-valid playbackId і падає, якщо `postgres_queries != 0`;
- `M3 {false_403_or_503, requests:400}` і падає, якщо live-stack allow не дав 400 × 200;
- `M4 {segment_seconds:6, segments_per_minute:10, startup_playlists:2, authorize_per_first_minute}`; очікуване й assert-нуте число першої хвилини — **12** (10 сегментів + 2 стартові плейлисти).

## Приймальні рядки T1

- B37 — `playback-authorize.sh`: container `nginx -t` + `cache_entries=0`; **не запущено тут**.
- B38 — той самий script: M1, 400 запитів / concurrency 20, вимагає 400 × 200; **не запущено тут**.
- R01 — anonymous master playlist і segment мусять дати 403; **не запущено тут**.
- R10 — зовнішній `/internal/playback/authorize` через edge мусить дати 404; **не запущено тут**.
- R11 — script зупиняє API, segment мусить дати 503, потім відновлює API; **не запущено тут**.
- R12 — script зупиняє Redis, segment мусить дати 503, потім відновлює Redis/ownership; **не запущено тут**.
- R13 — script зупиняє Postgres, segment мусить дати 503, потім відновлює Postgres; **не запущено тут**.
- R20 — два malformed path id мусять дати 403, а HTTP counter `authorize` має дельту 0; **не запущено тут**.

## Не запущені Docker-гейти

Через відсутність Docker не запускались як Make targets: `make format-check` (локальні frontend+gofmt половини PASS), `make lint` (локальні frontend+go vet половини PASS), `make test` (frontend PASS; Go partial PASS, один loopback test sandbox-blocked), `make integration`, `make check-generated`, `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` і повний `make check`. Жоден із них не позначено пройденим.

КРОК 59 (code review): Spec-вісь знайшла blocking конфлікт benchmark із media rate limit і неточну назву M2-доказу; Standards-вісь — відсутній isolated-Postgres integration test, bare HTTP error statuses і persistent benchmark fixtures. Зауваження про відсутні команди/acceptance mapping уже закрито кроком 58. Решту виправляю до повторної верифікації.

### D4 — B38 зі спайку суперечить production media rate limit

ФАКТ: `docs/playback/spike-edge.conf:5-58` не має `limit_req`, але production `infrastructure/nginx/nginx.conf:31,65` має 120r/s, burst 240 на один source IP. Непейсований прогін 400/20 з одного frontend container у `infrastructure/tests/playback-authorize.sh` може закономірно дати 429 після burst, хоча auth здоровий; тоді B38 тестує limiter, а не таймаути, і benchmark не запускабельний як задумано.

РІШЕННЯ: M1/M2 вимірюватимуть сам реальний Go endpoint напряму всередині Docker network (400/20, без edge limiter). B38 стане окремим edge-прогоном: 20 одночасних запитів у кожній пачці, пачки paced нижче 120r/s, усі 400 мусять дати 200. Запропонована поправка plan B38: явно назвати pacing або окремий source-IP budget. Plan не редагую через NOT-список.

### D5 — «Postgres не торкаються» не може означати весь request

ФАКТ: `docs/playback/plan.md:197-203` прямо каже, що умова 1 (`Authenticate`) коштує Postgres; `work/brief-t1.md` M2 одночасно просить довести, що Postgres «не торкаються». Наявний counter у `playback/adapters/postgres/access.go` рахує лише condition-3 access query, тому назва output `postgres_queries=0` була б хибно широкою: identity query уже відбулась.

РІШЕННЯ: M2 друкуватиме дві дельти: всі instrumented database queries (мають включати identity) і `condition3_postgres_queries` (мусить бути 0). Точне твердження: після успішної умови 1 випадковий playbackId коштує Redis GET і не запускає condition 3. Запропонована поправка brief: замінити «Postgres не торкаються» на «condition-3 Postgres query не виконується».

КРОК 60 (review fix): Playback handler тепер використовує спільний `httpserver.WriteError` для власних 403/503 з явними safe codes `playback_forbidden` / `dependency_unavailable` і requestId; 204 лишився без тіла. HTTP tests тепер пінують envelopes. `go test ./internal/playback/adapters/http` — PASS, 1.240s.

КРОК 61 (review fix): Додано build-tagged isolated-Postgres integration test, який підіймає PostgreSQL 17.5, застосовує реальні міграції, виконує справжній join/cast для published asset і доводить deny після `evicted_at`. `go test -tags=integration -run '^$' ./internal/playback/adapters/postgres` — compile PASS, test навмисно не запускався без Docker.

КРОК 62 (review fix): Benchmark fixture тепер має run-unique playback/module/lesson/video IDs; EXIT trap відновлює сервіси й видаляє exact Redis key, exact HLS directory та DB rows у FK-safe порядку. M1/M2 перенесено на direct internal Go endpoint із 400/20; M2 окремо друкує `all_postgres_queries` і нульову `condition3_postgres_queries`. B38 лишився edge-перевіркою 400 запитів, але йде пачками 20 з pacing 200ms (100r/s < media 120r/s); M3 береться з нього. `sh -n` і embedded `node --check` — PASS. Docker runtime не запускався.

КРОК 63 (integration review): В isolated-Postgres fixture прибрано parameterized multi-statement Exec, який pgx extended protocol міг відхилити; кожен insert тепер окрема команда. Integration package знову compile PASS із `-tags=integration -run '^$'`; runtime test не запускався.

КРОК 64: Обидві незалежні review-осі відправлено на повторну перевірку виправлень; фінальний verdict буде записано окремо, без змішування Standards і Spec.

КРОК 65 (post-review format/syntax): `gofmt -l` порожній, `sh -n infrastructure/tests/playback-authorize.sh` PASS, embedded Node `--check` PASS.

КРОК 66 (post-review backend): `go vet ./...` — PASS; `go test -race` для всіх packages крім sandbox-blocked `platform/httpserver` — PASS після review fixes.

КРОК 67 (review return): Обидві осі незалежно знайшли `paths.length` після перейменування параметра benchmark helper на `requests`; це був runtime `ReferenceError`, якого syntax check не ловить. Виправлено на `requests.length`. Standards також вказав, що fixed admin ownership/session не зберігали зовнішній стан; benchmark переведено на run-unique fixture admin, його session видаляється logout/cascade, key унікальний.

КРОК 68: Перевірено актуальний script text пошуком: `paths.length` відсутній; є лише один `authorizeCount`; output має `all_postgres_queries` і `condition3_postgres_queries`; B38 — окремий paced edge run.

## Поправка до раннього опису M1–M4 і B38

Ранній список у кроці 58 описує попередню редакцію скрипта. Актуальний контракт output:

- M1 — direct real Go `authorize`, 400/20, очікує 400 × **204**, друкує p50/p95/p99/RPS;
- M2 — direct real Go `authorize`, 400/20 випадкових playbackId, очікує 400 × 403, друкує `all_postgres_queries` (identity path) і `condition3_postgres_queries` (**0**);
- B38 — окремо 400 edge-запитів пачками по 20, pacing 200ms між пачками (номінально 100r/s < media 120r/s), очікує 400 × 200;
- M3 — кількість хибних 403/503 саме у B38, очікує 0/400;
- M4 — 12 authorize першої хвилини (2 playlists + 10 six-second segments).

Отже актуальний доказ B38 — рядок `B38` скрипта, не M1. Точна команда прогону не змінилась: `sh infrastructure/tests/playback-authorize.sh` після `make up && make seed-identities && make seed`.

КРОК 69 (review fix): Signal traps розділено: `INT`/`TERM` тепер завершують process (130/143), а єдиний EXIT trap виконує cleanup; script не може продовжити роботу після перерваного cleanup із вимкненим `-e`.

КРОК 70 (benchmark dry runtime): Embedded Node benchmark реально виконано з in-memory fetch/metrics simulator, не лише parse check. Отримано очікувану форму: M1 400×204; M2 400×403, all Postgres 400, condition-3 0; B38 400×200; M3 0/400; R20 delta 0; M4 12; R01 403/403; R10 404; exit 0. Ці latency/RPS числа належать simulator і **не є M1–M4 production заміром**; у фінальні виміри їх не підставляю.

КРОК 71: Обидві review-осі запущено на фінальний короткий verdict після останніх fixes; очікую явний Standards/Spec pass або нову знахідку.

## Code review — фінальний verdict

### Standards

PASS — isolated PostgreSQL coverage додано й compile-checked; playback HTTP errors мають safe envelopes/requestId; benchmark fixtures run-unique і self-cleaning; INT/TERM завершують process через EXIT cleanup. Матеріальних standards findings не лишилось.

### Spec

PASS — `requests.length` runtime defect закрито; M1/M2/B38/M3 форми відповідають актуальному механізму; M2 розрізняє total identity-Postgres і нуль condition-3 Postgres; точні команди й mapping усіх обов’язкових acceptance rows записані. Матеріальних T1 spec findings не лишилось.

ПІДСУМОК REVIEW: Standards 0 findings; Spec 0 findings.

КРОК 72 (final static): benchmark `sh -n`, executable-bit check, embedded Node parse і backend `gofmt -l` — PASS після всіх review fixes.

ФАКТ: Реалізовано мінімальний T1 slice: справжній 204/403/503 `authorize` у Go з несучим порядком identity → Redis → Postgres, production edge без auth cache, M2 condition-3 counter, unit/HTTP/adapter tests, isolated-Postgres integration test і self-cleaning Docker benchmark для M1–M4 та B37/B38/R01/R10–R13/R20.

РІШЕННЯ: NOT-список дотримано: не додано play/heartbeat/progress/read model/player/migration/cache; не змінено plan/spec/_wip/tasks/generated; git-команд і коміту не було.

ДАЛІ: Рецензенту з Docker виконати чотири точні команди з розділу «Команди виміру M1–M4», вставити фактичні M1–M4 числа в цей звіт і лише тоді позначати Docker acceptance/gates пройденими.

---

## Верифікація рецензентом (2026-09-08)

Виконано те, чого виконавець не міг: докерні гейти й сам замір M1–M4.

### Гейти

| Гейт | Результат |
|---|---|
| `format-check`, `lint`, `typecheck`, `test` | зелені; фронт **334** |
| `integration` | **зелений окремим прогоном** (див. нижче) |
| `check-generated`, `build`, `up`, `seed-identities`, `seed`, `smoke` | exit 0 |
| `e2e` | exit 0, **35/35** |

**Чому не одним `make check`.** Два послідовні прогони повного ланцюга впали на `integration`
з **різних середовищних причин**: спершу таймаут `host.docker.internal` при зʼєднанні з
Postgres (306 с), потім гонка ryuk-reaper (`unexpected container status "created"`, 1.4 с) —
та сама, що траплялась у цій сесії раніше. Обидва рази наступний прогін проходив: `tus` за
**6.4 с** проти 306 с. Це контенція Docker на цій машині під повним ланцюгом, не код: усі
чотири нові пакети `playback` зелені в обох прогонах. Ланцюг виконано двома частинами, кожна
exit 0, і це названо тут явно, а не подано як один чистий `make check`.

### Замір

```
M1 {"requests":400,"concurrency":20,"p50_ms":2.795,"p95_ms":8.004,"p99_ms":18.372,
    "rps":4069.59,"statuses":{"204":400}}
M2 {"requests":400,"concurrency":20,"p50_ms":3.393,"p95_ms":6.142,"p99_ms":7.703,
    "rps":3901.916,"statuses":{"403":400},
    "all_postgres_queries":800,"condition3_postgres_queries":0}
B38 {"requests":400,"concurrency":20,"p50_ms":10.014,"p95_ms":53.843,"p99_ms":99.179,
     "rps":94.561,"statuses":{"200":400},"pacing_ms_between_batches":200}
M3 {"false_403_or_503":0,"requests":400}
M4 {"segment_seconds":6,"segments_per_minute":10,"startup_playlists":2,
    "authorize_per_first_minute":12}
B37 nginx_config=valid cache_entries=0
R20 statuses=403,403 go_requests=0
R01 playlist=403 segment=403   R10 404   R11 503   R12 503   R13 503
```

**O2 закрито.** Ключове — `condition3_postgres_queries: 0` при 400 випадкових `playbackId`:
порядок умов працює, підібраний id помирає на Redis-GET. `all_postgres_queries: 800` — це
умова 1 (`Authenticate`), рівно як передбачала Discussion D5; формулювання «Postgres не
торкаються» справді було хибно широким.

**Вартість перегляду:** 12 авторизацій на першу хвилину (2 плейлисти + 10 сегментів по 6 с),
далі 10/хв. При p95 8 мс це несуттєво.

### Два дефекти у скрипті заміру, знайдені запуском

Обидва невидимі без виконання, тому виконавець їх побачити не міг:

1. **Регекс не переживав heredoc.** `<<JS` неквотований, тож `'\\s'` у файлі ставало `'\s'` у
   JS, а в одинарних лапках `\s` — це просто `s`. Вираз перетворювався на
   `...totals+([^s]+)$` і не збігався **ніколи**: метрика була на місці зі значенням 400, а
   скрипт звітував `missing metric`. Замінено на пошук рядка без регексу.
2. **Мітка `route` несе метод.** Фільтр шукав `route="/internal/playback/authorize"`, тоді як
   Prometheus віддає `route="GET /internal/playback/authorize"` — лічильник давав би нуль.

### Discussion

Усі пʼять прийнято; правки внесено в `docs/playback/plan.md` і `work/tasks.md`
(коміт `e8935f4`). D4 і D5 були дефектами **моїх** документів: без D4 рядок B38 міряв би
`limit_req`, а не таймаути; без D5 M2 доводив би твердження, яке не може бути істинним.
