NEXT: Read AI Factory entry rules, confirm repository scope, then read the complexity-lens contract.
FACT: AI Factory entry rules require append-only progress. The owner assigned a single read-only complexity lens.
NEXT: Verify the root and read local project rules and the lens contract.
FACT: Repository root verified and AGENTS.local.md read. Scope is one returned lens and one writable file.
NEXT: Read project briefing, review contract reference, and locate revision 1 and its source audit.
FACT: f6a4611a matches the checked-out plan and source audit. The briefing identifies ACCT as part of the owner-initiated design pass. The shared reference preserves the 44px control requirement.
NEXT: Read the full ACCT section, source findings and owner answers.
FACT: V1–V15 cover three resend surfaces, two token routes, 2FA state transitions, and shared validation and styling.
NEXT: Read the complete verification and implementation sequence plus the owner’s answers.
FACT: Owner authorized the audit’s 16 fixes and taste item. ACCT-1 lists the same five resend scenarios for three surfaces, including two surfaces specified as signed-in only.
NEXT: Trace existing mutation ownership, auth cache, and route/view seams.
FACT: All three resend callers already use session.resend and can keep local React Query mutation state. The 2FA route already has one mutation per action but collapses errors into a first-match list. FieldHelp has four callers, including Select and ComboBox.
NEXT: Check backend contracts, e2e fixture flow, and existing frontend utilities and tests.
FACT: Application contracts support anonymous resend, invalidation of earlier verification links, and client-local 2FA cancellation. Shared identity composition already exposes current/refresh and other composition modules use mutationOptions.
NEXT: Trace definitive token errors, existing notice spacing, and current test helpers without running the suite.
FACT: Native validation and current mutation objects already supply the inputs V1, V6 and V7 need. PasswordResetView already uses the A5 notice primary-link class. The registration e2e owns a run-specific address and captures the initial verification link.
NEXT: Confirm the notice/token reuse candidates, guest-mail fixture details, and phase snapshot before recording the findings.
FACT: main is f6a4611a. Source verification covered the requested routes, identity views, learning gate, shared controls and CSS, session layers, backend handlers and services, and account e2e. Test and runtime execution were excluded by the owner.
DECISION: Four actionable simplifications, each non-blocking. Keep the owner-approved behavior, shared resend result, and required verification.

[complexity] Test resend outcomes once and verify each caller’s wiring
where: docs/design-pass/plan.md:641 (ACCT-1) and :662 (shared result).
why: The table repeats a five-case matrix across three views. Guest success and guest field errors are reachable only on /verify-email (V1, :562–567). The banner is gated by session.data.emailVerified === false (__root.tsx:389), and the lesson gate requires an unverified signed-in user (EmailVerificationGate.tsx:27–28). Testing guest variants there would require state the product never supplies.
action: Put the applicable outcome cases on the shared result and the guest-specific field behavior on /verify-email. Extend the existing banner and gate tests with one representative result-wiring assertion each. Preserve the real anonymous resend and Mailpit proof in ACCT-2.
weight: non-blocking

[complexity] Keep localized validation within the two text controls
where: docs/design-pass/plan.md:589 (V6), :625 (ACCT-8), and :663 (FieldHelp scope).
why: FieldHelp also serves Select and ComboBox (controls.tsx:182, :236). An unconditional fallback there changes two additional control types and expands review and verification beyond the stated TextField/TextArea scope.
action: Put the localized fallback on the TextField/TextArea path, using FieldError’s existing validation details and preserving explicit messages. State the text-control scope in ACCT-8. Keep the full ordered gate.
weight: non-blocking

[complexity] Use each existing 2FA mutation as its form’s error source
where: docs/design-pass/plan.md:594 (V7).
why: settings.account.tsx:31–62 already has an independent mutation per action. Its current first-match aggregate at :65–67 discards that association. Sending a separately tracked source or creating an error dispatcher would repeat ownership that the mutation objects already express.
action: Replace the aggregate error in the plan with the corresponding mutation error per action. Bind regeneration and disable directly to their forms and use enroll and confirm errors in their existing steps.
weight: non-blocking

[complexity] Extend the existing CSS declarations for matching account treatments
where: docs/design-pass/plan.md:602–606 (V11–V13).
why: main.css already holds the AUTH intrinsic-width declaration (:514–516), the C3b container threshold and wrapping rule (:465–472), and the A5 primary-link margin (:524–528). Parallel copies of the same values would give later spacing and zoom fixes multiple places to update.
action: Specify extending the existing selectors where behavior matches. Reuse the existing primary-link margin for A3/A6 and narrowly cover the direct non-link actions that need 16px, preserving the ordinary AUTH form margin at :509–513.
weight: non-blocking

FACT: No plan, evidence, code, or other working file was edited. No make check, Docker, or e2e commands ran.
DECISION: 1 assigned / 1 returned.
NEXT: Return the four findings and the complexity-lens verdict for consolidation.
Lens verdict: no blocking findings.
