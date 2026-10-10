# Admin/creator analytics — інтернет-ресерч зрілих платформ

| | |
|---|---|
| Дата | 2026-10-10 |
| Автор | Courses (дослідницький агент, на запит власника «спроєктувати повноцінний розділ аналітики») |
| Що це | Каталог того, що показують у адмін/creator-аналітиці зрілі платформи онлайн-курсів і LMS, згрупований по 10 темах, + режим single-instructor, + ранжований shortlist віджетів для нашого дашборда, + прив'язка до нашого беку. |
| Метод | WebSearch по офіційних help-центрах/доках платформ і репутабельних UX/analytics-джерелах. Кожне твердження про конкретну метрику підкріплене URL (див. «Джерела»). Де офіційних деталей не знайдено — позначено «не знайдено». |
| Контекст | `docs/design/admin-analytics-plan.md` (залізне обмеження: у беку НЕМА агрегатних/stats-ендпоінтів; реально рахуються лише users×3 і courses×3 через `pagination.total`; усе цінне потребує нового stats-шару). |
| Платформи | Teachable, Thinkific, Kajabi, LearnWorlds, Podia, Udemy (instructor), Skillshare (teacher), Moodle, Canvas, TalentLMS, Circle, Gumroad, + MasterStudy LMS (multi-instructor), ChartMogul/Paddle (subscription benchmark). |

> Як читати «корисність»: **MUST** — зустрічається майже всюди й прямо відповідає на питання власника; **важливо** — часто, цінне на другому екрані; **nice** — вузька/нішева цінність.

---

## Каталог аналітики по 10 темах

### 1. Продажі й дохід

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Gross revenue over time** (загальний дохід за період, графік) | Kajabi (Gross Revenue report), Teachable, Thinkific, Gumroad (bar-графік по днях), Skillshare (month-by-month earnings) | **MUST** |
| **Дохід за курсом / топ-продуктом** | Teachable (внесок курсів у дохід), Thinkific (revenue dashboard — top products), Gumroad (фільтр по продукту) | **MUST** |
| **Refunds / повернення** | Kajabi (Refunds report — повні й часткові), Teachable | важливо |
| **MRR (Monthly Recurring Revenue)** | Kajabi, Teachable, Thinkific (subscription health), ChartMogul/Paddle (індустрійний стандарт) | важливо (лише якщо є підписки) |
| **New / Canceled subscriptions** | Kajabi (окремі репорти) | важливо (для підписок) |
| **AOV / ARPU** (середній чек / дохід на користувача) | Thinkific (ARPU в revenue dashboard) | важливо |
| **Earnings / fees / payouts** (нетто після комісій) | Teachable (earnings, fees, payouts), Skillshare (estimated earnings) | важливо |
| **Failed payments / dunning** | Teachable (failed payments), ChartMogul/Paddle (dunning recovery) | nice |
| **Дохід за купоном/промо** | Kajабі (offer purchase + знижки/trial враховуються), Gumroad | nice |

Джерела: Teachable analytics-and-reporting; Kajabi analytics-reports; Thinkific analytics; Gumroad help/article/74; Skillshare articles/360044058112; ChartMogul cohort; Paddle retain/metrics.

### 2. Популярність курсів

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Enrollments per course** | Udemy (enrollments в Overview), Thinkific (enrollment reporting по courses/bundles), усі LMS | **MUST** |
| **Топ-курси** (за enrollments / доходом) | Teachable (які курси performing best), Thinkific (top products), LearnWorlds (radar chart top-performing courses) | **MUST** |
| **Конверсія відвідувачі→покупці** | Gumroad (conversion rate + primary traffic source), Udemy (Traffic & conversion), Teachable | важливо |
| **Landing / sales-page conversion rate** | галузевий бенчмарк ~2.3% середнє, топ 8–12% (course landing) | важливо |
| **Нові vs повторні покупці** | згадується в funnel-аналітиці; окремого стандартного віджета рідко — здебільшого через когорти | nice |
| **Seasonal trends** | Teachable (сезонні тренди, ефективність кампаній) | nice |

