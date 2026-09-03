# Migrations

Phase 1 has no domain schema. `migrate` reports a successful no-op for this empty directory.
Phase 2 adds versioned Goose SQL migrations. Demo data is an explicit seed command, never a schema migration.
