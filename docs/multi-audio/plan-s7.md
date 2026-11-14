# План S7 — вибір аудіодоріжки в плеєрі (multi-audio)

Статус: **ACCEPTED** (plan-critique Codex, раунд 4, 2026-11-14; раунди 1–3 BLOCKED — §8). Гілка `multi-audio-s7` від `main@3f0e1c8`. Батьківський план:
`docs/multi-audio/plan.md` §5 S7 («`hls.audioTracks` у gear-меню; перемикання з мінімальним gap; e2e на
реальному multi-audio HLS»). Обсяг — лише frontend-плеєр; backend, контракт і master.m3u8 не змінюються
(D2 батьківського плану: доріжки приходять маніфестом, playback-DTO незмінний).

## 1. Рішення власника (2026-11-14, ця сесія, AskUserQuestion)

- **R1. Обрана мова аудіо запам'ятовується між уроками** (localStorage, як швидкість/гучність): якщо в
  наступному уроці є доріжка тієї ж мови — вона стартує одразу; інакше грає основна.
- **R2. Субтитри для кожної доріжки — бажана поведінка, але НЕ в S7.** Власник: «весь урок має бути
  мультимовним»; обсяг — «S7 зараз, субтитри — далі» окремим планом (ASR на доріжку, редактор/публікація
  по мовах, читання за мовою). Семантика для того плану зафіксована зараз: **субтитри слідують за мовою
  обраного аудіо**; коли транскрипта цієї мови немає — перемикач субтитрів це показує. У S7 субтитри
  лишаються як на `main` (транскрипт уроку, одна мова) — обмеження §6 з механізмом.

## 2. Факти (перевірено 2026-11-14)

- hls.js **1.7.2** (`frontend/node_modules/hls.js/package.json`). `audio-track-controller.ts`:
  `set nextAudioTrack(id)` → `setAudioTrack(id, false)` (без негайного flush — мінімальний gap, S0);
  `setAudioOption(option)` пише `hls.config.audioPreference = option` і, якщо доріжка в активній групі,
  робить той самий `setAudioTrack(groupIndex, flushImmediate=false)`. `onManifestLoading` скидає
  `currentTrack`, а при оновленні групи `config.audioPreference` перемагає `DEFAULT=YES` (`findTrackId`,
  другий цикл матчить за name/lang без перевірки default). Отже перевага, встановлена через
  `setAudioOption`, **переживає `loadSource` того ж інстансу** (ретрай 503, нова сесія → новий `mediaPath`).
- Master (`backend/internal/media/adapters/ffmpeg/master.go:95-105`): одна група `group_aac`,
  `NAME="<label>"`, `LANGUAGE="<lang>"`, `DEFAULT=YES|NO`, `AUTOSELECT=YES`. Мова **точно** унікальна серед
  нефейлених доріжок ассета (`00026_audio_track_language_unique.sql:6`) → точний рядок мови = стабільна
  ідентичність доріжки між маніфестами; індекс hls — ні (новий master може додати доріжку).
- **Але матчинг hls.js — префіксний:** `languagesMatch` (`utils/rendition-helper.ts:388`) приймає, якщо один
  рядок — префікс іншого (`und` ↔ `und-x-original`, `en` ↔ `en-us`); `setAudioOption` повертається без
  перемикання, якщо поточна доріжка «матчить» (`audio-track-controller.ts:269`). Пара `und`/`und-x-original`
  реальна (`00028_audio_track_embedded_backfill.sql:42`). `NAME` теж не унікальний. Тому `setAudioOption` і
  `config.audioPreference` для вибору **не використовуються**.
- **Порядок у контролері:** `AUDIO_TRACKS_UPDATED` тригериться синхронно, потім контролер читає
  `this.trackId` і ставить свою (default) доріжку **лише якщо** `trackId === -1`
  (`audio-track-controller.ts:195-198`). Обробник, що синхронно присвоїть `hls.audioTrack = i`, вибирає
  доріжку **до** запиту плейлиста default. `onManifestLoading` скидає `trackId` на кожен `loadSource`
  (`:70-77`), тож подія приходить і після ретраю/нової сесії.
- Мова в бекенді: 1–35 символів, ASCII-літери **обох регістрів**, цифри, `-`, без нормалізації
  (`domain/audiotrack.go:267`, DB `00025_multi_audio.sql:13`).
