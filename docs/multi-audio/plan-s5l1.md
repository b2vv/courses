# Plan — multi-audio S5-L1 (authoring contract)

Base `origin/main@640b6e2` (S5-L0 merged). Writer=Courses 1, review_mode=independent_subagent.
Decision surface for a dependent openapigen implementation → потребує plan-critic ПЕРЕД кодом
(pre-code checkpoint). Фронтенд UI — окремий слайс S5-L2 (не тут).

## Мета
Авторам (admin/instructor власник курсу) керувати аудіо-доріжками уроку: додати (аплоад), список,
переназвати (language/label), призначити default, видалити. Закриває **L3** (DeleteAudioTrack без
RegenerateMaster) і **self-eviction** (S5-L0 finding), бо саме цей слайс вмикає uploaded-track шлях.

## Поточний seam (факти, surface-map 2026-11-08)
- Repo має `CreateAudioTrack/ListAudioTracks/SetDefaultAudioTrack/DeleteAudioTrack` — НЕ на порту,
  БЕЗ actor/ownership. `EnqueueAudioTrack`/`RenderAudioTrack` без call-site.
- Ownership: SQL-join `course_lessons→course_modules→courses.instructor_id` + `actor.CanEditLesson`
  всередині repo (напр. `Mint` repository.go:99-130). Аудіо-методи його НЕ мають.
- `audio_tracks` (00025): без declared/received bytes, без claim_token. Унікальність лише
  `one_default_per_asset`. Немає (language, asset).
- Video upload: `POST /admin/lessons/{lessonId}/video`→MintTicket→`uploadTicketDTO{VideoID,UploadURL
  "/uploads/{id}",...}`; tus `/uploads/{videoId}`; Enqueue у application AppendChunk при переході в
  Queued; source leaf "original". Audio source leaf = `audio-<trackId>` у source-дирі ассета.
- openapigen: RouteBinding (contract.go) + NameMap + overlay (operationId/summary/desc/params/headers).
  Gate: coverage-тест (mux.Handle set == Routes set), contract-тест (hard-coded media paths +
  byte-identical generation) — нові шляхи додати в обидва + overlay.

## Рішення (кожне закриває альтернативу; DATED 2026-11-08, author Courses 1)

### D-S5L1-1 — Ендпоінти ключовані на videoId (assetId), вкладені track-операції
Доріжка належить АСЕТУ. Дзеркалимо `/videos/{videoId}/retry`. UI резолвить урок→поточний Ready-асет→
videoId. Усі track-операції вкладені (несуть assetID у шляху, що репо-методи вже беруть):
- `POST   /api/v1/admin/videos/{videoId}/audio-tracks` — mint audio-ticket; body {language,label,
  declaredBytes,originalName} → 201 `MediaAudioUploadTicket{trackId,uploadURL,chunkBytes,maxBytes,expiresAt}`.
- `GET    /api/v1/admin/videos/{videoId}/audio-tracks` → 200 `MediaAudioTracks{tracks:[MediaAudioTrack]}`.
- `PATCH  /api/v1/admin/videos/{videoId}/audio-tracks/{trackId}` — body {language?,label?} → 200 `MediaAudioTrack`.
- `POST   /api/v1/admin/videos/{videoId}/audio-tracks/{trackId}/default` → 200 `MediaAudioTrack`.
- `DELETE /api/v1/admin/videos/{videoId}/audio-tracks/{trackId}` → 204.
Закрита альтернатива: trackId-only шляхи (`/audio-tracks/{trackId}`) — відкинуто, бо ownership-join
стартує з ассета, а репо-методи вже беруть (assetID,trackID); trackId-only потребував би зайвого lookup.

### D-S5L1-2 — Upload аудіо ⚠️OWNER (BLOCKING L1b, під-специфіковано — critic finding 1+2)
`audio_tracks` не має byte-колонок, lean lifecycle (owner 2026-11-08) пропускає Uploading-handshake.
**Це ВІДХИЛЕННЯ від D3 «реюз upload-машини»:** наявний tus/Upload жорстко зв'язаний із
`video_assets.received_bytes/declared_bytes` (upload.go:143) — byte-less таблицю as-is не бере.
Чорнова ПРОПОЗИЦІЯ (НЕ закрита, бо без механізму): tus-роут `/uploads/audio/{trackId}`, offset із
РОЗМІРУ scratch-файлу. **Три під-проблеми, які мусить закрити owner-рішення, перш ніж це decision:**
  (а) **детекція завершення** — без received/declared у БД як дізнатись, що аплоад завершено?
      (tus-фінальний чанк? окремий complete-POST? declaredBytes має десь персиститись);
  (б) **ліміт розміру** — server-enforced max (video має MaxBytesReader+declared); де тут;
  (в) **покинутий Created-трек** — `audio_tracks` без expires_at/sweep (на відміну від video_assets):
      minted-but-never-completed трек ВІЧНО тримає (language,asset)-слот (D4 index WHERE state<>failed
      звільняє лише failed). Треба expiry/cleanup-відповідь.
