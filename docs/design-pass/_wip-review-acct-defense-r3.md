NEXT: Read the shared entry rules and verify the repository root.
FACT: Read the shared entry rules and verified the repository root.
NEXT: Read the scoped project rules, briefing, and plan-defense contract.
FACT: Read the project routing and the five-question defense contract.
NEXT: Pin revision 3 and recover the revision 1 and 2 findings.
FACT: Read §18 at efc0bcf9 and the two earlier defense logs.
DECISION: Use the pinned plan line numbers and put the defense table in this working file under the read-only instruction.
NEXT: Read the source audit and trace the specified frontend and backend paths.
FACT: Read the full ACCT decisions, acceptance rows, and source-audit findings.
NEXT: Verify the session, mutation, token and control behavior in the named implementation.
FACT: Traced the named routes, views, controls, CSS, client adapters, API handlers, services, and browser spec by source.
NEXT: Verify candidate gaps around resend recovery, overlapping session reads, and duplicate token submissions.
FACT: Checked the exact mutation cache lifetime, shared session cache, and partial server write paths.
NEXT: Resolve which old findings close and which remaining cases leave their fixed question open.
FACT: Verified client defaults and cancellation behavior, browser test sequencing, and parity with the pinned reviewed sources.
DECISION: Count each fixed question once and record earlier closures separately from revision 3 gaps.
NEXT: Append the findings, five-row defense table and verdict.
FACT: Confirmed the browser defaults leave mutation retention at five minutes and that cancellation rejects waiters with CancelledError.
NEXT: Write the final record with exact plan and source citations.

FACT: Reviewed §18 at efc0bcf9 on main against the source audit and the named frontend, backend and e2e sources. This was source review. No Make, Docker or e2e was run.

FACT: Revision 1 follow-up. Guest validation and localized feedback close row 1 (V1/V6). V14 closes the clipboard failure with manual copy (line 666). V2/V3 give same-token retries an explicit recovery path after token consumption (lines 606–621). V1 assigns resend ownership to a router-scoped synchronous guard and explicitly bounds tab coordination (line 596). The rollback clause closes row 5 (line 723).

FACT: Revision 2 follow-up. V1 removes the unconditional delivery promise (line 589). V5a defines the screen and read-only retry when a 2FA write and /me both fail (lines 639–644). It warns that old codes may no longer work (line 645). The recovery instruction assumes authenticator access, leaving the recovery-code-only case open. The same-render resend gap closes (lines 596, 699). V2 now cancels a pre-attempt /me before starting a new one (lines 609, 702). V4a covers StrictMode and a pending A→B→A return (lines 624, 704). The plan still needs an answer for overlapping reconciliations and the lifetime of settled attempts.

FACT: New dependency gap. V7 expressly allows authorization with a recovery code. A user who lost the authenticator can authorize regeneration that way. The server consumes that proof, deletes the old set, then generates and inserts replacements. If the response is lost or insert fails after deletion, the instruction to use a new authenticator code is impossible for this user (plan lines 645, 649–652; two_factor.go:227–241, 258–281). The plan needs a recovery path or an explicit limit with its consequence. V1's 204 copy also leaves the no-letter case implicit. There is no named next action if the mailer fails after invalidating the old link (plan lines 589–596; accounts.go:133–140; handler.go:373–377).

FACT: New race gaps. V4a promises reusing a settled attempt whenever the token is reopened in the same app. The router has no mutation retention override, and TanStack Query drops an unobserved settled mutation after five minutes. Open A, let it settle, navigate away, wait past collection and reopen A on the same router. The cache is empty and the planned check sends a second automatic POST (plan lines 624, 680, 704; router.tsx:35–37; Query core removable.ts:14–28 and mutation.ts:143–161). State the retention or post-expiry behavior and prove it after a settled result. Separately, a late A reconciliation can cancel the /me B is already awaiting. V4a isolates displayed mutation entries but does not scope or sequence cancellation of the shared session key. Specify ownership or cancellation handling so B does not enter S0 from a client-side cancellation (plan lines 604, 609, 624, 702, 704; identity-queries.ts:19–34; Query core queryClient.ts:283–295 and query.ts:401–415).

FACT: V1 explicitly excludes cross-tab resend coordination, but its 204 copy still asserts that prior links are invalid when a new letter arrives. Even without simultaneous writes, the mailer can deliver two successive messages out of order. With overlapping requests, both invalidations can also finish before either token is issued (plan lines 589, 596; accounts.go:133–140; generated/identity.sql.go:474–480). Keep the copy within the stated guarantee.

DECISION: Count each fixed question once. The remaining gaps are dependency failure and race/repeated invocation.
NEXT: Return the five-row defense and verdict to the plan author.

| # | Question | §18 at efc0bcf9 | Verdict |
|---|---|---|---|
| 1 | Malformed input | Lines 591–594, 606, 618, 647. Invalid guest email is blocked with localized field feedback. Missing and rejected tokens render S4/E4. Server-side resend 400 gets a general alert. | Closed |
| 2 | Dependency failure | Lines 589–592, 604–609, 619–621, 639–652, 666. Session and clipboard failures have visible recovery. There is no explicit next action when resend returns 204 and mail never arrives. Regeneration can discard old codes after accepting a recovery code, yet the prescribed retry assumes access to the authenticator. | **Gap** |
| 3 | Race and repeated invocation | Lines 596, 609, 624, 699, 702, 704. The guard covers same-router resend and the plan covers immediate remount and StrictMode. Settled attempt eviction can trigger another automatic POST. An older token's reconciliation can cancel the current token's shared /me read. V1's link-validity copy also exceeds the cross-tab guarantee. | **Gap** |
| 4 | Empty and boundary data | Lines 594, 604–607, 618, 656–658, 666–668, 692. The plan covers missing tokens and session data, cleared forms, hiding the one-time code list, and displaying the full fixed-size secret at the specified widths. | Closed |
| 5 | Rollback | Line 723. Revert the frontend commits. Sent mail, invalidated and consumed tokens, account changes, and 2FA effects already applied by the API survive the revert. | Closed |

**Verdict: 2 gaps (rows 2 and 3).**
