# Codex brief — LR4: frontend of the learner overview

РОБОЧИЙ ФАЙЛ: `work/_wip-lr4-codex.md`
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого інструменту:
рядок `ДАЛІ:` з тим, що збираєшся робити. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ файлу.

## Scope (що і що НЕ робимо)

Фаза `learning`, зріз **LR4** — лише **фронтенд огляду учнів**. Бекенд (порти/SQL/HTTP/OpenAPI)
уже зроблено й схвалено в LR1–LR3; генеровані моделі вже на місці
(`frontend/src/adapters/api/generated/models/learner*.ts`, `listTeachLearnersParams.ts`,
`listAdminLearnersParams.ts`). **Нічого в бекенді, OpenAPI чи `generated/` не чіпаємо.**
Кнопки блокування в черзі модерації — це LR5, НЕ цей зріз.

Закриває рядки приймання (`docs/learning/spec.md`):
- **Lo10** — з рядка огляду перехід у ростер курсу (e2e).
- **Lo11** — UI огляду: клавіатура, видимий фокус, ≥44px, світла/темна, 200% тексту, axe (e2e).
- частково Lo07 на рівні фронту: домен НЕ має поля `email` взагалі; видалений акаунт
  (`displayName === ""`) показується анонімно.

## Де живе код (точні файли та прецеденти — читай їх перед кодом)

Огляд кладемо в домен **`teach`** (не `enrollment`), поряд із рештою робочого місця інструктора.
Прецедент, який повторюємо 1-в-1 за структурою:

- Домен: `frontend/src/domain/teach/teach.ts` (додати типи + коди помилок).
- Порт: `frontend/src/ports/TeachRepository.ts` (додати метод).
- Адаптер: `frontend/src/adapters/api/teach-repository.ts` (+ тест поряд) — **валідація
  відповіді схемою**. Еталон валідації schema-guards: `frontend/src/adapters/api/enrollment-repository.ts`
  (функції `record`/`uuid`/`date`/`card`). Повтори цей стиль у teach-repository власними guard-ами.
- Query: `frontend/src/app/composition/teach-queries.ts` (+ тест) — `infiniteQueryOptions`,
  `unwrap`/`TeachQueryError` уже є в цьому файлі.
- Presentation: новий `frontend/src/presentation/teach/LearnerOverview.tsx` (+ тест).
  Еталон списку з load-more/фокусом/isPending/прогрес-баром: `frontend/src/presentation/teach/CourseRoster.tsx`.
- Маршрути: `frontend/src/routes/teach.learners.tsx`, `frontend/src/routes/admin.learners.tsx`.
  Еталон guard-а `/teach`: `frontend/src/routes/teach.courses.$courseId.tsx` (+ `teach.tsx`).
  Еталон `/admin`-сторінки: `frontend/src/routes/admin.participant-reports.tsx`.
  Еталон `validateSearch` + `useSearch`/`useNavigate`: `frontend/src/routes/admin.users.index.tsx`.
- Навігація: `frontend/src/routes/__root.tsx` (масиви `navItems` для teach і admin, ~рядки 143–155).
- i18n: `frontend/src/app/providers/locales/uk.ts` і `en.ts` (обидва файли, однакові ключі).

## 1. Домен — `domain/teach/teach.ts`

Додай (імена й форма мусять збігатися з generated-моделлю, БЕЗ email):

```ts
export type LearnerCourseState = 'active' | 'blocked';
export type LearnerCourse = {
  readonly courseId: string;
  readonly slug: string;
  readonly title: string;
  readonly state: LearnerCourseState;
  readonly progressPercent: number; // 0..100 ціле
  readonly enrolledAt: string;
};
export type LearnerRow = {
  readonly userId: string;
  readonly displayName: string;        // "" для видаленого акаунта — показувати анонімно
  readonly lastActivityAt: string | null;
  readonly courses: readonly LearnerCourse[]; // ≥1
};
export type LearnerPage = {
  readonly entries: readonly LearnerRow[];
  readonly nextCursor: string | null;
};
export type LearnerScope = 'teach' | 'admin';
export type LearnerFilters = {
  readonly courseId?: string;
  readonly state?: LearnerCourseState;
  readonly q?: string;
};
```

