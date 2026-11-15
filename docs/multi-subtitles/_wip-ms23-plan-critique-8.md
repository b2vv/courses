**VERDICT: BLOCKED**

**Blocker:** none.

**Major**

1. **Changing `multilingual` can leave a cached draft loadable.**  
   [Plan:454](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:454), [guard:344](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:344).

   The guard relies on server `draft.stale`, but the existing rename mutation invalidates only audio tracks ([admin-audio-queries.ts:137](/Users/strelia/projects/courses/frontend/src/app/composition/admin-audio-queries.ts:137)). Completed drafts stop polling, and the draft DTO exposes no `trackMultilingual` for comparison locally. Toggling the flag can therefore leave cached `stale=false` with Load enabled.

   **Required plan change:** explicitly refresh/invalidate affected drafts or compare their claimed mode locally. Add acceptance: completed draft → toggle mode → Load disabled in the open editor → no automatic replacement → Regenerate → new mode applied and Load enabled. Include an `und`-origin draft: its language exception must not bypass mode staleness.

2. **O8 acceptance stops before the Python→faster-whisper boundary.**  
   [M2-R12:477](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:477), [M3-R14:494](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:494).

   CLI-argument checks and successful English recognition can pass while Python accepts but ignores `--multilingual`. Existing engine tests substitute shell scripts; the actual engine call is separate ([transcribe.py:49](/Users/strelia/projects/courses/backend/internal/transcript/adapters/whisper/transcribe.py:49)). The proposed worker-log assertion also lacks a producer: current logging captures stderr, not argv.

   **Required plan change:** test the real Python entry point with a captured model call, asserting `multilingual=True` and `language=None`, plus the false-mode behavior. Retain Go argument coverage. Specify the diagnostic producer if retaining log evidence.

**Minor**

3. **Qualify the “no master rewrite” PATCH rule.**  
   [Plan:445](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms2-3.md:445).

   This endpoint also changes language/label, which require master regeneration ([audiotrack.go:287](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:287)). State that flag-only changes skip regeneration; language/label and combined changes retain it. Define how a flag-only request preserves existing metadata—the current decoder requires language and label.

Four independent lenses launched / four returned. Complexity was clean; findings above consolidate logic, edge-case and acceptance results.

O7 pending/requested-only claim rules, claim snapshots, token-scoped completion/recovery, revised S6 scenarios and the `und` language exception are otherwise consistent. O1–O8 remain fixed. Upstream [faster-whisper 1.2.1 source](https://github.com/SYSTRAN/faster-whisper/blob/v1.2.1/faster_whisper/transcribe.py) supports the native multilingual option; installed-package inspection was blocked by Docker socket permissions.

Reviewed HEAD `a1eaffb96878be4aa922701b7aad3cbed79eac4f`; plan SHA-256 `09ce240abafbd6853594089dbd90a65a2aa9a0c2a2b6e5ec32e7bf54da38f268`. Read-only inspection; no files changed or tests run.

**Open blocking findings: 2 Majors.**