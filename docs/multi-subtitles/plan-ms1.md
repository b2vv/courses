# План зрізу MS1 — мова аудіодоріжки (D2) і легасі-мова (D8)

Фаза: [plan.md](plan.md) (ACCEPTED r6). Зріз реалізує D2 і D8 і нічого з D1/D3–D7: плеєр, ASR і
редактор транскрипту не змінюються. Статус: **ACCEPTED (Codex plan-critic r3, 2026-11-14; r1–r2 BLOCKED, закриті — `_wip-ms1-plan-critique-{1..3}.md`).**

## 1. Межі

**В межах:** правило мови доріжки (backend-домен, DB CHECK, frontend-домен); унікальність мови без
урахування регістру; міграція, що відмовляє на невідповідних рядках; легасі-міграція `und`→мова
primary (O6); режим `media-worker regenerate-masters` і його запуск у compose.

**Поза межами:** будь-яка зміна читання транскрипту, ASR, редактора, плеєра (MS2–MS4); переклад;
видалення транскриптів; мова S7-пам'яті плеєра (нормалізує MS4).

## 2. Факти (звірено з кодом 2026-11-14, HEAD `895b2e2`)

- Правило мови доріжки — `validLanguage` (`backend/internal/media/domain/audiotrack.go:267-281`):
  1–35 символів `[A-Za-z0-9-]`, регістр зберігається; ним користуються конструктори й rename
  (`validateLanguageLabel`, `:241`). DB CHECK — лише довжина (`00025_multi_audio.sql:13`).
- Унікальність — `audio_tracks_language_per_asset ON audio_tracks (video_asset_id, language) WHERE
  state <> 'failed'` (`00026`), **з урахуванням регістру**. Порушення мапиться на
  `ErrDuplicateLanguage` **за ім'ям обмеження** (`media/adapters/postgres/audiotrack.go:855`) → HTTP
  409 `duplicate_language` (`media/adapters/http/handler.go:525`).
- Правило мови транскрипту (ціль D2) — `transcript/domain/transcript.go:36-61` і CHECK `00007`:
  `^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$`, 2–35.
- Frontend: `validAudioLanguage` (`frontend/src/modules/media/domain/audio-track.ts:90`); список мов
  автора `['uk','en','pl','de','fr','es','und']` (`:80`); `offeredLanguages` порівнює мови точно
  (`:130-142`). Тест приймає `'x'.repeat(35)` (`audio-track.test.ts:38`).
- Embedded нового аплоаду — `und` (`media/application/transcode.go:468-471`); backfill 00028 дає `und`
  або `und-x-original` при колізії; прецедент доказу міграції на populated data —
  `media/adapters/postgres/backfill_integration_test.go`.
- Регенерація master: репозиторій `RegenerateMaster` блокує ассет `FOR UPDATE`, потім доріжки, і
  викликає `rewrite(ready)` під замками (`media/adapters/postgres/audiotrack.go:297-330`); застосунок
  `AudioTracks.rewrite` пропускає порожній набір і кличе writer `RegenerateMaster`
  (`media/application/audiotrack.go:72-83`; writer — `media/adapters/ffmpeg/master.go:28`).
- `cmd/media-worker/main.go:39-140`: `run` відкриває залежності, робить `Recover`, `Sweep`, стартує
  asynq і HTTP. Підкоманд немає. Образ `worker` має `ENTRYPOINT ["/app/media-worker"]`
  (`backend/Dockerfile:47`). У compose `media-worker` залежить від `migrate` (completed) і `upload`
  (started, K13) (`docker-compose.yml:102-125`); `make e2e` піднімає стек `up -d --build --wait` без
  скидання томів (`Makefile:85-86`), тож retained dev-дані зустрінуть міграцію.

## 3. Рішення

