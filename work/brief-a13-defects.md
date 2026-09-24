# Brief A13 / крок 2 — дефекти C10 і C14

Частина A13 (закрити приймання Phase 8 до 22/22). Крок 1 (model-path) — окремо, уже зроблений.
Три дефекти, підтверджені проти коду на 54fc22c: C10, C14 і перегенерація, мовчки заблокована
ключем завершеної задачі (ламає C04; див. C10, дефект 2). Тримай C10+C04 і C14 логічно окремо —
комітимуться двома комітами. Бриф переписала сесія Courses 1 2026-09-24 після звірки з кодом
фронтенду й asynq; рецензент — Courses 2.

## Read first
- `docs/asr/spec.md` §11 (**D8-6** ідемпотентність за `assetID:generation`, конфлікт **видимий**;
  **D8-5** відновлення на старті: `running` зі старим `updated_at` → в чергу або `failed` з
  названою причиною) і §12 рядки **C10**, **C14**, **C03**.
- `backend/internal/transcript/application/recognize.go` — весь файл (`Process`, `fail`,
  `complete`, `Recover`, `failBeforeStart`, `TaskTimeout`).
- `backend/internal/transcript/adapters/asynq/queue.go` — весь файл.
- `backend/internal/transcript/adapters/http/draft_handler.go` (handler `Regenerate` ~175-196)
  і `backend/internal/transcript/adapters/http/handler.go` `writeTranscriptError` (119-146).
- `backend/internal/transcript/domain/errors.go` (коди помилок — рядок-тип `Error`).
- Прецедент тестів черги: `backend/internal/media/adapters/asynq/queue_test.go`.
- Прецедент integration-харнесу: `backend/internal/transcript/adapters/postgres/draft_integration_test.go`
  (`//go:build integration`, testcontainers Postgres) і будь-який redis-integration прецедент
  (`backend/internal/playback/adapters/redis/*_integration_test.go`).

---

## Задача C10 — «вже виконується» має бути видимим і правдивим; перегенерація після завершення має працювати

**Дефект 1 (перевірено):** `queue.go:44` будує `taskID` з `LessonID`, а D8-6/C10 вимагають
`assetID:generation`. `queue.go:50-52` перетворює `ErrTaskIDConflict` на `nil`: це тихий успіх, HTTP
віддає 202 (`draft_handler.go:195`). C10 (`docs/asr/spec.md:1196`, D8-6 `:1150`) вимагає видиму
відповідь «вже виконується».

**Дефект 2 (перевірено за кодом asynq v0.26.0; runtime-доказ — твоїм тестом):** asynq вирішує
конфлікт через `EXISTS` на ключ задачі (`internal/rdb/rdb.go:98-101`, `enqueueCmd`). Завершена задача
ключ **не видаляє**: `markAsCompleteCmd` робить `HSET … state completed` і `EXPIREAT` через
`Retention` (`rdb.go:382-395`), а в нас `Retention(24h)`. Задача, що повернула помилку, з
`MaxRetry(0)` іде в архів, і ключ живе до 90 днів (`rdb.go:840-841`, `archivedExpirationInDays`).
Прогін, що впав із названою причиною (`engine_error` тощо), `Process` завершує без помилки →
`completed` → блок на 24 год. **Наслідок:** перегенерація того самого `assetID:generation` після
будь-якого завершеного прогону зараз мовчки нічого не робить — до 24 год, а після архіву до 90 днів.
Це ламає C04 («перегенерація перезаписує чернетку») у проді; доказ C04 стояв на фейковій черзі.
Якщо просто замінити `nil` на помилку, автор бачитиме «вже виконується» про задачу, що давно впала.
**Відповідь «вже виконується» мусить означати справжній незавершений прогін, а не залишок ключа.**

**Код помилки — `already_running`, HTTP 409.** Це вже контракт фронтенду:
`frontend/src/adapters/api/draft-repository.ts:112` (`409: ['already_running']`),
`frontend/src/domain/transcript/draft.ts:31`, показ — `LessonDraftField.tsx:51`. Бекенд його досі
ніколи не віддавав. **Інший рядок (напр. `asr_recognition_in_progress`) — баг**: UI не впізнає код.

**Рекомендований шов (інший — через `## Discussion` з доказом):**
1. `domain/errors.go`: `ErrAlreadyRunning Error = "already_running"` (стиль наявних кодів).
2. `Drafting.Regenerate`: перед постановкою прочитати чернетку уроку. Якщо вона для **поточних**
   `assetID:generation` у стані `running` — повернути `ErrAlreadyRunning`, у чергу не ставити.
3. `taskID` = `assetID:generation:<маркер>`, де маркер — токен чернетки (її `updated_at`) на
   момент запиту, або стала `none`, якщо чернетки ще нема. Два одночасні запити до `Claim`
   прочитають той самий токен → той самий ID → конфлікт → `ErrAlreadyRunning`.
   Маркер — стабільний рядок: `updated_at` у фіксованому форматі (UTC, фіксована точність), щоб
   однаковий токен завжди давав однаковий ID, а `taskID` лишався валідним рядком для asynq. Після завершення
   токен уже інший → новий ID → постановка проходить, хоч старий ключ ще живий. Дотримайся того
   самого правила й у `EnqueuePending` (автопостановка) — один конструктор ID, не два.
4. `queue.go`: `ErrTaskIDConflict` → `domain.ErrAlreadyRunning` (не `nil`).
5. `handler.go` `writeTranscriptError`: `ErrAlreadyRunning` → 409, код `"already_running"`.

