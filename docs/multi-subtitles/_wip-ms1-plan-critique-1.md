`stage: plan` · Critic: Codex `/root`, selected independent context. Scope: MS1’s implementation contract and production-path evidence; O1–O6 retained.

1. **BLOCKER — OPEN: D8 can erase data that D2 requires the migration to reject.**

   Location: [plan-ms1.md:63](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:63).

   **Counterexamples traced:**

   - Ready asset, default track `und-`, primary transcript `uk`. The old domain accepts `und-` ([audiotrack.go:267](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:267)). Step 1 matches `und-%` and changes it to `uk`; step 2 then finds no invalid language.
   - Ready asset with `und` and `UND`, primary `uk`. Both are legal under the old case-sensitive index. Step 1 elects one and changes it to `uk`, removing the normalized collision before step 3 detects it.

   These outcomes contradict D2’s refusal of invalid rows and normalized collisions ([phase plan:91](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:91)). The claim that step 1 creates no invalid output does not establish that it preserves refusal of invalid **input**.

   **Required plan change:** preflight the original rows under the migration locks before relabelling. Add both counterexamples to M-R4 and assert rollback leaves every original row unchanged.

2. **BLOCKER — OPEN: migration diagnostics do not reach the promised production remediation path.**

   Locations: [plan-ms1.md:67](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:67), [plan-ms1.md:124](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:124).

   **Counterexample traced:** retained data contains language `e`; migration raises the planned exception with its ID. The actual migration executable discards `provider.Up`’s error and returns only a generic message ([main.go:67](/Users/strelia/projects/courses/backend/cmd/migrate/main.go:67)). Its logger therefore cannot report the offending ID/count. No application `OnNotice` handler was found either, so `RAISE NOTICE` is not an established logging mechanism.

   A direct goose integration test can receive detailed errors while the production executable hides them.

   **Required plan change:** specify a safe, concrete route for offending IDs/counts and changed/skipped reporting—either structured migration diagnostics or an explicit read-only diagnostic procedure. Cover the actual executable/logging path. Do not rely solely on provider-level tests.

3. **BLOCKER — OPEN: M-R7/M-R8 do not prove the populated migration-to-master path.**

   Locations: [plan-ms1.md:111](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:111), [plan-ms1.md:112](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:112).

   M-R7 assigns the proof to application fakes, repository-list integration and command dispatch tests. None necessarily connects migrated metadata, real repository locking, real filesystem rewriting and process completion.

   **Counterexample traced:** the command is wired to the wrong media root or writer; fake application tests, list tests and no-queue tests can all pass. An ordered `make check` starts with empty data volumes ([Makefile:166](/Users/strelia/projects/courses/Makefile:166)), so successful startup may regenerate zero masters.

   **Required plan change:** add a populated fixture with a real master file, apply D8, run the actual command twice, then inspect `LANGUAGE="uk"` and compare file bytes. Include command failure propagation and observed service ordering. Keep fakes for targeted failure branches.

   **Byte-identical output is feasible:** the existing writer preserves deterministic ordering and already has an idempotence test ([master_test.go:73](/Users/strelia/projects/courses/backend/internal/media/adapters/ffmpeg/master_test.go:73)). The missing piece is proof through the new command.

4. **NONBLOCKING — ACTIONABLE: M-D1 misstates ADR 013 and underspecifies its mirror.**

   Locations: [plan-ms1.md:47](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:47), [plan-ms1.md:113](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:113).

   ADR 013 does not restrict shared predicates to modules participating “in one decision”; it explicitly chooses one pure rule over copies plus a stronger test ([ADR 013:19](/Users/strelia/projects/courses/docs/adr/013-access-contract.md:19)). A shared grammar predicate is a credible alternative with two actual consumers and no new port.

   **Counterexample traced:** `" uk "` is accepted by `transcript.NewLanguage`, which trims whitespace ([transcript.go:36](/Users/strelia/projects/courses/backend/internal/transcript/domain/transcript.go:36)); M-D1’s raw grammar rejects it. Thus constructor verdicts are not universally identical after lowercase conversion alone.

   Prefer a shared pure grammar predicate with module-specific normalization. If retaining duplication, correct the ADR rationale, define the mirror’s canonical-input scope and place cross-module comparison outside the media module, following the existing platform architecture-test pattern. Copied tables plus comments alone do not detect divergence.

