`stage: plan` · Critic: Codex `/root`, selected independent context · Round 4, affected scope only.

Plan: [docs/multi-audio/plan-s7.md](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md), SHA-256 `c8a30d71806c85eb05016d5a3777f9f442c030aa7bdb6bfedc41257acaaf02dd`.

Snapshot: `/Users/strelia/projects/courses`, branch `multi-audio-s7`, HEAD `3f0e1c84e97eedd6aa2537410803cc14165e5afe`. Plan/probe/prior verdicts are untracked. Dirty playback implementation is excluded; player-path comparisons use HEAD.

Prompt: **Plan critic — version 1**, `prompts/plan-critic.md` in candidate `17a9527421e8fac3e43a16c287c404118699833c`, verified SHA-256 `514e4c376c2551afe76e390c5779867ecf7f04ad2a552c97c2b4f713c43dcab6`. Package declares `0.8.0-alpha.1+native-handoff.1`, unreleased; release integrity remains unverified.

1. **NONBLOCKING — OPEN: preserve reproducibility and qualify browser identity.**  
   [Probe:5](/Users/strelia/projects/courses/docs/multi-audio/_wip-s7-probe/probe.mjs:5) uses default headless Chromium and `Desktop Chrome`, matching the relevant [project configuration:49](/Users/strelia/projects/courses/frontend/playwright.config.ts:49). Installed Playwright/hls.js versions match the plan.

   However, the retained packet lacks the fixture-generation command, generated HLS/JS hashes, invocation/exit and runtime browser version. Consequently, identical browser/build bytes are **not established**. Retain these details with the implementation evidence. This limits reproducibility; it does not negate the supplied capture observation.

2. **NONBLOCKING — OPEN: correct the sample-count description.**  
   [Plan:121](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:121) describes stability for “30 samples.” The log contains 30 post-switch observations: four still show 440 Hz, followed by **26** at 880 Hz, beginning at 1221 ms. Correct the count; the five-consecutive-sample requirement remains comfortably supported.

The three requested checks resolve as follows:

- **Capture route — round-3 blocker RESOLVED.** [Probe HTML:10–16](/Users/strelia/projects/courses/docs/multi-audio/_wip-s7-probe/index.html:10) captures the video’s decoded audio, connects only to an analyser and records one audio track with a running context. The log demonstrates baseline 439.45 Hz followed by sustained 878.91 Hz while playback advances. HEAD’s player uses an ordinary video element with hls.js MSE and no separate WebAudio output path. Nginx serves the media on the application origin; authorization changes delivery, without introducing a different capture mechanism. The probe’s real click followed by `play()` supports the planned Play gesture. These observations establish feasibility, **not an integrated player/e2e PASS**.
- **Classifier — round-3 nonblocker RESOLVED.** [Plan:115–118](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:115) specifies FFT size, frequency tolerance, signal floor, consecutive samples and deadline. The recorded peaks satisfy those thresholds. Persistent original audio, silence and a single transient cannot satisfy the proposed transition. The 10-second window is a test budget supported by this local observation, not a general switching guarantee.
- **Helper invocation — round-3 nonblocker RESOLVED.** [Plan:108](/Users/strelia/projects/courses/docs/multi-audio/plan-s7.md:108) now supplies `'Англійська'`, matching the helper’s accessible-name lookup at [ASR:156](/Users/strelia/projects/courses/frontend/e2e/asr.spec.ts:156) and existing caller at line 441.

All three raw prior verdicts were read. Earlier language-identity, validation and SWITCHING dispositions are retained without reopening their implementation scope. Plan B remains inactive and requires its stated fresh critique.

Evidence SHA-256:

```text
probe.mjs  a165ab18504c72175ff1dbd75be0839956d6f95645a2b40761500aad0c62d8d0
index.html 681d147e304b177eb59a83c0e8764f8574c183decf694370a2b7464f2d4b44a3
run.log    870d09643b6a01715031302eb96be29602383a13aeb0d798da0750f1302ee732
```

**ACCEPTED — the three named round-4 planning questions, at this exact snapshot. Zero open blockers; two nonblocking findings.**

No tests, probes, graph refresh or writes were performed. **0 additional reviewers launched / 0 returned**; this verdict does not establish the separate formal four-lens gate. The record is returned here because the filesystem is read-only.