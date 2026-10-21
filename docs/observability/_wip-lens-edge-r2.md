# Lens: Edge cases — GATE 2 round 2

ДАЛІ: прочитав plan.md + spec.md; читаю transport.ts (ok:false), docker-compose.e2e.yml, Makefile e2e, nginx login zone pattern.

ФАКТ: transport.ts TransportOutcome ok:false → {code}; 429 та >=500 → dependency_unavailable (unnamedStatusCode). Адаптер за планом ковтає ok:false → self-report не відбувається (B6 закрито для transport-помилки).
ФАКТ: nginx api zone 10r/s burst20; login 10r/m burst5; e2e-варіанти лише для login+api (login-req.e2e.conf burst50, api-req.e2e.conf). Усі браузер-запити в Compose йдуть з однієї IP (docker network) → zone keyed $binary_remote_addr акумулює ВЕСЬ прогін.
ФАКТ: e2e target піднімає api/frontend/nginx; план додає clientlog-req.e2e.conf mount у S6 (ще не існує).
ФАКТ: план value-scrub лише message/error.stack/url-like; поле error.name не скрабиться (обмежене, не секрет). Вільного context-bag немає — набір полів закритий.

РІШЕННЯ (мої знахідки): див. Handback. blocking=3.
ДАЛІ: фінал.