**Закрита альтернатива A:** declared/received_bytes+claim+expires_at у audio_tracks (schema) — важче,
але дає server-invariant і sweep як у video. **B:** single-PUT без чанків — ~50-100MB незручно.
⚠️ **ВИНЕСТИ ВЛАСНИКУ з названими (а)(б)(в):** leaner byte-less tus vs byte-колонки+sweep.
**РІШЕННЯ ВЛАСНИКА 2026-11-08: byte-колонки + sweep.** L1b додає `declared_bytes`/`received_bytes`/
`expires_at` у `audio_tracks` і реюзить проваджений video tus-патерн: server-enforced ліміт (MaxBytesReader
+declared), детекція завершення (received==declared), sweep покинутих Created за expires_at. Закриває
під-проблеми (а)(б)(в); справжній реюз upload-машини (D3), без відкритих хвостів. Закрита альтернатива:
byte-less tus — відкинуто (3 невирішені під-проблеми). Схема — у міграції L1b, не L1a.

### D-S5L1-3 — Ownership на аудіо-репо-методах через той самий join
`CreateAudioTrack/ListAudioTracks/SetDefaultAudioTrack/DeleteAudioTrack` отримують `actor domain.Actor`
і роблять `course_lessons→…→courses.instructor_id` join + `CanEditLesson` (точно як Mint). Асет-scoped:
join від video_assets.lesson_id. Закрита альтернатива: перевірка в application (load asset→owner) —
відкинуто для консистентності з наявним repo-паттерном (ownership у SQL під тим самим локом).

### D-S5L1-4 — Унікальність (language, asset): partial unique index WHERE state<>'failed'
Міграція `00026`: `CREATE UNIQUE INDEX audio_tracks_language_per_asset ON audio_tracks
(video_asset_id, language) WHERE state <> 'failed'`. WHERE state<>'failed' → провалений аплоад не
блокує повтор тією ж мовою (low-risk, випливає з index). embedded "und" + uploaded-мови не конфліктують.
**409-дискримінація (critic finding 4, у L1a):** `audioTrackConflict` (audiotrack.go:458-470) зараз
мапить `23505` генерично на `ErrInvalidRequest`/400. Додати гілку за ІМЕНЕМ констрейнта
`audio_tracks_language_per_asset` → новий `ErrDuplicateLanguage`/409. Досяжно вже в L1a через
PATCH-зміну мови (не лише L1b mint). ⚠️ **ВИНЕСТИ ВЛАСНИКУ (minor):** чи дозволені дві "und"
(дефолт — ні, index забороняє; uploaded-трек рідко ставить "und").

### D-S5L1-5 — L3: DeleteAudioTrack + RegenerateMaster під локом
`DeleteAudioTrack(... rewrite func(ready))` (дзеркалить PromoteAudioTrack): після DELETE+re-election,
перед commit — `rewrite(readySet)` = RegenerateMaster. Application Delete передає RegenerateMaster.
Закриває L3. Edge: видалення останньої Ready embedded з Ready-ассета вже заборонено ValidateTrackDeletion.

### D-S5L1-5b — Regen master на ВСІХ трьох мутаціях, з API-процесу (рішення 2026-11-08, Courses 1, під час L1a)
Знахідка під час реалізації (critique її не бачила): **SetDefault і Rename теж змінюють master** —
`DEFAULT=YES`, `NAME`, `LANGUAGE` живуть у `EXT-X-MEDIA` (master.go formatMediaLine). Regen лише на
delete лишав би stale default/label у плеєрі. Тож `SetDefaultAudioTrack`, `UpdateAudioTrackMetadata`,
`DeleteAudioTrack` усі приймають `rewrite func(ready)` і викликають його під тим самим асет-локом перед
commit. Application пропускає regen, коли Ready-набір порожній (`RegenerateMaster` відкидає `len==0`;
не-Ready ассет не має audio-секції для публікації).
**Хто пише master:** API-процес, через `mediaffmpeg.New("/media")` (лише `RegenerateMaster` — чистий
Go, без ffmpeg-бінарника). **Механізм перевірки:** api і media-worker — той самий uid `10001:10001`
(Dockerfile `runtime` stage), обидва монтують `media-data:/media` rw (compose), а запис серіалізується
DB-локом `FOR UPDATE OF assets`, який тримають і promote, і ці мутації — D7 single-writer тримається
локом, не процесом; atomic temp+rename — nginx не бачить напівзапис. **Закрита альтернатива:** окрема
asynq-задача `regenerate_master` у воркер після commit — відкинуто: eventual-consistency вікно (stale
master між commit і задачею) + ризик втраченого enqueue після commit, без виграшу в безпеці.
**Зсув межі:** API вже пише `/media/i` (images); тепер ще й `master.m3u8` у HLS-дереві. Винести власнику
в звіті як boundary-зміну (не блокує — механізм вище).

