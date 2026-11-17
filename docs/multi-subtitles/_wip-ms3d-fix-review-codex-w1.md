**One Major blocker: destination selection still exceeds the accepted contract.**

1. **Major — BLOCKER — confidence: high.** [recovery.ts:230](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.ts:230) restores orphan tabs as destinations when `tracks.length === 0`. With only `lang:uk` present, a detached `language: 'UK'` snapshot is offered there. The checkpoint explicitly says an orphan is **never** a destination ([report.md:796](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:796)).

   **Impact:** zero-track input retains the prohibited orphan placement. The adjacent fallback also chooses the first non-default track when no default exists, rather than following the specified default-track rule.

   **Verified:** executed the production placement function directly:
   
   | Tab inventory | Destination |
   |---|---|
   | Default `en` track + matching `uk` orphan | `track:en` — correct |
   | Only matching `uk` orphan | `lang:uk` — contract violation |
   | No tabs | `''` — existing single-editor behavior |
   | Non-default `en` track + matching orphan | `track:en` — unspecified fallback |

   The new [recovery.test.ts:446](/Users/strelia/projects/courses/frontend/src/modules/transcript/application/recovery.test.ts:446) explicitly expects the orphan placement. Remove these fallback exceptions or obtain a recorded contract change. The backend Ready-asset invariant normally supplies a default track, limiting normal production reachability; the requested zero-track contract remains unsatisfied.

The other checks found:

- **Identity migration is complete.** No production `capturedAt` identity comparisons remain. Both adapters, hook ownership, accepted-source cleanup, ancestry deduplication and Discard use `revision`. The sole production capture writer generates a fresh UUID per write.
- **Tests discriminate the timestamp collision.** Memory and IndexedDB replacements preserve equal timestamps with different revisions; the hook freezes `Date.now()` and checks distinct generated revisions and competing-window survival. The default-track-plus-matching-orphan fixture discriminates the original F1. No existing assertion weakening was found.
- **Revisionless persistence behaves as authorized:** reads return `null`, preventing recovery. `keys()` remains independent of schema parsing, so viewer sweeps still delete foreign revisionless entries. The planted entries at `transcript.spec.ts:567` therefore remain useful sweep fixtures. Inspected ASR scenarios capture through the editor rather than planting timestamp-only snapshots.
- **UUID availability:** this remains a secure-context API dependency already used by transcript row creation and Save. Capture catches UUID failures, but recovery then writes nothing. Browser availability was not exercised here.
- No additional TypeScript or Architecture violations found.

`npm run typecheck` **passed**. Rstest **could not start**: sandbox denied `listen ::1:3000`; zero tests executed. No edits, git writes or Docker operations. The frontend diff still matches frozen SHA-256 `20279856c7e02377cf610a463ead11c96ecf8b1be8db480efcc80dccfe1a6165`.

Counts: **Critical 0 · Major 1 · Minor 0**.

**VERDICT (ms3d-fix-w1): BLOCK**