1. **MAJOR — Wave-2 finding #1: CLOSED.** [audio-coverage.ts:50](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/audio-coverage.ts:50), [hls-media.ts:123](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/hls-media.ts:123).

   The counterexample `de → partially appended/failed en → de` now retains `en` among the uncertain span’s possible languages. Returning to `de` establishes a fresh 1.5-second settle window and preserves any later inherited deadline. Helper and adapter regressions assert pending immediately before expiry and confirmation at expiry. These prove the bookkeeping and production adapter wiring; the existing E4b supplies the independent decoded-audio evidence behind the settle policy. **No further code change required.**

The other wave-2 dispositions remain **CLOSED**: portable Compose invocation, exact-language requests beginning before audio confirmation, and scheduled verification excluded from code blockers.

The final commit’s additional fixes are correct by inspection:

- [learn.$.tsx:226](/Users/strelia/projects/courses/frontend/src/routes/learn.$.tsx:226) uses **`isPending`, without `isFetching`**. A held same-language refetch therefore preserves captions and mounted search/follow state, matching D7 and the dated M5 decision. The regression exercises real Query refetch state.
- [subtitles-cleanup.ts:10](/Users/strelia/projects/courses/frontend/e2e/subtitles-cleanup.ts:10) deletes audit, grants, assets and course transactionally. Setup registers the course before subsequent validation can fail; dependent audio, transcript and progress rows cascade. This closes the wave-2 standards review’s retained-fixture finding.

**No new BLOCKER, MAJOR or MINOR findings** in `f50f6ec..af483ce6`, including the final fix delta.

Acceptance coverage inspected:

| Rows | Evidence and assessment |
|---|---|
| M01 | Production HLS event wiring, uncertainty/error handling, actual audio-buffer intersection, reset and cancellation; E4a/E4b/E13 exercise real media. Failed-attempt return regression now discriminates the reported defect. |
| M01b | Reducer rejects foreign mounts and older same-mount generations; accepts a new mount’s lower generation. |
| M02 | Native lifecycle transitions use the explicitly permitted fake `AudioTrackList` seam. |
| M03 | PostgreSQL integration covers playable-asset selection, eviction, Ready filtering, ordering and actor-specific redaction; HTTP tests cover serialization. |
| M04, M08 | Shared case table checks prediction against adapter startup, including case-insensitive whole-tag matching. |
| M05 | State-table tests cover prediction, pending, unavailable, generation rejection and no prediction fallback after load. |
| M06 | Real-route tests cover locked/email-gated/non-ready lessons, primary/exact requests, foreign responses and the `und` family. |
| M07 | Current-element track mode, language remounting and unchanged `captionsOn`; E5 checks restoration through the real player. |
| M09 | Five panel states, retry and both search/follow resets are covered. |
| M10–M14 | E1, E2, E3, E4a, E4b use real playback, authored transcripts, actual caption controls and decoded-audio observations. |
| M15–M19 | E5–E9 cover absent captions, untouched `und`, timing, initial single-track confirmation and transcript-error recovery. |
| M20–M24 | E10–E14 cover navigation, A→B→A, same-instance retry, fingerprinted real MSE append failure/recovery and cross-mount generations. |
| M25 | **PENDING:** full existing E2E set and final ordered gate. |
| M26 | **PENDING:** scheduled design-pass. |

Read-only source review; no tests executed. SHIP applies to the reviewed code, without claiming completion of the pending gates.

VERDICT: SHIP