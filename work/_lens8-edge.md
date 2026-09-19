# ЛІНЗА EDGE CASES — /Users/strelia/projects/courses/docs/asr/spec.md

ДАЛІ: читаю spec.md цілком; потім backend/internal/transcript/{domain,adapters/postgres}, backend/internal/media/ (generation, витіснення ассетів, життєвий цикл), backend/cmd/media-worker/main.go. Шукаю: порожні/граничні дані, повторний і паралельний виклик, відмова залежності, часткове застосування, заміна відео під час ASR, видалення уроку/курсу з чернеткою, рестарт під час 38-хв задачі.

ФАКТ: spec.md прочитано цілком (976 рядків). Опорні тези для лінзи:
- §4.1 ASR стартує САМ, коли відео «готове до перегляду» (ready-ассет).
- §7/§7.10: машина пише ЧЕРНЕТКУ в нову таблицю модуля transcript (ключ lesson_id, ОДНА чернетка на урок, перегенерація перезаписує), транскрипт не чіпає. Стани чернетки running/ready/failed, є мова.
- §7.10 п.4: оптимістичний замок і в чернетки — «два письменники: автор і ASR».
- §8: asynq Timeout(30 хв) у media/adapters/asynq/queue.go:41, Concurrency:1 у cmd/media-worker/main.go:70; нова іменована черга `asr` із власним таймаутом.
- §6.8: 38 хв на годинний урок (turbo, 6 CPU, 8 ГіБ); стелю A5 знято.
- §4.5: тиша/музика/чужа мова → «названий» результат (не сказано ЯКИЙ і хто його виробляє).
- §9.6: cmd/asr-worker, ports/transcriber.go, adapters/whisper, application/recognise.go; крос-модульне ЧИТАННЯ video_assets — прецедент §9.2.
- §11 out of scope: переклад, діаризація, GPU.

ДАЛІ: читаю код — transcript/domain, transcript/adapters/postgres, media/domain/asset.go (generation, витіснення), media/application/transcode.go, media/adapters/asynq/*, cmd/media-worker/main.go. Шукаю: чи є generation/eviction, що тримає lesson_id, як видаляється урок/курс, як воркер реагує на скасування.

ФАКТ (код, звірено 2026-09-19):
- media/adapters/postgres/publish.go: при readyCount > 2 найстаріший ready-ассет отримує evicted_at; далі transcode.Process → purgeAsset(source, hls, thumbnails, temporary) + DeleteSwept. Тобто БАЙТИ ДЖЕРЕЛА старої генерації видаляються, поки ASR може їх читати.
- transcript/domain/transcript.go:155 + 00007: транскрипт несе authored_against_generation, Stale() деривує застарілість. У чернетки (§7) ключ lesson_id — генерації немає.
- cmd/media-worker/main.go:70 — Concurrency:1, ShutdownTimeout 10s; docker-compose stop_grace_period 40s. 38-хв задача гине за 10 с.
- media закриває це стартовим Recover(): RecoverProcessingMediaAssets (processing→queued) + RecoverQueuedMediaAssets (queued старші 10 хв, attempt+1). Аналогу для чернетки спека не називає.
- media/adapters/asynq/queue.go: TaskID(assetID+":"+attempt), Retention(24h), ErrTaskIDConflict → return nil (ТИХИЙ успіх). Переюз цієї форми з lesson-ключем робить «перегенерувати» no-op на 24 год.
- 00005: video_assets.lesson_id ON DELETE SET NULL; 00007: transcripts.lesson_id ON DELETE CASCADE. course/adapters/postgres/authoring.go:37 deleteMissingAuthoringLessons — видалення уроку є рядовим save-tree. media: dispositionOf() → processDone при LessonID == "" (тихий успіх без ретраю).
- transcript/domain/transcript.go:110 NewSegment: text 1..1000 рун і non-blank; 00007 CHECK те саме → порожні whisper-сегменти не збережуться.
- NewLanguage/validLanguageTag і CHECK у 00007 ПРОПУСКАЮТЬ рядок "auto"; getTranscript ORDER BY language LIMIT 1 → "auto" виграє в "en"/"uk".
- media/domain/asset.go:12 MaximumDeclaredBytes 2 ГіБ; AcceptSource (source.go:47) НЕ перевіряє тривалість. duration_seconds nullable (UsableDurationSeconds → nil при probe ≤ 0).
- frontend/src/application/transcript/recovery.ts: TranscriptSnapshot.expectedUpdatedAt — токен ТРАНСКРИПТА; коментар «recognition writes while the author is away» після §7.10 неправдивий (ASR пише чернетку). Токен чернетки знімок не несе → offer замість offer-stale.
- media/application/transcode.go: перевірка free < 3*DeclaredBytes перед роботою, StorageUnavailable, розділення бізнес-відмови й залежності. ASR-еквівалента спека не називає.
- docker-compose.yml: сервіси з <<: *runtime; nginx read_only+tmpfs; media-worker має том media-data. Для asr-worker §9.5 не називає ні тому, ні tmpfs під WAV, ні gid для дерева 0750/0640.

РІШЕННЯ: 12 знахідок — 8 blocking, 4 non-blocking. Перелік у фінальній відповіді.
