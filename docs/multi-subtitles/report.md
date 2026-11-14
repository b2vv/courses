# Звіт — багатомовні субтитри

План фази: [plan.md](plan.md) (ACCEPTED r6). Рішення власника O1–O6 — там, §1.

## MS1 — мова доріжки (D2) і легасі-мова (D8) — 2026-11-15

План зрізу: [plan-ms1.md](plan-ms1.md) (ACCEPTED r3). Гілка `multi-subtitles`; коміти коду:
`f00b0ff` (зріз), `7bab674` (рев'ю W1), `4ed7ed0` (фікстура, зловлена гейтом), `5778964`
(design-pass F1), `f415643` (звіт).

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

- Ordered `make check` на фінальному коді `f415643` (після F1): **GATE_EXIT=0**, e2e **138 passed
  (18.4 хв)**, frontend 948 unit. Попередній на `4ed7ed0`: GATE_EXIT=0, e2e 138, frontend 947. Перший прогін на `7bab674` впав на `make integration`: transcript-фікстура генерувала мови
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
доріжка Ready не-evicted ассета, мова primary уроку і причина, чому 00030 її (не) релейблить. Тепла БД
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
| U2 лише `und` — усі мови, крім `und` | PASS | PASS | PASS | PASS | N/A¹ | N/A¹ |
| U4 клавіатура: Tab з «Edit track» → тригер, Enter відкриває, ↓ рухає з видимою рамкою, Escape закриває й повертає фокус (`:focus-visible`) — і в діалозі | PASS | PASS | PASS | PASS | N/A¹ | N/A¹ |

¹ U2 і U4 — інший набір опцій і інший спосіб введення для тієї самої контроли; макет при 200%/320
перевірено U1/U3 на тій самій контролі. Набір опцій від ширини не залежить (`availableAudioLanguages`).

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

### Обмеження (магнітуда або механізм)

- Міграція відмовляє замість виправляти: retained-том з мовою на кшталт `e` зупинить `up` до ручного
  UPDATE/DELETE; повідомлення — id першого рядка й кількість у лозі migrate.
- `media-masters` переписує всі Ready master'и на кожному `up`: на локальному стеку після повного e2e —
  7–11 master'ів, сервіс завершується за ≈1 с (події start/die в одну секунду, табл. M-R12).
- Down `00030` не повертає мови релейблу — свідомо (plan-ms1 M-D3).
- Наступні зрізи не почато: MS2+MS3 (ASR на доріжку + редактор мов), MS4 (плеєр).
