# Критика плану — раунд 3

`stage: plan`

**S5-L2a — ACCEPTED. S5-L2b — ACCEPTED. Відкритих BLOCKING: 0.**

## Ідентичність

- Критик: Codex `/root`, незалежний від автора плану.
- Лінзи з чистими контекстами: `/root/slop`, `/root/logic`, `/root/complexity`, `/root/edge`. **Запущено 4 / отримано 4.**
- План: `docs/multi-audio/plan-s5l2.md`; SHA-256 `fe514c1c2dbe118b9d1bd250d0e0eb6b9ead9b43a92bf43b1937179216400ace`.
- Target: `origin/main@7c6b754c3b25e4b1479da2ac1b67cf5432cf5ed5`; HEAD: `132833afd50691491909dece38d5596a02c76c4c`. Tracked trees однакові: `98f9299021e0ac25819ff2461070e6af27d9d616`.
- Prompt: `/Users/strelia/project-data/ai-factory-kit/prompts/plan-critic.md`, version 1; SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`.
- Skill: `/Users/strelia/.agents/skills/plan/plan-critique/SKILL.md`; SHA-256 `540477ce2db1ffbf630b2b52a639945cdea9088adae70b419f2e017269416d11`.

Усі три SHA-256 перевірено.

## Перевірені джерела й межі

Прочитано repository briefing, AGENTS/workflow adapter, фазовий план §5/§8, S5-L1, surface survey і сирі findings раундів 1–2. Shared anti-patterns прочитано з `/Users/strelia/project-data/ai-factory-kit/ai-factory/reference-anti-patterns.md`, оскільки шлях від skill відсутній.

Нову критику обмежено D-L2-3, його перевірками та взаємодіями. Твердження звірено з frontend media domain/schema/application/tus/UI, `admin-video-queries.ts`, route/`CourseEditorView`, generated audio API та backend HTTP/domain/tus/persistence. E2E/setup переглянуто як джерела запланованої перевірки.

Поточний план є untracked artifact із наведеним hash. Інші dirty/untracked файли, зокрема notes і `backend/upload`, не включено в продуктовий scope. Файли не змінювалися; тести й мережа не запускалися. Runtime-поведінку не підтверджено.

## Findings і статус попередніх F1–F4

### 1. F2 — resolved; попередній BLOCKING закрито в плані

**Де:** [plan-s5l2.md:60](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:60), D-L2-3; перевірки на рядках 71–73.  
**Disposition:** actionable → incorporated.

План явно встановлює: відсутній snapshot і непорожній термінальний підпис означають зміну. Отже, перша відповідь уже з `ready` або `failed` запускає перечитування course videos.

Це виправляє відмінність від чинного draft-зразка: [admin-video-queries.ts:87](/Users/strelia/projects/courses/frontend/src/app/composition/admin-video-queries.ts:87) пропускає відсутній previous. Backend атомарно записує Ready-доріжку та можливий Ready-ассет ([audiotrack.go:395](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:395), `:419`, commit `:467`); failed-шлях також може змінити ассет (`:502–525`).

**Вплив:** попередній сценарій застарілого `awaiting_audio` після reload тепер має визначений механізм усунення. Відсортований `trackId:state` усуває залежність від порядку відповіді; однаковий підпис не повторює інвалідацію. Audio key відокремлений від course-video key, а чинний video query не інвалідує audio.

Тести розрізняють перший `ready`, перший `failed`, `processing→ready` і повторну тотожну відповідь. **Додаткова правка плану не потрібна.**

### 2. F1 — resolved; попередній BLOCKING залишається закритим у плані

**Де:** [plan-s5l2.md:92](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:92), D-L2-8.  
**Disposition:** actionable → incorporated.

Backend записує `awaiting_audio` у `postgres/publish.go:73`; чинний frontend enum його не містить (`domain/video.ts:1`), strict schema використовує цей enum (`lesson-video-schema.ts:23`), а `shownVideo` не вибирає цей стан (`LessonVideoField.tsx:44`).

План охоплює domain/schema/i18n, вибір ассета та video-only E2E. **Вплив попереднього блокера усувається запланованими змінами; реалізація ще попереду.**

### 3. F3 — resolved; попередній BLOCKING залишається закритим у плані

**Де:** [plan-s5l2.md:39](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:39), D-L2-2.  
**Disposition:** actionable → incorporated.

Resume визначено без ticket: URL із trackId, audio-константи, fingerprint і перевірка HEAD length/offset. Backend повертає обидва заголовки (`tus/audioserver.go:57–58`) та підтримує DELETE тієї самої цілі (`:84`).

План включає missing/mismatched fingerprint і cancel без ticket. **Попередня невизначеність контракту відновлення закрита.**

### 4. F4 — resolved; попередній NON-BLOCKING залишається закритим у плані

**Де:** [plan-s5l2.md:75](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:75), D-L2-4.  
**Disposition:** actionable → incorporated.

Guard тепер враховує Ready-стан ассета, Ready-ціль і відсутність іншої Ready-доріжки. Для досяжних станів це узгоджується з `domain/audiotrack.go:205–223`; серверна помилка залишається захистом від гонок. **Попереднє надмірне обмеження видалення усунуто у формулюванні плану.**

## Чотири незалежні лінзи

| Лінза | Висновок щодо дельти |
|---|---|
| AI-slop | **CLEAN:** підпис і snapshot обслуговують конкретний сценарій першого terminal-read; перевірки мають різні спостережувані наслідки. |
| Logical gaps | **CLEAN:** початкове термінальне спостереження явно враховане; зв’язок worker → audio-read → video refetch визначено. |
| Unnecessary complexity | **CLEAN:** повторно використано Query snapshot та інвалідацію; нового store або додаткового polling-механізму немає. |
| Edge cases | **CLEAN:** враховано перший ready/failed, повтори, порядок рядків, порожній набір, remount/cache та атомарність backend-переходів. |

## Зміна verdict і остаточний результат

Після раунду 2 додано terminal-signature механізм і перевірки першого термінального спостереження та повторної відповіді. Це закриває залишок F2. Докази закриття F1/F3/F4 збережено після звірки з кодом; нових findings у зміненому scope немає.

- **S5-L2a — ACCEPTED.**
- **S5-L2b — ACCEPTED.**

Це прийняття плану на зазначеному snapshot. Залежність **L2b ← merge L2a**, незалежне implementation review та обов’язкові ordered gates залишаються чинними.