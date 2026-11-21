# S5-L1b fix-forward — план (гілка multi-audio-s5l1b-fixes)

Власник обрав повний обсяг F1–F4 + тести (2026-11-10). Міграції не потрібно.

## Рішення (2026-11-10, артефакт за AGENTS.md)
Uploaded audio-трек у `failed` із `received_bytes < declared_bytes` (cancel / mint-eviction —
незавершене вивантаження) — це upload-junk без продуктової цінності: sweep чистить scratch-байти
й видаляє рядок. `failed` із `received == declared` (transcode-fail, видимий у ListAudioTracks для
retry) лишається й реапиться після 30 днів, як video failed-retention (`media.sql:69`). Перевірка:
sweep-предикат нижче; тест на обидві гілки. Дивергенція від video (video тримає created/uploading
+ evicted + failed-30d): audio не має evicted_at, тож незавершене розрізняємо за received<declared.

## Слайси
1. **repo + ports (F1+F3+F2):**
   - `SweepAudioTrackCandidates`: предикат `(created/uploading AND expires_at<now) OR (failed AND received<declared) OR (failed AND received=declared AND updated_at<now-30d)`.
   - замінити `DeleteAudioTrackSwept` → `DeleteSweptAudioTrackLocked(ctx, trackID, now, purge func() error)`: lock row FOR UPDATE, re-verify sweep-eligibility під локом, purge(), DELETE, commit; промоутнутий Queued → 0 rows → ErrInvalidStateTransition (skip, без purge). Закриває TOCTOU F1.
   - `RecoverQueuedAudioTracks(ctx, now, limit)`: queued uploaded, updated_at<now-10хв; повертає (assetID, trackID).
2. **app (F1+F2):**
   - `transcode.sweep`: audio-петля → `DeleteSweptAudioTrackLocked` з purge-колбеком (scratch+source); skip на ErrInvalidStateTransition.
   - `transcode.Recover`: після video — re-enqueue кожного RecoverQueuedAudioTracks через EnqueueAudioTrack(attempt=1), asynq дедупить.
3. **minors:** OpenAPI audio 400-опис «over 2 GiB»→«over 500 MiB». CHECK-minor — pre-launch, нема populated-рядків; задокументувати в report.
4. **tests (F4):** audioupload_test.go (append happy/offset/over-declared/complete→queued→enqueue/enqueue-fail/head-gating/cancel), audioserver_test.go (HEAD/PATCH/DELETE/offset/405), handler mint-тести (AT09–AT11 контракт), integration (sweep: failed-incomplete cleanup + promote-skip race; recover queued).
5. **gate + незалежне Codex-рев'ю** перед мерджем.

## Статус
- [x] slice1 repo+ports (sweep-предикат, DeleteSweptAudioTrackLocked, RecoverQueuedAudioTracks)
- [x] slice2 app (transcode.sweep locked purge+delete; transcode.Recover audio re-enqueue)
- [x] slice3 minors (OpenAPI 500 MiB; CHECK-minor задокументовано pre-launch)
- [x] slice4 tests: audioupload_test (7), transcode audio sweep/recover (3), handler mint AT09–AT12, tus audioserver_test (7), integration sweep/delete-locked/recover (3)
- [ ] slice5 gate+незалежне Codex-рев'ю

## Виправлені баги по ходу
- recover SQL: `$1` лише в `$1 - interval` → pgx виводив interval; каст `$1::timestamptz` (і в sweep-предикаті).
