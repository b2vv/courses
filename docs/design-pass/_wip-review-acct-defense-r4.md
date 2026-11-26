NEXT: Read the shared entry rules and verify the repository root.
FACT: The shared entry rules were read and the repository root is /Users/strelia/projects/courses.
NEXT: Read the local routes and repository briefing.
FACT: Local routing and the repository briefing were opened. This is a read-only defense of the pinned plan.
NEXT: Read the canonical defense contract, locate revision 4 and extract the previous findings.
FACT: 08fc8ae7 is on main. Read the five-question contract and all three prior defense logs. Revisions 2 and 3 each left dependency and race rows open.
DECISION: Put the table in the authorized working file, with citations to the pinned plan.
NEXT: Read the full pinned section and audit and isolate the revision 4 changes.
FACT: The audit provides real invalid-link and guest-resend responses, with mocked success and 2FA states. Revision 4 replaces settled mutation retention with route-instance ownership, adds no-mail recovery, and changes E5 copy and 2FA recovery instructions.
NEXT: Read the exact decision and acceptance lines in bounded sections.
FACT: Checkout HEAD is 08fc8ae7 and no reviewed source or plan changes were reported. V4a explicitly bounds reopen behavior and excludes cancellation by an unmounted screen.
NEXT: Trace routes, view wiring, session adapters, and backend write and error paths.
FACT: Server paths confirm invalidate-before-issue/send, consume-before-account-write, and delete-before-reinsert for recovery codes. Current routes and settings use independent mutations and a shared session key.
NEXT: Verify the views, CSS, error mapping, query ordering and existing browser proof.
FACT: The current adapter preserves 409 2FA state conflicts and 401 authentication_required. The redirect hook reacts to cached null, not to the mutation error itself. Current browser proof uses Mailpit and real ten-code enrollment but does not cover the newly planned branches.
NEXT: Check the minimum counterexamples and whether existing recovery paths are named in the plan.
FACT: The remaining resend claim can be falsified by both requests invalidating before either issues. Recovery-only regeneration can leave no usable proof after deletion. A new V2 inconsistency sends an already-verified viewer from S5 to S3 on retry even when S3’s table predicate still fails.
NEXT: Confirm exact proof references for remount ordering and check the remaining brief context.
FACT: Reviewed §18 at 08fc8ae7 on main, the source audit, the three previous defense logs and the named frontend/backend/e2e sources. Source review only. No Make, Docker or e2e ran.

FACT: Revision 1 follow-up. Guest validation and localized feedback remain closed (603, 658, 711, 718). V1 gives the same neutral 204 acknowledgement to both viewer types and explains the no-letter fallback (598). V14 retains manual copy on clipboard failure (677, 721). Token-consumed retries have S4/S3 or E4 recovery (614–618, 627–630), with the S5 retry branch inconsistency below. V1 retains one owner for the displayed result and the shared same-tab guard (605, 710). The rollback clause carries forward the effects that survive reverting the UI (734).

FACT: Revision 2 follow-up. V5a keeps the form mounted during reconciliation and gives an explicit read-only retry if both the write and /me fail (649–655, 717). V1's event-time isMutating check and two-trigger check address the earlier rendered-state guard (605, 710). V2 starts a new /me after cancelling the older request (618, 713). V4a provides one automatic submission per mounted instance, with a StrictMode test (633, 715).

FACT: Revision 3 follow-up. V1 closes the no-letter action (598), but the previous-link guarantee remains open (598, 605). V4a explicitly allows repeat POST after remount, closing the five-minute mutation-retention objection (633–634, 691, 715). V2 now denies an unmounted A permission to cancel B’s session query and ACCT-7 checks that ordering (618, 715). V5a still lacks a path for a user whose only usable proof was a recovery code (656).

FACT: Dependency counterexample. A user without the authenticator submits a valid recovery code. Authorization consumes it, then regeneration deletes the whole old set. A failed insert or lost response after deletion leaves no usable code to enter, even if an old code was never entered. The instruction at 656 offers only proofs the user may no longer possess (two_factor.go:227–241, 258–281). Name a fallback or explicitly bound the lockout case and its consequence. The existing administrator disable operation is at two_factor.go:244–255 and admin_handler.go:153–167. A regeneration 401 has a related recovery omission: 656 refers to useRedirectWhenSessionEnds, which only redirects on cached null. The mutation’s authentication_required is separate from /me. The plan lacks the read or transition that drives the hook when a session expires between page load and submit (session-repository.ts:79–81, 270–281, require-identity.ts:30–51).

FACT: Repeated resend counterexample. A and B can both invalidate before either issues a token. Both new links can then remain live. Even sequential sends can arrive out of order. Receipt of a letter cannot establish V1’s claim that prior links are already invalid or that the link in the arriving letter is usable. ACCT-2 tests a single sequential resend. Scope the copy to the guarantee the API makes (598, 605, 686, 725–730; accounts.go:109–140; generated/identity.sql.go:474–480).

FACT: New V2 conflict. For an already-verified viewer who gets a timeout on a link for another address, line 614 keeps S5. With the same session and no intervening write, the retry shortcut at 618 now sends them to S3 because /me still says verified. The same shortcut applies after a 403. Use the same predicate for both decisions or explicitly define the different outcome on retry (614–618, 713).

FACT: V4a states a predictable result for reopening an already-used token. The broader wording at 634 also covers a return while the first POST is pending. On A→B→A the new instance sends a second A POST. If the first A consumes the token and pauses before the write, the second A can receive token_used and the active screen can read /me before the first A commits. The now-unmounted first A's eventual result is ignored. The plan needs to bound that timing or define the recovery when the newest screen reports a dead link while the first attempt is still writing (618, 633–634, 715; accounts.go:95–106, 245–265). Separately, existing 409 twofactor_already_enabled/twofactor_not_enabled are valid outcomes when another tab changes 2FA first. V5a has no reconciliation outcome for those state conflicts (649–656; two_factor.go:84–90, 258–265; handler.go:851–854).

DECISION: Count gaps by fixed question. Rows 2 and 3 remain open. Record the result here under the read-only instruction.
NEXT: Return the five-row table and verdict.

| # | Question | §18 at 08fc8ae7 | Verdict |
|---|---|---|---|
| 1 | Malformed input | 603, 615, 627, 658, 711, 718. Guest email has required, format and length checks with localized feedback. Missing and rejected tokens get S4/E4, and resend 400 gets a general alert at 601. | Closed |
| 2 | Dependency failure | 598, 613–618, 628, 649–656, 677. No-mail retry, failed session read and clipboard refusal have a visible path. Regeneration can discard every usable proof for a recovery-only user, and its session-expiry branch names a redirect hook without driving the required session state. | **Gap** |
| 3 | Race and repeated invocation | 598, 605, 614–618, 633–634, 649–656, 710, 715. Same-tab resend, StrictMode and stale screen reads are bounded. The prior-link claim still exceeds the API guarantee. S5 retry can bypass S3’s predicate. The plan does not bound reopen while the original POST is still writing or reconcile 2FA state conflicts. | **Gap** |
| 4 | Empty and boundary data | 603, 613–618, 627, 667–669, 677–679, 703. Covers missing token or session data, cleared forms, one-time code display and full fixed-size secret readability. The adapter requires ten recovery codes (session-schema.ts:19–24). | Closed |
| 5 | Rollback | 734. Revert the frontend slice. The plan names sent mail, invalidated links, consumed tokens, credential changes and 2FA changes as effects that survive. | Closed |

**Verdict: 2 gaps (rows 2 and 3).**
