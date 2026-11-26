# A3, A6, A7 — email confirmation and account settings: ui-audit (before any fix)

| | |
|---|---|
| Date | 2026-11-26 |
| Build | `main` @ `c9ca9616` (#123). The running frontend image was built 2026-11-23 14:55 UTC from the C3b branch. Since then, #123 changed `main.css` (`.status-page` only) and `presentation/ui/states.tsx` (retry link); none of the files of these three screens. |
| Screens | A3 `/verify-email`, A6 `/email/change/confirm`, A7 `/settings/account` (plan §3). A1/A2/A4/A5 were audited and fixed in the AUTH slice (#115). |
| Role · task | A3: guest who opened the link from the verification email. Task (inferred): **"confirm my email and carry on"**. A6: signed-in student who opened the link sent to the new address. Task (inferred): **"finish changing my email"**. A7: signed-in student. Task (inferred): **"change my password or email, or turn two-factor sign-in on or off"**. |
| States | A3: no token, invalid token (real server answer 410), checking, success, resend pressed (real guest answer 400; mocked 204 and 429), signed-in student. A6: no token, invalid token (real 410) as student and as guest, success as student and as guest. A7: default; empty submit; password wrong / saved / saving; email taken / sent; 2FA setup, setup with a wrong code, recovery codes, 2FA on, turning off with a wrong code; guest (goes to sign-in). Light and dark, EN. |
| Viewports | 390×844, 768×1024, 1024×768, 1440×900, 390 with 200% text (simulated), 320×720 for A3 invalid/no token/success, A6 invalid/success, A7 default/setup/recovery codes/2FA on. Other states at 390 and 1440. |
| Data | Read-only. Success, 2FA and server-error states are the **real UI over a mocked API answer** (Playwright `route`); no live account was changed and no email was sent. The real server answered only the invalid-token reads and one guest resend (400, nothing sent). The stored e2e sessions had expired on 2026-11-24, so the test student signed in once through the sign-in page into a separate scratch session; `playwright/.auth` was not touched. |

Boxes are CSS px, `x y w×h` from the top of the document, at 1440 unless stated. Captures and metrics are in the session scratchpad (`a37/shots`, `a37/cap.json`, `a37/kb.json`) and are not committed.

## Problems

**#1 · Task · «Надіслати лист знову» never says what happened**
- sees: after pressing «Надіслати лист знову» the screen stays exactly the same, whether the email was sent, the server refused, or there were too many tries.
- costs: the person cannot tell whether to check the inbox or try again, so they press again or give up.
- proof: A3 at 1440, button `144 447 199×44`. Mocked 204, mocked 429 `rate_limited` and the real guest answer 400: the same text «Посилання недійсне або вже використане.», no `output`, no `role=alert`. Code: `routes/verify-email.tsx` passes `resend.error` only as `errorCode`, which the view turns into the same «expired» line, and never passes `resend.isSuccess`.
- rule: [nielsen] #1 Visibility of system status; [state] Error states actionable.
- fix: a sent message («Лист надіслано на вашу пошту») and named errors (too many tries, sign in first).
- sure: yes.

**#2 · Task · a guest is offered a resend that cannot work**
- sees: a guest on an invalid or used link gets the main button «Надіслати лист знову».
- costs: the only action on the page fails for the visitor most likely to be there (the link opens in a browser that is not signed in).
- proof: real guest press → `POST /api/v1/auth/resend-verification` **400**; the screen did not change (see #1). The page has no way to sign in except the header «Увійти».
- rule: [nielsen] #5 Error prevention; [state] Error states actionable.
- fix: for a guest, «Увійти, щоб надіслати лист знову» (sign in with a return to this page); for a signed-in person, the resend.
- sure: yes.

**#3 · Task · the email is confirmed, and then nothing**
- sees: «Пошту підтверджено. Дякуємо. Ваш обліковий запис підтверджено.» and no button or link.
- costs: a dead end at the moment the person wants to start: a guest has to find «Увійти» in the header, a student has to find «Моє навчання».
- proof: A3 success (mocked 204), `main` has 0 actions at every viewport.
- rule: [nielsen] #3 User control and freedom; [state] Success states lead somewhere.
- fix: a guest gets «Увійти»; a signed-in student gets «До навчання» (or «До каталогу»).
- sure: yes.

**#4 · Task · a used or invalid email-change link is a dead end**
- sees: «Підтверджуємо нову пошту / Посилання недійсне або вже використане.» and nothing to press.
- costs: the person cannot ask for a new link from here and is not told where to do it.
- proof: A6 invalid token (real 410) as student and as guest, `main` has 0 actions.
- rule: [state] Error states actionable.
- fix: «Перейти до налаштувань акаунта» (where the change is requested again); for a guest, sign in first.
- sure: yes.

**#5 · Words · a failed link keeps the "in progress" title**
- sees: when the link is invalid, A3 still says «Підтверджуємо пошту» and A6 «Підтверджуємо нову пошту», as if checking were still going on.
- costs: the heading says one thing and the line below another; the reader of the heading alone (and a screen reader user jumping by headings) is told the check is running.
- proof: A3 and A6 invalid token, `h1` above. Its twin A5 (`/password/reset` without a token) says «Посилання не працює» since AUTH-4.
- rule: [consistency] Same role, same treatment; [copy] Headings name the state.
- fix: the failure state gets its own title, like A5 («Посилання не працює»).
- sure: yes.

**#6 · Task · after «Увімкнути 2FA» the keyboard focus is lost**
- sees: pressing «Увімкнути 2FA» with the keyboard replaces the button with the QR code and the form, and the focus goes nowhere.
- costs: a keyboard or screen-reader user is thrown to the start of the page and must Tab through the header and two forms to reach the code field.
- proof: keyboard at 1440: focus on «Увімкнути 2FA», Enter → `document.activeElement` is `BODY` (`a37/kb.json`, `a7-enroll.afterAct`).
- rule: [wcag] 2.4.3 Focus Order.
- fix: move focus to the setup text or the code field when the setup appears; the same on the move to recovery codes and back.
- sure: yes.

**#7 · Words · empty-field messages appear in English**
- sees: pressing «Змінити пароль» with empty fields shows «Please fill out this field.» under each field, on a Ukrainian page.
- costs: the only explanation of what went wrong is in another language.
- proof: A7 empty submit, two `.ui-field-error` with that text (390 and 1440). The browser's own required-field message is shown.
- rule: [wcag] 3.1.2 Language of Parts; [copy] One language per page.
- fix: own messages («Введіть поточний пароль») from the locale; check the same for email and 2FA fields.
- sure: yes.

**#8 · Groups · with 2FA on, two forms share one card and one label**
- sees: the 2FA card has «Поточний пароль», «Код автентифікатора або резервний код», a red «Вимкнути 2FA», then again «Код автентифікатора або резервний код» and «Перегенерувати резервні коди», with no subheading or text between them.
- costs: it is unclear that the second code field belongs to the regeneration and not to turning 2FA off; the two code fields have the same accessible name.
- proof: A7 2FA on (mocked `/me`), fields `161 1215 638×44` and `161 1361 638×44`, both labelled «Код автентифікатора або резервний код».
- rule: [law] Common Region; [wcag] 2.4.6 Headings and Labels.
- fix: two separated groups with their own small heading or one line of explanation («Вимкнути 2FA», «Нові резервні коди»), the destructive one last.
- sure: yes.

**#9 · Task · once 2FA setup has started there is no way back**
- sees: after «Увімкнути 2FA» the card shows the QR code, the secret and «Підтвердити 2FA»; there is no «Скасувати».
- costs: a person who started by mistake, or has no authenticator app at hand, can only leave the page.
- proof: A7 setup (mocked enroll), `main` actions: secret, code, «Підтвердити 2FA».
- rule: [nielsen] #3 User control and freedom.
- fix: «Скасувати» returns the card to «Захист 2FA ще не ввімкнено».
- sure: yes.

**#10 · Words/Task · the email card does not say what the current email is**
- sees: «Змінити електронну пошту» asks for «Нова електронна пошта» but nowhere shows the current one.
- costs: the person cannot check which address they have now or whether the change went through after confirming.
- proof: A7 default, the card holds one field and the button. The current address is in the session (`/api/v1/auth/me`), not on the page.
- rule: [nielsen] #6 Recognition rather than recall.
- fix: «Поточна пошта: …» above the field.
- sure: yes.

**#11 · Task · after a saved password the typed passwords stay in the fields**
- sees: «Пароль змінено. Інші сеанси завершено.» appears, and the old and new passwords stay in the two fields; the success line looks like ordinary text.
- costs: the form looks unsent, inviting a second press, and the passwords sit in the page.
- proof: A7 password saved (mocked 204), 390: both fields filled, `output` in body style.
- rule: [state] Success state is distinct; [forms] Clear sensitive fields after submit.
- fix: clear both fields after success and make the message look like a success, not body text; the same for the email form.
- sure: yes.

**#12 · Twins/Squint · three equal full-width primary buttons**
- sees: «Змінити пароль», «Змінити пошту», «Увімкнути 2FA» are each a full-width filled button of the same weight, one per card.
- costs: nothing on the page stands out, and the buttons look unlike every other account form, where the button keeps its own width at the start (AUTH slice, `.notice form`).
- proof: A7 at 1440: all three `161 … 638×44`, same fill; login «Увійти» `144 641 78×44`.
- rule: [consistency] Same role, same treatment.
- fix: the same button width rule as the other account forms.
- sure: yes.

**#13 · Stress · at 200% the A3 title breaks mid-word**
- sees: «Підтвердж / уємо пошту» at 390 with 200% text.
- costs: a broken word for exactly the readers who enlarged the text.
- proof: `a3-badtoken-p200`. The status pages got a fix for this in C3b (`.status-page` container rule), the `.notice` pages did not.
- rule: [wcag] 1.4.4 Resize Text (readability, not loss of content).
- fix: the same rule for `.notice` titles.
- sure: yes.

**#14 · Lines · the resend button sticks to the line above**
- sees: «Надіслати лист знову» touches the text above it (about 4 px), while on A5 the action sits 16 px below its text.
- costs: the button reads as part of the sentence.
- proof: A3 1440: text bottom ≈443, button top 447; A5 text bottom 472, button top 488.
- rule: [law] Proximity.
- fix: the same gap as A5 (`.notice` actions).
- sure: yes.

## Suspects

**#15 · Task · recovery codes have no "done" step**
- sees: after confirming 2FA the card shows the ten codes and «Скопіювати коди»; there is no «Я зберіг коди» / «Готово», and the line «2FA увімкнено» is not shown while the codes are on screen.
- costs: the person does not know when it is safe to leave, or that 2FA is now on.
- proof: A7 recovery codes (mocked), 390 and 1440.
- rule: [nielsen] #1 Visibility of system status.
- fix: «2FA увімкнено» above the codes and «Готово» that closes them.
- sure: suspect (the codes are meant to be shown once; leaving them until reload may be intended).

**#16 · Stress · the manual secret does not fit the field at phone width**
- sees: «Секрет для ручного налаштування» shows «JBSWY3DPEHPK3PXPJBSWY3DPEHPK» and cuts the rest.
- costs: a person typing it by hand cannot read the whole secret; there is no copy button.
- proof: A7 setup at 390, 320 and 390/200%, `.ui-input` clipped (`scrollWidth > clientWidth`).
- rule: [forms] Read-only values readable in full.
- fix: show it as wrapping text in groups of four, or add «Скопіювати».
- sure: suspect (the field can be scrolled and selected).

## Taste

- A6 is labelled «Акаунт» and A3 «Підтвердження пошти»; both confirm an address from an email.

## Action inventory

A3, guest, 1440 (390 the same, with «Меню»):

```
header (skip link, wordmark, «Каталог», «Увійти», UA/EN) · checked in G0 · Tab PASS
«Надіслати лист знову» · resend the verification email · own label, under the failure line · 199×44 · Tab PASS, focus 3px · real guest → 400, no feedback (#1, #2); mocked 204/429 → no feedback
footer · checked in G0 · Tab PASS
```

A3 success, A3 checking, A6 invalid / no token: no action in `main` (#3, #4).

A6 success, student and guest, 1440:

```
«Налаштування акаунта» · go to account settings · the only action under «Пошту змінено» · 180×44 · Tab PASS, focus 3px · → /settings/account (a guest is sent to sign in)
```

A7, student, 1440:

```
header («Моє навчання», «Акаунт», «Вийти», UA/EN) · checked in G0 · Tab PASS
«Поточний пароль», «Новий пароль» · fields of «Змінити пароль» · card heading · Tab PASS
«Змінити пароль» · change the password · card «Змінити пароль» · 638×44 · Tab PASS · empty → English browser messages (#7); mocked 401 → «Поточний пароль неправильний.» role=alert; mocked 204 → message, fields kept (#11); pending → «Зберігаємо…», disabled. Real press: NOT_TESTED (unsafe)
«Нова електронна пошта» · field · card heading · Tab PASS
«Змінити пошту» · request the change · card «Змінити електронну пошту» · 638×44 · Tab PASS · mocked 409 → «Ця пошта вже використовується.»; mocked 204 → «Підтвердіть зміну на новій адресі.». Real press: NOT_TESTED (unsafe)
«Увімкнути 2FA» · start 2FA setup · card «Двофакторна автентифікація» · 638×44 · Tab PASS · mocked → QR + form, focus lost (#6). Real press: NOT_TESTED (unsafe)
«Секрет для ручного налаштування» · read-only secret · setup text · Tab PASS · clipped at phone (#16)
«Підтвердити 2FA» · confirm setup · under the code field · Tab PASS · mocked 422 → «Код автентифікатора неправильний.»; mocked 200 → recovery codes
«Скопіювати коди» · copy the recovery codes · under the codes · 164×44 · NOT_TESTED (clipboard permission in the test browser)
«Вимкнути 2FA» · turn 2FA off · no own group (#8) · danger · Tab PASS · mocked 422 → «Код автентифікатора неправильний.» above both forms
«Перегенерувати резервні коди» · new recovery codes · no own group (#8) · secondary · Tab PASS · activation NOT_TESTED
footer · checked in G0 · Tab PASS
```

Tab order matches the visual order on every captured state; focus is a visible 3 px outline on every stop.

## Not looked / NOT_TESTED

- **A real verification or email-change token:** NOT_TESTED. A real token needs a fresh email from Mailpit, which means registering or changing an address on the live stack. Success states are the real UI over a mocked 204.
- **A signed-in student whose email is not verified:** NOT_TESTED, no such fixture; the resend for a signed-in person was not pressed (it would send an email).
- **A signed-in verified student on an invalid A3 link:** seen (same screen as the guest, with the resend); what the resend would do for an already verified account is NOT_TESTED (unsafe).
- **2FA turned off successfully, regeneration success, «Скопіювати коди»:** NOT_TESTED.
- **2FA setup in dark (QR code on a dark card):** NOT_TESTED, not captured.
- **Keyboard at phone width:** NOT_TESTED; keyboard was walked at 1440 only.
- **Contrast:** axe (`wcag2a/aa/21aa/22aa`) found 0 violations and 0 contrast `incomplete` items in every A3/A6/A7 capture, light and dark (the A7 guest capture is the sign-in page, audited in AUTH). No `ui-measurements` contrast run (offer below).
- **Horizontal overflow:** 0 in every capture.
- **Real browser zoom:** NOT_TESTED, because 200% is simulated.
- **Screen reader:** NOT_TESTED.

## Owner answers

2026-11-26, owner: «Правити все» — all 16, including suspects #15 and #16 and the taste line, are to be fixed.

| # | Answer | Date |
|---|---|---|
| 1–14 | fix | 2026-11-26 |
| 15, 16 (suspect) | fix | 2026-11-26 |
