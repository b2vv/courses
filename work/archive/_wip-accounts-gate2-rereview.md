claude-opus-4-8

# GATE 2 re-review — accounts plan (independent, second pass)

ДАЛІ: читаю docs/accounts/plan.md (оновлений), docs/accounts/spec.md, попередні робочі файли рецензентів (critique, defense). Далі звіряю кожну попередню знахідку з кодом.

## ФАКТ — звірка коду (всі цитати підтверджено)
- access.go: Subject{UserID,Administrator}; CanView (рядок 22) не чіпає verified. Додавання поля EmailVerified zero-value-safe. ✓
- AST-тест (backend/platform/architecture/access_mapping_ast_test.go): рахує РІВНО методи з іменем CanBeViewedBy (==4), кожен один `return access.CanView(...)`. CanStream (функція, не метод CanBeViewedBy) НЕ рахується. ✓
- shared_rules_test.go TruthTable: будує доменні структури модулів, без виміру email_verified; лишається зелений. ✓
- 4 копії CanBeViewedBy: course/learning.go:95, transcript/access.go:38, playback/authorization.go:38, discussion/access.go:26. ✓
- httpserver.Actor ЄДИНА конструкція в non-test коді: identity/adapters/http/handler.go:137 (Authenticate). Більше ніхто Actor не будує → шов B2 архітектурно справний, default-false загрози нема. ✓
- SessionStore.Resolve → domain.User (ports/authentication.go:21). ✓
- DeleteIdentitySessionsForUser викликається 3× (administration.go:122/156/183); DeleteUserSessionsExceptCurrent НЕ існує (новий). ✓
- validDisplayName відкидає >120 (user.go:88) → допускає 120; план ріже на 100. ✓
- orders: StartPurchase:43 (гейт CanStartOrder), Checkout:108 (MarkCheckoutOpened) — verified-чеку нема. ✓
- discussion: Publish:144, ReportComment:244, PutReview:367, ReportReview:402 (DeleteReview:399 без гейта — але створення огляду вже гейтовано PutReview, не нова діра). ✓
- seed-identities.sql: INSERT-колонки без email_verified; ON CONFLICT SET без email_verified → план E1 (обидві частини) потрібен і покритий §5:105. ✓
- Міграції: найвища 00013_ratings → 00014 коректна. ✓
- ratings.spec.ts seedReview: вставляє user І review прямо SQL, не логіниться/не грає → гейт оминає, лишається false коректно (план Acc20). ✓

## РІШЕННЯ — статус попередніх знахідок
- B1 ЗАКРИТО (plan §4:60-64, §6:123, T5). Усі 4 CanBeViewedBy=CanView; CanStream окремий предикат у use-case. Обидва арх-тести лишаються зелені (підтверджено кодом).
- B2 ЗАКРИТО (plan §1:35, §2:34-36, T3). Шов названо повністю; ЄДИНА конструкція Actor підтверджує здійсненність.
- B3 ЗАКРИТО (plan §1:13, §2:36, T3). Новий DeleteUserSessionsExceptCurrent; reset=усі. Факт «3×» виправлено.
- Q2 ЗАКРИТО (plan §4b:81-89, Acc23). SMTP-даун: signup не відкочує (verificationEmailSent:false+resend); forgot нейтрально; email-change транзієнт. Redis fail-open з логом.
- Q3 ЗАКРИТО (plan §4b:90-92, Acc21). 23505→email_taken, не 500.
- Q4 ЗАКРИТО (plan §4b:93-94, Acc22). Повторна унікальність new_email на confirm (TOCTOU).
- E1 ЗАКРИТО non-blk (plan §5:105-106, Acc20). Обидві частини seed + аудит інлайн.
- E2 ЗАКРИТО (plan §4:72-74, T5). StartPurchase+Checkout, Publish+ReportComment+PutReview+ReportReview.
- E3 ЗАКРИТО (plan §4b:98-99). Ключ з поданого email незалежно від існування.
- E4 ЗАКРИТО (=Q4).
- Верхні межі ЗАКРИТО (Acc17:151 пароль>128/ім'я>100). Атомарний consume ЗАКРИТО (§4b:95-97, Acc24).
- Фактичні ЗАКРИТО: 3× виклик виправлено (plan:12); 120 vs 100 (plan:34).

## Нові спостереження (non-blocking)
- N1: Банер «підтвердіть пошту» (§2 фронт, Acc02 «банер зникає») потребує, щоб `GET /api/v1/auth/me` віддавав email_verified. План явно не називає розширення `me`-DTO цим полем (лише T6 openapi/Orval загально). Імпліцитно покрито T6/T8, але не названо точково. Не тримає гейт.
- N2: DeleteReview:399 без email_verified-гейта — коректно, бо PutReview (створення) вже гейтовано; видалення власного огляду неверифікованим недосяжне. Не діра.

## ВЕРДИКТ: ЗЕЛЕНИЙ
Усі blocking (B1-B3, Q2-Q4) ЗАКРИТО, звірено з кодом. Правки нових блокуючих дір не внесли. Один дрібний non-blocking уточнювальний пункт (N1).
