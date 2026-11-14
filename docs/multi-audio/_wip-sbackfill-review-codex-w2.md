Reviewed `git diff origin/main...HEAD` at `16fa3d009bf1e68a3f5376db0906bc1a48270b97`. Source inspection only: no files changed, tests or database commands executed.

### 1. PostgreSQL SQL, transaction and locking

Opened `backend/migrations/00028_audio_track_embedded_backfill.sql:12`, both DO blocks at `:17` and `:67`, and the cached goose implementation, `provider_run.go:213`.

Both DO blocks have `StatementBegin/End`. There is no `NO TRANSACTION`; goose executes migration SQL and its version update in one transaction. `SET LOCAL lock_timeout` precedes `LOCK`.

Traced Publish (`publish.go:77`), Mint (`audiotrack.go:625`), Promote (`:395`), Delete (`:237`), metadata (`:273`), eviction (`publish.go:114`), audio sweep (`audiotrack.go:808`) and video sweep (`publish.go:161`). Their DML takes `ROW EXCLUSIVE`, conflicting with the migration’s `SHARE ROW EXCLUSIVE`. Preliminary row-locking reads remain compatible; a conflicting interleaving can abort/retry rather than silently bypass serialization.

Up `:46` and Down `:93` use identical canonical UUID text and MD5 input. PostgreSQL accepts the resulting 32 hexadecimal digits as UUID without requiring UUID version bits. Down additionally requires `kind='embedded'`, preserving ordinary Publish rows.

**Result:** no SQL or rollback defect found.

### 2. Constraints and language selection

Opened `00025_multi_audio.sql:10`, `00026_audio_track_language_unique.sql:6` and `00027_audio_track_upload_columns.sql:3`.

Checked inserted kind, state, source leaf, lengths, FK, timestamps and byte-column defaults. They satisfy the schema. Crucially, “no Ready track” also implies “no default”: 00025 requires every default to be Ready. Insertion therefore cannot collide with the one-default index.

Tried occupied `und` in queued/processing states, failed language holders, and both fallback languages occupied. The occupancy predicates match 00026’s `state <> 'failed'` exactly: fallback succeeds where available; double occupancy deliberately refuses atomically.

**Result:** no reachable-state constraint violation found.

### 3. D-B3 gate

Opened migration `:72`, `publish.go:70`, `audiotrack.go:362` and `:210`.

Tried these counterexamples:

- **Mid-promote/default swap:** Ready state and default changes commit together; the migration cannot inspect the uncommitted intermediate state.
- **Evicted Ready generations:** explicitly excluded.
- **`awaiting_audio`:** explicitly excluded.
- **Ready uploaded default without embedded:** accepted, as D-B6 requires.
- **Ready uploaded track without a default:** skipped by insertion, then rejected by D-B3.

The schema already guarantees default ⇒ Ready and at most one default; the gate supplies the missing existence check.

**Result:** no false refusal or missed invariant violation found within D-B3’s defined scope.

### 4. Every downgrade below 28

An exhaustive `rg 'DownTo|UpTo' backend --glob '*test.go'` found the following. Paths below are relative to `backend/internal/`.

| Test and file:line | Versions below 28 | Gate/index assessment |
|---|---|---|
| Course authoring backfill/rollback — `course/adapters/postgres/repository_integration_test.go:82` | Down 0/2; Up 2 | No video data. Safe. |
| Course media migration — same file `:196` | Down 4; Up 5 | Seeds users/courses/modules/lessons only. Safe. |
| Course playback migration — same file `:278` | Down 5; Up 6 | Video inserted at `:284` is **queued**, not Ready. Safe. |
| Course enrollment backfill — same file `:344` | Down 14; Up 15 | Retains queued video; adds no Ready assets/audio. Safe. |
| Course lesson-document migration — same file `:425` | Down 17/18 | Retains queued video; document fixtures add no media. Safe. |
| Course playback lock-timeout/retry — same file `:517` | Down 5; Up 6 | Retains queued video. Expected timeout concerns migration 6; subsequent full Up is safe. |
| Course failed-migration rollback — same file `:567` | Down 6 | Synthetic migration 7 deliberately fails; real full Up encounters only queued video. Safe. |
| Media schema reversibility — `media/adapters/postgres/audiotrack_integration_test.go:593` | Down 24 | Previous test leaves Ready asset plus Created upload. Down drops `audio_tracks`; Up recreates index empty, then backfills embedded/default. Safe. |
| Identity reversibility — `identity/adapters/postgres/schema_integration_test.go:618` | Down 0 | Entire schema dropped; subsequent Up has no media. Safe. |
| New populated backfill — `media/adapters/postgres/backfill_integration_test.go:108` | Helper Down 27; Down 27 at `:182` | Positive fixtures and fallback succeed by inspection. |
| New language refusal — same file `:218` | Helper Down 27 | Deliberate double-language rejection. |
| New broken-default refusal — same file `:237` | Helper Down 27 | Deliberate D-B3 rejection. |
| New writer timeout — same file `:255` | Helper Down 27 | Deliberate timeout, followed by successful retry. |

Every course downgrade drops `audio_tracks`, so migration 26 rebuilds its language index on an empty table.

The named other packages have **no DownTo/UpTo calls**. Their initial full Up precedes fixtures: transcript repository `:146`, transcript draft `:87`, playback sessions `:88`, playback access `:61`, playback progress `:58`, adminstats snapshot `:51`. Their later Ready fixtures cannot trip this one-time migration.

**Result:** no cross-package migration regression found.

### 5. Production path, fixtures, timeout and cleanup

Opened `backfill_integration_test.go:44`, `:61`, `:123`, `:176`, `:237`, `:255`; registration/harness in `repository_integration_test.go:134` and `:203`.

Tests use the real migration directory, goose provider, pgx and disposable PostgreSQL **17.5**. Fixtures cover the planned states, language fallback, refusal, rollback and subsequent Promote behavior.

The timeout test completes its uncommitted INSERT before invoking goose (`:265–272`). No goroutine scheduling or fixed sleep establishes the lock. Rollback releases it before retry.

Cleanup ordering is sound for these scenarios: writer rollback runs before subtest cleanup; cleanup resets populated data before restoring head; provider connection cleanup runs afterward. Assertions establish core row fields, counts, refusal and rollback; they do not constitute a complete unchanged-row snapshot comparison.

**Result:** appropriate real-path tests; execution remains **NOT RUN** in this review.

### Findings

- **Minor — promised operational diagnostics are absent.** `docs/multi-audio/plan-sbackfill.md:82` requires read-only diagnostic queries in the report for language refusal, broken defaults and lock timeout. `docs/multi-audio/report.md` has no S-backfill diagnostics, while `backend/cmd/migrate/main.go:67` deliberately replaces the underlying error with a generic message. Add the promised queries so operators can distinguish these failures.

No verified Blocker or Major findings.

VERDICT: SHIP