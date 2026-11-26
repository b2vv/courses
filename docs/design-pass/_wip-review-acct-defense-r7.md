NEXT: Read the shared agent entry rules, then locate the revision 7 plan and review sources.
FACT: Shared entry rules read and Courses repository root verified.
NEXT: Read local rules, the briefing and the exact plan-defense contract.
FACT: HEAD and main equal ab3cae05. Read the five-question contract and project scope.
DECISION: Record the gate in the working file under the owner’s read-only instruction.
NEXT: Read the pinned plan, the six defense reviews and the source audit.
FACT: The audit distinguishes its live invalid-link and guest 400 results from mocked success and 2FA states.
NEXT: Read the operative plan with stable line numbers and extract each prior defense verdict.
FACT: Prior defenses converge on rows 2 and 3. V4b/V4c and the password recovery text are the material revision 7 changes.
NEXT: Verify session and write ordering against the named frontend and backend sources.
FACT: V4b waits for an older /me and V4c moves reconciliation outside the route component. V5a patches the cached 2FA flag before background reconciliation. The new 2FA gate references an existing pending value.
NEXT: Read the guard and cache consumers, then compare error and overlap paths to the plan.
FACT: requireIdentity uses fetchQuery at staleTime 0 and throws a session dependency error on failure. Current twoFactorPending is an OR of four rendered observer results. Password change writes before revoking other sessions.
NEXT: Verify default route recovery, login behavior, shared resend consumers, controls, styles and accounts evidence.
FACT: The new A6 fallback delegates to the guard, which retries the session read. The password recovery instructs a new window and the 2FA race test requires two dispatches before a render.
NEXT: Confirm the actual user-visible fallback and test the consistency of the post-result rules by tracing the source.
FACT: Logged-in LoginView has only “Continue”, so another ordinary window cannot verify the password. The ACCT-9 test checks the hint and the error, leaving its next user action untested. The route guard supplies an existing error/retry path after a second failed A6 read.
NEXT: Resolve any conflicting A3/2FA lines and assemble the prior-finding closure ledger.
FACT: Prior findings checked against revision 7 (all plan lines below refer to ab3cae05).
- R1. Guest address shape and localized errors are covered at 626–628, 640, 652 and 687. Neutral resend acknowledgement and no-mail action are at 623. Manual copy is at 708. Shared resend ownership, its tab boundary, and retry after a spent token are at 630, 640–643 and 652–659. Effects surviving rollback are named at 765.
- R2. The 204 delivery claim is gone at 623. Failed confirm/disable plus failed session read has a read-only retry at 676–680. Regeneration after proof loss gets its own path at 685. Same-render resend is guarded and tested at 630/741. A3 waits out a pre-result read at 643/660/744. StrictMode joins an in-flight attempt at 658/722/746.
- R3. No-mail fallback is at 623. A recovery-only user with no usable proof has the administrator path at 685. Settled attempts have a documented remount boundary at 658–659. Competing route continuations no longer cancel each other's session reads at 660–661. Resend copy makes no cross-tab claim about token validity or mail arrival order at 623/630.
- R4. Regeneration's authentication_required follows the shared cache-null redirect at 683–685. Same-token S5 retry uses the same S3/S4 predicate at 639–643. Pending remounts join the original POST at 658. All four 2FA actions have 409 reconciliation at 682. The no-mail and concurrent resend issues remain closed at 623/630.
- R5. A3/A6 403 has navigation without retry at 641/653. A7's 204 email-change no-mail path is explicit at 700. Old /me before settings 401 is awaited at 660/683/748. Enroll's conflict is included at 682. Password recovery and 2FA dispatch are revisited below.
- R6. A6 now keeps the settings action after a failed refresh and delegates the second read to the settings guard (655/661). The guard reads at staleTime 0, throws on dependency failure, and the existing root error offers a GET retry on settings, so this recovery does not need to replay the POST (require-identity.ts:11–26, identity-queries.ts:25–34, __root.tsx:108–122). The pre-result data-less /me objection is addressed by wait-then-fetch plus the held-read test at 660–661/745. V10 removes the previous instruction that could produce password_unchanged, but its replacement still needs a reachable credential prompt. Same-render 2FA is now an explicit target at 689/748, with the mechanism and scope issues below.

