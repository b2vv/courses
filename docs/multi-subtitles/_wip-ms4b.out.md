Implemented MS4b; details in the append-only [working file](<repo>/docs/multi-subtitles/_wip-ms4b.md).

- `npm run format` — EXIT 0
- `npm run format:check` — EXIT 0
- `npm run typecheck` — EXIT 0
- `npm run lint` — EXIT 0
- `npx rstest run src/modules/playback src/routes` — EXIT 1; latest exact retry: 160/164 passed.

All playback tests passed. Supplemental single-worker run: 162/164; two admin route failures remain, so the broader gate is not green.

No staging, commits, or branch changes. MS4a preserved; run-c behavior remains untouched.