### M-D1. Одне правило мови — чистий предикат у `internal/contracts/language`
Граматика тегу (`^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$`, 2–35) переїжджає з `transcript/domain/transcript.go:48-61`
(`validLanguageTag`) у `internal/contracts/language.ValidTag(lowered string) bool` — стандартна
бібліотека, без I/O, без імпорту модулів (ADR 013 `docs/adr/013-one-access-rule-as-a-contract.md:19`
обирає одне чисте правило замість копій). Нормалізація лишається модульною: transcript
`NewLanguage` — `TrimSpace` + `ToLower` + `ValidTag` (поведінка не змінюється, `" uk "` → `uk`);
media `validLanguage(v)` — **без** trim (мова доріжки вже потрапляє в маніфест як є) → `ValidTag(lower(v))`.
Тест-таблиця живе в `contracts/language`; кожен модуль має лише тест своєї нормалізації. Наявний
`platform/architecture/contracts_test.go` перевіряє, що пакет не імпортує модулів.
Frontend `validAudioLanguage` отримує ту саму граматику (регулярка + довжина) з тією самою таблицею
в тесті (дзеркало між мовами програмування неминуче; таблицю позначено посиланням на Go-тест).

### M-D2. Унікальність без регістру — той самий індекс, нове вираження
Міграція `00030` дропає `audio_tracks_language_per_asset` і створює **з тим самим ім'ям** на
`(video_asset_id, lower(language)) WHERE state <> 'failed'`. Ім'я зберігається, тож мапа помилки
(`audiotrack.go:855`) і HTTP 409 не змінюються. `availableAudioLanguages` на фронтенді
(`audio-track.ts:129`) порівнює `lower()`.

### M-D3. Міграція `00030` — preflight, потім релейбл
Одна транзакція, `SET LOCAL lock_timeout = '5s'` і `LOCK TABLE video_assets, audio_tracks, transcripts
IN SHARE ROW EXCLUSIVE MODE` (прецедент 00028). Кроки **в цьому порядку**:
1. **Preflight на вихідних рядках (до будь-якої зміни).** (a) Будь-яка доріжка (failed теж — CHECK діє
   на всі рядки), чия мова не проходить M-D1, → `RAISE EXCEPTION` з кількістю й id першого рядка.
   (b) Два не-failed рядки одного ассета з однаковою `lower(language)` → `RAISE EXCEPTION` з asset id.
   Так `und-` (валідне за старим правилом, невалідне за новим) і пара `und`/`UND` відмовляють, а не
   «лікуються» релейблом (r1 blocker 1).
2. **D8-релейбл (O6).** Для кожного Ready не-evicted ассета, чий урок має primary-транскрипт мовою P, і
   на якому немає не-failed доріжки з `lower(language) = P`: рівно одна не-failed доріжка з
   `lower(language) = 'und' OR lower(language) LIKE 'und-%'` отримує `language = P`,
   `updated_at = now()`. Вибір: `is_default` DESC, `kind = 'embedded'` DESC, `id` ASC
   (`DISTINCT ON (video_asset_id)`). Після preflight `und`-родина вже валідна, а P проходить граматику
   транскрипту, тож релейбл не створює ні невідповідних рядків, ні колізій (умова «немає доріжки P»).
3. `ALTER TABLE audio_tracks ADD CONSTRAINT audio_tracks_language_shape CHECK (…)` — граматика M-D1 на
   `lower(language)`.
4. Новий індекс M-D2.
Down: дропає CHECK, повертає case-sensitive індекс; мови кроку 2 **не повертає** (down не знає, які
були `und`, і повернення забрало б субтитри) — коментар у міграції.

**Діагностика (r1 blocker 2).** `cmd/migrate` зараз ховає будь-яку помилку за загальним текстом
(`backend/cmd/migrate/main.go:67`) — свідомо, щоб SQL-деталі не йшли в лог. Зміна вузька: якщо в
ланцюгу помилки є `*pgconn.PgError` з SQLSTATE `P0001` (`raise_exception` — лише наші власні
`RAISE EXCEPTION`, текст яких пишемо ми: id і кількості, без даних користувачів; P0001 сам по собі
не гарантує авторства — правило для майбутніх міграцій: повідомлення `RAISE EXCEPTION` містить лише
ідентифікатори й числа, коментар про це — у `cmd/migrate`), migrate логує
`migration`, `sqlstate`, `message`; інші помилки — як зараз. Успішні лічильники (змінено/пропущено)
міграція не друкує — `RAISE NOTICE` ніхто не слухає; для retained-тому в звіті зрізу — read-only
запит-діагностика (скільки `und`-доріжок Ready-ассетів з primary лишилось і чому) і його результат на
dev-томі до/після.

