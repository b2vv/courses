# M16 report — the first real video through the whole chain

## Files

- `frontend/e2e/media.setup.ts` — generates three 2-second clips and the dedicated 5-minute
  R35 clip with `-crf 32` inside `media-worker` `/tmp`, copies them to the ignored
  `frontend/playwright/.fixtures/` directory, and refuses empty copied fixtures.
- `frontend/e2e/media.spec.ts` — uses the stored admin/student states and a per-run course
  namespace to cover the browser upload journey, real tus wire, worker output, playlists,
  metrics, restart recovery, server authorization and post-publish media refusal.
- `frontend/playwright.config.ts` — adds `media.setup.ts` to the existing setup project.
- `.gitignore` — excludes the generated media fixture directory.
- `work/report-m16.md` — this report.

No application code, dependency, generated file, task tracker or phase report was changed.

## Behaviour and evidence written

- The setup project creates exactly the four §5.f fixtures with the worker image's ffmpeg:
  640×360, 1280×720 and 320×240 for 2 seconds, plus 1280×720 for 5 minutes. Only the long
  clip carries `-crf 32`; all source bytes begin in the container's `/tmp` and copied bytes
  live only under the ignored Playwright fixture directory.
- The serial journey creates one namespaced draft course with four saved lessons through the
  public admin API, opens the real editor with stored admin state, and adds zero logins.
- P04/P08/B05 upload the 640×360 fixture through the real file control, record a positive
  determinate progress value, wait for `ready`, verify readiness survives reload, and inspect
  the worker volume for the master, one 360p variant, shared audio playlist, one `init_*.mp4`
  and at least one `.m4s` in each rendition, plus the thumbnail. The 720p and 1080p paths are
  asserted absent.
- B25 scrapes the worker's real `/metrics` endpoint from inside the container and requires the
  `ready` histogram count to be at least one.
- B26/B27 transcode the 1280×720 fixture and assert two video variants, one audio media group
  referenced by both, one `audio/` directory, and video-only streams in both video variant
  playlists. B26 is repeated against the 320×240 output. Every stream line must carry
  `CODECS`, `RESOLUTION` and `BANDWIDTH`; `AVERAGE-BANDWIDTH` and `FRAME-RATE` are absent from
  every inspected playlist; `EXT-X-INDEPENDENT-SEGMENTS` is present only in video playlists,
  not the master or `audio/index.m3u8`. The 360p `BANDWIDTH` is exactly `542300` and
  byte-identical across the 1280×720 and 320×240 sources.
- B30 requires one 320×240 master variant and probes the HLS video stream with the worker's
  ffprobe, requiring `320x240` from the encoded output.
- R26 opens a student browser context from stored state, confirms the upload control is absent,
  then calls the mint endpoint through `page.request` and requires `403 forbidden`.
- P05/R35 upload the dedicated long fixture, observe the editor's real
  `data-video-state="processing"` before running `docker compose restart media-worker`, wait
  for the restarted worker to become healthy, and require the same open editor to reach
  `ready` without navigation.
- R13 repeats all three path shapes anonymously and with the admin request context after the
  first transcode is ready. The nested request addresses the real published master playlist,
  and all six responses must be `403`.
- Cleanup deletes the namespaced draft and restarts the worker so its startup sweep can remove
  the now-orphaned asset rows and media rather than leaking roughly 100 MB per E2E run.

## Acceptance rows

| Row           | Written evidence                                                                | Runtime verdict here |
| ------------- | ------------------------------------------------------------------------------- | -------------------- |
| P04           | real 640×360 transcode and exact published-file shape                           | not run              |
| P05           | long-clip editor visibly reaches processing then ready without navigation       | not run              |
| P08           | browser file control, positive determinate progress, ready, reload-stable ready | not run              |
| B05           | one 640×360 variant; 720p/1080p absent                                          | not run              |
| B25           | ready histogram count ≥1                                                        | not run              |
| B26           | exact 5.1.9 presences, absences and nominal-bandwidth equality on two sources   | not run              |
| B27           | one shared audio group/directory and video-only variant streams                 | not run              |
| B30           | master and ffprobe both report 320×240                                          | not run              |
| R26           | stored-student `page.request` receives `403 forbidden`                          | not run              |
| R35           | processing observed before restart; same editor later reaches ready             | not run              |
| R13 post half | real published master plus `/media` and `/media/`, two auth modes, all 403      | not run              |

## Discussion

### R13's literal nested path does not exercise the published file

Evidence: R13 names `/media/x/master.m3u8`, but §23 says M16 owns the post-transcode repeat
because an existence-dependent edge rule could refuse missing media while exposing a
published file. After a successful transcode, the actual master is
`/media/hls/{videoId}/master.m3u8` per D-8. The literal `x` path remains missing and cannot
detect the leak §23 describes.

Bug consequence: a configuration that returns 403 for missing paths but serves existing HLS
objects could pass all six literal R13 requests.

Proposed correction: keep M13's pre-transcode literal path matrix, but define M16's nested
post-transcode path as `/media/hls/{publishedVideoId}/master.m3u8`. The written spec uses that
real path; it does not silently claim that `/media/x/master.m3u8` proves publication privacy.

### A 2-second main fixture cannot deterministically expose the processing badge

Evidence: P05 and P08 ask the main 640×360 browser journey to visibly pass through
`processing`, while §5.f records that all 2-second fixtures render in well under one second
and D-12 polls every 3 seconds. The same section creates a 5-minute clip specifically because
the short clips cannot reliably expose `processing` to Playwright.

