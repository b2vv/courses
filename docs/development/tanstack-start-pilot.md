# TSSTART-01 — TanStack Start upgrade pilot

## Checkpoint v3 — defense revision — BLOCKED до product mutation

Дата: 2026-10-18. Owner: користувач; research/plan: delegated Codex. Primary record — цей файл; task TSSTART-01 у work/tasks.md, journal нижче. Scope доручення: цільове оновлення Start та необхідні сумісні frontend правки; без commit/push/deploy, зміни стеку/auth/permissions/hooks/graph DB, paid API чи billing fallback. Поточний rollout AI Factory 0.7 збережено byte-for-byte.

Baseline: main / 4c98a7f120bbb8fee52ab3c5e81e42dc8c9f9866; один worktree. npm package-lock v3; Node v24.11.1, npm 11.6.2. Manifest/lock/installed збігаються: Start 1.168.49, @tanstack/react-router 1.170.32, SSR Query 1.167.2, React/DOM 19.2.7, Rsbuild 2.2.2. Index чистий; tracked dirty — попередні 5 rollout-файлів, 201 untracked artifacts збережено. Повний public fingerprint і команди в `~/project-data/courses-tanstack-start-pilot/`.

Окремого активного TanStack upgrade assignment у перевірених work/docs не знайдено. Це не підтверджує відсутності незаписаного writer: інші Courses sessions/processes присутні. Перед install/build потрібні погоджений writer/ізольований checkout та повторний HEAD/dirty fingerprint; чужий devserver не використовувати, порти 3000/8080 не займати за замовчуванням. Branch не перемикається цим research.

## Стабільна ціль та залежності

Router тут означає **@tanstack/react-router**, не react-router іншої екосистеми. Версії — попередній official research snapshot, не повторний latest lookup після reconnect.

**Scope v2: exact Start 1.169.3 + @tanstack/react-router 1.170.45 + @tanstack/react-router-ssr-query 1.167.3 та необхідні транзитивні залежності. B1 змінює default на lifecycle-aligned SSR Query patch; Courses2 прийняв v2 на plan stage, 0 blocking; v3 додає defense та потребує delta binding.**

За official npm latest на дату дослідження:

| Пакет | Зараз | Пропозиція | Причина |
|---|---|---|---|
| @tanstack/react-start | 1.168.49 | **1.169.3** | Latest stable, GitHub prerelease=false; перехід через 1.169 stable |
| @tanstack/react-router | 1.170.32 | **1.170.45** | Exact dependency нового Start; уникнути дубльованого runtime |
| @tanstack/react-router-ssr-query | 1.167.2 | **1.167.3 EXACT** | Peer @tanstack/react-router >=1.170.33; router-ssr-query-core 1.169.3 замість 1.169.2. Release notes підтверджують узгодження request cleanup/stream ownership; це аргумент сумісності, не доведена несумісність старої версії |

Зберегти React/DOM 19.2.7, React Query 5.102.8, Rsbuild 2.2.2 та його plugins, Rstest, TypeScript і решту deps. Start вимагає Node >=22.12 і React >=18/19; Rsbuild peer ^2 підтримує поточний 2.2.2. Vite peer optional; Vite 8 правило стосується Vite builds, не привід мігрувати з Rsbuild. SSR Query 1.167.3 peers Router >=1.170.33, Query >=5.102.0 сумісні з пропозицією. Lockfile graph ще не вирішувався; semver metadata не є install/build proof.

