# A-батч — PR11 (останній): admin — ВИНЕСЕНО

## Слід (feature-local; contracts/identity.ts НЕ знадобився)
- domain/admin/{administration(+test),stats} → modules/admin/domain/ (обидва незалежні, 0 зовн. імпортів).
- application/admin/{authoring(+test),users(+test)} → modules/admin/application/ (тягнуть власні domain/ports + contracts/course).
- ports/{AdminCourseRepository,AdminStatsRepository,AdminUserRepository} → modules/admin/ports/.
- adapters/api/admin-{course,user,stats}-{repository,schema}(+tests) → modules/admin/adapters/; тягнуть flat shared через ../../../adapters/api/* (generated, course-schema, price-schema, transport, api-origin, timestamp-schema) + contracts/course.
- presentation/admin/* (18) → modules/admin/ui/; тягнуть shared presentation/ui і ЧУЖІ модулі через barrel (identity, transcript, purchase, media, enrollment) — ui→barrel дозволено.

## Ключові рішення (2026-10-18)
- **contracts/identity.ts НЕ потрібен** (всупереч прогнозу handoff): AuthenticatedUser чіпає лише admin UI (+ flat composition), а module ui може тягти чужий barrel (підтверджено правилами check-architecture рядок 54/64 обмежують лише domain/application/ports; teach/ui вже тягнув enrollment на зеленому гейті). Deep-шари admin identity-вільні. Код виграв за briefing.
- **Спільний timestamp-schema винесено з admin**: adminTimestampSchema (generic z.iso.datetime, НЕ admin-специфіка) живив і modules/media/adapters/lesson-video-schema. Якби лишився в admin-модулі — media тягнув би admin deep (заборонено) або через barrel → цикл media↔admin. Винесено в adapters/api/timestamp-schema.ts як isoTimestampSchema (flat shared); admin і media тягнуть звідти. Pre-launch cut-freely.
- barrel index.ts: domain*/application* wholesale, 3 ports (type), 3 фабрики create*, UI (8 зовнішніх + їх Pending/Failure/Dialog/NavSection + AdminNavIcon для AdminKit.stories). Внутрішні CoursePriceField/DocumentLibrary/ResetPasswordDialog — лише sibling.

## Споживачі (лишились flat, репойнт на barrel)
- app/composition/{admin,admin-queries(+test),admin-user-queries(+test),admin-stats-queries}; app/router/admin-search.
- routes: __root, admin.{refunds,analytics,courses.$courseId,participant-reports,index,users.index}, route-guards.test.
- modules/media/ui/LessonVideoField.test (CourseEditorView); presentation/ui/AdminKit.stories (AdminNavIcon).
- composition admin-grant-queries/admin-search — НЕ модуль, не чіпав.

## Верифікація (quick, до ордерного гейта)
- oxfmt чисто; tsc 0; oxlint + check-architecture (221 файл) 0 — жодного deep cross-module/cycle; rstest 773/0; admin adapter-тести під adapter-node (17 у 2 файлах).
- Далі: ордерний make check (owner-схвалення). ОСТАННІЙ модуль A-батчу.
