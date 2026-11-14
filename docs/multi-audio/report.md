# Report — multi-audio (кілька аудіо-доріжок на урок)

Фаза поза роадмапом §24. План і обидві брами — `docs/multi-audio/plan.md` (ЗАТВЕРДЖЕНО
власником 2026-11-07, plan-critique 4/4 0 blocking, plan-defense 5/5 0 gap).

Гілка `multi-audio` від `main@69d51a8`. Зрізи leaf-first (§5 плану), кожен зелений окремо,
кожен — незалежне рев'ю + ordered gate.

## S1 — домен (2026-11-07)

**Що зроблено.** Доменна модель multi-audio, без транспорту/БД/транскодера:

- `internal/media/domain/audiotrack.go` — сутність `AudioTrack` із двома життєвими циклами (D3):
  - `AudioTrackKind` = {`embedded`, `uploaded`}; `NewUploadedTrack` (старт `Created`, ніколи не
    default до Ready, leaf `audio-<trackID>`) і `NewEmbeddedTrack` (похідний `Ready`, leaf
    `original`, default дозволено — це аудіо, що веде ассет у Ready).
  - `ValidateAudioTrackState` — embedded лише `Ready`; uploaded повна дуга мінус `AwaitingAudio`.
  - `AcceptAudioTrackSource` — приймає лише AAC/MP3; порожній/непідтримуваний кодек =
    business-помилка (правило malformed-upload D4/Q1); video-only релаксація на нього НЕ поширюється.
  - Інваріанти D4 чистими функціями: `AssetMayBeReady` (≥1 Ready), `ValidateDefaultInvariant`
    (≤1 default до Ready; рівно 1 Ready-default у Ready), `ApplyDefault` (атомарний swap, ціль має
    бути Ready, вхід не мутується), `ElectDefault` (перша Ready за (CreatedAt, ID)),
    `ValidateTrackDeletion` (відмова прибрати останню Ready-доріжку у Ready-ассеті; missing → not-found).
- `internal/media/domain/asset.go` — новий стан `VideoAsset` `AwaitingAudio` + переходи (D4):
  T10 `Processing→AwaitingAudio`, T11 `AwaitingAudio→Ready`, T12 `AwaitingAudio→Failed`, усі
  пишуться Worker. Прямий T5 `Processing→Ready` (embedded) збережено. `States` розширено.
- Юніт-тести: `audiotrack_test.go` (конструктори, lifecycle, acceptance, усі інваріанти з
  межами й tie-break), `asset_test.go` (матриця переходів × 8 станів, writers T10–T12).

**Рішення зрізу (закривають альтернативу).**
- **Релаксація video-`AcceptSource` ВІДКЛАДЕНА в S3** — за правилом порядку плану (§3 D4: «мердж
  релаксації заборонено раніше за зміну транскодера»). Live `AcceptSource` і call-site
  `application/transcode.go:102` — не чіпано. S1 додає лише `AcceptAudioTrackSource`.
- **`AwaitingAudio` входить у публічний `States`** → адитивна зміна згенерованого контракту:
  `backend/openapi/openapi.yaml` і `frontend/src/adapters/api/generated/models/mediaVideoState.ts`
  отримали `awaiting_audio` (sqlc без змін). DB CHECK розшириться в S2 — безпечно, бо жоден код не
  пише цей стан до S3.

**Незалежне рев'ю (review_mode=independent_subagent, 2026-11-07).** Fresh-context субагент,
read-only по дифу. **VERDICT PASS**, 0 CONFIRMED-дефектів. Три minor-знахідки закрито в тому ж
зрізі: (1) `ValidateDefaultInvariant` — незалежний `readyDefaults` замість перезаписуваного
прапора; (2) коментар про навмисно невільний `IsDefault` у `NewEmbeddedTrack`; (3) покриття
uploaded `Uploading`/`Queued` у `ValidateAudioTrackState`.

