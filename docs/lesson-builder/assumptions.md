# lesson-builder — жива нота (assumptions / розвилки)

## Handoff (2026-10-13, оновлено після LB10b — LB10 ЗАКРИТО)

- **Гілка** `feat/lesson-builder` off main@9e7759e. **LB00–LB09a ЗАКОМІЧЕНО**; HEAD `e0a4551` (LB09a) → +
  LB09b-1 комітиться цим кроком. НЕ запушено, PR ще нема. **Standing-авторизація власника (2026-10-12):
  авто-коміт кожного зеленого слайсу до кінця фази.** Реалізація — **через Codex** (`codex exec
  "$(cat prompt)" < /dev/null`; CODEX_EXIT з лога; пісочниця Codex НЕ гонить Docker/listen —
  sqlc-generate/check-generated/integration/фронт-тести проганяю сам). Промпти — scratchpad
  `codex-lb0X-prompt.md`.
- **LB09b-1 ЗРОБЛЕНО (комітиться):** `internal/image/adapters/postgres/access.go` — `LoadImageAccess`
  (JOIN lesson_images→lessons→modules→courses LEFT JOIN course_grants; price/currency-mismatch guard;
  Published/Paid/Entitled/Membership/Preview) + `LoadLessonOwner` (instructor_id, not-found→("",false,nil))
  + `membership` helper + compile-time assertions AccessReader/AuthoringReader. Integration під Docker PASS
  1.96s/3.18s (owner, missing lesson, owner/non-member/active-member CanConsumeBy, missing image→
  ErrImageNotFound); gofmt/vet/build + boundary зелені. Підтверджено: published+unpaid+non-preview → не-член
  CanConsume=false.
- **LB09b-2 ЗРОБЛЕНО (комітиться):** `internal/image/adapters/http/handler.go` — upload
  `POST /api/v1/admin/lessons/{id}/images` (multipart, `MaxBytesReader = 5 MiB + 512 KiB envelope slack`,
  `FormFile("file")`, sniff→`DetectContentType`, `image.DecodeConfig` для розмірів без повного декоду,
  делегат у `Uploader.Upload`; помилки 404 lesson_not_found/403 forbidden/415 unsupported_image/413
  image_too_large/400 image_too_many_pixels/400 invalid_image/503) + authorize
  `GET /internal/image/authorize` (`X-Image-Id`, actor з контексту, 204=allowed, email_not_verified/
  forbidden/image_not_found→403/403/404, 503 default; `normalizeAuthenticationDenial` 401→403 скопійовано
  з playback) + service-інтерфейси `uploadService`/`authorizeService`. Contract-тести 24 кейси (httptest,
  реальний multipart 1×1 PNG, undecodable-PNG→invalid_image, MaxBytesReader→413, 401→403 wrapper). `x/image`
  v0.47.0 direct у go.mod (`go mod tidy` промоутнув + прибрав передіснуючий moby/client_model indirect-дрейф,
  підняв x/sync/sys/text). Гейти: gofmt/vet/build/http-тести + **повний backend unit 48pkg exit 0** + boundary
  зелені. **nginx-envelope tension:** LB10 upload-локація `client_max_body_size` МАЄ перевищувати 5 MiB на
  envelope, інакше справжнє 5-MiB зображення ріжеться на краю (прапор на LB10).
- **LB09b-3 ЗРОБЛЕНО (комітиться) — LB09 ЗАКРИТО.** Wiring `cmd/api/main.go`: аліаси imagefs/imagehttp/
  imagepostgres/imageapp; `const imageRoot = "/media/i"` (рішення у Журналі); один `imagepostgres.Repository`
  на всі 3 порти; `imagefs.New(imageRoot)` + Close-defer (дзеркало media-worker); `NewUploader(repo,storage,
  repo,newUUID,time.Now)`, `NewAuthorizer(repo)`, `NewHandler`; RegisterAdmin+RegisterInternal у Register-
  блоці. OpenAPI: `POST /api/v1/admin/lessons/{id}/images` — ПЕРШИЙ `multipart/form-data` у спеці (поле
  `file` binary, 201→`ImageUploadResponse{id,contentType,width,height}`, 400/401/403/404/413/415/503 на
  наявні компоненти); `/internal/image/authorize` НЕ в спеці (nginx-internal). Orval regen → новий
  `uploadAdminLessonImage` + моделі. Гейти: `make check-generated` exit 0, `go build ./cmd/api` ok, frontend
  `tsc --noEmit` 0 помилок, contract 24 + image-postgres integration зелені.
