# Brief — L6c: «Моє навчання» list endpoint (GET /api/v1/learn/memberships)

РОБОЧИЙ ФАЙЛ: `work/_wip-L6c-codex.md`
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого інструменту:
рядок `ДАЛІ:`. Далі ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ. Фінальна відповідь = стислий переказ файлу.
Стандартні правила — `work/codex-working-model.md`: name the denial; НЕ чіпати git/tasks.md/
docs report/generated-руками; missing contract → stop; inaccuracy that would cause a bug → `## Discussion`.

## Контекст
Фаза enrollment, гілка `feat/enrollment`. L5/L6a/L6b закомічені й **заморожені**. Це amendment-зріз
(`docs/enrollment/plan.md` §Amendment, 2026-10-07): «Моє навчання» (En01) потребує списку власних
курсів користувача; per-course membership недостатньо. Додаємо один read-ендпоінт в **orders**-модулі.

## Завдання
`GET /api/v1/learn/memberships` — членства ПОТОЧНОГО користувача у стані `active|blocked`, кожне з
карткою курсу. Лише `authenticate` (власні дані; **без** email-gate — це read, як наявний GET membership).
Реєстрація у `orders` `Register(...)` поряд із `GET .../membership` (`handler.go:42`).

Відповідь: `{ "items": [ { "courseId", "slug", "title", "state", "progressPercent", "enrolledAt" } ] }`.
Впорядкування: `granted_at DESC`. **Без пагінації** (кількість власних членств обмежена — десятки;
keyset не потрібен, межу названо в плані). Safety-cap LIMIT 200 у SQL — захист від патології.

## Шари (дзеркаль наявний Roster)
**Port** (`orders/ports/repository.go`): новий тип
`type MembershipCard struct { CourseID, Slug, Title string; State domain.GrantState; ProgressPercent int; EnrolledAt time.Time }`
і метод у repository-інтерфейсі: `MyMemberships(context.Context, userID string) ([]MembershipCard, error)`.

**Application** (`orders/application/service.go`, поряд із `Roster` ~240):
`func (service *Service) MyMemberships(ctx, actor Actor) ([]ports.MembershipCard, error)` —
тонко: `return service.repository.MyMemberships(ctx, actor.UserID)`. Без додаткового authz (authenticated
власник читає своє). Якщо `actor.UserID == ""` — нехай транспорт уже відсік (actorFromRequest); не додавай.

**Postgres** (`orders/adapters/postgres/repository.go`, дзеркаль `Roster` 132–204): той самий progress-
CASE `(completed.completed*100 + totals.total/2)/totals.total` з двома CROSS JOIN LATERAL по
course_modules/course_lessons/lesson_progress, але:
- `FROM course_grants g JOIN courses c ON c.id = g.course_id` (для slug/title);
- `WHERE g.user_id = $1::text::uuid AND g.state IN ('active','blocked')`;
- `ORDER BY g.granted_at DESC`; `LIMIT 200`.
- SELECT `g.course_id::text, c.slug, c.title, g.state, g.granted_at, <progress>`.
- scan: той самий switch по state (active/blocked інакше `domain.ErrDependencyUnavailable`), як `scanRosterEntry`.
Курс показуємо незалежно від publish-статусу (членство дає доступ; архів зберігає — En09). Не фільтруй по status.

**HTTP** (`orders/adapters/http/handler.go`): тонкий `myMemberships` — `actorFromRequest`, виклик,
мапінг у DTO `{ items: [...] }`. `noStore`/cache-conventions як у наявних learn-GET. Помилки — наявна
orders writeError-сітка.

## OpenAPI + Orval
Додай path `GET /api/v1/learn/memberships` + response schema (напр. `MyMembershipsPage` з `items: MembershipCard[]`,
`MembershipCard` з полями вище; `state` enum active|blocked). Регенеруй Orval (`make generate` / `frontend npm run generate`).
`make check-generated` має бути зелений **зі staged generated** (Courses 1 застейджить і перевірить).

## Тести
- Application: `MyMemberships` делегує з actor.UserID (fakeStore).
- Integration (`orders/adapters/postgres`): сід user із 2 членствами (active+blocked) на різних курсах
  + 1 revoked (НЕ має зʼявитись) + прогрес на одному; перевір порядок granted_at DESC, progressPercent,
  slug/title, що revoked виключено, і що членство іншого юзера не тече.
- Contract HTTP: happy (items shape, без email), порожній список → `{items: []}`.

## Межі
НЕ чіпати L5/L6a/L6b логіку, discussion, інші модулі. НЕ комітити. Реальні make integration/e2e — Courses 1.

## Перевірка перед здачею (у WIP)
`go build ./...`; `go test ./internal/orders/...`; `go test ./platform/architecture/...`;
`git diff --check` (якщо standing-rule дозволяє — інакше познач не-run); `cd frontend && npm run typecheck`.
Докер integration лишаєш reviewing session.