**Тести (C10 і C04 — integration проти справжнього Redis).** Харнес — testcontainers, прецедент
`backend/internal/playback/adapters/redis/*_integration_test.go`. Доведи на справжній asynq:
- (a) дві постановки з тим самим маркером → друга `ErrAlreadyRunning`;
- (b) **після того, як задачу справді оброблено** (підніми `asynq.Server` з обробником, що
  повертає `nil`, дочекайся completed) — постановка з новим маркером проходить, а зі старим дає
  конфлікт. Саме це доводить, що ключ після завершення живий і маркер це обходить;
- (c) те саме для задачі, що повернула помилку й пішла в архів.
Плюс unit на `Regenerate`: `running` для поточних `assetID:generation` → `ErrAlreadyRunning` без
виклику черги. Плюс контракт-тест HTTP: 409 + `{"error":{"code":"already_running"}}`.
Якщо Redis-харнес не піднімається без Docker — код і тести все одно напиши, запускати їх не
треба, **і так і познач**: не ганяв.

## Задача C14 — таймаут задачі має знімати чернетку з названою причиною

**Дефект (перевірено):** `recognize.go` рядки 73-74 і 82-83: на `ctx.Err() != nil` `Process`
повертає `ctx.Err()` **не викликаючи** `r.fail(...)`. Чернетка лишається `running`. Єдине, що її
згодом фейлить — `Recover` (рядок 178, причина `"timeout"`), але він біжить **лише на старті
воркера** (`cmd/asr-worker/main.go:63`). Отже після перевищення таймауту (похідного від
`duration_seconds`) автор бачить `running` **до рестарту воркера** — порушує C14 («знімається з
названою причиною») і б'є по C03 («після падіння — названу причину»).

**Зроби (рекомендований шов — якщо бачиш кращий, відкрий Discussion):**
1. У `Process`, коли `ctx.Err() != nil`, розрізняй:
   - `errors.Is(ctx.Err(), context.DeadlineExceeded)` — задача перевищила похідний таймаут:
     зафейлити чернетку з причиною `"timeout"` (та сама причина, що вживає `Recover`).
   - `errors.Is(ctx.Err(), context.Canceled)` — це graceful shutdown воркера: **лишити як є**
     (повернути `ctx.Err()`), бо чернетку підбере `Recover` на наступному старті (D8-5). Не
     фейлити на shutdown.
2. **Термінальний запис на мертвому контексті.** На `DeadlineExceeded` вихідний `ctx` вже
   протермінований, тож `r.drafts.Complete(ctx, ...)` одразу впаде. Зроби запис на контексті,
   **відчепленому від дедлайну задачі**: `context.WithoutCancel(ctx)` (Go 1.27 має його) +
   короткий `context.WithTimeout`. Причина — `"timeout"`, стан `DraftFailed`, тим самим шляхом
   `fail`/`complete`. Причина `timeout` уже є в зафіксованому списку метрик
   (`platform/observability/metrics.go:53`), тож новий лейбл не потрібен. Після такого `fail`
   `Process` повертає `nil`: задача стає `completed`, а не архівною, а маркер у `taskID` (C10) не дасть
   старому ключу заблокувати наступну перегенерацію.
3. Застосуй до **обох** місць (`audio.Extract` і `Transcribe`), щоб таймаут на будь-якій фазі
   давав названу відмову. `NULL`-тривалість (`failBeforeStart "unknown_duration"`) уже
   покрита — не чіпай.
4. **Тест рівня C14 (integration).** За прецедентом
   `postgres/draft_integration_test.go` (реальний Postgres): задача з fake-transcriber (або
   fake-audio), що блокується довше за похідний таймаут, лишає рядок чернетки у стані
   `failed`/`timeout`, **не** `running`. Якщо треба, задай малий таймаут через контрольований
   `duration_seconds`, щоб тест був швидким.

---

## NOT (не твоє в цьому кроці)
- НЕ чіпай model-path (Dockerfile, compose, `main.go` `modelReady`, smoke) — крок 1, зроблено.
- НЕ чіпай фронтенд і локалі — крок wiring.
- НЕ чіпай `backend/openapi/openapi.yaml` — відсутність draft-ендпоінтів там окремий борг, не цей крок.
- НЕ додавай speech-фікстур/e2e і НЕ став опційні integration-докази C21/C22/C09/C01 — окремі кроки.
- НЕ міняй `Recover`, `EnqueuePending`, `failBeforeStart` по суті (хіба що спільний хелпер для
  термінального запису — тоді мінімально й з тестом).

## Обмеження середовища
- Пісочниця **без Docker**: integration-тести (testcontainers) і ordered gate ти **не запускаєш**.
  Напиши код і тести, і в звіті **прямо познач**, що integration/gate НЕ ганяв — їх жену я.
  Ніколи не став «пройдено» те, чого не ганяв. Що можеш локально (gofmt, `go build ./...`,
  unit `go test` без тега `integration`) — жени й познач результат чесно.
- Не чіпай git; не редагуй `work/tasks.md`; не пиши `docs/asr/report.md`; не пиши в `generated/`.

## Якщо бачиш неточність, що дасть баг
Відкрий `## Discussion` у звіті (що не так, file:line, що станеться в рантаймі, пропозиція, що
зробив тим часом). Не обходь мовчки й не виконуй проти власного судження.

## Звіт
`work/report-a13-defects.md`
