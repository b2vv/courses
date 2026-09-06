# Brief M08 — the ffmpeg Transcoder adapter

Mode: ponytail-full. **No test here may require ffmpeg** — the integration image has none,
and a conditional skip is the trap AGENTS.md forbids. Prove argv and parsing exactly;
prove process behaviour against `/bin/sleep`.

## Read first, in this order
1. `docs/media/plan.md` — **D-11** (the argv: the Apple conformance constants — H.264 High
   ≤L4.1 8-bit 4:2:0, AAC-LC stereo, `-hls_time 6`, `-hls_flags independent_segments`, the
   2 s `force_key_frames expr:gte(t,n_forced*2)` which is frame-rate-independent so no
   probed fps enters argv, the audio `agroup`, ceilings 365k/3000k/6000k under capped CRF,
   `avc1` not `avc3` — and **`force_divisible_by=2`** plus the scale expression
   `scale=w='min(640,iw)':h='min(360,ih)':force_original_aspect_ratio=decrease:force_divisible_by=2`
   which must never upscale), **D-8** (`init_N.mp4` naming follows `var_stream_map` order,
   and NO consumer may construct that name), **§0's fact table** (every ffmpeg fact with the
   binary it was verified on — the worker ships **5.1.9**, not the host's 7.1.1),
   **§2.1/§2** (`Transcoder` signatures: `Probe`, `Render`, `Thumbnail`), and your owned
   rows **B13, B21, B28, R20, R21**.
2. `docs/adr/009-fmp4-cmaf-segments.md` (M01) — the recorded constants.
3. `backend/internal/media/ports/transcoder.go` + `domain/source.go` (M03/M03a) — what you
   implement and the ladder you serve.
4. Brief §11's hard rule: **no user-controlled argument reaches the FFmpeg command line.**
   The original filename must never appear in argv (nor in any error surfaced outward).

## The task (plan §10 M08, 1.75h)
- `backend/internal/media/adapters/ffmpeg/transcoder.go`: argv builders for render and
  thumbnail from the FIXED ladder plus probed dimensions only; ffprobe JSON parsing into
  `domain.SourceProbe`; `exec.CommandContext` with `WaitDelay` so a cancelled render cannot
  leave a child running.
- Tests (`make test`, **no ffmpeg, no Docker**):
  - **R21** — exact argv comparison for the render, including the `min()` scale expression
    and `force_divisible_by=2`; **B21** — the argv for 1920×1040 and 1080×1920 sources;
    **B28** — the argv pins the Apple conformance constants; **B13/R20** — read their row
    text (ffprobe parsing and the no-user-argument rule; assert the filename appears nowhere
    in argv).
  - Process behaviour: cancellation kills the process group, proven against `/bin/sleep`,
    not ffmpeg.

## What this task is NOT
No worker wiring or asynq (M12), no postgres, no HTTP, no tus, no frontend. Do not modify
the domain, ports or application — if a contract is missing, STOP and report.

## Environment limits — name the denial, do not work around it
No Docker: no Go compile or test. gofmt + read-through. **Do not add a test that needs
ffmpeg, and do not add a skip-if-missing guard.** List every unrun check. Denied write →
report the exact text and stop. **No git; do not edit work/tasks.md; do not write
docs/media/report.md.** Never write inside generated/.

## Report
`work/report-m08.md`.
