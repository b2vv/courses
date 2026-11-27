# ACCT: account pages after the A3–A7 audit

| | |
|---|---|
| Date | 2026-11-27 |
| Branch | `design-pass-acct`, root `58622b51` (plan §18, revision 12, accepted) |
| Contract | `plan.md` §18 «Зріз ACCT»; source: audit `evidence/A3-A7-audit.md` (#1–#16 and the taste line), owner answer «Правити все» (2026-11-26); scope narrowed by the owner's decision of 2026-11-27, with limitations Л1–Л6 |
| Build checked | frontend image `f5260afca8d1` on Kiri with the e2e compose configuration (`environment.sh`) |
| Data | Browser captures: the seed student's stored session (read-only) plus guests. Every write in the captures is a mocked response (`page.route`), so no live account changed. The e2e uses its own registered account, which the spec deletes in `afterAll` |

## Result against acceptance

| Item | Evidence | Status |
|---|---|---|
| ACCT-1 | Route test: two surfaces (banner and page) are pressed before a render, and they send one POST. Without the guard the same test fails with «expected 2 to be 1» (mutation check). The result shows only under the surface that was pressed. Unit: the four outcomes (204/429/401/other). Browser: banner result at 6 viewports | PASS |
| ACCT-2 | Route: tokenless entry sends no verify POST; a guest sends `{ email }`; a signed-in viewer sends `{}`; an empty field sends nothing («Заповніть це поле.»). e2e «Лист і посилання»: a guest resend gets a new Mailpit message (a different id). The first link then answers `410 token_used` → S4; the new link → S2 «До навчання»; the first link again → S3 | PASS |
| ACCT-3 | e2e: a context with an unknown session cookie → guest form → «Сеанс завершився.» with «Увійти». A copy of the registered page's state shows the banner. Its cookie is replaced without a reload; a real resend → `401 authentication_required` → «Увійти» → full load → password field | PASS |
| ACCT-4 | Route + e2e: S2 «До навчання» for a viewer, «Увійти» for a guest | PASS |
| ACCT-5 | Route: S0 paused offline with no POST, then one POST online. Read paused after 204 → no action and «Чекаємо на з'єднання…». 204 + failed read → «Увійти». S3/S4 + failed read → «Оновити сторінку». S5 retry → `token_used` → S4. 403 → «На головну», no retry. Л4 (unverified viewer, 204) → S2. A background session refresh keeps the S4 field, its value and its focus | PASS |
| ACCT-6 | Route: 204/409/410 in StrictMode, one POST each; 403 without a retry. E2 and its action show while `/me` is still held. e2e: the spent change link → E4 → «Перейти до налаштувань акаунта» → «Поточна пошта: <нова>» | PASS |
| ACCT-7 | Route: `?token=one` → `?token=two` on the same route gives two POSTs and the second screen's own result | PASS |
| ACCT-8 | e2e `toBeFocused()`: setup text after «Увімкнути 2FA»; the status after «Підтвердити 2FA» and after «Готово»; the status after «Вимкнути 2FA»; Enter on «Підтвердити 2FA» with a wrong code keeps the button focused. Keyboard probe (1440): cancel → «Увімкнути 2FA»; regenerate → «Нові резервні коди» → «Готово» → status | PASS |
| ACCT-9 | Route: uncertain confirm (503) → «Оновити сторінку», no «Скасувати», and still none after a later 400. Conflicting enroll (409) → reload. Disable 503/409 and regenerate 409 → reload; regenerate 503 → its own text. 401 `recovery_code_invalid`/`invalid_credentials` → error in the form. `authentication_required` on a write → cache `null` → `/login`. Two presses before a render → one password POST and one enroll POST | PASS |
| ACCT-10 | `controls.test.tsx`: `valueMissing` and `typeMismatch` in uk and en, on TextField and TextArea; an explicit `errorMessage` wins. Browser: an empty submit on `/settings/account` and on S4 shows «Заповніть це поле.» | PASS |
| ACCT-11–13 | Two named forms (`getByRole('form', { name })` in e2e and route tests), with the destructive one last; each error under its own form; «Скасувати» returns without a request; «Поточна пошта» shown | PASS |
| ACCT-14 | Pending: fields `readonly`, the button `aria-disabled`, focus stays on it («Зберігаємо…» → «Змінити пароль»). After 204 the fields are empty and the success is an `output.form-success`. axe 0 at all captures | PASS |
| ACCT-15 | The three actions keep their own width at the start of the row (captures at 6 viewports) | PASS |
| ACCT-16 | 390/200 %: `.notice` headings break between words; no clipped text in 158 captures | PASS |
| ACCT-17 | Text → action in `.notice` = 16 px at 100 % (32 px at 200 %, rem). S2 has no text by the plan (Л4), so its gap is the h1's own margin (31–52 px by width): N/A for this row | PASS |
| ACCT-18 | Unit: copy → «Скопійовано»; a refused clipboard → alert; «Готово» drops the block and its copied state | PASS |
| ACCT-19 | The secret field holds the full secret and does not scroll: `scrollHeight = clientHeight`, `scrollWidth = clientWidth` at 6 viewports (390/200 %: 157 = 157, 258 = 258; 320: 58 = 58, 252 = 252) | PASS |
| ACCT-20 | A6 eyebrow «Підтвердження пошти» | PASS |

## Captures

164 captures in 45 scenes (158 in the full run, plus 6 for the banner scene): 6 viewports (390, 768, 1024, 1440, 390 at 200 %, 320) for the main state of each screen, and phone plus desktop for the others. Dark theme and EN are included. axe (`wcag2a/aa/21aa/22aa`): 0 violations, 0 contrast `incomplete`. Horizontal overflow 0, clipped text 0. The scripts and images are kept outside the repository.

## Found and fixed during verification (not in the plan)

- **axe `role-img-alt` after a 2FA action (e2e Acc26).** A focused React Aria `Button` that turns `isPending` announces itself through the live announcer with `aria-labelledby=<its id>`, and the announcer keeps that node for 7 s. «Увімкнути 2FA», «Підтвердити 2FA», the two forms' buttons and «Спробувати ще раз» are unmounted by their own success, so the node pointed at nothing. Those buttons now use `aria-disabled` and the dispatch guards, as `CourseListView` does. `isPending` stays only on buttons that remain on screen (resend, password, email). The e2e `axeClean` now prints the node target.
- **The group headings in the enabled 2FA card read as body text** (h3 at body weight). They are now bold, between the card's h2 and its text.
- **V6 replaced the text of a `validate` function and of form-level errors** with «Перевірте значення.». Found by the admin course editor's unit tests (`CourseEditorView`, «Use at most 120 characters»). A `customError` now keeps its own text. New test in `controls.test.tsx`; with the fix removed it fails.
- **Form clearing used `setState` in an effect** (lint `set-state-in-effect`). It is now an adjustment during render, keyed by the success timestamp.

## Checks

| Check | Result |
|---|---|
| `npm run format:check`, `lint`, `typecheck` | PASS |
| Frontend unit | Before the last three fixes: 119 files, 1624/1624. After them, the full run under a host load average of 15–45 (Kiri VM ~224 % CPU from other work) had 9–11 failures in admin/media/route files: timeouts plus two `findBy` misses. The same four files run alone: 224/227, then 227/227. The ordered `make check` is the gate for this |
| e2e `accounts.spec.ts` (targeted, Kiri) | 8/8, B19 in budget (run 3; run 1 failed on the axe finding above, run 2 printed its node) |
| Full ordered `make check` | NOT_TESTED: needs the owner's consent |

## Limitations

Л1–Л6 stay as recorded in `plan.md` §18. The server race in `SetUserTotpSecret` (Л5) is a separate backend task that has not been created.
