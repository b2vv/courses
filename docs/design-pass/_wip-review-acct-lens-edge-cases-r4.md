NEXT: Read the shared entry rules and verify the repository root.
FACT: Verified the Courses root and read the shared entry rules.
NEXT: Read local rules, the briefing and the lens contract.
FACT: Local rules and lens contract loaded. Main resolves to 08fc8ae7.
DECISION: Review only edge cases and record only in this file.
NEXT: Read revision 4, the audit and earlier edge-case findings.
FACT: The audit and owner answer require all 16 items and the title inconsistency.
NEXT: Read the full scoped plan and historical verdicts in bounded sections.
FACT: V1–V5a specify fresh session reads, per-screen token state and a mounted form during ambiguous 2FA writes.
NEXT: Complete the plan read and enumerate earlier findings.
FACT: Read V6–V15 and the verification matrix. Revisions 1 and 2 cover resend, token, 2FA, validation and layout boundaries.
NEXT: Read revision 3 findings and the pinned plan delta.
FACT: V4a drops the cross-mount promise and ACCT-7 covers settled reopen, while V5a retains the focused form.
NEXT: Verify route, session and backend behavior against the revision.
FACT: Token handlers consume before writing. 2FA checks distinguish already enabled and no longer enabled, and consume proof before state updates.
NEXT: Check frontend error codes, background refresh, UI structure, CSS and test fixtures.
FACT: The identity adapter preserves authentication_required, twofactor_already_enabled and twofactor_not_enabled. The session-end hook observes the shared query and does not react directly to mutation errors.
NEXT: Verify retry branching, cache scheduling, and shared validation boundaries.
FACT: V2 retry can bypass the S3 predicate. A mounted 2FA form preserves focus on the initial read, but a later recheck starts from a different button. The existing browser test uses real account changes and scoped Mailpit reads.
NEXT: Confirm exact source lines and record individual dispositions.
DECISION: Revision 1 dispositions in original order.
- Resend 204 closed by V1's conditional wording and separate Mailpit proof.
- Retained invalid cookie closed for the reported path by V1 and ACCT-3. The plan explicitly gives duplicate-cookie 400 generic feedback.
- Coexisting resend controls closed by the dispatch-time cache guard and ACCT-1. The installed mutation implementation dispatches pending before its first await.
- Consumed-link recovery closed for the original ambiguity by S3/S4 and E3–E5. The separate S5 retry finding below concerns the revised state predicate.
- Cross-token result leakage closed by token-keyed screens and ACCT-7. Pending reopen is a separate duplicate-request case.
- Done with delayed session refresh addressed by V5a's cache write and ACCT-8. Ambiguous-write recheck has the focus, auth and concurrent-change gaps below.
- Stale copy feedback closed by V14's child lifecycle and late-copy test.
- Enable button width closed by V11's direct-child selector and ACCT-15.
- Mailpit message selection closed by excluding the original ID and exercising both links.
- Validation scope closed by V6's explicit TextField/TextArea boundary.

DECISION: Revision 2 dispositions in original order.
- 204 wording and same-tick resend both closed by V1 plus ACCT-1/2.
- Pre-write or failed A3 read closed for the original request-overlap case by cancel-then-read and S0, with the retry classification gap below.
- Pending reopen reopened. Revision 4's limit covers the settled outcome, whereas the original report specifically holds the first request in flight.
- Unknown 2FA while its session read fails addressed by the locked form and retry. The retry focus exit and direct auth response remain actionable.
- Ambiguous regeneration copy closed by V5a's old-code warning and fresh-proof wording.
- Post-submit edits cleared by success addressed by V10's read-only pending fields.
- Secret capacity addressed by content height and ACCT-19's full-value and no-internal-scrolling checks.

