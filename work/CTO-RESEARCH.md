# CTO research — брифінг репо `courses`

| | |
|---|---|
| Дата | 2026-10-31 (re-sync під teach-editor + SA13-close + LB09a + deploy-readiness; попередній re-sync — 2026-10-29 на `169c0fc`; повний rebuild — 2026-10-11) |
| Базовий коміт | `5fbaaff` (docs(teach-editor) record PR #3 review PASS; teach-editor злито у `main` ff ref-push) |
| Всього комітів у репо | **987** на `5fbaaff` (950 на `169c0fc`) |
| Гейт свіжості 2026-10-31 | **37 комітів** від `169c0fc` (база 2026-10-29). Зачепили продукт, межі, ризики — чотири блоки: **teach-editor** (owner-initiated **поза §24**; TE1–TE6, GitHub PR #93/#94/#95, злито 2026-10-31 → `5fbaaff`): нова `/teach`-поверхня редактора документа уроку для інструктора на власний курс — бекенд-ендпоінти `GET/POST /api/v1/teach/lessons/{lessonId}/document` + `/document/publish`, `POST /api/v1/teach/lessons/{id}/images`, owner-scoped `GET /api/v1/teach/courses/{id}` (реюз `getAdminCourse` через `courseHandler.RegisterTeach`; `image.RegisterTeach`; mount-alias у `cmd/api/main.go`, без нового auth — gate через authoring.Load + course-ownership); фронт — маршрут `teach.courses.$courseId_.lessons.$lessonId.tsx` (дзеркало admin), спільний `createDocumentAuthoringRepository`/`createDocumentQueries` параметризовано над `'admin'|'teach'` (байт-ідентичні admin-ключі), `TeachLessonList`. **Закрив lesson-builder-ліміт «нема /teach-поверхні».** **TE6 оголив пре-існуючий shared admin+teach publish-bug** (publish 200 порожнім тілом за контрактом `Empty:true` → клієнт парсив `{updatedAt}` на `undefined` → хибна «Не вдалося зберегти» після успішної публікації); виправлено frontend-only (publish→`void`, `onPublish` бере токен із re-read, зберігає overlay); 4 unit-моки кодували фейковий body-контракт — пастка. **SA13 security-батч повністю злито** (#83–#90 + SA13-A-R owner-accepted no-code `6efa393`): cursor-UUID 400-not-503, TOTP replay consumed-counter, CSP form-action, safeReturnTo (/teach,/settings,/orders), /metrics bearer-guard, sslmode prod-tail (no-code), panic value/stack log, HLS umask 0750/0640, totp_secret шифрування at-rest + keyed-HMAC recovery codes. **LB09a** (PR #91/#92): компенсаційне видалення щойно збережених image-байтів при падінні row-insert (+ detached-ctx N1) — закрив ліміт «байти-сироти лишаються»; named crash-window residual лишається. **deploy-readiness** (`5a671d5`): єдиний чеклист усіх deploy-gated хвостів. Перезібрано: шапку, §1 (teach-editor shipped, SA13 черга порожня), §6 (lesson-builder ліміти (1)/(3) закрито). Детальні доменні §2–§6 — кеш; головна дельта — teach-editor-поверхня. Історія re-sync'ів збережена. |
| Гейт свіжості 2026-10-29 | **171 комітів** від `63b26bf` (база 2026-10-13). Зачепили межі, інфру, залежності й ризики — чотири блоки: **security-audit SA01–SA13** (`docs/security/audit-2026-10-24.md`; SA01–SA12 + SA09-T1 merged #66–#80, S0 закрито; жива лишилась **SA13** LOW/hardening-батч + S01/S02 блоковані TLS); **frontend feature-modules** (A-batch PR1–PR11 #45–#55: `frontend/src/modules/{admin,catalog,discussion,document,enrollment,identity,media,observability,playback,purchase,teach,transcript}` + shared course-домен у contracts/); **TanStack Start upgrade** (TSSTART-01/02 #56/#57: Start 1.169.3 / Router 1.170.45 / Query 5.102.8, zod 4.6.7 jitless під enforcing-CSP); **observability** (B-1, PR #62 `0b69d04`: FE error-capture + SSR/API log-aggregation promtail→loki + Grafana `infrastructure/monitoring/`). **Головне архітектурне — generate-OpenAPI-from-Go (SA10/S0):** `openapi.yaml` тепер **генерується з Go transport-типів** і є authoritative (S0.2b cutover PR #81, follow-up PR #82); backend += 10-й модуль `clientlog`, новий build-time seam `internal/openapispec` (leaf-агрегатор) + `internal/platform/openapigen` (рефлектор kin-openapi) + `cmd/openapi-gen`. Перезібрано: шапку, §1 (10 модулів, +clientlog, архітектурний факт №1 = generated-authoritative contract), §3 (openapigen/openapispec seam + frontend feature-modules), §4 (observability + openapi-gen у make generate/check-generated), §5 (kin-openapi, Tiptap, TanStack bump), §6 (SA13 жива security-черга, B-1 знято). Детальні доменні розділи нижче — кеш, точний для доменних модулів; головна дельта — S0-seam. Історія re-sync'ів збережена. |
| Гейт свіжості 2026-10-13 | **43 коміти** від `25d520f` (база 2026-10-11). Зачепили продукт, межі, інфру — три блоки: **catalog full-text search** (`9178c04` — міграція 00016, `search_vector` tsvector GENERATED + GIN, config `simple` бо PG не має укр-словника, замінив ILIKE), **refund** (PR #19: міграція 00017, домен `PaymentRefunded` + refund-гілка в `DecidePaymentEvent`, WayForPay `Refunded`/`Voided`→подія, order-scoped revoke гранту в `Apply`, адмін-черга рефандів `/admin/refunds`; `make check` exit 0, e2e 115), **lesson-builder** (PR #20 `ddd459f`, LB00–LB15: структурований документ уроку — reader + lazy Tiptap editor під `/admin`, **новий 9-й доменний модуль `image`** з filesystem+postgres адаптерами й nginx image-serve гейтом під `/media/i/`, міграції 00018 (lesson_documents) / 00019 (drop flat lesson.description); `make check` exit 0, frontend 720 unit, e2e 118). Дефект, спійманий лише повним e2e на LB15: LB06-хвіст — lesson-level `description` у playback-фікстурах → 400. Перезібрано: шапку, §1 (9 модулів, +image, catalog/refund/lesson-builder закриті, +архітектурні факти №9 два media-гейти / №10 документ у `course`), §2 (продукт += документи уроку/зображення/full-text пошук/refund-черга), §3 (9-й модуль `image` seam + nginx image gate + lesson-documents у `course` + Tiptap), §4 (міграції 00016–00019, nginx image serve), §5 (**Tiptap — нова runtime-залежність**), §6 (refund-ризик знято, +ліміти lesson-builder). Історія re-sync'ів нижче збережена. |
| Гейт свіжості 2026-10-11 | **76 комітів** від `9c0e64f` (база 2026-10-07). Зачепили продукт, межі, інфру: фаза **learning** (LR0–LR6, огляд учнів інструктора/адміна + блок автора з черги, G8 «адмін — не учень»; merge PR #16 `066bef9`), **admin-реконструкція** (NorthWind shell S0 `98fc79f`, /admin/users Fleet-таблиця S2 `53a0476`), **admin-stats** (новий 8-й доменний модуль `adminstats`, T1–T11, /admin/analytics; merge PR #17 `f88cf96`), **рев'ю-фікси** (role+sort фільтр /admin/users end-to-end, hex→токени, formatLocale; PR #18 `25d520f`), **CI-перебудова K10→K12→K13** (CI знову гейт; легкий per-PR + e2e на тегах; Go-кеш; свіжа e2e-БД), nginx e2e-burst. Перезібрано: шапку, §1 (8 модулів, learning/admin закриті, P0 нема, +факти adminstats і G8), §2 (продукт += analytics/learner-overview/roster-filter), §3 (8-й модуль `adminstats` seam, roster role/sort, learner overview у `course`), §4 (CI K12/K13), §5 (Storybook dev-dep), §6 (CI-ризик знято, +ліміти admin-stats). Історія re-sync'ів нижче збережена. |
| Гейт свіжості (архів) | **38 комітів** від `2c0d29f` (база оплати). Зачепили продукт, межі й ризики: уся фаза **оплати, зрізи O01–O06a** (`e8ecd56`…`aeff91c`), дизайн-прохід клієнта й адмінки до мокапу (`9031e2f`, `1c90bfc`, `311ceb5`, `32c8a1d`, `22eead4`, `5e4ddb4`, `016de03`, `1b54c10`), чернетка спеки рейтингів (`d5543fd`), інструмент `visual-compare` (`1fe6cef`, `24381b2`). Перезібрано: шапку, §1 (оплата O01–O06a закрита, O06b/O06c/O07 відкриті; живий P0 лишається оплата), §2 (продукт += ціна/гранти/замовлення), §3 (новий модуль `orders`, ADR 014, розширений контракт доступу, дизайн-токени), §4 (без змін у гейтах, крім нового покриття), §5 (дві нові фронт-залежності — шрифти), §6 (два нові відкриті ризики — тестові флейки R23 і identity-інтеграція; старі ризики перенесено). Історія — у git. **Re-sync 2026-10-03:** точково оновлено під закриття фази оплати (хвіст O06b/O06c/O07, B-25, Pay25-live; `aeff91c`→`df5b380`) — §1, шапка, таблиці статусу та ризик 4; не повний rebuild. **Re-sync 2026-10-04:** точково оновлено під закриття фази рейтингів (spec-flow GATE 1/2, Rt01–Rt30, PR #5 `14aeb3b`, e2e 98 passed; 4 дефекти знайдено й закрито) та хвостів B-28/B-23 (PR #6 `83c5142`); `df5b380`→`83c5142` — §1, шапка, ризики; не повний rebuild. **Re-sync 2026-10-06:** точково оновлено під закриття фази **accounts** (self-service identity: реєстрація/верифікація/reset/зміна пароля-email/TOTP 2FA/CSP; GATE 1 G1–G14, GATE 2, Acc01–Acc31, PR #7 `d263045`, e2e 101; email-інфра Mailpit, міграція 00014, крос-модульний гейт верифікації через `access.CanStream`); `83c5142`→`d263045` — §1, шапка; не повний rebuild. **Re-sync 2026-10-07:** точково оновлено під закриття фази **enrollment** (членство як авторитетний факт доступу: один рядок `course_grants` на (user,course), стан active/blocked/revoked, free-контент вимагає членства; self-enrol/block/unblock/revoke, ростер із прогресом, скарги на учасників з адмін-чергою, предикати `CanConsume`/`CanCommunicate`, media-bytes гейт на nginx, backfill 00015 доведений на populated-даних; L1–L9, En01–En15, PR #8 `9c0e64f`, ordered `make check` green, e2e 107); `d263045`→`9c0e64f` — §1, шапка, таблиці статусу; не повний rebuild. |
| Тез із міткою `⚠️ неперевірено` | **2** (§2 — витримка `limit_req` під навантаженням, успадкована; §6 — механізм тестових флейків R23/identity невстановлений, хоч інструментовано K4/K5). CI-ризик (був ⚠️-подібний) знято K10/K12 — CI знову гейт. |
| Рішення людини (історичне) | **Оплата — остання фаза** (`PRODUCT-BRIEF.md` §24 п.21), **закрита 2026-10-03** (merge PR #3). **Модель оплати ухвалена 2026-09-28** (§24 п.22). **Провайдер ухвалено 2026-10-01:** WayForPay зараз, IBAN+QR пізніше, Paddle для ЄС (`docs/payment/provider-research.md` → Decision). **Ключ (власник, 2026-10-03):** Pay25-live пройшов на наявному SecretKey; тест-магазин `site_ex` видаляється, тож ротація не потрібна — наступний живий прогін на новому магазині з новим ключем (`docs/payment/assumptions.md`). Рейтинги — **фаза закрита 2026-10-04** (PR #5 `14aeb3b`), пройдено повним spec-flow (GATE 1 з 11 рішеннями власника про те, хто може оцінювати, самогрант заборонено, текст обов'язковий 20–2000, модерація за скаргами, видалення адміном). **Accounts (self-service identity) — закрито 2026-10-06** (PR #7 `d263045`): власник обрав identity-кластер першим (G1–G14 на GATE 1 — гейт верифікації read-only на контенті, автологін, 2FA опційне код-щоразу, remember-me 3 дні, без власного OAuth2). **Enrollment (K1) — закрито 2026-10-07** (PR #8 `9c0e64f`): власник ухвалив членство як один авторитетний рядок зі станом, free-контент за членством, self-enrol з верифікованим email, `GET /learn/memberships` як джерело «Моє навчання», revoke лише адміном. **Learning — закрито 2026-10-09** (merge PR #16): огляд учнів інструктора/адміна, блок автора з черги; **G8 (власник): адмін ніколи не учень** (не записується й не купує; модерує в адмінці). **Admin-реконструкція + admin-stats — закрито й злито 2026-10-10…11** (owner-initiated **поза §24**; PR #17 `f88cf96`, рев'ю-фікси PR #18 `25d520f`): NorthWind admin shell, /admin/users Fleet-таблиця з role/sort-фільтром, /admin/analytics (снапшот 6 доменів + лідерборд + engagement). Пройшло spec-flow GATE 1/2 + незалежне рев'ю Courses 2 (на PR #18 — до мерджу). **Refund — закрито 2026-10-11** (PR #19): WayForPay-рефанд у кабінеті → подія → order-scoped revoke гранту + адмін-черга видимості; закрив ризик §6:333 (`refund_required` без екрана). **Lesson-builder — закрито й злито 2026-10-13** (owner-initiated **поза §24**; PR #20 `ddd459f`): структурований документ уроку (reader + Tiptap editor + модуль `image`), спек/plan/GATE 1/2 пройдено; **рішення власника:** `contentType` у image-секції, редактор admin-scoped (без `/teach`-поверхні — follow-up, пізніше закрито teach-editor). **Teach-editor — закрито й злито 2026-10-31** (owner-initiated **поза §24**; GitHub PR #93/#94/#95 → `5fbaaff`): `/teach`-поверхня редактора документа уроку для інструктора на власний курс (Option B — реюз admin-редактора, продуктова межа /teach≠/admin), spec/plan/GATE 1/2 + незалежне рев'ю кожної партії PASS; виправлено пре-існуючий shared publish-bug. **Рішення власника (teach-editor):** lesson-list source = новий owner-scoped `GET /teach/courses/{id}` (реюз /admin відхилено — продуктова межа); backend-контракт порожнього publish-тіла (`Empty:true`, re-read-модель) лишено навмисно, фікс лише на фронті. **Роадмап §24 пройдено повністю; admin, lesson-builder і teach-editor — поза роадмапом.** Деплой — **відкладено власником 2026-10-03** (тест-режим, без прод-сервера); рішення людини, не черга. |

Брифінг — кеш первинних джерел, не істина. Розходиться з джерелом — править джерело,
а кеш переписується під нього.

## 1. Вердикт на один екран

`courses` — навчальна платформа: модульний Go-моноліт (**десять** доменних модулів:
`adminstats`, `clientlog`, `course`, `discussion`, `identity`, **`image`**, `media`, `orders`, `playback`,
`transcript`; плюс допоміжний `testsupport`, крос-модульний шар `contracts/` і **build-time seam
`openapispec`+`platform/openapigen`** — генерація OpenAPI з Go) і **шість** окремих виконуваних файлів
(`api`, `media-worker`, `upload`, `asr-worker`, `migrate`, **`openapi-gen`**), SSR-фронт на TanStack Start
(розбитий на feature-модулі `frontend/src/modules/*`), усе в одному Compose-стеку за nginx.

**Архітектурний факт №1 (найчастіше зрозуміють неправильно).** `backend/openapi/openapi.yaml` більше
**не редагується руками** — з 2026-10-29 (SA10/S0.2b cutover, PR #81) він **генерується рефлексією з Go
transport-типів** (`cmd/openapi-gen` → `openapispec.AllOpenAPIContributions()` → `openapigen.GenerateFromSource`,
kin-openapi) плюс prose-overlay `openapi/openapi.overlay.yaml`, і цей згенерований файл — **authoritative
контракт**, з якого Orval будує фронт-клієнт. Правка форми API йде в Go-тип або в overlay (multipart-тіла,
проза, деяка метадата) — **не в yaml**; `make generate` регенерує між sqlc і Orval, `check-generated` валить
гейт на дрейфі. Рефлектор не бачить custom `MarshalJSON` (форма H2 Block) — такі місця override-яться через
overlay й census-яться проти домену; **взаємна рекурсія named-типів fail-closed** (owner-рішення 2026-10-29).

**Поточний стан.** Фази 1–9, 7a/7b, робоче місце інструктора **та оплата** закриті (деталі нижче
в §2). **Фаза оплати закрита 2026-10-03** (merge PR #3 `df5b3802`), усі зрізи O01–O07: O01 (контракт
доступу += paid/entitled/preview, ADR 014) → O02 (міграція 00011, домен `orders`) → O03 (6 SQL-читань
права в 4 модулях) → O04 (application+HTTP замовлень, адмін-видача, grandfathering) → O05 (фронт:
каталог/урок/адмінка) → O06a (бекенд WayForPay) → O06b (фронт: `returnUrl` з `?course=`, checkout) →
O06c (e2e через стаб + тунель) → O07 (упорядкований `make check`, звіт Pay01–Pay27 порядково).
Знайдено й закрито дорогою **B-25** (повторний checkout → `(1112) Duplicate Order ID`; фікс: 409
`checkout_already_opened` + новий ордер; гроші без гранту → `refund_required` + warn). **Pay25-live
доведено живою Approved-оплатою 1 грн** на власному тест-режим-магазині власника (callback `POST→200`,
ордер `4396fba5…`→`paid`, застосовано рівно раз — `docs/payment/report.md` §6). Закриття за локальним
`make check` exit 0 (rstest 541, e2e 89); CI «verify» червоний суто environmental (`make up` не
піднімає compose у раннері), не регресія.

**Живого P0 немає — роадмап §24 пройдено повністю.** Оплата (остання фаза §24 п.21), рейтинги,
self-service **accounts**, **enrollment** (членство як авторитетний факт доступу), **learning** (огляд
учнів + блок автора), **admin** (shell + /admin/users + /admin/analytics, поза §24), **refund** (PR #19,
2026-10-11) і **lesson-builder** (PR #20, 2026-10-13, поза §24: структурований документ уроку + reader +
Tiptap editor + модуль `image`) — усі закриті; каталог отримав **full-text пошук** (міграція 00016,
tsvector замість ILIKE). **Наступного ПРОДУКТОВОГО ходу немає** — роадмап §24 пройдено; остання owner-initiated фіча **teach-editor** (`/teach` редактор документа уроку для інструктора) **закрита 2026-10-31** (`5fbaaff`). **Security-audit черга ПОРОЖНЯ:**
SA01–SA12 + SA09-T1 змерджено (#66–#80), **SA10/S0 закрито повністю**, **SA13 повністю злито 2026-10-31**
(#83–#90 + SA13-A-R owner-accepted no-code); лишаються лише S01/S02 блоковані TLS/deploy. lecturer-initiated revoke —
**відхилено власником 2026-10-13** (revoke лишається адмінським; лектор діє через block + report).
Деплой — рішення власника (відкладено 2026-10-03, test-mode). Деталі — `work/AGENDA.md` (ребілд 2026-10-29).

**Фаза рейтингів і відгуків закрита 2026-10-04** (PR #5 `14aeb3b`). Пройдено повним spec-flow: GATE 1
(R1–R6 + 11 рішень власника), GATE 2 (критика/захист свіжою 4.8-сесією), приймальна таблиця Rt01–Rt30,
T01–T08. Міграція `00013_ratings.sql`, контракт `internal/contracts/rating`, доменний `review`
(`discussion`), сторефронт/навчання/модерація на фронті, e2e `ratings.spec.ts` з axe. Повний
упорядкований `make check` зелений (e2e 98 passed). Під час e2e знайдено й закрито 4 реальні дефекти
за зеленим гейтом (strict-схема learning course без rating/lessonCount; курсор `""` замість `null`;
RadioGroup без CSS перехоплював кліки; застарілий список після запису). Рейтинги живуть у наявному
модулі `discussion` (доменний `review`) плюс новий крос-модульний контракт `internal/contracts/rating`
— **нового модуля не додано**, доменних модулів лишається сім (межі див. §3).

**Хвости B-28/B-23 закриті 2026-10-04** (PR #6 `83c5142`): B-28 — курсор останньої сторінки відгуків
тепер `null`, не `""` (handler→OpenAPI nullable→Orval→фронтенд); B-23 — застарілий знімок у B39/O1
playback-e2e (перечитування після poll). B-27 (видалити TEMPORARY ryuk-retry, коли реліз має
testcontainers-go #3868) **лишається відкритим** — тримається version-pin тестом.

**Фаза enrollment закрита 2026-10-07** (PR #8 `9c0e64f`, ordered `make check` green, e2e 107, En01–En15):
членство як один авторитетний рядок `course_grants` на (user,course) зі станом active/blocked/revoked
(міграція 00015: повний `UNIQUE(user,course)` замість часткового active-індексу, таблиця
`participant_reports`, backfill self-active для наявних споживачів free-курсів — доведений DownTo/UpTo-тестом
на populated-даних); free-контент тепер вимагає членства; self-enrol (email-first), instructor
block/unblock, admin revoke, ростер із round-half-up прогресом, скарги учень→лектор/лектор→учень з
адмін-чергою (discussion читає членство власним SQL, ADR-013); предикати `access.CanConsume`
(active|blocked) і `CanCommunicate` (active) у 7 читачах; гейти контенту (playback×2/transcript/course,
вкл. media-bytes через nginx `auth_request` — e2e на реальних байтах) і комунікації з окремими
`enrollment_required`/`enrollment_blocked`, названими в UI; `GET /learn/memberships` для «Моє навчання».
Рев'ю на кожному зрізі зловило до merge: revoked-as-member UI-гілку, два focus-drop'и, неназвану
comms-помилку, мертвий revoke на lecturer-скаргах, En13 proof-gap. Ліміти — `docs/enrollment/report.md` §6
(paid re-enrol після revoke відмовлено за дизайном; «Моє навчання» без пагінації — bounded).

**Фаза learning закрита 2026-10-09** (merge PR #16 `066bef9`, LR0–LR6, e2e 108): огляд навчання для
інструктора/адміна (`GET /teach/learners`, `/admin/learners`, backend у модулі `course`
`backend/internal/course/domain/learning.go`) і блок/розблок автора коментаря/відгуку прямо з черги
модерації (той самий `transitionMembership`, що й ростер). **G8 (власник): адміністратор ніколи не
учень** — не записується й не купує (змінило enrollment G10/En14; Rt05 переписано на студента),
модерує в адмінці, де доступно все. Ліміти — `docs/learning/report.md` §5.

**Admin-реконструкція + admin-stats закриті 2026-10-10…11** (owner-initiated **поза роадмапом §24**;
PR #17 `f88cf96` + рев'ю-фікси PR #18 `25d520f`). S0 — NorthWind admin shell (icon-rail + секційний
сайдбар + workspace-header); S2 — `/admin/users` як Fleet-таблиця з фільтр-баром (status-сегмент +
**role-фільтр + sort email/name/recent**, додані рев'ю-фіксом end-to-end). Міні-фаза **admin-stats**
(spec-flow GATE 1/2, T1–T11, e2e 114) — **новий 8-й доменний модуль `adminstats`** з двома admin-only
read-ендпойнтами (`GET /api/v1/admin/stats`, `/stats/timeseries`) і екраном `/admin/analytics`
(снапшот 6 доменів + лідерборд інструкторів + engagement + трендовий графік). Цикл рев'ю:
/code-review + /pattern-architecture-review → незалежне рев'ю Courses 2 (підтвердило, поглибило #4) →
рев'ю-фікси → повторне рев'ю Courses 2 зелено (на PR #18, **до** мерджу). Ліміти — `docs/admin-stats/report.md`
§5 (deferred watch-through/теги/популярність = placeholder; 9b доведено unit+repo не live-e2e; без
engagement-індексу; тренд лише day/week). **Урок:** PR #17 було змерджено до рев'ю Courses 2 (помилка,
виправлена постфактум); для код-фаз рев'ю Courses 2 — **до** мерджу.

**Фаза accounts (self-service identity) закрита 2026-10-06** (PR #7 `d263045`, e2e 101, Acc01–Acc31):
публічна реєстрація, верифікація email, self-service скидання пароля, зміна власного пароля/email,
опційне TOTP 2FA (двокроковий вхід, резервні коди, admin-disable), login rate-limit, remember-me —
на email-інфрі Mailpit (міграція 00014: `email_verified`/`totp_*`/`account_tokens`/`recovery_codes`).
Гейт верифікації — крос-модульний: новий чистий предикат `access.CanStream` (окремо від `CanBeViewedBy`,
щоб не ламати архітектурні AST/TruthTable-тести) + `email_not_verified` на мутаціях; факт `EmailVerified`
протягнуто через `httpserver.Actor`, не кожен модуль читає `users`. Нова shared-інфра `platform/mail`.
CSP: nonce у `server.mjs`, **enforcing у проді / report-only в e2e** (axe інжектить unnonce'd скрипт).
**Власний OAuth2-сервер свідомо НЕ піднімали** (рішення власника): сесія вже HttpOnly+Secure+SameSite-кукі;
сильніший захист дали CSP+2FA. Обмеження — `docs/accounts/report.md` §5 (TOTP-секрет у БД як є; style-src
unsafe-inline; e2e-only послаблення login/CSP; соцлогін/обов'язкове-2FA/WebAuthn відкладено).

**Архітектурний факт №1 — досі вірний.** Відмова 403 на сегменті нічого не вирішує про сесію
(`docs/adr/012`); рішення ухвалює heartbeat, не сегмент. Обмеження позиції тривалістю `ready`-ассета
навмисно продубльоване в чотирьох доменах (`course`, `playback`, `discussion`'s крос-тест); оплата
нічого тут не зачепила.

**Архітектурний факт №2.** Розпізнавання мовлення — усередині `backend/internal/transcript`, не
окремий модуль; `asr-worker` лише виконуваний файл (`backend/cmd/asr-worker/main.go:14-17`).

**Архітектурний факт №3.** Воркер ASR пише чернетку без HTTP-актора й без токена власника; конкурентність
ловить оптимістичний замок, не роль-перевірка.

**Архітектурний факт №4 — оновлено оплатою: контракт доступу тепер несе П'ЯТЬ фактів, мапінгів
лишилось ЧОТИРИ.** `contracts/access.Course` дістав `Paid`, `Entitled`, `Preview` (ADR 014, доповнює
013): `CanView = accessExempt(subject,course) ∨ Published ∧ (¬Paid ∨ Preview)`, де
`accessExempt = Administrator ∨ owner ∨ Entitled` (`backend/internal/contracts/access/access.go:20-30`).
Поряд — чиста `PurchaseRequired = Published ∧ Paid ∧ ¬Preview ∧ ¬accessExempt`: усі application-шари,
що відмовляють у перегляді платного курсу, кличуть цей **один** предикат і обирають бізнес-код
`course_purchase_required` замість generic `forbidden` (A6, `docs/payment/plan.md`). Мапінгів
лишилось **4** (`course`, `playback`, `transcript`, `discussion`) — оплата не додала п'ятого; кожен
і далі `CanBeViewedBy` = рівно `return access.CanView(…)`, пінить той самий AST-тест
(`platform/architecture/access_mapping_ast_test.go`, `expectedMethods=4`). Коли приходить право
власності на курс, це один факт `Entitled` — одне поле в контракті, один join `course_grants` у SQL
кожного модуля (6 запитів у 5 файлах, `docs/payment/plan.md` §«Де ляже join»).

**Архітектурний факт №5 — новий модуль `orders` володіє правом, грошима, inbox і аудитом, інші
модулі лише ЧИТАЮТЬ його таблицю.** `backend/internal/orders` (domain: `Money` у мінорних одиницях,
`Order`, `Grant`, `PaymentEvent`; application: `StartPurchase`/`PurchaseStatus`/`ApplyPaymentEvent`;
ports: `provider.go`, `repository.go`; adapters: `http`, `postgres`, `wayforpay`). Інші чотири модулі
джойнять `course_grants` у своїх SQL-адаптерах (прецедент — `discussion` читає `courses` напряму),
але **не імпортують** пакети `orders` — architecture-тест пінить відсутність імпорту в обидва боки
(дзеркало `discussion_ownership_test.go`, `docs/payment/plan.md` A2).

**Архітектурний факт №6 — порт провайдера з'явився разом із першим адаптером, не раніше (AGENTS.md).**
До O06a порту платіжного провайдера в коді не було (A4: «ports require an actual external
dependency»). `orders/ports/provider.go` і `orders/adapters/wayforpay/` народились в одному зрізі;
`ApplyPaymentEvent` — провайдер-незалежне ядро, яким користується і вебхук, і `CHECK_STATUS`.

**Архітектурний факт №7 — модуль `adminstats` лише ЧИТАЄ: сирий read-only pgx, без sqlc, агрегує
6 доменів власним SQL.** `backend/internal/adminstats` (domain: чисті value-типи + `UserSort`/валідація
range/metric/bucket; application: тонкий сервіс, що штампує `singleInstructor` і делегує; ports:
2-методний `StatsRepository`; adapters: `postgres` raw pgx над спільним пулом, `http`). `Snapshot`
читає users/courses/grants/orders/reports/media плюс лідерборд інструкторів і engagement; крос-табличні
читання — у власному адаптері модуля (прецедент orders/discussion, ADR 013), **без імпорту чужих
модулів**. Admin-гейт живе **лише** в `RequireAdministrator`-middleware (сервіс актора не бачить —
немає недосяжної defensive-гілки). Орвал-типи не покидають адаптер; Zod `.strict()` на фронті.
(`backend/internal/adminstats/**`, `docs/admin-stats/report.md` §1/§3).

**Архітектурний факт №8 — адміністратор не є учнем (G8), інваріант у сервісі, не лише в UI.** Адмін
не записується й не купує; enrol-CTA сховано на фронті, а сервіс відмовляє на бекенді
(`backend/internal/course/domain/learning.go`, `docs/learning/report.md`). Це змінило попередній
enrollment-інваріант (адміну грант не видається).

**Архітектурний факт №9 — під `/media/` тепер ДВА gated-шляхи, не один.** AGENTS.md заявляє, що
байти досяжні лише під `/media/p/<playbackId>/` через `auth_request`. Lesson-builder додав другий:
`/media/i/<id>.<ext>` для зображень документів уроку, з **власним** authorize-гейтом у модулі `image`
(nginx `auth_request` → image-handler, не playback). Зображення зберігаються як `<id>.<ext>`, щоб nginx
віддавав коректний content-type (рішення LB10a). Будь-яка зміна media-гейта тепер стосується обох
шляхів; читати інваріант як «лише playback» — помилка (`infrastructure/nginx/nginx.conf`,
`backend/internal/image/**`, `infrastructure/tests/image-authorize.sh`).

**Архітектурний факт №10 — документ уроку живе в модулі `course`, не в окремому модулі.** Lesson-builder
додав структурований документ уроку (reader + Tiptap editor), але доменна логіка документа — всередині
`course` (`lesson_documents`, PK `(lesson_id, locale)`), окремий доменний модуль додано **лише для
зображень** (`image`). Шукати «модуль document» на бекенді — хибний слід (`backend/internal/course/**`,
міграція 00018, `docs/lesson-builder/report.md` LB01–LB05).

## 2. Продукт

| Питання | Відповідь | Джерело |
|---|---|---|
| Що це | Платформа онлайн-курсів: публічний каталог, вхід, адмінське керування курсами, вивантаження відео, відтворення з прогресом, транскрипт, гілки коментарів, робоче місце інструктора, **разова покупка курсу з видачею права на перегляд** | `work/PRODUCT-BRIEF.md` §21, `docs/payment/plan.md` |
| Для кого | Студент (дивиться, читає транскрипт, коментує, **купує платний курс, бачить стан замовлення**), адміністратор/автор (курси, акаунти, модерація, **ціна курсу й прев'ю-уроки, ручна видача права**) | `AGENTS.md` §Scope, `docs/payment/assumptions.md` |
| Формат постачання | Compose-стек: `api`, `frontend` (SSR), `media-worker`, `upload`, `asr-worker`, `nginx`, `postgres`, `redis` | `docker-compose.yml`, `backend/cmd/` |
| Що працює й **виміряно** | Плеєр, навчальна поверхня, прогрес, транскрипт, ASR-чернетка, гілки коментарів, робоче місце інструктора — усе без змін із попередньої редакції (деталі § нижче, не повторено тут). **Нове, оплата:** контракт доступу з ціною/правом/прев'ю (O01, unit+property), модуль `orders` (O02, unit), 6 SQL-читань права (O03, integration), старт/статус покупки + адмін-видача + grandfathering (O04, contract+integration), фронт каталогу/уроку/адмінки з ціною й «Купити» (O05, 88/88 e2e), **бекенд WayForPay: checkout-форма, підписаний callback, CHECK_STATUS** (O06a, пройшло рев'ю, гонку даних у статусі закрито) | `docs/payment/plan.md`, `work/review-o0{1..6a}*.md` |
| Чого **ще немає** | Продакшн-розгортання (ФОП + публічний `serviceUrl`). Per-course блокування авторів і крос-курсовий learning-overview (залишки enrollment, за потребою). Соцлогін/обов'язкове-2FA/WebAuthn — поза accounts | `docs/enrollment/report.md` §6, `docs/accounts/report.md` §5, `docs/payment/report.md` §6 |
| Активний P0 | **Немає — роадмап §24 пройдено повністю, enrollment закрито (2026-10-07).** Наступного продуктового ходу немає; деплой — рішення власника | `work/AGENDA.md`, `docs/enrollment/report.md` |
| Порядок фаз далі | 7a ✓ · 8 ✓ · 9 ✓ · 7b ✓ · інструктор ✓ · оплата ✓ (10-03) · рейтинги ✓ (10-04) · accounts ✓ (10-06) · enrollment ✓ (10-07) · **learning ✓ (10-09)** — роадмап §24 завершено; **admin-реконструкція + admin-stats ✓ (10-10…11, поза §24)** | `PRODUCT-BRIEF.md` §24; `work/AGENDA.md` |
| Адмін-аналітика (admin-stats, ✓ 2026-10-11) | `/admin/analytics`: снапшот 6 доменів (users/courses/grants/orders/reports/media) + лідерборд інструкторів + engagement (active-learners-30d / сер. частка завершених уроків / completions) + трендовий графік (signups/enrollments/orders/revenue, day/week); admin-only, `Cache-Control: no-store`. Deferred-метрики (watch-through/теги/популярність) — чесний placeholder, не фабриковані числа | `docs/admin-stats/report.md`, `backend/internal/adminstats/**` |
| Огляд навчання + admin roster (learning + S2, ✓) | Інструктор/адмін бачать учнів курсу (`/teach/learners`, `/admin/learners`); `/admin/users` — Fleet-таблиця з пошуком, status-сегментом, **role-фільтром і sort (email/name/recent)**; блок автора коментаря/відгуку з черги модерації | `docs/learning/report.md`, `frontend/src/presentation/admin/UserListView.tsx` |
| Модель оплати (ухвалена, §24.22) | Разова покупка → безстроковий grant на курс (не підписка); один провайдер-адаптер; одна валюта (UAH); без повернень у MVP | `PRODUCT-BRIEF.md` §24 п.22, `docs/payment/assumptions.md` |
| Провайдер (ухвалено 2026-10-01) | **WayForPay** зараз (hosted redirect, підписаний callback, UAH); **IBAN+NBU QR** пізніше (ручне підтвердження адміном спершу, автоматичне через monobank-вебхук — окремий крок); **Paddle** як другий провайдер для ЄС (Merchant of Record, VAT) пізніше. Закриті альтернативи: LiqPay/Portmone/Fondy (Fondy — ризик: NBU 2024 скасував реєстрацію укр. юрособи), Stripe (не онбордить Україну напряму), Paddle як перший провайдер (UAH є, але ~6% комісія, виплати лише в USD/EUR/GBP/AUD/CAD) | `docs/payment/provider-research.md` §Decision, §Recommendation |
| Правило доступу (A1/ADR 014) | `CanView = admin ∨ owner ∨ entitled ∨ (published ∧ (¬paid ∨ preview))`; `PurchaseRequired` — спільний предикат відмови (A6) | `backend/internal/contracts/access/access.go:20-30` |
| Grandfathering (A7) | Коли автор уперше ставить курсу ціну, усі студенти з playback-прогресом на курсі дістають активний grant `source=grandfather` в одній транзакції з `UPDATE courses SET price_minor` | `docs/payment/plan.md` A7, `docs/payment/assumptions.md` |
| Дефект, знайдений і закритий 2026-10-01 | Навчальна сторінка платного курсу брала `Preview` лише з ПЕРШОГО рядка SQL (перший урок першого модуля) — платний курс із прев'ю не на першому уроці повністю відмовляв `course_purchase_required`, прев'ю ставало недосяжним. Рішення власника: нелогінований-без-права бачить **всю структуру** курсу, кожен урок має власний `locked`-прапор | `docs/payment/assumptions.md` «The learning outline of a paid course» |
| Дефект, знайдений і закритий 2026-10-01 | Запит навчальної сторінки ніколи не читав `price_minor`/`currency` — ціна завжди була `null`, доки покупка-панель це не виявила; виправлено мапінгом через `domain.NewCourse` | `docs/payment/assumptions.md` «The learning course never carried its price» |
| Дизайн-прохід (2026-09-30…10-01) | Редизайн-вирівнювання клієнта й адмінки під зовнішній мокап (shadcn/Tailwind-вигляд, **не** порт залежностей — заборонено AGENTS.md); слайси D-a (токени/примітиви) → D-b (сторінка каталогу: hero, картки з ціною/рейтингом) → D-c (урок) → D-d (адмінка, темний ops-дашборд-стиль на реальному контенті курсів/модерації, не фейкові ops-віджети) | `docs/design/portal-redesign.md` |
| Рейтинги й відгуки (**закрито 2026-10-04**, PR #5) | R1 оцінює лише студент/власник гранту, що почав курс (самогрант заборонено); R2 Steam-подібний % прогресу заморожено на момент відгуку; R3 1–5 зірок + обов'язковий текст 20–2000, модерація за скаргами через робоче місце інструктора, знімає лише адмін; R4 поріг показу — 3 оцінки («Новий курс» до того); R5 — реалізовано після оплати | `docs/ratings/report.md`, `docs/ratings/plan.md` (Rt01–Rt30) |
| Юрисдикція провайдера | Закрито: WayForPay обслуговує Україну зараз, провайдер-незалежний домен не блокується вибором (A4); ЄС — окремий адаптер пізніше (Paddle) | `docs/payment/provider-research.md` |
| Архівний курс і рейтинги | **Вирішено на GATE 1 + Rt26/Rt30:** власник гранту на архівному курсі зберігає свої відгуки; публічна сторінка курсу стає 404 (рейтинги публічно не показуються); після republish рейтинг картки відновлюється | `docs/ratings/report.md` (Rt26, Rt30) |

## 3. Кодова база й seams

Модулі бекенду (`backend/internal/`): **дев'ять** доменних — `adminstats` (read-only),
`course`, `discussion`, `identity`, **`image`** (новий, lesson-builder), `media`, `orders`, `playback`,
`transcript` — плюс `testsupport`
і крос-модульний шар `contracts/` (ADR 013, доповнений 014). Кожен доменний тримає
`domain`/`application`/`ports`/`adapters`, залежності дивляться всередину. П'ять виконуваних файлів
без змін (`backend/cmd/`).

| Межа | Що її перетинає | Джерело |
|---|---|---|
| **nginx ↔ playback** — несучий шов Phase 6 | Байти медіа віддає nginx, `auth_request` на кожен плейлист/сегмент; оплата не змінила цей шов, лише додала join на право в `loadVideoCourse` (найгарячіший з 6 запитів) | `infrastructure/nginx/nginx.conf:55-86`, `docs/payment/plan.md` §«Де ляже join» #2 |
| **nginx ↔ image** — ДРУГИЙ media-гейт (lesson-builder, 2026-10-13) | Зображення документів уроку під `/media/i/<id>.<ext>` з власним `auth_request` → image-handler (не playback). Зберігаються як `<id>.<ext>` щоб nginx дав коректний content-type (LB10a). Інваріант «байти лише через auth_request» тепер стосується ДВОХ шляхів (див. §1 факт №9) | `infrastructure/nginx/nginx.conf`, `backend/internal/image/**`, `infrastructure/tests/image-authorize.sh` |
| **Новий модуль `image` (lesson-builder, 2026-10-13)** | Повний доменний модуль: `domain`/`application` (upload + authorize use cases), `ports`, `adapters` (filesystem для байтів + postgres для рядків). Зберігає байти на диск і метадані в БД; гейт доступу до зображення — у самому модулі, не імпортує інші | `backend/internal/image/**`, `docs/lesson-builder/report.md` LB07–LB10 |
| **Документ уроку живе в `course` (lesson-builder)** | Структурований документ уроку (reader + Tiptap editor) — доменна логіка в `course` (`lesson_documents`, PK `(lesson_id, locale)`), НЕ окремий модуль; editor admin-scoped під `/admin/courses/$courseId/lessons/$lessonId`; HTTP-контракт у OpenAPI/Orval | `backend/internal/course/**`, міграція 00018, `docs/lesson-builder/report.md` LB01–LB05, LB12 |
| **Full-text пошук каталогу (2026-10-13)** | `courses.search_vector` tsvector GENERATED ALWAYS (`to_tsvector('simple', title‖summary)`) + GIN-індекс; читання через `websearch_to_tsquery`, замінив ILIKE substring-match. Config `simple` бо PG не має укр-словника (`simple` коректно токенізує кирилицю) | міграція 00016, `9178c04` |
| **Новий модуль `adminstats` (read-only, 2026-10-11)** | Повний доменний модуль, що лише ЧИТАЄ: `domain` (value-типи снапшоту + `UserSort`/`Metric`/`Bucket` + валідація range), `application/service.go` (штампує `singleInstructor`, делегує), `ports/stats_repository.go` (2 методи), `adapters/{postgres,http}`. Raw read-only pgx над спільним пулом, **без sqlc**; агрегує 6 доменних таблиць власним SQL, **не імпортує** жоден модуль (ADR 013). Два admin-only ендпойнти за `RequireAdministrator` | `backend/internal/adminstats/**`, `docs/admin-stats/report.md` |
| **Admin roster фільтр/сорт (end-to-end, PR #18)** | `/admin/users` (модуль `identity`): `ListAdminUsers`/`CountAdminUsers` дістали `role_filter` (whitelist student/instructor/admin) і CASE-сортування (`sort_by` email/name/recent); порт→application→handler парсять+валідують (OpenAPI enum + власний 400-guard); фронт — два `<Select>` + i18n. Дефолтна форма запиту незмінна | `backend/sql/queries/identity.sql`, `backend/internal/identity/adapters/http/admin_handler.go`, `frontend/src/presentation/admin/UserListView.tsx` |
| **Огляд учнів (learning, модуль `course`)** | `GET /teach/learners`/`/admin/learners` — ростер учнів курсу з прогресом; backend у `course` (не новий модуль); блок автора з черги модерації через `transitionMembership` | `backend/internal/course/domain/learning.go`, `docs/learning/report.md` |
| **Новий модуль `orders` (O01–O06a)** | Повний доменний модуль: `domain` (Money/Order/Grant/PaymentEvent/errors), `application/service.go` (StartPurchase, PurchaseStatus, ApplyPaymentEvent), `ports` (provider.go, repository.go), `adapters/{http,postgres,wayforpay}`. Інші модулі читають `course_grants` SQL-ом, не імпортують `orders`; `orders` не імпортує жоден модуль | `backend/internal/orders/**`, `docs/payment/plan.md` A2 |
| **Контракт доступу — 5 фактів, 4 мапінги (ADR 013+014)** | `contracts/access.CanView`/`PurchaseRequired` — чисті функції, лише stdlib; `course`/`playback`/`transcript`/`discussion` маплять свої факти через `CanBeViewedBy` = `return access.CanView(…)`; AST-тест `expectedMethods=4` незмінний (оплата не додала п'ятого мапінгу — `orders` сам не вирішує видимість, лише постачає факт `Entitled` через SQL) | `backend/internal/contracts/access/access.go`, `platform/architecture/access_mapping_ast_test.go` |
| **Провайдерний порт з'явився з першим адаптером (A4)** | `orders/ports/provider.go` (Checkout, OrderStatus) + `orders/adapters/wayforpay` (HMAC_MD5-підпис форми й callback, CHECK_STATUS) — стандартна бібліотека, жодної платіжної SDK | `backend/internal/orders/ports/provider.go`, `backend/internal/orders/adapters/wayforpay/wayforpay.go` |
| **Callback-ендпоінт — неавтентифікована поверхня** | `POST /api/v1/payments/wayforpay/callback` без сесії й без `sameOrigin`; авторизує лише підпис; невідоме замовлення → не термінальна відмова (503, ретрай), а не reject (Pay21b) | `docs/payment/plan.md` §Side-effects; `work/archive/brief-codex-o06a-wayforpay-backend.md` |
| **Міграція `00011_payment.sql`** | `courses.price_minor`/`currency` (NULL=безкоштовний), `course_lessons.is_preview`, нові таблиці `orders`, `course_grants` (частковий UNIQUE на активне право), `payment_events` (UNIQUE provider+event_id), `grant_audit`; без каскадів на видалення курсу/користувача (фінансовий слід) | `docs/payment/plan.md` §План міграції |
| **Ціна/замовлення-ендпоінти в OpenAPI/Orval** | Каталог несе `price`/`paid`; нові операції старту покупки, статусу, адмін-видачі, чекауту; `check-generated` у ланцюгу гейта | `backend/openapi/openapi.yaml`, `frontend/orval.config.ts` |
| Правило доступу у решти модулів | `course`/`playback`/`transcript`/`discussion` незмінні щодо своєї структури; лише 4 з 6 їхніх SQL-запитів навчились джойнити `course_grants` і читати `price_minor`/`is_preview` (O03) | `docs/payment/plan.md` §«Де ляже join» (6 запитів у 5 файлах) |
| **Машинний автор ASR обходить `CanBeEditedBy`** | Без змін із попередньої редакції — §1 факт №3 | `recognize.go` |
| Клієнт ↔ сервер | Orval-моделі в адаптерах, явний мапінг; новий `orders`-адаптер фронту (`routes/orders.$orderId.tsx`, `OrderStatusView`, `LockedLessonPanel`) | `frontend/src/routes/orders.$orderId.tsx`, `AGENTS.md` §Architecture |
| **Дизайн-токени — друга денситність на тій самій системі** | `D-a` додав pill-badge, icon-tile, floating-card+glow, stat-row, segmented control, load-meter, delta-pill, sidebar-nav, mono-identifier — спільні токени для «повітряного» сторфронту й «щільної» адмінки; без Tailwind/shadcn/Radix (заборонено) | `docs/design/portal-redesign.md`, `frontend/src/presentation/ui/*` |
| Гейт «логи без URI» | Без змін; оплата додала правило «ключі провайдера лише env, ніколи в логах» як продовження того ж принципу | `backend/platform/architecture/safe_request_logs_test.go`, `docs/payment/plan.md` §Конституція |
| Дві гілки відтворення | Без змін | `hls-media.ts:81` |

ADR, обов'язкові перед відповідними шарами: `006` transcript model, `010` предикат доступу, `011`
володіння й витіснення, `012` що означає 403, `013` одне правило доступу як контракт,
**`014` контракт доступу знає платність/право/прев'ю (доповнює 013) — перед будь-якою подальшою
зміною доступу й перед спекою рейтингів**.

## 4. Інфра й доставка

| Питання | Відповідь | Джерело |
|---|---|---|
| Гейти, у порядку | `format-check`, `lint`, `typecheck`, `test`, `integration`, `check-generated`, `build`, **`e2e-reset`** (K13: порожні томи), `up`, `drain-queue`, `seed-identities`, `seed`, `smoke`, `e2e` — локальний `make check` лишається авторитетним гейтом фаз | `Makefile:165` (`check:`) |
| CI (перебудовано K10→K12, 2026-10-07…08) | CI **знову гейт** (був non-gate 09-11…10-07). Два джоби: **легкий** (`format-check`→`build`, Go-кеш `actions/cache`, per-PR `concurrency` cancel, `timeout 30`) на кожен push у `main` і кожен PR-коміт; **e2e** лише на тегах `v[0-9]+.[0-9]+.0` і `workflow_dispatch` (`timeout 45`, стек без asr-worker K10). Без pre-push | `.github/workflows/ci.yml:8-21,37-39,78-82` |
| Свіжа e2e-БД (K13, ✓ 2026-10-08) | `e2e-reset` стирає томи postgres/redis/media перед авторитетним прогоном; `make e2e` без reset — для швидкого повтору; оголив fresh-start media-data race (фікс: upload заповнює том перший) | `Makefile` `e2e-reset`, AGENDA K13 |
| CI зелений доказ | Перший повний зелений CI з 09-11 — ран 5 `37677792282` (40m16), далі K12 light-гейт ~15 хв (6 з Go-кешем); PR #18 CI «verify» SUCCESS, e2e SKIPPED (за дизайном — лише на тегах) | AGENDA K10/K12 |
| nginx e2e api-burst (learning) | e2e-стек дістав більший api-burst (`api-req.e2e.conf`), щоб суїта не ловила anti-abuse 429; прод-burst (20) незмінний | `infrastructure/nginx/api-req.e2e.conf`, commit `fade1b8` |
| Гейт O05 (фронт покупки, 2026-10-01) | Ordered `make check` exit 0: rstest **520/520**, e2e **88/88** (87+`purchase.spec.ts`); окремий прогін після 200%-зум фіксу знову зелений | `work/archive/review-o05-purchase-ui.md` §Gate result |
| Гейт O05 slice 3 (адмін ціна/гранти) | Rstest/e2e без регресії; ComboBox a11y добре, 1 nb (loading-vs-empty SR-оголошення) | `work/archive/review-o05-grant-ui.md` |
| Гейт O06a (бекенд WayForPay) | `make generate format-check lint typecheck test integration` exit 0 (реальний Docker Postgres); **e2e НЕ прогнано** (відкладено на O06c); 1 BLOCKING (гонка даних на `statusAt`) знайдено й закрито раундом 5 | `work/archive/review-o06a-wayforpay-backend.md` |
| Два відкриті тестові флейки (нові, O05) | **R23** (`playback.spec.ts`, heartbeat 403): раз упав, 9 залежних не побігли; ізольований пересів 84/84 зелений; наступний повний гейт 88/88 без повтору; механізм звужено до `loadViewablePlayback` або `sameOrigin`-перевірки, непідтверджено. **`TestAdministrationRepositoryAgainstPostgreSQL`** (identity, пакет не чіпано слайсом O05): раз упав під час integration O05 slice2, 2 пересіви зелені, текст помилки не збережено — підозра на спільний стан БД/порядок тестів | `work/archive/review-o05-grant-ui.md` (кінець), `work/archive/review-o05-purchase-ui.md` (кінець) |
| Пастка `check-generated` | Без змін — порівнює з індексом | `Makefile:59-61` |
| Пастка docker (гейт) | Без змін | пам'ять `docker-credential-helper-hang` |
| Міграції з 2026-10-11 | `00016_catalog_search` (search_vector tsvector + GIN), `00017_refund` (refund_reason), `00018_lesson_document` (lesson_documents + backfill з flat description), `00019_drop_lesson_description` (дроп колонки — оголив LB06-хвіст у playback-фікстурах). goose, lock_timeout 5s | `backend/migrations/0001[6-9]_*.sql` |
| nginx image serve (lesson-builder) | Новий `location /media/i/` з `auth_request` → image-handler; окремо від `/media/p/` (playback) | `infrastructure/nginx/nginx.conf`, `infrastructure/tests/image-authorize.sh` |
| Розгортання | Немає, не заявлено | `AGENTS.md` §Scope |
| Якорі боргу в коді | Без змін (0 TODO/FIXME/ponytail:, 1 умовний ffmpeg-skip) | `grep -rn 't\.Skip' backend` |
| Брудне дерево після гейта | Без змін — B-18 | `work/backlog.md` §B-18 |

## 5. Залежності

Перечитано проти `frontend/package.json` і `backend/go.mod` (`git diff 25d520f..HEAD`). Go-залежності
без змін (refund/catalog/lesson-builder використали наявні pgx, Orval, Query; full-text пошук — чистий
PG, без бібліотеки; зображення — stdlib `os`/`io`). **Нове 2026-10-13 (lesson-builder) — Tiptap, перша
нова RUNTIME-залежність фронту за кілька фаз:** `@tiptap/{core,react,pm,starter-kit,extension-link}`
^3.31.4 (+транзитивний ProseMirror) для rich-text editor документа уроку. Editor завантажується **lazy**
(тільки на admin-маршруті редактора), reader його не тягне. Раніше додані: два шрифти (дизайн-прохід),
Storybook (dev-only, `^10.6.1`). WayForPay-підпис — стандартна бібліотека Go.

| Що | Версія | Чому важлива |
|---|---|---|
| Go | 1.27.1 | `backend/go.mod:3` |
| **`@fontsource-variable/inter`** (нове) | ^5.3.0 | Дизайн-прохід D-a, `frontend/package.json` |
| **`@fontsource-variable/space-grotesk`** (нове) | ^5.3.0 | Дизайн-прохід D-a, комміт `7132565` |
| hls.js | 1.7.2 | ABR увімкнено, `startLevel` не задано |
| React | 19.2.7 | |
| TanStack Router / Query | 1.170.32 / 5.102.8 | Query-кеш request-scoped; `useInfiniteQuery`, замовлення інвалідують outline |
| React Aria Components | 1.21.0 | |
| Rsbuild / Rstest | 2.2.2 / 0.11.11 | |
| Playwright | 1.62.1 | `workers: 1`, `fullyParallel: false`; також рушій `frontend/tools/visual-compare` |
| **Storybook (dev)** | ^10.6.1 (+`storybook-react-rsbuild` ^3.6.0) | Каталог admin-UI-примітивів (`AdminKit.stories`); dev-only, не в прод-бандлі |
| **Tiptap (нове, runtime)** | ^3.31.4 (`core`/`react`/`pm`/`starter-kit`/`extension-link`) | Rich-text editor документа уроку; **lazy** лише на admin-editor-маршруті, reader не тягне (`docs/lesson-builder/report.md` LB12b) |
| axe-core | 4.13.0 | |
| MSW | 2.15.0 | |
| faster-whisper / Silero VAD | — | без змін |

Заборонене явно (AGENTS.md §TypeScript): Vite, Next, Vitest, Jest, ESLint, Axios, Redux Toolkit,
Radix, shadcn, Rust, ORM, service locator. **Tailwind НЕ заборонений** і є реальною залежністю
**з найпершого foundation-коміту** (`5608a03`): `tailwindcss 4.3.3` + `@rsbuild/plugin-tailwindcss`,
`@import 'tailwindcss'` у `frontend/src/app/styles/main.css:1`. House-style компонентів — семантичні
токени + scoped CSS + React Aria (не utility-класи), але Tailwind присутній і дозволений — попередня
редакція брифінгу помилково числила його забороненим (виправлено 2026-10-13, звірено з AGENTS.md + git).

## 6. Ризики

**Статус 2026-10-13 (після refund + catalog-search + lesson-builder):**
- **Refund-ризик ЗНЯТО (PR #19).** `refund_required` тепер має адмін-екран (`/admin/refunds`) + order-scoped revoke гранту на WayForPay-рефанд-подію; закрив §6:333. Залишковий борг refund (`docs/refund/report.md` §6): ми НЕ ініціюємо рефанд (лише реагуємо на зроблений у кабінеті WayForPay); черга показує сирі id, не назви; `{items}` без пагінації (магнітуда — O(десятки) за весь час, механізм росту названо); refund-before-paid reordering не реконсилюється.
- **Ліміти lesson-builder (`docs/lesson-builder/report.md` §Limitations), статус на 2026-10-31:** (1) ~~editor admin-scoped — нема `/teach`-поверхні~~ — **ЗАКРИТО teach-editor** (`/teach`-редактор, PR #93/#94/#95, `5fbaaff`); (2) один locale в UI — схема/ендпоінти locale-aware (PK `(lesson_id, locale)`), але editor/reader single-locale без перемикача (лишається); (3) ~~нема sweeper для orphaned-image~~ — **пом'якшено LB09a** (PR #91/#92): байти видаляються компенсаційно при падінні row-insert (+ detached-ctx), named crash-window residual лишається (compensation, не фоновий sweeper — рішення власника 2026-10-30); (4) scenario 7 (migrated flat description → документ) доведено на integration/domain, не e2e (колонку дропнуто 00019) (лишається); (5) Tiptap лишив один pre-existing moderate advisory (`fast-uri`, транзитивний), нова runtime-advisory не внесена (лишається).
- **Ліміти teach-editor (`docs/teach-editor/report.md`):** (1) non-blocking — якщо post-publish re-read сам падає, optimistic-lock токен відкочується до saved-токена (неминуче за контрактом порожнього publish-тіла, «re-read»-модель); (2) single-locale успадковано від lesson-builder; (3) publish-контракт лишає тіло порожнім навмисно — будь-який новий споживач publish-відповіді мусить re-read draft, не очікувати `{updatedAt}` (пастка, що оголила shared publish-bug).
- **CI знову гейт (ризик 5 — ЗНЯТО, K10→K12).** CI був non-gate з 09-11; відновлено 2026-10-07 (ран 5 зелений) і перебудовано K12: легкий per-PR гейт + e2e на тегах. Повний виклад нижче лишаю як історію того, що зловили (три класи прихованих тест-дефектів).
- **e2e config/profile-drift (ризик 6 — ЗАКРИТО, K11+K13).** Абсолютний `COMPOSE_FILE`, `--no-recreate` у recovery.spec, свіжі томи за замовчуванням (K13). Детектори drift — ID/`.Created`/hash.
- **Нові відкладені ліміти (admin-stats, `docs/admin-stats/report.md` §5):** метрики watch-through/теги/популярність — чесний placeholder, не фабриковані числа; 9b (видимий лідерборд) доведено unit+repo, не live-e2e (один e2e-стек = одне значення `PLATFORM_SINGLE_INSTRUCTOR`); без engagement-індексу (fine at scale, revisit якщо `lesson_progress` виросте); тренд лише day/week (cap 366 бакетів). **#6 з рев'ю:** визначення timeseries-метрики розкидане по доменному enum + SQL + FE — єдина мапа-джерело лише за реального росту набору метрик (датоване рішення, не борг).
- **Latent axe у `Delta`-примітиві** (`aria-label` на bare span) — не вжито на жодній сторінці, залишено (власний тест залежить від поточної поведінки).

Нові ризики попереднього прогону (оплата, O05/O06a):

1. **R23 — heartbeat 403, механізм не встановлено.** `playback.spec.ts` раз відповів 403 на
   `POST /…/playback/heartbeat`, де очікувався 200; ізольований пересів і наступний повний гейт
   (88/88) пройшли чисто. Звужено до двох джерел — правило доступу в `loadViewablePlayback`
   (`playback_forbidden`/`course_purchase_required`) або `sameOrigin`-перевірка — непідтверджено;
   неавтентифікований шлях виключено (`normalizeAuthenticationDenial` огортає лише
   `GET /internal/playback/authorize`, не heartbeat). Наступного разу — зберігати тіло відповіді й
   лог API (`work/archive/review-o05-grant-ui.md`). **Інструментовано 2026-10-07 (K4):** R23 валиться з кодом
   і requestId відмовленого heartbeat'а; `make e2e` стрімить api+nginx лог усього прогону у
   `frontend/test-logs/api-nginx.log` (переживає recreate — доведено 7× «server starting» у 107/107).
2. **`TestAdministrationRepositoryAgainstPostgreSQL` — одноразовий фейл, пакет identity не чіпано.**
   Два пересіви зелені, текст помилки втрачено; підозра на спільний стан БД/порядок паралельних
   тестів, не на зміни O05 (`work/archive/review-o05-purchase-ui.md`). Тест має власний контейнер і послідовні
   `t.Run` — міжпакетний стан виключено. **Інструментовано 2026-10-07 (K5):** `make integration`
   персистить вивід у `backend/test-results/integration.log` зі статусом `go test`'а; retry в
   `containers.Start` логує помилку. Механізм досі невстановлений.
5. **CI не був гейтом з 2026-09-11 (знайдено 2026-10-07 під час K4).** 100/100 ранів `ci.yml` у
   retention-вікні — failure; усі свіжі падають на `make up`: `asr-worker` `cpus: 6.0`/`mem_limit: 5g`
   (compose:130–131, ASR-фаза) > 2 vCPU GitHub-runner'а — детермінований range-error на create. Під
   ним латентно: CI підіймає BASE-стек (`make up`), а `purchase-wayforpay.spec` потребує overlay-порту
   8090 без skip-guard. **Цілісність фаз не зачеплена:** усі звіти/AGENTS заявляють «behind an ordered
   `make check`» — локальний гейт, єдина заявлена авторитетність; «CI green» ніде не claim'ився.
   Рішення власника 2026-10-07: полагодити як окремий хвіст після K4/K5 (CI-override для cpus/mem,
   не послаблення дефолту; CI підіймає той самий стек, що `make e2e-local`). **Стек обрано 2026-10-07
   (варіант B, AGENDA K10):** CI без asr-worker (образ 2.94 GB, whisper з HF на build, cpus 6.0 — не
   вміщається в 2 vCPU/7 GB/30 хв), решта на overlay через job-level абсолютний `COMPOSE_FILE`;
   asr.spec пропускається (`E2E_SKIP_ASR`), smoke C20 — за відсутності воркера; ASR авторитетний лише
   локально (датований ліміт); час CI міряється раном PR; застереження Courses 2 —
   повний overlay-e2e на 2 vCPU може бути повільним/флакі (тюнінг під потужний хост) → можливо larger
   runner або вужчий CI-сабсет. **Реалізовано 2026-10-07 (гілка `chore/ci-gate-k10`, PR #11):**
   `docker-compose.ci.yml` (asr-worker → `profiles: [local-only]`; media-worker cpus 1.5/mem 1536m);
   `ci.yml` job-level `COMPOSE_FILE` = 3 абсолютні файли; `E2E_SKIP_ASR` у `playwright.config.ts`;
   `smoke.sh` став config-aware — читає факт зі змонтованого стека замість припущення base (на
   overlay було ЧОТИРИ false-red, не один: ASR readyz/C20, C15 asr-метрики, WFP callback 404 base /
   401 `invalid_callback` overlay, login burst 5 base / 50 `login-req.e2e.conf`). Strong-host доказ
   CI-шляху (down -v → 3 файли → smoke → `E2E_SKIP_ASR=1` e2e): `make up` 37 с, smoke 0, e2e 101
   passed за 750 с, media 91 с (+28 % проти 71 с без капу), playback 229 с (+12 % проти 204),
   media-worker `OOMKilled=false RestartCount=0` — кап коректний; час на 2 vCPU — лише з CI-рану PR.
   **CI-рани PR #11 на 2 vCPU (2026-10-07):** вперше з 09-11 пройшли `make up`/seed/smoke і оголили
   три класи прихованих тест-дефектів за зеленими локальними гейтами (reaper sleep-race у TEMPORARY
   пакеті; discussion K19/K20 — `getByText` збігався з відкритою textarea, не з опублікованою
   відповіддю; playback P04/P05/P07/B18 — читання `videoTime` після `playing`+poll міряло швидкість
   тест-процесу, P05 — рухома ціль, бо перша сторінка грає далі) — усі закриті test-only, датовано в
   `docs/discussion/report.md` §5 і `docs/playback/report.md` §6. Час: pre-browser ≈16.5 хв (Go без
   кешу), браузер ≈2.4× strong-host → зелений ран ≈38–40 хв; рішення — `timeout-minutes: 45`, один
   ран на зміну (`push` лише `main` + `pull_request`). **Закрито 2026-10-07: ран 5 (37677792282) —
   перший повний зелений CI з 09-11, 40m16s** (pre-browser ≈15 хв, Install browser 6m19, Browser
   tests 18m49 — 101 passed, ×1.5 проти strong-host); PR #9/#10/#11 злиті (`ba53961`). **K12
   (рішення власника):** CI на кожен коміт PR — лише легкі кроки з Go-кешем і per-PR concurrency,
   таймаут назад до 30; e2e в CI — на теги `v*.*.0` і `workflow_dispatch`; без pre-push; локальний
   ordered `make check` — авторитетний гейт фаз (AGENDA K12).
6. **e2e config-drift усередині suite — закрито 2026-10-07 (AGENDA K11), profile-drift лишається.** Bare
   `docker compose` у специ резолвились у base-конфіг і перестворювали api/nginx/frontend посеред
   overlay-suite (CSP report-only→enforcing, api без WayForPay-стаба). Фікс: рецепт `e2e` експортує
   **абсолютний** `COMPOSE_FILE` (відносний вимикає parent-discovery й ламає frontend-cwd виклики —
   виміряно `config --hash`), `--no-deps` у `recreateApi()` (виміряно: un-scoped recreate на compose
   v5.5.1 не каскадив — hardening). Доказ `make e2e-local` 107/107: frontend/nginx `.Created` = старт
   suite крізь recovery, CSP тримається. **Лишається:** api один раз перестворюється в recovery через
   порядок проєктів (playback.setup 2/6 → recovery → teardown #106 30/180): bare `up` повертає дефолти
   (`d3ffda`→`01d97f`), виміряно ізольовано; бенігно, варіанти в K11. Детектори drift — ID/`.Created`/hash;
   «server starting» і `ps` «Up X» — НЕ детектори (старти процесу / час від старту).
3. **Провайдерна поверхня — callback без автентифікації (прийнято, не дефект).** `POST
   /api/v1/payments/wayforpay/callback` авторизує лише підпис; `zone=api` nginx-ліміт і тіло-ліміт —
   єдиний захист від зловживання до підпису (`docs/payment/plan.md` §Side-effects).
4. **Фаза оплати закрита (2026-10-03).** Pay25-live доведено production-шляхом: жива Approved-оплата
   1 грн на власному тест-режим-магазині через ngrok, справжній callback `POST→200`, застосовано
   рівно раз (`docs/payment/report.md` §6). **Оновлено 2026-10-11 (PR #19):** повернення
   (`Refunded`/`Voided` після `Approved`) тепер обробляються — order-scoped revoke гранту + адмін-черга
   `/admin/refunds`; попередній «refund_required без екрана» знято (див. статус угорі §6).
   Тест-магазин `site_ex` видаляється власником — ключ, що світився в чаті, зникає разом із ним.

Ризики, перенесені без змін (детальний виклад — попередня редакція, не дублюється тут):
5. Нативна гілка HLS недоведена (D-9). 6. Стеля 5000 реплік не заміряна під навантаженням.
7. Другий мовний контент не наповнюється. 8. GATE 2 плану `7a` не зійшовся (9 знахідок третього
раунду). 9. Дві паралельні 9-кейсові таблиці істинності (`coursevisibility.Cases()` vs
`accesstest.TruthTable`) — дубль не закрито. 10. Межі Phase 7b (безповоротне прибрання коментаря,
без per-user ліміту публікацій, без пагінації відповідей). 11. `B-22` копії селектора «поточне
відео». 12. `B-23` **закрито 2026-10-04** (PR #6): це був дефект тесту — застарілий знімок до
poll у B39/O1; фінальні перевірки тепер перечитують `readWrites()` після poll. 13. Межі ASR (integration C14 500мс не справжній таймаут).
14. `B-18` брудне дерево скриншотів. 15. `B-19` ранг 360p. 16. `B-21` читання транскрипту без мови.
17. Успадковані Phase 6 обмеження. 18. `S01`/`S02` заблоковані на TLS; `B-1` спостережуваність
фронту. 19. ⚠️ `limit_req zone=media` під реальним NAT-профілем не заміряно.

Джерела рядків 5–19: `docs/playback/report.md` §6, `docs/transcript/report.md` §6, `docs/asr/report.md`
§6–7, `docs/discussion/report.md` §5, `docs/readiness/report.md` §7, `work/backlog.md`
(B-18, B-19, B-21, B-22, B-23), `work/tasks.md:36-37`.

## 7. Що читати далі

`AGENTS.md` — контракт репо. `work/AGENDA.md` — ранжування наступного ходу: фази оплати й рейтингів
закриті, хвости B-28/B-23 закриті, **роадмап пройдено, автоматичного наступного ходу немає** —
потрібен вибір власника між осями (K1 enrollment без деплою / K2 деплой). Деплой відкладено (власник,
2026-10-03: поки все в тестовому варіанті, без прод-сервера). `docs/ratings/report.md` — фаза рейтингів.
`docs/payment/{spec,plan,
assumptions,provider-research,iban-qr-research}.md` — уся фаза оплати, рішення A1–A7, провайдер.
`docs/design/portal-redesign.md` — дизайн-прохід D-a…D-d. `docs/ratings/spec.md` — наступна
(не почата) фаза. `docs/adr/014` — контракт доступу знає платність (доповнює 013); `010`/`011`/`012`/`013`.
`PRODUCT-BRIEF.md` §24 п.21/п.22 — ухвалена модель оплати. `work/backlog.md` — відкриті записи
(B-18, B-19, B-21, B-22, B-23) + два нові тестові флейки з цього прогону (§6 п.1–2, ще не заведені
в backlog під номером — варто завести при наступному дотику).
