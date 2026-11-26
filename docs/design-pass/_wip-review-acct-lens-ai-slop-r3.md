NEXT: Read the AI Factory entry rules and the AI-slop lens contract, then verify the ACCT revision.
FACT: Read the installed entry rules. The owner scoped all writes to this append-only file.
NEXT: Verify the repository root and read the local rules, briefing and named skill.
FACT: HEAD and main both resolve to efc0bcf9. Read the local rules and the single-lens contract.
NEXT: Read section 18, prior AI-slop findings, the owner audit and the reference anti-patterns.
FACT: Owner approved the 16 audit items and the label preference. Read the anti-pattern reference.
NEXT: Read the full prior findings and untruncated revision 3 implementation and proof rows.
FACT: Revision 3 partitions the resend tests by surface, scopes validation messages to text controls, directs A6 failures to the current address, and uses a controllable response in the focus test.
NEXT: Check the matched route, UI, controls and CSS implementations for inaccurate descriptions and added ceremony.
FACT: No source drift found from the plan fact baseline in the named source trees. The revision 3 delta retains an unconditional statement about invalidating old links in V1.
NEXT: Verify the live branches and current locale usage against the identified copy issues.
FACT: A6 renders account.error.<code>, so the earlier factual correction is still missing. Text field and signed-in surface scoping match the current wiring.
NEXT: Check the service branches, session query options, and browser fixture red/green behavior.
FACT: The resend service skips invalidation for missing and verified accounts and can invalidate old tokens before issuing a replacement. A6 E3/E4 now direct users to the current address, and its e2e follows that action. ACCT-8 now holds and releases the response explicitly.
NEXT: Confirm the exact 2FA regeneration response paths and inspect the existing session and Mailpit proof details.
FACT: Regeneration authenticates and decodes before it touches recovery codes, then accepts either a fresh authenticator code or an unused recovery code. S3 prescribes reopening the same link even when verification just rejected it as invalid or used.
NEXT: Pin source lines for the copy and classify the remaining findings by user impact.
FACT: Compared section 18 at efc0bcf9 against the cited account routes and UI, shared styles and controls, session layers, identity handlers and services, locale strings, and existing accounts browser spec. The observed audit findings supply actual users for the shared resend UI, focus transitions, validation copy, recovery-code feedback and visual proof.
DECISION: Rev 1 guest-case scope is closed by ACCT-1/2/3. Rev 1 validation scope is closed by V6. Rev 1 and rev 2 delivery promise is closed by V1's conditional acknowledgment, with the stale invalidation explanation still open. Rev 2 E3/E4 user-copy and navigation proof are closed, with the inaccurate current-A6 description still open. Rev 2 fixed-delay finding is closed by ACCT-8's explicit response release.
DECISION: Two new copy blockers. The suggested action in S3 sends a user back through a link the API already rejected. V5a tells a user with a known authorization or request failure to obtain another authenticator code. Both have existing recovery instructions or capabilities the plan can use.
NEXT: Record the results in the fixed finding format and return this single lens's verdict.

[ai-slop] S3 sends users back to a rejected link
where: Plan §18 V2 S3 (line 605), including the already-invalidated-link proof (lines 717–719).
why: S3 instructs the user to sign out and reopen the same link if it belonged to another account. VerifyEmail uses only the token to select the account (handler.go:321–343, accounts.go:95–106). Signing out changes whose status the page can show, but it cannot make a definitively rejected link work. The Mailpit journey already establishes this by using the same rejected link as a guest and as a signed-in user.
action: Keep the statement about the signed-in account. When the link was definitively rejected, tell the user to request a new verification link for the other address and point them to the guest resend path. Reserve reuse of the same link for an inconclusive result.
weight: blocking

[ai-slop] The regeneration fallback gives the wrong remedy for known failures
where: Plan §18 V5a regeneration (line 645), V7 (line 650), and ACCT-9.
why: The catch-all directs the user to fetch a new authenticator code even after a malformed request or a rejected session. Authentication and decoding precede the regeneration call (handler.go:126, 280–296). A new authenticator code cannot repair either failure. For uncertain failures, the message also suggests an authenticator is the only recovery path even though an unused recovery code may still work if the request stopped before deletion (two_factor.go:227–241, 258–282).
action: Keep session and request errors actionable with their own messages. Reserve the warning about previous codes for outcomes where the write is uncertain. Recommend a fresh authenticator code if available and preserve the option to use a different unused recovery code if it still works.
weight: blocking

[ai-slop] V1 still overstates when old links stop working
where: Plan §18 V1 (lines 589 and 596).
why: The cross-tab note says every request invalidates links and the rationale ties invalidation to issuance or delivery. Missing and verified accounts return before invalidation, and invalidation may complete before issuance fails (accounts.go:123–140). The 204 acknowledgment works without these claims. They give the implementer two conflicting versions of the same rule.
action: State that replacement attempts for a found, unverified account invalidate prior links before issuance. Keep the uniform 204 acknowledgment and ACCT-2's Mailpit verification of the successful replacement path.
weight: non-blocking

[ai-slop] The A6 baseline still flattens distinct error messages
where: Plan §18 Facts (line 577).
why: ConfirmEmailChangeView.tsx:18–24 already looks up account.error.<code>. The incorrect heading and missing next action are real. The blanket description makes the reader rediscover which error copy already exists.
action: Describe A6 as using code-specific error copy under the in-progress heading. Leave E3/E4 and the navigation proof as revised.
weight: non-blocking

FACT: Only this working file was edited. Verification was by source inspection. No make, Docker, or e2e commands ran. This is the one requested lens, which returned a verdict.

Lens verdict: 2 blocking findings.
