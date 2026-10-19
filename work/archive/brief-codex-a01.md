# Codex brief — A01: email-інфраструктура (`platform/mail` + Mailpit)

Репо `courses`, гілка `feat/accounts-selfservice`. Фаза `accounts` (spec/plan у `docs/accounts/`).
Роль: ти реалізуєш A01; план/гейти/коміти — Courses 1; рев'ю — Courses 2.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a01-codex.md
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — ДО першого інструменту: рядок `ДАЛІ:`
з тим, що робитимеш. Далі ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінальна відповідь = стислий переказ файлу.

## Що зробити (A01 — лише транспорт email + Mailpit; auth-шаблони й порти — пізніші таски)

1. **`backend/platform/config` — SMTP-конфіг.** Додай `MailConfig(getenv Getenv) (MailConfig, error)` у стилі
   наявного `dsn.go` (той самий функціональний `Getenv`, ті самі правила валідації й помилок). Поля з env:
   `SMTP_HOST`, `SMTP_PORT` (дефолт 1025), `MAIL_FROM` (валідуй як email). Без паролів/TLS (Mailpit локальний,
   без авторизації) — але структуру лиши розширюваною під майбутні `SMTP_USERNAME/PASSWORD`. Юніт-тест
   `*_test.go` на валідні/невалідні входи (як `dsn_test.go`).

2. **`backend/platform/mail/` — Sender (SMTP-транспорт).** Новий пакет shared-інфри:
   - Тип `Message { To, Subject, TextBody, HTMLBody string }` (multipart/alternative, text+html).
   - Тип `Sender` із методом `Send(ctx context.Context, msg Message) error`, що кодує лист (RFC 5322 заголовки,
     коректний MIME multipart/alternative, UTF-8) і шле через `net/smtp` на endpoint із конфігу (без auth).
   - **Жодних секретів у логах**; помилки — явні, обгорнуті.
   - Поважай `ctx` (дедлайн/скасування) — хоч `net/smtp.SendMail` контексту не бере, оберни діал у dialer
     із дедлайном з ctx.
   - **Без нових залежностей** (stdlib `net/smtp`, `mime/multipart`, `net/textproto`). Шаблонів auth тут НЕ
     роби — це A03; за потреби додай крихітний generic-хелпер рендеру `text/template`, не прив'язаний до auth.

3. **Тести Sender:**
   - **Юніт** — кодування Message (MIME-структура, заголовки, екранування) без мережі; таблиця.
   - **Інтеграція** (`*_integration_test.go`, під `make integration`, як repository-тести) — підніми Mailpit
     контейнером через наявний `internal/testsupport/containers.Start` (generic testcontainers-go, образ
     `axllent/mailpit`), відправ лист, прочитай назад через Mailpit HTTP API (`GET /api/v1/messages`,
     `GET /api/v1/message/{id}`), перевір To/Subject/тіло. Дотримуйся наявних патернів тегів/збірки
     integration-тестів у репо.

4. **Compose — Mailpit.** Додай сервіс `mailpit` (образ `axllent/mailpit`, фіксована версія-тег) у
   `docker-compose.yml` і `docker-compose.e2e.yml`: SMTP 1025, web/API 8025, healthcheck на 8025. Сервісу
   `api` (і де треба) додай env `SMTP_HOST=mailpit`, `SMTP_PORT=1025`, `MAIL_FROM=no-reply@courseforge.local`.
   Не ламай наявний стек (depends_on/healthcheck як у інших сервісів).

## Правила (AGENTS / конституція)
- Domain-safe межі: `platform/mail` — shared-інфра, не імпортує жодного модуля. Адаптери всередину.
- TS не чіпаємо (A01 суто бекенд + compose).
- Маленькі логічні коміти НЕ роби сам — лишай робоче дерево Courses 1 для коміту; просто познач у робочому
  файлі, що готово.
- TDD: тест → реалізація → рефактор. Якщо впираєшся в ліміт часу — спершу дороби тести вже-написаного коду,
  не лишай код без тестів; познач у робочому файлі, що лишилось.

## Критерій готовності A01
- `go build ./...` і `go vet` чисті; `gofmt` чистий.
- `go test ./platform/config/... ./platform/mail/...` — юніт зелені.
- Integration-тест Mailpit компілюється й проходить під наявним integration-харнесом (не вигадуй окремий).
- Compose валідний (`docker compose config` ок), стек піднімається з Mailpit.
- Жодних секретів у логах; жодних нових залежностей поза testcontainers (уже є).
