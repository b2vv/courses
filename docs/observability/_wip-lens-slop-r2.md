# Lens: AI-slop — round 2 review of docs/observability/plan.md

ДАЛІ: прочитати plan.md + spec.md, звірити round-2 доповнення (двохрівневий §17 scrub, dedup signature+TTL+StrictMode, server-entry wrap, окремий ssr-logger, monitoring-stack.sh) проти реальних викликів/замовника; перевірити чи reuse-first справжній.

ФАКТ: server.ts = bare re-export (+zod-jitless import) → реального 500-сема немає; wrap обґрунтований.
ФАКТ: transport.ts і request-headers.ts існують → reuse-first справжній, не вигадана абстракція.
ФАКТ: image-authorize.sh / playback-authorize.sh існують → monitoring-stack.sh дзеркалить наявний патерн out-of-gate proof, не церемонія.
ФАКТ: login-req.e2e.conf + api-req.e2e.conf існують → clientlog e2e-variant дзеркалить наявний патерн.
ФАКТ: немає глобального reporting error boundary / window.onerror у frontend/src (errorComponent — це per-route TanStack fallback, інший концерн) → новий boundary+bridge не реінвент.
ФАКТ: EventSink port ДРОПНУТО (B-NB), batching ДРОПНУТО (B8) — це анти-slop ходи round-2.

РІШЕННЯ (slop-лінза):
[AI-slop NB] Клієнтський pre-send §17 scrub = третя копія redaction.
  де: New code → Frontend, рядок "A client-side pre-send scrub mirrors the backend never-log set as defence in depth [G2:NB AC2]"
  чому: backend уже скрабить на field+value рівні ПЕРЕД emit (B5), трафік in-network. Клієнтський scrub дублює never-log set у TS (drift між TS і Go), а marginal threat (секрет перетинає in-network HTTP до власного backend, що одразу скрабить) слабкий. "defence in depth" як обґрунтування без окремої загрози = тонка церемонія.
  що зробити: або прибрати клієнтський scrub (backend-scrub закриває AC2), або назвати конкретну загрозу, яку backend-шар НЕ закриває (напр. секрет у transport-логах nginx/proxy ДО ingest), інакше лишити один авторитетний never-log set.
  вага: non-blocking

[AI-slop NB] StrictMode як драйвер dedup — не ганяється авторитетним гейтом.
  де: New code → Frontend, "dedup ... that also absorbs React StrictMode's dev double-invoke"
  чому: StrictMode double-invoke — dev-only; e2e (AC1) ганяється на prod-білді (repo-норма), тож цей драйвер гейт ніколи не вправляє. Сам dedup виправданий реальним prod-овелапом (boundary-caught error React форвардить у window.onerror), але StrictMode-аргумент = наповнювач, що роздуває обґрунтування.
  що зробити: лишити dedup, але прибрати StrictMode як мотивацію (або позначити dev-only); тримати обґрунтування на prod boundary→onerror овелапі.
  вага: non-blocking

ФАКТ: narrative-неточність "channels separated to avoid double-count" — розділення каналів призначає ownership, double-count знімає саме dedup; не slop-робота, лише формулювання. Залишаю як зауваження, не знахідку.

ВИСНОВОК: решта чиста. Перевірено: §17 two-level scrub (закриває реальний B5-gap, не церемонія), server-entry wrap (реальний брак сема), ssr-logger окремим leaf (бо Go-логер недоступний у Node — мінімум нового коду), monitoring-stack.sh (наявний патерн), promtail source-label pipeline (реальний механізм, бо client-лінії виходять через api-контейнер), dropped EventSink+batching (анти-slop).
2 non-blocking, 0 blocking.
