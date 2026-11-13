**S5-L2a — BLOCKED. S5-L2b — BLOCKED.** Залишився **1 відкритий BLOCKING**: F2 частково виправлено, але синхронізація після першого audio-read не визначена повністю.

## Ідентичність перевірки

- `stage: plan`; раунд 2; критик: Codex `/root`.
- Незалежні лінзи з чистим контекстом: `/root/slop`, `/root/logic`, `/root/complexity`, `/root/edge`. **Запущено 4 / отримано 4.**
- План: `docs/multi-audio/plan-s5l2.md`, SHA-256 `bb67040dcf0386e6b73b03eb789805d26697b02951d3350a2409096fb8a3372d`.
- Target: `origin/main@7c6b754c3b25e4b1479da2ac1b67cf5432cf5ed5`; фактичний HEAD: `132833afd50691491909dece38d5596a02c76c4c`. Обидва tracked trees: `98f9299021e0ac25819ff2461070e6af27d9d616`.
- Prompt: `/Users/strelia/project-data/ai-factory-kit/prompts/plan-critic.md`, version 1, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`.
- Skill: `/Users/strelia/.agents/skills/plan/plan-critique/SKILL.md`, SHA-256 `540477ce2db1ffbf630b2b52a639945cdea9088adae70b419f2e017269416d11`.

## Перевірені джерела й межі

Прочитано AGENTS, briefing, workflow adapter, фазовий план §5/§8, S5-L1, surface survey та сирі findings першого раунду. Перевірено відповідні frontend media domain/application/adapters/UI, composition, route і `CourseEditorView`, generated audio API, backend HTTP/domain/tus/persistence, media e2e/setup. Використано shared reference anti-patterns із `/Users/strelia/project-data/ai-factory-kit/ai-factory/reference-anti-patterns.md`.

Surface survey використано як набір тверджень для перевірки. Сторонні dirty/untracked файли й локальний `backend/upload` не включено в продуктовий scope.

Файли не змінювалися; тести й мережа не запускалися. Це verdict плану, не доказ працездатності реалізації.

## Знахідки та disposition першого раунду

### 1. F2 — частково виправлено; BLOCKING для L2a/L2b

**Де:** [plan-s5l2.md:57](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:57), D-L2-3; перевірки на рядках 62–63, 114–115.  
**Disposition:** actionable, open. Незалежно виявлено лінзами logical gaps і edge cases.

Інвалідація на спостереженому `processing→ready/failed` закриває основний сценарій. Проте перше спостереження вже термінальної доріжки не визначене:

1. Після reload course-video query повертає `awaiting_audio`.
2. Worker завершує обробку й атомарно переводить ассет у Ready — [audiotrack.go:419](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:419).
3. Перша відповідь audio-list уже містить `ready`; попереднього audio snapshot немає.

План вимагає інвалідацію «на переході», але не встановлює, що перше термінальне спостереження також її викликає. Названий зразок явно пропускає відсутнє попереднє значення — [admin-video-queries.ts:87](/Users/strelia/projects/courses/frontend/src/app/composition/admin-video-queries.ts:87). Audio polling зупиниться на термінальному списку; video polling не охоплює `awaiting_audio` — [admin-video-queries.ts:36](/Users/strelia/projects/courses/frontend/src/app/composition/admin-video-queries.ts:36).

**Вплив:** реалізація за наведеним зразком може залишити застарілий відеостатус до стороннього refetch. Обіцянка «після reload працює так само» не гарантована.

**Потрібна правка плану:** явно включити перше термінальне спостереження без попереднього audio cache у синхронізацію course videos. Визначити захист від повторної інвалідації на однакових відповідях.

Додати query-тест: cached video=`awaiting_audio`, previous audio=`undefined`, first audio response=`ready` → course videos перечитуються; повторна тотожна відповідь не створює цикл. Аналогічно покрити перший `failed`, оскільки worker може змінити й стан ассета.

### 2. F1 — resolved; попередній BLOCKING закрито в плані

**Де:** [plan-s5l2.md:82](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:82), D-L2-8.  
**Disposition:** actionable, incorporated.

Backend справді пише `awaiting_audio` — `postgres/publish.go:73`. Чинний frontend enum його не містить; strict schema відхиляє список, а `shownVideo` не вибирає цей стан — `domain/video.ts:1`, `lesson-video-schema.ts:23`, `LessonVideoField.tsx:44`.

План тепер прямо включає domain/schema/i18n у L2a, вибір і пояснення стану в L2b та video-only e2e. Попередню прогалину закрито на рівні запланованих змін.

### 3. F3 — resolved; попередній BLOCKING закрито в плані

**Де:** [plan-s5l2.md:39](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:39), D-L2-2.  
**Disposition:** actionable, incorporated.

Визначено URL і константи з trackId, fingerprint, перевірку HEAD length/offset, missing/mismatched fingerprint та cancel без ticket. Backend HEAD повертає потрібні `Upload-Length` і `Upload-Offset` — `tus/audioserver.go:51`; DELETE підтримується на тій самій цілі — `:84`.

Додані перевірки стартують без ticket. Змінювати audio-list DTO для цього механізму не потрібно.

### 4. F4 — resolved; попередній NON-BLOCKING закрито в плані

**Де:** [plan-s5l2.md:65](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:65), D-L2-4.  
**Disposition:** actionable, incorporated.

Guard тепер враховує Ready-стан ассета, Ready-ціль і відсутність іншої Ready-доріжки. Для досяжних станів це узгоджується з `domain/audiotrack.go:205`. Component-перевірку guard включено.

## Підсумок лінз

| Лінза | Результат |
|---|---|
| AI-slop | CLEAN: змінені рішення мають конкретних споживачів і перевірювані результати |
| Logical gaps | 1 BLOCKING: початкова синхронізація без попереднього audio snapshot |
| Unnecessary complexity | CLEAN: повторно використано tus і Query; нових зайвих механізмів немає |
| Edge cases | Той самий BLOCKING; F3/F4 закрито |

Порівняно з раундом 1 закрито F1, F3, F4; F2 звужено до першого термінального спостереження. Незмінені висновки AI-slop і complexity збережено після повторної перевірки зміненого scope.

**Остаточно: S5-L2a — BLOCKED; S5-L2b — BLOCKED.** Єдиний відкритий блокер належить query-механізму L2a та впливає на обіцяне UI-оновлення L2b.