# Review: access-control enforcement for paid courses (feat/payment)

ДАЛІ: прочитати diff main..HEAD по contracts/access, playback, transcript, discussion, course; знайти bypass авторизації для locked-уроків платних курсів.

ФАКТ: contract access.go — CanView = exempt || Published && (!Paid || Preview); PurchaseRequired = Published && Paid && !Preview && !exempt; exempt = admin||owner||Entitled. Чиста функція, без I/O. Тести (14 рядків truth-table + mutual-exclusion) проходять.
ФАКТ: усі 4 модулі (playback access.go+playback.go, transcript repository.go, discussion repository.go lessonFacts+commentQuery, course learning.go) читають lessons.is_preview для КОНКРЕТНОГО уроку й LEFT JOIN course_grants ... state='active' по actor.UserID. Preview — per-lesson, не per-course.
ФАКТ: playback 3 content-шляхи (nginx authorize, Play, Heartbeat) усі через CanBeViewedBy+PurchaseRequired. authorize re-check щозапит → ревокація діє одразу (no TOCTOU leak).
ФАКТ: learning CanBeViewedBy хардкодить Preview:true (видно лише outline/метадані); Locked рахується per-lesson через PurchaseRequiredBy(lesson.IsPreview). DTO не віддає playbackID/transcript для locked. Owner decision 2026-10-01.
ФАКТ: усі content-read хендлери (transcript/discussion/learning) вимагають auth (401 без юзера) → userID завжди валідний uuid; порожній userID не байпасить (fail-closed 500 у гіршому разі).
ФАКТ: schema course_grants state IN (active,revoked), partial unique one-active-per-user-course; grant join state='active' виключає revoked. Grandfather грантить усім з lesson_progress на курсі (ON CONFLICT ... DO NOTHING).
РІШЕННЯ: блокуючих bypass немає. Low/info: (1) discussion Update/Remove не re-check view-entitlement (автор редагує/видаляє власні коменти після refund — не leak контенту); (2) grandfather over/under-grant — політика, fail-closed; (3) learning outline метадані для платних — задокументоване рішення.
