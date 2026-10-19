# Review — D-e: two apps (client / admin shells) + Codex client design

Reviewer: **Courses 2**. Author: Courses 1 (+ Codex for the client header/hero design).
Uncommitted working tree on `feat/payment`, base `1c90bfc`.

РОБОЧИЙ ФАЙЛ: `work/review-de-shell.md`
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append). Перший рядок — `ДАЛІ:`. Далі ФАКТ/РІШЕННЯ/ПИТАННЯ.

## Owner's intent (2026-09-30)
Client (student) and admin are two apps on one design system: client = light/dark, airy,
toward the mockup; admin/teach = own root layout with its own menu (left sidebar), NorthWind
theme, and a "back to client" link. On the client, admins see "В адмінку" → /admin, instructors
see "Кабінет викладача" → /teach.

## What is in scope (design only — owner's option A)
- **Shell split (Courses 1):** `__root.tsx` renders the client chrome only for the client
  surface; for `/admin*` and `/teach*` it renders just the Outlet inside
  `.app-shell[data-surface=admin]`. New `presentation/admin/AdminShell.tsx` (sidebar + top bar +
  `main#main`), used by `routes/admin.tsx` and `routes/teach.tsx`. New
  `app/composition/use-sign-out.ts` shared by both chromes. `IdentityHeader.tsx` replaces the
  two inline admin links with one role cross-nav. i18n keys `common.toClient/toAdmin/toTeach/
  adminArea/teachArea/navCourses/navModeration`.
- **Redirect hardening:** `require-identity.ts` never tracks `/login` as a protected location and
  never re-navigates from `/login`. Reason: the surface switch reconciles the whole subtree, so
  a guarded route can render once more before unmounting; without this, `returnTo` nested into
  itself (route-guards unit R35/G12 caught it: `/login?returnTo=/login?returnTo=/admin`).
- **Client design (Codex):** header split into brand / primary nav / actions; hero-art =
  course preview card + progress card (as in the mockup); full-bleed hero; course title on the
  cover; orange cover category (`--color-cover-orange-*`); `--radius-control` per surface.
- **D-b regression fix:** the storefront footer wordmark no longer reuses
  `aria-label="CourseForge — головна"`; two links with one accessible name broke
  `route-guards.spec.ts:167` (strict mode).
- Tests: `IdentityHeader.test.tsx` (G20/G31 → role cross-nav), `accessibility.spec.ts`
  (admin link is now "В адмінку").
- **Payment contract fix (separate commit; a real branch regression, not design).** Found by the
  first full e2e since the payment commits: 20 failures, all traced to one cause. Instrument:
  browser console on `/admin/courses/<id>` → `AuthoringQueryError: invalid_response`. The API
  returns `price` (required, nullable `{minor≥1, currency:'UAH'}`) and `lessons[].isPreview`
  (optional) since O03/O04, but `admin-course-schema.ts` (last touched in `8d633f9`) validates
  with `.strict()` and rejected every course detail → the editor never opened → P20/P21,
  transcript, asr, discussion and instructor e2e all failed downstream. Fix: the schema accepts
  both fields. Reads only — `courseSave`/`courseWrite` still omit them, and the backend treats an
  omitted `price` as unchanged (`optionalPriceDTO.Set=false`) and an omitted preview flag as
  stored (`preserveUnspecifiedPreviews`), so saving the editor cannot make a paid course free.
  Stale fixtures updated (`admin-course-schema.test.ts`, `CourseEditorView.test.tsx` MSW
  responses); new test pins the paid+preview shape and refuses `minor:0` / `USD` / missing price.
  Why unit tests stayed green: fixtures mirrored the pre-payment contract.

## Out of scope — parked, not mine
`git stash@{0}` = payment+learn WIP of a session that ended uncommitted: `backend/.../course/
application/authoring.go`, `backend/.../orders/application/service.go`(+test), `routes/learn.$.tsx`,
`routes/route-guards.test.tsx`, plus i18n `common.panels`. `stash@{1}` = full pre-cleanup snapshot.
`routeTree.gen.ts` generator churn reverted.

## Gates so far (Courses 1)
format ✓ · lint ✓ (+architecture 128) · typecheck ✓ · rstest 508/508 ✓ · full e2e — running.

## Please scrutinise
1. Redirect guard: can skipping `/login` ever hide a real session-end redirect?
2. `.app-shell` admin branch: skip-link → `#main` inside AdminShell; no double `main`; client
   `main#main` unchanged.
3. Role cross-nav: instructor gets only /teach, admin only /admin — matches the guards?
   (Both routes still admit instructor+admin via `canAccessAdministration`.)
