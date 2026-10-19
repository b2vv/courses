ДАЛІ: розширити identity domain, SessionRepository та API adapter для self-service операцій; потім додати application-команди й тести.

2026-10-05: прочитано CTO briefing, A07 brief, чинні session/domain/transport seams і OpenAPI. Generated Orval містить потрібні операції; його не буде змінено.
2026-10-05: розширено доменний session contract (`emailVerified`, rememberMe, self-service inputs/errors), порт, Orval-based adapter та application session commands. Початковий `npm run typecheck` після оновлення наявних фікстур — exit 0.
2026-10-05: додано перевірки repository/application (business-code mapping, нейтральний forgot, rememberMe, `emailVerified`) і `identity.refresh(client)` для мутацій A08, що мають перечитати `/me` після verify/email change.
2026-10-05: `npx oxfmt` на зачеплених TS-файлах, `npm run lint` (включно з architecture) і `npm run typecheck` — exit 0; `git diff --check` чистий. Цільовий і повний `npm run test` не стартують у sandbox до тестів: `listen EPERM ::1:3000`. Ручний перегляд diff: generated/UI не змінені, усі self-service виклики проходять через Orval + request-scoped transport, а зовнішній user/empty response парсяться zod. Коміт не створено.
