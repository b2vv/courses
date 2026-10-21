# GATE 2 review — observability (B-1) — РАУНД 2

ДАЛІ: прочитав round1-review, поточний plan.md, spec.md. Далі звіряю технічні твердження
доопрацьованого плану проти коду (server.ts wrap, Rsbuild source.define, transport.ts,
promtail pipeline, nginx e2e-зона, image-authorize.sh), тоді по кожному з B1–B8:
ЗАКРИТО/ЧАСТКОВО/ВІДКРИТО. Тоді plan-critique (4 лінзи) і plan-defense (5 питань) наново.

ФАКТ (технічна звірка доопрацьованого плану проти коду):
- server.ts: bare re-export `export { default } from '@tanstack/react-start/server-entry'` + `import './app/zod-jitless'`. Default export = `createServerEntry({ fetch })` → обʼєкт `{ fetch: RequestHandler }` (dist/default-entry/esm/server.js). Обгортка `.fetch` з try/catch + `res.status>=500` ТЕХНІЧНО ЗДІЙСНЕННА. B1 seam реальний і названий точно.
- Rsbuild `source.define` УЖЕ вживається в rsbuild.config.ts (ряд.13-19, PUBLIC_WAYFORPAY_PAY_HOST через process.env). Плановий `__APP_RELEASE__` через source.define — реальний, доведений важіль. B2 механізм здоровий.
- transport.ts повертає `TransportOutcome` {ok:true|false, code}; 429/5xx→'dependency_unavailable'. Adapter «ковтає ok:false, не самозвітує» (B6) — реалістично.
- infrastructure/tests/image-authorize.sh існує (5854B, +x). «monitoring-stack.sh дзеркалить image-authorize.sh» (B3) — реальний патерн, поза make check. B3 механізм здоровий.
- nginx e2e-variant патерн ПІДТВЕРДЖЕНО: docker-compose.e2e.yml монтує login-req.e2e.conf/api-req.e2e.conf поверх prod-шляху (last-mount-wins). Плановий `clientlog-req.e2e.conf` дзеркалить реальний патерн. Закриває round1 FACT#3.
- Makefile e2e: явний список `api frontend wayforpay-stub nginx`, БЕЗ monitoring. AC7 тримається; e2e-bridge POSTить у /client-logs через api-контейнер.
- promtail JSON pipeline-stage (парс поля source у Loki-мітку) — стандартна фіча promtail, не кастом; B4 механізм реалістичний.

ФАКТ (статус 8 blocking раунду 1 проти поточного плану):
- B1 SSR-500 seam: ЗАКРИТО. plan ряд.50-56 [G2:B1] — визнає, що server.ts — bare re-export, називає реальний seam (обгортка fetch default export), ssr-logger.ts у stdout source:"ssr", не імпортує Go; S3 спершу локалізує 500/throw, тоді emit. Технічно звірено: здійсненне.
- B2 release id: ЗАКРИТО. plan ряд.47-49 [G2:B2] — `__APP_RELEASE__` через Rsbuild source.define з APP_RELEASE (git short SHA, compose build-arg), dev/gate default "dev", крок у S1/S2. source.define уже в репо.
- B3 Compose-proof поза гейтом: ЗАКРИТО. plan ряд.140-144 [G2:B3] — infrastructure/tests/monitoring-stack.sh (дзеркалить image-authorize.sh): profile-on, LogQL на кожен source, Grafana HTTP API, docker ps/network, поза make check, результат у report.md. Патерн реальний.
- B4 promtail source-мітка: ЗАКРИТО. plan ряд.57-60 [G2:B4] — pipeline-stage парсить JSON-поле `source` у Loki-мітку (не per-container), тож client/ssr/api/worker розділяються хоча client виходить через stdout api-контейнера. Стандартна promtail-фіча.
- B5 §17 value-scrub: ЗАКРИТО. plan ряд.26-30 [G2:B5] — двохрівнева редакція: (a) заборонити поле з never-log, (b) value-scrub message/stack/url за патернами (password=, Bearer, token=, cookie/JWT) ПЕРЕД emit; §Конституція ряд.121 тест на nested+in-text секрет.
- B6 adapter self-loop: ЗАКРИТО. plan ряд.38-40 [G2:B6] — adapter НІКОЛИ не репортить власну transport-невдачу, failed POST ковтається; Side-effects ряд.84-86 закріплює (no retry/self-report/loop). AC-таблиця не має явного негатив-тесту «падіння POST→0 повторних POST», але механізм названо чітко (див. NEW нижче).
- B7 крос-канальний дедуп/StrictMode: ЗАКРИТО (з застереженням складності→критика). plan ряд.40-43 [G2:B7] — канали розділені: error boundary = render-errors; onerror/unhandledrejection bridge = решта; обидва через dedup (signature=message+stack, short TTL) що також поглинає StrictMode double-invoke → 1 подія. AC1 e2e await на один POST.
- B8 batching vs one-POST: ЗАКРИТО. plan ряд.43-44 [G2:B8] — «No batching in v1»: один POST/подія, per-session ceiling лише як storm-guard. Знято невизначений flush-trigger/sendBeacon. Open ряд.149 узгоджено (ceiling, не batching).

