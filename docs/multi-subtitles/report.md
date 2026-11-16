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

## MS3a — data and language tabs in the lesson editor — checkpoint 2026-11-16

Plan: [plan-ms2-3.md](plan-ms2-3.md) (ACCEPTED r9, `_wip-ms23-plan-critique-{1..9}.md`), MS-D6, MS-D7, MS-D10 (frontend),
on `ms2-3-asr-per-track` @ `a985543`. Owner go-ahead: "Починай", 2026-11-16 (chat, Courses 1).

### Pre-code checkpoint
- **Scope / acceptance (full).** Rows M3-R1, M3-R2, M3-R3, M3-R5, M3-R7, M3-R10, M3-R14, M3-R15 at unit level, and the save-path parts
  of M3-R11 (new segment ids on every PUT) and M3-R13 (token choice, branches a–d). Their e2e parts are MS3c. Non-goals: the player and panel
  (MS4), the UX fixes from the editor review (owner, 2026-11-15: at the end).
- **Decision (2026-11-16, Courses 1): what MS3a leaves to MS3b.** The plan's split names only "data and tabs" / "`und` flow and recovery".
  MS3a therefore excludes M3-R4 (the `und` confirm dialog, PATCH → PUT), M3-R6 / M3-R12 (detached snapshots, flush on unmount) and the
  `baseLanguage` half of M3-R13. Until MS3b, inside the one PR: an `und` tab loads its draft, but Save is refused with a message, so 0 PUT
  with `language=und`. Recovery keys are already per tab (MS-D9 key), so a legacy `user:lesson` snapshot is not offered between MS3a and MS3b.
  Check: unit "und tab Save → 0 PUT"; MS3b reopens both.
- **Grilling / decisions.** None open. O7 and O8 hold. The current "stale draft = pending recognition, wait 15 min" poll in
  `admin-draft-queries.ts` contradicts O7 and is removed (M3-R5).
- **Dependencies / writers.** MS2c committed (`a985543`); the HTTP drafts, inventory and media `multilingual` exist in the generated
  client. Writer: Codex (implementation runs), Courses 1 (gates, browser, commit). Peer sessions share the working tree: only files of this
  slice are touched.
- **Scoped impact.** `modules/admin/ui/CourseEditorView.tsx` (mount point of the field), new tabs component under `modules/admin/ui`;
  `modules/transcript/{domain,adapters,application,ui}` (`TranscriptEditor`, `LessonDraftField`, `LessonTranscriptField`, recovery key);
  `modules/media/{adapters,ui}` (`LessonAudioTracksField`, rename/mint); `app/composition/{admin-draft-queries,admin-audio-queries}.ts`;
  `routes/admin.courses.$courseId.tsx` (wiring). Existing e2e (`transcript.spec`, `asr.spec`, `media*.spec`) select by role. Hidden panels
  use `display:none`, so they leave the accessibility tree and keep strict-mode selectors unique; checked in MS3c by the full e2e run.
- **Plan critique / defense.** Covered by the whole-plan critique r9 ACCEPTED; MS3a adds no contract. Post-code: Codex read-only review
  (`review_mode=independent_subagent` → Codex runtime, per owner memory 26.10).
- **Admission: READY** (owner go-ahead 2026-11-16; no open required question).

### What was done (Codex runs 1–2, Courses 1 fixes)
- Run 1 (`_wip-ms3a-run1.md`): the drafts list per lesson replaces the `drafts[0]` bridge; the inventory adapter (strict schema);
  pending markers keyed `lessonId:trackId`; the stale-draft wait (`awaitingReplacement`) removed (O7); an audio-track PATCH invalidates
  the lesson's drafts; the pure guard `draftLoadable(draft, track)`; the "Multiple languages" switch in the new-track form
  (initial `multilingualDefault`) and on each track row (PATCH with the full body).
- Run 2 (`_wip-ms3a-run2.md`): `LessonLanguageTabs` (`modules/admin/ui`). One tab per ready track of the latest Ready asset, plus
  orphans; the no-Ready fallback is the old single editor. The editor is keyed by tab identity; recovery keys are per tab; the
  guard text is shown instead of Load. Save re-reads the tracks and inventory directly, with the token rules of MS-D6 and new ids
  on every PUT. The interim `und` rule: Save refused, 0 PUT.
- **Courses 1, browser defect 1 (2026-11-16):** the tab set was computed from `tracks.data ?? []` while the tracks were still
  loading. A transcript in a track's language showed as an orphan, React Aria selected that orphan tab and kept it after the tracks
  arrived, and its editor mounted under the orphan's recovery key. Seen live: the "Ukrainian" orphan was selected instead of the
  default track. Fix: tabs render only when both lists are loaded; a failed read shows an alert. Two unit tests were added.
