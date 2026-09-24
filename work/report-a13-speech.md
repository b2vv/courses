# Report A13 / крок 3 — рушій на справжньому мовленні і speech-фікстура

Дата: 2026-09-24

## Зроблено

- У `backend/internal/transcript/adapters/whisper/transcribe.py` виправлено назву поля faster-whisper з `avg_log_prob` на `avg_logprob` у коментарі та в обох входженнях коду.
- У `infrastructure/tests/smoke.sh` збережено синусоїдну перевірку `no_speech` і додано окремий прогін на WAV, синтезованому `flite`. Перевіряються успішний exit, валідний JSON, `speech_seconds >= 1` та щонайменше один сегмент із непорожнім `text`. Обидва тимчасові WAV прибираються trap-ом.
- У `frontend/e2e/media.setup.ts` додано `/tmp/e2e-speech-640x360.mp4` із `testsrc2` та тією самою `flite`-фразою, а також його вивантаження через наявний цикл `cat` із перевіркою ненульового розміру. Наявні синусоїдні фікстури не змінено.

## Перевірки

- `sh -n infrastructure/tests/smoke.sh` — пройдено.
- `frontend: npm run typecheck` — пройдено.
- `frontend: npm run lint` — пройдено.

## Не запускалося

- `smoke.sh`, Docker Compose, перезбірка образів і `media.setup.ts` — не запускалися: Docker недоступний у середовищі, а бриф прямо забороняє ці запуски.
- Повний ordered gate та e2e — не запускалися з тієї ж причини; це не заявляється як пройдене.

Git не змінювався.
