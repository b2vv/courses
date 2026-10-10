# Незалежні gates foundation

Дата: 2026-09-03. Reviewer: `plan_defense`, незалежний від автора плану. Обсяг: намір Phase 1 → Phase 2 у [плані](plan.md) і [специфікації](spec.md), звірений з [AGENTS.md](../../AGENTS.md). Це review документації, а не доказ виконаних runtime-перевірок.

## Plan-critique: перевірка рішень

Координатор передав результати чотирьох незалежних лінз: logic — 2 blocking, edges — 2 blocking, slop — 2 non-blocking, complexity — без знахідок. [Ledger](../../work/archive/ledger.md) підтверджує 4 launched / 4 returned, зокрема `plan_complexity` — 0 blocking / 0 non-blocking. Reviewer нижче перевірив наявність виправлень у поточному плані, не підміняючи цим авторство початкових critiques.

| Лінза / знахідка | Рішення у плані | Перевірка |
|---|---|---|
| Logic: Phase 1 gate залежав від майбутнього Course slice | [Phase 1 shell gate](plan.md#phase-1-shell-gate-та-відновлення), рядок 32: P01–P03 перевіряють shell/runtime; Course acceptance починається після C01–C04 | Blocking закрито |
| Logic: відновлення після dependency failure не визначене | Той самий розділ, рядок 34: доступний retry/reload з повним URL і E11 після відновлення API | Blocking закрито |
| Edges: невдала міграція могла допустити запуск застосунку | Той самий розділ, рядок 30: `service_completed_successfully`, failed migration fails `make up`, volumes збережені | Blocking закрито |
| Edges: pagination могла створювати неприйнятний URL або безвихідний порожній екран | Той самий розділ, рядок 38: reset offset, обмежений Next, total і перехід на першу сторінку | Blocking закрито |
| Slop: повний UI inventory та session-aware forwarding до auth | [Dependency direction](plan.md#реалізація-та-dependency-direction), рядки 14, 16–17: явні вимоги користувача збережені; auth залишається поза поточним scope | 2 non-blocking прийнято з причиною |
| Complexity | `plan_complexity` — без знахідок; отримання зафіксовано в ledger | Чисто, 0/0 |

## Plan-defense: п'ять питань

Посилання ведуть у відповідні розділи плану; номери рядків відповідають версії, прочитаній 2026-09-03.

| # | Питання | Закрито рядком | Вердикт |
|---|---|---|---|
| 1 | Кривий інпут | [API та дані](plan.md#api-та-дані), рядки 22–24: canonical slug, межі query, duplicate/nonnumeric rejection, 400 `invalid_query`/`invalid_slug` envelope; [E01](plan.md#приймальні-сценарії), рядок 62 задає спостережувану відповідь | Закрито |
| 2 | Відмова залежності | [API та дані](plan.md#api-та-дані), рядки 22–24: 503, bounded timeout/cancellation, SSR failure не є empty success; [shell/recovery](plan.md#phase-1-shell-gate-та-відновлення), рядки 30, 34: failed migration блокує запуск, доступний retry зі збереженим URL; E03–E06 перевіряють DB/API failure | Закрито |
| 3 | Гонка й повторний виклик | [API та дані](plan.md#api-та-дані), рядки 24–26: snapshot list/count, stable order, read-only endpoint; [shell/recovery](plan.md#phase-1-shell-gate-та-відновлення), рядок 36: усі нормалізовані параметри в Query key, rendering лише active key, obsolete abort і C03/C04 reversed A/B test; рядок 38: atomic concurrent seed зі stable UUID/`ON CONFLICT DO NOTHING` | Закрито |
| 4 | Порожні та граничні дані | [B01–B09](plan.md#приймальні-сценарії), рядки 53–61: 0/1/max, ties, hidden records, zoom/themes, repeated seed; [shell/recovery](plan.md#phase-1-shell-gate-та-відновлення), рядок 38: pagination ceiling та first-page recovery; [API та дані](plan.md#api-та-дані), рядки 22–24: плоскі Course list/detail і bounded input | Закрито |
| 5 | Rollback | [Rollback](plan.md#rollback), рядок 99: попередні image/config revisions, збережені volumes, forward schema fix, destructive reset лише окремою явною дією; [міграції](plan.md#план-міграції), рядок 95: down лише в ephemeral DB; [side effects](plan.md#сайд-ефекти), рядок 91: немає платежів, повідомлень чи публічного deployment | Закрито |

**Підсумковий вердикт: зелений — 0 незакритих питань.** На першому проході рядок 3 був діркою: план не визначав захист від запізнілої відповіді попереднього URL. Після правки автора повторний review підтвердив явний Query key, active-key rendering, cancellation та reversed A/B acceptance у рядку 36 плану. Це закриває дірку на рівні плану; виконання тесту ще належить реалізації C03/C04.

Reviewer оновив підсумковий статус і п'ять рядків незалежного defense у [плані](plan.md#план-defense--незалежний-review-завершено), як вимагає `plan-defense`. Координатор відображає цей результат у ledger. Зелені planning gates не означають завершення F06 або C05.