- **Courses 1, browser defect 2:** the row switch and the new-track switch had the same accessible name "Multiple languages". The
  row switch is now named per track, like Make default and Delete ("Multiple languages in track Original (Undetermined)").

### design-pass (browser, Kiri stack rebuilt from the branch, migrate `applied: 1` = 00031)
Fixture: course "Субтитри e2e-transcript-…" (lesson `7539aed1`): one `und` default track with a `no_speech` draft, plus an orphan
`uk` transcript.
| Check | Result | Evidence |
|---|---|---|
| Tabs and order (M3-R1) | PASS | "Unknown language · Default" selected first, then "Ukrainian"; after fix 1 |
| Unique ids, inert panels, names (M3-R7) | PASS | `querySelectorAll('[id]')` → 0 duplicates; inactive panel `inert` + `display:none`; `aria-labelledby` → active tab |
| 1440 / 1024 / 768 / 390 | PASS | `scrollWidth` ≤ viewport at each width; tabs 44 px high, stacked at 390 |
| 320 px and 200 % text | PASS | 0 elements of the tabs/panel past the viewport; tabs 304×188 / 193×88 |
| Keyboard | PASS | ArrowRight selects "Ukrainian", 3 px focus outline (`:focus-visible`); Tab enters the active panel ("Line 1 start, ms") |
| `und` tab Save (interim rule) | PASS | line typed → "Set the track's language in Audio tracks first", 0 non-GET requests, line kept |
| Orphan Save (M3-R2/R10 orphan branch) | PASS | requests: GET inventory → PUT transcript → refetch; "Transcript saved."; text restored afterwards |
| Row switch (M3-R14) | PASS | PATCH body `{"language":"und","label":"Original","multilingual":false/true}` → refetch of audio-tracks AND drafts (M3-R15 invalidation); restored to `false` (DB checked) |
| Two track tabs, Load guard on a ready draft, rename flows | NOT_TESTED in browser | need a second (dub) track; unit-tested here, e2e in MS3c (M3-R8, R11, R15) |

The browser evidence above is of the revision before the review-w1 fixes. Those fixes change only the refresh-failure and
post-confirmation paths, not the rendered tabs, and are covered by unit tests (below); they were not re-walked in the browser.

### Review
- Codex read-only w1 — **BLOCK, 2 Major**, both in `LessonLanguageTabs.tsx`:
  1. My fix for browser defect 1 unmounted every editor when a background refresh of the tracks or inventory failed, losing
     unsaved lines. Now the alert sits above the tabs while cached lists exist. Test: "a failed background refresh keeps the editors…".
  2. A track renamed or deleted while the replace dialog was open still got `PUT language=<new>` (M3-R10). Now `trackMatches` is
     re-checked after confirmation, and the confirmed token is kept. Test: "a track renamed while the replace dialog is open sends
     zero PUT". Mutation-checked: without the re-check the test fails (26/27).
- Codex read-only w2 — **SHIP, no findings** (`_wip-ms3a-review-codex-w2.md`).

### Gates (final MS3a revision)
- Frontend: `oxfmt --check`, `tsc --noEmit`, `npm run lint` (oxlint + architecture check), `npm test` (**1047 passed, 0 skipped**)
  — all exit 0.
- Not run for this slice: `make integration` (no backend change since `a985543`, where it was green); `make e2e` (ASR specs expect
  the pre-MS2 URLs, MS3c updates them); the full ordered `make check` runs once on the final MS3 commit.

## MS3b — `und` flow and recovery per tab — checkpoint 2026-11-16

Plan: [plan-ms2-3.md](plan-ms2-3.md) (ACCEPTED r9), MS-D8, MS-D9, the `baseLanguage` part of MS-D6, on `ms2-3-asr-per-track` @ `9ad04e3`.
Owner go-ahead: "Так, MS3b", 2026-11-16 (chat, Courses 1).

### Pre-code checkpoint
- **Scope / acceptance (full).** Rows M3-R4 (the `und` confirm dialog, PATCH → PUT, every failure branch), M3-R6 (key per tab, detached
  snapshots offered by reading only), M3-R12 (c) and flush on unmount at unit level, and the `baseLanguage` half of M3-R13 (snapshot
  writes the pair, restore returns it, a snapshot without `baseLanguage` takes the "other language" branch, e/f/g at unit level). Their e2e
  parts (M3-R4 e2e, M3-R6 two-tab reload, M3-R12 a/b, M3-R13 a/b) are MS3c. Non-goals: player/panel (MS4), editor UX fixes (owner,
  2026-11-15: at the end).
