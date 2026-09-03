# Identity Phase 3 delivery report

Updated: 2026-09-03. Current completed task: I01.

## I01 — users, sessions and domain

1. **Implemented.** Added the Identity domain foundation for users, roles and time-bounded sessions, plus the PostgreSQL schema for credential records and opaque-session digests.
2. **Files and modules.** `backend/internal/identity/domain` owns stable errors, `User`, `Role`, `Session`, email normalization and capability policy. `backend/migrations/00002_identity.sql` owns `users` and `user_sessions`. `backend/internal/identity/adapters/postgres/schema_integration_test.go` verifies the schema against PostgreSQL 17.5.
3. **Architecture decisions.** Domain objects contain no password hashes, raw session tokens, pgx types or transport DTOs. Persistence stores an Argon2id PHC string beside the user record and only a unique 32-byte session-token digest. Session lifetime uses the half-open interval `created_at <= now < expires_at`. `student` can learn; `instructor` and `admin` can enter the administration shell.
4. **Verification commands.** Targeted TDD red/green runs used the pinned `golang:1.27.1-bookworm` toolchain. The final I01 gate ran `gofmt` cleanliness, `go vet ./internal/identity/domain`, `go test -race ./internal/identity/domain`, and `go test -tags=integration ./internal/course/adapters/postgres ./internal/identity/adapters/postgres`.
5. **Passed checks.** Domain tests pass with the race detector. Migration up/down/up, database constraints, unique digests and user-to-session cascade pass in Testcontainers. The existing Course repository integration suite also passes with migration 00002 present.
6. **Remaining limitations.** I01 does not authenticate credentials, create/revoke sessions, expose HTTP endpoints, seed demo identities or render authenticated frontend state. Those behaviors belong to I02–I06.
7. **Next smallest task.** I02 adds only the application ports/use cases and Argon2id/PostgreSQL adapters, with fake-port tests and transactional rotation integration evidence.
8. **User decision.** No product decision blocks I02. The GitHub remote remains unconfigured because external repository creation/push needs explicit approval for the exact destination.
