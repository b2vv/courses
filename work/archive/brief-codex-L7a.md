# Brief — L7a: learn-side enrollment UI (enroll / «Ви учасник» / «Моє навчання» / стани + 409)

РОБОЧИЙ ФАЙЛ: `work/_wip-L7a-codex.md`
Append ПІСЛЯ КОЖНОГО КРОКУ; перший рядок `ДАЛІ:`. ФАКТ/РІШЕННЯ/ДАЛІ/ПИТАННЯ. Фінал = переказ файлу.
Стандартні правила — `work/codex-working-model.md`: name the denial; НЕ чіпати git/tasks.md/docs report/
generated-руками; missing contract → stop; inaccuracy that would cause a bug → `## Discussion`.
Великий зріз — **можна дробити власну роботу й робити resume**: спершу data-layer+сторінка курсу,
потім «Моє навчання», потім i18n+тести. Кожен крок — у WIP.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. Бекенд enrollment ПОВНІСТЮ готовий і заморожений
(L5/L6a/L6b/L6c). Фронт enrollment — **greenfield** (є лише generated-клієнт). Архітектура сувора,
її енфорсить `frontend/scripts/check-architecture.mjs` (через `npm run lint`): шари під `frontend/src/`:
`domain/` (0 імпортів), `ports/`+`application/` (inward-only), `adapters/` (єдине місце `fetch`; generated тут),
`app/composition/*-queries.ts` (wiring: TanStack Query option-factories), `app/providers/` (i18n, locales),
`presentation/` (тупі React-в'юхи), `routes/` (file-based TanStack Router). **`fetch(` у routes/presentation
ВАЛИТЬ білд.** Дотримуйся шарів точно, дзеркаль наявні фічі.

QueryClient defaults (`app/router.tsx`): `staleTime 60s, retry:false`. Мутації — **invalidate-on-success**,
без оптимізму (виняток лише purchase-seed). i18n: `app/providers/i18n.tsx` + `app/providers/locales/{uk,en}.ts`,
`fallbackLng:'uk'`; `const {t}=useTranslation(); t('ns.key')`; `common.myLearning:'Моє навчання'` вже є.

Generated-клієнт (вже є, `adapters/api/generated/`):
- `enrollInCourse(courseId, options)` → POST `/learn/courses/{id}/enroll` (201 membershipGrant | помилки).
- `getMyCourseMembership(courseId, options)` → `MembershipState { member:boolean; state: 'active'|'blocked'|'revoked'|null }`.
- `getMyCourseMemberships(options)` → `MyMembershipsPage { items: MembershipCard[] }`;
  `MembershipCard { courseId, slug, title, state:'active'|'blocked', progressPercent, enrolledAt }`.
Error enum (`errorResponseErrorCode.ts`) має: `enrollment_required`, `enrollment_blocked`,
`membership_conflict`, `course_not_enrollable`, `enrollment_forbidden`, `course_already_owned`,
`email_not_verified`.

## Частина A — data layer (дзеркаль `teach`/`catalog`)
1. **ports** `ports/EnrollmentRepository.ts`: інтерфейс із методами, що повертають доменний `Result<T>`
   (дзеркаль `ports/TeachRepository.ts`): `membership(courseId): Result<Membership>`,
   `enroll(courseId): Result<Membership>`, `myMemberships(): Result<MembershipCard[]>`. Доменні типи
   `Membership { member:boolean; state:'active'|'blocked'|'revoked'|null }` і `MembershipCard {...}`
   — у `domain/` (як catalog моделює `CourseDetail`), error-code union `EnrollmentErrorCode` з кодами вище.
2. **adapter** `adapters/api/enrollment-repository.ts`: через `createApiTransport()` (`adapters/api/transport.ts`),
   `send(op, init=>generatedFn(...,init), expectedStatus, signal)` + `namedError(outcome, allowed)` мапить
   `{status, error.code}` у `EnrollmentErrorCode` (дзеркаль `teach-repository.ts` ~94-134; 5xx/429→`dependency_unavailable`).
   Валідуй тіла локальними type-guard'ами перед `{ok:true,value}`.
3. **composition** `app/composition/enrollment-queries.ts`: `createEnrollmentQueries(repo)` →
   `membership(courseId)` queryOptions (key `['enrollment','membership',courseId]`),
   `myMemberships()` queryOptions (key `['enrollment','myMemberships']`),
   `enroll()` mutationOptions: `mutationFn: courseId=>repo.enroll(courseId).then(unwrap)`,
   `onSuccess: (_,courseId)=>{ client.invalidateQueries({queryKey:['enrollment','membership',courseId]}); client.invalidateQueries({queryKey:['enrollment','myMemberships']}); client.invalidateQueries({queryKey:['course']}); }`.
   `unwrap()` кидає `EnrollmentQueryError{code}` на `!ok` (дзеркаль існуючих).
4. **route context**: зареєструй bundle (напр. `enrollment`) у `app/router.tsx` поряд із `learn`/`catalog`
   (створи factory, передай `queryClient`), щоб routes діставали його з context (як вони дістають `catalog`).

## Частина B — сторінка курсу: «Записатися» / «Ви учасник» (En01, En03, En05)
Файли: `presentation/storefront/CourseView.tsx` (`<aside className="availability">`), route `routes/courses.$slug.tsx`.
- У route: додай `useQuery(context.enrollment.membership(courseId), { enabled: signedIn })` (дзеркаль
  gated `catalog.ownReview`), і `useMutation(context.enrollment.enroll())`. Передай membership+enroll
  як props у `CourseView` (presentation не фетчить). courseId бери так само, як `ownReview` (детально
  глянь наявний виклик — чи це course.id).
- Логіка CTA в `<aside className="availability">`:
  - member && state==='active' → «Ви учасник» + наявний `<Link href=/learn/${slug}>` «Почати навчання».
  - state==='blocked' → «Ви учасник», лінк «Почати навчання» ЛИШАЄТЬСЯ (контент доступний blocked), +
    `role="status"` нотатка `t('enrollment.blockedNote')` («Доступ до контенту відкрито, участь в
    обговореннях обмежено»).
  - (не member) && безкоштовний курс (`course.price`==0/безкоштовний — звір, як CourseDetail несе ціну) →
    кнопка «Записатися» (`enrollment.enroll`), що кличе enroll-мутацію.
  - (не member) && платний → **НЕ чіпай** наявну поведінку (лінк веде в плеєр, де покупка). Опис+аутлайн
    видно завжди (метадані).
  - state==='revoked' на безкоштовному → «Записатися» знову дозволено (re-enroll); на платному → як не-член платний.
- Кнопка: `isPending` (НЕ `isDisabled`) поки мутація летить, щоб не губити фокус (див. коментарі в
  `CourseView.tsx`/`LockedLessonPanel.tsx`). Успіх → membership інвалідовано, CTA сам перемкнеться.
- **409 `membership_conflict`**: у error-гілці enroll (код `membership_conflict`) — показати retry-able
  `role="alert"` `t('enrollment.errors.membership_conflict')` і `invalidateQueries(['enrollment','membership',courseId])`
  (перечитати стан; користувач уже член в іншій вкладці). Решта кодів → `enrollErrorKey(code)` switch → `t(...)`
  (дзеркаль `purchaseErrorKey`/`errorKey` у LockedLessonPanel/CourseGrantPanel). `email_not_verified` →
  повідомлення «підтвердіть email».

## Частина C — «Моє навчання» список (En01)
Файли: `routes/learn.tsx` (зараз `beforeLoad` РЕДІРЕКТИТЬ `/learn`→`/courses` і рендерить `LearnEmptyState`).
- **Прибери редірект /learn index** (ЛИШЕ index; splat-плеєр `routes/learn.$.tsx` і requireIdentity не чіпай).
  На `/learn` index рендер списку карток із `useQuery(context.enrollment.myMemberships())` (prefetch у loader
  через `ensureRouteQuery`). Presentation-компонент `presentation/learn/MyLearningList.tsx`:
  сітка карток {title, progressPercent (ProgressBar з `ui/data`), state-бейдж, `<Link to=/learn/${slug}>}`,
  blocked → бейдж/нотатка. Порожньо → `EmptyState` (`ui/states`) з CTA «Переглянути курси» → `/courses`.
- Якщо треба назва/обкладинка — усе вже в `MembershipCard`.

## Частина D — i18n (uk + en, ОБИДВА файли)
Новий namespace `enrollment` у `app/providers/locales/uk.ts` і `en.ts`:
`enroll`, `member` («Ви учасник»), `startLearning` (якщо ще нема — звір `course.startLearning`), `blockedNote`,
`myLearning.title`, `myLearning.empty`, `myLearning.browse`, `progress` (напр. «{{percent}}% пройдено»),
і `errors: { membership_conflict, enrollment_required, enrollment_blocked, course_not_enrollable,
enrollment_forbidden, course_already_owned, email_not_verified, generic }`. Тексти укр/англ, людські.

## Частина E — a11y
Усі нові контроли: 44px (примітиви вже дають), видимий фокус, клавіатура; нотатки — `role="status"`/`aria-live`,
помилки — `role="alert"`. ProgressBar із `ui/data` (має aria). Світла/темна теми — лише токени/класи, без inline-кольорів.

## Тести
- Rstest на presentation: CTA-гілки (member/blocked/free-nonmember/revoked), enroll success→CTA flip (мок мутації),
  membership_conflict→alert+invalidate, MyLearningList (список/порожньо/blocked-бейдж). Дзеркаль наявні
  presentation-тести (шукай `*.test.tsx` біля CourseView/LockedLessonPanel).
- Repository/composition unit, якщо є патерн у teach/catalog (мапінг кодів).
- e2e `enrollment.spec.ts` — **НЕ тут**, це L8.

## Межі
НЕ чіпати бекенд, generated-руками, інші фічі (ростер=L7b, скарги/адмін=L7c). НЕ комітити. git/tasks/report — ні.

## Перевірка перед здачею (у WIP)
`cd frontend && npm run typecheck`; `npm run lint` (архітектура!); `npm test` (Rstest) або цільові;
`git diff --check` (якщо standing-rule дозволяє, інакше познач не-run). Білд/e2e — reviewing session.
