# ADR 008 — tus ingest in a separate executable
Status: accepted design, 2026-09-06; implementation scheduled for Phase 5.

Terminate tus 1.0.0 core plus `expiration` in a third Go executable beside `api` and
`media-worker`. Its entire surface is `HEAD`, `PATCH`, and `DELETE`, with no `OPTIONS`,
`creation`, `creation-defer-length`, `concatenation`, or `checksum`. The cookie-authenticated
JSON mint remains the only creation point, so ownership, declared length, and the 2 GiB limit
are decided once. Upload writes chunks directly to the media volume; the Go API and frontend
SSR server never receive video bytes.

The minted `video_assets` row is the capability: its random UUID is both ticket and asset id,
and the upload executable validates row state and expiry without reading cookies, roles, or
sessions. This avoids both a signed token's secret/replay protocol and an authorization
callback that would expose the browser cookie to the byte receiver. Swap the bounded core for
tusd when parallel chunks, checksum verification, or an S3 backing store is required.

Named Q1 deviation: completion advances media state and enqueues through shared backend ports
instead of notifying the API. Brief rule 12 permits this and the API still sees metadata only.
Add one internal completion endpoint when a second ingest terminator needs a single owner for
state advancement.

Named Q2 deviation, accepted as an amendment on 2026-09-06: sniff the container on the first
`PATCH`, but check H.264/HEVC and AAC/MP3 with ffprobe in the worker. Unsupported codecs become
explicit non-retryable processing failures; FFmpeg and subprocess execution stay out of the
exposed upload executable. Revisit only when early codec refusal is measured as necessary and
can preserve that boundary. Phase 5 keeps `/media/*` refused and claims no playback.
