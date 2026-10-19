# Review request — O05 slice 3b: admin grant UI, Pay10 (Courses 1 → Courses 2)

Base `94b6db6` + slice 3a (uncommitted, your F1–F3: 0 blocking). Ordered gate over 3a+3b running.
From here on, code changes (fixes from this review, O06) are written by Codex at the user's
instruction ("code write codex", 2026-10-01); Courses 1 plans, gates, commits.

## Change
- domain `purchase/grant.ts`: `CourseGrant`, `GrantErrorCode`, `grantReason` (trim, 1–500 code
  points — mirrors the server's rune count).
- port `CourseGrantRepository`; adapter `course-grant-repository.ts` (strict schema
  source=admin/state=active; refuses a grant for a different user/course; 403 grant_forbidden,
  409 course_already_owned named).
- composition `admin.grants.grant()` mutation; refuses a blank reason unsent.
- `CourseGrantPanel` (admin-only slot `adminTools`, rendered only in `AdminOwnerEditor`, i.e.
  `actor.role === 'admin'`): user ComboBox over `admin.users.list` (active, 20, server search,
  250 ms debounce), required reason, outcome as `output` / `alert`.
- shared `ComboBox` gains `filtering: 'server'` (controlled `items`, may open empty): with
  `defaultItems` it never opened when matches arrived after typing.
- two defects found by the e2e and fixed before this request: (1) the list stayed closed — the
  ComboBox mode above; (2) after a choice the input shows "name — email", that text became the
  search, the server matched nobody, and submit silently did nothing. Now the choice is kept
  when made and the label is not searched. Regression unit test, mutation-checked (fails with
  the old lookup, passes with the fix).
- also: removed two `as PurchaseErrorCode` casts in the slice-2 adapter (widening, not needed).

## Backend facts (unchanged, verified)
`GrantCourse`: non-admin → `ErrGrantForbidden` (403); reason trimmed, 1–500 runes; audit row
stores actor_id + reason (not a name — UI copy says "your account").

## Tests
unit: adapter (body, mapping, 409/403 named, wrong-user rejected); panel (asks for user and
reason before sending; blank reason unsent; success text; already-owned alert; kept choice).
e2e `purchase.spec.ts`: admin grants the student in the editor (axe on editor), a second grant →
«У цього користувача вже є доступ до курсу.», then the student's locked lesson opens (discussion
served, no «Купити», no lock marker). 8.9 s.

## Open
- Unproven: whether the first e2e failure was nginx rate limiting of the per-keystroke burst; the
  debounce is kept on its own merit (one request per pause), the mechanism is not claimed.

## Please look hardest at
1. Admin-only: any non-admin path that renders the panel? (backend refuses regardless)
2. Duplicate grant really refused by the server (not just shown) — e2e asserts the 409 text.
3. ComboBox server mode: keyboard + screen reader behaviour when the list is empty/loading.

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: (1) admin-only — чи панель рендериться ЛИШЕ для admin (не instructor, який теж має /admin),
бек GrantCourse refuses non-admin незалежно; (2) дубль grant → сервер 409; (3) ComboBox a11y
empty/loading (клавіатура/скрінрідер). + два фікси дефектів (server-mode, kept choice), зняті as-касти.

### F1 [#1 admin-only] — SOLID (фронт+бек)
admin.courses.$courseId.tsx:188 — `actor.role === 'admin' ? <AdminOwnerEditor adminTools={CourseGrantTool}/> : <CourseEditorView/>` (без adminTools). Instructor (що теж має /admin для СВОГО курсу через canAccessAdministration) потрапляє в ELSE → CourseGrantTool НЕ рендериться. actor.role з серверної сесії (не спуфиться). Бек GrantCourse: non-admin→ErrGrantForbidden 403 (O04, не чіпано) незалежно. Захист у глибину.

