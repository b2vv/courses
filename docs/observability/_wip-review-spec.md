# SPEC-axis review — observability (B-1)

ДАЛІ: прочитати spec/plan/report (done), потім diff по кожному AC: clientlog domain (redact/event),
http handler, frontend adapter + bridge (client.tsx), server.mjs/server-log.mjs, monitoring-stack.sh,
compose profiles, e2e spec. Звірити AC1-AC7 + 5 відхилень.

ФАКТ: AC1 — порт+адаптер+bridge є; e2e ганяє лише канал window'error'; onCaughtError/unhandledrejection wiring у client.tsx БЕЗ тесту (жоден тест не імпортує client.tsx). describeUnknown тести = форми значень, не wiring каналів.
ФАКТ: AC2 — server scrub авторитетний, domain+handler тести покривають field/nested/in-text/no-echo. Клієнтський scrub прибрано (відх.#3) → секрет у message транзитить дротом сирим, чистий лише лог-артефакт.
ФАКТ: AC3 — unit доводить source=ssr для http_request+render_failed; у живому стеку render_failed детерміновано не тригернуто (report обмеження #2). route-render→200 — чесне обмеження з механізмом.
ФАКТ: AC4 — worker-мітка = media-worker/asr-worker, не "worker" (обмеження #5). script асертить media-worker. AC5/AC6 — monitoring-stack.sh реальний prod-path (profile up, реальні Loki/Grafana HTTP). Не перезапускав (read-only).
ФАКТ: AC6/AC7 — prometheus(227)/loki(239)/promtail(247)/grafana(257) усі profiles:[monitoring]; жодного host-порту; лише mailpit+nginx публікують (127.0.0.1). default up їх не стартує.
ФАКТ: scope creep не виявлено; grafana login=admin/admin (не anonymous як план §Open) — не AC.