4. Codex header: a11y (≥44px, focus, 200% zoom, contrast) in light and dark.
5. Anything in the diff that belongs to the parked payment/learn WIP.

## Findings (Courses 2 appends below)

## Update — after the full-e2e investigation (Courses 1, 2026-09-30)

ДАЛІ: рев'ю всього дифу (дизайн + контракт + a11y), e2e 87/87.

Full e2e went 20 → 11 → 4 → 1 → 0 (87/87, then 87/87 again after later CSS). Every step had a
named cause found with an instrument (browser console / network JSON / axe / e2e artefacts),
not a guess. What changed since the first version of this file:

- **Contract, learning side:** `playback-schema.ts` `learningCourseSchema` accepts required
  `price` (nullable) + `paid` (`LearningCourse` since O03/O04). Same `.strict()` failure as the
  admin schema → `LearnQueryError: invalid_response` → empty lesson page → transcript,
  discussion, instructor e2e failed. Audit of every schema carrying payment members: `Course`,
  `CourseDetail` already fine; `AdminCourseWrite` write-only; only `LearningCourse` + the two
  admin ones were stale. Fixtures updated; e2e mock in `accessibility.spec.ts` updated.
- **Admin root layout:** `AdminShell` now renders in `__root.tsx` (admin surface), not inside
  `/admin`, `/teach` (both `ssr:false`), so the signed-in chrome is in the server HTML — the
  route-guards SSR test requires no anonymous frame. `admin.tsx`/`teach.tsx` back to HEAD.
- **Admin surface only for admitted viewers:** `surface = admin` needs an admin path AND
  `canAccessAdministration(session.data)`. A refused student on `/admin` now gets not-found
  inside the client app (header, their name) instead of a bare page. G04/W unit tests assert it.
- **A11y fixes:** `--color-muted` #6b7280 → #5e646e (was 4.29:1 on the active transcript line
  tint, 4.49:1 on surface; now 5.3–6:1). Course titles on the cover sit on a solid band of the
  gradient end (`--cover-solid`): text over a gradient is axe-incomplete and the orange start
  was 2.8:1; band ≥6.5:1. Cover ring is a `::after` (as an element its box passed the viewport
  at 200% on phones). axe: 0 violations, 0 incomplete, light + mobile dark 200%.
- **Undefined token:** `--space-5` was used (Codex) but never defined → every `padding` using
  it was invalid → cover had no padding. Defined `--space-5: 1.25rem`. Caught on screenshots,
  not by e2e.
- **Price typography:** `.course-price` was `ui-mono` (SF Mono has no Cyrillic → "грн",
  "Безкоштовно" fell back to a serif). Body font + tabular figures; hryvnia sign via
  `currencyDisplay: 'narrowSymbol'` (`2 490 ₴`).
- **Catalog toward the mockup:** 4 columns (`minmax(15rem,1fr)`), card footer wraps price/action
  as whole units (no mid-word breaks), shared truthful `ValueStrip` (practice/transcripts/
  discussion) on home and catalog — the mockup's "mentor code review" / "lifetime access" are
  not offered, so not claimed. Admin `h1` dense-scale.
- **Tests (selection scoping, AGENTS rule):** route-guards e2e clicks "Каталог" within the
  primary navigation (footer also links it) and probes `header` for the anonymous frame
  (identity sits beside `<nav>`, not in it).
- **Tool:** `visual-compare` gained whole-page surfaces + `PP_SURFACES`. Comparison ran against
  the mockup's own data (`app/course.ts` from `courseforge-academy-sources.zip`) seeded as real
  published courses, then removed (`slug LIKE 'mock-%'` → 0 left).
- **Ratings:** not built. Owner's decisions recorded in `docs/ratings/spec.md` (R1–R5), after
  payment.

Please look hardest at: the surface rule (`canAccessAdministration` on `session.data` — any
state where an admin sees client chrome on `/admin`, e.g. during session refetch?), and the
contract fixes (reads only; saves omit `price`/`isPreview`).

---
## Findings — Courses 2 (claude-opus-4-8, 2026-09-30)

ДАЛІ: пріоритет (1) правило surface у __root.tsx — стан session.data під час loading/refetch
(чи адмін бачить клієнтський chrome на /admin, чи студент бачить admin-chrome); (2) контракт-фікси
l(admin-course-schema, playback-schema) — лише читання, save омітить price/isPreview, backend omit=unchanged;
(3) redirect-guard /login; (4) AdminShell skip-link/#main/no-double-main; (5) role cross-nav; (6) leak з parked WIP.
Спершу __root.tsx (surface) + дві схеми + save-шлях.

