# Припущення та журнал рішень foundation

## Вихідний стан

- 2026-09-03: робоча директорія `/Users/strelia/projects/courses`; на початку огляду файлів репозиторію та `AGENTS.md` немає. Сторонні дані не видаляються.
- Існуючі модулі, індекс повторного використання та code graph відсутні. Нова структура необхідна для цього репозиторію.
- `spec-flow` і `acceptance-spec` прочитані. Manifest стандартів і persist configuration створює основний агент. Цей документ є durable записом припущень, `work/ledger.md` — виконання.
- Користувач прямо задав рішення й дозволив не блокувати роботу питаннями за наявності безпечного оборотного вибору. Повторне grilling-підтвердження цих рішень не потрібне. Незалежні технічні gates залишаються обов'язковими.

## Обрані припущення

| Питання | Рішення | Причина / межа |
|---|---|---|
| Обсяг | Phase 1 та Phase 2; Phase 3–8 лише roadmap | Foundation DoD включає перший публічний Course slice |
| Runtime | Go 1.27.1 у контейнері, Node 24, PostgreSQL 17, Redis 7 | Конкретний baseline; доступність образів і сумісність перевіряються до green gate |
| Frontend dependency management | npm, точні версії та committed lockfile | Відтворюваний `npm ci`; нова бібліотека має поточне використання |
| SSR | TanStack Start + Rsbuild plugin; client і server builds | Заданий користувачем стек; без Vite/Next.js fallback |
| API contract | OpenAPI 3.0.3 у backend; Orval native fetch | Контракт версіонований, generated DTO ізольовані й перевірені |
| Course identity | UUID id; lowercase ASCII slug з цифрами та одиночними внутрішніми дефісами, максимум 120 символів | Однозначні URL; malformed slug → 400, відсутній валідний slug → 404 |
| Catalog | limit 1–50 (default 20), offset 0–10000 (default 0), q максимум 100 Unicode code points, sort `recent` або `title` (default `recent`) | Обмеження вартості запиту та відтворювана адреса; q trims surrounding whitespace |
| Catalog ordering | `recent`: published_at DESC, id ASC; `title`: title ASC, id ASC; явна PostgreSQL collation у запиті | Стабільність на ties; offset pagination може зміщуватися між окремими запитами під час публікації |
| Catalog consistency | List і count в одній read-only repeatable-read transaction | Узгоджений snapshot в межах відповіді; не обіцяє snapshot між сторінками |
| Demo content | Окрема явна idempotent seed-команда; міграція без fixtures | Схема не публікує демонстраційні курси автоматично |
| Media foundation | Nginx `/media/*` fail-closed; worker Asynq запускається без handlers і має статус/health | Немає доступу до storage або удаваного processing до Phase 5–6 |
| Storage | Named media volume, Nginx read-only, worker writable; file layout задокументовано | Реальний storage adapter і atomic publishing з'являться з upload use case |
| HLS майбутній формат | H.264/AAC, MPEG-TS `.ts` segments; 360p/720p, 1080p лише за достатнього source | Найпростіший початковий pipeline; немає заяви DRM |
| Login/admin/learn | `/login` public SSR інформаційна сторінка; `/admin/*` та `/learn/*` client fail-closed placeholders | Немає fake auth, захищених даних чи бізнес-логіки в route components |
| Caching | Public success data only; жодного shared кешу сесій або errors; CDN cache disabled foundation | Уникає витоку персонального стану до появи identity policy |
| Telemetry | JSON logs, request/trace IDs, HTTP status/duration, DB duration, queue depth/status; internal metrics ports | Playback/transcoding метрики додаються разом із відповідними use cases, не вигадуються |
| Rollback | Stop/revert images/config, volumes retained; destructive migration down лише ephemeral test DB | Жодного автоматичного видалення користувацьких даних |

## Відкладені залежності

Не встановлювати лише для переліку: React Hook Form, Zod для складних форм (якщо не потрібен для наявного contract adapter), Zustand, hls.js, TanStack Table, TanStack Virtual, TipTap, ECharts, i18next/react-i18next, Sentry adapter і frontend OpenTelemetry SDK, Argon2id. Кожна додається з реальним сценарієм та власним gate. Оркестрація worker використовує Asynq зараз; FFmpeg/ffprobe потрібні образу майбутнього worker, але foundation не має production transcoding pipeline. PaymentGateway і transcription ports додаються разом із реальним application caller; затверджені майбутні контракти поки живуть у документації, без порожніх runtime interfaces.

## Зміни припущень

Поки немає. Після перевірки версій або незалежного review основний агент дописує тут факт, причину, ADR/файл і перевірку; початкові рішення не стираються.
