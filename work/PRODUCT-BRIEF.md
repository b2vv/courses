# Product brief — платформа продажу онлайн-курсів

> Джерело: початкове ТЗ від замовника. Це робочий документ — модифікуємо його в міру
> уточнення scope. Де він розходиться з кодом, перемагає код, а брифінг переписується.

Ти працюєш як senior software architect та coding agent. Потрібно створити production-oriented фундамент платформи продажу онлайн-курсів із власним відеосховищем, HLS streaming, транскрипціями, кабінетом студента та адміністративною панеллю.

Не намагайся реалізувати весь продукт одним великим кроком. Спочатку проаналізуй вимоги, створи план із малих перевірюваних задач, після чого послідовно реалізуй фундамент. Після кожного етапу запускай відповідні перевірки.

## 1. Цільовий результат

Створити monorepo courses:

```
courses/
├── frontend/
├── backend/
├── infrastructure/
├── docs/
├── docker-compose.yml
├── Makefile
├── .env.example
├── README.md
└── AGENTS.md
```

Система повинна містити:

* публічний каталог курсів;
* SSR сторінки каталогу та курсу;
* кабінет студента;
* HLS video player;
* транскрипцію із прив'язкою тексту до часу;
* посилання на документацію для сегментів транскрипції;
* адміністративну панель;
* керування курсами, модулями, уроками й відео;
* Go API;
* PostgreSQL;
* Redis;
* FFmpeg media worker;
* Nginx для HLS data plane;
* локальне файлове сховище;
* можливість майбутньої заміни файлової системи на MinIO/S3 без зміни domain/application layers.

Поточний етап — фундамент і перший вертикальний slice, а не повна production-платформа.

## 2. Архітектурні принципи

Використовуй hexagonal architecture окремо для frontend і backend.

Обов'язкові правила:

1. Domain layer не залежить від framework, transport, database, React або зовнішніх бібліотек.
2. Application layer містить use cases і залежить тільки від domain та ports.
3. Ports описують контракти взаємодії із зовнішнім світом.
4. Adapters реалізують ports.
5. Infrastructure відповідає за composition root і підключення adapters.
6. HTTP handlers і React route components не містять бізнес-логіки.
7. PostgreSQL models не використовуються як domain entities.
8. Generated API types не проникають безпосередньо в domain.
9. Бізнес-помилки моделюються явно, а не через випадкові exception strings.
10. Не створюй мікросервіси без реальної необхідності.
11. Backend починається як modular monolith.
12. Media worker є окремим executable, але використовує спільні backend modules і ports.
13. Відеобайти не повинні проксуватися через Go API або frontend SSR server.
14. Nginx напряму віддає HLS playlists і segments після перевірки доступу.
15. Не використовуй Rust у першій версії.

## 3. Frontend stack

Використовуй:

* React 19;
* TypeScript strict;
* TanStack Start;
* TanStack Router;
* TanStack Query;
* Rsbuild;
* Rspack;
* React Aria Components;
* Tailwind CSS 4;
* React Hook Form;
* Zod 4;
* Zustand;
* Orval;
* native fetch;
* hls.js;
* TanStack Table;
* TanStack Virtual;
* TipTap;
* ECharts;
* i18next та react-i18next;
* Rstest;
* Testing Library;
* MSW;
* Playwright;
* Oxlint;
* tsc --noEmit;
* Oxfmt;
* OpenTelemetry;
* Sentry adapter.

Не використовуй:

* Next.js;
* Vite;
* Vitest;
* Jest;
* ESLint;
* Redux Toolkit;
* Axios;
* Radix UI;
* shadcn/ui;
* бізнес-логіку в React components;
* прямі fetch виклики з presentation layer.

## 4. Frontend SSR

Використовуй TanStack Start із Rsbuild plugin.

SSR потрібен для:

```
/
/courses
/courses/$slug
/login
```

Selective або вимкнений SSR:

```
/learn                  data-only або client
/learn/$course/$lesson  client
/admin/*                 client
```

Frontend production build повинен створювати client і server bundles.

SSR server:

* відповідає лише за React rendering;
* отримує дані через Go API;
* передає session cookies у Go API;
* не містить domain-логіки;
* не працює як media proxy;
* має health endpoint;
* коректно завершує роботу;
* підтримує hydration TanStack Query.

Nginx маршрутизує:

```
/            -> TanStack Start SSR
/api/*       -> Go API
/media/*     -> HLS storage
```