**Верифікація.**
- `gofmt -l` чисто; `go vet ./internal/media/domain` ок; `go test -race ./internal/media/domain` PASS.
- Ordered-префікс `make format-check lint typecheck test check-generated` на цілому дереві:
  format/lint/typecheck/test ✅ (frontend enum-exhaustiveness пройшов із новим значенням;
  `go test -race ./...`). `check-generated` зелений після staging згенерованих файлів.
- `make build` (docker compose build) ✅.
- **Повний `make integration` + `make e2e` ВІДКЛАДЕНО до S2** — рішення власника 2026-11-07:
  S1 домен-only, нове enum-значення адитивне й без споживачів; усі out-of-diff вектори покриті
  typecheck+build+test; e2e не може дійти до нового стану. Це **deferred, не passed** — повний
  ordered gate обов'язковий на S2 (persist), першому зрізі зі злитою/персистною поведінкою.

## S2 — persist (2026-11-08)

**Що зроблено.** Персист-шар multi-audio, без app/HTTP/транскодера:

- `migrations/00025_multi_audio.sql` — ALTER `video_assets` state-CHECK (+`awaiting_audio`, up/down
  симетрично до набору 00005) + `CREATE TABLE audio_tracks` з DB-рівневими інваріантами D3/D4:
  FK `ON DELETE CASCADE`, `embedded⇒ready`, `leaf↔kind` (`original` / `audio-<id>`), `default⇒ready`,
  partial unique `audio_tracks_one_default_per_asset` (≤1 default/asset).
- `internal/media/adapters/postgres/audiotrack.go` — raw-tx repo-методи (патерн Mint/Claim):
  `CreateAudioTrack` (app-minted id + RETURNING), `ListAudioTracks` (порядок (created_at,id)),
  `SetDefaultAudioTrack` (двоетапний clear→set swap під `FOR UPDATE`), `DeleteAudioTrack`
  (re-election через `domain.ElectDefault`, guard останньої Ready через `ValidateTrackDeletion`);
  `mapAudioTrack`; `mapAsset` розширено на `AwaitingAudio`.
- Integration-тести (Postgres): персист+порядок, schema-guards (включно з прийняттям `awaiting_audio`
  на `video_assets`), set-default swap +adverse-order +concurrent, delete+re-election+last-ready-guard,
  create+orphan, міграція 00025 down/up round-trip.

**Незалежне рев'ю (independent_subagent, 2026-11-08).** VERDICT **PASS-WITH-CONCERNS**.
- #1 [PLAUSIBLE-high] — single-statement swap `is_default=(id=$2)` під НЕ-deferrable partial unique
  індексом міг дати order-залежний transient 23505 (rank-swap). **Закрито**: двоетапний clear→set
  swap + `testSetDefaultAdverseOrder` (своп на раніше-створену доріжку, поки пізніша тримає default).
- #2 [GAP] відсутній concurrency-тест — **закрито** `testConcurrentSetDefault` (дві паралельні SetDefault,
  інваріант «рівно один default» тримається; подія-керований, без sleep).
- #3/#4 — не блокери: reversibility CHECK-симетрія доведена кодом (down test-only, §10); evicted-guard
  на Create/SetDefault — app-політика наступного зрізу (DeleteAudioTrack уже перевіряє `evicted_at IS NULL`).

**Верифікація.** Повний ordered `make check` **зелений**: `CHECK_EXIT=0`, e2e **129 passed** (17.6m),
+ integration (media Postgres, усі S2-підтести), check-generated (sqlc без дрейфу — нові методи raw-tx,
без нових queries), build. Це перший зріз зі злитою/персистною поведінкою — повний gate обов'язковий і
виконаний (на відміну від S1, де e2e був deferred).

**Наступний крок — S3 (транскод доріжки + regen master):** розширити payload `media:transcode`;
ffmpeg пакетує аудіофайл; повна регенерація master під per-asset локом, атомарно; матеріалізація
embedded; відео-only гілка без `-map 0:a:0` разом із релаксацією video-`AcceptSource` (той самий зріз).
Передумова S0-борг: захопити реальний master.m3u8. Повний ordered `make check`.

