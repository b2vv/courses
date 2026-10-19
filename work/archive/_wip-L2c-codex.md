# _wip L2c (Codex) — course roster with progress

(Codex дописує тут після кожного кроку.)

ДАЛІ: L2c — додати read-model ростера в orders (порт, PostgreSQL, application та тести); HTTP/frontend/generated/sqlc поза межами.
Крок 1: контекст перевірено. Робоче дерево вже містить сторонні зміни; їх не чіпаю. Рішення: ростер включає лише `active` і `blocked`; `revoked` свідомо виключено. Прогрес буде розрахований SQL-виразом за `lesson_progress.completed_at`, без повторної реалізації порогу playback.

Крок 2: додано порт `Roster` і типи без email-поля, application-авторизацію та clamp (default 20, min 1, max 100), PostgreSQL read-query й RawURL base64 cursor. `ErrInvalidCursor` — новий у `orders/domain` (`roster_cursor_invalid`). Для `granted_at DESC, user_id ASC` наступна сторінка використовує рівносильний keyset-предикат: старіша дата або та сама дата й більший user id.

Крок 3: додано application unit з owner/admin/стороннім, clamp і пробросом cursor/page; усі наявні фейки порту доповнено `Roster`. Додано інтеграційний En08: active+blocked/revoked, 100/50/0%, нуль уроків, порожній ростер, limit=1/keyset/порожній останній cursor, битий cursor і authz. `gofmt`, `GOCACHE=/private/tmp/courses-l2c-gocache go test ./internal/orders/...`, `go build ./...`, `go vet ./internal/orders/...`, `git diff --check` — green. Docker integration за домовленістю лишається Courses 1.

Крок 4: локальний перегляд L2c-дифу завершено: у production SQL немає `users.email`; read-model не робить write; HTTP/frontend/generated/sqlc не змінено. Декодер також відхиляє base64-валідний cursor з не-UUID id як `ErrInvalidCursor`. ДАЛІ: Courses 1 запускає Docker integration; L2c готовий до передачі без коміту.
