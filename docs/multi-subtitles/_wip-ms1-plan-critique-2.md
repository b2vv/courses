`stage: plan` · Critic: Codex `/root`, selected independent context · Round 2.

1. **BLOCKER — OPEN: M-R12 lacks a repeatable startup precondition.**

   Location: [plan-ms1.md:129](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:129), related claim at line 108.

   **New counterexample:** run `make up` successfully, then run it again without stopping the stack or changing its configuration. The existing `media-worker` can remain running: there is no second worker-start event against which to prove “`media-masters` completed before worker startup.” A dependency condition orders startup; it does not itself stop an already-running consumer.

   **Required change:** distinguish:

   - Two starts with **retained volumes and stopped containers**, proving regeneration completes before each worker startup.
   - An `up` against an already-running stack, with its intended behavior explicitly stated.

   Include a failed-regeneration startup case and assert nonzero Compose exit and no worker startup. Capture events before invoking `up`; `ps -a` alone cannot establish ordering. This concerns the new M-R12 acceptance, not the previously closed dependency-cycle finding.

2. **NONBLOCKING — strengthen M-R9’s error-chain test.**

   Location: [plan-ms1.md:126](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:126).

   Test a **wrapped** `PgError` through the production diagnostic/sanitization function, asserting the emitted fields and continued redaction of other errors. A bare `PgError` test could pass while extraction from Goose’s wrapper fails. The installed Goose source provides `PartialError.Unwrap()`, so this route is feasible.

Explicit dispositions of round-1 findings:

| Finding | Disposition | Counterexample retried against r2 |
|---|---|---|
| **1 — preflight ordering** | **Resolved** | Lines 65–69 reject both `und-` and `und`/`UND` before relabel. M-R4 explicitly requires every original row unchanged after refusal. |
| **2 — migration diagnostics** | **Resolved at planning level** | Lines 82–89 now name a concrete `PgError.Message` logging route and separate retained-volume diagnostics. The original generic-error-only path is expressly changed. Finding 2 above strengthens its planned proof. |
| **3 — populated regeneration path** | **Resolved at planning level** | M-R7 now connects populated v29 data, migration, real adapters, real master bytes and a second identical rewrite. The wrong-root counterexample should fail this test. Dispatcher/failure tests and Compose evidence complete the intended process-level coverage. |
| **4 — shared grammar** | **Resolved** | M-D1 preserves transcript normalization while media rejects surrounding whitespace. The original `" uk "` counterexample is now explicit acceptance. |
| **5 — HTTP and picker coverage** | **Resolved** | M-R10 covers mint/rename rejection and unchanged metadata; M-R11 names `availableAudioLanguages`, case-insensitive exclusion and preservation of the edited track’s own option. |

Specific checks:

- **Transcript behavior/error codes:** moving the existing predicate preserves behavior if `NewLanguage` retains its UTF-8 check, trimming, lowercasing and `ErrInvalidRequest` return. No transcript read or HTTP contract change is required.
- **Architecture:** [contracts_test.go:13](/Users/strelia/projects/courses/backend/platform/architecture/contracts_test.go:13) recursively inspects new contract packages; there is no package allowlist. Standard-library-only `contracts/language` fits it. Domain imports are allowed by [boundaries_test.go:99](/Users/strelia/projects/courses/backend/platform/architecture/boundaries_test.go:99).
- **P0001 safety:** repository-wide migration inspection found only the two existing raises in migration 00028; both interpolate UUIDs. No existing migration-defined trigger/function supplies a user-text counterexample. Logging specifically `PgError.Message` is safe for those messages and the proposed ID/count messages at this snapshot. **P0001 itself is not an authorship guarantee.**
- **Integration placement:** `cmd/media-worker` is compatible with the runner’s `go test -count=1 -tags=integration ./...`. Use `//go:build integration`, the existing `containers.Start` helper, isolated Postgres and temporary filesystem storage. The package-local media harness is not directly reusable.

**BLOCKED — MS1 r2. One substantive blocker remains: finding 1 above.**

Binding and limitations:

- Root: `/Users/strelia/projects/courses`; branch `multi-subtitles`; HEAD `895b2e2537dd2ad428239254444666d67c62c1b3`.
- Untracked plan, r2: SHA-256 `07b794f5bea6cfcced6a0878c29cc36eb4ef02eab97605c2620eb818979fa77b`.
- Phase-plan SHA-256: `cbe299106678907bbf4494196e182ff0c16d37e000e0ab84c267e2ebd3c59f43`.
- Raw round-1 record inspected: SHA-256 `fd2ffc8e8e9db2d946bede36030d011acb9566b4a8b1d3e11d037c23be57f6fc`.
- Prompt: candidate release directory `17a9527421e8fac3e43a16c287c404118699833c`, `prompts/plan-critic.md`, version 1, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package declares `0.8.0-alpha.1+native-handoff.1`; release commit remains unverified.
- Relevant tracked product sources are clean. Unrelated dirty documentation/configuration and untracked `backend/upload` were excluded.
- Earlier closed findings remain closed. This round freshly assessed the changed contracts, diagnostics and verification strategy; no historical test PASS transfers.
- No tests, writes, network probes, graph refresh or reviewer spawning ran. Four-lens gate remains **NOT_RUN: 0 launched / 0 returned**.
- Record returned here because the workspace is read-only; no review artifact was saved.