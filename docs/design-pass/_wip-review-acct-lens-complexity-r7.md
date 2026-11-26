NEXT: Read shared entry rules, verify the repository and revision, then apply the complexity lens.
FACT: Shared entry rules read. The review record is the only writable deliverable.
NEXT: Verify local routing and read the skill contract.
FACT: Repository root verified. HEAD is ab3cae05 on main; local rules and the chosen lens contract are loaded.
DECISION: Review only the complexity lens, with the working file as the sole review record.
NEXT: Locate §18 and read the briefing, audit, and prior complexity findings.
FACT: Prior records identify 14 findings across revisions 1–6, including one CSS item carried into revision 2.
NEXT: Read §18 and the audit in full, with line numbers for comparison.
FACT: §18 and audit owner scope are loaded. Revision 7 centralizes read-after-result; V2 says S5 needs no read while V4c calls for all A3 results.
NEXT: Verify routes, composition, UI, CSS, backend outcomes and existing test seams at the reviewed checkout.
FACT: Source routes already share the exact session key; settings renders its 2FA branch from that cache. Current twoFactorPending is assembled from render-time mutation observer snapshots.
NEXT: Trace pending reads, the settings guard, and the 2FA mutation ordering against ACCT-8 and ACCT-9.
FACT: Settings protects its 2FA branch only with current(). Query setQueryData does not stop a fetch; V4b explicitly waits for older responses.
NEXT: Confirm the exact cache write sequence and review the revision diff and proof timing for actionable findings.
FACT: Query core writes a successful older response into the session cache even after a manual setQueryData; the held replacement-read check alone misses that window.
NEXT: Check when the pending-guard language was introduced and pin the finding references.
FACT: Revision 7 introduced the same-tick 2FA rule. The existing four mutation snapshots cannot serve as its imperative guard; V1 already names the cache primitive.
NEXT: Reconcile each historical disposition against the audit and current proof text, then record the final findings.
DECISION: Rev 1 closed 4/4. ACCT-1 tests resend outcomes once and caller wiring separately (plan.md:741). V6 limits validation to TextField/TextArea (:687). V7 owns errors in each action's mutation (:691). V11–V13 extend the existing CSS treatments (:702–706).
DECISION: Rev 2 closed 3/3 for its original cases. V2 and V4b require a read that starts after the old read settles, and ACCT-5 holds the old response (:643, :660, :744). V14 ties copy feedback to the mounted codes panel and owns its delayed-completion proof (:708, :752). V13 extends the A5 margin (:706).
DECISION: Rev 3 closed 3/3. V4a retains in-flight work only, with pending and completed re-entry in ACCT-7 (:658–659, :746). V5a keeps the initiating form mounted while checking an uncertain write (:676–682). V1 uses the existing mutation cache for synchronous resend admission (:630, :741).
DECISION: Rev 4 closed for user-visible classification. V2 uses one S3 condition, sends Retry directly, and uses the known refusal for S5 (:639–643, :744). Revision 7 reintroduces an unnecessary background request for known A3 refusals through V4c, recorded below.
DECISION: Rev 5 closed 1/1. All four 2FA actions share the conflict rule and keep the initiating control mounted (:682, :748).
DECISION: Rev 6 closed 2/2 for its stated cases. A6 uses the shared read boundary, keeps E2–E5 independent of read failure, withholds the settings action through settlement, and specifies the cold-cache proof (:655, :660–661, :745). One predicate owns all three uncertainty codes (:660).
FACT: Compared the pinned plan and audit against all requested source groups, the local Query implementation and the existing e2e specs. Static review only.

[complexity] Keep the confirmed 2FA phase through older session reads
where: docs/design-pass/plan.md:660, :674–675, :747 (V4b, V5a, ACCT-8).
why: V5a's immediate setQueryData uses the same cache field that currently selects the 2FA branch (settings.account.tsx:20, :76; AccountSettingsView.tsx:131, :177, :223). Under V4b, a pre-write /me is allowed to finish, and Query still commits its response after the manual cache patch (queryClient.ts:202–211; query.ts:575, :656–665). When that old response arrives after Done or successful Disable, the card can switch back and unmount the focus target while the fresh read is pending. A failure of the fresh read can leave the wrong branch visible. Holding only the post-write read misses this sequence.
action: Make the existing route-owned card phase retain the successful mutation outcome through reconciliation. Define when the post-write read can take over, with the last proven phase retained if it fails. Use the shared cache refresh after the older read settles. In ACCT-8, release a pre-write snapshot with the opposite 2FA value while the replacement /me remains held, and verify the phase and focus.
weight: blocking

[complexity] Reuse the resend admission check for the four 2FA writes
where: docs/design-pass/plan.md:630, :689, :748 (V1, one 2FA action at a time, ACCT-9).
why: Existing twoFactorPending is an OR of four render-time mutation snapshots (settings.account.tsx:63–64), and the four handlers close over that render (:83–86). Looking at that boolean at click time still admits two same-tick calls before rerender. The query cache already exposes synchronous pending state through isMutating (queryClient.ts:116–119).
action: Specify a shared exact mutation key for the four 2FA writes and use the V1 imperative cache check at dispatch. Use the existing reconciliation phase to cover the interval when checking is still active after the mutation settles. Keep the ACCT-9 same-tick test across two distinct 2FA handlers.
weight: blocking

[complexity] Skip the session refresh after an A3 refusal before the write
where: docs/design-pass/plan.md:641–643, :661, :744 (V2, V4c, ACCT-5).
why: V4c starts a session read for every completed A3 attempt, including origin_forbidden. The origin check rejects before the token handler (handler.go:115–121, :514–523), so the read cannot reconcile an identity change or alter the S5 action. V2 and ACCT-5 already specify that 403 needs no session read.
action: Narrow the A3 V4c trigger, keeping the shared read for outcomes whose actions depend on session state or whose write may have changed it. Exempt known pre-write S5 refusals and keep the ACCT-5 assertion with /me unavailable.
weight: non-blocking

DECISION: One requested lens completed and returned. Two blocking findings and one non-blocking finding. The plan and outside review files remained untouched.
Lens verdict: 2 blocking findings.
