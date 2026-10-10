
# Лінза «надлишкова складність» — docs/asr/spec.md

ДАЛІ: читаю spec.md цілком, потім AGENTS.md (Architecture), далі наявні механізми:
asynq-адаптер медіа, media-worker main, backend/internal/transcript, frontend recovery.ts,
docker-compose.yml. Ціль — чернетка як окрема серверна сутність (§7.10), окремий сервіс/контейнер,
окрема черга.

ФАКТ: спека прочитана (976 рядків). Нові сутності, які вона вводить:
1) `cmd/asr-worker` — окремий Go-виконуваний файл (§9.6);
2) окремий образ `debian:bookworm-slim` + whisper.cpp + модель 1549 МіБ (§2.2, §6.6);
3) окремий сервіс у docker-compose з власними cpus/mem_limit ≈4 ГіБ (§9.5);
4) окрема черга asynq `asr` з власним таймаутом (§10, §8);
5) серверна таблиця `transcript_drafts` зі станом (running/ready/failed), мовою,
   власним оптимістичним замком і власною дією «зберегти чернетку» (§7.10);
6) два збереження в UI (чернетка / транскрипт) (§7.10);
7) уже відвантажене локальне відновлення IndexedDB (порт+2 адаптери+сервіс+хук, §7.5-7.6),
   яке §7.10 лишає поруч із (5).
ДАЛІ: дивлюсь наявне — transcript domain/ports/adapters, recovery.ts, asynq queue/handler,
media-worker main, docker-compose — щоб перевірити кожну сутність на «виражається наявним».

ФАКТ (наявне, звірено з деревом):
- `media/adapters/asynq/queue.go:35-48` — черга, TaskID, MaxRetry(3), Timeout(30m), Retention.
  Payload = assetId; тип задачі один.
- `media/adapters/asynq/handler.go` — Handler і ErrorHandler **прив'язані до media**:
  metrics.TranscodingJobDuration/TranscodingFailures, domain.FailureReason, валідація assetID.
  Тобто «повтори й обробник помилок переюзаються» — лише на рівні platform, не цього адаптера.
- Реально спільне: `platform/redis` (QueueOption, QueueLogger, MonitorQueue), `platform/httpserver`,
  `platform/runtime.Open`, `observability.Metrics`.
- `platform/redis/queue.go:36` — MonitorQueue **захардкоджено `queue == "media"`**, і гейдж
  QueueDepth пишеться з літералом "media". Нова черга `asr` метрики не отримає без параметра.
- `backend/Dockerfile` — вже є стадія `runtime` = debian:bookworm-slim + ca-certificates +
  USER 10001; стадії api/migrate/upload/worker будуються з неї. 4 cmd-таргети в одному Dockerfile.
- `migrations/00007_transcript.sql` — transcripts(lesson_id, language) + transcript_segments
  (нормалізовані, ordinal, індекс) + course_documents + segment_documents.
- `transcript/ports/repository.go` — SaveLessonTranscript(replace, one tx) повертає токен;
  GetLessonTranscript при порожній мові бере «єдину або першу» → читач НЕ скоупить вибір.
- Фронт: порт TranscriptHistoryStore (22) + 2 адаптери (88+25) + application/transcript/recovery.ts
  (171) + use-transcript-recovery.ts + recovery-dismissal.ts + i18n en/uk — уже відвантажено.

ФАКТ: LessonTranscriptField.tsx:66 — мова жорстко `data?.language ?? 'uk'`, селекта мов у
дереві немає; локалей рівно дві (app/providers/locales = en.ts, uk.ts).
ФАКТ: TranscriptEditor монтується з ключем `${lessonId}-${expectedUpdatedAt ?? 'new'}`,
actions.save — один; TranscriptSnapshot несе РІВНО ОДИН expectedUpdatedAt (recovery.ts:33).

РІШЕННЯ (знахідки лінзи, 10 шт.):
1 blocking — публікація чернетки = наявний SaveLessonTranscript, не новий promote-ендпойнт (§7.10).
2 blocking — три сховища й два токени на один редактор (§7.5/7.6 vs §7.10).
3 non-blocking — чернетка як один документ, без дзеркала transcript_segments.
4 non-blocking — образ ASR = наявна стадія `runtime`; §2.2 ухвалює те, що Dockerfile уже фіксує.
5 non-blocking — «повтори й обробник помилок переюзані» стосується platform, не media/asynq/handler.
6 non-blocking — MonitorQueue хардкодить "media": нова черга без параметра невидима в метриках.
7 non-blocking — селект мов вводить каталог; продукт має дві локалі.
8 non-blocking — нарізка §8.3 знята заміром, але лишилась у §8 як напрям.
9 non-blocking — окремий Redis (§10) винести в §11 з тригером.
10 non-blocking — валідація чернетки тими самими доменними конструкторами.
ЧИСТО: cmd/asr-worker (форма 4 наявних cmd), окремий сервіс у compose (свої cpus/mem_limit),
іменована черга asr, оптимістичний замок чернетки — усе виражається наявною формою, нового
механізму не вводять.
