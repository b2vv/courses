NEXT: Read the shared entry rules and verify the repository root.
FACT: Read shared entry rules and verified the courses repository root.
NEXT: Read repository instructions, briefing, and the requested lens contract.
FACT: main is at f6a4611a. Read the local rules and the plan-critique lens contract.
DECISION: Apply AI-slop only, using the requested working file as the review record.
NEXT: Read the pinned plan, audit evidence, and the shared anti-pattern reference.
FACT: Audit records concrete user impact for findings 1–16 and the owner approved all findings plus the naming change.
NEXT: Map the full decision and verification text to the existing frontend routes and views.
FACT: The three resend surfaces exist. Only VerifyEmailView has a guest branch under the plan, but ACCT-1 lists the same five test cases for all three.
FACT: FieldHelp is also called by Select and ComboBox, beyond the TextField and TextArea scope named in V6.
NEXT: Check the exact shared helper, styles, API contract, and e2e usage to separate filler from legitimate shared behavior.
FACT: The frontend session and repository already accept an optional email. Existing CSS supplies the status-page zoom and AUTH spacing patterns. The available account journey already exercises real tokens and 2FA.
NEXT: Confirm the resend success semantics, session guards, and proposed test assertions against the exact source lines.
FACT: Existing account e2e covers registration, Mailpit, both verification barriers, and the 2FA transitions. The fixture can extend those flows.
NEXT: Verify the 204 delivery claim and identify current consumers affected by the proposed shared defaults.
FACT: The resend service returns false with nil error on a mailer failure, and the HTTP handler discards the boolean. The adapter sees the same 204 as a successful send. The signed-in banner and gate have no email input and cannot show a guest field error.
DECISION: Report one blocking claim in the proposed copy and two non-blocking sources of extra scope. The shared resend result has three actual consumers, and the plan's mandatory verification is justified by the shared field change.
NEXT: Return this lens's findings and verdict.

[ai-slop] The resend confirmation promises more than the API knows
where: Plan §18 V1, ACCT-1/2 (lines 555–559, 618–619).
why: The signed-in text «Лист надіслано на {email}.» and the guest text's conditional claim can both falsely tell a user to look for a new message. accounts.go:123–140 returns false/nil for a no-op or failed mail send and handler.go:373–377 drops the boolean and sends 204. The successful Mailpit scenario establishes delivery for that fixture only.
action: Specify 204 copy that acknowledges the request without promising delivery. Keep the guest response uniform and retain Mailpit receipt as the separate delivery assertion.
weight: blocking

[ai-slop] Signed-in-only surfaces inherit guest test cases
where: Plan §18 ACCT-1 in both tables (lines 618, 641).
why: The matrix names guest success and guest field errors for VerifyEmailView, EmailVerificationBanner and EmailVerificationGate, but the root renders the banner only for a signed-in unverified user and the gate enforces the same condition (__root.tsx:389–394, EmailVerificationGate.tsx:27–28). Faking a guest form in all three adds tests with no user path.
action: Specify the reachable outcomes by surface. Put guest success and invalid-email field feedback on VerifyEmailView. For each signed-in surface, verify that it shows the shared result.
weight: non-blocking

[ai-slop] Localized text-field errors spill into unrelated controls
where: Plan §18 V6 and Scoped impact (lines 589, 663).
why: The audit shows English required-field messages on account text inputs. FieldHelp also serves Select and ComboBox (controls.tsx:182, 236). Putting the new default there for every caller expands validation-copy work to choice controls without a finding or owner decision for those controls.
action: Specify that TextField and TextArea opt into the localized fallback and retain the existing explicit error-message precedence.
weight: non-blocking

Lens verdict: 1 blocking finding.
