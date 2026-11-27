NEXT: Read repository entry rules, then apply the plan-critique complexity lens to revision 8.

FACT: Recorded the initial NEXT and read the shared agent entry rules.

NEXT: Verify the repository root and read its local rules and the requested skill contract.
FACT: Inspected the repository root, local routing rules and plan-critique contract.
DECISION: Run only complexity and keep all review notes in the assigned working file.
NEXT: Locate the exact revision and read the repository briefing, task pointer, plan slice and audit.
FACT: HEAD matches 34212f73 and the plan and audit have no worktree diff.
NEXT: Read §18 in full and the relevant audit evidence before checking its proposed mechanisms against code.
FACT: Read the revision 8 implementation and acceptance text and traced the audit findings and owner decisions.
NEXT: Check the current token, navigation, session, form, and test mechanisms for needless duplication or added work.
FACT: Mapped the operative V1–V15, ACCT rows and session/token call sites against the revision 8 scope.
NEXT: Inspect implementation details and existing test helpers, focusing on mechanisms the plan can reuse.
FACT: Verified route and view boundaries, current session refresh hooks, shared controls and CSS, identity endpoint behavior, and account e2e anchors.
NEXT: Check terminal error paths, local pending state and the Mailpit helper to distinguish real simplifications from required behavior.
FACT: Verified refresh is exact invalidation, A6 navigation labels are session independent, and deterministic rejection paths can avoid extra reads. Verified the secret is a display-only value with no form reads.
NEXT: Validate proposed simplifications against the fixed acceptance matrix and record weighted findings.
FACT: The session adapter already exposes all needed result codes and shapes. Production settings entry reads /me and the root and learning surfaces use the existing session/resend seams.
DECISION: Keep the shared resend mechanism and keyed token instance. Report only simplifications with the same audit outcome, treating Л1–Л6 as accepted.
NEXT: Append the weighted findings and final verdict.

DECISION: Record three non-blocking simplifications. The audit fixes remain achievable under the accepted limits.

[complexity] Reuse the settings guard for A6 recovery
where:   §18 V3 (658), ACCT-6 (743, 767)
why:     E3/E4 and the settings action in E5 use the same destination for every viewer. The guard already reads /me on entry and sends guests to login with returnTo. Refreshing on every A6 result adds route-level cache work even for a rejected token or email_taken, neither of which updates the address. See require-identity.ts:11–27, identity-queries.ts:35–37 and accounts.go:245–265.
action:  Keep the existing success invalidation, publish the settings link directly from the result, and let the settings guard load the session for recovery. Preserve the spent-link-to-current-address test.
weight:  non-blocking

[complexity] Keep 2FA action completion tied to the write
where:   §18 V5/V5a (666–682)
why:     settings.account.tsx:36–64 currently awaits refresh in confirm, disable and regenerate. If carried forward, the shared pending guard can continue to block actions after the successful write has already changed the card. V5a supplies confirm/disable state directly and regeneration only replaces recovery codes, which /me does not expose. See two_factor.go:227–241 and identity.ts:2–9, 42.
action:  Specify that confirm/disable use the accepted cache update and start the existing invalidation in the background. Let regeneration end when its own result is available. Keep the focus targets and exact mutation-key guard.
weight:  non-blocking

[complexity] Use the existing 2FA forms as the two named regions
where:   §18 V7 (692–696), ACCT-11 verification (772)
why:     AccountSettingsView.tsx:185–221 already has a separate form per action. The proposed group role adds a second semantic boundary per operation when the form itself can be named by its heading and provide the same scope for the field, error and test.
action:  Put each heading and error inside its existing form, name the form with aria-labelledby, and scope the verification to the named form. Keep the distinct headings and the destructive form last.
weight:  non-blocking

FACT: The review used static repository reads. No make check, Docker or e2e was run. The assigned working file is the only file edited.
DECISION: lens verdict (no blocking findings).
FACT: Appended the findings and checked that the plan and audit remain clean.