## S3 — транскод доріжки + regen master (2026-11-08)

**Що зроблено.** Транскодер, пакування аудіо й повна регенерація master; коміти:
- `df9a313` **S3a** — video-only render argv-гілка (`Render(…, hasAudio)`), `RenderAudioTrack`,
  чиста `rewriteMasterText` / атомарний `RegenerateMaster` (D7, single-writer під per-asset локом),
  релаксація `AcceptSource` (scoped: `AudioCodec==""||aac||mp3`; непідтримуваний кодек досі reject),
  `validLabel`/`validLanguage` manifest-safe (відмова `"`, CR/LF, 0x7f). Golden-master'и
  `testdata/master_*.m3u8` реальним ffmpeg — S0-борг (захопити справжній master.m3u8) закрито.
- `c62f662` **S3b-1** — embedded-матеріалізація + AwaitingAudio: `Publish(embedded *AudioTrack)`
  (атомарно Ready+track, або AwaitingAudio без eviction-проходу — за D4); координатор
  `repo.RegenerateMaster`.
- `f838882` **S3b-2** — uploaded-track transcode-гілка: payload `trackId`, `Job.TrackID`,
  `EnqueueAudioTrack`; `processAudioTrack`; storage audio-leaf; `LoadAudioTrack` /
  `PromoteAudioTrack` (T11 + default-election + regen під локом) / `FailAudioTrack` (T12).
- `133a20f` **gate-fix (test-only)** — integration-seed AwaitingAudio-ассета несе renditions
  (інакше promote→Ready валив `video_assets_ready_has_renditions`, 23514) + integration-only фейк
  `tusQueue.EnqueueAudioTrack`. Без змін продакшн-коду.

**Рішення зрізу (закривають альтернативу).**
- **Lean lifecycle для uploaded-track (рішення власника 2026-11-08).** Uploaded-доріжка йде
  `Created→(upload)→Queued→processAudioTrack→Promote/Fail` без проміжного `Uploading`-підтвердження
  tus-callback'ом: мінтимо id, приймаємо байти, ставимо в чергу. Закрита альтернатива — повний
  upload-handshake як у video — відкладена, бо uploaded-track UI активується лише в S5; механізм
  доведення: промоут/фейл під per-asset локом із integration-тестами (`testPromoteAudioTrack*`,
  `testFailAudioTrack*`, `testConcurrentPromoteAudioTrack`).
- **S0-факти master-регенерації (виміряні реальним ffmpeg, 2026-11-08):** аудіо-група master'а має
  `GROUP-ID="group_aac"`; `EXT-X-STREAM-INF` несе виміряний `BANDWIDTH` (не оцінку); embedded-доріжка
  адресується відносним URI під `audio/` (а uploaded — `audio-<trackID>/`). Ці значення зафіксовані
  в golden-master'ах `testdata/master_*.m3u8`.

**Незалежне рев'ю (independent_subagent, 2026-11-08; повний звіт — `docs/multi-audio/_wip-s3-review.md`).**
VERDICT **PASS (S3 scope).** Live-шлях S3 — embedded-матеріалізація, video-only→AwaitingAudio,
атомарність/вибір-стану/eviction `Publish`, повний rewrite master (атомарний, ідемпотентний, video-рядки
verbatim), path-safety audio-leaf, quoted-string-safety, рівно-один-default — коректний; інваріант D4
«no silent playback» тримається. Три CONFIRMED-знахідки — у uploaded-track шляху, що **не має
продакшн-виклику в S3** (активується в S5): недосяжні в S3-as-shipped, блокувальні для S5.

