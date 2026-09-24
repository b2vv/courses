# Report A13 / C09

## Зроблено

- У `backend/internal/transcript/adapters/postgres/draft.go:118-138` додано UPSERT-guard `WHERE asr_drafts.generation <= EXCLUDED.generation`.
- Поруч із guard зафіксовано інваріант: він утримується в місці запису й не залежить від порядку черги (Concurrency 1, FIFO).
- `pgx.ErrNoRows` після відсікання guard повертає `ports.ClaimedDraft{Claimed: false}` без помилки; гілка стоїть перед `persistenceError`. Обробка FK `23503 → ErrNotFound` збережена.
- У `draft_integration_test.go:261-304` оновлено сценарій запізнілого claim: він перевіряє `Claimed: false` і збереження готової чернетки генерації 2.
- Додано сценарій повторного `Claim` тієї самої генерації; він перевіряє успішний claim і новий optimistic-lock token.

## Перевірки

- `gofmt` — пройшов; `gofmt -d` не показав відмінностей.
- `go vet ./internal/transcript/...` — пройшов.
- `go test ./internal/transcript/...` — пройшов.
- `go test -tags integration -run '^$' ./internal/transcript/...` — пройшов, integration-пакети скомпільовано без запуску тестів.
- Реальні integration-тести не запускалися: для них потрібен Docker, а working model забороняє цей запуск у Codex-сесії; їх має прогнати reviewing session / Courses 2.

Зміни обмежені `backend/internal/transcript/adapters/postgres`; цей звіт записано у вказаний брифом шлях.
