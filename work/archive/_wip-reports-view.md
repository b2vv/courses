# V-track діагностика — view 7 `#reports` (ParticipantReportQueue)

**Гілка:** `feat/reports-view-designpass` (від main @c9aa39e). **Автор:** Courses 1. **Старт:** 2026-10-17.

## ФАКТ: прод уже card-grid, багатший за proto
`ParticipantReportQueue` (`admin/participant-reports`, admin-only) — уже `.admin-queue-card` черга:
`.admin-queue-heading` (icon-tile + eyebrow id·date + h2 kind + status-pill), `.admin-queue-details`
(3-col reporter/course/target), `.admin-queue-reason` (inline, proto ховає за Details-модалку),
`.admin-queue-actions` (Resolve + Revoke-діалог **gated on_student** — багатше за proto Resolve-only).
CSS `.admin-queue-*` вже стилізований (спільний із refund). proto `.queue-grid` — теж 1 колонка.

## ФАКТ: proto `#reports` (src.jsx case 'reports')
PageHeader eyebrow «SAFETY & TRUST» + title «Reports» + description + actions=Pill «N open»;
QueueFilter **All/Open/Resolved**; `.report-card` grid; Empty з reset-«All».

## РОЗБІЖНОСТІ
- **D1** PageHeader (eyebrow+title+**description**) — прод має bare `.admin-list-heading`. Головний гап.
- **Фільтр стану** All/Open/Resolved — proto має, прод ні.
- icon-tile info-tint (proto `.queue-heading>.ui-icon-tile{color:info}`), прод — default.
- Details-модалка — proto має, прод показує reason інлайн (**багатше, лишаємо**).

## ⚠️ КЛЮЧОВА ЗНАХІДКА: backend-черга open-only
`ParticipantReportQueue` SQL = `WHERE r.state='open'`; `ResolveParticipantReport` → `state='resolved'`
(випадає з черги). Тож proto-фільтр проти **поточного** бекенду беззмістовний (Resolved завжди
порожній, All==Open). Фільтр = фіча (SQL/port/http/query), не чистий design-pass.

## РІШЕННЯ scope (2026-10-17, Courses 1, вибір власника «C» через AskUserQuestion)
**Збудувати фільтр стану** (повна фіча): backend-черга фільтрується за станом і повертає resolved;
`?state=open|resolved|all`; фронт QueueFilter + PageHeader. Закриває альтернативи A (layout-only) і
B (+count-pill). Більший обсяг, окремий гейт.

## РІШЕННЯ default=Open (2026-10-17, Courses 1) — закриває proto-default=All
Default (без `?state=`) = **open**, не proto-`All`. Причина: (1) зберігає resolve→зникає-з-активної-
черги, тож e2e En07 (`toHaveCount(0)` після Вирішити) лишається валідним без правок; (2) модератор
бачить actionable-звернення першими. All/Resolved — явні вкладки. Перевірка: En07 на default-в'ю
зелений; новий e2e-кейс — перемикання на Resolved показує щойно вирішене. HTTP-мапа:
absent→open; `?state=all`→порт `""` (без WHERE); open/resolved — як є; інше → 400.

## ВЕРИФІКОВАНО + ЗДАНО (2026-10-18, Courses 1)
Фічу (scope C) реалізовано, доведено, відкрито PR.
- **Backend (discussion):** `ParticipantReportQueue(ctx, cursor, stateFilter, limit)` — SQL `WHERE ($4::text='' OR r.state=$4)`; threading port→app→SQL. HTTP `?state=open|resolved|all` (absent→open, all→"", інше→400; OpenAPI enum + Orval regen, check-generated 0). Тести: http-мапінг state, app-threading (fake фіксує filter), integration (resolved=1, all=3).
- **Frontend:** route sparse `?state` (open опущено), query key `[...base, state]` (prefix-інвалідація family у resolve/revoke); `ParticipantReportQueue` — PageHeader (D1 eyebrow «Безпека та довіра»+title+lede), SegmentedControl Усі/Відкриті/Вирішені (`.admin-queue-filter{justify-self:start}`), icon-tile `color:var(--color-info)`, empty-reset «Показати відкриті», Resolve/Revoke приховано для resolved. Домен `ParticipantReportStateFilter`. Локалі uk/en.
- **Гейт:** ордерний `make check` **exit 0, e2e 121 passed (15.7m)**. **1-й прогін почервонів на format-check (gofmt)** — я форматував лише frontend oxfmt, backend-тести лишились не-gofmt'нуті (вирівнювання полів структур); `gofmt -w` + перепрогін → зелено. Лог `frontend/test-logs/reports-view-gate-20261017T202928Z.log`.
- **Браузер (1440):** D1 PageHeader + QueueFilter рендер точний; open default (без param) показує open on_student з Resolve+Revoke; клік «Вирішені» → `?state=resolved` показує resolved on_lecturer зі status-pill «Вирішена» і БЕЗ action-кнопок. axe цього в'ю зелений через En07 (`wcagViolations`) у гейті.
- **PR #43** проти main: https://github.com/b2vv/courses/pull/43. Далі — view 8 #refunds, тоді A-батч.