**Обмеження (§S3, усі латентні до S5 — uploaded-track шлях без продакшн-виклика).**
1. [severe] `processAudioTrack` класифікує ВСІ probe/`RenderAudioTrack`-фейли як transient → після
   `MaxRetry` asynq кидає задачу, доріжка лишається `Queued`, ассет навіки `AwaitingAudio` (T12
   `FailAudioTrack` не спрацьовує, бо `Queued` рахується як pending). Video-шлях натомість робить
   термінальний `Fail` на вичерпанні. **Фікс до S5:** термінальний `FailAudioTrack` при
   `RetryCount>=MaxRetry`.
2. `PromoteAudioTrack` піднімає `AwaitingAudio→Ready`, але НЕ ганяє per-lesson eviction-прохід
   (на відміну від `Publish`) → video-only уроки з повторними аплоадами накопичують Ready-генерації
   понад cap=2; старі байти не чистяться. Суперечить D4 «eviction при переході ассета в Ready».
   **Фікс до S5.**
3. `DeleteAudioTrack` (unwired, S5) міняє БД, але не кличе `RegenerateMaster` → видалення Ready-
   доріжки з multi-track Ready-ассета лишає stale `EXT-X-MEDIA` URI на видалений каталог.
   **Фікс при wiring S5.**
- Minor (4): самотній embedded-Ready ассет тримає ffmpeg-ів master verbatim (`NAME="audio_0"`, без
  LANGUAGE/AUTOSELECT), поки DB-рядок — `Label="Original"`/`Language="und"`; сходяться до «Original»/«und»
  на першому regen (коли впаде uploaded-доріжка). Косметика.
- By-design (5): video-only аплоад тепер ПРИЙМАЄТЬСЯ (був reject до S3) і йде в `AwaitingAudio` без
  forward-шляху до S5 — за D4-sequencing, але це жива зміна поведінки вже зараз.

**Верифікація.** Повний ordered gate **зелений** (2026-11-08):
- `make integration` — 64 пакети `ok`, 0 FAIL; media Postgres (усі S3-підтести, включно з
  `testPromoteAudioTrackReadiesVideoOnlyAsset` і `testConcurrentPromoteAudioTrack`, раніше червоними).
- `make check-generated` — без дрейфу (S3 raw-tx методи, без нових sqlc-queries).
- `make build` ✅.
- `make e2e-local` (reset-volumes + reseed) — **129 passed (18.0m)**, exit 0. e2e поки покриває
  embedded/video-шлях; upload→Ready (uploaded-track UI) — S5; лічба 129 стабільна від S2.
- format-check (gofmt чисто на змінених файлах; фронт не чіпано), lint/typecheck/test не зачеплені
  зміною (integration-tagged Go test-онлі).

**Наступний крок — S5 (wiring uploaded-track) з попереднім закриттям обмежень 1–3** або S-backfill/
S6/S7 за `plan.md` §5. Мердж гілки `multi-audio` — через PR (код-фаза), після незалежного рев'ю
дельти й go власника.

---

## S5-L1b fix-forward (2026-11-10) — закриття знахідок незалежного рев'ю PR #100

PR #100 (S5-L1b: upload+enqueue+sweep для audio-доріжок) злито 2026-10/11 під **fork**-рев'ю, яке
повернуло CLEAN, бо успадкувало контекст автора. Запущене згодом **незалежне** рев'ю (Codex, свіжий
контекст) повернуло **BLOCK** із 4 Major. Усі звірені з кодом і закриті у `547b1f3`+HLS-fix
(гілка `multi-audio-s5l1b-fixes`):

- **F1 небезпечний sweep.** Стара `DeleteAudioTrackSwept` видаляла за id без фільтра стану → sweep
  міг почистити/видалити доріжку, яку конкурентний PATCH щойно підняв у Queued. Замінено на
  `DeleteSweptAudioTrackLocked`: під `FOR UPDATE` перевіряє eligibility тим самим предикатом, що й
  вибірка, чистить сховище й видаляє в одній транзакції; промоутнутий рядок → `ErrInvalidStateTransition`,
  пропуск без purge живих байтів.
