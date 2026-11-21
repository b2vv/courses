MS4a завершено. SQL-фікстуру виправлено; результати дописано в [working file](<repo>/docs/multi-subtitles/_wip-ms4a.md).

Усі перевірки пройшли (`EXIT=0`):

- `go test ./internal/course/...` — із тимчасовим `GOCACHE`.
- `go test -tags integration ./internal/course/...` — реальний Postgres через Kiri.
- `make generate`.
- `npm run format`.
- `npm run format:check`.
- `npm run typecheck`.
- `npm run lint`.
- `npx rstest run src/modules/playback src/routes` — 116/116.

Статус generated-файлів записано. Комітів і staging немає; гілка незмінна. Повний e2e та незалежне рев’ю не запускалися.