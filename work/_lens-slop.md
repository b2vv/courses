
## Лінза AI-slop — сесія 2026-09-07
ДАЛІ: читаю docs/playback/plan.md (предмет), потім spec.md, _wip-phase6.md, AGENTS.md, .claude/standards.md, work/CTO-RESEARCH.md. Шукаю: абстракції під один виклик, церемонію без користувача, перекази спеки, «на майбутнє», непроверювані приймальні рядки, цифри без походження.

ФАКТ: plan.md прочитано (324 рядки). Структура: 1 Розміщення, 2 Reuse-first, 3 Рішення D-1..D-7, 4 Сайд-ефекти, 5 Міграція, 6 Rollback, 7 Конституція, 8 Приймальні (P01-P13, B01-B19, R01-R23), 9 Задачі (порожньо, "після GATE 2").
ДАЛІ: читаю spec.md (щоб знати, що є переказом спеки), _wip-phase6.md (G1-G14), standards.md.

ФАКТ: посилання плану на код перевірено — точні: handler.go:123 Authenticate (RequireSameOrigin насправді :143, план каже :142 — off-by-one, ігнорую), transcoder.go:150-152, nginx.conf:45-46, docker-compose.yml:26/39/141, openapi.yaml:619-624, docs/media/report.md B07/B08 (справді міграційні канарки). Стеля 40×120 підтверджена openapi.yaml maxItems 40/120.
ФАКТ: D-2 «пастка». Голий образ nginxinc/nginx-unprivileged:1.30.4-alpine → /var/cache/nginx drwxrwxr-x nginx:root = WRITABLE. Під compose → NOTWRITABLE, бо docker-compose.yml має `read_only: true` + `tmpfs: [/tmp, /var/cache/nginx]`, і tmpfs монтується root:root 0775. Тобто ВИСНОВОК правильний, але названа причина («скомпільовані шляхи не записувані для uid 101») хибна, а наведений «транскрипт» не міг дати показаний вивід (`|| echo no` друкує `no`, а не «not writable ...»; `id` реально має `groups=`).
ФАКТ: Makefile:8-12 `up` = `docker compose up -d --wait` + `nginx -s reload`. Отже «контейнер здоровий» і «nginx -t exit 0» у B10 вже гарантовані гейтом up.
ФАКТ: videoPollMs=3000 (admin-video-queries.ts:12), поллінг умовний (лише queued/processing, адмінський екран). Порівняння §4 «того ж порядку, що поллінг Phase 5» не міряне: 3 с проти 30 с і зовсім інша популяція акторів.
ФАКТ: spec.md §4 виносить «Обсервабельність фронту» в беклог B-1, а §2 плану заявляє переюз backend platform/observability для «метрики проби AV1». Шлях браузер→метрика не названо ніде.
РІШЕННЯ: 11 знахідок, 2 blocking (§8 «ручних рядків 0» проти B19/P12; D-7/P12 метрика — переказ спеки без ЯК).
