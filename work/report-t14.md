ДАЛІ: прочитати work/brief-t14.md, обов’язковий репозиторний брифінг і скіл implement; потім звірити вимоги T14 з поточним кодом та визначити перший логічний крок.
ФАКТ: work/brief-t14.md прочитано повністю (190 рядків); T14 обмежений доказами B23 і B36 у smoke та R30 на рівні unit/lint, без змін продуктового nginx-конфігу й без Docker-прогонів у цьому середовищі.
РІШЕННЯ: виконувати через наявні шви, наприкінці застосувати code-review, зробити малий логічний коміт; не торкатися заблокованих B20/B21/B22/B28/R23, рядків T15 та заборонених документів.
ДАЛІ: повністю прочитати обов’язкові джерела репозиторію, точні приймальні рядки й п’ять питань; потім оглянути робоче дерево та суміжні тести.
ФАКТ: первинний огляд підтвердив шви брифу: `media` має `rate=120r/s`, `burst=240 nodelay` лише на `/media/p/`; auth_request-503 перетворюється nginx через 500 у `@unavailable`; Go логер передає шаблон `route`.
ФАКТ: репозиторний брифінг застарів щодо стану Phase 6, як і попереджає AGENTS.md; у цій роботі джерелом істини є поточний код і playback-план/звіт.
ДАЛІ: дочитати цільові секції плану, звіту, handoff і весь task tracker окремими неперерізаними вибірками; зняти точний стан git та знайти архітектурні тестові взірці.
ФАКТ: `work/tasks.md` підтверджує T1–T13 done, а T14 ще не заведений у tracker і за брифом tracker змінювати заборонено. До початку роботи untracked лише наданий `work/brief-t14.md` і створений журнал `work/report-t14.md`; чужих tracked-змін не виявлено.
РІШЕННЯ: R30 розмістити поруч із backend structural invariants у `backend/platform/architecture`, якщо огляд повного тестового файлу підтвердить придатний патерн; це один тест, який читає обидва джерела й атомарно захищає наскрізну log-safety властивість.
ДАЛІ: прочитати без усічення `plan.md` §8, `report.md` §2/§6, методологічну секцію handoff та архітектурний тест; перевірити Makefile/скрипти, щоб новий тест входив у дозволений Go-гейт.
ФАКТ: §8 ставить п’ять питань: що саме доводить тест; чи артефакт створений продакшн-шляхом; чи можливий зелений з неправильної причини; чи саме цей доказ закриває рядок; що мусить зламатись для червоного. Handoff додає три наскрізні перевірки: фактичний доказ, поведінка за кращого продукту, умови заміру.
ФАКТ: playback-звіт прямо позначає B23, B36 і R30 як `not verified`; §6 забороняє плутати агрегатний зелений гейт із доказом конкретного рядка.
ФАКТ: `backend/platform/architecture/boundaries_test.go` вже читає структуру репозиторію й запускається в `go test -race ./...`; отже це найближчий наявний шов для одного структурного інваріанта над nginx і Go.
РІШЕННЯ: тест R30 парситиме тільки директиву `log_format safe`, токенізуватиме її змінні та перевірятиме точний заборонений набір; Go-виклик читатиме через AST і вимагатиме пару `"route", route`, одночасно відкидаючи ключі/вирази, що несуть URI.
ДАЛІ: додати спершу структурний тест R30, потім запустити тільки його; після зеленого окремо інʼєктувати дві тимчасові поломки та зняти обидва тексти падіння.
ФАКТ: додано `TestSafeRequestLogsDoNotIncludeURI` у `backend/platform/architecture/boundaries_test.go`. Nginx-половина виділяє лише директиву `log_format safe` й порівнює точні токени змінних; Go-половина парсить AST, знаходить єдиний `Info("http request", ...)`, вимагає `"route", route` і відкидає URI-ключі/читання `r.URL` або `r.RequestURI`.
РІШЕННЯ: перелік nginx для R30 — `$request_uri`, `$uri`, `$request`, `$args`, `$query_string`: це саме повний перелік альтернатив, названий контрактом T14; точне токенізування не плутає небезпечний `$request` із безпечними `$request_id`/`$request_method`.
ДАЛІ: відформатувати лише змінений Go-файл через gofmt і запустити цільовий тест R30 без Docker.
ФАКТ: `gofmt` застосовано лише до зміненого Go-тесту; `git diff --check` для нього завершився без помилок.
ДАЛІ: запустити `go test ./platform/architecture -run TestSafeRequestLogsDoNotIncludeURI -count=1` із каталогу backend.
ФАКТ: перший цільовий запуск не дістався компіляції: Go намагався писати у sandbox-заборонений `/Users/strelia/Library/Caches/go-build` і завершився `operation not permitted` (`FAIL [setup failed]`). Це відмова середовища, не результат тесту.
РІШЕННЯ: повторити той самий тест із task-specific `GOCACHE=/tmp/courses-t14-gocache`; не змінювати HOME і не трактувати перший запуск як доказ.
ДАЛІ: повторити цільовий тест із доступним кешем.
ФАКТ: `GOCACHE=/tmp/courses-t14-gocache go test ./platform/architecture -run TestSafeRequestLogsDoNotIncludeURI -count=1` пройшов: `ok courses/platform/architecture 0.688s`.
ДАЛІ: тимчасово додати `$request_uri` саме до `log_format safe`, запустити R30-тест, записати очікуване падіння й одразу повернути nginx.conf точно до початкового стану.
ФАКТ: для mutation-check тимчасово додано поле `"uri":"$request_uri"` саме в директиву `log_format safe`; це не продуктова зміна й має бути зняте після заміру.
ДАЛІ: запустити лише R30-тест та перевірити, що він червоний через названу nginx-змінну.
ФАКТ: nginx mutation-check очікувано впав (exit 1): `boundaries_test.go:72: log_format safe must not include URI-bearing nginx variable $request_uri`.
ДАЛІ: прибрати тимчасове URI-поле з nginx.conf та перевірити, що файл не лишився у продуктовій дельті.
ФАКТ: nginx mutation повністю знято; `git diff --exit-code -- infrastructure/nginx/nginx.conf` завершився exit 0.
ДАЛІ: тимчасово замінити Go-пару `"route", route` на читання `r.URL.Path`, запустити R30-тест, записати падіння й повернути server.go точно як був.
ФАКТ: для другого mutation-check тимчасово замінено `"route", route` на `"path", r.URL.Path` у Go request logger; зміна не входить у результат T14.
ДАЛІ: запустити лише Go-половину R30 та перевірити, що вона називає URI-bearing поле/відсутній шаблон route.
ФАКТ: Go mutation-check очікувано впав (exit 1) із трьома діагностичними рядками: `must not include URI-bearing field "path"`; `field "path" must not read the request URI`; `must log the matched template as "route", route`.
ДАЛІ: повернути Go logger до `"route", route`, перевірити нульову дельту server.go, потім знову запустити повний R30-тест на відновлених джерелах.
ФАКТ: тимчасову Go-поломку знято, і `git diff --exit-code -- backend/platform/httpserver/server.go` пройшов. Комбінована команда далі не запустила тест, бо її помилково виконано з кореня репозиторію, де немає `go.mod`: `go: cannot find main module`; це не результат R30.
ДАЛІ: повторити повний R30-тест із правильного каталогу `backend`.
ФАКТ: після відновлення обох джерел повний R30-тест знову зелений: `ok courses/platform/architecture 0.564s`. Обидві тимчасові поломки окремо доведено червоними й повністю прибрано.
ДАЛІ: оглянути Compose-стан/healthcheck Redis і наявні recovery-патерни в smoke-скриптах; потім додати B36 з безумовним відновленням та B23 після нього, щоб burst не завадив переходу `503 → 403`.
ФАКТ: Redis має healthcheck, але `docker compose start redis` не чекає його; наявний playback-authorize smoke чекає Redis/API окремо, проте T14 вимагає сильніший зовнішній доказ тим самим media-шляхом.
РІШЕННЯ: B36 виконати перед B23. На EXIT/HUP/INT/TERM cleanup безумовно стартує Redis і опитує той самий URL до 403; нормальний шлях робить те саме, знімає trap лише після 403, тоді запускає rate burst. Так 429-борг не маскує 503/403.
ДАЛІ: додати до `smoke.sh` B36 із числовим `body.byteLength === 0`, наявним `Retry-After`, переходом 503→403 і cleanup; після нього додати B23 burst до `/media/p/`.
ФАКТ: `smoke.sh` тепер перевіряє B36 трьома спостережуваними — статус 503, присутній `Retry-After`, `body.byteLength === 0` — після `docker compose stop redis`; далі стартує Redis і опитує той самий шлях кожні 250 мс до 403 або 30 с.
ФАКТ: EXIT/HUP/INT/TERM trap тримає прапорець зупиненого Redis і виконує те саме відновлення навіть якщо B36 або проміжна команда впаде; trap знімається лише після спостереженого 403. Відтак наступні гейти не успадковують навмисно зупинену залежність.
ФАКТ: B23 додано після B36 як 720 одночасних GET саме до валідно сформованого `/media/p/<playbackId>/hls/<videoId>/...`; перевіряється наявність хоча б одного 429, а очікувані 403 не забороняються.
РІШЕННЯ: `N=720`: початковий `burst=240`; решта 480 потребують `480 / 120r/s = 4s` поповнення, щоб усі пройшли. `Promise.all` створює один локальний сплеск, тому принаймні один запит мусить перевищити бюджет; число втричі більше burst і не вимагає припущення про валідну сесію.
ДАЛІ: статично перевірити shell-синтаксис, формат/дельту та уважно прорецензувати cleanup на збереження первинного exit status і відсутність взаємного впливу B36/B23.
ФАКТ: `sh -n infrastructure/tests/smoke.sh` і `git diff --check` пройшли. Цільова дельта наразі — один Go structural test і 66 рядків smoke; у worktree паралельно з’явилися чужі tracked-зміни в `AGENTS.md`, `work/AGENDA.md`, `work/CTO-RESEARCH.md`, яких T14 не торкається і не включатиме в коміт.
РІШЕННЯ: cleanup зберігає первинний exit status, на помилці вимикає рекурсивні traps, робить best-effort start+503→403 poll і лише тоді виходить. B23 стоїть після успішного 403, тому його 720 запитів не можуть підмінити B36 кодом 429.
ДАЛІ: посилити аварійний шлях: позначати Redis як такий, що потребує відновлення, ще до `docker compose stop`, аби навіть частково невдалий stop завершився спробою start; повторити статичні перевірки.
ФАКТ: прапорець `redis_stopped=true` тепер встановлюється до stop; отже будь-який вихід після початку руйнівного кроку запускає відновлення. Повторні `sh -n` і `git diff --check` пройшли (exit 0).
ДАЛІ: скласти обов’язкову таблицю доказів B23/B36/R30 з п’ятьма питаннями, явно позначивши B23/B36 як очікування до Docker-гейта, а R30 — як виміряний доказ.

ФАКТ: матриця «рядок → чим доводиться» за п’ятьма питаннями §8:

