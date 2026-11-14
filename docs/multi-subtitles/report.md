# Звіт — багатомовні субтитри

План фази: [plan.md](plan.md) (ACCEPTED r6). Рішення власника O1–O6 — там, §1.

## MS1 — мова доріжки (D2) і легасі-мова (D8) — 2026-11-15

План зрізу: [plan-ms1.md](plan-ms1.md) (ACCEPTED r3). Гілка `multi-subtitles`; коміти коду:
`f00b0ff` (зріз), `7bab674` (рев'ю W1), `4ed7ed0` (фікстура, зловлена гейтом), `5778964`
(design-pass F1).

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

- Ordered `make check` на `4ed7ed0`: **GATE_EXIT=0**, e2e **138 passed (18.1 хв)**, frontend 947
  unit. Перший прогін на `7bab674` впав на `make integration`: transcript-фікстура генерувала мови
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

### design-pass (plan-ms1 §4a; вказівка власника 2026-11-14)

Маршрут `/admin/courses/<id>` → вкладка Lessons → «Audio tracks», роль admin (локальна фікстура
`frontend/e2e/auth.ts`), UI-мова English; ревізія фронтенду — `5778964` (зібраний образ). Фікстура U1/U3:
embedded-доріжка перейменована в `EN` SQL-ом на локальній БД (UI таку мову не пропонує), після проходу
повернута в `uk`. U2 — інший курс, лише `und`-доріжка. Скриншоти — [ms1-design-pass/](ms1-design-pass/),
кожен переглянутий; списки опцій зняті з DOM у тому ж стані.

| Поверхня × стан | 390×844 | 768×1024 | 1024×768 | 1440×900 | 200% тексту (CSS, 390) | reflow 320×640 |
|---|---|---|---|---|---|---|
| U1 нова доріжка, listbox відкрито, `EN` на ассеті — `English` немає | PASS | PASS | PASS | PASS | — | PASS |
| U3 перейменування `EN`, listbox — `English*` один, `EN` немає | PASS | PASS | PASS | PASS | PASS | PASS |
| U2 лише `und` — усі мови, крім `und` | — | — | — | PASS (ширина панелі) | — | — |
| U4 клавіатура: Enter відкриває, ↓ рухає з видимою рамкою, Escape закриває й повертає фокус (`:focus-visible`) — і в діалозі | PASS | — | — | — | — | — |

Горизонтального переповнення немає ніде (`scrollWidth` ≤ `innerWidth`: 390/753/1009/1425/320).
200%: listbox відкривається над тригером і прокручується, опції не обрізані; модаль прокручується, Close
досяжна. U2/U4 — стани однієї контроли, макет той самий, що U1/U3; на інших класах не повторювались
(**NOT_TESTED**, причина — дублювання того самого компонента без зміни макета).

**Знахідка F1 (виправлено, `5778964`):** доріжка `EN` показувалась сирим кодом (`EN`) у тригері,
опції й рядку доріжки, хоча вибір уже вважав її англійською. `languageName` тепер шукає назву без
урахування регістру; unit-тест; перевірено знову на всіх класах — скриншоти вище після виправлення
(«before-fix-390-U3-raw-EN» — доказ дефекту). **F2 (не дефект):** інструмент `find` показав опції з
порожнім ім'ям; DOM — ім'я з вмісту `"EN"`, без `aria-label`; артефакт інструмента.

Лінзи interface-review для зміненої контроли: ієрархія й порядок опцій без змін; стани
(відкрито/фокус/вибрано) мають видимі відмінності, не лише колір (рамка фокусу); copy — назви мов
локалізовані, код показується лише для мов поза списком (наприклад `pt-BR`) — це наявна поведінка S5.

### Обмеження (магнітуда або механізм)

- Міграція відмовляє замість виправляти: retained-том з мовою на кшталт `e` зупинить `up` до ручного
  UPDATE/DELETE; повідомлення — id першого рядка й кількість у лозі migrate.
- `media-masters` переписує всі Ready master'и на кожному `up`: на локальному стеку після повного e2e —
  7–11 master'ів, сервіс завершується за ≈1 с (події start/die в одну секунду, табл. M-R12).
- Down `00030` не повертає мови релейблу — свідомо (plan-ms1 M-D3).
- Наступні зрізи не почато: MS2+MS3 (ASR на доріжку + редактор мов), MS4 (плеєр).