- **LB10a ЗРОБЛЕНО (комітиться):** формат сховища з розширенням (рішення власника про content-type під
  nosniff — Журнал). `ports.Storage.Put(ctx, id, extension, r)`; filesystem пише `<id>.<ext>` з
  `validExtension` {jpg,png,webp} у traversal-guard; `domain.ExtensionFor(contentType)` (jpeg→jpg/png→png/
  webp→webp); `application.Upload` мапить `image.ContentType`→ext і передає в Put. Тести: storage_test
  (ext-арг + `TestPutInvalidExtension`: "","gif","jpeg","exe","png/../escape","PNG"), upload_test
  (fakeStorage ловить extension, асерт `ExtensionFor(contentType)`). Гейти: gofmt/vet/unit
  (filesystem+http+app+domain)+boundary+full-build зелені.
- **LB10b ЗРОБЛЕНО (комітиться) — LB10 ЗАКРИТО.** nginx: `map $uri $image_id` (uuid до крапки з
  `/media/i/<uuid>.(jpg|png|webp)`), `location ^~ /media/i/` → `auth_request /_image_auth` → rewrite strip
  → `root /media` (mime за ext → Content-Type), security-headers + Cache-Control private; `/_image_auth` →
  `proxy_pass $api/internal/image/authorize` X-Image-Id $img; upload-size = **вкладена** regex-локація в
  `^~ /api/` (`~ ^/api/v1/admin/lessons/<uuid>/images$`, `client_max_body_size 6m`) — лишає top-level
  precedence недоторканим (R30-валідатор проходить). compose: api `volumes:[media-data:/media]` +
  `depends_on upload:service_started`; `os.MkdirAll(imageRoot,0750)`+Chmod у main.go (bootstrap дерева).
  **ФІКС, знайдений доказом:** authorize `image_not_found` 404→**403** (auth_request розуміє лише
  2xx/401/403; 404→500→@unavailable=503; 403 ще й не тече існування). Dockerfile НЕ чіпав (api-стадія без
  /media; runtime MkdirAll). Гейти: nginx -t ok, R30 exit 0, build/vet/unit(filesystem+http+app+domain)+
  boundary зелені; **`infrastructure/tests/image-authorize.sh` PASS у Compose** (base+e2e, api/migrate/nginx
  пересоздані, DB догнав 00018+00019): owner=200+image/png, anonymous=403, outsider(не-член)=403,
  malformed=403, missing=403, no-ext=403.
- **ДАЛІ: LB11** фронтенд reader `LessonDocumentView` + query (віддача зображень новим шляхом
  `/media/i/<id>.<ext>`; документ зберігає id+contentType або повний відносний URL — уточнити в LB11). Тоді
  LB12 editor (Tiptap lazy, move БЕЗ isDisabled — [[isdisabled-on-focused-button-drops-focus]]), LB13 bundle
  grep (нема base64), LB14 DOM flatness, LB15 ordered `make check` + повний e2e + звіт. **Нотатка для LB15:**
  `image-authorize.sh` варто зачепити у фазовий гейт (як playback-authorize.sh).
- **(нижче — старий детальний план LB10b, залишено для історії):** nginx: `map $uri $image_id` вирізає uuid ДО крапки
  з `/media/i/<uuid>.(jpe?g|png|webp)`; `location ^~ /media/i/` → `auth_request /_image_auth` (помилки →
  наявні @denied/@unavailable) → rewrite strip prefix → `root /media` (types{} дають Content-Type за ext) +
  security-headers + `Cache-Control private,no-store`; `location = /_image_auth` internal →
  `proxy_pass $api/internal/image/authorize`, `proxy_set_header X-Image-Id $image_id`; upload-size: regex
  `~ ^/api/v1/admin/lessons/<uuid>/images$` з `client_max_body_size 6m`+`proxy_request_buffering off`
  (УВАГА precedence: `^~ /api/` короткозамикає regex → конвертувати `^~ /api/`→`/api/` plain-prefix, regex
  перевіряється першим; наразі інших regex-локацій нема, тож безпечно). compose: api `volumes:
  [media-data:/media]` (read_only root + volume rw ок), `depends_on upload: service_started` (K13 populate-
  race), api створює `/media/i` на старті — `os.MkdirAll(imageRoot, 0750)` у `cmd/api/main.go` ПЕРЕД
  `imagefs.New` (композиція бутстрапить середовище; LB08-Storage навмисно без MkdirAll). nginx вже
  `group_add:['10001']`+ro — читає 0640/0750 gid-10001. integration: Compose up → upload через API → fetch
  через nginx з членством (200 + Content-Type) і без (403). Тоді LB11–13 фронт, LB14 DOM, LB15 gate+e2e+звіт.
- **ГОТЧА підтверджена знову:** фоновий `make … | tail; echo EXIT=$pipestatus` маскує код — перший запуск
  integration показав «exit 0» (обгортка), а насправді make впав «No rule to make target» бо cd не спрацював
  для фонового шелла → гони `make -C <root>` і читай EXIT із лога. [[background-exit-code-is-last-pipe]].