| Рядок | 1. Що доводить | 2. Походження артефакту | 3. Хибно-зелена причина | 4. Саме цей рядок | 5. Що робить червоним |
|---|---|---|---|---|---|
| B23 | Smoke над production nginx показує хоча б один 429 у зоні `media` | Байти не потрібні: валідно сформований шлях іде тим самим `/media/p/` location; rate limit працює у preaccess до auth | Не б’є в загальний `/media/`; не вимагає 200 чи відсутності нормальних 403; 429 може дати лише ліміт цільового location | Так: окрема перевірка `Media rate limit not enforced` відповідає B23 | Видалення/перенесення `limit_req zone=media`, або бюджет, який пропустить усі 720, прибере 429 і впустить тест |
| B36 | Smoke при зупиненому Redis вимагає 503, наявний `Retry-After`, рівно 0 байтів тіла; після start той самий шлях мусить перейти у 403 | Реальні Compose-сервіси, production nginx/auth_request і той самий media URL; медіафайл не потрібен, бо відмова залежності передує читанню байтів | Не приймає сам status: перевіряє header+числову довжину; відновлений 403 відрізняє політику від dependency failure; `error_page` не змінено | Так: усі три спостережувані B36 і обов’язкове відновлення живуть в одному блоці | Інший статус, зниклий `Retry-After`, будь-яке непорожнє тіло або відсутність переходу до 403 за 30 с роблять smoke червоним |
| R30 | Unit/structural test читає лише `log_format safe` і AST єдиного Go `http request` logger; nginx не має URI-змінних, Go має `"route", route` | Артефакти — самі production source-файли `nginx.conf` і `server.go`, а не копії | Точні токени не плутають `$request` із `$request_id`; тест не сканує весь nginx-файл; дві mutation-поломки реально дали red | Так: тест названий `TestSafeRequestLogsDoNotIncludeURI` і перевіряє обидві половини R30 | Додавання `$request_uri` дало `must not include ... $request_uri`; заміна на `"path", r.URL.Path` дала три URI/route помилки |

РІШЕННЯ: якщо продукт «кращий», перевірки не карають його за більше 429 чи швидше відновлення: B23 просить лише наявність 429, B36 приймає будь-який момент переходу до 403 в межах 30 с. Умови заміру явно обмежені: B23/B36 — Compose smoke (тут не запускається), R30 — джерельний unit-тест (тут запущений).
ФАКТ: числове очікування рецензента для B23 — `N=720`, `>=1` відповідь 429; 403 дозволені. Для B36 — послідовність `503`, `Retry-After != null`, `body.byteLength=0`, потім той самий URL дає `403` у межах 30 с при polling кожні 250 мс.
ДАЛІ: перевірити наявність Docker як окремий факт, не запускаючи заборонені Docker-гейти; потім виконати дозволені гейти у заданому порядку.
ФАКТ: у PATH є Docker-клієнт (`docker-client-present`), але бриф прямо фіксує відсутність доступного Docker-середовища і забороняє тут `integration/build/up/seed/smoke/e2e/check-generated`; сам факт бінарника не є доказом доступного daemon і не змінює обмеження.
ФАКТ: Docker-залежні Go-тести відокремлені build tag `integration`, тому звичайний `go test -race ./...` є дозволеним Go unit-гейтом без Docker.
ДАЛІ: запустити `npm run format:check` у frontend; не запускати форматер із записом.
ФАКТ: `npm run format:check` пройшов (exit 0): 164 файли, усі у правильному форматі. `npm run format` не запускався.
ДАЛІ: запустити frontend `npm run lint` без `--fix`.
ФАКТ: `npm run lint` пройшов (exit 0); oxlint без warnings, architecture boundaries перевірено для 91 файла.
ДАЛІ: запустити `npm run typecheck` (`tsc --noEmit`).
ФАКТ: `npm run typecheck` пройшов (exit 0), `tsc --noEmit`.
ДАЛІ: запустити frontend `npm test`.
ФАКТ: `npm test` пройшов (exit 0): 45 файлів, 379/379 tests passed, 0 skipped/todo, snapshots без змін.
ДАЛІ: запустити весь дозволений Go unit-гейт `go test -race ./...` без integration build tag, із task-specific GOCACHE.
ФАКТ: `go test -race ./...` виконав усі звичайні Go-пакети, але загальний exit 1 через sandbox-заборону локального bind у наявному `platform/httpserver` тесті: `TestShutdownDrainsAnInFlightResponse: listen tcp 127.0.0.1:0: bind: operation not permitted`. Усі інші показані пакети, включно з `courses/platform/architecture`, пройшли; це відмова середовища, не зелений повний Go-гейт.
РІШЕННЯ: не маскувати падіння і не послаблювати тест. Окремий R30-тест уже пройшов; для повнішого локального сигналу запустити решту httpserver unit-тестів без єдиного sandbox-заблокованого bind-сценарію, явно не називаючи це повним гейтом.
ДАЛІ: перелічити тести `platform/httpserver`, потім виконати пакет із regex-виключенням лише `TestShutdownDrainsAnInFlightResponse`.
ФАКТ: `go test ./platform/httpserver -list '^Test'` пройшов і перелічив п’ять тестів; єдиний bind-сценарій — `TestShutdownDrainsAnInFlightResponse`, решта: ActorFromContext, HealthAndReadiness, RequestObservabilityExcludesSecrets, SafeErrorsAndCorrelation.
ДАЛІ: запустити чотири незаблоковані httpserver-тести з `-race`.
ФАКТ: чотири незаблоковані `platform/httpserver` тести пройшли з race (exit 0, 2.529s). Повний Go-гейт усе одно лишається незеленим через один sandbox-заблокований bind-тест.
ДАЛІ: застосувати обов’язковий зі скілу implement фінальний `code-review`: спершу прочитати інструкцію скілу повністю, потім оглянути лише дельту T14 проти HEAD і виправити blocking-знахідки.
ФАКТ: `code-review/SKILL.md` прочитано повністю. Він вимагає дві незалежні осі — Standards і Spec — у паралельних субагентах та фіксовану базу.
РІШЕННЯ: база — поточний pre-T14 `HEAD`; оскільки реалізація ще навмисно не закомічена, оглядатиметься working-tree diff лише трьох T14-файлів плюс новий journal, а не чужі паралельні tracked-зміни. Джерело spec — `work/brief-t14.md`; стандарти — `AGENTS.md` і smell baseline скілу.
ДАЛІ: зафіксувати SHA HEAD і цільову дельту, потім паралельно запустити Standards та Spec reviews.
ФАКТ: fixed point резолвиться у `fd2482c469094b5c60450f6ab3e4c4033fcdf294`; цільова code delta непорожня: +134 рядки structural test, +66 рядків smoke; `work/brief-t14.md` і append-only `work/report-t14.md` нові. Комітів після бази ще немає, бо review передує коміту.
ДАЛІ: запустити два незалежні read-only reviews без редагування файлів.
ФАКТ: перший виклик Standards reviewer не стартував через технічне обмеження імені агента (`agent_name must use only lowercase letters, digits, and underscores`); жодного review ще не виконано і файли не змінено.
РІШЕННЯ: повторити з допустимими underscore-іменами й тим самим scope.
ДАЛІ: запустити обидві review-осі паралельно.
ФАКТ: два незалежні read-only subreviews запущено паралельно: Standards проти AGENTS.md+smell baseline і Spec проти повного `work/brief-t14.md`; обом задано ігнорувати чужі зміни й не редагувати файли.
ДАЛІ: дочекатися обох висновків, записати їх окремо без злиття пріоритетів і виправити лише підтверджені blocking-знахідки.
ФАКТ: Standards review завершено: hard-порушень AGENTS.md немає. Є одна неблокуюча judgment-call знахідка `Duplicated Code`: довгий media URL повторено у трьох Node heredoc; reviewer вважає залишення захищених quoted-heredoc літералів виправданим, хоча є малий synchronization risk.
РІШЕННЯ: не передавати URL через shell interpolation/env лише заради DRY: це додасть межу екранування й приховає явну test fixture; повтор у трьох ізольованих процесах лишається свідомим і неблокуючим.
ДАЛІ: дочекатися Spec review.
ФАКТ: Spec review завершено з двома знахідками, scope creep — 0.

## Standards

- Hard findings: 0.
- Non-blocking: `Duplicated Code` для трьох явних копій media URL у різних quoted heredoc; свідомо прийнято, бо безпечне винесення створює зайву межу передачі session/path між процесами.

## Spec

- **Blocking:** B36 без cookie зупиняється в identity middleware як 403 до звернення в Redis; потрібна валідна сесія, створена до stop, а після recovery невідоме playback ownership має дати 403.
- **Blocking/partial:** R30 ловить пряме `r.URL`, але може пропустити URI через проміжну змінну й новий ключ; потрібен fail-closed allowlist точних безпечних key/value пар.
- Scope creep: 0.

