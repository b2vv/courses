# WIP: Backend modules review (feat/payment vs main) — excl orders/**

ДАЛІ: отримати diff main..feat/payment по backend (крім orders), знайти міграцію, прочитати access contract.

ФАКТ: Прочитав повний diff + обидві міграції (00011, 00012). Access contract (access.go): CanView = exempt || (Published && (!Paid||Preview)); PurchaseRequired = Published && Paid && !Preview && !exempt; exempt = admin||owner||Entitled. Логіка узгоджена в 4 модулях (course/discussion/transcript/playback) — усі мапять Paid/Entitled/Preview і роблять if !CanView {if PurchaseRequired -> ErrPurchaseRequired else ErrForbidden}.
ФАКТ: Нові SQL у discussion/transcript/playback/learning додали `LEFT JOIN course_grants ... ON grants.user_id = $1::text::uuid`. userID тепер каститься до uuid. РИЗИК: якщо endpoint доступний анонімно (userID==""), `''::text::uuid` -> помилка Postgres -> 500. Треба перевірити authn кожного endpoint.
ДАЛІ: перевірити authn для transcript/discussion list/playback — чи може actor.UserID бути порожнім; прочитати application/authoring.go цілком (firstPriceSet + optimistic concurrency); перевірити grandfather call expectedUpdatedAt.

ФАКТ: Authenticate middleware (identity/adapters/http/handler.go:123) відхиляє порожній токен -> ErrAuthenticationRequired. transcript/discussion/playback/learning усі за Authenticate. Отже actor.UserID НЕ буває порожнім -> `$1::text::uuid` не отримує '' -> НЕ баг (перевірено, не вгадано).
ФАКТ: go build in-scope пакетів -> exit 0. Grant WRITE SQL поза orders лише в тестах; продакшн grandfather SQL у orders (поза скоупом).
ФАКТ: contracts/access_test.go тестує інваріант PurchaseRequired => !CanView по всій truth table + окремо owner/entitled/preview. 14 кейсів.
ФАКТ: playback access_integration_test перевіряє, що hot-query використовує індекс course_grants_one_active_per_user_course (reliability ок). Partial unique index гарантує <=1 active grant -> LEFT JOIN не множить рядки.
ФАКТ: firstPriceSet гілка (authoring.go:150-175): не викликає repository.Save; price пише grandfatherer з expectedUpdatedAt=saved.UpdatedAt (свіжий Load, не клієнтський токен). Захищено DeepEqual(original, storedForCompare) + DB optimistic lock на свіжому UpdatedAt + ON CONFLICT партіального індексу -> подвійний grant/гонка безпечні. Клієнтський expectedUpdatedAt у цій гілці ігнорується, але поведінка fail-safe. НЕ блокер.
ФАКТ: main.go grandfatherer завжди підключений (orders service будується завжди, навіть без WayForPay). nil-гілка -> ErrDependencyUnavailable (fail-closed, для тестів).
ФАКТ: catalog (repository.mapCourse) віддає Price/Paid, але НЕ Entitled -> без per-user даних у публічному кеші. Learning outline віддає titles/videoState/duration неоплаченому (owner decision 2026-10-01); реальний контент гейтиться окремо playback authorize (defense-in-depth).

РІШЕННЯ: немає блок-ворзі дефектів. Логіка узгоджена між 4 модулями, міграція безпечна/оборотна, authz fail-closed. Дрібні спостереження нижче.
