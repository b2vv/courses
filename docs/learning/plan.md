# Фаза `learning` — план (до GATE 2)

> ЯК. Spec: `docs/learning/spec.md` (GATE 1 2026-10-08). Автор: Courses 1, 2026-10-08. Код пише Codex
> вузькими прогонами (брифи `work/brief-codex-lr*.md`), рев'ю — Courses 2, гейт — ordered `make check`.

## 1. Де живе код (модулі)

Усе нове — у модулі **`orders`**: він уже володіє членством (`course_grants`), ростером, блоком/розблоком і
аудитом (`grant_audit`). Нового модуля, нового контракту в `internal/contracts` і нової міграції немає.

- **Огляд** — нові порт/запит/сервіс/HTTP поруч із `Roster` (`orders/ports`, `adapters/postgres`,
  `application/service.go`, `adapters/http`). Той самий SQL-підхід, що ростер: `course_grants` + `users` +
  `courses` + прогрес із `course_modules`/`course_lessons`/`lesson_progress` (читання чужих таблиць у SQL
  адаптера — наявний прецедент: ростер читає `lesson_progress`, `discussion` читає `course_grants`).
- **Блок з черги** — нові HTTP-маршрути в `orders`; адаптер `orders` знаходить автора й курс за
  ідентифікатором коментаря (`discussion_comments.author_id` → `lesson_id` → `course_lessons` →
  `course_modules.course_id`) або відгуку (`discussion_reviews.author_id`, `course_id`) — новий метод порту
  `AuthorOf(kind, id)`. Далі — **той самий** `transitionMembership` (одна правда, один аудит, Lo17).
  Модуль `discussion` не імпортується (ADR 013): лише читання його таблиць у SQL адаптера orders.

## 2. API (OpenAPI → Orval/oapi; generated лише генерацією)

| Метод | Шлях | Хто | Відповідь |
|---|---|---|---|
| GET | `/api/v1/teach/learners` | інструктор (свої курси) | `LearnerPage` |
| GET | `/api/v1/admin/learners` | адмін (усі курси) | `LearnerPage` |
| POST | `/api/v1/teach/comments/{commentId}/block-author` | власник курсу коментаря або адмін | `MembershipGrant` |
| POST | `/api/v1/teach/comments/{commentId}/unblock-author` | те саме | `MembershipGrant` |
| POST | `/api/v1/teach/reviews/{reviewId}/block-author` | власник курсу відгуку або адмін | `MembershipGrant` |
| POST | `/api/v1/teach/reviews/{reviewId}/unblock-author` | те саме | `MembershipGrant` |

Query огляду: `courseId` (uuid, необов'язковий), `state` (`active|blocked`), `q` (ім'я, 1–80 символів,
ILIKE з екрануванням `%`/`_`), `cursor`, `limit` (1–100, 20).

```
LearnerPage   { entries: LearnerRow[] (≤100), nextCursor: string|null }
LearnerRow    { userId: uuid, displayName: string ("" для видаленого), lastActivityAt: date-time|null,
                courses: LearnerCourse[] (≥1) }
LearnerCourse { courseId: uuid, slug, title, state: active|blocked, progressPercent: 0–100, enrolledAt }
```
Без `email` (G3). `ModerationQueueItem` отримує **`authorBlocked: boolean`** (стан автора на курсі, щоб
кнопка знала «Заблокувати»/«Розблокувати») — **не** ідентифікатор автора (Lo14).

