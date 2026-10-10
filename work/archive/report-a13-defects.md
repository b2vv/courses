# Звіт A13 / дефекти C10 і C14

Дата: 2026-09-24

## Зроблено

- Додано доменну помилку `already_running`.
- `Regenerate` читає поточну чернетку: для поточних `assetID:generation` у стані `running` повертає `ErrAlreadyRunning` і не ставить задачу в чергу.
- Task ID asynq тепер має форму `assetID:generation:marker`. Маркер — стабілізований `updated_at` до мікросекунд або `none`; той самий конструктор використовується для ручної та автоматичної постановки.
- Конфлікт `asynq.ErrTaskIDConflict` більше не є тихим успіхом: він мапиться на `ErrAlreadyRunning`.
- HTTP повертає `409` з кодом `already_running`.
- Таймаут під час `audio.Extract` або `Transcribe` завершує чернетку як `failed` з причиною `timeout`. Запис виконується через `context.WithoutCancel` і короткий 5-секундний timeout. `context.Canceled` лишається graceful-shutdown шляхом без фейлу чернетки.

## Тести

- Додано unit-тести для task ID, конфлікту черги, `Regenerate` на running-чернетці та HTTP-контракту `409/already_running`.
- Додано unit-тести для timeout на обох фазах і перевірку, що terminal write не використовує протермінований context.
- Додано Redis integration harness: конфлікт однакового маркера, повторна постановка після completed key і після archived key.
- Додано PostgreSQL integration harness для перевірки `failed/timeout`, а не `running`, після дедлайну.

## Перевірки

- `gofmt` — пройшов.
- `GOCACHE=/private/tmp/courses-gocache go test ./internal/transcript/...` — пройшов.
- `GOCACHE=/private/tmp/courses-gocache go test -tags integration -run '^$' ./internal/transcript/adapters/asynq ./internal/transcript/adapters/postgres` — integration-тести скомпільовані, але не виконані.
- `GOCACHE=/private/tmp/courses-gocache go build ./...` — пройшов.
- `GOCACHE=/private/tmp/courses-gocache go test ./...` — не пройшов лише `platform/httpserver/TestShutdownDrainsAnInFlightResponse`: sandbox заборонив bind на `127.0.0.1:0` (`operation not permitted`); інші пакети пройшли.

Redis/PostgreSQL integration, Docker-залежні тести та ordered gate **не запускалися** через обмеження середовища без Docker. Це не позначено як пройдене.

## Обмеження

Integration-тест C14 використовує короткий контрольований context deadline, щоб не чекати продуктивний мінімум `TaskTimeout` у 300 секунд; PostgreSQL перевіряє саме запис terminal timeout у реальну чернетку. Перевірку повного queue-derived deadline має виконати reviewing session разом із Docker integration gate.

## Правка після рев’ю

`EnqueuePending` виконується періодично, тому вже наявний task ID для pending-асета є очікуваним станом, а не помилкою. Додано спеціальне пропускання `domain.ErrAlreadyRunning` без `Error`-логу; інші помилки й далі логуються як раніше. Додано unit-тест: перший pending повертає `ErrAlreadyRunning`, другий обробляється в тому самому sweep.

Перевірено: `gofmt` і `GOCACHE=/private/tmp/courses-gocache go test ./internal/transcript/...` — пройшли. Git не змінювався.
