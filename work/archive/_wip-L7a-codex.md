ДАЛІ: Перевірено brief, working model і наявні frontend seams; реалізую data layer enrollment, потім CTA сторінки курсу.

## ФАКТ

- Робоче дерево вже має сторонні незакомічені файли; їх не чіпаю.
- Контракт generated містить усі три потрібні операції та схеми `MembershipGrant`, `MembershipState`, `MyMembershipsPage`.
- Безкоштовність у `CourseDetail` представлена як `price === null`; це перевірено в існуючому catalog view.

ДАЛІ: Data layer готовий; підключаю membership query і enroll-CTA до course route/view.

## РІШЕННЯ

- Додано enrollment domain/port/adapter/composition та router bundle. Адаптер суворо валідовує кожне тіло API; успішний `MembershipGrant` з `state: active` мапиться у доменний `Membership`.
- Мутація інвалідує конкретний membership, список навчання й cache курсу після успіху; оптимістичного стану немає.

ДАЛІ: Course CTA підключено; реалізую `/learn` index і його presentation list.

## ФАКТ

- `membership_conflict` у course route інвалідує лише відповідний membership query, тому повторне читання змінює CTA за фактичним станом з іншої вкладки.
- CTA не вважає платний курс безкоштовним: лише `price === null` отримує enroll-кнопку.

ДАЛІ: `/learn` тепер має query-backed список; додаю переклади й тести, тоді запускаю дозволені frontend checks.

## РІШЕННЯ

- Прибрано redirect лише для learn index. Splat плеєр і `requireIdentity` збережені; loader prefetch-ить memberships тільки на `/learn`.
- Новий presentation list отримує лише доменні картки через props, відображає aria-enabled ProgressBar, blocked-стан і порожній стан з посиланням на каталог.

ДАЛІ: Переклади додано в обидві локалі; пишу targeted presentation та adapter/composition tests.

## ФАКТ

- Усі нові тексти знаходяться у новому `enrollment` namespace; fallback лишається українським за чинним i18n provider.

ДАЛІ: Targeted presentation tests додані; запускаю typecheck, lint і Rstest. Build/e2e лишаю reviewing session за brief.

## РІШЕННЯ

- Course presentation tests покривають free CTA, успішний перехід до active CTA через rerender, blocked note, revoked re-enroll та conflict alert.
- MyLearningList tests покривають картку/progress/link, blocked note і порожній стан.

ДАЛІ: Typecheck і lint пройшли; Rstest заблокований sandbox-обмеженням `listen EPERM ::1:3000`. Додаю перевірку syntax/typecheck після test-case conflict invalidation і передаю результат.

## ФАКТ

- `npm run typecheck` — пройшов.
- `npm run lint` — пройшов; architecture checker підтвердив 170 файлів, без `fetch` поза adapters.
- `npm test` — НЕ запущено до тестів: Rstest не може відкрити `::1:3000` у sandbox (`EPERM`). Це denial, не успішний тест.

## РІШЕННЯ

- Інвалідацію membership після `membership_conflict` перенесено до enrollment mutation options, де є і error, і `courseId`; route лише підключає цю mutation. Додано unit test exact query-key invalidation.

ДАЛІ: Передати на reviewing session: виконати Rstest поза sandbox, потім build/e2e у належному gate. Не запускати тут.

## ФАКТ

- Після фінальних змін повторно пройшли `cd frontend && npm run typecheck` і `npm run lint` (architecture: 170 files).
- `npm test` лишається не пройденим через sandbox denial `listen EPERM: operation not permitted ::1:3000`.
- Build, e2e та `git diff --check` не запускалися: перші два прямо залишені reviewing session, останнє не запускалося через інструкцію не чіпати git.

## ПИТАННЯ

- Немає. Контракти, потрібні для L7a, наявні.
