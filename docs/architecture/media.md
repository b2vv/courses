# Media design and foundation boundary

Foundation starts an Asynq worker with health/readiness and queue metrics, but no transcoding handlers. Unrecognized tasks fail; no job is marked ready. Nginx denies all media requests. No upload, playback token or transcription API is advertised as working.

Phase 5 introduces application ports `MediaStorage`, `Transcoder` and job repository. Storage keys, not filesystem paths, cross the application boundary. The filesystem adapter resolves keys inside an owned root and rejects traversal/symlinks; an S3/MinIO adapter can implement the same logical publish/read/write contract. Local publication uses a same-filesystem atomic rename from temporary work output. Object storage publishes immutable version prefixes and commits a manifest pointer, rather than pretending S3 rename is atomic.

Layout: source/{videoId}/original, hls/{videoId}/master.m3u8 and 360p/720p/1080p, thumbnails/, temporary/. ffprobe supplies dimensions/duration; fixed FFmpeg command generation yields H.264/AAC MPEG-TS HLS, no user flags. Only profiles supported by source are generated, with no upscaling. A worker lease and idempotent publication protect retries; cancel terminates subprocess group and cleans temporary output. Failed publication never advertises ready.

Phase 6 uses session + enrollment checks, 3-minute playback grants, concurrency limits and Nginx auth_request. Authorization must cover each playlist and segment. Token delivery will use an HttpOnly scoped media cookie so relative segment URLs inherit access; logout/revocation and refresh are part of the future contract. Never log grants/cookies/query strings. Nginx alone serves bytes with range/cache/CORS rules. This is access control, not DRM.

Future transcription port accepts an asset reference and returns timestamped segments with confidence/speaker and documentation links. Manual fixtures and a versioned transcript contract precede an adapter. No Python runtime is installed in foundation.
