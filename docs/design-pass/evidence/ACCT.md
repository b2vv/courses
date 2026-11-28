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

## Code review

Codex, R1 on `58622b51..78a572b1`: **BLOCK** (`_wip-review-acct-code.md`). All five findings were confirmed and fixed:

- **P2: no session read after S5.** V2 asks for one read after every result but 403; the code read only after S2/S4. A lost 204 behind a timeout left the cached viewer unverified and the banner on screen. The read now follows S5 too (not 403). Route test: a 503 whose write went through hides the banner; it fails with the old trigger.
- **P2: a resend result crossed accounts.** The banner's resend hook lived in the root layout, which survives sign-out and sign-in. The banner is now its own component keyed by the viewer id. Route test: account A's «Запит прийнято» is gone for account B; it fails with a constant key.
- **P2: two retry presses before a render.** Both token routes now key their mutation by token and refuse a press at dispatch while that token's request is in flight. Route test: two presses in one `act` send one request on each route; both fail without the guard.
- **P3: the E2 held-read test passed without the refresh.** It now asserts that exactly one `/me` read started after the 204. Added an E0 offline route test and a V6 length-message test.
- **P3: the e2e «wrong code» could be valid.** It now picks a code that is none of the secret's codes for the five steps around now.

Codex, R2 on `78a572b1..6fb97644`: **BLOCK** (`_wip-review-acct-code-r2.md`). Round 1's three P2 closed. New:

- **P2: the token-keyed guard counted another screen's request.** Leave a token route while its request is pending and reopen the same token: when the new attempt fails first, its retry button did nothing until the old request settled. The guard is now this screen's own `ref`, set at dispatch and cleared in that call's `onSettled`. Route test: held first request, reopen, 503, retry → 204; it fails with the R1 route file.
- **P3: the keyed banner lost its result on a visit to the admin surface.** The resend hook is back in the root layout. A result is shown only to the viewer who pressed (the hook records the viewer id at dispatch), so the account switch stays covered. Route test: unverified admin resends, opens `/admin`, returns, and the result is still there; it fails with the R1 root.
- **P3: `tooLong` had no test.** Added; it fails when only `tooLong` is dropped.

Codex, R3 on `6fb97644..1b1d70a7`: **BLOCK**, three P3 (`_wip-review-acct-code-r3.md`); all round-2 findings closed. Fixed:

- **A dismissed banner stayed dismissed for the next account** (A → B without a sign-out between). The banner is keyed by the viewer id; the resend hook stays in the root.
- **«Сеанс завершився» came back** after the same account signed out and in again. Any change of viewer, sign-out included, now drops the recorded press.
- **The overlap test covered `/verify-email` only.** It now runs on both token routes; the `/email/change/confirm` row fails with the R1 route file.

Each new test fails on the R2 code.

Codex, R4 on `1b1d70a7..fb092d77`: **PASS**, no P1/P2/P3 (`_wip-review-acct-code-r4.md`). All round-3 findings closed; for each, the reviewer traced that reverting the fix fails its test. Л1–Л6 not worsened.

The decision «A3 reads only after S2/S4» in the first version of this record was wrong against V2: the plan was right.

## Found and fixed during verification (not in the plan)

- **axe `role-img-alt` after a 2FA action (e2e Acc26).** A focused React Aria `Button` that turns `isPending` announces itself through the live announcer with `aria-labelledby=<its id>`, and the announcer keeps that node for 7 s. «Увімкнути 2FA», «Підтвердити 2FA», the two forms' buttons and «Спробувати ще раз» are unmounted by their own success, so the node pointed at nothing. Those buttons now use `aria-disabled` and the dispatch guards, as `CourseListView` does. `isPending` stays only on buttons that remain on screen (resend, password, email). The e2e `axeClean` now prints the node target.
- **The group headings in the enabled 2FA card read as body text** (h3 at body weight). They are now bold, between the card's h2 and its text.
- **V6 replaced the text of a `validate` function and of form-level errors** with «Перевірте значення.». Found by the admin course editor's unit tests (`CourseEditorView`, «Use at most 120 characters»). A `customError` now keeps its own text. New test in `controls.test.tsx`; with the fix removed it fails.
- **Form clearing used `setState` in an effect** (lint `set-state-in-effect`). It is now an adjustment during render, keyed by the success timestamp.

