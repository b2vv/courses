ДАЛІ: Прочитати work/brief-t3.md, репозиторний брифінг і правила implement; звірити контракти перед першою зміною коду.
ФАКТ: `work/report-t3.md` був відсутній; створено лише цей append-only журнал. `work/brief-t3.md`, `work/CTO-RESEARCH.md` і `work/tasks.md` прочитано; T3 має статус `planned`, HTTP/reader/player/edge/media/generated/git поза scope.
РІШЕННЯ: Виконувати через TDD на доменному й адаптерному seams та завершити окремим code review, як вимагає скіл `implement`.
ДАЛІ: Прочитати повні контракти D-1/D-4/D-5, B05/B35/B41–B44/R31/R33, G3–G6/G11–G13, код playback і звіти T1/T2; зупинитися, якщо контракту бракує.
ФАКТ: Прочитано названі розділи plan/spec, `work/report-t1.md`, `work/report-t2.md`, увесь наявний `backend/internal/playback/`, Redis client, identity config і backend standards. Підтверджено один ключ `playback:owner:{userId}`, take=`SET ... EX`, renew=`EVAL` із результатами -1/0/1, прогресовий PK `(user_id, lesson_id)`, nullable positive duration і заборону HTTP у T3.
РІШЕННЯ: T3 нижче порогу `spec-flow`: це один backend-модуль без public API/frontend. Погоджені seams: domain validation/policy; application orchestration via ownership/progress ports; Redis adapter behavior; PostgreSQL adapter behavior. Git-команди не виконуватимуться через NOT-список.
ДАЛІ: Дочитати обов’язкові TDD references і перевірити фактичні repository/config/test patterns; сформувати мінімальні публічні контракти без третього порту.
ФАКТ: Прочитано `tdd/tests.md` і `tdd/mocking.md`; `CONTEXT.md` у репо немає. Досліджено composition/config/Postgres patterns і всі згадки progress reasons. Брифом уже погоджено тестування лише через публічні domain/application/adapter seams.

## Discussion

