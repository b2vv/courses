# Brief — L7d: surface enrollment comms-gate errors in the UI (named, not generic)

РОБОЧИЙ ФАЙЛ: `work/_wip-L7d-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`.

## Контекст / чому (Courses 2 verdict: BLOCKING для En05/En11)
Бекенд (L4b) віддає окремі коди для комунікації-гейту: `enrollment_required` (не член),
`enrollment_blocked` (blocked), поряд із `email_not_verified`. Сенс окремих кодів (plan §2/§4) — щоб UI
показав **зрозумілу відмову** (En11), тобто користувач бачить ЧОМУ. Але фронт їх не несе → blocked-юзер
бачить generic «course_forbidden», не «вас заблоковано». Треба провести коди до видимого повідомлення на
двох шляхах комунікації: **коментарі** (discussion) і **відгуки** (reviews). Backend-поведінка вже є; це
суто frontend-surface.

Гейтовані бекендом комунікаційні виклики (усі можуть дати 403 `enrollment_required`/`enrollment_blocked`):
publish-comment, report-comment, put-review, report-review.

## Шлях A — коментарі (discussion)
- `frontend/src/domain/discussion/discussion.ts`: додай до `DiscussionErrorCode` два коди
  `enrollment_required`, `enrollment_blocked`.
- `frontend/src/adapters/api/discussion-repository.ts`: додай обидва коди до `403` у `publishErrors`
  і в map для report-comment (знайди відповідну таблицю, напр. reportErrors). Нічого іншого не чіпай.
- Presentation: `DiscussionPanel.tsx` рендерить `t('learn.discussion.errors.${code}')` автоматично —
  **switch чіпати не треба**, лише додай i18n-ключі.
- i18n: `learn.discussion.errors.enrollment_required` / `.enrollment_blocked` (uk+en), людські:
  напр. uk «Щоб коментувати, запишіться на курс.» / «Вас заблоковано в цьому курсі — коментування недоступне.»

## Шлях B — відгуки (reviews, course-repository)
- `frontend/src/adapters/api/course-repository.ts`: review-гілка (`putReview` ~197, `reportReview` ~160)
  мапить помилки review_* (див. switch ~40-70). Додай обробку 403 `enrollment_required`/`enrollment_blocked`
  для put-review і report-review. Знайди тип error-code union відгуків (там, де оголошені review_* коди —
  domain/course чи inline) і додай туди ці два коди.
- Presentation: форма відгуку у `frontend/src/presentation/storefront/CourseView.tsx` (review submit/report
  error-display). Якщо рендерить по коду через i18n — лише додай ключі; якщо є switch — додай випадки.
- i18n: відповідні ключі (напр. `ratings.errors.enrollment_required`/`.enrollment_blocked` або та
  namespace, що її читає форма відгуку), uk+en.

## Тести (КРИТИЧНО — Courses 2: асертити ВИДИМЕ повідомлення, не network-код)
FE unit (Rstest) — достатньо, але асерти на **відрендереному тексті**:
- Коментар: `DiscussionPanel` з `error='enrollment_blocked'` → видно «Вас заблоковано…»; з
  `error='enrollment_required'` → видно «…запишіться на курс.» (дзеркаль наявні DiscussionPanel-тести).
- Відгук: форма відгуку у CourseView з цими кодами → видно відповідні тексти (blocked І non-member).
Обидва коди, обидва шляхи. Без цього gap тихо регресне (network-код-асерт e2e його не ловить).

## Межі
НЕ чіпати бекенд/generated/інші фічі; лише discussion+review frontend + i18n + тести. НЕ комітити.
git/tasks/report — ні.

## Перевірка (у WIP)
`cd frontend && npm run typecheck`; `npm run lint`; `npm test`. Якщо sandbox-порт блокує Rstest — mark not-run,
reviewing session прожене. Білд/e2e — reviewing session.
