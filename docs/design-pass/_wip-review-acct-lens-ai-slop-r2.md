NEXT: Read the AI Factory entry rules, then confirm the repository root and critique scope.
FACT: AI Factory entry rules read. Repository root is /Users/strelia/projects/courses.
DECISION: Use only the AI-slop lens and record progress in this file.
NEXT: Read local rules, repository briefing, and critique contract.
FACT: Local route and plan-critique contract loaded. CTO briefing identifies the existing identity and design-pass work.
NEXT: Pin revision 2 and read the revision-1 findings, owner answers, and anti-pattern reference.
FACT: Revision 2 is commit 8d171f63. The audit's owner answer covers findings 1–16 and the A6 label. The previous lens returned one blocking 204 claim and two non-blocking scope findings.
FACT: V1 still predicts delivery within minutes and says prior links are invalid. ACCT-1 now tests the shared component once, with guest input on the verify route. V6 expressly scopes the default to TextField and TextArea.
NEXT: Confirm these dispositions against the code and check revision-2 additions for unsupported work or copy.
FACT: HEAD and main are both 8d171f63, with no source drift in the requested paths. Root banner and lesson gate only render for signed-in accounts with unverified email. FieldHelp also serves both choice controls, which V6 now explicitly excludes.
NEXT: Inspect response handling, 2FA state, shared styling, and the accounts spec on the pinned source.
FACT: The HTTP handler ignores the resend boolean. Missing and already-verified users return before invalidation, while mailer failure can return after invalidation. Revision 2's uniform message still claims an unobservable outcome.
FACT: The email-change server consumes the token before updating the address. A used token cannot tell the route whether that address changed. V3 E4 currently tells users to request the change again.
NEXT: Check existing locale, session, focus test, and Mailpit paths before weighing any additional finding.
FACT: Mailpit already exposes exact message IDs, so filtering the existing inbox helper by ID serves the new resend scenario. The existing 2FA e2e supplies a real browser journey for the focus assertions.
DECISION: Keep the real-browser focus check. Evaluate whether prescribing a fixed two-second pause serves that check.
NEXT: Check issuing and lookup semantics for email-change copy, and look for an existing deterministic browser response gate.
FACT: Current A6 uses account.error.<code>, including distinct email_taken, rate_limited, and dependency_unavailable strings. The plan's blanket description of current A6 copy is inaccurate. Its current heading still incorrectly says confirmation is in progress after a failure.
FACT: RequestEmailChange does not invalidate previous tokens. The account_tokens user/purpose index is nonunique. ConfirmEmailChange returns email_taken for any matching user, including the same account if an earlier link already set the address.
DECISION: One new blocking finding covers the unsupported instructions in A6 E3/E4. The fixed two-second test delay is a separate non-blocking finding because the acceptance condition is an outstanding /me response.
NEXT: Record closure dispositions, fixed-format findings, and the lens verdict.

## Revision 1 dispositions

- 204 confirmation claim — open. The uniform 204 copy still promises arrival within minutes and states that prior links are invalid.
- Guest tests on signed-in-only surfaces — closed. ACCT-1 verifies shared outcomes once and checks that the banner and gate each display their action's result. ACCT-2/3 cover the guest form and stale-cookie path on the verify route.
- Localized validation spilling into choice controls — closed. V6 limits the fallback to TextField/TextArea, keeps explicit error-message precedence, and excludes Select/ComboBox.

[ai-slop] Resend copy still promises delivery and link invalidation
where: Plan §18 V1, lines 576 and 583, and ACCT-1/2, lines 652–653.
why: A 204 can be a no-op for a missing or verified account, or a mailer failure after invalidation. Neither “надійде за кілька хвилин” nor “попередні посилання більше не діють” follows from that response. The person can wait for a message that never arrives or discard a link that still works. See accounts.go:123–140 and handler.go:373–377.
action: Use the same acknowledgment for both audiences, such as «Запит прийнято. Якщо отримаєте новий лист, скористайтеся посиланням із нього.» Keep the actual delivery and prior-link rejection checks in the Mailpit scenario. Qualify the cross-tab note with the branch that actually replaces a token.
weight: blocking

[ai-slop] Email-change error copy tells users to redo a change that may already be complete
where: Plan §18 V3, E3–E5, lines 603–607, and the repeated-link e2e at line 681.
why: E3 attributes email_taken to another account, but the lookup can find the same account after it confirmed an earlier link for the same address. E4 tells the person to request the change again even when the first confirmation succeeded, which is the exact setup of the proposed repeated-link e2e. Both messages can send the person back through an already completed flow. See accounts.go:226–265 and migration 00014_accounts.sql:9–25.
action: Keep E3 account-neutral. In E4, direct the person to check the current address in settings and request another link only if the change is still needed. In the reused-link e2e, follow the settings action and verify that the already-changed address is shown. In the Facts subsection, describe A6 as using code-specific copy under the wrong heading, as ConfirmEmailChangeView.tsx:18–24 does.
weight: blocking

[ai-slop] The focus test prescribes a two-second pause for a state it can control
where: Plan §18 verification table, ACCT-8, line 683.
why: The required observation is correct focus while /me is outstanding. Two seconds adds a timed window that a slow runner can miss. The real browser assertion is still warranted by audit #6.
action: Hold the relevant /me response during the focus assertion, release it deterministically, and verify that focus remains on the expected element when the response settles.
weight: non-blocking

FACT: Reviewed the pinned plan and code by inspection. No make, Docker, or e2e commands ran. Only this working file was edited.

Lens verdict: 2 blocking findings.
