VERDICT: BLOCKED

Reviewed `/Users/strelia/projects/courses` at HEAD `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`. Revised plan SHA-256: `534f714c903f4f86f385d9d2018d4f5e95a18659020cc136801b17bcd5c8742e`. The plan is untracked; inspected production-code paths have no tracked changes.

**Four independent lenses launched / four returned:** logic gaps, unnecessary complexity, edge cases, acceptance sufficiency. Read-only inspection; no tests run or files changed. O1–O6 remain fixed.

**Round-1 dispositions**

| R1 | Disposition | Reason |
|---|---|---|
| 1 — `und` draft segment-ID collision | **Resolved for the reported sequence.** | MS-D7 assigns fresh IDs on ASR import; M3-R11 names repeated publication. Retained published rows have a separate collision below. |
| 2 — recovery migration collisions | **Resolved.** | Track keys remain stable through rename/O5-PATCH; legacy offers do not overwrite destination storage. |
| 3 — processing→Ready loses text | **Resolved for the reported first-promotion sequence.** | Unmount flush plus legacy fallback and M3-R12 address it. Storage refusal is explicitly best-effort. Replacement of an existing Ready asset remains uncovered below. |
| 4 — relinquished D5 guarantees | **Rebuttal accepted.** | The table now explicitly accounts for attempt identity, generation provenance, idempotence and the client/server race window. The single-request provenance alternative is acknowledged. This accepts the narrower contract, not equivalence to D5. |
| 5 — duplicate accessible IDs | **Resolved.** | Instance-scoped IDs and M3-R7 cover multiple mounted editors. |
| 6 — wrong recovery suite / C17–C18 migration | **Resolved.** | M3-R6 names transcript recovery; M3-R9 explicitly migrates C17/C18 and scopes helpers. |
| 7 — German fixture claim | **Resolved.** | M3-R8 correctly limits the proof to routing, hints and language-keyed storage. |
| 8 — rename limitation versus `und` exception | **Resolved.** | §6 explicitly preserves the `und`-origin exception through subsequent renames. |

**Blocker**

1. **Retained published rows still collide when their track changes language.**  
   Where: [plan-ms2-3.md:261](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:261), [fresh-ID rule:290](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:290).

   Sequence: open a saved `en` transcript → edit existing rows → rename its track to unused `de` → Save without another ASR Load. MS-D6 preserves the rows, including IDs already persisted under `en`. PUT deletes only destination-language segments before inserting ([write.go:83](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:83)); those IDs remain globally occupied ([00007_transcript.sql:28](/Users/strelia/projects/courses/backend/migrations/00007_transcript.sql:28)). Saving `de` fails.

   The same transition needs destination-token handling: the old `en` token cannot update absent `de` ([write.go:69](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:69)). Today’s token-derived editor key also remounts on token changes ([LessonTranscriptField.tsx:87](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/LessonTranscriptField.tsx:87)).

   **Required plan change:** separate retained working rows from the destination baseline/token; assign fresh IDs when copying retained or recovered rows across transcript languages. Preserve rows while the destination loads and block Save on a failed destination read. Add acceptance for renaming to both absent and existing destination transcripts, preserving the original transcript.

**Major**

2. **Track-based recovery becomes unreachable after ordinary tab-identity changes.**  
   Where: [plan-ms2-3.md:243](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:243), [recovery lookup rules:325](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:325).

   An author edits track A while replacement video B processes. Promotion creates new track identities ([transcode.go:478](/Users/strelia/projects/courses/backend/internal/media/application/transcode.go:478)). Tabs then contain only B’s tracks plus transcript orphans. A’s flushed snapshot remains under `track:A`; neither the new track nor the orphan reads it. The only fallback specified is `user:lesson`.

   Likewise, adding or renaming a track to an orphan’s language removes the dirty `lang:<language>` editor, while the replacement `track:<id>` editor does not read its snapshot. These losses exceed the explicit track-deletion limitation.

   **Required plan change:** define recovery when a track tab becomes an orphan, an orphan becomes attached, or a newer Ready asset replaces the tab set. A read-only offer from detached snapshots can preserve the no-transfer decision. Add browser coverage for replacement promotion during typing and attachment of a dirty orphan.

3. **The replacement freshness check has no failure contract.**  
   Where: [plan-ms2-3.md:267](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:267), [M3-R10:390](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:390).

   Missing-track and changed-language responses are covered; failed refetches are not. TanStack Query can return an error result while retaining cached data. Inspecting that data after `refetch()` can permit PUT without the successful fresh read promised by MS-D1.

   **Required plan change:** require a successful fresh response. Network errors, timeouts and error results must prevent PUT, preserve edits and show a retryable failure. Add unit/browser evidence asserting zero PUT requests after a failed freshness GET.

4. **M3-R11’s assigned fixture cannot execute its regression sequence.**  
   Where: [plan-ms2-3.md:391](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:391), [M3-R8:388](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:388).

   M3-R11 places `en→de` rename/republication inside M3-R8, whose asset already contains a Ready `de` dub. The current case-insensitive language uniqueness constraint rejects that rename ([00030_audio_track_language_shape.sql:88](/Users/strelia/projects/courses/backend/migrations/00030_audio_track_language_shape.sql:88)). The second PUT never occurs.

   **Required plan change:** use an unused third language or a separate one-track fixture, retaining assertions that both published transcripts survive.

**Minor**

5. **Remove the publish-only dependency chain as part of endpoint cleanup.**  
   Where: [plan-ms2-3.md:92](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:92).

   Deleting `Drafting.Publish` leaves its exclusively used `TranscriptPublisher` interface, field and constructor parameter unless explicitly cleaned up ([draft.go:16](/Users/strelia/projects/courses/backend/internal/transcript/application/draft.go:16)).

   **Plan change:** include their composition wiring and publish-only test fakes in removal scope. No broader complexity blocker was found.

**Open blocking findings: 4 — one Blocker and three Majors.**