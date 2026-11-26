NEXT: Read the AI Factory entry rules and verify the repository root.
FACT: Repository root verified and shared entry rules read.
NEXT: Load local routing, repository briefing, and the requested lens contract.
FACT: HEAD is 08fc8ae7 on main. Loaded local rules and scoped the critique to AI-slop.
NEXT: Read the briefing, linked anti-pattern rules, task pointers, and prior findings.
FACT: Prior notes capture the resend, validation, current A6 copy, error recovery, and focus-test findings. The rev 4 diff changes only E5.
NEXT: Read the complete current section and compare its contract with the owner audit.
FACT: The owner approved fixes for all 16 audit items and the A6 label. Rev 4 visibly addresses S3, the inaccurate A6 baseline and the resend invalidation note.
NEXT: Compare the untruncated plan rows with the named source paths.
FACT: Read all of §18, including its acceptance and proof tables. ACCT-1, V6, ACCT-8, E3/E4 and the E4 browser journey retain their earlier corrections.
NEXT: Trace current account API errors, translations, and 2FA redirects, then compare remaining view, style and browser evidence.
FACT: The requested source trees match the plan fact baseline and pinned commit. Resend 204 can mean no-op or mailer failure. Both confirmation services consume the token before writing. Regeneration accepts either a fresh TOTP or an existing recovery code and deletes codes before inserting replacements.
NEXT: Check exact translated error text and whether the regeneration session remedy names behavior the existing hook actually provides.
FACT: Shared controls, status-page styles and the existing account browser journey supply the mechanisms named by the revision. The recovery redirect observes cached /me; it does not itself fetch. The adapter also maps a concrete 409 twofactor_not_enabled branch.
NEXT: Verify the exact 2FA error codes and locale keys, and decide which residual issues belong to the requested lens.
FACT: Both locales supply the E5/S5 error strings. The 403 handler refuses the origin before calling the token action. The regeneration POST can reject an expired cookie while settings still has a cached user.
DECISION: Treat the named regeneration redirect as an unsupported recovery instruction. Treat the same-origin 403 retry as non-blocking because the refusal is visible and E5 still offers settings.
NEXT: Record dispositions and fixed-format findings.

## Earlier findings

- Rev 1 and 2, resend delivery and invalidation promises: closed. V1 acknowledges the request and gives conditional advice for receipt of a new message. ACCT-2 tests issuance and invalidation in Mailpit.
- Rev 1, guest cases on signed-in-only surfaces: closed. ACCT-1 tests shared outcomes once and checks the signed-in integrations. ACCT-2/3 cover the guest field on the verify route.
- Rev 1, validation copy spreading to choice controls: closed. V6 limits the fallback to TextField/TextArea and preserves explicit error messages.
- Rev 2, A6 copy sending users back through an already completed change: closed. E3/E4 direct users to check the current address and ACCT-6 verifies it after following E4.
- Rev 2, fixed-delay focus test: closed. ACCT-8 explicitly holds and releases the response.
- Rev 3, S3 instructing users to reopen a rejected link: closed. S3 directs someone holding a link for another address to request a new one.
- Rev 3, generic regeneration recovery: partially closed. V5a separates known errors from uncertain writes and reflects both accepted code types. The new session-expired instruction has a remaining unsupported claim.
- Rev 3, V1 invalidation explanation: closed. V1 scopes the replacement path to a found, unverified account and describes invalidation before issuance.
- Rev 3, flattened A6 baseline: closed. §18 Facts describes code-specific copy under the pending heading.

[ai-slop] Regeneration names a sign-in recovery the existing hook cannot start
where: §18 V5a regeneration (line 656) and ACCT-9 (lines 693, 717).
why: The regeneration POST can return 401 authentication_required while settings still has a cached user. useRedirectWhenSessionEnds navigates only when the session query contains null. Settings disables refetch on mount and refreshes the session on regeneration success only, so the named hook alone leaves the user at the regeneration controls. See handler.go:126,495–511; settings.account.tsx:20–23,54–60; require-identity.ts:30–51.
action: Specify a fresh session read on authentication_required so a null result triggers the existing redirect. Alternatively, show the localized session-expired message with a sign-in action. In ACCT-9 start with a cached user and check that the POST's 401 leads to sign-in.
weight: blocking

[ai-slop] The origin-refusal rows offer a retry with the same origin
where: §18 S5 (lines 616–618), E5 (line 628), and ACCT-5/6 (lines 713–714).
why: The updated copy names the security refusal, but the retry sends the same origin. The origin check runs before token handling, so another click from that page repeats the refusal. See handler.go:514–523.
action: For origin_forbidden, show the refusal and keep applicable navigation. Offer the same-token retry for failures it can resolve.
weight: non-blocking

FACT: Reviewed §18 at 08fc8ae7 by source inspection. Only this working file was edited. No make check, Docker or e2e ran. One requested lens, one verdict returned.

Lens verdict: 1 blocking finding.
