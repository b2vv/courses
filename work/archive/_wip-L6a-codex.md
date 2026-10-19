# _wip L6a (Codex) — orders HTTP endpoints

(Codex дописує тут після кожного кроку.)

ДАЛІ: L6a реалізовано; перед інтеграцією з іншими L6-зрізами прогнати повний репозиторний гейт.

2026-10-07 — Додано тільки orders HTTP: шість маршрутів enroll/membership/roster/block/unblock/revoke, `RegisterTeach` і його виклик у `cmd/api/main.go`. Хендлери лише дістають actor, перевіряють transport input і викликають application; authz не перенесено з application. Додано DTO `MembershipGrant`, `MembershipState`, `RosterPage` (у roster немає email), нові OpenAPI paths/schemas та `membership_conflict`, `course_not_enrollable`, `enrollment_forbidden` до error enum; також внесено вже відповідний HTTP-коду `email_not_verified`, який contract-тест виявив відсутнім. `writeError`: conflict для перших двох і forbidden для enrollment. Contract-validated tests покривають happy/error шістьох маршрутів, включно з roster без email і bad cursor. `go test ./internal/orders/...`, `go build ./...` та `git diff --check` зелені з ізольованим GOCACHE. Orval/sqlc не генерувалися; discussion L6b не чіпався; коміту немає.
