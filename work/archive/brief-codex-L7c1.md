# Brief — L7c1: report forms (учень→лектор, лектор/адмін→учень)

РОБОЧИЙ ФАЙЛ: `work/_wip-L7c1-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`:
name the denial; НЕ чіпати git/tasks.md/docs report/generated-руками; missing contract → stop;
inaccuracy that would cause a bug → `## Discussion`.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. Бекенд (L5–L6c) і фронт L7a/L7b закомічені. L7c розбито на
**L7c1 (форми скарг — цей брифінг)** і L7c2 (адмін-черга+resolve+revoke — окремо). Розширюємо наявну
фічу **enrollment** (`domain/enrollment`, `ports/EnrollmentRepository`, `adapters/api/enrollment-repository`,
`app/composition/enrollment-queries`). Dialog — з `presentation/ui/overlays` (`Dialog`, Aria, focus-trap,
isDismissable). Форми — дзеркаль `presentation/admin/CourseGrantPanel.tsx` (form noValidate + TextArea +
submit + `role="alert"`/`<output>` + локальний error-switch).

Generated-клієнт (вже є): `reportCourseLecturer(courseId, { reason }, options?)` (POST, 201),
`reportCourseParticipant(courseId, userId, { reason }, options?)` (POST, 201). Тип тіла — `ParticipantReportRequest { reason: string }`.
Error-коди (в enum уже є): `report_own_person`, `report_invalid`, `report_forbidden`, `report_already_open`,
плюс `email_not_verified`.
**ВИПРАВЛЕННЯ (помилка рев'ювера, 2026-10-07):** раніше тут було «404 для цих НЕ існує» — це НЕВІРНО
(я рефлекторно переніс правило з L7b). Report-ендпоінти МАЮТЬ 404: generated декларує
`reportCourseLecturerResponse404`/`reportCourseParticipantResponse404` (`NotFoundResponse`), бекенд
`writeParticipantReportError` мапить `ErrNotFound`→404 `course_not_found` (курс не знайдено). Тож:
додай `course_not_found` до `EnrollmentErrorCode`, мапь 404→`course_not_found` у адаптері, і додай
i18n-ключ (`report`/`errors`: «Курс не знайдено»). Мапь рівно статуси, що декларує generated для цих fn.

## Data layer — розшир enrollment
- `domain/enrollment/enrollment.ts`: додай до `EnrollmentErrorCode` коди `report_own_person`,
  `report_invalid`, `report_forbidden`, `report_already_open` (якщо ще нема). Тип `ParticipantReport`
  (id, courseId, reporterId, targetId, kind, reason, state, createdAt) — якщо потрібен для 201-відповіді;
  якщо форма лише показує успіх/помилку і не рендерить створений звіт, можеш не моделювати тіло
  детально (але 201 має пройти type-guard — мінімальний guard на об'єкт).
- `ports/EnrollmentRepository.ts`: `reportLecturer(courseId, reason, signal?): Result<void>` (або Result<ParticipantReport>),
  `reportStudent(courseId, userId, reason, signal?): Result<void>`.
- `adapters/api/enrollment-repository.ts`: реалізуй через transport (`send`/`namedError`); мапінг рівно на
  коди вище (403 report_forbidden, 409 report_already_open, 400 report_invalid/report_own_person,
  403 email_not_verified; звір реальні статуси у generated для цих двох fn і мапь що декларовано, без вигадок).
- `app/composition/enrollment-queries.ts`: `reportLecturer()` і `reportStudent()` як mutationOptions.
  onSuccess — нічого інвалідувати не треба (скарга не змінює видимий стан користувача); лише повернути успіх.

## UI
### A. «Поскаржитися на викладача» — сторінка курсу (En07)
- `presentation/storefront/CourseView.tsx`: ЛИШЕ для члена (activeMember — тобто state active|blocked;
  revoked/none — без кнопки). Кнопка «Поскаржитися на викладача» відкриває `Dialog` з `TextArea`
  (reason, 1..500, required) + submit. Викликає reportLecturer-мутацію (пробрось через props з route
  `routes/courses.$slug.tsx`, як enroll). Успіх → закрити діалог + `role="status"` «Скаргу надіслано».
  Помилки → `role="alert"` + локальний switch коду → i18n (`report_already_open` → «Скарга вже на розгляді»,
  `report_own_person` → «Не можна поскаржитися на себе», тощо). isPending поки летить.
- route `courses.$slug.tsx`: додай `useMutation(context.enrollment.reportLecturer())`, пробрось у CourseView.

### B. «Поскаржитися» на учня — рядок ростера (En07, лектор/адмін→учень)
- `presentation/teach/CourseRoster.tsx`: додай per-row дію «Поскаржитися» → `Dialog` з `TextArea` →
  reportStudent(courseId, entry.userId, reason). Успіх → status; помилки → alert-switch (ті самі коди).
  Authz на бекенді (owner∨admin); фронт лише показує кнопку (вона в інструкторському/адмін-екрані).
- Проброс мутації з `routes/teach.courses.$courseId.tsx` (useMutation(enrollment.reportStudent())).

## i18n (uk + en, ОБИДВА)
`enrollment.report` namespace: `lecturer` («Поскаржитися на викладача»), `student` («Поскаржитися»),
`reasonLabel`, `submit`, `sent` («Скаргу надіслано»), `cancel`, і error-ключі
`report_already_open`/`report_own_person`/`report_invalid`/`report_forbidden`/`email_not_verified`
(переваж наявний `enrollment.errors.*`, якщо підходить).

## a11y
Dialog: focus-trap, Esc, повернення фокусу (Aria `Dialog` дає); TextArea label; submit 44px; status/alert ролі;
кнопка «Поскаржитися» видимий фокус.

## Межі
НЕ чіпати бекенд/generated-руками; НЕ робити адмін-чергу/resolve/revoke (L7c2); НЕ ламати roster block/unblock
(L7b) чи CTA (L7a). НЕ комітити. git/tasks/report — ні.

## Перевірка (у WIP)
`cd frontend && npm run typecheck`; `npm run lint`; `npm test` (якщо sandbox-порт блокує — mark not-run).
Білд/e2e — reviewing session (e2e=L8).
