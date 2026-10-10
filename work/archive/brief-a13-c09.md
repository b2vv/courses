# Brief A13 / крок 5b — C09: стара генерація не перезаписує нову (інваріант у місці запису)

Частина A13 (Phase 8). Бриф — сесія Courses 1, 2026-09-24. Рішення про фікс — рецензент
Courses 2 (UPSERT-guard), вона ж жене integration-гейт.

## Дефект (доведено тестом кроку 5, механізм — `work/report-a13-integration.md`)

Заміна відео soft-evict-ить старий рядок `video_assets` і вставляє новий з наступною `generation`
(`backend/internal/media/adapters/postgres/repository.go:132-146`). Старий рядок живе до `DELETE` у
sweeper-і. `Claim` (`backend/internal/transcript/adapters/postgres/draft.go:101-105`) перевіряє лише
наявність рядка `(lesson, asset, generation)`, а UPSERT (`:116-126`, `ON CONFLICT (lesson_id) DO
UPDATE`) перезаписує безумовно. Запізніла задача старої генерації переводить готову чернетку нової
генерації в `running`, а `Complete` потім пише старий текст. Тест `testDelayedOldClaim` у
`draft_integration_test.go` — червоний. У продакшні цей порядок зараз не виникає (Concurrency 1,
FIFO, у чергу ставиться лише поточна генерація), але цю безпеку дає порядок черги, а не інваріант у
коді.

## Зроби

1. У `Claim` додай до UPSERT умову
   `ON CONFLICT (lesson_id) DO UPDATE SET … WHERE asr_drafts.generation <= EXCLUDED.generation`.
   Саме `<=`, а не `<`: повторна `Claim` тієї самої генерації (перегенерація) мусить проходити.
2. Коли умова відсікає запис, `RETURNING` не повертає рядка, і `Scan` дає `pgx.ErrNoRows`. Це не
   помилка сховища: поверни `ports.ClaimedDraft{Claimed: false}` без помилки. `Process` уже
   обробляє `!claimed.Claimed` (`application/recognize.go`) — задача мовчки завершується, нічого не
   пише. Помилка `23503` (FK) лишається `ErrNotFound`, як є.
   **Порядок гілок обов'язковий:** зараз після `Scan` є лише `23503 → ErrNotFound` і
   `persistenceError`. Без явної гілки `errors.Is(err, pgx.ErrNoRows)` **перед** `persistenceError`
   фікс перетворить запізнілу задачу на помилку сховища. Жодної failure-метрики й Error-логу на цій
   гілці — рішення рецензента: це «не твоя черга», як і наявний `!Claimed`-шлях.
3. Коментар поряд з умовою — одне речення: інваріант тримається в місці запису, щоб не залежати
   від порядку черги (Concurrency 1, FIFO).
4. Тести: `testDelayedOldClaim` стає зеленим: чернетка генерації 2 лишається `ready` з її текстом,
   `Claim` генерації 1 повертає `Claimed: false`. Звичайний порядок
   (стара `Claim` → нова `Claim` → стара `Complete` = `ErrDraftSuperseded`) лишається зеленим.
   Додай unit або integration на те, що повторна `Claim` **тієї самої** генерації проходить.
   Перевір, що інші integration-тести `draft_integration_test.go` не зачеплено.

## NOT
- Не чіпай `media`, міграції, фронтенд, e2e, `openapi.yaml`. Git, `tasks.md`, `report.md` — ні.

## Перевірки
`gofmt`, `go vet ./internal/transcript/...`, `go test ./internal/transcript/...`,
`go test -tags integration -run '^$' ./internal/transcript/...`. Самі integration-тести без
Docker не запускаєш — познач чесно; їх жене Courses 2.

## Звіт
`work/report-a13-c09.md`
