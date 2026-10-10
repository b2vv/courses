# Brief A13 / крок 1 — фікс model-path (asr-worker)

Частина A13 (закрити приймання Phase 8 до 22/22). Це перший із кроків; наступні
(C10/C14, speech-фікстур, wiring, e2e, integration) — окремі брифи. Тримайся ЛИШЕ цього.

## Read first
- `docs/asr/spec.md` §11 — **D8-14**, **D8-14a** (ваги `large-v3-turbo` **всередині образу**,
  рантайм `faster-whisper`/CTranslate2, `mem_limit ≥ 5 ГіБ`) і §12 рядок **C16**
  («`/readyz` воркера червоний, поки файл моделі недоступний»).
- `backend/cmd/asr-worker/main.go` — `modelReady` (рядки ~148-160), `NewEngine` (~55-60).
- `backend/internal/transcript/adapters/whisper/engine.go` рядки 70-75 — як спавниться
  `transcribe.py`: `--model <engine.model>`, де `engine.model = WHISPER_MODEL_DIR`.
- `backend/internal/transcript/adapters/whisper/transcribe.py` — контракт (аргумент `--model`
  іде прямо у `WhisperModel(args.model, ...)`).
- `backend/Dockerfile` рядки ~48-64 — стадія `asr-worker` (там запікається модель).
- `docker-compose.yml` — сервіс `asr-worker`, змінна `WHISPER_MODEL_DIR`.
- `infrastructure/tests/smoke.sh` — блок asr (~рядки 13-27).

## Проблема (перевірено в живому контейнері 54fc22c — це ФАКТИ, не гіпотези)
Движок спавнить `python3 /opt/whisper/transcribe.py --model $WHISPER_MODEL_DIR --audio ...`
(`engine.go:70`). `WHISPER_MODEL_DIR=/opt/whisper/models` (`main.go:58`, `compose`,
`Dockerfile` `ENV HF_HOME=/opt/whisper/models`). Але `/opt/whisper/models` — це **HF_HOME-корінь**:
містить лише `hub/` і `xet/`, **без плаского `model.bin`**. Прямий прогін як у движка:

```
WhisperModel("/opt/whisper/models", device="cpu", compute_type="int8")
→ RuntimeError: Unable to open file 'model.bin' in model '/opt/whisper/models'
```

Наслідок: **жодна реальна транскрипція не проходить**; усі e2e-рядки (C03/C06/C11/C13/C17)
блоковані на завантаженні моделі (модель вантажиться ПЕРШОЮ, до VAD). `modelReady` робить лише
`os.Stat` на теку → `/readyz` зелений при зламаному рушії, що **порушує C16**.

Перевірені факти (використай, НЕ передоказуй):
- `WhisperModel(<тека, де лежить model.bin>)` вантажиться **ОК** (ваги в образі є — зламаний
  саме аргумент шляху).
- `faster_whisper.utils.download_model(size_or_id, output_dir=None, local_files_only=False,
  cache_dir=None, revision=None, use_auth_token=None)` **існує** у faster-whisper 1.2.1.
- `large-v3-turbo` → HF repo `mobiuslabsgmbh/faster-whisper-large-v3-turbo`; плаский набір
  файлів: `model.bin`, `config.json`, `tokenizer.json`, `vocabulary.json`,
  `preprocessor_config.json`.

## Задача
1. **Dockerfile (стадія `asr-worker`):** запекти ваги у **стабільну пласку теку** з реальним
   `model.bin` за шляхом, що **не залежить від revision-хешу** (наприклад
   `/opt/whisper/models/large-v3-turbo`). Годиться
   `download_model('large-v3-turbo', output_dir='/opt/whisper/models/large-v3-turbo')`. Ваги
   лишаються всередині образу (D8-14). Тека має бути **самодостатньою** — `model.bin` читається
   в рантаймі (символьні лінки в межах образу прийнятні, дангліючі — ні).
2. **`WHISPER_MODEL_DIR`** → нова пласка тека у Dockerfile `ENV` **і** в `docker-compose.yml`
   (сервіс `asr-worker`). Значення мають збігатися; одне джерело правди.
3. **`main.go` `modelReady`:** перевіряти наявність **файла моделі** всередині теки
   (`filepath.Join(dir, "model.bin")` через `os.Stat`), а не лише що тека існує. Стиль помилок
   лишити (`errors.New("model weights not available")`). Це закриває C16 по суті: відсутній файл
   моделі → `/readyz` червоний. (Стеж, щоб `os.Stat` файла, а не теки; імпорт `path/filepath`
   за потреби.)
4. **`smoke.sh` (блок asr):** додати **один справжній прогін движка** — довести, що модель
   вантажиться через той самий `--model`, що й у проді. Згенеруй короткий (~1с) 16 kHz mono sine
   wav у контейнері `asr-worker` (`ffmpeg` там є), запусти
   `/opt/whisper/bin/python3 /opt/whisper/transcribe.py --model "$WHISPER_MODEL_DIR" --audio <wav>`
   і перевір: **exit 0** І stdout — валідний JSON із числовим полем `speech_seconds`. Синусоїди
   достатньо (доводить завантаження моделі); реальний speech-сегмент тут НЕ потрібен. Дотримуйся
   стилю наявного smoke.sh (`docker compose exec -T asr-worker ...`, `set -e`, зрозуміле
   повідомлення при провалі).

## NOT (не твоє в цьому кроці)
- НЕ чіпай код C10/C14: `recognize.go`, `application/draft.go`, `adapters/asynq/queue.go`,
  `adapters/http/draft_handler.go` — окремий бриф.
- НЕ чіпай фронтенд: `CourseEditorView.tsx`, `LessonTranscriptField.tsx`, `TranscriptEditor.tsx`,
  локалі — крок wiring.
- НЕ додавай speech-фікстур, e2e-тести чи integration-тести (C21/C22/C09/C01) — окремі кроки.
- НЕ чіпай `mem_limit` (5 ГіБ уже стоїть), не змінюй базу образу, не чіпай `media-worker`.

## Обмеження середовища
- Пісочниця **без Docker**: ти НЕ можеш зібрати образ, запустити smoke чи будь-який гейт.
  Напиши код і в звіті **прямо познач**, що збірку образу / smoke / ordered gate НЕ запускав —
  їх жену я. **Ніколи не став «пройдено» те, чого не ганяв** (назви відмову, не обходь її).
- Не чіпай git; не редагуй `work/tasks.md`; не пиши `docs/asr/report.md`; не пиши в `generated/`.

## Якщо бачиш неточність, що дасть баг
Відкрий секцію `## Discussion` у звіті — що не так, доказ (file:line, що станеться в рантаймі),
пропонована правка, і що зробив тим часом (спинився / зробив із позначкою розбіжності). Не
обходь мовчки і не виконуй проти власного судження.

## Звіт
`work/report-a13-model-path.md`
