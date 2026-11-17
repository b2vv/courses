**VERDICT: BLOCKED**

Reviewed HEAD `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`. Plan SHA-256: `82ecb668eab0029daa2d529ecb8966f23bd17ca6d6689cad4349a2b6bb31c07e`.

Four independent lenses launched / four returned: logic gaps, unnecessary complexity, edge cases, acceptance sufficiency. Read-only inspection against the requested code paths; no files changed or tests run. O1–O6 remain fixed.

**Round 3 dispositions**

| Finding | Disposition | Reason |
|---|---|---|
| Major 1 — recovery loses row provenance | **Open, partially resolved.** | MS-D6 correctly persists `originLanguage`, treats absent provenance as unknown, and adds reload/restore coverage. MS-D9 still explicitly replaces origin with snapshot language for detached recovery, retaining the collision path below. |
| Major 2 — M3-R8 cannot distinguish recognized audio | **Resolved.** | M3-R8 now requires distinct `Welcome`/`dubbed` recognition results, then checks each saved transcript’s language and contents against its PUT payload. The existing speech fixtures support this proof. `en-GB` maps to the English hint; embedded `und` remains auto-detected. |

**Blocker:** none.

**Major — detached recovery still overwrites genuine provenance.**  
Where: [plan-ms2-3.md:377](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:377).  
Lenses: logic gaps + edge cases. **Blocking.**

The instruction sets `originLanguage` to the snapshot’s **language**, contradicting MS-D6’s separate provenance field.

Concrete sequence:

1. Open saved `fr`, edit, rename the track to `pl`, and capture.
2. Snapshot contains `language=pl`, `originLanguage=fr`.
3. Replace the video, making the old track’s snapshot detached.
4. Restore into a current `pl` track.

Following line 377 sets origin to `pl`, so Save retains IDs still owned by `fr`. Segment IDs are globally unique ([00007_transcript.sql:28](/Users/strelia/projects/courses/backend/migrations/00007_transcript.sql:28)); PUT deletes only destination-language segments before inserting ([write.go:83](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:83)). Publication therefore fails.

**Required plan edit:** every restore path must use `snapshot.originLanguage`, with absence remaining unknown. Extend M3-R6/R12 with this detached dirty-rename sequence, asserting successful destination publication and unchanged source transcript.

**Minor — clarify M3-R11’s fixture sequence.**  
Where: [plan-ms2-3.md:430](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:430).  
Lens: acceptance sufficiency. **Nonblocking.**

Case (c) leaves the track at `en`; case (d) starts with `pl→es`. Specify independent fixtures or an explicit return to `pl`. The parenthetical also still says M3-R8 occupies `de`, although r4 uses `en-gb`.

The complexity lens identified fresh IDs on every outbound PUT as an optional simplification of provenance handling. The bounded provenance design is also workable once the contradiction above is removed; redesign is unnecessary.

**Open blocking findings: 1 Major. VERDICT: BLOCKED.**