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
