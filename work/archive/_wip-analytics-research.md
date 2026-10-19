# WIP — admin/creator analytics research (2026-10-10)

ДАЛІ(done): прочитав admin-analytics-plan.md (залізне обмеження: нема stats-ендпоінтів; реально лише users×3, courses×3).

ФАКТ (Teachable): sales trends, student progress, engagement; revenue analytics з MRR, student churn, failed payments, earnings/fees/payouts; unified view revenue/enrollments/engagement/funnels/community; KPI conversion rate, engagement, completion rate; video-аналітика (які відео реплеять, де drop-off, які викликають питання); топ-курси, seasonal trends. URL: teachable.com/analytics-and-reporting, teachable.com/blog/new-teachable-reporting-tools-2024

ФАКТ (Kajabi): Analytics > Reports. Revenue: Gross Revenue over time, Refunds. Subscriptions: New Subscriptions, Canceled, MRR. Offers: Offer Purchase Quantities. Analytics Overview customizable. URL: help.kajabi.com/articles/analytics/analytics-reports, kajabi.com/features/analytics

ФАКТ (Udemy instructor): Overview — course revenue, enrollments, ratings. Секції: Students (мапа, мови, інтереси), Reviews, Engagement (course+lecture level, total minutes consumed daily), Traffic & conversion. Marketplace vs Udemy Business split. URL: teach.udemy.com/publishing/using-your-dashboard/

ФАКТ (Thinkific): revenue/enrollments/engagement/funnels/community в одному; revenue dashboards — top products, trends, ARPU, subscription health; enrollment+order reporting (courses/bundles/memberships/communities); engagement+completion correlated to revenue/retention; identify products strong enrollment low engagement; Advanced Analytics — custom dashboards, role-based access, automated reporting. URL: thinkific.com/blog/thinkific-vs-kajabi-analytics/

ФАКТ (LearnWorlds): 24+ charts engagement/progress, radar charts top-performing courses; automated scheduled reports + export; user progression, exam results, survey, participation history; per-video engagement stats (Video Library); metrics — completed activities, viewers, avg study time, total study time, avg score; AI-powered insights. URL: learnworlds.com/analytics-and-reports/, support.learnworlds.com/.../12000077480-course-insights-activities-breakdown

ФАКТ (Podia): нативних деталей help-доків не знайдено в пошуку; знайдено лише сторонні (API-дашборди). Потрібен уточнюючий запит/WebFetch.

ФАКТ (Skillshare Teacher Stats, Creator Hub): class performance overall — total earnings, students, minutes watched, reviews across all classes; estimated earnings current month; charts minutes-watched та #students over time; month-by-month earnings breakdown. Payout залежить від minutes watched (частка від усіх teacher minutes). URL: help.skillshare.com/hc/articles/360044058112

ФАКТ (Moodle): standard reports — Logs (filterable events), Activity (view count), Activity completion (матриця студент×активність), Live logs; Course completion (прогрес до завершення за критеріями); Learning analytics (prediction models). URL: docs.moodle.org/en/Learning_analytics, docs.moodle.org/en/Reports

ФАКТ (Canvas): Course Analytics — grades (scatterplot class averages), Weekly Online Activity (page views + participations per week), Students, Reports, Online Attendance; New Analytics — avg course grades, participation vs classmates, refresh 24h, лише для published. URL: community.canvaslms.com (Course Analytics / New Analytics)

ФАКТ (TalentLMS): customizable dashboards (progress, user activity, course engagement); Infographics (system stats, Home>Reports>Infographics); Training matrix (візуальна матриця хто почав/прогресує/завершив, хто відстає, compliance по командах); Timeline (кожна подія, фільтр за user/course/period, аудит); course completion rates. URL: help.talentlms.com/.../360017623073, talentlms.com/features/lms-reporting

ФАКТ (Gumroad): Sales Analytics — всі продукти за замовч., фільтр по продукту; combined graph (кожен бар = день, grey=views, line=sales volume); hover — sales/views/absolute $; date range до першого продажу, daily/monthly toggle; метрики — sales+views, country data, primary traffic source, organic daily traffic, conversion rates. URL: gumroad.com/help/article/74-the-analytics-dashboard

ФАКТ (Circle): analytics hub — member journey; member growth + paid-membership mix; space-level engagement + quiet spaces; event registrations/attendance; course progress; member leaderboards; community health (response times, unanswered, contributor concentration). URL: circle.so/analytics

ФАКТ (Podia): all-in-one (сайт, цифрові продукти, спільнота, email); дашборд combines course sales, membership revenue, community metrics; офіційних help-доків із детальним переліком віджетів у пошуку не знайдено (solo-creator орієнтація). URL: rapidevelopers.com/retool-integrations/podia (3rd party); help.podia.com (офіц.)

ФАКТ (sales funnel / conversion): opt-in — landing page views, form submissions, conversion rate, lead source; sales page — sales page visits, checkout clicks, completed purchases, refunds; бенчмарк course landing conversion ~2.3% середнє, топ 8-12%; воронка lead→purchase. URL: zenler.com/blog/online-course-analytics-after-launch, thoughtmetric.io/calculate/how-to-calculate-conversion-rate-in-teachable, learndash.com (boost course sales)

ФАКТ (subscription/churn): MRR, Churn Rate (% скасованих за 30д), Active Subscribers, LTV, dunning recovery; cohort retention grid — групує за місяцем старту, показує коли drop-off; «cohort grid — найкорисніша в'юшка». URL: help.chartmogul.com/hc/en-us/articles/213789989 (cohort analysis), developer.paddle.com/concepts/retain/metrics

ФАКТ (Teachable Video Stats — конкретні метрики!): per-video — Engagement (% відео переглянуто в середньому), Play rate (% хто побачив і клікнув play), Loads, Visitors, Plays, Hours watched; Retention Report (popup — як довго дивились, drop-off); Lesson completion report (avg completion rate по курсу, скільки уроків позначено complete). URL: support.teachable.com/en/articles/11682570 (course reporting tools)

ФАКТ (multi-instructor LMS): multiple instructors per course, revenue splits (primary author + co-authors); аналітика для instructor згрупована в таби Revenue / Engagement / Students / Reviews (MasterStudy LMS); admin-side бачить revenue breakdown platform vs instructor share; окремі admin-analytics vs instructor-analytics в'юшки. URL: docs.stylemixthemes.com/masterstudy-lms/lms-analytics/analytics-for-instructors, .../analytics-for-admin

РІШЕННЯ: даних достатньо (12+ платформ покрито). Пишу docs/design/admin-analytics-research.md.

ФАКТ (done): docs/design/admin-analytics-research.md створено — шапка, каталог по 10 темах з URL, single-instructor розділ, топ-12 віджетів, відповіді на 4 прямі запити власника, прив'язка до беку (лише ТОП-12), джерела. Жоден із ТОП-12 віджетів не реалізується без нового stats-беку; реально 🟢 лише users×3+courses×3.


