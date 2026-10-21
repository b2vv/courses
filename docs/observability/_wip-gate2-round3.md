# GATE 2 — раунд 3 (незалежний рецензент)

ДАЛІ: читаю spec.md, plan.md, _wip-gate2-round2.md; далі звіряю N1–N6 проти реального коду (server.mjs, Dockerfile, rsbuild.config.ts, transport.ts, docker-compose.e2e.yml, monitoring, nginx).

## Крок 1 — прочитав spec.md, plan.md @1a14cb9, round2. HEAD звірено: 1a14cb9.

## Крок 2 — N1–N6 проти реального коду (6 заявлено / 6 перевірено)

ФАКТ server.mjs (63 рядки): УЖЕ обгортає `entry.fetch` у try/catch (ряд.36-60); пише JSON
`http_request`+status у STDOUT на успіх (ряд.42-50) і `render_failed` у STDERR→503 на throw
(ряд.52-59). Коментар ряд.12-14 дослівно: «Start 1.169.3 wraps route components in Suspense: route
render errors resolve as HTTP 200 with the error deferred to the client». Жодного source/release/route
зараз нема — S3 має їх додати.

ФАКТ Dockerfile: ряд.27 `COPY --from=build .../server.mjs ./` — копіює ДОСЛІВНО (не бандлить);
ряд.30 `CMD ["node","server.mjs"]`. Build-ARG лише PUBLIC_WAYFORPAY_PAY_HOST; APP_RELEASE ще нема
(планова додача).

ФАКТ rsbuild.config.ts: `source.define` реально вживається (ряд.14-19, трансформує
`process.env.PUBLIC_WAYFORPAY_PAY_HOST` для клієнта). server.mjs копіюється поза rsbuild → define
туди НЕ доходить. Підтверджує N3-механізм.

ФАКТ transport.ts: TransportOutcome {ok:true|false,code}; `unnamedStatusCode` 429|>=500→
`dependency_unavailable` (ряд.59-61). Swallow реалістичний.

ФАКТ docker-compose.e2e.yml: nginx монтує login-req.e2e.conf + api-req.e2e.conf поверх того ж
шляху (ряд.28-30, last-mount-wins). Clientlog-варіант ще НЕ змонтовано (планова додача S6).

ФАКТ nginx: зони визначені в nginx.conf через `limit_req_zone $binary_remote_addr` (ряд.21-31);
login burst e2e=50 проти prod=5; api e2e=200 проти prod=20. $binary_remote_addr спільний на всі
e2e-POST з однієї docker-IP — підтверджує ризик N6.

СТАТУС:
- N1 (SSR-wrap у ТОМУ файлі): ЗАКРИТО. plan ряд.56-62 — збагачує НАЯВНИЙ catch у server.mjs
  (source/release/route), без паралельного wrap, без src/server.ts. Звірено: server.mjs уже
  обгортає+логує. Механізм здійсненний.
- N2 (AC3 route-render→200 обмеження): ЗАКРИТО. plan ряд.63-67 + AC3-табл ряд.146 — AC3 скоупнуто
  до shell-throw/503; route-render→200 як задокументоване обмеження (власник — клієнтський канал).
  Несе МЕХАНІЗМ (Suspense, звірений у коментарі server.mjs). Не hand-picked throw.
- N3 (release-define не доходить до server.mjs): ЗАКРИТО. plan ряд.51-55 — клієнт через
  source.define(`__APP_RELEASE__`), SSR через `process.env.APP_RELEASE` у рантаймі. Звірено:
  Dockerfile копіює server.mjs дослівно, define туди не доходить. Механізм коректний.
- N4 (крос-канальний dedup ПРИБРАНО): ЗАКРИТО. plan ряд.42-48 — dedup ВИДАЛЕНО; поділ каналів
  (error boundary=render через React19 `onCaughtError`; window.onerror+unhandledrejection=решта) —
  ЄДИНА гарантія «1 збій→1 подія». Явно названо, чому dedup був ШКІДЛИВИЙ (non-Error reject без
  message/stack колапсує N відмов у 1). React19 exclusivity здорова: boundary-спіймана помилка не
  доходить до window.onerror; помилки в обробниках подій не ловляться boundary. Канали взаємно
  виключні в prod-білді. Фікс round1-B7 (несправний) справді прибрано, не повернуто.
- N5 (short TTL без величини): ЗАКРИТО — зникло разом із dedup. У плані TTL більше нема.
- N6 (AC1 вимагає 204 + e2e-зона розмірена): ЗАКРИТО. AC1-табл ряд.144 тепер вимагає «exactly one
  /client-logs POST that returns 204… awaiting the network response» — ковтнутий 429 тепер ВАЛИТЬ
  тест (не хибно-зелений). Open ряд.160-162 розмірює clientlog-req.e2e.conf під увесь сьют на одну
  docker-IP (дзеркалить login-e2e burst), точний burst → S6. Обидві половини присутні.

