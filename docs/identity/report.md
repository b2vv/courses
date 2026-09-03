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

## I02 — credential and session application slice

1. **Implemented.** Added login, authenticate and logout use cases; real Argon2id verification; sqlc Identity queries; and a PostgreSQL session adapter with opaque tokens, digest-only persistence, expiry cleanup, revocation and bounded collision retry.
2. **Files and modules.** `backend/internal/identity/application` owns orchestration and the 12-hour lifetime. `ports/authentication.go` defines only the credential and session dependencies used by those cases. `adapters/argon2id` owns bounded PHC parsing and constant-time digest comparison. `adapters/postgres/repository.go` owns credential lookup, user mapping and transactional session operations. Identity-generated sqlc code lives under the adapter.
3. **Architecture decisions.** Password hashes stay inside the credential adapter. A raw token crosses the session port only in an explicitly secret-bearing issuance result and never enters domain/generated persistence models. Login rotation confirms one transaction before returning a token; unique collisions retry the full transaction up to three times. An indeterminate commit fails closed without a cookie claim. Application and adapter boundaries both validate untrusted tokens.
4. **Verification commands.** TDD red/green runs covered application and Argon2id packages. The final backend gate ran `gofmt`, `go vet ./...`, `go test -race ./...`, and `go test -tags=integration ./internal/course/adapters/postgres ./internal/identity/adapters/postgres` in the pinned Go container.
5. **Passed checks.** All backend vet and race tests pass. Testcontainers verifies real Argon2id login, exactly one dummy verification for unknown users, disabled-user rejection, same-account/cross-account rotation, digest-only storage, one-collision recovery, three-collision rollback, concurrent distinct sessions, exact expiry cleanup, idempotent revoke, cancellation rollback and recovery. The Course integration suite remains green.
6. **Remaining limitations.** No HTTP/cookie/origin contract exists yet. Demo-user seed values are deferred until I06, where documented credentials and repeat/concurrency evidence are delivered together. A connection loss during PostgreSQL commit remains an intentionally reported indeterminate 503 outcome.
7. **Next smallest task.** I03 adds typed runtime security configuration, the three HTTP handlers, same-origin enforcement, cookie handling and authentication middleware without adding a role-restricted business endpoint.
8. **User decision.** No product decision blocks I03. GitHub remote creation/push still requires explicit authorization for the exact external destination.
