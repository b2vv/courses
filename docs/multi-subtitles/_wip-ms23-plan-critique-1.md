VERDICT: BLOCKED

Reviewed HEAD `37ec4b0`; plan SHA-256 `116638481011ab8f82f103d4db655b0047a7084170906a90507a56b526631963`. Four independent lenses launched / four returned. Read-only inspection; no tests run or files changed. O1–O6 remain fixed.

The publication-path discovery is correct: `/draft/publish` has no frontend caller beyond generated code. `CourseEditorView` passes draft lines into `TranscriptEditor`, which saves through `LessonTranscriptField` and ordinary PUT. Removing the unused endpoint and `SaveDraft` is reasonable. The claim that this makes **all** D5 protections meaningless is not sound.

**Blocker**

1. **Reusing an `und` draft under another language will collide on segment IDs.**  
   Where: [plan-ms2-3.md:268](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:268).

   Valid sequence: recognize under `und` → publish `en` → rename track to `de` → load the same exempt, non-stale draft → save `de`. The retained `en` transcript becomes an orphan, but still owns its segment IDs. [CourseEditorView.tsx:303](/Users/strelia/projects/courses/frontend/src/modules/admin/ui/CourseEditorView.tsx:303) and [TranscriptEditor.tsx:58](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:58) preserve those IDs. They are globally unique in [00007_transcript.sql:28](/Users/strelia/projects/courses/backend/migrations/00007_transcript.sql:28); PUT deletes only the destination language before inserting them ([write.go:83](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:83)). The second publication rolls back on uniqueness violation.

   **Plan change:** require fresh transcript-segment identities when importing ASR lines. Add acceptance for this exact reuse sequence, preserving both language transcripts.

2. **Recovery migration can overwrite unrelated unsaved work.**  
   Where: [plan-ms2-3.md:299](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:299).

   Both the pre-PATCH target write and legacy migration lack a destination-collision policy. A newer `user:lesson:en` snapshot can be overwritten by an older legacy or `und` snapshot. If PATCH fails, the target-language copy remains despite no rename. This also drops the parent’s explicit collision/repeated-migration requirement ([plan.md:224](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:224)).

   **Plan change:** define collision preservation, repeat-safe migration, and cleanup after failed PATCH. Keep live editor state through successful PATCH/failed PUT; IndexedDB alone cannot guarantee transfer because its writes are best-effort. Extend M3-R4/R6 with occupied destination, rejected storage, PATCH failure, and retry cases.

3. **Processing → Ready can discard the author’s current text.**  
   Where: [plan-ms2-3.md:244](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:244), [line 302](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:302).

   An author types Ukrainian lines while the first video processes. Promotion creates an `und` track; no saved `uk` inventory entry exists. Switching to tabs removes the old editor, and the stated migration rule discards its snapshot because there is no matching language tab. Uncaptured text is also exposed: [use-transcript-recovery.ts:89](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:89) cancels the idle timer on unmount without capturing.

   **Plan change:** specify preservation across the single-editor→tabs transition, and rename/delete transitions. Add a browser scenario that types before promotion, allows promotion without blur/reload, and verifies the text remains recoverable.

**Major**

4. **MS-D1 confuses permission to author arbitrary text with protection against accidental stale publication.**  
   Where: [plan-ms2-3.md:96](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:96), [line 120](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:120).

   Manual PUT means the system cannot guarantee semantic language correctness against a deliberately bypassing author. It does not eliminate D5’s useful guarantees for the normal ASR workflow: reviewed-attempt identity, publication-time track/language/existence checks, and generation provenance. A load-time guard cannot protect against rename, deletion, regeneration, or promotion after loading.

   Generation freshness and response idempotence are explicitly relinquished, but the other protections disappear without equivalent accounting. Language-tag equality also cannot prove that text matches the spoken language.

   The rejected alternative is a false dichotomy: preserving provenance does **not** require HTTP `SaveDraft` plus two requests. One request can carry corrected segments and ASR provenance, reusing transcript persistence.

   **Plan change:** either preserve those checks in a single publication request, or explicitly record the additional relinquished guarantees and expected race outcomes. Add checkable acceptance for the chosen behavior.

5. **Force-mounted editors duplicate accessible IDs.**  
   Where: [plan-ms2-3.md:257](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:257).

   Existing editor heading, recovery, and draft-offer IDs use only `lessonId` ([TranscriptEditor.tsx:166](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:166), :176, :208). Multiple mounted language editors duplicate them; active content can reference a hidden panel’s heading.

   **Plan change:** give each editor a unique language/track identity. Extend M3-R7 to verify unique IDs and correct accessible names with multiple panels mounted.

6. **The recovery evidence points at the wrong suite, and test migration is incomplete.**  
   Where: [plan-ms2-3.md:335](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:335), [line 338](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:338).

   `e2e/recovery.spec.ts` tests infrastructure outages, not transcript-edit recovery. Actual coverage lives in [transcript.spec.ts:499](/Users/strelia/projects/courses/frontend/e2e/transcript.spec.ts:499) and ASR C18 ([asr.spec.ts:389](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:389)). C18 still invokes the removed language Select at :407. C17 and helpers also assume a single editor/draft, which multiple tabs invalidate.

   **Plan change:** explicitly migrate C17/C18 and scope helpers to the selected track. Name browser proofs for language-isolated snapshots, legacy migration, and PATCH-success/PUT-failure→reload→restore. Existing no-video recovery coverage cannot prove these.

**Minor**

7. **M3-R8’s German fixture proves routing/storage, not German recognition quality.**  
   Where: [plan-ms2-3.md:337](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:337).

   Both speech fixtures contain English flite speech ([media.setup.ts:74](/Users/strelia/projects/courses/frontend/e2e/media.setup.ts:74)). Keeping them is reasonable for per-track processing, hint plumbing, and language-keyed publication. State that narrower proof; use German speech if claiming recognition fidelity.

8. **The rename limitation contradicts the accepted `und` exception.**  
   Where: [plan-ms2-3.md:359](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:359).

   Automatic re-recognition applies to drafts claimed with a defined language. Drafts claimed under the `und` family remain exempt, including subsequent renames, as the parent explicitly permits. Qualify the statement.

No separate backend blocker was found in the proposed backfill, failed-track exclusion, Down reduction, single-statement claim, per-track completion/recovery, reformulated C09, TaskID, or hint mapping. The leaf-uniqueness assumption is maintained by current production writers, although not by a database uniqueness constraint. The old-task rejection strategy is coherent.

The acceptance rows are a useful base, but they omit the concrete failure sequences above. **Six blocking findings remain: three Blocker and three Major.**