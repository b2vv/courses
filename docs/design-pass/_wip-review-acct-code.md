NEXT: Read the shared entry rules, then record this opening step in the review file and verify the ACCT contract at the review tip.
FACT: Read the installed agent-entry rules.
NEXT: Verify the repository and load its scoped instructions and briefing.
FACT: Verified repository root, local wiring, briefing, DP2 pointer and exact review tip. Existing unrelated WIP files are outside the review.
DECISION: Review the pinned commit directly in the requested single working file.
NEXT: Read accepted §18, audit, and evidence before inspecting the diff.
FACT: Found revision 12 ACCEPTED, the acceptance and checks tables, recovery journey, rollback, and the 25-file diff.
NEXT: Compare current decision tables and source audit with the session and route code.
FACT: Located token-keyed routes, synchronous settings mutation guards, and accepted session race boundaries.
NEXT: Isolate current V1–V6, the one-shot read, and the evidence deviations.
FACT: V1 and link-failure code membership align. V2 requires a session read after any result except 403, while the trigger only covers verified/link-failed.
NEXT: Check retry timing and test assertions, then inspect 2FA, shared controls/styles and e2e.
FACT: Settings guards synchronously serialize writes by key. Focus intent is recorded in the view before the dispatch guard. Resend ownership follows hook instances, and pending survives unmount in the mutation cache.
NEXT: Trace validation merge behavior, CSS consumers, and browser fixtures.
FACT: Browser contexts close in finally; the mail journey checks a distinct ID and a real 410. No added form login. The held E2 read test does not assert that a GET actually started.
NEXT: Inspect manually validated field consumers and confirm exact fixture/backend behavior.
FACT: Existing manual isInvalid callers supply their explicit errorMessage when invalid. Shared CSS affects existing AUTH notices and the teach error notice through inline-size containment; cluster/form-success class consumers are confined to the slice.
NEXT: Check identity transitions and run the four permitted focused test files.
FACT: Focused tests are running. Backend reads confirm both stale-cookie calls fail before the resend limiter, and the guest call uses the email/IP bucket.
NEXT: Confirm whether sign-out/sign-in remounts the banner hook and classify findings.
FACT: Explicit validation text is preserved by current callers and React Aria validation prioritizes controlled, server, then custom errors. The controls test omits the required length branch.
NEXT: Confirm the root banner survives client-side identity changes and collect focused test results.
FACT: Focused Rstest could not start because the sandbox denied listen on ::1:3000. Diff whitespace check passed.
NEXT: Run the permitted typecheck and substantiate stale-banner and retry dispatch findings.
FACT: Root sign-out and sign-in are SPA transitions. The banner callback/result is owned by the persistent Layout, so an earlier viewer’s resend outcome survives the banner unmount. The token retry functions also use a render snapshot rather than synchronous dispatch locking.
NEXT: Collect typecheck result and record verdict with source lines and test gaps.
FACT: TypeScript passed at 78a572b1 with no emit. Rstest ran no tests because localhost listen was denied. No Docker, e2e, or make gate was run. Tracked checkout remains clean.
FACT: The manual validation callers, the separate 2FA forms, one-time recovery block, current-email display, and stateUnknown/reset rules were checked. The shared style additions use existing notice consumers and slice-owned cluster/form-success consumers. Accepted Л1–Л6 were excluded from findings.
FACT: Mailpit journey verifies a different message ID and a real 410; all added contexts close in finally, the account already has teardown, and no form login was added. Only the guest resend reaches the 5/15-minute email/IP bucket. Author-reported captures and Kiri e2e are plausible from the fixture shape but were not independently rerun.
DECISION: BLOCK.
FACT: P2, frontend/src/routes/verify-email.tsx:62. A3 skips its V2 session read on S5 results other than 403. If the server verifies the address and its response is lost, the cached unverified viewer and banner persist until another session read, despite the explicit post-result refresh. The S5 route test checks the copy and retry outcome but does not assert the intervening read. Confidence high.
FACT: P2, frontend/src/routes/__root.tsx:237. Layout owns the banner resend observer and survives in-app sign-out and sign-in. After an accepted resend or another outcome for account A, sign out and sign in to unverified account B. The new banner displays A's result without any press by B. A request finishing after the banner unmounts can produce the same leak. Confidence high.
FACT: P2, frontend/src/routes/verify-email.tsx:108 and frontend/src/routes/email.change.confirm.tsx:67. Two retry activations before the observer result reaches React both see isPending=false and dispatch. The later observer can report token_used while the earlier write succeeds, including E4 in place of E2. State is correctly scoped by token and StrictMode is separately guarded. Confidence high for same-render dispatch.
FACT: P3, frontend/src/routes/route-guards.test.tsx:3734. The E2 held-read test also passes if the refresh is deleted. It proves E2 renders before release but never observes a post-204 GET. The specified E0 offline route and length flag translation also lack direct regression assertions. Confidence high.
FACT: P3, frontend/e2e/accounts.spec.ts:618. The wrong-code focus probe uses a syntactically valid fixed TOTP for a random live secret, so it can occasionally enable 2FA and wait for an alert that never appears. Confidence high.
NEXT: Return the pinned-range verdict and verification limits to the owner.
