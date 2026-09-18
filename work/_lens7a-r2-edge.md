# Лінза EDGE CASES — раунд 2, редакція 5 плану

ДАЛІ: прочитати `docs/transcript/plan.md` (ред. 5) і `docs/transcript/spec.md` (§5.3 — шість рішень grilling),
скласти карту рядків §6 (P/B/N/R) і рівнів доказу; потім по кожному з G1–G6 шукати межу без рядка.

ФАКТ: план ред.5 прочитано цілком (528 рядків), spec §5.3 прочитано. §6 = P01–P10, B01–B22, N01–N21, R01–R11.
ФАКТ: D-9 (оптимістичний замок, `transcript_modified`) — у §6 НЕМАЄ жодного рядка (перевірено переглядом усіх чотирьох таблиць).
ФАКТ: D-7 обіцяє «Окремий рядок міряє саме це» про збій посеред 5000 рядків — такого рядка в §6 немає.
ФАКТ: D-5 вирішує «клік у такому уроці вимкнений» (duration=0) — рядка немає; B09 міряє лише читання.
ДАЛІ: звірити код — progress.go (ApplyProgress, завершений урок), LessonPlayer (seekTo/duration), володіння+витіснення Phase 6, maximumAdminWriteBody, nginx (client_max_body_size, gzip), пін стель, як біжать integration vs e2e.

## Перевірені координати (усі відкриті в джерелі)

ФАКТ: `infrastructure/nginx/nginx.conf:20` — `client_max_body_size 1m;` глобально; `location ^~ /api/`
(:116-126) його НЕ перевизначає; перевизначення є лише в `/uploads/` (:92, 16m).
ФАКТ: `nginx.conf` не містить жодної директиви `gzip*` (перевірено grep) — G2 правий.
ФАКТ: `make integration` = `infrastructure/tests/integration.sh` = host-awk по nginx.conf + `docker run golang … go test -tags=integration ./…`
з монтуванням лише `backend/`. HTTP через nginx там неможливий. Через nginx ідуть `smoke`
(`smoke.sh`, `http://nginx:8080`) і `e2e` (playwright baseURL `http://localhost:8080`).
ФАКТ: у `smoke.sh` уже є дослівний прецедент перевірки ліміту тіла на nginx: 32 MiB PATCH → 413 + лічильник
доводить, що запит не дійшов до сервісу.
ФАКТ: `LessonPlayer.tsx` — `const [duration, setDuration] = useState(0)`; `duration` оновлюється лише
`onDurationChange` / `onReady` медіа, а медіа вантажиться лише після `loadSession` (кнопка Play,
`sessionRef.current === undefined → loadSession(position,'initial')`). `seekTo` (:404-409) клампить у
`Math.max(1, duration)` → до першого Play будь-який клік дає 1 c у БУДЬ-ЯКОМУ уроці, не лише без відео.
Кнопка Play стартує сесію з `position` — тобто зіпсованого кліком значення.
ФАКТ: `heartbeatNow` повертає 'stop' при `sessionRef.current === undefined` → до старту сесії клік нічого
не пише; `ApplyProgress` (`playback/domain/progress.go:68-70`) на ЗАВЕРШЕНОМУ уроці відкидає рух назад
навіть для подієвих reason (`saved.CompletedAt != nil`).
ФАКТ: `playback/application/sessions.go:57-63` — прогрес пишеться лише при `OwnershipRenewed`; при
витісненні heartbeat від кліка повертає `playback_taken_over`, прогрес не пишеться, плеєр іде в
`terminalFailure('taken-over')` і `<video>` зникає з DOM.
ФАКТ: `contract_ceilings_test.go` порівнює доменну константу з `maxItems`/`maxLength` в openapi. Для
байтової константи тіла контрактного літерала не існує — «зареєструвати тим самим піном» (plan:117-118)
у цій формі нездійсненно; справжня друга копія числа лежить у nginx.
ФАКТ: покриття критеріїв рядками §6 — усі, КРІМ `A24` (медіабайти в обхід `auth_request`). Рядка немає.
ФАКТ: `G3` не згадується в `plan.md` жодного разу (grep = 0); G1/G2/G4/G5/G6 мають рядки N15/N14/N09-11/N16/N05-06.
ФАКТ: `backend/sqlc.yaml` має 3 пакети (course, identity, media); `make check-generated` (Makefile:79-81)
дифить лише course+identity+frontend — media вже випав. Новий `transcript`-пакет так само не буде покритий.

## Знахідки

[edge] BLOCKING 1 — nginx ріже тіло на 1 MiB раніше, ніж діє власний ліміт G1
[edge] BLOCKING 2 — ліміт тіла не виводиться з арифметики 5000 × maxLength(text) × 3 документи
[edge] BLOCKING 3 — gzip (G2) міряється рівнем, який nginx не бачить; «не для медіабайтів» без рядка
[edge] BLOCKING 4 — duration=0 не лише в уроці без відео: клік до першого Play дає 1 c і псує старт
[edge] BLOCKING 5 — G3 не має жодного рядка; «дія пише» хибне до старту сесії й у завершеному уроці
[edge] BLOCKING 6 — два транскрипти різними мовами (їх створює сам N08) × N10 «віддає єдину наявну»
[edge] BLOCKING 7 — D-9 без рядків; гонка двох імпортів у порожній транскрипт не має токена
[edge] BLOCKING 8 — обіцяний «окремий рядок» на збій посеред заміни відсутній; заміна порожнім без рядка
[edge] BLOCKING 9 — стеля на результаті імпорту (G5) не має рядка
[edge] non-blocking 10 — A24 єдиний критерій без рядка, а `<track>` дивиться саме в цей бік
[edge] non-blocking 11 — фікстура N14 не названа (без documentLinks замір не є стелею)
[edge] non-blocking 12 — клік після витіснення/помилки плеєра не визначений
[edge] non-blocking 13 — `check-generated` не покриє sqlc-генерацію transcript
[edge] non-blocking 14 — дублікат документа в репліці / однакові `ordinal` → PK-конфлікт замість коду
[edge] non-blocking 15 — документ під посиланням невилучний (D-8) + стеля 200 = курс без шляху вниз

РІШЕННЯ: 9 blocking, 6 non-blocking. Усі координати відкриті в джерелі, жодна не взята з переказу.