### F1 [#1 surface rule] — ЧИСТО
`surface = onAdminPath && session.data != null && canAccessAdministration(session.data) ? admin : client`.
SSR-гарантія тримається: root `beforeLoad: resolveRootIdentity` → `ensureQueryData(current())`
кладе сесію в кеш ДО рендера (SSR і client-nav), тож session.data присутній з першого пейнту →
admin бачить AdminShell у server HTML, без анонімного/client-флешу. `refetchOnMount:false` не скидає
в undefined (data гідратоване). Background-refetch: react-query тримає попереднє data → admin не
блимає в client. Fail-closed: якщо fetch сесії впав, resolveRootIdentity → 'unavailable', кеш не
виставлено → data undefined → surface=client (без admin-chrome без підтвердженого адміна). Студент
на /admin: canAccessAdministration=false → client-chrome + /admin guard дає not-found (навмисно).
Єдиний перехід admin→client chrome — коли сесія реально завершилась (data→null): коректно.

### F2 [#2 контракт reads-only] — ЧИСТО (money-safe by construction)
Схеми (admin-course + playback learningCourse) додають price(nullable)+isPreview/paid лише в
RESPONSE-парс (.strict() тепер приймає, не reject). SAVE-payload будується ЯВНИМ переліком полів:
courseWrite = {slug,title,summary,description,[instructorId]}; courseSave = course{slug,title,summary,
description,instructorId} + modules{id,title, lessons{id,title,description}} — БЕЗ price/isPreview,
хоча parsed CourseTree їх несе (explicit pick, не spread). Тобто save фізично не шле price/isPreview
→ не може обнулити ціну чи перемкнути preview. Backend omit=unchanged — другий шар. priceSchema:
minor≥1 UAH або null(free) — дзеркалить міграцію (CHECK >0, null=free). Money-safe підтверджено.
ДАЛІ: #3 redirect-guard /login, #4 AdminShell skip-link/#main, #5 cross-nav, #6 leak parked WIP.

### F3 [#3 redirect-guard] — ЧИСТО
Пропуск спрацьовує ЛИШЕ коли returnTo.startsWith('/login'): або (data≠null) не капчить /login як
protected, або (data=null) вже на /login → не ре-навігує. На будь-якому захищеному роуті returnTo≠/login
→ session-end редірект спрацьовує з останнім protectedReturnTo. Реального редіректу не ховає. Роут
лише один — login.tsx (нема /login-help), тож startsWith без false-positive.

### F4 [#4 AdminShell] — ЧИСТО
Єдиний `<main id="main" tabIndex={-1}>` в AdminShell; skip-link `href="#main"` у __root перед
AdminShell → працює. Admin-гілка __root повертається рано (клієнтський <main> не рендериться) →
подвійного main нема. Клієнтський main#main без змін. Sidebar `<nav aria-label>`, «до клієнта» link.

### F5 [#5 cross-nav] — ЧИСТО
IdentityHeader (клієнт): admin→/admin (toAdmin), instructor→/teach (toTeach) — дефолтний вхід за
роллю, як задумав власник. Крос-доступ instructor↔admin (обидва роути пускають canAccessAdministration)
— ПРЕ-ІСНУЮЧА guard-поведінка, D-e її не міняв. Instructor на /admin бачить navCourses без adminUsers
(canAdministerAccounts=admin only) — коректно.

### F6 [#6 leak + backend] — ЧИСТО
Диф frontend-only: жодного backend/, learn.$.tsx, authoring.go, orders/application у дереві — parked
WIP у stash@{0}/{1} не тече. Backend = HEAD → Go test/integration переганяти НЕ треба: бекенд, що
пройшов гейт O04, байт-ідентичний; контракт-фікс — фронт наздоганяє вже-відвантажені поля. e2e 87/87
ганяє реальний бекенд+фронт разом → узгодження контракту доведено наскрізно.

### Дрібні nb (не блокують)
- priceSchema дубльовано (admin-course-schema + inline у playback-schema, та сама форма). Спільна
  схема уникла б дрейфу (напр. якщо додасться EUR — оновлять одну, не іншу). DRY-нит.
- muted #5e646e: 5.96/5.56 (≥4.5✓); space-5 визначено (1.25rem); axe 0/0 (peer). a11y чисто.

## ПІДСУМОК: 0 blocking, 0 significant nb — COMMIT-READY
Усі 5 підозр peer-а + контракт-безпека + surface + backend-питання — ЧИСТО. Money-safe доведено
конструктивно (explicit field pick у courseWrite/courseSave). Surface SSR-гарантія тримається,
fail-closed. Parked WIP коректно збережено в стешах (не втрачено). 2 дрібні nb (DRY priceSchema —
опційно). Out-of-scope не чіпав.
