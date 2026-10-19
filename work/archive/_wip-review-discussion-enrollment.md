ДАЛІ: diff discussion domain (participant_report, errors, review) + ports

ФАКТ: domain participant_report — NewParticipantReport перевіряє reporter==target (ErrReportSelf), courseID/reporter/target non-empty, kind∈{on_lecturer,on_student}, reason trim 1..500 рун, utf8. OK.
ФАКТ: ReportLecturer — EmailVerified→ReportSubject(actor)→courseFound→actor≠owner→IsParticipant(actor.membership active|blocked)→target=ownerID. Blocked member може репортити лектора (ймовірно навмисно).
ФАКТ: ReportStudent — ReportSubject(target)→courseFound→(admin||actor==owner)→IsParticipant(target.membership). Немає IDOR: ownerID з БД за courseID, звіряється з actor. OK.
ФАКТ: ReportQueue/ResolveReport — admin-only в application; роут authenticate+administrator; wired main.go:165 RegisterAdmin. OK.
ФАКТ: InsertParticipantReport ON CONFLICT (course,reporter,target) WHERE state=open DO NOTHING RETURNING → ErrReportDuplicate (409). Частковий unique index збігається. OK.
ФАКТ: ReportSubject LEFT JOIN course_grants без state-фільтра — безпечно лише бо міграція зробила один grant на (user,course) (course_grants_one_per_user_course). Підтверджено.
ФАКТ: ParticipantReportQueue — keyset (created_at,id)<($1,$2) DESC/DESC, limit+1, NextCursor з items[limit-1]. Коректно. limit клампиться 1..100 default 20.
ФАКТ: requireCommunication застосовано до Publish, ReportComment, PutReview, ReportReview, LearningReviews(CanConsume) — посилення доступу під enrollment (навмисне, задокументовано в коментарях).
РІШЕННЯ: модуль по суті чистий. Дві дрібні знахідки нижче.
ЗНАХІДКА (low): ResolveParticipantReport — Exec UPDATE ... WHERE id=.. AND state='open', ігнорує RowsAffected=0; резолв неіснуючого/вже-резолвленого reportId → 204 success без 404. Ідемпотентно, але адмін не бачить що ID хибний. Фікс: перевірити CommandTag.RowsAffected()==0 → domain.ErrNotFound.
ЗНАХІДКА (very-low): decodeParticipantReportCursor не валідує що parts[1] — UUID; підроблений серверний base64-курсор з не-uuid id → помилка на SQL-касті ($2::text::uuid) замість чистого ErrInvalidCursor(400). Лише для втручання в курсор. Фікс: не критично.
