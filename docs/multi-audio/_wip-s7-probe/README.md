# S7 capture-feasibility probe (2026-11-14)

Question: does Playwright headless Chromium expose an hls.js MSE video's decoded audio through
`captureStream()` so an AnalyserNode can tell 440 Hz from 880 Hz? Answer: yes (`run.log`).

- Browser: Playwright 1.62.1 headless Chromium 151.0.7922.34, `devices['Desktop Chrome']` (as `playwright.config.ts` project `chromium`).
- hls.js 1.7.2 (`frontend/node_modules/hls.js/dist/hls.min.js`, copied next to `index.html`).
- Fixture (in `<probe>/hls`):

```bash
ffmpeg -nostdin -loglevel error -f lavfi -i testsrc2=size=320x180:rate=15 -f lavfi -i sine=frequency=440 -f lavfi -i sine=frequency=880 -t 20 -map 0:v -map 1:a -map 2:a -c:v libx264 -preset veryfast -c:a aac -ac 2 -b:a 128k -f hls -hls_time 4 -hls_playlist_type vod -hls_segment_type fmp4 -master_pl_name master.m3u8 -var_stream_map "v:0,agroup:aac a:0,agroup:aac,name:orig,language:und,default:yes a:1,agroup:aac,name:dub,language:en" -hls_segment_filename "s%v_%03d.m4s" "p%v.m3u8"
```

- Run: `cd frontend && node <probe>/probe.mjs <probe>` → exit 0, output `run.log`.
- Result: 6 baseline samples at 439.5 Hz (−48…−38 dB, first sample empty before data); after
  `nextAudioTrack = 1`: 4 samples still 439.5 Hz, then **26** at 878.9 Hz from 1221 ms; `currentTime` 2.8→11.6 s.
- An unmuted `play()` without a gesture is refused (`NotAllowedError`); the probe clicks the video first.
- SHA-256: probe.mjs a165ab18504c72175ff1dbd75be0839956d6f95645a2b40761500aad0c62d8d0; index.html 681d147e304b177eb59a83c0e8764f8574c183decf694370a2b7464f2d4b44a3; run.log 870d09643b6a01715031302eb96be29602383a13aeb0d798da0750f1302ee732; 
