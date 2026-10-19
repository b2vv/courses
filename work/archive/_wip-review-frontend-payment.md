# WIP — рев'ю фронтенду payment (feat/payment vs main)

ДАЛІ: беру повний diff фронтенду (frontend/src, крім generated і styles), складаю мапу змінених файлів, читаю ключові: access-gating, order polling, purchase, authoring.

ФАКТ: Trust model solid — `locked`/`isPreview`/`price` надходять із сервера (playback-schema), не ре-деривуються. LockedLesson рендерить панель замість плеєра → жодного playback/transcript/discussion запиту для locked.
ФАКТ: WayForPay redirect — submitCheckoutForm POSTить форму на checkout.action; action валідується zod-ом проти allowlist (secure.wayforpay.com https, або e2e host http). fields вставляються як hidden inputs через input.value (не innerHTML) → без XSS/open redirect. Server-provided form POST. OK.
ФАКТ: Idempotency — кожен start новий crypto.randomUUID(); коментар пояснює, що retry свідомо новий ключ; подвійний сабміт блокує кнопка (isDisabled isPending). OK.
ФАКТ: order schema .strict(), grant schema .strict(), перевірка userId/courseId/id відповідності. Контракт норм.
ПИТАННЯ/РИЗИК: OrderStatusView polling — refetchInterval повертає false лише коли data!==undefined && isSettled. При стійкій помилці (404 order_not_found / auth) data лишається undefined → повертає 2000 → нескінченний полінг кожні 2с. Треба перевірити QueryClient defaults (retry, чи refetchInterval триває в error-стані).

ДАЛІ: читаю QueryClient defaults (retry/refetchInterval on error), route-guards, admin.courses route, CourseGrantPanel, CoursePriceField, require-identity, transport unnamedStatusCode, price-schema, format-price.

ФАКТ: QueryClient (router.tsx:18) global retry:false, staleTime 60s. Отже на помилці — 1 запит на інтервал, без ретраїв.
ФАКТ: Підтверджено по джерелу query-core (queryObserver.js:300-313): _computeRefetchInterval викликає refetchInterval(query) без перевірки error-стану; setInterval фаєрить _executeFetch поки focused. При стійкій помилці status-запиту (order_not_found 404, authentication_required 401, invalid_request) data лишається undefined → refetchInterval повертає 2000 → нескінченний полінг кожні 2с поки вкладка у фокусі. Фікс: у purchase-queries.ts:56 додати `if (query.state.status === 'error') ...` зі стопом на термінальних кодах (лишити ретрай лише для dependency_unavailable/timeout). Confidence: high що полінг триває; medium що блокує мердж.
ФАКТ: isPreview round-trips через save — formValues (CourseEditorView:144 {...lesson}) і onSave(course, values) (567) несуть isPreview; CourseTreeEdit.modules→AuthoringLesson має isPreview. Регулярний save НЕ несе price (CourseFormValues без price) → ціна не зачіпається; setPrice окремо через unchangedEdit+price. Поділ коректний, previews не стираються.
ФАКТ: parseUahAmount — межі ок (до 13 цифр, safe integer, validCoursePrice гейт); "1890.5"/"1890.05"/"1 890,5" парсяться вірно.
ФАКТ: makePublishedCourse інваріант input.paid === (price!==null); price валідується. priceSchema .strict() min1. Контракт норм.
ФАКТ: admin grant/price — компоненти лише в admin-гілці за beforeLoad-гардом; сервер відмовляє grant_forbidden. Подвійний сабміт блокують pending-прапорці. onGrant ловить помилки, authentication_required → onAuthenticationRequired.
ФАКТ: __root.tsx / use-sign-out — без dangerouslySetInnerHTML/innerHTML/window.location/eval/необроблених redirect. OrderSearch slug валідується courseSlug.test, інакше дропається (test покриває ..%2Fadmin → /courses).
РІШЕННЯ: Єдина предметна знахідка — unbounded polling на термінальних помилках status-запиту (medium). Решта (security/контракт/idempotency/trust-model) чисто. Готую фінальний звіт.