5. **NONBLOCKING — ACTIONABLE: add narrow HTTP and picker acceptance.**

   Location: [plan-ms1.md:105](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:105).

   Add mint and rename requests with now-invalid `e`/`und-`, asserting **400 `invalid_request`** and unchanged persisted metadata. The existing mapping supports that contract ([handler.go:507](/Users/strelia/projects/courses/backend/internal/media/adapters/http/handler.go:507)).

   M-R10 should separately cover upload selection and editing an uppercase existing language, including retaining the edited track’s own option. The actual helper is `availableAudioLanguages`, not `offeredLanguages` ([audio-track.ts:129](/Users/strelia/projects/courses/frontend/src/modules/media/domain/audio-track.ts:129)).

**BLOCKED — MS1. Three substantive blockers remain open.**

Other requested checks:

- **Migration election:** default → embedded → ID matches accepted D8. Failed rows remain subject to CHECK but excluded from uniqueness; in-flight uploads participate in uniqueness. Evicted/non-Ready assets and assets without a currently associated lesson primary are excluded as specified. Step 1 does not inherently create a language-shape failure or normalized collision; its demonstrated problem is hiding existing ones.
- **Locks:** adding `transcripts` fits the table-lock mechanism. Explicitly retain `SET LOCAL lock_timeout = '5s'` from migration 00028; the plan currently references the precedent without stating that setting.
- **Index consumers:** repository-wide searches found no old-expression `ON CONFLICT` target or audio-index-definition assertion. Preserving the index name preserves the existing error mapping.
- **Runtime:** `runtime.Open` requires Redis and initializes tracing/metrics, but starts no HTTP server. Startup has a 10-second context; tracing shutdown has a five-second context. Storage initialization creates missing directories; its readiness probe is separate. No unavoidable one-shot hang was established.
- **Compose:** `make e2e` reaches the new dependency transitively through `nginx → media-worker → media-masters`. Upload already depends on healthy Redis and migration completion. The inspected dev/stub/e2e/CI overlays introduce no dependency cycle. API and ASR need no additional dependency for the stated command contract. Warm-start rerun and failure semantics still need explicit evidence.
- **Rewrite concurrency:** repository locks serialize metadata/master writers; atomic rename protects playback readers. No additional writer race was established.
- **Fixtures:** no incompatible persisted audio language was found in the inspected Go integration fixtures, browser fixtures or seed SQL. The known frontend acceptance of `x.repeat(35)` is already identified for replacement.

Binding and retained evidence:

- Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `895b2e2537dd2ad428239254444666d67c62c1b3`.
- Plan: untracked `docs/multi-subtitles/plan-ms1.md`, initial draft, SHA-256 `6b0bd337ad1682dfe0b2aea5b4810a07ddfa17045c0276407263ca73b1162f84`.
- Current phase-plan SHA-256: `cbe299106678907bbf4494196e182ff0c16d37e000e0ab84c267e2ebd3c59f43`. Raw rounds 1–6 were inspected. This is the first MS1 verdict; earlier D8 election/command architecture conclusions are retained, while its detailed migration and proof strategy receive the findings above. Product sources are unchanged since the phase-review snapshot; unrelated dirty documentation/configuration and untracked `backend/upload` were excluded.
- Prompt: candidate release directory `17a9527421e8fac3e43a16c287c404118699833c`, `prompts/plan-critic.md`, version 1, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package declares `0.8.0-alpha.1+native-handoff.1`; release commit remains unverified.
- No tests, writes, network probes or graph refresh ran. Counterexamples are file-based analyses. Required four-lens gate remains **NOT_RUN: 0 launched / 0 returned**, preserving the selected reviewer without spawning.

The review could not be saved or linked from the plan because the workspace is read-only.