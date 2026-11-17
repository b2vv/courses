NEXT: Read shared entry rules and repository guidance, then inspect the two review findings and MS3d-fix checkpoint before editing permitted frontend files.
FACT: Read the shared entry rules; no delegation is planned. The owner-scoped MS3d fix is the active task.
NEXT: Verify repository root/state and read local guidance, briefing and the required review/spec/checkpoint sources.
FACT: Verified root /Users/strelia/projects/courses and read AGENTS.local.md; shared tree contains unrelated owner work, which will be preserved. Broad source output was truncated, so required sources will be reread in bounded sections.
NEXT: Read scoped workflow/skills and focused review/spec sections; inspect transcript identity and destination sites.
FACT: Branch/HEAD match ms2-3-asr-per-track @ 38c7915. Standards finding 1 and Spec F1 agree with the owner checkpoint. Scope is two frontend fixes plus this mandated journal; no Git mutations, backend, generated files, e2e or Docker.
DECISION: Small fix route, existing checkpoint READY; owner is current requester, delivery is this journal. Author checks now; independent review/design evidence and whole-PR mutations remain later owner-coordinated work.
NEXT: Read managed ponytail/receiving-code-review skills, repository standards and implementation/test sites.
FACT: Skills and standards read. Existing UUID generation and Rstest are available; no dependency needed. capturedAt identity spans application, both adapters, hook ownership, editor ancestry and tests; admin fixtures must also change to typecheck.
NEXT: Read complete recovery implementations and focused test fixtures before the first fix.
FACT: Verified root cause: store cleanup and editor ancestry use timestamps as identity; detached matching admits orphan tabs. IndexedDB tests use a request/event stub, not fake-indexeddb.
DECISION: Require revision for snapshots/ancestry, remove the obsolete pre-list acceptedSource transform (old entries have no revision), and add an explicit track/orphan kind to recovery tabs. Preserve capturedAt solely for time/order.
NEXT: Finish focused hook/application/editor test inspection, then implement collision fix first.
FACT: Implemented revision identity in snapshot/ancestry types, port, both adapters, hook seen/written ownership, editor source mapping/dedup/discard, and clearAcceptedTranscriptSnapshot comparison/deletion. Added revisions to existing snapshot fixtures, including required admin fixtures.
NEXT: Inspect fixture edits and run focused collision-path tests/typecheck; then add discriminating collision tests.
FACT: Initial typecheck command ran at repository root and exited ENOENT (no root package.json); no compiler ran. Correct directory is frontend. Fixture inspection found timestamp-shaped ancestry examples needing explicit revision conversion.
NEXT: Correct fixtures and add equal-time regression coverage before rerunning from frontend.
FACT: Corrected ancestry fixtures to key+revision and strengthened existing memory/application/hook collision fixtures to equal capturedAt with distinct revisions. Frontend typecheck exited 0.
NEXT: Run the focused transcript/admin suites to verify the first fix, then implement track-only destination selection.
FACT: Focused Rstest could not start in sandbox: listen EPERM ::1:3000, exit 1; zero tests executed. The suite needs its local listener permission.
NEXT: Retry the same authorized frontend suite with sandbox escalation; no Docker or e2e.
FACT: Escalated unit-suite launch was accepted and is running; no result yet. The test log is initially empty.
NEXT: Implement the independent track-only destination fix and explicit recovery tab kind while the baseline suite completes.
FACT: Added explicit track/orphan recovery-tab kind in application type and admin mapping; matching now uses tracks only, fallback default track → first track → prior no-track fallback. First suite completed with fixture assertion failures; production identity paths compiled.
NEXT: Correct the remaining randomized expected-value fixture, add named regressions and rerun.
FACT: First suite: 445 passed, 2 failed (random UUIDs recomputed in expected fixtures); corrected both. Added named MS3d memory, IndexedDB, missing-revision parser and two-window hook regression tests. Existing hook ownership matrix now covers equal timestamps across blur/idle/unmount/Discard/Save.
NEXT: Add destination regression coverage and run the authorized frontend checks in order.
FACT: A command used frontend as cwd while its Python/journal paths were root-relative: the intended destination-test edits did not run. Format did run and exited 0 (430 files); npm script expands to oxfmt ., so path arguments do not narrow it. Checking the diff for unauthorized changes before continuing.
NEXT: Apply the pending destination tests from repository root, then verify changed-file scope and rerun ordered frontend checks.
FACT: Scope check shows only allowed frontend files changed. Ordered attempt: format exit 0, typecheck exit 2 (destinations declaration missed a formatted fallback line), lint exit 0 (240 architecture files). No suite started on this invalid snapshot.
NEXT: Correct the destination declaration and repeat ordered checks before tests.
FACT: Corrected destination declaration; final ordered static checks: npm run format exit 0, npm run typecheck exit 0, npm run lint exit 0 (240 architecture files).
NEXT: Run transcript/admin unit suites on the corrected snapshot, then execute one-line mutations for each of the five new MS3d tests in a disposable copy.
FACT: Final scoped suite exit 0: 452 passed, 0 failed, 0 skipped, 27 files. All five new MS3d tests are included. Raw evidence: /tmp/ms3d-tests.log; static logs /tmp/ms3d-{format,typecheck,lint}.log.
NEXT: Create a disposable frontend copy under /private/tmp; confirm its five-test baseline, mutate one production line per new test, capture attributable assertion failures, restore exact bytes and recheck.
FACT: Disposable copy /private/tmp/ms3d-mutations-jvi74brj/frontend; source/config copied, installed node_modules linked; no Git clone/worktree or shared source mutation.
FACT: Disposable baseline: exit 0; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript src/modules/admin --testNamePattern MS3d; raw log /private/tmp/ms3d-mutations-jvi74brj/baseline.log.
FACT: "counts": { "testFiles": 27, "failedFiles": 0, "tests": 452, "failedTests": 0, "passedTests": 5, "skippedTests": 447, "todoTests": 0 }
NEXT: Mutation memory: `if (snapshots.get(key)?.revision !== revision) return Promise.resolve(false);` → `if (snapshots.get(key) === undefined) return Promise.resolve(false);`; only disposable file src/modules/transcript/adapters/transcript-history-memory.ts.
FACT: Disposable memory: exit 1; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript/application/recovery.test.ts --testNamePattern MS3d memory cleanup preserves a replacement with equal capturedAt and different revision; raw log /private/tmp/ms3d-mutations-jvi74brj/memory.log.
FACT: "counts": { "testFiles": 1, "failedFiles": 1, "tests": 31, "failedTests": 1, "passedTests": 0, "skippedTests": 30, "todoTests": 0 }
FACT: Named test failed:
```text
401 |   await store.write(first);
402 |   await store.write(second);
403 |   expect(await store.clearIfVersion(first.key, first.revision)).toBe(false);
    |                                                                 ^
404 |   expect(await store.read(first.key)).toEqual(second);
405 |   expect(await store.clearIfVersion(second.key, second.revision)).toBe(true);
```
FACT: Restored src/modules/transcript/adapters/transcript-history-memory.ts; SHA-256 23a49707b395d0bd3348c6093c9639d8ea356d452dd13d96de5ab3ef5f2b9acd.
NEXT: Continue with the next single-line mutation, or restored baseline after the fifth.
NEXT: Mutation indexeddb: `result.data.revision === revision` → `true`; only disposable file src/modules/transcript/adapters/transcript-history-indexed-database.ts.
FACT: Disposable indexeddb: exit 1; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts --testNamePattern MS3d IndexedDB cleanup preserves a replacement with equal capturedAt and different revision; raw log /private/tmp/ms3d-mutations-jvi74brj/indexeddb.log.
FACT: "counts": { "testFiles": 1, "failedFiles": 1, "tests": 25, "failedTests": 1, "passedTests": 0, "skippedTests": 24, "todoTests": 0 }
FACT: Named test failed:
```text
193 |   await store.write(snapshot);
194 |   await store.write(second);
195 |   expect(await store.clearIfVersion(snapshot.key, snapshot.revision)).toBe(false);
    |                                                                       ^
196 |   expect(await store.read(snapshot.key)).toEqual(second);
197 |   expect(await store.clearIfVersion(second.key, second.revision)).toBe(true);
```
FACT: Restored src/modules/transcript/adapters/transcript-history-indexed-database.ts; SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
NEXT: Continue with the next single-line mutation, or restored baseline after the fifth.
NEXT: Mutation schema: `  revision: z.string(),` → `  revision: z.string().optional(),`; only disposable file src/modules/transcript/adapters/transcript-history-indexed-database.ts.
FACT: Disposable schema: exit 1; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript/adapters/transcript-history-indexed-database.test.ts --testNamePattern MS3d IndexedDB rejects a stored snapshot without revision; raw log /private/tmp/ms3d-mutations-jvi74brj/schema.log.
FACT: "counts": { "testFiles": 1, "failedFiles": 1, "tests": 25, "failedTests": 1, "passedTests": 0, "skippedTests": 24, "todoTests": 0 }
FACT: Named test failed:
```text
149 |   const { revision: _revision, ...legacy } = snapshot;
150 |   storedValue(legacy);
151 |   expect(await createIndexedDbTranscriptHistory()?.read(snapshot.key)).toBeNull();
    |                                                                        ^
152 | });
153 | 
```
FACT: Restored src/modules/transcript/adapters/transcript-history-indexed-database.ts; SHA-256 b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59.
NEXT: Continue with the next single-line mutation, or restored baseline after the fifth.
NEXT: Mutation hook: `const revision = crypto.randomUUID();` → `const revision = String(capturedAt);`; only disposable file src/modules/transcript/ui/use-transcript-recovery.ts.
FACT: Disposable hook: exit 1; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript/ui/use-transcript-recovery.test.tsx --testNamePattern MS3d hook captures unique revisions in one millisecond and Save preserves the other window; raw log /private/tmp/ms3d-mutations-jvi74brj/hook.log.
FACT: "counts": { "testFiles": 1, "failedFiles": 1, "tests": 21, "failedTests": 1, "passedTests": 0, "skippedTests": 20, "todoTests": 0 }
FACT: Named test failed:
```text
252 |   if (a === null || b === null) throw new Error('Both captures must exist');
253 |   expect(a.capturedAt).toBe(b.capturedAt);
254 |   expect(a.revision).not.toBe(b.revision);
    |                          ^
255 |   act(() => first.result.current.finalize());
256 |   await act(async () => {});
```
FACT: Restored src/modules/transcript/ui/use-transcript-recovery.ts; SHA-256 034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34.
NEXT: Continue with the next single-line mutation, or restored baseline after the fifth.
NEXT: Mutation destination: `const destinations = tracks.length === 0 ? tabs : tracks;` → `const destinations = tabs;`; only disposable file src/modules/transcript/application/recovery.ts.
FACT: Disposable destination: exit 1; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript/application/recovery.test.ts --testNamePattern MS3d detached uk snapshot skips a matching orphan for default or first track; raw log /private/tmp/ms3d-mutations-jvi74brj/destination.log.
FACT: "counts": { "testFiles": 1, "failedFiles": 1, "tests": 31, "failedTests": 1, "passedTests": 0, "skippedTests": 30, "todoTests": 0 }
FACT: Named test failed:
```text
427 |     isDefault: true,
428 |   };
429 |   expect(place([detached], [defaultTrack, orphan])[0]?.instanceId).toBe('track:en');
    |                                                                    ^
430 |   expect(place([detached], [orphan, first, defaultTrack])[0]?.instanceId).toBe('track:en');
431 |   expect(place([detached], [{ ...orphan, isDefault: true }, first])[0]?.instanceId).toBe(
```
FACT: Restored src/modules/transcript/application/recovery.ts; SHA-256 513f1b3ea6b29e391310d3209990f5b4008988396367a5ffc5d5a5e4f58de33e.
NEXT: Continue with the next single-line mutation, or restored baseline after the fifth.
FACT: Disposable restored: exit 0; command (cwd /private/tmp/ms3d-mutations-jvi74brj/frontend): /Users/strelia/projects/courses/frontend/node_modules/.bin/rstest run src/modules/transcript src/modules/admin --testNamePattern MS3d; raw log /private/tmp/ms3d-mutations-jvi74brj/restored.log.
FACT: "counts": { "testFiles": 27, "failedFiles": 0, "tests": 452, "failedTests": 0, "passedTests": 5, "skippedTests": 447, "todoTests": 0 }
FACT: Five new tests each killed their single-line mutation; restored five-test baseline is green. Shared production sources remained untouched during mutations.
NEXT: Audit final diff/identity sites and record changed-file hashes for handoff.
FACT: Final production diff audit: capturedAt remains only in snapshot time type/schema/read, Date.now capture, and sorting; no timestamp identity comparisons remain. git diff --check exited 0. All five disposable mutations failed at their named assertions and restored baseline passed 5/5.
NEXT: Review fixture diff for accidental weakening, then record the complete identity-site inventory and final SHA-256 manifest.
FACT: Final fixture review replaced a broad admin cleanup argument matcher with the exact accepted source revision, preserving the original assertion strength. No new production changes.
NEXT: Repeat the requested ordered checks on the final fixture snapshot, then append the site inventory and final file hashes. Mutation proofs remain applicable: their production and new-test files did not change.
FACT: Final snapshot static checks repeated in order: npm run format exit 0 (400 files), npm run typecheck exit 0, npm run lint exit 0 (240 architecture files).
NEXT: Finish the exact transcript/admin suite on the final snapshot.
FACT: Final suite on final shared-tree snapshot: `cd frontend && npx rstest run src/modules/transcript src/modules/admin` exit 0; 452 passed, 0 failed, 0 skipped, 27 test files. Requested ordered checks all pass: format 0 → typecheck 0 → lint 0 → Rstest 0. Logs: /tmp/ms3d-format.log, /tmp/ms3d-typecheck.log, /tmp/ms3d-lint.log, /tmp/ms3d-tests.log.
FACT: Final `git diff --check` exit 0; frontend changed-file allowlist verified (12 files). All mutation-target production files and five-test source files byte-match the restored disposable baseline. No backend/generated/e2e/Docker or Git write commands were used; unrelated WIP was preserved.
FACT: Complete identity-site inventory (all paths relative to frontend/src/modules/transcript):
- application/recovery.ts:32 AcceptedTranscriptSource key+revision, :42 TranscriptSnapshot revision, :259 clearAcceptedTranscriptSnapshot stored-version match, :267 atomic clear argument.
- ports/TranscriptHistoryStore.ts:17 clearIfVersion revision:string contract.
- adapters/transcript-history-memory.ts:23–24 conditional deletion argument/comparison.
- adapters/transcript-history-indexed-database.ts:7 ancestry revision schema, :11 required snapshot revision, :106 read mapping, :122/:132 atomic deletion argument/comparison.
- ui/use-transcript-recovery.ts:69–70 seen/written types, :100 initial-read ownership, :130 conditional cleanup, :135 refreshed-read ownership, :159 UUID capture, :163 snapshot write, :179 seen/written capture ownership.
- ui/TranscriptEditor.tsx:168 captured ancestry key+revision, :259 ancestry deduplication, :265 detached Discard conditional deletion.
FACT: Destination sites: application/recovery.ts:202 explicit tab kind; :229–231 track candidates/default→first fallback, :236 case-insensitive matching; admin/ui/LessonLanguageTabs.tsx:159 maps track existence to kind. With zero tracks the previous tab-language/default behavior is retained. Existing fixture updates cover transcript application/adapters/hook/editor and the necessary admin test.
DECISION: Pre-launch stored snapshots without revision parse as null. The old pre-list acceptedSource transform is removed because those timestamp-only sources have no unique revision; inventing one would grant cleanup ownership without evidence. No schema version bump/migration shim.
FACT: Changed frontend file SHA-256 manifest (journal is the append-only evidence record):
- `frontend/src/modules/admin/ui/LessonLanguageTabs.test.tsx` — `a83395dd07cda5f5f8c0cfbe705f9fa7ba9dbbe92d2c9746d13259290e60847c`
- `frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` — `5751be5c1083b53ea3484f30cf2e83edb06982181838098d7ad696665ee07460`
- `frontend/src/modules/transcript/adapters/transcript-history-indexed-database.test.ts` — `794d636e9ad0990c5dbb3eb418b812ff1875e17a9f0aa5779cb3efe87b04914d`
- `frontend/src/modules/transcript/adapters/transcript-history-indexed-database.ts` — `b9c36a8d73679b1d25c1b65bea1fb95a7d7081e8a76e8d53c177b933b5f28b59`
- `frontend/src/modules/transcript/adapters/transcript-history-memory.ts` — `23a49707b395d0bd3348c6093c9639d8ea356d452dd13d96de5ab3ef5f2b9acd`
- `frontend/src/modules/transcript/application/recovery.test.ts` — `5095abdf3ffe147d2b72129de9037def089bf91b90243fc5604e2670dc97ab1c`
- `frontend/src/modules/transcript/application/recovery.ts` — `513f1b3ea6b29e391310d3209990f5b4008988396367a5ffc5d5a5e4f58de33e`
- `frontend/src/modules/transcript/ports/TranscriptHistoryStore.ts` — `4378d8e7b9a32570fb395a6c093265187652bf16885b5401941a56acc94ad2a7`
- `frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx` — `3d6e78aa978cf29fc9e09c13cdd0b967872d05eee6026e08c08715b246839cf7`
- `frontend/src/modules/transcript/ui/TranscriptEditor.tsx` — `1498b2eee1604b14fe8b5ead1f09620acfec8095277cf4cae2f1d4bece1764b2`
- `frontend/src/modules/transcript/ui/use-transcript-recovery.test.tsx` — `cd866a604a415e339af0cc8b874612b03117a9ba1601c2c99534322894af5d33`
- `frontend/src/modules/transcript/ui/use-transcript-recovery.ts` — `034118e97557b01be7f8fe264690a6cad8830caf9c12fb68552afba6e38d8b34`
FACT: MS3d-fix scope complete: both defects fixed, equal-time replacement survival proven in memory/IndexedDB/hook paths, orphan-only destination proven, and all five new tests mutation-discriminated. IndexedDB tests exercise the existing request/event boundary stub, not live browser persistence.
NEXT: Hand off this exact uncommitted snapshot for the owner-coordinated independent review and later PR design evidence/whole-PR mutations/full gate. This slice does not claim PR #109 merge readiness or those broader evidence gaps closed.