ФАКТ: `work/brief-t3.md:39-43` і `docs/playback/plan.md:237-242` задають Lua `local v = GET ... if v ~= ARGV[1]`, але значення `v` за `work/brief-t3.md:27` / `docs/playback/plan.md:120-123` є JSON `{playbackId,videoId,lessonId}`, тоді як `ARGV[1]` названо самим `playbackId`. Наступний абзац `work/brief-t3.md:45` прямо вимагає порівняння за `playbackId`, не за цілим JSON.
ПИТАННЯ: Буквальний скрипт завжди дасть `0` для наявного коректного JSON, тому heartbeat чинного власника стане 409 і TTL ніколи не поновиться. Пропонована правка контракту: після `GET` виконати захищений `cjson.decode(v)`, звірити `stored.playbackId` з `ARGV[1]`, а malformed/missing `playbackId` класифікувати як dependency failure, не takeover.
РІШЕННЯ: Документи не редагую через NOT-список. Реалізую недвозначний намір — порівняння декодованого поля `playbackId`; окремі тести доведуть `-1`, `0`, `1` і corrupt-value failure.
ДАЛІ: Зафіксувати domain/application/port API, почати вертикальний RED→GREEN із валідації progress і трьох результатів renew.
ФАКТ: RED — додано domain-seam тести для межі `0 <= client_seq < 2^31`, закритого reason, finite non-negative position, 95.0/94.9, відомої/NULL duration, зростання позиції, monotonic seq, event regression, clamp і незворотності `completed_at`. `GOCACHE=/tmp/courses-t3-go-cache go test ./internal/playback/domain` упав на відсутньому API (`NewProgressUpdate`, `Progress`, `ApplyProgress`, `ErrInvalidProgress`), як очікувалось.
РІШЕННЯ: «Подія виходу» реалізується як event-triggered reason, тобто всі дозволені причини крім periodic `tick`; це відповідає plan B10, де до таких подій явно віднесено pause/seek/hide/nav, а `ended` є завершальною подією. Completion від reason не залежить.
ДАЛІ: Реалізувати мінімальний domain API, достатній для зеленого progress suite, без persistence/framework залежностей.
ФАКТ: GREEN — реалізовано stdlib-only `ProgressReason`, валідований `ProgressUpdate`, `Progress`, `ApplyProgress` та explicit domain errors. Після gofmt `GOCACHE=/tmp/courses-t3-go-cache go test ./internal/playback/domain` — PASS.
РІШЕННЯ: Completion використовує вже обрізану позицію, потребує відомої duration і строгого зростання проти збереженої; existing `completed_at` ніколи не скидається. Monotonic gate спершу вимагає більший seq, потім ріст позиції або event-triggered reason.
ДАЛІ: Додати RED application-тести для take, трьох renew outcomes і гарантії «не-власник не пише прогрес»; потім мінімальні ownership/progress ports і orchestration.
ФАКТ: RED — додано application tests: take передає один ownership+TTL; успішний heartbeat renew-ить і зберігає; missing→expired, mismatch→taken-over, Redis failure→unavailable; у всіх відмовах progress port не викликається. `go test ./internal/playback/application` упав на відсутніх `OwnershipRenewal` і `NewSessions`.
ДАЛІ: Додати рівно два потрібні capability contracts — ownership writer і progress writer — та мінімальний `Sessions` orchestrator; existing T1 reader/access ports не ламати.
ФАКТ: GREEN — додано `OwnershipWriter` і `ProgressWriter`, domain `OwnershipRenewal` та application `Sessions`. `Heartbeat` спершу renew-ить; missing/mismatch/dependency не можуть дійти до progress write. Після gofmt `go test ./internal/playback/application` з task-local GOCACHE — PASS.
РІШЕННЯ: Existing T1 `OwnershipReader` і `VideoAccessReader` лишено вузькими й сумісними; нового «майбутнього» порту немає, лише два capabilities з реальними Redis/Postgres залежностями.
ДАЛІ: RED→GREEN Redis adapter: безумовний SET одного JSON key з EX TTL; Lua decode+compare playbackId; числові -1/0/1 із тестом кожного; malformed storage як unavailable.
ФАКТ: RED — Redis adapter tests пінять два безумовні SET виклики на тому самому єдиному key, точний трьохпольовий JSON+TTL, mapping Lua `-1/0/1`, decoded-field comparison і dependency failures. Після виправлення лише test-local constants suite падає виключно на відсутньому `NewOwnershipWriter`.
ДАЛІ: Реалізувати Redis writer та Lua без четвертого числового результату: corrupt JSON/shape має повертати Redis error reply, нормальні гілки — рівно -1/0/1.
ФАКТ: GREEN — `OwnershipWriter.Take` робить plain `SET key JSON expiration`; `Renew` виконує Lua через `EVAL`, декодує JSON, порівнює лише `stored.playbackId`, робить `EXPIRE` лише власнику й мапить -1/0/1. Corrupt payload дає Redis error reply. Redis unit suite — PASS.
ФАКТ: Lua повертає саме три різні числові значення для нормальних станів: `-1` (ключ відсутній), `0` (playbackId чужий), `1` (TTL поновлено). `TestRenewMapsEachDistinctLuaValue` має окремий subtest для кожного.
ДАЛІ: Додати integration-tagged тест проти Redis 7.4, який виконує сам Lua й напряму спостерігає кожне з -1/0/1, TTL renewal та corrupt-value failure; compile-check без Docker.
ФАКТ: Додано `TestRenewOwnershipLuaAgainstRedis` з образом production-паритету `redis:7.4-bookworm`: три поіменні raw-EVAL subtests очікують -1/0/1, успішна гілка перевіряє TTL, corrupt JSON — dependency unavailable. `go test -run '^$' -tags=integration ./internal/playback/adapters/redis` — compile-only PASS `[no tests to run]`; runtime не запускався без Docker.
ДАЛІ: RED→GREEN PostgreSQL progress adapter через публічний `Save`: unknown duration, clamp, completion/growth, seq+position monotonicity, event regression, missing lesson і dependency failure; integration tests за чинним testcontainers pattern.
ФАКТ: RED — unit seam для PostgreSQL writer пінить параметри validated update, latest ready duration/clamp, atomic `ON CONFLICT`, seq+position guards, completion і error classification. Після виправлення test-local IDs suite падає лише на відсутньому `NewProgressWriter`.
ДАЛІ: Реалізувати один atomic PostgreSQL statement: знайти lesson+latest ready duration, clamp candidate, insert/update з monotonic WHERE, preserve completed_at; missing lesson окремо від dependency failure.
ФАКТ: GREEN — `ProgressWriter.Save` реалізовано одним atomic CTE/upsert: latest non-evicted ready duration, one write-time clamp, completion лише за threshold+growth, seq AND (growth OR event) guard, незворотний completed_at. Unit PostgreSQL suite — PASS.
ФАКТ: Додано integration-tagged `TestProgressWriterAgainstPostgreSQL` із поіменними B05/B41/B42/B44, NULL duration, stale seq, tick/event regression, preserved completion і missing lesson. `go test -run '^$' -tags=integration ./internal/playback/adapters/postgres` — compile-only PASS; runtime Docker test не запускався.
ДАЛІ: RED→GREEN typed playback config: defaults 30/180, positive decimal environment values, startup invariant `heartbeatSeconds * 3 <= ownershipTtlSeconds`; підключити loading у API composition root без HTTP endpoint.
ФАКТ: RED — config tests пінять defaults 30/180, test profile 2/6, named B43 error для 30/60 і відмову non-positive/non-decimal values. HTTP adapter suite падає на відсутніх `ConfigFromEnvironment`/`NewConfig`.
РІШЕННЯ: Назви env-полів обрано модульно й явно: `PLAYBACK_HEARTBEAT_SECONDS`, `PLAYBACK_OWNERSHIP_TTL_SECONDS`; порожні значення означають задокументовані defaults. Це не розширює product contract і лишає T5 готові getters.
ДАЛІ: Реалізувати typed config із overflow-safe invariant check; завантажити його до відкриття runtime dependencies у `cmd/api`.
ФАКТ: GREEN — додано playback HTTP-adapter config із defaults 30/180, positive integer parsing, getters і overflow-safe B43 guard (`heartbeat > ttl/3`). `cmd/api` валідовує config до відкриття Redis/Postgres. `go test ./internal/playback/adapters/http ./cmd/api` — PASS.
РІШЕННЯ: HTTP endpoints не додавались; у composition root з’явилось лише startup loading, потрібне B43.
ДАЛІ: Запустити весь playback unit suite й architecture boundary; перевірити interface satisfaction та прогнати typecheck перед інтеграційним review.
ФАКТ: `go test ./internal/playback/... ./platform/architecture` — PASS; `make typecheck` — PASS.
ФАКТ: RED regression — domain policy порівнювала raw position до clamp: saved 94, input 100, new duration 90, tick seq10 помилково приймався й записував 90. Новий тест очікувано впав (`accepted=true position=90`). PostgreSQL statement уже порівнює clamped candidate і цього дефекту не має.
РІШЕННЯ: Monotonicity оцінювати за єдиною write-time clamped position, інакше саме обрізання створює непередбачений регрес.
ДАЛІ: Перенести clamp перед position monotonicity gate у `ApplyProgress`, повторити domain і весь playback suite.
ФАКТ: GREEN regression — clamp перенесено перед position gate; domain тепер збігається з SQL для replacement case. `go test ./internal/playback/domain ./internal/playback/...` — PASS.
ДАЛІ: Перевірити integration test compilation для всього playback, host Go suite, а потім виконати дозволені ordered gates `format-check`, `lint`, `typecheck`, `test` з чесною фіксацією Docker-відмов.
ФАКТ: `go test -run '^$' -tags=integration ./internal/playback/...` — compile-only PASS для Redis/Postgres integration tests; `[no tests to run]`, runtime не називаю пройденим.
ДАЛІ: Запустити повний host `go test ./...` без integration tag із task-local GOCACHE.
ФАКТ: Повний host `go test ./...` виконав усі пакети; T3/playback і решта пройшли, але загальний exit 1 через відомий sandbox-only `platform/httpserver TestShutdownDrainsAnInFlightResponse`: bind `127.0.0.1:0` заборонено. Це НЕ повний PASS і не дефект T3.
ДАЛІ: Запустити ordered `make format-check`; якщо Docker-підкрок недоступний, зафіксувати partial result і перейти до `make lint`.
ФАКТ: `make format-check` — НЕ ЗАВЕРШИВСЯ: frontend oxfmt PASS; backend Docker `gofmt` підкрок не стартував через permission denied до Docker socket, exit 2/Make Error 1. Gate не позначаю green.
ДАЛІ: Запустити ordered `make lint`.
ФАКТ: `make lint` — НЕ ЗАВЕРШИВСЯ: frontend oxlint + architecture PASS; backend Docker `go vet` не стартував через Docker socket permission, gate не green.
ДАЛІ: Запустити ordered `make typecheck` повторно на фінальному change-set.
ФАКТ: `make typecheck` — PASS, exit 0.
ДАЛІ: Запустити ordered `make test`; очікуваний Docker blocker назвати, не обходити як пройдений gate.
ФАКТ: `make test` — НЕ ЗАВЕРШИВСЯ ще до backend Docker: frontend Rstest не зміг bind `::1:3000` (`listen EPERM`, sandbox), make exit 2. Gate не green; це середовищна відмова, не тестова assertion T3.
ДАЛІ: Запустити доступні host-equivalent `gofmt -l` і `go vet ./...`; потім виконати code review двома незалежними каналами без git, як дозволяє прецедент T2 і вимагає NOT-список.
ФАКТ: Host `gofmt -l .` — exit 0 без виводу; host `go vet ./...` — exit 0. Це додаткові checks, не заміна незавершених Docker make-gates.
ФАКТ: Запущено два незалежні read-only code reviews без git: Standards за `AGENTS.md`/`.claude/standards.md`/backend architecture+smell baseline і Spec за `work/brief-t3.md`+точними plan/spec rows. Обом передано вичерпний T3 file list і заборонено редагування.
ДАЛІ: Поки reviews працюють, локально перечитати production SQL/Lua/config/application change-set на concurrency, corrupt-state і contract drift; потім звести findings окремо.
ФАКТ: Локальний review підтвердив atomic SQL semantics і знайшов два hardening-пункти: Lua corrupt-shape guard тепер відхиляє також порожній `playbackId`; Redis integration test доводить, що result 0 не подовжує TTL чужого ключа. Видалено мертве test field. Playback unit suite і Redis integration compile-check після правок — PASS.
ДАЛІ: Отримати незалежні Standards/Spec findings; виправити підтверджені blocking проблеми й повторити фінальні доступні перевірки.
ФАКТ: Standards review: hard violation — completion/clamp/monotonic business policy дублюється в PostgreSQL SQL, тоді як domain `ApplyProgress` у production не викликається (`AGENTS.md`: policy inward). Spec review: exit override дозволяє completed lesson регреснути після повторного відкриття; domain zero-value також помилково відкидає перший seq=0. Scope creep/NOT violations не знайдено; Lua/config/ownership order підтверджено.

