# Report M08 — ffmpeg Transcoder adapter

## Outcome

M08 is implemented in the adapter slice only:

- `backend/internal/media/adapters/ffmpeg/transcoder.go` implements `ports.Transcoder` over
  `ffprobe`, `ffmpeg`, and storage paths derived only from validated UUID/claim-token keys.
- `backend/internal/media/adapters/ffmpeg/transcoder_test.go` records exact argv fixtures and
  the `/bin/sleep` cancellation observation without requiring ffmpeg.

No domain, port, application, worker, PostgreSQL, HTTP, tus, frontend, dependency, or
generated file was changed. `work/tasks.md` and `docs/media/report.md` were not edited.

## Implemented contract

### Probe and thumbnail

- Probe argv is exactly `ffprobe -v error -show_streams -show_format -print_format json --
  <source>`.
- The JSON decoder ignores unknown members, extracts the first video and audio streams, and
  populates width, height, codecs, and `format.duration` into `domain.SourceProbe`, including
  `DurationSeconds`.
- Missing codec/dimension values are refused at the existing parse → `domain.AcceptSource`
  seam before rendering; missing, negative, non-finite, or non-numeric duration is refused by
  the parser rather than becoming zero.
- Thumbnail argv seeks to `0` when `DurationSeconds < 1`, otherwise `1`, and uses
  `scale='min(640,iw)':-2`, so it cannot upscale.

### Render argv

The fixed one-, two-, and three-rung argv fixtures pin:

- H.264 High, `yuv420p`, capped CRF 21, and ceilings/buffers `365k/730k`,
  `3000k/6000k`, `6000k/12000k`;
- level 4.1 on 360p/720p and no explicit level on 1080p;
- frame-rate-independent `expr:gte(t,n_forced*2)` on every video stream and no probed frame
  rate in argv;
- AAC-LC stereo at 128k, one audio output, and an audio-first shared `agroup`;
- fMP4 HLS, six-second VOD segments, `independent_segments`, the `init.mp4` template, and
  `master.m3u8`;
- each rung's `min(width,iw)` / `min(height,ih)` scale with
  `force_original_aspect_ratio=decrease:force_divisible_by=2`.

The fixtures cover 1920×1040, 1080×1920 portrait, and 320×240 sources. The last keeps the
mandatory 360p-labelled rendition while the `min()` caps prevent enlargement. The adapter
supplies only the `init.mp4` template; it never constructs `init_N.mp4`.

### No-user-argument and cancellation rules

The argv builders accept no original filename. Source and output path components are fixed
directories plus a strictly validated asset UUID and claim token; a non-prefix ladder or an
invalid storage key is rejected. R21's malicious original name is absent from render and
thumbnail argv, and both lists are compared element-by-element with fixed literals.

Every process is created with `exec.CommandContext` and `WaitDelay = 2s`. Cancellation uses
the command's default direct-process kill; there is no `Setpgid`, `SysProcAttr`, process-group
signal, shell, or hand-written cancellation hook. R20 starts `/bin/sleep`, observes that it
started before cancellation, requires a returned `context.Canceled`, and bounds completion by
`WaitDelay`.

## Review

The implementation workflow's code-review step was adapted to the explicit M08 files because
Git is forbidden. Two independent read-only axes reviewed the files.

### Standards axis

No hard violation was found. The adapter is stdlib-only, stays in the media adapter layer,
and follows ADR 009. Three non-blocking smell judgements were considered and retained:

- literal one-/two-/three-rung expected argv repeat stable sections intentionally, because
  deriving expected data from the production builder would make the exact-argv proof
  tautological;
- the private `resolveWork(..., thumbnail bool)` keeps one validation body for the two fixed
  temporary-key shapes; a new artifact abstraction would add machinery without another
  consumer;
- probe, render, thumbnail, and process lifecycle remain in the one file the M08 placement
  plan explicitly assigns to this adapter.

### Specification axis

The first pass found that the malicious-filename test exact-compared render argv but not the
thumbnail argv. The test was corrected to exact-compare both. The follow-up pass found no
remaining blocking specification issue and confirmed B13, B21, B28, R20, and R21 are
represented at their prescribed seams.

## Verification performed

- `gofmt -w backend/internal/media/adapters/ffmpeg/transcoder.go
  backend/internal/media/adapters/ffmpeg/transcoder_test.go` — exit 0.
- `gofmt -d` over the same two files — exit 0 with no diff.
- Source read-through checked the exact fixed argv, the duration branch, storage-key
  validation, parser fields, `CommandContext`, and `WaitDelay`.
- Static read-through found no `Setpgid`, `SysProcAttr`, `Process.Kill`, direct `exec.Command`,
  or constructed `init_N` name. `OriginalName` occurs only in the R21 test fixture/assertion.
- No filesystem write was denied, so there is no denied-write error text to report.
- No Git command was run and no commit was created.

## Checks not run — denied by the brief

The exact environment denial is: **“No Docker: no Go compile or test. gofmt + read-through.”**
The following checks are unrun and none is represented as passed:

- focused `go test ./internal/media/adapters/ffmpeg` — not run; Go compilation and test
  execution are denied;
- the R20 `/bin/sleep` test — written but not run because it is a Go test;
- direct Go compilation, `go test`, `go vet`, and `go build` — not run;
- `make format-check` — not run; its Go path uses Docker;
- `make lint` — not run; its Go vet path compiles through Docker;
- `make typecheck` — not run; Go compilation is denied;
- `make test` — not run; Go compilation and tests are denied;
- `make integration` — not run; Docker is denied and it is outside M08's unit evidence;
- `make check-generated` — not run; it performs Git-backed drift checking and Git is denied;
- `make build` — not run; Docker and Go compilation are denied;
- `make up`, `make seed-identities`, `make seed`, and `make smoke` — not run; Docker Compose
  is denied and these checks are outside M08;
- `make e2e` — not run; Docker Compose is denied and it is outside M08;
- aggregate `make check` — not run because it contains the denied checks above;
- ffmpeg and ffprobe execution — not run; the brief requires argv/parsing evidence without
  either binary and forbids a test that requires ffmpeg or skips when it is absent.

No skip-if-missing guard was added.

## Reviewer verification (measured in the reviewing session)

`make test` and `make integration` pass — `courses/internal/media/adapters/ffmpeg` green,
with **zero `exec.LookPath` and zero `Skip(`**: the suite genuinely needs no ffmpeg rather
than hiding behind a conditional skip, which AGENTS.md forbids. Verified by inspection:
the scale expression is the `min(640,iw)`/`min(360,ih)` form with `force_divisible_by=2`
(the shape execution proved necessary, twice); the Apple constants `-hls_time 6`,
`-hls_flags independent_segments` and the frame-rate-independent
`force_key_frames expr:gte(t,n_forced*2)` are pinned; cancellation uses `WaitDelay` with
**no `Setpgid`**, matching D-10 — the point my own brief had got backwards. The
no-user-argument rule has a real adversarial test: the original name is set to
`"; rm -rf / --preset x.mp4` and asserted absent from both render and thumbnail argv.
format-check and lint pass.

This task stopped once and was right to: D-11's thumbnail seek needed a duration no
contract carried, and the brief contradicted D-10 on cancellation. Both were repaired
before it resumed (plan §17). Provenance: implemented with Codex (gpt-5.6-sol); reviewed
and verified here.
