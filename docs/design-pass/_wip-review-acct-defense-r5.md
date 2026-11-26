NEXT: Read the shared entry rules, verify the repository root, and load the review contract.
FACT: Read the shared entry rules and local repository routes.
NEXT: Read the plan-defense contract, briefing, and task pointer.
FACT: Loaded the plan-defense contract, repository briefing and task pointer, and checked the revision reference on main.
NEXT: Read the pinned plan section, source audit, and four prior defense files.
FACT: Read the pinned ACCT plan, the A3/A6/A7 source audit, and the four prior defense records.
NEXT: Trace the audited routes, controls, session behavior and account handlers at the reviewed base.
FACT: Captured the line-numbered V2–V8 decisions and ACCT acceptance and rollback rules.
NEXT: Extract the previous findings and verify their implementation assumptions.
FACT: Indexed prior findings and verified route and handler entry points against the pinned source.
NEXT: Check backend effects and frontend caching, form semantics, styles and test seams.
FACT: Verified the source audit against route/view wiring, field and CSS primitives, API adapters, write order, and current e2e selectors.
NEXT: Test revision 5 recovery and ordering claims against concrete source-level counterexamples.
DECISION: Count gaps by fixed question. The 204 copy and 2FA ordering leave row 3 open. Check the scope of ambiguous account-write recovery for row 2.
NEXT: Record disposition for all four prior defenses, new findings, and the five-row verdict.
FACT: Confirmed that enrollment also returns the 409 state conflict and that a manual Query cache write leaves an existing fetch active. Checked the shared origin check and the settings forms’ existing error paths.
NEXT: Append the final evidence and verdict.

DECISION: Use the five fixed questions in order and count each open row once. Citations below use the line numbers of docs/design-pass/plan.md at 6a8d1a3e. The owner authorized writing only this file.

### Earlier findings

FACT: Revision 1. Guest validation and localized feedback are specified (610–612, 624, 636, 669). V1 now gives both viewers the same receipt acknowledgement and tells them what to do if no mail arrives (607). Clipboard refusal has a manual alternative (688). V1 gives same-router resend a synchronous guard and assigns feedback to the initiating surface (614, 721). V2/V3 specify recovery after an ambiguous token write (625–627, 636–639). The rollback clause lists the effects that survive (745). The concurrent resend validity claim remains open below.

FACT: Revision 2. V5a names the visible recovery when a 2FA write and /me both fail (658–662) and warns about lost recovery codes (667). V1 checks same-render overlap at the event and in its unit scenario (614, 721). V2 restarts /me after the relevant token attempt, V4a deduplicates StrictMode initiation, and ACCT-5/7 test both (627, 642, 724–726).

FACT: Revision 3. The no-mail action is explicit (607). The recovery-code-only path now names the administrator reset operation (667; handler.go:134, admin_handler.go:153–167, two_factor.go:244–255). V4a bounds settled results to the screen instance and names the behavior on reopen, so the five-minute mutation retention issue no longer applies (642–643, 702, 726). A stale instance may neither act on its result nor cancel the active instance's read (627, 726). The 204 copy still claims more than the server guarantees across tabs (607, 614).

FACT: Revision 4. Same-token retry now goes through the same S3/S4 predicate, so a verified session cannot short-circuit S5 (623–627, 700, 724). The in-flight map binds a remount to the existing promise (642–643, 702, 726). Explicit cached null is now wired to the redirect hook, closing the missing transition from authentication_required (665–666, 704, 728); the new write still has the in-flight read race below. Confirm/disable/regenerate 409s have a reconciliation path (664, 728); the same 409 from enroll remains uncovered.

### Open findings

