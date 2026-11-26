NEXT: Read the AI Factory entry rules, then apply the edge-cases lens to ACCT revision 2.
FACT: Verified the Courses repository root and read the shared entry rules.
DECISION: Keep every write in this working file and apply only the assigned lens.
NEXT: Read local rules, the repository briefing, and the assigned critique contract.
FACT: Local routes and the installed critique contract are available. The assigned lens covers boundary and duplicate data, repeated or concurrent actions, dependency failures, and partial application.
NEXT: Read the repository briefing and isolate the pinned plan, prior findings, and audit evidence.
FACT: HEAD is the requested revision 8d171f63. The briefing records the shipped accounts flows and the account design pass.
NEXT: Read the full ACCT section, revision-1 findings, and owner decisions before reading implementation.
FACT: The owner approved all 16 audit items and the A6 label. Revision 2 adds V1–V5 failure and repeat handling.
NEXT: Compare all ten prior findings with the revised acceptance, browser evidence, and rollout clauses.
FACT: Mapped the ten revision-1 items to V1, V2–V4a, V5a, V6, V11, V14, and the Mailpit journey.
NEXT: Check the actual route, session and 2FA UI state transitions against those clauses.
FACT: Both token routes auto-submit from an effect and refresh identity after success. Settings has four 2FA mutations and a render-derived pending flag.
NEXT: Check server write boundaries, session refresh and browser fixture details.
FACT: Server returns 204 even when no mail was sent. Confirmation and regeneration write state in multiple steps, and identity refresh shares one session query.
NEXT: Verify the shared Button, field and layout semantics, current E2E flow, and relevant query dependencies.
FACT: `current()` already uses staleTime 0 with the shared session key. Token routes receive named origin errors. V5a covers confirm/disable but V7 does not yet account for regeneration deleting codes before insertion.
NEXT: Confirm pending-guard mechanics, token reopening, and the 2FA secret layout boundary before weighing findings.
FACT: Browser app uses StrictMode; token-keyed remounts reset a local mutation. Shared query and mutation-cache behavior determines whether a read or submit is actually fresh.
NEXT: Read installed query source and exact CSS/guard inputs, then record revision-1 dispositions and revision-2 findings.
FACT: Installed Query Core reuses an in-flight fetch when fetchQuery supplies no cancelRefetch option. React Query publishes the useIsMutating snapshot through a scheduled subscriber. The route client is wrapped in StrictMode.
NEXT: Verify remaining mechanical details and write the final findings and dispositions.
FACT: Checked the request adapters, Go handlers and write ordering, the shared input and button primitives, CSS, and existing Playwright fixtures. No runtime checks were run.
DECISION: Revision-1 dispositions in original order:
- Resend 204: open. V1 acknowledges the request and separates Mailpit proof, but the user-facing sentence still predicts delivery and invalidation.
- Retained invalid cookie: closed for the reported one-cookie path. V1/ACCT-3 have a login recovery route and browser coverage. Duplicate-cookie 400 stays a generic failure by explicit plan choice.
- Coexisting resend controls: open. Shared observation is specified, but a same-tick dispatch guard is missing.
- Consumed links: partly closed. E3 is terminal, S4 offers a new link, and E4/E5 point to settings. V2 still leaves failed or stale session checks unresolved.
- Link swap: partly closed. A→B and a late A display are covered. Reopening A while its original request is pending can submit it again.
- Stale 2FA after Done: partly closed. Acknowledged success is seeded in cache and slow refresh is exercised. The failed reconciliation read after an ambiguous write has no transition.
- Copy state: closed in behavior. V14 keys state to the displayed set and handles missing/rejected clipboard and late completion.
- Standalone enable width: closed by V11 and ACCT-15.
- Mailpit identity: closed by the distinct ID and old-link-before-new-link journey.
- Validation scope: closed by V6's explicit TextField/TextArea scope and Select/ComboBox exclusion.

[edge-cases] Resend acknowledgement still promises delivery
where: §18 V1, line 576, ACCT-1/2
why: The copy tells an unverified user to expect a new email within minutes. accounts.go:123–140 and handler.go:373–377 can return 204 after a mailer failure, and 204 can also precede any token invalidation for a neutral no-op.
action: Keep the acknowledgement conditional on what the client knows. Describe what happens to the previous link if a replacement is issued. Check identical 204 copy for a successful send and for an accepted request with no delivered email.
weight: blocking

