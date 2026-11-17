# Delta review: package `ms3d-fix-w2`

You are the same independent reviewer as w1 (`docs/multi-subtitles/_wip-ms3d-fix-review-codex-w1.md`, BLOCK, 1 Major:
destination fell back to orphan tabs with zero tracks and to the first track with no default). Read-only; no edits, no git
writes, no Docker. Rstest cannot listen in your sandbox; `npm run typecheck` works.

- Delta: uncommitted `git diff -- frontend/` (frozen `docs/multi-subtitles/_wip-ms3d-fix-review-diff-w2.patch`, sha256 `61b2e0c186f00eb0fea78cf9f4fd039947650f744c5915fc3edfc812ee24075f`).
  Since w1 only `frontend/src/modules/transcript/application/recovery.ts` (`detachedTranscriptSnapshots`) and its test changed.
- The author's decision and its mechanism are in the last MS3d-fix entry of `docs/multi-subtitles/report.md`. Verify the
  mechanism claim (tabs only with a Ready video; one Ready default track) against `LessonLanguageTabs.tsx` and
  `backend/internal/media/domain/audiotrack.go`.

Confirm the Major is closed, that dropping (not offering) a snapshot with no default track loses nothing (key not deleted,
offered later), and that the single-editor (no tabs) path is unchanged. Finding format as w1. End with
`VERDICT (ms3d-fix-w2): SHIP | BLOCK` and counts.
