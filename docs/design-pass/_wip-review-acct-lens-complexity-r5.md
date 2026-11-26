NEXT: Read the shared entry rules, then verify the repository root and load the local rules.
FACT: Shared entry rules read. This is a single complexity lens with one append-only file.
NEXT: Verify repository root, revision, and local instructions.
FACT: Verified repository root and HEAD 6a8d1a3e, and read the local rules and design-pass task pointer.
NEXT: Read the installed complexity lens contract and the ACCT sources.
FACT: The installed contract defines this lens as reuse-first and ranks rework as blocking. The working tree has no tracked changes.
NEXT: Compare §18 at revision 5 with owner decisions and each prior complexity verdict.
FACT: The owner approved all 16 audit items and the title alignment. Revision 5 targets the prior S3/retry blocker and adds coordination for in-flight link submissions.
NEXT: Read the current decision text and isolate all prior complexity findings.
FACT: V2 now sends ambiguous failures straight to S5 and applies one S3 predicate after a terminal result. ACCT-5 includes a second POST for an already-verified user and a 403 with unavailable /me.
NEXT: Check the current route/session mechanisms and the new V4a and V5a paths against the source.
FACT: Both link routes currently use route-local useMutation observers; an observer disappears on unmount. The backend confirms terminal link failures and the 2FA conflict codes.
NEXT: Check how existing composition handles shared reads, sign-out, and protected-route redirects.
FACT: EnrollTwoFactor and ConfirmTwoFactor return the same already-enabled error; disable and regeneration share the not-enabled error through authorizeFactor. V5a names only three of the four callers.
NEXT: Check the adapter and plan verification rows for any existing implicit coverage of the omitted start action.
FACT: The adapter preserves both 409 codes. ACCT-9 covers confirm, disable and regeneration conflicts, and omits enroll. The plan baseline has no frontend/backend delta at the reviewed commit.
NEXT: Confirm the existing UI paths and finish the prior-finding disposition.
DECISION: Rev 1 closed 4/4. ACCT-1 tests resend results once plus caller wiring (:721). V6 scopes fallback to text controls (:669). V7 keeps each form's existing mutation as its error source (:671). V11–V13 extend the existing CSS (:682–686).
DECISION: Rev 2 closed 3/3. V2 cancels the older session fetch before using current() for S3/S4, with the held-read proof in ACCT-5 (:627, :724). V14 uses the recovery panel lifetime, with delayed copy and new-set checks (:688, :732). V13 extends the existing A5 margin (:686).
DECISION: Rev 3 closed 3/3. V4a holds only in-flight work and explicitly permits another POST after a completed attempt, with both re-entry cases in ACCT-7 (:642–643, :726). V5a preserves the mounted form and its focus during the read (:658–662, :728). V1 derives shared pending state from the mutation cache (:614, :721).
DECISION: Rev 4 closed 1/1. V2 uses one S3 predicate after a terminal result, and S5 Retry sends the token without a session precheck. ACCT-5 verifies retry for an already verified user and 403 with unavailable /me (:623–627, :724).
NEXT: Record the new V5a finding and return the single-lens verdict.

[complexity] Reuse the 2FA conflict recovery for every action that returns either state code
where: docs/design-pass/plan.md:664 (V5a) and :728 (ACCT-9).
why: V5a dispatches the same reconciliation by action and code, omitting the setup-start action. EnrollTwoFactor and ConfirmTwoFactor both return twofactor_already_enabled (backend/internal/identity/application/two_factor.go:58–65, :84–91), and the adapter preserves that code. With stale disabled state after another tab enables 2FA, clicking «Увімкнути 2FA» can leave the card showing disabled and offering the same rejected action. The route already has the same current() query for all four 2FA actions.
action: Define one route-level conflict rule for either state-specific 409 from any settings 2FA mutation. Reuse the planned fresh current() read and reconciliation notice, preserving the initiating control while the read runs, including the standalone enroll button. Add an enroll → 409 case to ACCT-9.
weight: blocking

FACT: Static review of main at 6a8d1a3e. Only this working file was edited. No other lens working file was read. No make check, Docker or e2e commands ran.
DECISION: 1 assigned / 1 returned.
Lens verdict: 1 blocking finding.
