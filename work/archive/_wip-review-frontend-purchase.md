
## Review session 2026-10-02 (frontend purchase/order/access-gating)

ДАЛІ: беру diff main..HEAD по перелічених purchase/order/learn/admin файлах; дивлюсь валідацію external-data, idempotency-key, polling, cache invalidation, locked/preview/owned логіку.

ФАКТ: price-schema (minor int>=1, UAH), purchase-repository (zod validation, checkout action host allowlist), format-price (minor/100). Idempotency key = crypto.randomUUID() per mutate call, свідомо (коментар). polling refetchInterval=false коли settled, інакше 2000 навіть коли data undefined (error) → можливий unbounded poll на 404. onPaid useEffect invalidate learningCourseQueryKey. locked lesson рендерить LockedLessonPanel замість плеєра.
ДАЛІ: playback domain (locked/isPreview/price на LearningLesson/Course), playback-schema, learning course schema parsing, QueryClient mutation retry defaults, тести.

ФАКТ: QueryClient (router.tsx:17-19) request-scoped per getRouter(), queries retry:false; mutations не мають retry-default → react-query default mutation retry = 0, тож newIdempotencyKey() per mutate() = 1 network call, без дублю ордерів від авто-ретраю. Idempotency design свідомий.
ФАКТ: locked/isPreview/price приходять із сервера (playback.ts коментар "decided by the server, never re-derived"), мапляться у playback-repository.ts:70,81-82. Клієнтського рішення locked немає → leak неможливий. LockedLessonPanel рендериться замість плеєра (learn.$.tsx) → жодного playback/transcript запиту для locked.
ФАКТ: no `any`/unsafe asserts; лише `as ReturnType<typeof mutationOptions<...>>` у purchase-queries.ts:70 — звуження generics, не маскування input.
ФАКТ: checkout action host-allowlist (secure.wayforpay.com https / e2e host http) → submitCheckoutForm POST безпечний.
РІШЕННЯ: немає blocking. F1 (low-med): status polling не зупиняється на persistent error (data===undefined → refetchInterval=2000) → нескінченний 2s-poll на order_not_found/invalid_response; тест не покриває цей кейс. F2 (nit): format-price minimumFractionDigits:0 → "1 890,5 ₴" губить trailing нуль для некруглих сум. F3 (nit): stale коментар playback-schema.ts "learning model does not map them" — price МАПИТЬСЯ.
