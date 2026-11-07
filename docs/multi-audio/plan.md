# План — multi-audio (кілька аудіо-доріжок на урок)

Фаза поза роадмапом §24. Крок B роботи над перемиканням аудіо (крок A = seamless quality switch,
змерджено PR #96 → `69d51a8`). Статус: **ПЛАН, пройшов plan-critique (4/4 лінзи, 0 blocking) і
plan-defense (5/5, 0 gap); до затвердження власником; коду не писати**.

Факти про код — з карти seam'ів (Explore) і S0-ресьорчу (2026-11-07), посилання file:line нижче.

## 1. Мета й модель власника

Урок може мати кілька аудіо-доріжок (оригінал + дубляж). Глядач перемикає доріжку в меню плеєра з
мінімальним gap. Джерело доріжок (рішення власника 2026-11-07): відео вантажиться **з вбудованим
аудіо або без**; додаткові аудіо — **окремими файлами**. Доріжки уроку = {вбудована, якщо є} ∪
{довантажені}. Німе відтворення заборонене (D4).

## 2. Поточний стан (seam'и, що доведеться змінити)

- Домен `VideoAsset` (`asset.go:52-68`), один ассет = один урок = рівно один файл (tus `server.go:32`).
  Стани `Created/Uploading/Queued/Processing/Ready/Failed` (`asset.go:19-29`). Поняття «доріжка» немає.
- `AcceptSource` **ВИМАГАЄ** прийнятного аудіокодека (`source.go:50-57`). `parseProbe` коректно лишає
  `AudioCodec==""` для відео-only (`transcoder.go:230-243`) — падає не probe, а транскод-крок.
- Транскодер безумовно `-map 0:a:0` + `-c:a:0 aac -profile:a:0 aac_low -b:a:0 128k -ac 2`
  (`transcoder.go:164-169`); fmp4, hls_time 6, init.mp4 (`:170-180`); var_stream_map фіксує одну
  аудіогрупу `a:0,agroup:aac,name:audio,default:yes` (`:294-300`). Аудіо вже пакується у власний
  `%v`-каталог `audio/` незалежно від відео-кейфреймів.
- Вихід: `hls/<assetID>/{master.m3u8, audio/, 360p/, 720p/...}`, права 0750/0640
  (`transcode.go:120-124`, `transcoder.go:23-27`). **У репо НЕ захоплено реальний master.m3u8**
  (`transcoder_test.go:29-32` перевіряє лише наявність файлів, не текст) — див. S3-передумову.
- Черга: Redis asynq, task `media:transcode`, payload `{assetId, attempt}` (`asynq/queue.go`),
  generic; тригер — завершення upload (`upload.go:143`). Upload-сервіс ключований на generic assetID
  + `StorageKey.Leaf` (відео-специфіка лише sniff контейнера + leaf `original`).
- Playback-контракт: `playResponse{playbackId, mediaPath=…/master.m3u8, …}` (`handler.go:77-120`) —
  аудіо/доріжок НЕМАЄ; фронт бере лише master.m3u8.
- ASR бере вхід ЖОРСТКО з `source/<assetID>/original`, один assetID → один шлях (`audio.go:46`).

## 3. Ключові архітектурні рішення (кожне закриває альтернативу)

### D1. Аудіо-доріжка пакується незалежно; відео НЕ ре-транскодиться; транскод — РОЗШИРЕННЯ наявної job
Кожна доріжка пакується окремим ffmpeg-прогоном у `audio-<trackId>/{init.mp4, segment_%05d.m4s,
index.m3u8}` тими самими параметрами (fmp4, hls_time 6, aac_low/128k/stereo). Відео-варіанти
незмінні. **НЕ нова asynq-task:** розширити payload наявного `media:transcode` опційним
track-дискримінатором і розгалузити в одному handler, переВикориставши чергу/error-handler/метрики
(crit: complexity). Детермінований **regen master.m3u8** (D7) — окремий малий крок, не job.
- Відкинуто: один ffmpeg-прогін з кількома `-map 0:a:N` (ре-рендер відео при кожному дубляжі).

### D2. Доріжки віддаються через HLS-маніфест; playback-DTO НЕ змінюється
hls.js читає `EXT-X-MEDIA:TYPE=AUDIO` і виставляє `hls.audioTracks` (підтверджено S0:
`audio-track-controller.ts:79-84`, парсинг `m3u8-parser.ts:240`). `playResponse`/`PlaybackSession`,
mediaPath-regex — незмінні; **конфлікту з security-spike на playback-поверхні немає**.
- Відкинуто: список доріжок у playResponse (дублює маніфест, чіпає openapigen дарма).

### D3. Домен: lean `AudioTrack` metadata, що ПЕРЕВИКОРИСТОВУЄ наявну upload/queue-машинерію
Нова таблиця/сутність per-track: `AudioTrack{ID, VideoAssetID, Kind(embedded|uploaded), Language,
Label, IsDefault, State, SourceLeaf}`. **Машинерію upload/queue/transcode НЕ клонувати** (crit:
complexity, blocking): uploaded-доріжка переВикористовує наявні `Tickets.MintTicket` + `Upload` +
asynq-чергу, лише з `StorageKey.Leaf="audio-<trackId>"` (замість `original`) і track-дискримінатором
у payload. Нове й мале — тільки track-metadata.
- **Два Kind — різні життєві цикли** (crit: ai-slop+complexity): `Kind:uploaded` — повний arc
  (Created→Uploading→Queued→Processing→Ready/Failed); `Kind:embedded` — **похідний Ready-рядок** без
  upload-станів, матеріалізується при завершенні відео-транскоду (хто саме ставить Ready — D4).

### D4. Повна стан-машина; Ready = ≥1 Ready-доріжка, інваріант діє ВЕСЬ час (рішення власника + crit)
- `AcceptSource` послаблюється приймати `AudioCodec==""` **лише для video-source (leaf `original`)**.
  **АЛЕ** разом із цим (той самий зріз — crit edge+gaps, blocking) відео-транскод для audio-less
  джерела генерує аргументи **БЕЗ `-map 0:a:0`** і var_stream_map **БЕЗ аудіогрупи**; master стартує
  без `EXT-X-MEDIA`; embedded не матеріалізується. Мердж релаксації AcceptSource ЗАБОРОНЕНО раніше за
  цю зміну транскодера.
- **Malformed input — завантажений аудіофайл (leaf `audio-<trackId>`)** (plan-defense Q1): проходить
  той самий `AcceptSource` + container-sniff, що й video (переВикористання, D3), **але послаблення
  вище на нього НЕ поширюється** — `AudioCodec==""` (файл без аудіопотоку, PDF-під-виглядом-аудіо,
  непідтримуваний кодек) для audio-доріжки = business-помилка на mint/accept, доріжка **не
  створюється**; інваріант «≥1 Ready» не зачіпається, бо битий вхід відхилено до появи рядка. Збій уже
  ПІД ЧАС транскоду прийнятого файлу → доріжка `Failed` (не відхилення входу) і йде стан-машиною вище.
- Новий проміжний стан `VideoAsset`: **`AwaitingAudio`** — відео-транскод завершився, Ready-доріжки
  ще нема. Переходи: відео-транскод ok + є embedded → одразу Ready; відео ok + немає аудіо →
  AwaitingAudio; у AwaitingAudio перша Ready uploaded-доріжка → Ready; **усі доріжки Failed і embedded
  нема → термінальний Failed уроку** (crit gaps). Повна таблиця переходів VideoAsset × AudioTrack — у S1.
- **Ready вимагає ≥1 Ready-доріжки й ЗАЛИШАЄТЬСЯ істинним весь час Ready** (crit edge, blocking):
  delete відмовляється business-помилкою, якщо опустив би к-сть Ready-доріжок нижче 1. Німе
  відтворення заборонене завжди, не лише на вході в Ready.
- **Рівно одна default** (crit edge, blocking): ≤1 default до Ready, рівно 1 після Ready. set-default
  = атомарний swap (знімає прапор з попередньої в тій самій транзакції). Видалення default-доріжки в
  тій самій транзакції переобирає нову (перша Ready за детермінованим порядком) або відмовляється.

### D5. ASR бере вхід із default-доріжки; timing і stale-транскрипт визначені (crit gaps)
Для embedded шлях лишається `source/<assetID>/original`; для uploaded-only —
`source/<assetID>/audio-<defaultTrackId>`. **ASR стартує лише коли default призначено** (а не «перша
Ready»), щоб не плодити stale-транскрипт. **Зміна IsDefault інвалідує й ре-ранить транскрипт**
(datoване рішення зрізу S6). До backfill старих ассетів — ASR fallback на `source/<id>/original`.

### D6. Авторинг-ендпоінти — у media-модулі, через openapigen
Нові HTTP (admin/instructor, як mint video): mint audio-ticket, list/delete track, set default, set
language/label; **унікальність (language, asset)** — datoване рішення S5 (або заборона дубля
business-помилкою, або дозвіл із обов'язковим різним Label і детермінованим порядком). Проходять через
`mediahttp.OpenAPIContribution()` (`contributions.go:21`) → чіпають openapigen-генератор (не
playback-DTO). Координація зі security-spike — §6.

### D7. master.m3u8 — ПОВНА регенерація з набору Ready-доріжок у БД, серіалізовано й атомарно
**Найважливіша структурна правка** (crit edge #2/#3/#8 + gaps #3, blocking; закриває 4 знахідки):
- НЕ інкрементальний append. master перегенеровується **повністю** зі списку Ready `AudioTrack` у БД
  (+ незмінні відео-варіанти) при кожній зміні.
- Під **per-asset локом** (єдиний письменник) → жодних lost-update від конкурентних transcode-audio.
- **Ідемпотентно** → повтор/ретрай задачі безпечний (набір той самий).
- **Атомарний запис**: temp-файл + `rename` на місце, права 0640, group-ownership як дерево
  (0750/0640) → nginx ніколи не читає напів-записаний master (AGENTS «Nginx serves media bytes»).
- master містить `EXT-X-MEDIA` **лише для Ready-доріжок**; доріжка стає Ready в ТІЙ САМІЙ
  серіалізованій операції, що регенерує master → failed/processing у master не потрапляють.
- Рядок: `#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID="aac",NAME="<label>",LANGUAGE="<lang>",DEFAULT=<YES|NO>,
  AUTOSELECT=YES,URI="audio-<trackId>/index.m3u8"`; `EXT-X-STREAM-INF` відео НЕ міняються (вже
  несуть `AUDIO="aac"`). Рівно один `DEFAULT=YES` на групу (з D4).
- **Передумова S3** (S0): спершу захопити реальний ffmpeg-вивід master у тест-прогоні, перш ніж
  писати regen — у репо його нема.

## 4. «Без ями» при перемиканні доріжки — вбудовано в hls.js (S0 підтвердив)
Перемикати доріжку треба `hls.nextAudioTrack = n` (НЕ `audioTrack`, що флашить негайно). S0:
`set nextAudioTrack` → `setAudioTrack(n,false)` → `flushImmediate:false` → префетч нової доріжки
(`nextLevelSwitch`), старий буфер флашиться лише від позиції завантаження вперед, коли PTS нової
відомий (`audio-stream-controller.ts:485,931-943,961`; дефолт `nextAudioTrackBufferFlushForwardOffset
0.25`). Тобто «не глушити стару, поки нова буферизована» — **бібліотекою, без ручної логіки** (аналог
`nextLevel` з кроку A). Справжній мікс-фейд (WebAudio, 2 джерела) на одному MSE неможливий — ліміт §7.

## 5. Розбивка на зрізи (leaf-first; кожен зелений окремо; незалежне рев'ю кожного — §8)

- **S0 — ресьорч (ЗРОБЛЕНО, без коду):** hls.js `nextAudioTrack`-семантика + ffmpeg-пакетування
  підтверджують D1/D2/D4/D7. Залишковий борг S0→S3: захопити реальний master.m3u8 з тест-прогону.
- **S1 — домен:** `AudioTrack` + два життєві цикли (D3); послаблення AcceptSource; **повна таблиця
  переходів VideoAsset × AudioTrack** з `AwaitingAudio` і fail-шляхами (D4); інваріанти «≥1 Ready» та
  «рівно одна default». Юніт-тести домену. Без транспорту/БД/транскодера.
- **S2 — persist:** міграція `audio_tracks`, repository, storage-ключі `source/<id>/audio-<trackId>`;
  set-default атомарний swap, delete з re-election — на рівні транзакцій. Repo-тести проти Postgres.
- **S3 — транскод доріжки + regen master (D1+D7), включно з embedded (злито S3+S4, crit complexity):**
  розширити payload `media:transcode`; ffmpeg пакетує один аудіофайл; **повна регенерація master під
  per-asset локом, атомарно** з набору Ready-доріжок; матеріалізація embedded-доріжки при відео-
  транскоді; **відео-only гілка без `-map 0:a:0`** (D4) у ТОМУ Ж зрізі, що релаксація AcceptSource.
  **Передумова:** захопити реальний master (S0-борг). Тести: вихідне дерево, валідність маніфесту,
  конкурентний regen (гонка двох доріжок), ретрай (ідемпотентність), відео-only.
- **S-backfill — окремий зріз (crit edge, blocking):** міграція `000NN` матеріалізує
  `AudioTrack{embedded,default}` для кожного наявного Ready-ассета з аудіо, **доведена на populated
  data** (прецедент: enrollment 00015); визначити ASR-fallback до backfill. Злити РАЗОМ з активацією
  Ready-ґейта (gaps #1), не раніше.
- **S5 — авторинг-контракт+UI:** mint audio-ticket + list/delete/default/language (openapigen,
  узгодити зі spike); admin/instructor UI «доріжки уроку»; унікальність (language,asset) — D6. **Лише
  цей зріз залежить від security-spike cutover** (gaps #4). Contract-тести HTTP.
- **S6 — ASR:** вхід із default-доріжки; ASR стартує по призначенні default; ре-ран при зміні default;
  fallback до backfill (D5). Тести, щоб наявний транскрипт не зламався.
- **S7 — frontend-плеєр:** `hls.audioTracks` у gear-меню (поряд speed/quality/captions); перемикання
  `hls.nextAudioTrack = n` (S0). e2e на реальному multi-audio HLS (нова фікстура ≥2 аудіо).

(S8 прибрано як окремий зріз — дублював per-slice контракт §8, crit ai-slop.)

## 6. Продуктові рішення власника (2026-11-07) — закриті

- **Відео-only:** урок не може бути без аудіо; Ready вимагає ≥1 Ready-доріжки й весь час Ready (D4).
- **Вирівнювання (ліміт):** автор вантажить повнометражне аудіо, вирівняне з початку; зсув/обрізання
  **і нерівну тривалість** (crit edge) базовий план НЕ обробляє — §7.
- **Порядок зі security-spike:** власник обрав гейтити старт **усієї** фази на **security S0.2
  cutover** (злиття s0.0-spike у main). Примітка критики (gaps #4, non-blocking): за логікою плану
  cutover строго потрібен **лише S5** (openapigen-авторинг); S1–S3, S-backfill, S6, S7 openapigen не
  чіпають. Рішення власника лишається (спільне робоче дерево з піром — додаткова причина); зафіксовано,
  щоб за бажання звузити.
- **Crossfade:** «мінімізувати gap» вбудовано (§4); справжній мікс-фейд виключено.

## 7. Ризики та ліміти

- Справжній crossfade на одному MSE неможливий (§4) — базовий план дає мінімальний gap (hls.js).
- Окремий дубляж без автосинхронізації: розсинхрон при невирівняному аудіо (зсув) **і при нерівній
  тривалості** (аудіо кінчається раніше/пізніше за відео) — не обробляється; S3 може додати валідацію
  тривалості з магнітудою допустимої дельти, або це лишається явним лімітом.
- Нові доріжки, додані під час активної VOD-сесії, стають видимі глядачу **лише в наступній сесії**
  (hls.js читає VOD-master один раз) — очікувано, не дефект.
- openapigen-поверхня спільна зі spike (лише S5).
- Старі ассети — закрито окремим S-backfill (не «гіпотеза в костюмі ліміту»).
- Збільшення storage/transcode на кожну доріжку.
- **Межа кількості доріжок (plan-defense Q4):** верхній ліміт доріжок на урок не форсується окремим
  квотуванням — обмежений лише унікальністю `(language, asset)` (D6), тобто фактично числом мов. Явний
  квот-ліміт **поза обсягом** базового плану (нема продуктової вимоги); маніфест і gear-меню коректні
  при 1..N доріжках, деградації при зростанні N не очікується в межах розумного числа мов.

## 8. Верифікація
Кожен зріз: відповідні юніт/repo/contract-тести → повний ordered `make check` на зрізах, що міняють
злиту/персистну поведінку → e2e на реальному multi-audio HLS → браузер для свапу → **незалежне рев'ю
кожного зрізу** (правило власника 26.10). Додатково (gaps #1): e2e, що **наявний upload→Ready
лишається зеленим** після зрізу, який вмикає Ready-ґейт.

## 9. Результат plan-critique (2026-11-07)
4/4 лінзи повернулись. **9 blocking** знахідок вбудовано: reuse замість клону машинерії (D3);
відео-only без `-map 0:a:0` разом з AcceptSource (D4/S3); regen master серіалізовано+атомарно+з-Ready-
набору (D7); рівно одна default + swap/re-election (D4); delete не нижче 1 Ready (D4); backfill окремим
зрізом (S-backfill); master лише Ready-доріжки (D7); активація Ready-ґейта разом із S4/S-backfill (S8-
нотатка+§8 e2e); fail-шлях стан-машини (D4). Non-blocking враховано: S8 прибрано, два Kind розділено,
job розширено а не додано, S3+S4 злито, нумерація S0 vs security-S0.2 розведена, D5 timing визначено,
ліміт тривалості додано.
**Відкритих blocking: 0 → план готовий до plan-defense.**

## 10. Rollback (plan-defense Q5)
Відкат — по-зрізово, бо кожен зріз зелений окремо (§5). Що оборотне й що ні:
- **Код (S1, S3, S5, S6, S7):** revert коміту/PR. Фаза не задеплоєна (test mode, рішення власника
  03.10) → прод-відкату немає; зворотність = git-revert гілки до мерджу.
- **Схема (S2 `audio_tracks`, S-backfill `000NN`):** goose-міграції з `DownTo`; down-міграція
  `audio_tracks` оборотна (drop table). **S-backfill — data-міграція**: down має прибирати лише
  згенеровані embedded-рядки (ідемпотентно, за ознакою `Kind:embedded`), не чіпаючи uploaded; довести
  up+down на populated data (прецедент enrollment 00015, D3/S-backfill).
- **НЕОБОРОТНИЙ зовнішній ефект — регенеровані `master.m3u8` на диску** (D7): revert коду НЕ
  переписує вже згенеровані файли; старий master для multi-track ассета лишиться з `EXT-X-MEDIA`, що
  вказує на `audio-<trackId>/`. Оскільки regen **ідемпотентний і повністю з БД** (D7), відновлення =
  прогнати стару (одно-доріжкову) логіку regen ще раз або лишити файли (вони валідні, доки каталоги
  доріжок на місці). Видалення каталогів `audio-<trackId>/` — ручний крок, не частина code-revert;
  зафіксувати в report зрізу S3 як known irreversible.
- **ASR (S6):** зміна default ре-ранить транскрипт (D5); відкат default повертає вхід, транскрипт
  ре-раниться назад — оборотно через той самий механізм.

## 11. Результат plan-defense (2026-11-07)
| # | Питання | Закрито рядком | Вердикт |
|---|---|---|---|
| 1 | malformed input | §3 D4 «Malformed input — завантажений аудіофайл»; §2 AcceptSource | closed |
| 2 | dependency failure | §3 D1 (reuse error-handler/retry), D4 (Failed-переходи), D7 (ідемпотентний ретрай) | closed |
| 3 | race / repeated | §3 D7 (per-asset лок, єдиний письменник, ідемпотентність), D4 (атомарний swap/re-election), §5 S2/S3 | closed |
| 4 | empty / boundary | §3 D4 (0→AwaitingAudio/Failed, 1→delete відмовляє), D6 (duplicates), §7 (нерівна тривалість; межа к-сті доріжок — out of scope) | closed |
| 5 | rollback | §10 Rollback (код/схема оборотні; regen master — known irreversible з механізмом відновлення) | closed |

**Відкритих gap: 0 → gate green. План переходить у розбивку на задачі.**

## Статус / блокер
- Продуктові рішення зібрані (§6). План **ЗАТВЕРДЖЕНО власником 2026-11-07**.
- S0 зроблено. plan-critique (0 open blocking) і plan-defense (0 gap) пройдено — обидві брами зелені.
- **РОЗБЛОКОВАНО 2026-11-07:** security S0.2 cutover ЗАВЕРШЕНО й у main — swap-коміт `93690df`
  (generated openapi.yaml authoritative), `openapi.generated.yaml` видалено, follow-ups PR #82
  (`9d42bb9`/`0445184`/`f8377b6`/`c426504`/`0ce4371`), закриття `169c0fc` (PR #81/#82). Гейт фази
  виконано. (Незакомічені `docs/security/_wip-s0.1*`/`_wip-s0.2-cutover.md` — лише робочі журнали піра,
  не код; код злитий.)
- **Наступний крок — S1 (домен)**, але НЕ в цій сесії: транскрипт >4МБ, потрібна свіжа CLI-сесія.

## Рішення власника
- **2026-11-07:** план ЗАТВЕРДЖЕНО («Го»), пройшовши plan-critique (0 blocking) і plan-defense (0 gap).
  Що саме закрито: обсяг §1, джерело аудіо §1, 7 архітектурних рішень D1–D7, розбивка S0–S7+backfill §5,
  ліміти §7, rollback §10. Старт кодування лишається гейтнутим на security S0.2 cutover (нижче).

## ДАЛІ
1. ✅ Затвердження власником — 2026-11-07.
2. ✅ security S0.2 cutover — завершено й у main (`93690df`+PR #82, закриття `169c0fc`).
3. **Старт S1 (домен) у СВІЖІЙ сесії** (транскрипт цієї >4МБ): `AudioTrack` + стан-машина (D3/D4),
   юніт-тести, без транспорту/БД/транскодера; далі leaf-first S2–S7+S-backfill, кожен зріз —
   незалежне рев'ю + ordered gate.
