WORKING FILE: docs/multi-subtitles/_wip-ms4c.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 run (c) — caption/panel state helper, learn route, TranscriptPanel, LessonPlayer captions

Repository `/Users/strelia/projects/courses`, branch `ms4-player-captions`. You are the only writer in
this tree (owner checked). Do not commit, stage, switch branches, or git checkout/reset/stash/clean.
Runs (a) `b397662` and (b) `2297320` are committed: learner lessons carry `audioTracks`; the port has
`onCaptionSelection`; `LessonPlayer` forwards `{generation, selection}` through an optional prop;
`predictedAudioLanguage` exists in `frontend/src/modules/playback/application/audio-language.ts`.

Contract: `docs/multi-subtitles/plan-ms4.md` (ACCEPTED r9). Read §0 (O7), §3 M3–M6, M8 and §4 rows
M04–M07, M09 in full. Key rules, all from the plan:
- Lesson matrix M4: locked → no query (already short-circuited); email-gated → no query, no panel;
  non-ready → `primary` (no language, as today); Ready before the first load of this mount → `exact`
  in the predicted language (remembered preference from `application/preferences.ts`, case-insensitive,
  else default, else first); Ready after load → `exact` in the confirmed language.
- Three query modes `off | primary | exact(language)`; `exact` enabled only with a language; no
  `keepPreviousData`/`placeholderData`; render only when `result.language === language` (lowercased).
- Mount identity M5: route `mount = useMemo(() => Symbol('player'), [lessonId])` passed to LessonPlayer
  (which echoes it in its callback; adjust the (b) prop shape to `{mount, generation, selection}`).
  Reducer: event of another mount than current → drop; record of another mount → accept the first
  event of the current mount without comparing generation; same mount and older generation → drop;
  once a load happened in this mount never fall back to prediction. Computed during render, no reset effect.
- `und` family (`und`, `und-*`) → "none for this track", no request.
- States table M5 (pending/loading, none-for-track, transcript, error with retry → `refetch`); captions
  only in the `transcript` state and never before Play (no media).
- M6: `<track key={language} srcLang={language}>` (remove the `'uk'` default); the `captionsOn` effect
  applies mode to the CURRENT track element via a ref, re-runs on language change; in "none" the gear
  shows a non-interactive text row `learn.video.captionsNoneForTrack` instead of the switch; the VALUE of
  `captionsOn` is never changed by "none". `data-caption-selection` on `<section class="lesson-player">`
  = `pending | confirmed:<lang> | unavailable` (report only).
- M8: `TranscriptPanel` takes `state: {kind:'loading'} | {kind:'none-for-track'} | {kind:'none'} |
  {kind:'error', onRetry} | {kind:'transcript', transcript}`; route keys the panel by language (or
  `'none'`) so search and follow reset. i18n keys in uk and en: `learn.transcript.noneForTrack`,
  `learn.transcript.loadFailed`, `learn.transcript.retry`, `learn.video.captionsNoneForTrack`
  (Ukrainian copy: «Немає субтитрів для цієї доріжки», «Не вдалося завантажити транскрипт»,
  «Спробувати ще», «Субтитри: немає для цієї доріжки»). Loading copy: reuse an existing loading string if
  one exists, else add `learn.transcript.loading` «Завантаження транскрипту…».
- Architecture (AGENTS.md): no fetch in presentation/routes beyond Query options; pure helper in the
  transcript or playback domain/application layer; semantic tokens; ≥44 px controls; visible focus.

## Tests (part of this run)
- Unit: the pure state helper — every row of M4 and M5 tables, mount/generation reducer (other mount;
  A gen 2 → B gen 1 accepted; same mount older gen dropped; never back to prediction after load).
- Route test (`frontend/src/routes`, the existing learn route test pattern): matrix M4 incl. that the
  `primary` request actually happens for a non-ready lesson; exact request with the predicted language
  before Play; foreign-language response never rendered; `und` → no request; error → retry refetches.
- Unit LessonPlayer: `<track>` recreated on language change; mode applied to the current element;
  `captionsOn` value unchanged across "none"; gear text row in "none"; `data-caption-selection`.
- Unit TranscriptPanel: five states; search AND follow reset when the key changes.

## Verify before finishing
`cd frontend && npm run format && npm run format:check && npm run typecheck && npm run lint && npx rstest run src --testTimeout 30000`
(the host is heavily loaded; the 5 s default times out there — report both if you also run the default).
Report each command and its real result; a skipped check is not a pass. Do not run e2e (run d).

## Two channels
- Missing or contradictory contract → STOP, QUESTION in the working file.
- An inaccurate plan statement that would cause a bug → argue with evidence (DECISION).
