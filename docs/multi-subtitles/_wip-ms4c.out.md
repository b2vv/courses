MS4(c) реалізовано: helpers, route matrix, п’ять станів панелі, keyed captions і збереження `captionsOn`. Додано тести навігації та пізніх callback.

Фінальні перевірки у `frontend/`:

| Команда | Результат |
|---|---|
| `npm run format` | PASS, exit 0 |
| `npm run format:check` | PASS, exit 0 |
| `npm run typecheck` | PASS, exit 0 |
| `npm run lint` | PASS, exit 0 |
| `npx rstest run src --testTimeout 30000` | PASS, 1415/1415, exit 0 |

[Журнал роботи](<repo>/docs/multi-subtitles/_wip-ms4c.md) доповнювався після кожного кроку; phase report оновлено. E2E залишено для run (d). Staging і commit не виконувалися.