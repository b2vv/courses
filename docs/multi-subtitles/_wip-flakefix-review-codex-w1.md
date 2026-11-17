Reviewed `flakefix-w1` at `ea1fe47`; the three-file diff exactly matches frozen patch `bd927dbd…`.

One finding:

- [frontend/e2e/playback.spec.ts:1087](/Users/strelia/projects/courses/frontend/e2e/playback.spec.ts:1087) — **non-blocking, high confidence.** “hls.js keeps 30 s ahead” overstates the diagnosis. Installed hls.js computes its buffer target from bitrate and `maxBufferSize`, with `maxBufferLength=30` and a 600-second ceiling (`base-stream-controller.ts:1451–1462`). The trace establishes request inactivity, without measuring the actual buffered range. Impact: misleading maintenance guidance; the fix remains valid. Prefer describing the observed absence of requests.

The requested checks otherwise hold:

1. **S5 product fix:** Correct. Installed `router-core` **1.171.38** selects `_pendingLocation || latestLocation` in `buildLocation`, then passes that location’s search to the functional updater. `adminListState(previous)` preserves existing per-field validation/defaulting. Callers retain their explicit behavior: filters reset pagination, paging preserves filters, and clearing removes search/status while preserving sort. Untouched fields now come from the latest navigation.

2. **S5 test:** It discriminates by construction: the held search loader keeps the old component rendered, so the old implementation drops `search` on sort. Changing `defaultPendingMs` before loading is sound: the router reads the option during loading. Extending the browser’s normal one-second window makes the race deterministic without inventing a different state. I could not independently execute the reported old-code failure.

3. **B18:** The trace claim is verified. Network line **84** starts at `1217991.751`; `setOffline(true)` starts at `1218042.463`—**50.712 ms later**. No media requests, or other requests, follow during the failed wait. The replacement proves an observable network failure and retains the persisted-position assertion: acknowledged pause position, fresh page, first `playing` position within ±1 second.

   `no-store` bypasses HTTP cache; it does **not** bypass service workers. No service-worker registration exists in the inspected application/test sources. A cached successful response would fail this assertion, rather than silently pass it.

4. **Sibling rows:** No other offline/unavailable wait remains. B28’s unavailable assertion follows an observed missing-master **404**, so it lacks B18’s buffered-player mechanism.

5. **Standards:** No material architecture, TypeScript, or dependency violations found against `AGENTS.md` and `.claude/standards.md`.

Verification: typecheck **passed**. Rstest was **BLOCKED before execution** by sandbox `listen EPERM ::1:3000`. No edits, Git mutations, or Docker operations performed. This review does not establish a green merge gate.

**VERDICT (flakefix-w1): SHIP — 0 BLOCKER, 1 non-blocking.**