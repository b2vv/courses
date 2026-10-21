# Lens: Надлишкова складність (GATE 2 round 2)

ДАЛІ: прочитав plan.md + spec.md; перевіряю 4 нові механізми (§17 scrub 2-рівневий, крос-канальний dedup, ssr-logger.ts, per-session ceiling) на over-engineering. Далі гляну transport.ts і image-authorize.sh за потреби.

## ФАКТ transport.ts
- transport.ts = request-scoped `createApiTransport`, вже класифікує failure (timeout/cancel/5xx/429), tracer span. Client-log adapter постить через нього — чистий reuse, не новий механізм. Round-1 зняв EventSink-порт і батчинг — у плані їх немає (§25-36 «no EventSink», «No batching in v1»). Не перефлагую.

## РІШЕННЯ по 4 нових механізмах
1. 2-рівневий §17 scrub — NON-BLOCKING: spec AC2 вимагає лише field-level («payload ... does not carry it»); план САМ розширив AC2 до in-text і збудував 4-патерновий value-scrub (incl. JWT-shaped). field+size-cap закриває binding AC; value-scrub на message/stack має реальну користь (free-text може нести секрет), але JWT-shaped heuristic — false-positive-prone і неповний. Тримати exact-патерни (Bearer/token=/password=), зняти JWT-heuristic.
2. крос-канальний dedup (signature+TTL+StrictMode-absorb) — BLOCKING: план уже РОЗДІЛЯЄ канали (boundary=render, onerror/rejection=решта). У prod-білді (e2e path) StrictMode-подвоєння й React dev-rethrow НЕ відбуваються; React19 caught→onCaughtError (не window.onerror). Отже prod-подвоєння немає → AC1 «exactly one» дає сам поділ каналів. TTL-кеш сигнатур = stateful механізм під dev-only проблему — той самий патерн, що round-1 зняв у EventSink/батчингу.
3. ssr-logger.ts — CLEAN: leaf-адаптер, що пише один JSON-рядок у process.stdout, не імпортує Go/app-модулі, тримає server.ts тонким seam і лежить поруч із client-adapter у adapters/observability. Мінімальний, не overbuilt.
4. per-session ceiling counter — NON-BLOCKING: nginx `clientlog` zone уже тротлить на транспорті, adapter swallow-ить (B6) → без loop. Клієнтський лічильник = другий storm-guard під роботу, яку вже робить nginx. Тривіальний, але для v1 зайвий як «configured ceiling».

blocking = 1 (dedup)
