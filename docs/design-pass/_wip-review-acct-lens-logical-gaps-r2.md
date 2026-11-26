NEXT: Read the repository rules and review contract, then trace revision 1 findings against revision 2.
FACT: Verified the Courses root and read the shared entry rules. NEXT: Read local rules, project briefing, lens contract and prior findings.
FACT: Read the logical-gaps contract and the DP2 context. DECISION: Keep this pass to one lens and the working file. NEXT: Compare the fixed plan snapshot with all prior findings and owner answers.
FACT: HEAD and main both match 8d171f63; scoped tracked sources match the snapshot. Revision 2 adds explicit precedence, test rows, and recovery. NEXT: Trace the actual request, session and redirect paths before deciding whether the earlier findings close.
FACT: A3/A6 currently use token mutations with session refresh on success; A7 uses cached current-user data and refetchOnMount false. Owner approved all audit rows. NEXT: Check guard freshness, transport outcomes, and whether planned browser fixtures reach their claimed branches.
FACT: Handler confirms 204 can be a no-op or mailer failure; a valid cookie overrides the submitted address. The A7 guard does force a fresh current-user read on entry. NEXT: Verify UI state ownership, 2FA code formats and existing e2e assertions; inspect revision 2 additions for conflicts.
FACT: The test can distinguish the Mailpit messages and run old-link failure before signed-in success. root-identity already distinguishes unavailable from anonymous. NEXT: Verify precise 2FA render precedence, resend error coverage, and shared style behavior against the existing sources.
FACT: The current 2FA view renders recovery codes first, then enrollment, then enabled/disabled. Disable also writes state before cleaning up codes. V6 and V11 now cover the specific prior scope gaps. NEXT: Resolve which remaining issues require plan changes and validate the exact code references.
FACT: Compared all eight revision 1 findings. The 204 wording and unreadable-session branch remain open (the stale-cookie 401 subcase is covered). The heading order and old-link fixture, A6/A3 post-consumption recovery behavior, successful 2FA transitions, non-form button sizing, text-field scope, and A6 label check are specified. The A6 recovery sequence and delayed-read focus proof still need sharper verification.
DECISION: Record two remaining revision 1 blockers and two new revision 2 blockers. Distinguish the new post-error behavior from the successful 2FA transition addressed in revision 1.

[logical-gaps] The 204 acknowledgment still promises mail and link invalidation
where: §18 V1 and ACCT-1/2 (plan.md:575–576, 652–653).
why: The proposed copy says an unverified address will receive mail within minutes and old links have stopped working. ResendVerification returns without invalidating anything for missing or verified accounts. It can also return 204 after invalidating the old link and failing to send its replacement (accounts.go:123–140, handler.go:373–377). ACCT-2 proves one successful delivery and cannot make those promises valid in every 204 branch.
action: Use conditional copy that acknowledges the request and tells people to use the latest link if a message arrives. Keep the distinct Mailpit delivery and old-link assertions.
weight: blocking

[logical-gaps] An unreadable session still submits a guest address that the server can ignore
where: §18 V2 S4 and ACCT-2/3 (plan.md:592–595, 678–680).
why: Root identity distinguishes unavailable from anonymous (root-identity.ts:6–9, 26–33). S4 still gives both the same address form. If /me fails with a valid cookie, resend can recover and use the cookie's user ID, ignoring the typed address and potentially replacing that account's link (handler.go:357–373). The new 401 branch handles a rejected cookie only.
action: Give unreadable identity its own recovery action to recheck the session before offering the appropriate form. Keep the stale-cookie 401 handling and add a case where /me fails but the cookie remains valid.
weight: blocking

[logical-gaps] A fresh session read does not establish which email the failed token was for
where: §18 V2 S3/S5 and ACCT-5 (plan.md:591–595, 656).
why: Revision 2 extends S3 to ambiguous failures. /me can say only that the current viewer's address is verified. VerifyEmail consumes the link and uses the token's user ID (accounts.go:95–106, handler.go:321–343). An already verified viewer opening another account's link can therefore get «Пошту вже підтверджено» and «До навчання» after a failure without any evidence that the address named by the link was verified. The same ambiguity applies to the old-link override.
action: Scope the S3 message explicitly to the signed-in account and keep a way to recover a link meant for another account. For an ambiguous failure, treat a verified session as evidence of completed verification only when the flow can establish which account is being verified. Add a route case for an already verified viewer opening a link for another account.
weight: blocking

[logical-gaps] A partially applied 2FA confirmation needs an explicit exit from setup
where: §18 V5/V5a and ACCT-9 (plan.md:614–623, 660, 684).
why: Following confirm → 503 → fresh /me with 2FA enabled, setting the enabled flag alone leaves the setup screen mounted. Only the existing success callback clears enrollment, and the view picks enrollment before the enabled branch (settings.account.tsx:21, 36–43, AccountSettingsView.tsx:146–177). V5 also says errors leave content and focus alone, while V5a requires both to move for this outcome.
action: Define the reconciled server outcome as a phase transition that clears enrollment and the submitted code, renders the enabled card with the recovery notice, and focuses that notice. Make it an explicit exception to ordinary validation-error behavior and extend ACCT-9 to assert that setup disappears, regeneration is available, and focus reaches the notice.
weight: blocking

[logical-gaps] The 2FA timing check can pass after the delayed read completes
where: §18 ACCT-8 browser proof (plan.md:683).
why: A fixed two-second wait on /me followed by Playwright's auto-waiting focus assertion can pass even if both transitions still depend on the read. The result would prove eventual focus and leave the independence requirement in V5a unmeasured.
action: Hold the real /me response behind a test-controlled release and assert that «Готово» and Disable land on their focus targets while that response is still held. Then release it and check that the focus target persists.
weight: non-blocking

[logical-gaps] The A6 checks do not connect uncertain confirmation to the settings recovery path
where: §18 V3 and ACCT-6 (plan.md:603–607, 681).
why: Running standalone E1–E5 tests and reusing an already successful link covers E4 but does not prove the E5 → retry → E4 → settings sequence. The settings guard already fetches fresh identity (require-identity.ts:11–20), so the planned behavior is implementable without changing the API.
action: Check an ambiguous confirmation result followed by token_used, then follow the settings link and assert the address shown comes from the fresh read. Include the guest's sign-in return to settings in the route proof if covered by the same fixture.
weight: non-blocking

FACT: Static review of 8d171f63 on main. No make check, Docker, or e2e commands ran. Only this working file was edited.
NEXT: Return findings to the plan author.
lens verdict: 4 blocking findings