## Checks

| Check | Result |
|---|---|
| `npm run format:check`, `lint`, `typecheck` | PASS |
| Frontend unit (after R3 fixes) | 119 files, 1658/1658 at load average ~10 |
| Frontend unit (after R2 fixes) | 119 files, 1655/1655 at load average ~12 |
| Frontend unit (after R1 fixes) | 119 files, 1652/1652 at load average ~15 |
| Frontend unit (first version) | Before the last three fixes: 119 files, 1624/1624. After them, the full run under a host load average of 15–45 (Kiri VM ~224 % CPU from other work) had 9–11 failures in admin/media/route files: timeouts plus two `findBy` misses. The same four files run alone: 224/227, then 227/227. The ordered `make check` is the gate for this |
| e2e `accounts.spec.ts` (targeted, Kiri) | 8/8, B19 in budget, also after the R1, R2 and R3 fixes (runs 4–6; run 3 8/8; run 1 failed on the axe finding above, run 2 printed its node) |
| Full ordered `make check` on `cf3e7cf0` | **Red three times**, each time one playback test only (everything before e2e green: unit 1658/1658; B19 7/7). Run 1: P08 «watched 24.7 s, want > 25». Run 2: audio-tracks S7 «Пауза» not shown within 5 s. See «Gate attempts» below |

## Gate attempts

Both red gates failed in playback only, with the same shape: the video did not start, or play long enough, inside the test's time budget.

- P08 alone, accelerated profile, same commit: 3/3 pass.
- The S7 block alone: `main` 6/6 runs green; the branch 2 of 5, then 2 of 6 runs with a failure. One of those was a 10-minute transcode timeout in the fixture.
- The trace of one branch failure: the second lesson's `POST /api/v1/playback/sessions` left the browser at 12:29:53.85. The API logged it as finished at 12:29:58.84 with `durationMs 3713`, against 1–4 ms for the same call a few seconds earlier, so about 1.3 s passed before the API even got it. The video cannot start before that answer. The slice does not change the backend or this request.
- The Kiri VM at the time: 12 GiB, 98 MiB free and 578 MiB available, no swap. Memory pressure `full avg60` = 7.7 % (all tasks stalled about 8 % of the time). Other projects' containers share the VM (elasticsearch at 88 % CPU and 1.2 GiB, kibana, cassiopeia, factory-storage). This matches the recorded gotcha on Kiri memory stalls.

- Run 3, after the owner had Cassiopeia's 10 containers stopped (VM available memory 578 MiB → 5666 MiB): red again on one playback test, subtitles E4b «en showing cue returns after release» (unit 1658/1658, B19 7/7, 129 passed, 47 did not run). The MS4 block alone on the branch then: 3/3 runs, 45/45.

Each red gate failed on a different playback test, and each of them passes alone.

- The same ordered `make check` on `main` (`58622b51`, without the slice), the same VM, Cassiopeia still stopped: **red too**, on one playback test, `playback.spec.ts` R23 «a complete 100-second lesson…» (7.0 min). Unit 1589/1589, B19 7/7, 167 passed, 9 did not run.

- Runs 4 and 5 on the branch (`5daf6413`, 2026-11-27/28) stopped at `make test`: 11 and 15 unit failures, nearly all 5 s/20 s timeouts, in `CourseEditorView.test.tsx` and `route-guards.test.tsx`. These are the same files that passed 1658/1658 earlier the same day at a host load average of ~10–12. Each run started at a host load average of 36–44: other work on the Mac (another project's Storybook build and headless Chromium, Spotify), not the VM. The owner had cargo's and Cassiopeia's containers stopped, but `make test` runs on the host, so this did not help.

So the playback e2e family fails one test per full run today on this host with and without the slice; the slice's three red gates are that same instability, not a regression the slice adds. This is a conclusion from four gates and the isolated reruns, not a pass: no full ordered gate is green on the slice.

No mechanism links the slice to these failures. The `main`/branch difference is 0/6 against 2/5, a small sample. The two red gates are not counted as a pass.

## Limitations

Л1–Л6 stay as recorded in `plan.md` §18. The server race in `SetUserTotpSecret` (Л5) is a separate backend task that has not been created.
