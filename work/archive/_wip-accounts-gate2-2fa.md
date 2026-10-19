model: claude-opus-4-8

ДАЛІ: читаю docs/accounts/spec.md (G13/G14, AC14–AC18) і docs/accounts/plan.md (§4c, §4d, §5, Acc25–Acc30, T11–T15); потім звіряю з кодом identity Login/handler, httpserver middleware, SSR.

## ФАКТИ (звірка з кодом)

- Login однокроковий: `identity.Login` завжди робить `sessions.Rotate` і повертає Token (identity.go:60-68); handler.login завжди ставить cookie + 200 userDTO (handler.go:91-97). Дельта робить його двокроковим (plan §4c 110-111, T14 line 223). Сідовані totp_enabled=false → однокрокова гілка не змінюється (plan 116-117) — правдоподібно; контракт-тест логіну (handler_test.go) 200+userDTO лишається для не-2FA.
- Actor НЕ має EmailVerified (actor.go:7-10) — план додає (ок, плановий пункт).
- RateLimiter у коді НЕ існує (grep порожній) — новий порт, план §1.
- Pending != сесія: Authenticate резолвить лише session-токени через SessionStore (identity.go:71-88, validToken вимагає 43-char/32-byte base64, handler.go:103-109); pending у Redis окремо → не можна подати як session cookie. Acc30 (I) структурно тримається.
- Admin reset пароля НЕ журналюється (administration.go:92-100 — лише SetPassword, жодного audit); в identity audit-механізму НЕМА (grep audit → лише orders + discussion/moderation). План (§4c 114-115) заявляє admin-disable-2FA "журналюється" — механізму немає й не названо.
- SSR HTML рендерить Node server.mjs (frontend:3000), TanStack `<Scripts/>` (__root.tsx:85) → інлайн hydration-скрипти. server.mjs ставить лише Cache-Control/Vary (server.mjs:23-24), НЕ CSP.
- Go `platform/httpserver` обслуговує лише /api/v1 JSON (nginx.conf:116-145; server.mjs:18-19 → 404 на /api,/media). Жодного HTML.
- nginx: `location /` проксі на frontend:3000 (nginx.conf:153-156), БЕЗ власного add_header → серверні add_header (nginx.conf:49-51: XCTO, Referrer-Policy, X-Frame-Options always) УСПАДКОВУЮТЬСЯ в проксі-HTML. Тобто nginx server-level add_header CSP покрив би й динамічний HTML.

## РІШЕННЯ (знахідки)

BLOCKING B1 — CSP розміщено там, де немає скриптоносної поверхні.
План §4d (line 121): CSP middleware у `platform/httpserver` + "дублем у nginx для статичних відповідей". Але:
- platform/httpserver віддає лише JSON /api (nginx.conf:116-145) — CSP на JSON браузер як документ не рендерить, захисту 0.
- Скрипти несе SSR-HTML з server.mjs (TanStack `<Scripts/>`, __root.tsx:85) — план server.mjs не згадує.
- nginx у плані явно звужено до "статичних відповідей", тобто динамічний HTML-документ НЕ отримує CSP від жодного названого компонента → AC14 «кожна відповідь несе CSP» не досягається як написано.
- Nonce-шлях (plan 124-125) реалізовний ЛИШЕ в server.mjs (генерує nonce + прокидає в TanStack `<Scripts/>`); nginx/Go не можуть вставити per-request nonce у тіло → заявлений спосіб no-unsafe-inline безхатній.
Дешева правка: CSP header на HTML-документ — nginx server-level add_header (успадковується в location /, покриває проксі-HTML) для статичного/report-only; strict+nonce — у server.mjs. Перемістити первинний механізм з platform/httpserver → server.mjs/nginx-HTML. Поки не перенесено — гейт червоний на дельті CSP.

NON-BLOCKING N1 — admin-disable-2FA "журналюється" без механізму (plan 114-115). В identity немає audit; наявний admin reset не журналюється (administration.go:92-100). Назвати механізм (audit-таблиця / структурований лог-івент) або зняти обіцянку (порушує правило «обмеження/рішення несе механізм»).

NON-BLOCKING N2 — атомарний consume резервних кодів не прописано. §4b line 95 дає атомарний UPDATE...WHERE consumed_at IS NULL RETURNING лише для account_tokens; recovery_codes має колонку consumed_at (line 138), але запит consume не названо. Acc28 (I/E, гонкоздатний). Прописати той самий атом для recovery_codes, інакше check-then-update double-use.

NON-BLOCKING N3 — AC17 «перегенерувати коди (старі недійсні)» не має Acc-рядка. Acc26 показує коди раз, Acc28 — одноразовість; регенерація-інвалідація старих не тестується. Додати рядок/розширити Acc.

NON-BLOCKING N4 — pending-токен: названо user+IP + TTL ~5хв (plan 110-111), але НЕ названо одноразовість pending-токена й де зберігається remember-me-вибір між кроком 1 і 2 (line 111 «видає сесію з remember-me», та pending має нести цей прапор). IP-прив'язка успадковує залежність від довіри X-Forwarded-For (та сама, що per-IP AC8 бази). Назвати одноразовість pending + носій remember-me.

NON-BLOCKING N5 — rate-limit 2FA: per-user ключ названо (line 112), поріг/вікно для 2FA-кроку явно не задано (посилання на AC8 ~5/15хв). Перебір 10^6 обмежений 5 спроб/pending (TTL 5хв) + re-login лімітований per-email+IP — прийнятно, але поріг 2FA варто зафіксувати явно.

ПІДТВЕРДЖЕНО ОК: TOTP stdlib hmac/sha1/base32 вікно ±1 реалістично (plan 103); секрет-plaintext обмеження названо з механізмом (spec 82-83, прод-KMS-обгортка відкладена); pending!=сесія структурно (Acc30); гонка email 23505→email_taken (§4b 91-92); атомарний consume account_tokens (§4b 95-97).

ВЕРДИКТ: ЧЕРВОНИЙ (1 blocking: B1 CSP placement; 5 non-blocking N1–N5).