[edge-cases] Two resend activations can run before the shared pending render
where: §18 V1, line 583, ACCT-1
why: useIsMutating is a rendered snapshot. Its subscription schedules a render notification, and mutationKey alone allows multiple mutations. Both eligible handlers can enter before either receives a pending prop. Each resend can invalidate the other's link.
action: Require a synchronous check at dispatch against the shared pending state. In the two-surface test invoke both handlers before a pending render, assert one POST, and check that a later activation can proceed after settlement.
weight: blocking

[edge-cases] A3 recovery can read pre-write identity or fail during the read
where: §18 V2, lines 591–595, ACCT-5
why: fetchQuery with staleTime 0 still joins an in-flight read on the same key (identity-queries.ts:19–37 and Query Core query.ts:397–415). A /me request that started before confirmation can return emailVerified false after the writes, and a failed /me rejects. The specified flow can miss S3, show an unusable resend button for the now-verified account, or leave retry without a defined result.
action: Specify a bounded reconciliation state and require a read started after the confirmation outcome. Define the visible fallback and next action for a failed read. Check an overlapping pre-write /me and a rejected reconciliation /me.
weight: blocking

[edge-cases] Reopening a pending link can post its one-time token twice
where: §18 V4a, line 610, ACCT-7
why: The current effect on each route submits from its local idle state. For A→B→A, the new A mount is idle while the original A request can still consume the token. The new A can then display token_used while the original A succeeds, especially visibly as E4 on the email-change route. StrictMode effect replay is another duplicate invocation boundary.
action: Preserve a single in-flight attempt per route and token across remounts. Define which attempt the reopened screen observes and when explicit retry can start. Verify A→B→A before A settles and effect replay without two POSTs for A.
weight: blocking

[edge-cases] An ambiguous 2FA write can outlast the reconciliation read
where: §18 V5a and V8, lines 623 and 632, ACCT-8/9/12
why: Confirm can enable TOTP before failing to return codes, and disable can clear TOTP before recovery-code deletion fails (two_factor.go:101–110, 211–224). A failed or already-in-flight /me does not settle either write. If pending ends on this failure, stale controls can offer Cancel or a second write even though the initial change may have applied.
action: Define a recoverable unknown state for confirm and disable until a post-mutation /me settles. Give it a session recheck action, and define the confirmed-on, confirmed-off and signed-out exits with focus behavior. Test confirm/disable error followed by a failed /me and later recovery.
weight: blocking

[edge-cases] A failed regeneration may have invalidated the saved codes
where: §18 V7 and V14, lines 627–630 and 644, ACCT-11/18
why: RegenerateRecoveryCodes consumes the supplied proof and deletes previous codes before generating and inserting replacements (two_factor.go:227–241). A 5xx or lost response can leave the user without a usable saved code set. V7's group-local generic failure never says the old codes may be gone or that the just-used authenticator step may be consumed.
action: Specify separate feedback for rejected proof and uncertain replacement. Tell the user to obtain a new displayed set with a fresh authenticator code after uncertain replacement, and cover a failure after deletion.
weight: blocking

[edge-cases] Saving can clear edits made after submission
where: §18 V10, line 636, ACCT-14
why: AccountSettingsView.tsx:76–126 leaves the controlled inputs editable while only the action is disabled. Clearing on 204 can discard a later input the user entered during a delayed request. For email the success message then applies to the submitted address while the newer draft disappears.
action: Define whether fields stay editable during a write. If they stay editable, clear the submitted values only and preserve a newer draft. Verify a delayed acknowledgement with an edit in flight.
weight: blocking

[edge-cases] Two visible lines cannot be assumed to fit the full secret
where: §18 V15, line 646, ACCT-19
why: The secret is 32 unpadded Base32 characters (identity/domain/two_factor.go:26–31). At 390 with 200% text the shell, card and input padding leave roughly 280 CSS px in the current layout. A 32px monospace font at roughly 0.6em per character fits about 14 characters per line, so a two-line viewport can still hide the end. The shared TextArea also applies its own minimum size and supplies no rows prop to the inner textarea (controls.tsx:75–89 and tokens.css:251–254).
action: Treat two lines as a minimum or let the content determine height. In the boundary check verify the last character is visible without scrolling inside the field and the accessible value is still the raw secret.
weight: blocking

DECISION: lens verdict (8 blocking findings).
FACT: Verified eight fixed-format findings and no tracked working-tree changes. Only the assigned working file was appended.
NEXT: Return the scoped lens verdict.
