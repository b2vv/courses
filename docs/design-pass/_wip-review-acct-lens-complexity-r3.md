NEXT: Read the AI Factory entry rules and confirm the repository root.
FACT: AI Factory entry rules read. Repository root verified as /Users/strelia/projects/courses.
NEXT: Read local rules, project briefing, and the canonical critique contract.
FACT: Local routing and critique contract read. The user scoped this pass to one lens and one append-only file.
NEXT: Read revision 3, both prior complexity reviews, and the owner answers.
FACT: main points to efc0bcf9. The first combined read exceeded the display limit.
NEXT: Read the prior findings, revision 3 decisions, and owner answers in bounded chunks.
FACT: Recovered four revision-1 findings and three revision-2 findings. The plan explicitly extends A5 spacing, scopes clipboard state to the mounted codes panel, and cancels stale session requests before a fresh read.
NEXT: Trace frontend route and session ownership and inspect the remaining 2FA plan decisions.
FACT: Frontend routes use separate mutations for resend and each 2FA action. Identity composition already owns the current-session query and refresh.
NEXT: Check backend boundaries, view lifetimes, the Mailpit fixture and the installed Query implementation.
FACT: Owner approved the full audit. Existing recovery codes and enabled management are exclusive branches. The session query already has an exact-key refresh helper.
NEXT: Verify whether TanStack pending state can replace the resend flag, and check token entry retention.
FACT: Query dispatches pending synchronously. A shared mutation key can expose the same pending state to the click guard and all three callers. Default retention needs checking against V4a’s application-lifetime promise.
NEXT: Confirm exact source lines for retention, cancellation and session ownership, and check the pinned source snapshot.
FACT: Unobserved settled mutations expire after five browser minutes by default. The existing token acceptance checks cover only immediate remounts. V5a hides controls in its checking phase despite the same-form error path retaining focus.
NEXT: Confirm the exact cache APIs and focus-related control behavior before writing findings.
FACT: QueryClient.isMutating reads pending mutations synchronously; useIsMutating observes the same filter. The code diff for scoped frontend and backend files between the plan source baseline and revision 3 is empty.
NEXT: Record closure of the earlier items and classify the three new findings.
DECISION: Revision 1 is closed across all four items. ACCT-1 tests shared outcomes once and checks banner/gate wiring separately (:699). V6 limits localized fallback to text controls (:647). V7 assigns errors to their existing mutations (:649). V11–V13 extend existing width, zoom and narrowly scoped spacing rules (:660–664), including the deferred A5 margin item.
DECISION: Revision 2 is closed across all three items. V2 cancels the pending exact-key session read before fetching current(), covered by the held older /me case (:609, :702). V14 scopes copy feedback to the mounted codes branch and covers completion after dismissal (:666, :710). V11–V13 close the remaining CSS reuse item (:660–664).

[complexity] Make token deduplication last as long as V4a promises
where: docs/design-pass/plan.md:558, :624, :680, :704.
why: V4a depends on the presence of a prior mutation entry to suppress a POST. The installed Query core removes unobserved settled mutations after the default five browser minutes (removable.ts:24–29, mutation.ts:143–161). After leaving the route and waiting past that window, reopening the same token in the same SPA can auto-POST and display token_used in place of its previous success. The immediate A→B→A proof passes even with that behavior. Adding a parallel token registry would duplicate the cache.
action: Specify the existing mutation cache as the sole owner and set a retention policy for these two keyed mutations that matches the stated SPA lifetime. Add a same-SPA revisit after the default GC window to ACCT-7. If only in-flight deduplication is required, revise V4a and the revision summary to state that narrower contract explicitly.
weight: blocking

[complexity] Preserve the 2FA form while checking an uncertain write
where: docs/design-pass/plan.md:635, :639–643, :706.
why: V5a hides the controls for the fresh /me read, then promises that focus will not move if confirm returns to setup with an error. The current setup branch contains the focused submit button or code field (AccountSettingsView.tsx:147–175). Removing or hiding that form drops focus before /me resolves, even when the same form must return. Restoring it would require extra focus bookkeeping for a phase that can keep the form in place.
action: Keep the current form visible and mounted while reconciling, announce the check, and block edits and new writes with focus-preserving pending or read-only controls plus submission guards. Keep Cancel unavailable. Move focus only when the result replaces the form or enters the explicit read-failure state. Extend ACCT-9 with a held /me followed by the same-form result and assert focus through both states.
weight: blocking

[complexity] Derive the shared resend guard from the mutation cache
where: docs/design-pass/plan.md:586, :596, :699.
why: The router-owned flag duplicates pending state already held by the three mutations. MutationObserver.mutate builds and executes a cache entry synchronously (mutationObserver.ts:136–150). execute dispatches pending before its first await (mutation.ts:208–225), and QueryClient.isMutating reads that state directly (queryClient.ts:116–119). An imperative click guard can see the first invocation before React renders.
action: Give the three hook instances one exact resend mutation key. Check isMutating at dispatch time and use useIsMutating for pending display. Keep each instance's local result and the existing same-tick ACCT-1 proof.
weight: non-blocking

FACT: Source-only review of main at efc0bcf9. No make check, Docker or e2e commands ran. Only the assigned working file was edited. One lens assigned and one returned.
NEXT: Return revision closure, findings and the lens verdict for consolidation.
Lens verdict: 2 blocking findings.