### M-D4. `media-worker regenerate-masters` — одноразовий режим
`main` розгалужується за `os.Args[1]`: `regenerate-masters` → `regenerate(ctx, logger)`, інакше —
наявний `run`. `regenerate` відкриває ті самі залежності (`runtime.Open` вимагає Redis, ініціалізує
tracing/metrics, HTTP не стартує) і storage з `mediaRoot`, **не** робить Recover/Sweep, не стартує
asynq/HTTP. Тіло — функція `regenerateMasters(ctx, database, root string) (Result, error)` у пакеті
`cmd/media-worker`, яка будує **ті самі** production-адаптери (`mediapostgres.NewRepository`,
`mediaffmpeg.New(root)`, `mediaapplication.NewAudioTracks`) і кличе
`AudioTracks.RegenerateAllMasters(ctx)`: список id Ready не-evicted ассетів (новий метод репозиторію,
SELECT за `generation`, `id`) → для кожного `repository.RegenerateMaster(ctx, id, service.rewrite(ctx,
id))` — та сама серіалізована регенерація, що й при rename. Помилка одного ассета не зупиняє інші;
`Result{Regenerated, Skipped, Failed []assetID}`; ассет, що став evicted між списком і регенерацією
(`ErrVideoNotFound`), — Skipped. Є Failed → лог з id і **exit 1**.

### M-D5. Запуск у compose — сервіс `media-masters`
Новий сервіс з тим самим `build: target: worker`, `command: ["regenerate-masters"]`, томом
`media-data`, змінними БД/Redis як у `media-worker`, `restart: 'no'`, залежить від `migrate`
(completed) і `upload` (started, K13). `media-worker` отримує `media-masters: { condition:
service_completed_successfully }`. Тож кожен `up` (dev, `make e2e`) після міграцій регенерує master'и,
а збій валить `up --wait`. Ціна: перезапис усіх Ready master'ів на кожен старт стеку; master — кілька
сотень байт тексту на ассет, запис атомарний (`master.go`). Відкинуто: крок у `run` media-worker'а
(суперечить plan.md D8 «окремий режим, не стартує recovery…») і ручна make-ціль (забувається; MS4
вимагає, щоб регенерація гарантовано відбулась у потоці).

## 4. Рядки приймання

