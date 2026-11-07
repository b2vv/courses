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

**Наступний крок — S2 (persist):** міграція `audio_tracks` (+ розширення DB CHECK `VideoAsset`
станів на `awaiting_audio`), repository, storage-ключі `source/<id>/audio-<trackId>`, атомарний
set-default swap і delete з re-election на рівні транзакцій. Repo-тести проти Postgres + повний
ordered `make check`.