РІШЕННЯ: обидві Spec-знахідки підтверджені. Перенести наявну admin-сесію першого smoke-блоку через mode-0600 файл у tmpfs frontend-контейнера, не робити повторний login після насичення login rate limit; logout+unlink виконувати під тим самим cleanup. Go logger перевіряти точним allowlist поточної безпечної форми.
ДАЛІ: переробити R30 на allowlist, повторити green і обидва mutation-red; потім виправити lifecycle cookie для B36 та оновити матрицю звіту.
ФАКТ: Go-половину R30 переведено на fail-closed allowlist шести точних key/value пар (`method`, `route`, `status`, `durationMs`, `requestId`, `traceId`). Будь-яке нове поле або підміна значення тепер червоніє до явного security-review, тому непрямий `requestPath` не пройде під довільним ключем.
ДАЛІ: gofmt + цільовий green, після чого повторно заміряти обидві мутації вже фінальною версією тесту.
ФАКТ: фінальний allowlist-тест відформатовано і він зелений: `ok courses/platform/architecture 0.765s`.
ДАЛІ: повторити nginx mutation `$request_uri`, зафіксувати red і відновити файл.
ФАКТ: фінальний nginx mutation-check знову червоний (exit 1): `boundaries_test.go:74: log_format safe must not include URI-bearing nginx variable $request_uri`.
ДАЛІ: негайно прибрати тимчасове поле й підтвердити нульову nginx-дельту.
ФАКТ: nginx mutation знято; `git diff --exit-code -- infrastructure/nginx/nginx.conf` exit 0.
ДАЛІ: перевірити сильніший непрямий Go mutation: присвоїти `r.URL.Path` у `requestPath` і замінити route-пару на `"urlPath", requestPath`; зафіксувати red фінального allowlist та відновити server.go.
ФАКТ: непрямий Go mutation-check червоний (exit 1): `contains unapproved field "urlPath"; review it for URI data before adding it to the safe allowlist` і `is missing approved field "route"`. Отже обхід через локальну змінну більше не зелений.
ДАЛІ: прибрати `requestPath`, повернути `"route", route`, підтвердити нульову server.go-дельту й green фінального R30.
ФАКТ: server.go повністю відновлено (`git diff --exit-code` exit 0); фінальний R30 з обома чистими джерелами зелений: `ok courses/platform/architecture 0.514s`.
ДАЛІ: змінити smoke lifecycle так, щоб уже створена до login-burst admin-сесія дожила до B36, зберігалася mode 0600 лише у frontend `/tmp`, передавалась у 503 і recovery-403 fetch, а cleanup завжди робив logout+unlink.
ФАКТ: B36 тепер використовує валідну admin-сесію: перший Node-блок не logout-ить її, а після усіх своїх перевірок пише тільки cookie у `/tmp/courses-smoke-media-cookie` з mode 0600 всередині tmpfs frontend-контейнера. І 503-запит, і recovery polling читають cookie та надсилають його.
ФАКТ: cleanup розширено: після гарантованого Redis start+403 він робить session logout і unlink cookie-файлу; первинна помилка має пріоритет, але помилка restore чи cleanup також дає ненульовий exit на інакше зеленому шляху. На нормальному шляху trap знімається лише після обох відновлень.
РІШЕННЯ: повторний login після 30-запитного login burst не використовується, бо сам отримав би 429. Невідомий playbackId з живою автентифікованою сесією доходить до playback/Redis і після відновлення закономірно дає policy 403.
ДАЛІ: запустити `sh -n`, перевірити відсутність логування cookie, звірити identity/auth order у коді й оновити матрицю B36 про автентифікаційну передумову.
ФАКТ: `sh -n` і `git diff --check` пройшли; identity `Authenticate` справді відхиляє порожню cookie до application handler, тож reviewer знайшов реальний blocker. Пошук виводу cookie не знайшов збігів; саме тому комбінована read-only команда завершилася exit 1 на фінальному `rg` (no matches), а не на syntax/diff check.
ФАКТ: поправка до попередньої матриці B36: production-path походження включає валідну session cookie, створену до зупинки Redis; без неї тест доводив би лише ранню anonymous policy 403. Поточний тест з cookie доходить до Redis-залежного authorize, а після AOF-відновлення та сама cookie + невідоме ownership дають 403.
ДАЛІ: провести власний статичний review фінального smoke control flow і allowlist-тесту, потім повторити дозволені гейти, яких торкнулися виправлення.
ФАКТ: self-review знайшов cleanup-взаємодію: logout через nginx міг успадкувати насичену раніше зону `api`. Cleanup переведено на внутрішній `http://api:8080` з тим самим cookie й Origin; це не доказ acceptance, а гарантоване прибирання тестової сесії після відновлення Redis.
РІШЕННЯ: acceptance-запити B36 лишаються через production nginx; лише несучий cleanup обходить уже навмисно навантажений edge budget.
ДАЛІ: повторити `sh -n`, `git diff --check`, цільовий R30 та повні дозволені frontend-гейти; Go full race вже має зафіксовану sandbox-відмову, тому повторити target package з race після фінальної зміни тесту.
ФАКТ: перша комбінована фінальна перевірка не стартувала через хибний cwd `backend` для кореневого шляху `infrastructure/tests/smoke.sh`: `No such file or directory`, exit 127. Жоден із трьох checks цієї команди не зараховано.
ДАЛІ: повторити syntax/diff з кореня, а Go target — окремо з backend.
ФАКТ: фінальні `sh -n infrastructure/tests/smoke.sh` і scoped `git diff --check` пройшли (exit 0).
ДАЛІ: запустити фінальний R30 target із race з каталогу backend.
ФАКТ: фінальний `go test -race ./platform/architecture -run TestSafeRequestLogsDoNotIncludeURI -count=1` пройшов: `ok ... 1.815s`.
ДАЛІ: повторити frontend format-check → lint → typecheck → test одним строго впорядкованим read-only ланцюгом; frontend після review не змінювався, це фінальне підтвердження.
ФАКТ: фінальний frontend-ланцюг: format-check, lint і typecheck пройшли; `npm test` цього разу не стартував через sandbox bind `listen EPERM: operation not permitted ::1:3000` (exit 1). Раніше в цій сесії той самий `npm test` пройшов 379/379; повторна відмова — середовищна, але весь фінальний ланцюг не зелений.
РІШЕННЯ: один раз повторити лише `npm test`, не змінюючи конфіг/таймаути; обидва результати лишаються у звіті.
ДАЛІ: повторити `npm test`.
ФАКТ: повторний `npm test` пройшов (exit 0): 45 файлів, 379/379 passed, 0 skipped/todo, snapshots без змін. Попередній EPERM запуск лишається зафіксованим як середовищна відмова.
ДАЛІ: повернути фінальну дельту тим самим Standards/Spec reviewers для перевірки виправлень; жодних нових правок до їхнього вердикту.
ФАКТ: обидві review-осі повторно запущено на фінальній дельті після виправлень, read-only.
ДАЛІ: дочекатися обох фінальних вердиктів.
ФАКТ: Standards re-review — 0 findings. Spec re-review підтвердив усе, крім одного non-blocking, але реального parser blind spot: `index(..., ";")` вважає термінатором крапку з комою всередині quoted log value, тому `'{"note":"x;y","uri":"$request_uri"}'` міг сховати URI після `x;`.
РІШЕННЯ: виправити, а не приймати: завершальна `;` означає першу крапку з комою поза одинарними/подвійними nginx quotes; quote/escape state має переживати багаторядкову директиву.
ДАЛІ: додати POSIX-awk пошук unquoted terminator, перевірити baseline і додаткову mutation з quoted semicolon перед `$request_uri`, потім ще раз повернути zero diff nginx.conf.
ФАКТ: `directive_terminator` тепер посимвольно трекає одинарні/подвійні quotes і backslash escape крізь рядки; `;` завершує захоплення лише поза quotes. Одинарну лапку задано через `sprintf("%c", 39)`, щоб зберегти shell single-quoted awk program POSIX-сумісним.
ДАЛІ: запустити `sh -n` і baseline host integration; green визначати як відсутність R30 error та досягнення Docker-межі, навіть якщо Docker API знову недоступний.
ФАКТ: quote-aware версія пройшла `sh -n`; baseline host R30 зелений і скрипт дійшов до Docker API, де отримав очікуваний sandbox `permission denied` (overall exit 1, але жодного `nginx log_format safe...` error).
ДАЛІ: додаткова mutation — вставити quoted `"note":"x;y"` перед `"uri":"$request_uri"`, прогнати `sh infrastructure/tests/integration.sh` і вимагати, щоб внутрішня `;` не сховала URI.
ФАКТ: додаткову quoted-semicolon mutation прогнано точно `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker з `nginx log_format safe contains unapproved variable $request_uri`. Внутрішня `;` у `"x;y"` більше не завершує директиву.
ДАЛІ: повернути nginx.conf точно як був і підтвердити `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: quoted-semicolon mutation знято; nginx.conf zero diff, `git diff --exit-code` exit 0.
РІШЕННЯ: після зміни parser state повторити обидві обов’язкові мутації ще раз фінальною реалізацією, щоб старі red не спиралися на попередню версію awk.
ДАЛІ: фінально переміряти однорядкову `$request_uri` mutation командою host integration, повернути zero diff.
ФАКТ: фінальна quote-aware однорядкова mutation прогнана `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker: `nginx log_format safe contains unapproved variable $request_uri`.
ДАЛІ: зняти mutation і підтвердити zero diff nginx.conf.
ФАКТ: фінальну однорядкову mutation знято; nginx.conf zero diff, exit 0.
ДАЛІ: фінально переміряти дворядкову mutation з `$request_uri` на другому рядку, повернути zero diff.
ФАКТ: фінальна quote-aware дворядкова mutation прогнана `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker: `nginx log_format safe contains unapproved variable $request_uri`.
ДАЛІ: зняти mutation, підтвердити zero diff nginx.conf і повторити фінальний baseline host R30.
ФАКТ: фінальну дворядкову mutation знято; `git diff --exit-code -- infrastructure/nginx/nginx.conf` exit 0. Наступний baseline `sh infrastructure/tests/integration.sh` пройшов повністю, exit 0, включно з host R30 та всіма Docker integration-пакетами.
ДАЛІ: короткий фінальний Spec/Standards re-review quote-aware parser; після вердикту дописати підсумок цього кола без commit.
ФАКТ: обидва фінальні read-only re-reviews запущено; git writes і smoke виключені.
ДАЛІ: дочекатися вердиктів.
ФАКТ: фінальний Standards re-review — 0 findings. Spec re-review знайшов ще один blocking parser edge: `#` поза quotes завершує nginx-рядок, тому `;` у коментарі не є термінатором; поточний scanner міг зупинитися на comment semicolon і не прочитати `$request_uri` на наступному рядку.
ФАКТ: друга заявлена ним «documentation gap» не відповідає файлу: записи про quote-aware fix, quoted-semicolon mutation, повтор обох required mutations, zero diffs і фінальний integration exit 0 уже фізично стоять перед рядком `ДАЛІ: дочекатися вердиктів.`. Нічого не переписується; цей факт додано на EOF.
РІШЕННЯ: додати comment-aware boundary: `#` поза quotes обриває лише фізичний рядок, не директиву; `#` у quoted value лишається даними. Переміряти спеціальну comment-semicolon mutation.
ДАЛІ: виправити awk scanner для nginx comments, baseline, mutation `log_format safe ... # explanatory ;` + `$request_uri` на наступному рядку, відновлення zero diff.
ФАКТ: scanner тепер повертає `scanned_length`: unquoted `#` обрізає лише поточний фізичний рядок і не ставить `complete`; unquoted `;` завершує директиву; quoted/escaped `#` та `;` не мають синтаксичного ефекту.
ДАЛІ: `sh -n` і baseline host integration на чистому nginx.conf.
ФАКТ: перша спроба дописати цей запис не змінила файлів через typo шляху `workE/work/report-t14.md` (`No such file or directory`). Фактичний baseline braced-aware scanner пройшов `sh -n`; host R30 зелений і дійшов до Docker API, де sandbox дав `permission denied` (overall exit 1, без R30 error).
ДАЛІ: тимчасово додати `${request_uri}` у safe format, прогнати `sh infrastructure/tests/integration.sh`, повернути конфіг і zero diff.
ФАКТ: braced mutation прогнано точно `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker: `nginx log_format safe contains unapproved variable ${request_uri}`.
ДАЛІ: повернути nginx.conf, підтвердити zero diff і фінальний baseline.
ФАКТ: braced mutation знято; nginx.conf zero diff (exit 0). Фінальний baseline `sh infrastructure/tests/integration.sh` пройшов повністю, exit 0, включно з host R30 та Docker integration-пакетами.
ДАЛІ: фінальний read-only review braced/comment/quote-aware scanner; після нього лише EOF-підсумок, без нових git або smoke дій.
ФАКТ: останні Standards/Spec re-reviews запущено read-only; поточний фізичний EOF-якір унікальний.
ДАЛІ: дочекатися фінального вердикту.
ФАКТ: comment-aware scanner пройшов `sh -n`; baseline host R30 зелений і дійшов до Docker API, де sandbox дав `permission denied` (overall exit 1, без R30 error).
ДАЛІ: внести comment-semicolon mutation з `$request_uri` на наступному рядку й прогнати точну host integration-команду.
ФАКТ: comment-semicolon mutation прогнано `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker: `nginx log_format safe contains unapproved variable $request_uri`. `;` у comment не сховала другий рядок.
ДАЛІ: повернути nginx.conf, zero diff, фінальний baseline і короткий re-review.
ФАКТ: comment mutation знято; nginx.conf zero diff (exit 0). Фінальний baseline `sh infrastructure/tests/integration.sh` пройшов повністю, exit 0, включно з host R30 і Docker integration-пакетами.
ДАЛІ: останній read-only Standards/Spec re-review comment-aware parser; після нього дописати підсумок і зупинитися без git commit.
ФАКТ: останні два read-only re-reviews запущено; git writes і smoke виключені.
ДАЛІ: дочекатися остаточних вердиктів.
ФАКТ: фінальний Standards re-review: hard findings 0; non-blocking — allowlist-поля перелічено двічі (реальний synchronization seam) і свідомий повтор media URL. Фінальний Spec re-review: кодових знахідок 0; єдина report-знахідка — бракує консолідованого явного списку run/not-run гейтів.
РІШЕННЯ: прибрати подвійний список allowlist через ітерацію одного map; URL не виносити з quoted heredoc. Додати нижче повний gate inventory з буквальним статусом Docker-гейтів.
ДАЛІ: усунути дубль allowlist, gofmt і повторити R30 race target.
ФАКТ: duplicate allowlist seam усунуто: обов’язковість полів перевіряється ітерацією того самого `expected` map. Після gofmt R30 race target зелений: `ok courses/platform/architecture 1.807s`.

