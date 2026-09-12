# Phase 6 — playback: delivery report

Delta to [plan.md](plan.md) and [spec.md](spec.md). Written 2026-09-12 at phase close from
the task evidence in `work/report-t1.md` … `work/report-t12.md` and the reviewer's ordered
gate recorded in T11 and T12. There is no production deployment to check.

## 1. What was done

A signed-in student can open a published course's learning surface, see its ordered outline,
start the server-selected lesson and play real HLS through nginx. The API mints a server-side
playback identity; nginx asks the playback module for authorization on every playlist and
segment request and serves the media bytes itself. Heartbeats renew the account-wide playback
slot and persist monotonic progress. Reloading or changing browser context restores an
unfinished position, reaching 95% marks the lesson watched, and reopening a completed lesson
starts it from zero. A second tab, device or course playback takes the same account-wide slot
and the displaced session loses media access. These claims are exercised by T4–T10 and the
60-test browser run in the ordered gate.

The same surface exposes explicit empty, no-video, processing and failed states, keyboard
controls, Ukrainian and English copy, light and dark themes, visible focus, 200% zoom support
and 44 px controls. T11 measured those browser properties with the qualifications in §6.
T12 added a bounded, once-per-page AV1 capability observation to the playback mint seam; its
frontend and backend halves are contract-bound but are not one wire-crossing test.

This phase did not add payments, orders or enrolment, comment threads, transcripts, an
instructor workspace or a production deployment.

## 2. Acceptance scenarios

The header below §8 in the plan says **94**, but the three tables actually contain **102
unique rows**: P01–P14, B01–B47 plus B21a, and R01–R39. The evidence supports **89/102**.
Rows not verified at the planned outcome are **B07, B08, B20, B21, B21a, B22, B23, B28,
B31, B36, R23, R30 and R34**. A green aggregate gate is not substituted for a missing
assertion.

The reviewer ran the ordered `make check` on 2026-09-12 against the tree that became `0959fca`
(T12) and `efc9851` (T11) — the gate preceded those commits and saw that content, not the
committed state:
`GATE_EXIT=0`; format-check, lint and typecheck passed; frontend unit tests were **379 passed,
0 failed**; Go `-race ./...` reported **48 packages ok**; integration, generated-code check,
build, Compose startup, both seed steps and smoke passed; e2e was **60 passed, 0 failed,
0 did not run** in **7.7 minutes**.

In the table, `gate` means that ordered run. The named T-report is the primary row-specific
proof; `gate` only records the final execution of proofs that actually exist.

