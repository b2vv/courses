# T01 report — migration 00004, query gate, canary and identity seed

## Files

- `backend/migrations/00004_identity_admin.sql` — adds nullable `users.deleted_at`,
  the timestamp and deleted-implies-disabled CHECKs, replaces `users_email_key` with
  the partial `users_email_active_key`, and provides the test-only destructive Down.
- `backend/sql/queries/identity.sql` — gates `GetIdentityCredential` on
  `deleted_at IS NULL`; no other status query changed.
- `backend/sql/seed-identities.sql` — resets `deleted_at` and adds the three
  disposable student fixtures on ids ending `004`–`006`.
- `backend/internal/course/adapters/postgres/repository_integration_test.go` — proves
  migration 00004 preserves five populated courses and three pre-00004 identities,
  adds the real 00004 to the failing-migration MapFS, and renumbers the canary to
  00005.
- `backend/internal/identity/adapters/postgres/schema_integration_test.go` — proves
  the new schema shape and constraints, seed reconciliation of deleted rows, and an
  HTTP login after email reuse; its reuse pair is deleted before the terminal
  `DownTo`.
- `work/report-t01.md` — this report.

No file under a `generated/` directory was written.

## Decisions

- Used the plan's exact filename, `00004_identity_admin.sql`; the longer filename in
  the brief points back to the plan for resolution.
- Used the brief's stricter existing timestamp idiom,
  `deleted_at BETWEEN created_at AND updated_at`, while retaining the plan's named
  constraint `users_deleted_at_valid`.
- Assigned the three separately named disposable fixtures the next stable identity
  ids (`...004`, `...005`, `...006`); plan sections 2/5 name the emails and separation
  requirement but do not enumerate ids.
- The course migration harness uses an inline three-row pre-administration seed while
  held at migration 00002. The current seed now references `deleted_at` and therefore
  cannot execute before migration 00004 exists.
- B08 exercises the actual login handler with a deleted/active email-reuse pair and
  removes both rows in a subtest defer, before the file's terminal migration Down.

## Verified

- Read through the Up/Down order against migrations 00002 and 00003: the Up adds the
  column and CHECKs before replacing uniqueness; the Down removes the partial index
  and soft-delete schema before restoring table-wide email uniqueness.
- Read through all T01 query and seed edits: only `GetIdentityCredential` gained a
  status gate; seed conflict reconciliation clears both status timestamps.
- Read through B06/B07/B08 placement and teardown against their live harness order.
- Ran `gofmt` on the two edited Go test files.
- Ran `make -n` from the repository root: exit 0; the default target expanded only
  the `.env` existence/copy recipe and executed nothing.

## Checks not run

The brief denies Docker, sqlc, Go compilation and Testcontainers, and permits only
read-through plus `make -n`. Consequently none of the ordered gates was represented
as passed:

- `make format-check` — not run; its Go half requires Docker.
- `make lint` — not run; its Go vet half requires Docker/Go compilation.
- `make typecheck` — not run under the brief's verification limit.
- `make test` — not run; its Go half compiles in Docker.
- `make integration` — not run; denied Testcontainers/Docker. B06, B07 and B08 remain
  written evidence awaiting the reviewing environment.
- `make check-generated` — not run; sqlc is denied, generated writes belong to the
  reviewer, and the target also invokes Git, which this task forbids touching.
- `make build` — not run; Docker/compilation denied.
- `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e` — not run;
  they require the Compose environment.
- `make check` — not run because it includes every denied step above.

No write was denied, so the brief's stop-on-denied-write condition did not trigger.

## Reviewer verification (measured in the reviewing session)

`make generate` regenerated `GetIdentityCredential` with the `deleted_at IS NULL` filter.
One review fix: the three disposable fixtures collided on `users_pkey` with the authoring
integration test's own fixtures (`...004`=other-instructor, `...005`=disabled-owner) — the
disposables were relocated to the free `...00d1/d2/d3` block, and the identity seed test's
fixture ids synced to match. After the fix: `make integration` 12 packages ok (course
authoring and identity schema/seed suites green, including B06/B07/B08), format-check,
lint, typecheck, test and check-generated all pass. Provenance: implemented with Codex
(gpt-5.6-sol); reviewed, corrected and verified here.
