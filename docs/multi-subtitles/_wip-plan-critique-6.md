`stage: plan` · Critic: Codex `/root`, selected independent context, round 6. Scope: round-5 finding 1 and new contradictions in the revised D5 lock protocol.

1. **R5 finding 1 — RESOLVED; rebuttal accepted.**

   At [plan.md:173](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:173), the revised contract permits publication after its latest-Ready check succeeds, while stamping **the draft’s generation A**.

   `PromoteAudioTrack(B)` reads and writes asset/audio facts, not transcripts or drafts ([audiotrack.go:370](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:370)). Consequently, the two writers’ results match **publish→promote**: publication succeeds, B becomes latest Ready, and A’s published text reads stale. The stale predicate compares the stamped generation against the latest Ready generation ([transcript.go:220](/Users/strelia/projects/courses/backend/internal/transcript/domain/transcript.go:220)).

   My previous finding imposed the stronger requirement that promotion wait or publication refuse. That requirement is unnecessary for the revised contract.

2. **Nonblocking — qualify “serializability” as the two writers’ outcome guarantee.**

   Location: [plan.md:180](/Users/strelia/projects/courses/docs/multi-subtitles/plan.md:180). Disposition: wording clarification recommended.

   A reader between B’s commit and publication’s commit can see B with the previously committed transcript—or no transcript. Publishing immediately before promotion would expose the newly published text instead. Thus the argument establishes equivalent **completed writer outcomes**, not identical observations for every intervening reader.

   This supplies no wrong-language or false-freshness counterexample: uncommitted publication remains invisible, and A’s text is stale after publication commits. The phase does not promise atomic snapshots spanning playback, transcript inventory and publication.

Downstream assessment after both commits:

| Consumer | Result |
|---|---|
| Player | Selects latest Ready B; transcript selection remains language-based; A-stamped text reads stale. |
| Editor inventory | Published language exists with stale status; Ready-track tabs come from B. |
| ASR pending | B’s tracks remain eligible under D4’s track/generation keys; publication changes A’s draft, not B’s. |

No new substantive contradiction found in D5. The forced-interleaving tests now explicitly cover uploaded-track promotion without eviction and with eviction; these remain planned evidence.

**ACCEPTED — round-6 review scope. Zero open substantive blockers.** Earlier closed findings remain closed. This accepts the revised D5 contract at phase level; slice plans retain their separate critique requirements.

Binding:

- Plan: untracked `docs/multi-subtitles/plan.md`, **r6**, SHA-256 `9d3f0b1af464f14d7f9dfb84b8a53bf03f009bfa3f960633b3629111450cf428`, unchanged on final reread.
- Target: `/Users/strelia/projects/courses`, branch `multi-subtitles`, HEAD `5cbb44bb6d0f9888750f008e50e4cca1f2093c1b`. Tracked backend/frontend sources are clean. Unrelated documentation/configuration WIP and untracked `backend/upload` excluded.
- Raw rounds 1–5 inspected; earlier recorded hashes match. R5 SHA-256: `2e313cc9c68a9a484afc42e3a8e1e15df4b7cc656e01a427224890944c3e8c25`.
- Prompt: candidate `17a9527421e8fac3e43a16c287c404118699833c/prompts/plan-critic.md`, **version 1**, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`; package declares `0.8.0-alpha.1+native-handoff.1`, release commit unverified.

File-based review only; no tests, writes, network probes or graph refresh. Unchanged prior evidence is retained without transferring historical test results. The stronger four-lens gate remains **NOT_RUN: 0 launched / 0 returned**; this scoped verdict does not complete it.

The record could not be saved or linked because the workspace is read-only.