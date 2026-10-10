ДАЛІ: прочитати обов’язкові інструкції implement і work/brief-t7.md; перед змінами перевірити контракти та NOT-обмеження.

ФАКТ: створено відсутній `work/report-t7.md`; початковий запис `ДАЛІ:` зафіксовано. Перша спроба append не змінила файлову систему, бо файл був відсутній.

ФАКТ: прочитано `implement`, `work/CTO-RESEARCH.md` і `work/brief-t7.md`. T7 обмежений frontend-плеєром, таблицею відновлення та запускабельним O1 e2e; backend, edge, міграції, generated, playback-документи, tasks і git заборонені. Docker-гейти не запускаються. Конфлікт із загальним `implement` щодо коміту вирішено на користь конкретної заборони T7: git не торкаюся.

ДАЛІ: прочитати TDD-інструкцію та всі названі T7 контракти; зупинитися, якщо хоч одного контракту бракує.

ФАКТ: TDD-інструкцію прочитано повністю; контрактні файли та всі названі секції/рядки існують. Узгоджені seams уже задані brief/plan: компонентний public UI seam для B02–B04/B24 та браузерний route/network seam для P01/P08–P11/B24a/B25/B39/B40/R04/O1. Виправлено технічну мітку першого запису на дозволену `ФАКТ`.

ДАЛІ: дочитати точний текст D-1/D-6/D-7/D-8, §8-рядків, O1, spec-сценаріїв, T5/T6 контрактів, composition/transport/e2e-конфігурації та локальних стандартів.

ФАКТ: перший пакет читання перевищив ліміт виводу й був обрізаний; його не зараховано як повне прочитання контрактів. Видима частина підтвердила `play`→mint→`loadSource`, один ownership key, 409 takeover, 410 expiry/remint, learning-модель і заборону перевіряти HLS-файли в API.

ДАЛІ: перечитати названі секції малими неперекривними блоками до EOF, не покладаючись на обрізаний вивід.

ФАКТ: D-6/D-7/D-8, O1 і потрібні spec-сценарії/A11–A17 прочитано. Плановий D-7 містить застарілі формулювання, які brief уже виправляє: 403-клієнт не розрізняє takeover від archive; B40 не очікує 404 при штатній заміні, бо дві генерації збережені. Це не нова невиявлена неточність і не окремий `## Discussion`; реалізація слідує виправленому brief.

ФАКТ: другий пакет також обрізано на великій §8-таблиці; потрібні target-рядки видимі, але повні T5/T6 звіти ще не зараховані.

ДАЛІ: прочитати target-рядки §8 точним фільтром, а T5/T6 — послідовними блоками; потім composition, transport, Playwright і AGENTS.

ФАКТ: усі 14 цільових рядків §8 знайдено точним фільтром. `work/report-t5.md` має 301 рядок, `work/report-t6.md` — 495; спроба прочитати перші 180 рядків T5 разом із inventory знову обрізалась, тому переходжу на блоки по 100 рядків.

ДАЛІ: прочитати `work/report-t5.md` 1–100, 101–200, 201–301 окремо й фіксувати кожен блок.

ФАКТ: `work/report-t5.md` 1–100 прочитано повністю. Public contracts: `POST /api/v1/playback/sessions` приймає рівно `{lessonId}`, повертає 201 з server-minted 32-hex `playbackId`, `mediaPath`, cadence і `clientSeq`; heartbeat має distinct 204/400/403/404/409/410/503. `ready` — DB-передумова, missing bytes обробляє T7/B28.

ДАЛІ: прочитати `work/report-t5.md` 101–200.

ФАКТ: `work/report-t5.md` 101–200 прочитано повністю. Для T7 важливі точні transport invariants: 409/410 належать heartbeat, `play` не повертає їх; generated client мав бути materialized рецензентом; runtime smoke/generation були Docker-only і не приписувались локальному прогону.

ДАЛІ: прочитати `work/report-t5.md` 201–301.

ФАКТ: `work/report-t5.md` 201–301 прочитано повністю. Рецензент підтвердив materialized generated client, повний `make check`, policy-відмови через справжній mint і прийнятий B28 recovery contract. Отже T7 може спиратися на generated `createPlaybackSession`/`heartbeatPlayback` та exact responses.

ДАЛІ: прочитати `work/report-t6.md` п’ятьма блоками 1–100 … 401–495.

ФАКТ: `work/report-t6.md` 1–100 прочитано повністю. T6 зафіксував learning endpoint, deterministic `continueLessonId`, видимі no-video/processing/failed стани та проєкцію media state: newest non-evicted ready wins; інакше newest asset, причому created/uploading/queued/processing → processing.

ДАЛІ: прочитати `work/report-t6.md` 101–200.

ФАКТ: `work/report-t6.md` 101–200 прочитано повністю. Learning transport/OpenAPI/composition реалізовано окремо від public catalog; response має nullable `videoState` і exact default progress. Query/read path request-scoped, без нових globals.

ДАЛІ: прочитати `work/report-t6.md` 201–300.

ФАКТ: `work/report-t6.md` 201–300 прочитано повністю. Критичне для T7: heartbeat після T6 перевіряє course visibility до renewal/write, тому archive дає 403 `playback_forbidden`; R17 виправив реальний write bug. Docker-only checks лишались named limitations.

ДАЛІ: прочитати `work/report-t6.md` 301–400.

ФАКТ: `work/report-t6.md` 301–400 прочитано повністю. Остаточний heartbeat 403 контракт посилається на `PlaybackForbidden`; `Play` і `Heartbeat` поділяють `loadViewablePlayback`. T6 фінальні review-осі не мали findings після виправлень.

ДАЛІ: прочитати `work/report-t6.md` 401–495.

ФАКТ: `work/report-t6.md` 401–495 прочитано повністю; обидва попередні звіти тепер прочитані до EOF. Рецензент підтвердив T6 generated learning client і повний `make check`, тож named frontend contracts наявні й перевірені.

ДАЛІ: прочитати composition/query polling pattern, transport, Playwright config та локальний `AGENTS.md`; потім інвентаризувати фактичний frontend і перевірити, чи немає missing contract.

ФАКТ: required composition, transport, Playwright config і локальний `AGENTS.md` прочитано повністю. Request-scoped Query factory і shared transport є; презентації/routes не можуть робити fetch. Існуючі media/recovery e2e серіалізовані через окремі Playwright projects.

РІШЕННЯ: новий `playback.spec.ts` піде в новий ізольований `playback` project після `media`, а `recovery` залежатиме від нього: він потребує вже транскодованої media fixture та виконує довгі/відновлювальні media flows, тому не може працювати поруч із `chromium`; окремий project дає адресний O1 command/output і не змішує T7 з Phase-5 media journey.

ДАЛІ: виконати вимогу AGENTS — дочитати повний phase plan і `work/tasks.md`, потім інвентаризувати frontend seams.

ФАКТ: `docs/playback/plan.md` 1–180 прочитано. Placement підтверджує `frontend/src/{domain,ports,adapters/api}/playback`, `presentation/learn`, `routes/learn.*`, composition `learn.ts`/`learn-queries.ts`; hls.js — єдина нова frontend dependency.

ДАЛІ: прочитати `docs/playback/plan.md` 181–360.

ФАКТ: `docs/playback/plan.md` 181–360 прочитано. Confirmed server semantics: failed `play` never changes ownership; heartbeat `reason` closed enum; 409/410 distinct; B28 edge 404 is legitimate after mint.

ДАЛІ: прочитати `docs/playback/plan.md` 361–540.

ФАКТ: `docs/playback/plan.md` 361–540 прочитано. Plan D-7 legacy statements remain superseded by brief corrections. Frontend architecture and machine gates are explicit; B28 requires a visible error state, not a white screen.

ДАЛІ: прочитати `docs/playback/plan.md` 541–720 і `work/tasks.md` повністю.

ФАКТ: `docs/playback/plan.md` 541–720 прочитано; phase plan прочитано до EOF. Останній абзац §11 суперечить попередньому тексту й таблиці, помилково кажучи, що T1 закриває O1; brief-t7 явно й правильно призначає O1 T7. Оскільки brief уже виправляє цю історичну неточність і реалізація не спиратиметься на помилковий абзац, окремого нового Discussion не відкриваю.

