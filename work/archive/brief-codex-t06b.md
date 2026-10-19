# Codex brief — ratings T06b: learning-page review, admin buy entry, teach queue reviews

Repository /Users/strelia/projects/courses, branch `feat/ratings`, frontend in `frontend/`. Read first (they
bind): `AGENTS.md`, `docs/architecture/frontend.md`, `docs/ratings/spec.md` (AC1–AC7, AC10, G1–G10),
`docs/ratings/plan.md` §6 and §11 (Rt01–Rt05, Rt10, Rt11, Rt15–Rt18, Rt21–Rt23, Rt26, Rt29, Rt30). T06a is in
the tree (commit c5468dc): `RatingSummary`, review read model + `percent`, `CourseRepository.listReviews /
reportReview / ownReview`, locales `ratings.*` — reuse them. Orval already has every review operation
(`putOwnCourseReview`, `deleteOwnCourseReview`, `listLearningCourseReviews`, teach `removeReview`,
`dismissReviewReport`, the queue with `kind`) and `LearningCourse.canPurchase`. Log each step to
`work/codex-t06b.log` (first line before the first tool: `NEXT:`). Do **not** commit.

**Your last three frontend/backend hand-backs stopped without the required tests. This slice is accepted
only with every test in «Tests» written** — you cannot run `npm test` in your sandbox (port binding); write
them anyway, Courses 1 runs them. Write tests first.

Layering exactly as T06a: adapter (schema-validated, named error codes) → domain → application →
`app/composition` (Query, request-scoped) → presentation/routes. No fetch in presentation/routes; React Aria
primitives from `presentation/ui/`; every string in `locales/uk.ts` and `en.ts`; a11y per AGENTS. Never use
browser `confirm()`/`alert()` — use the UI dialog primitives.

## 1. Learning page review block (`routes/learn.$.tsx`, `presentation/learn/…`)
Read own review + eligibility (`ownReview`). Show, in a «Ваш відгук» section:
- **Own review exists** → the review (stars, text, «пройдено N%…», date, «змінено») with «Редагувати» and
  «Видалити». Edit opens the same form prefilled. Delete asks in a React Aria dialog, then calls DELETE;
  after success the form for a new review appears if still eligible (G10). A removed own review cannot be
  replaced: on `review_removed` show «Відгук знято модератором» and no form.
- **No review, eligibility `ok`** → the form: stars as a React Aria RadioGroup of 5 (each option named «1 з
  5» … «5 з 5»), a TextField/TextArea with a live counter «N / 2000» and the 20-character minimum shown
  before submit; submit disabled until valid. PUT → 201 shows the new review; 200 shows the updated one.
- **Not eligible** → a sentence per code, no form: `review_owner_forbidden` «Ви автор цього курсу — оцінювати
  його не можна»; `review_purchase_required` «Відгук можна залишити після покупки курсу»;
  `review_not_started` «Почніть проходити курс, щоб залишити відгук»; `review_not_available` «Курс зараз
  недоступний для відгуків».
- Write failures keep the typed text in the form and show the mapped message (`review_invalid`,
  `review_removed`, `dependency_unavailable` «Не вдалося зберегти, спробуйте ще», Rt29).
- Below it, «Відгуки студентів»: the learning-scoped list (`GET /learn/courses/{id}/reviews`, same item
  rendering as the course page — extract a shared review list component from T06a's CourseView instead of
  duplicating it), with «Показати ще» and the error/empty states. This is what makes an **archived** course's
  reviews visible to its grant holders (Rt30, AC10).

## 2. Admin buy entry (G2, Rt05)
When the learning course has `canPurchase: true`, show «Купити курс» in the course outline header even
though no lesson is locked for this viewer (admins). It runs the existing purchase start
(`learn.purchase.start` in `routes/learn.$.tsx`) and goes to the order page exactly as the locked-lesson
button does. Do not show it when `canPurchase` is false.

## 3. Teach queue reviews (`routes/teach.courses.$courseId.tsx`, `teach-queries.ts`, teach repository)
Render queue items of `kind: 'review'` (T05 left them filtered out): author, stars, text, open report count.
Actions: «Відхилити скаргу» (owner or admin → POST dismiss) and **«Зняти» only for an admin** (role from the
identity query; the course owner never sees it — G9). After an action the queue refetches. Comment items
keep their current behaviour.

## Tests (all required)
- adapter: put (201/200 and each refusal code), delete (204, `review_removed`, `review_not_found`), learning
  list mapping, teach remove/dismiss mapping, queue review item schema.
- presentation, learning block: form validation (19 vs 20 chars, 2000 / 2001, no star → submit disabled),
  submit success shows the review; edit prefilled; delete through the dialog; each not-eligible sentence;
  write failure keeps the text (Rt29); `review_removed` message.
- presentation, learning list: renders reviews for an archived course's viewer; error state with retry.
- admin buy entry: shown with `canPurchase` true, hidden when false; pressing it starts a purchase (spy).
- teach queue: review item renders; «Зняти» visible for an admin and absent for the owner; dismiss calls
  its mutation.
- a11y: the star radio group has an accessible name and each radio «N з 5»; the counter is announced politely.

## Verify
`cd frontend && npm run typecheck && npm run lint && npm run format:check`; list `npm test` as not run if denied.
Final answer: files changed, every test name added (per file), what you ran, what you could not.
