NEXT: Read the AI Factory entry rules, then verify the repository root and load the edge-cases contract.
FACT: Read the shared agent entry rules. The initial NEXT was issued before the first tool call.
FACT: Verified repository root, read local routing and located DP2 in the briefing and tracker.
DECISION: Review only edge cases in efc0bcf9 and use only the assigned working file for progress and findings.
NEXT: Read the fixed plan, audit and prior edge-case findings.
FACT: Confirmed the requested commit is on main and recovered the R1/R2 findings.
NEXT: Read the remaining revision-3 acceptance clauses and the owner answers in full.
FACT: Owner approved all 16 audit items and the A6 label. Revision 3 adds cache-based token attempts, post-write session checks, and new transient 2FA phases.
NEXT: Compare token and session claims with the pinned routes and identity adapters.
FACT: Read ACCT-1 through ACCT-20 and the Mailpit journey. The first route-read command shadowed zsh PATH; retrying with a different loop variable.
FACT: Confirmed token auto-submit and success-refresh behavior, the shared current() query, and the independent resend entry points.
NEXT: Verify query-cache behavior, partial account writes and existing browser fixture mechanics.
FACT: Confirm, disable and regenerate each have partial-write boundaries. The settings route refreshes a shared session query and Playwright uses a single account for the verification journey.
NEXT: Read UI branches and verify cache eviction, client deadlines and error-classification boundaries.
FACT: React Query cache subscribers do not keep mutation entries alive. The installed useMutationState also refreshes its selected results only on a cache notification. The new 2FA reconciliation phase hides the focused form.
NEXT: Check production query defaults, token handlers and the exact revision-1 titles before weighing findings.
FACT: The tracked tree is clean and the reviewed code matches the pinned commit. The token handlers have a definitive 403 origin refusal before the write.
NEXT: Confirm the router defaults and distinguish the original reported transitions from newly introduced cases.
FACT: The router config sets query defaults only, so mutations use the installed five-minute client GC default. The read-only source review covered the requested route, UI, adapter, backend, and E2E files.
DECISION: Revision 1 dispositions in original order.
- Resend 204: closed by V1's conditional, audience-neutral acknowledgement and distinct Mailpit proof.
- Retained invalid cookie: closed for the original one-cookie case by V1/ACCT-3. Duplicate-cookie 400 has generic feedback. Unavailable identity has S0.
- Coexisting resend controls: closed by the synchronous shared guard and the same-render-window test in ACCT-1.
- Consumed links: closed for the original outcome ambiguity. S4 allows a new link, S3 checks current identity, and E3–E5 direct the user to check the stored address. A definitive origin rejection still needs its own covered outcome.
- Link swap: partly closed. V4a defines the right behavior and ACCT-7 covers uncached tokens and in-flight reopen. The selected subscription and retention gaps below leave settled/removed-token transitions open.
- Stale 2FA after Done: closed for the acknowledged write path by immediately seeding the session cache before background refresh. Ambiguous writes now have recovery, with a separate focus gap below.
- Copy state: closed by component-local feedback and the delayed/failing clipboard tests.
- Standalone enable width: closed in the plan by the direct-child selector and ACCT-15.
- Mailpit identity: closed by excluding the original message ID and opening the old link before confirming the replacement.
- Validation scope: closed by binding the fallback to TextField/TextArea and excluding Select/ComboBox.

DECISION: Revision 2 dispositions in original order.
- Resend copy: closed by the conditional 204 acknowledgement.
- Same-tick resend: closed by the synchronous guard and pre-render double activation check.
- Pre-write /me and failed A3 read: closed for the reported overlap and rejection by cancel-then-read and S0 retry.
- Pending-token reopen: closed for duplicate POSTs by the synchronous cache lookup. Reading the right cached state on reopen and retaining it after settlement are still open.
- Failed 2FA reconciliation read: closed for the indeterminate state and recheck path, including signed-out exit. The transitional focus path is still open.
- Failed code regeneration: closed by the warning about old codes and the instruction to use a fresh authenticator code.
- Edits lost during save: closed by read-only fields during the write.
- Secret clipping: closed in the plan by content-based height, fallback capacity and the visible-last-character/raw-value check at the acceptance boundary.

[edge-cases] A cached token can display the previous token's result
where: §18 V4a and ACCT-7, lines 623–624 and 704.
why: Installed React Query useMutationState.ts:71–98 initializes its selection once. Changing the filter only sets an options ref. The selection is recalculated on a cache notification. A same-instance switch from A to already-cached B creates no new mutation, so A's result can remain visible indefinitely. Removing the token also needs to clear the prior display without waiting for a notification.
action: Require the displayed selection to update synchronously with a token change or removal even when the cache is quiet. Test navigation between two settled cached tokens with different outcomes and then removal of the token, with no extra POST.
weight: blocking

[edge-cases] Cache eviction can resubmit a settled one-time link
where: §18 V4a and ACCT-7, lines 623–624 and 704.
why: router.tsx:35–37 sets query defaults only. Query Core removable.ts:24–29 and mutation.ts:128–162 remove unobserved settled mutations after five minutes on the client. useMutationState observes cache events without retaining the mutation. Leaving A6 after an acknowledged change and reopening the same link after eviction in the same router can send a second POST and replace E2 with E4. V4a promises that only an explicit action starts another attempt.
action: State a retention policy that preserves the promised result for the router lifetime, or explicitly specify the bounded reopening behavior. Verify reopening after the default GC interval using controlled time.
weight: blocking

[edge-cases] A failed 2FA write can drop focus while the check is pending and after the form returns
where: §18 V5/V5a and ACCT-8/9, lines 628–644 and 705–706.
why: V5a unmounts the focused confirm or disable control while the session check runs. It gives no focus target to the checking phase and says focus does not move when confirm returns to setup with an error. The original node is gone and the newly mounted form cannot keep its focus. The disable-still-enabled exit has the same restoration gap. AccountSettingsView.tsx:146–205 shows the conditional form branches that the new phase interrupts.
action: Name a focus target while the read is pending and restore focus to the relevant remounted form or error when the write did not apply. Check keyboard focus during a held /me and after release for both non-applied exits.
weight: blocking

[edge-cases] A definitive token request refusal has no state
where: §18 V2/V3 and ACCT-5/6, lines 598–619 and 702–703.
why: handler.go:115–121 and 514–523 guard both token actions and can return 403 origin_forbidden before invoking the write. session-repository.ts:85–87 preserves that code. It matches neither the unusable-link states nor the listed uncertain outcomes. Sending it through S5/E5 instead would suggest the write may have applied and offer an identical retry even though the request was refused before the token was used.
action: Define an explicit refusal or fallback state with truthful copy and a recovery action for both routes. Include the named 403 in route coverage and continue to distinguish it from a consumed link or a lost response.
weight: blocking

FACT: Four open blocking findings. No runtime checks were run. Findings and progress were appended only to the designated working file.
NEXT: Return the scoped findings and lens verdict.
DECISION: lens verdict (4 blocking findings).