### D2 — буквальна формула exit override пропускає саме заборонений сценарій

ФАКТ: `docs/playback/plan.md:283-287` пояснює monotonicity тим, що completed lesson за A25 відкривається з 0 і не повинен знищити збережену позицію. Але формула `seq > saved` AND (`position > saved` OR exit event) у `work/brief-t3.md:74` дозволяє `nav`/`hide` після такого відкриття: більший seq + exit=true записує 0. Поточний тест `progress_integration_test.go` навіть очікував completed 100→10 на `nav`.
ПИТАННЯ: У runtime завершений урок після простого reopen→nav втратить позицію, всупереч прямо названій меті D-5. Пропонована правка: event override дозволяє non-growing position лише для ще незавершеного progress; якщо `completed_at` уже є, позиція й seq не регресують. Completion mark як і раніше незворотний.
РІШЕННЯ: Документи не редагую. Додаю server-side completed guard і regression test на reopen-at-zero; жодного покладання на client reason для completion.
ДАЛІ: Перебудувати progress persistence так, щоб PostgreSQL транзакційно load/lock/map/write, але єдину policy викликав із domain; одночасно змоделювати absent progress окремо, щоб перший seq=0 був валідний.
ФАКТ: RED review-fix — domain tests тепер вимагають окремий `Present` стан, прийняття першого seq=0 і відмову completed progress регресувати на event. Suite не компілюється на ще відсутньому `Progress.Present`, як очікувалось.
ДАЛІ: Реалізувати absent/present semantics і completed regression guard у єдиній `ApplyProgress` policy.
ФАКТ: GREEN — `Progress.Present` відрізняє відсутній row від persisted seq=0; перший seq=0 приймається. Event regression completed progress тепер відхиляється, тож reopen-at-zero не знищує position/seq/completed_at. Domain suite — PASS.
ФАКТ: Standards fix — SQL policy видалено. PostgreSQL adapter тепер в одній транзакції читає latest duration, створює лише transaction-local sentinel для відсутнього row, `FOR UPDATE` читає progress, викликає production `domain.ApplyProgress`, а потім лише мапить результат в UPDATE/commit. Unit adapter test доводить виклик domain completion та completed-regression no-write; suites PASS.
ФАКТ: PostgreSQL integration table оновлено сценаріями completed reopen→nav (position/seq незмінні) і first seq=0. Integration-tagged package compile-only PASS.
ДАЛІ: Закрити review finding щодо доказів B35/R31/R33 настільки, наскільки дозволяє T3 без HTTP: додати exact 2^62 unit input і combined integration-tagged Sessions test з real Redis+Postgres, що перевіряє expired/taken-over no-write та valid owner write.
ФАКТ: Додано exact B35 `client_seq=2^62` до domain validation. Додано combined integration-tagged `TestSessionsOwnershipFailuresCannotWriteProgress` з real Redis 7.4 + PostgreSQL 17.5: R31 missing→expired/no row change, R33 takeover→409-class/no row change, owner renew→write, B35 huge seq→validation/no row change.
ФАКТ: Після review fixes `go test ./internal/playback/...` — PASS; `go test -run '^$' -tags=integration ./internal/playback/...` — compile-only PASS. Runtime integration досі не запускався без Docker.
ДАЛІ: Провести короткий повторний Standards/Spec review виправленого change-set, після чого повторити фінальні host/gate checks і завершити обов’язкові таблиці звіту.
ФАКТ: Повторний Standards review закрив hard finding: PostgreSQL лишився transaction/load-lock/map-write adapter, єдина policy — `domain.ApplyProgress`. Нових blocking violations немає; дубль testcontainer setup названо non-blocking test-only judgment call за repo precedent.
ФАКТ: Повторний Spec review закрив D2, first seq=0, B35/R31/R33 evidence gaps і policy drift. HTTP 400 свідомо лишається T5 за NOT-списком. Нових blocking findings або scope creep немає.
РІШЕННЯ: Не витягувати shared integration harness у межах T3: це розширило б test infrastructure заради трьох isolated fixtures; duplication не торкається production policy й відповідає наявному pattern.
ДАЛІ: Повторити після review fixes ordered доступні gates `format-check`, `lint`, `typecheck`, `test`, host Go suite, integration compile; фіналізувати таблицю §8 і перелік незапущених гейтів.
ФАКТ: Фінальний `make format-check` — НЕ ЗАВЕРШИВСЯ: frontend oxfmt PASS; backend Docker/gofmt не стартував через Docker socket permission. Gate не green.
ДАЛІ: Фінальний `make lint`.
ФАКТ: Фінальний `make lint` — НЕ ЗАВЕРШИВСЯ: frontend lint+architecture PASS; backend Docker/go-vet не стартував через socket permission. Gate не green.
ДАЛІ: Фінальний `make typecheck`.
ФАКТ: Фінальний `make typecheck` — PASS, exit 0.
ДАЛІ: Фінальний `make test`.
ФАКТ: Фінальний `make test` — НЕ ЗАВЕРШИВСЯ: frontend Rstest зупинено sandbox bind EPERM на `::1:3000`; backend Docker tests не досягнуті. Gate не green.
ДАЛІ: Запустити фінальний host `go test ./...`, окремо playback+architecture, integration compile-only, host gofmt/go-vet.
ФАКТ: Фінальний host `go test ./...` — загальний exit 1 лише через sandbox bind EPERM у `platform/httpserver`; усі playback та інші виконані пакети green. Окремо `go test ./internal/playback/... ./platform/architecture` — PASS.
ФАКТ: `go test -run '^$' -tags=integration ./internal/playback/...` — compile-only PASS; `gofmt -l .` — exit 0 без виводу; `go vet ./...` — PASS. Runtime testcontainers не запускались.
ДАЛІ: Дописати фінальну таблицю «рядок §8 → доказ», gate matrix, reviewer runtime commands і явне підтвердження трьох Lua values.

