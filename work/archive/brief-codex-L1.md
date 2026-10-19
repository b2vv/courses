# Codex brief — L1: домен membership + міграція 00015 + backfill + seed

Репо `courses`, гілка `feat/enrollment`. План `docs/enrollment/plan.md` (§3, §4, §5). Роль: реалізуєш L1;
коміт/генерація/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-L1-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Факти (звірено)
- `course_grants` (00011): `source CHECK IN (purchase,admin,grandfather)`, `state CHECK IN (active,revoked)`,
  `CHECK ((source='purchase')=(order_id IS NOT NULL))`, partial-unique `course_grants_one_active_per_user_course`
  (WHERE state='active'). Домен — `orders/domain/grant.go` (`GrantSource`, `GrantState`, `NewGrant` з інваріантом).
- free курс = `courses.price_minor IS NULL`. `lesson_progress` — таблиця прогресу (playback). Discussion-активність —
  знайди таблиці коментарів/відгуків (00009_discussion / 00013_ratings) для backfill-union.

## 1. Міграція `backend/migrations/00015_enrollment.sql` (up+down, стиль наявних goose)
UP:
- `ALTER TABLE course_grants`: розширити CHECK `source` на `+'self'`; CHECK `state` на `+'blocked'`.
  (Інваріант `(source='purchase')=(order_id IS NOT NULL)` лишається — self/admin/grandfather без order.)
- **DROP INDEX `course_grants_one_active_per_user_course`; ADD `UNIQUE(user_id, course_id)`** (один рядок на пару).
- `CREATE TABLE participant_reports (id uuid PK, course_id uuid NOT NULL REFERENCES courses(id), reporter_id uuid
  NOT NULL REFERENCES users(id), target_id uuid NOT NULL REFERENCES users(id), kind text NOT NULL CHECK (kind IN
  ('on_lecturer','on_student')), reason text NOT NULL, state text NOT NULL DEFAULT 'open' CHECK (state IN
  ('open','resolved')), created_at timestamptz NOT NULL DEFAULT now(), CHECK (reporter_id <> target_id))`;
  **partial-unique** `UNIQUE(course_id, reporter_id, target_id) WHERE state='open'` (дедуп).
- **Backfill (union, щоб нікого не замкнути):** `INSERT INTO course_grants (id,user_id,course_id,source,state,...)`
  self/active для кожного (user,course), де курс free (`price_minor IS NULL`) і опублікований, user має рядок у
  `lesson_progress` на цьому курсі **АБО** discussion-активність (коментар/відгук) на цьому курсі, і ще немає
  рядка course_grants (user,course). `ON CONFLICT (user_id,course_id) DO NOTHING`. Генеруй id (gen_random_uuid),
  order_id NULL, granted_by NULL.
DOWN (дзеркально, конверсія перед звуженням):
- `DROP TABLE participant_reports;`
- `DELETE FROM course_grants WHERE source='self';`
- `UPDATE course_grants SET state='revoked' WHERE state='blocked';`
- DROP `UNIQUE(user_id,course_id)`; відтворити partial-active unique `course_grants_one_active_per_user_course`;
  звузити CHECK source→(purchase,admin,grandfather), state→(active,revoked).

## 2. Домен `orders/domain/grant.go`
- `GrantSource += GrantSelf = "self"`; `GrantState += GrantBlocked = "blocked"`.
- `NewGrant` прийняти нові source/state; інваріант лишається (self без order → ок).
- Доменні переходи (чисті функції/методи + помилки у стилі `orders/domain/errors.go`):
  `Block` (active→blocked), `Unblock` (blocked→active), `Revoke` (active|blocked→revoked),
  `self-enroll` створює active(source=self). Заборонені переходи — доменна помилка.
- Коментар: «цей запис — authoritative membership; стан = влада доступу (enrollment старший за grant)».
- TDD: table-driven unit на source/state валідацію, інваріант, усі дозволені/заборонені переходи.

## 3. Seed (`backend/sql/seed.sql` / `seed-identities`)
Щоб наявні e2e-студенти не втратили доступ до сідованих **free** курсів (free тепер вимагає членства):
додай self/active course_grants для відповідних сідованих користувачів на сідованих free-курсах. Звір, які
курси free (price_minor NULL) і які користувачі в них «вчаться» в наявних e2e, щоб не зламати suite (клас E1).
Якщо не певен щодо конкретних пар — познач у робочому файлі, Courses 1 уточнить на верифікації.

## Правила
- Домен чистий (stdlib+contracts). НЕ запускай sqlc (нові queries — у L2; L1 лише домен+міграція+seed). generated не чіпай.
- НЕ комість. TDD; час — спершу тести вже-написаного.

## Критерій готовності
- `go build ./...`, `go vet ./internal/orders/...`, `gofmt` чисті.
- `go test ./internal/orders/domain/...` зелені (нові source/state/переходи).
- Міграція 00015 up+down синтаксично валідна (застосування на чистій БД перевірить Courses 1 через integration).
- У робочому файлі: які seed-пари додано й чому; чи backfill-union покриває наявні e2e-патерни.