- **F2 втрата enqueue.** Доріжка, що закомітила Queued, але чий post-commit `EnqueueAudioTrack`
  загубився (збій Redis / крах), застрягала. Додано `RecoverQueuedAudioTracks` (queued > 10 хв) і
  audio-re-enqueue у `Transcode.Recover`; asynq дедупить за task id.
- **F3 витік байтів.** Cancel/mint-eviction ставлять `failed`, а sweep брав лише created/uploading.
  Розширено на незавершені failed (received<declared) і завершені-та-failed понад 30-денне вікно
  (дзеркало video failed-reap). Published HLS failed-after-publish доріжки теж чиститься (`audioHLSKey`).
- **Minor.** Опис `MediaBadRequest` тепер називає межу на медіа (2 GiB video, 500 MiB audio).

**Рішення (2026-11-10, артефакт):** uploaded-доріжка у `failed` із `received<declared` (cancel /
mint-eviction — незавершене вивантаження) — upload-junk; sweep чистить і видаляє рядок. `failed` із
`received=declared` (transcode-fail, видимий у ListAudioTracks) лишається, реапиться після 30 днів.
Перевірка: предикат `audioSweepEligible`; integration-тест на обидві гілки. Дивергенція від video
(video має evicted_at) — audio розрізняє незавершене за received<declared.

**Обмеження (магнітуда/механізм за AGENTS.md):**
1. `Recover` виконується **лише один раз при старті** media-worker (`cmd/media-worker/main.go:61`),
   не періодично — ідентично video-шляху. Загублений enqueue, що стався, поки worker працює, не
   відвідується до наступного рестарту (і лише якщо рядок старший за 10-хв grace). `Sweep` натомість
   НЕ стартовий: крім старту він біжить на початку кожного video-`Process` (`transcode.go:47`), тож
   F3-cleanup частий. Періодична **recovery** — покращення всього worker-а (стосується й video),
   поза обсягом цього audio-фіксу.

**Верифікація.** `format-check` / `lint` / `typecheck` / `test` / `check-generated` / `build` /
`integration` — зелені (2026-11-10). Нові тести: `audioupload_test` (7: head-gating, offset/over-declared,
completion→enqueue, enqueue-fail, cancel), `transcode` audio sweep/recover (3), `tus/audioserver_test`
(7: HEAD/PATCH/DELETE/offset/415/405/412/404), HTTP mint-контракт AT09–AT12, postgres integration
(sweep-вибірка, locked-delete purge+promote-skip, recover-вибірка). Повний упорядкований `make check` на `1e4eeb2` (2026-11-13) — `GATE_EXIT=0`, e2e **129 passed**, 0 skipped (15.8 хв). Попередній червоний e2e (Acc26–31 TOTP, `login→503`) був інфра-шумом: argon2id-timeout під CPU-контенцією від сторонніх контейнерів на хості (load 55–116); після їх зупинки (load ~8) той самий тест `Acc26–Acc29 and Acc31` зелений (2.4 хв) без змін у коді.

