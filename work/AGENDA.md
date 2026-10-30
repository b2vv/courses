# Агенда — що робити далі

| | |
|---|---|
| Дата | 2026-10-30 (попередній прогін — 2026-10-29 на `169c0fc`) |
| Спирається на брифінг | [work/CTO-RESEARCH.md](CTO-RESEARCH.md) **re-sync 2026-10-29** (база `169c0fc`). Доменні §2–§6 — кеш; **перед значущим ПРОДУКТОВИМ ходом — повний `cto-research`** (брифінг відстав на SA13-батч + S0-cutover-наслідки). |
| HEAD на момент агенди | `6efa393`. gh: **0 відкритих issues, 0 відкритих PR**. |
| Чи протух попередній вибір | Так — у хорошому сенсі. Агенда 2026-10-29 ставила наступним **SA13**; відтоді **весь SA13-батч змерджено** (D-F2 #83 … I-F6 #90) і **SA13-A-R residual закрито** датованим owner-accepted-bar (`6efa393`). **Жива не-блокована черга агента тепер ПОРОЖНЯ** — рівно як передбачала попередня агенда. |

**Зроблено з попередньої агенди (одним рядком):**
- **SA13-батч закрито повністю:** D-F2 cursor-UUID #83 · I-F5 TOTP replay #84 · T-F2 CSP form-action #85 · T-F1 safeReturnTo #86 · P-F3 /metrics guard #87 · P-F6 sslmode prod-tail (no-code) · P-F5 panic-log #88 · M7 HLS umask #89 · I-F6 totp_secret-шифрування #90. Кожен — незалежне рев'ю Courses 2 PASS.
- **SA13-A-R residual** (non-canonical UUID курсор → 503) — owner-ACCEPTED as dated limitation, no code (`6efa393`).

## Джерела та кількість кандидатів (усі 6 пройдено)

| Джерело | К-сть | Суть |
|---|---:|---|
| 1. Відкриті таски `work/` + gh | **0 живих** | SA13 закрито; **S01/S02** блоковані TLS/deploy; gh 0/0. Уся audit-черга вичерпана. |
| 2. Зафіксований борг (`backlog.md`) | 0 actionable | B-27 **блок upstream** (testcontainers #3868); B-2 DONE; B-3 DONE + format-residual (JSON Schema не виражає суму по модулях — не вирішуване в контракті). |
| 3. Ризики брифінгу §6 | 2 (низький сигнал) | 2 `⚠️ неперевірено`: витримка `limit_req` під навантаженням; механізм R23/identity-флейку (інструментовано K4/K5, не відтворено). Investigation, не фіча. |
| 4. `[ ]` + §Limitations | **3** (deferred follow-up) | lesson-builder §Lim (`docs/lesson-builder/report.md`): **`/teach`-редактор** (§Lim1:688), **друга локаль UI** (§Lim2:694), **orphaned-image sweep** (§Lim3:698, LB09a). Усі `[ ]` = 0. |
| 5. skip/TODO/FIXME/ponytail | 0 | 2 легітимні guard-skip (no-build-info, ffmpeg-absent); 0 боргових якорів у проді. |
| 6. Продуктові розвилки | 3 (+Р1 deploy) | мультивалюта · друга мова контенту · деплой (Р1, відкладено власником 2026-10-03). |

Кандидатів без джерела: **0**.

## Кандидати по двох осях (розблоковує / зворотність)

| Хід | Severity | Розблоковує | Зворотність |
|---|---|---|---|
| **orphaned-image sweep** (LB09a) | LOW hardening | нічого; закриває resource-leak байтів на диску при fail-after-upload | **зворотний** — окремий sweeper-job/крок, малий |
| **`/teach`-поверхня редактора** | feature | інструктор редагує власні документи уроків без admin-доступу | зворотний (нова FE-поверхня + backend-scoping `CanEditLesson`), **середній**; **owner-initiated** |
| limit_req під-навантаженням verify | — | підтверджує успадковану `⚠️` тезу | зворотний (load-тест); investigation, низький сигнал |

**Блоковані (не ходи зараз):** S01/S02 (TLS/deploy), B-27 (upstream #3868), деплой (owner), SA13 prod-хвіст (Grafana pw, COOKIE_SECURE, prometheus metrics-auth, POSTGRES_SSLMODE=verify-full+CA — усі потребують реального prod-середовища), I-F6 N1/N2 (key rotation/AAD — deploy/rotation tail).

## Рішення людини (не ходи агента)

1. **Деплой (Р1).** Відкладено власником 2026-10-03 (test-mode). **Що змінилось:** увесь security-audit (SA01–SA13) тепер закрито — головний pre-deploy блокер знято. Розблоковує найбільше: S01/S02 TLS, усі SA13 prod-хвости, фінальний live-payment на новому магазині+ключі. **Рекомендація: винести на перегляд** — деплой тепер набагато захищеніший, ніж коли його відкладали; але це людське рішення й зовнішній незворотний ефект.
2. **Друга мова контенту** (транскрипт primary є; редактор/reader single-locale — lesson-builder §Lim2 + форк B-21). **Рекомендація: не зараз** — без сигналу попиту.
3. **Мультивалюта** (ціни лише UAH). **Рекомендація: не зараз** — немає попиту, test-mode.

## Наступний хід

**Рішення власника 2026-10-30: послідовність (1) deploy-readiness → (3) orphaned-image sweep → (2) `/teach`-редактор.**

- **(1) ✅ ЗАКРИТО — deploy-readiness checklist** → [work/DEPLOY-READINESS.md](DEPLOY-READINESS.md) (`5a671d5`). Зведено всі deploy-gated хвости (A TLS/S01/S02 · B SA13 prod-хвости · C I-F6 ключ/ротація · D payment live · E observability) в один gate з джерелом+перевіркою+власник/агент на кожен.
- **(3) ✅ ЗАКРИТО — orphaned-image compensating delete (LB09a)** — PR #91 (`9a2756e`). Компенсаційне видалення при збої Insert (owner choice, не sweeper); повний `make check` EXIT=0, e2e 127; незалежне рев'ю Courses 2 PASS 0 blocking. 1 non-blocking N1 (cancelled-ctx робить компенсацію no-op — у report §Lim3 з механізмом; опційний майбутній close через `context.WithoutCancel`).
- **(2) ⬜ ДАЛІ — `/teach`-редактор** (owner-initiated фіча, поза §24; `docs/lesson-builder/report.md` §Lim1) — повний `cto-research` → `spec-flow`. Потребує owner-go на старт (значуща фіча, не gardening).

Чому не паралельно / не в іншому порядку: власник задав саме цей порядок.

Відкладено (не в цій послідовності): друга мова контенту, мультивалюта (продуктові розвилки без сигналу); B-27 (upstream #3868); limit_req verify (investigation без сигналу).