ФАКТ: консолідований перелік гейтів:

| Гейт / перевірка | Статус |
|---|---|
| `npm run format:check` | пройшов двічі; 164 файли |
| `npm run lint` | пройшов двічі; 91 architecture files |
| `npm run typecheck` | пройшов двічі |
| `npm test` | пройшов: 379/379, 0 skipped/todo; один проміжний повтор не стартував через sandbox `listen EPERM ::1:3000`, наступний повтор пройшов 379/379 |
| `go test -race ./...` | запускався, **не пройшов повністю**: єдиний `TestShutdownDrainsAnInFlightResponse` не зміг `bind 127.0.0.1:0` через sandbox; усі інші показані пакети пройшли |
| `go test -race ./platform/architecture -run TestSafeRequestLogsDoNotIncludeURI -count=1` | пройшов фінально; R30 доведено |
| 4 незаблоковані `platform/httpserver` тести з race | пройшли; це не заміна повного Go-гейта |
| `sh -n infrastructure/tests/smoke.sh`; scoped `git diff --check` | пройшли |
| R30 mutation: nginx `$request_uri` | навмисно впав із очікуваною діагностикою; mutation знято |
| R30 mutation: непрямий `requestPath`/`urlPath` | навмисно впав із очікуваною діагностикою; mutation знято |

