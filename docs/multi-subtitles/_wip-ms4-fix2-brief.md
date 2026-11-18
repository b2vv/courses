WORKING FILE: docs/multi-subtitles/_wip-ms4-fix2.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 fix 2 — review wave 2 (both axes BLOCK) on `ms4-player-captions` @ 3da8fb0f

Repository `/Users/strelia/projects/courses`. Only writer in this tree. No commits, no staging, no branch
switch, never git checkout/reset/stash/clean/rm. Read `docs/multi-subtitles/_wip-ms4-review2-codex-spec.md`
and `_wip-ms4-review2-codex-standards.md`. Fix exactly these, each with a test that fails before the fix:

1. **Possible languages of uncertain spans** (spec w2 #1, `audio-coverage.ts:55`, `hls-media.ts:125`). Record the
   attempted language at APPENDING (the adapter knows `hls.audioTracks[frag.level]` then; pass it through
   `captionLanguageOf`). A span's "possible languages" = its attempted language (when the attempt is not a
   clean success) ∪ its inherited `replaced` history. A new attempt's `replaced` = completed-clean labels it
   overlaps ∪ possible languages of overlapping uncertain spans; the settle guard applies whenever that set
   contains a language other than the new one. Unit regression: `de` [0,6) → `en` APPENDING + APPENDED + ERROR
   near playhead 2.5 → `de` re-append succeeds at 2.6 → pending until ≥ 2.6 + 1.5.
2. **Owner decision (Courses 1, 2026-11-18): a same-language background refetch is not "loading".**
   `learn.$.tsx:226` passes `isPending || isFetching`. With `isFetching`, a window-focus refetch of the SAME
   language removes captions and remounts the panel (search/follow lost) — the reviewer confirmed the
   mechanism. D7's loading state is for a NEW language, which is a new query key with no data, i.e.
   `isPending`. Use `isPending` only (no `keepPreviousData` — a new language still has no data and still shows
   loading). Route test: data present for `de`, a refetch in flight → transcript and captions stay, search
   text in the panel is kept. Add one line to `docs/multi-subtitles/plan-ms4.md` M5 recording this dated
   decision.
3. **e2e cleanup of globally counted fixtures** (standards w2 #5, `subtitles.spec.ts:646`). Follow
   `frontend/e2e/audio-tracks.spec.ts:26-31,65,227-240`: register each created course for deletion before
   anything can fail, and in `afterAll` delete it through SQL (grant_audit → course_grants → courses, plus any
   other rows the spec creates that do not cascade: check `progress`, media assets, transcripts against the
   FK definitions in `backend/migrations`) with the same portable `docker compose exec` call. Verify after a run
   that no `e2e-subtitles*`/your fixture slugs remain (query, record the count).

## Verify
`cd frontend && npm run format && npm run format:check && npm run typecheck && npm run lint && npx rstest run src --testTimeout 30000`.
Do not rebuild images; stop and report — the owner session rebuilds `frontend` and runs the browser checks.
Report each command's real result.

## Two channels
Missing/contradictory contract → STOP, QUESTION. A finding you believe is wrong → argue it with evidence (DECISION).
