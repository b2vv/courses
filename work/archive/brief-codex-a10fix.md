# Codex brief — A10-fix: verified-subject у integration-фікстурах (E1 після A05-гейту)

Гілка feat/accounts-selfservice. Повний `make check` на стадії integration впав: гейт верифікації з A05
(`access.CanStream` для playback/transcript і `email_not_verified` для мутацій orders/discussion) тепер
правильно відмовляє НЕверифікованим subject, але integration-фікстури будують subject/actor БЕЗ
`EmailVerified=true`, бо писались до гейту.
РОБОЧИЙ ФАЙЛ: work/_wip-a10fix-codex.md (append, перший рядок `ДАЛІ:`).

Падають (email_not_verified):
- `internal/discussion/adapters/postgres/repository_integration_test.go` (~:450)
- `internal/orders/adapters/postgres/repository_integration_test.go` (~:100/164/203/248/323)
- `internal/playback/application/sessions_integration_test.go` (~:142/162/201/242/254/272/291/312/329)
- `internal/transcript/adapters/postgres/repository_integration_test.go` (~:212/240/250/272/290) та
  `transcript/adapters/postgres/write_integration_test.go` (~:50/126)

Зроби: у КОЖНОМУ з цих тестів скрізь, де будується `access.Subject` (чи actor/viewer, що мапиться в Subject),
для сценаріїв, які МАЮТЬ пройти, постав `EmailVerified: true`. Тести, що перевіряють відмову за ІНШОЮ причиною
(право/ownership/entitlement/preview), теж мають verified-subject — інакше відмова підміняється на
email_not_verified і змінює сенс тесту. Якщо є спільний хелпер побудови subject — онови його (найменша зміна).

Правила: міняй ЛИШЕ ці integration-тести (і їхні локальні хелпери); НЕ чіпай продакшн-код, generated, sqlc, інші
тести. Зберігай оригінальні перевірки/очікування (право, стейл-токен, ownership). НЕ комість.

Критерій: `go build ./...` ок; (Courses 1 прожене `make integration`/повний гейт — ці 4 пакети мають стати зелені).
Познач у робочому файлі, чи був спільний хелпер, чи правив поштучно.
