# ADR 005 — HLS data plane in Nginx
Status: accepted design; media serving denied in foundation.

Superseded in part: the MPEG-TS `.ts` segment clause is replaced by [ADR 009](009-fmp4-cmaf-segments.md). The Nginx data plane, authorization and the denial of public media access before Phase 6 stand.

Choose MPEG-TS `.ts` segments for first HLS implementation: broad H.264/AAC compatibility and straightforward FFmpeg/debug tooling. fMP4 saves overhead and supports future low-latency/CMAF workflows but is not required. Nginx serves playlists/segments after auth_request, Go checks session/enrollment and issues a 3-minute scoped grant. Every segment request must be authorized; relative playlist URLs must retain credentials.

No public media alias is enabled before Phase 6. Grants, refresh, replay/concurrency policy, range behavior and negative authorization tests are required before enabling it. This does not prevent an authorized user from recording video and is not DRM.