## Рядок §8 → чим доводиться

| Рядок | Доказ T3 | Статус у цьому середовищі |
|---|---|---|
| B05 | `TestApplyProgressUsesServerCompletionAndMonotonicityRules` (95.0/94.9); PostgreSQL subtests `B05 exact threshold completes` / `below threshold remains incomplete` | unit PASS; integration compile PASS, runtime чекає Docker |
| B35 | domain validation на `2^31` і точний `2^62`; combined integration subtest `B35 huge sequence is rejected before persistence` перевіряє незмінний row | unit PASS; integration compile PASS, runtime чекає Docker; HTTP 400 належить T5 |
| B41 | domain replacement case 94/100→duration90 без completion; PostgreSQL subtest `B41 replacement restore does not complete without growth` | unit PASS; integration compile PASS, runtime чекає Docker |
| B42 | domain `playing to end completes`; PostgreSQL subtest `B42 playing to the end completes` | unit PASS; integration compile PASS, runtime чекає Docker |
| B43 | `TestPlaybackConfigRejectsBrokenCadenceAtStartup`; `cmd/api` load до runtime dependencies | unit/API compile PASS |
| B44 | domain clamp 120→90; PostgreSQL subtest `B44 position over duration is clamped` | unit PASS; integration compile PASS, runtime чекає Docker |
| R31 | application missing→`ErrPlaybackSessionExpired`+no save; raw Lua -1; combined real Redis/Postgres subtest `R31 missing playback returns expired without progress` | unit PASS; integration compile PASS, runtime чекає Docker |
| R33 | application takeover→`ErrPlaybackTakenOver`+no save; Lua 0 не renew-ить; combined subtest `R33 taken over playback cannot write progress` перевіряє незмінний PostgreSQL row | unit PASS; integration compile PASS, runtime чекає Docker |