- Плеєр: порт `PlaybackMedia` (`ports/PlaybackMedia.ts`) знає лише рівні якості; `hls-media.ts` віддає
  `onLevels` на `MANIFEST_PARSED`; native-гілка лише коли `Hls.isSupported()` = false. Gear-меню
  `PlayerSettingsMenu.tsx`: speed, quality (лише hls), captions, autoplay. Переваги —
  `application/preferences.ts`, ключ `courses.player.prefs.v1`, парсинг поле-за-полем з fallback.

## 3. Рішення

### D-S7-1. Порт: доріжки й перемикання — у `PlaybackMedia`; ідентичність = точна мова
`PlaybackAudioTrack = { language: string; label: string; isDefault: boolean }` (без індексу hls).
Колбеки: `onAudioTracks(tracks)` і `onAudioTrack(language)` — **обрана** доріжка (на
`AUDIO_TRACK_SWITCHING`, яку контролер тригерить синхронно при стартовому виборі й при реальному перемиканні,
`audio-track-controller.ts:347`; повторний вибір уже завантаженої поточної VOD-доріжки повертається раніше без
події, `:339` — і адаптер теж не кличе hls, тож радіо вже правильне). Метод `selectAudioTrack(language): boolean` — `false`, коли точної мови в списку немає.
Фабрика отримує стартову мову: `createMedia(video, callbacks, { audioLanguage })`.
- hls: адаптер тримає `wanted: string | null` (старт = опція; оновлюється успішним `selectAudioTrack`).
  Обробник `AUDIO_TRACKS_UPDATED`: індекс `i` з `hls.audioTracks[i].lang === wanted` (строге `===`); якщо
  знайдено — синхронно `hls.audioTrack = i` (до запиту default); інакше нічого (грає `DEFAULT=YES`, `wanted`
  не змінюється). Перемикання: точний `i`, `hls.nextAudioTrack = i` (`setAudioTrack(i, false)` — без
  негайного flush, S0); вибір уже обраної — no-op `true`.
- Ретрай (`load()` того ж інстансу) і нова сесія (`load(newMediaPath)`) → новий маніфест → той самий
  обробник знову застосовує `wanted`.
- native: не повідомляє доріжок, `selectAudioTrack` → `false`.
- Відкинуто: `setAudioOption`/`audioPreference` — префіксний матчинг обирає не ту доріжку (§2).

### D-S7-2. Перевага `audioLanguage: string | null` у `PlayerPreferences` (R1)
Адитивне поле в тому ж ключі `v1` (старі блоби без поля → `null`). Парсинг — та сама форма, що бекенд
(`domain/audiotrack.go:267`): 1–35 символів `[A-Za-z0-9-]`, **регістр зберігається**, порівняння точне;
інакше `null`. Пишеться **лише на явний вибір студента**, і лише коли `selectAudioTrack` повернув `true`;
автоматичний старт на default або відсутність мови в уроці перевагу не змінює. Вибір основної доріжки
теж явний → пишеться її мова. Плеєр тримає актуальну мову в ref і передає її в `createMedia` на кожне
створення інстансу (не `initialPrefs` з моменту монтування); вибір доріжки інстанс **не** перестворює.

### D-S7-3. UI: група «Аудіо» в gear-меню
- Показується лише для hls і коли доріжок ≥ 2 (одна доріжка — вибору немає, як quality для native).
- Значення радіо = мова з `onAudioTrack` (обрана в hls). Клік → `selectAudioTrack`; `false` → радіо
  лишається на обраній, перевага не пишеться. Окремого «pending» стану немає: `SWITCHING` синхронний; no-op
  повертає `true` без колбеку (unit-тест).
  Чутність нової доріжки (`AUDIO_TRACK_SWITCHED`, лише під час відтворення —
  `audio-stream-controller.ts:792/1114`) UI не чекає; її доводить e2e за частотою звуку (§4).
- Мітка опції = `NAME`; порожній `NAME` → мова; однакові мітки в уроці → `мітка (мова)`.
- Новий урок / новий інстанс media → список і поточна скидаються (як `qualities`).
- i18n uk/en: `learn.video.audio`. Розмітка — наявний `RadioGroup` (`presentation/ui/controls.tsx:113`).
  Мітки — React-текст (`<b>Dub</b>` лишається літералом). Панель `.ui-popover` уже `overflow: auto`
  (`tokens.css:262`); придатність при 200% і довгих мітках — браузерна перевірка (§4).

### D-S7-4. Субтитри не змінюються в S7 (R2)
`captionsUrl` і `<track>` — як на `main`. Багатомовні субтитри — окремий план після S7 (запис у
`work/AGENDA.md`/report як наступний напрям, не як відкритий зріз цієї фази).