ДАЛІ: прочитати `work/tasks.md` повністю.

ФАКТ: `work/tasks.md` прочитано до EOF. T1–T6 done, T7 planned; prerequisites complete. Tracker itself за NOT не змінюю. Повний перелік обов’язкового читання та phase-context завершено без missing contract.

ДАЛІ: інвентаризувати frontend domain/application/ports/adapters/routes/tests/styles/generated contracts і package scripts; визначити мінімальні TDD slices без T8/T9/T10/T12.

ФАКТ: frontend inventory підтвердив materialized generated learning/playback operations і models. `hls.js` ще відсутній, тому його можна додати лише разом зі споживачем. Поточний learn surface — layout + splat placeholder; окремих playback domain/port/adapter/composition/presentation файлів ще немає. Широкий rg-output обрізано після inventory, тому code details читаю адресно.

ДАЛІ: прочитати точні route/context/composition patterns, generated shapes, API adapter schemas, UI primitives, styles і e2e fixtures.

ФАКТ: current route/context/composition/API patterns прочитано. `/learn` зараз client-only protected placeholder, `/learn/$` — порожній splat; root context не має learning dependency. Catalog patterns show request-scoped isomorphic factory, query options and explicit zod mapping.

ДАЛІ: знайти в усіх phase artifacts точний lesson route contract; якщо exact URL ніде не визначено й від нього залежить public behavior, застосувати канал «Стоп» до коду.

ФАКТ: exact lesson URL pattern не заданий, але контракт вимагає лише URL із slug та lesson id; existing protected-path tests already use two path segments under `/learn`. Це не missing external contract: обираю мінімальний `/learn/$slug/$lessonId`, де обидві contract values видимі й T8 зможе будувати links без query state. API request/response contracts повні.

ДАЛІ: прочитати `frontend-design` skill перед UI-рішеннями, потім exact generated/domain/e2e/style seams.

ФАКТ: `frontend-design` skill прочитано повністю. Artifact — system; subject: навчальний відеоурок для автентифікованого студента; one job: надійно почати/продовжити HLS playback і ясно назвати стан або відмову. Palette/type/radius/spacing лишаються закритими existing token system; свобода — компактна media-console композиція та повторювана status-strip signature.

ДАЛІ: прочитати skill `WRITING.md`, token/theme/component/style sources і exact generated API shapes.

ФАКТ: `WRITING.md`, tokens, core styles і UI primitives прочитано. Token source: `tokens.css`; theme: `prefers-color-scheme` plus root `data-theme`; components: React Aria `Button`/`Select`/states; styling idiom: semantic BEM-like classes in `main.css`; parity gates: plan §8, axe unit and real-browser accessibility. Closed axes remain unchanged. Signature: a thin utility-font status strip above the media frame, encoding state without decorative card chrome.

ДАЛІ: прочитати exact generated models/functions, existing test setup, e2e media fixture lifecycle and accessibility harness; then define first RED slice.

ФАКТ: exact generated learning/session/heartbeat shapes, Orval calls, unit setup, media fixture generation and accessibility harness прочитано. Generated calls accept request-scoped `RequestInit`; heartbeat 204 has no body; media project deletes its temporary course, so T7 e2e needs its own course/upload fixture but can reuse generated media files.

## Discussion — нативний HLS не дає клієнту HTTP status сегмента

ФАКТ: §2.2/D-7 вимагає розгалужувати segment failures за 403/404/503 і водночас підтримати нативний Safari branch. У hls.js `ERROR` містить response code, але нативний `<video src="…m3u8">` повідомляє лише media error/stall; Fetch/XHR response status внутрішніх playlist/segment requests JavaScript не отримує. Отже однакова status-based таблиця не реалізується для обох гілок суто frontend-кодом.

ПИТАННЯ: без server/edge side-channel нативна гілка не може відрізнити «архівовано/витіснено» від «байтів немає» чи тимчасового 503 в момент media error. Це може або безкінечно ретраїти terminal 404/403, або передчасно зупиняти recoverable 503.

РІШЕННЯ: у межах NOT не змінюю backend/edge. hls.js branch виконає exact 403/404/503 table; native branch покладається на exact heartbeat 403/409/410 для policy/ownership, а opaque media error робить bounded backoff retry після reread learning model і потім показує named failure state. У звіті це лишається відкритою contract limitation, не видається за повний parity proof.

ДАЛІ: прочитати adapter/unit/e2e test conventions і решту media lifecycle; написати перший RED slice для learning adapter/states.

ФАКТ: adapter/unit/e2e conventions і media lifecycle прочитано. Media project deletes its course in `afterAll`; T7 cannot reuse that DB fixture. New isolated playback project will create/publish/upload its own course inside a serial spec, reuse `playwright/.fixtures` bytes, and clean up. Tests validate observable API/network/UI outcomes, not private hls.js state except O1’s required browser measurements. Попередня append-спроба не змінила файл через неточний patch context.

ДАЛІ: перевірити test-profile heartbeat TTL, route splat params and package-lock; потім RED adapter slice.

ФАКТ: search found no playback timing variables in Compose or frontend; only backend config unit tests instantiate 2/6. Frontend has one splat route and a package lock; `hls.js` is absent. Exact Compose service block still needs inspection before classifying P08 testability.

ДАЛІ: прочитати API Compose block and env interpolation. Якщо 2/6 profile справді не wire-иться, записати Discussion before code.

## Discussion — P08 test profile неможливо увімкнути в Compose

ФАКТ: `docker-compose.yml` не передає `PLAYBACK_HEARTBEAT_SECONDS` або `PLAYBACK_OWNERSHIP_TTL_SECONDS` у `api`; production process тому завжди бачить defaults 30/180. Лише Go unit test викликає config із 2/6. Команда на кшталт `PLAYBACK_HEARTBEAT_SECONDS=2 docker compose up` зараз нічого не змінює.

ПИТАННЯ: P08 вимагає 30 секунд перегляду в test profile 2/6 і ≥4 справжні heartbeat renewals. Без Compose wiring тест або триває понад 180 секунд, або підміняє response/Redis і стає хибнопозитивом.

РІШЕННЯ: додати до `api.environment` лише pass-through із незмінними defaults `${PLAYBACK_HEARTBEAT_SECONDS:-30}` і `${PLAYBACK_OWNERSHIP_TTL_SECONDS:-180}`. Це не backend/edge code, не змінює default runtime і робить exact reviewer command запускабельною. E2E project сам не мутує Compose; reviewer явно recreates `api` з 2/6 перед адресним прогоном і повертає defaults після.

ДАЛІ: написати RED adapter/domain slice для exact learning/mint/heartbeat outcomes before production code.

ФАКТ: перший RED slice додано на public adapter seam: learning slug validation, explicit-play mint ordering/body, exact play errors і distinct heartbeat 204/403/404/409/410/503. Production files ще не створено.

ДАЛІ: запустити лише adapter test і зарахувати RED тільки з причини відсутнього production module.

ФАКТ: adapter test RED як очікувалось: 0 tests executed, build failed only because `./playback-repository` did not exist.

ФАКТ: мінімальний GREEN implementation додано: inward playback domain/result, external repository port, strict zod schemas and generated-call adapter. Session validation pins mediaPath↔playbackId and heartbeat×3≤TTL; statuses remain distinct.

ДАЛІ: прибрати небезпечний TypeScript assertion у external-error guard до запуску, потім format і targeted adapter test.

ФАКТ: unsafe external-error assertion прибрано; code is narrowed by lookup in the status-specific allowed code list. External input remains zod-validated.

ДАЛІ: запустити formatter on new files and targeted adapter test.

ФАКТ: `npx oxfmt` на п’яти new adapter/domain files PASS, exit 0.

ДАЛІ: запустити targeted adapter test.

ФАКТ: targeted adapter suite GREEN: 10/10 tests, adapter-node project, exit 0.

РІШЕННЯ: hls/native media інтеграція матиме окремий port backed by hls.js/HTMLMediaElement. Це реальна external dependency, дозволяє монтувати engine без source і тестувати UI через public actions без module mocks.

ДАЛІ: додати RED component tests for B02–B04 and B24: no broken video, explicit play, keyboard controls, hls-only quality and axe.