| ID | Planned observation | Actual proof and run | Verdict |
|---|---|---|---|
| P01 | Real HLS plays and segments return 200 | T7 real-playback e2e; gate 60/60 | matched |
| P02 | Outline marks the lesson completed through real playback | T8 playback→heartbeat→learning-model e2e; T8 47/47 and gate | matched |
| P03 | Start uses server-selected `continueLessonId` | T8 e2e; T8 47/47 and gate | matched |
| P04 | Position survives reload | T9 e2e; T9 52/52 and gate | matched |
| P05 | Position survives another browser context | T9 e2e; T9 52/52 and gate | matched |
| P06 | A 100 s asset completes at 95% | T9 dedicated-fixture e2e; T9 52/52 and gate | matched |
| P07 | A completed lesson reopens at zero | T9 e2e; T9 52/52 and gate | matched |
| P08 | Playback outlives ownership TTL through renewals | T7 e2e, at the 2/6 test cadence; gate | matched with §6 timing limit |
| P09 | hls.js manual quality selection requests that rung | T7 e2e; gate | matched |
| P10 | Seeking into unloaded media requests the covering segment | T7 e2e; gate | matched |
| P11 | Admin plays an archived course | T7 admin e2e; gate | matched |
| P12 | AV1 probe result reaches the mint seam and increments a series | T12 frontend/backend tests plus OpenAPI; frontend 379 and Go 48-package gate | matched with split-proof limit |
| P13 | A second device stops the first | T10 two-context e2e; T10 59/59 and gate | matched |
| P14 | The limit applies across courses | T10 two-course e2e; T10 59/59 and gate | matched |
| B01 | Empty course has an empty state and no player | T6 domain/HTTP/PostgreSQL evidence; Go/integration gate | matched |
| B02 | Lesson without video has an explicit state | T7 component test; frontend gate | matched |
| B03 | Processing video has an explicit state | T7 component test; frontend gate | matched |
| B04 | Failed video has an explicit state | T7 component test; frontend gate | matched |
| B05 | 95.0 completes and 94.9 does not | T3 domain boundary tests; Go gate | matched |
| B06 | NULL duration saves position without completion or clamp | T3 domain and PostgreSQL integration; Go/integration gate | matched |
| B07 | Shorter replacement clamps a saved position | No task report contains the required integration observation | **not verified** |
| B08 | Replacing video preserves the watched mark | T2 explicitly did not close it; no later report supplies the integration observation | **not verified** |
| B09 | 90 s of playing time produces exactly three writes | T9 frontend cadence test; frontend gate | matched |
| B10 | Pause, seek, hide and navigation each write | T9 frontend event tests; frontend gate | matched |
| B11 | Stale `clientSeq` cannot rewind progress | T9 PostgreSQL integration; integration gate | matched |
| B12 | Nginx is healthy and its cache stays empty | T4 smoke; smoke gate | matched |
| B13 | Three media requests cause three authorizations | T4 smoke counter; smoke gate | matched |
| B14 | Anonymous replay of an allowed URL returns 403 | T4 same-URL smoke transition; smoke gate | matched |
| B15 | The same segment changes 200→403 immediately after takeover | T10 e2e; T10 59/59 and gate | matched |
| B16 | An abandoned session does not delay a new play | T10 e2e; T10 59/59 and gate | matched |
| B17 | Repeated play gives the caller the fresh slot | T5 Redis/PostgreSQL integration; integration gate | matched |
| B18 | Network loss retains the last saved position | T9 e2e; T9 52/52 and gate | matched |
| B19 | The 120-lesson boundary renders all rows | T8 unit fixture; frontend gate | matched for rendering only |
| B20 | A 1,024-byte range returns 206 and exactly 1,024 bytes | No row-specific smoke proof was found | **not verified** |
| B21 | Three ranges cause three authorizations | No row-specific smoke proof was found | **not verified** |
| B21a | Generated playlists contain no `EXT-X-BYTERANGE` | No named integration assertion was found | **not verified** |
| B22 | Media responses carry no-store and security headers | Smoke requests `/media`, `/media/` and `/media/x/master.m3u8` — all refusal paths. It never requests an allowed segment, so headers on a 200 are outside what it asserts | **not verified** |
| B23 | Media rate limiting returns 429 | Smoke measures 429 in the API, login and SSR zones; there is no media zone in it, and its own PASS line names only those three | **not verified** |
| B24 | Player is keyboard-operable and has no axe violations | T7 component/axe tests; frontend gate | matched |
| B24a | Focus is visible and controls are at least 44 px | T11 real-engine e2e in light/dark and before zoom; gate | matched |
| B25 | Contrast and 200% layout pass | T11 real-engine e2e; gate, with `readyState=0`, width/height 0/0 | matched only over empty video; see §6 |
| B26 | Locale base keys/forms agree and values are non-empty | T8 locale unit traversal; frontend gate | matched |
| B27 | Missing progress reads as concrete zero values | T6 domain/HTTP/PostgreSQL evidence; Go/integration gate | matched |
| B28 | DB `ready` with missing files gives edge 404 and visible error | T5 delegated this to T7; T7 proves only the UI 404 directive, not the edge/filesystem precondition | **not verified** |
| B29 | A new session writes on its first heartbeat | T9 PostgreSQL/integration sequence seeding; integration gate | matched |
| B30 | Zero probe duration persists NULL and cannot complete | T2 worker/PostgreSQL integration; integration gate | matched |
| B31 | Returning caller retakes the slot and the displaced UI stops | T5 Redis/PostgreSQL integration proves fresh ID and old-denied/new-allowed backend semantics, but no browser test proves the displaced UI stops | **not verified at the planned e2e level** |
| B32 | Takeover removes byte access immediately | T10 same-segment e2e; T10 59/59 and gate | matched |
| B33 | Two tabs displace one another | T10 same-browser e2e; T10 59/59 and gate | matched |
| B34 | Restore alone cannot complete after a shorter replacement | Same scenario and proof as duplicate B41: T3 domain/T5 PostgreSQL; Go/integration gate | matched |
| B35 | `clientSeq=2^62` returns 400 without a write | T3/T5 validation and integration evidence; Go/integration gate | matched |
| B36 | Redis failure returns 503 with `Retry-After` and no bytes | T1 proves 503/no bytes but does not record `Retry-After` | **not verified** |
| B37 | Edge starts healthy with an empty cache | T1 smoke; final smoke gate | matched |
| B38 | Paced 400-request parallel run is 400/400 HTTP 200 | T1 measured 400/400, no false 403/429; final smoke gate | matched |
| B39 | Expiry remints silently rather than reporting takeover | T7/T9 e2e; gate | matched |
| B40 | Existing generation plays through replacement; reopen gets new | T7 e2e 200/no 403 or 404, then fresh generation; gate | matched |
| B41 | Position must grow before completion | T5 PostgreSQL integration; integration gate | matched |
| B42 | Growing to the end completes | T5 PostgreSQL integration; integration gate | matched |
| B43 | Invalid heartbeat/TTL ratio refuses API startup | T3 configuration unit test; Go gate | matched |
| B44 | Oversized position is accepted and clamped | T5 HTTP plus PostgreSQL integration; Go/integration gate | matched |
| B45 | Migration succeeds over populated data | T2 live PostgreSQL integration; integration gate | matched |
| B46 | Synthetic failing migration rolls back only itself | T2 live PostgreSQL integration; integration gate | matched |
| B47 | Lock timeout rolls back and retry succeeds | T2 live PostgreSQL integration; integration gate | matched |
| R01 | Anonymous playlists and segments return 403 | T1/T4 smoke; smoke gate | matched |
| R02 | Student cannot play a draft course | T5 real-mint smoke; smoke gate | matched |
| R03 | Student cannot play an archived course | T5 real-mint smoke; smoke gate | matched |
| R04 | Archiving stops a started student view | T7 e2e checks a previously served segment becomes 403; gate | matched |
| R05 | Expired ownership refuses the first segment | T5 real-mint, real-TTL smoke; smoke gate | matched |
| R06 | Another user cannot use the playback ID | T5 real-mint smoke; smoke gate | matched |
| R07 | Course visibility is rechecked per media request | T5 exact-ownership then inaccessible-course smoke; smoke gate | matched |
| R08 | Non-HLS media trees are closed | T5 smoke; smoke gate | matched |
| R09 | Traversal variants return 403/404 without bytes | T5 active-ID smoke; smoke gate | matched |
| R10 | Internal authorize is not externally routed | T1 smoke; smoke gate | matched |
| R11 | Unavailable Go returns 503 | T1 dependency-failure smoke; smoke gate | matched |
| R12 | Unavailable Redis returns 503 without bytes | T1 dependency-failure smoke; smoke gate | matched |
| R13 | Unavailable Postgres returns 503 | T1 dependency-failure smoke; smoke gate | matched |
| R14 | Displaced playback cannot bypass client stop | T10's stronger same-segment 200→403 proof; e2e gate | matched by stronger observation |
| R15 | Displaced heartbeat returns 409 and changes neither slot nor progress | T5 Redis/PostgreSQL integration; integration gate | matched |
| R16 | Non-owner heartbeat does not extend another owner's TTL | T5 Redis integration; integration gate | matched |
| R17 | Inaccessible-course heartbeat returns 403 and creates no progress | T6 distinct-student Redis/PostgreSQL integration; integration gate | matched |
| R18 | Unknown lesson returns 404 and the UI has a not-found state | T5 HTTP 404 plus T7 recovery/UI tests; Go/frontend gate | matched by split proof |
| R19 | Malformed lesson ID returns 400 before application lookup | T5 HTTP contract test; Go gate | matched |
| R20 | Malformed edge IDs return 403 without calling Go | T1 smoke counter; smoke gate | matched |
| R21 | Anonymous public course response exposes no lesson IDs | T6 HTTP contract test; Go gate | matched |
| R22 | Malformed heartbeat fields return 400 without changing progress | T5 HTTP/application/PostgreSQL evidence; Go/integration gate | matched |
| R23 | API and SSR network output stay far below video size | No `docker stats` byte-volume measurement was found | **not verified** |
| R24 | The generated route set contains no “my courses” route | T8 exact route-set unit test plus generated check; frontend/generated gate | matched |
| R25 | Two accounts behind one IP play together | T10 e2e; T10 59/59 and gate | matched |
| R26 | One account from two network addresses still takes over | T10 mixed browser/container request using nginx source address; gate | matched |
| R27 | Logout changes the same segment from 200 to 403 | T5 real-mint smoke; smoke gate | matched |
| R28 | A displaced client stops sending progress | T9 frontend recovery test; frontend gate | matched |
| R29 | Anonymous and non-owner receive byte-identical 403 bodies | T4 smoke; smoke gate | matched |
| R30 | Nginx/Go safe logs cannot acquire request URI | Safe configuration exists, but no named unit/lint assertion was found | **not verified** |
| R31 | Unknown playback ID heartbeat returns 410 | T3 Redis/integration evidence; integration gate | matched |
| R32 | Deleted lesson heartbeat returns 404 without FK error | T5 application/HTTP/integration evidence; Go/integration gate | matched |
| R33 | Displaced heartbeat cannot write progress | T3/T5 Redis/PostgreSQL integration; integration gate | matched |
| R34 | Course owner plays their own archived course and receives segments | T6 proves shared policy, application and PostgreSQL read; it explicitly records no segment-200 e2e. T7 P11 is admin, not owner | **not verified** |
| R35 | hls.js ABR steps down under a narrowed channel | T11 e2e; two reviewer runs and gate | matched only as one step; see diagnostics and §6 |
| R36 | `/learn` leads to the catalog | T8 behavior e2e; T8 47/47 and gate | matched |
| R37 | Huge position clamps to the known 300 s duration | T8 dedicated e2e; T8 47/47 and gate | matched |
| R38 | Media 403 is classified by heartbeat before recovery/denial | T9 unit+e2e with O1 counters; T9 52/52 and gate | matched |
| R39 | A non-button pause advances persisted progress | T10 paired before/after learning-model e2e; T10 59/59 and gate | matched |