ФАКТ: **Docker відсутній, ці не запускались:** `make integration`, `make check-generated`, `make build` / `docker compose build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` і агрегатний `make check`. Через це B23 і B36 реалізовані, але **не оголошуються доведеними**; reviewer має перевірити числові очікування, записані вище.
ДАЛІ: оглянути фінальний scoped diff/status, переконатися, що nginx.conf і server.go чисті, а чужі паралельні зміни не входять у T14; потім зробити один малий коміт T14.
ФАКТ: фінальний scope чистий: змінені лише `backend/platform/architecture/boundaries_test.go` і `infrastructure/tests/smoke.sh`; нові — наданий `work/brief-t14.md` і цей `work/report-t14.md`. Тимчасові `nginx.conf`/`server.go` mutations відсутні; чужих tracked-змін більше немає. Попередній `work/brief-t13.md` tracked, тому brief T14 також належить логічному артефакту задачі.
РІШЕННЯ: комітити рівно ці чотири файли одним малим комітом; жодні generated/docs plan/report/task/status файли не включати.
ДАЛІ: staged diff check і коміт `test: cover remaining playback acceptance rows`.
ДАЛІ: git не чіпати записом; прочитати дописану поправку брифа, поточний `safe_request_logs_test.go`, G37 в `integration.sh` і перенести лише nginx-половину R30 у host-side static check до першого Docker-виклику.
ФАКТ: оновлений `work/brief-t14.md` прочитано повністю. Поправка визнає попередню рекомендацію помилкою брифа: Go-гейт бачить лише змонтований `backend/`, тому nginx-половина R30 мусить жити на хості в `integration.sh` до будь-якого Docker-виклику; Go-половина лишається в окремому `safe_request_logs_test.go`.
РІШЕННЯ: інструкція `implement` про коміт поступається прямій поправці рецензента: git-записів не буде. Після реалізації застосую read-only code-review, але commit не робитиму.
ДАЛІ: дочитати актуальний repository briefing, G37 і обидва поточні R30/smoke файли; зафіксувати точну дельту без пошуку в B23/B36.
ФАКТ: актуальний `work/CTO-RESEARCH.md` прочитано повністю; він підтверджує backend-only `GO_RUN` і host-owned Docker/integration seam. `integration.sh` обчислює `task_root`, потім одразу викликає `docker context inspect`; G37 нижче вже пояснює, чому root-файли перевіряються host-side.
ФАКТ: рецензент уже виніс R30 у `backend/platform/architecture/safe_request_logs_test.go`; `boundaries_test.go` чіпати не треба. Поточний файл усе ще містить nginx-підтест та його `os`/`regexp`/directive helpers; B23/B36 у `smoke.sh` прочитані лише для підтвердження scope і залишаються без змін.
РІШЕННЯ: host-перевірка буде одним POSIX `awk` до `docker context inspect`: state machine захоплює тільки `log_format safe` до першої завершальної `;`, токенізує цілі `$[[:alnum:]_]+` і fail-closed дозволяє лише `$request_id`, `$request_method`, `$status`, `$request_time`.
ДАЛІ: прибрати nginx-підтест і непотрібні helpers/imports із `safe_request_logs_test.go`; потім додати host awk між `task_root` і Docker.
ФАКТ: nginx-підтест, `findNginxDirective`, regex і root-file imports повністю прибрані з `safe_request_logs_test.go`; у Go лишився тільки AST allowlist виклику логера в сусідньому `platform/httpserver/server.go`.
ДАЛІ: додати host-side nginx allowlist у `integration.sh` точно після `task_root` і до `docker context inspect`.
ФАКТ: host-side R30 додано між `task_root` і першим Docker-викликом. `awk` входить у стан лише на `log_format safe`, читає всі рядки до `;`, виділяє повні токени змінних і дозволяє тільки `$request_id`, `$request_method`, `$status`, `$request_time`; будь-який інший токен друкує названу помилку й завершує скрипт до Docker.
РІШЕННЯ: розміщення несуче: статичний конфіг-інваріант не залежить від Docker і тепер може дати власний green/red навіть у цьому sandbox; G37 лишається нижче незмінним.
ДАЛІ: gofmt Go-файлу, `sh -n` integration, цільовий Go green; потім baseline `sh infrastructure/tests/integration.sh` і розрізнення його очікуваної Docker-відмови за текстом.
ФАКТ: `gofmt` і `sh -n infrastructure/tests/integration.sh` пройшли, але комбінована команда завершилась exit 1 на Go, бо її запущено з кореня без `go.mod`: `go: cannot find main module`. Go green не зараховано.
ДАЛІ: повторити точний Go target із каталогу `backend`, де рамка рівноцінна `GO_RUN` для `server.go`, бо і host, і контейнер бачать той самий backend subtree.
ФАКТ: `GOCACHE=/tmp/courses-t14-gocache go test -count=1 ./platform/architecture/` з `backend/` пройшов: `ok courses/platform/architecture 0.871s`. Для Go-половини це рівноцінно гейту: `server.go` лежить усередині backend subtree, який `GO_RUN` монтує без зміни шляху `/work`.
ДАЛІ: запустити baseline точно гейтовою командою `sh infrastructure/tests/integration.sh` з кореня й за текстом підтвердити, що R30 пройшов до Docker-відмови.
ФАКТ: baseline `sh infrastructure/tests/integration.sh` з кореня не впав на Docker у цьому відновленому середовищі, а пройшов повністю (exit 0), включно з host R30 та Docker integration-пакетами. Важливий discriminant: повідомлення R30 про unapproved variable відсутнє, скрипт пройшов host-перевірку й пішов далі; очікувана брифом Docker-відмова цього разу не відтворилась.
ДАЛІ: mutation (а) — тимчасово додати `$request_uri` в однорядковий `log_format safe`, запустити точну гейтову команду з кореня й вимагати власну помилку до Docker; потім повернути конфіг і виконати обов’язковий read-only `git diff --exit-code`.
ФАКТ: mutation (а) внесено тимчасово: `"uri":"$request_uri"` додано в однорядкову директиву `log_format safe`; це не продуктова дельта.
ДАЛІ: виконати `sh infrastructure/tests/integration.sh` з кореня й зафіксувати текст падіння host R30.
ФАКТ: mutation (а) прогнано точно командою `sh infrastructure/tests/integration.sh` з кореня; exit 1 стався до Docker з текстом `nginx log_format safe contains unapproved variable $request_uri`. Це власний red R30, не Docker-відмова.
ДАЛІ: повернути однорядковий `nginx.conf` точно як був і виконати `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: mutation (а) знято; `git diff --exit-code -- infrastructure/nginx/nginx.conf` завершився exit 0.
ДАЛІ: mutation (б) — тимчасово розбити `log_format safe` на два рядки з `$request_uri` на другому, прогнати точну host integration-команду, потім повернути конфіг і знову підтвердити нульову дельту.
ФАКТ: mutation (б) внесено: початок `log_format safe` лишився на першому рядку без `;`, повний формат із `$request_uri` перенесено на другий рядок.
ДАЛІ: виконати `sh infrastructure/tests/integration.sh` з кореня й перевірити, що багаторядкове захоплення дає власний red до Docker.
ФАКТ: mutation (б) прогнано точно командою `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker, текст `nginx log_format safe contains unapproved variable $request_uri`. Отже перенесення небезпечної змінної на другий рядок не робить перевірку сліпою.
ДАЛІ: повернути `log_format safe` до початкового одного рядка без URI і виконати `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: mutation (б) знято; `git diff --exit-code -- infrastructure/nginx/nginx.conf` завершився exit 0.
ДАЛІ: Go mutation — тимчасово замінити `"route", route` на `"path", r.URL.Path`, прогнати `GOCACHE=/tmp/courses-t14-gocache go test -count=1 ./platform/architecture/` з `backend/`, записати red і повернути нульову дельту server.go.
ФАКТ: Go mutation внесено точно як задано: logger тимчасово має `"path", r.URL.Path`; це не продуктова дельта.
ДАЛІ: виконати цільовий Go-тест із `backend/` і зафіксувати текст allowlist-падіння.
ФАКТ: Go mutation прогнано точно командою `GOCACHE=/tmp/courses-t14-gocache go test -count=1 ./platform/architecture/` з `backend/`; exit 1: `safe_request_logs_test.go:27: Go http request logger contains unapproved field "path"; review it for URI data before adding it to the safe allowlist` і `... is missing approved field "route"`.
ФАКТ: цей прямий Go-прогін рівноцінний `make test` для R30, бо і тут, і в `GO_RUN` доступний той самий `backend/platform/httpserver/server.go`; на відміну від старої nginx-половини жодного root-relative файла він не читає.
ДАЛІ: повернути logger до `"route", route`, виконати `git diff --exit-code -- backend/platform/httpserver/server.go`, потім повторити Go green.
ФАКТ: Go mutation знято; `git diff --exit-code -- backend/platform/httpserver/server.go` exit 0, а відновлений `GOCACHE=/tmp/courses-t14-gocache go test -count=1 ./platform/architecture/` зелений: `ok ... 0.510s`.
ДАЛІ: повторити фінальний baseline `sh infrastructure/tests/integration.sh` після всіх відновлень; потім перевірити syntax/format і scope, не запускаючи smoke та не торкаючись git-індексу.
ФАКТ: фінальний baseline `sh infrastructure/tests/integration.sh` з кореня пройшов повністю, exit 0; host nginx R30 виконався до Docker, після чого всі integration-пакети також зелені. Це сильніше за очікуваний у брифі «дійшов до Docker і впав», але рамка та порядок ті самі.
ДАЛІ: виконати фінальні read-only syntax/format/diff checks; підтвердити, що `boundaries_test.go`, `nginx.conf`, `server.go` дорівнюють HEAD, а змінені R30-файли — лише `safe_request_logs_test.go` та `integration.sh`. Smoke не запускати й не змінювати.
ФАКТ: фінальні read-only checks пройшли: `sh -n`, `gofmt -l` порожній, scoped `git diff --check` чистий; `boundaries_test.go`, `nginx.conf`, `server.go` дорівнюють HEAD. Tracked diff показує `integration.sh` і раніше змінений `smoke.sh`; новий `safe_request_logs_test.go` та append-only report не входять у `git diff --name-only`, бо ще untracked. У цьому колі `smoke.sh` не змінювався й не запускався.
ДАЛІ: виконати обов’язковий read-only code-review лише дельти цього кола за осями Standards/Spec; git index/commit не чіпати.
ФАКТ: `code-review/SKILL.md` прочитано повністю. Fixed point для незакоміченого reviewer-fix — поточний HEAD; review scope: host R30 у `integration.sh`, Go R30 у новому `safe_request_logs_test.go` і дописи цього кола в report. `smoke.sh` виключено, бо рецензент наказав лишити його як є.
ДАЛІ: паралельно запустити read-only Standards review проти AGENTS.md+smell baseline і Spec review проти виправлення рецензента/оновленого brief; файли не редагувати.
ФАКТ: обидві незалежні review-осі запущено паралельно й read-only; scope явно виключає `smoke.sh` та git-записи.
ДАЛІ: дочекатися Standards і Spec вердиктів, записати окремо й виправити лише підтверджені blocking-знахідки.
ФАКТ: staging/commit не відбулися: sandbox заборонив створити `.git/index.lock` — `fatal: Unable to create '/Users/strelia/projects/courses/.git/index.lock': Operation not permitted` (exit 128). Робочі файли збережені, але не staged і не committed; обійти read-only `.git` неможливо в межах дозволів цієї сесії.
РІШЕННЯ: не видавати коміт за створений і не обходити sandbox через інший інтерфейс. Реалізація готова для reviewer Docker-smoke та зовнішнього коміту рівно чотирьох названих файлів.
ДАЛІ: передати результат: R30 доведено локально; B23/B36 чекають reviewer smoke; повний Go/Compose gate не зелений і причини названі вище.

---

ФАКТ (рецензент, 2026-09-12): R30 винесено з `backend/platform/architecture/boundaries_test.go` у
новий `backend/platform/architecture/safe_request_logs_test.go` того ж пакета. Причина: до T14
`boundaries_test.go` містив рівно один тест — `TestDomainAndApplicationDependenciesPointInward`,
правило про напрям залежностей шарів. Формат логу — інший рід інваріанта; пакет правильний, файл ні.

ФАКТ: перенос дослівний, за діапазоном рядків. Єдина зміна — прибрано імпорт `io/fs`, який
компілятор назвав невикористаним у новому файлі. Перевірено: `git diff HEAD --
backend/platform/architecture/boundaries_test.go` порожній, тобто файл повернуто до стану до T14
байт у байт, і T14 у ньому лише додавав. `gofmt -l` чистий. На хості з `-race`:
`TestDomainAndApplicationDependenciesPointInward` PASS, `TestSafeRequestLogsDoNotIncludeURI` PASS
(обидва підтести — `nginx_safe_log_format`, `Go_request_logger`), `ok courses/platform/architecture`.

ФАКТ: тексти мутаційних падінь вище (`boundaries_test.go:72: log_format safe must not include
URI-bearing nginx variable $request_uri` і три рядки Go-мутації) знято ДО переносу й посилаються на
старе розташування. Після переносу той самий провал показуватиме шлях
`safe_request_logs_test.go`, а не `boundaries_test.go:72`.

РІШЕННЯ (рецензент): мутації на новому розташуванні свідомо не переміряно. Логіка перевірки
ідентична — перенос дослівний і перевірний без мутацій (див. вище). Повторна мутація натомість
знову зробила б `infrastructure/nginx/nginx.conf` і `backend/platform/httpserver/server.go`
тимчасово зламаними, а паралельна сесія щойно отримала підтвердження, що ці файли безпечно читати.
Тобто перевірка дала б мало, а відновила б гонку, яку ми вже закрили. Якщо рецензент T14 вважає
інакше — переміряти двома прогонами, попередивши паралельну сесію.

ФАКТ: розміщення R30 у `boundaries_test.go` першою помітила паралельна рецензент-сесія, не цей
прохід рев'ю.

---

ФАКТ (рецензент, гейт 2026-09-12): `make check` → **GATE_EXIT=2** на цілі `test`, до `integration`.
`TestSafeRequestLogsDoNotIncludeURI/nginx_safe_log_format` FAIL:
`safe_request_logs_test.go:25: open ../../../infrastructure/nginx/nginx.conf: no such file or directory`.
Go-підтест не впав. Стек не зачеплено: `test` у ланцюжку `check` четверта ціль, `up` — дев'ята;
Redis не зупинявся, B23 і B36 не міряні.

ФАКТ: причина — середовище, у якому гейт жене Go. `Makefile:3`:
`GO_RUN := docker run --rm -v "$(CURDIR)/backend:/work" -w /work …` — у контейнер змонтовано лише
`backend/`. `infrastructure/nginx/nginx.conf` лежить поза ним. Codex і рецензент обидва міряли R30
прямим `go test` на повному checkout, де шлях існує, — не в умовах гейта.

ФАКТ: отже в гейті nginx-половина R30 — перевірка, що **падає завжди**, незалежно від конфігу.
Мутаційний доказ «червоніє на `$request_uri`» знято в прямому середовищі; для гейта він нічого не
означає, бо там вона червона й без мутації.

ФАКТ: репо це обмеження вже знає. `infrastructure/tests/integration.sh`, коментар G37: «The Go tests
cannot assert this - their container mounts only backend/, so the repository root is invisible from
inside them», — і тому свої перевірки кореневих файлів робить на хості, до `docker run`.

РІШЕННЯ (рецензент): це дефект **брифа**, не лише реалізації. `work/brief-t14.md` §2.3 сам
запропонував `backend/platform/architecture` як дім перевірки, що читає `nginx.conf`, не звіривши
пропозицію з монтуванням `GO_RUN`. Codex пішов за підказкою. Поправку дописано в бриф.

РІШЕННЯ: розділити R30. Go-половина (читає `server.go` усередині `backend/`) лишається в
`safe_request_logs_test.go`. nginx-половина переїжджає в хостову частину `integration.sh` поруч із
G37. Відкинуто: змонтувати корінь репо в `GO_RUN` — ламає свідоме backend-only монтування; `t.Skip`,
коли файлу немає, — гейт зелений, а nginx-половина не виконується ніколи, тобто маскування.

ДАЛІ: повернути Codex через resume з виправленням. Обидві мутації переміряти **в тих умовах, у яких
їх жене гейт**: nginx — хостовою перевіркою з кореня репо; Go — у контейнері `GO_RUN`.

ФАКТ: фінальний Standards review цього кола — 0 порушень AGENTS.md, 0 actionable smells. Spec review підтвердив R30 по суті, але знайшов: (1) blocking-порушення хронології journal — перший apply_patch цього кола прив’язався до неостаннього однакового `ДАЛІ:` і вставив нові записи перед уже дописаним reviewer-блоком; (2) non-blocking false positive — awk токенізує також текст після завершальної `;` на тому самому рядку.
РІШЕННЯ: попередні записи не переміщати й не видаляти, бо це було б rewrite; ця примітка на фактичному EOF відновлює явну хронологію та визнає незворотну помилку append-позиції. Надалі додавати лише після цього EOF. Awk обрізати перед першою завершальною `;` до токенізації.
ДАЛІ: виправити межу awk «до `;`», повторити baseline і обидві nginx-мутації в host integration-рамці, не змінюючи smoke чи git state.
ФАКТ: awk тепер спершу знаходить завершальну `;`, обрізає поточний рядок перед нею і лише потім токенізує змінні; текст/коментар після директиви більше не може дати false positive.
ДАЛІ: запустити `sh -n` і baseline `sh infrastructure/tests/integration.sh` з кореня.
ФАКТ: після awk-fix `sh -n` пройшов; baseline `sh infrastructure/tests/integration.sh` завершився exit 1 лише після host R30 на Docker API: `permission denied while trying to connect to the docker API at unix:///Users/strelia/.docker/run/docker.sock`. Власного R30 error немає, отже host static check зелений і дійшов до Docker-межі.
ДАЛІ: повторити однорядкову `$request_uri` mutation фінальним awk, зафіксувати власний red до Docker й повернути nginx.conf з zero diff.
ФАКТ: фінальну однорядкову mutation внесено й прогнано точно `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker: `nginx log_format safe contains unapproved variable $request_uri`.
ДАЛІ: прибрати mutation та виконати `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: фінальну однорядкову mutation знято; nginx.conf zero diff, `git diff --exit-code` exit 0.
ДАЛІ: повторити фінальну дворядкову mutation з `$request_uri` на другому рядку, прогнати host integration, повернути й перевірити zero diff.
ФАКТ: фінальну дворядкову mutation прогнано точно `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker: `nginx log_format safe contains unapproved variable $request_uri`. Multi-line захоплення лишається доказаним після awk-fix.
ДАЛІ: повернути однорядковий safe format без URI та виконати `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: фінальну дворядкову mutation знято; nginx.conf zero diff, `git diff --exit-code` exit 0.
ДАЛІ: віддати awk-fix і EOF-корекцію тим самим reviewers на короткий read-only re-review; потім записати фінальний стан без git commit.
ФАКТ: Standards і Spec re-review запущені read-only; git і smoke поза scope.
ДАЛІ: дочекатися обох фінальних вердиктів.