ФАКТ: RED component slice added. It asserts no `<video>` for null/processing/failed; ready player mounts media without source, mints only on explicit keyboard play, restores position, exposes seek and hls-only quality, and passes axe.

ДАЛІ: запустити only `LessonPlayer.test.tsx`; expected RED must be missing component/media port.

ФАКТ: component suite RED as expected: 0 tests executed; build failed only because `LessonPlayer` does not exist. This is the intended red edge before media/component production.

ДАЛІ: add the sole permitted dependency `hls.js` pinned by npm, immediately followed by its media adapter consumer.

ФАКТ: `npm install --save-exact hls.js` succeeded; one package added, audit 0 vulnerabilities. No other dependency requested or added.

ДАЛІ: implement PlaybackMedia port, hls.js/native adapter and minimal LessonPlayer GREEN behavior for the current tests.

ФАКТ: installed `hls.js` version is 1.7.2. Local primary types/source confirm `Hls.isSupported`, `attachMedia`, `loadSource`, `stopLoad`, `MANIFEST_PARSED`, `LEVEL_SWITCHED` and typed `ErrorData`.

ДАЛІ: add media port/adapter and LessonPlayer consumer now that dependency API is verified.

ФАКТ: media port, hls.js/native adapters and `LessonPlayer` added. hls.js attaches at mount without calling `loadSource`; explicit play mints then loads. 410 stops loading before one debounced remint and restores time on ready; 403 is terminal+refresh without mint; 409 stops; 404 refreshes and allows one reload; 503/opaque native errors use bounded exponential backoff.

ФАКТ: localized state/control copy and existing-token CSS added; controls preserve ≥44px and responsive one-column layout. Compose pass-through for exact P08 2/6 profile added with unchanged 30/180 defaults, per Discussion.

ДАЛІ: format changed frontend files, then run component suite and fix only observed GREEN failures.

ФАКТ: first format invocation did not run: oxfmt rejected path `../docker-compose.yml` because paths may not contain `..`; exit 1. No format result is claimed.

ДАЛІ: run oxfmt from repository root with explicit in-root paths.

ФАКТ: root-scoped oxfmt on eight changed files PASS, exit 0.

ДАЛІ: run `LessonPlayer` suite and inspect actual GREEN failures.

ФАКТ: component suite 3/4 PASS. B02–B04 pass. B24 failed because happy-dom does not apply native range ArrowRight behavior (`currentTime` stayed 12); this is test-environment behavior, but an explicit key handler also makes keyboard seeking deterministic across browsers.

ДАЛІ: add exact ArrowLeft/ArrowRight seek behavior to the public range control, then rerun B24.

ФАКТ: range control now handles ArrowLeft/ArrowRight explicitly in 5-second bounded steps while retaining native change handling. This is observable keyboard behavior, not a test-only hook.

ДАЛІ: format LessonPlayer and rerun component suite.

ФАКТ: LessonPlayer format PASS.

ДАЛІ: rerun component suite.

ФАКТ: report tail contains additional B24 iterations after the earlier snapshot; the latest source change synchronizes duration in onReady and leaves the suite rerun pending. Process inspection itself was denied by the sandbox (`ps: operation not permitted`) and establishes no runtime fact.

ДАЛІ: inspect the current LessonPlayer/test files as the source of truth, then run the focused suite once if the sandbox permits it.

ФАКТ: current source includes the duration snapshot, single React Aria quality label, the unknown-duration/adapter-stability regression, and the completed public keyboard sequence. No conflicting source edit is present.

ДАЛІ: run the focused LessonPlayer suite once against the current source.

БЛОКЕР ПЕРЕВІРКИ: current focused LessonPlayer suite again cannot start because Rstest is denied binding ::1:3000 (`EPERM`); no assertion result is claimed after the duration/lifecycle changes.

ДАЛІ: append the mandatory §8 proof matrix, exact reviewer O1 commands, project rationale and currently known unrun gates before the final review/gate pass.

## Рядок §8 → чим доводиться

| Рядок | Доказ |
|---|---|
| P01 | `frontend/e2e/playback.spec.ts` — real ready HLS, explicit play, growing `currentTime`, media 200 |
| P08 | `frontend/e2e/playback.spec.ts` — asserts exact 2/6 profile, 30 seconds, ≥4 successful heartbeats, no segment 403 |
| P09 | `frontend/e2e/playback.spec.ts` — selects 360p through the hls.js-only control and observes a 360p segment 200 |
| P10 | `frontend/e2e/playback.spec.ts` — seeks to 80 seconds and observes playback beyond 70 plus a newly requested rung segment |
| P11 | `frontend/e2e/playback.spec.ts` — admin opens the archived course and observes an HLS segment 200 |
| B02 | `frontend/src/presentation/learn/LessonPlayer.test.tsx` — null video state renders text and no `<video>` |
| B03 | `frontend/src/presentation/learn/LessonPlayer.test.tsx` — processing renders a state and no `<video>` |
| B04 | `frontend/src/presentation/learn/LessonPlayer.test.tsx` — failed renders a state and no `<video>` |
| B24 | `frontend/src/presentation/learn/LessonPlayer.test.tsx` — keyboard play/seek/quality path plus axe |
| B24a | `frontend/e2e/accessibility.spec.ts` — keyboard order, visible ≥3px focus ring and measured ≥44px play/seek targets in Chromium |
| B25 | `frontend/e2e/accessibility.spec.ts` — dark + reduced-motion + 200% at 390px, no overflow/overlap; B24 axe supplies automated contrast analysis and the existing suite covers light theme |
| B39 | `frontend/e2e/playback.spec.ts` — pause beyond TTL, resume causes exactly the second mint, no takeover message, time continues |
| B40 | `frontend/e2e/playback.spec.ts` — replace while playing, old generation has no 403/404 and ends, next opening gets a different video path |
| R04 | `frontend/e2e/playback.spec.ts` — archive while playing, real heartbeat/media 403, named terminal state |

## O1 — точні команди для рецензента

```sh
PLAYBACK_HEARTBEAT_SECONDS=2 PLAYBACK_OWNERSHIP_TTL_SECONDS=6 docker compose up -d --force-recreate api
cd frontend
npm run e2e -- --project=playback --grep 'B39 O1' --reporter=line
cd ..
docker compose up -d --force-recreate api
```

ФАКТ: test prints: old hls.js requests in flight immediately before remint; old requests still in flight when the second mint begins after `stopLoad`; whether that count is zero; `currentTime` immediately before the explicit restore, the assigned restore, and pre-remint time. No plan-assumed number is asserted.

РІШЕННЯ: `playback` is a new isolated Playwright project ordered `media → playback → recovery`, excluded from `chromium`. It owns a real five-minute transcode, a 30-second renewal observation and expiry/remint network accounting; serialization prevents CPU saturation and request-count races with the Phase-5 media journey while keeping O1 directly addressable.

НЕ ЗАПУЩЕНО: Docker is unavailable by brief constraint, so `make integration`, `make build`, `make up`, `make e2e` and `make check-generated` are not run. Browser O1 numbers are intentionally left to the reviewer. Rstest is additionally blocked in this sandbox by `listen EPERM ::1:3000`; it is not represented as passed.

ДАЛІ: run the required code-review discipline against the current T7 files, resolve blocking findings, then execute the permitted frontend gates in the required order.

РІШЕННЯ: code-review skill normally requires a git fixed point, but T7 explicitly forbids touching git and supplies no ref. I will preserve the stronger brief rule: two independent read-only reviews will inspect the enumerated T7 working set directly, one against AGENTS/architecture standards and one against `work/brief-t7.md`; no git command will be used.

ДАЛІ: launch the two required independent read-only review axes in parallel and continue local contract/lifecycle inspection.

ФАКТ: independent standards and specification reviewers are running read-only against the explicit T7 file set; both were forbidden from git, edits, and report writes.

ДАЛІ: locally inspect the complete playback media adapter, API adapter, domain/ports and composition for lifecycle or boundary errors while reviews run.

ФАКТ: local lifecycle review found two code defects before handoff: a successful 410 remint returns `stop`, so the existing heartbeat loop never schedules again; and resetting media retry count at every manifest allows a repeatedly failing segment to retry forever. Both are implementation defects, not contract inaccuracies, so they do not open a Discussion.

ДАЛІ: add focused recovery tests that pin post-remint heartbeat continuation and bounded consecutive media retries before fixing the callbacks.

