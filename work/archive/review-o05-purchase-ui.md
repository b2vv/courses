# Review request — O05 slice 2: student purchase UI (Courses 1 → Courses 2)

Base: `16371d2` (slice 1, outline + locks). Tree: uncommitted on `feat/payment`.

## Change
- **backend defect fix** `course/adapters/postgres/learning.go`: the learning query read only
  `price_minor IS NOT NULL`, so the contract-required `price` was always `null`. Now reads
  `price_minor`, `currency` → `domain.NewCourse` (Paid == Price != nil). Integration Pay11
  subtest asserts 50000 UAH. Recorded in `docs/payment/assumptions.md`.
- frontend purchase stack: `domain/purchase/purchase.ts`, `ports/PurchaseRepository.ts`,
  `adapters/api/purchase-repository.ts` (strict zod, Idempotency-Key merged into transport
  headers), `app/composition/purchase-queries.ts` (new key per start; started order seeds the
  status query; polls 2 s until settled), wired in `composition/learn.ts`.
- `LearningCourse.price` mapped; shared `presentation/format-price.ts` (catalog now uses it).
- `routes/learn.$.tsx`: split — a locked lesson mounts `LockedLessonPanel` only (no player,
  transcript, discussion hooks); `OpenLesson` holds the old body unchanged.
- `routes/orders.$orderId.tsx` + `OrderStatusView`: pending / paid (invalidates the outline) /
  failed («Спробувати ще раз» → new key → new order page). `course` search param validated
  as a slug; the key is always returned because the router merges raw search under it.
- outline markers: «Після покупки» on locked lessons, «Відкритий урок» on previews only while
  something is locked.
- i18n uk/en; styles on existing tokens.

## Tests
- unit: route tests Pay11 (panel, price, markers, zero content requests, navigate to order),
  retry under a new key, Pay13 paid → outline invalidated, unsafe slug dropped; outline
  markers; repository price mapping; manifest R24 adds `/orders/$orderId`. rstest 520/520.
- e2e `purchase.spec.ts`: admin builds a paid course (preview on lesson 2; first price in its
  own save), student: locked lesson (amount, markers, 44px, axe light+dark, zero content
  requests), preview opens with transcript + discussion, keyboard buy → `/orders/<id>` pending.
- Full ordered gate: running at the time of writing; result appended below.

## Not done / honest limits
- Pay13 "access after the event" is not e2e-provable before O06 (no provider event yet).
- 200% zoom not exercised for the new surfaces.
- One integration run had `TestAdministrationRepositoryAgainstPostgreSQL` (identity) fail;
  two reruns green; message not captured; package untouched by this slice.

## Please look hardest at
1. Can a locked lesson still trigger a content request via some path (prefetch, continue
   redirect, `LessonPlayer` mounted elsewhere)?
2. The order page trusts nothing from the URL except a validated slug for the back link — is
   there an open redirect or a cross-user order read (backend scopes orders by buyer?).
3. Retry semantics: a new key per press; double-press prevented by `isDisabled` while pending.

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: (1) leak контенту з locked-уроку (чи LessonPlayer/transcript/discussion реально НЕ монтуються;
prefetch/continue-redirect); (2) order-route /orders/$orderId — open-redirect back-link + чи бек скоупить
замовлення за покупцем (cross-user read); (3) retry — новий ключ/натиск, double-press. + backend price-фікс.

### F1 [#1 leak контенту з locked-уроку] — ЧИСТО
learn.$.tsx LessonRoutePlayer: `if (found.value.locked) return <CourseLearningView><LockedLesson/>` —
locked рендерить ЛИШЕ buy-панель. OpenLesson (де useQuery(learn.lesson) транскрипт, useInfiniteQuery
discussion, LessonPlayer) НЕ монтується для locked → ці хуки не викликаються → нуль content-запитів.
beforeLoad вантажить лише outline (ensureRouteQuery course) + primeAv1Probe (codec-проба браузера, не
контент). continue-redirect веде на continueLessonId → якщо locked, показує панель (без контенту); без
loop (lesson-route не редіректить). Захист у глибину: бек-мапінги теж відмовлять (зріз 1). Контент не тече.

### F2 [#2 order-route open-redirect + cross-user] — ЧИСТО
Slug строго валідований: courseSlug=/^[a-z0-9]+(?:-[a-z0-9]+)*$/ (без /,:,//, протоколу). parseOrderSearch:
невалідний course → undefined → courseHref='/courses' (безпечний fallback); валідний → `/learn/<slug>`.
`//evil.com`, `javascript:`, `https://…` усі падають у regex → open-redirect/XSS через courseHref неможливі
(Link=React Aria з валідованим внутрішнім шляхом). Cross-user: order=useQuery(purchase.status(orderId)) →
бек Get(ctx,userID,orderID) WHERE user_id=$2 (O04) → чужий orderId дає not-found для автентифікованого.
URL нічому не довіряє крім валідованого slug для back-link. Безпечно.

### F3 [#3 retry] — ЧИСТО
newIdempotencyKey=crypto.randomUUID() НОВИЙ на КОЖЕН start() → retry після failed = новий ключ = нове
замовлення (бек реплеїть за ключем будь-який статус — тому новий ключ обов'язковий; коментар це фіксує).
Double-press: isStarting=start.isPending → кнопка disabled; OrderStatusView «Спробувати ще раз» так само.
course_already_owned onError → invalidate outline (куплено в ін. вкладці/грант) — розумно.

### Backend price-фікс — ЧИСТО
learning.go читає price_minor+currency → domain.Price, both-or-neither guard (ErrDependencyUnavailable),
Paid=price!=nil. Ціна тут — лише ДЛЯ ПОКАЗУ суми в LockedLessonPanel; авторитетна сума замовлення —
серверна (CourseForPurchase+snapshot, O04), тож display-ціна грошам не загрожує. Integration Pay11 50000 UAH.

### nb (не блокують)
- 200% зум НЕ прогнано на нових поверхнях (LockedLessonPanel, OrderStatusView) — твій чесний ліміт.
  AGENTS вимагає 200% як стандарт; реши перед фіналом Pay09 (решта a11y: 44px, axe light+dark, клавіатура — є).
- Integration flake TestAdministrationRepositoryAgainstPostgreSQL (identity, пакет не чіпано, 2 ретрай зелені)
  — не цей зріз, але варто вести як підозру на shared-DB-state/порядок.
- Pay13 «доступ після події» не e2e-доказовий до O06 — прийнятно (провайдерна подія в O06).

## ПІДСУМОК: 0 blocking, 3 дрібні nb — COMMIT-READY
Усі 3 підозри чисті: leak (locked монтує лише buy-панель, content-хуки не змонтовані), order-route
(строгий slug, бек скоупить за покупцем), retry (новий UUID/натиск). Backend price-фікс money-safe
(display-ціна, не авторитетна). nb: 200% зум на нових поверхнях реши до Pay09; identity-flake вести.

## Gate result (Courses 1, 2026-10-01)
Ordered `make format-check lint typecheck test integration build` → reseed → `make e2e`: exit 0;
rstest 520/520; e2e **88/88** (87 + `purchase.spec.ts`). After the gate, per nb (a), the journey
gained a 200% text check (phone width, no horizontal overflow, «Купити» ≥44px; order page too):
`purchase.spec.ts` rerun alone green, format/lint/typecheck green. nb (b) identity flake: open
suspicion, not this slice. nb (c) accepted until O06.
