# Brief — L7c2: admin participant-report queue + resolve + revoke

РОБОЧИЙ ФАЙЛ: `work/_wip-L7c2-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`:
name the denial; НЕ чіпати git/tasks.md/docs report/generated-руками; missing contract → stop;
inaccuracy that would cause a bug → `## Discussion`.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. Бекенд (L5–L6c) і фронт L7a/L7b/L7c1 закомічені. Це остання
UI-частина L7. Розширюємо фічу **enrollment**. Адмін-екрани рендеряться в `AdminShell` (змонтовано в
`routes/__root.tsx` `Layout`), nav-items будуються там; admin-роути `ssr:false`. Шаблон admin-екрана зі
списком+діями — `presentation/admin/UserListView.tsx` (АЛЕ він OFFSET-пагінований; наша черга KEYSET —
бери infinite-query патерн з roster/moderation, а chrome Table/EmptyState/action-row — з UserListView).

**ВАЖЛИВО (урок L7b):** будь-яка рядкова дія — **per-row pending** (keyed on acted id + `isPending`,
НЕ list-wide `isDisabled`), і load-more → `isPending` + фокус на перший дозавантажений рядок (дзеркаль
`CourseRoster.tsx` L7b). **КОДИ помилок — рівно те, що декларує generated для кожної fn; нічого не вигадуй**
(минулі дві зупинки були про 404 — не повторюй у жоден бік: звір generated).

Generated-клієнт (вже є):
- `listParticipantReports({ cursor }?, options?)` → `ParticipantReportQueuePage { items: ParticipantReportQueueItem[] (max 20); nextCursor: string|null }`;
  `ParticipantReportQueueItem { id, courseId, reporterId, targetId, kind, reason, state, createdAt, reporterName|null, targetName|null, courseTitle }` (БЕЗ email).
- `resolveParticipantReport(reportId, options?)` → 204.
- `revokeCourseParticipant(courseId, userId, options?)` → (revoke цільового учасника).

## Data layer — розшир enrollment
- `domain/enrollment`: типи `ParticipantReportQueueItem`/`ParticipantReportQueuePage` (+ kind enum on_lecturer|on_student,
  state open|resolved). Додай до `EnrollmentErrorCode` будь-які коди, що їх декларує generated для list/resolve/revoke
  і яких ще нема (звір; напр. revoke 403/409 вже покриті enrollment_forbidden/membership_conflict).
- `ports/EnrollmentRepository.ts`: `participantReports(cursor?, signal?): Result<ParticipantReportQueuePage>`,
  `resolveReport(reportId, signal?): Result<void>`, `revoke(courseId, userId, signal?): Result<void>`.
- `adapters/api/enrollment-repository.ts`: transport + type-guards; мапінг рівно за generated-статусами.
- `app/composition/enrollment-queries.ts`:
  - `participantReports()` — **infiniteQueryOptions** (queryKey `['enrollment','participantReports']`,
    `getNextPageParam: p=>p.nextCursor ?? undefined`).
  - `resolveReport(client)` — mutationOptions, onSuccess → invalidate `['enrollment','participantReports']`.
  - `revoke(client)` — mutationOptions, onSuccess → invalidate `['enrollment','participantReports']` **і** roster
    цього курсу (`enrollmentRosterQueryKey(courseId)`); onError → 409 membership_conflict → invalidate ті самі ключі
    (дзеркаль block/unblock B-L7b.2).

## UI — адмін-екран черги скарг
- Новий роут `routes/admin.participant-reports.tsx` → `/admin/participant-reports` (flat dotted; під `admin.tsx`
  layout). Loader prefetch `prefetchInfiniteQuery(context.enrollment.participantReports())`. Компонент
  `presentation/admin/ParticipantReportQueue.tsx`:
  - keyset infinite-список; кожен елемент: `kind` (людською — «Скарга на викладача»/«на учня»),
    `reporterName ?? formerAuthor`, `targetName ?? formerAuthor`, `courseTitle`, `reason`, `createdAt`, `state`.
    **Email ніде.**
  - Дії (per-row pending):
    - «Вирішити» (resolve) → resolveReport(item.id). Idempotent; успіх → рядок зникне після invalidate.
    - «Відкликати доступ» (revoke) → **через confirm-Dialog** (revoke ріже доступ, лише адмін відновить) →
      revoke(item.courseId, item.targetId). (Дзеркаль confirm-патерн `presentation/admin/ResetPasswordDialog.tsx`.)
  - keyset «Показати ще» (isPending + фокус першого нового рядка), EmptyState «Скарг немає», query-error retry, alert-switch.
- Nav: додай пункт `/admin/participant-reports` у admin-nav у `routes/__root.tsx` (лише admin-гілка navItems,
  поряд із `/admin`, `/admin/users`), i18n-лейбл.

## i18n (uk + en, ОБИДВА)
`enrollment.reports` (черга): `title`, `kind.on_lecturer`/`kind.on_student`, `reporter`, `target`, `course`,
`reason`, `resolve`, `revoke`, `revokeConfirm`, `empty` («Скарг немає»), `more`; error-ключі для resolve/revoke
(переваж наявний `enrollment.errors.*`). Nav-лейбл `common.navParticipantReports` (або в adminUsers-стилі).

## a11y
Per-row pending (не губити фокус); confirm-Dialog focus-trap; таблиця/список із підписами; 44px; alert/status ролі;
kind/state не лише кольором.

## Межі
НЕ чіпати бекенд/generated-руками; НЕ ламати L7a/L7b/L7c1; НЕ комітити. git/tasks/report — ні.

## Перевірка (у WIP)
`cd frontend && npm run typecheck`; `npm run lint`; `npm test` (якщо sandbox-порт блокує — mark not-run).
Білд/e2e — reviewing session (e2e=L8).
