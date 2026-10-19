claude-opus-4-8

# GATE 2 critique — accounts plan

ДАЛІ: прочитати spec.md + plan.md; підняти plan-critique; звірити з кодом access/identity/orders/discussion/compose.

## ФАКТИ (звірка з деревом)

- `contracts/access/access.go`: Subject = {UserID, Administrator}; CanView ігнорує будь-який verified-факт. Додавання поля EmailVerified у Subject саме по собі CanView не ламає (zero-value).
- 4 копії `CanBeViewedBy` (playback, transcript, discussion, course/learning) — усі `return access.CanView(...)`.
- ГЕЙТ-ТЕСТИ архітектури:
  - `platform/architecture/access_mapping_ast_test.go`: вимагає РІВНО 4 методи CanBeViewedBy, кожен — РІВНО один statement `return access.CanView(...)`. CanStream тут не пройде.
  - `platform/architecture/shared_rules_test.go` `TestTheAccessPredicateAgreesAcrossModules`: усі 4 копії мусять збігатися з TruthTable (accesstest), у якій НЕМАЄ виміру email_verified.
- Принципал: identity `Authenticate` middleware (adapters/http/handler.go:136-137) кладе `httpserver.Actor{UserID, Role}` у контекст; усі модулі читають `httpserver.ActorFromContext` (напр. playback handler.go:261). Модулі НЕ читають таблицю users — це факт identity. SessionStore.Resolve повертає domain.User.
- domain.User (identity/domain/user.go) — поля без EmailVerified; validDisplayName ≤120 рун (spec каже 1–100).
- Мутації:
  - orders: StartPurchase (service.go:43) гейт = domain.CanStartOrder (гранти), НЕ CanView; Checkout (108) = MarkCheckoutOpened. Жодного verified-чеку.
  - discussion: Publish(149), ReportComment(249) гейт = CanBeViewedBy; PutReview(367)/OwnReview гейт = domain.CanReview; DeleteReview(399) — БЕЗ гейта.
- seed-identities.sql: 6 рядків, INSERT-колонки без email_verified; ON CONFLICT DO UPDATE без email_verified.
- ІНШІ місця створення users: `frontend/e2e/ratings.spec.ts:139` вставляє users інлайн колонками (id,email,display_name,role,password_hash) — без email_verified; + ~15 integration-тестів INSERT INTO users. seed.sql користувачів не має.
- platform/mail — НЕ існує (новий, коректно). platform/redis — існує (client.go, queue.go). playback/adapters/redis — існує.
- argon2id Verifier.Hash + VerifyDummy (constant-time) — є (verifier.go:51,56). VerifyDummy викликається у login path (repository.go:47).
- DeleteIdentitySessionsForUser — plan каже «не викликається»: ХИБА, викликається 3× (administration.go:122,156,183); семантично видаляє ВСІ сесії user (для change-password «усі крім поточної» не підходить).
- mailpit/1025/8025/SMTP — у жодному docker-compose*.yml НЕМАЄ (plan додає — коректно, нова робота).

## ЛІНЗИ (4 запущено / 4 повернулось; інлайн, бо Task-tool недоступний лист-агенту — незалежність послаблена, усі 4 відпрацьовано)

### Логічні діри
B1 [blocking] CanBeViewedBy→CanStream ламає два гейт-тести.
  де: plan §4 рядок 61.
  чому: якщо playback/transcript CanBeViewedBy переписати на CanStream → access_mapping_ast_test (вимагає CanView) червоний; shared_rules_test (4 копії мусять збігатися з TruthTable без verified-виміру) червоний. Впорядкований гейт стає RED.
  що зробити: лишити всі 4 CanBeViewedBy=CanView; CanStream застосувати ОКРЕМИМ чеком у use-case Play (playback) і read транскрипту, не всередині CanBeViewedBy; розширити обидва архітектурні тести під CanStream. Переписати §4 рядки 60-61 і §5 рядок 70.

