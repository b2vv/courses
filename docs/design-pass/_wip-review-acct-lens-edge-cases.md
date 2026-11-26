NEXT: Read the AI Factory entry rules, then verify the repository root and load the critique contract.
FACT: Read the shared entry rules. Current owner scope permits only the named working file.
NEXT: Verify repository and review shared and project instructions.
FACT: Repository root is courses and HEAD is f6a4611a. Read the local routing and single-lens contract.
DECISION: Apply only edge-cases. Keep all progress and findings in this file.
NEXT: Read the ACCT revision and its owner-approved audit source.
FACT: main points at the requested revision. The shared reference treats scoped accessibility criteria as requirements.
NEXT: Compare section 18 acceptance and exclusions to owners’ answers.
FACT: Owner approved all 16 audit items and the taste item. Section 18 defines status mapping, resend, 2FA, current email and shared validation/layout.
NEXT: Trace confirmation, session cache, banner and lesson-gate callers.
FACT: A3/A6 only auto-submit while the mutation is idle. Both refresh session after success. The root banner and lesson gate own separate resend mutations, and the banner can coexist with either route’s resend.
NEXT: Trace server transaction order, token reuse and 2FA races.
FACT: Resend 204 discards the mailer success boolean and can follow invalidation plus failed delivery. Both confirmations consume the token before updating the account. 2FA writes can also partially apply before an error.
NEXT: Read the full route test paths, field controls, layouts and existing UI variants; check which failure modes the proposed tests distinguish.
FACT: FieldHelp also serves Select and ComboBox; the proposed V11 selector misses the direct-child enable button in the off state. Current E2E owns a run-unique guest email and can distinguish verification links by Mailpit ID.
NEXT: Verify decoder boundaries, Mailpit lookup and 2FA post-mutation transitions before weighing findings.
FACT: Compared scoped code at the plan baseline and review commit: no diff. Mailpit helper selects the first matching path, so two verify emails require explicit message identity. Confirm email also has a terminal 409 after token consumption.
NEXT: Confirm stale-cookie handling and exact locations; record actionable findings.
FACT: Authenticate returns 401 without clearing the cookie. /me maps that 401 to anonymous, while resend still authenticates the cookie before using a body email. Recovery-code copy stores a view-wide boolean that survives the new Done → regenerate flow.
DECISION: State concrete failures as plan edits; keep the Mailpit selector and shared FieldHelp boundary as proof refinements.
NEXT: Append the fixed-format findings and verify the count.

DECISION: Actionable findings for §18 revision 1.

[edge-cases] A resend 204 can mean no mail was sent
where: §18 V1, lines 555–559, ACCT-1/2
why: accounts.go:123–140 returns no error for an already verified account or a failed mail send. handler.go:373–377 discards the delivery boolean and returns 204 even after invalidating the previous link.
action: Use acknowledgement copy for signed-in 204 and keep the guest copy neutral. Check the signed-in no-op or mailer-failure case separately from successful Mailpit delivery.
weight: blocking

[edge-cases] A revoked cookie makes the guest resend form unusable
where: §18 V1, lines 562–567, ACCT-2
why: The session adapter interprets /me 401 as anonymous (session-repository.ts:270–281). Resend still reads the stale cookie first and returns 401 even with a valid body email (handler.go:357–367, 495–506). Generic retry leaves the same blocker. Duplicate cookies also produce 400 independently of email validity.
action: Specify recovery for authentication_required with a retained cookie so users can sign in again or reach a cookie-free resend. Keep session-related 400 failures out of the field-level email diagnosis where distinguishable. Test a real retained but revoked cookie.
weight: blocking

[edge-cases] Coexisting resend buttons bypass separate pending guards
where: §18 V1, lines 555–567, ACCT-1
why: The root banner coexists with the lesson gate and signed-in A3. Each current caller has a separate mutation (__root.tsx:236–238, 389–395, learn.$.tsx:186–188, 355–359, verify-email.tsx:22–24). Both can send while the other is pending. Each server call invalidates earlier tokens and issues a replacement (accounts.go:133–140), so the user may receive a link invalidated by the duplicate request.
action: Make the pending guard apply across concurrently visible controls for the same signed-in account. Test rapid activation from both controls against one request and define result placement so feedback is announced once.
weight: blocking

