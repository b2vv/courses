# Агенда — що робити далі

| | |
|---|---|
| Дата | 2026-11-19; повний rebuild після агенди 2026-10-31 |
| Спирається на брифінг | [work/CTO-RESEARCH.md](CTO-RESEARCH.md), re-sync **`7b78727f`** над продуктовою базою **`06c1c220`** (#111); §1–§6 і первинні звіти звірено читанням. |
| HEAD на момент агенди | `main` @ `7b78727f42386f2210cc3d781d444be1c02c4b03`; `gh issue list --state open --limit 100 --json number,title,url` → `[]` (**0 issues**). PR inventory цим запитом не вимірювався. |
| Чи протух попередній вибір | Так. «Черга порожня / винести деплой» вже не описує стан: є датований Start/Rsbuild request (`cbb645a0`) і виміряні R5/MS4 хвости. Роадмап §24 завершено, нової відкритої фази немає (`CTO-RESEARCH.md` §1–2). Деплой лишається відкладеним людським рішенням. |

**Зроблено з попередньої агенди:**

- **done: seamless quality switch через `nextLevel` → PR #96 / `69d51a8a`.**
- **done: multi-audio S1–S3 → PR #97 / `ca6634bd`; worker correctness → #98 / `640b6e2b`; authoring contract → #99 / `57d3832d`; upload/enqueue/sweep → #100 / `0b04f5af`; fix-forward → #101 / `7c6b754c`.**
- **done: audio data layer → #102 / `15afa6c9`; UI → #103 / `22928697`; embedded backfill → #104 / `ef48c150`; ASR source → #105 / `3f0e1c84`; player audio choice → #106 / `5cbb44bb`.**
- **done: multi-subtitles MS1 → #107 / `6ad8d8d3`; F3 phone/200% editor → #108 / `37ec4b0a`; MS2+MS3 → #109 / `f50f6ec3`; MS4 audible captions/panel → #110 / `10436cdf`.**
- **done: admin-editor-ux Ux01–Ux18 → #111 / `06c1c220`; briefing re-sync → `7b78727f`.**

Джерела closing: `git log --first-parent 5fbaaff..7b78727f`, `docs/multi-audio/report.md` §S1–S7,
`docs/multi-subtitles/report.md` §MS1/F3/MS2–MS4, `docs/admin-editor-ux/report.md` §1–3.
Історичні deploy-readiness (`5a671d5`), LB09a (#91/#92) і teach-editor (#93–#95) були закриті ще в
попередній агенді; вони не нові кандидати. Останній записаний фазовий gate — admin-editor-ux на
`8f7923ad` (у squash #111): 168 e2e, GATE_EXIT=0 (`docs/admin-editor-ux/report.md` §2).
Цей rebuild не запускав builds/tests/Docker і не заявляє нового runtime/CI PASS.

## Джерела та кількість кандидатів (усі 6 пройдено)

К-сть — окремі ходи/розвилки з реєстру нижче, які постачає джерело, **з перетинами** між джерелами.
Це не сума задач. Унікальних: **24** (M1–M18, D1–D2, Р1–Р4); дві продуктові розвилки сплячі.
Паралельна design-pass сесія — окремий стан роботи, не новий пакет задач цієї агенди.

| Джерело | К-сть | Суть / фільтр |
|---|---:|---|
| 1. Відкриті таски `work/` + gh | **3** | D1/S01, D2/S02 (`work/tasks.md:38–39`) і M1 — датований request у старій AGENDA (`cbb645a0`, блок збережено нижче). Сирих open-status rows у tasks **7**; B01/A10/A13/SA13/TE4 відфільтровані closing evidence нижче. gh **0 open issues**. |
| 2. Зафіксований борг / limitations | **17** | M2–M15, Р1/Р2/Р4. Прочитано `docs/multi-audio/report.md` limitations S3/S5/backfill/S6/S7; `docs/multi-subtitles/report.md` MS1/F3/MS3b/MS4; `docs/admin-editor-ux/report.md` §5 + plan §5 R5; playback §6; ASR §7; `work/DEPLOY-READINESS.md`; `work/backlog.md`. Додаткових tech-debt/CRITIQUE файлів пошук не знайшов. Lesson-document locale — `docs/lesson-builder/report.md:694`. |
| 3. Ризики брифінгу §6 | **18** | M2–M7, M9–M12, M14–M18, Р1/Р2/Р4. Нові щодо боргу — лише M16–M18: три непідтверджені механізми; інше — перетини. Джерело: `CTO-RESEARCH.md:192–276`. |
| 4. Unchecked boxes у live specs/plans | **4** | У `work/`/`docs/` spec/plan/REQUIREMENTS поза archive та WIP — **0 `[ ]`**. Є **15** unchecked пунктів A1–E2 readiness checklist (`DEPLOY-READINESS.md:16–42`) → D1/D2/Р1/Р4, усі вже враховані. Архівні boxes не повертають виконану роботу. |
| 5. Skipped tests / disabled gates / TODO у коді | **0** | Точний `git grep -n -E "test\.skip|t\.Skip\(|\.only\(|TODO|FIXME" -- backend frontend/src frontend/e2e` → **1** `t.Skip("no build info")`, `backend/internal/testsupport/containers/guard_test.go:17`; 0 інших anchors. CI tag/manual e2e та `E2E_SKIP_ASR=1` — явна межа конфігурації через відсутній asr-worker (`.github/workflows/ci.yml:78,114–115`, `frontend/playwright.config.ts:58`), не новий дефект і не повне CI ASR-покриття. |
| 6. Невирішені продуктові розвилки | **4** | Р1 deploy, Р4 rotation policy до deploy; Р2 lesson-document locale та Р3 мультивалюта — майбутні scope-вибори, не оголошені «відкритими питаннями власника». Джерела й опції нижче. |

Кандидатів з міткою `⚠️ no source`: **0**.

**Closing-фільтр, не нова робота:** B01 → B-1 PR #62 / `0b69d04` (`backlog.md:127–143`);
A10/A13 → accounts закрито (`docs/accounts/report.md:4,21,44–49`, PR #7 / `d263045`);
SA13 → #83–#90 + `6efa393` (closing у самому `tasks.md:466`);
TE4 → frontend PR #94 / `5b6b47b`, TE6 #95 / `5fbaaff` (`docs/teach-editor/report.md` §TE4–TE6,
`CTO-RESEARCH.md` §1). Ці stale statuses не редагуються в цьому scope.
B-2–B-26/B-28 з DONE/Outcome/Closure не поновлюються: зокрема B-20 `e8c238d`, B-21 #59,
B-18 #32, B-19 #61, B-22 `def0db3e` — deliberate limitation (`backlog.md` відповідні sections).
Архів використовувався лише для перевірки closing-посилань, не як джерело кандидатів.

## Кандидати по двох осях (розблоковує / зворотність)

Перші три — найближчі альтернативи; решта — зафіксовані follow-up/investigation, не обіцянка виконання.

| Хід | Джерело / стан | Розблоковує | Зворотність / ціна помилки |
|---|---|---|---|
| **M2. R5: прибрати published-fixture leak у lesson-builder/refund/teach-editor e2e** | `docs/admin-editor-ux/plan.md:148–151`, report:93–95; producers `frontend/e2e/{lesson-builder,refund,teach-editor}.spec.ts` | Повторні прогони без витіснення seed з catalog page 1; надійніший повний e2e для наступних змін. Виміряно 28/31 published rows після повторів. | Дешевий edit тестів; cleanup лише власних fixture IDs, без загального стирання. Shared-state зміна потім потребує повного e2e. |
| **M1. Start + Rsbuild compatibility-check** | Owner request **2026-11-14**, `cbb645a0:work/AGENDA.md`, дослівно нижче; prerequisite MS4 виконано #110 | Доказ рішення перед оновленням: peers/Node Linux-amd64, SPA, streaming, gzip окремо. | Дешеве ізольоване дослідження; Docker/emulation + gate коштують вікна й місткості. Оновлення автоматично не дозволене. |
| **M3. Normal transcript +5 px при 390/200%** | `docs/multi-subtitles/report.md:1008,1023` (395 проти 390 px) | Читання header/search/list без горизонтального виходу; локальний UX-фікс для learn. | Дешевий CSS/layout edit; повторити той самий viewport/zoom і перевірити спільні стилі. Наявний фазовий PASS цього не закрив. |
| M4. Періодичне recovery загубленого upload enqueue | `docs/multi-audio/report.md:190–196` | Відновлення video/audio без рестарту worker; нині Recover тільки на старті та після 10-хв grace. | Середня: повторні queue effects потребують ідемпотентності й доказу; код відкотний, уже відправлені jobs — ні. Не зараз без такого сценарію. |
| M5. Справжній sweep/PATCH race test | `docs/multi-audio/report.md:204–210` | Доказ row-lock при PATCH→Queued проти sweep; нинішній тест не ловить видалення FOR UPDATE. | Дешевий test-only edit; детермінований two-transaction proof, без підняття timeout. |
| M6. `awaiting_audio` у adminstats | `docs/multi-audio/report.md:239–240`, `CTO-RESEARCH.md:204` | Правдивий media-state лічильник: стан нині не входить ні в ready, ні в in_progress. | Невеликий SQL/DTO/UI edit із перевіркою семантики метрики; без міграції. |
| M7. Видима причина failed audio / заміни | `docs/multi-audio/report.md:275–278` | Автор розуміє падіння доріжки/заміни за наявності старшого Ready; DTO не несе причину. | Середня: контракт+UI та перевірка безпечних повідомлень. Продуктову семантику визначити до розширення DTO. |
| M8. Recovery version id замість лише `capturedAt` | `docs/multi-subtitles/report.md:541–544` | Збереження новішої чернетки іншого вікна за колізії в одну ms; відтворено лише з примусово однаковим clock. | Середня: persisted IndexedDB формат/atomic comparison; зворотність дорожча за звичайний UI edit. Не зараз без повторення реального випадку (trigger звіту). |
| M9. Kiri capacity перед heavy-вікном | `docs/admin-editor-ux/report.md:96` + §2 | Місце для compatibility/gates: 63 GB, 74–99% used, cache +≈17 GB за кілька builds; attempt 1 ENOSPC. | Read-only оцінка дешева; prune видаляє rebuildable cache і коштує повторної збірки. Фазовий дозвіл на prune не переноситься сюди; live health не перевірявся. |
| M10. Real Safari native playback/captions proof | `docs/playback/report.md:297–302`, `docs/asr/report.md:166`, `docs/multi-subtitles/report.md:1021–1022` | Відповідь про native branch; Chromium/Playwright WebKit ішли hls.js, fake AudioTrackList не доводить Safari. | Дешевий за кодом експеримент, потрібен реальний Safari/device; можливий фікс визначається лише результатом. |
| M11. Межі playback proof: ABR settling / 30–180 cadence / AV1 wire | `docs/playback/report.md:312–320,344–348` | Доказ утримання rung, shipped heartbeat/TTL та наскрізного telemetry field; нині step, 2/6 cadence, часткові тести. | Відкотні fixture/test edits, дорожчі тривалі прогони. Не заявляти ці властивості до proof; phase 6 102/102 не переозначується. |
| M12. ASR deadline/startup failure-path proof | `docs/asr/report.md:155–165` | Queue-derived timeout ≥300 s замість controlled 500 ms; startup/end-to-end і named-failure шлях понад часткові докази. | Відкотні тести, дорожчий runtime; звіт прямо фіксує відсутність чесної named-failure fixture. Спершу знайти механізм, не підміняти fake доказом. |
| M13. OpenTelemetry end-to-end spans verify | `work/backlog.md:143` (partial, B-1 в цілому closed) | Відповідь, чи пов'язуються SSR→API→worker spans; явного споживача наступної фічі не зафіксовано. | Дешеве дослідження; додавання spans — окремий кодовий scope лише після результату. Не зараз. |
| M14. B-27: перевірка release-trigger й видалення temporary retry | `work/backlog.md:7–33`; pin `backend/go.mod:17` | Прибирає wrapper/call sites, **якщо** обрана release містить upstream #3868; MS3 conflict retry не закрив B-27 (`multi-subtitles/report.md` §Gate flakes fix). | Dependency upgrade середньої ціни; removal + три integration прогони за записаним контрактом. Поточний upstream release status не перевірявся; не називати це доведеним upstream-блоком. |
| M15. Visual proof gaps: translucent contrast / Firefox text zoom | `docs/admin-editor-ux/report.md:87–92`, `docs/multi-subtitles/report.md:248–258,1007,1010` | Виміряний composited contrast destructive confirm/active line та text-only zoom понад CSS simulation; none-row ≈10 px — minor. | Read-only proof дешевий за кодом, потрібні відповідні засоби/браузер. NOT_TESTED не означає FAIL; не вигадувати ratio. |
| M16. R23 heartbeat 403 — встановити механізм при повторі | `work/CTO-RESEARCH.md:274` / §6 (K4 logs, closing pointer на `work/archive/review-o05-grant-ui.md`) | Діагноз access vs sameOrigin, якщо 403 повториться; MS3 B18 fix — інша причина. | Дешева інструментація/читання збереженого логування, без відтворення не доведено. Не зараз без повтору. |
| M17. Identity integration одноразове падіння — діагноз при повторі | `work/CTO-RESEARCH.md:275`; `infrastructure/tests/integration.sh:245` | Відновлення точного failure text через K5; ізольований контейнер вже відкинув shared-DB гіпотезу. | Дешеве читання логів; без відтворення причина невідома. Не зараз. |
| M18. NAT-профіль `limit_req` — endurance/load verify | `work/CTO-RESEARCH.md:276`; `docs/readiness/report.md` §7; `infrastructure/nginx/nginx.conf` zone=media | Доказ capacity під реальним NAT, якого сам конфіг не дає; конкретний launch traffic ще не задано. | Відкотний load experiment, потребує погодженого стенда й профілю; перед exposure, не зараз. |
| D1. S01 TLS-залежні headers/cookies | `work/tasks.md:38`, `DEPLOY-READINESS.md:16–17,22` | HSTS/Permissions-Policy/Secure на prod HTTPS. CSP вже enforcing; «зробити CSP» вдруге не задача. | Конфіг відкотний, невдалий транспортний cutover впливає на доступність; **блок Р1/A1**. |
| D2. S02 `__Host-` session-cookie | `work/tasks.md:39`, `DEPLOY-READINESS.md:18` | Прибирає duplicate-cookie передумову; потрібні Secure/Path=/, без Domain. | Дорожче: перейменування чинної cookie/сесійний cutover; **блок D1/Р1**, не зараз. |

Інші named boundaries зберігаються як межі доставлених фаз, а не нові обіцянки: backfill не
реконструює втрачений embedded original (multi-audio §S-backfill D-B6); manifest reload/немиттєве
перемикання (S7); metadata→ручний Regenerate, und без captions, orphan не показується учневі,
1.5 s settle і ASR ≈19 хв/год/доріжку (multi-subtitles plan §6, report MS4); D7 reload замість
Discard, D6 raw-ms через IDB v1, duration відсутня в MediaVideo, слабший observer-unit bound
(admin-editor-ux §5); image Put→Insert crash-window та migrated-description proof не e2e
(lesson-builder report:711–720). Історичний fast-uri advisory у тому звіті не є новим dependency audit.
Adminstats placeholders і discussion pagination/rate межі лишаються у phase reports
(`CTO-RESEARCH.md` §6). Transcript 87 rows не перераховано, 5000-line load не виміряно
(`docs/transcript/report.md` §6–7, briefing:273). Жодна з цих меж не оголошується закритою цим документом.

## Рішення людини (не ходи агента)

**Р1 — деплой лишається deferred.** Точний чинний запис `AGENTS.md:5`:
> Production deployment was deferred by the owner on 2026-10-03 (test mode, no prod server) and is a human decision, not a queued phase (`work/AGENDA.md` Р1).

| Розвилка / джерело | Опції та ціна | Рекомендація з причиною | Розблоковує / зворотність |
|---|---|---|---|
| **Р1. Коли відкривати production?** `AGENTS.md:5`; `DEPLOY-READINESS.md:7–8,16–52` | Зберегти test-mode: немає prod-витрат, немає публічного запуску. Відкрити launch: сервер/TLS, секрети, зовнішній email, monitoring, shop/key/ФОП/callback і live Approved proof; readiness A1–E2 стає gate. | Зберегти standing defer; повернутися до вибору, коли є намір і середовище запуску. Відсутність відкритої фази сама не є owner-go. | D1/D2 і всі prod tails. Дорогий зовнішній ефект/гроші/дані; не самостійний хід агента. |
| **Р4. Політика ротації 2FA master до deploy.** `DEPLOY-READINESS.md:29–31` | Наявний формат без key-id: rotation = re-enrol усіх 2FA, ціна — комунікація/відновлення. Версійований key-id/keyring: ціна — новий формат, міграція й перевірка читання/ротації. Опційний AAD=user_id — окреме hardening C3, не заміна rotation policy. | До першого launch описати re-enrol процедуру чинного формату; keyring брати лише за вимоги безперервної ротації. Так обсяг відповідає записаному механізму, без непогодженої міграції. | Закриває C2 deploy gate; рішення на папері дешеве, ротація ключа й persisted формат дорогі. До Р1 не запускати. |
| **Р2. Друга locale lesson-document UI.** `docs/lesson-builder/report.md:694–697` | Одна locale: без коду. Додати switch/create/read для другої: UI + persisted locale rows, правила fallback й recovery та перевірки. | Не зараз: спочатку конкретний курс/контент і вимоги до locale. MS1–MS4 вже закрили багатомовний transcript, але не lesson-document. Це майбутній scope-вибір, не засвідчене відкрите питання власника. | Мовні документи; код відкотний, створені тексти/locale semantics дорожчі. |
| **Р3. Мультивалюта.** `docs/payment/spec.md:35,39,97` | UAH: чинний контракт. Друга валюта: дані/валідація, підтримка провайдера, checkout/refund/display і фінансові докази. | Не зараз: spec:97 прямо лишає розвилку за попитом; нового запиту в зібраних джерелах немає. | Продаж за другою валютою; фінансові записи/зовнішні платежі дорогі для відкату. |

**Revoke не питання.** `AGENTS.md:8` прямо фіксує: “Revoke stays admin-only by the owner's decision
(2026-10-13): a lecturer acts through block (communication) + report-on-learner to the admin queue,
and the admin decides on revoke”. Це чинна межа (`docs/enrollment/report.md` §4), до розвилок не включена.

## Owner-initiated, у черзі (додано 2026-11-14)

Після фази multi-subtitles (`docs/multi-subtitles/plan.md`), не замість неї.

1. **Перевірка зв'язки TanStack Start + Rsbuild перед оновленням** (власник, 2026-11-14). Зараз:
   `@tanstack/react-start` 1.169.3, `@tanstack/react-router` 1.170.45, `@rsbuild/core` 2.2.2
   (`frontend/package.json`), образ `node:24-bookworm-slim` (`frontend/Dockerfile`).
   - **Окрема тестова гілка**, три перевірки окремо, кожна зі своїм доказом:
     1. **SPA-навігація** — клієнтські переходи між маршрутами без повного перезавантаження.
     2. **SSR streaming** — сторінки йдуть потоком через nginx з `proxy_buffering off`
        (`infrastructure/nginx/nginx.conf:241-243`), а не одним буфером.
     3. **Компресія** — gzip лише там, де він налаштований (`nginx.conf:250-252`, `gzip_proxied any`,
        `gzip_vary on`), і він не буферизує SSR-потік.
   - **Не оновлювати автоматично за номером версії.** Спершу перевірити сумісність залежностей
     (peer-залежності Start ↔ Router ↔ Rsbuild) і Node **в Linux/x86 Docker** (`--platform linux/amd64`;
     хост розробки — arm64, тож локальний зелений прогін цього не доводить). Лише після цього —
     оновлення й повний ordered `make check`.

M1 збережено дослівно з `cbb645a0:work/AGENDA.md` (дата/формулювання/джерела вище); поточна
агенда не переобирає стек і не починає оновлення. Після MS4 #110 його sequencing prerequisite виконано.

**Паралельна сесія — стан, без перенесення плану.** Read-only `git show
claude/design-pass-all-screens-085f0e:docs/design-pass/{plan,findings}.md`, tip **`5073e800`**:
5 невлитих комітів. Plan:9 ще каже **S0 in_progress**, Q3/Q4 open; §9:113 містить точну owner-відповідь
«Так, погоджую» про S0/PR cadence, §10:117 — «Спершу auth-зріз». AUTH fix **`5518cda0`** + evidence
**`ef3aea3f`** позначають A1-1/A2-1/A5-1 закритими **на тій гілці** (`findings.md` §Auth).
Це не merge/review/gate PASS у main. Розблоковує доступні auth journeys; presentation edit відкотний,
але delivery належить тій сесії, дублювати його тут не слід.
S0 знімав образ невідомої ревізії між MS1/F3 (`plan.md:114`), отже findings — кандидати на повторну
перевірку, не доведені дефекти main; D0/D2 вже закриті #111, D3/T2 потребують звірки після F3.
Q3 heavy-вікно/Q4 reviewer лишаються координацією тієї сесії; тут не обираються. Її план не прийнято,
не злито й не перетворено на новий backlog. Джерела зберігаються на branch tip, checkout не виконувався.

## Наступний хід

**Один рекомендований хід — M2: закрити R5 fixture leak у трьох e2e specs.** Обсяг: кожен producer
прибирає власні published fixtures, включно з failure-path, без глобального очищення каталогу;
читачі вибирають власний seed за ідентичністю (AGENTS.md §Decisions). Доказ майбутньої задачі — повтори без reset не
накопичують її published rows і не ламають catalog page 1, після чого повний e2e як для shared-state
зміни. Джерело проблеми: `docs/admin-editor-ux/plan.md` §5 R5, report §5; це не реалізація в agenda.
Передати як малий набір задач через `to-tickets` за локальним task route; код/гейт потребуватиме
окремого робочого вікна й review route, тут їх не запущено.

Чому не найближчі альтернативи:

- **M1 Start/Rsbuild:** явний request збережено і його prerequisite вже закрито; R5 спочатку усуває
  виміряний шум shared-state перед repeat/gate доказами сумісності. Це рекомендований порядок,
  не приписаний власнику новий пріоритет. Оновлення не починати автоматично.
- **M3 +5 px overflow:** конкретний UX-дефект і дешевий фікс, але він локальний; R5 впливає на
  достовірність наступних повних e2e, включно з перевіркою самого фіксу.
- **Продовжити all-screens design-pass тут:** інша сесія вже має AUTH-код/evidence; новий паралельний
  виконавець дублював би відповідальність та змішував ревізії. Її delivery/узгодження лишається там.

Деплой не конкурує з M2 як агентська задача: standing defer **Р1** чинний. Агенда рекомендує хід,
але не є дозволом запускати наступну задачу.
