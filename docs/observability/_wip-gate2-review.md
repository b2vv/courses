# GATE 2 review — observability (B-1) plan

ДАЛІ: прочитав spec.md + plan.md + standards.md. Далі звіряю цитовані планом файли
(transport.ts, check-architecture.mjs, httpserver, observability, docker-compose,
infrastructure/monitoring, nginx login zone, jsonbody.ValidateUniqueKeys, server.ts),
тоді дві лінзи: plan-critique і plan-defense.

ФАКТ: spec AC1–AC7 прочитані. Plan має Reuse-first, Architecture, Side-effects, Migration,
Rollback, Constitution, Acceptance scenarios (AC→slice), Open items. Слайси S1–S6.

ФАКТ (звірка цитат плану з кодом):
- transport.ts, request-headers.ts, composition/, httpserver/, jsonbody/unique.go (ValidateUniqueKeys), observability/ — існують.
- observability/tracing.go має NewLogger(service) *slog.Logger → реюз "емітить через platform/observability" ТОЧНИЙ.
- httpserver/server.go має WriteJSON/WriteError/RequestID/ErrorResponse envelope — реюз точний.
- jsonbody.ValidateUniqueKeys існує; DisallowUnknownFields — stdlib json.Decoder (не jsonbody), ок.
- identity Redis limiter існує (identity/adapters/redis/rate_limiter.go); план ЯВНО його НЕ реюзає (ADR 013) — коректно.
- nginx: zones у nginx.conf (login 10r/m, api 10r/s...). Патерн rate-limit: location + include <zone>-req.conf (prod) + <zone>-req.e2e.conf (mount через docker-compose.e2e.yml). login має location = /api/v1/auth/login.
- docker-compose: prometheus під `profiles:[monitoring]`, без host-портів. `make up` = `docker compose up` БЕЗ --profile monitoring → Loki/Grafana/promtail НЕ стартують. `make check`...e2e теж явний список сервісів (api frontend wayforpay-stub nginx). AC7 ТРИМАЄТЬСЯ.
- frontend `frontend` service = SSR. workers: media-worker/asr-worker/upload.

ФАКТ (проблемні):
1. РЕЛІЗ-ID: grep RELEASE|VERSION|BUILD_|GIT_SHA|COMMIT по compose+frontend/src+config = ПОРОЖНЬО. План каже "release id from a build/version env" — але такої env НЕ існує, план не називає змінну/дефолт/місце ін'єкції. AC1 вимагає release context.
2. server.ts = 91 байт, ре-експорт @tanstack/react-start/server-entry. План: "in frontend/src/server.ts emit source=ssr on route-lane 500". Файл НЕ містить route-lane обробки; механізм перехоплення 500 (wrap handler / middleware / onError) не описаний. AC3 залежить. Також SSR — Node/TS, не може імпортувати Go observability.NewLogger; TS-логер для source=ssr не названо.
3. nginx clientlog zone: план згадує нову зону (S6), але НЕ згадує e2e-варіант (clientlog-req.e2e.conf + mount у docker-compose.e2e.yml), хоча це обов'язковий патерн репо. AC1 e2e POSTить у /client-logs; глобальний bridge ловить усі client-errors під час усього e2e — strict-зона (login-стиль 10r/m) може 429-ити. Ставка лишена "modest req/s" відкритою. Ризик для AC7 e2e-green.

ФАКТ (edge для AC2): редакція §17 — field-level (дропає поля з never-log set). Секрет у вільному тексті message/stack (вставлений токен) не скрабиться. AC2-тест перевіряє лише дропнуте поле, не підрядок. Механізм field-only — ок як межа, якщо названо явно.

ДАЛІ: запускаю skill plan-critique, тоді plan-defense.

ДАЛІ: запущено 4/4 лінзи plan-critique (AI-slop, логічні діри, надлишкова складність, edge cases) у свіжих сабагентах. Чекаю всі 4. Потім plan-defense (5 питань).

