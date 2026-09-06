# Report M01 — media ADRs

- `docs/adr/008-tus-ingest.md` chooses the bounded tus core in a third executable, the
  `video_assets` row as the upload capability, the JSON mint as the only creation point, and
  records the Q1 completion-path deviation and dated Q2 container/codec split with their
  triggers.
- `docs/adr/009-fmp4-cmaf-segments.md` chooses fMP4/CMAF `.m4s` segments and the H.264/AAC
  ladder, supersedes only ADR 005's MPEG-TS segment clause, records ffmpeg 5.1.9's verified
  `init_N.mp4` naming and playlist gaps, and preserves the three-condition AV1 trigger.

Verified all existing ADRs before matching their compact heading, status, decision, rationale,
and consequence format. Cross-checked the decisions against `docs/media/plan.md` D-2, D-8,
D-11, D-16, and M01; `docs/media/spec.md` Q1–Q3 including the 2026-09-06 Q2 amendment;
`work/PRODUCT-BRIEF.md` §2, §11, and §24.11–§24.14; and
`docs/media/research-codecs.md`. Confirmed ADR 009 names ffmpeg 5.1.9 as the shipping muxer
binary; both ADRs keep `/media/*` refused during Phase 5 and never claim that the Go API receives
video bytes or that playback exists. `.claude/standards.md` already covers both files through
its `docs/adr/*.md` row, so it required no change. No Docker verification was needed for this
documentation-only task.

## Reviewer verification (measured in the reviewing session)

Both ADRs match the repo format (compared against 001 and 005). ADR 009 correctly names
the SHIPPING binary — ffmpeg 5.1.9 from `debian:bookworm-slim`, verified in
`courses-media-worker-1` — rather than the host's 7.1.1, which is the discipline GATE 2
round 5 established. Two review additions: dates on both status lines, and a
"Superseded in part" back-reference in ADR 005 pointing at 009, so a reader of 005 learns
its segment clause is replaced without having to find 009 first. Docs-only change; no gate
surface touched. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and verified
here.