ФАКТ: component tests now pin a second heartbeat after silent 410 remint and terminal failure after three consecutive manifest-success/segment-503 reloads.

ДАЛІ: make 410 return to the heartbeat scheduler, reset media retry budget only when playback time actually advances, then format/typecheck the slice.

ФАКТ: successful 410 recovery now returns `continue` so cadence resumes; media retry count no longer resets at manifest parse and instead resets on observable playback progress, bounding consecutive failures while forgiving later isolated failures.

ДАЛІ: format LessonPlayer and its test, then typecheck the new regression slice.

ФАКТ: formatter passes on LessonPlayer, its tests and the adjusted playback e2e.

ДАЛІ: run frontend typecheck for the new recovery tests and callbacks.

ПЕРЕВІРКА: frontend typecheck passes with the new 410-cadence and bounded-retry regressions.

## Review — Standards

ФАКТ: independent standards review found one blocking architecture issue (recovery policy concentrated in presentation), two medium issues (concrete HLS adapter wired in presentation; hard-coded frame color), plus low duplicated test geometry. No other standards findings.

РІШЕННЯ: move recovery decisions/backoff calculation into an application module that imports only domain types; inject the concrete media factory from the composition root; add a semantic media-frame token. Extract the repeated test geometry only if it remains duplicated after the focused changes.

ДАЛІ: inspect token and composition test shapes, then perform the standards refactor without expanding T7 behavior.

ФАКТ: tokens.css has a single semantic palette with dark overrides; one root `--color-media-frame` token can replace the hard-coded frame color. Query composition currently returns repository operations only, and route-guard tests construct that object directly. The attempted router read used a stale path and stopped at the missing file without changing anything.

ДАЛІ: locate the real router/composition files, then add recovery-policy unit seams and composition-root media injection.

ФАКТ: added an application-layer recovery policy with exact heartbeat/media directives, backoff calculation and direct unit cases; this moves status-to-action decisions out of presentation. Router source is `frontend/src/router.tsx` and already creates request-scoped learn dependencies.

## Review — Specification

ФАКТ: independent spec review found blocking gaps: mint-side 503 was terminal; O1 sampled one count before the TTL wait; R04 accepted heartbeat and skipped the specified >30s segment proof. Medium gaps: B40 lacked a positive old-segment 200 assertion, P10 did not prove which segment covers 80s, B25 did not measure both themes/contrast, and hls.js forwarded nonfatal statusless errors into destructive reloads. The media retry-reset issue was independently found and already fixed before this report arrived. No scope creep found.

ДАЛІ: extend the application policy for bounded mint retry, consume all recovery directives from LessonPlayer, and distinguish nonfatal/fatal hls.js errors before fixing browser proof gaps.

ФАКТ: application policy now includes bounded exponential retry for dependency/timeout mint failures, exact terminal/refresh outcomes, and a retryability flag so opaque native failures can retry while fatal statusless hls.js failures cannot.

ДАЛІ: update the media port/adapters to emit retryability and refactor LessonPlayer to consume application directives and manage the mint retry timer.

ФАКТ: retryability touches only the PlaybackMedia port, hls/native adapter, LessonPlayer, and two synthetic 503 calls in its test; no other consumer exists.

ДАЛІ: implement that narrow media-error contract update.

ФАКТ: hls.js now reports only HTTP-status errors or fatal statusless failures; nonfatal statusless events are ignored. Native opaque errors remain retryable. The port makes this distinction explicit.

ДАЛІ: inspect current LessonPlayer callback blocks and replace inline decisions with the application directives plus bounded mint retry.

ФАКТ: LessonPlayer now consumes application directives for mint, heartbeat and media failures; dependency mint failures retry at 500/1000/2000ms before a named unavailable state, and retry timers are cleared on stop/unmount. Heartbeat delay calculation also moved inward.

ДАЛІ: inject the concrete media factory through learn composition/root route, update test harness injection, and add the semantic frame token.

ФАКТ: concrete hls/native adapter wiring now lives in `app/composition/learn.ts` and flows through route context; LessonPlayer requires the injected port. The test harness provides an explicit unused fake. Player frame color now uses `--color-media-frame` from tokens.css.

ДАЛІ: repair the O1/P10/B40/R04 browser proofs exactly as specified, then extend B25 to both themes with real axe contrast checks.

ФАКТ: media generation pins six-second HLS targets and rendition playlists expose exact `#EXTINF` durations plus segment names. P10 can therefore derive the actual segment covering second 80 from the authenticated playlist instead of assuming a filename index.

РІШЕННЯ: keep B40 before irreversible archive but replace with the five-minute fixture; R04 can then play >30 seconds and demand a segment 403 exactly as §8 states.

ДАЛІ: implement the exact playlist-derived P10, corrected O1 sampling, positive B40 200, and strict R04 evidence.

ФАКТ: P10 now parses the authenticated rendition playlist and waits for the exact segment whose EXTINF interval covers 80s. O1 samples immediately after TTL wait and before resume. B40 requires an old-generation 200 and uses a five-minute replacement. R04 plays 31s, then requires a media-prefix 403 after archive.

ДАЛІ: inspect and extend B25 to exercise both color schemes with axe `color-contrast`, retaining the 200% dark/mobile layout measurements.

ФАКТ: current B24a/B25 test performs dark/200% geometry only; axe is already imported and exposed on `window`, so both live color schemes can use the real browser color-contrast rule without new dependencies.

ДАЛІ: add light and dark browser contrast assertions before the existing dark/200% geometry checks.

ФАКТ: B25 now runs axe's real-browser `color-contrast` rule in both light and dark schemes, then applies dark-mode 200% mobile geometry assertions.

ДАЛІ: make the hls.js auto-quality control present before mint so accessibility.spec can measure every rendered player control, then extend B24a focus/size assertions to it.

ФАКТ: hls.js branch now exposes its Auto quality control as soon as the adapter mounts; rendition choices populate after manifest. B24a measures and keyboard-focuses play, seek and quality, each at ≥44px.

ДАЛІ: make API-to-domain mappings explicit, then format/typecheck and inspect the resulting complete files before rerunning both review axes.

ФАКТ: validated wire payloads are now explicitly copied into PlaybackSession/LearningCourse domain objects; adapter-only timestamps and generated shapes no longer leak inward structurally.

ДАЛІ: format all files changed by review fixes, then run typecheck.

ФАКТ: formatter passes on all 14 files touched by review fixes.

ДАЛІ: run frontend typecheck for policy extraction, composition injection and strengthened e2e proofs.

ПЕРЕВІРКА: frontend typecheck passes after all first-round standards/specification fixes.

ДАЛІ: inspect the complete recovery policy, player and strengthened e2e code for timing or cleanup mistakes before the second independent review.

ФАКТ: inspection found two remaining player lifecycle defects: mint backoff recursively released `mintRef` between attempts, so it was request-debounced rather than player-debounced; and terminal failure unmounted the media element, so the visible explicit retry could mint before a new adapter existed and drop the source. Stop also did not clear a pending media retry.

ДАЛІ: make one mint promise span all dependency retries, cancel retry waits on stop/unmount, and remount the adapter before explicit retry minting.

ФАКТ: one player-level mint promise now spans all backoff attempts, so concurrent triggers share it; stop/unmount clears and resolves pending waits. Explicit failure retry first clears the terminal state, remounts the adapter, then mints through an effect, preventing a dropped mediaPath.

ДАЛІ: add component regressions for mint debounce/backoff and explicit terminal retry, then format/typecheck.

ФАКТ: localized public failure/retry labels are `Playback stopped. This course is no longer available.` and `Start playback again`, providing a stable observable seam for the explicit-action regression.

ДАЛІ: add the two component regressions through public controls and media callbacks.

ФАКТ: component coverage now requires one continuous mint cycle across a 500ms dependency backoff and proves explicit 403 recovery remounts media before loading the successful second mint.

ДАЛІ: align the existing unavailable assertion with current locale copy, format the player/policy/tests, and typecheck.

ФАКТ: bounded-retry regression now asserts the actual English locale string `This video is unavailable right now.`.

ДАЛІ: format the player, policy and tests.

ФАКТ: formatter passes on the player, application recovery policy and their tests.