## Lua: три різні значення

ФАКТ: `renewOwnershipScript` має рівно три числові результати нормальних станів: `-1` — key відсутній; `0` — decoded `stored.playbackId` не збігається; `1` — `EXPIRE` виконано. `TestRenewMapsEachDistinctLuaValue` має окремий unit subtest на кожне; `TestRenewOwnershipLuaAgainstRedis` виконує сам Lua проти Redis 7.4 і напряму очікує кожне з `-1/0/1`. Corrupt JSON/shape не додає четвертого числа — Lua повертає error reply, adapter мапить його в dependency unavailable.

## Гейти

| Перевірка | Результат |
|---|---|
| `make format-check` | НЕ ЗАВЕРШИВСЯ: frontend PASS; backend Docker/gofmt не стартував через Docker socket permission |
| `make lint` | НЕ ЗАВЕРШИВСЯ: frontend lint/architecture PASS; backend Docker/go-vet не стартував |
| `make typecheck` | PASS, exit 0 |
| `make test` | НЕ ЗАВЕРШИВСЯ: frontend Rstest bind `::1:3000` → EPERM; backend Docker step не досягнуто |
| `go test ./internal/playback/... ./platform/architecture` | PASS |
| host `go test ./...` | НЕ ПОВНИЙ PASS: усі playback packages green; один sandbox bind failure у `platform/httpserver` на `127.0.0.1:0` |
| host `gofmt -l .` / `go vet ./...` | PASS; не замінюють Docker halves make-gates |
| integration-tagged playback packages | compile-only PASS з `-run '^$'`; runtime НЕ ЗАПУСКАВСЯ |

