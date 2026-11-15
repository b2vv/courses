# F3 — редактор уроку не вміщується при 200% тексту на телефоні

Статус: **ACCEPTED r3** (Codex plan-critic; r1, r2 BLOCKED — `_wip-f3-plan-critique-{1,2,3}.md`). Знайдено в design-pass MS1 (`report.md`, MS1 → F3), класифіковано як наявний дефект.
Власник 2026-11-15: «Можливо дефект виправити?» → виправляємо окремим PR від `main@6ad8d8d`.

## 1. Дефект (заміряно, 2026-11-15)

Маршрут `/admin/courses/<id>` → Lessons, урок з відео й доріжкою `und`, роль admin, UI англійською,
в'юпорт 390×844, 200% тексту (CSS `html{font-size:200%}`).

- `document.documentElement.scrollWidth` = **465** при `innerWidth` 390; `body{overflow-x:clip}`
  обрізає правий край. Шеврон тригера «Choose a language» (`right` 400), праві краї полів «Track name»
  і підказок — за межею в'юпорту, недосяжні. Це не лише перенос рядків, а втрата вмісту (WCAG 1.4.10).
- `fieldset.admin-lesson` має ширину 224, а його дитина `section.admin-video-field` — 331 (переповнює).
- Мінімальний вміст (min-content, заміряно `inline-size:min-content`): тригер мови 233, слово
  «Undetermined» 215.
- Ланцюг вкладених горизонтальних відступів при 200% (padding+border+margin, px):
  `main.admin-content` 64 → `.admin-program.ui-card` 66 → `.admin-module` 36 → `.admin-lesson` 33 →
  `.admin-video-field.ui-card` 66 → `fieldset.admin-audio-new` 50 = **315**. Вміст 233 + 315 = 548 > 390.
- Механізм: усі ці відступи — токени в `rem` (`--space-*`, `--card-padding`), тож при збільшенні тексту
  подвоюються разом зі шрифтом. Додатково `fieldset.admin-audio-new` має UA `min-inline-size:
  min-content` (у `.admin-module/.admin-lesson` це вже знято `min-inline-size:0`).

## 2. Рішення

F-D1 (2026-11-15, Courses 1; r2 після plan-critic r1). Вкладені **горизонтальні** відступи редактора
уроку обмежуються шириною в'юпорту — `min(<поточний токен>, 2.5vw)`, логічними властивостями
(`padding-inline` / `padding-inline-start`, не фізичний `left`):

| Селектор | Зараз (100% тексту) | Стане |
|---|---|---|
| `.admin-program.ui-card` | `--card-padding` (admin override 1rem = 16) | `padding-inline: min(var(--card-padding), 2.5vw)` |
| `.admin-module` (+ override `@media (max-width:40rem)`, `main.css:1455`) | start `--space-6` 24 / mobile `--space-4` 16 | `padding-inline-start: min(<той самий токен>, 2.5vw)` в обох місцях; shorthand переписано логічно |
| `.admin-lesson` | start `--space-4` 16 | `padding-inline-start: min(var(--space-4), 2.5vw)`; shorthand логічно |
| `.admin-video-field.ui-card` | `--card-padding` 16 | `padding-inline: min(var(--card-padding), 2.5vw)` |
| `li.admin-audio-track` | `--space-3` 12 | `padding-inline: min(var(--space-3), 2.5vw)` |
| `fieldset.admin-audio-new` | `--space-3` 12, UA `min-inline-size: min-content` | `padding-inline: min(var(--space-3), 2.5vw)`; `min-inline-size: 0` |

F-D2 (діалог перейменування, r1 Major 1). Діалог поза цим ланцюгом: `.ui-modal-overlay` `padding:
var(--space-4)` і `.ui-dialog` `padding: var(--space-6)` (`tokens.css:380,393`) теж подвоюються:
390 − 64 − 96 − 2 = 228 < 233. Той самий прийом у глобальних примітивах: overlay `padding-inline:
min(var(--space-4), 4vw)`, `.ui-dialog` `padding-inline: min(var(--space-6), 5vw)`. На 1440 при 100%
(57.6/72 px ≥ токенів) — без змін; на 390 при 100%: overlay 15.6 (було 16), dialog 19.5 (було 24).
Зачіпає всі діалоги застосунку → повний e2e і огляд щонайменше двох інших діалогів.