ФАКТ: лінза #3 надлишкова-складність повернулась (a93140):
- NB: клієнтський батчинг передчасний (Side-effects р.56 + Open р.100) — лишити простий per-session ceiling, викинути "batches"; закрити open рішенням "один POST/подія v1".
- NB: порт EventSink — шар без 2-го клієнта/зовн.залежності (р.23-27, Arch р.39-41); ADR "ports require actual external dependency". Ін'єктувати observability.NewLogger напряму АБО назвати майбутній seam.
- Чисто: reuse-first, promtail-вибір, дашборд v1, clientlog domain.
Повернулось 1/4. Чекаю AI-slop, логічні діри, edge cases.

ФАКТ: лінза #1 AI-slop повернулась (aabe99): 2 NB — обидві ДУБЛЮЮТЬ over-complexity:
- EventSink порт під єдиний випадок (ports рядок, Arch р.39-41, Reuse р.16) — 2 лінзи бачать → вага зростає.
- Батчинг/ceiling "на майбутнє" (Side-effects останній пункт + Open п.2) — 2 лінзи бачать.
- Чисто: frontend port+boundary+bridge (прямо з AC1), clientlog domain, S1-S6, reuse-таблиця, out-of-scope.
Повернулось 2/4. Чекаю: логічні діри, edge cases.

ФАКТ: лінза #2 логічні діри (ab185e) — 4 BLOCKING + 2 NB:
- B1 S3 чіпляє SSR-500 у server.ts, де хука немає (р.32-33, AC3→S3). Назвати реальний seam (defaultErrorComponent / обгортка server-entry / SSR middleware) + крок "знайти точку 500" ПЕРЕД emit.
- B2 release id для AC1 зі стану, якого ніхто не ставить (р.6, р.22, AC1). grep порожній. Додати крок, що ВСТАНОВЛЮЄ release (Rsbuild define/compose arg) + dev-дефолт.
- B3 "Compose proof" AC4/5/6 недосяжний шляхом гейта (р.92-94, р.66, р.52-53): make up без --profile monitoring, e2e лише api/frontend/stub/nginx. Нема визначеного profile-on раннера/команди/асерту. Додати явний profile-on verification-крок (команда, LogQL, Grafana API, ps network-only), поза make check.
- B4 (НОВЕ, сильне) promtail "source label per container" НЕ розрізняє client від api (р.34, AC4): client-логи виходять зі stdout КОНТЕЙНЕРА api (ingest там) разом із source=api. Per-container мітка дасть обом container=api. Потрібен promtail pipeline-stage, що парсить JSON-поле source у Loki-мітку. Blocking.
- NB: AC4 мапінг S4,S6 пропускає S5 (Grafana) + worker має бути піднятий у proof. 
- NB: AC2 "payload does not carry it" — редакція лише в бекенд-домені (S1); frontend payload постить §17 як є, секрет перетинає мережу й дропається лише в лог-рядку. Узгодити формулювання AC2 або додати клієнт-side не-збір.
- Чисто: порядок S1→S2, rollback пер-слайс, відсутність БД-міграції, AC7-консистентність.
Повернулось 3/4. Чекаю edge cases (a9664b).

ФАКТ: лінза #4 edge cases (a9664b) — 4 BLOCKING + 4 NB:
- E1 §17 секрет у вільному тексті message/stack/url field-редакція НЕ ловить (§22-27, AC2). Додати value-scrubbing за патернами (password=, Bearer, cookie/JWT) + size-cap; тест на вкладений секрет. [перетин з моєю знахідкою]
- E2 bridge може зациклити на відмові /client-logs (429/5xx) (§49-56, §28-30). Явний рядок: adapter НІКОЛИ не репортить власні transport-невдачі; негативний тест "падіння POST не породжує 2-й POST".
- E3 дедуплікація між 3 каналами відсутня (boundary+onerror+unhandledrejection+StrictMode) (§28-30, AC1) — 2+ події на одну відмову, ламає AC1 "one event" і e2e "один POST". Крос-канальний дедуп / чіткий поділ + StrictMode.
- E4 batching vs "рівно один POST" + тригер флашу не визначено + sendBeacon на unload (§56, AC1 e2e р.89). Детермінований флаш, sendBeacon не fetch-transport, e2e await на POST.
- NB: порожні release/route → sentinel, не дропати; invalid body 204vs400 (→400 envelope); rollback у зворотному порядку (S2 до S1) або tolerate 404; Loki-down degradation (stdout fallback, poll readiness).

