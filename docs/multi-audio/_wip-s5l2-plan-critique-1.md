**S5-L2a: BLOCKED. S5-L2b: BLOCKED.** Відкрито 3 блокуючі прогалини плану.

### Ідентичність перевірки

- `stage: plan`; критик: Codex `/root`.
- Незалежні лінзи з чистим контекстом: `/root/slop`, `/root/logic`, `/root/complexity`, `/root/edge`. **Запущено 4 / отримано 4.**
- План: `docs/multi-audio/plan-s5l2.md`, SHA-256 `49c39db5027ffe2f134387bc4cea6bda54589a526ee9c60877bc0cdf6370b4f8`.
- Snapshot: `origin/main@7c6b754c3b25e4b1479da2ac1b67cf5432cf5ed5`; фактичний HEAD `132833afd50691491909dece38d5596a02c76c4c`. Tracked trees однакові: `98f9299021e0ac25819ff2461070e6af27d9d616`.
- Prompt: `/Users/strelia/project-data/ai-factory-kit/prompts/plan-critic.md`, version 1, SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`.
- Skill: `/Users/strelia/.agents/skills/plan/plan-critique/SKILL.md`, SHA-256 `540477ce2db1ffbf630b2b52a639945cdea9088adae70b419f2e017269416d11`.

### Перевірені джерела й межі

Прочитано AGENTS, briefing, workflow adapter, `work/tasks.md`, фазовий план §5/§8, попередній S5-L1-план, surface survey, сирі попередні reviews і відповідний report. Твердження звірено з frontend media domain/application/adapters/UI, composition і route, generated audio API, backend audio HTTP/domain/tus/persistence та media e2e/setup.

Reference зі skill за шляхом `~/.agents/ai-factory/reference-anti-patterns.md` відсутній; прочитано explicit fallback `/Users/strelia/project-data/ai-factory-kit/ai-factory/reference-anti-patterns.md`.

Поточні сторонні dirty/untracked файли не включено в продуктовий scope. Файли не змінювалися; тести й мережа не запускалися. Це оцінка плану, не runtime PASS.

### Знахідки

**1. BLOCKING — L2a/L2b: відео `awaiting_audio` не доходить до аудіосекції.**

**Де:** [plan-s5l2.md:34](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:34), декомпозиція й тести.

Backend після транскоду відео без звуку записує `awaiting_audio` — [publish.go:73](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/publish.go:73). Frontend `videoStates` цього значення не містить. `courseVideosSchema` використовує цей enum, тому адаптер відхиляє **весь список відео курсу** — [lesson-video-repository.ts:108](/Users/strelia/projects/courses/frontend/src/modules/media/adapters/lesson-video-repository.ts:108).

На свіжому відкритті редактор отримує порожні `rows`; автор не має аудіосекції, через яку міг би довести відео до Ready. Навіть після розширення enum чинний `shownVideo` не вибирає `awaiting_audio`.

**Правка плану:** включити в L2a domain/schema-підтримку цього стану; у L2b — вибір відповідного ассета, його відображення/i18n та сценарій **video-only → awaiting_audio → upload audio → Ready**. Наявний тест із уже Ready-відео прогалину не виявляє.

**Disposition:** actionable, відкрита. Лінза: logical gaps.

**2. BLOCKING — L2a/L2b: інвалідація пропускає завершення асинхронної обробки аудіо.**

**Де:** [plan-s5l2.md:43](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:43).

План називає інвалідацію course-video list після default/delete. Проте worker переводить ассет `awaiting_audio → ready` саме після промоуту першої аудіодоріжки — [audiotrack.go:419](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:419); невдала обробка може перевести його у Failed.

Чинне video polling працює лише для queued/processing або локального video transfer. Інвалідація після tus transfer відбувається раніше за завершення worker-а. Аудіодоріжка може вже показувати Ready, а відео залишатися в попередньому стані до стороннього refetch.

**Правка плану:** визначити синхронізацію course-video query після термінального результату обробки аудіо, включно з polling після reload. Перевірити оновлення відеостатусу без ручного перезавантаження, default або delete.

**Disposition:** actionable, відкрита. Лінза: logical gaps.

**3. BLOCKING — L2a/L2b: resume після reload не має визначеного контракту відновлення.**

**Де:** [plan-s5l2.md:37](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:37), [plan-s5l2.md:68](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:68).

Конфігурація transfer береться «з ticket-а», але після reload ticket відсутній. Audio list DTO не повертає upload URL, declared bytes чи ліміти — [handler.go:75](/Users/strelia/projects/courses/backend/internal/media/adapters/http/handler.go:75). Наявний video-resume натомість очікує `declaredBytes`.

Заплановані component/tus-тести можуть пройти з ticket у пам’яті, не довівши потрібний сценарій.

**Правка плану:** зафіксувати frontend-only механізм: відновити `/uploads/audio/{trackId}` і константи аудіо, перевірити вибраний файл за збереженим fingerprint, отримати та перевірити length/offset через HEAD. Визначити ціль також для cancel. Додати тест зі стартом **лише з list data + localStorage**, без ticket, та випадки відсутнього/невідповідного fingerprint.

**Disposition:** actionable, відкрита. Лінза: edge cases.

**4. NON-BLOCKING — L2b: delete-gating сформульовано ширше за backend-правило.**

**Де:** [plan-s5l2.md:48](/Users/strelia/projects/courses/docs/multi-audio/plan-s5l2.md:48).

«Вимкнено для останньої Ready» пропускає умову стану ассета. Backend забороняє таке видалення лише для **Ready-ассета** — [audiotrack.go:205](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:205).

**Правка плану:** уточнити guard через `asset.state === 'ready'`, Ready-стан цільової доріжки та відсутність іншої Ready. Це усуває розбіжність формулювання з контрактом; поточний звичайний promote атомарно переводить ассет у Ready, тому блокером не вважаю.

**Disposition:** actionable, відкрита. Лінза: edge cases.

### Підсумок лінз

| Лінза | Висновок |
|---|---|
| AI-slop | Clean: рішення прив’язані до реальних сценаріїв і owner decisions |
| Logical gaps | 2 blocking: `awaiting_audio`, асинхронна синхронізація |
| Unnecessary complexity | Clean: tus має два споживачі; використано чинні шари й Query |
| Edge cases | 1 blocking: reload-resume; 1 non-blocking: delete guard |

Це перший наданий verdict для цього S5-L2 snapshot. Попередні S5-L1 findings прочитано як контекст; їхні PASS і результати тестів на L2 не перенесено.

**Остаточно: S5-L2a — BLOCKED; S5-L2b — BLOCKED.** Для обох залишаються відкритими F1–F3; L2b також залежить від мерджу L2a. Потрібні правки плану та повторна критика зачепленого scope.