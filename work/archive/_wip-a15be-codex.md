ДАЛІ: Courses 1 має перегенерувати sqlc; тоді `mapResolved` мусить замінити тимчасове `false` на `row.TotpEnabled`.

2026-10-05 — A15-be виконано: resolve SQL вибирає `users.totp_enabled`; `domain.User`, HTTP DTO/мапінг та OpenAPI `AuthenticatedUser` несуть `twoFactorEnabled`. Generated-файли не змінювалися, sqlc/orval не запускались. До regen resolve тимчасово віддає `twoFactorEnabled: false`.