Не запускались: `make integration`, `make build`, `make up`, `make e2e` — Docker недоступний за брифом; `make check-generated` — NOT-список забороняє git і generated не змінювався; `seed-identities`, `seed`, `smoke` — поза T3 та потребують Compose. Жоден skipped/blocked gate не названо passed.

## Команди рецензента для runtime integration

Запускати з `backend/` у Docker-enabled середовищі:

```sh
GOCACHE=/tmp/courses-t3-go-cache go test -count=1 -tags=integration -run '^TestRenewOwnershipLuaAgainstRedis$' ./internal/playback/adapters/redis
GOCACHE=/tmp/courses-t3-go-cache go test -count=1 -tags=integration -run '^TestProgressWriterAgainstPostgreSQL$' ./internal/playback/adapters/postgres
GOCACHE=/tmp/courses-t3-go-cache go test -count=1 -tags=integration -run '^TestSessionsOwnershipFailuresCannotWriteProgress$' ./internal/playback/application
```

ФАКТ: T3 реалізовано в дозволеному обсязі: domain/application/Redis/PostgreSQL/config, без нових HTTP endpoints. `internal/media/`, reader model, player, edge, `docs/playback/*`, `work/tasks.md`, `generated/` і git не змінювались.
РІШЕННЯ: Обидві неточності, здатні спричинити runtime bug, зафіксовано в `## Discussion`: JSON-vs-playbackId у Lua і exit override для completed reopen. Реалізація використовує decoded field та server-side completed regression guard.
ДАЛІ: Рецензент запускає три runtime integration-команди вище й повний ordered Docker gate; T5 підключає ці application/adapters до `play`/heartbeat HTTP та мапить validation/takeover/expired у 400/409/410.

