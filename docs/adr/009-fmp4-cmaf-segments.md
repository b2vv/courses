# ADR 009 — fMP4/CMAF HLS segments
Status: accepted design, 2026-09-06; implementation scheduled for Phase 5.

Use HLS with CMAF-style fMP4 `.m4s` segments and separate video and audio streams. This ADR
supersedes only ADR 005's MPEG-TS `.ts` segment clause; its Nginx data plane, authorization,
and denial of public media access before Phase 6 remain. fMP4 lets the Phase 6 player, future
DASH, and a future AV1 rendition reuse the same bytes. Evidence is in
[`research-codecs.md`](../media/research-codecs.md).

Encode H.264 High Profile, 8-bit 4:2:0 with capped CRF 21 at 360p/365 kbit/s,
720p/3000 kbit/s, and 1080p/6000 kbit/s only when supported by the source; `maxrate` is the
ceiling and `bufsize` twice it, with level 4.1 only on the lower two rungs. Never upscale, but
produce one source-sized rendition below 360p. Use one shared AAC-LC stereo rendition at
128 kbit/s, `avc1`, 6-second target segments, IDR every 2 seconds, and
`EXT-X-INDEPENDENT-SEGMENTS`.

The shipping binary is ffmpeg 5.1.9 from `debian:bookworm-slim`. Verified in
`courses-media-worker-1`, `-hls_fmp4_init_filename init.mp4` produces one `init_N.mp4` per
variant in `var_stream_map` order: audio first yields `audio/init_0.mp4`, then
`360p/init_1.mp4` and `720p/init_2.mp4`. Consumers follow variant playlists and never construct
the init name.

ffmpeg 5.1.9 emits `CODECS` and `RESOLUTION`, but its three accepted Apple non-conformances are
absent `AVERAGE-BANDWIDTH` (9.14), absent `FRAME-RATE` (9.15), and nominal rather than measured
peak `BANDWIDTH` (9.13, 1.26/1.27). It puts `EXT-X-INDEPENDENT-SEGMENTS` only in video variant
playlists and omits audio `CHANNELS`. Write no custom playlist; when the Phase 6 player needs
measured ABR bandwidths, upgrade the worker binary or append the missing master-playlist lines.

H.264/AAC is the only Phase 5 delivery codec; HEVC is excluded. Per `work/PRODUCT-BRIEF.md`
§24.14, add AV1 only beside H.264 and only when egress has real cost or measured
bandwidth-bound sessions, Phase 6 telemetry shows at least 50% stable AV1 decoder support, and
the worker can afford a second SVT-AV1 pass without breaking publication latency. Phase 5
keeps `/media/*` refused and claims no playback.
