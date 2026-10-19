ДАЛІ: Прочитати enrollment-план, наявні e2e-шаблони та L7 UI/контракти; визначити точні селектори і лише тоді змінювати specs.

## Крок 1 — старт

Прочитано `work/brief-codex-L8.md`, `work/codex-working-model.md`, `AGENTS.md` і репозиторний briefing. Межі зафіксовано: змінюю лише `frontend/e2e` specs; git, `work/tasks.md`, фазовий report, generated і product/app code не чіпаю. Docker/e2e не запускатиму.

ДАЛІ: Оновити два застарілі очікування `/learn` у route guards, потім додати ізольований run-namespaced enrollment e2e spec на фактичних L7 селекторах.

## Крок 2 — звірка контракту й UI

Прочитано enrollment plan/spec, L7 routes/components, переклади, generated API surface та шаблони `purchase`, `learn-access`, `instructor`, `discussion`. Контракт для всіх потрібних операцій уже існує: self-enroll, membership/roster, block/unblock, revoke, reports і resolve. Реальні селектори: `Записатися`, `Ви учасник`, `Почати навчання`, `Моє навчання`, `Учасники`, `Поскаржитися на викладача`, `Черга скарг учасників`.

En04 media-bytes-blocked, En13 backfill та En10 grandfather/preview свідомо не додаватиму: brief вимагає не вигадувати e2e для integration/L3 доказів. Інших наявних e2e-конфліктів із L7, крім двох `/learn` redirect assertions у `route-guards.spec.ts`, під час статичного проходу не виявлено; Courses 1 має підтвердити це повним `make e2e`.

## Discussion

**Неточність brief щодо R36 empty state.** Brief наказує перевірити на signed-in `/learn` «заголовок/порожній стан». Але `docs/enrollment/plan.md` §5 прямо каже, що seed-студент має self-active членства на seeded free courses, а `frontend/e2e/route-guards.spec.ts` використовує саме `authStatePath('student')`. У runtime `GET /api/v1/learn/memberships` тому поверне щонайменше одну картку, і `MyLearningList` (`frontend/src/presentation/learn/MyLearningList.tsx:7-13`) не рендерить empty state. Очікування empty state зробило б spec хибно червоним. Пропонована поправка: R36 пінить стабільну зміну — URL лишається `/learn` і видимий заголовок «Моє навчання»; сам empty state покривається окремим run-scoped або fixture-control сценарієм. Тимчасово реалізовано цю вузьку, фактичну перевірку; потрібне рішення Courses 1, якщо потрібен окремий empty-state journey.

ДАЛІ: Побудувати `enrollment.spec.ts` з API lifecycle helpers і UI journeys En01/En03/En04/En05/En07/En08/En14/En15 та L7 accessibility checks.

## Крок 3 — існуючі route guards

Оновлено `frontend/e2e/route-guards.spec.ts`: усі три застарілі signed-in `/learn` очікування тепер перевіряють `/learn` + «Моє навчання», не каталог. `/learn/go/intro` не змінювався.

## Знахідка для Courses 1

**L7 communication error не доходить до UI кодом.** Backend планує `enrollment_blocked` для publish/comment/review, але `frontend/src/adapters/api/discussion-repository.ts:119-132` дозволяє для publish 403 лише `origin_forbidden` і `course_forbidden`; `frontend/src/domain/discussion/discussion.ts:1-18` взагалі не має `enrollment_blocked`/`enrollment_required`. Отже відповідь backend з точним кодом деградує в generic HTTP 403, і UI не може показати named enrollment error. Це не змінюю (поза межами product/app code). Enrollment e2e перевірить real backend response code через мережеву відповідь; Courses 1 має вирішити окремий frontend fix, якщо acceptance вимагає текстовий UI error.

ДАЛІ: Запустити дозволені `frontend` typecheck і lint; виправити тільки помилки e2e spec.

## Крок 4 — новий enrollment e2e

Додано `frontend/e2e/enrollment.spec.ts` з run-namespaced published free courses і `afterAll` archive. Покрито UI/API journeys En01, En03 (free), En04, En05, En07 (учень→лектор, лектор→учень, admin resolve), En08, En14, En15 та axe/44px/200%-overflow на нових L7 екранах. En04 перевіряє реальну 403-відмову до повторного self-enroll; En05 пінить реальний backend `enrollment_blocked` з POST коментаря. En03 paid path не додано: L8 brief вимагає paid і free, але лише free-course lifecycle задокументований як доступний API journey; paid setup у наявному `purchase.spec.ts` проходить через editor UI і окрему price flow. Це кандидат Courses 1 для e2e-прогону/доповнення, якщо він вважає обидва полюси потрібними в одному enrollment spec.

ДАЛІ: Передати зміни Courses 1 для `make e2e` та тріажу відомих кандидатів; Docker/e2e в цьому середовищі не запускалися.

## Крок 5 — дозволена перевірка

У `frontend/` виконано успішно:

- `npm run typecheck` — passed.
- `npm run lint` — passed (`oxlint --deny-warnings`, architecture boundaries: 174 files).

Не виконано навмисно: `make e2e`, integration та повний ordered gate — середовище Codex не має Docker, а brief прямо делегує цей прогін Courses 1. Це **not-run**, не passed.

ДАЛІ: Додати En04/En10 на справжні HLS-байти через nginx.

## Крок 6 — уточнення media-bytes (2026-10-07)

Прочитано оновлений brief, working model, playback.spec/setup і media.setup. Попереднє твердження WIP про достатність integration для media-bytes/preview скасовано уточненням Courses 2: Go CanConsume не доводить nginx auth_request. Додаю один ізольований курс із двосекундним відео media.setup, seeded student, реальним upload/transcode і playback session. Перевірка того самого сегмента до/після revoke відокремлює membership denial від відсутнього файла; повторна видача playback session після ввімкнення preview доведе виняток published-preview. Курс архівуватиметься наявним afterAll; глобальні seeded курси не змінюються.

ДАЛІ: Запустити тільки frontend typecheck + lint для доповнення.

## Крок 7 — En04/En10 media-bytes scenario

Додано один сценарій у `frontend/e2e/enrollment.spec.ts`: upload двосекундного fixture через editor; poll ready саме виданого videoId; seeded student self-enroll; POST playback sessions; читання справжніх master/rendition playlists; GET існуючого `.m4s` через `/media/p/<playbackId>/…` → 200 і непорожнє тіло. Admin revoke → той самий GET не-preview сегмента → 403. Preview увімкнено реальним editor на published курсі; membership API підтверджує `revoked`; той самий старий URL знову → 200 (контроль проти TTL/takeover як причини попередньої відмови). Новий playback session для revoked preview-відвідувача → GET сегмента 200 і тіло ідентичне попередньому. Це перевіряє nginx-шлях без mock/route intercept або прямого виклику internal authorize. Після preview-save оновлюється token для afterAll archive. Використовується стандартний chromium 30/180 playback profile, не прискорений playback-проєкт.

## Крок 8 — перевірка доповнення (2026-10-07)

`cd frontend && npm run typecheck` — passed, exit 0. `npm run lint` — passed, exit 0, architecture boundaries 174 files. `make e2e` / Docker — **not-run**, відповідно до brief; runtime-результати 200/403 вище є очікуваннями нового spec, ще не виміряними результатами. Courses 1 запускає повний e2e і тріажить. У цьому доповненні змінено тільки `frontend/e2e/enrollment.spec.ts` та append до цього WIP.

ДАЛІ: Courses 1 — виконати `make e2e`, підтвердити En04/En10 через реальний Compose/nginx.
