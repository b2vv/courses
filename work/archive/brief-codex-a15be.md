# Codex brief — A15-be: two_factor_enabled у /auth/me

Гілка feat/accounts-selfservice. Треба, щоб фронт бачив стан 2FA після reload.
РОБОЧИЙ ФАЙЛ: work/_wip-a15be-codex.md (append, перший рядок `ДАЛІ:`).

Зроби (backend identity):
1. Resolve-запит користувача-по-сесії у `sql/queries/identity.sql` — додай `totp_enabled` у SELECT
   (поряд із уже доданим email_verified). НЕ запускай sqlc (Courses 1).
2. `domain.User` += `TwoFactorEnabled bool`.
3. Mapper у `adapters/postgres/repository.go` (resolve) — наповнюй `TwoFactorEnabled` з нового стовпця
   (після того як Courses 1 перегенерує). Якщо generated ще не має поля — познач TODO й лиши компільовано.
4. `userDTO` (handler.go, біля EmailVerified) += `TwoFactorEnabled bool json:"twoFactorEnabled"`; мапінг у
   рядку toDTO (~787) += TwoFactorEnabled.
5. `openapi.yaml` `AuthenticatedUser` += `twoFactorEnabled: boolean` (required).

Правила: generated не чіпай, sqlc/orval не запускай, НЕ комість. `go build`/`gofmt` мають лишатись зеленими
(якщо generated ще без поля — мапер тимчасово не читає його, але код компілюється; Courses 1 завершить після regen).
Критерій: openapi оновлено; domain/DTO/handler готові; чітко напиши в робочому файлі, що мапер чекає regen.