У `TeachErrorCode` додай `'learner_course_forbidden'` (403, фільтр чужим курсом). `course_forbidden`,
`course_not_found` вже є — не дублюй. Звір точні рядки з наявним enum, нові не вигадуй понад цей.

## 2. Порт — `ports/TeachRepository.ts`

```ts
listLearners(
  scope: LearnerScope,
  filters: LearnerFilters,
  cursor?: string,
  signal?: AbortSignal,
): Promise<TeachResult<LearnerPage>>;
```

## 3. Адаптер — `adapters/api/teach-repository.ts`

- Виклич generated `listTeachLearners` (scope `teach`) або `listAdminLearners` (scope `admin`)
  з params `{ courseId, state, q, cursor, limit: 20 }` (не слати undefined-ключі — як у ростері).
- **Валідуй відповідь schema-guard-ами** (не довіряй generated-типам як runtime-правді): кожен
  `LearnerRow` — `uuid(userId)`, `displayName` рядок (|| ""), `lastActivityAt` null або date,
  `courses` непорожній масив валідних `LearnerCourse` (uuid, slug-pattern, title 1..200,
  state ∈ {active,blocked}, progressPercent ціле 0..100, enrolledAt date). Невалідна форма →
  `invalid_response`. Відображення кодів помилок транспорту — як у наявних методах teach-repository.
- **Поля `email` в мапінгу немає** (Lo07). Нічого з generated не копіюй наосліп — мапь явно в домен.

## 4. Query — `app/composition/teach-queries.ts`

Додай у `createTeachQueries`:

```ts
learners: (scope: LearnerScope, filters: LearnerFilters) =>
  infiniteQueryOptions({
    queryKey: teachLearnersQueryKey(scope, filters),
    initialPageParam: undefined as string | undefined,
    queryFn: ({ signal, pageParam }) => teach.listLearners(scope, filters, pageParam, signal).then(unwrap),
    getNextPageParam: (page) => page.nextCursor ?? undefined,
  }),
```

+ експортована `teachLearnersQueryKey(scope, filters)` → `['teach','learners', scope, filters] as const`.
Пагінація — **через `useInfiniteQuery` + «Показати ще»** (як ростер/відгуки), курсор НЕ в URL.

## 5. Маршрути

Два тонкі файли, спільний presentation-компонент.

- `routes/teach.learners.tsx` (під guard-ом `/teach`), `routes/admin.learners.tsx` (під `/admin`).
  `ssr: false`, guard/pending/error — дзеркаль сусідні сторінки відповідної поверхні.
- `validateSearch` парсить **courseId? (uuid), state? ('active'|'blocked'), q? (trim, 1..80, інакше
  відкинути)**. Зразок парсера/`useSearch`/`useNavigate` — `admin.users.index.tsx`. Курсор у search
  НЕ кладемо.
- Компонент читає search → `LearnerFilters`, викликає `useInfiniteQuery(teach.learners(scope, filters))`,
  передає все в `LearnerOverview`. Зміна фільтра (q/state/clear-course) → `navigate({ search })`.

## 6. Presentation — `presentation/teach/LearnerOverview.tsx`

Спільний для обох маршрутів; props: `entries`, `isLoading`, `error`, `hasNextPage`,
`isFetchingNextPage`, `filters`, `onRetry`, `onLoadMore`, `onFilterChange`. Структура рядка:
- заголовок = `displayName || t('ratings.formerAuthor')`;
- `lastActivityAt`: якщо null — `t('teach.learners.neverActive')`, інакше `<time dateTime=...>`
  локалізована дата (null-safe, як ростер робить із датами);
