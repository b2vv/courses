Wave-1 disposition:

- **#1 OPEN.** The original same-language retry counterexample is fixed, but failed attempts still lose a possible audible language. See finding 1.
- **#2 CLOSED.** `subtitles.spec.ts:89` now invokes portable `docker compose` with the inherited environment.
- **#3 CLOSED.** `pending.requested` starts the exact-language request before confirmation. The route test verifies the HTTP query and keeps the resolved transcript hidden until confirmation.
- **#4 CLOSED as a code blocker.** M25, M26, and the final ordered gate are **PENDING**, as explicitly scheduled by the owner.

1. **MAJOR — A partially successful failed append can bypass the replacement settle window when returning to the original language.** [audio-coverage.ts:55](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/audio-coverage.ts:55), [hls-media.ts:125](/Users/strelia/projects/courses/frontend/src/modules/playback/adapters/hls-media.ts:125).

   Trace: `de` covers `[0,6)`. An `en` attempt emits APPENDING, successfully appends a chunk, then receives a nonfatal fragment ERROR. Its span becomes uncertain and retains `replaced = ['de']`; the attempted `en` language is never recorded because language is supplied only at FRAG_BUFFERED. The viewer returns to `de`, and a replacing `de` append succeeds near playhead 2.6. The inherited history contains only `de`, so `replacesAnother` is false and no new settle window is established. The map immediately confirms `de`, although already decoded `en` samples from the failed attempt may still sound—the same decoder mechanism documented in the dated E4b decision.

   **Required change:** retain the attempted language alongside inherited possible languages for uncertain spans, and apply the settle guard whenever those possibilities include another language. Add a discriminating regression for `de → partially appended/failed en → de`. The new “audio event tap” test at `hls-media.test.ts:428` uses `FakeHls`; it cannot establish decoder behavior. The browser recovery test still waits before recovery and does not check immediate recovery continuously against decoded audio.

Acceptance coverage checked by source inspection:

| Rows | Assessment |
|---|---|
| M01 | **Incomplete:** finding 1; existing E4a/E4b/E13 exercise real media, but miss this failed-append return path. |
| M01b, M05 | Reducer/helper tests cover foreign mounts, lower generations, fresh A→B→A identity, and no return to prediction after load. |
| M02 | Native lifecycle transitions use the explicitly permitted fake `AudioTrackList` seam. |
| M03 | Real PostgreSQL tests cover asset selection, eviction, Ready filtering, ordering and actor-specific redaction; HTTP tests cover serialization. |
| M04, M08 | One case table exercises prediction and adapter startup, including case-insensitive whole-tag matching. |
| M06 | Route tests cover locked/email-gated/non-ready lessons, exact queries, foreign responses and the `und` family. |
| M07 | Player tests cover current-track mode, language remounting and persisted `captionsOn`; E5 checks return from a language without captions. |
| M09 | Five panel states, retry/live roles, search and follow reset have component tests. |
| M10–M14 | E1, E2, E3, E4a and E4b use real playback and decoded-audio observations. |
| M15–M19 | E5–E9 cover absent captions, untouched `und`, timing, initial single-track confirmation and transcript-error recovery. |
| M20–M24 | E10–E14 cover navigation, fresh mounts, same-instance retry, real append failure and cross-mount generation changes. Finding 1 remains outside their discriminating coverage. |
| M25, M26 | **PENDING**, without blocking the code merely because these scheduled checks are absent. |

The same-language background-refetch behavior is real: `learn.$.tsx:226` includes `isFetching`, so refetching removes captions and unmounts `TranscriptLines`, resetting search/follow. This is broader than D7’s wording about a **new-language** request, but MS4 M5 explicitly maps an in-flight query to loading. I therefore do **not** classify that documented behavior alone as a specification violation.

No tests were executed during this read-only review.

VERDICT: BLOCK