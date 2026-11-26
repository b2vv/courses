NEXT: Read the shared entry rules, then verify the repository root and load the local rules.
FACT: Read the shared entry rules and verified the root at /Users/strelia/projects/courses.
NEXT: Read the local rules and briefing, then the named plan-defense contract.
FACT: Read the local rules, briefing context and five-question contract.
DECISION: Put the defense table in the working file under the owner’s read-only instruction.
NEXT: Pin revision 6, locate §18 and the task record, then read earlier defense findings.
FACT: cf5c5477 is on main and the working plan matches it. DP2 points to the design-pass plan. The previous defenses consistently left dependency and race gaps after revision 1.
NEXT: Read the pinned ACCT plan and its audit, then compare the revision 5 to revision 6 delta.
FACT: Revision 6 adds no-mail and password recovery, revises 403 behavior and resend copy, and adds session-read ordering and four-action 2FA conflict handling.
NEXT: Read the exact V1–V17 rules and the audit in bounded sections.
FACT: The audit used real invalid-link and guest-resend responses and mocked success and 2FA. V3 newly promises a post-result session read after every E2–E5 outcome.
NEXT: Trace the routes, session query and adapter, and exact revision 6 setting rules.
FACT: All three resend entry points and the settings mutations are independently wired today; requireIdentity uses fetchQuery on the shared session key.
NEXT: Verify root session ownership, Query cancellation semantics, backend write order, and the named UI and e2e seams.
FACT: Confirmed the audited UI seams and the existing Mailpit and ten-code e2e assertions. The V5a conflict and cached-null paths now name mechanisms, while the earlier event-time 2FA guard finding remains open.
NEXT: Read precise Query conditions and backend branches to test the new A6 refresh and V5a success paths.
FACT: Installed Query only cancels a running invalidate/refetch when cached data exists. A6 newly relies on that option alone and ACCT-6 exercises only a cached-address case. Server writes confirm token consumption precedes the account update and 2FA replacement spans multiple operations.
NEXT: Check whether a data-less session is reachable and whether the plan handles a failed refreshed read or concurrent email-change requests.
FACT: The data-less A6 read is reachable after root identity resolves unavailable. The new password advice also leaves the desired next password unspecified, while the API rejects equal current and next values before credential verification.
NEXT: Recheck earlier dispositions and pinpoint the A6 failure and overlap cases to score.
FACT: Confirmed the password validation order, the data-less A6 scenario, and the production bounds for the secret and codes. The A6 route guard uses the shared query, and the root error view supplies a general retry if the settings guard fails. All reviewed plan, audit and named source files match the pinned revision.
DECISION: Count open questions by row. Rows 2 and 3 remain open. Keep their causes separate below.
NEXT: Record every prior finding's disposition and the final defense table.

### Earlier findings

FACT: Revision 1. V1/V6 specify guest validation and localized feedback (619–621, 633, 645, 678). V1 names a same-router resend owner and V14 offers manual copy after clipboard failure (623, 697). V2/V3 specify recovery after an ambiguous token response (634–636, 645–648), subject to the new A6 refresh gaps below. Rollback names the effects that survive (754).

FACT: Revision 2. The 204 copy has no delivery promise (616). The failed 2FA write plus failed session read preserves the form and gives a read-only retry (667–671). Same-render resend is guarded and tested (623, 730). A3 cancels a pre-result /me and starts another (636, 733). V4a and ACCT-7 cover StrictMode and a repeated opening while the client promise is pending (651, 735).

FACT: Revision 3. V1 gives a no-letter action and removes the unsupported claim about the validity or arrival order of links across tabs (616, 623). The admin reset path covers a user with neither authenticator nor usable recovery code (676). V4a explicitly bounds settled attempts to the mounted instance and covers pending A→B→A, so settled-mutation retention is no longer required (651–652, 735). The stale A3 reconciliation exclusion carries forward (636, 651, 735). A6's new refresh depends on a separate ordering condition below.

FACT: Revision 4. A3 repeat goes through the original S3/S4 predicate after the POST (632–636, 733). An in-flight remount joins the same promise (651, 735). Regeneration 401 now drives the redirect hook, and all four 2FA actions use the 409 reconciliation (673–676, 737). The previous resend ordering claim is gone (616, 623).

FACT: Revision 5. V1 closes the cross-tab copy finding (616, 623). Settings 401 cancels the held session read before publishing null (674, 737). Enroll 409 now reconciles along with confirm/disable/regenerate (673, 737). S5/E5 both suppress same-screen retry on origin_forbidden (634, 646, 733–734). The email-change 204 now has a no-mail recovery action (689). Password recovery has added copy but still does not give an executable next step, and settings 2FA write overlap remains unspecified.