- список курсів: кожен — назва = **`<Link to="/teach/courses/$courseId" params={{courseId}}>`**
  (ростер курсу — ціль Lo10), стан (`t('enrollment.roster.state.*')` або власний ключ), `ProgressBar`
  (як у ростері, `value`=progressPercent, 0..100, лейбл `t('enrollment.progress',{percent})`).
- Фільтри: поле пошуку `q` (submit → search param), селект стану (всі/активні/заблоковані),
  якщо `filters.courseId` заданий — чип «курс: …» з кнопкою «очистити». Повного пікера курсів не
  роби (огляд глобальний; courseId приходить із URL). Це свідоме мінімальне рішення — познач коментарем.
- Load-more кнопка: **`isPending={isFetchingNextPage}`, НЕ `isDisabled`** (пам'ять: native disabled
  на фокусованій кнопці губить фокус у `<body>`, B-L7b.1). Фокус на перший довантажений рядок — як
  `CourseRoster` (useRef/useEffect, рядки 59–70).
- Стани loading / error+retry / empty — як ростер (`<output>`, `role="alert"`, `EmptyState`).
- a11y: секція з `aria-labelledby`, список з `aria-label`, усі контроли ≥44px, видимий фокус,
  працює в темній темі й на 200% зумі (Lo11).

## 7. Навігація — `__root.tsx`

- teach navItems: додати `{ to: '/teach/learners', label: t('common.navLearners') }`.
- admin navItems: додати `{ to: '/admin/learners', label: t('common.navLearners') }`.
- i18n: `common.navLearners` = «Учні» (uk) / «Learners» (en) + `teach.learners.{title,empty,more,
  searchLabel,searchPlaceholder,stateAll,neverActive,courseFilter,clearCourse}` у ОБОХ локалях.

## 8. Тести

**unit (Rstest, поряд із кодом):**
- `teach-repository.test.ts`: валідна сторінка → мапиться (без email у результаті); зіпсована
  форма (не-uuid userId / порожній courses / progressPercent 101) → `invalid_response`; транспортна
  помилка 403 → `learner_course_forbidden`.
- `teach-queries.test.tsx`: `teachLearnersQueryKey` і unwrap помилки → `TeachQueryError`.
- `LearnerOverview.test.tsx`: рендер рядків; `displayName===""` → formerAuthor; `lastActivityAt===null`
  → neverActive; порожній стан; error+retry; load-more `isPending` (кнопка лишається фокусованою);
  курс рендериться як посилання на `/teach/courses/{id}` (перевір href).

**e2e (Playwright, `frontend/e2e/`):** дзеркаль наявні teach/accessibility специ.
- **Lo10**: на `/teach/learners` клік по курсу в рядку → URL = `/teach/courses/{id}`, видно ростер.
- **Lo11**: axe на `/teach/learners` без порушень; tab-порядок проходить фільтри й перший рядок;
  видимий фокус; контрол ≥44px (виміряй boundingBox, не око); темна тема; 200% зум без втрати контенту.

## Порядок, гейти, розбиття на прогони

Пам'ять: великий зріз у Codex упирається в ліміт часу й тести випадають — **розбий на 2 прогони**:
1. **impl**: домен → порт → адаптер → query → маршрути → presentation → i18n → навігація.
2. **tests**: unit + e2e (resume за id першого прогону).

Після КОЖНОГО прогону: `npm run format` (Codex не ганяє format сам; інакше `make format-check`
завалить гейт на невідформатованому фронті). Далі targeted: `npm run typecheck`, `npm run lint`,
`npm test` для нових файлів. Повний `make e2e-local` НЕ тут — він за планом після LR5 (спільний
стан членства чіпає LR5, не LR4); але нові LR4-специ ганяй таргетно й доведи зеленими.

НЕ комітити — коміт і фінальний ordered-гейт робить Courses 1 після рев'ю Courses 2.
