ДАЛІ: Створюю A01 SMTP-конфігурацію, `platform/mail` Sender, її unit/integration тести та Mailpit у Compose; кожен завершений крок фіксуватиму нижче.
ФАКТ: WIP-файл, вказаний у брифі, був відсутній на момент початку. Його створено; наявні не пов’язані зміни робочого дерева не чіпаю.
ФАКТ: Перевірено патерни: `config.endpoint` уже централізує host/port валідацію; integration-тести мають тег `integration` і обгортають створення testcontainers у `testsupport/containers.Start`. `api` у Compose ще не має SMTP env.
РІШЕННЯ: `platform/mail` буде stdlib SMTP-адаптером з конструктором від `config.MailConfig`; MIME-кодування винесене у чисту непублічну функцію для unit-тесту. Для зовнішніх заголовків/адрес забороняю CR/LF, щоб не допустити header injection.
ДАЛІ: Напишу спершу unit-тести конфігурації та MIME-кодування, а потім мінімальну реалізацію, що їх задовольняє.
ФАКТ: Додано червоні unit-тести `platform/config/mail_test.go` і `platform/mail/sender_test.go`: дефолт/валідація SMTP, MIME multipart/alternative, UTF-8 subject та відмова від header injection.
ФАКТ: Перший запуск `go test` заблокував sandbox за межами workspace у стандартному Go build cache, до компіляції тестів не дійшов. Наступні Go-команди виконуватиму з `GOCACHE=/tmp/courses-go-build-cache`.
ДАЛІ: Реалізую `MailConfig` і Sender на stdlib, після чого повторю unit-тести з локальним cache.
ФАКТ: Реалізовано `config.MailConfig`, `config.SMTPConfig` і `mail.Sender`: endpoint через `net.JoinHostPort`, контекстний `net.Dialer`, SMTP без auth, RFC 5322/MIME multipart/alternative з quoted-printable UTF-8 частинами, явні помилки без логування даних листа.
РІШЕННЯ: Буквальна сигнатура з брифу `MailConfig(getenv) (MailConfig, error)` некомпільована в Go, бо функція й тип ділять namespace. Збережено потрібну функцію `MailConfig`; її результат названо `SMTPConfig`, що лишає місце для майбутніх username/password.
ФАКТ: `GOCACHE=/tmp/courses-go-build-cache go test ./platform/config ./platform/mail` — PASS.
ДАЛІ: Додам tagged integration-тест із `testsupport/containers.Start`, перевірю Mailpit HTTP API на фактично прийнятому листі.
ФАКТ: Додано `sender_integration_test.go` з тегом `integration`: контейнер Mailpit `v1.31.2` запускається тільки через `containers.Start`, лист надсилається SMTP, а `GET /api/v1/messages` і `GET /api/v1/message/{id}` перевіряють recipient, subject, text та HTML.
ФАКТ: `go test -tags=integration ./platform/mail` компілює тест, але запуск зупинено середовищем: sandbox не має доступу до Docker sockets (`permission denied` на `/var/run/docker.sock` і `~/.docker/run/docker.sock`). Це не позначаю як пройдений integration gate.
ФАКТ: Додано Mailpit `axllent/mailpit:v1.31.2` у базовий Compose з перевіркою `/api/v1/messages` на 8025; `api` отримує SMTP env і залежить від healthy Mailpit. E2E override явно зберігає ті самі env/depends_on і сервіс `mailpit`.
РІШЕННЯ: Обрано pin `v1.31.2`, а не рухомий тег; healthcheck робить HTTP-запит саме до API на 8025.
ДАЛІ: Запущу build/vet/unit-перевірки й валідацію Compose; за потреби виправлю знайдене, після чого перегляну diff без коміту.
ФАКТ: Перевірки після фінального виправлення: `go build ./...` — PASS; `go vet ./...` — PASS; `go test ./platform/config/... ./platform/mail/...` — PASS; `docker compose -f docker-compose.yml -f docker-compose.e2e.yml config` — PASS; `git diff --check` — PASS.
ФАКТ: Під час локального review виправлено один суттєвий крайовий випадок: скасування `context.Context` без deadline тепер перериває вже відкритий SMTP socket через `context.AfterFunc` + immediate deadline, а не діє лише на dial.
ФАКТ: Integration-тест скомпілювався; його реальне виконання залишається заблокованим відсутнім доступом sandbox до Docker daemon, що зафіксовано вище. Compose-стек у цьому середовищі з тієї ж причини не підіймався.
РІШЕННЯ: Змін не комічено — усі A01-артефакти лишено в робочому дереві для Courses 1. Не пов’язані файли/скриншоти не змінювались.
ДАЛІ: A01 готовий до перевірки Courses 1 у Docker-доступному середовищі: виконати `make integration` або щонайменше `go test -tags=integration ./platform/mail`, а також `docker compose up --wait`.
