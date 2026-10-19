# Codex brief — A13b: postgres TwoFactorStore + RecoveryCodeStore (на generated)

Репо `courses`, гілка `feat/accounts-selfservice`. A13a у дереві (порти TwoFactorStore/RecoveryCodeStore/
PendingTwoFactorStore, application 2FA, redis pending). Courses 1 прогнав `sqlc generate`; у
`internal/identity/adapters/postgres/generated` є: `GetUserTwoFactor`, `SetUserTotpSecret`, `EnableUserTotp`,
`DisableUserTotp`, `InsertRecoveryCode` (однорядковий), `ConsumeRecoveryCode` (атомарний, RETURNING id),
`DeleteRecoveryCodesForUser`. Роль: реалізуєш A13b; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a13b-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити (у `internal/identity/adapters/postgres`, патерн наявного `repository.go`)
1. **`ports.TwoFactorStore`**: `GetTwoFactor` (мапить `GetUserTwoFactor`-row → доменні факти: secret, enabled),
   `SetTotpSecret`, `EnableTotp`, `DisableTotp` на відповідних generated-функціях. Явний маппінг; помилки обгортай.
2. **`ports.RecoveryCodeStore`**:
   - `InsertCodes(ctx, userID, digests [][]byte|[]string)` — **у транзакції** зацикли `InsertRecoveryCode`
     (все-або-нічого; id генерує адаптер). 
   - `ConsumeCode(ctx, userID, digest)` → `ConsumeRecoveryCode`; 0 рядків (`pgx.ErrNoRows`) → «невірний/вжитий»
     (булеве false або доменна помилка, як очікує application-порт з A13a).
   - `DeleteCodes(ctx, userID)` → `DeleteRecoveryCodesForUser`.
3. **Compile-time assertions**, що `Repository` реалізує обидва нові порти (як зроблено для UserWriter/TokenStore).
4. **Integration-тести** (`*_integration_test.go`, тег як наявні, через `testsupport/containers`):
   - Totp: Get(default false/nil) → SetTotpSecret → Get бачить secret, enabled=false → EnableTotp → enabled=true →
     DisableTotp → secret NULL, enabled=false.
   - Recovery: InsertCodes(10) → ConsumeCode(дійсний) успіх один раз → повторний Consume того ж → провал →
     DeleteCodes прибирає решту; невідомий код → провал.

## Правила
- generated у адаптері, явний маппінг у домен; домен не імпортує адаптер. Без ORM. Без секретів/кодів у логах.
- НЕ запускай sqlc (вже згенеровано), generated не редагуй вручну. НЕ комість. TDD/integration; час — спершу тести.

## Критерій готовності
- `go build ./...`, `go vet ./internal/identity/...`, `gofmt` чисті.
- `go test ./internal/identity/...` unit зелені; integration postgres 2FA проходить під `make integration`-харнесом
  (Courses 1 прожене).
- `Repository` компіляційно реалізує TwoFactorStore і RecoveryCodeStore.