## 5. Frontend architecture

Базова структура:

```
frontend/
├── src/
│   ├── app/
│   │   ├── router/
│   │   ├── providers/
│   │   ├── composition/
│   │   ├── config/
│   │   └── styles/
│   │
│   ├── routes/
│   │   ├── __root.tsx
│   │   ├── index.tsx
│   │   ├── courses/
│   │   ├── learn/
│   │   ├── account/
│   │   └── admin/
│   │
│   ├── domain/
│   │   ├── course/
│   │   ├── lesson/
│   │   ├── transcript/
│   │   ├── enrollment/
│   │   ├── order/
│   │   └── identity/
│   │
│   ├── application/
│   │   ├── catalog/
│   │   ├── learning/
│   │   ├── checkout/
│   │   └── administration/
│   │
│   ├── ports/
│   │   ├── CourseRepository.ts
│   │   ├── TranscriptRepository.ts
│   │   ├── PlaybackGateway.ts
│   │   ├── SessionGateway.ts
│   │   └── PaymentGateway.ts
│   │
│   ├── adapters/
│   │   ├── api/
│   │   ├── player/
│   │   ├── storage/
│   │   └── telemetry/
│   │
│   └── presentation/
│       ├── storefront/
│       ├── learning/
│       ├── admin/
│       └── ui/
│
├── e2e/
├── public/
├── rsbuild.config.ts
├── rstest.config.ts
├── oxlint.config.ts
├── playwright.config.ts
└── tsconfig.json
```

Допускається коригування структури, якщо воно обґрунтовано dependency direction і не створює зайвих абстракцій.

## 6. UI system

Побудуй власний UI kit над React Aria Components.

Початкові компоненти:

* Button;
* Link;
* TextField;
* TextArea;
* Select;
* ComboBox;
* Checkbox;
* Switch;
* Dialog;
* Menu;
* Tabs;
* Tooltip;
* ProgressBar;
* Table;
* Pagination;
* FileUpload;
* Toast;
* EmptyState;
* ErrorState;
* Skeleton.

Вимоги:

* keyboard navigation;
* focus management;
* visible focus;
* semantic HTML;
* ARIA attributes через React Aria;
* responsive layout;
* light і dark themes через design tokens;
* відсутність hardcoded кольорів у feature components;
* підтримка 200% text zoom;
* мінімальний target size для інтерактивних елементів;
* accessibility tests для критичних компонентів.

## 7. Frontend state rules

TanStack Query використовується для:

* каталогу;
* деталей курсу;
* користувача;
* enrollment;
* замовлень;
* progress;
* transcript;
* адміністративних mutations.

TanStack Router search params використовуються для:

* пошуку;
* фільтрів;
* сортування;
* пагінації;
* активних admin tabs;
* timestamp ?t=754.

Zustand використовується тільки для:

* video player instance;
* високочастотного playback state;
* тимчасового upload state;
* локальних panels;
* локального кошика до синхронізації.

Не дублюй server state у Zustand.

## 8. Backend stack

Використовуй:

* Go 1.25 або останню сумісну stable версію;
* стандартний net/http або Chi;
* PostgreSQL;
* pgx;
* sqlc;
* Goose;
* Redis;
* Asynq;
* OpenAPI;
* structured logging;
* OpenTelemetry;
* Prometheus metrics;
* Argon2id;
* Testcontainers для integration tests;
* FFmpeg та ffprobe.

Не використовуй важкий framework, ORM із runtime magic або глобальний service locator.

## 9. Backend structure

```
backend/
├── cmd/
│   ├── api/
│   │   └── main.go
│   └── media-worker/
│       └── main.go
│
├── internal/
│   ├── identity/
│   ├── course/
│   ├── enrollment/
│   ├── order/
│   ├── payment/
│   ├── media/
│   └── transcript/
│
├── platform/
│   ├── postgres/
│   ├── redis/
│   ├── httpserver/
│   ├── observability/
│   ├── filesystem/
│   └── ffmpeg/
│
├── migrations/
├── openapi/
├── sql/
├── sqlc.yaml
├── go.mod
└── go.sum
```

Кожен бізнес-модуль організовуй так:

```
internal/course/
├── domain/
├── application/
├── ports/
└── adapters/
    ├── http/
    └── postgres/
```

Не створюй порожні interfaces «на майбутнє». Port додається тільки тоді, коли application use case реально має зовнішню залежність.

