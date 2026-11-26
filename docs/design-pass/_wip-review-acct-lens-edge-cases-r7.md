NEXT: Read the shared agent entry rules, then verify the repository root.
FACT: Shared entry rules read. Repository root verified at /Users/strelia/projects/courses.
NEXT: Read local rules, repository briefing and the nominated skill contract.
FACT: Read the local rules, briefing and plan-critique skill. HEAD is main at ab3cae05.
DECISION: Apply the single edge-cases lens to the pinned plan; the working file is the sole write.
NEXT: Read earlier findings, the pinned plan and owner evidence.
FACT: Indexed the prior notes and DP2. Loaded the shared anti-pattern reference.
NEXT: Compare the revision 7 state tables and verification rows with each earlier finding.
FACT: Enumerated 35 earlier edge-case findings and isolated the revision 7 changes and acceptance matrix.
NEXT: Follow V4b and the affected routes through their actual reader and writer paths.
FACT: V4b introduces wait-then-read and wait-then-null. V4c gives A3/A6 refresh ownership to the link attempt; its stated treatment of definitive refusals must be checked against the tables.
NEXT: Inspect frontend route, session, control and CSS implementations and backend failure boundaries.
FACT: Located the actual route, adapter, UI, CSS, backend and e2e anchors. Existing settings observes shared /me data and disables refetch on mount.
NEXT: Resolve the stale-read publication, timeout/refusal, focus and password-recovery cases against the code.
FACT: LoginView renders only Continue for a signed-in user; a normal second window reuses the original login cookie. The 2FA card derives its enabled branch from the shared query in the current route.
NEXT: Confirm the installed query scheduling and cache overwrite behavior, then compare the precise revision delta.
FACT: The installed query-core can pause a default online query before invoking the transport. V4b also leaves the older query free to publish after an immediate V5a cache update. ACCT-4 newly promises that the banner and gate disappear.
NEXT: Verify both Query Core interleavings in memory and confirm whose account the token handlers write.
FACT: In-memory QueryClient confirmed that a held pre-disable read overwrites the successful cache patch, and that an offline query pauses before the transport starts. The verification POST uses the user ID from the token, independently of the browser session.
NEXT: Confirm router defaults, phase handling and the prior findings’ individual dispositions.
FACT: Main is ab3cae05 and the tracked working tree matches that commit. The router uses default online Query Core scheduling with retries disabled. The audit owner approved all 16 fixes and the label change.
DECISION: Rev 1 findings in file order: (1) closed by neutral 204 copy; (2) closed by stale-cookie 401 sign-in action; (3) closed by dispatch-time shared resend guard; (4) closed for settled token recovery by S3/S4 and E3–E5, subject to the A6 refresh failures below; (5) closed by token-keyed screens; (6) immediate 2FA patch added, but the revision 7 wait rule reopens the stale-state case below; (7) closed by code-set component lifecycle; (8) closed by the standalone-button CSS selector; (9) closed by selecting the second Mailpit ID; (10) closed by the TextField/TextArea scope.
DECISION: Rev 2 findings in file order: (1) closed by conditional 204 copy; (2) closed by synchronous dispatch check; (3) closed for a settling pre-result read by V4b and S0, with paused reads open below; (4) closed by rejoining the token's pending request; (5) closed for a failed read that settles by the locked 2FA form and fresh recheck, with paused reads open below; (6) closed by warning about old codes, fresh proof and admin fallback; (7) closed by read-only fields during writes; (8) closed by content sizing and a full-value visibility check.
DECISION: Rev 3 findings in file order: (1) closed by keyed screens; (2) closed by the documented settled-reopen policy; (3) closed by keeping the form mounted and prescribing the focus return; (4) refusal-specific copy and actions are present, but revision 7 reinstates a session read for A3 refusal via V4c.
DECISION: Rev 4 findings in file order: (1) closed by A→B→A promise joining; (2) closed by a real retry and the same S3 predicate; (3) resolved for a settling pre-401 read by wait-then-null, with paused reads open below; (4) closed by four-action 409 reconciliation; (5) closed by restoring focus after manual recheck.
DECISION: Rev 5 findings in file order: (1) closed for a settling pre-401 read, with paused reads open below; (2) closed by including enroll in 409; (3) closed for a settling initial read by wait-then-fetch; (4) closed by showing the actual response on settled reopen; (5) closed by neutral resend copy with no inbox-order inference.
DECISION: Rev 6 findings in file order: (1) closed for a settling initial read by explicit sequencing and the held-first-read case; (2) closed for successful refresh by composition ownership and the late-A route case, with a failed refresh still open below; (3) identical-password instructions removed, but the replacement can open a login page without a credential form.
NEXT: Record the six actionable findings and verify the sole-file write.