6/6 N-findings закрито механізмом, названим на рядку І звіреним проти коду, що існує зараз.

## Крок 3 — plan-critique наново (4 лінзи, фокус на НОВИХ blocking від round-2 рефакторингу)

AI-slop: 0 blocking. Маркери обґрунтовані, EventSink/batching/dedup дропнуто (анти-slop). NB
(перенесений): клієнтський pre-send scrub = 3-я копія never-log (drift TS↔Go).

Логічні діри: 0 blocking. НОВІ NB від рефакторингу:
- (NB) render_failed пише в STDERR (server.mjs ряд.53), success — у STDOUT (ряд.42). План
  ряд.99/147 каже «stdout». Docker json-file зводить обидва потоки в один лог-файл, що promtail і
  читає → працює; але якщо promtail скрейпить лише stdout-потік, найважливіша SSR-помилка (AC3)
  випаде. Механізм (json-file бере обидва) робить це робочим; monitoring-stack.sh LogQL-проба
  мусить перевірити саме render_failed (stderr) рядок. Не blocking — доказовий скрипт ловить.
- (NB) release-значення: клієнт бере через build-ARG (build-time SHA), SSR — process.env у
  рантаймі. План не фіксує, що обидва мусять нести ТЕ САМЕ значення; інакше client/SSR release-мітки
  розійдуться. Механічна деталь wiring, не blocking.

Надлишкова складність: 0 blocking (рефакторинг СПРОСТИВ — головний over-eng round2, dedup,
прибрано). NB (перенесені): 2-рівневий §17 scrub ширший за AC2 (field-level); per-session ceiling
дублює nginx+swallow.

Edge cases: 0 blocking. E1/E2/E3 round2 закриті (dedup прибрано → non-Error collapse зникла; TTL
зник; AC1 вимагає 204 + зона розмірена). НОВІ NB:
- (NB) `__APP_RELEASE__` без ambient-декларації валить typecheck (AC7) — план ряд.51-52 не згадує
  declare/.d.ts. AC7-гейт ловить; фікс механічний. NB.
- (NB) AC1 «exactly one» потребує bounded-wait на відсутність 2-го POST; план каже «awaiting the
  network response», але не fiксує вікно стабільності count. NB.

РІШЕННЯ критики: НОВИХ blocking 0. Рефакторинг round-2 не вніс нової крихкості; навпаки — зняв
dedup-over-eng. Усі залишкові знахідки — NB, жодна не форклоузить роботу й не дає тихо-червоного
гейта (AC7 ловить typecheck; proof-скрипт ловить stderr).

## Крок 4 — plan-defense наново (5 питань, кожна відповідь на рядок плану)

Q1 кривий інпут: ЗАКРИТО. plan ряд.33-36 — ValidateUniqueKeys + DisallowUnknownFields + body cap +
validating; invalid→400 standard envelope (no secrets), valid→204.
Q2 відмова залежності: ЗАКРИТО. Side-effects ряд.95-100 — 429/5xx→adapter swallows (no
retry/self-report/loop); Loki-down ряд.98-100 — stdout завжди, probe event-driven.
Q3 ГОНКА/ПОВТОР→1 подія: ЗАКРИТО ПОДІЛОМ КАНАЛІВ, НЕ поверненим dedup. plan ряд.42-48 — dedup
видалено; поділ каналів (boundary=render/onCaughtError, bridge=решта) + React19 exclusivity =
гарантія. non-Error collapse (round2-реопен) усунено саме видаленням dedup. Реопен round2 закрито
по суті, а не косметично.
Q4 порожні/граничні: ЗАКРИТО. plan ряд.31-32 — sentinel release→"unknown"/route→"unknown" +
size caps; найраніша boot-помилка не дропається.
Q5 rollback: ЗАКРИТО. §Rollback ряд.121-126 — порядок S2→S1; swallow робить 404 тихим дропом;
infra revert = profile+volumes; no data migration.

Дірок defense: 0. Усі 5 закриті named-рядком, що вказує на ЗДІЙСНЕННИЙ механізм. Q3 (round2-реопен)
тепер справді закритий.

## Крок 5 — ЗВЕДЕННЯ

Перевірено: 6/6 N-findings (усі заявлені). Усі ЗАКРИТО механізмом на рядку + звірено проти коду.
Нові blocking від критики: 0. Defense Q1–Q5: усі ЗАКРИТО (Q3 реопен round2 — тепер закрито).
Залишкові NB (не форклоузять, не дають тихо-червоного): render_failed stderr vs «stdout» формулювання
(proof-скрипт має перевірити stderr-рядок); client/SSR release мусять нести те саме значення;
ambient `__APP_RELEASE__` declare для typecheck (AC7 ловить); AC1 bounded-wait на 2-й POST;
3-я копія scrub (drift); 2-рівневий scrub ширший за AC2; per-session ceiling дублює nginx.

ФІНАЛЬНИЙ ВЕРДИКТ GATE 2 РАУНД 3: **ЗЕЛЕНИЙ**.