ФАКТ: фізичний EOF перевірено через `nl -ba work/report-t14.md | tail`: попередні quote-aware/comment-aware записи справді були вставлені раніше (рядки 146–180), бо короткі контекстні anchors повторювалися. Вони описують реальні кроки, але їхня позиція нехронологічна. Цей запис зроблено після унікального трирядкового EOF-anchor; історію не переписано.
ФАКТ: для відновлення читабельної хронології: після рядка 335 реалізація вже (а) стала quote-aware; (б) почала ігнорувати unquoted `# ...` до кінця фізичного рядка; (в) червоніла на plain, multiline, quoted-semicolon і comment-semicolon `$request_uri` мутаціях; (г) після кожної nginx.conf повертався в zero diff; (ґ) останній baseline integration проходив exit 0. Точні команди й тексти є в фізично ранніших append-записах.
ФАКТ: останній Spec review знайшов новий blocking edge: nginx дозволяє braced `${request_uri}`, а regex бачить лише `$request_uri`; Standards findings — 0. Журнал також був позначений blocking саме через описану позиційну помилку.
РІШЕННЯ: додати до tokenizer обидві форми, canonicalize `${name}` у `$name` і перевіряти тим самим allowlist; переміряти `${request_uri}` host-командою. Надалі append робити лише за повним унікальним EOF-блоком.
ДАЛІ: виправити braced-token parsing, baseline, `${request_uri}` mutation, restore zero diff і остаточний review; git записом не чіпати.
ФАКТ: tokenizer тепер розпізнає повні `$name` і `${name}` токени; braced-форму canonicalize-ить до `$name` перед тим самим чотириелементним allowlist, а діагностика друкує оригінальний токен.
ДАЛІ: `sh -n` і baseline host integration на чистому nginx.conf.
ФАКТ: braced-aware scanner пройшов `sh -n`; перший baseline пройшов host R30 і впав лише на Docker API permission (`permission denied ... docker.sock`), без R30 error. Одна спроба допису не змінила файл через typo шляху `workE/work/report-t14.md`; помилку не приховано.
ФАКТ: `${request_uri}` mutation прогнано точно `sh infrastructure/tests/integration.sh` з кореня; exit 1 до Docker з текстом `nginx log_format safe contains unapproved variable ${request_uri}`. Після повернення `git diff --exit-code -- infrastructure/nginx/nginx.conf` дав exit 0.
ФАКТ: фінальний baseline після відновлення пройшов повністю, `sh infrastructure/tests/integration.sh` exit 0, включно з host R30 та Docker integration-пакетами.

## Standards — фінальний re-review цього кола

- Findings: 0. Задокументованих порушень і actionable smells не лишилось.

## Spec — фінальний re-review цього кола

- Кодових findings: 0. Підтверджено host-before-Docker розміщення; directive-local, multiline, quote/comment/escape-aware межі; `$name`/`${name}` exact-token allowlist; Go-only backend test; mutation red і zero-diff restore.
- Історичний blocking journal: частина записів цього кола фізично вставилася перед reviewer-блоком через неунікальні patch anchors. Переміщення/видалення було б rewrite, тому помилку явно зафіксовано, а authoritative chronology двічі підсумовано на перевіреному EOF. Останній reviewer визнав таку immutable-корекцію прийнятною, але наступний re-review читав файл до цього фінального append і попросив саме наведені вище braced evidence.

РІШЕННЯ: R30 виправлено за поправкою брифа. `safe_request_logs_test.go` перевіряє лише Go logger у backend-visible рамці; `integration.sh` перевіряє nginx host-side до першого Docker-виклику. B23/B36 і `smoke.sh` у цьому колі не змінювались. Git index/commit не чіпалися; коміт лишається рецензенту.
ДАЛІ: рецензент запускає свій Docker-гейт/коміт; у цій сесії додаткових дій немає.

---

ФАКТ (рецензент, повторний гейт 2026-09-12, після кола з виправленням R30): `make check` →
**GATE_EXIT=2** на цілі `smoke`. До неї пройшло все: `format-check`, `lint`, `typecheck`, `test`,
`integration`, `check-generated`, `build`, `up`, `seed-identities`, `seed`. `e2e` не запускався.

ФАКТ: **R30 зелений в умовах гейта, обидві половини.** Go-половина пройшла `make test` у контейнері
`GO_RUN` (лише `backend/`). nginx-половина пройшла на хості в `sh infrastructure/tests/integration.sh`,
до першого `docker context inspect` — жодного повідомлення `log_format safe` у лозі гейта.

ФАКТ: **B36 — дві частини з трьох.** Скрипт перевіряє послідовно: статус **503 ✓**, заголовок
**`Retry-After` ✓**, тіло ✗ — `media with Redis unavailable: expected an empty body, got 190 bytes`.

ФАКТ: 190 байтів — це стандартна сторінка помилки 503 у nginx. `@unavailable` робить `return 503;`
без тексту, і nginx віддає вбудований HTML. Перевірено арифметикою з CRLF і `server_tokens off`:
`<html>` 8 + `<head><title>503 Service Temporarily Unavailable</title></head>` 65 + `<body>` 8 +
`<center><h1>…</h1></center>` 63 + `<hr><center>nginx</center>` 28 + `</body>` 9 + `</html>` 9 = **190**.

ФАКТ: **безумовне відновлення спрацювало на реальному падінні.** Після провалу B36 trap вивів
`PASS Redis recovery restored media authorization policy (503 -> 403)`; Redis `running healthy`;
`/tmp/courses-smoke-media-cookie` прибрано. B23 не запускався — smoke зупинився на B36.

РІШЕННЯ (рецензент): провал B36 — дефект **брифа**, не реалізації. `work/brief-t14.md` §2.2 казав
«тіло порожнє — жодного байта медіа», а вимагав довжину тіла 0, тобто злив «немає медіа» з «порожнє
тіло». Codex виконав буквально.

РІШЕННЯ (рецензент): і глибше — застереження «жодного байта» в smoke не може бути доведене змістовно.
Запит іде з **невідомим** `playbackId`, тож медіа-байтів немає за будь-якого стану Redis: буквальна
перевірка падає на HTML nginx, змістовна виконувалась би порожньо. Зробити тіло `@unavailable`
порожнім лише заради тесту — означає підлаштувати продукт під перевірку; відкинуто. Рішення, де живе
доказ «жодного байта медіа», — того самого роду, що шість рядків без фікстури, і воно за користувачем.

ФАКТ (кандидат, не блокує T14): перевірка R30 доводить, що **визначення** `log_format safe` не несе URI,
але не те, що nginx **пише лог у ньому**. `log_format full '… $request_uri'` плюс
`access_log /dev/stdout full;` (`nginx.conf:16`) лишили б її зеленою. Рядок плану R30 сформульований
так само («прочитати `log_format safe`»), тож T14 рядок виконує; кандидат — стверджувати ще, що кожен
`access_log` посилається на `safe` або `off`. Звуження до директиви було в брифі.

ФАКТ: попередні записи Codex цього кола лягали над рецензентськими блоками, не в кінець; Codex сам це
помітив і дописав у справжній кінець примітку про хронологію, нічого не переносячи й не видаляючи.

ФАКТ (рецензент, ДІАГНОСТИКА — не доказ; гейт до B23 не дійшов): фрагмент B23 зі `smoke.sh` прогнано
вручну на піднятому після гейта стеку, `docker compose exec -T frontend node`, 720 запитів до того самого
`/media/p/…` шляху. Результат: **403 × 245, 429 × 475**. 245 допущених = `burst=240` плюс поповнення за
час сплеску при `rate=120r/s` — рівно арифметика з §2.1 брифа. Доказом B23 стане лише прогін у гейті.

РІШЕННЯ (користувач, 2026-09-12): шість рядків без фікстури — B20, B21, B21a, B22, B28, R23 — перенести
на e2e датованою поправкою плану з причиною; окремий таск T16.

РІШЕННЯ (користувач, 2026-09-12): B36 розділити. Smoke доводить 503 і `Retry-After`; «жодного байта
медіа» переходить на e2e зі справжнім сегментом, разом із T16. Порожнє тіло `@unavailable` у nginx
відкинуто як підлаштування продукту під тест.

ДАЛІ (рецензент): третє коло Codex T14 — зняти в `smoke.sh` лише перевірку довжини тіла B36; потім
повний гейт у новий лог.
ДАЛІ: git не чіпати; прочитати поправку §2.2, змінити тільки B36 у `smoke.sh` — прибрати буквальну перевірку порожнього HTTP-тіла, лишивши 503, `Retry-After`, trap і recovery без змін — та перевірити лише syntax і scoped diff.
ФАКТ: правила `implement` прочитано; його вимогу коміту в цьому колі перекриває пряма заборона користувача — git-записів не буде. `work/CTO-RESEARCH.md` підтверджує межу середовища: Docker-гейт виконує рецензент. Актуальний B36 досі окремо читає тіло й вимагає `byteLength===0`; status, `Retry-After`, trap і recovery розташовані окремо й можуть лишитися без змін.
ДАЛІ: дочитати актуальний `work/brief-t14.md` без обрізання, особливо датовану поправку §2.2, і зафіксувати точний контракт третього кола.
ФАКТ: актуальний `work/brief-t14.md` прочитано повністю (205 рядків). Датована поправка §2.2 фіксує рішення користувача: smoke лишає доказ 503 і `Retry-After`, а «жодного байта медіа» переходить у T16 e2e зі справжнім сегментом; `nginx.conf` не змінюється. Залишкову стару фразу брифа про числову довжину тіла перекриває це рішення та буквальна вказівка третього кола прибрати `body.byteLength!==0`.
РІШЕННЯ: видалити тільки читання response body і assertion його нульової довжини; PASS називатиме лише реально перевірені 503 та `Retry-After`. Перед перевіркою додати захисний коментар про невідомий `playbackId` і перенесення медіа-доказу в T16 рішенням користувача 2026-09-12.
ДАЛІ: внести єдину правку у B36, не торкаючись B23, trap/recovery, `nginx.conf`, `integration.sh` і Go.
ФАКТ: у B36 `smoke.sh` видалено тільки `response.arrayBuffer()`, змінну `body` та assertion `body.byteLength!==0`; перевірки status 503 і присутності `Retry-After` не змінено. PASS тепер каже лише `PASS unavailable Redis returns 503 with Retry-After`.
ФАКТ: дворядковий коментар над запитом закріплює причину: `playbackId` невідомий, тому медіа немає незалежно від Redis; доказ `no media bytes` перенесено в e2e T16 рішенням користувача 2026-09-12.
ДАЛІ: виконати рівно замовлені перевірки — `sh -n infrastructure/tests/smoke.sh` і `git diff -- infrastructure/tests/smoke.sh`; Docker smoke не запускати.
ФАКТ: точна команда `sh -n infrastructure/tests/smoke.sh` завершилась exit 0.
ФАКТ: точна команда `git diff -- infrastructure/tests/smoke.sh` завершилась exit 0, але від HEAD показує весь раніше незакомічений T14-диф цього файла: збереження cookie, B36 із trap/recovery та B23. Тому буквальна умова «diff показує лише цей блок» у поточній робочій рамці не виконується; це не нова дельта третього кола, і заборона «B23/trap/recovery лишити як є» не дозволяє прибирати її задля косметики diff.
РІШЕННЯ: не stage/commit і не переписувати попередні зміни. Дельта саме цього кола обмежена B36: два коментарі додано, `body`/перевірку довжини видалено, PASS звужено; інших рядків не змінювалось.
ФАКТ: Docker smoke не запускався відповідно до прямої вказівки та межі середовища. B36 цим колом не оголошується доведеним; його 503/`Retry-After` має повторно підтвердити рецензентський Docker-гейт.
ДАЛІ: виконати обов’язковий read-only code review цієї вузької дельти; жодних додаткових тестів або git-записів.
ФАКТ: `code-review` прочитано. Його звичайний fixed point (`git diff <ref>...HEAD`) тут непридатний: користувач заборонив git-записи, а HEAD-відмінність уже містить попередні незакомічені B23/B36 зміни. Spec source є — актуальний `work/brief-t14.md` плюс буквальна інструкція третього кола.
РІШЕННЯ: дати двом read-only осям review точну ізольовану дельту третього кола (два коментарі, видалення `body`/assertion, звуження PASS), не весь історичний diff від HEAD. Це перевіряє саме замовлену зміну без stage/commit і без торкання файлів.
ДАЛІ: паралельно перевірити Standards та Spec; прийняти лише findings у межах B36.
ФАКТ: фінальна інспекція `smoke.sh:169-174` підтвердила дворядковий коментар, незмінені assertions status 503 і `Retry-After`, відсутність читання/перевірки body та звужений PASS. Сусідні trap/recovery і B23 лишилися поза дельтою третього кола.