Джерела: Udemy using-your-dashboard; Thinkific; Gumroad; Teachable; zenler.com online-course-analytics; thoughtmetric.io teachable conversion.

### 3. Залученість / споживання (відео)

Це найдетальніша тема в зрілих платформ — прямо відповідає на «що саме дивляться».

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Watch time / minutes consumed** | Skillshare (minutes watched — ядро виплат), Udemy (total minutes consumed, daily), Teachable (Hours watched), LearnWorlds (total/avg study time) | **MUST** |
| **Per-video engagement** (% відео переглянуто в середньому) | Teachable (Engagement %), LearnWorlds (per-video stats) | **MUST** |
| **Retention / drop-off curve** (де кидають відео) | Teachable (Retention Report popup — як довго дивились), Udemy (course+lecture reporting), edX Insights (video engagement) | **MUST** (прямий запит власника «що дивляться») |
| **Play rate** (% хто побачив і натиснув play) | Teachable (Play rate) | важливо |
| **Найбільш / найменш переглядані лекції** | Teachable (Video Stats per-video: Plays, Loads, Visitors), Udemy (lecture-level), Moodle (Activity report — view count) | **MUST** |
| **Які відео реплеять / викликають питання** | Teachable (які відео реплеять, які викликають більше питань) | важливо |
| **Products strong enrollment but low engagement** | Thinkific (прямо виділяє такі курси) | важливо |

Джерела: Teachable articles/11682570 (Video Stats: Engagement, Play rate, Loads, Visitors, Plays, Hours watched, Retention Report); Udemy dashboard; Skillshare; LearnWorlds course-insights; Moodle Reports.

