1. **MAJOR — Wave-2 #1 CLOSED**, `frontend/e2e/subtitles.spec.ts:89`. Progress reset uses portable `docker compose` with the inherited environment, removing the machine-specific launcher failure.

2. **MAJOR — Wave-2 #2 CLOSED**, `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:35`. Errors use `alert`; loading and absence use `status`. Scoped tests assert exactly one appropriate live region.

3. **MAJOR — Wave-2 #3 OPEN — PENDING**, `docs/multi-subtitles/plan-ms4.md:338`. Ordered `make check`, full-e2e/M25 and design-pass/M26 remain scheduled owner work. They are **not code blockers** for this review.

4. **MINOR — Wave-2 #4 CLOSED**, `frontend/e2e/subtitles.spec.ts:1060`. E9 polls actual playback progress with a 15-second bound instead of a fixed readiness sleep.

5. **MAJOR — Wave-2 #5 CLOSED**, `frontend/e2e/subtitles-cleanup.ts:10` and `frontend/e2e/subtitles.spec.ts:568`. Teardown transactionally deletes audit records, grants, linked assets and the course; schema cascades remove dependent audio, transcripts and progress. Registering the course before timestamp validation covers later partial-setup failures. Repeated successful runs no longer retain the globally counted rows identified in wave 2.

No new standards findings in the fix delta. Uncertain audio spans now retain the attempted language, with a regression covering failed replacement and return. The same-language refetch change matches the dated M5 owner decision and has a held-refetch regression checking captions and search/follow preservation.

Static review only; no tests executed.

VERDICT: SHIP