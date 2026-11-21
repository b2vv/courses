# S5-L1b — незалежне рев'ю (Codex), вердикт BLOCK

Дата: 2026-11-10. Рецензент: Codex (codex-cli 0.163, read-only sandbox, свіжий контекст).
Діфф: `57d3832..0b04f5a` (PR #100, уже злито в main). Усі знахідки звірені мною з кодом.

## F1 — Major, CONFIRMED: небезпечний sweep
`adapters/postgres/audiotrack.go:767` `DeleteAudioTrackSwept` робить `DELETE … WHERE id=$1`
**без фільтра стану/expiry**. `SweepAudioTrackCandidates:744` бере кандидатів (created/uploading,
expires_at<now) без FOR UPDATE. PATCH, що стартував до expiry з pre-expiry `now`, тримає lock,
комітить Queued; sweep уже вибрав рядок, почистив файли (`transcode.go:297`), і DELETE зносить
уже-Queued трек. Video має той самий unguarded DELETE — «як у video» не виправдання.

## F2 — Major, CONFIRMED: втрата enqueue назавжди
`application/audioupload.go:117`. Completion комітить Queued У ТРАНЗАКЦІЇ `AppendAudioChunk`,
а `EnqueueAudioTrack` — ПІСЛЯ коміту. Збій Redis / смерть процесу між ними → трек Queued без
задачі. HEAD/PATCH відкидають Queued, sweep його не бере, `Recover` обробляє лише video_assets.
Трек застрягає назавжди.

## F3 — Major, CONFIRMED (гірше за video): витік байтів
`adapters/postgres/audiotrack.go:724` Cancel ставить `state='failed'` (не evicted_at, байти не
чистить). `SweepAudioTrackCandidates:744` бере лише `created/uploading`. Доказ дивергенції:
video-sweep `sql/queries/media.sql:66-69` прибирає `evicted_at IS NOT NULL` **І** `failed` старші
30 днів. Audio-sweep має лише гілку created/uploading → scratch-байти скасованого/eviced-at-mint
треку течуть назавжди, рядок лишається Failed. Термінальний transcode-fail лишає source-байти так само.

## F4 — Major, CONFIRMED: обіцяного покриття немає
Немає `audioupload_test.go`, немає `audioserver_test.go`. Усі audio-рефи в тестах —
стаби (`upload_test.go`, `handler_test.go`) + фікстури інтеграційного. Нема тестів на: mint-авторизацію,
append/resume/offset/size, completion+enqueue-fail, cancel, sweep/concurrency, tus HEAD/PATCH/DELETE,
HTTP AT09–AT11. План обіцяв application/integration/transport-покриття.

## Minor
- `migrations/00027:9` ADD CHECK дасть збій, якщо є pre-existing uploaded created/uploading/queued/processing
  рядок із declared_bytes=0 (PLAUSIBLE; baseline прод не мав upload-шляху).
- `openapi_contract.go:34` audio реюзить MediaBadRequest опис «over 2 GiB», а реально межа 500 MiB.

## Виправдані відмінності (не дефекти)
Окремий сервіс/сервер, відкладена codec-валідація, track-ключі (колізії з video немає),
wiring у cmd/upload. Offset-рівність + declared-bounds + 8-MiB chunk не дають received-overflow.
`persistenceError(nil)`→dependency-unavailable, але всі нові call-site гардять non-nil.
Авторизація = як video: володіння UUID = capability (нема session/minter-перевірки) — свідомий existing design.

## Процесний висновок
Моє раннє fork-рев'ю повернуло CLEAN — бо fork успадкував мій контекст і сліпу пляму. Мердж
пройшов під несправжнім «незалежним» рев'ю. Урок записано: [[fork-review-shares-blind-spot]].
