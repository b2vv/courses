ДАЛІ: L5-backend реалізовано; фактичний Docker integration-прогін лишається для Courses 1.

# _wip L5 (Codex) — person-reports

Рішення: resolve неіснуючої або вже resolved скарги — idempotent no-op; дублі ловимо `INSERT … ON CONFLICT … DO NOTHING` + `RowsAffected()==0`; integration fixtures мають новий UUID-курс і users, не `d1`/`b1`.

Зроблено: domain `ParticipantReport`/коди/валідація і participant predicate; порт із точним `ReportSubject(ownerID, membership, courseFound)`; postgres-власний SQL по `courses` + `course_grants`, dedup, keyset admin queue лише з display_name/title, resolve; application 4 операції з email-first і authz; domain/application/integration тести. HTTP і generated/sqlc не змінювались.

Перевірено: `go test ./internal/discussion/...`, compile інтеграційного тесту (`-tags=integration -run '^$'`), `go test ./platform/architecture/...`, `go build ./internal/discussion/...`, `git diff --check` — green. Фактичний integration test не стартував: Docker socket у sandbox permission denied до запуску Postgres.
