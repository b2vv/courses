**R5 disposition: resolved for the reported overwrite.** MS-D6 preserves the editor’s baseline token for same-language saves; inventory cannot advance it. Cross-language replacement requires confirmation bound to the shown version. M3-R13 covers stale editors, stale recovery and destination races. These rules match the backend’s conditional UPDATE and unique-key protections.

**Blocker:** none.

**Major — recovery does not preserve which language owns the baseline token.**

Where: [plan-ms2-3.md:284](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:284), [line 384](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:384).  
Lenses: logic gaps, edge cases, acceptance sufficiency. **Blocking.**

The revised rule treats `snapshot.language` as the baseline language. Existing snapshots define that field as the **editor language at capture**, while `expectedUpdatedAt` belongs to the transcript originally edited ([recovery.ts:33](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:33), [capture:105](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:105)).

Concrete sequence:

1. Open saved `en` at token T1; edit.
2. Rename the track to absent `es`; capture before Save.
3. Snapshot contains `language=es`, token T1 belonging to `en`.
4. Reload, restore into `es`, Save.
5. MS-D6 classifies this as same-language and sends T1. The backend attempts UPDATE of nonexistent `es`, returning `transcript_modified` ([write.go:69](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:69)).

This contradicts M3-R11(d)’s required successful publication. Fresh segment IDs address ID collisions but cannot fix token ownership.

**Required plan edit:** preserve the baseline **language/token pair** separately from the current destination through rename, capture, restore and recapture. Define safe handling for older snapshots whose token ownership is unknown. Extend M3-R11/M3-R13 to assert the outbound token after rename→reload→restore: absent destination sends none; existing destination requires version-bound confirmation; unchanged-language stale work retains its original token.

**Minor:** none.

The unnecessary-complexity lens is clean: existing publication, recovery storage and unconditional outbound UUIDs remain reasonable reuse choices. All four requested lenses completed in this critic session; no subagents launched.

Reviewed HEAD `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`; untracked plan SHA-256 `a4bbff6e9ab845bd717720c1058b2617285faad3cdfe09cf94977c64673e5ae6`. O1–O6 remain fixed. Read-only inspection; no files changed or tests run.

**Open blocking findings: 1 Major.**

**VERDICT: BLOCKED**