DECISION: Revision 3 dispositions in original order.
- Cached selection after a token change closed by the keyed instance. Removing a token selects the no-token instance as well.
- Settled mutation eviction closed by the explicit per-instance resubmission limit and ACCT-7.
- Initial pending 2FA focus and the immediate unchanged exits closed by retaining the form. The new focus target for a failed read needs a separate exit.
- Definitive refusal classified and covered in both initial state tables. A3's retry can still override that classification as described below.

[edge-cases] Reopening a pending link can replace its eventual success with a failure
where: §18 V4a and ACCT-7, lines 633–634 and 715.
why: In A→B→A while the original A request is pending, revision 4 starts a second A request and discards the first result. If the first A consumed its token but is still writing, the second immediately gets token_used (accounts.go:95–106, 245–265). A6 can display E4 even when the discarded request completes successfully. ACCT-7 only reopens an already-used link after settlement.
action: Specify how a reopened screen waits for or reuses an existing pending attempt for the same route and token. Retain the documented settled-reopen behavior. Hold the first A in flight in an A→B→A route test and assert a single A POST and the eventual A outcome.
weight: blocking

[edge-cases] S5 retry can turn a definitive refusal into confirmation
where: §18 V2 and ACCT-5, lines 614–618 and 713.
why: After origin_forbidden, an already-verified signed-in user properly sees S5. V2's Retry then reads /me and sends any verified user to S3, although handler.go:514–523 refused the token write and S3's predicate excludes this case. The same shortcut changes an ambiguous failure for an account already verified before the attempt into S3. A cached pre-attempt snapshot from a different account also cannot establish the required unverified-to-verified transition.
action: Apply the same S3 predicate on the initial reconciliation and on Retry, comparing pre- and post-attempt identity for the same user when inferring a transition. Cover a verified user retrying a refused link and a verified user retrying an ambiguous failure.
weight: blocking

[edge-cases] Session expiry during a 2FA write can leave protected controls on screen
where: §18 V5a and ACCT-9, lines 649–656 and 717.
why: A direct 401 authentication_required from a guarded 2FA mutation can arrive while the shared session cache still contains a user. The adapter preserves the code, but useRedirectWhenSessionEnds only navigates once session.data is null (require-identity.ts:30–50). V5a relies on that hook and describes a fresh /me for ambiguous errors. It gives no cache or navigation transition for a direct 401, including the regeneration case where it promises a sign-in redirect.
action: Define what a direct authentication_required triggers so the live redirect can run with a stale signed-in cache. Cover an expired session during a 2FA command and distinguish the 401 recovery_code_invalid response.
weight: blocking

[edge-cases] Concurrent 2FA changes leave stale controls after a state conflict
where: §18 V5a and V7, lines 647–663.
why: A second tab can enable 2FA while setup is open, or disable it while the first tab shows both enabled forms. Confirm returns twofactor_already_enabled and disable/regenerate return twofactor_not_enabled (two_factor.go:84–91, 211–229, 258–265; handler.go:851–854). The adapter preserves both codes. Neither is an ambiguous write nor a named reconciliation outcome, so group-local feedback can leave a form that will fail again in the same state.
action: Route the named state conflicts through reconciliation and specify the resulting card and focus. Cover setup followed by already-enabled and an enabled form followed by no-longer-enabled.
weight: blocking

[edge-cases] A successful 2FA recheck can remove the focused retry button
where: §18 V5a and ACCT-9, lines 649–653 and 717.
why: The first failed read explicitly moves focus from the submit button to «Перевірити ще раз». When the retry confirms that the original write did not apply, the retry control disappears and the mounted form becomes available. The instruction to leave focus on the submit button only works on the first reconciliation path, where it was still focused.
action: On an unchanged-state result after a manual recheck, move focus from the disappearing retry action back to the restored submit control. Assert the unsuccessful read, manual retry and unchanged-state exit for both confirm and disable.
weight: blocking

FACT: Static inspection only. No make check, Docker or e2e ran. Only this working file was appended.
DECISION: lens verdict (5 blocking findings).
FACT: Verified five fixed-format blocking findings and no tracked changes.
NEXT: Return the scoped verdict.