- **Closes the MS3a interim rule.** The "`und` tab Save refused, 0 PUT" rule is replaced by the dialog; its unit test is rewritten, not
  deleted (the dialog's Cancel still sends 0 PATCH and 0 PUT).
- **Grilling / decisions.** None open: O5 (PATCH then PUT), O7, O8 hold. Split into two Codex runs (memory: implementation and tests per
  narrow run): run 1 = M3-R4 + `baseLanguage`; run 2 = detached snapshots + flush on unmount.
- **Dependencies / writers.** MS3a committed (`9ad04e3`). Writer: Codex; Courses 1 runs gates, browser, review, commit. Shared tree: only
  this slice's files.
- **Scoped impact.** `modules/admin/ui/LessonLanguageTabs.tsx` (+ test), `modules/transcript/{application/recovery.ts,
  ui/use-transcript-recovery.ts, ui/TranscriptEditor.tsx, ui/LessonTranscriptField.tsx, ports/TranscriptHistoryStore.ts}`, locales. The
  IndexedDB adapter keeps its schema: `baseLanguage` is an optional field of the stored object, so old snapshots stay readable (they take
  branch g). No backend change.
- **Plan critique / defense.** Covered by the whole-plan critique r9 ACCEPTED; MS3b adds no contract. Post-code: Codex read-only review
  (`review_mode=independent_subagent` → Codex runtime).
- **Admission: READY** (owner go-ahead 2026-11-16; no open required question).

### MS3b run 1 — Codex implementation evidence, 2026-11-16

Scope: M3-R4 and the `baseLanguage` half of M3-R13 at unit level. Journal and complete changed-file list:
[`_wip-ms3b-run1.md`](_wip-ms3b-run1.md). No commit; based on `9ad04e3`, retaining shared WIP.

- `und` Save now opens the React Aria language confirmation after direct fresh track/inventory reads. Options exclude `und`,
  existing track languages (case-insensitive), and inventory languages; detected draft language is preselected only if available.
  Empty selection refuses confirmation with text; no options offers only Cancel. Confirm invokes the existing injected full-body
  audio rename, then a tokenless transcript save with fresh segment ids. PATCH failure names the error, sends no PUT and leaves
  recovery untouched; capture is suspended while Save/dialog is pending. Cancel/Escape return focus to Save.
- PATCH success + PUT failure keeps the `track:<id>` editor and typed lines. A discovered production-path gap was fixed:
  `LessonTranscriptField` now preserves an already-mounted editor while the new-language transcript query initially loads.
  The retry test includes this loading interval and checks the exact textarea node, then ordinary fresh-read PUT in the new language.
- Capture/restore/recapture carries (`baseLanguage`, `expectedUpdatedAt`). **Decision (Codex, 2026-11-16):** a legacy snapshot restores
  `baseLanguage` as `''`, which cannot equal a real language tag; the existing language-inequality branch therefore handles unknown
  token ownership everywhere. Check: M3-R13 g absent/present target tests and recapture assertion. IndexedDB retains version 1 and
  validates optional fields on read, accepting snapshots both with and without `baseLanguage`.
- Final requested ordered frontend gate: `npm run format` **exit 0** (398 files), `npm run typecheck` **exit 0**,
  `npm run lint` **exit 0** (239 architecture files), `npm test` **exit 0 — 1064 passed, 105 files, 0 skipped/todo**.
  `git diff --check -- frontend` **exit 0**. Earlier exploratory harness/type errors and the 64/64 targeted run are in the journal;
  only the final ordered run above is the delivery gate.
- Limits / next: browser persistence is not proven by the IndexedDB event mock; browser/e2e, independent review and the full ordered
  `make check` were **NOT_RUN** in this run. Run 2 still owns detached snapshots and flush on unmount; MS3c owns browser/e2e and the
  final phase gate. No backend, generated or e2e files changed. No phase-wide closure claimed.


### MS3b run 2 — Codex implementation evidence, 2026-11-16

Scope: MS-D9, M3-R6 and M3-R12c/flush at unit/component level, plus the run-1 failed-Save capture correction.
Append-only journal, changed files and final frontend hashes: [`_wip-ms3b-run2.md`](_wip-ms3b-run2.md).
Shared run-1 WIP retained; branch/HEAD unchanged at `ms2-3-asr-per-track` / `9ad04e3`; no commit.

- Pending idle capture now flushes on unmount using the latest lines and baseline. No pending change means no unmount write;
  returning to baseline clears instead of writes. Mount alone no longer arms a timer. Store rejection remains best-effort.
- Pure detached placement uses the exact `user:lesson` / `user:lesson:` boundary, excludes every current tab key, matches the
  current language case-insensitively, falls back to default, and orders newest first. A transcript-public lesson hook reads the
  keys/inventory once per tab set for both tabs and the single editor. Admin imports only the public transcript surface.
- Own offers have priority, including while their async read is pending. Detached offers stay inside their panel, carry source
  and known/unknown language copy in uk/en, and show one at a time. Restore keeps the base-language/token pair and records the
  accepted detached key; captures write only the own key. Successful Save finalizes own and accepted keys; explicit Discard clears
  only that detached key and advances. Restore without Save offers again after remount. Viewer sweep policy is unchanged and tested
  against track/lang key forms. Existing lesson+instance ids remain unique.
- **Run-1 correction:** failed/cancelled saves immediately capture current lines after re-enabling capture; they never clear/finalize
  recovery. The pending-capture failure test observes one write and no stale unmount write. This supersedes run 1's wording that
  PATCH failure leaves recovery wholly untouched: a protective capture is now allowed. Single-editor fallback also honors MS-D6:
  foreign/unknown base-language tokens are omitted from PUT.
- Final requested ordered frontend gate: `npm run format` **exit 0** (400 files), `npm run typecheck` **exit 0**, `npm run lint`
  **exit 0** (240 architecture files), `npm test` **exit 0 — 1092 passed, 106 files, 0 skipped/todo**, 12.806 s.
  There are 28 additional tests versus run 1. Scoped `git diff --check` **exit 0**. Earlier fixture/lint failures and their fixes
  remain recorded in the journal; the counts above belong to the final implementation.
- Limits / next: production React paths are exercised with in-memory history, including dirty orphan attachment and single→tabs
  unmount/read ordering. Native IndexedDB teardown completion, browser layout/focus/persistence, browser/e2e, independent review and
  full ordered `make check` are **NOT_TESTED / NOT_RUN** here. No missing in-scope unit test identified. Courses 1 owns independent
  review; MS3c owns M3-R6 browser reloads, M3-R12 a/b and final full gate. No generated, e2e or backend file changed; no phase closure.

### MS3b fix runs 1–6, review w1–w7 and final state (Courses 1), 2026-11-16
The run 1/2 notes above are superseded where they disagree with this section. Briefs and append-only journals:
`_wip-ms3b-fix{,2,3,4,5,6}-brief.md` / `_wip-ms3b-fix{,2,3,4,5,6}.md`. Each fix carries a test that fails with the fix
reverted (revert checks recorded per fix in the journals).

**Review (Codex read-only, `review_mode=independent_subagent` → Codex runtime), `_wip-ms3b-review-codex-w{1..7}.md`:**
| Round | Verdict | Findings → fix |
|---|---|---|
| w1 | BLOCK | Restore A then B, Save deleted both source keys; a clean rename cleared an unresolved offer on unmount; Cancel said "Replacement cancelled" → fix 1 |
| w2 | BLOCK | fix 1's "pending offer refuses capture" lost new edits; a write before the initial read overwrote an unread snapshot → fix 2: the editor is read-only until the initial read resolves and while its own offer is unresolved |
| w3 | BLOCK | Save retired a source key that had become another mounted tab's own key holding newer lines → fix 3: version-checked retirement + relinquishment; focus after Restore/Discard |
| w4 | BLOCK | compare-then-delete was two operations; cleanup ownership lost on reload → fix 4: `clearIfVersion` (one IndexedDB `readwrite` transaction), `acceptedSource` stored in the snapshot |
| w5 | BLOCK | detached Discard deleted another window's newer lines; ancestry lost through a second detachment → fix 5: versioned Discard + refreshed offer; bounded ancestry (8) |
| w6 | BLOCK | a clean blur after own Discard cleared a newer foreign version → fix 6 closes the CLASS: every deletion path inventoried (`_wip-ms3b-fix6.md`), all ordinary deletions are `clearIfVersion` with the version this editor read/offered/wrote; unknown version → no delete; only the cross-author sweep is unconditional |
| w7 | **SHIP** | 1 Minor (below, accepted as a limitation) |

**Decisions (dated, with a check).**
- 2026-11-16, Courses 1: while a tab's own recovery state is unknown or unresolved, its lines are read-only and Add/Remove/Save refuse
  with «Спершу відновіть або відкиньте незбережені зміни». Closes the alternative "silently refuse capture" (lost edits, w2). Check:
  fix-2 tests; browser below.
- 2026-11-16, Courses 1 (w2 Major 3 kept): MS-D8 "recovery keys untouched on PATCH failure" = no clear, no finalize, no move to another
  key. Final mechanism (fix 6): Save flushes the pending author change BEFORE suspending capture, and a failed Save does not re-capture
  after the response (it could overwrite intervening work). Check: failed-Save integration test (one write, no extra unmount write).
- 2026-11-16, Codex fix 6, accepted by Courses 1: own-offer Discard deletes exactly the offered version (Phase 8 only remembered a
  dismissal per tab session, but the next clean blur cleared the key anyway, so the visible behaviour is unchanged). A successful Save
  retires only the version this editor last wrote; a newer capture from another window survives and is offered next visit.
- 2026-11-16, Codex run 1, accepted: a legacy snapshot restores `baseLanguage = ''` (never a real tag), so the "different language"
  token branch handles unknown ownership (M3-R13 g).

**Limitation (accepted 2026-11-16, Courses 1; magnitude + mechanism).** Version ownership compares `capturedAt` (ms). Two windows that
capture DIFFERENT lines under the SAME key within the same millisecond get equal versions; a Save in one can then retire the other's
unsaved lines. Reproduced by w7 only with a clock forced to equal values. Fix if it is ever seen: a random capture id compared atomically,
`capturedAt` kept for ordering.

### design-pass MS3b (browser, Kiri stack rebuilt from the branch after each fix run)
Fixture as MS3a (lesson `7539aed1`: `und` default track, `no_speech` draft, orphan `uk` transcript). PATCH/PUT were stubbed in the page
(`window.fetch`) for failure paths, so the fixture data was not changed; IndexedDB test snapshots were removed after each check (0 keys).
| Check | Result | Evidence |
|---|---|---|
| `und` Save → dialog (M3-R4) | PASS | keyboard Enter on Save → "Confirm the transcript language", focus in dialog; options English/Polish/German/French/Spanish (no `und`, no `uk` from the inventory); nothing preselected (draft has no language); empty confirm → "Choose a language before saving." |
| PATCH `duplicate_language` | PASS | stubbed 409 → "A track in that language already exists.", requests GET inventory, GET tracks, PATCH, **0 PUT**, line kept, focus on Save |
| Cancel / Escape | PASS (after fix 1) | before: Escape → focus `<body>`, copy "Replacement cancelled"; after: focus on Save, "Language not changed; your lines are still here.", 0 writes; same for the replace dialog |
| Legacy snapshot placement (M3-R6) | PASS | legacy `user:lesson` with `uk` → offered in the inert "Ukrainian" panel ("Unsaved lines from before tabs (language: uk)"); without language → default tab, "language unknown" |
| Legacy restore → Save (M3-R13 g) | PASS | replace confirmation for `uk` instead of a PUT with a token; after fix 1 the heading reads "Ukrainian" and a localized date; before fix 1 a false "track language is now uk" notice in the orphan tab — gone |
| Lock while own offer pending (fix 2) | PASS | Add/Save `aria-disabled`, hint shown, Add click → 0 lines, focus stays; Discard unlocks |
| Focus after Restore/Discard (fix 3) | PASS | own Discard → next offer's Restore; Restore → "Line 1 text" (3 px focus outline) |
| `acceptedSource` in real IndexedDB (fix 4) | PASS | after restoring legacy A in `lang:uk`, the stored `lang:uk` snapshot carries `acceptedSource` = A's key + exact `capturedAt`, `baseLanguage: ""` |
| Versioned Discard in real IndexedDB (fix 5) | PASS | offer A shown, store overwritten with newer C → Discard keeps C and offers it (focus on Restore); Discard of C → key deleted |
| 1440 / 1024 / 768 / 390, 320 px, 200 % | PASS | dialog: 0 elements past the viewport at each width, `scrollWidth` = viewport; buttons 44 px; at 320 px + 200 % the modal scrolls (`overflow-y: auto`) |
| Final revision smoke | PASS | tabs, `und` dialog, Cancel copy, 0 writes, 0 duplicate ids |
| Save that retires a source (`clearIfVersion` after PUT), PATCH ok + PUT fail, two windows | NOT_TESTED in browser | would change the fixture / need two windows; unit/component tests + adapter test (mocked IDB requests); e2e in MS3c |

### Gates (final MS3b revision)
- Frontend (Courses 1 rerun on the final tree): `oxfmt --check` 0, `tsc --noEmit` 0, `npm run lint` 0, `npm test` **1185 passed, 0
  failed, 0 skipped**.
- Not run for this slice: `make integration` (no backend change), `make e2e` (MS3c), the full ordered `make check` (once, on the final
  MS3 commit).
