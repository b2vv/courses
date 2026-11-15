| R8 finding | Disposition |
|---|---|
| Major 1 — cached draft remains loadable after mode change | **Resolved.** MS-D10 requires local snapshot comparison and dual query invalidation; M3-R15 covers toggle → disabled Load → Regenerate → enabled Load, including `und`. |
| Major 2 — acceptance misses Python→faster-whisper boundary | **Resolved.** M2-R12 executes the real `transcribe.py`, captures model kwargs, and checks both multilingual/no-hint and ordinary/hint behavior. |
| Minor 3 — PATCH master-rewrite rule | **Resolved.** Flag-only changes skip rewriting; language/label changes retain it. Required metadata and omitted-flag semantics are explicit. |

Ranked findings:

- **Blocker:** none.
- **Major:** none.
- **Minor, nonblocking — incomplete DTO enumeration.** [plan-ms2-3.md:246](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:246) omits `trackMultilingual` from `GET drafts`, although [MS-D10:466](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:466) explicitly requires it. Add it to the endpoint field list.

**Four independent lenses launched / four returned.** Pending, requested-only claim, snapshots, stale rules, MS-D7’s additional mode guard, revised S6 acceptance and the language-only `und` exception are consistent. No unnecessary complexity requiring changes was found. O1–O8 remain fixed.

Verified against the requested code paths. Installed faster-whisper **1.2.1** inspection confirms native `multilingual` support and per-window language detection. Acceptance remains planned evidence; no tests were run and no files changed.

Reviewed HEAD: `a1eaffb96878be4aa922701b7aad3cbed79eac4f`  
Plan SHA-256: `e2a6a646888a7b992fab97c84342174ae4a03806233c203c8f44dc64660f0cc1`

**Open blocking findings: 0.**

**VERDICT: ACCEPTED**