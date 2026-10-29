# Агенда — що робити далі

| | |
|---|---|
| Дата | 2026-10-29 (попередній прогін — 2026-10-26 на `11d0b7e`) |
| Спирається на брифінг | [work/CTO-RESEARCH.md](CTO-RESEARCH.md) **re-sync 2026-10-29** (база `169c0fc`): шапка + §1 перезібрані під security-audit + frontend feature-modules + generate-OpenAPI-from-Go. Архітектурний факт №1 — `openapi.yaml` тепер генерується з Go й authoritative. Детальні доменні розділи §2–§6 — кеш (точний для доменних модулів); **перед значущим ПРОДУКТОВИМ ходом — повний `cto-research`** (§3/§4/§5 доменних фаз не переписувались порядково). |
| HEAD на момент агенди | `169c0fc`. gh: **0 відкритих issues, 0 відкритих PR**. |
| Чи протух попередній вибір | Так — у хорошому сенсі. Агенда 2026-10-26 ставила наступним **SA08**; відтоді SA08 і вся решта security-черги (SA09, SA09-T1, SA10/M8+S0, SA11, SA12) **змерджені**. Висновок уточнено: **роадмап §24 пройдено; security-audit вичерпано, крім SA13 (не блокована) і TLS-gated S01/S02.** |

**Зроблено з попередньої агенди (одним рядком):**
- **SA08** M1 fail-closed limiter (#75) · **SA09** test-gaps M6/M10/O-F2 (#78) · **SA09-T1** O-F3 concurrent events (#79) · **SA11** M11 error-codes (#76) · **SA12** M12 enumeration (#77).
- **SA10:** M8 cross-module isolation rule (#80); **S0 generate-OpenAPI-from-Go — повний цикл закрито**: S0.1 (11 контриб'юторів, 85/85 покриття) → S0.2a CLI+make-wire → S0.2b cutover (`openapi.yaml` authoritative, PR #81) → follow-up зріз (prune + 2 census-binding + within-graph collision guard + mutual-recursion fail-closed, PR #82). Кожен зріз — незалежне рев'ю Courses 2 PASS.
- Усі merge пройшли **обов'язкове рев'ю Courses 2** (правило власника 2026-10-26).

## Джерела та кількість кандидатів (усі 6 пройдено)

| Джерело | К-сть | Суть |
|---|---:|---|
| 1. Відкриті таски `work/` + gh | **3** | **SA13** (1, cite `audit-2026-10-24.md` §LOW) — єдина жива не-блокована. **S01/S02** (2) — блоковані TLS/deploy. gh 0/0. SA08–SA12/SA09-T1/SA10 — усі `done/merged`. |
| 2. Зафіксований борг (`backlog.md`) | 3 | B-27 (reaper race — **блок upstream testcontainers** #3868), B-2 (adapter-DOM коштує header-coverage), B-3 (контракт дозволяє 4× уроків за домен). **B-1 observability — закрито** (PR #62). B-7…B-18 — пост-мортеми. |
| 3. Ризики брифінгу §6 | re-synced | §6 ще несе доменні ризики; CI-ризик знято (локальний гейт авторитетний), B-1 знято. Нові gotchas (mutual-recursion fail-closed, within-graph collision) — у `_gotchas.md`, не нові таски. |
| 4. `[ ]` чекбокси + §Limitations | 0 + ~N | невиконаних `[ ]` = 0. Named-limitations у звітах фаз + security §Remediation — усі свідомі, з магнітудою/механізмом. |
| 5. skip/TODO/FIXME/ponytail | 0 | легітимні guard-skip (no-build-info, ffmpeg-absent); жодного боргового якоря. |
| 6. Продуктові розвилки | 3 (+Р1 deploy) | мультивалюта · друга мова контенту (editor+UI) · деплой (Р1, відкладено власником). lecturer-revoke **відхилено 2026-10-13**. |

Кандидатів без джерела: **0** — security-черга cite `audit-2026-10-24.md`, борг cite `backlog.md`.

## Кандидати по двох осях (розблоковує / зворотність)

| Хід | Severity | Розблоковує | Зворотність |
|---|---|---|---|
| **SA13** LOW/hardening batch | LOW | закриває security-audit (крім TLS-gated) | змішана — більшість зворотні (cursor-UUID→400, CSP form-action, /metrics guard, safeReturnTo, panic-лог, HLS umask, sslmode env); `totp_secret`-шифрування — data-ish міграція; Grafana-pw + `COOKIE_SECURE` — **prod-only, відкласти до deploy** |
| B-2 adapter-DOM | — | header-coverage | висока (тест-інфра) |
| B-3 контракт 4× уроків/домен | — | мало | висока (контракт-ліміт) |

**Блоковані (не ходи зараз):** S01/S02 (TLS/deploy), B-27 (upstream #3868), деплой (owner).

## Рішення людини (не ходи агента)

1. **Мультивалюта** (ціни лише UAH). **Рекомендація: не зараз** — немає попиту в джерелах, test-mode.
2. **Друга мова контенту** (транскрипт primary є, редактор-UI нема — форк із B-21). **Рекомендація: не зараз** — owner-initiated, без сигналу попиту.
3. **Деплой (Р1).** Відкладено власником (test-mode). Людське рішення, не черга. Розблоковує S01/S02 + prod-хвіст SA13.
4. **Після SA13 — що далі.** Security-audit буде вичерпано (лишаться тільки deploy-gated). Жива черга стане **порожньою**: наступний хід — рішення власника (деплой, нова owner-initiated фіча, або добивання блокованих хвостів після TLS). Варто винести на AGENDA окремо, коли SA13 закрито.

## Наступний хід

**SA13 — LOW/hardening batch** (cite `docs/security/audit-2026-10-24.md` §LOW), розбити на дрібні незалежні зрізи, кожен зі своїм failing-test/доказом.

Чому саме він:
- **Єдина жива не-блокована security-задача.** Уся решта audit-черги (SA01–SA12, SA09-T1, SA10/S0) змерджена; SA13 закриває хвіст.
- **Дешевий і подрібнюваний:** cursor-decoder UUID→400 (D-F2), TOTP replay consumed-counter (I-F5), CSP `form-action` (T-F2), `safeReturnTo` += teach/settings/orders (T-F1), `/metrics` in-app guard (P-F3), PostgresDSN `sslmode` env (P-F6), panic value/stack лог (P-F5), HLS umask 0027 (M7) — усі малі, test-mode-сумісні.
- **Зворотний** у більшості; виняток — шифрування `totp_secret` (I-F6, міграція) робити окремим обережним зрізом.

Що **винести/відкласти в межах SA13:**
- Prod-only хвіст (Grafana pw, `COOKIE_SECURE=true`) — до deploy; він перетинається з блокованими S01/S02.

Чому не інші:
- **B-2/B-3** — борг без severity; прибирають мало тертя, не закривають дірок. Після SA13, якщо власник не відкриє продуктовий хід.
- **S01/S02, деплой** — блоковані TLS / рішення власника.

Далі: SA13 (по зрізах) → pre-code checkpoint кожного → ordered `make check`/e2e за впливом → PR → **обов'язкове рев'ю Courses 2** → owner merge. Коли SA13 закрито — нова AGENDA: черга порожня, наступне — рішення власника.