T11's mandatory diagnostics were:

- `B25 video at contrast audit`: `readyState=0`, `videoWidth=0`, `videoHeight=0`.
- `R35 master BANDWIDTH`: 360p **542300 bit/s**, 720p **3440800 bit/s**; derived
  throttle **170750 B/s** (**1365996 bit/s**).
- R35 run 1: baseline 360p=1, 720p=15; after throttle 360p=2, 720p=19.
- R35 run 2: baseline 360p=1, 720p=5; after throttle 360p=2, 720p=8.
- The decisive request sequence was `720p/segment_00024 → 720p/segment_00025 →
  360p/segment_00025`: abandon-and-retry of the same segment index at the lower rung.

T12's live metric scrape found exactly three pre-seeded zero series: `probe_failed`,
`supported` and `unsupported`. Its Help text counts browsers whose bounded probe reached the
mint endpoint, including denied mints. The live API run did not itself observe a 0→1 mint or
a repeated mint without the field; those transitions are the named Go and frontend tests in
the ordered unit/integration portions of the gate.

## 3. Non-obvious findings

- **The plan's own count drifted.** Its “94 rows” footer hides eight rows; direct enumeration
  yields 102. The delivery count therefore uses the IDs, not the stale footer.
- **A read-only proof hid a progress-write authorization bug.** T6's review replaced it with
  a real heartbeat attempt by a distinct student and moved the shared visibility check before
  Redis renewal and PostgreSQL progress write.