### 4. Люди / студенти

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Нові реєстрації за період** (signups over time) | Circle (new members, net growth), Skillshare (#students over time), усі | **MUST** |
| **Активні за період** | Circle (member activity), Moodle (logs/live logs), Canvas (weekly online activity) | важливо |
| **Географія** (мапа де учні) | Udemy (мапа learners, мови), Gumroad (country data) | важливо |
| **За роллю** (student/instructor/admin) | адмін-зріз у будь-якій multi-role системі | важливо (для нас) |
| **Інтереси / мови аудиторії** | Udemy (languages, topics of interest) | nice |

Джерела: Udemy; Gumroad; Circle analytics; Skillshare; Moodle; Canvas.

### 5. Прогрес і завершення

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Completion rate** (середня по курсу) | Teachable (lesson completion report — avg), TalentLMS (course completion rates), Moodle (course completion), Canvas | **MUST** |
| **Середній прогрес учнів** | Thinkific (progress), LearnWorlds (user progression), Circle (course progress) | **MUST** |
| **Training matrix** (хто почав/прогресує/завершив, хто відстає) | TalentLMS (Training matrix — візуальна матриця user×course) | важливо |
| **Activity completion matrix** | Moodle (completion matrix студент×активність) | важливо |
| **Сертифікати** | Moodle, TalentLMS (compliance) | nice (у нас нема сертифікатів) |

Джерела: Teachable; TalentLMS reporting; Moodle Learning_analytics; Thinkific; LearnWorlds; Canvas.

### 6. Інструктори (прямий запит власника — «в яких лекторів»)

Marketplace- і team-платформи виносять інструктора як окремий вимір аналітики.

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Per-instructor analytics у табах** Revenue / Engagement / Students / Reviews | MasterStudy LMS (analytics-for-instructors: саме ці 4 таби), Udemy (ті самі секції на рівні інструктора) | **MUST** (для multi-instructor) |
| **Дохід по інструктору + split platform/instructor** | MasterStudy (admin бачить revenue breakdown platform vs instructor share), multi-instructor LMS (revenue splits, ledger rows) | **MUST** |
| **Enrollments / ratings / watch по інструктору** | Udemy (course ratings, enrollments на рівні інструктора), Skillshare (per-teacher earnings/students/minutes) | **MUST** |
| **Admin-view vs instructor-view** (адмін бачить усіх, лектор — лише себе) | MasterStudy (окремі analytics-for-admin та analytics-for-instructors), Udemy (instructor бачить власний Performance dashboard) | важливо |
| **Primary author + co-instructors** | multi-instructor LMS (primary + co-authors, split ≤ line item) | nice (у нас 1 автор на курс) |

Як marketplace це показують: Udemy дає кожному інструктору власний Performance dashboard (Overview: revenue, enrollments, ratings; + Students/Reviews/Engagement/Traffic); Skillshare — Teacher Stats у Creator Hub (earnings, students, minutes watched, reviews across all classes). Адмін платформи (MasterStudy) має **окремий** admin-analytics екран, де інструктор — це вимір/фільтр, і revenue split platform↔instructor.

Джерела: docs.stylemixthemes.com masterstudy analytics-for-instructors / analytics-for-admin; Udemy dashboard; Skillshare; wbcomdesigns multi-instructor.

### 7. Воронки й конверсія

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Visit → enroll → complete** (повна воронка навчання) | Teachable/Thinkific (unified revenue/enrollments/engagement/funnels), LearnWorlds (funnel stage view) | важливо |
| **Opt-in funnel** (landing views → form submissions → conversion) | галузеві практики (lead magnet 25–40%) | важливо (у нас нема lead-magnet) |
| **Checkout funnel** (sales page visits → checkout clicks → completed purchases → refunds) | галузеві практики, Gumroad (views→sales) | важливо |
| **Traffic & conversion** (джерела трафіку + конверсія) | Udemy (Traffic & conversion), Gumroad (primary traffic source, organic daily) | важливо |

Джерела: Teachable/Thinkific; LearnWorlds; zenler.com; Udemy; Gumroad.

### 8. Утримання / відтік (підписки, когорти)

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Churn rate** (% скасованих за 30д) | Teachable (student churn), Kajabi (canceled subscriptions), ChartMogul/Paddle | важливо (для підписок) |
| **Cohort retention grid** (групування за місяцем старту, коли drop-off) | ChartMogul («найкорисніша в'юшка»), Paddle (retention metrics) | важливо |
| **Active subscribers / net growth** | Circle (net growth, activation of recent signups), ChartMogul | важливо |
| **LTV (customer lifetime value)** | ChartMogul/Paddle, Thinkific | nice |

> Застереження для нас: ми продаємо разовий доступ (WayForPay-замовлення), не підписки. Churn/MRR/cohort релевантні лише якщо з'являться підписки/членства з продовженням. Зараз це **не must-have**.

Джерела: help.chartmogul.com cohort-analysis; developer.paddle.com retain/metrics; Teachable; Kajabi; Circle.

### 9. Відгуки / рейтинги як аналітика

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Reviews розподіл + тренд** | Udemy (Reviews секція), Skillshare (reviews across classes), Circle | важливо |
| **Рейтинг по курсу / по інструктору** | Udemy (course ratings в Overview, marketplace vs business split), MasterStudy (Reviews tab per-instructor) | важливо |
| **Community health** (час відповіді, без відповіді, concentration контриб'юторів) | Circle (community health dashboard) | nice (ближче до наших коментарів/модерації) |

Джерела: Udemy; Skillshare; Circle analytics.

### 10. Операції (тех-стани, обробка)

| Метрика / віджет | Платформи-приклади | Корисність |
|---|---|---|
| **Логи подій / timeline** (аудит, фільтр user/course/period) | TalentLMS (Timeline — кожна подія), Moodle (Logs, Live logs) | важливо (операційно/аудит) |
| **Активність по активностях** (view count) | Moodle (Activity report) | nice |
| **Обробка відео / стани транскодування** | Публічної per-platform аналітики саме по video-processing станах **не знайдено** (платформи ховають це як інфраструктуру; показують лише готовий контент) | nice (наша специфіка — медіа-воркер/ASR) |

Джерела: TalentLMS help 360017623073; Moodle Reports. Для video-processing станів — не знайдено публічного прикладу.

---

## Single-instructor режим

Запит власника: «зробити налаштування, якщо на порталі буде тільки один лектор».

### Як платформи різняться solo vs multi/marketplace

- **Solo-creator (один автор за замовчуванням):** Teachable, Thinkific, Kajabi, Podia, Gumroad. Аналітика **не має виміру «за інструктором»** взагалі — весь дохід/enrollments/watch належать власнику акаунту. Дашборд = «мій бізнес»: revenue, enrollments, engagement, funnels. Немає колонки/фільтра «instructor», немає split-рядків. Podia/Gumroad — найбільш спрощені (один storefront, один власник).
- **Marketplace (багато незалежних авторів):** Udemy, Skillshare. Кожен інструктор бачить **лише власний** Performance dashboard (Udemy: Overview + Students/Reviews/Engagement/Traffic; Skillshare: Teacher Stats у Creator Hub). Платформа-адмін агрегує по всіх, інструктор — вимір. Виплати рахуються per-instructor (Skillshare — частка minutes watched).
- **Team / multi-instructor (кілька авторів під одним брендом):** LearnWorlds, TalentLMS, MasterStudy LMS. Є **і** admin-view (усі інструктори, revenue split platform↔instructor, фільтр за автором) **і** instructor-view (лише свої курси). MasterStudy навіть має два окремі екрани: `analytics-for-admin` та `analytics-for-instructors`.

### Що міняється, коли лектор один

Спостережена закономірність: **solo-режим ховає весь вимір «інструктор» і спрощує навігацію.**
- Прибирають per-instructor розбивки (таб/колонка/фільтр «автор») — вони тривіальні (все належить одному).
- Прибирають revenue-split (platform vs instructor) — split 100% одному.
- Overview стає «мій бізнес», а не «порівняння авторів»; менше вкладок.
- «Автор» трактується як сам власник порталу — не окрема сутність в аналітиці.

### Що МАЄ робити наше налаштування single-instructor

Прапорець на рівні порталу (напр. `platform.singleInstructor = true`), і тоді:

1. **Ховати секцію/картку «Інструктори»** (тема 6) повністю — вона порожня при одному авторі.
2. **Прибирати колонку й фільтр «інструктор»** у таблицях курсів і в розбивках доходу/enrollments/watch.
3. **Не показувати per-instructor leaderboard / порівняння авторів** — немає з ким порівнювати.
4. **Спрощувати Overview-банд:** замість «дохід по інструкторах» — просто «дохід по курсах» і «топ-курси» (вимір курсу, не автора).
5. **Трактувати автора як платформу:** «дохід» = дохід порталу, без split-рядків platform↔instructor.
6. Тримати все **курсо-** і **контенто-центричним** (курс/урок/відео — виміри), бо саме вони лишаються інформативними при одному лекторі.
7. Коли прапорець знято (з'являється 2-й лектор) — ті самі віджети повертають вимір «інструктор» без переробки (дизайнити розбивки одразу, лише приховувати їх прапорцем).

---

## ТОП-віджети на дашборд (ранжований shortlist)

Критерій відбору: прямо відповідає на питання власника (популярні курси / дохід / що дивляться / в яких лекторів), зустрічається в більшості зрілих платформ, і дає дію (не просто число).

1. **Дохід за період (графік) + усього за діапазон** — ядро майже кожного creator-дашборда (Kajabi, Teachable, Thinkific, Gumroad, Skillshare). Прямо: «скільки заплатили».
2. **Топ-курси за enrollments і за доходом** (таблиця/бар) — Teachable, Thinkific, Udemy. Прямо: «популярні курси».
3. **Enrollments за період** (нові членства, графік) — усі LMS. Пульс попиту.
4. **Watch time / minutes consumed (всього + по курсах)** — Skillshare, Udemy, Teachable. Прямо: «що дивляться».
5. **Retention / drop-off по відео + топ і анти-топ лекцій** — Teachable (Retention Report + Video Stats), Udemy (lecture-level). Прямо: «що саме дивляться, де кидають».
6. **Completion rate / середній прогрес** — Teachable, TalentLMS, Moodle, Canvas. Чи доходять до кінця.
7. **Per-instructor панель (Revenue/Students/Reviews/Engagement)** — Udemy, MasterStudy. Прямо: «в яких лекторів». **Ховається в single-instructor режимі.**
8. **Нові реєстрації користувачів за період** — Circle, Skillshare. Ріст аудиторії.
9. **Conversion: відвідувачі курсу → покупці** — Udemy (Traffic & conversion), Gumroad. Чи працює sales-page.
10. **Рейтинги/відгуки: розподіл + тренд по курсах/лекторах** — Udemy, Skillshare. Якість очима учнів.
11. **Воронка visit→enroll→complete** — Teachable/Thinkific/LearnWorlds. Де втрати.
12. **Refunds / failed payments** — Kajabi, Teachable. Здоров'я продажів (nice-to-have на старті).

Чому саме ці перші: #1–#5 покривають усі чотири прямі запити власника (дохід, популярні курси, що дивляться, в яких лекторів) і є у 4+ незалежних платформах кожен; #6–#9 — наступний рівень цінності (якість навчання + ріст + конверсія); #10–#12 — доповнюють, але не критичні для першого екрана. Churn/MRR/cohort (тема 8) свідомо **не** в топі — ми не підписочна модель.

---

## Відповіді на прямі запити власника

**(а) Популярні курси** → тема 2 + віджети #2, #3. Стандарт: таблиця «топ-курси за enrollments і за доходом» + графік нових enrollments за період (Teachable/Thinkific/Udemy).

**(б) Скільки заплатили / дохід** → тема 1 + віджети #1, #12. Стандарт: графік gross revenue за період, дохід по курсах, + refunds. (Для нас: суми з WayForPay-замовлень, поки sandbox 1-UAH.)

**(в) Що дивляться і в яких лекторів** → теми 3 і 6 + віджети #4, #5, #7. «Що дивляться»: watch time, retention/drop-off по відео, топ/анти-топ лекцій (Teachable Video Stats + Retention Report, Udemy lecture-level). «В яких лекторів»: per-instructor панель Revenue/Students/Reviews/Engagement (Udemy, MasterStudy) + watch-time по інструктору (Skillshare).

**(г) Single-instructor налаштування** → розділ «Single-instructor режим». Коротко: прапорець порталу ховає вимір «інструктор» (секцію #7, колонку/фільтр автора, split), спрощує Overview до курсо-центричного; віджети проєктуються з виміром автора одразу, але приховуються прапорцем, щоб 2-й лектор не вимагав переробки.

---

## Прив'язка до нашого беку (лише для ТОП-віджетів)

Спираючись на `admin-analytics-plan.md` §2–3. Легенда: 🟢 є count зараз (`pagination.total`) · 🔴 треба новий stats-ендпоінт · 🕒 потребує timeseries (`created_at`-бакети).

| # | Віджет | Статус | Що бракує |
|---|---|---|---|
| 1 | Дохід за період | 🔴 🕒 | Нема admin-orders-агрегату взагалі; потрібен `orders.revenueMinorUAH` + timeseries по `created_at`. |
| 2 | Топ-курси за enrollments/доходом | 🔴 | Нема агрегату grants по курсах і orders по курсах; потрібен stats по course_id. |
| 3 | Enrollments за період | 🔴 🕒 | `learners` cursor без total; потрібен count grants + timeseries. |
| 4 | Watch time (всього + по курсах) | 🔴 | Прогрес/watched-marks читаються поштучно; нема агрегату watch-секунд. |
| 5 | Retention/drop-off + топ лекцій | 🔴 | Нема per-lesson/per-video агрегату переглядів; дорогий новий шар (подієвий збір позицій відтворення). |
| 6 | Completion rate / середній прогрес | 🔴 | Прогрес поштучно; потрібен агрегат avg progress / completion по курсах. |
| 7 | Per-instructor панель | 🔴 | Є `instructors` список, але розподіл дохід/enroll/watch/rating по автору — все через новий stats; **ховається прапорцем single-instructor**. |
| 8 | Нові реєстрації за період | 🔴 🕒 | Нема `created_at` у вибірці `AdminUser`; потрібен signups-timeseries. |
| 9 | Conversion відвідувачі→покупці | 🔴 | Нема обліку переглядів sales-page (нема аналітики трафіку взагалі); найдорожче — потрібен окремий збір visit-подій. |
| 10 | Рейтинги/відгуки розподіл+тренд | 🔴 🕒 | Рейтинги/відгуки існують як дані, але admin-агрегату (розподіл зірок, avg по курсах) нема. |
| 11 | Воронка visit→enroll→complete | 🔴 | Поєднує #9 (visit, нема) + #3 (enroll) + #6 (complete); найповніше залежить від нового беку. |
| 12 | Refunds / failed payments | 🔴 | Нема admin-orders-ендпоінта; поля refunded/declined недосяжні. |

**Єдине, що реально 🟢 зараз:** users×3 (active/blocked/deleted) і courses×3 (draft/published/archived) + два total — косметичний status-блок, **жоден із ТОП-12 віджетів без нового беку не реалізується.** Висновок плану підтверджено: справжня аналітика = спроєктувати `GET /api/v1/admin/stats` (знімок, SQL GROUP BY) для 🔴 і `stats/timeseries` для 🕒.

---

## Джерела

**Teachable**
- https://teachable.com/analytics-and-reporting
- https://teachable.com/blog/new-teachable-reporting-tools-2024
- https://support.teachable.com/en/articles/11682570 (Course reporting tools — Video Stats, Retention, lesson completion)
- https://teachable.com/blog/how-to-use-analytics-to-increase-your-course-sales-and-engagement

**Thinkific**
- https://www.thinkific.com/blog/thinkific-vs-kajabi-analytics/
- https://www.thinkific.com/blog/thinkfific-analytics-vs-teachable/

**Kajabi**
- https://help.kajabi.com/articles/analytics/analytics-reports
- https://help.kajabi.com/articles/analytics/analytics-overview
- https://kajabi.com/features/analytics

**Udemy (instructor)**
- https://teach.udemy.com/publishing/using-your-dashboard/
- https://support.udemy.com/hc/articles/360007889294
- https://teach.udemy.com/ub-insights-launch/

**Skillshare (teacher)**
- https://help.skillshare.com/hc/articles/360044058112
- https://help.skillshare.com/hc/en-us/articles/41694886825229-Skillshare-Teacher-Earnings-Update-Effective-January-1-2026

**LearnWorlds**
- https://www.learnworlds.com/analytics-and-reports/
- https://support.learnworlds.com/support/solutions/articles/12000077480-course-insights-activities-breakdown
- https://support.learnworlds.com/support/solutions/articles/12000063333-how-to-track-user-engagement

**Podia**
- https://help.podia.com (офіційний help; детального переліку віджетів у відкритому пошуку не знайдено)
- https://rapidevelopers.com/retool-integrations/podia (3rd-party опис даних)

**Moodle**
- https://docs.moodle.org/en/Learning_analytics
- https://docs.moodle.org/en/Reports

**Canvas / Instructure**
- https://community.canvaslms.com (Course Analytics / New Analytics — grades, weekly activity, participation, page views)

**TalentLMS**
- https://help.talentlms.com/hc/en-us/articles/360017623073-What-types-of-reports-are-available
- https://help.talentlms.com/hc/en-us/articles/360014659353-A-guide-to-the-portal-infographic
- https://www.talentlms.com/features/lms-reporting

**Circle**
- https://circle.so/analytics
- https://www.metabase.com/integrations/circle

**Gumroad**
- https://gumroad.com/help/article/74-the-analytics-dashboard

**Multi-instructor / per-instructor analytics**
- https://docs.stylemixthemes.com/masterstudy-lms/lms-analytics/analytics-for-instructors
- https://docs.stylemixthemes.com/masterstudy-lms/lms-analytics/analytics-for-admin
- https://wbcomdesigns.com/best-lms-with-multi-instructor-support/

**Funnel / conversion / subscription benchmarks**
- https://www.zenler.com/blog/online-course-analytics-after-launch
- https://thoughtmetric.io/calculate/how-to-calculate-conversion-rate-in-teachable
- https://help.chartmogul.com/hc/en-us/articles/213789989 (cohort analysis)
- https://developer.paddle.com/concepts/retain/metrics
