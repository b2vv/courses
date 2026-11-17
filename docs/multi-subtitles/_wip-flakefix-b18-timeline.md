# B18 attempt-8 trace: request timeline (derived; no headers, cookies or bodies)

Source: Playwright trace of `playback.spec.ts:1073` B18, `make check` attempt 8, 2026-11-17. Monotonic ms.

- 1217947.7 GET 720p/segment_00024.m4s 200
- 1217959.9 GET audio/segment_00025.m4s 200
- 1217960.1 GET 720p/segment_00025.m4s 200
- 1217971.5 GET audio/segment_00026.m4s 200
- 1217971.5 GET 720p/segment_00026.m4s 200
- 1217972.3 GET audio/segment_00027.m4s 200
- 1217972.5 GET 720p/segment_00027.m4s 200
- 1217989.9 GET audio/segment_00028.m4s 200
- 1217990.1 GET 720p/segment_00028.m4s 200
- 1217991.7 GET audio/segment_00029.m4s 200
- 1217991.8 GET 720p/segment_00029.m4s 200
- 1218033.7 POST /api/v1/playback/heartbeat 204

- **1218042.5 `setOffline(true)`** — 8.8 ms after the last request above; no request follows until the 60 s wait times out.
