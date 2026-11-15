# Звіт — багатомовні субтитри

План фази: [plan.md](plan.md) (ACCEPTED r6). Рішення власника O1–O6 — там, §1.

## MS1 — мова доріжки (D2) і легасі-мова (D8) — 2026-11-15

План зрізу: [plan-ms1.md](plan-ms1.md) (ACCEPTED r3). Гілка `multi-subtitles`; коміти коду:
`f00b0ff` (зріз), `7bab674` (рев'ю W1), `4ed7ed0` (фікстура, зловлена гейтом), `5778964`
(design-pass F1), `f415643` (звіт), `c34c0d0` (design-pass F4, глобальний `.ui-popover`/`.ui-listbox`), `3759d07` (рев'ю W5 F5: клас замість `:has()`).

### Що зроблено

- `internal/contracts/language.ValidTag` — одна граматика тегу для transcript і media (ADR 013);
  нормалізація лишилась модульною: transcript `TrimSpace`+`ToLower` (без змін), media — ASCII-guard і
  `ToLower`, регістр зберігається.
- Міграція `00030`: preflight на вихідних рядках → легасі-релейбл `und`→мова primary (O6) → CHECK
  `audio_tracks_language_shape` → унікальність `(asset, lower(language))` під тим самим ім'ям індексу.
- `cmd/migrate` логує текст власного `RAISE EXCEPTION` (лише SQLSTATE `P0001`); решта — загальним текстом.
- `media-worker regenerate-masters` + compose one-shot `media-masters` перед `media-worker`.
- Фронтенд: те саме правило в `validAudioLanguage`; вибір мови без урахування регістру; назва мови
  без урахування регістру (`EN` = English).
- OpenAPI: `pattern` мови в запитах mint/rename і у відповіді доріжки.

### Рядки приймання

| # | Статус | Доказ |
|---|---|---|
| M-R1 | PASS | unit `contracts/language`, `media/domain` `TestTrackLanguageFollowsTranscriptGrammar`, `transcript/domain` (без змін) |
| M-R2 | PASS | integration `MS1 track language CHECK and case-insensitive uniqueness` (прямий INSERT/UPDATE) |
| M-R3 | PASS | той самий integration (rename `de`→`EN`, create `EN` → `duplicate_language`; failed не тримає мову) |
| M-R4 | PASS | integration `MS1 00030 preflight…`: `e`, `und-`, `und`+`UND` → відмова, версія 29, рядки незмінні |
| M-R5 | PASS | integration `MS1 00030 relabels…`: 13 доріжок у 8 ассетах, кожна мова перевірена |
| M-R6 | PASS | той самий тест: DownTo 29 лишає `uk`, приймає `UK` поруч з `uk` і `e` |
| M-R7 | PASS | `cmd/media-worker` `TestRegenerateMastersAfterLegacyRelabel`: legacy master від production writer → 00030 → `LANGUAGE="uk"`, другий прогін байт-у-байт, evicted не торкнуто |
| M-R8 | PASS | unit `TestRegenerateAllMastersKeepsGoingAndClassifies`, `TestCommandMode` |
| M-R9 | PASS | unit `TestAuthoredRefusalSurfacesOnlyRaiseException` (загорнуто як у goose; 23514 і не-SQL — редакція) |
| M-R10 | PASS | contract `TestAudioTrackContracts/MS1…` (`e`, `und-`, `1x` → 400, репозиторій не викликано) |
| M-R11 | PASS | unit `audio-track.test.ts` (таблиця, `EN` займає `en`, власна `EN` лишається) + `LessonAudioTracksField.test.tsx` |
| M-R12 | PASS | живий прогін нижче |

### Гейти і мутації

- Ordered `make check` на фінальному коді `3759d07` (після F5): перший прогін **червоний** на
  `make integration` — `TestPlaybackSessionsAndAuthorizationAgainstRedisAndPostgreSQL`:
  `port "5432/tcp" not found` (testcontainers не віддав мапінг порту Postgres-контейнера; дифф зрізу
  на цьому коміті — лише фронтенд-CSS і два `className`, playback-backend не зачеплено). Повний
  повторний прогін: **GATE_EXIT=0**, e2e **138 passed (18.2 хв)**, frontend 948 unit. Червоний прогін
  не зараховано як зелений. Точна причина в Docker/testcontainers не встановлена; збій — на
  отриманні рядка підключення (`sessions_integration_test.go:74–76`), до відкриття БД і міграцій.
- На `c34c0d0` (після F4, глобальний CSS): GATE_EXIT=0, e2e 138 (18.3 хв), frontend 948. Попередній на `f415643` (після F1): GATE_EXIT=0, e2e 138
  (18.4 хв), frontend 948. Попередній на `4ed7ed0`: GATE_EXIT=0, e2e 138, frontend 947. Перший прогін на `7bab674` впав на `make integration`: transcript-фікстура генерувала мови
  `'l-'||md5…` (однолітерний перший subtag) — новий CHECK її відкинув. Ні таргетний media-integration,
  ні рев'ю плану цього не знайшли; виправлено в `4ed7ed0`. Пастка — в Obsidian `_gotchas.md`.
- Мутації (кожна повернута, `cmp` чистий), кожна червонить названий тест: індекс без `lower()` →
  M-R3; релейбл вимкнено → M-R5; preflight після релейблу → M-R4 (b) і (c), (a) лишається зеленим;
  домен без граматики → M-R1; `regenerateMasters` з іншим root → M-R7; фронтенд без `toLowerCase` у
  зайнятих мовах → M-R11.

### M-R12 — compose, живий прогін (2026-11-15, Docker Compose, локальний стек)

Тепла БД після `make check`; легасі-випадок створила сама e2e: embedded `und` доріжка
(`c6b8acd9…`) Ready-ассета `bf464489…`, урок з primary `uk`. Схему відкочено до v29 (індекс, CHECK,
рядок `goose_db_version`), `docker events` запущено до кожного `up`.

| Сценарій | Подія (час, unix) | Результат |
|---|---|---|
| (a)1 `stop` → `make up` | `migrate` die 0 @…760 → `media-masters` start @…761, die 0 @…761 → `media-worker` start @…762 | версія 30; доріжка `uk`; master `LANGUAGE="uk"`; `regenerated: 7`; sha256 `2fd26465…` |
| (a)2 повтор | `migrate` die 0 @…808 → `media-masters` die 0 @…809 → `media-worker` start @…809 | master sha256 той самий `2fd26465…` |
| (b) `make up` по запущеному | `media-masters` start/die 0 @…833; подій `media-worker` немає | `StartedAt` worker'а до/після однаковий `2026-11-14T22:03:29.902Z` |
| (c) override `POSTGRES_PASSWORD` лише для `media-masters` | `media-masters` die **1** @…855; `media-worker` start немає | `up --wait` exit 1, «service "media-masters" didn't complete successfully» |

Після (c) — `make up` exit 0, стек відновлено.

(a)3 — третій прогін на теплому томі після фінального гейту (діагностика нижче): `migrate` →
`media-masters` → `media-worker`, події @1794695628–630; `regenerated: 7`; master ассета `6031021c…`
`LANGUAGE="und"` → `"uk"`.

### Діагностика retained-тому (plan-ms1 M-D3, read-only)

Запит — [ms1-diagnostic/diagnostic.sql](ms1-diagnostic/diagnostic.sql): кожна не-failed `und`/`und-*`
доріжка Ready не-evicted ассета, мова primary уроку і причина, чому 00030 її (не) релейблить; предикат
кандидата й вибір (default → embedded → менший id) — ті самі, що в кроці 2 міграції (рев'ю W3: перша
версія називала eligible кожного кандидата, без вибору). `before.txt` знято першою версією: у кожному
ассеті там один кандидат, тож вибір нічого б не змінив; `after.txt` — поточною. Тепла БД
після фінального гейту, схему відкочено до v29, запит до і після `make up`.

| | Рядків | eligible | no primary transcript | already has primary language |
|---|---|---|---|---|
| до ([before.txt](ms1-diagnostic/before.txt)) | 7 | 1 (доріжка `8d5ab454…`, ассет `6031021c…`, primary `uk`) | 6 | 0 |
| після ([after.txt](ms1-diagnostic/after.txt)) | 6 | 0 | 6 | 0 |

Решта шість `und` — ассети без primary-транскрипту: за O6 мову їм немає звідки взяти, вони лишаються
`und` і субтитрів не отримують (D1). Невиясненого `und` на ассеті з primary не лишилось.

### design-pass (plan-ms1 §4a; вказівка власника 2026-11-14)

Маршрут `/admin/courses/<id>` → вкладка Lessons → «Audio tracks», роль admin (локальна фікстура
`frontend/e2e/auth.ts`), UI-мова English; ревізія фронтенду — `5778964` (зібраний образ; `f415643`
фронтенд не змінює). Фікстура U1/U3/U4:
embedded-доріжка перейменована в `EN` SQL-ом на локальній БД (UI таку мову не пропонує), після проходу
повернута в `uk`. U2 — інший курс, лише `und`-доріжка. Скриншоти — [ms1-design-pass/](ms1-design-pass/),
кожен переглянутий; списки опцій зняті з DOM у тому ж стані.

| Поверхня × стан | 390×844 | 768×1024 | 1024×768 | 1440×900 | 200% тексту (CSS, 390) | reflow 320×640 |
|---|---|---|---|---|---|---|
| U1 нова доріжка, listbox відкрито, `EN` на ассеті — `English` немає | PASS | PASS | PASS | PASS | PASS (див. F3) | PASS |
| U3 перейменування `EN`, listbox — `English*` один, `EN` немає | PASS | PASS | PASS | PASS | PASS | PASS |
| U2 лише `und` — усі мови, крім `und` | PASS | PASS | PASS | PASS | PASS | PASS |
| U4 клавіатура: Tab з «Edit track» → тригер, Enter відкриває, ↓ рухає з видимою рамкою, Escape закриває й повертає фокус (`:focus-visible`) — і в діалозі | PASS | PASS | PASS | PASS | PASS (після F4) | PASS |

U2/U4 при 200% і 320 — на `und`-курсі (U2-набір опцій), разом: нова доріжка — Tab → Enter → ↓ (English
у фокусі), End прокручує до Spanish, Escape → фокус на тригері `:focus-visible`; діалог — Enter на
«Edit track» → тригер → Enter → ↑/↓, Escape закриває список (фокус на тригері діалогу, `:focus-visible`),
другий Escape закриває діалог і повертає фокус на «Edit track Original (Undetermined)».

Скриншоти: `<в'юпорт>-U<n>-….jpg`. Списки опцій і активний елемент зняті з DOM у тому ж стані, що
кадр: U1/U4 — Ukrainian, Polish, German, French, Spanish, Undetermined (English немає; ↓ → Polish);
U2 — Ukrainian, English, Polish, German, French, Spanish (Undetermined немає). Горизонтального
переповнення немає ніде (`scrollWidth` ≤ `innerWidth`: 390/753/1009/1425/320).
200%: listbox відкривається над тригером і прокручується до останньої опції, опції не обрізані; модаль
прокручується, Close досяжна.

**Походження скриншотів.** Усі кадри без префікса `before-fix-` зняті після `5778964`.
`before-fix-390-U3-raw-EN.jpg` і `before-fix-390-U4-keyboard-focus.jpg` зняті до F1 (сирий `EN` у рядку
доріжки) і лишені як доказ дефекту; `desktop-pane-U2-und-only-open.jpg` — перший U2-кадр у ширині панелі
браузера (не базовий клас), замінений рядками 390–1440.

**Знахідка F1 (виправлено, `5778964`):** доріжка `EN` показувалась сирим кодом (`EN`) у тригері,
опції й рядку доріжки, хоча вибір уже вважав її англійською. `languageName` тепер шукає назву без
урахування регістру; unit-тест; перевірено знову на всіх класах — скриншоти вище після виправлення
(`before-fix-*` — доказ дефекту). **F2 (не дефект):** інструмент `find` показав опції з
порожнім ім'ям; DOM — ім'я з вмісту `"EN"`, без `aria-label`; артефакт інструмента.

Лінзи interface-review для зміненої контроли: ієрархія й порядок опцій без змін; стани
(відкрито/фокус/вибрано) мають видимі відмінності, не лише колір (рамка фокусу); copy — назви мов
локалізовані, код показується лише для мов поза списком (наприклад `pt-BR`) — це наявна поведінка S5.

**F3 (наявний до MS1, Minor, не виправлено в MS1):** при 200% тексту на 390 px поле «New audio track»
має ширину 139 px — тригер ламає «Choose a language» у три рядки, «Track name» у два
(`390-text200-U1-open-top.jpg`). Механізм — вкладені відступи редактора уроку (сторінка → модуль →
урок → поле), не MS1: розмітку й стилі контроли MS1 не міняв. Спроба виправити в `tokens.css`
(`overflow-wrap:anywhere` на значенні + `flex:none` на шевроні) розбила текст по складах — гірше,
повернуто. Контрола лишається працездатною (відкривається, прокручується, вибір працює). Пропоную
окремою задачею: стиснути відступи вкладених карток адмін-редактора уроку на вузькій ширині.

**F4 (наявний до MS1, Major за рев'ю W4, виправлено `c34c0d0`):** при 200% тексту на 390 px діалог перейменування
відкриває список із фокусом на вибраній опції (Undetermined, сьома з семи), але не прокручує до неї:
опція 557–645 px, видима межа поповера 397 px (`390-text200-U4-dialog-F4-selected-obscured.jpg`) —
сфокусований елемент повністю схований (WCAG 2.4.11). Механізм: прокручується `.ui-popover`
(`overflow:auto`, `tokens.css`), а не `.ui-listbox`; React Aria прокручує до фокуса саме listbox, тож
початковий фокус лишається за краєм. Перша ж стрілка прокручує поповер (`…-after-arrow.jpg`). На 320
при 100% і на всіх базових класах список уміщається — не відтворюється. MS1 не змінював ні поповер, ні
listbox (рев'ю W3: `git diff 5cbb44b..HEAD -- frontend` без CSS/розмітки); дефект глобального примітиву
`Select`. Рев'ю W4: повне зникнення фокуса в перевіреній контролі — Major, виправляється в MS1.
Виправлення (`tokens.css`): поповер, чия пряма дитина `.ui-listbox`, — flex-колонка без відступу;
прокручується listbox (`overflow:auto`, `min-block-size:0`), відступ і `scroll-padding` перенесено на
нього, щоб рамка фокуса опції лишалась усередині. Поповер плеєра (`Dialog` усередині) не зачеплено.
Перевірка після: діалог при 200%/390 відкривається з Undetermined у видимій зоні (опція 292–380 px,
listbox 13–396, `scrollTop` 265; поповер не прокручується), Home → Ukrainian, рамка повна (по 16 px з
боків), Escape → фокус на тригері (`390-text200-U4-dialog-open-selected-visible.jpg`,
`…-dialog-home.jpg`; кадри до — `before-fix-390-text200-U4-dialog-*`). Регресія примітиву (той самий
стенд, 4 класи, справжні клавіші): Select власника (Overview), нова доріжка (End), мова розпізнавання
(End), ComboBox користувача (Access, ↓) — у кожному сфокусована опція видима, поповер сам не
прокручується, рамка з відступом 8 px з боків, `scrollWidth` ≤ ширини (390/753/1009/1425);
ComboBox на 390 (7 опцій) прокручує listbox. Кадри `F4-regression-*`. `Menu` у застосунку не
використовується. Повний ordered `make check` на `c34c0d0` — GATE_EXIT=0, e2e 138 (розділ «Гейти»).

**F5 (рев'ю W5, Major, виправлено `3759d07`):** `:has()` не підтримується Firefox до 121, а дефолтний
browserslist Rsbuild (у проєкті без override) включає Firefox ≥104 — там F4 повертався. Тепер `Select`,
`ComboBox` і `Menu` явно мають клас `ui-popover-list` (`controls.tsx`, `overlays.tsx`), правило — на
класі; поповер плеєра (`PlayerSettingsMenu`, усередині `Dialog`) класу не має й не змінюється.
Live-прогону у Firefox 104–120 немає (браузера немає на стенді); механізм — звичайний клас-селектор.

**Перевірка на фінальній ревізії `3759d07` (F6 рев'ю W5).** Стенд після гейту (новий e2e-курс
`982aa7a6…` з однією `und`-доріжкою; попередній курс прибрав гейт), повторний вхід тестовим admin з
фікстури. Діалог при 200%/390: відкривається з Undetermined видимим (`scrollTop` 265, поповер не
прокручується, рамка 16/16 px), Escape → тригер `:focus-visible`, другий Escape → «Edit track …».
Reflow 320×640: нова доріжка — Tab → Enter → ↓ (English видима), End → Spanish видима, Escape → тригер
`:focus-visible`; діалог — відкриття з Undetermined видимим, Escape ×2 з поверненням фокуса; ComboBox —
7 опцій, прокручується listbox, сфокусована видима. Чотири класи (390/768/1024/1440) × Select власника,
нова доріжка (End), мова розпізнавання (End), ComboBox (↓): у кожному поповер має `ui-popover-list`,
сфокусована опція видима, поповер не прокручується, рамка 8/8 px, `scrollWidth` 390/753/1009/1425.
Кадри `final-*`. Повний ordered `make check` на `3759d07` — GATE_EXIT=0 з другого прогону (розділ «Гейти»).

Рев'ю Codex: W1–W5 BLOCK (докази, F4, F5/F6), **W6 SHIP** — записи `_wip-ms1-review-codex-w{1..6}.md`.

### Обмеження (магнітуда або механізм)

- Міграція відмовляє замість виправляти: retained-том з мовою на кшталт `e` зупинить `up` до ручного
  UPDATE/DELETE; повідомлення — id першого рядка й кількість у лозі migrate.
- `media-masters` переписує всі Ready master'и на кожному `up`: на локальному стеку після повного e2e —
  7–11 master'ів, сервіс завершується за ≈1 с (події start/die в одну секунду, табл. M-R12).
- Down `00030` не повертає мови релейблу — свідомо (plan-ms1 M-D3).
- Наступні зрізи не почато: MS2+MS3 (ASR на доріжку + редактор мов), MS4 (плеєр).

## F3 — редактор уроку при 200% тексту на телефоні — 2026-11-15

План: [plan-f3.md](plan-f3.md) (ACCEPTED r3; r1, r2 BLOCKED — `_wip-f3-plan-critique-{1,2,3}.md`).
Гілка `f3-lesson-editor-narrow` від `main@6ad8d8d`. Запит власника 2026-11-15 («Можливо дефект
виправити?»).

### Дефект (до, ревізія `6ad8d8d`)

390×844, 200% тексту (CSS `html{font-size:200%}`, Chromium-панель): документ `scrollWidth` **465** при
`clientWidth` 390, правий край обрізаний `body{overflow-x:clip}`; 53 видимі нащадки уроку за його
контентною областю (правий край 325 → до 465) — шеврон тригера, поля, підказки недосяжні
(`f3-design-pass/before-390-text200-new-track.jpg`). Механізм: шість рівнів вкладених горизонтальних
відступів у `rem` (оболонка 64 + програма 66 + модуль 36 + урок 33 + картка відео 66 + fieldset 50 =
315 px при 200%) плюс UA `min-inline-size: min-content` у `fieldset.admin-audio-new`. MS1 класифікував
F3 як «перенос рядків» — замір показав обрізання вмісту.

### Що змінено

- `main.css`: горизонтальні відступи `.admin-program`, `.admin-module` (і mobile override),
  `.admin-lesson`, `.admin-video-field`, `li.admin-audio-track`, `fieldset.admin-audio-new` —
  `min(<токен>, 2.5vw)`, логічні властивості; `fieldset.admin-audio-new { min-inline-size: 0 }`.
- `tokens.css`: `.ui-modal-overlay` `padding-inline: min(var(--space-4), 4vw)`, `.ui-dialog`
  `padding-inline: min(var(--space-6), 5vw)` (усі діалоги).
- Текст автора переноситься будь-де: назва доріжки (≤64), назви документів у рядку транскрипту
  (≤200) і **список матеріалів курсу** `.document-library-items a`. Останнє — нова знахідка цього
  зрізу: фікстура з назвою документа 200 символів без пробілів розтягнула всю сітку редактора до
  `scrollWidth` 4717 (`section.document-library` на тій самій вкладці); правило F-R1, що міряло лише
  нащадків уроку, цього не бачило — винуватця знайдено бінарним пошуком по DOM.

### Рядки приймання (ревізія коду `4d31192`, образ перезібрано з неї)

F-R1, F-R2, F-R5 зняті на проміжній збірці без правила `.document-library-items a` (решта CSS — як
`4d31192`); у курсі тоді не було документів, тож це правило ні на що не впливало. F-R3, F-R4, F-R6,
F-R7, F-R9 — на збірці, ідентичній `4d31192`.

| # | Статус | Доказ |
|---|---|---|
| F-R1 | PASS | 390/200%: `scrollWidth` 390 = `clientWidth`; 61 in-flow нащадок уроку, 0 за межами (5 кандидатів — приховані службові елементи React Aria під `clip-path: inset(50%)`: прихований checkbox і `HiddenSelect`; правило уточнено — виключати нащадків візуально прихованого предка); `scrollWidth`=`clientWidth` для уроку 290, картки відео 278, audio-field 259, fieldset 257, li 257; тригер 89–326, шеврон 301; «Choose a / language» — два рядки між словами. `after-390-text200-new-track.jpg` |
| F-R2 | PASS | діалог при 200%/390: `.ui-modal` 357/357, overlay 15.6, dialog 19.5 px; тригер 36–354, шеврон 329; відкриття — Undetermined видима; Escape → тригер `:focus-visible`; другий Escape → «Edit track Original (Undetermined)». `after-390-text200-dialog-*.jpg` |
| F-R3 | PASS | 320×640/100%: `scrollWidth` 320; 75 in-flow, 0 за межами; усі відступи 8 px (2.5vw); фікстури довгих назв увімкнені. `after-320-reflow-new-track.jpg` |
| F-R4 | PASS | база `6ad8d8d` vs після, 100%, той самий курс/образ/мова, шрифти `loaded`, класичний скролбар 15 px (`innerWidth` 1440 / `clientWidth` 1425; 1024/1009; 768/753): **1440 і 1024 — відступи й `left/width` усіх шести контейнерів і тригера ідентичні**; 768 — лише module start 24 → 19.2, похідні зсуви 4.8 px (урок 61→56.2, ширина 659→663.8); 390 — відступи 12–16 → 9.75, тригер 100→72.8 (ширина 193.5 без змін). `after-390-100-new-track.jpg`, `after-768-100-editor.jpg` |
| F-R5 | PASS | 390/200%: Tab з «Edit track» → тригер, Enter, ↓ → English (видима, у межах в'юпорту), Escape → тригер `:focus-visible` |
| F-R6 | PASS | назва доріжки 64×«W»: 200%/390 — 237 px завширшки, переноситься (563 px заввишки), урок без переповнення; 320/100% — так само. `after-390-text200-long-track-label.jpg` |
| F-R7 | PASS | `padding-inline` за формулою (±0.1): 390 — overlay 15.6, dialog 19.5; 1440 — 16 / 24 (= база); `.ui-modal` `scrollWidth`=`clientWidth` (357, 559/574): перейменування доріжки (адмінка, 390 і 1440), «Create course» (адмінка, 390 і 1440), «Delete review?» (клієнт, студент з фікстури, 390 і 1440, світла й темна тема). Адмінка має лише темну тему (`tokens.admin.css` ігнорує `prefers-color-scheme`) — світла N/A. Діалоги закривались Escape, нічого не створено й не видалено. `dialog-*.jpg` |
| F-R8 | PASS | ordered `make check` на `4d31192`: **GATE_EXIT=0** з першого прогону, e2e **138 passed (18.4 хв)**, frontend 948 unit |
| F-R9 | PASS | документ курсу 200×«D» + незбережений рядок транскрипту: 200%/390 — fieldset 280/280, label 277 px, переноситься; чекбокс 13 px у межах; 320/100% — fieldset 249/249. Список матеріалів курсу — 324/324. Фікстури видалено (`DELETE 1`, назва доріжки повернута `Original`). `after-390-text200-long-*.jpg` |

Стани ASR на фікстурі: «No speech detected in the video» + вибір мови розпізнавання + «Regenerate
transcript»; інші стани ASR не відрендерені й не вважаються покритими.

Рев'ю Codex: **W1 SHIP** (без знахідок) — `_wip-f3-review-codex-w1.md`.

### Обмеження

- Комбінація 320 px **і** 200% тексту не перевірялась і не обіцяється (оболонка адмінки 64 px +
  мінімальні відступи лишають ≈181 px при вмісті 233) — потребувала б зміни оболонки.
- Метод — CSS-симуляція в Chromium; text-only zoom Firefox — **NOT_TESTED** (немає Firefox на
  стенді); медіа-запити в `rem` під симуляцією рахуються від 16 px.
- На 390 при 100% вкладеність стала щільнішою (відступи 12–16 → 9.75 px); ієрархію тримають рамки
  `border-inline-start`.
- Інші місця з текстом автора поза редактором курсу (наприклад, матеріали рядка транскрипту в плеєрі
  студента) не перевірялись.

## MS2a — чернетка ASR на доріжку: міграція 00031 і репозиторій — 2026-11-15

План: [plan-ms2-3.md](plan-ms2-3.md) (ACCEPTED r9), рішення MS-D1 (бекенд), MS-D2, MS-D3, MS-D10
(лише сховище). Коміт `c8fa93f` на `ms2-3-asr-per-track`, не злитий; один PR і один ordered gate
разом із MS2b/MS2c/MS3 (план §5).

### Що зроблено
- Міграція `00031`: PK `asr_drafts` → `audio_track_id`, знімки `track_language`/`track_multilingual`,
  бекфіл за (ассет, leaf), незіставлені чернетки видалено, `source_leaf` прибрано;
  `audio_tracks.multilingual` і унікальність `(video_asset_id, source_leaf)` (Codex w1 minor 1).
- Pending, claim, complete, recover на доріжку; TaskID і payload з `trackId`; claim `Superseded`/
  `NotFound` завершує задачу без збою.
- Видалено `/draft/publish`, `Drafting.Publish`, `TranscriptPublisher`, `SaveDraft`; openapi і
  клієнт перегенеровано. Старі `GET /draft` і `/draft/regenerate` тимчасово працюють на default-доріжку до MS2c.

### Рядки приймання (ревізія `c8fa93f`)
| Рядок | Статус | Доказ |
|---|---|---|
| M2-R1 | PASS | integration, включно з вибором у Down (default, потім найновіша) |
| M2-R2, R3, R4, R7, R8, R9 | PASS | integration, `go test -tags=integration -v`: 29 підтестів M2 PASS, 0 FAIL/SKIP |
| M2-R5 | PASS (частково) | unit `queue_test`: TaskID містить доріжку; «дві доріжки → 2 задачі» — MS2b |
| M2-R11 | PASS | unit: payload без `trackId` → `invalid_task_payload` |
| M2-R6, R10, R12 | NOT_TESTED | MS2b/MS2c |

### Гейти
- `go build`, `go vet`, `go test ./internal/transcript/... ./internal/media/...` — ok; `make format-check` — ok; frontend `typecheck` — ok.
- Повний `make integration` двічі (до і після фіксу w1): `GATE_EXIT=0`, 67 пакетів ok, 0 FAIL.
- Не запускались для цього зрізу: `make e2e`. Наявні C06/S6 e2e очікують стару модель; їх оновлює MS3 (M3-R9).
- Рев'ю: Codex read-only w1 — SHIP, 2 minor закрито (`_wip-ms2a-review-codex-w1.md`).

## MS2b — language hint from the track, multilingual mode, `ASR_MULTILINGUAL_DEFAULT` — 2026-11-15

Plan: [plan-ms2-3.md](plan-ms2-3.md) (ACCEPTED r9), decisions MS-D4 and MS-D10 (env default and the media
application; media HTTP is MS2c). On `ms2-3-asr-per-track`, not merged; one PR and one ordered gate together with MS2c/MS3.

### What was done
- The `Transcriber` port takes the track language and its `multilingual` from the claim snapshot. `RecognitionJob.Language`,
  the payload's `language` and the `language` parameter of `Regenerate` were removed (MS-D4). The legacy `/draft/regenerate`
  still validates the body but ignores the language until MS2c.
- The whisper adapter: `multilingual` → `--multilingual`; the `und` family → no hint; otherwise the first subtag,
  if it is in the 100 faster-whisper 1.2.1 codes; otherwise no hint plus `hint_unsupported` (tag only).
- `transcribe.py`: `--multilingual` → `language=None, multilingual=True`.
- `ASR_MULTILINGUAL_DEFAULT` (`platform/config`): `true`/`false`; empty = `false`; anything else → api and
  media-worker fail to start. The value goes into the embedded track (transcode) and into the uploaded-track mint;
  added to `.env.example` and docker-compose (api, media-worker). media-masters passes `false`: it does not create tracks.

### Acceptance rows (MS2b revision)
| Row | Status | Evidence |
|---|---|---|
| M2-R5 | PASS | unit `queue_test`: two tracks of one asset → two different TaskIDs |
| M2-R6 | PASS (adapter + Process) | unit `engine_test`: `en-US`→`en`, `de`, `haw`, `und`/`und-x-original`→none, `tlh`→none + log, list = 100 codes; `recognize_test`: Process passes the claim snapshot |
| M2-R11 | PASS | unit: payload without `trackId` → error, process not called, `MaxRetry(0)` |
| M2-R12 | PASS (Python boundary, env, mint/embedded, claim) | `TestRealTranscribePythonModelBoundary`: the real `transcribe.py` via `python3` with a stub `faster_whisper`, kwargs `True/None`, `False/"en"`, `False/None`; ran on the host and in `golang:1.27.3-bookworm`. Config table; mint/embedded true/false. Media HTTP (PATCH/GET `multilingual`, `multilingualDefault`) — MS2c |
| M2-R10 | NOT_TESTED | MS2c |

### Gates
- `make format-check`, `make lint`, `make test` — exit 0; the whisper tests in `golang:1.27.3-bookworm` — exit 0.
- Full `make integration`: `GATE_EXIT=0`, 67 packages ok, 0 FAIL.
- `make e2e` was not run (as in MS2a: C06/S6 are updated by MS3).
- Review: Codex read-only w1 — SHIP, no findings (`_wip-ms2b-review-codex-w1.md`).
- Not verified: inference by the real model in multilingual mode (the test stubs only the model itself).

## MS2c — HTTP for drafts per track, transcript inventory, media `multilingual` — 2026-11-15

Plan: [plan-ms2-3.md](plan-ms2-3.md) (ACCEPTED r9), MS-D5, MS-D10 (HTTP), phase D3. On `ms2-3-asr-per-track`, not merged.

### What was done
- `GET …/lessons/{id}/drafts` — one draft per track of the latest Ready asset, in the order default → embedded →
  by language → id; `stale` uses the MS-D5 rule. `POST …/drafts/{trackId}/regenerate` — no body, `202`;
  `404 track_not_found` (new code), `409 already_running`. `/draft` and `/draft/regenerate` removed (404).
  The temporary "empty trackID = default" convention from MS2a removed.
- `GET …/lessons/{id}/transcripts` — `{language, updatedAt, stale, isPrimary}` by `language`. Access follows the course-editor
  rule of the admin `PUT …/transcript`. **Decision (2026-11-15, Courses 1/Codex):** the plan cited "the same access as the admin
  `GET …/transcript`", but that endpoint does not exist. Checked by the HTTP cases owner/foreign/anonymous and by integration.
- Media: mint takes an optional `multilingual`, absent → env. PATCH takes an optional `multilingual`; the master is
  rewritten only when language/label changed, compared under the asset lock (`FOR UPDATE`). GET returns `multilingual` and
  `multilingualDefault`.
- openapi and the client regenerated. **Temporary frontend bridge (until MS3a):** the editor takes `drafts[0]` after strict
  validation of every item; regenerate runs on the trackId of the draft shown; the Regenerate language Select removed; the strict audio-track
  schema accepts `multilingual`/`multilingualDefault`. Without it the Audio tracks list would fail at runtime while typecheck stays green.

### Acceptance rows (MS2c revision)
| Row | Status | Evidence |
|---|---|---|
| M2-R10 | PASS (backend + client) | HTTP contract tests for the new and removed routes and codes; integration for the inventory (0/1/2 languages, stale after a Ready replacement, no Ready → not stale, foreign course → forbidden); `make check-generated` exit 0. The editor's tabs come in MS3 |
| M2-R12 (HTTP/media) | PASS | mint without the field → env, with it → the field; PATCH of the flag alone → 0 `RegenerateMaster` calls, of the language → 1; the repository write in integration; GET returns both fields |

### Gates
- `make format-check`, `lint`, `typecheck`, `test`, `integration` (67 packages ok, 0 FAIL), `check-generated` — exit 0.
- The first integration run caught a defect in the new fixture: two pending assets per lesson against
  `video_assets_pending_per_lesson`. It now demotes only the latest one.
- The w1 Minor test (no Ready asset) was added while the gate was already running. It passed separately
  (`go test -tags=integration -run ^TestTranscriptReadAgainstPostgreSQL$`); its run inside the full integration gate is not confirmed.
- `make e2e` was not run: the ASR specs expect the old URLs; MS3 updates them (M3-R9).
- Review: Codex read-only w1 — SHIP, 1 Minor closed (`_wip-ms2c-review-codex-w1.md`).
