# Brief — L8fix: e2e triage fixes (spec bugs) + a11y contrast + L7d comms-error UI

РОБОЧИЙ ФАЙЛ: `work/_wip-L8fix-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`.
Можеш дробити власну роботу (specs → a11y → L7d). Docker/e2e ти НЕ ганяєш — Courses 1.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. Перший `make e2e` (Courses 1) дав 6 падінь. Тріаж:
4 spec-баги (не продукт), 1 реальний a11y-дефект, 1 інфраструктура (reseed — Courses 1 зробить).
Плюс лишається L7d (Courses 2 verdict: BLOCKING для En05/En11). Усе нижче — frontend.

## Частина 1 — spec-фікси (`frontend/e2e/enrollment.spec.ts`, `route-guards.spec.ts`)
Коренева помилка specs: `GET /api/v1/learn/courses/{slug}` — це ПУБЛІЧНИЙ outline (план N3: опис+аутлайн —
метадані, НЕ gated; повертає 200 з per-lesson `locked`). Контент gated НЕ там, а на playback/transcript/
**comments** (читання коментарів = CanConsume, L4a). Виправ:

- **En03** (`:242`): прибери `expect(learning.status()).toBe(403)` на outline. Замість: лишаються
  `expect(...course metadata/Програма курсу...).toBeVisible()`; додай доказ **закритого контенту**
  не-члену — `GET /api/v1/learn/lessons/${course.lessonID}/comments` (або `/transcript`) → **403**.
- **En04** (`:289`): рядок `const denied = GET /learn/courses/{slug}; expect 403` замінити на той самий
  content-ендпоінт (`…/lessons/${course.lessonID}/comments` або `/transcript`) → **403** після revoke.
  Решту (re-enroll «Записатися»→«Ви учасник», single-row roster) лишити.
- **En05** (`:257`): прибери `expect(getByRole('heading',{name:'Урок en05'})).toBeVisible()` — сторінка
  уроку НЕ рендерить назву уроку заголовком. Доведи, що blocked РЕТЕЙНИТЬ контент реальним сигналом:
  напр. `await expect(studentPage.getByRole('region',{name:'Обговорення'})).toBeVisible()` (регіон
  обговорення рендериться = не LockedLessonPanel). Лиши перевірку POST коментаря → 403 `enrollment_blocked`.
- **route-guards.spec.ts:145** (тест «public SPA transitions …», рядок ~186): асерт `/learn`→`/courses`
  застарів (L7a зняв редірект). Онови на нову поведінку: після навігації на `/learn` URL лишається
  `/learn` і видно «Моє навчання» (для signed-in), як у R36. `currentUserRequests`/revalidate-логіку
  (ядро тесту) не ламай — лише ціль навігації й видимий заголовок.

## Частина 2 — a11y contrast (реальний дефект)
En07 `wcagViolations` на `/admin/participant-reports` повернув `['color-contrast']` (teach-ростер passed,
тож це специфіка черги). Знайди низькоконтрастний елемент у `presentation/admin/ParticipantReportQueue.tsx`
(та/або його класи/токени в `app/styles/*`) — ймовірні підозрювані: текст стану/виду, `<time>`, вторинні
підписи. Доведи контраст ≥ 4.5:1 семантичними токенами (не inline-кольори). Якщо винен спільний admin-токен,
що проявляється лише тут — усе одно виправ так, щоб не зламати інші admin-екрани (вони axe-зелені).

## Частина 3 — L7d: named comms-error у UI (BLOCKING для En05/En11)
Бекенд (L4b) віддає `enrollment_required` (не член) / `enrollment_blocked` (blocked) для publish-comment,
report-comment, put-review, report-review. Фронт їх не несе → generic-повідомлення. Проведи до видимого тексту.
- **Коментарі**: `domain/discussion/discussion.ts` → додай `enrollment_required`,`enrollment_blocked` до
  `DiscussionErrorCode`; `adapters/api/discussion-repository.ts` → додай обидва у `403` для publish і report;
  `DiscussionPanel` рендерить `t('learn.discussion.errors.${code}')` сам → лише i18n-ключі (uk+en).
- **Відгуки**: `adapters/api/course-repository.ts` review-гілка (putReview/reportReview) → мапь 403
  `enrollment_required`/`enrollment_blocked`; додай коди в review error-union; форма відгуку в
  `presentation/storefront/CourseView.tsx` → показати повідомлення; i18n (uk+en).
- i18n тексти людські: uk «Щоб коментувати/залишити відгук, запишіться на курс.» (required) /
  «Вас заблоковано в цьому курсі — {дія} недоступна.» (blocked).
- **ТЕСТИ (Courses 2: асертити ВИДИМИЙ текст, не network-код)**: FE unit — DiscussionPanel з
  `error='enrollment_blocked'`→видно текст блокування; `='enrollment_required'`→видно текст про запис;
  форма відгуку з обома кодами→видно відповідні тексти. Обидва коди, обидва шляхи (коментар+відгук).

## Межі
НЕ чіпати бекенд/generated; інші фічі — лише там, де вказано. НЕ комітити. git/tasks/report — ні.

## Перевірка (у WIP)
`cd frontend && npm run typecheck`; `npm run lint`; `npm test`. e2e/Docker — Courses 1 (reseed + full make e2e).
