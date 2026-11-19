# e2e-hygiene R5 — report (2026-11-18)

## Done
- **New shared helper.** `frontend/e2e/course-cleanup.ts` (`deleteE2ECourse`) deletes a course and every dependant that
  does not cascade, in one transaction through the portable `docker compose exec postgres` call. It validates that the
  id is a UUID, and each FK it relies on is cited by migration in the file.
- **Callers.** `subtitles.spec.ts` and `audio-tracks.spec.ts` now use the helper; the old `subtitles-cleanup.ts` and the
  inline copy are removed.
- **H1.** `lesson-builder.spec.ts`, `refund.spec.ts` and `teach-editor.spec.ts` register each created course right
  after the create response and delete it in `test.afterAll`. No assertion changed.

## Verification
- `npm run format`, `format:check`, `typecheck` and `lint` exit 0 (Codex run, `_wip-r5.md`).
- **H4 proof.** On the Kiri stack, without `e2e-reset`, after a fresh `--project=setup` run, I ran lesson-builder, refund,
  teach-editor, subtitles, audio-tracks and catalog twice (chromium, `--no-deps`).
  - Each run: 31 passed / 0 failed.
  - Courses `e2e-%` / published `e2e-%` / all: 47 / 7 / 52 before, 47 / 7 / 52 after run 1, 47 / 7 / 52 after run 2.
  - Before the fix, every such run left one published course per spec. Measured 2026-11-18: 28 of 31 published rows
    after four runs, which pushed the seeded course off catalog page 1 (`catalog.spec.ts:3`).
  - The 7 published rows that remain (`e2e-lb` 4, `e2e-te` 2, `e2e-refund` 1) were left by gate runs on the pre-fix code.
- **Independent review.** Codex, `_wip-r5-review.md`: SHIP on both axes, 0 findings.
- **Ordered `make check`:** pending owner consent.

## Limitations (named)
- SQL cleanup leaves image/video/audio bytes on the media volume. They are not counted anywhere, and `e2e-reset`
  removes them.
- **Archived leftovers (outside R5).** Other specs archive their courses instead of deleting them. Grouped count after
  the gate plus the two proof runs: ratings 9, enrollment 8, instructor 5, discussion 4, learning 4, wayforpay 3,
  transcript 2, admin 2, accounts 1, purchase 1. Archived rows stay out of the catalog but are counted by admin
  statistics. Follow-up candidate: move these specs to `deleteE2ECourse`.
