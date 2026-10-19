# Brief — L7b: instructor roster + block/unblock

РОБОЧИЙ ФАЙЛ: `work/_wip-L7b-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. Стандартні правила — `work/codex-working-model.md`:
name the denial; НЕ чіпати git/tasks.md/docs report/generated-руками; missing contract → stop;
inaccuracy that would cause a bug → `## Discussion`.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. Бекенд (L5–L6c) і L7a (learn-фронт + enrollment data layer)
закомічені. L7a створив фічу **enrollment** (`domain/enrollment`, `ports/EnrollmentRepository`,
`adapters/api/enrollment-repository`, `app/composition/enrollment-queries`, зареєстровану в route-context
у `app/router.tsx`). L7b **розширює цю ж фічу** інструкторським ростером — ростер/блок — це enrollment,
НЕ discussion-moderation (teach).

Generated-клієнт (вже є): `getCourseRoster(courseId, params?, options?)` (keyset: `cursor`, `limit` 1..100 →
`RosterPage { entries: RosterEntry[]; nextCursor: string|null }`, `RosterEntry { userId, displayName, state:'active'|'blocked', enrolledAt, progressPercent }`),
`blockCourseParticipant(courseId, userId, options?)`, `unblockCourseParticipant(courseId, userId, options?)`.
(revoke — АДМІН, це L7c; тут НЕ додавай.)

## Data layer — розшир фічу enrollment (дзеркаль teach keyset+mutation)
- `ports/EnrollmentRepository.ts`: додай `roster(courseId, cursor?, signal?): Result<RosterPage>`,
  `block(courseId, userId, signal?): Result<void>`, `unblock(courseId, userId, signal?): Result<void>`.
  Доменні типи `RosterPage`/`RosterEntry` у `domain/enrollment`. Валідуй тіла type-guard'ами в адаптері.
- `adapters/api/enrollment-repository.ts`: реалізуй через наявний transport (`send`/`namedError`), мапінг
  кодів у enrollment error-union. **ВИПРАВЛЕННЯ (помилка рев'ювера, 2026-10-07):** раніше тут було «/404» —
  його НЕМАЄ. Generated-клієнт для block/unblock декларує лише 400/401/403/409/500/503; бекенд НЕ повертає
  404 для цих операцій. Мапь рівно наявні коди enrollment-union: `403 enrollment_forbidden`,
  `409 membership_conflict` (решту 400/5xx — наявною generic/`dependency_unavailable`-сіткою transport, як у L7a).
  Нових кодів/404 НЕ додавай, OpenAPI не чіпай.
- `app/composition/enrollment-queries.ts`: додай `roster(courseId)` як **infiniteQueryOptions** (дзеркаль
  teach `moderationQueue`: `initialPageParam: undefined`, `getNextPageParam: p=>p.nextCursor ?? undefined`,
  queryKey `['enrollment','roster',courseId]`), і `block(client)`/`unblock(client)` як mutationOptions,
  `onSuccess` → `client.invalidateQueries({queryKey:['enrollment','roster',courseId]})`. Інвалідація — лише
  ростер цього курсу (не чіпай learn-side membership/myMemberships keys).

## UI — інструкторський екран ростера
- Де: **розшир `routes/teach.courses.$courseId.tsx`** (наявний moderation-екран) вкладками `Tabs`
  (`presentation/ui/overlays`): «Модерація» (наявне) | «Учасники» (ростер). Якщо `Tabs` там незручно
  вписати — додай роздільну секцію з заголовком; дзеркаль наявну структуру екрана. НЕ ламай moderation.
- Контекст: ростер читай з `Route.useRouteContext().enrollment.roster(courseId)` через `useInfiniteQuery`
  (prefetch у loader через `prefetchInfiniteQuery`, як teach moderation у цьому ж файлі). Якщо teach-роут НЕ
  отримує `enrollment` з context — розшир надання context у `app/router.tsx` (enrollment уже створено там
  для learn-роутів; дай його і teach-роутам). Presentation не фетчить.
- Компонент `presentation/teach/CourseRoster.tsx` (дзеркаль `ReviewModerationRow`/moderation list):
  - рядок учасника: `displayName` (порожнє ім'я видаленого → анонімізуй як review-список, НЕ лий ""→сирцем),
    state-бейдж (active/blocked), `enrolledAt`, `ProgressBar` з `progressPercent` (**0 уроків → 0%**).
  - дія: active → кнопка «Заблокувати» (block), blocked → «Розблокувати» (unblock). `run()`-обгортка з
    `actionError` (дзеркаль moderation), `isPending` поки летить (не губити фокус). Успіх → ростер
    інвалідовано, рядок перемкнеться.
  - keyset: кнопка «Показати ще» gated на `hasNextPage`/`isFetchingNextPage` (як moderation).
  - **порожній ростер → порожній стан** (`EmptyState`, `ui/states`): «Поки немає учасників».
  - query error → `role="alert"` + retry (дзеркаль moderation).
- Навігація: якщо треба пункт/таб у teach-nav (`routes/__root.tsx` AdminShell teach-nav) — додай, дзеркаль
  наявний «Модерація».

## i18n (uk + en, ОБИДВА)
Namespace `enrollment.roster` (або поряд): `tab`/`title` («Учасники»), `block` («Заблокувати»),
`unblock` («Розблокувати»), `empty` («Поки немає учасників»), `state.active`/`state.blocked` (якщо
`enrollment.member`/`enrollment.blocked` із L7a не підходять за контекстом — переваж наявні),
`enrolledAt` label, і error-ключі для block/unblock (переваж наявний `enrollment.errors.*` з L7a).

## a11y
44px, видимий фокус, клавіатура; бейджі не лише кольором (текст); ProgressBar aria; alert/status ролі.

## Межі
НЕ чіпати бекенд, generated-руками, learn-side CTA/«Моє навчання» (L7a), скарги/адмін-чергу/revoke (L7c),
інші фічі. НЕ комітити. git/tasks/report — ні.

## Перевірка перед здачею (у WIP)
`cd frontend && npm run typecheck`; `npm run lint` (architecture!); `npm test` (Rstest) — якщо sandbox
не дає підняти порт, познач як denial (reviewing session прожене). Білд/e2e — reviewing session (e2e=L8).
