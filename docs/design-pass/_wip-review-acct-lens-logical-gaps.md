NEXT: Read the installed agent-entry rules before reviewing the ACCT plan.
FACT: Read the shared entry rules.
NEXT: Verify the repository root, then read the scoped rules, briefing and lens contract.
FACT: HEAD and main match f6a4611a, and the plan and audit have no local changes.
DECISION: Apply only the logical-gaps lens and keep the plan read-only.
NEXT: Read the briefing, project rules, plan revision and owner evidence.
FACT: The owner approved all sixteen audit findings and the A3/A6 label alignment. The audit marks real token success and several 2FA paths as untested.
NEXT: Read the complete §18 and its verification rows alongside the audit causes.
FACT: DP2 is the open design-pass implementation task. The owner evidence explicitly covers sixteen numbered rows plus label alignment.
NEXT: Trace §18 state transitions through the frontend routes and server endpoints.
FACT: §18 defines V1–V15, ACCT-1–18, and an ordered gate for the implementation.
NEXT: Read route orchestration and session data ownership.
FACT: Confirmation routes keep mutation state by mounted route and refresh identity on success. Account settings stores enrollment and recovery locally and reads 2FA enabled from the current-user query.
NEXT: Inspect backend token semantics, UI state precedence and actual browser test entry points.
FACT: Resend returns neutral 204 even when the mailer fails. Both confirmation use cases consume before applying the account update. The proposed old-link e2e can conflict with the verified-session branch.
NEXT: Check CSS and 2FA transitions and narrow each issue to a concrete plan correction.
FACT: V11 only sizes form children, while the enable and recovery actions sit outside forms as stretched card grid items. Current() maps a 401 to null, even when the browser still has a cookie.
NEXT: Confirm stale-cookie behavior and the recovery-code transition, then finalize the actionable findings.
FACT: Recovery codes render as soon as confirm succeeds, while identity refresh remains in flight. The existing enabled/disabled branch reads only current-user data. Existing Mailpit helper returns the first link matching the pathname.
NEXT: Verify prior CSS scope and shared field call sites before recording the verdict.
FACT: The reviewed frontend/backend paths have no diff between the plan's fact baseline fa74f8f4 and f6a4611a. FieldHelp is shared by Select and ComboBox too, and the verification table omits ACCT-18.
DECISION: Record the following plan corrections. Review scope was static source inspection at f6a4611a.

[logical-gaps] A 204 cannot confirm that a verification email was sent
where: §18 V1, ACCT-1 (plan.md:555–559, 618)
why: accounts.go:123–140 returns a no-op for missing or verified accounts and turns mailer failure into false with no error. handler.go:373–377 ignores the boolean. The signed-in UI can claim delivery after a no-op or send failure.
action: Define 204 copy as an acknowledgment of the request for both audiences, preserve the guest's neutral response, and keep actual delivery as a separate Mailpit assertion.
weight: blocking

[logical-gaps] The guest resend branch assumes an absent session cookie
where: §18 V1, ACCT-2 (plan.md:562–567, 619, 642)
why: currentUser maps authentication_required to null (session-repository.ts:270–281). A browser with a revoked session can therefore see the guest form while still sending the cookie. Resend tries that cookie before using the body and returns the authentication error (handler.go:357–373). When identity is unreadable, a still-valid cookie can also override the typed address.
action: Distinguish a confirmed guest from an unreadable session, define an actionable recovery for the resend 401 such as sign-in or session reset, and add proof with a revoked session cookie alongside the clean guest.
weight: blocking

[logical-gaps] A3 has conflicting headings and an underspecified old-link fixture
where: §18 V1/V4, ACCT-3/6 (plan.md:562–576, 620–623, 642–645)
why: V1 requires «Пошту вже підтверджено» for a verified viewer on a bad link, while V4 requires «Посилання не працює». The current account spec uses the signed-in page for confirmation (accounts.spec.ts:283–284). After it consumes the replacement link, that viewer is verified. Also, readMailLink returns the first matching pathname (accounts.spec.ts:150–177), so finding a link alone does not distinguish the second message.
action: Specify heading precedence by token and session state. Capture the original message ID, get the replacement from a different message, prove the old link fails in the guest context or before verification, then consume the replacement in the signed-in context and exercise the verified override separately.
weight: blocking

[logical-gaps] Same-token retry has no recovery after the token is consumed
where: §18 V3/V4 and ACCT-5/6 (plan.md:571–577, 622–623, 645)
why: Both handlers consume the token before updating the account (accounts.go:95–107, 245–265). A timeout or 5xx can therefore leave the token spent, including when the account update already succeeded but the response was lost. A6 also returns email_taken after consumption, which V4 sends to the same futile retry. A mock that fails before the write cannot prove recovery for either branch.
action: Specify email_taken as a terminal A6 outcome with account settings as the next step. Define the status and recovery path when an ambiguous failure is followed by token_used, including how to refresh account state or request a fresh link. Add these sequences to ACCT-6.
weight: blocking

[logical-gaps] The 2FA focus target depends on an unfinished session refresh
where: §18 V5/V14, ACCT-7/16 (plan.md:580–587, 608, 646–648)
why: On confirmation, settings.account.tsx:36–43 displays recovery codes before awaiting the current-user refresh. Done can hide the codes while the cache still says disabled, so the enabled focus target is absent. Disable has the reverse dependency (settings.account.tsx:45–53, 76). The browser plan checks focus for Enable and Confirm only, although Done and Disable depend on the refresh order.
action: Define when Done can proceed and what establishes the rendered enabled/disabled state on mutation success. Give focus to the target once it renders, and verify Done and Disable in-browser with a delayed current-user read.
weight: blocking

[logical-gaps] The width rule misses the audited Enable 2FA button
where: §18 V11 and ACCT-13 (plan.md:602, 630, 650)
why: V11 sizes direct children of forms. Enable 2FA is a direct grid child of the card (AccountSettingsView.tsx:223–237, main.css:536–548), so it still stretches across the card. Recovery actions are also outside forms (AccountSettingsView.tsx:131–145).
action: Include the audited non-form Enable action and define the recovery action layout. Measure the three buttons named in audit #12 individually.
weight: blocking

[logical-gaps] The shared validation helper has callers outside the stated scope
where: §18 V6 and ACCT-8 (plan.md:589, 625, 647)
why: FieldHelp also renders errors for Select and ComboBox (controls.tsx:170–182, 236). Applying the default there as written adds behavior outside the stated text-field and textarea work and expands verification.
action: Bound the localized default to TextField/TextArea or list the additional callers and verify their default and explicit messages as part of V6.
weight: non-blocking

[logical-gaps] The verification table leaves the A6 label without a check
where: §18 ACCT-18 and verification table (plan.md:635, 639–651)
why: The owner approved the label alignment and the acceptance table records it, but the corresponding proof row is absent.
action: Add a view or browser assertion for the A6 label in the existing verification table.
weight: non-blocking

FACT: Review completed through static inspection. No make check, Docker, or e2e ran. Only this working file was edited.
NEXT: Return the lens findings to the plan author.
lens verdict: 6 blocking findings