### D-S5L1-6 — Self-eviction: РЕКОМЕНДАЦІЯ лишити як є ⚠️OWNER (UX-семантика, НЕ forced fix — critic finding 3)
S5-L0 finding: пізній промоут awaiting_audio-генерації, витісненої новішими, self-evict'иться.
**ВАЖЛИВА корекція (critic):** поточна поведінка НЕ марнотратна-і-strand — вона РЕКЛЕЙМИТЬ байти
застарілої генерації (`PromoteAudioTrack` evict → `purgeAsset`+`DeleteSwept`, transcode.go:228-235);
cap-consistent, без strand/corruption (підтверджено S5-L0 рев'ю). **РЕКОМЕНДАЦІЯ: лишити як є** —
тільки вже-виправлений правдивий коментар; жодного коду. Закрита альтернатива (моя попередня):
short-circuit→Failed-«superseded» — **відкинуто**, бо лишає video-байти застарілої генерації без
reclaim (СТВОРЮЄ strand, якого зараз нема). ⚠️ **ВИНЕСТИ ВЛАСНИКУ (опційно):** прийняти «лишити як є»
(рек.) чи все ж short-circuit заради економії одного RegenerateMaster (ціною нового sweep-шляху).

## Декомпозиція (leaf-first; кожен — незалежне рев'ю + ordered gate; окремий PR)
- **S5-L1a — metadata+delete (БЕЗ upload):** port+app+ownership для List/SetDefault/SetLanguage(PATCH)/
  Delete(+L3 regen); міграція 00026 унікальність (D4); openapigen 4 ендпоінти (GET/PATCH/default/DELETE)
  + overlay; contract+coverage тести; integration ownership+uniqueness. НЕ вмикає worker-шлях.
- **S5-L1b — upload+enqueue+self-eviction:** mint-ticket ендпоінт + tus audio-роут (D2) +
  EnqueueAudioTrack wiring + self-eviction fix (D6). Вмикає uploaded-track шлях end-to-end; повний
  ordered gate + e2e (новий multi-audio HLS-сценарій частково — повне відтворення у S7).
Чому так: L1a — чисті CRUD+ownership+uniqueness без нового upload-механізму (нижчий ризик, не чіпає
proven upload); L1b — єдиний, що вмикає worker і несе owner-decision про upload+self-eviction.

## Тести (per-leaf)
- Domain: uniqueness-перевірка, PATCH-валідація, delete-invariant.
- Repo integration: ownership-відмова (чужий instructor), uniqueness-конфлікт **409 через PATCH-зміну
  мови** (L1a, не лише mint), **delete embedded-track edge** (critic finding 5), delete→regen,
  set-default swap через API-шлях, set-default на non-Ready → відмова.
- HTTP contract: 4(L1a)+1(L1b) нові routes; coverage-тест (mux==Routes); overlay byte-identical.
- L1b: e2e upload→Queued→Ready; наявний upload→Ready лишається зеленим.

## Plan-critique (GATE 2, independent_subagent, 2026-11-08) — VERDICT READY
L1a surface звірено по коду — sound, можна імплементувати. 2 blocking-item — лише для L1b, свідомо
за owner-рішеннями й L1b-gate (не сховані під зеленим гейтом): (1) D2 upload-механізм під-специфіковано;
(2) abandoned-Created strand (expiry). Non-blocking, інкорпоровано вище: (3) self-eviction лишити як є
(reclaim-byte корекція); (4) 409-дискримінація в L1a; (5) embedded-delete тест-кейс.

## Відкриті owner-рішення
- **BLOCKING L1b (не L1a):** D-S5L1-2 upload-механізм — названо під-проблеми (а)детекція завершення,
  (б)ліміт, (в)expiry покинутого Created. Byte-less tus vs byte-колонки+sweep.
- **Minor:** D-S5L1-4 дубль "und" (дефолт — ні). D-S5L1-6 self-eviction (рек. лишити як є).
- **L1a НЕ заблоковано** цими — metadata+delete+ownership+uniqueness не чіпають upload.

## Журнал
- Surface зібрано (Explore). План написано. Plan-critic → READY (L1a), 2 blocking quarantined у L1b.
- Інкорпоровано critique findings 3-5. Owner: старт L1a зараз; D2=byte-колонки+sweep (L1b).
- **S5-L1a data+app ЗАКОМІЧЕНО `c10cd7d`** (гілка `multi-audio-s5l1a` від origin/main@640b6e2):
  migration 00026 (uniqueness D6); repo ownership на List/SetDefault/Delete + новий
  UpdateAudioTrackMetadata; DeleteAudioTrack+RegenerateMaster (закрив L3); ErrDuplicateLanguage/409;
  app AudioTracks service; port+4 методи; 4 integration-тести (ownership/409/delete-regen/embedded).
  Media postgres integration ЗЕЛЕНИЙ; backend build/vet зелені.
- **2026-11-09 handoff #2 — НЕЗАКОМІЧЕНО на `multi-audio-s5l1a` поверх `c10cd7d` (перевірено: build OK,
  media postgres integration зелений, media unit зелений, gofmt/vet чисті):** D-S5L1-5b —
  `SetDefaultAudioTrack`/`UpdateAudioTrackMetadata` тепер теж приймають `rewrite` (хелпер
  `rewriteReadyTracks` у audiotrack.go, Delete переведено на нього); `ports.MasterWriter` (transcoder.go);
  app `AudioTracks` бере `MasterWriter`, regen на всіх 3 мутаціях, пропуск порожнього Ready-набору;
  fakes/stubs/integration-виклики оновлено; тест rename: rewrite бачить нові NAME/LANGUAGE + failing
  rewrite відкочує rename. Файли: http/handler_test.go, postgres/audiotrack.go,
  postgres/audiotrack_integration_test.go, application/audiotrack.go, application/upload_test.go,
  ports/repository.go, ports/transcoder.go.
  **Перший крок свіжої сесії:** (а) додати асерт у `testSetDefaultAudioTrack` — rewrite отримує новий
  default (зараз там `noMasterRewrite`); (б) закомітити цей D-5b шар явними шляхами; (в) далі HTTP-кроки нижче.
  Composition root (`cmd/api/main.go:167`): `mediaapplication.NewAudioTracks(mediaRepository,
  mediaffmpeg.New("/media"), time.Now)` → передати в `mediahttp.NewHandler`.