ДАЛІ: run frontend typecheck after mint-cycle and explicit-retry changes.

ПЕРЕВІРКА: frontend typecheck passes after the player-level mint debounce and explicit-retry regressions.

ДАЛІ: correct B40's positive evidence to count only `.m4s` responses, then ask both independent reviewers to re-audit the resolved working set.

ФАКТ: B40's old-generation status list now includes only `.m4s` segment responses; its required 200 cannot be satisfied by a playlist or init file.

ДАЛІ: run second-round standards and specification reviews in parallel against current files and first-round findings.

ФАКТ: both independent reviewers are re-auditing the current files against their first-round findings; no edits or git access were delegated.

ДАЛІ: format the last B40 edit and run a local architecture/import scan while reviews complete.

ПЕРЕВІРКА: playback.spec formatting passes and the repository architecture checker passes across 89 files after policy extraction/composition injection.

ДАЛІ: receive the two second-round review results; fix any remaining blocking issue before final gates.

## Review round 2 — Standards

ФАКТ: standards reviewer PASS. It confirms recovery decisions are inward, concrete HLS wiring is in composition, media-frame color is tokenized, and API wire objects map explicitly to domain. It recommends leaving the small test-only overlap arithmetic duplication because extraction would add serialization indirection.

ФАКТ: specification reviewer had not returned within the first two-minute wait; no result is inferred from the timeout.

ДАЛІ: wait once more for the specification re-review while preserving the standards PASS.

## Review round 2 — Specification

ФАКТ: specification re-review confirms every first-round fix, then finds four remaining issues: native 410 captured time after native stopLoad reset it; the e2e range helper bypassed React's change tracker; O1 could label a zero→zero observation as successful cancellation; B24 did not keyboard-activate Pause.

ДАЛІ: capture 410 time before stopLoad, dispatch e2e seeks through the native input setter, print an explicit inconclusive O1 result when pre-count is zero, and add keyboard Pause proof.

ФАКТ: all four round-2 findings are patched: native currentTime is captured first; range e2e uses HTMLInputElement's prototype setter; O1 distinguishes yes/no/inconclusive; B24 activates the live Pause button by keyboard and observes pause().

ДАЛІ: inspect the exact patched blocks, then format/typecheck and request a final specification confirmation.

ФАКТ: patched blocks are present at the intended public seams: Pause remains the focused renamed button after play; native input setter precedes bubbled input/change; O1 reports an explicit inconclusive string for zero pre-count; 410 captures a local resumeAt before stopLoad.

ДАЛІ: brace the remint switch case for lexical-scope lint safety, then format/typecheck.

ФАКТ: remint case now has an explicit lexical block around captured resumeAt.

ДАЛІ: format the three round-2 fix files.

ФАКТ: formatter passes on the three round-2 fix files.

ДАЛІ: run frontend typecheck after round-2 fixes.

ПЕРЕВІРКА: frontend typecheck passes after native 410, seek dispatch, O1 labeling and keyboard Pause fixes.

ДАЛІ: request one final read-only specification confirmation on the four round-2 findings.

ФАКТ: final narrow spec confirmation is running read-only against the four patched findings.

ДАЛІ: while it runs, verify dependency inventory contains hls.js as the only T7 addition and inspect report tail for final-matrix corrections.

ФАКТ: component suite GREEN 4/4, including three state rows and B24 keyboard/quality/axe.

ДАЛІ: add RED composition/route tests for request-scoped learning query and exact `/learn/{slug}/{lessonId}` parsing; then wire the existing splat route without touching routeTree/generated files.

ФАКТ: RED query/route tests added. They pin exact two-segment slug+UUID parsing and one Query key with concurrent deduplication. No routeTree/generated edits are required because the existing `/learn/$` route remains the transport shell.

ДАЛІ: run only those two tests; expected RED is missing `learn-queries` and `learn-route` modules.

ФАКТ: query/route suite RED exactly as expected: two files, 0 tests executed, only missing `learn-queries` and `learn-route` modules.

ДАЛІ: implement parser/query factory, request-scoped `learn.ts`, root context wiring and the existing splat route’s lesson component.

ФАКТ: parser/query/composition/root/route wiring implemented. Existing `/learn/$` now validates exactly slug+UUID, preloads request-scoped learning query, verifies lesson membership, and renders `LessonPlayer`; `/learn` keeps its prior empty state. Presentation receives actions only and performs no fetch.

ДАЛІ: format route/composition files, run RED slice tests, then typecheck early for router-context drift.

ФАКТ: nine route/composition files format PASS.

ДАЛІ: rerun query/parser slice.

ФАКТ: query/parser slice GREEN 8/8.

ДАЛІ: run frontend typecheck early to expose router/media typing drift before e2e work.

ФАКТ: early typecheck RED, exit 2. Findings are mechanical and scoped: React 19 requires explicit initial values for five refs; assignments to undefined require union refs; existing route-guard test context needs the new request-scoped `learn` dependency.

ДАЛІ: fix ref declarations and add a test learn dependency to existing route-guard harness, then rerun typecheck.

ФАКТ: exact typecheck sites inspected: five zero-argument refs and existing route context literal at `route-guards.test.tsx:146`. No broader router drift found in compiler output.

ДАЛІ: patch those exact sites, using typed fake PlaybackRepository in route-guard harness.

ФАКТ: route-guard imports inspected; it composes repositories directly, so adding `createLearnQueries` + typed `PlaybackRepository` matches existing harness style.

ДАЛІ: apply ref and harness patch.

ФАКТ: React refs now explicitly include/initialize undefined; existing route-guard harness receives a typed fake playback repository through `createLearnQueries`.

ДАЛІ: format both files, then rerun typecheck.

ФАКТ: ref/harness format PASS.

ДАЛІ: rerun frontend typecheck.

ФАКТ: early frontend typecheck GREEN, exit 0.

ДАЛІ: inspect exact admin publish/upload APIs and auth fixtures needed by isolated playback e2e; then add runnable `playback.spec.ts`, O1 console measurements, project wiring and player layout checks in `accessibility.spec.ts`.

ФАКТ: auth fixtures and media upload helper pattern inspected. Playback e2e can use existing admin/student storage states, generated 5-minute fixture, admin UI upload, and direct admin API setup/cleanup without extra logins.

ДАЛІ: inspect exact status transition endpoint/body and course token updates before writing e2e fixture.

ФАКТ: admin e2e confirms UI transition flow and delete contract; fixture helpers parse only `{id,updatedAt}` from external JSON. Exact generated status operation still needs a narrow read.

ДАЛІ: read generated status URL/body and media upload path anchors.

ФАКТ: narrow read landed just before transition implementation; operation name is `transitionAdminCourse`, exact body type `AdminCourseStatusRequest`. Need final lines 960–1025 for URL/method.

ДАЛІ: read exact transition function.

ФАКТ: exact admin transition contract confirmed: `POST /api/v1/admin/courses/{id}/status` with `{status, expectedUpdatedAt}`, returning a fresh course token. E2E fixture contract is complete.

РІШЕННЯ: isolated project order will be `media → playback → recovery`. Playback uses the real 5-minute 720p fixture, creates/publishes its own course, and deletes only its run-scoped course via PostgreSQL in `afterAll` because public lifecycle has no published→draft/delete transition.

ДАЛІ: add runnable real-browser playback spec with P01/P08–P11/B39/B40/R04 and printed O1 measurements; then add project wiring and accessibility layout assertions.

ФАКТ: the large e2e patch tool output was truncated, so its application status is not assumed.

ДАЛІ: inspect playback.spec.ts and Playwright configuration directly before making further edits.

ФАКТ: playback.spec.ts exists (397 lines); Playwright excludes it from chromium and runs a serialized playback project after media and before recovery.

ДАЛІ: inspect the full playback spec and existing accessibility suite for type, sequencing, cleanup, and assertion gaps.

ФАКТ: full e2e inspection found three concrete gaps: response arrays use the wrong Playwright type, P11 begins waiting after playback already starts, and B24a/B25 are absent from accessibility.spec.ts. Cleanup also hard-codes database credentials.

ДАЛІ: inspect the player effect/callback graph and current stylesheet before fixing lifecycle restoration and adding the missing browser accessibility proof.