**Відома test-robustness follow-up (non-blocking, SHIP-вердикт):** F1 integration-тест
`testDeleteSweptAudioTrackLockedPurgesAndSkipsPromoted` доводить skip промоутнутого треку за
eligibility-предикатом, але не ганяє СПРАВЖНЮ конкурентну транзакцію (PATCH, що промоутить у Queued,
проти sweep). Тож мутація «прибрати `FOR UPDATE`» тест би не завалила. Production-код коректно тримає
row-lock (підтверджено незалежним рев'ю). Підсилення — детермінований two-transaction race за
патерном `blockingMintTracer`/`testConcurrentPromoteAudioTrack` — залишено follow-up, щоб не вносити
флак-ризик наприкінці.

## S5-L2a — data layer аудіодоріжок + фікс `awaiting_audio` (2026-11-13)

План: `docs/multi-audio/plan-s5l2.md` (plan-critique Codex, 3 раунди → ACCEPTED). Гілка `multi-audio-s5l2a`.

- **Дефект на main (з S3), закрито:** фронтенд `videoStates` не мав `awaiting_audio`, тож strict-схема
  відкидала **весь** список відео курсу через один video-only рядок — редактор без відео для всіх уроків.
  Learner-SQL (`course/adapters/postgres/learning.go`) віддавав `awaiting_audio` сирим поза
  `LearningVideoStates` і playback-схемою; тепер це `processing` (D-L2-9).
- **tus:** `createTusUploads(resource, …)` — шлях, ключ fingerprint і ліміти задає ресурс; відео-клієнт
  зв'язаний із тими самими значеннями, відео-tus-тести не змінено.
- **Аудіо-шар:** domain (стани, ліміт 500 MiB, перелік мов-константа, правила language/label як на сервері,
  guards default/delete, термінальний підпис), port, application, Orval+zod adapter (resume лише з trackId +
  fingerprint), Query options; operationId `mintAdminAudioTicket`. Споживачів у UI ще нема (S5-L2b).

**Верифікація.** Повний упорядкований `make check` 2026-11-13 — `GATE_EXIT=0`: frontend unit 884 passed,
e2e **129 passed** (15.7 хв). Нові тести: domain 21, repository (msw, включно з tus на `/uploads/audio/`) 15,
queries 6, schema +1, learner integration +1. Мутаційні перевірки: (1) правило «перше термінальне
спостереження інвалідує список відео» → прибране, падає 3/6 query-тестів; (2) learner-SQL без
`awaiting_audio` → падає підтест `video-only asset awaiting its audio track reads as processing`
(прицільний `-v` прогін у runner-образі integration; у повному гейті integration-лог не verbose).

**Незалежне рев'ю (Codex, свіжий контекст, 2026-11-13):** SHIP на `ea54313` з двома Minor (audio-resume з
розбіжним `Upload-Length` давав `invalid_request` замість `upload_resume_unverified`; tus 409
`upload_offset_conflict` губився як `invalid_response`) — обидва виправлено в `b4ee658` лише для аудіо;
delta-рев'ю — SHIP на `b4ee658`. Повторний повний `make check` на `b4ee658` — `GATE_EXIT=0`, frontend unit
885, e2e **129 passed** (16.1 хв). Raw: `docs/multi-audio/_wip-s5l2a-review-codex.md`.

**Обмеження.** `adminstats` `snapshot.go:137` не рахує `awaiting_audio` ні в `in_progress`, ні в `ready`
(лічильник дашборду, не wire-парсинг) — поза обсягом. e2e-шлях video-only → аудіо → ready — у S5-L2b.

## S5-L2b — UI аудіодоріжок уроку + e2e (2026-11-14)

План: `docs/multi-audio/plan-s5l2.md` (D-L2-1…D-L2-5, D-L2-8, D-L2-10). Гілка `multi-audio-s5l2b` від `main@15afa6c`.

- **Секція «Аудіодоріжки»** у `LessonVideoField` для поточного відео в `ready`/`awaiting_audio`
  (`modules/media/ui/LessonAudioTracksField.tsx`): список зі станами, нова доріжка (Select мови з
  константи без зайнятих мов + назва), resume/cancel перерваної доріжки лише зі списку, rename у
  діалозі, «Зробити основною» лише на Ready не-default, видалення з підтвердженням і заблоковане для
  останньої Ready-доріжки Ready-ассета; помилки сервера (409 тощо) мапляться в названі тексти uk/en.
- **`shownVideo`:** `awaiting_audio` — поточне відео, лише коли новіше за найновіше `ready` (D-L2-10).
- **Wiring:** `useCourseEditorAudio` у `routes/admin.courses.$courseId.tsx` — один query на ассет,
  лічильник трансферів per video для опитування.
