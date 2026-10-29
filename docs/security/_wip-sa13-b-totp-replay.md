ДАЛІ (НЕ СТАРТУВАТИ до merge PR #83): pre-code план I-F5 TOTP consumed-counter. Код тримати, поки D-F2 не злитий (одна гілка в польоті).

# SA13-B — I-F5 TOTP replay consumed-counter

## Pre-code checkpoint (2026-10-29) — PLAN ONLY, код не розпочато
- **Задача / готовність:** SA13, другий зріз (AGENDA-порядок після D-F2). Джерело cite `audit-2026-10-24.md §LOW` I-F5: «TOTP replay within ~90s window (no consumed-counter)». READY-на-план; **implementation BLOCKED до merge PR #83** (щоб не тримати 2 незлиті гілки — рішення власника про темп + моя пересторога).
- **Дефект (механізм, звірено в коді):** `domain.VerifyTOTP(secret, code, now) bool` (two_factor.go:36) приймає код із вікна counters [now/30−1 .. +1] (±1 крок ≈ 90 с) і повертає лише bool — не фіксує, який counter збігся, і ніде не персиститься «останній спожитий крок». Отже валідний TOTP можна **повторно** подати в межах його вікна (і ±skew). Три replay-чутливі шляхи, усі в `application/two_factor.go`:
  1. `CompleteTwoFactor` :147 — **логін 2-й фактор** (названий аудитом; головний).
  2. `authorizeFactor` :249 — disable / regenerate-recovery (код, уже вжитий на логін, не має вимикати 2FA).
  3. `ConfirmTwoFactor` :90 (via `ValidateTwoFactorEnrollment`) — enrollment-confirm (enrollment-код не має одразу зіграти як перший логін).
  (Recovery-коди вже one-time через атомарний `ConsumeCode` — поза scope.)
- **Storage (факт):** 2FA-стан — колонки `users.totp_secret`, `users.totp_enabled` (міграція 00014). Остання міграція — `00022`. sqlc-queries → `adapters/postgres/generated/identity.sql.go` (+ .sql джерело); порт `ports.TwoFactorStore` (GetTwoFactor/SetTotpSecret/EnableTotp/DisableTotp).
- **Фікс (план):**
  1. **Міграція 00023**: `ALTER TABLE users ADD COLUMN totp_last_counter bigint NOT NULL DEFAULT 0;` (down: DROP). `0` = ніколи не вживано (реальний counter = unix/30 ≫ 0, тож перший успіх проходить).
  2. **Domain**: додати `VerifyTOTPAt(secret, code, now) (counter int64, ok bool)` — повертає counter, що збігся; `VerifyTOTP` лишити тонкою bool-обгорткою для місць, де counter не потрібен. Правило replay: приймати лише якщо matched > lastCounter.
  3. **Port**: `TwoFactor` отримує поле `LastCounter int64` (GetTwoFactor читає). Додати **атомарний** `AdvanceTotpCounter(ctx, userID string, counter int64) (bool, error)` → `UPDATE users SET totp_last_counter=$2 WHERE id=$1 AND totp_last_counter<$2` + rowsAffected==1. Умовний UPDATE закриває гонку одночасного повтору (plan-defense «гонка й повторний виклик»): два паралельні Complete з тим самим кодом → виграє один UPDATE.
  4. **Application**: у 3 шляхах — після доменного матчу взяти matched counter, викликати `AdvanceTotpCounter`; якщо повернув false (вже спожито ≥) → трактувати як невалідний. Будь-який успішний TOTP-proof просуває counter (спільний для login/disable/enroll — код не переносимо між шляхами).
  5. **Error code:** лишити `ErrInvalidTOTP` (без оракула replay-vs-wrong — однаковий «invalid»).
- **Поза scope:** I-F6 (`totp_secret` plaintext + unsalted recovery hash) — окремий обережний зріз (data-ish міграція), НЕ тут.
- **Доказ (план, production path):**
  - domain unit: `VerifyTOTPAt` дає правильний counter для now і ±1; відхиляє невалідне.
  - identity integration (RED→GREEN): enroll+enable → логін кодом кроку N (успіх, counter N персиститься) → **повтор того самого коду в межах вікна → відхилено** (RED: зараз приймає). Плюс concurrent: два Complete з тим самим кодом → рівно один успіх (атомарний UPDATE).
- **Impact / гейти:** backend Go (domain+application+port+repo) + міграція + sqlc-query → **`make check-generated`** стає релевантним (sqlc regen; memory check-generated-needs-staging: `git add` перед ціллю). Міграція міняє persisted schema → повний `make integration` + за впливом e2e логін-2FA-флоу (якщо e2e має 2FA-сценарій — звірити). Ordered make check перед PR.
- **Review:** `review_mode=existing_session`; reviewer `courses-96 [557618]` (власник підтвердив ref 557618 після /clear, попри що назва сесії тепер «courses-96», не «Courses 2»); writer = Courses13 [c49267]; contact = SendMessage; обов'язкове незалежне рев'ю перед merge.
- **Transfer (SA13-B review):** snapshot = commit `7442749` на гілці `sa13-totp-replay`. Стан: `verdict_received` — **PASS, 0 blocking** від `courses-96 [557618]` (2026-10-29), verdict стоїть на прочитанні коду (усі 5 фокусів), не лише на зеленому гейті. HEAD=7442749, дерево чисте → рев'ю коміту = те, що змерджиться. Non-blocking: N1 (розподіл покриття e2e-live/unit-replay нормальний), N2 (skew +1-крок DoS вимагає повного компромісу секрету, самолікується ≤30с — не новий вектор), N3 (Acc26 2.4m vs 240s ~40% запас — перший підозрюваний за флаку під навантаженням). Required final verdict отримано. Наступне: push + PR + self-merge.

## Рішення (власник, 2026-10-29)
- **Scope:** консюмити counter на **ВСІХ 3 шляхах** (login + disable/regen + enrollment-confirm) через спільний `AdvanceTotpCounter`. Закриває: «лише логін». Перевірка: integration покриває всі 3 шляхи на replay.
- **Sequencing:** **чекати merge PR #83**, потім починати I-F5-код (одна гілка в польоті). Закриває: стек на D-F2 / паралель від main. Перевірка: I-F5-гілка відгалужується від main уже з D-F2 всередині.

## BLOCKER (виявлено 2026-10-29 під час гейту) — e2e реюз коду
- Код+unit+integration готові й зелені: domain `VerifyTOTPAt`, port+repo `AdvanceTotpCounter` (атомарний conditional-UPDATE), 3 шляхи wired, міграція 00023, sqlc regen. Unit `TestCompleteTwoFactorRejectsReplayedTOTP` (replay на логіні → ErrInvalidTOTP), integration `AdvanceTotpCounter ... (I-F5)` (replay/stale/newer + concurrent 8 racers = 1 winner). gofmt/vet/`go test -race ./...`/identity-integration зелені.
- **Конфлікт:** `frontend/e2e/accounts.spec.ts` Acc26–Acc31 реюзить той самий `totp(secret)` багато разів за секунди (confirm@406, login@429/447/462, regenerate@452, disable@467). Сервер валідує за реальним годинником (±1 крок ≈90с) — кілька дій у тому самому 30-с кроці тепер відхиляються. enrollment-consume (вибір «усі 3 шляхи») ламає @429 майже завжди (той самий крок, що @406). Навіть login-only ламало б @447/@462, коли діляться кроком.
- **Причина глибша:** тест кодує стару (вразливу) поведінку «код можна вжити кілька разів у вікні». Фікс правильний — тест застарів.
- **Опції (owner call):** (A) лишити сильний фікс (3 шляхи) + переписати e2e: wait-to-next-step хелпер там, де крок колайдить (+до ~30с/колізія, ~1–2хв до прогону) та/або recovery-коди для повторних логінів; (B) звузити до login-only — усе одно треба правити e2e; (C) прибрати enrollment-consume (дозволити confirm→перший-логін тим самим кодом) — менше e2e-правок, але слабший гейт. **Рекомендація: A** (фікс коректний; тест має вживати кожен крок один раз).
- **Стан:** backend-гілка `sa13-totp-replay`, НЕ комічено; повний `make check`/e2e НЕ ганявся (зламався б на Acc26). Чекаю вибір перед правкою e2e.

## Рішення (власник, 2026-10-29, продовження після /clear)
- **Вибір A:** лишити сильний фікс (consume на всіх 3 шляхах) + переписати e2e так, щоб кожен успішний TOTP-проф вживав СВІЙ 30-с крок. Закриває: тест кодував стару (вразливу) поведінку «код багаторазово у вікні». Перевірка: повний `make e2e` зелений на Acc26 без послаблення фіксу.
- **Реалізація e2e (факт коду):** сервер на disable ресетить `totp_last_counter=0` (`DisableUserTotp`), а re-enable з новим секретом стартує з 0 → повторний enroll (accounts.spec L475) spacing НЕ потребує. Строго-зростаючі кроки потрібні лише для 6 проофів на `secret` (L406 confirm, L429/L447/L462 login, L452 regen, L467 disable). Хелпер `freshTotp` чекає до наступного невжитого кроку; трекер скидається після disable (дзеркалить серверний ресет). Таймаут Acc26 піднятий під найгірший spacing (6×~30с).

## Журнал
- 2026-10-29: investigation I-F5 завершено (3 шляхи, storage на users, атомарний conditional-UPDATE проти гонки, міграція 00023, check-generated стає в гру). Код НЕ розпочато — план-гейт.
- 2026-10-29 (після /clear): власник обрав A. Правлю `frontend/e2e/accounts.spec.ts` — `freshTotp` spacing-хелпер + reset після disable + підняти таймаут Acc26.
