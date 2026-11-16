# MS3a — run 2 brief for Codex: language tabs in the lesson editor

Repository: current directory, branch `ms2-3-asr-per-track` @ `a985543` + the uncommitted MS3a run 1 (data layer,
per-track draft queries, inventory, `draftLoadable` guard, "Multiple languages" switch — read
`docs/multi-subtitles/_wip-ms3a-run1.md` for what exists and where). Shared working tree with other sessions: do NOT
run git checkout/reset/stash/clean/rm, do not commit, do not touch generated files, e2e, or backend.

WORKING FILE: docs/multi-subtitles/_wip-ms3a-run2.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md`; `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED) — authoritative: **MS-D6** (all of it), **MS-D7**,
  **MS-D9 key part only**, **MS-D10** "tab shows the flag"; rows **M3-R1, M3-R2, M3-R3, M3-R5, M3-R7, M3-R10**, and the
  save-path parts of **M3-R11** (new ids) and **M3-R13** (token branches a–d).
- `docs/multi-subtitles/report.md` → "MS3a … checkpoint": MS3a leaves to MS3b the `und` confirm dialog (M3-R4),
  detached snapshots / flush on unmount (M3-R6, M3-R12) and `baseLanguage` in the snapshot. Interim rule for MS3a: an
  `und` tab can Load its draft, but Save is refused with a short message ("confirm the language — coming next" is NOT
  the copy; say what the author must do: «Спершу вкажіть мову доріжки в Audio tracks» / "Set the track's language in
  Audio tracks first"), and sends 0 PUT.

## Scope
1. **Tabs component** in `frontend/src/modules/admin/ui` (it combines media and transcript; modules never import each
   other — admin may import both public `index.ts` APIs). Mounted by `CourseEditorView` instead of the separate
   `LessonDraftField` + `LessonTranscriptField` pair. Replace the `// MS3a run 2: tabs` bridge spots from run 1.
   - With a Ready asset: one tab per `ready` track of the latest Ready asset in the order default → embedded → rest by
     language → id, plus one tab per inventory language without a track (orphan). Identity `track:<id>` / `lang:<lang>`.
   - Label: language name (reuse the S5 language list/labels used by `LessonAudioTracksField`) or the tag; «Невідома
     мова» for the `und` family; a "default" marker; the `multilingual` flag next to the track language (MS-D10).
   - No Ready asset: one editor without tabs, exactly as today (`instanceId` empty, ids unchanged). The existing
     `transcript.spec` must stay green without edits — keep its ids, roles and copy.
   - `Aria.Tabs` with `shouldForceMount`; inactive panels `inert` and hidden by a LOCAL rule
     `.lesson-language-panel[inert] { display: none }` (not a global rule). Gotcha: Aria.Tabs + shouldForceMount keeps a
     single `<form>` — check the editors are not nested in an outer form that would submit the wrong panel.
2. **Track tab content.** `LessonDraftField` of that track (state, Load draft, Regenerate on that trackId, no language
   Select) + `TranscriptEditor` with language = current `lower(track.language)`, fixed (no language picker in a tab).
   Read `?language=` of that language. `TranscriptEditor` and `LessonDraftField` take `instanceId` (= tab identity); every
   DOM id is built from `lessonId + instanceId` (heading, recovery, draft offer, labels; `TranscriptEditor.tsx:166,176,208`).
   Recovery key per tab: `transcriptHistoryKey` gains the tab part (`user:lesson:track:<id>` / `user:lesson:lang:<lang>`;
   no tab → `user:lesson` unchanged).
3. **Editor key and base (MS-D6 "Зміна мови T").** The editor's React `key` is the tab identity only — never the
   language or token (today `LessonTranscriptField.tsx:87` keys by token). Base (language, token, transcript lines) comes
   in by props. Clean editor + new base → lines replaced. Dirty editor + new base → lines kept, message «мова доріжки тепер
   X; рядки буде збережено як X».
4. **Guard (MS-D7, M3-R3).** Load draft only when run 1's `draftLoadable(draft, track)` is true; otherwise the text
   «Чернетка розпізнана для іншої мови чи режиму доріжки — натисніть Regenerate». Remove the old
   `draftOffer.languageMismatch` branch. Stale after rename/flag change → that text, no automatic re-recognition (O7).
5. **Save (MS-D6 "Save: свіже читання", M3-R10, M3-R11, M3-R13).** In a tab, before PUT:
   - direct repository reads (NOT the Query cache) of the asset's tracks and the transcript inventory; PUT only if both
     returned ok. Network error / timeout / error result → repeatable error «Не вдалося перевірити доріжку — спробуйте ще»,
     lines kept, 0 PUT.
   - track gone, or `lower(language)` ≠ save language → refuse with a message, 0 PUT. Orphan tab: inventory only.
   - token: save language = base language → PUT with the BASE token (inventory never advances it silently); a newer
     write → server `transcript_modified` as today. Save language L ≠ base language: L absent in fresh inventory → PUT
     without token (`transcript_exists` if created meanwhile); L present → confirm «Транскрипт мовою L (оновлено <час>)
     уже є — замінити його цими рядками?», PUT with exactly that token.
   - every PUT gives each line a new `crypto.randomUUID()` id, wherever the lines came from.
   - orchestration of the two reads lives in `app/composition` or admin UI with injected functions — not in presentation
     fetching. After a successful save: invalidate the inventory and the lesson transcript query.
   - `onTranscriptKnown` = inventory non-empty.
6. **Copy.** i18n keys in every locale the editor already uses; keep copy short and say what to do.

## Tests (rstest + Testing Library; update what you break)
Unit: tab set and order incl. orphan and no-Ready fallback (M3-R1); tab save language and PUT language (M3-R2); guard
texts (M3-R3); per-track Regenerate without Select and per-track pending (M3-R5); no duplicate ids across mounted
editors (`querySelectorAll('[id]')`), inactive panel `inert`, accessible names point at the active panel (M3-R7);
save re-read: rename/delete after Load → refuse, tracks/inventory read failing → error, **PUT counter = 0** (M3-R10);
new ids on every PUT (M3-R11); token branches (M3-R13 a–d); `und` tab Save → 0 PUT + message.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint && npm test` — exit codes and counts; changed files.
If the slice is too big for one run, finish the implementation with passing existing tests first, list the missing new
tests as NEXT, and stop — Courses 1 runs a follow-up. Plan ambiguous or a bug → QUESTION and argue.