- **Розширення файлу** перевіряється в application `mint` до будь-якого запиту (`unsupported_media`).

**Верифікація.** Повний упорядкований `make check` на `a8c1005` (2026-11-14) — `GATE_EXIT=0`:
frontend unit **910 passed**, e2e **131 passed** (15.7 хв), включно з двома новими:
(1) ready-відео → друга доріжка «English» → ready → default → видалення оригіналу; `master.m3u8`
містить 1 → 2 → 1 `EXT-X-MEDIA:TYPE=AUDIO`, `DEFAULT=YES` на English; axe на секції;
(2) video-only фікстура → `awaiting_audio` → мова з клавіатури → m4a → відео `ready` без навігації.
Нові тести: component 13, domain +8. Мутації: `shownVideo` `>`→`<` (2 падіння), delete-guard → `true`
(1), фільтр мов прибрано (1).

**Дефект, спійманий лише повним гейтом.** Гейт на `5927dcd` червоний (117 passed, 1 failed): playback
B40 — strict-mode violation, бо в полі ready-уроку тепер два `input[type="file"]`; таргетні прогони
зелені. Фікс `a8c1005`: усі video-upload локатори e2e скоуплено `:not(.admin-audio-field input)`.

**Незалежне рев'ю (Codex `exec`, свіжий контекст), три хвилі:** W1 на `4f29e50` — BLOCK: Major #2
(нема перевірки розширення до mint) виправлено в `5927dcd`; Major #1 (заміна, що впала через аудіо,
ховається за старшим ready) — відповідь із механізмом (D-12; `failed` аудіо не відновлює,
`postgres/audiotrack.go:421/:519`), записано в D-L2-10. W2 на `5927dcd` — SHIP, 0 знахідок, прийняв
обидві відповіді. W3 (delta e2e-фікса) на `a8c1005` — SHIP, 0 знахідок. Raw:
`_wip-s5l2b-review-codex.md`, `-w2.md`, `-w3.md`.