## 4. Тести (доказ на production-шляху)

- **Unit `preferences.test.ts`:** `audioLanguage` round-trip; відсутнє/не-рядок/погана форма/занадто
  довге → `null`; старий блоб без поля → `null`, решта полів ціла.
- **Unit `LessonPlayer.test.tsx` (фейковий `createMedia`):** група прихована при 0/1 доріжці і для
  native; ≥2 → показана з мітками (включно з дедуплікацією); вибір кличе `selectAudioTrack(lang)` і
  пише перевагу; `onAudioTrack` змінює обране; `createMedia` отримує збережену мову; новий урок скидає.
- **Unit адаптера hls** (`hls-media.test.ts`, новий): чиста функція вибору індексу
  `exactAudioTrackIndex(tracks, language)` винесена й протестована на контрприкладах критики
  (`und`/`und-x-original` в обох порядках, `en`/`en-us`, `en-US` ≠ `en-us`, відсутня мова → `-1`, однакові
  `NAME`). Сам обробник на реальному hls.js доводить e2e (happy-dom без MSE).
- **e2e на реальному multi-audio HLS** (новий `audio-tracks.spec.ts`, проєкт `chromium`, бюджет тесту
  300 с; аплоад-хелпер як у `asr.spec.ts` ≤240 с). Нові фікстури в `media.setup.ts`:
  `e2e-640x360-30s.mp4` (вбудоване `sine 440`, 30 с) і `e2e-tone-30s.m4a` (`sine 880`, 30 с — та сама
  тривалість, щоб аудіо не закінчилось раніше за відео). Автор: API-seed курсу/уроку як `asr.spec.ts:69`, відео й
  доріжка — через UI-хелпери як `asr.spec.ts:137/153`; хелпер обирає мову за **ім'ям опції** UI, тож виклик —
  `addAudioTrack(page, 'Англійська', 'English dub', …)`, як `asr.spec.ts:441` (код `en`;
  регістр/префікси мов доводить unit адаптера, не e2e), публікація, грант студенту; `afterAll` видаляє курс
  (DELETE → 204, як `asr.spec.ts:250`; правило «хто додав глобально лічене — прибирає»). Слухачі запитів реєструються
  **до** дій. Студент: (1) gear показує «Аудіо» з двома опціями, обрано основну; (2) на `currentTime` ≤ 10 с
  вибір дубляжу → відповідь **200** на `…/audio-<trackId>/index.m3u8` і ≥1 сегмент цієї доріжки, радіо на
  дубляжі, `currentTime` росте без `waiting`; **доказ чутності** — частота декодованого аудіо елемента (не
  фізичного виходу): після кліку Play (жест — без нього `play()` = `NotAllowedError`) тест підключає
  `AnalyserNode` (`fftSize` 8192) через `video.captureStream()` → `AudioContext.createMediaStreamSource`; до
  `destination` **не** підключає (вихід плеєра не змінюється). Класифікація зразка: бін максимуму в межах
  ±12 Гц від цільової частоти **і** рівень ≥ −60 дБFS. Базова лінія: 3 зразки поспіль (крок 300 мс) ≈440 Гц,
  `state === 'running'`, інакше тест падає. Після кліку дубляжу — ≥ 5 зразків поспіль ≈880 Гц протягом ≤ 10 с.
  **Здійсненність виміряно 2026-11-14** (`_wip-s7-probe/`: ffmpeg-HLS з двома тонами, hls.js 1.7.2, Playwright
  1.62.1 Chromium headless, `devices['Desktop Chrome']`): до перемикання пік 439.5 Гц на −48…−38 дБ, після
  `nextAudioTrack` — 878.9 Гц від 1221 мс і стабільно до кінця (з 30 зразків після перемикання 4 ще 440 Гц, далі 26 — 880 Гц; відтворення — `_wip-s7-probe/README.md`), `currentTime` 2.8→11.6 с без
  зупинки. Поріг 10 с = ~8× виміряної затримки. Тест логує мс «клік → перший сегмент» (download latency) і
  «клік → пік 880 Гц» (audible-switch latency) — §6; (3) reload → **перший** запит аудіо-плейлиста — дубляжу (R1 застосовано до старту);
  (4) клавіатура: Tab до gear, Enter, стрілки між опціями «Аудіо», видимий фокус, Escape повертає фокус на
  gear; (5) урок з однією доріжкою (наявна фікстура) — групи немає. Повний e2e — наявні playback-тести
  лишаються зеленими. Фонове ASR нового уроку (тон → `no_speech`) не паузить черги: глобальний pause/resume
  (`asr.spec.ts:37`) не чіпаємо; `asr.spec.ts` за алфавітом іде раніше в тому ж одноворкерному проєкті.
