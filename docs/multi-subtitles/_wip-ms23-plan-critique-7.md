**R6 Major: resolved.** [plan-ms2-3.md:284](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:284) separates the baseline `(language, token)` from the save language, preserves the pair through capture, restore and recapture, and sends older snapshots without `baseLanguage` through the cross-language branch.

This matches [write.go:51](/Users/strelia/projects/courses/backend/internal/transcript/adapters/postgres/write.go:51): no token means INSERT; a token means UPDATE of the specified language/version. Unknown token ownership therefore cannot silently authorize an overwrite.

Four independent lenses launched / four returned:

| Lens | Result |
|---|---|
| Logic gaps | Clean: token selection follows baseline ownership. |
| Unnecessary complexity | Clean: existing PUT and recovery mechanisms are reused. |
| Edge cases | Clean: legacy snapshots, destination races and stale recovery have explicit outcomes. |
| Acceptance sufficiency | Clean: [M3-R13:444](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:444) covers pair persistence and outbound-token branches; M3-R11(d) adds browser coverage for rename→reload→restore→Save. |

**Ranked findings**

- **Blocker:** none.
- **Major:** none.
- **Minor:** none.

Verified against the requested backend, transcript UI/application and e2e sources. O1–O6 remain fixed. This accepts the revised plan; implementation and test execution remain outstanding.

Reviewed HEAD `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`; untracked plan SHA-256 `a441ca6da34e142a2019daed0a69f412f2ba5fe0f042f735f0e5d2d93e664580`. Read-only inspection; no files changed or tests run.

**Open blocking findings: 0.**

**VERDICT: ACCEPTED**