- **An edge denial needs a before-state.** T4 and T10 found that a bare 403 could pass because
  bytes were already unreadable for another reason. The lasting proofs use the same URL as
  200 before logout/takeover and 403 after it.
- **An ABR request is not a settled quality level.** The repeated segment index is strong
  evidence of a downward abandon-and-retry, while the `+1` lower-rung count is insufficient
  evidence that playback remained there.

## 4. Screenshots

No static screenshots were captured. Browser-visible behavior was exercised by the Compose
Playwright suite; T11 recorded computed theme, geometry, zoom, axe and video-state diagnostics
rather than treating a screenshot as an assertion.

## 5. User flow walk-through

Verified by the named browser scenarios, not by hand. A signed-in student starts in the public
catalog, opens a published course and enters its learning route at the server-selected lesson.
The outline shows modules, lessons and watched marks. Pressing play mints a session and real HLS
starts; quality can be left on Auto or chosen manually in hls.js. Progress is written while
playing and on exit events, restored after reload and in another browser context, and reflected
in the outline. At 95% the lesson becomes watched; reopening it starts at zero. If the same
account starts another playback, the former view stops and explains why. Archive and logout
remove byte access on the next checked request. Empty, unavailable and processing media states
remain readable UI states rather than blank pages.

## 6. Risks and limitations

