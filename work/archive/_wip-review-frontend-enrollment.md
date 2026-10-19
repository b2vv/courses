# Frontend review — feat/enrollment vs main

ДАЛІ: отримати список змінених frontend .ts/.tsx файлів (крім generated), зрозуміти обсяг

## КРОК 1 — core files read
ФАКТ: enrollment-repository.ts validators. rosterPage дозволяє nextCursor === "" (лише typeof string), а participantReportPage вимагає length>0. getNextPageParam роблять `page.nextCursor ?? undefined` — "" не nullish → hasNextPage=true → можливий infinite load / fetch cursor="". ПІДОЗРА на pagination bug у roster (залежить від бекенду).
ФАКТ: enrollment-queries invalidation виглядає повним (enroll→membership+myMemberships+['course']; block/unblock→roster; revoke→reports+roster; resolveReport→reports).
ДАЛІ: читаю presentation (CourseRoster, MyLearningList, ReportDialog, ParticipantReportQueue, CourseView, DiscussionPanel, ReviewBlock) і routes.

## КРОК 2 — routes + presentation + verification
ФАКТ: admin.participant-reports.tsx неститься під admin.tsx, який у beforeLoad робить requireIdentity + canAccessAdministration→notFound. Authz гейт є + бекенд enforce. OK.
ФАКТ: DiscussionQueryError/CatalogQueryError/TeachQueryError/EnrollmentQueryError усі super(code) → message===code. Гілки `error.message === 'enrollment_required'` у DiscussionPanel/ReviewBlock/ReviewCard спрацюють. OK.
ФАКТ: усі i18n ключі (enrollment.errors.*, enrollment.report.errors.*, ratings.errors.enrollment_*, learn.discussion.errors.*) присутні в en.ts. OK.
ФАКТ: roster nextCursor — бекенд handler.go:259 `if page.NextCursor != "" { response.NextCursor = &... }`, *string json → emits null або непорожній рядок, НІКОЛИ "". Отже валідатор, що приймає "", не веде до реального loop. Downgrade → nit.
ФАКТ: invalidation повна і коректна (enroll→membership+myMemberships+['course']; block/unblock→roster + onError membership_conflict→roster; revoke→reports+roster; resolveReport→reports). mutateAsync-проміси ловляться локально в діалогах/runRosterAction.
РІШЕННЯ: блокерів немає. Є кілька low/nit.
