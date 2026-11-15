Reviewed `f00b0ff..f415643` on `multi-subtitles`. No files changed or tests run; counter-examples were traced through source. No production code defect found.

**Ranked findings**

- **Major — required shipping evidence remains incomplete.** [report.md:43](docs/multi-subtitles/report.md#L43), [report.md:77](docs/multi-subtitles/report.md#L77), [report.md:54](docs/multi-subtitles/report.md#L54):
  - The recorded ordered gate passed on `4ed7ed0`, before the frontend fix. The final-HEAD gate is still pending; [plan-ms1.md:152](docs/multi-subtitles/plan-ms1.md#L152) requires it.
  - [Plan line 146](docs/multi-subtitles/plan-ms1.md#L146) requires every U1–U4 row across all listed conditions. U1 lacks 200% text verification; U2 is desktop-only; U4 is 390-only. U2/U4 omissions are honestly labelled **NOT_TESTED**, but remain required. U1’s omission is only a dash. Counter-example: U3’s enlarged modal does not establish U1’s enlarged inline popover placement.
  - The retained-volume diagnostic query and before/after results required by [plan lines 89–91](docs/multi-subtitles/plan-ms1.md#L89) are absent. Counter-example: one successful relabel and seven regenerated masters can coexist with unexplained remaining `und` tracks on another Ready asset with a primary transcript.

- **Minor — screenshot provenance contradicts the report.** [report.md:89](docs/multi-subtitles/report.md#L89) says the screenshots above are after the fix, under revision `5778964`. However, [390-U4-keyboard-focus.jpg](docs/multi-subtitles/ms1-design-pass/390-U4-keyboard-focus.jpg) shows `Original EN`; the corrected U1 screenshot shows `Original English`. Relabel the older evidence or replace it. This does not itself establish a keyboard defect.

No Blocker finding.

**Wave-1 dispositions**

- **Major: OPEN, partly resolved.** Opened the wave-1 record, MS1 plan/report and all 15 screenshots. The recorded Compose events cover both stopped-stack passes, repeated-up worker stability and failed-regeneration non-start. Their stated ordering satisfies the report requirement; same-second timestamps alone would not establish it. Browser and final-gate gaps remain above.
- **Minor M-R7: RESOLVED.** Opened [regenerate_integration_test.go:115](backend/cmd/media-worker/regenerate_integration_test.go#L115), command wiring and production writer. Both initial masters now pass through `writer.RegenerateMaster`. Assertions retain migrated `uk`, absence of `und`, video URI preservation, repeatable bytes and unchanged evicted bytes. Traced wrong-root, unchanged-language and unintended-eviction-rewrite counter-examples: each is detected. The shared embedded rendition ID is harmless because embedded URI generation ignores it.

**Delta checks**

- [Fixture:204](backend/internal/transcript/adapters/postgres/draft_integration_test.go#L204): `zz-` plus eight hexadecimal characters is valid, 11 characters total, and derived separately from each new track UUID. It is probabilistically unique, not mathematically guaranteed; truncation is unchanged. Searches across backend/frontend/e2e/SQL expressions found no other invalid positive language generator. Migration 00028 generates valid `und`/`und-x-original`; oversized generated tags are intentional rejection tests.
- [languageName:72](frontend/src/modules/media/ui/LessonAudioTracksField.tsx#L72): lowercasing affects display lookup only. Option IDs, rename state and submitted codes remain raw; `pt-BR` stays unchanged. Checked audio-tracks and ASR selectors, including localized `Dub (Англійська)`: no selector regression found.
- All screenshot files exist. U1/U3 images support the shown option states; U2 supports its desktop cell. Static screenshots cannot independently prove the reported Escape/focus-return sequence.

**VERDICT: BLOCK** — required evidence remains incomplete.