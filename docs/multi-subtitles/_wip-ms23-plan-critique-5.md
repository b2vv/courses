**VERDICT: BLOCKED**

Reviewed HEAD `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`; plan SHA-256 `030d561af447519ce1e6d186a9f03bb5403ff45478808251e3498ea6b68ba6e8`.

Four independent lenses launched / four returned. Read-only inspection; no files changed or tests run. O1–O6 remain fixed.

**Round 4 dispositions**

| Finding | Disposition |
|---|---|
| Major — detached recovery loses provenance and collides on segment IDs | **Resolved.** Fresh UUIDs on every outbound editor PUT eliminate dependence on provenance, including detached recovery. |
| Minor — ambiguous M3-R11 sequence and stale `de` parenthetical | **Resolved.** The separate fixture avoids occupied languages; (c) ends at `en`, (d) proceeds `en→es`, and (e) continues `es→de` through video replacement. |

The fresh-ID mechanism is supported by the code. The only database reference to `transcript_segments.id` is the cascading FK in [00007_transcript.sql:60](/Users/strelia/projects/courses/backend/migrations/00007_transcript.sql:60). PUT deletes destination-language segments and rebuilds their document links within one transaction ([write.go:83](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:83), :100–113).

Frontend IDs support React keys, active-row comparison and local recovery equality; none requires persisted segment identities across saves. UUID replacement at the outbound boundary is compatible with those uses. No e2e assertion requires segment IDs to survive an editor save.

**Blocker:** none.

**Major — fresh inventory tokens bypass optimistic concurrency.**  
Where: [plan-ms2-3.md:282](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:282).  
Lenses: logic gaps + edge cases. **Blocking; high confidence.**

The plan takes the destination’s freshly fetched `updatedAt` before every PUT:

1. A opens `en` at token T1 and edits.
2. B saves newer `en`, producing T2.
3. A’s Save fetches T2 and submits A’s older rows with T2.
4. The server accepts and silently overwrites B’s work.

The replacement confirmation does not apply because A’s destination language has not changed. This contradicts the existing optimistic-lock decision ([transcript/plan.md:368](/Users/strelia/projects/courses/docs/transcript/plan.md:368)). The backend rejects stale writes only when given the original baseline token ([write.go:69](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:69)).

**Required plan edit:** retain the editor’s baseline token for same-language saves. Fresh inventory may detect a conflict but must not silently advance that token. Cross-language replacement may use a destination token after explicit confirmation bound to that version. Add acceptance for two same-language editors and stale recovered work, proving newer server edits cannot be silently overwritten.

**Minor:** none.

The complexity lens is clean: unconditional UUIDs simplify recovery without introducing another mechanism. Acceptance sufficiency is clean for the revised ID/recovery scenarios; add the concurrency coverage required above.

**Open blocking findings: 1 Major. VERDICT: BLOCKED.**