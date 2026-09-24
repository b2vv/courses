# A13 / крок 1 — звіт про model-path

Дата: 2026-09-24

## Результат

- `asr-worker` запікає `large-v3-turbo` через `download_model` у стабільну пласку теку
  `/opt/whisper/models/large-v3-turbo`; збірка додатково вимагає наявності файла `model.bin`
  через `test -f`, а `chmod -R a+rX` робить його читабельним у runtime.
- `WHISPER_MODEL_DIR` однаково вказує на цю теку в образі, Compose і fallback композиційного
  кореня воркера.
- `/readyz` перевіряє регулярний файл `<WHISPER_MODEL_DIR>/model.bin`, а не саму теку. Відсутній
  або непридатний файл повертає `model weights not available`.
- Smoke генерує одноразовий 1-секундний 16 kHz mono sine WAV у `/tmp`, прибирає його через
  `trap`, запускає той самий продакшн-шлях
  `/opt/whisper/bin/python3 /opt/whisper/transcribe.py --model "$WHISPER_MODEL_DIR"` і вимагає
  exit 0 та валідний JSON із числовим `speech_seconds`. `NaN`/`Infinity` явно відхиляються як
  невалідний JSON.

## Змінені файли

- `backend/Dockerfile`
- `backend/cmd/asr-worker/main.go`
- `docker-compose.yml`
- `infrastructure/tests/smoke.sh`

Заборонені брифом C10/C14-файли, фронтенд, speech-фікстури, e2e/integration-тести,
`work/tasks.md`, `docs/asr/report.md` і `generated/` не редагував. Наявні чужі незакомічені
зміни в них збережені.

## Перевірки

Виконано:

- `gofmt` для `backend/cmd/asr-worker/main.go`.
- `sh -n infrastructure/tests/smoke.sh` — exit 0 після остаточної правки.
- Двоосьовий огляд змін проти брифу й стандартів репозиторію: блокуючих знахідок не лишилось.
  Специфікаційний огляд знайшов приймання `NaN`/`Infinity` стандартним `json.load`; валідатор
  виправлено. Огляд стандартів порушень не знайшов.

Не запускав відповідно до секції «Обмеження середовища» (у пісочниці немає Docker, а бриф
прямо забороняє запускати будь-який гейт):

- збірку образу `asr-worker`;
- реальний smoke, включно із завантаженням моделі та прогоном `transcribe.py`;
- ordered gate / `make check`: `make format-check`, `make lint`, `make typecheck`, `make test`,
  `make integration`, `make check-generated`, `make build`, `make up`, `make drain-queue`,
  `make seed-identities`, `make seed`, `make smoke`, `make e2e`.

Жодну з цих незапущених перевірок не позначено як пройдену. Git-стан не змінював: комітів,
стейджингу чи інших git-операцій над деревом не робив.
