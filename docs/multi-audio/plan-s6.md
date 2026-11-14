# Plan — multi-audio S6 (ASR бере вхід із default-доріжки)

Base `origin/main@ef48c15` (S-backfill злито, PR #104). Гілка `multi-audio-s6`. Writer = Courses 1,
review_mode = independent_subagent (Codex `exec` напряму). Вибір зрізу — власник, 2026-11-14 («S6»).
Джерело вимоги: `plan.md` D5 і §5 (S6): «вхід із default-доріжки; ASR стартує по призначенні default;
ре-ран при зміні default; fallback до backfill; тести, щоб наявний транскрипт не зламався».

## Проблема (виміряно на коді, 2026-11-14)
ASR-адаптер завжди читає `<mediaRoot>/source/<assetID>/original` (`transcript/adapters/media/audio.go:49`),
а черга й чернетка не знають про доріжки. Наслідки на `main` зараз:
1. **Video-only урок** (джерело без звуку, Ready з першою uploaded-доріжкою — S3/S5) → ffmpeg `-vn` з
   `original` не має аудіо → чернетка `failed/audio_extraction_failed`. Розпізнати такий урок неможливо.
2. **Урок із дубляжем як default** → розпізнається оригінал, а не те, що учень чує за замовчуванням.
3. Зміна default не перезапускає ASR: pending-вибірка дивиться лише на generation
   (`postgres/draft.go:79`), а default міняється без зміни generation.

Fallback «до backfill» більше не потрібен: S-backfill (PR #104) дав кожному Ready-ассету default-доріжку
(ґейт 00028), embedded має leaf `original` — той самий шлях, що й сьогодні.

## Рішення (2026-11-14, author Courses 1)

### D-S6-1 — Чернетка запам'ятовує leaf свого аудіо; міграція `00029`
`ALTER TABLE asr_drafts ADD COLUMN source_leaf text NULL` з `CHECK (source_leaf IS NULL OR
source_leaf = 'original' OR source_leaf ~ '^audio-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$')`
— та сама форма, що `audio_tracks.source_leaf` (00025: embedded → `original`, uploaded → `audio-<id>`).
`NULL` = чернетка до S6; вона зроблена з `original`, тож читається як `COALESCE(source_leaf,'original')` —
старі чернетки **не** перезапускаються масово (магнітуда: 0 зайвих ASR-прогонів на ассетах з embedded
default; перезапуск лише там, де default уже не `original`, — це і є дефект 1/2).
Альтернатива «зберігати `audio_track_id` з FK» відкинута: id embedded-доріжки змінився б при backfill/
повторній матеріалізації без зміни аудіо, а FK на чужу таблицю додає крос-модульний cascade; leaf рівно
називає байти, які чулися, і саме його потребує адаптер.

### D-S6-2 — Leaf визначається в `Claim`; автоматична задача не перезаписує вже актуальну чернетку
`Claim` (`postgres/draft.go:106`) читає `source_leaf` default Ready-доріжки ассета
(`audio_tracks WHERE video_asset_id=$asset AND is_default AND state='ready'`) і записує його в чернетку
разом зі `state='running'`; `ClaimedDraft` повертає leaf, `Process` передає його в `Extract`.
Нема default → `ErrNotFound` (як нема ассета) — «ASR стартує лише коли default призначено» (D5).
**Дублі автоматичних задач (critique #1 F1).** Послідовність: A розпізнається; default → B; poll ставить
заміну з marker = claim-час A; A завершується (marker змінився); наступний poll ставить **другу** заміну;
обидві згодом візьмуть B, і друга перезапише вже готову (можливо, виправлену) чернетку B. Механізм
відсіву: `RecognitionJob` отримує `Requested bool` (ручний Regenerate = `true`, pending = `false`;
payload `requested`, відсутнє поле старих задач → `false`). `Claim` автоматичної задачі приймає роботу
лише якщо чернетки нема, `generation` старіша або `COALESCE(source_leaf,'original')` ≠ поточного default
leaf; наявний guard монотонності generation (`WHERE asr_drafts.generation <= EXCLUDED.generation`)
лишається. Перервана спроба **не** відновлюється автоматичним claim: черга має `MaxRetry(0)`
(`adapters/asynq/queue.go:47`), asynq архівує задачу з простроченою орендою, а `Recover`
(`application/recognize.go:198`) переводить чернетку в `failed/timeout` без нового enqueue — так само, як
сьогодні; далі автор тисне Regenerate (critique #2 F6). Друга заміна
бачить чернетку `ready` з leaf B = default → `Claimed:false`, нічого не пише. Ручний Regenerate лишає
безумовний same-generation reclaim (наявний тест `testReclaimSameGeneration`,
`draft_integration_test.go:337`). Перевірка: детермінований integration-тест — дві автоматичні заміни
в черзі до/після завершення першої, друга не claim-иться, текст першої лишається.
Чому не з payload черги: задача, поставлена до зміни default, при виконанні бере **поточний** default —
застаріла задача не пише чернетку зі старого звуку. Читання з `audio_tracks` у transcript-адаптері —
той самий прецедент, що наявні читання `video_assets` (`postgres/draft.go:67,110,258`).

### D-S6-3 — Pending-вибірка перезапускає при зміні default
`PendingRecognitions` (`postgres/draft.go:65`) приєднує default Ready-доріжку поточного ассета
(INNER JOIN — без default ассет не кандидат) і бере ассет, якщо чернетки нема, generation інша **або**
`COALESCE(drafts.source_leaf,'original') <> tracks.source_leaf`. Ідентифікатор задачі
`asset:generation:marker` (`adapters/asynq/queue.go:151-160`), marker = `drafts.updated_at` — після
завершеної чернетки він новий, тож повторна задача для того самого generation не конфліктує з
утриманою (Retention 24 h). Виявлення зміни — на наступному poll (`pollInterval` = 60 с, `cmd/asr-worker/main.go:30`); початок
розпізнавання далі залежить від черги (concurrency 1, `cmd/asr-worker/main.go:72`; вибірка ≤50 за poll,
`application/recognize.go:19`) — межі часу старту без припущень про backlog не обіцяємо (critique #1 F3).
Альтернатива «подія зі зміни default у media → transcript» відкинута: потребувала б крос-модульного
порту/черги, а pending-опитування вже є і саме відновлюється після збою.

### D-S6-4 — `Extract(ctx, assetID, leaf)`; адаптер перевіряє leaf
Порт `AudioSource.Extract` отримує leaf; адаптер будує `source/<assetID>/<leaf>` лише після перевірки
leaf тим самим правилом, що CHECK (D-S6-1) — інакше `ErrInvalidRequest` (path-safety: leaf із БД, але
адаптер не довіряє). ffmpeg-аргументи (`-vn -ac 1 -ar 16000 pcm_s16le`) уже працюють і для
аудіо-only джерела (m4a/mp3/aac uploaded-доріжки).

### D-S6-6 — Інвалідація при зміні default: лише чернетка; опублікований транскрипт — ні
**Рішення власника 2026-11-14** (питання в сесії, варіант «Не чіпати»; закриває «позначати опублікований
транскрипт застарілим для учня»). Звужує D5 «зміна IsDefault інвалідує й ре-ранить транскрипт»:
- **Чернетка:** `stale` = generation чернетки ≠ ready generation **або** `COALESCE(source_leaf,'original')` ≠
  leaf поточної default-доріжки. Отже одразу після зміни default (ще до старту заміни, critique #1 F2)
  `GetDraft` віддає `stale=true`; текст банера (`LessonDraftField.tsx:121`) — «…зі старшого відео або
  іншої аудіодоріжки» (uk/en). `ports.LessonDraft` отримує leaf default-доріжки поряд із
  `ReadyGeneration`; `domain.Draft.Stale` порівнює обидва.
- **Доставка у відкритий редактор (critique #2 F5).** Зміна default не міняє generation, тож наявна
  інвалідація чернетки за generation (`admin-video-queries.ts:87`) не спрацює, а опитування чернетки
  зупиняється на термінальному стані (`admin-draft-queries.ts:114`). Тому: (1) `setDefault` і `remove`
  (видалення default переобирає іншу) у `admin-audio-queries.ts` після успіху інвалідують усі
  lesson-draft запити префіксом `['admin','draft']` (перечитуються лише змонтовані — редактор тримає
  чернетку уроку, який відкрито); те саме робить queryFn списку доріжок, коли змінюється термінальний
  підпис (worker міг призначити першу default на video-only ассеті); (2) `refetchInterval` чернетки
  опитує ще й поки `data.stale === true`, кожні `draftPollMs` = 5 с: кожна stale-чернетка є
  pending-кандидатом (D-S6-3), тож опитування спиняється, коли заміна дасть не-stale чернетку
  (ready/no_speech/failed з новим leaf). Це залежить від прогресу worker'а (critique #4 F8), тому вікно
  обмежене тим самим `draftTaskTimeoutMs` = 15 хв, що вже діє для ручного Regenerate
  (`admin-draft-queries.ts:8`), від першого спостереження stale (клієнтська мітка на lesson; знімається,
  коли чернетка не-stale). Магнітуда за недоступного worker'а: ≤180 GET на змонтовану чернетку, далі
  опитування стоп; банер stale лишається, Regenerate доступний. Текст у редакторі й токен-захист (`SaveDraft` за `updated_at`)
  не змінюються: ре-ран лише пропонує нову чернетку через наявний «Завантажити чернетку».
- **Перша чернетка уроку без чернетки — поза S6 (critique #3 F7; рішення 2026-11-14, Courses 1).** Урок,
  у якого чернетки ще нема, після першої Ready-генерації (будь-якого відео, і video-only з першою
  default теж) не отримує автоматичної доставки першої чернетки у відкритий редактор: відповідь без
  чернетки — `state=null, stale=false`, опитування не стартує, а інвалідація за generation пропускає
  перше спостереження (`admin-video-queries.ts:88`, `previousGeneration !== undefined`). Це **наявна**
  поведінка `main` для кожного першого відео уроку, не новий дефект S6. Механізм, що лишає автора не в
  глухому куті: редактор без чернетки показує «Згенерувати транскрипт» (`LessonDraftField.tsx:149`,
  ручний Regenerate), а запит перечитується на reload або на повернення у вкладку після `staleTime` 60 с
  (`router.tsx:36`, TanStack `refetchOnWindowFocus` за замовчуванням). Магнітуда: затримка показу до
  першого з цих подій; бекенд-розпізнавання при цьому вже йде. Закрита альтернатива — опитування
  чернетки, поки урок має Ready-відео без чернетки: потребує знання стану відео в draft-запиті й межі
  часу (worker недоступний → вічне опитування), окремий frontend-зріз. Інвалідацію draft-префіксу на
  зміну термінального підпису лишаю (дешевий refetch), але доставку першої чернетки нею не заявляю.
- **Опублікований транскрипт** (`transcripts`, позначка учневі `TranscriptPanel.tsx:53`) не змінюється й
  не стає stale: це текст автора; публікувати нову чернетку — його рішення. Магнітуда закритої
  альтернативи: колонка `authored_against_leaf` у `transcripts` + learner-читання + тести learner-позначки.
Перевірка: unit `Stale` (generation / leaf / legacy NULL); integration GetDraft stale після зміни default;
тест, що ре-ран не змінює рядок `transcripts` і його `stale`.

### D-S6-5 — Що НЕ змінюється (закриті альтернативи)
- **Опублікований транскрипт уроку не чіпається:** ре-ран пише лише `asr_drafts`; transcript
  (`transcripts`) міняється тільки явним Publish автора. Перевірка: тест «ре-ран не змінює transcript».
- **Відредагована, але не опублікована чернетка перезаписується ре-раном** (власне рішення 2026-11-14,
  Courses 1; не D8-15, що стосується каскадного видалення — critique #1 F4). Механізм той самий, що вже
  діє для нової generation і ручного Regenerate (`Claim` upsert скидає segments, `postgres/draft.go:124-134`).
  Зміна default — явна дія автора, D5 вимагає ре-рану, а редактор показує stale-банер до старту (D-S6-6).
  Обмеження в report.
- **Мова-підказка не береться з доріжки:** pending-задачі лишаються auto-detect (як сьогодні); мова
  доріжки — метадані автора без обмеження на підтримувані Whisper мови (`und`, `und-x-original`, `pt-BR`),
  хибна підказка зламала б розпізнавання; ручний Regenerate з вибором мови лишається.
- **Тайм-аут** лишається від тривалості відео (`TaskTimeout`); uploaded-аудіо іншої тривалості — наявне
  обмеження вирівнювання фази (`plan.md` §7).

## Тести
- **Unit (application):** `Process` передає в `Extract` leaf із claim; claim без default → нічого не
  розпізнається; pending-задача має `Requested=false`, Regenerate — `true`; `Stale` за generation/leaf/NULL.
- **Unit (adapter media):** leaf `original` / `audio-<uuid>` → правильний шлях; `../x`, `audio-`, порожній →
  `ErrInvalidRequest` без ffmpeg.
- **Integration (transcript postgres):** pending — (a) ассет з embedded default і чернеткою `NULL` leaf → не
  кандидат; (b) default → uploaded → кандидат; (c) чернетка з тим самим leaf → не кандидат; (d) ассет без
  default → не кандидат; Claim записує leaf default-доріжки; claim після зміни default бере новий leaf;
  автоматичний claim: дублікат після завершення → `Claimed:false`; ручний
  claim того ж generation/leaf → claim; GetDraft `stale=true` одразу після зміни default; ре-ран не змінює
  рядок `transcripts`. Міграція 00029 down/up.
- **Unit (asynq adapter):** payload round-trip `requested=true`, `false` і відсутнє поле старої задачі →
  `false` (critique #2 F6).
- **Frontend query/component:** закешована ready-чернетка A → `setDefault` → чернетку перечитано, банер
  stale видно → опитування триває, поки `stale` → `running` → `ready` (не-stale) без reload і без
  Regenerate; те саме для delete з переобранням; опитування спиняється на не-stale.
- **e2e (реальний whisper, production-шлях asr-worker):** (1) video-only урок + uploaded m4a з мовленням
  (flite) → чернетка `ready` з текстом (дефект 1) — перевірка через `page.reload()` / API-читання чернетки,
  бо доставка першої чернетки у відкритий редактор поза S6 (D-S6-6); (2) змонтований редактор: ready-чернетка з embedded →
  друга доріжка з мовленням → «Зробити основною» → stale-банер до claim → нова чернетка без reload
  (тайм-аут з урахуванням poll 60 с). Наявні C03/C06/C11/C13/C17/C18 лишаються зеленими.
- Повний ordered `make check`.

## Pre-code checkpoint
| задача | READY? | що бракує |
|---|---|---|
| S6 | **READY** (plan-critique #4 ACCEPTED, 2026-11-14) | — |

## Журнал
- 2026-11-14: plan-critique #1 (Codex) — BLOCKED: F1 дублі автоматичних задач → `Requested` + умовний
  автоматичний claim; F2 інвалідація D5 → питання власнику → «не чіпати» опублікований транскрипт,
  stale чернетки за leaf (D-S6-6); F3 межа 60 с → лише cadence виявлення; F4 хибне посилання D8-15 → власне
  датоване рішення. Raw: `_wip-s6-plan-critique-1.md`.
- 2026-11-14: plan-critique #2 (Codex) — BLOCKED: F1/F3/F4 resolved, F2 narrowed (бекенд-контракт); новий F5
  (blocking) доставка stale/ре-рану у змонтований редактор → інвалідація draft-префіксу + опитування
  поки stale + frontend/e2e тести; F6 (non-blocking) хибне «asynq-ретрай» → прибрано `running`-виняток,
  описано наявне Recover, payload round-trip тести. Raw: `_wip-s6-plan-critique-2.md`.
- 2026-11-14: plan-critique #3 (Codex) — BLOCKED: F5 narrowed (заміна термінальної чернетки resolved), F6
  resolved; новий F7 (blocking) перша чернетка не доходить до змонтованого редактора → звірено: та сама
  поведінка на `main` для кожного першого відео уроку → звужено з механізмом (Generate-кнопка, reload,
  refocus після staleTime 60 с) і магнітудою; e2e (1) читає результат через reload. Raw:
  `_wip-s6-plan-critique-3.md`.
- 2026-11-14: plan-critique #4 (Codex) — **ACCEPTED** на plan sha256 `daf694b8…79e4`, snapshot `ef48c15`; F7
  narrowed (виключення з механізмом), F1–F6 retained; F8 (non-blocking) опитування stale без межі →
  обмежено 15-хв вікном (як ручний Regenerate). Raw: `_wip-s6-plan-critique-4.md`. Implementation review
  і ordered gate — окремі gates.