**Обмеження.** Секція не показується під час `queued/processing` (D-L2-10). Автор не бачить причини
падіння заміни, поки на уроці є старше `ready` (як для відео на `main`). Причина `failed` доріжки не
показується — DTO її не несе. `adminstats` `awaiting_audio` — як у L2a. GitHub CI не запускається
(account-wide, див. пам'ять) — авторитетний гейт локальний.

## S-backfill — embedded-доріжка для Ready-ассетів до S3 (2026-11-14)

План: `docs/multi-audio/plan-sbackfill.md` (D-B1…D-B6, D-B3a; plan-critique Codex, 3 раунди → ACCEPTED).
Гілка `multi-audio-sbackfill` від `main@2292869`. Вибір зрізу — власник, 2026-11-14.

- **Дефект на `main`, закрито:** Ready-ассети, транскодовані до S3, не мали рядка `audio_tracks`, тож перша
  вивантажена доріжка перегенеровувала master лише з неї — оригінальний звук (`audio/index.m3u8`) зникав із
  плеєра; редактор показував порожній список доріжок.
- **Міграція `00028`:** кожен `ready`, не evicted ассет **без жодної Ready-доріжки** отримує embedded
  (`und`/«Original», default, Ready, leaf `original`) — як `Publish` після S3. Мова `und-x-original`, якщо `und`
  тримає не-failed upload; обидві зайняті → відмова. `LOCK TABLE … SHARE ROW EXCLUSIVE` після
  `SET LOCAL lock_timeout='5s'` серіалізує з mint/promote/delete/publish/evict/sweep. Детермінований id
  `md5('multi-audio-backfill:'||asset_id)::uuid` — Down прибирає рівно ці рядки. Ready-ґейт для наявних
  даних: після INSERT кожен Ready-ассет має ≥1 Ready-доріжку й рівно одну default, інакше відмова.
- **Обсяг — звуження, не ремонт усіх (D-B6):** старий ассет, чия uploaded-доріжка **вже стала Ready** до
  міграції, у БД не відрізнити від video-only — міграція його не чіпає; оригінальний звук такого ассета не
  відновлено. Стан досяжний лише між S5 і цією міграцією, у test mode.

**Верифікація.** Integration (реальний goose-провайдер, Postgres 17.5, testcontainers) — 4 нові підтести
зелені: populated data 10 станів (A/A2/A3/H — backfill; B evicted, C embedded від Publish, D video-only з
Ready upload, E processing, F failed, G awaiting_audio — без змін) + down прибирає лише backfill-рядки +
повторний up відтворює ті самі id + promote на backfilled-ассеті дає 2 rendition-и з embedded default;
відмова при зайнятих `und`/`und-x-original` (версія лишилась 27, рядків нема); відмова ґейта (Ready-доріжка
без default) з відкатом backfill-кандидата; незакомічений writer → міграція чекає й падає з lock timeout
(5.17 с), після rollback — успіх. Мутації (кожна ловиться своїм тестом): умова «нуль рядків» → падає
populated; без `LOCK` → падає writer-тест; down «за kind» → падає populated.

**Незалежне рев'ю (Codex `exec`):** W1 на `16fa3d0` — SHIP без доказів (не зараховано як перевірку);
W2 — SHIP з доказами по 5 пунктах, включно з усіма 13 викликами `DownTo/UpTo` нижче 28 у backend (жоден не
зачіпає ґейт: course-фікстури мають лише `queued` відео; інші пакети не понижують схему). Minor W2 —
діагностика нижче. Raw: `_wip-sbackfill-review-codex-w1.md`, `-w2.md`.

**Діагностика відмови `make migrate` (D-B3a).** Runner навмисно ховає помилку goose
(`cmd/migrate/main.go:67-69`), тож при відмові 00028 оператор запускає read-only запити:

```sql
-- 1) Обидві мови зайняті (відмова «already uses both und and und-x-original»):
SELECT a.id FROM video_assets a
WHERE a.state='ready' AND a.evicted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM audio_tracks t WHERE t.video_asset_id=a.id AND t.state='ready')
  AND EXISTS (SELECT 1 FROM audio_tracks t WHERE t.video_asset_id=a.id AND t.state<>'failed' AND t.language='und')
  AND EXISTS (SELECT 1 FROM audio_tracks t WHERE t.video_asset_id=a.id AND t.state<>'failed' AND t.language='und-x-original');
-- 2) Ready-ассет, що лишився б без Ready default (відмова «has no ready default audio track»):
SELECT a.id FROM video_assets a
WHERE a.state='ready' AND a.evicted_at IS NULL
  AND EXISTS (SELECT 1 FROM audio_tracks t WHERE t.video_asset_id=a.id AND t.state='ready')
  AND (SELECT count(*) FROM audio_tracks t WHERE t.video_asset_id=a.id AND t.is_default) <> 1;
-- 3) Lock timeout — хто тримає таблиці:
SELECT l.pid, l.mode, l.granted, a.state, a.query_start
FROM pg_locks l JOIN pg_stat_activity a USING (pid)
WHERE l.relation IN ('video_assets'::regclass, 'audio_tracks'::regclass);
-- D-B6: кандидати, яких міграція свідомо не чіпає (старий ассет з уже Ready upload, без embedded):
SELECT a.id FROM video_assets a
WHERE a.state='ready' AND a.evicted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM audio_tracks t WHERE t.video_asset_id=a.id AND t.kind='embedded');
```
Останній запит дає й video-only ассети (вони не мають embedded за визначенням) — розрізнити їх можна лише
за наявністю каталогу `audio/` на диску, тож це список для ручного огляду, не автоматичний ремонт.

**Обмеження.** D-B6 (вище). master.m3u8 на диску не змінюється (D-B4): назва «audio_0» лишається до
першого regen. ASR не змінено (D-B5, S6). Down — test-only; після нього старі ассети знову без доріжок.
