claude-opus-4-8

ДАЛІ: прочитати skill plan-defense, docs/ratings/spec.md, docs/ratings/plan.md, AGENTS.md, .claude/standards.md — потім прогнати 5 фіксованих питань проти рядків plan.md.

ФАКТ: Звірено цитовані факти коду plan/spec:
- repository.go lessonFactsQuery (~20-34): grant LEFT JOIN state='active', courses.status, price_minor IS NOT NULL, is_preview, current_asset duration — reuse-claim тримається.
- comment.go:11 MaximumBodyRunes=2000; :40 utf8.ValidString + RuneCountInString 1..2000 — база для 20..2000.
- access.go: CanView(22) відкриває published paid через Preview; PurchaseRequired(26-28); accessExempt(31) = Administrator|owner|Entitled — G1/G2 тримаються.
- orders/service.go: StartPurchase(~43) не блокує адміна; GrantCourse(131) має лише !Administrator та валідацію reason, БЕЗ self-grant — G5 «немає перевірки сьогодні» вірно.
- playback/progress.go (~74): completed при >=0.95*duration — G6.
- migrations/00010:13 discussion_comment_reports.comment_id NOT NULL ... UNIQUE(comment_id,reporter_id) — reports/log-reuse тримається.
Усі цитовані факти збігаються з кодом.

РІШЕННЯ: 5 питань:
1 Кривий інпут — ЗАКРИТО: §4 ряд.90 (review_invalid для stars/body) + Rt10/Rt11 (19/20/2000/2001→400/201; 0/1/5/6→400/201/400) + §6 ряд.120 (текст причини на кожен код), CHECK у §3 ряд.47-48.
2 Відмова залежності — ДІРКА: жоден рядок не каже, як деградує нова секція відгуків при помилці/таймауті запиту, і що показує каталог, коли агрегат-join у ListPublishedCourses (§7 ряд.134 — найгарячіший публічний read) падає. §9 покриває лише ПОВІЛЬНО (>2×→matview), не ПОМИЛКУ. Нема і явного «поза скоупом». Новий каплінг: збій у даних відгуків валить увесь каталог — не адресовано.
3 Гонка/повтор — ЗАКРИТО: §3 ряд.54 UNIQUE(course_id,author_id)→review_already_exists (§4 ряд.88); §5 ряд.101 report 200 на повтор + Rt25; §7 ряд.137.
4 Порожні/граничні — ЗАКРИТО: Rt14 (0→empty-state; 25→20+«Показати ще»), Rt06-08 (поріг 2/3), Rt10/Rt11 (межі формату).
5 Rollback — ЗАКРИТО: §8 ряд.139-146 (Down: видалити log з review_id, drop таблиць, restore NOT NULL; revert merge; поля каталогу адитивні; єдина незворотність comment_id DROP NOT NULL відкочується Down; test-mode, prod даних нема).

ВЕРДИКТ: 1 дірка (питання 2). Гейт ЧЕРВОНИЙ, тримає рядок 2.

=== RE-CHECK commit adbe54b (рядки 2 і 3) ===
ФАКТ: handler.go:395 `errors.Is(err, domain.ErrDependencyUnavailable)` → :396 `return http.StatusServiceUnavailable`. План цитує 394-395; фактична case-лінія 395-396 (дрібний off-by-one на 2-му рядку, механізм реальний і присутній).
ФАКТ: CatalogStates.tsx існує, :13 <ErrorState title=catalog.failure retryHref retryLabel=common.retry> — error state + «retry» реальний.
ФАКТ: §7 ряд.147-160 новий булет «Dependency failure»: Postgres єдина залежність; каталог падає цілком як сьогодні (503 + CatalogStates) — rating:null свідомо відкинуто бо R6 забороняє 2-й запит; сторінка курсу — окремий запит reviews, на збій секція показує «Спробувати ще», решта рендериться (Rt28); learning — own-review збій не чіпає плеєр, невдалий write лишає текст (Rt29); writes/moderation → dependency_unavailable 503 через statusFor, remove=update+log у одній транзакції.
ФАКТ: Rt28 (ряд.227) 503→error «Спробувати ще» + outline; Rt29 (ряд.228) write 503→error+текст лишається. FE unit.

РІШЕННЯ ряд.2: ЗАКРИТО. §7 ряд.147-160 + Rt28/Rt29. Мій колишній закид «збій відгуків валить увесь каталог» тепер явно визнано й обґрунтовано (агрегат = один statement; graceful=2-й запит, R6 забороняє; no partial card) — це легітимне звуження скоупу рядком з механізмом.

ФАКТ: §4 ряд.93-96 create = INSERT … ON CONFLICT (course_id,author_id) DO NOTHING; 0 рядків → follow-up read → review_removed (removed) або review_already_exists; дві конкурентні create → один 201, один 409, без сирого constraint-error. UNIQUE §3 ряд.54 лишається backing index; узгоджено з §3 ряд.67-69 (removed тримає slot). Report-повтор §5 ряд.109 + Rt25 без змін.
РІШЕННЯ ряд.3: ЗАКРИТО, зміна ЗМІЦНює (детермінований 201/409 замість сирого 23505), нічого не відкриває. Не-блокуюче спостереження: мікрогонка create-vs-власний-delete між 0-рядків і follow-up read не адресована, але поза рамками Q3 (double-click/2 вкладки) і нехтовна.

ВЕРДИКТ RE-CHECK: 0 відкритих рядків (1,4,5 закриті раніше; 2,3 закриті тепер). ГЕЙТ ЗЕЛЕНИЙ.
