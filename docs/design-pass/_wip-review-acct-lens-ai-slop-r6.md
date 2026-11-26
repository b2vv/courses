NEXT: Read the shared agent entry rules before starting the revision 6 review.
FACT: Shared entry rules read and repository root verified.
NEXT: Read local routes, briefing, task index, and critique contract.
FACT: Local routes located the open design-pass slice and the skill defines filler, invented ceremony, and unsupported future scope as the AI-slop lens.
DECISION: Apply only the requested lens and write only to this working file.
NEXT: Read revision 6, audit evidence, prior AI-slop reviews, and cross-cutting anti-patterns.
FACT: Revision 6 is on main. The initial combined read was truncated, so the findings and plan need bounded reads.
NEXT: Read the revision 6 delta and enumerate the findings across revisions 1 through 5.
FACT: Enumerated the 14 prior findings and confirmed owner approval for all 16 audit items and the A6 label.
NEXT: Read all current decision and proof rows with stable line numbers.
FACT: Both token views remove the 403 retry. The 2FA conflict text describes an observed refresh without inventing another window.
NEXT: Check exact recovery copy and trace the handlers and settings session path.
FACT: Password update precedes session cleanup, email-change 204 can follow a mailer failure, and cached null is what triggers the settings sign-in redirect.
NEXT: Check password validation, frontend error mapping, current views and browser fixtures.
FACT: Checked all named frontend paths. The shared resend and text-control scopes match real callers, and Mailpit and focus proofs extend existing journeys.
NEXT: Check the exact password policy, proof text and fresh-session implementation before weighing residual findings.
FACT: V10 directs the user to reuse the requested password as current after a failed retry; the domain rejects equal passwords before checking credentials.
NEXT: Confirm the mapped error and recovery route, then distinguish duplicate proof wording from separate proof obligations.
DECISION: Treat the impossible V10 password remedy as blocking. Treat the repeated ACCT-6 route-test prescription as non-blocking because one stale-read scenario can prove both statements.
NEXT: Verify the reviewed sources match the pinned snapshot and record final dispositions.
FACT: The plan, audit, and inspected source matched cf5c5477. The HTTP handler and frontend adapter preserve password_unchanged as a distinct 422 error. The application already has a sign-in flow and password recovery.
DECISION: Close every earlier AI-slop finding. Record one new blocking recovery-copy issue and one non-blocking proof duplication.

## Earlier findings

- Revision 1 (three closed). V1 acknowledges acceptance conditionally and ACCT-2 supplies the Mailpit proof. ACCT-1 tests the shared outcomes once and the signed-in callers for integration, with guest behavior in ACCT-2/3. V6 limits fallback validation copy to TextField/TextArea, preserves explicit errors, and excludes Select/ComboBox.
- Revision 2 (three closed). V1 no longer promises delivery or invalidation from the order of messages. E3 is account-neutral, E4/E5 direct users to check the current address, and the A6 browser journey follows the settings action to verify that address. ACCT-8 holds and explicitly releases the response.
- Revision 3 (four closed). S3 directs the holder of a rejected link for another address to request a new message. V5a separates code, session, request, and uncertain-write errors, with both supported kinds of code and the existing admin reset path for uncertain regeneration. V1 leaves prior-link rejection to the sequential Mailpit scenario. The Facts subsection describes A6's existing code-specific errors.
- Revision 4 (two closed). V5a now specifies how authentication_required cancels the old session read and publishes null for the existing redirect. Both token routes remove the retry for an origin refusal.
- Revision 5 (two closed). S5/E5 and ACCT-5/6 drop the same-origin reload/retry. V5a and ACCT-9 show the observed 2FA state with cause-neutral copy.

[ai-slop] Password recovery sends the user into an invalid resubmission
where: §18 V10 (line 689) and ACCT-9 verification (line 737).
why: The new guidance tells a user whose retry rejected the old password to enter the requested new password as the current one. The form preserves both values after a failure and submits both (AccountSettingsView.tsx:32–41). Following the instruction leaves the same password in both fields. ValidatePasswordChange rejects equal values before checking credentials (accounts.go:195–208, account_policy.go:84–91), and the handler returns 422 password_unchanged (handler.go:839–840). Following the guidance cannot establish whether the first write succeeded.
action: Keep the uncertainty message and replace the field-swap instruction with an existing recovery path. For example, if the original password is rejected on retry, tell the user to try signing in with the requested password and use password recovery if sign-in fails. Have ACCT-9 assert the supported instruction.
weight: blocking

[ai-slop] ACCT-6 prescribes the same retry journey twice
where: §18 ACCT-6 Unit / route proof (line 734).
why: The new held-/me scenario already goes from E5 through retry and E4 to the fresh address in settings. The remainder of the cell asks for that sequence again without adding an observation. This can yield two fixtures for the same user journey.
action: Specify one E5 → retry → token_used → E4 → settings scenario. Retain both observations in it, that the old read is cancelled and that settings displays the fresh address. Keep the separate 403 assertion and the live repeated-link browser journey.
weight: non-blocking

FACT: One lens requested, one verdict returned. Only this working file was edited. Source inspection only; no make check, Docker, or e2e ran.

Lens verdict: 1 blocking finding.