---

## Верифікація рецензентом (2026-09-08)

### Гейти

**Повний `make check` — чистий з першого прогону**: 334 фронт-тести, `integration` без
флаків, e2e **35/35**. Виконавець чесно записав, що його фінальний `format-check` не
завершився через права на Docker-сокет у пісочниці — тут він пройшов.

Усі пакети playback під тегом `integration`:
```
ok  .../playback/adapters/http      ok  .../playback/adapters/postgres
ok  .../playback/adapters/redis     ok  .../playback/application
ok  .../playback/domain
```

### Три значення Lua — доведено проти справжнього Redis, поіменно

```
=== RUN TestRenewOwnershipLuaAgainstRedis/missing_key_returns_minus_one     -> -1 -> 410
=== RUN TestRenewOwnershipLuaAgainstRedis/foreign_playback_id_returns_zero  ->  0 -> 409
=== RUN TestRenewOwnershipLuaAgainstRedis/matching_playback_id_returns_one  ->  1 -> 204
=== RUN TestSessionsOwnershipFailuresCannotWriteProgress/R31_missing_playback_returns_expired_without_progress
=== RUN TestSessionsOwnershipFailuresCannotWriteProgress/R33_taken_over_playback_cannot_write_progress
```

Не «пакет зелений», а кожне значення окремим іменованим підтестом проти `redis:7.4`, і
комбінований сценарій проти живих Redis + PostgreSQL.

### Discussion — обидві прийнято, обидві дефекти рецензента

**D1 — скрипт, який не міг працювати.** План і бриф писали
`if v ~= ARGV[1] then return 0 end`, де `v` — JSON `{playbackId,videoId,lessonId}`, а
`ARGV[1]` — голий `playbackId`. Для будь-якого коректного запису порівняння істинне
**завжди** → завжди `0` → heartbeat **чинного власника** щоразу отримував би 409, TTL не
поновлювався б **ніколи**, і володіння протухало б у всіх через `ownershipTtlSeconds`.

Найгірше — **наступний абзац того самого брифу вимагав порівнювати за `playbackId`, не за
цілим JSON**. Документ суперечив власному скрипту, і скрипт прожив так **чотири редакції**:
його читали дві осі гейта, чотири лінзи й автор. Усі бачили «compare-and-renew» замість того,
що написано.

Виправлено в плані: `cjson.decode`, порівняння **поля**, пошкоджений запис → помилка (503),
а не 409.

**D2 — правило, що перемагало власну мету.** Монотонність писалась, щоб відкриття завершеного
уроку (яке за A25 починає з нуля) не знищило збережену позицію. Але формула
`seq > збереженого AND (позиція зросла OR подія виходу)` **дозволяє** `nav` після такого
відкриття записати **0** — тобто виняток відчиняв рівно ті двері, які правило зачиняло.

Знайдено **написанням тесту**, і виконавець виявив, що його ж попередній тест **очікував
неправильної поведінки** (completed 100→10 на `nav`). Він не лишив зелений тест, що пінить
баг, а переписав обидва.

Виправлено в плані: override за подією діє лише для ще **не завершеного** прогресу.

**Понад бриф:** виконавець сам додав розрізнення «рядка немає» від «збережений `seq = 0`» —
інакше правило `seq > збереженого` відкидало б **найперший** heartbeat кожного уроку. Цього
не було ні в плані, ні в брифі.

**Власне Standards-рев'ю** окремо повернуло policy з SQL у домен: адаптер тепер лише
завантажує під `FOR UPDATE`, викликає `domain.ApplyProgress` і мапить результат.