FACT: Dependency failure at A7. V10 defines the success treatment after a 204 and V5a defines session expiry (665, 680, 709, 731). A password write can succeed before deleting other sessions fails, leaving a changed password with an error response (accounts.go:195–222). A request for an email change discards mailer failure after issuing the token and returns 204, so V10's success output can direct the user to confirm at the new address when there is no mail (accounts.go:225–242; AccountSettingsView.tsx:108–113). V1's no-letter fallback is for verification resend only (604–607). A7 needs observable recovery for both outcomes or an explicit scope boundary. Also clarify whether E5 uses S5's full reload after origin_forbidden: the tables share the error condition, but only S5 names reload and ACCT-6 specifies only the E5 copy for 403 (625, 637, 725; handler.go:115–121, 514–523).

FACT: Resend ordering remains open. Sequential invalidate→issue does not imply that the latest arriving letter is the last issued link. Concurrent resends can both invalidate before either issues, leaving both new links live. The service makes separate calls in this order (accounts.go:109–140; repository.go:321; backend/sql/queries/identity.sql:122). Changing the copy to “найновішого листа” leaves the categorical invalidation claim at 607 and 614. ACCT-2's one sequential resend (697, 736–741) cannot support it across tabs.

FACT: Expiry reconciliation can lose to an older session read. /me may have captured the viewer before the session expired, while the write later returns authentication_required. The cache write at 665 does not cancel the pending read. TanStack Query setQueryData is a manual success that leaves the fetch active (queryClient.ts:202–211, query.ts:656–669). A delayed authenticated result can replace null, and the effect at require-identity.ts:38–51 redirects only when its observed data is null. ACCT-9 seeds a cached viewer but never holds and releases an older /me across the 401 (728). Define which event wins, with an ordering check.

FACT: 2FA reconciliation misses enrollment. A stale disabled card can show “Увімкнути 2FA” after another tab enables it. The enroll service then returns the very same 409 twofactor_already_enabled as confirm (two_factor.go:58–65, 84–90; handler.go:201–212, 851–854; session-repository.ts:94–98). V5a sends only confirm's 409 through reconciliation (664, 728). The visible card can remain disabled with no refresh path in the plan. The settings card also needs an explicit same-router serialization rule or scope limit for overlapping 2FA writes. Its existing button and form guards use the last rendered pending value (settings.account.tsx:63–64; AccountSettingsView.tsx:48–61), and V5a blocks new writes during reconciliation without specifying the event-time guard when the writes start (658). Enroll replaces the pending secret and regeneration replaces the entire code set (two_factor.go:58–81, 227–241), so a late response can display material a different request replaced.

FACT: Source review at 6a8d1a3e on main. The audit ran real invalid-token and guest-400 cases and mocked most success states (audit lines 9–11, 188–199). This review used the named route, UI, styles, identity client, backend and e2e source files. No make check, Docker or e2e ran.

| # | Question | Plan lines and defense | Verdict |
|---|---|---|---|
| 1 | Malformed input | 610–612, 624, 636, 669, 722, 729. Guest address validation blocks missing, malformed and overlength input with localized feedback. Missing or rejected tokens get S4/E4. Resend 400 gets a general alert. | Closed |
| 2 | Dependency failure | 607, 622, 625–627, 637, 658–667, 680, 688. V1's no-mail fallback, failed session reads, lost 2FA proofs and clipboard refusal have recovery paths. A7 still lacks account-write recovery when the password applied before an error or the email-change request returned 204 without mail. E5 leaves the 403 retry mechanism unclear. | **Gap** |
| 3 | Race and repeated invocation | 607, 614, 623–627, 642–643, 657–667, 721, 726, 728. The guard and map address same-router resend and token overlap. The cross-tab link-validity claim, /me ordering after a write returns 401, and settings 2FA overlap and enroll conflict remain open. | **Gap** |
| 4 | Empty and boundary data | 612, 623–627, 636, 669, 678–680, 688–690, 733. The plan covers missing tokens and session values, cleared forms, hiding one-time codes, and the full fixed-size secret at the specified widths. | Closed |
| 5 | Rollback | 745. Revert the frontend slice. Sent mail, invalidated links, spent tokens, credential changes and 2FA effects already applied through the API survive. | Closed |

**Verdict: 2 gaps (rows 2 and 3).**
