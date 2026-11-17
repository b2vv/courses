All 25 rows are accepted. The supplied patches reconstruct the current ten test files exactly; their hashes match the journals. The scoped package changes only tests, and the six removed lines in `draft_handler_test.go` preserve or strengthen the original assertions.

| Rows | Accepted evidence |
|---|---|
| B055 | Existing cancellation-signal test also asserts `multilingual=true` |
| B145, B148, B184, B185 | Existing guarded-editor test fails when language tabs disappear |
| B156 | Existing guarded-Save test fails when fresh tracks become empty |
| B007 | New query/application test detects lost multilingual mode |
| B146, B147 | New real-route tests detect incorrect upload/switch metadata |
| B152 | New sibling-tab test detects incorrect draft selection |
| B153 | New regeneration test detects incorrect track identity |
| B155 | New failure/retry test detects swallowed track-read errors |
| B192 | New inventory test detects incorrect lesson identity |
| X-019 | Named async-retirement assertion fails after verifying editing and both Saves |
| A14 | Selected-attempt marker assertion |
| R06 | No-speech completion track assertion |
| H11, H14, H16, H19 | Serialized state, asset, non-null segments and segment-ID assertions |
| Q05 | Real task handler → recognition application → recorded Claim |
| S68 | Real PostgreSQL duration assertion |
| H38–H40 | Real route registration and identity middleware; valid/missing/invalid cookies, no pre-injected Actor |

Every detector is CI-selected through `make test` or `make integration`. I checked the raw baseline/mutant records: all failures are attributable named-test failures, with no collection or infrastructure errors counted. All six FE “existing kill” claims hold. X-019’s canonical evidence correctly cites **F02**, the explicit assertion, rather than global F01.

One **nonblocking cleanup issue**: [TranscriptEditor.test.tsx:1839](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.test.tsx:1839) installs the process listener before `await editor(...)`, outside `try/finally`. A mounting failure would bypass listener removal and spy restoration. Move setup under the cleanup guard. Successful baseline and mutant paths clean up correctly; no added sleeps were found.

Reviewer execution: `npm run typecheck` **passed**. Go unit tests and vet were **sandbox-blocked** by temporary/cache directory permissions. Mutation and integration conclusions rely on inspected, snapshot-matched execution records. No Docker or edits performed. This package verdict does not waive PR #109’s separate final ordered gate.

**VERDICT (gaps-e2e): SHIP**  
Counts: **6 KILLED-by-existing + 19 CLOSED; 0 blockers, 1 nonblocking Minor.**  
**rows accepted 25/25**