Коди відмов (нові — у `ErrorCode` enum): `learner_course_forbidden` (фільтр чужим курсом, 403),
`membership_author_not_blockable` (ціль — адмін, 409), наявні
`membership_transition_invalid` (автор без членства, 409), `course_forbidden`/`comment_not_found`/
`review_not_found` (404 невідомий об'єкт). Точні наявні імена звіряє бриф із `ErrorCode` enum — не вигадує.

## 3. Огляд — запит

- Видимість: інструктор — `courses.instructor_id = actor`; адмін — усі курси. Членства `active|blocked`
  (G2). Рядок = користувач, що має ≥1 видиме членство після фільтрів.
- Сортування й keyset: `(lower(coalesce(display_name,'')), user_id)` ASC; курсор — base64 цієї пари
  (стабільний при вставці між сторінками — Lo06). Видалені акаунти (`deleted_at`) — `displayName = ""`,
  сортуються першими, анонімні в UI (як ростер).
- Курси рядка — агрегат (`jsonb_agg` або другий запит по сторінці userId — вибір у брифі за читабельністю;
  обмеження: ≤ кількості курсів інструктора/юзера). `progressPercent` — **та сама формула**, що ростер
  (винести в спільну SQL-функцію або спільний фрагмент, щоб Lo02 не міг розійтись).
- `lastActivityAt` = `max(lesson_progress.updated_at)` по уроках видимих курсів рядка (G4).
- Фільтр `courseId` не свого курсу → `learner_course_forbidden`; `state`/`q` звужують членства/людей.

## 4. Блок із черги

`AuthorOf(kind, id)` → `(userID, courseID)` або not-found. Сервіс: `BlockAuthor(actor, kind, id)` →
перевірка G6 — ціль за `users.role`: **адміна не блокує ніхто**; адмін блокує студентів та інструкторів;
власник курсу — студентів та інструкторів-учасників свого курсу. Чиста функція
`CanBlock(actorAdmin bool, targetRole string) bool` в `orders/domain` з юніт-таблицею (актор адмін/власник ×
ціль student/instructor/admin) → `transitionMembership(…,
"block")` з `reason = "moderation_queue:<kind>:<id>"` у `grant_audit` (стан і дія ті самі, що з ростера;
reason показує походження — Lo17). Перевірку G6 додати і в шлях ростера — одне правило для обох.
`authorBlocked` у черзі: `EXISTS` по `course_grants` (state=blocked) для автора елемента.

## 5. Frontend

- Домени/адаптери/Query: `domain/teach` (LearnerRow), `adapters/api` (валідація відповіді схемою), Query
  ключі; без fetch у presentation/routes.
- Маршрути: `/teach/learners` (новий пункт навігації «Учні»), `/admin/learners`. Фільтри — Router search
  params (`courseId`, `state`, `q`, курсор — «Показати ще» як у ростері/відгуках). Рядок: ім'я, остання
  активність, список курсів (назва → посилання на `/teach/courses/{id}` ростер; стан; прогрес-бар).
- Черга: кнопка «Заблокувати автора на курсі»/«Розблокувати» біля коментаря й відгуку; `isPending`, не
  `isDisabled` (пам'ять: isDisabled губить фокус); помилки з кодом — названий текст.
- Токени, клавіатура, фокус, ≥44px, світла/темна, 200% (Lo11), axe.

## 6. Зрізи (кожен — окремий Codex-прогін, реалізація і тести окремо, де великі)

| Зріз | Зміст | Рядки | Доказ |
|---|---|---|---|
| LR0 | G8: `Enroll`/`StartPurchase` відмовляють адміну (новий код у `ErrorCode`, напр. `enrollment_admin_not_learner`, 409 — звірити наявні); UI ховає запис/покупку для адміна; **Rt05 переписати на студента-покупця** (зараз адмін купує й рецензує); grep e2e на інші admin-як-учасник | Lo18, Lo19 | unit + contract + integration; повний `make e2e-local` (зміна злитої поведінки enrollment) |
| LR1 | OpenAPI (огляд + block-author + `authorBlocked`) + генерація | Lo07, Lo14 (контракт) | `make check-generated`, contract-тести |
| LR2 | Backend огляду: порт, SQL, сервіс, HTTP | Lo01–Lo09 | integration (Postgres) + contract |
| LR3 | Backend блоку з черги + G6 у ростері + `authorBlocked` | Lo12–Lo17 | integration + contract |
| LR4 | Frontend огляду (teach + admin) | Lo10, Lo11 | unit + e2e |
| LR5 | Frontend кнопок у черзі | Lo12, Lo13 | unit + e2e (`CanCommunicate`: заблокований не публікує) |
| LR6 | Звіт `docs/learning/report.md`, рядки по одному, обмеження | усі | ordered `make check` |

Після кожного зрізу: `npm run format` (Codex його не ганяє), targeted тести, рев'ю Courses 2. Після LR3
і LR5 — повний `make e2e-local` (спільний стан: членство). Фінальний гейт — ordered `make check`.

## 7. Ризики й рішення, що закривають альтернативи

- **Без нової міграції/індексу (2026-10-08, Courses 1).** Огляд читає ті самі таблиці, що ростер; індекс
  на `course_grants(user_id, course_id)` уже є (unique). Перевірка: `EXPLAIN` запиту огляду на сідах +
  ~200 фікстурних членств в integration-тесті; якщо seq scan по `lesson_progress` дає >50 мс — індекс
  `lesson_progress(user_id, updated_at)` окремим рішенням.
- **Агрегація в одному SQL vs два запити** — вибір у LR2 з виміром; критерій — читабельність і один
  джерельний рядок формули прогресу.
- **`authorBlocked` розширює контракт черги** — свідомо замість окремого запиту стану (менше раундтріпів);
  не розкриває ідентифікатор (Lo14 — асерт у contract-тесті, що полів `authorId`/`userId` немає).
