ДАЛІ: Courses 1 має запустити integration там, де Docker доступний, і забрати зміни в свій коміт; browser e2e навмисно не запускався.

# _wip L4b (Codex) — communication gate (discussion)

2026-10-07 — L4b-backend виконано в `discussion`:

- Домен: `Course.CanCommunicateBy`/`CanConsumeBy`; нові коди `enrollment_required` та `enrollment_blocked`. `CanBeViewedBy` не змінювався.
- Application: `Publish`, `ReportComment`, `PutReview`, `ReportReview` ідуть через `requireCommunication` після email-first; `PutReview` перевіряє комунікацію до `CanReview`. `ReportReview` спершу читає відгук, потім вантажить `ReviewFacts` за його `course_id`. `List` перейшов на `CanConsumeBy`. `PublicReviews` не чіпав.
- HTTP: обидва коди = 403 із повідомленнями; додано до `backend/openapi/openapi.yaml`, generated/sqlc не чіпав.
- Unit: table-driven precedence: unverified+blocked → email, blocked → blocked, none/revoked → required, active → success; List (none/blocked/active); ReportComment; PutReview (communication до CanReview); ReportReview завантажує курс.
- Integration: доданий test із реальним repository та окремими direct-seeded `active`/`blocked`/`revoked`: active Publish, blocked/revoked Publish, blocked/revoked List.

Перевірено: `go test ./internal/discussion/... ./internal/contracts/...`, `go test ./platform/architecture`, `go build ./internal/discussion/...` — green (із `GOCACHE=/private/tmp/courses-l4b-gocache`). Integration-команда скомпілювалася, але запуск зупинився до тесту: поточне середовище не має доступу до Docker socket (`permission denied`). Browser e2e не запускався за scope.
