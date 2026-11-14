# Plan — multi-audio S5-L2 (UI «аудіодоріжки уроку»)

Base `origin/main@7c6b754` (S5-L1b fix-forward merged, PR #101). Writer = Courses 1,
review_mode = independent_subagent (Codex, `codex exec` напряму — не fork, не gate-reviewer subagent).
Surface map: `docs/multi-audio/_wip-s5l2-surface.md` (2026-11-13, звірено вибірково з кодом).

## Мета
Автор уроку (admin або instructor-власник курсу) у редакторі курсу бачить для кожного уроку з відео
список аудіодоріжок і може: вивантажити нову доріжку (файл + мова + назва), бачити стан обробки,
змінити мову/назву, зробити доріжку default, видалити. Backend-контракт уже злитий (L1a+L1b);
цей зріз — **лише frontend** (+ один косметичний operationId, D-L2-7).

## Поточний seam (факти, surface-map)
- Відео-UI живе тільки в `/admin/courses/$courseId` → `CourseEditorView` → `LessonVideoField` на кожен урок
  (`modules/media/ui/LessonVideoField.tsx`); `/teach` відео не має. Instructor проходить `/admin`-гард
  (`canAccessAdministration` = instructor|admin), ownership перевіряє backend (`courses.instructor_id`).
- tus-клієнт `modules/media/adapters/tus-client.ts` прибитий до відео: URL `/uploads/${videoId}` (`:45`),
  ліміт `maximumVideoBytes` (`:257`), fingerprint-ключ по videoId; `application/lesson-video.ts`
  `validTransferTarget` вимагає `uploadUrl === /uploads/${videoId}`.
- Orval-функції аудіо є, споживачів нуль: `listAdminAudioTracks`, `postApiV1AdminVideosVideoIdAudioTracks`
  (без operationId), `renameAdminAudioTrack`, `setDefaultAdminAudioTrack`, `deleteAdminAudioTrack`.
- Backend-правила, які UI має відзеркалити (а не відкривати як 400): default — лише Ready-доріжка
  (`ApplyDefault`); видалення останньої Ready на Ready-ассеті заборонено (`ValidateTrackDeletion`);
  409 `duplicate_language` / `upload_in_progress`; формат — AAC/MP3 (`AcceptAudioTrackSource`), ≤500 MiB;
  language 1..35 `[A-Za-z0-9-]`, label 1..64 без `"`/control.
- Опитування відео: 3 с, поки є queued/processing або трансфер у вкладці (`admin-video-queries.ts:36-44`).
- e2e: аудіо-фікстури немає; фікстури генеруються ffmpeg у `media.setup.ts`; серійний `media.spec.ts`
  уже має ready-відео в ланцюгу.

## Рішення (кожне закриває альтернативу; 2026-11-13, author Courses 1)

### D-L2-1 — Розміщення: секція під відео-полем уроку в `/admin/courses/$courseId`
Альтернатива `/teach` відкинута: там нема відео взагалі, інструктор і так редагує відео через `/admin`.
Секція рендериться лише коли в уроку є video asset (є `videoId`). Перевірка: component-тест «без відео — секції нема».

### D-L2-2 — tus-клієнт параметризується ціллю, а не дублюється (уточнено після critique F3)
`createTusClient.transfer` отримує ціль `{uploadUrl, maxBytes, chunkBytes, fingerprintKey}` замість
виводити її з videoId. Відео-шлях передає ті самі значення, що й зараз → поведінка відео без змін.
**Ціль виводиться з trackId і констант аудіо, а не з ticket-а**: `/uploads/audio/{trackId}`,
`maximumAudioBytes = 500 MiB`, `audioChunkBytes = 8 MiB` (=`mediaChunkBytes` backend). Ticket лише
перевіряється на збіг із цими значеннями (як `validTransferTarget` для відео). Отже, після reload
resume стартує **лише зі списку доріжок + localStorage**: fingerprint за trackId
(`courses.media.upload-fingerprint.v1.audio.<trackId>`) звіряє повторно вибраний файл, HEAD дає
`Upload-Length` і offset; `Upload-Length ≠ file.size` або відсутній/чужий fingerprint →
`upload_resume_unverified`. Cancel — DELETE на ту саму виведену ціль + видалення fingerprint.
Audio list DTO не має declared/received bytes — і не потребує: усе береться з HEAD (backend не міняємо).
Перевірка: tus/application-тест resume без ticket (лише trackId + storage), кейси missing/mismatched
fingerprint, cancel без ticket.
Альтернатива «копія клієнта для аудіо» відкинута: 288 рядків протоколу двічі. Перевірка: наявні 372 рядки
`tus-client.test.ts` зелені без зміни асертів + нові audio-кейси.

### D-L2-3 — Стан і опитування; синхронізація стану відео (уточнено після critique F2)
Окремий query `['admin','audio-tracks', videoId]`, refetch 3 с, поки є доріжка queued/processing або
трансфер у вкладці; інакше вимкнено. Інвалідація після кожної мутації.
Стан ассета змінює **worker** (перша Ready-доріжка переводить `awaiting_audio → ready`,
`postgres/audiotrack.go:419`), а не мутація, тому інвалідації після мутацій недостатньо. Механізм —
той самий, що вже є для lesson-draft (`admin-video-queries.ts:76-96`): queryFn аудіо-списку порівнює
попередній і новий стан кожної доріжки й на переході у термінальний (`ready`/`failed`) інвалідує
course-video list.
**Перше спостереження теж рахується (critique #2):** на відміну від draft-зразка, який пропускає
`previous === undefined` (`admin-video-queries.ts:87`), аудіо-queryFn рахує **термінальний підпис**
доріжок відео — відсортовані `trackId:state` для `ready`/`failed` — і інвалідує course-video list,
коли підпис ≠ останньому збереженому для цього videoId (`setQueryData` snapshot-ключа, як у draft);
відсутній snapshot + непорожній підпис = зміна. Тотожні відповіді не інвалідують (підпис той самий);
циклу немає, бо інвалідація course videos не зачіпає ключ аудіо. Магнітуда: ≤1 зайвий
`GET /admin/courses/{id}/videos` на відкриття редактора для кожного відео з термінальними доріжками.
Отже, після reload: перша відповідь аудіо вже `ready` → course videos перечитуються → `awaiting_audio`
змінюється на `ready`.
Плюс інвалідація course-video list після default/delete. Опитування відео на `awaiting_audio` **не**
вмикаємо (вічне опитування уроку, куди ще нічого не вантажили).
Перевірка: query-тести — (а) processing→ready інвалідує course videos; (б) snapshot відсутній, перша
відповідь `ready` → інвалідує; (в) те саме для першого `failed`; (г) повторна тотожна відповідь не
інвалідує; component/e2e: відео-статус оновлюється без ручного reload.

### D-L2-4 — UI запобігає відомим 400, а не лише показує їх (уточнено після critique F4)
«Зробити default» — тільки на Ready не-default. «Видалити» вимкнено (з поясненням), коли ассет
`ready`, ціль — Ready і іншої Ready-доріжки немає (дзеркало `ValidateTrackDeletion`, `domain/audiotrack.go:205`).
Помилки сервера все одно мапляться (гонки). Перевірка: component-тести на обидва вимкнення.

### D-L2-5 — Типи файлів: `.m4a`, `.mp3`, `.aac`; ліміт 500 MiB перевіряється до mint
Backend перевіряє кодек після вивантаження; клієнт ріже очевидне (розширення, розмір) до витрати трафіку.

### D-L2-6 — Розбивка на два PR
- **S5-L2a — data layer без UI:** domain (типи, ліміти, валідація language/label), port, application,
  adapter (zod-схеми + Orval + tus), query options, параметризація tus (D-L2-2). Unit-тести.
  Споживачів у UI нема → гейт-префікс за §8-уточненням (format/lint/typecheck/test/check-generated/build);
  але параметризація tus міняє злитий відео-шлях → **повний ordered `make check` обов'язковий і тут**.
- **S5-L2b — UI + e2e:** `LessonAudioTracksField`, wiring у route/`CourseEditorView`, i18n uk/en,
  component-тести, audio-фікстура в `media.setup.ts`, e2e у `media.spec.ts`; повний `make check`.
Кожен — незалежне Codex-рев'ю перед мерджем.

### D-L2-8 — `awaiting_audio` у фронтенді (critique F1; живий дефект на main із S3)
Backend пише `awaiting_audio` для відео без звуку (`postgres/publish.go:73`), а `videoStates`
(`modules/media/domain/video.ts:1-8`) його не знає → `courseVideosSchema` відкидає **весь** список
відео курсу (`lesson-video-repository.ts:108`) і редактор показує порожні рядки для всіх уроків.
L2a: стан у domain + schema + i18n uk/en (щоб список знову парсився — це фікс дефекту, з тестом
«список з awaiting_audio парситься»). L2b: `shownVideo` вибирає `awaiting_audio` як поточне відео
уроку; відео-поле пояснює «додайте аудіодоріжку»; секція аудіо показується. e2e: video-only фікстура
(testsrc2 без аудіо) → `awaiting_audio` → upload m4a → доріжка ready → відео ready без reload.

### D-L2-9 — learner-читання мапить `awaiting_audio` у `processing` (розширення обсягу L2a, 2026-11-13, Courses 1)
Під час коду L2a знайдено той самий клас дефекту в backend: `course/adapters/postgres/learning.go:35-37`
мапив лише `created/uploading/queued/processing` у `processing`, а інше віддавав сирим, тож
`awaiting_audio` виходив на wire поза контрактом `LearningVideoStates` (`processing|ready|failed`) і
поза фронтенд-схемою `playback-schema.ts:42` → сторінка курсу учня для такого уроку не парситься.
Фікс — додати `awaiting_audio` у той самий `IN (...)`: для учня відео без звуку ще не готове, як і
`processing`. Закриває альтернативу «новий learner-стан» (зміна контракту без продуктової вимоги).
Перевірка: integration-підтест у `learning_integration_test.go` (asset → `awaiting_audio` → `VideoProcessing`).
Поза обсягом (обмеження): `adminstats` `in_progress` (`adminstats/adapters/postgres/snapshot.go:137`)
не рахує `awaiting_audio` — лічильник, не wire-парсинг; відео без звуку зараз не потрапляє в
`in_progress` і не в `ready` дашборду.

### D-L2-10 — рішення під час коду L2b (2026-11-13, Courses 1)
- **Секція аудіо — лише для поточного відео в `ready` або `awaiting_audio`.** Закриває «показувати й
  під час `queued/processing`»: доріжки з'являються після транскоду (embedded створює worker), а
  додавання до необробленого ассета не має продуктової потреби. Перевірка: component-тест D-L2-1
  (processing → секції нема).
- **`shownVideo`: `awaiting_audio` є поточним, лише якщо його generation > найновішого `ready`.**
  Закриває «завжди вище за ready» і «завжди нижче»: новіша заміна без звуку мусить бути видимою, щоб
  автор додав аудіо; старша застрягла генерація (поза pending-index, див. `_gotchas`) не ховає
  ready-відео. Перевірка: два component-тести D-L2-8 + мутація `>`→`<` (2 падіння).
- **Бейдж доріжки — `data-audio-state`, не `data-video-state`.** Закриває спільний атрибут: наявні
  e2e-локатори `field.locator('[data-video-state]')` стали б неоднозначними (strict mode) на кожному
  ready-уроці з embedded-доріжкою. Перевірка: повний e2e media-проєкт.
- **Заміна, що впала через аудіо, ховається за старшим `ready` — як будь-яка впала заміна (D-12).**
  Відповідь на Codex-рев'ю L2b (Major #1), 2026-11-13. Механізм: коли падає остання доріжка
  `awaiting_audio`-ассета, backend ставить ассету `failed/render_failed` (`postgres/audiotrack.go:519`),
  а worker підвищує стан лише з `awaiting_audio` (`:421`) — аудіо такий ассет уже не відновить; шлях
  відновлення — «Замінити відео» (видно на ready-уроці). Закриває «показувати failed-заміну з
  аудіо-контролами»: вони пропонували б дію без ефекту. Обмеження: автор не бачить причини падіння
  заміни, поки є старше ready — те саме, що для відео-заміни на `main`. Перевірка: тест «D-12 badge order».
- **Розширення файлу перевіряється до mint** (Codex Major #2, прийнято): `acceptedAudioFileName` у
  application `mint` → `unsupported_media` без запиту; drag-and-drop обходить фільтр пікера.
  Перевірка: domain-тест + adapter-тест «refuses a non-audio file name before any request».
- **Мова в Select не пропонує зайнятих мов** (не-failed доріжки того ж ассета, міграція 00026);
  409 `duplicate_language` все одно мапиться (гонка). Перевірка: component-тести + domain-тест.

### D-L2-7 — operationId `mintAdminAudioTicket` для POST create
Overlay-рядок + `make generate`; дзеркало `mintAdminVideoTicket`. Генерована назва
`postApiV1AdminVideosVideoIdAudioTracks` інакше протече в adapter. Зміна — лише ім'я в generated client.

## Рішення власника (2026-11-13)
- **Q2 → resume як у відео** (власник, 2026-11-13; закриває «лише скасувати»). Fingerprint у localStorage
  за trackId, повторний вибір того самого файлу, HEAD-offset; перервана доріжка показує resume + cancel.
  Перевірка: component-тест interrupted-стану + tus-тест resume з audio-ціллю.
- **Q1 → Select, не вільне поле** (власник, 2026-11-13). Перелік: `uk, en, pl, de, fr, es, und`.
  `und` обов'язковий у будь-якому разі: embedded-доріжка створюється з `language="und"`
  (`application/transcode.go:471`), і автор має її перейменувати. Доріжка з мовою поза переліком
  (дані, внесені до появи списку) показується як є, а Select пропонує її поточне значення.
- **Q1b → константа зараз** (власник «Approve» на рекомендацію, 2026-11-13; закриває адмін-словник у
  межах S5-L2). Перелік живе однією константою в `modules/media/domain`; словник — окремий зріз, якщо
  перелік почне часто мінятися. Перевірка: domain-тест, що перелік містить `und` і всі коди проходять
  backend-правило мови. Контекст питання: власник запитав, чи не простіше словник, який налаштовує адмін.
  Магнітуда: словник = міграція таблиці + repo + 2–4 admin-ендпоінти (openapigen) + admin-екран CRUD
  + e2e — окремий backend+frontend зріз, більший за весь S5-L2. Константа в `modules/media/domain` —
  ~10 рядків, зміна переліку = PR. Рекомендація: константа зараз, словник — окремий зріз, якщо знадобиться.

## Тести
- L2a: unit domain (валідація, перелік мов), application (ids/inputs до порту), adapter (zod на всі
  статуси, wireError-мапа, tus із audio-ціллю: URL, ліміт 500 MiB, chunk), resume без ticket
  (D-L2-2), список відео з `awaiting_audio` парситься (D-L2-8), перехід доріжки в термінальний стан
  інвалідує course videos (D-L2-3); відео-тести tus без змін асертів.
- L2b: component (список/стани, upload-флоу, interrupted → resume/cancel зі старту без ticket, rename,
  default-gating, delete-gating з умовою стану ассета + confirm, помилки 409, `awaiting_audio` у
  відео-полі), e2e: (1) video-only → `awaiting_audio` → upload m4a → відео ready без reload;
  (2) ready-відео → upload m4a «en» → ready → default → delete; master.m3u8 містить 2 аудіо після
  ready; axe на секції, keyboard-навігація.

## Pre-code checkpoint (per-task)
| задача | READY? | що бракує |
|---|---|---|
| S5-L2a | **READY** (critique #3 ACCEPTED) | — |
| S5-L2b | **READY** (L2a злито `15afa6c`, PR #102, 2026-11-13) | — |
Залежності: L2b ← L2a злито. Impact L2a: tus-клієнт відео (D-L2-2) → повний e2e відео-upload.

## Ризики
- Параметризація tus ламає відео-resume (fingerprint-ключ) — ловиться наявними тестами + e2e media.
- e2e аудіо подовжує media-проєкт (транскод 1–2 с аудіо — мала магнітуда; фікстура `sine` 2 с).
- Спільне робоче дерево з піром — гілка `multi-audio-s5l2a` від `origin/main`, явні шляхи в `git add`.

## Журнал
- 2026-11-13: plan-critique #3 (Codex, 4/4 лінзи, delta D-L2-3) — **ACCEPTED** S5-L2a і S5-L2b на plan sha256 `fe514c1c…0ace`, snapshot `7c6b754`; 0 blocking. Raw: `_wip-s5l2-plan-critique-3.md`. Implementation review і ordered gates лишаються окремими gates.
- 2026-11-13: plan-critique #2 (Codex, 4/4) — BLOCKED: F1/F3/F4 resolved; F2 звужено до першого термінального спостереження без snapshot → внесено «термінальний підпис» у D-L2-3 + тести (б)–(г). Raw: `_wip-s5l2-plan-critique-2.md`.
- 2026-11-13: plan-critique #1 (Codex, 4/4 лінзи) — BLOCKED: F1 awaiting_audio, F2 синхронізація стану відео, F3 resume без ticket (blocking), F4 delete-guard (non-blocking); усі звірено з кодом і внесено (D-L2-8, D-L2-3, D-L2-2, D-L2-4). Raw: `_wip-s5l2-plan-critique-1.md`. Потрібна повторна критика зачепленого scope.
- 2026-11-13: surface map (окрема сесія, read-only), план написано. Власник: Q2=resume, Q1=Select (uk,en,pl,de,fr,es,und); Q1b=константа (Approve).