ФАКТ: player inspection exposed a runtime lifecycle bug: position state participates in the media error callback, so every timeupdate can recreate/destroy the HLS adapter. Restoration also clamps to zero when duration is not yet known at MANIFEST_PARSED.

ДАЛІ: locate the actual stylesheet and inspect existing component tests before applying a focused lifecycle fix with regression proof.

ФАКТ: player CSS is in frontend/src/app/styles/main.css and already provides semantic tokens, responsive stacking, visible focus, and a tokenized minimum slider target. Existing B24 coverage has a stable injectable media seam suitable for lifecycle regression tests.

ДАЛІ: add a failing regression test for unknown-duration restoration and adapter stability across timeupdate.

ФАКТ: added a focused component regression that requires restoring position when duration is not yet known and keeping one media adapter after timeupdate.

ДАЛІ: run the focused LessonPlayer test to confirm the regression is red before changing production code.

БЛОКЕР ПЕРЕВІРКИ: focused Rstest did not reach assertions because its server bind to ::1:3000 was denied with EPERM in this sandbox; this check is not recorded as red or passed.

ДАЛІ: inspect the Rstest configuration for a non-network execution path, without changing the requested implementation scope.

ФАКТ: Rstest config has no test-server override; port 3000 appears only in the application Rsbuild config. The failed bind is runner/sandbox infrastructure, not a test assertion or product regression.

ДАЛІ: fix the player by removing position state from the media-effect dependency path and avoiding a zero clamp before metadata provides duration.

ФАКТ: LessonPlayer now keeps playback position in a ref for error recovery, so timeupdate does not recreate the media adapter; MANIFEST_PARSED restoration uses the pending position until a positive duration is known.

ДАЛІ: repair playback.spec.ts response typing, start P11 segment observation before playback, and make database cleanup honor Compose credentials.

ФАКТ: playback e2e now uses Playwright Response, observes the archived-admin segment before starting playback, and cleans only its run-scoped course using the Compose service's POSTGRES_USER/POSTGRES_DB values.

ДАЛІ: add the required B24a/B25 browser accessibility test with intercepted learning data, student auth, dark/reduced-motion/200% conditions, target-size, focus, overflow, and overlap assertions.

ФАКТ: confirmed the strict learning payload and client-only lesson route; accessibility can be exercised without minting media by intercepting only GET /api/v1/learn/courses/{slug} under the stored student session.

ДАЛІ: implement B24a/B25 in frontend/e2e/accessibility.spec.ts.

ФАКТ: accessibility.spec.ts now covers the ready-player layout under student auth at mobile width, dark mode, reduced motion and 200% text; it checks keyboard order/focus ring, 44px targets, no overflow and no control overlap.

ДАЛІ: format the T7-owned frontend files, then typecheck to surface compile-time issues before reviewing behavior coverage.

ФАКТ: oxfmt completed successfully on the 24 T7-owned TypeScript, TSX, CSS and Playwright files.

ДАЛІ: run frontend TypeScript checking.

ПЕРЕВІРКА: frontend typecheck reached code and failed on playback.spec.ts DOM narrowing only (video currentTime/ended and accessor closure narrowing); no production TypeScript errors were reported.

ДАЛІ: add explicit HTMLVideoElement guards and stable accessor locals in the O1 instrumentation, then rerun typecheck.

ФАКТ: the failing locations are limited to five video-property reads and the nested getter/setter used by O1 instrumentation.

ДАЛІ: centralize guarded video reads and retain narrowed currentTime accessors outside the nested property descriptor callbacks.

ФАКТ: the combined patch did not apply because oxfmt had changed the multiline openAndPlay expression; no source file was changed by that failed patch.

ДАЛІ: apply the same fix in smaller context-stable edits.

ФАКТ: guarded videoTime/videoEnded helpers were added and openAndPlay now uses the typed time helper.

ДАЛІ: replace the remaining inline video reads and stabilize the O1 descriptor accessors.

ФАКТ: every playback e2e video property read now validates HTMLVideoElement, and O1 stores narrowed getter/setter functions before installing instrumentation.

ДАЛІ: format playback.spec.ts and rerun frontend typecheck.

ФАКТ: playback.spec.ts formatting completed successfully.

ДАЛІ: rerun frontend typecheck after DOM narrowing fixes.

ПЕРЕВІРКА: frontend typecheck passes after the e2e and player lifecycle fixes.

ДАЛІ: re-read the T7 acceptance/verification table and audit each required behavior against code or a named test before running final gates.

ФАКТ: brief re-audit confirms the implementation remains inside T7 scope; it also reveals R04 must tolerate the retained replacement being short enough that forcing a new segment is unreliable, while a 2-second heartbeat still proves the same 403 recovery row.

ДАЛІ: make R04 wait for either a heartbeat or media-prefix 403, then audit every §8 row against its proof.

ФАКТ: R04 now accepts the same terminal 403 from either the heartbeat contract or the media prefix, so it remains deterministic after B40 replaces the five-minute source with a short generation.

ДАЛІ: extract the exact §8 row wording and enumerate the code/test evidence for the mandatory report matrix.

ФАКТ: named proof exists for P01, P08–P11, B02–B04, B24/B24a, B39, B40, R04 and O1. The plan also places B25 under T11 while the T7 brief explicitly requires B24a/B25 here; the browser test intentionally follows the brief and covers both.

ДАЛІ: inspect the exact B25 acceptance row, then add the complete §8-to-proof matrix and reviewer commands to the report.

ФАКТ: popover is now observable and axe stage not yet reached, but happy-dom leaves focus on the trigger rather than the selected option after open, so ArrowDown did not reach listbox keyboard logic. DOM shows correct listbox/options/selection semantics.

ФАКТ: exact B25 requires both light/dark contrast and an intact 200% layout. The added test directly exercises dark/200%; axe unit provides automated contrast analysis, while existing application accessibility coverage already exercises the light theme.

ДАЛІ: inspect the report tail and component test state because a late B24 test note appeared after the prior green record; reconcile it before writing the acceptance matrix.

ДАЛІ: focus the selected `Auto` option after public listbox opens, then use ArrowDown+Enter and rerun.

ФАКТ: test now restores the browser focus transition that happy-dom omits, then exercises only keyboard selection on the public option.

ДАЛІ: format test and rerun component suite.

ФАКТ: test format PASS.

ДАЛІ: rerun component suite.

ФАКТ: B24 reached/focused quality but selected-level assertion ran before the React Aria popover was observable; callback stayed empty. This is test interaction ordering, not a production finding yet.

ДАЛІ: make the keyboard test wait for `listbox` after Enter, then move to 360p and commit.

ФАКТ: B24 test now waits for the public listbox before keyboard selection and for the observable media callback after commit.

ДАЛІ: format test and rerun component suite.

ФАКТ: B24 test format PASS.

ДАЛІ: rerun component suite.

ФАКТ: B24 now reaches the quality control, but exact-name lookup exposed duplicate accessible naming (`Auto Quality Quality`) caused by both `aria-label` and React Aria `Label`. This is a real a11y naming defect; 3/4 still pass.

ДАЛІ: remove redundant `aria-label`, assert the labeled control by `/Quality/`, then rerun.

ФАКТ: redundant quality `aria-label` removed; React Aria Label is now the single accessible name source. Test locates the resulting labeled button without depending on its selected-value prefix.

ДАЛІ: format component/test and rerun.

ФАКТ: component/test format PASS.

ДАЛІ: rerun component suite.

ФАКТ: component suite still 3/4. Explicit ArrowRight ran, but bounded to 1 because duration state remained 0: fake manifest readiness set media duration without emitting a later `durationchange`. Production should snapshot known duration at ready as well as listen for changes.

ДАЛІ: synchronize duration state in `onReady`, then rerun.

ФАКТ: `onReady` now snapshots finite media duration before restoring position; seek bounds no longer depend on a separate event ordering.

ДАЛІ: format and rerun component suite.

ФАКТ: format PASS after duration synchronization.

ДАЛІ: rerun component suite.

ФАКТ: dependency inventory confirms the sole T7 dependency is `hls.js@1.7.2`, pinned in package.json and package-lock.json.

## Review round 3 — Specification

ФАКТ: final independent specification review PASS. It confirms native 410 captures time before stopLoad, e2e seek reaches React through the native setter, O1 labels zero-precount evidence inconclusive, and B24 keyboard-activates Pause. No regression from those fixes was found.

