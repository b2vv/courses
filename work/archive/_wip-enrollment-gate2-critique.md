claude-opus-4-8

# GATE 2 critique — enrollment plan

ДАЛІ: прочитати spec.md + plan.md, звірити claims про access/grants/learning/playback з кодом.

ФАКТ: access.go:22-24 CanView пускає free published без Entitled. access.go:27-28 CanStream=CanView&&EmailVerified (у КОНТРАКТІ, не use-case). access.go:37 accessExempt читає course.Entitled → Entitled не можна прибрати.
ФАКТ: grant.go — source{purchase,admin,grandfather}, state{active,revoked}; інваріант (source==purchase)!=hasOrder (ряд.43); CanStartOrder рахує тільки active.
ФАКТ: entitled-JOIN state='active' у: course/learning.go:46, transcript/repository.go:38, playback/playback.go:28, **playback/access.go:19-20 (LoadVideoCourse — nginx auth_request media-bytes gate!)**, discussion/repository.go:27/74, discussion/review.go:19, orders/repository.go:101.
ФАКТ: План (§0/§1/§3/§5) називає ЛИШЕ playback/playback.go:29 — playback/access.go НЕ згадано ніде.
ФАКТ: AST-тест access_mapping_ast_test.go вимагає рівно 4 CanBeViewedBy, кожен = access.CanView(...). План не чіпає CanBeViewedBy → OK.
ФАКТ: TruthTable (accesstest/table.go) 14 кейсів, "stranger published" free→Want:true. Тестує CanView; план не чіпає → OK. Фікстура-ізоляція тест забороняє імпорт accesstest у прод.
ФАКТ: 00011_payment.sql — `course_grants_one_active_per_user_course ... WHERE state='active'` (partial unique ТІЛЬКИ на active). Purchase-insert та grandfather-insert спираються на ON CONFLICT (user_id,course_id) WHERE state='active'.
ФАКТ: discussion мапить ОДИН boolean entitled (grants.id IS NOT NULL, state='active') у Course.Entitled → годує і CanBeViewedBy(read) і Publish(communicate). Publish/ReportComment: email-гейт потім CanBeViewedBy (application.go:146,153,249,256).

РІШЕННЯ: 4 лінзи застосовано одним рецензентом (свіжий контекст, автор — не я); окремих сабагентів не піднімав — називаю це явно. 4/4 лінзи дали знахідки.

ВЕРДИКТ: ЧЕРВОНИЙ — 5 blocking (див. фінальний звіт).