Bug consequence: making P05/P08 require the 640×360 badge would be a timing race. The upload
can move from `queued` through `processing` to `ready` between two legitimate UI polls, so a
correct product would intermittently fail the test.

Proposed correction: let the 640×360 journey own determinate upload progress, ready and
reload-stable ready; co-assert P05's no-navigation `processing → ready` observation in R35's
long-fixture test. The long clip is still uploaded solely for R35; its already-required
processing observation supplies the deterministic UI-state evidence rather than creating a
second long upload.

## Verification performed

- Targeted `oxfmt --check` passed for `media.setup.ts`, `media.spec.ts`,
  `playwright.config.ts` and this report after formatting the new spec.
- `cd frontend && npm run typecheck` passed (`tsc --noEmit`, including both new Playwright
  files).
- `cd frontend && npm run lint` passed with zero warnings; the architecture check reported
  `Architecture boundaries checked (78 files).`
- `cd frontend && npm run build` passed for client and SSR. It emitted the pre-existing
  warning that `src/routes/route-guards.test.tsx` is discovered as a route file without a
  `Route` export; both builds completed.
- A read-only standards/spec review rechecked the stored-state login budget, all owned rows,
  the B26 ffmpeg 5.1.9 distinctions, the R13 post half, the real M15 DOM, runtime-only media,
  cleanup, and every NOT-list item.
- No filesystem write was denied. No Git command was run. No file in `generated/`,
  `work/tasks.md`, or `docs/media/report.md` was written.

## Environment denials and unrun checks

Environment denial: **Docker is unavailable here.** No Docker command was run and no
alternative container route was attempted. Fixture generation, worker ffmpeg/ffprobe,
Compose restart/health, storage inspection, metrics scraping, cleanup sweep, `make
integration`, `make build`, `make up`, both seeds, and `make smoke` remain unrun.

Environment denial: **a browser is unavailable here.** Playwright was not invoked, including
its setup project. The real tus headers, upload progress, polling, storageState behavior,
authorization calls and all acceptance rows in the table remain unverified. `make e2e` was
not run; the reviewer must reseed identities between the two consecutive E2E runs used for
the 429 check, exactly as the brief states.

The remaining ordered checks not run are `make format-check`, `make lint`, `make typecheck`,
`make test`, `make integration`, `make check-generated`, `make build`, and `make e2e`;
therefore aggregate `make check` was not run. Their frontend-local format/lint/typecheck/build
subsets are the only checks represented as passing above. `npm test` was also not run because
M16 changes no Rstest seam and it does not execute Playwright specs. `make check-generated`
is additionally unavailable because it invokes forbidden Git and generation would write
inside forbidden `generated/` directories. No substitute drift check was attempted.

## Review

### Standards

No open finding in the permitted read-only review. The change stays in E2E/config/report
scope, uses the existing Compose and stored-session idioms plus Node stdlib, adds no dependency,
validates unknown response data, keeps media bytes ignored, and contains no app-code or
generated edit.

### Specification

All M16-owned rows and R13's post half have explicit written assertions. Runtime evidence is
pending the denied reviewer environment. The two bug-causing plan inaccuracies are not hidden;
they are isolated above with evidence and proposed corrections.

The repository `code-review` skill's prescribed fixed-point `git diff`/`git log` review and
parallel diff sub-agents were not run: the user forbids every Git operation and supplied no
legal fixed point. The manual two-axis source/spec review is not represented as that formal
diff review.

Review summary: Standards — 0 open findings. Specification — 2 plan inaccuracies handled in
Discussion; 11 acceptance rows written and unverified under the named Docker/browser denials.

## Reviewer verification (measured in the reviewing session)

Both Discussion items accepted, recorded as plan **§25** — and the first one caught a flaw in
my own §23: repeating R13's six refusals after a transcode proves nothing while the row names
a path that is still missing afterwards, so the post-transcode check now uses the real
published path. The second removed a test that would have been flaky by construction: a
2-second fixture renders in under a second against a 3-second poll, so demanding a visible
`processing` badge in the short journey would fail intermittently on a correct product.

**Three review fixes, all in the test layer, none in the product:**
1. `docker compose cp` cannot read the worker's `/tmp` because it is a tmpfs mount, so the
   fixtures never reached the host. Bytes are streamed out through `exec cat` instead.
2. ffprobe repeats a stream entry per playlist segment, so `toEqual(['video'])` and
   `toBe('320x240')` failed on correct output. Both assertions now test the row's subject —
   no audio muxed into a video variant, and the sub-360p source never enlarged — as sets
   rather than as line counts.
3. The media journey restarts the worker and saturates its two CPUs with a five-minute
   transcode, starving specs running beside it (a 4b dialog timed out at 5s). It now has its
   own Playwright project running after `chromium`, exactly as `recovery` already does for the
   same reason.

**Result: `make e2e` exit 0, 35 passed — twice consecutively**, with an identity reseed and a
limiter reset between, each run printing `total=7, maximum consecutive=3`. The first real
video went the whole way: a person picks a file, tus carries the bytes past the Go API, the
worker claims the job, ffmpeg 5.1.9 encodes it, the playlist publishes atomically, and the
asset reaches `ready` — while `/media/*` keeps refusing everyone, now proven on the real
published path.

Provenance: implemented with Codex (gpt-5.6-sol); reviewed, corrected and verified here.