[edge-cases] Retrying a consumed token cannot determine the outcome
where: §18 V3–V4, lines 571–576, ACCT-5/6
why: Both Go confirmation paths consume before writing (accounts.go:95–106, 245–265). A timeout can follow a successful update or a failure after consumption. A6 can also return email_taken after consumption, so the catch-all retry is already guaranteed to fail in that case.
action: Give email_taken a terminal path to settings. Describe timeouts and lost responses as an uncertain outcome and offer a way to check current state or get a fresh link alongside any retry. Test a retry after consumption and a conflict separately from a request that never reached the server.
weight: blocking

[edge-cases] A second link can inherit the first link’s result
where: §18 V4, lines 573–578, ACCT-6
why: Both current routes auto-submit only while their mutation is idle (verify-email.tsx:17–27, email.change.confirm.tsx:18–27). A search-only navigation can retain the route instance and the first token’s result. An older in-flight completion can also arrive after the token changes.
action: Key the visible state and automatic submission to the current token. Define behavior for a changed or removed token, a late response for the old token and repeated activation of Retry. Add a same-route navigation test with two tokens and no document reload.
weight: blocking

[edge-cases] Done can expose stale 2FA state
where: §18 V5/V7/V14, lines 580–587, 591–594, 608, ACCT-7/16
why: settings.account.tsx:39–43 publishes recovery codes before /me refresh finishes, while the on/off branch reads cached /me at line 76. Done during a slow or failed refresh can expose the off branch and miss its focus target despite successful enablement. Server confirm can also enable 2FA before returning an error while creating codes (two_factor.go:101–110).
action: Require Done to land on the enabled state once enablement is acknowledged, independently of refresh timing. For ambiguous write failures, reconcile current 2FA state and show a recovery path if enabled without codes. Cover delayed and failed /me plus a partially applied confirmation.
weight: blocking

[edge-cases] A new recovery-code set can appear already copied
where: §18 V14, line 608, ACCT-16
why: AccountSettingsView.tsx:37, 63–66, 141–144 keeps copied for the full view lifetime. The new Done action lets a user copy, dismiss and regenerate in the same mount. The old success message can immediately reappear for uncopied codes, including after a late clipboard completion.
action: Tie copy feedback to the currently visible code set. Test copy, Done and regenerate without a reload, including late or failed clipboard completion.
weight: blocking

[edge-cases] The off-state enable button still fills the card
where: §18 V11, line 602, ACCT-13
why: The selector only matches children of forms. In the off state the enable button is a direct grid child (AccountSettingsView.tsx:223–239, main.css:536–547). It is one of the three full-width controls reported by audit #12.
action: Include standalone actions and prove the enable button has its own width after initial render, cancellation and successful disablement.
weight: blocking

[edge-cases] The Mailpit helper can pick the first verification email
where: §18 ACCT-2/6 verification, lines 642–645
why: readMailLink returns the first match for email and path (accounts.spec.ts:150–171). Two messages now match, so a rerun can depend on list order unless the test captures message identity.
action: Record the original message ID and wait for a distinct ID. Open the original link while the viewer is still unverified and confirm by the replacement link.
weight: non-blocking

[edge-cases] Validation fallback also reaches select and combobox
where: §18 V6, line 589, ACCT-8
why: FieldHelp is shared with Select and ComboBox (controls.tsx:170–182, 213–236). Implementing the fallback there alone changes their implicit validation copy even though V6 names only text inputs.
action: Bind the fallback to TextField and TextArea or explicitly specify and verify the behavior for empty required selection controls.
weight: non-blocking

FACT: Static source review only. No runtime checks were run. No sibling review files were read.
NEXT: Return the findings and lens verdict.
FACT: Verified ten fixed-format findings. The working file is the only scoped change.
DECISION: Lens verdict: 8 blocking findings and 2 non-blocking findings.