| # | Поведінка | Доказ (продакшн-шлях) |
|---|---|---|
| M-R1 | Предикат приймає `uk`, `en-us`, `pt-br`, `und`, `und-x-original`, `zh-hant-tw`; відмовляє `e`, `1x`, `-en`, `en-`, `a--b`, `und-`, `x`×35, `x`×36, `uk_ua`, `укр`. Media: `en-US` ок, `" uk "` відмова; transcript: `" uk "` → `uk` (без змін) | unit `contracts/language`, `media/domain/audiotrack_test.go`, `transcript/domain/transcript_test.go` |
| M-R2 | DB відмовляє невідповідну мову прямим INSERT/UPDATE (обхід домену) | integration, `audio_tracks_language_shape` |
| M-R3 | `en` потім `EN` на одному ассеті → 409 `duplicate_language`; rename `de`→`EN` при наявній `en` → 409; failed `en` + новий `EN` → дозволено | integration repo + HTTP contract-тест |
| M-R4 | Міграція на populated data: (a) мова `e` → відмова з id; (b) `und-` у Ready-ассета з primary `uk` → відмова (не релейбл); (c) `und`+`UND` з primary `uk` → відмова з asset id; (d) чисті дані → проходить. Після відмови **кожен вихідний рядок незмінний** (rollback) | integration у стилі `backfill_integration_test.go` (UpTo 29 → seed → Up) |
| M-R5 | D8: `und` default + primary `uk` → `uk`; `und`+`und-x-original` → одна `uk` (default/embedded/id), інша лишається; є доріжка `uk` → без змін; evicted / не-Ready / урок без primary → без змін; primary `en-us` → доріжка `en-us` | той самий integration, перевірка кожного рядка |
| M-R6 | Down 00030 повертає case-sensitive індекс, прибирає CHECK, мови кроку 2 лишаються | integration DownTo 29 |
| M-R7 | **Наскрізно:** populated v29 + справжній master, записаний writer'ом зі старими даними (`LANGUAGE="und"`) → Up 00030 → `regenerateMasters` (production-адаптери, тимчасовий root) → master містить `LANGUAGE="uk"`; другий запуск — байт-у-байт той самий файл; evicted ассет не торкнуто | integration у `cmd/media-worker` проти testcontainers Postgres |
| M-R8 | Гілки збою: writer падає на одному ассеті → інші регенеровані, `Failed` містить id, процес — exit 1; evicted між списком і регенерацією → Skipped; `regenerate` не викликає Recover/Sweep/queue | unit застосунку (фейк лише writer'а для failure-гілки) + тест диспетчера `main` |
| M-R9 | migrate логує `sqlstate=P0001` і текст нашого `RAISE EXCEPTION`; іншу SQL-помилку — загальним текстом | unit `cmd/migrate`: production-функція діагностики на `*pgconn.PgError`, **загорнутому** як у goose (`PartialError.Unwrap`), і на загорнутій не-P0001 помилці (редакція лишається) |
| M-R10 | HTTP: mint і rename з `e` / `und-` → 400 `invalid_request`, збережені метадані без змін | contract-тест HTTP (`handler.go:507`) |
| M-R11 | Frontend: `validAudioLanguage` — та сама таблиця; `availableAudioLanguages` не пропонує мову, наявну в іншому регістрі (новий аплоад), а при редагуванні доріжки з мовою `EN` лишає її власну опцію | unit `audio-track.test.ts` |
| M-R12 | Compose, три сценарії; `docker events --filter type=container` запускається **до** `up`, порядок — за подіями `start`/`die`: (a) **зупинений стек, збережений том** (`docker compose stop`, потім `make up`) — `media-masters` `die` з exit 0 **до** `start` `media-worker`; master легасі-ассета містить мову primary; повторено двічі; (b) **`up` по запущеному стеку** — `media-masters` (завершений one-shot) запускається знову й завершується 0, `media-worker` лишається запущеним без перезапуску; це задумано: регенерація ідемпотентна й серіалізована тими самими замками ассета, що й писачі worker'а (`audiotrack.go:297`); (c) **збій регенерації** на зупиненому стеку (`media-masters` з недійсним `POSTGRES_PASSWORD` через override-файл у scratchpad) — `up --wait` завершується ненульовим кодом, `media-worker` не стартує (немає події `start`). На `make check` (порожній том) — 0 регенерацій, e2e зелений | живий прогін, події й коди виходу у звіті + ordered `make check` |

## 5. Перевірка

Ordered `make check` зеленим на фінальному коміті; повний `make integration` (змінюється злита
поведінка media). Мутації: прибрати `lower()` з індексу → M-R3 червоний; прибрати крок релейблу → M-R5
червоний; поміняти місцями preflight і релейбл → M-R4(b,c) червоний; прибрати граматику з домену →
M-R1 червоний; зламати `root` у `regenerateMasters` → M-R7 червоний. Незалежне рев'ю Codex хвилями до SHIP.

## 6. Ліміти

- Міграція відмовляє замість виправляти: retained dev-том з доріжкою на кшталт `e` зупинить `up`
  до ручного виправлення (механізм: повідомлення з id і кількістю; ремедіація — SQL UPDATE/DELETE).
- `media-masters` перезаписує всі Ready master'и на кожному старті стеку; на кількості ассетів dev/e2e
  (десятки) це секунди — замір часу сервісу в звіті зрізу.
- Down не повертає мови D8 (M-D3) — свідомо.
