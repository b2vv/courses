ДАЛІ: прочитати work/brief-t11.md цілком, репозиторний брифінг, план поточної фази, список задач і застосовні інструкції implement; визначити точний обсяг T11 та порядок перевірки.
ФАКТ: work/brief-t11.md прочитано цілком. T11 має довести B24a (видимий фокус і геометрія ≥44×44 для кнопки play/pause, повзунка позиції та hls.js-вибору якості), B25 (axe WCAG 2 AA без мовчазного відкидання incomplete; світла/темна теми; 200% із конкретними layout-критеріями) і R35 (автоперехід hls.js на нижчу сходинку через наявний CDP-дросель). Нативна HLS-гілка в цьому Playwright-стенді свідомо недоведена; Hls.isSupported() підміняти заборонено. Єдина дозволена зміна продукту — короткий коментар про несучий порядок умов у hls-media.ts.
РІШЕННЯ: виконувати T11 переважно в frontend/e2e/accessibility.spec.ts, не чіпати AV1/T12, 403 і hls.stopLoad(), generated, docs/playback/*, work/tasks.md, AGENTS.md чи залежності. workers лишити 1. Локалі змінювати лише парно, якщо взагалі знадобляться.
ДАЛІ: дочитати без усічення всі обовʼязкові джерела з §1 брифа та точні секції §8 плану/spec, зелений гейт T10, backlog-міни й поточний accessibility.spec.ts; перевірити робоче дерево перед змінами.
ФАКТ: implement вимагає red→green на погодженому публічному шві та фінальне code-review. Погоджений брифом шов — реальний браузер у frontend/e2e/accessibility.spec.ts; приймальні рядки B24a, B25 і R35 вже явно призначені цьому рівню. Робоче дерево до моїх змін чисте, крім нового work/report-t11.md. У accessibility.spec.ts вже є чернетковий B24a/B25-тест: він знаходить три названі контроли, міряє ≥44px, таб-порядок, лише focus ring quality, відсутність overlap і горизонтального scroll; але contrast-перевірка все ще викидає axe incomplete, не перевіряє обрізаний текст/вихід кожного контролу за контейнер і охоплює теми несиметрично відносно 200%.
РІШЕННЯ: вважати браузерний e2e єдиним тестовим seam для нових доказів; не додавати unit-підміни layout/Hls.isSupported(). Розвивати існуючий тест вертикальними зрізами: спочатку чесний axe-result, потім точна геометрія/фокус/zoom у двох темах, потім R35 через справжній hls.js і CDP.
ДАЛІ: знайти точний hls-media.ts, дочитати хвіст accessibility.spec.ts, релевантні хелпери playback.spec.ts і всі обовʼязкові джерела без усічення; окремо перевірити, чи існує CONTEXT.md.
ФАКТ: прочитано docs/playback/plan.md:1–440 і видимі частини 441–842. План закріплює hls.js як MediaSource-гілку з ручною якістю й точними HTTP-статусами, а нативну Safari-гілку — без ручної якості та без статусів сегментів (D-7/D-9). Приймальні B24a/B25/R35 належать T11; R35 вимагає саме автоматичний вибір у hls.js. Ризик-реєстр прямо визнає непаритет нативної гілки.
РІШЕННЯ: не видавати WebKit за Safari і не розширювати скоуп. R35 будувати на тому самому продакшн-шляху hls.js, на реальному master із двома сходинками та на браузерному CDP-дроселі; геометрію й контраст доводити після реального рендеру, не за CSS-класами.
ДАЛІ: дочитати plan.md дрібнішими діапазонами, бо сукупний вивід інструмента був усічений; потім прочитати spec.md, CTO-RESEARCH.md і work/tasks.md цілком окремими командами.
ФАКТ: docs/playback/plan.md дочитано по діапазонах, включно з повними рядками B24/B24a/B25/B26 і R35, ризиками, T11 та D-9. docs/playback/spec.md прочитано цілком: A12 вимагає автоматичну й ручну якість, A15 — клавіатуру, видимий фокус, ≥44px, світлу/темну теми, 200% і uk/en. T11 у work/tasks.md досі planned, залежність T8 виконана; T9/T10 мають повні зелені make check.
РІШЕННЯ: B24 уже належить попередньому unit+axe доказу; T11 не дублює його, а закриває браузерні B24a/B25 та hls.js R35. B26 локалізаційний parity-гейт уже належить T8; T11 має перевірити фактичну uk/en доступність плеєрних контролів у браузері без вигадування нових ключів, якщо чинні підписи достатні.
ДАЛІ: прочитати work/CTO-RESEARCH.md і work/tasks.md без усічення, точні backlog B-8/B-9/B-11/B-13/B-16/B-17, повний T10-звіт і AGENTS.md; потім інвентаризувати код, хелпери й fixture-шлях R35.
ФАКТ: CTO-RESEARCH.md, work/tasks.md, AGENTS.md, потрібні B-8/B-9/B-11/B-13/B-16/B-17 і work/report-t10.md прочитані. AGENTS.md актуальніший за старий CTO-кеш: playback, outline, watched marks і progress уже працюють; T1–T10 виміряні make check 59/59. B-8/B-9 вимагають мислити фактичним станом у момент тесту; B-11 — module-level heredoc; B-13 — workers:1; B-16 — три тихі заміри перед будь-якою зміною timeout; B-17 — не використовувати git add -A під час паралельної роботи.
РІШЕННЯ: git не змінювати: бриф T11 прямо забороняє чіпати git, тому це сильніше за загальну вимогу implement про коміт. Docker-залежні гейти не запускати й не називати пройденими. CTO-RESEARCH.md не переписувати попри застарілі твердження, бо T11 прямо забороняє docs/playback/*, work/tasks.md/AGENTS.md, а briefing не є продукт-кодом T11.
ДАЛІ: інвентаризувати реальні шляхи hls-media, плеєр, CSS, локалі, Playwright config/setup/teardown і CDP-дросель у playback.spec.ts; встановити, чи поточна accessibility-фікстура створює справжню hls.js-гілку чи лише макет.
ФАКТ: createPlaybackMedia справді перевіряє Hls.isSupported() перед canPlayType; Chromium/WebKit через Playwright тому обирають hls.js. Чинний B24a/B25 layout-тест маршрутизує тільки learning read і не запускає play: hls.js-конструктор монтується, тому quality trigger існує, але медіа не мінтиться й сходинки не завантажуються. Це придатний реальний-render seam для геометрії, але не для R35.
ФАКТ: реальний двосходинковий/багатосходинковий HLS створюється production upload→worker→publish шляхом усередині frontend/e2e/playback.spec.ts; там уже є openAndPlay, 5-хвилинний asset, request observers і CDP Network.emulateNetworkConditions. playback.teardown.ts після цього проєкту дренує чергу й повертає 30/180. workers:1 незмінний.
РІШЕННЯ: B24a/B25 залишити в accessibility.spec.ts як ізольований browser-layout доказ; R35 додати до чинного serial playback.spec.ts, де доступний production-origin master і не виникає ще одного транскоду/проєкту. Це дотримується «accessibility.spec.ts — основний файл» та прямої вказівки переюзати CDP-механізм з playback.spec.ts без дублювання fixture machinery.
ДАЛІ: прочитати точну форму HLS master/ladder і локалізаційний browser-flow; сформулювати незалежні observables для axe incomplete, layout/text clipping/focus усіх трьох контролів у двох темах і автоматичного 720p→360p переходу без user quality action.
ФАКТ: production fixture з 1280×720 source дає 360p і 720p HLS-renditions; P09 уже вручну обирає 360p. CSS дає кожному control мінімум 2.75rem, seek окремий :focus-visible outline 3px, Button/Select — спільний 3px focus token; на ≤42rem control grid стає одноколонковим. uk/en вже мають парні ключі play/pause/position/quality/qualityAuto.
РІШЕННЯ: B24a oracle для кожного з трьох конкретних контролів: послідовний Tab-фокус; computed outline style≠none, width≥3, непрозорий колір, відмінний від фону; bounding box width≥44 і height≥44. B25 oracle у кожній темі після 200%: axe violations порожні; axe incomplete дорівнює зафіксованому точному allowlist; document без horizontal scroll; усі три control rectangles усередині `.lesson-player-controls`; жоден видимий текстовий елемент у controls не має scrollWidth/Height більших за clientWidth/Height; controls не перетинаються.
РІШЕННЯ: uk/en browser oracle — українські Play/Position/Quality доступні, перемикач EN змінює їх на Play/Playback position/Quality і `html[lang=en]`; нових locale keys не потрібно. R35 oracle: на auto (жодного вибору quality) спочатку спостерігається 720p segment request на нормальному каналі, потім після CDP-дроселя й seek у незавантажену ділянку — 360p request; це червоніє від fixed-high/manual-only implementation.
ДАЛІ: внести перший тестовий зріз: чесний axe-result із violations та incomplete allowlist і симетричний B24a/B25 layout helper для light/dark на 200%; потім запустити локальні format/lint/typecheck.
ФАКТ: перший test-only зріз внесено в frontend/e2e/accessibility.spec.ts. Обидва axe-прогони тепер повертають точні `{id, targets}` для violations і incomplete та вимагають порожні обидва списки. B24a перевіряє всі три названі контроли окремо; B25 перевіряє light і dark уже при 200%, containment, відсутність overlap/page overflow/text clipping; browser-flow також перевіряє українські й англійські accessible names та `html lang`.
РІШЕННЯ: очікуваний axe incomplete allowlist наразі порожній. Контроли в актуальній верстці не лежать поверх video, а на окремій непрозорій semantic surface; якщо runtime покаже неминучий incomplete, reviewer має зафіксувати лише виміряний exact id+selector, не послаблювати перевірку.
ДАЛІ: запустити frontend format-check, lint і typecheck; виправити лише статичні помилки цього зрізу, не форматуючи чужі файли масово.
ФАКТ: `cd frontend && npm run format:check` — RED, exit 1; єдина названа проблема форматування — e2e/accessibility.spec.ts. Інших файлів formatter не назвав.
РІШЕННЯ: застосувати repo formatter лише до frontend/e2e/accessibility.spec.ts як механічну правку, потім повторити format-check перед lint/typecheck.
ДАЛІ: відформатувати один змінений test file і повторити format-check.
ФАКТ: `npx oxfmt e2e/accessibility.spec.ts` змінив лише цільовий test file; повторний `npm run format:check` — GREEN, exit 0, 162 файли.
ДАЛІ: запустити `npm run lint` для архітектурного й статичного гейта першого зрізу.
ФАКТ: `cd frontend && npm run lint` — GREEN, exit 0; oxlint без warnings, architecture boundaries checked 90 files.
ДАЛІ: запустити `npm run typecheck`; особливо перевірити Locator/evaluate та axe result typings.
ФАКТ: `cd frontend && npm run typecheck` — GREEN, exit 0; нові Locator/Page helpers і повні axe findings типізуються без assertions/any.
ДАЛІ: додати R35 у наявний serial playback fixture: спостерігати auto 720p на нормальному каналі, увімкнути CDP-дросель, змусити запит у незавантажену частину й вимагати новий 360p request без взаємодії з quality control; гарантовано відновити мережу в finally.
ФАКТ: R35 додано в frontend/e2e/playback.spec.ts одразу після P01/P09/P10, усередині чинного student serial block із production upload→worker→publish fixture. Test записує лише `.m4s` requests, підтверджує `Автоматично`, чекає 720p, фіксує baseline 360p, вмикає той самий CDP Network.emulateNetworkConditions (100ms/1024 Bps), seek-ає на 150s і вимагає новий 360p request; мережа відновлюється й CDP detach виконується у finally.
ФАКТ: у frontend/src/adapters/playback/hls-media.ts додано короткий дозволений коментар: порядок Hls.isSupported() перед native canPlayType несучий, бо Chromium брехливо повертає `maybe`, а hls.js дає робоче playback, quality selection і HTTP-status errors.
РІШЕННЯ: R35 не клікає quality і не підміняє Hls.isSupported(); видимий hls.js quality control та значення Auto є branch/selection oracle. Нативна гілка лишається недоведеною в цьому стенді за D-9.
ДАЛІ: прогнати format-check; якщо RED лише на двох змінених code/test files, форматувати їх точково й повторити, потім lint/typecheck/test.
ФАКТ: `cd frontend && npm run format:check` після R35 — RED, exit 1; formatter назвав лише e2e/playback.spec.ts, hls-media.ts і попередній accessibility.spec.ts уже форматні.
РІШЕННЯ: відформатувати лише e2e/playback.spec.ts і повторити повний frontend format-check.
ДАЛІ: виконати точковий oxfmt для playback.spec.ts і повторний format-check.
ФАКТ: `npx oxfmt e2e/playback.spec.ts` змінив лише R35 test file; повторний `npm run format:check` — GREEN, exit 0, 162 файли.
ДАЛІ: запустити lint/architecture на повній frontend-кодобазі.
ФАКТ: `cd frontend && npm run lint` після R35/comment — GREEN, exit 0; oxlint без warnings, 90 architecture-boundary files checked.
ДАЛІ: запустити typecheck для R35 CDP payload, request observers і hls comment scope.
ФАКТ: `cd frontend && npm run typecheck` після повної T11-правки — GREEN, exit 0.
ДАЛІ: запустити дозволений брифом повний `npm test`; browser e2e не запускати без Docker.
ФАКТ: `cd frontend && npm test` — GREEN, exit 0; 44 files, 373/373 tests passed, 0 failed/skipped/todo, snapshots unchanged, 5.607s total.
ДАЛІ: виконати лише Playwright discovery (`--list`) для chromium accessibility та playback R35, не запускаючи браузерні сценарії; перевірити, що нові labels потрапили у правильні проєкти.
ФАКТ: Playwright discovery — GREEN, exit 0. Chromium chain discover-ить B24a/B25 у accessibility.spec.ts; playback chain discover-ить новий R35 у playback.spec.ts:604. `--project=playback --list` показує 56 tests у 13 files разом із залежностями/teardown. Це лише discovery, не runtime e2e PASS.
ДАЛІ: застосувати обовʼязковий implement-етап code-review: прочитати скіл повністю, окремо перевірити Standards і Specification, а також причинність пʼяти питань; не запускати Docker-гейти.
ФАКТ: code-review skill прочитано цілком. Він вимагає два незалежні паралельні огляди. Зміни T11 навмисно незакомічені, бо бриф забороняє чіпати git; тому фіксована база огляду — поточний HEAD, а review surface — `git diff HEAD -- frontend/e2e/accessibility.spec.ts frontend/e2e/playback.spec.ts frontend/src/adapters/playback/hls-media.ts work/report-t11.md` плюс untracked report як окремий артефакт.
РІШЕННЯ: Standards source — AGENTS.md плюс Fowler smell baseline зі скіла; Specification source — work/brief-t11.md, docs/playback/plan.md §8 і docs/playback/spec.md A12/A15. Обидва огляди read-only; жоден агент не редагує файли.
ДАЛІ: запустити паралельно Standards і Spec review; після повернення перевірити кожну знахідку локально й виправити лише підтверджені blocking проблеми.
ФАКТ: під час власного review R35 посилено до причинного доказу: test тепер читає server-minted mediaPath і вимагає production master рівно з `360p/index.m3u8` та `720p/index.m3u8`; контрольний seek на 100s при нормальному каналі мусить породити новий 720p request, і лише після цього throttle+seek 150s мусить породити новий 360p request.
РІШЕННЯ: пара normal-channel 720p → throttled 360p відсікає false green, де lower request був лише стартовим вибором або наслідком seek незалежно від каналу. Exact two-rung master окремо пінує передумову R35.
ДАЛІ: дочекатися обох незалежних review; паралельно повторити format-check/lint/typecheck після causal strengthening.
ФАКТ: format-check після causal strengthening — RED, exit 1; лише e2e/playback.spec.ts потребує механічного форматування.
РІШЕННЯ: точково відформатувати playback.spec.ts, не торкаючись інших файлів.
ДАЛІ: виконати точковий oxfmt і повторний format-check.
ФАКТ: точковий oxfmt завершено; повторний frontend format-check — GREEN, exit 0, 162 файли.
ДАЛІ: послідовно повторити lint, typecheck і npm test після остаточного R35 strengthening.
ФАКТ: фінальний локальний allowed chain після R35 strengthening — GREEN: `npm run lint` exit 0 (90 architecture files), `npm run typecheck` exit 0, `npm test` exit 0 (44 files, 373/373 passed, 0 skipped/todo, 5.832s).
ДАЛІ: отримати обидва незалежні review verdicts і перевірити кожну знахідку проти коду/брифа.
ФАКТ: незалежний Standards review знайшов два HIGH: ≥44×44 перевірялось лише після 200% (22px при 100% могло б хибно пройти), а dark-прогін не мав oracle, що resolved theme справді dark. LOW judgement call — дубль мапінгу axe findings. Незалежний Spec review знайшов HIGH лише у ще не завершеному звіті (немає фінальної пʼятипитальної таблиці/незапущених гейтів/команд) і MEDIUM — треба дослівно записати замір native limitation для обох рушіїв.
РІШЕННЯ: обидва Standards HIGH підтверджені й блокують завершення. Додати 100%-геометрію до zoom, а кожну тему пінувати resolved `colorScheme` та semantic surface background. Axe-дубль прибрати спільним browser-evaluate helper без нової абстракції продукту. Report findings закрити фінальними append-блоками після коду й гейтів.
ДАЛІ: виправити accessibility test: спільний axeAudit helper; до zoom перевірити 44×44; у light/dark вимагати `getComputedStyle(document.documentElement).colorScheme` і різні виміряні background colors; потім повторити локальні гейти.
ФАКТ: обидва Standards HIGH виправлено в accessibility.spec.ts. `expectPlaybackLayout` тепер викликається до 200% і тому пінує фізичні 44×44 при 100%; після zoom light runtime мусить мати resolved `colorScheme: light`, dark — `colorScheme: dark`, а semantic controls surface мусить реально змінити background. Дубль axe mapping замінено одним test helper `axeAudit(page, 'wcag'|'contrast')`.
РІШЕННЯ: 200%-layout checks залишено окремо в обох темах; 100%-перевірка не підміняє zoom evidence. Theme oracle не пінує конкретний hex і дозволяє редизайн, але червоніє, якщо dark media rule або surface tokens перестають застосовуватись.
ДАЛІ: запустити format-check, потім повний дозволений lint→typecheck→test chain; після цього повторно перевірити diff і review findings.
ФАКТ: post-review `npm run format:check` — GREEN, exit 0, 162 файли; review-fix уже відповідає formatter.
ДАЛІ: виконати lint, typecheck, npm test у канонічному дозволеному порядку.
ФАКТ: post-review allowed chain — GREEN: lint/architecture exit 0 (90 files), typecheck exit 0, npm test exit 0 (44 files, 373/373 passed, 0 failed/skipped/todo, 6.375s).
ДАЛІ: попросити обидва review-канали повторно перевірити виправлені HIGH і знайти залишкові blockers; паралельно локально перевірити diff scope/whitespace.
ФАКТ: локальний diff audit — `git diff --check` GREEN; змінені лише frontend/e2e/accessibility.spec.ts, frontend/e2e/playback.spec.ts, frontend/src/adapters/playback/hls-media.ts та append-only work/report-t11.md. workers:1 незмінний; forbidden docs/playback/*, work/tasks.md, AGENTS.md, generated і dependencies не змінені.
ФАКТ: повторний Standards review — PASS: обидва HIGH і axe duplication закриті; залишкових documented-standard blockers або серйозних Fowler smells немає. Повторний Spec review — PASS: B24a/B25/R35 причинні й буквальні; єдина product-code зміна — дозволений hls-media comment.
РІШЕННЯ: code-review gate закрито після одного fix round. Standards: 0 findings. Specification code/test delta: 0 findings. Початкові report-only знахідки закриваються наступним фінальним append, не переписуванням історії.
ДАЛІ: повторити Playwright discovery після остаточних правок, потім дописати exact native limitation, таблицю §8 з пʼятьма відповідями, конкретні контроли, незапущені гейти й reviewer commands.
ФАКТ: фінальний Playwright discovery `--project=chromium --project=playback --grep 'B24a B25|R35 hls\.js' --list` — GREEN, exit 0; discover-ить B24a/B25 в accessibility.spec.ts:256 і R35 в playback.spec.ts:604, разом із належним setup/media/playback-profile/teardown dependency chain. Це не runtime browser PASS.
ДАЛІ: зафіксувати обовʼязковий фінальний доказовий блок і точні команди рецензента.

## Конкретні контроли B24a

1. Одна кнопка відтворення/паузи — виміряна у стані `Відтворити`/`Play`; та сама кнопка змінює дію й назву на pause після старту.
2. Повзунок позиції — `Позиція відтворення` / `Playback position`.
3. hls.js-вибір якості — `Якість: Автоматично` / `Quality: Auto`; у native branch цього контролу за дизайном немає.

Для кожного з трьох test перевіряє окремо `width >= 44` і `height >= 44` при 100%, клавіатурне фокусування і видимий outline; при 200% повторно перевіряє containment/layout у light і dark.

## Native HLS limitation — точний замір

Нативна гілка лишається недоведеною в цьому стенді. В обох Playwright-рушіях — Chromium і WebKit — замір був `{hlsJsSupported:true, mse:true, canPlayType:"maybe"}`. Через несучий порядок `Hls.isSupported()` першим обидва входять у hls.js; Playwright WebKit — не Safari. Підміни `Hls.isSupported()` немає. Справжній Safari потрібен для native-branch доказу; це свідоме обмеження D-9, не заявлений PASS.

## §8: рядок → чим доводиться, пʼять питань

| Рядок | 1. Що саме доводить | 2. Production artifact/path | 3. Як виключено false green | 4. Чому саме цей рядок | 5. Що зробить test red |
|---|---|---|---|---|---|
| B24a | Реальний Chromium рендерить кнопку play/pause, position slider і hls.js quality trigger не меншими за 44×44 при 100%; Tab/Shift+Tab досягає кожного, computed outline видимий у light і dark. | Компонент LessonPlayer і UI kit рендеряться справжнім SSR/browser path; лише learning read замінено валідною course fixture, бо row міряє layout, а не походження media bytes. Hls.isSupported не підмінено. | 44px перевіряються до zoom, тому 22px не подвояться до хибного PASS; кожен control має власні size/focus assertions, а не груповий текст. | B24 уже доводить keyboard actions unit+axe; цей test додає рівно браузерну геометрію та focus visibility, яких jsdom не має. | Будь-який control <44px у width або height, пропущений у tab order, outline none/<3px/transparent/того ж кольору, або зниклий quality trigger зробить test red. |
| B25 | При 200% у фактично light і фактично dark темах: axe `color-contrast` має порожні violations і incomplete; page не має horizontal scroll; кожен control у container; sibling controls не overlap; direct text не clipped. English names також поміщаються, `html lang=en`. | Ті самі product CSS semantic tokens, responsive media query, React Aria controls і браузерний layout engine; test не перевіряє наявність CSS-класу замість computed behavior. | Resolved `colorScheme` і зміна computed surface доводять, що dark справді застосувався; axe incomplete не відкидається; containment/scroll/overlap/clipping — незалежні observables. | Це точне розгортання нечіткого «layout не ламається» B25 у пораховані browser criteria та WCAG AA rule, а не повтор B24a. | Зникла dark rule/token, новий axe violation/incomplete selector, overflow за viewport/container, overlap чи clipped uk/en text зробить test red. |
| R35 | На hls.js Auto production master спочатку робить новий 720p `.m4s` request після control seek при нормальному каналі, а після CDP throttle та такого ж seek — новий 360p request без quality action. | Server-minted mediaPath веде до master, створеного справжнім admin upload → worker ffmpeg 5.1.9 → publish шляхом; test пінує рівно `360p/index.m3u8` і `720p/index.m3u8`. | Baseline counts не дають зарахувати старий startup 360p; normal-channel control seek мусить лишитись 720p; Auto перевіряється до й після; Hls.isSupported не mock-иться. | P09 уже доводить ручний rung select; R35 окремо доводить автоматичне channel-driven пониження hls.js. Native limitation названа вище. | Fixed-high/manual-only ABR, відсутня 720/360 rung, selection не Auto, seek-only downgrade або ігнорований CDP throttle зробить test red. |

## Незапущені гейти

Через заявлену брифом відсутність Docker НЕ ЗАПУСКАЛИСЯ і не позначені passed: `make integration`, `make check-generated`, `make build`, `make up`, `make smoke`, `make e2e`, `make e2e-local`, повний `make check`. Відповідно runtime B24a/B25/R35, actual axe incomplete list і ABR transition ще authored/discovered, але не виміряні в цій сесії.

## Команди для reviewer

Локальні без Docker, у перевіреному порядку:

```sh
cd /Users/strelia/projects/courses/frontend
npm run format:check
npm run lint
npm run typecheck
npm test
```

Повний канонічний гейт на Docker-capable host:

```sh
cd /Users/strelia/projects/courses
make check
```

Точковий browser replay після підняття/seed середовища через правильну локальну ціль:

```sh
cd /Users/strelia/projects/courses
make up
make seed-identities
make seed
cd frontend
npx playwright test --project=chromium --grep 'B24a B25' --reporter=line
npx playwright test --project=playback --grep 'R35 hls\.js' --reporter=line
```

ФАКТ: обовʼязкова таблиця містить усі пʼять відповідей для B24a, B25 і R35; окремо названі три вимірювані контроли, точний native limitation, незапущені Docker-гейти і reviewer commands.
РІШЕННЯ: T11 code/test scope завершено локально. Доступні гейти GREEN: format-check, lint/architecture, typecheck, 373/373 unit tests, Playwright discovery; Standards review PASS, Specification review PASS. Runtime e2e не оголошено passed.
ДАЛІ: Docker-capable reviewer запускає `make check`; лише після runtime GREEN B24a/B25/R35 можна називати виміряними. T12 AV1 лишається окремим паралельним таском, T13 — фазовим гейтом після T11/T12.
ПИТАННЯ: немає; scope, limitation і наступний verification step визначені брифом.
ФАКТ: фінальний audit work/report-t11.md пройдено: 142 рядки до цього запису; перший рядок — обовʼязковий `ДАЛІ:`, історія кроків збережена append-only, фінал містить усі обовʼязкові докази/обмеження/команди. `git diff --check` GREEN; change scope рівно три code/test files плюс цей журнал.
РІШЕННЯ: роботу завершено без коміту відповідно до прямої заборони брифа чіпати git; секрети, media bytes, dependency caches і заборонені шляхи не додані.
ДАЛІ: Повністю прочитати базовий і дельта-бриф T11, обов’язкові репозиторні документи та інструкції implement, потім звірити вимоги з поточним кодом і тестами.
ФАКТ: Skill implement вимагає працювати за специфікацією, регулярно запускати вузькі тести й typecheck, наприкінці виконати повну перевірку, code review та зробити логічний коміт.
ДАЛІ: Прочитати обидва брифи T11 цілком, не підміняючи дельтою чинні вимоги базового брифу.
ФАКТ: Базовий бриф вимагає доказів B24a/B25/R35, точного обліку axe incomplete, незмінності продуктового коду крім уже дозволеного коментаря та чесного переліку недоведеного. Дельта фіксує два корені: один очікуваний nonBmp incomplete від шеврона Select і недосяжний R35 через дросель O1 у 1 КБ/с; рішення й заборонені альтернативи задані буквально.
ДАЛІ: Дочитати решту дельта-брифу, зокрема вимоги до перевірок, review і звіту, потім прочитати чинний report та обов’язкові план/spec/backlog матеріали.
ФАКТ: Дельта прочитана повністю. Docker і браузерний доказ у цьому sandbox недоступні; B24a/B25/R35 заборонено називати зеленими. Дозволені гейти: frontend format:check/lint/typecheck/test і Go-тести без Docker. Фінальний append має дати п’ять відповідей на кожен рядок, числові очікування браузерного прогону, точні команди рецензента й спростування передчасного PASS попереднього проходу.
РІШЕННЯ: Не запускати Docker-залежні цілі й не торкатися T12 або заборонених product-файлів; браузерні виміри оформити як очікування для рецензента, а не як власний результат.
ДАЛІ: Прочитати поточний журнал, рев’ю попереднього прогону, релевантні секції plan/spec/backlog/tasks і CTO briefing, а потім інвентаризувати робоче дерево.
ФАКТ: Інвентаризація обов’язкових джерел локалізувала §8 plan (B24a рядок 592, B25 593, R35 655), spec A12/A15, T11 у task tracker, backlog B-8/B-9/B-11/B-13/B-16/B-17, секцію зеленого гейта T10 і 157-рядковий review першого Docker-прогону.
ДАЛІ: Прочитати ці джерела без урізання потрібного контексту, включно з усім CTO briefing, поточним report і review.
ФАКТ: CTO briefing прочитано повністю. Він є кешем стану до playback і в частині «програвання не існує» застарів відносно чинного коду та AGENTS.md; актуальна архітектурна межа лишається незмінною: медіа-байти віддає Nginx, не Go/SSR, а browser e2e є єдиним доказом layout/ABR.
РІШЕННЯ: Для T11 спиратися на актуальний код, plan/spec і фазові звіти; не переписувати CTO briefing, бо це поза буквальним scope дельти.
ДАЛІ: Прочитати чинний work/report-t11.md і Docker-review повністю, щоб не перезаписати історію та не повторити спростований діагноз.
ФАКТ: Docker-review прочитано: перший make check дійшов до e2e й впав на єдиному nonBmp incomplete; залежний playback тоді не запустився. Окремий R35 прогін теж red: 360p count лишився 1 після 60 с. Review сам виправляє свій ранній хибний висновок про seek: дефолтний asset 300 с, справжня проблема — 8192 біт/с проти master BANDWIDTH 542300/3440800.
ФАКТ: Поточний report підтверджує передчасний запис «T11 code/test scope завершено… Specification review PASS» і містить попередні test-only зміни, локальні зелені гейти та недоведений runtime. Сукупний вивід урізав середину журналу, тому повне читання ще не завершене.
ДАЛІ: Дочитати work/report-t11.md окремими діапазонами, потім перейти до нормативних секцій plan/spec/tasks/backlog.
ФАКТ: Повторне читання report підтвердило його append-only структуру й повний попередній доказовий блок. Видно попередні causal strengthening R35, review-виправлення 44×44 при 100% і resolved theme oracle, але один короткий середній діапазон знову урізаний виводом.
ДАЛІ: Прочитати лише урізаний діапазон журналу з номерами рядків, тоді завершити читання нормативних джерел.
ФАКТ: work/report-t11.md тепер прочитано повністю. Попередній review уже ловив і виправляв false-green ризики 44px після zoom та непідтвердженої dark theme; поточна дельта має замінити лише хибні припущення про порожній incomplete і дросель, не ламати ці докази.
ДАЛІ: Прочитати plan §8 з п’ятьма питаннями, spec A12/A15, task tracker, потрібні backlog-міни та зелений гейт T10.
ФАКТ: Нормативні plan/spec секції прочитані. П’ять питань вимагають назвати точний доказ, production-origin артефакта, false-green захист, власника рядка й конкретний red trigger. A12 вимагає auto та manual quality; A15 — keyboard, visible focus, ≥44px, themes, 200%, uk/en. D-7/D-9 свідомо не дають native HLS паритету зі hls.js.
ДАЛІ: Прочитати task tracker, backlog B-8/B-9/B-11/B-13/B-16/B-17 і T10 green gate, після чого звірити diff та поточний код точними рядками.
ФАКТ: Task tracker і мінне поле прочитані. T11 залежить від виконаного T8; прогрес і черга є спільним станом, heredoc лишаються module-level, Chromium workers=1 через shared catalog, таймаут не змінюється без триразового тихого заміру, git staging має бути лише явними шляхами. T10 green gate зафіксував make check EXIT=0, e2e 59/59 і O1 із навмисним 1 КБ/с для завислого сегмента — ці числа не можна переносити в ABR-доказ R35.
ДАЛІ: Перевірити робоче дерево та точний diff, потім прочитати цільові test/code фрагменти й transcoder BANDWIDTH origin.
ФАКТ: Робоче дерево містить паралельну T12-дельту в заборонених для T11 backend/OpenAPI/generated/API/composition файлах; їх не чіпаю. T11 surface лишається accessibility.spec.ts, playback.spec.ts, уже дозволений коментар hls-media.ts і append-only report. Поточний B25 досі вимагає incomplete=[], а R35 досі використовує 1024 Б/с і не має post-throttle high baseline/diagnostic.
РІШЕННЯ: Редагувати лише два e2e-файли та журнал; hls-media.ts уже відповідає базовому брифу й не потребує дельта-правок.
ДАЛІ: Прочитати точні реалізації axeAudit/B24a-B25/R35/O1, master parser context і ffmpeg master-generation code перед патчем.
ФАКТ: Цільовий код підтверджує точні seams дельти. axeAudit агрегує targets на rule-рівні й губить node check data; B25 робить два однакові порожні incomplete asserts. R35 читає master один раз, але відкидає `#EXT-X-STREAM-INF`; O1 демонструє безумовний diagnostic до assert. ffmpeg генерує master сам, а `-maxrate` 365k/3000k не є BANDWIDTH, за яким ухвалює рішення hls.js.
ДАЛІ: Звірити axe-core типи для безпечного `messageKey` мапінгу та перевірити, чи в репо вже є parser/helper для HLS attribute lists, щоб не дублювати зайве.
ФАКТ: Пошук знайшов чинний parser BANDWIDTH у frontend/e2e/media.spec.ts і axe NodeResult/CheckResult типи. Загальний пошук зачепив minified axe bundle та урізав деталі, тому потрібні два вузькі читання перед реалізацією.
ДАЛІ: Прочитати лише axe type declarations і media.spec parser, потім сформувати мінімальний патч без нових product abstractions.
ФАКТ: Типи axe підтверджують, що `messageKey` живе у validated `CheckResult.data`; media e2e вже використовує regex `BANDWIDTH=(\d+)`. Патч внесено лише в accessibility.spec.ts і playback.spec.ts: axe findings тепер node-level `{id,messageKey,target}`, allowlist вимагає рівно один nonBmp decorative-chevron selector, а R35 парсить обидва BANDWIDTH, виводить геометричне середнє в bytes/s і збирає окремий high baseline перед throttle.
РІШЕННЯ: Діагностика R35 безумовна до асертів: master bandwidths, derived throttle, low/high baselines, low/high post-throttle counts і весь список post-throttle rendition requests. Нуль нових 720p після throttle тепер окремо червонить дискримінант, не маскуючись контрольним seek.
ДАЛІ: Переглянути змінені фрагменти, запустити точковий format-check на файлах і typecheck; якщо статичні гейти знайдуть проблему, правити лише T11-файли.
ФАКТ: Diff audit і `git diff --check` для двох T11-файлів пройшли. Allowlist не фільтрує rule цілком: загальна довжина incomplete мусить бути 1 і цей єдиний запис мусить задовольнити id/messageKey/stable-suffix. R35 перевіряє, що геометричне середнє справді лежить між сходинками, і відновлює мережу в наявному finally.
ДАЛІ: Запустити `npm run format:check`; за потреби застосувати formatter лише до двох змінених T11 test files, не до всього дерева.
ФАКТ: `npm run format:check` — RED, exit 1; formatter назвав лише `frontend/e2e/playback.spec.ts`. accessibility.spec.ts і паралельна T12-дельта форматні.
РІШЕННЯ: Не запускати заборонений `npm run format`; застосувати `npx oxfmt` лише до playback.spec.ts, як дозволену точкову механічну правку.
ДАЛІ: Відформатувати один T11 test file і повторити повний format:check.
ФАКТ: `npx oxfmt e2e/playback.spec.ts` змінив лише цільовий файл; повторний `npm run format:check` — GREEN, exit 0, 164 файли.
ДАЛІ: Запустити `npm run lint`, потім окремо `npm run typecheck`; через паралельний T12 будь-яку чужу відмову класифікувати за файлом і повторити, не лагодити її.
ФАКТ: `npm run lint` — GREEN, exit 0; oxlint без warnings, architecture boundaries checked 91 files.
ДАЛІ: Запустити `npm run typecheck`, перевіряючи типи axe check data, node-level findings, master parser і CDP throughput.
ФАКТ: `npm run typecheck` — RED, exit 2 лише в `frontend/e2e/accessibility.spec.ts:347-349`: Playwright `locator('video').evaluate` типізує вузол як загальний HTMLElement/SVGElement, тому readyState/videoWidth/videoHeight недоступні без runtime narrowing.
РІШЕННЯ: Додати `instanceof HTMLVideoElement` guard усередині browser evaluate; не використовувати type assertion, що маскує DOM input.
ДАЛІ: Внести guard і повторити typecheck, потім format-check/lint для остаточної статичної дельти.
ФАКТ: DOM-вузол у B25 тепер звужується runtime guard-ом `instanceof HTMLVideoElement`; числовий замір не покладається на TypeScript assertion.
ДАЛІ: Повторити `npm run typecheck`.
ФАКТ: Повторний `npm run typecheck` — GREEN, exit 0; axe mapping, video-state probe, HLS parser і CDP payload типізуються.
ДАЛІ: Повторити `npm run format:check` і `npm run lint` після guard-правки, потім запустити `npm test`.
ФАКТ: Post-guard `npm run format:check` — GREEN, exit 0, 164 файли.
ДАЛІ: Повторити `npm run lint`.
ФАКТ: Post-guard `npm run lint` — GREEN, exit 0; architecture boundaries checked 91 files.
ДАЛІ: Запустити `npm test`; це повна frontend unit suite, але не browser e2e доказ B24a/B25/R35.
ФАКТ: `npm test` — GREEN, exit 0: 45 files, 379/379 tests passed, 0 failed/skipped/todo, snapshots unchanged, 5.700s. Зростання від попередніх 373 тестів належить паралельній T12-дельті; падінь у чужих файлах немає.
ДАЛІ: Запустити дозволені Go-тести без Docker, спочатку з’ясувавши канонічну команду з Makefile/go module; Docker-залежні integration/build/e2e не запускати.
ФАКТ: Makefile target `test` викликає `go test -race ./...` через змінну GO_RUN; integration є окремою ціллю.
ДАЛІ: Прочитати визначення GO_RUN і виконати еквівалентний host Go unit test без Docker.
ФАКТ: `GO_RUN` у Makefile завжди використовує Docker image `golang:1.27.1-bookworm`, тому `make test` у цьому sandbox не є дозволеним способом. Без-Docker еквівалент — host `go test -race ./...` у backend, якщо host Go сумісний.
ДАЛІ: Перевірити host Go version і, якщо доступний, запустити `go test -race ./...` у backend.
ФАКТ: Host Go — точна потрібна версія `go1.27.1 darwin/amd64`, але перший `go test -race ./...` — RED до компіляції: sandbox заборонив запис у `/Users/strelia/Library/Caches/go-build` (`operation not permitted`). Це середовищна відмова, не тестова.
РІШЕННЯ: Повторити ту саму suite з task-specific `GOCACHE` у дозволеному `/tmp`; код і таймаути не змінювати.
ДАЛІ: Запустити `GOCACHE=/tmp/courses-t11-go-build go test -race ./...` у backend.
ФАКТ: Повторний host Go test стартував із дозволеним task-specific cache і не завершився за перші 30 секунд; процес продовжує працювати, раннього compile/test failure немає.
ДАЛІ: Дочекатися завершення наявної test session, не запускати дубль.
ФАКТ: Host `go test -race ./...` завершився RED, exit 1 лише в `courses/platform/httpserver.TestShutdownDrainsAnInFlightResponse`: sandbox заборонив `listen tcp 127.0.0.1:0` (`bind: operation not permitted`). Усі інші виведені Go-пакети пройшли, включно з T12-зміненими playback HTTP/observability; це середовищне блокування socket bind, не продуктова відмова.
РІШЕННЯ: Назвати повну suite незеленою; окремо прогнати всі Go-пакети крім sandbox-blocked `platform/httpserver`, не приховуючи первинний exit 1.
ДАЛІ: Перевірити склад tests у platform/httpserver і запустити race-suite для решти пакетів явним списком.
ФАКТ: platform/httpserver має п’ять tests; socket bind використовують lifecycle/server integration-style tests, тоді як ActorFromContext є чистим. Перша `go list ./...` команда була запущена з repo root і закономірно не знайшла module; це помилка робочої директорії, не гейт.
ДАЛІ: Повторити package listing із backend і запустити всі пакети крім platform/httpserver; чистий ActorFromContext прогнати окремо.
ФАКТ: Go race-suite без socket-bound package — GREEN, exit 0 для всіх перелічених пакетів; окремий чистий `platform/httpserver.TestActorFromContext` — GREEN, 2.589s. Повний `go test -race ./...` лишається RED тільки через sandbox bind prohibition і не записується як passed.
ДАЛІ: Виконати Playwright discovery для B24a/B25/R35 без запуску браузерів, потім перевірити точний diff scope і числову арифметику очікуваного throttle.
РІШЕННЯ: Playwright discovery не запускати: дельта §4 дозволяє рівно frontend static/unit команди та Go-тести без Docker, а браузерний доказ має виконувати reviewer. Попередній рядок `ДАЛІ` зберігається як історія, це рішення його явно скасовує.
ДАЛІ: Read-only перевіркою звірити final diff scope, arithmetic `sqrt(542300×3440800)/8`, відсутність правок forbidden T11 paths і точні рядки для review.
ФАКТ: Final static audit: `git diff --check` GREEN; T11 не змінив controls.tsx, LessonPlayer.tsx, docs/playback/*, work/tasks.md чи AGENTS.md. Для review-master 542300/3440800 формула дає 1365996 біт/с і `170750` Б/с — 2.52× вище 360p та 2.52× нижче 720p. Цільові рядки: allowlist 45–53, video-state probe 346–360, R35 parser/throttle/diagnostic 617–703.
ДАЛІ: Провести обов’язковий code review за skill implement: спочатку прочитати skill code-review повністю, потім окремо звірити Standards і Specification без редагування чужої T12-дельти.
ФАКТ: Skill code-review прочитано повністю. Він прямо вимагає два незалежні паралельні sub-agent огляди; це є винятком, дозволеним multi-agent правилом. Через незакомічену паралельну T12 fixed point для review — HEAD, але surface задається явними T11-шляхами, а spec sources — обидва брифи плюс plan §8/spec A12/A15.
РІШЕННЯ: Запустити read-only Standards і Specification review паралельно; агенти не редагують файлів. Commit-list непридатний для незакоміченої дельти, тому джерелом істини є explicit working-tree diff T11-шляхів.
ДАЛІ: Отримати обидва незалежні verdicts, локально перевірити кожну знахідку і виправити лише підтверджені blockers.
ФАКТ: Два незалежні read-only review запущені паралельно: Standards перевіряє AGENTS/brief prohibitions і Fowler baseline; Specification звіряє обидва брифи, §8 та A12/A15. Обом явно заборонено редагувати й наказано ігнорувати чужу T12-дельту.
ДАЛІ: Поки review працюють, власноруч перевірити причинність allowlist, video-scope pin і R35 failure ordering; потім зіставити з їхніми висновками.
ФАКТ: Standards review: один HIGH лише про ще не дописаний фінальний report, один LOW про повторення rendition-count filters; code hard violations не знайдені. Specification review: той самий report HIGH і LOW про зайвий пін порядку `360p,720p` у master; scope creep немає. Обидва підтвердили точність allowlist, master-derived throttle, окремого pre-throttle high baseline та unconditional diagnostics.
РІШЕННЯ: Report HIGH закрити фінальним append після повторних гейтів. Обидва LOW підтверджені: порядок variants не належить R35, а локальний count helper зменшує ризик переплутати highRequestsBeforeControlSeek з highRequestsBeforeThrottle. Виправити без product abstraction.
ДАЛІ: Зробити order-independent assertion двох URI та локальний count helper у R35; одночасно уточнити messageKey extraction так, щоб шукати саме string messageKey, а не перший довільний non-null check data.
ФАКТ: Review-fix внесено test-only: variant URI звіряються як точний відсортований набір із двох сходинок, локальний `renditionRequestCount('360p'|'720p')` володіє всіма count-ами, а axe mapper шукає перший валідований string `messageKey` серед check data.
РІШЕННЯ: Product-код не змінено; LOW findings закриті без розширення scope або нової залежності.
ДАЛІ: Повторити повний дозволений frontend chain у порядку format:check → lint → typecheck → test.
ФАКТ: Final review-fix `npm run format:check` — GREEN, exit 0, 164 файли.
ДАЛІ: Запустити final `npm run lint`.
ФАКТ: Final review-fix `npm run lint` — GREEN, exit 0; architecture boundaries checked 91 files.
ДАЛІ: Запустити final `npm run typecheck`.
ФАКТ: Final review-fix `npm run typecheck` — GREEN, exit 0.
ДАЛІ: Запустити final `npm test`; не інтерпретувати unit GREEN як runtime B24a/B25/R35.
ФАКТ: Final review-fix `npm test` — GREEN, exit 0: 45 files, 379/379 tests, 0 failed/skipped/todo, snapshots unchanged, 4.897s.
ДАЛІ: Повторно попросити обидва review-канали перевірити закриття своїх LOW у коді; report HIGH закрити після їхніх verdicts фінальним append-блоком.
ФАКТ: Обидва незалежні review-канали отримали follow-up на re-review коду після виправлень; report навмисно виключено з цього кола, бо його обов’язковий фінальний append залежить від verdicts.
ДАЛІ: Дочекатися re-review, перевірити final diff scope, потім дописати superseding таблицю й усі числові очікування/обмеження/команди.
ФАКТ: Повторний Standards code review — PASS, 0 hard violations і 0 actionable Fowler smells. Повторний Specification code/test review — PASS, 0 findings/blockers; exact-one allowlist, order-independent exact ladder, окремий pre-throttle baseline, master-derived throttle, diagnostics-before-assert і finally-відновлення підтверджені.
РІШЕННЯ: Code-review gate закрито. Залишився report-only HIGH із першого кола: append-нути виправлений підсумок, що явно supersede-ить старий PASS, але не стирає його.
ДАЛІ: Зняти фінальний read-only scope/status audit і line references, після чого дописати обов’язковий фінальний доказовий блок.
ФАКТ: Фінальний scope audit — `git diff --check` GREEN. T11-рядки тепер живуть у `accessibility.spec.ts:12-53,345-377`, `playback.spec.ts:604-697` і в уже дозволеному коментарі `hls-media.ts:81-82`. У дереві паралельно лишається T12-дельта; жоден її файл цією роботою не редагувався.

## Виправлення висновку попереднього проходу

ФАКТ: Старий запис «T11 code/test scope завершено локально… Specification review PASS» був передчасним. Перший справжній `make check` на Docker-capable host дав `EXIT=2`: B24a/B25 почервоніли на одному axe `incomplete`, а залежний playback-проєкт і R35 не запустилися. Окремий reviewer-run R35 дав `EXIT=1`: 360p count лишився `1` протягом 60 секунд. Отже на момент старого PASS усі три рядки були невиміряні, а два незалежні корені стали видимими лише runtime-прогонами.
РІШЕННЯ: Старий текст не змінюється як історичний слід, але цей блок його supersede-ить. Поточний code/test review PASS означає лише відповідність дельта-брифу; B24a, B25 і R35 не оголошуються GREEN без нового Docker/browser run.

## B25: exact-one allowlist і межа контрастного доказу

ФАКТ: axe mapper тепер повертає один запис на DOM-вузол: `{id, messageKey, target}`. Іменований allowlist приймає рівно один `incomplete`, для якого `id === 'color-contrast'`, `messageKey === 'nonBmp'`, а `target.endsWith('span[aria-hidden="true"]')`. Генерований React Aria id не піниться; стабільний лише суфікс декоративного шеврона Select. Загальна довжина також мусить бути `1`, тому другий такий самий вузол або будь-який новий `incomplete` робить рядок RED.
РІШЕННЯ: Пін знімається, коли axe навчиться виносити контрастний вердикт для non-BMP гліфа або UI-kit змінить DOM шеврона й стабільний суфікс зникне. Це відмова axe від вердикту, не доказ дефекту контрасту; controls.tsx не змінювався.
ФАКТ: Числове очікування browser-run у момент обох contrast audits: `<video>` має `readyState=0`, `videoWidth=0`, `videoHeight=0`; test друкує ці три числа безумовно перед асертом і пінує `0/0/0`. У цьому sandbox браузер не запускався, тому це очікування, а не власний runtime-замір; reviewer має повернути надрукований рядок.
ФАКТ: B25 доводить WCAG AA `color-contrast` над порожнім плеєром: у light очікується `violations=0`, `incomplete=1`; у dark — `violations=0`, `incomplete=1`, і обидва incomplete мусять бути тим самим типом `color-contrast/nonBmp` зі стабільним суфіксом. Рядок не доводить контраст контролів над відтворюваним відеокадром. Передбачення базового брифу §2.2 про `incomplete` через тло `<video>` перший Docker-прогін не підтвердив: єдиний incomplete був від декоративного гліфа. Розширення до media fixture — інший таск.

## R35: master-derived throttle і числовий дискримінант

ФАКТ: Попередній reviewer зняв із production master `BANDWIDTH=542300` для 360p і `BANDWIDTH=3440800` для 720p. Новий test читає обидва значення з того самого server-minted master і обчислює `sqrt(542300 × 3440800) = 1365996` біт/с, тобто після округлення `170750` Б/с для CDP. Це 2.52× вище нижньої сходинки і 2.52× нижче верхньої. O1 лишається на 1024 Б/с за іншим призначенням — завислий запит; уніфікувати числа назад не можна.
ФАКТ: Старий R35 не друкував окрему пару `highRequestsBeforeThrottle/highRequestsAfterThrottle`, тому фактичні два числа з того прогону відсутні й не вигадуються. Новий test безумовно друкує обидва baselines, обидва post-throttle counts і весь список post-throttle `.m4s` requests до асертів. Числове очікування reviewer-run: `highRequestsBeforeThrottle >= 2` (стартовий 720p + контрольний seek), `highRequestsAfterThrottle >= highRequestsBeforeThrottle + 1` (мінімум `3`); попередній low baseline був `1`, отже очікується `lowRequestsAfterThrottle >= 2` і строго більше за baseline. Якщо після дроселя приріст 720p дорівнює `0`, це `## Discussion`: корінь вище за перевірювану ABR-гіпотезу, а не привід підбирати інше число.
РІШЕННЯ: R35 не змінює seek 100/150, timeout, hls.js config чи `toBeGreaterThan`; не клікає quality і не підміняє `Hls.isSupported()`. Master мусить містити рівно дві URI 360p/720p незалежно від їх порядку; мережа відновлюється в `finally`.

## §8: виправлена таблиця «рядок → чим доводиться», п’ять питань

| Рядок | 1. Що саме доводить | 2. Production artifact/path | 3. Як виключено false green | 4. Чому саме цей рядок | 5. Що зробить test RED |
|---|---|---|---|---|---|
| B24a | У реальному Chromium рівно три контроли — play/pause button, position slider, hls.js quality Select — мають width і height `>=44px` при 100%, досягаються Tab/Shift+Tab і мають видимий computed outline у light/dark. | Справжні LessonPlayer, React Aria UI-kit і CSS browser layout; route підміняє лише валідну learning read model. `Hls.isSupported()` не mock-иться. | Геометрія міряється до 200% zoom, кожен із трьох контролів має окремі size/focus asserts; quality trigger є branch oracle hls.js. | B24 unit+axe вже доводить keyboard actions; B24a володіє браузерною геометрією та focus visibility, яких jsdom не рахує. | Хоч один width/height `<44`, пропуск у tab order, outline none/`<3px`/transparent/злитий із surface або відсутній quality control. |
| B25 | При 200% у resolved light і dark: layout failures `0`, horizontal page overflow `0`, кожен контрол усередині container, overlaps `0`, clipped direct-text `0`; axe contrast violations `0`, incomplete рівно `1` allowlisted nonBmp у кожній темі. uk/en names і `html[lang]` перевіряються тим самим browser layout. | Product semantic tokens, responsive CSS, React Aria controls і реальний layout engine. Числовий scope pin для video — `readyState/videoWidth/videoHeight = 0/0/0`. | Resolved `colorScheme` і різні surface backgrounds доводять фактичну зміну теми; exact-one allowlist не приховує новий вузол; containment/scroll/overlap/clipping є незалежними observables. | Це конкретизація нечіткого B25 «розкладка не ламається» плюс чесний WCAG AA доказ над порожнім player; контраст над playing frame прямо виключений із покриття. | Будь-який новий/другий incomplete, contrast violation, theme що не застосувалась, video state не `0/0/0`, overflow, вихід контрола, overlap, clipped uk/en text. |
| R35 | На hls.js Auto production master спочатку дає новий 720p `.m4s` після control seek на нормальному каналі; при master-derived `170750` Б/с з’являються нові 720p і 360p requests без quality action, quality лишається Auto. | Server-minted mediaPath і master із production admin upload → ffmpeg 5.1.9 worker → publish; BANDWIDTH читається з `#EXT-X-STREAM-INF`, URI — рівно 360p/720p. | Окремі baselines зняті безпосередньо перед CDP; control-seek baseline не перевикористовується; усі post-throttle requests друкуються до asserts; Hls.isSupported/config не підмінено. | P09 володіє manual rung selection; R35 окремо доводить channel-driven auto downgrade лише досяжної hls.js-гілки. | Нуль нових 720p після throttle показує інший корінь і RED; нуль нових 360p, fixed-high/manual-only ABR, неправильний ladder/BANDWIDTH, quality не Auto або ігнорований CDP також RED. |

## Нативна HLS-гілка й позаскоупне обмеження

ФАКТ: Native branch лишається недоведеною: у Chromium і Playwright WebKit замір однаковий — `{hlsJsSupported:true,mse:true,canPlayType:"maybe"}`; через несучий порядок `Hls.isSupported()` обидва рушії входять у hls.js, а Playwright WebKit не є Safari. Мертвий `return createNativeMedia(...)` після невдалого `canPlayType` лишився недоторканим і жоден новий test його не покриває; це відоме D-9 обмеження, не заявлений PASS і не привід змінювати product code в T11.

## Запущені й незапущені гейти

ФАКТ: Запущено й GREEN після review-fix: `npm run format:check` exit 0 (164 файли), `npm run lint` exit 0 (91 architecture file), `npm run typecheck` exit 0, `npm test` exit 0 (45 files, 379/379 tests, 0 skipped/todo, 4.897s). Два незалежні code reviews — Standards PASS і Specification PASS після fix round.
ФАКТ: Host `go1.27.1` повний `go test -race ./...` НЕ GREEN: exit 1 через sandbox-відмову `listen tcp 127.0.0.1:0: bind: operation not permitted` у `platform/httpserver.TestShutdownDrainsAnInFlightResponse`. Повтор без socket-bound package GREEN; окремий чистий `TestActorFromContext` GREEN. Первинна повна команда лишається чесно записаною як blocked/RED, не passed.
ФАКТ: Docker відсутній, тому НЕ ЗАПУСКАЛИСЯ: `make integration`, `make check-generated`, `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`, `make e2e-local`, повний `make check`. Playwright runtime/discovery також не запускався в цьому колі. B24a/B25/R35 після fix — authored і statically verified, але НЕ ВИМІРЯНІ та НЕ GREEN у цій сесії.

## Числові очікування reviewer-run

| Рядок | Мінімальний результат, що відрізняє «спрацювало» від «не впало» |
|---|---|
| B24a | `3` конкретні контроли; для кожного `width>=44`, `height>=44`; усі `3` отримують keyboard focus і видимий outline. |
| B25 | video `0/0/0`; light `violations=0,incomplete=1`; dark `violations=0,incomplete=1`; обидва incomplete — exact nonBmp pin; layout failures `0`. |
| R35 | master `542300/3440800` очікувано дає throttle `170750` Б/с; high baseline `>=2`, high after `>=baseline+1` (мінімум `3`), low baseline з попереднього run `1`, low after `>=2`; Auto лишається видимим. |

## Точні команди для reviewer

Повний канонічний гейт на Docker-capable host:

```sh
cd /Users/strelia/projects/courses
make check
```

Точкове повторення після `make up`, identities і seed, якщо потрібні повні diagnostics другого кола:

```sh
cd /Users/strelia/projects/courses
make up
make seed-identities
make seed
cd frontend
npx playwright test --project=chromium --grep 'B24a B25' --reporter=line
npx playwright test --project=playback --no-deps --grep 'R35 hls\.js' --reporter=line
```

Рецензент має повернути безумовні рядки `B25 video at contrast audit`, `R35 master BANDWIDTH`, `R35 baselines before throttle`, `R35 counts after throttle` і `R35 rendition requests after throttle` навіть при RED.

РІШЕННЯ: Git не змінювався і коміт не створено: прямий T11-бриф забороняє чіпати git, що сильніше за загальну вимогу implement. Staging чужої T12-дельти також порушив би B-17.
ПИТАННЯ: Немає. Наступний крок визначений: Docker-capable reviewer запускає `make check` і повертає п’ять надрукованих diagnostic lines; лише після числового runtime evidence можна оголосити B24a/B25/R35 зеленими.
ФАКТ: Фінальний audit журналу завершено: до цього запису 315 рядків, перший рядок лишився обов’язковим `ДАЛІ:`, історичні твердження не переписані, а новий блок явно їх виправляє. Є п’ятипитальна таблиця B24a/B25/R35, exact allowlist і removal trigger, video `0/0/0` та межа B25, BANDWIDTH/throttle/count expectations, native limitation, повний облік гейтів і точні reviewer commands; `git diff --check` GREEN.
РІШЕННЯ: Реалізаційна дельта T11-fix завершена настільки, наскільки дозволяє sandbox; runtime browser verdict свідомо лишено Docker-capable reviewer, без заяви про GREEN.

---

ФАКТ (рецензент, 2026-09-12): упорядкований гейт прогнано на хості, `make check` → **EXIT=0**.
Дерево містило обидві дельти — T11 і T12, — тож це спільний гейт обох.

| Ціль | Результат |
|---|---|
| `format-check`, `lint`, `typecheck` | ok |
| `npm test` | **379 passed, 0 failed** |
| Go `-race ./...` | **48 пакетів ok**, включно з `courses/platform/httpserver` 1.062s |
| `integration` | ok |
| `check-generated` | ok |
| `build`, `up`, `seed-identities`, `seed`, `smoke` | ok |
| `e2e` | **60 passed, 0 failed (7.7m)** — попередній стан був 1 failed · 29 passed · 30 did not run |

`platform/httpserver` окремо: цей пакет був єдиним Discussion у T12 — пісочниця Codex не давала
loopback bind. На хості зелений, маскування не знадобилось.

ФАКТ: B25 — надруковане число, відтворене двічі (ізольований прогін і гейт):
`readyState=0, videoWidth=0, videoHeight=0`. Детерміноване, не флейк.

ФАКТ: R35 — числа двох прогонів:

```
BANDWIDTH: 360p=542300 bit/s, 720p=3440800 bit/s
derived throttle = 170750 B/s (1 365 996 bit/s)
прогін 1: before 360p=1 720p=15 → after 360p=2 720p=19
прогін 2: before 360p=1 720p=5  → after 360p=2 720p=8
```

Приріст 360p в обох прогонах рівно `+1`; 720p гуляє втричі (15 vs 5) — емпірична підстава того,
що високий лічильник міряє таймінг, а не претензію.

ФАКТ: послідовність запитів після дроселя доводить механізм точніше за лічильники:
`720p/segment_00024 → 720p/segment_00025 → 360p/segment_00025`. Той самий індекс сегмента,
перезапитаний на нижчій сходинці — підпис abandon-and-retry. ABR **ступив** на нижчу сходинку.

РІШЕННЯ (рецензент): `expect(highRequestsAfterThrottle).toBeGreaterThan(...)` знято, число
лишилось друкованим. Причина: чи піде ще один високий фрагмент до пониження — це таймінг, а не
претензія R35; якби hls.js знизився до того, як seek щось потягне, продукт зробив би краще за
вимогу, а рядок почервонів би. Друкована пара і далі розрізняє «ABR не мав що покинути» від
«ABR покинув і ступив нижче» — на червоному.

ФАКТ (межа R35, для звіту фази): доведено, що ABR ступив на нижчу сходинку на звуженому каналі.
**Не** доведено, що плеєр на ній осів — вимір зупиняється на першому 360p-запиті, канал одразу
відновлюється. Дросель синтетичний: середнє геометричне `BANDWIDTH` із master, не реальна
мережева умова.

ДАЛІ: коміт T11 окремо від T12 явними шляхами — за словом користувача, не раніше.
