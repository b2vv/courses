# Report A13 / крок 5 — integration-докази C21, C22, C09, C01

Дата: 2026-09-24

## Зроблено

- У `backend/internal/transcript/adapters/postgres/draft_integration_test.go` додано справжні Postgres-тести:
  - C21: `Claim` для ассета іншого уроку повертає `domain.ErrNotFound`, рядок для чужого уроку не створюється.
  - C22: справжній `DELETE FROM video_assets` каскадно прибирає готову чернетку. Коментар у тесті фіксує рішення D8-15: `CASCADE` свідомий, `SET NULL` був би хибним виправленням.
  - C09: звичайний порядок `стара Claim → нова Claim → стара Complete` дає `ErrDraftSuperseded`.
  - C09: форсований порядок із soft-evicted старим ассетом; він є тестом на те, що стара генерація не затирає нову.
  - C01: `PendingRecognitions` перевіряє ready-ассет без чернетки поточної генерації, новішу генерацію проти старої чернетки, а також виключає evicted і not-ready ассети.
- У `backend/internal/transcript/adapters/asynq/queue_integration_test.go` додано Redis/asynq-тест повторної постановки тієї самої задачі без маркера: другий виклик повертає `domain.ErrAlreadyRunning`.
- Продуктовий код не змінювався. Git не використовувався.

## C09 / Discussion

Тест `testDelayedOldClaim` навмисно має бути червоним на живому Postgres за поточним кодом; його не позначено skip і не замасковано.

Механізм дефекту:

- заміна відео в `backend/internal/media/adapters/postgres/repository.go:132-146` soft-evict-ить старий рядок (`evicted_at`), а нове відео вставляється окремим рядком із наступною `generation`; старий рядок лишається до sweeper DELETE;
- `backend/internal/transcript/adapters/postgres/draft.go:101-105` у `Claim` перевіряє лише `lesson_id`, `asset_id` і `generation`, але не `evicted_at` і не те, що це поточна ready-генерація;
- після цього `draft.go:116-125` безумовно перезаписує рядок `asr_drafts` через `ON CONFLICT (lesson_id)`, а `Complete` пише результат за новим токеном (`draft.go:192-210`).

Тому запізніла задача старої генерації може пройти `Claim`, перевести чернетку нової генерації в `running`, а потім записати старий текст. Це не виникає в поточній реальній черзі завдяки `Concurrency: 1`, FIFO та постановці лише поточної генерації, але це емерджентна гарантія порядку, а не інваріант репозиторію. Кандидат виправлення для окремого рішення рецензента: обмежити `Claim` поточною ready-генерацією уроку або додати до UPSERT умову на кшталт `WHERE asr_drafts.generation <= EXCLUDED.generation`.

Обсяг не розширювався до продуктового фікса: за брифом C09 залишено як доказ для рішення Courses 2.

## Перевірки

Пройшли:

- `gofmt` для двох змінених integration-файлів;
- `GOCACHE=/tmp/courses-gocache go vet ./internal/transcript/...`;
- `GOCACHE=/tmp/courses-gocache go test ./internal/transcript/...`;
- `GOCACHE=/tmp/courses-gocache go test -tags integration -run '^$' ./internal/transcript/...` — integration-тести скомпільовані.

Не запускалися:

- самі integration-тести з Docker (Postgres/Redis testcontainers): у робочому середовищі Docker недоступний за `work/codex-working-model.md`; їх має запустити Courses 2;
- повний ordered gate, e2e та інші перевірки поза брифом.

Отже C01 наразі доведено частинами — реальний Postgres-селектор і реальна Redis/asynq-ідемпотентність — без одного наскрізного тесту `EnqueuePending` через обидва контейнери.
