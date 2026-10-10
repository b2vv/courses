# P03 report — одна таблиця доступу через SQL кожного адаптера

## Done

- Додано `backend/internal/contracts/access/accesstest` із експортованою таблицею `TruthTable`: 9 рядків `{administrator, owner, stranger} × {draft, published, archived}` з нейтральними `Role`, `Status` і `Want`.
- `access_test.go` та `platform/architecture/shared_rules_test.go` переведено на спільну таблицю. Окремий рядок порожнього суб’єкта в unit-тесті збережено.
- Додано SQL integration-підтести в наявні harness-и:
  - `R06` — `course/adapters/postgres/learning_integration_test.go`, через `GetLearningCourse`;
  - `R07` — `playback/adapters/postgres/access_integration_test.go`, через `LoadVideoCourse` і `LoadLessonPlayback`;
  - `R08` — `transcript/adapters/postgres/repository_integration_test.go`, через `GetLessonTranscript`.
- Кожен SQL-підтест засіває окремі курс, модуль, урок і, для playback, ready-asset; кожен проганяє всі 9 рядків і перевіряє `CanBeViewedBy(actor) == want`. Архівні рядки покриті тією самою таблицею: адмін і власник — доступні, сторонній — ні.
- Додаткових адаптерів, які читають факти для `CanBeViewedBy`, не знайдено.
- Продуктовий код, SQL адаптерів, домени, `docs/`, `work/tasks.md` і git не змінювалися.

## Verification

- `gofmt` — успішно.
- `go vet ./...` — успішно.
- `go test ./...` — не пройшов повністю через обмеження sandbox: `courses/platform/httpserver.TestShutdownDrainsAnInFlightResponse` не зміг відкрити `127.0.0.1:0` (`operation not permitted`). Інші пакети цього прогону пройшли.
- `go test -tags integration -run '^$' ./...` — успішно: integration-тести компілюються.
- Самі integration-тести не запускалися: середовище Codex не має Docker, тому PostgreSQL testcontainers недоступні. Це не вважається pass; їх має прогнати reviewing-сесія.

## Discussion

Немає. Під час перевірки всі чотири перелічені в брифі адаптери дали змогу прочитати факти й застосувати очікування спільної таблиці; розбіжностей вердикту не виявлено.

## Правка після рев'ю

- У `R07` playback integration-підтест тепер використовує власні `tableMetric` і `tableReader` для 9 викликів `LoadVideoCourse`. Наявний `reader` залишено для початкового сценарію та його перевірки `metric=2`; очікування не змінено.
