**VERDICT: BLOCKED**

Reviewed HEAD `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`; untracked plan SHA-256 `f268ea3df044c6b8e76662c9936ac7b2ea52d803d4fe3c5e62c5a21548e4fb50`. Inspected production-code paths have no tracked changes. O1–O6 remain fixed.

**Four independent lenses launched / four returned:** logic gaps, unnecessary complexity, edge cases, acceptance sufficiency. Read-only inspection; no files changed or tests run.

**Round 2 dispositions**

| R2 finding | Disposition | Reason |
|---|---|---|
| Blocker 1 — retained IDs and wrong destination token | **Resolved for the reported in-memory sequence.** | Stable editor identity, dirty-row retention, cross-language ID regeneration and destination-token selection address it ([plan:264](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:264)). Recovery introduces a related gap below. |
| Major 2 — unreachable snapshots after replacement or orphan attachment | **Resolved.** | MS-D9 enumerates detached keys, defines offers and preserves source snapshots until Save/discard; M3-R6/R12 cover the transitions. |
| Major 3 — freshness-read failure contract | **Resolved.** | Direct repository reads must both succeed; errors preserve edits and produce **0 PUT**, with unit/browser evidence specified. |
| Major 4 — occupied `de` fixture | **Resolved.** | M3-R11 now uses a separate one-track fixture and unused destination languages. |
| Minor 5 — publish-only dependency chain | **Resolved.** | Removal explicitly includes the interface, field, constructor parameter, composition wiring and test fakes. |

**Blocker:** none.

**Major**

1. **Recovery loses row provenance and reintroduces segment-ID collisions.**  
   **Where:** [plan-ms2-3.md:274](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:274), [line 368](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:368).  
   **Lenses:** logic gaps + edge cases. **Blocking.**

   Sequence: open saved `fr` → edit → rename track to `pl` → capture → reload → restore → Save. Dirty rows retain IDs owned by `fr`, but captures currently store the **current editor language** ([use-transcript-recovery.ts:102](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/use-transcript-recovery.ts:102)). The snapshot schema has no separate origin ([recovery.ts:30](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:30)).

   Restoring `originLanguage` from snapshot language therefore produces `pl`. Save skips ID regeneration, although the surviving `fr` transcript still owns those globally unique IDs. PUT deletes only destination-language segments before inserting ([write.go:83](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:83)); the save fails.

   **Required plan change:** preserve genuine row provenance through capture/restore, or regenerate IDs when recovered provenance is uncertain. Define compatibility for existing snapshots. Add **dirty rename → capture/reload → restore → Save** to M3-R11/R6, proving successful destination publication and an unchanged source transcript.

2. **The replacement ASR acceptance can pass while both tracks recognize the same audio.**  
   **Where:** [plan-ms2-3.md:418](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:418), [line 419](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:419).  
   **Lens:** acceptance sufficiency. **Blocking.**

   Two drafts, inventory `[de,en]` and successful GETs do not establish which audio each recognition used. Extracting the default source for both tracks could satisfy these assertions while violating O2. Existing coverage distinguishes original and dub through recognized text ([asr.spec.ts:495](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:495)); the replacement rows omit that proof. GET status alone is also insufficient: an absent transcript returns 200 ([handler.go:77](/Users/strelia/projects/courses/backend/internal/transcript/adapters/http/handler.go:77)).

   **Required plan change:** explicitly retain production-path evidence distinguishing the recordings, then assert each saved language and segment contents against its corresponding editor payload. A separate English-variant fixture can prove source routing if the German hint makes English phrase assertions unsuitable.

**Minor:** none.

The unnecessary-complexity lens returned clean after checking publication reuse, dependency cleanup, per-track persistence/queue changes, adapter ownership and recovery-store reuse.

**Open blocking findings: 2 Majors. VERDICT: BLOCKED.**