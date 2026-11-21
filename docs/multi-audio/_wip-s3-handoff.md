# HANDOFF — multi-audio S3 (транскод доріжки + regen master)

Гілка `multi-audio`. **✅ S3 ЗАКРИТО Й ЗМЕРДЖЕНО 2026-11-08 — PR #97 (squash `ca6634b`).** Повний
ordered gate зелений (integration 64 пакети 0 FAIL, check-generated без дрейфу, build, e2e 129
passed); незалежне рев'ю дельти (test-seed fix) — VERDICT PASS; squash-merge за go власника.
Комміти на гілці: `133a20f` (gate-fix test-only), `737ebdf` (закриття S3 у report.md §S3 з
limitations 1–3 латентними до S5). Далі роадмап: S5 (wiring uploaded-track, з попереднім закриттям
обмежень 1–3) або S-backfill/S6/S7 за `plan.md` §5.

Нижче — історичний стан на момент підготовки handoff (виконано цією сесією):

## Що зроблено й закомічено (гілка `multi-audio`)
- `df9a313` **S3a** — video-only render argv-гілка (`Render(…, hasAudio)`), `RenderAudioTrack`,
  чиста `rewriteMasterText`/атомарний `RegenerateMaster` (D7), relax `AcceptSource`, `validLabel`
  manifest-safe. Golden-master'и `testdata/master_*.m3u8` (S0-борг закрито реальним ffmpeg).
- `c62f662` **S3b-1** — embedded-матеріалізація + AwaitingAudio: `Publish(embedded *AudioTrack)`
  (атомарно Ready+track, або AwaitingAudio без eviction); координатор `repo.RegenerateMaster`.
- `f838882` **S3b-2** — uploaded-track transcode-гілка (lean lifecycle, рішення власника 2026-11-08):
  payload `trackId`, `Job.TrackID`, `EnqueueAudioTrack`; `processAudioTrack`; storage audio-leaf;
  `LoadAudioTrack`/`PromoteAudioTrack`(T11+default-election+regen під локом)/`FailAudioTrack`(T12).

## Незакомічено в робочому дереві (НЕ комітила за hook-директивою)
- `backend/internal/media/adapters/tus/server_test.go` — доданий фейк `EnqueueAudioTrack`
  (integration-тег; без нього `make integration` падав build-ом — пастка integration-only).

## Стан gate (ordered `make check`)
- ✅ `format-check`, `lint`, `typecheck`, `test` (повний unit-набір, exit 0).
- ❌ `make integration` — **1 реальний фейл, тест-онлі** (після фіксу tus). Деталі нижче.
- ⬜ `check-generated`, `build`, `e2e` — ще не ганялись.

### Integration-фейл (тест-сетап, НЕ продакшн-баг)
Падають 2 підтести в `audiotrack_integration_test.go`:
`S3 promote readies a video-only asset…` і `S3 concurrent promote keeps exactly one default`.
Помилка: `video_assets_ready_has_renditions` (SQLSTATE 23514) при promote AwaitingAudio→Ready.
Причина: тести сідають AwaitingAudio-ассет через `insertAsset(... domain.AwaitingAudio ...)` з
порожніми renditions, а CHECK вимагає непорожні renditions у стані ready. **У продакшні це не
виникає**: `Publish` (postgres/publish.go) ставить `renditions=$3` тим самим UPDATE і для
`awaiting_audio`, тож реальний AwaitingAudio-ассет завжди має renditions.
**Фікс:** у цих двох тестах сідати AwaitingAudio-ассет із непорожніми renditions (напр. розширити
`insertAsset`/додати UPDATE renditions після вставки, перш ніж promote). Правка тест-онлі.

## Незалежне рев'ю S3 — VERDICT PASS (повний звіт: `docs/multi-audio/_wip-s3-review.md`)
Live-шлях S3 (embedded + video-only Publish) коректний, D4 (no silent playback) тримається,
regen/atomic/lock/path-safety/quoted-string — ок. 3 CONFIRMED-знахідки **латентні в S3**
(uploaded-track шлях не має продакшн-виклику до S5), **блокери для S5** — записати як limitations у
`report.md` при закритті:
1. [severe] `processAudioTrack` класифікує ВСІ probe/render-фейли як transient → після MaxRetry
   asynq кидає задачу, доріжка лишається Queued, ассет навіки AwaitingAudio (T12 не спрацьовує, бо
   Queued рахується як pending). Video-шлях натомість робить термінальний Fail на вичерпанні.
   Треба термінальний FailAudioTrack при RetryCount>=MaxRetry.
2. `PromoteAudioTrack` піднімає AwaitingAudio→Ready, але НЕ ганяє per-lesson eviction (на відміну
   від `Publish`) → video-only уроки з повторними аплоадами накопичують Ready-генерації понад cap=2,
   старі байти не чистяться.
3. `DeleteAudioTrack` (unwired, S5) міняє БД, але не кличе `RegenerateMaster` → видалення Ready-
   доріжки лишає stale `EXT-X-MEDIA` на видалений каталог.
Minor: (4) самотній embedded лишає ffmpeg-ів `NAME="audio_0"` до першого regen (косметика);
(5) video-only аплоад тепер приймається (був reject до S3) і йде в AwaitingAudio без forward-шляху
до S5 — за D4-sequencing, але це жива зміна поведінки вже зараз.

## Наступні дії (у свіжій сесії)
1. Пофіксити 2 integration-тести (renditions у AwaitingAudio-seed) — тест-онлі.
2. Закомітити tus-фейк + тест-фікс (fold у S3b-2 fixup або окремий коміт на гілці).
3. Догнати gate: `make integration` (зелений), `check-generated`, `build`, `e2e`.
   e2e поки покриває лише embedded/video-шлях (uploaded-track UI — S5); upload→Ready має лишитись зеленим.
4. Записати рев'ю-знахідки 1–3 як limitations у `docs/multi-audio/report.md` §S3; записати lean-
   lifecycle як датоване рішення (власник 2026-11-08) і S0-знахідки (GROUP-ID=group_aac, виміряні
   BANDWIDTH, embedded URI=audio/).
5. Після зеленого повного gate — закрити S3 у report.md, мердж гілки за звичаєм (PR для код-фази).
6. Далі роадмап S3 → S-backfill/S5/S6/S7 (`plan.md` §5).