ФАКТ: earlier report additions used matching context lines that recur in this append-only journal, so some later facts may appear before older repeated `ДАЛІ:` entries. Facts remain present; from this unique round-3 marker onward all additions target the physical EOF.

ДАЛІ: verify that the mandatory proof matrix/reviewer commands and all review outcomes remain present in the report, then run final permitted gates in required order.

ФАКТ: report contains the mandatory §8 proof matrix, exact four-command O1 reviewer block, isolated-project rationale, unrun Docker gate list, both Discussions, standards PASS and final specification PASS.

ФАКТ: proof-strengthening supersedes terse earlier matrix wording: P10 derives the exact EXTINF interval covering 80s; B40 requires an old-generation `.m4s` 200; R04 plays 31s then requires a segment 403; B25 runs browser contrast in both themes before dark/200% geometry.

ДАЛІ: run final permitted gate 1/4 exactly: `npm run format:check` in frontend.

ПЕРЕВІРКА: final gate 1/4 `npm run format:check` PASS; 157 files checked, exit 0.

ДАЛІ: run final permitted gate 2/4 exactly: `npm run lint` in frontend.

ПЕРЕВІРКА: final lint attempt reached code but failed on one `jsx-a11y/media-has-caption` warning for the new `<video>`; architecture checker did not run after oxlint exit 1. This is not passed.

ДАЛІ: inspect repository suppression conventions and playback contracts for captions; do not invent a fake caption source.

ФАКТ: no playback/spec/T7 caption or subtitle source contract exists, and the only repo convention uses a targeted `oxlint-disable-next-line` with rationale. Adding an empty `<track>` would falsely claim captions and produce no accessible content.

РІШЕННЯ: add a narrow lint suppression on the video element stating that T7 has no caption-source contract; this does not implement or claim captions and does not broaden scope.

ДАЛІ: add the targeted rationale comment, format the file, and rerun lint gate 2/4.

ФАКТ: LessonPlayer carries a single-line `media-has-caption` suppression with the missing caption-source rationale; no fake track or caption claim was added.

ДАЛІ: format LessonPlayer before repeating lint.

ФАКТ: LessonPlayer formatter PASS after the targeted lint rationale.

ДАЛІ: rerun final permitted gate 2/4 `npm run lint`.

ПЕРЕВІРКА: final gate 2/4 `npm run lint` PASS; oxlint has zero warnings and architecture checker passes 89 files.

ДАЛІ: run final permitted gate 3/4 exactly: `npm run typecheck` in frontend.

ПЕРЕВІРКА: final gate 3/4 `npm run typecheck` PASS, exit 0.

ДАЛІ: run final permitted gate 4/4 exactly: `npm test` in frontend; record the sandbox bind refusal if it recurs.

ПЕРЕВІРКА: final gate 4/4 executed 364 tests: 359 passed, 5 failed. One adapter expectation still compared the wire timestamps intentionally removed by explicit domain mapping. Four route-guard cases show `/learn` is being matched by `/learn/$` and rendered not-found, regressing the existing learning empty state. This gate is not passed.

ДАЛІ: inspect the adapter fixture and learn route-tree shape; fix only the mapping expectation and `/learn` collision, then restart the four final gates in order.

ФАКТ: adapter failure is an outdated test expectation for three adapter-only timestamps. Route failure is caused by child `/learn/$` beforeLoad rejecting an empty splat before the parent can render its preserved `/learn` empty state; generated tree confirms the child is nested under `/learn`.

РІШЕННЯ: accept only an empty splat as the parent index case in child beforeLoad, returning no child UI; all nonempty malformed paths remain not-found. This fixes routing without adding an index file or touching generated routeTree.

ДАЛІ: update the domain expectation and empty-splat branch, then run the focused failing files.

ФАКТ: adapter test now expects the explicit domain shape; empty splat no longer throws not-found. The first route patch introduces an early return before hooks, so it needs a child component split before lint/testing.

ДАЛІ: split the nullable route shell from the hook-using lesson body, preserving unconditional hooks per component.

ФАКТ: route body has four hooks after the nullable context check; a small shell/body split can preserve behavior without new domain or route abstractions.

ДАЛІ: implement that shell/body split.

ФАКТ: nullable splat shell now returns before rendering a separate LessonRoutePlayer; all query/memo hooks are unconditional inside the body component.

ДАЛІ: format the two test-failure fixes, then rerun only playback-repository and route-guards suites.

ФАКТ: formatter passes on the domain-mapping expectation and empty-splat route fix.

ДАЛІ: rerun the two previously failing test files together.

БЛОКЕР ПЕРЕВІРКИ: targeted rerun did not start because Rstest again hit sandbox `listen EPERM ::1:3000`; no focused assertion result is claimed.

ДАЛІ: restart final permitted gates from 1/4 after the two fixes: format-check, lint, typecheck, then full npm test.

ПЕРЕВІРКА: restarted final gate 1/4 format-check PASS; 157 files, exit 0.

ДАЛІ: run restarted final gate 2/4 lint.

ПЕРЕВІРКА: restarted final gate 2/4 lint PASS; zero warnings and architecture checker passes 89 files.

ДАЛІ: run restarted final gate 3/4 typecheck.

ПЕРЕВІРКА: restarted final gate 3/4 typecheck PASS, exit 0.

ДАЛІ: run restarted final gate 4/4 full `npm test`.

ПЕРЕВІРКА: full suite improved to 363/364 pass. Adapter mapping and all three `/learn` index-role cases pass. Remaining G06 expected the old parent empty state at a signed-in nested `/learn/lessons/intro`; T7 now intentionally treats malformed/non-contract lesson paths as not-found. Gate remains red pending test-contract inspection.

ДАЛІ: read G06's stated matrix and surrounding route guard intent; decide whether the assertion is stale under T7's exact slug+UUID route or reveals a real guard regression.

ФАКТ: G06 is a durable completed identity contract (`docs/identity/i06/acceptance.md`): signed-in `/learn/go/intro` must open the learning shell, while anonymous users redirect. T7's route choice must preserve that fallback. The remaining failure is a real regression, not a stale test.

РІШЕННЯ: valid slug+UUID paths load the lesson; invalid/non-contract nested paths retain the existing authenticated learning empty state. Valid paths whose course/lesson is absent still return not-found.

ДАЛІ: share the learning empty-state component with the splat fallback and rerun G06/full gates.

ФАКТ: learn parent now exports its existing empty-state view; invalid splat paths render it after the parent identity guard, while valid lesson paths continue through learning preload and lesson membership checks.

ДАЛІ: format both learn routes, then run the single G06 repro.

ФАКТ: both learn route files format successfully.

ДАЛІ: run the exact remaining G06 repro command from Rstest.

БЛОКЕР ПЕРЕВІРКИ: exact G06 repro again could not start due intermittent sandbox `listen EPERM ::1:3000`; no assertion outcome claimed.

ДАЛІ: perform the final four-gate sequence once more after the G06 compatibility fix.

ПЕРЕВІРКА: final sequence gate 1/4 format-check PASS; 157 files, exit 0.

ДАЛІ: final sequence gate 2/4 lint.

ПЕРЕВІРКА: lint found one mechanical unused `t` left in LearnRoute after extracting LearnEmptyState; exit 1, not passed.

ДАЛІ: remove the unused parent translation binding, format, and restart the required sequence.

ФАКТ: removed only the unused parent `t`; LearnEmptyState retains its own translation hook.

ДАЛІ: format learn.tsx, then restart final gates at format-check.

ФАКТ: learn.tsx formatter PASS.

ДАЛІ: final clean sequence gate 1/4 format-check.

ПЕРЕВІРКА: final clean gate 1/4 format-check PASS; 157 files, exit 0.

ДАЛІ: final clean sequence gate 2/4 lint.

ПЕРЕВІРКА: final clean gate 2/4 lint PASS; zero warnings and architecture checker passes 89 files.

ДАЛІ: final clean sequence gate 3/4 typecheck.

ПЕРЕВІРКА: final clean gate 3/4 typecheck PASS, exit 0.

ДАЛІ: final clean sequence gate 4/4 full npm test.

ПЕРЕВІРКА: final clean gate 4/4 `npm test` PASS: 43 files, 364/364 tests, 0 skipped/todo, exit 0.