### Remaining and new findings

FACT: Password recovery after an ambiguous write (dependency). V10 says to enter the previous proposed new password as the current password after a retry returns incorrect current password (689, 737). The form keeps the proposed new value after errors (689, current AccountSettingsView.tsx:83–99). Following the instruction alone submits the same value twice and returns password_unchanged before verifying credentials (account_policy.go:84–91, accounts.go:199–200, handler.go:839–840). The case is reachable when the password update commits and deletion of other sessions fails (accounts.go:219–222). Specify how to check or proceed with the applied password and what remains uncertain about other sessions. The route test at 737 checks the hint, not the resulting next action.

FACT: A6 refreshed read can fail (dependency). V3 says E4/E5 lead to settings so the person can tell whether the earlier write applied, and ACCT-13 promises an address from a post-result read (645–648, 717). The refresh at 648 can fail. invalidateQueries normally swallows refetch errors (Query core queryClient.ts:328–343). The existing protected settings guard retries the read on entry, but on failure it raises a generic route error (require-identity.ts:11–26, __root.tsx:108–122). V9 only says to omit the row if the session is unreadable (687). The plan does not say whether E2–E5 remain actionable when this read fails, how the address can be checked without re-posting the token, or when the generic settings retry is the intended recovery. Name the failure behavior and cover it in ACCT-6/13 (734, 739).

FACT: A6 can join a pre-result session read (race, new in r6). V3's invalidateQueries with cancelRefetch does not cancel an initial /me with undefined data, even with an active observer. Query reuses that promise (Query core query.ts:401–415). The initial root reads can fail with no data and still render the layout, which observes the shared key and can start a new /me during the A6 POST (root-identity.ts:21–33, __root.tsx:173–176, identity-queries.ts:19–37). If that /me started before the email write and remains pending after the A6 result, settings fetchQuery can join it too (require-identity.ts:11–17), contrary to 648/717. ACCT-6 only holds a read in a cached-address scenario (734). Define a read that starts after the result even without cached data and retain V4a's mounted-screen ownership when canceling.

FACT: Settings 2FA writes are still unsynchronized at dispatch (race, carried from r5). V5a blocks new writes once reconciliation starts, specifies the 409 outcome, and distinguishes session expiry, but gives no event-time rule for starting two writes or explicit scope boundary (665–676, 737). The current settings callbacks call separate mutations directly and use the last rendered pending state (settings.account.tsx:31–64, 83–86, AccountSettingsView.tsx:48–61, 234). Two enroll requests can both succeed before activation, with the response observed last containing a secret the other request replaced. Two regeneration requests can overlap deletion and insertion with distinct valid proofs (two_factor.go:58–81, 227–241). Define the same-router ordering and bound what is guaranteed across tabs.

FACT: Read-only source defense of cf5c5477 on main. No Make, Docker or e2e ran. The audit itself used a real invalid-link response and one real guest 400, with success and 2FA states mostly mocked (audit 9–11, 188–199). This working file is the only edit.
NEXT: Return the five-row table and verdict.

| # | Question | §18 at cf5c5477 | Verdict |
|---|---|---|---|
| 1 | Malformed input | 619–621, 633, 645, 678. Missing or rejected links get S4/E4. Required, malformed and overlength guest addresses are blocked with localized field feedback. Resend 400 gets a general alert. | Closed |
| 2 | Dependency failure | 616–619, 631–648, 667–676, 687–689, 697. Resend, 403, 2FA proof loss and clipboard failure have recovery. The new password advice can end at password_unchanged. A6 adds a required fresh session read without defining its failure path through to the settings address. | **Gap** |
| 3 | Race and repeated invocation | 623, 636, 648, 651–652, 665–676, 730, 734–737. Resend, link remounts, four-action 2FA conflicts and stale /me after settings 401 are covered. A6 can reuse an initial /me that predates the result, and settings 2FA writes still lack dispatch ordering. | **Gap** |
| 4 | Empty and boundary data | 621, 633, 645, 687–699, 718, 722–723, 742. The plan covers absent tokens and session values, cleared forms, the one-time recovery-code display, and full secret readability in the narrow and enlarged-text viewports. | Closed |
| 5 | Rollback | 754. Revert the frontend commits. Sent mail, invalidated links, used tokens, changed credentials and applied 2FA changes survive. | Closed |

**Verdict: 2 gaps (rows 2 and 3).**