F-D3 (текст користувача, r1 Major 3). Назва доріжки — `<strong>` до 64 символів
(`maximumAudioLabelLength`, `audio-track.ts:83`) без переносу; відступи її не обмежують. Для тексту,
який вводить користувач, перенос будь-де — правильна поведінка: `.admin-audio-track-head > strong {
overflow-wrap: anywhere; min-inline-size: 0 }`. Фіксовані тексти інтерфейсу (тригер, «Undetermined»)
лишаються з переносом між словами (спроба MS1 з `anywhere` на тригері — відкинута).

F-D4 (редактор транскрипту, r2 Major 1). `TranscriptEditor` (`TranscriptEditor.tsx:262–280`) у
кожному рядку рендерить простий `<fieldset>` (UA `min-inline-size: min-content`) з `<label>`-ами назв
документів курсу — текст користувача до **200** символів (`transcript-schema.ts:42`) без переносу; CSS
для `.transcript-editor-lines` немає взагалі. Додаємо: `.transcript-editor-lines fieldset {
min-inline-size: 0; margin-inline: 0 }`, `.transcript-editor-lines fieldset label { overflow-wrap:
anywhere }` (текст користувача — перенос будь-де; легенда «Documents for line N» — фіксований текст,
між словами).

- Чому `vw`, а не rem/em: `vw` не росте з розміром тексту (при методі нижче) — саме залежність rem від
  розміру тексту і є механізмом дефекту.
- Чому не container query / `:has()`: `min()`+`vw` підтримуються всіма цілями Rsbuild за замовчуванням
  (Firefox ≥104). `:has()` — Firefox 121 (W5 F5); container queries — Firefox 110 (caniuse, окремо від W5).
- Магнітуда при 200%/390 (CSS-симуляція): ланцюг уроку 64 (оболонка) + 10 відступів × 9.75 + рамки 13
  = 174.5 → вміст ≈ 215. Мінімум вмісту: тригер 233, слово «Undetermined» 215. **Тригер не вміщується
  на 18 px** при 10 відступах — тому відступи `li.admin-audio-track` і `fieldset.admin-audio-new` —
  паралельні (не вкладені один в одного): у ланцюгу до тригера 8 відступів (program×2, module, lesson,
  video×2, audio-new×2) + рамки 11 = 89 → вміст 237 ≥ 233; до «Undetermined» — 8 (… li×2) → 237 ≥ 215.
  Діалог: 390 − 2×15.6 − 2×19.5 − 2 = 317.8 ≥ 233.
- Діапазон токенів 12–24 px (r1 Minor 7). Де 2.5vw ≥ токен — значення без змін: 1440 (36) і 1024 (25.6)
  ≥ усіх; 768 (19.2) змінює лише module start 24 → 19.2.
- Закрита альтернатива: `overflow-wrap:anywhere` на значенні тригера — ламає слова по складах.
- Межа: `main.admin-content` (оболонка, 64 px) не чіпаємо — спільна для всіх сторінок адмінки.

### Метод збільшення тексту і межі твердження (r1 Major 4)

