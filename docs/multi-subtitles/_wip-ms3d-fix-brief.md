# MS3d-fix brief for Codex — two PR #109 review defects (frontend transcript recovery)

Repository: current directory, branch `ms2-3-asr-per-track` @ `38c7915`. Shared working tree: do NOT run git
checkout/reset/stash/clean/rm/add/commit. Touch only `frontend/src/modules/transcript/**` (and
`frontend/src/modules/admin/ui/LessonLanguageTabs.tsx` + its test only if strictly needed). No backend, no generated files,
no e2e, no Docker.

WORKING FILE: docs/multi-subtitles/_wip-ms3d-fix.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line. Then FACT / DECISION /
NEXT / QUESTION. Final answer = concise summary of the file.

Read first: `docs/multi-subtitles/_wip-pr109-review-codex-standards.md` (finding 1) and
`docs/multi-subtitles/_wip-pr109-review-codex-spec.md` (F1); spec `docs/multi-subtitles/plan-ms2-3.md:398-415`; the
MS3d-fix checkpoint at the end of `docs/multi-subtitles/report.md`.

## 1. Capture version collision (Standards finding 1)

`capturedAt = Date.now()` (`ui/use-transcript-recovery.ts:158`) is used as the snapshot's identity. Two captures of the same
key in the same millisecond share it, so cleanup owned by the first deletes the second, violating
`ports/TranscriptHistoryStore.ts:16`.

- Add `revision: string` to `TranscriptSnapshot`, generated uniquely per capture (`crypto.randomUUID()` is already used in
  `ui/TranscriptEditor.tsx`; reuse the same approach). `capturedAt` stays as time for sorting/labels only.
- Move **every** identity comparison from `capturedAt` to `revision` — the class, not one place: `clearIfVersion` (port,
  memory adapter, IndexedDB adapter), the hook's `seen`/`written` ownership, `clearAcceptedTranscriptSnapshot`,
  `AcceptedTranscriptSource` (key + revision), and any other `capturedAt ===`/`!==` comparison you find
  (`grep -rn capturedAt frontend/src/modules/transcript`). List each site you changed in the journal.
- IndexedDB schema: `revision` required; a stored snapshot without it fails parsing → read returns null (no copy). Pre-launch:
  no migration shim. Keep the pre-list `acceptedSource` transform only if it still parses meaningfully; otherwise drop it and say why.
- Tests (discriminating): equal `capturedAt`, different `revision` → `clearIfVersion` with the first revision returns false and
  the second snapshot survives — for the memory adapter, the IndexedDB adapter (existing fake-indexeddb test setup if present)
  and the hook-level ownership path. Update existing fixtures to carry `revision`.

## 2. Detached-snapshot destination (Spec F1)

`application/recovery.ts:232` matches all tabs with the language and falls back to `matches[0]`, so an orphan `lang:*` tab can
receive a detached snapshot. Spec: the **track** tab whose current language equals `snapshot.language` (case-insensitive), else
the default-track tab.

- Restrict matches to track tabs. Prefer an explicit tab kind over the `instanceId.startsWith('track:')` string test if the tab
  type can carry it cheaply (`TranscriptRecoveryTab`, built in `LessonLanguageTabs.tsx:157-162`); otherwise keep the prefix but
  in one named helper.
- Fallback: default track; if no tab is default, the first track tab; only with no track tabs at all, the existing behavior.
- Tests in `application/recovery.test.ts`: tabs `track:en` (default) + `lang:uk` orphan, snapshot `language: 'uk'` (detached key,
  e.g. `user:lesson:track:gone`) → `track:en`. Keep the existing cases green.

## Done = report

`cd frontend && npm run format && npm run typecheck && npm run lint && npx rstest run src/modules/transcript src/modules/admin`
(use the repo's actual test command if different — check `package.json`) exit codes and pass counts; for each new test, a
mutation check in the journal: revert that one fix line temporarily → the named test fails (copy the failing line) → restore;
changed files + SHA-256 of each.
