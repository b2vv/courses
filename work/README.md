# Courses: робота з агентом за AI Factory

Це короткий вхід у наявний процес Courses, не новий набір правил або tracker.
Для початку достатньо відкрити агента в корені цього репозиторію й передати задачу
разом із промптом нижче. Workflow не потребує запуску сервера чи scanner.

## Що прочитати спочатку

1. [AGENTS.md](../AGENTS.md) (розділ «AI workflow» — правила Courses поверх kit) →
   [standards](../.claude/standards.md). Локальний `AGENTS.local.md` (поза git) підключає
   керований kit. [CLAUDE.md](../CLAUDE.md) уже веде до AGENTS і checkpoint; Codex починає з AGENTS.
2. [CTO-RESEARCH.md](CTO-RESEARCH.md) — брифінг, [AGENDA.md](AGENDA.md) — контекст
   наступної роботи, [tasks.md](tasks.md) — індекс задач. Історичні записи не означають,
   що зараз відкрита відповідна фаза або вже дозволено наступний пункт agenda.
3. Чинні spec/acceptance/plan/report конкретної задачі в `docs/<phase>/`, застосовні
   [ADR](../docs/adr/) і вкладені instructions. Їх визначають задача та її pointers,
   а не припущення про «останній» файл. Локальні contracts і доручення мають пріоритет.

## Який AI Factory підключений

Керований AI Factory kit на цій машині, підключений через локальний `AGENTS.local.md`
(поза git). Його правила й skills — завжди остання встановлена версія; ідентичність kit
читається з його `installed.json` (`latest`) у момент роботи. У цьому репозиторії немає
pin версії, commit чи хешу kit.

Головний вхід kit — `ai-factory/ai-factory-plan.md`; потрібні деталі:
`ai-factory/workflow.md`, `pre-code-checkpoint.md`, `feature-lifecycle.md` і
`review-handoff.md` (усі в `ai-factory/` kit). Правила Courses поверх kit (task branch,
один writer, де живе checkpoint) — у [AGENTS.md](../AGENTS.md#ai-workflow).
Цей README не створює залежності від іншого продуктового репозиторію.

## Почати або продовжити задачу

```text
Працюй у Courses за work/README.md. Прочитай AGENTS.md,
.claude/standards.md і потрібні контракти керованого kit. Задача: [мета або ID і шлях чинного plan/report].
Спочатку покажи scope/acceptance, root/branch/HEAD і relevant WIP,
власника запису, маршрут small/full/resume та pre-code READY/BLOCKED.
Використовуй наявні task artifacts і погоджений незалежний review route.
Збережи локальні gates, чужу роботу й межі дозволів; не роби auto-install,
зміну kit або запуск scanner. Далі виконуй саме погоджену задачу.
```

- **Small:** короткий plan/checkpoint у чинному record і перевірки за впливом;
  required gates не скасовуються. **Full:** застосувати поріг зі standards та один
  обраний процес, без паралельного дублювання spec-flow.
- До коду: scope/acceptance, рішення й невідоме, залежності/writer, scoped impact,
  застосовна planning critique/defense та READY/BLOCKED. Деталі —
  pre-code checkpoint (`ai-factory/pre-code-checkpoint.md` у kit; правила Courses — AGENTS.md).
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
  commands/exits і reviewer reference.

Завершення: кожен acceptance item має перевірений результат, застосовні gates та
незалежний verdict стосуються фінального snapshot, limitations і next action названі.
PASS/FAIL/BLOCKED/NOT_TESTED/N/A не підміняють одне одного; aggregate green або build
не доводять усі поведінкові рядки. [Testing](../docs/development/testing.md) і standards
визначають перевірки та side effects.
`make check`/`e2e-local` не є безпечними перевірками цього README: вони змінюють стан,
а важкі gates потребують погодженого вікна. Commit/push/deploy — лише в дозволеному scope.

## Якщо Factory недоступний

Назви відсутній шлях і BLOCKED саме залежної дії. Не роби auto-clone, install чи relink.
Локальні AGENTS/standards і task artifacts залишаються доступним джерелом для читання та
планування. Якщо required shared contract/review недоступний, не оголошуй його виконаним:
потрібне відновлення kit або окремо погоджений fallback. Відсутній optional CRG/Obsidian
допускає файловий fallback, без автоматичного refresh.