Офіційні джерела:
- [npm Start](https://registry.npmjs.org/@tanstack%2freact-start/latest)
- [Start 1.169.3 release](https://github.com/TanStack/router/releases/tag/%40tanstack%2Freact-start%401.169.3)
- [Stable migration](https://tanstack.com/start/latest/docs/framework/react/migrate-from-release-candidate)
- [Rsbuild integration](https://tanstack.com/start/latest/docs/framework/react/build-from-scratch)
- [SSR Query patch](https://github.com/TanStack/router/releases/tag/%40tanstack%2Freact-router-ssr-query%401.167.3)

## Acceptance / non-goals

A1 Exact approved three-package set and npm lock resolution; official registry only. Зберегти окремий `npm ls --all --json <package>` з exit/artifact для кожного: @tanstack/react-start, @tanstack/react-router, @tanstack/react-router-ssr-query, @tanstack/router-core, @tanstack/router-ssr-query-core, @tanstack/start-client-core, @tanstack/start-server-core та @tanstack/query-core. Lock-tree звірити з runtime: один очікуваний resolved version кожного пакета, dedup paths пояснені, жодних invalid peers/неочікуваних копій. Різні назви core-пакетів не мусять мати однаковий version. Actual duplicates до install/lock resolution НЕ встановлені.
A2 Existing Rsbuild dev/prod and handwritten srvx wrapper server.mjs (imports generated dist/server/index.js) work; no Vite/new services/RSC adoption or backend/schema migration.
A3 Existing TypeScript/unit/component/route tests preserve assertions; request-scoped Query/identity, cancellation, redirects/search/guards remain valid.
A4 Routing/SSR/hydration/auth/session/returned-response evidence: reuse existing route-guards, accounts, shell and recovery E2E for their covered acceptance; full E2E required by local shared-state contract. Додаткові вузькі observables O1–O5 нижче закривають реальні прогалини, не повторюють всю E2E матрицю вручну. No form-only substitute for result/view.
A5 Storybook shared Rsbuild config still builds with Start disabled; render smoke only for impacted stories. Existing addons=[] gives no automated Storybook a11y/visual PASS. Existing component axe evidence remains distinct.
A6 Every command records exit/artifact and current snapshot; required unavailable check BLOCKED/NOT_RUN. No lowered assertions, exclusions or gates. Independent plan + implementation verdict bind exact snapshots. No commit/push/deploy.

## Grilling / assumptions / impact

User selected the upgrade and preservation of stack; no unresolved preference on framework replacement is assumed. Open decisions are operational: manual delivery/delta binding by identified Courses2; writer ownership; isolated runtime/ports and permitted heavy-gate window. A new auth/session library or irreversible migration requires a new user decision, not implicit acceptance.

Framework bump зачіпає shared SSR/auth: v2 independent critique ACCEPTED; авторська defense v3 нижче, операційний допуск ще BLOCKED; цей record лишається єдиним планом.

Graph fallback: no verified fresh scoped graph supplied and graph refresh forbidden. Direct file evidence:
- frontend/rsbuild.config.ts:4,8–11: Start Rsbuild plugin and STORYBOOK bypass.
- frontend/rstest.config.ts:3–15: alias to start-server-core/request-response; transitive export compatibility needs recheck.
- frontend/server.mjs:4,28–57: handwritten srvx wrapper around generated dist/server/index.js, nonce context, CSP and response-header mutation plus catch fallback.
- frontend/src/router.tsx:17–33,48–61: getGlobalStartContext nonce fallback, request-scoped QueryClient, SSR Query integration.
- frontend/src/app/composition/identity.ts:2–14 and catalog/admin/learn/teach/enrollment/locale: getRequestHeader/getCookie APIs; missing versus empty header behavior matters.
- frontend/src/routes/__root.tsx and route-guards.test.tsx; app/composition/identity-queries.test.tsx, catalog-queries.test.tsx: cache/guard/navigation coverage.
- frontend/e2e/route-guards.spec.ts:38–102,207–225; accessibility.spec.ts, recovery.spec.ts and shell.spec.ts: existing integration/browser routes.

Read-only search frontend/src + server.mjs found no inputValidator, useSession, custom createStartHandler/createServerEntry or response-helper setter use. This is scoped absence, not proof all transitive behavior is unchanged. Go owns sessions, so removed Start useSession does not by itself justify cookie replacement. Stable changes to error conversion/returned response, headers/cookies and SSR stream cancellation require real validation. No server-function RawStream migration inferred without usage evidence.

## Proposed implementation / migration / rollback

1. Передати defense delta v3 тому самому Courses2 вручну через Віталія для оцінки delta та прив’язки збереженого ACCEPTED v2; answer raw findings. Confirm writer and isolated runtime plan. Rebind if sources changed.
2. In approved isolated checkout or exclusive scope, backup exact package/lock; use npm with explicit versions and registry.npmjs.org, no blanket update, no third-party installer. Inspect changed transitive packages, peers and install lifecycle scripts before execution; do not silently enable new scripts.
3. Only frontend/package.json, package-lock.json initially; minimal compatibility fixes to existing build/router/server/tests only if observed failures require them. Do not edit tests to hide changed behavior. Re-review material scope expansion before implementation.
4. Run below impact-based checks; preserve all failure evidence. Compare generated route/build outputs to baseline and retain only explained task changes.
5. Fresh existing-reviewer handoff; no author verdict substitution. Stop before commit/push.

Rollback is reversible files/dependency set, no DB migration: restore only exact task-owned package/lock/compatibility hunks after before/after hash check; refuse on foreign edits. Recreate dependencies from original lock only in the isolated environment; never reset shared checkout/node_modules or kill someone else's server. Перед restore перевірити ВСІ task-owned файли, не починати пофайловий rollback до проходження всіх guards. У разі foreign edit: STOP, записати expected/current hashes і partial-apply inventory, зберегти обидва стани та запросити узгоджений merge/restore; не заявляти rollback і не продовжувати роботу в partial state. Success: package/lock/owned files дорівнюють backups (SHA-256 lockfile EXACT), а per-package npm ls в ізольованому середовищі показує початкові versions без invalid peers. Partial install/build до невдачі лишається явно BLOCKED, не видається за відновлене дерево.

## Planned checks — all NOT_RUN

| Criterion | Existing check | Scope / side effects / route |
|---|---|---|
| A1 | Per-package npm ls --all --json + exact lock graph у A1 | No install yet; package-specific dedup/peers evidence, no automatic npm dedupe |
| A2/A3 | frontend npm run format:check; npm run lint; npm run typecheck; npm test; npm run build | After critique/writer gate; build writes generated routes/dist/cache, use isolation. Full frontend tests due shared router |
| A3 | route-guards, identity-queries, catalog-queries component/unit cases in existing suite | No duplicate test harness; add tests only for demonstrated gaps |
| A5 | frontend npm run build-storybook | Relevant because shared rsbuild config imports Start; no new addon/test runner |
| A2/A4 | Existing E2E + narrow O1–O5; isolated dev/prod startup | Explicit free private ports and coordinated environment; no current service touched |
| A4 | existing ordered/full E2E per AGENTS; make e2e/e2e-local/check only in coordinated isolated stack | Compose up/build, resets/seeds, setup/teardown mutate shared resources. BLOCKED pending separate coordination; never run as research |
| Backend gates | Standalone Go unit/integration NOT_RUN: Go packages outside this frontend change dependency graph | Не запускати окремий backend/reset pipeline для research. Якщо чинний обов’язковий ordered gate включає backend stages, він лишається BLOCKED до coordinated isolated run; це не waiver локальних gates |

## Reviewer route / admission

Courses2: independent Courses Claude, reference supplied by Віталій — `claude.ai/code/session_01WTzvny2ZqFKZNhJtoyAVb1`. Отриманий critique v1: **BLOCKED**, B1–B3, 12 non-blocking; raw response збережено без змін. Автор не перевіряв фактичний запуск чотирьох лінз, лише отриманий report та його exact binding.

Courses2 re-critique v2 **ACCEPTED (plan stage), 0 blocking** — артефакт `~/project-data/courses-tanstack-start-pilot/revision-2/critique-response-courses2-v2.md` звірено з plan SHA `150960144633e4c00928fa5be406d530c0e18fabc591e2323df6f720adbea2db`, snapshot `6887deab70c32b00389c1d62d1fd0a2f281a554b708f96cddc99b896f70c167a` і HEAD. B1–B3 закрито рецензентом; N1–N12 addressed, обидва factual rebuttals прийняті. Raw findings/dispositions v1/v2 лишаються історичними; це не impl acceptance.

V3 додає авторську defense. Pinned plan-defense skill не вимагає окремої незалежної defense-сесії; авторська структурна перевірка не є незалежним PASS. За ai-factory.md:90 той самий Courses2 має оцінити delta та явно прив’язати збережений verdict; якщо вважає clarification material — повторити критику лише зачепленого scope. Не повторювати весь фан-аут автоматично.

Delivery — **вручну через Віталія**, NOT_CONFIRMED для v3. `01a13bd8-49f1-7279-806c-4c38b62cf056` — voice child, НЕ local Codex author/rollout ID; `codex exec resume` до нього не використовувати. “No rollout found” попри exit 0 не доводить delivery; local rollout ID не вигадується.

**Pre-code admission BLOCKED:** delta binding v3 та підтвердження exclusive writer/isolated checkout/runtime ownership і плану heavy-gate window. Відповідь власника щодо coordination ще очікується. **Post-implementation acceptance:** A1, A2/A3, O1–O5, Storybook build, applicable ordered/full E2E і independent implementation review — NOT_RUN/PENDING; вони виконуються ПІСЛЯ дозволеної реалізації, а не є неможливою вимогою мати runtime PASS до зміни коду. Не називати upgrade завершеним до цих доказів.

## Observable checks O1–O5 — proposed, NOT_RUN

- **O1 / B2, N9 — nonce та enforced CSP.** На погодженому isolated production server unset `CSP_REPORT_ONLY`; перевірити наявність enforcing `Content-Security-Policy` (не лише Report-Only). Для двох concurrent SSR GET: витягти nonce з script-src кожної response і звірити з nonce кожного executable SSR script у тому ж HTML; nonce непорожній і різний між responses. Не вимагати executable nonce від inert JSON/data scripts без причини. Browser smoke має реально hydrate/interact, без blocked application scripts/CSP violations; не інжектувати axe/addStyleTag, які самі можуть порушити policy. Report-only E2E не зараховується за цей proof.
- **O2 / B3 — anonymous session fetch.** Реальний anonymous SSR GET БЕЗ Cookie: на isolated API boundary порахувати `/api/v1/auth/me` requests, scoped саме до цього SSR request, до hydration: **0**; `shouldLoadCurrentUser=false`, identity anonymous. Missing cookie, empty Cookie header і present cookie з empty value — різні випадки, не змішувати; present-empty не оголошувати missing. Доповнювальний helper test має виконати actual target getCookie у request context, не mocked return. **НЕ** вимагати відсутності dehydrated session key: root-identity.ts:21–23 навмисно setQueryData(['session'], null), що не доводить fetch. Після hydration чинний client revalidation contract окремий.
- **O3 / N7–N8 — response vs throw.** Reuse recovery E2E для фактичних returned-error paths; зафіксувати returned status, `Cache-Control: private, no-store`, `Vary` з Cookie, CSP у handwritten wrapper для пройдених ним responses. В окремому ізольованому test boundary змусити `entry.fetch` throw: очікувати wrapper catch — 503, тіло `Rendering unavailable`, `Cache-Control: no-store`, `render_failed`; не підміняти це returned 500/503. Start може конвертувати render error у Response, тому сам “force render error” не гарантує catch. Pre-entry malformed-path/error може мати інший маршрут: спочатку довести, що досягнув wrapper; не переписувати production error semantics для зручності тесту.
- **O4 / N10 — cancellation/deadline.** Reuse transport.test.ts (SSR abort vs deadline) + catalog-queries cancellation; вони не доводять whole Start runtime. На isolated delayed upstream SSR abort mid-loader: requestSignal передає abort у fetch, downstream робота завершується без hang/leak. Для identity з реальною SSR конфігурацією timeoutMs=1000 простежити deadline/request_timeout у допустимому timing margin, не exact wall-clock equality. No shared worker/service restart.
- **O5 / N12 — URL search/returnTo.** Reuse login-search.test.ts rejected //evil.example, external URL/backslash cases і catalog-search roundtrip. Звірити composed router parse/stringify для normalized URL state зі спецсимволами та percent-encoding і кінцевий redirect лишається same-origin; рядок //evil.com не стає зовнішнім origin. Не створювати другу повну матрицю там, де existing assertions це вже доводять.

## Disposition v2 — авторські правки, не verdict Courses2

| Finding | Рішення / доказ |
|---|---|
| B1 | Default SSR Query 1.167.3; cached official metadata: peer Router >=1.170.33, ssr-query-core 1.169.3 (old 1.169.2); release notes lifecycle alignment. Твердження про вже наявні duplicates не прийнято: Router тягне router-core 1.171.38, це інший пакет; A1 перевірить actual graph/dedup. |
| B2 | O1: response nonce equality + enforcing browser proof; docker-compose.e2e.yml:21 report-only недостатньо. |
| B3 | O2 додає прямий zero-session-fetch observable. Target start-server-core 1.169.44 internal-request-response.ts:1413–1435 повертає cookie map lookup; missing→empty change не підтверджено. Stable getRequestHeader empty semantics — інший API. Negative dehydrated-key assertion відхилено через root-identity set null. |
| N1 | Existing E2E для covered matrix; нові runtime checks тільки O1–O5 gaps, не дубль all-flows manual pass. |
| N2 | Standalone Go checks NOT_RUN за impact, локальний ordered gate не скасовано. |
| N3 | HANDOFF v2 скорочено до binding/scope/route та link на цей план. |
| N4 | Єдиний exact three-package default, optional відкладання прибрано. |
| N5 | Risk classification скорочено до дії: незалежна критика shared SSR/auth. |
| N6 | Самооцінку lifecycle з journal прибрано; лише факти й NOT_RUN. |
| N7 | server.mjs названо handwritten wrapper; O3 returned-response headers/statuses. |
| N8 | O3 окремо returned response і forced entry throw; render error не ототожнюється з catch. |
| N9 | O1 concurrent nonces: own header/script match і міжзапитна відмінність. |
| N10 | O4 reuse existing tests + runtime propagation/deadline gap. |
| N11 | All-files preflight rollback guards, original hashes/versions, STOP на foreign edit без partial success. |
| N12 | O5 reuse malicious returnTo/roundtrip tests; перевірити composed boundary без дублювання. |

## Journal

2026-10-18 reconnect: verified dependency hashes unchanged, recovered and saved checkpoint without new network/install/runtime. Existing reviewer pointer still has null reference and no response; bounded app thread listing did not identify a current Courses Claude reviewer. No session contacted.

2026-10-18: inspected versions/lock/installed packages, worktree and dirty scope; official npm/GitHub/docs research; identified stable migration risks and compatible target. Prepared plan and registered blocked task. No dependency files, product code, node_modules, devserver, tests, graphs, hooks, auth or paid APIs changed/executed. Upgrade correctness and runtime behavior remain NOT_TESTED.

2026-10-18 v2: отримано critique Courses2 через Віталія; v1 hash/snapshot/HEAD і всі pinned source hashes збігаються. Перевірено exact npm source tarball server-core 1.169.44 (SHA-512 integrity), лише прочитано source, код не виконувався. Внесено B1–B3/N1–N12 dispositions лише до цього plan-doc; install/build/product checks NOT_RUN. Re-critique та defense PENDING.


## Plan-defense v3 — авторські відповіді, не незалежний PASS

Pinned skill: `<release>/skills/plan/plan-defense/SKILL.md` з exact release c427bd807f250e8d4b7d9c42d9bc894d8f8dce71. Рішення target/стек/A1/O1–O5 та B/N dispositions v2 збережено. Нижче уточнено поведінку на чинних boundaries, без нових features або послаблення gates.

**D1 — кривий інпут.** O5 і чинний `frontend/src/app/router/catalog-search.test.ts:34–65` задають malformed/out-of-range/repeated URL params: reject як not-found замість server error; external/protocol-relative returnTo відкидається, користувач не переходить на інший origin. Missing/empty cookie cases розрізняються O2; не перетворювати anonymous на “signed in”. Invalid percent-encoding перевірити за target Start contract (400); pre-entry path не прирівнювати до handwritten wrapper catch (O3). Зміна accepted product schema чи нова auth policy поза scope цього dependency upgrade, бо їхня поведінка не замовлена; чинні validations/assertions лишаються.

**D2 — відмова залежності.** O3 розрізняє returned error response і entry throw: користувач бачить чинний error state або `Rendering unavailable`/503 для доведеного wrapper catch, не false success/empty data. O4 обмежує identity SSR deadline 1000 ms із timing margin і відрізняє abort від timeout; existing transport/recovery tests повторно використовуються. Для npm registry/peer/install failure — stop, зберегти raw error і partial inventory, не міняти target/registry/пороги щоб отримати green; rollback за D5. Відсутність середовища для O1/O2 означає BLOCKED цього proof, не PASS.

**D3 — гонка й повторний виклик.** O1: два SSR requests мають різні nonces і кожен script/header збігається у власній response; request-scoped Query/identity не переноситься між користувачами. O4 та existing catalog-queries cancellation: за navigation race попередній loader скасовується, stale result не підміняє актуальний маршрут. Повторний запуск upgrade спочатку порівнює lock/package/current hashes: already-at-target — inspect/verify, не другий update або seed; foreign change — stop. Business double-submit/idempotency не перепроєктовується: доменні mutation semantics поза scope без змін contracts, але наявні E2E assertions не видаляються. Не заявляти “один результат” для двох незалежних SSR GET — інваріант тут ізоляція, не дедуп двох різних запитів.

**D4 — порожні та граничні дані.** O2: absent cookie → anonymous, 0 SSR session-fetch, session cache null дозволений. Existing catalog-search tests: plain URL дає q='', sort=recent, limit=20, offset=0; invalid/duplicate params відхиляються, special-character roundtrip збережений (O5). Порожній список/один елемент і доступні boundary fixtures рендеряться через наявні component/E2E assertions без залежності від попередньої SSR request; degraded response не маскується empty success (D2). Нові максимуми доменних колекцій чи глибини lesson tree поза scope: upgrade не змінює data model/limits; чинні unit/full-E2E coverage збережені, непокритий релевантний regression фіксується gap, не “перевірено”.

**D5 — rollback.** Розділ Proposed implementation / migration / rollback задає all-files preflight, exact backups і refusal on foreign edits; success — SHA-256 package-lock == backup та per-package npm ls original versions в ізольованому середовищі. Немає DB/schema migration, deploy або нових зовнішніх business writes у scope; їх dependency rollback не “відкочує”, і запускати їх цим планом не дозволено. Partial npm/build failures зупиняють роботу з exact hashes/inventory, без hard reset/shared node_modules/server termination; відновлення тільки owned changes після узгодженого guard. Якщо foreign conflict блокує restore — оператор узгоджує merge/відновлення, статус BLOCKED, не удаваний rollback PASS.

| # | Питання | Закрито рядком плану | Авторський вердикт |
|---|---|---|---|
| 1 | Кривий інпут | §Plan-defense D1; O2/O3/O5 | Відповідь записана |
| 2 | Відмова залежності | §Plan-defense D2; O3/O4 | Відповідь записана |
| 3 | Гонка й повторний виклик | §Plan-defense D3; O1/O4 | Відповідь записана; domain redesign явно поза scope |
| 4 | Порожні та граничні дані | §Plan-defense D4; O2/O5 | Відповідь записана; нові domain limits поза scope |
| 5 | Rollback | §Plan-defense D5; §Proposed implementation / migration / rollback | Відповідь записана |

Незаповнених defense-питань: **0/5**. Це завершення авторської структурної defense, не runtime proof, не незалежний verdict і не зняття operational blockers.

**Environment owner / capability contract:** Віталій підтверджує exclusive writer і власника isolated environment/heavy window (ім’я/checkout/ports/slot ще pending); призначений виконавець після допуску перевіряє isolated production serving з enforcing CSP для O1, per-request API-boundary counter для O2, delayed/cancellable upstream route для O4 та data isolation для full E2E. Не перекладати proof на непідтверджений чужий server. Якщо capability відсутня, відповідний check BLOCKED, вибрати узгоджений еквівалентний доказ через reviewer або надати capability; не заявляти, що runtime check пройшов.

2026-10-18 defense journal: прочитано binding-matched ACCEPTED v2 від Courses2, B1–B3 resolved і N1–N12 addressed; додано п’ять авторських defense-відповідей, 0 структурних дірок. Install/build/code/runtime/commit NOT_RUN; writer/runtime coordination pending; manual v3 delta handoff prepared, delivery NOT_CONFIRMED.

2026-10-19 implementation (owner go, isolated worktree `~/project-data/courses-tanstack-start-pilot/checkout` @ 4c98a7f; no commit/push/deploy). Backup package/lock (sha256 recorded). Baseline `npm ci` exit 0. Upgrade `npm install --save-exact @tanstack/react-start@1.169.3 @tanstack/react-router@1.170.45 @tanstack/react-router-ssr-query@1.167.3` (registry.npmjs.org) exit 0. **A1:** each of 8 target/core packages one resolved version, no invalid peers, react-router@1.170.45 deduped (single runtime — closes B1); lock diff tanstack-only + transitive closure; no new install scripts. Transitive: srvx 0.11.22→1.0.1 (major) — server.mjs symbols `serve`/`staticMiddleware` still exported (signatures NOT runtime-verified → O3). **Two minimal compat fixes:** (1) `teach.courses.$courseId.tsx` ErrorComponentProps.error Error→unknown (adopt ErrorComponentProps + narrow; cleared cascaded useLoaderData error); (2) `root-identity.test.tsx` Start removed `requestHandler` → migrated to `createServerEntry` (request scope) + router-core `createRequestHandler` (rstest alias omits the SSR handler), all G10/G15/G22/G23 assertions unchanged. **Gates (isolated):** format:check 0, lint 0 (arch boundaries OK), typecheck 0, test 0 — **773 passed/0 failed**, build 0 (dist/server/index.js emitted), build-storybook 0. Task-owned diff = 4 files (package.json, lock, 2 compat fixes); foreign AI Factory 0.7 dirty files byte-preserved. Evidence: `~/project-data/courses-tanstack-start-pilot/impl/IMPL-SUMMARY.md`. **Still gated (post-implementation acceptance):** A4 full E2E + O1–O5 (coordinated isolated Compose stack), independent implementation review, commit/PR. Upgrade NOT declared complete.

2026-10-19 A4 + O1–O5 (owner go, coordinated isolated Compose): owner's head `courses` e2e stack brought down (volumes preserved); checkout run under isolated project `tsstart` (COMPOSE_PROJECT_NAME override of compose `name: courses`), own volumes, default ports, temp DOCKER_CONFIG (no credsStore). **Lockfile gotcha S0:** host npm 11.6.2 produced a darwin-biased lock → in-container `npm ci` failed (missing `@emnapi/*`, `@floating-ui/dom`); regenerated platform-complete via node:24 + npm@11 `npm install --package-lock-only` (per docs/admin+foundation precedent), re-verified darwin `npm ci`/test 773/build. **`make e2e-local`: 121 passed / 0 failed (14.7m).** **O2 PASS** (live boundary: anonymous SSR → 0 `/api/v1/auth/me`; cookie control → 1). **O1 PASS** (enforcing CSP; two concurrent anonymous SSR → enforcing header, per-request non-empty nonces matching own scripts, nonces differ). **O3 happy-path** proven (server.mjs/srvx 1.0.1 served all 121 E2E); **O3 throw-path PASS** (isolated srvx-1.0.1 harness → 503 `Rendering unavailable`/no-store/render_failed). **O4** covered by unit G10 (1s deadline), **O5** by unit login/catalog-search + E2E route-guards. tsstart torn down (`down -v`), owner's `courses` volumes untouched. **Remaining:** independent implementation review (fresh existing-reviewer handoff), then commit/PR. Upgrade still NOT declared complete.

2026-10-19 branch-workflow reconciliation (owner routing override): the normal route is now a **task branch in the main repo `/Users/strelia/projects/courses`, one active writer, no automatic worktree/clone/detached checkout** (`docs/development/ai-factory.md` §"Owner routing override — task branches"; shared `~/.agents/ai-factory/workflow.md#task-branches`). This narrow owner override wins over the pinned bundle's Git-isolation defaults; pin/cache and all local gates are unchanged. Consequences for TSSTART-01: (a) the isolated worktree `~/project-data/courses-tanstack-start-pilot/checkout` is **INACTIVE** — its implementation artifacts (A1, two compat fixes, gates 773/0, A4 E2E 121/0, O1–O5) are **preserved as scoped prior evidence**, not an automatic acceptance under the new workflow; (b) the task-owned diff (4 files) must be **coordinated onto an agreed main-repo task branch**, not auto-copied into the main tree; (c) pilot-root `HANDOFF.md` is **SUPERSEDED** (not current routing); (d) the earlier Courses2 implementation ACCEPTED stands as prior evidence, but the **branch/writer/WIP/runtime execution delta** needs a fresh Courses2 reconciliation (ai-factory.md review routing). **Agreed task-branch facts (proposed, pending owner confirm):** root `/Users/strelia/projects/courses`; base = current `main` HEAD `71a0d04`; one active writer = this session; task-owned diff = 4 files with sha256 — `frontend/package.json 9d06c9fc…`, `frontend/package-lock.json 04c416e5…` (authoritative platform-complete lock), `frontend/src/routes/teach.courses.$courseId.tsx 1db681cd…`, `frontend/src/app/composition/root-identity.test.tsx 25b9d75a…`. **Open coordination blocker:** the main tree currently carries foreign uncommitted WIP (AI Factory 0.7 rollout + this workflow correction: `AGENTS.md`, `docs/development/ai-factory-*`, `work/tasks.md`, `work/archive/ledger.md`); per the task-branch contract a branch must not be cut over foreign dirty WIP and that WIP must not be cleared without separate authorization — owner must decide ownership of that WIP / authorize cutting the task branch beside it before the diff transfer and commit/PR proceed. Runtime/test isolation (compose project `tsstart`, ports, volumes, heavy-gate slot) remains a separate contract, already exercised for A4. No branch, commit, push or deploy performed by this reconciliation.

2026-10-19 landed: plan recorded direct to `main` (`2f4e4ab`, docs). Task branch `tsstart-01-tanstack-start-1.169.3` (base `main` `2f4e4ab`) carries the 4 task-owned files only (`7f5c5a4`); foreign WIP left uncommitted/byte-preserved per owner Variant A. **PR #56** opened (`https://github.com/b2vv/courses/pull/56`). Courses2 [eb5c1a] delta-review **ACCEPTED** (hashes verified, scope + foreign-WIP preservation confirmed, implementation ACCEPTED bound under the task-branch workflow). **CI is red for an account-wide GitHub Actions provisioning failure** (every recent run, incl. merged PR #52–#55 and main pushes, fails at startup in 3–4s, `runner_id 0`) — not this change. **Compensating main-context gate-close (owner go, task branch, upgraded deps via `npm ci`):** format:check 0, lint 0 (arch boundaries 221 files), typecheck 0, test **773/0**, build 0 (`dist/server/index.js` emitted). Shared tree restored to `main` baseline afterward. Pending: owner restores Actions billing → rerun CI on #56; then merge.
