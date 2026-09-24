# Brief A13 / крок 3 — рушій на справжньому мовленні і speech-фікстура

Частина A13 (Phase 8): закрити приймання ASR. Кроки 1 (`23fe407`, шлях до моделі) і 2 (`62f3993`,
C10/C04/C14) закомічені. Бриф — сесія Courses 1, 2026-09-24; рецензент і гейти — Courses 2.

## Навіщо

Після виправлення шляху до моделі рушій уперше дійшов до справжнього мовлення — і впав. Доведено
в живому `asr-worker` (`23fe407`) тим самим шляхом, що в рушії
(`transcribe.py --model $WHISPER_MODEL_DIR --audio …`):

- синусоїда 2 с → `{"speech_seconds": 0.0, "segments": []}`, exit 0;
- справжнє мовлення → `AttributeError: 'Segment' object has no attribute 'avg_log_prob'. Did you
  mean: 'avg_logprob'?`. У faster-whisper 1.2.1 `Segment` — dataclass, його поле зветься
  `avg_logprob` (перевірено `dataclasses.fields(Segment)` і `hasattr` в образі; Courses 2 відтворила). Отже будь-яке розпізнавання з мовленням закінчується `engine_error`,
  і чернетки з текстом продакшн ніколи не створював.
- На копії скрипта з `avg_logprob` фраза, синтезована `flite` (4,3 с), розпізнається дослівно:
  `"Welcome to the course, today we will learn how video transcripts work."`, confidence 0.8522,
  exit 0, ~7 с на CPU.

Smoke із кроку 1 цього не ловить: він ганяє синусоїду, сегментів немає, цикл по сегментах не
виконується. E2e-рядкам C03, C06, C11 і C17 потрібна відео-фікстура з мовленням.

## Read first
- `backend/internal/transcript/adapters/whisper/transcribe.py` — весь файл (цикл сегментів :61-75).
- `backend/internal/transcript/adapters/whisper/engine.go` — як читається JSON скрипта.
- `infrastructure/tests/smoke.sh` — ASR-блок із кроку 1 (~рядки 25-55): синусоїда → `transcribe.py`.
- `frontend/e2e/media.setup.ts` — як генеруються й виносяться з tmpfs наявні фікстури (:52-141).
- `.gitignore` — `/frontend/playwright/.fixtures/` уже ігнорується.

## Зроби

1. **`transcribe.py`:** `avg_log_prob` → `avg_logprob` (обидва входження й коментар :62). Інших змін
   у скрипті немає.

2. **`smoke.sh`: друга перевірка — на справжньому мовленні.** Синусоїдну не прибирай: вона доводить
   `no_speech`. Додай поруч прогін на синтезованій фразі в тому самому контейнері `asr-worker`:
   `ffmpeg -nostdin -loglevel error -f lavfi -i "flite=text=Welcome to the course. Today we will
   learn how video transcripts work.:voice=slt" -ar 16000 -ac 1 -y <tmp>.wav` (ffmpeg в образі
   зібраний з `--enable-libflite`, перевірено `ffmpeg -filters`). Далі той самий виклик
   `transcribe.py --model "$WHISPER_MODEL_DIR"`. Вимагай exit 0, валідний JSON, **≥1 сегмент із
   непорожнім `text`** і `speech_seconds >= 1`. Тимчасові файли прибирай так само, як синусоїду.
   Жодної перевірки на точний текст: модель не мусить повторювати слово в слово, і smoke від цього
   став би крихким.
   **Квотування:** у filtergraph ffmpeg кома розділяє фільтри, а двокрапка — опції. Тому значення
   `text` для `flite` бери в одинарні лапки всередині filtergraph (`flite=text='…':voice=slt`) і
   тримай фразу без ком. Квотуй однаково в `smoke.sh` і в `media.setup.ts` — у TS аргумент іде в
   `execFile` без shell, тож лапки там — частина рядка filtergraph, а не shell-синтаксис.

3. **`media.setup.ts`: speech-фікстура.** Поряд із трьома двосекундними згенеруй у `media-worker`
   `/tmp/e2e-speech-640x360.mp4`: відео `testsrc2=size=640x360:rate=15`, аудіо
   `flite=text=<та сама фраза>:voice=slt`, `-shortest`, `-c:v libx264 -preset veryfast -c:a aac`.
   Винеси її тим самим циклом `cat` через `exec` у `.fixtures/e2e-speech-640x360.mp4` і перевір, що
   розмір > 0. Звичайні фікстури не чіпай: синусоїда в `e2e-640x360.mp4` лишається для C13.

## NOT
- Не пиши e2e-тести й не чіпай фронтенд `src/` — це крок 4.
- Не чіпай `Dockerfile`, compose, `main.go`, Go-код рушія, `openapi.yaml`.
- Не додавай `say`, espeak-ng і нових пакетів в образи: `flite` уже є.
- Не редагуй `work/tasks.md`, `docs/asr/report.md`, `generated/`. Git не чіпай.

## Обмеження середовища
Docker у пісочниці немає: `smoke.sh`, перезбірку образу й `media.setup.ts` ти **не запускаєш**.
Перевір синтаксис (`sh -n infrastructure/tests/smoke.sh`, `npm run typecheck` і `npm run lint` у
`frontend`, якщо доступні) і прямо познач у звіті, що не ганяв. Ніколи не став «пройдено» того,
чого не запускав.

## Якщо бачиш неточність, що дасть баг
`## Discussion` у звіті: що не так, file:line, що станеться в рантаймі, пропозиція.

## Звіт
`work/report-a13-speech.md`