## 10. Початкові backend modules

Створи основу для:

### Identity

* User;
* Role;
* Session;
* authentication middleware;
* authorization policy.

Ролі:

```
student
instructor
admin
```

### Course

* Course;
* Module;
* Lesson;
* CourseStatus.

Стани курсу:

```
draft
published
archived
```

### Enrollment

* Enrollment;
* перевірка доступу студента до курсу;
* порт для playback authorization.

### Media

* VideoAsset;
* Upload;
* TranscodingJob;
* MediaStorage port;
* Transcoder port.

Стани відео:

```
created
uploading
queued
processing
ready
failed
```

### Transcript

* Transcript;
* TranscriptSegment;
* DocumentLink;
* TranscriptionJob.

### Order

На першому етапі створити тільки domain skeleton і PaymentGateway port. Не підключати реальний платіжний provider у foundation phase.

## 11. Media architecture

Використовуй:

* local filesystem як перший storage adapter;
* FFmpeg для transcoding;
* ffprobe для metadata;
* Redis + Asynq для jobs;
* Nginx для HLS data plane;
* hls.js у frontend.

Структура файлів:

```
media/
├── source/
│   └── {videoId}/original
├── hls/
│   └── {videoId}/
│       ├── master.m3u8
│       ├── 360p/
│       ├── 720p/
│       └── 1080p/
├── thumbnails/
└── temporary/
```

Профілі першої версії:

* 360p;
* 720p;
* 1080p тільки якщо source підтримує цю роздільну здатність;
* H.264;
* AAC;
* HLS;
* сегменти .m4s або .ts, вибір задокументувати ADR.

Media worker повинен:

1. отримати job;
2. перевірити source;
3. отримати metadata через ffprobe;
4. створити тимчасову робочу директорію;
5. запустити FFmpeg;
6. створити HLS variants;
7. створити master playlist;
8. створити thumbnail;
9. атомарно перенести результат у final storage;
10. оновити статус;
11. очистити temporary files;
12. коректно обробити retry та cancellation.

Не передавай довільні аргументи користувача в FFmpeg command.

## 12. Playback security

Flow:

1. Student відкриває lesson.
2. Frontend викликає playback endpoint.
3. Go перевіряє session.
4. Go перевіряє enrollment.
5. Go створює короткоживучий playback token.
6. Frontend отримує HLS URL.
7. Nginx перевіряє token через secure_link або auth_request.
8. Nginx напряму віддає playlist і segments.

Вимоги:

* TTL токена 2–5 хвилин;
* відсутність прямого public доступу до storage;
* rate limiting;
* обмеження паралельних playback sessions;
* підтримка HTTP range requests там, де потрібно;
* cache headers;
* CORS тільки для дозволених origin;
* логування playback access без збереження секретів.

Це не DRM. Не заявляй захист, якого система фактично не забезпечує.

## 13. Transcript

Модель сегмента:

```ts
type TranscriptSegment = {
  id: string;
  lessonId: string;
  startMs: number;
  endMs: number;
  text: string;
  speaker?: string;
  confidence?: number;
  documentLinks: DocumentLink[];
};
```

Transcript UI повинен підтримувати:

* timestamp;
* click-to-seek;
* підсвічування активного сегмента;
* autoscroll;
* тимчасове вимкнення autoscroll після ручної прокрутки;
* пошук;
* deep link через ?t=;
* відкриття документації;
* keyboard navigation.

На першому етапі реалізуй ручні transcript fixtures і API contract. Автоматичне розпізнавання мовлення не є частиною foundation slice.

Підготуй port для майбутнього transcription adapter.

Майбутні реалізації:

* faster-whisper worker;
* або whisper.cpp subprocess.

Не додавай Python runtime у foundation phase.

## 14. OpenAPI

Go API є source of truth для HTTP contract.

Потрібно:

* versioned /api/v1;
* OpenAPI specification;
* generated TypeScript client через Orval;
* стандартизований error response;
* request ID;
* pagination contract;
* idempotency key для майбутніх payments;
* generated code зберігати окремо;
* generated DTO мапити в frontend domain models через adapters.

Базова помилка:

```json
{
  "error": {
    "code": "course_not_found",
    "message": "Course was not found",
    "requestId": "..."
  }
}
```

## 15. Перший вертикальний slice

Перший slice повинен проходити через усі необхідні layers.

Реалізуй:

1. PostgreSQL migration для courses.
2. Course domain entity.
3. CourseRepository port.
4. PostgreSQL adapter.
5. ListPublishedCourses use case.
6. GetPublishedCourse use case.
7. HTTP handlers.
8. OpenAPI contract.
9. Generated frontend client.
10. Frontend API adapter.
11. SSR каталогу курсів.
12. SSR сторінки одного курсу.
13. Loading, empty і error states.
14. Unit tests domain/application.
15. Integration test repository.
16. Frontend component tests.
17. Playwright test каталогу.
18. Docker Compose запуск усього slice.

Не починай із payments, автоматичної транскрипції чи повного admin CRUD.

## 16. Infrastructure

Docker Compose повинен містити:

* frontend SSR;
* Go API;
* media worker;
* PostgreSQL;
* Redis;
* Nginx;
* optional monitoring profile.

Додай:

* health checks;
* named volumes;
* non-root containers, де можливо;
* read-only mounts для HLS output у Nginx;
* окремий writable mount для media worker;
* resource limits для FFmpeg worker;
* log rotation;
* graceful shutdown;
* dependency health conditions;
* .env.example;
* development overrides.

Не додавай secrets у repository.

## 17. Observability

Для першої версії:

* structured JSON logs;
* request ID;
* trace ID;
* HTTP duration;
* HTTP status counters;
* database query duration;
* queue depth;
* transcoding job duration;
* transcoding failures;
* active playback sessions;
* health/readiness endpoints.

Не логуй:

* паролі;
* session cookies;
* playback tokens;
* payment secrets;
* повні приватні transcript documents.

## 18. Testing strategy

### Frontend

Rstest:

* domain tests;
* application tests;
* React component tests;
* hook tests;
* adapters з MSW;
* player store tests.

Playwright:

* каталог завантажується через SSR;
* відкриття сторінки курсу;
* keyboard navigation;
* admin route guard;
* transcript click-to-seek у майбутньому slice.

### Backend

* table-driven unit tests;
* use-case tests із fake ports;
* PostgreSQL integration tests;
* HTTP contract tests;
* migration tests;
* media command generation tests;
* path traversal security tests.

### CI порядок

```
format-check
→ lint
→ typecheck
→ unit tests
→ integration tests
→ build
→ E2E
```

Oxlint відповідає за lint. tsc --noEmit залишається окремою обов'язковою перевіркою.

## 19. Documentation

Створи:

```
docs/
├── architecture/
│   ├── context.md
│   ├── containers.md
│   ├── frontend.md
│   ├── backend.md
│   └── media.md
├── adr/
│   ├── 001-modular-monolith.md
│   ├── 002-tanstack-start-rsbuild.md
│   ├── 003-react-aria.md
│   ├── 004-local-media-storage.md
│   ├── 005-hls-nginx.md
│   └── 006-transcript-model.md
└── development/
    ├── local-setup.md
    ├── testing.md
    └── troubleshooting.md
```

README повинен містити:

* призначення системи;
* стек;
* prerequisites;
* запуск однією командою;
* migrations;
* tests;
* build;
* основні URLs;
* структуру repository;
* відомі обмеження foundation phase.

## 20. Правила виконання агентом

Перед змінами:

1. Перевір поточну директорію.
2. Перевір наявність існуючих файлів.
3. Не видаляй і не перезаписуй сторонні дані.
4. Прочитай AGENTS.md, якщо він існує.
5. Зафіксуй припущення.
6. Створи робочий план.
7. Розбий роботу на задачі по 0.5–2 години.
8. Познач залежності.
9. Виконуй тільки один логічний slice за раз.

Під час реалізації:

* не залишай неперевірений generated boilerplate;
* не створюй абстракції без використання;
* не приховуй errors;
* не використовуй any;
* не застосовуй type assertions для маскування проблем;
* не допускай циклічних залежностей;
* не змішуй domain models і DTO;
* не дублюй server state;
* не створюй глобальні mutable singleton-и;
* не додавай бібліотеки без конкретної потреби;
* не оновлюй залежності поза scope;
* зберігай маленькі логічні commits;
* після кожного етапу запускай мінімально необхідні перевірки.

Якщо рішення неоднозначне:

1. Визнач варіанти.
2. Назви конкретні trade-offs.
3. Обери найпростіший варіант, що покриває поточний scope.
4. Зафіксуй рішення в ADR.
5. Не блокуй роботу питанням, якщо можна зробити безпечне оборотне припущення.

