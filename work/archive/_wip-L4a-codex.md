ДАЛІ: L4a завершено локально; перед передачею — перевірити diff і зберегти без коміту.

Змінено тільки content-gates: playback authorize (media auth_request), playback sessions
(Play/Heartbeat через loadViewablePlayback) і transcript Reading.Get. Усі три спершу кличуть
CanConsumeBy, потім окремо від фактів видимості перевіряють actor.EmailVerified.

Додано CanConsumeBy в playback і transcript; їхні CanStreamBy-обгортки видалено. Contract
access.CanStream не змінювався. CanBeViewedBy не змінювався: architecture AST підтвердив рівно 4.

Denial/allow matrix у доменних і application-тестах: active/blocked member -> доступ; revoked
member -> forbidden; free non-member -> forbidden; paid non-member -> purchase required; published
preview -> доступ, preview unverified/member unverified -> email_not_verified; owner/admin -> доступ.
Оновлено наявні HTTP позитивні fixtures: вони тепер явно active member, а не published stranger.

Перевірено: GOCACHE=/private/tmp/courses-l4a-gocache go test ./internal/playback/... ./internal/transcript/... ./internal/contracts/...; go test ./platform/architecture; go build ./... — green.
Не комічено. generated/sqlc і discussion не чіпалися. Наявні сторонні змінені screenshots лишені як були.
