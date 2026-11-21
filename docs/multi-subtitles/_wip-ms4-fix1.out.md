У [журналі](<repo>/docs/multi-subtitles/_wip-ms4-fix1.md) зафіксовано всі 5 виправлень і red→green регресії.

Усі команди завершились з exit 0:

- `npm run format`
- `npm run format:check`
- `npm run typecheck`
- `npm run lint`
- `npx rstest run src --testTimeout 30000` — **1432/1432**, без пропусків.

Комітів і staging немає. Rebuild та browser checks не запускав; наступний крок — owner-сесія на Kiri.