- **Пройдено:** GATE 1 + GATE 2 (вердикт свіжим Opus-4.8; блокери закриті). Рішення власника:
  секції video/text/image/row/task (gallery→фаза `portfolio`); Description→односекційний text тоді
  DROP; i18n схема locale-aware UI одна локаль; draft+published дві ревізії; task=rich-text; окремий
  editor-route явне збереження; гейт тіла+зображень = **CanConsume+EmailVerified** (дзеркало
  транскрипта, без lease).
- **Слайси (durable статуси — `work/tasks.md`):**
  - **LB00** спайк GO · **LB01** домен · **LB02** міграція 00018+backfill (integration ok) ·
    **LB03** ports+app (gate) · **LB04** postgres-адаптер (integration ok 3.2s) — закомічені зелені.
  - **LB05** HTTP документа — `course/adapters/http/document_handler.go`: admin GET draft / PUT / POST
    publish + learn GET published (гейт делеговано сервісу EmailVerified→CanConsume, ОКРЕМО від outline);
    OpenAPI 4 шляхи + рекурсивні схеми LessonDocument/Section/RichText/Block/Span; Orval regen; contract-
    тести на фейках. ЗАКОМІЧЕНО `5cdb1bf`.
  - **LB06** DROP `course_lessons.description` — міграція 00019 + lesson-only прибирання (домен, sqlc-джерело
    + inline-upsert, admin DTO/handler, OpenAPI, фронт форма/zod/locales, + фікстури ВСІХ модулів і УСІ e2e-
    спеки, бо колонку сіяли скрізь). Сесія: generate/check-generated/build/vet/**integration INTEG=0**
    (`course/adapters/postgres` ok 23.3s) + фронт-тести зелені. Guardrail: курс-description/summary цілі.
    **ЗРОБЛЕНО** (staged, не закомічено).
  - **LB07** `[P]` модуль `internal/image` домен+порти — **ЗРОБЛЕНО** Codex 2026-10-12 (ще НЕ закомічено
    на момент запису): `domain/image.go` (`Image`, `DetectContentType` magic jpeg/png/webp sniff 12,
    `AcceptDimensions` int64×int64 ≤40 MP, `NewImage` allowlist+5 MiB+dims у фіксованому порядку),
    `domain/errors.go` (5 sentinel), `ports/{repository,storage}.go`. Ліміти (byte 5 MiB = nginx 5m,
    pixel 40 MP) — рішення в Журналі нижче. Сесія незалежно: gofmt/vet/test зелені, boundary-тест ok.
    **Access-reader порт (cross-module lesson→course CanConsume+EmailVerified) свідомо НЕ в LB07** —
    без consumer був би unused interface (конституція); заводиться в LB09 разом з application.Authorize.
  - **LB08** `internal/image` adapters — **ЗРОБЛЕНО** Codex 2026-10-12 (ще НЕ закомічено на момент запису):
    `adapters/filesystem/storage.go` (os.Root на images-дир, `Put` O_WRONLY|O_CREATE|O_EXCL + fsync +
    partial-cleanup, validUUID проти traversal, fileMode 0640, flat `<id>`), `adapters/postgres/
    repository.go` (`Repository{pool}`, Insert/FindByID, `$::text::uuid`, pgx.ErrNoRows→ErrImageNotFound,
    created_by nullable→""). Сесія: gofmt/vet/unit зелені; **integration під Docker PASS 3.0s** (round-trip
    усіх полів+created_at.Equal, missing, dup 23505, FK 23503, nullable creator). directoryMode прибрано.
  - **LB09a** image domain-access + application — **ЗРОБЛЕНО** Codex 2026-10-13 (ще НЕ закомічено на момент
    запису): `domain/access.go` (`Actor{UserID,Administrator,EmailVerified}`+`CanEditLesson`;
    `ImageAccess`+`CanConsumeBy` через `contracts/access`), errors +forbidden/lesson_not_found/
    email_not_verified, `application/{upload,authorize}.go`, `ports/access.go` (AccessReader/
    AuthoringReader). Unit на фейках зелені (Put/Insert call-counts + порядок Put→Insert). gofmt/vet/
    test/boundary зелені. Orphan-bytes-on-insert-fail — named limitation (нема sweeper у v1).
  - **LB09b-1 ЗРОБЛЕНО** Codex+сесія 2026-10-13: postgres `LoadImageAccess`+`LoadLessonOwner`+integration
    (деталі у Handoff вище). **ДАЛІ: LB09b-2/b-3** `internal/image` HTTP+wiring: HTTP `POST /api/v1/admin/lessons/
    {id}/images` (multipart, MaxBytesReader 5m, sniff→DecodeConfig→Upload) + internal `GET /internal/image/
    authorize` (`X-Image-Id`, actor з cookie, normalizeAuthenticationDenial як playback); OpenAPI(upload)+
    Orval; composition `cmd/api/main.go` (NewUploader(repo,fsStorage,repo,newUUID,time.Now), NewAuthorizer,
    RegisterAdmin+RegisterInternal); `golang.org/x/image/webp` у go.mod; contract-тести AC9. Гейт: contract+
    integration+check-generated. Тоді LB10 nginx.
  - **(старий текст LB09, замінено розбивкою вище):** `POST /api/v1/admin/lessons/{id}/images` (multipart,
    admin|instructor, requireSameOrigin, MaxBytesReader 5m; sniff→DetectContentType, image.DecodeConfig→
    AcceptDimensions, NewImage, Storage.Put, Repo.Insert) + internal `GET /internal/image/authorize`
    (власний reader imageID→lessonID→course→**CanConsume+EmailVerified**, без lease). ТУТ заводиться
    access-reader порт (cross-module seam) + application.Authorize-use-case (його consumer). Composition-
    root wiring course+image ендпоінтів. Contract-тести (крива mime/розмір/bomb→явні коди AC9:
    unsupported_media_type/payload_too_large/invalid_request). Тоді LB10 nginx.
    (HTTP: admin multipart upload + internal `/internal/image/authorize` reader lesson→course→CanConsume+
    EmailVerified, без lease)→LB10 (nginx `^~ /media/i/` + нова upload-location 5m; compose; auth_request
    integration). Потім LB11–13 фронт (reader чистий / lazy Tiptap editor / bundle-grep), LB14 DOM-вимір,
    LB15 ordered gate+повний e2e+звіт.
- **Докази Codex** — `docs/lesson-builder/report.md` (LB05+LB06 секції + session-verification). Жива схема-
  контракт — LB01 `document.go` (json-теги schemaVersion/sections/content/imageId/columns/blocks/spans/text/
  marks/href). Мірник гейта тіла — `transcript read.go` (EmailVerified→CanConsume).
- **Відкритих питань власника немає.** **Standing-авторизація (власник 2026-10-12):** авто-коміт кожного
  зеленого слайсу до кінця фази lesson-builder без повторного питання. LB06 закомічено `6156afe`.

## Ключові рішення плану (reuse-first, 3 агенти 2026-10-12)

- **Бекенд розміщення:** розширити модуль `course` (володіє уроками) — document.go у domain/app/
  ports/adapters; новий окремий модуль `internal/image` (НЕ розширювати відео-media).
- **Персистенс:** таблиця `lesson_documents` PK `(lesson_id, locale)`, draft_body+published_body
  jsonb, updated_at optimistic-lock; raw pgx (asr-прецедент), не sqlc. Publish = `UPDATE SET
  published_body = draft_body` (атомарно). PK з locale → мультимовність без зміни схеми.
- **Гейт (AC7):** новий `GET /learn/lessons/{id}/document` gated `access.CanConsume`, ОКРЕМО від
  outline (`getLearningCourse` — Preview:true-видимий будь-кому; body туди НЕ класти). Повний
  факт-сет (owner/status/paid/entitled/membership/preview), не transcript-ів edit-гейт.
- **Reader без редактора (AC4):** route-компоненти auto-split; АЛЕ важкі vendor-деп з eager
  router-контексту падають у 1.64MB main-chunk (доказ: hls.js). → reader = чистий компонент
  секції→HTML, нуль редактор-деп; Tiptap лише в `React.lazy` editor-піддереві, НЕ в
  createAdminDependencies/router-контексті.
- **Зображення:** модуль `internal/image`; multipart POST на `api` (admin, requireSameOrigin,
  magic-bytes jpeg/png/webp, MaxBytesReader, піксельний бюджет); store `/media/images/<id>`;
  nginx `/media/i/` → `/internal/image/authorize` (власний, `CanConsume`, без lease/EmailVerified);
  окремий upload-location (більший body-limit, `/api/`=1m). CSP не чіпати (`img-src 'self'`).
- **Міграція:** 00018 (goose); backfill Description→односекційний text (locale uk), потім DROP
  COLUMN description (pre-launch cut-freely). Міграцію-down втрата published-документів — named.
- **Факт-корекція:** Tailwind НЕ в бан-листі AGENTS.md і вже залежність (4.3.3) — дозволено.

## GATE 2 (закрито 2026-10-12, свіжий Opus-4.8, вердикт виносив не автор плану)

- **Вердикт:** NOT READY → усі несучі anchor'и ПІДТВЕРДЖЕНО (learning-запит не селектить lesson
  description; `access.CanConsume` — правильний гейт; media відео-only; route auto-split але eager
  router-ctx vendor-деп у main-chunk — доказ hls.js; 00018 наступний; конституція ✓). 2 блокери +
  прогалина — усі дешеві правки плану, ВНЕСЕНО:
  - **B1 (міграція):** зафіксовано ТОЧНУ форму backfill-документа (JSON у plan §Міграція); backfill
    через доменний валідатор, не сирий SQL; сценарій №7 сідить НЕПОРОЖНІЙ description і доводить
    валідність+рендер (не «майже-порожній» прогін). DROP COLUMN лише після доведеного backfill.
  - **B2 (AC3 reorder):** move-кнопки секцій НЕ копіюють `isDisabled` (пастка
    [[isdisabled-on-focused-button-drops-focus]] — фокус злітає на body, jsdom ховає); межа = no-op
    фокусованої кнопки / явне перенесення фокуса; сценарій №6 перевіряє real-browser keyboard.
  - **Прогалини закрито:** повний `grep description` + sqlc/Orval regen у side-effects; nginx
    upload-location нова, `client_max_body_size 5m` (не переюз /uploads tus); reader на відсутній
    рядок/0 секцій → порожній стан, не 404.
- **Найризикованіше припущення:** Tiptap тримається поза learn/main-chunk — де-ризик спайком
  (вимір бандла, сценарій №10); hls.js — живий доказ ризику, mitigation адресна.

## Рішення власника про гейт (ЗАКРИТО 2026-10-12)

- **Тіло-документ + зображення = `CanConsume` + EmailVerified** (власник: «є спека, що лише
  верифіковані мають доступ до курсів»). Authorize дзеркалить гейт транскрипта, без playback-lease.
  Прев'ю видно лише верифікованому акаунту, не аноніму. Питання публічності прев'ю-зображень закрите.

## Рішення (датовано, хто вирішив, що закриває)

- **2026-10-12, власник:** фаза `white-label` згорнута (fork-per-brand знімає потребу);
  єдина реальна інженерія — **редактор уроку (WYSIWYG)**. Гілка `feat/white-label →
  feat/lesson-builder`; дім `docs/white-label → docs/lesson-builder` (white-label артефакти
  були untracked, прибрані без втрат; гілка не мала комітів понад main 9e7759e).
- **2026-10-12, Codex (адверсар-критика, прийнято):** відхилити рекомендацію Craft.js «як є».
  Перейти на **власний структурований документ уроку** (упорядковані типізовані секції +
  обмежені пресети). Повну критику зберегти в артефакти фази. Несучі пункти прийнято:
  (1) не зберігати сирий JSON стороннього рушія — власна валідована версійована схема;
  (2) `learn.$.tsx:33` — `ssr:false`, тож «reader-SSR» був хибний claim; рендерер = власний скоуп;
  (3) один наявний відео-слот, не довільні відео-блоки; (4) безпека = валідований документ, не
  «довірений автор+CSP»; outline(public) ≠ content(gated) — тіло не підшивати в outline-відповідь
  (`learning.go CanBeViewedBy` Preview:true); (5) Tiptap MIT — правильний rich-text шар; (6)
  images/публікація/i18n — реальні підсистеми, не деталі JSON.
- **2026-10-12, власник (приклади уроків — вирішальний вхід Codex #4):**
  - Езотерика: відео → текст(замовляння) → ряд[зображення|пояснення] → завдання; частина без відео
    (текст+зображення). → розкладається в упорядковані секції, depth-1 вкладення в `row`.
  - Б'юті: відео-уроки + портфоліо **галереєю** (масштабована). → `gallery` як окрема типізована
    секція, не емуляція `row`+`image`.
  - Вигляд обох клієнтів ≈ totaltypescript.com → візуальний north-star клієнт-поверхні; тема/дизайн
    downstream (per-brand у копії), НЕ ця фаза.
- **2026-10-12, я (рекомендація, прийнято):** модель = власний документ + секції; drag-рушій НЕ
  потрібен (жоден приклад його не виправдовує). Інвентар v1: `video`(0..1) `text` `image`
  `row`(1–3 кол, діти text/image, depth-1) `task` `gallery`.

## Факти (ґрунт, 2026-10-12)

- `backend/internal/course/domain/authoring.go:62` — `Lesson{ID,Title,Description string,Position,
  IsPreview}`. Тіло уроку = **один плоский рядок `Description`**; структури немає.
- `frontend/src/routes/learn.$.tsx:33` — `ssr:false` (learn-шлях клієнтський; reader-SSR = новий скоуп).
- `backend/internal/course/domain/learning.go CanBeViewedBy` — outline видимий (Preview:true), контент
  гейтиться по-уроково; тіло не можна підшивати в широко-видиму outline-відповідь (AC7).
- Відео/транскрипт/прогрес/дискусія вже прив'язані до уроку (playback/transcript/discussion).
- Зображення: media-pipeline — nginx `/media/` auth_request, upload-модуль; загальної img-бібліотеки
  нема (потрібен шлях авторизації/лімітів для не-playback зображень — уточнити в plan).
- CSP (server.mjs): `img-src 'self' data:`, `style-src 'self' 'unsafe-inline'`, `script-src 'self'
  'nonce-…'`; e2e CSP report-only — зелений звичайний ран не доказ під enforcing.
- **(2026-10-12) Довільних зображень ніде зберігати/віддавати.** Media-модуль суто відео: tus по
  `videoId`, HLS, thumbnail'и генеруються з відео (`asset.go` StorageKind: source/hls/thumbnails/
  temporary; tus `server.go` keyed videoId). nginx (`nginx.conf`): `/media/p/<hex>/hls/` auth-gated
  відео (:67), решта `/media/` відмовляється (:102). → `image`/`row`-image/`gallery` ВСІ вимагають
  НОВОЇ підсистеми зображень (upload+store+authorize+serve), неминучої навіть для одного зображення.
- **(2026-10-12) Q5: сотні + before/after пари** (власник). Галерея = підсистема, не секція: альбоми/
  колекції на сотні, тип-пара «до/після» + UI-порівняння, пагінація/lazy. Поверх підсистеми зображень.

## Припущення (підтвердити на GATE 1)

- Pre-launch (власник 08.10): плоский `Description` можна мігрувати в односекційний `text`-документ
  сміливо, без shim для «старих даних» — але спершу підтвердити, чи є наповнені уроки (Q2).
- Чернетка/публікація потрібні (AC8) — форма (ревізії vs прапорець) відкрита (Q6).
- `task` v1 = стилізований rich-text, НЕ структуроване завдання зі здачею (Q7) — інакше окрема фаза.

## Поріг

Строго за `.claude/standards.md`: тригер «cross-cutting, more than one layer of BOTH backend and
frontend» СПРАЦЮВАВ (бек: таблиця документа+міграція+API+валідація; фронт: admin-білдер+learn-рендерер).
Повний spec-flow виправданий.

## Рішення GATE 1 (закрито 2026-10-12, підтвердив власник — вердикт гейту)

- **Розбивка на дві фази (Q10):** `lesson-builder` (ця) = схема документа + секції video/text/image/
  row/task + **підсистема довільних зображень** + draft/publish + learn-рендерер + безпека +
  міграція. `portfolio` (наступна) = gallery на сотні + альбоми + before/after. Причина: кожна фаза
  поставна; підсистема зображень усе одно потрібна тут і її переюзає portfolio; Codex-ризик «другого
  продукту». Езотерика після цієї фази працює повністю; б'юті-відео+одиночні зображення теж.
- **Q1–Q9** (деталі) — у spec.md «Рішення GATE 1». Ключове: gallery OUT; Description→односекційний
  text, старе поле геть; i18n — схема locale-aware, UI одна локаль; rich-text мінімальний allowlist
  без код-блоків; draft+published дві ревізії, атомарна публікація, без історії; task=rich-text;
  окремий маршрут редактора, явне збереження; спайк доводить 5 найгостріших ризиків.
- **Q11 (before/after, для фази `portfolio`, зафіксовано поки свіже):** елемент галереї = одиночне
  зображення АБО пара «до/після»; порівняння side-by-side (доступний дефолт), драг-повзунок —
  підсилення пізніше; галерея посилається на альбом, не інлайнить сотні референсів.

## Журнал

### LB11 — як документ посилається на зображення = `contentType` у image-секції (2026-10-13, ВЛАСНИК — рішення, закриває альтернативу)

- **Зберігати `contentType` (image/jpeg|png|webp) у image-секції документа; reader виводить
  розширення клієнтськи й будує `/media/i/<imageId>.<ext>`.** Механізм/чому: nginx (LB10b) матчить
  `^~ /media/i/` лише з розширенням (`map $uri $image_id` вирізає uuid ДО крапки), тож reader мусить
  знати ext. Збережена секція несла тільки `imageId`+`alt`; content_type живе в таблиці `lesson_images`,
  якою володіє модуль `image`, а `course` туди лізти не може (межі модулів). Три альтернативи:
  (A) `contentType` у секції — обрано; (B) повний URL у секції — відкинуто (бейкає serve-шлях у
  persisted jsonb, зміна префіксу вимагає міграції документів); (C) backend резолвить при read —
  відкинуто (course-модуль читав би чужу таблицю lesson_images, порушення меж; або важкий контракт-шов).
  Денормалізація мінімальна: і `imageId`, і `contentType` приходять з однієї upload-відповіді
  (`ImageUploadResponse{id,contentType}`), editor (LB12) кладе їх разом, тож розбіжності неможливі; якщо
  ext не збігся б з файлом на диску — URL дав би 404, видимо. **Даних для міграції нема: жодної
  image-секції ще не збережено** (backfill 00018 робить лише text-секції). Перевірка: reader будує
  `/media/i/<id>.png` для `contentType:"image/png"`, `<img>` рендерить під nosniff; domain-валідація
  відмовляє image-секцію з contentType поза allowlist jpeg/png/webp. Рішення власника через
  AskUserQuestion 2026-10-13.
- **Розбивка:** LB11a (бекенд: поле `ContentType` у `domain.Section` + валідація image-секції
  [allowlist, обов'язкове для image, заборонене для не-image], OpenAPI `LessonSection.contentType`
  enum, Orval regen; unit domain + check-generated), LB11b (фронт: `LessonDocumentView` чистий
  компонент, порт+адаптер+zod дзеркало транскрипта, query `learn.lessonDocument` неса-suspending,
  слот у `OpenLesson`, порожній стан, unit + e2e render).

### LB10 — content-type зображення під nosniff = розширення у файлі (2026-10-13, ВЛАСНИК — рішення, закриває альтернативу)

- **Зберігати `<id>.<ext>` (jpg/png/webp), serve-URL `/media/i/<id>.<ext>`; nginx mime-types дають
  Content-Type за розширенням.** Механізм/чому: глобальний `X-Content-Type-Options: nosniff`
  (nginx.conf:49) + пласке `<id>` без розширення → nginx віддає `application/octet-stream`, а під nosniff
  браузер ВІДМОВЛЯЄТЬСЯ рендерити це в `<img>` (серв зламаний). Stock-образ `nginxinc/nginx-unprivileged`
  не має `headers-more`, `default_type` не бере змінну з auth-сабреквесту → content-type із заголовка
  авторизації підставити НЕ можна. Розширення у файлі — стандартний stock-nginx патерн (types{} за ext).
  Відкинуто: X-Accel-Redirect (інший патерн серва ніж playback auth_request, API у шляху серва) і
  flat+custom-nginx (нестандартний образ, крихко). Рішення власника через AskUserQuestion 2026-10-13.
  Наслідок (pre-launch, ріжу сміливо — [[pre-launch-cut-freely]]): ревізія закомічених LB08 (Storage.Put
  +extension) і LB09 (Upload мапить content-type→ext). ext: image/jpeg→jpg, png→png, webp→webp. Перевірка:
  upload пише `/media/i/<id>.jpg`, nginx віддає з `Content-Type: image/jpeg`, `<img>` рендерить під nosniff.
- **Розбивка:** LB10a (бекенд: формат сховища з розширенням + domain ExtensionFor + Upload + тести,
  unit+filesystem), LB10b (nginx serve `^~ /media/i/`+auth_request+upload-size-локація, compose api-монт
  media-data rw + /media/i, Dockerfile за потреби, integration auth_request у Compose).

### LB09b-3 — корінь сховища зображень (2026-10-13, я — рішення, закриває альтернативу)

- **image bytes на диску = `/media/i/<uuid>` (НЕ окремий том, НЕ config/env-поле).** Механізм/перевірка:
  media-root `/media` захардкоджений у воркерах (`cmd/media-worker/main.go:26` тощо), API медіа-байтів НЕ
  пише; `platform/config` будує лише DSN/mail — поля storage-root НЕ існує. Зображення вантажаться
  **синхронно в API-процесі**, тож API потрібен власний корінь. Обрано `const imageRoot = "/media/i"` в
  `cmd/api/main.go` (під той самий `/media`-дерево-том, 0750/0640, group-readable nginx), бо LB10 додає
  serve-локацію `^~ /media/i/`→`/internal/image/authorize`: nginx віддає `/media/i/<id>`, тож API пише саме
  туди. Альтернатива (окремий env `IMAGE_ROOT`) відкинута — немає прецеденту env для roots, а хардкод-const
  дзеркалить наявний `mediaRoot`. Наслідок: дерево `/media/i` має існувати з правами (os.OpenRoot не робить
  MkdirAll після LB08) — створення у compose/deploy (LB10). Перевірка: upload пише `/media/i/<id>`, nginx
  під auth віддає його учневі (LB10 integration auth_request).

### LB09 — розвилки HTTP/application image (2026-10-12, я — рішення, закривають альтернативи)

- **webp-розміри через `golang.org/x/image/webp` (НЕ ручний парс RIFF-заголовка).** Механізм:
  stdlib `image/jpeg`+`image/png` дають `DecodeConfig` (лише заголовок, без декоду пікселів → жодної
  bomb), але webp у stdlib нема. x/image — golang.org-модуль, не в бан-листі AGENTS.md, used (webp у
  погодженому allowlist). Альтернатива (ручний парс VP8/VP8L/VP8X) — більше коду й крихкіше. Дод. у
  go.mod у LB09b, side-effect import `_ "golang.org/x/image/webp"` поряд з `image/jpeg`,`image/png`.
- **Cross-module authorize-reader = власний SQL-JOIN у image-postgres (НЕ Go-імпорт course).** Прецедент
  `playback/adapters/postgres/access.go:39 LoadVideoCourse` — playback JOIN-ить course-таблиці своїм SQL
  і мапить у власний домен-тип. Image так само: `LoadImageAccess(userID,imageID)` JOIN lesson_images→
  course_lessons→course_modules→courses→course_grants → `domain.ImageAccess` (дзеркало course
  `LessonDocumentAccess`). ADR 013 ✓ (SQL-факт, не імпорт модуля).
- **Upload авторизує автора через окремий owner-read (НЕ по imageID — рядка ще нема).** `LoadLessonOwner
  (lessonID)→instructorID`; `actor.CanEditLesson(ownerID)` (admin|owner); тоді NewImage→Storage.Put→
  Repo.Insert. Гейт serve (authorize) = EmailVerified+CanConsume (дзеркало playback); гейт upload =
  CanEditLesson (авторський, як media mint).
- **LB09 розбито** (завеликий на 1 Codex-прогін, [[codex-split-impl-and-tests]]): **LB09a** домен
  (ImageAccess/Actor/+errors) + application (Upload+Authorize) + порти (AccessReader/AuthoringReader) +
  unit на фейках; **LB09b** postgres-методи (LoadImageAccess/LoadLessonOwner)+integration + HTTP
  (multipart upload + internal authorize `X-Image-Id` як playback `X-Playback-Id`) + OpenAPI(upload)+
  Orval + composition wiring (`newUUID`, `filesystem.New(imagesPath)`) + contract-тести (AC9 коди).

### LB07 — ліміти зображень (2026-10-12, я — рішення, закриває альтернативу)

План фіксує `internal/image` домен+валідацію, але НЕ числа. Зафіксовано:
- **Byte-ліміт `MaximumImageBytes = 5<<20` (5 MiB).** Магнітуда+механізм: дорівнює
  `client_max_body_size 5m`, який LB10 ставить на upload-location nginx — одне число на двох шарах,
  інакше або nginx ріже те, що домен пустив, або навпаки. Enforcement на дроті — MaxBytesReader
  (LB09); домен валідує `byteSize > max → ErrImageTooLarge` як другий бар'єр.
- **Pixel-бюджет `MaximumImagePixels = 40_000_000` (40 MP).** Механізм: `width*height` (int64, щоб
  не переповнити int32) перевіряється ДО повного декоду — `image.DecodeConfig` (LB09) дає розміри з
  заголовка дешево. 40 MP при 4 B/px ≈ 160 MB RAM на картинку — стеля проти decompression-bomb, але
  вище за будь-яку реальну камеру (12–50 MP), тож легітимні б'юті/портфоліо-фото проходять.
- Magic-bytes allowlist jpeg(FF D8 FF)/png(89 50 4E 47…)/webp(RIFF….WEBP); sniff 12 байтів (webp
  потребує 12). CSP `img-src 'self'` не чіпається (serve через nginx `/media/i/`, не base64/data:).

### LB03 — відкрите питання контракту (2026-10-12, Codex)

**STOP до уточнення власником.** Запитаний порт
`LoadLessonAccess(ctx, lessonID string) (domain.LessonDocumentAccess, error)` не приймає
ідентифікатор читача, хоча `Entitled` і `Membership` — факти конкретного користувача в курсі.
Самого `lessonID` недостатньо для JOIN `course_grants` за користувачем. Наявний transcript-read
передає `actor.UserID` явно в `GetLessonTranscript(ctx, actor.UserID, lessonID, language)`;
course learning-reader також приймає ідентифікатор користувача. Контракту передачі actor через
`context.Context` або прив'язування access-reader до користувача для LB03 не задано.

**Питання:** чи змінюємо порт на
`LoadLessonAccess(ctx context.Context, userID, lessonID string) (domain.LessonDocumentAccess, error)`
і передаємо `actor.UserID` з усіх чотирьох application-методів? Рекомендація — так, за наявним
патерном явного userID. Перевірка після погодження: fake reader фіксує отриманий userID;
два верифіковані студенти одного уроку з різними membership отримують різні результати гейту;
LB04 перевіряє той самий сценарій на реальному JOIN `course_grants`.

Код LB03 і тести не змінено; перевірки Go не запускалися, бо реалізацію зупинено на контракті.

**РІШЕННЯ (2026-10-12, я — погоджено):** ТАК, порт =
`LoadLessonAccess(ctx, userID, lessonID string) (domain.LessonDocumentAccess, error)`; усі чотири
application-методи передають `actor.UserID`. `Entitled`/`Membership` — пер-глядацькі факти (JOIN
`course_grants` за користувачем), дзеркалить transcript-read і course learning-reader. Перевірка:
fake фіксує userID; два верифіковані студенти з різними membership → різні гейти; LB04 — реальний JOIN.

- 2026-10-12: пивот white-label→lesson-builder; гілка/дім перейменовано; spec.md написано.
- 2026-10-12: GATE 1 Раунд 1 (Q1–Q8) — власник «погоджую все», Q5 уточнив: сотні + before/after.
- 2026-10-12: факт — підсистема зображень не існує (media відео-only); Раунд 2 (Q9–Q11) + розбивка
  фази. Власник підтвердив. GATE 1 ЗАКРИТО. spec.md оновлено (gallery OUT, AC9 переписано під нову
  підсистему, рішення зафіксовані). Далі — plan.md.
