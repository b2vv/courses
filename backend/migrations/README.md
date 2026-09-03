# Migrations

`00001_courses.sql` creates the Course schema, constraints and published-list indexes.
The migration runner applies pending Goose SQL migrations; an empty directory remains a successful no-op.
Demo data lives in `backend/sql/seed.sql` and is applied explicitly with `make seed`, never by migration.

Destructive `DownTo(0)` verification runs only inside disposable Testcontainers databases. Retained development/production volumes use reviewed forward migrations.
