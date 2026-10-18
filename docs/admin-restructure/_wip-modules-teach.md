# A-батч — PR10: teach — ВИНЕСЕНО

## Слід (feature-local, contracts НЕ треба)
- domain/teach/teach.ts (чистий leaf, 0 імпортів) → modules/teach/domain/teach.ts.
- ports/TeachRepository.ts → modules/teach/ports/.
- adapters/api/teach-repository.ts (+test) → modules/teach/adapters/; тягне flat shared: generated/courses, course-schema (reviewSchema), transport, api-origin — через ../../../adapters/api/*.
- presentation/teach/{CourseRoster,LearnerOverview,ReviewModerationRow}(+tests) → modules/teach/ui/; залежать від власного domain, shared presentation/ui і modules/enrollment (RosterEntry, ReportDialog) через ../../enrollment.
- barrel modules/teach/index.ts: domain* + TeachRepository(type) + createApiTeachRepository + 3 UI.

## Лишено flat (репойнт на barrel ../../modules/teach або ../modules/teach)
- app/composition/teach-queries.ts(+test), teach.ts (wiring); app/router/learner-search.ts.
- routes: teach.tsx, teach.courses.$courseId.tsx, teach.learners.tsx, admin.learners.tsx, route-guards.test.tsx.

## Рішення
- Feature-local: teach споживає enrollment лише через barrel (крос-модуль дозволено для ui/routes), власний domain чистий — contracts не потрібні.
- adapter-тест під node: rstest adapter-node include вже ловить src/modules/*/adapters/** (підтверджено 8 тестів).

## Верифікація (quick, до ордерного гейта)
- oxfmt чисто; tsc 0; oxlint + check-architecture (219 файлів) 0; rstest повний 0 fail.
- Далі: ордерний make check (owner-схвалення, скидає спільні дані).

## ПІСЛЯ: admin (ВАЖКИЙ, останній) — заводиться contracts/identity.ts.
