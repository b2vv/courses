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
- **Ordered `make check`** on `328c9d57` (owner consent 2026-11-19, 11:38→12:11): **GATE_EXIT=0** — unit 1482 passed /
  0 failed / 0 skipped, Go 130 ok / 0 FAIL, real `e2e-reset` (postgres/redis/media volumes removed), e2e **168 passed**.
  After that full run on the fresh database: published `e2e-%` courses **0**, all `e2e-%` courses 40 (archived, see
  limitations).

## Limitations (named)
- SQL cleanup leaves image/video/audio bytes on the media volume. They are not counted anywhere, and `e2e-reset`
  removes them.
- **Archived leftovers (outside R5).** Other specs archive their courses instead of deleting them. Grouped count after
  the gate plus the two proof runs: ratings 9, enrollment 8, instructor 5, discussion 4, learning 4, wayforpay 3,
  transcript 2, admin 2, accounts 1, purchase 1. Archived rows stay out of the catalog but are counted by admin
  statistics. Follow-up candidate: move these specs to `deleteE2ECourse`.

# e2e-hygiene R6 — report (2026-11-19)

## Done
- **H5.** Ten specs (ratings, enrollment, instructor, discussion, learning, purchase-wayforpay, admin, transcript,
  purchase, accounts) register every course they create right after the create response's status assertion, and
  delete it in `test.afterAll` through `deleteE2ECourse` (`Promise.allSettled` + `AggregateError`). 15 creation paths,
  including transcript's draft and the admin UI create dialogs. The admin student negative-create registers an id only
  on an unexpected 201, best-effort, so the 403 assertion stays the primary failure.
- Cleanup-only archive calls are removed. Archive calls that are the behaviour under test (admin P20, ratings
  Rt26/Rt30) stay; their courses are deleted afterwards. Seeded fixtures are not touched. `course-cleanup.ts` unchanged.
- **H6.** No assertion changed (AST comparison of expect chains against HEAD: 530 unchanged, `_wip-r6.md`).

## Verification
- `npm run format`, `format:check`, `typecheck`, `lint` exit 0 after the final edit.
- **H7 proof.** On Kiri, without reset, after `seed-identities` and a fresh `--project=setup` each time, the ten specs +
  catalog ran twice (chromium, `--no-deps`), 55 passed / 0 failed each run. `e2e-%` courses by status: archived 39 /
  draft 1 before, after run 1 and after run 2 — the same for the first implementation and again after the F1/F2 fix
  (four runs in total). The 40 rows were left by earlier gate runs on the pre-R6 code. The final F3 edit (a branch taken
  only on an unexpected 201) ran only inside the gate below.
- **Independent review.** Codex, `_wip-r6-review.md`: BLOCK (F1 status assertion masked by body parse at 13 sites;
  F2 unexpected-201 negative-create not cleaned up) → fixed → BLOCK (F3: the F2 branch masked the 403 assertion) →
  fixed → **SHIP** on both axes.
- **Ordered `make check`** on `3aa5eaca` (owner consent 2026-11-19, 21:26→21:57): **GATE_EXIT=0** — unit 1482 passed /
  0 failed, Go 130 ok / 0 FAIL, real `e2e-reset` (postgres/redis/media volumes removed), e2e **168 passed**.
  **H8:** after that full run on the fresh database, `e2e-%` courses = **0** in every status.

## Limitations (named)
- As in R5: SQL cleanup leaves image/video/audio bytes on the media volume; `e2e-reset` removes them.
- Playback's own SQL cleanup (not in R6 scope) removes grants and courses but does not delete `video_assets` rows
  explicitly; no residue measured (H8 counts courses only).
