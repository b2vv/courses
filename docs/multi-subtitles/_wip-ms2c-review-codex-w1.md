Blocker: none.  
Major: none.

Minor — inventory’s no-Ready boundary lacks a nonempty repository test. [inventory_integration_test.go:49](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/inventory_integration_test.go:49) checks zero transcripts, then creates a Ready asset before saving any transcript. Add an existing transcript with no Ready asset and assert `stale=false`; this exercises the `COALESCE(..., 0)` scan in [repository.go:186](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/repository.go:186). The implementation appears correct.

The remaining checks align with scope:

- Endpoint access, error distinctions, body strictness, ordering, stale computation, cache headers, removed-route 404s and OpenAPI agree.
- Media metadata comparison occurs under the asset lock with a `FOR UPDATE` pre-read; master rewriting depends only on language/label changes ([audiotrack.go:272](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:272)).
- Frontend validators accept and validate the new fields, validate every draft, select `drafts[0]`, and regenerate by track ID. No new UI fetch or `any` found.
- Existing ASR e2e URL changes are explicitly scheduled in M3-R9.

Reviewed staged, unstaged and relevant untracked files at `6d11e76`; staged diff was empty. No files changed. No tests or Docker run; runtime integration and full gates remain unverified here.

VERDICT: SHIP

## Resolution (Courses 1, 2026-11-15)
- Minor (no-Ready inventory): added subtest `M2-R10 inventory without any Ready asset…` in `inventory_integration_test.go`; `go test -tags=integration -run ^TestTranscriptReadAgainstPostgreSQL$` — both M2-R10 inventory subtests PASS.
