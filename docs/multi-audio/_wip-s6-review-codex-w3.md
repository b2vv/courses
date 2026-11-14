Reviewed only `fe203457bded8d82ab990bd13ae64fad5b630304` on `multi-audio-s6`. No files changed. Verification used in-memory execution of the parent and current factories with installed TanStack Query; Rstest and full gates were **not run**.

1. **Merged condition: equivalent.** Opened [admin-draft-queries.ts:67](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:67), [:114](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:114), and the parent’s two checks.

   Counter-example: exercised both actual factories across **108 combinations**: absent/live/expired marker × all six server states × null/non-null marker timestamp × null/equal/different server timestamp. Expiry included exactly 900,000 ms. Compared complete returned drafts, including a repeated read: **zero differences**.

   Both implementations behave as follows:

   | Marker | Server response | Result |
   |---|---|---|
   | Absent/expired | Any | Server draft |
   | Live | State null | Synthetic pending |
   | Live | Any state, same non-null timestamp | Synthetic pending |
   | Live | Non-null state, different timestamp or either timestamp null | Server draft |

   Terminal responses clear the marker only when they reach the terminal branch. Equal **null** timestamps do not establish an old draft. The removed second check was redundant.

2. **Lifetime and scope: request isolation holds; cache-reset semantics change.** Opened [admin.ts:31](/Users/strelia/projects/courses/frontend/src/app/composition/admin.ts:31), [:44](/Users/strelia/projects/courses/frontend/src/app/composition/admin.ts:44), [router.tsx:17](/Users/strelia/projects/courses/frontend/src/router.tsx:17), [:35](/Users/strelia/projects/courses/frontend/src/router.tsx:35), and [:42](/Users/strelia/projects/courses/frontend/src/router.tsx:42).

   Each router creates its own QueryClient and draft factory. The Map is closure-local, with no global mutable store. Separate factories returned ordinary null despite another factory’s pending marker.

   Counter-example: regenerate, call `client.clear()`, then reread; also read through a second client sharing the factory. **Old code returned ready; new code returned pending in both cases.** Thus the Map belongs to the factory, not the supplied client.

   This does not block the current wiring: production pairs one factory with one client. [use-sign-out.ts:21](/Users/strelia/projects/courses/frontend/src/app/composition/use-sign-out.ts:21) changes the session query and navigates; it does **not** clear the client or recreate dependencies. A browser identity switch can therefore retain this lesson-level wait until expiry. It retains a timestamp/token, not another user’s draft contents; repository authorization still governs every read. A future client-reset requirement must explicitly reset or recreate the factory.

3. **Growth: session-scoped retention is acceptable here.** Opened [admin-draft-queries.ts:73](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:73), [:100](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:100), and [:125](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.ts:125).

   Counter-example: regenerate distinct lessons, leave each before completion, and never revisit them. Expiry alone does **not** remove those entries; retention is O(distinct abandoned lesson IDs) for that router’s lifetime. Repeated regeneration of one lesson overwrites its entry. Mounted lessons read the marker during polling and remove expired entries.

   There is no hard memory bound, but each retained value is one lesson key plus two scalar fields, and the entire Map dies with the router. No session-scale blocker is substantiated.

4. **The tests discriminate the GC failure.** Opened [admin-draft-queries.test.ts:157](/Users/strelia/projects/courses/frontend/src/app/composition/admin-draft-queries.test.ts:157), [rstest.config.ts:49](/Users/strelia/projects/courses/frontend/rstest.config.ts:49), and installed Query’s `removable.ts:24` / `query.ts:225`.

   Counter-example: replayed the new assertions against both implementations using controlled timers:

   | Implementation | GC | State at 10 minutes | Reads next 5 seconds |
   |---|---|---|---|
   | Parent | Default | null / ready | 0 |
   | Parent | Infinity | pending / pending | 1 |
   | Current | Default | pending / pending | 1 |

   Both new cases therefore fail on the parent under browser defaults. Their happy-dom project supplies the browser environment needed for five-minute GC.

   `rg` across **all `frontend/src` and `frontend/e2e`** found no remaining `draft-pending` references or readers.

Findings ranked:

- **Blocker:** none.
- **Major:** none.
- **Minor:** none substantiated; factory-reset semantics and lazy retention are documented above.

**VERDICT: SHIP**