B2 [blocking] Джерело факту EmailVerified не визначене; «кожен модуль мапить свій факт» архітектурно неможливо.
  де: plan §4 рядок 65, §1 рядок 35.
  чому: модулі читають принципала лише з httpserver.Actor{UserID,Role} (identity.Authenticate handler.go:137). email_verified — факт identity (таблиця users); інші модулі її не читають. Без додавання EmailVerified у domain.User → SessionStore.Resolve → httpserver.Actor → Actor кожного модуля гейтам НЕМА ЧОГО перевіряти. ADR-013 застосовано хибно (це не «власний факт модуля»).
  що зробити: у плані назвати шов: EmailVerified у domain.User + httpserver.Actor, заповнення в identity.Authenticate, протяжка в Actor 4 модулів і orders. Виправити §1 рядок 35, §4 рядок 65.

B3 [blocking] change-password «усі сесії крім поточної» — механізму немає, цитований reuse вбиває поточну.
  де: plan §1 рядок 12 + §2 рядок 34 (ChangePassword) + AC6/G8.
  чому: єдиний цитований інструмент DeleteIdentitySessionsForUser видаляє ВСІ сесії user. Для change-password AC6 вимагає лишити поточну. Реалізація за планом → користувача розлогінить, AC6 порушено (хибний результат).
  що зробити: додати в план новий запит «delete sessions WHERE user_id=$1 AND token_digest<>current»; розмежувати reset (усі) vs change-password (крім поточної) у §2/§5.

### Edge cases
E1 [non-blocking] Сіди verified лише для 6 рядків; інші місця створення users лишаться false.
  де: plan §5 рядок 78, Acc20.
  чому: ratings.spec.ts:139 та ~15 integration-тестів вставляють users без email_verified → DEFAULT false. Інлайн-автори у ratings.spec пишуть reviews прямо в SQL (гейт оминають, ймовірно не ламаються), але план цього не аудіював. Ризик прихований.
  що зробити: розширити §5/Acc20 аудитом «жоден e2e не логіниться інлайн-створеним неверифікованим і не грає/мутує»; якщо логіниться — додати email_verified у той INSERT.
E2 [non-blocking] StartPurchase не названо серед гейтованих мутацій.
  де: plan §4 рядок 62 («orders checkout»).
  чому: перша мутація неверифікованого — StartPurchase (створення order, service.go:43), не Checkout. Якщо чек лише на Checkout — створення order тече.
  що зробити: назвати точні методи (StartPurchase + Checkout), а також PutReview/ReportComment/ReportReview.
E3 [non-blocking] Enumeration через per-email rate-limit.
  де: plan AC8/§5 rate-limiter.
  чому: якщо per-email лічильник інкрементиться лише для наявних акаунтів — оракул існування (всупереч VerifyDummy). План не каже, що ключ — на поданий email незалежно від існування.
  що зробити: зафіксувати: per-email ключ на поданий email завжди, однакова поведінка/тайминг для існуючих і ні.
E4 [non-blocking] Унікальність нового email не перевіряється на confirm email-change.
  де: plan §2 ConfirmEmailChange, §5 account_tokens.new_email.
  чому: email міг стати зайнятим між request і confirm; повторна перевірка унікальності на confirm не описана.
  що зробити: додати перевірку унікальності new_email на кроці confirm.

### Надлишкова складність
— чисто: reuse-first дисциплінований (argon2id, SessionStore, platform/redis, platform/httpserver, Clock переюзані), нових портів рівно під реальні зовнішні залежності (Mailer/SMTP, RateLimiter/Redis, UserWriter/TokenStore/Postgres), спекулятивного шва (Google) свідомо немає. Єдина фактична похибка — reuse-рядок про DeleteIdentitySessionsForUser (non-blocking, під B3).

### AI-slop
— чисто: без вигаданих абстракцій під один виклик, без церемонії; приймальні рядки Acc01–Acc20 конкретні, з рівнями U/C/I/E; 0 ручних. Зауваження: §6 «конституція ✔» — самозвіт без доказу на B1/B2 (архітектурні тести не згадані), але це не наповнювач.

## ВЕРДИКТ: ЧЕРВОНИЙ — 3 blocking відкриті (B1, B2, B3). 4 non-blocking (E1–E4).
Лінз: 4/4 (інлайн, не окремі субагенти — Task недоступний; назвав уголос).
