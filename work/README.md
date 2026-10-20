# Courses: робота з агентом за AI Factory

Це короткий вхід у наявний процес Courses, не новий набір правил або tracker.
Для початку достатньо відкрити агента в корені цього репозиторію й передати задачу
разом із промптом нижче. Workflow не потребує запуску сервера чи scanner.

## Що прочитати спочатку

1. [AGENTS.md](../AGENTS.md) → [standards](../.claude/standards.md) →
   [локальний AI Factory adapter](../docs/development/ai-factory.md).
   [CLAUDE.md](../CLAUDE.md) уже веде до AGENTS і checkpoint; Codex починає з AGENTS.
2. [CTO-RESEARCH.md](CTO-RESEARCH.md) — брифінг, [AGENDA.md](AGENDA.md) — контекст
   наступної роботи, [tasks.md](tasks.md) — індекс задач. Історичні записи не означають,
   що зараз відкрита відповідна фаза або вже дозволено наступний пункт agenda.
3. Чинні spec/acceptance/plan/report конкретної задачі в `docs/<phase>/`, застосовні
   [ADR](../docs/adr/) і вкладені instructions. Їх визначають задача та її pointers,
   а не припущення про «останній» файл. Локальні contracts і доручення мають пріоритет.

## Який AI Factory підключений

Джерело істини — [package metadata](../docs/development/ai-factory-package.json)
і [mapping в adapter](../docs/development/ai-factory.md#portable-package-pin).
Перевірено 2026-10-20: repository `https://github.com/b2vv/ai-factory`,
version **0.7.0-candidate**, commit **`c427bd807f250e8d4b7d9c42d9bc894d8f8dce71`**.

Immutable cache на цій машині:
`~/project-data/ai-factory-rollout/releases/c427bd807f250e8d4b7d9c42d9bc894d8f8dce71`.
Manifest SHA-256: `723c3477e8b47ae1ecc3d0b8a411d2c9e196a1e4f39c218de58238f30ce22608`.
Manifest і всі 147 файлів із `required_files` збіглися з pin metadata.
Це перевірка bytes, не доказ виконання workflow агентом. Metadata досі позначає
незалежне review rollout як `PENDING`, native discovery — `NOT_TESTED`.

У цьому release головний вхід — `ai-factory/ai-factory-plan.md`; потрібні деталі:
`ai-factory/workflow.md`, `pre-code-checkpoint.md`, `feature-lifecycle.md` і
`review-handoff.md` (останні три також у `ai-factory/`). Exact-file читання дозволяє
застосувати instructions без припущення, що однойменний skill автоматично discovered.
Чинний owner override щодо task branches описаний окремо в локальному adapter.

Не підміняти pin робочим checkout `~/project-data/ai-factory` і не підключати його
незакомічений scanner baseline. Оновлення shared package не перемикає Courses автоматично.
Цей README не створює залежності від іншого продуктового репозиторію.

## Почати або продовжити задачу

```text
Працюй у Courses за work/README.md. Прочитай AGENTS.md,
.claude/standards.md і docs/development/ai-factory.md; перевір чинний pin
та потрібні shared файли. Задача: [мета або ID і шлях чинного plan/report].
Спочатку покажи scope/acceptance, root/branch/HEAD і relevant WIP,
власника запису, маршрут small/full/resume та pre-code READY/BLOCKED.
Використовуй наявні task artifacts і погоджений незалежний review route.
Збережи локальні gates, чужу роботу й межі дозволів; не роби auto-install,
оновлення pin або запуск scanner. Далі виконуй саме погоджену задачу.
```

- **Small:** короткий plan/checkpoint у чинному record і перевірки за впливом;
  required gates не скасовуються. **Full:** застосувати поріг зі standards та один
  обраний процес, без паралельного дублювання spec-flow.
- До коду: scope/acceptance, рішення й невідоме, залежності/writer, scoped impact,
  застосовна planning critique/defense та READY/BLOCKED. Деталі —
  [pre-code checkpoint](../docs/development/ai-factory.md#pre-code-checkpoint).
- **Resume:** прочитати поточний plan/report/handoff, звірити HEAD, scoped dirty/untracked
  inputs і докази; повторно перевірити лише зачеплене, зберігши сильніші локальні вимоги.
  Незмінений завершений handoff не відкриває задачу знову.
- Review route та exact reviewer reference беруться з поточного доручення/record.
  Якщо їх немає, уточнити перед spawn/надсиланням; безпечне читання може тривати.
  Наявність reviewer або авторський self-check не означають незалежного PASS.

Нова сесія не потрібна для явного читання цих файлів. Свіжий контекст корисний для
перевірки startup routing; ні промпт, ні наявність pointer не гарантують автоматичної
дисципліни агента. Проси конкретні джерела та докази застосування правил до задачі.

## Артефакти, перевірки та завершення

- `docs/<phase>/`: чинні spec, acceptance, plan, gates, evidence і `report.md`.
  Під час роботи — `assumptions.md` або `_wip-<topic>.md`; не створювати другу фазу/трекер
  лише для AI Factory.
- [tasks.md](tasks.md): стани й result/archive pointers; завершені рядки залишаються.
  [archive/ledger.md](archive/ledger.md): coordination/history, append-only за змістом.
- Review evidence/verdict і pointer на пакет зберігаються в чинному task/phase record
  або його evidence-файлах. Пакет прив'язаний до exact diff/snapshot, acceptance,
  commands/exits і reviewer reference. Історія підключення Factory окремо —
  [readiness](../docs/development/ai-factory-readiness.md); її старий PASS не приймає нову задачу.

Завершення: кожен acceptance item має перевірений результат, застосовні gates та
незалежний verdict стосуються фінального snapshot, limitations і next action названі.
PASS/FAIL/BLOCKED/NOT_TESTED/N/A не підміняють одне одного; aggregate green або build
не доводять усі поведінкові рядки. [Testing](../docs/development/testing.md), standards
і [adapter](../docs/development/ai-factory.md) визначають перевірки та side effects.
`make check`/`e2e-local` не є безпечними перевірками цього README: вони змінюють стан,
а важкі gates потребують погодженого вікна. Commit/push/deploy — лише в дозволеному scope.

## Якщо Factory недоступний

Назви відсутній шлях або hash mismatch і BLOCKED саме залежної дії. Не роби auto-clone,
install, relink чи заміну pin. Локальні AGENTS/standards/adapter і task artifacts
залишаються доступним джерелом для незалежного читання та планування. Якщо required
shared contract/review недоступний, не оголошуй його виконаним: потрібне відновлення
того самого погодженого release або окремо погоджений fallback. Відсутній optional
CRG/Obsidian допускає файловий fallback за adapter, без автоматичного refresh.