FACT: Dependency failure. V10 at 700 says to verify the new password in “another window”. The authenticated session remains valid when the password write is followed by revocation failure, or when the response is lost after success (accounts.go:195–222). A normal window in the same browser profile shares that session. LoginView renders only “Continue” for a signed-in viewer (login.tsx:45–50, LoginView.tsx:35–42). Specify a private window or other signed-out browser context, and indicate that successful fresh sign-in confirms the password while revocation of other sessions can still be uncertain. Line 748 verifies the hint only.

FACT: Dependency failure. A3's 403 acceptance case at 744 explicitly requires S5 without a session read when /me is unavailable. V2 at 643 says S1/S5 do not require a read, while V4c at 661 requires one after any A3 result in the in-flight continuation. For 403 the write is rejected before token processing (handler.go:115–121, 514–523). Specify the A3 outcomes that schedule a refresh and keep the no-read test consistent. Clarify S2 when the read fails as well: its own row at 637 offers login even when the session could not be read, while 643 maps failure to S0 and 661 preserves the onward action after the failed read.

FACT: Race and repeated invocation. The two-before-render contract (689/748) is now explicit. Literal reuse of the existing twoFactorPending cannot meet it: that value is an OR of four rendered isPending values (settings.account.tsx:63–64, 76–86, AccountSettingsView.tsx:48–61, 231–234). Specify live synchronous admission through the write and reconciliation, as V1 already does at 630, and name its tab boundary. The 409 rule at 682 covers already changed state. It cannot cover two concurrent enrolls that both see disabled and overwrite the pending secret, or two regenerations that both authorize before separate deletes/inserts interleave (two_factor.go:58–81, 227–241). Define or explicitly bound the user outcome in those cases.

FACT: Race introduced by the revision 7 read rule. V5a immediately patches the /me cache on successful confirm/disable (675), while V4b lets a pre-write /me complete before the fresh read (660). A held false response can overwrite true after confirmation; if “Done” is pressed then, the existing card selects the disabled branch from the cached flag (settings.account.tsx:76, AccountSettingsView.tsx:131–146, 177, 223–234). A failed follow-up can leave the old value in place. The reverse applies to disable. Specify which result the card uses through the old response and a failed fresh read. ACCT-8 at 747 needs the held response to carry the pre-write flag to prove the transition stays stable.

DECISION: Count gaps by fixed question. This was source review of revision 7 on main against the audit and all six earlier defense records. The review record is the sole edit. No Make, Docker or e2e ran.
NEXT: Return the five-row defense and verdict.

| # | Question | §18 at ab3cae05 | Verdict |
|---|---|---|---|
| 1 | Malformed input | 626–628, 640, 652, 687. Reject missing, malformed and overlength guest addresses with localized feedback. Missing or rejected links reach S4/E4. A resend 400 gets the general alert. | Closed |
| 2 | Dependency failure | 623, 637–643, 655–661, 676–685, 700, 708, 744. Mail, clipboard, 2FA and A6 have recovery. The password check needs a signed-out context. A3 gives conflicting rules for session reads after 403 and for a failed read after 204. | **Gap** |
| 3 | Race and repeated invocation | 630, 643, 658–661, 675, 683, 689, 745–748. Resend, link remounts and initial stale reads are covered. 2FA still needs synchronous admission with a tab boundary and a rule to keep a successful write visible when an older /me completes. | **Gap** |
| 4 | Empty and boundary data | 628, 640, 652, 687, 698–700, 708–710, 734. Covers missing tokens and session values, cleared forms, one-time codes and the full secret at the named widths. | Closed |
| 5 | Rollback | 765. Revert the frontend commits. Sent mail, invalidated links, spent tokens, credential changes and 2FA effects survive. | Closed |

**Verdict: 2 gaps (rows 2 and 3).**