## 21. Порядок реалізації

Виконуй у такому порядку:

### Phase 1 — foundation

* repository;
* frontend scaffold;
* backend scaffold;
* Docker Compose;
* PostgreSQL;
* Redis;
* Nginx;
* CI;
* architecture docs.

### Phase 2 — first vertical slice

* Course domain;
* migrations;
* repository;
* use cases;
* API;
* OpenAPI;
* generated frontend client;
* SSR catalog;
* SSR course page;
* tests.

### Phase 3 — identity

* users;
* sessions;
* roles;
* route guards;
* API authorization.

### Phase 4 — admin course management

* course CRUD;
* modules;
* lessons;
* publish/archive workflow.

### Phase 5 — media upload

* upload;
* storage;
* jobs;
* FFmpeg;
* HLS generation;
* admin processing status.

### Phase 6 — playback

* signed access;
* Nginx validation;
* hls.js;
* progress;
* session limits.

### Phase 7 — transcript

* transcript API;
* editor;
* synchronized transcript;
* document links;
* WebVTT;
* deep links.

### Phase 8 — orders

* order;
* checkout;
* payment adapter;
* webhook;
* enrollment.

Не переходь до наступної phase, поки попередня не має working result і перевірок.

## 22. Definition of Done для foundation

Foundation вважається готовим, якщо:

* monorepo має погоджену структуру;
* frontend запускається на TanStack Start + Rsbuild;
* SSR працює для каталогу та сторінки курсу;
* React Aria UI kit має базові компоненти;
* Go API запускається;
* PostgreSQL migrations працюють;
* Redis доступний;
* Course vertical slice проходить через domain/application/ports/adapters;
* OpenAPI генерує TypeScript client;
* Nginx маршрутизує frontend, API і media;
* media worker запускається та має health/status механізм;
* Docker Compose піднімає систему;
* unit, integration та E2E tests проходять;
* Oxlint проходить без errors;
* TypeScript typecheck проходить;
* Go lint/tests проходять;
* production builds створюються;
* документація відповідає фактичній реалізації;
* немає secrets у repository;
* немає незадокументованих критичних TODO;
* README дозволяє іншому розробнику запустити систему без усних пояснень.

## 23. Формат звіту після кожного етапу

Після завершення етапу повідом:

1. Що реалізовано.
2. Які файли та модулі створено.
3. Які архітектурні рішення прийнято.
4. Які команди перевірки виконано.
5. Які перевірки пройшли.
6. Які обмеження залишилися.
7. Яка наступна найменша задача.
8. Чи потрібне рішення користувача.

Почни з Phase 1. Спочатку покажи детальний план маленьких задач і залежностей. Після цього реалізуй foundation послідовно, не намагаючись одразу побудувати весь продукт.

## 24. Уточнення scope після вихідного ТЗ

Рішення замовника, ухвалені під час планування Phase 3–4; модифікують §21. Дати — дати
рішень.

1. **Порядок фаз підтверджено (2026-09-05).** Замовлення й оплата лишаються Phase 8:
   продати можна лише те, що вміємо наповнити (Phase 4) й доставити (Phase 5–6).
2. **Власник курсу (2026-09-04).** `instructor_id` — обов'язкове поле першої міграції
   Phase 4, хоча споживач з'являється значно пізніше. Це найдешевший момент: порожня
   таблиця courses проти пізнішої міграції курсів, модулів, уроків і всіх запитів.
3. **Робоче місце інструктора (2026-09-04).** Окрема фаза після Phase 6, не розділ
   адмінки. Спирається на чотири опори у своїх фазах: власник курсу (4), відео (5),
   прогрес (6), гілки бесід (7).
4. **Гілки бесід — коментарі студентів до відео (2026-09-05).** Розширення Phase 7
   (синхронізований навчальний досвід, поряд із транскриптом), не частина робочого місця
   інструктора: коментує студент, інструктор лише модерує через власне робоче місце.
5. **Handler-level авторизація за роллю.** Перенесена з Phase 3 у Phase 4 і в'їжджає
   разом із першим захищеним admin-ендпоїнтом (рішення: `docs/identity/plan.md`
   §Decisions).
6. **React Hook Form.** Девіація I05: логін із двох полів зроблено без RHF. RHF
   підключається з першими адмін-формами Phase 4 — залежність лише зі споживачем.