- **Native HLS is unproven (D-9).** Production checks `Hls.isSupported()` first, and both
  Chromium and Playwright WebKit measured `{hlsJsSupported:true,mse:true,canPlayType:"maybe"}`;
  both therefore used hls.js. Playwright WebKit is not Safari. A real Safari run is required,
  and no claim is made that the native branch works.
- **A native fallback is dead and uncovered.** No test reaches the
  `return createNativeMedia(...)` after a failed `canPlayType` in `hls-media.ts`.
- **B25 audited an empty video element.** The course read model was supplied, but mint/media
  were not, so the measured element was `readyState=0`, `videoWidth=0`, `videoHeight=0`.
  Contrast over actual playing imagery was not tested. The earlier prediction that a video
  background would create an axe `incomplete` was not confirmed because the fixture never
  created that state.
- **One axe `incomplete` remains:** `color-contrast`/`nonBmp` for React Aria Select's
  decorative chevron. It is axe declining to decide a glyph-only node, not a measured contrast
  defect. Remove the pin when axe can decide non-BMP glyph contrast, or when an updated UI kit
  changes the chevron DOM so the current stable suffix no longer identifies it.
- **R35 proves a step, not settling.** The channel is restored immediately after the first
  360p request, and both runs add exactly one 360p request. The supported claim is “ABR steps
  down under a narrowed channel,” not “the player transitions to and stays at 360p.”
- **R35 uses a synthetic channel.** The **170750 B/s** throttle is the geometric mean of the
  two master-playlist BANDWIDTH values, not a real network condition. It proves reaction to a
  controlled capacity between rungs, not correctness on real networks.
- **Ownership timing was compressed.** Playback e2e recreates the API at heartbeat/TTL **2/6**,
  while the shipped Compose defaults are **30/180**. Both satisfy `heartbeat*3 <= ttl`, but
  renewal, expiry and takeover behavior was not timed at the shipped cadence.
- **AV1 telemetry begins only at the mint endpoint.** It includes denied mints, but a client
  consumes its page-scope probe before sending; if that request never reaches the server, the
  observation is lost. This narrow loss window avoids double counting when a response is lost.
- **P12 has no single wire-crossing test.** Frontend proves the field is sent on a denied mint;
  backend proves a received field increments the series; shared `openapi.yaml` binds the two.
- **Thirteen planned observations remain unverified:** B07, B08, B20, B21, B21a, B22, B23,
  B28, B31, B36, R23, R30 and R34. Their exact missing evidence is recorded in §2.
- **One audio rendition only; no AV1 rendition.** Multi-language audio delivery is not built,
  and the AV1 work is telemetry/probe only.
- **Still out of scope:** payments, orders and enrolment; transcripts; comments/discussions;
  instructor workspace; “my courses”; notes, bookmarks, search, offline mode and certificates;
  frontend observability; TLS/S01/S02; and production deployment. None is claimed here.

The next task is to rebuild the stale 2026-09-05 `work/AGENDA.md` against this phase close,
then take the selected Phase 7 slice through `spec-flow`. The roadmap currently names
transcripts and synchronized document links, with student video comment threads added by
`work/PRODUCT-BRIEF.md` §24.4; none of that roadmap work is delivered by this phase.
