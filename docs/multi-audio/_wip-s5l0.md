# S5-L0 — worker correctness (close S3 limitations L1 + L2)

Гілка `multi-audio-s5l0` від `origin/main@ca6634b` (S1–S3 змерджено PR #97). Writer=Courses 1 (Courses13),
review_mode=independent_subagent. Власник обрав S5-L0 worker-correctness першим (2026-11-08).

## Checkpoint (pre-code)
- **Scope:** закрити 2 з 3 обмежень S3 (`docs/multi-audio/report.md §S3`), пре-специфіковані рев'ю S3.
  L3 (DeleteAudioTrack+RegenerateMaster) ВІДКЛАДЕНО у S5-L1 (delete-wiring) — рев'ю S3 прямо сказало
  «fix when wiring S5»; його єдиний майбутній caller — S5 delete-handler.
- **Material change?** Ні нового контракту/openapigen. Зміна сигнатури `PromoteAudioTrack` (internal port) —
  симетрична наявному `Publish(…) (published, evicted, err)`; не decision surface, а віддзеркалення патерну.
- **Gates:** повний ordered `make check` (міняє злиту/персистну eviction-поведінку → обов'язковий integration+e2e).
- **Review:** незалежний субагент по дельті перед мерджем (норма власника 26.10). PR (код-фаза).

## L1 — термінальний fail на вичерпанні retry (severe)
`application/transcode.go processAudioTrack`: усі probe/render/publish/promote-фейли → `mediaOperationError`
(transient). Після MaxRetry asynq кидає задачу → трек Queued/Processing, ассет навіки AwaitingAudio.
**Фікс:** helper `audioTrackTransientFailure(ctx, job, cause, reason)` — `RetryCount<MaxRetry` → transient;
інакше `FailAudioTrack` (термінально, T12) + `TranscodeOutcome{Failed:true}`. Дзеркалить video-шлях
`businessFailure` (`reason.Permanent() || RetryCount>=MaxRetry`). finalizationContext (WithoutCancel) → fail
виконується навіть при скасованому ctx. Застосувати на 5 transient-сайтах (probe/purge/render/publish/promote).

## L2 — per-lesson eviction у PromoteAudioTrack
`adapters/postgres/audiotrack.go PromoteAudioTrack`: lift AwaitingAudio→Ready БЕЗ eviction-проходу (на
відміну від `Publish`) → video-only повторні аплоади накопичують Ready понад cap=2. **Фікс:** після
lift (лише коли `assetState==AwaitingAudio`) — той самий readyCount+evict, що `publish.go:107-127`
(ORDER BY generation LIMIT 1; just-promoted = найвища generation → не self-evict). Сигнатура →
`(evicted *domain.VideoAsset, err error)`. Worker (`transcode.go`) purgе+DeleteSwept evicted (дзеркалить
video-шлях рядки 157-164). Оновити: port, impl, worker-caller, fake (upload_test.go), closure
(transcode_test.go), 3 integration-сайти (audiotrack_integration_test.go 212/238/308).

## Журнал
- L1 закомічено `f1b54e4`: helper `audioTrackTransientFailure` на 5 transient-сайтах; юніт-тест
  (table) retries-remain→transient vs exhausted→FailAudioTrack. go test+gofmt зелені.
- L2 закомічено `609d79a`: PromoteAudioTrack → `(evicted, err)`, eviction на AwaitingAudio→Ready
  (дзеркалить publish.go), worker purgе+DeleteSwept evicted. Оновлено port/impl/worker/fake/closure/
  3 integration-сайти + новий integration-тест (3-я Ready-генерація евіктить найстаршу). vet+build
  (incl. integration-tag)+go test -race media зелені; media-postgres integration (targeted) зелений.
- Повний `make check` #1 → RED на `make test`: `TestStartTerminatesRaceContainersBeforeRetrying`
  (testsupport/containers) — відомий B-27 reaper-race flake під навантаженням (помилка прямо каже
  «#3868»). Доведено flake: isolated `-count=20` зелений; пакет media-only дифа не чіпає.
- Повний `make check` #2 → GATE_EXIT=0: testsupport/containers ok, integration (0 FAIL, incl. новий
  evict-subtest), check-generated, build, **e2e 129 passed (17.9m)**.
- Незалежне рев'ю дельти ca6634b..HEAD (gate-reviewer, Opus 4.8) → **VERDICT PASS** (gate green,
  L1/L2 коректні, сигнатура пропагована всюди). Одна реальна латентна знахідка (non-blocking).
- NEXT: виправити коментар (зроблено) → легкий re-verify (build/gofmt, поведінка без змін) → мердж
  PR #98 → S5-L1.

## Знахідка рев'ю (латентна, закрити в S5-L1)
**Self-eviction при промоуті застарілої awaiting_audio-генерації.** `awaiting_audio` НЕ входить у
pending-per-lesson unique index (`00005:36` — лише created/uploading/queued/processing; `00025`
розширив CHECK, але index не чіпав). Тож стара awaiting_audio-генерація (gen1) може лишатися, поки
новіші (gen2,gen3) аплоадяться й стають Ready. Коли аудіо gen1 нарешті приходить → promote gen1 →
readyCount=3 → `ORDER BY generation LIMIT 1` обирає gen1 = just-promoted → **self-eviction**.
- Наслідок: cap-consistent (лишає новіші 2), БЕЗ strand/corruption/double-count; шкода — змарнований
  RegenerateMaster gen1 + UX-дивина (завершений audio-add тихо видаляє стару генерацію).
- Reachability СЬОГОДНІ: НЕМА — `CreateAudioTrack`/`EnqueueAudioTrack` без non-test callers
  (audio-upload endpoint не зварено до S5-L1). Scaffolding-шлях.
- **Фікс у S5-L1 (перед wiring audio-upload):** або guard промоуту лише до latest-генерації (short-
  circuit superseded), або явно виключати just-promoted з eviction. Рішення — продуктова семантика
  S5-L1. Коментар `audiotrack.go` виправлено на правдивий (було хибне «highest generation»).