## Standards — review третього кола

- Findings: 0. Дельта локальна й цілісна, не додає залежностей або абстракцій, не послаблює дві ознаки, що залишені за smoke, і не порушує задокументованих правил репо.

## Spec — review третього кола

- Findings: 0. Виконано всі буквальні вимоги рішення користувача: body-read/assertion прибрані, 503 і `Retry-After` збережені, PASS не заявляє порожнього тіла, дворядковий коментар фіксує невідомий `playbackId` та передачу доказу в T16. Scope creep у цій дельті відсутній.

РІШЕННЯ: третє коло T14 завершити на цій вузькій зміні. Git не чіпали; Docker smoke не запускався; B36 не оголошено доведеним у цій сесії.
ДАЛІ: рецензент запускає Docker-гейт, підтверджує B36 за status 503 і `Retry-After` та робить коміт; доказ «жодного байта медіа» виконується окремо в T16 e2e зі справжнім сегментом.
ДАЛІ: git не чіпати; перевірити чинний host-side R30 scanner, додати до нього лише allowlist для кожної `access_log` директиви (`safe` або `off`) до першого Docker-виклику та переміряти baseline і дві задані мутації точними gate-командами.
ФАКТ: `work/CTO-RESEARCH.md`, `receiving-code-review` і `implement` прочитано повністю. Пряма заборона користувача перекриває загальну рекомендацію `implement` про коміт; жодного git-запису не буде.
ФАКТ: вимога рецензента технічно підтверджена кодом. Чинний awk на `integration.sh:6-72` знаходить і валідовує лише визначення `log_format safe`; після цього `access_log ... combined` або `access_log шлях;` справді не впливають на результат R30.
РІШЕННЯ: розширити той самий awk і його `directive_terminator` другим станом збору всіх line-start `access_log` директив до їхнього нецитованого/неекранованого `;`. Окремий quote/escape-aware tokenizer визнає лише точний двотокенний `access_log off` або директиву, де третій токен — точний формат `safe`; відсутній третій токен і `combined` червоні.
ДАЛІ: звірити дві фактичні `access_log` директиви у `nginx.conf`, потім внести тільки цю host-side перевірку перед Docker.
ФАКТ: `nginx.conf` підтверджує рівно дві чинні директиви: `access_log /dev/stdout safe;` на рядку 16 і inline `access_log off;` усередині `location = /healthz { ... }` на рядку 52. Історичний `git diff -- infrastructure/tests/integration.sh` містить лише попередню host-side R30 перевірку відносно HEAD.
РІШЕННЯ: початковий задум шукати лише line-start `access_log` був би хибним — він пропустив би фактичну inline-директиву `/healthz`. Сканер має лексично проходити весь конфіг, відділяти statements за нецитованими/неекранованими `;` і блоки за `{}`; тоді він бачить обидві директиви, коментарі та лапки не створюють фальшивих збігів.
ДАЛІ: замінити вузький збирач однієї директиви на єдиний POSIX awk scanner statements, зберігши точний allowlist змінних `log_format safe` й додавши тільки правило `access_log safe|off`.
ФАКТ: перша `apply_patch` спроба рефакторингу scanner не застосувалась (`Failed to find expected lines`): контекст patch не збігся з фактичним завершенням awk-блока. `integration.sh` цією спробою не змінено.
РІШЕННЯ: не вгадувати контекст; перечитати точний блок і замінити його за стабільними межами `awk '` … `' "$task_root/..."`.
ДАЛІ: отримати точний поточний текст awk-блока та повторити одну локальну правку.
ФАКТ: точний awk-блок перечитано. Його поточний END-контракт — `nginx log_format safe not found or missing semicolon`, а діагностики йдуть через загальне `>&2`; попередній patch очікував іншу редакцію повідомлень.
ДАЛІ: застосувати scanner-заміну за фактичним текстом, зберігаючи host-before-Docker розміщення і stderr-рамку.
ФАКТ: host-side R30 awk тепер лексить увесь nginx.conf на statements, розділяючи лише нецитовані й неекрановані `;`, ігноруючи лише нецитовані `#`-коментарі та скидаючи block headers на `{`/`}`. Тому inline `/healthz { access_log off; ... }` також перевіряється.
ФАКТ: `log_format safe` зберіг попередній exact-token allowlist для `$name` і `${name}`. Для кожної знайденої `access_log` валідні лише два контракти: рівно `access_log off;` або щонайменше три токени з точним третім `safe`; отже `combined` і пропущений формат червоні.
ДАЛІ: перевірити POSIX sh syntax, потім базовий `sh infrastructure/tests/integration.sh` і відрізнити проходження host R30 від очікуваної Docker-відмови за текстом.
ФАКТ: `sh -n infrastructure/tests/integration.sh` завершилась exit 0; shell-синтаксис після scanner-зміни валідний.
ДАЛІ: прогнати baseline точно `sh infrastructure/tests/integration.sh` з кореня репо й зафіксувати перше повідомлення після host-перевірки.
ФАКТ: перший baseline, точна команда `sh infrastructure/tests/integration.sh` з кореня, дав exit 2 до Docker з текстом `awk: syntax error at source line 2 in function reset_statement` і контекстом `reset_statement(index)`. Це падіння нашого scanner: системний awk трактує `index` як builtin, тому baseline ще не зелений.
РІШЕННЯ: перейменувати лише локальний параметр `index` на `key`; логіка не змінюється.
ДАЛІ: застосувати portability-fix і повторити `sh -n`, потім baseline тією самою gate-командою.
ФАКТ: awk-local параметр `index` перейменовано на `key`; це єдина зміна після першого baseline-падіння.
ДАЛІ: повторно перевірити shell/awk parsing через `sh -n` і реальний запуск host-префікса.
ФАКТ: повторний `sh -n infrastructure/tests/integration.sh` після portability-fix — exit 0.
ДАЛІ: повторити baseline точно `sh infrastructure/tests/integration.sh` з кореня.
ФАКТ: повторний baseline `sh infrastructure/tests/integration.sh` не вивів R30-помилки, пройшов host-префікс і почав Docker Go integration (`courses/cmd/api`, `media-worker`, `migrate`, `upload`), але перший 10-секундний yield не повернув фінального exit-коду. Спроба read-only перевірити процес через `ps` дала `operation not permitted` — межа пісочниці, не результат гейта.
РІШЕННЯ: повторити точну baseline-команду з довшим yield і зберегти session id, якщо вона знову триватиме; не трактувати частковий лог як pass.
ДАЛІ: отримати завершений baseline або явний Docker-failure text.
ФАКТ: завершений baseline прогнано точно `sh infrastructure/tests/integration.sh` з кореня репо; exit 0. Host R30 не вивів помилок, після нього Docker integration теж завершився зеленим (усі перелічені Go packages `ok`/`[no test files]`). Це baseline (в), виміряний у повній gate-команді.
ДАЛІ: мутація (а) — тимчасово замінити `access_log /dev/stdout safe;` на `access_log /dev/stdout combined;`, прогнати ту саму команду й вимагати падіння нашої перевірки до Docker.
ФАКТ: мутацію (а) застосовано тимчасово: активний stdout `access_log` називає `combined` замість `safe`.
ДАЛІ: виконати точно `sh infrastructure/tests/integration.sh` з кореня й зафіксувати діагностику до Docker.
ФАКТ: мутація (а), точна команда `sh infrastructure/tests/integration.sh` з кореня, дала exit 1 до будь-якого Docker output з текстом `nginx access_log must use safe format or be off:    access_log /dev/stdout combined`. Отже небезпечний явний формат червоніє саме на новій host-перевірці.
ДАЛІ: повернути `safe` і виконати точно `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: мутацію (а) знято, stdout format повернуто на `safe`.
ДАЛІ: підтвердити нульову дельту nginx.conf командою `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: після мутації (а) точна команда `git diff --exit-code -- infrastructure/nginx/nginx.conf` дала exit 0; nginx.conf повернуто байт у байт відносно HEAD.
ДАЛІ: мутація (б) — тимчасово додати `access_log /tmp/extra.log;` без формату, прогнати точну gate-команду й вимагати падіння до Docker.
ФАКТ: мутацію (б) застосовано тимчасово: додано окрему `access_log /tmp/extra.log;` без третього токена формату.
ДАЛІ: виконати точно `sh infrastructure/tests/integration.sh` з кореня й зафіксувати host-side падіння.
ФАКТ: мутація (б), точна команда `sh infrastructure/tests/integration.sh` з кореня, дала exit 1 до будь-якого Docker output з текстом `nginx access_log must use safe format or be off:    access_log /tmp/extra.log`. Отже відсутній format не трактується нейтрально: перевірка червоніє до fallback `combined`.
ДАЛІ: видалити тимчасову директиву й підтвердити `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: мутацію (б) знято; тимчасову `/tmp/extra.log` директиву видалено.
ДАЛІ: перевірити нульову дельту nginx.conf точною git diff-командою.
ФАКТ: після мутації (б) точна команда `git diff --exit-code -- infrastructure/nginx/nginx.conf` дала exit 0; продуктова конфігурація знову має нульову дельту.
ДАЛІ: повторити baseline на фінальному конфігу точно `sh infrastructure/tests/integration.sh`; це остаточний зелений доказ (в) після обох restore.
ФАКТ: фінальний baseline (в), точна команда `sh infrastructure/tests/integration.sh` з кореня, пройшов host R30 без повідомлення `nginx access_log...` і дійшов до Docker/Testcontainers. Загальний exit 1 виник пізніше в Docker з текстом `wait for reaper b89ec6fa: unexpected container status "created"` у `TestAdministrationRepositoryAgainstPostgreSQL`; це «зелене» R30 за заданим критерієм, але не зелений повний integration-гейт. Раніший baseline тією самою командою мав exit 0.
РІШЕННЯ: розрізняти результати за текстом і місцем: обидві мутації впали негайно на `nginx access_log must use safe format or be off` до Docker; фінальний baseline не мав цієї діагностики й виконував Docker Go packages.
ДАЛІ: виконати обов’язковий read-only `code-review` вузької R30 дельти, потім оглянути фінальний scoped diff; git-записів і додаткових продуктових правок не робити.
ФАКТ: `code-review` прочитано повністю; scoped `git diff -- infrastructure/tests/integration.sh` підтверджує, що робочий файл змінює лише host-префікс до `docker context inspect`, але diff від HEAD містить також попередню прийняту R30 реалізацію. Окремого commit fixed point для четвертого кола немає через заборону git-записів.
РІШЕННЯ: дві review-осі отримають фактичний scoped diff і точне розмежування: нова дельта — перехід від collector `log_format safe` до спільного statement scanner плюс правило `access_log`; попереднє host-before-Docker розміщення не є новою зміною.
ДАЛІ: паралельний read-only Standards/Spec review, без файлів, тестів і git-записів.
ФАКТ: завершальний двоосьовий code review повернув Standards findings: 0 і Spec findings: 0. Standards підтвердив POSIX sh+awk, host-before ambiguous Docker, відсутність нових залежностей і виправдану scanner-складність; Spec підтвердив охоплення inline/multiline `access_log`, quotes/escapes/comments/`${...}`, точні дозволені форми `safe`/`off` і червоний implicit `combined`.
РІШЕННЯ: четверте коло завершене без додаткових правок. Змінено тільки hostcodes R30 у `infrastructure/tests/in Nancytegration.sh`; `smoke.sh`, Go і фінальний `nginx.conf` не змінено. Git index/commit не чіпалися.
ДАЛІ: рецензент запускає Docker-гейт і робить коміт. Докази цього кола: baseline host-check зелений (один повний exit 0; фінальний прогін дійшов до Docker й упав лише на Testcontainers reaper), обидві задані мутації червоні до Docker з точними текстами вище, після кожної nginx.conf мав `git diff --exit-code` exit 0.
ФАКТ: виправлення двох описок у попередньому append без rewrite: `host-before ambiguous Docker` слід читати як `host-before-Docker`; `infrastructure/tests/in Nancytegration.sh` слід читати як `infrastructure/tests/integration.sh`. Фактичний файл зміни — тільки `infrastructure/tests/integration.sh`; код і результати перевірок описки не зачепили.
ДАЛІ: git не чіпати; прочитати поправку «п’яте коло», додати лише парний `403` assertion до B23 і вимогу хоча б одного активного `access_log ... safe` до R30, потім виконати замовлені syntax, mutation/restore та baseline перевірки без змін Go, nginx.conf, B36 або trap.
ФАКТ: перед успішним append дві оркестраційні спроби не дійшли до `apply_patch`: JavaScript wrapper упав спершу на випадковому `_AWAY`, потім на `Bot is not defined`. Жоден файл цими двома викликами не змінено; перший фактичний файловий запис цього кола — наведений вище `ДАЛІ:` у EOF.
ДАЛІ: прочитати актуальний репо-брифінг, скіли та поправку «п’яте коло» до будь-якої кодової правки.
ФАКТ: `work/CTO-RESEARCH.md` і весь актуальний `work/brief-t14.md` прочитано; поправка «п’яте коло» дослівно підтверджує дві діри: B23 потребує одночасної наявності 429 і 403, R30 — хоча б однієї `access_log`, що явно називає `safe`. `receiving-code-review` підтверджує порядок verify-before-change; `implement` вимагає review, але його загальну вимогу коміту перекриває пряма заборона користувача.
РІШЕННЯ: для B23 додати один `statuses.includes(403)` assertion та однореченнєвий коментар без числових квот; для R30 рахувати лише успішно розібрані директиви з третім токеном `safe` і в END червоніти, якщо лічильник нульовий, з поясненням implicit `combined`/URI.
ДАЛІ: перечитати точні чинні B23 і R30 блоки, щоб внести лише дві однорядкові вимоги й потрібний коментар.
ФАКТ: скіли перечитано без обрізання. Чинний B23 має лише `statuses.includes(429)` після 720 запитів. Чинний R30 `validate_access_log` повертає успіх для `safe`, але не фіксує існування такої директиви; END перевіряє тільки `log_format_safe_found`.
РІШЕННЯ: B23 отримає один сусідній `includes(403)` assertion і одне речення-коментар про пару фаз. R30 встановлюватиме `safe_access_log_found=1` лише на валідній директиві з третім токеном `safe`; END вимагатиме цей прапорець і пояснюватиме, що без explicit safe nginx defaults to combined and logs URI.
ДАЛІ: внести B23 зміну окремо, не торкаючись B36/trap.
ФАКТ: B23 доповнено рівно парною вимогою: після `includes(429)` тепер обов’язковий `includes(403)` з помилкою `Media rate limit did not admit a request to authorization`. Однореченнєвий коментар пояснює, що пара доводить відсікання лише надлишку до авторизації, а не відмову всім запитам. Конкретні кількості не зафіксовано.
ДАЛІ: додати R30 existence-прапорець і END assertion у той самий scanner, не змінюючи решту правил.
ФАКТ: R30 scanner тепер ставить `safe_access_log_found` лише коли розібрана `access_log` має явний третій токен `safe`; `access_log off;` лишається дозволеною, але не задовольняє existence-вимогу. END без active safe падає з текстом `nginx requires an access_log using safe; otherwise it defaults to combined and logs URI`.
ДАЛІ: виконати обидві замовлені syntax-перевірки; Docker smoke не запускати.
ФАКТ: точна команда `sh -n infrastructure/tests/smoke.sh` — exit 0; точна команда `sh -n infrastructure/tests/integration.sh` — exit 0.
ФАКТ: B23 runtime smoke не запускався; на цьому етапі новий `403` assertion має лише syntax-доказ і B23 не оголошується доведеним.
ДАЛІ: R30 мутація — тимчасово видалити `access_log /dev/stdout safe;`, прогнати точно `sh infrastructure/tests/integration.sh` з кореня й вимагати наш текст до Docker.
ФАКТ: R30 existence-мутацію застосовано тимчасово: єдину активну `access_log /dev/stdout safe;` видалено; inline `access_log off;` лишилася.
ДАЛІ: виконати точну gate-команду й зафіксувати exit/text до Docker.
ФАКТ: existence-мутація, точна команда `sh infrastructure/tests/integration.sh` з кореня, дала exit 1 до будь-якого Docker output з точним текстом `nginx requires an access_log using safe; otherwise it defaults to combined and logs URI`. Отже scanner не приймає вакуумну істину, коли лишається тільки `access_log off;`.
ДАЛІ: повернути `access_log /dev/stdout safe;` і виконати точно `git diff --exit-code -- infrastructure/nginx/nginx.conf`.
ФАКТ: тимчасову R30 мутацію знято; активний stdout `access_log ... safe` повернуто на початкове місце.
ДАЛІ: підтвердити нульову дельту nginx.conf заданою git diff-командою.
ФАКТ: після restore точна команда `git diff --exit-code -- infrastructure/nginx/nginx.conf` дала exit 0; продуктова конфігурація знову байт у байт відносно HEAD.
ДАЛІ: прогнати baseline точно `sh infrastructure/tests/integration.sh` з кореня; зелене R30 визначити за проходженням host scanner і переходом до Docker, незалежно від подальшого Docker outcome.
ФАКТ: baseline прогнано точно `sh infrastructure/tests/integration.sh` з кореня; exit 0. Host R30 пройшов без діагностики й повний Docker integration цього разу також завершився зеленим (`ok`/`[no test files]` для всіх виведених Go packages).
ДАЛІ: виконати обов’язковий read-only `code-review` двох вузьких дельт, потім фінально звірити scoped diff; B23 smoke не запускати й git-записів не робити.
ФАКТ: `code-review` прочитано повністю. Scoped diff від HEAD містить усі попередні незакомічені T14 правки, тому fixed point п’ятого кола не існує як commit; ізольована нова дельта — три рядки B23 (коментар+403 assert) і сім рядків R30 (safe flag+END error).
РІШЕННЯ: передати паралельним Standards/Spec reviewer-ам саме ізольовану дельту п’ятого кола, водночас дозволивши читати чинні повні блоки для контексту; жодних edit/test/git дій у subagents.
ДАЛІ: двовісний read-only review, потім фінальний підсумок у EOF.
## Standards — review п’ятого кола

- Findings: 0. B23 не фіксує timing-dependent counts і точно описує пару 429/403; R30 fail-closed existence assertion повторно використовує чинний POSIX awk scanner, не додає залежностей і пояснює небезпечний default `combined`. Задокументованих порушень або material smells немає.

## Spec — review п’ятого кола

- Findings: 0. B23 має сусідні `includes(429)`/`includes(403)` та однореченнєвий phase-order коментар; R30 рахує лише explicit safe access log, не зараховує `off` і падає без safe. Обидва `sh -n` зелені; R30 mutation червона потрібним текстом до Docker, restore nginx.conf має exit 0, baseline integration exit 0. B23 smoke не запускався й не заявлений доведеним; Go, B36, trap і фінальний nginx.conf не змінювалися.

РІШЕННЯ: п’яте коло завершено двома вузькими assertions. Git index/commit не чіпалися; коміт і Docker smoke B23 лишаються рецензенту.
ДАЛІ: рецензент запускає Docker smoke/повний гейт; очікування B23 — у відповідях присутні і 429, і 403 без вимоги конкретного розподілу.

---

ФАКТ (рецензент, фінальний гейт 2026-09-12, після п'ятого кола): `make check` → **GATE_EXIT=0**, лог
`make-check-t14-final.log`. `format-check`, `lint`, `typecheck` ok; `npm test` **379 passed, 0 failed**;
Go `-race` **48 пакетів ok, 0 FAIL**; `integration` ok, жодного повідомлення хостової перевірки R30;
`check-generated`, `build`, `up`, `seed-identities`, `seed` ok; smoke —
`PASS unavailable Redis returns 503 with Retry-After`, `PASS Redis recovery restored media authorization
policy (503 -> 403)`, `PASS media rate limit`; e2e **60 passed (8.6m)**.

ФАКТ: отже T14 закриває **B23** (пара 429/403) і **R30** (хостова перевірка nginx + Go AST, обидві половини в
умовах гейта) і **частково B36** (503 і `Retry-After`; «жодного байта медіа» — T16, поправка плану `8cc1d72`).

ФАКТ: обмеження, записані в `docs/playback/report.md`: сканер R30 читає лише `nginx.conf`, файли з `include`
не скануються. Гейт прогнано до коміту T14, на вмісті, який цей коміт фіксує.
