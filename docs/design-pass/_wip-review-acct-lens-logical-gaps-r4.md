NEXT: Read the shared entry rules, then verify the repository and review scope.
FACT: Shared entry rules read. This is a single, explicitly scoped review.
DECISION: Use the requested working file as the only write target.
NEXT: Verify the root and read local project rules.
FACT: Verified main at 08fc8ae7 and read AGENTS.local.md.
NEXT: Read the lens contract, repository briefing, and task pointer.
FACT: The contract defines logical gaps as missing prerequisites, state, ordering, or proof. The task tracker identifies ACCT as part of DP2.
DECISION: Assess only this lens at the pinned revision, without consolidation or edits to the plan.
NEXT: Read §18, its source audit, and all three prior logical-gaps records.
FACT: Collected eight revision 1, six revision 2, and four revision 3 findings. The revision 4 commit itself changes one plan line.
NEXT: Read all of §18 and the source audit, including owner answers.
FACT: The plan and scoped tracked sources match HEAD. The owner approved all 16 audit items and the A6 label. The tip patch changes E5 copy, while the full revision also changes session freshness, token lifetime and 2FA reconciliation.
NEXT: Trace the current state tables and their proof rows against the route and query code.
FACT: V2 now starts a fresh session read on tokenless entry, and V4a defines a new attempt per mount. V5a keeps forms mounted during reconciliation.
QUESTION: Do S5 retry and manual 2FA reconciliation follow the same state and focus rules?
NEXT: Match acceptance and proof rows to current frontend and backend behavior.
FACT: ACCT-5 expects both retry-to-S3 and token_used-to-S4. ACCT-9 tests the first recheck and only says “recovery” after a manual recheck. The current settings card uses the shared /me cache for its enabled flag.
NEXT: Verify API error mapping, the exact session-refresh behavior, real browser fixtures and the focus primitive.
FACT: The API has explicit 409 already-enabled/not-enabled outcomes. The existing session-expiry hook watches query data and does not consume mutation errors.
NEXT: Check current translation keys, token/account identity, focus after manual recheck, and whether the e2e fixtures support the proof.
FACT: The token endpoints update the token owner independently of the browser viewer. The mounted 2FA forms already guard submit handlers independently of the button. Mailpit and accounts fixtures can exercise the planned link and success paths.
NEXT: Isolate remaining logical gaps and compare the final E5 amendment against existing locale and transport contracts.
FACT: Both locales already provide the status messages referenced by the final E5 amendment, including the 403 refusal. The settings guard and safeReturnTo support the A6 recovery destination. The scoped frontend and backend sources did not change between the plan's fact baseline and 08fc8ae7.
DECISION: Revision 1 dispositions in original order. (1) Closed by the conditional V1 acknowledgment and separate Mailpit delivery check. (2) Closed by the current screen's session read, tokenless entry coverage and stale-cookie 401 recovery. (3) Closed by the ordered state table and different message ID plus guest-first check. (4) E3 and E5 → E4 → settings are closed. A3's retry precedence still conflicts with the state table, as recorded below. (5) Closed for the success paths by V5a and the controlled /me hold. (6) Closed by V11 and the three separate measurements. (7) Closed by the explicit V6 TextField/TextArea scope. (8) Closed by ACCT-20 and its view check.
DECISION: Revision 2 dispositions in original order. (1) Closed by conditional 204 copy. (2) Closed by S0, a fresh tokenless read and a stale-null route check. (3) The initial already-verified case is now S5 and S3 wording scopes the result to the current account. The retry instruction still bypasses that condition, and the pre-attempt snapshot is not authoritative. (4) Closed for partially applied enable by the explicit transition and cleanup. (5) Closed by holding /me behind a controlled release. (6) Closed by the E5 → E4 → settings proof and existing guard/returnTo.
DECISION: Revision 3 dispositions in original order. (1) Closed by V2's screen-owned read and ACCT-3. (2) The old route cache trap is removed by V4a and the S3 recovery points to a new link. The S5 retry shortcut still strands the already-verified viewer on an unused link. (3) Closed by the explicit per-mount lifetime and fresh-instance proof for a used token. A fresh mount repeats the POST, and its result follows the server's actual response. (4) Closed for the first reconciliation read by keeping the form mounted. The failed-read/manual-recheck path still needs a focus destination when it restores the same form.
NEXT: Record the remaining implementation-blocking plan corrections and verify the working file.

[logical-gaps] S5 retry bypasses S3's verification condition
where: §18 V2, ACCT-5 and its proof row (plan.md:614–618, 689, 713).
why: V2 correctly places an already-verified viewer in S5 after an ambiguous failure. The retry prose then sends any signed-in viewer with a verified /me result to S3, so that viewer never retries a potentially unused link. The same shortcut also applies to a definitive 403. For a claimed transition to verified, the root snapshot can be cached (root-identity.ts:26–32), and the write acts on the token's user independently of the viewer (accounts.go:95–106).
action: Apply the same S3 condition on retry and in the table. If using a before/after transition, require fresh observations of the same account around that attempt. When the observations cannot establish a transition, keep a usable token retry. Give token_used the same S3/S4 precedence in the prose and acceptance. Test already-verified viewer → ambiguous failure or 403 → retry → second POST.
weight: blocking

[logical-gaps] A successful manual 2FA recheck can remove the focused retry control
where: §18 V5a and ACCT-9 (plan.md:649–653, 717).
why: A failed /me moves focus to «Перевірити ще раз». If that action succeeds and returns the original 2FA state, the plan applies the same-form branch, which assumes the submit button still holds focus. Removing the focused retry control loses focus despite retaining the form (the prior review's unmount issue on a different branch).
action: Keep the retry control focused while its read runs. If the original form returns, focus its submit button when retiring the retry control. If the card changes, use the existing replacement target. Add a browser assertion for failed read → manual retry → unchanged form, including final focus.
weight: blocking

[logical-gaps] 2FA rejection responses do not establish the UI state the plan relies on
where: §18 V5a, ACCT-9 (plan.md:649–656, 717).
why: Protected 2FA POSTs can return authentication_required, while useRedirectWhenSessionEnds watches session.data === null and a rejected mutation does not update that query (handler.go:122–126, require-identity.ts:30–51, settings.account.tsx:31–62). The existing hook cannot redirect from the POST alone. The API also returns twofactor_already_enabled on Confirm and twofactor_not_enabled on Disable/Regenerate (two_factor.go:84–91, 211–231, 258–265), proving the form is out of date. Neither code has a defined reconciliation path.
action: Specify which mutation errors initiate a fresh /me read. Establish null before relying on the redirect hook, and reconcile the card when the server reports the opposite enabled state. Add ACCT-9 checks with cached identity for expired-session POST and the 409 responses.
weight: blocking

FACT: Static review at 08fc8ae7 covered the requested plan, audit, prior logical-gaps findings, frontend and backend code, and accounts browser spec. Runtime checks were not run. Only this working file was edited.
NEXT: None. 1 lens assigned / 1 returned.
lens verdict: 3 blocking findings
