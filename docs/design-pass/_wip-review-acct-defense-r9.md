NEXT: Read the shared entry rules, verify the repository, then load the plan defense contract.
FACT: Git verified /Users/strelia/projects/courses as the repository root. Read the shared entry rules, local scope, briefing, and defense contract.
NEXT: Pin revision 9 and map the prior findings to the five questions.
DECISION: Use the requested working file for the five-row result. Treat Л1–Л6 as accepted scope and use only the defense r8 working file for prior findings.
NEXT: Read the pinned plan and source audit with line numbers.
FACT: HEAD and main are d9136f15 and the requested plan, audit, and source areas have no tracked diff from that commit. The r8 defense had two gap rows and four underlying issues.
NEXT: Trace the named source and compare each r8 issue against the changed plan.
FACT: Audit #1–#16 and the taste item are in scope. Its successes and most write failures used mocked API responses, so those observations are not evidence of server semantics.
NEXT: Verify the revised rules, acceptance rows, and client and server mechanisms.
FACT: Revision 9 explicitly scopes offline writes in Л6, switches A3 action timing to the session observer, adds action-time password/email guards, and explicitly admits simultaneous enrolls can succeed.
NEXT: Check those mechanisms for residual contradictions and trace the transport, backend, and existing e2e paths.
FACT: The named frontend files confirm the existing route, field, button and card seams. ACCT-1 through ACCT-20 map to the audit and taste request.
NEXT: Validate the offline A3 state and concurrent 2FA transitions against exact conditions.
FACT: V2 lines 639–642 solve the prior paused-refresh-promise issue while preserving Л1. V5a provides the missing same-render guards at line 678. The source distinguishes paused transport, no-op 204, and server rejection.
NEXT: Check the boundaries of the new guarantees and the corresponding verification rows.
FACT: The current server performs one-use token consumption and hits named 2FA conflicts only if the corresponding state check observes the changed state. Existing Playwright fixtures use live Mailpit and the real 2FA routes.
NEXT: Compare revision 9 directly with revision 8 and record defensible findings only.
FACT: The revision diff confirms that the affected rules are new in r9: observer timing and failure identification in V2, and the widened offline/cross-tab descriptions in Л5–Л6.
NEXT: Write the closure check and five-row verdict based on those mechanisms.
FACT: Revision 8 closure check. (1) Offline mutations before transport are now admitted by Л6 (729). (2) The early completion of a paused refresh promise is addressed by observing fetchStatus (639–642). (3) Password and email receive action-time same-key guards (678). (4) Л5 now acknowledges that two enrolls may both succeed (728). Its replacement description leaves a separate contradiction below.
FACT: Row 2, new offline outcome mismatch. Л6 (729) promises A3 “Чекаємо на з'єднання…” and describes paused writes using pending buttons. V2 (633, 639–642) applies the waiting copy to the post-result session read only. Opening a token while the app is offline can pause the automatic verification mutation before transport. There is no initiating button, no result to start the session refresh, and S1 specifies only “Перевіряємо посилання…”. The paused /me test at 763 does not exercise this path. The default pause is accepted. The plan must say what the A3 token screen shows during that pause and align it with Л6.
FACT: Row 2, new refresh-failure edge. V2 (634, 639–642) lists sign-in when the session cannot be read in S2, but specifies “Оновити сторінку” on a refresh error. Even with the latter taking precedence, line 641 uses a strict errorUpdatedAt comparison to recognize that error. The query reducer sets errorUpdatedAt with Date.now() and separately increments errorUpdateCount (query-core/src/query.ts:672–682). A refresh that fails in the same clock tick can keep cached data and fall through to a session-dependent action, violating ACCT-5 (739). Identify the refresh failure by an observer result change and specify one S2 action for that failure.
FACT: Row 3, new Л5 mechanism and outcome mismatch. Suppose tab A already has its setup secret. Tab B starts enroll and reads Enabled=false, then tab A confirms successfully before B saves. B's SetTotpSecret still succeeds, replaces the secret, and writes totp_enabled=false. The backend separates the read and write (two_factor.go:58–81, 84–111) and the SQL update has no enabled-state predicate (generated/identity.sql.go:799–806). This produces neither the stated 409 once state changes nor a failed confirmation in A. A has already shown “2FA увімкнено” and recovery codes (V5a/V14, 673 and 712). Keep cross-tab consistency out of scope and correct the stated outcomes in Л5.
DECISION: Count one gap per unresolved fixed question. Static review covered the named sources and audit #1–#16. No make, Docker, or e2e was run. No plan edit or commit was made.
NEXT: Return the five-row verdict. Rows 2 and 3 need plan clarification.

| # | Fixed question | Plan lines and answer | Verdict |
|---|---|---|---|
| 1 | Malformed input | §18:625, 635–636, 650–652, 681–687. Guest email is validated, unusable links get recovery actions, and field errors use the page language. | Closed |
| 2 | Dependency failure | §18:619–623, 633–642, 652, 674–676, 729, 739. The automatic A3 write can pause before the waiting state applies, and the refresh-error check can miss a same-tick failure. | Gap |
| 3 | Race and repeated invocation | §18:627, 657, 678, 728. Same-client guards cover repeated submits. In Л5, an enroll already in flight can disable 2FA after another tab confirms successfully. | Gap |
| 4 | Empty and boundary data | §18:635–636, 651, 681–687, 697, 716, 753, 772. Covers missing tokens and session data, empty and length-invalid fields, and the full secret at narrow widths. | Closed |
| 5 | Rollback | §18:784. Revert frontend commits. Emails, consumed or invalidated tokens, and account changes already made through the API persist. | Closed |

**Verdict** — 2 gaps, rows 2 and 3.