Перевірка — CSS-симуляція `html{font-size:200%}` у Chromium-панелі Claude (в'юпорт емулюється), як у
MS1. Вона еквівалентна налаштуванню «розмір шрифту» браузера для властивостей у `rem` (обидва
міняють базу rem). Відмінності, які фіксуємо, а не замовчуємо: (а) медіа-запити в `rem/em` рахуються від
**початкового** розміру шрифту — при симуляції `@media (max-width:40rem)` = 640 px, при справжньому
налаштуванні 32 px — 1280 px (override module тоді діє ширше; з `min()` обидва випадки ≤ 2.5vw);
(б) page zoom змінює ширину в'юпорту в CSS px — це випадок reflow (320), а не 200% тексту; (в)
text-only zoom Firefox на стенді недоступний (немає Firefox) → **NOT_TESTED** з причиною; твердження
«vw не росте з текстом» обмежене протестованим механізмом.

### Не входить (r1 Major 2 — заперечення)

320 px **разом** з 200% тексту не перевіряємо й не обіцяємо. Контракт design-pass
(`viewport-coverage.md`, «Scaling and accessibility — separate checks») і plan-ms1 §4a ставлять їх
окремими перевірками, і цей зріз обіцяє саме їх. Це не твердження про загальну відповідність WCAG:
W3C (Understanding 1.4.10, «Resize text») розглядає 1.4.4 і 1.4.10 і в поєднанні; комбінований випадок
лишається відомим обмеженням (запис у звіті). Арифметика r1 (320 − 64 − 75 = 181 < 233) правильна, але це
комбінація, якої жоден із двох критеріїв не вимагає; вона потребувала б зміни оболонки адмінки
(поза межею F3).

## 3. Рядки приймання

Фікстура: локальний стенд, e2e-курс з уроком з однією `und`-доріжкою (id — у звіті), admin з
`frontend/e2e/auth.ts`, UI English, Chromium-панель; шрифти завантажені (`document.fonts.status ===
'loaded'`), скролбар панелі — оверлейний (запис `innerWidth` і `clientWidth` документа). Для F-R9 —
тимчасовий рядок `course_documents` цього курсу з назвою 200 символів без пробілів (SQL на локальній
БД, видалити після) і один незбережений рядок транскрипту через «Add a line». Стани ASR —
ті, що є на фікстурі (фіксуються у звіті поіменно; невидимі стани не вважаються покритими). Базова ревізія для «без змін» — `6ad8d8d`
(той самий образ, курс, в'юпорт, мова).

| # | Перевірка | Як |
|---|---|---|
| F-R1 | 390×844, 200%: документ `scrollWidth` ≤ `clientWidth`; для кожного відрендереного in-flow нащадка `.admin-lesson` (`offsetParent` ≠ null, `position` не fixed/absolute): `left` ≥ лівого й `right` ≤ правого краю контентної області `.admin-lesson` (±0.5 px); для `.admin-lesson`, `.admin-video-field`, `.admin-audio-field`, `fieldset.admin-audio-new`, `li.admin-audio-track`, полів транскрипту — `scrollWidth` ≤ `clientWidth`; тригер з шевроном, «Track name», підказки видимі; фіксований текст переноситься лише між словами. Поповери (портал) — окремо F-R2/F-R5 | DOM-заміри + скрин до/після |
| F-R2 | діалог перейменування при 200%/390: `.ui-modal` `scrollWidth` ≤ `clientWidth`, тригер з шевроном усередині, вибрана опція видима при відкритті, Escape → фокус на тригері, другий Escape → «Edit track …» | браузер, справжні клавіші |
| F-R3 | 320×640 при 100%: `scrollWidth` ≤ 320, вкладене переповнення відсутнє (як F-R1) | браузер |
| F-R4 | 1024/1440 при 100%: обчислені `padding-inline*` і `getBoundingClientRect` `.admin-program`, `.admin-module`, `.admin-lesson`, `.admin-video-field`, тригера — **рівні** базі `6ad8d8d`; 768 — дельта лише module start 24 → 19.2 і похідні ширини; 390 при 100% — записати до/після | DOM-заміри на обох ревізіях |
| F-R5 | клавіатура: Tab → тригер → Enter → ↓ → Escape, видимий фокус не обрізаний, при 200%/390 | браузер |
| F-R6 | довга назва: доріжка з назвою 64 символи без пробілів (SQL на локальній БД, після — повернути) при 200%/390 і 100%/320 — переноситься в межах картки | браузер |
| F-R7 | глобальні примітиви: обчислені `padding-inline` `.ui-modal-overlay` = `min(16px, 4vw)` і `.ui-dialog` = `min(24px, 5vw)` за формулою, допуск ±0.1 px (390: 15.6/19.5; 1440: 16/24) на трьох діалогах — перейменування доріжки, ще один діалог адмінки, один діалог клієнтської частини — у світлій і темній темі на 390 і 1440 при 100%; без переповнення (`.ui-modal` `scrollWidth` ≤ `clientWidth`) | браузер |
| F-R9 | редактор транскрипту: рядок з документом-назвою 200 символів — при 200%/390 і 100%/320 fieldset і label у межах `.admin-lesson` (правило F-R1), назва переноситься, чекбокс ≥ видимий і клікабельний | браузер |
| F-R8 | ordered `make check` (спільний CSS → повний e2e) | гейт |

design-pass: 4 базові класи + 200% + 320 reflow, скрини в `docs/multi-subtitles/f3-design-pass/`.
Рев'ю — Codex до SHIP. Звіт — розділ «F3» у `report.md`.

## 4. Ризики

- Інші сторінки: `grep` по `frontend/src` — класи ланцюга задають лише `CourseEditorView`,
  `LessonVideoField`, `LessonAudioTracksField`; змонтовані тільки в `routes/admin.courses.$courseId.tsx`.
  `.ui-modal-overlay`/`.ui-dialog` — глобальні (усі діалоги) → F-R7 + повний e2e.
- На 390 при 100% тексту вкладені відступи зменшуються (12–24 → 9.75 px) — видима зміна вкладеності на
  телефоні; рамки `border-inline-start` лишаються індикатором ієрархії. Скрин до/після.
- Застосунок не має RTL-локалі (uk/en); логічні властивості — для коректності, без окремого RTL-прогону.