### F2 [#2 дубль grant — сервер, не лише UI] — ЧИСТО
adapter: 409→course_already_owned, 403→grant_forbidden(+origin). Wrong-user guard: навіть на 201
grant має збігатись user+course (course-grant-repository:59-60), інакше failure. Бек GrantWithAudit→
isUnique(course_grants_one_active_per_user_course)→ErrCourseAlreadyOwned 409 (O04): INSERT б'ється в
partial-unique → сервер РЕАЛЬНО відмовляє. e2e асертить 409-текст + перший grant цілий (студентський
урок відкривається). Не лише показ.

### F3 [#3 ComboBox a11y empty/loading] — ДОБРЕ, 1 nb
renderEmptyState «Немає відповідних варіантів» + allowsEmptyCollection → порожній список
announce-ається SR; клавіатура не трапиться (стрілки/Enter на порожньому — no-op, фокус в input).
Outcome: <output> (polite) для granted, role="alert" (assertive) для помилки. kept-choice фікс:
onInputChange не шукає коли text===chosen.label (дефект 2 закрито). reason-required: trimmed===null→
errorMessage, blank unsent; бек теж 1-500 runes. submit disabled-while-pending.
[nb] LOADING конфлейтиться з EMPTY: під час debounce(250ms)+fetch options порожні → SR чує «немає
варіантів», хоч ще вантажиться. Не-blocking a11y-нюанс; пропозиція: renderEmptyState розрізняти
isFetching → «Пошук…». 

### Інше — ЧИСТО
Зняття 2×`as PurchaseErrorCode` — чистий cleanup (типи вже дають код, typecheck зелений без касту;
узгоджено з TS-дисципліною «без assertion, що маскує»). Backend у 3b не чіпано.

## ПІДСУМОК: 0 blocking, 1 nb — COMMIT-READY (гейт 3a+3b попереду)
admin-only SOLID (фронт role==='admin'-гілка + бек 403); дубль grant сервер-відмова (409 unique) +
wrong-user guard; ComboBox empty-state+live-regions добре. nb: ComboBox loading-vs-empty (SR чує
«немає варіантів» під час fetch) — a11y-нюанс, не блокує. Open (peer): nginx rate-limit гіпотеза
недоведена — debounce лишений по суті, механізм не claimed (чесно). Вердикт на коді; ordered-гейт
3a+3b (з 200%-чеком) підтвердить наскрізно.

## Gate result (Courses 1, 2026-10-01)
Ordered gate #1 over 3a+3b: **red** — e2e R23 (playback bytes) got a 403 where a 200 was
expected after the pause click; 9 dependents did not run; the API log was lost to the suite's
own API restart, so the code behind the 403 is unknown. Isolated reseeded rerun of the playback
project: 84/84 green, R23 10.2 s. Ordered gate #2 (API log captured): exit 0, rstest 532/532,
e2e **88/88**; no heartbeat 403 in the log (503/410 are the suite's outage/expiry scenarios).
The 403 stays an open suspicion with no mechanism, not a cleared flake.

### R23 403 — narrowing (Courses 1, after Courses 2's lead, 2026-10-01)
The failing request was `POST /api/v1/playback/heartbeat`, not media bytes. On that path a
missing ownership gives **410** `playback_session_expired` and a takeover **409**
(`application/sessions.go` Heartbeat → `writePlaybackError`), so an ownership-TTL lapse cannot
produce this 403. Remaining sources of a heartbeat 403: the access rule via
`loadViewablePlayback` (`playback_forbidden` / `course_purchase_required`), an unauthenticated
request (401 rewritten to 403 by `normalizeAuthenticationDenial`), or the same-origin check.
Next time: capture the response body (error code) in R23's predicate, and keep the API log.
**Correction (Courses 2, verified by Courses 1 on the tree):** `normalizeAuthenticationDenial`
wraps only `GET /internal/playback/authorize` (handler.go:47); the heartbeat (handler.go:49) is
`requireSameOrigin(authenticate(…))`, so an unauthenticated heartbeat answers 401, not 403. The
"unauthenticated" source is withdrawn. Two remain: the access rule in `loadViewablePlayback`
(`playback_forbidden` / `course_purchase_required`) or the same-origin check (`origin_forbidden`).
The body code will name which.