- **NEXT (HTTP+openapigen layer L1a, свіжа сесія):**
  1. HTTP handler: 4 методи у `adapters/http/handler.go` (GET list, PATCH rename, POST default,
     DELETE) + DTO (audioTrackDTO, audioTracksDTO, renameRequest) + strict JSON decode (як decodeMintVideo);
     actor via mediaActor; map domain errors (ErrDuplicateLanguage→409, ErrCourseForbidden→403,
     ErrVideoNotFound→404, ErrInvalidRequest→400). uploadService interface — додати AudioTracks-методи
     АБО новий audioService interface + поле Handler.
  2. RegisterAdmin: 4 нові `mux.Handle` (GET без requireSameOrigin; PATCH/POST/DELETE з requireSameOrigin).
  3. openapigen contract.go: 4 RouteBinding + NameMap (MediaAudioTrack/MediaAudioTracks/MediaAudioRename)
     + Enums (kind/state) + overlay 4 path-блоки (operationId/summary/desc/params/RequestID+NoStore headers).
  4. Тести: openapi_coverage (mux==Routes — ОНОВИТИ hard-coded path list), openapi_contract fidelitySurface
     (+4 paths, byte-identical), handler contract-тести.
  5. Composition root: зареєструвати AudioTracks service (NewAudioTracks) + wire у RegisterAdmin.
  6. Повний ordered `make check` (міняє контракт → integration+e2e обов'язкові; check-generated після
     generate+staging) → незалежне рев'ю дельти → PR S5-L1a → мердж.
- **2026-11-09 D-5b закомічено `3a78349`:** rewrite callback у SetDefault/Rename/Delete;
  ports.MasterWriter; тест setDefault перевіряє, що rewrite бачить нового default.
- **2026-11-09 HTTP шар закомічено `93f48ff`:** 4 ендпоінти (GET list, PATCH rename, POST default,
  DELETE) + audioService interface + composition root (mediaffmpeg.New → MasterWriter);
  decodeRenameTrack strict JSON; ErrDuplicateLanguage→409; openapigen: 7 routes, overlay,
  MediaAudioTrack/MediaAudioTracks/MediaAudioRename schemas, AudioTrackKind enum,
  DuplicateLanguage response; coverage+fidelity тести (7 paths); 8 handler contract тестів
  (AT01–AT08). Tus test fix `151b508`.
- **Gate status:** format-check ✅, lint ✅, typecheck ✅, test ✅, check-generated ✅, build ✅,
  integration media+tus ✅ (identity flake — Ryuk race, не наш код), e2e 93/93 passed
  (1 unrelated flake admin.spec.ts:182 — sort URL param, жоден admin-файл у дифі).
  Залишається: незалежне рев'ю, PR → мердж.
- Потім **S5-L1b** (upload+enqueue+self-eviction, byte-колонки+sweep за D2-рішенням).