[edge-cases] A paused session query can block recovery indefinitely
where: §18 V4b/V4c, lines 660–661, V5a line 683, ACCT-4/6/9.
why: The router and session query use Query Core's default online scheduling. An offline /me pauses before the transport starts, which an in-memory check reproduced. It cannot reach the adapter timeout. A success or error screen can indefinitely withhold its exit while awaiting that read, and a settings write that already returned authentication_required can indefinitely postpone publishing null and redirecting.
action: Define a bounded or non-pausing recovery path for session reads and a 401 exit that stays authoritative despite old queries. Add cases for current() pausing after a link result and a paused read across a settings 401.
weight: blocking

[edge-cases] A pre-write session read can undo the successful 2FA card change
where: §18 V4b and V5a, lines 660 and 674–675, ACCT-8.
why: V5a immediately patches the cache, but V4b lets an older /me finish and publish. A held pre-disable read can restore twoFactorEnabled=true after successful disable, removing the focused off-state message before the fresh read finishes. The inverse can remove the enabled branch after Done. The installed QueryClient reproduced the overwrite.
action: Preserve both the authoritative card state and its focus across the old read and a failed fresh read. Add coverage with a pre-write /me released after success while the post-write read stays held.
weight: blocking

[edge-cases] A3 refusal regains the session dependency that S5 excludes
where: §18 V2 and V4c, lines 641–643 and 661, ACCT-5 route proof at line 744.
why: V4c starts /me for any A3 completion, while ACCT-5 requires the 403 path to perform no session read. Origin rejection exits before touching the token at handler.go:514–523. If the awaited read is part of the link completion, an unavailable session dependency can delay the only exit from a definitive refusal.
action: List the exact A3 outcomes that trigger shared reconciliation and keep pre-write 403 outside that list. Assert no /me call for a 403 during a session outage and that Home is immediately available.
weight: blocking

[edge-cases] A valid link may verify a different account from the active viewer
where: §18 V2 line 637 and ACCT-4 lines 719 and 744.
why: VerifyEmail writes the user ID from the token (accounts.go:95–106). A viewer signed in as unverified A can open B's valid link and receive 204 while fresh /me correctly still reports A as unverified. ACCT-4 currently requires the viewer's banner and lesson gate to disappear after that 204.
action: Scope banner and gate removal to a fresh session that reports the viewer verified. Keep accurate success copy and a navigation action when a valid link leaves the active viewer unverified. Verify both account combinations.
weight: blocking

[edge-cases] A failed late email refresh can leave settings without a current address
where: §18 V4c and V9, lines 661 and 698, ACCT-6/13.
why: If A commits after settings has mounted and its refresh fails, the guard has already returned and will not do the retry V4c relies on. Settings observes the same query with refetchOnMount disabled. Under V9 it can hide the current address indefinitely in that visit, with no action to recover it after a transient failure.
action: Specify the mounted settings recovery for a failed late refresh, including how the viewer obtains the current address. Extend A→B→settings to cover a late A commit, one failed read, then successful recovery.
weight: blocking

[edge-cases] Another normal window cannot verify the new password
where: §18 V10 line 700 and ACCT-9 line 748.
why: ChangePassword preserves the current session (accounts.go:219–222). A normal second window shares that session, and LoginView.tsx:35–43 gives a signed-in viewer only Continue. The proposed recovery can therefore give the user no way to submit the new password.
action: Specify a signed-out context, such as a private window, and verify the recovery after an ambiguous commit while the original session is still valid.
weight: blocking

FACT: Static review and two in-memory QueryClient probes only. Make, Docker and browser e2e were not run. One assigned lens returned.
DECISION: lens verdict (6 blocking findings).
FACT: Verified six fixed-format findings and no tracked changes. Only this working file was appended.
DECISION: lens verdict (6 blocking findings).