ФАКТ: permitted final sequence is fully green in required order: format-check → lint/architecture → typecheck → unit tests. Independent standards and specification reviews are PASS after fixes.

НЕ ЗАПУЩЕНО: per brief, Docker-dependent `make integration`, `make build`, `make up`, `make e2e`, `make check-generated` and real-browser O1 measurements. They are not represented as passed. Native Safari HTTP-status parity remains the documented Discussion limitation; P08 Compose profile wiring remains documented with exact reviewer setup.

ДАЛІ: reviewer runs the exact O1 command block above, records its printed counts/currentTime/stopLoad observation (including `inconclusive` if no request was active), then runs the Docker gates before accepting T7 and moving to T8.

---

## Рецензія T7 (сесія-приймач, 2026-09-08)

ФАКТ: агента перепідвʼязано до сесії, що веде фазу; Docker-гейти, недоступні йому,
прогнано тут.

ФАКТ: `make test` дав 364/364, вихід 0. Чотири падіння `CourseEditorView.test.tsx` у
першому прогоні (три таймаути по 5000мс і один незнайдений текст) не відтворились ані
ізольовано, ані в повторі повної свити: перший прогін ділив машину з контейнерами
інтеграції. Це флак під навантаженням, не дефект; при 364 тестах дефолтні 5с стають
тісними на завантаженій машині.

ФАКТ: `make check-generated`, `make build`, `make format-check`, `make lint`,
`make typecheck` — усі вихід 0.

ФАКТ: `make e2e` — 42/42, вихід 0, після трьох виправлень нижче.

### Три дефекти, знайдені першим справжнім прогоном у браузері

ФАКТ (F-1, продуктовий): `master.m3u8` віддавав 403 при `auth_request` = 204. Дерево
транскодера має `0750`/`0640` на uid 10001, nginx працює під uid 101 і не міг відкрити
каталог відео. Теза «nginx віддає медіабайти» не працювала з фази 5 — її тести не тягнули
байтів через edge. Виправлено `group_add: ['10001']` для nginx; права не послаблено.
Рішення записано як F-1 у `docs/playback/plan.md §12`.

ФАКТ (F-2, тестовий): `beforeAll` у `playback.spec.ts` транскодує пʼятихвилинний фікстур,
але `describe.configure({ timeout })` накриває тести, а не хук — той лишався на 30с.
Додано `test.setTimeout(420_000)` у самому хуку.

ФАКТ (F-3, тестовий): очищення в `afterAll` передавало `:'course_id'` через `psql -c`,
де підстановки змінних не відбувається — сервер отримував літерал і відповідав `42601`.
Оператор переведено на stdin. Перевірено окремим прогоном проти живої бази.

ФАКТ (F-4, тестовий): генерація фікстур мала бюджет 120с, тоді як сам пʼятихвилинний
енкод займає 111с на завантаженій машині. Бюджет піднято до 600с на тест і 300с на виклик.
Фікстур не скорочено: його довжина потрібна для seek у незавантажену частину.

### O1

ФАКТ: заміряно під `PLAYBACK_HEARTBEAT_SECONDS=2`, `PLAYBACK_OWNERSHIP_TTL_SECONDS=6`.
Позиція відновлюється точно (0 → 0.254666 → 0.254666). Ключове питання — чи `stopLoad`
знімає запити в польоті — лишилось **неостаточним**: у момент перекарбування щоразу було
нуль запитів у польоті. Числа й наслідок записано в `docs/playback/plan.md §12`.

### Discussion агента

ФАКТ: `## Discussion` про непрозорість статусу в нативному HLS переведено з цього звіту в
рішення **D-9** плану. Нерівність гілок названо свідомою, а не паритетом.

ВЕРДИКТ: T7 прийнято, позначено `done` у `work/tasks.md`. Наступний — T8.

---

## Верифікація рецензентом (2026-09-08)

### Що виміряно зеленим — і в якому прогоні

| Гейт | Результат | Прогін |
|---|---|---|
| `format-check`, `lint`, `typecheck` | exit 0 | усі |
| `test` (фронт) | **364/364** | 9 |
| `integration` | 25–26 із 26 пакетів `ok`; падіння лише на підйомі контейнерів | 5–8 |
| `check-generated`, `build`, `up`, `seed`, `smoke` | exit 0 | 6–10 |
| `e2e` | **35/36**; єдине падіння — хибнопозитив у хелпері, виправлено | 8 |

### Чого НЕ досягнуто, і це не приховується

**Одного повного зеленого ланцюга за один прогін на цій машині отримати не вдалося.**
Одинадцять спроб; щоразу падав інший елемент, і жодного разу — код T7.

Причина остання й відтворювана: важкий a11y-тест
`CourseEditorView B09 keeps a maximal module list keyboard reachable and free of axe violations`
не встигає у ліміт **5000 мс** при `load average` 43–52, який на цій машині створюють шість
сторонніх проєктів. Доказ, що це **повільність, а не зависання**: той самий файл дає
**12/12 PASS** із `--testTimeout=30000`, і в прогоні 9 при нижчому навантаженні — **364/364**.

**Ліміт у конфізі свідомо не піднято.** 5 с — вибір проєкту, і він тримається на
розвантаженій машині; піднімати його назавжди заради тимчасового чужого навантаження було б
підгонкою мірника під обставини.

### Хронологія одинадцяти прогонів — жоден не вказав на код T7

| # | Впало | Причина |
|---|---|---|
| 1 | `build` | лок згенеровано npm 11.6.2, перевіряє 11.19.0 |
| 2 | `smoke` | воркер відновив стару задачу → метрика не нульова |
| 3 | `test` | load 43 → таймаути |
| 4 | — | **моя команда**: не та робоча тека |
| 5–6 | `integration` | гонка ryuk на перевантаженому демоні |
| 7 | `e2e` | стан від моїх часткових прогонів (`delete-me`, лічильник логінів) |
| 8 | `e2e` B40 | **хибнопозитив у хелпері** — єдине, що стосувалось T7 |
| 9 | `e2e` | **моя правка**: тест-файл імпортував тест-файл |
| 10–11 | `test` | load 51 → один таймаут |

### Дві правки рецензента

**1. Хибнопозитив у `uploadReady`.** Хелпер чекав на атрибут **уроку**
`data-video-state='ready'`, а урок уже мав готову генерацію — тож умова ставала істинною
**миттєво**, до того як нове відео дотранскодувалось. Тепер опитує стан **свого** `videoId`
через `/api/v1/admin/courses/{id}/videos`.

Це **четвертий хибнопозитив фази** і єдиний, що дав **червоне** — тому й знайшовся.

**Код при цьому був правильний.** Обидва запити роблять
`state='ready' AND evicted_at IS NULL ORDER BY generation DESC`, тобто беруть найновішу
**готову** генерацію; поки нова транскодується, чесно віддається стара — рівно те, заради
чого Phase 5 тримає дві генерації. Полагодити код під тест означало б зламати цю властивість.

**2. Профіль P08 переїхав у шар e2e.** Проєкт `playback` отримав власний setup (api з
каденцією 2/6) і teardown (повернення 30/180). `docker-compose.yml` **не змінюється** —
розробка й решта e2e бачать продакшн-значення, а `make check` лишається однією командою.
Спільна операція живе в `e2e/playback-profile.ts`, бо Playwright забороняє тест-файлам
імпортувати одне одного.

### Discussion — обидві прийнято

**D1 — нативний HLS не віддає HTTP-статус сегмента.** `<video src="…m3u8">` у Safari дає
JavaScript лише загальну media-помилку, тож розрізнити 403/404/503 у нативній гілці
**неможливо фронтовим кодом**. D-7 писався так, ніби застосовний до обох гілок — це було
неправдою, і план виправлено: нативна гілка бере політику з **точних кодів heartbeat**, а
непрозору помилку трактує обмеженим ретраєм. Записано в §9 як обмеження, не як паритет.

**D2 — P08 недосяжний із продакшн-каденцією.** Вирішено профілем у шарі e2e (вище).

### O1 — лишається відкритим

Заміри hls.js (запити в польоті, доля `currentTime`) написані й запускабельні, але потребують
прогону браузера, якого стабільного на цій машині зараз не отримати. **Не заявляється
закритим.**