РІШЕННЯ: усі 8 blocking раунду 1 — ЗАКРИТО механізмом, названим на конкретному рядку плану; технічні seam-и звірені проти коду і здійсненні.

ДАЛІ: plan-critique наново (4 лінзи) — чи зʼявились НОВІ blocking від доопрацювання (складність дедупу/двохрівневого scrub/wrap-крихкість). Тоді plan-defense (5 питань).

=== plan-defense РАУНД 2 (план НЕ редагую — мандат рев'ювера; таблиця тут) ===
| # | Питання | Закрито рядком плану | Вердикт |
|---|---|---|---|
| 1 | Кривий інпут | New code ряд.31-36: adapters/http `ValidateUniqueKeys`+`DisallowUnknownFields`+body cap+validating; **invalid→400 standard envelope (no secrets), valid→204** [G2:NB]; Side-effects ряд.82-83 | ЗАКРИТО (round1-дірка «204 vs 400» усунена явним рядком) |
| 2 | Відмова залежності | Side-effects ряд.84-86 [G2:defense Q2]: 429/5xx→adapter swallows (no retry/self-report/loop, B6), логи втрачаються, нічого не ламається; Loki-down ряд.87-89 (stdout завжди, probe event-driven) | ЗАКРИТО (round1 тверда дірка закрита named-рядком) |
| 3 | Гонка/повторний виклик | New code ряд.40-43 [G2:B7]: канали розділені (boundary=render, bridge=решта) + dedup (signature=message+stack, short TTL) поглинає StrictMode → one failure→one event; no batching, 1 POST/подія [G2:B8] | ЗАКРИТО рядком (якість dedup — предмет критики-лінзи; якщо edge-лінза дасть blocking на злитті різних помилок, Q3 частково реоткривається) |
| 4 | Порожні/граничні | New code ряд.26-30: size caps + **sentinel** release→"unknown"/route→"unknown" щоб найраніша boot-помилка не дропалась [G2:NB] | ЗАКРИТО (round1 вузька дірка sentinel усунена) |
| 5 | Rollback | §Rollback ряд.108-114 [G2:NB]: порядок S2→S1; swallow робить 404 тихим дропом не штормом; infra revert = profile+volumes; nginx zone з блоком; no data migration | ЗАКРИТО (round1 застереження reverse-order тепер явне) |

Дірок defense: 0 (усі 5 закриті named-рядком). plan-defense РАУНД 2: ЗЕЛЕНИЙ — за умови, що лінзи критики не відкриють НОВИЙ blocking, який реоткриває Q2/Q3.

ДАЛІ: чекаю 4 лінзи plan-critique (4 запущено). Зведу, назву blocking, тоді фінальний вердикт.

ФАКТ: лінза #1 AI-slop (aeae03) повернулась: 0 blocking, 2 NB.
- NB: клієнтський pre-send §17 scrub = 3-я копія redaction (drift TS↔Go, загроза поверх backend не названа) → один авторитетний never-log набір АБО назвати загрозу (секрет у nginx/proxy-логах ДО ingest).
- NB: StrictMode як драйвер dedup не вправляється гейтом (e2e на prod-білді, StrictMode dev-only); dedup сам виправданий prod boundary→onerror овелапом.
- Чисто: wrap обґрунтований, reuse справжній, monitoring-stack.sh дзеркалить реальний патерн, clientlog e2e-variant реальний, §17 two-level закриває B5, ssr-logger виправданий, promtail source-label реальний, EventSink+batching дропнуто (анти-slop).
Повернулось 1/4.

ФАКТ: лінза #3 надлишкова-складність (a0aa40) повернулась: 1 BLOCKING, 2 NB.
- BLOCKING: крос-канальний dedup (ряд.40-44) — механізм під dev-only проблему. Поділ каналів УЖЕ: boundary=render, onerror/unhandledrejection=решта. У prod-білді (e2e на prod) StrictMode НЕ подвоює; у React 19 boundary-спіймана помилка → onCaughtError, НЕ window.onerror → prod-подвоєння немає, AC1 гарантує сам поділ каналів. TTL-кеш сигнатур = stateful-надлишок заради неіснуючої в prod проблеми (той самий патерн, що round-1 зняв з EventSink/batching). Правка: поділ каналів = ЄДИНА гарантія «1 збій→1 подія»; dedup прибрати з v1 або dev-only; e2e доводить 1 POST через поділ, не через dedup.
- NB: 2-рівневий §17 scrub ширший за spec AC2 (AC2 вимагає лише field-level); JWT-regex false-positive-prone і неповний → лишити field+size-cap як гарантію, value-scrub звузити до exact-патернів, in-text = best-effort.
- NB: per-session ceiling дублює nginx-троттл+swallow → прибрати з v1 або 1 рядок-guard.
- Чисто: ssr-logger leaf мінімальний; EventSink/batching відсутні (round-1 закрито); reuse наявний.
Повернулось 2/4. Чекаю logic-gaps (a4c7aa), edge-cases (ab4794).

ФАКТ: лінза #4 edge-cases (ab4794) повернулась: 3 BLOCKING, 5 NB.
BLOCKING:
- E1 non-Error reject колапс: Promise.reject("A")/undefined/object без .message/.stack → signature=message+stack однакова → N різних відмов зливаються в ОДНУ → губиться N-1 (підриває AC1 у зворотний бік). [КОРОБОРАЦІЯ з over-complexity BLOCKING: dedup не лише зайвий — ще й шкідливий]. Правка: задати екстракцію message/stack для non-Error + стабільний компонент підпису (type/constructor/serialize); юніт.
- E2 TTL без величини: «short TTL» не має числа — AGENTS.md вимагає величину/механізм. AC1-детермінізм висить на невказаному TTL в обидва боки (замалий→StrictMode-double не поглинається→2 події; завеликий→реальні повтори зливаються). Гейт AC1 невідтворюваний. Правка: конкретний TTL(мс)+обґрунтування+юніт.
- E3 e2e clientlog-зона не розмірена + акумуляція всього прогону на одну docker-IP ($binary_remote_addr) → цільовий POST AC1 може отримати 429; AC1 перевіряє лише «1 network POST», а 429 теж рахується як 1 POST при 0 залогованих подій → AC1 ХИБНО-ЗЕЛЕНИЙ. Правка: (а) розмірити clientlog-req.e2e.conf під об'єм усього сьюту (burst як login-e2e); (б) AC1-assert має вимагати 204 цільового POST, не лише факт POST.
NB: boundary+onerror різні stack-представлення (component-stack vs JS-stack) не колапсують dedup (prod ексклюзивність каналів або нормалізувати signature); self-loop не закрито для throw у САМОМУ capture-pipeline (re-entrancy guard); boot-вікно до install бриджа (sentinel лише про метадані — звузити твердження або ставити onerror найраніше); 400-envelope — фіксований code не raw decoder-рядок (тест body-з-секретом→400 без echo); AC4 Loki-probe retry-until-indexed конкретного рядка per source з bounded timeout.
Чисто: free-text scrub закрито; backend handler stateless (паралельний POST ок); batching закрито.
Повернулось 3/4. Чекаю logic-gaps (a4c7aa).

ФАКТ: лінза #2 logic-gaps (a4c7aa) повернулась: 3 BLOCKING, 3 NB. ЗВІРЕНО проти коду (підтверджено):
- server.mjs ІСНУЄ; Dockerfile prod `CMD ["node","server.mjs"]`, копіює server.mjs ДОСЛІВНО + dist + node_modules; package.json start=`node server.mjs`.
- server.mjs імпортує `entry from './dist/server/index.js'` (= збілджений src/server.ts), УЖЕ обгортає entry.fetch у try/catch, УЖЕ пише JSON: `http_request`+status(успіх), `render_failed`→503(throw).
- Коментар у server.mjs: «Start 1.169.3 wraps route components in Suspense: route render errors resolve as HTTP 200 with the error deferred to the client».
BLOCKING (підтверджені):
- L1 SSR-seam у НЕ ТОМУ файлі: реальний wrap живе в server.mjs, не src/server.ts; плановий 2-й wrap у src/server.ts (→dist) дублює наявну емісію (2 рядки на 1 throw) й ігнорує наявну success-лінію. ПЕРЕКРИВАЄ моє закриття B1. → збагатити НАЯВНИЙ catch у server.mjs (source:"ssr",release,route), не додавати паралельний wrap.
- L2 AC3 «5xx route lane» — неіснуючий стан: route-render-помилки → HTTP 200 (Suspense), відкладені на клієнт (їх ловить S2), НЕ 5xx/throw. Плановий wrap (throw|5xx) структурно не бачить найтиповіший клас SSR-помилок; AC3 зелений на hand-picked throw = та сама «зелений гейт на вигідній фікстурі». ПЕРЕКРИВАЄ моє закриття AC3. → скоупити AC3 до shell-throw/503; route-render→200 як задокументоване обмеження.
- L3 release-define не доходить до server.mjs: source.define трансформує лише src/*→dist; server.mjs копіюється дослівно → `__APP_RELEASE__` там undefined/ReferenceError. ПЕРЕКРИВАЄ моє закриття B2 для SSR. → SSR читає release з process.env.APP_RELEASE у рантаймі, define лише для клієнта.
NB: `__APP_RELEASE__` без ambient-декларації валить typecheck (AC7) — додати declare у S1/S2; dedup signature stack-less rejection (= E1); B3 one-shot proof поза гейтом — назвати стелю явно.
Чисто: B1 shape {fetch} підтверджено; AC7 default-up без профілю; B6 swallow/rollback; nginx e2e-variant; StrictMode no-op на prod-гейті.
Повернулось 4/4.

=== ЗВЕДЕННЯ plan-critique РАУНД 2: 4 запущено / 4 повернулось (усі дали знахідки) ===
НОВІ BLOCKING (дедуп між лінзами):
N1 [logic L1] SSR-wrap у НЕ ТОМУ файлі — server.mjs уже обгортає+логує; src/server.ts-wrap дублює. ПЕРЕКРИВАЄ round1-B1.
N2 [logic L2] AC3 «5xx route lane» неіснуючий; route-помилки→200/клієнт, не source=ssr; зелений на hand-picked throw. ПЕРЕКРИВАЄ round1-AC3.
N3 [logic L3] release source.define не доходить до дослівно-копійованого server.mjs → SSR release undefined; треба runtime-env. ПЕРЕКРИВАЄ round1-B2(SSR).
N4 [over-complexity BLOCKING + edge E1 + logic-NB — 3 лінзи] крос-канальний dedup: (a) ЗАЙВИЙ у prod (поділ каналів+React19 onCaughtError гарантують 1 подію), (b) ШКІДЛИВИЙ: non-Error reject ("x"/undefined/object) без message+stack → однакова signature → N різних відмов зливаються в 1 → губиться N-1 → ламає AC1 у зворотний бік. Сам фікс round1-B7 несправний.
N5 [edge E2] dedup «short TTL» без величини (порушує AGENTS.md: обмеження несе величину/механізм) → AC1-детермінізм невідтворюваний. (Знімається, якщо dedup прибрати за N4.)
N6 [edge E3] e2e clientlog-зона не розмірена + акумуляція всього прогону на одну docker-IP → цільовий AC1-POST може 429; AC1 перевіряє лише «1 POST», а ковтнутий 429 рахується як 1 POST при 0 подіях → AC1 ХИБНО-ЗЕЛЕНИЙ. Діра в доказі AC1/AC7.
NON-BLOCKING (дедуп): клієнт pre-send scrub = 3-я копія redaction (drift); StrictMode як мотивація dedup (dev-only); 2-рівневий scrub ширший за AC2 + JWT-regex спекулятивний; per-session ceiling дублює nginx; throw-у-capture-pipeline self-loop (re-entrancy guard); boot-вікно до install бриджа; 400-envelope фіксований code не raw decoder; AC4 Loki-probe retry-until-indexed; typecheck ambient __APP_RELEASE__; B3 one-shot-ceiling назвати явно.
Відкритих НОВИХ blocking: 6 (N1–N6; N5 злитий з N4 як dedup) → критика РАУНД 2 ЧЕРВОНА.

=== СТАТУС 8 blocking раунду 1 (переглянуто після server.mjs) ===
- B1 SSR-seam: РЕОПЕН (N1) — seam названо в НЕ тому файлі.
- B2 release id: ЧАСТКОВО/РЕОПЕН (N3) — ок для клієнта, НЕ для SSR.
- B3 Compose-proof: ЗАКРИТО (NB: назвати one-shot-стелю).
- B4 promtail source-label: ЗАКРИТО.
- B5 §17 value-scrub: ЗАКРИТО (NB: звузити scope до AC2).
- B6 self-loop: ЗАКРИТО для transport (NB: throw-у-pipeline).
- B7 dedup: РЕОПЕН — сам фікс несправний+зайвий (N4/N5).
- B8 batching: ЗАКРИТО.
→ Чисто закрито 4/8 (B3,B4,B5,B8); реопен 3 (B1,B2,B7); B6 закрито з NB.

=== DEFENSE переглянуто ===
Q3 (гонка/повтор→1 подія): РЕОПЕН — рядок плану (dedup) існує, але механізм, на який він вказує, несправний (N4 non-Error collapse ламає «1 подія»). Q1,Q2,Q4,Q5 лишаються закриті named-рядком. Defense: 1 реопен (Q3) → ЧЕРВОНИЙ.

=== ПІДСУМКОВИЙ ВЕРДИКТ GATE 2 РАУНД 2: ЧЕРВОНИЙ ===
Критика: 6 нових blocking (3 перекривають round1-B1/B2/AC3 через server.mjs; dedup-фікс несправний; e2e AC1 хибно-зелений). Defense: Q3 реопен. Корінь: план планувався проти src/server.ts, а реальний SSR-процес — server.mjs (уже логує) → B1/B2/AC3 не закриті; B7-dedup — несправний over-eng.
