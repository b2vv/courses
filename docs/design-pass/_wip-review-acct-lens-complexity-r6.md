NEXT: Read the AI Factory entry rules, then review revision 6 through the complexity lens only.
FACT: AI Factory entry rules read.
NEXT: Verify repository root and target revision.
FACT: Verified Courses root and HEAD/main at cf5c5477 with no tracked modifications.
NEXT: Read the local rules, repository briefing, and complexity lens contract.
FACT: Local rules and the single-lens contract are in scope. DP2 points to the ACCT phase materials.
NEXT: Read the briefing, the exact section, the revision delta, and the prior complexity records.
FACT: Read the relevant briefing, anti-pattern reference, and prior closure history. The combined source output was truncated.
NEXT: Read the revision-6 delta and extract every prior finding title directly from its own file.
FACT: The five prior reviews contain 4 + 3 + 3 + 1 + 1 finding entries, including the repeated CSS concern. Rev 6 changes A6 session refresh, 401 handling, 403 actions, and 2FA conflict recovery.
NEXT: Read the current V1–V5a decisions and owner evidence.
FACT: Owner approved findings 1–16. V5a now applies one conflict rule to all four 2FA mutations. V3 uses invalidateQueries with cancelRefetch for every A6 result.
NEXT: Trace A6 and the session helper, settings guard, and shared query implementation.
FACT: Current A6 already calls identityDependencies.refresh on success. Query defaults cancelRefetch to true, and Query.fetch only takes that cancellation branch when cached data is defined.
NEXT: Verify the cold-cache root path, guard behavior, and existing refresh reuse across A3, A6, and settings.
FACT: Root beforeLoad can return unavailable without seeding data, and the mounted root then observes current(). requireIdentity uses the same fetchQuery and joins a pending read. A6 therefore cannot rely on invalidateQueries to replace a cold read.
NEXT: Verify adapter and server behavior, A6 proof coverage, and the remaining prior closure details.
FACT: ACCT-6 holds an old read with a cached address path and does not require data to be undefined. V6–V14 and ACCT-1 retain prior reuse fixes. The revision delta has no frontend/backend changes.
NEXT: Confirm typed error behavior and the scope of a reusable uncertainty decision, then record prior dispositions.
FACT: Verified the four server 2FA conflict paths, the token action and Origin middleware ordering, 204 mailer semantics, and existing typed session error codes.
NEXT: Verify the installed Query cold remount and cancellation behavior, and check for an existing typed uncertainty predicate.
FACT: No typed uncertainty predicate appeared in the identity sources inspected. The adapter and domain already expose a single typed code vocabulary for all five call sites.
NEXT: Confirm the cold query observer branch and read the exact revision-1 dispositions before writing the findings.
FACT: Installed Query can start the first observer load with data undefined even with refetchOnMount false. All twelve prior entries are closed by their current decision and proof text.
NEXT: Check the transport code union and the existing banner/result components before recording the verdict.
FACT: The banner and lesson gate both accept local resend state through their existing callers. Installed Query explicitly defaults cancelRefetch to true, so adding that option is redundant as well as insufficient on the cold path.
NEXT: Verify the transport taxonomy and append dispositions and findings.
FACT: The shared transport collapses network faults and unnamed 5xx to dependency_unavailable, keeps request_timeout and invalid_response distinct, and the adapter preserves named refusals. V10 is the new consumer of the same uncertain-write classification.
DECISION: Rev 1 closed 4/4. ACCT-1 tests the resend outcomes once and checks the two authenticated callers' wiring (plan.md:730). V6 confines the fallback to TextField/TextArea (:678). V7 uses each group's mutation (:680). V11–V13 extend the existing width, zoom, and spacing treatments (:691–695).
DECISION: Rev 2 closed 3/3. V2 cancels the old exact-key read before fetching current(), with the held-read proof in ACCT-5 (:636, :733). V14 ties copy feedback to the mounted codes branch (:697, :741). V13 explicitly extends the A5 spacing rule (:695).
DECISION: Rev 3 closed 3/3. V4a retains only in-flight work, with joined and completed re-entry cases (:651–652, :735). V5a holds the current form and focus during the new read (:667–671, :737). V1 derives cross-caller pending state from the shared mutation key (:623, :730).
DECISION: Rev 4 closed 1/1. V2 uses one terminal-result S3 predicate and keeps S5 off the session-read path. Retry for retryable S5 sends another POST, and 403 is navigation-only (:632–636, :733).
DECISION: Rev 5 closed 1/1. V5a applies both conflict codes to all four actions using one reconciliation rule and ACCT-9 checks the enroll path, including the mounted initiating button (:673, :737).
NEXT: Record actionable revision-6 findings, then return the complexity verdict.

[complexity] Reuse the fresh session read for A6
where: docs/design-pass/plan.md:648 (V3), :717 (ACCT-13), and :734 (ACCT-6).
why: A6 already uses the exact-key refresh (email.change.confirm.tsx:23, identity-queries.ts:35–37). Query already defaults cancelRefetch to true and only cancels a pending request under that option if cached data is defined (queryClient.ts:320–334, query.ts:397–415). If root beforeLoad receives unavailable with no cached data, its mounted observer can start a new initial /me read before the token attempt finishes (root-identity.ts:26–32, __root.tsx:176, queryObserver.ts:724–746). A6 refresh can then join that older read. The settings guard calls fetchQuery for the same key and can join it too, showing the pre-change address despite ACCT-13. The held-read proof does not specify a cold cache.
action: In V3 reuse V2/V5a’s explicit exact-key cancellation and fetchQuery(current()) through identity composition. Start the fresh request before the settings action can enter its guard, and keep the E2–E5 token result independent of a read error. Extend the held-read proof to start with data === undefined and navigate while the replacement is pending.
weight: blocking

[complexity] Share the uncertain-write classification
where: docs/design-pass/plan.md:634 (V2), :646 (V3), :667 and :676 (V5a), :689 (V10).
why: V10 adds another consumer of the same uncertainty list. The transport and adapter already expose the same typed codes for all these writes (transport.ts:45–60, session-repository.ts:138–141, identity.ts:28–32). Repeating the list at each consumer adds decisions that must stay synchronized, while the different guidance and recovery steps are specific to each action.
action: Name one predicate over the existing identity error codes for dependency_unavailable, request_timeout, and invalid_response. Reuse it across A3, A6, and settings, leaving each action to choose its own recovery.
weight: non-blocking

FACT: Static review at cf5c5477. Only this working file was edited. No other lens files were read. No make check, Docker, or e2e commands ran.
DECISION: 1 assigned / 1 returned.
Lens verdict: 1 blocking finding.