- **План Б, якщо headless Chromium не дає звуку в `captureStream`** (вимірюється першим прогоном, не
  припускається): адаптер віддає `AUDIO_TRACK_SWITCHED` як `onAudioTrackPlaying(language)`, плеєр пише
  `data-audio-playing` на кадр, e2e чекає мову дубляжу. Перехід на план Б — датоване рішення з виміром у report
  і повторна критика цього пункту.
- **Браузер (design-pass, ручний):** gear-панель із трьома доріжками і довгою міткою при 200% тексту на
  ширині 375 px — усе досяжне прокруткою, без обрізання.

## 5. Порядок виконання

1. Порт + адаптер hls/native + перевага (unit). 2. UI-група + i18n + unit плеєра. 3. e2e-спека.
4. Codex-рев'ю хвилями → повний ordered `make check` → report → PR.

## 6. Обмеження (з механізмом)

- **Субтитри = транскрипт уроку незалежно від обраного аудіо** (R2): `captionsUrl` будується з одного
  транскрипта (`LessonPlayer.tsx:803`), а ASR/редактор одномовні (`asr_drafts` PK = `lesson_id`). Знімає
  окремий план «багатомовні субтитри».
- **Native-гілка без вибору:** вона активна лише коли `Hls.isSupported()` = false (немає MSE/
  ManagedMediaSource); там грає `DEFAULT=YES`. Перевірити: `createPlaybackMedia` у `hls-media.ts`.
- **Доріжка, додана під час перегляду,** видна з наступного маніфесту (батьківський §7).
- **Перемикання не миттєве:** `nextAudioTrack` не скидає поточний буфер одразу, але пізніше скидає
  forward-буфер аудіо (`audio-stream-controller.ts:931`) і чекає завантаження нової доріжки — затримка
  мережозалежна; e2e логує дві магнітуди на локальному Compose: download latency (клік → перший сегмент) і
  audible-switch latency (клік → пік 880 Гц). Мікс-фейду
  немає (батьківський §4/§7).

## 7. Відкат
Revert PR; перевага `audioLanguage` у localStorage ігнорується старим кодом (парсер читає лише відомі поля).

## 8. Історія критики
- **Раунд 1 (Codex, `_wip-s7-plan-critique-1.md`) — BLOCKED.** B1 префіксний матчинг мов у hls.js →
  D-S7-1 переписано на точний індекс + синхронний вибір в `AUDIO_TRACKS_UPDATED`; B2 валідація мови суперечила
  бекенду → D-S7-2 бере правило `domain/audiotrack.go:267`, регістр зберігається; N3 pending/відмова/
  перестворення → `selectAudioTrack: boolean`, радіо з `SWITCHING`, мова з ref; N4 e2e → 30-с фікстури, бюджет,
  cleanup, 200 + сегменти + ріст `currentTime`; N5 a11y і формулювання затримки → кроки (4) e2e, ручна
  перевірка 200%, §6 з механізмом.
- **Раунд 2 (Codex, `_wip-s7-plan-critique-2.md`) — BLOCKED.** B1/B2 закрито. Новий B: e2e могла пройти без
  того, щоб дубляж зазвучав → доказ за частотою виходу (440→880 Гц) з обов'язковою базовою лінією + план Б;
  N: SWITCHING не на кожен `setAudioTrack` → формулювання й unit no-op; N: `en-US` недосяжний UI-хелпером →
  e2e з `en`, регістр — у unit.
- **Раунд 3 (Codex, `_wip-s7-plan-critique-3.md`) — BLOCKED.** B: не доведено, що headless Chromium віддає MSE-
  аудіо в `captureStream` → **виміряно** (`_wip-s7-probe/run.log`, вище в §4): так; N: пороги класифікатора
  (±12 Гц, ≥ −60 дБFS, 3/5 зразків поспіль, ≤ 10 с, без `destination`) названо; N: хелпер бере ім'я опції →
  `'Англійська'`. План Б лишається запасним і не активний.
- **Раунд 4 (Codex, `_wip-s7-plan-critique-4.md`) — ACCEPTED**, 0 blocker. N: відтворюваність заміру → `_wip-s7-probe/
  README.md` (команда фікстури, версія браузера, хеші); N: «30 зразків» → 4 + 26 — виправлено.