=== ЗВЕДЕННЯ plan-critique: 4 запущено / 4 повернулось (всі дали знахідки) ===
BLOCKING (8, деду-пнуто):
B1 SSR-500 seam у server.ts відсутній (логічні діри)
B2 release id джерело не встановлене (логічні діри)
B3 Compose-proof раннер AC4/5/6 не визначений поза гейтом (логічні діри)
B4 promtail per-container мітка не ділить client/api (логічні діри)
B5 §17 free-text scrubbing message/stack (edge cases + я)
B6 bridge self-loop на відмові транспорту (edge cases)
B7 дедуп 3 каналів захоплення → ламає AC1/one-POST (edge cases)
B8 batching vs one-POST + flush trigger + sendBeacon (edge cases)
NON-BLOCKING (дедуп): EventSink порт без 2-го клієнта [2 лінзи]; батчинг передчасний [2 лінзи, зводиться з B8 → "1 POST/подія v1"]; AC4 мапінг S5/worker; AC2 бекенд-only редакція формулювання; sentinel release/route; invalid-body 204vs400; rollback reverse; Loki-down degradation.
Відкритих blocking: 8 → критика ЧЕРВОНА.

ДАЛІ: plan-defense — 5 фіксованих питань з посиланням на рядок плану.

=== plan-defense (таблиця в робочому файлі — план НЕ редагую, мандат рев'ювера) ===
| # | Питання | Закрито рядком плану | Вердикт |
|---|---|---|---|
| 1 | Кривий інпут | Reuse-first(httpserver "returns the standard envelope") + New code clientlog adapters/http ("ValidateUniqueKeys, DisallowUnknownFields, body cap, validating") | ДІРКА(вузька): механізм валідації названо, але reject-результат (204 vs 400, що бачить клієнт) не визначено; взаємодіє з self-loop B6 |
| 2 | Відмова залежності | Side-effects ("promtail reads ... read-only mount; no app change") — лише інфра | ДІРКА: деградацію КЛІЄНТ→ingest (429/5xx) план не описує; transport.ts віддає ok:false, але план ніде не каже, що adapter це мовчки ковтає й не репортить (B6); Loki-down backfill/stdout-fallback не зафіксовано |
| 3 | Гонка й повторний виклик | Side-effects ("adapter batches/caps ... per-session ceiling") — лише обсяг | ДІРКА: дедупу немає; одна помилка в boundary+onerror/+StrictMode = 2+ події → ламає AC1 "one event"/e2e "один POST" (B7); тригер флашу батча не визначено (B8) |
| 4 | Порожні/граничні дані | New code ("caps sizes") + body cap | ДІРКА(вузька): size-cap є, але fallback для порожніх release/route (sentinel) не визначено — найраніша boot-помилка або дропається, або шле "" (NB) |
| 5 | Rollback | §Rollback (цілий блок: "each slice reverts independently ... no persisted state ... infra reverts by removing profile services+volumes ... no data migration to reverse") | ЗАКРИТО (застереження: критика E-NB показала, що "independently" хибне — реверт S1 при живому S2 → 404-шторм; потрібен зворотний порядок/толеранс 404) |

Дірок: 4 (Q1,Q2,Q3,Q4; Q2 і Q3 — тверді). Кожна відповідь мала вказати рядок плану → Q2/Q3 не закриті планом. plan-defense ЧЕРВОНИЙ.

=== ПІДСУМКОВИЙ ВЕРДИКТ GATE 2: ЧЕРВОНИЙ ===
Критика: 8 відкритих blocking. Defense: 4 незакриті питання (Q1-Q4). Обидва гейти червоні.
