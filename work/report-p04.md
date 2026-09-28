# P04 — звіт

## Зроблено

- Додано HTTP contract-тест `TestR09LearningForeignDraftDeniedAndPublishedAllowed` у `backend/internal/course/adapters/http/learning_handler_test.go`.
- Додано HTTP contract-тест `TestR09TranscriptForeignDraftDeniedAndPublishedAllowed` у `backend/internal/transcript/adapters/http/handler_test.go`.
- Додано окремі R09-тести для обох playback-точок у `backend/internal/playback/adapters/http/handler_test.go`:
  - створення сесії;
  - `authorize` для nginx `auth_request`.
- У кожному тесті перевірено пару рядків `accesstest`: `stranger × draft` відмовляється, `stranger × published` дозволяється.
- Факти ролі та статусу курсу мапляться через `internal/contracts/access/accesstest`; playback-тести проходять через application-сервіси, а не підміняють відмову помилкою фейка.

## Перевірки

- `gofmt` — PASS для трьох змінених тестових файлів.
- `GOCACHE=/private/tmp/courses-go-cache go vet ./...` — PASS.
- `GOCACHE=/private/tmp/courses-go-cache go test ./internal/...` — PASS.
- Початковий запуск без окремого `GOCACHE` був заблокований правами на системний Go build cache (`operation not permitted`); повторний запуск у дозволеному `/private/tmp` пройшов.

## Discussion

Немає. Продуктовий код, SQL, домени, контракти, фронтенд, `docs/`, `work/tasks.md` і git не змінювалися.
