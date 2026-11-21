# Агенда — що робити далі

| | |
|---|---|
| Дата | 2026-11-21 (rebuild після M1 #114 і design-pass #115); попередній rebuild `63e5ebca` (2026-11-19) |
| Спирається на брифінг | [work/CTO-RESEARCH.md](CTO-RESEARCH.md), delta до **`f8dd8589`** у тому ж коміті, що й ця агенда (рядок «Гейт свіжості 2026-11-21»). Продуктові межі/модулі/контракт не змінились; змінились §5 (піни TanStack/Rsbuild) і §6 (design-pass, межі M1). |
| HEAD на момент агенди | `main` @ `f8dd8589`; `gh issue list --state open` → `[]` (**0 issues**); `gh pr list --state open` → `[]` (**0 open PR**); на GitHub лише гілка `main`. |
| Чи протух попередній вибір | Так: обраний хід M1 виконано (#114). Роадмап §24 завершено; деплой — відкладене людське рішення (Р1). |
| Хеші | **Історію `main` переписано 2026-11-21.** Хеші нижче — нові. Хеші після 2026-10-18 у phase reports і старих агендах — до переписування; шукай коміт за номером PR. |

**Зроблено з попередньої агенди:**

- **done: M2 / R5 published e2e leftovers → #112 / `0db1c0de`** (ordered `make check`, e2e 168; `docs/e2e-hygiene/report.md`).
- **done: M19 / R6 archived e2e leftovers → #113 / `4d0dfc23`** (ordered `make check`, e2e 168; там само §R6).
- **done: M1 Start + Rsbuild compatibility-check → #114 / `da9e447f`** — піни Start 1.169.7 / Router 1.170.46 / Rsbuild 2.3.3 / plugin-react 2.1.1; C0–C5 PASS; ordered `make check` GATE_EXIT=0, e2e 168; `docs/start-rsbuild-check/report.md`.
- **done: AI Factory adapter files local-only → `bdc7048c`** + переписування історії (рішення власника 2026-11-21).
- **done: design-pass S0 + AUTH-зріз → #115 / `d6147d71`** — S0-реєстр 31 знахідки (`docs/design-pass/findings.md`); A1-1 (blocker), A2-1, A5-1 закрито; Codex-рев'ю PASS (`docs/design-pass/_wip-review*.md`).

Цей rebuild не запускав builds/tests/Docker і не заявляє нового runtime/CI PASS.

## Джерела та кількість кандидатів (усі 6 пройдено)

К-сть — ходи/розвилки з реєстру нижче, **з перетинами** між джерелами. Унікальних: **30**
(DP1–DP3, M3–M18, M20–M21, D1–D2, Р1–Р4, Q3, Q4, Q5).

| Джерело | К-сть | Суть / фільтр |
|---|---:|---|
| 1. Відкриті таски `work/` + gh | **5** | DP1 planned, DP2 in_progress, DP3 planned (`work/tasks.md` §design-pass); D1/S01, D2/S02 (`work/tasks.md:38–39`). gh **0 open issues**. B01/A10/A13/SA13/TE4 відфільтровані closing evidence (як у попередній агенді, `63e5ebca`). |
| 2. Зафіксований борг / limitations | **18** | M3–M15, M20, M21, Р1/Р2/Р4. Нові: M20 — `docs/design-pass/plan.md` §10 (retry-форми `.notice` NOT_TESTED); M21 — `docs/start-rsbuild-check/report.md` §Limitations. M19 знято (#113). |
| 3. Ризики брифінгу §6 | **21** | M3–M7, M9–M12, M14–M18, Р1/Р2/Р4 (як раніше) + нові рядки §6: design-pass (DP1, DP2, M20) і межі M1 (M21). |
| 4. Unchecked boxes у live specs/plans | **4** | Unchecked пунктів поза archive/`_wip` — **15**, усі в `work/DEPLOY-READINESS.md` (A1–E2, `grep -n -- '- \[ \]' work/DEPLOY-READINESS.md`) → D1/D2/Р1/Р4, уже враховані. |
| 5. Skipped tests / disabled gates / TODO у коді | **0** | `git grep -n -E "test\.skip\|t\.Skip\(\|\.only\(\|TODO\|FIXME" -- backend frontend/src frontend/e2e` → **1**: `t.Skip("no build info")`, `backend/internal/testsupport/containers/guard_test.go:17` — явна межа, не дефект. |
| 6. Невирішені продуктові розвилки | **7** | Р1–Р4 (як раніше); design-pass Q3/Q4 (`docs/design-pass/plan.md` §8) і Q5 — вибір шрифту для S-3 (`findings.md` S-3: «веб-моно-шрифт (нова залежність → рішення власника)»). |

Кандидатів з міткою `⚠️ no source`: **0**.

## Кандидати по двох осях (розблоковує / зворотність)

| Хід | Джерело / стан | Розблоковує | Зворотність / ціна помилки |
|---|---|---|---|
| **DP1. design-pass S1 — спільні примітиви** | `docs/design-pass/plan.md` §5 п.2 (S1 обов'язковий, бо S0 знайшов спільні major); `findings.md` S-1…S-6 | Усі поекранні зрізи DP2: S-1 (`overflow-wrap:anywhere` рве слова на 200%), S-3 (`ui-monospace`=Courier), S-4 (однаковий `<title>`, WCAG 2.4.2), S-5 (wordmark 38 px < 44) змінюють кожен екран; без S1 докази поекранних зрізів доведеться переробляти. | Дешевий CSS/route-head edit, відкотний; глобальний стиль → повний e2e (урок `AGENTS.md` §Decisions). S-3 з веб-шрифтом — нова залежність (Q5). |
| DP2. design-pass поекранні зрізи | `plan.md` §5 п.3 (G0 → C1 → … → T3); 28 відкритих знахідок (16 major, 12 minor) | Видимі дефекти на клієнті/адмінці: C1-1/2/3 (головна), C3-1 (немає ціни/CTA на платному курсі), L2-1…3 (урок без h1), T0-1 (порожня область `/teach`), T1-1 (робоче місце без стилів). D0-1 (`/admin/<невідомий>`) уже закрито #111 (`docs/admin-editor-ux/report.md:13`, catch-all кидає `notFound()`) — лише повторна перевірка, не фікс. | Presentation edits, відкотні; ≈5 PR за Q2. Після DP1. |
| DP3. design-pass report + фінальний gate | `work/tasks.md` DP3 | Закриття задачі. | Docs + gate; після DP2. |
| M20. Retry-форми `.notice` після #115 | `docs/design-pass/plan.md` §10 Scoped impact (NOT_TESTED) | Доказ, що нове правило `.notice form` не зіпсувало помилку сторінки (`__root.tsx`) і збій логіну (`routes/login.tsx`). | Read-only браузерна перевірка; фікс (якщо треба) — один CSS-рядок. Природно входить у DP1 (той самий `main.css`). |
| M21. M1: native amd64 proof | `docs/start-rsbuild-check/report.md` §Limitations | C5 доведено через Rosetta на task-owned builder; native amd64 (CI) не запускався, r2 exit 1 без причини. | Дешево, коли CI живий. Зараз workflow `foundation` має стан `disabled_manually` (`gh api …/actions/workflows`); останній run 2026-10-19 падав на старті з повідомленням про платежі/spending limit — історичний доказ, поточна причина не перевірена. Ввімкнення CI — рішення власника. |
| M3. Normal transcript +5 px при 390/200% | `docs/multi-subtitles/report.md:1008,1023` | Читання панелі без горизонтального виходу. | Дешевий CSS; може зачепитись S-1 у DP1 — перевірити там само. |
| M4. Періодичне recovery загубленого upload enqueue | `docs/multi-audio/report.md:190–196` | Відновлення без рестарту worker. | Середня; не зараз без сценарію. |
| M5. Справжній sweep/PATCH race test | `docs/multi-audio/report.md:204–210` | Доказ row-lock. | Дешевий test-only. |
| M6. `awaiting_audio` у adminstats | `docs/multi-audio/report.md:239–240` | Правдивий media-лічильник. | Невеликий SQL/DTO/UI. |
| M7. Видима причина failed audio / заміни | `docs/multi-audio/report.md:275–278` | Автор розуміє падіння доріжки. | Середня: контракт+UI. |
| M8. Recovery version id замість `capturedAt` | `docs/multi-subtitles/report.md:541–544` | Колізія в одну ms. | Persisted формат; не зараз без повтору. |
| M9. Kiri capacity перед heavy-вікном | `docs/start-rsbuild-check/report.md` §Environment: VM 13 GB free після cleanup, `kiri-build` cache 14.5→≈21 GB, не pruned | Місце для DP-гейтів і знімків. | Read-only оцінка дешева; prune — лише з дозволу. Перевірити перед першим gate DP1. |
| M10. Real Safari native playback/captions | `docs/playback/report.md:297–302`, `docs/multi-subtitles/report.md:1021–1022` | Native branch. | Потрібен реальний Safari. |
| M11. Playback proof: ABR / cadence / AV1 wire | `docs/playback/report.md:312–320,344–348` | Доказ shipped властивостей. | Тривалі прогони. |
| M12. ASR deadline/startup failure-path | `docs/asr/report.md:155–165` | Чесна named-failure fixture. | Спершу механізм. |
| M13. OpenTelemetry end-to-end spans | `work/backlog.md:143` | SSR→API→worker spans. | Дослідження read-only, помилка нічого не коштує; додавання spans — відкотний код. Не зараз: немає споживача. |
| M14. B-27 testcontainers retry removal | `work/backlog.md:7–33`; `backend/go.mod:17` | Прибирає wrapper, якщо release містить #3868. | Upstream status не перевірявся. |
| M15. Visual proof gaps: translucent contrast / Firefox zoom | `docs/admin-editor-ux/report.md:87–92`, `docs/multi-subtitles/report.md:248–258` | Виміряний composited contrast. | NOT_TESTED ≠ FAIL. |
| M16. R23 heartbeat 403 — механізм при повторі | `work/CTO-RESEARCH.md` §6 | Діагноз. | Читання логів/інструментація, повністю відкотні; без повтору нема що міряти. |
| M17. Identity integration одноразове падіння | `work/CTO-RESEARCH.md` §6 | Діагноз. | Читання K5-логу, без змін коду — дешево й відкотно; без повтору не відтворити. |
| M18. NAT-профіль `limit_req` | `work/CTO-RESEARCH.md` §6; `docs/readiness/report.md` §7 | Capacity під NAT. | Перед exposure. |
| D1. S01 TLS-залежні headers/cookies | `work/tasks.md:38`, `DEPLOY-READINESS.md` | HSTS/Secure на prod. | Конфіг відкотний, але HSTS кешується браузерами (max-age) — помилку не відкликати миттєво. **Блок Р1.** |
| D2. S02 `__Host-` session-cookie | `work/tasks.md:39` | Duplicate-cookie. | Дорожче: перейменування cookie розлогінює всіх, назад — ще раз. **Блок D1/Р1.** |

Named boundaries доставлених фаз (multi-audio S7/backfill, multi-subtitles plan §6, admin-editor-ux §5,
lesson-builder crash-window, adminstats placeholders, discussion pagination, transcript 87 rows) лишаються
у phase reports і цим документом не закриваються (повний перелік — попередня агенда `63e5ebca`).

## Рішення людини (не ходи агента)

**Р1 — деплой лишається deferred** (`AGENTS.md:5`: “Production deployment was deferred by the owner on 2026-10-03”).

| Розвилка / джерело | Опції та ціна | Рекомендація з причиною | Розблоковує / зворотність |
|---|---|---|---|
| **Q3. Heavy-вікно для design-pass.** ✅ вирішено 2026-11-21: (a). `docs/design-pass/plan.md` §8 | (a) Ексклюзивне вікно в основному checkout: стек на 8080 перезбирається з поточної гілки, `make check` по черзі. (b) Окремий compose-проєкт/порт — зміна інфри. | **(a)**: умова питання змінилась — паралельної гілки більше немає, вся робота в одному checkout, і #115 так і пройшов (ordered `make check` на `5e55c88a` GATE_EXIT=0, e2e 168 — тіло PR #115). (b) коштує інфри без споживача. | DP1–DP3; відкотне. |
| **Q4. Рев'юер design-pass.** ✅ вирішено 2026-11-21: Codex напряму. `plan.md` §8 | Codex напряму (independent_subagent) або Courses 2 (existing_session). | **Codex напряму**: так рев'юили #114/#115, обидва знайшли реальні дефекти (BLOCK → fix → PASS). Courses 2 — якщо власник хоче другу сесію, як у пам'яті «рев'ю обов'язкове». | Кожен PR DP; відкотне. |
| **Q5. Шрифт для S-3 (`--font-utility`).** ✅ вирішено 2026-11-21: Inter + `tabular-nums`. `findings.md` S-3 | Inter + `tabular-nums`: без залежності, цифри рівні, мітки втрачають «технічний» вигляд. Веб-моно-шрифт: нова npm-залежність, вага шрифту, ліцензія. | **Inter + `tabular-nums`**: прибирає Courier без нової залежності; мітки й ціни лишаються читабельними. Моно — лише якщо власник хоче цей стиль. | DP1; дешевий відкат CSS. |
| **Р1. Коли відкривати production?** `AGENTS.md:5`; `DEPLOY-READINESS.md` | Test-mode vs launch (сервер/TLS, секрети, email, monitoring, live shop). | Зберегти defer до наміру й середовища запуску. | D1/D2, prod tails; дорогий зовнішній ефект. |
| **Р4. Ротація 2FA master до deploy.** `DEPLOY-READINESS.md:29–31` | Re-enrol при ротації vs версійований keyring. | До launch описати re-enrol; keyring — за вимоги безперервної ротації. | C2 deploy gate; до Р1 не запускати. |
| **Р2. Друга locale lesson-document.** `docs/lesson-builder/report.md:694–697` | Одна locale vs switch/create/read другої. | Не зараз: немає конкретного курсу з вимогою. | Відкотний код, дорожчі дані. |
| **Р3. Мультивалюта.** `docs/payment/spec.md:35,39,97` | UAH vs друга валюта. | Не зараз: spec:97 лишає за попитом. | Фінансові записи дорогі для відкату. |

**Revoke не питання** (`AGENTS.md:8`, admin-only за рішенням власника 2026-10-13).

## Наступний хід

**Один рекомендований хід — DP1: design-pass S1, спільні примітиви** (S-1, S-3, S-4, S-5 + перевірка M20 і M3
на тих самих viewport). Підстава — план, який власник уже погодив (Q1/Q2 2026-11-16, `plan.md` §9): S1 іде
перед поекранними зрізами, бо S0 знайшов спільні major. S-2 (phone-хедер, нова mobile-навігація) — окремий
зріз у DP1, якщо пре-код checkpoint покаже, що він більший за CSS.
Передумови: Q3/Q4/Q5 **вирішено власником 2026-11-21** («Так, усі три» — прийнято рекомендації нижче; `docs/design-pass/plan.md` §9); лишається read-only перевірка місця на Kiri (M9).
Шлях — pre-code checkpoint у `docs/design-pass/plan.md` → коміт на екран → ordered `make check` → Codex-рев'ю → PR.

Чому не найближчі альтернативи:

- **DP2 одразу з C1:** план вимагає S1 спершу; поекранні докази, зняті до S-1/S-3/S-4, довелося б знімати вдруге.
- **M3 +5 px / M20 retry-форми окремо:** обидва дрібні й лежать у тих самих стилях, тож дешевше перевірити їх у DP1, ніж відкривати окремі PR.
- **M21 native amd64:** CI-workflow вимкнено (`disabled_manually`); без native amd64 доказ не отримати, а ввімкнення — рішення власника.

Деплой не конкурує як агентська задача (Р1). Агенда рекомендує хід, але не є дозволом його запускати.
