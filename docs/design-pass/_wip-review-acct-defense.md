
NEXT: Record the review scope and read the shared entry rules.
FACT: Read the shared entry rules and verified the repository root.
NEXT: Read the local repository route, briefing, and defense contract, and locate the pinned section.
FACT: main is at f6a4611a. The defense contract calls for five questions in order and accepts only plan text as a defense.
DECISION: Keep the plan read-only per the owner and put the table in this working file.
NEXT: Read the numbered §18 revision and source audit.
FACT: §18 names resend outcomes, link retry, 2FA cancellation and focus, localized field errors, and responsive limits. The audit used real invalid-link and guest-resend responses, with success and 2FA states mostly mocked.
NEXT: Trace the frontend route, validation, view, styling and existing e2e behavior.
FACT: The three resend entry points have separate mutations today. Both link routes use one idle-triggered mutation per mounted route. Settings has independent mutation state, and clipboard copy can silently return or reject. ACCT-2/6 already call out stale links and ACCT-14/17 specify display boundaries.
NEXT: Read the exact server and adapter failure and write semantics against the five questions.
FACT: Verification and email change consume the token before the account write. Resend invalidates existing tokens before issuing and sending a new one. Its mailer failure becomes false with no error. 2FA enrollment stores a replacement secret and code regeneration performs separate delete and insert operations.
NEXT: Confirm HTTP behavior, field boundaries, and the exact verification fixture outcomes before scoring.
FACT: The handler discards the service's resend boolean and returns 204. It authenticates a presented cookie before using the body value. The client accepts exactly ten recovery codes and the existing e2e asserts that count and reads the full enrollment secret.
NEXT: Verify precise references for the remaining boundary and failure cases and score each row.
FACT: The plan, audit and reviewed sources match f6a4611a. The server generates a 32-character secret and ten distinct recovery codes. No checks requiring Make, Docker, or a browser were run.
DECISION: Count each fixed question once. §18 answers the supported display boundaries. Partial answers leave four rows open.
NEXT: Use this table as the input to the next plan revision.

| # | Question | Plan line at `f6a4611a` and defense | Verdict |
|---|---|---|---|
| 1 | Malformed input | §18 lines 558, 563, 575, 589, 647. V4 covers missing and malformed tokens. V6 localizes field errors when a constraint fires. V1 does not specify the guest field's required, format, or length checks. A nonempty malformed address can reach the neutral 204 path because failed normalization leaves the original identifier for lookup (`handler.go:357–377`, `accounts.go:118–126`, `domain/user.go:69–84`). Specify what prevents that submission and what the guest sees. | **Gap** |
| 2 | Dependency failure | §18 lines 556–559, 562–566, 576, 608. V4 names the link transport failure state and retry, and V1 names errors the API surfaces. The mailer can fail after the old link is invalidated and the HTTP handler still returns 204, which V1 renders as “Лист надіслано” for the signed-in user (`accounts.go:133–140`, `handler.go:373–377`). Use copy consistent with the 204 contract and specify the recovery path. V14 also leaves clipboard refusal undefined before “Готово” hides the only displayed copy (`AccountSettingsView.tsx:63–66`, audit lines 180, 193). | **Gap** |
| 3 | Race and repeated invocation | §18 lines 555–567, 576, 596, 642, 645, 653. V1 suppresses a pending repeat and V8 refuses cancellation during confirmation. The stale-link checks cover sequential reuse. The banner and lesson gate currently have independent mutations (`__root.tsx:236–238,389–394`, `learn.$.tsx:186–188,355–359`). The plan does not say who owns the in-flight resend across those co-visible controls or how concurrent tabs are handled. Rate limiting does not serialize invalidate/issue/send (`accounts.go:109–140`). V4 also retries the same token after an ambiguous failure. Both writes consume it before updating the account, so a retry can show “Посилання не працює” whether the first write finished or failed after consuming it (`accounts.go:95–106,245–265`). Define the intended outcome for overlap and ambiguous retries, or bound each explicitly. | **Gap** |
| 4 | Empty and boundary data | §18 lines 565–566, 575, 589, 598–604, 608–610, 625, 631–634, 647–650. The plan defines missing-token and required-field states, omits the address while the session is unavailable, clears form values after success, and specifies the narrow and enlarged-text display of the full secret. “Готово” defines the state after the codes are hidden. Existing production data has fixed secret and recovery-code sizes (`domain/two_factor.go:16–32,82–99`, `session-schema.ts:19–24`). | **Closed** |
| 5 | Rollback | §18 line 655 gives the frontend revert. Lines 608 and 642 describe one-time display and old-link invalidation separately. The rollback clause does not say which effects survive: reverting the frontend cannot recall mail, restore invalidated links or replaced recovery codes, or undo account changes already applied through the UI (`accounts.go:133–140,195–265`, `two_factor.go:84–111,211–241`). Record these effects alongside the revert. | **Gap** |

**Verdict: 4 gaps (rows 1, 2, 3 and 5).**
