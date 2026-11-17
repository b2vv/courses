# MS2a — Codex review w1 (read-only, 2026-11-15)

- **Minor — [00031_asr_draft_per_track.sql:21](/Users/strelia/projects/courses/backend/migrations/00031_asr_draft_per_track.sql:21): backfill uniqueness is assumed, not enforced.** The existing schema permits two embedded tracks on one asset with different languages, both using `source_leaf='original'`. `UPDATE … FROM` would then choose an unspecified track and language snapshot. Current production writers appear to prevent this shape, so this is non-blocking hardening: reject ambiguous matches explicitly or enforce their uniqueness.

- **Minor — [draft_integration_test.go:639](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/draft_integration_test.go:639): rollback selection is untested.** The populated Down/Up fixture retains only one draft per lesson. It never exercises Down’s default preference, newest-attempt fallback, or track-ID tie-break. Add multiple same-lesson drafts before Down and assert the survivor. The SQL ordering itself is deterministic.

No Blocker or Major findings. Claim validates the source tuple and latest Ready asset within its write; the later diagnostic read cannot authorize an invalid write. Pending implements O7, and Complete/Recover fence attempts per track. Task identity, payload propagation, terminal Claim-error handling, architecture boundaries, and adapter mappings match the scoped decisions.

M2-R1..R4 and R7..R9 exercise production migrations/repository methods, subject to the rollback coverage gap above. Caller searches found no consumers of removed publication endpoints, incompatible seed inserts, or outside-module draft readers. Existing C06/S6 e2e expectations require the replacements already scheduled in MS3.

Review was static and read-only; supplied runtime results were not independently rerun. Full e2e and mutation checks remain unverified here.

**VERDICT: SHIP**
## Resolution (Courses 1, 2026-11-15)
- Minor 1: closed — 00031 adds UNIQUE audio_tracks_leaf_per_asset (video_asset_id, source_leaf) before the backfill; Down drops it.
- Minor 2: closed — M2-R1 seeds a default-vs-newer pair (lesson B) and two non-default drafts (lesson C) before Down and asserts default-first then newest survivors; PASS.
- Full make integration rerun after the fix: GATE_EXIT=0, 67 ok packages, 0 FAIL.
